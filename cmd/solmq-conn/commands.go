package main

//go:generate go test . -run TestCommandsDocInSync -update

import "strings"

// This file is the single source of truth for the solmq-conn command tree. Both
// the in-binary help (usageText, printed by usage) and the generated reference
// docs/commands.md (renderCommandsDoc) are produced from cliVerbs/cliFlags, and
// TestCommandsDocInSync / TestCommandsModelMatchesUsage gate the two against it.

const bt = "`" // backtick, for building markdown code spans without raw-string clashes

// cliFlag documents one global flag.
type cliFlag struct {
	Short, Long, AppliesTo, Meaning string
}

// cliTarget documents a second-argument target/platform under a verb.
type cliTarget struct {
	Name, Summary, Example string
}

// cliVerb documents a top-level command. A verb either fans out into Targets
// (generate/deploy/delete) or is a leaf (validate/examples/help).
type cliVerb struct {
	Name, Args, Detail string
	Summary, Example   string // leaf verbs only (verbs with targets carry these per-target)
	TreeSuffix         string // leaf-verb suffix in the command-tree bullet
	Flags              []string
	InUsage            bool // appears as a command line in the -h/help summary
	Targets            []cliTarget
}

var cliFlags = []cliFlag{
	{Short: "-e", Long: "--env", AppliesTo: "all except " + bt + "examples" + bt, Meaning: "config file, relative or absolute path (default: " + bt + "env.yaml" + bt + ")"},
	{Short: "-o", Long: "--out", AppliesTo: bt + "generate" + bt, Meaning: "write output to a file (default: stdout)"},
	{Short: "-f", Long: "--force", AppliesTo: bt + "examples" + bt, Meaning: "overwrite existing files"},
}

var cliVerbs = []cliVerb{
	{
		Name: "generate", Args: "[-e env.yaml] [-o out]", Flags: []string{"-e", "-o"}, InUsage: true,
		Detail: "Renders the target's artifacts from " + bt + "env.yaml" + bt + " and prints them to stdout (or " + bt + "-o" + bt + "). Fails fast: stops at the first error and writes nothing; output is buffered, so a failed run never leaves a half-written " + bt + "-o" + bt + " file.",
		Targets: []cliTarget{
			{Name: "config", Summary: "Emit application.yml", Example: "solmq-conn generate config -e env.yaml -o application.yml"},
			{Name: "kubernetes", Summary: "Emit ConfigMap+Deployment+Service (+Secrets)", Example: "solmq-conn generate kubernetes -e env.yaml -o k8s.yaml"},
			{Name: "docker", Summary: "Emit docker-compose.yml (application.yml inlined)", Example: "solmq-conn generate docker -e env.yaml -o docker-compose.yml"},
			{Name: "podman", Summary: "Emit a podman run script or quadlet unit", Example: "solmq-conn generate podman -e env.yaml -o run.sh"},
		},
	},
	{
		Name: "deploy", Args: "[-e env.yaml]", Flags: []string{"-e"}, InUsage: true,
		Detail: "Generates for the platform, then applies it by shelling out to the section's " + bt + "command:" + bt + " (" + bt + "kubectl" + bt + "/" + bt + "oc" + bt + ", " + bt + "docker" + bt + ", or " + bt + "podman" + bt + " + " + bt + "systemctl" + bt + ") through an argv slice -- never a shell. The env file must contain the matching section.",
		Targets: []cliTarget{
			{Name: "kubernetes", Summary: "kubectl/oc apply -f - (manifest on stdin)", Example: "solmq-conn deploy kubernetes -e env.yaml"},
			{Name: "docker", Summary: "docker compose up -d", Example: "solmq-conn deploy docker -e env.yaml"},
			{Name: "podman", Summary: "write the quadlet unit; systemctl start", Example: "solmq-conn deploy podman -e env.yaml"},
		},
	},
	{
		Name: "delete", Args: "[-e env.yaml]", Flags: []string{"-e"}, InUsage: true,
		Detail: "Tears down what " + bt + "deploy" + bt + " created for the platform, the same way (via the section's " + bt + "command:" + bt + ").",
		Targets: []cliTarget{
			{Name: "kubernetes", Summary: "kubectl/oc delete -f -", Example: "solmq-conn delete kubernetes -e env.yaml"},
			{Name: "docker", Summary: "docker compose down", Example: "solmq-conn delete docker -e env.yaml"},
			{Name: "podman", Summary: "systemctl stop; remove the unit", Example: "solmq-conn delete podman -e env.yaml"},
		},
	},
	{
		Name: "validate", Args: "[-e env.yaml]", Flags: []string{"-e"}, InUsage: true,
		Summary: "Lint the whole env.yaml + workflows", Example: "solmq-conn validate -e env.yaml",
		Detail: "Runs every check across the whole " + bt + "env.yaml" + bt + " (including any " + bt + "kubernetes:" + bt + "/" + bt + "docker:" + bt + "/" + bt + "podman:" + bt + " sections) and its workflows, printing all findings. Non-zero exit if any errors. Use it as a linter.",
	},
	{
		Name: "examples", Args: "[dir] [-f]", Flags: []string{"-f"}, InUsage: true, TreeSuffix: bt + "[dir]" + bt,
		Summary: "Write a starter env.yaml + workflows", Example: "solmq-conn examples ./myconfig",
		Detail: "Writes a starter " + bt + "env.yaml" + bt + " plus workflow files into " + bt + "dir" + bt + " (default: the current directory). Use " + bt + "-f" + bt + " to overwrite existing files.",
	},
	{
		Name: "help", Args: "", Flags: nil, InUsage: false, TreeSuffix: "(" + bt + "-h" + bt + ", " + bt + "--help" + bt + ")",
		Summary: "Print the usage summary (also -h, --help)", Example: "solmq-conn help",
		Detail: "Prints the usage summary. Same as " + bt + "-h" + bt + " / " + bt + "--help" + bt + ".",
	},
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
				parts = append(parts, bt+f.Short+bt+", "+bt+f.Long+bt)
			}
		}
	}
	return "Flags: " + strings.Join(parts, "; ") + "."
}

// invocation builds "solmq-conn <verb> [target] [args]".
func invocation(v cliVerb, target string) string {
	s := "solmq-conn " + v.Name
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

	add("# solmq-conn command reference")
	add("")
	add("<!-- GENERATED -- do not edit by hand.")
	add("Source of truth: cmd/solmq-conn/commands.go (the cliSpec model).")
	add("Regenerate: go test ./cmd/solmq-conn -run TestCommandsDocInSync -update")
	add("TestCommandsDocInSync fails the build if this file drifts from the model. -->")
	add("")
	add("The full " + bt + "solmq-conn" + bt + " command tree. The first argument is a **verb**; where a verb")
	add("takes a second argument it names the **target** (" + bt + "generate" + bt + ") or **platform**")
	add("(" + bt + "deploy" + bt + "/" + bt + "delete" + bt + "). Generated from the command model in")
	add("[" + bt + "cmd/solmq-conn/commands.go" + bt + "](../cmd/solmq-conn/commands.go); see")
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
		add("| " + bt + f.Short + bt + ", " + bt + f.Long + bt + " | " + f.AppliesTo + " | " + f.Meaning + " |")
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
	return `solmq-conn -- Solace PubSub+ Connector for IBM MQ config generator + deployer

Usage:
  solmq-conn generate config     [-e env.yaml] [-o out]   Emit application.yml
  solmq-conn generate kubernetes [-e env.yaml] [-o out]   Emit ConfigMap+Deployment+Service (+Secrets)
  solmq-conn generate docker     [-e env.yaml] [-o out]   Emit docker-compose.yml (application.yml inlined)
  solmq-conn generate podman     [-e env.yaml] [-o out]   Emit a podman run script or quadlet unit
  solmq-conn deploy  kubernetes  [-e env.yaml]            kubectl/oc apply -f - (manifest on stdin)
  solmq-conn delete  kubernetes  [-e env.yaml]            kubectl/oc delete -f -
  solmq-conn deploy  docker      [-e env.yaml]            docker compose up -d
  solmq-conn delete  docker      [-e env.yaml]            docker compose down
  solmq-conn deploy  podman      [-e env.yaml]            write the quadlet unit; systemctl start
  solmq-conn delete  podman      [-e env.yaml]            systemctl stop; remove the unit
  solmq-conn validate            [-e env.yaml]            Lint the whole env.yaml + workflows
  solmq-conn examples [dir] [-f]                          Write a starter env.yaml + workflows

Flags:
  -e, --env     Config file, relative or absolute (default: env.yaml)
  -o, --out     Generate output file (default: stdout)
  -f, --force   examples: overwrite existing files

Workflows and per-target settings all come from env.yaml. The env file is always
excluded from the workflow set. Deploy commands run the CLI named by each
section's 'command:' via an argv slice -- never a shell -- and every command
token is checked against a safe charset.
`
}
