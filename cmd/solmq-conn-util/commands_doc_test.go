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

	leaves := 0
	for _, v := range cliVerbs {
		if !v.InUsage {
			continue
		}
		if len(v.Targets) == 0 {
			leaves++
			if !strings.Contains(collapsed, "solmq-conn-util "+v.Name) {
				t.Errorf("usage() is missing command %q", v.Name)
			}
			continue
		}
		for _, tg := range v.Targets {
			leaves++
			inv := "solmq-conn-util " + v.Name + " " + tg.Name
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
}
