package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRunExamplesGeneratesAndConfigs(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "ex")
	if code := run([]string{"examples", dir}); code != 0 {
		t.Fatalf("examples exit=%d", code)
	}
	for _, n := range []string{"workflow-0.yaml", "workflow-1.yaml", "workflow-2.yaml", "workflow-3.yaml", "defaults.yaml", "kubernetes.yaml"} {
		if _, err := os.Stat(filepath.Join(dir, n)); err != nil {
			t.Errorf("missing example %s: %v", n, err)
		}
	}
	// The shipped examples must always generate an application.yml with no errors.
	if code := run([]string{"config", dir}); code != 0 {
		t.Fatalf("config on generated examples exit=%d", code)
	}
}

func TestRunExamplesSkipsThenForces(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "ex")
	if code := run([]string{"examples", dir}); code != 0 {
		t.Fatalf("first exit=%d", code)
	}
	f := filepath.Join(dir, "defaults.yaml")
	if err := os.WriteFile(f, []byte("touched\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if code := run([]string{"examples", dir}); code != 0 { // no -f: must skip existing
		t.Fatalf("re-run exit=%d", code)
	}
	if b, _ := os.ReadFile(f); string(b) != "touched\n" {
		t.Error("existing file should be skipped without -f")
	}
	if code := run([]string{"examples", "-f", dir}); code != 0 { // -f before dir too
		t.Fatalf("force exit=%d", code)
	}
	if b, _ := os.ReadFile(f); string(b) == "touched\n" {
		t.Error("-f should overwrite the existing file")
	}
}

func TestRunExamplesDefaultDir(t *testing.T) {
	tmp := t.TempDir()
	wd, _ := os.Getwd()
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(wd)
	if code := run([]string{"examples"}); code != 0 {
		t.Fatalf("default-dir exit=%d", code)
	}
	if _, err := os.Stat(filepath.Join(tmp, "examples", "workflow-0.yaml")); err != nil {
		t.Errorf("default dir 'examples' not created: %v", err)
	}
}
