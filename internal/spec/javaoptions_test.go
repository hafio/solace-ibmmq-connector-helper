package spec

import (
	"strings"
	"testing"
)

// TestParseEnvJavaOptionsSpellings covers the two spellings the key is meant to
// take -- a plain string and a >- folded block -- plus the literal block an
// author might reach for instead, and pins that all three reach the renderers
// as the same single line once normalized.
func TestParseEnvJavaOptionsSpellings(t *testing.T) {
	cases := []struct {
		name, doc, wantTool, wantJDK string
	}{
		{
			name:     "plain strings",
			doc:      "java-options:\n  tool: -Xms512m -Xmx1g\n  jdk: --add-opens=java.base/java.lang=ALL-UNNAMED\n",
			wantTool: "-Xms512m -Xmx1g",
			wantJDK:  "--add-opens=java.base/java.lang=ALL-UNNAMED",
		},
		{
			name:     "folded block",
			doc:      "java-options:\n  tool: >-\n    -Xms512m\n    -Xmx1g\n    -XX:+HeapDumpOnOutOfMemoryError\n",
			wantTool: "-Xms512m -Xmx1g -XX:+HeapDumpOnOutOfMemoryError",
		},
		{
			// A literal block keeps its newlines, and a more-indented folded line
			// keeps its own; normalizing is what makes both safe to render.
			name:    "literal block",
			doc:     "java-options:\n  jdk: |\n    -Duser.language=en\n    -Duser.country=SG\n",
			wantJDK: "-Duser.language=en -Duser.country=SG",
		},
		{
			name:     "only one sub-key",
			doc:      "java-options:\n  tool: -Xmx1g\n",
			wantTool: "-Xmx1g",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			e, err := ParseEnv([]byte(c.doc))
			if err != nil {
				t.Fatalf("ParseEnv: %v", err)
			}
			if e.JavaOptions == nil {
				t.Fatal("java-options did not parse")
			}
			if got := NormalizeJavaOptions(e.JavaOptions.Tool); got != c.wantTool {
				t.Errorf("tool = %q, want %q", got, c.wantTool)
			}
			if got := NormalizeJavaOptions(e.JavaOptions.JDK); got != c.wantJDK {
				t.Errorf("jdk = %q, want %q", got, c.wantJDK)
			}
		})
	}
}

// TestParseEnvJavaOptionsAbsent pins that an env.yaml without the block, or
// with it left empty, sets no JVM options at all.
func TestParseEnvJavaOptionsAbsent(t *testing.T) {
	for _, doc := range []string{"timezone: UTC\n", "java-options:\n"} {
		e, err := ParseEnv([]byte(doc))
		if err != nil {
			t.Fatalf("ParseEnv(%q): %v", doc, err)
		}
		if tool, jdk := JavaEnv(false, e.JavaOptions); tool != "" || jdk != "" {
			t.Errorf("ParseEnv(%q) sets JVM options: tool %q, jdk %q", doc, tool, jdk)
		}
	}
}

// TestParseEnvJavaOptionsShapeErrors covers the mistakes ParseEnv would
// otherwise swallow: it decodes without KnownFields, so a misspelled sub-key or
// a list of options would start the connector without them and say nothing.
// Each has to fail at parse, naming the key and what it should be.
func TestParseEnvJavaOptionsShapeErrors(t *testing.T) {
	cases := []struct {
		name, doc string
		want      []string
	}{
		{"a bare string instead of the block", "java-options: -Xmx1g\n", []string{"java-options must be a mapping", "tool:", "jdk:"}},
		{"a misspelled sub-key", "java-options:\n  tools: -Xmx1g\n", []string{"java-options.tools is not a key", "JAVA_TOOL_OPTIONS", "JDK_JAVA_OPTIONS"}},
		{"a list of options", "java-options:\n  tool:\n    - -Xms512m\n    - -Xmx1g\n", []string{"java-options.tool must be one string", ">-", "got a list"}},
		{"a mapping as a value", "java-options:\n  jdk:\n    xmx: 1g\n", []string{"java-options.jdk must be one string", "got a mapping"}},
		{"a sub-key given twice", "java-options:\n  tool: -Xms512m\n  tool: -Xmx1g\n", []string{"java-options.tool is set twice"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := ParseEnv([]byte(c.doc))
			if err == nil {
				t.Fatal("expected a parse error")
			}
			for _, w := range c.want {
				if !strings.Contains(err.Error(), w) {
					t.Errorf("error %q should contain %q", err, w)
				}
			}
		})
	}
}

// TestParseEnvJavaOptionsAlias pins that a YAML alias to a scalar is followed,
// so an anchor shared with another value is not reported as the wrong shape.
func TestParseEnvJavaOptionsAlias(t *testing.T) {
	e, err := ParseEnv([]byte("x-heap: &heap -Xmx1g\njava-options:\n  tool: *heap\n"))
	if err != nil {
		t.Fatalf("ParseEnv: %v", err)
	}
	if e.JavaOptions == nil || e.JavaOptions.Tool != "-Xmx1g" {
		t.Errorf("tool = %+v, want -Xmx1g", e.JavaOptions)
	}
}

// TestJavaEnvMergesTheTLSFlag pins the one rule every renderer relies on: the
// tool's MQ TLS flag comes first in JAVA_TOOL_OPTIONS and the operator's
// options after it, so an operator who sets the same property wins, and
// JDK_JAVA_OPTIONS never carries the TLS flag.
func TestJavaEnvMergesTheTLSFlag(t *testing.T) {
	cases := []struct {
		name              string
		mqTLS             bool
		j                 *JavaOptions
		wantTool, wantJDK string
	}{
		{"nothing set", false, nil, "", ""},
		{"TLS only", true, nil, MQTLSJavaToolOption, ""},
		{"TLS with an empty block", true, &JavaOptions{}, MQTLSJavaToolOption, ""},
		{"operator only", false, &JavaOptions{Tool: "-Xmx1g", JDK: "-Duser.language=en"}, "-Xmx1g", "-Duser.language=en"},
		{"TLS then operator", true, &JavaOptions{Tool: "-Xmx1g"}, MQTLSJavaToolOption + " -Xmx1g", ""},
		{"jdk never gets the TLS flag", true, &JavaOptions{JDK: "-Xmx1g"}, MQTLSJavaToolOption, "-Xmx1g"},
		{"whitespace-only values set nothing", false, &JavaOptions{Tool: " \n ", JDK: "\t"}, "", ""},
		{"normalized before merging", true, &JavaOptions{Tool: "  -Xms512m\n\n  -Xmx1g  "}, MQTLSJavaToolOption + " -Xms512m -Xmx1g", ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			tool, jdk := JavaEnv(c.mqTLS, c.j)
			if tool != c.wantTool {
				t.Errorf("JAVA_TOOL_OPTIONS = %q, want %q", tool, c.wantTool)
			}
			if jdk != c.wantJDK {
				t.Errorf("JDK_JAVA_OPTIONS = %q, want %q", jdk, c.wantJDK)
			}
		})
	}
}

// TestExpandJavaOptions pins that both sub-keys take ${VAR} and ${VAR:default}
// like every other non-credential value, so a heap size can come from the
// deploying shell.
func TestExpandJavaOptions(t *testing.T) {
	e := &Env{JavaOptions: &JavaOptions{Tool: "-Xmx${HEAP:512m}", JDK: "-Dregion=${REGION}"}}
	Expand(Expander{Lookup: lookupOf(map[string]string{"REGION": "apac"})}, e, nil)
	if got, want := e.JavaOptions.Tool, "-Xmx512m"; got != want {
		t.Errorf("tool = %q, want %q", got, want)
	}
	if got, want := e.JavaOptions.JDK, "-Dregion=apac"; got != want {
		t.Errorf("jdk = %q, want %q", got, want)
	}
}
