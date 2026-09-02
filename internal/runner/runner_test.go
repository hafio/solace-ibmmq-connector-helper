package runner

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/spec"
	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/validate"
)

// fakeRunner records every invocation so tests can assert the exact argv/stdin
// crossing the exec boundary without starting a process.
type fakeRunner struct {
	calls     []call
	err       error         // returned by every Run when errByCall has no entry (nil = success)
	errByCall map[int]error // per-call-index override, so a later call can fail while earlier ones succeed
	out       string        // returned by every Run when outByCall has no entry ("" falls back to "out")
	outByCall map[int]string
}

type call struct {
	argv  []string
	stdin string
	env   []string
}

func (f *fakeRunner) Run(c Cmd) (string, error) {
	idx := len(f.calls)
	f.calls = append(f.calls, call{argv: c.Argv, stdin: c.Stdin, env: c.Env})
	out, ok := f.outByCall[idx]
	if !ok {
		out = f.out
		if out == "" {
			out = "out"
		}
	}
	if e, ok := f.errByCall[idx]; ok {
		return out, e
	}
	return out, f.err
}

// ---- OS.Run (the real exec.Command boundary) --------------------------------

// helperProcessArgv builds an OS.Run argv slice that re-execs this test binary
// under go test's own harness with -test.run=TestHelperProcess, so
// TestHelperProcess below runs as the child process instead of the actual test
// suite. program is the value used for argv[0] -- tests pass either os.Args[0]
// or an absolute path from os.Executable so both forms are exercised.
func helperProcessArgv(program, mode string) []string {
	return []string{program, "-test.run=TestHelperProcess", "--", mode}
}

// TestHelperProcess is not a real test: it returns immediately unless
// GO_WANT_HELPER_PROCESS is set. OS.Run leaves cmd.Env nil, so the child
// inherits the parent's environment and sees the variable the tests below set
// with t.Setenv, letting them drive a re-exec of this binary as the child
// process without depending on any external command.
func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	args := os.Args
	for len(args) > 0 {
		if args[0] == "--" {
			args = args[1:]
			break
		}
		args = args[1:]
	}
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "TestHelperProcess: no mode given")
		os.Exit(2)
	}
	switch args[0] {
	case "stdin":
		b, err := io.ReadAll(os.Stdin)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		os.Stdout.Write(b)
	case "both":
		fmt.Fprintln(os.Stdout, "stdout-line")
		fmt.Fprintln(os.Stderr, "stderr-line")
	case "fail":
		fmt.Fprintln(os.Stdout, "before-exit")
		os.Exit(3)
	case "drip":
		// Print immediately, then outlive any reasonable test: a reader that
		// sees the line has seen it while the process was still running, which
		// is what separates Stream from Run.
		fmt.Fprintln(os.Stdout, "drip-line")
		time.Sleep(10 * time.Minute)
	case "env":
		fmt.Fprintf(os.Stdout, "AMBIENT_ONLY=%s OVERRIDE=%s EXTRA_ONLY=%s\n",
			os.Getenv("RUNNER_TEST_AMBIENT_ONLY"),
			os.Getenv("RUNNER_TEST_OVERRIDE"),
			os.Getenv("RUNNER_TEST_EXTRA_ONLY"))
	case "stdiokind":
		// Reports whether each standard file the parent handed over is a plain
		// regular file. If os/exec had interposed a pipe -- which it does for
		// any io.Writer that is not an *os.File -- these would read as pipes
		// instead, and that substitution is exactly what Attach must never
		// allow: a pipe is not a terminal, so the engine would refuse the tty.
		fmt.Fprintf(os.Stdout, "in=%v out=%v err=%v\n",
			regularFile(os.Stdin), regularFile(os.Stdout), regularFile(os.Stderr))
	default:
		fmt.Fprintf(os.Stderr, "TestHelperProcess: unknown mode %q\n", args[0])
		os.Exit(2)
	}
	os.Exit(0)
}

func TestOSRunWiresStdinToChild(t *testing.T) {
	t.Setenv("GO_WANT_HELPER_PROCESS", "1")
	out, err := (OS{}).Run(Cmd{Argv: helperProcessArgv(os.Args[0], "stdin"), Stdin: "hello-stdin\n"})
	if err != nil {
		t.Fatalf("Run returned error: %v (output %q)", err, out)
	}
	if out != "hello-stdin\n" {
		t.Errorf("stdin content not visible to child: output = %q", out)
	}
}

func TestOSRunCombinesStdoutAndStderr(t *testing.T) {
	t.Setenv("GO_WANT_HELPER_PROCESS", "1")
	out, err := (OS{}).Run(Cmd{Argv: helperProcessArgv(os.Args[0], "both")})
	if err != nil {
		t.Fatalf("Run returned error: %v (output %q)", err, out)
	}
	if !strings.Contains(out, "stdout-line") {
		t.Errorf("combined output missing stdout write: %q", out)
	}
	if !strings.Contains(out, "stderr-line") {
		t.Errorf("combined output missing stderr write: %q", out)
	}
}

func TestOSRunNonZeroExitReturnsErrorWithOutput(t *testing.T) {
	t.Setenv("GO_WANT_HELPER_PROCESS", "1")
	out, err := (OS{}).Run(Cmd{Argv: helperProcessArgv(os.Args[0], "fail")})
	if err == nil {
		t.Fatalf("non-zero child exit must return a non-nil error, output = %q", out)
	}
	if !strings.Contains(out, "before-exit") {
		t.Errorf("output written before the failing exit must still be captured: %q", out)
	}
}

// TestOSRunEnvReachesChildAndAmbientInherited pins the two guarantees in the Cmd.Env
// doc comment: a supplied entry reaches the child, an ambient variable the child
// process would otherwise see (via os.Environ()) is still inherited when Env carries
// unrelated entries, and a supplied entry wins over an ambient one of the same name.
func TestOSRunEnvReachesChildAndAmbientInherited(t *testing.T) {
	t.Setenv("GO_WANT_HELPER_PROCESS", "1")
	t.Setenv("RUNNER_TEST_AMBIENT_ONLY", "ambient-only-value")
	t.Setenv("RUNNER_TEST_OVERRIDE", "ambient-should-lose")
	out, err := (OS{}).Run(Cmd{
		Argv: helperProcessArgv(os.Args[0], "env"),
		Env: []string{
			"RUNNER_TEST_OVERRIDE=override-wins",
			"RUNNER_TEST_EXTRA_ONLY=extra-only-value",
		},
	})
	if err != nil {
		t.Fatalf("Run returned error: %v (output %q)", err, out)
	}
	want := "AMBIENT_ONLY=ambient-only-value OVERRIDE=override-wins EXTRA_ONLY=extra-only-value\n"
	if out != want {
		t.Errorf("output = %q, want %q", out, want)
	}
}

// TestOSRunAcceptsAbsolutePathArgv0 pins the trust decision at the nolint on
// OS.Run's exec.Command call -- argv[0] can come from an env.yaml command
// token, which validate.SafeToken allows to name any binary (letters, digits,
// and / . : - _ among the allowed chars), so an absolute path must resolve via
// exec.LookPath and run exactly like a bare command name does.
func TestOSRunAcceptsAbsolutePathArgv0(t *testing.T) {
	exe, err := os.Executable()
	if err != nil {
		t.Fatalf("os.Executable: %v", err)
	}
	if !filepath.IsAbs(exe) {
		t.Fatalf("os.Executable() returned a non-absolute path %q", exe)
	}
	t.Setenv("GO_WANT_HELPER_PROCESS", "1")
	out, err := (OS{}).Run(Cmd{Argv: helperProcessArgv(exe, "stdin"), Stdin: "abs-argv0-ok\n"})
	if err != nil {
		t.Fatalf("Run with absolute argv[0] returned error: %v (output %q)", err, out)
	}
	if out != "abs-argv0-ok\n" {
		t.Errorf("output = %q", out)
	}
}

func TestParseCommand(t *testing.T) {
	cases := []struct {
		in   string
		want []string
		ok   bool
	}{
		{"kubectl", []string{"kubectl"}, true},
		{"kubectl --context prod -n solace", []string{"kubectl", "--context", "prod", "-n", "solace"}, true},
		{"  oc   ", []string{"oc"}, true},
		{"", nil, false},
		{"   ", nil, false},
		{"kubectl; rm -rf /", nil, false},          // ';' is unsafe
		{"kubectl $(evil)", nil, false},            // '$' and '(' unsafe
		{"kubectl `id`", nil, false},               // backtick unsafe
		{`kubectl --kubeconfig "a b"`, nil, false}, // quote+space unsafe
		{"curl", nil, false},                       // not on the kubernetes allowlist
		{"/tmp/evil", nil, false},                  // path, not a bare name
	}
	for _, c := range cases {
		got, err := ParseCommand(validate.PlatformKubernetes, c.in, nil)
		if c.ok {
			if err != nil {
				t.Errorf("ParseCommand(%q) unexpected error: %v", c.in, err)
				continue
			}
			if !reflect.DeepEqual(got, c.want) {
				t.Errorf("ParseCommand(%q) = %v, want %v", c.in, got, c.want)
			}
			continue
		}
		if err == nil {
			t.Errorf("ParseCommand(%q) should have failed, got %v", c.in, got)
		}
	}
}

// TestParseCommandExtraAllowed pins the escape hatch: a chained binary (e.g.
// `sudo podman`) is rejected on the platform allowlist alone but accepted once
// its argv[0] is named in extraAllowed, and the platform's own binary still
// follows in argv unmodified.
func TestParseCommandExtraAllowed(t *testing.T) {
	if _, err := ParseCommand(validate.PlatformPodman, "sudo podman", nil); err == nil {
		t.Fatal("sudo podman must be rejected without --allow-command sudo")
	}
	got, err := ParseCommand(validate.PlatformPodman, "sudo podman", []string{"sudo"})
	if err != nil {
		t.Fatalf("ParseCommand with extraAllowed=[sudo] unexpected error: %v", err)
	}
	want := []string{"sudo", "podman"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("ParseCommand(%q) = %v, want %v", "sudo podman", got, want)
	}
}

func TestKubernetesDeployApplyOnStdin(t *testing.T) {
	f := &fakeRunner{}
	manifest := "apiVersion: v1\nkind: Namespace\n"
	if _, err := Kubernetes(f, "kubectl --context prod", ActionDeploy, manifest, nil); err != nil {
		t.Fatal(err)
	}
	if len(f.calls) != 1 {
		t.Fatalf("want 1 call, got %d", len(f.calls))
	}
	wantArgv := []string{"kubectl", "--context", "prod", "apply", "-f", "-"}
	if !reflect.DeepEqual(f.calls[0].argv, wantArgv) {
		t.Errorf("argv = %v, want %v", f.calls[0].argv, wantArgv)
	}
	if f.calls[0].stdin != manifest {
		t.Errorf("manifest must be on stdin, got %q", f.calls[0].stdin)
	}
}

func TestKubernetesRemoveUsesDeleteVerb(t *testing.T) {
	f := &fakeRunner{}
	if _, err := Kubernetes(f, "oc", ActionRemove, "x", nil); err != nil {
		t.Fatal(err)
	}
	wantArgv := []string{"oc", "delete", "-f", "-"}
	if !reflect.DeepEqual(f.calls[0].argv, wantArgv) {
		t.Errorf("argv = %v, want %v", f.calls[0].argv, wantArgv)
	}
}

func TestKubernetesRejectsUnsafeCommand(t *testing.T) {
	f := &fakeRunner{}
	if _, err := Kubernetes(f, "kubectl; rm -rf /", ActionDeploy, "x", nil); err == nil {
		t.Fatal("unsafe command must be rejected")
	}
	if len(f.calls) != 0 {
		t.Fatal("nothing must be executed when the command is rejected")
	}
}

func TestKubernetesUnknownAction(t *testing.T) {
	f := &fakeRunner{}
	if _, err := Kubernetes(f, "kubectl", "restart", "x", nil); err == nil {
		t.Fatal("unknown action must be rejected")
	}
	if len(f.calls) != 0 {
		t.Fatal("nothing must run for an unknown action")
	}
}

func TestDockerUpAndDown(t *testing.T) {
	f := &fakeRunner{}
	env := []string{"MQ_CONN_1_USER=admin", "MQ_CONN_1_PASSWORD=s3cr3t"}
	if _, err := Docker(f, "docker", ActionDeploy, "/tmp/docker-compose.yml", env, nil); err != nil {
		t.Fatal(err)
	}
	wantUp := []string{"docker", "compose", "-f", "/tmp/docker-compose.yml", "up", "-d"}
	if !reflect.DeepEqual(f.calls[0].argv, wantUp) {
		t.Errorf("up argv = %v, want %v", f.calls[0].argv, wantUp)
	}
	if !reflect.DeepEqual(f.calls[0].env, env) {
		t.Errorf("up env = %v, want %v (credential values must cross via Cmd.Env, never argv)", f.calls[0].env, env)
	}
	if _, err := Docker(f, "docker", ActionRemove, "/tmp/docker-compose.yml", env, nil); err != nil {
		t.Fatal(err)
	}
	wantDown := []string{"docker", "compose", "-f", "/tmp/docker-compose.yml", "down"}
	if !reflect.DeepEqual(f.calls[1].argv, wantDown) {
		t.Errorf("down argv = %v, want %v", f.calls[1].argv, wantDown)
	}
	if !reflect.DeepEqual(f.calls[1].env, env) {
		t.Errorf("down env = %v, want %v (down must resolve the same secret declarations up did)", f.calls[1].env, env)
	}
}

func TestDockerRejectsUnsafeCommand(t *testing.T) {
	f := &fakeRunner{}
	if _, err := Docker(f, "docker `id`", ActionDeploy, "x", nil, nil); err == nil {
		t.Fatal("unsafe command must be rejected")
	}
	if len(f.calls) != 0 {
		t.Fatal("nothing must run when the command is rejected")
	}
}

func TestResolveQuadletScope(t *testing.T) {
	sys, err := ResolveQuadletScope(spec.QuadletScopeSystem, "")
	if err != nil {
		t.Fatal(err)
	}
	if sys.UserMode || sys.Dir != quadletSystem {
		t.Errorf("system scope = %+v", sys)
	}
	usr, err := ResolveQuadletScope(spec.QuadletScopeUser, "")
	if err != nil {
		t.Fatal(err)
	}
	if !usr.UserMode || !strings.HasSuffix(filepath.ToSlash(usr.Dir), quadletUserSub) {
		t.Errorf("user scope = %+v", usr)
	}
	over, err := ResolveQuadletScope(spec.QuadletScopeUser, "/custom/dir")
	if err != nil {
		t.Fatal(err)
	}
	if over.Dir != "/custom/dir" {
		t.Errorf("dir override ignored: %+v", over)
	}
	if _, err := ResolveQuadletScope("bogus", ""); err == nil {
		t.Fatal("unknown scope must be rejected")
	}
}

func TestPodmanDeployReloadThenStart(t *testing.T) {
	f := &fakeRunner{}
	sc := QuadletScope{Dir: t.TempDir(), UserMode: true}
	if _, err := PodmanDeploy(f, sc, []string{"solmq-connector.service"}); err != nil {
		t.Fatal(err)
	}
	want := [][]string{
		{"systemctl", "--user", "daemon-reload"},
		{"systemctl", "--user", "start", "solmq-connector.service"},
	}
	for i, w := range want {
		if !reflect.DeepEqual(f.calls[i].argv, w) {
			t.Errorf("call %d argv = %v, want %v", i, f.calls[i].argv, w)
		}
	}
}

func TestPodmanDeploySystemModeNoUserFlag(t *testing.T) {
	f := &fakeRunner{}
	sc := QuadletScope{Dir: t.TempDir(), UserMode: false}
	if _, err := PodmanDeploy(f, sc, []string{"a.service"}); err != nil {
		t.Fatal(err)
	}
	if reflect.DeepEqual(f.calls[0].argv, []string{"systemctl", "--user", "daemon-reload"}) {
		t.Error("system mode must not pass --user")
	}
	if !reflect.DeepEqual(f.calls[0].argv, []string{"systemctl", "daemon-reload"}) {
		t.Errorf("argv = %v", f.calls[0].argv)
	}
}

func TestPodmanRemoveStopsRemovesReloads(t *testing.T) {
	f := &fakeRunner{}
	dir := t.TempDir()
	unit := "solmq-connector.container"
	if err := os.WriteFile(filepath.Join(dir, unit), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	sc := QuadletScope{Dir: dir, UserMode: false}
	if _, err := PodmanRemove(f, sc, []string{"solmq-connector.service"}, []string{unit}); err != nil {
		t.Fatal(err)
	}
	want := [][]string{
		{"systemctl", "stop", "solmq-connector.service"},
		{"systemctl", "daemon-reload"},
	}
	for i, w := range want {
		if !reflect.DeepEqual(f.calls[i].argv, w) {
			t.Errorf("call %d argv = %v, want %v", i, f.calls[i].argv, w)
		}
	}
	if _, err := os.Stat(filepath.Join(dir, unit)); !os.IsNotExist(err) {
		t.Error("unit file must be removed")
	}
}

func TestPodmanDeployStartFailureIsReported(t *testing.T) {
	// daemon-reload (call 0) succeeds so the per-service start (call 1) is actually
	// reached; injecting the failure only on call 1 exercises the start branch.
	f := &fakeRunner{errByCall: map[int]error{1: fmt.Errorf("boom")}}
	sc := QuadletScope{Dir: t.TempDir(), UserMode: true}
	_, err := PodmanDeploy(f, sc, []string{"a.service"})
	if err == nil {
		t.Fatal("a systemctl start failure must surface")
	}
	if !strings.Contains(err.Error(), "start a.service") {
		t.Errorf("error should name the failed start, got %v", err)
	}
	if len(f.calls) != 2 {
		t.Fatalf("want daemon-reload + start = 2 calls, got %d", len(f.calls))
	}
}

func TestDockerUnknownAction(t *testing.T) {
	f := &fakeRunner{}
	if _, err := Docker(f, "docker", "restart", "x", nil, nil); err == nil {
		t.Fatal("unknown action must be rejected")
	}
	if len(f.calls) != 0 {
		t.Fatal("nothing must run for an unknown action")
	}
}

func TestPodmanRemoveStopFailureIsReported(t *testing.T) {
	// The first call PodmanRemove makes is the per-service stop, so a runner that
	// fails every call exercises the stop-failure branch.
	f := &fakeRunner{err: fmt.Errorf("boom")}
	sc := QuadletScope{Dir: t.TempDir(), UserMode: false}
	_, err := PodmanRemove(f, sc, []string{"solmq-connector.service"}, []string{"solmq-connector.container"})
	if err == nil {
		t.Fatal("a systemctl stop failure must surface")
	}
	if !strings.Contains(err.Error(), "stop solmq-connector.service") {
		t.Errorf("error should name the failed stop, got %v", err)
	}
}

// ---- PodmanSecretCreate / PodmanSecretRemove --------------------------------

func TestPodmanSecretCreateRemovesThenCreatesValueOnStdin(t *testing.T) {
	f := &fakeRunner{}
	const name = "mq-conn-1-password"
	const value = "s3cr3t"
	if _, err := PodmanSecretCreate(f, "podman", name, value, nil); err != nil {
		t.Fatal(err)
	}
	if len(f.calls) != 2 {
		t.Fatalf("want rm then create = 2 calls, got %d", len(f.calls))
	}
	wantRm := []string{"podman", "secret", "rm", "--ignore", name}
	if !reflect.DeepEqual(f.calls[0].argv, wantRm) {
		t.Errorf("rm argv = %v, want %v", f.calls[0].argv, wantRm)
	}
	if f.calls[0].stdin != "" {
		t.Errorf("rm must not carry stdin, got %q", f.calls[0].stdin)
	}
	wantCreate := []string{"podman", "secret", "create", name, "-"}
	if !reflect.DeepEqual(f.calls[1].argv, wantCreate) {
		t.Errorf("create argv = %v, want %v", f.calls[1].argv, wantCreate)
	}
	if f.calls[1].stdin != value {
		t.Errorf("create stdin = %q, want %q", f.calls[1].stdin, value)
	}
	for i, c := range f.calls {
		for _, a := range c.argv {
			if strings.Contains(a, value) {
				t.Errorf("call %d: secret value must never appear in argv, found in %q", i, a)
			}
		}
	}
}

func TestPodmanSecretCreateSkipsCreateWhenRmFails(t *testing.T) {
	f := &fakeRunner{err: fmt.Errorf("boom")}
	_, err := PodmanSecretCreate(f, "podman", "mq-conn-1-password", "s3cr3t", nil)
	if err == nil {
		t.Fatal("a failed rm must surface as an error")
	}
	if !strings.Contains(err.Error(), "podman secret rm mq-conn-1-password") {
		t.Errorf("error should name the failed rm, got %v", err)
	}
	if len(f.calls) != 1 {
		t.Fatalf("create must not run when rm fails, got %d calls", len(f.calls))
	}
}

func TestPodmanSecretCreateReportsCreateFailure(t *testing.T) {
	f := &fakeRunner{errByCall: map[int]error{1: fmt.Errorf("boom")}}
	_, err := PodmanSecretCreate(f, "podman", "mq-conn-1-password", "s3cr3t", nil)
	if err == nil {
		t.Fatal("a failed create must surface as an error")
	}
	if !strings.Contains(err.Error(), "podman secret create mq-conn-1-password") {
		t.Errorf("error should name the failed create, got %v", err)
	}
	if len(f.calls) != 2 {
		t.Fatalf("want rm then create = 2 calls, got %d", len(f.calls))
	}
}

func TestPodmanSecretCreateRejectsUnsafeCommand(t *testing.T) {
	f := &fakeRunner{}
	if _, err := PodmanSecretCreate(f, "podman; rm -rf /", "name", "value", nil); err == nil {
		t.Fatal("unsafe command must be rejected")
	}
	if len(f.calls) != 0 {
		t.Fatal("nothing must run when the command is rejected")
	}
}

func TestPodmanSecretRemoveBatchesNames(t *testing.T) {
	f := &fakeRunner{}
	names := []string{"mq-conn-1-user", "mq-conn-1-password", "truststore-password"}
	if _, err := PodmanSecretRemove(f, "podman", names, nil); err != nil {
		t.Fatal(err)
	}
	if len(f.calls) != 1 {
		t.Fatalf("want a single batched call, got %d", len(f.calls))
	}
	want := append([]string{"podman", "secret", "rm", "--ignore"}, names...)
	if !reflect.DeepEqual(f.calls[0].argv, want) {
		t.Errorf("argv = %v, want %v", f.calls[0].argv, want)
	}
}

func TestPodmanSecretRemoveNoNamesIsNoop(t *testing.T) {
	f := &fakeRunner{}
	out, err := PodmanSecretRemove(f, "podman", nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if out != "" {
		t.Errorf("output = %q, want empty", out)
	}
	if len(f.calls) != 0 {
		t.Fatal("no names must not invoke the runner")
	}
}

func TestPodmanSecretRemoveRejectsUnsafeCommand(t *testing.T) {
	f := &fakeRunner{}
	if _, err := PodmanSecretRemove(f, "podman; rm -rf /", []string{"name"}, nil); err == nil {
		t.Fatal("unsafe command must be rejected")
	}
	if len(f.calls) != 0 {
		t.Fatal("nothing must run when the command is rejected")
	}
}

func TestPodmanSecretRemoveReportsFailure(t *testing.T) {
	f := &fakeRunner{err: fmt.Errorf("boom")}
	_, err := PodmanSecretRemove(f, "podman", []string{"name"}, nil)
	if err == nil {
		t.Fatal("a failed rm must surface as an error")
	}
	if !strings.Contains(err.Error(), "podman secret rm") {
		t.Errorf("error should name the failed operation, got %v", err)
	}
}

func TestWriteFileCreatesDirsAndMode(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "nested", "creds.env")
	if err := WriteFile(p, "USER=admin\n", 0o600); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(p)
	if err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != "USER=admin\n" {
		t.Errorf("content = %q", b)
	}
	// Unix perms are not faithfully reproduced on the windows-2025 CI runner.
	if runtime.GOOS != "windows" {
		if info.Mode().Perm() != 0o600 {
			t.Errorf("mode = %v, want 0600", info.Mode().Perm())
		}
	}
}

// TestWriteFileDoesNotTightenExistingFileMode pins that os.WriteFile does not
// chmod a file that already exists: writing with mode 0600 over a file already
// present with 0644 replaces the content but leaves the mode untouched. Callers
// must not rely on WriteFile to tighten an existing file's mode (matters for
// creds.env secrets rewritten by a later deploy).
func TestWriteFileDoesNotTightenExistingFileMode(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "creds.env")
	if err := os.WriteFile(p, []byte("OLD=1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := WriteFile(p, "NEW=2\n", 0o600); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != "NEW=2\n" {
		t.Errorf("content = %q, want replaced", b)
	}
	if runtime.GOOS != "windows" {
		info, err := os.Stat(p)
		if err != nil {
			t.Fatal(err)
		}
		if info.Mode().Perm() != 0o644 {
			t.Errorf("mode = %v, want unchanged 0644", info.Mode().Perm())
		}
	}
}

// TestWriteFileParentIsFileReturnsError covers the MkdirAll error branch: when
// the parent path already exists as a regular file, MkdirAll cannot create a
// directory there and the error must surface with the offending path.
func TestWriteFileParentIsFileReturnsError(t *testing.T) {
	dir := t.TempDir()
	blocker := filepath.Join(dir, "blocker")
	if err := os.WriteFile(blocker, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	p := filepath.Join(blocker, "creds.env")
	err := WriteFile(p, "x", 0o600)
	if err == nil {
		t.Fatal("expected error when the parent path is an existing file")
	}
	if !strings.Contains(err.Error(), blocker) {
		t.Errorf("error should name the offending path %q, got %v", blocker, err)
	}
}

// TestWriteFileTargetIsDirectoryReturnsError covers the os.WriteFile error
// branch: when the target path already exists as a directory, the write fails
// and the error must surface with the offending path.
func TestWriteFileTargetIsDirectoryReturnsError(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "creds.env")
	if err := os.Mkdir(target, 0o755); err != nil {
		t.Fatal(err)
	}
	err := WriteFile(target, "x", 0o600)
	if err == nil {
		t.Fatal("expected error when the target path is an existing directory")
	}
	if !strings.Contains(err.Error(), target) {
		t.Errorf("error should name the offending path %q, got %v", target, err)
	}
}

// TestOSRunWritesNothingToStderr pins that Run prints nothing of its own: the
// child's streams are captured into the returned combined output, and the
// caller (cmd/solmq-conn-util's report) decides what an operator sees. It
// reuses the TestHelperProcess re-exec trick already used above so no external
// binary is needed, and temporarily swaps os.Stderr for a pipe to capture
// anything that leaks past the buffer.
func TestOSRunWritesNothingToStderr(t *testing.T) {
	t.Setenv("GO_WANT_HELPER_PROCESS", "1")
	exe, err := os.Executable()
	if err != nil {
		t.Fatalf("os.Executable: %v", err)
	}
	origStderr := os.Stderr
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	os.Stderr = w

	argv := helperProcessArgv(exe, "stdin")
	out, runErr := (OS{}).Run(Cmd{Argv: argv, Stdin: "echo-test\n"})

	w.Close()
	os.Stderr = origStderr
	data, readErr := io.ReadAll(r)
	if readErr != nil {
		t.Fatalf("reading captured stderr: %v", readErr)
	}
	if runErr != nil {
		t.Fatalf("Run returned error: %v", runErr)
	}
	if got := string(data); got != "" {
		t.Errorf("Run wrote %q to stderr, want nothing", got)
	}
	// The child's own output still reaches the caller through the return value,
	// which is what the removed echo used to duplicate.
	if !strings.Contains(out, "echo-test") {
		t.Errorf("combined output = %q, want it to carry the child's output", out)
	}
}

// TestOSRunRejectsUnresolvableArgv0 pins that a binary LookPath cannot find is
// an error rather than a silent exec attempt (exec.Command would otherwise
// defer the same failure to Start).
func TestOSRunRejectsUnresolvableArgv0(t *testing.T) {
	_, err := (OS{}).Run(Cmd{Argv: []string{"solmq-conn-test-nonexistent-binary"}})
	if err == nil {
		t.Fatal("Run must fail when argv[0] cannot be resolved on PATH")
	}
}

// ---- Preflight ----------------------------------------------------------------

func TestPreflightKubernetesArgvDeployNoNamespace(t *testing.T) {
	f := &fakeRunner{}
	if _, err := Preflight(f, validate.PlatformKubernetes, "kubectl --context prod", ActionDeploy, "", nil); err != nil {
		t.Fatal(err)
	}
	want := []string{"kubectl", "--context", "prod", "auth", "can-i", "create", "deployment"}
	if !reflect.DeepEqual(f.calls[0].argv, want) {
		t.Errorf("argv = %v, want %v", f.calls[0].argv, want)
	}
}

func TestPreflightKubernetesArgvRemoveWithNamespace(t *testing.T) {
	f := &fakeRunner{}
	if _, err := Preflight(f, validate.PlatformKubernetes, "oc", ActionRemove, "solace", nil); err != nil {
		t.Fatal(err)
	}
	want := []string{"oc", "auth", "can-i", "delete", "deployment", "--namespace", "solace"}
	if !reflect.DeepEqual(f.calls[0].argv, want) {
		t.Errorf("argv = %v, want %v", f.calls[0].argv, want)
	}
}

func TestPreflightDockerArgvIsInfo(t *testing.T) {
	f := &fakeRunner{}
	if _, err := Preflight(f, validate.PlatformDocker, "docker --context foo", ActionDeploy, "", nil); err != nil {
		t.Fatal(err)
	}
	want := []string{"docker", "--context", "foo", "info"}
	if !reflect.DeepEqual(f.calls[0].argv, want) {
		t.Errorf("argv = %v, want %v", f.calls[0].argv, want)
	}
}

func TestPreflightPodmanArgvIsInfo(t *testing.T) {
	f := &fakeRunner{}
	if _, err := Preflight(f, validate.PlatformPodman, "podman", ActionRemove, "", nil); err != nil {
		t.Fatal(err)
	}
	want := []string{"podman", "info"}
	if !reflect.DeepEqual(f.calls[0].argv, want) {
		t.Errorf("argv = %v, want %v", f.calls[0].argv, want)
	}
}

func TestPreflightFailureWrapsPlatformHint(t *testing.T) {
	cases := []struct {
		platform string
		command  string
		hint     string
	}{
		{validate.PlatformKubernetes, "kubectl", "oc login"},
		{validate.PlatformDocker, "docker", "docker daemon"},
		{validate.PlatformPodman, "podman", "podman machine"},
	}
	for _, c := range cases {
		f := &fakeRunner{err: fmt.Errorf("connection refused")}
		_, err := Preflight(f, c.platform, c.command, ActionDeploy, "", nil)
		if err == nil {
			t.Fatalf("%s: preflight failure must surface as an error", c.platform)
		}
		if !strings.Contains(err.Error(), "preflight failed for "+c.platform) {
			t.Errorf("%s: error missing platform prefix: %v", c.platform, err)
		}
		if !strings.Contains(err.Error(), "connection refused") {
			t.Errorf("%s: error must preserve the underlying cause: %v", c.platform, err)
		}
		if !strings.Contains(err.Error(), c.hint) {
			t.Errorf("%s: error missing actionable hint (want to contain %q): %v", c.platform, c.hint, err)
		}
	}
}

func TestPreflightRejectsDisallowedBinaryBeforeRunning(t *testing.T) {
	f := &fakeRunner{}
	if _, err := Preflight(f, validate.PlatformKubernetes, "curl", ActionDeploy, "", nil); err == nil {
		t.Fatal("a binary outside the platform allowlist must be rejected")
	}
	if len(f.calls) != 0 {
		t.Fatal("the probe must never run when the command itself is rejected")
	}
}

func TestPreflightExtraAllowedThreadsThrough(t *testing.T) {
	f := &fakeRunner{}
	if _, err := Preflight(f, validate.PlatformPodman, "sudo podman", ActionDeploy, "", nil); err == nil {
		t.Fatal("sudo podman must be rejected without extraAllowed")
	}
	f = &fakeRunner{}
	if _, err := Preflight(f, validate.PlatformPodman, "sudo podman", ActionDeploy, "", []string{"sudo"}); err != nil {
		t.Fatalf("sudo podman with extraAllowed=[sudo] unexpected error: %v", err)
	}
	want := []string{"sudo", "podman", "info"}
	if !reflect.DeepEqual(f.calls[0].argv, want) {
		t.Errorf("argv = %v, want %v", f.calls[0].argv, want)
	}
}

func TestPreflightUnknownAction(t *testing.T) {
	f := &fakeRunner{}
	if _, err := Preflight(f, validate.PlatformKubernetes, "kubectl", "restart", "", nil); err == nil {
		t.Fatal("unknown action must be rejected")
	}
	if len(f.calls) != 0 {
		t.Fatal("nothing must run for an unknown action")
	}
}

// ---- the status verb's read-only queries -------------------------------------
//
// Every case here asserts the exact argv crossing the exec boundary, because
// that argv is the security boundary: no shell is involved, so what these
// slices say is precisely what runs. The tool-authored tokens (subcommands,
// flags, Go templates) are asserted verbatim for the same reason.

func TestKubernetesPodsJSONArgv(t *testing.T) {
	cases := []struct {
		name          string
		namespace     string
		selector      string
		names         []string
		allNamespaces bool
		want          []string
	}{
		{
			name: "by selector, no namespace", selector: "app=solmq-connector",
			want: []string{"kubectl", "get", "pods", "-l", "app=solmq-connector", "-o", "json"},
		},
		{
			name: "by selector in a namespace", namespace: "solace", selector: "app=solmq-connector",
			want: []string{"kubectl", "get", "pods", "-n", "solace", "-l", "app=solmq-connector", "-o", "json"},
		},
		{
			name: "explicit names", namespace: "solace", names: []string{"pod-a", "pod-b"},
			want: []string{"kubectl", "get", "pods", "pod-a", "pod-b", "-n", "solace", "-o", "json"},
		},
		{
			// --all is explicitly a cluster-wide search, so a namespace resolved
			// from env.yaml must not narrow it back down.
			name: "every namespace outranks a resolved namespace", namespace: "solace", allNamespaces: true,
			want: []string{"kubectl", "get", "pods", "--all-namespaces", "-o", "json"},
		},
	}
	for _, c := range cases {
		f := &fakeRunner{}
		if _, err := KubernetesPodsJSON(f, []string{"kubectl"}, c.namespace, c.selector, c.names, c.allNamespaces); err != nil {
			t.Fatalf("%s: %v", c.name, err)
		}
		if !reflect.DeepEqual(f.calls[0].argv, c.want) {
			t.Errorf("%s: argv = %v, want %v", c.name, f.calls[0].argv, c.want)
		}
	}
}

func TestKubernetesPodsJSONRunFailureWraps(t *testing.T) {
	f := &fakeRunner{err: fmt.Errorf("boom")}
	_, err := KubernetesPodsJSON(f, []string{"kubectl"}, "", "app=solmq-connector", nil, false)
	if err == nil {
		t.Fatal("a run failure must surface as an error")
	}
	if !strings.Contains(err.Error(), "listing pods") {
		t.Errorf("error should name the failed operation, got %v", err)
	}
}

func TestKubernetesGetJSONArgv(t *testing.T) {
	cases := []struct {
		name      string
		namespace string
		kind      string
		obj       string
		want      []string
	}{
		{"deployment in a namespace", "solace", "deployment", "solmq-connector",
			[]string{"kubectl", "get", "deployment", "solmq-connector", "-n", "solace", "-o", "json"}},
		{"no namespace", "", "service", "solmq-connector",
			[]string{"kubectl", "get", "service", "solmq-connector", "-o", "json"}},
		{"a referenced object", "solace", "configmap", "solmq-connector-config",
			[]string{"kubectl", "get", "configmap", "solmq-connector-config", "-n", "solace", "-o", "json"}},
	}
	for _, c := range cases {
		f := &fakeRunner{}
		if _, err := KubernetesGetJSON(f, []string{"kubectl"}, c.namespace, c.kind, c.obj); err != nil {
			t.Fatalf("%s: %v", c.name, err)
		}
		if !reflect.DeepEqual(f.calls[0].argv, c.want) {
			t.Errorf("%s: argv = %v, want %v", c.name, f.calls[0].argv, c.want)
		}
	}
}

// TestKubernetesGetJSONMissingObjectIsAnError covers the answer the components
// check is actually asking for: a missing object exits non-zero, and the caller
// turns that into "MISSING" rather than a failed run.
func TestKubernetesGetJSONMissingObjectIsAnError(t *testing.T) {
	f := &fakeRunner{err: fmt.Errorf("secrets \"nope\" not found")}
	_, err := KubernetesGetJSON(f, []string{"kubectl"}, "solace", "secret", "nope")
	if err == nil {
		t.Fatal("a missing object must surface as an error")
	}
	if !strings.Contains(err.Error(), "secret/nope") {
		t.Errorf("error should name the object, got %v", err)
	}
}

func TestKubernetesTopArgv(t *testing.T) {
	cases := []struct {
		name          string
		namespace     string
		selector      string
		names         []string
		allNamespaces bool
		want          []string
	}{
		{
			// --containers attributes the sample to one container rather than
			// summing the pod, so it is comparable with that container's limits.
			name: "by selector", namespace: "solace", selector: "app=solmq-connector",
			want: []string{"kubectl", "top", "pod", "-n", "solace", "-l", "app=solmq-connector", "--containers", "--no-headers"},
		},
		{
			name: "explicit names", names: []string{"pod-a"},
			want: []string{"kubectl", "top", "pod", "pod-a", "--containers", "--no-headers"},
		},
		{
			name: "every namespace", allNamespaces: true,
			want: []string{"kubectl", "top", "pod", "--all-namespaces", "--containers", "--no-headers"},
		},
	}
	for _, c := range cases {
		f := &fakeRunner{}
		if _, err := KubernetesTop(f, []string{"kubectl"}, c.namespace, c.selector, c.names, c.allNamespaces); err != nil {
			t.Fatalf("%s: %v", c.name, err)
		}
		if !reflect.DeepEqual(f.calls[0].argv, c.want) {
			t.Errorf("%s: argv = %v, want %v", c.name, f.calls[0].argv, c.want)
		}
	}
}

// TestKubernetesTopWithoutMetricsAPIWraps covers the cluster without
// metrics-server: kubectl fails, and the caller degrades to a note instead of
// dropping the whole report.
func TestKubernetesTopWithoutMetricsAPIWraps(t *testing.T) {
	f := &fakeRunner{err: fmt.Errorf("Metrics API not available")}
	_, err := KubernetesTop(f, []string{"kubectl"}, "", "app=x", nil, false)
	if err == nil {
		t.Fatal("an unavailable metrics API must surface as an error")
	}
	if !strings.Contains(err.Error(), "sampling pod resource usage") {
		t.Errorf("error should name the failed operation, got %v", err)
	}
}

// TestEngineInspectJSONArgv covers the one call that answers the whole
// container view on docker and podman: every target in a single invocation.
func TestEngineInspectJSONArgv(t *testing.T) {
	f := &fakeRunner{}
	if _, err := EngineInspectJSON(f, []string{"docker"}, []string{"a", "b", "c"}); err != nil {
		t.Fatal(err)
	}
	want := []string{"docker", "inspect", "a", "b", "c"}
	if !reflect.DeepEqual(f.calls[0].argv, want) {
		t.Errorf("argv = %v, want %v", f.calls[0].argv, want)
	}
	// A chained command keeps its own tokens ahead of the subcommand.
	f2 := &fakeRunner{}
	if _, err := EngineInspectJSON(f2, []string{"sudo", "podman"}, []string{"solmq"}); err != nil {
		t.Fatal(err)
	}
	want2 := []string{"sudo", "podman", "inspect", "solmq"}
	if !reflect.DeepEqual(f2.calls[0].argv, want2) {
		t.Errorf("argv = %v, want %v", f2.calls[0].argv, want2)
	}
}

func TestEngineInspectJSONNoNamesIsAnErrorAndRunsNothing(t *testing.T) {
	f := &fakeRunner{}
	if _, err := EngineInspectJSON(f, []string{"docker"}, nil); err == nil {
		t.Fatal("inspecting nothing must be an error")
	}
	if len(f.calls) != 0 {
		t.Errorf("nothing must run, got %v", f.calls)
	}
}

func TestEngineInspectJSONFailureNamesTheTargets(t *testing.T) {
	f := &fakeRunner{err: fmt.Errorf("no such object")}
	_, err := EngineInspectJSON(f, []string{"docker"}, []string{"a", "b"})
	if err == nil {
		t.Fatal("a run failure must surface as an error")
	}
	if !strings.Contains(err.Error(), "a, b") {
		t.Errorf("error should name the targets, got %v", err)
	}
}

func TestEngineImageInspectJSONArgv(t *testing.T) {
	f := &fakeRunner{}
	if _, err := EngineImageInspectJSON(f, []string{"podman"}, "solace/x:2.14.1"); err != nil {
		t.Fatal(err)
	}
	want := []string{"podman", "image", "inspect", "solace/x:2.14.1"}
	if !reflect.DeepEqual(f.calls[0].argv, want) {
		t.Errorf("argv = %v, want %v", f.calls[0].argv, want)
	}
}

// TestEngineStatsArgv pins the template as well as the argv: docker and podman
// disagree about the shape of their JSON stats, and these four template fields
// are what both render identically.
func TestEngineStatsArgv(t *testing.T) {
	f := &fakeRunner{}
	if _, err := EngineStats(f, []string{"docker"}, []string{"a", "b"}); err != nil {
		t.Fatal(err)
	}
	want := []string{"docker", "stats", "--no-stream", "--format",
		"{{.Name}}\t{{.CPUPerc}}\t{{.MemUsage}}\t{{.MemPerc}}", "a", "b"}
	if !reflect.DeepEqual(f.calls[0].argv, want) {
		t.Errorf("argv = %v, want %v", f.calls[0].argv, want)
	}
	// --no-stream is what keeps this one sample rather than a stream that never
	// returns; its absence would hang a status run.
	if !contains(f.calls[0].argv, "--no-stream") {
		t.Error("stats must not stream")
	}
}

func TestEngineListArgv(t *testing.T) {
	f := &fakeRunner{}
	if _, err := EngineList(f, []string{"podman"}); err != nil {
		t.Fatal(err)
	}
	want := []string{"podman", "ps", "--all", "--no-trunc", "--format", "{{.Names}}\t{{.Image}}"}
	if !reflect.DeepEqual(f.calls[0].argv, want) {
		t.Errorf("argv = %v, want %v", f.calls[0].argv, want)
	}
	// --all is deliberate: an instance that died is exactly what a search like
	// this is for.
	if !contains(f.calls[0].argv, "--all") {
		t.Error("the discovery list must include stopped containers")
	}
}

func contains(argv []string, want string) bool {
	for _, a := range argv {
		if a == want {
			return true
		}
	}
	return false
}

// TestSystemctlNRestarts covers the only truthful restart count under podman
// quadlet: systemd recreates the container, so the container's own counter
// reads 0 however many times the instance has died.
func TestSystemctlNRestarts(t *testing.T) {
	cases := []struct {
		name     string
		scope    QuadletScope
		out      string
		err      error
		want     int
		wantErr  bool
		wantArgv []string
	}{
		{
			name: "user scope", scope: QuadletScope{UserMode: true}, out: "7\n", want: 7,
			wantArgv: []string{"systemctl", "--user", "show", "solmq-connector.service", "-p", "NRestarts", "--value"},
		},
		{
			name: "system scope", scope: QuadletScope{}, out: "0\n", want: 0,
			wantArgv: []string{"systemctl", "show", "solmq-connector.service", "-p", "NRestarts", "--value"},
		},
		{
			// systemd answers an unknown unit with an empty property rather than a
			// failure, so the caller has to be able to fall back.
			name: "unknown unit answers nothing", scope: QuadletScope{}, out: "\n", wantErr: true,
			wantArgv: []string{"systemctl", "show", "solmq-connector.service", "-p", "NRestarts", "--value"},
		},
		{
			name: "no systemd on this host", scope: QuadletScope{}, err: fmt.Errorf("executable file not found"), wantErr: true,
			wantArgv: []string{"systemctl", "show", "solmq-connector.service", "-p", "NRestarts", "--value"},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			f := &fakeRunner{outByCall: map[int]string{0: c.out}, err: c.err}
			got, err := SystemctlNRestarts(f, c.scope, "solmq-connector.service")
			if (err != nil) != c.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, c.wantErr)
			}
			if !c.wantErr && got != c.want {
				t.Errorf("restarts = %d, want %d", got, c.want)
			}
			if !reflect.DeepEqual(f.calls[0].argv, c.wantArgv) {
				t.Errorf("argv = %v, want %v", f.calls[0].argv, c.wantArgv)
			}
		})
	}
}

// ---- ScriptInstalled -----------------------------------------------------------

// probePayload is the sh -c program ScriptInstalled sends: it reports presence
// as a marker on stdout and exits 0 either way, so a missing file cannot be
// confused with an unreachable target (see ScriptInstalled's doc comment).
const probePayload = "if [ -f /tmp/solmq-status.sh ]; then echo " + ScriptPresentMarker + "; else echo " + ScriptAbsentMarker + "; fi"

func TestScriptInstalledArgv(t *testing.T) {
	cases := []struct {
		name      string
		cmd       []string
		platform  string
		target    string
		namespace string
		want      []string
	}{
		{
			name: "kubernetes no namespace", cmd: []string{"kubectl"}, platform: validate.PlatformKubernetes,
			target: "pod-0",
			want:   []string{"kubectl", "exec", "pod-0", "-c", spec.ConnectorContainerName, "--", "sh", "-c", probePayload},
		},
		{
			name: "kubernetes with namespace", cmd: []string{"oc"}, platform: validate.PlatformKubernetes,
			target: "pod-0", namespace: "solace",
			want: []string{"oc", "exec", "pod-0", "-n", "solace", "-c", spec.ConnectorContainerName, "--", "sh", "-c", probePayload},
		},
		{
			name: "docker", cmd: []string{"docker"}, platform: validate.PlatformDocker,
			target: "solmq-connector",
			want:   []string{"docker", "exec", "solmq-connector", "sh", "-c", probePayload},
		},
		{
			name: "podman", cmd: []string{"podman"}, platform: validate.PlatformPodman,
			target: "solmq-connector",
			want:   []string{"podman", "exec", "solmq-connector", "sh", "-c", probePayload},
		},
	}
	for _, c := range cases {
		f := &fakeRunner{out: ScriptPresentMarker}
		if _, err := ScriptInstalled(f, c.cmd, c.platform, c.target, c.namespace, "/tmp/solmq-status.sh"); err != nil {
			t.Fatalf("%s: %v", c.name, err)
		}
		if !reflect.DeepEqual(f.calls[0].argv, c.want) {
			t.Errorf("%s: argv = %v, want %v", c.name, f.calls[0].argv, c.want)
		}
	}
}

// TestScriptInstalledReadsMarkers pins the marker contract, including that a
// marker is believed even when the engine also reported a non-zero exit: the
// probe answers on stdout precisely so the exit status does not have to carry
// two different meanings.
func TestScriptInstalledReadsMarkers(t *testing.T) {
	cases := []struct {
		name string
		out  string
		err  error
		want bool
	}{
		{name: "present marker", out: ScriptPresentMarker + "\n", want: true},
		{name: "absent marker", out: ScriptAbsentMarker + "\n", want: false},
		{name: "marker among engine chatter", out: "Defaulted container to connector\n" + ScriptAbsentMarker + "\n", want: false},
		{name: "present marker despite a non-zero exit", out: ScriptPresentMarker + "\n", err: fmt.Errorf("exit status 1"), want: true},
	}
	for _, c := range cases {
		f := &fakeRunner{out: c.out, err: c.err}
		got, err := ScriptInstalled(f, []string{"kubectl"}, validate.PlatformKubernetes, "pod-0", "", "/tmp/solmq-status.sh")
		if err != nil {
			t.Fatalf("%s: unexpected error %v", c.name, err)
		}
		if got != c.want {
			t.Errorf("%s: got %v, want %v", c.name, got, c.want)
		}
	}
}

// TestScriptInstalledUnreachableTargetIsError covers the case the markers
// exist to separate out: no answer at all means the probe never ran, which must
// surface as an error rather than a silent "absent" that would trigger an
// install against a target that cannot be reached.
func TestScriptInstalledUnreachableTargetIsError(t *testing.T) {
	cases := []struct {
		name string
		out  string
		err  error
	}{
		{name: "engine error, no marker", out: "Error: No such container: solmq-connector\n", err: fmt.Errorf("exit status 1")},
		{name: "clean exit but no marker", out: "unexpected\n"},
	}
	for _, c := range cases {
		f := &fakeRunner{out: c.out, err: c.err}
		got, err := ScriptInstalled(f, []string{"docker"}, validate.PlatformDocker, "solmq-connector", "", "/tmp/solmq-status.sh")
		if err == nil {
			t.Fatalf("%s: a probe with no marker must surface as an error", c.name)
		}
		if got {
			t.Errorf("%s: want false alongside the error", c.name)
		}
	}
}

func TestScriptInstalledUnknownPlatform(t *testing.T) {
	f := &fakeRunner{}
	if _, err := ScriptInstalled(f, []string{"kubectl"}, "bogus", "pod-0", "", "/tmp/solmq-status.sh"); err == nil {
		t.Fatal("unknown platform must be rejected")
	}
	if len(f.calls) != 0 {
		t.Fatal("nothing must run for an unknown platform")
	}
}

// ---- InstallScript --------------------------------------------------------------

func TestInstallScriptArgv(t *testing.T) {
	cases := []struct {
		name      string
		cmd       []string
		platform  string
		target    string
		namespace string
		want      []string
	}{
		{
			name: "kubernetes no namespace", cmd: []string{"kubectl"}, platform: validate.PlatformKubernetes,
			target: "pod-0",
			want:   []string{"kubectl", "exec", "-i", "pod-0", "-c", spec.ConnectorContainerName, "--", "sh", "-c", "mkdir -p /tmp && cat > /tmp/solmq-status.sh"},
		},
		{
			name: "kubernetes with namespace", cmd: []string{"oc"}, platform: validate.PlatformKubernetes,
			target: "pod-0", namespace: "solace",
			want: []string{"oc", "exec", "-i", "pod-0", "-n", "solace", "-c", spec.ConnectorContainerName, "--", "sh", "-c", "mkdir -p /tmp && cat > /tmp/solmq-status.sh"},
		},
		{
			name: "docker", cmd: []string{"docker"}, platform: validate.PlatformDocker,
			target: "solmq-connector",
			want:   []string{"docker", "exec", "-i", "solmq-connector", "sh", "-c", "mkdir -p /tmp && cat > /tmp/solmq-status.sh"},
		},
		{
			name: "podman", cmd: []string{"podman"}, platform: validate.PlatformPodman,
			target: "solmq-connector",
			want:   []string{"podman", "exec", "-i", "solmq-connector", "sh", "-c", "mkdir -p /tmp && cat > /tmp/solmq-status.sh"},
		},
	}
	for _, c := range cases {
		f := &fakeRunner{}
		if _, err := InstallScript(f, c.cmd, c.platform, c.target, c.namespace, "/tmp", "/tmp/solmq-status.sh", "#!/bin/sh\necho ok\n"); err != nil {
			t.Fatalf("%s: %v", c.name, err)
		}
		if !reflect.DeepEqual(f.calls[0].argv, c.want) {
			t.Errorf("%s: argv = %v, want %v", c.name, f.calls[0].argv, c.want)
		}
	}
}

func TestInstallScriptPassesScriptOnStdinNotArgv(t *testing.T) {
	f := &fakeRunner{}
	script := "#!/bin/sh\necho ok\n"
	if _, err := InstallScript(f, []string{"docker"}, validate.PlatformDocker, "solmq-connector", "", "/tmp", "/tmp/solmq-status.sh", script); err != nil {
		t.Fatal(err)
	}
	if f.calls[0].stdin != script {
		t.Errorf("stdin = %q, want the script body", f.calls[0].stdin)
	}
	for _, a := range f.calls[0].argv {
		if strings.Contains(a, "echo ok") {
			t.Errorf("script body must never appear in argv, found in %q", a)
		}
	}
}

func TestInstallScriptUnknownPlatform(t *testing.T) {
	f := &fakeRunner{}
	if _, err := InstallScript(f, []string{"kubectl"}, "bogus", "pod-0", "", "/tmp", "/tmp/x.sh", "x"); err == nil {
		t.Fatal("unknown platform must be rejected")
	}
	if len(f.calls) != 0 {
		t.Fatal("nothing must run for an unknown platform")
	}
}

// ---- RunStatusScript ------------------------------------------------------------

func TestRunStatusScriptArgv(t *testing.T) {
	cases := []struct {
		name      string
		cmd       []string
		platform  string
		target    string
		namespace string
		want      []string
	}{
		{
			name: "kubernetes no namespace", cmd: []string{"kubectl"}, platform: validate.PlatformKubernetes,
			target: "pod-0",
			want:   []string{"kubectl", "exec", "pod-0", "-c", spec.ConnectorContainerName, "--", "sh", "/tmp/solmq-status.sh"},
		},
		{
			name: "kubernetes with namespace", cmd: []string{"oc"}, platform: validate.PlatformKubernetes,
			target: "pod-0", namespace: "solace",
			want: []string{"oc", "exec", "pod-0", "-n", "solace", "-c", spec.ConnectorContainerName, "--", "sh", "/tmp/solmq-status.sh"},
		},
		{
			name: "docker", cmd: []string{"docker"}, platform: validate.PlatformDocker,
			target: "solmq-connector",
			want:   []string{"docker", "exec", "solmq-connector", "sh", "/tmp/solmq-status.sh"},
		},
		{
			name: "podman", cmd: []string{"podman"}, platform: validate.PlatformPodman,
			target: "solmq-connector",
			want:   []string{"podman", "exec", "solmq-connector", "sh", "/tmp/solmq-status.sh"},
		},
	}
	for _, c := range cases {
		f := &fakeRunner{}
		if _, err := RunStatusScript(f, c.cmd, c.platform, c.target, c.namespace, "/tmp/solmq-status.sh"); err != nil {
			t.Fatalf("%s: %v", c.name, err)
		}
		if !reflect.DeepEqual(f.calls[0].argv, c.want) {
			t.Errorf("%s: argv = %v, want %v", c.name, f.calls[0].argv, c.want)
		}
	}
}

// TestRunStatusScriptReturnsOutputAlongsideNonZeroExit pins the contract: the
// status script's own exit convention (1 standby, 2 error) is expected, not a
// runner failure, so the output must never be dropped just because the exit
// code is non-zero.
func TestRunStatusScriptReturnsOutputAlongsideNonZeroExit(t *testing.T) {
	f := &fakeRunner{err: fmt.Errorf("exit status 1"), out: "STANDBY\n"}
	out, err := RunStatusScript(f, []string{"kubectl"}, validate.PlatformKubernetes, "pod-0", "", "/tmp/solmq-status.sh")
	if err == nil {
		t.Fatal("a non-zero script exit is expected and must be returned, not swallowed")
	}
	if out != "STANDBY\n" {
		t.Errorf("output = %q, want %q alongside the error", out, "STANDBY\n")
	}
}

func TestRunStatusScriptUnknownPlatform(t *testing.T) {
	f := &fakeRunner{}
	if _, err := RunStatusScript(f, []string{"kubectl"}, "bogus", "pod-0", "", "/tmp/x.sh"); err == nil {
		t.Fatal("unknown platform must be rejected")
	}
	if len(f.calls) != 0 {
		t.Fatal("nothing must run for an unknown platform")
	}
}

// ---- OS.Stream (the real streaming boundary) --------------------------------

// writerFunc adapts a function to io.Writer, so a test can observe output at the
// moment it is written rather than after the process has exited -- which is the
// whole distinction between Stream and Run.
type writerFunc func(p []byte) (int, error)

func (f writerFunc) Write(p []byte) (int, error) { return f(p) }

// TestOSStreamDeliversOutputBeforeExitAndCancelIsCleanEnd is the test that
// justifies the Streamer seam existing at all: the child prints one line and
// then blocks for far longer than this test will wait, so seeing that line
// proves the output was not buffered until exit. Cancelling then ends the run,
// and reports success -- a follow that the operator stopped did not fail.
func TestOSStreamDeliversOutputBeforeExitAndCancelIsCleanEnd(t *testing.T) {
	t.Setenv("GO_WANT_HELPER_PROCESS", "1")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var once sync.Once
	arrived := make(chan struct{})
	var mu sync.Mutex
	var got strings.Builder
	w := writerFunc(func(p []byte) (int, error) {
		mu.Lock()
		got.Write(p)
		seen := strings.Contains(got.String(), "drip-line")
		mu.Unlock()
		if seen {
			once.Do(func() { close(arrived) })
		}
		return len(p), nil
	})

	errc := make(chan error, 1)
	go func() {
		errc <- (OS{}).Stream(ctx, Cmd{Argv: helperProcessArgv(os.Args[0], "drip")}, w, io.Discard)
	}()

	select {
	case <-arrived:
	case err := <-errc:
		t.Fatalf("the child exited before its first line was seen (err=%v); Stream is buffering", err)
	case <-time.After(30 * time.Second):
		t.Fatal("no output arrived while the child was still running; Stream is buffering")
	}

	cancel()
	select {
	case err := <-errc:
		if err != nil {
			t.Errorf("a cancelled stream is a clean end, got %v", err)
		}
	case <-time.After(30 * time.Second):
		t.Fatal("Stream did not return after the context was cancelled")
	}
}

// TestOSStreamKeepsStdoutAndStderrApart pins the reason Stream takes two
// writers where Run merges into one: `logs > app.log` must land log lines in the
// file and leave the platform's own diagnostics on the terminal.
func TestOSStreamKeepsStdoutAndStderrApart(t *testing.T) {
	t.Setenv("GO_WANT_HELPER_PROCESS", "1")
	var out, errOut strings.Builder
	if err := (OS{}).Stream(context.Background(), Cmd{Argv: helperProcessArgv(os.Args[0], "both")}, &out, &errOut); err != nil {
		t.Fatalf("Stream: %v", err)
	}
	if !strings.Contains(out.String(), "stdout-line") || strings.Contains(out.String(), "stderr-line") {
		t.Errorf("stdout writer should carry only stdout, got %q", out.String())
	}
	if !strings.Contains(errOut.String(), "stderr-line") || strings.Contains(errOut.String(), "stdout-line") {
		t.Errorf("stderr writer should carry only stderr, got %q", errOut.String())
	}
}

// TestOSStreamReportsAFailureThatWasNotCancelled is the other half of the
// cancellation rule: an uncancelled non-zero exit is a real failure and must
// still be reported.
func TestOSStreamReportsAFailureThatWasNotCancelled(t *testing.T) {
	t.Setenv("GO_WANT_HELPER_PROCESS", "1")
	var out strings.Builder
	err := (OS{}).Stream(context.Background(), Cmd{Argv: helperProcessArgv(os.Args[0], "fail")}, &out, io.Discard)
	if err == nil {
		t.Fatal("a non-zero exit with no cancellation must be an error")
	}
	if !strings.Contains(out.String(), "before-exit") {
		t.Errorf("output written before the failure should still reach the writer, got %q", out.String())
	}
}

// TestOSStreamRejectsEmptyAndUnresolvableArgv pins that Stream refuses exactly
// what Run refuses -- both go through resolveArgv0, so the two cannot drift.
func TestOSStreamRejectsEmptyAndUnresolvableArgv(t *testing.T) {
	if err := (OS{}).Stream(context.Background(), Cmd{}, io.Discard, io.Discard); err == nil {
		t.Error("an empty argv must be refused")
	}
	err := (OS{}).Stream(context.Background(), Cmd{Argv: []string{"solmq-no-such-binary-anywhere"}}, io.Discard, io.Discard)
	if err == nil || !strings.Contains(err.Error(), "resolving") {
		t.Errorf("an unresolvable binary must be refused by name, got %v", err)
	}
}

// ---- LogsArgv ---------------------------------------------------------------

// TestLogsArgvPerPlatform pins the two shapes, which differ in more than
// spelling: kubectl takes the pod as a positional with the namespace and
// container as flags, docker and podman take their options first and the
// container name last. The container is always named on kubernetes, whatever
// the pod holds -- see ExecArgv for why.
func TestLogsArgvPerPlatform(t *testing.T) {
	for _, c := range []struct {
		name      string
		platform  string
		namespace string
		opts      LogsOpts
		want      []string
	}{
		{
			name: "kubernetes bare", platform: validate.PlatformKubernetes, namespace: "prod",
			opts: LogsOpts{Tail: TailAll},
			want: []string{"cli", "logs", "pod-a", "-n", "prod", "-c", spec.ConnectorContainerName},
		},
		{
			name: "kubernetes without a namespace still names the container", platform: validate.PlatformKubernetes,
			opts: LogsOpts{Tail: TailAll},
			want: []string{"cli", "logs", "pod-a", "-c", spec.ConnectorContainerName},
		},
		{
			name: "kubernetes with every option", platform: validate.PlatformKubernetes, namespace: "prod",
			opts: LogsOpts{Follow: true, Previous: true, Timestamps: true, Tail: 50, Since: "10m0s"},
			want: []string{"cli", "logs", "pod-a", "-n", "prod", "-c", spec.ConnectorContainerName, "-p", "-f", "--timestamps", "--tail", "50", "--since", "10m0s"},
		},
		{
			name: "docker bare", platform: validate.PlatformDocker,
			opts: LogsOpts{Tail: TailAll},
			want: []string{"cli", "logs", "pod-a"},
		},
		{
			name: "docker with every option it has", platform: validate.PlatformDocker,
			opts: LogsOpts{Follow: true, Timestamps: true, Tail: 50, Since: "10m0s"},
			want: []string{"cli", "logs", "-f", "--timestamps", "--tail", "50", "--since", "10m0s", "pod-a"},
		},
		{
			name: "podman reads the container, never the journal", platform: validate.PlatformPodman,
			opts: LogsOpts{Tail: TailAll},
			want: []string{"cli", "logs", "pod-a"},
		},
		{
			// 0 is a real request -- the options only, no history -- so it cannot
			// double as the "unset" marker TailAll is.
			name: "tail zero is not tail all", platform: validate.PlatformDocker,
			opts: LogsOpts{Tail: 0},
			want: []string{"cli", "logs", "--tail", "0", "pod-a"},
		},
	} {
		t.Run(c.name, func(t *testing.T) {
			got, err := LogsArgv([]string{"cli"}, c.platform, "pod-a", c.namespace, c.opts)
			if err != nil {
				t.Fatalf("LogsArgv: %v", err)
			}
			if !reflect.DeepEqual(got, c.want) {
				t.Errorf("argv = %v, want %v", got, c.want)
			}
		})
	}
}

// TestLogsArgvRefusesPreviousOffKubernetes pins the guard behind the CLI's own
// refusal: neither engine keeps a prior run's log under the same name, so the
// option is refused rather than dropped -- a caller that asked for the previous
// log must never be handed the current one instead.
func TestLogsArgvRefusesPreviousOffKubernetes(t *testing.T) {
	for _, platform := range []string{validate.PlatformDocker, validate.PlatformPodman} {
		_, err := LogsArgv([]string{"docker"}, platform, "c", "", LogsOpts{Previous: true, Tail: TailAll})
		if err == nil {
			t.Fatalf("%s: --previous must be refused", platform)
		}
		if !strings.Contains(err.Error(), platform) {
			t.Errorf("%s: the error should name the platform, got %v", platform, err)
		}
	}
}

// TestLogsArgvUnknownPlatform mirrors ExecArgv's refusal: an unrecognised
// platform names the three that exist rather than producing a half-built argv.
func TestLogsArgvUnknownPlatform(t *testing.T) {
	_, err := LogsArgv([]string{"kubectl"}, "nomad", "c", "", LogsOpts{Tail: TailAll})
	if err == nil {
		t.Fatal("an unknown platform must be refused")
	}
	for _, want := range []string{validate.PlatformKubernetes, validate.PlatformDocker, validate.PlatformPodman} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("the error should name %q, got %v", want, err)
		}
	}
}

// ---- ExecArgv (the shared exec prefix) -----------------------------------------

// TestExecArgvPerPlatform pins the shape each engine's exec actually parses,
// which is a parser difference rather than a spelling one: kubectl accepts its
// own flags after the pod positional and needs the "--" terminator to know
// where the in-container command begins, while docker and podman stop parsing
// their own flags at the container name -- so theirs must come first, and a
// "--" would reach the container as an argument.
//
// The container is asserted on every kubernetes row, including the one with no
// namespace: it is a constant, not something discovery supplies.
func TestExecArgvPerPlatform(t *testing.T) {
	for _, c := range []struct {
		name      string
		platform  string
		namespace string
		opts      ExecOpts
		want      []string
	}{
		{
			name: "kubernetes bare", platform: validate.PlatformKubernetes,
			want: []string{"cli", "exec", "pod-0", "-c", spec.ConnectorContainerName, "--"},
		},
		{
			name: "kubernetes with a namespace", platform: validate.PlatformKubernetes, namespace: "prod",
			want: []string{"cli", "exec", "pod-0", "-n", "prod", "-c", spec.ConnectorContainerName, "--"},
		},
		{
			name: "kubernetes with stdin", platform: validate.PlatformKubernetes, namespace: "prod",
			opts: ExecOpts{Stdin: true},
			want: []string{"cli", "exec", "-i", "pod-0", "-n", "prod", "-c", spec.ConnectorContainerName, "--"},
		},
		{
			name: "kubernetes with a terminal", platform: validate.PlatformKubernetes, namespace: "prod",
			opts: ExecOpts{Stdin: true, TTY: true},
			want: []string{"cli", "exec", "-i", "-t", "pod-0", "-n", "prod", "-c", spec.ConnectorContainerName, "--"},
		},
		{
			// No -c and no "--": both would be handed to the container, which
			// stops parsing engine flags at its own name.
			name: "docker", platform: validate.PlatformDocker, namespace: "ignored",
			opts: ExecOpts{Stdin: true, TTY: true},
			want: []string{"cli", "exec", "-i", "-t", "pod-0"},
		},
		{
			name: "podman", platform: validate.PlatformPodman,
			opts: ExecOpts{Stdin: true},
			want: []string{"cli", "exec", "-i", "pod-0"},
		},
	} {
		t.Run(c.name, func(t *testing.T) {
			got, err := ExecArgv([]string{"cli"}, c.platform, "pod-0", c.namespace, c.opts)
			if err != nil {
				t.Fatalf("ExecArgv: %v", err)
			}
			if !reflect.DeepEqual(got, c.want) {
				t.Errorf("argv = %v, want %v", got, c.want)
			}
		})
	}
}

// TestExecArgvRefusesATTYWithoutStdin pins the guard: a terminal the child
// cannot read from can only be a caller mistake, so it is refused rather than
// quietly produced.
func TestExecArgvRefusesATTYWithoutStdin(t *testing.T) {
	_, err := ExecArgv([]string{"kubectl"}, validate.PlatformKubernetes, "pod-0", "", ExecOpts{TTY: true})
	if err == nil {
		t.Fatal("a tty without stdin must be refused")
	}
	if !strings.Contains(err.Error(), "stdin") {
		t.Errorf("the error should say what is missing, got %v", err)
	}
}

// TestExecArgvUnknownPlatform mirrors LogsArgv's refusal: an unrecognised
// platform names the three that exist rather than producing a half-built argv.
func TestExecArgvUnknownPlatform(t *testing.T) {
	_, err := ExecArgv([]string{"kubectl"}, "nomad", "c", "", ExecOpts{})
	if err == nil {
		t.Fatal("an unknown platform must be refused")
	}
	for _, want := range []string{validate.PlatformKubernetes, validate.PlatformDocker, validate.PlatformPodman} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("the error should name %q, got %v", want, err)
		}
	}
}

// ---- OS.Attach (the terminal-handover boundary) --------------------------------

// regularFile reports whether f is a plain file rather than a pipe, which is
// how the stdiokind helper mode tells an inherited descriptor from one os/exec
// substituted a pipe for. Go reports an anonymous pipe as ModeNamedPipe on
// Windows and unix alike, so IsRegular separates the two on both.
func regularFile(f *os.File) bool {
	fi, err := f.Stat()
	return err == nil && fi.Mode().IsRegular()
}

// attachFiles opens three temp files to stand in for a terminal, since a test
// has no terminal to hand over. They are regular files, which is the whole
// point: the child reports back what kind of descriptor it was given.
func attachFiles(t *testing.T) (stdin, stdout, stderr *os.File) {
	t.Helper()
	dir := t.TempDir()
	open := func(name string) *os.File {
		f, err := os.Create(filepath.Join(dir, name))
		if err != nil {
			t.Fatalf("creating %s: %v", name, err)
		}
		t.Cleanup(func() { _ = f.Close() })
		return f
	}
	return open("stdin"), open("stdout"), open("stderr")
}

// readFile returns what the child wrote to one of those files.
func readFile(t *testing.T, f *os.File) string {
	t.Helper()
	b, err := os.ReadFile(f.Name())
	if err != nil {
		t.Fatalf("reading %s: %v", f.Name(), err)
	}
	return string(b)
}

// TestOSAttachHandsTheChildTheCallersFilesNotPipes is the test that justifies
// the seam existing at all. Stream copies through io.Writers, and os/exec
// answers any writer that is not an *os.File with a pipe -- so a future
// "simplification" of Attach onto io.Writer would silently take the terminal
// away from the engine and leave an operator with a shell that has no prompt.
// The child reports what it was actually given.
func TestOSAttachHandsTheChildTheCallersFilesNotPipes(t *testing.T) {
	t.Setenv("GO_WANT_HELPER_PROCESS", "1")
	stdin, stdout, stderr := attachFiles(t)
	code, err := (OS{}).Attach(Cmd{Argv: helperProcessArgv(os.Args[0], "stdiokind")}, stdin, stdout, stderr)
	if err != nil || code != 0 {
		t.Fatalf("Attach = (%d, %v), want (0, nil); stderr %q", code, err, readFile(t, stderr))
	}
	if got, want := readFile(t, stdout), "in=true out=true err=true\n"; !strings.Contains(got, want) {
		t.Errorf("the child saw %q, want %q -- os/exec interposed a pipe", got, want)
	}
}

// TestOSAttachStdinIsTheHandedFile pins the other half: what the caller gives
// as stdin is what the child reads, with nothing copied through this process.
func TestOSAttachStdinIsTheHandedFile(t *testing.T) {
	t.Setenv("GO_WANT_HELPER_PROCESS", "1")
	stdin, stdout, stderr := attachFiles(t)
	if _, err := stdin.WriteString("hello-attach\n"); err != nil {
		t.Fatalf("writing stdin: %v", err)
	}
	if _, err := stdin.Seek(0, io.SeekStart); err != nil {
		t.Fatalf("rewinding stdin: %v", err)
	}
	code, err := (OS{}).Attach(Cmd{Argv: helperProcessArgv(os.Args[0], "stdin")}, stdin, stdout, stderr)
	if err != nil || code != 0 {
		t.Fatalf("Attach = (%d, %v), want (0, nil); stderr %q", code, err, readFile(t, stderr))
	}
	if got := readFile(t, stdout); got != "hello-attach\n" {
		t.Errorf("child echoed %q, want the handed stdin", got)
	}
}

// TestOSAttachReportsTheChildExitStatus pins the fact the cli verb's exit-code
// contract rests on: a session that ended non-zero is reported as that status,
// not as an error, because only the caller knows whether it means the engine
// could not attach or the operator's last command failed.
func TestOSAttachReportsTheChildExitStatus(t *testing.T) {
	t.Setenv("GO_WANT_HELPER_PROCESS", "1")
	stdin, stdout, stderr := attachFiles(t)
	code, err := (OS{}).Attach(Cmd{Argv: helperProcessArgv(os.Args[0], "fail")}, stdin, stdout, stderr)
	if err != nil {
		t.Fatalf("a child that ran and exited non-zero is not an error here: %v", err)
	}
	if code != 3 {
		t.Errorf("code = %d, want 3", code)
	}
	if got := readFile(t, stdout); !strings.Contains(got, "before-exit") {
		t.Errorf("the child's output should still have reached the caller's file, got %q", got)
	}
}

// TestOSAttachEnvReachesChild covers the third call site of the applyCmdEnv
// split, so the credential channel cannot behave differently here than under
// Run.
func TestOSAttachEnvReachesChild(t *testing.T) {
	t.Setenv("GO_WANT_HELPER_PROCESS", "1")
	t.Setenv("RUNNER_TEST_AMBIENT_ONLY", "ambient")
	t.Setenv("RUNNER_TEST_OVERRIDE", "ambient")
	stdin, stdout, stderr := attachFiles(t)
	cmd := Cmd{
		Argv: helperProcessArgv(os.Args[0], "env"),
		Env:  []string{"RUNNER_TEST_OVERRIDE=supplied", "RUNNER_TEST_EXTRA_ONLY=extra"},
	}
	code, err := (OS{}).Attach(cmd, stdin, stdout, stderr)
	if err != nil || code != 0 {
		t.Fatalf("Attach = (%d, %v), want (0, nil); stderr %q", code, err, readFile(t, stderr))
	}
	want := "AMBIENT_ONLY=ambient OVERRIDE=supplied EXTRA_ONLY=extra\n"
	if got := readFile(t, stdout); got != want {
		t.Errorf("child env = %q, want %q", got, want)
	}
}

// TestOSAttachRefusesACmdCarryingStdinText pins the contradiction guard: a Cmd
// that wants a string written to the child cannot also be a session reading the
// operator's keyboard, and honouring one of them silently is the class of bug
// LogsArgv's --previous guard exists to prevent.
func TestOSAttachRefusesACmdCarryingStdinText(t *testing.T) {
	stdin, stdout, stderr := attachFiles(t)
	_, err := (OS{}).Attach(Cmd{Argv: []string{"go"}, Stdin: "payload"}, stdin, stdout, stderr)
	if err == nil {
		t.Fatal("a Cmd carrying stdin text must be refused")
	}
	if !strings.Contains(err.Error(), "Cmd.Stdin") {
		t.Errorf("the error should name the field, got %v", err)
	}
}

// TestOSAttachRefusesANilFile fails here rather than at the child's first
// write, where os/exec panics on a typed-nil *os.File in its io.Writer field.
func TestOSAttachRefusesANilFile(t *testing.T) {
	stdin, stdout, stderr := attachFiles(t)
	for _, c := range []struct {
		name          string
		in, out, errf *os.File
	}{
		{"no stdin", nil, stdout, stderr},
		{"no stdout", stdin, nil, stderr},
		{"no stderr", stdin, stdout, nil},
	} {
		t.Run(c.name, func(t *testing.T) {
			if _, err := (OS{}).Attach(Cmd{Argv: []string{"go"}}, c.in, c.out, c.errf); err == nil {
				t.Fatal("a nil file must be refused")
			}
		})
	}
}

// TestOSAttachRejectsEmptyAndUnresolvableArgv mirrors the Run and Stream
// refusals: all three go through resolveArgv0, so the rules cannot drift apart.
func TestOSAttachRejectsEmptyAndUnresolvableArgv(t *testing.T) {
	stdin, stdout, stderr := attachFiles(t)
	if _, err := (OS{}).Attach(Cmd{}, stdin, stdout, stderr); err == nil {
		t.Error("an empty argv must be refused")
	}
	if _, err := (OS{}).Attach(Cmd{Argv: []string{"solmq-no-such-binary-xyz"}}, stdin, stdout, stderr); err == nil {
		t.Error("an unresolvable argv[0] must be refused")
	}
}

// ---- OS.RunSplit and runParsed (the parse-safety boundary) ---------------------

// TestOSRunSplitKeepsTheStreamsApart is the regression test for the bug this
// seam exists to fix: `oc get -o json` on OpenShift prints "Warning: ... is
// deprecated" on stderr, and Run merges that into stdout ahead of the JSON,
// where it fails to parse on its first character. Split, the answer is clean.
func TestOSRunSplitKeepsTheStreamsApart(t *testing.T) {
	t.Setenv("GO_WANT_HELPER_PROCESS", "1")
	out, errOut, err := (OS{}).RunSplit(Cmd{Argv: helperProcessArgv(os.Args[0], "both")})
	if err != nil {
		t.Fatalf("RunSplit returned error: %v", err)
	}
	if out != "stdout-line\n" {
		t.Errorf("stdout = %q, want only the stdout line", out)
	}
	if errOut != "stderr-line\n" {
		t.Errorf("stderr = %q, want only the stderr line", errOut)
	}
}

// TestOSRunSplitWiresStdinAndEnv pins that the split path is otherwise Run: the
// credential channel and stdin behave identically, so a caller does not have to
// know which one a helper used.
func TestOSRunSplitWiresStdinAndEnv(t *testing.T) {
	t.Setenv("GO_WANT_HELPER_PROCESS", "1")
	out, _, err := (OS{}).RunSplit(Cmd{Argv: helperProcessArgv(os.Args[0], "stdin"), Stdin: "hello-split\n"})
	if err != nil || out != "hello-split\n" {
		t.Fatalf("stdin: out=%q err=%v", out, err)
	}

	t.Setenv("RUNNER_TEST_AMBIENT_ONLY", "ambient")
	t.Setenv("RUNNER_TEST_OVERRIDE", "ambient")
	out, _, err = (OS{}).RunSplit(Cmd{
		Argv: helperProcessArgv(os.Args[0], "env"),
		Env:  []string{"RUNNER_TEST_OVERRIDE=supplied", "RUNNER_TEST_EXTRA_ONLY=extra"},
	})
	if err != nil {
		t.Fatalf("env: %v", err)
	}
	if want := "AMBIENT_ONLY=ambient OVERRIDE=supplied EXTRA_ONLY=extra\n"; out != want {
		t.Errorf("env: out = %q, want %q", out, want)
	}
}

// TestOSRunSplitRejectsEmptyAndUnresolvableArgv mirrors Run, Stream and Attach:
// all four go through resolveArgv0, so the refusals cannot drift apart.
func TestOSRunSplitRejectsEmptyAndUnresolvableArgv(t *testing.T) {
	if _, _, err := (OS{}).RunSplit(Cmd{}); err == nil {
		t.Error("an empty argv must be refused")
	}
	if _, _, err := (OS{}).RunSplit(Cmd{Argv: []string{"solmq-no-such-binary-xyz"}}); err == nil {
		t.Error("an unresolvable argv[0] must be refused")
	}
}

// splitRunner is a fake that satisfies Splitter as well as Runner, so the
// helpers can be tested on the path production actually takes. Run is left
// deliberately wrong -- it returns the two streams merged, the way OS.Run does
// -- so a helper that failed to use the split path produces unparseable output
// and the test says so.
type splitRunner struct {
	calls  []call
	stdout string
	stderr string
	err    error
}

func (s *splitRunner) Run(c Cmd) (string, error) {
	s.calls = append(s.calls, call{argv: c.Argv, stdin: c.Stdin, env: c.Env})
	return s.stderr + s.stdout, s.err
}

func (s *splitRunner) RunSplit(c Cmd) (string, string, error) {
	s.calls = append(s.calls, call{argv: c.Argv, stdin: c.Stdin, env: c.Env})
	return s.stdout, s.stderr, s.err
}

// TestParsingHelpersIgnoreAWarningOnStderr walks every helper whose output is
// parsed rather than scanned, with a warning waiting on stderr. Each must
// return the payload alone.
//
// The list is the point of the test as much as the assertion is: a new helper
// that parses output and forgets runParsed is a bug that only shows up against
// a cluster that warns, which is the one place it is hardest to notice.
func TestParsingHelpersIgnoreAWarningOnStderr(t *testing.T) {
	const warning = "Warning: apps.openshift.io/v1 DeploymentConfig is deprecated in v4.14+\n"
	const payload = `{"items":[]}`
	sc := QuadletScope{UserMode: false}

	for _, c := range []struct {
		name string
		call func(r Runner) (string, error)
	}{
		{"KubernetesPodsJSON", func(r Runner) (string, error) {
			return KubernetesPodsJSON(r, []string{"oc"}, "ns", "", nil, false)
		}},
		{"KubernetesGetJSON", func(r Runner) (string, error) {
			return KubernetesGetJSON(r, []string{"oc"}, "ns", "deployment", "d")
		}},
		{"KubernetesListJSON", func(r Runner) (string, error) {
			return KubernetesListJSON(r, []string{"oc"}, "ns", "all")
		}},
		{"KubernetesTop", func(r Runner) (string, error) {
			return KubernetesTop(r, []string{"oc"}, "ns", "", nil, false)
		}},
		{"EngineInspectJSON", func(r Runner) (string, error) {
			return EngineInspectJSON(r, []string{"docker"}, []string{"c"})
		}},
		{"EngineImageInspectJSON", func(r Runner) (string, error) {
			return EngineImageInspectJSON(r, []string{"docker"}, "img:1")
		}},
		{"EngineStats", func(r Runner) (string, error) {
			return EngineStats(r, []string{"docker"}, []string{"c"})
		}},
		{"EngineList", func(r Runner) (string, error) {
			return EngineList(r, []string{"docker"})
		}},
	} {
		t.Run(c.name, func(t *testing.T) {
			r := &splitRunner{stdout: payload, stderr: warning}
			got, err := c.call(r)
			if err != nil {
				t.Fatalf("%s: %v", c.name, err)
			}
			if got != payload {
				t.Errorf("output = %q, want the payload alone -- a warning on stderr must not reach the parser", got)
			}
		})
	}

	// SystemctlNRestarts parses an integer, so the same warning would turn a
	// restart count into an error. It answers a number rather than a string,
	// hence its own case.
	t.Run("SystemctlNRestarts", func(t *testing.T) {
		r := &splitRunner{stdout: "3\n", stderr: warning}
		n, err := SystemctlNRestarts(r, sc, "x.service")
		if err != nil {
			t.Fatalf("SystemctlNRestarts: %v", err)
		}
		if n != 3 {
			t.Errorf("restarts = %d, want 3", n)
		}
	})
}

// TestParsingHelpersFallBackToRun pins the other half of runParsed: a Runner
// with no Splitter still works, on exactly the combined output Run has always
// returned. Every argv-recording fake in the suite depends on that.
func TestParsingHelpersFallBackToRun(t *testing.T) {
	f := &fakeRunner{out: `{"items":[]}`}
	got, err := KubernetesListJSON(f, []string{"kubectl"}, "ns", "all")
	if err != nil {
		t.Fatalf("KubernetesListJSON: %v", err)
	}
	if got != `{"items":[]}` {
		t.Errorf("output = %q, want what Run returned", got)
	}
	if len(f.calls) != 1 {
		t.Errorf("want 1 call through Run, got %d", len(f.calls))
	}
}

// TestParsedFailureCarriesBothStreams pins that splitting did not cost the
// error context Run was merging stderr in for: a failed command still reports
// whatever it said, on whichever stream it said it on.
func TestParsedFailureCarriesBothStreams(t *testing.T) {
	r := &splitRunner{
		stdout: "partial-output",
		stderr: "Error from server (Forbidden): pods is forbidden",
		err:    fmt.Errorf("exit status 1"),
	}
	_, err := KubernetesListJSON(r, []string{"oc"}, "ns", "all")
	if err == nil {
		t.Fatal("a failed command must be reported")
	}
	for _, want := range []string{"Forbidden", "partial-output"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("the error should carry %q, got %v", want, err)
		}
	}
}
