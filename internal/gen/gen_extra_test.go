package gen

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/consolidate"
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

func TestNamesAndPaths(t *testing.T) {
	if pathIn("", "a") != "a" || pathIn("/base/", "a") != "/base/a" || pathIn("/base", "a") != "/base/a" {
		t.Error("pathIn")
	}
}

func TestTargetMounts(t *testing.T) {
	tls := spec.TLSConfig{
		Truststore: &spec.Store{File: "certs/t.jks"},
		Keystore:   &spec.Store{File: "certs/k.jks"},
	}
	res := Resolver{Abs: func(p string) string { return "/abs/" + p }}
	// The store bind-mount target is always the fixed in-container dir; the
	// supplied stores.MountPath ("/mnt") is deliberately ignored (a non-default
	// value is rejected in validate). Only the host Source comes from res.Abs.
	sm, lm := targetMounts(tls, &spec.StoresMount{MountPath: "/mnt"}, &spec.LibsMount{Dir: "libs", MountPath: "/libs"}, res)
	if len(sm) != 2 {
		t.Fatalf("store mounts=%d want 2", len(sm))
	}
	if sm[0].Source != "/abs/certs/t.jks" || sm[0].Target != spec.DefaultStoresMountPath+"/t.jks" {
		t.Errorf("store mount 0 = %+v", sm[0])
	}
	if sm[1].Target != spec.DefaultStoresMountPath+"/k.jks" {
		t.Errorf("store mount 1 = %+v", sm[1])
	}
	if lm == nil || lm.Source != "/abs/libs" || lm.Target != "/libs" {
		t.Errorf("libs mount = %+v", lm)
	}
	// Opt-out: nil stores and libs yield nothing (bind mounts are opt-in).
	if sm2, lm2 := targetMounts(tls, nil, nil, res); sm2 != nil || lm2 != nil {
		t.Errorf("nil sections should yield no mounts: %v %v", sm2, lm2)
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
	envData := "kubernetes:\n  deployment:\n    name: solmq\n    namespace: ns\n    image: img\n  service:\n    enabled: true\n    port: 8090\n"
	req := Request{
		Env:       &File{Name: "env.yaml", Data: []byte(envData)},
		Workflows: synthWorkflowFiles(21),
	}
	out, errs, _ := GenerateKubernetes(req, Resolver{})
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
// ${STABLE} placeholder, with exactly one carve-out: security defaults to
// effectively enabled (no security: block here), so
// consolidate.applyStatusAccess unconditionally appends the reserved
// spec.StatusUserName account carrying its password as a literal -- the
// generated status script reads that literal back out of application.yml at
// run time, so it has nothing else to read. This guards both invariants
// end-to-end through Config: every ordinary password is a placeholder, and
// the one literal that exists is exactly the expected status-account value.
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
	if !strings.Contains(out, "${SOL_CONN_1_CLIENT_USERNAME}") || !strings.Contains(out, "${MQ_CONN_1_USER}") {
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
	envData := `tls:
  truststore:
    file: ./certs/truststore.jks
    password: ts
    type: JKS
docker:
  command: docker
  image: solace/connector:9.9
  name: solmq-connector
  restart: unless-stopped
  ports:
    - 8090
  timezone: UTC
  stores:
    mount-path: /app/external/classpath/truststores
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

	// One workflow, inline solace source (auto sol-conn-1) + inline mq target
	// (auto mq-conn-1): four credential positions, all literal.
	want := []consolidate.SecretRef{
		{Stable: "SOL_CONN_1_CLIENT_USERNAME", Literal: "u"},
		{Stable: "SOL_CONN_1_CLIENT_PASSWORD", Literal: "p"},
		{Stable: "MQ_CONN_1_USER", Literal: "u"},
		{Stable: "MQ_CONN_1_PASSWORD", Literal: "p"},
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
	}
}

func TestGeneratePodmanRunAndQuadlet(t *testing.T) {
	envData := `podman:
  command: podman
  mode: run
  image: solace/connector:9.9
  name: solmq-connector
  restart: unless-stopped
  ports:
    - 8090
  timezone: UTC
`
	req := Request{Env: &File{Name: "env.yaml", Data: []byte(envData)}, Workflows: synthWorkflowFiles(1)}
	res := Resolver{Env: func(string) (string, bool) { return "v", true }}

	// mode: run -> a run script, no quadlet unit.
	plan, errs, _ := GeneratePodman(req, res, PodmanOpts{})
	if len(errs) > 0 {
		t.Fatalf("run: unexpected errors: %v", errs)
	}
	if plan.Mode != spec.PodmanModeRun || plan.RunScript == "" || plan.Unit != (podmangen.Unit{}) {
		t.Errorf("run plan mode=%q script=%d unit=%+v", plan.Mode, len(plan.RunScript), plan.Unit)
	}
	if plan.AppYAML.Name != "solmq-connector-application.yml" {
		t.Errorf("app yaml = %+v", plan.AppYAML)
	}
	if plan.Service != "solmq-connector.service" {
		t.Errorf("service = %q", plan.Service)
	}
	if len(plan.Secrets) != 4 {
		t.Fatalf("secrets = %+v, want 4 entries", plan.Secrets)
	}
	// The run script loads each secret into podman's store, namespaced by the
	// container, before any `podman run`; it never carries the value itself.
	for _, s := range plan.Secrets {
		store := PodmanSecretStoreName("solmq-connector", s.Stable)
		if !strings.Contains(plan.RunScript, "podman secret create "+store) {
			t.Errorf("run script missing secret-create for %q:\n%s", store, plan.RunScript)
		}
		if !strings.Contains(plan.RunScript, "--secret "+store+",type=mount,target="+s.Stable) {
			t.Errorf("run script missing --secret mount for %q:\n%s", store, plan.RunScript)
		}
	}

	// ForceQuadlet -> a quadlet unit, no run script (deploy path).
	q, errs, _ := GeneratePodman(req, res, PodmanOpts{ForceQuadlet: true, BaseDir: "/base"})
	if len(errs) > 0 {
		t.Fatalf("quadlet: unexpected errors: %v", errs)
	}
	if q.Mode != spec.PodmanModeQuadlet || q.Unit == (podmangen.Unit{}) || q.RunScript != "" {
		t.Errorf("quadlet plan mode=%q unit=%+v script=%d", q.Mode, q.Unit, len(q.RunScript))
	}
	if len(q.Secrets) != 4 {
		t.Errorf("quadlet secrets = %+v, want 4 entries", q.Secrets)
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
	envData := "kubernetes:\n  command: kubectl\n  deployment:\n    name: solmq\n    namespace: ns\n    image: img\n"
	req := Request{Env: &File{Name: "env.yaml", Data: []byte(envData)}, Workflows: synthWorkflowFiles(1)}
	out, errs, _ := GenerateKubernetes(req, Resolver{Rand: fixedStatusRand})
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
	envData := "docker:\n  command: docker\n  image: img\n  name: solmq-connector\n"
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

// TestGeneratePodmanCarriesStatusScript pins the podman wiring:
// PodmanPlan.StatusScript names <name>-status and carries the rendered
// script, and the on-disk mount path is BaseDir-resolved exactly like
// AppYAML.
func TestGeneratePodmanCarriesStatusScript(t *testing.T) {
	envData := "podman:\n  command: podman\n  mode: run\n  image: img\n  name: solmq-connector\n"
	req := Request{Env: &File{Name: "env.yaml", Data: []byte(envData)}, Workflows: synthWorkflowFiles(1)}
	res := Resolver{Env: func(string) (string, bool) { return "v", true }, Rand: fixedStatusRand}

	plan, errs, _ := GeneratePodman(req, res, PodmanOpts{BaseDir: "/base"})
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if plan.StatusScript.Name != "solmq-connector-status" {
		t.Errorf("StatusScript.Name = %q, want %q", plan.StatusScript.Name, "solmq-connector-status")
	}
	if !strings.Contains(plan.StatusScript.Data, "USER_NAME="+spec.StatusUserName) {
		t.Errorf("StatusScript.Data missing the rendered script:\n%s", plan.StatusScript.Data)
	}
	// Same BaseDir resolution as AppYAML (pathIn), not a bare name.
	if want := "/base/solmq-connector-application.yml"; !strings.Contains(plan.RunScript, want) {
		t.Errorf("run script missing BaseDir-resolved AppYAML mount %q:\n%s", want, plan.RunScript)
	}
	if want := "/base/solmq-connector-status:" + statusscript.ContainerPath; !strings.Contains(plan.RunScript, want) {
		t.Errorf("run script missing BaseDir-resolved status mount %q:\n%s", want, plan.RunScript)
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
			func(r Request, res Resolver) []Issue { _, e, _ := GenerateKubernetes(r, res); return e }},
		{"docker", "docker target requires a 'docker:' section in env.yaml",
			func(r Request, res Resolver) []Issue { _, e, _ := GenerateDocker(r, res); return e }},
		{"podman", "podman target requires a 'podman:' section in env.yaml",
			func(r Request, res Resolver) []Issue { _, e, _ := GeneratePodman(r, res, PodmanOpts{}); return e }},
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
	envData := `kubernetes:
  deployment:
    name: c
    namespace: ns
    image: img
  secrets:
    credentials:
      create:
        name: s
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
