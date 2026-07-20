package gen

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-gen/internal/deploy"
	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-gen/internal/spec"
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
	if b64([]byte("ABC")) != "QUJD" {
		t.Errorf("b64=%q", b64([]byte("ABC")))
	}
	if iss := toIssues([]string{"a", "b"}); len(iss) != 2 || iss[0].Msg != "a" {
		t.Errorf("toIssues=%v", iss)
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

func TestDeploySharding(t *testing.T) {
	req := Request{
		Workflows:  synthWorkflowFiles(21),
		Kubernetes: &File{Name: "kubernetes.yaml", Data: []byte("deployment:\n  name: solmq\n  namespace: ns\n  image: img\nservice:\n  enabled: true\n  port: 8090\n")},
	}
	out, errs, _ := Deploy(req, Resolver{})
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
	kubeData := `
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
		Workflows:  []File{{Name: "10.yaml", Data: []byte(wfData)}},
		Kubernetes: &File{Name: "kubernetes.yaml", Data: []byte(kubeData)},
	}
	res := Resolver{ReadFile: func(string) ([]byte, error) { return []byte("SOL=x\n"), nil }}
	if _, warns := Validate(req, res); warns == nil {
		_ = warns // path exercised
	}
	k, err := spec.ParseKubernetes([]byte(kubeData))
	if err != nil {
		t.Fatal(err)
	}
	keys, iss := valuesFileKeys(k, res)
	if iss != nil || !keys["SOL"] {
		t.Fatalf("keys=%v iss=%v", keys, iss)
	}
	if kk, ii := valuesFileKeys(nil, res); kk != nil || ii != nil {
		t.Error("nil kube -> nil,nil")
	}
}
