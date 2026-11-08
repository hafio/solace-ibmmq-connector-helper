package runner

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/hafio/go-deploy-utility/execguard"
	"github.com/hafio/go-deploy-utility/exitcode"
	"github.com/hafio/go-deploy-utility/kube"
	"github.com/hafio/go-deploy-utility/kube/nsgate"
	"github.com/hafio/go-deploy-utility/kube/rbac"
	"github.com/hafio/go-deploy-utility/podman"
	gdurunner "github.com/hafio/go-deploy-utility/runner"
	"github.com/hafio/go-deploy-utility/runner/runnertest"
	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/spec"
	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/validate"
)

// noApprovals is the zero Approvals: a run without --allow-command.
var noApprovals execguard.Approvals

// approve builds the approvals a run with --allow-command <names> would carry,
// failing the test on a value Approve refuses.
func approve(t *testing.T, names ...string) execguard.Approvals {
	t.Helper()
	a, err := execguard.Approve("solmq-conn-util", names)
	if err != nil {
		t.Fatalf("Approve(%v) = %v", names, err)
	}
	return a
}

// fakeRunner records every invocation so tests can assert the exact argv/stdin
// crossing the exec boundary without starting a process.
type fakeRunner struct {
	calls     []call
	err       error         // returned by every Run when errByCall has no entry (nil = success)
	errByCall map[int]error // per-call-index override, so a later call can fail while earlier ones succeed
	out       string        // returned by every Run when outByCall has no entry ("" falls back to "out", or "yes" for auth can-i)
	outByCall map[int]string
}

type call struct {
	argv  []string
	stdin string
	env   []string
}

func (f *fakeRunner) Run(_ context.Context, c gdurunner.Cmd, stdout, _ io.Writer) error {
	idx := len(f.calls)
	f.calls = append(f.calls, call{argv: c.Argv, stdin: string(c.Stdin), env: c.Env})
	out, ok := f.outByCall[idx]
	if !ok {
		out = f.out
		switch {
		case out != "":
		case isCanI(c.Argv):
			// rbac.Check refuses anything but a "yes", so a probe answers it by
			// default and a test scripts the refusal it wants.
			out = "yes\n"
		default:
			out = "out"
		}
	}
	emit(stdout, out)
	if e, ok := f.errByCall[idx]; ok {
		return e
	}
	return f.err
}

// Attach is never reached: nothing in this package attaches a terminal.
func (f *fakeRunner) Attach(c gdurunner.Cmd, _, _, _ *os.File) (int, error) {
	return 0, fmt.Errorf("%s: this fake does not attach", c.Argv[0])
}

// emit writes a fake's scripted output to the stream a caller handed it. A nil
// writer discards the stream, as the shared runner's does.
func emit(w io.Writer, s string) {
	if w != nil && s != "" {
		fmt.Fprint(w, s)
	}
}

// isCanI reports whether argv is a `kubectl auth can-i` probe.
func isCanI(argv []string) bool {
	for i := 0; i+1 < len(argv); i++ {
		if argv[i] == "auth" && argv[i+1] == "can-i" {
			return true
		}
	}
	return false
}

// bg is the context the helper tests run under.
func bg() context.Context { return context.Background() }

func TestParseCommand(t *testing.T) {
	cases := []struct {
		in   string
		want []string
		ok   bool
	}{
		{"kubectl", []string{"kubectl"}, true},
		{"kubectl --context prod -n solace", []string{"kubectl", "--context", "prod", "-n", "solace"}, true},
		{"  oc   ", []string{"oc"}, true},
		{"kubectl.Exe", []string{"kubectl.Exe"}, true},
		{"kubectl --kubeconfig=C:/RUNNER~1/config", []string{"kubectl", "--kubeconfig=C:/RUNNER~1/config"}, true},
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
		got, err := ParseCommand(validate.PlatformKubernetes, c.in, noApprovals)
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
		} else if !exitcode.IsUsage(err) {
			t.Errorf("ParseCommand(%q) refusal must be a usage error (exit 2), got %v", c.in, err)
		}
	}
}

// TestParseCommandExtraAllowed pins the escape hatch: a chained binary (e.g.
// `lima podman`) is rejected on the platform allowlist alone but accepted once
// its argv[0] is approved, and the platform's own binary still follows in argv
// unmodified. A privilege-escalation wrapper is the one chained binary that no
// approval reaches: `sudo podman` is refused as a usage error whatever the
// approvals say, and --allow-command sudo is refused by Approve itself.
func TestParseCommandExtraAllowed(t *testing.T) {
	if _, err := ParseCommand(validate.PlatformPodman, "lima podman", noApprovals); err == nil {
		t.Fatal("lima podman must be rejected without --allow-command lima")
	}
	got, err := ParseCommand(validate.PlatformPodman, "lima podman", approve(t, "lima"))
	if err != nil {
		t.Fatalf("ParseCommand with lima approved unexpected error: %v", err)
	}
	want := []string{"lima", "podman"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("ParseCommand(%q) = %v, want %v", "lima podman", got, want)
	}

	_, err = ParseCommand(validate.PlatformPodman, "sudo podman", approve(t, "lima"))
	if err == nil || !strings.Contains(err.Error(), `podman.command[0] = "sudo" is a privilege-escalation wrapper (sudo), which is never permitted`) {
		t.Fatalf("sudo podman must be refused as an escalation wrapper, got %v", err)
	}
	if !exitcode.IsUsage(err) {
		t.Errorf("an escalation refusal must be a usage error (exit 2), got %v", err)
	}
	if _, err := execguard.Approve("solmq-conn-util", []string{"sudo"}); err == nil || !strings.Contains(err.Error(), "sudo solmq-conn-util") {
		t.Errorf("--allow-command sudo must be refused with the run-it-under-sudo advice, got %v", err)
	}
}

func TestKubernetesDeployApplyOnStdin(t *testing.T) {
	f := &fakeRunner{}
	manifest := "apiVersion: v1\nkind: Namespace\n"
	if _, err := Kubernetes(bg(), f, "kubectl --context prod", ActionDeploy, manifest, noApprovals); err != nil {
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
	if _, err := Kubernetes(bg(), f, "oc", ActionRemove, "x", noApprovals); err != nil {
		t.Fatal(err)
	}
	wantArgv := []string{"oc", "delete", "-f", "-", "--ignore-not-found"}
	if !reflect.DeepEqual(f.calls[0].argv, wantArgv) {
		t.Errorf("argv = %v, want %v", f.calls[0].argv, wantArgv)
	}
}

func TestKubernetesRejectsUnsafeCommand(t *testing.T) {
	f := &fakeRunner{}
	if _, err := Kubernetes(bg(), f, "kubectl; rm -rf /", ActionDeploy, "x", noApprovals); err == nil {
		t.Fatal("unsafe command must be rejected")
	}
	if len(f.calls) != 0 {
		t.Fatal("nothing must be executed when the command is rejected")
	}
}

// TestKubernetesRolloutStatusRejectsUnsafeCommand: the rollout wait goes through the
// same guard as the apply it follows, so a refused command waits on nothing.
func TestKubernetesRolloutStatusRejectsUnsafeCommand(t *testing.T) {
	f := &fakeRunner{}
	if _, err := KubernetesRolloutStatus(bg(), f, "kubectl; rm -rf /", "ns", "c", noApprovals); err == nil {
		t.Fatal("unsafe command must be rejected")
	}
	if len(f.calls) != 0 {
		t.Fatal("nothing must be executed when the command is rejected")
	}
}

func TestKubernetesUnknownAction(t *testing.T) {
	f := &fakeRunner{}
	if _, err := Kubernetes(bg(), f, "kubectl", "restart", "x", noApprovals); err == nil {
		t.Fatal("unknown action must be rejected")
	}
	if len(f.calls) != 0 {
		t.Fatal("nothing must run for an unknown action")
	}
}

func TestDockerUpAndDown(t *testing.T) {
	f := &fakeRunner{}
	env := []string{"MQ_CONN_1_USER=admin", "MQ_CONN_1_PASSWORD=s3cr3t"}
	const compose = "name: proj\nservices: {}\n"
	if _, err := Docker(bg(), f, "docker", ActionDeploy, compose, "/srv/spec", env, noApprovals); err != nil {
		t.Fatal(err)
	}
	wantUp := []string{"docker", "compose", "-f", "-", "--project-directory", "/srv/spec", "up", "-d"}
	if !reflect.DeepEqual(f.calls[0].argv, wantUp) {
		t.Errorf("up argv = %v, want %v", f.calls[0].argv, wantUp)
	}
	if f.calls[0].stdin != compose {
		t.Errorf("up stdin = %q, want the compose file %q (it never touches the disk)", f.calls[0].stdin, compose)
	}
	if !reflect.DeepEqual(f.calls[0].env, env) {
		t.Errorf("up env = %v, want %v (credential values must cross via Cmd.Env, never argv)", f.calls[0].env, env)
	}
	if _, err := Docker(bg(), f, "docker", ActionRemove, compose, "/srv/spec", env, noApprovals); err != nil {
		t.Fatal(err)
	}
	wantDown := []string{"docker", "compose", "-f", "-", "--project-directory", "/srv/spec", "down"}
	if !reflect.DeepEqual(f.calls[1].argv, wantDown) {
		t.Errorf("down argv = %v, want %v", f.calls[1].argv, wantDown)
	}
	if f.calls[1].stdin != compose {
		t.Errorf("down stdin = %q, want the compose file", f.calls[1].stdin)
	}
	if !reflect.DeepEqual(f.calls[1].env, env) {
		t.Errorf("down env = %v, want %v (down must resolve the same secret declarations up did)", f.calls[1].env, env)
	}
}

func TestDockerRejectsUnsafeCommand(t *testing.T) {
	f := &fakeRunner{}
	if _, err := Docker(bg(), f, "docker `id`", ActionDeploy, "x", "/srv/spec", nil, noApprovals); err == nil {
		t.Fatal("unsafe command must be rejected")
	}
	if len(f.calls) != 0 {
		t.Fatal("nothing must run when the command is rejected")
	}
}

// scopeEnv is a ScopeEnv for a test: a fixed euid, the given environment, home
// /home/alice, and a runtime directory that exists only where dirs says so. It
// reads nothing from the real process.
func scopeEnv(euid int, vars map[string]string, dirs ...string) ScopeEnv {
	return ScopeEnv{
		Euid:   euid,
		Getenv: func(k string) string { return vars[k] },
		Home:   func() (string, error) { return "/home/alice", nil },
		IsDir: func(p string) (bool, error) {
			for _, d := range dirs {
				if d == p {
					return true, nil
				}
			}
			return false, nil
		},
		Linux: true,
	}
}

// TestResolveQuadletScope pins the scope to the rootless rule, the quadlet
// directory and the user session, each read from the inputs it is given rather
// than from the process: so it needs no skip as root and redirects no HOME.
func TestResolveQuadletScope(t *testing.T) {
	const runtimeDir = "/run/user/1000"
	user := "/home/alice/.config/containers/systemd"
	sess := []string{"XDG_RUNTIME_DIR=" + runtimeDir, "DBUS_SESSION_BUS_ADDRESS=unix:path=" + runtimeDir + "/bus"}
	t.Run("the six rootless decisions", func(t *testing.T) {
		windows := scopeEnv(-1, nil)
		windows.Linux = false
		for _, c := range []struct {
			name    string
			env     ScopeEnv
			rootful bool
			want    QuadletScope
			wantErr string
		}{
			{"non-root, rootful unset: rootless", scopeEnv(1000, nil, runtimeDir), false,
				QuadletScope{Dir: user, UserMode: true, Euid: 1000, Env: sess}, ""},
			{"non-root asking for rootful: refused", scopeEnv(1000, nil, runtimeDir), true,
				QuadletScope{}, "podman.rootful=true requires root, but running as euid 1000"},
			{"root asking for rootful: rootful", scopeEnv(0, nil), true,
				QuadletScope{Dir: "/etc/containers/systemd", Euid: 0}, ""},
			{"root, rootful unset: refused", scopeEnv(0, nil), false,
				QuadletScope{}, "set podman.rootful=true to deploy the rootful engine"},
			{"no euid, rootful unset: rootless, no session", windows, false,
				QuadletScope{Dir: "/home/alice/.config/containers/systemd", UserMode: true, Euid: -1}, ""},
			{"no euid, rootful: rootful", windows, true,
				QuadletScope{Dir: "/etc/containers/systemd", Euid: -1}, ""},
		} {
			t.Run(c.name, func(t *testing.T) {
				got, err := ResolveQuadletScope(c.env, c.rootful)
				if c.wantErr != "" {
					if err == nil || !strings.Contains(err.Error(), c.wantErr) {
						t.Fatalf("err = %v, want it to contain %q", err, c.wantErr)
					}
					if !exitcode.IsUsage(err) {
						t.Errorf("a refused rootless rule is a usage error (exit 2), got %v", err)
					}
					return
				}
				if err != nil {
					t.Fatal(err)
				}
				if !reflect.DeepEqual(got, c.want) {
					t.Errorf("scope = %+v, want %+v", got, c.want)
				}
			})
		}
	})

	t.Run("XDG_CONFIG_HOME", func(t *testing.T) {
		got, err := ResolveQuadletScope(scopeEnv(1000, map[string]string{"XDG_CONFIG_HOME": "/srv/xdg"}, runtimeDir), false)
		if err != nil || got.Dir != "/srv/xdg/containers/systemd" {
			t.Errorf("an absolute XDG_CONFIG_HOME is honoured: dir=%q err=%v", got.Dir, err)
		}
		_, err = ResolveQuadletScope(scopeEnv(1000, map[string]string{"XDG_CONFIG_HOME": "rel/xdg"}, runtimeDir), false)
		if err == nil || !strings.Contains(err.Error(), "XDG_CONFIG_HOME") || !exitcode.IsUsage(err) {
			t.Errorf("a relative XDG_CONFIG_HOME is a usage error naming it, got %v", err)
		}
		// Rootful never reads the user's config directory.
		got, err = ResolveQuadletScope(scopeEnv(0, map[string]string{"XDG_CONFIG_HOME": "rel/xdg"}), true)
		if err != nil || got.Dir != "/etc/containers/systemd" {
			t.Errorf("rootful ignores XDG_CONFIG_HOME: dir=%q err=%v", got.Dir, err)
		}
	})

	t.Run("unresolvable home", func(t *testing.T) {
		env := scopeEnv(1000, nil, runtimeDir)
		env.Home = func() (string, error) { return "", errors.New("$HOME is not defined") }
		_, err := ResolveQuadletScope(env, false)
		if err == nil || !strings.Contains(err.Error(), "$HOME is not defined") || !exitcode.IsUsage(err) {
			t.Errorf("an unresolvable home is a usage error carrying the cause, got %v", err)
		}
	})

	t.Run("session", func(t *testing.T) {
		got, err := ResolveQuadletScope(scopeEnv(1000, nil, runtimeDir), false)
		if err != nil || !reflect.DeepEqual(got.Env, sess) {
			t.Errorf("a stripped session is derived: env=%v err=%v", got.Env, err)
		}
		inherited := map[string]string{
			"XDG_RUNTIME_DIR":          "/run/user/1000",
			"DBUS_SESSION_BUS_ADDRESS": "unix:path=/run/user/1000/bus",
		}
		got, err = ResolveQuadletScope(scopeEnv(1000, inherited, runtimeDir), false)
		if err != nil || len(got.Env) != 0 {
			t.Errorf("an inherited session is left as it is: env=%v err=%v", got.Env, err)
		}
		_, err = ResolveQuadletScope(scopeEnv(1000, nil), false)
		if err == nil || exitcode.IsUsage(err) || !strings.Contains(err.Error(), runtimeDir) {
			t.Errorf("a missing runtime directory is a plain failure (exit 1) naming it, got %v", err)
		}
		probeFails := scopeEnv(1000, nil)
		probeFails.IsDir = func(string) (bool, error) { return false, errors.New("permission denied") }
		if _, err = ResolveQuadletScope(probeFails, false); err == nil || !strings.Contains(err.Error(), "permission denied") {
			t.Errorf("a probe that cannot run is an error carrying the cause, got %v", err)
		}
		// Off Linux there is no session to repair, and nothing is probed.
		other := scopeEnv(1000, nil)
		other.Linux = false
		other.IsDir = func(string) (bool, error) { t.Error("no probe off Linux"); return false, nil }
		if got, err = ResolveQuadletScope(other, false); err != nil || len(got.Env) != 0 {
			t.Errorf("non-Linux has no session: env=%v err=%v", got.Env, err)
		}
		// A rootful scope has no user session either.
		if got, err = ResolveQuadletScope(scopeEnv(0, nil), true); err != nil || len(got.Env) != 0 {
			t.Errorf("a rootful scope has no session: env=%v err=%v", got.Env, err)
		}
	})
}

// LiveScopeEnv reads this process, and statIsDir tells a missing path (no
// answer to give) from one that cannot be looked at (an error).
func TestLiveScopeEnv(t *testing.T) {
	env := LiveScopeEnv()
	if env.Euid != os.Geteuid() || env.Linux != (runtime.GOOS == "linux") || env.Getenv == nil || env.Home == nil {
		t.Errorf("LiveScopeEnv() = %+v, want this process's values", env)
	}
	dir := t.TempDir()
	file := filepath.Join(dir, "f")
	if err := os.WriteFile(file, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	for _, c := range []struct {
		path string
		want bool
	}{{dir, true}, {file, false}, {filepath.Join(dir, "gone"), false}} {
		if got, err := env.IsDir(c.path); err != nil || got != c.want {
			t.Errorf("IsDir(%q) = %v, %v, want %v, nil", c.path, got, err, c.want)
		}
	}
	if _, err := env.IsDir("bad\x00path"); err == nil {
		t.Error("a path that cannot be looked at is an error, not a missing directory")
	}
}

// WithEnv attaches the session to a rootless scope's commands, never aliasing the
// scope's slice, and leaves a system scope's commands as they are.
func TestQuadletScopeWithEnv(t *testing.T) {
	user := QuadletScope{UserMode: true, Env: []string{"XDG_RUNTIME_DIR=/run/user/1000"}}
	c := user.WithEnv(gdurunner.Cmd{Argv: []string{"x"}, Env: []string{"A=1"}})
	if want := []string{"A=1", "XDG_RUNTIME_DIR=/run/user/1000"}; !reflect.DeepEqual(c.Env, want) {
		t.Errorf("env = %v, want %v", c.Env, want)
	}
	c.Env[0] = "changed"
	if user.Env[0] != "XDG_RUNTIME_DIR=/run/user/1000" {
		t.Error("WithEnv must copy, not append into the scope's own slice")
	}
	if got := (QuadletScope{Env: user.Env}).WithEnv(gdurunner.Cmd{}); got.Env != nil {
		t.Errorf("a system scope carries no session, got %v", got.Env)
	}
}

// EnsureLinger asks logind, makes no change when the user lingers, enables it for
// the user's own account when not, and says what to ask an administrator when
// that is refused. Every call carries the session.
func TestEnsureLinger(t *testing.T) {
	sc := QuadletScope{UserMode: true, Euid: 1000, Env: []string{"XDG_RUNTIME_DIR=/run/user/1000"}}
	show := []string{"loginctl", "show-user", "1000", "--property=Linger"}
	enable := []string{"loginctl", "enable-linger"}
	for _, c := range []struct {
		name        string
		f           *fakeRunner
		wantEnabled bool
		wantErr     []string
		wantArgv    [][]string
	}{
		{"already lingers: no enable call", &fakeRunner{out: "Linger=yes\n"}, false, nil, [][]string{show}},
		{"off: enabled", &fakeRunner{outByCall: map[int]string{0: "Linger=no\n"}}, true, nil, [][]string{show, enable}},
		{"enable refused: the administrator form", &fakeRunner{outByCall: map[int]string{0: "Linger=no\n", 1: "Access denied"},
			errByCall: map[int]error{1: errors.New("exit status 1")}}, false,
			[]string{"Linger=no", "Access denied", "sudo loginctl enable-linger alice"}, [][]string{show, enable}},
		{"probe fails: an error naming the command", &fakeRunner{err: errors.New("exit status 1")}, false,
			[]string{"loginctl show-user 1000 --property=Linger", "logind"}, [][]string{show}},
	} {
		t.Run(c.name, func(t *testing.T) {
			enabled, err := EnsureLinger(bg(), c.f, sc, "alice")
			if enabled != c.wantEnabled {
				t.Errorf("enabled = %v, want %v", enabled, c.wantEnabled)
			}
			if len(c.wantErr) == 0 && err != nil {
				t.Fatal(err)
			}
			for _, w := range c.wantErr {
				if err == nil || !strings.Contains(err.Error(), w) {
					t.Errorf("err = %v, want it to contain %q", err, w)
				}
			}
			if len(c.f.calls) != len(c.wantArgv) {
				t.Fatalf("calls = %+v, want %v", c.f.calls, c.wantArgv)
			}
			for i, w := range c.wantArgv {
				if !reflect.DeepEqual(c.f.calls[i].argv, w) || !reflect.DeepEqual(c.f.calls[i].env, sc.Env) {
					t.Errorf("call %d = %v env %v, want %v env %v", i, c.f.calls[i].argv, c.f.calls[i].env, w, sc.Env)
				}
			}
		})
	}
	state, err := ProbeLinger(bg(), &splitRunner{stdout: "Linger=no\n", stderr: "a warning\n"}, sc)
	if err != nil || state != "no" {
		t.Errorf("ProbeLinger = %q, %v, want the stdout state alone", state, err)
	}
}

// PodmanIDMap reads the mapping from stdout alone, for the uid map or the gid map.
func TestPodmanIDMap(t *testing.T) {
	for _, c := range []struct {
		gid  bool
		want string
	}{{false, "UIDMap"}, {true, "GIDMap"}} {
		r := &splitRunner{stdout: "0:1 1:65536 \n", stderr: "WARN[0000] a warning\n"}
		out, err := PodmanIDMap(bg(), r, "podman", c.gid, noApprovals)
		if err != nil || out != "0:1 1:65536 \n" {
			t.Fatalf("out=%q err=%v", out, err)
		}
		if want := podman.IDMap([]string{"podman"}, c.gid).Argv; !reflect.DeepEqual(r.calls[0].argv, want) {
			t.Errorf("argv = %v, want the %s query %v", r.calls[0].argv, c.want, want)
		}
	}
	if _, err := PodmanIDMap(bg(), &splitRunner{err: errors.New("boom")}, "podman", false, noApprovals); err == nil ||
		!strings.Contains(err.Error(), "boom") {
		t.Errorf("a failed query is an error carrying the cause, got %v", err)
	}
	f := &fakeRunner{}
	if _, err := PodmanIDMap(bg(), f, "podman; rm -rf /", false, noApprovals); err == nil || len(f.calls) != 0 {
		t.Errorf("an unsafe command is refused before anything runs: err=%v calls=%d", err, len(f.calls))
	}
}

// TestPodmanDeployReloadThenStart pins how a deploy brings the unit up: reload,
// ask systemd for the service's state, then restart it when it is active or
// reloading (a running container keeps the secrets it was created with, so only a
// restart loads the new ones) and start it for any other answer, or none.
// deployFns are the two seams PodmanDeploy takes, recording when they ran so a test
// can assert the order: the question comes before the write.
type deployFns struct {
	asked, wrote int
	answer       bool
	askErr       error
	writeErr     error
	callsAtWrite int
	f            *fakeRunner
}

func (d *deployFns) approve() (bool, error) { d.asked++; return d.answer, d.askErr }
func (d *deployFns) write() (string, error) {
	d.wrote++
	if d.f != nil { // nil where the test records calls on another runner
		d.callsAtWrite = len(d.f.calls)
	}
	return "wrote\n", d.writeErr
}

func TestPodmanDeployReloadThenStart(t *testing.T) {
	for _, c := range []struct {
		name    string
		user    bool
		state   string
		changed bool
		verb    string // "" leaves the service alone
	}{
		{"inactive unit, user scope", true, "inactive\n", true, "start"},
		{"active unit, user scope", true, "active\n", true, "restart"},
		{"reloading unit, user scope", true, "reloading\n", true, "restart"},
		{"failed unit, user scope", true, "failed\n", true, "start"},
		{"unanswered probe starts, user scope", true, "", true, "start"},
		{"inactive unit, system scope", false, "inactive\n", true, "start"},
		{"active unit, system scope", false, "active\n", true, "restart"},
		{"reloading unit, system scope", false, "reloading\n", true, "restart"},
		{"unchanged inactive unit is started, nothing reloaded", true, "inactive\n", false, "start"},
		{"unchanged active unit is left running", true, "active\n", false, ""},
		{"unchanged reloading unit is left running", false, "reloading\n", false, ""},
	} {
		t.Run(c.name, func(t *testing.T) {
			// is-active exits non-zero for every state but active, so a state
			// other than active or reloading arrives with an exit error.
			f := &fakeRunner{outByCall: map[int]string{0: c.state}}
			if c.state != "active\n" && c.state != "reloading\n" {
				f.errByCall = map[int]error{0: fmt.Errorf("exit status 3")}
			}
			d := &deployFns{answer: true, f: f}
			sc := QuadletScope{Dir: t.TempDir(), UserMode: c.user, Env: []string{"XDG_RUNTIME_DIR=/run/user/1000"}}
			out, cancelled, err := PodmanDeploy(bg(), f, sc, []string{"solmq-connector.service"}, c.changed, d.approve, d.write)
			if err != nil || cancelled {
				t.Fatalf("err = %v, cancelled = %v", err, cancelled)
			}
			prefix := []string{"systemctl"}
			if c.user {
				prefix = append(prefix, "--user")
			}
			sysctl := func(args ...string) []string { return append(append([]string(nil), prefix...), args...) }
			want := [][]string{sysctl("is-active", "solmq-connector.service")}
			if c.changed {
				want = append(want, sysctl("daemon-reload"))
			}
			if c.verb != "" {
				want = append(want, sysctl(c.verb, "solmq-connector.service"))
			}
			if len(f.calls) != len(want) {
				t.Fatalf("want %d calls, got %+v", len(want), f.calls)
			}
			for i, w := range want {
				if !reflect.DeepEqual(f.calls[i].argv, w) {
					t.Errorf("call %d argv = %v, want %v", i, f.calls[i].argv, w)
				}
				if c.user && !reflect.DeepEqual(f.calls[i].env, sc.Env) {
					t.Errorf("call %d env = %v, want the session %v", i, f.calls[i].env, sc.Env)
				}
				if !c.user && f.calls[i].env != nil {
					t.Errorf("call %d env = %v, a system scope carries no session", i, f.calls[i].env)
				}
			}
			running := c.state == "active\n" || c.state == "reloading\n"
			wantAsked := 0
			if c.changed && running {
				wantAsked = 1
			}
			wantWrote := 0
			if c.changed {
				wantWrote = 1
			}
			if d.asked != wantAsked || d.wrote != wantWrote {
				t.Errorf("asked %d times (want %d), wrote %d times (want %d)", d.asked, wantAsked, d.wrote, wantWrote)
			}
			if d.wrote == 1 && d.callsAtWrite != 1 {
				t.Errorf("write ran after %d calls, want after the is-active probe alone", d.callsAtWrite)
			}
			if said := strings.Contains(out, "was running; restarting it"); said != (c.verb == "restart") {
				t.Errorf("output %q: restart notice present = %v, want %v", out, said, c.verb == "restart")
			}
			if said := strings.Contains(out, "unchanged and left running"); said != (c.verb == "" && running) {
				t.Errorf("output %q: left-running note present = %v", out, said)
			}
			if c.verb == "" && !strings.Contains(out, "`solmq-conn-util restart`") {
				t.Errorf("output %q should say how to restart it", out)
			}
		})
	}
}

// A running service that would restart asks first, and a "no" leaves the host
// exactly as it was: nothing written, nothing reloaded, nothing restarted.
func TestPodmanDeployDeclinedChangesNothing(t *testing.T) {
	f := &fakeRunner{outByCall: map[int]string{0: "active\n"}}
	d := &deployFns{answer: false, f: f}
	sc := QuadletScope{Dir: t.TempDir(), UserMode: true}
	out, cancelled, err := PodmanDeploy(bg(), f, sc, []string{"a.service"}, true, d.approve, d.write)
	if err != nil || !cancelled || out != "" {
		t.Fatalf("out = %q, cancelled = %v, err = %v; want a cancelled, silent deploy", out, cancelled, err)
	}
	if d.asked != 1 || d.wrote != 0 {
		t.Errorf("asked %d, wrote %d; want one question and no write", d.asked, d.wrote)
	}
	if len(f.calls) != 1 || f.calls[0].argv[len(f.calls[0].argv)-2] != "is-active" {
		t.Errorf("only the state probe may run, got %+v", f.calls)
	}
}

// A question that cannot be asked ends the deploy with its own error, before any
// write: no terminal and no --no-prompt is the typical cause.
func TestPodmanDeployUnaskableQuestionStopsBeforeAnyWrite(t *testing.T) {
	f := &fakeRunner{outByCall: map[int]string{0: "active\n"}}
	d := &deployFns{askErr: fmt.Errorf("confirming the redeploy interactively: no terminal; pass --no-prompt instead"), f: f}
	sc := QuadletScope{Dir: t.TempDir()}
	_, cancelled, err := PodmanDeploy(bg(), f, sc, []string{"a.service"}, true, d.approve, d.write)
	if err == nil || cancelled || !strings.Contains(err.Error(), "--no-prompt") {
		t.Fatalf("err = %v, cancelled = %v; want the question's own error", err, cancelled)
	}
	if d.wrote != 0 || len(f.calls) != 1 {
		t.Errorf("wrote %d times, %d calls; want no write and only the probe", d.wrote, len(f.calls))
	}
}

// A write that fails stops the deploy where it fails: no reload, no start.
func TestPodmanDeployWriteFailureStopsBeforeSystemctl(t *testing.T) {
	f := &fakeRunner{outByCall: map[int]string{0: "inactive\n"}, errByCall: map[int]error{0: fmt.Errorf("exit status 3")}}
	d := &deployFns{writeErr: fmt.Errorf("boom"), f: f}
	sc := QuadletScope{Dir: t.TempDir()}
	out, _, err := PodmanDeploy(bg(), f, sc, []string{"a.service"}, true, d.approve, d.write)
	if err == nil || err.Error() != "boom" {
		t.Fatalf("err = %v, want the write's own", err)
	}
	if !strings.Contains(out, "wrote") || len(f.calls) != 1 {
		t.Errorf("out = %q, %d calls; want the write's output kept and only the probe run", out, len(f.calls))
	}
}

func TestPodmanDeploySystemModeNoUserFlag(t *testing.T) {
	f := &fakeRunner{}
	d := &deployFns{answer: true, f: f}
	sc := QuadletScope{Dir: t.TempDir(), UserMode: false}
	if _, _, err := PodmanDeploy(bg(), f, sc, []string{"a.service"}, true, d.approve, d.write); err != nil {
		t.Fatal(err)
	}
	// Call 0 is the state probe, call 1 the reload.
	if reflect.DeepEqual(f.calls[1].argv, []string{"systemctl", "--user", "daemon-reload"}) {
		t.Error("system mode must not pass --user")
	}
	if !reflect.DeepEqual(f.calls[1].argv, []string{"systemctl", "daemon-reload"}) {
		t.Errorf("argv = %v", f.calls[1].argv)
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
	if _, err := PodmanRemove(bg(), f, sc, []string{"solmq-connector.service"}, []string{unit}); err != nil {
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
	// is-active is call 0 and decides the verb, daemon-reload is call 1 and the
	// start or restart is call 2; a failure injected at any step must surface
	// naming the step, and nothing after it may run. The verb follows the state
	// is-active prints, as systemctl does: "inactive" with exit 3 starts, and
	// "active" restarts.
	for _, c := range []struct {
		name      string
		errs      map[int]error
		probe     string
		wantErr   string
		wantCalls int
	}{
		{"daemon-reload", map[int]error{1: fmt.Errorf("boom")}, "active\n", "systemctl daemon-reload: boom", 2},
		{"start", map[int]error{0: fmt.Errorf("exit status 3"), 2: fmt.Errorf("boom")}, "inactive\n",
			"systemctl start a.service: boom", 3},
		{"restart", map[int]error{2: fmt.Errorf("boom")}, "active\n", "systemctl restart a.service: boom", 3},
	} {
		t.Run(c.name, func(t *testing.T) {
			f := &fakeRunner{errByCall: c.errs, outByCall: map[int]string{0: c.probe}}
			d := &deployFns{answer: true, f: f}
			sc := QuadletScope{Dir: t.TempDir(), UserMode: true}
			_, _, err := PodmanDeploy(bg(), f, sc, []string{"a.service"}, true, d.approve, d.write)
			if err == nil || !strings.Contains(err.Error(), c.wantErr) {
				t.Fatalf("err = %v, want it to contain %q", err, c.wantErr)
			}
			if len(f.calls) != c.wantCalls {
				t.Fatalf("want %d calls, got %d", c.wantCalls, len(f.calls))
			}
		})
	}
}

func TestDockerUnknownAction(t *testing.T) {
	f := &fakeRunner{}
	if _, err := Docker(bg(), f, "docker", "restart", "x", "/srv/spec", nil, noApprovals); err == nil {
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
	_, err := PodmanRemove(bg(), f, sc, []string{"solmq-connector.service"}, []string{"solmq-connector.container"})
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
	if _, err := PodmanSecretCreate(bg(), f, "podman", name, value, false, noApprovals); err != nil {
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

// With replace (podman 4.7+) a secret is replaced in one create --replace,
// the value on stdin, so a failed create leaves the previous one in place.
func TestPodmanSecretCreateReplacesInOneCall(t *testing.T) {
	f := &fakeRunner{}
	if _, err := PodmanSecretCreate(bg(), f, "podman", "c-tls-ts", "bytes", true, noApprovals); err != nil {
		t.Fatal(err)
	}
	if want := []string{"podman", "secret", "create", "--replace", "c-tls-ts", "-"}; len(f.calls) != 1 || !reflect.DeepEqual(f.calls[0].argv, want) || f.calls[0].stdin != "bytes" {
		t.Errorf("calls = %+v, want one %v with the value on stdin", f.calls, want)
	}
	failing := &fakeRunner{err: fmt.Errorf("boom")}
	if _, err := PodmanSecretCreate(bg(), failing, "podman", "c-tls-ts", "bytes", true, noApprovals); err == nil || !strings.Contains(err.Error(), "podman secret create --replace c-tls-ts") {
		t.Errorf("err = %v", err)
	}
}

// PodmanVersion reads the client version from stdout alone, and
// podman.SupportsReplace holds it to 4.7, whatever space surrounds it.
func TestPodmanVersion(t *testing.T) {
	r := &splitRunner{stdout: "5.8.2\n", stderr: "WARN[0000] a warning\n"}
	v, err := PodmanVersion(bg(), r, "podman", noApprovals)
	if err != nil || v != "5.8.2" {
		t.Fatalf("v=%q err=%v", v, err)
	}
	if want := []string{"podman", "version", "--format", "{{.Client.Version}}"}; !reflect.DeepEqual(r.calls[0].argv, want) {
		t.Errorf("argv = %v, want %v", r.calls[0].argv, want)
	}
	for v, want := range map[string]bool{"5.8.2": true, "  5.8.2\n": true, "\t4.7.0 \n": true, "4.7.0": true, "4.9.4-rhel": true, "v4.7.1": true, "4.6.2": false, "4.5": false, "3.4.4": false, "": false, "out": false, "4": false} {
		if got := podman.SupportsReplace(v); got != want {
			t.Errorf("podman.SupportsReplace(%q) = %v, want %v", v, got, want)
		}
	}
	padded := &splitRunner{stdout: "  4.7.0 \n"}
	if v, err := PodmanVersion(bg(), padded, "podman", noApprovals); err != nil || !podman.SupportsReplace(v) {
		t.Errorf("a version with surrounding space must parse: v=%q err=%v", v, err)
	}
	if _, err := PodmanVersion(bg(), &splitRunner{err: errors.New("boom")}, "podman", noApprovals); err == nil {
		t.Error("a failed version call is an error")
	}
	if _, err := PodmanVersion(bg(), &splitRunner{}, "podman; id", noApprovals); err == nil {
		t.Error("an unsafe command is refused")
	}
}

func TestPodmanSecretCreateSkipsCreateWhenRmFails(t *testing.T) {
	f := &fakeRunner{err: fmt.Errorf("boom")}
	_, err := PodmanSecretCreate(bg(), f, "podman", "mq-conn-1-password", "s3cr3t", false, noApprovals)
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
	_, err := PodmanSecretCreate(bg(), f, "podman", "mq-conn-1-password", "s3cr3t", false, noApprovals)
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
	if _, err := PodmanSecretCreate(bg(), f, "podman; rm -rf /", "name", "value", false, noApprovals); err == nil {
		t.Fatal("unsafe command must be rejected")
	}
	if len(f.calls) != 0 {
		t.Fatal("nothing must run when the command is rejected")
	}
}

func TestPodmanSecretRemoveBatchesNames(t *testing.T) {
	f := &fakeRunner{}
	names := []string{"mq-conn-1-user", "mq-conn-1-password", "truststore-password"}
	if _, err := PodmanSecretRemove(bg(), f, "podman", names, noApprovals); err != nil {
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
	out, err := PodmanSecretRemove(bg(), f, "podman", nil, noApprovals)
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
	if _, err := PodmanSecretRemove(bg(), f, "podman; rm -rf /", []string{"name"}, noApprovals); err == nil {
		t.Fatal("unsafe command must be rejected")
	}
	if len(f.calls) != 0 {
		t.Fatal("nothing must run when the command is rejected")
	}
}

func TestPodmanSecretRemoveReportsFailure(t *testing.T) {
	f := &fakeRunner{err: fmt.Errorf("boom")}
	_, err := PodmanSecretRemove(bg(), f, "podman", []string{"name"}, noApprovals)
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

// ---- Preflight ----------------------------------------------------------------

// deployProbe is the one probe the failure tests hand a kubernetes Preflight.
var deployProbe = []rbac.Probe{{Verb: "create", Resource: "deployments"}}

func TestPreflightKubernetesArgvDeployNoNamespace(t *testing.T) {
	f := &fakeRunner{}
	if _, err := Preflight(bg(), f, validate.PlatformKubernetes, "kubectl --context prod", false, deployProbe, noApprovals); err != nil {
		t.Fatal(err)
	}
	want := []string{"kubectl", "--context", "prod", "auth", "can-i", "create", "deployments"}
	if len(f.calls) != 1 || !reflect.DeepEqual(f.calls[0].argv, want) {
		t.Errorf("calls = %+v, want one with argv %v", f.calls, want)
	}
}

func TestPreflightKubernetesArgvRemoveWithNamespace(t *testing.T) {
	f := &fakeRunner{}
	probes := []rbac.Probe{{Verb: "delete", Resource: "deployments", Namespace: "solace"}}
	if _, err := Preflight(bg(), f, validate.PlatformKubernetes, "oc", false, probes, noApprovals); err != nil {
		t.Fatal(err)
	}
	want := []string{"oc", "auth", "can-i", "delete", "deployments", "-n", "solace"}
	if len(f.calls) != 1 || !reflect.DeepEqual(f.calls[0].argv, want) {
		t.Errorf("calls = %+v, want one with argv %v", f.calls, want)
	}
}

func TestPreflightDockerArgvIsInfo(t *testing.T) {
	f := &fakeRunner{}
	if _, err := Preflight(bg(), f, validate.PlatformDocker, "docker --context foo", false, nil, noApprovals); err != nil {
		t.Fatal(err)
	}
	want := []string{"docker", "--context", "foo", "info"}
	if !reflect.DeepEqual(f.calls[0].argv, want) {
		t.Errorf("argv = %v, want %v", f.calls[0].argv, want)
	}
}

func TestPreflightPodmanArgvIsInfo(t *testing.T) {
	f := &fakeRunner{}
	if _, err := Preflight(bg(), f, validate.PlatformPodman, "podman", false, nil, noApprovals); err != nil {
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
		rootless bool
		hint     string
	}{
		{validate.PlatformKubernetes, "kubectl", false, "oc login"},
		{validate.PlatformDocker, "docker", false, "sudo systemctl start docker"},
		{validate.PlatformPodman, "podman", true, "systemctl --user start podman.socket"},
		{validate.PlatformPodman, "podman", false, "sudo systemctl start podman.socket"},
	}
	for _, c := range cases {
		f := &fakeRunner{err: fmt.Errorf("connection refused"), out: "dial unix /run/podman.sock: no such file"}
		var probes []rbac.Probe
		if c.platform == validate.PlatformKubernetes {
			probes = deployProbe
		}
		out, err := Preflight(bg(), f, c.platform, c.command, c.rootless, probes, noApprovals)
		if err == nil {
			t.Fatalf("%s: preflight failure must surface as an error", c.platform)
		}
		if !strings.Contains(err.Error(), "connection refused") {
			t.Errorf("%s: error must preserve the underlying cause: %v", c.platform, err)
		}
		if !strings.Contains(err.Error(), c.hint) {
			t.Errorf("%s: error missing actionable hint (want to contain %q): %v", c.platform, c.hint, err)
		}
		if c.platform == validate.PlatformKubernetes {
			if !strings.Contains(err.Error(), "preflight failed for "+c.platform) {
				t.Errorf("%s: error missing platform prefix: %v", c.platform, err)
			}
			continue
		}
		// One message: the engine's own output is inside the error, and the output
		// the caller prints is empty, so nothing is said twice.
		if !strings.Contains(err.Error(), "cannot talk to the") || !strings.Contains(err.Error(), fmt.Sprintf("%q info", c.command)) ||
			!strings.Contains(err.Error(), "no such file") || out != "" {
			t.Errorf("%s: want containercli's one message carrying the engine's output, got out=%q err=%v", c.platform, out, err)
		}
	}
}

// The engine's output is appended to the error, so a control character in it is
// dropped there: a terminal escape from the engine is never acted on.
func TestPreflightFailureDropsControlCharactersFromEngineOutput(t *testing.T) {
	f := &fakeRunner{err: fmt.Errorf("refused"), out: "bad\x1b[2Jnews\nsecond line"}
	_, err := Preflight(bg(), f, validate.PlatformDocker, "docker", false, nil, noApprovals)
	if err == nil || strings.Contains(err.Error(), "\x1b") {
		t.Fatalf("want an error without the escape, got %q", err)
	}
	if !strings.Contains(err.Error(), "bad[2Jnews\nsecond line") {
		t.Errorf("the output should survive with its line break, got %q", err)
	}
}

// TestKubernetesReadVerbsAskNothing: status, logs and cli change nothing, so on
// kubernetes they ask for no permission up front -- the read or exec itself reports
// a refusal -- and a view-only operator is never turned away by a question. docker
// and podman have no permission to ask about: reachability only.
func TestKubernetesReadVerbsAskNothing(t *testing.T) {
	for _, action := range []string{ActionStatus, ActionLogs, ActionCLI} {
		probes, err := KubernetesProbes(action, "prod", nil, false)
		if err != nil || len(probes) != 0 {
			t.Errorf("%s: probes = %+v, err = %v, want none", action, probes, err)
		}
		f := &fakeRunner{}
		if _, err := Preflight(bg(), f, validate.PlatformKubernetes, "kubectl", false, probes, noApprovals); err != nil || len(f.calls) != 0 {
			t.Errorf("%s: err = %v, calls = %+v, want nothing run", action, err, f.calls)
		}
		f = &fakeRunner{}
		if _, err := Preflight(bg(), f, validate.PlatformDocker, "docker", false, nil, noApprovals); err != nil || !reflect.DeepEqual(f.calls[0].argv, []string{"docker", "info"}) {
			t.Errorf("%s on docker: argv = %v, err = %v, want docker info", action, f.calls[0].argv, err)
		}
	}
}

// TestPreflightRefusalNamesThePermission: can-i exits 1 both when the cluster
// is unreachable and when the answer is no. A no is not fixed by logging in, so
// it names the permission and the namespace instead of the login hint, and is a
// *rbac.RefusedError behind the platform prefix.
func TestPreflightRefusalNamesThePermission(t *testing.T) {
	for _, c := range []struct {
		out, ns, want string
	}{
		{"no\n", "prod", `not allowed to create deployments in namespace "prod" (auth can-i answered "no")`},
		{"Warning: something\nno - RBAC: access denied\n", "", `not allowed to create deployments in the current namespace (auth can-i answered "no - RBAC: access denied")`},
	} {
		f := &fakeRunner{out: c.out, err: fmt.Errorf("exit status 1")}
		probes := []rbac.Probe{{Verb: "create", Resource: "deployments", Namespace: c.ns}}
		out, err := Preflight(bg(), f, validate.PlatformKubernetes, "kubectl", false, probes, noApprovals)
		var refused *rbac.RefusedError
		if err == nil || !strings.Contains(err.Error(), c.want) || !errors.As(err, &refused) {
			t.Errorf("out %q: error = %v, want a *RefusedError containing %s", c.out, err, c.want)
		}
		if err != nil && (strings.Contains(err.Error(), "oc login") || !strings.Contains(err.Error(), "preflight failed for kubernetes")) {
			t.Errorf("out %q: a refusal must not blame the login, and names the platform: %v", c.out, err)
		}
		if out != "" {
			t.Errorf("out %q: the answer is in the error; nothing else to print, got %q", c.out, out)
		}
	}
	// An unreachable cluster still gets the login hint.
	f := &fakeRunner{out: "error: You must be logged in to the server (Unauthorized)\n", err: fmt.Errorf("exit status 1")}
	if _, err := Preflight(bg(), f, validate.PlatformKubernetes, "kubectl", false, deployProbe, noApprovals); err == nil || !strings.Contains(err.Error(), "oc login") {
		t.Errorf("an unreachable cluster keeps the login hint, got %v", err)
	}
}

// TestPreflightStopsAtTheFirstRefusal: every probe of a command runs before its
// first write, in order, and the first refusal ends the run -- later ones are not
// asked.
func TestPreflightStopsAtTheFirstRefusal(t *testing.T) {
	probes := []rbac.Probe{
		{Verb: "create", Resource: "deployments", Namespace: "ns"},
		{Verb: "create", Resource: "secrets", Namespace: "ns"},
		{Verb: "create", Resource: "services", Namespace: "ns"},
	}
	f := &fakeRunner{outByCall: map[int]string{1: "no\n"}, errByCall: map[int]error{1: fmt.Errorf("exit status 1")}}
	_, err := Preflight(bg(), f, validate.PlatformKubernetes, "kubectl", false, probes, noApprovals)
	var refused *rbac.RefusedError
	if !errors.As(err, &refused) || refused.Probe.Resource != "secrets" {
		t.Fatalf("want a refusal of secrets, got %v", err)
	}
	if len(f.calls) != 2 {
		t.Errorf("want 2 probes asked (deployments, secrets), got %+v", f.calls)
	}
	f = &fakeRunner{}
	if _, err := Preflight(bg(), f, validate.PlatformKubernetes, "kubectl", false, probes, noApprovals); err != nil || len(f.calls) != 3 {
		t.Errorf("all allowed: err = %v, want 3 probes asked, got %+v", err, f.calls)
	}
}

// A probe whose answer is not a yes is never taken for one: a wrapper that changed
// can-i's output exits 0 without a verdict.
func TestPreflightRefusesAnAnswerThatIsNotYes(t *testing.T) {
	f := &fakeRunner{out: "ok\n"}
	if _, err := Preflight(bg(), f, validate.PlatformKubernetes, "kubectl", false, deployProbe, noApprovals); err == nil ||
		!strings.Contains(err.Error(), "could not read the answer") {
		t.Errorf("an exit 0 without yes must be refused, got %v", err)
	}
}

func TestPreflightRejectsDisallowedBinaryBeforeRunning(t *testing.T) {
	f := &fakeRunner{}
	if _, err := Preflight(bg(), f, validate.PlatformKubernetes, "curl", false, nil, noApprovals); err == nil {
		t.Fatal("a binary outside the platform allowlist must be rejected")
	}
	if len(f.calls) != 0 {
		t.Fatal("the probe must never run when the command itself is rejected")
	}
}

func TestPreflightExtraAllowedThreadsThrough(t *testing.T) {
	f := &fakeRunner{}
	if _, err := Preflight(bg(), f, validate.PlatformPodman, "lima podman", false, nil, noApprovals); err == nil {
		t.Fatal("lima podman must be rejected without an approval")
	}
	f = &fakeRunner{}
	if _, err := Preflight(bg(), f, validate.PlatformPodman, "lima podman", false, nil, approve(t, "lima")); err != nil {
		t.Fatalf("lima podman with lima approved unexpected error: %v", err)
	}
	want := []string{"lima", "podman", "info"}
	if !reflect.DeepEqual(f.calls[0].argv, want) {
		t.Errorf("argv = %v, want %v", f.calls[0].argv, want)
	}

	// sudo is refused before the probe runs, approvals or not.
	f = &fakeRunner{}
	if _, err := Preflight(bg(), f, validate.PlatformPodman, "sudo podman", false, nil, approve(t, "lima")); err == nil ||
		!strings.Contains(err.Error(), "privilege-escalation wrapper") {
		t.Fatalf("sudo podman must be refused as an escalation wrapper, got %v", err)
	}
	if len(f.calls) != 0 {
		t.Fatal("the probe must never run when the command itself is rejected")
	}
}

func TestKubernetesProbesUnknownAction(t *testing.T) {
	if probes, err := KubernetesProbes("rollback", "", nil, false); err == nil || len(probes) != 0 {
		t.Fatalf("unknown action must be rejected, got %+v, %v", probes, err)
	}
	if _, err := KubernetesProbes(ActionDeploy, "", []nsgate.Object{{Kind: "CronJob", Name: "x"}}, false); err == nil ||
		!strings.Contains(err.Error(), "CronJob") {
		t.Fatalf("a kind with no resource mapped must be an error naming it, got %v", err)
	}
}

// The probe set of a deploy and a remove is the plural resource of every object the
// manifest writes, once each, plus persistentvolumes (cluster-wide) for the libs
// claim's volume, and never namespaces: the namespace flow asks for that only when
// the namespace is absent.
func TestKubernetesProbesDeployAndRemoveSets(t *testing.T) {
	base := []nsgate.Object{{Kind: "Deployment", Name: "d"}, {Kind: "Service", Name: "d"}, {Kind: "ConfigMap", Name: "d-config"}}
	withSecrets := append(append([]nsgate.Object(nil), base...),
		nsgate.Object{Kind: "Secret", Name: "d-credentials"}, nsgate.Object{Kind: "Secret", Name: "d-stores"})
	withPVC := append(append([]nsgate.Object(nil), withSecrets...), nsgate.Object{Kind: "PersistentVolumeClaim", Name: "d-libs"})
	for _, c := range []struct {
		name    string
		written []nsgate.Object
		pv      bool
		want    []string
	}{
		{"no secret, no volume", base, false, []string{"deployments", "services", "configmaps"}},
		{"secrets are asked once", withSecrets, false, []string{"deployments", "services", "configmaps", "secrets"}},
		{"libs claim adds the claim and the volume", withPVC, true,
			[]string{"deployments", "services", "configmaps", "secrets", "persistentvolumeclaims", "persistentvolumes"}},
	} {
		for _, verb := range []struct{ action, can string }{{ActionDeploy, "create"}, {ActionRemove, "delete"}} {
			probes, err := KubernetesProbes(verb.action, "ns", c.written, c.pv)
			if err != nil {
				t.Fatalf("%s %s: %v", c.name, verb.action, err)
			}
			var got []string
			for _, p := range probes {
				got = append(got, p.Resource)
				if p.Verb != verb.can || p.Namespace != "ns" || p.Resource == "namespaces" {
					t.Errorf("%s %s: probe %+v, want verb %s in ns", c.name, verb.action, p, verb.can)
				}
			}
			if !reflect.DeepEqual(got, c.want) {
				t.Errorf("%s %s: resources = %v, want %v", c.name, verb.action, got, c.want)
			}
		}
	}
	// The volume is cluster-scoped, so rbac.CanI asks it of the whole cluster.
	probes, _ := KubernetesProbes(ActionDeploy, "ns", withPVC, true)
	argv := rbac.CanI([]string{"kubectl"}, probes[len(probes)-1]).Argv
	if want := []string{"kubectl", "auth", "can-i", "create", "persistentvolumes", "--all-namespaces"}; !reflect.DeepEqual(argv, want) {
		t.Errorf("persistentvolumes probe argv = %v, want %v", argv, want)
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
		if _, err := KubernetesPodsJSON(bg(), f, []string{"kubectl"}, c.namespace, c.selector, c.names, c.allNamespaces); err != nil {
			t.Fatalf("%s: %v", c.name, err)
		}
		if !reflect.DeepEqual(f.calls[0].argv, c.want) {
			t.Errorf("%s: argv = %v, want %v", c.name, f.calls[0].argv, c.want)
		}
	}
}

func TestKubernetesPodsJSONRunFailureWraps(t *testing.T) {
	f := &fakeRunner{err: fmt.Errorf("boom")}
	_, err := KubernetesPodsJSON(bg(), f, []string{"kubectl"}, "", "app=solmq-connector", nil, false)
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
		if _, err := KubernetesGetJSON(bg(), f, []string{"kubectl"}, c.namespace, c.kind, c.obj); err != nil {
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
	_, err := KubernetesGetJSON(bg(), f, []string{"kubectl"}, "solace", "secret", "nope")
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
		if _, err := KubernetesTop(bg(), f, []string{"kubectl"}, c.namespace, c.selector, c.names, c.allNamespaces); err != nil {
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
	_, err := KubernetesTop(bg(), f, []string{"kubectl"}, "", "app=x", nil, false)
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
	if _, err := EngineInspectJSON(bg(), f, []string{"docker"}, []string{"a", "b", "c"}); err != nil {
		t.Fatal(err)
	}
	want := []string{"docker", "container", "inspect", "a", "b", "c"}
	if !reflect.DeepEqual(f.calls[0].argv, want) {
		t.Errorf("argv = %v, want %v", f.calls[0].argv, want)
	}
	// A chained command keeps its own tokens ahead of the subcommand.
	f2 := &fakeRunner{}
	if _, err := EngineInspectJSON(bg(), f2, []string{"sudo", "podman"}, []string{"solmq"}); err != nil {
		t.Fatal(err)
	}
	want2 := []string{"sudo", "podman", "container", "inspect", "solmq"}
	if !reflect.DeepEqual(f2.calls[0].argv, want2) {
		t.Errorf("argv = %v, want %v", f2.calls[0].argv, want2)
	}
}

func TestEngineInspectJSONNoNamesIsAnErrorAndRunsNothing(t *testing.T) {
	f := &fakeRunner{}
	if _, err := EngineInspectJSON(bg(), f, []string{"docker"}, nil); err == nil {
		t.Fatal("inspecting nothing must be an error")
	}
	if len(f.calls) != 0 {
		t.Errorf("nothing must run, got %v", f.calls)
	}
}

func TestEngineInspectJSONFailureNamesTheTargets(t *testing.T) {
	f := &fakeRunner{err: fmt.Errorf("no such object")}
	_, err := EngineInspectJSON(bg(), f, []string{"docker"}, []string{"a", "b"})
	if err == nil {
		t.Fatal("a run failure must surface as an error")
	}
	if !strings.Contains(err.Error(), "a, b") {
		t.Errorf("error should name the targets, got %v", err)
	}
}

// What the engine said is appended to the error; a terminal escape in it is dropped,
// the line breaks are kept.
func TestEngineFailuresDropControlCharactersFromTheEnginesOutput(t *testing.T) {
	for name, call := range map[string]func(r gdurunner.Runner) error{
		"EngineInspectJSON": func(r gdurunner.Runner) error {
			_, err := EngineInspectJSON(bg(), r, []string{"docker"}, []string{"c"})
			return err
		},
		"EngineImageInspectJSON": func(r gdurunner.Runner) error {
			_, err := EngineImageInspectJSON(bg(), r, []string{"docker"}, "img:1")
			return err
		},
		"EngineStats": func(r gdurunner.Runner) error {
			_, err := EngineStats(bg(), r, []string{"docker"}, []string{"c"})
			return err
		},
		"EngineList": func(r gdurunner.Runner) error {
			_, err := EngineList(bg(), r, []string{"docker"}, false)
			return err
		},
	} {
		t.Run(name, func(t *testing.T) {
			r := &splitRunner{stderr: "no such\x1b[2J thing\nsecond", err: fmt.Errorf("exit status 1")}
			err := call(r)
			if err == nil || strings.Contains(err.Error(), "\x1b") || !strings.Contains(err.Error(), "no such[2J thing\nsecond") {
				t.Errorf("err = %q, want the output with the escape dropped and the line break kept", err)
			}
		})
	}
}

func TestEngineImageInspectJSONArgv(t *testing.T) {
	f := &fakeRunner{}
	if _, err := EngineImageInspectJSON(bg(), f, []string{"podman"}, "solace/x:2.14.1"); err != nil {
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
	if _, err := EngineStats(bg(), f, []string{"docker"}, []string{"a", "b"}); err != nil {
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
	if _, err := EngineList(bg(), f, []string{"podman"}, false); err != nil {
		t.Fatal(err)
	}
	want := []string{"podman", "ps", "--all", "--no-trunc", "--format", "{{.Names}}\t{{.Image}}\t{{.Status}}"}
	if !reflect.DeepEqual(f.calls[0].argv, want) {
		t.Errorf("argv = %v, want %v", f.calls[0].argv, want)
	}
	// Only the running ones, for the redeploy check.
	f = &fakeRunner{}
	if _, err := EngineList(bg(), f, []string{"docker"}, true); err != nil {
		t.Fatal(err)
	}
	want = []string{"docker", "ps", "--all", "--no-trunc", "--filter", "status=running", "--format", "{{.Names}}\t{{.Image}}\t{{.Status}}"}
	if !reflect.DeepEqual(f.calls[0].argv, want) {
		t.Errorf("running argv = %v, want %v", f.calls[0].argv, want)
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
			// A count below zero is no count at all.
			name: "negative count is refused", scope: QuadletScope{}, out: "-1\n", wantErr: true,
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
			got, err := SystemctlNRestarts(bg(), f, c.scope, "solmq-connector.service")
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
		if _, err := ScriptInstalled(bg(), f, c.cmd, c.platform, c.target, c.namespace, "/tmp/solmq-status.sh"); err != nil {
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
		got, err := ScriptInstalled(bg(), f, []string{"kubectl"}, validate.PlatformKubernetes, "pod-0", "", "/tmp/solmq-status.sh")
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
		got, err := ScriptInstalled(bg(), f, []string{"docker"}, validate.PlatformDocker, "solmq-connector", "", "/tmp/solmq-status.sh")
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
	if _, err := ScriptInstalled(bg(), f, []string{"kubectl"}, "bogus", "pod-0", "", "/tmp/solmq-status.sh"); err == nil {
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
			want:   []string{"kubectl", "exec", "-i", "pod-0", "-c", spec.ConnectorContainerName, "--", "sh", "-c", "mkdir -p '/tmp' && cat > '/tmp/solmq-status.sh'"},
		},
		{
			name: "kubernetes with namespace", cmd: []string{"oc"}, platform: validate.PlatformKubernetes,
			target: "pod-0", namespace: "solace",
			want: []string{"oc", "exec", "-i", "pod-0", "-n", "solace", "-c", spec.ConnectorContainerName, "--", "sh", "-c", "mkdir -p '/tmp' && cat > '/tmp/solmq-status.sh'"},
		},
		{
			name: "docker", cmd: []string{"docker"}, platform: validate.PlatformDocker,
			target: "solmq-connector",
			want:   []string{"docker", "exec", "-i", "solmq-connector", "sh", "-c", "mkdir -p '/tmp' && cat > '/tmp/solmq-status.sh'"},
		},
		{
			name: "podman", cmd: []string{"podman"}, platform: validate.PlatformPodman,
			target: "solmq-connector",
			want:   []string{"podman", "exec", "-i", "solmq-connector", "sh", "-c", "mkdir -p '/tmp' && cat > '/tmp/solmq-status.sh'"},
		},
	}
	for _, c := range cases {
		f := &fakeRunner{}
		if _, err := InstallScript(bg(), f, c.cmd, c.platform, c.target, c.namespace, "/tmp/solmq-status.sh", "#!/bin/sh\necho ok\n"); err != nil {
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
	if _, err := InstallScript(bg(), f, []string{"docker"}, validate.PlatformDocker, "solmq-connector", "", "/tmp/solmq-status.sh", script); err != nil {
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

// An engine upload with nothing to name is refused by containercli.Upload before
// anything runs.
func TestInstallScriptRefusesAnEmptyTargetOnAnEngine(t *testing.T) {
	f := &fakeRunner{}
	if _, err := InstallScript(bg(), f, []string{"docker"}, validate.PlatformDocker, "", "", "/tmp/x.sh", "x"); err == nil {
		t.Fatal("an upload with no container must be refused")
	}
	if len(f.calls) != 0 {
		t.Fatal("nothing must run")
	}
}

func TestInstallScriptUnknownPlatform(t *testing.T) {
	f := &fakeRunner{}
	if _, err := InstallScript(bg(), f, []string{"kubectl"}, "bogus", "pod-0", "", "/tmp/x.sh", "x"); err == nil {
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
		if _, err := RunStatusScript(bg(), f, c.cmd, c.platform, c.target, c.namespace, "/tmp/solmq-status.sh"); err != nil {
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
	out, err := RunStatusScript(bg(), f, []string{"kubectl"}, validate.PlatformKubernetes, "pod-0", "", "/tmp/solmq-status.sh")
	if err == nil {
		t.Fatal("a non-zero script exit is expected and must be returned, not swallowed")
	}
	if out != "STANDBY\n" {
		t.Errorf("output = %q, want %q alongside the error", out, "STANDBY\n")
	}
}

func TestRunStatusScriptUnknownPlatform(t *testing.T) {
	f := &fakeRunner{}
	if _, err := RunStatusScript(bg(), f, []string{"kubectl"}, "bogus", "pod-0", "", "/tmp/x.sh"); err == nil {
		t.Fatal("unknown platform must be rejected")
	}
	if len(f.calls) != 0 {
		t.Fatal("nothing must run for an unknown platform")
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
		// containercli.NoPrevious names the engine as a message spells it (Docker).
		if !strings.Contains(strings.ToLower(err.Error()), platform) || !strings.Contains(err.Error(), "--previous is a kubernetes option") {
			t.Errorf("%s: the error should name the platform and the option, got %v", platform, err)
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
			// The in-container command is the caller's, appended after the flags
			// (and after the "--" on kubernetes).
			got, err := ExecArgv([]string{"cli"}, c.platform, "pod-0", c.namespace, c.opts, "id", "-u")
			if err != nil {
				t.Fatalf("ExecArgv: %v", err)
			}
			if want := append(append([]string(nil), c.want...), "id", "-u"); !reflect.DeepEqual(got, want) {
				t.Errorf("argv = %v, want %v", got, want)
			}
		})
	}
}

// An exec with nothing to run is a caller mistake on every platform, refused rather
// than built.
func TestExecArgvRefusesAnEmptyCommand(t *testing.T) {
	for _, platform := range []string{validate.PlatformKubernetes, validate.PlatformDocker, validate.PlatformPodman} {
		if _, err := ExecArgv([]string{"cli"}, platform, "pod-0", "", ExecOpts{}); err == nil {
			t.Errorf("%s: an exec with no command must be refused", platform)
		}
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

// ---- runParsed (the parse-safety boundary) -------------------------------------

// splitRunner is a fake that writes its two scripted streams to the two writers
// a caller hands it, so the helpers can be tested on the path production takes.
// A helper that merged the streams (Combined, rather than Split) would parse the
// warning along with the payload, and the test says so.
type splitRunner struct {
	calls  []call
	stdout string
	stderr string
	err    error
}

func (s *splitRunner) Run(_ context.Context, c gdurunner.Cmd, stdout, stderr io.Writer) error {
	s.calls = append(s.calls, call{argv: c.Argv, stdin: string(c.Stdin), env: c.Env})
	emit(stdout, s.stdout)
	emit(stderr, s.stderr)
	return s.err
}

// Attach is never reached: nothing in this package attaches a terminal.
func (s *splitRunner) Attach(c gdurunner.Cmd, _, _, _ *os.File) (int, error) {
	return 0, fmt.Errorf("%s: this fake does not attach", c.Argv[0])
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
		call func(r gdurunner.Runner) (string, error)
	}{
		{"KubernetesPodsJSON", func(r gdurunner.Runner) (string, error) {
			return KubernetesPodsJSON(bg(), r, []string{"oc"}, "ns", "", nil, false)
		}},
		{"KubernetesGetJSON", func(r gdurunner.Runner) (string, error) {
			return KubernetesGetJSON(bg(), r, []string{"oc"}, "ns", "deployment", "d")
		}},
		{"KubernetesListJSON", func(r gdurunner.Runner) (string, error) {
			return KubernetesListJSON(bg(), r, []string{"oc"}, "ns")
		}},
		{"KubernetesTop", func(r gdurunner.Runner) (string, error) {
			return KubernetesTop(bg(), r, []string{"oc"}, "ns", "", nil, false)
		}},
		{"EngineInspectJSON", func(r gdurunner.Runner) (string, error) {
			return EngineInspectJSON(bg(), r, []string{"docker"}, []string{"c"})
		}},
		{"EngineImageInspectJSON", func(r gdurunner.Runner) (string, error) {
			return EngineImageInspectJSON(bg(), r, []string{"docker"}, "img:1")
		}},
		{"EngineStats", func(r gdurunner.Runner) (string, error) {
			return EngineStats(bg(), r, []string{"docker"}, []string{"c"})
		}},
		{"EngineList", func(r gdurunner.Runner) (string, error) {
			return EngineList(bg(), r, []string{"docker"}, false)
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

	// DockerImageCommand parses two JSON lines; a warning on stderr made the
	// line count wrong and the image check pass without comparing anything.
	t.Run("DockerImageCommand", func(t *testing.T) {
		r := &splitRunner{stdout: "[\"java\"]\n[\"--x\"]\n", stderr: warning}
		ep, cmd, found, err := DockerImageCommand(bg(), r, "docker", "img:1", noApprovals)
		if err != nil || !found || !reflect.DeepEqual(ep, []string{"java"}) || !reflect.DeepEqual(cmd, []string{"--x"}) {
			t.Errorf("ep=%q cmd=%q found=%v err=%v, want the stdout answer despite the warning", ep, cmd, found, err)
		}
	})

	// SystemctlNRestarts parses an integer, so the same warning would turn a
	// restart count into an error. It answers a number rather than a string,
	// hence its own case.
	t.Run("SystemctlNRestarts", func(t *testing.T) {
		r := &splitRunner{stdout: "3\n", stderr: warning}
		n, err := SystemctlNRestarts(bg(), r, sc, "x.service")
		if err != nil {
			t.Fatalf("SystemctlNRestarts: %v", err)
		}
		if n != 3 {
			t.Errorf("restarts = %d, want 3", n)
		}
	})
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
	_, err := KubernetesListJSON(bg(), r, []string{"oc"}, "ns")
	if err == nil {
		t.Fatal("a failed command must be reported")
	}
	for _, want := range []string{"Forbidden", "partial-output"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("the error should carry %q, got %v", want, err)
		}
	}
}

// TestHelpersPassTheCallersContextToTheRunner pins that every helper hands the
// caller's context to the shared runner rather than minting its own: a context
// that is already done starts nothing, whether the helper captures combined
// (Combined), keeps the streams apart (runParsed) or runs several commands in a
// row. runnertest.Fake refuses a done context before it records a call, so an
// empty call list is the proof.
func TestHelpersPassTheCallersContextToTheRunner(t *testing.T) {
	done, cancel := context.WithCancel(context.Background())
	cancel()
	for _, c := range []struct {
		name string
		call func(r gdurunner.Runner) error
	}{
		{"Kubernetes (combined)", func(r gdurunner.Runner) error {
			_, err := Kubernetes(done, r, "kubectl", ActionDeploy, "x", noApprovals)
			return err
		}},
		{"KubernetesGetJSON (split)", func(r gdurunner.Runner) error {
			_, err := KubernetesGetJSON(done, r, []string{"kubectl"}, "ns", "secret", "s")
			return err
		}},
		{"PodmanDeploy (several commands)", func(r gdurunner.Runner) error {
			// The state probe's error is ignored, so the service reads as not
			// running and the write and the reload follow; the reload is what the
			// done context stops.
			d := &deployFns{answer: true}
			_, _, err := PodmanDeploy(done, r, QuadletScope{}, []string{"a.service"}, true, d.approve, d.write)
			return err
		}},
		{"Preflight", func(r gdurunner.Runner) error {
			_, err := Preflight(done, r, validate.PlatformDocker, "docker", false, nil, noApprovals)
			return err
		}},
	} {
		t.Run(c.name, func(t *testing.T) {
			f := &runnertest.Fake{}
			err := c.call(f)
			if !errors.Is(err, context.Canceled) {
				t.Errorf("err = %v, want it to wrap context.Canceled", err)
			}
			if n := len(f.Calls()); n != 0 {
				t.Errorf("a done context must start nothing, got %d calls", n)
			}
		})
	}
}

// DockerImageCommand reads the image's entrypoint and command through the
// engine, one inspect call; an engine that cannot answer -- the image not
// pulled yet -- or an answer that does not parse leaves nothing to compare,
// and an unsafe command never runs.
func TestDockerImageCommand(t *testing.T) {
	f := &fakeRunner{out: "[\"java\",\"-cp\",\"@/app/jib-classpath-file\"]\n[\"--x=1\"]\n"}
	ep, cmd, found, err := DockerImageCommand(bg(), f, "docker", "img:1", noApprovals)
	if err != nil || !found || !reflect.DeepEqual(ep, []string{"java", "-cp", "@/app/jib-classpath-file"}) || !reflect.DeepEqual(cmd, []string{"--x=1"}) {
		t.Fatalf("ep=%q cmd=%q found=%v err=%v", ep, cmd, found, err)
	}
	want := []string{"docker", "image", "inspect", "--format", "{{json .Config.Entrypoint}}\n{{json .Config.Cmd}}", "img:1"}
	if len(f.calls) != 1 || !reflect.DeepEqual(f.calls[0].argv, want) {
		t.Errorf("argv = %v, want %v", f.calls, want)
	}
	for _, c := range []struct {
		name string
		f    *fakeRunner
	}{
		{"not pulled", &fakeRunner{err: fmt.Errorf("Error: No such image: img:1")}},
		{"one line", &fakeRunner{out: "[\"java\"]"}},
		{"not json", &fakeRunner{out: "out\nout"}},
	} {
		if _, _, found, err := DockerImageCommand(bg(), c.f, "docker", "img:1", noApprovals); found || err != nil {
			t.Errorf("%s: found=%v err=%v, want nothing to compare", c.name, found, err)
		}
	}
	unsafe := &fakeRunner{}
	if _, _, _, err := DockerImageCommand(bg(), unsafe, "docker; rm -rf /", "img:1", noApprovals); err == nil || len(unsafe.calls) != 0 {
		t.Errorf("an unsafe command must be rejected before running: err=%v calls=%d", err, len(unsafe.calls))
	}
}

// DockerImagePull pulls through the configured engine command, and refuses an
// unsafe one before running anything.
func TestDockerImagePull(t *testing.T) {
	f := &fakeRunner{out: "pulled"}
	if out, err := DockerImagePull(bg(), f, "docker --context remote", "img:1", noApprovals); err != nil || out != "pulled" {
		t.Fatalf("out=%q err=%v", out, err)
	}
	if want := []string{"docker", "--context", "remote", "image", "pull", "img:1"}; len(f.calls) != 1 || !reflect.DeepEqual(f.calls[0].argv, want) {
		t.Errorf("argv = %v, want %v", f.calls, want)
	}
	unsafe := &fakeRunner{}
	if _, err := DockerImagePull(bg(), unsafe, "docker; rm -rf /", "img:1", noApprovals); err == nil || len(unsafe.calls) != 0 {
		t.Errorf("an unsafe command must be rejected before running: err=%v calls=%d", err, len(unsafe.calls))
	}
}

// The lifecycle helpers act on what deploy created, in place, through the
// configured command: kubernetes scales the Deployment to 0 and back to its
// replicas, or rolls it; docker starts, stops or restarts the container;
// podman does the same to the unit through systemctl, in the unit's scope.
func TestLifecycleArgv(t *testing.T) {
	for _, c := range []struct {
		name string
		run  func(r gdurunner.Runner) (string, error)
		want []string
	}{
		{"kubernetes stop", func(r gdurunner.Runner) (string, error) {
			return KubernetesLifecycle(bg(), r, "kubectl", ActionStop, "ns", "c", 3, noApprovals)
		}, []string{"kubectl", "scale", "deployment/c", "-n", "ns", "--replicas=0"}},
		{"kubernetes start", func(r gdurunner.Runner) (string, error) {
			return KubernetesLifecycle(bg(), r, "oc", ActionStart, "ns", "c", 3, noApprovals)
		}, []string{"oc", "scale", "deployment/c", "-n", "ns", "--replicas=3"}},
		{"kubernetes restart, current namespace", func(r gdurunner.Runner) (string, error) {
			return KubernetesLifecycle(bg(), r, "kubectl", ActionRestart, "", "c", 1, noApprovals)
		}, []string{"kubectl", "rollout", "restart", "deployment/c"}},
		{"kubernetes restart, in a namespace", func(r gdurunner.Runner) (string, error) {
			return KubernetesLifecycle(bg(), r, "kubectl", ActionRestart, "ns", "c", 1, noApprovals)
		}, []string{"kubectl", "rollout", "restart", "deployment/c", "-n", "ns"}},
		{"kubernetes rollout status", func(r gdurunner.Runner) (string, error) {
			return KubernetesRolloutStatus(bg(), r, "kubectl", "ns", "c", noApprovals)
		}, []string{"kubectl", "rollout", "status", "deployment/c", "-n", "ns", "--timeout=300s"}},
		{"docker stop", func(r gdurunner.Runner) (string, error) {
			return DockerLifecycle(bg(), r, "docker --context prod", ActionStop, "connector", noApprovals)
		}, []string{"docker", "--context", "prod", "stop", "connector"}},
		{"docker start", func(r gdurunner.Runner) (string, error) {
			return DockerLifecycle(bg(), r, "docker", ActionStart, "connector", noApprovals)
		}, []string{"docker", "start", "connector"}},
		{"docker restart", func(r gdurunner.Runner) (string, error) {
			return DockerLifecycle(bg(), r, "docker", ActionRestart, "connector", noApprovals)
		}, []string{"docker", "restart", "connector"}},
		{"podman start, user unit", func(r gdurunner.Runner) (string, error) {
			return PodmanLifecycle(bg(), r, QuadletScope{UserMode: true}, ActionStart, "connector.service")
		}, []string{"systemctl", "--user", "start", "connector.service"}},
		{"podman stop, system unit", func(r gdurunner.Runner) (string, error) {
			return PodmanLifecycle(bg(), r, QuadletScope{}, ActionStop, "connector.service")
		}, []string{"systemctl", "stop", "connector.service"}},
	} {
		f := &fakeRunner{}
		if _, err := c.run(f); err != nil {
			t.Errorf("%s: %v", c.name, err)
			continue
		}
		if len(f.calls) != 1 || !reflect.DeepEqual(f.calls[0].argv, c.want) {
			t.Errorf("%s: argv = %+v, want %v", c.name, f.calls, c.want)
		}
	}
	for _, bad := range []func(r gdurunner.Runner) (string, error){
		func(r gdurunner.Runner) (string, error) {
			return KubernetesLifecycle(bg(), r, "kubectl", ActionDeploy, "ns", "c", 1, noApprovals)
		},
		func(r gdurunner.Runner) (string, error) {
			return DockerLifecycle(bg(), r, "docker", "kill", "c", noApprovals)
		},
		func(r gdurunner.Runner) (string, error) {
			return PodmanLifecycle(bg(), r, QuadletScope{}, ActionRemove, "c.service")
		},
		func(r gdurunner.Runner) (string, error) {
			return DockerLifecycle(bg(), r, "docker; rm -rf /", ActionStop, "c", noApprovals)
		},
	} {
		f := &fakeRunner{}
		if _, err := bad(f); err == nil || len(f.calls) != 0 {
			t.Errorf("an unknown action or unsafe command must fail before running: err %v, calls %+v", err, f.calls)
		}
	}
}

// The lifecycle verbs' kubernetes preflight asks to patch the Deployment,
// which scale and rollout restart both do.
func TestPreflightKubernetesLifecycleAsksToPatch(t *testing.T) {
	for _, action := range []string{ActionStart, ActionStop, ActionRestart} {
		probes, err := KubernetesProbes(action, "ns", nil, false)
		if err != nil {
			t.Fatalf("%s: %v", action, err)
		}
		f := &fakeRunner{}
		if _, err := Preflight(bg(), f, validate.PlatformKubernetes, "kubectl", false, probes, noApprovals); err != nil {
			t.Fatalf("%s: %v", action, err)
		}
		if want := []string{"kubectl", "auth", "can-i", "patch", "deployments", "-n", "ns"}; len(f.calls) != 1 || !reflect.DeepEqual(f.calls[0].argv, want) {
			t.Errorf("%s: calls = %+v, want one with argv %v", action, f.calls, want)
		}
	}
}

// The namespace read is stdout alone and never taken for an answer it is not: exactly
// namespace/<ns> exists, nothing is absent, anything else and any failure is unknown.
func TestKubernetesNamespaceState(t *testing.T) {
	for _, c := range []struct {
		name string
		out  string
		err  error
		want kube.NamespaceState
	}{
		{"exists", "namespace/ns\n", nil, kube.NamespaceExists},
		{"absent", "", nil, kube.NamespaceAbsent},
		{"somebody else's", "namespace/other\n", nil, kube.NamespaceUnknown},
		{"read refused", "namespace/ns", fmt.Errorf("exit status 1"), kube.NamespaceUnknown},
	} {
		f := &fakeRunner{outByCall: map[int]string{0: c.out}, err: c.err}
		if got := KubernetesNamespaceState(bg(), f, []string{"kubectl"}, "ns"); got != c.want {
			t.Errorf("%s: state = %v, want %v", c.name, got, c.want)
		}
		if want := []string{"kubectl", "get", "namespace", "ns", "--ignore-not-found", "-o", "name"}; !reflect.DeepEqual(f.calls[0].argv, want) {
			t.Errorf("%s: argv = %v, want %v", c.name, f.calls[0].argv, want)
		}
	}
}

// The namespace gate's two calls: nsgate.List's one query, and kube.DeleteNamespace.
func TestKubernetesNamespaceGateArgv(t *testing.T) {
	f := &fakeRunner{}
	if _, err := KubernetesListJSON(bg(), f, []string{"oc", "--context", "p"}, "ns"); err != nil {
		t.Fatal(err)
	}
	if want := []string{"oc", "--context", "p", "get", "all,persistentvolumeclaims,secrets,configmaps", "-n", "ns", "-o", "json"}; !reflect.DeepEqual(f.calls[0].argv, want) {
		t.Errorf("list argv = %v, want %v", f.calls[0].argv, want)
	}
	if _, err := KubernetesDeleteNamespace(bg(), f, []string{"oc"}, "ns"); err != nil {
		t.Fatal(err)
	}
	if want := []string{"oc", "delete", "namespace", "ns", "--ignore-not-found"}; !reflect.DeepEqual(f.calls[1].argv, want) {
		t.Errorf("delete argv = %v, want %v", f.calls[1].argv, want)
	}
}

// CheckProbe reads the verdict from stdout alone, so a warning kubectl prints on
// stderr after it cannot hide it.
func TestCheckProbeReadsStdoutAlone(t *testing.T) {
	r := &splitRunner{stdout: "yes\n", stderr: "Warning: a deprecated flag\n"}
	if err := CheckProbe(bg(), r, []string{"kubectl"}, rbac.Probe{Verb: "create", Resource: "secrets"}); err != nil {
		t.Errorf("a yes with a warning on stderr must pass, got %v", err)
	}
}

// ---- redeploy probes: DockerDeployDigest, PodmanSecretMatches -------------------------

// routeRunner answers each call from a function of its argv, on stdout, so a test
// can give `ps` and `inspect` different answers. It records every argv.
type routeRunner struct {
	calls [][]string
	reply func(argv []string) (string, error)
}

func (r *routeRunner) Run(_ context.Context, c gdurunner.Cmd, stdout, _ io.Writer) error {
	r.calls = append(r.calls, c.Argv)
	out, err := r.reply(c.Argv)
	emit(stdout, out)
	return err
}

func (r *routeRunner) Attach(c gdurunner.Cmd, _, _, _ *os.File) (int, error) {
	return 0, fmt.Errorf("%s: this fake does not attach", c.Argv[0])
}

const deployKey = "solace-connector/deploy-sha256"

// DockerDeployDigest asks `ps` for the running containers and matches the name
// exactly in Go, so a sibling such as solmq-connector-2 is not solmq-connector; then,
// for a running container, reads the deploy digest label off its inspect.
func TestDockerDeployDigest(t *testing.T) {
	inspectDoc := func(name, label string) string {
		return `[{"Name":"/` + name + `","State":{"Status":"running"},"Config":{"Image":"i","Labels":{` + label + `}}}]`
	}
	for _, c := range []struct {
		name        string
		ps          string
		psErr       error
		inspect     string
		inspectErr  error
		wantRunning bool
		wantDigest  string
		wantErr     bool
	}{
		{name: "not running", ps: "other\timg\tUp 1 hour\n"},
		{name: "a sibling is not the container", ps: "solmq-connector-2\timg\tUp 1 hour\n"},
		{name: "running, label present", ps: "solmq-connector\timg\tUp 1 hour\n",
			inspect: inspectDoc("solmq-connector", `"`+deployKey+`":"abc"`), wantRunning: true, wantDigest: "abc"},
		{name: "running, label missing", ps: "solmq-connector\timg\tUp 1 hour\n",
			inspect: inspectDoc("solmq-connector", `"other":"x"`), wantRunning: true},
		{name: "running, inspect fails", ps: "solmq-connector\timg\tUp 1 hour\n",
			inspectErr: fmt.Errorf("boom"), wantRunning: true},
		{name: "running, inspect is not json", ps: "solmq-connector\timg\tUp 1 hour\n",
			inspect: "not json", wantRunning: true},
		{name: "running, inspect answers another name", ps: "solmq-connector\timg\tUp 1 hour\n",
			inspect: inspectDoc("elsewhere", `"`+deployKey+`":"abc"`), wantRunning: true},
		{name: "the listing fails", psErr: fmt.Errorf("daemon down"), wantErr: true},
	} {
		t.Run(c.name, func(t *testing.T) {
			r := &routeRunner{reply: func(argv []string) (string, error) {
				if contains(argv, "ps") {
					return c.ps, c.psErr
				}
				return c.inspect, c.inspectErr
			}}
			running, digest, err := DockerDeployDigest(bg(), r, "docker", "solmq-connector", deployKey, noApprovals)
			if (err != nil) != c.wantErr {
				t.Fatalf("err = %v, want error = %v", err, c.wantErr)
			}
			if c.wantErr {
				if !strings.Contains(err.Error(), "solmq-connector") || !strings.Contains(err.Error(), "daemon down") {
					t.Errorf("the error should name the container and carry the cause, got %v", err)
				}
				return
			}
			if running != c.wantRunning || digest != c.wantDigest {
				t.Errorf("running = %v, digest = %q; want %v, %q", running, digest, c.wantRunning, c.wantDigest)
			}
			wantPS := []string{"docker", "ps", "--all", "--no-trunc", "--filter", "status=running", "--format", "{{.Names}}\t{{.Image}}\t{{.Status}}"}
			if !reflect.DeepEqual(r.calls[0], wantPS) {
				t.Errorf("listing argv = %v, want %v", r.calls[0], wantPS)
			}
			if c.wantRunning && !reflect.DeepEqual(r.calls[1], []string{"docker", "container", "inspect", "solmq-connector"}) {
				t.Errorf("inspect argv = %v", r.calls[1])
			}
			if !c.wantRunning && len(r.calls) != 1 {
				t.Errorf("a container that is not running is not inspected, got %v", r.calls)
			}
		})
	}
	if _, _, err := DockerDeployDigest(bg(), &routeRunner{reply: nil}, "docker; id", "c", deployKey, noApprovals); err == nil {
		t.Error("an unsafe command is refused before anything runs")
	}
}

// PodmanSecretMatches reads the stored value back from stdout alone and compares it
// exactly. Anything that is not a clear match reads as false: the caller counts that
// as a change. The value never appears in an argv.
func TestPodmanSecretMatches(t *testing.T) {
	const value = "s3cr3t"
	r := &splitRunner{stdout: value + "\n", stderr: "WARN[0000] a warning\n"}
	if !PodmanSecretMatches(bg(), r, "podman", "solmq-conn-app", value, noApprovals) {
		t.Error("the stored value equal to the new one matches, despite a warning on stderr")
	}
	want := []string{"podman", "secret", "inspect", "--showsecret", "--format", "{{.SecretData}}", "solmq-conn-app"}
	if !reflect.DeepEqual(r.calls[0].argv, want) {
		t.Errorf("argv = %v, want %v", r.calls[0].argv, want)
	}
	for _, a := range r.calls[0].argv {
		if strings.Contains(a, value) {
			t.Errorf("the value reached argv: %q", a)
		}
	}
	for name, rr := range map[string]*splitRunner{
		"a different value":   {stdout: "other\n"},
		"no trailing newline": {stdout: value},
		"an extra newline":    {stdout: value + "\n\n"},
		"a failed probe":      {stdout: value + "\n", err: fmt.Errorf("no such secret")},
		"nothing printed":     {},
	} {
		if PodmanSecretMatches(bg(), rr, "podman", "n", value, noApprovals) {
			t.Errorf("%s must read as a change", name)
		}
	}
	unsafe := &splitRunner{stdout: value + "\n"}
	if PodmanSecretMatches(bg(), unsafe, "podman; id", "n", value, noApprovals) || len(unsafe.calls) != 0 {
		t.Errorf("an unsafe command runs nothing: calls=%d", len(unsafe.calls))
	}
	if PodmanSecretMatches(bg(), unsafe, "podman", "", value, noApprovals) || len(unsafe.calls) != 0 {
		t.Errorf("an empty name runs nothing: calls=%d", len(unsafe.calls))
	}
}
