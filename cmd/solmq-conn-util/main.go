// Command solmq-conn-util generates and deploys the Solace PubSub+ Connector for IBM
// MQ from a single env.yaml: it consolidates a folder of per-workflow YAML files
// into an application.yml and, per platform, into Kubernetes manifests, a docker
// compose file, or podman run/quadlet units -- and can apply, tear down, check the
// status of, or read the logs of those by shelling out to kubectl/oc, docker, or
// podman/systemctl.
//
//	solmq-conn-util generate [config] [--platform kubernetes|docker|podman] [-e env.yaml] [-o out]
//	solmq-conn-util deploy   [--platform kubernetes|docker|podman] [-e env.yaml]
//	solmq-conn-util remove   [--platform kubernetes|docker|podman] [-e env.yaml]
//	solmq-conn-util status   <container|application|all> [-d] [-w] [--all] [--output table|json] [--install] [--platform kubernetes|docker|podman] [-e env.yaml]
//	solmq-conn-util logs     [--follow] [--previous] [--tail N] [--since d] [--timestamps] [--all] [--platform kubernetes|docker|podman] [-e env.yaml]
//	solmq-conn-util version
//	solmq-conn-util validate [-e env.yaml]
//	solmq-conn-util examples [dir] [-f]
//	solmq-conn-util download jar mq|syslog [dir] [--url u] [--version v] [--omit-lib-file file] [--include-provided] [-f]
//	solmq-conn-util auto-complete bash|zsh|fish|powershell
package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/examples"
	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/gen"
	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/libs"
	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/runner"
	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/scan"
	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/spec"
	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/validate"
)

const (
	defaultEnvFile = "env.yaml"
	composeFile    = "docker-compose.yml"

	tgtConfig     = "config"
	tgtKubernetes = "kubernetes"
	tgtDocker     = "docker"
	tgtPodman     = "podman"
)

// platformNames lists the three deploy-platform section keys, in the order
// commands.go's --platform meaning documents them. generate (without
// "config"), deploy, remove, and status all resolve to one of these three via
// resolvePlatform; every platform-keyed map in this file (platformGenerators,
// actTargets) is gated against this same list by TestPlatformMapsCoverThreeNames.
var platformNames = []string{tgtKubernetes, tgtDocker, tgtPodman}

// platformAliasList is the accepted short spellings of a --platform value,
// paired with the canonical section key each resolves to. A slice rather than
// a bare map so the order is fixed: platformSpellings renders it into the
// "must be" error, and map iteration order would make that message -- and any
// test asserting it -- vary between runs once a platform has more than one
// alias.
//
// Curated rather than derived from a prefix rule, for the same reason the verb
// aliases are (cliVerb.Aliases): a prefix scheme silently changes meaning when
// a name is added. Only kubernetes has a widely recognized short form; dk and
// pm are this tool's own, which is why the --platform meaning and the user
// guide spell them out rather than leaving them to be guessed.
var platformAliasList = []struct{ Alias, Canonical string }{
	{"kube", tgtKubernetes},
	{"dk", tgtDocker},
	{"pm", tgtPodman},
}

// platformAliases is platformAliasList keyed for lookup.
var platformAliases = buildPlatformAliases()

func buildPlatformAliases() map[string]string {
	m := make(map[string]string, len(platformAliasList))
	for _, e := range platformAliasList {
		m[e.Alias] = e.Canonical
	}
	return m
}

// platformSpellings lists every accepted --platform value, canonical names
// first, for the "must be" error so a rejected value names its alternatives.
func platformSpellings() []string {
	out := append([]string(nil), platformNames...)
	for _, e := range platformAliasList {
		out = append(out, e.Alias)
	}
	return out
}

// resolvePlatformAlias returns val's canonical platform name when val is a
// known alias, otherwise val unchanged -- resolvePlatform's validation is what
// rejects an unknown value.
func resolvePlatformAlias(val string) string {
	if canon, ok := platformAliases[val]; ok {
		return canon
	}
	return val
}

// version is solmq-conn-util's own version, stamped at build time via
// -ldflags "-X main.version=<tag>"; "dev" is what an un-injected local build
// reports.
var version = "dev"

func main() {
	os.Exit(run(os.Args[1:]))
}

// run is the CLI entry point: it dispatches with the production Runner. Tests
// that need to inspect what would run call dispatch directly with a fake.
func run(args []string) int {
	return dispatch(args, runner.OS{})
}

// verbHandlers maps each modeled verb (cliVerbs in commands.go) to its
// implementation, threading r down to whichever action ends up shelling out.
// Keying dispatch off the model instead of a parallel switch is what
// TestDispatchHandlersMatchModel gates: a verb added to one side and not the
// other fails that test rather than drifting silently (see task 2's
// examples-default-dir drift for what this class of bug looks like).
var verbHandlers = map[string]func(args []string, r runner.Runner) int{
	"generate":      func(args []string, r runner.Runner) int { return runGenerate(args) },
	"deploy":        func(args []string, r runner.Runner) int { return runAction(runner.ActionDeploy, args, r) },
	"remove":        func(args []string, r runner.Runner) int { return runAction(runner.ActionRemove, args, r) },
	"status":        func(args []string, r runner.Runner) int { return runStatus(args, r) },
	"logs":          func(args []string, r runner.Runner) int { return runLogs(args, r) },
	"version":       func(args []string, r runner.Runner) int { return actVersion() },
	"validate":      func(args []string, r runner.Runner) int { return runValidate(args) },
	"examples":      func(args []string, r runner.Runner) int { return runExamples(args) },
	"download":      func(args []string, r runner.Runner) int { return runDownload(args) },
	"auto-complete": func(args []string, r runner.Runner) int { return runAutoComplete(args) },
	"help":          func(args []string, r runner.Runner) int { return runHelp(args) },
}

// verbAliases maps each alternate spelling (cliVerb.Aliases in commands.go) to
// its canonical verb name. Built from the model instead of hand-written so it
// can never drift from cliVerbs -- see resolveVerb.
var verbAliases = buildVerbAliases()

func buildVerbAliases() map[string]string {
	m := make(map[string]string)
	for _, v := range cliVerbs {
		for _, a := range v.Aliases {
			m[a] = v.Name
		}
	}
	return m
}

// resolveVerb returns name's canonical verb if name is a modeled alias,
// otherwise name unchanged (already canonical, or unknown -- either way
// dispatch's verbHandlers lookup is what decides that).
func resolveVerb(name string) string {
	if canon, ok := verbAliases[name]; ok {
		return canon
	}
	return name
}

// dispatch resolves args[0] against verbHandlers. -h/--help are aliases of the
// modeled "help" verb, handled before the lookup since they are not spellable
// as args[0] in the model itself. An alias from cliVerb.Aliases is resolved to
// its canonical verb the same way, right after that -h/--help normalization;
// the unknown-command message below still echoes args[0] as typed, never the
// resolved name.
func dispatch(args []string, r runner.Runner) int {
	if len(args) == 0 {
		usage(os.Stderr)
		return 2
	}
	if args[0] == "-h" || args[0] == "--help" {
		usage(os.Stdout)
		return 0
	}
	verb := resolveVerb(args[0])
	h, ok := verbHandlers[verb]
	if !ok {
		fmt.Fprintf(os.Stderr, "unknown command %q\n\n", args[0])
		usage(os.Stderr)
		return 2
	}
	// `<command> -h` straight after the verb prints that command's own page --
	// for every verb, including the ones that parse no flags of their own
	// (version, help). A -h later among the arguments reaches the same page
	// through the flag parser; see flagExit.
	if len(args) > 1 && (args[1] == "-h" || args[1] == "--help") {
		fmt.Fprint(os.Stdout, verbUsage(verb))
		return 0
	}
	return h(args[1:], r)
}

// runHelp implements the help verb: the top-level summary, or -- with a
// command name -- that command's own page, exactly what `<command> -h` prints.
// Requested help goes to stdout; only an unknown name is a usage error.
//
// The name is checked against the model (verbUsage answers "" for a verb it
// does not know) rather than against verbHandlers: this function is itself a
// verbHandlers entry, and consulting the map it lives in would be an
// initialization cycle. TestDispatchHandlersMatchModel keeps the two sets
// equal, so the answer is the same either way.
func runHelp(args []string) int {
	if len(args) == 0 {
		usage(os.Stdout)
		return 0
	}
	page := verbUsage(resolveVerb(args[0]))
	if page == "" {
		fmt.Fprintf(os.Stderr, "unknown command %q\n\n", args[0])
		usage(os.Stderr)
		return 2
	}
	fmt.Fprint(os.Stdout, page)
	return 0
}

// verbFlagSet is the flag set every verb parses with. Its Usage is silenced
// deliberately: Go's default dump lists both spellings of every flag
// alphabetically, with no targets and no examples, so a parse error is
// answered by flagExit's one-line hint and -h by the verb's own modeled page
// instead of that dump.
func verbFlagSet(name string) *flag.FlagSet {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.Usage = func() {}
	return fs
}

// flagExit routes a flag-parse failure. -h/--help is a request, not a mistake:
// the verb's own page goes to stdout and the exit is 0. Anything else is a
// usage error -- the flag package has already said what was wrong on stderr,
// so one hint naming the help page follows it, rather than a page nobody asked
// for scrolling the error away.
func flagExit(verb string, err error) int {
	if errors.Is(err, flag.ErrHelp) {
		fmt.Fprint(os.Stdout, verbUsage(verb))
		return 0
	}
	fmt.Fprintf(os.Stderr, "run 'solmq-conn-util help %s' for its arguments and flags\n", verb)
	return 2
}

// targetNames returns the modeled target/platform names for verb, in the order
// cliVerbs declares them ("" for a leaf verb with none).
func targetNames(verb string) []string {
	for _, v := range cliVerbs {
		if v.Name != verb {
			continue
		}
		names := make([]string, 0, len(v.Targets))
		for _, tg := range v.Targets {
			names = append(names, tg.Name)
		}
		return names
	}
	return nil
}

// targetSetNames returns the modeled Sets names for verb's target, in the
// order cliVerbs declares them ("" for an unknown verb/target, or a target
// with no Sets). Mirrors targetNames one level deeper, for download/jar's
// third command level.
func targetSetNames(verb, target string) []string {
	for _, v := range cliVerbs {
		if v.Name != verb {
			continue
		}
		for _, tg := range v.Targets {
			if tg.Name != target {
				continue
			}
			names := make([]string, 0, len(tg.Sets))
			for _, s := range tg.Sets {
				names = append(names, s.Name)
			}
			return names
		}
	}
	return nil
}

// pipeList joins names as "a|b|c", matching the wording of the existing
// "missing target/platform" messages.
func pipeList(names []string) string { return strings.Join(names, "|") }

// wantList joins names as "a, b, or c" (or "a or b" for two), matching the
// wording of the existing "unknown target/platform" messages.
func wantList(names []string) string {
	switch len(names) {
	case 0:
		return ""
	case 1:
		return names[0]
	case 2:
		return names[0] + " or " + names[1]
	default:
		return strings.Join(names[:len(names)-1], ", ") + ", or " + names[len(names)-1]
	}
}

// ---- generate ----------------------------------------------------------------

// genTargets maps generate's one modeled positional target ("config", in
// cliVerbs -- kubernetes/docker/podman are no longer positional, see
// platformGenerators) to its renderer, so runGenerate's accepted set can
// never drift from the model -- see TestDispatchHandlersMatchModel.
var genTargets = map[string]func(envPath, out string) int{
	tgtConfig: genConfig,
}

// platformGenerators maps each of platformNames to the renderer generate
// falls back to when no positional target is given: the resolved platform's
// artifacts instead of application.yml. Gated against platformNames by
// TestPlatformMapsCoverThreeNames.
var platformGenerators = map[string]func(envPath, out string) int{
	tgtKubernetes: genKubernetes,
	tgtDocker:     genDocker,
	tgtPodman:     genPodman,
}

// runGenerate handles both of generate's forms: the "config" positional
// (application.yml, no platform involved) and the platform form, which
// resolves --platform (or infers/prompts, see resolvePlatform) and renders
// that platform's artifacts. A positional naming an old-grammar platform
// (generate kubernetes, ...) is rejected with a hint at --platform instead of
// being resolved, per platformResolutionDetail in commands.go.
func runGenerate(args []string) int {
	fs := verbFlagSet("generate")
	env := envFlag(fs)
	out := outFlag(fs)
	platform := platformFlag(fs)
	pos, err := collectFlagsAndDirs(fs, args)
	if err != nil {
		return flagExit("generate", err)
	}
	if len(pos) > 0 {
		// Through resolveTarget, so the modeled short spelling (`cfg`) reaches the
		// same branch as the canonical word, the way status resolves its own target.
		switch {
		case resolveTarget("generate", pos[0]) == tgtConfig:
			return genConfig(*env, *out)
		case contains(platformNames, pos[0]):
			fmt.Fprintf(os.Stderr, "generate: %q is no longer a positional target; pass --platform %s instead\n", pos[0], pos[0])
			fmt.Fprint(os.Stderr, verbUsage("generate"))
			return 2
		default:
			fmt.Fprintf(os.Stderr, "generate: unknown target %q (want %s)\n", pos[0], wantList(append(targetNames("generate"), platformNames...)))
			return 2
		}
	}
	e, lerr := loadEnvFile(*env)
	if lerr != nil {
		return errExit(lerr)
	}
	platformName, perr := resolvePlatform(*platform, presentPlatforms(e), true)
	if perr != nil {
		return errExit(perr)
	}
	return platformGenerators[platformName](*env, *out)
}

func genConfig(envPath, out string) int {
	req, _, envDir, err := loadEnv(envPath)
	if err != nil {
		return errExit(err)
	}
	doc, errs, warns := gen.Config(req, resolver(envDir))
	printWarnings(warns)
	if len(errs) > 0 {
		return failFast(errs) // config: stop at the first error
	}
	return emit(out, doc)
}

func genKubernetes(envPath, out string) int {
	req, _, envDir, err := loadEnv(envPath)
	if err != nil {
		return errExit(err)
	}
	manifest, errs, warns := gen.GenerateKubernetes(req, resolver(envDir))
	printWarnings(warns)
	if len(errs) > 0 {
		return failFast(errs)
	}
	return emit(out, manifest)
}

func genDocker(envPath, out string) int {
	req, _, envDir, err := loadEnv(envPath)
	if err != nil {
		return errExit(err)
	}
	plan, errs, warns := gen.GenerateDocker(req, resolver(envDir))
	printWarnings(warns)
	if len(errs) > 0 {
		return failFast(errs)
	}
	return emit(out, plan.Compose)
}

func genPodman(envPath, out string) int {
	req, _, envDir, err := loadEnv(envPath)
	if err != nil {
		return errExit(err)
	}
	plan, errs, warns := gen.GeneratePodman(req, resolver(envDir), gen.PodmanOpts{})
	printWarnings(warns)
	if len(errs) > 0 {
		return failFast(errs)
	}
	if plan.Mode == spec.PodmanModeQuadlet {
		return emit(out, plan.Unit.Content)
	}
	return emit(out, plan.RunScript)
}

// ---- deploy / remove ---------------------------------------------------------

// actTargets maps each of platformNames to its deploy/remove implementation.
// Gated against platformNames by TestPlatformMapsCoverThreeNames (deploy/
// remove no longer model a positional Targets list in commands.go -- the
// platform is resolved by resolvePlatform, not looked up from args[0]).
// extraAllowed carries the values of a repeatable --allow-command flag (nil
// when unused).
var actTargets = map[string]func(action, envPath string, r runner.Runner, extraAllowed []string) int{
	tgtKubernetes: actKubernetes,
	tgtDocker:     actDocker,
	tgtPodman:     actPodman,
}

// runAction resolves --platform (or infers/prompts, see resolvePlatform) for
// deploy/remove and dispatches to actTargets. A positional argument is never
// a platform anymore: one naming an old-grammar platform (deploy kubernetes,
// ...) is rejected with a hint at --platform, and anything else is an
// unexpected argument -- both usage errors (exit 2), per
// platformResolutionDetail in commands.go.
func runAction(action string, args []string, r runner.Runner) int {
	fs := verbFlagSet(action)
	env := envFlag(fs)
	platform := platformFlag(fs)
	allow := allowCommandFlag(fs)
	pos, err := collectFlagsAndDirs(fs, args)
	if err != nil {
		return flagExit(action, err)
	}
	if len(pos) > 0 {
		if contains(platformNames, pos[0]) {
			fmt.Fprintf(os.Stderr, "%s: %q is no longer a positional platform; pass --platform %s instead\n", action, pos[0], pos[0])
			fmt.Fprint(os.Stderr, verbUsage(action))
			return 2
		}
		fmt.Fprintf(os.Stderr, "%s: unexpected argument %q (the platform is now selected with --platform)\n", action, pos[0])
		return 2
	}
	e, lerr := loadEnvFile(*env)
	if lerr != nil {
		return errExit(lerr)
	}
	platformName, perr := resolvePlatform(*platform, presentPlatforms(e), true)
	if perr != nil {
		return errExit(perr)
	}
	return actTargets[platformName](action, *env, r, *allow)
}

// preflight runs the read-only login/reachability probe (runner.Preflight)
// before any generate-write or mutating exec. On failure it reports the
// probe's combined output and wrapped error (which already names the platform
// and carries a login hint) and the caller must stop without writing or
// executing anything further.
func preflight(r runner.Runner, action, platform, command, namespace string, extraAllowed []string) (code int, ok bool) {
	out, err := runner.Preflight(r, platform, command, action, namespace, extraAllowed)
	if err != nil {
		return report(action, platform, out, err), false
	}
	return 0, true
}

func actKubernetes(action, envPath string, r runner.Runner, extraAllowed []string) int {
	req, e, envDir, err := loadEnv(envPath)
	if err != nil {
		return errExit(err)
	}
	if e.Kubernetes == nil {
		return errExit(fmt.Errorf("env.yaml has no kubernetes: section to %s", action))
	}
	manifest, errs, warns := gen.GenerateKubernetes(req, resolver(envDir), extraAllowed...)
	printWarnings(warns)
	if len(errs) > 0 {
		return failFast(errs)
	}
	if code, ok := preflight(r, action, validate.PlatformKubernetes, e.Kubernetes.Command, e.Kubernetes.Deployment.Namespace, extraAllowed); !ok {
		return code
	}
	out, rerr := runner.Kubernetes(r, e.Kubernetes.Command, action, manifest, extraAllowed)
	return report(action, tgtKubernetes, out, rerr)
}

func actDocker(action, envPath string, r runner.Runner, extraAllowed []string) int {
	req, e, envDir, err := loadEnv(envPath)
	if err != nil {
		return errExit(err)
	}
	if e.Docker == nil {
		return errExit(fmt.Errorf("env.yaml has no docker: section to %s", action))
	}
	res := resolver(envDir)
	plan, errs, warns := gen.GenerateDocker(req, res, extraAllowed...)
	printWarnings(warns)
	if len(errs) > 0 {
		return failFast(errs)
	}
	// Resolve before writing anything: an unset credential should fail with a
	// named variable, not leave a compose file behind for a run that never began.
	kvs, cerr := gen.ResolveCredentials(plan.Secrets, res)
	if cerr != nil {
		return errExit(cerr)
	}
	if code, ok := preflight(r, action, validate.PlatformDocker, e.Docker.Command, "", extraAllowed); !ok {
		return code
	}
	path := filepath.Join(envDir, composeFile)
	if werr := runner.WriteFile(path, plan.Compose, 0o644); werr != nil {
		return errExit(werr)
	}
	out, rerr := runner.Docker(r, e.Docker.Command, action, path, envPairs(kvs), extraAllowed)
	// The compose file is regenerated by every deploy and remove, so it is
	// scratch. It is kept after a failure: a half-started `up` still needs a
	// compose file to `down` with.
	if rerr == nil {
		_ = os.Remove(path)
	}
	return report(action, tgtDocker, out, rerr)
}

// envPairs renders resolved credentials as KEY=VALUE entries for a child
// process's environment. The result is secret material: it goes straight to the
// runner and is never printed.
func envPairs(kvs []gen.KV) []string {
	out := make([]string, 0, len(kvs))
	for _, kv := range kvs {
		out = append(out, kv.Key+"="+kv.Val)
	}
	return out
}

func actPodman(action, envPath string, r runner.Runner, extraAllowed []string) int {
	req, e, envDir, err := loadEnv(envPath)
	if err != nil {
		return errExit(err)
	}
	if e.Podman == nil {
		return errExit(fmt.Errorf("env.yaml has no podman: section to %s", action))
	}
	sc, serr := runner.ResolveQuadletScope(e.Podman.Quadlet.Scope, e.Podman.Quadlet.Dir)
	if serr != nil {
		return errExit(serr)
	}
	res := resolver(envDir)
	// deploy/remove are always quadlet; BaseDir bakes absolute on-disk paths into
	// the units so systemd resolves them regardless of cwd.
	plan, errs, warns := gen.GeneratePodman(req, res, gen.PodmanOpts{ForceQuadlet: true, BaseDir: sc.Dir}, extraAllowed...)
	printWarnings(warns)
	if len(errs) > 0 {
		return failFast(errs)
	}
	if code, ok := preflight(r, action, validate.PlatformPodman, e.Podman.Command, "", extraAllowed); !ok {
		return code
	}
	if action == runner.ActionRemove {
		return podmanRemove(sc, plan, e.Podman, r, extraAllowed)
	}
	return podmanDeploy(sc, plan, e.Podman, res, r, extraAllowed)
}

func podmanDeploy(sc runner.QuadletScope, plan gen.PodmanPlan, p *spec.Podman, res gen.Resolver, r runner.Runner, extraAllowed []string) int {
	// Credentials first: a missing one should stop the deploy before any unit is
	// written, and the units reference secrets that must already exist.
	kvs, cerr := gen.ResolveCredentials(plan.Secrets, res)
	if cerr != nil {
		return errExit(cerr)
	}
	for _, kv := range kvs {
		out, err := runner.PodmanSecretCreate(r, p.Command, gen.PodmanSecretStoreName(p.Name, kv.Key), kv.Val, extraAllowed)
		if err != nil {
			return report(runner.ActionDeploy, tgtPodman, out, err)
		}
	}
	// application.yml now carries a live read-only credential (the reserved
	// status account's password, read back out of this same file by the
	// generated status script at run time -- see statusscript.Render), so it is
	// written 0600 rather than world/group-readable. The status script and the
	// unit carry no secret of their own -- the script only reads one back out,
	// the unit only stable secret names -- so 0644 is right for both.
	if err := runner.WriteFile(filepath.Join(sc.Dir, plan.AppYAML.Name), plan.AppYAML.Data, 0o600); err != nil {
		return errExit(err)
	}
	if err := runner.WriteFile(filepath.Join(sc.Dir, plan.StatusScript.Name), plan.StatusScript.Data, 0o644); err != nil {
		return errExit(err)
	}
	if err := runner.WriteFile(filepath.Join(sc.Dir, plan.Unit.Filename), plan.Unit.Content, 0o644); err != nil {
		return errExit(err)
	}
	out, rerr := runner.PodmanDeploy(r, sc, []string{plan.Service})
	return report(runner.ActionDeploy, tgtPodman, out, rerr)
}

func podmanRemove(sc runner.QuadletScope, plan gen.PodmanPlan, p *spec.Podman, r runner.Runner, extraAllowed []string) int {
	out, rerr := runner.PodmanRemove(r, sc, []string{plan.Service}, []string{plan.Unit.Filename})
	// Best-effort cleanup of the files we generated.
	_ = os.Remove(filepath.Join(sc.Dir, plan.AppYAML.Name))
	_ = os.Remove(filepath.Join(sc.Dir, plan.StatusScript.Name))
	// Credentials are removed from podman's store last, after the units that
	// referenced them are gone. Leaving credential material behind is worth
	// reporting even when the teardown itself succeeded, so a failure here
	// surfaces rather than being swallowed like the file cleanup above.
	names := make([]string, 0, len(plan.Secrets))
	for _, s := range plan.Secrets {
		names = append(names, gen.PodmanSecretStoreName(p.Name, s.Stable))
	}
	so, serr := runner.PodmanSecretRemove(r, p.Command, names, extraAllowed)
	out += so
	if rerr == nil {
		rerr = serr
	}
	return report(runner.ActionRemove, tgtPodman, out, rerr)
}

// ---- platform resolution (shared by generate/deploy/remove/status) ----------

// contains reports whether s appears in names.
func contains(names []string, s string) bool {
	for _, n := range names {
		if n == s {
			return true
		}
	}
	return false
}

// presentPlatforms returns which of platformNames have a section in e, in
// platformNames order. A nil e (status with no env.yaml at all -- see
// loadStatusEnv) reports none present.
func presentPlatforms(e *spec.Env) []string {
	if e == nil {
		return nil
	}
	var out []string
	if e.Kubernetes != nil {
		out = append(out, tgtKubernetes)
	}
	if e.Docker != nil {
		out = append(out, tgtDocker)
	}
	if e.Podman != nil {
		out = append(out, tgtPodman)
	}
	return out
}

// resolvePlatform implements the order platformResolutionDetail (commands.go)
// documents for generate/deploy/remove/status: flagVal if set; otherwise the
// single entry in present, echoed to stderr so the operator sees what ran;
// otherwise an interactive numbered menu over present; zero entries in
// present is a loud error naming all three section keys.
//
// requireSection gates whether a given flagVal must name an entry actually in
// present: true for every caller except status with explicit --pod/--container
// targets, which names its own targets and so needs no section to run
// against them (see actStatus).
func resolvePlatform(flagVal string, present []string, requireSection bool) (string, error) {
	if flagVal != "" {
		// Resolved first so the section check, the messages below, and every
		// caller see a canonical name -- an operator who typed an alias is told
		// which section: key to add, not which alias they used.
		flagVal = resolvePlatformAlias(flagVal)
		if !contains(platformNames, flagVal) {
			return "", fmt.Errorf("--platform %q must be %s", flagVal, wantList(platformSpellings()))
		}
		if requireSection && !contains(present, flagVal) {
			if len(present) == 0 {
				return "", fmt.Errorf("--platform %s: env.yaml has no kubernetes:, docker:, or podman: section", flagVal)
			}
			return "", fmt.Errorf("--platform %s: env.yaml has no %s: section (present: %s)", flagVal, flagVal, strings.Join(present, ", "))
		}
		return flagVal, nil
	}
	switch len(present) {
	case 0:
		return "", fmt.Errorf("env.yaml has no kubernetes:, docker:, or podman: section; pass --platform to select one")
	case 1:
		fmt.Fprintf(os.Stderr, "platform: %s (the only section present in env.yaml)\n", present[0])
		return present[0], nil
	default:
		return promptPlatformMenu(present)
	}
}

// promptLine is the injectable stdin-read seam behind the interactive
// platform menu and the status install confirmation: readStdinLine in
// production, a canned fake in tests (mirroring the runner.Runner seam
// dispatch already threads explicitly -- see useFakeRunner in main_test.go).
var promptLine = readStdinLine

// readStdinLine refuses to read when stdin is not a character device (a
// script, a CI job, or anything else that is not an interactive terminal) so
// a non-interactive invocation fails fast with actionable guidance instead of
// blocking forever on a read that will never return.
func readStdinLine(question string) (string, error) {
	fi, err := os.Stdin.Stat()
	if err != nil || fi.Mode()&os.ModeCharDevice == 0 {
		return "", fmt.Errorf("stdin is not a terminal, so this prompt cannot be answered interactively")
	}
	fmt.Fprint(os.Stderr, question)
	line, rerr := bufio.NewReader(os.Stdin).ReadString('\n')
	if rerr != nil && line == "" {
		return "", fmt.Errorf("reading input: %w", rerr)
	}
	return strings.TrimSpace(line), nil
}

// promptPlatformMenu lists present as a numbered menu and reads the operator's
// choice via promptLine. A non-TTY refusal is wrapped with a --platform hint
// so the operator has an immediate next step instead of a bare "not a
// terminal" error.
func promptPlatformMenu(present []string) (string, error) {
	var b strings.Builder
	b.WriteString("multiple platforms are configured in env.yaml; choose one:\n")
	for i, p := range present {
		fmt.Fprintf(&b, "  %d) %s\n", i+1, p)
	}
	b.WriteString("> ")
	line, err := promptLine(b.String())
	if err != nil {
		return "", fmt.Errorf("choosing a platform interactively: %w; pass --platform instead", err)
	}
	n, cerr := strconv.Atoi(line)
	if cerr != nil || n < 1 || n > len(present) {
		return "", fmt.Errorf("%q is not a valid choice (want 1-%d); pass --platform instead", line, len(present))
	}
	return present[n-1], nil
}

// platformFlag registers the shared --platform selector (generate/deploy/
// remove/status; no short alias) and returns its raw value for
// resolvePlatform to interpret.
func platformFlag(fs *flag.FlagSet) *string {
	const u = "the platform: kubernetes, docker, or podman (default: resolved from env.yaml, or an interactive menu)"
	return fs.String("platform", "", u)
}

// loadEnvFile reads and parses env.yaml only, with no workflow scan: platform
// resolution and status need the parsed sections but never the workflow set
// loadEnv additionally reads for generate/deploy/remove.
func loadEnvFile(envPath string) (*spec.Env, error) {
	data, err := os.ReadFile(envPath)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", envPath, err)
	}
	return spec.ParseEnv(data)
}

// ---- version ---------------------------------------------------------------

// actVersion prints solmq-conn-util's own version (see the package-level
// version var) plus the Go toolchain and OS/arch it was built with. It takes
// no flags and always succeeds.
func actVersion() int {
	fmt.Println("solmq-conn-util", version, runtime.Version(), runtime.GOOS+"/"+runtime.GOARCH)
	return 0
}

// ---- validate / examples -----------------------------------------------------

func runValidate(args []string) int {
	fs := verbFlagSet("validate")
	env := envFlag(fs)
	if _, err := collectFlagsAndDirs(fs, args); err != nil {
		return flagExit("validate", err)
	}
	req, _, envDir, lerr := loadEnv(*env)
	if lerr != nil {
		return errExit(lerr)
	}
	errs, warns := gen.Validate(req, resolver(envDir))
	printWarnings(warns)
	if len(errs) > 0 {
		for _, e := range errs { // validate: report every error
			fmt.Fprintln(os.Stderr, "error:", e.String())
		}
		fmt.Fprintf(os.Stderr, "\n%d error(s)\n", len(errs))
		return 1
	}
	fmt.Fprintln(os.Stderr, "ok: no errors")
	return 0
}

// runExamples writes the embedded starter env.yaml + workflows to dir (default
// the current directory).
func runExamples(args []string) int {
	fs := verbFlagSet("examples")
	force := fs.Bool("f", false, "overwrite existing files")
	fs.BoolVar(force, "force", false, "overwrite existing files")
	pos, err := collectFlagsAndDirs(fs, args)
	if err != nil {
		return flagExit("examples", err)
	}
	dir := "."
	if len(pos) > 0 {
		dir = pos[0]
	}
	written, skipped, werr := examples.Write(dir, *force)
	if werr != nil {
		return errExit(werr)
	}
	for _, p := range written {
		fmt.Fprintln(os.Stderr, "wrote:", p)
	}
	for _, p := range skipped {
		fmt.Fprintln(os.Stderr, "exists (use -f to overwrite):", p)
	}
	fmt.Fprintf(os.Stderr, "\n%d written, %d skipped in %q\n", len(written), len(skipped), dir)
	fmt.Fprintf(os.Stderr, "next: solmq-conn-util generate config -e %s\n", filepath.Join(dir, defaultEnvFile))
	return 0
}

// ---- download --------------------------------------------------------------

// downloadSets maps each modeled download-jar set name (cliVerbs's "jar"
// target Sets in commands.go) to the libs.Input.Set value it selects, so
// runDownload's accepted set names can never drift from the model -- see
// TestDispatchHandlersMatchModel and TestDownloadSetMapMatchesModel, which
// also gate this map against internal/libs.SetNames() directly.
var downloadSets = map[string]string{
	libs.SetMQ:     libs.SetMQ,
	libs.SetSyslog: libs.SetSyslog,
}

// downloadFn is the injectable seam behind runDownload's call into
// internal/libs: libs.Download in production, a fake in tests that captures
// the Input it received instead of touching the network -- mirrors the
// runner.Runner seam dispatch threads explicitly, so this package's own tests
// stay decoupled from internal/libs's HTTP/Maven behavior. download never
// shells out, so it carries no runner.Runner of its own; this is its only
// injected boundary.
var downloadFn = libs.Download

// runDownload parses download's flags and validates all three positional
// words (target, set, optional dir) against the model before calling
// internal/libs -- mirrors runAutoComplete's missing/unknown-word handling
// (including its usage()-only-on-missing asymmetry) so the words accepted by
// help, completion, and dispatch can never drift from what actually runs.
// downloadDeployedImage reads the connector image reference out of env.yaml so
// download can tell the operator when the jar list it omits against was
// captured from a DIFFERENT image than the one being deployed. It is the only
// thing download reads config for -- never credentials, a platform, or
// workflows -- and the value only ever reaches Report.OmitListImageMismatch.
//
// The read is advisory because download is the command you run BEFORE you have
// a deployment: it must work in an empty directory with no config at all. So
// the failure modes follow the precedent --omit-lib-file already sets, where a
// file the operator NAMED failing to load is systemic but an absent default is
// not:
//
//   - -e given explicitly and unreadable or malformed -> error, nothing runs.
//   - -e defaulted and env.yaml absent -> "", no error, no warning.
//   - -e defaulted and env.yaml malformed -> "", and a note on stderr, because
//     silently ignoring a config that IS there would hide the very mismatch
//     this exists to surface.
//   - parsed but no image: block -> "", nothing to compare.
//
// fs.Visit is what separates the first two: comparing *envPath against the
// default would read a deliberately typed "-e env.yaml" as if it were implicit.
func downloadDeployedImage(fs *flag.FlagSet, envPath string) (string, error) {
	explicit := false
	fs.Visit(func(f *flag.Flag) {
		if f.Name == "e" || f.Name == "env" {
			explicit = true
		}
	})

	e, err := loadEnvFile(envPath)
	if err != nil {
		if explicit {
			return "", fmt.Errorf("reading the image reference from %s: %w", envPath, err)
		}
		if errors.Is(err, os.ErrNotExist) {
			return "", nil
		}
		fmt.Fprintf(os.Stderr, "note: %s could not be read (%v), so the omit list cannot be checked against the image you deploy\n", envPath, err)
		return "", nil
	}
	return e.Image.Ref(), nil
}

func runDownload(args []string) int {
	fs := verbFlagSet("download")
	var urls []string
	urlVal := repeatableName{&urls}
	fs.Var(urlVal, "url", "exact URL to download instead of Maven resolution (repeatable)")
	version := fs.String("version", "", "pin the seed release instead of resolving latest stable (default: latest stable)")
	omitLibFile := fs.String("omit-lib-file", "", "a jar list that replaces the embedded default the omission rule compares against (default: the embedded default; an empty file omits nothing)")
	includeProvided := fs.Bool("include-provided", false, "download the whole closure even where the connector image already provides a jar")
	envPath := envFlag(fs)
	force := fs.Bool("f", false, "overwrite existing files")
	fs.BoolVar(force, "force", false, "overwrite existing files")
	pos, err := collectFlagsAndDirs(fs, args)
	if err != nil {
		return flagExit("download", err)
	}

	if len(pos) == 0 {
		fmt.Fprintf(os.Stderr, "download: missing target (%s)\n\n", pipeList(targetNames("download")))
		fmt.Fprint(os.Stderr, verbUsage("download"))
		return 2
	}
	target := pos[0]
	if !contains(targetNames("download"), target) {
		fmt.Fprintf(os.Stderr, "download: unknown target %q (want %s)\n", target, wantList(targetNames("download")))
		return 2
	}

	setWords := targetSetNames("download", target)
	if len(pos) < 2 {
		fmt.Fprintf(os.Stderr, "download: missing set (%s)\n\n", pipeList(setWords))
		fmt.Fprint(os.Stderr, verbUsage("download"))
		return 2
	}
	setArg := pos[1]
	set, ok := downloadSets[setArg]
	if !ok {
		fmt.Fprintf(os.Stderr, "download: unknown set %q (want %s)\n", setArg, wantList(setWords))
		return 2
	}

	dir := "./libs"
	if len(pos) > 2 {
		dir = pos[2]
	}
	if len(pos) > 3 {
		fmt.Fprintf(os.Stderr, "download: unexpected argument %q\n", pos[3])
		return 2
	}

	deployedImage, ierr := downloadDeployedImage(fs, *envPath)
	if ierr != nil {
		return errExit(ierr)
	}

	rep, derr := downloadFn(libs.Input{
		Dir:             dir,
		Set:             set,
		Version:         *version,
		URLs:            urls,
		Force:           *force,
		OmitLibFile:     *omitLibFile,
		IncludeProvided: *includeProvided,
		DeployedImage:   deployedImage,
	})
	if derr != nil {
		return errExit(derr)
	}
	return reportDownload(rep, dir, *omitLibFile)
}

// reportDownload prints libs.Download's Report mirroring runExamples' shape:
// "wrote:"/"exists (use -f to overwrite):"/"failed:" lines, then three
// clearly distinct blocks that must never blur together -- "omitted:" for an
// artifact the connector image already provides (Report.Omitted; not a
// problem, so it never affects the exit code), "guessed version:" for one
// whose POM chain never resolved a version and fell back to latest stable
// (Report.Fallback), and "unverified:" for one written without a digest
// match, e.g. an explicit --url whose ".sha1" sidecar 404s (Report.Unverified;
// also not a failure, but worth an operator's attention) -- then which omit
// list was in effect (Report.OmitListProvenance, annotated "(built in)" when
// omitLibFile was left empty) and any per-line omit list warnings
// (Report.OmitListWarnings), then a counts footer, then a "next:" hint pointing at
// the docker/podman libs.dir and kubernetes libs: config keys. When every
// resolved artifact was omitted, the hint instead names --include-provided,
// since there is no dir to point deployment config at and -f cannot revive a
// jar the omission step already dropped.
//
// Exit code convention: this returns 1 whenever Report.Failed is non-empty,
// whether one artifact failed among many or every one of them did -- the
// exit code only ever signals "at least one failure"; a caller that needs to
// tell a partial failure from a total one reads the written/skipped/omitted/
// failed counts in the footer, not the exit code. A run where every artifact
// was omitted (the image already has all of it) has an empty Failed and so
// exits 0: that is success, not failure.
func reportDownload(rep libs.Report, dir string, omitLibFile string) int {
	for _, p := range rep.Written {
		fmt.Fprintln(os.Stderr, "wrote:", p)
	}
	for _, p := range rep.Skipped {
		fmt.Fprintln(os.Stderr, "exists (use -f to overwrite):", p)
	}
	for _, f := range rep.Failed {
		fmt.Fprintf(os.Stderr, "failed: %s: %v\n", f.Name, f.Err)
	}
	for _, note := range rep.Omitted {
		fmt.Fprintln(os.Stderr, "omitted:", note)
	}
	for _, note := range rep.Fallback {
		fmt.Fprintln(os.Stderr, "guessed version:", note)
	}
	for _, note := range rep.Unverified {
		fmt.Fprintln(os.Stderr, "unverified:", note)
	}
	if rep.OmitListProvenance != "" {
		line := "omit list: " + rep.OmitListProvenance
		if omitLibFile == "" {
			line += " (built in; describes " + libs.EmbeddedListMinVersion + " and later)"
		}
		fmt.Fprintln(os.Stderr, line)
	}
	for _, w := range rep.OmitListWarnings {
		fmt.Fprintln(os.Stderr, "omit list warning:", w)
	}
	if rep.OmitListImageMismatch != "" {
		fmt.Fprintln(os.Stderr, "omit list warning:", rep.OmitListImageMismatch)
	}
	fmt.Fprintf(os.Stderr, "\n%d written, %d skipped, %d omitted, %d unverified, %d failed in %q\n",
		len(rep.Written), len(rep.Skipped), len(rep.Omitted), len(rep.Unverified), len(rep.Failed), dir)

	everythingOmitted := len(rep.Omitted) > 0 && len(rep.Written) == 0 && len(rep.Skipped) == 0 && len(rep.Failed) == 0
	if everythingOmitted {
		fmt.Fprintln(os.Stderr, "next: every resolved jar is already on the connector image, so nothing was downloaded; --include-provided downloads the full closure anyway (-f only overwrites a file already on disk, it never revives an omitted jar)")
	} else {
		next := fmt.Sprintf("next: point the docker/podman libs.dir or kubernetes libs: config key at %q", dir)
		if len(rep.Omitted) > 0 {
			next += "; the omitted jars are already on the connector image and need no copy"
		}
		fmt.Fprintln(os.Stderr, next)
	}
	if len(rep.Failed) > 0 {
		return 1
	}
	return 0
}

// ---- auto-complete -------------------------------------------------------

// completionShellRenderers maps each modeled "auto-complete" shell (cliVerbs in
// commands.go) to its renderer, so runAutoComplete's accepted set can never
// drift from the model -- see TestDispatchHandlersMatchModel.
var completionShellRenderers = map[string]func() string{
	"bash":       renderBashCompletion,
	"zsh":        renderZshCompletion,
	"fish":       renderFishCompletion,
	"powershell": renderPowerShellCompletion,
}

// runAutoComplete prints the named shell's completion script on stdout, for
// the caller to redirect or source. Nothing is read or written beyond that:
// the script is rendered from the compiled-in command model, so it needs no
// env.yaml and matches the binary that printed it.
func runAutoComplete(args []string) int {
	fs := verbFlagSet("auto-complete")
	pos, err := collectFlagsAndDirs(fs, args)
	if err != nil {
		return flagExit("auto-complete", err)
	}
	if len(pos) == 0 {
		fmt.Fprintf(os.Stderr, "auto-complete: missing shell (%s)\n\n", pipeList(targetNames("auto-complete")))
		fmt.Fprint(os.Stderr, verbUsage("auto-complete"))
		return 2
	}
	h, ok := completionShellRenderers[pos[0]]
	if !ok {
		fmt.Fprintf(os.Stderr, "auto-complete: unknown shell %q (want %s)\n", pos[0], wantList(targetNames("auto-complete")))
		return 2
	}
	return emit("", h())
}

// ---- shared flags + loading --------------------------------------------------

// envFlag registers the shared -e/--env config-file selector (default env.yaml).
func envFlag(fs *flag.FlagSet) *string {
	const u = "config file; relative or absolute path (default: env.yaml)"
	e := fs.String("e", defaultEnvFile, u)
	fs.StringVar(e, "env", defaultEnvFile, u)
	return e
}

// outFlag registers the shared -o/--out generate destination (default stdout).
func outFlag(fs *flag.FlagSet) *string {
	const u = "write output to a file (default: stdout)"
	o := fs.String("o", "", u)
	fs.StringVar(o, "out", "", u)
	return o
}

// allowCommandValue implements flag.Value for a repeatable --allow-command: it
// validates each value against the same bare-name/safe-charset rules
// CheckDeployCommand applies to the command itself (a path or an unsafe
// character in the escape hatch would just move the hole), then appends it.
type allowCommandValue struct{ vals *[]string }

func (allowCommandValue) String() string { return "" } // flag.Value needs a zero-value String; nothing to show before parsing

func (v allowCommandValue) Set(s string) error {
	if strings.ContainsRune(s, '/') {
		return fmt.Errorf("--allow-command %q: a path is not accepted here; use a bare binary name, resolved from PATH", s)
	}
	if !validate.SafeToken(s) {
		return fmt.Errorf("--allow-command %q: contains an unsafe character (%s)", s, validate.UnsafeTokenReason)
	}
	*v.vals = append(*v.vals, s)
	return nil
}

// allowCommandFlag registers the repeatable --allow-command flag (deploy/remove
// only -- runGenerate/runValidate never call this) and returns the accumulated,
// already-validated values. A bad value fails fs.Parse, which callers surface
// as the standard usage-error exit (2).
func allowCommandFlag(fs *flag.FlagSet) *[]string {
	vals := &[]string{}
	fs.Var(allowCommandValue{vals}, "allow-command", "approve an extra command binary for this invocation (repeatable)")
	return vals
}

// collectFlagsAndDirs parses fs while tolerating flags before, after, or between
// positional args (a bare flag.Parse stops at the first non-flag token). It
// returns the positional args in order.
func collectFlagsAndDirs(fs *flag.FlagSet, args []string) ([]string, error) {
	var pos []string
	for len(args) > 0 {
		if err := fs.Parse(args); err != nil {
			return nil, err
		}
		if fs.NArg() == 0 {
			break
		}
		pos = append(pos, fs.Arg(0))
		args = fs.Args()[1:]
	}
	return pos, nil
}

// loadEnv reads+parses env.yaml, scans workflows.dir (resolved relative to the
// env file) for the matching workflow files, and returns the request plus the
// parsed Env (for deploy commands) and the env file's directory (for path
// resolution).
func loadEnv(envPath string) (gen.Request, *spec.Env, string, error) {
	data, err := os.ReadFile(envPath)
	if err != nil {
		return gen.Request{}, nil, "", fmt.Errorf("reading %s: %w", envPath, err)
	}
	e, err := spec.ParseEnv(data)
	if err != nil {
		return gen.Request{}, nil, "", err
	}
	envDir := filepath.Dir(envPath)
	absEnv, _ := filepath.Abs(envPath)
	wfDir := e.Workflows.Dir
	if !filepath.IsAbs(wfDir) {
		wfDir = filepath.Join(envDir, wfDir)
	}
	sr, err := scan.Scan(wfDir, e.Workflows.FilePattern, absEnv)
	if err != nil {
		return gen.Request{}, nil, "", err
	}
	req := gen.Request{Env: &gen.File{Name: filepath.Base(envPath), Data: data}}
	for _, p := range sr.WorkflowFiles {
		wd, rerr := os.ReadFile(p)
		if rerr != nil {
			return gen.Request{}, nil, "", rerr
		}
		req.Workflows = append(req.Workflows, gen.File{Name: filepath.Base(p), Data: wd})
	}
	return req, e, envDir, nil
}

func resolver(envDir string) gen.Resolver {
	return gen.Resolver{Env: os.LookupEnv, ReadFile: fileReader(envDir), Abs: absResolver(envDir)}
}

// fileReader resolves values-file / .jks paths relative to the env file's dir.
func fileReader(dir string) func(string) ([]byte, error) {
	return func(p string) ([]byte, error) { return os.ReadFile(absPath(dir, p)) }
}

func absResolver(dir string) func(string) string {
	return func(p string) string { return absPath(dir, p) }
}

func absPath(dir, p string) string {
	if filepath.IsAbs(p) {
		return p
	}
	return filepath.Join(dir, p)
}

// ---- output helpers ----------------------------------------------------------

func emit(out, content string) int {
	if out == "" {
		fmt.Print(content)
		return 0
	}
	if err := os.WriteFile(out, []byte(content), 0o600); err != nil {
		return errExit(err)
	}
	return 0
}

// report prints a deployer's combined output and a final ok/error line.
func report(action, platform, out string, err error) int {
	if s := strings.TrimSpace(out); s != "" {
		fmt.Fprintln(os.Stderr, s)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s %s failed: %v\n", action, platform, err)
		return 1
	}
	fmt.Fprintf(os.Stderr, "ok: %s %s\n", action, platform)
	return 0
}

func errExit(err error) int {
	fmt.Fprintln(os.Stderr, "error:", err)
	return 1
}

func failFast(errs []gen.Issue) int {
	fmt.Fprintln(os.Stderr, "error:", errs[0].String())
	return 1
}

func printWarnings(warns []gen.Issue) {
	for _, w := range warns {
		fmt.Fprintln(os.Stderr, "warning:", w.String())
	}
}

// usage prints the top-level summary. The stream is the caller's statement of
// intent: stdout when help was asked for (help, -h, --help -- pipeable, exit
// 0), stderr when it is the side effect of a usage mistake (exit 2).
func usage(w *os.File) {
	fmt.Fprint(w, usageText())
}
