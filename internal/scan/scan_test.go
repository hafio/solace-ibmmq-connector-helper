package scan

import (
	"os"
	"path/filepath"
	"testing"
)

func write(t *testing.T, dir, name, body string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
}

func bases(paths []string) []string {
	out := make([]string, len(paths))
	for i, p := range paths {
		out[i] = filepath.Base(p)
	}
	return out
}

func TestScanClassifies(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "20.yaml", "x")
	write(t, dir, "10.yml", "x")
	write(t, dir, "defaults.yaml", "x")
	write(t, dir, "kubernetes.yaml", "x")
	write(t, dir, "notes.txt", "x")
	write(t, dir, "out.yaml", "x")
	if err := os.Mkdir(filepath.Join(dir, "sub"), 0o755); err != nil {
		t.Fatal(err)
	}
	res, err := Scan(dir, "kubernetes.yaml", filepath.Join(dir, "out.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	got := bases(res.WorkflowFiles)
	if len(got) != 2 || got[0] != "10.yml" || got[1] != "20.yaml" {
		t.Fatalf("workflows = %v, want [10.yml 20.yaml]", got)
	}
	if filepath.Base(res.DefaultsPath) != "defaults.yaml" {
		t.Errorf("defaults = %q", res.DefaultsPath)
	}
	if filepath.Base(res.KubernetesPath) != "kubernetes.yaml" {
		t.Errorf("kube = %q", res.KubernetesPath)
	}
	if res.Dir != dir {
		t.Errorf("dir = %q", res.Dir)
	}
}

func TestScanCustomKubeNoDefaults(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "a.yaml", "x")
	write(t, dir, "myk8s.yaml", "x")
	res, err := Scan(dir, "myk8s.yaml", "")
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Base(res.KubernetesPath) != "myk8s.yaml" {
		t.Errorf("custom kube: %q", res.KubernetesPath)
	}
	if res.DefaultsPath != "" {
		t.Errorf("expected no defaults: %q", res.DefaultsPath)
	}
	if len(res.WorkflowFiles) != 1 || filepath.Base(res.WorkflowFiles[0]) != "a.yaml" {
		t.Errorf("workflows = %v", bases(res.WorkflowFiles))
	}
}

func TestScanOutFileOutsideDirNotExcluded(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "a.yaml", "x")
	res, err := Scan(dir, "kubernetes.yaml", filepath.Join(t.TempDir(), "a.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if len(res.WorkflowFiles) != 1 {
		t.Errorf("out file outside dir must not be excluded: %v", bases(res.WorkflowFiles))
	}
}

func TestScanEmptyKubeDefaultsToKubernetesYAML(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "kubernetes.yaml", "x")
	res, err := Scan(dir, "", "")
	if err != nil {
		t.Fatal(err)
	}
	if res.KubernetesPath == "" {
		t.Error("empty kubeFile should default to kubernetes.yaml")
	}
}

func TestScanErrorMissingDir(t *testing.T) {
	if _, err := Scan(filepath.Join(t.TempDir(), "nope"), "kubernetes.yaml", ""); err == nil {
		t.Fatal("expected error for missing dir")
	}
}

func TestIsYAML(t *testing.T) {
	for name, want := range map[string]bool{
		"a.yaml": true, "a.yml": true, "A.YAML": true,
		"a.txt": false, "yaml": false, "a.yamlx": false,
	} {
		if got := isYAML(name); got != want {
			t.Errorf("isYAML(%q)=%v want %v", name, got, want)
		}
	}
}
