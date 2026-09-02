package main

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strings"

	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/runner"
	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/validate"
)

// This file implements the cli verb: a shell inside one running instance, or a
// single command run in it.
//
// Every Go identifier here is spelled "shell" rather than "cli" because
// cliVerb/cliVerbs/cliFlags/cliTarget in commands.go already mean the command
// model, and a runCli sitting next to a cliVerb would read as part of it. Only
// the operator-facing word is "cli"; verbShell in instances.go is where the two
// spellings meet.

// argSeparator ends cli's own flags and begins the command to run inside the
// instance.
const argSeparator = "--"

// shellOpts is one parsed cli invocation, separated from how it was spelled the
// same way logsOpts is.
type shellOpts struct {
	envPath  string
	platform string
	pods     []string
	conts    []string
	ns       string
	command  string
	allow    []string

	// remote is what to run inside the instance: everything after the "--"
	// separator. Empty is the interactive form, an attached shell.
	remote []string
}

// interactive reports whether this invocation is the attached-shell form.
func (o shellOpts) interactive() bool { return len(o.remote) == 0 }

// splitAtSeparator divides args at the first bare "--", returning what precedes
// it, what follows it, and whether one was there at all.
//
// This has to happen before the flag package sees anything. collectFlagsAndDirs
// re-enters fs.Parse once per positional it finds, and Go's flag package treats
// "--" as the end of parsing -- so `cli -- ls -la` would hand back "ls" as a
// positional and then re-parse "-la" as an unknown flag of cli's own. Splitting
// first means the flag set never sees the separator or anything past it, and
// the in-container command keeps its own flags.
//
// sawSep is what distinguishes `cli --` (a separator with nothing after it,
// which is a mistake worth naming) from a plain `cli` (which is the shell).
func splitAtSeparator(args []string) (before, after []string, sawSep bool) {
	for i, a := range args {
		if a == argSeparator {
			return args[:i], args[i+1:], true
		}
	}
	return args, nil, false
}

// runShell parses cli's flags and hands them to actShell.
func runShell(args []string, r runner.Runner) int {
	flagArgs, remote, sawSep := splitAtSeparator(args)

	fs := verbFlagSet(verbShell)
	env := envFlag(fs)
	platform := platformFlag(fs)
	var pods, containers []string
	fs.Var(repeatableName{&pods}, "pod", "open the session in this kubernetes pod, by name or by index")
	fs.Var(repeatableName{&containers}, "container", "open the session in this docker/podman container, by name or by index")
	namespace := fs.String("namespace", "", "kubernetes namespace to query")
	command := fs.String("command", "", "override the platform CLI binary used to reach the instance")
	allow := allowCommandFlag(fs)

	pos, err := collectFlagsAndDirs(fs, flagArgs)
	if err != nil {
		return flagExit(verbShell, err)
	}
	if len(pos) > 0 {
		// cli has no target word: an instance is named with --pod/--container,
		// and a command to run in it goes after "--", so a bare word here is a
		// mistake worth naming rather than a name to guess at.
		fmt.Fprintf(os.Stderr, "cli: unexpected argument %q (name an instance with %s or %s; a command to run goes after %s)\n",
			pos[0], podFlagName, containerFlagName, argSeparator)
		return 2
	}
	if sawSep && len(remote) == 0 {
		fmt.Fprintf(os.Stderr, "cli: %s was given with no command after it (leave it off to open a shell)\n", argSeparator)
		return 2
	}

	o := shellOpts{
		envPath:  *env,
		platform: *platform,
		pods:     pods,
		conts:    containers,
		ns:       *namespace,
		command:  *command,
		allow:    *allow,
		remote:   remote,
	}
	if code := checkShellFlags(o); code != 0 {
		return code
	}
	return actShell(o, r)
}

// checkShellFlags rejects the combinations that cannot mean anything, before
// any query runs -- the same contract checkLogsFlags keeps.
func checkShellFlags(o shellOpts) int {
	// cli reaches one instance, so naming two is a question it cannot answer.
	// Refused loudly rather than by quietly using the first, exactly as logs
	// refuses it: a repeated flag is far more often a slip than an intention,
	// and the picker is the supported way to see what there is to choose from.
	for _, f := range []struct {
		vals []string
		name string
	}{
		{o.pods, podFlagName},
		{o.conts, containerFlagName},
	} {
		if len(f.vals) > 1 {
			fmt.Fprintf(os.Stderr, "cli: %s may be given once -- cli reaches one instance; run cli with no %s to list what matches\n", f.name, f.name)
			return 2
		}
	}
	return 0
}

// actShell resolves the platform, settles on exactly one instance, and hands
// the terminal to it. Everything that can fail the whole run -- a runner that
// cannot attach, an unsafe command token, a non-interactive stdin, an
// unreadable env.yaml, a platform that cannot be resolved, a failed preflight
// -- fails here, before any process starts.
func actShell(o shellOpts, r runner.Runner) int {
	a, ok := r.(runner.Attacher)
	if !ok {
		return errExit(errors.New("cli needs a runner that can hand a command the terminal, and this one cannot"))
	}

	// Every token of the in-container command is held to the same charset every
	// other operator-supplied argv token is. No shell is involved -- the tokens
	// reach the engine as separate argv entries, never through sh -c -- so a
	// metacharacter has nothing to expand it, which also means one in the
	// command was a mistake rather than an intention. Checked before anything
	// else so a typo costs no process at all.
	for _, t := range o.remote {
		if !validate.SafeToken(t) {
			return errExit(fmt.Errorf("the command token %q contains an unsafe character (%s); cli runs an argv, not a shell line, so a pipe or a glob has to be written inside the session instead", t, validate.UnsafeTokenReason))
		}
	}

	// The same non-TTY refusal the platform menu, the status install prompt and
	// the remove confirmation make, for the same reason: a run that cannot
	// possibly work should say so with a next step rather than hang or open a
	// session nobody can type into. Only the shell form needs a terminal; the
	// one-shot form is exactly what a script should be using instead.
	if o.interactive() && !stdinIsTerminal() {
		return errExit(fmt.Errorf("cli opens an interactive shell and stdin is not a terminal; run `cli %s <command>` to run one command instead", argSeparator))
	}

	sess, serr := resolveInstanceSession(instanceRequest{
		verb:     verbShell,
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

	// The same read-only preflight probe status and logs run, for the same
	// reason: fail on an unreachable engine once, up front, rather than at the
	// point of attach.
	if code, ok := preflight(r, runner.ActionDeploy, sess.platform, sess.command, sess.ns, o.allow); !ok {
		return code
	}

	target, found, terr := resolveOneInstance(r, sess, namedInstance(sess.platform, o.pods, o.conts), shellInvocation(sess, o))
	if terr != nil {
		return errExit(terr)
	}
	if !found {
		// The picker was printed: the run answered an ambiguous request with
		// exactly what to type next, which is not a failure.
		return 0
	}
	return attachShell(a, sess, o, target)
}

// shellInvocation rebuilds the command that was run, minus the instance, so the
// picker lines carry the flags the operator already typed instead of dropping
// them. The resolved --platform is always included: a pasted line must not
// re-open the interactive platform menu.
//
// The in-container command becomes the suffix rather than part of the prefix,
// because it has to stay behind its "--" and therefore behind the --pod the
// picker is telling the operator to add.
func shellInvocation(sess instanceSession, o shellOpts) pickerLine {
	parts := []string{"solmq-conn-util", verbShell, platformFlagName, sess.platform}
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
	line := pickerLine{prefix: strings.Join(parts, " ")}
	if !o.interactive() {
		line.suffix = argSeparator + " " + strings.Join(o.remote, " ")
	}
	return line
}

// attachShell hands the operator's own terminal to the engine and waits for the
// session to end.
//
// The status it returns is the child's own, not folded into this tool's 0/1/2:
// what a script wants back from `cli -- <command>` is the command's answer. It
// cannot be made more precise than that, which is why the verb's help says so
// outright -- kubectl exec reports "the pod could not be reached" and "the
// command exited 1" with the same status, so a non-zero cli exit means one of
// the two and the engine's own message on stderr is what distinguishes them.
func attachShell(a runner.Attacher, sess instanceSession, o shellOpts, t instanceRef) int {
	opts := runner.ExecOpts{Stdin: true, TTY: true}
	payload := []string{runner.ContainerShell}
	if !o.interactive() {
		// No terminal for a one-shot: there is no prompt to draw and nothing to
		// line-edit. Stdin is attached only when something is actually being
		// piped in, which is what a stdin that is not a terminal means here.
		// Asking for it unconditionally would leave a command that reads stdin
		// waiting on a terminal that never sends EOF.
		opts = runner.ExecOpts{Stdin: !stdinIsTerminal()}
		payload = o.remote
	}

	argv, err := runner.ExecArgv(sess.cmdArgv, sess.platform, t.name, t.namespace, opts)
	if err != nil {
		return errExit(err)
	}
	argv = append(argv, payload...)

	if o.interactive() {
		// A busybox prompt's escape sequences print as literal bytes in a plain
		// conhost window otherwise. Best-effort and silent, as at its other call
		// site.
		enableVirtualTerminal(os.Stdout)
	}

	restore := ignoreInterruptWhileAttached()
	defer restore()

	code, aerr := a.Attach(runner.Cmd{Argv: argv}, os.Stdin, os.Stdout, os.Stderr)
	if aerr != nil {
		return errExit(fmt.Errorf("opening a session on %s: %w", t.name, aerr))
	}
	return code
}

// ignoreInterruptWhileAttached stops this process reacting to Ctrl-C for the
// life of an attached session, and returns the function that restores it.
//
// Ctrl-C typed at an attached shell belongs to the shell, not to
// solmq-conn-util. Without this, Go's default handling kills the parent: on
// unix os/exec leaves the child in the parent's process group, so the terminal
// signals both; on Windows a console Ctrl-C raises CTRL_C_EVENT in every
// process attached to the console, which os/signal surfaces as the same
// os.Interrupt. Either way the child already has its own copy straight from the
// terminal and needs nothing forwarded to it -- the only thing to fix is the
// parent dying first, which would take the session down with it before the
// engine could restore the console mode it changed.
//
// signal.Notify rather than signal.Ignore, and the difference is load-bearing:
// signal.Ignore sets the disposition to SIG_IGN, and SIG_IGN is inherited
// across exec, so the shell in the container would come up unable to be
// interrupted at all. A Go handler is reset to the default by exec instead.
//
// No draining goroutine, unlike readLog's follow handler: that one notifies in
// order to act, this one notifies in order not to. Delivery to a notified
// channel is a non-blocking send, so a buffered channel nobody reads simply
// drops every signal, which is the whole intent. And no build tag, unlike
// enableVirtualTerminal: os/signal already maps CTRL_C_EVENT onto os.Interrupt,
// so one spelling serves both platforms.
//
// What it actually covers is narrower than it looks, and worth being honest
// about: once the engine has the TTY it puts the local terminal into raw mode,
// so Ctrl-C becomes a plain 0x03 byte on the child's stdin and the remote pty
// turns it into a signal for whatever is running inside the container -- which
// is the behaviour wanted. This handler covers the gap before raw mode is set,
// a run where the engine could not get a TTY at all, and keeping the parent
// alive long enough for the engine to put the console back as it found it.
func ignoreInterruptWhileAttached() func() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt)
	return func() { signal.Stop(sigs) }
}
