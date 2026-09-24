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

func TestScanSortsYAMLOnly(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "20.yaml", "x")
	write(t, dir, "10.yml", "x")
	write(t, dir, "notes.txt", "x") // non-YAML ignored
	if err := os.Mkdir(filepath.Join(dir, "sub"), 0o755); err != nil {
		t.Fatal(err)
	}
	res, err := Scan(dir, "*", "")
	if err != nil {
		t.Fatal(err)
	}
	got := bases(res.WorkflowFiles)
	if len(got) != 2 || got[0] != "10.yml" || got[1] != "20.yaml" {
		t.Fatalf("workflows = %v, want [10.yml 20.yaml]", got)
	}
	if res.Dir != dir {
		t.Errorf("dir = %q", res.Dir)
	}
}

// TestScanSortsNumericallyNotLexically pins the ordering operators actually
// number files by: 2 before 9 before 10 before 19. Lexical order would give
// 10, 19, 2, 9 -- and since a workflow's id is its position in this slice,
// that would hand the files each other's ids.
func TestScanSortsNumericallyNotLexically(t *testing.T) {
	dir := t.TempDir()
	for _, n := range []string{"19.yaml", "2.yaml", "10.yaml", "9.yaml"} {
		write(t, dir, n, "x")
	}
	res, err := Scan(dir, "*", "")
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"2.yaml", "9.yaml", "10.yaml", "19.yaml"}
	got := bases(res.WorkflowFiles)
	if len(got) != len(want) {
		t.Fatalf("workflows = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("workflows = %v, want %v", got, want)
		}
	}
}

// TestScanSortsPrefixedNamesNumerically covers the same ordering on the naming
// the examples verb writes (workflow-0.yaml, workflow-1.yaml, ...), where the
// digits sit after a shared prefix rather than at the start of the name.
func TestScanSortsPrefixedNamesNumerically(t *testing.T) {
	dir := t.TempDir()
	for _, n := range []string{"workflow-10.yaml", "workflow-2.yaml", "workflow-1.yaml"} {
		write(t, dir, n, "x")
	}
	res, err := Scan(dir, "*", "")
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"workflow-1.yaml", "workflow-2.yaml", "workflow-10.yaml"}
	got := bases(res.WorkflowFiles)
	if len(got) != len(want) {
		t.Fatalf("workflows = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("workflows = %v, want %v", got, want)
		}
	}
}

func TestScanExcludesEnvFile(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "env.yaml", "x")
	write(t, dir, "workflow-0.yaml", "x")
	envAbs, _ := filepath.Abs(filepath.Join(dir, "env.yaml"))
	res, err := Scan(dir, "*", envAbs)
	if err != nil {
		t.Fatal(err)
	}
	got := bases(res.WorkflowFiles)
	if len(got) != 1 || got[0] != "workflow-0.yaml" {
		t.Fatalf("env.yaml must be excluded even under '*': %v", got)
	}
}

func TestScanEnvFileExcludedRegardlessOfPattern(t *testing.T) {
	// A pattern that would otherwise match env.yaml must still exclude it.
	dir := t.TempDir()
	write(t, dir, "env.yaml", "x")
	write(t, dir, "envoy.yaml", "x")
	envAbs, _ := filepath.Abs(filepath.Join(dir, "env.yaml"))
	res, err := Scan(dir, "env*", envAbs)
	if err != nil {
		t.Fatal(err)
	}
	got := bases(res.WorkflowFiles)
	if len(got) != 1 || got[0] != "envoy.yaml" {
		t.Fatalf("only envoy.yaml should remain: %v", got)
	}
}

func TestScanEmptyPatternDefaultsToStar(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "a.yaml", "x")
	res, err := Scan(dir, "", "")
	if err != nil {
		t.Fatal(err)
	}
	if len(res.WorkflowFiles) != 1 {
		t.Errorf("empty pattern should behave as '*': %v", bases(res.WorkflowFiles))
	}
}

func TestScanErrorMissingDir(t *testing.T) {
	if _, err := Scan(filepath.Join(t.TempDir(), "nope"), "*", ""); err == nil {
		t.Fatal("expected error for missing dir")
	}
}

func TestScanPatternWildcards(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "workflow-0.yaml", "x")
	write(t, dir, "workflow-1.yaml", "x")
	write(t, dir, "adhoc.yaml", "x")

	// Trailing star.
	res, err := Scan(dir, "workflow-*", "")
	if err != nil {
		t.Fatal(err)
	}
	got := bases(res.WorkflowFiles)
	if len(got) != 2 || got[0] != "workflow-0.yaml" || got[1] != "workflow-1.yaml" {
		t.Fatalf("workflow-* = %v, want the two workflow files", got)
	}

	// Mid-string star.
	if res, _ := Scan(dir, "*hoc*", ""); len(res.WorkflowFiles) != 1 || filepath.Base(res.WorkflowFiles[0]) != "adhoc.yaml" {
		t.Errorf("*hoc* should match only adhoc.yaml, got %v", bases(res.WorkflowFiles))
	}

	// Leading star.
	if res, _ := Scan(dir, "*-1.yaml", ""); len(res.WorkflowFiles) != 1 || filepath.Base(res.WorkflowFiles[0]) != "workflow-1.yaml" {
		t.Errorf("*-1.yaml should match only workflow-1.yaml, got %v", bases(res.WorkflowFiles))
	}
}

func TestScanPatternNoMatchIsEmptyNotError(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "workflow-0.yaml", "x")
	res, err := Scan(dir, "nope*", "")
	if err != nil {
		t.Fatalf("a non-matching pattern is not an error: %v", err)
	}
	if len(res.WorkflowFiles) != 0 {
		t.Errorf("want no matches, got %v", bases(res.WorkflowFiles))
	}
}

func TestScanRejectsNonStarMetachars(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "workflow-0.yaml", "x")
	for _, pat := range []string{"[bad", "wf?.yaml", "a]b", `a\b`} {
		if _, err := Scan(dir, pat, ""); err == nil {
			t.Errorf("pattern %q should be rejected (only '*' allowed)", pat)
		}
	}
}

func TestMatchStar(t *testing.T) {
	cases := []struct {
		pattern, name string
		want          bool
	}{
		{"*", "anything.yaml", true},
		{"exact.yaml", "exact.yaml", true},
		{"exact.yaml", "other.yaml", false},
		{"pre*", "prefix.yaml", true},
		{"*.yaml", "x.yaml", true},
		{"*.yaml", "x.yml", false},
		{"a*b*c", "axxbyyc", true},
		{"a*b*c", "axxc", false},
		{"**", "anything", true},
	}
	for _, c := range cases {
		if got := matchStar(c.pattern, c.name); got != c.want {
			t.Errorf("matchStar(%q,%q)=%v want %v", c.pattern, c.name, got, c.want)
		}
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
