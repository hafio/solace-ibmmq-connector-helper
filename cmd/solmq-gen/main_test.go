package main

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const validWF = `
source:
  solace:
    host: tcp://b:55555
    msg-vpn: prod
    client-username: connector
    client-password: ${SOL_PASSWORD}
    queue: Q.IN
target:
  mq:
    conn-name: mqhost(1414)
    queue-manager: QM1
    channel: CH
    user: app
    password: ${MQ_PASSWORD}
    queue: OUT
`

const invalidWF = `
source:
  mq:
    conn-name: not-a-conn
    queue-manager: QM1
    channel: CH
    user: u
    password: p
    queue: Q
target:
  solace:
    host: tcp://b:1
    msg-vpn: v
    client-username: u
    client-password: p
    queue: Q
`

const kubeEnv = `
deployment:
  name: solmq-connector
  namespace: solace-connectors
  image: img:1
  replicas: 1
service:
  enabled: true
  port: 8090
secrets:
  credentials:
    create:
      name: creds
      source: env
      variables: [SOL_PASSWORD, MQ_PASSWORD]
`

const kubeFileSrc = `
deployment:
  name: solmq-connector
  namespace: solace-connectors
  image: img:1
  replicas: 1
secrets:
  credentials:
    create:
      name: creds
      source: file
      values-file: secrets.env
`

func write(t *testing.T, dir, name, body string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
}

func specDir(t *testing.T, wf string) string {
	dir := t.TempDir()
	write(t, dir, "10.yaml", wf)
	return dir
}

func captureStdout(t *testing.T, f func()) string {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w
	f()
	_ = w.Close()
	os.Stdout = old
	b, _ := io.ReadAll(r)
	return string(b)
}

func TestRunConfigToFile(t *testing.T) {
	dir := specDir(t, validWF)
	out := filepath.Join(dir, "out.yml")
	if code := run([]string{"config", dir, "-o", out}); code != 0 {
		t.Fatalf("exit=%d", code)
	}
	b, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), "spring:") || !strings.Contains(string(b), "type: undefined") {
		t.Errorf("unexpected output:\n%s", b)
	}
}

func TestRunConfigToStdout(t *testing.T) {
	dir := specDir(t, validWF)
	var code int
	out := captureStdout(t, func() { code = run([]string{"config", dir}) })
	if code != 0 {
		t.Fatalf("exit=%d", code)
	}
	if !strings.Contains(out, "spring:") {
		t.Errorf("stdout missing spring:\n%s", out)
	}
}

func TestRunConfigFailFast(t *testing.T) {
	dir := specDir(t, invalidWF)
	if code := run([]string{"config", dir}); code != 1 {
		t.Fatalf("want exit 1 for invalid spec, got %d", code)
	}
}

func TestRunValidateOKAndErrors(t *testing.T) {
	if code := run([]string{"validate", specDir(t, validWF)}); code != 0 {
		t.Fatalf("valid spec should validate, exit=%d", code)
	}
	if code := run([]string{"validate", specDir(t, invalidWF)}); code != 1 {
		t.Fatalf("invalid spec should fail validate, exit=%d", code)
	}
}

func TestRunDeployEnvSource(t *testing.T) {
	t.Setenv("SOL_PASSWORD", "s")
	t.Setenv("MQ_PASSWORD", "m")
	dir := specDir(t, validWF)
	write(t, dir, "kubernetes.yaml", kubeEnv)
	out := filepath.Join(dir, "manifests.yml")
	if code := run([]string{"deploy", dir, "-o", out}); code != 0 {
		t.Fatalf("exit=%d", code)
	}
	b, _ := os.ReadFile(out)
	for _, w := range []string{"kind: ConfigMap", "kind: Deployment", "kind: Service", "kind: Secret", "SOL_PASSWORD:"} {
		if !strings.Contains(string(b), w) {
			t.Errorf("missing %q in manifests", w)
		}
	}
}

func TestRunDeployFileSource(t *testing.T) {
	dir := specDir(t, validWF)
	write(t, dir, "kubernetes.yaml", kubeFileSrc)
	write(t, dir, "secrets.env", "SOL_PASSWORD=a\nMQ_PASSWORD=b\n")
	out := filepath.Join(dir, "m.yml")
	if code := run([]string{"deploy", dir, "-o", out}); code != 0 {
		t.Fatalf("exit=%d", code)
	}
	b, _ := os.ReadFile(out)
	if !strings.Contains(string(b), `SOL_PASSWORD: "a"`) {
		t.Errorf("values-file not read via fileReader:\n%s", b)
	}
}

func TestRunDeployMissingKube(t *testing.T) {
	if code := run([]string{"deploy", specDir(t, validWF)}); code != 1 {
		t.Fatalf("deploy without kubernetes.yaml should fail, got %d", code)
	}
}

func TestRunUsageAndErrors(t *testing.T) {
	if code := run(nil); code != 2 {
		t.Errorf("no args -> 2, got %d", code)
	}
	if code := run([]string{"bogus"}); code != 2 {
		t.Errorf("unknown subcommand -> 2, got %d", code)
	}
	if code := run([]string{"--help"}); code != 0 {
		t.Errorf("help -> 0, got %d", code)
	}
	if code := run([]string{"config"}); code != 2 {
		t.Errorf("missing dir -> 2, got %d", code)
	}
	if code := run([]string{"config", "-nope", "x"}); code != 2 {
		t.Errorf("bad flag -> 2, got %d", code)
	}
}

func TestRunConfigReadError(t *testing.T) {
	if code := run([]string{"config", filepath.Join(t.TempDir(), "does-not-exist")}); code != 1 {
		t.Errorf("missing dir scan error -> 1, got %d", code)
	}
}

func TestRunConfigFlagsBeforeDir(t *testing.T) {
	dir := specDir(t, validWF)
	out := filepath.Join(dir, "b.yml")
	if code := run([]string{"config", "-o", out, dir}); code != 0 {
		t.Fatalf("flags-before-dir exit=%d", code)
	}
	if _, err := os.Stat(out); err != nil {
		t.Errorf("flags-before-dir: output not written: %v", err)
	}
}

func TestRunConfigFlagsAfterDir(t *testing.T) {
	dir := specDir(t, validWF)
	out := filepath.Join(dir, "a.yml")
	if code := run([]string{"config", dir, "-o", out}); code != 0 {
		t.Fatalf("flags-after-dir exit=%d", code)
	}
	if _, err := os.Stat(out); err != nil {
		t.Errorf("flags-after-dir (documented usage): output not written: %v", err)
	}
}

func TestRunValidateCustomKubeFlag(t *testing.T) {
	t.Setenv("SOL_PASSWORD", "s")
	t.Setenv("MQ_PASSWORD", "m")
	dir := specDir(t, validWF)
	write(t, dir, "myk8s.yaml", kubeEnv)
	// -k after dir; myk8s.yaml treated as the kube settings (excluded from workflows)
	if code := run([]string{"validate", dir, "-k", "myk8s.yaml"}); code != 0 {
		t.Fatalf("validate with -k exit=%d", code)
	}
}

const loopWF = `
source:
  mq: {conn-name: h(1414), queue-manager: QM, channel: C, user: u, password: p, queue: SAME}
target:
  mq: {conn-name: h(1414), queue-manager: QM, channel: C, user: u, password: p, queue: SAME}
`

func TestRunConfigEmitsWarnings(t *testing.T) {
	// source==target binder+destination -> message-loop warning exercises printWarnings
	dir := specDir(t, loopWF)
	if code := run([]string{"config", dir, "-o", filepath.Join(dir, "o.yml")}); code != 0 {
		t.Fatalf("config with warnings should still succeed, exit=%d", code)
	}
}

func TestRunConfigEmitWriteError(t *testing.T) {
	dir := specDir(t, validWF)
	bad := filepath.Join(dir, "no-such-subdir", "o.yml") // parent dir does not exist
	if code := run([]string{"config", dir, "-o", bad}); code != 1 {
		t.Fatalf("emit to unwritable path should return 1, got %d", code)
	}
}
