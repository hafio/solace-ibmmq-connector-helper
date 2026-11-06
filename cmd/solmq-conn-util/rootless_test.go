package main

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"

	gdurunner "github.com/hafio/go-deploy-utility/runner"
)

// probeRunner answers the probes a rootless podman deploy makes on a Linux host,
// by what each command is rather than by its position: loginctl's linger state,
// the enable call, and podman's uid and gid maps. Every other command succeeds
// with no output. The zero answers describe a healthy host.
type probeRunner struct {
	noAttach
	calls []fakeCall

	linger    string // Linger= value; "" answers "yes"
	lingerErr error  // fails show-user
	enableErr error  // fails enable-linger
	uidMap    string // "" answers "0:1 1:65536 "
	gidMap    string
	noMaps    bool // answer the maps with nothing at all
	mapErr    error
}

func (p *probeRunner) Run(_ context.Context, c gdurunner.Cmd, stdout, _ io.Writer) error {
	p.calls = append(p.calls, fakeCall{argv: c.Argv, stdin: string(c.Stdin), env: c.Env})
	joined := strings.Join(c.Argv, " ")
	stock := func(s string) string {
		if p.noMaps {
			return ""
		}
		if s == "" {
			return "0:1 1:65536 "
		}
		return s
	}
	switch {
	case c.Argv[0] == "loginctl" && c.Argv[1] == "show-user":
		state := p.linger
		if state == "" {
			state = "yes"
		}
		scriptOut(stdout, "Linger="+state+"\n")
		return p.lingerErr
	case c.Argv[0] == "loginctl" && c.Argv[1] == "enable-linger":
		return p.enableErr
	case strings.Contains(joined, "UIDMap"):
		scriptOut(stdout, stock(p.uidMap))
		return p.mapErr
	case strings.Contains(joined, "GIDMap"):
		scriptOut(stdout, stock(p.gidMap))
		return p.mapErr
	}
	return nil
}

// ran reports whether any command began with prefix.
func (p *probeRunner) ran(prefix ...string) bool {
	return slices.ContainsFunc(p.calls, func(c fakeCall) bool {
		return len(c.argv) >= len(prefix) && reflect.DeepEqual(c.argv[:len(prefix)], prefix)
	})
}

// idProbes counts the podman info calls that asked for an id map.
func (p *probeRunner) idProbes() int {
	n := 0
	for _, c := range p.calls {
		if j := strings.Join(c.argv, " "); strings.Contains(j, "UIDMap") || strings.Contains(j, "GIDMap") {
			n++
		}
	}
	return n
}

// linuxHost is testHost on a Linux host whose session was stripped (as sudo or
// cron does), so the scope derives and carries it, with the given subid files.
func linuxHost(home string, files map[string]string) hostFacts {
	h := testHost(home)
	h.scope.Linux = true
	h.readFile = fakeHostFiles(files)
	return h
}

// derivedSession is the environment a stripped session for uid 1000 gets.
var derivedSession = []string{"XDG_RUNTIME_DIR=/run/user/1000", "DBUS_SESSION_BUS_ADDRESS=unix:path=/run/user/1000/bus"}

// runPodmanVerb runs verb against a podman env.yaml made of section and reports
// the exit code and stderr.
func runPodmanVerb(t *testing.T, verb, section string, r gdurunner.Runner) (int, string) {
	t.Helper()
	dir := t.TempDir()
	write(t, dir, "env.yaml", sharedEnv+section)
	write(t, dir, "10.yaml", validWF)
	args := []string{verb, "-e", filepath.Join(dir, "env.yaml")}
	if verb != "validate" { // validate reads every section; it takes no --platform
		args = append(args, "--platform", "podman")
	}
	if verb == "remove" { // remove asks first; this is about what follows the answer
		withPrompter(t, true, "y\n")
	}
	var code int
	stderr := captureStderr(t, func() { code = dispatch(args, r) })
	return code, stderr
}

func unitWritten(quadletDir string) bool {
	_, err := os.Stat(filepath.Join(quadletDir, "solmq-conn.container"))
	return err == nil
}

// A rootless deploy on Linux asks logind about linger after the preflight and
// before it writes anything: a user that lingers is left alone, one that does
// not is enabled for its own account, a refused enable names what an
// administrator runs, and a probe that cannot run is an error. Every loginctl and
// systemctl call carries the repaired session, and no podman call does.
func TestDeployPodmanRootlessLinger(t *testing.T) {
	for _, c := range []struct {
		name       string
		r          *probeRunner
		wantCode   int
		wantErr    []string
		wantEnable bool
		wantUnit   bool
	}{
		{"already lingers: no enable call", &probeRunner{}, 0, nil, false, true},
		{"off: enabled for the user's own account", &probeRunner{linger: "no"}, 0,
			[]string{"note: enabled linger for alice"}, true, true},
		{"enable refused: the administrator form", &probeRunner{linger: "no", enableErr: errors.New("exit status 1")}, 1,
			[]string{"linger is not enabled for alice (Linger=no)", "sudo loginctl enable-linger alice"}, true, false},
		{"probe fails: an error naming the command", &probeRunner{lingerErr: errors.New("exit status 1")}, 1,
			[]string{"cannot read the linger state with `loginctl show-user 1000 --property=Linger`"}, false, false},
	} {
		t.Run(c.name, func(t *testing.T) {
			home := t.TempDir()
			useHost(t, linuxHost(home, nil))
			code, stderr := runPodmanVerb(t, "deploy", podmanEnv(), c.r)
			if code != c.wantCode {
				t.Fatalf("exit=%d, want %d\n%s", code, c.wantCode, stderr)
			}
			for _, w := range c.wantErr {
				if !strings.Contains(stderr, w) {
					t.Errorf("stderr should carry %q, got:\n%s", w, stderr)
				}
			}
			if got := c.r.ran("loginctl", "enable-linger"); got != c.wantEnable {
				t.Errorf("enable-linger called = %v, want %v", got, c.wantEnable)
			}
			if got := unitWritten(filepath.Join(home, ".config", "containers", "systemd")); got != c.wantUnit {
				t.Errorf("unit written = %v, want %v", got, c.wantUnit)
			}
			if !c.wantUnit && c.r.ran("podman", "secret") {
				t.Errorf("nothing may be stored once linger is refused, got %+v", c.r.calls)
			}
			for i, call := range c.r.calls {
				switch call.argv[0] {
				case "loginctl", "systemctl":
					if !reflect.DeepEqual(call.env, derivedSession) {
						t.Errorf("call %d %v env = %v, want the session %v", i, call.argv, call.env, derivedSession)
					}
				default:
					if call.env != nil {
						t.Errorf("call %d %v must not carry the session, got %v", i, call.argv, call.env)
					}
				}
			}
			// Linger is settled after the preflight and before the first secret.
			if c.r.calls[0].argv[0] != "podman" || c.r.calls[1].argv[0] != "loginctl" {
				t.Errorf("want preflight then the linger probe, got %+v", c.r.calls)
			}
		})
	}
}

// validate changes nothing, so a user that does not linger is a host error
// (exit 1) and so is a probe that cannot run; enabling is deploy's.
func TestValidatePodmanRefusesWithoutLinger(t *testing.T) {
	for _, c := range []struct {
		name string
		r    *probeRunner
		want string
	}{
		{"linger off", &probeRunner{linger: "no"}, "linger is not enabled for uid 1000 (Linger=no)"},
		{"probe fails", &probeRunner{lingerErr: errors.New("exit status 1")}, "cannot read the linger state"},
	} {
		t.Run(c.name, func(t *testing.T) {
			useHost(t, linuxHost(t.TempDir(), nil))
			code, stderr := runPodmanVerb(t, "validate", podmanEnv(), c.r)
			if code != 1 || !strings.Contains(stderr, c.want) {
				t.Errorf("exit=%d, want 1 naming %q:\n%s", code, c.want, stderr)
			}
			if c.r.ran("loginctl", "enable-linger") {
				t.Error("validate must not enable linger")
			}
		})
	}
	useHost(t, linuxHost(t.TempDir(), nil))
	r := &probeRunner{}
	if code, stderr := runPodmanVerb(t, "validate", podmanEnv(), r); code != 0 {
		t.Errorf("a lingering user with a mappable run-as validates: exit=%d\n%s", code, stderr)
	}
}

// Deploy and validate ask podman for the user's id map and refuse a numeric
// run-as outside it, with the usermod fix naming both ranges computed from the
// subid files, then this tool's own alternative. Id 0 and run-as image need no
// range, and unreadable files say the range was not computed.
func TestPodmanRunAsAgainstTheSubIDRanges(t *testing.T) {
	files := map[string]string{"/etc/subuid": "alice:100000:65536\n", "/etc/subgid": "other:165536:65536\n"}
	const fix = "sudo usermod --add-subuids 231072-301072 --add-subgids 231072-301072 alice"
	for _, c := range []struct {
		name    string
		runAs   string
		r       *probeRunner
		files   map[string]string
		wantErr []string
		probes  int
	}{
		{"inside the range", "", &probeRunner{}, files, nil, 1},
		{"uid outside the range", "70000:0", &probeRunner{}, files,
			[]string{"podman.run-as subuid 70000 is outside this user's subuid ranges (0:1 1:65536).", fix,
				"podman system migrate", "or to image"}, 1},
		{"gid outside the range", "1000:70000", &probeRunner{}, files,
			[]string{"podman.run-as subgid 70000 is outside this user's subgid ranges (0:1 1:65536)."}, 2},
		{"no allocation at all", "1000:0", &probeRunner{noMaps: true}, files,
			[]string{`this user has no subuid allocation, so podman.run-as "1000:0" cannot work`}, 1},
		{"id 0 needs no range", "0:0", &probeRunner{noMaps: true}, files, nil, 0},
		{"image needs no range", "image", &probeRunner{noMaps: true}, files, nil, 0},
		{"unreadable files: the range is not computed", "70000:0", &probeRunner{}, nil,
			[]string{"--add-subuids 100000-170000 --add-subgids 100000-170000 alice   # /etc/subuid unreadable: check this does not overlap"}, 1},
		{"probe fails", "1000:0", &probeRunner{mapErr: errors.New("exit status 125")}, files,
			[]string{`cannot check podman.run-as 1000:0 against this user's subuid ranges`, "exit status 125"}, 1},
	} {
		section := podmanEnv()
		if c.runAs != "" {
			section += "  run-as: \"" + c.runAs + "\"\n"
		}
		// Deploy exits 1 before it stores or writes anything.
		t.Run("deploy "+c.name, func(t *testing.T) {
			home := t.TempDir()
			useHost(t, linuxHost(home, c.files))
			code, stderr := runPodmanVerb(t, "deploy", section, c.r)
			if (len(c.wantErr) > 0) != (code == 1) || (len(c.wantErr) == 0 && code != 0) {
				t.Fatalf("exit=%d, want errors %q\n%s", code, c.wantErr, stderr)
			}
			for _, w := range c.wantErr {
				if !strings.Contains(stderr, w) {
					t.Errorf("stderr should carry %q, got:\n%s", w, stderr)
				}
			}
			if got := c.r.idProbes(); got != c.probes {
				t.Errorf("id map probes = %d, want %d", got, c.probes)
			}
			if wrote := unitWritten(filepath.Join(home, ".config", "containers", "systemd")); wrote != (len(c.wantErr) == 0) {
				t.Errorf("unit written = %v with errors %q", wrote, c.wantErr)
			}
			if len(c.wantErr) > 0 && c.r.ran("podman", "secret") {
				t.Errorf("nothing may be stored after a refusal, got %+v", c.r.calls)
			}
		})
		// validate reports it as a host error, exit 1.
		t.Run("validate "+c.name, func(t *testing.T) {
			useHost(t, linuxHost(t.TempDir(), c.files))
			code, stderr := runPodmanVerb(t, "validate", section, c.r)
			if (len(c.wantErr) > 0) != (code == 1) || (len(c.wantErr) == 0 && code != 0) {
				t.Fatalf("exit=%d, want errors %q\n%s", code, c.wantErr, stderr)
			}
			for _, w := range c.wantErr {
				if !strings.Contains(stderr, w) {
					t.Errorf("stderr should carry %q, got:\n%s", w, stderr)
				}
			}
		})
	}
}

// Rootless is the default. A non-root user who asks for rootful is refused, and
// so is root who does not, with exit 2 and no command run -- deploy, remove,
// start and validate alike. Root asking for rootful gets the system manager.
func TestPodmanRootfulRule(t *testing.T) {
	const rootful = "  rootful: true\n"
	asUser := func(h hostFacts) hostFacts { return h }
	asRoot := func(h hostFacts) hostFacts { h.scope.Euid = 0; return h }
	for _, c := range []struct {
		name    string
		host    func(hostFacts) hostFacts
		section string
		want    string
	}{
		{"non-root asking for rootful", asUser, podmanEnv() + rootful, "podman.rootful=true requires root, but running as euid 1000"},
		{"root without rootful", asRoot, podmanEnv(), "set podman.rootful=true to deploy the rootful engine"},
	} {
		for _, verb := range []string{"deploy", "remove", "start", "validate"} {
			t.Run(c.name+" "+verb, func(t *testing.T) {
				useHost(t, c.host(testHost(t.TempDir())))
				f := useFakeRunner(t)
				code, stderr := runPodmanVerb(t, verb, c.section, f)
				if code != 2 || !strings.Contains(stderr, c.want) {
					t.Errorf("exit=%d, want 2 naming %q:\n%s", code, c.want, stderr)
				}
				if len(f.calls) != 0 {
					t.Errorf("nothing may run, got %+v", f.calls)
				}
			})
		}
	}
	t.Run("root asking for rootful drives the system manager", func(t *testing.T) {
		useHost(t, asRoot(linuxHost(t.TempDir(), nil)))
		r := &probeRunner{}
		code, stderr := runPodmanVerb(t, "start", podmanEnv()+rootful, r)
		if code != 0 {
			t.Fatalf("exit=%d\n%s", code, stderr)
		}
		if want := []string{"systemctl", "start", "solmq-conn.service"}; !reflect.DeepEqual(r.calls[len(r.calls)-1].argv, want) {
			t.Errorf("calls = %+v, want a system-manager start", r.calls)
		}
		if r.calls[len(r.calls)-1].env != nil {
			t.Errorf("a rootful scope carries no session, got %v", r.calls[len(r.calls)-1].env)
		}
	})
}

// A relative XDG_CONFIG_HOME is refused like any usage mistake (exit 2, nothing
// run), and a session that cannot be repaired is a host failure (exit 1) before
// the preflight; both at deploy and, for the first, at validate.
func TestPodmanScopeRefusals(t *testing.T) {
	t.Run("relative XDG_CONFIG_HOME", func(t *testing.T) {
		for _, verb := range []string{"deploy", "validate"} {
			h := testHost(t.TempDir())
			h.scope.Getenv = func(k string) string {
				if k == "XDG_CONFIG_HOME" {
					return "rel/config"
				}
				return ""
			}
			useHost(t, h)
			f := useFakeRunner(t)
			code, stderr := runPodmanVerb(t, verb, podmanEnv(), f)
			if code != 2 || !strings.Contains(stderr, `XDG_CONFIG_HOME is "rel/config"`) || len(f.calls) != 0 {
				t.Errorf("%s: exit=%d, calls=%+v, want 2 naming it with nothing run:\n%s", verb, code, f.calls, stderr)
			}
		}
	})
	t.Run("no user runtime directory", func(t *testing.T) {
		h := linuxHost(t.TempDir(), nil)
		h.scope.IsDir = func(string) (bool, error) { return false, nil }
		useHost(t, h)
		for _, verb := range []string{"deploy", "remove", "start"} {
			f := useFakeRunner(t)
			code, stderr := runPodmanVerb(t, verb, podmanEnv(), f)
			if code != 1 || !strings.Contains(stderr, "/run/user/1000 does not exist") || len(f.calls) != 0 {
				t.Errorf("%s: exit=%d, calls=%+v, want 1 naming the directory with nothing run:\n%s", verb, code, f.calls, stderr)
			}
		}
		// validate reports it as a host error.
		f := useFakeRunner(t)
		if code, stderr := runPodmanVerb(t, "validate", podmanEnv(), f); code != 1 || !strings.Contains(stderr, "/run/user/1000 does not exist") {
			t.Errorf("validate: exit=%d, want 1 naming the directory:\n%s", code, stderr)
		}
	})
}
