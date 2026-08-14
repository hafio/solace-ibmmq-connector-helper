package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/runner"
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

// envCredWF mixes a literal MQ credential with a -env Solace credential, so the
// docker child-env test exercises both resolution paths through the real
// resolver (os.LookupEnv), not just literal passthrough.
const envCredWF = `
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
    client-username-env: SOLACE_USER
    client-password-env: SOLACE_PASS
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
// used to push the workflow count past validate.MaxWorkflows (20), which is
// now a fatal cap error rather than a sharding trigger.
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
// larger than it (e.g. a config with many workflows) would block the writer if
// we read only after f() returns.
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

// captureStderr redirects os.Stderr for the duration of f and returns everything
// written (see captureStdout for why the drain runs concurrently).
func captureStderr(t *testing.T, f func()) string {
	t.Helper()
	old := os.Stderr
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stderr = w
	done := make(chan string, 1)
	go func() {
		b, _ := io.ReadAll(r)
		done <- string(b)
	}()
	f()
	_ = w.Close()
	os.Stderr = old
	return <-done
}

// ---- deploy/delete seam helpers -------------------------------------------------

// fakeRunner records every invocation reaching the deploy/delete seam so tests
// can assert the exact argv/stdin/env without starting a process (mirrors
// internal/runner/runner_test.go's fakeRunner).
//
// err, when set, fails every call from index failFrom onward (0 by default, so
// it fails from the very first call -- the preflight probe, now that one always
// precedes the mutating command). A test that needs the probe to succeed but
// the real deploy/delete call to fail sets failFrom to the index of that call.
type fakeRunner struct {
	calls    []fakeCall
	err      error
	failFrom int
}

type fakeCall struct {
	argv  []string
	stdin string
	env   []string
}

func (f *fakeRunner) Run(c runner.Cmd) (string, error) {
	idx := len(f.calls)
	f.calls = append(f.calls, fakeCall{argv: c.Argv, stdin: c.Stdin, env: c.Env})
	if f.err != nil && idx >= f.failFrom {
		return "out", f.err
	}
	return "out", nil
}

// useFakeRunner returns a fake that captures argv instead of starting a
// process. Callers pass it to dispatch (not run, which always uses the
// production runner.OS{}) so the fake reaches the deploy/delete seam as an
// explicit argument rather than through mutable package state.
func useFakeRunner(t *testing.T) *fakeRunner {
	t.Helper()
	return &fakeRunner{}
}

// ---- a) exit-code contract -----------------------------------------------------

func TestExitCodeContract(t *testing.T) {
	validDir := workflowDir(t, validWF)
	invalidDir := workflowDir(t, invalidWF)
	// validDir's env.yaml (bareEnv) has no kubernetes:/docker:/podman: section, so
	// it doubles as the fixture for the "missing section" deploy/delete cases below.
	noSectionEnv := filepath.Join(validDir, "env.yaml")
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
		{"generate no target", []string{"generate"}, 2},
		{"generate bogus target", []string{"generate", "bogus"}, 2},
		{"deploy no platform", []string{"deploy"}, 2},
		{"deploy bogus platform", []string{"deploy", "bogus"}, 2},
		{"deploy kubernetes no section", []string{"deploy", "kubernetes", "-e", noSectionEnv}, 1},
		{"deploy docker no section", []string{"deploy", "docker", "-e", noSectionEnv}, 1},
		{"deploy podman no section", []string{"deploy", "podman", "-e", noSectionEnv}, 1},
		{"delete kubernetes no section", []string{"delete", "kubernetes", "-e", noSectionEnv}, 1},
		{"delete docker no section", []string{"delete", "docker", "-e", noSectionEnv}, 1},
		{"delete podman no section", []string{"delete", "podman", "-e", noSectionEnv}, 1},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			f := &fakeRunner{}
			if got := dispatch(c.args, f); got != c.want {
				t.Errorf("dispatch(%v) = %d, want %d", c.args, got, c.want)
			}
			// Every case here is a rejection (bad usage or a missing deploy section):
			// none of them should ever reach the runner.
			if len(f.calls) != 0 {
				t.Errorf("dispatch(%v): runner must not be invoked on a rejection, got %d calls", c.args, len(f.calls))
			}
		})
	}
}

// ---- a2) command-model / dispatch drift guard ------------------------------------

// TestDispatchHandlersMatchModel guards against the class of drift task 2 fixed
// by hand (examples' default dir: the model and docs agreed while the code
// disagreed with both): the handler maps dispatch/runGenerate/runAction actually
// use must name exactly the verbs/targets cliVerbs (commands.go) documents, in
// both directions -- a modeled entry with no handler, or a handler with no
// modeled entry, must fail.
func TestDispatchHandlersMatchModel(t *testing.T) {
	verbNames := make(map[string]bool, len(cliVerbs))
	for _, v := range cliVerbs {
		verbNames[v.Name] = true
	}
	assertSameNameSet(t, "verb", verbNames, keySet(verbHandlers))
	assertSameNameSet(t, "generate target", nameSet(targetNames("generate")), keySet(genTargets))
	assertSameNameSet(t, "deploy platform", nameSet(targetNames("deploy")), keySet(actTargets))
	assertSameNameSet(t, "delete platform", nameSet(targetNames("delete")), keySet(actTargets))
}

func nameSet(names []string) map[string]bool {
	m := make(map[string]bool, len(names))
	for _, n := range names {
		m[n] = true
	}
	return m
}

func keySet[V any](m map[string]V) map[string]bool {
	out := make(map[string]bool, len(m))
	for k := range m {
		out[k] = true
	}
	return out
}

// assertSameNameSet fails if either set contains a name the other lacks.
func assertSameNameSet(t *testing.T, label string, modeled, handlers map[string]bool) {
	t.Helper()
	for name := range modeled {
		if !handlers[name] {
			t.Errorf("%s %q is modeled (commands.go) but has no handler", label, name)
		}
	}
	for name := range handlers {
		if !modeled[name] {
			t.Errorf("%s handler %q has no modeled (commands.go) entry", label, name)
		}
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

// ---- c) workflow count cap -------------------------------------------------------

// TestGenerateConfigWorkflowCapExceeded pins the new hard-validation behavior:
// a folder with more than validate.MaxWorkflows (20) workflows is a fatal error
// (checkWorkflowCount), not something the tool splits across instances, and
// generate config must write no output file at all in that case.
func TestGenerateConfigWorkflowCapExceeded(t *testing.T) {
	dir := manyWorkflowsDir(t, 21) // > MaxWorkflows (20)
	outDir := t.TempDir()          // outside the env dir, which the scanner reads wholesale
	out := filepath.Join(outDir, "out.yml")

	var code int
	stderr := captureStderr(t, func() {
		code = run([]string{"generate", "config", "-e", filepath.Join(dir, "env.yaml"), "-o", out})
	})
	if code != 1 {
		t.Fatalf("exit=%d, want 1", code)
	}
	if !strings.Contains(stderr, "21 workflows found") || !strings.Contains(stderr, "at most 20") {
		t.Errorf("stderr missing cap error, got:\n%s", stderr)
	}
	if _, err := os.Stat(out); !os.IsNotExist(err) {
		t.Errorf("no output file should be written when the workflow cap is exceeded, stat err=%v", err)
	}
}

// ---- d) flags before/after the positional target --------------------------------

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

// ---- e) emit write error via run() -----------------------------------------------

func TestGenerateConfigEmitWriteError(t *testing.T) {
	dir := workflowDir(t, validWF)
	bad := filepath.Join(dir, "no-such-subdir", "o.yml") // parent dir does not exist
	if code := run([]string{"generate", "config", "-e", filepath.Join(dir, "env.yaml"), "-o", bad}); code != 1 {
		t.Fatalf("emit to unwritable path should return 1, got %d", code)
	}
}

// ---- f) loadEnv path resolution --------------------------------------------------

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

// ---- g) deploy/delete via the runner seam ----------------------------------------
//
// These tests call dispatch directly (not run) so they can pass a fakeRunner
// explicitly instead of mutating package-level state.

func TestDeployKubernetesSeamHappyPath(t *testing.T) {
	f := useFakeRunner(t)
	dir := t.TempDir()
	write(t, dir, "env.yaml", kubeEnv)
	write(t, dir, "10.yaml", validWF)

	if code := dispatch([]string{"deploy", "kubernetes", "-e", filepath.Join(dir, "env.yaml")}, f); code != 0 {
		t.Fatalf("exit=%d", code)
	}
	// Call 0 is the read-only preflight probe (runner.Preflight), which always
	// precedes the mutating apply/delete call.
	if len(f.calls) != 2 {
		t.Fatalf("want 2 calls to the runner (preflight, apply), got %d: %+v", len(f.calls), f.calls)
	}
	wantPreflight := []string{"kubectl", "auth", "can-i", "create", "deployment", "--namespace", "solace-connectors"}
	if !reflect.DeepEqual(f.calls[0].argv, wantPreflight) {
		t.Errorf("preflight argv = %v, want %v", f.calls[0].argv, wantPreflight)
	}
	want := []string{"kubectl", "apply", "-f", "-"}
	if !reflect.DeepEqual(f.calls[1].argv, want) {
		t.Errorf("argv = %v, want %v", f.calls[1].argv, want)
	}
	if !strings.Contains(f.calls[1].stdin, "kind: Deployment") {
		t.Errorf("manifest on stdin missing kind: Deployment:\n%s", f.calls[1].stdin)
	}
}

func TestDeleteKubernetesSeamHappyPath(t *testing.T) {
	f := useFakeRunner(t)
	dir := t.TempDir()
	write(t, dir, "env.yaml", kubeEnv)
	write(t, dir, "10.yaml", validWF)

	if code := dispatch([]string{"delete", "kubernetes", "-e", filepath.Join(dir, "env.yaml")}, f); code != 0 {
		t.Fatalf("exit=%d", code)
	}
	if len(f.calls) != 2 {
		t.Fatalf("want 2 calls to the runner (preflight, delete), got %d: %+v", len(f.calls), f.calls)
	}
	wantPreflight := []string{"kubectl", "auth", "can-i", "delete", "deployment", "--namespace", "solace-connectors"}
	if !reflect.DeepEqual(f.calls[0].argv, wantPreflight) {
		t.Errorf("preflight argv = %v, want %v", f.calls[0].argv, wantPreflight)
	}
	want := []string{"kubectl", "delete", "-f", "-"}
	if !reflect.DeepEqual(f.calls[1].argv, want) {
		t.Errorf("argv = %v, want %v", f.calls[1].argv, want)
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

	if code := dispatch([]string{"deploy", "kubernetes", "-e", filepath.Join(dir, "env.yaml")}, f); code != 1 {
		t.Fatalf("unsafe kubernetes.command should exit 1, got %d", code)
	}
	if len(f.calls) != 0 {
		t.Fatalf("nothing must reach the runner when the command is rejected, got %d calls", len(f.calls))
	}
}

// ---- h) validate subcommand -------------------------------------------------------

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

// ---- i) examples CLI contract ------------------------------------------------------

func TestExamplesWriteSkipForceThenGenerate(t *testing.T) {
	// The shipped example's connections/workflows/tls/security all reference
	// their credentials via -env (never a literal), so generating it for real
	// needs every one of those host variables set.
	for _, v := range []string{
		"SOL_PASSWORD", "MQ_ARCHIVE_PASSWORD", "MQ_CORE_PASSWORD", "EDGE_SOL_PASSWORD",
		"TRUSTSTORE_PASSWORD", "KEYSTORE_PASSWORD", "HEALTHCHECK_PASSWORD",
	} {
		t.Setenv(v, "x")
	}
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
	if _, err := os.Stat(filepath.Join(tmp, "workflow-0.yaml")); err != nil {
		t.Errorf("default dir '.' (current directory) not written to: %v", err)
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

// podmanEnvSudo mirrors podmanEnv but with a chained `sudo podman` command:
// rejected outright (sudo is not on the podman allowlist) unless the caller
// approves it via --allow-command sudo, which is the escape-hatch scenario
// the flag exists for.
func podmanEnvSudo(quadletDir string) string {
	return fmt.Sprintf(`
podman:
  command: sudo podman
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

// TestGeneratePodmanQuadletStdout drives genPodman's quadlet arm, which now
// prints the single rendered unit's content directly (no per-unit banner --
// there is only ever one instance).
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
	for _, w := range []string{"[Container]", "ContainerName=solmq-conn"} {
		if !strings.Contains(stdout, w) {
			t.Errorf("quadlet preview missing %q:\n%s", w, stdout)
		}
	}
}

// ---- deploy / delete: docker and podman seams -------------------------------------

func TestDeployDockerSeamWritesComposeAndRuns(t *testing.T) {
	f := useFakeRunner(t)
	dir := t.TempDir()
	write(t, dir, "env.yaml", dockerEnv)
	write(t, dir, "10.yaml", validWF)

	if code := dispatch([]string{"deploy", "docker", "-e", filepath.Join(dir, "env.yaml")}, f); code != 0 {
		t.Fatalf("exit=%d", code)
	}
	compose := filepath.Join(dir, "docker-compose.yml")
	// The compose file is scratch, regenerated by every deploy/delete, and is
	// removed once a deploy succeeds.
	if _, err := os.Stat(compose); !os.IsNotExist(err) {
		t.Fatalf("compose file should be removed after a successful deploy, stat err=%v", err)
	}
	// Call 0 is the read-only preflight probe (runner.Preflight), which runs
	// before the compose file is even written.
	if len(f.calls) != 2 {
		t.Fatalf("want 2 runner calls (preflight, up), got %d: %+v", len(f.calls), f.calls)
	}
	if want := []string{"docker", "info"}; !reflect.DeepEqual(f.calls[0].argv, want) {
		t.Errorf("preflight argv = %v, want %v", f.calls[0].argv, want)
	}
	want := []string{"docker", "compose", "-f", compose, "up", "-d"}
	if !reflect.DeepEqual(f.calls[1].argv, want) {
		t.Errorf("argv = %v, want %v", f.calls[1].argv, want)
	}
}

// TestDeployDockerSeamComposeFileSurvivesFailedRun covers the other half of the
// compose file's lifecycle: a half-started stack (docker compose up failed
// partway through) still needs the compose file on disk to `down` with, so a
// failed run must leave it in place instead of cleaning it up. failFrom: 1
// lets the preflight probe (call 0) succeed so the compose file is actually
// written, and fails only the real `up` call (call 1).
func TestDeployDockerSeamComposeFileSurvivesFailedRun(t *testing.T) {
	f := useFakeRunner(t)
	f.err = fmt.Errorf("boom")
	f.failFrom = 1
	dir := t.TempDir()
	write(t, dir, "env.yaml", dockerEnv)
	write(t, dir, "10.yaml", validWF)

	if code := dispatch([]string{"deploy", "docker", "-e", filepath.Join(dir, "env.yaml")}, f); code != 1 {
		t.Fatalf("exit=%d, want 1 for a failed deploy", code)
	}
	compose := filepath.Join(dir, "docker-compose.yml")
	if _, err := os.Stat(compose); err != nil {
		t.Fatalf("compose file must survive a failed run: %v", err)
	}
}

// TestDeployDockerSeamChildEnvCarriesCredentials asserts the credential values
// reach the docker child process's environment as STABLE=value pairs (never in
// argv or on disk), covering both a literal credential (mq) and a -env one
// (solace) resolved from the real process environment.
func TestDeployDockerSeamChildEnvCarriesCredentials(t *testing.T) {
	t.Setenv("SOLACE_USER", "envuser")
	t.Setenv("SOLACE_PASS", "envpass")
	f := useFakeRunner(t)
	dir := t.TempDir()
	write(t, dir, "env.yaml", dockerEnv)
	write(t, dir, "10.yaml", envCredWF)

	if code := dispatch([]string{"deploy", "docker", "-e", filepath.Join(dir, "env.yaml")}, f); code != 0 {
		t.Fatalf("exit=%d", code)
	}
	// Call 0 is the preflight probe (no Env); call 1 is the real `up`, which
	// carries the resolved credentials.
	if len(f.calls) != 2 {
		t.Fatalf("want 2 runner calls, got %d: %+v", len(f.calls), f.calls)
	}
	want := []string{
		"MQ_CONN_1_USER=app",
		"MQ_CONN_1_PASSWORD=pw",
		"SOL_CONN_1_CLIENT_USERNAME=envuser",
		"SOL_CONN_1_CLIENT_PASSWORD=envpass",
	}
	if !reflect.DeepEqual(f.calls[1].env, want) {
		t.Errorf("child env = %v, want %v", f.calls[1].env, want)
	}
}

func TestDeleteDockerSeam(t *testing.T) {
	f := useFakeRunner(t)
	dir := t.TempDir()
	write(t, dir, "env.yaml", dockerEnv)
	write(t, dir, "10.yaml", validWF)

	if code := dispatch([]string{"delete", "docker", "-e", filepath.Join(dir, "env.yaml")}, f); code != 0 {
		t.Fatalf("exit=%d", code)
	}
	wantCalls := [][]string{
		{"docker", "info"},
		{"docker", "compose", "-f", filepath.Join(dir, "docker-compose.yml"), "down"},
	}
	if len(f.calls) != len(wantCalls) {
		t.Fatalf("calls = %+v, want %v", f.calls, wantCalls)
	}
	for i, w := range wantCalls {
		if !reflect.DeepEqual(f.calls[i].argv, w) {
			t.Errorf("call %d argv = %v, want %v", i, f.calls[i].argv, w)
		}
	}
}

func TestDeployPodmanSeamWritesUnitsAndStarts(t *testing.T) {
	f := useFakeRunner(t)
	quadletDir := t.TempDir()
	dir := t.TempDir()
	write(t, dir, "env.yaml", podmanEnv(quadletDir))
	write(t, dir, "10.yaml", validWF)

	if code := dispatch([]string{"deploy", "podman", "-e", filepath.Join(dir, "env.yaml")}, f); code != 0 {
		t.Fatalf("exit=%d", code)
	}
	// The app yaml and the quadlet unit land in the (overridden) quadlet dir.
	for _, name := range []string{"solmq-conn-application.yml", "solmq-conn.container"} {
		if _, err := os.Stat(filepath.Join(quadletDir, name)); err != nil {
			t.Errorf("%s not written to quadlet dir: %v", name, err)
		}
	}
	// validWF's four credentials (mq user/password, solace
	// client-username/client-password) must each be stored (secret rm --ignore,
	// then secret create) BEFORE daemon-reload/start: the unit being started
	// references these secrets by name, so they must already exist. Call 0 is
	// the read-only preflight probe, which runs before any of that.
	wantCalls := [][]string{
		{"podman", "info"},
		{"podman", "secret", "rm", "--ignore", "solmq-conn-MQ_CONN_1_USER"},
		{"podman", "secret", "create", "solmq-conn-MQ_CONN_1_USER", "-"},
		{"podman", "secret", "rm", "--ignore", "solmq-conn-MQ_CONN_1_PASSWORD"},
		{"podman", "secret", "create", "solmq-conn-MQ_CONN_1_PASSWORD", "-"},
		{"podman", "secret", "rm", "--ignore", "solmq-conn-SOL_CONN_1_CLIENT_USERNAME"},
		{"podman", "secret", "create", "solmq-conn-SOL_CONN_1_CLIENT_USERNAME", "-"},
		{"podman", "secret", "rm", "--ignore", "solmq-conn-SOL_CONN_1_CLIENT_PASSWORD"},
		{"podman", "secret", "create", "solmq-conn-SOL_CONN_1_CLIENT_PASSWORD", "-"},
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

	if code := dispatch([]string{"delete", "podman", "-e", filepath.Join(dir, "env.yaml")}, f); code != 0 {
		t.Fatalf("exit=%d", code)
	}
	// Secrets are removed from podman's store only AFTER the unit is stopped and
	// the generator reloaded, mirroring deploy's create-before-start ordering in
	// reverse: a failure removing them still surfaces (see main.go's
	// podmanDelete), but the units referencing them are gone first. Call 0 is
	// the read-only preflight probe, which runs before any of that.
	wantCalls := [][]string{
		{"podman", "info"},
		{"systemctl", "--user", "stop", "solmq-conn.service"},
		{"systemctl", "--user", "daemon-reload"},
		{"podman", "secret", "rm", "--ignore",
			"solmq-conn-MQ_CONN_1_USER", "solmq-conn-MQ_CONN_1_PASSWORD",
			"solmq-conn-SOL_CONN_1_CLIENT_USERNAME", "solmq-conn-SOL_CONN_1_CLIENT_PASSWORD"},
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

// ---- j) --allow-command flag --------------------------------------------------

// TestAllowCommandFlagBadValueExitsUsageError covers the flag's own input
// validation (SafeToken-clean, no path separator) independent of
// CheckDeployCommand: a bad value must fail fs.Parse (exit 2, the usage-error
// contract) before anything reaches loadEnv/validate/the runner.
func TestAllowCommandFlagBadValueExitsUsageError(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "env.yaml", kubeEnv)
	write(t, dir, "10.yaml", validWF)
	envPath := filepath.Join(dir, "env.yaml")

	cases := []struct {
		name string
		val  string
	}{
		{"path", "/usr/bin/sudo"},
		{"unsafe character", "sudo;rm"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			f := &fakeRunner{}
			code := dispatch([]string{"deploy", "kubernetes", "-e", envPath, "--allow-command", c.val}, f)
			if code != 2 {
				t.Errorf("--allow-command %q: exit=%d, want 2", c.val, code)
			}
			if len(f.calls) != 0 {
				t.Errorf("--allow-command %q: runner must not be invoked, got %d calls", c.val, len(f.calls))
			}
		})
	}
}

// TestAllowCommandFlagRejectedOnGenerateAndValidate pins --allow-command as
// deploy/delete-only: generate and validate never register the flag, so
// passing it is an unknown-flag usage error, same as any other undefined flag.
func TestAllowCommandFlagRejectedOnGenerateAndValidate(t *testing.T) {
	dir := workflowDir(t, validWF)
	envPath := filepath.Join(dir, "env.yaml")
	cases := [][]string{
		{"generate", "config", "-e", envPath, "--allow-command", "sudo"},
		{"validate", "-e", envPath, "--allow-command", "sudo"},
	}
	for _, args := range cases {
		if code := run(args); code != 2 {
			t.Errorf("%v: exit=%d, want 2 (flag not defined on this verb)", args, code)
		}
	}
}

// TestAllowCommandFlagRepeatableThreadsToRunner is the threading assertion:
// a chained `sudo podman` command: fails before the runner without the flag,
// and reaches every runner call (preflight, secrets) once --allow-command sudo
// approves it. Passing the flag twice with the same value also covers that it
// is genuinely repeatable (flag.Value.Set called more than once).
func TestAllowCommandFlagRepeatableThreadsToRunner(t *testing.T) {
	quadletDir := t.TempDir()
	dir := t.TempDir()
	write(t, dir, "env.yaml", podmanEnvSudo(quadletDir))
	write(t, dir, "10.yaml", validWF)
	envPath := filepath.Join(dir, "env.yaml")

	fReject := useFakeRunner(t)
	if code := dispatch([]string{"deploy", "podman", "-e", envPath}, fReject); code != 1 {
		t.Fatalf("without --allow-command: exit=%d, want 1", code)
	}
	if len(fReject.calls) != 0 {
		t.Fatalf("without --allow-command: runner must not be invoked, got %d calls", len(fReject.calls))
	}

	fAccept := useFakeRunner(t)
	args := []string{"deploy", "podman", "-e", envPath, "--allow-command", "sudo", "--allow-command", "sudo"}
	if code := dispatch(args, fAccept); code != 0 {
		t.Fatalf("with --allow-command sudo: exit=%d, want 0", code)
	}
	if len(fAccept.calls) == 0 {
		t.Fatalf("with --allow-command sudo: expected the runner to be reached")
	}
	// The preflight probe runs first, resolving argv[0] "sudo" via the flag and
	// "podman" as its own allowlisted chained token.
	if want := []string{"sudo", "podman", "info"}; !reflect.DeepEqual(fAccept.calls[0].argv, want) {
		t.Errorf("preflight argv = %v, want %v", fAccept.calls[0].argv, want)
	}
	foundSecretCall := false
	for _, c := range fAccept.calls {
		if len(c.argv) >= 3 && c.argv[0] == "sudo" && c.argv[1] == "podman" && c.argv[2] == "secret" {
			foundSecretCall = true
			break
		}
	}
	if !foundSecretCall {
		t.Errorf("expected a podman secret call prefixed with [sudo podman secret ...], got %+v", fAccept.calls)
	}
}

// ---- k) preflight ordering -----------------------------------------------------
//
// These pin runner.Preflight's placement: after generate/validate, before any
// file write or mutating exec. A failing probe must stop the action with
// nothing written to disk and nothing beyond the probe itself reaching the
// runner.

func TestDeployKubernetesPreflightFailureStopsBeforeApply(t *testing.T) {
	f := useFakeRunner(t)
	f.err = fmt.Errorf("not logged in")
	dir := t.TempDir()
	write(t, dir, "env.yaml", kubeEnv)
	write(t, dir, "10.yaml", validWF)

	if code := dispatch([]string{"deploy", "kubernetes", "-e", filepath.Join(dir, "env.yaml")}, f); code != 1 {
		t.Fatalf("exit=%d, want 1", code)
	}
	if len(f.calls) != 1 {
		t.Fatalf("want exactly 1 runner call (the preflight probe), got %d: %+v", len(f.calls), f.calls)
	}
	want := []string{"kubectl", "auth", "can-i", "create", "deployment", "--namespace", "solace-connectors"}
	if !reflect.DeepEqual(f.calls[0].argv, want) {
		t.Errorf("preflight argv = %v, want %v", f.calls[0].argv, want)
	}
}

func TestDeployDockerPreflightFailureStopsBeforeWrite(t *testing.T) {
	f := useFakeRunner(t)
	f.err = fmt.Errorf("daemon unreachable")
	dir := t.TempDir()
	write(t, dir, "env.yaml", dockerEnv)
	write(t, dir, "10.yaml", validWF)

	if code := dispatch([]string{"deploy", "docker", "-e", filepath.Join(dir, "env.yaml")}, f); code != 1 {
		t.Fatalf("exit=%d, want 1", code)
	}
	compose := filepath.Join(dir, "docker-compose.yml")
	if _, err := os.Stat(compose); !os.IsNotExist(err) {
		t.Errorf("compose file must not be written when preflight fails, stat err=%v", err)
	}
	if len(f.calls) != 1 {
		t.Fatalf("want exactly 1 runner call (the preflight probe), got %d: %+v", len(f.calls), f.calls)
	}
	if want := []string{"docker", "info"}; !reflect.DeepEqual(f.calls[0].argv, want) {
		t.Errorf("preflight argv = %v, want %v", f.calls[0].argv, want)
	}
}

func TestDeployPodmanPreflightFailureStopsBeforeWrite(t *testing.T) {
	f := useFakeRunner(t)
	f.err = fmt.Errorf("podman unreachable")
	quadletDir := t.TempDir()
	dir := t.TempDir()
	write(t, dir, "env.yaml", podmanEnv(quadletDir))
	write(t, dir, "10.yaml", validWF)

	if code := dispatch([]string{"deploy", "podman", "-e", filepath.Join(dir, "env.yaml")}, f); code != 1 {
		t.Fatalf("exit=%d, want 1", code)
	}
	for _, name := range []string{"solmq-conn-application.yml", "solmq-conn.container"} {
		if _, err := os.Stat(filepath.Join(quadletDir, name)); !os.IsNotExist(err) {
			t.Errorf("%s must not be written when preflight fails", name)
		}
	}
	if len(f.calls) != 1 {
		t.Fatalf("want exactly 1 runner call (the preflight probe), got %d: %+v", len(f.calls), f.calls)
	}
	if want := []string{"podman", "info"}; !reflect.DeepEqual(f.calls[0].argv, want) {
		t.Errorf("preflight argv = %v, want %v", f.calls[0].argv, want)
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
