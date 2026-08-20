package main

import (
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// -update regenerates the committed docs from the model instead of asserting.
// Run: go test ./cmd/solmq-conn-util -run TestCommandsDocInSync -update
var updateDoc = flag.Bool("update", false, "regenerate generated docs (docs/commands.md)")

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

// TestCommandsModelMatchesUsage anchors the model to the authoritative in-binary
// help, in both directions: every documented command/flag appears in usage(),
// and usage() lists no command the model omits.
func TestCommandsModelMatchesUsage(t *testing.T) {
	u := usageText()
	collapsed := strings.Join(strings.Fields(u), " ") // fold the help's alignment padding

	// usageName is how a verb heads its usage line: bare when it has no alias,
	// and <name|alias> when it does, so -h shows both spellings. Building the
	// expectation from the model is also what makes the alias genuinely
	// asserted here -- a bare strings.Contains(u, "gen") would pass on the
	// "generate" already in the text without the alias appearing at all.
	usageName := func(v cliVerb) string {
		if len(v.Aliases) == 0 {
			return v.Name
		}
		return "<" + strings.Join(append([]string{v.Name}, v.Aliases...), "|") + ">"
	}

	leaves := 0
	for _, v := range cliVerbs {
		if !v.InUsage {
			continue
		}
		if len(v.Targets) == 0 {
			leaves++
			if want := "solmq-conn-util " + usageName(v); !strings.Contains(collapsed, want) {
				t.Errorf("usage() is missing command %q", want)
			}
			continue
		}
		for _, tg := range v.Targets {
			leaves++
			inv := "solmq-conn-util " + usageName(v) + " " + tg.Name
			if !strings.Contains(collapsed, inv) {
				t.Errorf("usage() is missing command %q", inv)
			}
		}
	}

	got := 0
	for _, ln := range strings.Split(u, "\n") {
		if strings.HasPrefix(ln, "  solmq-conn-util ") {
			got++
		}
	}
	if got != leaves {
		t.Errorf("usage() lists %d command lines but the model has %d InUsage commands", got, leaves)
	}

	for _, f := range cliFlags {
		if !strings.Contains(u, f.Short) || !strings.Contains(u, f.Long) {
			t.Errorf("usage() is missing flag %s/%s", f.Short, f.Long)
		}
	}

	// --platform accepts short spellings too, and -h is where someone looks
	// first: the docs carried them while usage() did not, so this keeps the two
	// from diverging again as the table changes. Asserted as the pipe-joined
	// list rather than one alias at a time, because "kube" on its own is a
	// substring of "kubernetes" and would pass against usage text that never
	// mentions a short spelling at all.
	shorts := make([]string, 0, len(platformAliasList))
	for _, e := range platformAliasList {
		shorts = append(shorts, e.Alias)
	}
	if want := strings.Join(shorts, "|"); !strings.Contains(u, want) {
		t.Errorf("usage() should list the platform short spellings as %q", want)
	}
}
