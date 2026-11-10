package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	pkcs12 "software.sslmate.com/src/go-pkcs12"

	"github.com/hafio/go-deploy-utility/confirm"
	"github.com/hafio/go-deploy-utility/containercli"
	"github.com/hafio/go-deploy-utility/kube/nsgate"
	"github.com/hafio/go-deploy-utility/notice"
	gdurunner "github.com/hafio/go-deploy-utility/runner"
	gduwatch "github.com/hafio/go-deploy-utility/watch"
	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/certfixture"
	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/dockergen"
	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/libs"
	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/runner"
	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/spec"
	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/statusreport"
	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/statusscript"
	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/storescript"
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

// tlsWF is validWF with its MQ source on TLS, validating the queue manager
// against the store named ts -- so the spec mounts that store.
var tlsWF = strings.Replace(validWF, "    queue: Q.IN\n", "    tls: true\n    cipher: TLS_RSA_WITH_AES_256_CBC_SHA256\n    truststore: ts\n    queue: Q.IN\n", 1)

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
// platform-resolution seam tests -- a safe command, the fields checkKube
// requires, and the credentials Secret any referenced credential needs.
const kubeEnv = `
kubernetes:
  command: kubectl
  deployment:
    name: solmq-connector
    namespace: solace-connectors
  secrets:
    credentials:
      create: true
`

// kubeEnvUnsafeCommand mirrors kubeEnv but with a command containing a shell
// metacharacter -- it reaches the CLI and is refused by the shared command
// guard (execguard.Check), at validate and again at runner.ParseCommand.
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

// withPrompter swaps the run's prompter (the seam behind the platform menu, the
// status install confirmation, both remove questions and cli's terminal check)
// for one whose terminal answer and input the test controls, restoring the
// previous one on cleanup. input is every line the run may read, one per
// question: the prompter holds a single reader, so a run that asks twice needs
// two lines. It returns the buffer the questions are written to. It never
// touches the real os.Stdin, so it carries no risk of blocking on a read
// regardless of how the test binary itself was invoked.
func withPrompter(t *testing.T, terminal bool, input string) *bytes.Buffer {
	t.Helper()
	out := &bytes.Buffer{}
	old := prompter
	prompter = confirm.New(confirm.Options{
		IsTerminal: func() bool { return terminal },
		In:         strings.NewReader(input),
		Out:        out,
	})
	t.Cleanup(func() { prompter = old })
	return out
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
	noAttach
	calls    []fakeCall
	err      error
	failFrom int
}

// noAttach is the Attach half of gdurunner.Runner for the fakes that never open
// a session: a call that reaches it is a bug in the test, so it fails loudly.
type noAttach struct{}

func (noAttach) Attach(c gdurunner.Cmd, _, _, _ *os.File) (int, error) {
	return 0, fmt.Errorf("%s: this fake does not attach", c.Argv[0])
}

// scriptOut writes a fake's scripted output to the stream a caller handed it. A nil
// writer discards the stream, as the shared runner's does.
func scriptOut(w io.Writer, s string) {
	if w != nil && s != "" {
		fmt.Fprint(w, s)
	}
}

type fakeCall struct {
	argv  []string
	stdin string
	env   []string
}

func (f *fakeRunner) Run(_ context.Context, c gdurunner.Cmd, stdout, _ io.Writer) error {
	idx := len(f.calls)
	f.calls = append(f.calls, fakeCall{argv: c.Argv, stdin: string(c.Stdin), env: c.Env})
	scriptOut(stdout, defaultAnswer(c.Argv))
	if f.err != nil && idx >= f.failFrom {
		return f.err
	}
	return nil
}

// defaultAnswer is what a fake says to a command a test did not script: "yes" to a
// kubectl auth can-i (rbac.Check refuses anything else), namespace/<ns> to the
// namespace read (an existing namespace, so a deploy has nothing more to ask), and
// "out" otherwise.
func defaultAnswer(argv []string) string {
	for i := 0; i+1 < len(argv); i++ {
		switch {
		case argv[i] == "auth" && argv[i+1] == "can-i":
			return "yes\n"
		case argv[i] == "get" && argv[i+1] == "namespace" && i+2 < len(argv):
			return "namespace/" + argv[i+2] + "\n"
		}
	}
	return "out"
}

// useFakeRunner returns a fake that captures argv instead of starting a
// process. Callers pass it to dispatch (not run, which always uses the
// production gdurunner.Exec{}) so the fake reaches the deploy/remove seam as an
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
// "fail from index N onward". Calls beyond len(resp) succeed with no output,
// except that a kubectl auth can-i still answers "yes" and the namespace read
// still finds the namespace (defaultAnswer).
type queueRunner struct {
	noAttach
	calls []fakeCall
	resp  []queuedResp
}

func (q *queueRunner) Run(_ context.Context, c gdurunner.Cmd, stdout, _ io.Writer) error {
	idx := len(q.calls)
	q.calls = append(q.calls, fakeCall{argv: c.Argv, stdin: string(c.Stdin), env: c.Env})
	if idx < len(q.resp) {
		scriptOut(stdout, q.resp[idx].out)
		return q.resp[idx].err
	}
	scriptOut(stdout, strings.TrimPrefix(defaultAnswer(c.Argv), "out"))
	return nil
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
		{"missing env file", []string{"generate", "config", "-e", filepath.Join(validDir, "does-not-exist.yaml")}, 2},
		{"invalid spec", []string{"generate", "config", "-e", filepath.Join(invalidDir, "env.yaml")}, 2},
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
	if code != 2 {
		t.Fatalf("exit=%d, want 2", code)
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
	// The read-only probes come first, one per kind the manifest writes
	// (deployments, configmaps and the credentials secrets), then the namespace
	// read (it exists, so nothing is created), the apply, and the rollout wait.
	if len(f.calls) != kubeProbes+3 {
		t.Fatalf("want %d calls to the runner (probes, namespace read, apply, rollout status), got %d: %+v", kubeProbes+3, len(f.calls), f.calls)
	}
	for i, resource := range []string{"deployments", "configmaps", "secrets"} {
		want := []string{"kubectl", "auth", "can-i", "create", resource, "-n", "solace-connectors"}
		if !reflect.DeepEqual(f.calls[i].argv, want) {
			t.Errorf("probe %d argv = %v, want %v", i, f.calls[i].argv, want)
		}
	}
	if want := []string{"kubectl", "get", "namespace", "solace-connectors", "--ignore-not-found", "-o", "name"}; !reflect.DeepEqual(f.calls[kubeProbes].argv, want) {
		t.Errorf("namespace read argv = %v, want %v", f.calls[kubeProbes].argv, want)
	}
	apply := f.calls[kubeProbes+1]
	if want := []string{"kubectl", "apply", "-f", "-"}; !reflect.DeepEqual(apply.argv, want) {
		t.Errorf("argv = %v, want %v", apply.argv, want)
	}
	if !strings.Contains(apply.stdin, "kind: Deployment") {
		t.Errorf("manifest on stdin missing kind: Deployment:\n%s", apply.stdin)
	}
	if want := []string{"kubectl", "rollout", "status", "deployment/solmq-connector", "-n", "solace-connectors", "--timeout=300s"}; !reflect.DeepEqual(f.calls[kubeProbes+2].argv, want) {
		t.Errorf("rollout wait argv = %v, want %v", f.calls[kubeProbes+2].argv, want)
	}
}

func TestRemoveKubernetesSeamHappyPath(t *testing.T) {
	// The probes, the manifest delete, the occupancy probe that every kubernetes
	// teardown makes (nsList() is an empty namespace, so the run reaches the
	// namespace question), the permission probe for deleting it, and the delete.
	f := &queueRunner{resp: removeResp(queuedResp{nsList(), nil}, queuedResp{"yes\n", nil}, queuedResp{"", nil})}
	// remove asks before it tears down; this test is about the argv that
	// follows, so it answers yes. The prompt itself is covered below.
	withPrompter(t, true, "y\ny\n")
	dir := t.TempDir()
	write(t, dir, "env.yaml", sharedEnv+kubeEnv)
	write(t, dir, "10.yaml", validWF)

	if code := dispatch([]string{"remove", "--platform", "kubernetes", "-e", filepath.Join(dir, "env.yaml")}, f); code != 0 {
		t.Fatalf("exit=%d", code)
	}
	if len(f.calls) != kubeProbes+4 {
		t.Fatalf("want %d calls (probes, delete, listing, delete probe, namespace delete), got %d: %+v", kubeProbes+4, len(f.calls), f.calls)
	}
	// The teardown asks delete of every kind it deletes, including the stores
	// Secret it always names; the Service is not enabled, so it is not asked.
	for i, resource := range []string{"deployments", "configmaps", "secrets"} {
		want := []string{"kubectl", "auth", "can-i", "delete", resource, "-n", "solace-connectors"}
		if !reflect.DeepEqual(f.calls[i].argv, want) {
			t.Errorf("probe %d argv = %v, want %v", i, f.calls[i].argv, want)
		}
	}
	want := []string{"kubectl", "delete", "-f", "-", "--ignore-not-found"}
	if !reflect.DeepEqual(f.calls[kubeProbes].argv, want) {
		t.Errorf("argv = %v, want %v", f.calls[kubeProbes].argv, want)
	}
}

// kubeProbes is how many permission probes a deploy or remove of kubeEnv asks before
// its first write: deployments, configmaps and secrets (kubeEnv enables no Service
// and no libs claim, and creates the credentials Secret).
const kubeProbes = 3

// yesProbes answers n permission probes with a yes.
func yesProbes(n int) []queuedResp {
	out := make([]queuedResp, n)
	for i := range out {
		out[i] = queuedResp{"yes\n", nil}
	}
	return out
}

// removeResp is the answers a kubernetes remove of kubeEnv reads: its probes, the
// manifest delete (no output), then rest, the namespace gate's calls.
func removeResp(rest ...queuedResp) []queuedResp {
	return append(append(yesProbes(kubeProbes), queuedResp{"", nil}), rest...)
}

// TestDeployKubernetesProbesOperatorSecrets pins the deploy-time check of a
// Secret the operator manages and the pod mounts: after the permission probes,
// one `get secret` per Secret; a missing Secret or key stops the deploy before
// anything is applied, a Secret the identity may not read only warns, and a
// complete one lets the apply through. remove never probes.
func TestDeployKubernetesProbesOperatorSecrets(t *testing.T) {
	env := sharedEnv + "tls:\n  stores:\n    ts:\n      password: pw\n      existing:\n        secret:\n          name: broker-cert\n          key: ca.p12\n" + kubeEnv
	notFound := errors.New("exit status 1")
	for _, c := range []struct {
		name      string
		reply     queuedResp
		wantCode  int
		wantApply bool
	}{
		{"missing Secret", queuedResp{`Error from server (NotFound): secrets "broker-cert" not found`, notFound}, 1, false},
		{"missing key", queuedResp{`{"kind":"Secret","data":{"tls.crt":"QQ=="}}`, nil}, 1, false},
		{"not allowed to read it", queuedResp{`Error from server (Forbidden): secrets "broker-cert" is forbidden`, notFound}, 0, true},
		{"complete", queuedResp{`{"kind":"Secret","data":{"ca.p12":"QQ=="}}`, nil}, 0, true},
	} {
		t.Run(c.name, func(t *testing.T) {
			f := &queueRunner{resp: append(yesProbes(kubeProbes), c.reply)}
			dir := t.TempDir()
			write(t, dir, "env.yaml", env)
			write(t, dir, "10.yaml", tlsWF)
			var code int
			stderr := captureStderr(t, func() {
				code = dispatch([]string{"deploy", "--platform", "kubernetes", "-e", filepath.Join(dir, "env.yaml")}, f)
			})
			if code != c.wantCode {
				t.Fatalf("exit = %d, want %d\n%s", code, c.wantCode, stderr)
			}
			if want := []string{"kubectl", "get", "secret", "broker-cert", "-n", "solace-connectors", "-o", "json"}; len(f.calls) <= kubeProbes || !reflect.DeepEqual(f.calls[kubeProbes].argv, want) {
				t.Fatalf("call %d = %+v, want %v", kubeProbes, f.calls, want)
			}
			applied := false
			for _, call := range f.calls {
				applied = applied || reflect.DeepEqual(call.argv, []string{"kubectl", "apply", "-f", "-"})
			}
			if applied != c.wantApply {
				t.Errorf("applied = %v, want %v: %+v\n%s", applied, c.wantApply, f.calls, stderr)
			}
			if c.name != "complete" && !strings.Contains(stderr, "the Secret broker-cert") {
				t.Errorf("stderr names no Secret:\n%s", stderr)
			}
		})
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

	if code := dispatch([]string{"deploy", "--platform", "kubernetes", "-e", filepath.Join(dir, "env.yaml")}, f); code != 2 {
		t.Fatalf("unsafe kubernetes.command should exit 2, got %d", code)
	}
	if len(f.calls) != 0 {
		t.Fatalf("nothing must reach the runner when the command is rejected, got %d calls", len(f.calls))
	}
}

// TestRemoveKubernetesSeamPromptsBeforeTearingDown is the confirmation half of
// TestRemoveKubernetesSeamHappyPath above: an accepted answer produces exactly
// the same calls, so the prompt gates the teardown without changing it.
func TestRemoveKubernetesSeamPromptsBeforeTearingDown(t *testing.T) {
	for _, answer := range []string{"y", "yes", "YES", "Y"} {
		t.Run(answer, func(t *testing.T) {
			f := &queueRunner{resp: removeResp(queuedResp{nsList(), nil}, queuedResp{"yes\n", nil}, queuedResp{"", nil})}
			withPrompter(t, true, answer+"\n"+answer+"\n")
			dir := t.TempDir()
			write(t, dir, "env.yaml", sharedEnv+kubeEnv)
			write(t, dir, "10.yaml", validWF)

			if code := dispatch([]string{"remove", "--platform", "kubernetes", "-e", filepath.Join(dir, "env.yaml")}, f); code != 0 {
				t.Fatalf("exit=%d, want 0", code)
			}
			// The same answer approves the namespace question too, so every call
			// happens: probes, delete, listing, delete probe, namespace delete.
			if len(f.calls) != kubeProbes+4 {
				t.Fatalf("want %d calls, got %d: %+v", kubeProbes+4, len(f.calls), f.calls)
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
			withPrompter(t, true, answer+"\n")
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

// TestRemoveNoPromptSkipsThePromptEntirely covers the scripted path: --no-prompt must write
// no question at all, rather than asking it and ignoring the answer.
func TestRemoveNoPromptSkipsThePromptEntirely(t *testing.T) {
	f := &queueRunner{resp: removeResp(queuedResp{nsList(), nil}, queuedResp{"yes\n", nil}, queuedResp{"", nil})}
	out := withPrompter(t, true, "n\nn\n")

	dir := t.TempDir()
	write(t, dir, "env.yaml", sharedEnv+kubeEnv)
	write(t, dir, "10.yaml", validWF)

	if code := dispatch([]string{"remove", "--platform", "kubernetes", noPromptFlagName, "-e", filepath.Join(dir, "env.yaml")}, f); code != 0 {
		t.Fatalf("exit=%d, want 0", code)
	}
	// Neither question is asked, and the empty namespace goes with it:
	// probes, delete, listing, delete probe, namespace delete.
	if len(f.calls) != kubeProbes+4 {
		t.Fatalf("want %d calls, got %d: %+v", kubeProbes+4, len(f.calls), f.calls)
	}
	if out.Len() != 0 {
		t.Errorf("--no-prompt must not prompt at all, got %q", out.String())
	}
}

// TestRemoveNonTTYFailsFastNamingTheFlag is the CI case: rather than blocking on
// a read that will never return, the prompt refuses and the error names the flag
// that skips it -- the same shape the platform menu and the status install
// confirmation use.
func TestRemoveNonTTYFailsFastNamingTheFlag(t *testing.T) {
	f := useFakeRunner(t)
	out := withPrompter(t, false, "y\ny\n")

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
	if out.Len() != 0 {
		t.Errorf("a non-terminal must not be asked anything, got %q", out.String())
	}
}

// TestRemovePromptNamesWhatItWillDestroy is why the prompt is worth having: the
// question has to carry the identifier that tells one target from another, or an
// operator confirms "kubernetes" and loses production. On kubernetes that is the
// namespace.
func TestRemovePromptNamesWhatItWillDestroy(t *testing.T) {
	for _, c := range []struct {
		platform string
		env      string
		want     []string
	}{
		{"kubernetes", kubeEnv, []string{"solmq-connector", "solace-connectors"}},
		{"docker", dockerEnv, []string{"container"}},
		{"podman", podmanEnv(), []string{"container", ".service"}},
	} {
		t.Run(c.platform, func(t *testing.T) {
			f := useFakeRunner(t)
			out := withPrompter(t, true, "n\n")

			dir := t.TempDir()
			write(t, dir, "env.yaml", sharedEnv+c.env)
			write(t, dir, "10.yaml", validWF)

			captureStderr(t, func() {
				dispatch([]string{"remove", "--platform", c.platform, "-e", filepath.Join(dir, "env.yaml")}, f)
			})
			asked := out.String()
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

// nsListTypes is what nsgate.List asks for, in one query.
const nsListTypes = "all,persistentvolumeclaims,secrets,configmaps"

// kubeFixture writes env.yaml (the shared keys plus kube) and the valid workflow
// into a fresh directory and returns the env file's path.
func kubeFixture(t *testing.T, kube string) string {
	t.Helper()
	dir := t.TempDir()
	write(t, dir, "env.yaml", sharedEnv+kube)
	write(t, dir, "10.yaml", validWF)
	return filepath.Join(dir, "env.yaml")
}

// runKube dispatches a kubernetes verb against r and returns its exit code and what
// it wrote to stderr.
func runKube(t *testing.T, r gdurunner.Runner, envPath string, args ...string) (int, string) {
	t.Helper()
	var code int
	stderr := captureStderr(t, func() {
		code = dispatch(append(append([]string(nil), args...), "--platform", "kubernetes", "-e", envPath), r)
	})
	return code, stderr
}

// canIAsked lists the permission questions among calls, as "verb resource [flags]".
func canIAsked(calls []fakeCall) []string {
	var out []string
	for _, c := range calls {
		for i := 0; i+1 < len(c.argv); i++ {
			if c.argv[i] == "auth" && c.argv[i+1] == "can-i" {
				out = append(out, strings.Join(c.argv[i+2:], " "))
				break
			}
		}
	}
	return out
}

// TestRemoveTeardownManifestOmitsTheNamespace is the bug this whole file exists
// for: the manifest piped to `kubectl delete -f -` used to carry a Namespace,
// which cascades to every object inside it -- including workloads this tool
// never deployed. The namespace is a separate, checked step now; the manifest
// must never carry it.
func TestRemoveTeardownManifestOmitsTheNamespace(t *testing.T) {
	f := &queueRunner{resp: removeResp(queuedResp{nsList(), nil}, queuedResp{"yes\n", nil}, queuedResp{"", nil})}
	withPrompter(t, true, "y\ny\n")
	env := kubeFixture(t, kubeEnv)

	if code, stderr := runKube(t, f, env, "remove"); code != 0 {
		t.Fatalf("exit=%d, want 0\n%s", code, stderr)
	}
	if strings.Contains(f.calls[kubeProbes].stdin, "kind: Namespace") {
		t.Fatalf("a teardown manifest must not carry the namespace:\n%s", f.calls[kubeProbes].stdin)
	}
	// The namespace goes by its own `delete namespace`, not by a document piped
	// in, and only after the occupancy check and the permission probe passed.
	last := f.calls[len(f.calls)-1]
	if want := []string{"kubectl", "delete", "namespace", "solace-connectors", "--ignore-not-found"}; !reflect.DeepEqual(last.argv, want) || last.stdin != "" {
		t.Errorf("the namespace step = %v with stdin %q, want %v with none", last.argv, last.stdin, want)
	}
}

// TestDeployManifestOmitsTheNamespace: a deploy creates the namespace on its own,
// only when it is absent, so the manifest it applies carries no Namespace document
// whether or not the namespace exists -- and `generate kubernetes`, which is applied
// by hand, still does.
func TestDeployManifestOmitsTheNamespace(t *testing.T) {
	env := kubeFixture(t, kubeEnv)
	f := useFakeRunner(t)
	if code, stderr := runKube(t, f, env, "deploy"); code != 0 {
		t.Fatalf("deploy exit=%d\n%s", code, stderr)
	}
	if strings.Contains(f.calls[kubeProbes+1].stdin, "kind: Namespace") {
		t.Errorf("the deploy manifest must carry no Namespace:\n%s", f.calls[kubeProbes+1].stdin)
	}
	var code int
	out := captureStdout(t, func() { code = run([]string{"generate", "--platform", "kubernetes", "-e", env}) })
	if code != 0 || !strings.HasPrefix(out, "apiVersion: v1\nkind: Namespace\n") {
		t.Errorf("generate kubernetes keeps the Namespace first, exit %d:\n%s", code, out)
	}
}

// TestDeployNamespaceFlow is every branch of the namespace step, which runs after the
// probes and before the apply: nothing for a namespace that exists (no cluster
// permission asked), and otherwise the create probe, then the Namespace applied on
// its own before the manifest.
func TestDeployNamespaceFlow(t *testing.T) {
	notFoundErr := errors.New("exit status 1")
	const nsDoc = "apiVersion: v1\nkind: Namespace\nmetadata:\n  name: solace-connectors\n"
	exists := queuedResp{"namespace/solace-connectors\n", nil}
	absent := queuedResp{"", nil}
	unreadable := queuedResp{`Error from server (Forbidden): namespaces "solace-connectors" is forbidden`, notFoundErr}
	allowed := queuedResp{"yes\n", nil}
	refused := queuedResp{"no\n", notFoundErr}
	unanswered := queuedResp{"error: You must be logged in to the server (Unauthorized)\n", notFoundErr}
	for _, c := range []struct {
		name        string
		read, probe queuedResp // the namespace read, and the create probe when it is asked
		applyNS     queuedResp // the Namespace's own apply, when it is made
		wantCode    int
		wantProbe   bool   // `create namespaces` asked
		wantNSApply bool   // the Namespace applied on its own
		wantManif   bool   // the manifest applied
		wantStderr  string // contained in stderr
	}{
		{name: "exists asks nothing more", read: exists, wantManif: true},
		{name: "absent and allowed creates it first", read: absent, probe: allowed, wantProbe: true, wantNSApply: true, wantManif: true},
		{name: "absent and refused is MissingNamespace", read: absent, probe: refused, wantCode: 1, wantProbe: true,
			wantStderr: `namespace "solace-connectors" does not exist, and this identity may not create it`},
		{name: "absent and unanswered is its own error", read: absent, probe: unanswered, wantCode: 1, wantProbe: true,
			wantStderr: "cannot check permission to create namespaces"},
		{name: "unknown and allowed creates it", read: unreadable, probe: allowed, wantProbe: true, wantNSApply: true, wantManif: true},
		{name: "unknown and refused warns and carries on", read: unreadable, probe: refused, wantProbe: true, wantManif: true,
			wantStderr: "warning: namespace solace-connectors could not be read (that needs get namespaces) and this identity may not create namespaces"},
		{name: "unknown and unanswered is an error", read: unreadable, probe: unanswered, wantCode: 1, wantProbe: true,
			wantStderr: "cannot check permission to create namespaces"},
		{name: "a failed Namespace apply stops the deploy", read: absent, probe: allowed, applyNS: queuedResp{"denied", notFoundErr},
			wantCode: 1, wantProbe: true, wantNSApply: true, wantStderr: "create namespace solace-connectors for kubernetes failed"},
	} {
		t.Run(c.name, func(t *testing.T) {
			resp := append(yesProbes(kubeProbes), c.read)
			if c.wantProbe {
				resp = append(resp, c.probe)
			}
			if c.wantNSApply {
				resp = append(resp, c.applyNS)
			}
			q := &queueRunner{resp: resp}
			code, stderr := runKube(t, q, kubeFixture(t, kubeEnv), "deploy")
			if code != c.wantCode {
				t.Fatalf("exit = %d, want %d\n%s", code, c.wantCode, stderr)
			}
			if c.wantStderr != "" && !strings.Contains(stderr, c.wantStderr) {
				t.Errorf("stderr lacks %q:\n%s", c.wantStderr, stderr)
			}
			var probed, nsApplied, applied int
			for _, call := range q.calls {
				switch {
				case reflect.DeepEqual(call.argv, []string{"kubectl", "auth", "can-i", "create", "namespaces", "--all-namespaces"}):
					probed++
				case call.stdin == nsDoc:
					nsApplied++
				case reflect.DeepEqual(call.argv, []string{"kubectl", "apply", "-f", "-"}):
					applied++
				}
			}
			if (probed == 1) != c.wantProbe || probed > 1 {
				t.Errorf("the create-namespaces probe was asked %d times, want %v", probed, c.wantProbe)
			}
			if (nsApplied == 1) != c.wantNSApply || nsApplied > 1 {
				t.Errorf("the Namespace was applied on its own %d times, want %v: %+v", nsApplied, c.wantNSApply, q.calls)
			}
			if (applied == 1) != c.wantManif || applied > 1 {
				t.Errorf("the manifest was applied %d times, want %v", applied, c.wantManif)
			}
			if c.wantNSApply && c.wantManif {
				// The Namespace is its own apply, before the manifest's.
				ns := indexOfCall(q.calls, func(fc fakeCall) bool { return fc.stdin == nsDoc })
				man := indexOfCall(q.calls, func(fc fakeCall) bool { return strings.Contains(fc.stdin, "kind: Deployment") })
				if ns < 0 || man < 0 || ns > man {
					t.Errorf("the Namespace apply (call %d) must come before the manifest (call %d)", ns, man)
				}
			}
		})
	}
}

// indexOfCall is the index of the first call match accepts, or -1.
func indexOfCall(calls []fakeCall, match func(fakeCall) bool) int {
	for i, c := range calls {
		if match(c) {
			return i
		}
	}
	return -1
}

// TestDeployWaitsForTheRollout: after a successful apply the deploy waits for the
// Deployment (kube.RolloutStatus, 300s), keeping the one ok: line at the end. A
// refused wait warns and succeeds, since the change is applied; a wait that times
// out or sees the rollout fail is the deploy's failure, exit 1.
func TestDeployWaitsForTheRollout(t *testing.T) {
	forbidden := `Error from server (Forbidden): deployments.apps "solmq-connector" is forbidden: User "u" cannot get resource "deployments/scale"`
	for _, c := range []struct {
		name       string
		wait       queuedResp
		wantCode   int
		wantStderr []string
		notStderr  string
	}{
		{"rolled out", queuedResp{`deployment "solmq-connector" successfully rolled out`, nil}, 0,
			[]string{`deployment "solmq-connector" successfully rolled out`, "ok: deploy kubernetes"}, "warning: the wait"},
		{"refused", queuedResp{forbidden, errors.New("exit status 1")}, 0,
			[]string{"warning: the wait for deployment/solmq-connector to roll out was refused", "the change was applied", "ok: deploy kubernetes"}, "error:"},
		{"timed out", queuedResp{"error: timed out waiting for the condition", errors.New("exit status 1")}, 1,
			[]string{"error: deploy kubernetes failed", "deployment/solmq-connector did not finish rolling out"}, "ok:"},
	} {
		t.Run(c.name, func(t *testing.T) {
			q := &queueRunner{resp: append(yesProbes(kubeProbes), queuedResp{"namespace/solace-connectors\n", nil}, queuedResp{"", nil}, c.wait)}
			code, stderr := runKube(t, q, kubeFixture(t, kubeEnv), "deploy")
			if code != c.wantCode {
				t.Fatalf("exit = %d, want %d\n%s", code, c.wantCode, stderr)
			}
			for _, want := range c.wantStderr {
				if !strings.Contains(stderr, want) {
					t.Errorf("stderr lacks %q:\n%s", want, stderr)
				}
			}
			if strings.Contains(stderr, c.notStderr) {
				t.Errorf("stderr must not carry %q:\n%s", c.notStderr, stderr)
			}
			if len(q.calls) != kubeProbes+3 {
				t.Errorf("want probes, namespace read, apply and one wait, got %d calls: %+v", len(q.calls), q.calls)
			}
		})
	}
	// A failed apply is not waited on.
	q := &queueRunner{resp: append(yesProbes(kubeProbes), queuedResp{"namespace/solace-connectors\n", nil}, queuedResp{"denied", errors.New("exit status 1")})}
	if code, _ := runKube(t, q, kubeFixture(t, kubeEnv), "deploy"); code != 1 || len(q.calls) != kubeProbes+2 {
		t.Errorf("a failed apply must not wait: exit %d, calls %+v", code, q.calls)
	}
}

// TestKubernetesProbeSets: every probe of a command is asked before its first write.
// A deploy asks create, and a remove delete, of each kind the manifest writes, once:
// the credentials Secret only when created, the Service only when enabled, and the
// libs claim and its cluster-scoped volume (asked cluster-wide) only for
// libs.pvc.create. The Namespace is never among them.
func TestKubernetesProbeSets(t *testing.T) {
	const withLibs = `
kubernetes:
  command: kubectl
  deployment:
    name: solmq-connector
    namespace: solace-connectors
  service:
    enabled: true
  secrets:
    credentials:
      create: true
  libs:
    pvc:
      create:
        nfs:
          server: nfs1
          path: /libs
`
	const theirCreds = `
kubernetes:
  command: kubectl
  deployment:
    name: solmq-connector
    namespace: solace-connectors
  secrets:
    credentials:
      existing: their-creds
`
	in := func(verb string, resources ...string) []string {
		var out []string
		for _, r := range resources {
			if r == "persistentvolumes" {
				out = append(out, verb+" "+r+" --all-namespaces")
				continue
			}
			out = append(out, verb+" "+r+" -n solace-connectors")
		}
		return out
	}
	for _, c := range []struct {
		name, env    string
		deploy, drop []string
	}{
		{"credentials secret, no service, no libs", kubeEnv,
			in("create", "deployments", "configmaps", "secrets"), in("delete", "deployments", "configmaps", "secrets")},
		{"a service and the libs claim with its volume", withLibs,
			in("create", "deployments", "services", "configmaps", "secrets", "persistentvolumeclaims", "persistentvolumes"),
			in("delete", "deployments", "services", "configmaps", "secrets", "persistentvolumeclaims", "persistentvolumes")},
		// The credentials Secret is the operator's, so a deploy writes no Secret at
		// all; a teardown still deletes the stores Secret by name, whose name the
		// instance owns whether or not an earlier deploy made it.
		{"no created secret", theirCreds, in("create", "deployments", "configmaps"), in("delete", "deployments", "configmaps", "secrets")},
	} {
		t.Run(c.name, func(t *testing.T) {
			f := useFakeRunner(t)
			runKube(t, f, kubeFixture(t, c.env), "deploy")
			if got := canIAsked(f.calls); !reflect.DeepEqual(got, c.deploy) {
				t.Errorf("deploy asked %v, want %v", got, c.deploy)
			}
			f = useFakeRunner(t)
			withPrompter(t, true, "n\ny\n")
			runKube(t, f, kubeFixture(t, c.env), "remove", noPromptFlagName)
			// The remove's last question is the namespace delete, asked after the
			// occupancy check, so only the first len(c.drop) belong to the teardown.
			got := canIAsked(f.calls)
			if len(got) < len(c.drop) || !reflect.DeepEqual(got[:len(c.drop)], c.drop) {
				t.Errorf("remove asked %v, want %v first", got, c.drop)
			}
			for _, q := range got {
				if strings.Contains(q, " namespaces") && !strings.HasPrefix(q, "delete namespaces") {
					t.Errorf("a namespaces probe other than the gate's delete: %q", q)
				}
			}
		})
	}
}

// Every probe runs before the first write, and the first refusal stops the command:
// nothing is applied, and the later probes are not asked.
func TestKubernetesRefusedProbeStopsBeforeAnyWrite(t *testing.T) {
	resp := append(yesProbes(1), queuedResp{"no\n", errors.New("exit status 1")})
	q := &queueRunner{resp: resp}
	code, stderr := runKube(t, q, kubeFixture(t, kubeEnv), "deploy")
	if code != 1 {
		t.Fatalf("exit = %d, want 1\n%s", code, stderr)
	}
	if want := `not allowed to create configmaps in namespace "solace-connectors"`; !strings.Contains(stderr, want) {
		t.Errorf("stderr lacks %q:\n%s", want, stderr)
	}
	if len(q.calls) != 2 {
		t.Errorf("want the first two probes only, got %+v", q.calls)
	}
}

// TestRemoveNamespaceOccupiedLeavesItAlone is the case that motivated the flag:
// another connector, or anything else, living in the same namespace.
func TestRemoveNamespaceOccupiedLeavesItAlone(t *testing.T) {
	q := &queueRunner{resp: removeResp(
		queuedResp{nsList(
			nsItemJSON("Deployment", "billing-api"),
			nsItemJSON("StatefulSet", "postgres"),
			nsItemJSON("PersistentVolumeClaim", "billing-data"),
		), nil},
	)}
	withPrompter(t, true, "y\n")
	env := kubeFixture(t, kubeEnv)

	code, stderr := runKube(t, q, env, "remove")
	if code != 0 {
		t.Fatalf("exit=%d, want 0 -- the teardown succeeded", code)
	}
	// probes, delete, the occupancy listing -- and nothing after it.
	if len(q.calls) != kubeProbes+2 {
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
// around it, so it is asked on its own -- after the permission probe, which is
// never wasted on a question that would be refused.
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
			q := &queueRunner{resp: removeResp(queuedResp{nsList(), nil}, queuedResp{"yes\n", nil}, queuedResp{"", nil})}
			// Two questions: yes to the teardown, then the case's answer to the
			// namespace. That separation is the point of the feature.
			withPrompter(t, true, "y\n"+c.answer+"\n")
			env := kubeFixture(t, kubeEnv)

			if code, stderr := runKube(t, q, env, "remove"); code != 0 {
				t.Fatalf("exit=%d, want 0\n%s", code, stderr)
			}
			want := kubeProbes + 3
			if c.deleted {
				want = kubeProbes + 4
			}
			if len(q.calls) != want {
				t.Fatalf("want %d calls, got %d: %+v", want, len(q.calls), q.calls)
			}
			if c.deleted {
				if got, want := q.calls[kubeProbes+3].argv, []string{"kubectl", "delete", "namespace", "solace-connectors", "--ignore-not-found"}; !reflect.DeepEqual(got, want) {
					t.Errorf("the namespace delete = %v, want %v", got, want)
				}
			}
		})
	}
}

// TestRemoveSharesOneReaderAcrossBothQuestions pins the one-prompter-per-run
// rule: the teardown and the namespace questions read from the same buffered
// reader, so one write answering both ("y\ny\n") is not half lost to the first
// question's read-ahead.
func TestRemoveSharesOneReaderAcrossBothQuestions(t *testing.T) {
	q := &queueRunner{resp: removeResp(queuedResp{nsList(), nil}, queuedResp{"yes\n", nil}, queuedResp{"", nil})}
	out := withPrompter(t, true, "y\ny\n")
	env := kubeFixture(t, kubeEnv)

	if code, stderr := runKube(t, q, env, "remove"); code != 0 {
		t.Fatalf("exit=%d, want 0\n%s", code, stderr)
	}
	if n := strings.Count(out.String(), "[y/N]"); n != 2 {
		t.Errorf("want 2 questions asked, got %d: %q", n, out.String())
	}
	if len(q.calls) != kubeProbes+4 {
		t.Fatalf("both answers must be read, want %d calls (probes, delete, listing, delete probe, namespace delete), got %d: %+v", kubeProbes+4, len(q.calls), q.calls)
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
	q := &queueRunner{resp: removeResp(queuedResp{nsList(nsItemJSON("Deployment", "billing-api")), nil})}
	out := withPrompter(t, true, "n\nn\n")
	env := kubeFixture(t, kubeEnv)

	code, stderr := runKube(t, q, env, "remove", noPromptFlagName)
	if code != 0 {
		t.Fatalf("exit=%d, want 0 -- the teardown succeeded", code)
	}
	if out.Len() != 0 {
		t.Errorf("--no-prompt must not prompt at all, got %q", out.String())
	}
	// probes, delete, listing -- and nothing after it, whatever the flags say.
	if len(q.calls) != kubeProbes+2 {
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
	q := &queueRunner{resp: removeResp(queuedResp{nsList(), nil}, queuedResp{"yes\n", nil}, queuedResp{"", nil})}
	out := withPrompter(t, true, "y\ny\n")
	env := kubeFixture(t, kubeEnv)

	if code, stderr := runKube(t, q, env, "remove"); code != 0 {
		t.Fatalf("exit=%d, want 0\n%s", code, stderr)
	}
	// Two questions with no flags at all: the teardown, then the namespace.
	asked := out.String()
	if n := strings.Count(asked, "[y/N]"); n != 2 {
		t.Fatalf("want 2 prompts (teardown, namespace), got %d: %q", n, asked)
	}
	if !strings.Contains(asked, "remove the namespace too?") {
		t.Errorf("the second question should be about the namespace, got %q", asked)
	}
	if len(q.calls) != kubeProbes+4 {
		t.Fatalf("want %d calls (probes, delete, listing, delete probe, namespace delete), got %d: %+v", kubeProbes+4, len(q.calls), q.calls)
	}
	if !containsToken(q.calls[kubeProbes+1].argv, nsListTypes) {
		t.Errorf("call %d should be the occupancy listing, got %v", kubeProbes+1, q.calls[kubeProbes+1].argv)
	}
	if want := []string{"kubectl", "auth", "can-i", "delete", "namespaces", "--all-namespaces"}; !reflect.DeepEqual(q.calls[kubeProbes+2].argv, want) {
		t.Errorf("call %d should be the delete probe, got %v, want %v", kubeProbes+2, q.calls[kubeProbes+2].argv, want)
	}
}

// TestRemoveNamespaceRefusesClusterNamespaces pins the guard that no occupancy
// result can override. Deleting one of these breaks the cluster rather than an
// application, and a wrong namespace in env.yaml is the likeliest way to get here.
func TestRemoveNamespaceRefusesClusterNamespaces(t *testing.T) {
	for _, ns := range []string{"default", "kube-system", "kube-public", "kube-node-lease"} {
		t.Run(ns, func(t *testing.T) {
			env := strings.Replace(kubeEnv, "namespace: solace-connectors", "namespace: "+ns, 1)

			q := &queueRunner{resp: removeResp()}
			code, stderr := runKube(t, q, kubeFixture(t, env), "remove", noPromptFlagName)
			if code != 0 {
				t.Fatalf("exit=%d, want 0", code)
			}
			// Refused before the occupancy listing: the probes and the delete only.
			if len(q.calls) != kubeProbes+1 {
				t.Fatalf("%s must be refused without even being listed, got %d calls: %+v", ns, len(q.calls), q.calls)
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
	q := &queueRunner{resp: removeResp(queuedResp{"boom", fmt.Errorf("forbidden")})}
	code, stderr := runKube(t, q, kubeFixture(t, kubeEnv), "remove", noPromptFlagName)
	if code != 0 {
		t.Fatalf("exit=%d, want 0 -- the teardown itself succeeded", code)
	}
	if len(q.calls) != kubeProbes+2 {
		t.Fatalf("the namespace must not be deleted after a failed listing, got %d calls", len(q.calls))
	}
	if !strings.Contains(stderr, "left in place") {
		t.Errorf("stderr should say the namespace was left alone, got:\n%s", stderr)
	}
}

// TestRemoveFailedTeardownSkipsTheNamespace: whatever did not come down is still
// in there, so the namespace is not even considered.
func TestRemoveFailedTeardownSkipsTheNamespace(t *testing.T) {
	q := &queueRunner{resp: append(yesProbes(kubeProbes), queuedResp{"boom", fmt.Errorf("denied")})}
	code, _ := runKube(t, q, kubeFixture(t, kubeEnv), "remove", noPromptFlagName)
	if code != 1 {
		t.Fatalf("exit=%d, want 1", code)
	}
	if len(q.calls) != kubeProbes+1 {
		t.Fatalf("a failed teardown must not list or delete the namespace, got %d calls: %+v", len(q.calls), q.calls)
	}
}

// TestRemoveKubernetesSeamRejectsUnsafeCommand: the namespace gate takes the argv the
// command was guarded into, so a command the guard refuses fails the remove as a usage
// error before anything reaches the runner -- there is no second path around it.
func TestRemoveKubernetesSeamRejectsUnsafeCommand(t *testing.T) {
	f := useFakeRunner(t)
	code, stderr := runKube(t, f, kubeFixture(t, kubeEnvUnsafeCommand), "remove", noPromptFlagName)
	if code != 2 {
		t.Fatalf("exit=%d, want 2\n%s", code, stderr)
	}
	if len(f.calls) != 0 {
		t.Errorf("nothing may run, got %d calls", len(f.calls))
	}
}

// gateKube is the section the direct calls to removeNamespace use.
func gateKube(ns string) *spec.Kubernetes {
	k := &spec.Kubernetes{Command: spec.DefaultKubeCommand}
	k.Deployment.Name = "solmq-connector"
	k.Deployment.Namespace = ns
	return k
}

// gateBase is the guarded kubectl argv the caller hands the gate.
var gateBase = []string{"kubectl"}

// TestRemoveNamespaceDiscountsWhatTheReleaseOwnsByKind is the live half of the
// occupancy check, now nsgate's: the objects this release created are passed with
// their kinds, so the Deployment and what it owns do not count, but another kind of
// object that merely shares a name does.
func TestRemoveNamespaceDiscountsWhatTheReleaseOwnsByKind(t *testing.T) {
	list := nsList(
		nsItemJSON("Deployment", "solmq-connector"),
		nsItemJSON("ConfigMap", "solmq-connector-config"),
		`{"kind":"ReplicaSet","metadata":{"name":"solmq-connector-7d4","ownerReferences":[{"kind":"Deployment","name":"solmq-connector"}]}}`,
		`{"kind":"Pod","metadata":{"name":"solmq-connector-7d4-abc","ownerReferences":[{"kind":"ReplicaSet","name":"solmq-connector-7d4"}]}}`,
		// Not ours, though the name is: a ConfigMap is not our Deployment.
		nsItemJSON("ConfigMap", "solmq-connector"),
	)
	q := &queueRunner{resp: []queuedResp{{list, nil}}}
	var code int
	stderr := captureStderr(t, func() {
		code = removeNamespace(q, actionOpts{action: "remove", noPrompt: true}, gateKube("prod"), gateBase)
	})
	if code != 0 || len(q.calls) != 1 {
		t.Fatalf("exit=%d, calls %+v, want 0 and the listing only", code, q.calls)
	}
	if !strings.Contains(stderr, "still holds 1 resource(s)") || !strings.Contains(stderr, "configmap/solmq-connector") ||
		strings.Contains(stderr, "pod/") || strings.Contains(stderr, "deployment/") {
		t.Errorf("only the foreign ConfigMap is an occupant, got:\n%s", stderr)
	}
}

// TestOwnedObjectsCoverEverythingThisReleaseCreates is the safety net behind the
// occupancy check. The check runs after the teardown, so these objects are
// normally gone -- but one that has not finished terminating and carries no
// deletionTimestamp yet must not be mistaken for someone else's workload and
// keep the namespace alive forever.
func TestOwnedObjectsCoverEverythingThisReleaseCreates(t *testing.T) {
	// The retired name keys are set on purpose: what the teardown created and
	// deletes is the derived name, so that -- not the key -- is what is ours.
	k := &spec.Kubernetes{
		Secrets: spec.Secrets{
			Credentials: &spec.CredentialsSecret{Create: &spec.CredCreate{Name: "solmq-creds"}},
			Stores:      &spec.StoresSecret{Create: &spec.StoreCreate{Name: "solmq-tls"}},
			ImagePull:   &spec.ImagePullSecret{Name: "solmq-pull", Create: true},
		},
		Libs: &spec.Libs{PVC: &spec.LibsPVC{Create: &spec.PVCCreate{Name: "old-libs"}}},
	}
	k.Deployment.Name = "solmq-connector"

	owned := map[nsgate.Object]bool{}
	for _, o := range k.CreatedNames() {
		owned[nsgate.Object{Kind: o.Kind, Name: o.Name}] = true
	}
	for _, want := range []nsgate.Object{
		{Kind: "Deployment", Name: "solmq-connector"}, {Kind: "Service", Name: "solmq-connector"},
		{Kind: "ConfigMap", Name: "solmq-connector-config"}, {Kind: "Secret", Name: "solmq-connector-credentials"},
		{Kind: "Secret", Name: "solmq-connector-stores"}, {Kind: "Secret", Name: "solmq-connector-image-pull"},
		{Kind: "PersistentVolumeClaim", Name: "solmq-connector-libs"},
	} {
		if !owned[want] {
			t.Errorf("CreatedNames should carry %v, got %v", want, owned)
		}
	}
	for _, retired := range []string{"solmq-creds", "solmq-tls", "solmq-pull", "old-libs"} {
		for o := range owned {
			if o.Name == retired {
				t.Errorf("%q is a retired name key the tool never created, so it is not ours, got %v", retired, owned)
			}
		}
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
	for _, o := range k2.CreatedNames() {
		for _, no := range []string{"their-creds", "their-tls", "their-pull", "their-pvc"} {
			if o.Name == no {
				t.Errorf("%q is referenced, not created, so it is not ours to discount", no)
			}
		}
	}
}

// TestRemoveNamespaceWithoutANamespaceSaysSo covers the guard for a kubernetes
// section carrying no namespace: there is nothing to remove, and that is worth
// saying rather than deleting something named "".
func TestRemoveNamespaceWithoutANamespaceSaysSo(t *testing.T) {
	f := useFakeRunner(t)

	var code int
	stderr := captureStderr(t, func() {
		code = removeNamespace(f, actionOpts{action: "remove"}, gateKube(""), gateBase)
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

// TestRemoveNamespaceUnreadableProbeLeavesItAlone: output that cannot be parsed,
// or a listing with nothing in it to read, must not read as "empty" and authorise
// the delete -- nsgate.Occupants errors on both, and the namespace is kept with a
// warning naming why.
func TestRemoveNamespaceUnreadableProbeLeavesItAlone(t *testing.T) {
	for _, c := range []struct{ name, out, want string }{
		{"not JSON", "not json at all", "not JSON"},
		{"no items", "{}", "no items"},
		{"empty", "  ", "came back empty"},
	} {
		t.Run(c.name, func(t *testing.T) {
			q := &queueRunner{resp: []queuedResp{{c.out, nil}}}
			var code int
			stderr := captureStderr(t, func() {
				code = removeNamespace(q, actionOpts{action: "remove", noPrompt: true}, gateKube("prod"), gateBase)
			})
			if code != 0 {
				t.Fatalf("exit=%d, want 0", code)
			}
			if len(q.calls) != 1 {
				t.Fatalf("the namespace must not be deleted, got %d calls: %+v", len(q.calls), q.calls)
			}
			if !strings.Contains(stderr, "left in place") || !strings.Contains(stderr, c.want) {
				t.Errorf("stderr should say the namespace was left alone and why (%q), got:\n%s", c.want, stderr)
			}
		})
	}
}

// TestRemoveNamespaceNonTTYKeepsItNamingTheFlag: the namespace question has no
// terminal to ask on, which used to fail the remove. The gate never fails the
// remove: the namespace is kept with a warning that names the flag that answers
// without a terminal, and the exit code stays 0.
func TestRemoveNamespaceNonTTYKeepsItNamingTheFlag(t *testing.T) {
	q := &queueRunner{resp: []queuedResp{{nsList(), nil}, {"yes\n", nil}}}
	withPrompter(t, false, "y\n")

	var code int
	stderr := captureStderr(t, func() { code = removeNamespace(q, actionOpts{action: "remove"}, gateKube("prod"), gateBase) })
	if code != 0 {
		t.Fatalf("exit=%d, want 0 -- the teardown succeeded", code)
	}
	if !strings.Contains(stderr, "warning:") || !strings.Contains(stderr, noPromptFlagName) {
		t.Errorf("the warning must name %s, got:\n%s", noPromptFlagName, stderr)
	}
	if len(q.calls) != 2 {
		t.Errorf("the namespace must not be deleted, got %d calls: %+v", len(q.calls), q.calls)
	}
}

// TestRemoveNamespaceAsksThePermissionBeforeTheQuestion: the delete probe runs after
// the occupancy check and before the prompt, so nobody is asked a question that would
// be refused. A refusal, and a probe that could not be answered, keep the namespace
// with a warning and exit 0, without asking.
func TestRemoveNamespaceAsksThePermissionBeforeTheQuestion(t *testing.T) {
	for _, c := range []struct {
		name  string
		probe queuedResp
		want  string
	}{
		{"refused", queuedResp{"no\n", errors.New("exit status 1")}, `not allowed to delete namespaces in the cluster (auth can-i answered "no")`},
		{"unanswered", queuedResp{"error: no context\n", errors.New("exit status 1")}, "cannot check permission to delete namespaces"},
	} {
		t.Run(c.name, func(t *testing.T) {
			q := &queueRunner{resp: []queuedResp{{nsList(), nil}, c.probe}}
			out := withPrompter(t, true, "y\n")
			var code int
			stderr := captureStderr(t, func() { code = removeNamespace(q, actionOpts{action: "remove"}, gateKube("prod"), gateBase) })
			if code != 0 {
				t.Fatalf("exit=%d, want 0", code)
			}
			if out.Len() != 0 {
				t.Errorf("nobody is asked a question that would be refused, got %q", out.String())
			}
			if len(q.calls) != 2 {
				t.Fatalf("want the listing and the probe only, got %+v", q.calls)
			}
			if want := []string{"kubectl", "auth", "can-i", "delete", "namespaces", "--all-namespaces"}; !reflect.DeepEqual(q.calls[1].argv, want) {
				t.Errorf("probe argv = %v, want %v", q.calls[1].argv, want)
			}
			for _, want := range []string{"warning:", "left in place", c.want} {
				if !strings.Contains(stderr, want) {
					t.Errorf("stderr lacks %q:\n%s", want, stderr)
				}
			}
		})
	}
}

// TestRemoveNamespaceDeleteOutcome: a confirmed delete reports ok:, and one that
// fails is a warning that keeps the namespace -- the remove already succeeded, so the
// exit stays 0.
func TestRemoveNamespaceDeleteOutcome(t *testing.T) {
	q := &queueRunner{resp: []queuedResp{{nsList(), nil}, {"yes\n", nil}, {"namespace \"prod\" deleted", nil}}}
	var code int
	stderr := captureStderr(t, func() {
		code = removeNamespace(q, actionOpts{action: "remove", noPrompt: true}, gateKube("prod"), gateBase)
	})
	if code != 0 || !strings.Contains(stderr, "ok: remove namespace kubernetes") {
		t.Errorf("a deleted namespace reports ok, exit %d:\n%s", code, stderr)
	}

	q = &queueRunner{resp: []queuedResp{{nsList(), nil}, {"yes\n", nil}, {"Error from server (Forbidden): namespaces \"prod\" is forbidden", errors.New("exit status 1")}}}
	stderr = captureStderr(t, func() {
		code = removeNamespace(q, actionOpts{action: "remove", noPrompt: true}, gateKube("prod"), gateBase)
	})
	if code != 0 {
		t.Fatalf("a failed namespace delete must not fail the remove: exit %d\n%s", code, stderr)
	}
	for _, want := range []string{"warning: namespace prod was left in place, since deleting it failed", "is forbidden"} {
		if !strings.Contains(stderr, want) {
			t.Errorf("stderr lacks %q:\n%s", want, stderr)
		}
	}
	if strings.Contains(stderr, "ok:") || strings.Contains(stderr, "error:") {
		t.Errorf("a kept namespace is neither ok nor an error:\n%s", stderr)
	}
}

// TestRemoveNamespaceProbeArgvIsOneQuery pins the shape of the occupancy query:
// one call listing every kind whose loss would matter, in the namespace being
// considered (nsgate.List).
func TestRemoveNamespaceProbeArgvIsOneQuery(t *testing.T) {
	q := &queueRunner{resp: []queuedResp{{nsList(nsItemJSON("Deployment", "other")), nil}}}

	captureStderr(t, func() { removeNamespace(q, actionOpts{action: "remove", noPrompt: true}, gateKube("prod"), gateBase) })
	if len(q.calls) != 1 {
		t.Fatalf("want 1 probe call, got %d", len(q.calls))
	}
	want := []string{"kubectl", "get", nsListTypes, "-n", "prod", "-o", "json"}
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
	if code := run([]string{"validate", "-e", filepath.Join(invalidDir, "env.yaml")}); code != 2 {
		t.Errorf("invalid spec should fail validate, exit=%d", code)
	}
}

// validate makes the checks a docker deploy makes on its host: each
// bind-mounted store that is missing, or a directory where its PKCS12 belongs,
// is an error, every one reported in the same run, and nothing is executed.
func TestValidateChecksDockerHost(t *testing.T) {
	f := useFakeRunner(t)
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "certs", "ks.p12"), 0o755); err != nil {
		t.Fatal(err)
	}
	stores := "\ntls:\n  stores:\n    ts:\n      password: ts\n      existing:\n        pkcs12: ./certs/missing.p12\n" +
		"    ks:\n      password: ks\n      existing:\n        pkcs12: ./certs/ks.p12\n"
	write(t, dir, "env.yaml", sharedEnv+stores+dockerEnv)
	write(t, dir, "10.yaml", strings.Replace(tlsWF, "    truststore: ts\n", "    truststore: ts\n    keystore: ks\n", 1))
	var code int
	stderr := captureStderr(t, func() {
		code = dispatch([]string{"validate", "-e", filepath.Join(dir, "env.yaml")}, f)
	})
	for _, want := range []string{
		"docker mounts tls.stores.ts.existing.pkcs12 " + filepath.Join(dir, "certs", "missing.p12") + ", which does not exist",
		"docker mounts tls.stores.ks.existing.pkcs12 " + filepath.Join(dir, "certs", "ks.p12") + ", which is a directory",
	} {
		if !strings.Contains(stderr, want) {
			t.Errorf("validate does not report %q:\n%s", want, stderr)
		}
	}
	if code != 1 {
		t.Errorf("exit = %d, want 1", code)
	}
	if len(f.calls) != 0 {
		t.Errorf("validate must run nothing, got %+v", f.calls)
	}
}

// validate warns, as deploy does, about each host mount the connector could
// not read: a 0600 docker truststore owned by another uid than its run-as, and
// a rootless podman libs.dir only its owner may enter. Warnings alone exit 0,
// and nothing is executed. Linux only, like the deploy check.
func TestValidateWarnsAboutHostMountsTheConnectorCannotRead(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("the mount check runs on Linux only")
	}
	f := useFakeRunner(t)
	podmanQuadletHome(t) // podman is rootless, as an ordinary user
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, "certs"), 0o755); err != nil {
		t.Fatal(err)
	}
	truststore := filepath.Join(dir, "certs", "truststore.p12")
	// A real store its password opens: validate opens every store podman
	// loads, and this test is about the mode, not the content.
	write(t, filepath.Join(dir, "certs"), "truststore.p12", string(certfixture.TrustStorePKCS12(t, "ts", certfixture.CA(t, "Root"))))
	if err := os.Chmod(truststore, 0o600); err != nil {
		t.Fatal(err)
	}
	libs := filepath.Join(dir, "libs")
	if err := os.Mkdir(libs, 0o700); err != nil {
		t.Fatal(err)
	}
	runAs := fmt.Sprintf("%d:%d", os.Getuid()+1, os.Getgid()+1)
	write(t, dir, "env.yaml", sharedEnv+"\ntls:\n  stores:\n    ts:\n      password: ts\n      existing:\n        pkcs12: ./certs/truststore.p12\n"+
		dockerEnv+"  run-as: \""+runAs+"\"\n"+podmanEnv()+"  libs:\n    dir: ./libs\n")
	write(t, dir, "10.yaml", tlsWF)
	var code int
	stderr := captureStderr(t, func() {
		code = dispatch([]string{"validate", "-e", filepath.Join(dir, "env.yaml")}, f)
	})
	if code != 0 {
		t.Fatalf("exit=%d, want warnings only:\n%s", code, stderr)
	}
	for _, want := range []string{
		"docker mounts tls.stores.ts.existing.pkcs12 " + truststore + " (mode 0600) is not readable by the connector, which runs as " + runAs,
		"podman mounts libs.dir " + libs + " (mode 0700) is not readable by the connector",
	} {
		if !strings.Contains(stderr, want) {
			t.Errorf("validate does not warn %q:\n%s", want, stderr)
		}
	}
	if len(f.calls) != 0 {
		t.Errorf("validate must run nothing, got %+v", f.calls)
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
		"SOLACE_TRUSTSTORE_PASSWORD", "SOLACE_KEYSTORE_PASSWORD", "MQ_TRUSTSTORE_PASSWORD", "MQ_KEYSTORE_PASSWORD",
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
// annotated "(built in; describes <range>)" from Report.OmitListRange for a
// built-in list (Report.OmitListProvenance then names it), printed bare when
// an explicit path was given, and any per-line Report.OmitListWarnings
// (entries the omit list loader rejected as unparseable) surfaced as their
// own lines.
//
// The annotation carries the range because the filename alone names one
// tag, and an operator reading "2.13.0" while deploying 2.14.1 would
// reasonably conclude the omissions were judged against the wrong image.
func TestDownloadReportPrintsOmitListProvenance(t *testing.T) {
	t.Run("built in", func(t *testing.T) {
		f := &fakeDownload{resp: libs.Report{
			Written:            []string{"libs/a.jar"},
			OmitListProvenance: "solace-pubsub-connector-ibmmq-2.13.0",
			OmitListRange:      "2.10.0 and later, before 3.0.0",
		}}
		useFakeDownload(t, f)
		stderr := captureStderr(t, func() { dispatch([]string{"download", "jar", "mq"}, &fakeRunner{}) })
		want := "omit list: solace-pubsub-connector-ibmmq-2.13.0 (built in; describes 2.10.0 and later, before 3.0.0)"
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

// TestDownloadReportPrintsSeedChoice pins the "seed:" line: why the syslog
// encoder resolved to the release it did, printed when the connector line
// picked it and absent otherwise, so a pinned --version or the mq set reads
// as it always has.
func TestDownloadReportPrintsSeedChoice(t *testing.T) {
	const why = "logstash-logback-encoder 8.1, the newest 8.x: connector 2.14.1 ships Jackson 2"
	useFakeDownload(t, &fakeDownload{resp: libs.Report{Written: []string{"libs/a.jar"}, SeedChoice: why}})
	stderr := captureStderr(t, func() { dispatch([]string{"download", "jar", "syslog"}, &fakeRunner{}) })
	if !strings.Contains(stderr, "seed: "+why+"\n") {
		t.Errorf("stderr missing the seed line, got:\n%s", stderr)
	}

	useFakeDownload(t, &fakeDownload{resp: libs.Report{Written: []string{"libs/a.jar"}}})
	stderr = captureStderr(t, func() { dispatch([]string{"download", "jar", "mq"}, &fakeRunner{}) })
	if strings.Contains(stderr, "seed:") {
		t.Errorf("a report with no SeedChoice printed a seed line:\n%s", stderr)
	}
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
		if code := dispatch(args, &fakeRunner{}); code != 2 {
			t.Errorf("exit=%d, want 2: the operator named this file", code)
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

// testHost is the host a podman test runs on unless it says otherwise: an
// ordinary user (euid 1000, "alice") whose home is home, with an empty
// environment, no session to repair (not Linux, so no linger or subordinate-id
// probe) and no subuid or subgid file. It reads nothing from the real process, so
// a podman test passes the same as root, as another user, or on Windows.
func testHost(home string) hostFacts {
	return hostFacts{
		scope: runner.ScopeEnv{
			Euid:   1000,
			Getenv: func(string) string { return "" },
			Home:   func() (string, error) { return home, nil },
			IsDir:  func(string) (bool, error) { return true, nil },
			Linux:  false,
		},
		readFile: func(string) ([]byte, error) { return nil, os.ErrNotExist },
		userName: func() string { return "alice" },
	}
}

// useHost installs h as the host the podman paths read for the length of the test.
func useHost(t *testing.T, h hostFacts) {
	t.Helper()
	prev := host
	host = func() hostFacts { return h }
	t.Cleanup(func() { host = prev })
}

func init() {
	// Whatever a test does not set up for itself runs as an ordinary user with
	// no session, never as the euid of the machine the suite is on.
	host = func() hostFacts {
		h := liveHost()
		h.scope.Euid, h.scope.Linux = 1000, false
		return h
	}
}

// podmanQuadletHome runs the test as an ordinary user whose home is a temp
// directory (testHost) and returns the unit directory the scope derives from it.
//
// There is no config key for the unit directory -- it is $XDG_CONFIG_HOME or
// $HOME/.config -- so a test keeps its units out of the developer's real
// ~/.config by redirecting the home, through the host seam and, for whatever
// else reads it, the environment, with XDG_CONFIG_HOME cleared so a developer's
// own does not move the directory. It no longer depends on the euid of the
// process, so nothing is skipped as root.
func podmanQuadletHome(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)        // unix
	t.Setenv("USERPROFILE", home) // windows
	t.Setenv("XDG_CONFIG_HOME", "")
	useHost(t, testHost(home))
	return filepath.Join(home, ".config", "containers", "systemd")
}

// podmanEnv renders a minimal-but-valid podman: section. It names no directory:
// the unit's comes from the invoking user (see podmanQuadletHome), and
// everything else the unit mounts comes from podman's secret store.
func podmanEnv() string {
	return `
podman:
  command: podman
  name: solmq-conn
  ports:
    - 8090
`
}

// podmanEnvChained mirrors podmanEnv but with a chained `lima podman` command:
// rejected outright (lima is not on the podman allowlist) unless the caller
// approves it via --allow-command lima, which is the escape-hatch scenario
// the flag exists for.
func podmanEnvChained() string {
	return `
podman:
  command: lima podman
  name: solmq-conn
  ports:
    - 8090
`
}

// podmanEnvSudo mirrors podmanEnv but with `sudo podman`: a privilege-escalation
// wrapper, refused whatever --allow-command says. Run `sudo solmq-conn-util`
// instead.
func podmanEnvSudo() string {
	return `
podman:
  command: sudo podman
  name: solmq-conn
  ports:
    - 8090
`
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
	if want := "name: \"" + spec.DefaultComposeProject + "\"\n"; !strings.HasPrefix(string(b), want) {
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
	write(t, dir, "env.yaml", sharedEnv+podmanEnv())
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

// TestGeneratePodmanOnlyLibsIsAHostPath pins the one host path a podman unit
// still names, against the way an operator actually invokes the tool: `-e
// env.yaml` from the file's own directory, which used to leave envDir as "."
// and emit a relative source.
//
// A relative source is not a near-miss in a quadlet unit. systemd starts the unit
// with no useful cwd, and podman reads a source with no ./ or / prefix as a NAMED
// VOLUME -- so `Volume=libs:/app/external/libs:ro` silently mounts an empty volume
// over the jars instead of failing, and the connector dies looking for classes
// that were never mounted. The relative spelling has to be impossible to emit,
// which is why this asserts on the rendered unit rather than on absPath. The
// store, selected in the same spec, must not be a Volume= at all: it comes from
// podman's secret store.
func TestGeneratePodmanOnlyLibsIsAHostPath(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "env.yaml", sharedEnv+`
tls:
  stores:
    ts:
      existing:
        pkcs12: ./certs/truststore.jks
      password: ts
`+podmanEnv()+"  libs:\n    dir: ./libs\n")
	write(t, dir, "10.yaml", tlsWF)
	t.Chdir(dir)

	var code int
	stdout := captureStdout(t, func() {
		code = run([]string{"generate", "--platform", "podman", "-e", "env.yaml"})
	})
	if code != 0 {
		t.Fatalf("exit=%d", code)
	}
	var volumes []string
	for _, line := range strings.Split(stdout, "\n") {
		if strings.HasPrefix(line, "Volume=") {
			volumes = append(volumes, line)
		}
	}
	if want := "Volume=" + filepath.Join(dir, "libs") + ":/app/external/libs:ro"; len(volumes) != 1 || volumes[0] != want {
		t.Errorf("Volume= lines = %q, want only %q", volumes, want)
	}
	if want := "Secret=solmq-conn-tls-ts,type=mount,target=/app/external/classpath/truststores/ts\n"; !strings.Contains(stdout, want) {
		t.Errorf("missing %q in:\n%s", want, stdout)
	}
}

// TestDeployPodmanUnresolvedInputFailsBeforeAnyWrite pins where deploy
// resolves what the unit mounts -- the credentials' variables and the TLS
// stores -- before anything with a side effect outside this process: secrets
// placed in podman's store, then the unit, then systemctl. A missing variable
// or store found late would leave a half-built deployment behind.
//
// So each case asserts exit 2 naming what is missing, no runner call
// beyond the read-only preflight probe, and nothing written to the quadlet
// directory.
func TestDeployPodmanUnresolvedInputFailsBeforeAnyWrite(t *testing.T) {
	const unsetVar = "SOLMQ_TEST_PODMAN_UNSET_PW"
	for _, c := range []struct {
		name, env, wf, want string
	}{
		{"unreadable store", sharedEnv + `
tls:
  stores:
    ts:
      existing:
        pkcs12: ./certs/missing.jks
      password: ts
` + podmanEnv(), tlsWF, `tls.stores.ts.existing.pkcs12 "./certs/missing.jks"`},
		{"unset credential variable", sharedEnv + podmanEnv(),
			strings.Replace(validWF, "    password: pw\n", "    password-env: "+unsetVar+"\n", 1),
			"environment variable " + unsetVar + ", which is not set"},
	} {
		t.Run(c.name, func(t *testing.T) {
			// t.Setenv restores whatever was there; Unsetenv makes sure the
			// variable is absent for the deploy, whatever the shell exported.
			t.Setenv(unsetVar, "")
			os.Unsetenv(unsetVar)
			f := useFakeRunner(t)
			dir := t.TempDir()
			quadletDir := podmanQuadletHome(t)
			write(t, dir, "env.yaml", c.env)
			write(t, dir, "10.yaml", c.wf)

			var code int
			stderr := captureStderr(t, func() {
				code = dispatch([]string{"deploy", "--platform", "podman", "-e", filepath.Join(dir, "env.yaml")}, f)
			})
			if code != 2 {
				t.Fatalf("exit=%d, want 2: a deploy with an unresolved input is refused for what env.yaml names", code)
			}
			if !strings.Contains(stderr, c.want) {
				t.Errorf("the error should name %q, got %q", c.want, stderr)
			}
			if len(f.calls) != 1 || !reflect.DeepEqual(f.calls[0].argv, []string{"podman", "info"}) {
				t.Errorf("only the preflight probe may run before the inputs are resolved, got %+v", f.calls)
			}
			// The directory is only created by a write, so "does not exist" is
			// the strongest form of "nothing was written".
			ents, rerr := os.ReadDir(quadletDir)
			if rerr != nil && !os.IsNotExist(rerr) {
				t.Fatal(rerr)
			}
			if len(ents) != 0 {
				t.Errorf("no file should be written on a rejected deploy, got %d", len(ents))
			}
		})
	}
}

// TestDeployPodmanSetButEmptyCredentialFailsBeforeAnyWrite pins the empty half
// of the -env rule: a variable that is set but empty is refused at deploy like an
// unset one (exit 2, only the preflight probe run, nothing written), and validate
// warns about it, since a blank password is worse than refusing to deploy.
func TestDeployPodmanSetButEmptyCredentialFailsBeforeAnyWrite(t *testing.T) {
	const emptyVar = "SOLMQ_TEST_PODMAN_EMPTY_PW"
	t.Setenv(emptyVar, "")
	f := useFakeRunner(t)
	dir := t.TempDir()
	quadletDir := podmanQuadletHome(t)
	write(t, dir, "env.yaml", sharedEnv+podmanEnv())
	write(t, dir, "10.yaml", strings.Replace(validWF, "    password: pw\n", "    password-env: "+emptyVar+"\n", 1))
	envPath := filepath.Join(dir, "env.yaml")

	var code int
	stderr := captureStderr(t, func() {
		code = dispatch([]string{"deploy", "--platform", "podman", "-e", envPath}, f)
	})
	if code != 2 {
		t.Fatalf("exit=%d, want 2: a set but empty credential variable is refused like an unset one", code)
	}
	if want := "environment variable " + emptyVar + ", which is set but empty"; !strings.Contains(stderr, want) {
		t.Errorf("the error should say %q, got %q", want, stderr)
	}
	if len(f.calls) != 1 || !reflect.DeepEqual(f.calls[0].argv, []string{"podman", "info"}) {
		t.Errorf("only the preflight probe may run before the inputs are resolved, got %+v", f.calls)
	}
	if ents, rerr := os.ReadDir(quadletDir); rerr != nil && !os.IsNotExist(rerr) {
		t.Fatal(rerr)
	} else if len(ents) != 0 {
		t.Errorf("no file should be written on a rejected deploy, got %d", len(ents))
	}

	vstderr := captureStderr(t, func() { run([]string{"validate", "-e", envPath}) })
	if want := "environment variable " + emptyVar + ", which is set but empty; export it before running"; !strings.Contains(vstderr, want) {
		t.Errorf("validate should warn %q, got %q", want, vstderr)
	}
}

// TestDeployPodmanFailureStopsBeforeSystemctl pins that a failing step --
// storing a credential, storing a document, deleting a file secret the spec no
// longer mounts, or writing the unit -- stops the deploy where it fails, and
// systemctl never reloads or starts anything: the unit would mount a secret that
// is missing or stale, or be the one left from the previous deploy. With validWF
// and no stores or syslog the calls run: 0 preflight, 1 the podman version (the
// fake's "out" is no version, so no value is read back), 2 the state probe, 3-10
// the four credentials (rm, create), 11-14 application.yml and the status script,
// 15 the one batched rm of the three unmounted file secrets; the unit is written
// after that.
func TestDeployPodmanFailureStopsBeforeSystemctl(t *testing.T) {
	for _, c := range []struct {
		name      string
		failFrom  int // the fake fails every call from this index on; -1 never
		blockUnit bool
		want      string
		wantCalls int
	}{
		{"credential", 3, false, "podman secret rm solmq-conn-_GEN_MQ_CONN_1_USER: boom", 4},
		{"document", 11, false, "podman secret rm solmq-conn-application.yml: boom", 12},
		{"unmounted file secret", 15, false, "podman secret rm: boom", 16},
		{"unit write", -1, true, "creating ", 16},
	} {
		t.Run(c.name, func(t *testing.T) {
			f := useFakeRunner(t)
			if c.failFrom >= 0 {
				f.err = fmt.Errorf("boom")
				f.failFrom = c.failFrom
			}
			quadletDir := podmanQuadletHome(t)
			want := c.want
			if c.blockUnit {
				// A file where ~/.config should be makes the unit directory
				// impossible to create, on every OS.
				if err := os.WriteFile(filepath.Dir(filepath.Dir(quadletDir)), []byte("x"), 0o600); err != nil {
					t.Fatal(err)
				}
				want += quadletDir
			}
			dir := t.TempDir()
			write(t, dir, "env.yaml", sharedEnv+podmanEnv())
			write(t, dir, "10.yaml", validWF)

			var code int
			stderr := captureStderr(t, func() {
				code = dispatch([]string{"deploy", "--platform", "podman", "-e", filepath.Join(dir, "env.yaml")}, f)
			})
			if code != 1 {
				t.Fatalf("exit=%d, want 1", code)
			}
			if !strings.Contains(stderr, want) {
				t.Errorf("the error should name the failed step %q, got %q", want, stderr)
			}
			if len(f.calls) != c.wantCalls {
				t.Errorf("deploy must stop after %d calls, got %d: %+v", c.wantCalls, len(f.calls), f.calls)
			}
			for _, call := range f.calls {
				// The state probe is read-only and runs before anything is written.
				if call.argv[0] == "systemctl" && !containsToken(call.argv, "is-active") {
					t.Errorf("systemctl must not reload or start after a failed step, got %v", call.argv)
				}
			}
			if _, err := os.Stat(filepath.Join(quadletDir, "solmq-conn.container")); err == nil {
				t.Errorf("the unit must not be written after a failed step")
			}
		})
	}
}

// ---- deploy / remove: docker and podman seams -------------------------------------

// TestDeployDockerSeamPipesComposeAndRuns is CR-0003: the compose file inlines
// application.yml and the status account's password, so it reaches compose on
// stdin and never the disk -- it used to sit beside env.yaml at 0644, kept
// whenever compose failed, where the next run read it as a workflow (the
// default workflows.dir is env.yaml's own).
func TestDeployDockerSeamPipesComposeAndRuns(t *testing.T) {
	f := useFakeRunner(t)
	dir := t.TempDir()
	write(t, dir, "env.yaml", sharedEnv+dockerEnv)
	write(t, dir, "10.yaml", validWF)

	if code := dispatch([]string{"deploy", "--platform", "docker", "-e", filepath.Join(dir, "env.yaml")}, f); code != 0 {
		t.Fatalf("exit=%d", code)
	}
	if _, err := os.Stat(filepath.Join(dir, "docker-compose.yml")); !os.IsNotExist(err) {
		t.Errorf("deploy must write no compose file, stat err=%v", err)
	}
	// Call 0 is the read-only preflight probe (runner.Preflight), call 1 the
	// listing that says whether the container is running (it is not: the fake
	// answers no row), call 2 compose up.
	if len(f.calls) != 3 {
		t.Fatalf("want 3 runner calls (preflight, ps, up), got %d: %+v", len(f.calls), f.calls)
	}
	if want := []string{"docker", "info"}; !reflect.DeepEqual(f.calls[0].argv, want) {
		t.Errorf("preflight argv = %v, want %v", f.calls[0].argv, want)
	}
	if want := []string{"docker", "ps", "--all", "--no-trunc", "--filter", "status=running", "--format", "{{.Names}}\t{{.Image}}\t{{.Status}}"}; !reflect.DeepEqual(f.calls[1].argv, want) {
		t.Errorf("running probe argv = %v, want %v", f.calls[1].argv, want)
	}
	// No -p/--project-name: the compose project is declared by the file's own
	// top-level name: key instead. Pinned here so the choice is a recorded one
	// rather than an omission. --project-directory is env.yaml's directory, the
	// one compose took from the file's location when it was written there.
	want := []string{"docker", "compose", "-f", "-", "--project-directory", dir, "up", "-d"}
	if !reflect.DeepEqual(f.calls[2].argv, want) {
		t.Errorf("argv = %v, want %v", f.calls[2].argv, want)
	}
	in := f.calls[2].stdin
	if !strings.Contains(in, "\nservices:\n") || !strings.Contains(in, "content: |") {
		t.Errorf("the compose file must reach compose on stdin, got:\n%s", in)
	}
	// The file carries the deploy-digest label, and its value is the digest of the
	// file and of the child's environment taken before the variable was added.
	if !strings.Contains(in, "      "+spec.LabelDeployDigestKey+": ${"+dockergen.DeployDigestVar+":-}\n") {
		t.Errorf("the compose file lacks the deploy digest label:\n%s", in)
	}
	env := f.calls[2].env
	digest, ok := envValue(env, dockergen.DeployDigestVar)
	if !ok || env[len(env)-1] != dockergen.DeployDigestVar+"="+digest {
		t.Fatalf("the digest is the last entry of the child env, got %v", env)
	}
	if want := containercli.ComposeDigest([]byte(in), env[:len(env)-1]); digest != want {
		t.Errorf("deploy digest = %q, want ComposeDigest of the file and the rest of the env %q", digest, want)
	}
}

// TestDeployDockerSeamFailedRunLeavesNoComposeFile: a failed `up` used to keep
// the compose file -- password and all -- for a `down` that remove renders
// afresh anyway. failFrom: 2 lets the preflight probe (call 0) and the running
// probe (call 1) succeed and fails only the real `up` call (call 2).
func TestDeployDockerSeamFailedRunLeavesNoComposeFile(t *testing.T) {
	f := useFakeRunner(t)
	f.err = fmt.Errorf("boom")
	f.failFrom = 2
	dir := t.TempDir()
	write(t, dir, "env.yaml", sharedEnv+dockerEnv)
	write(t, dir, "10.yaml", validWF)

	if code := dispatch([]string{"deploy", "--platform", "docker", "-e", filepath.Join(dir, "env.yaml")}, f); code != 1 {
		t.Fatalf("exit=%d, want 1 for a failed deploy", code)
	}
	if _, err := os.Stat(filepath.Join(dir, "docker-compose.yml")); !os.IsNotExist(err) {
		t.Errorf("a failed run must leave no compose file behind, stat err=%v", err)
	}
	if len(f.calls) != 3 || f.calls[2].stdin == "" {
		t.Errorf("the failed up still had its file on stdin: %+v", f.calls)
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
	// Call 0 is the preflight probe (no Env), call 1 the running probe; call 2 is
	// the real `up`, which carries the resolved credentials and, last, the deploy
	// digest.
	if len(f.calls) != 3 {
		t.Fatalf("want 3 runner calls, got %d: %+v", len(f.calls), f.calls)
	}
	want := []string{
		"_GEN_MQ_CONN_1_USER=app",
		"_GEN_MQ_CONN_1_PASSWORD=pw",
		"SOLACE_USER=envuser",
		"SOLACE_PASS=envpass",
	}
	env := f.calls[2].env
	if len(env) != len(want)+1 || !reflect.DeepEqual(env[:len(want)], want) || !strings.HasPrefix(env[len(want)], "_GEN_DEPLOY_SHA256=") {
		t.Errorf("child env = %v, want %v and then the deploy digest", env, want)
	}
}

// TestDeployDockerWarnsAboutMountsTheConnectorCannotRead pins the deploy-time
// readability check on docker: a 0600 truststore and a 0600 jar under libs.dir,
// owned by a uid and gid other than the run-as ones, are each named on stderr
// with their mode, and the deploy still goes ahead. Both paths are written with
// ${VAR}s, so the warning has to name the expanded path the compose file
// mounts. Linux only: elsewhere Docker Desktop's file sharing does not carry
// the host's modes into the container, so nothing is checked.
func TestDeployDockerWarnsAboutMountsTheConnectorCannotRead(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("the mount check runs on Linux only")
	}
	f := useFakeRunner(t)
	dir := t.TempDir()
	libs := filepath.Join(dir, "libs")
	for _, d := range []string{filepath.Join(dir, "certs"), libs} {
		if err := os.Mkdir(d, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.Chmod(d, 0o755); err != nil { // umask aside
			t.Fatal(err)
		}
	}
	truststore := filepath.Join(dir, "certs", "truststore.jks")
	// A real store its password opens, since deploy opens it before the
	// digest; the mode is what this test is about.
	write(t, filepath.Join(dir, "certs"), "truststore.jks", string(certfixture.TrustStorePKCS12(t, "ts", certfixture.CA(t, "Root"))))
	if err := os.Chmod(truststore, 0o600); err != nil {
		t.Fatal(err)
	}
	jar := filepath.Join(libs, "com.ibm.mq.jakarta.client.jar")
	writeMode(t, jar, 0o600)
	runAs := fmt.Sprintf("%d:%d", os.Getuid()+1, os.Getgid()+1)
	t.Setenv("SOLMQ_TEST_CERTS", filepath.Join(dir, "certs"))
	t.Setenv("SOLMQ_TEST_LIBS", libs)
	write(t, dir, "env.yaml", sharedEnv+`
tls:
  stores:
    ts:
      existing:
        pkcs12: ${SOLMQ_TEST_CERTS}/truststore.jks
      password: ts
`+dockerEnv+"  libs:\n    dir: ${SOLMQ_TEST_LIBS}\n  run-as: \""+runAs+"\"\n")
	write(t, dir, "10.yaml", tlsWF)

	var code int
	stderr := captureStderr(t, func() {
		code = dispatch([]string{"deploy", "--platform", "docker", "-e", filepath.Join(dir, "env.yaml")}, f)
	})
	if code != 0 {
		t.Fatalf("exit=%d, want the deploy to go ahead:\n%s", code, stderr)
	}
	for _, want := range []string{
		"warning: docker mounts tls.stores.ts.existing.pkcs12 " + truststore + " (mode 0600) is not readable by the connector, which runs as " + runAs,
		"warning: docker mounts libs.dir jar " + jar + " (mode 0600) is not readable by the connector, which runs as " + runAs,
	} {
		if !strings.Contains(stderr, want) {
			t.Errorf("stderr should carry %q, got:\n%s", want, stderr)
		}
	}
	if len(f.calls) != 3 {
		t.Errorf("want the preflight, the running probe and the up call, got %+v", f.calls)
	}
}

func TestRemoveDockerSeam(t *testing.T) {
	f := useFakeRunner(t)
	// remove asks before it tears down; this test is about the argv that
	// follows, so it answers yes. The prompt itself is covered in section g.
	withPrompter(t, true, "y\n")
	dir := t.TempDir()
	write(t, dir, "env.yaml", sharedEnv+dockerEnv)
	write(t, dir, "10.yaml", validWF)

	if code := dispatch([]string{"remove", "--platform", "docker", "-e", filepath.Join(dir, "env.yaml")}, f); code != 0 {
		t.Fatalf("exit=%d", code)
	}
	wantCalls := [][]string{
		{"docker", "info"},
		{"docker", "compose", "-f", "-", "--project-directory", dir, "down"},
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
	quadletDir := podmanQuadletHome(t)
	legacyDir := filepath.Join(t.TempDir(), "old-base-dir")
	dir := t.TempDir()
	// A binary truststore -- a real PKCS12, which deploy opens with its
	// password before loading it -- to show the bytes reach podman unchanged,
	// and a leftover base-dir, to show it is ignored.
	truststore := string(certfixture.TrustStorePKCS12(t, "ts", certfixture.CA(t, "Root")))
	if err := os.MkdirAll(filepath.Join(dir, "certs"), 0o755); err != nil {
		t.Fatal(err)
	}
	write(t, filepath.Join(dir, "certs"), "truststore.p12", truststore)
	write(t, dir, "env.yaml", sharedEnv+`
tls:
  stores:
    ts:
      existing:
        pkcs12: ./certs/truststore.p12
      password: ts
`+podmanEnv()+"  base-dir: "+filepath.ToSlash(legacyDir)+"\n")
	write(t, dir, "10.yaml", tlsWF)
	// The unit an earlier deploy left, mounting a store this spec no longer
	// selects: deploy reads it before replacing it and deletes that secret.
	if err := os.MkdirAll(quadletDir, 0o755); err != nil {
		t.Fatal(err)
	}
	// It is 0644 as deploy writes it: os.WriteFile keeps an existing file's mode.
	write(t, quadletDir, "solmq-conn.container", "[Container]\nSecret=solmq-conn-tls-old-ca,type=mount,target=/app/external/classpath/truststores/old-ca\n")
	if err := os.Chmod(filepath.Join(quadletDir, "solmq-conn.container"), 0o644); err != nil {
		t.Fatal(err)
	}

	if code := dispatch([]string{"deploy", "--platform", "podman", "-e", filepath.Join(dir, "env.yaml")}, f); code != 0 {
		t.Fatalf("exit=%d", code)
	}
	// The unit is the one file deploy writes, to the quadlet dir the invoking
	// user resolves to; it carries only secret names, so 0644. Nothing goes to
	// the old base-dir, which is not even created.
	info, err := os.Stat(filepath.Join(quadletDir, "solmq-conn.container"))
	if err != nil {
		t.Fatalf("unit not written to %s: %v", quadletDir, err)
	}
	// Unix perms are not faithfully reproduced on the windows-2025 CI runner.
	if runtime.GOOS != "windows" && info.Mode().Perm() != 0o644 {
		t.Errorf("unit mode = %v, want 0644", info.Mode().Perm())
	}
	if ents, _ := os.ReadDir(quadletDir); len(ents) != 1 {
		t.Errorf("the quadlet dir should hold only the unit, got %d entries", len(ents))
	}
	if _, err := os.Stat(legacyDir); !os.IsNotExist(err) {
		t.Errorf("deploy must not write to podman.base-dir any more: %v", err)
	}
	// tlsWF's five credentials (mq user/password, the ts store's literal
	// password, solace client-username/client-password), then
	// application.yml, the ts store and the status script. Call 0 is the read-only
	// preflight probe. Then the version is read BEFORE anything is compared; the
	// fake answers every call with "out", which is not a version (so no value is
	// read back and every secret counts as changed) and not a running state (so
	// the unit is started); the unit's `is-active` follows. Only then is each
	// secret stored (secret rm --ignore and secret create, no label) BEFORE
	// daemon-reload/start: the unit being started references these secrets by
	// name, so they must already exist. The file secrets this spec does not mount
	// (no syslog, and the old-ca store the previous unit mounted) are deleted in
	// one call.
	names := []string{
		"solmq-conn-_GEN_MQ_CONN_1_USER", "solmq-conn-_GEN_MQ_CONN_1_PASSWORD", "solmq-conn-_GEN_TLS_TS_PASSWORD",
		"solmq-conn-_GEN_SOL_CONN_1_CLIENT_USERNAME", "solmq-conn-_GEN_SOL_CONN_1_CLIENT_PASSWORD",
		"solmq-conn-application.yml", "solmq-conn-tls-ts", "solmq-conn-status-script",
	}
	wantCalls := [][]string{
		{"podman", "info"},
		{"podman", "version", "--format", "{{.Client.Version}}"},
		{"systemctl", "--user", "is-active", "solmq-conn.service"},
	}
	first := len(wantCalls)
	for _, n := range names {
		wantCalls = append(wantCalls, []string{"podman", "secret", "rm", "--ignore", n}, []string{"podman", "secret", "create", n, "-"})
	}
	wantCalls = append(wantCalls,
		[]string{"podman", "secret", "rm", "--ignore", "solmq-conn-logback-spring.xml", "solmq-conn-tls-old-ca"},
		[]string{"systemctl", "--user", "daemon-reload"},
		[]string{"systemctl", "--user", "start", "solmq-conn.service"},
	)
	if len(f.calls) != len(wantCalls) {
		t.Fatalf("want %d runner calls, got %+v", len(wantCalls), f.calls)
	}
	for i, w := range wantCalls {
		if !reflect.DeepEqual(f.calls[i].argv, w) {
			t.Errorf("call %d argv = %v, want %v", i, f.calls[i].argv, w)
		}
	}
	// The values travel on stdin, never argv: the rendered application.yml, and
	// the truststore byte for byte.
	if app := f.calls[first+2*5+1].stdin; !strings.Contains(app, "solace:") || !strings.Contains(app, "conn-name: mqhost(1414)") {
		t.Errorf("application.yml secret should carry the rendered document, got %q", app)
	}
	if got := f.calls[first+2*6+1].stdin; got != truststore {
		t.Errorf("truststore secret = %q, want the file's bytes %q", got, truststore)
	}
}

// TestDeployPodmanRootlessWarnsAboutAnUnlistableLibsDir pins the deploy-time
// check on rootless podman: the default run-as 1000:0 maps to no host uid and
// to the invoking user's own group, so a libs.dir only its owner may enter
// (0700) is named on stderr -- at its expanded path -- with the remedy, and the
// deploy still goes ahead. Linux only, like the docker check.
func TestDeployPodmanRootlessWarnsAboutAnUnlistableLibsDir(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("the mount check runs on Linux only")
	}
	f := useFakeRunner(t)
	podmanQuadletHome(t) // an ordinary user, so this deploy is rootless
	dir := t.TempDir()
	libs := filepath.Join(dir, "libs")
	if err := os.Mkdir(libs, 0o700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("SOLMQ_TEST_LIBS", libs)
	write(t, dir, "env.yaml", sharedEnv+podmanEnv()+"  libs:\n    dir: ${SOLMQ_TEST_LIBS}\n")
	write(t, dir, "10.yaml", validWF)

	var code int
	stderr := captureStderr(t, func() {
		code = dispatch([]string{"deploy", "--platform", "podman", "-e", filepath.Join(dir, "env.yaml")}, f)
	})
	if code != 0 {
		t.Fatalf("exit=%d, want the deploy to go ahead:\n%s", code, stderr)
	}
	want := "warning: podman mounts libs.dir " + libs + " (mode 0700) is not readable by the connector, which runs as 1000:0 with no capability to bypass file permissions: it would start and then fail when it needs it. Make it listable: chmod o+rx, or give it to group " + fmt.Sprint(os.Getgid()) + " with g+rx -- or change podman.run-as"
	if !strings.Contains(stderr, want) {
		t.Errorf("stderr should carry %q, got:\n%s", want, stderr)
	}
}

func TestRemovePodmanSeamStopsRemovesReloads(t *testing.T) {
	f := useFakeRunner(t)
	quadletDir := podmanQuadletHome(t)
	legacyDir := t.TempDir()
	dir := t.TempDir()
	// remove asks before it tears down; this test is about the argv that
	// follows, so it answers yes. The prompt itself is covered in section g.
	withPrompter(t, true, "y\n")
	write(t, dir, "env.yaml", sharedEnv+podmanEnv()+"  base-dir: "+filepath.ToSlash(legacyDir)+"\n")
	write(t, dir, "10.yaml", validWF)
	// Pre-seed the unit a deploy would have written, and the files an older
	// deploy left under base-dir. The quadlet dir is under a redirected HOME
	// that does not exist yet, and write() does not create parents -- deploy
	// gets that for free from runner.WriteFile's MkdirAll.
	if err := os.MkdirAll(quadletDir, 0o755); err != nil {
		t.Fatal(err)
	}
	write(t, quadletDir, "solmq-conn.container", "[Container]\nSecret=solmq-conn-tls-old-ca,type=mount,target=/app/external/classpath/truststores/old-ca\n")
	write(t, legacyDir, "solmq-conn-application.yml", "x\n")
	write(t, legacyDir, "solmq-conn-status", "#!/bin/sh\n")

	if code := dispatch([]string{"remove", "--platform", "podman", "-e", filepath.Join(dir, "env.yaml")}, f); code != 0 {
		t.Fatalf("exit=%d", code)
	}
	// Secrets are removed from podman's store only AFTER the unit is stopped and
	// the generator reloaded, mirroring deploy's create-before-start ordering in
	// reverse: a failure removing them still surfaces (see main.go's
	// podmanRemove), but the units referencing them are gone first. Every file
	// secret the instance can own goes too, mounted by this spec or not -- the
	// stores read from the deployed unit's Secret= lines. Call 0
	// is the read-only preflight probe, which runs before any of that.
	wantCalls := [][]string{
		{"podman", "info"},
		{"systemctl", "--user", "stop", "solmq-conn.service"},
		{"systemctl", "--user", "daemon-reload"},
		{"podman", "secret", "rm", "--ignore",
			"solmq-conn-_GEN_MQ_CONN_1_USER", "solmq-conn-_GEN_MQ_CONN_1_PASSWORD",
			"solmq-conn-_GEN_SOL_CONN_1_CLIENT_USERNAME", "solmq-conn-_GEN_SOL_CONN_1_CLIENT_PASSWORD",
			"solmq-conn-application.yml", "solmq-conn-status-script", "solmq-conn-logback-spring.xml",
			"solmq-conn-tls-old-ca"},
	}
	if len(f.calls) != len(wantCalls) {
		t.Fatalf("want %d runner calls, got %+v", len(wantCalls), f.calls)
	}
	for i, w := range wantCalls {
		if !reflect.DeepEqual(f.calls[i].argv, w) {
			t.Errorf("call %d argv = %v, want %v", i, f.calls[i].argv, w)
		}
	}
	if _, err := os.Stat(filepath.Join(quadletDir, "solmq-conn.container")); !os.IsNotExist(err) {
		t.Errorf("the unit should be removed by the remove verb")
	}
	// What an older deploy left under base-dir is not remove's: validate names
	// those files for the operator to delete.
	for _, name := range []string{"solmq-conn-application.yml", "solmq-conn-status"} {
		if _, err := os.Stat(filepath.Join(legacyDir, name)); err != nil {
			t.Errorf("%s under the old base-dir must be left alone: %v", name, err)
		}
	}
}

// ---- j) --allow-command flag --------------------------------------------------

// TestAllowCommandFlagBadValueExitsUsageError covers the flag's own input
// validation, done once per run by execguard.Approve (a bare name, no unsafe
// character, never a privilege-escalation wrapper): a bad value is a usage error
// (exit 2) with the shared message before anything reaches loadEnv/validate/the
// runner.
func TestAllowCommandFlagBadValueExitsUsageError(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "env.yaml", sharedEnv+kubeEnv)
	write(t, dir, "10.yaml", validWF)
	envPath := filepath.Join(dir, "env.yaml")

	cases := []struct {
		name string
		val  string
		want string
	}{
		{"path", "/usr/bin/lima", "it must be a bare binary name, not a path"},
		{"unsafe character", "lima;rm", `which is not allowed in a command token`},
		{"escalator", "sudo", "--allow-command sudo is never permitted"},
		{"escalator in another case with .exe", "Doas.EXE", "elevate the tool instead"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			f := &fakeRunner{}
			var code int
			stderr := captureStderr(t, func() {
				code = dispatch([]string{"deploy", "--platform", "kubernetes", "-e", envPath, "--allow-command", c.val}, f)
			})
			if code != 2 {
				t.Errorf("--allow-command %q: exit=%d, want 2", c.val, code)
			}
			if !strings.Contains(stderr, c.want) {
				t.Errorf("--allow-command %q: stderr should contain %q, got:\n%s", c.val, c.want, stderr)
			}
			if len(f.calls) != 0 {
				t.Errorf("--allow-command %q: runner must not be invoked, got %d calls", c.val, len(f.calls))
			}
		})
	}
}

// TestAllowCommandSudoNamesTheSupportedAlternative pins the message an operator
// meets on `--allow-command sudo`: it names `sudo solmq-conn-util`, the supported
// way to run elevated, on every verb that takes the flag.
func TestAllowCommandSudoNamesTheSupportedAlternative(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "env.yaml", sharedEnv+kubeEnv)
	write(t, dir, "10.yaml", validWF)
	envPath := filepath.Join(dir, "env.yaml")
	for _, verb := range []string{"deploy", "remove", "start", "status", "logs", "cli"} {
		t.Run(verb, func(t *testing.T) {
			f := &fakeRunner{}
			var code int
			stderr := captureStderr(t, func() {
				code = dispatch([]string{verb, "--platform", "kubernetes", "-e", envPath, "--allow-command", "sudo"}, f)
			})
			if code != 2 {
				t.Errorf("%s --allow-command sudo: exit=%d, want 2", verb, code)
			}
			if !strings.Contains(stderr, "sudo solmq-conn-util") {
				t.Errorf("%s: stderr should name sudo solmq-conn-util, got:\n%s", verb, stderr)
			}
			if len(f.calls) != 0 {
				t.Errorf("%s: runner must not be invoked, got %d calls", verb, len(f.calls))
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
		{"generate", "config", "-e", envPath, "--allow-command", "lima"},
		{"validate", "-e", envPath, "--allow-command", "lima"},
	}
	for _, args := range cases {
		if code := run(args); code != 2 {
			t.Errorf("%v: exit=%d, want 2 (flag not defined on this verb)", args, code)
		}
	}
}

// TestAllowCommandFlagRepeatableThreadsToRunner is the threading assertion:
// a chained `lima podman` command: fails before the runner without the flag,
// and reaches every runner call (preflight, secrets) once --allow-command lima
// approves it. Passing the flag twice with the same value also covers that it
// is genuinely repeatable (flag.Value.Set called more than once).
func TestAllowCommandFlagRepeatableThreadsToRunner(t *testing.T) {
	// The approved run completes a deploy, so the unit directory has to be
	// redirected away from the developer's real ~/.config.
	podmanQuadletHome(t)
	dir := t.TempDir()
	write(t, dir, "env.yaml", sharedEnv+podmanEnvChained())
	write(t, dir, "10.yaml", validWF)
	envPath := filepath.Join(dir, "env.yaml")

	fReject := useFakeRunner(t)
	if code := dispatch([]string{"deploy", "--platform", "podman", "-e", envPath}, fReject); code != 2 {
		t.Fatalf("without --allow-command: exit=%d, want 2", code)
	}
	if len(fReject.calls) != 0 {
		t.Fatalf("without --allow-command: runner must not be invoked, got %d calls", len(fReject.calls))
	}

	fAccept := useFakeRunner(t)
	args := []string{"deploy", "--platform", "podman", "-e", envPath, "--allow-command", "lima", "--allow-command", "lima"}
	if code := dispatch(args, fAccept); code != 0 {
		t.Fatalf("with --allow-command lima: exit=%d, want 0", code)
	}
	if len(fAccept.calls) == 0 {
		t.Fatalf("with --allow-command lima: expected the runner to be reached")
	}
	// The preflight probe runs first, resolving argv[0] "lima" via the flag and
	// "podman" as its own allowlisted chained token.
	if want := []string{"lima", "podman", "info"}; !reflect.DeepEqual(fAccept.calls[0].argv, want) {
		t.Errorf("preflight argv = %v, want %v", fAccept.calls[0].argv, want)
	}
	foundSecretCall := false
	for _, c := range fAccept.calls {
		if len(c.argv) >= 3 && c.argv[0] == "lima" && c.argv[1] == "podman" && c.argv[2] == "secret" {
			foundSecretCall = true
			break
		}
	}
	if !foundSecretCall {
		t.Errorf("expected a podman secret call prefixed with [lima podman secret ...], got %+v", fAccept.calls)
	}
}

// TestDeployRefusesSudoCommandWhateverIsApproved pins the escalator ban: an env
// file whose podman.command is `sudo podman` is refused at deploy, exit 2 with
// the runner never reached, with no approval and with an unrelated one. The only
// supported way to run elevated is `sudo solmq-conn-util`.
func TestDeployRefusesSudoCommandWhateverIsApproved(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "env.yaml", sharedEnv+podmanEnvSudo())
	write(t, dir, "10.yaml", validWF)
	envPath := filepath.Join(dir, "env.yaml")
	for _, extra := range [][]string{nil, {"--allow-command", "lima"}} {
		f := useFakeRunner(t)
		args := append([]string{"deploy", "--platform", "podman", "-e", envPath}, extra...)
		var code int
		stderr := captureStderr(t, func() { code = dispatch(args, f) })
		if code != 2 {
			t.Errorf("%v: exit=%d, want 2", args, code)
		}
		if !strings.Contains(stderr, `podman.command[0] = "sudo" is a privilege-escalation wrapper (sudo), which is never permitted`) ||
			!strings.Contains(stderr, "run the tool itself under sudo instead") {
			t.Errorf("%v: stderr should refuse the wrapper and name the alternative, got:\n%s", args, stderr)
		}
		if len(f.calls) != 0 {
			t.Errorf("%v: runner must not be invoked, got %d calls", args, len(f.calls))
		}
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
	want := []string{"kubectl", "auth", "can-i", "create", "deployments", "-n", "solace-connectors"}
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
	quadletDir := podmanQuadletHome(t)
	dir := t.TempDir()
	write(t, dir, "env.yaml", sharedEnv+podmanEnv())
	write(t, dir, "10.yaml", validWF)

	if code := dispatch([]string{"deploy", "--platform", "podman", "-e", filepath.Join(dir, "env.yaml")}, f); code != 1 {
		t.Fatalf("exit=%d, want 1", code)
	}
	// The unit is the one file deploy writes, and preflight runs before it.
	if _, err := os.Stat(filepath.Join(quadletDir, "solmq-conn.container")); !os.IsNotExist(err) {
		t.Errorf("the unit must not be written when preflight fails")
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
		t.Run(c.alias, func(t *testing.T) {
			if c.wantBinary == "podman" {
				// A podman deploy writes its unit to the invoking user's quadlet
				// directory, so it is redirected like every other podman deploy
				// test -- or this one would install a fixture unit for real.
				podmanQuadletHome(t)
			}
			f := useFakeRunner(t)
			dir := t.TempDir()
			write(t, dir, "env.yaml", sharedEnv+kubeEnv+dockerEnv+podmanEnv())
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
		})
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
		if code != 2 {
			t.Fatalf("exit=%d, want 2", code)
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
		if code != 2 {
			t.Fatalf("exit=%d, want 2", code)
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
	if got := platformAbbrev.Shorts(); len(got) != len(platformAliasList) {
		t.Errorf("platform set has %d short forms for %d declared aliases", len(got), len(platformAliasList))
	}
}

// TestAliasDeclarationsFollowTheSharedRules builds the real sets -- the verbs,
// each verb's targets and the platforms -- through the same builders the binary
// uses. abbrev.New panics on a declaration that breaks the shared rule (lower-case
// letters and digits, shorter than the word, unique, never a canonical name), so
// building them here is the check; the loop then pins that every declared alias
// expands to its canonical word.
func TestAliasDeclarationsFollowTheSharedRules(t *testing.T) {
	commands := newCommandAbbrev()
	for _, v := range cliVerbs {
		for _, a := range v.Aliases {
			if got, ok := commands.Expand(a); !ok || got != v.Name {
				t.Errorf("verb alias %q expands to %q (known=%v), want %q", a, got, ok, v.Name)
			}
		}
	}
	sets := newTargetAbbrev()
	for _, v := range cliVerbs {
		set, ok := sets[v.Name]
		if len(v.Targets) == 0 {
			if ok {
				t.Errorf("%s models no targets but has a target set", v.Name)
			}
			continue
		}
		if !ok {
			t.Errorf("%s models targets but has no target set", v.Name)
			continue
		}
		for _, tg := range v.Targets {
			for _, a := range tg.Aliases {
				if got, ok := set.Expand(a); !ok || got != tg.Name {
					t.Errorf("%s target alias %q expands to %q (known=%v), want %q", v.Name, a, got, ok, tg.Name)
				}
			}
		}
	}
	platforms := newPlatformAbbrev()
	for _, e := range platformAliasList {
		if got, ok := platforms.Expand(e.Alias); !ok || got != e.Canonical {
			t.Errorf("platform alias %q expands to %q (known=%v), want %q", e.Alias, got, ok, e.Canonical)
		}
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
		if code != 2 {
			t.Fatalf("exit=%d, want 2", code)
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
	if len(f.calls) != kubeProbes+3 {
		t.Errorf("want %d runner calls (probes, namespace read, apply, rollout wait), got %d: %+v", kubeProbes+3, len(f.calls), f.calls)
	}
}

// TestPlatformMenuOnMultipleSections asserts that with no --platform and more
// than one section present, the interactive menu (via the injected prompter
// seam) picks the platform.
func TestPlatformMenuOnMultipleSections(t *testing.T) {
	withPrompter(t, true, "2\n") // present order is [kubernetes, docker]; 2 = docker
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
// never return, and that the resulting error points at --platform. The
// prompter is told it has no terminal rather than relying on whatever terminal
// state the test binary happened to inherit, so this cannot flake or block
// regardless of how the tests are invoked.
func TestPlatformMenuNonTTYRefusesWithPlatformHint(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "env.yaml", sharedEnv+kubeEnv+dockerEnv)

	withPrompter(t, false, "")

	f := &fakeRunner{}
	stderr := captureStderr(t, func() {
		code := dispatch([]string{"deploy", "-e", filepath.Join(dir, "env.yaml")}, f)
		if code != 2 {
			t.Fatalf("exit=%d, want 2", code)
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
		if code != 2 {
			t.Fatalf("exit=%d, want 2", code)
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
	q := &queueRunner{resp: []queuedResp{{podsJSON, nil}}}
	var code int
	captureStdout(t, func() {
		code = dispatch([]string{"sts", "cnt", "--platform", "kubernetes", "--pod", "pod-a", "--pod", "pod-b"}, q)
	})
	if code != 0 {
		t.Fatalf("exit=%d, want 0", code)
	}
	if len(q.calls) != 1 {
		t.Errorf("the container view should cost one get, got %d calls: %+v", len(q.calls), q.calls)
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
		{"raw on the container view", []string{"status", "container", "--raw"}, "--raw applies to the application"},
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
	if err := w.Set("true"); err != nil || !w.on || w.interval != time.Duration(watchDefaultSeconds)*time.Second {
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

// ---- --watch -------------------------------------------------------------------

// watchTestCollector is a docker container-view collector over one scripted
// runner, so a watch refresh costs one inspect call and prints one table.
func watchTestCollector(r gdurunner.Runner, view statusreport.View) (*statusCollector, statusOpts) {
	o := statusOpts{view: view, conts: []string{"solmq-connector"}, watch: watchFlag{on: true, interval: 5 * time.Second}}
	return &statusCollector{opts: o, platform: validate.PlatformDocker, cmdArgv: []string{"docker"}, r: r}, o
}

// inspectRunner answers every call with the same inspect document, or fails
// every call with err.
type inspectRunner struct {
	noAttach
	err   error
	calls int

	// onRun, when set, runs at the start of every call, so a test can look at
	// what the loop has written by the time a collection begins, or interrupt it.
	onRun func()
}

func (i *inspectRunner) Run(_ context.Context, _ gdurunner.Cmd, stdout, _ io.Writer) error {
	i.calls++
	if i.onRun != nil {
		i.onRun()
	}
	if i.err != nil {
		return i.err
	}
	scriptOut(stdout, dockerInspect)
	return nil
}

// sleepAfter returns a sleep seam that records each interval it is asked for and
// cancels the context once it has been called n times, which is what ends the
// loop in a test.
func sleepAfter(n int, cancel context.CancelFunc, got *[]time.Duration) func(context.Context, time.Duration) error {
	return func(ctx context.Context, d time.Duration) error {
		*got = append(*got, d)
		if len(*got) >= n {
			cancel()
			return ctx.Err()
		}
		return nil
	}
}

// TestWatchStatusRedrawsUnderTheHeader covers a terminal: each refresh starts
// with the shared clear-screen sequence and the header naming the interval, the
// view and the time, and the report follows it.
func TestWatchStatusRedrawsUnderTheHeader(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var slept []time.Duration
	c, o := watchTestCollector(&inspectRunner{}, statusreport.ViewContainer)
	o.view = statusreport.ViewAll // the title only; the collector stays on the engine view
	var out bytes.Buffer
	code := watchStatus(ctx, o, c, watchIO{redraw: true, out: &out, sleep: sleepAfter(1, cancel, &slept)})
	if code != 0 {
		t.Fatalf("exit=%d, want 0", code)
	}
	got := out.String()
	if !strings.HasPrefix(got, gduwatch.ClearScreen+"Every 5s: solmq-conn-util status all   ") {
		t.Errorf("a redraw should open with the clear and the header, got:\n%q", got)
	}
	if !strings.Contains(got, "   Ctrl-C to stop\n\n") || !strings.Contains(got, "solmq-connector") {
		t.Errorf("the header should end the line and the report follow, got:\n%q", got)
	}
}

// TestWatchStatusAppendsSectionsWhenNotATerminal covers a pipe or a file: each
// refresh is appended under a timestamped section line and no escape byte is
// written.
func TestWatchStatusAppendsSectionsWhenNotATerminal(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var slept []time.Duration
	c, o := watchTestCollector(&inspectRunner{}, statusreport.ViewContainer)
	var out bytes.Buffer
	code := watchStatus(ctx, o, c, watchIO{out: &out, sleep: sleepAfter(2, cancel, &slept)})
	if code != 0 {
		t.Fatalf("exit=%d, want 0", code)
	}
	got := out.String()
	if strings.Contains(got, "\x1b") {
		t.Errorf("no escape byte belongs in a pipe, got:\n%q", got)
	}
	if n := strings.Count("\n"+got, "\n=== "); n != 2 {
		t.Errorf("want one section line per refresh (2), got %d in:\n%s", n, got)
	}
	if strings.Contains(got, "Every ") {
		t.Errorf("the redraw header is for a terminal only, got:\n%s", got)
	}
}

// TestWatchStatusSleepsTheIntervalAfterEachRefresh pins that the pause is the
// flag's interval and is taken once per refresh, after the report.
func TestWatchStatusSleepsTheIntervalAfterEachRefresh(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var slept []time.Duration
	r := &inspectRunner{}
	c, o := watchTestCollector(r, statusreport.ViewContainer)
	o.watch.interval = 7 * time.Second
	watchStatus(ctx, o, c, watchIO{out: io.Discard, sleep: sleepAfter(3, cancel, &slept)})
	if !reflect.DeepEqual(slept, []time.Duration{7 * time.Second, 7 * time.Second, 7 * time.Second}) {
		t.Errorf("slept %v, want 7s after each of 3 refreshes", slept)
	}
	if r.calls != 3 {
		t.Errorf("want one collection per refresh (3), got %d", r.calls)
	}
}

// TestWatchStatusReportsAFailedCollectionOnceAndCarriesOn covers a refresh the
// engine refuses: the collection error goes back to the shared loop, which warns
// once per refresh on stderr naming the retry, nothing else reports it, and the
// next refresh still runs.
func TestWatchStatusReportsAFailedCollectionOnceAndCarriesOn(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var slept []time.Duration
	r := &inspectRunner{err: errors.New("engine down")}
	c, o := watchTestCollector(r, statusreport.ViewContainer)
	var code int
	stderr := captureStderr(t, func() {
		code = watchStatus(ctx, o, c, watchIO{out: io.Discard, sleep: sleepAfter(2, cancel, &slept)})
	})
	if code != 0 {
		t.Errorf("exit=%d, want 0: a watch ended on purpose did not fail", code)
	}
	if n := strings.Count(stderr, "engine down"); n != 2 {
		t.Errorf("want the failure once per refresh (2), got %d in:\n%s", n, stderr)
	}
	if n := strings.Count(stderr, "warning: inspecting solmq-connector: engine down -- trying again in 5s\n"); n != 2 {
		t.Errorf("want the loop's warning naming the retry once per refresh (2), each on one line, got %d in:\n%s", n, stderr)
	}
	if strings.Contains(stderr, "error:") {
		t.Errorf("a failed refresh is a warning, not an error, got:\n%s", stderr)
	}
	if r.calls != 2 {
		t.Errorf("the loop should carry on after a failure, got %d collections", r.calls)
	}
}

// TestWatchStatusCancelledContextExitsZero covers Ctrl-C arriving while the
// loop is paused: the real sleep returns the interruption and the exit is 0.
func TestWatchStatusCancelledContextExitsZero(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	c, o := watchTestCollector(&inspectRunner{}, statusreport.ViewContainer)
	if code := watchStatus(ctx, o, c, watchIO{out: io.Discard}); code != 0 {
		t.Errorf("exit=%d, want 0", code)
	}
}

// TestWatchStatusKeepsTheLastReportUntilTheNextIsCollected covers the redraw
// order: a refresh collects before it clears, so while the next collection runs
// the previous report is still the whole screen.
func TestWatchStatusKeepsTheLastReportUntilTheNextIsCollected(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var slept []time.Duration
	var out bytes.Buffer
	var onScreen []string
	r := &inspectRunner{}
	r.onRun = func() { onScreen = append(onScreen, out.String()) }
	c, o := watchTestCollector(r, statusreport.ViewContainer)
	if code := watchStatus(ctx, o, c, watchIO{redraw: true, out: &out, sleep: sleepAfter(2, cancel, &slept)}); code != 0 {
		t.Fatalf("exit=%d, want 0", code)
	}
	if len(onScreen) != 2 {
		t.Fatalf("want 2 collections, got %d", len(onScreen))
	}
	if onScreen[0] != "" {
		t.Errorf("nothing may be drawn before the first report is collected, got %q", onScreen[0])
	}
	if n := strings.Count(onScreen[1], gduwatch.ClearScreen); n != 1 || !strings.Contains(onScreen[1], "solmq-connector") {
		t.Errorf("the second collection should run under the first report, got:\n%q", onScreen[1])
	}
	// Each frame is still the clear, the header, then the report.
	frames := strings.Split(out.String(), gduwatch.ClearScreen)
	if len(frames) != 3 || frames[0] != "" {
		t.Fatalf("want two frames, each opening with the clear, got:\n%q", out.String())
	}
	for _, f := range frames[1:] {
		if !strings.HasPrefix(f, "Every 5s: solmq-conn-util status container   ") || !strings.Contains(f, "solmq-connector") {
			t.Errorf("a frame should be the header and then the report, got:\n%q", f)
		}
	}
}

// TestWatchStatusDrawsNothingForARefreshCutShort covers Ctrl-C arriving during a
// collection: whatever came back was cut short, so nothing is drawn, not even
// the header, and the watch still exits 0.
func TestWatchStatusDrawsNothingForARefreshCutShort(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	r := &inspectRunner{onRun: cancel}
	c, o := watchTestCollector(r, statusreport.ViewContainer)
	var out bytes.Buffer
	if code := watchStatus(ctx, o, c, watchIO{redraw: true, out: &out}); code != 0 {
		t.Errorf("exit=%d, want 0", code)
	}
	if out.Len() != 0 {
		t.Errorf("a refresh cut short must draw nothing, got:\n%q", out.String())
	}
	if r.calls != 1 {
		t.Errorf("want the one interrupted collection, got %d", r.calls)
	}
}

// TestWatchStatusRejectsAnIntervalTheLoopCannotUse covers options the shared
// loop refuses before any refresh: the error goes through errExit and nothing
// is collected.
func TestWatchStatusRejectsAnIntervalTheLoopCannotUse(t *testing.T) {
	r := &inspectRunner{}
	c, o := watchTestCollector(r, statusreport.ViewContainer)
	o.watch.interval = 0
	var code int
	stderr := captureStderr(t, func() {
		code = watchStatus(context.Background(), o, c, watchIO{out: io.Discard})
	})
	if code == 0 || !strings.Contains(stderr, "error:") || r.calls != 0 {
		t.Errorf("exit=%d, collections=%d, stderr:\n%s", code, r.calls, stderr)
	}
}

// TestWatchTitleNamesTheViewAndItsFlags pins what follows "Every <interval>: ".
func TestWatchTitleNamesTheViewAndItsFlags(t *testing.T) {
	cases := []struct {
		name string
		opts statusOpts
		want string
	}{
		{"container", statusOpts{view: statusreport.ViewContainer}, "solmq-conn-util status container"},
		{"application", statusOpts{view: statusreport.ViewApplication}, "solmq-conn-util status application"},
		{"all", statusOpts{view: statusreport.ViewAll}, "solmq-conn-util status all"},
		{"details", statusOpts{view: statusreport.ViewAll, level: statusreport.LevelDetails}, "solmq-conn-util status all --details"},
		{"all instances", statusOpts{view: statusreport.ViewContainer, all: true}, "solmq-conn-util status container --all"},
		{"both", statusOpts{view: statusreport.ViewApplication, level: statusreport.LevelDetails, all: true}, "solmq-conn-util status application --details --all"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := watchTitle(c.opts); got != c.want {
				t.Errorf("title = %q, want %q", got, c.want)
			}
		})
	}
}

// ---- the container view ---------------------------------------------------------

// TestStatusContainerViewReadsEngineFactsWithoutExecing is the shape of the new
// view: one read-only query answers discovery and the whole table, and nothing
// is ever exec'd into -- the in-container script is not involved at all.
func TestStatusContainerViewReadsEngineFactsWithoutExecing(t *testing.T) {
	q := &queueRunner{resp: []queuedResp{
		{podsJSON, nil}, // get pods -o json: discovery and facts together; no permission probe first
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
	if len(q.calls) != 1 {
		t.Fatalf("want 1 call (get pods), got %d: %+v", len(q.calls), q.calls)
	}
	// Nothing was exec'd: no probe, no install, no script run.
	for _, c := range q.calls {
		for _, tok := range c.argv {
			if tok == "exec" {
				t.Errorf("the container view must not exec into anything: %v", c.argv)
			}
		}
	}
	if got := q.calls[0].argv; got[len(got)-1] != "json" {
		t.Errorf("discovery should ask for json, got %v", got)
	}
}

// TestStatusContainerDetailsSamplesAndChecksComponents covers what --details
// adds on kubernetes: one sampling call for the whole run, and one presence
// check per distinct referenced object -- deduplicated, since every pod of a
// deployment references the same ones.
func TestStatusContainerDetailsSamplesAndChecksComponents(t *testing.T) {
	q := &queueRunner{resp: []queuedResp{
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
	if len(q.calls) != 3 {
		t.Fatalf("want 3 calls (get pods, top, one component check), got %d: %+v", len(q.calls), q.calls)
	}
	if !containsToken(q.calls[1].argv, "--containers") {
		t.Errorf("the sample must be per container, got %v", q.calls[1].argv)
	}
}

// TestStatusContainerDetailsWithoutMetricsServerDegradesToANote covers the
// cluster with no metrics API: an optional add-on being absent must cost the
// resource lines and nothing else -- the report still prints, and the exit code
// is still about the instances, not about the collection.
func TestStatusContainerDetailsWithoutMetricsServerDegradesToANote(t *testing.T) {
	q := &queueRunner{resp: []queuedResp{
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
	podmanInspect := `[{"Name":"solmq-connector","State":{"Status":"running","StartedAt":"2026-08-18T04:12:07Z",
	  "Health":{"Status":"healthy","FailingStreak":0,"Log":[{"ExitCode":0,"Output":"UP\n"}]}},
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
	// The quadlet this tool generates declares a HealthCmd, so a running
	// instance reports a real verdict rather than the n/a the column showed
	// while no healthcheck was emitted at all.
	if !strings.Contains(stdout, "healthy") {
		t.Errorf("the podman healthcheck verdict should reach the HEALTH column, got:\n%s", stdout)
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
	// This fixture declares no healthcheck, which is what a container created
	// before the quadlet carried HealthCmd looks like: the column has to fall
	// back rather than show an empty cell.
	if !strings.Contains(stdout, statusreport.NotApplicable) {
		t.Errorf("a container with no healthcheck should read %q, got:\n%s", statusreport.NotApplicable, stdout)
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
		q := &queueRunner{resp: []queuedResp{{pods, nil}}}
		var code int
		stdout := captureStdout(t, func() {
			code = dispatch([]string{"status", "container", "--platform", "kubernetes", "--all"}, q)
		})
		if code != 0 {
			t.Fatalf("exit=%d, want 0", code)
		}
		if !containsToken(q.calls[0].argv, "--all-namespaces") {
			t.Errorf("--all must search the whole cluster, got %v", q.calls[0].argv)
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
		list := "solmq-connector\tsolace/solace-pubsub-connector-ibmmq:2.14.1\tUp 3 days\nredis\tredis:7\tUp 3 days\n"
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
	q := &queueRunner{resp: []queuedResp{{`{"items":[]}`, nil}}}
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
		{onePodJSON, nil}, // get pods
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

	// report returns what was printed and the argv the script was run with.
	report := func(args ...string) (string, []string) {
		q := &queueRunner{resp: []queuedResp{{onePodJSON, nil}, {full, nil}}}
		out := captureStdout(t, func() {
			if code := dispatch(args, q); code != 0 {
				t.Fatalf("%v: exit=%d, want 0", args, code)
			}
		})
		return out, q.calls[len(q.calls)-1].argv
	}

	details, argv := report("status", "app", "-d", "--platform", "kubernetes", "--pod", "pod-a")
	// The details level asks the script for what it renders, so the script
	// spends its extra actuator calls and the JVM start only here. Both levels
	// ask for the raw documents the connection verdict is read from.
	if !slices.Contains(argv, statusscript.DetailsArg) || !slices.Contains(argv, statusscript.RawArg) {
		t.Errorf("-d should run the script with %s and %s, got argv %v", statusscript.DetailsArg, statusscript.RawArg, argv)
	}
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

	// The basic level prints none of it, even from a script that still reports
	// it: an instance deployed by an older release ignores the argument.
	basic, argv := report("status", "app", "--platform", "kubernetes", "--pod", "pod-a")
	if slices.Contains(argv, statusscript.DetailsArg) || !slices.Contains(argv, statusscript.RawArg) {
		t.Errorf("the basic level must ask for %s and not for the details calls, got argv %v", statusscript.RawArg, argv)
	}
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
	withScriptWorkers(t, 1) // the answers below are in report order
	q := &queueRunner{resp: []queuedResp{
		{podsJSON, nil},  // get pods: pod-a fine, pod-b crash looping
		{appReport, nil}, // pod-a run
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
			{onePodJSON, nil},                       // get pods
			{runner.ScriptAbsentMarker + "\n", nil}, // run: the script is absent
		}, extra...)}
	}

	t.Run("--install installs without asking", func(t *testing.T) {
		q := base(queuedResp{"", nil}, queuedResp{appReport, nil})
		captureStdout(t, func() {
			if code := dispatch([]string{"status", "app", "--install", "--platform", "kubernetes", "--pod", "pod-a"}, q); code != 0 {
				t.Fatalf("exit=%d, want 0", code)
			}
		})
		if len(q.calls) != 4 {
			t.Fatalf("want 4 calls (get, run finding it absent, install, run), got %d: %+v", len(q.calls), q.calls)
		}
	})

	t.Run("prompt answered yes installs", func(t *testing.T) {
		withPrompter(t, true, "y\n")
		q := base(queuedResp{"", nil}, queuedResp{appReport, nil})
		captureStdout(t, func() {
			if code := dispatch([]string{"status", "app", "--platform", "kubernetes", "--pod", "pod-a"}, q); code != 0 {
				t.Fatalf("exit=%d, want 0", code)
			}
		})
		if len(q.calls) != 4 {
			t.Fatalf("want 4 calls, got %d: %+v", len(q.calls), q.calls)
		}
	})

	t.Run("prompt declined skips the instance and exits 1", func(t *testing.T) {
		withPrompter(t, true, "n\n")
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
		if len(q.calls) != 2 {
			t.Fatalf("nothing should be installed or run, got %d calls: %+v", len(q.calls), q.calls)
		}
	})
}

// TestStatusInstallPromptNonTTYRefusesWithInstallHint is the confirmInstall
// counterpart to TestPlatformMenuNonTTYRefusesWithPlatformHint: the install
// confirmation shares the same prompter, so it must refuse rather than block
// on a read that would never return, and point at --install.
func TestStatusInstallPromptNonTTYRefusesWithInstallHint(t *testing.T) {
	withPrompter(t, false, "")

	q := &queueRunner{resp: []queuedResp{
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
	if len(q.calls) != 2 {
		t.Fatalf("nothing installed or run, got %d calls: %+v", len(q.calls), q.calls)
	}
}

// TestStatusScriptStillMissingAfterInstallFails covers an install that reported
// success while the run after it still found nothing at the path: the instance
// says so in its block and the run exits 1, rather than being offered the same
// install again.
func TestStatusScriptStillMissingAfterInstallFails(t *testing.T) {
	q := &queueRunner{resp: []queuedResp{
		{onePodJSON, nil},
		{runner.ScriptAbsentMarker + "\n", nil}, // run: absent
		{"", nil},                               // install
		{runner.ScriptAbsentMarker + "\n", nil}, // run: still absent
	}}
	var code int
	stdout := captureStdout(t, func() {
		code = dispatch([]string{"status", "app", "--install", "--platform", "kubernetes", "--pod", "pod-a"}, q)
	})
	if code != 1 {
		t.Fatalf("exit=%d, want 1", code)
	}
	if !strings.Contains(stdout, "status: the status script is still missing after the install") {
		t.Errorf("the instance should say the install did not take, got:\n%s", stdout)
	}
	if len(q.calls) != 4 {
		t.Fatalf("want 4 calls (get, run, install, run) and no second install, got %d: %+v", len(q.calls), q.calls)
	}
}

// TestStatusStandbyIsAnAnswerNotAFailure keeps the contract the whole verb is
// built on: the script always exits 0 and carries the state in its output, so a
// standby instance prints like any other and the run still exits 0.
func TestStatusStandbyIsAnAnswerNotAFailure(t *testing.T) {
	withScriptWorkers(t, 1) // the answers below are in report order
	q := &queueRunner{resp: []queuedResp{
		{podsJSON, nil},
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

// reconnectingReport is the script's --raw report from an instance whose MQ
// binder is reconnecting, with workflow 0 riding on it in the given state.
func reconnectingReport(leState, wfState string) string {
	doc := `{"status":"RECONNECTING","components":{"binders":{"status":"RECONNECTING","components":{` +
		`"mq-conn-1":{"status":"RECONNECTING","details":{"error":"JMSWMQ0018: Failed to connect to queue manager 'QM1'"}},` +
		`"prod-solace":{"status":"UP"}}}}}`
	return "leader-election mode: active_standby\nleader-election state: " + leState + "\n" +
		"health: RECONNECTING\nhealth-detail: " + doc + "\nhealth-json: " + doc + "\n" +
		"binding: input-0 mq-conn-1\nbinding: output-0 prod-solace\n" +
		"workflows:\n  0: " + wfState + "\n"
}

// TestStatusRunningWorkflowNotConnectedFailsTheRun covers the one answer the
// report treats as a failure: a workflow that reads running while a connection
// it rides on is not UP. The row says which connection, the basic level shows
// the failing binder with its error, and the run exits 1. A standby beside the
// same binder is an answer, since its workflows are stopped.
func TestStatusRunningWorkflowNotConnectedFailsTheRun(t *testing.T) {
	run := func(t *testing.T, leState, wfState string) (string, int) {
		t.Helper()
		q := &queueRunner{resp: []queuedResp{{onePodJSON, nil}, {reconnectingReport(leState, wfState), nil}}}
		var code int
		stdout := captureStdout(t, func() {
			code = dispatch([]string{"status", "app", "--platform", "kubernetes", "--pod", "pod-a"}, q)
		})
		return stdout, code
	}

	t.Run("active", func(t *testing.T) {
		stdout, code := run(t, "active", "running")
		if code != 1 {
			t.Errorf("exit=%d, want 1 for a running workflow that is not connected", code)
		}
		for _, want := range []string{
			"  connections:\n    mq-conn-1: RECONNECTING  JMSWMQ0018: Failed to connect to queue manager 'QM1'\n",
			"    0: running (mq-conn-1: RECONNECTING)\n",
		} {
			if !strings.Contains(stdout, want) {
				t.Errorf("stdout missing %q, got:\n%s", want, stdout)
			}
		}
	})

	t.Run("stopped workflow on a standby", func(t *testing.T) {
		stdout, code := run(t, "standby", "stopped")
		if code != 0 {
			t.Errorf("exit=%d, want 0: a stopped workflow holds no connection to judge", code)
		}
		if !strings.Contains(stdout, "    0: stopped\n") || strings.Contains(stdout, "(mq-conn-1") {
			t.Errorf("a stopped workflow must not be marked, got:\n%s", stdout)
		}
	})
}

// TestStatusRawPrintsTheHealthDocumentInFull covers status --raw: the health
// document the instance answered is printed whole under its block, and is the
// healthDocument object of the JSON output. Without --raw neither appears, and
// an instance whose script sent no document says so instead of printing
// nothing.
func TestStatusRawPrintsTheHealthDocumentInFull(t *testing.T) {
	run := func(t *testing.T, report string, args ...string) (string, int) {
		t.Helper()
		q := &queueRunner{resp: []queuedResp{{onePodJSON, nil}, {report, nil}}}
		var code int
		stdout := captureStdout(t, func() {
			code = dispatch(append([]string{"status", "app", "--platform", "kubernetes", "--pod", "pod-a"}, args...), q)
		})
		return stdout, code
	}

	t.Run("table", func(t *testing.T) {
		stdout, _ := run(t, reconnectingReport("active", "running"), "--raw")
		for _, want := range []string{
			"  health document:\n    {\n      \"status\": \"RECONNECTING\",\n",
			"\"error\": \"JMSWMQ0018: Failed to connect to queue manager 'QM1'\"",
		} {
			if !strings.Contains(stdout, want) {
				t.Errorf("stdout missing %q, got:\n%s", want, stdout)
			}
		}
		// The document in full makes the one-line copy redundant.
		if strings.Contains(stdout, "health-detail:") {
			t.Errorf("health-detail must give way to the full document, got:\n%s", stdout)
		}
	})

	t.Run("json", func(t *testing.T) {
		stdout, _ := run(t, reconnectingReport("active", "running"), "--raw", "--output", "json")
		var doc struct {
			Instances []struct {
				Application struct {
					HealthDocument map[string]any `json:"healthDocument"`
				} `json:"application"`
			} `json:"instances"`
		}
		if err := json.Unmarshal([]byte(stdout), &doc); err != nil {
			t.Fatalf("stdout must be one json document: %v\n%s", err, stdout)
		}
		if len(doc.Instances) != 1 || doc.Instances[0].Application.HealthDocument["status"] != "RECONNECTING" {
			t.Errorf("healthDocument should be the instance's document as an object, got:\n%s", stdout)
		}
	})

	t.Run("without raw", func(t *testing.T) {
		stdout, _ := run(t, reconnectingReport("active", "running"), "--output", "json")
		if strings.Contains(stdout, "healthDocument") {
			t.Errorf("the document is only kept under --raw, got:\n%s", stdout)
		}
		table, _ := run(t, reconnectingReport("active", "running"))
		if strings.Contains(table, "health document:") {
			t.Errorf("the document is only printed under --raw, got:\n%s", table)
		}
	})

	t.Run("an older script sends no document", func(t *testing.T) {
		stdout, code := run(t, appReport, "--raw")
		if code != 0 {
			t.Errorf("exit=%d, want 0: a missing document is a note, not a failure", code)
		}
		if !strings.Contains(stdout, "status: no health document from this instance") {
			t.Errorf("the missing document should be said in the block, got:\n%s", stdout)
		}
	})
}

// podListJSON is onePodJSON's running pod under each of names, for a test that
// needs more instances than the fixed fixtures carry.
func podListJSON(names ...string) string {
	item := strings.TrimSuffix(strings.TrimPrefix(onePodJSON, `{"items":[`), `]}`)
	items := make([]string, len(names))
	for i, n := range names {
		items[i] = strings.Replace(item, `"name":"pod-a"`, `"name":"`+n+`"`, 1)
	}
	return `{"items":[` + strings.Join(items, ",") + `]}`
}

// targetRunner answers each exec with a report naming the pod it reached, and
// any other call with its pod list. It is safe to call from several goroutines
// at once and counts how many execs are in flight together. It holds the first
// `together` execs until that many have arrived, which a serial run never
// reaches, and records a timeout instead of hanging when they do not.
type targetRunner struct {
	noAttach
	pods     string
	together int
	met      chan struct{}

	mu                sync.Mutex
	arrived, inFlight int
	maxInFlight       int
	timedOut          bool
}

func newTargetRunner(pods string, together int) *targetRunner {
	return &targetRunner{pods: pods, together: together, met: make(chan struct{})}
}

func (r *targetRunner) Run(_ context.Context, c gdurunner.Cmd, stdout, _ io.Writer) error {
	if len(c.Argv) < 3 || c.Argv[1] != "exec" {
		scriptOut(stdout, r.pods)
		return nil
	}
	pod := c.Argv[2]
	r.mu.Lock()
	r.arrived++
	r.inFlight++
	r.maxInFlight = max(r.maxInFlight, r.inFlight)
	if r.arrived == r.together {
		close(r.met)
	}
	r.mu.Unlock()
	if r.together > 0 {
		select {
		case <-r.met:
		case <-time.After(5 * time.Second):
			r.mu.Lock()
			r.timedOut = true
			r.mu.Unlock()
		}
	}
	r.mu.Lock()
	r.inFlight--
	r.mu.Unlock()
	scriptOut(stdout, "leader-election mode: mode-"+pod+"\nleader-election state: active\n")
	return nil
}

// TestStatusRunsTheScriptsConcurrently covers the batch the application half
// runs in: the execs overlap, never more of them than the worker limit, and
// each answer still lands on the instance that gave it.
func TestStatusRunsTheScriptsConcurrently(t *testing.T) {
	names := []string{"pod-a", "pod-b", "pod-c", "pod-d", "pod-e"}
	withScriptWorkers(t, 2)
	r := newTargetRunner(podListJSON(names...), 2)
	args := []string{"status", "app", "--output", "json", "--platform", "kubernetes"}
	for _, n := range names {
		args = append(args, "--pod", n)
	}
	var code int
	stdout := captureStdout(t, func() { code = dispatch(args, r) })
	if code != 0 {
		t.Fatalf("exit=%d, want 0", code)
	}
	if r.timedOut {
		t.Error("the first two execs never overlapped: the scripts ran one after another")
	}
	if r.maxInFlight != 2 {
		t.Errorf("%d execs were in flight at once, want exactly the 2-worker limit", r.maxInFlight)
	}
	var doc struct {
		Instances []struct {
			Name        string `json:"name"`
			Application struct {
				Mode string `json:"leaderElectionMode"`
			} `json:"application"`
		} `json:"instances"`
	}
	if err := json.Unmarshal([]byte(stdout), &doc); err != nil {
		t.Fatalf("stdout must be one json document: %v\n%s", err, stdout)
	}
	if len(doc.Instances) != len(names) {
		t.Fatalf("want %d instances, got %d:\n%s", len(names), len(doc.Instances), stdout)
	}
	for _, inst := range doc.Instances {
		if inst.Application.Mode != "mode-"+inst.Name {
			t.Errorf("%s carries %q, another instance's answer", inst.Name, inst.Application.Mode)
		}
	}
}

// ---- output format ---------------------------------------------------------------

// TestStatusJSONOutputIsOneDocument covers --output json: the same model the
// tables render, as one parseable document on stdout.
func TestStatusJSONOutputIsOneDocument(t *testing.T) {
	q := &queueRunner{resp: []queuedResp{
		{onePodJSON, nil},
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
		q := &queueRunner{resp: []queuedResp{{onePodJSON, nil}, {deployment, nil}}}
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
			{onePodJSON, nil}, {deployment, nil},
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
		q := &queueRunner{resp: []queuedResp{{onePodJSON, nil}, {deployment, nil}}}
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
		if code != 2 {
			t.Errorf("--user %q: exit=%d, want 2", user, code)
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
			if code := dispatch(c.args, f); code != 2 {
				t.Errorf("%s: exit=%d, want 2", c.name, code)
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
			if code := dispatch(args, f); code != 2 {
				t.Errorf("--management-port %s: exit=%d, want 2", p, code)
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
	q := &queueRunner{resp: []queuedResp{{`{"items":[]}`, nil}}}
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
	list := "bad;name\t" + img + "\tUp 3 days\nev\x1b[2Jil\t" + img + "\tUp 3 days\nsolmq-connector\t" + img + "\tUp 3 days\n"
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
		if containsToken(c.argv, "bad;name") || containsToken(c.argv, "ev\x1b[2Jil") {
			t.Errorf("an unsafe name must never reach an argv: %v", c.argv)
		}
	}
	// A name carrying a terminal escape is skipped too (containercli.ValidName),
	// and the note quotes it, so the escape is printed as text, never sent.
	if strings.Contains(stdout, "\x1b") {
		t.Errorf("a control character from the engine reached the terminal: %q", stdout)
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
	list := "zeta\t" + img + "\tUp 3 days\nalpha\t" + img + "\tUp 3 days\n"
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

// ---- progress: the spinner, and --verbose steps ---------------------------------
//
// The surface is stderr-only, so the cases here always assert two things: what
// the operator sees while collection runs, and that stdout is untouched by it.
// The report is the artifact, and a `1>` capture has to be byte-identical with
// and without a terminal to spin on.

// withStderrTerminal injects an answer for the stderrIsTerminal seam so a test
// can ask for the spinner: go test's own stderr is a pipe, which would otherwise
// make the default rendering unreachable. withPrompter is the stdin counterpart.
func withStderrTerminal(t *testing.T, yes bool) {
	t.Helper()
	prev := stderrIsTerminal
	stderrIsTerminal = func() bool { return yes }
	t.Cleanup(func() { stderrIsTerminal = prev })
}

// withProgressClock pins progressNow to a clock that advances by step on every
// read, so an elapsed time is arithmetic rather than timing: each step reports
// exactly one step's worth whatever the machine was doing. Atomic because the
// frame goroutine reads the same clock.
func withProgressClock(t *testing.T, step time.Duration) {
	t.Helper()
	prev := progressNow
	base := time.Date(2026, 8, 18, 4, 0, 0, 0, time.UTC)
	var reads int64
	progressNow = func() time.Time {
		return base.Add(time.Duration(atomic.AddInt64(&reads, 1)) * step)
	}
	t.Cleanup(func() { progressNow = prev })
}

// withProgressFrameRate replaces the spinner's tick. A case wants one of two
// extremes: an hour, so only the draws the call itself makes are captured, or a
// millisecond, so the goroutine's own redraw is observable without a sleep.
func withProgressFrameRate(t *testing.T, d time.Duration) {
	t.Helper()
	prev := progressFrameRate
	progressFrameRate = d
	t.Cleanup(func() { progressFrameRate = prev })
}

// withScriptWorkers sets how many instances the status script runs on at once.
// 1 pins the calls to report order, which a queueRunner's ordered answers need
// once a test has more than one instance.
func withScriptWorkers(t *testing.T, n int) {
	t.Helper()
	prev := statusScriptWorkers
	statusScriptWorkers = n
	t.Cleanup(func() { statusScriptWorkers = prev })
}

// progressErase is what blanking a live line of n columns looks like on the
// wire: back to column 0, n spaces, back again. Spelled out here so the
// expectations below read as sequences of draws and erases.
func progressErase(n int) string { return "\r" + strings.Repeat(" ", n) + "\r" }

// pipeReader is an os.Pipe with an incremental drain, for the one case
// captureStderr cannot serve: the spinner's goroutine writes while the call
// under test is still running, so the test has to see what has been written so
// far rather than only the total.
type pipeReader struct {
	w     *os.File
	seen  string
	chunk chan string
}

func newPipeReader(t *testing.T) *pipeReader {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	p := &pipeReader{w: w, chunk: make(chan string, 64)}
	go func() {
		buf := make([]byte, 512)
		for {
			n, rerr := r.Read(buf)
			if n > 0 {
				p.chunk <- string(buf[:n])
			}
			if rerr != nil {
				close(p.chunk)
				return
			}
		}
	}()
	t.Cleanup(func() {
		_ = w.Close()
		_ = r.Close()
	})
	return p
}

// waitFor accumulates writes until want appears and returns everything seen so
// far. Bounded rather than a sleep: what it waits for is a goroutine's next
// tick, and a fixed sleep would be both slower and flakier.
func (p *pipeReader) waitFor(t *testing.T, want string) string {
	t.Helper()
	deadline := time.After(10 * time.Second)
	for !strings.Contains(p.seen, want) {
		select {
		case c, ok := <-p.chunk:
			if !ok {
				t.Fatalf("the stream closed before %q appeared, got:\n%q", want, p.seen)
			}
			p.seen += c
		case <-deadline:
			t.Fatalf("timed out waiting for %q, got:\n%q", want, p.seen)
		}
	}
	return p.seen
}

// TestNewProgressPicksOneRendering pins the precedence the flags were settled
// with: --watch owns the screen, --verbose is an explicit request and so does
// not need a terminal, and only the spinner does.
func TestNewProgressPicksOneRendering(t *testing.T) {
	cases := []struct {
		name string
		opts statusOpts
		tty  bool
		want progressMode
	}{
		{"watch beats verbose and a terminal", statusOpts{watch: watchFlag{on: true}, verbose: true}, true, progressOff},
		{"verbose prints without a terminal", statusOpts{verbose: true}, false, progressSteps},
		{"a terminal gets the spinner", statusOpts{}, true, progressSpinner},
		{"no terminal and no verbose: nothing", statusOpts{}, false, progressOff},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			withStderrTerminal(t, c.tty)
			if got := newProgress(c.opts, os.Stderr).mode; got != c.want {
				t.Errorf("mode=%d, want %d", got, c.want)
			}
		})
	}
}

// TestStderrIsTerminalReportsAPipe covers the real probe every other case
// injects over: a redirected stderr is not a screen, which is what keeps a
// rewritten line out of `2>steps.log`.
func TestStderrIsTerminalReportsAPipe(t *testing.T) {
	var got bool
	captureStderr(t, func() { got = stderrIsTerminal() })
	if got {
		t.Error("a pipe is not a terminal")
	}
}

// TestProgressOffAndNilWriteNothing is the rule every call site depends on:
// nothing branches on the mode, so both the nil receiver and progressOff have to
// be silent -- and pause still has to run what it was given.
func TestProgressOffAndNilWriteNothing(t *testing.T) {
	var nilp *progress
	paused := 0
	out := captureStderr(t, func() {
		nilp.step("preflight")
		nilp.pause(func() { paused++ })
		nilp.stop()

		off := &progress{w: os.Stderr, mode: progressOff}
		off.step("preflight")
		off.pause(func() { paused++ })
		off.stop()
	})
	if out != "" {
		t.Errorf("nothing should reach stderr, got %q", out)
	}
	if paused != 2 {
		t.Errorf("pause must still run its function, ran %d/2", paused)
	}
}

// TestProgressSpinnerRewritesOneLineAndErases pins the whole rendering as bytes:
// one line, rewritten in place, and erased when the step ends -- no ANSI, so it
// needs no VT processing, and nothing on stdout.
func TestProgressSpinnerRewritesOneLineAndErases(t *testing.T) {
	withProgressFrameRate(t, time.Hour) // only the draws step and stop make themselves
	withProgressClock(t, 0)             // a pinned clock: no seconds counter in this case
	var p *progress
	stdout := captureStdout(t, func() {
		stderr := captureStderr(t, func() {
			p = &progress{w: os.Stderr, mode: progressSpinner}
			p.step("list pods")
			p.step("read service solmq-connector")
			p.stop()
			p.stop() // idempotent: the second one has nothing left to end
		})
		const first, second = "- list pods", "- read service solmq-connector"
		want := "\r" + first + progressErase(len(first)) + "\r" + second + progressErase(len(second))
		if stderr != want {
			t.Errorf("stderr:\n%q\nwant:\n%q", stderr, want)
		}
	})
	if stdout != "" {
		t.Errorf("the spinner must never reach stdout, got %q", stdout)
	}
	// stop joined the goroutine rather than leaving it to draw over whatever is
	// written next -- which in this suite includes a swapped-back os.Stderr.
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.quit != nil || p.joined != nil {
		t.Error("stop must leave no frame goroutine running")
	}
}

// TestProgressSpinnerGoroutineAdvancesAndCountsSeconds is the one case that has
// to watch the live line while it is still live, so it writes to its own pipe:
// captureStderr only hands back what was written in total.
func TestProgressSpinnerGoroutineAdvancesAndCountsSeconds(t *testing.T) {
	withProgressFrameRate(t, time.Millisecond)
	withProgressClock(t, 600*time.Millisecond)
	const label = "status script pod-a 1/2"
	pr := newPipeReader(t)
	p := &progress{w: pr.w, mode: progressSpinner}

	p.step(label)
	// The second frame is the goroutine's work, not step's own first draw.
	pr.waitFor(t, "\\ "+label)
	// And once the step passes a second the counter appears -- whole seconds, so
	// what an operator watches is a number that changes once a second.
	pr.waitFor(t, label+"  1s")

	p.stop()
	fmt.Fprint(pr.w, "the report")
	seen := pr.waitFor(t, "the report")
	// Whatever width the line had reached, the bytes before the report are an
	// erase: spaces, then back to column 0.
	if before := seen[:strings.LastIndex(seen, "the report")]; !strings.HasSuffix(before, " \r") {
		t.Errorf("the live line should be blanked before the report, got:\n%q", before)
	}
}

// TestProgressPauseHandsBackTheStream pins the ordering the install prompt
// needs: the live line is gone before the question is asked, no frame is drawn
// over it while it waits for an answer, and the next step opens a fresh line.
func TestProgressPauseHandsBackTheStream(t *testing.T) {
	withProgressFrameRate(t, time.Hour)
	withProgressClock(t, 0)
	const runs = "- status script on 1 instance(s)"
	const installs = "- install script pod-a 1/1"
	const question = "install the status script on pod-a? [y/N] "
	stderr := captureStderr(t, func() {
		p := &progress{w: os.Stderr, mode: progressSpinner}
		p.step("status script on 1 instance(s)")
		p.pause(func() { fmt.Fprint(os.Stderr, question) })
		p.step("install script pod-a 1/1")
		p.stop()
	})
	want := "\r" + runs + progressErase(len(runs)) + question + "\r" + installs + progressErase(len(installs))
	if stderr != want {
		t.Errorf("stderr:\n%q\nwant:\n%q", stderr, want)
	}
}

// TestProgressStepsPauseFinishesTheStepFirst is the same handover under
// --verbose: the step ends before the question, so its line reports the time the
// call took rather than the time the operator took to answer.
func TestProgressStepsPauseFinishesTheStepFirst(t *testing.T) {
	withProgressClock(t, 600*time.Millisecond)
	const question = "install the status script on pod-a? [y/N] "
	stderr := captureStderr(t, func() {
		p := &progress{w: os.Stderr, mode: progressSteps}
		p.step("status script on 1 instance(s)")
		p.pause(func() { fmt.Fprint(os.Stderr, question) })
		p.stop()
	})
	want := "step: status script on 1 instance(s) 0.6s\n" + question
	if stderr != want {
		t.Errorf("stderr:\n%q\nwant:\n%q", stderr, want)
	}
}

// TestProgressDoneLinesCarryTheirOwnTime pins the line a call run alongside
// others gets: under --verbose its own time, printed at once, while the batch
// step stays in flight and is timed as a whole. The spinner and the off mode
// print nothing for it.
func TestProgressDoneLinesCarryTheirOwnTime(t *testing.T) {
	withProgressFrameRate(t, time.Hour)
	withProgressClock(t, 600*time.Millisecond)
	steps := captureStderr(t, func() {
		p := &progress{w: os.Stderr, mode: progressSteps}
		p.step("status script on 2 instance(s)")
		p.done(2100*time.Millisecond, "status script %s %d/%d", "pod-b", 1, 2)
		p.done(90*time.Second, "status script %s %d/%d", "pod-a", 2, 2)
		p.stop()
	})
	want := "step: status script pod-b 1/2 2.1s\n" +
		"step: status script pod-a 2/2 1m30s\n" +
		"step: status script on 2 instance(s) 0.6s\n"
	if steps != want {
		t.Errorf("stderr:\n%q\nwant:\n%q", steps, want)
	}

	for _, mode := range []progressMode{progressSpinner, progressOff} {
		got := captureStderr(t, func() {
			p := &progress{w: os.Stderr, mode: mode}
			p.step("status script on 1 instance(s)")
			p.done(time.Second, "status script %s %d/%d", "pod-a", 1, 1)
			p.stop()
		})
		if strings.Contains(got, "pod-a") {
			t.Errorf("mode %d printed a done line, got:\n%q", mode, got)
		}
	}
	var none *progress
	none.done(time.Second, "status script %s %d/%d", "pod-a", 1, 1) // a nil progress is a no-op
}

// TestProgressElapsedShapes pins the two shapes a step's duration is printed in.
// The negative case is not hypothetical: progressNow is a var, and a clock that
// goes backwards should read as no time at all rather than as a minus sign.
func TestProgressElapsedShapes(t *testing.T) {
	cases := []struct {
		d    time.Duration
		want string
	}{
		{-time.Second, "0.0s"},
		{0, "0.0s"},
		{300 * time.Millisecond, "0.3s"},
		{4100 * time.Millisecond, "4.1s"},
		{59500 * time.Millisecond, "59.5s"},
		{time.Minute, "1m00s"},
		{125 * time.Second, "2m05s"},
	}
	for _, c := range cases {
		if got := progressElapsed(c.d); got != c.want {
			t.Errorf("progressElapsed(%s)=%s, want %s", c.d, got, c.want)
		}
	}
}

// TestProgressTrimElidesTheMiddle covers the one-row cap. The tail is the half
// that has to survive: it carries the instance and its i/N, which is what says
// the run is advancing rather than stuck on the same exec.
func TestProgressTrimElidesTheMiddle(t *testing.T) {
	const full = "status script solmq-connector-7d9f8c-x2n4q 1/2"
	for _, c := range []struct {
		name  string
		label string
		width int
		want  string
	}{
		{"shorter than the line", "list pods", 20, "list pods"},
		{"exactly the line", "list pods", 9, "list pods"},
		{"middle elided, both ends kept", "abcdefghij", 9, "abc...hij"},
		{"odd remainder favours the head", "abcdefghij", 8, "abc...ij"},
		{"a real label keeps its i/N", full, 20, "status sc...2n4q 1/2"},
		{"no room for the elision", "abcdefghij", 3, "abc"},
		{"one column", "abcdefghij", 1, "a"},
		{"no columns", "abcdefghij", 0, ""},
		{"negative, which the head and tail widths can produce", "abcdefghij", -4, ""},
	} {
		t.Run(c.name, func(t *testing.T) {
			got := progressTrim(c.label, c.width)
			if got != c.want {
				t.Errorf("progressTrim(%q, %d) = %q, want %q", c.label, c.width, got, c.want)
			}
			if len(got) > c.width && c.width >= 0 {
				t.Errorf("%q is wider than the %d columns asked for", got, c.width)
			}
		})
	}
}

// TestProgressSpinnerNeverExceedsOneRow is why the trim exists: a line that
// wraps leaves a row above the cursor that \r cannot reach, so the erase would
// leave residue for the report to land under. Nothing the spinner writes --
// neither a draw nor the blanking that follows it -- may be wider than the cap.
func TestProgressSpinnerNeverExceedsOneRow(t *testing.T) {
	withProgressFrameRate(t, time.Hour)
	withProgressClock(t, 0)
	label := "status script " + strings.Repeat("a", 100) + " 12/34"
	stderr := captureStderr(t, func() {
		p := &progress{w: os.Stderr, mode: progressSpinner}
		p.step("%s", label) // step is printf-like, and this label is built, not constant
		p.stop()
	})
	for _, seg := range strings.Split(stderr, "\r") {
		if len(seg) > progressLineWidth {
			t.Errorf("segment is %d columns, over the %d cap:\n%q", len(seg), progressLineWidth, seg)
		}
	}
	if !strings.Contains(stderr, "...") || !strings.Contains(stderr, "12/34") {
		t.Errorf("the label should be elided with its i/N kept, got:\n%q", stderr)
	}
}

// TestStatusVerboseNamesEverySlowCall is the point of --verbose on a slow
// environment: one durable line per call, in the order they were made, each with
// the time it took -- and nothing on stdout for a parser to trip over.
func TestStatusVerboseNamesEverySlowCall(t *testing.T) {
	withProgressClock(t, 600*time.Millisecond)
	withStderrTerminal(t, false) // --verbose prints whether or not there is a screen
	withScriptWorkers(t, 1)      // one worker reads the clock in a fixed order
	q := &queueRunner{resp: []queuedResp{
		{podsJSON, nil},  // get pods
		{appReport, nil}, // pod-a run
		{appReport, nil}, // pod-b run
	}}
	var code int
	var stderr string
	stdout := captureStdout(t, func() {
		stderr = captureStderr(t, func() {
			code = dispatch([]string{"status", "app", "-v", "--platform", "kubernetes", "--pod", "pod-a", "--pod", "pod-b"}, q)
		})
	})
	if code != 0 {
		t.Fatalf("exit=%d, want 0", code)
	}
	var steps []string
	for _, ln := range strings.Split(stderr, "\n") {
		if strings.HasPrefix(ln, "step: ") {
			steps = append(steps, ln)
		}
	}
	// No "preflight" line: kubernetes asks no permission up front, so there is no
	// call for the step to name. No "resolve instance names" line either: both
	// --pod values here are names, and resolveIndexes only enumerates when an
	// index was typed. Each script run gets its own line with its own time as it
	// finishes, and the batch they ran in is one step around them: its 3.0s is
	// the five clock reads between its start and its end.
	want := []string{
		"step: list pods 0.6s",
		"step: status script pod-a 1/2 0.6s",
		"step: status script pod-b 2/2 0.6s",
		"step: status script on 2 instance(s) 3.0s",
	}
	if !reflect.DeepEqual(steps, want) {
		t.Errorf("steps:\n%s\nwant:\n%s", strings.Join(steps, "\n"), strings.Join(want, "\n"))
	}
	if strings.Contains(stdout, "step: ") {
		t.Errorf("no step line may reach stdout, got:\n%s", stdout)
	}
}

// TestStatusVerboseUnderWatchSaysWhichWins keeps the pair from being silently
// ignored. They do not contradict each other -- the redraw simply owns the
// screen -- so this is a note rather than a refusal, said out loud either way.
func TestStatusVerboseUnderWatchSaysWhichWins(t *testing.T) {
	for _, spelling := range []string{"-v", "--verbose"} {
		t.Run(spelling, func(t *testing.T) {
			// The note is the last thing checkStatusFlags does, so reaching it
			// means every refusal passed -- which would otherwise drop this case
			// into the redraw loop. A failing preflight is what ends the run
			// instead: it is the first call actStatus makes, before the loop. docker
			// has one where kubernetes, which asks nothing up front, does not.
			f := &fakeRunner{err: fmt.Errorf("no engine")}
			var code int
			stderr := captureStderr(t, func() {
				code = dispatch([]string{"status", "container", "-w", spelling, "--platform", "docker", "--container", "solmq-connector"}, f)
			})
			if code != 1 {
				t.Fatalf("exit=%d, want 1", code)
			}
			if !strings.Contains(stderr, "note: --verbose prints no steps under --watch") {
				t.Errorf("the winner should be named, got:\n%s", stderr)
			}
			if len(f.calls) != 1 {
				t.Errorf("only the preflight probe should run, got %d calls", len(f.calls))
			}
		})
	}
}

// TestStatusStdoutIsIdenticalWithAndWithoutTheSpinner is the stream contract the
// whole feature rests on: the document is the artifact, so a redirected stdout
// cannot depend on whether there was a terminal to spin on.
func TestStatusStdoutIsIdenticalWithAndWithoutTheSpinner(t *testing.T) {
	withProgressFrameRate(t, time.Hour)
	withProgressClock(t, 0)
	run := func(t *testing.T, tty bool) (string, string) {
		withStderrTerminal(t, tty)
		q := &queueRunner{resp: []queuedResp{
			{onePodJSON, nil},
			{appReport, nil},
		}}
		var out string
		errs := captureStderr(t, func() {
			out = captureStdout(t, func() {
				if code := dispatch([]string{"status", "app", "--output", "json", "--platform", "kubernetes", "--pod", "pod-a"}, q); code != 0 {
					t.Fatalf("exit=%d, want 0", code)
				}
			})
		})
		return out, errs
	}
	spinning, frames := run(t, true)
	quiet, silence := run(t, false)
	if spinning != quiet {
		t.Errorf("stdout differs with the spinner on:\n%q\nwithout:\n%q", spinning, quiet)
	}
	if !strings.Contains(frames, "\r- ") {
		t.Errorf("the terminal case should have spun, got:\n%q", frames)
	}
	if silence != "" {
		t.Errorf("a redirected stderr should stay empty, got:\n%q", silence)
	}
	var doc map[string]any
	if err := json.Unmarshal([]byte(spinning), &doc); err != nil {
		t.Errorf("the document must still parse with the spinner on: %v", err)
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

// streamed is one log read's canned result. interrupt makes the fake deliver
// an os.Interrupt to this process once the read is running and then wait for the
// follow's context to end, as an operator's Ctrl-C does.
type streamed struct {
	stdout    string
	stderr    string
	err       error
	interrupt bool
}

// streamRunner is the fake logs needs: logs reads every log with Run writing to
// the process's own stdout and stderr (so the platform's own diagnostics stay on
// stderr rather than landing in a redirected log), while the preflight probe and
// discovery capture theirs. Both kinds of call are recorded in one ordered slice,
// while the canned answers are queued per kind, told apart by the stdout writer
// the call carries: captured calls are served from resp, reads from streams.
type streamRunner struct {
	noAttach
	calls []fakeCall

	resp []queuedResp
	runN int

	streams []streamed
	streamN int
}

func (s *streamRunner) Run(ctx context.Context, c gdurunner.Cmd, stdout, stderr io.Writer) error {
	s.calls = append(s.calls, fakeCall{argv: c.Argv, stdin: string(c.Stdin), env: c.Env})
	if stdout != io.Writer(os.Stdout) {
		i := s.runN
		s.runN++
		if i < len(s.resp) {
			scriptOut(stdout, s.resp[i].out)
			return s.resp[i].err
		}
		return nil
	}
	i := s.streamN
	s.streamN++
	if i >= len(s.streams) {
		return nil
	}
	st := s.streams[i]
	scriptOut(stdout, st.stdout)
	scriptOut(stderr, st.stderr)
	if st.interrupt {
		p, err := os.FindProcess(os.Getpid())
		if err == nil {
			err = p.Signal(os.Interrupt)
		}
		if err != nil {
			return fmt.Errorf("sending the interrupt: %w", err)
		}
		select {
		case <-ctx.Done():
			return fmt.Errorf("%s: stopped: %w", c.Argv[0], ctx.Err())
		case <-time.After(5 * time.Second):
			return errors.New("the interrupt never ended the follow")
		}
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
			// containercli.NoPrevious spells the engine as a message does (Docker).
			for _, w := range []string{"--previous", "kubernetes", platform} {
				if !strings.Contains(strings.ToLower(stderr), w) {
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
	q := &streamRunner{}
	if code := dispatch([]string{"logs", "--platform", "kubernetes", "--pod", "pod-a", "--previous"}, q); code != 0 {
		t.Fatalf("exit=%d, want 0", code)
	}
	if !containsToken(q.calls[0].argv, "-p") {
		t.Errorf("--previous should reach the argv as -p, got %v", q.calls[0].argv)
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

// TestLogsFollowCancelledByTheOperatorEndsClean pins the A-specific policy over
// the shared runner's cancel rule: the runner reports an interrupted follow as an
// error wrapping context.Canceled, and logs reads that as the operator ending
// the follow, so it exits 0 with nothing on stderr.
func TestLogsFollowCancelledByTheOperatorEndsClean(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("a process cannot be sent os.Interrupt on windows")
	}
	q := &streamRunner{
		resp:    []queuedResp{{"", nil}},
		streams: []streamed{{interrupt: true}},
	}
	var code int
	stderr := captureStderr(t, func() {
		code = dispatch([]string{"logs", "--platform", "docker", "--container", "solmq-conn", "--follow"}, q)
	})
	if code != 0 {
		t.Fatalf("exit=%d, want 0 (Ctrl-C is how a follow ends)", code)
	}
	if stderr != "" {
		t.Errorf("a clean end should print nothing, got:\n%s", stderr)
	}
}

// TestLogsFollowFailureIsNotACleanEnd pins the other half: a follow that fails
// without the operator cancelling it exits 1 and names the instance -- including
// a failure that merely wraps context.Canceled while the follow's own context
// is still live -- and a one-shot read is never a clean end whatever it wraps.
func TestLogsFollowFailureIsNotACleanEnd(t *testing.T) {
	for _, c := range []struct {
		name string
		args []string
		err  error
	}{
		{"follow, engine failure", []string{"--follow"}, fmt.Errorf("docker: exit status 1")},
		{"follow, canceled error but a live context", []string{"--follow"}, fmt.Errorf("docker: %w", context.Canceled)},
		{"one-shot read, canceled error", nil, fmt.Errorf("docker: %w", context.Canceled)},
	} {
		t.Run(c.name, func(t *testing.T) {
			q := &streamRunner{
				resp:    []queuedResp{{"", nil}},
				streams: []streamed{{stderr: "daemon said no\n", err: c.err}},
			}
			args := append([]string{"logs", "--platform", "docker", "--container", "solmq-conn"}, c.args...)
			var code int
			stderr := captureStderr(t, func() { code = dispatch(args, q) })
			if code != 1 {
				t.Fatalf("exit=%d, want 1", code)
			}
			if !strings.Contains(stderr, "reading solmq-conn") {
				t.Errorf("the error should name the instance, got:\n%s", stderr)
			}
		})
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
			if code != 2 {
				t.Fatalf("exit=%d, want 2", code)
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
	withPrompter(t, true, "1\n") // kubernetes, the first section present

	q := &streamRunner{resp: []queuedResp{{onePodJSON, nil}}}
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
	if !containsToken(q.calls[0].argv, "app=solmq-connector") {
		t.Errorf("discovery should use the deployment selector, got %v", q.calls[0].argv)
	}
	if !containsToken(q.calls[0].argv, "solace-connectors") {
		t.Errorf("discovery should use the deployment namespace, got %v", q.calls[0].argv)
	}
}

// TestLogsPlatformFlagSkipsTheMenu is the same env.yaml with the flag given: the
// flag is the first step of the resolution order, so nothing is asked.
func TestLogsPlatformFlagSkipsTheMenu(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "env.yaml", sharedEnv+kubeEnv+dockerEnv)
	// Answering the menu with kubernetes would be wrong; the flag must win
	// before the prompter is ever consulted.
	withPrompter(t, true, "1\n")

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
	if code != 2 {
		t.Fatalf("exit=%d, want 2", code)
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

	q := &streamRunner{resp: []queuedResp{{`{"items":[]}`, nil}}}
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
	if code != 2 {
		t.Fatalf("exit=%d, want 2", code)
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

// TestReadVerbsAskKubernetesNoPermissionFirst is CR-0006 taken to its end: status,
// logs and cli change nothing, so on kubernetes they ask for no permission up front.
// The first call is the read itself, and a view-only operator is never turned away by
// a question -- what the cluster refuses, the read reports in the platform's own
// words.
func TestReadVerbsAskKubernetesNoPermissionFirst(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "env.yaml", sharedEnv+kubeEnv)
	t.Chdir(dir)
	forbidden := `Error from server (Forbidden): pods is forbidden: User "u" cannot list resource "pods" in namespace "prod"`
	refused := fmt.Errorf("exit status 1")
	status := &queueRunner{resp: []queuedResp{{forbidden, refused}}}
	logs := &streamRunner{streams: []streamed{{stderr: forbidden, err: refused}}}
	cli := &attachRunner{attaches: []attached{{code: 1, err: refused}}}
	withPrompter(t, true, "")
	for _, c := range []struct {
		args  []string
		r     gdurunner.Runner
		calls func() []fakeCall
		want  []string // the first call's leading arguments after the command
	}{
		{[]string{"status", "container", "--platform", "kubernetes", "--namespace", "prod"}, status, func() []fakeCall { return status.calls }, []string{"get", "pods"}},
		{[]string{"logs", "--platform", "kubernetes", "--pod", "p", "--namespace", "prod"}, logs, func() []fakeCall { return logs.calls }, []string{"logs", "p"}},
		{[]string{"cli", "--platform", "kubernetes", "--pod", "p", "--namespace", "prod"}, cli, func() []fakeCall { return cli.calls }, []string{"exec"}},
	} {
		var code int
		captureStderr(t, func() { code = dispatch(c.args, c.r) })
		if code == 0 {
			t.Errorf("%s: a refused read must not exit 0", c.args[0])
		}
		calls := c.calls()
		if len(calls) == 0 || !containsToken(calls[0].argv, c.want[0]) || (len(c.want) > 1 && !containsToken(calls[0].argv, c.want[1])) {
			t.Errorf("%s: the first call should be the read itself (%v), got %+v", c.args[0], c.want, calls)
		}
		if asked := canIAsked(calls); len(asked) != 0 {
			t.Errorf("%s: kubernetes asks no permission up front, got %v", c.args[0], asked)
		}
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
		streams: []streamed{{stderr: "Error from server (NotFound): pods \"ghost\" not found\n", err: fmt.Errorf("exit status 1")}},
	}
	var code int
	stderr := captureStderr(t, func() {
		code = dispatch([]string{"logs", "--platform", "kubernetes", "--pod", "ghost", "--namespace", "prod"}, q)
	})
	if code != 1 {
		t.Fatalf("exit=%d, want 1", code)
	}
	// No permission question either: the read is the whole run.
	if len(q.calls) != 1 {
		t.Fatalf("want 1 call (the read), got %d: %+v", len(q.calls), q.calls)
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

	q := &streamRunner{resp: []queuedResp{{onePodJSON, nil}}}
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
	q := &streamRunner{}
	if code := dispatch([]string{"logs", "--platform", "kubernetes", "--pod", "pod-a", "--command", "oc"}, q); code != 0 {
		t.Fatalf("exit=%d, want 0", code)
	}
	if q.calls[0].argv[0] != "oc" {
		t.Errorf("--command should pick the binary, got %v", q.calls[0].argv)
	}

	// A binary outside the allowlist is refused before anything runs, exactly as
	// it is for deploy/remove/status, and it is argv, so the exit is a usage error.
	q2 := &streamRunner{}
	var code int
	stderr := captureStderr(t, func() {
		code = dispatch([]string{"logs", "--platform", "kubernetes", "--pod", "pod-a", "--command", "curl"}, q2)
	})
	if code != 2 {
		t.Fatalf("exit=%d, want 2", code)
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
		streams: []streamed{{stdout: "line-a\n"}},
	}
	var code int
	stdout := captureStdout(t, func() {
		code = dispatch([]string{"logs", "--platform", "kubernetes", "--namespace", "prod", "--pod", "pod-a"}, q)
	})
	if code != 0 {
		t.Fatalf("exit=%d, want 0", code)
	}
	if len(q.calls) != 1 {
		t.Fatalf("want 1 call (the log read), got %d: %+v", len(q.calls), q.calls)
	}
	if want := []string{"kubectl", "logs", "pod-a", "-n", "prod", "-c", spec.ConnectorContainerName}; !reflect.DeepEqual(q.calls[0].argv, want) {
		t.Errorf("log argv = %v, want %v", q.calls[0].argv, want)
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
			if code != 2 {
				t.Fatalf("exit=%d, want 2", code)
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

	q := &streamRunner{resp: []queuedResp{{unsortedPodsJSON, nil}}}
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
		resp:    []queuedResp{{onePodJSON, nil}},
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
				resp:    []queuedResp{{unsortedPodsJSON, nil}},
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
		resp:    []queuedResp{{podsNamedZero, nil}},
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

	q := &streamRunner{resp: []queuedResp{{unsortedPodsJSON, nil}}}
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
// reporting a missing pod named "0". With no env file there is nothing to
// discover from, which is about what the tool was given, so the exit is 2.
func TestLogsIndexWithNothingToIndexInto(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "no-such-env.yaml")
	q := &streamRunner{}
	var code int
	stderr := captureStderr(t, func() {
		code = dispatch([]string{"logs", "--platform", "kubernetes", "-e", missing, "--pod", "0"}, q)
	})
	if code != 2 {
		t.Fatalf("exit=%d, want 2", code)
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
		{unsortedPodsJSON, nil}, // enumeration for the index
		{unsortedPodsJSON, nil}, // the report query itself
	}}
	if code := dispatch([]string{"status", "container", "-e", filepath.Join(dir, "env.yaml"), "--pod", "1"}, q); code != 0 {
		t.Fatalf("exit=%d, want 0", code)
	}
	// calls: 0 the enumeration for the index, 1 the report query. Later calls
	// are the workload summary, which is not what this pins.
	report := q.calls[1].argv
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
	q := &queueRunner{resp: []queuedResp{{podsJSON, nil}}}
	if code := dispatch([]string{"status", "container", "--platform", "kubernetes", "--pod", "pod-a", "--pod", "pod-b"}, q); code != 0 {
		t.Fatalf("exit=%d, want 0", code)
	}
	argv := q.calls[0].argv
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

	q := &streamRunner{resp: []queuedResp{{unsortedPodsJSON, nil}}}
	var code int
	stdout := captureStdout(t, func() {
		code = dispatch([]string{
			"logs", "-e", envPath,
			"--namespace", "prod",
			"--command", "oc",
			"--allow-command", "lima",
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
		"--allow-command lima",
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
	withPodman := &spec.Env{Podman: &spec.Podman{Command: "podman --remote"}}

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
		{"podman section", validate.PlatformPodman, "", withPodman, "podman --remote"},
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
// The first line is the one a script reads the version from, so the support
// notice must follow it rather than change it.
func TestVersionOutputShape(t *testing.T) {
	var code int
	stdout := captureStdout(t, func() { code = run([]string{"version"}) })
	if code != 0 {
		t.Fatalf("exit=%d, want 0", code)
	}
	first := fmt.Sprintf("solmq-conn-util dev %s %s/%s\n", runtime.Version(), runtime.GOOS, runtime.GOARCH)
	want := first + strings.Join(notice.Short(), "\n") + "\n"
	if stdout != want {
		t.Errorf("stdout = %q, want %q", stdout, want)
	}
	if !strings.HasPrefix(stdout, first) {
		t.Errorf("the version line must come first and unchanged, got %q", stdout)
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

// attachRunner is the fake cli needs: cli reaches its instance through Attach
// while the preflight probe and discovery go through Run. Both kinds of call are
// recorded in one ordered slice, while the canned answers are queued per kind --
// the same split streamRunner uses for logs.
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

func (a *attachRunner) Run(_ context.Context, c gdurunner.Cmd, stdout, _ io.Writer) error {
	a.calls = append(a.calls, fakeCall{argv: c.Argv, stdin: string(c.Stdin), env: c.Env})
	i := a.runN
	a.runN++
	if i < len(a.resp) {
		scriptOut(stdout, a.resp[i].out)
		return a.resp[i].err
	}
	return nil
}

func (a *attachRunner) Attach(c gdurunner.Cmd, _, _, _ *os.File) (int, error) {
	a.calls = append(a.calls, fakeCall{argv: c.Argv, env: c.Env})
	i := a.attachN
	a.attachN++
	if i >= len(a.attaches) {
		return 0, nil
	}
	return a.attaches[i].code, a.attaches[i].err
}

// TestCliRefusesAShellWithoutATerminal is the same non-TTY contract the
// platform menu, the status install prompt and the remove confirmation keep:
// fail with the next step rather than open a session nobody can type into. The
// next step here is the one-shot form, so the message has to name it.
func TestCliRefusesAShellWithoutATerminal(t *testing.T) {
	withPrompter(t, false, "")
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
// discovery call, and kubernetes asks no permission first, so the whole run is
// the attach.
func TestCliKubernetesArgvShape(t *testing.T) {
	withPrompter(t, true, "")
	q := &attachRunner{}
	if code := dispatch([]string{"cli", "--platform", "kubernetes", "--pod", "pod-a", "--namespace", "prod"}, q); code != 0 {
		t.Fatalf("exit=%d, want 0", code)
	}
	if len(q.calls) != 1 {
		t.Fatalf("want 1 call (the attach), got %d: %+v", len(q.calls), q.calls)
	}
	want := []string{"kubectl", "exec", "-i", "-t", "pod-a", "-n", "prod", "-c", spec.ConnectorContainerName, "--", runner.ContainerShell}
	if !reflect.DeepEqual(q.calls[0].argv, want) {
		t.Errorf("session argv = %v, want %v", q.calls[0].argv, want)
	}
}

// TestCliEngineArgvShape is the other half: docker and podman stop parsing
// their own flags at the container name, so the options come first and there is
// no "--" and no -c to add.
func TestCliEngineArgvShape(t *testing.T) {
	for _, platform := range []string{"docker", "podman"} {
		t.Run(platform, func(t *testing.T) {
			withPrompter(t, true, "")
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
	withPrompter(t, true, "")
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
	withPrompter(t, false, "")
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
	withPrompter(t, true, "")
	q := &attachRunner{attaches: []attached{{code: 3}}}
	if code := dispatch([]string{"cli", "--platform", "docker", "--container", "solmq-conn"}, q); code != 3 {
		t.Errorf("exit=%d, want the session status 3", code)
	}
}

// TestCliSessionThatCouldNotStartIsAnError separates the two: a session that
// never began is this tool failing, which is exit 1 and names the instance.
func TestCliSessionThatCouldNotStartIsAnError(t *testing.T) {
	withPrompter(t, true, "")
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
			code: 2,
		},
		{
			name: "an unsafe instance name",
			args: []string{"cli", "--platform", "kubernetes", "--pod", "pod-a;rm -rf /"},
			want: []string{"unsafe character"},
			code: 2,
		},
	} {
		t.Run(c.name, func(t *testing.T) {
			withPrompter(t, true, "")
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
	withPrompter(t, true, "")
	dir := t.TempDir()
	write(t, dir, "env.yaml", sharedEnv+kubeEnvForLogs)
	envPath := filepath.Join(dir, "env.yaml")

	q := &attachRunner{resp: []queuedResp{{unsortedPodsJSON, nil}}}
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
	withPrompter(t, true, "")
	dir := t.TempDir()
	write(t, dir, "env.yaml", sharedEnv+kubeEnvForLogs)

	q := &attachRunner{resp: []queuedResp{{unsortedPodsJSON, nil}}}
	if code := dispatch([]string{"cli", "-e", filepath.Join(dir, "env.yaml"), "--pod", "1"}, q); code != 0 {
		t.Fatalf("exit=%d, want 0", code)
	}
	if len(q.calls) != 2 {
		t.Fatalf("want 2 calls (get pods, the attach), got %d: %+v", len(q.calls), q.calls)
	}
	// pod-b: the middle of the sorted list, not the middle of the JSON.
	if !containsToken(q.calls[1].argv, "pod-b") {
		t.Errorf("index 1 should select pod-b from the sorted list, got %v", q.calls[1].argv)
	}
}

// TestCliPreflightFailureOpensNothing keeps cli on the same contract as status
// and logs: an engine that cannot be reached is reported once, up front, and no
// session is attempted.
func TestCliPreflightFailureOpensNothing(t *testing.T) {
	withPrompter(t, true, "")
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

// ---- docker: created stores ---------------------------------------------------

// createdStoreDockerSpec writes env.yaml with a docker section and one created
// truststore, ts, built from ca.pem beside it with its password in
// SOLMQ_TEST_TS_PW, and tlsWF selecting it.
func createdStoreDockerSpec(t *testing.T, dir string) {
	t.Helper()
	write(t, dir, "ca.pem", string(certfixture.PEM(certfixture.CA(t, "Root"))))
	write(t, dir, "env.yaml", sharedEnv+`
tls:
  stores:
    ts:
      password-env: SOLMQ_TEST_TS_PW
      create:
        certs:
          - ./ca.pem
`+dockerEnv)
	write(t, dir, "10.yaml", tlsWF)
}

// envValue is the value of key in a child environment, as KEY=VALUE pairs.
func envValue(env []string, key string) (string, bool) {
	for _, kv := range env {
		if k, v, ok := strings.Cut(kv, "="); ok && k == key {
			return v, true
		}
	}
	return "", false
}

// imageStartReply is what `image inspect` prints for an image that starts the
// way storescript records: its entrypoint, then its command, as JSON lines.
func imageStartReply(t *testing.T) string {
	t.Helper()
	ep, err := json.Marshal(storescript.ImageEntrypoint)
	if err != nil {
		t.Fatal(err)
	}
	cmd, err := json.Marshal(storescript.ImageCmd)
	if err != nil {
		t.Fatal(err)
	}
	return string(ep) + "\n" + string(cmd) + "\n"
}

// A docker deploy carrying a created store builds it and hands it to compose
// in the child environment -- base64-encoded under its secret's name, with the
// digest the stores label is filled from -- never in argv. The image is
// inspected first and starts the way the storescript hands over to.
func TestDeployDockerCarriesCreatedStores(t *testing.T) {
	f := &queueRunner{resp: []queuedResp{{out: "ok"}, {out: imageStartReply(t)}}}
	dir := t.TempDir()
	createdStoreDockerSpec(t, dir)
	t.Setenv("SOLMQ_TEST_TS_PW", "store-pass-123")
	if code := dispatch([]string{"deploy", "--platform", "docker", "-e", filepath.Join(dir, "env.yaml")}, f); code != 0 {
		t.Fatalf("exit=%d", code)
	}
	if len(f.calls) != 4 || f.calls[0].argv[1] != "info" || !reflect.DeepEqual(f.calls[1].argv[:3], []string{"docker", "image", "inspect"}) ||
		f.calls[2].argv[1] != "ps" || f.calls[3].argv[len(f.calls[3].argv)-2] != "up" {
		t.Fatalf("calls = %+v, want info, image inspect, ps, compose up", f.calls)
	}
	up := f.calls[3]
	enc, ok := envValue(up.env, "_GEN_TLS_TS_STORE")
	if !ok {
		t.Fatalf("compose up env carries no _GEN_TLS_TS_STORE: %v", up.env)
	}
	raw, err := base64.StdEncoding.DecodeString(enc)
	if err != nil {
		t.Fatal(err)
	}
	if certs, err := pkcs12.DecodeTrustStore(raw, "store-pass-123"); err != nil || len(certs) != 1 {
		t.Errorf("the environment does not carry the built truststore: %d certificates, err %v", len(certs), err)
	}
	if digest, ok := envValue(up.env, "_GEN_TLS_STORES_SHA256"); !ok || len(digest) != 64 {
		t.Errorf("stores digest = %q, want 64 hex characters", digest)
	}
	for _, a := range up.argv {
		if strings.Contains(a, enc[:16]) {
			t.Errorf("store bytes reached argv: %q", a)
		}
	}
}

// An image the engine does not have yet is pulled and then inspected, so the
// check always compares something; one it still cannot describe after the
// pull stops the deploy before anything is written.
func TestDeployDockerPullsAnImageItCannotInspect(t *testing.T) {
	noImage := queuedResp{out: "Error: No such image: img:1", err: errors.New("exit status 1")}
	for _, c := range []struct {
		name  string
		resp  []queuedResp
		code  int
		calls []string // the verb after "docker" in each call
	}{
		{"pulled, then it starts as expected", []queuedResp{{out: "ok"}, noImage, {out: "pulled"}, {out: imageStartReply(t)}}, 0, []string{"info", "image", "image", "image", "ps", "compose"}},
		{"pulled, still not described", []queuedResp{{out: "ok"}, noImage, {out: "pulled"}, noImage}, 1, []string{"info", "image", "image", "image"}},
		{"the pull fails", []queuedResp{{out: "ok"}, noImage, {out: "denied", err: errors.New("exit status 1")}}, 1, []string{"info", "image", "image"}},
	} {
		t.Run(c.name, func(t *testing.T) {
			q := &queueRunner{resp: c.resp}
			dir := t.TempDir()
			createdStoreDockerSpec(t, dir)
			t.Setenv("SOLMQ_TEST_TS_PW", "store-pass-123")
			var code int
			stderr := captureStderr(t, func() {
				code = dispatch([]string{"deploy", "--platform", "docker", "-e", filepath.Join(dir, "env.yaml")}, q)
			})
			if code != c.code {
				t.Fatalf("exit=%d, want %d:\n%s", code, c.code, stderr)
			}
			if len(q.calls) != len(c.calls) {
				t.Fatalf("calls = %+v, want %v", q.calls, c.calls)
			}
			for i, v := range c.calls {
				if q.calls[i].argv[1] != v {
					t.Errorf("call %d = %v, want docker %s ...", i, q.calls[i].argv, v)
				}
			}
			if len(q.calls) > 2 && !reflect.DeepEqual(q.calls[2].argv, []string{"docker", "image", "pull", "img:1"}) {
				t.Errorf("call 2 = %v, want the pull", q.calls[2].argv)
			}
		})
	}
}

// An image the engine reports starting any other way than the command the
// storescript hands over to fails the deploy before a compose file is written
// or anything is started.
func TestDeployDockerRefusesAnImageStartingOtherwise(t *testing.T) {
	q := &queueRunner{resp: []queuedResp{{out: "ok"}, {out: "[\"java\",\"-jar\",\"/app.jar\"]\n[]"}}}
	dir := t.TempDir()
	createdStoreDockerSpec(t, dir)
	t.Setenv("SOLMQ_TEST_TS_PW", "store-pass-123")
	var code int
	stderr := captureStderr(t, func() {
		code = dispatch([]string{"deploy", "--platform", "docker", "-e", filepath.Join(dir, "env.yaml")}, q)
	})
	if code == 0 || !strings.Contains(stderr, `the image img:1 starts with entrypoint ["java" "-jar" "/app.jar"] and command []`) {
		t.Fatalf("exit=%d, stderr:\n%s", code, stderr)
	}
	if len(q.calls) != 2 {
		t.Errorf("calls = %+v, want only info and image inspect", q.calls)
	}
	if _, err := os.Stat(filepath.Join(dir, "docker-compose.yml")); !os.IsNotExist(err) {
		t.Errorf("no compose file may be written: %v", err)
	}
}

// remove tears down with placeholders for the created stores' secrets: their
// sources may be gone, since down never reads them. The store's password is a
// credential the compose file declares like any other, so it is resolved as
// every credential is.
func TestRemoveDockerNeedsNoCreatedStoreSources(t *testing.T) {
	f := useFakeRunner(t)
	withPrompter(t, true, "y\n")
	dir := t.TempDir()
	createdStoreDockerSpec(t, dir)
	t.Setenv("SOLMQ_TEST_TS_PW", "store-pass-123")
	if err := os.Remove(filepath.Join(dir, "ca.pem")); err != nil {
		t.Fatal(err)
	}
	if code := dispatch([]string{"remove", "--platform", "docker", "-e", filepath.Join(dir, "env.yaml")}, f); code != 0 {
		t.Fatalf("exit=%d", code)
	}
	down := f.calls[len(f.calls)-1]
	if v, _ := envValue(down.env, "_GEN_TLS_TS_STORE"); v != "-" {
		t.Errorf("_GEN_TLS_TS_STORE = %q, want the placeholder", v)
	}
	if v, _ := envValue(down.env, "_GEN_TLS_STORES_SHA256"); v != "-" {
		t.Errorf("_GEN_TLS_STORES_SHA256 = %q, want the placeholder", v)
	}
}

// podman remove deletes a created store's secret even when the unit that
// mounted it is already gone -- a deploy that failed before writing it, or a
// unit deleted by hand -- and needs no store source to do so.
func TestRemovePodmanDeletesACreatedStoreSecretWithoutTheUnit(t *testing.T) {
	f := useFakeRunner(t)
	podmanQuadletHome(t) // no unit written: the fallback is what is tested
	withPrompter(t, true, "y\n")
	dir := t.TempDir()
	write(t, dir, "env.yaml", sharedEnv+`
tls:
  stores:
    ts:
      password-env: SOLMQ_TEST_TS_PW
      create:
        certs:
          - ./gone.pem
`+podmanEnv())
	write(t, dir, "10.yaml", tlsWF)
	if code := dispatch([]string{"remove", "--platform", "podman", "-e", filepath.Join(dir, "env.yaml")}, f); code != 0 {
		t.Fatalf("exit=%d", code)
	}
	rm := f.calls[len(f.calls)-1].argv
	if !slices.Contains(rm, "solmq-conn-tls-ts") || rm[1] != "secret" || rm[2] != "rm" {
		t.Errorf("last call = %v, want a secret rm naming solmq-conn-tls-ts", rm)
	}
}

// A docker deploy whose existing store is missing stops before anything runs
// -- docker would otherwise create an empty directory in its place -- while
// remove, which mounts nothing, still tears down.
func TestDeployDockerRefusesAMissingStoreFile(t *testing.T) {
	f := useFakeRunner(t)
	dir := t.TempDir()
	write(t, dir, "env.yaml", sharedEnv+"\ntls:\n  stores:\n    ts:\n      password: ts\n      existing:\n        pkcs12: ./certs/missing.p12\n"+dockerEnv)
	write(t, dir, "10.yaml", tlsWF)
	var code int
	stderr := captureStderr(t, func() {
		code = dispatch([]string{"deploy", "--platform", "docker", "-e", filepath.Join(dir, "env.yaml")}, f)
	})
	if code != 1 || !strings.Contains(stderr, "docker mounts tls.stores.ts.existing.pkcs12 "+filepath.Join(dir, "certs", "missing.p12")+", which does not exist") {
		t.Fatalf("exit=%d, stderr:\n%s", code, stderr)
	}
	if len(f.calls) != 0 {
		t.Errorf("nothing may run before the store is there, got %+v", f.calls)
	}
	withPrompter(t, true, "y\n")
	if code := dispatch([]string{"remove", "--platform", "docker", "-e", filepath.Join(dir, "env.yaml")}, useFakeRunner(t)); code != 0 {
		t.Errorf("remove must not need the store: exit=%d", code)
	}
}
