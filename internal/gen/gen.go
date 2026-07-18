// Package gen is the in-memory generation core behind the CLI. It ties together parse -> validate -> consolidate ->
// render/deploy; all I/O (folder scan, env, file reads) is injected by callers.
package gen

import (
	"encoding/base64"
	"fmt"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-gen/internal/consolidate"
	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-gen/internal/deploy"
	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-gen/internal/render"
	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-gen/internal/spec"
	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-gen/internal/validate"
)

// File is one named input document.
type File struct {
	Name string
	Data []byte
}

// Request is a full set of inputs (workflow files + optional defaults/kubernetes).
type Request struct {
	Workflows  []File
	Defaults   *File
	Kubernetes *File
}

// Resolver injects the environment-dependent bits deploy needs.
type Resolver struct {
	// Env looks up an environment variable (CLI: os.LookupEnv; nil skips checks).
	Env func(string) (string, bool)
	// ReadFile reads a values-file or .jks referenced by config (from disk).
	// Nil means no file access.
	ReadFile func(string) ([]byte, error)
}

// Issue re-exports validate.Issue for callers that only import gen.
type Issue = validate.Issue

// Config parses+validates the request and returns application.yml on success.
func Config(r Request, res Resolver) (out string, errs, warns []Issue) {
	wfs, d, _, pissues := parse(r)
	e, w := validate.Run(validate.Context{Workflows: wfs, Defaults: d, Env: res.Env})
	errs = append(pissues, e...)
	warns = w
	if len(errs) > 0 {
		return "", errs, warns
	}
	m, cw := consolidate.Build(wfs, d, false)
	warns = append(warns, toIssues(cw)...)
	return render.Application(m), nil, warns
}

// Validate runs every check and returns all findings (errors + warnings).
func Validate(r Request, res Resolver) (errs, warns []Issue) {
	wfs, d, k, pissues := parse(r)
	valuesKeys, vErr := valuesFileKeys(k, res)
	e, w := validate.Run(validate.Context{
		Workflows:      wfs,
		Defaults:       d,
		Kube:           k,
		Deploy:         k != nil,
		Env:            res.Env,
		ValuesFileKeys: valuesKeys,
	})
	errs = append(pissues, e...)
	if vErr != nil {
		errs = append(errs, *vErr)
	}
	warns = w
	return errs, warns
}

// Deploy parses+validates and returns the manifest set on success.
func Deploy(r Request, res Resolver) (out string, errs, warns []Issue) {
	wfs, d, k, pissues := parse(r)
	if k == nil {
		pissues = append(pissues, Issue{Msg: "deploy requires a kubernetes.yaml settings file"})
	}
	valuesKeys, vErr := valuesFileKeys(k, res)
	e, w := validate.Run(validate.Context{
		Workflows:      wfs,
		Defaults:       d,
		Kube:           k,
		Deploy:         true,
		Env:            res.Env,
		ValuesFileKeys: valuesKeys,
	})
	errs = append(pissues, e...)
	if vErr != nil {
		errs = append(errs, *vErr)
	}
	warns = w
	if len(errs) > 0 {
		return "", errs, warns
	}

	m, cw := consolidate.Build(wfs, d, true)
	warns = append(warns, toIssues(cw)...)
	in := deploy.Input{Kube: k, Defaults: d, Model: m, AppYAML: render.Application(m)}

	if c := k.Secrets.Credentials; c != nil && c.Create != nil {
		kvs, err := resolveCred(c.Create, res)
		if err != nil {
			return "", []Issue{{File: "kubernetes.yaml", Msg: err.Error()}}, warns
		}
		in.CredKVs = kvs
	}
	if s := k.Secrets.Stores; s != nil && s.Create != nil {
		files, err := resolveStores(d, res)
		if err != nil {
			return "", []Issue{{File: "kubernetes.yaml", Msg: err.Error()}}, warns
		}
		in.Stores = files
	}
	return deploy.Render(in), nil, warns
}

// ---- parsing -----------------------------------------------------------------

func parse(r Request) (wfs []spec.Workflow, d *spec.Defaults, k *spec.Kubernetes, issues []Issue) {
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
	d = &spec.Defaults{}
	if r.Defaults != nil {
		pd, err := spec.ParseDefaults(r.Defaults.Data)
		if err != nil {
			issues = append(issues, Issue{File: "defaults.yaml", Msg: err.Error()})
		} else {
			d = pd
		}
	}
	if r.Kubernetes != nil {
		pk, err := spec.ParseKubernetes(r.Kubernetes.Data)
		if err != nil {
			issues = append(issues, Issue{File: "kubernetes.yaml", Msg: err.Error()})
		} else {
			k = pk
		}
	}
	return wfs, d, k, issues
}

// ---- secret resolution -------------------------------------------------------

func resolveCred(c *spec.CredCreate, res Resolver) ([]deploy.KV, error) {
	switch c.Source {
	case spec.SourceEnv:
		var out []deploy.KV
		for _, v := range c.Variables {
			val := ""
			if res.Env != nil {
				if got, ok := res.Env(v); ok {
					val = got
				} else {
					return nil, fmt.Errorf("credentials variable %q is not set in the environment", v)
				}
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
		return nil, &Issue{File: "kubernetes.yaml", Msg: fmt.Sprintf("reading values-file %q: %v", c.ValuesFile, err)}
	}
	kvs, err := parseValues(data)
	if err != nil {
		return nil, &Issue{File: "kubernetes.yaml", Msg: fmt.Sprintf("parsing values-file %q: %v", c.ValuesFile, err)}
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
