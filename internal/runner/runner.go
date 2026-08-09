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

// Runner runs one process. argv[0] is the program and argv[1:] its arguments;
// stdin, when non-empty, is written to the process's standard input. It returns
// the combined stdout+stderr (for error context) and the process error.
type Runner interface {
	Run(argv []string, stdin string) (string, error)
}

// OS is the production Runner: it runs argv via os/exec with no shell.
type OS struct{}

// Run starts argv[0] with argv[1:], feeds stdin when non-empty, and returns the
// combined output. It never passes the string through a shell.
func (OS) Run(argv []string, stdin string) (string, error) {
	if len(argv) == 0 {
		return "", fmt.Errorf("empty command")
	}
	cmd := exec.Command(argv[0], argv[1:]...) //nolint:gosec // argv tokens are SafeToken-validated upstream; no shell is involved
	if stdin != "" {
		cmd.Stdin = strings.NewReader(stdin)
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
	tokens := strings.Fields(s)
	if len(tokens) == 0 {
		return nil, fmt.Errorf("command is empty")
	}
	for _, t := range tokens {
		if !validate.SafeToken(t) {
			return nil, fmt.Errorf("command token %q contains an unsafe character (no spaces, quotes, backslash, control chars, or shell metacharacters)", t)
		}
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
	return r.Run(argv, manifest)
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
// `<command> compose -f <composeFile> down` (delete). The compose file and any
// credential env-file must already be written by the caller.
func Docker(r Runner, command, action, composeFile string) (string, error) {
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
	return r.Run(argv, "")
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
	o, err := r.Run(sc.systemctlArgs("daemon-reload"), "")
	out.WriteString(o)
	if err != nil {
		return out.String(), fmt.Errorf("systemctl daemon-reload: %w", err)
	}
	for _, s := range services {
		o, err = r.Run(sc.systemctlArgs("start", s), "")
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
		o, err := r.Run(sc.systemctlArgs("stop", s), "")
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
	o, err := r.Run(sc.systemctlArgs("daemon-reload"), "")
	out.WriteString(o)
	if err != nil {
		return out.String(), fmt.Errorf("systemctl daemon-reload: %w", err)
	}
	return out.String(), nil
}
