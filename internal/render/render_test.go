package render

import (
	"fmt"
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
	m, _ := consolidate.Build([]spec.Workflow{wf(t, "10.yaml", richSrcYAML)}, d, true)
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
		"management:", "server:", "port: 8090", "include: health,info", "show-details: always",
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
            password: ${T}
            type: JKS
          keystore:
            location: /app/external/classpath/truststores/keystore.jks
            password: ${K}
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
                user: app
                password: ${MQ}
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
                client-username: connector
                client-password: ${SOL}
                connect-retries: -1
                api-properties:
                  SSL_VALIDATE_CERTIFICATE: true
                  SSL_TRUST_STORE: /app/external/classpath/truststores/truststore.jks
                  SSL_TRUST_STORE_PASSWORD: ${T}
                  SSL_TRUST_STORE_FORMAT: JKS
                  SSL_KEY_STORE: /app/external/classpath/truststores/keystore.jks
                  SSL_KEY_STORE_PASSWORD: ${K}
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
          password: ${H}
management:
  server:
    port: 8090
  endpoints:
    web:
      exposure:
        include: health,info
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
	m, _ := consolidate.Build([]spec.Workflow{wf(t, "10.yaml", src)}, &spec.Defaults{}, true)
	out := Application(m)
	for _, no := range []string{"ssl:", "management:", "logging:", "security:"} {
		if strings.Contains(out, no) {
			t.Errorf("unexpected %q with empty defaults:\n%s", no, out)
		}
	}
	if !strings.Contains(out, "type: undefined") {
		t.Error("undefined binder always emitted")
	}
}

func TestApplicationLeaderElection(t *testing.T) {
	src := `
source:
  solace: {conn-ref: edge, queue: Q.IN}
target:
  mq: {conn-ref: qm, queue: OUT}
`
	defs := `
connections:
  edge:
    solace:
      host: tcps://b:55443
      msg-vpn: prod
      client-username: u
      client-password: ${SOL}
      key-alias: sc
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
`
	d, err := spec.ParseDefaults([]byte(defs))
	if err != nil {
		t.Fatal(err)
	}
	m, _ := consolidate.Build([]spec.Workflow{wf(t, "10.yaml", src)}, d, true)
	out := Application(m)
	for _, w := range []string{
		"management:", "leader-election:", "mode: active_standby",
		"fail-over:", "max-attempts: 5", "back-off-multiplier: 1.5",
		"queue: mgmt-q", "session:", "host: tcps://b:55443", "msg-vpn: prod",
		"client-username: u", "client-password: ${SOL}",
		"SSL_TRUST_STORE: /app/external/classpath/truststores/truststore.jks",
		"SSL_PRIVATE_KEY_ALIAS: sc",
	} {
		if !strings.Contains(out, w) {
			t.Errorf("missing %q\n---\n%s", w, out)
		}
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
	m, _ := consolidate.Build([]spec.Workflow{wf(t, "10.yaml", src)}, &spec.Defaults{}, true)
	out := Application(m)
	// Binders still render their identity...
	for _, want := range []string{"conn-name: h(1414)", "queue-manager: QM1", "host: tcp://b:55555", "msg-vpn: v"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q\n---\n%s", want, out)
		}
	}
	// ...but the empty credential lines are omitted entirely (no `key:` null value).
	for _, no := range []string{"user:", "password:", "client-username:", "client-password:"} {
		if strings.Contains(out, no) {
			t.Errorf("empty credential line %q should be omitted\n---\n%s", no, out)
		}
	}
}
