package main

//go:generate go test . -run "TestCommandsDocInSync|TestAbbreviationDocInSync|TestCompletionGoldenInSync" -update

import (
	"strings"

	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/libs"
	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/spec"
	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/statusreport"
)

// This file is the single source of truth for the solmq-conn-util command tree. The
// in-binary help (usageText, printed by usage), the generated reference
// docs/commands.md (renderCommandsDoc), the generated abbreviation lookup
// docs/abbreviation.md (renderAbbreviationDoc, abbreviation.go) and the shell
// completion scripts (completion.go) are all produced from cliVerbs/cliFlags, and
// TestCommandsDocInSync / TestCommandsModelMatchesUsage / TestAbbreviationDocInSync /
// TestCompletionGoldenInSync gate the four against it.

const bt = "`" // backtick, for building markdown code spans without raw-string clashes

// allowCommandFlagName is the repeatable deploy/remove/status/logs flag's plain
// name (no backticks): the cliFlag key used in cliVerb.Flags, and the literal
// usageText() must contain.
const allowCommandFlagName = "--allow-command"

// allowCommandSpan and cmdFieldSpan are the markdown code spans for
// --allow-command and command:, reused across the deploy/remove/status/logs
// Detail text.
const allowCommandSpan = bt + allowCommandFlagName + bt
const cmdFieldSpan = bt + "command:" + bt

// platformFlagName is generate/deploy/remove/status/logs's platform selector (no
// short alias): the cliFlag key, and the literal usageText() must contain.
const platformFlagName = "--platform"

// platformArgBracket is the "[--platform kubernetes|docker|podman]" fragment,
// shared verbatim by every verb's Args/TreeSuffix that takes the flag so the
// enumerated platform list is spelled exactly once.
const platformArgBracket = "[" + platformFlagName + " kubernetes|docker|podman]"

// platformSpan is the markdown code span for --platform, reused across the
// generate/deploy/remove/status/logs Detail text.
const platformSpan = bt + platformFlagName + bt

// installFlagName is status's --install flag (no short alias).
const installFlagName = "--install"

// noPromptFlagName is remove's confirmation skip (no short alias). It names
// what it does -- ask nothing -- which covers both questions a teardown asks:
// the confirmation, and whether to remove a namespace that turned out empty.
//
// Deliberately not -f/--force: cliFlags is keyed by Short across every verb, and
// -f already means "overwrite existing files" for examples/download. One key
// carrying two unrelated meanings would make that entry's own prose untrue.
const noPromptFlagName = "--no-prompt"

// noPromptSpan is the markdown code span for --no-prompt, for the remove Detail.
const noPromptSpan = bt + noPromptFlagName + bt

// urlFlagName is download's repeatable explicit-URL override (no short
// alias): when given, no Maven resolution happens at all.
const urlFlagName = "--url"

// versionFlagName pins the seed release (the IBM MQ client jar, or the
// syslog encoder jar) instead of resolving latest stable. An empty value
// means latest stable, exactly as omitting the flag does.
const versionFlagName = "--version"

// omitLibFileFlagName points at a jar list captured from a different
// connector image, replacing the embedded default the omission rule compares
// against. It REPLACES the embedded list completely rather than merging with
// it, so a named file containing nothing omits nothing.
const omitLibFileFlagName = "--omit-lib-file"

// includeProvidedFlagName downloads the whole closure even where the
// connector image already provides a jar, disabling omission entirely.
const includeProvidedFlagName = "--include-provided"

// versionSpan, omitLibFileSpan and includeProvidedSpan are the markdown code
// spans for the three flags above, reused across the download Detail text.
const versionSpan = bt + versionFlagName + bt
const omitLibFileSpan = bt + omitLibFileFlagName + bt
const includeProvidedSpan = bt + includeProvidedFlagName + bt

// status's remaining flags (no short alias). commandFlagName overrides the
// platform CLI binary for status; it is distinct from cmdFieldSpan, which is
// the env.yaml "command:" key these flags read a default from.
const (
	detailsFlagShort       = "-d"
	detailsFlagName        = "--details"
	watchFlagShort         = "-w"
	watchFlagName          = "--watch"
	allFlagName            = "--all"
	outputFlagName         = "--output"
	podFlagName            = "--pod"
	containerFlagName      = "--container"
	namespaceFlagName      = "--namespace"
	managementPortFlagName = "--management-port"
	userFlagName           = "--user"
	commandFlagName        = "--command"
)

// logs's own flags (no short aliases). followFlagName deliberately has no -f:
// cliFlags is keyed by Short across every verb, and -f is already --force.
const (
	followFlagName     = "--follow"
	previousFlagName   = "--previous"
	timestampsFlagName = "--timestamps"
	tailFlagName       = "--tail"
	sinceFlagName      = "--since"
)

// Flag spans for the logs Detail text, mirroring allowCommandSpan's shape.
var (
	followSpan     = bt + followFlagName + bt
	previousSpan   = bt + previousFlagName + bt
	timestampsSpan = bt + timestampsFlagName + bt
	tailSpan       = bt + tailFlagName + bt
	sinceSpan      = bt + sinceFlagName + bt
)

// statusTargetArgBracket is how status's required target word is written in
// every usage line and doc: the words themselves, since one of them must be
// typed. Short spellings are documented in the verb Detail rather than here, so
// the bracket stays readable at a glance.
//
// Built from the target-word constants rather than from the model with
// targetNames("status"): cliVerbs' own initialiser uses this value, so reading
// the model here would be an initialisation cycle. TestStatusTargetsMatchModel
// is what keeps the two in step instead.
const statusTargetArgBracket = "<" + statusTargetContainer + "|" + statusTargetApplication + "|" + statusTargetAll + ">"

// Flag spans for the status Detail text, mirroring allowCommandSpan's shape so
// each flag is spelled once in code.
var (
	detailsSpan = bt + detailsFlagShort + bt + "/" + bt + detailsFlagName + bt
	watchSpan   = bt + watchFlagShort + bt + "/" + bt + watchFlagName + bt
	allSpan     = bt + allFlagName + bt
	outputSpan  = bt + outputFlagName + bt
)

// platformResolutionDetail explains how generate/deploy/remove/status/logs pick
// a platform when --platform is not given. Rendered once as the doc's "Platform
// resolution" section; the five verbs' Detail text carries
// platformResolutionPointer instead of restating it, so the resolution order is
// described in exactly one place.
const platformResolutionDetail = "The platform is resolved in order: " + platformSpan + " (which accepts the short spellings " + bt + "kube" + bt + ", " + bt + "dk" + bt + " and " + bt + "pm" + bt + "), if given; otherwise the single " +
	bt + "kubernetes:" + bt + "/" + bt + "docker:" + bt + "/" + bt + "podman:" + bt +
	" section in env.yaml, when exactly one is present; otherwise an interactive menu, when more than one is present. A " +
	platformSpan + " value with no matching section in env.yaml is a loud error, and so are zero sections. " +
	bt + "status" + bt + ", " + bt + "logs" + bt + " and " + bt + "cli" + bt + " are the exception, and only when the operator has already named the instances themselves (" + bt + podFlagName + bt + "/" + bt + containerFlagName + bt + ", or " + bt + allFlagName + bt + " under " + bt + "status" + bt + ") alongside an explicit " + platformSpan +
	": there is then nothing left to read from env.yaml, so a missing file and a section-less platform are both fine -- which is how an instance this tool never deployed is reached. The menu -- and " +
	bt + "status" + bt + "'s install confirmation, and " + bt + "remove" + bt + "'s teardown confirmation -- never block when stdin is not a TTY; all three fail with the same guidance (naming the flag that skips them) instead of hanging."

// platformResolutionPointer is the one-line cross-reference each platform verb's
// Detail ends with instead of restating platformResolutionDetail five times.
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
	// Usage is the terse one-liner terminal help shows (`help <command>` and
	// `<command> -h`). Meaning above stays the full prose for docs/commands.md;
	// two texts on purpose, because a paragraph that reads well in a reference
	// table drowns a help page. Plain ASCII, no backticks, no trailing period.
	Usage string
}

// cliTarget documents a second-argument target/platform under a verb.
type cliTarget struct {
	Name, Summary, Example string
	// Aliases are alternate spellings for Name, recognised wherever the
	// canonical word is (dispatch, and a verb's own completion) but never
	// offered as a completion candidate, so the TAB menu keeps showing one
	// spelling per target. Same rule as cliVerb.Aliases, and the same charset
	// restriction: entries are spliced unquoted into shell case patterns and
	// fish conditions, so assertShellSafeName gates them too.
	Aliases []string
	// Sets is a third command level: the words a target itself fans out into
	// (download jar mq / download jar syslog). Only download/jar uses it today;
	// every other target leaves it nil.
	Sets []cliTarget
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
	// Synopsis is the short invocation the terminal help prints ("status
	// <container|application|all> [flags]") -- Args above stays the full
	// spelled-out form for docs/commands.md. No aliases: those are md-only.
	Synopsis string
	Targets  []cliTarget
	// Aliases are alternate spellings for Name: recognized wherever the
	// canonical name is (dispatch, shell completion of a verb's own
	// flags/targets), but deliberately never offered as a position-1
	// completion candidate -- the TAB menu keeps showing only canonical verbs.
	// Entries must be [a-z0-9-] only: names are spliced unquoted into shell
	// case patterns and fish conditions, and assertShellSafeName enforces it.
	Aliases []string
}

var cliFlags = []cliFlag{
	{Short: "-e", Long: "--env", AppliesTo: "all except " + bt + "examples" + bt + "/" + bt + "download" + bt, Meaning: "config file, relative or absolute path (default: " + bt + "env.yaml" + bt + ")", Arg: argFile, Usage: "config file (default: env.yaml)"},
	{Short: "-o", Long: "--out", AppliesTo: bt + "generate" + bt, Meaning: "write output to a file (default: stdout)", Arg: argFile, Usage: "write output to a file (default: stdout)"},
	{Short: "-f", Long: "--force", AppliesTo: bt + "examples" + bt + "/" + bt + "download" + bt, Meaning: "overwrite existing files", Arg: argNone, Usage: "overwrite existing files"},
	// argName, not argFile: the value must be a bare, PATH-resolved binary name
	// (allowCommandValue rejects a path), so offering filenames would suggest
	// exactly what the flag refuses.
	{Short: allowCommandFlagName, Long: allowCommandFlagName, AppliesTo: bt + "deploy" + bt + "/" + bt + "remove" + bt + "/" + bt + "status" + bt + "/" + bt + "logs" + bt + "/" + bt + "cli" + bt, Meaning: "approve an extra command binary beyond the " + cmdFieldSpan + " allowlist; repeatable", Arg: argName, Usage: "approve an extra command binary beyond the platform allowlist (repeatable)"},
	// argName: the model has no enumerated-value kind (only argNone/argFile/
	// argName), so this value offers no shell suggestions even though it is one
	// of three fixed words. Adding a kind means teaching all four renderers.
	{Short: platformFlagName, Long: platformFlagName, AppliesTo: bt + "generate" + bt + "/" + bt + "deploy" + bt + "/" + bt + "remove" + bt + "/" + bt + "status" + bt + "/" + bt + "logs" + bt + "/" + bt + "cli" + bt, Meaning: "the platform: " + bt + "kubernetes" + bt + ", " + bt + "docker" + bt + ", or " + bt + "podman" + bt + " (short: " + bt + "kube" + bt + ", " + bt + "dk" + bt + ", " + bt + "pm" + bt + "; default: resolved from env.yaml, or an interactive menu -- see Platform resolution)", Arg: argName, Usage: "kubernetes, docker, or podman (short: kube, dk, pm; default: from env.yaml, or a menu)"},
	// argName, not argFile: a URL is not a filesystem path, so offering file
	// suggestions would suggest exactly the wrong kind of value.
	{Short: urlFlagName, Long: urlFlagName, AppliesTo: bt + "download" + bt, Meaning: "exact URL to download instead of Maven resolution; repeatable; when given, no resolution happens at all", Arg: argName, Usage: "exact URL to download instead of Maven resolution (repeatable)"},
	{Short: versionFlagName, Long: versionFlagName, AppliesTo: bt + "download" + bt, Meaning: "pin the seed release (the IBM MQ client jar, or the syslog encoder jar) instead of resolving latest stable; empty means latest stable", Arg: argName, Usage: "pin the seed release (default: latest stable)"},
	{Short: omitLibFileFlagName, Long: omitLibFileFlagName, AppliesTo: bt + "download" + bt, Meaning: "a jar list that replaces (never merges with) the embedded default the omission rule compares against; an empty file omits nothing", Arg: argFile, Usage: "replace the built-in jar list the omission rule compares against"},
	{Short: includeProvidedFlagName, Long: includeProvidedFlagName, AppliesTo: bt + "download" + bt, Meaning: "download the whole closure even where the connector image already provides a jar, instead of omitting it", Arg: argNone, Usage: "download the whole closure even where the image already ships a jar"},
	{Short: installFlagName, Long: installFlagName, AppliesTo: bt + "status" + bt, Meaning: "install the status script on every instance without prompting", Arg: argNone, Usage: "install the status script on every instance without prompting"},
	{Short: noPromptFlagName, Long: noPromptFlagName, AppliesTo: bt + "remove" + bt, Meaning: "tear down without asking anything -- what a script or CI job passes, since the prompts refuse to read a non-TTY rather than hang. It covers both questions: the teardown confirmation, and whether to remove a namespace that turned out to be empty. It cannot authorise more than that: a namespace holding anything this release does not own is never removed, with or without it", Arg: argNone, Usage: "tear down without asking for confirmation"},
	{Short: detailsFlagShort, Long: detailsFlagName, AppliesTo: bt + "status" + bt, Meaning: "add the enrichment lines each view can report: worker node, CPU/memory use against allocation, image digest and referenced components; app version, java version, config path and heap", Arg: argNone, Usage: "add node, cpu/memory, digest, components; application version, java, config, heap"},
	{Short: watchFlagShort, Long: watchFlagName, AppliesTo: bt + "status" + bt, Meaning: "re-render the report every 5s until interrupted (Ctrl-C)", Arg: argNone, Usage: "re-render the report every 5s until interrupted"},
	{Short: allFlagName, Long: allFlagName, AppliesTo: bt + "status" + bt, Meaning: "reach every connector instance found by image name (" + bt + statusreport.ImageMatch + bt + ") instead of the ones " + bt + "env.yaml" + bt + " describes -- every namespace on kubernetes, every container on docker/podman; cannot be combined with " + bt + podFlagName + bt + "/" + bt + containerFlagName + bt, Arg: argNone, Usage: "reach every instance found by image name instead of the ones env.yaml describes"},
	{Short: outputFlagName, Long: outputFlagName, AppliesTo: bt + "status" + bt, Meaning: "output format: " + bt + "table" + bt + " (default) or " + bt + "json" + bt + ", one machine-readable document per run; " + bt + "json" + bt + " cannot be combined with " + bt + watchFlagName + bt, Arg: argName, Usage: "table (default) or json"},
	{Short: podFlagName, Long: podFlagName, AppliesTo: bt + "status" + bt + "/" + bt + "logs" + bt + "/" + bt + "cli" + bt, Meaning: "the kubernetes pod to reach, by name or by index into the listed order (alphabetical, the order " + bt + "status" + bt + " prints); a name always wins over the index reading. Repeatable on " + bt + "status" + bt + "; on " + bt + "logs" + bt + " and " + bt + "cli" + bt + ", which reach one instance, it may be given once. Default: every running pod on " + bt + "status" + bt + ", and on " + bt + "logs" + bt + "/" + bt + "cli" + bt + " the matching instances are listed instead. No effect on docker/podman", Arg: argName, Usage: "the kubernetes pod, by name or index (repeatable on status, once on logs)"},
	{Short: containerFlagName, Long: containerFlagName, AppliesTo: bt + "status" + bt + "/" + bt + "logs" + bt + "/" + bt + "cli" + bt, Meaning: "the docker/podman container to reach, by name or by index into the listed order (alphabetical, the order " + bt + "status" + bt + " prints); a name always wins over the index reading. Repeatable on " + bt + "status" + bt + "; on " + bt + "logs" + bt + " and " + bt + "cli" + bt + ", which reach one instance, it may be given once. Default: every running container on " + bt + "status" + bt + ", and on " + bt + "logs" + bt + "/" + bt + "cli" + bt + " the one the section in env.yaml names. No effect on kubernetes", Arg: argName, Usage: "the docker/podman container, by name or index (repeatable on status, once on logs)"},
	{Short: namespaceFlagName, Long: namespaceFlagName, AppliesTo: bt + "status" + bt + "/" + bt + "logs" + bt + "/" + bt + "cli" + bt, Meaning: "kubernetes namespace to query (default: the namespace of the deployment in env.yaml); no effect on docker/podman", Arg: argName, Usage: "kubernetes namespace to query"},
	{Short: managementPortFlagName, Long: managementPortFlagName, AppliesTo: bt + "status" + bt, Meaning: "actuator management port to reach inside each instance (default: the configured management port)", Arg: argName, Usage: "actuator management port inside each instance"},
	{Short: userFlagName, Long: userFlagName, AppliesTo: bt + "status" + bt, Meaning: "actuator account the status script authenticates as (default " + bt + spec.StatusUserName + bt + ")", Arg: argName, Usage: "actuator account the status script authenticates as (default solmq-status)"},
	{Short: commandFlagName, Long: commandFlagName, AppliesTo: bt + "status" + bt + "/" + bt + "logs" + bt + "/" + bt + "cli" + bt, Meaning: "override the platform CLI binary (" + bt + "kubectl" + bt + "/" + bt + "oc" + bt + ", " + bt + "docker" + bt + ", or " + bt + "podman" + bt + ") used to reach each instance, instead of the " + cmdFieldSpan + " in that section", Arg: argName, Usage: "override the platform CLI binary used to reach each instance"},
	{Short: followFlagName, Long: followFlagName, AppliesTo: bt + "logs" + bt, Meaning: "keep the log open and print new lines as they arrive, until interrupted (Ctrl-C); reads one instance, so it cannot be combined with " + bt + allFlagName + bt + " or " + bt + previousFlagName + bt, Arg: argNone, Usage: "keep the log open and print new lines until interrupted"},
	{Short: previousFlagName, Long: previousFlagName, AppliesTo: bt + "logs" + bt, Meaning: "read the log of the previous container instead of the running one -- what a pod that is restarting printed before it died; kubernetes only, since neither docker nor podman keeps a prior run under the same name", Arg: argNone, Usage: "read the log of the previous container instead of the running one (kubernetes only)"},
	{Short: tailFlagName, Long: tailFlagName, AppliesTo: bt + "logs" + bt, Meaning: "read only the last N lines, or " + bt + "all" + bt + " for the whole log (default: " + bt + "all" + bt + ")", Arg: argName, Usage: "read only the last N lines, or all (default: all)"},
	{Short: sinceFlagName, Long: sinceFlagName, AppliesTo: bt + "logs" + bt, Meaning: "read only lines newer than this duration, spelled as a Go duration (" + bt + "30s" + bt + ", " + bt + "10m" + bt + ", " + bt + "2h" + bt + ")", Arg: argName, Usage: "read only lines newer than this duration, e.g. 10m"},
	{Short: timestampsFlagName, Long: timestampsFlagName, AppliesTo: bt + "logs" + bt, Meaning: "prefix every line with the time the platform recorded for it", Arg: argNone, Usage: "prefix every line with the time the platform recorded for it"},
}

var cliVerbs = []cliVerb{
	{
		Name: "generate", Args: platformArgBracket + " [-e env.yaml] [-o out]",
		Flags: []string{platformFlagName, "-e", "-o"}, PosArg: posNone, Aliases: []string{"gen"},
		Synopsis:   "generate [config] [flags]",
		Blurb:      "Render application.yml, or the deploy artifacts for the resolved platform",
		Summary:    "Render the artifacts for the resolved platform to stdout or a file",
		Example:    "solmq-conn-util generate --platform kubernetes -e env.yaml -o k8s.yaml",
		TreeSuffix: bt + "[config]" + bt + " " + bt + platformArgBracket + bt,
		Detail:     "Renders artifacts and prints them to stdout (or " + bt + "-o" + bt + "). Fails fast: stops at the first error and writes nothing; output is buffered, so a failed run never leaves a half-written " + bt + "-o" + bt + " file. The " + bt + "config" + bt + " positional renders " + bt + "application.yml" + bt + " from env.yaml and never involves a platform (" + platformSpan + " is ignored); leaving it off renders the resolved platform's artifacts instead. " + platformResolutionPointer,
		Targets: []cliTarget{
			{Name: "config", Aliases: []string{"cfg"}, Summary: "Emit application.yml", Example: "solmq-conn-util generate config -e env.yaml -o application.yml"},
		},
	},
	{
		Name: "deploy", Args: platformArgBracket + " [-e env.yaml] [" + allowCommandFlagName + " name]",
		Flags: []string{platformFlagName, "-e", allowCommandFlagName}, PosArg: posNone, Aliases: []string{"dp"},
		Synopsis:   "deploy [flags]",
		Summary:    "Generate for a platform, then apply it",
		Example:    "solmq-conn-util deploy --platform kubernetes -e env.yaml",
		TreeSuffix: bt + platformArgBracket + bt,
		Detail:     "Generates for the platform, then applies it by shelling out to the section's " + cmdFieldSpan + " (" + bt + "kubectl" + bt + "/" + bt + "oc" + bt + ", " + bt + "docker" + bt + ", or " + bt + "podman" + bt + " + " + bt + "systemctl" + bt + ") through an argv slice -- never a shell. The env file must contain the matching section. " + cmdFieldSpan + "'s argv[0] must be a bare, allowlisted binary name (path-free, PATH-resolved); " + allowCommandSpan + " approves an extra binary for this invocation (e.g. a " + bt + "sudo" + bt + " prefix). Before anything is written or applied, a read-only preflight probe (login/permission check) must succeed, or the run stops with a login hint. " + platformResolutionPointer,
	},
	{
		Name: "remove", Args: platformArgBracket + " [" + noPromptFlagName + "] [-e env.yaml] [" + allowCommandFlagName + " name]",
		Flags: []string{noPromptFlagName, platformFlagName, "-e", allowCommandFlagName}, PosArg: posNone, Aliases: []string{"rm"},
		Synopsis:   "remove [flags]",
		Summary:    "Tear down what deploy created for a platform",
		Example:    "solmq-conn-util remove --platform kubernetes -e env.yaml",
		TreeSuffix: bt + platformArgBracket + bt + " " + bt + "[" + noPromptFlagName + "]" + bt,
		Detail: "Tears down what " + bt + "deploy" + bt + " created for the platform, the same way (via the section's " + cmdFieldSpan + ", the same binary allowlist, " + allowCommandSpan + ", and the same read-only preflight probe before anything is torn down). " +
			"Because it is the one verb that destroys rather than creates, it asks for confirmation first, naming what it is about to tear down -- the deployment and its namespace on kubernetes, the container on docker/podman -- so a run pointed at the wrong " + bt + "env.yaml" + bt + " is caught before anything is deleted. " +
			"Answering anything but " + bt + "y" + bt + "/" + bt + "yes" + bt + " cancels and exits 0, having touched nothing at all: the question comes before even the preflight probe. " +
			noPromptSpan + " skips the prompt, which is what a script or CI job passes -- without it a non-TTY run fails fast with that hint rather than hanging on a read that will never return. " +
			"On kubernetes the namespace is handled separately, and never as part of the manifest delete: deleting a Namespace cascades to every object inside it, including workloads this tool never deployed. " +
			"So once the teardown succeeds, " + bt + "remove" + bt + " looks at what is left in the namespace. Anything this release does not own -- another deployment, a stateful set, a volume claim, a secret -- is listed and the namespace is kept. Only a namespace with nothing else in it is offered for removal, as its own separate question. " +
			"That is an invariant rather than a default: no flag removes a namespace that still holds someone else's work. The cluster namespaces (" + bt + "default" + bt + ", " + bt + "kube-system" + bt + ", " + bt + "kube-public" + bt + ", " + bt + "kube-node-lease" + bt + ") are never removed either, and a check that cannot run leaves the namespace alone rather than assuming it is empty. " + platformResolutionPointer,
	},
	{
		Name: "status", Args: statusTargetArgBracket + " [" + detailsFlagName + "] [" + watchFlagName + "] [" + allFlagName + "] [" + outputFlagName + " table|json] [" + installFlagName + "] " + platformArgBracket + " [-e env.yaml] [" + podFlagName + " name] [" + containerFlagName + " name] [" + namespaceFlagName + " ns] [" + managementPortFlagName + " port] [" + userFlagName + " name] [" + commandFlagName + " name] [" + allowCommandFlagName + " name]",
		Flags:  []string{detailsFlagShort, watchFlagShort, allFlagName, outputFlagName, installFlagName, platformFlagName, "-e", podFlagName, containerFlagName, namespaceFlagName, managementPortFlagName, userFlagName, commandFlagName, allowCommandFlagName},
		PosArg: posNone, Aliases: []string{"sts"},
		Synopsis:   "status <container|application|all> [flags]",
		Blurb:      "Report each instance: container (engine), application (connector), or all",
		TreeSuffix: bt + statusTargetArgBracket + bt + " " + bt + "[" + detailsFlagName + "]" + bt + " " + bt + platformArgBracket + bt,
		Detail: "Reports the state of every connector instance of the resolved platform. The target word picks which half is reported, because they answer different questions and come from different places: " +
			bt + "container" + bt + " is what the container engine knows, read from outside through read-only " + bt + "kubectl" + bt + "/" + bt + "docker" + bt + "/" + bt + "podman" + bt + " queries -- state, restarts, age and the image actually running, in one table per platform; " +
			bt + "application" + bt + " is what the connector knows about itself, read from inside by running the generated status script in each instance -- leader-election mode and state, health, and per-workflow state; " +
			bt + "all" + bt + " reports both, container first, since a container that is not running is the reason the application half is missing. " +
			"Each word has a short spelling (" + bt + "cnt" + bt + ", " + bt + "app" + bt + "); the word itself is required, and " + bt + "status" + bt + " on its own prints this list. " +
			detailsSpan + " adds the enrichment lines to whichever view is being printed: worker node, CPU and memory use against allocation, the image digest, and the objects the workload references (secrets, config maps, volume claims, mounts) on the container side; app version, java version, the configuration file the report was read from, and JVM heap use on the application side. The one sampling query it needs (" + bt + "kubectl top" + bt + ", " + bt + "docker stats" + bt + ") is why those lines are opt-in: on kubernetes it also needs a metrics API in the cluster, and reports a note instead of the lines when there is none. " +
			watchSpan + " re-renders the report every 5s until interrupted. " + outputSpan + " " + bt + "json" + bt + " emits one machine-readable document per run instead of the tables, carrying every fact either view collected. " +
			allSpan + " ignores the instance names in " + bt + "env.yaml" + bt + " and reports every connector instance it can find by image name (" + bt + statusreport.ImageMatch + bt + ") -- across every namespace on kubernetes, and every container, running or not, on docker/podman. " +
			"For the application views, " + bt + installFlagName + bt + " installs the status script without asking where it is missing; without it, a declined install prompt just skips the instances that lack it. " +
			bt + podFlagName + bt + " and " + bt + containerFlagName + bt + " (both repeatable) narrow which instances are reported; " + bt + namespaceFlagName + bt + " overrides the kubernetes namespace and " + bt + managementPortFlagName + bt + " the actuator port; " + bt + userFlagName + bt + " names the read-only actuator account the script authenticates as, for an instance whose config does not carry the reserved " + bt + spec.StatusUserName + bt + " account. " + bt + commandFlagName + bt + " overrides the platform CLI binary used to reach each instance, and " + allowCommandSpan + " approves an extra one, the same as deploy/remove. " + platformResolutionPointer,
		Targets: []cliTarget{
			{
				Name: statusTargetContainer, Aliases: []string{"cnt"},
				Summary: "Report what the engine knows: state, restarts, age and image per instance",
				Example: "solmq-conn-util status container --platform kubernetes -e env.yaml",
			},
			{
				Name: statusTargetApplication, Aliases: []string{"app"},
				Summary: "Report what the connector knows: leader-election state, health and workflows",
				Example: "solmq-conn-util status application -e env.yaml",
			},
			{
				Name:    statusTargetAll,
				Summary: "Report both halves: the container table, then the application block per instance",
				Example: "solmq-conn-util status all -d",
			},
		},
	},
	{
		Name: "logs", Args: "[" + followFlagName + "] [" + previousFlagName + "] [" + tailFlagName + " N] [" + sinceFlagName + " d] [" + timestampsFlagName + "] " + platformArgBracket + " [-e env.yaml] [" + podFlagName + " name|index] [" + containerFlagName + " name|index] [" + namespaceFlagName + " ns] [" + commandFlagName + " name] [" + allowCommandFlagName + " name]",
		Flags:  []string{followFlagName, previousFlagName, tailFlagName, sinceFlagName, timestampsFlagName, platformFlagName, "-e", podFlagName, containerFlagName, namespaceFlagName, commandFlagName, allowCommandFlagName},
		PosArg: posNone, Aliases: []string{"lg"},
		Synopsis:   "logs [flags]",
		Summary:    "Print one instance log, where status says what but not why",
		Example:    "solmq-conn-util logs " + tailFlagName + " 100 -e env.yaml",
		TreeSuffix: bt + "[" + followFlagName + "]" + bt + " " + bt + "[" + previousFlagName + "]" + bt + " " + bt + platformArgBracket + bt,
		Detail: "Prints what one connector instance of the resolved platform has written, read through the same read-only " + bt + "kubectl" + bt + "/" + bt + "docker" + bt + "/" + bt + "podman" + bt + " path " + bt + "status" + bt + " uses -- the same " + cmdFieldSpan + ", the same binary allowlist, " + allowCommandSpan + ", and the same preflight probe -- and discovering the same instances from " + bt + "env.yaml" + bt + ", so the two verbs can never disagree about which ones they mean. " +
			"It is the answer to the question " + bt + "status" + bt + " leaves open: that view reports a restart count and an exit code, and this one reports the lines that preceded them. " +
			"One instance is read per run. When discovery finds several and none was named, nothing is read: the matching instances are listed on stdout as commands that can be pasted back verbatim, carrying the flags already typed, and the run exits 0. " +
			bt + podFlagName + bt + " and " + bt + containerFlagName + bt + " name the instance, and each may be given once -- a second is refused rather than silently losing the first. " +
			"Either accepts an **index** as well as a name: " + bt + podFlagName + " 0" + bt + " is the first instance in the listed order, which is alphabetical by name and the same order " + bt + "status" + bt + " prints. A name always wins, so an instance genuinely called " + bt + "0" + bt + " is still reachable by name. " +
			previousSpan + " is the pairing that matters most -- it reads what the previous container printed before it died, which is the only place a crash loop explains itself; kubernetes keeps that log, docker and podman do not, so it is refused there rather than quietly ignored. " +
			followSpan + " keeps the log open and prints new lines until interrupted (Ctrl-C, which is a clean exit, not a failure). " +
			tailSpan + " limits how far back the read starts (" + bt + tailFlagName + " all" + bt + " is the default), " + sinceSpan + " limits it by age, and " + timestampsSpan + " prefixes each line with the time the platform recorded. " +
			bt + namespaceFlagName + bt + " overrides the kubernetes namespace, " + bt + commandFlagName + bt + " overrides the platform CLI binary, and " + allowCommandSpan + " approves an extra one, the same as status. " +
			"The log itself goes to stdout and everything else to stderr, so " + bt + "logs > app.log" + bt + " captures the log and nothing but. " + platformResolutionPointer,
	},
	{
		Name: "cli", Args: platformArgBracket + " [-e env.yaml] [" + podFlagName + " name|index] [" + containerFlagName + " name|index] [" + namespaceFlagName + " ns] [" + commandFlagName + " name] [" + allowCommandFlagName + " name] [-- command ...]",
		Flags:      []string{platformFlagName, "-e", podFlagName, containerFlagName, namespaceFlagName, commandFlagName, allowCommandFlagName},
		PosArg:     posNone,
		Synopsis:   "cli [flags] [-- command ...]",
		Summary:    "Open a shell inside one instance, or run one command in it",
		Example:    "solmq-conn-util cli -e env.yaml",
		TreeSuffix: bt + "[-- command ...]" + bt + " " + bt + platformArgBracket + bt,
		Detail: "Opens an interactive shell inside one connector instance of the resolved platform, reached through the same read-only " + bt + "kubectl" + bt + "/" + bt + "docker" + bt + "/" + bt + "podman" + bt + " path " + bt + "status" + bt + " and " + bt + "logs" + bt + " use -- the same " + cmdFieldSpan + ", the same binary allowlist, " + allowCommandSpan + ", the same preflight probe, and the same instance discovery from " + bt + "env.yaml" + bt + ", so the three verbs can never disagree about which instance they mean. " +
			"It is where the questions " + bt + "status" + bt + " and " + bt + "logs" + bt + " cannot answer get settled: whether the truststore really mounted, what is actually in " + bt + "/app/external/libs" + bt + ", which " + bt + "application.yml" + bt + " the process is running. " +
			"Everything after " + bt + "--" + bt + " is run in the instance instead of a shell, non-interactively, which is the form for a script -- an interactive run with no terminal on stdin is refused with that as the next step rather than opening a session nobody can type into. " +
			"The shell is " + bt + "sh" + bt + ": the connector image is Alpine, so its userland is busybox and there is no " + bt + "bash" + bt + " to ask for. " +
			"One instance is reached per run. When discovery finds several and none was named, nothing is opened: the matching instances are listed on stdout as commands that can be pasted back verbatim, carrying the flags already typed, and the run exits 0. " +
			bt + podFlagName + bt + " and " + bt + containerFlagName + bt + " name the instance, each may be given once, and either accepts an **index** as well as a name -- " + bt + podFlagName + " 0" + bt + " is the first instance in the listed order, alphabetical by name and the same order " + bt + "status" + bt + " prints. A name always wins. " +
			"On kubernetes the container is always named explicitly (" + bt + "-c " + spec.ConnectorContainerName + bt + "), so a pod carrying a sidecar cannot be entered by mistake; a pod with no such container is refused by the platform rather than guessed at. " +
			"Every token of a " + bt + "--" + bt + " command is held to the same safe charset as every other value that reaches an argv, because " + bt + "cli" + bt + " runs an argv and never a shell line: a pipe, a redirect or a glob has to be written inside the session instead. " +
			"**Exit status** is the one place " + bt + "cli" + bt + " leaves the usual contract: once the session or command is running, whatever it exited with is passed straight back. The engines give no way to be more precise -- " + bt + "kubectl exec" + bt + " reports an unreachable pod and a command that exited non-zero with the same status -- so a non-zero " + bt + "cli" + bt + " exit means one of the two, and the message the engine printed on stderr is what says which. " + platformResolutionPointer,
	},
	{
		Name: "version", Args: "", Flags: nil, PosArg: posNone, Aliases: []string{"ver"},
		Synopsis: "version",
		Summary:  "Print the utility name, version, Go version and OS/arch",
		Example:  "solmq-conn-util version",
		Detail:   "Prints solmq-conn-util's own version (stamped in at build time), the Go version it was built with, and its OS/arch (" + bt + "GOOS" + bt + "/" + bt + "GOARCH" + bt + ") -- for bug reports and to confirm which build is installed. Takes no flags.",
	},
	{
		Name: "validate", Args: "[-e env.yaml]", Flags: []string{"-e"}, PosArg: posNone, Aliases: []string{"vld"},
		Synopsis: "validate [flags]",
		Summary:  "Lint the whole env.yaml + workflows", Example: "solmq-conn-util validate -e env.yaml",
		Detail: "Runs every check across the whole " + bt + "env.yaml" + bt + " (including any " + bt + "kubernetes:" + bt + "/" + bt + "docker:" + bt + "/" + bt + "podman:" + bt + " sections) and its workflows, printing all findings. Non-zero exit if any errors. Use it as a linter.",
	},
	{
		Name: "examples", Args: "[dir] [-f]", Flags: []string{"-f"}, TreeSuffix: bt + "[dir]" + bt, PosArg: posDir, Aliases: []string{"eg"},
		Synopsis: "examples [dir] [flags]",
		Summary:  "Write a starter env.yaml + workflows", Example: "solmq-conn-util examples ./myconfig",
		Detail: "Writes a starter " + bt + "env.yaml" + bt + " plus workflow files into " + bt + "dir" + bt + " (default: the current directory). Use " + bt + "-f" + bt + " to overwrite existing files.",
	},
	{
		Name: "download", Args: "[dir] [-e env.yaml] [" + urlFlagName + " u] [" + versionFlagName + " v] [" + omitLibFileFlagName + " file] [" + includeProvidedFlagName + "] [-f]",
		Flags: []string{"-e", urlFlagName, versionFlagName, omitLibFileFlagName, includeProvidedFlagName, "-f"}, PosArg: posDir, Aliases: []string{"dl"},
		Synopsis: "download jar <mq|syslog> [dir] [flags]",
		Blurb:    "Download IBM MQ or syslog encoder jars and their dependencies",
		Detail: "Downloads a fixed set of jars and their dependencies into a directory. All three words -- " +
			bt + "jar" + bt + ", then " + bt + libs.SetMQ + bt + " or " + bt + libs.SetSyslog + bt + " -- are required; a missing or unknown word is a loud error listing the valid words. " +
			bt + libs.SetMQ + bt + " seeds from the IBM MQ client jar; " + bt + libs.SetSyslog + bt + " seeds from the logstash syslog encoder jar. " +
			"The " + bt + libs.SetMQ + bt + " seed is the Jakarta build of the client, and there is no flag to change it: the connector image is a Jakarta stack, so the javax build could only ever produce a classpath that fails at run time. " +
			bt + "-e" + bt + " is read for one thing only: the " + bt + "image" + bt + " block, so the command can say when the jar list it omits against does not describe the image being deployed. It reads no credentials, no platform and no workflows, and a missing env.yaml is not an error -- download is the command you run before you have a deployment. " +
			"The seed artifact resolves to its latest stable release, or to the exact release named by " + versionSpan + " when given; an empty value means latest stable, the same as leaving the flag off. Every dependency version instead comes from the Maven POM chain of the resolved seed release. " +
			bt + urlFlagName + bt + " (repeatable) overrides all of that: when given, exactly those URLs are downloaded and no Maven resolution happens. " +
			"By default, an artifact resolved through Maven is omitted when the connector image already ships that jar, matched by artifact base name, at a version equal to or newer than the one resolved here; every omission is reported, never silent. The seed artifact -- the jar the command was run to fetch in the first place -- is never a candidate for omission no matter what the omit file says about it, so the command stays useful against an older image that lacks it entirely, and a stale or hostile omit file can never cause the one jar that matters to be skipped. " +
			omitLibFileSpan + " replaces the embedded jar list (captured from the shipped connector image) with one captured from a different image, so the comparison runs against that image instead; it REPLACES the embedded list completely rather than merging with it, so an omit file containing nothing omits nothing. " +
			includeProvidedSpan + " disables omission entirely and downloads the whole closure regardless of what the image already has. " +
			"Omission never applies to an explicit " + urlFlagName + ": the operator named that URL directly, so it is always downloaded verbatim and never second-guessed. " +
			"Matching is by jar artifact base name plus version, since a jar filename carries no groupId; this is why Jackson 3 (" + bt + "tools.jackson.core" + bt + ") still downloads for the " + bt + libs.SetSyslog + bt + " set even though the image already ships Jackson 2: its 3.x versions compare higher than the 2.x copies the image carries, so the version comparison still gets the right answer. " +
			"The destination is the trailing " + bt + "[dir]" + bt + " positional (default " + bt + "./libs" + bt + "); " + bt + "env.yaml" + bt + " is never read and there is no " + bt + "-e" + bt + " flag. " +
			"Every jar is checked against the sha1 digest the repository publishes beside it before it is written, catching a truncated or corrupted transfer; that is integrity, not authenticity -- it is not proof against a compromised repository. https is still required on the initial URL and on every redirect hop. An existing file is skipped unless " + bt + "-f" + bt + " is given, exactly like " + bt + "examples" + bt + ".",
		Targets: []cliTarget{
			{
				Name: "jar", Summary: "Download a set of jars and their dependencies into a directory", Example: "solmq-conn-util download jar " + libs.SetMQ + " ./libs",
				Sets: []cliTarget{
					{Name: libs.SetMQ, Summary: "Download the IBM MQ client jar and its dependencies", Example: "solmq-conn-util download jar " + libs.SetMQ + " ./libs"},
					{Name: libs.SetSyslog, Summary: "Download the logstash syslog encoder jar and its dependencies", Example: "solmq-conn-util download jar " + libs.SetSyslog + " ./libs"},
				},
			},
		},
	},
	{
		Name: "auto-complete", Args: "", Flags: nil, PosArg: posNone,
		Synopsis: "auto-complete <bash|zsh|fish|powershell>",
		Blurb:    "Print a shell completion script",
		Detail:   "Prints a completion script for the named shell on stdout, for you to source or drop into the shell's completion directory (see the per-shell examples below). The script is rendered from the same command model as this help, so the completion a binary prints always matches the commands that binary accepts.",
		Targets: []cliTarget{
			{Name: "bash", Summary: "Print the bash completion script", Example: "solmq-conn-util auto-complete bash > /etc/bash_completion.d/solmq-conn-util"},
			{Name: "zsh", Summary: "Print the zsh completion script", Example: "solmq-conn-util auto-complete zsh > ~/.zsh/completions/_solmq-conn-util"},
			{Name: "fish", Summary: "Print the fish completion script", Example: "solmq-conn-util auto-complete fish > ~/.config/fish/completions/solmq-conn-util.fish"},
			{Name: "powershell", Summary: "Print the PowerShell completion script", Example: "solmq-conn-util auto-complete powershell > solmq-conn-util-completion.ps1"},
		},
	},
	{
		Name: "help", Args: "[command]", Flags: nil, TreeSuffix: "(" + bt + "-h" + bt + ", " + bt + "--help" + bt + ")", PosArg: posNone,
		Synopsis: "help [command]",
		Summary:  "Print this summary, or the help page of one command",
		Example:  "solmq-conn-util help status",
		Detail:   "Prints the command summary (same as " + bt + "-h" + bt + " / " + bt + "--help" + bt + "). With a command name, prints that command's own page -- its arguments, flags, and examples -- exactly like " + bt + "<command> -h" + bt + ". Requested help goes to stdout and exits 0; the same pages printed after a usage mistake go to stderr with exit 2.",
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

// invocation builds "solmq-conn-util <verb> [words...] [args]", skipping any
// empty word -- invocation(v, "") (no target) and invocation(v, tg.Name,
// s.Name) (download/jar's three-level form) share the same helper.
func invocation(v cliVerb, words ...string) string {
	s := "solmq-conn-util " + v.Name
	for _, w := range words {
		if w != "" {
			s += " " + w
		}
	}
	if v.Args != "" {
		s += " " + v.Args
	}
	return s
}

// targetSuffix renders one target's own arrow suffix when it fans out into
// Sets (download/jar's mq | syslog), so a third command level shows in the
// command tree the same way a verb's Targets do -- "" when tg has none.
func targetSuffix(tg cliTarget) string {
	if len(tg.Sets) == 0 {
		return ""
	}
	names := make([]string, 0, len(tg.Sets))
	for _, s := range tg.Sets {
		names = append(names, bt+s.Name+bt)
	}
	return " -> " + strings.Join(names, " | ")
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
	add(bt + "env.yaml" + bt + ". Every short spelling in here is also listed on its own, keyed")
	add("by the abbreviation, in [abbreviation.md](abbreviation.md). Generated from the")
	add("command model in")
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
				names = append(names, bt+tg.Name+bt+targetSuffix(tg))
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
			if len(tg.Sets) == 0 {
				add("| " + bt + tableCell(invocation(v, tg.Name)) + bt + " | " + tg.Summary + " |")
				continue
			}
			for _, s := range tg.Sets {
				add("| " + bt + tableCell(invocation(v, tg.Name, s.Name)) + bt + " | " + s.Summary + " |")
			}
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
	add("| _other_ | " + bt + "cli" + bt + " only: once a session or command is running inside the instance, its own exit status is passed straight back. Anything that fails before that still uses the three codes above |")
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
			if len(tg.Sets) == 0 {
				add("#### " + bt + invocation(v, tg.Name) + bt)
				add("")
				add(tg.Summary + ".")
				add("")
				add("```sh")
				add(tg.Example)
				add("```")
				add("")
				continue
			}
			for _, s := range tg.Sets {
				add("#### " + bt + invocation(v, tg.Name, s.Name) + bt)
				add("")
				add(s.Summary + ".")
				add("")
				add("```sh")
				add(s.Example)
				add("```")
				add("")
			}
		}
	}

	return strings.TrimRight(strings.Join(l, "\n"), "\n") + "\n"
}

// verbUsage renders one command's own help page from the model: description,
// invocation, target words, the command's flags with their terse Usage texts,
// and an example. This page is the routing target for `help <command>`,
// `<command> -h`, and the "run 'solmq-conn-util help <command>'" hint after a
// usage error -- the detail deferred from the main summary lands here.
//
// Like the summary, it shows no aliases: the short spellings keep working but
// are documented only in docs/commands.md and the user guide.
func verbUsage(name string) string {
	for _, v := range cliVerbs {
		if v.Name != name {
			continue
		}
		var b strings.Builder
		b.WriteString(plainText(verbBlurb(v)) + "\n")
		b.WriteString("\nUsage:\n  solmq-conn-util " + v.Synopsis + "\n")

		if entries := targetEntries(v); len(entries) > 0 {
			b.WriteString("\nTargets:\n")
			writeAligned(&b, entries)
		}
		if entries := flagEntries(v); len(entries) > 0 {
			b.WriteString("\nFlags:\n")
			writeAligned(&b, entries)
		}
		if ex := verbExample(v); ex != "" {
			b.WriteString("\nExample:\n  " + ex + "\n")
		}
		return b.String()
	}
	return ""
}

// helpEntry is one aligned label/description pair on a help page.
type helpEntry struct{ label, desc string }

// helpPageWidth is the terminal width every help page fits in; a description
// longer than its column wraps with a hanging indent rather than running over.
const helpPageWidth = 100

// targetEntries flattens a verb's target words -- and their own further words
// (download jar mq) -- into help entries, canonical spellings only.
func targetEntries(v cliVerb) []helpEntry {
	var out []helpEntry
	for _, tg := range v.Targets {
		out = append(out, helpEntry{tg.Name, plainText(tg.Summary)})
		for _, st := range tg.Sets {
			out = append(out, helpEntry{tg.Name + " " + st.Name, plainText(st.Summary)})
		}
	}
	return out
}

// flagEntries renders the verb's flags in their modeled order: the offered
// spellings, a value placeholder for a flag that takes one, and the terse
// Usage text.
func flagEntries(v cliVerb) []helpEntry {
	byShort := make(map[string]cliFlag, len(cliFlags))
	for _, f := range cliFlags {
		byShort[f.Short] = f
	}
	var out []helpEntry
	for _, sh := range v.Flags {
		f, ok := byShort[sh]
		if !ok {
			continue // an unmodeled flag key; TestCompletionModelMetadataComplete catches it
		}
		label := strings.Join(flagOffered(f), ", ")
		switch f.Arg {
		case argFile:
			label += " <file>"
		case argName:
			label += " <value>"
		}
		out = append(out, helpEntry{label, f.Usage})
	}
	return out
}

// writeAligned prints entries as two columns: labels padded to the widest, and
// descriptions wrapped at helpPageWidth with a hanging indent, so a long Usage
// text folds under itself rather than past the edge of the terminal.
func writeAligned(b *strings.Builder, entries []helpEntry) {
	width := 0
	for _, e := range entries {
		if len(e.label) > width {
			width = len(e.label)
		}
	}
	indent := "  " + strings.Repeat(" ", width) + "  "
	for _, e := range entries {
		lines := wrapText(e.desc, helpPageWidth-len(indent))
		b.WriteString("  " + pad(e.label, width) + "  " + lines[0] + "\n")
		for _, more := range lines[1:] {
			b.WriteString(indent + more + "\n")
		}
	}
}

// wrapText breaks text on spaces into lines of at most width characters; a
// single word longer than width gets a line of its own rather than being cut.
// Always returns at least one line, so a caller can print lines[0] unguarded.
func wrapText(text string, width int) []string {
	words := strings.Fields(text)
	if len(words) == 0 {
		return []string{""}
	}
	lines := []string{words[0]}
	for _, w := range words[1:] {
		last := lines[len(lines)-1]
		if len(last)+1+len(w) <= width {
			lines[len(lines)-1] = last + " " + w
		} else {
			lines = append(lines, w)
		}
	}
	return lines
}

// verbExample is the example a help page shows: the verb's own, or its first
// target's when the verb-level invocation is not the interesting one
// (download's example lives on its jar target).
func verbExample(v cliVerb) string {
	if v.Example != "" {
		return v.Example
	}
	for _, tg := range v.Targets {
		if tg.Example != "" {
			return tg.Example
		}
	}
	return ""
}

// resolveTarget returns name's canonical target word for verb if name is a
// modeled alias, otherwise name unchanged (already canonical, or unknown --
// either way the caller's own switch decides that). Model-driven for the same
// reason resolveVerb is: an alias added to cliVerbs is accepted everywhere
// without a second table to keep in step.
func resolveTarget(verb, name string) string {
	for _, v := range cliVerbs {
		if v.Name != verb {
			continue
		}
		for _, tg := range v.Targets {
			if tg.Name == name {
				return name
			}
			for _, a := range tg.Aliases {
				if a == name {
					return tg.Name
				}
			}
		}
	}
	return name
}

// usageText is the in-binary help summary: one line per command, description
// from the model, nothing else -- everything more belongs to the command's own
// page (`help <command>` / `<command> -h`), which is where the old page's
// 30-line flag table and full argument forms went. Rendered from cliVerbs so
// it cannot drift, kept narrow enough never to wrap in a 100-column terminal
// (TestCommandsModelMatchesUsage gates the width), and deliberately free of
// aliases: the short spellings keep working but are documented only in
// docs/commands.md and the user guide.
func usageText() string {
	width := 0
	for _, v := range cliVerbs {
		if len(v.Name) > width {
			width = len(v.Name)
		}
	}
	var b strings.Builder
	b.WriteString("solmq-conn-util -- generate, deploy, and check the Solace PubSub+ Connector for IBM MQ\n")
	b.WriteString("\nUsage:\n  solmq-conn-util <command> [arguments] [flags]\n")
	b.WriteString("\nCommands:\n")
	for _, v := range cliVerbs {
		b.WriteString("  " + pad(v.Name, width) + "  " + plainText(verbBlurb(v)) + "\n")
	}
	b.WriteString("\nRun 'solmq-conn-util help <command>' (or '<command> -h') for its arguments, flags, and examples.\n")
	return b.String()
}

// pad right-pads s with spaces to width, the alignment every help column uses.
func pad(s string, width int) string {
	for len(s) < width {
		s += " "
	}
	return s
}
