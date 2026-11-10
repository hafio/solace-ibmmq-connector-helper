package main

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/hafio/go-deploy-utility/notice"
	"github.com/hafio/go-deploy-utility/notice/noticetest"
	"github.com/hafio/go-deploy-utility/paritydoc"
)

// supportNoticeDocs are the hand-written docs that must carry supportNotice
// verbatim, as paths relative to the repository root. They have no renderer to
// take it from, so this list is what keeps them from drifting away from
// support.go.
var supportNoticeDocs = []string{
	"README.md",
	"docs/getting-started.md",
	"docs/configuration.md",
	"docs/operations.md",
	"docs/troubleshooting.md",
	"docs/DESIGN.md",
	"docs/DEVELOPMENT.md",
	"docs/parity.md",
}

const supportNoticePagePath = "../../solmq-conn-util-generator.html"

// supportNoticeElemRE finds the generator page's notice <div> by its id, so the
// class, role and attribute order are free to change. It ends at the first
// </div>, not the first closing tag: the notice nests a <strong>.
var supportNoticeElemRE = regexp.MustCompile(`(?s)<div[^>]*\bid="support-notice"[^>]*>(.*?)</div>`)

var htmlTagRE = regexp.MustCompile(`<[^>]+>`)

// TestSupportNoticeInHandWrittenDocs pins the support statement at the top of
// the README and every hand-written doc under docs/ -- each one a page a
// reader can land on directly without passing through any other -- and holds
// both forms of the statement to plain ASCII, since they reach terminals, logs
// and the in-binary output.
func TestSupportNoticeInHandWrittenDocs(t *testing.T) {
	noticetest.AssertASCII(t, "the support notice", supportNotice.Lines())
	noticetest.AssertASCII(t, "the short support notice", notice.Short())
	noticetest.AssertDocs(t, supportNotice, "../..", noticetest.Options{Files: supportNoticeDocs})
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
		noticetest.Assert(t, name, []byte(render()), supportNotice, noticetest.Options{})
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
	if want := supportNotice.Plain(); got != want {
		t.Errorf("the generator page's support notice has drifted from supportNotice (cmd/solmq-conn-util/support.go)\n got: %s\nwant: %s", got, want)
	}
	if body := strings.Index(page, `<div class="page">`); body < 0 || body < m[0] {
		t.Errorf("the support notice must sit above the page body, directly under the header")
	}
	if strings.Contains(page, "100vh - 60px") {
		t.Errorf("the page is still sized to a guessed header height (calc(100vh - 60px)); the notice strip needs body to lay out as a flex column instead")
	}
}

// TestParityDocMatchesTheSharedBody holds docs/parity.md to the body the shared
// module owns (paritydoc/body.md): both tools carry it unchanged below their own
// support notice, so the page cannot drift from the other tool's copy.
func TestParityDocMatchesTheSharedBody(t *testing.T) {
	path := filepath.FromSlash("../../docs/parity.md")
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	if err := paritydoc.Check(b); err != nil {
		t.Fatal(err)
	}
}
