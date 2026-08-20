// Command solmq-conn-util generates and deploys the Solace PubSub+ Connector for IBM
// MQ from a single env.yaml: it consolidates a folder of per-workflow YAML files
// into an application.yml and, per platform, into Kubernetes manifests, a docker
// compose file, or podman run/quadlet units -- and can apply, tear down, or check
// the status of those by shelling out to kubectl/oc, docker, or podman/systemctl.
//
//	solmq-conn-util generate [config] [--platform kubernetes|docker|podman] [-e env.yaml] [-o out]
//	solmq-conn-util deploy   [--platform kubernetes|docker|podman] [-e env.yaml]
//	solmq-conn-util remove   [--platform kubernetes|docker|podman] [-e env.yaml]
//	solmq-conn-util status   [--install] [--platform kubernetes|docker|podman] [-e env.yaml]
//	solmq-conn-util version
//	solmq-conn-util validate [-e env.yaml]
//	solmq-conn-util examples [dir] [-f]
//	solmq-conn-util auto-complete bash|zsh|fish|powershell
package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/examples"
	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/gen"
	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/runner"
	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/scan"
	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/spec"
	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/statusscript"
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
	"version":       func(args []string, r runner.Runner) int { return actVersion() },
	"validate":      func(args []string, r runner.Runner) int { return runValidate(args) },
	"examples":      func(args []string, r runner.Runner) int { return runExamples(args) },
	"auto-complete": func(args []string, r runner.Runner) int { return runAutoComplete(args) },
	"help":          func(args []string, r runner.Runner) int { usage(); return 0 },
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
		usage()
		return 2
	}
	if args[0] == "-h" || args[0] == "--help" {
		usage()
		return 0
	}
	verb := resolveVerb(args[0])
	h, ok := verbHandlers[verb]
	if !ok {
		fmt.Fprintf(os.Stderr, "unknown command %q\n\n", args[0])
		usage()
		return 2
	}
	return h(args[1:], r)
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
	fs := flag.NewFlagSet("generate", flag.ContinueOnError)
	env := envFlag(fs)
	out := outFlag(fs)
	platform := platformFlag(fs)
	pos, err := collectFlagsAndDirs(fs, args)
	if err != nil {
		return 2
	}
	if len(pos) > 0 {
		switch {
		case pos[0] == tgtConfig:
			return genConfig(*env, *out)
		case contains(platformNames, pos[0]):
			fmt.Fprintf(os.Stderr, "generate: %q is no longer a positional target; pass --platform %s instead\n", pos[0], pos[0])
			usage()
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
	fs := flag.NewFlagSet(action, flag.ContinueOnError)
	env := envFlag(fs)
	platform := platformFlag(fs)
	allow := allowCommandFlag(fs)
	pos, err := collectFlagsAndDirs(fs, args)
	if err != nil {
		return 2
	}
	if len(pos) > 0 {
		if contains(platformNames, pos[0]) {
			fmt.Fprintf(os.Stderr, "%s: %q is no longer a positional platform; pass --platform %s instead\n", action, pos[0], pos[0])
			usage()
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

// ---- status --------------------------------------------------------------

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

// runStatus parses status's flags and hands them to actStatus.
func runStatus(args []string, r runner.Runner) int {
	fs := flag.NewFlagSet("status", flag.ContinueOnError)
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
	pos, err := collectFlagsAndDirs(fs, args)
	if err != nil {
		return 2
	}
	if len(pos) > 0 {
		fmt.Fprintf(os.Stderr, "status: unexpected argument %q\n", pos[0])
		return 2
	}
	return actStatus(*env, *install, *platform, pods, containers, *namespace, *port, *user, *command, *allow, r)
}

// loadStatusEnv loads env.yaml the same way loadEnvFile does, except when the
// operator has named explicit --pod/--container targets and an explicit
// --platform: status then needs nothing from the file at all (see
// platformResolutionDetail's status exception), so a missing file is not an
// error and status proceeds against a zero Env (built-in command/port/user
// defaults). A file that does exist but fails to parse is still reported --
// its presence means the operator expects it to be read.
func loadStatusEnv(envPath string, skipIfMissing bool) (*spec.Env, error) {
	data, err := os.ReadFile(envPath)
	if err != nil {
		if skipIfMissing {
			return &spec.Env{}, nil
		}
		return nil, fmt.Errorf("reading %s: %w", envPath, err)
	}
	return spec.ParseEnv(data)
}

// statusCommand resolves the CLI binary status execs through: the --command
// override if given, else the platform's section command: in env.yaml when
// the section is present, else the platform's own default -- the same
// defaults applyKubeDefaults/applyDockerDefaults/applyPodmanDefaults fill in
// at parse time, needed again here because status can run with no section at
// all (explicit --pod/--container targets).
func statusCommand(platform, override string, e *spec.Env) string {
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

// statusNamespace resolves the kubernetes namespace status queries: the
// --namespace override if given, else the kubernetes section's
// deployment.namespace when present, else "" (the CLI's current-context
// default). No effect on docker/podman.
func statusNamespace(platform, override string, e *spec.Env) string {
	if override != "" {
		return override
	}
	if platform == validate.PlatformKubernetes && e != nil && e.Kubernetes != nil {
		return e.Kubernetes.Deployment.Namespace
	}
	return ""
}

// resolveStatusTargets returns the targets status checks: the operator's own
// --pod/--container values when given, otherwise discovered from env.yaml --
// kubernetes lists running pods (runner.KubernetesPodNames, selector
// app=<deployment name>), docker/podman use the section's configured
// instance name.
func resolveStatusTargets(platform string, pods, containers []string, namespace string, e *spec.Env, r runner.Runner, cmdArgv []string) ([]string, error) {
	switch platform {
	case validate.PlatformKubernetes:
		if len(pods) > 0 {
			return pods, nil
		}
		if e == nil || e.Kubernetes == nil {
			return nil, fmt.Errorf("no kubernetes: section in env.yaml to discover pods from; pass --pod explicitly")
		}
		selector := "app=" + e.Kubernetes.Deployment.Name
		names, err := runner.KubernetesPodNames(r, cmdArgv, namespace, selector)
		if err != nil {
			return nil, err
		}
		if len(names) == 0 {
			return nil, fmt.Errorf("no pods found for selector %q in namespace %q; pass --pod explicitly", selector, namespace)
		}
		return names, nil
	case validate.PlatformDocker:
		if len(containers) > 0 {
			return containers, nil
		}
		if e == nil || e.Docker == nil {
			return nil, fmt.Errorf("no docker: section in env.yaml to discover the container name from; pass --container explicitly")
		}
		return []string{e.Docker.Name}, nil
	case validate.PlatformPodman:
		if len(containers) > 0 {
			return containers, nil
		}
		if e == nil || e.Podman == nil {
			return nil, fmt.Errorf("no podman: section in env.yaml to discover the container name from; pass --container explicitly")
		}
		return []string{e.Podman.Name}, nil
	default:
		return nil, fmt.Errorf("unknown platform %q", platform)
	}
}

// actStatus resolves the platform and its targets, ensures the generated
// status script is present on each (installing per install/prompt policy),
// runs it, and prints the per-target report. See platformResolutionDetail and
// the status Detail text in commands.go for the resolution order and the
// install-or-skip behavior.
func actStatus(envPath string, install bool, platformFlagVal string, pods, containers []string, namespaceFlagVal string, portFlagVal int, userFlagVal, commandFlagVal string, extraAllowed []string, r runner.Runner) int {
	explicitTargets := len(pods) > 0 || len(containers) > 0

	e, lerr := loadStatusEnv(envPath, explicitTargets && platformFlagVal != "")
	if lerr != nil {
		return errExit(lerr)
	}

	platform, perr := resolvePlatform(platformFlagVal, presentPlatforms(e), !explicitTargets)
	if perr != nil {
		return errExit(perr)
	}

	namespace := statusNamespace(platform, namespaceFlagVal, e)
	if namespace != "" && !validate.SafeToken(namespace) {
		return errExit(fmt.Errorf("--namespace %q contains an unsafe character (%s)", namespace, validate.UnsafeTokenReason))
	}
	for _, p := range pods {
		if !validate.SafeToken(p) {
			return errExit(fmt.Errorf("--pod %q contains an unsafe character (%s)", p, validate.UnsafeTokenReason))
		}
	}
	for _, c := range containers {
		if !validate.SafeToken(c) {
			return errExit(fmt.Errorf("--container %q contains an unsafe character (%s)", c, validate.UnsafeTokenReason))
		}
	}

	user := userFlagVal
	if user == "" {
		user = spec.StatusUserName
	}
	// Stricter than the other flags: the account name is also spliced into a
	// sed address inside the generated script, where a '/' or a regex
	// metacharacter would break the password lookup rather than fail loudly.
	if !validate.SafeActuatorUser(user) {
		return errExit(fmt.Errorf("--user %q is not a usable account name (%s)", user, validate.SafeActuatorUserReason))
	}

	port := portFlagVal
	if port == 0 {
		port = e.Defaults.EffectiveManagementPort()
	}
	if port < 1 || port > 65535 {
		return errExit(fmt.Errorf("--management-port %d must be 1-65535", port))
	}

	command := statusCommand(platform, commandFlagVal, e)
	cmdArgv, cerr := runner.ParseCommand(platform, command, extraAllowed)
	if cerr != nil {
		return errExit(cerr)
	}

	// Reuse the read-only preflight probe before touching anything, same as
	// deploy/remove. The action argument only steers kubernetes' can-i verb
	// (docker/podman ignore it); status never creates or deletes a deployment,
	// so deploy's "create" is used as the closer of the two existing checks.
	if code, ok := preflight(r, runner.ActionDeploy, platform, command, namespace, extraAllowed); !ok {
		return code
	}

	targets, terr := resolveStatusTargets(platform, pods, containers, namespace, e, r, cmdArgv)
	if terr != nil {
		return errExit(terr)
	}

	script := statusscript.Render(port, user)
	return runStatusOnTargets(r, cmdArgv, platform, namespace, e, targets, script, install)
}

// runStatusOnTargets ensures the status script is present on each target and
// runs it, printing "<target>: " followed by the script's own output (one
// line per continuation, so a multi-line report -- leader-election mode,
// leader-election state, one line per workflow -- stays grouped under its
// target). It never aborts the loop on a target's own non-zero exit: that is
// the script's documented convention (1 standby, 2 error), data to report,
// not a crash. It returns 0 only when every target reached the run step; a
// probe/install failure or a declined install both count toward exit 1.
func runStatusOnTargets(r runner.Runner, cmdArgv []string, platform, namespace string, e *spec.Env, targets []string, script string, install bool) int {
	var present, missing []string
	failed := false

	for _, t := range targets {
		ok, err := runner.ScriptInstalled(r, cmdArgv, platform, t, namespace, statusscript.ContainerPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: error: %v\n", t, err)
			failed = true
			continue
		}
		if ok {
			present = append(present, t)
		} else {
			missing = append(missing, t)
		}
	}

	toRun := present
	if len(missing) > 0 {
		doInstall := install
		if !doInstall {
			var cerr error
			doInstall, cerr = confirmInstall(missing)
			if cerr != nil {
				fmt.Fprintln(os.Stderr, "error:", cerr)
				return 1
			}
		}
		if doInstall {
			for _, t := range missing {
				if _, err := runner.InstallScript(r, cmdArgv, platform, t, namespace, statusscript.ContainerDir, statusscript.ContainerPath, script); err != nil {
					fmt.Fprintf(os.Stderr, "%s: error installing status script: %v\n", t, err)
					failed = true
					continue
				}
				toRun = append(toRun, t)
			}
		} else {
			for _, t := range missing {
				fmt.Fprintf(os.Stderr, "%s: skipped (status script not installed, and the install prompt was declined)\n", t)
			}
			failed = true
		}
	}

	// The deployment name exists only on kubernetes, and only when env.yaml was
	// read at all -- explicit --pod targets can run with no section present.
	deployment := ""
	if platform == validate.PlatformKubernetes && e != nil && e.Kubernetes != nil {
		deployment = e.Kubernetes.Deployment.Name
	}
	reported := false
	for _, t := range toRun {
		// The script always exits 0 and puts its findings in the output, so an
		// error here is the exec failing rather than a standby instance -- report
		// it and keep going, so one unreachable target does not hide the rest.
		out, err := runner.RunStatusScript(r, cmdArgv, platform, t, namespace, statusscript.ContainerPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: error running the status script: %v\n", t, err)
			failed = true
			continue
		}
		// Only docker has a group above the container, and it is read back from
		// the container rather than assumed from the compose file's directory --
		// one extra read-only inspect per target, skipped on the other platforms.
		group := ""
		if platform == validate.PlatformDocker {
			group = runner.DockerComposeProject(r, cmdArgv, t)
		}
		if reported {
			fmt.Println()
		}
		printTargetReport(statusBanner(platform, namespace, deployment, group, t), out)
		reported = true
	}

	if failed {
		return 1
	}
	return 0
}

// statusBanner is the one-line identity printed above a target's report: the
// platform, then the names that locate that instance on it -- namespace /
// deployment / pod on kubernetes, compose project / container on docker, the
// container alone on podman. None of it can come from the report itself: the
// script runs inside the container and knows nothing of what surrounds it.
//
// A name that is not set (no namespace configured, a container compose did not
// create) is dropped rather than rendered as an empty segment, so a separator
// always sits between two real names.
func statusBanner(platform, namespace, deployment, group, target string) string {
	parts := make([]string, 0, 4)
	for _, s := range []string{namespace, deployment, group, target} {
		if s != "" {
			parts = append(parts, s)
		}
	}
	return platform + "  " + strings.Join(parts, " / ")
}

// printTargetReport prints one target's report under its banner, indenting
// every line by two. The identity is a banner rather than a lead-in on the
// first line so the report lines stay left-aligned with each other however
// long the pod name is.
func printTargetReport(banner, out string) {
	fmt.Printf("=== %s ===\n", banner)
	body := strings.TrimRight(out, "\n")
	if body == "" {
		return
	}
	for _, l := range strings.Split(body, "\n") {
		fmt.Printf("  %s\n", l)
	}
}

// confirmInstall asks once whether to install the status script on the listed
// missing targets, via the same promptLine seam and non-TTY refusal
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
	fs := flag.NewFlagSet("validate", flag.ContinueOnError)
	env := envFlag(fs)
	if _, err := collectFlagsAndDirs(fs, args); err != nil {
		return 2
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
	fs := flag.NewFlagSet("examples", flag.ContinueOnError)
	force := fs.Bool("f", false, "overwrite existing files")
	fs.BoolVar(force, "force", false, "overwrite existing files")
	pos, err := collectFlagsAndDirs(fs, args)
	if err != nil {
		return 2
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
	fs := flag.NewFlagSet("auto-complete", flag.ContinueOnError)
	pos, err := collectFlagsAndDirs(fs, args)
	if err != nil {
		return 2
	}
	if len(pos) == 0 {
		fmt.Fprintf(os.Stderr, "auto-complete: missing shell (%s)\n", pipeList(targetNames("auto-complete")))
		usage()
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

func usage() {
	fmt.Fprint(os.Stderr, usageText())
}
