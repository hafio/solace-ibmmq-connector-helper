package yamlwriter

import "testing"

func TestWriterLineIndent(t *testing.T) {
	w := &Writer{}
	w.Line(0, "spring:")
	w.Line(2, "config:")
	w.Line(0, "top:")
	want := "spring:\n  config:\ntop:\n"
	if got := w.String(); got != want {
		t.Errorf("String() = %q, want %q", got, want)
	}
}

func TestWriterRawPassthrough(t *testing.T) {
	w := &Writer{}
	w.Line(2, "content: |")
	w.Raw("\n")
	w.Line(4, "line2")
	want := "  content: |\n\n    line2\n"
	if got := w.String(); got != want {
		t.Errorf("String() = %q, want %q", got, want)
	}
}

func TestSplitLines(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want []string
	}{
		{"trailing newline dropped", "a\nb\n", []string{"a", "b"}},
		{"no trailing newline kept", "a\nb", []string{"a", "b"}},
		{"empty string", "", []string{}},
		{"blank line preserved", "a\n\nb\n", []string{"a", "", "b"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := SplitLines(c.in)
			if len(got) != len(c.want) {
				t.Fatalf("SplitLines(%q) = %v, want %v", c.in, got, c.want)
			}
			for i := range got {
				if got[i] != c.want[i] {
					t.Errorf("SplitLines(%q)[%d] = %q, want %q", c.in, i, got[i], c.want[i])
				}
			}
		})
	}
}
