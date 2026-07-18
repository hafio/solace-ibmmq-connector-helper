package render

import (
	"strings"
	"testing"

	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-gen/internal/consolidate"
	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-gen/internal/spec"
)

func wf(t *testing.T, name, y string) spec.Workflow {
	t.Helper()
	w, err := spec.ParseWorkflow([]byte(y), name)
	if err != nil {
		t.Fatalf("%s: %v", name, err)
	}
	return *w
}

func TestApplicationRich(t *testing.T) {
	src := `
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
	defs := `
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
	d, err := spec.ParseDefaults([]byte(defs))
	if err != nil {
		t.Fatal(err)
	}
	m, _ := consolidate.Build([]spec.Workflow{wf(t, "10.yaml", src)}, d, true)
	out := Application(m)
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
