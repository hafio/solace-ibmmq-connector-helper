package consolidate

import (
	"strings"
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/spec"
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
	cases := []struct {
		name string
		node *yaml.Node
		want string
	}{
		{"plain passthrough", &yaml.Node{Kind: yaml.ScalarNode, Value: "plain"}, "plain"},
		{"double-quoted requoting", &yaml.Node{Kind: yaml.ScalarNode, Value: "dq", Style: yaml.DoubleQuotedStyle}, `"dq"`},
		{"single-quote-doubling escape", &yaml.Node{Kind: yaml.ScalarNode, Value: "a'b", Style: yaml.SingleQuotedStyle}, `'a''b'`},
	}
	for _, c := range cases {
		if got := FormatScalar(c.node); got != c.want {
			t.Errorf("%s: FormatScalar=%q, want %q", c.name, got, c.want)
		}
	}

	// depth>=2 quote-preserving passthrough: FormatScalar is also called by the
	// render package while walking a nested (Sub) passthrough tree, so it must
	// preserve quoting on a scalar reached two or more levels below the map that
	// owns it, not just on the depth-1 scalars nodeToProps formats directly.
	// Parse real YAML (not a hand-built node) so the nested node's Style comes
	// from the same path render.renderContainer walks.
	var doc yaml.Node
	if err := yaml.Unmarshal([]byte("consumer:\n  retry:\n    label: \"R1\"\n"), &doc); err != nil {
		t.Fatal(err)
	}
	// doc -> DocumentNode; Content[0] -> top mapping {consumer: ...}.
	consumerVal := doc.Content[0].Content[1] // consumer's value: {retry: ...}
	retryVal := consumerVal.Content[1]       // retry's value: {label: "R1"}
	labelVal := retryVal.Content[1]          // label's value node (depth 2 below consumer)
	if got := FormatScalar(labelVal); got != `"R1"` {
		t.Errorf("depth>=2 nested quoted passthrough: FormatScalar=%q, want %q", got, `"R1"`)
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
	if got := displayName(&acc{kind: spec.SystemSolace, vpn: "myvpn"}); got != "solace:myvpn" {
		t.Errorf("displayName solace tuple fallback = %q, want solace:myvpn", got)
	}
	if got := displayName(&acc{kind: spec.SystemMQ, qm: "QM1"}); got != "mq:QM1" {
		t.Errorf("displayName mq tuple fallback = %q, want mq:QM1", got)
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
	raw, _ := Build(wfs, d, false) // config: reflect the env.yaml path verbatim
	if raw.Bundles[0].TruststoreLoc != "./certs/t.jks" {
		t.Fatalf("config (mount=false) truststore loc = %q, want ./certs/t.jks", raw.Bundles[0].TruststoreLoc)
	}
	mnt, _ := Build(wfs, d, true) // deploy: rewrite to the container mount path
	if mnt.Bundles[0].TruststoreLoc != "/app/external/classpath/truststores/t.jks" {
		t.Fatalf("deploy (mount=true) truststore loc = %q", mnt.Bundles[0].TruststoreLoc)
	}
}

// trustStoreVal extracts SSL_TRUST_STORE from a leader-election session's
// APIProps ("" if absent).
func trustStoreVal(le *LeaderElectionModel) string {
	for _, p := range le.Session.APIProps {
		if p.Key == "SSL_TRUST_STORE" {
			return p.Val
		}
	}
	return ""
}

// TestBuildLeaderElection pins buildLeaderElection's branches directly. It was
// previously exercised only cross-package (render_test.go's
// TestApplicationLeaderElection and gen_extra_test.go's
// TestBuildShardsLeaderQueueSuffix); this moves ownership of those branches
// into the package that defines them.
func TestBuildLeaderElection(t *testing.T) {
	// dangling conn-ref: validate.go:291 (checkLeaderElection) catches an unknown
	// conn-ref before Build ever runs; buildLeaderElection itself has no guard and
	// emits no warning, so a dangling ref silently leaves Session nil.
	dangling := &spec.Defaults{LeaderElection: spec.LeaderElection{
		Present: true, Mode: spec.LeaderActiveStby, Queue: "mgmt-q", ConnRef: "missing",
	}}
	le := buildLeaderElection(dangling, true)
	if le == nil || le.Mode != spec.LeaderActiveStby || le.Queue != "mgmt-q" {
		t.Fatalf("dangling conn-ref: le = %+v", le)
	}
	if le.Session != nil {
		t.Errorf("dangling conn-ref: Session = %+v, want nil", le.Session)
	}

	// conn-ref happy path with a tcps:// host: the TLS APIProps branch populates
	// (SSL trust-store + mTLS key-store props from the shared truststore/keystore).
	happy := &spec.Defaults{
		LeaderElection: spec.LeaderElection{Present: true, Mode: spec.LeaderActiveActive, Queue: "mgmt-q", ConnRef: "edge"},
		Connections: map[string]spec.Side{
			"edge": {System: spec.SystemSolace, Host: "tcps://b:55443", MsgVPN: "prod", ClientUser: "u", ClientPass: "p", KeyAlias: "sc"},
		},
		TLS: spec.TLSConfig{
			Truststore: &spec.Store{File: "./certs/truststore.jks", Password: "${T}", Type: "JKS"},
			Keystore:   &spec.Store{File: "./certs/keystore.jks", Password: "${K}", Type: "JKS"},
		},
	}
	le = buildLeaderElection(happy, true)
	if le == nil || le.Session == nil || le.Session.Host != "tcps://b:55443" || le.Session.MsgVPN != "prod" {
		t.Fatalf("conn-ref happy path: le = %+v", le)
	}
	got := map[string]string{}
	for _, p := range le.Session.APIProps {
		got[p.Key] = p.Val
	}
	want := map[string]string{
		"SSL_TRUST_STORE":       "/app/external/classpath/truststores/truststore.jks",
		"SSL_KEY_STORE":         "/app/external/classpath/truststores/keystore.jks",
		"SSL_PRIVATE_KEY_ALIAS": "sc",
	}
	for k, w := range want {
		if got[k] != w {
			t.Errorf("conn-ref happy path: APIProps[%s] = %q, want %q", k, got[k], w)
		}
	}

	// inline session (no conn-ref), non-tcps host: no TLS APIProps are added.
	inline := &spec.Defaults{LeaderElection: spec.LeaderElection{
		Present: true, Mode: spec.LeaderActiveStby, Queue: "mgmt-q",
		Session: &spec.Side{System: spec.SystemSolace, Host: "tcp://b:55555", MsgVPN: "v", ClientUser: "u", ClientPass: "p"},
	}}
	le = buildLeaderElection(inline, true)
	if le == nil || le.Session == nil || le.Session.Host != "tcp://b:55555" || le.Session.ClientUser != "u" {
		t.Fatalf("inline session: le = %+v", le)
	}
	if len(le.Session.APIProps) != 0 {
		t.Errorf("inline session: non-tcps host should have no TLS APIProps, got %+v", le.Session.APIProps)
	}

	// mount rewrite in both states: config (mount=false) keeps the env.yaml path
	// verbatim; deploy (mount=true) rewrites to the container mount path.
	mountable := &spec.Defaults{
		LeaderElection: spec.LeaderElection{Present: true, Mode: spec.LeaderActiveStby, Queue: "mgmt-q", ConnRef: "edge"},
		Connections: map[string]spec.Side{
			"edge": {System: spec.SystemSolace, Host: "tcps://b:55443", MsgVPN: "prod", ClientUser: "u", ClientPass: "p"},
		},
		TLS: spec.TLSConfig{Truststore: &spec.Store{File: "./certs/truststore.jks", Password: "${T}", Type: "JKS"}},
	}
	raw := buildLeaderElection(mountable, false)
	mnt := buildLeaderElection(mountable, true)
	if rawTS := trustStoreVal(raw); rawTS != "./certs/truststore.jks" {
		t.Errorf("config (mount=false) truststore = %q, want ./certs/truststore.jks", rawTS)
	}
	if mntTS := trustStoreVal(mnt); mntTS != "/app/external/classpath/truststores/truststore.jks" {
		t.Errorf("deploy (mount=true) truststore = %q", mntTS)
	}
}
