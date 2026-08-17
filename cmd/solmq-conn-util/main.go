// Command solmq-conn-util generates and deploys the Solace PubSub+ Connector for IBM
// MQ from a single env.yaml: it consolidates a folder of per-workflow YAML files
// into an application.yml and, per target, into Kubernetes manifests, a docker
// compose file, or podman run/quadlet units -- and can apply or tear those down
// by shelling out to kubectl/oc, docker, or podman/systemctl.
//
//	solmq-conn-util generate config|kubernetes|docker|podman [-e env.yaml] [-o out]
//	solmq-conn-util deploy   kubernetes|docker|podman        [-e env.yaml]
//	solmq-conn-util delete   kubernetes|docker|podman        [-e env.yaml]
//	solmq-conn-util validate                                 [-e env.yaml]
//	solmq-conn-util examples [dir] [-f]
//	solmq-conn-util completion bash|zsh|fish|powershell
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/examples"
	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/gen"
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
	"generate":   func(args []string, r runner.Runner) int { return runGenerate(args) },
	"deploy":     func(args []string, r runner.Runner) int { return runAction(runner.ActionDeploy, args, r) },
	"delete":     func(args []string, r runner.Runner) int { return runAction(runner.ActionDelete, args, r) },
	"validate":   func(args []string, r runner.Runner) int { return runValidate(args) },
	"examples":   func(args []string, r runner.Runner) int { return runExamples(args) },
	"completion": func(args []string, r runner.Runner) int { return runCompletion(args) },
	"help":       func(args []string, r runner.Runner) int { usage(); return 0 },
}

// dispatch resolves args[0] against verbHandlers. -h/--help are aliases of the
// modeled "help" verb, handled before the lookup since they are not spellable
// as args[0] in the model itself.
func dispatch(args []string, r runner.Runner) int {
	if len(args) == 0 {
		usage()
		return 2
	}
	if args[0] == "-h" || args[0] == "--help" {
		usage()
		return 0
	}
	h, ok := verbHandlers[args[0]]
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

// genTargets maps each modeled "generate" target (cliVerbs in commands.go) to
// its renderer, so runGenerate's accepted set can never drift from the model --
// see TestDispatchHandlersMatchModel.
var genTargets = map[string]func(envPath, out string) int{
	tgtConfig:     genConfig,
	tgtKubernetes: genKubernetes,
	tgtDocker:     genDocker,
	tgtPodman:     genPodman,
}

func runGenerate(args []string) int {
	fs := flag.NewFlagSet("generate", flag.ContinueOnError)
	env := envFlag(fs)
	out := outFlag(fs)
	pos, err := collectFlagsAndDirs(fs, args)
	if err != nil {
		return 2
	}
	if len(pos) == 0 {
		fmt.Fprintf(os.Stderr, "generate: missing target (%s)\n", pipeList(targetNames("generate")))
		usage()
		return 2
	}
	h, ok := genTargets[pos[0]]
	if !ok {
		fmt.Fprintf(os.Stderr, "generate: unknown target %q (want %s)\n", pos[0], wantList(targetNames("generate")))
		return 2
	}
	return h(*env, *out)
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

// ---- deploy / delete ---------------------------------------------------------

// actTargets maps each modeled deploy/delete platform (cliVerbs in
// commands.go) to its implementation, so runAction's accepted set can never
// drift from the model -- see TestDispatchHandlersMatchModel. extraAllowed
// carries the values of a repeatable --allow-command flag (nil when unused).
var actTargets = map[string]func(action, envPath string, r runner.Runner, extraAllowed []string) int{
	tgtKubernetes: actKubernetes,
	tgtDocker:     actDocker,
	tgtPodman:     actPodman,
}

func runAction(action string, args []string, r runner.Runner) int {
	fs := flag.NewFlagSet(action, flag.ContinueOnError)
	env := envFlag(fs)
	allow := allowCommandFlag(fs)
	pos, err := collectFlagsAndDirs(fs, args)
	if err != nil {
		return 2
	}
	if len(pos) == 0 {
		fmt.Fprintf(os.Stderr, "%s: missing platform (%s)\n", action, pipeList(targetNames(action)))
		usage()
		return 2
	}
	h, ok := actTargets[pos[0]]
	if !ok {
		fmt.Fprintf(os.Stderr, "%s: unknown platform %q (want %s)\n", action, pos[0], wantList(targetNames(action)))
		return 2
	}
	return h(action, *env, r, *allow)
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
	// The compose file is regenerated by every deploy and delete, so it is
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
	// deploy/delete are always quadlet; BaseDir bakes absolute on-disk paths into
	// the units so systemd resolves them regardless of cwd.
	plan, errs, warns := gen.GeneratePodman(req, res, gen.PodmanOpts{ForceQuadlet: true, BaseDir: sc.Dir}, extraAllowed...)
	printWarnings(warns)
	if len(errs) > 0 {
		return failFast(errs)
	}
	if code, ok := preflight(r, action, validate.PlatformPodman, e.Podman.Command, "", extraAllowed); !ok {
		return code
	}
	if action == runner.ActionDelete {
		return podmanDelete(sc, plan, e.Podman, r, extraAllowed)
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
	// application.yml and the unit carry only stable secret names, never values,
	// so they are readable by the container user and by systemd (0644).
	if err := runner.WriteFile(filepath.Join(sc.Dir, plan.AppYAML.Name), plan.AppYAML.Data, 0o644); err != nil {
		return errExit(err)
	}
	if err := runner.WriteFile(filepath.Join(sc.Dir, plan.Unit.Filename), plan.Unit.Content, 0o644); err != nil {
		return errExit(err)
	}
	out, rerr := runner.PodmanDeploy(r, sc, []string{plan.Service})
	return report(runner.ActionDeploy, tgtPodman, out, rerr)
}

func podmanDelete(sc runner.QuadletScope, plan gen.PodmanPlan, p *spec.Podman, r runner.Runner, extraAllowed []string) int {
	out, rerr := runner.PodmanDelete(r, sc, []string{plan.Service}, []string{plan.Unit.Filename})
	// Best-effort cleanup of the file we generated.
	_ = os.Remove(filepath.Join(sc.Dir, plan.AppYAML.Name))
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
	return report(runner.ActionDelete, tgtPodman, out, rerr)
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

// ---- completion --------------------------------------------------------------

// completionShellRenderers maps each modeled "completion" shell (cliVerbs in
// commands.go) to its renderer, so runCompletion's accepted set can never drift
// from the model -- see TestDispatchHandlersMatchModel.
var completionShellRenderers = map[string]func() string{
	"bash":       renderBashCompletion,
	"zsh":        renderZshCompletion,
	"fish":       renderFishCompletion,
	"powershell": renderPowerShellCompletion,
}

// runCompletion prints the named shell's completion script on stdout, for the
// caller to redirect or source. Nothing is read or written beyond that: the
// script is rendered from the compiled-in command model, so it needs no env.yaml
// and matches the binary that printed it.
func runCompletion(args []string) int {
	fs := flag.NewFlagSet("completion", flag.ContinueOnError)
	pos, err := collectFlagsAndDirs(fs, args)
	if err != nil {
		return 2
	}
	if len(pos) == 0 {
		fmt.Fprintf(os.Stderr, "completion: missing shell (%s)\n", pipeList(targetNames("completion")))
		usage()
		return 2
	}
	h, ok := completionShellRenderers[pos[0]]
	if !ok {
		fmt.Fprintf(os.Stderr, "completion: unknown shell %q (want %s)\n", pos[0], wantList(targetNames("completion")))
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

// allowCommandFlag registers the repeatable --allow-command flag (deploy/delete
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
