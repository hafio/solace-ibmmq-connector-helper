package validate

import (
	"testing"

	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-gen/internal/spec"
)

func wfOK() []spec.Workflow {
	return []spec.Workflow{wf("x.yaml", vSolace("Q", spec.DestQueue, ""), vMQ("M", spec.DestQueue, false))}
}

func TestCheckSideMQMissingFields(t *testing.T) {
	m := spec.Side{System: spec.SystemMQ, DestKind: spec.DestQueue, Dest: "Q"}
	errs, _ := Run(Context{Workflows: []spec.Workflow{wf("x.yaml", vSolace("Q", spec.DestQueue, ""), m)}, Defaults: &spec.Defaults{}})
	for _, sub := range []string{"missing 'conn-name'", "missing 'queue-manager'", "missing 'channel'"} {
		if !hasErr(errs, sub) {
			t.Errorf("want %q, got %v", sub, errs)
		}
	}
	// user/password are optional (cert-based or channel auth) — never flagged missing.
	for _, sub := range []string{"missing 'user'", "missing 'password'"} {
		if hasErr(errs, sub) {
			t.Errorf("user/password are optional; did not want %q, got %v", sub, errs)
		}
	}
}

func TestCheckSideSolaceMissingAndBadScheme(t *testing.T) {
	empty := spec.Side{System: spec.SystemSolace, DestKind: spec.DestQueue, Dest: "Q"}
	errs, _ := Run(Context{Workflows: []spec.Workflow{wf("x.yaml", empty, vMQ("M", spec.DestQueue, false))}, Defaults: &spec.Defaults{}})
	for _, sub := range []string{"missing 'host'", "missing 'msg-vpn'"} {
		if !hasErr(errs, sub) {
			t.Errorf("want %q, got %v", sub, errs)
		}
	}
	// client-username/client-password are optional (cert-only/OAuth auth) — never flagged missing.
	for _, sub := range []string{"missing 'client-username'", "missing 'client-password'"} {
		if hasErr(errs, sub) {
			t.Errorf("client credentials are optional; did not want %q, got %v", sub, errs)
		}
	}
	bad := spec.Side{System: spec.SystemSolace, Host: "http://x", MsgVPN: "v", ClientUser: "u", ClientPass: "p", DestKind: spec.DestQueue, Dest: "Q"}
	errs2, _ := Run(Context{Workflows: []spec.Workflow{wf("x.yaml", bad, vMQ("M", spec.DestQueue, false))}, Defaults: &spec.Defaults{}})
	if !hasErr(errs2, "must start with tcp:// or tcps://") {
		t.Errorf("want scheme error, got %v", errs2)
	}
}

func TestSolaceKeyAliasRequiresTCPSAndKeystore(t *testing.T) {
	s := spec.Side{System: spec.SystemSolace, Host: "tcp://x", MsgVPN: "v", ClientUser: "u", ClientPass: "p", KeyAlias: "a", DestKind: spec.DestQueue, Dest: "Q"}
	errs, _ := Run(Context{Workflows: []spec.Workflow{wf("x.yaml", s, vMQ("M", spec.DestQueue, false))}, Defaults: defsWithStores()})
	if !hasErr(errs, "requires a tcps:// host") {
		t.Errorf("want tcps requirement, got %v", errs)
	}
	s2 := spec.Side{System: spec.SystemSolace, Host: "tcps://x", MsgVPN: "v", ClientUser: "u", ClientPass: "p", KeyAlias: "a", DestKind: spec.DestQueue, Dest: "Q"}
	errs2, _ := Run(Context{Workflows: []spec.Workflow{wf("x.yaml", s2, vMQ("M", spec.DestQueue, false))}, Defaults: &spec.Defaults{}})
	if !hasErr(errs2, "no keystore defined") {
		t.Errorf("want keystore requirement, got %v", errs2)
	}
}

func TestMQKeyAliasRequiresKeystore(t *testing.T) {
	m := vMQ("M", spec.DestQueue, true)
	m.KeyAlias = "mc"
	errs, _ := Run(Context{Workflows: []spec.Workflow{wf("x.yaml", vSolace("Q", spec.DestQueue, ""), m)}, Defaults: &spec.Defaults{}})
	if !hasErr(errs, "no keystore defined") {
		t.Errorf("want keystore error, got %v", errs)
	}
}

func TestCheckKubeRequiredAndReplicas(t *testing.T) {
	k := &spec.Kubernetes{Deployment: spec.Deployment{Replicas: 3}}
	errs, _ := Run(Context{Workflows: wfOK(), Defaults: &spec.Defaults{}, Kube: k, Deploy: true})
	for _, sub := range []string{"deployment.name is required", "deployment.namespace is required", "deployment.image is required", "replicas: 1"} {
		if !hasErr(errs, sub) {
			t.Errorf("want %q, got %v", sub, errs)
		}
	}
}

func TestCheckKubeCredentialSources(t *testing.T) {
	base := spec.Deployment{Name: "c", Namespace: "ns", Image: "img", Replicas: 1}
	k1 := &spec.Kubernetes{Deployment: base, Secrets: spec.Secrets{Credentials: &spec.CredentialsSecret{Create: &spec.CredCreate{Name: "s", Source: spec.SourceEnv}}}}
	if e, _ := Run(Context{Workflows: wfOK(), Defaults: &spec.Defaults{}, Kube: k1, Deploy: true}); !hasErr(e, "non-empty 'variables'") {
		t.Errorf("want empty-variables, got %v", e)
	}
	k2 := &spec.Kubernetes{Deployment: base, Secrets: spec.Secrets{Credentials: &spec.CredentialsSecret{Create: &spec.CredCreate{Name: "s", Source: spec.SourceFile}}}}
	if e, _ := Run(Context{Workflows: wfOK(), Defaults: &spec.Defaults{}, Kube: k2, Deploy: true}); !hasErr(e, "requires 'values-file'") {
		t.Errorf("want values-file, got %v", e)
	}
	k3 := &spec.Kubernetes{Deployment: base, Secrets: spec.Secrets{Credentials: &spec.CredentialsSecret{Create: &spec.CredCreate{Name: "s", Source: "weird"}}}}
	if e, _ := Run(Context{Workflows: wfOK(), Defaults: &spec.Defaults{}, Kube: k3, Deploy: true}); !hasErr(e, "source must be") {
		t.Errorf("want bad-source, got %v", e)
	}
}

func TestCheckKubeStoresRequireTruststore(t *testing.T) {
	k := &spec.Kubernetes{Deployment: spec.Deployment{Name: "c", Namespace: "ns", Image: "img", Replicas: 1}, Secrets: spec.Secrets{Stores: &spec.StoresSecret{Create: &spec.StoreCreate{Name: "t"}}}}
	if errs, _ := Run(Context{Workflows: wfOK(), Defaults: &spec.Defaults{}, Kube: k, Deploy: true}); !hasErr(errs, "requires defaults.yaml tls.truststore") {
		t.Errorf("want truststore requirement, got %v", errs)
	}
}

func TestStoresNotWiredWarning(t *testing.T) {
	k := &spec.Kubernetes{Deployment: spec.Deployment{Name: "c", Namespace: "ns", Image: "img", Replicas: 1}}
	_, warns := Run(Context{Workflows: []spec.Workflow{wf("x.yaml", vSolace("Q", spec.DestQueue, ""), vMQ("M", spec.DestQueue, true))}, Defaults: defsWithStores(), Kube: k, Deploy: true})
	if !hasErr(warns, "secrets.stores is omitted") {
		t.Errorf("want stores-omitted warning, got %v", warns)
	}
}

func TestStoresWiredExistingNoWarning(t *testing.T) {
	k := &spec.Kubernetes{Deployment: spec.Deployment{Name: "c", Namespace: "ns", Image: "img", Replicas: 1}, Secrets: spec.Secrets{Stores: &spec.StoresSecret{Existing: "my-tls"}}}
	_, warns := Run(Context{Workflows: []spec.Workflow{wf("x.yaml", vSolace("Q", spec.DestQueue, ""), vMQ("M", spec.DestQueue, true))}, Defaults: defsWithStores(), Kube: k, Deploy: true})
	if hasErr(warns, "secrets.stores is omitted") {
		t.Errorf("stores wired via existing should not warn: %v", warns)
	}
}

func TestUnsuppliedVarsWarning(t *testing.T) {
	src := vSolace("Q", spec.DestQueue, "")
	tgt := vMQ("M", spec.DestQueue, false)
	d := defsWithStores()
	d.Security = spec.Security{Present: true, Enabled: true, Users: []spec.User{{Name: "hc", Password: "${HC}"}}}
	k := &spec.Kubernetes{
		Deployment: spec.Deployment{Name: "c", Namespace: "ns", Image: "img", Replicas: 1},
		Secrets:    spec.Secrets{Credentials: &spec.CredentialsSecret{Create: &spec.CredCreate{Name: "s", Source: spec.SourceEnv, Variables: []string{"P"}}}},
	}
	env := func(string) (string, bool) { return "x", true }
	_, warns := Run(Context{Workflows: []spec.Workflow{wf("10.yaml", src, tgt)}, Defaults: d, Kube: k, Deploy: true, Env: env})
	if !hasErr(warns, "${T} is used but not supplied") || !hasErr(warns, "${HC} is used but not supplied") {
		t.Errorf("want unsupplied warnings, got %v", warns)
	}
}

func TestSolaceQueueDestinationWarnsNotErrors(t *testing.T) {
	// MQ source -> Solace queue target: allowed, but flagged with an EDA advisory.
	errs, warns := Run(Context{Workflows: []spec.Workflow{wf("x.yaml", vMQ("MQ", spec.DestQueue, false), vSolace("Q", spec.DestQueue, ""))}, Defaults: &spec.Defaults{}})
	if len(errs) != 0 {
		t.Fatalf("Solace queue destination should be allowed, got errors %v", errs)
	}
	if !hasErr(warns, "point-to-point") {
		t.Fatalf("want EDA warning for a Solace queue destination, got %v", warns)
	}
}

func TestIdiomaticSolaceCombosNoEDAWarn(t *testing.T) {
	// Idiomatic pair: publish to a Solace topic (target), consume from a Solace queue (source).
	errs, warns := Run(Context{Workflows: []spec.Workflow{
		wf("0.yaml", vMQ("MQ", spec.DestQueue, false), vSolace("evt/out", spec.DestTopic, "")),
		wf("1.yaml", vSolace("Q.IN", spec.DestQueue, ""), vMQ("MQ2", spec.DestQueue, false)),
	}, Defaults: &spec.Defaults{}})
	if len(errs) != 0 {
		t.Fatalf("idiomatic combos should have no errors, got %v", errs)
	}
	if hasErr(warns, "non-durable subscription") || hasErr(warns, "point-to-point") {
		t.Errorf("idiomatic combos should emit no EDA warnings, got %v", warns)
	}
}

func connDefaults() *spec.Defaults {
	d := defsWithStores()
	d.Connections = map[string]spec.Side{
		"edge": {System: spec.SystemSolace, Host: "tcps://b:55443", MsgVPN: "prod", ClientUser: "u", ClientPass: "${P}"},
		"qm":   {System: spec.SystemMQ, ConnName: "h(1414)", QueueManager: "QM", Channel: "C", User: "u", Password: "${MQ}"},
	}
	return d
}

func TestConnRefStrictOnlyDestination(t *testing.T) {
	d := connDefaults()
	// A conn-ref side that also sets a connection field (host) is rejected.
	src := spec.Side{System: spec.SystemSolace, ConnRef: "edge", Host: "tcps://other", DestKind: spec.DestQueue, Dest: "Q"}
	errs, _ := Run(Context{Workflows: []spec.Workflow{wf("x.yaml", src, vMQ("M", spec.DestQueue, false))}, Defaults: d})
	if !hasErr(errs, "may set only queue/topic") {
		t.Fatalf("want strict conn-ref error, got %v", errs)
	}
}

func TestConnRefUnknownAndSystemMismatch(t *testing.T) {
	d := connDefaults()
	e1, _ := Run(Context{Workflows: []spec.Workflow{wf("x.yaml", spec.Side{System: spec.SystemSolace, ConnRef: "nope", DestKind: spec.DestQueue, Dest: "Q"}, vMQ("M", spec.DestQueue, false))}, Defaults: d})
	if !hasErr(e1, "is not defined under connections") {
		t.Errorf("want unknown-ref error, got %v", e1)
	}
	// MQ side referencing a solace connection.
	e2, _ := Run(Context{Workflows: []spec.Workflow{wf("x.yaml", vSolace("Q", spec.DestQueue, ""), spec.Side{System: spec.SystemMQ, ConnRef: "edge", DestKind: spec.DestQueue, Dest: "Q"})}, Defaults: d})
	if !hasErr(e2, `is a "solace" connection but is referenced under "mq"`) {
		t.Errorf("want system-mismatch error, got %v", e2)
	}
}

func TestConnRefValidResolvesNoError(t *testing.T) {
	d := connDefaults()
	src := spec.Side{System: spec.SystemSolace, ConnRef: "edge", DestKind: spec.DestQueue, Dest: "Q"}
	tgt := spec.Side{System: spec.SystemMQ, ConnRef: "qm", DestKind: spec.DestQueue, Dest: "MQ"}
	if errs, _ := Run(Context{Workflows: []spec.Workflow{wf("x.yaml", src, tgt)}, Defaults: d}); len(errs) != 0 {
		t.Fatalf("valid conn-ref should pass, got %v", errs)
	}
}

func baseKubeDeploy() spec.Deployment {
	return spec.Deployment{Name: "c", Namespace: "ns", Image: "img", Replicas: 1}
}

func TestCheckSyslog(t *testing.T) {
	run := func(sys *spec.Syslog) (errs, warns []Issue) {
		k := &spec.Kubernetes{Deployment: baseKubeDeploy(), Logging: &spec.Logging{Syslog: sys}}
		return Run(Context{Workflows: wfOK(), Defaults: &spec.Defaults{}, Kube: k, Deploy: true})
	}
	if e, _ := run(&spec.Syslog{Port: 514, Protocol: spec.SyslogUDP}); !hasErr(e, "logging.syslog.host is required") {
		t.Errorf("want missing host, got %v", e)
	}
	if e, _ := run(&spec.Syslog{Host: "bad host;", Port: 514, Protocol: spec.SyslogUDP}); !hasErr(e, "may only contain") {
		t.Errorf("want bad host chars, got %v", e)
	}
	if e, _ := run(&spec.Syslog{Host: "h", Port: 0, Protocol: spec.SyslogUDP}); !hasErr(e, "must be 1-65535") {
		t.Errorf("want bad port (0), got %v", e)
	}
	if e, _ := run(&spec.Syslog{Host: "h", Port: 70000, Protocol: spec.SyslogUDP}); !hasErr(e, "must be 1-65535") {
		t.Errorf("want bad port (70000), got %v", e)
	}
	if e, _ := run(&spec.Syslog{Host: "h", Port: 514, Protocol: "xxx"}); !hasErr(e, `must be "udp" or "tcp"`) {
		t.Errorf("want bad protocol, got %v", e)
	}
	if e, w := run(&spec.Syslog{Host: "h", Port: 514, Protocol: spec.SyslogTCP}); len(e) != 0 || !hasErr(w, "logstash-logback-encoder") {
		t.Errorf("want tcp warning only, errs=%v warns=%v", e, w)
	}
	if e, _ := run(&spec.Syslog{Host: "h", Port: 514, Protocol: spec.SyslogUDP}); len(e) != 0 {
		t.Errorf("valid udp syslog should have no errors, got %v", e)
	}
}

func TestCheckLibs(t *testing.T) {
	run := func(lb *spec.Libs) (errs, warns []Issue) {
		k := &spec.Kubernetes{Deployment: baseKubeDeploy(), Libs: lb}
		return Run(Context{Workflows: wfOK(), Defaults: &spec.Defaults{}, Kube: k, Deploy: true})
	}
	if e, _ := run(&spec.Libs{}); !hasErr(e, "exactly one of 'pvc' or 'download'") {
		t.Errorf("want zero-modes error, got %v", e)
	}
	if e, _ := run(&spec.Libs{PVC: &spec.LibsPVC{Existing: "x"}, Download: &spec.LibsDownload{URLs: []string{"https://x/a.jar"}}}); !hasErr(e, "exactly one of 'pvc' or 'download'") {
		t.Errorf("want two-modes error, got %v", e)
	}
	if e, _ := run(&spec.Libs{PVC: &spec.LibsPVC{}}); !hasErr(e, "exactly one of 'create' or 'existing'") {
		t.Errorf("want pvc-neither error, got %v", e)
	}
	if e, _ := run(&spec.Libs{PVC: &spec.LibsPVC{Create: &spec.PVCCreate{Name: "p", NFS: spec.NFS{Server: "s", Path: "/x"}}, Existing: "x"}}); !hasErr(e, "exactly one of 'create' or 'existing'") {
		t.Errorf("want pvc-both error, got %v", e)
	}
	if e, _ := run(&spec.Libs{PVC: &spec.LibsPVC{Create: &spec.PVCCreate{Name: "p"}}}); !hasErr(e, "requires nfs.server and nfs.path") {
		t.Errorf("want missing nfs error, got %v", e)
	}
	if e, _ := run(&spec.Libs{PVC: &spec.LibsPVC{Create: &spec.PVCCreate{Name: "Bad_Name", NFS: spec.NFS{Server: "s", Path: "/x"}}}}); !hasErr(e, "DNS-1123") {
		t.Errorf("want DNS-1123 name error, got %v", e)
	}
	if e, _ := run(&spec.Libs{Download: &spec.LibsDownload{}}); !hasErr(e, "non-empty 'urls' list") {
		t.Errorf("want empty urls error, got %v", e)
	}
	if e, _ := run(&spec.Libs{Download: &spec.LibsDownload{URLs: []string{"ftp://x/a.jar"}}}); !hasErr(e, "must be http(s)") {
		t.Errorf("want non-http url error, got %v", e)
	}
	if e, _ := run(&spec.Libs{Download: &spec.LibsDownload{URLs: []string{"https://x/a';rm -rf /'"}}}); !hasErr(e, "no spaces, quotes, or control characters") {
		t.Errorf("want injection url error, got %v", e)
	}
	if e, _ := run(&spec.Libs{Download: &spec.LibsDownload{URLs: []string{"https://x/a$(id).jar"}}}); !hasErr(e, "no spaces, quotes, or control characters") {
		t.Errorf("want dollar-injection url error, got %v", e)
	}
	if e, _ := run(&spec.Libs{Download: &spec.LibsDownload{URLs: []string{"https://x/a.jar"}, PVC: "Bad"}}); !hasErr(e, "DNS-1123") {
		t.Errorf("want DNS-1123 pvc error, got %v", e)
	}
	if e, _ := run(&spec.Libs{PVC: &spec.LibsPVC{Existing: "my-pvc"}}); len(e) != 0 {
		t.Errorf("valid pvc.existing should have no errors, got %v", e)
	}
	if e, _ := run(&spec.Libs{Download: &spec.LibsDownload{URLs: []string{"https://x/a.jar"}}}); len(e) != 0 {
		t.Errorf("valid download should have no errors, got %v", e)
	}
}

func TestConnectionDefinitionValidation(t *testing.T) {
	d := defsWithStores()
	d.Connections = map[string]spec.Side{
		"bad-dest":   {System: spec.SystemSolace, Host: "tcps://b", MsgVPN: "v", ClientUser: "u", ClientPass: "p", DestKind: spec.DestQueue, Dest: "Q"},
		"incomplete": {System: spec.SystemMQ, ConnName: "h(1414)"},
	}
	errs, _ := Run(Context{Workflows: wfOK(), Defaults: d})
	if !hasErr(errs, "must not define queue/topic") {
		t.Errorf("want no-destination error, got %v", errs)
	}
	if !hasErr(errs, "connections.incomplete mq: missing 'queue-manager'") {
		t.Errorf("want incomplete-connection error, got %v", errs)
	}
}
