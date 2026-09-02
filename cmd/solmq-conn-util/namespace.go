package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/deploy"
	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/runner"
	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/spec"
	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/validate"
)

// This file is the namespace half of a kubernetes teardown, which is deliberately
// not part of the manifest delete.
//
// Deleting a Namespace cascades to everything inside it, so a manifest carrying
// one -- piped to `kubectl delete -f -` -- takes workloads this tool never
// deployed along with its own. remove therefore renders without the Namespace
// document (deploy.Input.OmitNamespace) and the namespace is settled here
// instead: refused outright for the namespaces no tool should ever delete, and
// offered only once nothing else is living in it.
//
// The invariant this file exists to hold: a namespace containing anything this
// release does not own is never removed. Not interactively, not under
// --no-prompt, not on any path. The occupancy check runs first and an occupant
// of any kind ends it; --no-prompt only removes the question in the case where
// the namespace is already empty.

// namespaceProbeTypes is what the occupancy check asks kubectl for. `all` covers
// the workload and service kinds; the three after it do not appear under `all`
// and are exactly the ones whose loss would hurt most.
//
// A tool-authored constant, never operator input.
const namespaceProbeTypes = "all,persistentvolumeclaims,secrets,configmaps"

// wellKnownNamespaces are refused outright, ahead of any occupancy result.
// Deleting one of these is unrecoverable and breaks the cluster rather than one
// application, so no emptiness check should be able to authorise it -- an
// operator who has typed the wrong namespace into env.yaml is the likeliest way
// to arrive here.
var wellKnownNamespaces = map[string]bool{
	"default":         true,
	"kube-system":     true,
	"kube-public":     true,
	"kube-node-lease": true,
}

// clusterDefaultOccupants exist in every namespace kubernetes creates, so they
// are not evidence that anyone is using it.
var clusterDefaultOccupants = map[string]string{
	"ConfigMap":      "kube-root-ca.crt",
	"ServiceAccount": "default",
}

// nsItem is the slice of a kubectl object this check needs.
//
// Parsed here rather than through internal/statusreport because that package
// models pods and containers for the status report; this is a different
// question about arbitrary object kinds, and the four fields below are the
// whole of it.
type nsItem struct {
	Kind     string `json:"kind"`
	Type     string `json:"type"` // secrets only: kubernetes.io/service-account-token
	Metadata struct {
		Name              string `json:"name"`
		DeletionTimestamp string `json:"deletionTimestamp"`
		OwnerReferences   []struct {
			Kind string `json:"kind"`
			Name string `json:"name"`
		} `json:"ownerReferences"`
	} `json:"metadata"`
}

// ownedNames is every object name this release put in the namespace, so a
// straggler of our own is not mistaken for someone else's workload.
//
// The check runs after the teardown, so these are mostly already gone; this is
// the safety net for one that has not finished terminating and carries no
// deletionTimestamp yet.
func ownedNames(k *spec.Kubernetes) map[string]bool {
	owned := map[string]bool{}
	if k == nil {
		return owned
	}
	name := k.Deployment.Name
	for _, n := range []string{name, name + "-config"} {
		if n != "" {
			owned[n] = true
		}
	}
	if c := k.Secrets.Credentials; c != nil && c.Create != nil {
		owned[c.Create.Name] = true
	}
	if s := k.Secrets.Stores; s != nil && s.Create != nil {
		owned[s.Create.Name] = true
	}
	if ip := k.Secrets.ImagePull; ip != nil && ip.Create {
		owned[ip.Name] = true
	}
	if lb := k.Libs; lb != nil && lb.PVC != nil && lb.PVC.Create != nil {
		owned[lb.PVC.Create.Name] = true
	}
	delete(owned, "")
	return owned
}

// namespaceOccupants lists what is living in the namespace that this release
// does not own, as sorted "kind/name" entries.
//
// Three things are deliberately not occupants:
//
//   - anything with a deletionTimestamp -- that is the teardown that just ran,
//     still terminating, and the single most likely false positive;
//   - our own objects, and anything owned by one of them (the ReplicaSets and
//     Pods behind the Deployment name their owner, so one rule covers them);
//   - the defaults kubernetes puts in every namespace it creates.
func namespaceOccupants(doc string, owned map[string]bool) ([]string, error) {
	var list struct {
		Items []nsItem `json:"items"`
	}
	if err := json.Unmarshal([]byte(strings.TrimSpace(doc)), &list); err != nil {
		return nil, fmt.Errorf("reading the namespace contents: %w", err)
	}
	var out []string
	for _, it := range list.Items {
		if isOurs(it, owned) || isClusterDefault(it) {
			continue
		}
		out = append(out, strings.ToLower(it.Kind)+"/"+it.Metadata.Name)
	}
	sort.Strings(out)
	return out, nil
}

// isOurs reports whether an item belongs to the release just torn down, either
// by name, by owner, or by already being on its way out.
func isOurs(it nsItem, owned map[string]bool) bool {
	if it.Metadata.DeletionTimestamp != "" {
		return true
	}
	if owned[it.Metadata.Name] {
		return true
	}
	for _, ref := range it.Metadata.OwnerReferences {
		if owned[ref.Name] {
			return true
		}
	}
	return false
}

// isClusterDefault reports whether an item is one kubernetes creates in every
// namespace, which says nothing about whether anyone is using it.
func isClusterDefault(it nsItem) bool {
	if name, ok := clusterDefaultOccupants[it.Kind]; ok && it.Metadata.Name == name {
		return true
	}
	// Older clusters mint a service-account token Secret per namespace.
	return it.Kind == "Secret" && it.Type == "kubernetes.io/service-account-token"
}

// removeNamespace is the second half of a kubernetes teardown, run after the
// manifest delete succeeds.
//
// It never removes a namespace it cannot prove is empty: a failed probe, a
// well-known name, or any occupant at all leaves it in place. The exit code
// stays 0 throughout -- the teardown the operator asked for did succeed, and
// declining to delete a namespace is an outcome, not a failure.
func removeNamespace(r runner.Runner, o actionOpts, k *spec.Kubernetes) int {
	command, ns := k.Command, k.Deployment.Namespace
	if ns == "" {
		fmt.Fprintf(os.Stderr, "remove: no namespace in env.yaml to remove\n")
		return 0
	}
	if wellKnownNamespaces[ns] {
		fmt.Fprintf(os.Stderr, "remove: %s is a cluster namespace and is never removed\n", ns)
		return 0
	}

	argv, err := runner.ParseCommand(validate.PlatformKubernetes, command, o.allow)
	if err != nil {
		return errExit(err)
	}
	doc, lerr := runner.KubernetesListJSON(r, argv, ns, namespaceProbeTypes)
	if lerr != nil {
		// A namespace is not worth deleting on a guess.
		fmt.Fprintf(os.Stderr, "remove: could not check what is in namespace %s, so it was left in place: %v\n", ns, lerr)
		return 0
	}
	occupants, perr := namespaceOccupants(doc, ownedNames(k))
	if perr != nil {
		fmt.Fprintf(os.Stderr, "remove: could not check what is in namespace %s, so it was left in place: %v\n", ns, perr)
		return 0
	}
	if len(occupants) > 0 {
		fmt.Fprintf(os.Stderr, "\nnamespace %s still holds %d resource(s) this release does not own:\n", ns, len(occupants))
		for _, occ := range occupants {
			fmt.Fprintln(os.Stderr, "  "+occ)
		}
		fmt.Fprintf(os.Stderr, "leaving namespace %s in place.\n", ns)
		return 0
	}

	if !o.noPrompt {
		ok, cerr := confirmNamespaceRemoval(ns)
		if cerr != nil {
			return errExit(cerr)
		}
		if !ok {
			fmt.Fprintf(os.Stderr, "remove: namespace %s left in place\n", ns)
			return 0
		}
	}
	out, rerr := runner.Kubernetes(r, command, runner.ActionRemove, deploy.NamespaceManifest(ns), o.allow)
	return report("remove namespace", tgtKubernetes, out, rerr)
}

// confirmNamespaceRemoval asks the second question, through the same promptLine
// seam and non-TTY refusal every other prompt uses.
//
// It is asked separately from the teardown confirmation because it destroys
// something the teardown does not, and something the operator may well not have
// created: saying yes to removing a deployment is not saying yes to removing
// the namespace around it.
func confirmNamespaceRemoval(ns string) (bool, error) {
	line, err := promptLine(fmt.Sprintf("namespace %s holds nothing else -- remove the namespace too? [y/N] ", ns))
	if err != nil {
		return false, fmt.Errorf("confirming the namespace removal interactively: %w; pass %s instead", err, noPromptFlagName)
	}
	switch strings.ToLower(line) {
	case "y", "yes":
		return true, nil
	default:
		return false, nil
	}
}
