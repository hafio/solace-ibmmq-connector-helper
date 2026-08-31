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
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

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
	cmd := exec.Command(resolved, c.Argv[1:]...) //nolint:gosec // argv tokens are SafeToken-validated upstream; no shell is involved
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
	runErr := cmd.Run()
	return buf.String(), runErr
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
	out, err := r.Run(Cmd{Argv: argv})
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
	out, err := r.Run(Cmd{Argv: argv})
	if err != nil {
		return out, fmt.Errorf("reading %s/%s: %w\n%s", kind, name, err, out)
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
	out, err := r.Run(Cmd{Argv: argv})
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
	out, err := r.Run(Cmd{Argv: append(argv, names...)})
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
	out, err := r.Run(Cmd{Argv: argv})
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
	out, err := r.Run(Cmd{Argv: append(argv, names...)})
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
	out, err := r.Run(Cmd{Argv: argv})
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
	out, err := r.Run(Cmd{Argv: sc.systemctlArgs("show", unit, "-p", "NRestarts", "--value")})
	if err != nil {
		return 0, fmt.Errorf("reading %s restart count from systemd: %w\n%s", unit, err, out)
	}
	n, cerr := strconv.Atoi(strings.TrimSpace(out))
	if cerr != nil {
		return 0, fmt.Errorf("reading %s restart count from systemd: unexpected answer %q", unit, strings.TrimSpace(out))
	}
	return n, nil
}

// execArgv builds the `<cmd> exec` argv prefix shared by ScriptInstalled,
// InstallScript, and RunStatusScript, keeping the per-platform shape in one
// place instead of repeating it in each helper.
//
// kubectl/oc need the namespace as a flag and a "--" separator before the
// in-container command; docker and podman exec take the command directly
// after the target, with no namespace concept and no separator. interactive
// adds -i, needed only when the caller pipes something on stdin.
//
// target and namespace are operator-supplied and must already be validated by
// the caller before reaching this function.
func execArgv(cmd []string, platform, target, namespace string, interactive bool) ([]string, error) {
	argv := append(append([]string(nil), cmd...), "exec")
	if interactive {
		argv = append(argv, "-i")
	}
	switch platform {
	case validate.PlatformKubernetes:
		argv = append(argv, target)
		if namespace != "" {
			argv = append(argv, "-n", namespace)
		}
		argv = append(argv, "--")
	case validate.PlatformDocker, validate.PlatformPodman:
		argv = append(argv, target)
	default:
		return nil, fmt.Errorf("unknown platform %q (want %q, %q, or %q)", platform, validate.PlatformKubernetes, validate.PlatformDocker, validate.PlatformPodman)
	}
	return argv, nil
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
	argv, err := execArgv(cmd, platform, target, namespace, false)
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
	argv = append(argv, "sh", "-c", "if [ -f "+path+" ]; then echo "+ScriptPresentMarker+"; else echo "+ScriptAbsentMarker+"; fi")
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
	argv, err := execArgv(cmd, platform, target, namespace, true)
	if err != nil {
		return "", err
	}
	argv = append(argv, "sh", "-c", "mkdir -p "+dir+" && cat > "+path)
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
	argv, err := execArgv(cmd, platform, target, namespace, false)
	if err != nil {
		return "", err
	}
	argv = append(argv, "sh", path)
	return r.Run(Cmd{Argv: argv})
}
