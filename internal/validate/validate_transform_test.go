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
// go": every transform-looking key outside a workflow file's two top-level
// blocks is an error on every run, naming the file and the path it was found
// at, and saying what to do. A real block in the wrong place is told to move,
// the legacy payload section how payload transforms are written now, any other
// key that it is not one -- naming both valid spellings -- and a transform in
// env.yaml that it belongs in a workflow file.
func TestMisplacedTransformsAreErrors(t *testing.T) {
	wfs := wfOK()
	wfs[0].MisplacedTransforms = []string{"source.solace.transform-headers", "source.solace.transform", "transform-header", "transform-payload", "target.mq.transform-payloads"}
	d := &spec.Defaults{MisplacedTransforms: []string{"connections.sol.transform-headers", "transform"}}
	errs, _ := Run(Context{Workflows: wfs, Defaults: d})

	for _, c := range []struct{ file, want string }{
		{"x.yaml", "source.solace.transform-headers is in the wrong place: header transforms apply to the whole workflow, so transform-headers: goes at the top level of this file"},
		{"x.yaml", "source.solace.transform is in the wrong place: a transform applies to the whole workflow, so transform: goes at the top level of this file"},
		{"x.yaml", "transform-header is not a key: a workflow file's transforms are written transform: (current) or transform-headers: (deprecated) at the top level of this file"},
		{"x.yaml", "transform-payload is not read by this tool: payload transforms are written as target['payload'] expressions in a transform: block"},
		{"x.yaml", "target.mq.transform-payloads is not read by this tool"},
		{fileEnv, "connections.sol.transform-headers is not read here: a transform belongs to one workflow, so write transform: (or the deprecated transform-headers:)"},
		{fileEnv, "transform is not read here: a transform belongs to one workflow"},
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

// runParsed parses text as a workflow file and validates the otherwise clean
// wfOK workflow carrying its transform blocks and misplaced keys, so the
// parser's placement rules are exercised along with the checks.
func runParsed(t *testing.T, text string, lint bool) (errs, warns []Issue) {
	t.Helper()
	p, err := spec.ParseWorkflow([]byte(text), "x.yaml")
	if err != nil {
		t.Fatalf("bad fixture: %v", err)
	}
	wfs := wfOK()
	wfs[0].Transform, wfs[0].TransformHeaders, wfs[0].MisplacedTransforms = p.Transform, p.TransformHeaders, p.MisplacedTransforms
	return Run(Context{Workflows: wfs, Defaults: &spec.Defaults{}, Lint: lint})
}

// transformIssues keeps the findings that mention a transform, so a fixture is
// judged on its transform alone and not on the rest of wfOK.
func transformIssues(issues []Issue) []Issue {
	var out []Issue
	for _, i := range issues {
		if strings.Contains(i.Msg, "transform") {
			out = append(out, i)
		}
	}
	return out
}

// TestTransformLegitimateBlocksStayQuiet runs the transform blocks Solace
// documents -- header copy, payload mapping with variables, functions and the
// dynamic destination header, the IBM MQ migration guide's example, content
// types alone -- plus ${...} placeholders, and pins that none raises anything,
// even under Lint. The migration example carries Solace's enabled: true, which
// is passed through without a word while the key is left unchecked.
func TestTransformLegitimateBlocksStayQuiet(t *testing.T) {
	for _, c := range []struct{ name, text string }{
		{"header copy", `transform:
  expressions:
    - transform: "target['headers']['my-header'] = source['headers']['my-header']"
`},
		{"payload mapping", `transform:
  source-payload:
    content-type: application/json
  target-payload:
    content-type: application/json
  expressions:
    # Set the dynamic destination header from the payload
    - transform: "target['headers']['scst_targetDestination'] = #joinString('/', source['payload']['airline'], source['payload']['destination'], source['payload']['origin'])"
    - transform: "target['payload'] = source['payload']"
    - transform: "var['passengerDistribution'] = #splitString(source['payload']['passengers'], ',', 3)"
    - transform: "target['payload']['passengers'] = {:}"
    - transform: "target['payload']['passengers']['capacity'] = #convertStringToNumber(var['passengerDistribution'][0])"
`},
		{"IBM MQ migration example", `transform:
  enabled: true
  expressions:
    - transform: "target['headers']['route'] = #joinString('/', source['headers']['region'], source['headers']['status'])"
    - transform: "target['headers']['count'] = #convertNumberToString(source['headers']['count'])"
`},
		{"content types alone", `transform:
  source-payload:
    content-type: application/vnd.solace.micro-integration.unspecified
  target-payload:
    content-type: application/json
`},
		{"placeholders", `transform:
  source-payload:
    content-type: ${PAYLOAD_TYPE}
  expressions:
    - transform: "target['headers']['region'] = '${REGION}'"
`},
		{"empty payload values", `transform:
  source-payload:
  target-payload:
    content-type:
  expressions:
    - transform: "target['payload'] = source['payload']"
`},
	} {
		t.Run(c.name, func(t *testing.T) {
			errs, warns := runParsed(t, c.text, true)
			if got := append(transformIssues(errs), transformIssues(warns)...); len(got) != 0 {
				t.Errorf("a documented transform must raise nothing, got %v", got)
			}
		})
	}
}

// TestTransformShapeErrors covers each way a transform block can be shaped so
// that the connector would not apply it, every one an error on every run that
// says what was found and how to write it. The transform-headers shape carried
// over under transform: gets the migration hint.
func TestTransformShapeErrors(t *testing.T) {
	for _, c := range []struct{ name, text, want string }{
		{"not a mapping", "transform: x\n", "transform must be a mapping (source-payload:, target-payload:, expressions:), got a scalar"},
		{"a list", "transform:\n  - transform: x\n", "transform must be a mapping (source-payload:, target-payload:, expressions:), got a list"},
		{"expressions as a mapping", "transform:\n  expressions:\n    h: \"'x'\"\n", "transform.expressions must be a list of - transform: <SpEL expression> items, got a mapping -- that is the transform-headers.expressions shape"},
		{"expressions as a scalar", "transform:\n  expressions: x\n", "transform.expressions must be a list of - transform: <SpEL expression> items, got a scalar"},
		{"a scalar item", "transform:\n  expressions:\n    - x\n", "transform.expressions[0] must be a mapping holding transform: <SpEL expression>, got a scalar"},
		{"an item without transform", "transform:\n  expressions:\n    - expression: x\n", "transform.expressions[0] has no transform: -- each item is - transform: <SpEL expression>"},
		{"a later item", "transform:\n  expressions:\n    - transform: x\n    - [x]\n", "transform.expressions[1] must be a mapping holding transform: <SpEL expression>, got a list"},
		{"a non-scalar expression", "transform:\n  expressions:\n    - transform: {a: b}\n", "transform.expressions[0].transform must be one SpEL expression, got a mapping"},
		{"an empty expression", "transform:\n  expressions:\n    - transform:\n", "transform.expressions[0].transform must be one SpEL expression, got an empty value"},
		{"an unquoted expression cut at its #", "transform:\n  expressions:\n    - transform: target['headers']['d'] = #joinString('/', 'a')\n", `transform.expressions[0].transform ends at "=": unquoted, the expression stops at its first " #"`},
		{"a payload block that is not a mapping", "transform:\n  source-payload: application/json\n", "transform.source-payload must be a mapping holding content-type:, got a scalar"},
		{"a content-type that is not one value", "transform:\n  target-payload:\n    content-type: [a, b]\n", "transform.target-payload.content-type must be one value such as application/json, got a list"},
		{"a key set twice", "transform:\n  expressions:\n    - transform: a\n  expressions:\n    - transform: b\n", `transform sets "expressions" twice; keep one`},
		{"a content-type set twice", "transform:\n  source-payload:\n    content-type: application/json\n    content-type: application/json\n", `transform.source-payload sets "content-type" twice; keep one`},
	} {
		t.Run(c.name, func(t *testing.T) {
			errs, _ := runParsed(t, c.text, false)
			if !issueWith(errs, c.want) {
				t.Errorf("want an error containing %q, got %v", c.want, errs)
			}
		})
	}
}

// TestTransformWarnings covers what is passed through but probably not meant:
// a key this tool does not know at each level (a later connector release might
// read it, so it is not refused), a content type the IBM MQ connector is not
// documented to read, and a block that does nothing. None is an error, and
// enabled -- left unchecked for now -- draws no warning of its own.
func TestTransformWarnings(t *testing.T) {
	for _, c := range []struct{ name, text, want string }{
		{"an unknown key", "transform:\n  expresions:\n    - transform: x\n", "transform.expresions is not a key this tool knows: it is passed through as written, but the connector reads a transform from source-payload, target-payload and expressions -- check the spelling"},
		{"an unknown key in a payload block", "transform:\n  source-payload:\n    content-typ: application/json\n", "transform.source-payload.content-typ is not a key this tool knows: it is passed through as written, but the connector reads only content-type here -- check the spelling"},
		{"an unknown key in an item", "transform:\n  expressions:\n    - transform: x\n      description: y\n", "transform.expressions[0].description is not a key this tool knows: it is passed through as written, but the connector reads one SpEL expression per item from transform -- check the spelling"},
		{"an undocumented content type", "transform:\n  source-payload:\n    content-type: text/plain\n", `transform.source-payload.content-type "text/plain" is not a value documented for the IBM MQ connector (application/json, application/vnd.solace.micro-integration.unspecified); it is passed through as written`},
		{"xml", "transform:\n  target-payload:\n    content-type: application/xml\n", "it is passed through as written -- the IBM MQ connector's release notes support JSON payloads only"},
		{"empty expressions", "transform:\n  expressions: []\n", "transform.expressions is empty, so this workflow applies no transform expressions"},
		{"an empty block", "transform: {}\n", "transform sets no expressions and no payload content-type, so it does nothing"},
		{"enabled alone", "transform:\n  enabled: true\n", "transform sets no expressions and no payload content-type, so it does nothing"},
	} {
		t.Run(c.name, func(t *testing.T) {
			errs, warns := runParsed(t, c.text, false)
			if got := transformIssues(errs); len(got) != 0 {
				t.Errorf("want no transform error, got %v", got)
			}
			if !issueWith(warns, c.want) {
				t.Errorf("want a warning containing %q, got %v", c.want, warns)
			}
			if issueWith(warns, "transform.enabled") {
				t.Errorf("enabled is passed through unchecked, got %v", warns)
			}
		})
	}
}

// TestTransformAndTransformHeadersCannotBeCombined pins Solace's rule that the
// current and the legacy section cannot be used together: both in one file is
// an error on every run, quoting the rule and naming the migration, while each
// block's own shape is still checked. Either alone is fine.
func TestTransformAndTransformHeadersCannotBeCombined(t *testing.T) {
	const both = "cannot use the two together"
	transform := "transform:\n  expressions:\n    - transform: \"target['headers']['h'] = 'x'\"\n"
	headers := "transform-headers:\n  expressions:\n    h: \"'x'\"\n"

	errs, _ := runParsed(t, transform+headers, false)
	if !issueWith(errs, both) || !issueWith(errs, `"The previous and current configuration sections cannot be used together."`) {
		t.Errorf("both blocks must be an error quoting Solace's rule, got %v", errs)
	}
	errs, _ = runParsed(t, transform+"transform-headers: x\n", false)
	if !issueWith(errs, both) || !issueWith(errs, "transform-headers must be a mapping") {
		t.Errorf("each block's own shape is still checked beside the pair, got %v", errs)
	}
	for name, text := range map[string]string{"transform alone": transform, "transform-headers alone": headers} {
		if errs, _ := runParsed(t, text, false); issueWith(errs, both) {
			t.Errorf("%s: want no pairing error, got %v", name, errs)
		}
	}
}

// TestTransformHeadersDeprecatedOnlyUnderLint pins where Solace's deprecation
// of transform-headers is reported: nowhere on a generate or deploy run, since
// the connector still reads it, and exactly once under Lint (the validate
// verb), naming the timeline and where the migration is described. A file on
// transform: alone draws nothing; one with both blocks still gets the notice.
func TestTransformHeadersDeprecatedOnlyUnderLint(t *testing.T) {
	const deprecated = "transform-headers is deprecated by Solace"
	headers := "transform-headers:\n  expressions:\n    h: \"'x'\"\n"
	transform := "transform:\n  expressions:\n    - transform: x\n"
	count := func(warns []Issue) int {
		n := 0
		for _, w := range warns {
			if strings.Contains(w.Msg, deprecated) {
				n++
			}
		}
		return n
	}

	if _, warns := runParsed(t, headers, false); count(warns) != 0 {
		t.Errorf("generate and deploy must stay quiet about the deprecation, got %v", warns)
	}
	_, warns := runParsed(t, headers, true)
	if count(warns) != 1 || !issueWith(warns, "since connector 2.9.0") || !issueWith(warns, "user guide section 6.7") {
		t.Errorf("validate must warn once, naming the timeline and section 6.7, got %v", warns)
	}
	if _, warns := runParsed(t, transform, true); count(warns) != 0 {
		t.Errorf("a transform-only file must draw no deprecation notice, got %v", warns)
	}
	if _, warns := runParsed(t, transform+headers, true); count(warns) != 1 {
		t.Errorf("both blocks under Lint still draw the notice, got %v", warns)
	}
}
