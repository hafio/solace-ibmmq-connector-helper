package validate

import (
	"strings"
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/spec"
)

// transformNode decodes one transform-headers: block from YAML text.
func transformNode(t *testing.T, src string) *yaml.Node {
	t.Helper()
	var doc yaml.Node
	if err := yaml.Unmarshal([]byte(src), &doc); err != nil {
		t.Fatalf("bad fixture: %v", err)
	}
	return doc.Content[0]
}

// runTransform validates the otherwise clean wfOK workflow carrying block as
// its transform-headers.
func runTransform(t *testing.T, block string) (errs, warns []Issue) {
	t.Helper()
	wfs := wfOK()
	wfs[0].TransformHeaders = transformNode(t, block)
	return Run(Context{Workflows: wfs, Defaults: &spec.Defaults{}})
}

// issueWith reports whether any issue's message contains sub.
func issueWith(issues []Issue, sub string) bool {
	for _, i := range issues {
		if strings.Contains(i.Msg, sub) {
			return true
		}
	}
	return false
}

// TestTransformHeadersValidBlockPasses pins that a well-formed block -- the
// shape the connector reads -- raises nothing.
func TestTransformHeadersValidBlockPasses(t *testing.T) {
	errs, warns := runTransform(t, "expressions:\n  solace_scst_targetDestination: \"'orders/' + headers.region\"\n  JMS_IBM_Format: \"'MQSTR'\"\n")
	if issueWith(errs, "transform-headers") || issueWith(warns, "transform-headers") {
		t.Errorf("a valid block must raise nothing, got errors %v, warnings %v", errs, warns)
	}
}

// TestTransformHeadersShapeErrors covers each way the block can be shaped so
// that the connector would not apply it. Each is an error on every run: the
// connector would start cleanly and transform nothing.
func TestTransformHeadersShapeErrors(t *testing.T) {
	for _, c := range []struct{ name, block, want string }{
		{"not a mapping", "just-a-string", "transform-headers must be a mapping with expressions: <header>: <SpEL expression>, got a scalar"},
		{"no expressions", "headers:\n  a: b\n", "transform-headers has no expressions:"},
		{"expressions as a list", "expressions:\n  - a\n", "transform-headers.expressions must be a mapping of <header>: <SpEL expression>, got a list"},
		{"an expression that is not a string", "expressions:\n  h:\n    nested: x\n", "transform-headers.expressions.h must be one SpEL expression, got a mapping"},
		{"a header set twice", "expressions:\n  h: \"'a'\"\n  h: \"'b'\"\n", `transform-headers.expressions sets "h" twice`},
	} {
		t.Run(c.name, func(t *testing.T) {
			errs, _ := runTransform(t, c.block)
			if !issueWith(errs, c.want) {
				t.Errorf("want an error containing %q, got %v", c.want, errs)
			}
		})
	}
}

// TestTransformHeadersWarnings covers what is passed through but probably not
// meant: a key beside expressions (which a later connector release might read,
// so it is not refused) and an empty expressions mapping.
func TestTransformHeadersWarnings(t *testing.T) {
	errs, warns := runTransform(t, "expressions:\n  h: \"'a'\"\nexpresions:\n  x: y\n")
	if issueWith(errs, "transform-headers") {
		t.Errorf("an unknown sibling key must not be an error, got %v", errs)
	}
	if !issueWith(warns, "transform-headers.expresions is not a key this tool knows") {
		t.Errorf("want a warning naming the unknown key, got %v", warns)
	}

	errs, warns = runTransform(t, "expressions: {}\n")
	if issueWith(errs, "transform-headers") {
		t.Errorf("an empty expressions mapping must not be an error, got %v", errs)
	}
	if !issueWith(warns, "transform-headers.expressions is empty") {
		t.Errorf("want an empty-expressions warning, got %v", warns)
	}
}

// TestMisplacedTransformsAreErrors pins the answer to "where did my transform
// go": every transform-looking key outside a workflow file's top-level
// transform-headers is an error on every run, naming the file and the path it
// was found at, and saying where it belongs. A transform-headers block in the
// wrong place is told to move; any other key is told it is not a key.
func TestMisplacedTransformsAreErrors(t *testing.T) {
	wfs := wfOK()
	wfs[0].MisplacedTransforms = []string{"source.solace.transform-headers", "transform"}
	d := &spec.Defaults{MisplacedTransforms: []string{"connections.sol.transform-headers"}}
	errs, _ := Run(Context{Workflows: wfs, Defaults: d})

	for _, c := range []struct{ file, want string }{
		{"x.yaml", "source.solace.transform-headers is in the wrong place: header transforms apply to the whole workflow, so transform-headers: goes at the top level of this file"},
		{"x.yaml", "transform is not a key: header transforms are written transform-headers:"},
		{fileEnv, "connections.sol.transform-headers is not read here: a header transform belongs to one workflow"},
	} {
		found := false
		for _, e := range errs {
			if e.File == c.file && strings.Contains(e.Msg, c.want) {
				found = true
			}
		}
		if !found {
			t.Errorf("want an error in %s containing %q, got %v", c.file, c.want, errs)
		}
	}
}
