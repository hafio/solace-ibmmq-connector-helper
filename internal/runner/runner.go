// Package runner executes the deploy/remove actions for each target by shelling
// out to the configured CLI (kubectl/oc, docker, podman/systemctl) through an
// argv slice -- never a shell string, so no token is ever re-interpreted by a
// shell. Command strings from env.yaml are tokenised and every token is checked
// against the safe charset (validate.SafeToken) before a process is started.
// Beyond the charset, ParseCommand enforces validate.CheckDeployCommand's
// per-platform binary allowlist and flag-shape rules, Preflight runs a
// read-only login/daemon probe before any mutating call, and the real Runner
// resolves argv[0] via exec.LookPath, refusing a current-directory resolution.
//
// The package is deliberately thin: rendering lives in gen/deploy/dockergen/
// podmangen, and the CLI owns path resolution. runner only parses commands,
// writes files (secrets 0600, never logged), and runs argv slices via an
// injected Runner so the exec boundary is unit-testable.
package runner

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/spec"
	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/validate"
)

// preflightHints gives an actionable next step per platform when Preflight's
// read-only probe fails, keyed by the same platform constants CheckDeployCommand
// uses.
var preflightHints = map[string]string{
	validate.PlatformKubernetes: "log in or select a context first (e.g. 'oc login', or check 'kubectl config current-context'), then re-run",
	validate.PlatformDocker:     "the docker daemon is unreachable; start it or select a working context, then re-run",
	validate.PlatformPodman:     "podman is unreachable; check the podman machine or socket, then re-run",
}

// Deploy actions shared by every target. These are the CLI verbs, and they reach
// the operator verbatim (report's "ok: remove kubernetes", runAction's flag-error
// prefix), so they stay spelled as the verbs are. The platform subcommands they
// map to are separate literals -- kubeVerb returns kubectl's own "delete",
// canIVerb the can-i permission of the same name, Docker compose's "down".
const (
	ActionDeploy = "deploy"
	ActionRemove = "remove"
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

// Streamer is the optional second capability a Runner may have: run one process
// and copy its output onward as it arrives, until the process exits or ctx is
// cancelled.
//
// It is a separate interface rather than a second method on Runner because Run
// answers a different question -- it returns only once the process has exited,
// which is exactly what a followed log cannot do -- and because most Runners
// (every fake that only ever needs to record argv) have no business growing a
// streaming implementation. A caller that needs to follow asks for this seam by
// type assertion and fails loudly when the Runner it was handed does not have
// it; OS does, pinned below.
//
// stdout and stderr are separate on purpose. Run merges them because its
// callers want one blob of error context, but a followed log is a stream an
// operator redirects to a file, and the CLI's own diagnostics must not land in
// the middle of it.
type Streamer interface {
	Stream(ctx context.Context, cmd Cmd, stdout, stderr io.Writer) error
}

// Splitter is the optional capability a Runner may have of returning a
// process's stdout and stderr apart from each other, both fully buffered.
//
// It exists because Run deliberately merges them, and that merge is wrong for
// every caller that has to *parse* the output. A platform CLI writes warnings
// to stderr whenever it feels like it -- OpenShift greets `oc get` with
// "Warning: apps.openshift.io/v1 DeploymentConfig is deprecated", kubectl warns
// on deprecated APIs -- and merged into stdout that line lands in front of the
// JSON, where it fails to parse on its first character. Nothing is wrong with
// the command, the cluster, or the request; the answer was simply concatenated
// with a diagnostic about it.
//
// stderr is returned rather than discarded because it is still the best error
// context there is when the command fails, which is what Run was merging it in
// for. Callers put it in the error and parse only stdout.
//
// A Runner that does not implement this falls back to Run, which is what the
// helpers below do: that is today's behaviour, and it keeps every fake that
// only records argv working unchanged. OS implements it, pinned below, so
// production always splits.
type Splitter interface {
	RunSplit(cmd Cmd) (stdout, stderr string, err error)
}

// Attacher is the optional third capability a Runner may have: run one process
// with the operator's own terminal handed straight to it, and report the status
// it exited with.
//
// It is a separate interface rather than a third method on Runner, or a flag on
// Cmd, because it answers a question neither of the other two can. Run returns
// the child's output as a string, which is meaningless for a session whose
// output never existed as a value -- it went to the terminal as it was typed.
// Stream copies output through a pair of io.Writers, and an io.Writer is
// exactly what must not happen here: os/exec interposes an OS pipe for any
// writer that is not an *os.File, and a pipe is not a terminal, so the engine
// would refuse the pseudo-terminal it was asked for and the shell would come up
// with no prompt, no line editing and no job control. The *os.File parameters
// state that guarantee in the type rather than in a comment -- only a real
// descriptor can be inherited, and os/exec passes an *os.File through untouched.
//
// There is deliberately no context.Context. Stream needs one because a followed
// log is ended from outside, by the operator pressing Ctrl-C at a terminal the
// child does not own. An attached session is ended from inside, by the operator
// typing exit or Ctrl-D at the shell itself, and a cancellable session would
// invite a caller to pull a shell out from under a half-typed command.
//
// A caller asks for this seam by type assertion and fails loudly when the
// Runner it was handed does not have it, exactly as actLogs does for Streamer;
// OS has it, pinned below.
type Attacher interface {
	Attach(cmd Cmd, stdin, stdout, stderr *os.File) (int, error)
}

// The production Runner is also the production Streamer and Attacher. Asserting
// both here makes a follow-mode or session caller's type assertion unreachable
// in production a compile-time fact rather than a hope.
var (
	_ Streamer = OS{}
	_ Attacher = OS{}
	_ Splitter = OS{}
)

// OS is the production Runner: it runs argv via os/exec with no shell.
type OS struct{}

// Run resolves Argv[0] via exec.LookPath (refusing an exec.ErrDot resolution --
// a binary found only relative to the current directory, e.g. an unzipped
// generator-page folder placed ahead of PATH on Windows), starts it with
// Argv[1:], feeds Stdin when non-empty, and returns the combined output. It
// never passes the string through a shell.
//
// It prints nothing of its own: the process's combined output is returned to
// the caller, which is the only thing an operator needs on the happy path and
// is reported in full alongside the error on a failure.
func (OS) Run(c Cmd) (string, error) {
	resolved, err := resolveArgv0(c)
	if err != nil {
		return "", err
	}
	cmd := exec.Command(resolved, c.Argv[1:]...) //nolint:gosec // argv tokens are SafeToken-validated upstream; no shell is involved
	applyCmdInput(cmd, c)
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	runErr := cmd.Run()
	return buf.String(), runErr
}

// RunSplit runs argv the way Run does -- same PATH resolution, same refusal of
// a current-directory hit, no shell, same Stdin and Env wiring -- but keeps the
// two output streams apart instead of merging them.
func (OS) RunSplit(c Cmd) (string, string, error) {
	resolved, err := resolveArgv0(c)
	if err != nil {
		return "", "", err
	}
	cmd := exec.Command(resolved, c.Argv[1:]...) //nolint:gosec // argv tokens are SafeToken-validated upstream; no shell is involved
	applyCmdInput(cmd, c)
	var out, errOut bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errOut
	runErr := cmd.Run()
	return out.String(), errOut.String(), runErr
}

// runParsed is how every helper that parses a command's output runs it: through
// Splitter when the Runner has it, so a warning on stderr cannot corrupt the
// answer, and through Run otherwise.
//
// The returned string is what the caller parses. On failure the error carries
// whatever the command said, from whichever stream it said it on, so a failure
// message is no less useful than it was when the two were merged.
func runParsed(r Runner, c Cmd) (string, error) {
	sp, ok := r.(Splitter)
	if !ok {
		return r.Run(c)
	}
	out, errOut, err := sp.RunSplit(c)
	if err != nil {
		return strings.TrimSpace(out + "\n" + errOut), err
	}
	return out, nil
}

// streamWaitDelay bounds how long a cancelled Stream waits for the child to
// exit on its own before the process is killed outright. A followed log ends
// with the operator pressing Ctrl-C, and a CLI that appears to hang after that
// reads as a broken tool, so the child gets a moment to shut down cleanly and
// then does not get to argue.
const streamWaitDelay = 2 * time.Second

// Stream runs argv the way Run does -- same PATH resolution, same refusal of a
// current-directory hit, no shell -- but copies the process's output onward as
// it is produced instead of buffering it, and returns when the process exits or
// ctx is cancelled.
//
// A cancelled ctx is the normal end of a follow, not a failure: the returned
// error is nil in that case, so the caller does not have to unpick "the
// operator pressed Ctrl-C" from "the engine went away".
func (OS) Stream(ctx context.Context, c Cmd, stdout, stderr io.Writer) error {
	resolved, err := resolveArgv0(c)
	if err != nil {
		return err
	}
	cmd := exec.CommandContext(ctx, resolved, c.Argv[1:]...) //nolint:gosec // argv tokens are SafeToken-validated upstream; no shell is involved
	applyCmdInput(cmd, c)
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	// Interrupt rather than Kill so the child can flush and close its own
	// connection. Windows cannot deliver os.Interrupt to another process at all,
	// so a refused signal falls straight through to Kill rather than leaving the
	// child running; WaitDelay is the backstop for a child that accepts the
	// signal and then ignores it.
	cmd.Cancel = func() error {
		if err := cmd.Process.Signal(os.Interrupt); err != nil {
			return cmd.Process.Kill()
		}
		return nil
	}
	cmd.WaitDelay = streamWaitDelay
	runErr := cmd.Run()
	if ctx.Err() != nil {
		// The process was ended on purpose. Whatever exit status that produced
		// (a signal, or the WaitDelay kill) describes how it was stopped, not
		// whether it worked.
		return nil
	}
	return runErr
}

// Attach runs argv the way Run does -- same PATH resolution, same refusal of a
// current-directory hit, no shell -- but gives the child the three files it was
// handed as its own standard input, output and error, and then waits. Nothing
// is buffered, nothing is copied, and this process writes nothing of its own to
// any of them.
//
// The returned int is the status the child exited with; err is non-nil only
// when the child could not be started at all. A child that ran and exited
// non-zero is not an error at this layer: only the caller knows whether that
// status means the engine could not attach or means the operator's last command
// failed, and unwrapping the ExitError here is what keeps os/exec out of the
// caller. A child killed by a signal reports no status of its own, so it is
// reported as 1 rather than as ExitCode's -1, which is not an exit status any
// process could have produced.
//
// A Cmd carrying Stdin text is refused: "write this string to the child" and
// "give the child the operator's keyboard" are contradictory requests, and
// silently honouring one of them is the class of bug LogsArgv's --previous
// guard exists to prevent. A nil file is refused too, because os/exec treats a
// typed-nil *os.File in its io.Writer field as a live writer and panics at the
// first write, which is a worse place to find out than here.
func (OS) Attach(c Cmd, stdin, stdout, stderr *os.File) (int, error) {
	if c.Stdin != "" {
		return 0, fmt.Errorf("an attached session takes the terminal's own stdin, so Cmd.Stdin must be empty")
	}
	if stdin == nil || stdout == nil || stderr == nil {
		return 0, fmt.Errorf("an attached session needs all three of stdin, stdout and stderr")
	}
	resolved, err := resolveArgv0(c)
	if err != nil {
		return 0, err
	}
	cmd := exec.Command(resolved, c.Argv[1:]...) //nolint:gosec // argv tokens are SafeToken-validated upstream; no shell is involved
	applyCmdEnv(cmd, c)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = stdin, stdout, stderr
	// No Cancel and no WaitDelay, unlike Stream: there is no context to cancel,
	// and a session the operator is still typing into must not be killed on a
	// timer.
	runErr := cmd.Run()
	var ee *exec.ExitError
	if errors.As(runErr, &ee) {
		if code := ee.ExitCode(); code >= 0 {
			return code, nil
		}
		return 1, nil
	}
	if runErr != nil {
		return 0, runErr
	}
	return 0, nil
}

// resolveArgv0 resolves Argv[0] via exec.LookPath for Run, Stream and Attach,
// so the refusal rules cannot drift apart between them.
func resolveArgv0(c Cmd) (string, error) {
	if len(c.Argv) == 0 {
		return "", fmt.Errorf("empty command")
	}
	resolved, err := exec.LookPath(c.Argv[0])
	if err != nil {
		// exec.LookPath returns a non-nil error both when the binary is not found
		// on PATH and when it resolves only via the current directory (wrapping
		// exec.ErrDot); both cases are rejected here rather than silently exec'd.
		return "", fmt.Errorf("resolving %q: %w", c.Argv[0], err)
	}
	return resolved, nil
}

// applyCmdEnv wires Env onto an exec.Cmd, shared by Run, Stream and Attach so
// the credential channel behaves identically in all three.
//
// It is split out of applyCmdInput below because Attach needs this half alone:
// it hands the child the operator's real terminal as stdin, so the string-stdin
// wiring below is not merely unnecessary there but would clobber it.
func applyCmdEnv(cmd *exec.Cmd, c Cmd) {
	if len(c.Env) > 0 {
		// Appended after the inherited environment so a supplied value wins over
		// an ambient one of the same name.
		cmd.Env = append(os.Environ(), c.Env...)
	}
}

// applyCmdInput wires Stdin and Env onto an exec.Cmd, shared by Run and Stream
// so the credential channel behaves identically in both.
func applyCmdInput(cmd *exec.Cmd, c Cmd) {
	if c.Stdin != "" {
		cmd.Stdin = strings.NewReader(c.Stdin)
	}
	applyCmdEnv(cmd, c)
}

// ParseCommand splits a command string into an argv slice on whitespace (the
// same tokenisation validate.checkCommand uses, so a command that passed
// validation parses identically here), rejects any token that is not
// SafeToken-clean, and then runs the full deploy-command guard
// (validate.CheckDeployCommand): argv[0] must be a bare name on the platform's
// allowlist or extraAllowed, and every later token must be flag-shaped. Quoted
// arguments containing spaces are intentionally unsupported: a token with a
// space or quote fails the safe-charset gate before the guard ever runs.
func ParseCommand(platform, cmd string, extraAllowed []string) ([]string, error) {
	tokens, bad := validate.Tokenize(cmd)
	if len(bad) > 0 {
		return nil, fmt.Errorf("command token %q contains an unsafe character (%s)", bad[0], validate.UnsafeTokenReason)
	}
	if bad != nil {
		return nil, fmt.Errorf("command is empty")
	}
	if err := validate.CheckDeployCommand(platform, cmd, extraAllowed); err != nil {
		return nil, err
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

// Kubernetes applies (deploy) or deletes (remove) the manifest by running
// `<command> apply -f -` / `<command> delete -f -` with the manifest on stdin,
// so the manifest bytes never appear on a command line.
func Kubernetes(r Runner, command, action, manifest string, extraAllowed []string) (string, error) {
	argv, err := ParseCommand(validate.PlatformKubernetes, command, extraAllowed)
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
	case ActionRemove:
		return "delete", nil
	default:
		return "", fmt.Errorf("unknown action %q (want %q or %q)", action, ActionDeploy, ActionRemove)
	}
}

// Docker runs `<command> compose -f <composeFile> up -d` (deploy) or
// `<command> compose -f <composeFile> down` (remove). The compose file must
// already be written by the caller.
//
// env carries the credential values the compose file's environment-provider
// secrets read, in this child process only -- which is why no secret material is
// ever written to disk for docker. It is passed on both actions: `down` resolves
// the same secret declarations `up` did.
func Docker(r Runner, command, action, composeFile string, env []string, extraAllowed []string) (string, error) {
	argv, err := ParseCommand(validate.PlatformDocker, command, extraAllowed)
	if err != nil {
		return "", err
	}
	argv = append(argv, "compose", "-f", composeFile)
	switch action {
	case ActionDeploy:
		argv = append(argv, "up", "-d")
	case ActionRemove:
		argv = append(argv, "down")
	default:
		return "", fmt.Errorf("unknown action %q (want %q or %q)", action, ActionDeploy, ActionRemove)
	}
	return r.Run(Cmd{Argv: argv, Env: env})
}

// Preflight runs a cheap, read-only probe before any generated file is written
// or any mutating command runs, so a deploy/remove that can never succeed
// (not logged in, daemon down) stops before touching anything:
//
//   - kubernetes: `<argv> auth can-i <create|delete> deployment [--namespace <ns>]`
//     -- create for deploy, delete for remove (can-i's own verbs, not apply/delete
//     used by kubeVerb for the eventual manifest command).
//   - docker/podman: `<argv> info`.
//
// A non-nil error wraps the probe's own output with a platform-specific hint
// (log in, start the daemon, ...) so the operator gets a next step instead of a
// bare CLI error.
func Preflight(r Runner, platform, command, action, namespace string, extraAllowed []string) (string, error) {
	argv, err := ParseCommand(platform, command, extraAllowed)
	if err != nil {
		return "", err
	}
	switch platform {
	case validate.PlatformKubernetes:
		verb, verr := canIVerb(action)
		if verr != nil {
			return "", verr
		}
		argv = append(argv, "auth", "can-i", verb, "deployment")
		if namespace != "" {
			argv = append(argv, "--namespace", namespace)
		}
	case validate.PlatformDocker, validate.PlatformPodman:
		argv = append(argv, "info")
	default:
		return "", fmt.Errorf("unknown platform %q (want %q, %q, or %q)", platform, validate.PlatformKubernetes, validate.PlatformDocker, validate.PlatformPodman)
	}
	out, runErr := r.Run(Cmd{Argv: argv})
	if runErr != nil {
		hint, ok := preflightHints[platform]
		if !ok {
			return out, fmt.Errorf("preflight failed for %s: %w", platform, runErr)
		}
		return out, fmt.Errorf("preflight failed for %s: %w\n%s", platform, runErr, hint)
	}
	return out, nil
}

// canIVerb maps a deploy/remove action to the verb `kubectl auth can-i` expects,
// which is create/delete -- not apply/delete, the actual manifest verbs kubeVerb
// returns -- because can-i asks about a permission, not a subcommand.
func canIVerb(action string) (string, error) {
	switch action {
	case ActionDeploy:
		return "create", nil
	case ActionRemove:
		return "delete", nil
	default:
		return "", fmt.Errorf("unknown action %q (want %q or %q)", action, ActionDeploy, ActionRemove)
	}
}

// PodmanSecretCreate stores one credential in podman's secret store, replacing
// any previous entry of that name.
//
// The value crosses on stdin, never in argv. It is a remove-then-create rather
// than `create --replace` because --replace needs podman 4.7 and the rest of the
// quadlet secret wiring only needs 4.5.
func PodmanSecretCreate(r Runner, command, name, value string, extraAllowed []string) (string, error) {
	argv, err := ParseCommand(validate.PlatformPodman, command, extraAllowed)
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
func PodmanSecretRemove(r Runner, command string, names []string, extraAllowed []string) (string, error) {
	if len(names) == 0 {
		return "", nil
	}
	argv, err := ParseCommand(validate.PlatformPodman, command, extraAllowed)
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

// PodmanRemove stops each service, removes its unit file from sc.Dir, and
// reloads the generator. units are the .container filenames to remove; services
// are the matching systemd unit names to stop.
//
// units are generator-built, pre-validated basenames (a DNS-1123 label plus
// ".container"), so no path-traversal containment guard is applied here -- see
// TestCheckContainerNameRejected in internal/validate/validate_extra_test.go.
func PodmanRemove(r Runner, sc QuadletScope, services, units []string) (string, error) {
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

// ---- status verb: discovery and in-container exec ---------------------------

// The status verb's read-only queries. Every one of them goes out as an argv
// slice through the same Runner seam the deploy path uses -- no shell, ever --
// and every operator-supplied token in them (namespace, selector, target name)
// must already be validated by the caller. Tokens this package adds itself
// (subcommands, flags, output formats, Go templates) are tool-authored
// constants: a brace or a quote in one of them is data to the process, never to
// a shell, because no shell is involved.
//
// They are deliberately one call for many targets wherever the CLI allows it
// (`get pods` for a whole selector, `inspect a b c`), so a status run costs a
// roughly fixed number of calls rather than one per instance.

// KubernetesPodsJSON lists pods as a JSON document: `<cmd> get pods -o json`
// scoped by exactly one of selector, explicit names, or every namespace.
//
// It is read-only and never creates, deletes, or execs into anything. An empty
// match is not an error -- kubectl answers with an empty List -- so the caller
// can produce an actionable "no pods found for selector ..." message naming the
// namespace and selector it used, rather than a generic wrapped error.
//
// -o json rather than -o name because the same call then answers the whole
// container view (state, restart counts, images, limits) for no extra request:
// discovery and the report read one document.
func KubernetesPodsJSON(r Runner, cmd []string, namespace, selector string, names []string, allNamespaces bool) (string, error) {
	argv := append(append([]string(nil), cmd...), "get", "pods")
	argv = append(argv, names...)
	switch {
	case allNamespaces:
		// --all-namespaces outranks a namespace: --all is explicitly a
		// cluster-wide search, so a namespace resolved from env.yaml must not
		// narrow it back down.
		argv = append(argv, "--all-namespaces")
	case namespace != "":
		argv = append(argv, "-n", namespace)
	}
	if selector != "" {
		argv = append(argv, "-l", selector)
	}
	argv = append(argv, "-o", "json")
	out, err := runParsed(r, Cmd{Argv: argv})
	if err != nil {
		return out, fmt.Errorf("listing pods: %w\n%s", err, out)
	}
	return out, nil
}

// KubernetesGetJSON reads one object as JSON: `<cmd> get <kind> <name> -o json`.
// Used for the workload summary (deployment, service) and for the presence of
// each object a pod references (secret, configmap, persistentvolumeclaim).
//
// kind is a tool-authored constant; name and namespace are operator-supplied and
// must already be validated by the caller. A missing object exits non-zero here,
// which the caller reports as "missing" rather than as a failure -- that is the
// answer it asked for.
func KubernetesGetJSON(r Runner, cmd []string, namespace, kind, name string) (string, error) {
	argv := append(append([]string(nil), cmd...), "get", kind, name)
	if namespace != "" {
		argv = append(argv, "-n", namespace)
	}
	argv = append(argv, "-o", "json")
	out, err := runParsed(r, Cmd{Argv: argv})
	if err != nil {
		return out, fmt.Errorf("reading %s/%s: %w\n%s", kind, name, err, out)
	}
	return out, nil
}

// KubernetesListJSON lists every object of the given types in a namespace:
// `<cmd> get <types> -n <ns> -o json`.
//
// types is a tool-authored constant (a comma-separated kubectl resource list),
// never operator input. It is separate from KubernetesGetJSON because that one
// appends a name unconditionally, and an empty name there would put an empty
// element in the argv rather than listing everything.
func KubernetesListJSON(r Runner, cmd []string, namespace, types string) (string, error) {
	argv := append(append([]string(nil), cmd...), "get", types)
	if namespace != "" {
		argv = append(argv, "-n", namespace)
	}
	argv = append(argv, "-o", "json")
	out, err := runParsed(r, Cmd{Argv: argv})
	if err != nil {
		return out, fmt.Errorf("listing %s in namespace %q: %w\n%s", types, namespace, err, out)
	}
	return out, nil
}

// KubernetesTop samples pod resource usage:
// `<cmd> top pod [names|-l selector] --containers --no-headers`.
//
// --containers attributes the sample to the connector container rather than
// summing every container in the pod, so the number is comparable with that
// container's own limits. It needs a metrics API in the cluster
// (metrics-server); when that is absent kubectl fails here and the caller
// degrades to a note rather than dropping the report.
func KubernetesTop(r Runner, cmd []string, namespace, selector string, names []string, allNamespaces bool) (string, error) {
	argv := append(append([]string(nil), cmd...), "top", "pod")
	argv = append(argv, names...)
	switch {
	case allNamespaces:
		argv = append(argv, "--all-namespaces")
	case namespace != "":
		argv = append(argv, "-n", namespace)
	}
	if selector != "" {
		argv = append(argv, "-l", selector)
	}
	argv = append(argv, "--containers", "--no-headers")
	out, err := runParsed(r, Cmd{Argv: argv})
	if err != nil {
		return out, fmt.Errorf("sampling pod resource usage: %w\n%s", err, out)
	}
	return out, nil
}

// EngineInspectJSON inspects one or more containers: `<cmd> inspect <names...>`
// on docker or podman, which answers a JSON array in the order the names were
// given. One call covers every target, and its output carries the compose
// project label too, so nothing else has to be asked for it.
//
// names are operator-supplied (or read back from the engine under --all) and
// must already be validated by the caller.
func EngineInspectJSON(r Runner, cmd []string, names []string) (string, error) {
	if len(names) == 0 {
		return "", fmt.Errorf("no container named to inspect")
	}
	argv := append(append([]string(nil), cmd...), "inspect")
	out, err := runParsed(r, Cmd{Argv: append(argv, names...)})
	if err != nil {
		return out, fmt.Errorf("inspecting %s: %w\n%s", strings.Join(names, ", "), err, out)
	}
	return out, nil
}

// EngineImageInspectJSON inspects an image: `<cmd> image inspect <ref>`, whose
// RepoDigests carry the registry digest the running image was pulled by --
// which the container's own inspect output does not report.
//
// ref comes back from the engine's container inspect rather than from the
// operator, so the caller re-validates it before it reaches this argv.
func EngineImageInspectJSON(r Runner, cmd []string, ref string) (string, error) {
	argv := append(append([]string(nil), cmd...), "image", "inspect", ref)
	out, err := runParsed(r, Cmd{Argv: argv})
	if err != nil {
		return out, fmt.Errorf("inspecting image %s: %w\n%s", ref, err, out)
	}
	return out, nil
}

// engineStatsFormat is the Go template `stats` renders each row with. A template
// rather than --format json deliberately: docker and podman disagree about the
// field names and even the shape (an array versus one object per line) of their
// JSON stats, while these four template fields mean the same thing on both.
// Tabs separate the fields, which no value here contains.
const engineStatsFormat = "{{.Name}}\t{{.CPUPerc}}\t{{.MemUsage}}\t{{.MemPerc}}"

// EngineStats samples container resource usage:
// `<cmd> stats --no-stream --format <template> <names...>`.
//
// --no-stream takes one sample and exits instead of streaming, but this is
// still a measurement rather than a metadata read: the engine has to watch the
// cgroup for an interval to compute a CPU percentage, which is why it is the
// one query the details level pays for rather than the default.
func EngineStats(r Runner, cmd []string, names []string) (string, error) {
	argv := append(append([]string(nil), cmd...), "stats", "--no-stream", "--format", engineStatsFormat)
	out, err := runParsed(r, Cmd{Argv: append(argv, names...)})
	if err != nil {
		return out, fmt.Errorf("sampling container resource usage: %w\n%s", err, out)
	}
	return out, nil
}

// engineListFormat is the Go template `ps` renders each row with: the name a
// container is addressed by, and the image reference --all matches against.
const engineListFormat = "{{.Names}}\t{{.Image}}"

// EngineList lists every container the engine knows about, running or not:
// `<cmd> ps --all --no-trunc --format <template>`. --all discovery filters the
// result by image reference, since without env.yaml or an explicit name the
// image is the only thing that identifies a connector instance.
//
// Stopped containers are included on purpose: an instance that died is exactly
// what a search like this is for.
func EngineList(r Runner, cmd []string) (string, error) {
	argv := append(append([]string(nil), cmd...), "ps", "--all", "--no-trunc", "--format", engineListFormat)
	out, err := runParsed(r, Cmd{Argv: argv})
	if err != nil {
		return out, fmt.Errorf("listing containers: %w\n%s", err, out)
	}
	return out, nil
}

// SystemctlNRestarts asks systemd how many times a quadlet-managed unit has
// been restarted: `systemctl [--user] show <unit> -p NRestarts --value`.
//
// This is the only truthful restart count under podman quadlet: systemd
// recreates the container rather than restarting it, so the container's own
// RestartCount reads 0 however many times the instance has died, and the unit is
// the only thing that remembers.
//
// unit is derived from the configured instance name (gen.PodmanServiceName) and
// must already be validated by the caller. An empty or unparseable answer is an
// error, so the caller can fall back to the container's own counter.
func SystemctlNRestarts(r Runner, sc QuadletScope, unit string) (int, error) {
	out, err := runParsed(r, Cmd{Argv: sc.systemctlArgs("show", unit, "-p", "NRestarts", "--value")})
	if err != nil {
		return 0, fmt.Errorf("reading %s restart count from systemd: %w\n%s", unit, err, out)
	}
	n, cerr := strconv.Atoi(strings.TrimSpace(out))
	if cerr != nil {
		return 0, fmt.Errorf("reading %s restart count from systemd: unexpected answer %q", unit, strings.TrimSpace(out))
	}
	return n, nil
}

// ContainerShell is the shell every in-container command is run through. The
// connector image is Alpine, so its userland is busybox: sh exists, bash does
// not -- the same assumption statusscript's package comment records for the
// script it renders. Spelled as a constant because it is the one token on these
// argvs that is a property of the image rather than of the engine.
const ContainerShell = "sh"

// ExecOpts is how one exec is shaped: whether the child reads standard input,
// and whether the engine is asked for a terminal.
//
// Stdin adds -i, needed whenever anything is written to the child's standard
// input -- a piped payload (InstallScript) or the operator's own keyboard. TTY
// adds -t, which asks the engine for a pseudo-terminal so a remote shell gets a
// prompt, line editing and job control.
//
// A struct rather than two positional bools, for the reason LogsOpts is one:
// execArgv(cmd, p, t, ns, true, false) says nothing at a call site about which
// flag is which, and the two are one transposition away from asking for a
// terminal that nothing can type into.
type ExecOpts struct {
	Stdin bool
	TTY   bool
}

// ExecArgv builds the `<cmd> exec` argv prefix shared by ScriptInstalled,
// InstallScript, RunStatusScript and the cli verb's attached session, keeping
// the per-platform shape in one place instead of repeating it in each caller.
//
// The shapes differ in more than spelling, and the difference is a parser
// difference rather than a style one. kubectl leaves its flag parser
// interspersed, so the namespace and container flags are accepted after the pod
// positional, and a "--" is required to mark where the in-container command
// begins. docker and podman deliberately stop interspersing on exec, so that a
// flag typed after the container name belongs to the in-container program --
// which means their own flags must come first, and a "--" would be handed
// through to the container as an argument rather than swallowed.
//
// The container is always spec.ConnectorContainerName rather than a name
// discovered from the pod. This tool renders that name, so naming it outright
// is what makes a multi-container pod fail loudly ("container connector not
// found in pod ...") instead of quietly reaching whichever container kubectl
// would have defaulted to -- the same judgement connectorIndex records for the
// reporting side, applied one layer out. The cost is that a pod this tool did
// not deploy, whose container is called something else, is no longer reachable;
// accepted, because silently reading a sidecar is the worse of the two answers.
//
// The caller appends the in-container command itself, which is either
// tool-authored (ContainerShell, or a payload built from constants) or already
// SafeToken-validated. target and namespace are operator-supplied and must
// already be validated by the caller before reaching this function.
func ExecArgv(cmd []string, platform, target, namespace string, o ExecOpts) ([]string, error) {
	if o.TTY && !o.Stdin {
		// Unreachable via the CLI, which never asks for one without the other.
		// Guarded anyway so a future caller cannot quietly get a terminal it has
		// no way to type into.
		return nil, fmt.Errorf("a tty needs stdin: ask for both or neither")
	}
	argv := append(append([]string(nil), cmd...), "exec")
	if o.Stdin {
		argv = append(argv, "-i")
	}
	if o.TTY {
		argv = append(argv, "-t")
	}
	switch platform {
	case validate.PlatformKubernetes:
		argv = append(argv, target)
		if namespace != "" {
			argv = append(argv, "-n", namespace)
		}
		argv = append(argv, "-c", spec.ConnectorContainerName, "--")
	case validate.PlatformDocker, validate.PlatformPodman:
		argv = append(argv, target)
	default:
		return nil, fmt.Errorf("unknown platform %q (want %q, %q, or %q)", platform, validate.PlatformKubernetes, validate.PlatformDocker, validate.PlatformPodman)
	}
	return argv, nil
}

// TailAll is the Tail value meaning "the whole log", spelled as a constant
// because 0 is a legitimate request (the flags only, no history) and so cannot
// double as the unset marker.
const TailAll = -1

// LogsOpts is how much of a log to read and how to render it, kept as an
// options struct so the per-platform shape below can grow a flag without every
// caller's signature changing.
//
// Since is already a canonical time.Duration string when it reaches here: the
// CLI parses the operator's spelling and passes the parsed form on, so no raw
// operator text is in this struct.
type LogsOpts struct {
	Follow, Previous, Timestamps bool
	Tail                         int
	Since                        string
}

// LogsArgv builds the argv that reads one instance's log, keeping the
// per-platform shape in one place the way ExecArgv does for exec.
//
// The two shapes differ in more than spelling: kubectl takes the pod as a
// positional with the namespace and container as flags, while docker and podman
// take their options first and the container name last. Previous has no
// docker/podman equivalent at all -- neither engine keeps a prior run's log
// under the same name -- and is refused by the caller before it gets here
// rather than being silently dropped.
//
// The container is spec.ConnectorContainerName, named outright for the reason
// ExecArgv records: the tool renders that name, and a pod it cannot find it in
// should say so rather than return some other container's log.
//
// target and namespace are operator-supplied and must already be validated by
// the caller. Everything else this function appends is a tool-authored constant
// or a number.
func LogsArgv(cmd []string, platform, target, namespace string, o LogsOpts) ([]string, error) {
	argv := append(append([]string(nil), cmd...), "logs")
	switch platform {
	case validate.PlatformKubernetes:
		argv = append(argv, target)
		if namespace != "" {
			argv = append(argv, "-n", namespace)
		}
		argv = append(argv, "-c", spec.ConnectorContainerName)
		if o.Previous {
			argv = append(argv, "-p")
		}
		argv = append(argv, logsCommonFlags(o)...)
	case validate.PlatformDocker, validate.PlatformPodman:
		if o.Previous {
			// Unreachable via the CLI, which refuses the combination with a
			// message naming the platform. Guarded anyway so a future caller
			// cannot quietly get a log that ignores what it asked for.
			return nil, fmt.Errorf("--previous is a kubernetes concept; %s has no prior-run log to read", platform)
		}
		argv = append(argv, logsCommonFlags(o)...)
		argv = append(argv, target)
	default:
		return nil, fmt.Errorf("unknown platform %q (want %q, %q, or %q)", platform, validate.PlatformKubernetes, validate.PlatformDocker, validate.PlatformPodman)
	}
	return argv, nil
}

// logsCommonFlags are the options every platform spells identically. They are
// appended in a fixed order so the argv a test asserts is the argv every run
// produces.
func logsCommonFlags(o LogsOpts) []string {
	var argv []string
	if o.Follow {
		argv = append(argv, "-f")
	}
	if o.Timestamps {
		argv = append(argv, "--timestamps")
	}
	if o.Tail != TailAll {
		argv = append(argv, "--tail", strconv.Itoa(o.Tail))
	}
	if o.Since != "" {
		argv = append(argv, "--since", o.Since)
	}
	return argv
}

// Markers the presence probe echoes, distinctive enough that engine chatter on
// the same combined-output stream cannot be mistaken for either answer.
const (
	ScriptPresentMarker = "solmq-script-present"
	ScriptAbsentMarker  = "solmq-script-absent"
)

// ScriptInstalled probes whether path already exists inside target, returning
// (present, nil) on a clear answer and an error only when the probe itself
// could not be carried out.
//
// The answer travels as a marker on stdout instead of an exit code, because an
// exit code cannot carry it unambiguously: a remote non-zero exit is how
// `test -f` says "no such file", but it is also how kubectl reports that it
// could not reach the pod at all, and both arrive here as a failed Run with
// text on combined output. Deciding between them by whether output happens to
// be empty misreads whichever case it guesses wrong, and guessing "error" on a
// missing file would refuse to install on precisely the targets that need it.
//
// path is a tool-authored constant, never operator input, so it is folded
// into the sh -c payload as-is. target and namespace are operator-supplied and
// must already be validated by the caller.
func ScriptInstalled(r Runner, cmd []string, platform, target, namespace, path string) (bool, error) {
	argv, err := ExecArgv(cmd, platform, target, namespace, ExecOpts{})
	if err != nil {
		return false, err
	}
	// The probe reports its answer as a marker on stdout and always exits 0,
	// rather than letting "file missing" be a non-zero exit. kubectl exec
	// surfaces a remote non-zero exit as its own failure plus a "command
	// terminated with exit code 1" line, so an absent script and an unreachable
	// pod are indistinguishable by exit status alone -- and treating the wrong
	// one as an error would refuse to install on exactly the targets that need
	// it. path is a tool constant, never operator input.
	argv = append(argv, ContainerShell, "-c", "if [ -f "+path+" ]; then echo "+ScriptPresentMarker+"; else echo "+ScriptAbsentMarker+"; fi")
	out, runErr := r.Run(Cmd{Argv: argv})
	switch {
	case strings.Contains(out, ScriptPresentMarker):
		return true, nil
	case strings.Contains(out, ScriptAbsentMarker):
		return false, nil
	case runErr != nil:
		return false, fmt.Errorf("probing %s on %s: %w\n%s", path, target, runErr, out)
	}
	return false, fmt.Errorf("probing %s on %s: expected %s or %s in the output, got:\n%s", path, target, ScriptPresentMarker, ScriptAbsentMarker, out)
}

// InstallScript pipes script onto the target's stdin and writes it to path,
// creating dir first: `sh -c "mkdir -p <dir> && cat > <path>"` with
// Cmd.Stdin = script. The script crosses only on stdin -- never cp, never a
// download, never the script text in argv.
//
// dir, path, and script are tool-authored constants, never operator input, so
// the sh -c payload is built from them as-is. target and namespace are
// operator-supplied and must already be validated by the caller.
func InstallScript(r Runner, cmd []string, platform, target, namespace, dir, path, script string) (string, error) {
	argv, err := ExecArgv(cmd, platform, target, namespace, ExecOpts{Stdin: true})
	if err != nil {
		return "", err
	}
	argv = append(argv, ContainerShell, "-c", "mkdir -p "+dir+" && cat > "+path)
	return r.Run(Cmd{Argv: argv, Stdin: script})
}

// RunStatusScript runs the already-installed status script (`sh <path>`) and
// returns its combined output.
//
// The script always exits 0 and reports through its output -- the leader-election
// state on stdout, anything that went wrong on stderr -- so a non-zero exit
// here did not come from the script and means the exec itself failed (target
// gone, engine unreachable). Callers should treat it as a failed target rather
// than as a state. Output is returned alongside any error so whatever the
// engine printed is available for the message.
//
// path is a tool-authored constant, never operator input. target and
// namespace are operator-supplied and must already be validated by the
// caller.
func RunStatusScript(r Runner, cmd []string, platform, target, namespace, path string) (string, error) {
	argv, err := ExecArgv(cmd, platform, target, namespace, ExecOpts{})
	if err != nil {
		return "", err
	}
	argv = append(argv, ContainerShell, path)
	return r.Run(Cmd{Argv: argv})
}
