package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/gen"
	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/podmangen"
	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/runner"
	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/spec"
)

// ---- fixtures -----------------------------------------------------------------

// validWF is a fully inline (no conn-ref, no TLS) workflow: IBM MQ -> Solace.
const validWF = `
source:
  mq:
    conn-name: mqhost(1414)
    queue-manager: QM1
    channel: CH
    user: app
    password: pw
    queue: Q.IN
target:
  solace:
    host: tcp://b:55555
    msg-vpn: prod
    client-username: connector
    client-password: pw
    queue: OUT
`

// invalidWF fails structural validation: conn-name does not match host(port).
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

// bareEnv carries no deploy sections, just the (already-default) workflow
// discovery config, spelled out for clarity.
const bareEnv = `workflows:
  dir: .
  file_pattern: "*"
`

// kubeEnv is a minimal-but-valid kubernetes: section for the deploy/delete seam
// tests -- a safe command, and the three fields checkKube requires.
const kubeEnv = `
kubernetes:
  command: kubectl
  deployment:
    name: solmq-connector
    namespace: solace-connectors
    image: img:1
`

// kubeEnvUnsafeCommand mirrors kubeEnv but with a command containing a shell
// metacharacter -- validate.go's checkKube never inspects kubernetes.command
// (only the docker/podman container-target checks do), so this reaches the CLI
// and is rejected only at the runner.ParseCommand/SafeToken gate.
const kubeEnvUnsafeCommand = `
kubernetes:
  command: "kubectl; rm -rf /"
  deployment:
    name: solmq-connector
    namespace: solace-connectors
    image: img:1
`

// ---- generic fixture helpers ---------------------------------------------------

func write(t *testing.T, dir, name, body string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
}

// workflowDir writes a bare env.yaml plus wf as the sole workflow file (10.yaml)
// into a fresh temp dir (workflows.dir default "." picks it up alongside env.yaml).
func workflowDir(t *testing.T, wf string) string {
	t.Helper()
	dir := t.TempDir()
	write(t, dir, "env.yaml", bareEnv)
	write(t, dir, "10.yaml", wf)
	return dir
}

// manyWorkflowsDir writes a bare env.yaml plus n distinct, valid workflow files
// (literal passwords, so no environment is needed) into a fresh temp dir --
// enough to exercise sharding once n exceeds validate.MaxWorkflowsPerInstance (20).
func manyWorkflowsDir(t *testing.T, n int) string {
	t.Helper()
	dir := t.TempDir()
	write(t, dir, "env.yaml", bareEnv)
	for i := 0; i < n; i++ {
		write(t, dir, fmt.Sprintf("wf-%02d.yaml", i), fmt.Sprintf(`
source:
  solace:
    host: tcp://b:55555
    msg-vpn: prod
    client-username: connector
    client-password: pw
    queue: Q.IN.%d
target:
  mq:
    conn-name: mqhost(1414)
    queue-manager: QM1
    channel: CH
    user: app
    password: pw
    queue: OUT.%d
`, i, i))
	}
	return dir
}

// captureStdout redirects os.Stdout for the duration of f and returns everything
// written. The drain runs concurrently: os.Pipe has a small buffer, so output
// larger than it (e.g. a multi-instance config) would block the writer if we
// read only after f() returns.
func captureStdout(t *testing.T, f func()) string {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w
	done := make(chan string, 1)
	go func() {
		b, _ := io.ReadAll(r)
		done <- string(b)
	}()
	f()
	_ = w.Close()
	os.Stdout = old
	return <-done
}

// ---- deploy/delete seam helpers -------------------------------------------------

// fakeRunner records every invocation reaching the deploy/delete seam so tests
// can assert the exact argv/stdin without starting a process (mirrors
// internal/runner/runner_test.go's fakeRunner).
type fakeRunner struct {
	calls []fakeCall
	err   error
}

type fakeCall struct {
	argv  []string
	stdin string
}

func (f *fakeRunner) Run(argv []string, stdin string) (string, error) {
	f.calls = append(f.calls, fakeCall{argv: argv, stdin: stdin})
	return "out", f.err
}

// useFakeRunner swaps newRunner for a fake that captures argv instead of
// starting a process, restoring the original on cleanup. Tests using this seam
// mutate package-level state, so they must not run in parallel with each other.
func useFakeRunner(t *testing.T) *fakeRunner {
	t.Helper()
	f := &fakeRunner{}
	old := newRunner
	newRunner = func() runner.Runner { return f }
	t.Cleanup(func() { newRunner = old })
	return f
}

// ---- a) exit-code contract -----------------------------------------------------

func TestExitCodeContract(t *testing.T) {
	validDir := workflowDir(t, validWF)
	invalidDir := workflowDir(t, invalidWF)
	cases := []struct {
		name string
		args []string
		want int
	}{
		{"nil args", nil, 2},
		{"unknown command", []string{"bogus"}, 2},
		{"help short", []string{"-h"}, 0},
		{"help long", []string{"--help"}, 0},
		{"help word", []string{"help"}, 0},
		{"unknown flag", []string{"generate", "config", "-nope"}, 2},
		{"missing env file", []string{"generate", "config", "-e", filepath.Join(validDir, "does-not-exist.yaml")}, 1},
		{"invalid spec", []string{"generate", "config", "-e", filepath.Join(invalidDir, "env.yaml")}, 1},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := run(c.args); got != c.want {
				t.Errorf("run(%v) = %d, want %d", c.args, got, c.want)
			}
		})
	}
}

// ---- b) generate config: stdout vs file ----------------------------------------

func TestGenerateConfigStdoutAndFileMatch(t *testing.T) {
	dir := workflowDir(t, validWF)
	envPath := filepath.Join(dir, "env.yaml")

	var code int
	stdout := captureStdout(t, func() { code = run([]string{"generate", "config", "-e", envPath}) })
	if code != 0 {
		t.Fatalf("stdout run exit=%d", code)
	}
	if !strings.Contains(stdout, "spring:") {
		t.Errorf("stdout missing spring:\n%s", stdout)
	}

	// Output outside the env dir: the workflow scanner reads every yaml there
	// (file_pattern "*") and would parse a previous run's output as a workflow.
	out := filepath.Join(t.TempDir(), "out.yml")
	if code := run([]string{"generate", "config", "-e", envPath, "-o", out}); code != 0 {
		t.Fatalf("file run exit=%d", code)
	}
	b, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != stdout {
		t.Errorf("file content differs from stdout:\nfile=%q\nstdout=%q", b, stdout)
	}
}

// ---- c/d) sharded generate ------------------------------------------------------

func TestGenerateConfigShardedFiles(t *testing.T) {
	dir := manyWorkflowsDir(t, 21) // > MaxWorkflowsPerInstance (20) -> 2 shards
	outDir := t.TempDir()          // outside the env dir, which the scanner reads wholesale
	out := filepath.Join(outDir, "out.yml")
	if code := run([]string{"generate", "config", "-e", filepath.Join(dir, "env.yaml"), "-o", out}); code != 0 {
		t.Fatalf("exit=%d", code)
	}
	for _, f := range []string{"out-1.yml", "out-2.yml"} {
		if _, err := os.Stat(filepath.Join(outDir, f)); err != nil {
			t.Errorf("expected %s to be written: %v", f, err)
		}
	}
	// emitConfigs' sharded path never writes the unsuffixed name.
	if _, err := os.Stat(out); !os.IsNotExist(err) {
		t.Errorf("unsuffixed %s should not be written when sharded", out)
	}
}

func TestGenerateConfigShardedStdout(t *testing.T) {
	dir := manyWorkflowsDir(t, 21)
	var code int
	stdout := captureStdout(t, func() {
		code = run([]string{"generate", "config", "-e", filepath.Join(dir, "env.yaml")})
	})
	if code != 0 {
		t.Fatalf("exit=%d", code)
	}
	for _, w := range []string{"CONNECTOR INSTANCE 1 OF 2", "CONNECTOR INSTANCE 2 OF 2"} {
		if !strings.Contains(stdout, w) {
			t.Errorf("stdout banner missing %q", w)
		}
	}
}

// ---- e) flags before/after the positional target --------------------------------

func TestGenerateFlagsBeforeAndAfterPositional(t *testing.T) {
	dir := workflowDir(t, validWF)
	envPath := filepath.Join(dir, "env.yaml")

	// Outputs go to a separate dir: the workflow scanner reads every yaml in
	// the env dir (file_pattern "*"), so an output written there would be
	// parsed as a workflow on the next run.
	outDir := t.TempDir()
	before := filepath.Join(outDir, "before.yml")
	if code := run([]string{"generate", "-e", envPath, "-o", before, "config"}); code != 0 {
		t.Fatalf("flags-before-target exit=%d", code)
	}
	if _, err := os.Stat(before); err != nil {
		t.Errorf("flags-before-target: output not written: %v", err)
	}

	after := filepath.Join(outDir, "after.yml")
	if code := run([]string{"generate", "config", "-e", envPath, "-o", after}); code != 0 {
		t.Fatalf("flags-after-target exit=%d", code)
	}
	if _, err := os.Stat(after); err != nil {
		t.Errorf("flags-after-target (documented usage): output not written: %v", err)
	}
}

// ---- f) suffixBeforeExt ---------------------------------------------------------

func TestSuffixBeforeExt(t *testing.T) {
	cases := []struct {
		in   string
		n    int
		want string
	}{
		{"out.yml", 1, "out-1.yml"},
		{"out", 1, "out-1"},
		{"a.tar.gz", 1, "a.tar-1.gz"},
		// filepath.Ext(".env") returns ".env" (the leading dot IS the extension
		// marker to Ext, dot-files have no "base name"), so the whole name is
		// trimmed and the suffix lands in front of it. Pinned current behavior --
		// do not change suffixBeforeExt to "fix" this.
		{".env", 1, "-1.env"},
	}
	for _, c := range cases {
		if got := suffixBeforeExt(c.in, c.n); got != c.want {
			t.Errorf("suffixBeforeExt(%q, %d) = %q, want %q", c.in, c.n, got, c.want)
		}
	}
}

// ---- g) writeCredEnvFile ---------------------------------------------------------

func TestWriteCredEnvFile(t *testing.T) {
	t.Run("empty name writes nothing", func(t *testing.T) {
		dir := t.TempDir()
		creds := &spec.CredentialsSecret{Create: &spec.CredCreate{Source: spec.SourceEnv, Variables: []string{"X"}}}
		if err := writeCredEnvFile(dir, "", creds, gen.Resolver{}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		entries, err := os.ReadDir(dir)
		if err != nil {
			t.Fatal(err)
		}
		if len(entries) != 0 {
			t.Errorf("expected no file written, got %v", entries)
		}
	})

	t.Run("resolver error propagates and writes nothing", func(t *testing.T) {
		dir := t.TempDir()
		creds := &spec.CredentialsSecret{Create: &spec.CredCreate{Source: spec.SourceEnv, Variables: []string{"MISSING"}}}
		res := gen.Resolver{Env: func(string) (string, bool) { return "", false }}
		if err := writeCredEnvFile(dir, "creds.env", creds, res); err == nil {
			t.Fatal("expected an error for an unset credentials variable")
		}
		if _, err := os.Stat(filepath.Join(dir, "creds.env")); !os.IsNotExist(err) {
			t.Error("no file should be written when the resolver errors")
		}
	})

	t.Run("nil kvs leaves an existing env-file untouched", func(t *testing.T) {
		dir := t.TempDir()
		write(t, dir, "existing.env", "PRESERVE=1\n")
		// Create == nil (Existing set instead) -> gen.ResolveCredentials returns a
		// nil kvs slice, the "existing env-file" case.
		creds := &spec.CredentialsSecret{Existing: "existing.env"}
		if err := writeCredEnvFile(dir, "existing.env", creds, gen.Resolver{}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		b, err := os.ReadFile(filepath.Join(dir, "existing.env"))
		if err != nil {
			t.Fatal(err)
		}
		if string(b) != "PRESERVE=1\n" {
			t.Errorf("existing env-file was modified: %q", b)
		}
	})

	t.Run("happy path writes ordered KVs at mode 0600", func(t *testing.T) {
		dir := t.TempDir()
		creds := &spec.CredentialsSecret{Create: &spec.CredCreate{Source: spec.SourceEnv, Variables: []string{"A", "B"}}}
		vals := map[string]string{"A": "1", "B": "2"}
		res := gen.Resolver{Env: func(k string) (string, bool) { v, ok := vals[k]; return v, ok }}
		if err := writeCredEnvFile(dir, "creds.env", creds, res); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		p := filepath.Join(dir, "creds.env")
		b, err := os.ReadFile(p)
		if err != nil {
			t.Fatal(err)
		}
		if string(b) != "A=1\nB=2\n" {
			t.Errorf("content = %q, want %q", b, "A=1\nB=2\n")
		}
		if runtime.GOOS != "windows" {
			fi, err := os.Stat(p)
			if err != nil {
				t.Fatal(err)
			}
			if fi.Mode().Perm() != 0o600 {
				t.Errorf("mode = %v, want 0600", fi.Mode().Perm())
			}
		}
	})
}

// ---- h) emit write error via run() -----------------------------------------------

func TestGenerateConfigEmitWriteError(t *testing.T) {
	dir := workflowDir(t, validWF)
	bad := filepath.Join(dir, "no-such-subdir", "o.yml") // parent dir does not exist
	if code := run([]string{"generate", "config", "-e", filepath.Join(dir, "env.yaml"), "-o", bad}); code != 1 {
		t.Fatalf("emit to unwritable path should return 1, got %d", code)
	}
}

// ---- i) loadEnv path resolution --------------------------------------------------

func TestLoadEnvWorkflowsDirRelativeToEnvFile(t *testing.T) {
	base := t.TempDir()
	envDir := filepath.Join(base, "config")
	wfDir := filepath.Join(envDir, "workflows")
	if err := os.MkdirAll(wfDir, 0o755); err != nil {
		t.Fatal(err)
	}
	write(t, envDir, "env.yaml", "workflows:\n  dir: workflows\n")
	write(t, wfDir, "10.yaml", validWF)

	// cwd differs from envDir, so a cwd-relative (rather than env-file-relative)
	// resolution of workflows.dir would miss the folder entirely.
	t.Chdir(t.TempDir())

	var code int
	stdout := captureStdout(t, func() {
		code = run([]string{"generate", "config", "-e", filepath.Join(envDir, "env.yaml")})
	})
	if code != 0 {
		t.Fatalf("exit=%d, stdout=%s", code, stdout)
	}
	if !strings.Contains(stdout, "spring:") {
		t.Errorf("expected rendered config, got:\n%s", stdout)
	}
}

func TestLoadEnvExcludesEnvFileFromWorkflowSet(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "env.yaml", bareEnv) // workflows.dir "." + file_pattern "*" also matches env.yaml itself
	write(t, dir, "10.yaml", validWF)
	// If scan.Scan did not exclude the env file, it would be parsed as a malformed
	// workflow (no source:/target:) and validation would fail.
	if code := run([]string{"generate", "config", "-e", filepath.Join(dir, "env.yaml")}); code != 0 {
		t.Fatalf("env.yaml must be excluded from its own workflow scan, exit=%d", code)
	}
}

// ---- j) deploy/delete via the runner seam ----------------------------------------
//
// These tests swap the package-level newRunner var, so none of them call
// t.Parallel() and they must not run concurrently with each other.

func TestDeployKubernetesSeamHappyPath(t *testing.T) {
	f := useFakeRunner(t)
	dir := t.TempDir()
	write(t, dir, "env.yaml", kubeEnv)
	write(t, dir, "10.yaml", validWF)

	if code := run([]string{"deploy", "kubernetes", "-e", filepath.Join(dir, "env.yaml")}); code != 0 {
		t.Fatalf("exit=%d", code)
	}
	if len(f.calls) != 1 {
		t.Fatalf("want 1 call to the runner, got %d", len(f.calls))
	}
	want := []string{"kubectl", "apply", "-f", "-"}
	if !reflect.DeepEqual(f.calls[0].argv, want) {
		t.Errorf("argv = %v, want %v", f.calls[0].argv, want)
	}
	if !strings.Contains(f.calls[0].stdin, "kind: Deployment") {
		t.Errorf("manifest on stdin missing kind: Deployment:\n%s", f.calls[0].stdin)
	}
}

func TestDeleteKubernetesSeamHappyPath(t *testing.T) {
	f := useFakeRunner(t)
	dir := t.TempDir()
	write(t, dir, "env.yaml", kubeEnv)
	write(t, dir, "10.yaml", validWF)

	if code := run([]string{"delete", "kubernetes", "-e", filepath.Join(dir, "env.yaml")}); code != 0 {
		t.Fatalf("exit=%d", code)
	}
	if len(f.calls) != 1 {
		t.Fatalf("want 1 call to the runner, got %d", len(f.calls))
	}
	want := []string{"kubectl", "delete", "-f", "-"}
	if !reflect.DeepEqual(f.calls[0].argv, want) {
		t.Errorf("argv = %v, want %v", f.calls[0].argv, want)
	}
}

// TestDeployKubernetesSeamRejectsUnsafeCommand mirrors the house pattern in
// internal/runner/runner_test.go's TestKubernetesRejectsUnsafeCommand: an unsafe
// command must fail before anything reaches the runner.
func TestDeployKubernetesSeamRejectsUnsafeCommand(t *testing.T) {
	f := useFakeRunner(t)
	dir := t.TempDir()
	write(t, dir, "env.yaml", kubeEnvUnsafeCommand)
	write(t, dir, "10.yaml", validWF)

	if code := run([]string{"deploy", "kubernetes", "-e", filepath.Join(dir, "env.yaml")}); code != 1 {
		t.Fatalf("unsafe kubernetes.command should exit 1, got %d", code)
	}
	if len(f.calls) != 0 {
		t.Fatalf("nothing must reach the runner when the command is rejected, got %d calls", len(f.calls))
	}
}

// ---- k) validate subcommand -------------------------------------------------------

func TestValidateOKAndErrors(t *testing.T) {
	validDir := workflowDir(t, validWF)
	if code := run([]string{"validate", "-e", filepath.Join(validDir, "env.yaml")}); code != 0 {
		t.Errorf("valid spec should validate, exit=%d", code)
	}
	invalidDir := workflowDir(t, invalidWF)
	if code := run([]string{"validate", "-e", filepath.Join(invalidDir, "env.yaml")}); code != 1 {
		t.Errorf("invalid spec should fail validate, exit=%d", code)
	}
}

// ---- l) examples CLI contract ------------------------------------------------------

func TestExamplesWriteSkipForceThenGenerate(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "ex")
	if code := run([]string{"examples", dir}); code != 0 {
		t.Fatalf("first exit=%d", code)
	}
	f := filepath.Join(dir, "env.yaml")
	if err := os.WriteFile(f, []byte("touched\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if code := run([]string{"examples", dir}); code != 0 { // no -f: must skip existing
		t.Fatalf("re-run exit=%d", code)
	}
	if b, _ := os.ReadFile(f); string(b) != "touched\n" {
		t.Error("existing file should be skipped without -f")
	}
	if code := run([]string{"examples", "-f", dir}); code != 0 { // -f before dir
		t.Fatalf("force exit=%d", code)
	}
	if b, _ := os.ReadFile(f); string(b) == "touched\n" {
		t.Error("-f should overwrite the existing file")
	}
	// The shipped examples must generate cleanly through the CLI; the
	// package-level equivalent lives in
	// internal/examples/examples_test.go's TestShippedExamplesGenerateConfig.
	if code := run([]string{"generate", "config", "-e", f}); code != 0 {
		t.Fatalf("generate on shipped examples exit=%d", code)
	}
}

func TestExamplesDefaultDir(t *testing.T) {
	tmp := t.TempDir()
	t.Chdir(tmp)
	if code := run([]string{"examples"}); code != 0 {
		t.Fatalf("default-dir exit=%d", code)
	}
	if _, err := os.Stat(filepath.Join(tmp, "examples", "workflow-0.yaml")); err != nil {
		t.Errorf("default dir 'examples' not created: %v", err)
	}
}

// ---- generate kubernetes / docker / podman ---------------------------------------

// dockerEnv is a minimal-but-valid docker: section (checkContainerTarget needs a
// DNS-1123 name, a safe command, a non-empty image, and in-range ports).
const dockerEnv = `
docker:
  command: docker
  image: img:1
  name: solmq-conn
  ports:
    - 8090
`

// podmanEnv renders a minimal-but-valid podman: section in quadlet mode with the
// unit dir overridden to a test-owned location (the default user-scope dir lives
// under the real home directory). ToSlash keeps the Windows temp path valid YAML.
func podmanEnv(quadletDir string) string {
	return fmt.Sprintf(`
podman:
  command: podman
  image: img:1
  name: solmq-conn
  ports:
    - 8090
  mode: quadlet
  quadlet:
    scope: user
    dir: %s
`, filepath.ToSlash(quadletDir))
}

func TestGenerateKubernetesStdout(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "env.yaml", kubeEnv)
	write(t, dir, "10.yaml", validWF)
	var code int
	stdout := captureStdout(t, func() {
		code = run([]string{"generate", "kubernetes", "-e", filepath.Join(dir, "env.yaml")})
	})
	if code != 0 {
		t.Fatalf("exit=%d", code)
	}
	if !strings.Contains(stdout, "kind: Deployment") {
		t.Errorf("manifest missing kind: Deployment:\n%s", stdout)
	}
}

func TestGenerateDockerToFile(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "env.yaml", dockerEnv)
	write(t, dir, "10.yaml", validWF)
	out := filepath.Join(t.TempDir(), "compose.yml")
	if code := run([]string{"generate", "docker", "-e", filepath.Join(dir, "env.yaml"), "-o", out}); code != 0 {
		t.Fatalf("exit=%d", code)
	}
	b, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	for _, w := range []string{"services:", "image: img:1"} {
		if !strings.Contains(string(b), w) {
			t.Errorf("compose missing %q:\n%s", w, b)
		}
	}
}

// TestGeneratePodmanQuadletStdout drives genPodman's quadlet arm, which previews
// every unit through joinUnits' filename banner.
func TestGeneratePodmanQuadletStdout(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "env.yaml", podmanEnv(t.TempDir()))
	write(t, dir, "10.yaml", validWF)
	var code int
	stdout := captureStdout(t, func() {
		code = run([]string{"generate", "podman", "-e", filepath.Join(dir, "env.yaml")})
	})
	if code != 0 {
		t.Fatalf("exit=%d", code)
	}
	if !strings.Contains(stdout, "# === solmq-conn.container ===") {
		t.Errorf("quadlet preview missing unit banner:\n%s", stdout)
	}
}

// TestJoinUnits pins the preview format directly: banner per unit, blank-line
// separator only between units.
func TestJoinUnits(t *testing.T) {
	units := []podmangen.Unit{
		{Filename: "a.container", Content: "A\n"},
		{Filename: "b.container", Content: "B\n"},
	}
	want := "# === a.container ===\nA\n\n# === b.container ===\nB\n"
	if got := joinUnits(units); got != want {
		t.Errorf("joinUnits = %q, want %q", got, want)
	}
	if got := joinUnits(units[:1]); got != "# === a.container ===\nA\n" {
		t.Errorf("single unit must have no separator, got %q", got)
	}
}

// ---- deploy / delete: docker and podman seams -------------------------------------

func TestDeployDockerSeamWritesComposeAndRuns(t *testing.T) {
	f := useFakeRunner(t)
	dir := t.TempDir()
	write(t, dir, "env.yaml", dockerEnv)
	write(t, dir, "10.yaml", validWF)

	if code := run([]string{"deploy", "docker", "-e", filepath.Join(dir, "env.yaml")}); code != 0 {
		t.Fatalf("exit=%d", code)
	}
	compose := filepath.Join(dir, "docker-compose.yml")
	if _, err := os.Stat(compose); err != nil {
		t.Fatalf("compose file not written: %v", err)
	}
	if len(f.calls) != 1 {
		t.Fatalf("want 1 runner call, got %d", len(f.calls))
	}
	want := []string{"docker", "compose", "-f", compose, "up", "-d"}
	if !reflect.DeepEqual(f.calls[0].argv, want) {
		t.Errorf("argv = %v, want %v", f.calls[0].argv, want)
	}
}

func TestDeleteDockerSeam(t *testing.T) {
	f := useFakeRunner(t)
	dir := t.TempDir()
	write(t, dir, "env.yaml", dockerEnv)
	write(t, dir, "10.yaml", validWF)

	if code := run([]string{"delete", "docker", "-e", filepath.Join(dir, "env.yaml")}); code != 0 {
		t.Fatalf("exit=%d", code)
	}
	want := []string{"docker", "compose", "-f", filepath.Join(dir, "docker-compose.yml"), "down"}
	if len(f.calls) != 1 || !reflect.DeepEqual(f.calls[0].argv, want) {
		t.Errorf("calls = %+v, want single %v", f.calls, want)
	}
}

func TestDeployPodmanSeamWritesUnitsAndStarts(t *testing.T) {
	f := useFakeRunner(t)
	quadletDir := t.TempDir()
	dir := t.TempDir()
	write(t, dir, "env.yaml", podmanEnv(quadletDir))
	write(t, dir, "10.yaml", validWF)

	if code := run([]string{"deploy", "podman", "-e", filepath.Join(dir, "env.yaml")}); code != 0 {
		t.Fatalf("exit=%d", code)
	}
	// The app yaml and the quadlet unit land in the (overridden) quadlet dir;
	// no credential env-file is written when there is no secrets section.
	for _, name := range []string{"solmq-conn-application.yml", "solmq-conn.container"} {
		if _, err := os.Stat(filepath.Join(quadletDir, name)); err != nil {
			t.Errorf("%s not written to quadlet dir: %v", name, err)
		}
	}
	wantCalls := [][]string{
		{"systemctl", "--user", "daemon-reload"},
		{"systemctl", "--user", "start", "solmq-conn.service"},
	}
	if len(f.calls) != len(wantCalls) {
		t.Fatalf("want %d runner calls, got %+v", len(wantCalls), f.calls)
	}
	for i, w := range wantCalls {
		if !reflect.DeepEqual(f.calls[i].argv, w) {
			t.Errorf("call %d argv = %v, want %v", i, f.calls[i].argv, w)
		}
	}
}

func TestDeletePodmanSeamStopsRemovesReloads(t *testing.T) {
	f := useFakeRunner(t)
	quadletDir := t.TempDir()
	dir := t.TempDir()
	write(t, dir, "env.yaml", podmanEnv(quadletDir))
	write(t, dir, "10.yaml", validWF)
	// Pre-seed the files a deploy would have written; delete must remove both
	// the unit and the generated app yaml.
	write(t, quadletDir, "solmq-conn.container", "[Container]\n")
	write(t, quadletDir, "solmq-conn-application.yml", "x\n")

	if code := run([]string{"delete", "podman", "-e", filepath.Join(dir, "env.yaml")}); code != 0 {
		t.Fatalf("exit=%d", code)
	}
	wantCalls := [][]string{
		{"systemctl", "--user", "stop", "solmq-conn.service"},
		{"systemctl", "--user", "daemon-reload"},
	}
	if len(f.calls) != len(wantCalls) {
		t.Fatalf("want %d runner calls, got %+v", len(wantCalls), f.calls)
	}
	for i, w := range wantCalls {
		if !reflect.DeepEqual(f.calls[i].argv, w) {
			t.Errorf("call %d argv = %v, want %v", i, f.calls[i].argv, w)
		}
	}
	for _, name := range []string{"solmq-conn.container", "solmq-conn-application.yml"} {
		if _, err := os.Stat(filepath.Join(quadletDir, name)); !os.IsNotExist(err) {
			t.Errorf("%s should be removed by delete", name)
		}
	}
}

// ---- absPath ----------------------------------------------------------------------

func TestAbsPath(t *testing.T) {
	// t.TempDir is absolute on every platform; a rooted path without a volume
	// (a bare backslash prefix) is NOT absolute on Windows, so it cannot serve here.
	abs := t.TempDir()
	if got := absPath("base", abs); got != abs {
		t.Errorf("absolute input must pass through: got %q", got)
	}
	if got := absPath("base", "rel.yaml"); got != filepath.Join("base", "rel.yaml") {
		t.Errorf("relative input must join onto dir: got %q", got)
	}
}
