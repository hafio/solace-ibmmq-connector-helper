package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/libs"
	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/runner"
	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/spec"
	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/statusreport"
	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/validate"
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

// withPromptAnswers is withPromptAnswer for a run that asks more than once,
// answering each prompt in turn. remove is the case that needs it: the teardown
// and the namespace are deliberately separate questions, so a test that wants
// "yes, tear down" then "no, keep the namespace" cannot use one canned answer
// for both. Answers beyond the list repeat the last one.
func withPromptAnswers(t *testing.T, answers ...string) {
	t.Helper()
	old := promptLine
	i := 0
	promptLine = func(string) (string, error) {
		a := answers[len(answers)-1]
		if i < len(answers) {
			a = answers[i]
		}
		i++
		return a, nil
	}
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
		// Requested help is never an error, wherever it is asked: after a verb,
		// as the help verb's argument, or somewhere among the flags (routed
		// through the flag parser as ErrHelp).
		{"verb help short", []string{"status", "-h"}, 0},
		{"cli help", []string{"cli", "-h"}, 0},
		{"verb help long", []string{"deploy", "--help"}, 0},
		{"verb help via flag parser", []string{"examples", "-f", "-h"}, 0},
		{"help with a command", []string{"help", "status"}, 0},
		{"help with an alias", []string{"help", "sts"}, 0},
		{"help with an unknown command", []string{"help", "bogus"}, 2},
		{"unknown flag", []string{"generate", "config", "-nope"}, 2},
		{"missing env file", []string{"generate", "config", "-e", filepath.Join(validDir, "does-not-exist.yaml")}, 1},
		{"invalid spec", []string{"generate", "config", "-e", filepath.Join(invalidDir, "env.yaml")}, 1},
		{"generate bogus target", []string{"generate", "bogus"}, 2},
		{"deploy bogus platform", []string{"deploy", "bogus"}, 2},
		{"download missing target", []string{"download"}, 2},
		// cli is the one verb whose exit code can leave this table, but only
		// once a session is actually running inside an instance. Everything
		// decided from the command line alone still answers 2.
		{"cli unexpected positional", []string{"cli", "bogus"}, 2},
		{"cli separator with no command", []string{"cli", "--platform", "docker", "--container", "c", "--"}, 2},
		{"download unknown target", []string{"download", "bogus"}, 2},
		{"download missing set", []string{"download", "jar"}, 2},
		{"download unknown set", []string{"download", "jar", "bogus"}, 2},
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
		// "dep" was deploy's short form before it was aligned with solace-util's
		// "dp". Same clean-break rule as the "completion" rename above: the old
		// spelling is not kept as a compatibility alias, and pinning it here is
		// what stops it quietly coming back.
		{"dep is no longer deploy", []string{"dep"}, 2},
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

// TestVerbAliasesDispatchLikeCanonical checks each of the nine modeled
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
		{"deploy", "dp", []string{"bogus"}},
		{"remove", "rm", []string{"bogus"}},
		{"status", "sts", []string{"bogus"}},
		{"logs", "lg", []string{"bogus"}},
		{"version", "ver", nil},
		{"validate", "vld", []string{"-e", missingEnv}},
		{"examples", "eg", []string{"-nope"}},
		{"download", "dl", []string{"jar", "bogus"}},
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

// TestGenerateConfigTargetAliasResolves covers the `cfg` short spelling the same
// way TestStatusTargetAliasesResolve covers cnt/app: the model declares it, so
// the positional must reach genConfig through resolveTarget rather than a
// literal comparison against the canonical word. Without the resolve call the
// alias is documented but exits 2, which is the failure this pins.
func TestGenerateConfigTargetAliasResolves(t *testing.T) {
	t.Setenv(spec.StatusUserPasswordEnvVar, "pinned-for-reproducible-output")

	if got := resolveTarget("generate", "cfg"); got != tgtConfig {
		t.Errorf("resolveTarget(generate, cfg) = %q, want %q", got, tgtConfig)
	}

	envPath := filepath.Join(workflowDir(t, validWF), "env.yaml")
	var code int
	stdout := captureStdout(t, func() { code = run([]string{"gen", "cfg", "-e", envPath}) })
	if code != 0 {
		t.Fatalf("`gen cfg` exit=%d, want 0", code)
	}
	if !strings.Contains(stdout, "spring:") {
		t.Errorf("`gen cfg` stdout missing spring:\n%s", stdout)
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
	// Three responses: the preflight probe, the manifest delete, and the
	// occupancy probe that every kubernetes teardown now makes. nsList() is an
	// empty namespace, so the run reaches the namespace question.
	f := &queueRunner{resp: []queuedResp{{"", nil}, {"", nil}, {nsList(), nil}, {"", nil}}}
	// remove asks before it tears down; this test is about the argv that
	// follows, so it answers yes. The prompt itself is covered below.
	withPromptAnswer(t, "y")
	dir := t.TempDir()
	write(t, dir, "env.yaml", sharedEnv+kubeEnv)
	write(t, dir, "10.yaml", validWF)

	if code := dispatch([]string{"remove", "--platform", "kubernetes", "-e", filepath.Join(dir, "env.yaml")}, f); code != 0 {
		t.Fatalf("exit=%d", code)
	}
	// preflight, delete, the occupancy probe, and the namespace delete the
	// prompt approved.
	if len(f.calls) != 4 {
		t.Fatalf("want 4 calls (preflight, delete, probe, namespace delete), got %d: %+v", len(f.calls), f.calls)
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

// TestRemoveKubernetesSeamPromptsBeforeTearingDown is the confirmation half of
// TestRemoveKubernetesSeamHappyPath above: an accepted answer produces exactly
// the same two calls, so the prompt gates the teardown without changing it.
func TestRemoveKubernetesSeamPromptsBeforeTearingDown(t *testing.T) {
	for _, answer := range []string{"y", "yes", "YES", "Y"} {
		t.Run(answer, func(t *testing.T) {
			f := &queueRunner{resp: []queuedResp{{"", nil}, {"", nil}, {nsList(), nil}, {"", nil}}}
			withPromptAnswer(t, answer)
			dir := t.TempDir()
			write(t, dir, "env.yaml", sharedEnv+kubeEnv)
			write(t, dir, "10.yaml", validWF)

			if code := dispatch([]string{"remove", "--platform", "kubernetes", "-e", filepath.Join(dir, "env.yaml")}, f); code != 0 {
				t.Fatalf("exit=%d, want 0", code)
			}
			// The same answer approves the namespace question too, so all four
			// calls happen: preflight, delete, probe, namespace delete.
			if len(f.calls) != 4 {
				t.Fatalf("want 4 calls, got %d: %+v", len(f.calls), f.calls)
			}
		})
	}
}

// TestRemoveDeclinedTouchesNothing pins the whole point of the prompt: a
// declined teardown runs nothing at all, not even the read-only preflight probe,
// because the question is asked before it. A blank line declines -- the safe
// answer is the one Enter gives.
func TestRemoveDeclinedTouchesNothing(t *testing.T) {
	for _, answer := range []string{"n", "no", "", "  ", "anything else"} {
		t.Run("answer="+answer, func(t *testing.T) {
			f := useFakeRunner(t)
			withPromptAnswer(t, answer)
			dir := t.TempDir()
			write(t, dir, "env.yaml", sharedEnv+kubeEnv)
			write(t, dir, "10.yaml", validWF)

			var code int
			stderr := captureStderr(t, func() {
				code = dispatch([]string{"remove", "--platform", "kubernetes", "-e", filepath.Join(dir, "env.yaml")}, f)
			})
			if code != 0 {
				t.Fatalf("exit=%d, want 0 -- a teardown declined on purpose did not fail", code)
			}
			if len(f.calls) != 0 {
				t.Fatalf("nothing may run when the teardown is declined, got %d calls: %+v", len(f.calls), f.calls)
			}
			if !strings.Contains(stderr, "cancelled") {
				t.Errorf("a declined teardown should say so, got:\n%s", stderr)
			}
		})
	}
}

// TestRemoveNoPromptSkipsThePromptEntirely covers the scripted path: --no-prompt must not
// reach promptLine at all, rather than reaching it and ignoring the answer.
func TestRemoveNoPromptSkipsThePromptEntirely(t *testing.T) {
	f := &queueRunner{resp: []queuedResp{{"", nil}, {"", nil}, {nsList(), nil}, {"", nil}}}
	old := promptLine
	promptLine = func(string) (string, error) {
		t.Error("--no-prompt must not prompt at all")
		return "n", nil
	}
	t.Cleanup(func() { promptLine = old })

	dir := t.TempDir()
	write(t, dir, "env.yaml", sharedEnv+kubeEnv)
	write(t, dir, "10.yaml", validWF)

	if code := dispatch([]string{"remove", "--platform", "kubernetes", noPromptFlagName, "-e", filepath.Join(dir, "env.yaml")}, f); code != 0 {
		t.Fatalf("exit=%d, want 0", code)
	}
	// Neither question is asked, and the empty namespace goes with it:
	// preflight, delete, probe, namespace delete.
	if len(f.calls) != 4 {
		t.Fatalf("want 4 calls, got %d: %+v", len(f.calls), f.calls)
	}
}

// TestRemoveNonTTYFailsFastNamingTheFlag is the CI case: rather than blocking on
// a read that will never return, the prompt refuses and the error names the flag
// that skips it -- the same shape the platform menu and the status install
// confirmation use.
func TestRemoveNonTTYFailsFastNamingTheFlag(t *testing.T) {
	f := useFakeRunner(t)
	old := promptLine
	promptLine = func(string) (string, error) {
		return "", fmt.Errorf("stdin is not a terminal, so this prompt cannot be answered interactively")
	}
	t.Cleanup(func() { promptLine = old })

	dir := t.TempDir()
	write(t, dir, "env.yaml", sharedEnv+kubeEnv)
	write(t, dir, "10.yaml", validWF)

	var code int
	stderr := captureStderr(t, func() {
		code = dispatch([]string{"remove", "--platform", "kubernetes", "-e", filepath.Join(dir, "env.yaml")}, f)
	})
	if code != 1 {
		t.Fatalf("exit=%d, want 1", code)
	}
	if !strings.Contains(stderr, noPromptFlagName) {
		t.Errorf("the refusal must name %s, got:\n%s", noPromptFlagName, stderr)
	}
	if len(f.calls) != 0 {
		t.Fatalf("nothing may run, got %d calls: %+v", len(f.calls), f.calls)
	}
}

// TestRemovePromptNamesWhatItWillDestroy is why the prompt is worth having: the
// question has to carry the identifier that tells one target from another, or an
// operator confirms "kubernetes" and loses production. On kubernetes that is the
// namespace.
func TestRemovePromptNamesWhatItWillDestroy(t *testing.T) {
	quadlet := t.TempDir()
	for _, c := range []struct {
		platform string
		env      string
		want     []string
	}{
		{"kubernetes", kubeEnv, []string{"solmq-connector", "solace-connectors"}},
		{"docker", dockerEnv, []string{"container"}},
		{"podman", podmanEnv(quadlet), []string{"container", ".service"}},
	} {
		t.Run(c.platform, func(t *testing.T) {
			f := useFakeRunner(t)
			var asked string
			old := promptLine
			promptLine = func(q string) (string, error) { asked = q; return "n", nil }
			t.Cleanup(func() { promptLine = old })

			dir := t.TempDir()
			write(t, dir, "env.yaml", sharedEnv+c.env)
			write(t, dir, "10.yaml", validWF)

			captureStderr(t, func() {
				dispatch([]string{"remove", "--platform", c.platform, "-e", filepath.Join(dir, "env.yaml")}, f)
			})
			if !strings.Contains(asked, c.platform) {
				t.Errorf("the prompt should name the platform, got %q", asked)
			}
			for _, w := range c.want {
				if !strings.Contains(asked, w) {
					t.Errorf("the prompt should carry %q so the target is identifiable, got %q", w, asked)
				}
			}
		})
	}
}

// TestDeployNeverPromptsAndRejectsNoPrompt keeps the flag where it belongs. deploy is
// additive and re-runnable, so it has nothing to confirm; accepting --no-prompt there
// would be a flag that silently does nothing, which is what checkStatusFlags
// refuses elsewhere for the same reason.
func TestDeployNeverPromptsAndRejectsNoPrompt(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "env.yaml", sharedEnv+kubeEnv)
	write(t, dir, "10.yaml", validWF)

	f := useFakeRunner(t)
	old := promptLine
	promptLine = func(string) (string, error) {
		t.Error("deploy must never prompt")
		return "n", nil
	}
	t.Cleanup(func() { promptLine = old })
	if code := dispatch([]string{"deploy", "--platform", "kubernetes", "-e", filepath.Join(dir, "env.yaml")}, f); code != 0 {
		t.Fatalf("deploy exit=%d, want 0", code)
	}

	f2 := useFakeRunner(t)
	var code int
	captureStderr(t, func() {
		code = dispatch([]string{"deploy", "--platform", "kubernetes", noPromptFlagName, "-e", filepath.Join(dir, "env.yaml")}, f2)
	})
	if code != 2 {
		t.Fatalf("deploy %s should be an unknown flag (exit 2), got %d", noPromptFlagName, code)
	}
	if len(f2.calls) != 0 {
		t.Fatalf("nothing may run, got %d calls", len(f2.calls))
	}
}

// TestRemoveTargetDescriptions covers removeTarget directly, including the
// branches the end-to-end tests cannot reach: a section-less env (guards, since
// resolvePlatform demands the section for deploy/remove) and a namespace-less
// deployment (which validate rejects moments later, so the prompt is the only
// place it can surface).
func TestRemoveTargetDescriptions(t *testing.T) {
	kube := func(name, ns string) *spec.Env {
		e := &spec.Env{Kubernetes: &spec.Kubernetes{}}
		e.Kubernetes.Deployment.Name = name
		e.Kubernetes.Deployment.Namespace = ns
		return e
	}
	for _, c := range []struct {
		name     string
		platform string
		env      *spec.Env
		want     string
	}{
		{"kube with namespace", "kubernetes", kube("solmq", "prod"), "deployment solmq in namespace prod"},
		{"kube without namespace", "kubernetes", kube("solmq", ""), "deployment solmq in the current namespace"},
		{"kube section missing", "kubernetes", &spec.Env{}, "the kubernetes deployment"},
		{"kube env nil", "kubernetes", nil, "the kubernetes deployment"},
		{"docker", "docker", &spec.Env{Docker: &spec.Docker{Name: "solmq-conn"}}, "container solmq-conn"},
		{"docker section missing", "docker", &spec.Env{}, "the docker container"},
		{"podman", "podman", &spec.Env{Podman: &spec.Podman{Name: "solmq-conn"}}, "container solmq-conn and its systemd unit solmq-conn.service"},
		{"podman section missing", "podman", &spec.Env{}, "the podman container"},
	} {
		t.Run(c.name, func(t *testing.T) {
			if got := removeTarget(c.platform, c.env); got != c.want {
				t.Errorf("removeTarget(%q) = %q, want %q", c.platform, got, c.want)
			}
		})
	}
}

// ---- the namespace half of a kubernetes teardown ----

// nsList builds the document `kubectl get all,...` returns, so the occupancy
// cases read as the objects they describe rather than as JSON.
func nsList(items ...string) string {
	return `{"apiVersion":"v1","kind":"List","items":[` + strings.Join(items, ",") + `]}`
}

func nsItemJSON(kind, name string) string {
	return `{"kind":"` + kind + `","metadata":{"name":"` + name + `"}}`
}

// TestRemoveTeardownManifestOmitsTheNamespace is the bug this whole file exists
// for: the manifest piped to `kubectl delete -f -` used to carry a Namespace,
// which cascades to every object inside it -- including workloads this tool
// never deployed. The namespace is a separate, checked step now; the manifest
// must never carry it.
func TestRemoveTeardownManifestOmitsTheNamespace(t *testing.T) {
	f := &queueRunner{resp: []queuedResp{{"", nil}, {"", nil}, {nsList(), nil}, {"", nil}}}
	withPromptAnswer(t, "y")
	dir := t.TempDir()
	write(t, dir, "env.yaml", sharedEnv+kubeEnv)
	write(t, dir, "10.yaml", validWF)

	if code := dispatch([]string{"remove", "--platform", "kubernetes", "-e", filepath.Join(dir, "env.yaml")}, f); code != 0 {
		t.Fatalf("exit=%d, want 0", code)
	}
	if strings.Contains(f.calls[1].stdin, "kind: Namespace") {
		t.Fatalf("a teardown manifest must not carry the namespace:\n%s", f.calls[1].stdin)
	}
	// The separate namespace delete is the ONLY place a Namespace document is
	// piped for a teardown, and only after the occupancy check passed.
	if !strings.Contains(f.calls[3].stdin, "kind: Namespace") {
		t.Errorf("the namespace step should pipe the Namespace document, got %q", f.calls[3].stdin)
	}
	// deploy still needs it, or the objects have nowhere to land.
	f2 := useFakeRunner(t)
	if code := dispatch([]string{"deploy", "--platform", "kubernetes", "-e", filepath.Join(dir, "env.yaml")}, f2); code != 0 {
		t.Fatalf("deploy exit=%d", code)
	}
	if !strings.Contains(f2.calls[1].stdin, "kind: Namespace") {
		t.Errorf("deploy must still create the namespace:\n%s", f2.calls[1].stdin)
	}
}

// TestRemoveNamespaceOccupiedLeavesItAlone is the case that motivated the flag:
// another connector, or anything else, living in the same namespace.
func TestRemoveNamespaceOccupiedLeavesItAlone(t *testing.T) {
	q := &queueRunner{resp: []queuedResp{
		{"", nil}, // preflight
		{"", nil}, // delete
		{nsList(
			nsItemJSON("Deployment", "billing-api"),
			nsItemJSON("StatefulSet", "postgres"),
			nsItemJSON("PersistentVolumeClaim", "billing-data"),
		), nil},
	}}
	withPromptAnswer(t, "y")
	dir := t.TempDir()
	write(t, dir, "env.yaml", sharedEnv+kubeEnv)
	write(t, dir, "10.yaml", validWF)

	var code int
	stderr := captureStderr(t, func() {
		code = dispatch([]string{"remove", "--platform", "kubernetes", "-e", filepath.Join(dir, "env.yaml")}, q)
	})
	if code != 0 {
		t.Fatalf("exit=%d, want 0 -- the teardown succeeded", code)
	}
	// preflight, delete, the occupancy probe -- and no fourth call.
	if len(q.calls) != 3 {
		t.Fatalf("the namespace must not be deleted, got %d calls: %+v", len(q.calls), q.calls)
	}
	for _, want := range []string{"deployment/billing-api", "statefulset/postgres", "persistentvolumeclaim/billing-data", "leaving namespace"} {
		if !strings.Contains(stderr, want) {
			t.Errorf("stderr should carry %q, got:\n%s", want, stderr)
		}
	}
}

// TestRemoveNamespaceEmptyPromptsSeparately covers the second question. Saying
// yes to tearing down a deployment is not saying yes to removing the namespace
// around it, so it is asked on its own.
func TestRemoveNamespaceEmptyPromptsSeparately(t *testing.T) {
	for _, c := range []struct {
		answer  string
		deleted bool
	}{
		{"y", true},
		{"n", false},
		{"", false},
	} {
		t.Run("answer="+c.answer, func(t *testing.T) {
			q := &queueRunner{resp: []queuedResp{{"", nil}, {"", nil}, {nsList(), nil}, {"", nil}}}
			// Two questions: yes to the teardown, then the case's answer to the
			// namespace. That separation is the point of the feature.
			withPromptAnswers(t, "y", c.answer)
			dir := t.TempDir()
			write(t, dir, "env.yaml", sharedEnv+kubeEnv)
			write(t, dir, "10.yaml", validWF)

			if code := dispatch([]string{"remove", "--platform", "kubernetes", "-e", filepath.Join(dir, "env.yaml")}, q); code != 0 {
				t.Fatalf("exit=%d, want 0", code)
			}
			want := 3
			if c.deleted {
				want = 4
			}
			if len(q.calls) != want {
				t.Fatalf("want %d calls, got %d: %+v", want, len(q.calls), q.calls)
			}
			if c.deleted && !strings.Contains(q.calls[3].stdin, "kind: Namespace") {
				t.Errorf("the namespace delete should pipe the Namespace document, got %q", q.calls[3].stdin)
			}
		})
	}
}

// TestRemoveNoPromptNeverRemovesAnOccupiedNamespace is the invariant, and the
// case that must never regress: no flag removes a namespace that still holds
// someone else's work.
//
// --no-prompt approves the questions a teardown asks. It does not approve a
// cascade -- the occupancy check runs first and an occupant of any kind ends it,
// silently or not.
func TestRemoveNoPromptNeverRemovesAnOccupiedNamespace(t *testing.T) {
	q := &queueRunner{resp: []queuedResp{
		{"", nil}, // preflight
		{"", nil}, // delete
		{nsList(nsItemJSON("Deployment", "billing-api")), nil},
	}}
	old := promptLine
	promptLine = func(string) (string, error) {
		t.Error("--no-prompt must not prompt at all")
		return "n", nil
	}
	t.Cleanup(func() { promptLine = old })

	dir := t.TempDir()
	write(t, dir, "env.yaml", sharedEnv+kubeEnv)
	write(t, dir, "10.yaml", validWF)

	var code int
	stderr := captureStderr(t, func() {
		code = dispatch([]string{"remove", "--platform", "kubernetes", noPromptFlagName, "-e", filepath.Join(dir, "env.yaml")}, q)
	})
	if code != 0 {
		t.Fatalf("exit=%d, want 0 -- the teardown succeeded", code)
	}
	// preflight, delete, probe -- and no fourth call, whatever the flags say.
	if len(q.calls) != 3 {
		t.Fatalf("an occupied namespace must never be deleted, got %d calls: %+v", len(q.calls), q.calls)
	}
	for _, want := range []string{"deployment/billing-api", "leaving namespace"} {
		if !strings.Contains(stderr, want) {
			t.Errorf("stderr should carry %q, got:\n%s", want, stderr)
		}
	}
}

// TestRemovePlainRunChecksTheNamespace is the regression this change exists for:
// a plain `remove` used to skip the namespace step entirely, so nothing was
// checked and nothing was offered. It is part of what remove does now, not
// something to opt into.
func TestRemovePlainRunChecksTheNamespace(t *testing.T) {
	q := &queueRunner{resp: []queuedResp{{"", nil}, {"", nil}, {nsList(), nil}, {"", nil}}}
	var asked []string
	old := promptLine
	promptLine = func(question string) (string, error) { asked = append(asked, question); return "y", nil }
	t.Cleanup(func() { promptLine = old })

	dir := t.TempDir()
	write(t, dir, "env.yaml", sharedEnv+kubeEnv)
	write(t, dir, "10.yaml", validWF)

	if code := dispatch([]string{"remove", "--platform", "kubernetes", "-e", filepath.Join(dir, "env.yaml")}, q); code != 0 {
		t.Fatalf("exit=%d, want 0", code)
	}
	// Two questions with no flags at all: the teardown, then the namespace.
	if len(asked) != 2 {
		t.Fatalf("want 2 prompts (teardown, namespace), got %d: %v", len(asked), asked)
	}
	if !strings.Contains(asked[1], "namespace") {
		t.Errorf("the second question should be about the namespace, got %q", asked[1])
	}
	if len(q.calls) != 4 {
		t.Fatalf("want 4 calls (preflight, delete, probe, namespace delete), got %d: %+v", len(q.calls), q.calls)
	}
	if !containsToken(q.calls[2].argv, namespaceProbeTypes) {
		t.Errorf("call 2 should be the occupancy probe, got %v", q.calls[2].argv)
	}
}

// TestRemoveNamespaceRefusesClusterNamespaces pins the guard that no occupancy
// result can override. Deleting one of these breaks the cluster rather than an
// application, and a wrong namespace in env.yaml is the likeliest way to get here.
func TestRemoveNamespaceRefusesClusterNamespaces(t *testing.T) {
	for _, ns := range []string{"default", "kube-system", "kube-public", "kube-node-lease"} {
		t.Run(ns, func(t *testing.T) {
			env := strings.Replace(kubeEnv, "namespace: solace-connectors", "namespace: "+ns, 1)
			dir := t.TempDir()
			write(t, dir, "env.yaml", sharedEnv+env)
			write(t, dir, "10.yaml", validWF)

			q := &queueRunner{resp: []queuedResp{{"", nil}, {"", nil}}}
			var code int
			stderr := captureStderr(t, func() {
				code = dispatch([]string{"remove", "--platform", "kubernetes", noPromptFlagName, "-e", filepath.Join(dir, "env.yaml")}, q)
			})
			if code != 0 {
				t.Fatalf("exit=%d, want 0", code)
			}
			// Refused before the occupancy probe: preflight and delete only.
			if len(q.calls) != 2 {
				t.Fatalf("%s must be refused without even being probed, got %d calls: %+v", ns, len(q.calls), q.calls)
			}
			if !strings.Contains(stderr, "never removed") {
				t.Errorf("the refusal should say so, got:\n%s", stderr)
			}
		})
	}
}

// TestRemoveNamespaceProbeFailureLeavesItAlone: a namespace is not worth
// deleting on a guess.
func TestRemoveNamespaceProbeFailureLeavesItAlone(t *testing.T) {
	q := &queueRunner{resp: []queuedResp{{"", nil}, {"", nil}, {"boom", fmt.Errorf("forbidden")}}}
	dir := t.TempDir()
	write(t, dir, "env.yaml", sharedEnv+kubeEnv)
	write(t, dir, "10.yaml", validWF)

	var code int
	stderr := captureStderr(t, func() {
		code = dispatch([]string{"remove", "--platform", "kubernetes", noPromptFlagName, "-e", filepath.Join(dir, "env.yaml")}, q)
	})
	if code != 0 {
		t.Fatalf("exit=%d, want 0 -- the teardown itself succeeded", code)
	}
	if len(q.calls) != 3 {
		t.Fatalf("the namespace must not be deleted after a failed probe, got %d calls", len(q.calls))
	}
	if !strings.Contains(stderr, "left in place") {
		t.Errorf("stderr should say the namespace was left alone, got:\n%s", stderr)
	}
}

// TestRemoveFailedTeardownSkipsTheNamespace: whatever did not come down is still
// in there, so the namespace is not even considered.
func TestRemoveFailedTeardownSkipsTheNamespace(t *testing.T) {
	q := &queueRunner{resp: []queuedResp{{"", nil}, {"boom", fmt.Errorf("denied")}}}
	dir := t.TempDir()
	write(t, dir, "env.yaml", sharedEnv+kubeEnv)
	write(t, dir, "10.yaml", validWF)

	var code int
	captureStderr(t, func() {
		code = dispatch([]string{"remove", "--platform", "kubernetes", noPromptFlagName, "-e", filepath.Join(dir, "env.yaml")}, q)
	})
	if code != 1 {
		t.Fatalf("exit=%d, want 1", code)
	}
	if len(q.calls) != 2 {
		t.Fatalf("a failed teardown must not probe or delete the namespace, got %d calls: %+v", len(q.calls), q.calls)
	}
}

// TestNamespaceOccupantsRules is the exclusion table. Each non-occupant is here
// because it would otherwise keep a namespace alive forever: our own leftovers,
// the objects kubernetes puts in every namespace, and -- the likeliest of all --
// the release that was deleted moments ago and is still terminating.
func TestNamespaceOccupantsRules(t *testing.T) {
	owned := map[string]bool{"solmq-connector": true, "solmq-connector-config": true}

	for _, c := range []struct {
		name string
		item string
		want bool // is it an occupant
	}{
		{"a foreign deployment", nsItemJSON("Deployment", "billing-api"), true},
		{"a foreign statefulset", nsItemJSON("StatefulSet", "postgres"), true},
		{"a foreign pvc", nsItemJSON("PersistentVolumeClaim", "data"), true},
		{"a foreign service", nsItemJSON("Service", "billing-svc"), true},
		{"our deployment", nsItemJSON("Deployment", "solmq-connector"), false},
		{"our configmap", nsItemJSON("ConfigMap", "solmq-connector-config"), false},
		{
			name: "a pod we own via ownerReferences",
			item: `{"kind":"Pod","metadata":{"name":"solmq-connector-7d4-abc","ownerReferences":[{"kind":"ReplicaSet","name":"solmq-connector"}]}}`,
			want: false,
		},
		{
			name: "anything already terminating",
			item: `{"kind":"Deployment","metadata":{"name":"whatever","deletionTimestamp":"2026-08-31T10:00:00Z"}}`,
			want: false,
		},
		{"the default root CA configmap", nsItemJSON("ConfigMap", "kube-root-ca.crt"), false},
		{"the default service account", nsItemJSON("ServiceAccount", "default"), false},
		{
			name: "a service-account token secret",
			item: `{"kind":"Secret","type":"kubernetes.io/service-account-token","metadata":{"name":"default-token-abc"}}`,
			want: false,
		},
	} {
		t.Run(c.name, func(t *testing.T) {
			got, err := namespaceOccupants(nsList(c.item), owned)
			if err != nil {
				t.Fatal(err)
			}
			if (len(got) > 0) != c.want {
				t.Errorf("occupant=%v, want %v (got %v)", len(got) > 0, c.want, got)
			}
		})
	}
}

// TestNamespaceOccupantsAreSortedAndLabelled pins the output shape: kind/name,
// lower-cased and sorted, so the list an operator reads is stable between runs.
func TestNamespaceOccupantsAreSortedAndLabelled(t *testing.T) {
	got, err := namespaceOccupants(nsList(
		nsItemJSON("StatefulSet", "postgres"),
		nsItemJSON("Deployment", "billing-api"),
	), nil)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"deployment/billing-api", "statefulset/postgres"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("= %v, want %v", got, want)
	}
}

// TestNamespaceOccupantsRejectsUnreadableOutput: a parse failure must be an
// error, never an empty list, or unreadable output would read as "empty" and
// authorise the delete.
func TestNamespaceOccupantsRejectsUnreadableOutput(t *testing.T) {
	if _, err := namespaceOccupants("not json", nil); err == nil {
		t.Fatal("unreadable output must not be treated as an empty namespace")
	}
}

// TestOwnedNamesCoversEverythingThisReleaseCreates is the safety net behind the
// occupancy check. The check runs after the teardown, so these objects are
// normally gone -- but one that has not finished terminating and carries no
// deletionTimestamp yet must not be mistaken for someone else's workload and
// keep the namespace alive forever.
func TestOwnedNamesCoversEverythingThisReleaseCreates(t *testing.T) {
	k := &spec.Kubernetes{
		Secrets: spec.Secrets{
			Credentials: &spec.CredentialsSecret{Create: &spec.CredCreate{Name: "solmq-creds"}},
			Stores:      &spec.StoresSecret{Create: &spec.StoreCreate{Name: "solmq-tls"}},
			ImagePull:   &spec.ImagePullSecret{Name: "solmq-pull", Create: true},
		},
		Libs: &spec.Libs{PVC: &spec.LibsPVC{Create: &spec.PVCCreate{Name: "solmq-libs"}}},
	}
	k.Deployment.Name = "solmq-connector"

	owned := ownedNames(k)
	for _, want := range []string{"solmq-connector", "solmq-connector-config", "solmq-creds", "solmq-tls", "solmq-pull", "solmq-libs"} {
		if !owned[want] {
			t.Errorf("ownedNames should carry %q, got %v", want, owned)
		}
	}
	if owned[""] {
		t.Error("an empty name must never be owned, or every unnamed object counts as ours")
	}

	// Referenced-but-not-created objects are NOT ours: the operator manages
	// them, the teardown does not delete them, and their presence is a real
	// reason to keep the namespace.
	k2 := &spec.Kubernetes{
		Secrets: spec.Secrets{
			Credentials: &spec.CredentialsSecret{Existing: "their-creds"},
			Stores:      &spec.StoresSecret{Existing: "their-tls"},
			ImagePull:   &spec.ImagePullSecret{Name: "their-pull"},
		},
		Libs: &spec.Libs{PVC: &spec.LibsPVC{Existing: "their-pvc"}},
	}
	k2.Deployment.Name = "solmq-connector"
	owned2 := ownedNames(k2)
	for _, no := range []string{"their-creds", "their-tls", "their-pull", "their-pvc"} {
		if owned2[no] {
			t.Errorf("%q is referenced, not created, so it is not ours to discount", no)
		}
	}

	if got := ownedNames(nil); len(got) != 0 {
		t.Errorf("a nil section owns nothing, got %v", got)
	}
}

// TestRemoveNamespaceWithoutANamespaceSaysSo covers the guard for a kubernetes
// section carrying no namespace: there is nothing to remove, and that is worth
// saying rather than deleting something named "".
func TestRemoveNamespaceWithoutANamespaceSaysSo(t *testing.T) {
	f := useFakeRunner(t)
	k := &spec.Kubernetes{Command: spec.DefaultKubeCommand}
	k.Deployment.Name = "solmq-connector"

	var code int
	stderr := captureStderr(t, func() {
		code = removeNamespace(f, actionOpts{action: "remove"}, k)
	})
	if code != 0 {
		t.Fatalf("exit=%d, want 0", code)
	}
	if !strings.Contains(stderr, "no namespace") {
		t.Errorf("stderr should say there is no namespace, got:\n%s", stderr)
	}
	if len(f.calls) != 0 {
		t.Errorf("nothing may run, got %d calls", len(f.calls))
	}
}

// TestRemoveNamespaceRejectsAnUnsafeCommand pins that the namespace delete goes
// through the same binary allowlist as everything else, rather than being a
// second path around it.
func TestRemoveNamespaceRejectsAnUnsafeCommand(t *testing.T) {
	f := useFakeRunner(t)
	k := &spec.Kubernetes{Command: "curl"}
	k.Deployment.Name = "solmq-connector"
	k.Deployment.Namespace = "prod"

	var code int
	stderr := captureStderr(t, func() {
		code = removeNamespace(f, actionOpts{action: "remove"}, k)
	})
	if code != 1 {
		t.Fatalf("exit=%d, want 1", code)
	}
	if !strings.Contains(stderr, "curl") {
		t.Errorf("the error should name the refused binary, got:\n%s", stderr)
	}
	if len(f.calls) != 0 {
		t.Errorf("nothing may run, got %d calls", len(f.calls))
	}
}

// TestRemoveNamespaceUnreadableProbeLeavesItAlone: output that cannot be parsed
// must not read as "empty" and authorise the delete.
func TestRemoveNamespaceUnreadableProbeLeavesItAlone(t *testing.T) {
	q := &queueRunner{resp: []queuedResp{{"not json at all", nil}}}
	k := &spec.Kubernetes{Command: spec.DefaultKubeCommand}
	k.Deployment.Name = "solmq-connector"
	k.Deployment.Namespace = "prod"

	var code int
	stderr := captureStderr(t, func() {
		code = removeNamespace(q, actionOpts{action: "remove", noPrompt: true}, k)
	})
	if code != 0 {
		t.Fatalf("exit=%d, want 0", code)
	}
	if len(q.calls) != 1 {
		t.Fatalf("the namespace must not be deleted, got %d calls: %+v", len(q.calls), q.calls)
	}
	if !strings.Contains(stderr, "left in place") {
		t.Errorf("stderr should say the namespace was left alone, got:\n%s", stderr)
	}
}

// TestRemoveNamespaceNonTTYFailsFastNamingTheFlag: the namespace question
// refuses a non-TTY the same way every other prompt does, naming the flag that
// skips it rather than blocking on a read that will never return.
func TestRemoveNamespaceNonTTYFailsFastNamingTheFlag(t *testing.T) {
	q := &queueRunner{resp: []queuedResp{{nsList(), nil}}}
	old := promptLine
	promptLine = func(string) (string, error) {
		return "", fmt.Errorf("stdin is not a terminal, so this prompt cannot be answered interactively")
	}
	t.Cleanup(func() { promptLine = old })

	k := &spec.Kubernetes{Command: spec.DefaultKubeCommand}
	k.Deployment.Name = "solmq-connector"
	k.Deployment.Namespace = "prod"

	var code int
	stderr := captureStderr(t, func() {
		code = removeNamespace(q, actionOpts{action: "remove"}, k)
	})
	if code != 1 {
		t.Fatalf("exit=%d, want 1", code)
	}
	if !strings.Contains(stderr, noPromptFlagName) {
		t.Errorf("the refusal must name %s, got:\n%s", noPromptFlagName, stderr)
	}
	if len(q.calls) != 1 {
		t.Errorf("the namespace must not be deleted, got %d calls", len(q.calls))
	}
}

// TestRemoveNamespaceProbeArgvIsOneQuery pins the shape of the occupancy query:
// one call listing every kind whose loss would matter, in the namespace being
// considered.
func TestRemoveNamespaceProbeArgvIsOneQuery(t *testing.T) {
	q := &queueRunner{resp: []queuedResp{{nsList(nsItemJSON("Deployment", "other")), nil}}}
	k := &spec.Kubernetes{Command: spec.DefaultKubeCommand}
	k.Deployment.Name = "solmq-connector"
	k.Deployment.Namespace = "prod"

	captureStderr(t, func() { removeNamespace(q, actionOpts{action: "remove", noPrompt: true}, k) })
	if len(q.calls) != 1 {
		t.Fatalf("want 1 probe call, got %d", len(q.calls))
	}
	want := []string{"kubectl", "get", namespaceProbeTypes, "-n", "prod", "-o", "json"}
	if !reflect.DeepEqual(q.calls[0].argv, want) {
		t.Errorf("probe argv = %v, want %v", q.calls[0].argv, want)
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

// ---- j) download subcommand -------------------------------------------------------

// fakeDownload records every Input reaching the injectable downloadFn seam
// (see main.go) so tests can assert exactly what runDownload built, without
// ever touching the network or internal/libs's own Maven/HTTP behavior --
// mirrors fakeRunner's role for the deploy/remove seam. resp/err are returned
// to every call.
type fakeDownload struct {
	calls []libs.Input
	resp  libs.Report
	err   error
}

func (f *fakeDownload) call(in libs.Input) (libs.Report, error) {
	f.calls = append(f.calls, in)
	return f.resp, f.err
}

// useFakeDownload installs f as the downloadFn seam for the duration of the
// test, restoring the production libs.Download on cleanup.
func useFakeDownload(t *testing.T, f *fakeDownload) {
	t.Helper()
	old := downloadFn
	downloadFn = f.call
	t.Cleanup(func() { downloadFn = old })
}

// TestDownloadMissingAndUnknownWordsRejected covers all three command-level
// validations runDownload enforces before ever calling downloadFn: a missing
// or unknown target, and a missing or unknown set -- mirroring
// runAutoComplete's messages and exit codes for the same shape of error.
func TestDownloadMissingAndUnknownWordsRejected(t *testing.T) {
	cases := []struct {
		name string
		args []string
		want string
	}{
		{"missing target", []string{"download"}, "missing target"},
		{"unknown target", []string{"download", "bogus"}, `unknown target "bogus"`},
		{"missing set", []string{"download", "jar"}, "missing set"},
		{"unknown set", []string{"download", "jar", "bogus"}, `unknown set "bogus"`},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			f := &fakeDownload{}
			useFakeDownload(t, f)
			r := &fakeRunner{}
			var code int
			stderr := captureStderr(t, func() { code = dispatch(c.args, r) })
			if code != 2 {
				t.Errorf("dispatch(%v) = %d, want 2", c.args, code)
			}
			if !strings.Contains(stderr, c.want) {
				t.Errorf("dispatch(%v) stderr = %q, want substring %q", c.args, stderr, c.want)
			}
			if len(f.calls) != 0 {
				t.Errorf("dispatch(%v): downloadFn must not be invoked on a rejection, got %d calls", c.args, len(f.calls))
			}
			if len(r.calls) != 0 {
				t.Errorf("dispatch(%v): runner must not be invoked, got %d calls", c.args, len(r.calls))
			}
		})
	}
}

// TestDownloadDirDefaultAndPositionalOverride covers the trailing [dir]
// positional: "./libs" when omitted, and the given value when present.
func TestDownloadDirDefaultAndPositionalOverride(t *testing.T) {
	cases := []struct {
		name string
		args []string
		want string
	}{
		{"default dir", []string{"download", "jar", "mq"}, "./libs"},
		{"explicit dir", []string{"download", "jar", "mq", "mylibs"}, "mylibs"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			f := &fakeDownload{}
			useFakeDownload(t, f)
			if code := dispatch(c.args, &fakeRunner{}); code != 0 {
				t.Fatalf("dispatch(%v) = %d, want 0", c.args, code)
			}
			if len(f.calls) != 1 {
				t.Fatalf("downloadFn calls = %d, want 1", len(f.calls))
			}
			if f.calls[0].Dir != c.want {
				t.Errorf("Input.Dir = %q, want %q", f.calls[0].Dir, c.want)
			}
		})
	}
}

// TestDownloadSetReachesInput covers both modeled sets threading through to
// libs.Input.Set unchanged.
func TestDownloadSetReachesInput(t *testing.T) {
	cases := []struct {
		name string
		set  string
		want string
	}{
		{"mq", "mq", libs.SetMQ},
		{"syslog", "syslog", libs.SetSyslog},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			f := &fakeDownload{}
			useFakeDownload(t, f)
			args := []string{"download", "jar", c.set}
			if code := dispatch(args, &fakeRunner{}); code != 0 {
				t.Fatalf("dispatch(%v) = %d, want 0", args, code)
			}
			if f.calls[0].Set != c.want {
				t.Errorf("Input.Set = %q, want %q", f.calls[0].Set, c.want)
			}
		})
	}
}

// TestDownloadURLFlagRepeatable covers --url collecting every occurrence, in
// order, into libs.Input.URLs.
func TestDownloadURLFlagRepeatable(t *testing.T) {
	f := &fakeDownload{}
	useFakeDownload(t, f)
	args := []string{"download", "jar", "mq", "--url", "https://example.com/a.jar", "--url", "https://example.com/b.jar"}
	if code := dispatch(args, &fakeRunner{}); code != 0 {
		t.Fatalf("dispatch(%v) = %d, want 0", args, code)
	}
	want := []string{"https://example.com/a.jar", "https://example.com/b.jar"}
	if !reflect.DeepEqual(f.calls[0].URLs, want) {
		t.Errorf("Input.URLs = %v, want %v", f.calls[0].URLs, want)
	}
}

// TestDownloadForceFlagReachesInput covers -f/--force defaulting to false and
// both spellings reaching libs.Input.Force.
func TestDownloadForceFlagReachesInput(t *testing.T) {
	cases := []struct {
		name string
		args []string
		want bool
	}{
		{"default false", []string{"download", "jar", "syslog"}, false},
		{"-f short", []string{"download", "jar", "syslog", "-f"}, true},
		{"--force long", []string{"download", "jar", "syslog", "--force"}, true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			f := &fakeDownload{}
			useFakeDownload(t, f)
			if code := dispatch(c.args, &fakeRunner{}); code != 0 {
				t.Fatalf("dispatch(%v) = %d, want 0", c.args, code)
			}
			if f.calls[0].Force != c.want {
				t.Errorf("Input.Force = %v, want %v", f.calls[0].Force, c.want)
			}
		})
	}
}

// TestDownloadVersionFlagReachesInput covers --version defaulting to "" (the
// libs.Input contract for "latest stable") and an explicit value reaching
// libs.Input.Version unchanged -- pinning the SEED release, per task 1.
func TestDownloadVersionFlagReachesInput(t *testing.T) {
	cases := []struct {
		name string
		args []string
		want string
	}{
		{"default empty", []string{"download", "jar", "mq"}, ""},
		{"explicit pin", []string{"download", "jar", "mq", "--version", "9.4.2.0"}, "9.4.2.0"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			f := &fakeDownload{}
			useFakeDownload(t, f)
			if code := dispatch(c.args, &fakeRunner{}); code != 0 {
				t.Fatalf("dispatch(%v) = %d, want 0", c.args, code)
			}
			if f.calls[0].Version != c.want {
				t.Errorf("Input.Version = %q, want %q", f.calls[0].Version, c.want)
			}
		})
	}
}

// TestDownloadOmitLibFileFlagReachesInput covers --omit-lib-file defaulting to
// "" (the libs.Input contract for the embedded jar list) and an explicit path
// reaching libs.Input.OmitLibFile unchanged.
func TestDownloadOmitLibFileFlagReachesInput(t *testing.T) {
	cases := []struct {
		name string
		args []string
		want string
	}{
		{"default empty", []string{"download", "jar", "mq"}, ""},
		{"explicit path", []string{"download", "jar", "mq", "--omit-lib-file", "other-image-libs.txt"}, "other-image-libs.txt"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			f := &fakeDownload{}
			useFakeDownload(t, f)
			if code := dispatch(c.args, &fakeRunner{}); code != 0 {
				t.Fatalf("dispatch(%v) = %d, want 0", c.args, code)
			}
			if f.calls[0].OmitLibFile != c.want {
				t.Errorf("Input.OmitLibFile = %q, want %q", f.calls[0].OmitLibFile, c.want)
			}
		})
	}
}

// TestDownloadIncludeProvidedFlagReachesInput covers --include-provided
// defaulting to false and reaching libs.Input.IncludeProvided true when given.
func TestDownloadIncludeProvidedFlagReachesInput(t *testing.T) {
	cases := []struct {
		name string
		args []string
		want bool
	}{
		{"default false", []string{"download", "jar", "syslog"}, false},
		{"--include-provided", []string{"download", "jar", "syslog", "--include-provided"}, true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			f := &fakeDownload{}
			useFakeDownload(t, f)
			if code := dispatch(c.args, &fakeRunner{}); code != 0 {
				t.Fatalf("dispatch(%v) = %d, want 0", c.args, code)
			}
			if f.calls[0].IncludeProvided != c.want {
				t.Errorf("Input.IncludeProvided = %v, want %v", f.calls[0].IncludeProvided, c.want)
			}
		})
	}
}

// TestDownloadReportExitCode covers reportDownload's exit contract: a
// systemic error (downloadFn's own error return) and a non-empty
// Report.Failed both exit non-zero; a clean write, a skip-only run, and a
// run where everything was omitted (the image already has all of it -- see
// task 2's version-aware omission) all exit 0.
//
// Exit-code convention (the finding's decision): a non-empty Report.Failed
// always exits 1, whether it holds one failure among several successes or
// every requested artifact -- the two "per-artifact failure" cases below
// (one of two written, and all failed) both exit 1 identically. This tool
// does not mint a distinct code for "partial" vs "total" failure, matching
// the documented exit codes (0 success, 1 processing error, 2 usage error)
// in docs/commands.md; a caller that must tell partial from total reads the
// written/skipped/omitted/failed counts reportDownload prints to stderr, not
// the exit code (see reportDownload's doc comment).
func TestDownloadReportExitCode(t *testing.T) {
	cases := []struct {
		name string
		rep  libs.Report
		err  error
		want int
	}{
		{"clean write", libs.Report{Written: []string{"libs/a.jar"}}, nil, 0},
		{"skip only", libs.Report{Skipped: []string{"libs/a.jar"}}, nil, 0},
		{"omitted only", libs.Report{Omitted: []string{"bcprov-jdk18on-1.84.jar: the image has 1.84"}}, nil, 0},
		{"partial failure", libs.Report{Written: []string{"libs/a.jar"}, Failed: []libs.Failure{{Name: "b.jar", Err: fmt.Errorf("404")}}}, nil, 1},
		{"total failure", libs.Report{Failed: []libs.Failure{{Name: "a.jar", Err: fmt.Errorf("404")}, {Name: "b.jar", Err: fmt.Errorf("404")}}}, nil, 1},
		{"systemic error", libs.Report{}, fmt.Errorf("boom"), 1},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			f := &fakeDownload{resp: c.rep, err: c.err}
			useFakeDownload(t, f)
			var code int
			stderr := captureStderr(t, func() { code = dispatch([]string{"download", "jar", "mq"}, &fakeRunner{}) })
			if code != c.want {
				t.Errorf("dispatch = %d, want %d\nstderr:\n%s", code, c.want, stderr)
			}
		})
	}
}

// TestDownloadReportPrintsWrittenSkippedFailedAndFallback pins
// reportDownload's line shapes, mirroring runExamples' "wrote:"/"exists (use
// -f to overwrite):" convention plus a "failed:" line and a Fallback note
// clearly labelled as a guessed version.
func TestDownloadReportPrintsWrittenSkippedFailedAndFallback(t *testing.T) {
	f := &fakeDownload{resp: libs.Report{
		Written:  []string{"libs/a.jar"},
		Skipped:  []string{"libs/b.jar"},
		Failed:   []libs.Failure{{Name: "c.jar", Err: fmt.Errorf("404")}},
		Fallback: []string{"c.jar: version could not be resolved from the POM chain, used latest stable"},
	}}
	useFakeDownload(t, f)
	var code int
	stderr := captureStderr(t, func() { code = dispatch([]string{"download", "jar", "mq"}, &fakeRunner{}) })
	if code != 1 {
		t.Errorf("dispatch = %d, want 1 (Report.Failed non-empty)", code)
	}
	for _, want := range []string{
		"wrote: libs/a.jar",
		"exists (use -f to overwrite): libs/b.jar",
		"failed: c.jar: 404",
		"guessed version:",
	} {
		if !strings.Contains(stderr, want) {
			t.Errorf("stderr missing %q, got:\n%s", want, stderr)
		}
	}
}

// TestDownloadReportPrintsOmittedBlockDistinctFromFallback covers the
// Report.Omitted, Report.Fallback, and Report.Unverified blocks: each is
// printed under its own prefix, so an operator can see at a glance that an
// artifact was skipped because the connector image already provides it
// ("omitted:"), had its version guessed ("guessed version:"), or could not be
// verified against a digest ("unverified:") -- never confusing any of those
// three with each other or with a "failed:" (something went wrong) line.
// Exit is 0: none of the three is a failure.
func TestDownloadReportPrintsOmittedBlockDistinctFromFallback(t *testing.T) {
	f := &fakeDownload{resp: libs.Report{
		Written:    []string{"libs/com.ibm.mq.jakarta.client-10.0.0.0.jar"},
		Omitted:    []string{"bcprov-jdk18on-1.84.jar: the image has 1.84", "jakarta.jms-api-3.0.0.jar: the image has 3.1.0"},
		Fallback:   []string{"logstash-logback-encoder-9.0.jar: version could not be resolved from the POM chain, used latest stable"},
		Unverified: []string{"custom.jar: no .sha1 sidecar was published for this url"},
	}}
	useFakeDownload(t, f)
	var code int
	stderr := captureStderr(t, func() { code = dispatch([]string{"download", "jar", "mq"}, &fakeRunner{}) })
	if code != 0 {
		t.Errorf("dispatch = %d, want 0 (no Failed entries)", code)
	}
	for _, want := range []string{
		"omitted: bcprov-jdk18on-1.84.jar: the image has 1.84",
		"omitted: jakarta.jms-api-3.0.0.jar: the image has 3.1.0",
		"guessed version: logstash-logback-encoder-9.0.jar",
		"unverified: custom.jar: no .sha1 sidecar was published for this url",
		"2 omitted",
	} {
		if !strings.Contains(stderr, want) {
			t.Errorf("stderr missing %q, got:\n%s", want, stderr)
		}
	}
	if strings.Contains(stderr, "failed:") {
		t.Errorf("an omission, fallback, or unverified note must never be reported as a failure, got:\n%s", stderr)
	}
}

// TestDownloadReportExitZeroWhenEverythingOmitted pins the requirement that
// "you already have all of it" is success: a Report whose every artifact was
// omitted (nothing written, nothing skipped, nothing failed) must still exit
// 0, the counts footer must say so, and the "next:" hint must name
// --include-provided as the way to download anyway rather than the normal
// dir-pointing hint -- an operator naturally reaches for -f there, which does
// not apply: -f only overwrites a file already on disk, never an artifact the
// omission step already dropped.
func TestDownloadReportExitZeroWhenEverythingOmitted(t *testing.T) {
	f := &fakeDownload{resp: libs.Report{
		Omitted: []string{
			"bcprov-jdk18on-1.84.jar: the image has 1.84",
			"bcpkix-jdk18on-1.84.jar: the image has 1.84",
			"bcutil-jdk18on-1.84.jar: the image has 1.84",
			"jakarta.jms-api-3.0.0.jar: the image has 3.1.0",
		},
	}}
	useFakeDownload(t, f)
	var code int
	stderr := captureStderr(t, func() { code = dispatch([]string{"download", "jar", "mq", "--version", "9.4.2.0"}, &fakeRunner{}) })
	if code != 0 {
		t.Fatalf("dispatch = %d, want 0 when every artifact was omitted, stderr:\n%s", code, stderr)
	}
	if !strings.Contains(stderr, "0 written, 0 skipped, 4 omitted, 0 unverified, 0 failed") {
		t.Errorf("stderr counts footer missing the omitted count, got:\n%s", stderr)
	}
	if !strings.Contains(stderr, "--include-provided") {
		t.Errorf("stderr missing the --include-provided hint when everything was omitted, got:\n%s", stderr)
	}
	if strings.Contains(stderr, "point the docker/podman") {
		t.Errorf("the dir-pointing hint must not appear when everything was omitted (there is nothing to point at), got:\n%s", stderr)
	}
}

// TestDownloadReportNextHint pins the previously-uncovered "next:" hint: the
// base line always appears, and the omitted-jars clause appears only when
// Report.Omitted is non-empty (and something else was also written or
// skipped -- the all-omitted case is covered separately by
// TestDownloadReportExitZeroWhenEverythingOmitted, which takes the other
// "next:" branch entirely).
func TestDownloadReportNextHint(t *testing.T) {
	const base = "next: point the docker/podman libs.dir or kubernetes libs: config key at"
	const omittedClause = "the omitted jars are already on the connector image and need no copy"

	t.Run("no omissions", func(t *testing.T) {
		f := &fakeDownload{resp: libs.Report{Written: []string{"libs/a.jar"}}}
		useFakeDownload(t, f)
		stderr := captureStderr(t, func() { dispatch([]string{"download", "jar", "mq"}, &fakeRunner{}) })
		if !strings.Contains(stderr, base) {
			t.Errorf("stderr missing base next: hint, got:\n%s", stderr)
		}
		if strings.Contains(stderr, omittedClause) {
			t.Errorf("omitted-jars clause must be absent with an empty Report.Omitted, got:\n%s", stderr)
		}
	})

	t.Run("with omissions", func(t *testing.T) {
		f := &fakeDownload{resp: libs.Report{
			Written: []string{"libs/a.jar"},
			Omitted: []string{"bcprov-jdk18on-1.84.jar: the image has 1.84"},
		}}
		useFakeDownload(t, f)
		stderr := captureStderr(t, func() { dispatch([]string{"download", "jar", "mq"}, &fakeRunner{}) })
		if !strings.Contains(stderr, base) {
			t.Errorf("stderr missing base next: hint, got:\n%s", stderr)
		}
		if !strings.Contains(stderr, omittedClause) {
			t.Errorf("stderr missing the omitted-jars clause, got:\n%s", stderr)
		}
	})
}

// TestDownloadReportPrintsOmitListProvenance covers the "omit list:" line:
// annotated "(built in; describes <floor> and later)" when --omit-lib-file
// was left empty (Report.OmitListProvenance then names the embedded
// default), printed bare when an explicit path was given, and any per-line
// Report.OmitListWarnings (entries the omit list loader rejected as
// unparseable) surfaced as their own lines.
//
// The annotation carries the floor because the filename alone names one
// tag, and an operator reading "2.13.0" while deploying 2.14.1 would
// reasonably conclude the omissions were judged against the wrong image.
func TestDownloadReportPrintsOmitListProvenance(t *testing.T) {
	t.Run("built in", func(t *testing.T) {
		f := &fakeDownload{resp: libs.Report{
			Written:            []string{"libs/a.jar"},
			OmitListProvenance: "solace-pubsub-connector-ibmmq-2.13.0",
		}}
		useFakeDownload(t, f)
		stderr := captureStderr(t, func() { dispatch([]string{"download", "jar", "mq"}, &fakeRunner{}) })
		want := "omit list: solace-pubsub-connector-ibmmq-2.13.0 (built in; describes " + libs.EmbeddedListMinVersion + " and later)"
		if !strings.Contains(stderr, want) {
			t.Errorf("stderr missing %q, got:\n%s", want, stderr)
		}
	})

	t.Run("explicit file", func(t *testing.T) {
		f := &fakeDownload{resp: libs.Report{
			Written:            []string{"libs/a.jar"},
			OmitListProvenance: "./my-image-libs.txt",
			OmitListWarnings:   []string{"jackson-core-9zzz.jar: version \"9zzz\" rejected as unparseable -- treated as not provided by the image"},
		}}
		useFakeDownload(t, f)
		args := []string{"download", "jar", "mq", "--omit-lib-file", "./my-image-libs.txt"}
		stderr := captureStderr(t, func() { dispatch(args, &fakeRunner{}) })
		if !strings.Contains(stderr, "omit list: ./my-image-libs.txt") {
			t.Errorf("stderr missing the explicit-file provenance line, got:\n%s", stderr)
		}
		if strings.Contains(stderr, "./my-image-libs.txt (built in)") {
			t.Errorf("an explicit --omit-lib-file must not be annotated (built in), got:\n%s", stderr)
		}
		if !strings.Contains(stderr, "omit list warning: jackson-core-9zzz.jar") {
			t.Errorf("stderr missing the omit list warning line, got:\n%s", stderr)
		}
	})
}

// TestDownloadSetMapMatchesModel guards download/jar's third command level
// the same way TestDispatchHandlersMatchModel guards verbs and generate's
// target: downloadSets (main.go's set-name dispatch table) must name exactly
// the sets cliVerbs models under download's "jar" target, and both must name
// exactly internal/libs's own SetNames() -- so the CLI model, dispatch, and
// the package can never quietly disagree about what "mq" or "syslog" means.
func TestDownloadSetMapMatchesModel(t *testing.T) {
	assertSameNameSet(t, "download jar set", nameSet(targetSetNames("download", "jar")), keySet(downloadSets))
	assertSameNameSet(t, "download jar set (libs)", nameSet(libs.SetNames()), keySet(downloadSets))
}

// TestDownloadJMSFlagIsGone pins the removal itself. --jms selected between the
// jakarta and javax IBM MQ clients; the javax build cannot satisfy the image's
// jakarta.jms binder, so the flag could only ever be set wrong, and both
// clients in one directory is worse still (they carry the same com.ibm.mq.*
// classes, and load order decides).
//
// The failure mode worth pinning is the quiet one: a script still passing the
// flag must fail as a usage error rather than appear to work while the argument
// is ignored. Go's flag package reports it as an unknown flag, which is exit 2
// here, and downloadFn is never reached.
func TestDownloadJMSFlagIsGone(t *testing.T) {
	for _, args := range [][]string{
		{"download", "jar", "mq", "--jms", "jakarta"},
		{"download", "jar", "mq", "--jms", "javax"},
		{"download", "jar", "syslog", "--jms", "jakarta"},
	} {
		f := &fakeDownload{}
		useFakeDownload(t, f)
		if code := dispatch(args, &fakeRunner{}); code != 2 {
			t.Errorf("dispatch(%v) = %d, want 2 (unknown flag)", args, code)
		}
		if len(f.calls) != 0 {
			t.Errorf("dispatch(%v): downloadFn must not run, got %d calls", args, len(f.calls))
		}
	}
}

// downloadEnvWithImage writes an env.yaml carrying just the top-level image
// block, which is the only part download reads.
func downloadEnvWithImage(t *testing.T, dir, tag string) string {
	t.Helper()
	body := "image:\n  name: solace/solace-pubsub-connector-ibmmq\n  tag: " + tag + "\n"
	write(t, dir, "env.yaml", body)
	return filepath.Join(dir, "env.yaml")
}

// TestDownloadReadsDeployedImageFromEnv pins the five precedence rows for the
// config read download just gained.
//
// The read exists only so the command can say when the jar list it omits
// against was captured from a different image than the one being deployed. It
// is advisory by design: download is the command you run BEFORE you have a
// deployment, so it has to work in an empty directory. That is why an absent
// default env.yaml is silent while a file the operator NAMED failing to load
// is systemic -- the same split --omit-lib-file already draws.
func TestDownloadReadsDeployedImageFromEnv(t *testing.T) {
	const wantRef = "solace/solace-pubsub-connector-ibmmq:2.14.1"

	t.Run("default env.yaml present: the image reaches libs", func(t *testing.T) {
		dir := t.TempDir()
		downloadEnvWithImage(t, dir, "2.14.1")
		f := &fakeDownload{}
		useFakeDownload(t, f)
		t.Chdir(dir)
		if code := dispatch([]string{"download", "jar", "mq"}, &fakeRunner{}); code != 0 {
			t.Fatalf("exit=%d, want 0", code)
		}
		if got := f.calls[0].DeployedImage; got != wantRef {
			t.Errorf("DeployedImage = %q, want %q", got, wantRef)
		}
	})

	t.Run("no env.yaml at all: silent, and still runs", func(t *testing.T) {
		f := &fakeDownload{}
		useFakeDownload(t, f)
		t.Chdir(t.TempDir())
		if code := dispatch([]string{"download", "jar", "mq"}, &fakeRunner{}); code != 0 {
			t.Fatalf("exit=%d, want 0: download must work with no config at all", code)
		}
		if got := f.calls[0].DeployedImage; got != "" {
			t.Errorf("DeployedImage = %q, want empty", got)
		}
	})

	t.Run("explicit -e that cannot be read is systemic", func(t *testing.T) {
		f := &fakeDownload{}
		useFakeDownload(t, f)
		args := []string{"download", "jar", "mq", "-e", filepath.Join(t.TempDir(), "nope.yaml")}
		if code := dispatch(args, &fakeRunner{}); code != 1 {
			t.Errorf("exit=%d, want 1: the operator named this file", code)
		}
		if len(f.calls) != 0 {
			t.Errorf("downloadFn must not run, got %d calls", len(f.calls))
		}
	})

	t.Run("defaulted env.yaml that is malformed does not stop the download", func(t *testing.T) {
		dir := t.TempDir()
		write(t, dir, "env.yaml", "image: [this is not a mapping\n")
		f := &fakeDownload{}
		useFakeDownload(t, f)
		t.Chdir(dir)
		if code := dispatch([]string{"download", "jar", "mq"}, &fakeRunner{}); code != 0 {
			t.Fatalf("exit=%d, want 0: a broken config must not break a download that does not need it", code)
		}
		if got := f.calls[0].DeployedImage; got != "" {
			t.Errorf("DeployedImage = %q, want empty", got)
		}
	})

	t.Run("env.yaml with no image block", func(t *testing.T) {
		dir := t.TempDir()
		write(t, dir, "env.yaml", "workflows:\n  dir: .\n")
		f := &fakeDownload{}
		useFakeDownload(t, f)
		t.Chdir(dir)
		if code := dispatch([]string{"download", "jar", "mq"}, &fakeRunner{}); code != 0 {
			t.Fatalf("exit=%d, want 0", code)
		}
		if got := f.calls[0].DeployedImage; got != "" {
			t.Errorf("DeployedImage = %q, want empty: nothing to compare", got)
		}
	})
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

// podmanEnv renders a minimal-but-valid podman: section with the unit dir
// overridden to a test-owned location (the default user-scope dir lives under the
// real home directory). ToSlash keeps the Windows temp path valid YAML.
//
// base-dir is pointed at that same directory so the callers asserting on written
// files can keep looking in one place. The two are independent in the schema --
// TestDeployPodmanSplitsBaseDirFromQuadletDir is what proves that.
func podmanEnv(quadletDir string) string {
	return fmt.Sprintf(`
podman:
  command: podman
  name: solmq-conn
  base-dir: %[1]s
  ports:
    - 8090
  quadlet:
    scope: user
    dir: %[1]s
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
  base-dir: %[1]s
  ports:
    - 8090
  quadlet:
    scope: user
    dir: %[1]s
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
	// dockerEnv sets no project-name, so this is the defaulted value, and it is
	// the document's first line -- the compose project this stack belongs to.
	if want := "name: " + spec.DefaultComposeProject + "\n"; !strings.HasPrefix(string(b), want) {
		t.Errorf("compose must open with %q:\n%s", want, b)
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

// TestGeneratePodmanVolumeSourcesAreAbsolute pins the host side of every Volume=
// line against the way an operator actually invokes the tool: `-e env.yaml` from
// the file's own directory, which used to leave envDir as "." and emit
// `Volume=certs/truststore.jks:...`.
//
// A relative source is not a near-miss in a quadlet unit. systemd starts the unit
// with no useful cwd, and podman reads a source with no ./ or / prefix as a NAMED
// VOLUME -- so `Volume=libs:/app/external/libs:ro` silently mounts an empty volume
// over the jars instead of failing, and the connector dies looking for classes
// that were never mounted. The relative spelling has to be impossible to emit,
// which is why this asserts on the rendered unit rather than on absPath.
func TestGeneratePodmanVolumeSourcesAreAbsolute(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "env.yaml", sharedEnv+`
tls:
  truststore:
    file: ./certs/truststore.jks
    password: ts
    type: JKS
`+podmanEnv(t.TempDir())+"  libs:\n    dir: ./libs\n")
	write(t, dir, "10.yaml", validWF)
	t.Chdir(dir)

	var code int
	stdout := captureStdout(t, func() {
		code = run([]string{"generate", "--platform", "podman", "-e", "env.yaml"})
	})
	if code != 0 {
		t.Fatalf("exit=%d", code)
	}
	// Only the two operator-supplied host paths are asserted. The application.yml
	// and status-script sources are deliberately bare names here: `generate`
	// passes an empty BaseDir for a portable preview, and only `deploy` resolves
	// them against the quadlet directory it writes them into.
	for _, want := range []string{
		filepath.Join(dir, "certs", "truststore.jks") + ":/app/external/classpath/truststores/truststore.jks:ro",
		filepath.Join(dir, "libs") + ":/app/external/libs:ro",
	} {
		if !strings.Contains(stdout, "Volume="+want) {
			t.Errorf("missing absolute mount %q in:\n%s", want, stdout)
		}
	}
}

// TestDeployPodmanSplitsBaseDirFromQuadletDir pins the split podman.base-dir
// introduces: the three mounted documents go where the operator asked, and only
// the .container unit goes to the quadlet directory, which is the one place
// systemd scans. The shared podmanEnv helper points both at one directory, so
// this is the only test that would catch them being wired together.
func TestDeployPodmanSplitsBaseDirFromQuadletDir(t *testing.T) {
	f := useFakeRunner(t)
	dir := t.TempDir()
	quadletDir := t.TempDir()
	baseDir := filepath.Join(t.TempDir(), "made-on-demand")
	write(t, dir, "env.yaml", sharedEnv+fmt.Sprintf(`
podman:
  command: podman
  name: solmq-conn
  base-dir: %s
  quadlet:
    scope: user
    dir: %s
`, filepath.ToSlash(baseDir), filepath.ToSlash(quadletDir)))
	write(t, dir, "10.yaml", validWF)

	if code := dispatch([]string{"deploy", "--platform", "podman", "-e", filepath.Join(dir, "env.yaml")}, f); code != 0 {
		t.Fatalf("exit=%d, calls=%+v", code, f.calls)
	}
	// base-dir did not exist: WriteFile creates it rather than failing.
	for _, name := range []string{"solmq-conn-application.yml", "solmq-conn-status"} {
		if _, err := os.Stat(filepath.Join(baseDir, name)); err != nil {
			t.Errorf("%s should be written to base-dir: %v", name, err)
		}
		if _, err := os.Stat(filepath.Join(quadletDir, name)); err == nil {
			t.Errorf("%s must not be written to the quadlet dir", name)
		}
	}
	if _, err := os.Stat(filepath.Join(quadletDir, "solmq-conn.container")); err != nil {
		t.Errorf("the unit belongs in the quadlet dir: %v", err)
	}
	// The unit's mounts must name base-dir, or the files it points at are not the
	// ones deploy just wrote.
	unit, rerr := os.ReadFile(filepath.Join(quadletDir, "solmq-conn.container"))
	if rerr != nil {
		t.Fatal(rerr)
	}
	// pathIn joins with "/" and keeps the separator style the spec was written in,
	// so the unit carries the ToSlash form above rather than the OS-native one the
	// Stat calls used. Both name the same file; only the unit's spelling is
	// asserted here, because that is what systemd and podman will read.
	if want := "Volume=" + filepath.ToSlash(baseDir) + "/solmq-conn-application.yml"; !strings.Contains(string(unit), want) {
		t.Errorf("unit missing %q:\n%s", want, unit)
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
	// No -p/--project-name: the compose project is declared by the file's own
	// top-level name: key instead, so the argv stays exactly as it was. Pinned
	// here so the choice is a recorded one rather than an omission.
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
// reach the docker child process's environment as NAME=value pairs (never in
// argv or on disk), covering both a literal credential (mq) and a -env one
// (solace) resolved from the real process environment.
//
// It also shows why an operator-chosen mount name cannot smuggle a dangerous
// variable into that child: an -env credential is keyed by the same variable it
// is read from, so the pair only ever restates what the child already inherited
// (SOLACE_PASS=envpass, from SOLACE_PASS). Literals keep a derived name.
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
		"_GEN_MQ_CONN_1_USER=app",
		"_GEN_MQ_CONN_1_PASSWORD=pw",
		"SOLACE_USER=envuser",
		"SOLACE_PASS=envpass",
	}
	if !reflect.DeepEqual(f.calls[1].env, want) {
		t.Errorf("child env = %v, want %v", f.calls[1].env, want)
	}
}

func TestRemoveDockerSeam(t *testing.T) {
	f := useFakeRunner(t)
	// remove asks before it tears down; this test is about the argv that
	// follows, so it answers yes. The prompt itself is covered in section g.
	withPromptAnswer(t, "y")
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
		{"podman", "secret", "rm", "--ignore", "solmq-conn-_GEN_MQ_CONN_1_USER"},
		{"podman", "secret", "create", "solmq-conn-_GEN_MQ_CONN_1_USER", "-"},
		{"podman", "secret", "rm", "--ignore", "solmq-conn-_GEN_MQ_CONN_1_PASSWORD"},
		{"podman", "secret", "create", "solmq-conn-_GEN_MQ_CONN_1_PASSWORD", "-"},
		{"podman", "secret", "rm", "--ignore", "solmq-conn-_GEN_SOL_CONN_1_CLIENT_USERNAME"},
		{"podman", "secret", "create", "solmq-conn-_GEN_SOL_CONN_1_CLIENT_USERNAME", "-"},
		{"podman", "secret", "rm", "--ignore", "solmq-conn-_GEN_SOL_CONN_1_CLIENT_PASSWORD"},
		{"podman", "secret", "create", "solmq-conn-_GEN_SOL_CONN_1_CLIENT_PASSWORD", "-"},
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
	// remove asks before it tears down; this test is about the argv that
	// follows, so it answers yes. The prompt itself is covered in section g.
	withPromptAnswer(t, "y")
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
			"solmq-conn-_GEN_MQ_CONN_1_USER", "solmq-conn-_GEN_MQ_CONN_1_PASSWORD",
			"solmq-conn-_GEN_SOL_CONN_1_CLIENT_USERNAME", "solmq-conn-_GEN_SOL_CONN_1_CLIENT_PASSWORD"},
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
//
// status now answers two questions behind a target word, so these cases come in
// three groups: the word itself and the flag combinations that cannot mean
// anything (rejected before any query runs), the container view (engine facts,
// no exec into anything), and the application view (the in-container script,
// still through the probe/install/run dance).
//
// The engine queries are asserted by their argv and their COUNT: one call for
// many instances is the whole cost argument for collecting engine facts by
// default, and a regression to one call per instance would pass every content
// assertion while quietly multiplying the cost.

// podsJSON is the `kubectl get pods -o json` answer these cases work from: one
// running pod and one crash-looping pod of the same deployment.
const podsJSON = `{"items":[
	  {"metadata":{"name":"pod-a","namespace":"prod"},
	   "spec":{"nodeName":"node-a","containers":[{"name":"connector","image":"solace/x:2.14.1",
	     "readinessProbe":{"tcpSocket":{"port":8090}},
	     "resources":{"limits":{"cpu":"1","memory":"1Gi"}},
	     "volumeMounts":[{"name":"config","mountPath":"/app/external/spring/config/application.yml"}]}],
	   "volumes":[{"name":"config","configMap":{"name":"solmq-connector-config"}}]},
	   "status":{"phase":"Running","containerStatuses":[{"name":"connector","ready":true,"restartCount":0,
	     "image":"solace/x:2.14.1","imageID":"docker-pullable://solace/x@sha256:9f2a1c",
	     "state":{"running":{"startedAt":"2026-08-18T04:12:07Z"}}}]}},
	  {"metadata":{"name":"pod-b","namespace":"prod"},
	   "spec":{"containers":[{"name":"connector","image":"solace/x:2.14.1",
	     "readinessProbe":{"tcpSocket":{"port":8090}}}]},
	   "status":{"phase":"Running","startTime":"2026-08-18T04:11:05Z",
	    "containerStatuses":[{"name":"connector","ready":false,"restartCount":7,
	     "image":"solace/x:2.14.1",
	     "state":{"waiting":{"reason":"CrashLoopBackOff"}},
	     "lastState":{"terminated":{"exitCode":137,"reason":"OOMKilled"}}}]}}]}`

// onePodJSON is the same document narrowed to one healthy pod, for the cases
// that are about the application half rather than the table.
const onePodJSON = `{"items":[{"metadata":{"name":"pod-a"},
	  "spec":{"containers":[{"name":"connector","image":"solace/x:2.14.1"}]},
	  "status":{"phase":"Running","containerStatuses":[{"name":"connector","ready":true,"restartCount":0,
	    "image":"solace/x:2.14.1","state":{"running":{"startedAt":"2026-08-18T04:12:07Z"}}}]}}]}`

// dockerInspect is the `docker inspect` answer for one composed container.
const dockerInspect = `[{"Name":"/solmq-connector","State":{"Status":"running","StartedAt":"2026-08-18T04:12:07Z","Health":{"Status":"healthy"}},
  "RestartCount":0,"Config":{"Image":"solace/x:2.14.1","Labels":{"com.docker.compose.project":"eg"}},
  "HostConfig":{"NanoCpus":0,"Memory":0}}]`

const appReport = "leader-election mode: active_standby\nleader-election state: active\nhealth: UP\n"

// ---- the target word and the flag combinations that cannot mean anything ----

// TestStatusTargetWordIsRequired covers the deliberate breaking change: the two
// views answer different questions, so neither is a safe default and a bare
// `status` prints the verb's own usage instead of guessing.
func TestStatusTargetWordIsRequired(t *testing.T) {
	f := &fakeRunner{}
	var code int
	stderr := captureStderr(t, func() { code = dispatch([]string{"status"}, f) })
	if code != 2 {
		t.Errorf("exit=%d, want 2", code)
	}
	for _, want := range []string{"missing target", "container", "application", "all", "Targets:"} {
		if !strings.Contains(stderr, want) {
			t.Errorf("stderr should carry %q, got:\n%s", want, stderr)
		}
	}
	// The short spellings (cnt, app) are deliberately absent: aliases keep
	// working but are documented only in the markdown docs, and this page is
	// rendered by the same verbUsage the no-alias gate covers
	// (TestVerbUsagePages).
	for _, alias := range []string{" cnt ", " app "} {
		if strings.Contains(stderr, alias) {
			t.Errorf("the help page should not show alias %q, got:\n%s", strings.TrimSpace(alias), stderr)
		}
	}
	if len(f.calls) != 0 {
		t.Errorf("nothing must run, got %d calls", len(f.calls))
	}
}

func TestStatusUnknownAndExtraTargetWords(t *testing.T) {
	cases := []struct {
		name string
		args []string
		want string
	}{
		{"unknown word", []string{"status", "bogus"}, "unknown target"},
		{"a second word", []string{"status", "all", "extra"}, "unexpected argument"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			f := &fakeRunner{}
			var code int
			stderr := captureStderr(t, func() { code = dispatch(c.args, f) })
			if code != 2 {
				t.Errorf("exit=%d, want 2", code)
			}
			if !strings.Contains(stderr, c.want) {
				t.Errorf("stderr should say %q, got:\n%s", c.want, stderr)
			}
			if len(f.calls) != 0 {
				t.Errorf("nothing must run, got %d calls", len(f.calls))
			}
		})
	}
}

// TestStatusTargetsMatchModel is the drift gate between the words the model
// documents, the constants the views switch on, and the bracket every usage
// line and doc renders. statusTargetArgBracket cannot be built from the model
// (cliVerbs' own initialiser uses it), so this is what keeps them in step.
func TestStatusTargetsMatchModel(t *testing.T) {
	modeled := targetNames("status")
	want := []string{statusTargetContainer, statusTargetApplication, statusTargetAll}
	if strings.Join(modeled, ",") != strings.Join(want, ",") {
		t.Errorf("modeled target words = %v, want %v", modeled, want)
	}
	if got := "<" + pipeList(modeled) + ">"; got != statusTargetArgBracket {
		t.Errorf("statusTargetArgBracket = %q, want %q", statusTargetArgBracket, got)
	}
	// Every word the views accept has to be a modeled word, and every modeled
	// word has to reach a view.
	for _, w := range modeled {
		if _, code := statusView([]string{w}); code != 0 {
			t.Errorf("modeled word %q does not resolve to a view", w)
		}
	}
}

func TestStatusTargetAliasesResolve(t *testing.T) {
	cases := []struct{ typed, want string }{
		{"cnt", statusTargetContainer},
		{"app", statusTargetApplication},
		{"container", statusTargetContainer},
		{"all", statusTargetAll},
		{"bogus", "bogus"}, // unknown words pass through for the caller to reject
	}
	for _, c := range cases {
		if got := resolveTarget("status", c.typed); got != c.want {
			t.Errorf("resolveTarget(%q) = %q, want %q", c.typed, got, c.want)
		}
	}
	// And an alias really drives the view: cnt runs the container half, which
	// never execs into anything.
	q := &queueRunner{resp: []queuedResp{{"", nil}, {podsJSON, nil}}}
	var code int
	captureStdout(t, func() {
		code = dispatch([]string{"sts", "cnt", "--platform", "kubernetes", "--pod", "pod-a", "--pod", "pod-b"}, q)
	})
	if code != 0 {
		t.Fatalf("exit=%d, want 0", code)
	}
	if len(q.calls) != 2 {
		t.Errorf("the container view should cost preflight + one get, got %d calls: %+v", len(q.calls), q.calls)
	}
}

// TestStatusRejectsImpossibleFlagCombinations covers every combination that
// cannot mean anything. Each is refused before a single query runs, rather than
// being silently ignored -- an operator who typed --install on the container
// view expects it to do something.
func TestStatusRejectsImpossibleFlagCombinations(t *testing.T) {
	cases := []struct {
		name string
		args []string
		want string
	}{
		{"unknown output format", []string{"status", "all", "--output", "yaml"}, "unknown --output"},
		{"json and watch", []string{"status", "all", "--output", "json", "-w"}, "cannot be combined with --watch"},
		{"all with an explicit pod", []string{"status", "all", "--all", "--pod", "pod-a"}, "cannot be combined with --pod/--container"},
		{"all with an explicit container", []string{"status", "all", "--all", "--container", "c1"}, "cannot be combined with --pod/--container"},
		{"install on the container view", []string{"status", "container", "--install"}, "--install applies to the application"},
		{"user on the container view", []string{"status", "cnt", "--user", "ops"}, "--user applies to the application"},
		{"management port on the container view", []string{"status", "container", "--management-port", "9090"}, "--management-port applies to the application"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			f := &fakeRunner{}
			var code int
			stderr := captureStderr(t, func() { code = dispatch(c.args, f) })
			if code != 2 {
				t.Errorf("exit=%d, want 2", code)
			}
			if !strings.Contains(stderr, c.want) {
				t.Errorf("stderr should explain the conflict (%q), got:\n%s", c.want, stderr)
			}
			if len(f.calls) != 0 {
				t.Errorf("nothing must run, got %d calls", len(f.calls))
			}
		})
	}
}

// TestWatchFlagAcceptsBareAndInterval covers the flag that is boolean in every
// documented sense but also takes an interval. The =<seconds> form is
// deliberately undocumented (see watchFlag), so this is the only place it is
// asserted at all.
func TestWatchFlagAcceptsBareAndInterval(t *testing.T) {
	var w watchFlag
	if !w.IsBoolFlag() {
		t.Fatal("the bare -w spelling needs IsBoolFlag")
	}
	if err := w.Set("true"); err != nil || !w.on || w.interval != watchDefaultSeconds*time.Second {
		t.Errorf("bare -w = %+v, err %v", w, err)
	}
	if err := w.Set("30"); err != nil || !w.on || w.interval != 30*time.Second {
		t.Errorf("-w=30 = %+v, err %v", w, err)
	}
	if err := w.Set("false"); err != nil || w.on {
		t.Errorf("-w=false should turn it off, got %+v err %v", w, err)
	}
	for _, bad := range []string{"0", "3601", "soon"} {
		if err := (&watchFlag{}).Set(bad); err == nil {
			t.Errorf("-w=%s should be rejected", bad)
		}
	}
}

// ---- the container view ---------------------------------------------------------

// TestStatusContainerViewReadsEngineFactsWithoutExecing is the shape of the new
// view: one read-only query answers discovery and the whole table, and nothing
// is ever exec'd into -- the in-container script is not involved at all.
func TestStatusContainerViewReadsEngineFactsWithoutExecing(t *testing.T) {
	q := &queueRunner{resp: []queuedResp{
		{"", nil},       // preflight: kubectl auth can-i create deployment
		{podsJSON, nil}, // get pods -o json: discovery and facts together
	}}
	var code int
	stdout := captureStdout(t, func() {
		code = dispatch([]string{"status", "container", "--platform", "kubernetes", "--pod", "pod-a", "--pod", "pod-b"}, q)
	})
	if code != 0 {
		t.Fatalf("exit=%d, want 0", code)
	}
	for _, want := range []string{
		// Both pods share a namespace, so it rides in the banner rather than
		// repeating down a column of its own.
		"== kubernetes  prod ==",
		"NAME", "STATE", "READY", "RESTARTS", "AGE", "IMAGE",
		"pod-a", "running", "yes",
		"pod-b", "restarting (CrashLoopBackOff)", "no", "7",
		"solace/x:2.14.1",
	} {
		if !strings.Contains(stdout, want) {
			t.Errorf("table is missing %q, got:\n%s", want, stdout)
		}
	}
	if strings.Contains(stdout, "NAMESPACE") {
		t.Errorf("one namespace belongs in the banner, not in a column, got:\n%s", stdout)
	}
	if len(q.calls) != 2 {
		t.Fatalf("want 2 calls (preflight, get pods), got %d: %+v", len(q.calls), q.calls)
	}
	// Nothing was exec'd: no probe, no install, no script run.
	for _, c := range q.calls {
		for _, tok := range c.argv {
			if tok == "exec" {
				t.Errorf("the container view must not exec into anything: %v", c.argv)
			}
		}
	}
	if got := q.calls[1].argv; got[len(got)-1] != "json" {
		t.Errorf("discovery should ask for json, got %v", got)
	}
}

// TestStatusContainerDetailsSamplesAndChecksComponents covers what --details
// adds on kubernetes: one sampling call for the whole run, and one presence
// check per distinct referenced object -- deduplicated, since every pod of a
// deployment references the same ones.
func TestStatusContainerDetailsSamplesAndChecksComponents(t *testing.T) {
	q := &queueRunner{resp: []queuedResp{
		{"", nil},       // preflight
		{podsJSON, nil}, // get pods
		{"pod-a   connector   120m   512Mi\n", nil},                                // top pod --containers
		{`{"kind":"ConfigMap","metadata":{"name":"solmq-connector-config"}}`, nil}, // get configmap (pod-a's only reference)
	}}
	var code int
	stdout := captureStdout(t, func() {
		code = dispatch([]string{"status", "container", "-d", "--platform", "kubernetes", "--pod", "pod-a", "--pod", "pod-b"}, q)
	})
	if code != 0 {
		t.Fatalf("exit=%d, want 0", code)
	}
	for _, want := range []string{
		"NODE", "node-a",
		"digest:", "sha256:9f2a1c",
		"cpu:", "120m of 1 (12%)",
		"memory:", "512Mi of 1Gi (50%)",
		"components:", "configmap", "solmq-connector-config", "present",
	} {
		if !strings.Contains(stdout, want) {
			t.Errorf("details view is missing %q, got:\n%s", want, stdout)
		}
	}
	if len(q.calls) != 4 {
		t.Fatalf("want 4 calls (preflight, get pods, top, one component check), got %d: %+v", len(q.calls), q.calls)
	}
	if !containsToken(q.calls[2].argv, "--containers") {
		t.Errorf("the sample must be per container, got %v", q.calls[2].argv)
	}
}

// TestStatusContainerDetailsWithoutMetricsServerDegradesToANote covers the
// cluster with no metrics API: an optional add-on being absent must cost the
// resource lines and nothing else -- the report still prints, and the exit code
// is still about the instances, not about the collection.
func TestStatusContainerDetailsWithoutMetricsServerDegradesToANote(t *testing.T) {
	q := &queueRunner{resp: []queuedResp{
		{"", nil},
		{onePodJSON, nil},
		{"error: Metrics API not available\n", fmt.Errorf("exit status 1")}, // top fails
	}}
	var code int
	stdout := captureStdout(t, func() {
		code = dispatch([]string{"status", "container", "-d", "--platform", "kubernetes", "--pod", "pod-a"}, q)
	})
	if code != 0 {
		t.Fatalf("exit=%d, want 0 (a missing add-on is not a failed status run)", code)
	}
	if !strings.Contains(stdout, "status: no resource usage") {
		t.Errorf("the report should carry a note, got:\n%s", stdout)
	}
	if !strings.Contains(stdout, "metrics-server") {
		t.Errorf("the note should say what to install, got:\n%s", stdout)
	}
	if !strings.Contains(stdout, "pod-a") {
		t.Errorf("the table must still print, got:\n%s", stdout)
	}
}

// TestStatusDockerContainerViewIsOneInspect covers docker: one inspect answers
// every target, and the compose project comes off the container's own label in
// that same call rather than costing a second lookup per target.
func TestStatusDockerContainerViewIsOneInspect(t *testing.T) {
	q := &queueRunner{resp: []queuedResp{
		{"", nil},            // preflight: docker info
		{dockerInspect, nil}, // inspect
	}}
	var code int
	stdout := captureStdout(t, func() {
		code = dispatch([]string{"status", "container", "--platform", "docker", "--container", "solmq-connector"}, q)
	})
	if code != 0 {
		t.Fatalf("exit=%d, want 0", code)
	}
	// docker reports a healthcheck verdict where kubernetes reports readiness.
	for _, want := range []string{"HEALTH", "healthy", "solmq-connector", "solace/x:2.14.1"} {
		if !strings.Contains(stdout, want) {
			t.Errorf("table is missing %q, got:\n%s", want, stdout)
		}
	}
	if strings.Contains(stdout, "READY") {
		t.Errorf("docker has no readiness concept, got:\n%s", stdout)
	}
	if len(q.calls) != 2 {
		t.Fatalf("want 2 calls (preflight, inspect), got %d: %+v", len(q.calls), q.calls)
	}
}

// TestStatusPodmanRestartCountComesFromSystemd covers the quadlet truth: podman
// recreates the container on restart, so its own counter reads 0 and systemd is
// the only thing that remembers.
func TestStatusPodmanRestartCountComesFromSystemd(t *testing.T) {
	podmanInspect := `[{"Name":"solmq-connector","State":{"Status":"running","StartedAt":"2026-08-18T04:12:07Z"},
	  "RestartCount":0,"Config":{"Image":"solace/x:2.14.1"},"HostConfig":{}}]`
	q := &queueRunner{resp: []queuedResp{
		{"", nil},            // preflight: podman info
		{podmanInspect, nil}, // inspect
		{"7\n", nil},         // systemctl show ... NRestarts
	}}
	var code int
	stdout := captureStdout(t, func() {
		code = dispatch([]string{"status", "container", "--platform", "podman", "--container", "solmq-connector"}, q)
	})
	if code != 0 {
		t.Fatalf("exit=%d, want 0", code)
	}
	if !strings.Contains(stdout, "7") {
		t.Errorf("the systemd count should reach the table, got:\n%s", stdout)
	}
	last := q.calls[len(q.calls)-1].argv
	if last[0] != "systemctl" || !containsToken(last, "NRestarts") {
		t.Errorf("last call = %v, want a systemctl NRestarts read", last)
	}
}

// TestStatusPodmanRestartCountFallsBackWhenSystemdCannotAnswer covers a
// container systemd knows nothing about (not quadlet-managed, or no systemd on
// the host): the container's own counter stays, and nothing fails.
func TestStatusPodmanRestartCountFallsBackWhenSystemdCannotAnswer(t *testing.T) {
	podmanInspect := `[{"Name":"solmq-connector","State":{"Status":"running","StartedAt":"2026-08-18T04:12:07Z"},
	  "RestartCount":3,"Config":{"Image":"solace/x:2.14.1"},"HostConfig":{}}]`
	q := &queueRunner{resp: []queuedResp{
		{"", nil},
		{podmanInspect, nil},
		{"", fmt.Errorf("systemctl: not found")},
	}}
	var code int
	stdout := captureStdout(t, func() {
		code = dispatch([]string{"status", "container", "--platform", "podman", "--container", "solmq-connector"}, q)
	})
	if code != 0 {
		t.Fatalf("exit=%d, want 0", code)
	}
	if !strings.Contains(stdout, "3") {
		t.Errorf("the container's own count should remain, got:\n%s", stdout)
	}
}

// TestStatusAllSearchesByImage covers --all on both kinds of platform: the
// operator names nothing, so the image reference is what identifies an
// instance.
func TestStatusAllSearchesByImage(t *testing.T) {
	t.Run("kubernetes searches every namespace", func(t *testing.T) {
		pods := `{"items":[
		  {"metadata":{"name":"ours","namespace":"prod"},
		   "spec":{"containers":[{"name":"connector","image":"reg/solace-pubsub-connector-ibmmq:2.14.1"}]},
		   "status":{"phase":"Running","containerStatuses":[{"name":"connector","ready":true,"state":{"running":{"startedAt":"2026-08-18T04:12:07Z"}}}]}},
		  {"metadata":{"name":"another-teams","namespace":"other"},
		   "spec":{"containers":[{"name":"connector","image":"reg/solace-pubsub-connector-ibmmq:2.13.0"}]},
		   "status":{"phase":"Running","containerStatuses":[{"name":"connector","ready":true,"state":{"running":{"startedAt":"2026-08-18T04:12:07Z"}}}]}},
		  {"metadata":{"name":"theirs","namespace":"other"},
		   "spec":{"containers":[{"name":"app","image":"nginx:1.27"}]},
		   "status":{"phase":"Running"}}]}`
		q := &queueRunner{resp: []queuedResp{{"", nil}, {pods, nil}}}
		var code int
		stdout := captureStdout(t, func() {
			code = dispatch([]string{"status", "container", "--platform", "kubernetes", "--all"}, q)
		})
		if code != 0 {
			t.Fatalf("exit=%d, want 0", code)
		}
		if !containsToken(q.calls[1].argv, "--all-namespaces") {
			t.Errorf("--all must search the whole cluster, got %v", q.calls[1].argv)
		}
		for _, want := range []string{"ours", "another-teams"} {
			if !strings.Contains(stdout, want) {
				t.Errorf("connector pod %q should be reported, got:\n%s", want, stdout)
			}
		}
		if strings.Contains(stdout, "theirs") {
			t.Errorf("an unrelated pod must be filtered out, got:\n%s", stdout)
		}
		// Instances found in different namespaces cannot share one banner, so each
		// row carries its own.
		if !strings.Contains(stdout, "NAMESPACE") {
			t.Errorf("a cluster-wide table spanning namespaces needs the namespace column, got:\n%s", stdout)
		}
		for _, want := range []string{"prod", "other"} {
			if !strings.Contains(stdout, want) {
				t.Errorf("namespace %q should appear, got:\n%s", want, stdout)
			}
		}
	})

	t.Run("docker lists then inspects the matches", func(t *testing.T) {
		list := "solmq-connector\tsolace/solace-pubsub-connector-ibmmq:2.14.1\nredis\tredis:7\n"
		inspect := `[{"Name":"/solmq-connector","State":{"Status":"running","StartedAt":"2026-08-18T04:12:07Z"},
		  "Config":{"Image":"solace/solace-pubsub-connector-ibmmq:2.14.1"},"HostConfig":{}}]`
		q := &queueRunner{resp: []queuedResp{{"", nil}, {list, nil}, {inspect, nil}}}
		var code int
		stdout := captureStdout(t, func() {
			code = dispatch([]string{"status", "container", "--platform", "docker", "--all"}, q)
		})
		if code != 0 {
			t.Fatalf("exit=%d, want 0", code)
		}
		if !containsToken(q.calls[1].argv, "ps") || !containsToken(q.calls[1].argv, "--all") {
			t.Errorf("discovery should list every container, got %v", q.calls[1].argv)
		}
		// Only the matching name reaches the inspect argv.
		if !containsToken(q.calls[2].argv, "solmq-connector") || containsToken(q.calls[2].argv, "redis") {
			t.Errorf("only matches should be inspected, got %v", q.calls[2].argv)
		}
		if !strings.Contains(stdout, "solmq-connector") {
			t.Errorf("got:\n%s", stdout)
		}
	})
}

// TestStatusAllWithNoMatchIsActionable covers the empty search: it names what
// it looked for, since there is no env.yaml in play to point at.
func TestStatusAllWithNoMatchIsActionable(t *testing.T) {
	q := &queueRunner{resp: []queuedResp{{"", nil}, {`{"items":[]}`, nil}}}
	var code int
	stderr := captureStderr(t, func() {
		code = dispatch([]string{"status", "container", "--platform", "kubernetes", "--all"}, q)
	})
	if code != 1 {
		t.Errorf("exit=%d, want 1", code)
	}
	if !strings.Contains(stderr, "solace-pubsub-connector-ibmmq") {
		t.Errorf("the error should name the image it searched for, got:\n%s", stderr)
	}
}

// ---- the application view -------------------------------------------------------

// TestStatusApplicationViewRunsTheScriptAndRendersItsFacts covers the path this
// verb has always had, now rendered from the parsed model rather than echoed
// verbatim: the values line up, and the banner is unchanged.
func TestStatusApplicationViewRunsTheScriptAndRendersItsFacts(t *testing.T) {
	q := &queueRunner{resp: []queuedResp{
		{"", nil},                                // preflight
		{onePodJSON, nil},                        // get pods
		{runner.ScriptPresentMarker + "\n", nil}, // probe: present
		{appReport + "workflows:\n   0: running\n  10: stopped\n", nil}, // run
	}}
	var code int
	stdout := captureStdout(t, func() {
		code = dispatch([]string{"status", "application", "--platform", "kubernetes", "--pod", "pod-a"}, q)
	})
	if code != 0 {
		t.Fatalf("exit=%d, want 0", code)
	}
	want := "=== kubernetes  pod-a ===\n" +
		"  leader-election mode:  active_standby\n" +
		"  leader-election state: active\n" +
		"  health:                UP\n" +
		"  workflows:\n" +
		"     0: running\n" +
		"    10: stopped\n"
	if stdout != want {
		t.Errorf("stdout =\n%q\nwant\n%q", stdout, want)
	}
	// The application view prints no table, even though it collected the engine
	// facts it needs to explain a failure.
	if strings.Contains(stdout, "STATE") {
		t.Errorf("the application view must not print the container table, got:\n%s", stdout)
	}
}

// TestStatusApplicationDetailsAddsTheEnrichmentLines covers -d on the
// application side: the script always prints every line it can, and the level
// decides which of them reach the operator.
func TestStatusApplicationDetailsAddsTheEnrichmentLines(t *testing.T) {
	full := appReport +
		"health components:\n  solace: UP\n  ibmmq: UP\n" +
		"uptime: 3d 4h 31m\nversion: 2.14.1\njava: openjdk 17.0.9\n" +
		"config: /app/external/spring/config/application.yml\n" +
		"heap: 432013312 of 1073741824\n"

	report := func(args ...string) string {
		q := &queueRunner{resp: []queuedResp{
			{"", nil}, {onePodJSON, nil},
			{runner.ScriptPresentMarker + "\n", nil}, {full, nil},
		}}
		return captureStdout(t, func() {
			if code := dispatch(args, q); code != 0 {
				t.Fatalf("%v: exit=%d, want 0", args, code)
			}
		})
	}

	details := report("status", "app", "-d", "--platform", "kubernetes", "--pod", "pod-a")
	for _, want := range []string{
		"uptime:", "3d 4h 31m",
		"version:", "2.14.1",
		"java:", "openjdk 17.0.9",
		"config:", "/app/external/spring/config/application.yml",
		// The script hands over raw bytes on purpose; the rendering happens here.
		"heap:", "412Mi of 1Gi (40%)",
		"health components:", "solace: UP", "ibmmq:  UP",
	} {
		if !strings.Contains(details, want) {
			t.Errorf("details block is missing %q, got:\n%s", want, details)
		}
	}

	// The same collection at the basic level prints none of it: one script run,
	// two levels of report.
	basic := report("status", "app", "--platform", "kubernetes", "--pod", "pod-a")
	for _, mustNot := range []string{"uptime:", "version:", "java:", "config:", "heap:", "health components:"} {
		if strings.Contains(basic, mustNot) {
			t.Errorf("the basic level must not carry %q, got:\n%s", mustNot, basic)
		}
	}
}

// TestStatusFailedScriptRunStillGetsItsOwnBlock is the other half of the
// rework: an instance whose script could not run keeps its banner, with the
// failure as a body line under the container facts that explain it. Before this
// change that instance produced one stderr line and no block at all.
func TestStatusFailedScriptRunStillGetsItsOwnBlock(t *testing.T) {
	q := &queueRunner{resp: []queuedResp{
		{"", nil},                                // preflight
		{podsJSON, nil},                          // get pods: pod-a fine, pod-b crash looping
		{runner.ScriptPresentMarker + "\n", nil}, // pod-a probe
		{runner.ScriptPresentMarker + "\n", nil}, // pod-b probe
		{appReport, nil},                         // pod-a run
		{"command terminated with exit code 126\n", fmt.Errorf("exit status 126")}, // pod-b run
	}}
	var code int
	stdout := captureStdout(t, func() {
		code = dispatch([]string{"status", "all", "--platform", "kubernetes", "--pod", "pod-a", "--pod", "pod-b"}, q)
	})
	if code != 1 {
		t.Fatalf("exit=%d, want 1 (an instance that could not be run)", code)
	}
	if !strings.Contains(stdout, "=== kubernetes  prod / pod-b ===") {
		t.Errorf("the failed instance needs its own banner, got:\n%s", stdout)
	}
	if !strings.Contains(stdout, "status: could not run the status script") {
		t.Errorf("the failure belongs in the block, got:\n%s", stdout)
	}
	// And the table above it says why: restarting, 7 restarts.
	if !strings.Contains(stdout, "restarting (CrashLoopBackOff)") {
		t.Errorf("the container facts should explain it, got:\n%s", stdout)
	}
	// A reachable instance in the same run still reports normally.
	if !strings.Contains(stdout, "=== kubernetes  prod / pod-a ===") {
		t.Errorf("one failure must not hide the rest, got:\n%s", stdout)
	}
}

func TestStatusInstallPaths(t *testing.T) {
	base := func(extra ...queuedResp) *queueRunner {
		return &queueRunner{resp: append([]queuedResp{
			{"", nil},                               // preflight
			{onePodJSON, nil},                       // get pods
			{runner.ScriptAbsentMarker + "\n", nil}, // probe: absent
		}, extra...)}
	}

	t.Run("--install installs without asking", func(t *testing.T) {
		q := base(queuedResp{"", nil}, queuedResp{appReport, nil})
		captureStdout(t, func() {
			if code := dispatch([]string{"status", "app", "--install", "--platform", "kubernetes", "--pod", "pod-a"}, q); code != 0 {
				t.Fatalf("exit=%d, want 0", code)
			}
		})
		if len(q.calls) != 5 {
			t.Fatalf("want 5 calls (preflight, get, probe, install, run), got %d: %+v", len(q.calls), q.calls)
		}
	})

	t.Run("prompt answered yes installs", func(t *testing.T) {
		withPromptAnswer(t, "y")
		q := base(queuedResp{"", nil}, queuedResp{appReport, nil})
		captureStdout(t, func() {
			if code := dispatch([]string{"status", "app", "--platform", "kubernetes", "--pod", "pod-a"}, q); code != 0 {
				t.Fatalf("exit=%d, want 0", code)
			}
		})
		if len(q.calls) != 5 {
			t.Fatalf("want 5 calls, got %d: %+v", len(q.calls), q.calls)
		}
	})

	t.Run("prompt declined skips the instance and exits 1", func(t *testing.T) {
		withPromptAnswer(t, "n")
		q := base()
		var code int
		stdout := captureStdout(t, func() {
			code = dispatch([]string{"status", "app", "--platform", "kubernetes", "--pod", "pod-a"}, q)
		})
		if code != 1 {
			t.Fatalf("exit=%d, want 1", code)
		}
		// The reason travels with the instance, not on stderr where nothing
		// explains it.
		if !strings.Contains(stdout, "status: the status script is not installed") {
			t.Errorf("the skip should be reported in the block, got:\n%s", stdout)
		}
		if len(q.calls) != 3 {
			t.Fatalf("nothing should be installed or run, got %d calls: %+v", len(q.calls), q.calls)
		}
	})
}

// TestStatusInstallPromptNonTTYRefusesWithInstallHint is the confirmInstall
// counterpart to TestPlatformMenuNonTTYRefusesWithPlatformHint: the install
// confirmation shares the same stdin seam, so it must refuse rather than block
// on a read that would never return, and point at --install.
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
	_ = pw.Close() // EOF: a read returns immediately, as it would in CI

	q := &queueRunner{resp: []queuedResp{
		{"", nil},
		{onePodJSON, nil},
		{runner.ScriptAbsentMarker + "\n", nil},
	}}
	var code int
	stdout := captureStdout(t, func() {
		code = dispatch([]string{"status", "app", "--platform", "kubernetes", "--pod", "pod-a"}, q)
	})
	if code != 1 {
		t.Fatalf("exit=%d, want 1", code)
	}
	if !strings.Contains(stdout, "--install") {
		t.Errorf("the refusal should point at --install, got:\n%s", stdout)
	}
	if len(q.calls) != 3 {
		t.Fatalf("nothing installed or run, got %d calls: %+v", len(q.calls), q.calls)
	}
}

// TestStatusStandbyIsAnAnswerNotAFailure keeps the contract the whole verb is
// built on: the script always exits 0 and carries the state in its output, so a
// standby instance prints like any other and the run still exits 0.
func TestStatusStandbyIsAnAnswerNotAFailure(t *testing.T) {
	q := &queueRunner{resp: []queuedResp{
		{"", nil},
		{podsJSON, nil},
		{runner.ScriptPresentMarker + "\n", nil},
		{runner.ScriptPresentMarker + "\n", nil},
		{"leader-election mode: active_standby\nleader-election state: standby\n", nil},
		{"leader-election mode: active_standby\nleader-election state: active\n", nil},
	}}
	var code int
	stdout := captureStdout(t, func() {
		code = dispatch([]string{"status", "app", "--platform", "kubernetes", "--pod", "pod-a", "--pod", "pod-b"}, q)
	})
	if code != 0 {
		t.Fatalf("exit=%d, want 0 (standby is an answer, not a failure)", code)
	}
	for _, want := range []string{"leader-election state: standby", "leader-election state: active"} {
		if !strings.Contains(stdout, want) {
			t.Errorf("stdout missing %q, got:\n%s", want, stdout)
		}
	}
}

// ---- output format ---------------------------------------------------------------

// TestStatusJSONOutputIsOneDocument covers --output json: the same model the
// tables render, as one parseable document on stdout.
func TestStatusJSONOutputIsOneDocument(t *testing.T) {
	q := &queueRunner{resp: []queuedResp{
		{"", nil},
		{onePodJSON, nil},
		{runner.ScriptPresentMarker + "\n", nil},
		{appReport, nil},
	}}
	var code int
	stdout := captureStdout(t, func() {
		code = dispatch([]string{"status", "all", "--output", "json", "--platform", "kubernetes", "--pod", "pod-a"}, q)
	})
	if code != 0 {
		t.Fatalf("exit=%d, want 0", code)
	}
	var doc map[string]any
	if err := json.Unmarshal([]byte(stdout), &doc); err != nil {
		t.Fatalf("stdout must be one json document: %v\n%s", err, stdout)
	}
	if doc["schemaVersion"] != float64(statusreport.SchemaVersion) {
		t.Errorf("schemaVersion = %v", doc["schemaVersion"])
	}
	insts, ok := doc["instances"].([]any)
	if !ok || len(insts) != 1 {
		t.Fatalf("instances = %v", doc["instances"])
	}
	inst := insts[0].(map[string]any)
	if inst["name"] != "pod-a" {
		t.Errorf("name = %v", inst["name"])
	}
	if _, ok := inst["container"]; !ok {
		t.Error("the container half belongs in the document")
	}
	if _, ok := inst["application"]; !ok {
		t.Error("the application half belongs in the document")
	}
}

// TestStatusImageMismatchIsReportedAtBothLevels covers the failed-rollout
// finding: a pod still on the old tag looks healthy in every other line of the
// report, so the comparison against `image:` in env.yaml has to surface at the
// basic level too -- where the per-instance detail block is not printed at all.
func TestStatusImageMismatchIsReportedAtBothLevels(t *testing.T) {
	dir := t.TempDir()
	envPath := filepath.Join(dir, "env.yaml")
	env := "image:\n  name: solace/x\n  tag: \"2.15.0\"\nkubernetes:\n  deployment:\n    name: solmq-connector\n    namespace: prod\n"
	if err := os.WriteFile(envPath, []byte(env), 0o600); err != nil {
		t.Fatal(err)
	}
	deployment := `{"kind":"Deployment","metadata":{"name":"solmq-connector"},"spec":{"replicas":1},"status":{"readyReplicas":1,"updatedReplicas":1,"availableReplicas":1}}`

	t.Run("basic reports it as a note", func(t *testing.T) {
		q := &queueRunner{resp: []queuedResp{{"", nil}, {onePodJSON, nil}, {deployment, nil}}}
		var code int
		stdout := captureStdout(t, func() {
			code = dispatch([]string{"status", "container", "-e", envPath, "--platform", "kubernetes", "--pod", "pod-a"}, q)
		})
		if code != 0 {
			t.Fatalf("exit=%d, want 0 (a wrong image is a finding, not a failed run)", code)
		}
		for _, want := range []string{"status:", "not running the image env.yaml configures", "solace/x:2.15.0", "--details"} {
			if !strings.Contains(stdout, want) {
				t.Errorf("the note should carry %q, got:\n%s", want, stdout)
			}
		}
	})

	t.Run("details reports it per instance", func(t *testing.T) {
		q := &queueRunner{resp: []queuedResp{
			{"", nil}, {onePodJSON, nil}, {deployment, nil},
			{"pod-a   connector   120m   512Mi\n", nil}, // top
		}}
		var code int
		stdout := captureStdout(t, func() {
			code = dispatch([]string{"status", "container", "-d", "-e", envPath, "--platform", "kubernetes", "--pod", "pod-a"}, q)
		})
		if code != 0 {
			t.Fatalf("exit=%d, want 0", code)
		}
		if !strings.Contains(stdout, "image-expected:") || !strings.Contains(stdout, "solace/x:2.15.0") {
			t.Errorf("the per-instance line is missing, got:\n%s", stdout)
		}
		// One statement of the finding, not two.
		if strings.Contains(stdout, "instance(s) are not running") {
			t.Errorf("the run-level note is redundant at the details level, got:\n%s", stdout)
		}
	})

	t.Run("a matching image says nothing at all", func(t *testing.T) {
		matching := "image:\n  name: solace/x\n  tag: \"2.14.1\"\nkubernetes:\n  deployment:\n    name: solmq-connector\n"
		matchPath := filepath.Join(dir, "match.yaml")
		if err := os.WriteFile(matchPath, []byte(matching), 0o600); err != nil {
			t.Fatal(err)
		}
		q := &queueRunner{resp: []queuedResp{{"", nil}, {onePodJSON, nil}, {deployment, nil}}}
		var code int
		stdout := captureStdout(t, func() {
			code = dispatch([]string{"status", "container", "-e", matchPath, "--platform", "kubernetes", "--pod", "pod-a"}, q)
		})
		if code != 0 {
			t.Fatalf("exit=%d, want 0", code)
		}
		if strings.Contains(stdout, "not running the image") {
			t.Errorf("a correct image must be silent, got:\n%s", stdout)
		}
	})
}

// TestStatusDockerProjectMismatchIsReported covers the compose-project finding.
// A stack brought up before docker.project-name existed is labelled with the
// basename of the directory its compose file was in, so `remove` -- which runs
// `compose down` against the configured project -- would tear down nothing and
// still report success. Nothing else in the report would say so, which is why
// the mismatch is a note.
func TestStatusDockerProjectMismatchIsReported(t *testing.T) {
	dir := t.TempDir()
	// dockerInspect labels the container with project "eg"; each env below names a
	// different project, so the arms differ only in whether they agree with it.
	writeEnv := func(t *testing.T, name, project string) string {
		t.Helper()
		path := filepath.Join(dir, name)
		env := "image:\n  name: solace/x\n  tag: \"2.14.1\"\ndocker:\n  command: docker\n  name: solmq-connector\n  project-name: " + project + "\n"
		if err := os.WriteFile(path, []byte(env), 0o600); err != nil {
			t.Fatal(err)
		}
		return path
	}

	t.Run("a different project is a note", func(t *testing.T) {
		envPath := writeEnv(t, "other.yaml", "solace-ibmmq-connectors")
		q := &queueRunner{resp: []queuedResp{{"", nil}, {dockerInspect, nil}}}
		var code int
		stdout := captureStdout(t, func() {
			code = dispatch([]string{"status", "container", "-e", envPath, "--platform", "docker", "--container", "solmq-connector"}, q)
		})
		if code != 0 {
			t.Fatalf("exit=%d, want 0 (a project mismatch is a finding, not a failed run)", code)
		}
		// The note must name both projects and the way out, so an operator can act
		// on it without reading the source.
		for _, want := range []string{"status:", "compose project", "eg", "solace-ibmmq-connectors", "docker.project-name"} {
			if !strings.Contains(stdout, want) {
				t.Errorf("the note should carry %q, got:\n%s", want, stdout)
			}
		}
	})

	t.Run("the configured project says nothing at all", func(t *testing.T) {
		envPath := writeEnv(t, "match.yaml", "eg")
		q := &queueRunner{resp: []queuedResp{{"", nil}, {dockerInspect, nil}}}
		var code int
		stdout := captureStdout(t, func() {
			code = dispatch([]string{"status", "container", "-e", envPath, "--platform", "docker", "--container", "solmq-connector"}, q)
		})
		if code != 0 {
			t.Fatalf("exit=%d, want 0", code)
		}
		if strings.Contains(stdout, "compose project") {
			t.Errorf("a matching project must be silent, got:\n%s", stdout)
		}
	})

	// A container compose never created carries no project label at all. That is
	// not the wrong project, it is a container outside compose entirely, and
	// reporting it as drift would be the likeliest false positive.
	t.Run("no compose label is not a mismatch", func(t *testing.T) {
		envPath := writeEnv(t, "nolabel.yaml", "solace-ibmmq-connectors")
		bare := `[{"Name":"/solmq-connector","State":{"Status":"running","StartedAt":"2026-08-18T04:12:07Z"},
		  "RestartCount":0,"Config":{"Image":"solace/x:2.14.1"},"HostConfig":{}}]`
		q := &queueRunner{resp: []queuedResp{{"", nil}, {bare, nil}}}
		var code int
		stdout := captureStdout(t, func() {
			code = dispatch([]string{"status", "container", "-e", envPath, "--platform", "docker", "--container", "solmq-connector"}, q)
		})
		if code != 0 {
			t.Fatalf("exit=%d, want 0", code)
		}
		if strings.Contains(stdout, "compose project") {
			t.Errorf("an unlabelled container must be silent, got:\n%s", stdout)
		}
	})
}

// TestStatusDockerDetailsAddsDigestAndStats covers what --details costs on
// docker and podman: the digest lives on the image rather than on the container
// there, so it is a second call, and the resource sample is a third.
func TestStatusDockerDetailsAddsDigestAndStats(t *testing.T) {
	imageInspect := `[{"Id":"sha256:local","RepoDigests":["solace/x@sha256:9f2a1c4d8e"]}]`
	stats := "solmq-connector\t0.15%\t512MiB / 1GiB\t50.00%\n"
	q := &queueRunner{resp: []queuedResp{
		{"", nil},            // preflight: docker info
		{dockerInspect, nil}, // inspect
		{imageInspect, nil},  // image inspect: the registry digest
		{stats, nil},         // stats --no-stream
	}}
	var code int
	stdout := captureStdout(t, func() {
		code = dispatch([]string{"status", "container", "-d", "--platform", "docker", "--container", "solmq-connector"}, q)
	})
	if code != 0 {
		t.Fatalf("exit=%d, want 0", code)
	}
	for _, want := range []string{"digest:", "sha256:9f2a1c4d8e", "cpu:", "0.15%", "memory:", "512MiB / 1GiB", "(50.00%)"} {
		if !strings.Contains(stdout, want) {
			t.Errorf("details block is missing %q, got:\n%s", want, stdout)
		}
	}
	if len(q.calls) != 4 {
		t.Fatalf("want 4 calls (preflight, inspect, image inspect, stats), got %d: %+v", len(q.calls), q.calls)
	}
	if !containsToken(q.calls[3].argv, "--no-stream") {
		t.Errorf("the sample must not stream, got %v", q.calls[3].argv)
	}
}

// ---- boundary checks, unchanged in intent -----------------------------------------

// TestStatusRejectsUnsafeUserBeforeAnyExec pins the boundary check on --user:
// the account name reaches a sed address inside the generated script, so a
// value carrying a regex or path metacharacter must be refused up front rather
// than silently producing an unauthenticated request.
func TestStatusRejectsUnsafeUserBeforeAnyExec(t *testing.T) {
	for _, user := range []string{"bad/user", "who$e", "a b", "quote'd"} {
		f := &fakeRunner{}
		code := dispatch([]string{"status", "app", "--platform", "docker", "--container", "c1", "--user", user}, f)
		if code != 1 {
			t.Errorf("--user %q: exit=%d, want 1", user, code)
		}
		if len(f.calls) != 0 {
			t.Errorf("--user %q: runner must not be invoked, got %d calls", user, len(f.calls))
		}
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
		{"bad pod name", []string{"status", "all", "--platform", "kubernetes", "--pod", "bad pod name"}},
		{"bad namespace", []string{"status", "all", "--platform", "kubernetes", "--pod", "goodpod", "--namespace", "bad;ns"}},
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
			args := []string{"status", "app", "--platform", "kubernetes", "--pod", "pod-a", "--management-port", p}
			if code := dispatch(args, f); code != 1 {
				t.Errorf("--management-port %s: exit=%d, want 1", p, code)
			}
			if len(f.calls) != 0 {
				t.Errorf("--management-port %s: runner must not be invoked, got %d calls", p, len(f.calls))
			}
		})
	}
}

// TestStatusNoPodsFoundNamesTheSelector covers discovery from env.yaml with
// nothing matching: the message has to carry the selector and namespace it
// used, since those are what the operator would fix.
func TestStatusNoPodsFoundNamesTheSelector(t *testing.T) {
	dir := t.TempDir()
	envPath := filepath.Join(dir, "env.yaml")
	if err := os.WriteFile(envPath, []byte("kubernetes:\n  deployment:\n    name: solmq-connector\n    namespace: prod\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	q := &queueRunner{resp: []queuedResp{{"", nil}, {`{"items":[]}`, nil}}}
	var code int
	stderr := captureStderr(t, func() {
		code = dispatch([]string{"status", "container", "--platform", "kubernetes", "-e", envPath}, q)
	})
	if code != 1 {
		t.Fatalf("exit=%d, want 1", code)
	}
	for _, want := range []string{"app=solmq-connector", "prod", "--pod"} {
		if !strings.Contains(stderr, want) {
			t.Errorf("the error should carry %q, got:\n%s", want, stderr)
		}
	}
}

// containsToken reports whether argv carries the exact token, so a case can
// assert one flag without pinning the whole slice.
func containsToken(argv []string, want string) bool {
	for _, a := range argv {
		if a == want {
			return true
		}
	}
	return false
}

// TestStatusAllSkipsAnUnsafeNameOutLoud covers the re-validation on the way back
// from the engine: under --all the container names come from `ps` rather than
// from the operator, so one that could not go into an argv is skipped with a
// note naming it, never passed on, and the instances around it are still
// reported.
//
// This lives here rather than under logs because --all is a status flag: logs
// reads one instance and has no image search.
func TestStatusAllSkipsAnUnsafeNameOutLoud(t *testing.T) {
	img := "reg/" + statusreport.ImageMatch + ":2.14.1"
	list := "bad;name\t" + img + "\nsolmq-connector\t" + img + "\n"
	// The inspect doc has to carry a matching image too: under --all,
	// ParseInspect filters by the same image match the ps list was filtered by.
	inspect := `[{"Name":"/solmq-connector","State":{"Status":"running","StartedAt":"2026-08-18T04:12:07Z"},
	  "RestartCount":0,"Config":{"Image":"` + img + `"},"HostConfig":{}}]`
	q := &queueRunner{resp: []queuedResp{{"", nil}, {list, nil}, {inspect, nil}}}
	var code int
	stdout := captureStdout(t, func() {
		code = dispatch([]string{"status", "container", "--platform", "docker", "--all"}, q)
	})
	if code != 0 {
		t.Fatalf("exit=%d, want 0", code)
	}
	for _, want := range []string{"bad;name", "unsafe character"} {
		if !strings.Contains(stdout, want) {
			t.Errorf("the skip should name the container and why, got:\n%s", want+"\n"+stdout)
		}
	}
	for _, c := range q.calls {
		if containsToken(c.argv, "bad;name") {
			t.Errorf("an unsafe name must never reach an argv: %v", c.argv)
		}
	}
	if !strings.Contains(stdout, "solmq-connector") {
		t.Errorf("the safe instance should still be reported, got:\n%s", stdout)
	}
}

// TestStatusAllSortsRowsAlphabetically pins the ordering an index depends on.
// docker ps returns creation order, so without the sort --pod/--container <n>
// would select a different instance than the row the operator counted to.
func TestStatusAllSortsRowsAlphabetically(t *testing.T) {
	img := "reg/" + statusreport.ImageMatch + ":2.14.1"
	// Newest first, which is what ps gives -- deliberately not alphabetical.
	list := "zeta\t" + img + "\nalpha\t" + img + "\n"
	inspect := `[{"Name":"/zeta","State":{"Status":"running","StartedAt":"2026-08-18T04:12:07Z"},
	  "RestartCount":0,"Config":{"Image":"` + img + `"},"HostConfig":{}},
	 {"Name":"/alpha","State":{"Status":"running","StartedAt":"2026-08-18T04:12:07Z"},
	  "RestartCount":0,"Config":{"Image":"` + img + `"},"HostConfig":{}}]`
	q := &queueRunner{resp: []queuedResp{{"", nil}, {list, nil}, {inspect, nil}}}
	var code int
	stdout := captureStdout(t, func() {
		code = dispatch([]string{"status", "container", "--platform", "docker", "--all"}, q)
	})
	if code != 0 {
		t.Fatalf("exit=%d, want 0", code)
	}
	if i, j := strings.Index(stdout, "alpha"), strings.Index(stdout, "zeta"); i < 0 || j < 0 || i > j {
		t.Errorf("rows must be sorted by name so an index means the same thing in both verbs, got:\n%s", stdout)
	}
}

// ---- n) logs --------------------------------------------------------------------
//
// logs shares status's platform resolution and instance discovery (instances.go),
// so these cases concentrate on what is its own: the per-platform argv it builds,
// the flag combinations it refuses, and the fact that every operator-supplied
// name is rejected before a process starts.
//
// Like the status cases, the assertions pin argv AND call count: discovery is one
// query for every instance, and a regression to one query per instance would pass
// every content assertion while quietly multiplying the cost.

// streamed is one Stream call's canned result.
type streamed struct {
	stdout string
	stderr string
	err    error
}

// streamRunner is the fake logs needs: it satisfies runner.Streamer as well as
// runner.Runner, since logs reads every log through the streaming seam (so the
// platform's own diagnostics stay on stderr rather than landing in a redirected
// log). Both kinds of call are recorded in one ordered slice, while the canned
// answers are queued per kind: Run serves the preflight probe and discovery,
// Stream serves the log reads.
type streamRunner struct {
	calls []fakeCall

	resp []queuedResp
	runN int

	streams []streamed
	streamN int
}

func (s *streamRunner) Run(c runner.Cmd) (string, error) {
	s.calls = append(s.calls, fakeCall{argv: c.Argv, stdin: c.Stdin, env: c.Env})
	i := s.runN
	s.runN++
	if i < len(s.resp) {
		return s.resp[i].out, s.resp[i].err
	}
	return "", nil
}

func (s *streamRunner) Stream(_ context.Context, c runner.Cmd, stdout, stderr io.Writer) error {
	s.calls = append(s.calls, fakeCall{argv: c.Argv, stdin: c.Stdin, env: c.Env})
	i := s.streamN
	s.streamN++
	if i >= len(s.streams) {
		return nil
	}
	st := s.streams[i]
	if st.stdout != "" {
		fmt.Fprint(stdout, st.stdout)
	}
	if st.stderr != "" {
		fmt.Fprint(stderr, st.stderr)
	}
	return st.err
}

// logsImage is an image reference carrying statusreport.ImageMatch, for the
// --all cases that find instances by image rather than by name.
var logsImage = "registry.example/solace/" + statusreport.ImageMatch + ":2.14.1"

// ---- the argv each platform gets ----

// TestLogsDockerArgvShape pins the docker read, which differs from kubernetes in
// more than spelling: the options come first and the container name last, and
// there is no namespace or container flag at all. It also pins that --since
// reaches the argv in the canonical duration form this tool produced, never the
// operator's spelling.
func TestLogsDockerArgvShape(t *testing.T) {
	q := &streamRunner{
		resp:    []queuedResp{{"", nil}},
		streams: []streamed{{stdout: "one line\n"}},
	}
	var code int
	stdout := captureStdout(t, func() {
		code = dispatch([]string{"logs", "--platform", "docker", "--container", "solmq-conn",
			"--tail", "100", "--since", "10m", "--timestamps"}, q)
	})
	if code != 0 {
		t.Fatalf("exit=%d, want 0", code)
	}
	if len(q.calls) != 2 {
		t.Fatalf("want 2 calls (preflight, one log read), got %d: %+v", len(q.calls), q.calls)
	}
	want := []string{"docker", "logs", "--timestamps", "--tail", "100", "--since", "10m0s", "solmq-conn"}
	if !reflect.DeepEqual(q.calls[1].argv, want) {
		t.Errorf("log argv = %v, want %v", q.calls[1].argv, want)
	}
	// One instance prints its log and nothing else, so the output pipes cleanly.
	if stdout != "one line\n" {
		t.Errorf("a single instance must print no heading, got %q", stdout)
	}
}

// TestLogsPodmanReadsTheContainerNotTheJournal pins that the quadlet path needs
// no new binary: podman logs reads a quadlet-managed container's output like any
// other, so journalctl never enters the allowlist.
func TestLogsPodmanReadsTheContainerNotTheJournal(t *testing.T) {
	q := &streamRunner{resp: []queuedResp{{"", nil}}, streams: []streamed{{stdout: "p\n"}}}
	if code := dispatch([]string{"logs", "--platform", "podman", "--container", "solmq-conn"}, q); code != 0 {
		t.Fatalf("exit=%d, want 0", code)
	}
	want := []string{"podman", "logs", "solmq-conn"}
	if !reflect.DeepEqual(q.calls[1].argv, want) {
		t.Errorf("log argv = %v, want %v", q.calls[1].argv, want)
	}
	for _, c := range q.calls {
		for _, tok := range c.argv {
			if tok == "journalctl" || tok == "systemctl" {
				t.Errorf("logs must read the container, not the unit journal: %v", c.argv)
			}
		}
	}
}

// TestLogsTailAllAndZero covers the two ends of --tail: the word "all" is the
// default and adds no flag at all, while 0 is a real request (the flags only, no
// history) and so cannot double as the unset marker.
func TestLogsTailAllAndZero(t *testing.T) {
	for _, c := range []struct {
		name string
		args []string
		want []string
	}{
		{"default", nil, []string{"docker", "logs", "solmq-conn"}},
		{"all", []string{"--tail", "all"}, []string{"docker", "logs", "solmq-conn"}},
		{"zero", []string{"--tail", "0"}, []string{"docker", "logs", "--tail", "0", "solmq-conn"}},
	} {
		t.Run(c.name, func(t *testing.T) {
			q := &streamRunner{resp: []queuedResp{{"", nil}}}
			args := append([]string{"logs", "--platform", "docker", "--container", "solmq-conn"}, c.args...)
			if code := dispatch(args, q); code != 0 {
				t.Fatalf("exit=%d, want 0", code)
			}
			if !reflect.DeepEqual(q.calls[1].argv, c.want) {
				t.Errorf("log argv = %v, want %v", q.calls[1].argv, c.want)
			}
		})
	}
}

// ---- the combinations that cannot mean anything ----

// TestLogsPreviousIsKubernetesOnly covers the one refusal that cannot be made
// until the platform is known: neither docker nor podman keeps a prior run's log
// under the same name, so --previous there is refused by name rather than
// silently ignored -- and refused before the preflight probe, since a flag that
// cannot work should not first make the operator wait on a daemon.
func TestLogsPreviousIsKubernetesOnly(t *testing.T) {
	for _, platform := range []string{"docker", "podman"} {
		t.Run(platform, func(t *testing.T) {
			q := &streamRunner{}
			var code int
			stderr := captureStderr(t, func() {
				code = dispatch([]string{"logs", "--platform", platform, "--container", "solmq-conn", "--previous"}, q)
			})
			if code != 2 {
				t.Fatalf("exit=%d, want 2", code)
			}
			for _, w := range []string{"--previous", "kubernetes", platform} {
				if !strings.Contains(stderr, w) {
					t.Errorf("stderr should name %q, got:\n%s", w, stderr)
				}
			}
			if len(q.calls) != 0 {
				t.Errorf("nothing may run, got %d calls: %+v", len(q.calls), q.calls)
			}
		})
	}
}

// TestLogsPreviousReachesTheKubernetesArgv is the other half of the rule: where
// the concept exists, the flag arrives as kubectl's own -p.
func TestLogsPreviousReachesTheKubernetesArgv(t *testing.T) {
	q := &streamRunner{resp: []queuedResp{{"", nil}, {onePodJSON, nil}}}
	if code := dispatch([]string{"logs", "--platform", "kubernetes", "--pod", "pod-a", "--previous"}, q); code != 0 {
		t.Fatalf("exit=%d, want 0", code)
	}
	if !containsToken(q.calls[1].argv, "-p") {
		t.Errorf("--previous should reach the argv as -p, got %v", q.calls[1].argv)
	}
}

// TestLogsFollowReadsTheOneInstance is the accepted case, pinning that -f
// reaches the argv and that a clean end is exit 0.
func TestLogsFollowReadsTheOneInstance(t *testing.T) {
	q := &streamRunner{
		resp:    []queuedResp{{"", nil}},
		streams: []streamed{{stdout: "tailing\n"}},
	}
	var code int
	stdout := captureStdout(t, func() {
		code = dispatch([]string{"logs", "--platform", "docker", "--container", "solmq-conn", "--follow"}, q)
	})
	if code != 0 {
		t.Fatalf("exit=%d, want 0", code)
	}
	want := []string{"docker", "logs", "-f", "solmq-conn"}
	if !reflect.DeepEqual(q.calls[1].argv, want) {
		t.Errorf("log argv = %v, want %v", q.calls[1].argv, want)
	}
	if stdout != "tailing\n" {
		t.Errorf("stdout = %q, want the streamed log", stdout)
	}
}

// ---- nothing unsafe reaches an argv ----

// TestLogsRejectsUnsafeNamesBeforeAnyCall pins that every operator-supplied name
// is checked before a process starts -- the preflight probe included, so a
// rejected name cannot even be observed by the platform.
func TestLogsRejectsUnsafeNamesBeforeAnyCall(t *testing.T) {
	for _, c := range []struct {
		name string
		args []string
	}{
		{"pod", []string{"logs", "--platform", "kubernetes", "--pod", "pod-a;rm -rf /"}},
		{"container", []string{"logs", "--platform", "docker", "--container", "solmq|nc"}},
		{"namespace", []string{"logs", "--platform", "kubernetes", "--pod", "pod-a", "--namespace", "prod$(id)"}},
	} {
		t.Run(c.name, func(t *testing.T) {
			q := &streamRunner{}
			var code int
			stderr := captureStderr(t, func() { code = dispatch(c.args, q) })
			if code != 1 {
				t.Fatalf("exit=%d, want 1", code)
			}
			if !strings.Contains(stderr, "unsafe character") {
				t.Errorf("stderr should say why, got:\n%s", stderr)
			}
			if len(q.calls) != 0 {
				t.Errorf("nothing may run, got %d calls: %+v", len(q.calls), q.calls)
			}
		})
	}
}

// TestLogsSinceAndTailAreValidatedAtParse covers the two flags that take a value
// this tool then puts in an argv. --since never travels as the operator typed it
// (it is parsed and re-spelled), and --tail is bounded, so neither can carry a
// surprise onward.
func TestLogsSinceAndTailAreValidatedAtParse(t *testing.T) {
	for _, c := range []struct{ name, flag, value string }{
		{"since not a duration", "--since", "yesterday"},
		{"since not positive", "--since", "-5m"},
		{"since with a metacharacter", "--since", "10m;id"},
		{"tail not a number", "--tail", "lots"},
		{"tail above the ceiling", "--tail", "9999999"},
		{"tail negative", "--tail", "-3"},
	} {
		t.Run(c.name, func(t *testing.T) {
			q := &streamRunner{}
			var code int
			captureStderr(t, func() {
				code = dispatch([]string{"logs", "--platform", "docker", "--container", "solmq-conn", c.flag, c.value}, q)
			})
			if code != 2 {
				t.Fatalf("exit=%d, want 2", code)
			}
			if len(q.calls) != 0 {
				t.Errorf("nothing may run, got %d calls: %+v", len(q.calls), q.calls)
			}
		})
	}
}

// ---- reporting across several instances ----

// TestLogsUnexpectedPositionalArgument pins that logs has no target word: an
// instance is named with a flag, so a bare word is a mistake worth naming rather
// than a name to guess at.
func TestLogsUnexpectedPositionalArgument(t *testing.T) {
	q := &streamRunner{}
	var code int
	stderr := captureStderr(t, func() { code = dispatch([]string{"logs", "pod-a"}, q) })
	if code != 2 {
		t.Fatalf("exit=%d, want 2", code)
	}
	for _, w := range []string{"pod-a", "--pod", "--container"} {
		if !strings.Contains(stderr, w) {
			t.Errorf("stderr should name %q, got:\n%s", w, stderr)
		}
	}
	if len(q.calls) != 0 {
		t.Errorf("nothing may run, got %d calls: %+v", len(q.calls), q.calls)
	}
}

// ---- platform resolution ----

// TestLogsPlatformMenuWhenSeveralSectionsArePresent covers the case --platform
// exists for: an env.yaml describing more than one platform cannot be resolved on
// its own, so the operator is asked. The answer then decides which binary the run
// reaches for and which instances it discovers.
func TestLogsPlatformMenuWhenSeveralSectionsArePresent(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "env.yaml", sharedEnv+kubeEnv+dockerEnv)
	withPromptAnswer(t, "1") // kubernetes, the first section present

	q := &streamRunner{resp: []queuedResp{{"", nil}, {onePodJSON, nil}}}
	var code int
	captureStdout(t, func() {
		captureStderr(t, func() {
			code = dispatch([]string{"logs", "-e", filepath.Join(dir, "env.yaml")}, q)
		})
	})
	if code != 0 {
		t.Fatalf("exit=%d, want 0", code)
	}
	if got := q.calls[0].argv[0]; got != "kubectl" {
		t.Errorf("the menu answer should select kubernetes, got %q", got)
	}
	// Discovery came from the deployment in env.yaml, not from a guess.
	if !containsToken(q.calls[1].argv, "app=solmq-connector") {
		t.Errorf("discovery should use the deployment selector, got %v", q.calls[1].argv)
	}
	if !containsToken(q.calls[1].argv, "solace-connectors") {
		t.Errorf("discovery should use the deployment namespace, got %v", q.calls[1].argv)
	}
}

// TestLogsPlatformFlagSkipsTheMenu is the same env.yaml with the flag given: the
// flag is the first step of the resolution order, so nothing is asked.
func TestLogsPlatformFlagSkipsTheMenu(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "env.yaml", sharedEnv+kubeEnv+dockerEnv)
	// Answering the menu with kubernetes would be wrong; the flag must win
	// before promptLine is ever consulted.
	withPromptAnswer(t, "1")

	q := &streamRunner{resp: []queuedResp{{"", nil}}}
	var code int
	captureStdout(t, func() {
		code = dispatch([]string{"logs", "--platform", "docker", "-e", filepath.Join(dir, "env.yaml")}, q)
	})
	if code != 0 {
		t.Fatalf("exit=%d, want 0", code)
	}
	if got := q.calls[0].argv[0]; got != "docker" {
		t.Errorf("--platform docker should win over the menu, got %q", got)
	}
	// The container name still comes from the docker: section.
	if got := q.calls[1].argv; got[len(got)-1] != "solmq-conn" {
		t.Errorf("the instance should come from the docker section, got %v", got)
	}
}

// TestLogsWithoutEnvFileNeedsAnExplicitPlatform covers the explicit-target
// exception: when the operator names the instance and the platform, there is
// nothing left to read from env.yaml, so a missing file is fine. That is how an
// instance this tool never deployed is reached.
func TestLogsWithoutEnvFileNeedsAnExplicitPlatform(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "no-such-env.yaml")

	q := &streamRunner{resp: []queuedResp{{"", nil}}, streams: []streamed{{stdout: "x\n"}}}
	if code := dispatch([]string{"logs", "--platform", "docker", "--container", "foreign", "-e", missing}, q); code != 0 {
		t.Fatalf("exit=%d, want 0", code)
	}
	if got := q.calls[1].argv; got[len(got)-1] != "foreign" {
		t.Errorf("log argv should name the container given, got %v", got)
	}

	// Without --platform there is nothing to fall back on, and the missing file
	// is reported rather than shrugged off.
	q2 := &streamRunner{}
	var code int
	stderr := captureStderr(t, func() {
		code = dispatch([]string{"logs", "--container", "foreign", "-e", missing}, q2)
	})
	if code != 1 {
		t.Fatalf("exit=%d, want 1", code)
	}
	if !strings.Contains(stderr, "no-such-env.yaml") {
		t.Errorf("the error should name the file it could not read, got:\n%s", stderr)
	}
	if len(q2.calls) != 0 {
		t.Errorf("nothing may run, got %d calls: %+v", len(q2.calls), q2.calls)
	}
}

// TestLogsNoInstancesFoundNamesTheFix covers discovery finding nothing: the
// message has to carry the selector and namespace it used, since those are what
// the operator would change.
func TestLogsNoInstancesFoundNamesTheFix(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "env.yaml", sharedEnv+kubeEnv)

	q := &streamRunner{resp: []queuedResp{{"", nil}, {`{"items":[]}`, nil}}}
	var code int
	stderr := captureStderr(t, func() {
		code = dispatch([]string{"logs", "-e", filepath.Join(dir, "env.yaml")}, q)
	})
	if code != 1 {
		t.Fatalf("exit=%d, want 1", code)
	}
	for _, w := range []string{"app=solmq-connector", "solace-connectors", "--pod"} {
		if !strings.Contains(stderr, w) {
			t.Errorf("the error should carry %q, got:\n%s", w, stderr)
		}
	}
}

// TestLogsNeedsAStreamingRunner pins the seam assertion itself: logs reads every
// log through runner.Streamer, and a Runner without it fails loudly rather than
// silently falling back to a buffered read that would merge the platform's
// diagnostics into the log.
func TestLogsNeedsAStreamingRunner(t *testing.T) {
	f := &fakeRunner{} // Run only -- no Stream
	var code int
	stderr := captureStderr(t, func() {
		code = dispatch([]string{"logs", "--platform", "docker", "--container", "solmq-conn"}, f)
	})
	if code != 1 {
		t.Fatalf("exit=%d, want 1", code)
	}
	if !strings.Contains(stderr, "stream") {
		t.Errorf("the error should say what is missing, got:\n%s", stderr)
	}
	if len(f.calls) != 0 {
		t.Errorf("nothing may run, got %d calls: %+v", len(f.calls), f.calls)
	}
}

// TestLogsPlatformWithNoSectionAtAllIsLoud is the earlier failure: an env.yaml
// that parses but describes no platform cannot answer --platform either, and
// says so before discovery is even attempted.
func TestLogsPlatformWithNoSectionAtAllIsLoud(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "env.yaml", sharedEnv)

	q := &streamRunner{}
	var code int
	stderr := captureStderr(t, func() {
		code = dispatch([]string{"logs", "--platform", "docker", "-e", filepath.Join(dir, "env.yaml")}, q)
	})
	if code != 1 {
		t.Fatalf("exit=%d, want 1", code)
	}
	for _, w := range []string{"kubernetes:", "docker:", "podman:"} {
		if !strings.Contains(stderr, w) {
			t.Errorf("the error should name the sections it looked for, got:\n%s", stderr)
		}
	}
	if len(q.calls) != 0 {
		t.Errorf("nothing may run, got %d calls: %+v", len(q.calls), q.calls)
	}
}

// TestLogsNamedPodIsNotPreChecked pins what naming a pod outright now costs,
// and what it no longer does. A name is taken at its word: there is no
// `get pods <name>` first, because the only thing that lookup ever established
// was which container to read, and that is a constant now. A pod that is not
// there is reported by the platform on the read itself, rather than by a probe
// this tool paid for on every run that named one.
func TestLogsNamedPodIsNotPreChecked(t *testing.T) {
	q := &streamRunner{
		resp:    []queuedResp{{"", nil}},
		streams: []streamed{{stderr: "Error from server (NotFound): pods \"ghost\" not found\n", err: fmt.Errorf("exit status 1")}},
	}
	var code int
	stderr := captureStderr(t, func() {
		code = dispatch([]string{"logs", "--platform", "kubernetes", "--pod", "ghost", "--namespace", "prod"}, q)
	})
	if code != 1 {
		t.Fatalf("exit=%d, want 1", code)
	}
	if len(q.calls) != 2 {
		t.Fatalf("want 2 calls (preflight, the read), got %d: %+v", len(q.calls), q.calls)
	}
	for _, c := range q.calls {
		if containsToken(c.argv, "get") {
			t.Errorf("a named pod must not cost a discovery query, got %v", c.argv)
		}
	}
	if !strings.Contains(stderr, "ghost") {
		t.Errorf("the failure should name the instance, got:\n%s", stderr)
	}
}

// TestLogsUsesTheSectionCommandOverride pins that logs reaches for the binary
// env.yaml names, not just the platform default -- the same resolution status
// uses, now shared through instanceCommand.
func TestLogsUsesTheSectionCommandOverride(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "env.yaml", sharedEnv+"\nkubernetes:\n  command: oc\n  deployment:\n    name: solmq-connector\n    namespace: prod\n")

	q := &streamRunner{resp: []queuedResp{{"", nil}, {onePodJSON, nil}}}
	var code int
	captureStdout(t, func() {
		captureStderr(t, func() {
			code = dispatch([]string{"logs", "-e", filepath.Join(dir, "env.yaml")}, q)
		})
	})
	if code != 0 {
		t.Fatalf("exit=%d, want 0", code)
	}
	for i, c := range q.calls {
		if c.argv[0] != "oc" {
			t.Errorf("call %d should use the section command oc, got %v", i, c.argv)
		}
	}
}

// TestLogsCommandFlagOverridesTheSection is the other half: --command wins over
// env.yaml, and still goes through the binary allowlist.
func TestLogsCommandFlagOverridesTheSection(t *testing.T) {
	q := &streamRunner{resp: []queuedResp{{"", nil}, {onePodJSON, nil}}}
	if code := dispatch([]string{"logs", "--platform", "kubernetes", "--pod", "pod-a", "--command", "oc"}, q); code != 0 {
		t.Fatalf("exit=%d, want 0", code)
	}
	if q.calls[0].argv[0] != "oc" {
		t.Errorf("--command should pick the binary, got %v", q.calls[0].argv)
	}

	// A binary outside the allowlist is refused before anything runs, exactly as
	// it is for deploy/remove/status.
	q2 := &streamRunner{}
	var code int
	stderr := captureStderr(t, func() {
		code = dispatch([]string{"logs", "--platform", "kubernetes", "--pod", "pod-a", "--command", "curl"}, q2)
	})
	if code != 1 {
		t.Fatalf("exit=%d, want 1", code)
	}
	if !strings.Contains(stderr, "curl") {
		t.Errorf("the error should name the refused binary, got:\n%s", stderr)
	}
	if len(q2.calls) != 0 {
		t.Errorf("nothing may run, got %d calls: %+v", len(q2.calls), q2.calls)
	}
}

// ---- one instance per run: the picker, and index selection ----

// TestLogsKubernetesArgvShape pins the kubernetes read: a named pod costs no
// discovery at all, and the log is read with the namespace and the connector
// container both named explicitly, so a pod carrying a sidecar cannot have the
// wrong half of it read.
func TestLogsKubernetesArgvShape(t *testing.T) {
	q := &streamRunner{
		resp:    []queuedResp{{"", nil}},
		streams: []streamed{{stdout: "line-a\n"}},
	}
	var code int
	stdout := captureStdout(t, func() {
		code = dispatch([]string{"logs", "--platform", "kubernetes", "--namespace", "prod", "--pod", "pod-a"}, q)
	})
	if code != 0 {
		t.Fatalf("exit=%d, want 0", code)
	}
	if len(q.calls) != 2 {
		t.Fatalf("want 2 calls (preflight, the log read), got %d: %+v", len(q.calls), q.calls)
	}
	if want := []string{"kubectl", "logs", "pod-a", "-n", "prod", "-c", spec.ConnectorContainerName}; !reflect.DeepEqual(q.calls[1].argv, want) {
		t.Errorf("log argv = %v, want %v", q.calls[1].argv, want)
	}
	// One instance prints its log and nothing else, so the output pipes cleanly.
	if stdout != "line-a\n" {
		t.Errorf("stdout should be the log alone, got:\n%s", stdout)
	}
}

// TestLogsRejectsImpossibleFlagCombinations covers the combinations refused
// before anything runs. --pod/--container given twice is among them: logs reads
// one instance, so a second is a question it cannot answer, and losing the first
// silently would be worse than saying so.
func TestLogsRejectsImpossibleFlagCombinations(t *testing.T) {
	for _, c := range []struct {
		name string
		args []string
		want []string
	}{
		{
			name: "follow with previous",
			args: []string{"logs", "--platform", "kubernetes", "--follow", "--previous"},
			want: []string{"--previous", "--follow"},
		},
		{
			name: "two pods",
			args: []string{"logs", "--platform", "kubernetes", "--pod", "a", "--pod", "b"},
			want: []string{"--pod", "once"},
		},
		{
			name: "two containers",
			args: []string{"logs", "--platform", "docker", "--container", "a", "--container", "b"},
			want: []string{"--container", "once"},
		},
	} {
		t.Run(c.name, func(t *testing.T) {
			q := &streamRunner{}
			var code int
			stderr := captureStderr(t, func() { code = dispatch(c.args, q) })
			if code != 2 {
				t.Fatalf("exit=%d, want 2", code)
			}
			for _, w := range c.want {
				if !strings.Contains(stderr, w) {
					t.Errorf("the refusal should name %q, got:\n%s", w, stderr)
				}
			}
			if len(q.calls) != 0 {
				t.Errorf("nothing may run, got %d calls: %+v", len(q.calls), q.calls)
			}
		})
	}
}

// TestLogsAllIsNotALogsFlag pins the removal: --all searches by image and
// returns many, which is the one thing logs cannot do, so it is an unknown flag
// here rather than one that is accepted and then refused.
func TestLogsAllIsNotALogsFlag(t *testing.T) {
	q := &streamRunner{}
	var code int
	captureStderr(t, func() {
		code = dispatch([]string{"logs", "--platform", "kubernetes", "--all"}, q)
	})
	if code != 2 {
		t.Fatalf("exit=%d, want 2", code)
	}
	if len(q.calls) != 0 {
		t.Errorf("nothing may run, got %d calls", len(q.calls))
	}
}

// TestLogsWithoutASectionToDiscoverFrom covers the discovery branch with nothing
// to work from, reached by naming an instance with the other platform's flag
// (--container on kubernetes, --pod on docker/podman): enough to make the run
// explicit, but leaving no name and no section.
//
// It also pins the verb-aware wording. status offers --all as a way out; logs
// does not have that flag, so naming it here would send an operator to a flag
// their verb does not accept.
func TestLogsWithoutASectionToDiscoverFrom(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "no-such-env.yaml")
	for _, c := range []struct {
		platform, wrongFlag, wrongVal, wantSection, wantFix string
	}{
		{"kubernetes", containerFlagName, "solmq-conn", "no kubernetes: section", podFlagName},
		{"docker", podFlagName, "pod-a", "no docker: section", containerFlagName},
		{"podman", podFlagName, "pod-a", "no podman: section", containerFlagName},
	} {
		t.Run(c.platform, func(t *testing.T) {
			q := &streamRunner{resp: []queuedResp{{"", nil}}}
			var code int
			stderr := captureStderr(t, func() {
				code = dispatch([]string{"logs", "--platform", c.platform, c.wrongFlag, c.wrongVal, "-e", missing}, q)
			})
			if code != 1 {
				t.Fatalf("exit=%d, want 1", code)
			}
			for _, w := range []string{c.wantSection, c.wantFix} {
				if !strings.Contains(stderr, w) {
					t.Errorf("the error should carry %q, got:\n%s", w, stderr)
				}
			}
			if strings.Contains(stderr, allFlagName) {
				t.Errorf("logs has no %s, so the hint must not name it, got:\n%s", allFlagName, stderr)
			}
			if q.streamN != 0 {
				t.Errorf("no log may be read, got %d stream calls", q.streamN)
			}
		})
	}
}

// unsortedPodsJSON lists three pods in an order kubectl would not return them
// in, so anything relying on the sort has to do the sorting itself.
const unsortedPodsJSON = `{"items":[
 {"metadata":{"name":"pod-c","namespace":"prod"},"spec":{"nodeName":"n1","containers":[{"name":"connector","image":"img:1"}]},"status":{"phase":"Running","containerStatuses":[{"name":"connector","image":"img:1","restartCount":0,"state":{"running":{"startedAt":"2026-01-01T00:00:00Z"}}}]}},
 {"metadata":{"name":"pod-a","namespace":"prod"},"spec":{"nodeName":"n1","containers":[{"name":"connector","image":"img:1"}]},"status":{"phase":"Running","containerStatuses":[{"name":"connector","image":"img:1","restartCount":0,"state":{"running":{"startedAt":"2026-01-01T00:00:00Z"}}}]}},
 {"metadata":{"name":"pod-b","namespace":"prod"},"spec":{"nodeName":"n1","containers":[{"name":"connector","image":"img:1"}]},"status":{"phase":"Running","containerStatuses":[{"name":"connector","image":"img:1","restartCount":0,"state":{"running":{"startedAt":"2026-01-01T00:00:00Z"}}}]}}
]}`

// kubeEnvForLogs is a kubernetes section whose deployment name gives discovery a
// selector to find pods with, so a bare logs run has something to discover.
const kubeEnvForLogs = `
kubernetes:
  command: kubectl
  deployment:
    name: solmq-connector
    namespace: prod
`

// TestLogsPickerListsPasteableCommands is the whole point of reading one
// instance: when discovery finds several and none was named, nothing is read and
// the operator is handed exactly what to type next.
func TestLogsPickerListsPasteableCommands(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "env.yaml", sharedEnv+kubeEnvForLogs)
	envPath := filepath.Join(dir, "env.yaml")

	q := &streamRunner{resp: []queuedResp{{"", nil}, {unsortedPodsJSON, nil}}}
	var code int
	stdout := captureStdout(t, func() {
		code = dispatch([]string{"logs", "-e", envPath, "--tail", "50", "--follow"}, q)
	})
	// The run answered an ambiguous request with what to type next, which is
	// not a failure.
	if code != 0 {
		t.Fatalf("exit=%d, want 0", code)
	}
	if q.streamN != 0 {
		t.Fatalf("the picker must read no log, got %d stream calls", q.streamN)
	}
	// Listed in sorted order, one pasteable command each, carrying the flags
	// already typed and the resolved platform so a paste cannot re-open the menu.
	for _, want := range []string{
		"3 pods match",
		"solmq-conn-util logs --platform kubernetes -e " + envPath + " --follow --tail 50 --pod pod-a",
		"--pod pod-b",
		"--pod pod-c",
		"or by index, 0-2",
		"--pod 0",
	} {
		if !strings.Contains(stdout, want) {
			t.Errorf("the picker is missing %q, got:\n%s", want, stdout)
		}
	}
	if i, j := strings.Index(stdout, "pod-a"), strings.Index(stdout, "pod-b"); i > j {
		t.Errorf("the picker must list in sorted order, got:\n%s", stdout)
	}
}

// TestLogsOneDiscoveredInstanceIsJustRead pins that the picker is for ambiguity
// only: a single match needs no choosing, so it is read directly.
func TestLogsOneDiscoveredInstanceIsJustRead(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "env.yaml", sharedEnv+kubeEnvForLogs)

	q := &streamRunner{
		resp:    []queuedResp{{"", nil}, {onePodJSON, nil}},
		streams: []streamed{{stdout: "only-line\n"}},
	}
	var code int
	stdout := captureStdout(t, func() {
		code = dispatch([]string{"logs", "-e", filepath.Join(dir, "env.yaml")}, q)
	})
	if code != 0 {
		t.Fatalf("exit=%d, want 0", code)
	}
	if stdout != "only-line\n" {
		t.Errorf("a single match should print its log alone, got:\n%s", stdout)
	}
}

// TestLogsIndexSelectsFromTheSortedList is why the sort is load-bearing: the
// pod list comes back unsorted, and index 0 must still be the instance an
// operator counts to in the picker and in the status table.
func TestLogsIndexSelectsFromTheSortedList(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "env.yaml", sharedEnv+kubeEnvForLogs)

	for _, c := range []struct{ index, want string }{
		{"0", "pod-a"},
		{"1", "pod-b"},
		{"2", "pod-c"},
	} {
		t.Run(c.index, func(t *testing.T) {
			q := &streamRunner{
				resp:    []queuedResp{{"", nil}, {unsortedPodsJSON, nil}},
				streams: []streamed{{stdout: "x\n"}},
			}
			if code := dispatch([]string{"logs", "-e", filepath.Join(dir, "env.yaml"), "--pod", c.index}, q); code != 0 {
				t.Fatalf("exit=%d, want 0", code)
			}
			read := q.calls[len(q.calls)-1].argv
			if !containsToken(read, c.want) {
				t.Errorf("--pod %s should read %s, got %v", c.index, c.want, read)
			}
		})
	}
}

// TestLogsNameWinsOverIndex covers the rule that keeps indexes from stealing a
// legitimate name: a pod genuinely called "0" is reachable by name, and the
// index reading only applies when nothing is actually named that.
func TestLogsNameWinsOverIndex(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "env.yaml", sharedEnv+kubeEnvForLogs)

	// Two pods, the second literally named "0".
	podsNamedZero := `{"items":[
 {"metadata":{"name":"pod-a","namespace":"prod"},"spec":{"containers":[{"name":"connector","image":"img:1"}]},"status":{"phase":"Running","containerStatuses":[{"name":"connector","image":"img:1","restartCount":0,"state":{"running":{"startedAt":"2026-01-01T00:00:00Z"}}}]}},
 {"metadata":{"name":"0","namespace":"prod"},"spec":{"containers":[{"name":"connector","image":"img:1"}]},"status":{"phase":"Running","containerStatuses":[{"name":"connector","image":"img:1","restartCount":0,"state":{"running":{"startedAt":"2026-01-01T00:00:00Z"}}}]}}
]}`
	q := &streamRunner{
		resp:    []queuedResp{{"", nil}, {podsNamedZero, nil}},
		streams: []streamed{{stdout: "x\n"}},
	}
	if code := dispatch([]string{"logs", "-e", filepath.Join(dir, "env.yaml"), "--pod", "0"}, q); code != 0 {
		t.Fatalf("exit=%d, want 0", code)
	}
	// Sorted, "0" is index 0 anyway -- so assert the name matched rather than
	// the position: the read must be for the pod called "0".
	read := q.calls[len(q.calls)-1].argv
	if !containsToken(read, "0") {
		t.Errorf("the pod named 0 should be read, got %v", read)
	}
}

// TestLogsIndexOutOfRangeNamesTheRange keeps the error actionable: the operator
// needs to know how many there were, not just that the number was wrong.
func TestLogsIndexOutOfRangeNamesTheRange(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "env.yaml", sharedEnv+kubeEnvForLogs)

	q := &streamRunner{resp: []queuedResp{{"", nil}, {unsortedPodsJSON, nil}}}
	var code int
	stderr := captureStderr(t, func() {
		code = dispatch([]string{"logs", "-e", filepath.Join(dir, "env.yaml"), "--pod", "3"}, q)
	})
	if code != 1 {
		t.Fatalf("exit=%d, want 1", code)
	}
	for _, want := range []string{"3 pods match", "0-2"} {
		if !strings.Contains(stderr, want) {
			t.Errorf("the error should carry %q, got:\n%s", want, stderr)
		}
	}
	if q.streamN != 0 {
		t.Errorf("no log may be read, got %d stream calls", q.streamN)
	}
}

// TestLogsIndexWithNothingToIndexInto covers an index given where discovery has
// no list to offer: the message has to say that is the problem, rather than
// reporting a missing pod named "0".
func TestLogsIndexWithNothingToIndexInto(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "no-such-env.yaml")
	q := &streamRunner{resp: []queuedResp{{"", nil}}}
	var code int
	stderr := captureStderr(t, func() {
		code = dispatch([]string{"logs", "--platform", "kubernetes", "-e", missing, "--pod", "0"}, q)
	})
	if code != 1 {
		t.Fatalf("exit=%d, want 1", code)
	}
	if !strings.Contains(stderr, "index") {
		t.Errorf("the error should say an index needs a list, got:\n%s", stderr)
	}
	if q.streamN != 0 {
		t.Errorf("no log may be read, got %d stream calls", q.streamN)
	}
}

// TestStatusAcceptsAnIndexToo pins that the two verbs share one vocabulary: the
// same --pod 1 means the same instance, resolved to a real name before status
// queries anything.
func TestStatusAcceptsAnIndexToo(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "env.yaml", sharedEnv+kubeEnvForLogs)

	q := &queueRunner{resp: []queuedResp{
		{"", nil},               // preflight
		{unsortedPodsJSON, nil}, // enumeration for the index
		{unsortedPodsJSON, nil}, // the report query itself
	}}
	if code := dispatch([]string{"status", "container", "-e", filepath.Join(dir, "env.yaml"), "--pod", "1"}, q); code != 0 {
		t.Fatalf("exit=%d, want 0", code)
	}
	// calls: 0 preflight, 1 the enumeration for the index, 2 the report query.
	// Later calls are the workload summary, which is not what this pins.
	report := q.calls[2].argv
	if !containsToken(report, "pod-b") {
		t.Errorf("--pod 1 should resolve to pod-b, got %v", report)
	}
	if containsToken(report, "1") {
		t.Errorf("the index must not reach the argv, got %v", report)
	}
}

// TestStatusKeepsPodRepeatable is the other half of the arity split: status
// reports many instances by design, so the one-entry limit is a logs rule only.
func TestStatusKeepsPodRepeatable(t *testing.T) {
	q := &queueRunner{resp: []queuedResp{{"", nil}, {podsJSON, nil}}}
	if code := dispatch([]string{"status", "container", "--platform", "kubernetes", "--pod", "pod-a", "--pod", "pod-b"}, q); code != 0 {
		t.Fatalf("exit=%d, want 0", code)
	}
	argv := q.calls[1].argv
	for _, want := range []string{"pod-a", "pod-b"} {
		if !containsToken(argv, want) {
			t.Errorf("both pods should be queried, got %v", argv)
		}
	}
}

// ---- the pieces the picker and indexes are built from ----

// TestIsIndex pins what counts as an index. Only a bare run of digits does: a
// name that merely starts with one is a name, or an instance called 0-abc could
// never be reached.
func TestIsIndex(t *testing.T) {
	for _, c := range []struct {
		in   string
		want bool
	}{
		{"0", true},
		{"12", true},
		{"", false},
		{"0-abc", false},
		{"abc", false},
		{"-1", false},
		{"1.0", false},
		{" 1", false},
	} {
		if got := isIndex(c.in); got != c.want {
			t.Errorf("isIndex(%q) = %v, want %v", c.in, got, c.want)
		}
	}
}

// TestResolveOneIndexRules covers the three outcomes in one place: an exact name
// always wins, a digit run in range selects positionally, and one past the end
// says how many there were.
func TestResolveOneIndexRules(t *testing.T) {
	refs := []instanceRef{{name: "alpha"}, {name: "0"}, {name: "zeta"}}
	for _, c := range []struct {
		name    string
		in      string
		want    string
		wantErr bool
	}{
		{"a plain name passes through", "alpha", "alpha", false},
		{"a name that is digits wins over the index", "0", "0", false},
		{"an index selects positionally", "2", "zeta", false},
		{"an unknown name passes through for discovery to judge", "ghost", "ghost", false},
		{"past the end is an error", "3", "", true},
	} {
		t.Run(c.name, func(t *testing.T) {
			got, err := resolveOneIndex(c.in, refs, podFlagName, validate.PlatformKubernetes)
			if c.wantErr {
				if err == nil {
					t.Fatalf("want an error for %q", c.in)
				}
				for _, w := range []string{"3 pods match", "0-2"} {
					if !strings.Contains(err.Error(), w) {
						t.Errorf("the error should carry %q, got %v", w, err)
					}
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != c.want {
				t.Errorf("resolveOneIndex(%q) = %q, want %q", c.in, got, c.want)
			}
		})
	}
}

// TestSortIsByNameThenNamespace pins the order an index selects into. The
// namespace tie-break matters for a cross-namespace status --all, where two
// deployments can run pods of the same name.
func TestSortIsByNameThenNamespace(t *testing.T) {
	refs := []instanceRef{
		{name: "b", namespace: "prod"},
		{name: "a", namespace: "prod"},
		{name: "a", namespace: "dev"},
	}
	sortRefs(refs)
	want := []string{"a/dev", "a/prod", "b/prod"}
	for i, w := range want {
		if got := refs[i].name + "/" + refs[i].namespace; got != w {
			t.Errorf("refs[%d] = %s, want %s", i, got, w)
		}
	}

	insts := []statusreport.Instance{
		{Name: "b", Namespace: "prod"},
		{Name: "a", Namespace: "prod"},
		{Name: "a", Namespace: "dev"},
	}
	sortInstances(insts)
	for i, w := range want {
		if got := insts[i].Name + "/" + insts[i].Namespace; got != w {
			t.Errorf("insts[%d] = %s, want %s", i, got, w)
		}
	}
}

// TestNamingHintIsVerbAware pins that a hint never names a flag the verb it came
// from does not accept: status offers --all as a way out, logs has no such flag.
func TestNamingHintIsVerbAware(t *testing.T) {
	for _, c := range []struct {
		verb, platform string
		wantFlag       string
		wantAll        bool
	}{
		{verbStatus, validate.PlatformKubernetes, podFlagName, true},
		{verbStatus, validate.PlatformDocker, containerFlagName, true},
		{verbLogs, validate.PlatformKubernetes, podFlagName, false},
		{verbLogs, validate.PlatformPodman, containerFlagName, false},
		{verbShell, validate.PlatformKubernetes, podFlagName, false},
		{verbShell, validate.PlatformDocker, containerFlagName, false},
	} {
		t.Run(c.verb+"/"+c.platform, func(t *testing.T) {
			got := namingHint(instanceSession{verb: c.verb, platform: c.platform})
			if !strings.Contains(got, c.wantFlag) {
				t.Errorf("hint should name %s, got %q", c.wantFlag, got)
			}
			if strings.Contains(got, allFlagName) != c.wantAll {
				t.Errorf("hint mentions %s = %v, want %v: %q", allFlagName, !c.wantAll, c.wantAll, got)
			}
		})
	}
}

// TestLogsPickerCarriesEveryFlagBack is what makes the picker worth printing:
// every flag already typed has to come back in the suggested command, or pasting
// one silently drops what the operator asked for.
func TestLogsPickerCarriesEveryFlagBack(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "env.yaml", sharedEnv+kubeEnvForLogs)
	envPath := filepath.Join(dir, "env.yaml")

	q := &streamRunner{resp: []queuedResp{{"", nil}, {unsortedPodsJSON, nil}}}
	var code int
	stdout := captureStdout(t, func() {
		code = dispatch([]string{
			"logs", "-e", envPath,
			"--namespace", "prod",
			"--command", "oc",
			"--allow-command", "sudo",
			"--previous",
			"--timestamps",
			"--since", "10m",
			"--tail", "50",
		}, q)
	})
	if code != 0 {
		t.Fatalf("exit=%d, want 0", code)
	}
	for _, want := range []string{
		"--platform kubernetes",
		"-e " + envPath,
		"--namespace prod",
		"--command oc",
		"--allow-command sudo",
		"--previous",
		"--timestamps",
		"--tail 50",
		"--since 10m0s",
	} {
		if !strings.Contains(stdout, want) {
			t.Errorf("the suggested command is missing %q, got:\n%s", want, stdout)
		}
	}
}

// TestInstanceCommandResolution covers the three-step binary resolution both
// verbs share: the --command override, then the section command in env.yaml,
// then the platform default. The defaults matter because a run can happen with
// no section at all (explicit targets), where parse-time defaulting never ran.
func TestInstanceCommandResolution(t *testing.T) {
	withKube := &spec.Env{Kubernetes: &spec.Kubernetes{Command: "oc"}}
	withDocker := &spec.Env{Docker: &spec.Docker{Command: "podman-compat"}}
	withPodman := &spec.Env{Podman: &spec.Podman{Command: "sudo podman"}}

	for _, c := range []struct {
		name     string
		platform string
		override string
		env      *spec.Env
		want     string
	}{
		{"override wins", validate.PlatformKubernetes, "kubectl --context prod", withKube, "kubectl --context prod"},
		{"kube section", validate.PlatformKubernetes, "", withKube, "oc"},
		{"kube default", validate.PlatformKubernetes, "", &spec.Env{}, spec.DefaultKubeCommand},
		{"kube nil env", validate.PlatformKubernetes, "", nil, spec.DefaultKubeCommand},
		{"docker section", validate.PlatformDocker, "", withDocker, "podman-compat"},
		{"docker default", validate.PlatformDocker, "", &spec.Env{}, spec.DefaultDockerCommand},
		{"podman section", validate.PlatformPodman, "", withPodman, "sudo podman"},
		{"podman default", validate.PlatformPodman, "", &spec.Env{}, spec.DefaultPodmanCommand},
		{"unknown platform", "nonsense", "", &spec.Env{}, ""},
	} {
		t.Run(c.name, func(t *testing.T) {
			if got := instanceCommand(c.platform, c.override, c.env); got != c.want {
				t.Errorf("instanceCommand(%q, %q) = %q, want %q", c.platform, c.override, got, c.want)
			}
		})
	}
}

// TestInstanceNamespaceResolution is the same three steps for the namespace,
// including that it stays empty on docker/podman, which have no such concept.
func TestInstanceNamespaceResolution(t *testing.T) {
	e := &spec.Env{Kubernetes: &spec.Kubernetes{}}
	e.Kubernetes.Deployment.Namespace = "prod"

	for _, c := range []struct {
		name     string
		platform string
		override string
		env      *spec.Env
		want     string
	}{
		{"override wins", validate.PlatformKubernetes, "other", e, "other"},
		{"from the section", validate.PlatformKubernetes, "", e, "prod"},
		{"no section", validate.PlatformKubernetes, "", &spec.Env{}, ""},
		{"nil env", validate.PlatformKubernetes, "", nil, ""},
		{"docker has none", validate.PlatformDocker, "", e, ""},
		{"podman has none", validate.PlatformPodman, "", e, ""},
	} {
		t.Run(c.name, func(t *testing.T) {
			if got := instanceNamespace(c.platform, c.override, c.env); got != c.want {
				t.Errorf("instanceNamespace(%q, %q) = %q, want %q", c.platform, c.override, got, c.want)
			}
		})
	}
}

// TestConfiguredInstanceName covers both engines and both missing-section
// errors, including that the message names the section that is absent rather
// than just reporting that nothing was found.
func TestConfiguredInstanceName(t *testing.T) {
	for _, c := range []struct {
		name    string
		sess    instanceSession
		want    string
		wantErr string
	}{
		{
			name: "docker",
			sess: instanceSession{platform: validate.PlatformDocker, env: &spec.Env{Docker: &spec.Docker{Name: "solmq-dk"}}},
			want: "solmq-dk",
		},
		{
			name: "podman",
			sess: instanceSession{platform: validate.PlatformPodman, env: &spec.Env{Podman: &spec.Podman{Name: "solmq-pm"}}},
			want: "solmq-pm",
		},
		{
			name:    "docker section absent",
			sess:    instanceSession{verb: verbStatus, platform: validate.PlatformDocker, env: &spec.Env{}},
			wantErr: "no docker: section",
		},
		{
			name:    "podman section absent",
			sess:    instanceSession{verb: verbStatus, platform: validate.PlatformPodman, env: &spec.Env{}},
			wantErr: "no podman: section",
		},
		{
			name:    "nil env",
			sess:    instanceSession{verb: verbLogs, platform: validate.PlatformDocker},
			wantErr: "no docker: section",
		},
	} {
		t.Run(c.name, func(t *testing.T) {
			got, err := configuredInstanceName(c.sess)
			if c.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), c.wantErr) {
					t.Fatalf("want an error carrying %q, got %v", c.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != c.want {
				t.Errorf("= %q, want %q", got, c.want)
			}
		})
	}
}

// TestNoPodsFoundExplainsTheBranchItCameFrom pins that an empty result is
// explained in the terms of whichever discovery produced it: a cluster-wide
// image search, a selector, or pods named outright. Quoting a selector back at
// someone who named pods directly would be noise.
func TestNoPodsFoundExplainsTheBranchItCameFrom(t *testing.T) {
	for _, c := range []struct {
		name     string
		selector string
		ns       string
		all      bool
		want     []string
	}{
		{"image search", "", "", true, []string{"anywhere in the cluster", statusreport.ImageMatch}},
		{"selector", "app=solmq", "prod", false, []string{"app=solmq", "prod", podFlagName}},
		{"named pods", "", "prod", false, []string{"none of the named pods", "prod"}},
	} {
		t.Run(c.name, func(t *testing.T) {
			err := noPodsFound(c.selector, c.ns, c.all)
			for _, w := range c.want {
				if !strings.Contains(err.Error(), w) {
					t.Errorf("the error should carry %q, got %v", w, err)
				}
			}
		})
	}
}

// ---- o) version -------------------------------------------------------------------

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

// ---- j) the cli verb: a session inside one instance ----------------------------

// attached is one Attach call's canned result: the status the session ended
// with, and an error only for a session that could not be started at all.
type attached struct {
	code int
	err  error
}

// attachRunner is the fake cli needs: it satisfies runner.Attacher as well as
// runner.Runner, since cli reaches its instance through the attaching seam
// while the preflight probe and discovery still go through Run. Both kinds of
// call are recorded in one ordered slice, while the canned answers are queued
// per kind -- the same split streamRunner uses for logs.
//
// The three *os.File values are ignored outright, and that is exactly why the
// seam takes them as parameters instead of reading os.Stdin itself: the fake
// needs no terminal, no pipe, and no temp file standing in for one.
type attachRunner struct {
	calls []fakeCall

	resp []queuedResp
	runN int

	attaches []attached
	attachN  int
}

func (a *attachRunner) Run(c runner.Cmd) (string, error) {
	a.calls = append(a.calls, fakeCall{argv: c.Argv, stdin: c.Stdin, env: c.Env})
	i := a.runN
	a.runN++
	if i < len(a.resp) {
		return a.resp[i].out, a.resp[i].err
	}
	return "", nil
}

func (a *attachRunner) Attach(c runner.Cmd, _, _, _ *os.File) (int, error) {
	a.calls = append(a.calls, fakeCall{argv: c.Argv, stdin: c.Stdin, env: c.Env})
	i := a.attachN
	a.attachN++
	if i >= len(a.attaches) {
		return 0, nil
	}
	return a.attaches[i].code, a.attaches[i].err
}

// withTerminal drives the stdinIsTerminal seam, mirroring withPromptAnswer and
// existing for the same reason: go test's own stdin is never a character
// device, so every path behind the real probe would be unreachable.
func withTerminal(t *testing.T, yes bool) {
	t.Helper()
	prev := stdinIsTerminal
	stdinIsTerminal = func() bool { return yes }
	t.Cleanup(func() { stdinIsTerminal = prev })
}

// TestCliNeedsAnAttachingRunner pins the type assertion: a Runner that cannot
// hand over the terminal fails loudly, before anything runs, rather than
// silently degrading to a session with no prompt.
func TestCliNeedsAnAttachingRunner(t *testing.T) {
	withTerminal(t, true)
	f := useFakeRunner(t)
	var code int
	stderr := captureStderr(t, func() {
		code = dispatch([]string{"cli", "--platform", "docker", "--container", "solmq-conn"}, f)
	})
	if code != 1 {
		t.Fatalf("exit=%d, want 1", code)
	}
	if !strings.Contains(stderr, "terminal") {
		t.Errorf("the error should say what the runner cannot do, got:\n%s", stderr)
	}
	if len(f.calls) != 0 {
		t.Errorf("nothing may run, got %d calls: %+v", len(f.calls), f.calls)
	}
}

// TestCliRefusesAShellWithoutATerminal is the same non-TTY contract the
// platform menu, the status install prompt and the remove confirmation keep:
// fail with the next step rather than open a session nobody can type into. The
// next step here is the one-shot form, so the message has to name it.
func TestCliRefusesAShellWithoutATerminal(t *testing.T) {
	withTerminal(t, false)
	q := &attachRunner{}
	var code int
	stderr := captureStderr(t, func() {
		code = dispatch([]string{"cli", "--platform", "docker", "--container", "solmq-conn"}, q)
	})
	if code != 1 {
		t.Fatalf("exit=%d, want 1", code)
	}
	if !strings.Contains(stderr, "cli -- <command>") {
		t.Errorf("the refusal should name the one-shot form, got:\n%s", stderr)
	}
	if len(q.calls) != 0 {
		t.Errorf("nothing may run, got %d calls: %+v", len(q.calls), q.calls)
	}
}

// TestCliKubernetesArgvShape pins the session argv: -i and -t both asked for,
// the namespace and the connector container both named, and the "--" that tells
// kubectl where the in-container command starts. Naming a pod costs no
// discovery call, so the whole run is the preflight probe and the attach.
func TestCliKubernetesArgvShape(t *testing.T) {
	withTerminal(t, true)
	q := &attachRunner{}
	if code := dispatch([]string{"cli", "--platform", "kubernetes", "--pod", "pod-a", "--namespace", "prod"}, q); code != 0 {
		t.Fatalf("exit=%d, want 0", code)
	}
	if len(q.calls) != 2 {
		t.Fatalf("want 2 calls (preflight, the attach), got %d: %+v", len(q.calls), q.calls)
	}
	want := []string{"kubectl", "exec", "-i", "-t", "pod-a", "-n", "prod", "-c", spec.ConnectorContainerName, "--", runner.ContainerShell}
	if !reflect.DeepEqual(q.calls[1].argv, want) {
		t.Errorf("session argv = %v, want %v", q.calls[1].argv, want)
	}
}

// TestCliEngineArgvShape is the other half: docker and podman stop parsing
// their own flags at the container name, so the options come first and there is
// no "--" and no -c to add.
func TestCliEngineArgvShape(t *testing.T) {
	for _, platform := range []string{"docker", "podman"} {
		t.Run(platform, func(t *testing.T) {
			withTerminal(t, true)
			q := &attachRunner{}
			if code := dispatch([]string{"cli", "--platform", platform, "--container", "solmq-conn"}, q); code != 0 {
				t.Fatalf("exit=%d, want 0", code)
			}
			if len(q.calls) != 2 {
				t.Fatalf("want 2 calls (preflight, the attach), got %d: %+v", len(q.calls), q.calls)
			}
			want := []string{platform, "exec", "-i", "-t", "solmq-conn", runner.ContainerShell}
			if !reflect.DeepEqual(q.calls[1].argv, want) {
				t.Errorf("session argv = %v, want %v", q.calls[1].argv, want)
			}
		})
	}
}

// TestCliOneShotRunsTheCommandWithNoTerminal pins the form a script uses:
// everything after "--" replaces the shell, no -t is asked for because there is
// no prompt to draw, and no -i either while stdin is a terminal -- a command
// that reads stdin would otherwise wait on one that never sends EOF.
func TestCliOneShotRunsTheCommandWithNoTerminal(t *testing.T) {
	withTerminal(t, true)
	q := &attachRunner{}
	code := dispatch([]string{"cli", "--platform", "docker", "--container", "solmq-conn", "--", "ls", "-la", "/app/external"}, q)
	if code != 0 {
		t.Fatalf("exit=%d, want 0", code)
	}
	want := []string{"docker", "exec", "solmq-conn", "ls", "-la", "/app/external"}
	if !reflect.DeepEqual(q.calls[1].argv, want) {
		t.Errorf("one-shot argv = %v, want %v", q.calls[1].argv, want)
	}
}

// TestCliOneShotAttachesStdinWhenSomethingIsPiped is the other side of that
// rule: a stdin that is not a terminal is a stdin something is being piped
// into, so -i is what makes `echo hi | cli -- cat` work. It is also the case
// that proves the one-shot form is allowed without a terminal at all.
func TestCliOneShotAttachesStdinWhenSomethingIsPiped(t *testing.T) {
	withTerminal(t, false)
	q := &attachRunner{}
	code := dispatch([]string{"cli", "--platform", "docker", "--container", "solmq-conn", "--", "cat"}, q)
	if code != 0 {
		t.Fatalf("exit=%d, want 0", code)
	}
	want := []string{"docker", "exec", "-i", "solmq-conn", "cat"}
	if !reflect.DeepEqual(q.calls[1].argv, want) {
		t.Errorf("one-shot argv = %v, want %v", q.calls[1].argv, want)
	}
}

// TestCliPropagatesTheSessionExitStatus is cli's one departure from the 0/1/2
// contract: once the session ran, its own status is the answer, so `exit 3`
// inside the shell reaches the operator as 3.
func TestCliPropagatesTheSessionExitStatus(t *testing.T) {
	withTerminal(t, true)
	q := &attachRunner{attaches: []attached{{code: 3}}}
	if code := dispatch([]string{"cli", "--platform", "docker", "--container", "solmq-conn"}, q); code != 3 {
		t.Errorf("exit=%d, want the session status 3", code)
	}
}

// TestCliSessionThatCouldNotStartIsAnError separates the two: a session that
// never began is this tool failing, which is exit 1 and names the instance.
func TestCliSessionThatCouldNotStartIsAnError(t *testing.T) {
	withTerminal(t, true)
	q := &attachRunner{attaches: []attached{{err: fmt.Errorf("docker: not found")}}}
	var code int
	stderr := captureStderr(t, func() {
		code = dispatch([]string{"cli", "--platform", "docker", "--container", "solmq-conn"}, q)
	})
	if code != 1 {
		t.Fatalf("exit=%d, want 1", code)
	}
	if !strings.Contains(stderr, "solmq-conn") {
		t.Errorf("the failure should name the instance, got:\n%s", stderr)
	}
}

// TestCliRejectsWhatCannotMeanAnything gathers the refusals that are decided
// from the command line alone. Every one of them must reach no process at all:
// a mistake in the invocation costs nothing.
func TestCliRejectsWhatCannotMeanAnything(t *testing.T) {
	for _, c := range []struct {
		name string
		args []string
		want []string
		code int
	}{
		{
			name: "a bare word is not a target",
			args: []string{"cli", "bogus"},
			want: []string{"unexpected argument", "--pod", "--"},
			code: 2,
		},
		{
			name: "a separator with no command",
			args: []string{"cli", "--platform", "docker", "--container", "c", "--"},
			want: []string{"no command after it"},
			code: 2,
		},
		{
			name: "two pods",
			args: []string{"cli", "--platform", "kubernetes", "--pod", "a", "--pod", "b"},
			want: []string{"--pod may be given once"},
			code: 2,
		},
		{
			name: "two containers",
			args: []string{"cli", "--platform", "docker", "--container", "a", "--container", "b"},
			want: []string{"--container may be given once"},
			code: 2,
		},
		{
			// cli runs an argv and never a shell line, so a glob has nothing to
			// expand it -- which makes one in the command a mistake, and the
			// message has to say where it can be written instead.
			name: "an unsafe command token",
			args: []string{"cli", "--platform", "docker", "--container", "c", "--", "ls", "*"},
			want: []string{"unsafe character", "argv, not a shell line"},
			code: 1,
		},
		{
			name: "an unsafe instance name",
			args: []string{"cli", "--platform", "kubernetes", "--pod", "pod-a;rm -rf /"},
			want: []string{"unsafe character"},
			code: 1,
		},
	} {
		t.Run(c.name, func(t *testing.T) {
			withTerminal(t, true)
			q := &attachRunner{}
			var code int
			stderr := captureStderr(t, func() { code = dispatch(c.args, q) })
			if code != c.code {
				t.Fatalf("exit=%d, want %d; stderr:\n%s", code, c.code, stderr)
			}
			for _, w := range c.want {
				if !strings.Contains(stderr, w) {
					t.Errorf("the refusal should mention %q, got:\n%s", w, stderr)
				}
			}
			if len(q.calls) != 0 {
				t.Errorf("nothing may run, got %d calls: %+v", len(q.calls), q.calls)
			}
		})
	}
}

// TestCliPickerCarriesTheCommandBehindTheSeparator is the picker's cli-specific
// twist: the in-container command has to stay behind its "--", and therefore
// behind the --pod the picker is telling the operator to add. A line that
// pasted the flag after the command would not parse.
func TestCliPickerCarriesTheCommandBehindTheSeparator(t *testing.T) {
	withTerminal(t, true)
	dir := t.TempDir()
	write(t, dir, "env.yaml", sharedEnv+kubeEnvForLogs)
	envPath := filepath.Join(dir, "env.yaml")

	q := &attachRunner{resp: []queuedResp{{"", nil}, {unsortedPodsJSON, nil}}}
	var code int
	stdout := captureStdout(t, func() {
		code = dispatch([]string{"cli", "-e", envPath, "--", "ls", "-la"}, q)
	})
	// An ambiguous request answered with what to type next is not a failure.
	if code != 0 {
		t.Fatalf("exit=%d, want 0", code)
	}
	if q.attachN != 0 {
		t.Fatalf("the picker must open no session, got %d attaches", q.attachN)
	}
	for _, want := range []string{
		"3 pods match; cli takes one",
		"solmq-conn-util cli --platform kubernetes -e " + envPath + " --pod pod-a -- ls -la",
		"--pod pod-b -- ls -la",
		"or by index, 0-2",
		"--pod 0 -- ls -la",
	} {
		if !strings.Contains(stdout, want) {
			t.Errorf("the picker is missing %q, got:\n%s", want, stdout)
		}
	}
}

// TestCliIndexSelectsFromTheSortedList pins that cli indexes into the same
// order logs and status print, so a number copied off one verb means the same
// instance in another.
func TestCliIndexSelectsFromTheSortedList(t *testing.T) {
	withTerminal(t, true)
	dir := t.TempDir()
	write(t, dir, "env.yaml", sharedEnv+kubeEnvForLogs)

	q := &attachRunner{resp: []queuedResp{{"", nil}, {unsortedPodsJSON, nil}}}
	if code := dispatch([]string{"cli", "-e", filepath.Join(dir, "env.yaml"), "--pod", "1"}, q); code != 0 {
		t.Fatalf("exit=%d, want 0", code)
	}
	if len(q.calls) != 3 {
		t.Fatalf("want 3 calls (preflight, get pods, the attach), got %d: %+v", len(q.calls), q.calls)
	}
	// pod-b: the middle of the sorted list, not the middle of the JSON.
	if !containsToken(q.calls[2].argv, "pod-b") {
		t.Errorf("index 1 should select pod-b from the sorted list, got %v", q.calls[2].argv)
	}
}

// TestCliPreflightFailureOpensNothing keeps cli on the same contract as status
// and logs: an engine that cannot be reached is reported once, up front, and no
// session is attempted.
func TestCliPreflightFailureOpensNothing(t *testing.T) {
	withTerminal(t, true)
	q := &attachRunner{resp: []queuedResp{{"cannot connect to the docker daemon", fmt.Errorf("exit status 1")}}}
	var code int
	captureStderr(t, func() {
		code = dispatch([]string{"cli", "--platform", "docker", "--container", "solmq-conn"}, q)
	})
	if code == 0 {
		t.Fatal("a failed preflight must not exit 0")
	}
	if q.attachN != 0 {
		t.Errorf("no session may be opened, got %d attaches", q.attachN)
	}
}
