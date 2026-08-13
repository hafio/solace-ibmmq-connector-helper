package gen_test

import (
	"flag"
	"os"
	"regexp"
	"strings"
	"testing"
)

// The spec generator page embeds a copy of the golden application.yml so its
// Self-test button can diff the JavaScript port of consolidate+render against
// what the Go code produces. docs/DEVELOPMENT.md documents keeping the two in
// sync as a manual rule; this test makes it a failing build instead, because a
// stale copy makes Self-test report a pass that means nothing.
const generatorPagePath = "../../solmq-conn-generator.html"

var updateHTMLGolden = flag.Bool("update-html-golden", false,
	"rewrite the embedded golden in solmq-conn-generator.html from testdata/golden/application.yml")

// goldenBlockRE captures the embedded copy. The block is `text/plain` so the
// browser never executes it and the YAML survives verbatim.
var goldenBlockRE = regexp.MustCompile(`(?s)(<script type="text/plain" id="golden">)(.*?)(</script>)`)

// normGolden mirrors the page's selfTest(): CRLF folded, then exactly one
// trailing newline, so an embedded block stored without its final newline still
// compares equal.
func normGolden(s string) string {
	return strings.TrimRight(norm(s), "\n") + "\n"
}

func TestGeneratorPageGoldenInSync(t *testing.T) {
	page, err := os.ReadFile(generatorPagePath)
	if err != nil {
		t.Fatalf("reading %s: %v", generatorPagePath, err)
	}
	m := goldenBlockRE.FindSubmatchIndex(page)
	if m == nil {
		t.Fatalf(`%s has no <script type="text/plain" id="golden"> block`, generatorPagePath)
	}

	// The page's own selfTest() normalizes CRLF and trailing newlines before
	// comparing, so match it exactly here -- otherwise this gate and Self-test
	// could disagree about the same two files.
	want := normGolden(string(mustRead(t, "../../testdata/golden/application.yml")))
	got := normGolden(string(page[m[4]:m[5]]))

	if got == want {
		return
	}
	if *updateHTMLGolden {
		updated := append([]byte(nil), page[:m[4]]...)
		updated = append(updated, strings.TrimSuffix(want, "\n")...)
		updated = append(updated, page[m[5]:]...)
		if err := os.WriteFile(generatorPagePath, updated, 0o644); err != nil {
			t.Fatalf("writing %s: %v", generatorPagePath, err)
		}
		t.Log("embedded golden updated; re-run without -update-html-golden")
		return
	}
	t.Errorf("the golden embedded in %s is out of sync with testdata/golden/application.yml.\n"+
		"Regenerate: go test ./internal/gen -run TestGeneratorPageGoldenInSync -update-html-golden\n%s",
		generatorPagePath, lineDiff(want, got))
}
