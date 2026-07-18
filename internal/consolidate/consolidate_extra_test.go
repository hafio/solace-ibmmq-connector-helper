package consolidate

import (
	"strings"
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-gen/internal/spec"
)

func containsSub(ss []string, sub string) bool {
	for _, s := range ss {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}

func TestFormatScalarQuoting(t *testing.T) {
	if got := formatScalar(&yaml.Node{Kind: yaml.ScalarNode, Value: "plain"}); got != "plain" {
		t.Errorf("plain=%q", got)
	}
	if got := formatScalar(&yaml.Node{Kind: yaml.ScalarNode, Value: "dq", Style: yaml.DoubleQuotedStyle}); got != `"dq"` {
		t.Errorf("double=%q", got)
	}
	if got := formatScalar(&yaml.Node{Kind: yaml.ScalarNode, Value: "a'b", Style: yaml.SingleQuotedStyle}); got != `'a''b'` {
		t.Errorf("single=%q", got)
	}
}

func TestSanitizeAndIsTCPS(t *testing.T) {
	if sanitize("A_b.2-x/y") != "A-b-2-x-y" {
		t.Errorf("sanitize=%q", sanitize("A_b.2-x/y"))
	}
	if !isTCPS("tcps://x") || isTCPS("tcp://x") {
		t.Error("isTCPS")
	}
}

func TestDisplayName(t *testing.T) {
	if displayName(&acc{name: "given"}) != "given" {
		t.Error("displayName should return name when set")
	}
	if displayName(&acc{connName: "my-conn"}) != "my-conn" {
		t.Error("displayName should fall back to the connection name")
	}
	if displayName(&acc{kind: spec.SystemMQ, qm: "QM1"}) != "mq:QM1" {
		t.Error("displayName tuple fallback")
	}
}

func TestMergeProp(t *testing.T) {
	var warns []string
	list := []Prop{{Key: "a", Val: "1"}}
	list = mergeProp(list, Prop{Key: "b", Val: "2"}, &warns, "bndr")
	if len(list) != 2 || len(warns) != 0 {
		t.Fatalf("append: %v warns=%v", list, warns)
	}
	list = mergeProp(list, Prop{Key: "a", Val: "9"}, &warns, "bndr")
	if list[0].Val != "9" || len(warns) != 1 {
		t.Fatalf("overwrite: %v warns=%v", list, warns)
	}
	list = mergeProp(list, Prop{Key: "a", Val: "9"}, &warns, "bndr")
	if len(warns) != 1 {
		t.Fatalf("same value should not warn again: %v", warns)
	}
}

func TestAppendPassthroughCollision(t *testing.T) {
	var warns []string
	out := appendPassthrough([]Prop{{Key: "T", Val: "x"}}, []Prop{{Key: "P", Val: "y"}, {Key: "T", Val: "z"}}, &warns, "bndr")
	if len(out) != 2 || len(warns) != 1 {
		t.Fatalf("out=%v warns=%v", out, warns)
	}
}

func TestNodeToProps(t *testing.T) {
	var n yaml.Node
	if err := yaml.Unmarshal([]byte("k1: v1\nk2:\n  nested: 1\n"), &n); err != nil {
		t.Fatal(err)
	}
	props := nodeToProps(n.Content[0])
	if len(props) != 2 || props[0].Key != "k1" || props[0].Val != "v1" || props[1].Sub == nil {
		t.Fatalf("props=%+v", props)
	}
	if nodeToProps(nil) != nil {
		t.Error("nil -> nil")
	}
}

func TestBuildMQmTLSBundle(t *testing.T) {
	mq := spec.Side{System: spec.SystemMQ, ConnName: "h(1414)", QueueManager: "QM1", Channel: "CH", User: "u", Password: "${MQ}", TLS: true, Cipher: "CIPH", KeyAlias: "mc", DestKind: spec.DestQueue, Dest: "IN"}
	sol := spec.Side{System: spec.SystemSolace, Host: "tcps://b", MsgVPN: "v", ClientUser: "u", ClientPass: "${S}", DestKind: spec.DestQueue, Dest: "OUT"}
	d := &spec.Defaults{TLS: spec.TLSConfig{
		Truststore: &spec.Store{File: "t.jks", Password: "${T}", Type: "JKS"},
		Keystore:   &spec.Store{File: "k.jks", Password: "${K}", Type: "PKCS12"},
	}}
	m, _ := Build([]spec.Workflow{{File: "10.yaml", Enabled: true, SourceSet: true, TargetSet: true, Source: mq, Target: sol}}, d, true)
	if !m.MQTLS || len(m.Bundles) != 1 {
		t.Fatalf("MQTLS=%v bundles=%d", m.MQTLS, len(m.Bundles))
	}
	b := m.Bundles[0]
	if !b.HasKeystore || b.KeyAlias != "mc" || b.KeystoreTyp != "PKCS12" || b.TruststoreTyp != "JKS" {
		t.Fatalf("bundle=%+v", b)
	}
}

func TestBuildCipherConflictWarning(t *testing.T) {
	mk := func(cipher, dest string) spec.Side {
		return spec.Side{System: spec.SystemMQ, ConnName: "h(1)", QueueManager: "QM1", Channel: "CH", User: "u", Password: "p", TLS: true, Cipher: cipher, DestKind: spec.DestQueue, Dest: dest}
	}
	sol := spec.Side{System: spec.SystemSolace, Host: "tcp://b", MsgVPN: "v", ClientUser: "u", ClientPass: "p", DestKind: spec.DestQueue, Dest: "X"}
	wfs := []spec.Workflow{
		{File: "10.yaml", Enabled: true, SourceSet: true, TargetSet: true, Source: mk("C1", "A"), Target: sol},
		{File: "20.yaml", Enabled: true, SourceSet: true, TargetSet: true, Source: mk("C2", "B"), Target: sol},
	}
	_, warns := Build(wfs, nil, true)
	if !containsSub(warns, "conflicting cipher") {
		t.Fatalf("want cipher conflict, got %v", warns)
	}
}

func TestBuildMessageLoopWarning(t *testing.T) {
	s := spec.Side{System: spec.SystemMQ, ConnName: "h(1)", QueueManager: "QM1", Channel: "CH", User: "u", Password: "p", DestKind: spec.DestQueue, Dest: "SAME"}
	_, warns := Build([]spec.Workflow{{File: "10.yaml", Enabled: true, SourceSet: true, TargetSet: true, Source: s, Target: s}}, nil, true)
	if !containsSub(warns, "message loop") {
		t.Fatalf("want loop warn, got %v", warns)
	}
}

func TestBuildMQTopicSourceAlwaysDurable(t *testing.T) {
	src := spec.Side{System: spec.SystemMQ, ConnName: "h(1)", QueueManager: "QM1", Channel: "CH", User: "u", Password: "p", DestKind: spec.DestTopic, Dest: "T"}
	tgt := spec.Side{System: spec.SystemSolace, Host: "tcp://b", MsgVPN: "v", ClientUser: "u", ClientPass: "p", DestKind: spec.DestQueue, Dest: "OUT"}
	m, _ := Build([]spec.Workflow{{File: "10.yaml", Enabled: true, SourceSet: true, TargetSet: true, Source: src, Target: tgt}}, nil, true)
	var in0 *JMSBinding
	for _, jb := range m.JMSBindings {
		if jb.Name == "input-0" {
			in0 = jb
		}
	}
	if in0 == nil || in0.Durable == "" {
		t.Fatalf("MQ topic source must always be durable, got %+v", in0)
	}
}

func TestBuildSolaceTopicSourceEmitsConsumerTopic(t *testing.T) {
	src := spec.Side{System: spec.SystemSolace, Host: "tcps://b", MsgVPN: "v", ClientUser: "u", ClientPass: "p", DestKind: spec.DestTopic, Dest: "evt/in"}
	tgt := spec.Side{System: spec.SystemMQ, ConnName: "h(1)", QueueManager: "QM", Channel: "C", User: "u", Password: "p", DestKind: spec.DestQueue, Dest: "OUT"}
	m, _ := Build([]spec.Workflow{{File: "10.yaml", Enabled: true, SourceSet: true, TargetSet: true, Source: src, Target: tgt}}, nil, true)
	var in0 *SolaceBinding
	for _, sb := range m.SolaceBindings {
		if sb.Name == "input-0" {
			in0 = sb
		}
	}
	if in0 == nil || in0.Role != "consumer" || in0.DestType != spec.DestTopic {
		t.Fatalf("Solace topic source should emit consumer destination-type=topic, got %+v", in0)
	}
}

func TestBuildStorePathsRawVsMount(t *testing.T) {
	mq := spec.Side{System: spec.SystemMQ, ConnName: "h(1)", QueueManager: "QM", Channel: "C", User: "u", Password: "p", TLS: true, KeyAlias: "mc", DestKind: spec.DestQueue, Dest: "IN"}
	sol := spec.Side{System: spec.SystemSolace, Host: "tcps://b", MsgVPN: "v", ClientUser: "u", ClientPass: "p", DestKind: spec.DestQueue, Dest: "OUT"}
	d := &spec.Defaults{TLS: spec.TLSConfig{
		Truststore: &spec.Store{File: "./certs/t.jks", Password: "${T}", Type: "JKS"},
		Keystore:   &spec.Store{File: "./certs/k.jks", Password: "${K}", Type: "JKS"},
	}}
	wfs := []spec.Workflow{{File: "10.yaml", Enabled: true, SourceSet: true, TargetSet: true, Source: mq, Target: sol}}
	raw, _ := Build(wfs, d, false) // config: reflect the defaults.yaml path verbatim
	if raw.Bundles[0].TruststoreLoc != "./certs/t.jks" {
		t.Fatalf("config (mount=false) truststore loc = %q, want ./certs/t.jks", raw.Bundles[0].TruststoreLoc)
	}
	mnt, _ := Build(wfs, d, true) // deploy: rewrite to the container mount path
	if mnt.Bundles[0].TruststoreLoc != "/app/external/classpath/truststores/t.jks" {
		t.Fatalf("deploy (mount=true) truststore loc = %q", mnt.Bundles[0].TruststoreLoc)
	}
}
