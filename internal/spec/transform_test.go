package spec

import (
	"reflect"
	"testing"

	"gopkg.in/yaml.v3"
)

// transformFile is a minimal valid workflow file carrying extra.
func transformFile(extra string) []byte {
	return []byte("source:\n  solace:\n    queue: IN\ntarget:\n  mq:\n    queue: OUT\n" + extra)
}

// TestParseWorkflowTransformHeaders pins the one right place for a header
// transform: the workflow file's top level. The block is captured verbatim --
// key order and each expression's quoting intact -- because it is passed
// through to the connector as written. An empty block is the same as none.
func TestParseWorkflowTransformHeaders(t *testing.T) {
	wf, err := ParseWorkflow(transformFile(`transform-headers:
  expressions:
    solace_scst_targetDestination: "'orders/' + headers.region"
    JMS_IBM_Format: 'MQSTR'
`), "wf.yaml")
	if err != nil {
		t.Fatalf("ParseWorkflow: %v", err)
	}
	th := wf.TransformHeaders
	if th == nil || th.Kind != yaml.MappingNode || len(th.Content) != 2 || th.Content[0].Value != "expressions" {
		t.Fatalf("transform-headers = %+v, want a mapping holding expressions", th)
	}
	exprs := th.Content[1]
	if got := []string{exprs.Content[0].Value, exprs.Content[2].Value}; !reflect.DeepEqual(got, []string{"solace_scst_targetDestination", "JMS_IBM_Format"}) {
		t.Errorf("header order = %v, want the file's own order", got)
	}
	if v := exprs.Content[1]; v.Value != "'orders/' + headers.region" || v.Style != yaml.DoubleQuotedStyle {
		t.Errorf("expression = %q (style %v), want it exactly as written, double-quoted", v.Value, v.Style)
	}
	if len(wf.MisplacedTransforms) != 0 {
		t.Errorf("a top-level transform-headers is in the right place, got misplaced %v", wf.MisplacedTransforms)
	}

	for name, extra := range map[string]string{"absent": "", "left empty": "transform-headers:\n"} {
		wf, err := ParseWorkflow(transformFile(extra), "wf.yaml")
		if err != nil {
			t.Fatalf("%s: ParseWorkflow: %v", name, err)
		}
		if wf.TransformHeaders != nil {
			t.Errorf("%s: transform-headers = %+v, want none", name, wf.TransformHeaders)
		}
	}
}

// TestParseWorkflowTransform pins where the current transform section is
// read: the workflow file's top level, beside transform-headers. Like that
// block it is captured verbatim -- key order, the list of one-key mappings and
// each expression's quoting intact -- and neither is reported as misplaced;
// refusing the pair is validate's job. An empty block is the same as none.
func TestParseWorkflowTransform(t *testing.T) {
	wf, err := ParseWorkflow(transformFile(`transform:
  source-payload:
    content-type: application/json
  target-payload:
    content-type: application/json
  expressions:
    - transform: "target['headers']['region'] = source['headers']['region']"
    - transform: 'target[''payload''] = source[''payload'']'
transform-headers:
  expressions:
    h: "'x'"
`), "wf.yaml")
	if err != nil {
		t.Fatalf("ParseWorkflow: %v", err)
	}
	tr := wf.Transform
	if tr == nil || tr.Kind != yaml.MappingNode || len(tr.Content) != 6 {
		t.Fatalf("transform = %+v, want a mapping of three keys", tr)
	}
	if got := []string{tr.Content[0].Value, tr.Content[2].Value, tr.Content[4].Value}; !reflect.DeepEqual(got, []string{"source-payload", "target-payload", "expressions"}) {
		t.Errorf("transform keys = %v, want the file's own order", got)
	}
	exprs := tr.Content[5]
	if exprs.Kind != yaml.SequenceNode || len(exprs.Content) != 2 {
		t.Fatalf("expressions = %+v, want a list of two items", exprs)
	}
	first, second := exprs.Content[0], exprs.Content[1]
	if first.Kind != yaml.MappingNode || first.Content[0].Value != "transform" || first.Content[1].Style != yaml.DoubleQuotedStyle ||
		first.Content[1].Value != "target['headers']['region'] = source['headers']['region']" {
		t.Errorf("first item = %+v, want transform: with its double-quoted expression exactly as written", first)
	}
	if v := second.Content[1]; v.Style != yaml.SingleQuotedStyle || v.Value != "target['payload'] = source['payload']" {
		t.Errorf("second expression = %q (style %v), want it single-quoted as written", v.Value, v.Style)
	}
	if wf.TransformHeaders == nil {
		t.Error("transform-headers beside it must be captured too")
	}
	if len(wf.MisplacedTransforms) != 0 {
		t.Errorf("a top-level transform and transform-headers are in the right place, got misplaced %v", wf.MisplacedTransforms)
	}

	for name, extra := range map[string]string{"absent": "", "left empty": "transform:\n"} {
		wf, err := ParseWorkflow(transformFile(extra), "wf.yaml")
		if err != nil {
			t.Fatalf("%s: ParseWorkflow: %v", name, err)
		}
		if wf.Transform != nil {
			t.Errorf("%s: transform = %+v, want none", name, wf.Transform)
		}
	}
}

// TestMisplacedWorkflowTransforms covers every place a transform plausibly
// lands by mistake. The file decodes without KnownFields, so each of these
// would otherwise be dropped in silence: a misspelling, the legacy
// transform-payload section (either spelling), and transform: or
// transform-headers: nested under a side, its solace:/mq: block, or that
// block's consumer:/producer: tuning. Each is reported as the dotted path it
// was found at, in file order; the two top-level blocks are not.
func TestMisplacedWorkflowTransforms(t *testing.T) {
	wf, err := ParseWorkflow([]byte(`x-side: &side
  transform-header: x
x-mq: &mq
  <<: *side
  queue: OUT
  producer:
    transforms: x
transform:
  expressions:
    - transform: "target['headers']['h'] = 'x'"
transform-header:
  expressions: {a: b}
transform-payload:
  expressions:
    - transform: x
source:
  transform: {expressions: []}
  transform-headers: {expressions: {a: b}}
  solace:
    queue: IN
    transform-headers: {expressions: {a: b}}
    consumer:
      transform: x
      transform-payloads: x
target:
  mq: *mq
transform-headers:
  expressions: {a: b}
`), "wf.yaml")
	if err != nil {
		t.Fatalf("ParseWorkflow: %v", err)
	}
	want := []string{
		"transform-header",
		"transform-payload",
		"source.transform",
		"source.transform-headers",
		"source.solace.transform-headers",
		"source.solace.consumer.transform",
		"source.solace.consumer.transform-payloads",
		"target.mq.transform-header",
		"target.mq.producer.transforms",
	}
	if !reflect.DeepEqual(wf.MisplacedTransforms, want) {
		t.Errorf("misplaced = %v, want %v", wf.MisplacedTransforms, want)
	}
	// The top-level blocks still parse: the misplaced ones sit beside them.
	if wf.Transform == nil || wf.TransformHeaders == nil {
		t.Error("the top-level transform and transform-headers must still be captured")
	}
}

// TestMisplacedEnvTransforms covers env.yaml, where no transform belongs at
// all: a transform is per workflow. Its top level and each connection, with
// that connection's solace:/mq: block, are scanned -- through both ParseEnv
// and ParseDefaults, so the two cannot disagree.
func TestMisplacedEnvTransforms(t *testing.T) {
	doc := []byte(`transform:
  expressions: []
transform-headers:
  expressions: {a: b}
connections:
  sol:
    transform: x
    solace:
      host: tcp://b
      transform-headers: {expressions: {a: b}}
  mq1:
    mq:
      conn-name: h(1414)
`)
	want := []string{"transform", "transform-headers", "connections.sol.transform", "connections.sol.solace.transform-headers"}
	e, err := ParseEnv(doc)
	if err != nil {
		t.Fatalf("ParseEnv: %v", err)
	}
	if !reflect.DeepEqual(e.MisplacedTransforms, want) {
		t.Errorf("ParseEnv misplaced = %v, want %v", e.MisplacedTransforms, want)
	}
	d, err := ParseDefaults(doc)
	if err != nil {
		t.Fatalf("ParseDefaults: %v", err)
	}
	if !reflect.DeepEqual(d.MisplacedTransforms, want) {
		t.Errorf("ParseDefaults misplaced = %v, want %v", d.MisplacedTransforms, want)
	}

	clean, err := ParseEnv([]byte("timezone: UTC\nconnections:\n  sol:\n    solace:\n      host: tcp://b\n"))
	if err != nil {
		t.Fatalf("ParseEnv: %v", err)
	}
	if len(clean.MisplacedTransforms) != 0 {
		t.Errorf("an env.yaml without transform keys reports %v", clean.MisplacedTransforms)
	}
	if empty, _ := ParseEnv(nil); len(empty.MisplacedTransforms) != 0 {
		t.Errorf("an empty env.yaml reports %v", empty.MisplacedTransforms)
	}
}
