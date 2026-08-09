// Command solmq-conn generates and deploys the Solace PubSub+ Connector for IBM
// MQ from a single env.yaml: it consolidates a folder of per-workflow YAML files
// into an application.yml and, per target, into Kubernetes manifests, a docker
// compose file, or podman run/quadlet units -- and can apply or tear those down
// by shelling out to kubectl/oc, docker, or podman/systemctl.
//
//	solmq-conn generate config|kubernetes|docker|podman [-e env.yaml] [-o out]
//	solmq-conn deploy   kubernetes|docker|podman        [-e env.yaml]
//	solmq-conn delete   kubernetes|docker|podman        [-e env.yaml]
//	solmq-conn validate                                 [-e env.yaml]
//	solmq-conn examples [dir] [-f]
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/examples"
	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/gen"
	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/podmangen"
	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/runner"
	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/scan"
	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/spec"
)

const (
	defaultEnvFile = "env.yaml"
	composeFile    = "docker-compose.yml"

	tgtConfig     = "config"
	tgtKubernetes = "kubernetes"
	tgtDocker     = "docker"
	tgtPodman     = "podman"
)

// newRunner builds the Runner deploy/delete actions run through. Tests swap it
// for a fake that records argv instead of starting a process.
var newRunner = func() runner.Runner { return runner.OS{} }

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	if len(args) == 0 {
		usage()
		return 2
	}
	switch args[0] {
	case "generate":
		return runGenerate(args[1:])
	case "deploy":
		return runAction(runner.ActionDeploy, args[1:])
	case "delete":
		return runAction(runner.ActionDelete, args[1:])
	case "validate":
		return runValidate(args[1:])
	case "examples":
		return runExamples(args[1:])
	case "-h", "--help", "help":
		usage()
		return 0
	default:
		fmt.Fprintf(os.Stderr, "unknown command %q\n\n", args[0])
		usage()
		return 2
	}
}

// ---- generate ----------------------------------------------------------------

func runGenerate(args []string) int {
	fs := flag.NewFlagSet("generate", flag.ContinueOnError)
	env := envFlag(fs)
	out := outFlag(fs)
	pos, err := collectFlagsAndDirs(fs, args)
	if err != nil {
		return 2
	}
	if len(pos) == 0 {
		fmt.Fprintln(os.Stderr, "generate: missing target (config|kubernetes|docker|podman)")
		usage()
		return 2
	}
	switch pos[0] {
	case tgtConfig:
		return genConfig(*env, *out)
	case tgtKubernetes:
		return genKubernetes(*env, *out)
	case tgtDocker:
		return genDocker(*env, *out)
	case tgtPodman:
		return genPodman(*env, *out)
	default:
		fmt.Fprintf(os.Stderr, "generate: unknown target %q (want config, kubernetes, docker, or podman)\n", pos[0])
		return 2
	}
}

func genConfig(envPath, out string) int {
	req, _, envDir, err := loadEnv(envPath)
	if err != nil {
		return errExit(err)
	}
	docs, errs, warns := gen.Config(req, resolver(envDir))
	printWarnings(warns)
	if len(errs) > 0 {
		return failFast(errs) // config: stop at the first error
	}
	return emitConfigs(out, docs)
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
		return emit(out, joinUnits(plan.Units))
	}
	return emit(out, plan.RunScript)
}

// joinUnits concatenates quadlet units into one previewable document, each
// preceded by a filename banner so `generate podman` (quadlet mode) shows every
// unit at once. Deploy writes them to disk as separate files instead.
func joinUnits(units []podmangen.Unit) string {
	var b strings.Builder
	for i, u := range units {
		if i > 0 {
			b.WriteString("\n")
		}
		b.WriteString("# === " + u.Filename + " ===\n")
		b.WriteString(u.Content)
	}
	return b.String()
}

// ---- deploy / delete ---------------------------------------------------------

func runAction(action string, args []string) int {
	fs := flag.NewFlagSet(action, flag.ContinueOnError)
	env := envFlag(fs)
	pos, err := collectFlagsAndDirs(fs, args)
	if err != nil {
		return 2
	}
	if len(pos) == 0 {
		fmt.Fprintf(os.Stderr, "%s: missing platform (kubernetes|docker|podman)\n", action)
		usage()
		return 2
	}
	switch pos[0] {
	case tgtKubernetes:
		return actKubernetes(action, *env)
	case tgtDocker:
		return actDocker(action, *env)
	case tgtPodman:
		return actPodman(action, *env)
	default:
		fmt.Fprintf(os.Stderr, "%s: unknown platform %q (want kubernetes, docker, or podman)\n", action, pos[0])
		return 2
	}
}

func actKubernetes(action, envPath string) int {
	req, e, envDir, err := loadEnv(envPath)
	if err != nil {
		return errExit(err)
	}
	if e.Kubernetes == nil {
		return errExit(fmt.Errorf("env.yaml has no kubernetes: section to %s", action))
	}
	manifest, errs, warns := gen.GenerateKubernetes(req, resolver(envDir))
	printWarnings(warns)
	if len(errs) > 0 {
		return failFast(errs)
	}
	out, rerr := runner.Kubernetes(newRunner(), e.Kubernetes.Command, action, manifest)
	return report(action, tgtKubernetes, out, rerr)
}

func actDocker(action, envPath string) int {
	req, e, envDir, err := loadEnv(envPath)
	if err != nil {
		return errExit(err)
	}
	if e.Docker == nil {
		return errExit(fmt.Errorf("env.yaml has no docker: section to %s", action))
	}
	res := resolver(envDir)
	plan, errs, warns := gen.GenerateDocker(req, res)
	printWarnings(warns)
	if len(errs) > 0 {
		return failFast(errs)
	}
	path := filepath.Join(envDir, composeFile)
	if werr := runner.WriteFile(path, plan.Compose, 0o644); werr != nil {
		return errExit(werr)
	}
	if werr := writeCredEnvFile(envDir, plan.EnvFileName, e.Docker.Secrets.Credentials, res); werr != nil {
		return errExit(werr)
	}
	out, rerr := runner.Docker(newRunner(), e.Docker.Command, action, path)
	return report(action, tgtDocker, out, rerr)
}

func actPodman(action, envPath string) int {
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
	plan, errs, warns := gen.GeneratePodman(req, res, gen.PodmanOpts{ForceQuadlet: true, BaseDir: sc.Dir})
	printWarnings(warns)
	if len(errs) > 0 {
		return failFast(errs)
	}
	if action == runner.ActionDelete {
		return podmanDelete(sc, plan, e.Podman.Secrets.Credentials)
	}
	return podmanDeploy(sc, plan, e.Podman.Secrets.Credentials, res)
}

func podmanDeploy(sc runner.QuadletScope, plan gen.PodmanPlan, creds *spec.CredentialsSecret, res gen.Resolver) int {
	for _, d := range plan.AppYAMLs {
		if err := runner.WriteFile(filepath.Join(sc.Dir, d.Name), d.Data, 0o644); err != nil {
			return errExit(err)
		}
	}
	if err := writeCredEnvFile(sc.Dir, plan.EnvFileName, creds, res); err != nil {
		return errExit(err)
	}
	for _, u := range plan.Units {
		if err := runner.WriteFile(filepath.Join(sc.Dir, u.Filename), u.Content, 0o644); err != nil {
			return errExit(err)
		}
	}
	out, rerr := runner.PodmanDeploy(newRunner(), sc, plan.Services)
	return report(runner.ActionDeploy, tgtPodman, out, rerr)
}

func podmanDelete(sc runner.QuadletScope, plan gen.PodmanPlan, creds *spec.CredentialsSecret) int {
	units := make([]string, 0, len(plan.Units))
	for _, u := range plan.Units {
		units = append(units, u.Filename)
	}
	out, rerr := runner.PodmanDelete(newRunner(), sc, plan.Services, units)
	// Best-effort cleanup of the files we generated (never a user-supplied
	// existing env-file).
	for _, d := range plan.AppYAMLs {
		_ = os.Remove(filepath.Join(sc.Dir, d.Name))
	}
	if plan.EnvFileName != "" && creds != nil && creds.Create != nil {
		_ = os.Remove(filepath.Join(sc.Dir, plan.EnvFileName))
	}
	return report(runner.ActionDelete, tgtPodman, out, rerr)
}

// writeCredEnvFile resolves a created credentials block and writes its env-file
// (0600, never logged). A nil/existing block writes nothing: an existing env-file
// is the user's to manage.
func writeCredEnvFile(dir, name string, creds *spec.CredentialsSecret, res gen.Resolver) error {
	if name == "" {
		return nil
	}
	kvs, err := gen.ResolveCredentials(creds, res)
	if err != nil {
		return err
	}
	if kvs == nil {
		return nil // existing env-file
	}
	return runner.WriteFile(filepath.Join(dir, name), gen.EnvFileContent(kvs), 0o600)
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
// "examples").
func runExamples(args []string) int {
	fs := flag.NewFlagSet("examples", flag.ContinueOnError)
	force := fs.Bool("f", false, "overwrite existing files")
	fs.BoolVar(force, "force", false, "overwrite existing files")
	pos, err := collectFlagsAndDirs(fs, args)
	if err != nil {
		return 2
	}
	dir := "examples"
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
	fmt.Fprintf(os.Stderr, "next: solmq-conn generate config -e %s\n", filepath.Join(dir, defaultEnvFile))
	return 0
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

// emitConfigs writes the per-instance application.yml documents. A single
// instance behaves exactly like emit (plain stdout / one -o file). With several
// instances, stdout gets a banner before each doc and -o writes <base>-<n><ext>.
func emitConfigs(out string, docs []string) int {
	if len(docs) == 1 {
		return emit(out, docs[0])
	}
	if out == "" {
		for i, doc := range docs {
			fmt.Print(instanceBanner(i+1, len(docs)))
			fmt.Print(doc)
		}
		return 0
	}
	for i, doc := range docs {
		if code := emit(suffixBeforeExt(out, i+1), doc); code != 0 {
			return code
		}
	}
	return 0
}

// instanceBanner is the separator printed before each instance's config when
// several are streamed to stdout.
func instanceBanner(n, total int) string {
	bar := "# " + strings.Repeat("=", 60)
	return fmt.Sprintf("%s\n%s\n#  CONNECTOR INSTANCE %d OF %d  --  application.yml\n%s\n%s\n", bar, bar, n, total, bar, bar)
}

// suffixBeforeExt inserts -<n> before the file extension: out.yml -> out-1.yml,
// out -> out-1.
func suffixBeforeExt(path string, n int) string {
	ext := filepath.Ext(path)
	return fmt.Sprintf("%s-%d%s", strings.TrimSuffix(path, ext), n, ext)
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
