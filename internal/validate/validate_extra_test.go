package validate

import (
	"strconv"
	"strings"
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
		ClientUser: "u", ClientPassEnv: "SOLACE_CLIENT_PASSWORD", DestKind: kind, Dest: dest}
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
	errs, _ := Run(Context{Workflows: wfOK(), Defaults: &spec.Defaults{}, Image: imageOK(), Kube: k, CheckKubernetes: true})
	for _, sub := range []string{"deployment.name is required", "deployment.namespace is required", "replicas: 1"} {
		if !hasErr(errs, sub) {
			t.Errorf("want %q, got %v", sub, errs)
		}
	}
}

// TestCheckKubeServicePort pins checkPortRange applied to kubernetes.service.port
// (a spec.Port, same host/container shape docker/podman ports already use):
// each side is range-checked independently, and a "scalar" port (host ==
// container, the common case) validates the same as a distinct host:container
// pair -- the shape itself is enforced at parse in spec.Port.UnmarshalYAML,
// not here.
func TestCheckKubeServicePort(t *testing.T) {
	run := func(p spec.Port) []Issue {
		k := &spec.Kubernetes{Deployment: baseKubeDeploy(), Command: spec.DefaultKubeCommand, Service: spec.Service{Port: p}}
		errs, _ := Run(Context{Workflows: wfOK(), Defaults: &spec.Defaults{}, Image: imageOK(), Kube: k, CheckKubernetes: true})
		return errs
	}
	// Scalar form (host == container): valid.
	if e := run(spec.Port{Host: 8090, Container: 8090}); hasErr(e, "kubernetes.service.port") {
		t.Errorf("valid scalar service.port should pass, got %v", e)
	}
	// host:container form with distinct values: valid.
	if e := run(spec.Port{Host: 8080, Container: 9090}); hasErr(e, "kubernetes.service.port") {
		t.Errorf("valid host:container service.port should pass, got %v", e)
	}
	// Host side out of range, container side left valid.
	if e := run(spec.Port{Host: 0, Container: 8090}); !hasErr(e, "kubernetes.service.port host port 0 must be 1-65535") {
		t.Errorf("want host-port range error, got %v", e)
	}
	if e := run(spec.Port{Host: 70000, Container: 8090}); !hasErr(e, "kubernetes.service.port host port 70000 must be 1-65535") {
		t.Errorf("want host-port range error, got %v", e)
	}
	// Container side out of range, host side left valid.
	if e := run(spec.Port{Host: 8090, Container: 0}); !hasErr(e, "kubernetes.service.port container port 0 must be 1-65535") {
		t.Errorf("want container-port range error, got %v", e)
	}
	if e := run(spec.Port{Host: 8090, Container: 70000}); !hasErr(e, "kubernetes.service.port container port 70000 must be 1-65535") {
		t.Errorf("want container-port range error, got %v", e)
	}
}

// TestCheckKubeCredentialCreateRemovedKeys covers CredCreate.RemovedKeys(): the
// values-file/env-var-list mechanism is gone (the Secret's keys are now derived
// from the config's own credential fields), so a config still setting
// source/variables/values-file must fail loudly instead of silently losing its
// credentials.
func TestCheckKubeCredentialCreateRemovedKeys(t *testing.T) {
	base := spec.Deployment{Name: "c", Namespace: "ns", Replicas: 1}
	run := func(c *spec.CredCreate) []Issue {
		k := &spec.Kubernetes{Deployment: base, Secrets: spec.Secrets{Credentials: &spec.CredentialsSecret{Create: c}}}
		errs, _ := Run(Context{Workflows: wfOK(), Defaults: &spec.Defaults{}, Image: imageOK(), Kube: k, CheckKubernetes: true})
		return errs
	}
	if e := run(&spec.CredCreate{Name: "s", Source: "env", Variables: []string{"P"}, ValuesFile: "vals.yaml"}); !hasErr(e, "no longer takes source/variables/values-file") || !hasErr(e, "Remove source, variables, values-file") {
		t.Errorf("want removed-keys error naming all three, got %v", e)
	}
	if e := run(&spec.CredCreate{Name: "s", Source: "file"}); !hasErr(e, "no longer takes source") || !hasErr(e, "Remove source") {
		t.Errorf("want removed-keys error naming source alone, got %v", e)
	}
	// The new shape -- just a name -- takes no removed keys.
	if e := run(&spec.CredCreate{Name: "s"}); hasErr(e, "no longer takes") {
		t.Errorf("a bare create.name should not trip the removed-keys check, got %v", e)
	}
}

func TestCheckKubeStoresRequireTruststore(t *testing.T) {
	k := &spec.Kubernetes{Deployment: spec.Deployment{Name: "c", Namespace: "ns", Replicas: 1}, Secrets: spec.Secrets{Stores: &spec.StoresSecret{Create: &spec.StoreCreate{Name: "t"}}}}
	if errs, _ := Run(Context{Workflows: wfOK(), Defaults: &spec.Defaults{}, Image: imageOK(), Kube: k, CheckKubernetes: true}); !hasErr(errs, "requires tls.truststore") {
		t.Errorf("want truststore requirement, got %v", errs)
	}
}

func TestStoresNotWiredWarning(t *testing.T) {
	k := &spec.Kubernetes{Deployment: spec.Deployment{Name: "c", Namespace: "ns", Replicas: 1}}
	_, warns := Run(Context{Workflows: []spec.Workflow{wf("x.yaml", vSolace("Q", spec.DestQueue, ""), vMQ("M", spec.DestQueue, true))}, Defaults: defsWithStores(), Image: imageOK(), Kube: k, CheckKubernetes: true})
	if !hasErr(warns, "secrets.stores is omitted") {
		t.Errorf("want stores-omitted warning, got %v", warns)
	}
}

func TestStoresWiredExistingNoWarning(t *testing.T) {
	k := &spec.Kubernetes{Deployment: spec.Deployment{Name: "c", Namespace: "ns", Replicas: 1}, Secrets: spec.Secrets{Stores: &spec.StoresSecret{Existing: "my-tls"}}}
	_, warns := Run(Context{Workflows: []spec.Workflow{wf("x.yaml", vSolace("Q", spec.DestQueue, ""), vMQ("M", spec.DestQueue, true))}, Defaults: defsWithStores(), Image: imageOK(), Kube: k, CheckKubernetes: true})
	if hasErr(warns, "secrets.stores is omitted") {
		t.Errorf("stores wired via existing should not warn: %v", warns)
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
		"edge": {System: spec.SystemSolace, Host: "tcps://b:55443", MsgVPN: "prod", ClientUser: "u", ClientPassEnv: "EDGE_CLIENT_PASSWORD"},
		"qm":   {System: spec.SystemMQ, ConnName: "h(1414)", QueueManager: "QM", Channel: "C", User: "u", PasswordEnv: "QM_PASSWORD"},
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
	return spec.Deployment{Name: "c", Namespace: "ns", Replicas: 1}
}

// baseKubeService is a valid kubernetes.service (matching the port docker/podman
// fixtures use) for tests that need checkPortRange to stay quiet while
// exercising an unrelated check.
func baseKubeService() spec.Service {
	return spec.Service{Port: spec.Port{Host: 8090, Container: 8090}}
}

func TestCheckSyslog(t *testing.T) {
	run := func(sys *spec.Syslog) (errs, warns []Issue) {
		k := &spec.Kubernetes{Deployment: baseKubeDeploy(), Command: spec.DefaultKubeCommand, Service: baseKubeService()}
		return Run(Context{Workflows: wfOK(), Defaults: &spec.Defaults{Syslog: sys}, Image: imageOK(), Kube: k, CheckKubernetes: true})
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
		k := &spec.Kubernetes{Deployment: baseKubeDeploy(), Command: spec.DefaultKubeCommand, Service: baseKubeService(), Libs: lb}
		return Run(Context{Workflows: wfOK(), Defaults: &spec.Defaults{}, Image: imageOK(), Kube: k, CheckKubernetes: true})
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
	return &spec.Docker{Command: "docker", Name: "c", ProjectName: "proj", Restart: "unless-stopped", Ports: []spec.Port{{Host: 8090, Container: 8090}}}
}

func TestCheckDocker(t *testing.T) {
	// Valid docker section: no errors.
	if e, _ := Run(Context{Workflows: wfOK(), Defaults: &spec.Defaults{}, Image: imageOK(), Docker: dockerOK(), CheckDocker: true}); len(e) != 0 {
		t.Fatalf("valid docker should pass, got %v", e)
	}
	// Empty command + bad port (the image is a top-level key now).
	d := &spec.Docker{Command: "", Ports: []spec.Port{{Host: 0, Container: 0}}}
	e, _ := Run(Context{Workflows: wfOK(), Defaults: &spec.Defaults{}, Image: imageOK(), Docker: d, CheckDocker: true})
	for _, sub := range []string{"docker.command must not be empty", "docker.ports host port 0 must be 1-65535"} {
		if !hasErr(e, sub) {
			t.Errorf("want %q, got %v", sub, e)
		}
	}
	// A command with a shell metacharacter is rejected as unsafe.
	d2 := &spec.Docker{Command: "docker; rm -rf /", Name: "c", Ports: []spec.Port{{Host: 8090, Container: 8090}}}
	if e, _ := Run(Context{Workflows: wfOK(), Defaults: &spec.Defaults{}, Image: imageOK(), Docker: d2, CheckDocker: true}); !hasErr(e, "unsafe character") {
		t.Errorf("want unsafe-command error, got %v", e)
	}
	// stores is a removed key: any present block is rejected by name.
	d3 := dockerOK()
	d3.Stores = &spec.StoresMount{}
	if e, _ := Run(Context{Workflows: wfOK(), Defaults: &spec.Defaults{}, Image: imageOK(), Docker: d3, CheckDocker: true}); !hasErr(e, "docker.stores is no longer configured") {
		t.Errorf("want stores-removed error, got %v", e)
	}
	// libs set but no dir.
	d4 := dockerOK()
	d4.Libs = &spec.LibsMount{}
	if e, _ := Run(Context{Workflows: wfOK(), Defaults: &spec.Defaults{}, Image: imageOK(), Docker: d4, CheckDocker: true}); !hasErr(e, "docker.libs.dir is required") {
		t.Errorf("want libs-dir error, got %v", e)
	}
	// Gate off: the docker section is not checked when CheckDocker is false. The
	// probe has to name an error an unchecked empty section WOULD produce -- the
	// old "docker.image is required" no longer exists, so asserting its absence
	// would pass whether the gate worked or not.
	if e, _ := Run(Context{Workflows: wfOK(), Defaults: &spec.Defaults{}, Docker: &spec.Docker{}}); hasErr(e, "docker.command must not be empty") {
		t.Errorf("docker checks must be gated by CheckDocker, got %v", e)
	}
}

// TestCheckDockerProjectName covers the compose project name. It is held to the
// same DNS-1123 rule as docker.name, which is deliberately stricter than docker
// compose's own grammar: compose would accept an underscore and a trailing
// hyphen, so those two are rejected here on purpose rather than by accident.
//
// The check is unconditional, unlike restart's -- ParseEnv always defaults
// project-name, so an empty value means the section reached the validator
// without going through it, and emitting a compose file with no project at all
// is not a state to pass silently.
func TestCheckDockerProjectName(t *testing.T) {
	for _, c := range []struct {
		name    string
		project string
		wantErr bool
	}{
		{"lowercase and hyphens", "solace-ibmmq-connectors", false},
		{"digits", "stack1", false},
		{"single char", "a", false},
		{"uppercase", "Solace-Connectors", true},
		{"underscore -- compose allows it, DNS-1123 does not", "solmq_prod", true},
		{"trailing hyphen -- compose allows it, DNS-1123 does not", "solmq-", true},
		{"leading hyphen", "-solmq", true},
		{"embedded space", "solmq prod", true},
		{"empty -- means defaults never ran", "", true},
	} {
		t.Run(c.name, func(t *testing.T) {
			d := dockerOK()
			d.ProjectName = c.project
			e, _ := Run(Context{Workflows: wfOK(), Defaults: &spec.Defaults{}, Image: imageOK(), Docker: d, CheckDocker: true})
			// The message must carry the offending value, so an operator sees what
			// was rejected rather than only which key.
			got := hasErr(e, "docker.project-name") && hasErr(e, strconv.Quote(c.project))
			if got != c.wantErr {
				t.Errorf("project-name %q: error = %v, want %v (errors: %v)", c.project, got, c.wantErr, e)
			}
		})
	}
}

// TestCheckPodmanHasNoProjectName pins that project-name is docker-only. It
// names a compose project, and podman has no equivalent grouping, so the shared
// containerTarget must not grow one and a podman section must not be held to it.
func TestCheckPodmanHasNoProjectName(t *testing.T) {
	e, _ := Run(Context{Workflows: wfOK(), Defaults: &spec.Defaults{}, Image: imageOK(), Podman: podmanOK(), CheckPodman: true})
	if hasErr(e, "project-name") {
		t.Errorf("podman must not be checked for a project name, got %v", e)
	}
}

// TestCheckDockerPodmanSecretsRemoved covers the docker/podman `.secrets`
// removal: credentials now come from the connection fields themselves and are
// delivered as platform secrets, so a config still setting `.secrets` (old
// env-var-list shape) must fail loudly instead of silently dropping it.
func TestCheckDockerPodmanSecretsRemoved(t *testing.T) {
	d := dockerOK()
	d.Secrets = &spec.Secrets{Credentials: &spec.CredentialsSecret{Create: &spec.CredCreate{Name: "s"}}}
	if e, _ := Run(Context{Workflows: wfOK(), Defaults: &spec.Defaults{}, Image: imageOK(), Docker: d, CheckDocker: true}); !hasErr(e, "docker.secrets is no longer configured") {
		t.Errorf("want docker secrets-removed error, got %v", e)
	}
	p := podmanOK()
	p.Secrets = &spec.Secrets{Stores: &spec.StoresSecret{Existing: "x"}}
	if e, _ := Run(Context{Workflows: wfOK(), Defaults: &spec.Defaults{}, Image: imageOK(), Podman: p, CheckPodman: true}); !hasErr(e, "podman.secrets is no longer configured") {
		t.Errorf("want podman secrets-removed error, got %v", e)
	}
	// Nil secrets -- the new default shape -- trips no such error.
	if e, _ := Run(Context{Workflows: wfOK(), Defaults: &spec.Defaults{}, Image: imageOK(), Docker: dockerOK(), CheckDocker: true}); hasErr(e, "secrets is no longer configured") {
		t.Errorf("nil docker.secrets should not be rejected, got %v", e)
	}
}

func TestCheckPodmanModeAndScope(t *testing.T) {
	if e, _ := Run(Context{Workflows: wfOK(), Defaults: &spec.Defaults{}, Image: imageOK(), Podman: podmanOK(), CheckPodman: true}); len(e) != 0 {
		t.Fatalf("valid podman should pass, got %v", e)
	}
	// mode: is removed. Both of its former values are rejected, "quadlet"
	// included -- that is now the only artifact, so the key decides nothing and a
	// section still carrying it should be told rather than quietly humoured.
	for _, mode := range []string{"run", "quadlet", "swarm"} {
		p := podmanOK()
		p.Mode = mode
		if e, _ := Run(Context{Workflows: wfOK(), Defaults: &spec.Defaults{}, Image: imageOK(), Podman: p, CheckPodman: true}); !hasErr(e, "podman.mode is no longer configured") {
			t.Errorf("mode %q should be rejected, got %v", mode, e)
		}
	}
	// Bad quadlet scope.
	p2 := podmanOK()
	p2.Quadlet = &spec.Quadlet{Scope: "root"}
	if e, _ := Run(Context{Workflows: wfOK(), Defaults: &spec.Defaults{}, Image: imageOK(), Podman: p2, CheckPodman: true}); !hasErr(e, "podman.quadlet.scope must be auto, user, or system") {
		t.Errorf("want bad-scope error, got %v", e)
	}
}

// TestCheckPodmanBaseDirRequired pins the one podman key with no default. It is
// where the mounted application.yml and status script are written, and the path
// is baked into the unit's Volume= lines, so there is nothing safe to guess.
func TestCheckPodmanBaseDirRequired(t *testing.T) {
	p := podmanOK()
	p.BaseDir = ""
	if e, _ := Run(Context{Workflows: wfOK(), Defaults: &spec.Defaults{}, Image: imageOK(), Podman: p, CheckPodman: true}); !hasErr(e, "podman.base-dir is required") {
		t.Errorf("an omitted base-dir must be rejected, got %v", e)
	}
	// It reaches a Volume= line unquoted, so it takes the same host-path gate as
	// libs.dir and the tls.*.file stores.
	p2 := podmanOK()
	p2.BaseDir = "/opt/my dir"
	if e, _ := Run(Context{Workflows: wfOK(), Defaults: &spec.Defaults{}, Image: imageOK(), Podman: p2, CheckPodman: true}); !hasErr(e, "podman.base-dir") {
		t.Errorf("an unsafe base-dir must be rejected, got %v", e)
	}
	// A relative value is accepted: it resolves against env.yaml at render time,
	// exactly as libs.dir and tls.*.file do.
	p3 := podmanOK()
	p3.BaseDir = "./data"
	if e, _ := Run(Context{Workflows: wfOK(), Defaults: &spec.Defaults{}, Image: imageOK(), Podman: p3, CheckPodman: true}); hasErr(e, "base-dir") {
		t.Errorf("a relative base-dir should be accepted, got %v", e)
	}
}

// TestSafeHostPathAllowsWindowsShortNames pins the tilde exception. A Windows 8.3
// short name is a real directory an operator cannot rename -- RUNNER~1 is what
// %TEMP% expands to on a GitHub Windows runner, and PROGRA~1 is Program Files --
// so a path gate that rejects '~' rejects valid input. It reached us as a CI-only
// failure precisely because no fixture had a tilde in it; this one does.
//
// The check is on safeHostPath directly because it guards every host path
// (tls.*.file, libs.dir, podman.base-dir, nfs.path), not just the one that
// surfaced it.
func TestSafeHostPathAllowsWindowsShortNames(t *testing.T) {
	for _, ok := range []string{
		`C:\Users\RUNNER~1\AppData\Local\Temp\x`,
		"C:/Users/RUNNER~1/AppData/Local/Temp/x",
		`C:\PROGRA~1\solmq`,
		"~/certs/truststore.jks",
	} {
		if !safeHostPath(ok) {
			t.Errorf("safeHostPath(%q) = false, want true: a tilde is legal in a host path", ok)
		}
	}
	// The rest of the metacharacter set is still refused -- the tilde is the only
	// concession, and only for paths.
	for _, bad := range []string{
		"/opt/a b", "/opt/a\nb", "/opt/$HOME", "/opt/a;rm", "/opt/a|b",
		"/opt/a*b", "/opt/a(b)", "/opt/a#b", "/opt/a!b", "/opt/a`b",
	} {
		if safeHostPath(bad) {
			t.Errorf("safeHostPath(%q) = true, want false", bad)
		}
	}
}

// TestCheckLibsMountPathRemoved pins the last of the fixed-path keys. dir stays --
// the host side is genuinely the operator's choice -- but mount-path is rejected
// for both sections, because the image launches with /app/external/libs literally
// on its classpath and a custom value silently put the jars out of its reach.
func TestCheckLibsMountPathRemoved(t *testing.T) {
	d := dockerOK()
	d.Libs = &spec.LibsMount{Dir: "./libs", MountPath: "/custom/libs"}
	if e, _ := Run(Context{Workflows: wfOK(), Defaults: &spec.Defaults{}, Image: imageOK(), Docker: d, CheckDocker: true}); !hasErr(e, "docker.libs.mount-path is no longer configured") {
		t.Errorf("want libs mount-path removed error, got %v", e)
	}
	p := podmanOK()
	p.Libs = &spec.LibsMount{Dir: "./libs", MountPath: spec.DefaultLibsMountPath}
	if e, _ := Run(Context{Workflows: wfOK(), Defaults: &spec.Defaults{}, Image: imageOK(), Podman: p, CheckPodman: true}); !hasErr(e, "podman.libs.mount-path is no longer configured") {
		t.Errorf("even the fixed value should be rejected once the key is gone, got %v", e)
	}
	// dir alone is the supported shape and trips no such error.
	d2 := dockerOK()
	d2.Libs = &spec.LibsMount{Dir: "./libs"}
	if e, _ := Run(Context{Workflows: wfOK(), Defaults: &spec.Defaults{}, Image: imageOK(), Docker: d2, CheckDocker: true}); hasErr(e, "mount-path") {
		t.Errorf("libs with dir alone should pass, got %v", e)
	}
}

// TestCheckPodmanStoresRemoved is the podman half of the removed stores: block;
// the docker half rides along in TestCheckDockerRules.
func TestCheckPodmanStoresRemoved(t *testing.T) {
	p := podmanOK()
	p.Stores = &spec.StoresMount{}
	if e, _ := Run(Context{Workflows: wfOK(), Defaults: &spec.Defaults{}, Image: imageOK(), Podman: p, CheckPodman: true}); !hasErr(e, "podman.stores is no longer configured") {
		t.Errorf("want stores-removed error, got %v", e)
	}
	// Nil stores -- the only shape left -- trips no such error.
	if e, _ := Run(Context{Workflows: wfOK(), Defaults: &spec.Defaults{}, Image: imageOK(), Podman: podmanOK(), CheckPodman: true}); hasErr(e, "stores is no longer configured") {
		t.Errorf("nil podman.stores should not be rejected, got %v", e)
	}
}

func TestCheckCommandMultiToken(t *testing.T) {
	// A multi-token command (extra kubeconfig/context args) is fine when every
	// token is safe; an unsafe token anywhere is flagged.
	d := &spec.Docker{Command: "docker --context prod", Name: "c", Ports: []spec.Port{{Host: 8090, Container: 8090}}}
	if e, _ := Run(Context{Workflows: wfOK(), Defaults: &spec.Defaults{}, Image: imageOK(), Docker: d, CheckDocker: true}); hasErr(e, "unsafe character") {
		t.Errorf("safe multi-token command should pass, got %v", e)
	}
	d2 := &spec.Docker{Command: "docker --host $(evil)", Name: "c", Ports: []spec.Port{{Host: 8090, Container: 8090}}}
	if e, _ := Run(Context{Workflows: wfOK(), Defaults: &spec.Defaults{}, Image: imageOK(), Docker: d2, CheckDocker: true}); !hasErr(e, "unsafe character") {
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

// TestSafeActuatorUser pins the tighter allowlist the status account name needs
// on top of SafeToken: the name is spliced into a sed address inside the
// generated status script, so '/' (which would end the address early and leave
// the script unauthenticated) and every other non-allowlisted character are
// rejected here even though SafeToken permits some of them as argv tokens.
func TestSafeActuatorUser(t *testing.T) {
	for name, want := range map[string]bool{
		"solmq-status": true, "healthcheck": true, "svc.status": true, "a_b-c.9": true,
		// Rejected here but accepted by SafeToken: these are what make the
		// separate, stricter gate necessary rather than reusing SafeToken.
		"bad/user": false, "user:1": false, "a=b": false, "a,b": false, "a+b": false, "a@b": false,
		// Already rejected by SafeToken; pinned so the two gates cannot diverge.
		"a b": false, "a$b": false, `a\b`: false, "a'b": false, `a"b`: false, "a*b": false, "a[b": false,
		// Empty is rejected: callers default to spec.StatusUserName instead.
		"": false,
	} {
		if got := SafeActuatorUser(name); got != want {
			t.Errorf("SafeActuatorUser(%q)=%v want %v", name, got, want)
		}
		// Anything this gate accepts must also be safe as an argv token, since
		// the same value reaches the exec call.
		if want && !SafeToken(name) {
			t.Errorf("SafeActuatorUser(%q) accepted a value SafeToken rejects", name)
		}
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
	return &spec.Podman{Command: "podman", Name: "c", BaseDir: "/opt/solmq", Ports: []spec.Port{{Host: 8090, Container: 8090}}, Quadlet: &spec.Quadlet{Scope: spec.QuadletScopeAuto}}
}

func TestCheckContainerNameRejected(t *testing.T) {
	// docker.name/podman.name flow into filesystem paths and a systemctl unit
	// token, so they must be valid DNS-1123 labels; traversal and uppercase/
	// underscore names are rejected. (Defaults fill a valid name before validate,
	// so only a user-supplied bad name reaches here.)
	for _, bad := range []string{"../evil", "Bad_Name"} {
		d := dockerOK()
		d.Name = bad
		if e, _ := Run(Context{Workflows: wfOK(), Defaults: &spec.Defaults{}, Image: imageOK(), Docker: d, CheckDocker: true}); !hasErr(e, "docker.name") {
			t.Errorf("docker.name %q should be rejected, got %v", bad, e)
		}
		p := podmanOK()
		p.Name = bad
		if e, _ := Run(Context{Workflows: wfOK(), Defaults: &spec.Defaults{}, Image: imageOK(), Podman: p, CheckPodman: true}); !hasErr(e, "podman.name") {
			t.Errorf("podman.name %q should be rejected, got %v", bad, e)
		}
	}
	// The default name is a valid DNS-1123 label and passes.
	d := dockerOK()
	d.Name = "solmq-connector"
	if e, _ := Run(Context{Workflows: wfOK(), Defaults: &spec.Defaults{}, Image: imageOK(), Docker: d, CheckDocker: true}); hasErr(e, "docker.name") {
		t.Errorf("default docker.name should be accepted, got %v", e)
	}
}

// TestDockerPodmanTLSNeedsNoStoresOptIn replaces the old TLS-without-stores
// warning. The store files are bind-mounted whenever tls.*.file is set, so a TLS
// workflow with no stores: block is now complete rather than half-wired, and
// neither section has anything to warn about.
func TestDockerPodmanTLSNeedsNoStoresOptIn(t *testing.T) {
	tlsWF := []spec.Workflow{wf("x.yaml", vSolace("Q", spec.DestQueue, ""), vMQ("M", spec.DestQueue, true))}

	if _, w := Run(Context{Workflows: tlsWF, Defaults: defsWithStores(), Image: imageOK(), Docker: dockerOK(), CheckDocker: true}); hasErr(w, "stores is omitted") {
		t.Errorf("docker should not warn about an omitted stores:, got %v", w)
	}
	if _, w := Run(Context{Workflows: tlsWF, Defaults: defsWithStores(), Image: imageOK(), Podman: podmanOK(), CheckPodman: true}); hasErr(w, "stores is omitted") {
		t.Errorf("podman should not warn about an omitted stores:, got %v", w)
	}
}

// TestDockerPodmanStorePathAlwaysGated pins the security boundary the derived
// mount widened. The tls.*.file paths become bind-mount sources in a compose
// document or a quadlet Volume= line with no stores: block to opt in, so the
// unsafe-character gate has to run on them unconditionally -- it used to fire
// only when the operator had opted in, which is exactly the case that no longer
// exists.
func TestDockerPodmanStorePathAlwaysGated(t *testing.T) {
	defs := defsWithStores()
	defs.TLS.Truststore.File = "./certs/$(evil).jks"

	if e, _ := Run(Context{Workflows: wfOK(), Defaults: defs, Image: imageOK(), Docker: dockerOK(), CheckDocker: true}); !hasErr(e, "unsafe character") {
		t.Errorf("docker should reject an unsafe tls.truststore.file with no stores: block, got %v", e)
	}
	if e, _ := Run(Context{Workflows: wfOK(), Defaults: defs, Image: imageOK(), Podman: podmanOK(), CheckPodman: true}); !hasErr(e, "unsafe character") {
		t.Errorf("podman should reject an unsafe tls.truststore.file with no stores: block, got %v", e)
	}
	// Kubernetes stays exempt: it embeds the store content in a Secret rather
	// than naming a host path, so it never reaches this gate.
	k := &spec.Kubernetes{Deployment: baseKubeDeploy(), Command: spec.DefaultKubeCommand}
	if e, _ := Run(Context{Workflows: wfOK(), Defaults: defs, Image: imageOK(), Kube: k, CheckKubernetes: true}); hasErr(e, "unsafe character") {
		t.Errorf("kubernetes should not gate the host store path, got %v", e)
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

// The two docker/podman stores-omitted warning tests that stood here are gone
// with the warning itself. What they were really guarding -- that usesTLS is
// reachable through the MQ TLS branch and not only a tcps Solace side -- is
// asserted directly by TestUsesTLS above, on the same workflow shape. usesTLS now
// has exactly one caller, the kubernetes stores warning, which still needs telling.

func TestBinderIdentityUsesTheCredentialPair(t *testing.T) {
	// Regression: validate and consolidate each kept a private dedupKey, and the
	// copies drifted once credentials became literal/-env pairs -- validate went
	// on keying off the literal username, so every side using client-username-env
	// keyed as if it had no username at all. Two DIFFERENT connections then
	// collapsed to one key here while staying separate in the renderer, and the
	// binder-level checks reported conflicts that do not exist. Both now share
	// spec.Side.DedupKey.
	a := vSolace("Q1", spec.DestQueue, "alias-a")
	a.ClientUser, a.ClientUserEnv = "", "TEAM_A_USER"
	b := vSolace("Q2", spec.DestQueue, "alias-b")
	b.ClientUser, b.ClientUserEnv = "", "TEAM_B_USER"
	wfs := []spec.Workflow{
		wf("0.yaml", a, vMQ("M1", spec.DestQueue, false)),
		wf("1.yaml", b, vMQ("M2", spec.DestQueue, false)),
	}
	if e, _ := Run(Context{Workflows: wfs, Defaults: defsWithStores()}); hasErr(e, "conflicting key-alias") {
		t.Errorf("different -env usernames are different binders, so their key-aliases cannot conflict: %v", e)
	}

	// The same -env username on the same host IS one binder, so a differing
	// key-alias there is still a real conflict.
	c := vSolace("Q3", spec.DestQueue, "alias-c")
	c.ClientUser, c.ClientUserEnv = "", "TEAM_A_USER"
	same := []spec.Workflow{
		wf("0.yaml", a, vMQ("M1", spec.DestQueue, false)),
		wf("1.yaml", c, vMQ("M2", spec.DestQueue, false)),
	}
	if e, _ := Run(Context{Workflows: same, Defaults: defsWithStores()}); !hasErr(e, "conflicting key-alias") {
		t.Errorf("one binder with two key-aliases must still conflict, got %v", e)
	}
}

// TestCheckContainerRestartUnsafe covers what is left of the per-platform
// charset gate now that image and timezone have moved to top-level keys (their
// retirement is covered in validate_image_test.go, their charset in
// TestImageBlockRequired and TestTopLevelTimezoneUnsafe).
//
// restart is still concatenated unquoted into the quadlet unit and the compose
// YAML, so an embedded newline or metacharacter is rejected
// at the gate. A realistic value (on-failure:N) must still pass.
func TestCheckContainerRestartUnsafe(t *testing.T) {
	dR := dockerOK()
	dR.Restart = "unless-stopped\nprivileged: true"
	if e, _ := Run(Context{Workflows: wfOK(), Defaults: &spec.Defaults{}, Image: imageOK(), Docker: dR, CheckDocker: true}); !hasErr(e, "docker.restart") {
		t.Errorf("want docker.restart rejection, got %v", e)
	}
	dOK := dockerOK()
	dOK.Restart = "on-failure:5"
	if e, _ := Run(Context{Workflows: wfOK(), Defaults: &spec.Defaults{}, Image: imageOK(), Docker: dOK, CheckDocker: true}); len(e) != 0 {
		t.Errorf("a realistic restart policy should pass, got %v", e)
	}
}

func TestCheckContainerHostPathsUnsafe(t *testing.T) {
	// The tls.*.file paths are bind-mount sources for docker/podman, and libs.dir
	// is one too: whitespace or a metacharacter would split or extend the mount
	// argument, so both are gated. A Windows-style path keeps validating -- '\'
	// and ':' cannot escape any of the three sinks.
	badStores := defsWithStores()
	badStores.TLS.Truststore.File = "t.jks\nprivileged: true"
	if e, _ := Run(Context{Workflows: wfOK(), Defaults: badStores, Image: imageOK(), Docker: dockerOK(), CheckDocker: true}); !hasErr(e, "tls.truststore.file") {
		t.Errorf("want unsafe truststore path rejection, got %v", e)
	}
	p := podmanOK()
	p.Libs = &spec.LibsMount{Dir: "/opt/my libs"}
	if e, _ := Run(Context{Workflows: wfOK(), Defaults: &spec.Defaults{}, Image: imageOK(), Podman: p, CheckPodman: true}); !hasErr(e, "podman.libs.dir") {
		t.Errorf("want unsafe libs.dir rejection, got %v", e)
	}
	winStores := defsWithStores()
	winStores.TLS.Truststore.File = `C:\certs\truststore.jks`
	winStores.TLS.Keystore.File = `C:\certs\keystore.jks`
	dWin := dockerOK()
	dWin.Libs = &spec.LibsMount{Dir: `C:\libs`}
	if e, _ := Run(Context{Workflows: wfOK(), Defaults: winStores, Image: imageOK(), Docker: dWin, CheckDocker: true}); hasErr(e, "unsafe character") {
		t.Errorf("Windows-style host paths should validate, got %v", e)
	}
}

func TestCheckKubeSecretNames(t *testing.T) {
	// Every Secret name is emitted verbatim as metadata.name / secretRef, so a
	// non-DNS-1123 or missing name is rejected before the manifest is built.
	base := spec.Deployment{Name: "c", Namespace: "ns", Replicas: 1}
	cases := []struct {
		name string
		sec  spec.Secrets
		want string
	}{
		{"cred create bad", spec.Secrets{Credentials: &spec.CredentialsSecret{Create: &spec.CredCreate{Name: "Bad_Name"}}}, "kubernetes.secrets.credentials.create.name"},
		{"cred create empty", spec.Secrets{Credentials: &spec.CredentialsSecret{Create: &spec.CredCreate{}}}, "kubernetes.secrets.credentials.create.name is required"},
		{"cred existing bad", spec.Secrets{Credentials: &spec.CredentialsSecret{Existing: "my secret\nfoo: bar"}}, "kubernetes.secrets.credentials.existing"},
		{"stores create bad", spec.Secrets{Stores: &spec.StoresSecret{Create: &spec.StoreCreate{Name: "../evil"}}}, "kubernetes.secrets.stores.create.name"},
		{"stores existing bad", spec.Secrets{Stores: &spec.StoresSecret{Existing: "UPPER"}}, "kubernetes.secrets.stores.existing"},
	}
	for _, c := range cases {
		k := &spec.Kubernetes{Deployment: base, Secrets: c.sec}
		e, _ := Run(Context{Workflows: wfOK(), Defaults: defsWithStores(), Image: imageOK(), Kube: k, CheckKubernetes: true, Env: func(string) (string, bool) { return "v", true }})
		if !hasErr(e, c.want) {
			t.Errorf("%s: want %q, got %v", c.name, c.want, e)
		}
	}
	// Valid names produce no name error.
	k := &spec.Kubernetes{Deployment: base, Secrets: spec.Secrets{
		Credentials: &spec.CredentialsSecret{Create: &spec.CredCreate{Name: "solmq-credentials"}},
		Stores:      &spec.StoresSecret{Create: &spec.StoreCreate{Name: "solmq-tls"}},
	}}
	e, _ := Run(Context{Workflows: wfOK(), Defaults: defsWithStores(), Image: imageOK(), Kube: k, CheckKubernetes: true, Env: func(string) (string, bool) { return "v", true }})
	if hasErr(e, "DNS-1123") || hasErr(e, "is required") {
		t.Errorf("valid secret names should pass, got %v", e)
	}
}

// TestCheckCredRejectsReservedPrefix pins the namespace reservation. An -env
// credential is mounted under the variable name it gives, sharing one namespace
// with the names derived for literals; rejecting the prefix here is what makes a
// derived name and an operator's name structurally unable to meet. envNameRE
// accepts a leading underscore, so this cannot be left to the charset check.
func TestCheckCredRejectsReservedPrefix(t *testing.T) {
	run := func(envVar string) []Issue {
		d := &spec.Defaults{Security: spec.Security{Users: []spec.User{{Name: "ops", PasswordEnv: envVar}}}}
		e, _ := Run(Context{Workflows: wfOK(), Defaults: d, Env: func(string) (string, bool) { return "v", true }})
		return e
	}
	want := "is reserved for the mount names this tool derives for itself"
	if e := run(spec.GeneratedNamePrefix + "SECURITY_USER_OPS_PASSWORD"); !hasErr(e, want) {
		t.Errorf("an -env inside the reserved prefix must be rejected, got %v", e)
	}
	// The prefix is rejected wherever it starts the name, not only on an exact
	// derived-name match -- the whole namespace is reserved.
	if e := run(spec.GeneratedNamePrefix + "ANYTHING"); !hasErr(e, want) {
		t.Errorf("the whole prefix namespace must be reserved, got %v", e)
	}
	// A name that merely contains it, or starts with a bare underscore, is fine.
	for _, ok := range []string{"MY_GEN_PASSWORD", "_MY_PASSWORD", "SOL_PASSWORD"} {
		if e := run(ok); hasErr(e, want) {
			t.Errorf("%q is outside the reserved prefix and must be accepted, got %v", ok, e)
		}
	}
}

// TestCheckKubeSecretsCreateXorExisting pins the create/existing exclusivity for
// both Secret kinds, the same rule libs.pvc already enforces. Both set is the
// dangerous one: deploy.Render takes the create branch, so it would emit a Secret
// doc over the very object 'existing' names. Neither set leaves the secrets
// mount out of the pod with nothing to say why.
func TestCheckKubeSecretsCreateXorExisting(t *testing.T) {
	base := spec.Deployment{Name: "c", Namespace: "ns", Replicas: 1}
	run := func(sec spec.Secrets) []Issue {
		k := &spec.Kubernetes{Deployment: base, Secrets: sec}
		e, _ := Run(Context{Workflows: wfOK(), Defaults: defsWithStores(), Image: imageOK(), Kube: k, CheckKubernetes: true, Env: func(string) (string, bool) { return "v", true }})
		return e
	}
	cases := []struct {
		name string
		sec  spec.Secrets
		want string
	}{
		{"cred both", spec.Secrets{Credentials: &spec.CredentialsSecret{Create: &spec.CredCreate{Name: "a"}, Existing: "b"}},
			"kubernetes.secrets.credentials must set exactly one of 'create' or 'existing'"},
		{"cred neither", spec.Secrets{Credentials: &spec.CredentialsSecret{}},
			"kubernetes.secrets.credentials must set exactly one of 'create' or 'existing'"},
		{"stores both", spec.Secrets{Stores: &spec.StoresSecret{Create: &spec.StoreCreate{Name: "a"}, Existing: "b"}},
			"kubernetes.secrets.stores must set exactly one of 'create' or 'existing'"},
		{"stores neither", spec.Secrets{Stores: &spec.StoresSecret{}},
			"kubernetes.secrets.stores must set exactly one of 'create' or 'existing'"},
	}
	for _, c := range cases {
		if e := run(c.sec); !hasErr(e, c.want) {
			t.Errorf("%s: want %q, got %v", c.name, c.want, e)
		}
	}

	// Exactly one set, either way round, passes for both kinds.
	ok := []struct {
		name string
		sec  spec.Secrets
	}{
		{"create only", spec.Secrets{
			Credentials: &spec.CredentialsSecret{Create: &spec.CredCreate{Name: "solmq-credentials"}},
			Stores:      &spec.StoresSecret{Create: &spec.StoreCreate{Name: "solmq-tls"}},
		}},
		{"existing only", spec.Secrets{
			Credentials: &spec.CredentialsSecret{Existing: "their-creds"},
			Stores:      &spec.StoresSecret{Existing: "their-tls"},
		}},
		// Omitting the blocks entirely stays legal: that is the documented "no
		// Secret is produced" path, not an undecided block.
		{"blocks omitted", spec.Secrets{}},
	}
	for _, c := range ok {
		if e := run(c.sec); hasErr(e, "exactly one of 'create' or 'existing'") {
			t.Errorf("%s: should pass the XOR check, got %v", c.name, e)
		}
	}
}

func TestCheckLibsNFSFields(t *testing.T) {
	// nfs.server/path land unquoted in the PersistentVolume manifest piped to
	// kubectl, so a newline (which would inject a sibling key) is rejected.
	base := spec.Deployment{Name: "c", Namespace: "ns", Replicas: 1}
	kube := func(server, path string) *spec.Kubernetes {
		return &spec.Kubernetes{Deployment: base, Libs: &spec.Libs{PVC: &spec.LibsPVC{
			Create: &spec.PVCCreate{Name: "libs-pvc", Storage: "1Gi", NFS: spec.NFS{Server: server, Path: path}},
		}}}
	}
	if e, _ := Run(Context{Workflows: wfOK(), Defaults: &spec.Defaults{}, Image: imageOK(), Kube: kube("nfs1.corp\nreadOnly: false", "/libs"), CheckKubernetes: true}); !hasErr(e, "nfs.server") {
		t.Errorf("want nfs.server rejection, got %v", e)
	}
	if e, _ := Run(Context{Workflows: wfOK(), Defaults: &spec.Defaults{}, Image: imageOK(), Kube: kube("nfs1.corp.example", "/libs\nreadOnly: false"), CheckKubernetes: true}); !hasErr(e, "nfs.path") {
		t.Errorf("want nfs.path rejection, got %v", e)
	}
	if e, _ := Run(Context{Workflows: wfOK(), Defaults: &spec.Defaults{}, Image: imageOK(), Kube: kube("nfs1.corp.example", "/solace-libs"), CheckKubernetes: true}); hasErr(e, "nfs.") {
		t.Errorf("valid nfs server/path should pass, got %v", e)
	}
}

func TestPasswordConflictOnSameBinder(t *testing.T) {
	// Two files declaring the same connection tuple with different passwords
	// collapse into one binder, so only the first password would survive. That is
	// an auth mismatch, not a preference -- it is fatal.
	a := vMQ("Q1", spec.DestQueue, false)
	b := vMQ("Q2", spec.DestQueue, false)
	b.PasswordEnv = "OTHER_MQ_PASSWORD"
	wfs := []spec.Workflow{
		wf("0.yaml", vSolace("S1", spec.DestQueue, ""), a),
		wf("1.yaml", vSolace("S2", spec.DestQueue, ""), b),
	}
	e, _ := Run(Context{Workflows: wfs, Defaults: &spec.Defaults{}})
	if !hasErr(e, "conflicting password for the same binder") {
		t.Errorf("want password-conflict error, got %v", e)
	}
	// The same tuple with the same password is the normal shared-binder case.
	same := []spec.Workflow{
		wf("0.yaml", vSolace("S1", spec.DestQueue, ""), vMQ("Q1", spec.DestQueue, false)),
		wf("1.yaml", vSolace("S2", spec.DestQueue, ""), vMQ("Q2", spec.DestQueue, false)),
	}
	if e, _ := Run(Context{Workflows: same, Defaults: defsWithStores()}); hasErr(e, "conflicting password") {
		t.Errorf("identical passwords on one binder should pass, got %v", e)
	}
	// Distinct tuples (different queue-manager) are distinct binders, so their
	// passwords are free to differ.
	other := vMQ("Q2", spec.DestQueue, false)
	other.QueueManager = "QM2"
	other.PasswordEnv = "OTHER_MQ_PASSWORD"
	distinct := []spec.Workflow{
		wf("0.yaml", vSolace("S1", spec.DestQueue, ""), vMQ("Q1", spec.DestQueue, false)),
		wf("1.yaml", vSolace("S2", spec.DestQueue, ""), other),
	}
	if e, _ := Run(Context{Workflows: distinct, Defaults: defsWithStores()}); hasErr(e, "conflicting password") {
		t.Errorf("different binders may hold different passwords, got %v", e)
	}
}

// TestRemovedDefaultsKeysRejected pins checkRemovedDefaultsKeys: security.enabled
// and management.exposure are gone from the schema (management security can no
// longer be turned off; the actuator exposure list is always the fixed one),
// but both still parse a value through into spec.Defaults so a stale key from
// an old env.yaml fails loudly here rather than being silently ignored.
func TestRemovedDefaultsKeysRejected(t *testing.T) {
	for _, enabled := range []bool{true, false} {
		enabled := enabled
		d := defsWithStores()
		d.Security.Enabled = &enabled
		errs, _ := Run(Context{Workflows: wfOK(), Defaults: d})
		if !hasErr(errs, "security.enabled is no longer configurable") {
			t.Errorf("security.enabled=%v: want removed-key error, got %v", enabled, errs)
		}
	}

	for _, exposure := range []string{"health,info", ""} {
		exposure := exposure
		d := defsWithStores()
		d.Management.Exposure = &exposure
		errs, _ := Run(Context{Workflows: wfOK(), Defaults: d})
		if !hasErr(errs, "management.exposure is no longer configurable") {
			t.Errorf("management.exposure=%q: want removed-key error, got %v", exposure, errs)
		}
	}

	// Neither key set (the new default shape): no removed-key error.
	if errs, _ := Run(Context{Workflows: wfOK(), Defaults: defsWithStores()}); hasErr(errs, "no longer configurable") {
		t.Errorf("neither removed key set should validate clean, got %v", errs)
	}
}

// TestSecurityUserRoles pins checkSecurityUserRoles: an empty role and an
// unsafe-charset role are each rejected naming both indices, while a real
// authority name validates clean. Deliberately not an allowlist -- the
// connector owns the role vocabulary, so an unrecognized-but-well-formed name
// must pass rather than being second-guessed here.
func TestSecurityUserRoles(t *testing.T) {
	cases := []struct {
		name  string
		roles []string
		want  string // "" means expect no roles error
	}{
		{"admin is fine", []string{"admin"}, ""},
		{"unknown but well-formed role is not second-guessed", []string{"auditor"}, ""},
		{"several roles", []string{"admin", "auditor"}, ""},
		{"empty role", []string{""}, "security.users[0].roles[0] is empty"},
		{"whitespace-only role", []string{"   "}, "security.users[0].roles[0] is empty"},
		{"shell metacharacter", []string{"admin;rm"}, "contains an unsafe character"},
		{"embedded space", []string{"read only"}, "contains an unsafe character"},
	}
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			d := defsWithStores()
			d.Security.Users = []spec.User{{Name: "ops", Password: "p", Roles: c.roles}}
			errs, _ := Run(Context{Workflows: wfOK(), Defaults: d})
			if c.want == "" {
				if hasErr(errs, "roles") {
					t.Errorf("roles %v should validate clean, got %v", c.roles, errs)
				}
				return
			}
			if !hasErr(errs, c.want) {
				t.Errorf("roles %v: want error containing %q, got %v", c.roles, c.want, errs)
			}
		})
	}

	// Both texts are mirrored verbatim by the generator page's JS validator, so
	// pin them in full the way TestCheckDeployCommandErrorTexts does for the
	// deploy-command wording: the substring checks above would let the two
	// validators drift in the half they do not cover.
	dw := defsWithStores()
	dw.Security.Users = []spec.User{{Name: "ops", Password: "p", Roles: []string{"", "bad;role"}}}
	werrs, _ := Run(Context{Workflows: wfOK(), Defaults: dw})
	for _, want := range []string{
		`security.users[0].roles[0] is empty: give the role a name (for example "admin", which grants the read/write access needed to POST to /actuator/workflows) or drop the entry`,
		`security.users[0].roles[1] "bad;role" contains an unsafe character (` + UnsafeTokenReason + `); a role is an authority name passed to the connector verbatim`,
	} {
		if !hasErr(werrs, want) {
			t.Errorf("error wording drifted from the text the generator page mirrors:\n  want %s\n  got  %v", want, werrs)
		}
	}

	// No roles at all is the read-only default and must stay clean.
	d := defsWithStores()
	d.Security.Users = []spec.User{{Name: "ops", Password: "p"}}
	if errs, _ := Run(Context{Workflows: wfOK(), Defaults: d}); hasErr(errs, "roles") {
		t.Errorf("a user with no roles should validate clean, got %v", errs)
	}
}

// TestStatusUserReservedName pins checkStatusUser's name-collision rule: a
// configured security.users entry named spec.StatusUserName collides with the
// tool's own reserved read-only status account and is rejected, naming the
// index like the surrounding checks do. A differently-named user is fine.
func TestStatusUserReservedName(t *testing.T) {
	d := defsWithStores()
	d.Security.Users = []spec.User{{Name: "ops", Password: "p"}, {Name: spec.StatusUserName, Password: "p"}}
	errs, _ := Run(Context{Workflows: wfOK(), Defaults: d})
	n := 0
	for _, e := range errs {
		if strings.Contains(e.String(), spec.StatusUserName) && strings.Contains(e.String(), "reserved") {
			n++
		}
	}
	if n != 1 {
		t.Fatalf("want exactly one reserved-name error, got %d in %v", n, errs)
	}
	if !hasErr(errs, "security.users[1].name") {
		t.Errorf("want the error to name the index, got %v", errs)
	}

	// A differently-named user does not collide.
	d2 := defsWithStores()
	d2.Security.Users = []spec.User{{Name: "ops", Password: "p"}}
	if e, _ := Run(Context{Workflows: wfOK(), Defaults: d2}); hasErr(e, "reserved") {
		t.Errorf("a differently-named user should not trip the reserved-name check, got %v", e)
	}
}

// TestStatusUserPasswordEnvCharset pins checkStatusUser's charset gate on
// spec.StatusUserPasswordEnvVar: absent or empty is fine (the tool generates a
// password), a value using only printable ASCII outside the excluded set is
// fine, and each excluded character -- whitespace, both quote styles,
// backslash, a ${...} placeholder character, a control character, and a
// non-ASCII byte -- is rejected. The error text never echoes the value.
func TestStatusUserPasswordEnvCharset(t *testing.T) {
	const secret = "s3cr3t-do-not-log-me"
	run := func(env func(string) (string, bool)) []Issue {
		errs, _ := Run(Context{Workflows: wfOK(), Defaults: defsWithStores(), Env: env})
		return errs
	}

	if e := run(nil); hasErr(e, spec.StatusUserPasswordEnvVar) {
		t.Errorf("nil Env should not trip the password check, got %v", e)
	}
	if e := run(func(string) (string, bool) { return "", false }); hasErr(e, spec.StatusUserPasswordEnvVar) {
		t.Errorf("unset env var should be fine, got %v", e)
	}
	if e := run(func(string) (string, bool) { return "", true }); hasErr(e, spec.StatusUserPasswordEnvVar) {
		t.Errorf("empty env var should be fine, got %v", e)
	}
	if e := run(func(string) (string, bool) { return secret, true }); hasErr(e, spec.StatusUserPasswordEnvVar) {
		t.Errorf("a good value should be fine, got %v", e)
	}

	bad := []struct {
		name string
		val  string
	}{
		{"space", secret + " x"},
		{"double quote", secret + `"`},
		{"single quote", secret + "'"},
		{"backslash", secret + `\`},
		{"dollar-brace", secret + "${x}"},
		{"control char", secret + "\x01"},
		{"non-ASCII byte", secret + "\xff"},
	}
	for _, c := range bad {
		e := run(func(string) (string, bool) { return c.val, true })
		if !hasErr(e, spec.StatusUserPasswordEnvVar) {
			t.Errorf("%s: want a password-charset error, got %v", c.name, e)
		}
		for _, issue := range e {
			if strings.Contains(issue.String(), secret) {
				t.Errorf("%s: error text must never echo the secret value, got %q", c.name, issue.String())
			}
		}
	}
}

func TestPasswordConflictSolaceSide(t *testing.T) {
	// The Solace branch keys on client-password rather than password.
	a := vSolace("Q1", spec.DestQueue, "")
	b := vSolace("Q2", spec.DestQueue, "")
	b.ClientPassEnv = "OTHER_SOLACE_PASSWORD"
	wfs := []spec.Workflow{
		wf("0.yaml", a, vMQ("M1", spec.DestQueue, false)),
		wf("1.yaml", b, vMQ("M2", spec.DestQueue, false)),
	}
	if e, _ := Run(Context{Workflows: wfs, Defaults: defsWithStores()}); !hasErr(e, "conflicting password for the same binder") {
		t.Errorf("want solace password-conflict error, got %v", e)
	}
}

// TestCheckSyslogRunsForEveryPlatform is why checkSyslog moved out of checkKube:
// syslog is a top-level key now, so a docker-only or podman-only run has to
// validate it too. Under the old placement it went unchecked everywhere it had
// just started applying.
func TestCheckSyslogRunsForEveryPlatform(t *testing.T) {
	bad := &spec.Syslog{Host: "h", Port: 514, Protocol: "xxx"}
	for _, c := range []struct {
		name string
		ctx  Context
	}{
		{"docker", Context{Docker: dockerOK(), CheckDocker: true}},
		{"podman", Context{Podman: podmanOK(), CheckPodman: true}},
		{"config only", Context{}},
	} {
		t.Run(c.name, func(t *testing.T) {
			ctx := c.ctx
			ctx.Workflows = wfOK()
			ctx.Image = imageOK()
			ctx.Defaults = &spec.Defaults{Syslog: bad}
			errs, _ := Run(ctx)
			if !hasErr(errs, `must be "udp" or "tcp"`) {
				t.Errorf("syslog must be validated on %s too, got %v", c.name, errs)
			}
		})
	}
}

// TestKubernetesLoggingIsRetired pins the migration path. ParseEnv decodes
// non-strict, so a leftover kubernetes.logging block would otherwise be dropped
// in silence and the instance would come up with no syslog and no diagnostic.
func TestKubernetesLoggingIsRetired(t *testing.T) {
	k := &spec.Kubernetes{
		Deployment: baseKubeDeploy(),
		Command:    spec.DefaultKubeCommand,
		Service:    baseKubeService(),
		Logging:    &spec.Logging{Syslog: &spec.Syslog{Host: "h", Port: 514, Protocol: spec.SyslogUDP}},
	}
	errs, _ := Run(Context{Workflows: wfOK(), Defaults: &spec.Defaults{}, Image: imageOK(), Kube: k, CheckKubernetes: true})
	if !hasErr(errs, "kubernetes.logging is no longer configured here") {
		t.Fatalf("a leftover kubernetes.logging must be rejected, got %v", errs)
	}
	if !hasErr(errs, "top-level logging:") {
		t.Errorf("the error must name where it moved to, got %v", errs)
	}
}

// ---- the credentials Secret has to be wired if anything needs it ---------------

// kubeNoSecrets is a valid kubernetes section with no secrets block at all,
// which is the shape that used to deploy a pod with no secrets mount and say
// nothing about it.
func kubeNoSecrets() *spec.Kubernetes {
	return &spec.Kubernetes{Deployment: spec.Deployment{Name: "c", Namespace: "ns", Replicas: 1}}
}

// TestCredentialsNotWiredWarning is the gate for the failure that cost an
// afternoon on OpenShift: the config references credentials, the kubernetes
// section omits secrets.credentials, so nothing is mounted and every ${...}
// stays unresolved -- and because the configtree import is optional, the
// connector starts anyway and only fails later, against the broker.
func TestCredentialsNotWiredWarning(t *testing.T) {
	_, warns := Run(Context{
		Workflows: []spec.Workflow{wf("x.yaml", vSolacePlainTCP("Q", spec.DestQueue), vMQ("M", spec.DestQueue, false))},
		Defaults:  &spec.Defaults{}, Image: imageOK(), Kube: kubeNoSecrets(), CheckKubernetes: true,
	})
	if !hasErr(warns, "kubernetes.secrets.credentials is omitted") {
		t.Errorf("want credentials-omitted warning, got %v", warns)
	}
	// The message has to name where they were going to be mounted.
	if !hasErr(warns, secretsMountPath) {
		t.Errorf("the warning should name %s, got %v", secretsMountPath, warns)
	}
}

// TestCredentialsWiredNoWarning covers both ways of wiring it: neither should
// warn, and the create/existing distinction is not this check's business.
func TestCredentialsWiredNoWarning(t *testing.T) {
	for _, c := range []struct {
		name string
		sec  spec.Secrets
	}{
		{"create", spec.Secrets{Credentials: &spec.CredentialsSecret{Create: &spec.CredCreate{Name: "s"}}}},
		{"existing", spec.Secrets{Credentials: &spec.CredentialsSecret{Existing: "my-creds"}}},
	} {
		t.Run(c.name, func(t *testing.T) {
			k := kubeNoSecrets()
			k.Secrets = c.sec
			_, warns := Run(Context{
				Workflows: []spec.Workflow{wf("x.yaml", vSolacePlainTCP("Q", spec.DestQueue), vMQ("M", spec.DestQueue, false))},
				Defaults:  &spec.Defaults{}, Image: imageOK(), Kube: k, CheckKubernetes: true,
			})
			if hasErr(warns, "kubernetes.secrets.credentials is omitted") {
				t.Errorf("credentials wired via %s should not warn: %v", c.name, warns)
			}
		})
	}
}

// TestNoCredentialsNoWarning is the other half, and the reason this is a
// warning about a mismatch rather than a rule that the block is mandatory: a
// config whose connections need no authentication needs no Secret either, and
// must not be nagged about one.
func TestNoCredentialsNoWarning(t *testing.T) {
	anon := spec.Side{System: spec.SystemSolace, Host: "tcp://b:55555", MsgVPN: "prod", DestKind: spec.DestQueue, Dest: "Q"}
	mq := spec.Side{System: spec.SystemMQ, ConnName: "h(1414)", QueueManager: "QM", Channel: "CH", DestKind: spec.DestQueue, Dest: "M"}
	_, warns := Run(Context{
		Workflows: []spec.Workflow{wf("x.yaml", anon, mq)},
		Defaults:  &spec.Defaults{}, Image: imageOK(), Kube: kubeNoSecrets(), CheckKubernetes: true,
	})
	if hasErr(warns, "kubernetes.secrets.credentials is omitted") {
		t.Errorf("a config with no credentials must not be warned about a Secret it does not need: %v", warns)
	}
}

// TestCredentialsFoundOutsideAWorkflowSide walks the positions that are easy to
// forget: the management session, a management account, and a store password
// all mount through the same Secret as a binder credential does.
func TestCredentialsFoundOutsideAWorkflowSide(t *testing.T) {
	anon := spec.Side{System: spec.SystemSolace, Host: "tcp://b:55555", MsgVPN: "prod", DestKind: spec.DestQueue, Dest: "Q"}
	mq := spec.Side{System: spec.SystemMQ, ConnName: "h(1414)", QueueManager: "QM", Channel: "CH", DestKind: spec.DestQueue, Dest: "M"}
	wfs := []spec.Workflow{wf("x.yaml", anon, mq)}

	for _, c := range []struct {
		name string
		defs *spec.Defaults
	}{
		{"a management account", &spec.Defaults{Security: spec.Security{Users: []spec.User{{Name: "ops", PasswordEnv: "OPS_PASSWORD"}}}}},
		{"a truststore password", &spec.Defaults{TLS: spec.TLSConfig{Truststore: &spec.Store{File: "t.jks", PasswordEnv: "TRUSTSTORE_PASSWORD", Type: "JKS"}}}},
	} {
		t.Run(c.name, func(t *testing.T) {
			_, warns := Run(Context{Workflows: wfs, Defaults: c.defs, Image: imageOK(), Kube: kubeNoSecrets(), CheckKubernetes: true})
			if !hasErr(warns, "kubernetes.secrets.credentials is omitted") {
				t.Errorf("%s is a credential too, and must trip the warning: %v", c.name, warns)
			}
		})
	}
}

// TestLibsPVNameLengthIsCapped covers what namespacing the PersistentVolume
// name costs: the derived name is still a single DNS-1123 label, so a long
// namespace plus a long claim name can exceed 63 characters. Caught at the gate
// with both fields named, rather than by the API server part-way through an
// apply that has already created other objects.
func TestLibsPVNameLengthIsCapped(t *testing.T) {
	long := strings.Repeat("a", 40)
	k := &spec.Kubernetes{
		Deployment: spec.Deployment{Name: "c", Namespace: strings.Repeat("n", 40), Replicas: 1},
		Libs: &spec.Libs{PVC: &spec.LibsPVC{Create: &spec.PVCCreate{
			Name: long, Storage: "1Gi", NFS: spec.NFS{Server: "nfs1", Path: "/libs"},
		}}},
	}
	errs, _ := Run(Context{Workflows: wfOK(), Defaults: &spec.Defaults{}, Image: imageOK(), Kube: k, CheckKubernetes: true})
	if !hasErr(errs, "exceeds the 63-char DNS-1123 limit") {
		t.Errorf("an over-long derived PV name must be refused, got %v", errs)
	}

	// A name that fits is not flagged, so the check cannot simply always fire.
	k.Libs.PVC.Create.Name = "jar-libs"
	k.Deployment.Namespace = "solconnector-tps-sit"
	errs, _ = Run(Context{Workflows: wfOK(), Defaults: &spec.Defaults{}, Image: imageOK(), Kube: k, CheckKubernetes: true})
	if hasErr(errs, "exceeds the 63-char DNS-1123 limit") {
		t.Errorf("a name that fits must not be flagged, got %v", errs)
	}
}
