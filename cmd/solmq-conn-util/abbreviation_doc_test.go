package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The -update flag and normLF live in commands_doc_test.go: one flag
// registration per package (a second flag.Bool("update", ...) panics with "flag
// redefined" before any test runs), and one line-ending helper for both gates.

const abbreviationDocPath = "../../docs/abbreviation.md"

const abbreviationRegenerate = "regenerate: go test ./cmd/solmq-conn-util -run TestAbbreviationDocInSync -update"

// TestAbbreviationDocInSync is the drift gate: docs/abbreviation.md must equal
// what the command model renders. It fails the build (and CI) whenever the two
// diverge -- the same contract TestCommandsDocInSync holds for commands.md.
func TestAbbreviationDocInSync(t *testing.T) {
	got := renderAbbreviationDoc()
	path := filepath.FromSlash(abbreviationDocPath)
	if *updateDoc {
		if err := os.WriteFile(path, []byte(got), 0o644); err != nil {
			t.Fatalf("update %s: %v", path, err)
		}
		t.Logf("regenerated %s", path)
		return
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v\n%s", path, err, abbreviationRegenerate)
	}
	if normLF(string(want)) != normLF(got) {
		t.Errorf("docs/abbreviation.md is out of sync with the command model (cmd/solmq-conn-util/commands.go).\n" +
			abbreviationRegenerate)
	}
}

// TestAbbreviationDocCoversModel gates the renderer against the model, which the
// byte comparison above cannot: a regenerated file agrees with a renderer that
// forgot a whole class of abbreviation. Every alias the binary accepts must
// appear on the page, and the page must carry no row the model does not back.
func TestAbbreviationDocCoversModel(t *testing.T) {
	doc := renderAbbreviationDoc()
	want := modeledAbbreviations()

	for short, what := range want {
		if !strings.Contains(doc, "| "+bt+short+bt+" |") {
			t.Errorf("docs/abbreviation.md has no row for %q (%s)\n%s", short, what, abbreviationRegenerate)
		}
	}

	// The other direction: a row for something the model does not declare. Set
	// comparison rather than a row count, so two kinds of abbreviation that ever
	// share a spelling (a verb alias and a target alias under another verb) do
	// not read as a spurious extra row.
	for _, short := range renderedAbbreviations(doc) {
		if _, ok := want[short]; !ok {
			t.Errorf("docs/abbreviation.md has a row for %q, which the model does not declare", short)
		}
	}
}

// renderedAbbreviations reads back the abbreviation each table row is keyed by:
// the first code span of a row, which is where every table puts it.
func renderedAbbreviations(doc string) []string {
	var out []string
	for _, ln := range strings.Split(doc, "\n") {
		rest, ok := strings.CutPrefix(ln, "| "+bt)
		if !ok {
			continue
		}
		if short, _, ok := strings.Cut(rest, bt); ok {
			out = append(out, short)
		}
	}
	return out
}

// modeledAbbreviations collects every short spelling the binary accepts, mapped
// to a description of where it comes from so a missing row names itself. Built
// by walking the same model the renderer walks, but independently of it: that is
// what makes it a check rather than a restatement.
func modeledAbbreviations() map[string]string {
	want := map[string]string{}
	for _, v := range cliVerbs {
		for _, a := range v.Aliases {
			want[a] = "alias of verb " + v.Name
		}
		addTargetAbbreviations(want, v.Name, v.Targets)
	}
	for _, e := range platformAliasList {
		want[e.Alias] = "short spelling of platform " + e.Canonical
	}
	for _, f := range cliFlags {
		if f.Short != f.Long {
			want[f.Short] = "short form of flag " + f.Long
		}
	}
	return want
}

// addTargetAbbreviations records the aliases of one level of target words and
// then of the sets beneath them, under the command path they hang off. Recursive
// so a command level deeper than download/jar/mq needs no change here.
func addTargetAbbreviations(want map[string]string, under string, targets []cliTarget) {
	for _, tg := range targets {
		for _, a := range tg.Aliases {
			want[a] = "alias of target " + under + " " + tg.Name
		}
		addTargetAbbreviations(want, under+" "+tg.Name, tg.Sets)
	}
}

// TestAbbreviationDocTableShape fails when a row does not have its table's
// column count: a literal "|" in a flag Meaning that reached the page unescaped
// splits the row into extra columns, which renders as a broken table rather than
// as an error.
func TestAbbreviationDocTableShape(t *testing.T) {
	doc := renderAbbreviationDoc()
	if !strings.Contains(doc, "| Short |") {
		t.Fatal("the rendered page has no abbreviation table at all")
	}
	cols := 0
	header := ""
	for i, ln := range strings.Split(doc, "\n") {
		switch {
		case !strings.HasPrefix(ln, "|"):
			cols, header = 0, ""
		case cols == 0:
			cols, header = countCells(ln), ln
		default:
			if got := countCells(ln); got != cols {
				t.Errorf("line %d has %d cells, want %d under header %q: %q", i+1, got, cols, header, ln)
			}
		}
	}
}

// countCells counts a markdown row's cells, honouring the backslash escape
// tableCell writes, so an escaped delimiter inside a cell does not read as a
// column boundary.
func countCells(row string) int {
	n := 0
	for i := 0; i < len(row); i++ {
		if row[i] == '\\' {
			i++
			continue
		}
		if row[i] == '|' {
			n++
		}
	}
	return n - 1 // the leading and trailing pipes bound n-1 cells
}
