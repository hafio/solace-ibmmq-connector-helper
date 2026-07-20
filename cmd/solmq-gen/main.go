// Command solmq-gen consolidates a folder of per-workflow YAML files into one
// application.yml for the Solace PubSub+ Connector for IBM MQ, and can emit the
// Kubernetes manifests that run it.
//
//	solmq-gen config   <dir> [-o out]
//	solmq-gen deploy   <dir> [-k kube] [-o out]
//	solmq-gen validate <dir>
//	solmq-gen examples [dir] [-f]
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-gen/internal/examples"
	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-gen/internal/gen"
	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-gen/internal/scan"
)

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	if len(args) == 0 {
		usage()
		return 2
	}
	sub := args[0]
	rest := args[1:]

	switch sub {
	case "config":
		return runConfig(rest)
	case "deploy":
		return runDeploy(rest)
	case "validate":
		return runValidate(rest)
	case "examples":
		return runExamples(rest)
	case "-h", "--help", "help":
		usage()
		return 0
	default:
		fmt.Fprintf(os.Stderr, "unknown subcommand %q\n\n", sub)
		usage()
		return 2
	}
}

// filterFlag registers the shared -f/--filter workflow glob on fs.
func filterFlag(fs *flag.FlagSet) *string {
	const usage = "only include workflow files whose base name matches this glob (e.g. 'workflow*.yaml')"
	filter := fs.String("f", "", usage)
	fs.StringVar(filter, "filter", "", usage)
	return filter
}

func runConfig(args []string) int {
	fs := flag.NewFlagSet("config", flag.ContinueOnError)
	out := fs.String("o", "", "write output to a file (default: stdout)")
	fs.StringVar(out, "output", "", "write output to a file (default: stdout)")
	filter := filterFlag(fs)
	dir, code := parseArgs(fs, args)
	if code != 0 {
		return code
	}
	req, err := loadRequest(dir, "", *out, *filter)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		return 1
	}
	req.Kubernetes = nil // config never reads kubernetes.yaml
	res := gen.Resolver{Env: os.LookupEnv, ReadFile: fileReader(dir)}
	docs, errs, warns := gen.Config(req, res)
	printWarnings(warns)
	if len(errs) > 0 {
		return failFast(errs) // config: stop at the first error
	}
	return emitConfigs(*out, docs)
}

func runDeploy(args []string) int {
	fs := flag.NewFlagSet("deploy", flag.ContinueOnError)
	out := fs.String("o", "", "write output to a file (default: stdout)")
	fs.StringVar(out, "output", "", "write output to a file (default: stdout)")
	kube := fs.String("k", "kubernetes.yaml", "Kubernetes settings file (always excluded from the scan)")
	fs.StringVar(kube, "kube", "kubernetes.yaml", "Kubernetes settings file")
	filter := filterFlag(fs)
	dir, code := parseArgs(fs, args)
	if code != 0 {
		return code
	}
	req, err := loadRequest(dir, *kube, *out, *filter)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		return 1
	}
	res := gen.Resolver{Env: os.LookupEnv, ReadFile: fileReader(dir)}
	manifests, errs, warns := gen.Deploy(req, res)
	printWarnings(warns)
	if len(errs) > 0 {
		return failFast(errs) // deploy: stop at the first error
	}
	return emit(*out, manifests)
}

func runValidate(args []string) int {
	fs := flag.NewFlagSet("validate", flag.ContinueOnError)
	kube := fs.String("k", "kubernetes.yaml", "Kubernetes settings file (always excluded from the scan)")
	fs.StringVar(kube, "kube", "kubernetes.yaml", "Kubernetes settings file")
	filter := filterFlag(fs)
	dir, code := parseArgs(fs, args)
	if code != 0 {
		return code
	}
	req, err := loadRequest(dir, *kube, "", *filter)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		return 1
	}
	res := gen.Resolver{Env: os.LookupEnv, ReadFile: fileReader(dir)}
	errs, warns := gen.Validate(req, res)
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

// runExamples writes the embedded sample spec files to dir (default "examples").
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
		fmt.Fprintln(os.Stderr, "error:", werr)
		return 1
	}
	for _, p := range written {
		fmt.Fprintln(os.Stderr, "wrote:", p)
	}
	for _, p := range skipped {
		fmt.Fprintln(os.Stderr, "exists (use -f to overwrite):", p)
	}
	fmt.Fprintf(os.Stderr, "\n%d written, %d skipped in %q\n", len(written), len(skipped), dir)
	fmt.Fprintf(os.Stderr, "next: solmq-gen config %s\n", dir)
	return 0
}

// ---- helpers -----------------------------------------------------------------

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

// parseArgs returns the single required positional <dir>, accepting flags in any
// position. On a parse error or missing <dir> it prints to stderr and returns 2.
func parseArgs(fs *flag.FlagSet, args []string) (string, int) {
	pos, err := collectFlagsAndDirs(fs, args)
	if err != nil {
		return "", 2
	}
	if len(pos) == 0 {
		fmt.Fprintf(os.Stderr, "%s: missing <dir> argument\n", fs.Name())
		return "", 2
	}
	return pos[0], 0
}

// loadRequest scans dir and reads the workflow/defaults/kubernetes files into a
// gen.Request. kubeFile "" uses the default name; filter "" includes every
// workflow file, otherwise it is a glob the base name must match.
func loadRequest(dir, kubeFile, outFile, filter string) (gen.Request, error) {
	res, err := scan.Scan(dir, kubeFile, outFile, filter)
	if err != nil {
		return gen.Request{}, err
	}
	var req gen.Request
	for _, p := range res.WorkflowFiles {
		data, err := os.ReadFile(p)
		if err != nil {
			return gen.Request{}, err
		}
		req.Workflows = append(req.Workflows, gen.File{Name: filepath.Base(p), Data: data})
	}
	if res.DefaultsPath != "" {
		data, err := os.ReadFile(res.DefaultsPath)
		if err != nil {
			return gen.Request{}, err
		}
		req.Defaults = &gen.File{Name: "defaults.yaml", Data: data}
	}
	if res.KubernetesPath != "" {
		data, err := os.ReadFile(res.KubernetesPath)
		if err != nil {
			return gen.Request{}, err
		}
		req.Kubernetes = &gen.File{Name: filepath.Base(res.KubernetesPath), Data: data}
	}
	return req, nil
}

// fileReader resolves values-file / .jks paths relative to the spec dir.
func fileReader(dir string) func(string) ([]byte, error) {
	return func(p string) ([]byte, error) {
		if !filepath.IsAbs(p) {
			p = filepath.Join(dir, p)
		}
		return os.ReadFile(p)
	}
}

func emit(out, content string) int {
	if out == "" {
		fmt.Print(content)
		return 0
	}
	if err := os.WriteFile(out, []byte(content), 0o600); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		return 1
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

// instanceBanner is the 5-line separator printed before each instance's config
// when several are streamed to stdout.
func instanceBanner(n, total int) string {
	bar := "# " + strings.Repeat("=", 60)
	return fmt.Sprintf("%s\n%s\n#  CONNECTOR INSTANCE %d OF %d  —  application.yml\n%s\n%s\n", bar, bar, n, total, bar, bar)
}

// suffixBeforeExt inserts -<n> before the file extension: out.yml -> out-1.yml,
// out -> out-1.
func suffixBeforeExt(path string, n int) string {
	ext := filepath.Ext(path)
	return fmt.Sprintf("%s-%d%s", strings.TrimSuffix(path, ext), n, ext)
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
	fmt.Fprint(os.Stderr, `solmq-gen — Solace IBM MQ Connector config generator

Usage:
  solmq-gen config   <dir> [-o out] [-f glob]            Emit application.yml (fails fast)
  solmq-gen deploy   <dir> [-k kube] [-o out] [-f glob]  Emit ConfigMap+Deployment+Service (+Secrets)
  solmq-gen validate <dir> [-k kube] [-f glob]           Lint only; report every error
  solmq-gen examples [dir] [-f]                          Write sample spec files (default dir: examples)

Flags:
  -o, --out     Output file (default: stdout)
  -k, --kube    Kubernetes settings file (default: kubernetes.yaml)
  -f, --filter  Only include workflow files matching this glob, e.g. 'workflow*.yaml'
                (config/deploy/validate). For examples, -f/--force overwrites existing files.

Reserved files (never treated as workflows): defaults.yaml, the -k file, the -o file.
`)
}
