package validate

import (
	"fmt"
	"strings"
	"testing"

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
func defsWithStores() *spec.Defaults {
	return &spec.Defaults{TLS: spec.TLSConfig{
		Truststore: &spec.Store{File: "t.jks", PasswordEnv: "TRUSTSTORE_PASSWORD", Type: "JKS"},
		Keystore:   &spec.Store{File: "k.jks", PasswordEnv: "KEYSTORE_PASSWORD", Type: "JKS"},
	}}
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
	wfs := []spec.Workflow{wf("10.yaml", vSolace("Q", spec.DestQueue, "sc"), vMQ("MQ", spec.DestQueue, true))}
	errs, _ := Run(Context{Workflows: wfs, Defaults: defsWithStores()})
	if len(errs) != 0 {
		t.Fatalf("expected no errors, got %v", errs)
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

func TestSolaceTopicSourceWarnsNotErrors(t *testing.T) {
	// A Solace topic source is now allowed, but flagged with an EDA advisory.
	errs, warns := Run(Context{Workflows: []spec.Workflow{wf("x.yaml", vSolace("t", spec.DestTopic, ""), vMQ("MQ", spec.DestQueue, false))}, Defaults: &spec.Defaults{}})
	if len(errs) != 0 {
		t.Fatalf("Solace topic source should be allowed now, got errors %v", errs)
	}
	if !hasErr(warns, "non-durable subscription") {
		t.Fatalf("want EDA warning for a Solace topic source, got %v", warns)
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
	errs, _ := Run(Context{Workflows: []spec.Workflow{wf("x.yaml", vSolace("Q", spec.DestQueue, "sc"), vMQ("MQ", spec.DestQueue, false))}, Defaults: &spec.Defaults{}})
	if !hasErr(errs, "no keystore defined") {
		t.Fatalf("want keystore error, got %v", errs)
	}
}

func TestKeyAliasConflict(t *testing.T) {
	a := vSolace("Q1", spec.DestQueue, "alias-a")
	b := vSolace("Q2", spec.DestQueue, "alias-b") // same solace tuple, different alias
	wfs := []spec.Workflow{
		wf("10.yaml", a, vMQ("M1", spec.DestQueue, false)),
		wf("20.yaml", vMQ("M2", spec.DestQueue, false), b),
	}
	errs, _ := Run(Context{Workflows: wfs, Defaults: defsWithStores()})
	if !hasErr(errs, "conflicting key-alias") {
		t.Fatalf("want key-alias conflict, got %v", errs)
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
	d.Connections = map[string]spec.Side{"mgmt": {System: spec.SystemSolace, Host: "tcps://b:55443", MsgVPN: "prod", ClientUser: "u", ClientPassEnv: "MGMT_CLIENT_PASSWORD"}}
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
	d := &spec.Defaults{TLS: spec.TLSConfig{
		Truststore: &spec.Store{File: "t.jks", PasswordEnv: "TRUSTSTORE_PASSWORD"},
	}}
	env := func(string) (string, bool) { return "", false } // nothing set
	errs, warns := Run(Context{
		Workflows: []spec.Workflow{wf("x.yaml", vSolace("Q", spec.DestQueue, ""), vMQ("MQ", spec.DestQueue, false))},
		Defaults:  d, Env: env,
	})
	if !hasErr(warns, `tls.truststore.password-env names TRUSTSTORE_PASSWORD, which is not set in this environment`) {
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
