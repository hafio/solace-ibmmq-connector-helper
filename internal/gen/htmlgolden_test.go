package gen_test

import (
	"flag"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/gen"
)

// The spec generator page embeds a copy of the golden application.yml, and a
// second block of the findings internal/validate reports for that same spec,
// so its Self-test button can diff the JavaScript port of consolidate+render
// AND validateModel against what the Go code produces. docs/DEVELOPMENT.md
// documents keeping the two in sync as a manual rule; these tests make it a
// failing build instead, because a stale copy makes Self-test report a pass
// that means nothing.
const generatorPagePath = "../../solmq-conn-util-generator.html"

var updateHTMLGolden = flag.Bool("update-html-golden", false,
	"rewrite the embedded goldens in solmq-conn-util-generator.html from testdata/golden/application.yml "+
		"and gen.Validate's findings for testdata/golden/specs")

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

// goldenFindingsBlockRE captures the embedded expected-findings copy, the
// sibling of goldenBlockRE sitting right below the application.yml block.
var goldenFindingsBlockRE = regexp.MustCompile(`(?s)(<script type="text/plain" id="golden-findings">)(.*?)(</script>)`)

// formatFindings renders errs then warns as the golden-findings block's line
// format: "ERROR <issue>" / "WARN <issue>", <issue> being Issue.String()
// (file: msg, or a bare msg when File is ""). The page's Self-test button
// builds this same format from its own validateModel() port before diffing
// it against the embedded block, so the two must stay byte-identical.
func formatFindings(errs, warns []gen.Issue) string {
	var b strings.Builder
	for _, e := range errs {
		b.WriteString("ERROR " + e.String() + "\n")
	}
	for _, w := range warns {
		b.WriteString("WARN " + w.String() + "\n")
	}
	return b.String()
}

// sampleFindings runs the same full lint the CLI's `validate` verb does
// (gen.Validate: internal/validate.Run, plus the parse-time ${VAR} expansion
// warnings and the post-consolidate mount-name conflict check it layers on
// top) against the golden spec folder that backs the id="golden"
// application.yml block above, so both embedded blocks describe one fixture.
// sampleModel() in solmq-conn-util-generator.html is a hand-written mirror of
// that same folder (see TestGeneratorPageGoldenInSync, which already assumes
// this for the application.yml output).
//
// The golden spec validates clean today (0 errors, 0 warnings), so this gate
// only guards against a regression that would make either side start
// reporting a finding for it -- it does not exercise the ERROR/WARN line
// format itself, since there is nothing to format. Closing that gap needs a
// second, deliberately-warned fixture that solmq-conn-util-generator.html's
// sample can also mirror; none exists yet.
func sampleFindings(t *testing.T) string {
	t.Helper()
	setGoldenCredEnv(t)
	req := loadSpecs(t)
	errs, warns := gen.Validate(req, gen.Resolver{Env: os.LookupEnv, ReadFile: dirReader()})
	return formatFindings(errs, warns)
}

func TestGeneratorPageFindingsGoldenInSync(t *testing.T) {
	page, err := os.ReadFile(generatorPagePath)
	if err != nil {
		t.Fatalf("reading %s: %v", generatorPagePath, err)
	}
	m := goldenFindingsBlockRE.FindSubmatchIndex(page)
	if m == nil {
		t.Fatalf(`%s has no <script type="text/plain" id="golden-findings"> block`, generatorPagePath)
	}

	want := normGolden(sampleFindings(t))
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
		t.Log("embedded findings golden updated; re-run without -update-html-golden")
		return
	}
	t.Errorf("the findings embedded in %s are out of sync with what gen.Validate reports for testdata/golden/specs.\n"+
		"Regenerate: go test ./internal/gen -run TestGeneratorPageFindingsGoldenInSync -update-html-golden\n%s",
		generatorPagePath, lineDiff(want, got))
}

// selfTestGotFRE captures selfTest()'s gotF assignment, the JS mirror of
// formatFindings above that Self-test diffs against wantF (the embedded
// golden-findings block, CRLF-folded then trimmed-and-padded to exactly one
// trailing newline).
var selfTestGotFRE = regexp.MustCompile(`const gotF = (.*);\n`)

// TestGeneratorPageSelfTestNormalizesEmptyFindings guards against gotF
// skipping the same trim-then-pad-one-newline normalization wantF applies.
// Without it, the golden spec's clean case (0 errors, 0 warnings,
// formatFindings() === "") compares "" against wantF's "\n" and Self-test
// reports FAIL on the page's own fixture every time -- a regression this test
// would have caught without needing a JS engine, since it checks the actual
// source text shipped in the page rather than a Go-side mirror of it.
func TestGeneratorPageSelfTestNormalizesEmptyFindings(t *testing.T) {
	page, err := os.ReadFile(generatorPagePath)
	if err != nil {
		t.Fatalf("reading %s: %v", generatorPagePath, err)
	}
	m := selfTestGotFRE.FindSubmatch(page)
	if m == nil {
		t.Fatalf(`%s has no "const gotF = ...;" assignment in selfTest()`, generatorPagePath)
	}
	got := string(m[1])
	if !strings.Contains(got, `.replace(/\n*$/, '')`) || !strings.HasSuffix(got, `+ '\n'`) {
		t.Errorf("%s selfTest()'s gotF = %s must trim trailing newlines then add exactly one, "+
			"mirroring wantF's normalization -- otherwise Self-test fails on the empty-findings case",
			generatorPagePath, got)
	}
}
