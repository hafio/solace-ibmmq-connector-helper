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

// testSecretRef mirrors the bookkeeping Build's own secretRef closure does
// (see consolidate.go), for white-box tests that call buildLeaderElection
// directly: one entry per stable name, "" for an unset credential, the same
// stable name collapsing repeats onto the first recording.
func testSecretRef() (secretFn, func() []SecretRef) {
	seen := map[string]bool{}
	var secrets []SecretRef
	fn := func(stable string, c spec.Cred) string {
		if c.Empty() {
			return ""
		}
		if !seen[stable] {
			seen[stable] = true
			secrets = append(secrets, SecretRef{Stable: stable, Literal: c.Literal, EnvVar: c.EnvVar})
		}
		return "${" + stable + "}"
	}
	return fn, func() []SecretRef { return secrets }
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
	mq := spec.Side{System: spec.SystemMQ, ConnName: "h(1414)", QueueManager: "QM1", Channel: "CH", UserEnv: "MQ_USER", PasswordEnv: "MQ_PASSWORD", TLS: true, Cipher: "CIPH", KeyAlias: "mc", DestKind: spec.DestQueue, Dest: "IN"}
	sol := spec.Side{System: spec.SystemSolace, Host: "tcps://b", MsgVPN: "v", ClientUserEnv: "SOL_USER", ClientPassEnv: "SOL_PASSWORD", DestKind: spec.DestQueue, Dest: "OUT"}
	d := &spec.Defaults{TLS: spec.TLSConfig{
		Truststore: &spec.Store{File: "t.jks", PasswordEnv: "TRUSTSTORE_PASSWORD_ENV", Type: "JKS"},
		Keystore:   &spec.Store{File: "k.jks", PasswordEnv: "KEYSTORE_PASSWORD_ENV", Type: "PKCS12"},
	}}
	m, _ := Build([]spec.Workflow{{File: "10.yaml", Enabled: true, SourceSet: true, TargetSet: true, Source: mq, Target: sol}}, d, Opts{MountStores: true})
	if !m.MQTLS || len(m.Bundles) != 1 {
		t.Fatalf("MQTLS=%v bundles=%d", m.MQTLS, len(m.Bundles))
	}
	b := m.Bundles[0]
	if !b.HasKeystore || b.KeyAlias != "mc" || b.KeystoreTyp != "PKCS12" || b.TruststoreTyp != "JKS" {
		t.Fatalf("bundle=%+v", b)
	}
	// The truststore password is referenced twice -- once by the MQ SSL bundle,
	// once by the Solace binder's tcps:// api-properties (both TLS-enabled sides
	// share the one truststore in d.TLS) -- and must collapse to a single
	// Model.Secrets entry, not be mounted/resolved twice under the same name.
	var truststoreRefs int
	for _, s := range m.Secrets {
		if s.Stable == TruststorePasswordName {
			truststoreRefs++
			if s.EnvVar != "TRUSTSTORE_PASSWORD_ENV" {
				t.Errorf("truststore secret EnvVar = %q, want TRUSTSTORE_PASSWORD_ENV", s.EnvVar)
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

// TestApplyStatusAccessSecurityAbsent pins the security-absent branch: absent
// means EffectivelyEnabled (spec.Security.EffectivelyEnabled's own doc), so
// Build synthesizes the management/security blocks from nothing and the
// reserved account is the only user, carrying the literal status password
// rather than a secretRef placeholder.
func TestApplyStatusAccessSecurityAbsent(t *testing.T) {
	wfs := []spec.Workflow{{File: "10.yaml", Enabled: true, SourceSet: true, TargetSet: true,
		Source: solaceSide("prod", "IN", spec.DestQueue, ""), Target: solaceSide("prod", "OUT", spec.DestQueue, "")}}
	m, _ := Build(wfs, nil, Opts{MountStores: true, StatusPassword: "status-literal"})

	if !m.Security.Present || !m.Security.Enabled {
		t.Fatalf("Security = %+v, want Present and Enabled both true", m.Security)
	}
	if len(m.Security.Users) != 1 || m.Security.Users[0].Name != spec.StatusUserName || m.Security.Users[0].Password != "status-literal" {
		t.Fatalf("Security.Users = %+v, want one reserved user carrying the literal password", m.Security.Users)
	}
	for _, s := range m.Secrets {
		if s.Literal == "status-literal" || s.Stable == securityUserPasswordName(spec.StatusUserName) {
			t.Errorf("status password must never enter Model.Secrets, got %+v", s)
		}
	}
}

// TestApplyStatusAccessSecurityEnabledAppendsAfterExisting covers security
// present+enabled with existing users: the reserved account is appended last,
// and existing users are still rewired through secretRef exactly as before.
func TestApplyStatusAccessSecurityEnabledAppendsAfterExisting(t *testing.T) {
	d := &spec.Defaults{Security: spec.Security{Present: true, Enabled: true, Users: []spec.User{
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
	if got := m.Security.Users[0]; got.Password != "${"+securityUserPasswordName("alice")+"}" {
		t.Errorf("existing user alice = %+v, want a secretRef placeholder", got)
	}
	if got := m.Security.Users[1]; got.Password != "${"+securityUserPasswordName("bob")+"}" {
		t.Errorf("existing user bob = %+v, want a secretRef placeholder", got)
	}
	for _, s := range m.Secrets {
		if s.Literal == "status-literal" {
			t.Errorf("status password must never enter Model.Secrets, got %+v", s)
		}
	}
}

// TestApplyStatusAccessSecurityDisabled covers explicit security.enabled:
// false: no reserved user is added and Enabled stays false.
func TestApplyStatusAccessSecurityDisabled(t *testing.T) {
	d := &spec.Defaults{Security: spec.Security{Present: true, Enabled: false, Users: []spec.User{
		{Name: "alice", Password: "alice-pass"},
	}}}
	wfs := []spec.Workflow{{File: "10.yaml", Enabled: true, SourceSet: true, TargetSet: true,
		Source: solaceSide("prod", "IN", spec.DestQueue, ""), Target: solaceSide("prod", "OUT", spec.DestQueue, "")}}
	m, _ := Build(wfs, d, Opts{MountStores: true, StatusPassword: "status-literal"})

	if m.Security.Enabled {
		t.Error("Security.Enabled should stay false when explicitly disabled")
	}
	if len(m.Security.Users) != 1 || m.Security.Users[0].Name != "alice" {
		t.Fatalf("Security.Users = %+v, want only the existing user (no reserved account)", m.Security.Users)
	}
}

// TestApplyStatusAccessExposure covers every exposure branch: empty defaults to
// health+leaderelection+workflows, an already-complete list is left unchanged,
// a partial list gets only the missing entry appended, and "*" (Spring's
// wildcard for "everything") is left untouched.
func TestApplyStatusAccessExposure(t *testing.T) {
	cases := []struct {
		name, in, want string
	}{
		{"empty", "", "health,leaderelection,workflows"},
		{"already both", "health,leaderelection,workflows", "health,leaderelection,workflows"},
		{"missing one", "health,leaderelection", "health,leaderelection,workflows"},
		{"wildcard", "*", "*"},
	}
	for _, c := range cases {
		d := &spec.Defaults{Management: spec.Management{Present: true, Exposure: c.in}}
		wfs := []spec.Workflow{{File: "10.yaml", Enabled: true, SourceSet: true, TargetSet: true,
			Source: solaceSide("prod", "IN", spec.DestQueue, ""), Target: solaceSide("prod", "OUT", spec.DestQueue, "")}}
		m, _ := Build(wfs, d, Opts{MountStores: true})
		if !m.Management.Present {
			t.Errorf("%s: Management.Present should always be forced true", c.name)
		}
		if m.Management.Exposure != c.want {
			t.Errorf("%s: Exposure = %q, want %q", c.name, m.Management.Exposure, c.want)
		}
	}
}

// TestHasExposureEntry pins the substring-vs-exact-match distinction: a
// same-prefix entry like "leaderelection2" or a suffixed "leaderelectionx"
// must never be read as "leaderelection", and spaces after a comma are
// trimmed before comparing.
func TestHasExposureEntry(t *testing.T) {
	cases := []struct {
		csv, entry string
		want       bool
	}{
		{"health,leaderelection,workflows", "leaderelection", true},
		{"health, leaderelection, workflows", "leaderelection", true},
		{"leaderelection2", "leaderelection", false},
		{"xleaderelection", "leaderelection", false},
		{"leaderelectionx", "leaderelection", false},
		{"*", "leaderelection", true},
		{"*", "workflows", true},
		{"health", "leaderelection", false},
		{"", "leaderelection", false},
	}
	for _, c := range cases {
		if got := hasExposureEntry(c.csv, c.entry); got != c.want {
			t.Errorf("hasExposureEntry(%q, %q) = %v, want %v", c.csv, c.entry, got, c.want)
		}
	}
}

func TestBuildStorePathsRawVsMount(t *testing.T) {
	mq := spec.Side{System: spec.SystemMQ, ConnName: "h(1)", QueueManager: "QM", Channel: "C", UserEnv: "MQ_USER", PasswordEnv: "MQ_PASSWORD", TLS: true, KeyAlias: "mc", DestKind: spec.DestQueue, Dest: "IN"}
	sol := spec.Side{System: spec.SystemSolace, Host: "tcps://b", MsgVPN: "v", ClientUserEnv: "SOL_USER", ClientPassEnv: "SOL_PASSWORD", DestKind: spec.DestQueue, Dest: "OUT"}
	d := &spec.Defaults{TLS: spec.TLSConfig{
		Truststore: &spec.Store{File: "./certs/t.jks", PasswordEnv: "TRUSTSTORE_PASSWORD_ENV", Type: "JKS"},
		Keystore:   &spec.Store{File: "./certs/k.jks", PasswordEnv: "KEYSTORE_PASSWORD_ENV", Type: "JKS"},
	}}
	wfs := []spec.Workflow{{File: "10.yaml", Enabled: true, SourceSet: true, TargetSet: true, Source: mq, Target: sol}}
	raw, _ := Build(wfs, d, Opts{MountStores: false}) // config: reflect the env.yaml path verbatim
	if raw.Bundles[0].TruststoreLoc != "./certs/t.jks" {
		t.Fatalf("config (mount=false) truststore loc = %q, want ./certs/t.jks", raw.Bundles[0].TruststoreLoc)
	}
	mnt, _ := Build(wfs, d, Opts{MountStores: true}) // deploy: rewrite to the container mount path
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
	noopSecretRef, _ := testSecretRef()
	dangling := &spec.Defaults{LeaderElection: spec.LeaderElection{
		Present: true, Mode: spec.LeaderActiveStby, Queue: "mgmt-q", ConnRef: "missing",
	}}
	le := buildLeaderElection(dangling, true, noopSecretRef)
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
			"edge": {System: spec.SystemSolace, Host: "tcps://b:55443", MsgVPN: "prod", ClientUserEnv: "EDGE_USER", ClientPassEnv: "EDGE_PASS", KeyAlias: "sc"},
		},
		TLS: spec.TLSConfig{
			Truststore: &spec.Store{File: "./certs/truststore.jks", PasswordEnv: "TRUSTSTORE_PASSWORD_ENV", Type: "JKS"},
			Keystore:   &spec.Store{File: "./certs/keystore.jks", PasswordEnv: "KEYSTORE_PASSWORD_ENV", Type: "JKS"},
		},
	}
	happySecretRef, happySecrets := testSecretRef()
	le = buildLeaderElection(happy, true, happySecretRef)
	if le == nil || le.Session == nil || le.Session.Host != "tcps://b:55443" || le.Session.MsgVPN != "prod" {
		t.Fatalf("conn-ref happy path: le = %+v", le)
	}
	// No credential value or host env-var name ever reaches the rendered
	// session: ClientUser/ClientPass carry the stable placeholder, not "u"/"p"
	// or the env-var name, and the leader-election secrets are recorded under
	// their own fixed stable names.
	if le.Session.ClientUser != "${"+LeaderUsernameName+"}" {
		t.Errorf("conn-ref happy path: Session.ClientUser = %q, want the %s placeholder", le.Session.ClientUser, LeaderUsernameName)
	}
	if le.Session.ClientPass != "${"+LeaderPasswordName+"}" {
		t.Errorf("conn-ref happy path: Session.ClientPass = %q, want the %s placeholder", le.Session.ClientPass, LeaderPasswordName)
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
	secrets := happySecrets()
	byStable := map[string]SecretRef{}
	for _, s := range secrets {
		byStable[s.Stable] = s
	}
	if s := byStable[LeaderUsernameName]; s.EnvVar != "EDGE_USER" {
		t.Errorf("Secrets[%s].EnvVar = %q, want EDGE_USER", LeaderUsernameName, s.EnvVar)
	}
	if s := byStable[LeaderPasswordName]; s.EnvVar != "EDGE_PASS" {
		t.Errorf("Secrets[%s].EnvVar = %q, want EDGE_PASS", LeaderPasswordName, s.EnvVar)
	}
	if s := byStable[TruststorePasswordName]; s.EnvVar != "TRUSTSTORE_PASSWORD_ENV" {
		t.Errorf("Secrets[%s].EnvVar = %q, want TRUSTSTORE_PASSWORD_ENV", TruststorePasswordName, s.EnvVar)
	}

	// inline session (no conn-ref), non-tcps host: no TLS APIProps are added.
	inline := &spec.Defaults{LeaderElection: spec.LeaderElection{
		Present: true, Mode: spec.LeaderActiveStby, Queue: "mgmt-q",
		Session: &spec.Side{System: spec.SystemSolace, Host: "tcp://b:55555", MsgVPN: "v", ClientUserEnv: "INLINE_USER", ClientPassEnv: "INLINE_PASS"},
	}}
	inlineSecretRef, _ := testSecretRef()
	le = buildLeaderElection(inline, true, inlineSecretRef)
	if le == nil || le.Session == nil || le.Session.Host != "tcp://b:55555" {
		t.Fatalf("inline session: le = %+v", le)
	}
	if le.Session.ClientUser != "${"+LeaderUsernameName+"}" {
		t.Errorf("inline session: Session.ClientUser = %q, want the %s placeholder (never the literal/env source)", le.Session.ClientUser, LeaderUsernameName)
	}
	if len(le.Session.APIProps) != 0 {
		t.Errorf("inline session: non-tcps host should have no TLS APIProps, got %+v", le.Session.APIProps)
	}

	// mount rewrite in both states: config (mount=false) keeps the env.yaml path
	// verbatim; deploy (mount=true) rewrites to the container mount path.
	mountable := &spec.Defaults{
		LeaderElection: spec.LeaderElection{Present: true, Mode: spec.LeaderActiveStby, Queue: "mgmt-q", ConnRef: "edge"},
		Connections: map[string]spec.Side{
			"edge": {System: spec.SystemSolace, Host: "tcps://b:55443", MsgVPN: "prod", ClientUserEnv: "EDGE_USER", ClientPassEnv: "EDGE_PASS"},
		},
		TLS: spec.TLSConfig{Truststore: &spec.Store{File: "./certs/truststore.jks", PasswordEnv: "TRUSTSTORE_PASSWORD_ENV", Type: "JKS"}},
	}
	rawSecretRef, _ := testSecretRef()
	mntSecretRef, _ := testSecretRef()
	raw := buildLeaderElection(mountable, false, rawSecretRef)
	mnt := buildLeaderElection(mountable, true, mntSecretRef)
	if rawTS := trustStoreVal(raw); rawTS != "./certs/truststore.jks" {
		t.Errorf("config (mount=false) truststore = %q, want ./certs/truststore.jks", rawTS)
	}
	if mntTS := trustStoreVal(mnt); mntTS != "/app/external/classpath/truststores/truststore.jks" {
		t.Errorf("deploy (mount=true) truststore = %q", mntTS)
	}
}
