// Package runner executes the deploy/remove actions for each target by shelling
// out to the configured CLI (kubectl/oc, docker, podman/systemctl) through an
// argv slice -- never a shell string, so no token is ever re-interpreted by a
// shell. Command strings from env.yaml are tokenised and ParseCommand checks
// every token with the shared guard (execguard.Check, through
// validate.CommandRules) before a process is started: the charset, the
// per-platform binary allowlist and the flag-shape rules. Preflight runs a
// read-only login/daemon probe before any mutating call. Every command crosses
// the shared runner (github.com/hafio/go-deploy-utility/runner), whose Exec
// resolves argv[0] through PATH, refusing a current-directory or other relative
// resolution.
//
// The package is deliberately thin: rendering lives in gen/deploy/dockergen/
// podmangen, and the CLI owns path resolution. runner only parses commands,
// writes files (secrets 0600, never logged), and runs argv slices via an
// injected shared Runner so the exec boundary is unit-testable. Every helper
// that runs a command takes the caller's context first.
package runner

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/hafio/go-deploy-utility/execguard"
	"github.com/hafio/go-deploy-utility/exitcode"
	gdurunner "github.com/hafio/go-deploy-utility/runner"
	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/spec"
	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/validate"
)

// preflightHints gives an actionable next step per platform when Preflight's
// read-only probe fails, keyed by the same platform constants CommandRules
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
// canIProbe the can-i permission of the same name, Docker compose's "down".
const (
	ActionDeploy = "deploy"
	ActionRemove = "remove"

	// The lifecycle actions act on an instance deploy already created, in
	// place: nothing is rendered or applied, and nothing deploy made is removed.
	ActionStart   = "start"
	ActionStop    = "stop"
	ActionRestart = "restart"

	// The read-only verbs change nothing; they name only the permission their
	// preflight asks for, so a view-only operator is not refused a write it
	// never makes.
	ActionStatus = "status"
	ActionLogs   = "logs"
	ActionCLI    = "cli"
)

// runParsed is how every helper that parses a command's output runs it: with the
// two streams kept apart, so a warning on stderr cannot corrupt the answer.
//
// The returned string is what the caller parses. On failure the error carries
// whatever the command said, from whichever stream it said it on, so a failure
// message is no less useful than it was when the two were merged.
func runParsed(ctx context.Context, r gdurunner.Runner, c gdurunner.Cmd) (string, error) {
	out, errOut, err := gdurunner.Split(ctx, r, c)
	if err != nil {
		return strings.TrimSpace(out + "\n" + errOut), err
	}
	return out, nil
}

// ParseCommand splits a command string into an argv slice on whitespace (the
// same split the validator uses, so a command that passed validation parses
// identically here) and runs the shared guard over it: every token passes the
// charset, argv[0] must be a bare name on the platform's allowlist or among the
// operator's approvals, and every later token must be flag-shaped. A privilege
// escalation wrapper is refused whatever the approvals say. Quoted arguments
// containing spaces are intentionally unsupported: the split tears them apart
// before the guard sees them.
//
// A refusal is marked as a usage error, because what it refuses came from the
// env file or the command line, so it is exit 2 wherever it surfaces.
func ParseCommand(platform, cmd string, approvals execguard.Approvals) ([]string, error) {
	tokens := strings.Fields(cmd)
	if err := execguard.Check(validate.CommandRules(platform), tokens, approvals); err != nil {
		return nil, exitcode.AsUsage(err)
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
// `<command> apply -f -` / `<command> delete -f - --ignore-not-found` with the
// manifest on stdin, so the manifest bytes never appear on a command line.
//
// A teardown names every object the instance can own -- the stores Secret
// always, whether or not this spec renders it -- so an object already gone
// is the expected case, not a failure that would stop remove before the
// namespace step.
func Kubernetes(ctx context.Context, r gdurunner.Runner, command, action, manifest string, approvals execguard.Approvals) (string, error) {
	argv, err := ParseCommand(validate.PlatformKubernetes, command, approvals)
	if err != nil {
		return "", err
	}
	verb, err := kubeVerb(action)
	if err != nil {
		return "", err
	}
	argv = append(argv, verb, "-f", "-")
	if action == ActionRemove {
		argv = append(argv, "--ignore-not-found")
	}
	return gdurunner.Combined(ctx, r, gdurunner.Cmd{Argv: argv, Stdin: []byte(manifest)})
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

// Docker runs `<command> compose -f - --project-directory <projectDir> up -d`
// (deploy) or `... down` (remove), handing compose the file on stdin. The file
// inlines application.yml, which carries the status account's password, so it
// is never written to disk -- where it sat beside env.yaml, readable by whoever
// could read that directory, and, kept by a failed run, was read as a workflow
// by the next one (the default workflows.dir is env.yaml's own).
//
// projectDir is env.yaml's directory, the one compose took from the file's
// location when it was written there, so an .env file and any relative path
// resolve as before.
//
// env carries the credential values the compose file's environment-provider
// secrets read, in this child process only -- which is why no secret material is
// ever written to disk for docker. It is passed on both actions: `down` resolves
// the same secret declarations `up` did.
func Docker(ctx context.Context, r gdurunner.Runner, command, action, compose, projectDir string, env []string, approvals execguard.Approvals) (string, error) {
	argv, err := ParseCommand(validate.PlatformDocker, command, approvals)
	if err != nil {
		return "", err
	}
	argv = append(argv, "compose", "-f", "-", "--project-directory", projectDir)
	switch action {
	case ActionDeploy:
		argv = append(argv, "up", "-d")
	case ActionRemove:
		argv = append(argv, "down")
	default:
		return "", fmt.Errorf("unknown action %q (want %q or %q)", action, ActionDeploy, ActionRemove)
	}
	return gdurunner.Combined(ctx, r, gdurunner.Cmd{Argv: argv, Stdin: []byte(compose), Env: env})
}

// KubernetesLifecycle stops, starts or restarts a Deployment in place: stop
// scales it to 0, start scales it back to replicas (the env.yaml value), and
// restart is `rollout restart`, which replaces the pods one by one as a
// rollout does.
func KubernetesLifecycle(ctx context.Context, r gdurunner.Runner, command, action, namespace, name string, replicas int, approvals execguard.Approvals) (string, error) {
	argv, err := ParseCommand(validate.PlatformKubernetes, command, approvals)
	if err != nil {
		return "", err
	}
	if namespace != "" {
		argv = append(argv, "--namespace", namespace)
	}
	switch action {
	case ActionStop:
		argv = append(argv, "scale", "deployment/"+name, "--replicas=0")
	case ActionStart:
		argv = append(argv, "scale", "deployment/"+name, "--replicas="+strconv.Itoa(replicas))
	case ActionRestart:
		argv = append(argv, "rollout", "restart", "deployment/"+name)
	default:
		return "", fmt.Errorf("unknown lifecycle action %q", action)
	}
	return gdurunner.Combined(ctx, r, gdurunner.Cmd{Argv: argv})
}

// DockerLifecycle runs `<command> start|stop|restart <container>` on the
// container a deploy created.
func DockerLifecycle(ctx context.Context, r gdurunner.Runner, command, action, container string, approvals execguard.Approvals) (string, error) {
	argv, err := ParseCommand(validate.PlatformDocker, command, approvals)
	if err != nil {
		return "", err
	}
	switch action {
	case ActionStart, ActionStop, ActionRestart:
	default:
		return "", fmt.Errorf("unknown lifecycle action %q", action)
	}
	return gdurunner.Combined(ctx, r, gdurunner.Cmd{Argv: append(argv, action, container)})
}

// PodmanLifecycle runs `systemctl [--user] start|stop|restart <service>` on
// the quadlet unit a deploy installed: under quadlet, systemd owns the
// container, so a podman stop would only be undone by the unit's restart
// policy, and a stopped unit's container is removed and created again from the
// unit on start.
func PodmanLifecycle(ctx context.Context, r gdurunner.Runner, sc QuadletScope, action, service string) (string, error) {
	switch action {
	case ActionStart, ActionStop, ActionRestart:
	default:
		return "", fmt.Errorf("unknown lifecycle action %q", action)
	}
	return gdurunner.Combined(ctx, r, gdurunner.Cmd{Argv: sc.systemctlArgs(action, service)})
}

// DockerImageCommand reads an image's ENTRYPOINT and CMD from the local engine
// (`<command> image inspect`), for a deploy that overrides the entrypoint to
// check that it hands over to the command the image really runs. found is
// false when the engine cannot answer -- above all when it does not have the
// image yet -- so the caller pulls it and asks again (DockerImagePull). The
// answer is read from stdout alone (runParsed): a warning the CLI prints on
// stderr must not turn a reply into "no answer".
func DockerImageCommand(ctx context.Context, r gdurunner.Runner, command, image string, approvals execguard.Approvals) (entrypoint, cmd []string, found bool, err error) {
	argv, err := ParseCommand(validate.PlatformDocker, command, approvals)
	if err != nil {
		return nil, nil, false, err
	}
	argv = append(argv, "image", "inspect", "--format", "{{json .Config.Entrypoint}}\n{{json .Config.Cmd}}", image)
	out, rerr := runParsed(ctx, r, gdurunner.Cmd{Argv: argv})
	if rerr != nil {
		return nil, nil, false, nil
	}
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) != 2 || json.Unmarshal([]byte(lines[0]), &entrypoint) != nil || json.Unmarshal([]byte(lines[1]), &cmd) != nil {
		return nil, nil, false, nil
	}
	return entrypoint, cmd, true, nil
}

// DockerImagePull pulls image through the engine (`<command> image pull`), for
// a deploy that has to see how an image starts before compose runs it. The
// pull prints its progress on both streams, so it runs through Run and its
// output is reported, never parsed.
func DockerImagePull(ctx context.Context, r gdurunner.Runner, command, image string, approvals execguard.Approvals) (string, error) {
	argv, err := ParseCommand(validate.PlatformDocker, command, approvals)
	if err != nil {
		return "", err
	}
	return gdurunner.Combined(ctx, r, gdurunner.Cmd{Argv: append(argv, "image", "pull", image)})
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
func Preflight(ctx context.Context, r gdurunner.Runner, platform, command, action, namespace string, approvals execguard.Approvals) (string, error) {
	argv, err := ParseCommand(platform, command, approvals)
	if err != nil {
		return "", err
	}
	var probe canI
	switch platform {
	case validate.PlatformKubernetes:
		var perr error
		if probe, perr = canIProbe(action); perr != nil {
			return "", perr
		}
		argv = append(argv, "auth", "can-i", probe.verb, probe.resource)
		if probe.subresource != "" {
			argv = append(argv, "--subresource", probe.subresource)
		}
		if namespace != "" {
			argv = append(argv, "--namespace", namespace)
		}
	case validate.PlatformDocker, validate.PlatformPodman:
		argv = append(argv, "info")
	default:
		return "", fmt.Errorf("unknown platform %q (want %q, %q, or %q)", platform, validate.PlatformKubernetes, validate.PlatformDocker, validate.PlatformPodman)
	}
	out, runErr := gdurunner.Combined(ctx, r, gdurunner.Cmd{Argv: argv})
	if runErr != nil {
		// can-i exits 1 both when it cannot reach the cluster and when the
		// answer is no; only the answer tells them apart, and a refusal is not
		// fixed by logging in again.
		if answer, denied := canIDenied(out); platform == validate.PlatformKubernetes && denied {
			where := "the current namespace"
			if namespace != "" {
				where = fmt.Sprintf("namespace %q", namespace)
			}
			return "", fmt.Errorf("preflight failed for %s: your account may not %s in %s (kubectl auth can-i answered %q): ask for that permission, or select a context that has it", platform, probe, where, answer)
		}
		hint, ok := preflightHints[platform]
		if !ok {
			return out, fmt.Errorf("preflight failed for %s: %w", platform, runErr)
		}
		return out, fmt.Errorf("preflight failed for %s: %w\n%s", platform, runErr, hint)
	}
	return out, nil
}

// canI is one permission `kubectl auth can-i` is asked about.
type canI struct{ verb, resource, subresource string }

// String spells the permission the way RBAC rules name it: "get pods/log".
func (c canI) String() string {
	if c.subresource != "" {
		return c.verb + " " + c.resource + "/" + c.subresource
	}
	return c.verb + " " + c.resource
}

// canIProbe maps an action to the permission its preflight asks about. The
// verb is can-i's own -- create/delete, not apply/delete, the manifest verbs
// kubeVerb returns -- because can-i asks about a permission, not a subcommand.
// A mutating action asks about the Deployment it changes; a read-only verb asks
// for the one thing it cannot work without. A subresource goes in
// --subresource, since "pods/log" would name a pod called log.
func canIProbe(action string) (canI, error) {
	switch action {
	case ActionDeploy:
		return canI{"create", "deployment", ""}, nil
	case ActionRemove:
		return canI{"delete", "deployment", ""}, nil
	case ActionStart, ActionStop, ActionRestart:
		// scale and rollout restart both patch the Deployment.
		return canI{"patch", "deployment", ""}, nil
	case ActionStatus:
		// Finding the pods is what status cannot do without; a pod it may not
		// exec into is reported against that pod.
		return canI{"get", "pods", ""}, nil
	case ActionLogs:
		return canI{"get", "pods", "log"}, nil
	case ActionCLI:
		return canI{"create", "pods", "exec"}, nil
	default:
		return canI{}, fmt.Errorf("unknown action %q", action)
	}
}

// canIDenied reports whether can-i's output is a "no" -- alone, or with the
// reason some authorizers add -- and returns that line.
func canIDenied(out string) (string, bool) {
	for _, line := range strings.Split(out, "\n") {
		if l := strings.TrimSpace(line); l == "no" || strings.HasPrefix(l, "no - ") {
			return l, true
		}
	}
	return "", false
}

// PodmanVersion is the podman client's version, as `podman version --format
// {{.Client.Version}}` prints it (5.8.2), read from stdout alone.
func PodmanVersion(ctx context.Context, r gdurunner.Runner, command string, approvals execguard.Approvals) (string, error) {
	argv, err := ParseCommand(validate.PlatformPodman, command, approvals)
	if err != nil {
		return "", err
	}
	out, err := runParsed(ctx, r, gdurunner.Cmd{Argv: append(argv, "version", "--format", "{{.Client.Version}}")})
	if err != nil {
		return "", fmt.Errorf("podman version: %w", err)
	}
	return strings.TrimSpace(out), nil
}

// PodmanSupportsReplace reports whether a podman version is 4.7 or newer,
// where `secret create --replace` exists. A version that does not parse is
// not, so the caller keeps to remove-then-create.
func PodmanSupportsReplace(version string) bool {
	num := func(s string) (int, bool) {
		end := 0
		for end < len(s) && s[end] >= '0' && s[end] <= '9' {
			end++
		}
		n, err := strconv.Atoi(s[:end])
		return n, err == nil
	}
	parts := strings.SplitN(strings.TrimPrefix(version, "v"), ".", 3)
	if len(parts) < 2 {
		return false
	}
	major, ok1 := num(parts[0])
	minor, ok2 := num(parts[1])
	return ok1 && ok2 && (major > 4 || major == 4 && minor >= 7)
}

// PodmanSecretCreate stores one credential in podman's secret store, replacing
// any previous entry of that name.
//
// The value crosses on stdin, never in argv. With replace -- podman 4.7 and
// newer, PodmanSupportsReplace -- it is one `secret create --replace`, so a
// failed create leaves the previous secret, and the unit still mounting it,
// intact. Below 4.7, where the rest of the quadlet wiring still works, it is a
// remove-then-create: a create that fails then leaves the installed unit
// naming a secret that is gone until deploy is run again.
func PodmanSecretCreate(ctx context.Context, r gdurunner.Runner, command, name, value string, replace bool, approvals execguard.Approvals) (string, error) {
	argv, err := ParseCommand(validate.PlatformPodman, command, approvals)
	if err != nil {
		return "", err
	}
	if replace {
		out, err := gdurunner.Combined(ctx, r, gdurunner.Cmd{Argv: append(append([]string(nil), argv...), "secret", "create", "--replace", name, "-"), Stdin: []byte(value)})
		if err != nil {
			return out, fmt.Errorf("podman secret create --replace %s: %w", name, err)
		}
		return out, nil
	}
	out, err := gdurunner.Combined(ctx, r, gdurunner.Cmd{Argv: append(append([]string(nil), argv...), "secret", "rm", "--ignore", name)})
	if err != nil {
		return out, fmt.Errorf("podman secret rm %s: %w", name, err)
	}
	o, err := gdurunner.Combined(ctx, r, gdurunner.Cmd{Argv: append(append([]string(nil), argv...), "secret", "create", name, "-"), Stdin: []byte(value)})
	out += o
	if err != nil {
		return out, fmt.Errorf("podman secret create %s: %w", name, err)
	}
	return out, nil
}

// PodmanSecretRemove deletes the named secrets, ignoring any that are already
// gone so a repeated delete stays quiet.
func PodmanSecretRemove(ctx context.Context, r gdurunner.Runner, command string, names []string, approvals execguard.Approvals) (string, error) {
	if len(names) == 0 {
		return "", nil
	}
	argv, err := ParseCommand(validate.PlatformPodman, command, approvals)
	if err != nil {
		return "", err
	}
	argv = append(argv, "secret", "rm", "--ignore")
	out, err := gdurunner.Combined(ctx, r, gdurunner.Cmd{Argv: append(argv, names...)})
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

// ResolveQuadletScope derives both the unit directory and the systemctl mode
// from the effective uid of whoever is running: system for root, user otherwise.
//
// It takes no arguments on purpose. These are the only two combinations an
// operator can both write to and start -- root owns /etc/containers/systemd and
// drives the system systemd, everyone else owns their own home and drives their
// own instance -- and systemd loads units only from its own generator
// directories, so there is no third directory worth naming. Making it
// unconfigurable is what turns a permission mismatch from something to validate
// against into something that cannot be expressed.
func ResolveQuadletScope() (QuadletScope, error) {
	var sc QuadletScope
	sc.UserMode = os.Geteuid() != 0
	if !sc.UserMode {
		sc.Dir = quadletSystem
		return sc, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return QuadletScope{}, fmt.Errorf("resolving user quadlet dir: %w", err)
	}
	sc.Dir = filepath.Join(home, quadletUserSub)
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

// PodmanDeploy reloads the systemd generator and brings up one service per unit.
// Unit files must already be written into sc.Dir, and the secrets they mount
// already be in podman's store, by the caller. services are the systemd unit
// names, e.g. "solmq-connector.service".
//
// A service that is already running is restarted rather than started: podman
// copies a mount secret into the container when it creates it, so a running
// container would keep the old application.yml and stores, and start on an
// active unit does nothing. Anything else -- a new unit, a stopped or a failed
// one -- is started.
func PodmanDeploy(ctx context.Context, r gdurunner.Runner, sc QuadletScope, services []string) (string, error) {
	var out strings.Builder
	o, err := gdurunner.Combined(ctx, r, gdurunner.Cmd{Argv: sc.systemctlArgs("daemon-reload")})
	out.WriteString(o)
	if err != nil {
		return out.String(), fmt.Errorf("systemctl daemon-reload: %w", err)
	}
	for _, s := range services {
		verb := "start"
		// is-active exits 0 only for an active unit. Any failure -- inactive,
		// failed, unknown, or systemctl itself erroring -- means start, whose own
		// error then says what is wrong.
		if _, aerr := gdurunner.Combined(ctx, r, gdurunner.Cmd{Argv: sc.systemctlArgs("is-active", "--quiet", s)}); aerr == nil {
			verb = "restart"
			out.WriteString(s + " was running; restarting it to load the new configuration\n")
		}
		o, err = gdurunner.Combined(ctx, r, gdurunner.Cmd{Argv: sc.systemctlArgs(verb, s)})
		out.WriteString(o)
		if err != nil {
			return out.String(), fmt.Errorf("systemctl %s %s: %w", verb, s, err)
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
func PodmanRemove(ctx context.Context, r gdurunner.Runner, sc QuadletScope, services, units []string) (string, error) {
	var out strings.Builder
	for _, s := range services {
		o, err := gdurunner.Combined(ctx, r, gdurunner.Cmd{Argv: sc.systemctlArgs("stop", s)})
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
	o, err := gdurunner.Combined(ctx, r, gdurunner.Cmd{Argv: sc.systemctlArgs("daemon-reload")})
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
func KubernetesPodsJSON(ctx context.Context, r gdurunner.Runner, cmd []string, namespace, selector string, names []string, allNamespaces bool) (string, error) {
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
	out, err := runParsed(ctx, r, gdurunner.Cmd{Argv: argv})
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
func KubernetesGetJSON(ctx context.Context, r gdurunner.Runner, cmd []string, namespace, kind, name string) (string, error) {
	argv := append(append([]string(nil), cmd...), "get", kind, name)
	if namespace != "" {
		argv = append(argv, "-n", namespace)
	}
	argv = append(argv, "-o", "json")
	out, err := runParsed(ctx, r, gdurunner.Cmd{Argv: argv})
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
func KubernetesListJSON(ctx context.Context, r gdurunner.Runner, cmd []string, namespace, types string) (string, error) {
	argv := append(append([]string(nil), cmd...), "get", types)
	if namespace != "" {
		argv = append(argv, "-n", namespace)
	}
	argv = append(argv, "-o", "json")
	out, err := runParsed(ctx, r, gdurunner.Cmd{Argv: argv})
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
func KubernetesTop(ctx context.Context, r gdurunner.Runner, cmd []string, namespace, selector string, names []string, allNamespaces bool) (string, error) {
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
	out, err := runParsed(ctx, r, gdurunner.Cmd{Argv: argv})
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
func EngineInspectJSON(ctx context.Context, r gdurunner.Runner, cmd []string, names []string) (string, error) {
	if len(names) == 0 {
		return "", fmt.Errorf("no container named to inspect")
	}
	argv := append(append([]string(nil), cmd...), "inspect")
	out, err := runParsed(ctx, r, gdurunner.Cmd{Argv: append(argv, names...)})
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
func EngineImageInspectJSON(ctx context.Context, r gdurunner.Runner, cmd []string, ref string) (string, error) {
	argv := append(append([]string(nil), cmd...), "image", "inspect", ref)
	out, err := runParsed(ctx, r, gdurunner.Cmd{Argv: argv})
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
func EngineStats(ctx context.Context, r gdurunner.Runner, cmd []string, names []string) (string, error) {
	argv := append(append([]string(nil), cmd...), "stats", "--no-stream", "--format", engineStatsFormat)
	out, err := runParsed(ctx, r, gdurunner.Cmd{Argv: append(argv, names...)})
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
func EngineList(ctx context.Context, r gdurunner.Runner, cmd []string) (string, error) {
	argv := append(append([]string(nil), cmd...), "ps", "--all", "--no-trunc", "--format", engineListFormat)
	out, err := runParsed(ctx, r, gdurunner.Cmd{Argv: argv})
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
func SystemctlNRestarts(ctx context.Context, r gdurunner.Runner, sc QuadletScope, unit string) (int, error) {
	out, err := runParsed(ctx, r, gdurunner.Cmd{Argv: sc.systemctlArgs("show", unit, "-p", "NRestarts", "--value")})
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
func ScriptInstalled(ctx context.Context, r gdurunner.Runner, cmd []string, platform, target, namespace, path string) (bool, error) {
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
	out, runErr := gdurunner.Combined(ctx, r, gdurunner.Cmd{Argv: argv})
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
func InstallScript(ctx context.Context, r gdurunner.Runner, cmd []string, platform, target, namespace, dir, path, script string) (string, error) {
	argv, err := ExecArgv(cmd, platform, target, namespace, ExecOpts{Stdin: true})
	if err != nil {
		return "", err
	}
	argv = append(argv, ContainerShell, "-c", "mkdir -p "+dir+" && cat > "+path)
	return gdurunner.Combined(ctx, r, gdurunner.Cmd{Argv: argv, Stdin: []byte(script)})
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
func RunStatusScript(ctx context.Context, r gdurunner.Runner, cmd []string, platform, target, namespace, path string) (string, error) {
	argv, err := ExecArgv(cmd, platform, target, namespace, ExecOpts{})
	if err != nil {
		return "", err
	}
	argv = append(argv, ContainerShell, path)
	return gdurunner.Combined(ctx, r, gdurunner.Cmd{Argv: argv})
}
