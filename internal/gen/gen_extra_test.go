package gen

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/deploy"
	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/spec"
)

func TestLooksDotenv(t *testing.T) {
	if !looksDotenv("A=1\nB=2\n") {
		t.Error("dotenv")
	}
	if looksDotenv("a: 1\n") {
		t.Error("yaml mapping should not be dotenv")
	}
	if looksDotenv("# c\n\n") {
		t.Error("comments/blank only -> false")
	}
	if !looksDotenv("URL=http://x:1\n") {
		t.Error("value with colon after = is dotenv")
	}
	if looksDotenv("KEY: v=1\n") {
		t.Error("colon before = -> yaml")
	}
}

func TestParseValuesDotenv(t *testing.T) {
	kvs, err := parseValues([]byte("# c\nA=1\n\nB = two \n"))
	if err != nil {
		t.Fatal(err)
	}
	if len(kvs) != 2 || kvs[0] != (deploy.KV{Key: "A", Val: "1"}) || kvs[1].Key != "B" || kvs[1].Val != "two" {
		t.Fatalf("kvs=%+v", kvs)
	}
}

func TestParseValuesYAML(t *testing.T) {
	kvs, err := parseValues([]byte("A: 1\nB: two\n"))
	if err != nil {
		t.Fatal(err)
	}
	if len(kvs) != 2 || kvs[0].Key != "A" || kvs[1].Val != "two" {
		t.Fatalf("kvs=%+v", kvs)
	}
	if _, err := parseValues([]byte("- just\n- a\n")); err == nil {
		t.Error("sequence should error (not a mapping)")
	}
}

func TestResolveCredEnv(t *testing.T) {
	c := &spec.CredCreate{Source: spec.SourceEnv, Variables: []string{"X", "Y"}}
	env := map[string]string{"X": "1", "Y": "2"}
	kvs, err := resolveCred(c, Resolver{Env: func(k string) (string, bool) { v, ok := env[k]; return v, ok }})
	if err != nil || len(kvs) != 2 {
		t.Fatalf("kvs=%v err=%v", kvs, err)
	}
	if _, err := resolveCred(c, Resolver{Env: func(string) (string, bool) { return "", false }}); err == nil {
		t.Error("missing env var should error")
	}
	kvs3, err := resolveCred(c, Resolver{})
	if err != nil || len(kvs3) != 2 || kvs3[0].Val != "" {
		t.Fatalf("nil env: %v %v", kvs3, err)
	}
}

func TestResolveCredFileAndBadSource(t *testing.T) {
	c := &spec.CredCreate{Source: spec.SourceFile, ValuesFile: "v.env"}
	kvs, err := resolveCred(c, Resolver{ReadFile: func(string) ([]byte, error) { return []byte("A=1\n"), nil }})
	if err != nil || len(kvs) != 1 || kvs[0].Key != "A" {
		t.Fatalf("kvs=%v err=%v", kvs, err)
	}
	if _, err := resolveCred(c, Resolver{}); err == nil {
		t.Error("no ReadFile should error")
	}
	if _, err := resolveCred(c, Resolver{ReadFile: func(string) ([]byte, error) { return nil, errors.New("boom") }}); err == nil {
		t.Error("read error should propagate")
	}
	if _, err := resolveCred(&spec.CredCreate{Source: "weird"}, Resolver{}); err == nil {
		t.Error("unknown source should error")
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

func TestBaseNameB64ToIssues(t *testing.T) {
	if baseName(`a\b\c.jks`) != "c.jks" || baseName("a/b/c.jks") != "c.jks" || baseName("x") != "x" {
		t.Error("baseName")
	}
	if iss := toIssues([]string{"a", "b"}); len(iss) != 2 || iss[0].Msg != "a" {
		t.Errorf("toIssues=%v", iss)
	}
}

// ---- names, paths, mounts (docker/podman plumbing) --------------------------

func TestNamesAndPaths(t *testing.T) {
	if credEnvFileName("n", nil) != "" {
		t.Error("nil creds -> empty")
	}
	if credEnvFileName("n", &spec.CredentialsSecret{Existing: "e.env"}) != "e.env" {
		t.Error("existing wins")
	}
	if credEnvFileName("n", &spec.CredentialsSecret{Create: &spec.CredCreate{}}) != "n.env" {
		t.Error("create -> <name>.env")
	}
	if pathIn("", "a") != "a" || pathIn("/base/", "a") != "/base/a" || pathIn("/base", "a") != "/base/a" {
		t.Error("pathIn")
	}
	if instanceName("b", 0, 1) != "b" || instanceName("b", 0, 2) != "b-1" || instanceName("b", 1, 2) != "b-2" {
		t.Error("instanceName")
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

func TestResolveCredentialsAndEnvFileContent(t *testing.T) {
	if kvs, err := ResolveCredentials(nil, Resolver{}); err != nil || kvs != nil {
		t.Errorf("nil creds -> nil,nil: %v %v", kvs, err)
	}
	if kvs, err := ResolveCredentials(&spec.CredentialsSecret{Existing: "x.env"}, Resolver{}); err != nil || kvs != nil {
		t.Errorf("existing -> nil,nil: %v %v", kvs, err)
	}
	creds := &spec.CredentialsSecret{Create: &spec.CredCreate{Source: spec.SourceEnv, Variables: []string{"A", "B"}}}
	env := map[string]string{"A": "1", "B": "2"}
	kvs, err := ResolveCredentials(creds, Resolver{Env: func(k string) (string, bool) { v, ok := env[k]; return v, ok }})
	if err != nil || len(kvs) != 2 {
		t.Fatalf("kvs=%v err=%v", kvs, err)
	}
	if EnvFileContent(kvs) != "A=1\nB=2\n" {
		t.Errorf("env-file = %q", EnvFileContent(kvs))
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

func TestShardWorkflows(t *testing.T) {
	for _, c := range []struct{ n, want int }{{0, 1}, {1, 1}, {20, 1}, {21, 2}, {40, 2}, {41, 3}} {
		if got := len(shardWorkflows(make([]spec.Workflow, c.n))); got != c.want {
			t.Errorf("n=%d chunks=%d want %d", c.n, got, c.want)
		}
	}
	chunks := shardWorkflows(synthWorkflows(21))
	if len(chunks) != 2 || len(chunks[0]) != 20 || len(chunks[1]) != 1 {
		t.Fatalf("chunk sizes = %d then %d,%d", len(chunks), len(chunks[0]), len(chunks[1]))
	}
	if chunks[1][0].File != "wf-20.yaml" {
		t.Errorf("chunk 2 first workflow = %q, want wf-20.yaml", chunks[1][0].File)
	}
}

func TestBuildShardsLeaderQueueSuffix(t *testing.T) {
	d := &spec.Defaults{LeaderElection: spec.LeaderElection{
		Present: true, Mode: spec.LeaderActiveStby, Queue: "mgmt-q",
		Session: &spec.Side{System: spec.SystemSolace, Host: "tcp://b", MsgVPN: "v", ClientUser: "u", ClientPass: "p"},
	}}
	shards, _ := buildShards(synthWorkflows(21), d, true)
	if len(shards) != 2 {
		t.Fatalf("shards=%d want 2", len(shards))
	}
	if shards[0].model.LeaderElection.Queue != "mgmt-q-1" || shards[1].model.LeaderElection.Queue != "mgmt-q-2" {
		t.Errorf("queues = %q, %q want mgmt-q-1, mgmt-q-2", shards[0].model.LeaderElection.Queue, shards[1].model.LeaderElection.Queue)
	}
	// Single instance keeps the queue unchanged.
	single, _ := buildShards(synthWorkflows(3), d, true)
	if len(single) != 1 || single[0].model.LeaderElection.Queue != "mgmt-q" {
		t.Errorf("single-instance queue = %q want mgmt-q", single[0].model.LeaderElection.Queue)
	}
}

// TestBuildShardsLeaderActiveActive pins the one field that actually
// distinguishes active_active from active_standby: the literal mode string
// carried through the model and into the rendered application.yml (queue
// suffixing is mode-agnostic; buildShards never branches on Mode).
func TestBuildShardsLeaderActiveActive(t *testing.T) {
	d := &spec.Defaults{LeaderElection: spec.LeaderElection{
		Present: true, Mode: spec.LeaderActiveActive, Queue: "mgmt-q",
		Session: &spec.Side{System: spec.SystemSolace, Host: "tcp://b", MsgVPN: "v", ClientUser: "u", ClientPass: "p"},
	}}
	shards, _ := buildShards(synthWorkflows(21), d, true)
	if len(shards) != 2 {
		t.Fatalf("shards=%d want 2", len(shards))
	}
	for i, s := range shards {
		if s.model.LeaderElection == nil || s.model.LeaderElection.Mode != spec.LeaderActiveActive {
			t.Errorf("shard %d LeaderElection = %+v want mode %q", i, s.model.LeaderElection, spec.LeaderActiveActive)
		}
	}
	if shards[0].model.LeaderElection.Queue != "mgmt-q-1" || shards[1].model.LeaderElection.Queue != "mgmt-q-2" {
		t.Errorf("queues = %q, %q want mgmt-q-1, mgmt-q-2", shards[0].model.LeaderElection.Queue, shards[1].model.LeaderElection.Queue)
	}
	if !strings.Contains(shards[0].appYAML, "mode: active_active") {
		t.Errorf("rendered appYAML missing 'mode: active_active':\n%s", shards[0].appYAML)
	}
	if strings.Contains(shards[0].appYAML, "active_standby") {
		t.Errorf("rendered appYAML should not mention active_standby:\n%s", shards[0].appYAML)
	}
}

func TestKubernetesSharding(t *testing.T) {
	envData := "kubernetes:\n  deployment:\n    name: solmq\n    namespace: ns\n    image: img\n  service:\n    enabled: true\n    port: 8090\n"
	req := Request{
		Env:       &File{Name: "env.yaml", Data: []byte(envData)},
		Workflows: synthWorkflowFiles(21),
	}
	out, errs, _ := GenerateKubernetes(req, Resolver{})
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	for _, c := range []struct {
		kind string
		want int
	}{{"kind: Namespace", 1}, {"kind: ConfigMap", 2}, {"kind: Deployment", 2}, {"kind: Service", 2}} {
		if got := strings.Count(out, c.kind); got != c.want {
			t.Errorf("%q count = %d, want %d", c.kind, got, c.want)
		}
	}
	for _, want := range []string{"name: solmq-1\n", "name: solmq-2\n", "name: solmq-1-config", "name: solmq-2-config"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q", want)
		}
	}
}

func TestConfigSharding(t *testing.T) {
	outs, errs, _ := Config(Request{Workflows: synthWorkflowFiles(21)}, Resolver{})
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if len(outs) != 2 {
		t.Fatalf("instances = %d, want 2", len(outs))
	}
	if !strings.Contains(outs[0], "input-19") || strings.Contains(outs[0], "input-20") {
		t.Errorf("instance 1 should hold workflows 0..19")
	}
	// Instance 2 renumbers its lone workflow from 0.
	if !strings.Contains(outs[1], "input-0") || strings.Contains(outs[1], "input-1") {
		t.Errorf("instance 2 should renumber from 0:\n%s", outs[1])
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
  secrets:
    credentials:
      existing: solmq.env
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
	if plan.EnvFileName != "solmq.env" {
		t.Errorf("env-file = %q want solmq.env (existing)", plan.EnvFileName)
	}
	if !strings.Contains(plan.Compose, "solace/connector:9.9") {
		t.Errorf("compose missing image:\n%s", plan.Compose)
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
  secrets:
    credentials:
      create:
        name: solmq-credentials
        source: env
        variables:
          - FOO
`
	req := Request{Env: &File{Name: "env.yaml", Data: []byte(envData)}, Workflows: synthWorkflowFiles(1)}
	res := Resolver{Env: func(string) (string, bool) { return "v", true }}

	// mode: run -> a run script, no quadlet units.
	plan, errs, _ := GeneratePodman(req, res, PodmanOpts{})
	if len(errs) > 0 {
		t.Fatalf("run: unexpected errors: %v", errs)
	}
	if plan.Mode != spec.PodmanModeRun || plan.RunScript == "" || len(plan.Units) != 0 {
		t.Errorf("run plan mode=%q script=%d units=%d", plan.Mode, len(plan.RunScript), len(plan.Units))
	}
	if plan.EnvFileName != "solmq-connector.env" {
		t.Errorf("env-file = %q want solmq-connector.env", plan.EnvFileName)
	}
	if len(plan.AppYAMLs) != 1 || plan.AppYAMLs[0].Name != "solmq-connector-application.yml" {
		t.Errorf("app yamls = %+v", plan.AppYAMLs)
	}
	if len(plan.Services) != 1 || plan.Services[0] != "solmq-connector.service" {
		t.Errorf("services = %+v", plan.Services)
	}

	// ForceQuadlet -> quadlet units, no run script (deploy path).
	q, errs, _ := GeneratePodman(req, res, PodmanOpts{ForceQuadlet: true, BaseDir: "/base"})
	if len(errs) > 0 {
		t.Fatalf("quadlet: unexpected errors: %v", errs)
	}
	if q.Mode != spec.PodmanModeQuadlet || len(q.Units) == 0 || q.RunScript != "" {
		t.Errorf("quadlet plan mode=%q units=%d script=%d", q.Mode, len(q.Units), len(q.RunScript))
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

func TestGenValidateAndValuesFileKeys(t *testing.T) {
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
        source: file
        values-file: v.env
`
	req := Request{
		Env:       &File{Name: "env.yaml", Data: []byte(envData)},
		Workflows: []File{{Name: "10.yaml", Data: []byte(wfData)}},
	}
	res := Resolver{ReadFile: func(string) ([]byte, error) { return []byte("SOL=x\n"), nil }}
	errs, warns := Validate(req, res)
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	wantWarn := "a TLS/mTLS connection exists but secrets.stores is omitted; the store files will be missing at runtime"
	if len(warns) != 1 || warns[0].File != fileEnv || warns[0].Msg != wantWarn {
		t.Fatalf("warns = %+v, want exactly one {%q, %q}", warns, fileEnv, wantWarn)
	}
	e, err := spec.ParseEnv([]byte(envData))
	if err != nil {
		t.Fatal(err)
	}
	keys, iss := valuesFileKeys(e.Kubernetes, res)
	if iss != nil || !keys["SOL"] {
		t.Fatalf("keys=%v iss=%v", keys, iss)
	}
	if kk, ii := valuesFileKeys(nil, res); kk != nil || ii != nil {
		t.Error("nil kube -> nil,nil")
	}
}
