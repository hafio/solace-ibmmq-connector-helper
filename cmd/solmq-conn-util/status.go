package main

// The status verb. It answers two different questions about the same
// instances, which is why it takes a target word rather than a pile of flags:
//
//   - container: what the engine knows -- state, restarts, age, image, and
//     under --details the node, resource use, digest, and referenced objects.
//     Collected from outside, through read-only kubectl/docker/podman queries.
//   - application: what the connector knows about itself -- leader election,
//     health, workflows, and under --details its version, JVM, config and heap.
//     Collected from inside, by running the generated script in the container.
//   - all: both, container first, since a container that is not running is why
//     the application half is missing.
//
// The engine queries are deliberately one call for many instances (`get pods`
// for a whole selector, `inspect a b c`), so the cost of a status run barely
// grows with the number of instances. Only the sampling calls under --details
// (kubectl top, docker stats) and the in-container script exec are per-run or
// per-instance costs beyond that.
//
// Every fact collected here lands in an internal/statusreport model and is
// rendered from it, so the table view and --output json can never disagree.

import (
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"time"

	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/gen"
	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/runner"
	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/spec"
	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/statusreport"
	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/statusscript"
	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/validate"
)

// The status target words. They are the modeled cliTarget names in commands.go;
// these constants are what the parsed word is compared against, so a rename
// there without one here fails TestStatusTargetsMatchModel rather than
// silently accepting a word no view implements.
const (
	statusTargetContainer   = "container"
	statusTargetApplication = "application"
	statusTargetAll         = "all"
)

// Output kinds for --output.
const (
	outputTable = "table"
	outputJSON  = "json"
)

// Watch interval bounds. The default is the only documented value; see
// watchFlag for why an override exists but is not advertised.
const (
	watchDefaultSeconds = 5
	watchMinSeconds     = 1
	watchMaxSeconds     = 3600
)

// repeatableName implements flag.Value for a repeatable, free-form name flag
// (--pod, --container): it only collects raw values here, since the platform
// (and therefore which flag applies) is not known until after parsing;
// actStatus validates each one against validate.SafeToken before it reaches
// an argv.
type repeatableName struct{ vals *[]string }

func (repeatableName) String() string { return "" } // flag.Value needs a zero-value String; nothing to show before parsing

func (v repeatableName) Set(s string) error {
	*v.vals = append(*v.vals, s)
	return nil
}

// watchFlag is --watch/-w: a flag that is boolean in every documented sense
// (bare -w re-renders every watchDefaultSeconds) but also accepts an interval
// as -w=<seconds>. IsBoolFlag is what makes the bare spelling legal, since Go's
// flag package otherwise demands a value for any flag.Value.
//
// The =<seconds> form is intentionally left out of the help text, the modeled
// flag meaning, docs/commands.md, the user guide, and the completion scripts:
// it exists for tightening the loop while working on the report, not as a knob
// operators are meant to tune, and one documented interval keeps the output's
// header line honest. This comment is the record of that being deliberate
// rather than an omission -- do not "fix" the docs by adding it.
type watchFlag struct {
	on       bool
	interval time.Duration
}

func (w *watchFlag) IsBoolFlag() bool { return true }

func (w *watchFlag) String() string { return "" }

func (w *watchFlag) Set(s string) error {
	switch s {
	case "true":
		w.on, w.interval = true, watchDefaultSeconds*time.Second
		return nil
	case "false":
		w.on = false
		return nil
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return fmt.Errorf("--watch takes no value (or a whole number of seconds), got %q", s)
	}
	if n < watchMinSeconds || n > watchMaxSeconds {
		return fmt.Errorf("--watch interval %d must be %d-%d seconds", n, watchMinSeconds, watchMaxSeconds)
	}
	w.on, w.interval = true, time.Duration(n)*time.Second
	return nil
}

// statusOpts is one parsed status invocation: the operator's whole intent,
// separated from how it was spelled, so actStatus and the collectors take a
// value rather than a dozen positional arguments.
type statusOpts struct {
	view  statusreport.View
	level statusreport.Level
	json  bool
	watch watchFlag

	envPath  string
	install  bool
	platform string
	pods     []string
	conts    []string
	ns       string
	port     int
	user     string
	command  string
	allow    []string
	all      bool

	// now is the clock ages are measured against, injected here rather than
	// read inside the collectors so a test can pin it.
	now time.Time
}

// runStatus parses status's target word and flags and hands them to actStatus.
func runStatus(args []string, r runner.Runner) int {
	fs := verbFlagSet("status")
	env := envFlag(fs)
	install := fs.Bool("install", false, "install the status script on every target without prompting")
	platform := platformFlag(fs)
	var pods, containers []string
	fs.Var(repeatableName{&pods}, "pod", "limit checks to this kubernetes pod name (repeatable)")
	fs.Var(repeatableName{&containers}, "container", "limit checks to this docker/podman container name (repeatable)")
	namespace := fs.String("namespace", "", "kubernetes namespace to query")
	port := fs.Int("management-port", 0, "actuator management port to reach inside each target")
	user := fs.String("user", "", "actuator account the status script authenticates as (default "+spec.StatusUserName+")")
	command := fs.String("command", "", "override the platform CLI binary used to reach each target")
	allow := allowCommandFlag(fs)
	const detailsUsage = "add the enrichment lines: node, resource use, digest and components; app version, java, config and heap"
	details := fs.Bool("details", false, detailsUsage)
	fs.BoolVar(details, "d", false, detailsUsage)
	all := fs.Bool("all", false, "report every connector instance on the platform, found by image name, instead of the ones env.yaml describes")
	output := fs.String("output", outputTable, "output format: table or json")
	var watch watchFlag
	// Not a const: the default interval is a number, and the usage text names it
	// so the two can never disagree.
	watchUsage := fmt.Sprintf("re-render the report every %ds until interrupted", watchDefaultSeconds)
	fs.Var(&watch, "watch", watchUsage)
	fs.Var(&watch, "w", watchUsage)

	pos, err := collectFlagsAndDirs(fs, args)
	if err != nil {
		return flagExit("status", err)
	}

	view, code := statusView(pos)
	if code != 0 {
		return code
	}

	opts := statusOpts{
		view: view, json: *output == outputJSON, watch: watch,
		envPath: *env, install: *install, platform: *platform,
		pods: pods, conts: containers, ns: *namespace, port: *port,
		user: *user, command: *command, allow: *allow, all: *all,
		now: time.Now(),
	}
	if *details {
		opts.level = statusreport.LevelDetails
	}
	if code := checkStatusFlags(opts, *output); code != 0 {
		return code
	}
	return actStatus(opts, r)
}

// statusView resolves the target word into a view. The word is required: the
// two views answer different questions and neither is a safe guess, so a bare
// `status` prints the verb's own usage and exits 2 rather than picking one --
// the same shape `download` uses for its missing target word.
func statusView(pos []string) (statusreport.View, int) {
	if len(pos) == 0 {
		fmt.Fprintf(os.Stderr, "status: missing target (%s)\n\n", pipeList(targetNames("status")))
		fmt.Fprint(os.Stderr, verbUsage("status"))
		return 0, 2
	}
	if len(pos) > 1 {
		fmt.Fprintf(os.Stderr, "status: unexpected argument %q\n", pos[1])
		return 0, 2
	}
	switch resolveTarget("status", pos[0]) {
	case statusTargetContainer:
		return statusreport.ViewContainer, 0
	case statusTargetApplication:
		return statusreport.ViewApplication, 0
	case statusTargetAll:
		return statusreport.ViewAll, 0
	}
	fmt.Fprintf(os.Stderr, "status: unknown target %q (want %s)\n", pos[0], wantList(targetNames("status")))
	return 0, 2
}

// checkStatusFlags rejects the combinations that cannot mean anything, loudly
// and before any query runs. A flag that would simply be ignored is the kind of
// silent surprise this tool refuses elsewhere too: an operator who typed
// --install on the container view expects it to do something.
func checkStatusFlags(o statusOpts, output string) int {
	switch output {
	case outputTable, outputJSON:
	default:
		fmt.Fprintf(os.Stderr, "status: unknown --output %q (want %s)\n", output, wantList([]string{outputTable, outputJSON}))
		return 2
	}
	if o.json && o.watch.on {
		// A redraw loop emitting a document per tick is neither a stream a parser
		// can read nor a screen a human can, so the two are refused together
		// rather than one silently winning.
		fmt.Fprintln(os.Stderr, "status: --output json cannot be combined with --watch")
		return 2
	}
	if o.all && (len(o.pods) > 0 || len(o.conts) > 0) {
		fmt.Fprintln(os.Stderr, "status: --all searches for every instance by image name, so it cannot be combined with --pod/--container")
		return 2
	}
	if o.view == statusreport.ViewContainer {
		// These three only steer the in-container script, which the container
		// view never runs.
		for _, f := range []struct {
			set  bool
			name string
		}{
			{o.install, "--install"},
			{o.user != "", "--user"},
			{o.port != 0, "--management-port"},
		} {
			if f.set {
				fmt.Fprintf(os.Stderr, "status: %s applies to the %s and %s views, which run the in-container script; %s does not\n",
					f.name, statusTargetApplication, statusTargetAll, statusTargetContainer)
				return 2
			}
		}
	}
	return 0
}

// actStatus resolves the platform and its targets and renders the report, once
// or on a watch loop. Everything that can fail the whole run -- an unreadable
// env.yaml, an unsafe name, a platform that cannot be resolved, a failed
// preflight -- fails here, before any instance is queried.
func actStatus(o statusOpts, r runner.Runner) int {
	sess, serr := resolveInstanceSession(instanceRequest{
		envPath:  o.envPath,
		platform: o.platform,
		ns:       o.ns,
		command:  o.command,
		pods:     o.pods,
		conts:    o.conts,
		allow:    o.allow,
		all:      o.all,
	})
	if serr != nil {
		return errExit(serr)
	}
	o.ns = sess.ns

	if o.user == "" {
		o.user = spec.StatusUserName
	}
	// Stricter than the other flags: the account name is also spliced into a
	// sed address inside the generated script, where a '/' or a regex
	// metacharacter would break the password lookup rather than fail loudly.
	if !validate.SafeActuatorUser(o.user) {
		return errExit(fmt.Errorf("--user %q is not a usable account name (%s)", o.user, validate.SafeActuatorUserReason))
	}

	if o.port == 0 {
		o.port = sess.env.Defaults.EffectiveManagementPort()
	}
	if o.port < 1 || o.port > 65535 {
		return errExit(fmt.Errorf("--management-port %d must be 1-65535", o.port))
	}

	// Reuse the read-only preflight probe before touching anything, same as
	// deploy/remove. The action argument only steers kubernetes' can-i verb
	// (docker/podman ignore it); status never creates or deletes a deployment,
	// so deploy's "create" is used as the closer of the two existing checks.
	if code, ok := preflight(r, runner.ActionDeploy, sess.platform, sess.command, sess.ns, o.allow); !ok {
		return code
	}

	c := &statusCollector{opts: o, platform: sess.platform, cmdArgv: sess.cmdArgv, env: sess.env, r: r}
	if o.watch.on {
		return watchStatus(o, c)
	}
	return c.renderOnce(os.Stdout)
}

// ---- collection --------------------------------------------------------------

// statusCollector holds everything one status run needs, so a watch tick can
// re-collect without re-resolving the platform, re-validating names, or
// re-running the preflight probe -- none of which can change between ticks.
type statusCollector struct {
	opts     statusOpts
	platform string
	cmdArgv  []string
	env      *spec.Env
	r        runner.Runner

	// promptedInstall records that the one install confirmation this run is
	// allowed has already been asked. A watch loop must not re-prompt on every
	// tick: the answer cannot change, and a prompt would block the redraw.
	promptedInstall bool
}

// wantsContainerFacts reports whether the operator asked for the engine's half
// of the report. The basic engine read happens either way -- the application
// view needs it to name each instance and to have something to say when the
// script cannot be run -- but the workload summary and the enrichment queries
// are collected only when something is going to render them.
func (c *statusCollector) wantsContainerFacts() bool {
	return c.opts.view != statusreport.ViewApplication
}

// session re-presents what actStatus already resolved in the shape the shared
// discovery helpers take, so status and logs ask them the same question.
func (c *statusCollector) session() instanceSession {
	return instanceSession{env: c.env, platform: c.platform, cmdArgv: c.cmdArgv, ns: c.opts.ns}
}

// collect gathers the whole report. It returns the report plus whether any
// instance could not be reached and run, which is what the exit code is about
// -- never which instance is active, and never whether an engine query
// degraded (that is a note in the report).
func (c *statusCollector) collect() (statusreport.Report, bool, error) {
	rep := statusreport.Report{Platform: c.platform, Namespace: c.opts.ns}
	var err error
	switch c.platform {
	case validate.PlatformKubernetes:
		err = c.collectKubernetes(&rep)
	case validate.PlatformDocker, validate.PlatformPodman:
		err = c.collectEngine(&rep)
	default:
		err = fmt.Errorf("unknown platform %q", c.platform)
	}
	if err != nil {
		return rep, true, err
	}
	statusreport.SortInstances(rep.Instances)
	c.compareImage(&rep)

	failed := false
	if c.opts.view != statusreport.ViewContainer {
		failed = c.collectApplication(&rep)
	}
	return rep, failed, nil
}

// collectKubernetes reads every pod's facts, and the workload above them, in
// as few calls as the CLI allows: one `get pods` answers discovery and the
// whole container view together.
func (c *statusCollector) collectKubernetes(rep *statusreport.Report) error {
	o := c.opts
	selector, match, group, derr := kubeDiscovery(c.session(), o.pods, o.all)
	if derr != nil {
		return derr
	}
	rep.Group = group

	doc, err := runner.KubernetesPodsJSON(c.r, c.cmdArgv, o.ns, selector, o.pods, o.all)
	if err != nil {
		return err
	}
	insts, perr := statusreport.ParsePods(doc, o.now, match)
	if perr != nil {
		return fmt.Errorf("reading the pod list: %w", perr)
	}
	if len(insts) == 0 {
		return noPodsFound(selector, o.ns, o.all)
	}
	rep.Instances = insts

	// The workload summary is the deployment and service this tool created; a
	// cluster-wide search has no single one, and an instance this tool never
	// deployed may have neither. Both reads are best-effort: a note, never a
	// failure, since the per-pod facts above are the report.
	if c.wantsContainerFacts() && !o.all && c.env != nil && c.env.Kubernetes != nil {
		name := c.env.Kubernetes.Deployment.Name
		if doc, derr := runner.KubernetesGetJSON(c.r, c.cmdArgv, o.ns, "deployment", name); derr != nil {
			rep.Notes = append(rep.Notes, fmt.Sprintf("could not read deployment %s: %v", name, derr))
		} else if w, werr := statusreport.ParseDeployment(doc); werr == nil {
			rep.Workload = w
		}
		if c.env.Kubernetes.Service.Enabled {
			if doc, serr := runner.KubernetesGetJSON(c.r, c.cmdArgv, o.ns, "service", name); serr == nil {
				rep.Workload, _ = statusreport.MergeService(rep.Workload, doc)
			}
		}
	}

	if o.level != statusreport.LevelDetails || !c.wantsContainerFacts() {
		return nil
	}
	if out, terr := runner.KubernetesTop(c.r, c.cmdArgv, o.ns, selector, o.pods, o.all); terr != nil {
		// The metrics API is an optional cluster add-on, so its absence is a
		// note against the resource lines rather than a failed run.
		rep.Notes = append(rep.Notes, fmt.Sprintf("no resource usage: %v; install metrics-server, or drop --details", terr))
	} else {
		statusreport.ApplyTop(rep.Instances, out)
	}
	c.checkComponents(rep)
	return nil
}

// checkComponents fills in whether each object a pod references actually
// exists. Deduplicated across instances first: every pod of one deployment
// references the same configmap and secrets, and asking once per pod would
// multiply the calls by the replica count for the same answer.
func (c *statusCollector) checkComponents(rep *statusreport.Report) {
	type key struct{ ns, kind, name string }
	seen := map[key]statusreport.Component{}
	for i := range rep.Instances {
		inst := &rep.Instances[i]
		if inst.Container == nil {
			continue
		}
		for j := range inst.Container.Components {
			comp := &inst.Container.Components[j]
			k := key{inst.Namespace, comp.Kind, comp.Name}
			if cached, ok := seen[k]; ok {
				comp.Status = cached.Status
				continue
			}
			if !validate.SafeToken(comp.Name) {
				// The name came back from the cluster rather than from the
				// operator; one that could not go into an argv is skipped and
				// said so, never passed on.
				comp.Status = "unchecked"
				seen[k] = *comp
				continue
			}
			doc, err := runner.KubernetesGetJSON(c.r, c.cmdArgv, inst.Namespace, comp.Kind, comp.Name)
			if err != nil {
				comp.Status = "MISSING"
			} else if ok, status := statusreport.ObjectExists(doc); ok {
				comp.Status = status
			} else {
				comp.Status = "MISSING"
			}
			seen[k] = *comp
		}
	}
}

// collectEngine reads every container's facts on docker or podman: one inspect
// for all targets, which also carries the compose project label, plus the
// digest and (on podman) the systemd restart count.
func (c *statusCollector) collectEngine(rep *statusreport.Report) error {
	o := c.opts
	match := ""
	if o.all {
		match = statusreport.ImageMatch
	}
	names, notes, nerr := engineNames(c.r, c.session(), o.conts, o.all)
	rep.Notes = append(rep.Notes, notes...)
	if nerr != nil {
		return nerr
	}

	doc, err := runner.EngineInspectJSON(c.r, c.cmdArgv, names)
	if err != nil {
		return err
	}
	insts, perr := statusreport.ParseInspect(doc, o.now, match)
	if perr != nil {
		return fmt.Errorf("reading the container details: %w", perr)
	}
	if len(insts) == 0 {
		if match != "" {
			return fmt.Errorf("no container on this host is running an image matching %q", match)
		}
		return fmt.Errorf("the engine reported nothing for %s", strings.Join(names, ", "))
	}
	rep.Instances = insts

	// The restart count is a basic column, so the truthful one has to be
	// collected at the basic level too -- one systemctl read per instance, on
	// podman only.
	c.applyPodmanRestarts(rep)

	if o.level != statusreport.LevelDetails || !c.wantsContainerFacts() {
		return nil
	}
	// The digest costs a second call here (docker and podman report it on the
	// image, not the container), so it belongs to the level that pays for
	// enrichment. On kubernetes it arrives free in the pod document either way.
	c.applyDigests(rep)
	if out, serr := runner.EngineStats(c.r, c.cmdArgv, names); serr != nil {
		rep.Notes = append(rep.Notes, fmt.Sprintf("no resource usage: %v", serr))
	} else {
		statusreport.ApplyStats(rep.Instances, out)
	}
	return nil
}

// applyPodmanRestarts replaces the container's own restart counter with the one
// systemd keeps, which is the only truthful count under quadlet: a restart
// there recreates the container, so its own counter is always 0.
//
// Best-effort by nature -- a container that is not quadlet-managed, a host
// without systemd, or a podman machine on another OS all simply leave the
// container's own number in place.
func (c *statusCollector) applyPodmanRestarts(rep *statusreport.Report) {
	if c.platform != validate.PlatformPodman {
		return
	}
	scope, dir := "", ""
	if c.env != nil && c.env.Podman != nil && c.env.Podman.Quadlet != nil {
		scope, dir = c.env.Podman.Quadlet.Scope, c.env.Podman.Quadlet.Dir
	}
	sc, err := runner.ResolveQuadletScope(scope, dir)
	if err != nil {
		return
	}
	for i := range rep.Instances {
		inst := &rep.Instances[i]
		if inst.Container == nil {
			continue
		}
		n, nerr := runner.SystemctlNRestarts(c.r, sc, gen.PodmanServiceName(inst.Name))
		if nerr != nil {
			continue
		}
		inst.Container.Restarts = n
		inst.Container.RestartSource = "systemd"
	}
}

// applyDigests fills in the registry digest each instance is actually running,
// which docker and podman report on the image rather than on the container.
// One lookup per distinct image reference, not per instance.
func (c *statusCollector) applyDigests(rep *statusreport.Report) {
	digests := map[string]string{}
	for i := range rep.Instances {
		inst := &rep.Instances[i]
		if inst.Container == nil || inst.Container.Image == "" {
			continue
		}
		ref := inst.Container.Image
		d, ok := digests[ref]
		if !ok {
			if !validate.SafeToken(ref) {
				// Read back off the container, so it is re-checked here before it
				// could reach an argv.
				digests[ref] = ""
				continue
			}
			if doc, err := runner.EngineImageInspectJSON(c.r, c.cmdArgv, ref); err == nil {
				d = statusreport.ParseImageDigest(doc)
			}
			digests[ref] = d
		}
		inst.Container.Digest = d
	}
}

// compareImage flags an instance whose running image is not the one env.yaml
// asks for -- the failed-rollout case, where a pod is still on the old tag and
// nothing else in the report would say so. Only ever set on a mismatch, so the
// line's presence is the finding.
func (c *statusCollector) compareImage(rep *statusreport.Report) {
	if c.env == nil || c.env.Image == nil {
		return
	}
	want := c.env.Image.Ref()
	if want == "" {
		return
	}
	mismatched := 0
	for i := range rep.Instances {
		inst := &rep.Instances[i]
		if inst.Container == nil {
			continue
		}
		if statusreport.ImageMismatch(inst.Container.Image, inst.Container.Digest, want) {
			inst.Container.ImageExpected = want
			mismatched++
		}
	}
	// The per-instance line above is only rendered by the details block, so at
	// the basic level the finding would otherwise be invisible -- and this is
	// the finding the comparison exists for. One note, only ever on a mismatch.
	if mismatched > 0 && c.opts.level != statusreport.LevelDetails {
		rep.Notes = append(rep.Notes, fmt.Sprintf(
			"%d of %d instance(s) are not running the image env.yaml configures (%s); run with %s for the per-instance detail",
			mismatched, len(rep.Instances), want, detailsFlagName))
	}
}

// ---- the application half ----------------------------------------------------

// collectApplication ensures the status script is present on each instance and
// runs it, filling in each instance's application block. It reports whether any
// instance could not be reached and run.
//
// The probe/install phase is separate from the run phase on purpose: the
// install confirmation is asked once for every instance that needs it, rather
// than once per instance.
func (c *statusCollector) collectApplication(rep *statusreport.Report) bool {
	script := statusscript.Render(c.opts.port, c.opts.user)
	failed := false

	// Indexed by namespace and name, not by name alone: under --all two pods in
	// different namespaces can share a name, and a bare-name key would report
	// one of them with the other's answer.
	var missing []int
	present := map[int]bool{}
	for i := range rep.Instances {
		inst := &rep.Instances[i]
		ok, err := runner.ScriptInstalled(c.r, c.cmdArgv, c.platform, inst.Name, inst.Namespace, statusscript.ContainerPath)
		if err != nil {
			inst.Error = fmt.Sprintf("could not check for the status script: %v", err)
			failed = true
			continue
		}
		if ok {
			present[i] = true
		} else {
			missing = append(missing, i)
		}
	}

	if len(missing) > 0 {
		doInstall := c.opts.install
		if !doInstall && !c.promptedInstall {
			c.promptedInstall = true
			answer, cerr := confirmInstall(instanceNames(rep, missing))
			if cerr != nil {
				// The prompt could not be asked at all (stdin is not a TTY). That
				// cannot be resolved for these instances, but the ones that already
				// carry the script are still worth reporting, so they are marked and
				// the run carries on rather than being abandoned.
				rep.Notes = append(rep.Notes, cerr.Error())
				markMissing(rep, missing, "the status script is not installed")
				missing = nil
				failed = true
			} else {
				doInstall = answer
			}
		}
		for _, i := range missing {
			inst := &rep.Instances[i]
			if !doInstall {
				inst.Error = "the status script is not installed, and the install prompt was declined"
				failed = true
				continue
			}
			if _, err := runner.InstallScript(c.r, c.cmdArgv, c.platform, inst.Name, inst.Namespace, statusscript.ContainerDir, statusscript.ContainerPath, script); err != nil {
				inst.Error = fmt.Sprintf("could not install the status script: %v", err)
				failed = true
				continue
			}
			present[i] = true
		}
	}

	for i := range rep.Instances {
		inst := &rep.Instances[i]
		if !present[i] {
			continue
		}
		// The script always exits 0 and puts its findings in the output, so an
		// error here is the exec failing rather than a standby instance -- report
		// it against this instance and keep going, so one unreachable instance
		// does not hide the rest.
		out, err := runner.RunStatusScript(c.r, c.cmdArgv, c.platform, inst.Name, inst.Namespace, statusscript.ContainerPath)
		if err != nil {
			inst.Error = fmt.Sprintf("could not run the status script: %v", err)
			failed = true
			continue
		}
		inst.Application = statusreport.ParseApplication(out)
	}
	return failed
}

// markMissing records why each of the given instances has no application
// block, keeping the reason with the instance rather than on stderr where
// nothing explains it.
func markMissing(rep *statusreport.Report, idx []int, reason string) {
	for _, i := range idx {
		rep.Instances[i].Error = reason
	}
}

// instanceNames is the list of names the install prompt shows, in report order.
func instanceNames(rep *statusreport.Report, idx []int) []string {
	names := make([]string, 0, len(idx))
	for _, i := range idx {
		names = append(names, rep.Instances[i].Name)
	}
	return names
}

// confirmInstall asks once whether to install the status script on the listed
// missing instances, via the same promptLine seam and non-TTY refusal
// promptPlatformMenu uses. "y"/"yes" (case-insensitive) installs; anything
// else, including a blank line, declines.
func confirmInstall(missing []string) (bool, error) {
	line, err := promptLine(fmt.Sprintf("status script missing on %s -- install it now? [y/N] ", strings.Join(missing, ", ")))
	if err != nil {
		return false, fmt.Errorf("confirming the status script install interactively: %w; pass --install instead", err)
	}
	switch strings.ToLower(line) {
	case "y", "yes":
		return true, nil
	default:
		return false, nil
	}
}

// ---- rendering ---------------------------------------------------------------

// renderOnce collects and prints one report. A collection error that stopped
// the run (nothing to report on at all) goes to stderr with the exit code, the
// way every other verb reports a failure; anything the report itself can carry
// is already in it.
func (c *statusCollector) renderOnce(w *os.File) int {
	rep, failed, err := c.collect()
	if err != nil {
		return errExit(err)
	}
	if c.opts.json {
		doc, jerr := statusreport.JSON(rep)
		if jerr != nil {
			return errExit(jerr)
		}
		fmt.Fprint(w, doc)
	} else {
		for _, line := range statusreport.Render(rep, c.opts.view, c.opts.level) {
			fmt.Fprintln(w, line)
		}
	}
	if failed {
		return 1
	}
	return 0
}

// watchStatus re-renders the report until the operator interrupts it. Each tick
// clears the screen and prints a header naming the interval and the time, so a
// glance tells whether what is on screen is current.
//
// Ctrl-C exits 0: a watch that was ended on purpose did not fail. The exit code
// of an individual tick is deliberately not carried out here -- there is no
// single answer for a loop that ran for an hour.
func watchStatus(o statusOpts, c *statusCollector) int {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	defer signal.Stop(stop)

	tick := time.NewTicker(o.watch.interval)
	defer tick.Stop()
	for {
		c.opts.now = time.Now()
		clearScreen(os.Stdout)
		fmt.Fprintf(os.Stdout, "Every %s: solmq-conn-util status  --  %s   (Ctrl-C to stop)\n\n",
			o.watch.interval, c.opts.now.Format("15:04:05"))
		c.renderOnce(os.Stdout)
		select {
		case <-stop:
			fmt.Fprintln(os.Stdout)
			return 0
		case <-tick.C:
		}
	}
}

// clearScreen moves the cursor home and clears what is below it, so a redraw
// replaces the previous report instead of scrolling it away. ANSI, which every
// terminal this tool targets understands once VT processing is on -- see
// enableVirtualTerminal, which turns it on where it is not on by default.
func clearScreen(w *os.File) {
	enableVirtualTerminal(w)
	fmt.Fprint(w, "\x1b[H\x1b[2J")
}
