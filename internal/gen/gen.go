// Package gen is the in-memory generation core behind the CLI. It ties together
// parse -> validate -> consolidate -> render for every target (application.yml,
// kubernetes manifests, docker compose, podman quadlet). All I/O (folder
// scan, env lookups, file reads, path resolution) is injected by callers so the
// package stays pure and testable.
package gen

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"runtime"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/consolidate"
	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/deploy"
	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/dockergen"
	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/logback"
	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/podmangen"
	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/render"
	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/spec"
	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/statusscript"
	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/tls"
	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/validate"
)

// fileEnv labels issues that are not tied to a specific workflow file.
const fileEnv = "env.yaml"

// File is one named input document.
type File struct {
	Name string
	Data []byte
}

// Request is a full set of inputs: the env.yaml document and the workflow files.
type Request struct {
	Env       *File
	Workflows []File
}

// Resolver injects the environment-dependent bits generation needs.
type Resolver struct {
	// Env looks up an environment variable (CLI: os.LookupEnv; nil skips checks).
	Env func(string) (string, bool)
	// ReadFile reads a values-file or .jks referenced by config (nil = no access).
	ReadFile func(string) ([]byte, error)
	// Abs resolves a config-relative path (store/libs dir) to a host path for
	// bind mounts. Nil leaves the path unchanged.
	Abs func(string) string
	// Rand fills b with cryptographically random bytes, used to generate the
	// reserved status account's password (see resolveStatusPassword). Tests
	// inject a deterministic or failing implementation; nil means
	// crypto/rand.Read.
	Rand func(b []byte) error
	// Now is the time a created store's certificates are judged against for
	// expiry. Tests pin it; nil means time.Now.
	Now func() time.Time
	// GOOS is the operating system the command runs on, which bounds what one
	// environment variable holds for a docker created store (EnvVarMax).
	// Tests pin it; empty means runtime.GOOS.
	GOOS string
}

func (res Resolver) now() time.Time {
	if res.Now == nil {
		return time.Now()
	}
	return res.Now()
}

func (res Resolver) goos() string {
	if res.GOOS == "" {
		return runtime.GOOS
	}
	return res.GOOS
}

func (res Resolver) abs(p string) string {
	if res.Abs == nil {
		return p
	}
	return res.Abs(p)
}

// KV is one resolved credential entry (re-exported for CLI env-file writing).
type KV = deploy.KV

// Issue re-exports validate.Issue for callers that only import gen.
type Issue = validate.Issue

// Config parses+validates the request and returns the application.yml -- the
// same document deploy ships on every platform: each store at its mount path,
// each credential a placeholder resolved from the mounted secrets. One folder
// is one connector instance: validate rejects a folder holding more than
// validate.MaxWorkflows workflows rather than splitting it.
func Config(r Request, res Resolver) (out string, errs, warns []Issue) {
	wfs, e, pissues, ewarns := parse(r, res)
	verrs, w := validate.Run(validate.Context{Workflows: wfs, Defaults: &e.Defaults, Env: res.Env})
	errs = append(pissues, verrs...)
	warns = append(ewarns, w...)
	if len(errs) > 0 {
		return "", errs, warns
	}
	statusPW, err := resolveStatusPassword(res)
	if err != nil {
		return "", []Issue{{File: fileEnv, Msg: err.Error()}}, warns
	}
	b, cw, err := build(wfs, &e.Defaults, statusPW)
	warns = append(warns, toIssues(cw)...)
	if err != nil {
		return "", []Issue{{File: fileEnv, Msg: err.Error()}}, warns
	}
	return b.appYAML, nil, warns
}

// Validate runs every check for every section present and returns all findings:
// every error generate and deploy would stop at before they reach a platform,
// short of what only the deploy host can tell -- the validate verb adds those
// (docker's bind-mount sources) -- so a clean validate leaves only the live
// system to fail: logins, the image, the cluster or engine itself. An unset
// -env variable is a warning.
func Validate(r Request, res Resolver) (errs, warns []Issue) {
	wfs, e, pissues, ewarns := parse(r, res)
	verrs, w := validate.Run(validate.Context{
		Workflows:       wfs,
		Defaults:        &e.Defaults,
		Image:           e.Image,
		Timezone:        e.Timezone,
		JavaOptions:     e.JavaOptions,
		Kube:            e.Kubernetes,
		Docker:          e.Docker,
		Podman:          e.Podman,
		CheckKubernetes: e.Kubernetes != nil,
		CheckDocker:     e.Docker != nil,
		CheckPodman:     e.Podman != nil,
		Env:             res.Env,
		Lint:            true,
	})
	errs, warns = append(pissues, verrs...), append(ewarns, w...)

	// A mount-name collision is only visible once names are assigned, which
	// happens in consolidate rather than in validate.Run -- so lint it here
	// instead of leaving it to generate/deploy, matching the rule every other
	// credential check follows: catch it while authoring, not mid-deploy. The
	// stores the rendered connections select are known only then too, so the
	// store files are read, and the created stores built, here as well
	// (validateStores).
	//
	// Only on an otherwise-clean spec: consolidate assumes validated input.
	// The status password is a placeholder because it is rendered as a literal
	// and never becomes a mount name, so it cannot affect the result, and the
	// rendered document is discarded.
	if len(errs) == 0 {
		b, _, berr := build(wfs, &e.Defaults, statusPasswordPlaceholder)
		if berr != nil {
			return append(errs, Issue{File: fileEnv, Msg: berr.Error()}), warns
		}
		serrs, swarns := validateStores(e, b.model.Stores, res)
		errs, warns = append(errs, serrs...), append(warns, swarns...)
	}
	return errs, warns
}

// statusPasswordPlaceholder stands in for the reserved status account's password
// on the validate path, which never emits the document it renders.
const statusPasswordPlaceholder = "validate-only-placeholder"

// GenerateKubernetes parses+validates and returns the manifest set on success.
// Credentials/stores are resolved and embedded (stringData/base64), so this is
// the same rendered output the deploy path applies.
//
// extraAllowed threads deploy/remove's --allow-command values into the
// kubernetes.command allowlist check; plain `generate kubernetes` calls this
// with none, so an exotic command validates clean only at deploy time.
// KubeOpts steers the kubernetes render, mirroring PodmanOpts rather than
// widening the signature every time a caller needs a variant.
type KubeOpts struct {
	// Teardown renders the set for `delete -f -`: no Namespace document, and the
	// manifest carrying a Namespace, piped to `kubectl delete -f -`, takes the
	// namespace and everything else living in it. Nothing is resolved for it:
	// every Secret is rendered by name alone (deploy.Input.Teardown), so remove
	// needs no exported credential and no store source.
	Teardown bool
	// Probes, when non-nil, receives the Secrets the rendered pod reads but the
	// tool does not write -- each operator Secret a store names
	// (existing.secret) and an existing credentials Secret -- with the keys the
	// pod needs from each, for deploy to check on the cluster before applying.
	Probes *[]SecretProbe
}

// SecretProbe is one Secret the operator manages that the pod mounts, and the
// data keys the pod reads from it.
type SecretProbe struct {
	Name string
	Keys []string
	// Field is the env.yaml position that names the Secret, for messages.
	Field string
}

func GenerateKubernetes(r Request, res Resolver, opts KubeOpts, extraAllowed ...string) (out string, errs, warns []Issue) {
	wfs, e, pissues, ewarns := parse(r, res)
	k := e.Kubernetes
	if k == nil {
		pissues = append(pissues, Issue{File: fileEnv, Msg: "kubernetes target requires a 'kubernetes:' section in env.yaml"})
	}
	verrs, w := validate.Run(validate.Context{
		Workflows:       wfs,
		Defaults:        &e.Defaults,
		Image:           e.Image,
		Timezone:        e.Timezone,
		JavaOptions:     e.JavaOptions,
		Kube:            k,
		CheckKubernetes: true,
		Teardown:        opts.Teardown,
		Env:             res.Env,
		AllowCommands:   extraAllowed,
	})
	errs = append(pissues, verrs...)
	warns = append(ewarns, w...)
	if len(errs) > 0 {
		return "", errs, warns
	}

	statusPW, err := resolveStatusPassword(res)
	if err != nil {
		return "", []Issue{{File: fileEnv, Msg: err.Error()}}, warns
	}
	b, cw, err := build(wfs, &e.Defaults, statusPW)
	warns = append(warns, toIssues(cw)...)
	if err != nil {
		return "", []Issue{{File: fileEnv, Msg: err.Error()}}, warns
	}

	runAs, rerrs := runAsOf("kubernetes", k.RunAs)
	if rerrs != nil {
		return "", rerrs, warns
	}
	in := deploy.Input{Kube: k, Defaults: &e.Defaults, Syslog: e.Defaults.Syslog, Teardown: opts.Teardown}
	switch {
	case opts.Teardown:
		// Delete needs names only: no credential is resolved, no store opened.
		if name, _ := k.ImagePullSecretName(); name != "" {
			in.ImagePull = &deploy.PullSecret{Name: name}
		}
	default:
		if c := k.Secrets.Credentials; c != nil && c.Create != nil {
			kvs, err := ResolveCredentials(b.model.Secrets, res)
			if err != nil {
				return "", []Issue{{File: fileEnv, Msg: err.Error()}}, warns
			}
			in.CredKVs = kvs
		}
		files, sw, err := resolveStores(b.model.Stores, e.Defaults.TLS, res)
		warns = append(warns, sw...)
		if err != nil {
			return "", []Issue{{File: fileEnv, Msg: err.Error()}}, warns
		}
		in.Stores = files
		if name, created := k.ImagePullSecretName(); name != "" {
			ps, err := resolvePullSecret(name, created, e.Image, res)
			if err != nil {
				return "", []Issue{{File: fileEnv, Msg: err.Error()}}, warns
			}
			in.ImagePull = ps
		}
	}
	if opts.Probes != nil {
		*opts.Probes = secretProbes(k, b.model)
	}
	in.Instance = deploy.Instance{
		Name:         k.Deployment.Name,
		Image:        e.Image.Ref(),
		Timezone:     e.Timezone,
		AppYAML:      b.appYAML,
		StatusScript: statusscript.Render(deploy.ManagementPort(in), spec.StatusUserName),
		Model:        b.model,
		JavaOptions:  e.JavaOptions,
		RunAs:        runAs,
	}
	return deploy.Render(in), nil, warns
}

// secretProbes lists the operator-managed Secrets the rendered pod mounts and
// the keys it reads from each: every Secret a selected store names, in
// first-use order, then an existing credentials Secret with every credential's
// mount name.
func secretProbes(k *spec.Kubernetes, m *consolidate.Model) []SecretProbe {
	var out []SecretProbe
	at := map[string]int{}
	for _, st := range m.Stores {
		if !st.ClusterHeld() {
			continue
		}
		key := st.Secret.EffectiveKey(st.Name)
		i, ok := at[st.Secret.Name]
		if !ok {
			i = len(out)
			at[st.Secret.Name] = i
			out = append(out, SecretProbe{Name: st.Secret.Name, Field: "tls.stores." + st.Name + ".existing.secret"})
		}
		if !slices.Contains(out[i].Keys, key) {
			out[i].Keys = append(out[i].Keys, key)
		}
	}
	if name, created := k.CredentialsSecretName(); name != "" && !created && len(m.Secrets) > 0 {
		keys := make([]string, 0, len(m.Secrets))
		for _, r := range m.Secrets {
			keys = append(keys, r.Stable)
		}
		out = append(out, SecretProbe{Name: name, Keys: keys, Field: "kubernetes.secrets.credentials.existing"})
	}
	return out
}

// DockerPlan is the rendered compose plus the secret references it declares.
// Values are not resolved here: the compose file names secrets only, and the CLI
// resolves the values into the `docker compose` child environment at deploy time
// so nothing secret is ever written to disk.
type DockerPlan struct {
	Compose string
	Secrets []consolidate.SecretRef
	RunAs   spec.RunAsID // the identity the compose service runs as, for deploy's mount checks
	Mounts  []HostMount  // the host paths the compose file bind-mounts, for the same checks
	// Stores are the created stores the compose file carries as secrets read
	// from the environment, built and encoded at deploy (ResolveDockerStores).
	Stores []DockerStore
	// Image is the reference the service runs, which deploy inspects before
	// overriding its entrypoint for the stores.
	Image string
}

// HostMount is one host path a compose file or podman unit bind-mounts, as the
// artifact names it -- expanded and absolute -- so deploy checks the path the
// container will read rather than the env.yaml spelling of it.
type HostMount struct {
	Field   string // the env.yaml key it comes from: StoreField(<name>) or MountLibs
	Path    string
	Private bool // a store some connection uses as a keystore: it holds a private key
}

// MountLibs is the env.yaml key of the libs directory mount.
const MountLibs = "libs.dir"

// StoreField is the env.yaml key an existing store's PKCS12 comes from, for
// messages.
func StoreField(store string) string { return "tls.stores." + store + ".existing.pkcs12" }

// GenerateDocker parses+validates and renders the docker-compose.yml (with
// application.yml inlined under compose configs:). Store/libs host paths are
// resolved to absolute so the compose is portable regardless of cwd.
//
// extraAllowed threads deploy/remove's --allow-command values into the
// docker.command allowlist check; plain `generate docker` calls this with none.
func GenerateDocker(r Request, res Resolver, extraAllowed ...string) (plan DockerPlan, errs, warns []Issue) {
	wfs, e, pissues, ewarns := parse(r, res)
	d := e.Docker
	if d == nil {
		pissues = append(pissues, Issue{File: fileEnv, Msg: "docker target requires a 'docker:' section in env.yaml"})
	}
	verrs, w := validate.Run(validate.Context{
		Workflows: wfs, Defaults: &e.Defaults, Image: e.Image, Timezone: e.Timezone, JavaOptions: e.JavaOptions, Docker: d, CheckDocker: true, Env: res.Env,
		AllowCommands: extraAllowed,
	})
	errs = append(pissues, verrs...)
	warns = append(ewarns, w...)
	if len(errs) > 0 {
		return DockerPlan{}, errs, warns
	}

	statusPW, err := resolveStatusPassword(res)
	if err != nil {
		return DockerPlan{}, []Issue{{File: fileEnv, Msg: err.Error()}}, warns
	}
	b, cw, err := build(wfs, &e.Defaults, statusPW)
	warns = append(warns, toIssues(cw)...)
	if err != nil {
		return DockerPlan{}, []Issue{{File: fileEnv, Msg: err.Error()}}, warns
	}

	runAs, rerrs := runAsOf("docker", d.RunAs)
	if rerrs != nil {
		return DockerPlan{}, rerrs, warns
	}
	sm, lm := targetMounts(b.model.Stores, d.Libs, res)
	stores := dockerStores(b.model.Stores, e.Defaults.TLS)
	in := dockergen.Input{
		Docker:  d,
		Syslog:  e.Defaults.Syslog,
		Secrets: stableNames(b.model.Secrets),
		Stores:  toDockerMounts(sm),
		Libs:    toDockerMount(lm),
		Created: dockerCreated(stores),
		Instance: dockergen.Instance{
			Name:         d.Name,
			Image:        e.Image.Ref(),
			Timezone:     e.Timezone,
			AppYAML:      b.appYAML,
			MQTLS:        b.model.MQTLS,
			StatusScript: statusscript.Render(e.Defaults.EffectiveManagementPort(), spec.StatusUserName),
			LeaderMode:   e.Defaults.LeaderElection.EffectiveMode(),
			JavaOptions:  e.JavaOptions,
			RunAs:        runAs,
		},
	}
	return DockerPlan{Compose: dockergen.Render(in), Secrets: b.model.Secrets, RunAs: runAs, Mounts: hostMounts(sm, lm), Stores: stores, Image: e.Image.Ref()}, nil, warns
}

// resolvePullSecret turns the image-pull block into the wiring deploy renders.
// A reference-only block resolves to the name alone -- deploy then emits the
// imagePullSecrets entry and no Secret, so the one the operator manages is left
// untouched. create additionally reads the registry password here and builds
// the payload, keeping deploy pure and the value inside this call.
func resolvePullSecret(name string, create bool, img *spec.Image, res Resolver) (*deploy.PullSecret, error) {
	out := &deploy.PullSecret{Name: name}
	if !create {
		return out, nil
	}
	user, pass := img.UserCred(), img.PassCred()
	// validate rejects create without both, so reaching here means a caller
	// skipped it. The guard makes that an error rather than a Secret built from
	// an empty account.
	if user.Empty() || pass.Empty() {
		return nil, fmt.Errorf("kubernetes.secrets.image-pull.create needs image.user and image.pass (or their -env forms)")
	}
	// Resolved through the same path as every other credential, so a literal
	// passes through, an -env form is read from the environment, and an unset
	// variable fails with a message naming both the position and the variable.
	kvs, err := ResolveCredentials([]consolidate.SecretRef{
		{Stable: "image.user", Literal: user.Literal, EnvVar: user.EnvVar},
		{Stable: "image.pass", Literal: pass.Literal, EnvVar: pass.EnvVar},
	}, res)
	if err != nil {
		return nil, err
	}
	doc, err := dockerConfigJSON(img.Registry(), kvs[0].Val, kvs[1].Val)
	if err != nil {
		return nil, fmt.Errorf("building the image-pull secret: %w", err)
	}
	out.DockerConfigJSON = doc
	return out, nil
}

// dockerConfigJSON renders the base64 .dockerconfigjson payload a
// kubernetes.io/dockerconfigjson Secret carries: an auths map keyed by registry
// host, holding the account plus the base64 "user:password" the engines send.
//
// It is marshalled rather than concatenated so a password carrying a quote or a
// backslash cannot break the document -- or, worse, reshape it.
func dockerConfigJSON(registry, user, pass string) (string, error) {
	type entry struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Auth     string `json:"auth"`
	}
	doc := struct {
		Auths map[string]entry `json:"auths"`
	}{Auths: map[string]entry{registry: {
		Username: user,
		Password: pass,
		Auth:     base64.StdEncoding.EncodeToString([]byte(user + ":" + pass)),
	}}}
	b, err := json.Marshal(doc)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(b), nil
}

// stableNames projects secret references down to the names an artifact declares.
func stableNames(refs []consolidate.SecretRef) []string {
	out := make([]string, 0, len(refs))
	for _, r := range refs {
		out = append(out, r.Stable)
	}
	return out
}

// PodmanFile is one file the quadlet unit mounts out of podman's secret store
// rather than from the host: a rendered document, carried whole in Data, or a
// TLS store -- an existing one named by Source, a created one by Build -- read
// or built only when deploy loads it (ResolvePodmanFiles); generate and remove
// never open a store.
type PodmanFile struct {
	StoreName string      // the secret's name in podman's store: <name>-<document>, or <name>-tls-<store>
	Data      string      // a rendered document; empty for a store
	Source    string      // an existing store's PKCS12 as written in env.yaml; empty otherwise
	Field     string      // a store's env.yaml key, e.g. tls.stores.mq-ca.existing.pkcs12, for errors
	Build     *StoreBuild // a created store, built when deploy loads it; nil otherwise
}

// PodmanPlan carries the rendered quadlet unit plus everything a deploy loads
// into podman's secret store before activating it. Nothing but the unit itself
// is written to the host.
type PodmanPlan struct {
	Unit podmangen.Unit // the .container quadlet unit
	// Files are the documents and stores the unit mounts from the secret store,
	// in unit order, under the same names its Secret= lines carry, so what
	// deploy creates cannot drift from what the unit mounts.
	Files   []PodmanFile
	Secrets []consolidate.SecretRef // credentials to place in podman's secret store
	Service string                  // systemd service name, e.g. name.service
	RunAs   spec.RunAsID            // the identity the unit runs as, for deploy's mount checks
	Mounts  []HostMount             // the host paths the unit bind-mounts (libs.dir alone), for the same checks
}

// The secret-store names of the documents a podman unit mounts, each appended
// to the container name by PodmanSecretStoreName. Every suffix holds a '-' or a
// '.', which no credential's stable name can (those match
// [A-Za-z_][A-Za-z0-9_]*), so a file never takes a credential's name. A store
// is named podmanStorePrefix + its tls.stores name.
const (
	podmanAppYAMLSuffix = "application.yml"
	podmanStatusSuffix  = "status-script"
	podmanLogbackSuffix = logback.FileName
	podmanStorePrefix   = "tls-"
)

// PodmanFileSecretNames lists every document secret an instance can own,
// mounted by its current spec or not: deploy deletes the ones it no longer
// mounts, and remove deletes them all. Store secrets are named after stores the
// spec may no longer define, so they cannot be listed from it; deploy and remove
// read them from the instance's own unit instead (PodmanUnitStoreSecrets).
func PodmanFileSecretNames(container string) []string {
	suffixes := []string{podmanAppYAMLSuffix, podmanStatusSuffix, podmanLogbackSuffix}
	out := make([]string, 0, len(suffixes))
	for _, s := range suffixes {
		out = append(out, PodmanSecretStoreName(container, s))
	}
	return out
}

// PodmanStoreSecretName is the secret-store name of a tls.stores entry the
// container's unit mounts.
func PodmanStoreSecretName(container, store string) string {
	return PodmanSecretStoreName(container, podmanStorePrefix+store)
}

// PodmanUnitStoreSecrets returns the store secrets a previously written unit
// for container mounts: the names of its Secret= lines that carry the
// container's store prefix. Reading the instance's own unit, rather than
// scanning podman's store by prefix, is what keeps another instance whose name
// happens to start the same way out of the result.
func PodmanUnitStoreSecrets(unit, container string) []string {
	prefix := PodmanStoreSecretName(container, "")
	var out []string
	for _, line := range strings.Split(unit, "\n") {
		v, ok := strings.CutPrefix(strings.TrimSpace(line), "Secret=")
		if !ok {
			continue
		}
		name, _, _ := strings.Cut(v, ",")
		if strings.HasPrefix(name, prefix) && len(name) > len(prefix) {
			out = append(out, name)
		}
	}
	return out
}

// podmanSecretMax is the most data one podman secret holds: containers/common
// refuses empty data and anything of 512000 bytes or more.
const podmanSecretMax = 511999

// podmanSecretSizeError refuses what podman's secret store cannot hold: an
// empty secret, or one over podmanSecretMax. what names the secret's source.
func podmanSecretSizeError(what string, n int) error {
	switch {
	case n == 0:
		return fmt.Errorf("%s is empty, and podman's secret store cannot hold an empty secret; point it at the real file", what)
	case n > podmanSecretMax:
		return fmt.Errorf("%s is %d bytes, over the %d bytes podman's secret store holds per secret; trim it to what this connector needs (for a truststore, the CA certificates of the brokers and queue managers it connects to)", what, n, podmanSecretMax)
	}
	return nil
}

// ResolvePodmanFiles turns a plan's files into the name/value pairs deploy
// loads into podman's secret store: a rendered document as it stands, an
// existing TLS store read from its file (resolved against env.yaml, as
// resolveStores reads it for kubernetes), a created one built (BuildStore),
// whose findings come back as warnings. Every value is held to podman's size
// limit here, so a file podman would refuse stops the deploy before any secret
// is created. The result is secret material -- callers hand it straight to the
// store.
//
// Errors name the file or secret and its size, never its content.
func ResolvePodmanFiles(files []PodmanFile, res Resolver) ([]KV, []Issue, error) {
	out := make([]KV, 0, len(files))
	var warns []Issue
	for _, f := range files {
		data, what := f.Data, f.StoreName
		switch {
		case f.Build != nil:
			what = f.Field
			b, w, err := BuildStore(*f.Build, res)
			if err != nil {
				return nil, nil, err
			}
			warns = append(warns, w...)
			data = string(b)
		case f.Source != "":
			what = fmt.Sprintf("%s %q", f.Field, f.Source)
			if res.ReadFile == nil {
				return nil, nil, fmt.Errorf("cannot read %s for podman's secret store (no file access)", what)
			}
			b, err := res.ReadFile(f.Source)
			if err != nil {
				return nil, nil, fmt.Errorf("reading %s for podman's secret store: %w", what, err)
			}
			data = string(b)
		}
		if err := podmanSecretSizeError(what, len(data)); err != nil {
			return nil, nil, err
		}
		out = append(out, KV{Key: f.StoreName, Val: data})
	}
	return out, warns, nil
}

// GeneratePodman parses+validates and renders the .container quadlet unit. The
// documents and stores it mounts are returned on the plan for deploy to load
// into podman's secret store; nothing is read from or written to the host here.
//
// extraAllowed threads deploy/remove's --allow-command values into the
// podman.command allowlist check; plain `generate podman` calls this with none.
func GeneratePodman(r Request, res Resolver, extraAllowed ...string) (plan PodmanPlan, errs, warns []Issue) {
	wfs, e, pissues, ewarns := parse(r, res)
	p := e.Podman
	if p == nil {
		pissues = append(pissues, Issue{File: fileEnv, Msg: "podman target requires a 'podman:' section in env.yaml"})
	}
	verrs, w := validate.Run(validate.Context{
		Workflows: wfs, Defaults: &e.Defaults, Image: e.Image, Timezone: e.Timezone, JavaOptions: e.JavaOptions, Podman: p, CheckPodman: true, Env: res.Env,
		AllowCommands: extraAllowed,
	})
	errs = append(pissues, verrs...)
	warns = append(ewarns, w...)
	if len(errs) > 0 {
		return PodmanPlan{}, errs, warns
	}

	statusPW, err := resolveStatusPassword(res)
	if err != nil {
		return PodmanPlan{}, []Issue{{File: fileEnv, Msg: err.Error()}}, warns
	}
	b, cw, err := build(wfs, &e.Defaults, statusPW)
	warns = append(warns, toIssues(cw)...)
	if err != nil {
		return PodmanPlan{}, []Issue{{File: fileEnv, Msg: err.Error()}}, warns
	}

	plan.Secrets = b.model.Secrets
	runAs, rerrs := runAsOf("podman", p.RunAs)
	if rerrs != nil {
		return PodmanPlan{}, rerrs, warns
	}
	plan.RunAs = runAs

	// Everything the unit mounts but the libs directory comes from podman's
	// secret store, in unit order: application.yml, the stores (at the path
	// application.yml names each one by), the status script, the logback config.
	// podman.base-dir, which used to hold the documents, is ignored.
	appSecret := PodmanSecretStoreName(p.Name, podmanAppYAMLSuffix)
	plan.Files = append(plan.Files, PodmanFile{StoreName: appSecret, Data: b.appYAML})
	var stores []podmangen.FileSecret
	for _, st := range b.model.Stores {
		name := PodmanStoreSecretName(p.Name, st.Name)
		plan.Files = append(plan.Files, PodmanFile{StoreName: name, Source: st.File, Field: storeField(st), Build: storeBuild(st, e.Defaults.TLS)})
		stores = append(stores, podmangen.FileSecret{StoreName: name, Target: tls.MountPath(st.Name)})
	}
	statusSecret := PodmanSecretStoreName(p.Name, podmanStatusSuffix)
	plan.Files = append(plan.Files, PodmanFile{StoreName: statusSecret, Data: statusscript.Render(e.Defaults.EffectiveManagementPort(), spec.StatusUserName)})
	logbackSecret := ""
	if sl := e.Defaults.Syslog; sl != nil {
		logbackSecret = PodmanSecretStoreName(p.Name, podmanLogbackSuffix)
		plan.Files = append(plan.Files, PodmanFile{StoreName: logbackSecret, Data: logback.XML(sl.Protocol)})
	}
	// Only the libs directory stays a host path, resolved against env.yaml like
	// every other; the stores are left out of targetMounts on purpose.
	_, lm := targetMounts(nil, p.Libs, res)
	plan.Mounts = hostMounts(nil, lm)
	plan.Service = PodmanServiceName(p.Name)
	in := podmangen.Input{
		Podman:  p,
		Syslog:  e.Defaults.Syslog,
		Secrets: podmanSecretRefs(p.Name, plan.Secrets),
		Stores:  stores,
		Libs:    toPodmanMount(lm),
		Instance: podmangen.Instance{
			Name:               p.Name,
			Image:              e.Image.Ref(),
			Timezone:           e.Timezone,
			AppYAMLSecret:      appSecret,
			MQTLS:              b.model.MQTLS,
			StatusScriptSecret: statusSecret,
			LogbackSecret:      logbackSecret,
			LeaderMode:         e.Defaults.LeaderElection.EffectiveMode(),
			JavaOptions:        e.JavaOptions,
			RunAs:              runAs,
		},
	}
	plan.Unit = podmangen.RenderQuadlet(in)
	return plan, nil, warns
}

// runAsOf parses a platform section's run-as for its renderer. validate.Run
// has already refused every value ParseRunAs rejects, so a failure here means a
// caller rendered without validating -- reported as an issue rather than
// rendered as the zero identity, which would be root.
func runAsOf(section, raw string) (spec.RunAsID, []Issue) {
	id, err := spec.ParseRunAs(raw)
	if err != nil {
		return spec.RunAsID{}, []Issue{{File: fileEnv, Msg: fmt.Sprintf("%s.run-as %q: %v", section, raw, err)}}
	}
	return id, nil
}

// ResolveCredentials turns a model's secret references into the ordered
// key/value pairs a platform's secret store needs: a literal contributes its own
// value, an -env reference the value of that host variable. The result is secret
// material -- callers hand it straight to the secret store and never log it.
//
// Errors name the stable secret name and the variable, never a value.
func ResolveCredentials(refs []consolidate.SecretRef, res Resolver) ([]KV, error) {
	out := make([]KV, 0, len(refs))
	for _, r := range refs {
		if r.EnvVar == "" {
			out = append(out, KV{Key: r.Stable, Val: r.Literal})
			continue
		}
		if res.Env == nil {
			return nil, fmt.Errorf("secret %s reads environment variable %s, but this command has no environment access", r.Stable, r.EnvVar)
		}
		v, ok := res.Env(r.EnvVar)
		if !ok {
			return nil, fmt.Errorf("secret %s: environment variable %s is not set; export it before deploying", r.Stable, r.EnvVar)
		}
		out = append(out, KV{Key: r.Stable, Val: v})
	}
	return out, nil
}

// resolveStatusPassword returns the password for the tool-reserved status
// actuator account (spec.StatusUserName): the operator's override from
// spec.StatusUserPasswordEnvVar when set and non-empty (validate.Run has
// already charset-checked it, so it is used verbatim), otherwise 16 random
// bytes hex-encoded into a 32-lowercase-hex-char literal. The result is
// rendered as a literal that the in-container status script (internal/
// statusscript) reads back out of application.yml, so it must be strong by
// default -- never a fixed or predictable fallback -- and it must stay
// stable within one generate call, which is why every build() call site
// resolves it exactly once and threads the same value through.
func resolveStatusPassword(res Resolver) (string, error) {
	if res.Env != nil {
		if v, ok := res.Env(spec.StatusUserPasswordEnvVar); ok && v != "" {
			return v, nil
		}
	}
	read := res.Rand
	if read == nil {
		read = func(b []byte) error { _, err := rand.Read(b); return err }
	}
	b := make([]byte, 16)
	if err := read(b); err != nil {
		return "", fmt.Errorf("generating the %s account password: %v; set %s to provide one explicitly", spec.StatusUserName, err, spec.StatusUserPasswordEnvVar)
	}
	return hex.EncodeToString(b), nil
}

// ---- config building ---------------------------------------------------------

// built is the consolidated model and its rendered application.yml. One folder
// yields exactly one of these: a connector instance runs at most
// validate.MaxWorkflows workflows, and validate rejects a folder holding more
// rather than splitting it.
type built struct {
	appYAML string
	model   *consolidate.Model
}

// ConfigImport is the Spring config-data import every generated application.yml
// carries. Credentials are mounted one file per name under spec.SecretsMountPath
// on all three platforms, so a single line makes the same config resolve
// everywhere -- which it must, since `generate config` takes no platform.
//
// It stays `optional:` because a config with no credentials mounts no such
// directory. The cost of that is worth knowing: a secrets directory that is
// missing, shadowed, or unreadable is not a startup failure either, so the
// connector comes up with placeholders unresolved and fails later instead.
const ConfigImport = "optional:configtree:" + spec.SecretsMountPath + "/"

// build consolidates the workflows and renders the application.yml.
// statusPassword is the already-resolved password for the reserved status
// account (see resolveStatusPassword); it flows straight into
// consolidate.Opts so it is threaded through, never recomputed.
func build(wfs []spec.Workflow, d *spec.Defaults, statusPassword string) (built, []string, error) {
	m, warns := consolidate.Build(wfs, d, consolidate.Opts{
		ConfigImport:   ConfigImport,
		StatusPassword: statusPassword,
	})
	// One mounted name holds one value, so two different credentials landing on
	// it means one of them would silently run with the other's value. Refuse
	// before anything is rendered; the operator renames the -env variable or the
	// connection it collides with.
	if err := secretConflictError(m.SecretConflicts); err != nil {
		return built{}, warns, err
	}
	return built{appYAML: render.Application(m), model: m}, warns, nil
}

// secretConflictError turns consolidate's collision records into the operator's
// error, or nil when there are none. It names *both* claiming positions, since
// the contested key alone does not say which field to edit -- a derived name
// appears nowhere in the spec.
//
// With spec.GeneratedNamePrefix reserved (validate rejects an -env variable
// inside it), a derived name and an operator's name can no longer meet. What is
// left is two derived names folding together: stableToken maps every
// non-alphanumeric run to one '_', so security.users "ops.1" and "ops-1" both
// reach _GEN_SECURITY_USER_OPS_1_PASSWORD. Hence "rename one of them" rather
// than advice about -env variables.
func secretConflictError(cs []consolidate.SecretConflict) error {
	if len(cs) == 0 {
		return nil
	}
	parts := make([]string, 0, len(cs))
	for _, c := range cs {
		parts = append(parts, fmt.Sprintf("%s (claimed by %s and %s)", c.Name, c.First, c.Second))
	}
	return fmt.Errorf("two different credentials are mounted under one name: %s. One mounted file holds one value, so rename one of them; names differing only in punctuation (\"ops.1\" vs \"ops-1\") fold to the same mount name", strings.Join(parts, "; "))
}

// ---- mount resolution --------------------------------------------------------

// mount is a resolved host->container bind mount, converted to the target
// renderer's own Mount type at the call site. Field is the env.yaml key the
// source comes from (StoreField(<name>) or MountLibs); Private marks a store
// holding a private key.
type mount struct {
	Field, Source, Target string
	Private               bool
}

// targetMounts resolves the stores the rendered connections select (each
// bind-mounted under its tls.stores name in the in-container store dir,
// tls.MountPath, which is where application.yml references it) and the libs
// directory into host->container mounts. The container side is fixed: the
// image fixes the directories, and the store's name the file in it; the spec
// only chooses the host source. Docker bind-mounts both; podman passes no
// stores (they come from its secret store) and takes only the libs mount.
//
// The stores are exactly consolidate's Model.Stores: application.yml references
// the mounted path for every one of them, and a store no connection selects is
// neither referenced nor mounted. libs stays opt-in -- it has a host dir to name.
func targetMounts(stores []consolidate.StoreUse, libs *spec.LibsMount, res Resolver) (sm []mount, lm *mount) {
	for _, st := range stores {
		if st.Created || st.File == "" {
			// No host file: a created store travels as a secret, and one held
			// only in a kubernetes Secret is refused for docker by validate.
			continue
		}
		sm = append(sm, mount{Field: StoreField(st.Name), Source: res.abs(st.File), Target: tls.MountPath(st.Name), Private: st.Private})
	}
	if libs != nil && libs.Dir != "" {
		lm = &mount{Field: MountLibs, Source: res.abs(libs.Dir), Target: spec.DefaultLibsMountPath}
	}
	return sm, lm
}

// hostMounts lists the host side of a plan's bind mounts, stores first.
func hostMounts(sm []mount, lm *mount) []HostMount {
	var out []HostMount
	for _, m := range sm {
		out = append(out, HostMount{Field: m.Field, Path: m.Source, Private: m.Private})
	}
	if lm != nil {
		out = append(out, HostMount{Field: lm.Field, Path: lm.Source})
	}
	return out
}

func toDockerMounts(ms []mount) []dockergen.Mount {
	out := make([]dockergen.Mount, 0, len(ms))
	for _, m := range ms {
		out = append(out, dockergen.Mount{Source: m.Source, Target: m.Target})
	}
	return out
}

func toDockerMount(m *mount) *dockergen.Mount {
	if m == nil {
		return nil
	}
	return &dockergen.Mount{Source: m.Source, Target: m.Target}
}

func toPodmanMount(m *mount) *podmangen.Mount {
	if m == nil {
		return nil
	}
	return &podmangen.Mount{Source: m.Source, Target: m.Target}
}

// PodmanServiceName is the systemd unit name a quadlet .container file
// generates for an instance. Shared with the status verb, which asks systemd
// for that unit's restart count: under quadlet a restart recreates the
// container, so the container's own counter reads 0 and only the unit knows how
// many times the instance has died.
func PodmanServiceName(name string) string { return name + ".service" }

// PodmanSecretStoreName is the name a credential is stored under in podman's
// secret store. That store is per-user and shared across every project on the
// host, so the container name namespaces it; the in-container file name comes
// from the mount target instead and stays the bare stable name.
func PodmanSecretStoreName(container, stable string) string {
	return container + "-" + stable
}

// podmanSecretRefs maps stable names onto the store-name/mount-target pairs the
// quadlet renderer emits.
func podmanSecretRefs(container string, refs []consolidate.SecretRef) []podmangen.SecretRef {
	out := make([]podmangen.SecretRef, 0, len(refs))
	for _, r := range refs {
		out = append(out, podmangen.SecretRef{
			StoreName: PodmanSecretStoreName(container, r.Stable),
			Target:    r.Stable,
		})
	}
	return out
}

// ---- parsing -----------------------------------------------------------------

func parse(r Request, res Resolver) (wfs []spec.Workflow, e *spec.Env, issues, warns []Issue) {
	files := append([]File(nil), r.Workflows...)
	// A workflow's id is its position here, in the order a directory listing
	// shows the names (spec.WorkflowFileLess), so the scan and an explicit
	// file list number the same files the same way.
	sort.Slice(files, func(i, j int) bool { return spec.WorkflowFileLess(files[i].Name, files[j].Name) })
	for _, f := range files {
		wf, err := spec.ParseWorkflow(f.Data, f.Name)
		if err != nil {
			issues = append(issues, Issue{File: f.Name, Msg: err.Error()})
			continue
		}
		wfs = append(wfs, *wf)
	}
	data := []byte(nil)
	if r.Env != nil {
		data = r.Env.Data
	}
	pe, err := spec.ParseEnv(data)
	if err != nil {
		issues = append(issues, Issue{File: fileEnv, Msg: err.Error()})
		pe, _ = spec.ParseEnv(nil) // fall back to a defaulted env so downstream code has a value
	}
	spec.Expand(spec.Expander{
		Lookup: res.Env,
		Warn: func(format string, a ...any) {
			warns = append(warns, Issue{File: fileEnv, Msg: fmt.Sprintf(format, a...)})
		},
	}, pe, wfs)
	return wfs, pe, issues, warns
}

// ---- secret resolution -------------------------------------------------------

// kubeApplyMax is the most one object can carry through kubectl's client-side
// `apply`, which is how deploy applies the manifest: apply records the whole
// object in its last-applied-configuration annotation, and the API server caps
// an object's annotations at 256 KiB -- well under the 1 MiB a Secret may hold.
const kubeApplyMax = 256 << 10

// kubeApplyAllowance is what the rest of the object -- its metadata, the keys,
// and the annotation's JSON re-encoding -- is allowed beside the base64 data.
const kubeApplyAllowance = 4096

// kubeStoresSizeError refuses a stores Secret kubectl apply would reject for
// size: total is its base64-encoded data, sizes each store's share of it.
func kubeStoresSizeError(total int, sizes []string) error {
	if total+kubeApplyAllowance <= kubeApplyMax {
		return nil
	}
	return fmt.Errorf("the stores Secret would hold %d bytes once base64-encoded (%s); with %d bytes for the rest of the object that is over the %d bytes kubectl apply carries in its last-applied-configuration annotation. Trim the stores to the certificates these connections need, or load a large store into a Secret yourself (kubectl create secret generic <name> --from-file=...) and name it on the store with existing: {secret: {name: <name>}}", total, strings.Join(sizes, ", "), kubeApplyAllowance, kubeApplyMax)
}

// resolveStores reads the files of the existing stores the rendered
// connections select, and builds the created ones (BuildStore), so the
// kubernetes stores Secret can carry their bytes. A store kept in an
// operator's Secret (existing.secret) is skipped: the pod reads it from there,
// and its file, if it names one, is never opened. Unlike credentials, these
// are files the cluster must hold verbatim, so they are base64-embedded rather
// than named. Each is keyed by its tls.stores name, the name it is mounted
// under, whatever the file is called. A set kubectl would refuse for size
// stops here, naming the stores, rather than at kubectl apply.
func resolveStores(stores []consolidate.StoreUse, defs spec.TLSConfig, res Resolver) ([]deploy.StoreFile, []Issue, error) {
	var own []consolidate.StoreUse
	for _, st := range stores {
		if !st.ClusterHeld() {
			own = append(own, st)
		}
	}
	if len(own) == 0 {
		return nil, nil, nil
	}
	if res.ReadFile == nil {
		return nil, nil, fmt.Errorf("cannot read store files (no file access)")
	}
	var out []deploy.StoreFile
	var warns []Issue
	total := 0
	var sizes []string
	for _, st := range own {
		var b []byte
		if sb := storeBuild(st, defs); sb != nil {
			built, w, err := BuildStore(*sb, res)
			if err != nil {
				return nil, nil, err
			}
			warns = append(warns, w...)
			b = built
		} else {
			read, err := res.ReadFile(st.File)
			if err != nil {
				return nil, nil, fmt.Errorf("reading %s %q: %v", StoreField(st.Name), st.File, err)
			}
			b = read
		}
		enc := b64(b)
		total += len(enc)
		sizes = append(sizes, fmt.Sprintf("%s %d bytes", st.Name, len(enc)))
		out = append(out, deploy.StoreFile{Name: st.Name, Base64: enc})
	}
	if err := kubeStoresSizeError(total, sizes); err != nil {
		return nil, nil, err
	}
	return out, warns, nil
}

func toIssues(ss []string) []Issue {
	out := make([]Issue, 0, len(ss))
	for _, s := range ss {
		out = append(out, Issue{Msg: s})
	}
	return out
}

func b64(b []byte) string { return base64.StdEncoding.EncodeToString(b) }
