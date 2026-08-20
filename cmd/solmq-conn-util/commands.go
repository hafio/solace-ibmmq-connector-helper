package main

//go:generate go test . -run "TestCommandsDocInSync|TestCompletionGoldenInSync" -update

import (
	"strings"

	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/spec"
)

// This file is the single source of truth for the solmq-conn-util command tree. The
// in-binary help (usageText, printed by usage), the generated reference
// docs/commands.md (renderCommandsDoc) and the shell completion scripts
// (completion.go) are all produced from cliVerbs/cliFlags, and
// TestCommandsDocInSync / TestCommandsModelMatchesUsage / TestCompletionGoldenInSync
// gate the three against it.

const bt = "`" // backtick, for building markdown code spans without raw-string clashes

// allowCommandFlagName is the repeatable deploy/remove/status flag's plain name
// (no backticks): the cliFlag key used in cliVerb.Flags, and the literal
// usageText() must contain.
const allowCommandFlagName = "--allow-command"

// allowCommandSpan and cmdFieldSpan are the markdown code spans for
// --allow-command and command:, reused across the deploy/remove/status Detail text.
const allowCommandSpan = bt + allowCommandFlagName + bt
const cmdFieldSpan = bt + "command:" + bt

// platformFlagName is generate/deploy/remove/status's platform selector (no
// short alias): the cliFlag key, and the literal usageText() must contain.
const platformFlagName = "--platform"

// platformArgBracket is the "[--platform kubernetes|docker|podman]" fragment,
// shared verbatim by every verb's Args/TreeSuffix that takes the flag so the
// enumerated platform list is spelled exactly once.
const platformArgBracket = "[" + platformFlagName + " kubernetes|docker|podman]"

// platformSpan is the markdown code span for --platform, reused across the
// generate/deploy/remove/status Detail text.
const platformSpan = bt + platformFlagName + bt

// installFlagName is status's --install flag (no short alias).
const installFlagName = "--install"

// status's remaining flags (no short alias). commandFlagName overrides the
// platform CLI binary for status; it is distinct from cmdFieldSpan, which is
// the env.yaml "command:" key these flags read a default from.
const (
	podFlagName            = "--pod"
	containerFlagName      = "--container"
	namespaceFlagName      = "--namespace"
	managementPortFlagName = "--management-port"
	userFlagName           = "--user"
	commandFlagName        = "--command"
)

// platformResolutionDetail explains how generate/deploy/remove/status pick a
// platform when --platform is not given. Rendered once as the doc's "Platform
// resolution" section; the four verbs' Detail text carries
// platformResolutionPointer instead of restating it, so the resolution order is
// described in exactly one place.
const platformResolutionDetail = "The platform is resolved in order: " + platformSpan + " (which accepts the short spellings " + bt + "kube" + bt + ", " + bt + "dk" + bt + " and " + bt + "pm" + bt + "), if given; otherwise the single " +
	bt + "kubernetes:" + bt + "/" + bt + "docker:" + bt + "/" + bt + "podman:" + bt +
	" section in env.yaml, when exactly one is present; otherwise an interactive menu, when more than one is present. A " +
	platformSpan + " value with no matching section in env.yaml is a loud error, and so are zero sections. The menu -- and, under " +
	bt + "status" + bt + ", the install confirmation prompt -- never block when stdin is not a TTY; both fail with the same guidance instead of hanging."

// platformResolutionPointer is the one-line cross-reference each platform verb's
// Detail ends with instead of restating platformResolutionDetail four times.
const platformResolutionPointer = "For how the platform is picked, see [Platform resolution](#platform-resolution)."

// Flag value kinds (cliFlag.Arg): what a flag's value completes to in the shell
// completion scripts. TestCompletionModelMetadataComplete rejects any other value,
// so a new flag cannot ship with a kind the renderers do not handle.
const (
	argNone = ""     // boolean flag; takes no value
	argFile = "file" // a filesystem path
	argName = "name" // a free-form name; no suggestions are offered
)

// Positional argument kinds (cliVerb.PosArg): what a verb's own positional
// argument completes to, beyond any target. Same gate as the flag kinds.
const (
	posNone = ""    // the verb takes no positional beyond its target, if any
	posDir  = "dir" // a directory
)

// cliFlag documents one global flag.
//
// Meaning, like every other description in this file (cliTarget.Summary,
// cliVerb.Blurb/Summary), must not contain an apostrophe: the fish, zsh and
// powershell emitters each escape one differently (quote-doubling idioms and
// backslashes), while TestCompletionCoversModel looks for the description
// verbatim in the rendered script, so an apostrophe fails that test and reads
// badly in the shell besides. Write "the namespace of the deployment" rather
// than "the deployment's namespace".
type cliFlag struct {
	Short, Long, AppliesTo, Meaning string
	Arg                             string // value kind for completion: argNone/argFile/argName
}

// cliTarget documents a second-argument target/platform under a verb.
type cliTarget struct {
	Name, Summary, Example string
}

// cliVerb documents a top-level command. A verb either fans out into Targets
// (auto-complete's shells; generate's optional "config") or has none (deploy,
// remove, status, version, validate, examples, help).
type cliVerb struct {
	Name, Args, Detail string
	Summary, Example   string // the verb's own invocation: every verb without Targets, plus generate's general (non-"config") form alongside its Targets
	Blurb              string // completion-script description for a verb with targets (leaf verbs use Summary; read both through verbBlurb)
	TreeSuffix         string // command-tree bullet suffix; takes priority over an auto-rendered Targets arrow (generate needs both a suffix and Targets)
	PosArg             string // positional-argument kind for completion: posNone/posDir
	Flags              []string
	InUsage            bool // appears as a command line in the -h/help summary
	Targets            []cliTarget
	// Aliases are alternate spellings for Name: recognized wherever the
	// canonical name is (dispatch, shell completion of a verb's own
	// flags/targets), but deliberately never offered as a position-1
	// completion candidate -- the TAB menu keeps showing only canonical verbs.
	// Entries must be [a-z0-9-] only: names are spliced unquoted into shell
	// case patterns and fish conditions, and assertShellSafeName enforces it.
	Aliases []string
}

var cliFlags = []cliFlag{
	{Short: "-e", Long: "--env", AppliesTo: "all except " + bt + "examples" + bt, Meaning: "config file, relative or absolute path (default: " + bt + "env.yaml" + bt + ")", Arg: argFile},
	{Short: "-o", Long: "--out", AppliesTo: bt + "generate" + bt, Meaning: "write output to a file (default: stdout)", Arg: argFile},
	{Short: "-f", Long: "--force", AppliesTo: bt + "examples" + bt, Meaning: "overwrite existing files", Arg: argNone},
	// argName, not argFile: the value must be a bare, PATH-resolved binary name
	// (allowCommandValue rejects a path), so offering filenames would suggest
	// exactly what the flag refuses.
	{Short: allowCommandFlagName, Long: allowCommandFlagName, AppliesTo: bt + "deploy" + bt + "/" + bt + "remove" + bt + "/" + bt + "status" + bt, Meaning: "approve an extra command binary beyond the " + cmdFieldSpan + " allowlist; repeatable", Arg: argName},
	// argName: the model has no enumerated-value kind (only argNone/argFile/
	// argName), so this value offers no shell suggestions even though it is one
	// of three fixed words. Adding a kind means teaching all four renderers.
	{Short: platformFlagName, Long: platformFlagName, AppliesTo: bt + "generate" + bt + "/" + bt + "deploy" + bt + "/" + bt + "remove" + bt + "/" + bt + "status" + bt, Meaning: "the platform: " + bt + "kubernetes" + bt + ", " + bt + "docker" + bt + ", or " + bt + "podman" + bt + " (short: " + bt + "kube" + bt + ", " + bt + "dk" + bt + ", " + bt + "pm" + bt + "; default: resolved from env.yaml, or an interactive menu -- see Platform resolution)", Arg: argName},
	{Short: installFlagName, Long: installFlagName, AppliesTo: bt + "status" + bt, Meaning: "install the status script on every instance without prompting", Arg: argNone},
	{Short: podFlagName, Long: podFlagName, AppliesTo: bt + "status" + bt, Meaning: "limit checks to this kubernetes pod name; repeatable (default: every running pod); no effect on docker/podman", Arg: argName},
	{Short: containerFlagName, Long: containerFlagName, AppliesTo: bt + "status" + bt, Meaning: "limit checks to this docker/podman container name; repeatable (default: every running container); no effect on kubernetes", Arg: argName},
	{Short: namespaceFlagName, Long: namespaceFlagName, AppliesTo: bt + "status" + bt, Meaning: "kubernetes namespace to query (default: the namespace of the deployment in env.yaml); no effect on docker/podman", Arg: argName},
	{Short: managementPortFlagName, Long: managementPortFlagName, AppliesTo: bt + "status" + bt, Meaning: "actuator management port to reach inside each instance (default: the configured management port)", Arg: argName},
	{Short: userFlagName, Long: userFlagName, AppliesTo: bt + "status" + bt, Meaning: "actuator account the status script authenticates as (default " + bt + spec.StatusUserName + bt + ")", Arg: argName},
	{Short: commandFlagName, Long: commandFlagName, AppliesTo: bt + "status" + bt, Meaning: "override the platform CLI binary (" + bt + "kubectl" + bt + "/" + bt + "oc" + bt + ", " + bt + "docker" + bt + ", or " + bt + "podman" + bt + ") used to reach each instance, instead of the " + cmdFieldSpan + " in that section", Arg: argName},
}

var cliVerbs = []cliVerb{
	{
		Name: "generate", Args: platformArgBracket + " [-e env.yaml] [-o out]",
		Flags: []string{platformFlagName, "-e", "-o"}, InUsage: true, PosArg: posNone, Aliases: []string{"gen"},
		Blurb:      "Render application.yml, or the artifacts for the resolved platform, to stdout or a file",
		Summary:    "Render the artifacts for the resolved platform to stdout or a file",
		Example:    "solmq-conn-util generate --platform kubernetes -e env.yaml -o k8s.yaml",
		TreeSuffix: bt + "[config]" + bt + " " + bt + platformArgBracket + bt,
		Detail:     "Renders artifacts and prints them to stdout (or " + bt + "-o" + bt + "). Fails fast: stops at the first error and writes nothing; output is buffered, so a failed run never leaves a half-written " + bt + "-o" + bt + " file. The " + bt + "config" + bt + " positional renders " + bt + "application.yml" + bt + " from env.yaml and never involves a platform (" + platformSpan + " is ignored); leaving it off renders the resolved platform's artifacts instead. " + platformResolutionPointer,
		Targets: []cliTarget{
			{Name: "config", Summary: "Emit application.yml", Example: "solmq-conn-util generate config -e env.yaml -o application.yml"},
		},
	},
	{
		Name: "deploy", Args: platformArgBracket + " [-e env.yaml] [" + allowCommandFlagName + " name]",
		Flags: []string{platformFlagName, "-e", allowCommandFlagName}, InUsage: true, PosArg: posNone, Aliases: []string{"dep"},
		Summary:    "Generate for a platform, then apply it",
		Example:    "solmq-conn-util deploy --platform kubernetes -e env.yaml",
		TreeSuffix: bt + platformArgBracket + bt,
		Detail:     "Generates for the platform, then applies it by shelling out to the section's " + cmdFieldSpan + " (" + bt + "kubectl" + bt + "/" + bt + "oc" + bt + ", " + bt + "docker" + bt + ", or " + bt + "podman" + bt + " + " + bt + "systemctl" + bt + ") through an argv slice -- never a shell. The env file must contain the matching section. " + cmdFieldSpan + "'s argv[0] must be a bare, allowlisted binary name (path-free, PATH-resolved); " + allowCommandSpan + " approves an extra binary for this invocation (e.g. a " + bt + "sudo" + bt + " prefix). Before anything is written or applied, a read-only preflight probe (login/permission check) must succeed, or the run stops with a login hint. " + platformResolutionPointer,
	},
	{
		Name: "remove", Args: platformArgBracket + " [-e env.yaml] [" + allowCommandFlagName + " name]",
		Flags: []string{platformFlagName, "-e", allowCommandFlagName}, InUsage: true, PosArg: posNone, Aliases: []string{"rm"},
		Summary:    "Tear down what deploy created for a platform",
		Example:    "solmq-conn-util remove --platform kubernetes -e env.yaml",
		TreeSuffix: bt + platformArgBracket + bt,
		Detail:     "Tears down what " + bt + "deploy" + bt + " created for the platform, the same way (via the section's " + cmdFieldSpan + ", the same binary allowlist, " + allowCommandSpan + ", and the same read-only preflight probe before anything is torn down). " + platformResolutionPointer,
	},
	{
		Name: "status", Args: "[" + installFlagName + "] " + platformArgBracket + " [-e env.yaml] [" + podFlagName + " name] [" + containerFlagName + " name] [" + namespaceFlagName + " ns] [" + managementPortFlagName + " port] [" + userFlagName + " name] [" + commandFlagName + " name] [" + allowCommandFlagName + " name]",
		Flags:   []string{installFlagName, platformFlagName, "-e", podFlagName, containerFlagName, namespaceFlagName, managementPortFlagName, userFlagName, commandFlagName, allowCommandFlagName},
		InUsage: true, PosArg: posNone, Aliases: []string{"sts"},
		Summary:    "Ensure and run the status script, printing per-instance leader-election and workflow state",
		Example:    "solmq-conn-util status --platform kubernetes -e env.yaml",
		TreeSuffix: bt + "[" + installFlagName + "]" + bt + " " + bt + platformArgBracket + bt,
		Detail:     "For the resolved platform, execs into each running instance (a " + bt + "kubernetes" + bt + " pod, or a " + bt + "docker" + bt + "/" + bt + "podman" + bt + " container) and ensures the status script is present: " + bt + installFlagName + bt + " installs it without asking; without it, a declined install prompt just skips the instances that are missing it. Then runs the script and prints, per instance, the leader-election result and workflow state from that instance's own actuator endpoint. " + bt + podFlagName + bt + " and " + bt + containerFlagName + bt + " (both repeatable) narrow which running instances are checked; " + bt + namespaceFlagName + bt + " overrides the kubernetes namespace and " + bt + managementPortFlagName + bt + " the actuator port; " + bt + userFlagName + bt + " names the read-only actuator account the script authenticates as, for an instance whose config does not carry the reserved " + bt + spec.StatusUserName + bt + " account. " + bt + commandFlagName + bt + " overrides the platform CLI binary used to reach each instance, and " + allowCommandSpan + " approves an extra one, the same as deploy/remove. " + platformResolutionPointer,
	},
	{
		Name: "version", Args: "", Flags: nil, InUsage: true, PosArg: posNone, Aliases: []string{"ver"},
		Summary: "Print the utility name, version, Go version and OS/arch",
		Example: "solmq-conn-util version",
		Detail:  "Prints solmq-conn-util's own version (stamped in at build time), the Go version it was built with, and its OS/arch (" + bt + "GOOS" + bt + "/" + bt + "GOARCH" + bt + ") -- for bug reports and to confirm which build is installed. Takes no flags.",
	},
	{
		Name: "validate", Args: "[-e env.yaml]", Flags: []string{"-e"}, InUsage: true, PosArg: posNone, Aliases: []string{"vld"},
		Summary: "Lint the whole env.yaml + workflows", Example: "solmq-conn-util validate -e env.yaml",
		Detail: "Runs every check across the whole " + bt + "env.yaml" + bt + " (including any " + bt + "kubernetes:" + bt + "/" + bt + "docker:" + bt + "/" + bt + "podman:" + bt + " sections) and its workflows, printing all findings. Non-zero exit if any errors. Use it as a linter.",
	},
	{
		Name: "examples", Args: "[dir] [-f]", Flags: []string{"-f"}, InUsage: true, TreeSuffix: bt + "[dir]" + bt, PosArg: posDir, Aliases: []string{"eg"},
		Summary: "Write a starter env.yaml + workflows", Example: "solmq-conn-util examples ./myconfig",
		Detail: "Writes a starter " + bt + "env.yaml" + bt + " plus workflow files into " + bt + "dir" + bt + " (default: the current directory). Use " + bt + "-f" + bt + " to overwrite existing files.",
	},
	{
		// InUsage false: usageText() is hand-aligned and already carries seven
		// command lines, so four more for a utility verb crowds the summary. It is
		// named in the usage footer prose instead, and listed in full in
		// docs/commands.md.
		Name: "auto-complete", Args: "", Flags: nil, InUsage: false, PosArg: posNone,
		Blurb:  "Print a shell completion script",
		Detail: "Prints a completion script for the named shell on stdout, for you to source or drop into the shell's completion directory (see the per-shell examples below). The script is rendered from the same command model as this help, so the completion a binary prints always matches the commands that binary accepts.",
		Targets: []cliTarget{
			{Name: "bash", Summary: "Print the bash completion script", Example: "solmq-conn-util auto-complete bash > /etc/bash_completion.d/solmq-conn-util"},
			{Name: "zsh", Summary: "Print the zsh completion script", Example: "solmq-conn-util auto-complete zsh > ~/.zsh/completions/_solmq-conn-util"},
			{Name: "fish", Summary: "Print the fish completion script", Example: "solmq-conn-util auto-complete fish > ~/.config/fish/completions/solmq-conn-util.fish"},
			{Name: "powershell", Summary: "Print the PowerShell completion script", Example: "solmq-conn-util auto-complete powershell > solmq-conn-util-completion.ps1"},
		},
	},
	{
		Name: "help", Args: "", Flags: nil, InUsage: false, TreeSuffix: "(" + bt + "-h" + bt + ", " + bt + "--help" + bt + ")", PosArg: posNone,
		Summary: "Print the usage summary (also -h, --help)", Example: "solmq-conn-util help",
		Detail: "Prints the usage summary. Same as " + bt + "-h" + bt + " / " + bt + "--help" + bt + ".",
	},
}

// verbBlurb is the one-line description the completion scripts show for a verb:
// Blurb where the verb fans out into targets (those carry no Summary of their
// own), Summary for a leaf. Kept as a lookup rather than a duplicated field so
// there is exactly one description string per verb to keep current.
func verbBlurb(v cliVerb) string {
	if v.Blurb != "" {
		return v.Blurb
	}
	return v.Summary
}

// flagSpan renders one flag as backtick-quoted markdown code spans: "`-e`,
// `--env`" for a short+long pair, or just "`--allow-command`" when the flag has
// no short alias (Short == Long).
func flagSpan(f cliFlag) string {
	if f.Short == f.Long {
		return bt + f.Long + bt
	}
	return bt + f.Short + bt + ", " + bt + f.Long + bt
}

// flagsLine renders the "Flags: ..." line for a verb's detail block ("" if none).
func flagsLine(v cliVerb) string {
	if len(v.Flags) == 0 {
		return ""
	}
	parts := make([]string, 0, len(v.Flags))
	for _, sh := range v.Flags {
		for _, f := range cliFlags {
			if f.Short == sh {
				parts = append(parts, flagSpan(f))
			}
		}
	}
	return "Flags: " + strings.Join(parts, "; ") + "."
}

// invocation builds "solmq-conn-util <verb> [target] [args]".
func invocation(v cliVerb, target string) string {
	s := "solmq-conn-util " + v.Name
	if target != "" {
		s += " " + target
	}
	if v.Args != "" {
		s += " " + v.Args
	}
	return s
}

// tableCell escapes a markdown table's column delimiter so a literal "|" in
// rendered text (--platform's enumerated kubernetes|docker|podman value) does
// not fracture the row into extra columns.
func tableCell(s string) string {
	return strings.ReplaceAll(s, "|", `\|`)
}

// renderCommandsDoc produces docs/commands.md from the model. Deterministic and
// byte-stable so TestCommandsDocInSync can gate the committed file against it.
func renderCommandsDoc() string {
	var l []string
	add := func(s string) { l = append(l, s) }

	add("# solmq-conn-util command reference")
	add("")
	add("<!-- GENERATED -- do not edit by hand.")
	add("Source of truth: cmd/solmq-conn-util/commands.go (the cliSpec model).")
	add("Regenerate: go test ./cmd/solmq-conn-util -run TestCommandsDocInSync -update")
	add("TestCommandsDocInSync fails the build if this file drifts from the model. -->")
	add("")
	add("The full " + bt + "solmq-conn-util" + bt + " command tree. The first argument is a **verb**.")
	add(bt + "generate" + bt + " takes an optional second argument, " + bt + "config" + bt + ", to render")
	add(bt + "application.yml" + bt + " instead of a platform's artifacts. " + bt + "generate" + bt + ", " + bt + "deploy" + bt + ",")
	add(bt + "remove" + bt + ", and " + bt + "status" + bt + " all accept " + bt + platformFlagName + bt + " to pick a **platform**")
	add("(" + bt + "kubernetes" + bt + ", " + bt + "docker" + bt + ", or " + bt + "podman" + bt + ") instead of resolving it from")
	add(bt + "env.yaml" + bt + ". Generated from the command model in")
	add("[" + bt + "cmd/solmq-conn-util/commands.go" + bt + "](../cmd/solmq-conn-util/commands.go); see")
	add("[DEVELOPMENT.md](DEVELOPMENT.md#testing) to regenerate.")
	add("")
	add("## Command tree")
	add("")
	for _, v := range cliVerbs {
		suffix := ""
		switch {
		case v.TreeSuffix != "":
			// TreeSuffix wins over an auto-rendered Targets arrow: generate needs
			// both (a completable "config" target and a hand-written flag sketch).
			suffix = " " + v.TreeSuffix
		case len(v.Targets) > 0:
			names := make([]string, 0, len(v.Targets))
			for _, tg := range v.Targets {
				names = append(names, bt+tg.Name+bt)
			}
			suffix = " -> " + strings.Join(names, " | ")
		}
		alias := ""
		if len(v.Aliases) > 0 {
			aliasSpans := make([]string, 0, len(v.Aliases))
			for _, a := range v.Aliases {
				aliasSpans = append(aliasSpans, bt+a+bt)
			}
			alias = " (" + strings.Join(aliasSpans, ", ") + ")"
		}
		add("- " + bt + v.Name + bt + alias + suffix)
	}
	add("")
	add("## All commands")
	add("")
	add("| Command | Summary |")
	add("|---------|---------|")
	for _, v := range cliVerbs {
		for _, tg := range v.Targets {
			add("| " + bt + tableCell(invocation(v, tg.Name)) + bt + " | " + tg.Summary + " |")
		}
		if v.Summary != "" {
			add("| " + bt + tableCell(invocation(v, "")) + bt + " | " + v.Summary + " |")
		}
	}
	add("")
	add("## Flags")
	add("")
	add("| Flag | Applies to | Meaning |")
	add("|------|-----------|---------|")
	for _, f := range cliFlags {
		add("| " + flagSpan(f) + " | " + f.AppliesTo + " | " + f.Meaning + " |")
	}
	add("")
	add("Flags may appear before, after, or between the positional arguments.")
	add("")
	add("## Platform resolution")
	add("")
	add(platformResolutionDetail)
	add("")
	add("## Exit codes")
	add("")
	add("| Code | Meaning |")
	add("|------|---------|")
	add("| " + bt + "0" + bt + " | success |")
	add("| " + bt + "1" + bt + " | processing error (bad input, unreadable file, missing env var, a deploy command that failed) |")
	add("| " + bt + "2" + bt + " | usage error (missing/unknown verb or target, unknown flag) |")
	add("")
	add("## Command details")
	add("")
	for _, v := range cliVerbs {
		add("### " + v.Name)
		add("")
		if len(v.Aliases) > 0 {
			aliasSpans := make([]string, 0, len(v.Aliases))
			for _, a := range v.Aliases {
				aliasSpans = append(aliasSpans, bt+a+bt)
			}
			add("Alias: " + strings.Join(aliasSpans, ", ") + ".")
			add("")
		}
		add(v.Detail)
		add("")
		if fl := flagsLine(v); fl != "" {
			add(fl)
			add("")
		}
		if v.Example != "" {
			add("```sh")
			add(v.Example)
			add("```")
			add("")
		}
		for _, tg := range v.Targets {
			add("#### " + bt + invocation(v, tg.Name) + bt)
			add("")
			add(tg.Summary + ".")
			add("")
			add("```sh")
			add(tg.Example)
			add("```")
			add("")
		}
	}

	return strings.TrimRight(strings.Join(l, "\n"), "\n") + "\n"
}

// usageText is the in-binary help (printed by usage). It stays hand-aligned for
// terminal readability; TestCommandsModelMatchesUsage checks it against cliVerbs.
func usageText() string {
	return `solmq-conn-util -- Solace PubSub+ Connector for IBM MQ config generator, deployer, and status checker

Usage:
  solmq-conn-util <generate|gen> config [-e env.yaml] [-o out]                                                      Emit application.yml
  solmq-conn-util <deploy|dep>          [--platform kubernetes|docker|podman] [-e env.yaml] [--allow-command name]  Generate for a platform, then apply it
  solmq-conn-util <remove|rm>           [--platform kubernetes|docker|podman] [-e env.yaml] [--allow-command name]  Tear down what deploy created for a platform
  solmq-conn-util <status|sts>          [--install] [--platform kubernetes|docker|podman] [-e env.yaml]             Ensure the status script is present and run it
  solmq-conn-util <version|ver>                                                                                     Print name, version, Go version, and OS/arch
  solmq-conn-util <validate|vld>        [-e env.yaml]                                                               Lint the whole env.yaml + workflows
  solmq-conn-util <examples|eg>         [dir] [-f]                                                                  Write a starter env.yaml + workflows

Flags:
  -e, --env            Config file, relative or absolute (default: env.yaml)
  -o, --out            Generate output file (default: stdout)
  -f, --force          examples: overwrite existing files
  --allow-command      deploy/remove/status: approve an extra command binary
                        beyond the platform allowlist (repeatable)
  --platform           generate/deploy/remove/status: kubernetes|docker|podman
                        (short: kube|dk|pm; default: resolved from env.yaml,
                        or a menu)
  --install            status: install the status script without prompting
  --pod                status: limit to this kubernetes pod (repeatable)
  --container          status: limit to this docker/podman container (repeatable)
  --namespace          status: kubernetes namespace to query
  --management-port    status: actuator port to reach inside each instance
  --user               status: actuator account the status script
                        authenticates as (default solmq-status)
  --command            status: override the platform CLI binary used to reach
                        each instance
`
}
