package main

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// supportNoticeDocs are the hand-written docs that must carry
// supportNoticeMarkdown verbatim. They have no renderer to take it from, so
// this list is what keeps them from drifting away from support.go.
var supportNoticeDocs = []string{
	"../../README.md",
	"../../docs/userguide.md",
	"../../docs/DEVELOPMENT.md",
}

const supportNoticePagePath = "../../solmq-conn-util-generator.html"

// supportNoticeElemRE finds the generator page's notice <div> by its id, so the
// class, role and attribute order are free to change. It ends at the first
// </div>, not the first closing tag: the notice nests a <strong>.
var supportNoticeElemRE = regexp.MustCompile(`(?s)<div[^>]*\bid="support-notice"[^>]*>(.*?)</div>`)

var htmlTagRE = regexp.MustCompile(`<[^>]+>`)

// assertNoticeAtTop requires the notice block verbatim, as its own paragraph,
// ahead of the doc's first section heading. The blank lines either side are
// part of the contract: a Markdown line that directly follows a blockquote is
// folded into it, so without them the doc's intro paragraph would render
// inside the warning.
func assertNoticeAtTop(t *testing.T, name, doc string) {
	t.Helper()
	block := strings.Join(supportNoticeMarkdown, "\n")
	i := strings.Index(doc, "\n\n"+block+"\n\n")
	if i < 0 {
		t.Errorf("%s does not carry the support notice verbatim, with a blank line either side; "+
			"paste supportNoticeMarkdown (cmd/solmq-conn-util/support.go) directly under the H1:\n\n%s", name, block)
		return
	}
	if h := strings.Index(doc, "\n## "); h >= 0 && h < i {
		t.Errorf("%s carries the support notice below its first section heading; move it directly under the H1", name)
	}
}

// supportNoticePlain reduces the Markdown block to its words: no quote
// markers, no alert tag, no bold or code-span markup.
func supportNoticePlain() string {
	var words []string
	for _, ln := range supportNoticeMarkdown {
		ln = strings.TrimSpace(strings.TrimPrefix(ln, ">"))
		if ln == "[!WARNING]" {
			continue
		}
		ln = strings.NewReplacer("**", "", bt, "").Replace(ln)
		words = append(words, strings.Fields(ln)...)
	}
	return strings.Join(words, " ")
}

// TestSupportNoticeInHandWrittenDocs pins the support statement at the top of
// the README, the user guide and the development guide -- the three docs a
// reader can land on directly without passing through any other -- and holds
// both forms of the statement to plain ASCII, since they reach terminals, logs
// and the in-binary output.
func TestSupportNoticeInHandWrittenDocs(t *testing.T) {
	for _, form := range [][]string{supportNoticeMarkdown, supportNoticeVersion} {
		for _, ln := range form {
			for i := 0; i < len(ln); i++ {
				if ln[i] > 0x7e || (ln[i] < 0x20 && ln[i] != '\t') {
					t.Errorf("support notice line has a non-ASCII or control byte 0x%02x at %d: %q", ln[i], i, ln)
				}
			}
		}
	}
	for _, rel := range supportNoticeDocs {
		path := filepath.FromSlash(rel)
		t.Run(filepath.Base(path), func(t *testing.T) {
			b, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("read %s: %v", path, err)
			}
			assertNoticeAtTop(t, rel, normLF(string(b)))
		})
	}
}

// TestSupportNoticeInGeneratedDocs pins the notice at the top of both
// generated reference docs. The committed files are then held to these
// renderers byte-for-byte by TestCommandsDocInSync and
// TestAbbreviationDocInSync.
func TestSupportNoticeInGeneratedDocs(t *testing.T) {
	for name, render := range map[string]func() string{
		"docs/commands.md":     renderCommandsDoc,
		"docs/abbreviation.md": renderAbbreviationDoc,
	} {
		assertNoticeAtTop(t, name, render())
	}
}

// TestSupportNoticeInGeneratorPage covers the one copy of the statement that
// is not Markdown. The page is the no-install entry point the README offers,
// so its users may never open a doc at all. Its text must read exactly as the
// Markdown block does, it must come before the page body, and the page must
// no longer size its body to a guessed header height -- that calc left no
// room for the strip and pushed the panes past the bottom of the viewport.
func TestSupportNoticeInGeneratorPage(t *testing.T) {
	path := filepath.FromSlash(supportNoticePagePath)
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	page := normLF(string(b))

	m := supportNoticeElemRE.FindStringSubmatchIndex(page)
	if m == nil {
		t.Fatalf("%s has no element with id=\"support-notice\"", supportNoticePagePath)
	}
	got := strings.Join(strings.Fields(htmlTagRE.ReplaceAllString(page[m[2]:m[3]], "")), " ")
	if want := supportNoticePlain(); got != want {
		t.Errorf("the generator page's support notice has drifted from supportNoticeMarkdown (cmd/solmq-conn-util/support.go)\n got: %s\nwant: %s", got, want)
	}
	if body := strings.Index(page, `<div class="page">`); body < 0 || body < m[0] {
		t.Errorf("the support notice must sit above the page body, directly under the header")
	}
	if strings.Contains(page, "100vh - 60px") {
		t.Errorf("the page is still sized to a guessed header height (calc(100vh - 60px)); the notice strip needs body to lay out as a flex column instead")
	}
}
