package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

const testCatalogPath = "../../docs/test.md"

// testCatalogRepoRoot is where the function-count walk starts: two levels up
// from this package is the repository root, the same root a plain
// `grep -rh '^func Test'` would be run from.
const testCatalogRepoRoot = "../.."

// testCatalogSkipDirs names the directories the walk never descends into.
// ".git" holds no source and walking its object store is wasted I/O and can
// hit files WalkDir cannot stat cleanly on Windows; "testdata" and
// "graphify-out" hold no *_test.go files today but are excluded on principle
// so a future fixture file named like a test file cannot inflate the count.
// Nothing else is excluded: this module has no vendor directory, and every
// other top-level directory (dist, eg, libs, scripts, .github, .claude)
// already has no *_test.go files, so this matches what
// `grep -rh --include=*_test.go '^func Test' .` sees from the repo root.
var testCatalogSkipDirs = map[string]bool{
	".git":         true,
	"testdata":     true,
	"graphify-out": true,
}

// testFuncRe is the doc's own documented counting rule, byte for byte: "func
// Test" at the very start of a line. It is deliberately as naive as the doc
// says it is -- it does not check for a *testing.T parameter or an exported
// name past "Test" -- so this gate and the human instruction it enforces stay
// the same formula.
var testFuncRe = regexp.MustCompile(`(?m)^func Test`)

// testCatalogSnapshotRe parses the "_Snapshot: N test functions, M case rows
// across P packages._" line docs/test.md carries just above the first
// package section. It intentionally does not anchor past "packages." so the
// human-authored parenthetical that follows can keep changing wording without
// breaking this test.
var testCatalogSnapshotRe = regexp.MustCompile(`_Snapshot: (\d+) test functions, (\d+) case rows across (\d+) packages\.`)

// TestTestCatalogSnapshotInSync is the drift gate: the three numbers in
// docs/test.md's snapshot line must match what the repo and the doc's own
// tables actually contain. Unlike TestCommandsDocInSync, docs/test.md is
// hand-maintained prose, not a rendered model, so there is no -update flag
// here -- a failure means a human recounts and edits the snapshot line, and
// each error below names which count drifted and how to recompute it.
func TestTestCatalogSnapshotInSync(t *testing.T) {
	path := filepath.FromSlash(testCatalogPath)
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	doc := normLF(string(b))

	wantFuncs, wantRows, wantPackages, err := parseTestCatalogSnapshot(doc)
	if err != nil {
		t.Fatalf("docs/test.md: %v", err)
	}

	gotRows, gotPackages := countDocShape(doc)
	gotFuncs, err := countTestFuncs(filepath.FromSlash(testCatalogRepoRoot))
	if err != nil {
		t.Fatalf("walk %s counting test functions: %v", testCatalogRepoRoot, err)
	}

	if gotFuncs != wantFuncs {
		t.Errorf("docs/test.md snapshot says %d test functions but the repo has %d; "+
			"recount with `grep -rho \"^func Test\" --include=*_test.go . | wc -l` from the repo root "+
			"(excluding testdata/, graphify-out/, .git/) and update the snapshot line",
			wantFuncs, gotFuncs)
	}
	if gotRows != wantRows {
		t.Errorf("docs/test.md snapshot says %d case rows but the doc's tables have %d; "+
			"recount as (pipe-prefixed lines) minus 2 per table (one header row, one separator "+
			"row) and update the snapshot line", wantRows, gotRows)
	}
	if gotPackages != wantPackages {
		t.Errorf("docs/test.md snapshot says %d packages but the doc has %d H2 package "+
			"sections (a `## ` heading starting with internal/ or cmd/); update the snapshot line",
			wantPackages, gotPackages)
	}
}

// parseTestCatalogSnapshot extracts the three counts the snapshot line
// claims, so the gate can compare a claim to a fact instead of parsing the
// claim and the fact the same way and trivially agreeing with itself.
func parseTestCatalogSnapshot(doc string) (funcs, rows, packages int, err error) {
	m := testCatalogSnapshotRe.FindStringSubmatch(doc)
	if m == nil {
		return 0, 0, 0, fmt.Errorf("no line matches %q -- the snapshot sentence was reworded or removed", testCatalogSnapshotRe.String())
	}
	funcs, _ = strconv.Atoi(m[1])
	rows, _ = strconv.Atoi(m[2])
	packages, _ = strconv.Atoi(m[3])
	return funcs, rows, packages, nil
}

// countDocShape derives case rows and package sections straight from the
// doc's own markdown, independent of the snapshot line: case rows are every
// pipe-prefixed line minus the header and separator row each table opens
// with (found by their separator row, since a header row alone looks just
// like a data row); package sections are the H2 headings that name a
// package rather than front matter like "## Contents".
func countDocShape(doc string) (caseRows, packageSections int) {
	pipeLines, tables := 0, 0
	for _, ln := range strings.Split(doc, "\n") {
		switch {
		case strings.HasPrefix(ln, "## "):
			if isPackageHeading(ln) {
				packageSections++
			}
		case strings.HasPrefix(ln, "|"):
			pipeLines++
			if isTableSeparatorRow(ln) {
				tables++
			}
		}
	}
	return pipeLines - 2*tables, packageSections
}

// isPackageHeading reports whether an H2 line names a Go package section
// (the only two top-level source roots this module has) rather than front
// matter such as "## Contents" or "## How the suite is built".
func isPackageHeading(heading string) bool {
	title := strings.TrimPrefix(heading, "## ")
	return strings.HasPrefix(title, "internal/") || strings.HasPrefix(title, "cmd/")
}

// isTableSeparatorRow reports whether a pipe-prefixed line is a markdown
// table's separator row (e.g. "|------|------|----------|") rather than a
// header or data row: every character between the bounding pipes is a dash,
// colon, or pipe, which no real cell content in this doc is.
func isTableSeparatorRow(ln string) bool {
	trimmed := strings.TrimSpace(ln)
	if !strings.HasPrefix(trimmed, "|") || !strings.HasSuffix(trimmed, "|") || len(trimmed) < 2 {
		return false
	}
	for _, r := range trimmed {
		if r != '|' && r != '-' && r != ':' {
			return false
		}
	}
	return true
}

// countTestFuncs walks root and counts lines matching testFuncRe across every
// *_test.go file, skipping testCatalogSkipDirs. It returns an error instead
// of taking a *testing.T so it stays a pure, independently testable count.
func countTestFuncs(root string) (int, error) {
	total := 0
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if testCatalogSkipDirs[d.Name()] {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, "_test.go") {
			return nil
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		total += len(testFuncRe.FindAll(b, -1))
		return nil
	})
	return total, err
}

// TestParseTestCatalogSnapshot covers both the happy path and the missing-line
// case: a reworded or deleted snapshot sentence must fail loudly rather than
// silently comparing zero to zero.
func TestParseTestCatalogSnapshot(t *testing.T) {
	tests := []struct {
		name    string
		doc     string
		want    [3]int
		wantErr bool
	}{
		{
			name: "well-formed line among other prose",
			doc: "some prose above\n" +
				"_Snapshot: 741 test functions, 1009 case rows across 18 packages. " +
				"(a human parenthetical that keeps changing)_\n" +
				"more prose below",
			want: [3]int{741, 1009, 18},
		},
		{
			name:    "no snapshot line present",
			doc:     "this doc has no snapshot sentence at all",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			funcs, rows, packages, err := parseTestCatalogSnapshot(tt.doc)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("parseTestCatalogSnapshot(%q) = %d,%d,%d,nil, want an error", tt.doc, funcs, rows, packages)
				}
				return
			}
			if err != nil {
				t.Fatalf("parseTestCatalogSnapshot(%q) error: %v", tt.doc, err)
			}
			if got := [3]int{funcs, rows, packages}; got != tt.want {
				t.Errorf("parseTestCatalogSnapshot(%q) = %v, want %v", tt.doc, got, tt.want)
			}
		})
	}
}

// TestIsTableSeparatorRow pins the shape that tells a separator row apart
// from the header and data rows it sits between, since a false positive on a
// data row (or a false negative on a separator) would throw off every table's
// contribution to countDocShape by one row.
func TestIsTableSeparatorRow(t *testing.T) {
	tests := []struct {
		name string
		line string
		want bool
	}{
		{"separator row", "|------|------|----------|", true},
		{"header row", "| Test | Case | Verifies |", false},
		{"data row", "| TestFoo | - | does the thing |", false},
		{"data row containing a dash", "| TestFoo | - | a dash-bearing sentence |", false},
		{"not pipe-prefixed", "some prose, not a table row", false},
		{"bare pipe", "|", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isTableSeparatorRow(tt.line); got != tt.want {
				t.Errorf("isTableSeparatorRow(%q) = %v, want %v", tt.line, got, tt.want)
			}
		})
	}
}

// TestCountDocShapeCountsCaseRowsAndPackageSections exercises countDocShape
// against a small hand-built doc rather than the real docs/test.md, so this
// test's expectation does not have to be updated every time a real case row
// is added -- that drift is exactly what TestTestCatalogSnapshotInSync is for.
func TestCountDocShapeCountsCaseRowsAndPackageSections(t *testing.T) {
	doc := strings.Join([]string{
		"# solmq-conn-util test catalogue",
		"",
		"## Contents",
		"- front matter, not a package section",
		"",
		"## internal/scan",
		"| Test | Case | Verifies |",
		"|------|------|----------|",
		"| TestA | - | one |",
		"| TestB | - | two |",
		"",
		"## cmd/solmq-conn-util",
		"| Test | Case | Verifies |",
		"|------|------|----------|",
		"| TestC | - | three |",
		"",
	}, "\n")

	gotRows, gotPackages := countDocShape(doc)
	if gotRows != 3 {
		t.Errorf("caseRows = %d, want 3", gotRows)
	}
	if gotPackages != 2 {
		t.Errorf("packageSections = %d, want 2 (Contents excluded)", gotPackages)
	}
}

// TestCountTestFuncsWalksTreeSkippingDataDirs builds a small fixture tree
// covering every branch countTestFuncs takes: a real match, a non-test .go
// file that must not be scanned even though it contains the same text, an
// indented occurrence that is not a real top-level func declaration, and one
// file under each skipped directory that would inflate the count if the skip
// list were ever dropped.
func TestCountTestFuncsWalksTreeSkippingDataDirs(t *testing.T) {
	root := t.TempDir()
	writeCatalogFixture(t, root, "pkg/a_test.go", "package pkg\n\nfunc TestOne(t *testing.T) {}\n\nfunc TestTwo(t *testing.T) {}\n")
	writeCatalogFixture(t, root, "pkg/helpers.go", "package pkg\n\nfunc TestThree(t *testing.T) {}\n") // wrong suffix ("_test.go" missing): must not be scanned
	writeCatalogFixture(t, root, "pkg/indented_test.go", "package pkg\n\nfunc real() {\n\t// func Test inside a comment, not at column 0\n}\n")
	writeCatalogFixture(t, root, "testdata/skipped_test.go", "func TestShouldNotCount(t *testing.T) {}\n")
	writeCatalogFixture(t, root, "graphify-out/skipped_test.go", "func TestShouldNotCount(t *testing.T) {}\n")
	writeCatalogFixture(t, root, ".git/skipped_test.go", "func TestShouldNotCount(t *testing.T) {}\n")

	got, err := countTestFuncs(root)
	if err != nil {
		t.Fatalf("countTestFuncs(%q): %v", root, err)
	}
	if got != 2 {
		t.Errorf("countTestFuncs(%q) = %d, want 2", root, got)
	}
}

// writeCatalogFixture creates rel under root, making its parent directories
// as needed, so each fixture tree in this file can be written in one line.
func writeCatalogFixture(t *testing.T, root, rel, body string) {
	t.Helper()
	full := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(full, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}
