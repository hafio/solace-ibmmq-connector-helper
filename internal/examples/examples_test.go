package examples

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWriteCreatesSkipsForces(t *testing.T) {
	dir := t.TempDir()
	w, s, err := Write(dir, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(w) != 6 || len(s) != 0 {
		t.Fatalf("first write: %d written, %d skipped", len(w), len(s))
	}
	for _, want := range []string{"workflow-0.yaml", "workflow-1.yaml", "workflow-2.yaml", "workflow-3.yaml", "defaults.yaml", "kubernetes.yaml"} {
		if _, err := os.Stat(filepath.Join(dir, want)); err != nil {
			t.Errorf("missing %s", want)
		}
	}
	// second run, no force: everything already there -> all skipped
	w2, s2, err := Write(dir, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(w2) != 0 || len(s2) != len(w) {
		t.Fatalf("skip run: %d written, %d skipped", len(w2), len(s2))
	}
	// force: all rewritten
	w3, s3, err := Write(dir, true)
	if err != nil {
		t.Fatal(err)
	}
	if len(w3) != len(w) || len(s3) != 0 {
		t.Fatalf("force run: %d written, %d skipped", len(w3), len(s3))
	}
	if b, _ := os.ReadFile(filepath.Join(dir, "workflow-0.yaml")); len(b) == 0 {
		t.Error("example file is empty")
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
