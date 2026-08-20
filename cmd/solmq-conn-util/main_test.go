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

// sharedEnv carries the keys that are declared once for every platform. It is
// a separate const because the fixtures below get concatenated to build
// multi-section configs, and two copies of a top-level key is a yaml error --
// which is exactly what the tool would hit if these keys were still per-section.
const sharedEnv = `
image:
  name: img
  tag: "1"
timezone: UTC
`

// kubeEnv is a minimal-but-valid kubernetes: section for the deploy/remove/
// platform-resolution seam tests -- a safe command, and the fields checkKube
// requires.
const kubeEnv = `
kubernetes:
  command: kubectl
  deployment:
    name: solmq-connector
    namespace: solace-connectors
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

// withPromptAnswer overrides the injectable promptLine seam (the interactive
// platform menu and the status install confirmation) to return answer for
// every call during the test, restoring the previous value on cleanup. It
// never touches the real os.Stdin, so it carries no risk of blocking on a
// read regardless of how the test binary itself was invoked.
func withPromptAnswer(t *testing.T, answer string) {
	t.Helper()
	old := promptLine
	promptLine = func(string) (string, error) { return answer, nil }
	t.Cleanup(func() { promptLine = old })
}

// ---- deploy/remove seam helpers -------------------------------------------------

// fakeRunner records every invocation reaching the deploy/remove seam so tests
// can assert the exact argv/stdin/env without starting a process (mirrors
// internal/runner/runner_test.go's fakeRunner).
//
// err, when set, fails every call from index failFrom onward (0 by default, so
// it fails from the very first call -- the preflight probe, now that one always
// precedes the mutating command). A test that needs the probe to succeed but
// the real deploy/remove call to fail sets failFrom to the index of that call.
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
// production runner.OS{}) so the fake reaches the deploy/remove seam as an
// explicit argument rather than through mutable package state.
func useFakeRunner(t *testing.T) *fakeRunner {
	t.Helper()
	return &fakeRunner{}
}

// queuedResp is one (out, err) pair a queueRunner call returns.
type queuedResp struct {
	out string
	err error
}

// queueRunner is a fakeRunner variant status's tests need: each call in a
// sequence returns its own (out, err) pair (e.g. probe-absent, then
// install-succeeds, then run-standby) rather than fakeRunner's uniform
// "fail from index N onward". Calls beyond len(resp) succeed with no output.
type queueRunner struct {
	calls []fakeCall
	resp  []queuedResp
}

func (q *queueRunner) Run(c runner.Cmd) (string, error) {
	idx := len(q.calls)
	q.calls = append(q.calls, fakeCall{argv: c.Argv, stdin: c.Stdin, env: c.Env})
	if idx < len(q.resp) {
		return q.resp[idx].out, q.resp[idx].err
	}
	return "", nil
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
		{"generate bogus target", []string{"generate", "bogus"}, 2},
		{"deploy bogus platform", []string{"deploy", "bogus"}, 2},
		{"auto-complete no shell", []string{"auto-complete"}, 2},
		{"auto-complete bogus shell", []string{"auto-complete", "bogus"}, 2},
		// "completion" was the verb's old name; the rename to "auto-complete" is a
		// deliberate clean break with no compatibility alias, so this now exits 2
		// as an unknown command rather than reaching runAutoComplete.
		{"completion is no longer a command", []string{"completion", "bash"}, 2},
		// Deliberately unassigned near misses for the alias table below: none of
		// these short forms were picked as an alias, so each must stay unknown.
		{"near miss d", []string{"d"}, 2},
		{"near miss v", []string{"v"}, 2},
		{"near miss s", []string{"s"}, 2},
		{"near miss g", []string{"g"}, 2},
		{"near miss comp", []string{"comp"}, 2},
		{"near miss h", []string{"h"}, 2},
		{"near miss hlp", []string{"hlp"}, 2},
		{"near miss stat", []string{"stat"}, 2},
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

// TestAutoCompleteDispatchPrintsScript covers the success path the exit-code
// table cannot: every modeled shell exits 0, writes its script to stdout
// (never a file), and never reaches the runner -- rendering is compiled-in, so
// it needs no env.yaml and no exec.
func TestAutoCompleteDispatchPrintsScript(t *testing.T) {
	for _, shell := range targetNames("auto-complete") {
		t.Run(shell, func(t *testing.T) {
			f := &fakeRunner{}
			code := -1
			out := captureStdout(t, func() { code = dispatch([]string{"auto-complete", shell}, f) })
			if code != 0 {
				t.Errorf("dispatch(auto-complete %s) = %d, want 0", shell, code)
			}
			if !strings.Contains(out, "solmq-conn-util") {
				t.Errorf("auto-complete %s: stdout does not look like a completion script: %q", shell, truncate(out))
			}
			if len(f.calls) != 0 {
				t.Errorf("auto-complete %s: runner must not be invoked, got %d calls", shell, len(f.calls))
			}
		})
	}
}

// truncate keeps a failure message readable when a whole script would otherwise
// be dumped into it.
func truncate(s string) string {
	const max = 120
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}

// ---- a1) alias resolution ---------------------------------------------------

// TestVerbAliasesDispatchLikeCanonical checks each of the seven modeled
// aliases (cliVerb.Aliases in commands.go) reaches exactly the same handler as
// its canonical verb: the same trailing args, dispatched once under the alias
// and once under the canonical name, must produce the same exit code, the
// same stdout/stderr, and the same runner calls. Each rest is picked to fail
// fast (a bad positional, a missing env file, an unknown flag) so the pair
// never reaches the runner or touches the filesystem beyond a failed open,
// keeping the comparison cheap and deterministic.
func TestVerbAliasesDispatchLikeCanonical(t *testing.T) {
	missingEnv := filepath.Join(t.TempDir(), "does-not-exist.yaml")
	cases := []struct {
		canonical string
		alias     string
		rest      []string
	}{
		{"generate", "gen", []string{"bogus"}},
		{"deploy", "dep", []string{"bogus"}},
		{"remove", "rm", []string{"bogus"}},
		{"status", "sts", []string{"bogus"}},
		{"version", "ver", nil},
		{"validate", "vld", []string{"-e", missingEnv}},
		{"examples", "eg", []string{"-nope"}},
	}
	for _, c := range cases {
		t.Run(c.alias, func(t *testing.T) {
			fCanon, fAlias := &fakeRunner{}, &fakeRunner{}
			var canonCode, aliasCode int
			var canonErrOut, aliasErrOut string
			canonOut := captureStdout(t, func() {
				canonErrOut = captureStderr(t, func() {
					canonCode = dispatch(append([]string{c.canonical}, c.rest...), fCanon)
				})
			})
			aliasOut := captureStdout(t, func() {
				aliasErrOut = captureStderr(t, func() {
					aliasCode = dispatch(append([]string{c.alias}, c.rest...), fAlias)
				})
			})
			if aliasCode != canonCode {
				t.Errorf("dispatch(%s %v) = %d, want %d (canonical %s)", c.alias, c.rest, aliasCode, canonCode, c.canonical)
			}
			if aliasOut != canonOut {
				t.Errorf("dispatch(%s %v) stdout = %q, want %q (canonical %s)", c.alias, c.rest, aliasOut, canonOut, c.canonical)
			}
			if aliasErrOut != canonErrOut {
				t.Errorf("dispatch(%s %v) stderr = %q, want %q (canonical %s)", c.alias, c.rest, aliasErrOut, canonErrOut, c.canonical)
			}
			if !reflect.DeepEqual(fAlias.calls, fCanon.calls) {
				t.Errorf("dispatch(%s %v) runner calls = %v, want %v (canonical %s)", c.alias, c.rest, fAlias.calls, fCanon.calls, c.canonical)
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
// modeled entry, must fail. deploy/remove no longer model a positional Targets
// list (the platform is resolved via --platform, not looked up from args[0]),
// so their platform-map coverage is gated separately by
// TestPlatformMapsCoverThreeNames instead of against cliVerbs here.
func TestDispatchHandlersMatchModel(t *testing.T) {
	verbNames := make(map[string]bool, len(cliVerbs))
	for _, v := range cliVerbs {
		verbNames[v.Name] = true
	}
	assertSameNameSet(t, "verb", verbNames, keySet(verbHandlers))
	assertSameNameSet(t, "generate target", nameSet(targetNames("generate")), keySet(genTargets))
	assertSameNameSet(t, "completion shell", nameSet(targetNames("auto-complete")), keySet(completionShellRenderers))
}

// TestPlatformMapsCoverThreeNames guards every platform-keyed map in this file
// (generate's platform fallback, deploy/remove's platform dispatch) against the
// same platformNames list resolvePlatform validates --platform against, so a
// platform added to one and not the other fails here instead of drifting
// silently.
func TestPlatformMapsCoverThreeNames(t *testing.T) {
	assertSameNameSet(t, "generate platform renderer", nameSet(platformNames), keySet(platformGenerators))
	assertSameNameSet(t, "deploy/remove platform handler", nameSet(platformNames), keySet(actTargets))
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
	// Two separate runs: without pinning the status account's password each
	// would mint its own, so the comparison would fail on the one line that is
	// designed to differ. Pinning it via the documented override leaves the
	// rest of the document -- what this test is actually about -- comparable.
	t.Setenv(spec.StatusUserPasswordEnvVar, "pinned-for-reproducible-output")

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

// ---- g) deploy/remove via the runner seam ----------------------------------------
//
// These tests call dispatch directly (not run) so they can pass a fakeRunner
// explicitly instead of mutating package-level state.

func TestDeployKubernetesSeamHappyPath(t *testing.T) {
	f := useFakeRunner(t)
	dir := t.TempDir()
	write(t, dir, "env.yaml", sharedEnv+kubeEnv)
	write(t, dir, "10.yaml", validWF)

	if code := dispatch([]string{"deploy", "--platform", "kubernetes", "-e", filepath.Join(dir, "env.yaml")}, f); code != 0 {
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

func TestRemoveKubernetesSeamHappyPath(t *testing.T) {
	f := useFakeRunner(t)
	dir := t.TempDir()
	write(t, dir, "env.yaml", sharedEnv+kubeEnv)
	write(t, dir, "10.yaml", validWF)

	if code := dispatch([]string{"remove", "--platform", "kubernetes", "-e", filepath.Join(dir, "env.yaml")}, f); code != 0 {
		t.Fatalf("exit=%d", code)
	}
	if len(f.calls) != 2 {
		t.Fatalf("want 2 calls to the runner (preflight, remove), got %d: %+v", len(f.calls), f.calls)
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
	write(t, dir, "env.yaml", sharedEnv+kubeEnvUnsafeCommand)
	write(t, dir, "10.yaml", validWF)

	if code := dispatch([]string{"deploy", "--platform", "kubernetes", "-e", filepath.Join(dir, "env.yaml")}, f); code != 1 {
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
	// The shipped example's connections/workflows/tls all reference their
	// credentials via -env (never a literal), so generating it for real needs
	// every one of those host variables set. The reserved status account needs
	// nothing here: its password is generated when the override is unset.
	for _, v := range []string{
		"SOL_PASSWORD", "MQ_ARCHIVE_PASSWORD", "MQ_CORE_PASSWORD", "EDGE_SOL_PASSWORD",
		"TRUSTSTORE_PASSWORD", "KEYSTORE_PASSWORD",
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
// DNS-1123 name, a safe command and in-range ports; the image is a top-level key).
const dockerEnv = `
docker:
  command: docker
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
	write(t, dir, "env.yaml", sharedEnv+kubeEnv)
	write(t, dir, "10.yaml", validWF)
	var code int
	stdout := captureStdout(t, func() {
		code = run([]string{"generate", "--platform", "kubernetes", "-e", filepath.Join(dir, "env.yaml")})
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
	write(t, dir, "env.yaml", sharedEnv+dockerEnv)
	write(t, dir, "10.yaml", validWF)
	out := filepath.Join(t.TempDir(), "compose.yml")
	if code := run([]string{"generate", "--platform", "docker", "-e", filepath.Join(dir, "env.yaml"), "-o", out}); code != 0 {
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
	write(t, dir, "env.yaml", sharedEnv+podmanEnv(t.TempDir()))
	write(t, dir, "10.yaml", validWF)
	var code int
	stdout := captureStdout(t, func() {
		code = run([]string{"generate", "--platform", "podman", "-e", filepath.Join(dir, "env.yaml")})
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

// ---- deploy / remove: docker and podman seams -------------------------------------

func TestDeployDockerSeamWritesComposeAndRuns(t *testing.T) {
	f := useFakeRunner(t)
	dir := t.TempDir()
	write(t, dir, "env.yaml", sharedEnv+dockerEnv)
	write(t, dir, "10.yaml", validWF)

	if code := dispatch([]string{"deploy", "--platform", "docker", "-e", filepath.Join(dir, "env.yaml")}, f); code != 0 {
		t.Fatalf("exit=%d", code)
	}
	compose := filepath.Join(dir, "docker-compose.yml")
	// The compose file is scratch, regenerated by every deploy/remove, and is
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
	write(t, dir, "env.yaml", sharedEnv+dockerEnv)
	write(t, dir, "10.yaml", validWF)

	if code := dispatch([]string{"deploy", "--platform", "docker", "-e", filepath.Join(dir, "env.yaml")}, f); code != 1 {
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
	write(t, dir, "env.yaml", sharedEnv+dockerEnv)
	write(t, dir, "10.yaml", envCredWF)

	if code := dispatch([]string{"deploy", "--platform", "docker", "-e", filepath.Join(dir, "env.yaml")}, f); code != 0 {
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

func TestRemoveDockerSeam(t *testing.T) {
	f := useFakeRunner(t)
	dir := t.TempDir()
	write(t, dir, "env.yaml", sharedEnv+dockerEnv)
	write(t, dir, "10.yaml", validWF)

	if code := dispatch([]string{"remove", "--platform", "docker", "-e", filepath.Join(dir, "env.yaml")}, f); code != 0 {
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
	write(t, dir, "env.yaml", sharedEnv+podmanEnv(quadletDir))
	write(t, dir, "10.yaml", validWF)

	if code := dispatch([]string{"deploy", "--platform", "podman", "-e", filepath.Join(dir, "env.yaml")}, f); code != 0 {
		t.Fatalf("exit=%d", code)
	}
	// The app yaml, the status script, and the quadlet unit all land in the
	// (overridden) quadlet dir. application.yml now carries a live credential
	// (the reserved status account's password), so it is 0600; the status
	// script and unit carry no secret of their own, so both stay 0644.
	wantModes := map[string]os.FileMode{
		"solmq-conn-application.yml": 0o600,
		"solmq-conn-status":          0o644,
		"solmq-conn.container":       0o644,
	}
	for name, wantMode := range wantModes {
		info, err := os.Stat(filepath.Join(quadletDir, name))
		if err != nil {
			t.Errorf("%s not written to quadlet dir: %v", name, err)
			continue
		}
		// Unix perms are not faithfully reproduced on the windows-2025 CI runner.
		if runtime.GOOS != "windows" && info.Mode().Perm() != wantMode {
			t.Errorf("%s mode = %v, want %v", name, info.Mode().Perm(), wantMode)
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

func TestRemovePodmanSeamStopsRemovesReloads(t *testing.T) {
	f := useFakeRunner(t)
	quadletDir := t.TempDir()
	dir := t.TempDir()
	write(t, dir, "env.yaml", sharedEnv+podmanEnv(quadletDir))
	write(t, dir, "10.yaml", validWF)
	// Pre-seed the files a deploy would have written; remove must clear all
	// three: the unit, the generated app yaml, and the status script.
	write(t, quadletDir, "solmq-conn.container", "[Container]\n")
	write(t, quadletDir, "solmq-conn-application.yml", "x\n")
	write(t, quadletDir, "solmq-conn-status", "#!/bin/sh\n")

	if code := dispatch([]string{"remove", "--platform", "podman", "-e", filepath.Join(dir, "env.yaml")}, f); code != 0 {
		t.Fatalf("exit=%d", code)
	}
	// Secrets are removed from podman's store only AFTER the unit is stopped and
	// the generator reloaded, mirroring deploy's create-before-start ordering in
	// reverse: a failure removing them still surfaces (see main.go's
	// podmanRemove), but the units referencing them are gone first. Call 0 is
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
	for _, name := range []string{"solmq-conn.container", "solmq-conn-application.yml", "solmq-conn-status"} {
		if _, err := os.Stat(filepath.Join(quadletDir, name)); !os.IsNotExist(err) {
			t.Errorf("%s should be removed by the remove verb", name)
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
	write(t, dir, "env.yaml", sharedEnv+kubeEnv)
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
			code := dispatch([]string{"deploy", "--platform", "kubernetes", "-e", envPath, "--allow-command", c.val}, f)
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
// deploy/remove/status-only: generate config and validate never register the
// flag, so passing it is an unknown-flag usage error, same as any other
// undefined flag.
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
	write(t, dir, "env.yaml", sharedEnv+podmanEnvSudo(quadletDir))
	write(t, dir, "10.yaml", validWF)
	envPath := filepath.Join(dir, "env.yaml")

	fReject := useFakeRunner(t)
	if code := dispatch([]string{"deploy", "--platform", "podman", "-e", envPath}, fReject); code != 1 {
		t.Fatalf("without --allow-command: exit=%d, want 1", code)
	}
	if len(fReject.calls) != 0 {
		t.Fatalf("without --allow-command: runner must not be invoked, got %d calls", len(fReject.calls))
	}

	fAccept := useFakeRunner(t)
	args := []string{"deploy", "--platform", "podman", "-e", envPath, "--allow-command", "sudo", "--allow-command", "sudo"}
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
	write(t, dir, "env.yaml", sharedEnv+kubeEnv)
	write(t, dir, "10.yaml", validWF)

	if code := dispatch([]string{"deploy", "--platform", "kubernetes", "-e", filepath.Join(dir, "env.yaml")}, f); code != 1 {
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
	write(t, dir, "env.yaml", sharedEnv+dockerEnv)
	write(t, dir, "10.yaml", validWF)

	if code := dispatch([]string{"deploy", "--platform", "docker", "-e", filepath.Join(dir, "env.yaml")}, f); code != 1 {
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
	write(t, dir, "env.yaml", sharedEnv+podmanEnv(quadletDir))
	write(t, dir, "10.yaml", validWF)

	if code := dispatch([]string{"deploy", "--platform", "podman", "-e", filepath.Join(dir, "env.yaml")}, f); code != 1 {
		t.Fatalf("exit=%d, want 1", code)
	}
	for _, name := range []string{"solmq-conn-application.yml", "solmq-conn-status", "solmq-conn.container"} {
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

// ---- l) --platform resolution ----------------------------------------------------
//
// These pin resolvePlatform's shared order (platformResolutionDetail in
// commands.go): --platform, then single-section inference, then an
// interactive menu, then a loud error -- shared by generate/deploy/remove
// (status has its own matrix further down, since it also has the
// explicit-target exception).

// TestPlatformFlagHitOverridesInference asserts --platform wins even when a
// different section could have been inferred (kubernetes is also present).
func TestPlatformFlagHitOverridesInference(t *testing.T) {
	f := useFakeRunner(t)
	dir := t.TempDir()
	write(t, dir, "env.yaml", sharedEnv+kubeEnv+dockerEnv)
	write(t, dir, "10.yaml", validWF)

	if code := dispatch([]string{"deploy", "--platform", "docker", "-e", filepath.Join(dir, "env.yaml")}, f); code != 0 {
		t.Fatalf("exit=%d, want 0", code)
	}
	if want := []string{"docker", "info"}; len(f.calls) == 0 || !reflect.DeepEqual(f.calls[0].argv, want) {
		t.Errorf("explicit --platform docker should be used even though kubernetes is also present, got %+v", f.calls)
	}
}

// TestPlatformAliasesResolveToCanonical asserts each short --platform spelling
// selects the same platform as its canonical name. Resolution happens before
// validation and before the section lookup, so an alias works anywhere the full
// name does. The assertion is on the binary the run reaches for rather than the
// whole argv: which platform got selected is the claim, and the exact preflight
// arguments are pinned by the per-platform deploy tests already.
func TestPlatformAliasesResolveToCanonical(t *testing.T) {
	for _, c := range []struct{ alias, wantBinary string }{
		{alias: "kube", wantBinary: "kubectl"},
		{alias: "dk", wantBinary: "docker"},
		{alias: "pm", wantBinary: "podman"},
	} {
		f := useFakeRunner(t)
		dir := t.TempDir()
		write(t, dir, "env.yaml", sharedEnv+kubeEnv+dockerEnv+podmanEnv(t.TempDir()))
		write(t, dir, "10.yaml", validWF)

		if code := dispatch([]string{"deploy", "--platform", c.alias, "-e", filepath.Join(dir, "env.yaml")}, f); code != 0 {
			t.Fatalf("--platform %s: exit=%d, want 0", c.alias, code)
		}
		if len(f.calls) == 0 {
			t.Fatalf("--platform %s reached no command", c.alias)
		}
		if got := f.calls[0].argv[0]; got != c.wantBinary {
			t.Errorf("--platform %s reached %q, want %q", c.alias, got, c.wantBinary)
		}
	}
}

// TestPlatformAliasMissingSectionNamesCanonicalSection asserts the alias is
// resolved before the section check, so the error tells the operator which
// section: key to add rather than echoing the alias they typed.
func TestPlatformAliasMissingSectionNamesCanonicalSection(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "env.yaml", sharedEnv+dockerEnv)
	f := &fakeRunner{}
	stderr := captureStderr(t, func() {
		code := dispatch([]string{"deploy", "--platform", "kube", "-e", filepath.Join(dir, "env.yaml")}, f)
		if code != 1 {
			t.Fatalf("exit=%d, want 1", code)
		}
	})
	if !strings.Contains(stderr, "kubernetes") {
		t.Errorf("error should name the kubernetes section, not just the alias, got:\n%s", stderr)
	}
	if len(f.calls) != 0 {
		t.Errorf("runner must not be invoked, got %d calls", len(f.calls))
	}
}

// TestPlatformUnknownValueListsEverySpelling asserts a bogus --platform value
// is rejected with all accepted spellings, canonical and short, so the operator
// can see the alias set without reaching for the docs.
func TestPlatformUnknownValueListsEverySpelling(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "env.yaml", sharedEnv+dockerEnv)
	f := &fakeRunner{}
	stderr := captureStderr(t, func() {
		code := dispatch([]string{"deploy", "--platform", "k8s", "-e", filepath.Join(dir, "env.yaml")}, f)
		if code != 1 {
			t.Fatalf("exit=%d, want 1", code)
		}
	})
	for _, want := range []string{"kubernetes", "docker", "podman", "kube", "dk", "pm"} {
		if !strings.Contains(stderr, want) {
			t.Errorf("rejection should list %q, got:\n%s", want, stderr)
		}
	}
	if len(f.calls) != 0 {
		t.Errorf("runner must not be invoked, got %d calls", len(f.calls))
	}
}

// TestPlatformSpellingsAreDeterministic pins the ordering contract behind the
// rejection message: platformSpellings is built from an ordered slice, not from
// map iteration, so the message cannot vary between runs.
func TestPlatformSpellingsAreDeterministic(t *testing.T) {
	first := platformSpellings()
	for i := 0; i < 20; i++ {
		if got := platformSpellings(); !reflect.DeepEqual(got, first) {
			t.Fatalf("platformSpellings varies between calls: %v then %v", first, got)
		}
	}
	// Canonical names lead, so the message reads as names-then-shorthand.
	for i, want := range platformNames {
		if first[i] != want {
			t.Errorf("position %d = %q, want the canonical %q first", i, first[i], want)
		}
	}
	if len(first) != len(platformNames)+len(platformAliasList) {
		t.Errorf("spellings = %v, want every canonical name plus every alias", first)
	}
}

// TestPlatformAliasesCoverEveryPlatformExactlyOnce guards the table itself: an
// alias pointing at an unknown platform, or a duplicate alias, would silently
// misroute a --platform value.
func TestPlatformAliasesCoverEveryPlatformExactlyOnce(t *testing.T) {
	seen := map[string]bool{}
	for _, e := range platformAliasList {
		if !contains(platformNames, e.Canonical) {
			t.Errorf("alias %q maps to %q, which is not a platform", e.Alias, e.Canonical)
		}
		if seen[e.Alias] {
			t.Errorf("alias %q is declared twice", e.Alias)
		}
		seen[e.Alias] = true
		if contains(platformNames, e.Alias) {
			t.Errorf("alias %q collides with a canonical platform name", e.Alias)
		}
	}
	if len(platformAliases) != len(platformAliasList) {
		t.Errorf("lookup map has %d entries for %d declared aliases", len(platformAliases), len(platformAliasList))
	}
}

// TestPlatformFlagMissingSectionIsLoudError asserts a --platform value with no
// matching section in env.yaml fails loudly and names the sections that ARE
// present, before anything reaches the runner.
func TestPlatformFlagMissingSectionIsLoudError(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "env.yaml", sharedEnv+dockerEnv)
	f := &fakeRunner{}
	stderr := captureStderr(t, func() {
		code := dispatch([]string{"deploy", "--platform", "kubernetes", "-e", filepath.Join(dir, "env.yaml")}, f)
		if code != 1 {
			t.Fatalf("exit=%d, want 1", code)
		}
	})
	if !strings.Contains(stderr, "kubernetes") || !strings.Contains(stderr, "docker") {
		t.Errorf("stderr should name both the requested and the present sections, got:\n%s", stderr)
	}
	if len(f.calls) != 0 {
		t.Errorf("runner must not be invoked, got %d calls", len(f.calls))
	}
}

// TestPlatformSingleSectionInferred asserts that with no --platform and
// exactly one section present, that section is used and echoed to stderr.
func TestPlatformSingleSectionInferred(t *testing.T) {
	f := useFakeRunner(t)
	dir := t.TempDir()
	write(t, dir, "env.yaml", sharedEnv+kubeEnv)
	write(t, dir, "10.yaml", validWF)

	stderr := captureStderr(t, func() {
		code := dispatch([]string{"deploy", "-e", filepath.Join(dir, "env.yaml")}, f)
		if code != 0 {
			t.Fatalf("exit=%d, want 0", code)
		}
	})
	if !strings.Contains(stderr, "platform: kubernetes") {
		t.Errorf("stderr should echo the inferred platform, got:\n%s", stderr)
	}
	if len(f.calls) != 2 {
		t.Errorf("want 2 runner calls (preflight, apply), got %d: %+v", len(f.calls), f.calls)
	}
}

// TestPlatformMenuOnMultipleSections asserts that with no --platform and more
// than one section present, the interactive menu (via the injected promptLine
// seam) picks the platform.
func TestPlatformMenuOnMultipleSections(t *testing.T) {
	withPromptAnswer(t, "2") // present order is [kubernetes, docker]; 2 = docker
	f := useFakeRunner(t)
	dir := t.TempDir()
	write(t, dir, "env.yaml", sharedEnv+kubeEnv+dockerEnv)
	write(t, dir, "10.yaml", validWF)

	if code := dispatch([]string{"deploy", "-e", filepath.Join(dir, "env.yaml")}, f); code != 0 {
		t.Fatalf("exit=%d, want 0", code)
	}
	if want := []string{"docker", "info"}; len(f.calls) == 0 || !reflect.DeepEqual(f.calls[0].argv, want) {
		t.Errorf("menu choice 2 should select docker, got %+v", f.calls)
	}
}

// TestPlatformMenuNonTTYRefusesWithPlatformHint asserts the menu refuses to
// block when stdin is not a terminal, instead of hanging on a read that would
// never return, and that the resulting error points at --platform. Stdin is
// swapped for a pipe (never a character device) rather than relying on
// whatever terminal state the test binary happened to inherit, so this cannot
// flake or block regardless of how the tests are invoked.
func TestPlatformMenuNonTTYRefusesWithPlatformHint(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "env.yaml", sharedEnv+kubeEnv+dockerEnv)

	oldStdin := os.Stdin
	pr, pw, perr := os.Pipe()
	if perr != nil {
		t.Fatal(perr)
	}
	os.Stdin = pr
	t.Cleanup(func() {
		os.Stdin = oldStdin
		_ = pw.Close()
		_ = pr.Close()
	})

	f := &fakeRunner{}
	stderr := captureStderr(t, func() {
		code := dispatch([]string{"deploy", "-e", filepath.Join(dir, "env.yaml")}, f)
		if code != 1 {
			t.Fatalf("exit=%d, want 1", code)
		}
	})
	if !strings.Contains(stderr, "--platform") {
		t.Errorf("non-tty menu error should mention --platform, got:\n%s", stderr)
	}
	if len(f.calls) != 0 {
		t.Errorf("runner must not be invoked, got %d calls", len(f.calls))
	}
}

// TestPlatformZeroSectionsIsLoudError asserts that with no --platform and no
// section present at all, the error names all three section keys.
func TestPlatformZeroSectionsIsLoudError(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "env.yaml", bareEnv)
	f := &fakeRunner{}
	stderr := captureStderr(t, func() {
		code := dispatch([]string{"deploy", "-e", filepath.Join(dir, "env.yaml")}, f)
		if code != 1 {
			t.Fatalf("exit=%d, want 1", code)
		}
	})
	for _, want := range []string{"kubernetes:", "docker:", "podman:"} {
		if !strings.Contains(stderr, want) {
			t.Errorf("stderr should name %q, got:\n%s", want, stderr)
		}
	}
	if len(f.calls) != 0 {
		t.Errorf("runner must not be invoked, got %d calls", len(f.calls))
	}
}

// TestOldPositionalFormsRejectedWithPlatformHint asserts the pre-rework
// positional grammar (deploy kubernetes, generate docker, ...) is now a usage
// error (exit 2) that points at --platform, rather than being resolved.
func TestOldPositionalFormsRejectedWithPlatformHint(t *testing.T) {
	cases := [][]string{
		{"deploy", "kubernetes"},
		{"remove", "docker"},
		{"generate", "podman"},
	}
	for _, args := range cases {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			f := &fakeRunner{}
			stderr := captureStderr(t, func() {
				code := dispatch(args, f)
				if code != 2 {
					t.Errorf("%v: exit=%d, want 2", args, code)
				}
			})
			if !strings.Contains(stderr, "--platform") {
				t.Errorf("%v: stderr should hint at --platform, got:\n%s", args, stderr)
			}
			if len(f.calls) != 0 {
				t.Errorf("%v: runner must not be invoked, got %d calls", args, len(f.calls))
			}
		})
	}
}

// ---- m) status ------------------------------------------------------------------

// TestStatusScriptPresentRunsAndReportsOutput covers the simplest matrix
// entry: the script is already installed, so status just runs it and prints
// its output under the target's banner, every line indented by two.
func TestStatusScriptPresentRunsAndReportsOutput(t *testing.T) {
	q := &queueRunner{resp: []queuedResp{
		{"", nil},                                // preflight: kubectl auth can-i create deployment
		{runner.ScriptPresentMarker + "\n", nil}, // ScriptInstalled: present
		{"leader-election mode: standalone\nleader-election state: active\n", nil}, // RunStatusScript
	}}
	var code int
	stdout := captureStdout(t, func() {
		code = dispatch([]string{"status", "--platform", "kubernetes", "--pod", "pod-a"}, q)
	})
	if code != 0 {
		t.Fatalf("exit=%d, want 0", code)
	}
	// No env.yaml behind an explicit --pod, so the banner carries the platform
	// and the pod alone -- an unresolved namespace or deployment is dropped
	// rather than printed as an empty segment.
	want := "=== kubernetes  pod-a ===\n" +
		"  leader-election mode: standalone\n" +
		"  leader-election state: active\n"
	if stdout != want {
		t.Errorf("stdout = %q, want %q", stdout, want)
	}
	if len(q.calls) != 3 {
		t.Fatalf("want 3 runner calls, got %d: %+v", len(q.calls), q.calls)
	}
}

// TestStatusAbsentPlusInstallFlagInstallsThenRuns covers --install: a missing
// script is installed without any prompt, then run.
func TestStatusAbsentPlusInstallFlagInstallsThenRuns(t *testing.T) {
	q := &queueRunner{resp: []queuedResp{
		{"", nil},                               // preflight
		{runner.ScriptAbsentMarker + "\n", nil}, // ScriptInstalled: absent
		{"", nil},                               // InstallScript
		{"leader-election mode: standalone\nleader-election state: active\n", nil}, // RunStatusScript
	}}
	if code := dispatch([]string{"status", "--install", "--platform", "kubernetes", "--pod", "pod-a"}, q); code != 0 {
		t.Fatalf("exit=%d, want 0", code)
	}
	if len(q.calls) != 4 {
		t.Fatalf("want 4 runner calls, got %d: %+v", len(q.calls), q.calls)
	}
}

// TestStatusAbsentPromptYesInstallsThenRuns covers the same absent-script case
// without --install: one prompt (via the injected promptLine seam) answered
// "y" installs, then runs.
func TestStatusAbsentPromptYesInstallsThenRuns(t *testing.T) {
	withPromptAnswer(t, "y")
	q := &queueRunner{resp: []queuedResp{
		{"", nil},                               // preflight
		{runner.ScriptAbsentMarker + "\n", nil}, // probe: absent
		{"", nil},                               // install
		{"leader-election mode: standalone\nleader-election state: active\n", nil}, // run
	}}
	if code := dispatch([]string{"status", "--platform", "kubernetes", "--pod", "pod-a"}, q); code != 0 {
		t.Fatalf("exit=%d, want 0", code)
	}
	if len(q.calls) != 4 {
		t.Fatalf("want 4 runner calls, got %d: %+v", len(q.calls), q.calls)
	}
}

// TestStatusAbsentPromptNoSkipsAndExitsOne covers declining the install
// prompt: the target is skipped (never installed, never run) and the overall
// exit code is 1.
func TestStatusAbsentPromptNoSkipsAndExitsOne(t *testing.T) {
	withPromptAnswer(t, "n")
	q := &queueRunner{resp: []queuedResp{
		{"", nil},                               // preflight
		{runner.ScriptAbsentMarker + "\n", nil}, // probe: absent
	}}
	stderr := captureStderr(t, func() {
		code := dispatch([]string{"status", "--platform", "kubernetes", "--pod", "pod-a"}, q)
		if code != 1 {
			t.Fatalf("exit=%d, want 1", code)
		}
	})
	if !strings.Contains(stderr, "pod-a: skipped") {
		t.Errorf("stderr should report the skipped target, got:\n%s", stderr)
	}
	if len(q.calls) != 2 {
		t.Fatalf("want 2 runner calls (preflight, probe) -- no install, no run -- got %d: %+v", len(q.calls), q.calls)
	}
}

// TestStatusInstallPromptNonTTYRefusesWithInstallHint is the confirmInstall
// counterpart to TestPlatformMenuNonTTYRefusesWithPlatformHint: the install
// confirmation shares the same stdin seam, so it must also refuse rather than
// block on a read that would never return, and point at --install (not
// --platform, which is already satisfied here).
func TestStatusInstallPromptNonTTYRefusesWithInstallHint(t *testing.T) {
	oldStdin := os.Stdin
	pr, pw, perr := os.Pipe()
	if perr != nil {
		t.Fatal(perr)
	}
	os.Stdin = pr
	t.Cleanup(func() {
		os.Stdin = oldStdin
		_ = pw.Close()
		_ = pr.Close()
	})

	q := &queueRunner{resp: []queuedResp{
		{"", nil},                               // preflight
		{runner.ScriptAbsentMarker + "\n", nil}, // probe: script absent
	}}
	stderr := captureStderr(t, func() {
		code := dispatch([]string{"status", "--platform", "kubernetes", "--pod", "pod-a"}, q)
		if code != 1 {
			t.Fatalf("exit=%d, want 1", code)
		}
	})
	if !strings.Contains(stderr, "--install") {
		t.Errorf("non-tty install prompt error should mention --install, got:\n%s", stderr)
	}
	if len(q.calls) != 2 {
		t.Fatalf("want 2 runner calls (preflight, probe) -- nothing installed or run -- got %d: %+v", len(q.calls), q.calls)
	}
}

// TestStatusBannerNamesTheInstancePerPlatform covers the identity line above
// each report. It exists because the report itself cannot carry any of this:
// the script runs inside the container and knows nothing of the namespace,
// deployment or compose project that locate it from outside.
//
// Every unresolved name is dropped rather than rendered as an empty segment,
// which is what an explicit --pod with no env.yaml, or a container compose did
// not create, actually produces.
func TestStatusBannerNamesTheInstancePerPlatform(t *testing.T) {
	tests := []struct {
		name                                           string
		platform, namespace, deployment, group, target string
		want                                           string
	}{
		{
			name: "kubernetes, everything resolved", platform: "kubernetes",
			namespace: "prod", deployment: "solmq-connector", target: "solmq-connector-7d9f8c6b5-x2n4q",
			want: "kubernetes  prod / solmq-connector / solmq-connector-7d9f8c6b5-x2n4q",
		},
		{
			name: "kubernetes, explicit --pod with no env.yaml", platform: "kubernetes",
			target: "pod-a", want: "kubernetes  pod-a",
		},
		{
			name: "docker, container in a compose project", platform: "docker",
			group: "eg", target: "solmq-connector", want: "docker  eg / solmq-connector",
		},
		{
			name: "docker, container compose did not create", platform: "docker",
			target: "solmq-connector", want: "docker  solmq-connector",
		},
		{
			name: "podman, container only", platform: "podman",
			target: "solmq-connector", want: "podman  solmq-connector",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := statusBanner(tt.platform, tt.namespace, tt.deployment, tt.group, tt.target); got != tt.want {
				t.Errorf("statusBanner = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestStatusDockerBannerCarriesComposeProject pins the one platform whose
// banner needs a second lookup: the compose project is read back from the
// container's own label rather than guessed from the compose file's directory,
// so it costs one read-only inspect per target, after the report is run.
func TestStatusDockerBannerCarriesComposeProject(t *testing.T) {
	q := &queueRunner{resp: []queuedResp{
		{"", nil},                                // preflight: docker info
		{runner.ScriptPresentMarker + "\n", nil}, // probe: present
		{"leader-election mode: standalone\nleader-election state: active\n", nil}, // run
		{"eg\n", nil}, // inspect: the compose project label
	}}
	var code int
	stdout := captureStdout(t, func() {
		code = dispatch([]string{"status", "--platform", "docker", "--container", "solmq-connector"}, q)
	})
	if code != 0 {
		t.Fatalf("exit=%d, want 0", code)
	}
	if !strings.Contains(stdout, "=== docker  eg / solmq-connector ===") {
		t.Errorf("banner missing the compose project, got:\n%s", stdout)
	}
	last := q.calls[len(q.calls)-1].argv
	if len(last) < 2 || last[1] != "inspect" {
		t.Errorf("last call = %v, want a docker inspect", last)
	}
}

// TestStatusDockerBannerWithoutComposeProject covers the same path when the
// container was not created by compose: the inspect answers nothing usable and
// the group segment is dropped instead of leaving a dangling separator.
func TestStatusDockerBannerWithoutComposeProject(t *testing.T) {
	q := &queueRunner{resp: []queuedResp{
		{"", nil},
		{runner.ScriptPresentMarker + "\n", nil},
		{"leader-election mode: standalone\nleader-election state: active\n", nil},
		{"<no value>\n", nil}, // inspect: no such label
	}}
	var code int
	stdout := captureStdout(t, func() {
		code = dispatch([]string{"status", "--platform", "docker", "--container", "solmq-connector"}, q)
	})
	if code != 0 {
		t.Fatalf("exit=%d, want 0", code)
	}
	if !strings.Contains(stdout, "=== docker  solmq-connector ===") {
		t.Errorf("banner should carry the container alone, got:\n%s", stdout)
	}
}

// TestStatusRejectsUnsafeUserBeforeAnyExec pins the boundary check on --user:
// the account name reaches a sed address inside the generated script, so a
// value carrying a regex or path metacharacter must be refused up front rather
// than silently producing an unauthenticated request.
func TestStatusRejectsUnsafeUserBeforeAnyExec(t *testing.T) {
	for _, user := range []string{"bad/user", "who$e", "a b", "quote'd"} {
		f := &fakeRunner{}
		code := dispatch([]string{"status", "--platform", "docker", "--container", "c1", "--user", user}, f)
		if code != 1 {
			t.Errorf("--user %q: exit=%d, want 1", user, code)
		}
		if len(f.calls) != 0 {
			t.Errorf("--user %q: runner must not be invoked, got %d calls", user, len(f.calls))
		}
	}
}

// TestStatusStandbyReportedWithoutAbortingOtherTargets asserts that standby is
// reported as the ordinary answer it is: the script always exits 0 and carries
// the state in its output, so a standby target prints like any other, the loop
// still reaches the next target, and the overall exit code stays 0.
func TestStatusStandbyReportedWithoutAbortingOtherTargets(t *testing.T) {
	// Every target is probed before any is run, so one install prompt can list
	// all the missing ones -- hence both probes ahead of both runs here.
	q := &queueRunner{resp: []queuedResp{
		{"", nil},                                // preflight
		{runner.ScriptPresentMarker + "\n", nil}, // pod-a probe: present
		{runner.ScriptPresentMarker + "\n", nil}, // pod-b probe: present
		{"leader-election mode: standalone\nleader-election state: standby\n", nil}, // pod-a run: standby
		{"leader-election mode: standalone\nleader-election state: active\n", nil},  // pod-b run: active
	}}
	var code int
	stdout := captureStdout(t, func() {
		code = dispatch([]string{"status", "--platform", "kubernetes", "--pod", "pod-a", "--pod", "pod-b"}, q)
	})
	if code != 0 {
		t.Fatalf("exit=%d, want 0 (standby is an answer, not a failure)", code)
	}
	for _, want := range []string{
		"=== kubernetes  pod-a ===\n  leader-election mode: standalone", "  leader-election state: standby",
		"=== kubernetes  pod-b ===\n  leader-election mode: standalone", "  leader-election state: active",
	} {
		if !strings.Contains(stdout, want) {
			t.Errorf("stdout missing %q, got:\n%s", want, stdout)
		}
	}
	if len(q.calls) != 5 {
		t.Fatalf("want 5 runner calls, got %d: %+v", len(q.calls), q.calls)
	}
}

// TestStatusRunFailureReportedAndExitsOne is the counterpart to the standby
// case: since the script itself always exits 0, a non-zero exit can only mean
// the exec failed, so that target is reported as an error and the overall exit
// code is 1 -- while a reachable target in the same run still prints normally.
func TestStatusRunFailureReportedAndExitsOne(t *testing.T) {
	q := &queueRunner{resp: []queuedResp{
		{"", nil},                                // preflight
		{runner.ScriptPresentMarker + "\n", nil}, // pod-a probe: present
		{runner.ScriptPresentMarker + "\n", nil}, // pod-b probe: present
		{"Error from server: pods \"pod-a\" not found\n", fmt.Errorf("exit status 1")}, // pod-a run: exec failed
		{"leader-election mode: standalone\nleader-election state: active\n", nil},     // pod-b run: fine
	}}
	var code int
	var stdout string
	stderr := captureStderr(t, func() {
		stdout = captureStdout(t, func() {
			code = dispatch([]string{"status", "--platform", "kubernetes", "--pod", "pod-a", "--pod", "pod-b"}, q)
		})
	})
	if code != 1 {
		t.Fatalf("exit=%d, want 1 (a target that could not be run)", code)
	}
	if !strings.Contains(stderr, "pod-a") {
		t.Errorf("stderr should name the failed target, got:\n%s", stderr)
	}
	if !strings.Contains(stdout, "=== kubernetes  pod-b ===") {
		t.Errorf("a reachable target must still report, got:\n%s", stdout)
	}
}

// TestStatusTargetValidationRejectsBadPodAndNamespace asserts a bad
// operator-supplied --pod or --namespace value is rejected before any exec,
// reusing validate.SafeToken rather than a hand-rolled check.
func TestStatusTargetValidationRejectsBadPodAndNamespace(t *testing.T) {
	cases := []struct {
		name string
		args []string
	}{
		{"bad pod name", []string{"status", "--platform", "kubernetes", "--pod", "bad pod name"}},
		{"bad namespace", []string{"status", "--platform", "kubernetes", "--pod", "goodpod", "--namespace", "bad;ns"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			f := &fakeRunner{}
			if code := dispatch(c.args, f); code != 1 {
				t.Errorf("%s: exit=%d, want 1", c.name, code)
			}
			if len(f.calls) != 0 {
				t.Errorf("%s: runner must not be invoked, got %d calls", c.name, len(f.calls))
			}
		})
	}
}

// TestStatusManagementPortBounds asserts an out-of-range --management-port is
// rejected before any exec (0 is not tested: it is the "not given, use the
// default" sentinel, not a value an operator would pass to mean "port 0").
func TestStatusManagementPortBounds(t *testing.T) {
	for _, p := range []string{"-1", "65536"} {
		t.Run(p, func(t *testing.T) {
			f := &fakeRunner{}
			args := []string{"status", "--platform", "kubernetes", "--pod", "pod-a", "--management-port", p}
			if code := dispatch(args, f); code != 1 {
				t.Errorf("--management-port %s: exit=%d, want 1", p, code)
			}
			if len(f.calls) != 0 {
				t.Errorf("--management-port %s: runner must not be invoked, got %d calls", p, len(f.calls))
			}
		})
	}
}

// ---- n) version -------------------------------------------------------------------

// TestVersionOutputShape pins the exact printed shape in an un-injected test
// build, where the package-level version var still holds its "dev" default.
func TestVersionOutputShape(t *testing.T) {
	var code int
	stdout := captureStdout(t, func() { code = run([]string{"version"}) })
	if code != 0 {
		t.Fatalf("exit=%d, want 0", code)
	}
	want := fmt.Sprintf("solmq-conn-util dev %s %s/%s\n", runtime.Version(), runtime.GOOS, runtime.GOARCH)
	if stdout != want {
		t.Errorf("stdout = %q, want %q", stdout, want)
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
