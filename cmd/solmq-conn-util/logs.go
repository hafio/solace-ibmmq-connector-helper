package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"time"

	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/runner"
	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/statusreport"
	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/validate"
)

// Tail bounds. The floor is 0 because "the flags only, no history" is a real
// request; the ceiling is not a platform limit but a typo guard -- a six-digit
// line count is a slipped keystroke far more often than an intention, and every
// platform reads the whole log for "all" anyway, which --tail all says
// directly.
const (
	tailMin     = 0
	tailMax     = 100000
	tailAllWord = "all"
)

// tailFlag is --tail: a whole number of trailing lines, or the word "all".
// runner.TailAll is the unset value rather than 0, since 0 is a request the
// platforms honour.
type tailFlag struct{ n int }

func (t *tailFlag) String() string { return "" } // flag.Value needs a zero-value String; nothing to show before parsing

func (t *tailFlag) Set(s string) error {
	if s == tailAllWord {
		t.n = runner.TailAll
		return nil
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return fmt.Errorf("--tail takes a whole number of lines or %q, got %q", tailAllWord, s)
	}
	if n < tailMin || n > tailMax {
		return fmt.Errorf("--tail %d must be %d-%d, or %q", n, tailMin, tailMax, tailAllWord)
	}
	t.n = n
	return nil
}

// sinceFlag is --since: how far back to start reading. The operator's spelling
// is parsed here and only the canonical form is kept, so what reaches an argv
// is a duration this tool produced rather than a string it was handed. That is
// what keeps --since out of the SafeToken checks the name flags need.
type sinceFlag struct{ d string }

func (f *sinceFlag) String() string { return "" } // flag.Value needs a zero-value String; nothing to show before parsing

func (f *sinceFlag) Set(s string) error {
	d, err := time.ParseDuration(s)
	if err != nil {
		return fmt.Errorf("--since %q is not a duration (want e.g. 30s, 10m, 2h)", s)
	}
	if d <= 0 {
		return fmt.Errorf("--since %q must be a positive duration", s)
	}
	f.d = d.String()
	return nil
}

// logsOpts is one parsed logs invocation, separated from how it was spelled the
// same way statusOpts is.
type logsOpts struct {
	envPath  string
	platform string
	pods     []string
	conts    []string
	ns       string
	command  string
	allow    []string
	all      bool

	follow     bool
	previous   bool
	timestamps bool
	tail       int
	since      string
}

// forRunner is the subset LogsArgv needs, per target: the container name is the
// one thing that varies between them.
func (o logsOpts) forRunner(container string) runner.LogsOpts {
	return runner.LogsOpts{
		Follow:     o.follow,
		Previous:   o.previous,
		Timestamps: o.timestamps,
		Tail:       o.tail,
		Since:      o.since,
		Container:  container,
	}
}

// logTarget is one instance's log, already addressed: what to name on the
// command line, and where.
type logTarget struct {
	name      string
	namespace string
	// container is which container inside a kubernetes pod to read, as
	// statusreport picked it. Empty on docker/podman, and empty for a pod whose
	// containers this tool cannot tell apart -- kubectl then says so itself,
	// which is a better answer than guessing.
	container string
}

// runLogs parses logs's flags and hands them to actLogs.
func runLogs(args []string, r runner.Runner) int {
	fs := verbFlagSet("logs")
	env := envFlag(fs)
	platform := platformFlag(fs)
	var pods, containers []string
	fs.Var(repeatableName{&pods}, "pod", "read this kubernetes pod's log (repeatable)")
	fs.Var(repeatableName{&containers}, "container", "read this docker/podman container's log (repeatable)")
	namespace := fs.String("namespace", "", "kubernetes namespace to query")
	command := fs.String("command", "", "override the platform CLI binary used to reach each target")
	allow := allowCommandFlag(fs)
	all := fs.Bool("all", false, "read every connector instance on the platform, found by image name, instead of the ones env.yaml describes")
	follow := fs.Bool("follow", false, "keep the log open and print new lines as they arrive, until interrupted")
	previous := fs.Bool("previous", false, "read the previous container's log instead of the running one (kubernetes only)")
	timestamps := fs.Bool("timestamps", false, "prefix every line with the time the platform recorded for it")
	tail := &tailFlag{n: runner.TailAll}
	fs.Var(tail, "tail", "read only the last N lines (default: the whole log)")
	var since sinceFlag
	fs.Var(&since, "since", "read only lines newer than this duration, e.g. 10m")

	pos, err := collectFlagsAndDirs(fs, args)
	if err != nil {
		return flagExit("logs", err)
	}
	if len(pos) > 0 {
		// logs has no target word: an instance is named with --pod/--container,
		// not positionally, so a bare word here is a mistake worth naming rather
		// than a name to guess at.
		fmt.Fprintf(os.Stderr, "logs: unexpected argument %q (name an instance with %s or %s)\n", pos[0], podFlagName, containerFlagName)
		return 2
	}

	o := logsOpts{
		envPath:    *env,
		platform:   *platform,
		pods:       pods,
		conts:      containers,
		ns:         *namespace,
		command:    *command,
		allow:      *allow,
		all:        *all,
		follow:     *follow,
		previous:   *previous,
		timestamps: *timestamps,
		tail:       tail.n,
		since:      since.d,
	}
	if code := checkLogsFlags(o); code != 0 {
		return code
	}
	return actLogs(o, r)
}

// checkLogsFlags rejects the combinations that cannot mean anything, before any
// query runs -- the same contract checkStatusFlags keeps. Only the
// platform-independent ones live here; --previous against an engine that has no
// such concept cannot be judged until the platform is resolved.
func checkLogsFlags(o logsOpts) int {
	if o.follow && o.previous {
		fmt.Fprintf(os.Stderr, "logs: %s reads a log that has already ended, so it cannot be combined with %s\n", previousFlagName, followFlagName)
		return 2
	}
	if o.follow && o.all {
		fmt.Fprintf(os.Stderr, "logs: %s searches for every instance by image name, and %s reads one; name a single instance with %s or %s instead\n",
			allFlagName, followFlagName, podFlagName, containerFlagName)
		return 2
	}
	if o.all && (len(o.pods) > 0 || len(o.conts) > 0) {
		fmt.Fprintf(os.Stderr, "logs: %s searches for every instance by image name, so it cannot be combined with %s/%s\n", allFlagName, podFlagName, containerFlagName)
		return 2
	}
	return 0
}

// actLogs resolves the platform and its targets and reads each one's log.
// Everything that can fail the whole run -- an unreadable env.yaml, an unsafe
// name, a platform that cannot be resolved, a failed preflight -- fails here,
// before any log is read.
func actLogs(o logsOpts, r runner.Runner) int {
	// Both paths stream: a one-shot read wants the platform's diagnostics on
	// stderr just as much as a followed one does, because `logs > app.log` is
	// the whole point of the verb and a redirect that swallowed them would be
	// the first thing to confuse an operator.
	s, ok := r.(runner.Streamer)
	if !ok {
		return errExit(errors.New("logs needs a runner that can stream command output, and this one cannot"))
	}

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

	if o.previous && sess.platform != validate.PlatformKubernetes {
		fmt.Fprintf(os.Stderr, "logs: %s is a kubernetes concept; %s keeps no prior-run log to read\n", previousFlagName, sess.platform)
		return 2
	}

	// The same read-only preflight probe status runs, for the same reason: fail
	// on an unreachable engine once, up front, rather than once per target.
	if code, ok := preflight(r, runner.ActionDeploy, sess.platform, sess.command, sess.ns, o.allow); !ok {
		return code
	}

	targets, terr := logTargets(r, sess, o)
	if terr != nil {
		return errExit(terr)
	}

	if o.follow && len(targets) > 1 {
		return errExit(fmt.Errorf("%s reads one instance, and %d match (%s); name one with %s",
			followFlagName, len(targets), strings.Join(targetNamesOf(targets), ", "), podOrContainerFlag(sess.platform)))
	}

	if o.follow {
		return followLog(s, sess, o, targets[0])
	}
	return readLogs(s, sess, o, targets)
}

// logTargets resolves which instances are meant, using the same discovery the
// status verb does so the two verbs can never disagree about which pods a bare
// invocation means.
func logTargets(r runner.Runner, s instanceSession, o logsOpts) ([]logTarget, error) {
	if s.platform == validate.PlatformKubernetes {
		selector, match, _, derr := kubeDiscovery(s, o.pods, o.all)
		if derr != nil {
			return nil, derr
		}
		doc, err := runner.KubernetesPodsJSON(r, s.cmdArgv, s.ns, selector, o.pods, o.all)
		if err != nil {
			return nil, err
		}
		// ParsePods is asked only for names here, so the clock it measures ages
		// against is never read; a zero time keeps the call honest about that.
		insts, perr := statusreport.ParsePods(doc, time.Time{}, match)
		if perr != nil {
			return nil, fmt.Errorf("reading the pod list: %w", perr)
		}
		if len(insts) == 0 {
			return nil, noPodsFound(selector, s.ns, o.all)
		}
		targets := make([]logTarget, 0, len(insts))
		for _, inst := range insts {
			ns := inst.Namespace
			if ns == "" {
				ns = s.ns
			}
			targets = append(targets, logTarget{name: inst.Name, namespace: ns, container: inst.ContainerName})
		}
		return targets, nil
	}

	names, notes, nerr := engineNames(r, s, o.conts, o.all)
	for _, n := range notes {
		fmt.Fprintln(os.Stderr, "logs:", n)
	}
	if nerr != nil {
		return nil, nerr
	}
	targets := make([]logTarget, 0, len(names))
	for _, n := range names {
		targets = append(targets, logTarget{name: n})
	}
	return targets, nil
}

// readLogs prints each target's log once, in discovery order.
//
// A target that cannot be read is reported against itself and the rest are
// still read, the same rule status keeps: one unreachable instance must not
// hide the others. The run then exits 1, because something the operator asked
// for did not arrive.
func readLogs(s runner.Streamer, sess instanceSession, o logsOpts, targets []logTarget) int {
	failed := false
	for i, t := range targets {
		if len(targets) > 1 {
			// A single target prints its log and nothing else, so the common case
			// pipes into another tool cleanly; several need saying apart.
			if i > 0 {
				fmt.Fprintln(os.Stdout)
			}
			fmt.Fprintf(os.Stdout, "==> %s <==\n", t.name)
		}
		argv, err := runner.LogsArgv(sess.cmdArgv, sess.platform, t.name, t.namespace, o.forRunner(t.container))
		if err != nil {
			fmt.Fprintf(os.Stderr, "logs: %s: %v\n", t.name, err)
			failed = true
			continue
		}
		if err := s.Stream(context.Background(), runner.Cmd{Argv: argv}, os.Stdout, os.Stderr); err != nil {
			fmt.Fprintf(os.Stderr, "logs: could not read %s: %v\n", t.name, err)
			failed = true
		}
	}
	if failed {
		return 1
	}
	return 0
}

// followLog keeps one target's log open until the operator interrupts it.
//
// Ctrl-C is how a follow is meant to end, so it exits 0 -- the same reasoning
// watchStatus records for --watch. Stream reports a cancelled context as
// success, so there is nothing here to unpick.
func followLog(s runner.Streamer, sess instanceSession, o logsOpts, t logTarget) int {
	argv, err := runner.LogsArgv(sess.cmdArgv, sess.platform, t.name, t.namespace, o.forRunner(t.container))
	if err != nil {
		return errExit(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	defer signal.Stop(stop)
	go func() {
		<-stop
		cancel()
	}()

	if err := s.Stream(ctx, runner.Cmd{Argv: argv}, os.Stdout, os.Stderr); err != nil {
		return errExit(fmt.Errorf("following %s: %w", t.name, err))
	}
	return 0
}

// targetNamesOf lists the instance names, for a message that has to say which
// ones it found.
func targetNamesOf(targets []logTarget) []string {
	names := make([]string, len(targets))
	for i, t := range targets {
		names[i] = t.name
	}
	return names
}

// podOrContainerFlag names the flag that narrows to one instance on this
// platform, so the message points at the flag that exists rather than at both.
func podOrContainerFlag(platform string) string {
	if platform == validate.PlatformKubernetes {
		return podFlagName
	}
	return containerFlagName
}
