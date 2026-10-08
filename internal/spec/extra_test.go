package spec

import (
	"reflect"
	"slices"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// extraKeyNames lists a captured Extra mapping's keys in order.
func extraKeyNames(n *yaml.Node) []string {
	if n == nil {
		return nil
	}
	var out []string
	for i := 0; i+1 < len(n.Content); i += 2 {
		out = append(out, n.Content[i].Value)
	}
	return out
}

// extraValue returns the value node under key in a captured Extra mapping.
func extraValue(t *testing.T, n *yaml.Node, key string) *yaml.Node {
	t.Helper()
	for i := 0; i+1 < len(n.Content); i += 2 {
		if n.Content[i].Value == key {
			return n.Content[i+1]
		}
	}
	t.Fatalf("no %q in %v", key, extraKeyNames(n))
	return nil
}

// TestKnownKeysReadsYAMLTags pins the schema derivation on yaml.v3's own
// naming: the tag's first element, options dropped, an untagged field by its
// lower-cased name, and a "-" tag or an unexported field not at all.
func TestKnownKeysReadsYAMLTags(t *testing.T) {
	type tagged struct {
		Plain    string `yaml:"plain-key"`
		Options  string `yaml:"with-options,omitempty"`
		Untagged string
		Skipped  string `yaml:"-"`
		hidden   string
	}
	got := knownKeys(reflect.TypeOf(tagged{}))
	want := map[string]bool{"plain-key": true, "with-options": true, "untagged": true}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("knownKeys = %v, want %v", got, want)
	}
	_ = tagged{}.hidden
}

// TestToolManagedKeysAreSchemaKeys keeps the three key sets consistent: every
// credential key is tool-managed, and every tool-managed key is one the schema
// reads -- except mq's ssl-bundle, which the tool derives rather than reads.
func TestToolManagedKeysAreSchemaKeys(t *testing.T) {
	for _, system := range []string{SystemSolace, SystemMQ} {
		known := KnownKeys(system)
		if !slices.IsSorted(known) || len(known) == 0 {
			t.Errorf("%s: KnownKeys = %v, want a sorted, non-empty list", system, known)
		}
		managed := ToolManagedKeys(system)
		for _, k := range managed {
			if system == SystemMQ && k == "ssl-bundle" {
				if slices.Contains(known, k) {
					t.Errorf("mq: ssl-bundle is derived, not read, so it must not be a schema key")
				}
				continue
			}
			if !slices.Contains(known, k) {
				t.Errorf("%s: tool-managed key %q is not a schema key", system, k)
			}
		}
		for _, k := range CredentialKeys(system) {
			if !slices.Contains(managed, k) {
				t.Errorf("%s: credential key %q is not tool-managed", system, k)
			}
		}
	}
	if KnownKeys("other") != nil || ToolManagedKeys("other") != nil || CredentialKeys("other") != nil {
		t.Error("an unknown system has no keys")
	}
}

// TestCanonicalKey pins the fold Spring's relaxed binding applies when it
// matches a key: case and the - and _ separators are ignored, nothing else.
func TestCanonicalKey(t *testing.T) {
	for _, c := range []struct{ in, want string }{
		{"user-authentication-mqcsp", "userauthenticationmqcsp"},
		{"userAuthenticationMQCSP", "userauthenticationmqcsp"},
		{"USER_AUTHENTICATION_MQCSP", "userauthenticationmqcsp"},
		{"Client-Name_x", "clientnamex"},
		{"pool.max-connections[0]", "pool.maxconnections[0]"},
	} {
		if got := CanonicalKey(c.in); got != c.want {
			t.Errorf("CanonicalKey(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

// TestParseWorkflowCapturesExtraKeys pins what a side's other keys come
// through as: every key the schema does not name, in file order, each value
// as written -- a bool stays a bool, a camelCase name keeps its case, a
// double-quoted string keeps its quotes, a nested mapping or list comes whole.
// A known key and a transform key (reported as misplaced) are left out, and
// a block of known keys alone captures nothing.
func TestParseWorkflowCapturesExtraKeys(t *testing.T) {
	wf, err := ParseWorkflow([]byte(`source:
  mq:
    conn-name: h(1414)
    queue-manager: QM1
    channel: CH
    queue: IN
    user-authentication-mqcsp: false
    applicationName: "orders bridge"
    pool:
      enabled: true
      max-connections: 5
    ccdt-url: ${CCDT}
    reconnect-options: [QMGR, DISABLED]
    transform-headers: {expressions: {a: b}}
target:
  solace:
    host: tcp://b
    msg-vpn: v
    queue: OUT
    client-name: ${HOSTNAME}
`), "wf.yaml")
	if err != nil {
		t.Fatalf("ParseWorkflow: %v", err)
	}
	mq := wf.Source.Extra
	want := []string{"user-authentication-mqcsp", "applicationName", "pool", "ccdt-url", "reconnect-options"}
	if got := extraKeyNames(mq); !reflect.DeepEqual(got, want) {
		t.Fatalf("mq extra keys = %v, want %v", got, want)
	}
	if v := extraValue(t, mq, "user-authentication-mqcsp"); v.Kind != yaml.ScalarNode || v.Value != "false" || v.ShortTag() != "!!bool" {
		t.Errorf("user-authentication-mqcsp = %+v, want the bool false as written", v)
	}
	if v := extraValue(t, mq, "applicationName"); v.Style != yaml.DoubleQuotedStyle || v.Value != "orders bridge" {
		t.Errorf("applicationName = %q (style %v), want it double-quoted as written", v.Value, v.Style)
	}
	if v := extraValue(t, mq, "pool"); v.Kind != yaml.MappingNode || len(v.Content) != 4 || v.Content[2].Value != "max-connections" {
		t.Errorf("pool = %+v, want the nested mapping whole", v)
	}
	if v := extraValue(t, mq, "reconnect-options"); v.Kind != yaml.SequenceNode || len(v.Content) != 2 {
		t.Errorf("reconnect-options = %+v, want the list whole", v)
	}
	if !slices.Contains(wf.MisplacedTransforms, "source.mq.transform-headers") {
		t.Errorf("the transform key is the transform scan's, got misplaced %v", wf.MisplacedTransforms)
	}
	if got := extraKeyNames(wf.Target.Extra); !reflect.DeepEqual(got, []string{"client-name"}) {
		t.Errorf("solace extra keys = %v, want [client-name]", got)
	}

	plain, err := ParseWorkflow([]byte("source:\n  mq: {conn-name: h(1414), queue-manager: QM, channel: C, queue: IN, client-id: app1}\ntarget:\n  solace:\n    host: tcp://b\n    msg-vpn: v\n    queue: OUT\n"), "wf.yaml")
	if err != nil {
		t.Fatalf("ParseWorkflow: %v", err)
	}
	if got := extraKeyNames(plain.Source.Extra); !reflect.DeepEqual(got, []string{"client-id"}) {
		t.Errorf("flow-style mq extra keys = %v, want [client-id]", got)
	}
	if plain.Target.Extra != nil {
		t.Errorf("a block of known keys alone captures nothing, got %v", extraKeyNames(plain.Target.Extra))
	}
}

// TestParseWorkflowExtraKeysFollowAliasesAndMergeKeys pins the YAML reuse
// forms env.yaml authors reach for: a << merge is expanded in place with the
// block's own key winning over a merged one, a block written as an alias
// reads as the block it names, and an aliased value is the value it names.
func TestParseWorkflowExtraKeysFollowAliasesAndMergeKeys(t *testing.T) {
	wf, err := ParseWorkflow([]byte(`x-base: &base
  application-name: base-app
  ccdt-url: file:///ccdt
x-mq: &mq
  <<: *base
  conn-name: h(1414)
  queue-manager: QM1
  channel: CH
  client-id: shared
  application-name: shared-app
x-id: &id app-7
x-sol: &sol
  host: tcp://b
  msg-vpn: v
  queue: OUT
  client-name: *id
source:
  mq:
    <<: *mq
    queue: IN
    client-id: mine
target:
  solace: *sol
`), "wf.yaml")
	if err != nil {
		t.Fatalf("ParseWorkflow: %v", err)
	}
	mq := wf.Source.Extra
	if got := extraKeyNames(mq); !reflect.DeepEqual(got, []string{"ccdt-url", "application-name", "client-id"}) {
		t.Fatalf("merged mq extra keys = %v, want the merged keys in place then the block's own client-id", got)
	}
	if v := extraValue(t, mq, "client-id"); v.Value != "mine" {
		t.Errorf("client-id = %q, want the block's own value over the merged one", v.Value)
	}
	if v := extraValue(t, mq, "application-name"); v.Value != "shared-app" {
		t.Errorf("application-name = %q, want the merge source's own value over what it merges in", v.Value)
	}
	if wf.Source.QueueManager != "QM1" {
		t.Errorf("the known keys still merge: queue-manager = %q", wf.Source.QueueManager)
	}
	sol := wf.Target.Extra
	if got := extraKeyNames(sol); !reflect.DeepEqual(got, []string{"client-name"}) {
		t.Fatalf("aliased solace block extra keys = %v, want [client-name]", got)
	}
	if v := extraValue(t, sol, "client-name"); v.Kind != yaml.ScalarNode || v.Value != "app-7" {
		t.Errorf("client-name = %+v, want the aliased value app-7", v)
	}
}

// TestParseDefaultsCapturesExtraKeys covers the env.yaml sites read through
// the same structs: each connection, the inline leader-election session, and
// mq-defaults as a block of its own; a connection naming both systems is
// ambiguous and carries nothing.
func TestParseDefaultsCapturesExtraKeys(t *testing.T) {
	d, err := ParseDefaults([]byte(`connections:
  qm:
    mq:
      conn-name: h(1414)
      queue-manager: QM1
      channel: CH
      user-authentication-mqcsp: false
  edge:
    solace:
      host: tcp://b
      msg-vpn: v
      client-name: edge-1
  both:
    solace:
      host: tcp://b
    mq:
      conn-name: h(1414)
leader-election:
  mode: active_standby
  queue: mgmt-q
  session:
    host: tcp://b
    msg-vpn: v
    connect-retries-per-host: 3
mq-defaults:
  application-name: fleet
`))
	if err != nil {
		t.Fatal(err)
	}
	if got := extraKeyNames(d.Connections["qm"].Extra); !reflect.DeepEqual(got, []string{"user-authentication-mqcsp"}) {
		t.Errorf("connections.qm extra keys = %v", got)
	}
	if got := extraKeyNames(d.Connections["edge"].Extra); !reflect.DeepEqual(got, []string{"client-name"}) {
		t.Errorf("connections.edge extra keys = %v", got)
	}
	if both := d.Connections["both"]; both.System != "" || both.Extra != nil {
		t.Errorf("a connection naming both systems = %+v, want the ambiguous empty side", both)
	}
	if got := extraKeyNames(d.LeaderElection.Session.Extra); !reflect.DeepEqual(got, []string{"connect-retries-per-host"}) {
		t.Errorf("session extra keys = %v", got)
	}
	if got := extraKeyNames(d.MQDefaults); !reflect.DeepEqual(got, []string{"application-name"}) {
		t.Errorf("mq-defaults = %v, want [application-name]", got)
	}
	if empty, err := ParseDefaults(nil); err != nil || empty.MQDefaults != nil {
		t.Errorf("an empty env.yaml has no mq-defaults, got %v (%v)", empty.MQDefaults, err)
	}
}

// TestSideSetsConnFieldsCountsExtraKeys pins that another key beside conn-ref
// is a connection field: a conn-ref side carrying one trips the strict rule.
func TestSideSetsConnFieldsCountsExtraKeys(t *testing.T) {
	extra := &yaml.Node{Kind: yaml.MappingNode, Content: []*yaml.Node{{Kind: yaml.ScalarNode, Value: "client-id"}, {Kind: yaml.ScalarNode, Value: "x"}}}
	if !(Side{System: SystemMQ, ConnRef: "qm", Extra: extra}).SetsConnFields() {
		t.Error("an extra key beside conn-ref must count as a connection field")
	}
	if (Side{System: SystemMQ, ConnRef: "qm", DestKind: DestQueue, Dest: "Q"}).SetsConnFields() {
		t.Error("a bare conn-ref side sets no connection field")
	}
}

// TestResolveCarriesTheConnectionsExtraKeys pins that a conn-ref side resolves
// to the connection's other keys along with its tuple.
func TestResolveCarriesTheConnectionsExtraKeys(t *testing.T) {
	extra := &yaml.Node{Kind: yaml.MappingNode, Content: []*yaml.Node{{Kind: yaml.ScalarNode, Value: "client-id"}, {Kind: yaml.ScalarNode, Value: "x"}}}
	d := &Defaults{Connections: map[string]Side{"qm": {System: SystemMQ, ConnName: "h(1414)", QueueManager: "QM", Channel: "C", Extra: extra}}}
	r := d.Resolve(Side{System: SystemMQ, ConnRef: "qm", DestKind: DestQueue, Dest: "Q"})
	if r.Extra != extra || r.Dest != "Q" {
		t.Errorf("resolved side = %+v, want the connection's extra keys with the side's destination", r)
	}
}

// TestMisplacedEnvTransformsCoversLeaderSession pins the scan over the inline
// management session, which is read through the same struct as a side and so
// would drop a transform just as silently.
func TestMisplacedEnvTransformsCoversLeaderSession(t *testing.T) {
	doc := []byte("leader-election:\n  mode: active_standby\n  session:\n    host: tcp://b\n    msg-vpn: v\n    transform: x\n")
	want := []string{"leader-election.session.transform"}
	e, err := ParseEnv(doc)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(e.MisplacedTransforms, want) {
		t.Errorf("ParseEnv misplaced = %v, want %v", e.MisplacedTransforms, want)
	}
	if e.Defaults.LeaderElection.Session.Extra != nil {
		t.Errorf("a transform key is the scan's, not an extra key: %v", extraKeyNames(e.Defaults.LeaderElection.Session.Extra))
	}
}

// TestParseWorkflowTypeErrorInsideASideStillNamesTheLine pins that the typed
// decode's own error -- a wrong scalar type on a known key -- still comes
// through the capture with its line number, so the capture never hides it.
func TestParseWorkflowTypeErrorInsideASideStillNamesTheLine(t *testing.T) {
	_, err := ParseWorkflow([]byte("source:\n  mq:\n    conn-name: h(1414)\n    tls: notabool\n"), "wf.yaml")
	if err == nil || !strings.Contains(err.Error(), "line 4") || !strings.Contains(err.Error(), "cannot unmarshal") {
		t.Errorf("err = %v, want the yaml type error naming line 4", err)
	}
}

// TestExpandLeavesExtraKeysAlone pins rule 6 for the new field: a ${...}
// inside another key reaches Spring as typed, never expanded here.
func TestExpandLeavesExtraKeysAlone(t *testing.T) {
	var doc yaml.Node
	if err := yaml.Unmarshal([]byte("client-name: ${SHOULD_NOT_EXPAND}\n"), &doc); err != nil {
		t.Fatal(err)
	}
	wfs := []Workflow{{Source: Side{Extra: doc.Content[0]}}}
	Expand(Expander{Lookup: lookupOf(map[string]string{"SHOULD_NOT_EXPAND": "changed"})}, &Env{}, wfs)
	if got := wfs[0].Source.Extra.Content[1].Value; got != "${SHOULD_NOT_EXPAND}" {
		t.Errorf("extra key expanded to %q", got)
	}
}
