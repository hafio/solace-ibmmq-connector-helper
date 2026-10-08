package consolidate

import (
	"reflect"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/spec"
)

// extraWF is a workflow carrying src and tgt, for the extra-key tests.
func extraWF(file string, src, tgt spec.Side) spec.Workflow {
	return spec.Workflow{File: file, Enabled: true, SourceSet: true, TargetSet: true, Source: src, Target: tgt}
}

// withExtra returns s carrying the extra keys y as its block's other keys.
func withExtra(t *testing.T, s spec.Side, y string) spec.Side {
	t.Helper()
	s.Extra = propsNode(t, y)
	return s
}

// binderOfKind returns the one binder of kind in m.
func binderOfKind(t *testing.T, m *Model, kind string) *Binder {
	t.Helper()
	var found *Binder
	for _, b := range m.Binders {
		if b.Kind != kind {
			continue
		}
		if found != nil {
			t.Fatalf("more than one %s binder: %v", kind, m.Binders)
		}
		found = b
	}
	if found == nil {
		t.Fatalf("no %s binder in %v", kind, m.Binders)
	}
	return found
}

// scalarProps reduces props to key -> value for the scalar entries, so a test
// can compare order-free where order is covered elsewhere.
func scalarProps(props []Prop) map[string]string {
	out := map[string]string{}
	for _, p := range props {
		if p.Sub == nil {
			out[p.Key] = p.Val
		}
	}
	return out
}

// TestOverlayExtras pins how a connection's own other keys lay over a defaults
// block: a key the connection also sets is replaced where the default stood,
// by Spring's reading of the name, anything else is appended, and nothing to
// lay on leaves the defaults exactly as they were.
func TestOverlayExtras(t *testing.T) {
	base := []Prop{{Key: "connect-retries", Val: "-1"}, {Key: "reconnect-retries", Val: "-1"}}
	for _, c := range []struct {
		name string
		over []Prop
		want []Prop
	}{
		{"replaces in place", []Prop{{Key: "connect-retries", Val: "5"}}, []Prop{{Key: "connect-retries", Val: "5"}, {Key: "reconnect-retries", Val: "-1"}}},
		{"another spelling replaces too", []Prop{{Key: "connectRetries", Val: "5"}}, []Prop{{Key: "connectRetries", Val: "5"}, {Key: "reconnect-retries", Val: "-1"}}},
		{"a new key is appended", []Prop{{Key: "client-name", Val: "n"}}, []Prop{{Key: "connect-retries", Val: "-1"}, {Key: "reconnect-retries", Val: "-1"}, {Key: "client-name", Val: "n"}}},
	} {
		if got := overlayExtras(base, c.over); !reflect.DeepEqual(got, c.want) {
			t.Errorf("%s: overlayExtras = %v, want %v", c.name, got, c.want)
		}
	}
	if got := overlayExtras(base, nil); !reflect.DeepEqual(got, base) {
		t.Errorf("no extras: got %v, want base untouched", got)
	}
	if got := overlayExtras(nil, nil); got != nil {
		t.Errorf("nil over nil = %v, want nil", got)
	}
	if got := overlayExtras(nil, []Prop{{Key: "a", Val: "1"}}); !reflect.DeepEqual(got, []Prop{{Key: "a", Val: "1"}}) {
		t.Errorf("extras over no defaults = %v", got)
	}
}

// TestMergePropNestedValues pins the structural compare mergeProp applies to a
// nested value: the same mapping arriving again is not a change, a different
// one is, and the same node is never one.
func TestMergePropNestedValues(t *testing.T) {
	var warns []string
	pool := propsNode(t, "max-connections: 5\n")
	list := mergeProp(nil, Prop{Key: "pool", Sub: pool}, &warns, "b")
	list = mergeProp(list, Prop{Key: "pool", Sub: propsNode(t, "max-connections: 5\n")}, &warns, "b")
	if len(warns) != 0 {
		t.Errorf("an identical nested value warned: %v", warns)
	}
	list = mergeProp(list, Prop{Key: "pool", Sub: pool}, &warns, "b")
	if len(warns) != 0 {
		t.Errorf("the same node warned: %v", warns)
	}
	list = mergeProp(list, Prop{Key: "pool", Sub: propsNode(t, "max-connections: 9\n")}, &warns, "b")
	if len(warns) != 1 || !strings.Contains(warns[0], `passthrough key "pool" set more than once`) {
		t.Errorf("a different nested value must warn once, got %v", warns)
	}
	if list[0].Sub.Content[1].Value != "9" {
		t.Errorf("last writer wins: %v", list[0].Sub.Content[1].Value)
	}
}

// TestBuildCarriesExtraKeysOntoBothBinders pins the whole path for an inline
// side: its other keys reach its own binder -- an MQ one as Extras, a Solace
// one laid over solace-defaults with the connection winning in place.
func TestBuildCarriesExtraKeysOntoBothBinders(t *testing.T) {
	d := &spec.Defaults{SolaceDefaults: propsNode(t, "connect-retries: -1\nreconnect-retries: -1\n")}
	src := withExtra(t, mqSide("QM", "IN", spec.DestQueue, false), "user-authentication-mqcsp: false\npool:\n  max-connections: 5\n")
	tgt := withExtra(t, solaceSide("v", "OUT", spec.DestQueue, ""), "connect-retries: 5\nclient-name: n\n")
	m, warns := Build([]spec.Workflow{extraWF("0.yaml", src, tgt)}, d, Opts{MountStores: true})
	if len(warns) != 0 {
		t.Errorf("unexpected warnings: %v", warns)
	}
	mq := binderOfKind(t, m, spec.SystemMQ).MQ
	if len(mq.Extras) != 2 || mq.Extras[0] != (Prop{Key: "user-authentication-mqcsp", Val: "false"}) || mq.Extras[1].Key != "pool" || mq.Extras[1].Sub == nil {
		t.Errorf("mq Extras = %+v, want the scalar then the pool mapping", mq.Extras)
	}
	if len(mq.AddlProps) != 0 {
		t.Errorf("other keys are siblings, not additional-properties: %v", mq.AddlProps)
	}
	sol := binderOfKind(t, m, spec.SystemSolace).Solace
	want := []Prop{{Key: "connect-retries", Val: "5"}, {Key: "reconnect-retries", Val: "-1"}, {Key: "client-name", Val: "n"}}
	if !reflect.DeepEqual(sol.Extras, want) {
		t.Errorf("solace Extras = %v, want %v", sol.Extras, want)
	}
}

// TestBuildMergesExtraKeysAcrossSidesOfOneBinder pins that other keys never
// split a binder: two sides on one tuple merge their keys into the one binder,
// a union when they differ in key, last (by filename) wins with a warning when
// they disagree on a value -- exactly as api-properties do.
func TestBuildMergesExtraKeysAcrossSidesOfOneBinder(t *testing.T) {
	a := extraWF("0.yaml", withExtra(t, mqSide("QM", "IN-0", spec.DestQueue, false), "client-id: c\n"), solaceSide("v", "OUT-0", spec.DestQueue, ""))
	b := extraWF("1.yaml", withExtra(t, mqSide("QM", "IN-1", spec.DestQueue, false), "application-name: app\n"), solaceSide("v", "OUT-1", spec.DestQueue, ""))
	m, warns := Build([]spec.Workflow{a, b}, &spec.Defaults{}, Opts{MountStores: true})
	if len(warns) != 0 {
		t.Errorf("a union of keys must not warn: %v", warns)
	}
	mq := binderOfKind(t, m, spec.SystemMQ).MQ
	if want := []Prop{{Key: "client-id", Val: "c"}, {Key: "application-name", Val: "app"}}; !reflect.DeepEqual(mq.Extras, want) {
		t.Errorf("Extras = %v, want the union %v", mq.Extras, want)
	}

	b.Source = withExtra(t, mqSide("QM", "IN-1", spec.DestQueue, false), "client-id: other\n")
	m, warns = Build([]spec.Workflow{a, b}, &spec.Defaults{}, Opts{MountStores: true})
	if len(warns) != 1 || !strings.Contains(warns[0], `passthrough key "client-id" set more than once; last (by filename) wins`) {
		t.Errorf("a disagreement must warn once, got %v", warns)
	}
	if got := binderOfKind(t, m, spec.SystemMQ).MQ.Extras; !reflect.DeepEqual(got, []Prop{{Key: "client-id", Val: "other"}}) {
		t.Errorf("Extras = %v, want the later file's value", got)
	}
}

// TestBuildNestedExtraKeyThroughSharedConnRefIsQuiet pins the case mergeProp's
// structural compare exists for: a connection carrying a nested block, reused
// by several workflows through conn-ref, merges the same block once per side
// and must not warn that it was set more than once.
func TestBuildNestedExtraKeyThroughSharedConnRefIsQuiet(t *testing.T) {
	d := &spec.Defaults{Connections: map[string]spec.Side{
		"qm": withExtra(t, spec.Side{System: spec.SystemMQ, ConnName: "h(1414)", QueueManager: "QM", Channel: "C", UserEnv: "MQ_USER", PasswordEnv: "MQ_PASSWORD"}, "pool:\n  enabled: true\n  max-connections: 5\n"),
	}}
	var wfs []spec.Workflow
	for _, n := range []string{"0.yaml", "1.yaml", "2.yaml"} {
		src := spec.Side{System: spec.SystemMQ, ConnRef: "qm", DestKind: spec.DestQueue, Dest: "IN-" + n}
		wfs = append(wfs, extraWF(n, src, solaceSide("v", "OUT-"+n, spec.DestQueue, "")))
	}
	m, warns := Build(wfs, d, Opts{MountStores: true})
	if len(warns) != 0 {
		t.Errorf("a shared nested extra key warned: %v", warns)
	}
	mq := binderOfKind(t, m, spec.SystemMQ).MQ
	if len(mq.Extras) != 1 || mq.Extras[0].Key != "pool" {
		t.Errorf("Extras = %+v, want the one pool mapping", mq.Extras)
	}
}

// TestBuildDropsExtraKeysTheToolManages pins the consolidate backstop under
// validate's refusal: another spelling of a key the tool writes itself is
// dropped with the passthrough warning, and the tool's value stays.
func TestBuildDropsExtraKeysTheToolManages(t *testing.T) {
	d := &spec.Defaults{TLS: spec.TLSConfig{Truststore: &spec.Store{File: "./certs/truststore.jks", PasswordEnv: "TS", Type: "JKS"}}}
	src := withExtra(t, mqSide("QM", "IN", spec.DestQueue, true), "ssl-bundle: theirs\nqueueManager: OTHER\n")
	tgt := withExtra(t, solaceSide("v", "OUT", spec.DestQueue, ""), "msgVpn: other\n")
	m, warns := Build([]spec.Workflow{extraWF("0.yaml", src, tgt)}, d, Opts{MountStores: true})
	mq := binderOfKind(t, m, spec.SystemMQ).MQ
	if mq.SSLBundle != "mq-conn-1-bundle" || mq.QueueManager != "QM" || len(mq.Extras) != 0 {
		t.Errorf("mq binder = %+v, want the tool's bundle and queue manager and no Extras", mq)
	}
	sol := binderOfKind(t, m, spec.SystemSolace).Solace
	if sol.MsgVPN != "v" || len(sol.Extras) != 0 {
		t.Errorf("solace binder = %+v, want the tool's msg-vpn and no Extras", sol)
	}
	for _, want := range []string{
		`binder "mq-conn-1": passthrough overrides tool-managed key "ssl-bundle"; tool value kept`,
		`binder "mq-conn-1": passthrough overrides tool-managed key "queueManager"; tool value kept`,
		`binder "sol-conn-1": passthrough overrides tool-managed key "msgVpn"; tool value kept`,
	} {
		found := false
		for _, w := range warns {
			found = found || w == want
		}
		if !found {
			t.Errorf("missing warning %q in %v", want, warns)
		}
	}
}

// TestBuildLeaderElectionSessionCarriesExtraKeys pins the session's share of
// this: its connection's other keys lay over solace-defaults exactly as a
// binder's do, and a managed-key spelling is dropped under the session's own
// label.
func TestBuildLeaderElectionSessionCarriesExtraKeys(t *testing.T) {
	d := &spec.Defaults{
		LeaderElection: spec.LeaderElection{Present: true, Mode: spec.LeaderActiveStby, Queue: "mgmt-q", ConnRef: "edge"},
		Connections: map[string]spec.Side{
			"edge": withExtra(t, spec.Side{System: spec.SystemSolace, Host: "tcp://b:55555", MsgVPN: "prod", ClientUserEnv: "EU", ClientPassEnv: "EP"}, "connect-retries: 5\nclient-name: ${HOSTNAME}\nmsgVpn: other\n"),
		},
		SolaceDefaults: propsNode(t, "connect-retries: -1\nreconnect-retries: -1\n"),
	}
	secretRef, _ := testSecretRef()
	var warns []string
	le := buildLeaderElection(d, true, secretRef, fixedLeaderNames, &warns)
	if le == nil || le.Session == nil {
		t.Fatalf("le = %+v", le)
	}
	want := []Prop{{Key: "connect-retries", Val: "5"}, {Key: "reconnect-retries", Val: "-1"}, {Key: "client-name", Val: "${HOSTNAME}"}}
	if !reflect.DeepEqual(le.Session.Extras, want) {
		t.Errorf("session Extras = %v, want %v", le.Session.Extras, want)
	}
	if len(warns) != 1 || warns[0] != `leader-election session: passthrough overrides tool-managed key "msgVpn"; tool value kept` {
		t.Errorf("warns = %v", warns)
	}
}

// TestBuildMQDefaultsOverriddenByConnection pins mq-defaults: every MQ binder
// gets the block, a connection's own key replaces a default in place, and a
// managed key in the block is dropped once, not once per binder.
func TestBuildMQDefaultsOverriddenByConnection(t *testing.T) {
	d := &spec.Defaults{MQDefaults: propsNode(t, "channel: nope\nuser-authentication-mqcsp: false\napplication-name: fleet\n")}
	a := extraWF("0.yaml", withExtra(t, mqSide("QM1", "IN-0", spec.DestQueue, false), "applicationName: mine\n"), solaceSide("v", "OUT-0", spec.DestQueue, ""))
	b := extraWF("1.yaml", mqSide("QM2", "IN-1", spec.DestQueue, false), solaceSide("v", "OUT-1", spec.DestQueue, ""))
	m, warns := Build([]spec.Workflow{a, b}, d, Opts{MountStores: true})
	if len(warns) != 1 || warns[0] != `mq-defaults: passthrough overrides tool-managed key "channel"; tool value kept` {
		t.Errorf("warns = %v, want the one mq-defaults warning", warns)
	}
	var qm1, qm2 *MQBinder
	for _, bd := range m.Binders {
		if bd.Kind != spec.SystemMQ {
			continue
		}
		switch bd.MQ.QueueManager {
		case "QM1":
			qm1 = bd.MQ
		case "QM2":
			qm2 = bd.MQ
		}
	}
	if qm1 == nil || qm2 == nil {
		t.Fatalf("binders = %v", m.Binders)
	}
	if want := []Prop{{Key: "user-authentication-mqcsp", Val: "false"}, {Key: "applicationName", Val: "mine"}}; !reflect.DeepEqual(qm1.Extras, want) {
		t.Errorf("QM1 Extras = %v, want %v (connection wins in place)", qm1.Extras, want)
	}
	if want := []Prop{{Key: "user-authentication-mqcsp", Val: "false"}, {Key: "application-name", Val: "fleet"}}; !reflect.DeepEqual(qm2.Extras, want) {
		t.Errorf("QM2 Extras = %v, want the defaults %v", qm2.Extras, want)
	}
	if qm1.Channel != "CH" || qm2.Channel != "CH" {
		t.Errorf("the tool's channel must stay: %q %q", qm1.Channel, qm2.Channel)
	}
	if got := scalarProps(qm2.Extras); got["channel"] != "" {
		t.Errorf("the managed key leaked into Extras: %v", got)
	}
}

// TestSameNode pins the structural compare behind mergeProp's nested case:
// nil only equals nil, a scalar by its formatted value (so quoting counts), a
// container element by element.
func TestSameNode(t *testing.T) {
	plain := &yaml.Node{Kind: yaml.ScalarNode, Value: "a"}
	quoted := &yaml.Node{Kind: yaml.ScalarNode, Value: "a", Style: yaml.DoubleQuotedStyle}
	for _, c := range []struct {
		name string
		a, b *yaml.Node
		want bool
	}{
		{"nil and nil", nil, nil, true},
		{"nil and a node", nil, plain, false},
		{"same value", plain, &yaml.Node{Kind: yaml.ScalarNode, Value: "a"}, true},
		{"quoting differs", plain, quoted, false},
		{"same mapping", propsNode(t, "a: 1\nb: 2\n"), propsNode(t, "a: 1\nb: 2\n"), true},
		{"different length", propsNode(t, "a: 1\n"), propsNode(t, "a: 1\nb: 2\n"), false},
		{"different kind", propsNode(t, "a: 1\n"), &yaml.Node{Kind: yaml.SequenceNode, Content: []*yaml.Node{plain, plain}}, false},
	} {
		if got := sameNode(c.a, c.b); got != c.want {
			t.Errorf("%s: sameNode = %v, want %v", c.name, got, c.want)
		}
	}
}
