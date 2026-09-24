package main

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"time"

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

// The verbs that reach into running instances, as instanceRequest.verb spells
// them. They gate the wording of messages that name a flag, since the three
// verbs do not offer the same set.
const (
	verbStatus = "status"
	verbLogs   = "logs"
	verbShell  = "cli"
)

// instanceRequest is what the operator asked for on the command line, before
// any of it has been resolved or validated.
type instanceRequest struct {
	// verb is the command being run ("status", "logs"). It decides which flags
	// a "nothing to discover from" message is allowed to name -- logs has no
	// --all, and pointing an operator at a flag their verb does not have is the
	// same mistake podOrContainerFlag exists to avoid.
	verb     string
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
	verb     string
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

	return instanceSession{verb: t.verb, env: e, platform: platform, command: command, cmdArgv: cmdArgv, ns: ns}, nil
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
			return "", "", "", fmt.Errorf("no kubernetes: section in env.yaml to discover pods from; %s", namingHint(s))
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
			return "", fmt.Errorf("no docker: section in env.yaml to discover the container name from; %s", namingHint(s))
		}
		return s.env.Docker.Name, nil
	}
	if s.env == nil || s.env.Podman == nil {
		return "", fmt.Errorf("no podman: section in env.yaml to discover the container name from; %s", namingHint(s))
	}
	return s.env.Podman.Name, nil
}

// ---- naming, sorting, and indexes ----------------------------------------------

// podOrContainerFlag names the flag that addresses one instance on this
// platform, so a message points at the flag that exists rather than at both.
func podOrContainerFlag(platform string) string {
	if platform == validate.PlatformKubernetes {
		return podFlagName
	}
	return containerFlagName
}

// instanceNoun is what to call the things being counted, for a message that has
// to say how many matched.
func instanceNoun(platform string, n int) string {
	word := "container"
	if platform == validate.PlatformKubernetes {
		word = "pod"
	}
	if n != 1 {
		word += "s"
	}
	return word
}

// namingHint lists the ways forward a verb actually offers when discovery has
// nothing to work from.
//
// It is verb-aware because neither logs nor cli has --all: naming it there
// would send an operator to a flag their verb does not have, the same mistake
// podOrContainerFlag exists to avoid for --pod/--container.
func namingHint(s instanceSession) string {
	flag := podOrContainerFlag(s.platform)
	if s.verb == verbLogs || s.verb == verbShell {
		return "pass " + flag + " explicitly"
	}
	return "pass " + flag + " explicitly, or " + allFlagName + " to search by image"
}

// instanceRef is one instance a verb could address.
//
// There is no container field: every kubernetes exec and log read names
// spec.ConnectorContainerName outright (see runner.ExecArgv), so which
// container to reach is a constant rather than something discovery has to carry
// back.
type instanceRef struct {
	name      string
	namespace string
}

// sortRefs orders candidates by name, tie-breaking on namespace for a
// cross-namespace --all search.
//
// The order is load-bearing rather than cosmetic: it is what an index selects
// into, so it has to be the order the operator saw. kubectl already returns
// pods name-sorted, but docker ps and podman ps return creation order, so
// without this an index would disagree with the table status prints.
func sortRefs(refs []instanceRef) {
	sort.Slice(refs, func(i, j int) bool {
		if refs[i].name != refs[j].name {
			return refs[i].name < refs[j].name
		}
		return refs[i].namespace < refs[j].namespace
	})
}

// sortInstances puts a report's instances in the same order sortRefs produces,
// so an index and the table it indexes into cannot drift apart.
func sortInstances(insts []statusreport.Instance) {
	sort.Slice(insts, func(i, j int) bool {
		if insts[i].Name != insts[j].Name {
			return insts[i].Name < insts[j].Name
		}
		return insts[i].Namespace < insts[j].Namespace
	})
}

// instanceCandidates enumerates every instance the session can see with no name
// filter -- the selector from env.yaml, or (status only) --all's image search --
// sorted by name.
//
// It is what both the logs picker and index resolution index into, so the two
// cannot disagree about which instance is number 0.
func instanceCandidates(r runner.Runner, s instanceSession, all bool) ([]instanceRef, error) {
	if s.platform == validate.PlatformKubernetes {
		selector, match, _, err := kubeDiscovery(s, nil, all)
		if err != nil {
			return nil, err
		}
		doc, derr := runner.KubernetesPodsJSON(r, s.cmdArgv, s.ns, selector, nil, all)
		if derr != nil {
			return nil, derr
		}
		// ParsePods is asked only for names here, so the clock it measures ages
		// against is never read; a zero time keeps the call honest about that.
		insts, perr := statusreport.ParsePods(doc, time.Time{}, match)
		if perr != nil {
			return nil, fmt.Errorf("reading the pod list: %w", perr)
		}
		if len(insts) == 0 {
			return nil, noPodsFound(selector, s.ns, all)
		}
		refs := make([]instanceRef, 0, len(insts))
		for _, inst := range insts {
			ns := inst.Namespace
			if ns == "" {
				ns = s.ns
			}
			refs = append(refs, instanceRef{name: inst.Name, namespace: ns})
		}
		sortRefs(refs)
		return refs, nil
	}

	// Notes from --all's unsafe-name skip are dropped here on purpose: this
	// enumeration answers "which instances exist", and the collect run that
	// follows reports them where they belong, against the report itself.
	names, _, err := engineNames(r, s, nil, all)
	if err != nil {
		return nil, err
	}
	refs := make([]instanceRef, 0, len(names))
	for _, n := range names {
		refs = append(refs, instanceRef{name: n})
	}
	sortRefs(refs)
	return refs, nil
}

// isIndex reports whether s is a bare non-negative integer, the only spelling an
// index takes.
func isIndex(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

// anyIndex reports whether any entry might be an index, which is what decides
// whether the candidate list has to be fetched at all.
func anyIndex(names []string) bool {
	for _, n := range names {
		if isIndex(n) {
			return true
		}
	}
	return false
}

// resolveIndexes replaces an all-digit entry with the instance it selects from
// the sorted candidate list, leaving every other entry alone.
//
// An exact instance name always wins: a value of "0" is an index only when no
// candidate is actually named "0". That is why an all-digit value forces the
// enumeration -- the name check needs the list to check against. When no entry
// looks like an index, nothing is fetched and the names pass straight through.
func resolveIndexes(r runner.Runner, s instanceSession, names []string, all bool) ([]string, error) {
	if !anyIndex(names) {
		return names, nil
	}
	flag := podOrContainerFlag(s.platform)
	refs, err := instanceCandidates(r, s, all)
	if err != nil {
		return nil, fmt.Errorf("%s was given an index, which needs the list of instances to index into: %w", flag, err)
	}
	out := make([]string, 0, len(names))
	for _, n := range names {
		resolved, rerr := resolveOneIndex(n, refs, flag, s.platform)
		if rerr != nil {
			return nil, rerr
		}
		out = append(out, resolved)
	}
	return out, nil
}

// resolveOneIndex applies the name-wins-over-index rule to a single entry.
func resolveOneIndex(n string, refs []instanceRef, flag, platform string) (string, error) {
	for _, ref := range refs {
		if ref.name == n {
			return n, nil
		}
	}
	if !isIndex(n) {
		return n, nil
	}
	i, err := strconv.Atoi(n)
	if err != nil || i >= len(refs) {
		return "", fmt.Errorf("%s %s is out of range: %d %s match, so the index must be 0-%d",
			flag, n, len(refs), instanceNoun(platform, len(refs)), len(refs)-1)
	}
	return refs[i].name, nil
}

// ---- settling on exactly one instance ------------------------------------------

// namedInstance is the single --pod/--container value for this platform, or ""
// when the operator named none. The one-instance verbs refuse a second one
// before they get here, so at most one entry can be in either slice.
//
// Only the flag that applies to the resolved platform is read: --pod has no
// effect on docker/podman and --container none on kubernetes, exactly as status
// treats them. Reading whichever happened to be set would turn a flag documented
// as inert into one that silently picks the instance.
func namedInstance(platform string, pods, conts []string) string {
	vals := conts
	if platform == validate.PlatformKubernetes {
		vals = pods
	}
	if len(vals) > 0 {
		return vals[0]
	}
	return ""
}

// pickerLine is how the picker rebuilds a paste-able command around the
// instance flag it suggests: everything the operator typed that belongs before
// the instance, and everything that has to follow it.
//
// suffix exists for cli, whose in-container command comes after a "--" and so
// cannot precede the --pod it is being told to add. logs leaves it empty.
type pickerLine struct {
	prefix string
	suffix string
}

// resolveOneInstance settles on the one instance a single-instance verb (logs,
// cli) will act on.
//
// found is false when the picker was printed instead: discovery turned up more
// than one instance and the operator named none, so there is nothing to act on
// until they say which. An error is a run that could not get that far at all.
//
// named is the operator's --pod/--container value, "" when they gave none, and
// line is what the picker rebuilds its paste-back commands from. Each verb
// renders its own, since each carries a different set of flags.
//
// A plain name is taken at its word with no query at all. That is what having
// the container be a constant buys: the only reason this used to cost a
// `get pods <name>` call on kubernetes was to learn which container to name,
// and the namespace it also read back was the one the query was already scoped
// by.
func resolveOneInstance(r runner.Runner, sess instanceSession, named string, line pickerLine) (instanceRef, bool, error) {
	var ref instanceRef
	flag := podOrContainerFlag(sess.platform)

	// An all-digit value still forces the enumeration, because a candidate might
	// really be called that and an exact name has to win over the index reading.
	if named != "" && !isIndex(named) {
		return instanceRef{name: named, namespace: sess.ns}, true, nil
	}

	refs, cerr := instanceCandidates(r, sess, false)
	if cerr != nil {
		if named != "" {
			return ref, false, fmt.Errorf("%s was given an index, which needs the list of instances to index into: %w", flag, cerr)
		}
		return ref, false, cerr
	}

	if named != "" {
		name, rerr := resolveOneIndex(named, refs, flag, sess.platform)
		if rerr != nil {
			return ref, false, rerr
		}
		for _, c := range refs {
			if c.name == name {
				return c, true, nil
			}
		}
		return ref, false, fmt.Errorf("%s %s did not match any instance", flag, named)
	}

	if len(refs) == 1 {
		return refs[0], true, nil
	}
	printInstancePicker(os.Stdout, sess, line, refs)
	return ref, false, nil
}

// printInstancePicker lists the matching instances as commands that can be
// pasted back verbatim, which is the point of it: an operator who ran the verb
// and got three pods should not have to retype the flags they already gave.
//
// It goes to stdout because it is the answer to the run, not a diagnostic about
// it.
func printInstancePicker(w io.Writer, sess instanceSession, line pickerLine, refs []instanceRef) {
	flag := podOrContainerFlag(sess.platform)
	fmt.Fprintf(w, "%d %s match; %s takes one. Run one of:\n\n", len(refs), instanceNoun(sess.platform, len(refs)), sess.verb)
	for _, ref := range refs {
		fmt.Fprintln(w, "  "+pickerCommand(line, flag, ref.name))
	}
	fmt.Fprintf(w, "\nor by index, 0-%d in the order listed:\n\n", len(refs)-1)
	fmt.Fprintln(w, "  "+pickerCommand(line, flag, "0"))
}

// pickerCommand assembles one paste-able line: what came before the instance,
// the flag that names it, and whatever has to trail it.
func pickerCommand(line pickerLine, flag, name string) string {
	cmd := line.prefix + " " + flag + " " + name
	if line.suffix != "" {
		cmd += " " + line.suffix
	}
	return cmd
}
