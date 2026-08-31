package main

import (
	"fmt"
	"os"

	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/runner"
	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/spec"
	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/statusreport"
	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/validate"
)

// This file holds what every verb that reaches into running instances needs
// before it can query one: which platform, through which binary, in which
// namespace, and which pods or containers are meant. status and logs both ask
// exactly those questions, and they have to arrive at the same answers -- a
// debugging pair that disagreed about which pods it meant would be worse than
// either verb alone -- so the resolution lives here once instead of in each.

// instanceRequest is what the operator asked for on the command line, before
// any of it has been resolved or validated.
type instanceRequest struct {
	envPath  string
	platform string
	ns       string
	command  string
	pods     []string
	conts    []string
	allow    []string
	all      bool
}

// explicitTargets reports whether the operator named the instances themselves
// rather than leaving them to be discovered from env.yaml.
func (t instanceRequest) explicitTargets() bool {
	return len(t.pods) > 0 || len(t.conts) > 0 || t.all
}

// instanceSession is the resolved, validated result: everything a run needs to
// address instances, settled once so a watch tick or a second target does not
// re-resolve it.
type instanceSession struct {
	env      *spec.Env
	platform string
	// command is the binary as spelled, kept alongside cmdArgv because Preflight
	// takes the string and parses it itself.
	command string
	cmdArgv []string
	ns      string
}

// resolveInstanceSession loads env.yaml, settles the platform and the namespace,
// rejects every operator-supplied name that could not safely reach an argv, and
// parses the platform binary through the same allowlist gate deploy uses.
//
// Everything here can fail the whole run, and does so before a single process
// is started. What it deliberately does not do is run the preflight probe: that
// reports through an exit code rather than an error, so it stays with the
// caller.
func resolveInstanceSession(t instanceRequest) (instanceSession, error) {
	var s instanceSession

	e, err := loadInstanceEnv(t.envPath, t.explicitTargets() && t.platform != "")
	if err != nil {
		return s, err
	}

	platform, err := resolvePlatform(t.platform, presentPlatforms(e), !t.explicitTargets())
	if err != nil {
		return s, err
	}

	ns := instanceNamespace(platform, t.ns, e)
	if ns != "" && !validate.SafeToken(ns) {
		return s, fmt.Errorf("--namespace %q contains an unsafe character (%s)", ns, validate.UnsafeTokenReason)
	}
	for _, p := range t.pods {
		if !validate.SafeToken(p) {
			return s, fmt.Errorf("--pod %q contains an unsafe character (%s)", p, validate.UnsafeTokenReason)
		}
	}
	for _, c := range t.conts {
		if !validate.SafeToken(c) {
			return s, fmt.Errorf("--container %q contains an unsafe character (%s)", c, validate.UnsafeTokenReason)
		}
	}

	command := instanceCommand(platform, t.command, e)
	cmdArgv, err := runner.ParseCommand(platform, command, t.allow)
	if err != nil {
		return s, err
	}

	return instanceSession{env: e, platform: platform, command: command, cmdArgv: cmdArgv, ns: ns}, nil
}

// loadInstanceEnv reads env.yaml the same way loadEnvFile does, except when the
// operator has named explicit targets (--pod/--container, or --all) and an
// explicit --platform: the verb then needs nothing from the file at all (the
// exception platformResolutionDetail spells out), so a missing file is not an
// error and it proceeds against a zero Env (built-in command/port/user
// defaults). A file that does exist but fails to parse is still reported --
// its presence means the operator expects it to be read.
func loadInstanceEnv(envPath string, skipIfMissing bool) (*spec.Env, error) {
	data, err := os.ReadFile(envPath)
	if err != nil {
		if skipIfMissing {
			return &spec.Env{}, nil
		}
		return nil, fmt.Errorf("reading %s: %w", envPath, err)
	}
	return spec.ParseEnv(data)
}

// instanceCommand resolves the CLI binary to reach instances through: the
// --command override if given, else the platform's section command: in env.yaml
// when the section is present, else the platform's own default -- the same
// defaults applyKubeDefaults/applyDockerDefaults/applyPodmanDefaults fill in at
// parse time, needed again here because a run can happen with no section at all
// (explicit targets, or --all).
func instanceCommand(platform, override string, e *spec.Env) string {
	if override != "" {
		return override
	}
	switch platform {
	case validate.PlatformKubernetes:
		if e != nil && e.Kubernetes != nil && e.Kubernetes.Command != "" {
			return e.Kubernetes.Command
		}
		return spec.DefaultKubeCommand
	case validate.PlatformDocker:
		if e != nil && e.Docker != nil && e.Docker.Command != "" {
			return e.Docker.Command
		}
		return spec.DefaultDockerCommand
	case validate.PlatformPodman:
		if e != nil && e.Podman != nil && e.Podman.Command != "" {
			return e.Podman.Command
		}
		return spec.DefaultPodmanCommand
	default:
		return ""
	}
}

// instanceNamespace resolves the kubernetes namespace to query: the --namespace
// override if given, else the kubernetes section's deployment.namespace when
// present, else "" (the CLI's current-context default). No effect on
// docker/podman.
func instanceNamespace(platform, override string, e *spec.Env) string {
	if override != "" {
		return override
	}
	if platform == validate.PlatformKubernetes && e != nil && e.Kubernetes != nil {
		return e.Kubernetes.Deployment.Namespace
	}
	return ""
}

// ---- discovery ---------------------------------------------------------------

// kubeDiscovery decides how pods are found, which is three different questions
// depending on what the operator said: --all searches the whole cluster and
// filters by image, named pods are taken as given, and otherwise the deployment
// in env.yaml supplies a label selector.
//
// group is the deployment name when there is one -- the report's heading, and
// empty for the other two branches, where no single workload owns the results.
func kubeDiscovery(s instanceSession, pods []string, all bool) (selector, match, group string, err error) {
	switch {
	case all:
		// Cluster-wide, filtered by image: the operator named nothing, so the
		// image reference is the only thing that identifies an instance.
		return "", statusreport.ImageMatch, "", nil
	case len(pods) > 0:
		return "", "", "", nil
	default:
		if s.env == nil || s.env.Kubernetes == nil {
			return "", "", "", fmt.Errorf("no kubernetes: section in env.yaml to discover pods from; pass --pod explicitly, or --all to search by image")
		}
		name := s.env.Kubernetes.Deployment.Name
		return "app=" + name, "", name, nil
	}
}

// noPodsFound explains an empty pod list in the terms of whichever branch
// produced it, so the message names the fix rather than the symptom.
func noPodsFound(selector, ns string, all bool) error {
	switch {
	case all:
		return fmt.Errorf("no pod anywhere in the cluster is running an image matching %q", statusreport.ImageMatch)
	case selector != "":
		return fmt.Errorf("no pods found for selector %q in namespace %q; pass --pod explicitly", selector, ns)
	default:
		return fmt.Errorf("none of the named pods exist in namespace %q", ns)
	}
}

// engineNames decides which docker/podman containers are meant: every one
// running a matching image under --all, the named ones, or the single container
// the section in env.yaml describes.
//
// notes carries the containers --all found but had to skip. Names come back
// from the engine there rather than from the operator, so they are re-checked
// before they can reach an argv, and a skip is said out loud rather than
// silently narrowing the result.
func engineNames(r runner.Runner, s instanceSession, conts []string, all bool) (names, notes []string, err error) {
	switch {
	case all:
		out, lerr := runner.EngineList(r, s.cmdArgv)
		if lerr != nil {
			return nil, nil, lerr
		}
		for _, n := range statusreport.EngineNamesByImage(out, statusreport.ImageMatch) {
			if validate.SafeToken(n) {
				names = append(names, n)
			} else {
				notes = append(notes, fmt.Sprintf("skipping container %q: its name contains an unsafe character (%s)", n, validate.UnsafeTokenReason))
			}
		}
		if len(names) == 0 {
			return nil, notes, fmt.Errorf("no container on this host is running an image matching %q", statusreport.ImageMatch)
		}
		return names, notes, nil
	case len(conts) > 0:
		return conts, nil, nil
	default:
		name, cerr := configuredInstanceName(s)
		if cerr != nil {
			return nil, nil, cerr
		}
		return []string{name}, nil, nil
	}
}

// configuredInstanceName is the single container name the docker/podman section
// of env.yaml describes -- what a verb addresses when the operator names no
// target.
func configuredInstanceName(s instanceSession) (string, error) {
	if s.platform == validate.PlatformDocker {
		if s.env == nil || s.env.Docker == nil {
			return "", fmt.Errorf("no docker: section in env.yaml to discover the container name from; pass --container explicitly, or --all to search by image")
		}
		return s.env.Docker.Name, nil
	}
	if s.env == nil || s.env.Podman == nil {
		return "", fmt.Errorf("no podman: section in env.yaml to discover the container name from; pass --container explicitly, or --all to search by image")
	}
	return s.env.Podman.Name, nil
}
