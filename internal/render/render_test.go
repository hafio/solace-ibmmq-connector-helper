package render

import (
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/consolidate"
	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/spec"
)

func wf(t *testing.T, name, y string) spec.Workflow {
	t.Helper()
	w, err := spec.ParseWorkflow([]byte(y), name)
	if err != nil {
		t.Fatalf("%s: %v", name, err)
	}
	return *w
}

// richSrcYAML and richDefsYAML are the shared fixture for the "rich" Application
// case: an MQ->Solace workflow exercising TLS/mTLS, cipher, verbatim passthrough
// (nested and quoted), a durable topic subscription, and the full set of
// defaults blocks (management/security/logging/solace-defaults). Used by both
// TestApplicationRich (Contains spot-checks) and TestApplicationRichExact (full
// output comparison) so the fixture is defined once.
//
// richDefsYAML's management.exposure: health,info is a removed key by this
// point in the change (see internal/spec): ParseDefaults still parses it, but
// consolidate.applyStatusAccess ignores the value entirely and always renders
// the fixed five-entry list. It is left in deliberately, differing from that
// fixed list, to prove the operator value has no effect.
const richSrcYAML = `
source:
  mq:
    conn-name: h(1414)
    queue-manager: QM1
    channel: CH
    user: app
    password: ${MQ}
    tls: true
    cipher: TLS_X
    key-alias: mc
    topic: T/EVENTS
    additional-properties:
      WMQ_SSL_PEER_NAME: "CN=x"
    consumer:
      retry:
        max: 3
        backoffs:
          - 1
          - 2
      items:
        - name: a
        - name: b
      note: "q:v"
target:
  solace:
    host: tcps://b:55443
    msg-vpn: prod
    client-username: connector
    client-password: ${SOL}
    key-alias: sc
    queue: Q.OUT
    api-properties:
      REAPPLY_SUBSCRIPTIONS: true
`

const richDefsYAML = `
tls:
  truststore:
    file: ./certs/truststore.jks
    password: ${T}
    type: JKS
  keystore:
    file: ./certs/keystore.jks
    password: ${K}
    type: JKS
logging:
  level:
    root: INFO
management:
  port: 8090
  exposure: health,info
  health-show-details: always
security:
  enabled: true
  users:
    - name: hc
      password: ${H}
solace-defaults:
  connect-retries: -1
`

// buildRich builds the consolidate.Model for the richSrcYAML/richDefsYAML fixture.
func buildRich(t *testing.T) *consolidate.Model {
	t.Helper()
	d, err := spec.ParseDefaults([]byte(richDefsYAML))
	if err != nil {
		t.Fatal(err)
	}
	m, _ := consolidate.Build([]spec.Workflow{wf(t, "10.yaml", richSrcYAML)}, d, consolidate.Opts{MountStores: true, StatusPassword: "status-literal-pw"})
	return m
}

func TestApplicationRich(t *testing.T) {
	out := Application(buildRich(t))
	for _, w := range []string{
		"spring:", "ssl:", "bundle:", "jks:", "-bundle:", "truststore:",
		"location: /app/external/classpath/truststores/truststore.jks", "keystore:", "key:", "alias: mc",
		"binders:", "type: jms", "queue-manager: QM1", "ssl-bundle:", "additional-properties:",
		"WMQ_SSL_CIPHER_SUITE: TLS_X", `WMQ_SSL_PEER_NAME: "CN=x"`,
		"type: solace", "host: tcps://b:55443", "connect-retries: -1", "api-properties:",
		"SSL_TRUST_STORE:", "SSL_PRIVATE_KEY_ALIAS: sc", "REAPPLY_SUBSCRIPTIONS: true", "type: undefined",
		"destination: T/EVENTS", "destination: Q.OUT",
		"jms:", "consumer:", "destination-type: topic", "durable-subscription-name: solmq-",
		"retry:", "backoffs:", "- 1", "items:", "name: a", "name: b", `note: "q:v"`,
		"destination-type: queue",
		"connector:", "workflows:", "enabled: true", "security:", "- name: hc",
		"management:", "server:", "port: 8090", "include: health,info,metrics,leaderelection,workflows", "show-details: always",
		"logging:", "level:", "root: INFO",
	} {
		if !strings.Contains(out, w) {
			t.Errorf("missing %q\n---\n%s", w, out)
		}
	}
}

// richApplicationWant is the full expected application.yml for the
// richSrcYAML/richDefsYAML fixture (see TestApplicationRichExact). The durable
// subscription name is deterministic: DurableName("h(1414)", "QM1", "T/EVENTS",
// "10.yaml") (uuid.go), pinned here as a literal so a change to that derivation
// is caught by a diff rather than silently reflected back into the const.
const richApplicationWant = `spring:
  ssl:
    bundle:
      jks:
        mq-conn-1-bundle:
          truststore:
            location: /app/external/classpath/truststores/truststore.jks
            password: ${_GEN_TRUSTSTORE_PASSWORD}
            type: JKS
          keystore:
            location: /app/external/classpath/truststores/keystore.jks
            password: ${_GEN_KEYSTORE_PASSWORD}
            type: JKS
          key:
            alias: mc
  cloud:
    stream:
      binders:
        mq-conn-1:
          type: jms
          environment:
            ibm:
              mq:
                queue-manager: QM1
                channel: CH
                conn-name: h(1414)
                user: ${_GEN_MQ_CONN_1_USER}
                password: ${_GEN_MQ_CONN_1_PASSWORD}
                ssl-bundle: mq-conn-1-bundle
                additional-properties:
                  WMQ_SSL_CIPHER_SUITE: TLS_X
                  WMQ_SSL_PEER_NAME: "CN=x"
        sol-conn-1:
          type: solace
          environment:
            solace:
              java:
                host: tcps://b:55443
                msg-vpn: prod
                client-username: ${_GEN_SOL_CONN_1_CLIENT_USERNAME}
                client-password: ${_GEN_SOL_CONN_1_CLIENT_PASSWORD}
                connect-retries: -1
                api-properties:
                  SSL_VALIDATE_CERTIFICATE: true
                  SSL_TRUST_STORE: /app/external/classpath/truststores/truststore.jks
                  SSL_TRUST_STORE_PASSWORD: ${_GEN_TRUSTSTORE_PASSWORD}
                  SSL_TRUST_STORE_FORMAT: JKS
                  SSL_KEY_STORE: /app/external/classpath/truststores/keystore.jks
                  SSL_KEY_STORE_PASSWORD: ${_GEN_KEYSTORE_PASSWORD}
                  SSL_KEY_STORE_FORMAT: JKS
                  SSL_PRIVATE_KEY_ALIAS: sc
                  REAPPLY_SUBSCRIPTIONS: true
        undefined:
          type: undefined
      bindings:
        input-0:
          destination: T/EVENTS
          binder: mq-conn-1
        output-0:
          destination: Q.OUT
          binder: sol-conn-1
      jms:
        bindings:
          input-0:
            consumer:
              destination-type: topic
              durable-subscription-name: solmq-decf3468-6dfe-5fbe-a194-54753cb38aa3
              retry:
                max: 3
                backoffs:
                  - 1
                  - 2
              items:
                -
                  name: a
                -
                  name: b
              note: "q:v"
      solace:
        bindings:
          output-0:
            producer:
              destination-type: queue
solace:
  connector:
    workflows:
      0:
        enabled: true
    security:
      enabled: true
      users:
        - name: hc
          password: ${_GEN_SECURITY_USER_HC_PASSWORD}
        - name: solmq-status
          password: status-literal-pw
management:
  server:
    port: 8090
  endpoints:
    web:
      exposure:
        include: health,info,metrics,leaderelection,workflows
  endpoint:
    health:
      show-details: always
logging:
  level:
    root: INFO
`

// TestApplicationRichExact is the full-output golden for the rich Application
// case, supplementing TestApplicationRich's Contains spot-checks with an exact
// byte-for-byte comparison (pattern: internal/dockergen/dockergen_test.go:23).
func TestApplicationRichExact(t *testing.T) {
	out := Application(buildRich(t))
	if out != richApplicationWant {
		t.Errorf("Application mismatch\n%s", lineDiff(richApplicationWant, out))
	}
}

// lineDiff returns a compact first-divergence report for two multi-line
// strings (pattern: internal/gen/golden_test.go:369).
func lineDiff(want, got string) string {
	wl := strings.Split(want, "\n")
	gl := strings.Split(got, "\n")
	n := len(wl)
	if len(gl) > n {
		n = len(gl)
	}
	for i := 0; i < n; i++ {
		var wv, gv string
		if i < len(wl) {
			wv = wl[i]
		}
		if i < len(gl) {
			gv = gl[i]
		}
		if wv != gv {
			return fmt.Sprintf("first diff at line %d:\n  want: %q\n  got:  %q", i+1, wv, gv)
		}
	}
	return "(strings differ only in length/trailing content)"
}

func TestApplicationMinimalNoOptionalBlocks(t *testing.T) {
	src := `
source:
  solace:
    host: tcp://b:55555
    msg-vpn: v
    client-username: u
    client-password: x
    queue: IN
target:
  solace:
    host: tcp://b:55555
    msg-vpn: v
    client-username: u
    client-password: x
    queue: OUT
`
	m, _ := consolidate.Build([]spec.Workflow{wf(t, "10.yaml", src)}, &spec.Defaults{}, consolidate.Opts{MountStores: true, StatusPassword: "status-literal-pw"})
	out := Application(m)
	// TLS and logging stay absent with nothing configured. management and
	// security are no longer part of this check: consolidate's
	// applyStatusAccess forces both on unconditionally so the generated status
	// script always has actuator access, regardless of what the operator set.
	for _, no := range []string{"ssl:", "logging:"} {
		if strings.Contains(out, no) {
			t.Errorf("unexpected %q with empty defaults:\n%s", no, out)
		}
	}
	for _, want := range []string{
		"management:", "include: health,info,metrics,leaderelection,workflows",
		"security:", "enabled: true", "- name: solmq-status", "password: status-literal-pw",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q\n---\n%s", want, out)
		}
	}
	if !strings.Contains(out, "type: undefined") {
		t.Error("undefined binder always emitted")
	}
}

// leaderSrc/leaderDefs are shared by the leader-election render tests: one
// connection ("edge") feeds both a workflow side and the management session, so
// the two rendered blocks are built from the same tuple and must agree.
const leaderSrc = `
source:
  solace: {conn-ref: edge, queue: Q.IN}
target:
  mq: {conn-ref: qm, queue: OUT}
`

const leaderDefs = `
connections:
  edge:
    solace:
      host: tcps://b:55443
      msg-vpn: prod
      client-username: u
      client-password: ${SOL}
      key-alias: sc
      api-properties:
        REAPPLY_SUBSCRIPTIONS: true
  qm:
    mq:
      conn-name: h(1414)
      queue-manager: QM1
      channel: C
      user: u
      password: ${MQ}
tls:
  truststore:
    file: ./certs/truststore.jks
    password: ${T}
    type: JKS
  keystore:
    file: ./certs/keystore.jks
    password: ${K}
    type: JKS
leader-election:
  mode: active_standby
  queue: mgmt-q
  conn-ref: edge
  fail-over:
    max-attempts: 5
    back-off-multiplier: 1.5
solace-defaults:
  connect-retries: -1
`

// renderLeaderFixture renders leaderSrc against leaderDefs with mounted stores.
func renderLeaderFixture(t *testing.T) string {
	t.Helper()
	d, err := spec.ParseDefaults([]byte(leaderDefs))
	if err != nil {
		t.Fatal(err)
	}
	m, _ := consolidate.Build([]spec.Workflow{wf(t, "10.yaml", leaderSrc)}, d, consolidate.Opts{MountStores: true})
	return Application(m)
}

func TestApplicationLeaderElection(t *testing.T) {
	out := renderLeaderFixture(t)
	for _, w := range []string{
		"management:", "leader-election:", "mode: active_standby",
		"fail-over:", "max-attempts: 5", "back-off-multiplier: 1.5",
		"queue: mgmt-q",
	} {
		if !strings.Contains(out, w) {
			t.Errorf("missing %q\n---\n%s", w, out)
		}
	}
	// The session block is asserted whole rather than line by line: key ORDER is
	// the contract -- solace-defaults between the credentials and api-properties,
	// verbatim passthrough last -- and a Contains loop cannot see order at all.
	// The credentials carry the edge binder's stable names rather than a second
	// LEADER_ELECTION_* pair, because the session resolves to that connection.
	want := "      session:\n" +
		"        host: tcps://b:55443\n" +
		"        msg-vpn: prod\n" +
		"        client-username: ${_GEN_EDGE_CLIENT_USERNAME}\n" +
		"        client-password: ${_GEN_EDGE_CLIENT_PASSWORD}\n" +
		"        connect-retries: -1\n" +
		"        api-properties:\n" +
		"          SSL_VALIDATE_CERTIFICATE: true\n" +
		"          SSL_TRUST_STORE: /app/external/classpath/truststores/truststore.jks\n" +
		"          SSL_TRUST_STORE_PASSWORD: ${_GEN_TRUSTSTORE_PASSWORD}\n" +
		"          SSL_TRUST_STORE_FORMAT: JKS\n" +
		"          SSL_KEY_STORE: /app/external/classpath/truststores/keystore.jks\n" +
		"          SSL_KEY_STORE_PASSWORD: ${_GEN_KEYSTORE_PASSWORD}\n" +
		"          SSL_KEY_STORE_FORMAT: JKS\n" +
		"          SSL_PRIVATE_KEY_ALIAS: sc\n" +
		"          REAPPLY_SUBSCRIPTIONS: true\n"
	if !strings.Contains(out, want) {
		t.Errorf("session block mismatch, want:\n%s\n---got---\n%s", want, out)
	}
}

// blockKeys returns the ordered key names inside the block introduced by header,
// each prefixed by its indent relative to that block, so two blocks written at
// different depths compare as the same shape.
func blockKeys(t *testing.T, out, header string) []string {
	t.Helper()
	_, rest, ok := strings.Cut(out, header+"\n")
	if !ok {
		t.Fatalf("block %q not found\n---\n%s", header, out)
	}
	base := len(header) - len(strings.TrimLeft(header, " ")) + 2
	var keys []string
	for _, ln := range strings.Split(rest, "\n") {
		if strings.TrimSpace(ln) == "" {
			break
		}
		ind := len(ln) - len(strings.TrimLeft(ln, " "))
		if ind < base {
			break
		}
		k, _, _ := strings.Cut(strings.TrimSpace(ln), ":")
		keys = append(keys, strings.Repeat(" ", ind-base)+k)
	}
	return keys
}

// TestApplicationLeaderElectionSessionMatchesBinderKeySet is the anti-drift
// guard. The connector documents solace.connector.management.session.* as the
// same interface as solace.java.*, but the two blocks are built and rendered by
// separate code and have silently diverged before -- the session used to drop
// solace-defaults and the connection's own api-properties. Rendered from one
// connection the two key sequences must be identical: values differ, keys
// must not.
func TestApplicationLeaderElectionSessionMatchesBinderKeySet(t *testing.T) {
	out := renderLeaderFixture(t)
	binder := blockKeys(t, out, "              java:")
	session := blockKeys(t, out, "      session:")
	if len(binder) == 0 {
		t.Fatal("binder solace.java block is empty")
	}
	if !slices.Equal(binder, session) {
		t.Errorf("session key set drifted from the binder\nbinder:  %q\nsession: %q\n---\n%s", binder, session, out)
	}
}

// TestApplicationLeaderElectionSessionPlaintext covers the non-TLS session: no
// tcps:// host means no tool-managed api-properties, but solace-defaults and the
// connection's own passthrough must still render. No workflow uses this broker,
// so there is no binder to share credential names with and the session falls
// back to the LEADER_ELECTION_* pair.
func TestApplicationLeaderElectionSessionPlaintext(t *testing.T) {
	defs := `
leader-election:
  mode: active_active
  queue: mgmt-q
  session:
    host: tcp://b:55555
    msg-vpn: prod
    client-username: u
    client-password: ${SOL}
    api-properties:
      REAPPLY_SUBSCRIPTIONS: true
solace-defaults:
  connect-retries: -1
  reconnect-retries: -1
`
	d, err := spec.ParseDefaults([]byte(defs))
	if err != nil {
		t.Fatal(err)
	}
	m, _ := consolidate.Build(nil, d, consolidate.Opts{})
	out := Application(m)
	want := "      session:\n" +
		"        host: tcp://b:55555\n" +
		"        msg-vpn: prod\n" +
		"        client-username: ${_GEN_LEADER_ELECTION_CLIENT_USERNAME}\n" +
		"        client-password: ${_GEN_LEADER_ELECTION_CLIENT_PASSWORD}\n" +
		"        connect-retries: -1\n" +
		"        reconnect-retries: -1\n" +
		"        api-properties:\n" +
		"          REAPPLY_SUBSCRIPTIONS: true\n"
	if !strings.Contains(out, want) {
		t.Errorf("plaintext session block mismatch, want:\n%s\n---got---\n%s", want, out)
	}
	if strings.Contains(out, "SSL_") {
		t.Errorf("a plaintext session must carry no tool TLS api-properties\n---\n%s", out)
	}
}

func TestApplicationOmitsEmptyCredentials(t *testing.T) {
	// A connection may omit user/password (cert-based or channel auth). Render
	// must skip the empty lines rather than emit a `key:` with a null value.
	src := `
source:
  mq:
    conn-name: h(1414)
    queue-manager: QM1
    channel: CH
    queue: IN
target:
  solace:
    host: tcp://b:55555
    msg-vpn: v
    topic: OUT
`
	m, _ := consolidate.Build([]spec.Workflow{wf(t, "10.yaml", src)}, &spec.Defaults{}, consolidate.Opts{MountStores: true, StatusPassword: "status-literal-pw"})
	out := Application(m)
	// Binders still render their identity...
	for _, want := range []string{"conn-name: h(1414)", "queue-manager: QM1", "host: tcp://b:55555", "msg-vpn: v"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q\n---\n%s", want, out)
		}
	}
	// ...but the empty credential lines are omitted entirely (no `key:` null
	// value) from the binder section. Everything from "solace:\n  connector:"
	// onward is the forced status-account security block (consolidate's
	// applyStatusAccess), which legitimately carries a password line, so only
	// the binder section checked here is scoped to before it.
	binderSection, _, _ := strings.Cut(out, "solace:\n  connector:")
	for _, no := range []string{"user:", "password:", "client-username:", "client-password:"} {
		if strings.Contains(binderSection, no) {
			t.Errorf("empty credential line %q should be omitted\n---\n%s", no, binderSection)
		}
	}
}

func TestApplicationQuotesRiskyScalars(t *testing.T) {
	// A value carrying ": ", " #", a leading indicator, or a YAML bool/number
	// lookalike would silently restructure or retype the document if emitted
	// plain, so the renderer double-quotes exactly those. Credentials never
	// reach render as literal text any more -- consolidate substitutes a
	// ${STABLE} placeholder for every credential position -- so this now
	// exercises the other spec-sourced scalars that still flow through q()
	// unchanged: MQ conn-name, a destination, and a security user's name.
	src := `
source:
  mq:
    conn-name: "h(1414) #1"
    queue-manager: QM1
    channel: CH
    queue: "key: value"
target:
  solace:
    host: tcp://b:55555
    msg-vpn: v
    topic: OUT
`
	defs := `
security:
  enabled: true
  users:
    - name: "no"
      password: x
`
	d, err := spec.ParseDefaults([]byte(defs))
	if err != nil {
		t.Fatal(err)
	}
	m, _ := consolidate.Build([]spec.Workflow{wf(t, "10.yaml", src)}, d, consolidate.Opts{MountStores: true})
	out := Application(m)
	for _, want := range []string{
		`conn-name: "h(1414) #1"`,   // " #" would start a comment
		`destination: "key: value"`, // ": " would open a nested mapping
		`- name: "no"`,              // reads back as false when plain
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q\n---\n%s", want, out)
		}
	}
	// Ordinary values stay plain, so existing output is byte-for-byte unchanged.
	for _, want := range []string{"queue-manager: QM1", "host: tcp://b:55555"} {
		if !strings.Contains(out, want) {
			t.Errorf("plain value should not be quoted: missing %q\n---\n%s", want, out)
		}
	}
}

func TestApplicationBlockScalarPassthrough(t *testing.T) {
	// A literal (|) passthrough value keeps its newlines through consolidate, so
	// the renderer must re-emit it as a block scalar; concatenating it into
	// "key: value" would push every line after the first to the document root.
	src := `
source:
  mq:
    conn-name: h(1414)
    queue-manager: QM1
    channel: CH
    queue: IN
    additional-properties:
      CERT: |
        -----BEGIN CERTIFICATE-----
        MIIBpayload
        -----END CERTIFICATE-----
target:
  solace:
    host: tcp://b:55555
    msg-vpn: v
    topic: OUT
`
	m, _ := consolidate.Build([]spec.Workflow{wf(t, "10.yaml", src)}, &spec.Defaults{}, consolidate.Opts{MountStores: true})
	out := Application(m)
	want := `                  CERT: |
                    -----BEGIN CERTIFICATE-----
                    MIIBpayload
                    -----END CERTIFICATE-----
`
	if !strings.Contains(out, want) {
		t.Errorf("literal block not re-emitted as a block scalar\nwant:\n%s\ngot:\n%s", want, out)
	}
	// The naive form (everything on the key's line) must not appear.
	if strings.Contains(out, "CERT: -----BEGIN") {
		t.Errorf("multi-line value was flattened onto the key line\n---\n%s", out)
	}
}

func TestApplicationSkipsBundleWithoutTruststore(t *testing.T) {
	// MQ tls: true with no tls.truststore has no store to point a bundle at.
	// Emitting one anyway wrote empty location/password/type lines, which the
	// connector reads as a configured-but-broken store.
	src := `
source:
  mq:
    conn-name: h(1414)
    queue-manager: QM1
    channel: CH
    tls: true
    queue: IN
target:
  solace:
    host: tcp://b:55555
    msg-vpn: v
    topic: OUT
`
	m, warns := consolidate.Build([]spec.Workflow{wf(t, "10.yaml", src)}, &spec.Defaults{}, consolidate.Opts{MountStores: true})
	out := Application(m)
	for _, no := range []string{"ssl:", "bundle:", "ssl-bundle:", "location: ", "truststore:"} {
		if strings.Contains(out, no) {
			t.Errorf("no truststore configured, so %q should not be emitted\n---\n%s", no, out)
		}
	}
	if !m.MQTLS {
		t.Error("the connection still uses TLS, so MQTLS must stay set")
	}
	var found bool
	for _, w := range warns {
		if strings.Contains(w, "tls.truststore is not configured") {
			found = true
		}
	}
	if !found {
		t.Errorf("want a warning that no bundle was emitted, got %v", warns)
	}
}

// TestApplicationSecurityUserRoles pins how roles render: a roles-bearing user
// emits a block-style sequence under its password, and a user without roles
// emits no roles key at all. The omission matters beyond tidiness -- an empty
// list is the connector's read-only default, and it is what keeps the reserved
// solmq-status account read-only and keeps role-less output byte-identical to
// what shipped before roles existed (so the golden fixtures do not move).
func TestApplicationSecurityUserRoles(t *testing.T) {
	src := `
source:
  solace:
    host: tcp://b:55555
    msg-vpn: v
    queue: IN
target:
  solace:
    host: tcp://b:55555
    msg-vpn: v
    queue: OUT
`
	d := &spec.Defaults{Security: spec.Security{Users: []spec.User{
		{Name: "ops", Password: "ops-pass", Roles: []string{"admin", "auditor"}},
		{Name: "probe", Password: "probe-pass"},
	}}}
	m, _ := consolidate.Build([]spec.Workflow{wf(t, "10.yaml", src)}, d,
		consolidate.Opts{MountStores: true, StatusPassword: "status-literal-pw"})
	out := Application(m)

	want := "        - name: ops\n" +
		"          password: ${_GEN_SECURITY_USER_OPS_PASSWORD}\n" +
		"          roles:\n" +
		"            - admin\n" +
		"            - auditor\n"
	if !strings.Contains(out, want) {
		t.Errorf("want the ops user rendered as\n%s---\ngot\n%s", want, out)
	}

	// probe and the reserved account both carry no roles, so neither may emit
	// the key. With the ops block asserted above, a total count of exactly one
	// is what proves the other two render none -- including the reserved
	// account, whose empty list is what keeps it read-only.
	if n := strings.Count(out, "roles:"); n != 1 {
		t.Errorf("roles: appears %d times, want exactly 1 -- only the user that declared roles may emit it (reserved account: %s)\n%s",
			n, spec.StatusUserName, out)
	}
}

// TestApplicationConfigImport verifies Application() leads with
// spring.config.import when Model.ConfigImport is set (so the mounted secret
// files under the secrets mount are read back as properties), and omits the block
// entirely when it is empty (matches gen.ConfigImport, the constant every
// production caller passes through consolidate.Opts.ConfigImport).
func TestApplicationConfigImport(t *testing.T) {
	src := `
source:
  solace:
    host: tcp://b:55555
    msg-vpn: v
    queue: IN
target:
  solace:
    host: tcp://b:55555
    msg-vpn: v
    queue: OUT
`
	const configImport = "optional:configtree:/app/external/var/secrets/"

	withImport, _ := consolidate.Build([]spec.Workflow{wf(t, "10.yaml", src)}, &spec.Defaults{},
		consolidate.Opts{MountStores: true, ConfigImport: configImport})
	out := Application(withImport)
	want := "spring:\n  config:\n    import: " + configImport + "\n"
	if !strings.HasPrefix(out, want) {
		t.Errorf("want Application to lead with %q when ConfigImport is set\n---\n%s", want, out)
	}

	noImport, _ := consolidate.Build([]spec.Workflow{wf(t, "10.yaml", src)}, &spec.Defaults{}, consolidate.Opts{MountStores: true})
	out2 := Application(noImport)
	if strings.Contains(out2, "config:") || strings.Contains(out2, "import:") {
		t.Errorf("want no config.import block when ConfigImport is empty\n---\n%s", out2)
	}
}
