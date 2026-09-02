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

	follow     bool
	previous   bool
	timestamps bool
	tail       int
	since      string
}

// forRunner is the subset LogsArgv needs. Which container to read is not among
// it: LogsArgv names spec.ConnectorContainerName itself.
func (o logsOpts) forRunner() runner.LogsOpts {
	return runner.LogsOpts{
		Follow:     o.follow,
		Previous:   o.previous,
		Timestamps: o.timestamps,
		Tail:       o.tail,
		Since:      o.since,
	}
}

// runLogs parses logs's flags and hands them to actLogs.
func runLogs(args []string, r runner.Runner) int {
	fs := verbFlagSet("logs")
	env := envFlag(fs)
	platform := platformFlag(fs)
	var pods, containers []string
	fs.Var(repeatableName{&pods}, "pod", "read this kubernetes pod log, by name or by index")
	fs.Var(repeatableName{&containers}, "container", "read this docker/podman container log, by name or by index")
	namespace := fs.String("namespace", "", "kubernetes namespace to query")
	command := fs.String("command", "", "override the platform CLI binary used to reach each target")
	allow := allowCommandFlag(fs)
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
	// logs reads one instance, so naming two is a question it cannot answer.
	// Refused loudly rather than by quietly using the first: a repeated flag is
	// far more often a slip than an intention, and the picker is the supported
	// way to see what there is to choose from.
	for _, f := range []struct {
		vals []string
		name string
	}{
		{o.pods, podFlagName},
		{o.conts, containerFlagName},
	} {
		if len(f.vals) > 1 {
			fmt.Fprintf(os.Stderr, "logs: %s may be given once -- logs reads one instance; run logs with no %s to list what matches\n", f.name, f.name)
			return 2
		}
	}
	return 0
}

// actLogs resolves the platform, settles on exactly one instance, and reads its
// log. Everything that can fail the whole run -- an unreadable env.yaml, an
// unsafe name, a platform that cannot be resolved, a failed preflight -- fails
// here, before anything is read.
func actLogs(o logsOpts, r runner.Runner) int {
	// Both paths stream: a one-shot read wants the platform diagnostics on
	// stderr just as much as a followed one does, because `logs > app.log` is
	// the whole point of the verb and a redirect that swallowed them would be
	// the first thing to confuse an operator.
	s, ok := r.(runner.Streamer)
	if !ok {
		return errExit(errors.New("logs needs a runner that can stream command output, and this one cannot"))
	}

	sess, serr := resolveInstanceSession(instanceRequest{
		verb:     verbLogs,
		envPath:  o.envPath,
		platform: o.platform,
		ns:       o.ns,
		command:  o.command,
		pods:     o.pods,
		conts:    o.conts,
		allow:    o.allow,
	})
	if serr != nil {
		return errExit(serr)
	}

	if o.previous && sess.platform != validate.PlatformKubernetes {
		fmt.Fprintf(os.Stderr, "logs: %s is a kubernetes concept; %s keeps no prior-run log to read\n", previousFlagName, sess.platform)
		return 2
	}

	// The same read-only preflight probe status runs, for the same reason: fail
	// on an unreachable engine once, up front, rather than at the point of read.
	if code, ok := preflight(r, runner.ActionDeploy, sess.platform, sess.command, sess.ns, o.allow); !ok {
		return code
	}

	target, found, terr := resolveOneInstance(r, sess, namedInstance(sess.platform, o.pods, o.conts), logsInvocation(sess, o))
	if terr != nil {
		return errExit(terr)
	}
	if !found {
		// The picker was printed: the run answered an ambiguous request with
		// exactly what to type next, which is not a failure.
		return 0
	}
	return readLog(s, sess, o, target)
}

// logsInvocation rebuilds the command that was run, minus the instance, so the
// picker lines carry the flags the operator already typed instead of dropping
// them. The resolved --platform is always included: a pasted line must not
// re-open the interactive platform menu.
//
// Every flag logs has belongs in front of the instance, so the line needs no
// suffix -- that half exists for cli, whose command has to stay behind its "--".
func logsInvocation(sess instanceSession, o logsOpts) pickerLine {
	parts := []string{"solmq-conn-util", "logs", platformFlagName, sess.platform}
	if o.envPath != defaultEnvFile {
		parts = append(parts, "-e", o.envPath)
	}
	if o.ns != "" {
		parts = append(parts, namespaceFlagName, o.ns)
	}
	if o.command != "" {
		parts = append(parts, commandFlagName, o.command)
	}
	for _, a := range o.allow {
		parts = append(parts, allowCommandFlagName, a)
	}
	if o.follow {
		parts = append(parts, followFlagName)
	}
	if o.previous {
		parts = append(parts, previousFlagName)
	}
	if o.timestamps {
		parts = append(parts, timestampsFlagName)
	}
	if o.tail != runner.TailAll {
		parts = append(parts, tailFlagName, strconv.Itoa(o.tail))
	}
	if o.since != "" {
		parts = append(parts, sinceFlagName, o.since)
	}
	return pickerLine{prefix: strings.Join(parts, " ")}
}

// readLog reads the one resolved instance log, following it when asked.
//
// Follow mode is the same read with a cancellable context and a signal handler:
// Ctrl-C is how a follow is meant to end, so it exits 0 -- the same reasoning
// watchStatus records for --watch. Stream reports a cancelled context as
// success, so there is nothing here to unpick.
func readLog(s runner.Streamer, sess instanceSession, o logsOpts, t instanceRef) int {
	argv, err := runner.LogsArgv(sess.cmdArgv, sess.platform, t.name, t.namespace, o.forRunner())
	if err != nil {
		return errExit(err)
	}

	ctx := context.Background()
	if o.follow {
		var cancel context.CancelFunc
		ctx, cancel = context.WithCancel(ctx)
		defer cancel()
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, os.Interrupt)
		defer signal.Stop(stop)
		go func() {
			<-stop
			cancel()
		}()
	}

	if err := s.Stream(ctx, runner.Cmd{Argv: argv}, os.Stdout, os.Stderr); err != nil {
		return errExit(fmt.Errorf("reading %s: %w", t.name, err))
	}
	return 0
}
