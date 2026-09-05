package main

import (
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// -update regenerates the committed docs from the model instead of asserting.
// One registration for the whole package: TestAbbreviationDocInSync reads the
// same flag, and a second flag.Bool("update", ...) would panic before any test
// runs. Run: go generate ./cmd/solmq-conn-util
var updateDoc = flag.Bool("update", false, "regenerate generated docs (docs/commands.md, docs/abbreviation.md)")

// normLF collapses CRLF to LF so the byte comparison is independent of the
// checkout's line-ending setting (Windows-first repo; output contract is LF).
func normLF(s string) string { return strings.ReplaceAll(s, "\r\n", "\n") }

const commandsDocPath = "../../docs/commands.md"

// TestCommandsDocInSync is the drift gate: docs/commands.md must equal what the
// command model renders. It fails the build (and CI) whenever the two diverge.
func TestCommandsDocInSync(t *testing.T) {
	got := renderCommandsDoc()
	path := filepath.FromSlash(commandsDocPath)
	if *updateDoc {
		if err := os.WriteFile(path, []byte(got), 0o644); err != nil {
			t.Fatalf("update %s: %v", path, err)
		}
		t.Logf("regenerated %s", path)
		return
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v\nregenerate: go test ./cmd/solmq-conn-util -run TestCommandsDocInSync -update", path, err)
	}
	if normLF(string(want)) != normLF(got) {
		t.Errorf("docs/commands.md is out of sync with the command model (cmd/solmq-conn-util/commands.go).\n" +
			"regenerate: go test ./cmd/solmq-conn-util -run TestCommandsDocInSync -update")
	}
}

// TestInvocationTargetArgs is the regression gate for the resolved-target
// invocation: once status's target word is given, its own placeholder
// (statusTargetArgBracket) must not be re-appended from Args, while the
// verb-level invocation with no target still shows it -- that line is the one
// place a reader who has not yet picked a target needs to see the choices.
func TestInvocationTargetArgs(t *testing.T) {
	var status cliVerb
	for _, v := range cliVerbs {
		if v.Name == "status" {
			status = v
		}
	}
	if status.Name == "" {
		t.Fatal("status verb not found in cliVerbs")
	}
	if got := invocation(status, ""); !strings.Contains(got, statusTargetArgBracket) {
		t.Errorf("invocation(status, \"\") = %q, want it to contain %q", got, statusTargetArgBracket)
	}
	for _, tg := range status.Targets {
		got := invocation(status, tg.Name)
		if strings.Contains(got, statusTargetArgBracket) {
			t.Errorf("invocation(status, %q) = %q, must not repeat the resolved placeholder %q", tg.Name, got, statusTargetArgBracket)
		}
		if !strings.HasPrefix(got, "solmq-conn-util status "+tg.Name+" ") {
			t.Errorf("invocation(status, %q) = %q, want it to start with the resolved target", tg.Name, got)
		}
	}
}

// TestCommandsModelMatchesUsage anchors the summary page to the model and to
// its own promises: one line per command carrying the model's description,
// detail deferred to the per-command pages, no aliases (md-only, by decision),
// and no line past the 100-column budget the page is designed never to wrap
// in. The full flag/argument coverage that used to be asserted against this
// page moved with the content to TestVerbUsagePages.
func TestCommandsModelMatchesUsage(t *testing.T) {
	u := usageText()
	lines := strings.Split(strings.TrimRight(u, "\n"), "\n")

	var commands []string
	in := false
	for _, ln := range lines {
		switch {
		case ln == "Commands:":
			in = true
		case in && strings.HasPrefix(ln, "  "):
			commands = append(commands, ln)
		case in:
			in = false
		}
	}
	if len(commands) != len(cliVerbs) {
		t.Fatalf("usage() lists %d command lines but the model has %d verbs:\n%s", len(commands), len(cliVerbs), u)
	}
	for i, v := range cliVerbs {
		ln := commands[i]
		if !strings.HasPrefix(ln, "  "+v.Name+" ") {
			t.Errorf("command line %d = %q, want it to start with %q", i, ln, v.Name)
		}
		if desc := plainText(verbBlurb(v)); !strings.Contains(ln, desc) {
			t.Errorf("command %q line is missing its description %q", v.Name, desc)
		}
	}

	assertNoAliases(t, "usage()", u)
	assertHelpWidth(t, "usage()", u)

	if !strings.Contains(u, "help <verb>") {
		t.Errorf("usage() should point at the per-command pages, got:\n%s", u)
	}
}

// TestVerbUsagePages is the drift gate for the per-command pages: every verb
// renders one, it carries the synopsis, every target word and its summary,
// every modeled flag the verb takes with its terse Usage text, and an example
// -- with no alias anywhere and no line past the page's width budget. This is
// where the detail deferred from the summary lives, so this is where its
// coverage is asserted.
func TestVerbUsagePages(t *testing.T) {
	byShort := make(map[string]cliFlag, len(cliFlags))
	for _, f := range cliFlags {
		byShort[f.Short] = f
	}
	for _, v := range cliVerbs {
		t.Run(v.Name, func(t *testing.T) {
			page := verbUsage(v.Name)
			if page == "" {
				t.Fatalf("verbUsage(%q) rendered nothing", v.Name)
			}
			if v.Synopsis == "" {
				t.Fatalf("verb %q has no Synopsis for its help page", v.Name)
			}
			// Folded for the Contains checks: a wrapped Usage text spans lines.
			folded := strings.Join(strings.Fields(page), " ")

			if !strings.Contains(page, "Usage:\n  solmq-conn-util "+v.Synopsis+"\n") {
				t.Errorf("page is missing the synopsis %q:\n%s", v.Synopsis, page)
			}
			if desc := plainText(verbBlurb(v)); !strings.Contains(folded, desc) {
				t.Errorf("page is missing the description %q", desc)
			}
			for _, tg := range v.Targets {
				if !strings.Contains(folded, tg.Name+" "+plainText(tg.Summary)) {
					t.Errorf("page is missing target %q with its summary", tg.Name)
				}
				for _, st := range tg.Sets {
					if !strings.Contains(folded, st.Name+" "+plainText(st.Summary)) {
						t.Errorf("page is missing set %q with its summary", st.Name)
					}
				}
			}
			for _, sh := range v.Flags {
				f, ok := byShort[sh]
				if !ok {
					continue // TestCompletionModelMetadataComplete reports the broken key
				}
				if f.Usage == "" {
					t.Errorf("flag %s has no terse Usage text, so the %q page cannot describe it", f.Long, v.Name)
					continue
				}
				for _, spelling := range flagOffered(f) {
					if !strings.Contains(page, spelling) {
						t.Errorf("page is missing flag spelling %q", spelling)
					}
				}
				if !strings.Contains(folded, f.Usage) {
					t.Errorf("page is missing %s's usage text %q", f.Long, f.Usage)
				}
			}
			if ex := verbExample(v); ex != "" && !strings.Contains(page, "\nExample:\n  "+ex+"\n") {
				t.Errorf("page is missing the example %q", ex)
			}
			assertNoAliases(t, "verbUsage("+v.Name+")", page)
			assertHelpWidth(t, "verbUsage("+v.Name+")", page)
		})
	}

	// A flag no verb lists would appear on no page at all -- unreachable from
	// every help surface, which is how documentation quietly rots.
	used := map[string]bool{}
	for _, v := range cliVerbs {
		for _, sh := range v.Flags {
			used[sh] = true
		}
	}
	for _, f := range cliFlags {
		if !used[f.Short] {
			t.Errorf("flag %s is listed by no verb, so no help page shows it", f.Long)
		}
	}

	// The platform short spellings must survive somewhere in terminal help;
	// they used to be asserted against the summary's flag table, which is gone.
	platform := byShort[platformFlagName]
	for _, e := range platformAliasList {
		if !strings.Contains(platform.Usage, e.Alias) {
			t.Errorf("--platform's Usage text should name the short spelling %q, got %q", e.Alias, platform.Usage)
		}
	}
}

// assertNoAliases fails when any verb or target alias appears as a standalone
// word in text: the short spellings keep working, but by decision they are
// documented only in docs/commands.md and the user guide, never in terminal
// help.
func assertNoAliases(t *testing.T, where, text string) {
	t.Helper()
	words := map[string]bool{}
	for _, w := range strings.FieldsFunc(text, func(r rune) bool {
		return !(r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '-')
	}) {
		words[w] = true
	}
	for _, v := range cliVerbs {
		for _, a := range v.Aliases {
			if words[a] {
				t.Errorf("%s shows alias %q; aliases are documented only in the markdown docs", where, a)
			}
		}
		for _, tg := range v.Targets {
			for _, a := range tg.Aliases {
				if words[a] {
					t.Errorf("%s shows target alias %q; aliases are documented only in the markdown docs", where, a)
				}
			}
		}
	}
}

// assertHelpWidth fails on any line past the width budget: the pages are
// designed never to wrap in a 100-column terminal, which is the failure mode
// the old 200-column summary page had.
func assertHelpWidth(t *testing.T, where, text string) {
	t.Helper()
	for _, ln := range strings.Split(text, "\n") {
		if len(ln) > helpPageWidth {
			t.Errorf("%s has a %d-column line (budget %d): %q", where, len(ln), helpPageWidth, ln)
		}
	}
}
