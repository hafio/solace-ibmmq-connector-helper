package validate

import (
	"fmt"
	"strings"
	"testing"

	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/spec"
)

func vSolace(dest, kind, keyAlias string) spec.Side {
	return spec.Side{System: spec.SystemSolace, Host: "tcps://b:55443", MsgVPN: "prod",
		ClientUser: "u", ClientPass: "${P}", DestKind: kind, Dest: dest, KeyAlias: keyAlias}
}
func vMQ(dest, kind string, tls bool) spec.Side {
	return spec.Side{System: spec.SystemMQ, ConnName: "h(1414)", QueueManager: "QM1", Channel: "CH",
		User: "u", Password: "${P}", TLS: tls, DestKind: kind, Dest: dest}
}
func wf(file string, src, tgt spec.Side) spec.Workflow {
	return spec.Workflow{File: file, Enabled: true, SourceSet: true, TargetSet: true, Source: src, Target: tgt}
}
func defsWithStores() *spec.Defaults {
	return &spec.Defaults{TLS: spec.TLSConfig{
		Truststore: &spec.Store{File: "t.jks", Password: "${T}", Type: "JKS"},
		Keystore:   &spec.Store{File: "k.jks", Password: "${K}", Type: "JKS"},
	}}
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

func TestNoWorkflowCap(t *testing.T) {
	// >MaxWorkflowsPerInstance workflows no longer fail: the gen layer shards them.
	errs, _ := Run(Context{Workflows: manyWorkflows(MaxWorkflowsPerInstance + 1), Defaults: &spec.Defaults{}})
	if hasErr(errs, "too many workflows") {
		t.Fatalf("workflow cap should be gone, got %v", errs)
	}
}

func TestDeployNameTooLongWhenSharded(t *testing.T) {
	// 21 workflows -> 2 instances -> longest generated name is "<name>-2-config".
	longName := strings.Repeat("a", 60) // 60 + len("-2-config")=9 -> 69 > 63
	k := &spec.Kubernetes{Deployment: spec.Deployment{Name: longName, Namespace: "ns", Image: "img", Replicas: 1}}
	errs, _ := Run(Context{Workflows: manyWorkflows(MaxWorkflowsPerInstance + 1), Defaults: &spec.Defaults{}, Kube: k, Deploy: true})
	if !hasErr(errs, "exceeds the 63-char DNS-1123 limit") {
		t.Fatalf("want too-long suffixed name error, got %v", errs)
	}
	// A short name with the same workflow count is fine.
	k.Deployment.Name = "solmq"
	errs2, _ := Run(Context{Workflows: manyWorkflows(MaxWorkflowsPerInstance + 1), Defaults: &spec.Defaults{}, Kube: k, Deploy: true})
	if hasErr(errs2, "exceeds the 63-char DNS-1123 limit") {
		t.Fatalf("short name should not trip the length guard, got %v", errs2)
	}
}

func leWorkflows() []spec.Workflow {
	return []spec.Workflow{wf("x.yaml", vSolace("Q", spec.DestQueue, ""), vMQ("MQ", spec.DestQueue, false))}
}

func TestLeaderElectionActiveStandbyValid(t *testing.T) {
	d := defsWithStores()
	d.Connections = map[string]spec.Side{"mgmt": {System: spec.SystemSolace, Host: "tcps://b:55443", MsgVPN: "prod", ClientUser: "u", ClientPass: "${P}"}}
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
	k := &spec.Kubernetes{Deployment: spec.Deployment{Name: "Bad_Name", Namespace: "ns", Image: ""}}
	errs, _ := Run(Context{
		Workflows: []spec.Workflow{wf("x.yaml", vSolace("Q", spec.DestQueue, ""), vMQ("MQ", spec.DestQueue, false))},
		Defaults:  &spec.Defaults{}, Kube: k, Deploy: true,
	})
	if !hasErr(errs, "not a valid DNS-1123 label") {
		t.Fatalf("want DNS-1123 error, got %v", errs)
	}
	if !hasErr(errs, "deployment.image is required") {
		t.Fatalf("want image-required error, got %v", errs)
	}
}

func TestCredentialsEnvChecks(t *testing.T) {
	k := &spec.Kubernetes{
		Deployment: spec.Deployment{Name: "c", Namespace: "ns", Image: "img", Replicas: 1},
		Secrets: spec.Secrets{Credentials: &spec.CredentialsSecret{
			Create: &spec.CredCreate{Name: "s", Source: spec.SourceEnv, Variables: []string{"MISSING_VAR"}},
		}},
	}
	env := func(string) (string, bool) { return "", false } // nothing set
	errs, _ := Run(Context{
		Workflows: []spec.Workflow{wf("x.yaml", vSolace("Q", spec.DestQueue, ""), vMQ("MQ", spec.DestQueue, false))},
		Defaults:  &spec.Defaults{}, Kube: k, Deploy: true, Env: env,
	})
	if !hasErr(errs, `variable "MISSING_VAR" is not set`) {
		t.Fatalf("want missing-env-var error, got %v", errs)
	}
}
