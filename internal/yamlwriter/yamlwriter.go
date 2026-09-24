// Package yamlwriter provides the indentation-aware line writer shared by
// every generator that emits YAML or line-oriented deployment artifacts
// (application.yml, the Kubernetes manifests, the compose file, the podman
// run script and quadlet unit). It exists because four packages each carried
// a private copy of this writer and had begun to drift in small ways, while
// every one of those generated artifacts is checked against a byte-for-byte
// golden file that depends on identical indentation.
package yamlwriter

import "strings"

// Writer accumulates lines with caller-chosen indentation (2 spaces per level
// is the convention every current caller uses).
type Writer struct{ b strings.Builder }

// Line writes s indented by indent spaces, followed by a newline. Callers with
// no nesting (e.g. podman's flat argument/unit lines) pass indent 0.
func (w *Writer) Line(indent int, s string) {
	w.b.WriteString(strings.Repeat(" ", indent))
	w.b.WriteString(s)
	w.b.WriteByte('\n')
}

// Raw writes s verbatim -- no indentation, no added newline -- for
// pre-formatted text such as a "---\n" document separator or a bare blank
// line inside a block scalar.
func (w *Writer) Raw(s string) { w.b.WriteString(s) }

// String returns the accumulated text.
func (w *Writer) String() string { return w.b.String() }

// SplitLines splits s on '\n' and drops a single trailing empty element that a
// terminating newline produces, so a block scalar built from the result keeps
// exactly one final newline.
func SplitLines(s string) []string {
	lines := strings.Split(s, "\n")
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return lines
}
