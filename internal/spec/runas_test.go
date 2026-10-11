package spec

import (
	"strings"
	"testing"
)

// TestParseRunAs pins the run-as grammar. The ids are written into compose
// files, quadlet units and manifests as-is, so anything that is not a plain
// decimal id -- a sign, a leading zero, an exponent, a name, whitespace, a
// third field -- is refused here rather than reinterpreted by a sink.
func TestParseRunAs(t *testing.T) {
	for _, c := range []struct {
		in       string
		want     RunAsID
		wantText string
	}{
		{"1000:0", RunAsID{UID: 1000, GID: 0}, "1000:0"},
		{"1000", RunAsID{UID: 1000, GID: 0}, "1000:0"},
		{"0", RunAsID{}, "0:0"},
		{"2147483647:2147483647", RunAsID{UID: 2147483647, GID: 2147483647}, "2147483647:2147483647"},
		{"image", RunAsID{Image: true}, "image"},
	} {
		got, err := ParseRunAs(c.in)
		if err != nil || got != c.want || got.String() != c.wantText {
			t.Errorf("ParseRunAs(%q) = %+v (%q), %v; want %+v (%q)", c.in, got, got.String(), err, c.want, c.wantText)
		}
	}
	for _, c := range []struct{ in, want string }{
		{"", "the uid is empty"},
		{":0", "the uid is empty"},
		{"1000:", "the gid is empty"},
		{"01000", `the uid "01000" has a leading zero`},
		{"1000:00", `the gid "00" has a leading zero`},
		{"-1", `the uid "-1" is not a decimal number`},
		{"1e3", `the uid "1e3" is not a decimal number`},
		{" 1000", `the uid " 1000" is not a decimal number`},
		{"root", `the uid "root" is not a decimal number`},
		{"1000:root", `the gid "root" is not a decimal number`},
		{"1000:0:0", `the gid "0:0" is not a decimal number`},
		{"2147483648", `the uid "2147483648" is larger than 2147483647`},
		{"99999999999999999999", `the uid "99999999999999999999" is larger than 2147483647`},
		{"Image", `the uid "Image" is not a decimal number`},
	} {
		if got, err := ParseRunAs(c.in); err == nil || !strings.Contains(err.Error(), c.want) {
			t.Errorf("ParseRunAs(%q) = %+v, %v; want an error containing %q", c.in, got, err, c.want)
		}
	}
}

// TestDefaultRunAsIsNonRootAndParses keeps the default honest: it parses, it
// is not root, and it fits a stock rootless-podman subuid range (65536 ids),
// so the same default works for every platform and both podman modes.
func TestDefaultRunAsIsNonRootAndParses(t *testing.T) {
	id, err := ParseRunAs(DefaultRunAs)
	if err != nil {
		t.Fatalf("DefaultRunAs %q does not parse: %v", DefaultRunAs, err)
	}
	if id.Image || id.UID == 0 {
		t.Errorf("DefaultRunAs %q must name a non-root uid", DefaultRunAs)
	}
	if id.UID >= 65536 || id.GID >= 65536 {
		t.Errorf("DefaultRunAs %q must fit a stock subuid/subgid range of 65536 ids", DefaultRunAs)
	}
}
