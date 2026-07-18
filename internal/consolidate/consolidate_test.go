package consolidate

import (
	"testing"

	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-gen/internal/spec"
)

func solaceSide(vpn, dest, destKind, keyAlias string) spec.Side {
	return spec.Side{
		System: spec.SystemSolace, Host: "tcps://broker:55443", MsgVPN: vpn,
		ClientUser: "connector", ClientPass: "${SOL}", DestKind: destKind, Dest: dest, KeyAlias: keyAlias,
	}
}

func mqSide(qm, dest, destKind string, tls bool) spec.Side {
	return spec.Side{
		System: spec.SystemMQ, ConnName: "host(1414)", QueueManager: qm, Channel: "CH",
		User: "app", Password: "${MQ}", TLS: tls, DestKind: destKind, Dest: dest,
	}
}

func binderNames(m *Model) []string {
	var out []string
	for _, b := range m.Binders {
		out = append(out, b.Name)
	}
	return out
}

func binderOf(m *Model, binding string) string {
	for _, b := range m.Bindings {
		if b.Name == binding {
			return b.Binder
		}
	}
	return ""
}

func eqStrs(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestBinderDedupAcrossWorkflows(t *testing.T) {
	wfs := []spec.Workflow{
		{File: "10.yaml", Enabled: true, SourceSet: true, TargetSet: true,
			Source: solaceSide("prod", "Q1", spec.DestQueue, ""), Target: mqSide("QM1", "MQ1", spec.DestQueue, false)},
		{File: "20.yaml", Enabled: true, SourceSet: true, TargetSet: true,
			Source: mqSide("QM1", "MQ2", spec.DestQueue, false), Target: solaceSide("prod", "Q2", spec.DestQueue, "")},
	}
	m, _ := Build(wfs, nil, true)
	// Inline (no conn-ref) sides get generated names, numbered per system by appearance.
	if got, want := binderNames(m), []string{"sol-conn-1", "mq-conn-1", "undefined"}; !eqStrs(got, want) {
		t.Fatalf("binders = %v, want %v", got, want)
	}
}

func TestSolaceToSolaceSingleBinder(t *testing.T) {
	wfs := []spec.Workflow{{File: "10.yaml", Enabled: true, SourceSet: true, TargetSet: true,
		Source: solaceSide("prod", "IN", spec.DestQueue, ""), Target: solaceSide("prod", "OUT", spec.DestQueue, "")}}
	m, _ := Build(wfs, nil, true)
	if got, want := binderNames(m), []string{"sol-conn-1", "undefined"}; !eqStrs(got, want) {
		t.Fatalf("binders = %v, want %v", got, want)
	}
	if in, out := binderOf(m, "input-0"), binderOf(m, "output-0"); in != "sol-conn-1" || out != "sol-conn-1" {
		t.Fatalf("both sides should reference sol-conn-1, got input=%q output=%q", in, out)
	}
}

func TestMQToMQSingleBinder(t *testing.T) {
	wfs := []spec.Workflow{{File: "10.yaml", Enabled: true, SourceSet: true, TargetSet: true,
		Source: mqSide("QM1", "IN", spec.DestQueue, false), Target: mqSide("QM1", "OUT", spec.DestQueue, false)}}
	m, _ := Build(wfs, nil, true)
	if got, want := binderNames(m), []string{"mq-conn-1", "undefined"}; !eqStrs(got, want) {
		t.Fatalf("binders = %v, want %v", got, want)
	}
	if in, out := binderOf(m, "input-0"), binderOf(m, "output-0"); in != "mq-conn-1" || out != "mq-conn-1" {
		t.Fatalf("both sides should reference mq-conn-1, got input=%q output=%q", in, out)
	}
}

func TestConnRefNamingAndClashSuffix(t *testing.T) {
	// Two differently-keyed connections that sanitize to the same binder base.
	d := &spec.Defaults{Connections: map[string]spec.Side{
		"svc.a": {System: spec.SystemSolace, Host: "tcps://h1", MsgVPN: "v1", ClientUser: "u", ClientPass: "p"},
		"svc/a": {System: spec.SystemSolace, Host: "tcps://h2", MsgVPN: "v2", ClientUser: "u", ClientPass: "p"},
	}}
	wfs := []spec.Workflow{{File: "10.yaml", Enabled: true, SourceSet: true, TargetSet: true,
		Source: spec.Side{System: spec.SystemSolace, ConnRef: "svc.a", DestKind: spec.DestQueue, Dest: "A"},
		Target: spec.Side{System: spec.SystemSolace, ConnRef: "svc/a", DestKind: spec.DestQueue, Dest: "B"}}}
	m, _ := Build(wfs, d, true)
	// Binders take the sanitized connection name; the second (same base) is disambiguated.
	if got, want := binderNames(m), []string{"svc-a", "svc-a-2", "undefined"}; !eqStrs(got, want) {
		t.Fatalf("clash names = %v", got)
	}
}

func TestConnRefDedupCollapsesToOneBinder(t *testing.T) {
	// Two workflows referencing the same connection collapse to a single binder.
	d := &spec.Defaults{Connections: map[string]spec.Side{
		"edge": {System: spec.SystemSolace, Host: "tcps://h", MsgVPN: "v", ClientUser: "u", ClientPass: "p"},
		"qm":   {System: spec.SystemMQ, ConnName: "h(1414)", QueueManager: "QM", Channel: "C", User: "u", Password: "p"},
	}}
	ref := func(sys, ref, dest string) spec.Side {
		return spec.Side{System: sys, ConnRef: ref, DestKind: spec.DestQueue, Dest: dest}
	}
	wfs := []spec.Workflow{
		{File: "10.yaml", Enabled: true, SourceSet: true, TargetSet: true, Source: ref(spec.SystemSolace, "edge", "A"), Target: ref(spec.SystemMQ, "qm", "B")},
		{File: "20.yaml", Enabled: true, SourceSet: true, TargetSet: true, Source: ref(spec.SystemSolace, "edge", "C"), Target: ref(spec.SystemMQ, "qm", "D")},
	}
	m, _ := Build(wfs, d, true)
	if got, want := binderNames(m), []string{"edge", "qm", "undefined"}; !eqStrs(got, want) {
		t.Fatalf("binders = %v, want %v", got, want)
	}
}

func TestDerivedDestinationTypes(t *testing.T) {
	wfs := []spec.Workflow{
		// MQ topic consumer (durable) -> Solace topic producer (no solace binding emitted).
		{File: "10.yaml", Enabled: true, SourceSet: true, TargetSet: true,
			Source: mqSide("QM1", "MQ/T", spec.DestTopic, false), Target: solaceSide("prod", "s/t", spec.DestTopic, "")},
		// Solace queue -> MQ queue.
		{File: "20.yaml", Enabled: true, SourceSet: true, TargetSet: true,
			Source: solaceSide("prod", "Q.IN", spec.DestQueue, ""), Target: mqSide("QM1", "MQ.OUT", spec.DestQueue, false)},
	}
	m, _ := Build(wfs, nil, true)

	// jms input-0 consumer topic + durable
	var in0 *JMSBinding
	for _, jb := range m.JMSBindings {
		if jb.Name == "input-0" {
			in0 = jb
		}
	}
	if in0 == nil || in0.Role != "consumer" || in0.DestType != spec.DestTopic || in0.Durable == "" {
		t.Fatalf("input-0 jms = %+v, want consumer/topic/durable", in0)
	}
	// solace producer topic -> no solace binding for output-0
	for _, sb := range m.SolaceBindings {
		if sb.Name == "output-0" {
			t.Fatalf("output-0 solace producer->topic should emit nothing, got %+v", sb)
		}
	}
	// mq producer queue -> jms output-1 producer queue
	var out1 *JMSBinding
	for _, jb := range m.JMSBindings {
		if jb.Name == "output-1" {
			out1 = jb
		}
	}
	if out1 == nil || out1.Role != "producer" || out1.DestType != spec.DestQueue {
		t.Fatalf("output-1 jms = %+v, want producer/queue", out1)
	}
}
