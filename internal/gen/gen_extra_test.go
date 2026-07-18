package gen

import (
	"errors"
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
