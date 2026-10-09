package gen

import (
	"errors"
	"fmt"
	"io/fs"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/consolidate"
	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/logback"
	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/podmangen"
	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/spec"
	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/statusscript"
)

// TestParseExpandsNonCredentialAndWarnsOnUnsetDefaultless pins the wiring:
// parse() must call spec.Expand after ParseWorkflow/ParseEnv, expanding
// non-credential fields, leaving credential fields verbatim, and surfacing
// one warning per unset defaultless variable in the returned warns list.
func TestParseExpandsNonCredentialAndWarnsOnUnsetDefaultless(t *testing.T) {
	env := File{Name: "env.yaml", Data: nil}
	wf := File{Name: "workflow-0.yaml", Data: []byte(`
source:
  solace:
    host: tcps://${HOST}:55443
    msg-vpn: ${VPN:fallback-vpn}
    client-username: literal-user
    client-password-env: ${TYPO_CRED}
    queue: Q.IN
target:
  mq:
    conn-name: ${TYPO}(1414)
    queue-manager: QM1
    channel: CH
    queue: Q.OUT
`)}
	lookup := func(name string) (string, bool) {
		if name == "HOST" {
			return "broker.internal", true
		}
		return "", false
	}
	wfs, _, issues, warns := parse(Request{Env: &env, Workflows: []File{wf}}, Resolver{Env: lookup})
	if len(issues) != 0 {
		t.Fatalf("unexpected parse issues: %v", issues)
	}
	src := wfs[0].Source
	if got, want := src.Host, "tcps://broker.internal:55443"; got != want {
		t.Errorf("Host = %q, want %q", got, want)
	}
	if got, want := src.MsgVPN, "fallback-vpn"; got != want {
		t.Errorf("MsgVPN = %q, want %q", got, want)
	}
	if got, want := src.ClientPassEnv, "${TYPO_CRED}"; got != want {
		t.Errorf("credential field must never expand: got %q, want %q", got, want)
	}
	if got, want := wfs[0].Target.ConnName, "${TYPO}(1414)"; got != want {
		t.Errorf("unset defaultless var must pass through verbatim: got %q, want %q", got, want)
	}
	if len(warns) != 1 || !strings.Contains(warns[0].Msg, "TYPO") {
		t.Fatalf("warns = %v, want exactly 1 warning naming TYPO", warns)
	}
}

func TestResolveStores(t *testing.T) {
	d := &spec.Defaults{TLS: spec.TLSConfig{
		Truststore: &spec.Store{File: "certs/t.jks"},
		Keystore:   &spec.Store{File: "certs/k.jks"},
	}}
	sf, err := resolveStores(d, Resolver{ReadFile: func(string) ([]byte, error) { return []byte("BYTES"), nil }})
	if err != nil || len(sf) != 2 || sf[0].Name != "t.jks" {
		t.Fatalf("sf=%+v err=%v", sf, err)
	}
	if _, err := resolveStores(d, Resolver{}); err == nil {
		t.Error("no ReadFile should error")
	}
	if _, err := resolveStores(d, Resolver{ReadFile: func(string) ([]byte, error) { return nil, errors.New("x") }}); err == nil {
		t.Error("read error should propagate")
	}
	if sf2, err := resolveStores(&spec.Defaults{}, Resolver{ReadFile: func(string) ([]byte, error) { return nil, nil }}); err != nil || len(sf2) != 0 {
		t.Fatalf("no stores: %v %v", sf2, err)
	}
}

func TestToIssues(t *testing.T) {
	// The path-basename cases moved to spec.TestBaseName with the helper itself.
	if iss := toIssues([]string{"a", "b"}); len(iss) != 2 || iss[0].Msg != "a" {
		t.Errorf("toIssues=%v", iss)
	}
}

// ---- names, paths, mounts (docker/podman plumbing) --------------------------

// TestPodmanFileSecretNames pins the five file secrets a podman instance can
// own, and that none of them can take a credential's name: a credential is
// stored as <name>-<stable>, and a stable name matches [A-Za-z_][A-Za-z0-9_]*,
// which no file suffix does.
func TestPodmanFileSecretNames(t *testing.T) {
	got := PodmanFileSecretNames("c")
	want := []string{"c-application.yml", "c-tls-truststore", "c-tls-keystore", "c-status-script", "c-logback-spring.xml"}
	if !slices.Equal(got, want) {
		t.Errorf("PodmanFileSecretNames = %q, want %q", got, want)
	}
	stable := regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)
	for _, n := range got {
		if suffix := strings.TrimPrefix(n, "c-"); stable.MatchString(suffix) {
			t.Errorf("file secret suffix %q could also be a credential's stable name", suffix)
		}
	}
}

// TestResolvePodmanFiles covers what deploy loads into podman's secret store
// for the files a unit mounts: a document as rendered, a store read byte for
// byte (binary content included), and every way a file can fail podman's
// limits -- each stopping before anything is created, naming the file and
// never its content.
func TestResolvePodmanFiles(t *testing.T) {
	const binary = "\x00\xfe\xedJKS\xff\r\n"
	files := map[string][]byte{
		"./certs/t.jks":     []byte(binary),
		"./certs/max.jks":   []byte(strings.Repeat("k", 511999)),
		"./certs/big.jks":   []byte(strings.Repeat("S", 512000)),
		"./certs/empty.jks": {},
	}
	res := Resolver{ReadFile: func(p string) ([]byte, error) {
		if b, ok := files[p]; ok {
			return b, nil
		}
		return nil, fs.ErrNotExist
	}}
	doc := PodmanFile{StoreName: "c-application.yml", Data: "a: 1\n"}
	store := func(src string) PodmanFile {
		return PodmanFile{StoreName: "c-tls-truststore", Source: src, Field: "tls.truststore.file"}
	}

	kvs, err := ResolvePodmanFiles([]PodmanFile{doc, store("./certs/t.jks"), store("./certs/max.jks")}, res)
	if err != nil {
		t.Fatalf("ResolvePodmanFiles: %v", err)
	}
	if len(kvs) != 3 || kvs[0] != (KV{Key: "c-application.yml", Val: "a: 1\n"}) || kvs[1] != (KV{Key: "c-tls-truststore", Val: binary}) || len(kvs[2].Val) != 511999 {
		t.Errorf("resolved = %q", kvs)
	}

	for _, c := range []struct {
		name  string
		file  PodmanFile
		res   Resolver
		want  string
		cause error
	}{
		{"missing store", store("./certs/gone.jks"), res, `reading tls.truststore.file "./certs/gone.jks" for podman's secret store`, fs.ErrNotExist},
		{"no file access", store("./certs/t.jks"), Resolver{}, `cannot read tls.truststore.file "./certs/t.jks" for podman's secret store (no file access)`, nil},
		{"empty store", store("./certs/empty.jks"), res, `tls.truststore.file "./certs/empty.jks" is empty, and podman's secret store cannot hold an empty secret`, nil},
		{"over the limit", store("./certs/big.jks"), res, `tls.truststore.file "./certs/big.jks" is 512000 bytes, over the 511999 bytes podman's secret store holds per secret`, nil},
		{"empty document", PodmanFile{StoreName: "c-application.yml"}, res, "c-application.yml is empty", nil},
	} {
		t.Run(c.name, func(t *testing.T) {
			kvs, err := ResolvePodmanFiles([]PodmanFile{doc, c.file}, c.res)
			if err == nil || !strings.Contains(err.Error(), c.want) {
				t.Fatalf("err = %v, want it to contain %q", err, c.want)
			}
			if kvs != nil {
				t.Errorf("a failure must return nothing to load, got %d entries", len(kvs))
			}
			if c.cause != nil && !errors.Is(err, c.cause) {
				t.Errorf("err = %v, want it to wrap %v", err, c.cause)
			}
			if strings.Contains(err.Error(), "SSSS") || strings.Contains(err.Error(), "a: 1") {
				t.Errorf("error carries file content: %v", err)
			}
		})
	}
}

func TestTargetMounts(t *testing.T) {
	tls := spec.TLSConfig{
		Truststore: &spec.Store{File: "certs/t.jks"},
		Keystore:   &spec.Store{File: "certs/k.jks"},
	}
	res := Resolver{Abs: func(p string) string { return "/abs/" + p }}
	// The store bind-mount target is always the fixed in-container dir; only the
	// host Source comes from res.Abs. Stores are not opt-in -- a configured
	// tls.*.file is mounted because application.yml already points at the mounted
	// path.
	sm, lm := targetMounts(tls, &spec.LibsMount{Dir: "libs"}, res)
	if len(sm) != 2 {
		t.Fatalf("store mounts=%d want 2", len(sm))
	}
	if sm[0].Source != "/abs/certs/t.jks" || sm[0].Target != spec.DefaultStoresMountPath+"/t.jks" {
		t.Errorf("store mount 0 = %+v", sm[0])
	}
	if sm[1].Target != spec.DefaultStoresMountPath+"/k.jks" {
		t.Errorf("store mount 1 = %+v", sm[1])
	}
	// The libs target is the fixed image path too; only Dir is the operator's.
	if lm == nil || lm.Source != "/abs/libs" || lm.Target != spec.DefaultLibsMountPath {
		t.Errorf("libs mount = %+v", lm)
	}
	// No TLS at all is now the only way to get no store mounts, and libs stays
	// opt-in: it has a host dir to name.
	if sm2, lm2 := targetMounts(spec.TLSConfig{}, nil, res); sm2 != nil || lm2 != nil {
		t.Errorf("no TLS and no libs should yield no mounts: %v %v", sm2, lm2)
	}
	// A store present but with no file set is skipped rather than mounted empty.
	if sm3, _ := targetMounts(spec.TLSConfig{Truststore: &spec.Store{}}, nil, res); sm3 != nil {
		t.Errorf("a store with no file should yield no mount: %v", sm3)
	}
}

// TestConfigRejectsSecretNameConflict pins gen.build's refusal. With
// spec.GeneratedNamePrefix reserved, an operator's -env name can no longer reach
// a derived one, so the reachable collision is two derived names folding
// together: stableToken maps runs of punctuation to a single '_', and nothing
// upstream rejects two management users differing only in punctuation. Config
// must fail naming the shared key and both positions, rather than render a
// config where one credential silently takes the other's password.
func TestConfigRejectsSecretNameConflict(t *testing.T) {
	env := File{Name: "env.yaml", Data: []byte(`
security:
  users:
    - name: ops.1
      password: first-pass
    - name: ops-1
      password: second-pass
`)}
	wf := File{Name: "workflow-0.yaml", Data: []byte(`
source:
  solace:
    host: tcp://broker.internal:55555
    msg-vpn: prod
    queue: Q.IN
target:
  mq:
    conn-name: host(1414)
    queue-manager: QM1
    channel: CH
    queue: Q.OUT
`)}
	res := Resolver{Env: func(string) (string, bool) { return "v", true }, Rand: fixedStatusRand}
	req := Request{Env: &env, Workflows: []File{wf}}

	// The message must name both claiming positions, not just the contested key:
	// a derived name appears nowhere in the spec, so the key alone does not tell
	// the operator which field to edit.
	wants := []string{
		"_GEN_SECURITY_USER_OPS_1_PASSWORD",
		"security.users[ops.1].password",
		"security.users[ops-1].password",
	}
	check := func(t *testing.T, what string, errs []Issue) {
		t.Helper()
		joined := ""
		for _, e := range errs {
			joined += e.String() + "\n"
		}
		for _, w := range wants {
			if !strings.Contains(joined, w) {
				t.Errorf("%s: error must mention %q, got:\n%s", what, w, joined)
			}
		}
	}

	out, errs, _ := Config(req, res)
	check(t, "Config", errs)
	if out != "" {
		t.Errorf("a rejected config must render nothing, got %d bytes", len(out))
	}

	// Validate must catch it too. A collision is only visible once consolidate
	// assigns names, so without the build call on the validate path this spec
	// would lint clean and fail only at generate/deploy.
	verrs, _ := Validate(req, res)
	check(t, "Validate", verrs)
}

// TestValidateCleanSpecStillPasses guards the build call Validate now makes: a
// spec with no collision must not pick up errors from it, and consolidate's own
// warnings must not start leaking into validate output.
func TestValidateCleanSpecStillPasses(t *testing.T) {
	env := File{Name: "env.yaml", Data: nil}
	wf := File{Name: "workflow-0.yaml", Data: []byte(`
source:
  solace:
    host: tcps://broker.internal:55443
    msg-vpn: prod
    client-username: connector
    client-password-env: SOL_PASSWORD
    queue: Q.IN
target:
  mq:
    conn-name: host(1414)
    queue-manager: QM1
    channel: CH
    password-env: MQ_CORE_PASSWORD
    queue: Q.OUT
`)}
	errs, _ := Validate(Request{Env: &env, Workflows: []File{wf}},
		Resolver{Env: func(string) (string, bool) { return "v", true }, Rand: fixedStatusRand})
	if len(errs) != 0 {
		t.Errorf("a clean spec must validate without errors, got %v", errs)
	}
}

// TestResolveCredentials covers the three behaviors ResolveCredentials must
// get right: a literal reference passes its value straight through, an -env
// reference is read from the resolver's environment, and an unset variable
// fails loud with a message that names the stable secret and the variable --
// never a value (S3).
func TestResolveCredentials(t *testing.T) {
	if kvs, err := ResolveCredentials(nil, Resolver{}); err != nil || len(kvs) != 0 {
		t.Errorf("nil refs -> no kvs, no error: %v %v", kvs, err)
	}

	refs := []consolidate.SecretRef{
		{Stable: "PROD_SOLACE_CLIENT_USERNAME", Literal: "connector"},
		{Stable: "PROD_SOLACE_CLIENT_PASSWORD", EnvVar: "SOL_PASSWORD"},
	}
	env := map[string]string{"SOL_PASSWORD": "s3cr3t"}
	kvs, err := ResolveCredentials(refs, Resolver{Env: func(k string) (string, bool) { v, ok := env[k]; return v, ok }})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []KV{
		{Key: "PROD_SOLACE_CLIENT_USERNAME", Val: "connector"},
		{Key: "PROD_SOLACE_CLIENT_PASSWORD", Val: "s3cr3t"},
	}
	if len(kvs) != len(want) || kvs[0] != want[0] || kvs[1] != want[1] {
		t.Fatalf("kvs = %+v, want %+v", kvs, want)
	}

	// Unset -env variable: fail loud, naming the stable secret and the variable.
	_, err = ResolveCredentials(
		[]consolidate.SecretRef{{Stable: "MQ_CONN_1_PASSWORD", EnvVar: "MQ_PW"}},
		Resolver{Env: func(string) (string, bool) { return "", false }},
	)
	if err == nil {
		t.Fatal("expected an error for an unset environment variable")
	}
	if !strings.Contains(err.Error(), "MQ_CONN_1_PASSWORD") || !strings.Contains(err.Error(), "MQ_PW") {
		t.Errorf("error %q should name the stable secret and the variable", err.Error())
	}

	// No environment access at all: also fails loud, not silently.
	_, err = ResolveCredentials(
		[]consolidate.SecretRef{{Stable: "MQ_CONN_1_PASSWORD", EnvVar: "MQ_PW"}},
		Resolver{},
	)
	if err == nil {
		t.Fatal("expected an error when the resolver has no environment access")
	}
}

// ---- sharding (>20 workflows) ----------------------------------------------

// synthWorkflows builds n minimal, valid workflows (distinct queues per index).
func synthWorkflows(n int) []spec.Workflow {
	var wfs []spec.Workflow
	for i := 0; i < n; i++ {
		wfs = append(wfs, spec.Workflow{
			File: fmt.Sprintf("wf-%02d.yaml", i), Enabled: true, SourceSet: true, TargetSet: true,
			Source: spec.Side{System: spec.SystemSolace, Host: "tcp://b", MsgVPN: "v", ClientUser: "u", ClientPass: "p", DestKind: spec.DestQueue, Dest: fmt.Sprintf("IN-%d", i)},
			Target: spec.Side{System: spec.SystemMQ, ConnName: "h(1414)", QueueManager: "QM", Channel: "C", User: "u", Password: "p", DestKind: spec.DestQueue, Dest: fmt.Sprintf("OUT-%d", i)},
		})
	}
	return wfs
}

// synthWorkflowFiles renders synthWorkflows as gen.File YAML for end-to-end tests.
func synthWorkflowFiles(n int) []File {
	var fs []File
	for i := 0; i < n; i++ {
		data := fmt.Sprintf(`source:
  solace:
    host: tcp://b
    msg-vpn: v
    client-username: u
    client-password: p
    queue: IN-%d
target:
  mq:
    conn-name: h(1414)
    queue-manager: QM
    channel: C
    user: u
    password: p
    queue: OUT-%d
`, i, i)
		fs = append(fs, File{Name: fmt.Sprintf("wf-%02d.yaml", i), Data: []byte(data)})
	}
	return fs
}

// TestConfigNumbersWorkflowsInLsOrder pins the second place workflow ids are
// assigned -- gen.parse over an explicit file list -- to the same order the
// folder scan uses: the one LC_ALL=C ls lists, so 10.yaml becomes workflow 0
// ahead of 2.yaml, whatever order the caller handed the files over in.
func TestConfigNumbersWorkflowsInLsOrder(t *testing.T) {
	files := synthWorkflowFiles(2)
	files[0].Name, files[1].Name = "2.yaml", "10.yaml"
	out, errs, _ := Config(Request{Workflows: files}, Resolver{Rand: fixedStatusRand})
	if len(errs) != 0 {
		t.Fatalf("Config: %v", errs)
	}
	for _, want := range []string{
		"        input-0:\n          destination: IN-1\n",
		"        input-1:\n          destination: IN-0\n",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q -- 10.yaml (IN-1) must be workflow 0:\n%s", want, out)
		}
	}
}

// extraKeysEnv is env.yaml for the extra-key end-to-end tests: a reusable
// Solace connection carrying another key, the management session built from
// it, and an mq-defaults block.
const extraKeysEnv = `connections:
  prod-solace:
    solace:
      host: tcp://b:55555
      msg-vpn: prod
      client-username: u
      client-password-env: SOL_PASSWORD
      client-name: ${HOSTNAME}
leader-election:
  mode: active_standby
  queue: mgmt-q
  conn-ref: prod-solace
mq-defaults:
  application-name: fleet
`

// extraKeysWorkflow is a workflow with an inline MQ source carrying other
// keys (one in IBM's camelCase spelling, one nested) and a conn-ref target.
const extraKeysWorkflow = `source:
  mq:
    conn-name: h(1414)
    queue-manager: QM
    channel: C
    user: u
    password: p
    queue: IN
    userAuthenticationMQCSP: false
    pool:
      max-connections: 5
target:
  solace:
    conn-ref: prod-solace
    queue: OUT
`

// extraKeysResolver resolves every -env credential, so the only findings are
// the ones under test.
var extraKeysResolver = Resolver{Env: func(string) (string, bool) { return "v", true }, Rand: fixedStatusRand}

// TestConfigCarriesExtraKeysEndToEnd pins the whole path through gen.Config:
// an inline side's other keys land under its own binder beside the tool's
// keys, after the mq-defaults entry; a connection's other key reaches both
// the binder a conn-ref side builds from it and the management session; and a
// ${...} inside one reaches application.yml as typed.
func TestConfigCarriesExtraKeysEndToEnd(t *testing.T) {
	req := Request{Env: &File{Name: "env.yaml", Data: []byte(extraKeysEnv)}, Workflows: []File{{Name: "wf-00.yaml", Data: []byte(extraKeysWorkflow)}}}
	out, errs, warns := Config(req, extraKeysResolver)
	if len(errs) != 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if issuesContain(warns, "HOSTNAME") || issuesContain(warns, "userAuthenticationMQCSP") {
		t.Errorf("no finding is due on these keys, got %v", warns)
	}
	extras := "                application-name: fleet\n                userAuthenticationMQCSP: false\n                pool:\n                  max-connections: 5\n"
	for _, want := range []string{
		extras,
		"                client-password: ${SOL_PASSWORD}\n                client-name: ${HOSTNAME}\n",
		"        client-password: ${SOL_PASSWORD}\n        client-name: ${HOSTNAME}\n",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing:\n%s\nin:\n%s", want, out)
		}
	}
	// Siblings of the tool's own keys, after them, not inside an
	// additional-properties block.
	if strings.Index(out, "                conn-name: h(1414)\n") > strings.Index(out, extras) || strings.Contains(out, "additional-properties") {
		t.Errorf("the extra keys must follow the tool's keys as siblings:\n%s", out)
	}
}

// TestConfigRejectsExtraKeysBesideConnRef pins that another key beside
// conn-ref fails the render like any connection field there would.
func TestConfigRejectsExtraKeysBesideConnRef(t *testing.T) {
	wf := strings.Replace(extraKeysWorkflow, "    conn-ref: prod-solace\n", "    conn-ref: prod-solace\n    client-name: mine\n", 1)
	req := Request{Env: &File{Name: "env.yaml", Data: []byte(extraKeysEnv)}, Workflows: []File{{Name: "wf-00.yaml", Data: []byte(wf)}}}
	if out, errs, _ := Config(req, extraKeysResolver); out != "" || !issuesContain(errs, "may set only queue/topic") {
		t.Errorf("an extra key beside conn-ref must fail the render, got out=%q errs=%v", out, errs)
	}
}

// TestConfigWarnsOnANearMissKeyButStillRenders pins that a probable typo is a
// warning, not a block: the key passes through and the config still renders.
func TestConfigWarnsOnANearMissKeyButStillRenders(t *testing.T) {
	wf := strings.Replace(extraKeysWorkflow, "    userAuthenticationMQCSP: false\n", "    queue-managr: QM\n", 1)
	req := Request{Env: &File{Name: "env.yaml", Data: []byte(extraKeysEnv)}, Workflows: []File{{Name: "wf-00.yaml", Data: []byte(wf)}}}
	out, errs, warns := Config(req, extraKeysResolver)
	if len(errs) != 0 || !strings.Contains(out, "                queue-managr: QM\n") {
		t.Errorf("the typo must pass through and render, got errs=%v out:\n%s", errs, out)
	}
	if !issuesContain(warns, `did you mean "queue-manager"?`) {
		t.Errorf("want the near-miss warning, got %v", warns)
	}
}

// TestConfigWorkflowCap pins the new hard cap: a folder holding more than
// validate.MaxWorkflows workflows is a fatal error through the real
// gen.Config path (no sharding, no output).
func TestConfigWorkflowCap(t *testing.T) {
	out, errs, _ := Config(Request{Workflows: synthWorkflowFiles(21)}, Resolver{})
	if out != "" {
		t.Errorf("expected no output over the cap, got:\n%s", out)
	}
	want := "21 workflows found, but one connector instance runs at most 20 (workflow ids 0..19). Split them across separate folders, each with its own env.yaml and its own deployment.name/docker.name/podman.name, and deploy each as its own connector"
	if !issuesContain(errs, want) {
		t.Fatalf("errs = %v, want one containing %q", errs, want)
	}
	// At the cap is still fine.
	if _, errs, _ := Config(Request{Workflows: synthWorkflowFiles(20)}, Resolver{}); len(errs) > 0 {
		t.Errorf("20 workflows should not hit the cap: %v", errs)
	}
}

// TestGenerateKubernetesWorkflowCap covers the same cap through the
// kubernetes target: over the limit is fatal and produces no manifest.
func TestGenerateKubernetesWorkflowCap(t *testing.T) {
	envData := "image:\n  name: img\n  tag: v1\nkubernetes:\n  deployment:\n    name: solmq\n    namespace: ns\n  service:\n    enabled: true\n    port: 8090\n"
	req := Request{
		Env:       &File{Name: "env.yaml", Data: []byte(envData)},
		Workflows: synthWorkflowFiles(21),
	}
	out, errs, _ := GenerateKubernetes(req, Resolver{}, KubeOpts{})
	if out != "" {
		t.Errorf("expected no manifest over the cap, got:\n%s", out)
	}
	if !issuesContain(errs, "21 workflows found, but one connector instance runs at most 20") {
		t.Fatalf("errs = %v, want the workflow-cap message", errs)
	}
}

// fixedStatusRand fills b with a repeating 0xab pattern, giving
// resolveStatusPassword a deterministic, obviously-synthetic result
// ("ab" x16 hex-encoded) instead of a fresh crypto/rand draw every run.
func fixedStatusRand(b []byte) error {
	for i := range b {
		b[i] = 0xab
	}
	return nil
}

// Every generated application.yml must import the mounted secret files and
// must never carry a credential value or a host variable name -- only the
// ${STABLE} placeholder, with exactly one carve-out: management security is
// always on (no operator toggle left, no security: block needed here), so
// consolidate.applyStatusAccess unconditionally appends the reserved
// spec.StatusUserName account carrying its password as a literal -- the
// generated status script reads that literal back out of application.yml at
// run time, so it has nothing else to read. This guards both invariants
// end-to-end through Config: every ordinary password is a placeholder, and
// the one literal that exists is exactly the expected status-account value.
// TestConfigCarriesSecurityUserRoles is the end-to-end proof that an operator's
// roles survive the whole chain -- validate accepts them, consolidate carries
// them, render emits them -- and that the reserved status account still renders
// without any, which is what keeps it read-only. Guards the gap this test was
// written for: roles used to be documented but silently dropped, since nothing
// in the model carried the key.
func TestConfigCarriesSecurityUserRoles(t *testing.T) {
	env := &File{Name: "env.yaml", Data: []byte(`
security:
  users:
    - name: ops
      password-env: OPS_PASS
      roles:
        - admin
`)}
	res := Resolver{
		Rand: fixedStatusRand,
		Env: func(k string) (string, bool) {
			if k == "OPS_PASS" {
				return "ops-pw", true
			}
			return "", false
		},
	}
	out, errs, _ := Config(Request{Env: env, Workflows: synthWorkflowFiles(1)}, res)
	if len(errs) > 0 {
		t.Fatalf("a roles-bearing env.yaml must generate cleanly, got: %v", errs)
	}
	want := "          roles:\n            - admin\n"
	if !strings.Contains(out, want) {
		t.Errorf("missing the ops user's rendered roles block:\n%s", out)
	}
	if n := strings.Count(out, "roles:"); n != 1 {
		t.Errorf("roles: appears %d times, want 1 -- the reserved account must not emit it\n%s", n, out)
	}
}

func TestConfigNoSecretsLeak(t *testing.T) {
	out, errs, _ := Config(Request{Workflows: synthWorkflowFiles(1)}, Resolver{Rand: fixedStatusRand})
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if !strings.HasPrefix(out, "spring:\n  config:\n    import: "+ConfigImport+"\n") {
		t.Errorf("application.yml must start with the config-tree import:\n%s", out)
	}
	if strings.Contains(out, "client-username: u") || strings.Contains(out, "user: u") {
		t.Errorf("rendered config leaks a literal credential value:\n%s", out)
	}
	if !strings.Contains(out, "${_GEN_SOL_CONN_1_CLIENT_USERNAME}") || !strings.Contains(out, "${_GEN_MQ_CONN_1_USER}") {
		t.Errorf("rendered config missing expected ${STABLE} placeholders:\n%s", out)
	}

	wantStatusPW := strings.Repeat("ab", 16)
	if !strings.Contains(out, "password: "+wantStatusPW) {
		t.Errorf("expected the reserved status account's literal test password %q:\n%s", wantStatusPW, out)
	}
	for _, ln := range strings.Split(out, "\n") {
		i := strings.Index(ln, "password:")
		if i < 0 {
			continue
		}
		val := strings.TrimSpace(ln[i+len("password:"):])
		if val == wantStatusPW {
			continue // the one permitted literal: the reserved status account
		}
		if !strings.HasPrefix(val, "${") || !strings.HasSuffix(val, "}") {
			t.Errorf("password line %q carries neither a ${STABLE} placeholder nor the reserved status account's literal:\n%s", ln, out)
		}
	}
}

// ---- docker / podman generation ---------------------------------------------

func TestGenerateDockerBasics(t *testing.T) {
	envData := `timezone: UTC
image:
  name: solace/connector
  tag: "9.9"
tls:
  truststore:
    file: ./certs/truststore.jks
    password: ts
    type: JKS
docker:
  command: docker
  name: solmq-connector
  restart: unless-stopped
  ports:
    - 8090
`
	req := Request{Env: &File{Name: "env.yaml", Data: []byte(envData)}, Workflows: synthWorkflowFiles(1)}
	plan, errs, _ := GenerateDocker(req, Resolver{})
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if plan.Compose == "" {
		t.Fatal("empty compose")
	}
	if !strings.Contains(plan.Compose, "solace/connector:9.9") {
		t.Errorf("compose missing image:\n%s", plan.Compose)
	}

	// The compose project comes from the spec, so it is the document's first
	// line -- and it is the defaulted value here, since the docker: section
	// above sets no project-name.
	wantName := "name: " + spec.DefaultComposeProject + "\n"
	if !strings.HasPrefix(plan.Compose, wantName) {
		t.Errorf("compose must open with %q:\n%s", wantName, plan.Compose)
	}

	// One workflow, inline solace source (auto sol-conn-1) + inline mq target
	// (auto mq-conn-1): four credential positions, all literal -- so all four
	// take derived names, which carry spec.GeneratedNamePrefix.
	want := []consolidate.SecretRef{
		{Stable: "_GEN_SOL_CONN_1_CLIENT_USERNAME", Literal: "u"},
		{Stable: "_GEN_SOL_CONN_1_CLIENT_PASSWORD", Literal: "p"},
		{Stable: "_GEN_MQ_CONN_1_USER", Literal: "u"},
		{Stable: "_GEN_MQ_CONN_1_PASSWORD", Literal: "p"},
	}
	if len(plan.Secrets) != len(want) {
		t.Fatalf("secrets = %+v, want %+v", plan.Secrets, want)
	}
	for i, w := range want {
		if plan.Secrets[i] != w {
			t.Errorf("secrets[%d] = %+v, want %+v", i, plan.Secrets[i], w)
		}
	}

	// Every declared secret is a compose top-level secret (environment provider)
	// and is listed under the service's own secrets:, never written as a value.
	for _, s := range want {
		if !strings.Contains(plan.Compose, s.Stable+":\n") {
			t.Errorf("compose missing top-level secret %q:\n%s", s.Stable, plan.Compose)
		}
		if strings.Contains(plan.Compose, s.Stable+": "+s.Literal) {
			t.Errorf("compose must never carry a secret value inline (%q):\n%s", s.Stable, plan.Compose)
		}
		// The ${STABLE} placeholder application.yml carries is Spring's, resolved
		// from the configtree import of the secrets mount. It reaches the container only
		// if it is escaped for compose's interpolation pass: left bare, compose
		// substitutes it from the environment the CLI hands the compose child, and
		// the plaintext credential is written into the document after all.
		if !strings.Contains(plan.Compose, "$${"+s.Stable+"}") {
			t.Errorf("compose placeholder for %q is not escaped against interpolation:\n%s", s.Stable, plan.Compose)
		}
	}
}

func TestGeneratePodmanQuadlet(t *testing.T) {
	envData := `timezone: UTC
image:
  name: solace/connector
  tag: "9.9"
podman:
  command: podman
  name: solmq-connector
  restart: unless-stopped
  ports:
    - 8090
`
	req := Request{Env: &File{Name: "env.yaml", Data: []byte(envData)}, Workflows: synthWorkflowFiles(1)}
	res := Resolver{Env: func(string) (string, bool) { return "v", true }}

	plan, errs, _ := GeneratePodman(req, res)
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if plan.Unit == (podmangen.Unit{}) {
		t.Error("no quadlet unit rendered")
	}
	if plan.Unit.Filename != "solmq-connector.container" {
		t.Errorf("unit filename = %q", plan.Unit.Filename)
	}
	var files []string
	for _, f := range plan.Files {
		files = append(files, f.StoreName)
	}
	if want := []string{"solmq-connector-application.yml", "solmq-connector-status-script"}; !slices.Equal(files, want) {
		t.Errorf("file secrets = %q, want %q", files, want)
	}
	if plan.Service != "solmq-connector.service" {
		t.Errorf("service = %q", plan.Service)
	}
	if len(plan.Secrets) != 4 {
		t.Fatalf("secrets = %+v, want 4 entries", plan.Secrets)
	}
	// Each secret is mounted from podman's store by its namespaced store name, at
	// an absolute target under the secrets mount rather than podman's default
	// /run/secrets. The value itself never reaches the unit.
	for _, s := range plan.Secrets {
		store := PodmanSecretStoreName("solmq-connector", s.Stable)
		want := "Secret=" + store + ",type=mount,target=" + spec.SecretsMountPath + "/" + s.Stable
		if !strings.Contains(plan.Unit.Content, want) {
			t.Errorf("unit missing secret directive %q:\n%s", want, plan.Unit.Content)
		}
	}
}

// TestConfigRendersTransformHeadersUnderItsWorkflow is the end-to-end pin for
// header transforms. The block is written in the workflow file because only the
// file knows which workflow number it becomes: here the second file by sorted
// name, so its block must land under solace.connector.workflows.1 -- verbatim,
// quoting and key order intact -- and nowhere under workflow 0.
func TestConfigRendersTransformHeadersUnderItsWorkflow(t *testing.T) {
	files := synthWorkflowFiles(2)
	files[1].Data = append(files[1].Data, []byte(`transform-headers:
  expressions:
    solace_scst_targetDestination: "'orders/' + headers.region"
    JMS_IBM_Format: 'MQSTR'
`)...)
	out, errs, _ := Config(Request{Workflows: files}, Resolver{Rand: fixedStatusRand})
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	want := `    workflows:
      0:
        enabled: true
      1:
        enabled: true
        transform-headers:
          expressions:
            solace_scst_targetDestination: "'orders/' + headers.region"
            JMS_IBM_Format: 'MQSTR'
`
	if !strings.Contains(out, want) {
		t.Errorf("application.yml missing the transform under workflow 1, want:\n%s\ngot:\n%s", want, out)
	}

	// A misspelt transform stops the render rather than dropping it.
	files[1].Data = append(synthWorkflowFiles(2)[1].Data, []byte("transform-header:\n  expressions:\n    h: x\n")...)
	if out, errs, _ := Config(Request{Workflows: files}, Resolver{Rand: fixedStatusRand}); out != "" || !issuesContain(errs, "transform-header is not a key") {
		t.Errorf("a misspelt transform must fail the render, got out=%q errs=%v", out, errs)
	}
}

// TestConfigRendersTransformUnderItsWorkflow is the end-to-end pin for the
// connector's current transform section: the second workflow file's
// transform: renders verbatim under solace.connector.workflows.1 and nowhere
// under 0, with a ${...} inside an expression left exactly as written --
// passthrough blocks are never expanded -- and no warning about it. The ways
// a transform goes wrong each fail the render instead of being dropped: the
// transform-headers shape carried over, a transform under a side, and both
// sections in one file.
func TestConfigRendersTransformUnderItsWorkflow(t *testing.T) {
	base := synthWorkflowFiles(2)[1].Data
	withExtra := func(extra string) []File {
		files := synthWorkflowFiles(2)
		files[1].Data = append(append([]byte(nil), base...), extra...)
		return files
	}

	out, errs, warns := Config(Request{Workflows: withExtra(`transform:
  target-payload:
    content-type: application/json
  expressions:
    - transform: "target['headers']['scst_targetDestination'] = #joinString('/', 'orders', '${DEPLOY_ENV}')"
`)}, Resolver{Env: func(string) (string, bool) { return "", false }, Rand: fixedStatusRand})
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	want := `    workflows:
      0:
        enabled: true
      1:
        enabled: true
        transform:
          target-payload:
            content-type: application/json
          expressions:
            -
              transform: "target['headers']['scst_targetDestination'] = #joinString('/', 'orders', '${DEPLOY_ENV}')"
`
	if !strings.Contains(out, want) {
		t.Errorf("application.yml missing the transform under workflow 1, want:\n%s\ngot:\n%s", want, out)
	}
	if issuesContain(warns, "DEPLOY_ENV") {
		t.Errorf("a ${...} inside a transform is the connector's, not expanded or warned about, got %v", warns)
	}

	sideTransform := strings.Replace(string(base), "    queue: IN-1\n", "    queue: IN-1\n    transform:\n      expressions:\n        - transform: x\n", 1)
	for _, c := range []struct {
		name  string
		files []File
		want  string
	}{
		{"the transform-headers shape", withExtra("transform:\n  expressions:\n    h: x\n"), "transform.expressions must be a list of - transform: <SpEL expression> items, got a mapping"},
		{"a transform under a side", []File{synthWorkflowFiles(2)[0], {Name: "wf-01.yaml", Data: []byte(sideTransform)}}, "source.solace.transform is in the wrong place"},
		{"both sections", withExtra("transform:\n  expressions:\n    - transform: x\ntransform-headers:\n  expressions:\n    h: x\n"), "cannot use the two together"},
	} {
		if out, errs, _ := Config(Request{Workflows: c.files}, Resolver{Rand: fixedStatusRand}); out != "" || !issuesContain(errs, c.want) {
			t.Errorf("%s must fail the render with %q, got out=%q errs=%v", c.name, c.want, out, errs)
		}
	}
}

// TestTransformHeadersDeprecationIsLintOnly pins where Solace's deprecation of
// transform-headers is reported end to end: validate asks for the migration,
// while generate -- and so deploy -- renders the same file without a word,
// since the connector still reads it.
func TestTransformHeadersDeprecationIsLintOnly(t *testing.T) {
	const deprecated = "transform-headers is deprecated by Solace"
	files := synthWorkflowFiles(2)
	files[1].Data = append(files[1].Data, []byte("transform-headers:\n  expressions:\n    h: \"'x'\"\n")...)

	if _, warns := Validate(Request{Workflows: files}, Resolver{Rand: fixedStatusRand}); !issuesContain(warns, deprecated) {
		t.Errorf("validate must warn that transform-headers is deprecated, got %v", warns)
	}
	out, errs, warns := Config(Request{Workflows: files}, Resolver{Rand: fixedStatusRand})
	if out == "" || len(errs) > 0 {
		t.Fatalf("a transform-headers file must still generate, got errs=%v", errs)
	}
	if issuesContain(warns, deprecated) {
		t.Errorf("generate must stay quiet about the deprecation, got %v", warns)
	}
}

// TestRetiredCreateNamesDeployButDoNotValidate is the end-to-end pin for the
// upgrade contract of the name keys the tool no longer honours. An env.yaml
// written before names were derived must keep generating -- every object
// under its derived name, the old keys ignored without a word -- while
// validate reports each old key and asks for it to be removed.
func TestRetiredCreateNamesDeployButDoNotValidate(t *testing.T) {
	envData := `image:
  name: img
  tag: v1
tls:
  truststore:
    file: ./certs/truststore.jks
    password: ts
    type: JKS
kubernetes:
  deployment:
    name: solmq
    namespace: ns
  secrets:
    credentials:
      create:
        name: shared-creds
    stores:
      create:
        name: shared-tls
  libs:
    pvc:
      create:
        name: shared-libs
        nfs:
          server: nfs1
          path: /libs
`
	req := Request{Env: &File{Name: "env.yaml", Data: []byte(envData)}, Workflows: synthWorkflowFiles(1)}
	res := Resolver{Rand: fixedStatusRand, ReadFile: func(string) ([]byte, error) { return []byte("JKS"), nil }}

	out, errs, _ := GenerateKubernetes(req, res, KubeOpts{})
	if len(errs) > 0 {
		t.Fatalf("generate must ignore the retired names quietly, got %v", errs)
	}
	for _, want := range []string{
		"  name: solmq-credentials\n", "  name: solmq-stores\n", "  name: solmq-libs\n",
		"claimName: solmq-libs\n", "  name: ns-solmq-libs-pv\n",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("manifest missing the derived name %q:\n%s", want, out)
		}
	}
	for _, retired := range []string{"shared-creds", "shared-tls", "shared-libs"} {
		if strings.Contains(out, retired) {
			t.Errorf("the retired name %q reached the manifest:\n%s", retired, out)
		}
	}

	verrs, _ := Validate(req, res)
	for _, field := range []string{
		"kubernetes.secrets.credentials.create.name is no longer accepted",
		"kubernetes.secrets.stores.create.name is no longer accepted",
		"libs.pvc.create.name is no longer accepted",
	} {
		found := false
		for _, e := range verrs {
			if strings.Contains(e.Msg, field) {
				found = true
			}
		}
		if !found {
			t.Errorf("validate must report %q, got %v", field, verrs)
		}
	}
}

// TestGenerateJavaOptionsReachEveryPlatform is the end-to-end pin for the
// top-level java-options: block. Written once, as a >- folded block with
// ${VAR} references in it, it reaches the kubernetes manifest, the compose file
// and the quadlet unit as the same single line; and an unsafe value stops every
// platform's generation before anything renders.
func TestGenerateJavaOptionsReachEveryPlatform(t *testing.T) {
	const (
		base   = "image:\n  name: img\n  tag: v1\n"
		block  = "java-options:\n  tool: >-\n    -Xms512m\n    -Xmx${HEAP:768m}\n  jdk: -Dregion=${REGION}\n"
		kube   = "kubernetes:\n  command: kubectl\n  deployment:\n    name: solmq\n    namespace: ns\n"
		docker = "docker:\n  command: docker\n  name: solmq\n"
		podman = "podman:\n  command: podman\n  name: solmq\n"
	)
	res := Resolver{Rand: fixedStatusRand, Env: func(name string) (string, bool) {
		if name == "REGION" {
			return "apac", true
		}
		return "", false
	}}
	req := func(env string) Request {
		return Request{Env: &File{Name: "env.yaml", Data: []byte(env)}, Workflows: synthWorkflowFiles(1)}
	}
	// valueOf returns the rest of the line that follows marker, so a check does
	// not depend on whether the MQ TLS flag leads JAVA_TOOL_OPTIONS.
	valueOf := func(out, marker string) string {
		i := strings.Index(out, marker)
		if i < 0 {
			return ""
		}
		rest := out[i+len(marker):]
		if j := strings.Index(rest, "\n"); j >= 0 {
			rest = rest[:j]
		}
		return rest
	}

	k8s, errs, _ := GenerateKubernetes(req(base+block+kube), res, KubeOpts{})
	if len(errs) > 0 {
		t.Fatalf("kubernetes: unexpected errors: %v", errs)
	}
	if v := valueOf(k8s, "- name: JAVA_TOOL_OPTIONS\n              value: "); !strings.HasSuffix(v, `-Xms512m -Xmx768m"`) {
		t.Errorf("kubernetes JAVA_TOOL_OPTIONS = %s, want it to end with the folded, expanded options:\n%s", v, k8s)
	}
	if v := valueOf(k8s, "- name: JDK_JAVA_OPTIONS\n              value: "); v != `"-Dregion=apac"` {
		t.Errorf("kubernetes JDK_JAVA_OPTIONS = %s, want \"-Dregion=apac\"", v)
	}

	compose, errs, _ := GenerateDocker(req(base+block+docker), res)
	if len(errs) > 0 {
		t.Fatalf("docker: unexpected errors: %v", errs)
	}
	if v := valueOf(compose.Compose, "JAVA_TOOL_OPTIONS: "); !strings.HasSuffix(v, `-Xms512m -Xmx768m"`) {
		t.Errorf("compose JAVA_TOOL_OPTIONS = %s, want it to end with the folded, expanded options:\n%s", v, compose.Compose)
	}
	if v := valueOf(compose.Compose, "JDK_JAVA_OPTIONS: "); v != `"-Dregion=apac"` {
		t.Errorf("compose JDK_JAVA_OPTIONS = %s, want \"-Dregion=apac\"", v)
	}

	quadlet, errs, _ := GeneratePodman(req(base+block+podman), res)
	if len(errs) > 0 {
		t.Fatalf("podman: unexpected errors: %v", errs)
	}
	if v := valueOf(quadlet.Unit.Content, "JAVA_TOOL_OPTIONS="); !strings.HasSuffix(v, `-Xms512m -Xmx768m"`) {
		t.Errorf("quadlet JAVA_TOOL_OPTIONS = %s, want the quoted, folded, expanded options:\n%s", v, quadlet.Unit.Content)
	}
	if v := valueOf(quadlet.Unit.Content, "Environment=JDK_JAVA_OPTIONS="); v != "-Dregion=apac" {
		t.Errorf("quadlet JDK_JAVA_OPTIONS = %s, want -Dregion=apac", v)
	}

	// The gate runs before any renderer, on every platform.
	const bad = "java-options:\n  tool: '-Dx=\"y\"'\n"
	gated := func(platform string, errs []Issue) {
		t.Helper()
		for _, e := range errs {
			if strings.Contains(e.Msg, "java-options.tool (JAVA_TOOL_OPTIONS)") {
				return
			}
		}
		t.Errorf("%s: an unsafe java-options.tool must be an error, got %v", platform, errs)
	}
	if out, errs, _ := GenerateKubernetes(req(base+bad+kube), res, KubeOpts{}); out != "" || len(errs) == 0 {
		t.Errorf("kubernetes rendered despite an unsafe java-options.tool")
	} else {
		gated("kubernetes", errs)
	}
	_, errs, _ = GenerateDocker(req(base+bad+docker), res)
	gated("docker", errs)
	_, errs, _ = GeneratePodman(req(base+bad+podman), res)
	gated("podman", errs)
}

// TestGeneratePodmanRejectsModeKey pins the removed podman.mode key. Both former
// values error: generate emits the quadlet unit either way, so a section asking
// for the old run script is told rather than silently given something else.
func TestGeneratePodmanRejectsModeKey(t *testing.T) {
	for _, mode := range []string{"run", "quadlet"} {
		t.Run(mode, func(t *testing.T) {
			envData := `image:
  name: solace/connector
  tag: "9.9"
podman:
  command: podman
  mode: ` + mode + `
  name: solmq-connector
`
			req := Request{Env: &File{Name: "env.yaml", Data: []byte(envData)}, Workflows: synthWorkflowFiles(1)}
			res := Resolver{Env: func(string) (string, bool) { return "v", true }}
			_, errs, _ := GeneratePodman(req, res)
			if !issuesContain(errs, "podman.mode is no longer configured") {
				t.Errorf("mode %q should be rejected, got %v", mode, errs)
			}
		})
	}
}

// TestGeneratePodmanNoModeKeyIsClean is the other half: an omitted mode: must not
// trip the rejection. It guards the removed default in applyPodmanDefaults -- were
// that still setting the field, every section would fail the check above for a
// value the operator never wrote.
func TestGeneratePodmanNoModeKeyIsClean(t *testing.T) {
	envData := `image:
  name: solace/connector
  tag: "9.9"
podman:
  command: podman
  name: solmq-connector
`
	req := Request{Env: &File{Name: "env.yaml", Data: []byte(envData)}, Workflows: synthWorkflowFiles(1)}
	res := Resolver{Env: func(string) (string, bool) { return "v", true }}
	_, errs, _ := GeneratePodman(req, res)
	if len(errs) > 0 {
		t.Errorf("an omitted mode: must be clean, got %v", errs)
	}
}

// ---- status password resolution + per-platform status script wiring -------

// TestResolveStatusPasswordFixedRand pins the generated branch: a fixed Rand
// hook yields the exact expected hex literal (16 bytes -> 32 lowercase hex
// chars), never a randomized value that would make the test flaky.
func TestResolveStatusPasswordFixedRand(t *testing.T) {
	res := Resolver{Rand: func(b []byte) error {
		for i := range b {
			b[i] = byte(i)
		}
		return nil
	}}
	got, err := resolveStatusPassword(res)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "000102030405060708090a0b0c0d0e0f"
	if got != want {
		t.Errorf("resolveStatusPassword = %q, want %q", got, want)
	}
}

// TestResolveStatusPasswordEnvOverride pins the override branch: a set,
// non-empty spec.StatusUserPasswordEnvVar is used verbatim and Rand is never
// consulted (validate.Run has already charset-checked the value).
func TestResolveStatusPasswordEnvOverride(t *testing.T) {
	randCalled := false
	res := Resolver{
		Env: func(k string) (string, bool) {
			if k == spec.StatusUserPasswordEnvVar {
				return "operator-chosen-pw", true
			}
			return "", false
		},
		Rand: func(b []byte) error { randCalled = true; return nil },
	}
	got, err := resolveStatusPassword(res)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "operator-chosen-pw" {
		t.Errorf("resolveStatusPassword = %q, want the env override verbatim", got)
	}
	if randCalled {
		t.Error("env override is set: Rand must not be consulted")
	}
}

// TestResolveStatusPasswordEmptyEnvFallsBackToRand covers the "set but empty"
// case: an empty override is treated the same as unset, so generation falls
// back to Rand rather than returning "".
func TestResolveStatusPasswordEmptyEnvFallsBackToRand(t *testing.T) {
	res := Resolver{
		Env:  func(string) (string, bool) { return "", true },
		Rand: fixedStatusRand,
	}
	got, err := resolveStatusPassword(res)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != strings.Repeat("ab", 16) {
		t.Errorf("resolveStatusPassword = %q, want the generated fallback", got)
	}
}

// TestResolveStatusPasswordRandError pins the hard-failure branch: Rand
// failing must surface as an actionable error naming the underlying cause,
// never a predictable fallback password.
func TestResolveStatusPasswordRandError(t *testing.T) {
	_, err := resolveStatusPassword(Resolver{Rand: func([]byte) error { return errors.New("entropy unavailable") }})
	if err == nil || !strings.Contains(err.Error(), "entropy unavailable") {
		t.Fatalf("err = %v, want an actionable error naming the underlying cause", err)
	}
}

// TestConfigStatusPasswordRandErrorNoOutput covers the same failure through
// the real Config path: a Rand error is a hard error and produces no output.
func TestConfigStatusPasswordRandErrorNoOutput(t *testing.T) {
	res := Resolver{Rand: func([]byte) error { return errors.New("entropy unavailable") }}
	out, errs, _ := Config(Request{Workflows: synthWorkflowFiles(1)}, res)
	if out != "" {
		t.Errorf("expected no output on a Rand failure, got:\n%s", out)
	}
	if !issuesContain(errs, "entropy unavailable") {
		t.Fatalf("errs = %v, want one naming the Rand failure", errs)
	}
}

// TestGenerateKubernetesCarriesStatusScript pins the k8s wiring: the
// ConfigMap gets a "status: |" key carrying the rendered script, addressed to
// the reserved account and the port GenerateKubernetes itself resolved (the
// same fallback chain deploy.ManagementPort implements).
func TestGenerateKubernetesCarriesStatusScript(t *testing.T) {
	envData := "image:\n  name: img\n  tag: v1\nkubernetes:\n  command: kubectl\n  deployment:\n    name: solmq\n    namespace: ns\n"
	req := Request{Env: &File{Name: "env.yaml", Data: []byte(envData)}, Workflows: synthWorkflowFiles(1)}
	out, errs, _ := GenerateKubernetes(req, Resolver{Rand: fixedStatusRand}, KubeOpts{})
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if !strings.Contains(out, "  status: |\n") {
		t.Errorf("ConfigMap missing the status: | key:\n%s", out)
	}
	if !strings.Contains(out, "USER_NAME="+spec.StatusUserName) {
		t.Errorf("rendered status script missing USER_NAME=%s:\n%s", spec.StatusUserName, out)
	}
	if !strings.Contains(out, "PORT=8090") { // no management.port/service.port set: falls back to 8090
		t.Errorf("rendered status script missing the fallback management port:\n%s", out)
	}
}

// TestGenerateDockerCarriesStatusScript pins the compose wiring: a second
// top-level config (<name>-status) inlines the rendered script and the
// service mounts it at statusscript.ContainerPath.
func TestGenerateDockerCarriesStatusScript(t *testing.T) {
	envData := "image:\n  name: img\n  tag: v1\ndocker:\n  command: docker\n  name: solmq-connector\n"
	req := Request{Env: &File{Name: "env.yaml", Data: []byte(envData)}, Workflows: synthWorkflowFiles(1)}
	plan, errs, _ := GenerateDocker(req, Resolver{Rand: fixedStatusRand})
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if !strings.Contains(plan.Compose, "solmq-connector-status:\n    content: |\n") {
		t.Errorf("compose missing the second (status) config entry:\n%s", plan.Compose)
	}
	if !strings.Contains(plan.Compose, "USER_NAME="+spec.StatusUserName) {
		t.Errorf("compose missing the rendered status script body:\n%s", plan.Compose)
	}
	if !strings.Contains(plan.Compose, "target: "+statusscript.ContainerPath) {
		t.Errorf("compose missing the status mount target %q:\n%s", statusscript.ContainerPath, plan.Compose)
	}
}

// TestGeneratePodmanCarriesStatusScript pins the podman wiring for the
// rendered documents: application.yml and the status script ride on the plan
// as <name>-application.yml and <name>-status-script with their content, and
// the unit mounts each from podman's secret store at its fixed path rather
// than from a host file.
func TestGeneratePodmanCarriesStatusScript(t *testing.T) {
	envData := "image:\n  name: img\n  tag: v1\npodman:\n  command: podman\n  name: solmq-connector\n"
	req := Request{Env: &File{Name: "env.yaml", Data: []byte(envData)}, Workflows: synthWorkflowFiles(1)}
	res := Resolver{Env: func(string) (string, bool) { return "v", true }, Rand: fixedStatusRand}

	plan, errs, _ := GeneratePodman(req, res)
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	byName := map[string]PodmanFile{}
	for _, f := range plan.Files {
		byName[f.StoreName] = f
	}
	if st := byName["solmq-connector-status-script"]; !strings.Contains(st.Data, "USER_NAME="+spec.StatusUserName) || st.Source != "" {
		t.Errorf("status script file = %+v, want the rendered script", st)
	}
	if app := byName["solmq-connector-application.yml"]; !strings.Contains(app.Data, "solace:") || app.Source != "" {
		t.Errorf("application.yml file = %+v, want the rendered document", app)
	}
	for _, want := range []string{
		"Secret=solmq-connector-application.yml,type=mount,target=/app/external/spring/config/application.yml\n",
		"Secret=solmq-connector-status-script,type=mount,target=" + statusscript.ContainerPath + "\n",
	} {
		if !strings.Contains(plan.Unit.Content, want) {
			t.Errorf("unit missing %q:\n%s", want, plan.Unit.Content)
		}
	}
	if strings.Contains(plan.Unit.Content, "Volume=") {
		t.Errorf("no host file may be mounted without libs:\n%s", plan.Unit.Content)
	}
}

// TestGeneratePodmanCarriesLogbackOnlyWithSyslog pins the podman half of the
// syslog block: with logging.syslog set the plan carries the rendered
// logback-spring.xml as <name>-logback-spring.xml and the unit mounts it from
// the secret store where the image reads it; without it there is neither.
func TestGeneratePodmanCarriesLogbackOnlyWithSyslog(t *testing.T) {
	const env = "image:\n  name: img\n  tag: v1\npodman:\n  command: podman\n  name: c\n"
	res := Resolver{Env: func(string) (string, bool) { return "v", true }, Rand: fixedStatusRand}
	for _, c := range []struct {
		name   string
		syslog bool
	}{{"no syslog", false}, {"syslog", true}} {
		t.Run(c.name, func(t *testing.T) {
			envData := env
			if c.syslog {
				envData += "logging:\n  syslog:\n    host: syslog.corp\n    port: 514\n"
			}
			req := Request{Env: &File{Name: "env.yaml", Data: []byte(envData)}, Workflows: synthWorkflowFiles(1)}
			plan, errs, _ := GeneratePodman(req, res)
			if len(errs) > 0 {
				t.Fatalf("unexpected errors: %v", errs)
			}
			var lb *PodmanFile
			for i := range plan.Files {
				if plan.Files[i].StoreName == "c-logback-spring.xml" {
					lb = &plan.Files[i]
				}
			}
			line := "Secret=c-logback-spring.xml,type=mount,target=" + logback.ContainerPath + "\n"
			if !c.syslog {
				if lb != nil || strings.Contains(plan.Unit.Content, "logback") {
					t.Errorf("no syslog must mean no logback file or mount: %+v\n%s", lb, plan.Unit.Content)
				}
				return
			}
			if lb == nil || !strings.Contains(lb.Data, "<configuration") || lb.Source != "" {
				t.Errorf("logback file = %+v, want the rendered logback-spring.xml", lb)
			}
			if !strings.Contains(plan.Unit.Content, line) {
				t.Errorf("unit missing %q:\n%s", line, plan.Unit.Content)
			}
		})
	}
}

// TestGeneratePodmanNeverReadsTheStores pins that the TLS stores ride on the
// plan by name only: generate (and remove, which renders the same plan) must
// work on a host where the store files are not present, so only deploy reads
// them. The unit mounts each from the secret store at the path application.yml
// reads it from.
func TestGeneratePodmanNeverReadsTheStores(t *testing.T) {
	envData := `image:
  name: img
  tag: v1
tls:
  truststore:
    file: ./certs/trust.p12
    password: ts
    type: PKCS12
  keystore:
    file: ./certs/key.p12
    password: ks
    type: PKCS12
podman:
  command: podman
  name: c
`
	req := Request{Env: &File{Name: "env.yaml", Data: []byte(envData)}, Workflows: synthWorkflowFiles(1)}
	res := Resolver{Env: func(string) (string, bool) { return "v", true }, Rand: fixedStatusRand, ReadFile: func(p string) ([]byte, error) {
		t.Errorf("generate read %s; only deploy may read a store", p)
		return nil, fs.ErrNotExist
	}}
	plan, errs, _ := GeneratePodman(req, res)
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	want := []PodmanFile{
		{StoreName: "c-tls-truststore", Source: "./certs/trust.p12", Field: "tls.truststore.file"},
		{StoreName: "c-tls-keystore", Source: "./certs/key.p12", Field: "tls.keystore.file"},
	}
	if len(plan.Files) < 3 || !slices.Equal(plan.Files[1:3], want) {
		t.Errorf("store files = %+v, want %+v after application.yml", plan.Files, want)
	}
	for _, line := range []string{
		"Secret=c-tls-truststore,type=mount,target=" + spec.DefaultStoresMountPath + "/trust.p12\n",
		"Secret=c-tls-keystore,type=mount,target=" + spec.DefaultStoresMountPath + "/key.p12\n",
	} {
		if !strings.Contains(plan.Unit.Content, line) {
			t.Errorf("unit missing %q:\n%s", line, plan.Unit.Content)
		}
	}
	if strings.Contains(plan.Unit.Content, "Volume=") {
		t.Errorf("a store must not be bind-mounted:\n%s", plan.Unit.Content)
	}
}

// TestGeneratePodmanIgnoresBaseDir pins that podman.base-dir decides nothing
// any more: with it set -- even to a value the old host-path gate refused,
// holding an unset variable -- generate renders the same unit and plan as
// without it, with no error and no warning (validate alone notes it).
func TestGeneratePodmanIgnoresBaseDir(t *testing.T) {
	const env = "image:\n  name: img\n  tag: v1\npodman:\n  command: podman\n  name: c\n"
	res := Resolver{Env: func(n string) (string, bool) { return "v", n != "UNSET" }, Rand: fixedStatusRand}
	gen := func(envData string) (PodmanPlan, []Issue, []Issue) {
		req := Request{Env: &File{Name: "env.yaml", Data: []byte(envData)}, Workflows: synthWorkflowFiles(1)}
		return GeneratePodman(req, res)
	}
	plain, errs, warns := gen(env)
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	with, errs2, warns2 := gen(env + "  base-dir: \"/old dir/${UNSET}\"\n")
	if len(errs2) > 0 {
		t.Fatalf("base-dir must not be an error, got %v", errs2)
	}
	if with.Unit != plain.Unit || !slices.Equal(with.Files, plain.Files) {
		t.Errorf("base-dir changed the plan:\n%s\nvs\n%s", with.Unit.Content, plain.Unit.Content)
	}
	if !slices.Equal(warns2, warns) {
		t.Errorf("base-dir must add no warning outside validate: %v", warns2)
	}
}

func issuesContain(errs []Issue, sub string) bool {
	for _, e := range errs {
		if strings.Contains(e.Msg, sub) {
			return true
		}
	}
	return false
}

// TestGenerateMissingTargetSection covers the nil-section guards: an env.yaml
// that parses (it has a tls: section) but omits the requested target section
// must fail loud with an actionable message rather than emit an empty artifact.
func TestGenerateMissingTargetSection(t *testing.T) {
	envData := `tls:
  truststore:
    file: ./certs/truststore.jks
    password: ts
    type: JKS
`
	cases := []struct {
		name string
		want string
		gen  func(Request, Resolver) []Issue
	}{
		{"kubernetes", "kubernetes target requires a 'kubernetes:' section in env.yaml",
			func(r Request, res Resolver) []Issue { _, e, _ := GenerateKubernetes(r, res, KubeOpts{}); return e }},
		{"docker", "docker target requires a 'docker:' section in env.yaml",
			func(r Request, res Resolver) []Issue { _, e, _ := GenerateDocker(r, res); return e }},
		{"podman", "podman target requires a 'podman:' section in env.yaml",
			func(r Request, res Resolver) []Issue { _, e, _ := GeneratePodman(r, res); return e }},
	}
	for _, c := range cases {
		req := Request{Env: &File{Name: "env.yaml", Data: []byte(envData)}, Workflows: synthWorkflowFiles(1)}
		errs := c.gen(req, Resolver{})
		if !issuesContain(errs, c.want) {
			t.Errorf("%s: want %q, got %v", c.name, c.want, errs)
		}
	}
}

// TestGenValidateStoresWarning covers the kubernetes credentials-secret path
// (create.name only -- the removed source/variables/values-file fields are
// covered by validate's own tests) together with the TLS-without-stores
// advisory warning.
func TestGenValidateStoresWarning(t *testing.T) {
	wfData := `
source:
  solace:
    host: tcps://b
    msg-vpn: v
    client-username: u
    client-password: p
    queue: IN
target:
  mq:
    conn-name: h(1414)
    queue-manager: QM
    channel: C
    user: u
    password: p
    queue: OUT
`
	envData := `image:
  name: img
  tag: v1
kubernetes:
  deployment:
    name: c
    namespace: ns
  secrets:
    credentials:
      create: true
`
	req := Request{
		Env:       &File{Name: "env.yaml", Data: []byte(envData)},
		Workflows: []File{{Name: "10.yaml", Data: []byte(wfData)}},
	}
	errs, warns := Validate(req, Resolver{})
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	wantWarn := "a TLS/mTLS connection exists but secrets.stores is omitted; the store files will be missing at runtime"
	if len(warns) != 1 || warns[0].File != fileEnv || warns[0].Msg != wantWarn {
		t.Fatalf("warns = %+v, want exactly one {%q, %q}", warns, fileEnv, wantWarn)
	}
}
