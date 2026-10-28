package validate

import (
	"fmt"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/spec"
)

func vSolace(dest, kind, keyAlias string) spec.Side {
	return spec.Side{System: spec.SystemSolace, Host: "tcps://b:55443", MsgVPN: "prod",
		ClientUser: "u", ClientPassEnv: "SOLACE_CLIENT_PASSWORD", DestKind: kind, Dest: dest, KeyAlias: keyAlias}
}
func vMQ(dest, kind string, tls bool) spec.Side {
	return spec.Side{System: spec.SystemMQ, ConnName: "h(1414)", QueueManager: "QM1", Channel: "CH",
		User: "u", PasswordEnv: "MQ_PASSWORD", TLS: tls, DestKind: kind, Dest: dest}
}
func wf(file string, src, tgt spec.Side) spec.Workflow {
	return spec.Workflow{File: file, Enabled: true, SourceSet: true, TargetSet: true, Source: src, Target: tgt}
}

// defsWithStores defines two stores, ts (a truststore) and ks (a keystore),
// for sides to select with Truststore: "ts" / Keystore: "ks".
func defsWithStores() *spec.Defaults {
	return &spec.Defaults{TLS: spec.TLSConfig{Stores: map[string]*spec.Store{
		"ts": {Existing: &spec.StoreExisting{PKCS12: "t.jks"}, PasswordEnv: "TRUSTSTORE_PASSWORD"},
		"ks": {Existing: &spec.StoreExisting{PKCS12: "k.jks"}, PasswordEnv: "KEYSTORE_PASSWORD"},
	}}}
}

// imageOK is the top-level image: block every platform-checking context needs
// now that the image is declared once rather than per platform. A context that
// enables any Check* flag without it fails on "image: is required", which is
// the point -- but it is not what most of these tests are about.
func imageOK() *spec.Image {
	return &spec.Image{Name: "solace/connector", Tag: "9.9"}
}

func hasErr(errs []Issue, sub string) bool {
	for _, e := range errs {
		if strings.Contains(e.String(), sub) {
			return true
		}
	}
	return false
}

func TestValidGoldenLikeInputPasses(t *testing.T) {
	sol := vSolace("Q", spec.DestQueue, "sc")
	sol.Truststore, sol.Keystore = "ts", "ks"
	mq := vMQ("MQ", spec.DestQueue, true)
	mq.Truststore = "ts"
	wfs := []spec.Workflow{wf("10.yaml", sol, mq)}
	errs, warns := Run(Context{Workflows: wfs, Defaults: defsWithStores()})
	if len(errs) != 0 {
		t.Fatalf("expected no errors, got %v", errs)
	}
	if hasErr(warns, "is not selected") {
		t.Errorf("both stores are selected, so neither is unused: %v", warns)
	}
}

func TestMissingSourceTarget(t *testing.T) {
	w := spec.Workflow{File: "x.yaml", Enabled: true} // neither side set
	errs, _ := Run(Context{Workflows: []spec.Workflow{w}, Defaults: &spec.Defaults{}})
	if !hasErr(errs, "missing 'source'") || !hasErr(errs, "missing 'target'") {
		t.Fatalf("want missing source+target errors, got %v", errs)
	}
}

func TestExactlyOneSystem(t *testing.T) {
	src := vSolace("Q", spec.DestQueue, "")
	src.System = "" // simulate both/neither system
	errs, _ := Run(Context{Workflows: []spec.Workflow{wf("x.yaml", src, vMQ("MQ", spec.DestQueue, false))}, Defaults: &spec.Defaults{}})
	if !hasErr(errs, "exactly one of 'solace:' or 'mq:'") {
		t.Fatalf("want system error, got %v", errs)
	}
}

func TestExactlyOneDestination(t *testing.T) {
	src := vSolace("Q", "", "") // DestKind empty => both/neither
	errs, _ := Run(Context{Workflows: []spec.Workflow{wf("x.yaml", src, vMQ("MQ", spec.DestQueue, false))}, Defaults: &spec.Defaults{}})
	if !hasErr(errs, "exactly one of 'queue:' or 'topic:'") {
		t.Fatalf("want destination error, got %v", errs)
	}
}

func TestSolaceTopicSourceIsError(t *testing.T) {
	// The connector's Solace consumer binds to an endpoint, so a topic source is
	// rejected outright -- not merely flagged with the EDA advisory it used to carry.
	errs, warns := Run(Context{Workflows: []spec.Workflow{wf("x.yaml", vSolace("t", spec.DestTopic, ""), vMQ("MQ", spec.DestQueue, false))}, Defaults: &spec.Defaults{}})
	if !hasErr(errs, "cannot be consumed from") {
		t.Fatalf("want an error for a Solace topic source, got %v", errs)
	}
	if hasErr(warns, "non-durable subscription") {
		t.Fatalf("the topic-source advisory should be gone, not emitted alongside the error: %v", warns)
	}
}

func TestConnNameFormat(t *testing.T) {
	bad := vMQ("MQ", spec.DestQueue, false)
	bad.ConnName = "not-a-conn"
	errs, _ := Run(Context{Workflows: []spec.Workflow{wf("x.yaml", vSolace("Q", spec.DestQueue, ""), bad)}, Defaults: &spec.Defaults{}})
	if !hasErr(errs, "host(port)") {
		t.Fatalf("want conn-name format error, got %v", errs)
	}
}

func TestKeyAliasNeedsKeystore(t *testing.T) {
	errs, _ := Run(Context{Workflows: []spec.Workflow{wf("x.yaml", vSolace("Q", spec.DestQueue, "sc"), vMQ("MQ", spec.DestQueue, false))}, Defaults: defsWithStores()})
	if !hasErr(errs, `source solace: key-alias "sc" selects a key from a keystore, but no keystore is selected; add keystore: <store>`) {
		t.Fatalf("want keystore error, got %v", errs)
	}
}

// TestStoreSelectionConflicts pins checkStoreSelectionConflicts: sides on one
// binder must select the same truststore, keystore and key-alias, strictly --
// a side selecting nothing conflicts with one selecting a store.
func TestStoreSelectionConflicts(t *testing.T) {
	const tail = ": one connection trusts one store and presents one certificate -- give every side the same value, or define the connection once under connections: and use conn-ref"
	with := func(dest, ts, ks, alias string) spec.Side {
		s := vSolace(dest, spec.DestQueue, alias)
		s.Truststore, s.Keystore = ts, ks
		return s
	}
	for _, c := range []struct {
		name string
		a, b spec.Side
		want string
	}{
		{"key-alias", with("Q1", "", "ks", "alias-a"), with("Q2", "", "ks", "alias-b"), `20.yaml: conflicting key-alias "alias-b" vs "alias-a" for the same binder (also in 10.yaml)` + tail},
		{"truststore set vs none", with("Q1", "ts", "", ""), with("Q2", "", "", ""), `20.yaml: conflicting truststore none vs "ts" for the same binder (also in 10.yaml)` + tail},
		{"keystore differs", with("Q1", "", "ks", ""), with("Q2", "", "ts", ""), `20.yaml: conflicting keystore "ts" vs "ks" for the same binder (also in 10.yaml)` + tail},
	} {
		t.Run(c.name, func(t *testing.T) {
			wfs := []spec.Workflow{
				wf("10.yaml", c.a, vMQ("M1", spec.DestQueue, false)),
				wf("20.yaml", c.b, vMQ("M2", spec.DestQueue, false)),
			}
			errs, _ := Run(Context{Workflows: wfs, Defaults: defsWithStores()})
			if !hasErr(errs, c.want) {
				t.Errorf("want %q, got %v", c.want, errs)
			}
		})
	}
	same := []spec.Workflow{
		wf("10.yaml", with("Q1", "ts", "ks", "a"), vMQ("M1", spec.DestQueue, false)),
		wf("20.yaml", with("Q2", "ts", "ks", "a"), vMQ("M2", spec.DestQueue, false)),
	}
	if errs, _ := Run(Context{Workflows: same, Defaults: defsWithStores()}); hasErr(errs, "conflicting") {
		t.Errorf("identical selections do not conflict: %v", errs)
	}
}

func manyWorkflows(n int) []spec.Workflow {
	var wfs []spec.Workflow
	for i := 0; i < n; i++ {
		wfs = append(wfs, wf(fmt.Sprintf("wf-%02d.yaml", i), vSolace("Q", spec.DestQueue, ""), vMQ("MQ", spec.DestQueue, false)))
	}
	return wfs
}

// TestWorkflowCap pins the new fatal cap: a folder holding more than
// MaxWorkflows workflows is rejected outright (no sharding survives to split
// them), and the message names the remedy (separate folders, each its own
// deployment.name/docker.name/podman.name, each deployed as its own connector).
func TestWorkflowCap(t *testing.T) {
	errs, _ := Run(Context{Workflows: manyWorkflows(MaxWorkflows + 1), Defaults: &spec.Defaults{}})
	if !hasErr(errs, "21 workflows found, but one connector instance runs at most 20 (workflow ids 0..19)") {
		t.Fatalf("want workflow cap error, got %v", errs)
	}
	if !hasErr(errs, "Split them across separate folders, each with its own env.yaml and its own deployment.name/docker.name/podman.name, and deploy each as its own connector") {
		t.Fatalf("want cap error to name the remedy, got %v", errs)
	}

	// Exactly MaxWorkflows must not trip the cap.
	errs2, _ := Run(Context{Workflows: manyWorkflows(MaxWorkflows), Defaults: &spec.Defaults{}})
	if hasErr(errs2, "workflows found") {
		t.Fatalf("MaxWorkflows workflows should not trip the cap, got %v", errs2)
	}
}

// TestDeployNameTooLong pins the kube ConfigMap name-length guard now that
// there is no per-shard instance suffix: it is simply deployment.name +
// "-config" <= 63.
func TestDeployNameTooLong(t *testing.T) {
	longName := strings.Repeat("a", 57) // 57 + len("-config")=7 -> 64 > 63
	k := &spec.Kubernetes{Deployment: spec.Deployment{Name: longName, Namespace: "ns", Replicas: 1}}
	errs, _ := Run(Context{Workflows: leWorkflows(), Defaults: &spec.Defaults{}, Image: imageOK(), Kube: k, CheckKubernetes: true})
	if !hasErr(errs, "exceeds the 63-char DNS-1123 limit") {
		t.Fatalf("want too-long name error, got %v", errs)
	}
	// A short name is fine.
	k.Deployment.Name = "solmq"
	errs2, _ := Run(Context{Workflows: leWorkflows(), Defaults: &spec.Defaults{}, Image: imageOK(), Kube: k, CheckKubernetes: true})
	if hasErr(errs2, "exceeds the 63-char DNS-1123 limit") {
		t.Fatalf("short name should not trip the length guard, got %v", errs2)
	}
}

func leWorkflows() []spec.Workflow {
	return []spec.Workflow{wf("x.yaml", vSolace("Q", spec.DestQueue, ""), vMQ("MQ", spec.DestQueue, false))}
}

func TestLeaderElectionActiveStandbyValid(t *testing.T) {
	d := defsWithStores()
	d.Connections = map[string]spec.Side{"mgmt": mgmtSide()}
	d.LeaderElection = spec.LeaderElection{Present: true, Mode: spec.LeaderActiveStby, Queue: "mgmt-q", ConnRef: "mgmt"}
	if errs, _ := Run(Context{Workflows: leWorkflows(), Defaults: d}); len(errs) != 0 {
		t.Fatalf("valid active_standby should pass, got %v", errs)
	}
}

func TestLeaderElectionActiveMissingQueueAndSession(t *testing.T) {
	d := defsWithStores()
	d.LeaderElection = spec.LeaderElection{Present: true, Mode: spec.LeaderActiveActive}
	errs, _ := Run(Context{Workflows: leWorkflows(), Defaults: d})
	if !hasErr(errs, "requires a 'queue'") || !hasErr(errs, "requires a solace session") {
		t.Fatalf("want queue + session errors, got %v", errs)
	}
}

func TestLeaderElectionConnRefMustBeSolace(t *testing.T) {
	d := defsWithStores()
	d.Connections = map[string]spec.Side{"qm": {System: spec.SystemMQ, ConnName: "h(1414)", QueueManager: "QM", Channel: "C", User: "u", Password: "p"}}
	d.LeaderElection = spec.LeaderElection{Present: true, Mode: spec.LeaderActiveStby, Queue: "q", ConnRef: "qm"}
	if errs, _ := Run(Context{Workflows: leWorkflows(), Defaults: d}); !hasErr(errs, "must be a solace connection") {
		t.Fatalf("want solace-session error, got %v", errs)
	}
}

func TestLeaderElectionInvalidMode(t *testing.T) {
	d := defsWithStores()
	d.LeaderElection = spec.LeaderElection{Present: true, Mode: "bogus"}
	if errs, _ := Run(Context{Workflows: leWorkflows(), Defaults: d}); !hasErr(errs, "is invalid") {
		t.Fatalf("want invalid-mode error, got %v", errs)
	}
}

// mgmtSide is a well-formed inline management session: a solace tuple with no
// destination, which is all a session may carry. Its broker is deliberately not
// leWorkflows' broker -- a session that shares a workflow's tuple shares its
// mounted credential too, which is TestLeaderSessionPasswordConflict's subject,
// not these tests'.
func mgmtSide() spec.Side {
	return spec.Side{System: spec.SystemSolace, Host: "tcps://mgmt:55443", MsgVPN: "mgmt", ClientUser: "u", ClientPassEnv: "MGMT_CLIENT_PASSWORD"}
}

// TestLeaderElectionSolaceKeyRenamed pins the retired spelling. It rides in
// checkRemovedDefaultsKeys rather than checkLeaderElection so it is reported
// whatever the mode is -- including standalone, where checkLeaderElection
// returns before it could look.
func TestLeaderElectionSolaceKeyRenamed(t *testing.T) {
	for _, c := range []struct {
		name string
		mode string
	}{
		{"active_standby", spec.LeaderActiveStby},
		{"standalone", spec.LeaderStandalone},
	} {
		t.Run(c.name, func(t *testing.T) {
			d := defsWithStores()
			sess := mgmtSide()
			d.LeaderElection = spec.LeaderElection{Present: true, Mode: c.mode, Queue: "mgmt-q", Session: &sess, SolaceKey: true}
			errs, _ := Run(Context{Workflows: leWorkflows(), Defaults: d})
			if !hasErr(errs, "leader-election.solace has been renamed to leader-election.session") {
				t.Fatalf("want the rename error, got %v", errs)
			}
		})
	}
}

// TestLeaderElectionConnRefAndInlineSession rejects the combination that used to
// resolve silently: consolidate prefers conn-ref, so the inline block is dead
// config -- yet credential checking still ran over it, reporting errors about a
// session nothing renders.
func TestLeaderElectionConnRefAndInlineSession(t *testing.T) {
	d := defsWithStores()
	sess := mgmtSide()
	d.Connections = map[string]spec.Side{"mgmt": mgmtSide()}
	d.LeaderElection = spec.LeaderElection{Present: true, Mode: spec.LeaderActiveStby, Queue: "mgmt-q", ConnRef: "mgmt", Session: &sess}
	if errs, _ := Run(Context{Workflows: leWorkflows(), Defaults: d}); !hasErr(errs, "sets both conn-ref") {
		t.Fatalf("want the conn-ref/session conflict error, got %v", errs)
	}
}

// TestLeaderElectionSessionRejectsBindingFields covers the four keys a session
// parses and then discards. leader-election.session.queue in particular is the
// near miss for leader-election.queue, so silence there costs an operator the
// management queue itself.
func TestLeaderElectionSessionRejectsBindingFields(t *testing.T) {
	for _, c := range []struct {
		name  string
		apply func(*spec.Side)
	}{
		{"queue", func(s *spec.Side) { s.DestKind, s.Dest = spec.DestQueue, "oops" }},
		{"topic", func(s *spec.Side) { s.DestKind, s.Dest = spec.DestTopic, "oops" }},
		{"consumer", func(s *spec.Side) { s.Consumer = &yaml.Node{Kind: yaml.MappingNode} }},
		{"producer", func(s *spec.Side) { s.Producer = &yaml.Node{Kind: yaml.MappingNode} }},
	} {
		t.Run(c.name, func(t *testing.T) {
			d := defsWithStores()
			sess := mgmtSide()
			c.apply(&sess)
			d.LeaderElection = spec.LeaderElection{Present: true, Mode: spec.LeaderActiveStby, Queue: "mgmt-q", Session: &sess}
			errs, _ := Run(Context{Workflows: leWorkflows(), Defaults: d})
			if !hasErr(errs, "leader-election.session may not set queue/topic/consumer/producer") {
				t.Fatalf("want the binding-field error, got %v", errs)
			}
		})
	}
}

// TestLeaderElectionInlineSessionValid is the negative control for the two tests
// above: a bare solace tuple under session: is exactly what a management session
// is, and must pass clean.
func TestLeaderElectionInlineSessionValid(t *testing.T) {
	d := defsWithStores()
	sess := mgmtSide()
	d.LeaderElection = spec.LeaderElection{Present: true, Mode: spec.LeaderActiveStby, Queue: "mgmt-q", Session: &sess}
	if errs, _ := Run(Context{Workflows: leWorkflows(), Defaults: d}); len(errs) != 0 {
		t.Fatalf("a bare inline session should pass, got %v", errs)
	}
}

// TestLeaderSessionPasswordConflict guards the hazard the shared credential name
// creates. The dedup tuple omits the password, so a session that reaches the same
// broker as a workflow binder now mounts ONE credential -- the binder's, recorded
// first -- and a session that meant to use a different password would silently
// authenticate with the wrong one.
func TestLeaderSessionPasswordConflict(t *testing.T) {
	// vSolace's tuple: tcps://b:55443 / prod / u, password-env SOLACE_CLIENT_PASSWORD.
	same := func() spec.Side {
		return spec.Side{System: spec.SystemSolace, Host: "tcps://b:55443", MsgVPN: "prod", ClientUser: "u", ClientPassEnv: "SOLACE_CLIENT_PASSWORD"}
	}
	differ := func() spec.Side {
		s := same()
		s.ClientPassEnv = "OTHER_PASSWORD"
		return s
	}
	elsewhere := func() spec.Side {
		s := differ()
		s.Host = "tcps://other:55443"
		return s
	}
	for _, c := range []struct {
		name    string
		inline  func() spec.Side
		connRef bool
		want    bool
	}{
		{"inline session disagrees", differ, false, true},
		{"conn-ref session disagrees", differ, true, true},
		{"session agrees", same, false, false},
		{"session is a different broker", elsewhere, false, false},
	} {
		t.Run(c.name, func(t *testing.T) {
			d := defsWithStores()
			sess := c.inline()
			d.LeaderElection = spec.LeaderElection{Present: true, Mode: spec.LeaderActiveStby, Queue: "mgmt-q"}
			if c.connRef {
				d.Connections = map[string]spec.Side{"mgmt": sess}
				d.LeaderElection.ConnRef = "mgmt"
			} else {
				d.LeaderElection.Session = &sess
			}
			errs, _ := Run(Context{Workflows: leWorkflows(), Defaults: d})
			if got := hasErr(errs, "reaches the same broker tuple with a different password"); got != c.want {
				t.Fatalf("conflict reported = %v, want %v (errs %v)", got, c.want, errs)
			}
		})
	}
}

func TestMQCipherRequiresTLS(t *testing.T) {
	m := vMQ("MQ", spec.DestQueue, false)
	m.Cipher = "TLS_RSA_WITH_AES_256_CBC_SHA256"
	errs, _ := Run(Context{Workflows: []spec.Workflow{wf("x.yaml", vSolace("Q", spec.DestQueue, ""), m)}, Defaults: &spec.Defaults{}})
	if !hasErr(errs, "require 'tls: true'") {
		t.Fatalf("want cipher-requires-tls, got %v", errs)
	}
}

func TestDeployKubeChecks(t *testing.T) {
	k := &spec.Kubernetes{Deployment: spec.Deployment{Name: "Bad_Name", Namespace: "ns"}}
	errs, _ := Run(Context{
		Workflows: []spec.Workflow{wf("x.yaml", vSolace("Q", spec.DestQueue, ""), vMQ("MQ", spec.DestQueue, false))},
		Defaults:  &spec.Defaults{}, Image: imageOK(), Kube: k, CheckKubernetes: true,
	})
	if !hasErr(errs, "not a valid DNS-1123 label") {
		t.Fatalf("want DNS-1123 error, got %v", errs)
	}
}

// TestCredentialsEnvChecks covers checkDefaultsCredentials: a store password-env
// that names a variable absent from the environment is a WARNING, not an error:
// authoring a spec or generating a config must not require every production
// credential to be exported. The deploy path resolves values and fails hard there.
func TestCredentialsEnvChecks(t *testing.T) {
	d := &spec.Defaults{TLS: spec.TLSConfig{Stores: map[string]*spec.Store{
		"ts": {Existing: &spec.StoreExisting{PKCS12: "t.jks"}, PasswordEnv: "TRUSTSTORE_PASSWORD"},
	}}}
	env := func(string) (string, bool) { return "", false } // nothing set
	errs, warns := Run(Context{
		Workflows: []spec.Workflow{wf("x.yaml", vSolace("Q", spec.DestQueue, ""), vMQ("MQ", spec.DestQueue, false))},
		Defaults:  d, Env: env,
	})
	if !hasErr(warns, `tls.stores.ts.password-env names TRUSTSTORE_PASSWORD, which is not set in this environment`) {
		t.Fatalf("want missing-env-var warning, got %v", warns)
	}
	if hasErr(errs, "TRUSTSTORE_PASSWORD") {
		t.Fatalf("an unset -env variable must not be fatal at validate time, got %v", errs)
	}
}

// TestCheckCredRules pins the checkCred rules for one credential position: the
// literal/-env forms are mutually exclusive, an -env value must be a bare
// variable name (not a ${...} reference), it must be a valid identifier, and --
// when Context.Env is supplied -- it must actually be set.
func TestCheckCredRules(t *testing.T) {
	side := func(mutate func(*spec.Side)) spec.Side {
		s := vMQ("MQ", spec.DestQueue, false)
		mutate(&s)
		return s
	}
	cases := []struct {
		name string
		s    spec.Side
		env  func(string) (string, bool)
		want string
		// asWarning marks the one case that is advisory rather than fatal: an
		// -env variable that is simply not exported on this machine.
		asWarning bool
	}{
		{
			name: "both literal and env set",
			s:    side(func(s *spec.Side) { s.Password = "secret" }), // PasswordEnv already set by vMQ
			want: "sets both a literal value and target password-env",
		},
		{
			name: "env value is a ${...} reference, not a bare name",
			s:    side(func(s *spec.Side) { s.PasswordEnv = "${MQ_PASSWORD}" }),
			want: "must be a bare variable name, not a ${...} reference",
		},
		{
			name: "env value is not a valid identifier",
			s:    side(func(s *spec.Side) { s.PasswordEnv = "1-bad-name" }),
			want: "is not a valid environment variable name",
		},
		{
			name:      "env var unset when Env is supplied",
			s:         side(func(s *spec.Side) {}),
			env:       func(string) (string, bool) { return "", false },
			want:      "which is not set in this environment",
			asWarning: true,
		},
	}
	for _, c := range cases {
		errs, warns := Run(Context{
			Workflows: []spec.Workflow{wf("x.yaml", vSolace("Q", spec.DestQueue, ""), c.s)},
			Defaults:  &spec.Defaults{}, Env: c.env,
		})
		got := errs
		if c.asWarning {
			got = warns
		}
		if !hasErr(got, c.want) {
			t.Errorf("%s: want %q, got errs=%v warns=%v", c.name, c.want, errs, warns)
		}
	}
}

// TestCheckCredLiteralLooksLikeEnvRefWarns covers the last checkCred branch: a
// literal containing "${" is almost always someone reaching for the -env key, so
// it warns (not errors) naming the -env key to use instead.
func TestCheckCredLiteralLooksLikeEnvRefWarns(t *testing.T) {
	m := vMQ("MQ", spec.DestQueue, false)
	m.PasswordEnv = ""
	m.Password = "${SOME_VAR}"
	_, warns := Run(Context{
		Workflows: []spec.Workflow{wf("x.yaml", vSolace("Q", spec.DestQueue, ""), m)},
		Defaults:  &spec.Defaults{},
	})
	if !hasErr(warns, "target password looks like a variable reference; it is used as a literal value. Use target password-env: VAR") {
		t.Fatalf("want literal-looks-like-env-ref warning, got %v", warns)
	}
}

// certsOf is a create: block listing the given certificate files.
func certsOf(files ...string) *spec.StoreSources {
	c := &spec.StoreSources{}
	for _, f := range files {
		c.Certs.Items = append(c.Certs.Items, spec.CertSource{File: f})
	}
	return c
}

// TestStoreDefinitions pins checkStoreDefinition: a store is created or
// existing, exactly one; an existing one names its PKCS12; a created one has
// its own password and sources the PKCS12 writer can build one store from --
// trusted certificates or one key entry, never both -- with every path and
// alias filled in. Every reader wires a store only when it is configured, so a
// store refused here would otherwise be skipped in silence.
func TestStoreDefinitions(t *testing.T) {
	key := spec.KeySource{Cert: "./c.pem", Key: "./c.key"}
	for _, c := range []struct {
		name string
		st   *spec.Store
		want string
	}{
		{"no block", nil, "env.yaml: tls.stores.s needs create: (the tool builds it) or existing: (a PKCS12 you provide, or a kubernetes Secret holding one)"},
		{"no source", &spec.Store{PasswordEnv: "PW"}, "tls.stores.s needs create: (the tool builds it) or existing: (a PKCS12 you provide, or a kubernetes Secret holding one)"},
		{"both", &spec.Store{PasswordEnv: "PW", Create: certsOf("./a.pem"), Existing: &spec.StoreExisting{PKCS12: "./t.p12"}}, "tls.stores.s takes create: or existing.pkcs12, not both: one host-side source per store"},
		{"existing naming nothing", &spec.Store{PasswordEnv: "PW", Existing: &spec.StoreExisting{}}, "tls.stores.s.existing names neither pkcs12 nor secret"},
		{"a secret without a name", &spec.Store{PasswordEnv: "PW", Existing: &spec.StoreExisting{Secret: &spec.StoreSecret{}}}, "tls.stores.s.existing.secret.name is required"},
		{"a secret name no Secret can have", &spec.Store{PasswordEnv: "PW", Existing: &spec.StoreExisting{Secret: &spec.StoreSecret{Name: "My_TLS"}}}, `tls.stores.s.existing.secret.name "My_TLS" is not a valid DNS-1123 label`},
		{"a key no Secret can hold", &spec.Store{PasswordEnv: "PW", Existing: &spec.StoreExisting{Secret: &spec.StoreSecret{Name: "my-tls", Key: "a/b"}}}, `tls.stores.s.existing.secret.key "a/b" is not a valid Secret data key`},
		{"created without a password", &spec.Store{Create: certsOf("./a.pem")}, "tls.stores.s is created and encrypted with its own password, but sets none; add password-env"},
		{"created from nothing", &spec.Store{PasswordEnv: "PW", Create: &spec.StoreSources{}}, "tls.stores.s.create names no source: list certs, give one keys entry, or name a jks"},
		{"certs and a key", &spec.Store{PasswordEnv: "PW", Create: &spec.StoreSources{Certs: certsOf("./a.pem").Certs, Keys: []spec.KeySource{key}}}, "tls.stores.s.create mixes trusted certificates and a key: a store is a truststore or a keystore -- define a second store for the other"},
		{"a jks and a key", &spec.Store{PasswordEnv: "PW", Create: &spec.StoreSources{JKS: "./old.jks", Keys: []spec.KeySource{key}}}, "tls.stores.s.create mixes trusted certificates and a key"},
		{"two keys", &spec.Store{PasswordEnv: "PW", Create: &spec.StoreSources{Keys: []spec.KeySource{key, key}}}, "tls.stores.s.create.keys has 2 entries: a created keystore holds exactly one key; define one store per identity"},
		{"a key entry without its key", &spec.Store{PasswordEnv: "PW", Create: &spec.StoreSources{Keys: []spec.KeySource{{Cert: "./c.pem"}}}}, "tls.stores.s.create.keys[0] needs cert and key: the certificate and the private key it belongs to"},
		{"an empty chain file", &spec.Store{PasswordEnv: "PW", Create: &spec.StoreSources{Keys: []spec.KeySource{{Cert: "./c.pem", Chain: []string{""}, Key: "./c.key"}}}}, "tls.stores.s.create.keys[0].chain[0] is empty: name a certificate file"},
		{"an empty certs entry", &spec.Store{PasswordEnv: "PW", Create: certsOf("")}, "tls.stores.s.create.certs[0] is empty: name a certificate file"},
		{"an empty named entry", &spec.Store{PasswordEnv: "PW", Create: &spec.StoreSources{Certs: spec.CertSources{Named: true, Items: []spec.CertSource{{Name: "root"}}}}}, "tls.stores.s.create.certs.root is empty: name a certificate file"},
		{"an empty alias", &spec.Store{PasswordEnv: "PW", Create: &spec.StoreSources{Certs: spec.CertSources{Named: true, Items: []spec.CertSource{{File: "./a.pem"}}}}}, "tls.stores.s.create.certs has an empty alias: name every certificate"},
		{"an alias twice", &spec.Store{PasswordEnv: "PW", Create: &spec.StoreSources{Certs: spec.CertSources{Named: true, Items: []spec.CertSource{{Name: "Root", File: "./a.pem"}, {Name: "root", File: "./b.pem"}}}}}, `tls.stores.s.create.certs: alias "root" appears twice (aliases ignore case)`},
	} {
		t.Run(c.name, func(t *testing.T) {
			d := &spec.Defaults{TLS: spec.TLSConfig{Stores: map[string]*spec.Store{"s": c.st}}}
			errs, _ := Run(Context{Workflows: []spec.Workflow{wf("10.yaml", vSolace("Q", spec.DestQueue, ""), vMQ("MQ", spec.DestQueue, false))}, Defaults: d})
			if !hasErr(errs, c.want) {
				t.Errorf("want %q, got %v", c.want, errs)
			}
		})
	}
	// Every well-formed shape is accepted: existing, a certs list, a named map,
	// one key entry with its chain, a jks alone, certs with a jks.
	ok := &spec.Defaults{TLS: spec.TLSConfig{Stores: map[string]*spec.Store{
		"have":  {PasswordEnv: "PW", Existing: &spec.StoreExisting{PKCS12: "./t.p12"}},
		"list":  {PasswordEnv: "PW", Create: certsOf("./a.pem", "./b.pem")},
		"named": {PasswordEnv: "PW", Create: &spec.StoreSources{Certs: spec.CertSources{Named: true, Items: []spec.CertSource{{Name: "root", File: "./a.pem"}}}}},
		"key":   {Password: "pw", Create: &spec.StoreSources{Keys: []spec.KeySource{{Cert: "./c.pem", Chain: []string{"./i.pem"}, Key: "./c.key"}}}},
		"jks":   {PasswordEnv: "PW", Create: &spec.StoreSources{JKS: "./old.jks"}},
		"both":  {PasswordEnv: "PW", Create: &spec.StoreSources{Certs: certsOf("./a.pem").Certs, JKS: "./old.jks"}},
		// A kubernetes Secret, alone, beside a file, or beside sources.
		"secret":         {PasswordEnv: "PW", Existing: &spec.StoreExisting{Secret: &spec.StoreSecret{Name: "my-tls", Key: "keystore.p12"}}},
		"file-secret":    {PasswordEnv: "PW", Existing: &spec.StoreExisting{PKCS12: "./t.p12", Secret: &spec.StoreSecret{Name: "my-tls"}}},
		"sources-secret": {PasswordEnv: "PW", Create: certsOf("./a.pem"), Existing: &spec.StoreExisting{Secret: &spec.StoreSecret{Name: "my-tls"}}},
	}}}
	if errs, _ := Run(Context{Workflows: []spec.Workflow{wf("10.yaml", vSolace("Q", spec.DestQueue, ""), vMQ("MQ", spec.DestQueue, false))}, Defaults: ok}); hasErr(errs, "tls.stores.") {
		t.Errorf("well-formed stores refused: %v", errs)
	}
}

// TestCreatedStoreSelections pins checkCreatedSelection: a store created from
// a key holds no trusted certificates, one created from certificates holds no
// key, and a created keystore's one key takes no key-alias -- on every
// selection site, here an inline workflow side.
func TestCreatedStoreSelections(t *testing.T) {
	d := &spec.Defaults{TLS: spec.TLSConfig{Stores: map[string]*spec.Store{
		"made-ca":  {PasswordEnv: "PW", Create: certsOf("./a.pem")},
		"made-key": {PasswordEnv: "PW", Create: &spec.StoreSources{Keys: []spec.KeySource{{Cert: "./c.pem", Key: "./c.key"}}}},
		"made-jks": {PasswordEnv: "PW", Create: &spec.StoreSources{JKS: "./old.jks"}},
	}}}
	sel := func(ts, ks, alias string) spec.Side {
		s := vSolace("Q", spec.DestQueue, alias)
		s.Truststore, s.Keystore = ts, ks
		return s
	}
	for _, c := range []struct {
		name string
		side spec.Side
		want string
	}{
		{"a key store as truststore", sel("made-key", "", ""), `source solace: truststore "made-key" is created from a key, so it holds no trusted certificates; create a second store from the CA certificates and select that`},
		{"a certificate store as keystore", sel("", "made-ca", ""), `source solace: keystore "made-ca" is created from certificates only, so it holds no private key; give it one keys entry`},
		{"a key-alias on a created keystore", sel("", "made-key", "client"), `source solace: remove key-alias "client": tls.stores.made-key is created with exactly one key, which needs no alias (Java names it "1")`},
	} {
		t.Run(c.name, func(t *testing.T) {
			errs, _ := Run(Context{Workflows: []spec.Workflow{wf("10.yaml", c.side, vMQ("MQ", spec.DestQueue, false))}, Defaults: d})
			if !hasErr(errs, c.want) {
				t.Errorf("want %q, got %v", c.want, errs)
			}
		})
	}
	// Roles that fit pass, and a jks-only store's role is left to deploy.
	for _, side := range []spec.Side{sel("made-ca", "made-key", ""), sel("made-jks", "made-jks", "")} {
		if errs, _ := Run(Context{Workflows: []spec.Workflow{wf("10.yaml", side, vMQ("MQ", spec.DestQueue, false))}, Defaults: d}); hasErr(errs, "is created") {
			t.Errorf("%s/%s refused: %v", side.Truststore, side.Keystore, errs)
		}
	}
}

// TestUnusedStoreWarns pins checkUnusedStores: a store nothing selects is not
// mounted, so it is reported -- usually a selector typo -- while a store any
// connection, side or session selects is not.
func TestUnusedStoreWarns(t *testing.T) {
	d := defsWithStores()
	d.Connections = map[string]spec.Side{"edge": {System: spec.SystemSolace, Host: "tcps://b:55443", MsgVPN: "v", Keystore: "ks"}}
	wfs := []spec.Workflow{wf("10.yaml", vSolace("Q", spec.DestQueue, ""), vMQ("MQ", spec.DestQueue, false))}
	_, warns := Run(Context{Workflows: wfs, Defaults: d})
	want := `env.yaml: tls.stores "ts" is not selected by any connection, so it is not mounted; remove it, or select it with truststore: or keystore:`
	if !hasErr(warns, want) {
		t.Errorf("want %q, got %v", want, warns)
	}
	if hasErr(warns, `tls.stores "ks" is not selected`) {
		t.Errorf("ks is selected by connections.edge: %v", warns)
	}
}
