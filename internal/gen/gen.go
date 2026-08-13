// Package gen is the in-memory generation core behind the CLI. It ties together
// parse -> validate -> consolidate -> render for every target (application.yml,
// kubernetes manifests, docker compose, podman run/quadlet). All I/O (folder
// scan, env lookups, file reads, path resolution) is injected by callers so the
// package stays pure and testable.
package gen

import (
	"encoding/base64"
	"fmt"
	"sort"
	"strings"

	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/consolidate"
	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/deploy"
	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/dockergen"
	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/podmangen"
	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/render"
	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/spec"
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
	b, cw := build(wfs, &e.Defaults, false)
	return b.appYAML, nil, append(warns, toIssues(cw)...)
}

// Validate runs every check for every section present and returns all findings.
func Validate(r Request, res Resolver) (errs, warns []Issue) {
	wfs, e, pissues, ewarns := parse(r, res)
	verrs, w := validate.Run(validate.Context{
		Workflows:       wfs,
		Defaults:        &e.Defaults,
		Kube:            e.Kubernetes,
		Docker:          e.Docker,
		Podman:          e.Podman,
		CheckKubernetes: e.Kubernetes != nil,
		CheckDocker:     e.Docker != nil,
		CheckPodman:     e.Podman != nil,
		Env:             res.Env,
	})
	return append(pissues, verrs...), append(ewarns, w...)
}

// GenerateKubernetes parses+validates and returns the manifest set on success.
// Credentials/stores are resolved and embedded (stringData/base64), so this is
// the same rendered output the deploy path applies.
func GenerateKubernetes(r Request, res Resolver) (out string, errs, warns []Issue) {
	wfs, e, pissues, ewarns := parse(r, res)
	k := e.Kubernetes
	if k == nil {
		pissues = append(pissues, Issue{File: fileEnv, Msg: "kubernetes target requires a 'kubernetes:' section in env.yaml"})
	}
	verrs, w := validate.Run(validate.Context{
		Workflows:       wfs,
		Defaults:        &e.Defaults,
		Kube:            k,
		CheckKubernetes: true,
		Env:             res.Env,
	})
	errs = append(pissues, verrs...)
	warns = append(ewarns, w...)
	if len(errs) > 0 {
		return "", errs, warns
	}

	b, cw := build(wfs, &e.Defaults, true)
	warns = append(warns, toIssues(cw)...)

	in := deploy.Input{Kube: k, Defaults: &e.Defaults}
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
	in.Instance = deploy.Instance{Name: k.Deployment.Name, AppYAML: b.appYAML, Model: b.model}
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
func GenerateDocker(r Request, res Resolver) (plan DockerPlan, errs, warns []Issue) {
	wfs, e, pissues, ewarns := parse(r, res)
	d := e.Docker
	if d == nil {
		pissues = append(pissues, Issue{File: fileEnv, Msg: "docker target requires a 'docker:' section in env.yaml"})
	}
	verrs, w := validate.Run(validate.Context{
		Workflows: wfs, Defaults: &e.Defaults, Docker: d, CheckDocker: true, Env: res.Env,
	})
	errs = append(pissues, verrs...)
	warns = append(ewarns, w...)
	if len(errs) > 0 {
		return DockerPlan{}, errs, warns
	}

	b, cw := build(wfs, &e.Defaults, true)
	warns = append(warns, toIssues(cw)...)

	sm, lm := targetMounts(e.Defaults.TLS, d.Stores, d.Libs, res)
	in := dockergen.Input{
		Docker:   d,
		Secrets:  stableNames(b.model.Secrets),
		Stores:   toDockerMounts(sm),
		Libs:     toDockerMount(lm),
		Instance: dockergen.Instance{Name: d.Name, AppYAML: b.appYAML, MQTLS: b.model.MQTLS},
	}
	return DockerPlan{Compose: dockergen.Render(in), Secrets: b.model.Secrets}, nil, warns
}

// stableNames projects secret references down to the names an artifact declares.
func stableNames(refs []consolidate.SecretRef) []string {
	out := make([]string, 0, len(refs))
	for _, r := range refs {
		out = append(out, r.Stable)
	}
	return out
}

// NamedDoc is one on-disk document (podman writes application.yml files to the
// unit/script directory because a container cannot inline file content).
type NamedDoc struct {
	Name string
	Data string
}

// PodmanOpts tunes podman generation. ForceQuadlet overrides the config mode
// (deploy/delete are always quadlet). BaseDir, when set, makes the on-disk file
// paths (application.yml, env-file) absolute under it so systemd quadlet units
// reference real locations; empty BaseDir keeps base names for a preview.
type PodmanOpts struct {
	ForceQuadlet bool
	BaseDir      string
}

// PodmanPlan carries whichever artifact the effective mode produced plus the
// on-disk material a deploy must write before activating the units.
type PodmanPlan struct {
	Mode      string                  // effective mode: run | quadlet
	RunScript string                  // set when Mode == run
	Unit      podmangen.Unit          // set when Mode == quadlet
	AppYAML   NamedDoc                // application.yml (write to disk)
	Secrets   []consolidate.SecretRef // credentials to place in podman's secret store
	Service   string                  // systemd service name (quadlet), e.g. name.service
}

// GeneratePodman parses+validates and renders either a `podman run` script
// (mode run) or one .container quadlet unit per instance (mode quadlet, or when
// opts.ForceQuadlet is set). application.yml documents are returned separately
// for the caller to write to disk.
func GeneratePodman(r Request, res Resolver, opts PodmanOpts) (plan PodmanPlan, errs, warns []Issue) {
	wfs, e, pissues, ewarns := parse(r, res)
	p := e.Podman
	if p == nil {
		pissues = append(pissues, Issue{File: fileEnv, Msg: "podman target requires a 'podman:' section in env.yaml"})
	}
	verrs, w := validate.Run(validate.Context{
		Workflows: wfs, Defaults: &e.Defaults, Podman: p, CheckPodman: true, Env: res.Env,
	})
	errs = append(pissues, verrs...)
	warns = append(ewarns, w...)
	if len(errs) > 0 {
		return PodmanPlan{}, errs, warns
	}

	b, cw := build(wfs, &e.Defaults, true)
	warns = append(warns, toIssues(cw)...)

	plan.Mode = p.Mode
	if opts.ForceQuadlet {
		plan.Mode = spec.PodmanModeQuadlet
	}
	plan.Secrets = b.model.Secrets

	sm, lm := targetMounts(e.Defaults.TLS, p.Stores, p.Libs, res)
	appName := p.Name + "-application.yml"
	plan.AppYAML = NamedDoc{Name: appName, Data: b.appYAML}
	plan.Service = p.Name + ".service"
	in := podmangen.Input{
		Podman:   p,
		Secrets:  podmanSecretRefs(p.Name, plan.Secrets),
		Stores:   toPodmanMounts(sm),
		Libs:     toPodmanMount(lm),
		Instance: podmangen.Instance{Name: p.Name, AppYAMLPath: pathIn(opts.BaseDir, appName), MQTLS: b.model.MQTLS},
	}
	if plan.Mode == spec.PodmanModeQuadlet {
		plan.Unit = podmangen.RenderQuadlet(in)
	} else {
		plan.RunScript = podmangen.RenderRunScript(in)
	}
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
// carries. Credentials are mounted one-file-per-name under this directory on all
// three platforms, so a single line makes the same config resolve everywhere.
const ConfigImport = "optional:configtree:/run/secrets/"

// build consolidates the workflows and renders the application.yml.
func build(wfs []spec.Workflow, d *spec.Defaults, mountStores bool) (built, []string) {
	m, warns := consolidate.Build(wfs, d, consolidate.Opts{
		MountStores:  mountStores,
		ConfigImport: ConfigImport,
	})
	return built{appYAML: render.Application(m), model: m}, warns
}

// ---- mount resolution --------------------------------------------------------

// mount is a resolved host->container bind mount, converted to the target
// renderer's own Mount type at the call site.
type mount struct{ Source, Target string }

// targetMounts resolves the store files (each tls.*.file bind-mounted onto the
// fixed in-container store dir spec.DefaultStoresMountPath, matching where
// application.yml references them) and the libs directory into host->container
// mounts. Both are nil/absent unless the section opted in.
func targetMounts(tls spec.TLSConfig, stores *spec.StoresMount, libs *spec.LibsMount, res Resolver) (sm []mount, lm *mount) {
	if stores != nil {
		for _, st := range []*spec.Store{tls.Truststore, tls.Keystore} {
			if st == nil || st.File == "" {
				continue
			}
			sm = append(sm, mount{Source: res.abs(st.File), Target: spec.DefaultStoresMountPath + "/" + spec.BaseName(st.File)})
		}
	}
	if libs != nil && libs.Dir != "" {
		lm = &mount{Source: res.abs(libs.Dir), Target: libs.MountPath}
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

func toPodmanMounts(ms []mount) []podmangen.Mount {
	out := make([]podmangen.Mount, 0, len(ms))
	for _, m := range ms {
		out = append(out, podmangen.Mount{Source: m.Source, Target: m.Target})
	}
	return out
}

func toPodmanMount(m *mount) *podmangen.Mount {
	if m == nil {
		return nil
	}
	return &podmangen.Mount{Source: m.Source, Target: m.Target}
}

// PodmanSecretStoreName is the name a credential is stored under in podman's
// secret store. That store is per-user and shared across every project on the
// host, so the container name namespaces it; the in-container file name comes
// from the mount target instead and stays the bare stable name.
func PodmanSecretStoreName(container, stable string) string {
	return container + "-" + stable
}

// podmanSecretRefs maps stable names onto the store-name/mount-target pairs the
// quadlet and run-script renderers emit.
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

// pathIn joins base and name with a forward slash (the on-target separator for
// docker/podman). An empty base yields the bare name for a portable preview.
func pathIn(base, name string) string {
	if base == "" {
		return name
	}
	return strings.TrimRight(base, "/\\") + "/" + name
}

// ---- parsing -----------------------------------------------------------------

func parse(r Request, res Resolver) (wfs []spec.Workflow, e *spec.Env, issues, warns []Issue) {
	files := append([]File(nil), r.Workflows...)
	sort.Slice(files, func(i, j int) bool { return files[i].Name < files[j].Name })
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
