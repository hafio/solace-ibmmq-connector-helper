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

	"gopkg.in/yaml.v3"

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

// Config parses+validates the request and returns one application.yml per
// connector instance (workflows are sharded at MaxWorkflowsPerInstance). A folder
// with <= that many workflows yields a single-element slice.
func Config(r Request, res Resolver) (outs []string, errs, warns []Issue) {
	wfs, e, pissues := parse(r)
	verrs, w := validate.Run(validate.Context{Workflows: wfs, Defaults: &e.Defaults, Env: res.Env})
	errs = append(pissues, verrs...)
	warns = w
	if len(errs) > 0 {
		return nil, errs, warns
	}
	shards, cw := buildShards(wfs, &e.Defaults, false)
	warns = append(warns, toIssues(cw)...)
	for _, s := range shards {
		outs = append(outs, s.appYAML)
	}
	return outs, nil, warns
}

// Validate runs every check for every section present and returns all findings.
func Validate(r Request, res Resolver) (errs, warns []Issue) {
	wfs, e, pissues := parse(r)
	valuesKeys, vErr := valuesFileKeys(e.Kubernetes, res)
	verrs, w := validate.Run(validate.Context{
		Workflows:      wfs,
		Defaults:       &e.Defaults,
		Kube:           e.Kubernetes,
		Docker:         e.Docker,
		Podman:         e.Podman,
		Deploy:         e.Kubernetes != nil,
		CheckDocker:    e.Docker != nil,
		CheckPodman:    e.Podman != nil,
		Env:            res.Env,
		ValuesFileKeys: valuesKeys,
	})
	errs = append(pissues, verrs...)
	if vErr != nil {
		errs = append(errs, *vErr)
	}
	return errs, w
}

// GenerateKubernetes parses+validates and returns the manifest set on success.
// Credentials/stores are resolved and embedded (stringData/base64), so this is
// the same rendered output the deploy path applies.
func GenerateKubernetes(r Request, res Resolver) (out string, errs, warns []Issue) {
	wfs, e, pissues := parse(r)
	k := e.Kubernetes
	if k == nil {
		pissues = append(pissues, Issue{File: fileEnv, Msg: "kubernetes target requires a 'kubernetes:' section in env.yaml"})
	}
	valuesKeys, vErr := valuesFileKeys(k, res)
	verrs, w := validate.Run(validate.Context{
		Workflows:      wfs,
		Defaults:       &e.Defaults,
		Kube:           k,
		Deploy:         true,
		Env:            res.Env,
		ValuesFileKeys: valuesKeys,
	})
	errs = append(pissues, verrs...)
	if vErr != nil {
		errs = append(errs, *vErr)
	}
	warns = w
	if len(errs) > 0 {
		return "", errs, warns
	}

	shards, cw := buildShards(wfs, &e.Defaults, true)
	warns = append(warns, toIssues(cw)...)

	in := deploy.Input{Kube: k, Defaults: &e.Defaults}
	if c := k.Secrets.Credentials; c != nil && c.Create != nil {
		kvs, err := resolveCred(c.Create, res)
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
	for i, s := range shards {
		name := instanceName(k.Deployment.Name, i, len(shards))
		in.Instances = append(in.Instances, deploy.Instance{Name: name, AppYAML: s.appYAML, Model: s.model})
	}
	return deploy.Render(in), nil, warns
}

// DockerPlan is the rendered compose plus the env-file base name the compose
// references (empty when there are no credentials). Credential values are not
// resolved here; the CLI writes the env-file (0600) at deploy time.
type DockerPlan struct {
	Compose     string
	EnvFileName string
}

// GenerateDocker parses+validates and renders the docker-compose.yml (with
// application.yml inlined under compose configs:). Store/libs host paths are
// resolved to absolute so the compose is portable regardless of cwd.
func GenerateDocker(r Request, res Resolver) (plan DockerPlan, errs, warns []Issue) {
	wfs, e, pissues := parse(r)
	d := e.Docker
	if d == nil {
		pissues = append(pissues, Issue{File: fileEnv, Msg: "docker target requires a 'docker:' section in env.yaml"})
	}
	verrs, w := validate.Run(validate.Context{
		Workflows: wfs, Defaults: &e.Defaults, Docker: d, CheckDocker: true, Env: res.Env,
	})
	errs = append(pissues, verrs...)
	warns = w
	if len(errs) > 0 {
		return DockerPlan{}, errs, warns
	}

	shards, cw := buildShards(wfs, &e.Defaults, true)
	warns = append(warns, toIssues(cw)...)

	sm, lm := targetMounts(e.Defaults.TLS, d.Stores, d.Libs, res)
	in := dockergen.Input{
		Docker:  d,
		EnvFile: credEnvFileName(d.Name, d.Secrets.Credentials),
		Stores:  toDockerMounts(sm),
		Libs:    toDockerMount(lm),
	}
	for i, s := range shards {
		in.Instances = append(in.Instances, dockergen.Instance{
			Name:    instanceName(d.Name, i, len(shards)),
			AppYAML: s.appYAML,
			MQTLS:   s.model.MQTLS,
		})
	}
	return DockerPlan{Compose: dockergen.Render(in), EnvFileName: in.EnvFile}, nil, warns
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
	Mode        string           // effective mode: run | quadlet
	RunScript   string           // set when Mode == run
	Units       []podmangen.Unit // set when Mode == quadlet
	AppYAMLs    []NamedDoc        // application.yml per instance (write to disk)
	EnvFileName string           // credential env-file base name ("" when none)
	Services    []string         // systemd service names (quadlet), e.g. name.service
}

// GeneratePodman parses+validates and renders either a `podman run` script
// (mode run) or one .container quadlet unit per instance (mode quadlet, or when
// opts.ForceQuadlet is set). application.yml documents are returned separately
// for the caller to write to disk.
func GeneratePodman(r Request, res Resolver, opts PodmanOpts) (plan PodmanPlan, errs, warns []Issue) {
	wfs, e, pissues := parse(r)
	p := e.Podman
	if p == nil {
		pissues = append(pissues, Issue{File: fileEnv, Msg: "podman target requires a 'podman:' section in env.yaml"})
	}
	verrs, w := validate.Run(validate.Context{
		Workflows: wfs, Defaults: &e.Defaults, Podman: p, CheckPodman: true, Env: res.Env,
	})
	errs = append(pissues, verrs...)
	warns = w
	if len(errs) > 0 {
		return PodmanPlan{}, errs, warns
	}

	shards, cw := buildShards(wfs, &e.Defaults, true)
	warns = append(warns, toIssues(cw)...)

	plan.Mode = p.Mode
	if opts.ForceQuadlet {
		plan.Mode = spec.PodmanModeQuadlet
	}
	plan.EnvFileName = credEnvFileName(p.Name, p.Secrets.Credentials)

	sm, lm := targetMounts(e.Defaults.TLS, p.Stores, p.Libs, res)
	in := podmangen.Input{Podman: p, Stores: toPodmanMounts(sm), Libs: toPodmanMount(lm)}
	// The quadlet EnvironmentFile= line references the env-file by an absolute host
	// path. A tool-created env-file is written into BaseDir (the quadlet dir) at
	// deploy, so reference it there; an existing env-file is the user's own and
	// lives at its env.yaml-resolved location -- reference it like stores/libs, via
	// res.abs, never joined onto BaseDir.
	if c := p.Secrets.Credentials; c != nil {
		switch {
		case c.Existing != "":
			in.EnvFile = res.abs(c.Existing)
		case c.Create != nil:
			in.EnvFile = pathIn(opts.BaseDir, plan.EnvFileName)
		}
	}
	for i, s := range shards {
		name := instanceName(p.Name, i, len(shards))
		appName := name + "-application.yml"
		plan.AppYAMLs = append(plan.AppYAMLs, NamedDoc{Name: appName, Data: s.appYAML})
		plan.Services = append(plan.Services, name+".service")
		in.Instances = append(in.Instances, podmangen.Instance{
			Name:        name,
			AppYAMLPath: pathIn(opts.BaseDir, appName),
			MQTLS:       s.model.MQTLS,
		})
	}
	if plan.Mode == spec.PodmanModeQuadlet {
		plan.Units = podmangen.RenderQuadlet(in)
	} else {
		plan.RunScript = podmangen.RenderRunScript(in)
	}
	return plan, nil, warns
}

// ResolveCredentials resolves a docker/podman credentials block to ordered KV
// pairs for the env-file. A nil block, or one that names an existing env-file
// (Existing), returns no pairs.
func ResolveCredentials(creds *spec.CredentialsSecret, res Resolver) ([]KV, error) {
	if creds == nil || creds.Create == nil {
		return nil, nil
	}
	return resolveCred(creds.Create, res)
}

// EnvFileContent renders ordered KV pairs as env-file lines (KEY=VALUE). Written
// 0600 by the caller; never logged.
func EnvFileContent(kvs []KV) string {
	var b strings.Builder
	for _, kv := range kvs {
		b.WriteString(kv.Key)
		b.WriteByte('=')
		b.WriteString(kv.Val)
		b.WriteByte('\n')
	}
	return b.String()
}

// ---- shared shard building ---------------------------------------------------

// shard is one connector instance's consolidated model and rendered config.
type shard struct {
	appYAML string
	model   *consolidate.Model
}

// shardWorkflows splits wfs into fill-to-N chunks preserving order (instance 1 =
// workflows 0..N-1, instance 2 = N..2N-1, ...). Always returns at least one chunk
// (an empty folder yields one empty chunk, matching the single-instance path).
func shardWorkflows(wfs []spec.Workflow) [][]spec.Workflow {
	per := validate.MaxWorkflowsPerInstance
	if len(wfs) <= per {
		return [][]spec.Workflow{wfs}
	}
	var out [][]spec.Workflow
	for i := 0; i < len(wfs); i += per {
		end := i + per
		if end > len(wfs) {
			end = len(wfs)
		}
		out = append(out, wfs[i:end])
	}
	return out
}

// buildShards builds one Model + application.yml per chunk. When there is more
// than one instance, the leader-election coordination queue is suffixed per
// instance (<queue>-<n>) so the independent connector clusters do not contend for
// a single election queue. The suffix is applied before rendering.
func buildShards(wfs []spec.Workflow, d *spec.Defaults, mountStores bool) ([]shard, []string) {
	chunks := shardWorkflows(wfs)
	n := len(chunks)
	var shards []shard
	var warns []string
	for i, c := range chunks {
		m, cw := consolidate.Build(c, d, mountStores)
		warns = append(warns, cw...)
		if n > 1 && m.LeaderElection != nil && m.LeaderElection.Queue != "" {
			m.LeaderElection.Queue = fmt.Sprintf("%s-%d", m.LeaderElection.Queue, i+1)
		}
		shards = append(shards, shard{appYAML: render.Application(m), model: m})
	}
	return shards, warns
}

// instanceName is the base name for one instance, suffixed -N (1-based) only when
// there is more than one instance.
func instanceName(base string, i, n int) string {
	if n > 1 {
		return fmt.Sprintf("%s-%d", base, i+1)
	}
	return base
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
			sm = append(sm, mount{Source: res.abs(st.File), Target: spec.DefaultStoresMountPath + "/" + baseName(st.File)})
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

// credEnvFileName is the env-file the compose/unit references: <name>.env for a
// created secret, the Existing value when one is named, or "" when there are no
// credentials.
func credEnvFileName(name string, creds *spec.CredentialsSecret) string {
	if creds == nil {
		return ""
	}
	if creds.Existing != "" {
		return creds.Existing
	}
	if creds.Create != nil {
		return name + ".env"
	}
	return ""
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

func parse(r Request) (wfs []spec.Workflow, e *spec.Env, issues []Issue) {
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
	return wfs, pe, issues
}

// ---- secret resolution -------------------------------------------------------

func resolveCred(c *spec.CredCreate, res Resolver) ([]deploy.KV, error) {
	switch c.Source {
	case spec.SourceEnv:
		var out []deploy.KV
		for _, v := range c.Variables {
			val := ""
			if res.Env != nil {
				got, ok := res.Env(v)
				if !ok {
					return nil, fmt.Errorf("credentials variable %q is not set in the environment", v)
				}
				val = got
			}
			out = append(out, deploy.KV{Key: v, Val: val})
		}
		return out, nil
	case spec.SourceFile:
		if res.ReadFile == nil {
			return nil, fmt.Errorf("cannot read values-file %q (no file access)", c.ValuesFile)
		}
		data, err := res.ReadFile(c.ValuesFile)
		if err != nil {
			return nil, fmt.Errorf("reading values-file %q: %v", c.ValuesFile, err)
		}
		return parseValues(data)
	default:
		return nil, fmt.Errorf("unknown credentials source %q", c.Source)
	}
}

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
		out = append(out, deploy.StoreFile{Name: baseName(s.File), Base64: b64(b)})
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

func valuesFileKeys(k *spec.Kubernetes, res Resolver) (map[string]bool, *Issue) {
	if k == nil || k.Secrets.Credentials == nil || k.Secrets.Credentials.Create == nil {
		return nil, nil
	}
	c := k.Secrets.Credentials.Create
	if c.Source != spec.SourceFile || res.ReadFile == nil {
		return nil, nil
	}
	data, err := res.ReadFile(c.ValuesFile)
	if err != nil {
		return nil, &Issue{File: fileEnv, Msg: fmt.Sprintf("reading values-file %q: %v", c.ValuesFile, err)}
	}
	kvs, err := parseValues(data)
	if err != nil {
		return nil, &Issue{File: fileEnv, Msg: fmt.Sprintf("parsing values-file %q: %v", c.ValuesFile, err)}
	}
	keys := map[string]bool{}
	for _, kv := range kvs {
		keys[kv.Key] = true
	}
	return keys, nil
}

// parseValues reads a values-file as either dotenv (KEY=VALUE, order-preserving)
// or a YAML mapping, returning ordered key/value pairs.
func parseValues(data []byte) ([]deploy.KV, error) {
	text := string(data)
	if looksDotenv(text) {
		var out []deploy.KV
		for _, ln := range strings.Split(text, "\n") {
			ln = strings.TrimSpace(ln)
			if ln == "" || strings.HasPrefix(ln, "#") {
				continue
			}
			i := strings.IndexByte(ln, '=')
			if i < 0 {
				continue
			}
			out = append(out, deploy.KV{Key: strings.TrimSpace(ln[:i]), Val: strings.TrimSpace(ln[i+1:])})
		}
		return out, nil
	}
	var node yaml.Node
	if err := yaml.Unmarshal(data, &node); err != nil {
		return nil, err
	}
	var m *yaml.Node
	if node.Kind == yaml.DocumentNode && len(node.Content) > 0 {
		m = node.Content[0]
	} else {
		m = &node
	}
	if m.Kind != yaml.MappingNode {
		return nil, fmt.Errorf("values-file must be a mapping or KEY=VALUE lines")
	}
	var out []deploy.KV
	for i := 0; i+1 < len(m.Content); i += 2 {
		out = append(out, deploy.KV{Key: m.Content[i].Value, Val: m.Content[i+1].Value})
	}
	return out, nil
}

func looksDotenv(text string) bool {
	sawLine := false
	for _, ln := range strings.Split(text, "\n") {
		ln = strings.TrimSpace(ln)
		if ln == "" || strings.HasPrefix(ln, "#") {
			continue
		}
		sawLine = true
		eq := strings.IndexByte(ln, '=')
		col := strings.IndexByte(ln, ':')
		// A dotenv line has '=' before any ':' (YAML mappings use ': ').
		if eq < 0 || (col >= 0 && col < eq) {
			return false
		}
	}
	return sawLine
}

func toIssues(ss []string) []Issue {
	out := make([]Issue, 0, len(ss))
	for _, s := range ss {
		out = append(out, Issue{Msg: s})
	}
	return out
}

// baseName returns the final path element, splitting on both '/' and '\' so it
// resolves the same regardless of the host OS the CLI runs on.
func baseName(p string) string {
	p = strings.ReplaceAll(p, "\\", "/")
	if i := strings.LastIndexByte(p, '/'); i >= 0 {
		return p[i+1:]
	}
	return p
}

func b64(b []byte) string { return base64.StdEncoding.EncodeToString(b) }
