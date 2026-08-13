// Package runner executes the deploy/delete actions for each target by shelling
// out to the configured CLI (kubectl/oc, docker, podman/systemctl) through an
// argv slice -- never a shell string, so no token is ever re-interpreted by a
// shell. Command strings from env.yaml are tokenised and every token is checked
// against the safe charset (validate.SafeToken) before a process is started.
//
// The package is deliberately thin: rendering lives in gen/deploy/dockergen/
// podmangen, and the CLI owns path resolution. runner only parses commands,
// writes files (secrets 0600, never logged), and runs argv slices via an
// injected Runner so the exec boundary is unit-testable.
package runner

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/spec"
	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/validate"
)

// Deploy actions shared by every target.
const (
	ActionDeploy = "deploy"
	ActionDelete = "delete"
)

// Cmd is one process invocation. Argv[0] is the program and Argv[1:] its
// arguments; Stdin, when non-empty, is written to the process's standard input.
//
// Env carries extra KEY=VALUE entries for the child's environment only. It is
// the channel credential values travel on -- they never appear in Argv (where
// `ps` would show them) and never in a file. Nothing in this package logs Env,
// and callers must not echo it.
type Cmd struct {
	Argv  []string
	Stdin string
	Env   []string
}

// Runner runs one process, returning the combined stdout+stderr (for error
// context) and the process error.
type Runner interface {
	Run(cmd Cmd) (string, error)
}

// OS is the production Runner: it runs argv via os/exec with no shell.
type OS struct{}

// Run starts Argv[0] with Argv[1:], feeds Stdin when non-empty, and returns the
// combined output. It never passes the string through a shell.
func (OS) Run(c Cmd) (string, error) {
	if len(c.Argv) == 0 {
		return "", fmt.Errorf("empty command")
	}
	cmd := exec.Command(c.Argv[0], c.Argv[1:]...) //nolint:gosec // argv tokens are SafeToken-validated upstream; no shell is involved
	if c.Stdin != "" {
		cmd.Stdin = strings.NewReader(c.Stdin)
	}
	if len(c.Env) > 0 {
		// Appended after the inherited environment so a supplied value wins over
		// an ambient one of the same name.
		cmd.Env = append(os.Environ(), c.Env...)
	}
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	err := cmd.Run()
	return buf.String(), err
}

// ParseCommand splits a command string into an argv slice on whitespace (the
// same tokenisation validate.checkCommand uses, so a command that passed
// validation parses identically here) and rejects any token that is not
// SafeToken-clean. Quoted arguments containing spaces are intentionally
// unsupported: a token with a space or quote fails the safe-charset gate.
func ParseCommand(s string) ([]string, error) {
	tokens, bad := validate.Tokenize(s)
	if len(bad) > 0 {
		return nil, fmt.Errorf("command token %q contains an unsafe character (%s)", bad[0], validate.UnsafeTokenReason)
	}
	if bad != nil {
		return nil, fmt.Errorf("command is empty")
	}
	return tokens, nil
}

// WriteFile writes content to path with mode, creating parent directories as
// needed. Secret material (credential env-files) is written with mode 0600 and
// its content is never logged; callers must not echo it.
func WriteFile(path, content string, mode os.FileMode) error {
	if dir := filepath.Dir(path); dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("creating %s: %w", dir, err)
		}
	}
	if err := os.WriteFile(path, []byte(content), mode); err != nil {
		return fmt.Errorf("writing %s: %w", path, err)
	}
	return nil
}

// Kubernetes applies (deploy) or deletes (delete) the manifest by running
// `<command> apply -f -` / `<command> delete -f -` with the manifest on stdin,
// so the manifest bytes never appear on a command line.
func Kubernetes(r Runner, command, action, manifest string) (string, error) {
	argv, err := ParseCommand(command)
	if err != nil {
		return "", err
	}
	verb, err := kubeVerb(action)
	if err != nil {
		return "", err
	}
	argv = append(argv, verb, "-f", "-")
	return r.Run(Cmd{Argv: argv, Stdin: manifest})
}

func kubeVerb(action string) (string, error) {
	switch action {
	case ActionDeploy:
		return "apply", nil
	case ActionDelete:
		return "delete", nil
	default:
		return "", fmt.Errorf("unknown action %q (want %q or %q)", action, ActionDeploy, ActionDelete)
	}
}

// Docker runs `<command> compose -f <composeFile> up -d` (deploy) or
// `<command> compose -f <composeFile> down` (delete). The compose file must
// already be written by the caller.
//
// env carries the credential values the compose file's environment-provider
// secrets read, in this child process only -- which is why no secret material is
// ever written to disk for docker. It is passed on both actions: `down` resolves
// the same secret declarations `up` did.
func Docker(r Runner, command, action, composeFile string, env []string) (string, error) {
	argv, err := ParseCommand(command)
	if err != nil {
		return "", err
	}
	argv = append(argv, "compose", "-f", composeFile)
	switch action {
	case ActionDeploy:
		argv = append(argv, "up", "-d")
	case ActionDelete:
		argv = append(argv, "down")
	default:
		return "", fmt.Errorf("unknown action %q (want %q or %q)", action, ActionDeploy, ActionDelete)
	}
	return r.Run(Cmd{Argv: argv, Env: env})
}

// PodmanSecretCreate stores one credential in podman's secret store, replacing
// any previous entry of that name.
//
// The value crosses on stdin, never in argv. It is a remove-then-create rather
// than `create --replace` because --replace needs podman 4.7 and the rest of the
// quadlet secret wiring only needs 4.5.
func PodmanSecretCreate(r Runner, command, name, value string) (string, error) {
	argv, err := ParseCommand(command)
	if err != nil {
		return "", err
	}
	out, err := r.Run(Cmd{Argv: append(append([]string(nil), argv...), "secret", "rm", "--ignore", name)})
	if err != nil {
		return out, fmt.Errorf("podman secret rm %s: %w", name, err)
	}
	o, err := r.Run(Cmd{Argv: append(append([]string(nil), argv...), "secret", "create", name, "-"), Stdin: value})
	out += o
	if err != nil {
		return out, fmt.Errorf("podman secret create %s: %w", name, err)
	}
	return out, nil
}

// PodmanSecretRemove deletes the named secrets, ignoring any that are already
// gone so a repeated delete stays quiet.
func PodmanSecretRemove(r Runner, command string, names []string) (string, error) {
	if len(names) == 0 {
		return "", nil
	}
	argv, err := ParseCommand(command)
	if err != nil {
		return "", err
	}
	argv = append(argv, "secret", "rm", "--ignore")
	out, err := r.Run(Cmd{Argv: append(argv, names...)})
	if err != nil {
		return out, fmt.Errorf("podman secret rm: %w", err)
	}
	return out, nil
}

// ---- podman quadlet ----------------------------------------------------------

const (
	systemctl      = "systemctl"
	flagUser       = "--user"
	quadletSystem  = "/etc/containers/systemd"
	quadletUserSub = ".config/containers/systemd"
)

// QuadletScope is the resolved placement for podman quadlet units: the directory
// the .container files live in and whether systemctl runs in --user mode.
type QuadletScope struct {
	Dir      string
	UserMode bool
}

// ResolveQuadletScope maps the configured scope (auto|user|system) and optional
// dir override to a concrete directory and systemctl mode. auto resolves to
// system for the root user (euid 0) and user otherwise. A dir override replaces
// the default directory for the resolved scope but does not change the mode.
func ResolveQuadletScope(scope, dirOverride string) (QuadletScope, error) {
	var sc QuadletScope
	switch scope {
	case "", spec.QuadletScopeAuto:
		sc.UserMode = os.Geteuid() != 0
	case spec.QuadletScopeSystem:
		sc.UserMode = false
	case spec.QuadletScopeUser:
		sc.UserMode = true
	default:
		return QuadletScope{}, fmt.Errorf("unknown quadlet scope %q (want auto, user, or system)", scope)
	}
	if sc.UserMode {
		home, err := os.UserHomeDir()
		if err != nil {
			return QuadletScope{}, fmt.Errorf("resolving user quadlet dir: %w", err)
		}
		sc.Dir = filepath.Join(home, quadletUserSub)
	} else {
		sc.Dir = quadletSystem
	}
	if dirOverride != "" {
		sc.Dir = dirOverride
	}
	return sc, nil
}

// systemctlArgs prefixes --user when the scope runs in user mode.
func (sc QuadletScope) systemctlArgs(args ...string) []string {
	argv := []string{systemctl}
	if sc.UserMode {
		argv = append(argv, flagUser)
	}
	return append(argv, args...)
}

// PodmanDeploy reloads the systemd generator and starts one service per unit.
// Unit files must already be written into sc.Dir by the caller (their bind-mount
// paths are baked into the rendered Volume= lines). services are the systemd
// unit names to start, e.g. "solmq-connector.service".
func PodmanDeploy(r Runner, sc QuadletScope, services []string) (string, error) {
	var out strings.Builder
	o, err := r.Run(Cmd{Argv: sc.systemctlArgs("daemon-reload")})
	out.WriteString(o)
	if err != nil {
		return out.String(), fmt.Errorf("systemctl daemon-reload: %w", err)
	}
	for _, s := range services {
		o, err = r.Run(Cmd{Argv: sc.systemctlArgs("start", s)})
		out.WriteString(o)
		if err != nil {
			return out.String(), fmt.Errorf("systemctl start %s: %w", s, err)
		}
	}
	return out.String(), nil
}

// PodmanDelete stops each service, removes its unit file from sc.Dir, and
// reloads the generator. units are the .container filenames to remove; services
// are the matching systemd unit names to stop.
//
// units are generator-built, pre-validated basenames (a DNS-1123 label plus
// ".container"), so no path-traversal containment guard is applied here -- see
// TestCheckContainerNameRejected in internal/validate/validate_extra_test.go.
func PodmanDelete(r Runner, sc QuadletScope, services, units []string) (string, error) {
	var out strings.Builder
	for _, s := range services {
		o, err := r.Run(Cmd{Argv: sc.systemctlArgs("stop", s)})
		out.WriteString(o)
		if err != nil {
			return out.String(), fmt.Errorf("systemctl stop %s: %w", s, err)
		}
	}
	for _, u := range units {
		if err := os.Remove(filepath.Join(sc.Dir, u)); err != nil && !os.IsNotExist(err) {
			return out.String(), fmt.Errorf("removing unit %s: %w", u, err)
		}
	}
	o, err := r.Run(Cmd{Argv: sc.systemctlArgs("daemon-reload")})
	out.WriteString(o)
	if err != nil {
		return out.String(), fmt.Errorf("systemctl daemon-reload: %w", err)
	}
	return out.String(), nil
}
