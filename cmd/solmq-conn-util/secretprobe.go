package main

import (
	"fmt"
	"sort"
	"strings"

	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/gen"
	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/runner"
	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/spec"
	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/statusreport"
	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/validate"
)

// probeOperatorSecrets checks, before a kubernetes deploy applies anything,
// that every Secret the operator manages and the pod mounts -- each one a
// store names (existing.secret) and an existing credentials Secret -- is in
// the namespace and holds the keys the pod reads from it. validate cannot
// know: the Secret lives only in the cluster. Without this a missing key would
// leave the pod waiting in ContainerCreating while deploy reported success.
//
// A missing Secret or key is an error. A Secret the deploying identity may not
// read is a warning: the apply may still be allowed, and the pod will say so
// at start if something is missing. Any other failure to read it is a warning
// too, naming what kubectl said; the preflight probe has already shown the
// cluster is reachable.
func probeOperatorSecrets(r runner.Runner, k *spec.Kubernetes, probes []gen.SecretProbe, extraAllowed []string) (errs []error, warns []string) {
	argv, err := runner.ParseCommand(validate.PlatformKubernetes, k.Command, extraAllowed)
	if err != nil {
		return []error{err}, nil
	}
	ns := k.Deployment.Namespace
	for _, p := range probes {
		doc, gerr := runner.KubernetesGetJSON(r, argv, ns, "secret", p.Name)
		if gerr != nil {
			switch msg := gerr.Error(); {
			case strings.Contains(msg, "NotFound") || strings.Contains(msg, "not found"):
				errs = append(errs, fmt.Errorf("the Secret %s, which %s names, does not exist in namespace %s; create it before deploying -- the pod would wait for it in ContainerCreating", p.Name, p.Field, ns))
			case strings.Contains(msg, "Forbidden") || strings.Contains(msg, "forbidden"):
				warns = append(warns, fmt.Sprintf("cannot read the Secret %s (which %s names) to check its keys: the deploying identity may not get Secrets in %s; the pod will report a missing key at start", p.Name, p.Field, ns))
			default:
				warns = append(warns, fmt.Sprintf("could not check the Secret %s (which %s names): %s", p.Name, p.Field, firstLine(msg)))
			}
			continue
		}
		keys, ok := statusreport.SecretDataKeys(doc)
		if !ok {
			warns = append(warns, fmt.Sprintf("could not check the Secret %s (which %s names): kubectl returned something that is not a Secret", p.Name, p.Field))
			continue
		}
		var missing []string
		for _, key := range p.Keys {
			if !keys[key] {
				missing = append(missing, key)
			}
		}
		if len(missing) > 0 {
			sort.Strings(missing)
			errs = append(errs, fmt.Errorf("the Secret %s, which %s names, lacks the key(s) %s the pod reads; add them before deploying -- the pod would wait for them in ContainerCreating", p.Name, p.Field, strings.Join(missing, ", ")))
		}
	}
	return errs, warns
}

// firstLine is s up to its first newline, for a one-line message.
func firstLine(s string) string {
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		return s[:i]
	}
	return s
}
