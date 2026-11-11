package main

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/hafio/go-deploy-utility/confirm"
	"github.com/hafio/go-deploy-utility/execguard"
	"github.com/hafio/go-deploy-utility/exitcode"
	"github.com/hafio/go-deploy-utility/kube"
	"github.com/hafio/go-deploy-utility/kube/nsgate"
	"github.com/hafio/go-deploy-utility/kube/rbac"
	gdurunner "github.com/hafio/go-deploy-utility/runner"
	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/runner"
	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/spec"
)

// This file is the namespace on each side of a kubernetes deploy: created before
// the apply, only when it is absent, and offered for deletion after a remove.
//
// Creating comes first (ensureNamespace). The deploy manifest carries no Namespace
// document (deploy.Input.OmitNamespace), because applying one asks the cluster for
// permission to create namespaces on every run, even when the namespace exists. The
// namespace is read first, and only an absent one is created, by its own apply and
// only when this identity may.
//
// Deleting is deliberately not part of the manifest delete. Deleting a Namespace
// cascades to everything inside it, so a manifest carrying one -- piped to
// `kubectl delete -f -` -- takes workloads this tool never deployed along with its
// own. remove therefore renders without the Namespace document (deploy.Input.Teardown)
// and the namespace is settled here instead, in nsgate's order: refused outright for
// the namespaces no tool should ever delete, and offered only once nothing else is
// living in it.
//
// The invariant this half exists to hold: a namespace containing anything this
// release does not own is never removed. Not interactively, not under --no-prompt,
// not on any path. The occupancy check runs first and an occupant of any kind ends
// it; --no-prompt only removes the question in the case where the namespace is
// already empty. And nothing in the gate fails the remove: every way out of it short
// of a confirmed delete keeps the namespace with a warning and exits 0.

// ensureNamespace makes sure ns exists before the apply, asking for no cluster-wide
// permission when it already does (docs/developer.md, the shared module's rules):
//
//   - exists: nothing more.
//   - absent: probe `create namespaces`, then apply the Namespace on its own. A
//     refused probe is kube.MissingNamespace, a failure; an unanswered one (no
//     login, no cluster) is its own error.
//   - unknown (the read was refused or failed): the same probe. Allowed, apply it.
//     Refused, warn and carry on, since the apply's own write stops a deploy into a
//     namespace that is not there. Unanswered is an error.
//
// ok is false when the deploy must stop, with the exit code to return.
func ensureNamespace(r gdurunner.Runner, base []string, command, ns string, approvals execguard.Approvals) (code int, ok bool) {
	state := runner.KubernetesNamespaceState(bg(), r, base, ns)
	if state == kube.NamespaceExists {
		return exitcode.OK, true
	}
	perr := runner.CheckProbe(bg(), r, base, rbac.Probe{Verb: "create", Resource: "namespaces"})
	var refused *rbac.RefusedError
	switch {
	case perr == nil:
		manifest, merr := kube.NamespaceManifest(ns)
		if merr != nil {
			return report(runner.ActionDeploy, tgtKubernetes, "", merr), false
		}
		if out, aerr := runner.Kubernetes(bg(), r, command, runner.ActionDeploy, string(manifest), approvals); aerr != nil {
			return report("create namespace "+ns+" for", tgtKubernetes, out, aerr), false
		}
	case !errors.As(perr, &refused):
		return report(runner.ActionDeploy, tgtKubernetes, "", perr), false
	case state == kube.NamespaceAbsent:
		return report(runner.ActionDeploy, tgtKubernetes, "", kube.MissingNamespace(base, ns, perr)), false
	default:
		fmt.Fprintf(os.Stderr, "warning: namespace %s could not be read (that needs get namespaces) and this identity may "+
			"not create namespaces (%v); carrying on, and the deploy stops at its first write if the namespace is missing\n", ns, perr)
	}
	return exitcode.OK, true
}

// removeNamespace is the second half of a kubernetes teardown, run after the
// manifest delete succeeds, in nsgate's order. base is the guarded kubectl argv.
//
// It never removes a namespace it cannot prove is empty: a failed listing, a
// protected name, or any occupant at all leaves it in place, and so does a refused
// delete permission (asked before the question, so nobody is asked and then
// refused), a declined question, a missing terminal and a failed delete. The exit
// code stays 0 throughout -- the teardown the operator asked for did succeed, and
// declining to delete a namespace is an outcome, not a failure.
func removeNamespace(r gdurunner.Runner, o actionOpts, k *spec.Kubernetes, base []string) int {
	ns := k.Deployment.Namespace
	if ns == "" {
		fmt.Fprintf(os.Stderr, "remove: no namespace in env.yaml to remove\n")
		return exitcode.OK
	}
	if nsgate.Protected(ns) {
		fmt.Fprintf(os.Stderr, "remove: %s is a cluster namespace and is never removed\n", ns)
		return exitcode.OK
	}

	doc, lerr := runner.KubernetesListJSON(bg(), r, base, ns)
	if lerr != nil {
		// A namespace is not worth deleting on a guess.
		fmt.Fprintf(os.Stderr, "remove: could not check what is in namespace %s, so it was left in place: %v\n", ns, lerr)
		return exitcode.OK
	}
	occupants, perr := nsgate.Occupants(doc, k.CreatedNames())
	if perr != nil {
		fmt.Fprintf(os.Stderr, "remove: could not check what is in namespace %s, so it was left in place: %v\n", ns, perr)
		return exitcode.OK
	}
	if len(occupants) > 0 {
		fmt.Fprintf(os.Stderr, "\nnamespace %s still holds %d resource(s) this release does not own:\n", ns, len(occupants))
		for _, occ := range occupants {
			fmt.Fprintln(os.Stderr, "  "+occ)
		}
		fmt.Fprintf(os.Stderr, "leaving namespace %s in place.\n", ns)
		return exitcode.OK
	}

	// The permission is asked before the question, so nobody is asked and then
	// refused.
	if derr := runner.CheckProbe(bg(), r, base, nsgate.DeleteProbe()); derr != nil {
		fmt.Fprintf(os.Stderr, "warning: namespace %s was left in place: %v\n", ns, derr)
		return exitcode.OK
	}
	if !o.noPrompt {
		ok, cerr := confirmNamespaceRemoval(o.prompter, ns)
		if cerr != nil {
			// The prompter's error already names --no-prompt as the way to answer
			// without a terminal.
			fmt.Fprintf(os.Stderr, "warning: namespace %s was left in place, since it could not be asked about: %v\n", ns, cerr)
			return exitcode.OK
		}
		if !ok {
			fmt.Fprintf(os.Stderr, "remove: namespace %s left in place\n", ns)
			return exitcode.OK
		}
	}
	out, derr := runner.KubernetesDeleteNamespace(bg(), r, base, ns)
	if derr != nil {
		if s := strings.TrimSpace(out); s != "" {
			fmt.Fprintln(os.Stderr, s)
		}
		fmt.Fprintf(os.Stderr, "warning: namespace %s was left in place, since deleting it failed: %v\n", ns, derr)
		return exitcode.OK
	}
	return report("remove namespace", tgtKubernetes, out, nil)
}

// confirmNamespaceRemoval asks the second question, through the same prompter
// and non-TTY refusal every other prompt uses.
//
// It is asked separately from the teardown confirmation because it destroys
// something the teardown does not, and something the operator may well not have
// created: saying yes to removing a deployment is not saying yes to removing
// the namespace around it.
func confirmNamespaceRemoval(p *confirm.Prompter, ns string) (bool, error) {
	return p.Gate(false, confirm.Lenient,
		fmt.Sprintf("namespace %s holds nothing else -- remove the namespace too? [y/N] ", ns), "the namespace removal")
}
