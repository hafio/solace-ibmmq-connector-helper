package main

//go:generate go test . -run "TestCommandsDocInSync|TestCompletionGoldenInSync" -update

import "strings"

// This file is the single source of truth for the solmq-conn-util command tree. The
// in-binary help (usageText, printed by usage), the generated reference
// docs/commands.md (renderCommandsDoc) and the shell completion scripts
// (completion.go) are all produced from cliVerbs/cliFlags, and
// TestCommandsDocInSync / TestCommandsModelMatchesUsage / TestCompletionGoldenInSync
// gate the three against it.

const bt = "`" // backtick, for building markdown code spans without raw-string clashes

// allowCommandFlagName is the repeatable deploy/delete flag's plain name (no
// backticks): the cliFlag key used in cliVerb.Flags, and the literal usageText()
// must contain.
const allowCommandFlagName = "--allow-command"

// allowCommandSpan and cmdFieldSpan are the markdown code spans for
// --allow-command and command:, reused across the deploy/delete Detail text.
const allowCommandSpan = bt + allowCommandFlagName + bt
const cmdFieldSpan = bt + "command:" + bt

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
type cliFlag struct {
	Short, Long, AppliesTo, Meaning string
	Arg                             string // value kind for completion: argNone/argFile/argName
}

// cliTarget documents a second-argument target/platform under a verb.
type cliTarget struct {
	Name, Summary, Example string
}

// cliVerb documents a top-level command. A verb either fans out into Targets
// (generate/deploy/delete/completion) or is a leaf (validate/examples/help).
type cliVerb struct {
	Name, Args, Detail string
	Summary, Example   string // leaf verbs only (verbs with targets carry these per-target)
	Blurb              string // completion-script description for a verb with targets (leaf verbs use Summary; read both through verbBlurb)
	TreeSuffix         string // leaf-verb suffix in the command-tree bullet
	PosArg             string // positional-argument kind for completion: posNone/posDir
	Flags              []string
	InUsage            bool // appears as a command line in the -h/help summary
	Targets            []cliTarget
}

var cliFlags = []cliFlag{
	{Short: "-e", Long: "--env", AppliesTo: "all except " + bt + "examples" + bt, Meaning: "config file, relative or absolute path (default: " + bt + "env.yaml" + bt + ")", Arg: argFile},
	{Short: "-o", Long: "--out", AppliesTo: bt + "generate" + bt, Meaning: "write output to a file (default: stdout)", Arg: argFile},
	{Short: "-f", Long: "--force", AppliesTo: bt + "examples" + bt, Meaning: "overwrite existing files", Arg: argNone},
	// argName, not argFile: the value must be a bare, PATH-resolved binary name
	// (allowCommandValue rejects a path), so offering filenames would suggest
	// exactly what the flag refuses.
	{Short: allowCommandFlagName, Long: allowCommandFlagName, AppliesTo: bt + "deploy" + bt + "/" + bt + "delete" + bt, Meaning: "approve an extra command binary beyond the " + cmdFieldSpan + " allowlist; repeatable", Arg: argName},
}

var cliVerbs = []cliVerb{
	{
		Name: "generate", Args: "[-e env.yaml] [-o out]", Flags: []string{"-e", "-o"}, InUsage: true,
		Blurb: "Render artifacts for one target to stdout or a file", PosArg: posNone,
		Detail: "Renders the target's artifacts from " + bt + "env.yaml" + bt + " and prints them to stdout (or " + bt + "-o" + bt + "). Fails fast: stops at the first error and writes nothing; output is buffered, so a failed run never leaves a half-written " + bt + "-o" + bt + " file.",
		Targets: []cliTarget{
			{Name: "config", Summary: "Emit application.yml", Example: "solmq-conn-util generate config -e env.yaml -o application.yml"},
			{Name: "kubernetes", Summary: "Emit ConfigMap+Deployment+Service (+Secrets)", Example: "solmq-conn-util generate kubernetes -e env.yaml -o k8s.yaml"},
			{Name: "docker", Summary: "Emit docker-compose.yml (application.yml inlined)", Example: "solmq-conn-util generate docker -e env.yaml -o docker-compose.yml"},
			{Name: "podman", Summary: "Emit a podman run script or quadlet unit", Example: "solmq-conn-util generate podman -e env.yaml -o run.sh"},
		},
	},
	{
		Name: "deploy", Args: "[-e env.yaml] [" + allowCommandFlagName + " name]", Flags: []string{"-e", allowCommandFlagName}, InUsage: true,
		Blurb: "Generate for a platform, then apply it", PosArg: posNone,
		Detail: "Generates for the platform, then applies it by shelling out to the section's " + cmdFieldSpan + " (" + bt + "kubectl" + bt + "/" + bt + "oc" + bt + ", " + bt + "docker" + bt + ", or " + bt + "podman" + bt + " + " + bt + "systemctl" + bt + ") through an argv slice -- never a shell. The env file must contain the matching section. " + cmdFieldSpan + "'s argv[0] must be a bare, allowlisted binary name (path-free, PATH-resolved); " + allowCommandSpan + " approves an extra binary for this invocation (e.g. a " + bt + "sudo" + bt + " prefix). Before anything is written or applied, a read-only preflight probe (login/permission check) must succeed, or the run stops with a login hint.",
		Targets: []cliTarget{
			{Name: "kubernetes", Summary: "kubectl/oc apply -f - (manifest on stdin)", Example: "solmq-conn-util deploy kubernetes -e env.yaml"},
			{Name: "docker", Summary: "docker compose up -d", Example: "solmq-conn-util deploy docker -e env.yaml"},
			{Name: "podman", Summary: "write the quadlet unit; systemctl start", Example: "solmq-conn-util deploy podman -e env.yaml"},
		},
	},
	{
		Name: "delete", Args: "[-e env.yaml] [" + allowCommandFlagName + " name]", Flags: []string{"-e", allowCommandFlagName}, InUsage: true,
		Blurb: "Tear down what deploy created for a platform", PosArg: posNone,
		Detail: "Tears down what " + bt + "deploy" + bt + " created for the platform, the same way (via the section's " + cmdFieldSpan + ", the same binary allowlist, " + allowCommandSpan + ", and the same read-only preflight probe before anything is torn down).",
		Targets: []cliTarget{
			{Name: "kubernetes", Summary: "kubectl/oc delete -f -", Example: "solmq-conn-util delete kubernetes -e env.yaml"},
			{Name: "docker", Summary: "docker compose down", Example: "solmq-conn-util delete docker -e env.yaml"},
			{Name: "podman", Summary: "systemctl stop; remove the unit", Example: "solmq-conn-util delete podman -e env.yaml"},
		},
	},
	{
		Name: "validate", Args: "[-e env.yaml]", Flags: []string{"-e"}, InUsage: true, PosArg: posNone,
		Summary: "Lint the whole env.yaml + workflows", Example: "solmq-conn-util validate -e env.yaml",
		Detail: "Runs every check across the whole " + bt + "env.yaml" + bt + " (including any " + bt + "kubernetes:" + bt + "/" + bt + "docker:" + bt + "/" + bt + "podman:" + bt + " sections) and its workflows, printing all findings. Non-zero exit if any errors. Use it as a linter.",
	},
	{
		Name: "examples", Args: "[dir] [-f]", Flags: []string{"-f"}, InUsage: true, TreeSuffix: bt + "[dir]" + bt, PosArg: posDir,
		Summary: "Write a starter env.yaml + workflows", Example: "solmq-conn-util examples ./myconfig",
		Detail: "Writes a starter " + bt + "env.yaml" + bt + " plus workflow files into " + bt + "dir" + bt + " (default: the current directory). Use " + bt + "-f" + bt + " to overwrite existing files.",
	},
	{
		// InUsage false: usageText() is hand-aligned and already carries twelve
		// command lines, so four more for a utility verb crowds the summary. It is
		// named in the usage footer prose instead, and listed in full in
		// docs/commands.md.
		Name: "completion", Args: "", Flags: nil, InUsage: false, PosArg: posNone,
		Blurb:  "Print a shell completion script",
		Detail: "Prints a completion script for the named shell on stdout, for you to source or drop into the shell's completion directory (see the per-shell examples below). The script is rendered from the same command model as this help, so the completion a binary prints always matches the commands that binary accepts.",
		Targets: []cliTarget{
			{Name: "bash", Summary: "Print the bash completion script", Example: "solmq-conn-util completion bash > /etc/bash_completion.d/solmq-conn-util"},
			{Name: "zsh", Summary: "Print the zsh completion script", Example: "solmq-conn-util completion zsh > ~/.zsh/completions/_solmq-conn-util"},
			{Name: "fish", Summary: "Print the fish completion script", Example: "solmq-conn-util completion fish > ~/.config/fish/completions/solmq-conn-util.fish"},
			{Name: "powershell", Summary: "Print the PowerShell completion script", Example: "solmq-conn-util completion powershell > solmq-conn-util-completion.ps1"},
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
	add("The full " + bt + "solmq-conn-util" + bt + " command tree. The first argument is a **verb**; where a verb")
	add("takes a second argument it names the **target** (" + bt + "generate" + bt + ") or **platform**")
	add("(" + bt + "deploy" + bt + "/" + bt + "delete" + bt + "). Generated from the command model in")
	add("[" + bt + "cmd/solmq-conn-util/commands.go" + bt + "](../cmd/solmq-conn-util/commands.go); see")
	add("[DEVELOPMENT.md](DEVELOPMENT.md#testing) to regenerate.")
	add("")
	add("## Command tree")
	add("")
	for _, v := range cliVerbs {
		suffix := ""
		if len(v.Targets) > 0 {
			names := make([]string, 0, len(v.Targets))
			for _, tg := range v.Targets {
				names = append(names, bt+tg.Name+bt)
			}
			suffix = " -> " + strings.Join(names, " | ")
		} else if v.TreeSuffix != "" {
			suffix = " " + v.TreeSuffix
		}
		add("- " + bt + v.Name + bt + suffix)
	}
	add("")
	add("## All commands")
	add("")
	add("| Command | Summary |")
	add("|---------|---------|")
	for _, v := range cliVerbs {
		if len(v.Targets) > 0 {
			for _, tg := range v.Targets {
				add("| " + bt + invocation(v, tg.Name) + bt + " | " + tg.Summary + " |")
			}
			continue
		}
		add("| " + bt + invocation(v, "") + bt + " | " + v.Summary + " |")
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
		add(v.Detail)
		add("")
		if fl := flagsLine(v); fl != "" {
			add(fl)
			add("")
		}
		if len(v.Targets) > 0 {
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
			continue
		}
		add("```sh")
		add(v.Example)
		add("```")
		add("")
	}

	return strings.TrimRight(strings.Join(l, "\n"), "\n") + "\n"
}

// usageText is the in-binary help (printed by usage). It stays hand-aligned for
// terminal readability; TestCommandsModelMatchesUsage checks it against cliVerbs.
func usageText() string {
	return `solmq-conn-util -- Solace PubSub+ Connector for IBM MQ config generator + deployer

Usage:
  solmq-conn-util generate config     [-e env.yaml] [-o out]   Emit application.yml
  solmq-conn-util generate kubernetes [-e env.yaml] [-o out]   Emit ConfigMap+Deployment+Service (+Secrets)
  solmq-conn-util generate docker     [-e env.yaml] [-o out]   Emit docker-compose.yml (application.yml inlined)
  solmq-conn-util generate podman     [-e env.yaml] [-o out]   Emit a podman run script or quadlet unit
  solmq-conn-util deploy  kubernetes  [-e env.yaml] [--allow-command name]  kubectl/oc apply -f - (manifest on stdin)
  solmq-conn-util delete  kubernetes  [-e env.yaml] [--allow-command name]  kubectl/oc delete -f -
  solmq-conn-util deploy  docker      [-e env.yaml] [--allow-command name]  docker compose up -d
  solmq-conn-util delete  docker      [-e env.yaml] [--allow-command name]  docker compose down
  solmq-conn-util deploy  podman      [-e env.yaml] [--allow-command name]  write the quadlet unit; systemctl start
  solmq-conn-util delete  podman      [-e env.yaml] [--allow-command name]  systemctl stop; remove the unit
  solmq-conn-util validate            [-e env.yaml]                        Lint the whole env.yaml + workflows
  solmq-conn-util examples [dir] [-f]                                      Write a starter env.yaml + workflows

Flags:
  -e, --env         Config file, relative or absolute (default: env.yaml)
  -o, --out         Generate output file (default: stdout)
  -f, --force       examples: overwrite existing files
  --allow-command   deploy/delete: approve an extra command binary beyond the
                     platform allowlist (repeatable)

Workflows and per-target settings all come from env.yaml. The env file is always
excluded from the workflow set. Deploy commands run the CLI named by each
section's 'command:' via an argv slice -- never a shell -- and every command
token is checked against a safe charset and a per-platform binary allowlist
(kubectl/oc, docker, podman); --allow-command approves an extra binary for one
invocation. Before anything is written or applied, a read-only preflight probe
checks login/permissions and stops with a hint on failure. The resolved binary
path is echoed before it runs.

Shell completion: 'solmq-conn-util completion bash|zsh|fish|powershell' prints a
completion script for that shell on stdout.
`
}
