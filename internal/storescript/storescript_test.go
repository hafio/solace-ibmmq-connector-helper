package storescript

import (
	"slices"
	"strings"
	"testing"
)

// The script decodes every encoded store into the in-memory store directory
// under its own name, through a temporary name, with the decoded files closed
// to everyone but the connector's identity, stops on any failure, and then
// becomes the connector by exec'ing the command it is handed.
func TestRenderDecodesEveryStoreThenExecs(t *testing.T) {
	got := Render()
	for _, want := range []string{
		"#!/bin/sh\n",
		"\nset -eu\n",
		"\numask 077\n",
		"\nfor f in /app/external/var/stores/*; do\n",
		"\n  [ -f \"$f\" ] || continue\n",
		"\n  name=${f##*/}\n",
		"\n  base64 -d \"$f\" > \"/app/external/classpath/truststores/$name.tmp\"\n",
		"\n  mv \"/app/external/classpath/truststores/$name.tmp\" \"/app/external/classpath/truststores/$name\"\n",
		"\ndone\nexec \"$@\"\n",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("script lacks %q:\n%s", want, got)
		}
	}
}

// The script lives beside the status script, outside every directory a volume
// mounts over, and the command it execs is the image's own: its entrypoint,
// then its command, as compose's command: must hand them over once
// entrypoint: has dropped the image's CMD.
func TestCommandIsTheImagesOwn(t *testing.T) {
	if ContainerPath != "/app/external/.stores-init" {
		t.Errorf("ContainerPath = %q", ContainerPath)
	}
	want := append(append([]string(nil), ImageEntrypoint...), ImageCmd...)
	got := Command()
	if !slices.Equal(got, want) || got[0] != "java" || got[len(got)-1] != ImageCmd[0] {
		t.Errorf("Command() = %q, want %q", got, want)
	}
	got[0] = "changed"
	if ImageEntrypoint[0] != "java" {
		t.Error("Command must return a copy, never the recorded entrypoint itself")
	}
}
