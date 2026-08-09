package validate

import (
	"testing"

	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/spec"
)

func wfOK() []spec.Workflow {
	return []spec.Workflow{wf("x.yaml", vSolace("Q", spec.DestQueue, ""), vMQ("M", spec.DestQueue, false))}
}

// vSolacePlainTCP is vSolace but with a plain tcp:// host (no TLS) -- for
// exercising the usesTLS negative path, where no side has a tcps:// Solace host.
func vSolacePlainTCP(dest, kind string) spec.Side {
	return spec.Side{System: spec.SystemSolace, Host: "tcp://b:55443", MsgVPN: "prod",
		ClientUser: "u", ClientPass: "${P}", DestKind: kind, Dest: dest}
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
	// key-alias-with-no-keystore is already pinned by TestKeyAliasNeedsKeystore (validate_test.go).
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
	// deployment.image-required is already pinned by TestDeployKubeChecks (validate_test.go).
	for _, sub := range []string{"deployment.name is required", "deployment.namespace is required", "replicas: 1"} {
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
	if errs, _ := Run(Context{Workflows: wfOK(), Defaults: &spec.Defaults{}, Kube: k, Deploy: true}); !hasErr(errs, "requires tls.truststore") {
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

func dockerOK() *spec.Docker {
	return &spec.Docker{Command: "docker", Image: "img", Name: "c", Restart: "unless-stopped", Ports: []spec.Port{{Host: 8090, Container: 8090}}}
}

func TestCheckDocker(t *testing.T) {
	// Valid docker section: no errors.
	if e, _ := Run(Context{Workflows: wfOK(), Defaults: &spec.Defaults{}, Docker: dockerOK(), CheckDocker: true}); len(e) != 0 {
		t.Fatalf("valid docker should pass, got %v", e)
	}
	// Missing image + empty command + bad port.
	d := &spec.Docker{Command: "", Ports: []spec.Port{{Host: 0, Container: 0}}}
	e, _ := Run(Context{Workflows: wfOK(), Defaults: &spec.Defaults{}, Docker: d, CheckDocker: true})
	for _, sub := range []string{"docker.image is required", "docker.command must not be empty", "docker.ports host port 0 must be 1-65535"} {
		if !hasErr(e, sub) {
			t.Errorf("want %q, got %v", sub, e)
		}
	}
	// A command with a shell metacharacter is rejected as unsafe.
	d2 := &spec.Docker{Command: "docker; rm -rf /", Image: "img", Name: "c", Ports: []spec.Port{{Host: 8090, Container: 8090}}}
	if e, _ := Run(Context{Workflows: wfOK(), Defaults: &spec.Defaults{}, Docker: d2, CheckDocker: true}); !hasErr(e, "unsafe character") {
		t.Errorf("want unsafe-command error, got %v", e)
	}
	// stores set but no truststore defined.
	d3 := dockerOK()
	d3.Stores = &spec.StoresMount{MountPath: "/x"}
	if e, _ := Run(Context{Workflows: wfOK(), Defaults: &spec.Defaults{}, Docker: d3, CheckDocker: true}); !hasErr(e, "docker.stores requires tls.truststore") {
		t.Errorf("want stores-truststore error, got %v", e)
	}
	// libs set but no dir.
	d4 := dockerOK()
	d4.Libs = &spec.LibsMount{MountPath: "/x"}
	if e, _ := Run(Context{Workflows: wfOK(), Defaults: &spec.Defaults{}, Docker: d4, CheckDocker: true}); !hasErr(e, "docker.libs.dir is required") {
		t.Errorf("want libs-dir error, got %v", e)
	}
	// Gate off: docker section is not checked when CheckDocker is false.
	if e, _ := Run(Context{Workflows: wfOK(), Defaults: &spec.Defaults{}, Docker: &spec.Docker{}}); hasErr(e, "docker.image is required") {
		t.Errorf("docker checks must be gated by CheckDocker, got %v", e)
	}
}

func TestCheckDockerCredentials(t *testing.T) {
	d := dockerOK()
	d.Secrets = spec.Secrets{Credentials: &spec.CredentialsSecret{Create: &spec.CredCreate{Name: "s", Source: spec.SourceEnv}}}
	if e, _ := Run(Context{Workflows: wfOK(), Defaults: &spec.Defaults{}, Docker: d, CheckDocker: true}); !hasErr(e, "docker.secrets.credentials.create source: env requires a non-empty 'variables'") {
		t.Errorf("want docker credential error, got %v", e)
	}
}

func TestCheckPodmanModeAndScope(t *testing.T) {
	if e, _ := Run(Context{Workflows: wfOK(), Defaults: &spec.Defaults{}, Podman: podmanOK(), CheckPodman: true}); len(e) != 0 {
		t.Fatalf("valid podman should pass, got %v", e)
	}
	// Bad mode.
	p1 := podmanOK()
	p1.Mode = "swarm"
	if e, _ := Run(Context{Workflows: wfOK(), Defaults: &spec.Defaults{}, Podman: p1, CheckPodman: true}); !hasErr(e, "podman.mode must be") {
		t.Errorf("want bad-mode error, got %v", e)
	}
	// Bad quadlet scope.
	p2 := podmanOK()
	p2.Quadlet = &spec.Quadlet{Scope: "root"}
	if e, _ := Run(Context{Workflows: wfOK(), Defaults: &spec.Defaults{}, Podman: p2, CheckPodman: true}); !hasErr(e, "podman.quadlet.scope must be auto, user, or system") {
		t.Errorf("want bad-scope error, got %v", e)
	}
}

func TestCheckCommandMultiToken(t *testing.T) {
	// A multi-token command (extra kubeconfig/context args) is fine when every
	// token is safe; an unsafe token anywhere is flagged.
	d := &spec.Docker{Command: "docker --context prod", Image: "img", Name: "c", Ports: []spec.Port{{Host: 8090, Container: 8090}}}
	if e, _ := Run(Context{Workflows: wfOK(), Defaults: &spec.Defaults{}, Docker: d, CheckDocker: true}); hasErr(e, "unsafe character") {
		t.Errorf("safe multi-token command should pass, got %v", e)
	}
	d2 := &spec.Docker{Command: "docker --host $(evil)", Image: "img", Name: "c", Ports: []spec.Port{{Host: 8090, Container: 8090}}}
	if e, _ := Run(Context{Workflows: wfOK(), Defaults: &spec.Defaults{}, Docker: d2, CheckDocker: true}); !hasErr(e, "unsafe character") {
		t.Errorf("want unsafe-token error, got %v", e)
	}
}

func TestSafeToken(t *testing.T) {
	for tok, want := range map[string]bool{
		"kubectl": true, "docker": true, "--context=prod": true,
		// Legitimate CLI tokens (paths, flags, server URLs) must pass -- guards
		// against the safe charset being tightened into rejecting real commands.
		"/usr/local/bin/kubectl": true, "--namespace=solace-connectors": true, "--server=https://api.k8s.local:6443": true,
		"a b": false, "a;b": false, "a$b": false, "a`b": false, `a\b`: false, "a'b": false, `a"b`: false,
		// shellMeta-only chars: safeShellChars alone would pass these, so they pin
		// the SafeToken metacharacter layer specifically.
		"a|b": false, "a&b": false, "a>b": false, "a<b": false, "a(b)": false, "a*b": false, "a?b": false, "a#b": false, "a!b": false,
		// Mutation-coverage rows: these protect shellMeta (validate.go:712) and the
		// 0x7f disjunct of safeShellChars (validate.go:488) -- 100 percent statement
		// coverage does not. Deleting the 0x7f check, or trimming brackets/braces/
		// tilde from shellMeta, would pass the suite without these rows.
		"a\x00b": false, "a\x7fb": false, "a[b": false, "a]b": false, "a{b": false, "a}b": false, "a~b": false,
	} {
		if got := SafeToken(tok); got != want {
			t.Errorf("SafeToken(%q)=%v want %v", tok, got, want)
		}
	}
	// SafeToken("") is true: unreachable from both callers (strings.Fields never
	// yields an empty token), pinned here as documented exported-API behavior.
	if !SafeToken("") {
		t.Errorf(`SafeToken("")=false want true`)
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

func podmanOK() *spec.Podman {
	return &spec.Podman{Command: "podman", Image: "img", Name: "c", Ports: []spec.Port{{Host: 8090, Container: 8090}}, Mode: spec.PodmanModeRun, Quadlet: &spec.Quadlet{Scope: spec.QuadletScopeAuto}}
}

func TestCheckContainerNameRejected(t *testing.T) {
	// docker.name/podman.name flow into filesystem paths and a systemctl unit
	// token, so they must be valid DNS-1123 labels; traversal and uppercase/
	// underscore names are rejected. (Defaults fill a valid name before validate,
	// so only a user-supplied bad name reaches here.)
	for _, bad := range []string{"../evil", "Bad_Name"} {
		d := dockerOK()
		d.Name = bad
		if e, _ := Run(Context{Workflows: wfOK(), Defaults: &spec.Defaults{}, Docker: d, CheckDocker: true}); !hasErr(e, "docker.name") {
			t.Errorf("docker.name %q should be rejected, got %v", bad, e)
		}
		p := podmanOK()
		p.Name = bad
		if e, _ := Run(Context{Workflows: wfOK(), Defaults: &spec.Defaults{}, Podman: p, CheckPodman: true}); !hasErr(e, "podman.name") {
			t.Errorf("podman.name %q should be rejected, got %v", bad, e)
		}
	}
	// The default name is a valid DNS-1123 label and passes.
	d := dockerOK()
	d.Name = "solmq-connector"
	if e, _ := Run(Context{Workflows: wfOK(), Defaults: &spec.Defaults{}, Docker: d, CheckDocker: true}); hasErr(e, "docker.name") {
		t.Errorf("default docker.name should be accepted, got %v", e)
	}
}

func TestCheckStoresMountPathRejected(t *testing.T) {
	// The in-container store path is fixed (application.yml always points at it),
	// so a custom stores.mount-path is rejected -- it would silently break TLS by
	// moving the mounted files. The fixed path is accepted.
	d := dockerOK()
	d.Stores = &spec.StoresMount{MountPath: "/custom/path"}
	if e, _ := Run(Context{Workflows: wfOK(), Defaults: defsWithStores(), Docker: d, CheckDocker: true}); !hasErr(e, "mount-path") || !hasErr(e, "is not supported") {
		t.Errorf("non-default docker.stores.mount-path should be rejected, got %v", e)
	}
	d2 := dockerOK()
	d2.Stores = &spec.StoresMount{MountPath: spec.DefaultStoresMountPath}
	if e, _ := Run(Context{Workflows: wfOK(), Defaults: defsWithStores(), Docker: d2, CheckDocker: true}); hasErr(e, "mount-path") {
		t.Errorf("the fixed docker.stores.mount-path should be accepted, got %v", e)
	}
}

func TestDockerPodmanTLSWithoutStoresWarning(t *testing.T) {
	// A TLS workflow with no host stores bind-mounted leaves application.yml
	// pointing at absent files, so each of docker/podman warns when stores is
	// omitted and stays quiet once stores is wired.
	tlsWF := []spec.Workflow{wf("x.yaml", vSolace("Q", spec.DestQueue, ""), vMQ("M", spec.DestQueue, true))}

	if _, w := Run(Context{Workflows: tlsWF, Defaults: defsWithStores(), Docker: dockerOK(), CheckDocker: true}); !hasErr(w, "docker.stores is omitted") {
		t.Errorf("want docker TLS-without-stores warning, got %v", w)
	}
	dYes := dockerOK()
	dYes.Stores = &spec.StoresMount{MountPath: spec.DefaultStoresMountPath}
	if _, w := Run(Context{Workflows: tlsWF, Defaults: defsWithStores(), Docker: dYes, CheckDocker: true}); hasErr(w, "docker.stores is omitted") {
		t.Errorf("docker stores wired should not warn, got %v", w)
	}

	if _, w := Run(Context{Workflows: tlsWF, Defaults: defsWithStores(), Podman: podmanOK(), CheckPodman: true}); !hasErr(w, "podman.stores is omitted") {
		t.Errorf("want podman TLS-without-stores warning, got %v", w)
	}
	pYes := podmanOK()
	pYes.Stores = &spec.StoresMount{MountPath: spec.DefaultStoresMountPath}
	if _, w := Run(Context{Workflows: tlsWF, Defaults: defsWithStores(), Podman: pYes, CheckPodman: true}); hasErr(w, "podman.stores is omitted") {
		t.Errorf("podman stores wired should not warn, got %v", w)
	}
}

func TestUsesTLS(t *testing.T) {
	cases := []struct {
		name string
		wfs  []spec.Workflow
		want bool
	}{
		{
			// Solace side with a tcps:// host hits the Solace branch (validate.go:561-563).
			name: "solace tcps host",
			wfs:  []spec.Workflow{wf("x.yaml", vSolace("Q", spec.DestQueue, ""), vMQ("M", spec.DestQueue, false))},
			want: true,
		},
		{
			// No Solace side at all (so no tcps:// anywhere): only the MQ TLS branch
			// (validate.go:564-566) can return true here. Every other call site in
			// this package pairs vSolace (always tcps://) with vMQ, which always
			// short-circuits on the Solace branch first -- this row is the only one
			// that actually exercises the MQ branch.
			name: "mq tls true, no tcps anywhere",
			wfs:  []spec.Workflow{wf("x.yaml", vMQ("Q1", spec.DestQueue, false), vMQ("M1", spec.DestQueue, true))},
			want: true,
		},
		{
			// Plain tcp:// Solace + MQ TLS false: neither branch matches, so the
			// function falls through to the final "return false" (validate.go:569).
			name: "plain tcp solace, mq tls false",
			wfs:  []spec.Workflow{wf("x.yaml", vSolacePlainTCP("Q", spec.DestQueue), vMQ("M", spec.DestQueue, false))},
			want: false,
		},
	}
	for _, c := range cases {
		if got := usesTLS(c.wfs); got != c.want {
			t.Errorf("%s: usesTLS()=%v want %v", c.name, got, c.want)
		}
	}
}

func TestMQOnlyTLSStoresOmittedWarning(t *testing.T) {
	// MQ-only TLS workflow (no tcps Solace side anywhere) against a docker target
	// with stores omitted: the stores-missing warning must still be emitted, since
	// usesTLS is reached via the MQ TLS branch rather than a tcps Solace side.
	mqOnlyTLS := []spec.Workflow{wf("x.yaml", vMQ("Q1", spec.DestQueue, false), vMQ("M1", spec.DestQueue, true))}
	_, warns := Run(Context{Workflows: mqOnlyTLS, Defaults: &spec.Defaults{}, Docker: dockerOK(), CheckDocker: true})
	if !hasErr(warns, "a TLS/mTLS connection exists but docker.stores is omitted; the store files will be missing at runtime") {
		t.Errorf("want docker stores-missing warning, got %v", warns)
	}
}

func TestPlainTCPStoresOmittedNoWarning(t *testing.T) {
	// Plain-TCP workflow (no TLS/mTLS anywhere), stores omitted: the
	// stores-missing warning must be absent. Other warnings may still fire, so
	// this asserts the specific text is missing rather than warns being empty.
	plainTCP := []spec.Workflow{wf("x.yaml", vSolacePlainTCP("Q", spec.DestQueue), vMQ("M", spec.DestQueue, false))}
	_, warns := Run(Context{Workflows: plainTCP, Defaults: &spec.Defaults{}, Docker: dockerOK(), CheckDocker: true})
	if hasErr(warns, "store files will be missing at runtime") {
		t.Errorf("plain-tcp workflow with no TLS should not warn about missing store files, got %v", warns)
	}
}
