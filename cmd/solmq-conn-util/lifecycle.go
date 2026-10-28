package main

import (
	"fmt"
	"os"
	"regexp"

	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/gen"
	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/runner"
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
func runLifecycle(action string, args []string, r runner.Runner) int {
	fs := verbFlagSet(action)
	env := envFlag(fs)
	platform := platformFlag(fs)
	allow := allowCommandFlag(fs)
	pos, err := collectFlagsAndDirs(fs, args)
	if err != nil {
		return flagExit(action, err)
	}
	if len(pos) > 0 {
		fmt.Fprintf(os.Stderr, "%s: unexpected argument %q (the platform is selected with --platform)\n", action, pos[0])
		fmt.Fprint(os.Stderr, verbUsage(action))
		return 2
	}
	e, lerr := loadEnvFile(*env)
	if lerr != nil {
		return errExit(lerr)
	}
	name, perr := resolvePlatform(*platform, presentPlatforms(e), true)
	if perr != nil {
		return errExit(perr)
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
			return errExit(err)
		}
		if ns := k.Deployment.Namespace; ns != "" {
			if err := checkName("kubernetes.deployment.namespace", ns); err != nil {
				return errExit(err)
			}
		}
		if code, ok := preflight(r, action, name, k.Command, k.Deployment.Namespace, *allow); !ok {
			return code
		}
		out, err = runner.KubernetesLifecycle(r, k.Command, action, k.Deployment.Namespace, k.Deployment.Name, k.Deployment.Replicas, *allow)
	case validate.PlatformDocker:
		d := e.Docker
		if err := checkName("docker.name", d.Name); err != nil {
			return errExit(err)
		}
		if code, ok := preflight(r, action, name, d.Command, "", *allow); !ok {
			return code
		}
		out, err = runner.DockerLifecycle(r, d.Command, action, d.Name, *allow)
	default:
		p := e.Podman
		if err := checkName("podman.name", p.Name); err != nil {
			return errExit(err)
		}
		sc, serr := runner.ResolveQuadletScope()
		if serr != nil {
			return errExit(serr)
		}
		if code, ok := preflight(r, action, name, p.Command, "", *allow); !ok {
			return code
		}
		out, err = runner.PodmanLifecycle(r, sc, action, gen.PodmanServiceName(p.Name))
	}
	return report(action, name, out, err)
}
