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

// TestMisplacedWorkflowTransforms covers every place a header transform
// plausibly lands by mistake. The file decodes without KnownFields, so each of
// these would otherwise be dropped in silence: the retired transform: key, a
// misspelling, and transform-headers nested under a side, its solace:/mq:
// block, or that block's consumer:/producer: tuning. Each is reported as the
// dotted path it was found at, in file order.
func TestMisplacedWorkflowTransforms(t *testing.T) {
	wf, err := ParseWorkflow([]byte(`transform:
  x: y
transform-header:
  expressions: {a: b}
source:
  transform-headers: {expressions: {a: b}}
  solace:
    queue: IN
    transform-headers: {expressions: {a: b}}
    consumer:
      transform-headers: {expressions: {a: b}}
target:
  mq:
    queue: OUT
    producer:
      transforms: x
transform-headers:
  expressions: {a: b}
`), "wf.yaml")
	if err != nil {
		t.Fatalf("ParseWorkflow: %v", err)
	}
	want := []string{
		"transform",
		"transform-header",
		"source.transform-headers",
		"source.solace.transform-headers",
		"source.solace.consumer.transform-headers",
		"target.mq.producer.transforms",
	}
	if !reflect.DeepEqual(wf.MisplacedTransforms, want) {
		t.Errorf("misplaced = %v, want %v", wf.MisplacedTransforms, want)
	}
	// The top-level block still parses: the misplaced ones sit beside it.
	if wf.TransformHeaders == nil {
		t.Error("the top-level transform-headers must still be captured")
	}
}

// TestMisplacedEnvTransforms covers env.yaml, where no transform belongs at
// all: a header transform is per workflow. Its top level and each connection,
// with that connection's solace:/mq: block, are scanned -- through both
// ParseEnv and ParseDefaults, so the two cannot disagree.
func TestMisplacedEnvTransforms(t *testing.T) {
	doc := []byte(`transform-headers:
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
	want := []string{"transform-headers", "connections.sol.transform", "connections.sol.solace.transform-headers"}
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
