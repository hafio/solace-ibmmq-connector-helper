package runner

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/spec"
)

// fakeRunner records every invocation so tests can assert the exact argv/stdin
// crossing the exec boundary without starting a process.
type fakeRunner struct {
	calls     []call
	err       error         // returned by every Run when errByCall has no entry (nil = success)
	errByCall map[int]error // per-call-index override, so a later call can fail while earlier ones succeed
}

type call struct {
	argv  []string
	stdin string
}

func (f *fakeRunner) Run(argv []string, stdin string) (string, error) {
	idx := len(f.calls)
	f.calls = append(f.calls, call{argv: argv, stdin: stdin})
	if e, ok := f.errByCall[idx]; ok {
		return "out", e
	}
	return "out", f.err
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
	default:
		fmt.Fprintf(os.Stderr, "TestHelperProcess: unknown mode %q\n", args[0])
		os.Exit(2)
	}
	os.Exit(0)
}

func TestOSRunWiresStdinToChild(t *testing.T) {
	t.Setenv("GO_WANT_HELPER_PROCESS", "1")
	out, err := (OS{}).Run(helperProcessArgv(os.Args[0], "stdin"), "hello-stdin\n")
	if err != nil {
		t.Fatalf("Run returned error: %v (output %q)", err, out)
	}
	if out != "hello-stdin\n" {
		t.Errorf("stdin content not visible to child: output = %q", out)
	}
}

func TestOSRunCombinesStdoutAndStderr(t *testing.T) {
	t.Setenv("GO_WANT_HELPER_PROCESS", "1")
	out, err := (OS{}).Run(helperProcessArgv(os.Args[0], "both"), "")
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
	out, err := (OS{}).Run(helperProcessArgv(os.Args[0], "fail"), "")
	if err == nil {
		t.Fatalf("non-zero child exit must return a non-nil error, output = %q", out)
	}
	if !strings.Contains(out, "before-exit") {
		t.Errorf("output written before the failing exit must still be captured: %q", out)
	}
}

// TestOSRunAcceptsAbsolutePathArgv0 pins the trust decision at the nolint on
// runner.go:47 -- argv[0] can come from an env.yaml command token, which
// validate.SafeToken allows to name any binary (letters, digits, and
// / . : - _ among the allowed chars), so an absolute path must run exactly
// like a bare command name does.
func TestOSRunAcceptsAbsolutePathArgv0(t *testing.T) {
	exe, err := os.Executable()
	if err != nil {
		t.Fatalf("os.Executable: %v", err)
	}
	if !filepath.IsAbs(exe) {
		t.Fatalf("os.Executable() returned a non-absolute path %q", exe)
	}
	t.Setenv("GO_WANT_HELPER_PROCESS", "1")
	out, err := (OS{}).Run(helperProcessArgv(exe, "stdin"), "abs-argv0-ok\n")
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
		{"kubectl; rm -rf /", nil, false},   // ';' is unsafe
		{"kubectl $(evil)", nil, false},     // '$' and '(' unsafe
		{"kubectl `id`", nil, false},        // backtick unsafe
		{`kubectl --kubeconfig "a b"`, nil, false}, // quote+space unsafe
	}
	for _, c := range cases {
		got, err := ParseCommand(c.in)
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

func TestKubernetesDeployApplyOnStdin(t *testing.T) {
	f := &fakeRunner{}
	manifest := "apiVersion: v1\nkind: Namespace\n"
	if _, err := Kubernetes(f, "kubectl --context prod", ActionDeploy, manifest); err != nil {
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

func TestKubernetesDeleteUsesDeleteVerb(t *testing.T) {
	f := &fakeRunner{}
	if _, err := Kubernetes(f, "oc", ActionDelete, "x"); err != nil {
		t.Fatal(err)
	}
	wantArgv := []string{"oc", "delete", "-f", "-"}
	if !reflect.DeepEqual(f.calls[0].argv, wantArgv) {
		t.Errorf("argv = %v, want %v", f.calls[0].argv, wantArgv)
	}
}

func TestKubernetesRejectsUnsafeCommand(t *testing.T) {
	f := &fakeRunner{}
	if _, err := Kubernetes(f, "kubectl; rm -rf /", ActionDeploy, "x"); err == nil {
		t.Fatal("unsafe command must be rejected")
	}
	if len(f.calls) != 0 {
		t.Fatal("nothing must be executed when the command is rejected")
	}
}

func TestKubernetesUnknownAction(t *testing.T) {
	f := &fakeRunner{}
	if _, err := Kubernetes(f, "kubectl", "restart", "x"); err == nil {
		t.Fatal("unknown action must be rejected")
	}
	if len(f.calls) != 0 {
		t.Fatal("nothing must run for an unknown action")
	}
}

func TestDockerUpAndDown(t *testing.T) {
	f := &fakeRunner{}
	if _, err := Docker(f, "docker", ActionDeploy, "/tmp/docker-compose.yml"); err != nil {
		t.Fatal(err)
	}
	wantUp := []string{"docker", "compose", "-f", "/tmp/docker-compose.yml", "up", "-d"}
	if !reflect.DeepEqual(f.calls[0].argv, wantUp) {
		t.Errorf("up argv = %v, want %v", f.calls[0].argv, wantUp)
	}
	if _, err := Docker(f, "docker", ActionDelete, "/tmp/docker-compose.yml"); err != nil {
		t.Fatal(err)
	}
	wantDown := []string{"docker", "compose", "-f", "/tmp/docker-compose.yml", "down"}
	if !reflect.DeepEqual(f.calls[1].argv, wantDown) {
		t.Errorf("down argv = %v, want %v", f.calls[1].argv, wantDown)
	}
}

func TestDockerRejectsUnsafeCommand(t *testing.T) {
	f := &fakeRunner{}
	if _, err := Docker(f, "docker `id`", ActionDeploy, "x"); err == nil {
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

func TestPodmanDeleteStopsRemovesReloads(t *testing.T) {
	f := &fakeRunner{}
	dir := t.TempDir()
	unit := "solmq-connector.container"
	if err := os.WriteFile(filepath.Join(dir, unit), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	sc := QuadletScope{Dir: dir, UserMode: false}
	if _, err := PodmanDelete(f, sc, []string{"solmq-connector.service"}, []string{unit}); err != nil {
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
	if _, err := Docker(f, "docker", "restart", "x"); err == nil {
		t.Fatal("unknown action must be rejected")
	}
	if len(f.calls) != 0 {
		t.Fatal("nothing must run for an unknown action")
	}
}

func TestPodmanDeleteStopFailureIsReported(t *testing.T) {
	// The first call PodmanDelete makes is the per-service stop, so a runner that
	// fails every call exercises the stop-failure branch.
	f := &fakeRunner{err: fmt.Errorf("boom")}
	sc := QuadletScope{Dir: t.TempDir(), UserMode: false}
	_, err := PodmanDelete(f, sc, []string{"solmq-connector.service"}, []string{"solmq-connector.container"})
	if err == nil {
		t.Fatal("a systemctl stop failure must surface")
	}
	if !strings.Contains(err.Error(), "stop solmq-connector.service") {
		t.Errorf("error should name the failed stop, got %v", err)
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
