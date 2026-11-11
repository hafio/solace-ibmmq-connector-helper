package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/hafio/go-deploy-utility/execguard"
	"github.com/hafio/go-deploy-utility/exitcode"
	"github.com/hafio/go-deploy-utility/kube"
	gdurunner "github.com/hafio/go-deploy-utility/runner"
	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/gen"
	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/runner"
	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/spec"
	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/validate"
)

// lifecycleNameRE is the DNS-1123 label every name a lifecycle verb puts in an
// argv must be: validate holds deployment.name, deployment.namespace and the
// docker/podman names to it, and these verbs read env.yaml without validating
// the rest of it, so the names are held to it here -- above all, none may
// start with '-' and be read as a flag.
var lifecycleNameRE = regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`)

// runLifecycle implements start, stop and restart: the instance env.yaml names
// on the resolved platform, acted on in place. It reads env.yaml for the
// section's names and command only -- nothing is rendered, no credential is
// resolved and no store opened, so it runs with no -env variable exported --
// and, like deploy, runs the platform's read-only preflight probe first.
//
//   - kubernetes: stop scales the Deployment to 0, start back to
//     deployment.replicas, restart is `rollout restart`.
//   - docker: `<command> stop|start|restart <docker.name>`; a created store is
//     decoded again on every start.
//   - podman: `systemctl [--user] stop|start|restart <name>.service`.
func runLifecycle(action string, args []string, a app) int {
	r := a.r
	fs := verbFlagSet(action)
	env := envFlag(fs)
	platform := platformFlag(fs)
	allow := allowCommandFlag(fs)
	pos, err := collectFlagsAndDirs(fs, args)
	if err != nil {
		return flagExit(action, err)
	}
	approvals, aerr := approvalsFrom(*allow)
	if aerr != nil {
		return usageExit(aerr)
	}
	if len(pos) > 0 {
		fmt.Fprintf(os.Stderr, "%s: unexpected argument %q (the platform is selected with --platform)\n", action, pos[0])
		fmt.Fprint(os.Stderr, verbUsage(action))
		return exitcode.Usage
	}
	e, lerr := loadEnvFile(*env)
	if lerr != nil {
		return usageExit(lerr)
	}
	name, perr := resolvePlatform(*platform, presentPlatforms(e), true, a.prompter)
	if perr != nil {
		return usageExit(perr)
	}
	checkName := func(field, v string) error {
		if !lifecycleNameRE.MatchString(v) || len(v) > 63 {
			return fmt.Errorf("%s %q is not a valid DNS-1123 label; run validate", field, v)
		}
		return nil
	}
	var out string
	switch name {
	case validate.PlatformKubernetes:
		k := e.Kubernetes
		if err := checkName("kubernetes.deployment.name", k.Deployment.Name); err != nil {
			return usageExit(err)
		}
		if ns := k.Deployment.Namespace; ns != "" {
			if err := checkName("kubernetes.deployment.namespace", ns); err != nil {
				return usageExit(err)
			}
		}
		kprobes, kerr := runner.KubernetesProbes(action, k.Deployment.Namespace, nil, false)
		if kerr != nil {
			return errExit(kerr)
		}
		if code, ok := preflight(r, action, name, k.Command, false, kprobes, approvals); !ok {
			return code
		}
		out, err = runner.KubernetesLifecycle(bg(), r, k.Command, action, k.Deployment.Namespace, k.Deployment.Name, k.Deployment.Replicas, approvals)
		// start and restart wait for the pods to roll; stop has nothing to wait for.
		if err == nil && action != runner.ActionStop {
			out, err = waitForRollout(r, k, out, approvals)
		}
	case validate.PlatformDocker:
		d := e.Docker
		if err := checkName("docker.name", d.Name); err != nil {
			return usageExit(err)
		}
		if code, ok := preflight(r, action, name, d.Command, false, nil, approvals); !ok {
			return code
		}
		out, err = runner.DockerLifecycle(bg(), r, d.Command, action, d.Name, approvals)
	default:
		p := e.Podman
		if err := checkName("podman.name", p.Name); err != nil {
			return usageExit(err)
		}
		sc, serr := runner.ResolveQuadletScope(a.host.scope, p.Rootful)
		if serr != nil {
			return errExit(serr)
		}
		if code, ok := preflight(r, action, name, p.Command, sc.UserMode, nil, approvals); !ok {
			return code
		}
		out, err = runner.PodmanLifecycle(bg(), r, sc, action, gen.PodmanServiceName(p.Name))
	}
	return report(action, name, out, err)
}

// waitForRollout waits until the Deployment has rolled out (kube.RolloutStatus, up
// to kube.RolloutTimeout), after a deploy's apply and after start and restart. out
// is what the command printed so far; the wait's output is appended to it, so the
// one ok:/error: line at the end still reports everything.
//
// The wait is a convenience, and the change is already applied: when kubectl refuses
// it (kube.Forbidden) the result is a warning and a success, since a rollout this
// identity may not watch is no failed deploy. A timeout or a rollout that fails is an
// error, exit 1.
func waitForRollout(r gdurunner.Runner, k *spec.Kubernetes, out string, approvals execguard.Approvals) (string, error) {
	name, ns := k.Deployment.Name, k.Deployment.Namespace
	wout, werr := runner.KubernetesRolloutStatus(bg(), r, k.Command, ns, name, approvals)
	out = strings.TrimSpace(out + "\n" + wout)
	switch {
	case werr == nil:
		return out, nil
	case kube.Forbidden(wout):
		return out + fmt.Sprintf("\nwarning: the wait for deployment/%s to roll out was refused, so it was not watched; "+
			"the change was applied", name), nil
	}
	return out, fmt.Errorf("deployment/%s did not finish rolling out: %w", name, werr)
}
