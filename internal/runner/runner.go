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
	"runtime"
	"strconv"
	"strings"

	"github.com/hafio/go-deploy-utility/containercli"
	"github.com/hafio/go-deploy-utility/containercli/inspect"
	"github.com/hafio/go-deploy-utility/execguard"
	"github.com/hafio/go-deploy-utility/exitcode"
	"github.com/hafio/go-deploy-utility/kube"
	"github.com/hafio/go-deploy-utility/kube/nsgate"
	"github.com/hafio/go-deploy-utility/kube/rbac"
	"github.com/hafio/go-deploy-utility/podman"
	"github.com/hafio/go-deploy-utility/rootless"
	gdurunner "github.com/hafio/go-deploy-utility/runner"
	"github.com/hafio/go-deploy-utility/systemd"
	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/spec"
	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/validate"
)

// safeOut is an engine's output as it may be appended to an error: every line
// has its control characters dropped (inspect.Printable), so a terminal escape
// in a container name or an error text the engine echoes is never acted on by
// the terminal that prints it. The line breaks themselves are kept.
func safeOut(out string) string {
	lines := strings.Split(out, "\n")
	for i, l := range lines {
		lines[i] = inspect.Printable(l)
	}
	return strings.Join(lines, "\n")
}

// Deploy actions shared by every target. These are the CLI verbs, and they reach
// the operator verbatim (report's "ok: remove kubernetes", runAction's flag-error
// prefix), so they stay spelled as the verbs are. The platform subcommands they
// map to are separate literals -- kube.Delete is kubectl's own "delete", the
// probe's verb the can-i permission of the same name, Docker compose's "down".
const (
	ActionDeploy = "deploy"
	ActionRemove = "remove"

	// The lifecycle actions act on an instance deploy already created, in
	// place: nothing is rendered or applied, and nothing deploy made is removed.
	ActionStart   = "start"
	ActionStop    = "stop"
	ActionRestart = "restart"

	// The read-only verbs change nothing and ask kubernetes for no permission
	// ahead of time, so a view-only operator is never refused a write it never
	// makes; what they cannot read, the command itself reports.
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

// Kubernetes applies (deploy) or deletes (remove) the manifest through kube.Apply
// (`apply -f -`) or kube.Delete (`delete -f - --ignore-not-found`), with the
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
	var c gdurunner.Cmd
	switch action {
	case ActionDeploy:
		c = kube.Apply(argv, []byte(manifest))
	case ActionRemove:
		c = kube.Delete(argv, []byte(manifest))
	default:
		return "", fmt.Errorf("unknown action %q (want %q or %q)", action, ActionDeploy, ActionRemove)
	}
	return gdurunner.Combined(ctx, r, c)
}

// Docker runs `<command> compose -f - --project-directory <projectDir> up -d`
// (deploy) or `... down` (remove), handing compose the file on stdin
// (containercli.ComposeUp and ComposeDown). The file inlines application.yml,
// which carries the status account's password, so it is never written to disk --
// where it sat beside env.yaml, readable by whoever could read that directory,
// and, kept by a failed run, was read as a workflow by the next one (the default
// workflows.dir is env.yaml's own).
//
// Deploy never forces a recreate: compose decides, from the file's own
// configuration hash, which carries the deploy digest label (see
// dockergen.DeployDigestVar), so a changed file or credential recreates the
// container and an unchanged one is left running.
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
	base := append(argv, "compose")
	var c gdurunner.Cmd
	switch action {
	case ActionDeploy:
		c = containercli.ComposeUp(base, []byte(compose), projectDir, env, false)
	case ActionRemove:
		c = containercli.ComposeDown(base, []byte(compose), projectDir, env)
	default:
		return "", fmt.Errorf("unknown action %q (want %q or %q)", action, ActionDeploy, ActionRemove)
	}
	return gdurunner.Combined(ctx, r, c)
}

// DockerDeployDigest reads what a redeploy needs to know before compose runs:
// whether the container is running (`ps --all --no-trunc --filter status=running`,
// matched by exact name in Go), and, when it is, the value of its label key -- the
// deploy digest the last deploy stamped on it. A label that is absent, or an
// inspect that fails, reads as "" (digest unknown), which the caller treats as a
// change.
//
// A listing that cannot be read is an error: with no answer the deploy cannot tell
// whether it would recreate a running service, and so cannot ask.
func DockerDeployDigest(ctx context.Context, r gdurunner.Runner, command, name, key string, approvals execguard.Approvals) (running bool, digest string, err error) {
	argv, err := ParseCommand(validate.PlatformDocker, command, approvals)
	if err != nil {
		return false, "", err
	}
	out, err := EngineList(ctx, r, argv, true)
	if err != nil {
		return false, "", fmt.Errorf("checking whether %s is running before deploying over it: %w", name, err)
	}
	if !containercli.HasName(containercli.ParseList(out), name) {
		return false, "", nil
	}
	doc, ierr := EngineInspectJSON(ctx, r, argv, []string{name})
	if ierr != nil {
		return true, "", nil
	}
	cs, derr := inspect.Decode(doc)
	if derr != nil {
		return true, "", nil
	}
	for _, c := range cs {
		if c.Name == name {
			return true, c.Labels[key], nil
		}
	}
	return true, "", nil
}

// KubernetesLifecycle stops, starts or restarts a Deployment in place: stop
// scales it to 0, start scales it back to replicas (the env.yaml value), and
// restart is `rollout restart`, which replaces the pods one by one as a
// rollout does (kube.Scale and kube.RolloutRestart).
func KubernetesLifecycle(ctx context.Context, r gdurunner.Runner, command, action, namespace, name string, replicas int, approvals execguard.Approvals) (string, error) {
	argv, err := ParseCommand(validate.PlatformKubernetes, command, approvals)
	if err != nil {
		return "", err
	}
	var c gdurunner.Cmd
	switch action {
	case ActionStop:
		c = kube.Scale(argv, "deployment", name, namespace, 0)
	case ActionStart:
		c = kube.Scale(argv, "deployment", name, namespace, replicas)
	case ActionRestart:
		c = kube.RolloutRestart(argv, "deployment", name, namespace)
	default:
		return "", fmt.Errorf("unknown lifecycle action %q", action)
	}
	return gdurunner.Combined(ctx, r, c)
}

// KubernetesRolloutStatus waits for the Deployment to roll out (kube.RolloutStatus,
// kube.RolloutTimeout) after a deploy, start or restart. The wait is a convenience:
// the caller turns a refusal (kube.Forbidden of the output) into a warning, and a
// timeout or a failed rollout into an error. The output is stdout and stderr merged,
// since both carry progress and the refusal.
func KubernetesRolloutStatus(ctx context.Context, r gdurunner.Runner, command, namespace, name string, approvals execguard.Approvals) (string, error) {
	argv, err := ParseCommand(validate.PlatformKubernetes, command, approvals)
	if err != nil {
		return "", err
	}
	return gdurunner.Combined(ctx, r, kube.RolloutStatus(argv, "deployment", name, namespace, ""))
}

// KubernetesNamespaceState reads whether namespace exists (kube.GetNamespace,
// stdout alone). A read that fails -- typically an identity that may not get
// namespaces -- is kube.NamespaceUnknown, never taken for absent.
func KubernetesNamespaceState(ctx context.Context, r gdurunner.Runner, base []string, namespace string) kube.NamespaceState {
	out, err := runParsed(ctx, r, kube.GetNamespace(base, namespace))
	if err != nil {
		return kube.NamespaceUnknown
	}
	return kube.ParseNamespace(out, namespace)
}

// KubernetesDeleteNamespace deletes the namespace (kube.DeleteNamespace). Run it
// only through the gate in nsgate order, which removeNamespace follows.
func KubernetesDeleteNamespace(ctx context.Context, r gdurunner.Runner, base []string, namespace string) (string, error) {
	return gdurunner.Combined(ctx, r, kube.DeleteNamespace(base, namespace))
}

// DockerLifecycle runs `<command> start|stop|restart <container>` on the
// container a deploy created.
func DockerLifecycle(ctx context.Context, r gdurunner.Runner, command, action, container string, approvals execguard.Approvals) (string, error) {
	argv, err := ParseCommand(validate.PlatformDocker, command, approvals)
	if err != nil {
		return "", err
	}
	var c gdurunner.Cmd
	switch action {
	case ActionStart:
		c = containercli.Start(argv, container)
	case ActionStop:
		c = containercli.Stop(argv, container)
	case ActionRestart:
		c = containercli.Restart(argv, container)
	default:
		return "", fmt.Errorf("unknown lifecycle action %q", action)
	}
	return gdurunner.Combined(ctx, r, c)
}

// PodmanLifecycle runs `systemctl [--user] start|stop|restart <service>` on
// the quadlet unit a deploy installed: under quadlet, systemd owns the
// container, so a podman stop would only be undone by the unit's restart
// policy, and a stopped unit's container is removed and created again from the
// unit on start.
func PodmanLifecycle(ctx context.Context, r gdurunner.Runner, sc QuadletScope, action, service string) (string, error) {
	var c gdurunner.Cmd
	switch action {
	case ActionStart:
		c = systemd.Start(sc.UserMode, service)
	case ActionStop:
		c = systemd.Stop(sc.UserMode, service)
	case ActionRestart:
		c = systemd.Restart(sc.UserMode, service)
	default:
		return "", fmt.Errorf("unknown lifecycle action %q", action)
	}
	return gdurunner.Combined(ctx, r, sc.WithEnv(c))
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

// Preflight runs cheap, read-only probes before any generated file is written
// or any mutating command runs, so a deploy/remove that can never succeed
// (not logged in, daemon down, a write the account may not make) stops before
// touching anything:
//
//   - kubernetes: `<argv> auth can-i <verb> <resource>` for each of probes, in
//     order, stopping at the first that is refused or unanswered (rbac.CanI and
//     rbac.Check; KubernetesProbes builds the set for an action). No probes, no
//     calls: the read verbs ask nothing ahead of time.
//   - docker/podman: `<argv> info` (containercli.Info).
//
// A non-nil error says what to do next: a refusal wraps *rbac.RefusedError (find it
// with errors.As) naming the permission to ask for, and an engine failure carries
// a platform-specific hint (log in, start the daemon, ...) so the operator gets a
// next step instead of a bare CLI error. rootless is whether a podman deployment
// runs rootless (the resolved quadlet scope's UserMode): it picks
// containercli.EngineHint's hint, and is ignored on every other platform.
func Preflight(ctx context.Context, r gdurunner.Runner, platform, command string, rootless bool, probes []rbac.Probe, approvals execguard.Approvals) (string, error) {
	argv, err := ParseCommand(platform, command, approvals)
	if err != nil {
		return "", err
	}
	var cmd gdurunner.Cmd
	switch platform {
	case validate.PlatformKubernetes:
		for _, p := range probes {
			if err := CheckProbe(ctx, r, argv, p); err != nil {
				return "", fmt.Errorf("preflight failed for %s: %w", platform, err)
			}
		}
		return "", nil
	case validate.PlatformDocker, validate.PlatformPodman:
		cmd = containercli.Info(argv)
	default:
		return "", fmt.Errorf("unknown platform %q (want %q, %q, or %q)", platform, validate.PlatformKubernetes, validate.PlatformDocker, validate.PlatformPodman)
	}
	out, runErr := gdurunner.Combined(ctx, r, cmd)
	if runErr != nil {
		// The engine's own output is in the message, so report prints it once.
		return "", containercli.InfoError(containercli.Engine(platform), argv, rootless, safeOut(out), runErr)
	}
	return out, nil
}

// CheckProbe asks one `auth can-i` question through base, the guarded kubectl argv,
// and reads the answer from stdout alone, so a warning on stderr cannot hide it
// (runner.Split, rbac.Check). nil is allowed; a refusal is a *rbac.RefusedError, and
// any other error means the question went unanswered.
func CheckProbe(ctx context.Context, r gdurunner.Runner, base []string, p rbac.Probe) error {
	out, errOut, runErr := gdurunner.Split(ctx, r, rbac.CanI(base, p))
	return rbac.Check(p, out, errOut, runErr)
}

// kindResources maps the kind of an object the tool writes to the plural resource
// `auth can-i` is asked about.
var kindResources = map[string]string{
	"deployment":            "deployments",
	"service":               "services",
	"configmap":             "configmaps",
	"secret":                "secrets",
	"persistentvolumeclaim": "persistentvolumeclaims",
}

// KubernetesProbes is the set of permissions an action asks for before its first
// write, and only what that write needs:
//
//   - deploy: create, and remove: delete, each resource of every object in written
//     (deploy.Written, so the probes and the owned objects cannot drift), once, plus
//     persistentvolumes when pv (libs.pvc.create). The Namespace is not among them:
//     the namespace flow asks for it only when it is absent.
//   - start, stop, restart: patch deployments in namespace (scale and rollout
//     restart both patch the Deployment).
//   - status, logs, cli: nothing. They read what they are allowed to read and report
//     what they are refused, so no up-front question stands between an operator and
//     a view.
//
// can-i's verbs are its own -- create/delete, not apply/delete, the manifest verbs --
// because can-i asks about a permission, not a subcommand.
func KubernetesProbes(action, namespace string, written []nsgate.Object, pv bool) ([]rbac.Probe, error) {
	switch action {
	case ActionDeploy, ActionRemove:
		verb := "create"
		if action == ActionRemove {
			verb = "delete"
		}
		var probes []rbac.Probe
		seen := map[string]bool{}
		add := func(resource string) {
			if !seen[resource] {
				seen[resource] = true
				probes = append(probes, rbac.Probe{Verb: verb, Resource: resource, Namespace: namespace})
			}
		}
		for _, o := range written {
			res, ok := kindResources[strings.ToLower(o.Kind)]
			if !ok {
				return nil, fmt.Errorf("no permission to ask about for the %s %s: add its resource to kindResources", o.Kind, o.Name)
			}
			add(res)
		}
		if pv {
			add("persistentvolumes")
		}
		return probes, nil
	case ActionStart, ActionStop, ActionRestart:
		return []rbac.Probe{{Verb: "patch", Resource: "deployments", Namespace: namespace}}, nil
	case ActionStatus, ActionLogs, ActionCLI:
		return nil, nil
	default:
		return nil, fmt.Errorf("unknown action %q", action)
	}
}

// PodmanVersion is the podman client's version, as `podman version --format
// {{.Client.Version}}` prints it (5.8.2), read from stdout alone. Pass it to
// podman.SupportsReplace, which ignores surrounding space.
func PodmanVersion(ctx context.Context, r gdurunner.Runner, command string, approvals execguard.Approvals) (string, error) {
	argv, err := ParseCommand(validate.PlatformPodman, command, approvals)
	if err != nil {
		return "", err
	}
	out, err := runParsed(ctx, r, podman.Version(argv))
	if err != nil {
		return "", fmt.Errorf("podman version: %w", err)
	}
	return strings.TrimSpace(out), nil
}

// PodmanSecretCreate stores one credential in podman's secret store, replacing
// any previous entry of that name.
//
// The value crosses on stdin, never in argv. With replace -- podman 4.7 and
// newer, podman.SupportsReplace -- it is one `secret create --replace`, so a
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
		out, err := gdurunner.Combined(ctx, r, podman.CreateSecret(argv, name, []byte(value), true))
		if err != nil {
			return out, fmt.Errorf("podman secret create --replace %s: %w", name, err)
		}
		return out, nil
	}
	rm, _ := podman.RemoveSecrets(argv, name)
	out, err := gdurunner.Combined(ctx, r, rm)
	if err != nil {
		return out, fmt.Errorf("podman secret rm %s: %w", name, err)
	}
	o, err := gdurunner.Combined(ctx, r, podman.CreateSecret(argv, name, []byte(value), false))
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
	rm, _ := podman.RemoveSecrets(argv, names...)
	out, err := gdurunner.Combined(ctx, r, rm)
	if err != nil {
		return out, fmt.Errorf("podman secret rm: %w", err)
	}
	return out, nil
}

// PodmanSecretMatches reports whether the named secret already holds value: it reads
// the stored value back (podman.SecretData, podman 4.7 and newer, which the caller
// has checked) from stdout alone and compares it exactly (podman.SecretMatches). The
// value stays inside this function, is never printed, and no error carries it. A
// missing secret, a probe that fails and a command the guard refuses all read as
// false: none of them proves the stored value is the one about to be written.
func PodmanSecretMatches(ctx context.Context, r gdurunner.Runner, command, name, value string, approvals execguard.Approvals) bool {
	argv, err := ParseCommand(validate.PlatformPodman, command, approvals)
	if err != nil {
		return false
	}
	c, err := podman.SecretData(argv, name)
	if err != nil {
		return false
	}
	out, _, rerr := gdurunner.Split(ctx, r, c)
	return rerr == nil && podman.SecretMatches(out, []byte(value))
}

// PodmanIDMap is podman's id mapping for the invoking user, its gid map when gid
// is set, read from stdout alone (see podman.IDMap). Ask only for a rootless
// deploy, since rootful podman reports no mappings.
func PodmanIDMap(ctx context.Context, r gdurunner.Runner, command string, gid bool, approvals execguard.Approvals) (string, error) {
	argv, err := ParseCommand(validate.PlatformPodman, command, approvals)
	if err != nil {
		return "", err
	}
	out, err := runParsed(ctx, r, podman.IDMap(argv, gid))
	if err != nil {
		return "", fmt.Errorf("podman info (id mappings): %w", err)
	}
	return out, nil
}

// ---- podman quadlet ----------------------------------------------------------

// QuadletScope is the resolved placement for podman quadlet units: the directory
// the .container files live in, whether systemctl runs in --user mode, and for
// a rootless scope the user session those commands need.
type QuadletScope struct {
	Dir      string
	UserMode bool

	// Euid is the effective uid the scope was resolved for, or -1 where there is
	// none (Windows). loginctl names it.
	Euid int

	// Env is the repaired user session as KEY=VALUE entries (empty when the
	// environment already carried it, and off Linux). It is attached to every
	// `systemctl --user` and loginctl command as Cmd.Env, never set process-wide.
	Env []string
}

// WithEnv attaches the scope's session to c. Only a rootless scope has one, so a
// system scope returns c untouched.
func (sc QuadletScope) WithEnv(c gdurunner.Cmd) gdurunner.Cmd {
	if sc.UserMode && len(sc.Env) > 0 {
		c.Env = append(append([]string(nil), c.Env...), sc.Env...)
	}
	return c
}

// ScopeEnv is what ResolveQuadletScope reads from the process, as values and
// functions so a test supplies its own and nothing is read from, or set in, the
// real environment.
type ScopeEnv struct {
	Euid   int                        // effective uid, -1 where there is none
	Getenv func(string) string        // os.Getenv
	Home   func() (string, error)     // os.UserHomeDir
	IsDir  func(string) (bool, error) // whether a directory exists; an error is a probe that could not run
	Linux  bool                       // the host runs systemd, so a rootless scope has a session to repair
}

// LiveScopeEnv is ScopeEnv read from this process.
func LiveScopeEnv() ScopeEnv {
	return ScopeEnv{
		Euid:   os.Geteuid(),
		Getenv: os.Getenv,
		Home:   os.UserHomeDir,
		IsDir:  statIsDir,
		Linux:  runtime.GOOS == "linux",
	}
}

// statIsDir reports whether path is an existing directory: a missing path is
// false with no error, and any other failure to look is an error.
func statIsDir(path string) (bool, error) {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return info.IsDir(), nil
}

// ResolveQuadletScope derives the unit directory, the systemctl mode and the
// user session from env and the podman.rootful setting.
//
// Rootless is the default: rootless.Decide refuses a non-root user who asks for
// rootful and root who does not. The directory is podman.QuadletDir, so an
// absolute XDG_CONFIG_HOME is honoured and a relative one is refused. There is
// no key for the directory: these are the only placements an operator can both
// write to and start, and systemd loads units only from its own generator
// directories.
//
// A refusal from Decide or QuadletDir is marked as a usage error (exit 2), since
// it is the operator's setting or environment that is wrong. A rootless scope on
// Linux then repairs the session sudo, su and cron strip (rootless.UserSession); a
// session that cannot be repaired is a plain error (exit 1), because the host is
// what is not ready.
func ResolveQuadletScope(env ScopeEnv, rootful bool) (QuadletScope, error) {
	rootlessMode, err := rootless.Decide(env.Euid, rootful, "podman.rootful")
	if err != nil {
		return QuadletScope{}, exitcode.AsUsage(err)
	}
	dir, err := podman.QuadletDir(rootlessMode, env.Getenv, env.Home)
	if err != nil {
		return QuadletScope{}, exitcode.AsUsage(err)
	}
	sc := QuadletScope{Dir: dir, UserMode: rootlessMode, Euid: env.Euid}
	if rootlessMode && env.Linux {
		session, err := rootless.UserSession(env.Euid, env.Getenv, env.IsDir)
		if err != nil {
			return QuadletScope{}, err
		}
		sc.Env = session.Env()
	}
	return sc, nil
}

// EnsureLinger makes sure the rootless user's manager outlives its login, so the
// quadlet unit survives logout and starts at boot. It asks logind, and when the
// answer is not "yes" enables linger for the user's own account, which needs no
// privilege. name is the user's login name, for the administrator form of the fix.
//
// A probe that cannot run is an error, since quadlets need logind; so is a refused
// enable, which names `sudo loginctl enable-linger <name>`. enabled reports that
// this call turned it on, for the caller to say so.
func EnsureLinger(ctx context.Context, r gdurunner.Runner, sc QuadletScope, name string) (enabled bool, err error) {
	state, err := ProbeLinger(ctx, r, sc)
	if err != nil {
		return false, err
	}
	if state == "yes" {
		return false, nil
	}
	cmd := sc.WithEnv(systemd.EnableLinger())
	if out, eerr := gdurunner.Combined(ctx, r, cmd); eerr != nil {
		return false, fmt.Errorf("linger is not enabled for %s (Linger=%s), and enabling it failed: %w\n%s\n"+
			"  Ask an administrator to run:\n"+
			"    sudo loginctl enable-linger %s",
			name, state, eerr, strings.TrimSpace(out), name)
	}
	return true, nil
}

// ProbeLinger is the user's linger state as logind reports it ("yes", "no"). A
// probe that cannot run is an error naming the command: a rootless quadlet unit
// needs logind, and a host without it cannot keep the unit up.
func ProbeLinger(ctx context.Context, r gdurunner.Runner, sc QuadletScope) (string, error) {
	cmd := sc.WithEnv(systemd.ShowLinger(sc.Euid))
	out, errOut, err := gdurunner.Split(ctx, r, cmd)
	if err != nil {
		return "", fmt.Errorf("cannot read the linger state with `%s`: %w\n%s\n"+
			"  Rootless quadlet units need systemd's logind: run on a systemd host, in the user's own login session",
			strings.Join(cmd.Argv, " "), err, strings.TrimSpace(out+"\n"+errOut))
	}
	return systemd.ParseLinger(out), nil
}

// PodmanDeploy brings up one service per unit after the caller has compared what
// this deploy would write with what is already installed. changed says that the
// unit or a secret differs (or is missing), and write puts the changed secrets and
// unit in place; it is not called when nothing changed. services are the systemd
// unit names, e.g. "solmq-connector.service".
//
// The order is the point: nothing is written before the answer.
//
//  1. Every service's state is read (is-active, whose stdout alone is used: it
//     exits non-zero for every state but an active one). A unit that is active or
//     reloading counts as running; any other answer -- inactive, failed, unknown,
//     or none because systemctl itself failed -- counts as not running, and start's
//     own error then says what is wrong.
//  2. When something changed and a service is running, approve is asked. An error
//     from it ends the deploy; a "no" ends it with cancelled set and the host
//     untouched, so the next deploy asks again.
//  3. write runs, then daemon-reload, only when something changed.
//  4. A service that is not running is started. A running one is restarted when
//     something changed, because podman copies a mount secret into the container
//     when it creates it, so a running container would keep the old application.yml
//     and stores, and start on an active unit does nothing. A running service
//     with nothing changed is left alone, and out says so.
func PodmanDeploy(ctx context.Context, r gdurunner.Runner, sc QuadletScope, services []string, changed bool, approve func() (bool, error), write func() (string, error)) (out string, cancelled bool, err error) {
	var b strings.Builder
	running := make(map[string]bool, len(services))
	anyRunning := false
	for _, s := range services {
		state, _, _ := gdurunner.Split(ctx, r, sc.WithEnv(systemd.IsActive(sc.UserMode, s)))
		st, answered := systemd.State(state)
		running[s] = answered && systemd.Active(st)
		anyRunning = anyRunning || running[s]
	}
	if changed && anyRunning {
		ok, aerr := approve()
		if aerr != nil {
			return "", false, aerr
		}
		if !ok {
			return "", true, nil
		}
	}
	if changed {
		o, werr := write()
		b.WriteString(o)
		if werr != nil {
			return b.String(), false, werr
		}
		o, rerr := gdurunner.Combined(ctx, r, sc.WithEnv(systemd.DaemonReload(sc.UserMode)))
		b.WriteString(o)
		if rerr != nil {
			return b.String(), false, fmt.Errorf("systemctl daemon-reload: %w", rerr)
		}
	}
	for _, s := range services {
		verb, build := "start", systemd.Start
		switch {
		case running[s] && !changed:
			b.WriteString(s + " is unchanged and left running; `solmq-conn-util restart` restarts it\n")
			continue
		case running[s]:
			verb, build = "restart", systemd.Restart
			b.WriteString(s + " was running; restarting it to load the new configuration\n")
		}
		o, serr := gdurunner.Combined(ctx, r, sc.WithEnv(build(sc.UserMode, s)))
		b.WriteString(o)
		if serr != nil {
			return b.String(), false, fmt.Errorf("systemctl %s %s: %w", verb, s, serr)
		}
	}
	return b.String(), false, nil
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
		o, err := gdurunner.Combined(ctx, r, sc.WithEnv(systemd.Stop(sc.UserMode, s)))
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
	o, err := gdurunner.Combined(ctx, r, sc.WithEnv(systemd.DaemonReload(sc.UserMode)))
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
	// --all-namespaces outranks a namespace (kube.GetOptions): --all is explicitly a
	// cluster-wide search, so a namespace resolved from env.yaml must not narrow it
	// back down.
	c := kube.Get(cmd, kube.GetOptions{Namespace: namespace, AllNamespaces: allNamespaces, Selector: selector,
		Output: "json"}, append([]string{"pods"}, names...)...)
	out, err := runParsed(ctx, r, c)
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
	out, err := runParsed(ctx, r, kube.Get(cmd, kube.GetOptions{Namespace: namespace, Output: "json"}, kind, name))
	if err != nil {
		return out, fmt.Errorf("reading %s/%s: %w\n%s", kind, name, err, out)
	}
	return out, nil
}

// KubernetesListJSON lists what is in a namespace in one call (nsgate.List: `get
// all,persistentvolumeclaims,secrets,configmaps -n <ns> -o json`), for the namespace
// gate. `all` covers the workload and service kinds; the three after it do not appear
// under `all` and are exactly the ones whose loss would hurt most.
func KubernetesListJSON(ctx context.Context, r gdurunner.Runner, cmd []string, namespace string) (string, error) {
	out, err := runParsed(ctx, r, nsgate.List(cmd, namespace))
	if err != nil {
		return out, fmt.Errorf("listing namespace %q: %w\n%s", namespace, err, out)
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

// EngineInspectJSON inspects one or more containers: `<cmd> container inspect
// <names...>` on docker or podman (inspect.Containers), which answers a JSON array
// in the order the names were given. The container subcommand answers for
// containers alone, so an image or a volume of the same name never does. One call
// covers every target, and its output carries the compose project label too, so
// nothing else has to be asked for it. Decode the answer with inspect.Decode.
//
// names are operator-supplied (or read back from the engine under --all) and
// must already be validated by the caller.
func EngineInspectJSON(ctx context.Context, r gdurunner.Runner, cmd []string, names []string) (string, error) {
	c, ok := inspect.Containers(cmd, names...)
	if !ok {
		return "", fmt.Errorf("no container named to inspect")
	}
	out, err := runParsed(ctx, r, c)
	if err != nil {
		return out, fmt.Errorf("inspecting %s: %w\n%s", strings.Join(names, ", "), err, safeOut(out))
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
		return out, fmt.Errorf("inspecting image %s: %w\n%s", ref, err, safeOut(out))
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
		return out, fmt.Errorf("sampling container resource usage: %w\n%s", err, safeOut(out))
	}
	return out, nil
}

// EngineList lists the containers the engine knows about, stopped ones included
// (`<cmd> ps --all --no-trunc --format <name, image, status>`, containercli.List),
// or only running ones when running is set. Read the answer with
// containercli.ParseList. --all discovery filters the result by image reference
// (containercli.ByImage), since without env.yaml or an explicit name the image is
// the only thing that identifies a connector instance; a stopped container is
// included on purpose, because an instance that died is exactly what a search like
// this is for.
func EngineList(ctx context.Context, r gdurunner.Runner, cmd []string, running bool) (string, error) {
	out, err := runParsed(ctx, r, containercli.List(cmd, running))
	if err != nil {
		return out, fmt.Errorf("listing containers: %w\n%s", err, safeOut(out))
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
	out, err := runParsed(ctx, r, sc.WithEnv(systemd.NRestarts(sc.UserMode, unit)))
	if err != nil {
		return 0, fmt.Errorf("reading %s restart count from systemd: %w\n%s", unit, err, out)
	}
	n, perr := systemd.ParseNRestarts(out)
	if perr != nil {
		return 0, fmt.Errorf("reading %s restart count from systemd: %w", unit, perr)
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

// ExecArgv builds the `<cmd> exec` argv shared by InstallScript,
// RunStatusScript and the cli verb's attached session, keeping the per-platform
// shape in one place instead of repeating it in each caller. command is the
// in-container program and its arguments, which the caller supplies and which must
// not be empty.
//
// The shapes differ in more than spelling, and the difference is a parser
// difference rather than a style one. kubectl leaves its flag parser
// interspersed, so the namespace and container flags are accepted after the pod
// positional, and a "--" is required to mark where the in-container command
// begins. docker and podman deliberately stop interspersing on exec, so that a
// flag typed after the container name belongs to the in-container program --
// which means their own flags must come first, and a "--" would be handed
// through to the container as an argument rather than swallowed. kubectl's shape is
// kube.Exec's and the two engines' is containercli.Exec's.
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
// The command is either tool-authored (ContainerShell, or a payload built from
// constants) or already SafeToken-validated. target and namespace are
// operator-supplied and must already be validated by the caller before reaching
// this function.
func ExecArgv(cmd []string, platform, target, namespace string, o ExecOpts, command ...string) ([]string, error) {
	if o.TTY && !o.Stdin {
		// Unreachable via the CLI, which never asks for one without the other.
		// Guarded anyway so a future caller cannot quietly get a terminal it has
		// no way to type into.
		return nil, fmt.Errorf("a tty needs stdin: ask for both or neither")
	}
	switch platform {
	case validate.PlatformKubernetes:
		// kube.Exec refuses an empty pod and an empty command.
		c, err := kube.Exec(cmd, target, kube.ExecOptions{Namespace: namespace,
			Container: spec.ConnectorContainerName, Stdin: o.Stdin, TTY: o.TTY}, command...)
		if err != nil {
			return nil, err
		}
		return c.Argv, nil
	case validate.PlatformDocker, validate.PlatformPodman:
		c, err := containercli.Exec(cmd, target, containercli.ExecOptions{Stdin: o.Stdin, TTY: o.TTY}, command...)
		if err != nil {
			return nil, err
		}
		return c.Argv, nil
	default:
		return nil, fmt.Errorf("unknown platform %q (want %q, %q, or %q)", platform, validate.PlatformKubernetes, validate.PlatformDocker, validate.PlatformPodman)
	}
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
// under the same name (containercli.NoPrevious) -- and is refused by the caller
// before it gets here rather than being silently dropped.
//
// The container is spec.ConnectorContainerName, named outright for the reason
// ExecArgv records: the tool renders that name, and a pod it cannot find it in
// should say so rather than return some other container's log.
//
// target and namespace are operator-supplied and must already be validated by
// the caller. Everything else this function appends is a tool-authored constant
// or a number.
func LogsArgv(cmd []string, platform, target, namespace string, o LogsOpts) ([]string, error) {
	switch platform {
	case validate.PlatformKubernetes:
		lo := kube.LogOptions{Namespace: namespace, Container: spec.ConnectorContainerName, Previous: o.Previous,
			Follow: o.Follow, Timestamps: o.Timestamps, Since: o.Since}
		if o.Tail != TailAll {
			lo.Tail = strconv.Itoa(o.Tail)
		}
		return kube.Logs(cmd, target, lo).Argv, nil
	case validate.PlatformDocker, validate.PlatformPodman:
		if o.Previous {
			// Unreachable via the CLI, which refuses the combination with the same
			// message. Guarded anyway so a future caller cannot quietly get a log
			// that ignores what it asked for.
			return nil, containercli.NoPrevious(containercli.Engine(platform))
		}
		lo := containercli.LogOptions{Follow: o.Follow, Timestamps: o.Timestamps, Since: o.Since}
		if o.Tail != TailAll {
			lo.Tail = strconv.Itoa(o.Tail)
		}
		return containercli.Logs(cmd, target, lo).Argv, nil
	default:
		return nil, fmt.Errorf("unknown platform %q (want %q, %q, or %q)", platform, validate.PlatformKubernetes, validate.PlatformDocker, validate.PlatformPodman)
	}
}

// ScriptAbsentMarker is the one line RunStatusScript's payload echoes when the
// script is not there. It is distinctive enough that engine chatter on the same
// combined-output stream cannot be mistaken for it, and a whole line rather than
// a substring, so a report line that happens to quote it does not count.
const ScriptAbsentMarker = "solmq-script-absent"

// statusScriptPayload is the sh -c program RunStatusScript sends. The script's
// path arrives as $0 and its arguments as $@, so neither is ever spliced into
// the shell text. exec hands the exit status and both streams straight to the
// script, which keeps the script's own always-exit-0 contract intact.
const statusScriptPayload = `if [ -f "$0" ]; then exec ` + ContainerShell + ` "$0" "$@"; else echo ` + ScriptAbsentMarker + `; fi`

// InstallScript pipes script onto the target's stdin and writes it to path,
// creating path's directory first: `sh -c "mkdir -p '<dir>' && cat > '<path>'"` with
// Cmd.Stdin = script. The script crosses only on stdin -- never cp, never a
// download, never the script text in argv. The command is kube.Upload on kubernetes
// and containercli.Upload on docker and podman, each of which single-quotes the path.
//
// path and script are tool-authored constants, never operator input. target and
// namespace are operator-supplied and must already be validated by the caller.
func InstallScript(ctx context.Context, r gdurunner.Runner, cmd []string, platform, target, namespace, path, script string) (string, error) {
	var c gdurunner.Cmd
	switch platform {
	case validate.PlatformDocker, validate.PlatformPodman:
		var err error
		if c, err = containercli.Upload(cmd, target, path, []byte(script), true); err != nil {
			return "", err
		}
	case validate.PlatformKubernetes:
		var err error
		if c, err = kube.Upload(cmd, target, kube.ExecOptions{Namespace: namespace, Container: spec.ConnectorContainerName},
			path, []byte(script), true); err != nil {
			return "", err
		}
	default:
		return "", fmt.Errorf("unknown platform %q (want %q, %q, or %q)", platform, validate.PlatformKubernetes, validate.PlatformDocker, validate.PlatformPodman)
	}
	return gdurunner.Combined(ctx, r, c)
}

// RunStatusScript runs the status script at path inside target when it is
// there, in one exec that also answers whether it is: present is false only when
// the payload echoed ScriptAbsentMarker on a line of its own, which it does
// instead of running anything. That folds the install probe into the run, so a
// deployed instance costs one exec rather than two.
//
// The script always exits 0 and reports through its output -- the leader-election
// state on stdout, anything that went wrong on stderr -- so a non-zero exit
// here did not come from the script and means the exec itself failed (target
// gone, engine unreachable). Callers should treat it as a failed target rather
// than as a state. Output is returned alongside any error so whatever the
// engine printed is available for the message.
//
// path and args are tool-authored constants, never operator input, and travel
// as positional parameters rather than inside the sh -c text. target and
// namespace are operator-supplied and must already be validated by the caller.
func RunStatusScript(ctx context.Context, r gdurunner.Runner, cmd []string, platform, target, namespace, path string, args ...string) (string, bool, error) {
	command := append([]string{ContainerShell, "-c", statusScriptPayload, path}, args...)
	argv, err := ExecArgv(cmd, platform, target, namespace, ExecOpts{}, command...)
	if err != nil {
		return "", false, err
	}
	out, runErr := gdurunner.Combined(ctx, r, gdurunner.Cmd{Argv: argv})
	for _, line := range strings.Split(out, "\n") {
		if strings.TrimSpace(line) == ScriptAbsentMarker {
			return "", false, nil
		}
	}
	return out, true, runErr
}
