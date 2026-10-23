package consolidate

import (
	"reflect"
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

// testSecretRef mirrors the bookkeeping Build's own secretRef closure does
// (see consolidate.go), for white-box tests that call buildLeaderElection
// directly: one entry per mount name, "" for an unset credential, the same
// name collapsing repeats onto the first recording, and an -env credential
// mounted under its own variable name. origin only labels collisions, which
// these white-box callers do not produce, so it is ignored here.
func testSecretRef() (secretFn, func() []SecretRef) {
	seen := map[string]bool{}
	var secrets []SecretRef
	fn := func(stable, _ string, c spec.Cred) string {
		if c.Empty() {
			return ""
		}
		name := stable
		if c.EnvVar != "" {
			name = c.EnvVar
		}
		if !seen[name] {
			seen[name] = true
			secrets = append(secrets, SecretRef{Stable: name, Literal: c.Literal, EnvVar: c.EnvVar})
		}
		return "${" + name + "}"
	}
	return fn, func() []SecretRef { return secrets }
}

// fixedLeaderNames is the leaderNameFn a management session gets when no
// binder shares its connection: the fixed LEADER_ELECTION_* pair. White-box
// calls to buildLeaderElection have no binders around them, so this is what
// Build would hand them.
func fixedLeaderNames(spec.Side) (string, string) { return LeaderUsernameName, LeaderPasswordName }

// propsNode parses a small YAML mapping into the *yaml.Node the spec carries
// for verbatim passthrough (api-properties) and for solace-defaults.
func propsNode(t *testing.T, y string) *yaml.Node {
	t.Helper()
	var n yaml.Node
	if err := yaml.Unmarshal([]byte(y), &n); err != nil {
		t.Fatal(err)
	}
	return n.Content[0]
}

// TestBuildCarriesEachWorkflowsTransformBlocks pins that a workflow's
// transform and transform-headers blocks reach its own WorkflowEnable entry,
// untouched, and never a neighbour's: the render places each under that
// workflow's id. A workflow with neither carries neither.
func TestBuildCarriesEachWorkflowsTransformBlocks(t *testing.T) {
	tr := propsNode(t, "expressions:\n  - transform: \"target['headers']['h'] = 'x'\"\n")
	th := propsNode(t, "expressions:\n  h: \"'x'\"\n")
	wfs := []spec.Workflow{
		{File: "a.yaml", Enabled: true, SourceSet: true, TargetSet: true, Source: solaceSide("v", "IN-A", spec.DestQueue, ""), Target: mqSide("QM", "OUT-A", spec.DestQueue, false), Transform: tr},
		{File: "b.yaml", Enabled: false, SourceSet: true, TargetSet: true, Source: solaceSide("v", "IN-B", spec.DestQueue, ""), Target: mqSide("QM", "OUT-B", spec.DestQueue, false), TransformHeaders: th},
		{File: "c.yaml", Enabled: true, SourceSet: true, TargetSet: true, Source: solaceSide("v", "IN-C", spec.DestQueue, ""), Target: mqSide("QM", "OUT-C", spec.DestQueue, false)},
	}
	m, _ := Build(wfs, &spec.Defaults{}, Opts{MountStores: true})
	if len(m.Workflows) != 3 {
		t.Fatalf("workflows = %+v, want 3", m.Workflows)
	}
	for i, c := range []struct {
		enabled bool
		tr, th  *yaml.Node
	}{
		{true, tr, nil},
		{false, nil, th},
		{true, nil, nil},
	} {
		got := m.Workflows[i]
		if got.ID != i || got.Enabled != c.enabled || got.Transform != c.tr || got.TransformHeaders != c.th {
			t.Errorf("workflow %d = %+v, want id %d, enabled %v, transform %p, transform-headers %p", i, got, i, c.enabled, c.tr, c.th)
		}
	}
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
		// A one-line block scalar lands on the key's line, so it is quoted only
		// when plain text would not read back the same -- a folded SpEL call
		// would lose everything from " #" on.
		{"one-line folded block, safe plain", &yaml.Node{Kind: yaml.ScalarNode, Value: "plain words", Style: yaml.FoldedStyle}, "plain words"},
		{"one-line folded block holding a comment marker", &yaml.Node{Kind: yaml.ScalarNode, Value: "x = #f('a')", Style: yaml.FoldedStyle}, `"x = #f('a')"`},
		{"multi-line literal block left to the render layer", &yaml.Node{Kind: yaml.ScalarNode, Value: "a\nb\n", Style: yaml.LiteralStyle}, "a\nb\n"},
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
	out := appendPassthrough([]Prop{{Key: "T", Val: "x"}}, []Prop{{Key: "P", Val: "y"}, {Key: "T", Val: "z"}}, &warns, binderOwner("bndr"))
	if len(out) != 2 || len(warns) != 1 {
		t.Fatalf("out=%v warns=%v", out, warns)
	}
	// The owner label is caller-supplied now that the leader-election session
	// shares this function; the binder wording it replaced is load-bearing for
	// anyone grepping their build output, so pin it byte for byte.
	want := `binder "bndr": passthrough overrides tool-managed key "T"; tool value kept`
	if warns[0] != want {
		t.Errorf("warning = %q, want %q", warns[0], want)
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
	mq := spec.Side{System: spec.SystemMQ, ConnName: "h(1414)", QueueManager: "QM1", Channel: "CH", UserEnv: "MQ_USER", PasswordEnv: "MQ_PASSWORD", TLS: true, Cipher: "CIPH",
		Truststore: "mq-ca", Keystore: "mq-client", KeyAlias: "mc", DestKind: spec.DestQueue, Dest: "IN"}
	sol := spec.Side{System: spec.SystemSolace, Host: "tcps://b", MsgVPN: "v", ClientUserEnv: "SOL_USER", ClientPassEnv: "SOL_PASSWORD", Truststore: "mq-ca", DestKind: spec.DestQueue, Dest: "OUT"}
	d := &spec.Defaults{TLS: spec.TLSConfig{Stores: map[string]*spec.Store{
		"mq-ca":     {File: "t.jks", PasswordEnv: "TRUSTSTORE_PASSWORD_ENV", Type: "JKS"},
		"mq-client": {File: "k.jks", PasswordEnv: "KEYSTORE_PASSWORD_ENV", Type: "PKCS12"},
	}}}
	m, _ := Build([]spec.Workflow{{File: "10.yaml", Enabled: true, SourceSet: true, TargetSet: true, Source: mq, Target: sol}}, d, Opts{MountStores: true})
	if !m.MQTLS || len(m.Bundles) != 1 {
		t.Fatalf("MQTLS=%v bundles=%d", m.MQTLS, len(m.Bundles))
	}
	b := m.Bundles[0]
	if !b.HasTruststore || !b.HasKeystore || b.KeyAlias != "mc" || b.KeystoreTyp != "PKCS12" || b.TruststoreTyp != "JKS" {
		t.Fatalf("bundle=%+v", b)
	}
	if want := []StoreUse{{Name: "mq-ca", File: "t.jks", Type: "JKS"}, {Name: "mq-client", File: "k.jks", Type: "PKCS12", Private: true}}; !reflect.DeepEqual(m.Stores, want) {
		t.Errorf("Stores = %+v, want %+v", m.Stores, want)
	}
	// The truststore password is referenced twice -- once by the MQ SSL bundle,
	// once by the Solace binder's tcps:// api-properties (both TLS-enabled sides
	// select the one mq-ca store) -- and must collapse to a single
	// Model.Secrets entry, not be mounted/resolved twice under the same name.
	// Keyed on the operator's own -env name, because that is what an -env
	// credential is mounted under: SecretRef.Stable is the variable written in
	// the spec, and only a literal takes a derived name.
	var truststoreRefs int
	for _, s := range m.Secrets {
		if s.EnvVar == "TRUSTSTORE_PASSWORD_ENV" {
			truststoreRefs++
			if s.Stable != s.EnvVar {
				t.Errorf("truststore secret Stable = %q, want the -env name %q", s.Stable, s.EnvVar)
			}
		}
	}
	if truststoreRefs != 1 {
		t.Fatalf("truststore password referenced by bundle + api-properties should collapse to 1 Secrets entry, got %d in %+v", truststoreRefs, m.Secrets)
	}
	// mq user/password, mq bundle truststore+keystore, sol client-user/pass: 6 distinct secrets.
	if len(m.Secrets) != 6 {
		t.Fatalf("Secrets = %+v, want 6 distinct entries", m.Secrets)
	}
}

func TestBuildCipherConflictWarning(t *testing.T) {
	mk := func(cipher, dest string) spec.Side {
		return spec.Side{System: spec.SystemMQ, ConnName: "h(1)", QueueManager: "QM1", Channel: "CH", UserEnv: "MQ_USER", PasswordEnv: "MQ_PASSWORD", TLS: true, Cipher: cipher, DestKind: spec.DestQueue, Dest: dest}
	}
	sol := spec.Side{System: spec.SystemSolace, Host: "tcp://b", MsgVPN: "v", ClientUserEnv: "SOL_USER", ClientPassEnv: "SOL_PASSWORD", DestKind: spec.DestQueue, Dest: "X"}
	wfs := []spec.Workflow{
		{File: "10.yaml", Enabled: true, SourceSet: true, TargetSet: true, Source: mk("C1", "A"), Target: sol},
		{File: "20.yaml", Enabled: true, SourceSet: true, TargetSet: true, Source: mk("C2", "B"), Target: sol},
	}
	_, warns := Build(wfs, nil, Opts{MountStores: true})
	if !containsSub(warns, "conflicting cipher") {
		t.Fatalf("want cipher conflict, got %v", warns)
	}
}

func TestBuildMessageLoopWarning(t *testing.T) {
	s := spec.Side{System: spec.SystemMQ, ConnName: "h(1)", QueueManager: "QM1", Channel: "CH", UserEnv: "MQ_USER", PasswordEnv: "MQ_PASSWORD", DestKind: spec.DestQueue, Dest: "SAME"}
	_, warns := Build([]spec.Workflow{{File: "10.yaml", Enabled: true, SourceSet: true, TargetSet: true, Source: s, Target: s}}, nil, Opts{MountStores: true})
	if !containsSub(warns, "message loop") {
		t.Fatalf("want loop warn, got %v", warns)
	}
}

func TestBuildSolaceTopicSourceEmitsConsumerTopic(t *testing.T) {
	src := spec.Side{System: spec.SystemSolace, Host: "tcps://b", MsgVPN: "v", ClientUserEnv: "SOL_USER", ClientPassEnv: "SOL_PASSWORD", DestKind: spec.DestTopic, Dest: "evt/in"}
	tgt := spec.Side{System: spec.SystemMQ, ConnName: "h(1)", QueueManager: "QM", Channel: "C", UserEnv: "MQ_USER", PasswordEnv: "MQ_PASSWORD", DestKind: spec.DestQueue, Dest: "OUT"}
	m, _ := Build([]spec.Workflow{{File: "10.yaml", Enabled: true, SourceSet: true, TargetSet: true, Source: src, Target: tgt}}, nil, Opts{MountStores: true})
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

// TestApplyStatusAccessNoOperatorUsers pins the no-configured-users case:
// Build synthesizes the security block from nothing and the reserved account
// is the only user, carrying the literal status password rather than a
// secretRef placeholder. Management security is unconditional now (no
// enabled/disabled toggle survives to this layer), so there is nothing left
// to branch on.
func TestApplyStatusAccessNoOperatorUsers(t *testing.T) {
	wfs := []spec.Workflow{{File: "10.yaml", Enabled: true, SourceSet: true, TargetSet: true,
		Source: solaceSide("prod", "IN", spec.DestQueue, ""), Target: solaceSide("prod", "OUT", spec.DestQueue, "")}}
	m, _ := Build(wfs, nil, Opts{MountStores: true, StatusPassword: "status-literal"})

	if len(m.Security.Users) != 1 || m.Security.Users[0].Name != spec.StatusUserName || m.Security.Users[0].Password != "status-literal" {
		t.Fatalf("Security.Users = %+v, want one reserved user carrying the literal password", m.Security.Users)
	}
	for _, s := range m.Secrets {
		if s.Literal == "status-literal" || s.Stable == securityUserPasswordName(spec.StatusUserName) {
			t.Errorf("status password must never enter Model.Secrets, got %+v", s)
		}
	}
}

// TestApplyStatusAccessCarriesOperatorRoles pins the roles pass-through: an
// operator's roles reach the model exactly as authored (only the password is
// rewritten), while the reserved account is appended with none -- an empty list
// is the connector's read-only default, which is what keeps that account
// read-only. It also proves applyStatusAccess does not mutate the caller's
// Defaults, which shares the roles backing array with the copy it makes.
func TestApplyStatusAccessCarriesOperatorRoles(t *testing.T) {
	d := &spec.Defaults{Security: spec.Security{Users: []spec.User{
		{Name: "ops", PasswordEnv: "OPS_PASS_ENV", Roles: []string{"admin"}},
		{Name: "probe", Password: "probe-pass"},
	}}}
	wfs := []spec.Workflow{{File: "10.yaml", Enabled: true, SourceSet: true, TargetSet: true,
		Source: solaceSide("prod", "IN", spec.DestQueue, ""), Target: solaceSide("prod", "OUT", spec.DestQueue, "")}}
	m, _ := Build(wfs, d, Opts{MountStores: true, StatusPassword: "status-literal"})

	if len(m.Security.Users) != 3 {
		t.Fatalf("Security.Users = %+v, want 3 (2 operator + 1 reserved)", m.Security.Users)
	}
	if got := m.Security.Users[0].Roles; len(got) != 1 || got[0] != "admin" {
		t.Errorf("ops roles = %v, want [admin] carried through verbatim", got)
	}
	if got := m.Security.Users[1].Roles; len(got) != 0 {
		t.Errorf("probe roles = %v, want none (no role was authored)", got)
	}
	if got := m.Security.Users[2]; got.Name != spec.StatusUserName || len(got.Roles) != 0 {
		t.Errorf("reserved account = %+v, want %s with no roles so it stays read-only", got, spec.StatusUserName)
	}
	// The caller's own Defaults must still describe what the operator wrote.
	if got := d.Security.Users[0].Roles; len(got) != 1 || got[0] != "admin" {
		t.Errorf("caller Defaults roles = %v, want [admin] unmutated", got)
	}
}

// TestApplyStatusAccessAppendsAfterExistingUsers covers operator-configured
// users: the reserved account is appended last, and existing users are still
// rewired through secretRef exactly as before.
func TestApplyStatusAccessAppendsAfterExistingUsers(t *testing.T) {
	d := &spec.Defaults{Security: spec.Security{Users: []spec.User{
		{Name: "alice", Password: "alice-pass"},
		{Name: "bob", PasswordEnv: "BOB_PASS_ENV"},
	}}}
	wfs := []spec.Workflow{{File: "10.yaml", Enabled: true, SourceSet: true, TargetSet: true,
		Source: solaceSide("prod", "IN", spec.DestQueue, ""), Target: solaceSide("prod", "OUT", spec.DestQueue, "")}}
	m, _ := Build(wfs, d, Opts{MountStores: true, StatusPassword: "status-literal"})

	if len(m.Security.Users) != 3 {
		t.Fatalf("Security.Users = %+v, want 3 (2 existing + 1 reserved)", m.Security.Users)
	}
	last := m.Security.Users[2]
	if last.Name != spec.StatusUserName || last.Password != "status-literal" {
		t.Errorf("last user = %+v, want the reserved account carrying the literal password", last)
	}
	// The two existing users differ on purpose: alice's password is a literal, so
	// it is mounted under the name its position derives, while bob's comes from
	// -env and keeps the variable the operator wrote. Both are placeholders --
	// neither value reaches the config.
	if got := m.Security.Users[0]; got.Password != "${"+securityUserPasswordName("alice")+"}" {
		t.Errorf("existing user alice = %+v, want the derived secretRef placeholder", got)
	}
	if got := m.Security.Users[1]; got.Password != "${BOB_PASS_ENV}" {
		t.Errorf("existing user bob = %+v, want the -env name as the placeholder", got)
	}
	for _, s := range m.Secrets {
		if s.Literal == "status-literal" {
			t.Errorf("status password must never enter Model.Secrets, got %+v", s)
		}
	}
}

// TestApplyStatusAccessExposureIsFixed pins the fixed actuator exposure list:
// applyStatusAccess always sets it to exactly this value and ignores whatever
// Defaults carries (management.exposure is a removed key from here on;
// validate rejects a non-nil value before Build ever runs, so consolidate
// need not read it at all).
func TestApplyStatusAccessExposureIsFixed(t *testing.T) {
	wfs := []spec.Workflow{{File: "10.yaml", Enabled: true, SourceSet: true, TargetSet: true,
		Source: solaceSide("prod", "IN", spec.DestQueue, ""), Target: solaceSide("prod", "OUT", spec.DestQueue, "")}}
	m, _ := Build(wfs, &spec.Defaults{}, Opts{MountStores: true})
	const want = "health,info,metrics,leaderelection,workflows"
	if m.Management.Exposure != want {
		t.Errorf("Exposure = %q, want %q", m.Management.Exposure, want)
	}
}

func TestBuildStorePathsRawVsMount(t *testing.T) {
	mq := spec.Side{System: spec.SystemMQ, ConnName: "h(1)", QueueManager: "QM", Channel: "C", UserEnv: "MQ_USER", PasswordEnv: "MQ_PASSWORD", TLS: true,
		Truststore: "mq-ca", Keystore: "mq-client", KeyAlias: "mc", DestKind: spec.DestQueue, Dest: "IN"}
	sol := spec.Side{System: spec.SystemSolace, Host: "tcps://b", MsgVPN: "v", ClientUserEnv: "SOL_USER", ClientPassEnv: "SOL_PASSWORD", DestKind: spec.DestQueue, Dest: "OUT"}
	d := &spec.Defaults{TLS: spec.TLSConfig{Stores: map[string]*spec.Store{
		"mq-ca":     {File: "./certs/t.jks", PasswordEnv: "TRUSTSTORE_PASSWORD_ENV", Type: "JKS"},
		"mq-client": {File: "./certs/k.jks", PasswordEnv: "KEYSTORE_PASSWORD_ENV", Type: "JKS"},
	}}}
	wfs := []spec.Workflow{{File: "10.yaml", Enabled: true, SourceSet: true, TargetSet: true, Source: mq, Target: sol}}
	raw, _ := Build(wfs, d, Opts{MountStores: false}) // config: reflect the env.yaml path verbatim
	if raw.Bundles[0].TruststoreLoc != "./certs/t.jks" {
		t.Fatalf("config (mount=false) truststore loc = %q, want ./certs/t.jks", raw.Bundles[0].TruststoreLoc)
	}
	mnt, _ := Build(wfs, d, Opts{MountStores: true}) // deploy: rewrite to the container mount path
	if mnt.Bundles[0].TruststoreLoc != "/app/external/classpath/truststores/mq-ca" {
		t.Fatalf("deploy (mount=true) truststore loc = %q", mnt.Bundles[0].TruststoreLoc)
	}
	if mnt.Bundles[0].KeystoreLoc != "/app/external/classpath/truststores/mq-client" {
		t.Fatalf("deploy (mount=true) keystore loc = %q", mnt.Bundles[0].KeystoreLoc)
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
	noopSecretRef, _ := testSecretRef()
	dangling := &spec.Defaults{LeaderElection: spec.LeaderElection{
		Present: true, Mode: spec.LeaderActiveStby, Queue: "mgmt-q", ConnRef: "missing",
	}}
	var warns []string
	le := buildLeaderElection(dangling, true, noopSecretRef, fixedLeaderNames, &warns)
	if le == nil || le.Mode != spec.LeaderActiveStby || le.Queue != "mgmt-q" {
		t.Fatalf("dangling conn-ref: le = %+v", le)
	}
	if le.Session != nil {
		t.Errorf("dangling conn-ref: Session = %+v, want nil", le.Session)
	}

	// conn-ref happy path with a tcps:// host: the TLS APIProps branch populates
	// (SSL trust-store + mTLS key-store props from the stores the connection selects).
	happy := &spec.Defaults{
		LeaderElection: spec.LeaderElection{Present: true, Mode: spec.LeaderActiveActive, Queue: "mgmt-q", ConnRef: "edge"},
		Connections: map[string]spec.Side{
			"edge": {System: spec.SystemSolace, Host: "tcps://b:55443", MsgVPN: "prod", ClientUserEnv: "EDGE_USER", ClientPassEnv: "EDGE_PASS",
				Truststore: "edge-ca", Keystore: "edge-client", KeyAlias: "sc",
				APIProps: propsNode(t, "REAPPLY_SUBSCRIPTIONS: true\n")},
		},
		SolaceDefaults: propsNode(t, "connect-retries: -1\nreconnect-retries: -1\n"),
		TLS: spec.TLSConfig{Stores: map[string]*spec.Store{
			"edge-ca":     {File: "./certs/truststore.jks", PasswordEnv: "TRUSTSTORE_PASSWORD_ENV", Type: "JKS"},
			"edge-client": {File: "./certs/keystore.jks", PasswordEnv: "KEYSTORE_PASSWORD_ENV", Type: "JKS"},
		}},
	}
	happySecretRef, happySecrets := testSecretRef()
	le = buildLeaderElection(happy, true, happySecretRef, fixedLeaderNames, &warns)
	if le == nil || le.Session == nil || le.Session.Host != "tcps://b:55443" || le.Session.MsgVPN != "prod" {
		t.Fatalf("conn-ref happy path: le = %+v", le)
	}
	// No credential value or host env-var name ever reaches the rendered
	// session: ClientUser/ClientPass carry the mount-name placeholder, never the
	// credential value. For an -env credential that mount name is the operator's
	// own variable, so an `existing:` Secret's keys are the names written in the
	// spec; the fixed LEADER_ELECTION_* names are what a literal would fall back
	// to instead.
	if le.Session.ClientUser != "${EDGE_USER}" {
		t.Errorf("conn-ref happy path: Session.ClientUser = %q, want the ${EDGE_USER} placeholder", le.Session.ClientUser)
	}
	if le.Session.ClientPass != "${EDGE_PASS}" {
		t.Errorf("conn-ref happy path: Session.ClientPass = %q, want the ${EDGE_PASS} placeholder", le.Session.ClientPass)
	}
	got := map[string]string{}
	for _, p := range le.Session.APIProps {
		got[p.Key] = p.Val
	}
	want := map[string]string{
		"SSL_TRUST_STORE":       "/app/external/classpath/truststores/edge-ca",
		"SSL_KEY_STORE":         "/app/external/classpath/truststores/edge-client",
		"SSL_PRIVATE_KEY_ALIAS": "sc",
	}
	for k, w := range want {
		if got[k] != w {
			t.Errorf("conn-ref happy path: APIProps[%s] = %q, want %q", k, got[k], w)
		}
	}
	secrets := happySecrets()
	byStable := map[string]SecretRef{}
	for _, s := range secrets {
		byStable[s.Stable] = s
	}
	// Stable is the -env variable itself, so each of these is keyed by its own
	// name and carries that name as EnvVar too.
	for _, name := range []string{"EDGE_USER", "EDGE_PASS", "TRUSTSTORE_PASSWORD_ENV"} {
		if s, ok := byStable[name]; !ok || s.EnvVar != name {
			t.Errorf("Secrets[%s] = %+v (present=%v), want an entry whose EnvVar is %s", name, s, ok, name)
		}
	}
	// solace-defaults reaches the session as direct session.* keys, in authored
	// order, exactly as it reaches a binder's solace.java.*.
	wantExtras := []Prop{{Key: "connect-retries", Val: "-1"}, {Key: "reconnect-retries", Val: "-1"}}
	if !reflect.DeepEqual(le.Session.Extras, wantExtras) {
		t.Errorf("conn-ref happy path: Extras = %+v, want %+v", le.Session.Extras, wantExtras)
	}
	// The connection's own verbatim api-properties land AFTER the tool-managed
	// TLS keys, so a tool key always wins the position it needs.
	last := le.Session.APIProps[len(le.Session.APIProps)-1]
	if last.Key != "REAPPLY_SUBSCRIPTIONS" || last.Val != "true" {
		t.Errorf("conn-ref happy path: last APIProp = %+v, want the connection passthrough", last)
	}

	// inline session (no conn-ref), non-tcps host: no TLS APIProps are added.
	inline := &spec.Defaults{
		LeaderElection: spec.LeaderElection{
			Present: true, Mode: spec.LeaderActiveStby, Queue: "mgmt-q",
			Session: &spec.Side{System: spec.SystemSolace, Host: "tcp://b:55555", MsgVPN: "v", ClientUserEnv: "INLINE_USER", ClientPassEnv: "INLINE_PASS",
				APIProps: propsNode(t, "REAPPLY_SUBSCRIPTIONS: true\n")},
		},
		SolaceDefaults: propsNode(t, "connect-retries: -1\n"),
	}
	inlineSecretRef, _ := testSecretRef()
	le = buildLeaderElection(inline, true, inlineSecretRef, fixedLeaderNames, &warns)
	if le == nil || le.Session == nil || le.Session.Host != "tcp://b:55555" {
		t.Fatalf("inline session: le = %+v", le)
	}
	// An inline session is named the same way a conn-ref one is: the placeholder
	// is the mount name, which for an -env credential is the operator's variable.
	// What it must never be is the credential value.
	if le.Session.ClientUser != "${INLINE_USER}" {
		t.Errorf("inline session: Session.ClientUser = %q, want the ${INLINE_USER} placeholder (never the credential value)", le.Session.ClientUser)
	}
	// A plaintext host adds no tool TLS props, but the inline block's own
	// api-properties and solace-defaults must still come through -- they are what
	// the session used to drop silently.
	if want := []Prop{{Key: "REAPPLY_SUBSCRIPTIONS", Val: "true"}}; !reflect.DeepEqual(le.Session.APIProps, want) {
		t.Errorf("inline session: APIProps = %+v, want only the inline passthrough %+v", le.Session.APIProps, want)
	}
	if want := []Prop{{Key: "connect-retries", Val: "-1"}}; !reflect.DeepEqual(le.Session.Extras, want) {
		t.Errorf("inline session: Extras = %+v, want %+v", le.Session.Extras, want)
	}

	// mount rewrite in both states: config (mount=false) keeps the env.yaml path
	// verbatim; deploy (mount=true) rewrites to the container mount path.
	mountable := &spec.Defaults{
		LeaderElection: spec.LeaderElection{Present: true, Mode: spec.LeaderActiveStby, Queue: "mgmt-q", ConnRef: "edge"},
		Connections: map[string]spec.Side{
			"edge": {System: spec.SystemSolace, Host: "tcps://b:55443", MsgVPN: "prod", ClientUserEnv: "EDGE_USER", ClientPassEnv: "EDGE_PASS", Truststore: "edge-ca"},
		},
		TLS: spec.TLSConfig{Stores: map[string]*spec.Store{"edge-ca": {File: "./certs/truststore.jks", PasswordEnv: "TRUSTSTORE_PASSWORD_ENV", Type: "JKS"}}},
	}
	rawSecretRef, _ := testSecretRef()
	mntSecretRef, _ := testSecretRef()
	raw := buildLeaderElection(mountable, false, rawSecretRef, fixedLeaderNames, &warns)
	mnt := buildLeaderElection(mountable, true, mntSecretRef, fixedLeaderNames, &warns)
	if rawTS := trustStoreVal(raw); rawTS != "./certs/truststore.jks" {
		t.Errorf("config (mount=false) truststore = %q, want ./certs/truststore.jks", rawTS)
	}
	if mntTS := trustStoreVal(mnt); mntTS != "/app/external/classpath/truststores/edge-ca" {
		t.Errorf("deploy (mount=true) truststore = %q", mntTS)
	}

	if len(warns) != 0 {
		t.Errorf("no case here collides with a tool-managed key; warns = %v", warns)
	}
}

// TestBuildLeaderElectionSessionPassthroughCollision pins the session half of
// the passthrough rule: a verbatim key colliding with a tool-managed TLS key is
// dropped, the tool value is kept, and the warning names the session rather than
// a binder -- it has no binder name to quote.
func TestBuildLeaderElectionSessionPassthroughCollision(t *testing.T) {
	d := &spec.Defaults{
		LeaderElection: spec.LeaderElection{Present: true, Mode: spec.LeaderActiveStby, Queue: "mgmt-q", ConnRef: "edge"},
		Connections: map[string]spec.Side{
			"edge": {System: spec.SystemSolace, Host: "tcps://b:55443", MsgVPN: "prod", ClientUserEnv: "EDGE_USER", ClientPassEnv: "EDGE_PASS", Truststore: "edge-ca",
				APIProps: propsNode(t, "SSL_TRUST_STORE: /wrong.jks\nREAPPLY_SUBSCRIPTIONS: true\n")},
		},
		TLS: spec.TLSConfig{Stores: map[string]*spec.Store{"edge-ca": {File: "./certs/truststore.jks", PasswordEnv: "TRUSTSTORE_PASSWORD_ENV", Type: "JKS"}}},
	}
	secretRef, _ := testSecretRef()
	var warns []string
	le := buildLeaderElection(d, false, secretRef, fixedLeaderNames, &warns)
	if ts := trustStoreVal(le); ts != "./certs/truststore.jks" {
		t.Errorf("SSL_TRUST_STORE = %q, want the tool value to win", ts)
	}
	if last := le.Session.APIProps[len(le.Session.APIProps)-1]; last.Key != "REAPPLY_SUBSCRIPTIONS" {
		t.Errorf("the non-colliding passthrough key should survive, got %+v", le.Session.APIProps)
	}
	want := `leader-election session: passthrough overrides tool-managed key "SSL_TRUST_STORE"; tool value kept`
	if len(warns) != 1 || warns[0] != want {
		t.Errorf("warns = %v, want exactly [%q]", warns, want)
	}
}

// TestBuildLeaderElectionWarningsReachBuild proves the session's passthrough
// warnings escape Build the way a binder's do. buildLeaderElection had no access
// to the warning slice at all before, so a warning nobody ever saw is exactly
// the failure this guards.
func TestBuildLeaderElectionWarningsReachBuild(t *testing.T) {
	d := &spec.Defaults{
		LeaderElection: spec.LeaderElection{Present: true, Mode: spec.LeaderActiveStby, Queue: "mgmt-q", ConnRef: "edge"},
		Connections: map[string]spec.Side{
			"edge": {System: spec.SystemSolace, Host: "tcps://b:55443", MsgVPN: "prod", ClientUserEnv: "EDGE_USER", ClientPassEnv: "EDGE_PASS", Truststore: "edge-ca",
				APIProps: propsNode(t, "SSL_TRUST_STORE: /wrong.jks\n")},
		},
		TLS: spec.TLSConfig{Stores: map[string]*spec.Store{"edge-ca": {File: "./certs/truststore.jks", PasswordEnv: "TRUSTSTORE_PASSWORD_ENV", Type: "JKS"}}},
	}
	_, warns := Build(nil, d, Opts{})
	if !containsSub(warns, "leader-election session: passthrough overrides tool-managed key") {
		t.Errorf("Build warns = %v, want the session collision among them", warns)
	}
}

// TestBuildLeaderElectionSharesBinderSecretNames covers the two halves of the
// credential-naming rule. A session that resolves to a connection a workflow
// already uses is the same credential, so it carries that binder's stable names
// and mounts one secret; a session on a broker no workflow touches has no binder
// to share with and falls back to the fixed LEADER_ELECTION_* pair.
func TestBuildLeaderElectionSharesBinderSecretNames(t *testing.T) {
	conn := spec.Side{System: spec.SystemSolace, Host: "tcps://b:55443", MsgVPN: "prod", ClientUserEnv: "EDGE_USER", ClientPassEnv: "EDGE_PASS"}
	defs := func(leRef string) *spec.Defaults {
		return &spec.Defaults{
			LeaderElection: spec.LeaderElection{Present: true, Mode: spec.LeaderActiveStby, Queue: "mgmt-q", ConnRef: leRef},
			Connections:    map[string]spec.Side{"edge": conn, "mgmt-only": {System: spec.SystemSolace, Host: "tcp://m:55555", MsgVPN: "m", ClientUserEnv: "M_USER", ClientPassEnv: "M_PASS"}},
		}
	}
	src := spec.Side{System: spec.SystemSolace, ConnRef: "edge", DestKind: spec.DestQueue, Dest: "Q1"}
	wfs := []spec.Workflow{{File: "10.yaml", Enabled: true, SourceSet: true, TargetSet: true,
		Source: src, Target: mqSide("QM1", "MQ1", spec.DestQueue, false)}}

	// A session sharing a binder's connection reuses that binder's credential
	// names, which under the -env rule are the operator's own variables.
	shared, _ := Build(wfs, defs("edge"), Opts{})
	if got, want := shared.LeaderElection.Session.ClientUser, "${EDGE_USER}"; got != want {
		t.Errorf("shared session ClientUser = %q, want %q", got, want)
	}
	if got, want := shared.LeaderElection.Session.ClientPass, "${EDGE_PASS}"; got != want {
		t.Errorf("shared session ClientPass = %q, want %q", got, want)
	}
	// Sharing means reusing, not mounting again: each of the binder's two
	// credentials must appear exactly once, whatever it is called.
	for _, name := range []string{"EDGE_USER", "EDGE_PASS"} {
		var n int
		for _, sec := range shared.Secrets {
			if sec.Stable == name {
				n++
			}
		}
		if n != 1 {
			t.Errorf("%s appears %d times in Secrets, want 1 -- a shared session must not mount a second copy: %+v", name, n, shared.Secrets)
		}
	}

	// A management-only connection is nobody's binder, but its credentials are
	// still -env, so it too is mounted under the operator's own names.
	lone, _ := Build(wfs, defs("mgmt-only"), Opts{})
	if got, want := lone.LeaderElection.Session.ClientUser, "${M_USER}"; got != want {
		t.Errorf("management-only session ClientUser = %q, want %q", got, want)
	}
}

// TestBuildSolaceBinderPassesValidateCertificateThrough pins the neutral stance:
// even selecting the truststore MQ uses, a tcps:// Solace binder and its
// leader-election session get only store wiring from the tool, so the
// connection's own SSL_VALIDATE_CERTIFICATE: false reaches the connector instead
// of being dropped as a collision with a tool-written true. The MQ bundle is
// untouched and the shared store's password is still one secret.
func TestBuildSolaceBinderPassesValidateCertificateThrough(t *testing.T) {
	d := &spec.Defaults{
		TLS: spec.TLSConfig{Stores: map[string]*spec.Store{"mq-ca": {File: "./certs/truststore.jks", PasswordEnv: "TRUSTSTORE_PASSWORD_ENV", Type: "JKS"}}},
		Connections: map[string]spec.Side{
			"edge": {System: spec.SystemSolace, Host: "tcps://b:55443", MsgVPN: "prod", ClientUserEnv: "EDGE_USER", ClientPassEnv: "EDGE_PASS", Truststore: "mq-ca",
				APIProps: propsNode(t, "SSL_VALIDATE_CERTIFICATE: false\nREAPPLY_SUBSCRIPTIONS: true\n")},
		},
		LeaderElection: spec.LeaderElection{Present: true, Mode: spec.LeaderActiveStby, Queue: "mgmt-q", ConnRef: "edge"},
	}
	mq := spec.Side{System: spec.SystemMQ, ConnName: "h(1414)", QueueManager: "QM1", Channel: "CH", UserEnv: "MQ_USER", PasswordEnv: "MQ_PASSWORD", TLS: true, Cipher: "CIPH", Truststore: "mq-ca", DestKind: spec.DestQueue, Dest: "IN"}
	sol := spec.Side{System: spec.SystemSolace, ConnRef: "edge", DestKind: spec.DestTopic, Dest: "out/t"}
	m, warns := Build([]spec.Workflow{{File: "10.yaml", Enabled: true, SourceSet: true, TargetSet: true, Source: mq, Target: sol}}, d, Opts{})
	for _, w := range warns {
		if strings.Contains(w, "tool-managed") {
			t.Errorf("unexpected collision warning: %s", w)
		}
	}
	wantKeys := "SSL_TRUST_STORE,SSL_TRUST_STORE_PASSWORD,SSL_TRUST_STORE_FORMAT,SSL_VALIDATE_CERTIFICATE,REAPPLY_SUBSCRIPTIONS"
	check := func(where string, props []Prop) {
		var keys []string
		for _, p := range props {
			keys = append(keys, p.Key)
			if p.Key == "SSL_VALIDATE_CERTIFICATE" && p.Val != "false" {
				t.Errorf("%s: SSL_VALIDATE_CERTIFICATE = %q, want the connection's false", where, p.Val)
			}
		}
		if got := strings.Join(keys, ","); got != wantKeys {
			t.Errorf("%s: api-properties keys = %s, want %s", where, got, wantKeys)
		}
	}
	var found bool
	for _, b := range m.Binders {
		if b.Kind == spec.SystemSolace {
			found = true
			check("binder "+b.Name, b.Solace.APIProps)
		}
	}
	if !found {
		t.Fatal("no Solace binder built")
	}
	if m.LeaderElection == nil || m.LeaderElection.Session == nil {
		t.Fatal("leader-election session not built")
	}
	check("leader-election session", m.LeaderElection.Session.APIProps)
	if len(m.Bundles) != 1 || m.Bundles[0].TruststoreLoc != "./certs/truststore.jks" {
		t.Fatalf("MQ bundle = %+v, want one bundle on the shared truststore", m.Bundles)
	}
	var refs int
	for _, s := range m.Secrets {
		if s.EnvVar == "TRUSTSTORE_PASSWORD_ENV" {
			refs++
		}
	}
	if refs != 1 {
		t.Errorf("truststore password secrets = %d, want 1", refs)
	}
}

// TestBuildSeparateStoresPerConnection pins named stores end to end in the
// model: each connection is wired to the stores it selects, every selected
// store is listed once in Model.Stores in first-use order, a store used as a
// keystore anywhere is Private, each store's password is its own secret, and an
// MQ binder that selects nothing gets no bundle and no warning.
func TestBuildSeparateStoresPerConnection(t *testing.T) {
	d := &spec.Defaults{
		TLS: spec.TLSConfig{Stores: map[string]*spec.Store{
			"mq-ca":         {File: "./certs/mq-ca.jks", Password: "mq-ca-pw", Type: "JKS"},
			"solace-ca":     {File: "./certs/sol-ca.p12", PasswordEnv: "SOL_TS_PW", Type: "PKCS12"},
			"solace-client": {File: "./certs/sol-client.p12", PasswordEnv: "SOL_KS_PW", Type: "PKCS12"},
			"unused":        {File: "./certs/unused.jks", PasswordEnv: "UNUSED_PW"},
		}},
		LeaderElection: spec.LeaderElection{Present: true, Mode: spec.LeaderActiveStby, Queue: "mgmt-q",
			Session: &spec.Side{System: spec.SystemSolace, Host: "tcps://mgmt:55443", MsgVPN: "v", ClientUserEnv: "L_USER", ClientPassEnv: "L_PASS", Truststore: "solace-ca", Keystore: "mq-ca"}},
	}
	mq := spec.Side{System: spec.SystemMQ, ConnName: "h(1414)", QueueManager: "QM1", Channel: "CH", UserEnv: "MQ_USER", PasswordEnv: "MQ_PASSWORD", TLS: true, Cipher: "CIPH", Truststore: "mq-ca", DestKind: spec.DestQueue, Dest: "IN"}
	plainMQ := spec.Side{System: spec.SystemMQ, ConnName: "h2(1414)", QueueManager: "QM2", Channel: "CH", UserEnv: "MQ_USER", PasswordEnv: "MQ_PASSWORD", TLS: true, Cipher: "CIPH", DestKind: spec.DestQueue, Dest: "OUT2"}
	sol := spec.Side{System: spec.SystemSolace, Host: "tcps://b:55443", MsgVPN: "v", ClientUserEnv: "SOL_USER", ClientPassEnv: "SOL_PASSWORD", Truststore: "solace-ca", Keystore: "solace-client", KeyAlias: "c", DestKind: spec.DestTopic, Dest: "out/t"}
	wfs := []spec.Workflow{
		{File: "10.yaml", Enabled: true, SourceSet: true, TargetSet: true, Source: mq, Target: sol},
		{File: "20.yaml", Enabled: true, SourceSet: true, TargetSet: true, Source: spec.Side{System: spec.SystemSolace, Host: "tcps://b:55443", MsgVPN: "v", ClientUserEnv: "SOL_USER", ClientPassEnv: "SOL_PASSWORD", Truststore: "solace-ca", Keystore: "solace-client", KeyAlias: "c", DestKind: spec.DestQueue, Dest: "Q"}, Target: plainMQ},
	}
	m, warns := Build(wfs, d, Opts{MountStores: true})
	for _, w := range warns {
		if strings.Contains(w, "truststore") || strings.Contains(w, "bundle") {
			t.Errorf("unexpected TLS warning: %s", w)
		}
	}
	want := []StoreUse{
		{Name: "mq-ca", File: "./certs/mq-ca.jks", Type: "JKS", Private: true},
		{Name: "solace-ca", File: "./certs/sol-ca.p12", Type: "PKCS12"},
		{Name: "solace-client", File: "./certs/sol-client.p12", Type: "PKCS12", Private: true},
	}
	if !reflect.DeepEqual(m.Stores, want) {
		t.Errorf("Stores = %+v, want %+v (binders in appearance order; mq-ca turns private as the session's keystore)", m.Stores, want)
	}
	if len(m.Bundles) != 1 || m.Bundles[0].TruststoreLoc != "/app/external/classpath/truststores/mq-ca" || m.Bundles[0].HasKeystore {
		t.Fatalf("Bundles = %+v, want one truststore-only bundle on mq-ca", m.Bundles)
	}
	for _, b := range m.Binders {
		if b.Kind == spec.SystemMQ && b.MQ.QueueManager == "QM2" && b.MQ.SSLBundle != "" {
			t.Errorf("an MQ binder selecting no store must reference no bundle, got %q", b.MQ.SSLBundle)
		}
	}
	got := map[string]string{}
	for _, b := range m.Binders {
		if b.Kind == spec.SystemSolace {
			for _, p := range b.Solace.APIProps {
				got[p.Key] = p.Val
			}
		}
	}
	if got["SSL_TRUST_STORE"] != "/app/external/classpath/truststores/solace-ca" || got["SSL_KEY_STORE"] != "/app/external/classpath/truststores/solace-client" || got["SSL_PRIVATE_KEY_ALIAS"] != "c" {
		t.Errorf("solace api-properties = %v", got)
	}
	sess := map[string]string{}
	for _, p := range m.LeaderElection.Session.APIProps {
		sess[p.Key] = p.Val
	}
	if sess["SSL_TRUST_STORE"] != "/app/external/classpath/truststores/solace-ca" || sess["SSL_KEY_STORE"] != "/app/external/classpath/truststores/mq-ca" {
		t.Errorf("session api-properties = %v", sess)
	}
	names := map[string]bool{}
	for _, s := range m.Secrets {
		names[s.Stable] = true
	}
	for _, n := range []string{StorePasswordName("mq-ca"), "SOL_TS_PW", "SOL_KS_PW"} {
		if !names[n] {
			t.Errorf("Secrets lack %s: %+v", n, m.Secrets)
		}
	}
	if names["UNUSED_PW"] {
		t.Error("a store no connection selects must not be referenced")
	}
}

// TestBuildMQKeystoreOnlyBundle pins that an MQ binder selecting only a keystore
// gets a bundle carrying its client certificate and no truststore section --
// before named stores the keystore was dropped whenever no truststore existed.
func TestBuildMQKeystoreOnlyBundle(t *testing.T) {
	d := &spec.Defaults{TLS: spec.TLSConfig{Stores: map[string]*spec.Store{"mq-client": {File: "./certs/k.jks", PasswordEnv: "KS_PW", Type: "JKS"}}}}
	mq := spec.Side{System: spec.SystemMQ, ConnName: "h(1414)", QueueManager: "QM1", Channel: "CH", UserEnv: "MQ_USER", PasswordEnv: "MQ_PASSWORD", TLS: true, Cipher: "CIPH", Keystore: "mq-client", KeyAlias: "app", DestKind: spec.DestQueue, Dest: "IN"}
	sol := spec.Side{System: spec.SystemSolace, Host: "tcp://b:55555", MsgVPN: "v", DestKind: spec.DestTopic, Dest: "t"}
	m, _ := Build([]spec.Workflow{{File: "10.yaml", Enabled: true, SourceSet: true, TargetSet: true, Source: mq, Target: sol}}, d, Opts{})
	if len(m.Bundles) != 1 {
		t.Fatalf("Bundles = %+v, want one", m.Bundles)
	}
	b := m.Bundles[0]
	if b.HasTruststore || !b.HasKeystore || b.KeystoreLoc != "./certs/k.jks" || b.KeyAlias != "app" {
		t.Errorf("bundle = %+v, want keystore only with alias app", b)
	}
	if want := []StoreUse{{Name: "mq-client", File: "./certs/k.jks", Type: "JKS", Private: true}}; !reflect.DeepEqual(m.Stores, want) {
		t.Errorf("Stores = %+v, want %+v", m.Stores, want)
	}
}
