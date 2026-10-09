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
	"sort"
	"strings"

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

// Config parses+validates the request and returns the application.yml. One
// folder is one connector instance: validate rejects a folder holding more than
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
	b, cw, err := build(wfs, &e.Defaults, false, statusPW)
	warns = append(warns, toIssues(cw)...)
	if err != nil {
		return "", []Issue{{File: fileEnv, Msg: err.Error()}}, warns
	}
	return b.appYAML, nil, warns
}

// Validate runs every check for every section present and returns all findings.
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
	// credential check follows: catch it while authoring, not mid-deploy.
	//
	// Only on an otherwise-clean spec: consolidate assumes validated input.
	// The status password is a placeholder because it is rendered as a literal
	// and never becomes a mount name, so it cannot affect the result, and the
	// rendered document is discarded.
	if len(errs) == 0 {
		if _, _, berr := build(wfs, &e.Defaults, false, statusPasswordPlaceholder); berr != nil {
			errs = append(errs, Issue{File: fileEnv, Msg: berr.Error()})
		}
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
	// namespace and everything else living in it.
	Teardown bool
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
	b, cw, err := build(wfs, &e.Defaults, true, statusPW)
	warns = append(warns, toIssues(cw)...)
	if err != nil {
		return "", []Issue{{File: fileEnv, Msg: err.Error()}}, warns
	}

	in := deploy.Input{Kube: k, Defaults: &e.Defaults, Syslog: e.Defaults.Syslog, Teardown: opts.Teardown}
	if c := k.Secrets.Credentials; c != nil && c.Create != nil {
		kvs, err := ResolveCredentials(b.model.Secrets, res)
		if err != nil {
			return "", []Issue{{File: fileEnv, Msg: err.Error()}}, warns
		}
		in.CredKVs = kvs
	}
	if s := k.Secrets.Stores; s != nil && s.Create != nil {
		files, err := resolveStores(&e.Defaults, res)
		if err != nil {
			return "", []Issue{{File: fileEnv, Msg: err.Error()}}, warns
		}
		in.Stores = files
	}
	if name, created := k.ImagePullSecretName(); name != "" {
		ps, err := resolvePullSecret(name, created, e.Image, res)
		if err != nil {
			return "", []Issue{{File: fileEnv, Msg: err.Error()}}, warns
		}
		in.ImagePull = ps
	}
	in.Instance = deploy.Instance{
		Name:         k.Deployment.Name,
		Image:        e.Image.Ref(),
		Timezone:     e.Timezone,
		AppYAML:      b.appYAML,
		StatusScript: statusscript.Render(deploy.ManagementPort(in), spec.StatusUserName),
		Model:        b.model,
		JavaOptions:  e.JavaOptions,
	}
	return deploy.Render(in), nil, warns
}

// DockerPlan is the rendered compose plus the secret references it declares.
// Values are not resolved here: the compose file names secrets only, and the CLI
// resolves the values into the `docker compose` child environment at deploy time
// so nothing secret is ever written to disk.
type DockerPlan struct {
	Compose string
	Secrets []consolidate.SecretRef
}

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
	b, cw, err := build(wfs, &e.Defaults, true, statusPW)
	warns = append(warns, toIssues(cw)...)
	if err != nil {
		return DockerPlan{}, []Issue{{File: fileEnv, Msg: err.Error()}}, warns
	}

	sm, lm := targetMounts(e.Defaults.TLS, d.Libs, res)
	in := dockergen.Input{
		Docker:  d,
		Syslog:  e.Defaults.Syslog,
		Secrets: stableNames(b.model.Secrets),
		Stores:  toDockerMounts(sm),
		Libs:    toDockerMount(lm),
		Instance: dockergen.Instance{
			Name:         d.Name,
			Image:        e.Image.Ref(),
			Timezone:     e.Timezone,
			AppYAML:      b.appYAML,
			MQTLS:        b.model.MQTLS,
			StatusScript: statusscript.Render(e.Defaults.EffectiveManagementPort(), spec.StatusUserName),
			LeaderMode:   e.Defaults.LeaderElection.EffectiveMode(),
			JavaOptions:  e.JavaOptions,
		},
	}
	return DockerPlan{Compose: dockergen.Render(in), Secrets: b.model.Secrets}, nil, warns
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
// TLS store, named by Source and read only when deploy loads it
// (ResolvePodmanFiles) -- generate and remove never open a store.
type PodmanFile struct {
	StoreName string // the secret's name in podman's store: <name>-<role>
	Data      string // a rendered document; empty for a store
	Source    string // a store's tls.*.file as written in env.yaml; empty for a document
	Field     string // a store's env.yaml key, e.g. tls.truststore.file, for errors
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
}

// The secret-store names of the files a podman unit mounts, one per role, each
// appended to the container name by PodmanSecretStoreName. Every suffix holds
// a '-' or a '.', which no credential's stable name can (those match
// [A-Za-z_][A-Za-z0-9_]*), so a file never takes a credential's name. The names
// follow the role rather than the file, so a renamed store or a dropped syslog
// block strands nothing: deploy deletes the ones it no longer mounts.
const (
	podmanAppYAMLSuffix    = "application.yml"
	podmanTruststoreSuffix = "tls-truststore"
	podmanKeystoreSuffix   = "tls-keystore"
	podmanStatusSuffix     = "status-script"
	podmanLogbackSuffix    = logback.FileName
)

// PodmanFileSecretNames lists every file secret an instance can own, mounted by
// its current spec or not: deploy deletes the ones it no longer mounts, and
// remove deletes them all.
func PodmanFileSecretNames(container string) []string {
	suffixes := []string{podmanAppYAMLSuffix, podmanTruststoreSuffix, podmanKeystoreSuffix, podmanStatusSuffix, podmanLogbackSuffix}
	out := make([]string, 0, len(suffixes))
	for _, s := range suffixes {
		out = append(out, PodmanSecretStoreName(container, s))
	}
	return out
}

// podmanSecretMax is the most data one podman secret holds: containers/common
// refuses empty data and anything of 512000 bytes or more.
const podmanSecretMax = 511999

// ResolvePodmanFiles turns a plan's files into the name/value pairs deploy
// loads into podman's secret store: a rendered document as it stands, a TLS
// store read from its tls.*.file (resolved against env.yaml, as resolveStores
// reads it for kubernetes). Every value is held to podman's size limit here, so
// a file podman would refuse stops the deploy before any secret is created.
// The result is secret material -- callers hand it straight to the store.
//
// Errors name the file or secret and its size, never its content.
func ResolvePodmanFiles(files []PodmanFile, res Resolver) ([]KV, error) {
	out := make([]KV, 0, len(files))
	for _, f := range files {
		data, what := f.Data, f.StoreName
		if f.Source != "" {
			what = fmt.Sprintf("%s %q", f.Field, f.Source)
			if res.ReadFile == nil {
				return nil, fmt.Errorf("cannot read %s for podman's secret store (no file access)", what)
			}
			b, err := res.ReadFile(f.Source)
			if err != nil {
				return nil, fmt.Errorf("reading %s for podman's secret store: %w", what, err)
			}
			data = string(b)
		}
		switch {
		case len(data) == 0:
			return nil, fmt.Errorf("%s is empty, and podman's secret store cannot hold an empty secret; point it at the real file", what)
		case len(data) > podmanSecretMax:
			return nil, fmt.Errorf("%s is %d bytes, over the %d bytes podman's secret store holds per secret; trim it to what this connector needs (for a truststore, the CA certificates of the brokers and queue managers it connects to)", what, len(data), podmanSecretMax)
		}
		out = append(out, KV{Key: f.StoreName, Val: data})
	}
	return out, nil
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
	b, cw, err := build(wfs, &e.Defaults, true, statusPW)
	warns = append(warns, toIssues(cw)...)
	if err != nil {
		return PodmanPlan{}, []Issue{{File: fileEnv, Msg: err.Error()}}, warns
	}

	plan.Secrets = b.model.Secrets

	// Everything the unit mounts but the libs directory comes from podman's
	// secret store, in unit order: application.yml, the stores (at the path
	// application.yml names each one by), the status script, the logback config.
	// podman.base-dir, which used to hold the documents, is ignored.
	appSecret := PodmanSecretStoreName(p.Name, podmanAppYAMLSuffix)
	plan.Files = append(plan.Files, PodmanFile{StoreName: appSecret, Data: b.appYAML})
	var stores []podmangen.FileSecret
	for _, st := range []struct {
		field, suffix string
		store         *spec.Store
	}{{"tls.truststore.file", podmanTruststoreSuffix, e.Defaults.TLS.Truststore}, {"tls.keystore.file", podmanKeystoreSuffix, e.Defaults.TLS.Keystore}} {
		if st.store == nil || st.store.File == "" {
			continue
		}
		name := PodmanSecretStoreName(p.Name, st.suffix)
		plan.Files = append(plan.Files, PodmanFile{StoreName: name, Source: st.store.File, Field: st.field})
		stores = append(stores, podmangen.FileSecret{StoreName: name, Target: tls.MountPath(st.store.File)})
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
	_, lm := targetMounts(spec.TLSConfig{}, p.Libs, res)
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
		},
	}
	plan.Unit = podmangen.RenderQuadlet(in)
	return plan, nil, warns
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
func build(wfs []spec.Workflow, d *spec.Defaults, mountStores bool, statusPassword string) (built, []string, error) {
	m, warns := consolidate.Build(wfs, d, consolidate.Opts{
		MountStores:    mountStores,
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
// renderer's own Mount type at the call site.
type mount struct{ Source, Target string }

// targetMounts resolves the store files (each tls.*.file bind-mounted onto the
// fixed in-container store dir spec.DefaultStoresMountPath, matching where
// application.yml references them) and the libs directory into host->container
// mounts. Both container-side paths are constants: the image fixes them, and the
// spec only chooses the host source. Docker bind-mounts both; podman passes no
// stores (they come from its secret store) and takes only the libs mount.
//
// The stores are not gated on an opt-in: application.yml always references the
// mounted path for docker (consolidate is called with MountStores true), so a
// configured store that was not mounted left the config pointing at a file
// nothing supplied. A store with no file set is skipped, so a spec without TLS
// still yields no store mounts. libs stays opt-in -- it has a host dir to name.
func targetMounts(tls spec.TLSConfig, libs *spec.LibsMount, res Resolver) (sm []mount, lm *mount) {
	for _, st := range []*spec.Store{tls.Truststore, tls.Keystore} {
		if st == nil || st.File == "" {
			continue
		}
		sm = append(sm, mount{Source: res.abs(st.File), Target: spec.DefaultStoresMountPath + "/" + spec.BaseName(st.File)})
	}
	if libs != nil && libs.Dir != "" {
		lm = &mount{Source: res.abs(libs.Dir), Target: spec.DefaultLibsMountPath}
	}
	return sm, lm
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

// resolveStores reads the truststore/keystore files so the kubernetes stores
// Secret can carry their bytes. Unlike credentials, these are files the cluster
// must hold verbatim, so they are base64-embedded rather than named.
func resolveStores(d *spec.Defaults, res Resolver) ([]deploy.StoreFile, error) {
	if res.ReadFile == nil {
		return nil, fmt.Errorf("cannot read store files (no file access)")
	}
	var out []deploy.StoreFile
	add := func(s *spec.Store) error {
		if s == nil || s.File == "" {
			return nil
		}
		b, err := res.ReadFile(s.File)
		if err != nil {
			return fmt.Errorf("reading store %q: %v", s.File, err)
		}
		out = append(out, deploy.StoreFile{Name: spec.BaseName(s.File), Base64: b64(b)})
		return nil
	}
	if err := add(d.TLS.Truststore); err != nil {
		return nil, err
	}
	if err := add(d.TLS.Keystore); err != nil {
		return nil, err
	}
	return out, nil
}

func toIssues(ss []string) []Issue {
	out := make([]Issue, 0, len(ss))
	for _, s := range ss {
		out = append(out, Issue{Msg: s})
	}
	return out
}

func b64(b []byte) string { return base64.StdEncoding.EncodeToString(b) }
