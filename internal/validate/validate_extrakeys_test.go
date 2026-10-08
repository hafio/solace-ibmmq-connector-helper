package validate

import (
	"strings"
	"testing"

	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/spec"
)

// extraMQ is a clean MQ side carrying y as its block's other keys; tls selects
// the TLS checks (with a truststore, see extraDefaults).
func extraMQ(t *testing.T, y string, tls bool) spec.Side {
	t.Helper()
	s := vMQ("M", spec.DestQueue, tls)
	s.Extra = transformNode(t, y)
	return s
}

// extraSolace is a clean Solace side carrying y as its block's other keys.
func extraSolace(t *testing.T, y string) spec.Side {
	t.Helper()
	s := vSolace("Q", spec.DestQueue, "")
	s.Extra = transformNode(t, y)
	return s
}

// runExtra validates a workflow of src -> tgt under d (empty defaults when
// nil) and returns only the findings about other keys -- those naming the
// site -- so the rest of the workflow's own findings stay out of the way.
func runExtra(t *testing.T, src, tgt spec.Side, d *spec.Defaults) (errs, warns []Issue) {
	t.Helper()
	if d == nil {
		d = &spec.Defaults{}
	}
	e, w := Run(Context{Workflows: []spec.Workflow{wf("x.yaml", src, tgt)}, Defaults: d})
	return siteIssues(e), siteIssues(w)
}

// siteIssues keeps the findings an extra-key check produces: they open with
// the site and system, or name a conn-ref side's rule. The event-driven
// advisory for a Solace queue target opens the same way and is not one.
func siteIssues(issues []Issue) []Issue {
	var out []Issue
	for _, i := range issues {
		if strings.HasPrefix(i.Msg, "target solace: producing to queue") {
			continue
		}
		for _, prefix := range []string{"source mq:", "source solace:", "target mq:", "target solace:", "connections.", "leader-election session solace:", "mq-defaults mq:"} {
			if strings.HasPrefix(i.Msg, prefix) {
				out = append(out, i)
				break
			}
		}
	}
	return out
}

// TestExtraKeysQuietForLegitimateProperties pins that the properties the
// connector reads pass through without a word: IBM's starter keys in either
// spelling, a nested block, a list, Solace's direct keys, and a secret-looking
// key that references a mounted credential.
func TestExtraKeysQuietForLegitimateProperties(t *testing.T) {
	for _, c := range []struct {
		name     string
		src, tgt spec.Side
	}{
		{"IBM starter keys", extraMQ(t, "userAuthenticationMQCSP: false\npool:\n  max-connections: 5\napplication-name: x\nclient-id: y\n", false), vSolace("Q", spec.DestQueue, "")},
		{"Solace direct keys", vMQ("M", spec.DestQueue, false), extraSolace(t, "client-name: n\nconnect-retries-per-host: 3\n")},
		{"a list value", extraMQ(t, "reconnect-options: [QMGR, DISABLED]\n", false), vSolace("Q", spec.DestQueue, "")},
		{"a mounted credential reference", extraMQ(t, "trust-store-password: \"${TRUSTSTORE_PASSWORD}\"\n", false), vSolace("Q", spec.DestQueue, "")},
	} {
		t.Run(c.name, func(t *testing.T) {
			errs, warns := runExtra(t, c.src, c.tgt, nil)
			if len(errs) != 0 || len(warns) != 0 {
				t.Errorf("want nothing, got errors %v, warnings %v", errs, warns)
			}
		})
	}
}

// TestExtraKeysManagedByTheToolIsAnError covers the keys the tool writes
// itself: ssl-bundle, derived from tls: and the truststore, and any other
// spelling of a key the tool reads, which the connector would otherwise get
// twice.
func TestExtraKeysManagedByTheToolIsAnError(t *testing.T) {
	for _, c := range []struct{ name, y, want string }{
		{"ssl-bundle", "ssl-bundle: theirs\n", `source mq: "ssl-bundle" is written by the tool itself (from tls: true and tls.truststore), so it cannot be set here; remove it and enable TLS with tls: true and a tls.truststore in env.yaml instead`},
		{"sslBundle", "sslBundle: theirs\n", `source mq: "sslBundle" is written by the tool itself`},
		{"queueManager", "queueManager: OTHER\n", `source mq: "queueManager" is another spelling of queue-manager, which the tool reads only in that form; write queue-manager`},
		{"additionalProperties", "additionalProperties: {}\n", `source mq: "additionalProperties" is another spelling of additional-properties`},
	} {
		t.Run(c.name, func(t *testing.T) {
			errs, _ := runExtra(t, extraMQ(t, c.y, false), vSolace("Q", spec.DestQueue, ""), nil)
			if !issueWith(errs, c.want) {
				t.Errorf("want an error containing %q, got %v", c.want, errs)
			}
		})
	}
	errs, _ := runExtra(t, vMQ("M", spec.DestQueue, false), extraSolace(t, "msgVpn: other\napiProperties: {}\n"), nil)
	for _, want := range []string{
		`target solace: "msgVpn" is another spelling of msg-vpn, which the tool reads only in that form; write msg-vpn`,
		`target solace: "apiProperties" is another spelling of api-properties`,
	} {
		if !issueWith(errs, want) {
			t.Errorf("want an error containing %q, got %v", want, errs)
		}
	}
}

// TestExtraKeysCredentialSpellingIsAnError pins the one spelling mistake that
// would defeat the secrets model: a credential written under another name
// would land in application.yml as a literal.
func TestExtraKeysCredentialSpellingIsAnError(t *testing.T) {
	errs, _ := runExtra(t, extraMQ(t, "Password: x\nUSER: u\n", false), extraSolace(t, "clientPassword: p\nclientUsername: u\n"), nil)
	for _, want := range []string{
		`source mq: "Password" is another spelling of password, a credential: written this way its value would land in application.yml as a literal instead of being mounted as a secret; write password (or password-env) instead`,
		`source mq: "USER" is another spelling of user, a credential`,
		`target solace: "clientPassword" is another spelling of client-password, a credential`,
		`target solace: "clientUsername" is another spelling of client-username, a credential`,
	} {
		if !issueWith(errs, want) {
			t.Errorf("want an error containing %q, got %v", want, errs)
		}
	}
}

// TestExtraKeysCipherSuiteClashesWithCipher covers the one property the tool
// writes under another name: cipher: becomes WMQ_SSL_CIPHER_SUITE, so the
// starter's own ssl-cipher-suite beside it sets the same thing twice. Alone,
// either is fine.
func TestExtraKeysCipherSuiteClashesWithCipher(t *testing.T) {
	d := defsWithStores()
	for _, c := range []struct {
		name, y string
		cipher  string
		want    bool
	}{
		{"ssl-cipher-suite beside cipher", "ssl-cipher-suite: T\n", "TLS_X", true},
		{"sslCipherSuite beside cipher", "sslCipherSuite: T\n", "TLS_X", true},
		{"ssl-cipher-suite alone", "ssl-cipher-suite: T\n", "", false},
	} {
		t.Run(c.name, func(t *testing.T) {
			src := extraMQ(t, c.y, true)
			src.Cipher = c.cipher
			errs, _ := runExtra(t, src, vSolace("Q", spec.DestQueue, ""), d)
			got := issueWith(errs, "both set the cipher suite (the tool writes cipher: as additional-properties.WMQ_SSL_CIPHER_SUITE); keep one of the two")
			if got != c.want {
				t.Errorf("clash error = %v, want %v (errors %v)", got, c.want, errs)
			}
		})
	}
}

// TestExtraKeysUnsafeKeyIsAnError pins the key charset: a key is written into
// application.yml unquoted, so one that is not a plain property name is
// refused rather than left to restructure the document.
func TestExtraKeysUnsafeKeyIsAnError(t *testing.T) {
	for _, c := range []struct{ name, y string }{
		{"a space", "\"a b\": 1\n"},
		{"a colon", "\"a:b\": 1\n"},
		{"a hash", "\"#x\": 1\n"},
		{"empty", "\"\": 1\n"},
	} {
		t.Run(c.name, func(t *testing.T) {
			errs, _ := runExtra(t, extraMQ(t, c.y, false), vSolace("Q", spec.DestQueue, ""), nil)
			if !issueWith(errs, "is not a plain property name (letters, digits and . _ - [ ] only), so it cannot be written into application.yml as a key; rename it") {
				t.Errorf("want the charset error, got %v", errs)
			}
		})
	}
}

// TestExtraKeysEnvSuffixWarns pins that an -env name outside the credential
// pairs is a plain property, not a credential the tool resolves: passed
// through, with a warning naming the pairs that are resolved.
func TestExtraKeysEnvSuffixWarns(t *testing.T) {
	_, warns := runExtra(t, extraMQ(t, "client-id-env: X\nenvironment: prod\n", false), extraSolace(t, "oauth2-token_env: Y\n"), nil)
	for _, want := range []string{
		`source mq: "client-id-env" is not a credential the tool resolves, so it is passed through as ibm.mq.client-id-env unchanged and the connector reads it as a plain property; the credential keys are user and password, each with an -env twin`,
		`target solace: "oauth2-token_env" is not a credential the tool resolves, so it is passed through as solace.java.oauth2-token_env unchanged and the connector reads it as a plain property; the credential keys are client-username and client-password, each with an -env twin`,
	} {
		if !issueWith(warns, want) {
			t.Errorf("want a warning containing %q, got %v", want, warns)
		}
	}
	if issueWith(warns, `"environment"`) {
		t.Errorf("environment does not end in -env, got %v", warns)
	}
}

// TestExtraKeysSecretLookingLiteralWarns covers the secret heuristic: a key
// that reads like a secret, at the top or inside a nested block, carrying a
// literal value warns; a ${...} reference, an empty value and an ordinary key
// do not.
func TestExtraKeysSecretLookingLiteralWarns(t *testing.T) {
	for _, c := range []struct {
		name, y string
		want    string
	}{
		{"a token literal", "token: abc\n", `source mq: "token" looks like a secret, but its value is written into application.yml as a literal; if it is sensitive, reference a credential the tool already mounts as ${NAME} instead (user guide section 9.2)`},
		{"a nested secret", "token-server:\n  endpoint: https://x\n  client-secret: abc\n", `source mq: "token-server.client-secret" looks like a secret`},
		{"a mounted reference", "token: ${TOKEN}\n", ""},
		{"an empty value", "token:\n", ""},
		{"an ordinary key", "application-name: x\n", ""},
	} {
		t.Run(c.name, func(t *testing.T) {
			_, warns := runExtra(t, extraMQ(t, c.y, false), vSolace("Q", spec.DestQueue, ""), nil)
			got := issueWith(warns, "looks like a secret")
			if c.want == "" && got {
				t.Errorf("want no secret warning, got %v", warns)
			}
			if c.want != "" && !issueWith(warns, c.want) {
				t.Errorf("want a warning containing %q, got %v", c.want, warns)
			}
		})
	}
}

// TestExtraKeysNearMissWarns covers the typo guard: a key within an edit or
// two of a key the tool reads is probably that key misspelt, so it is passed
// through with a question; an unrelated or very short key is not.
func TestExtraKeysNearMissWarns(t *testing.T) {
	for _, c := range []struct{ name, y, want string }{
		{"queue-managr", "queue-managr: QM\n", `source mq: "queue-managr" is not a key this tool knows; did you mean "queue-manager"? It is passed through as ibm.mq.queue-managr, which the connector ignores unless it reads that name`},
		{"chanel", "chanel: C\n", `did you mean "channel"?`},
		{"an unrelated key", "ccdt-url: file:///ccdt.json\n", ""},
		{"a very short key", "abc: 1\n", ""},
	} {
		t.Run(c.name, func(t *testing.T) {
			_, warns := runExtra(t, extraMQ(t, c.y, false), vSolace("Q", spec.DestQueue, ""), nil)
			got := issueWith(warns, "did you mean")
			if c.want == "" && got {
				t.Errorf("want no near-miss warning, got %v", warns)
			}
			if c.want != "" && !issueWith(warns, c.want) {
				t.Errorf("want a warning containing %q, got %v", c.want, warns)
			}
		})
	}
	// Another spelling of a key the tool reads is the error, not the question.
	errs, warns := runExtra(t, extraMQ(t, "keyAlias: k\n", false), vSolace("Q", spec.DestQueue, ""), nil)
	if !issueWith(errs, `"keyAlias" is another spelling of key-alias`) || issueWith(warns, "did you mean") {
		t.Errorf("keyAlias: errors %v, warnings %v", errs, warns)
	}
}

// TestExtraKeysTLSInterplayWarns covers the keys that fight the tool's own TLS
// wiring: a JVM cipher-mapping flag the tool already sets, the starter's jks
// stores that ssl-bundle supersedes, and a fixed client-name every replica of
// an active_* deployment would share.
func TestExtraKeysTLSInterplayWarns(t *testing.T) {
	d := defsWithStores()
	_, warns := runExtra(t, extraMQ(t, "use-ibm-cipher-mappings: true\njks:\n  trust-store: /x\n", true), vSolace("Q", spec.DestQueue, ""), d)
	for _, want := range []string{
		`source mq: "use-ibm-cipher-mappings" overrides the -Dcom.ibm.mq.cfg.useIBMCipherMappings=false the tool sets for a TLS connection, so the JCE cipher names in cipher: may no longer apply`,
		`source mq: jks.* is ignored by the connector while ssl-bundle is set; the stores in effect are env.yaml tls.truststore/keystore`,
	} {
		if !issueWith(warns, want) {
			t.Errorf("want a warning containing %q, got %v", want, warns)
		}
	}
	if _, warns := runExtra(t, extraMQ(t, "use-ibm-cipher-mappings: false\n", true), vSolace("Q", spec.DestQueue, ""), d); issueWith(warns, "overrides the -D") {
		t.Errorf("false agrees with the tool's flag, got %v", warns)
	}

	active := &spec.Defaults{LeaderElection: spec.LeaderElection{Present: true, Mode: spec.LeaderActiveStby, Queue: "mgmt-q", ConnRef: "edge"},
		Connections: map[string]spec.Side{"edge": {System: spec.SystemSolace, Host: "tcp://b:55555", MsgVPN: "v", ClientUserEnv: "EU", ClientPassEnv: "EP"}}}
	_, warns = runExtra(t, vMQ("M", spec.DestQueue, false), extraSolace(t, "client-name: fixed\n"), active)
	if !issueWith(warns, `target solace: "client-name" is fixed, but every replica shares this file and a client name must be unique per connection; use a per-instance value such as ${HOSTNAME}, or leave it for the connector to generate`) {
		t.Errorf("want the client-name warning, got %v", warns)
	}
	if _, warns := runExtra(t, vMQ("M", spec.DestQueue, false), extraSolace(t, "client-name: ${HOSTNAME}\n"), active); issueWith(warns, `"client-name" is fixed`) {
		t.Errorf("a per-instance value must not warn, got %v", warns)
	}
	if _, warns := runExtra(t, vMQ("M", spec.DestQueue, false), extraSolace(t, "client-name: fixed\n"), nil); issueWith(warns, `"client-name" is fixed`) {
		t.Errorf("standalone runs one instance, got %v", warns)
	}
}

// TestConnRefSideRejectsExtraKeys pins that another key beside conn-ref is a
// connection field like host or cipher: it belongs on the connection.
func TestConnRefSideRejectsExtraKeys(t *testing.T) {
	src := spec.Side{System: spec.SystemMQ, ConnRef: "qm", DestKind: spec.DestQueue, Dest: "Q", Extra: transformNode(t, "client-id: x\n")}
	errs, _ := Run(Context{Workflows: []spec.Workflow{wf("x.yaml", src, vSolace("Q", spec.DestQueue, ""))}, Defaults: connDefaults()})
	if !hasErr(errs, "may set only queue/topic") {
		t.Errorf("want the strict conn-ref error, got %v", errs)
	}
}

// TestExtraKeysCheckedOnConnectionsSessionAndDefaults pins the other sites,
// each under its own label and on every run (no Lint): a connection in
// env.yaml, the inline management session, and the mq-defaults block.
func TestExtraKeysCheckedOnConnectionsSessionAndDefaults(t *testing.T) {
	d := &spec.Defaults{
		Connections: map[string]spec.Side{
			"qm": {System: spec.SystemMQ, ConnName: "h(1414)", QueueManager: "QM", Channel: "C", Extra: transformNode(t, "queueManager: x\n")},
		},
		LeaderElection: spec.LeaderElection{Present: true, Mode: spec.LeaderActiveStby, Queue: "mgmt-q",
			Session: &spec.Side{System: spec.SystemSolace, Host: "tcp://b:55555", MsgVPN: "v", ClientUserEnv: "EU", ClientPassEnv: "EP", Extra: transformNode(t, "msgVpn: v\n")}},
		MQDefaults: transformNode(t, "channel: x\ntoken: literal\n"),
	}
	errs, warns := Run(Context{Workflows: wfOK(), Defaults: d})
	for _, want := range []string{
		`connections.qm mq: "queueManager" is another spelling of queue-manager`,
		`leader-election session solace: "msgVpn" is another spelling of msg-vpn`,
		`mq-defaults mq: "channel" is a key the tool reads on each connection, not a default; set channel on the connection instead`,
	} {
		if !issueWith(errs, want) {
			t.Errorf("want an error containing %q, got %v", want, errs)
		}
	}
	if !issueWith(warns, `mq-defaults mq: "token" looks like a secret`) {
		t.Errorf("want the mq-defaults secret warning, got %v", warns)
	}
}
