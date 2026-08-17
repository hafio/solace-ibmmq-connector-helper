package examples

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/gen"
	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/scan"
	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/spec"
)

// mustWrite wraps Write with the fatal-on-error boilerplate shared by the
// create/skip/force phases below.
func mustWrite(t *testing.T, dir string, force bool) (written, skipped []string) {
	t.Helper()
	w, s, err := Write(dir, force)
	if err != nil {
		t.Fatal(err)
	}
	return w, s
}

func TestWriteCreatesSkipsForces(t *testing.T) {
	dir := t.TempDir()
	w, s := mustWrite(t, dir, false)
	if len(w) != 5 || len(s) != 0 {
		t.Fatalf("first write: %d written, %d skipped", len(w), len(s))
	}
	for _, want := range []string{"workflow-0.yaml", "workflow-1.yaml", "workflow-2.yaml", "workflow-3.yaml", "env.yaml"} {
		if _, err := os.Stat(filepath.Join(dir, want)); err != nil {
			t.Errorf("missing %s", want)
		}
	}
	// second run, no force: everything already there -> all skipped
	w2, s2 := mustWrite(t, dir, false)
	if len(w2) != 0 || len(s2) != len(w) {
		t.Fatalf("skip run: %d written, %d skipped", len(w2), len(s2))
	}
	// overwrite one output file with junk, then force-rewrite: force must
	// restore the embedded original content, not just leave the path in place.
	if err := os.WriteFile(filepath.Join(dir, "workflow-0.yaml"), []byte("junk, not the real content"), 0o644); err != nil {
		t.Fatal(err)
	}
	want, err := files.ReadFile("files/workflow-0.yaml")
	if err != nil {
		t.Fatal(err)
	}

	// force: all rewritten
	w3, s3 := mustWrite(t, dir, true)
	if len(w3) != len(w) || len(s3) != 0 {
		t.Fatalf("force run: %d written, %d skipped", len(w3), len(s3))
	}
	got, err := os.ReadFile(filepath.Join(dir, "workflow-0.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(want) {
		t.Errorf("force did not restore workflow-0.yaml content: got %q, want %q", got, want)
	}
}

func TestWriteMkdirError(t *testing.T) {
	f := filepath.Join(t.TempDir(), "afile")
	if err := os.WriteFile(f, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, _, err := Write(filepath.Join(f, "sub"), false); err == nil {
		t.Error("expected error creating a dir under a regular file")
	}
}

// TestShippedExamplesGenerateConfig writes the embedded example set to disk and
// drives it through gen.Config, the same entry point cmd/solmq-conn-util's
// `generate config` calls (see main.go's genConfig). This enforces the package
// doc comment's promise (examples.go) -- a freshly written example set generates
// with no errors -- and catches schema drift in the embedded env.yaml
// independently of internal/gen's golden testdata copy, which has already
// diverged from this one in comments/sections.
func TestShippedExamplesGenerateConfig(t *testing.T) {
	dir := t.TempDir()
	if _, _, err := Write(dir, false); err != nil {
		t.Fatal(err)
	}

	envPath := filepath.Join(dir, "env.yaml")
	envData, err := os.ReadFile(envPath)
	if err != nil {
		t.Fatal(err)
	}
	e, err := spec.ParseEnv(envData)
	if err != nil {
		t.Fatal(err)
	}
	wfDir := e.Workflows.Dir
	if !filepath.IsAbs(wfDir) {
		wfDir = filepath.Join(dir, wfDir)
	}
	absEnv, err := filepath.Abs(envPath)
	if err != nil {
		t.Fatal(err)
	}
	sr, err := scan.Scan(wfDir, e.Workflows.FilePattern, absEnv)
	if err != nil {
		t.Fatal(err)
	}

	req := gen.Request{Env: &gen.File{Name: filepath.Base(envPath), Data: envData}}
	for _, p := range sr.WorkflowFiles {
		wd, rerr := os.ReadFile(p)
		if rerr != nil {
			t.Fatal(rerr)
		}
		req.Workflows = append(req.Workflows, gen.File{Name: filepath.Base(p), Data: wd})
	}

	out, errs, _ := gen.Config(req, testResolver(dir))
	if len(errs) != 0 {
		t.Fatalf("shipped examples must generate config with no errors, got: %v", errs)
	}
	if strings.TrimSpace(out) == "" {
		t.Fatal("rendered application.yml is empty")
	}
}

// testResolver builds a gen.Resolver the same way cmd/solmq-conn-util/main.go's
// resolver()/fileReader()/absResolver() do: real env lookups, and file/path
// resolution relative to dir.
func testResolver(dir string) gen.Resolver {
	abs := func(p string) string {
		if filepath.IsAbs(p) {
			return p
		}
		return filepath.Join(dir, p)
	}
	return gen.Resolver{
		Env:      os.LookupEnv,
		ReadFile: func(p string) ([]byte, error) { return os.ReadFile(abs(p)) },
		Abs:      abs,
	}
}
