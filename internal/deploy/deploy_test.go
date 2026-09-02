package deploy

import (
	"strconv"
	"strings"
	"testing"

	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/consolidate"
	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/spec"
)

func baseKube() *spec.Kubernetes {
	return &spec.Kubernetes{
		Deployment: spec.Deployment{
			Name: "solmq", Namespace: "ns", Replicas: 1,
			Resources: spec.Resources{CPU: "1", Memory: "1Gi"},
		},
		Service: spec.Service{Enabled: true, Port: spec.Port{Host: 8090, Container: 8090}},
	}
}

// one builds a connector instance (the common test-fixture case).
func one(name, appYAML string, m *consolidate.Model) Instance {
	return Instance{Name: name, Image: "img:1", Timezone: "UTC", AppYAML: appYAML, StatusScript: "#!/bin/sh\necho status\n", Model: m}
}

// fullFixtureInput builds the everything-on fixture shared by TestRenderFull
// and TestRenderFull_ExactDocument so the two cannot drift.
func fullFixtureInput() Input {
	k := baseKube()
	k.Secrets = spec.Secrets{
		Credentials: &spec.CredentialsSecret{Create: &spec.CredCreate{Name: "creds"}},
		Stores:      &spec.StoresSecret{Create: &spec.StoreCreate{Name: "tls"}},
	}
	return Input{
		Kube: k, Defaults: &spec.Defaults{},
		CredKVs:  []KV{{Key: "A", Val: `sec"ret`}},
		Stores:   []StoreFile{{Name: "truststore.jks", Base64: "QUJD"}},
		Instance: one(k.Deployment.Name, "spring:\n  x: 1\n\n  y: 2\n", &consolidate.Model{MQTLS: true}),
	}
}

func TestRenderFull(t *testing.T) {
	out := Render(fullFixtureInput())
	for _, w := range []string{
		"kind: ConfigMap", "name: solmq-config", "application.yml: |", "    spring:",
		"status: |", "    #!/bin/sh", "    echo status",
		"kind: Secret", "name: creds", "stringData:", `A: "sec\"ret"`,
		"name: tls", "truststore.jks: QUJD",
		"kind: Deployment", "automountServiceAccountToken: false", "name: JAVA_TOOL_OPTIONS", "useIBMCipherMappings=false",
		"solace-connector/le-mode: standalone", "solace-connector/role: active",
		"- name: secrets", "mountPath: " + SecretsMountPath, "defaultMode: 0400",
		"mountPath: /app/external/classpath/truststores", "livenessProbe:", "tcpSocket:", "readinessProbe:",
		"mountPath: /app/external/.status-script", "subPath: status",
		"requests:", "limits:", `cpu: "1"`, "memory: 1Gi",
		"kind: Service", "targetPort: 8090",
	} {
		if !strings.Contains(out, w) {
			t.Errorf("missing %q\n---\n%s", w, out)
		}
	}
	if !strings.Contains(out, "\n\n") {
		t.Error("blank line in app yaml not preserved in block scalar")
	}
}

// TestRenderFull_ExactDocument is a full-document exact comparison for the
// TestRenderFull fixture. Unlike the strings.Contains checks above -- which
// pass even if a doc were duplicated, mis-indented, or a field landed in the
// wrong block -- this pins the entire byte-for-byte output.
func TestRenderFull_ExactDocument(t *testing.T) {
	got := Render(fullFixtureInput())
	if got != wantRenderFull {
		t.Errorf("Render mismatch\n%s", lineDiff(wantRenderFull, got))
	}
}

// wantRenderFull is the byte-for-byte expected output for the
// TestRenderFull_ExactDocument fixture (traced by hand over Render).
const wantRenderFull = `apiVersion: v1
kind: Namespace
metadata:
  name: ns
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: solmq-config
  namespace: ns
data:
  application.yml: |
    spring:
      x: 1

      y: 2
  status: |
    #!/bin/sh
    echo status
---
apiVersion: v1
kind: Secret
metadata:
  name: creds
  namespace: ns
type: Opaque
stringData:
  A: "sec\"ret"
---
apiVersion: v1
kind: Secret
metadata:
  name: tls
  namespace: ns
type: Opaque
data:
  truststore.jks: QUJD
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: solmq
  namespace: ns
spec:
  replicas: 1
  selector:
    matchLabels:
      app: solmq
  template:
    metadata:
      labels:
        app: solmq
        solace-connector/le-mode: standalone
        solace-connector/role: active
    spec:
      automountServiceAccountToken: false
      containers:
        - name: connector
          image: img:1
          ports:
            - name: management
              containerPort: 8090
          env:
            - name: TZ
              value: UTC
            - name: JAVA_TOOL_OPTIONS
              value: "-Dcom.ibm.mq.cfg.useIBMCipherMappings=false"
          volumeMounts:
            - name: secrets
              mountPath: /app/external/var/secrets
              readOnly: true
            - name: config
              mountPath: /app/external/spring/config/application.yml
              subPath: application.yml
              readOnly: true
            - name: stores
              mountPath: /app/external/classpath/truststores
              readOnly: true
            - name: config
              mountPath: /app/external/.status-script
              subPath: status
              readOnly: true
          livenessProbe:
            tcpSocket:
              port: 8090
            initialDelaySeconds: 30
            periodSeconds: 15
          readinessProbe:
            tcpSocket:
              port: 8090
            initialDelaySeconds: 15
            periodSeconds: 10
          resources:
            requests:
              cpu: "1"
              memory: 1Gi
            limits:
              cpu: "1"
              memory: 1Gi
      volumes:
        - name: config
          configMap:
            name: solmq-config
        - name: secrets
          secret:
            secretName: creds
            defaultMode: 0400
        - name: stores
          secret:
            secretName: tls
---
apiVersion: v1
kind: Service
metadata:
  name: solmq
  namespace: ns
spec:
  selector:
    app: solmq
  ports:
    - name: management
      port: 8090
      targetPort: 8090
`

// TestRenderConfigMapStatusScript checks the ConfigMap always carries the
// status script under its own key, alongside application.yml.
func TestRenderConfigMapStatusScript(t *testing.T) {
	k := baseKube()
	out := Render(Input{Kube: k, Defaults: &spec.Defaults{}, Instance: one(k.Deployment.Name, "x: 1\n", &consolidate.Model{})})
	for _, want := range []string{"status: |", "    #!/bin/sh", "    echo status"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q\n---\n%s", want, out)
		}
	}
}

// lineDiff returns a compact first-divergence report for two multi-line
// strings (pattern: internal/gen/golden_test.go's lineDiff).
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
			return "first diff at line " + strconv.Itoa(i+1) + ":\n  want: " + strconv.Quote(wv) + "\n  got:  " + strconv.Quote(gv)
		}
	}
	return "(strings differ only in length/trailing content)"
}

func TestRenderNoSecretsNoServiceNoTLS(t *testing.T) {
	k := baseKube()
	k.Service.Enabled = false
	out := Render(Input{Kube: k, Defaults: &spec.Defaults{}, Instance: one(k.Deployment.Name, "x: 1\n", &consolidate.Model{MQTLS: false})})
	for _, no := range []string{
		"JAVA_TOOL_OPTIONS", "- name: secrets", "mountPath: " + SecretsMountPath, "kind: Secret", "kind: Service", "- name: stores",
		"logback-spring.xml", "LOGGING_SYSLOG_HOST", "- name: libs", "initContainers:",
	} {
		if strings.Contains(out, no) {
			t.Errorf("unexpected %q:\n%s", no, out)
		}
	}
}

func TestRenderSyslogUDP(t *testing.T) {
	k := baseKube()
	sys := &spec.Syslog{Host: "sys.host", Port: 514, Protocol: spec.SyslogUDP}
	out := Render(Input{Kube: k, Defaults: &spec.Defaults{}, Syslog: sys, Instance: one(k.Deployment.Name, "x: 1\n", &consolidate.Model{})})
	for _, want := range []string{
		"logback-spring.xml: |", "ch.qos.logback.classic.net.SyslogAppender",
		"- name: LOGGING_SYSLOG_APPNAME", "value: solmq",
		"- name: LOGGING_SYSLOG_HOST", "value: sys.host",
		"- name: LOGGING_SYSLOG_PORT", `value: "514"`,
		"mountPath: /app/external/classpath/logback-spring.xml", "subPath: logback-spring.xml",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q\n---\n%s", want, out)
		}
	}
	if strings.Contains(out, "LogstashTcpSocketAppender") {
		t.Error("udp must not use the TCP appender")
	}
}

func TestRenderSyslogTCP(t *testing.T) {
	k := baseKube()
	sys := &spec.Syslog{Host: "sys.host", Port: 6514, Protocol: spec.SyslogTCP}
	out := Render(Input{Kube: k, Defaults: &spec.Defaults{}, Syslog: sys, Instance: one(k.Deployment.Name, "x: 1\n", &consolidate.Model{})})
	for _, want := range []string{"LogstashTcpSocketAppender", "<destination>${SYSLOG_HOST}:${SYSLOG_PORT}</destination>"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q", want)
		}
	}
	for _, no := range []string{"ch.qos.logback.classic.net.SyslogAppender", "AsyncAppender"} {
		if strings.Contains(out, no) {
			t.Errorf("unexpected %q in tcp variant", no)
		}
	}
}

func TestRenderLibsPVCExisting(t *testing.T) {
	k := baseKube()
	k.Libs = &spec.Libs{PVC: &spec.LibsPVC{Existing: "my-pvc"}}
	out := Render(Input{Kube: k, Defaults: &spec.Defaults{}, Instance: one(k.Deployment.Name, "x: 1\n", &consolidate.Model{})})
	for _, want := range []string{"claimName: my-pvc", "mountPath: /app/external/libs"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q\n%s", want, out)
		}
	}
	for _, no := range []string{"initContainers:", "kind: PersistentVolume"} {
		if strings.Contains(out, no) {
			t.Errorf("unexpected %q", no)
		}
	}
}

func TestRenderLibsPVCCreate(t *testing.T) {
	k := baseKube()
	k.Libs = &spec.Libs{PVC: &spec.LibsPVC{Create: &spec.PVCCreate{
		Name: "jar-libs", Storage: "2Gi", NFS: spec.NFS{Server: "nfs1", Path: "/libs"},
	}}}
	out := Render(Input{Kube: k, Defaults: &spec.Defaults{}, Instance: one(k.Deployment.Name, "x: 1\n", &consolidate.Model{})})
	for _, want := range []string{
		// The PV name carries the namespace; the claim's does not, because the
		// claim is already namespaced. Both must agree on volumeName.
		"kind: PersistentVolume", "name: " + spec.LibsPVName(k.Deployment.Namespace, "jar-libs"),
		"persistentVolumeReclaimPolicy: Retain",
		"server: nfs1", "path: /libs", "readOnly: true",
		"kind: PersistentVolumeClaim", `storageClassName: ""`,
		"volumeName: " + spec.LibsPVName(k.Deployment.Namespace, "jar-libs"),
		"storage: 2Gi", "claimName: jar-libs",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q\n%s", want, out)
		}
	}
	if strings.Index(out, "kind: PersistentVolume") > strings.Index(out, "kind: Deployment") {
		t.Error("PV/PVC must precede the Deployment")
	}
}

func TestRenderLibsDownload(t *testing.T) {
	k := baseKube()
	k.Libs = &spec.Libs{Download: &spec.LibsDownload{
		URLs: []string{"https://repo/a.jar", "https://repo/b.jar"}, Image: "busybox:1.37",
	}}
	out := Render(Input{Kube: k, Defaults: &spec.Defaults{}, Instance: one(k.Deployment.Name, "x: 1\n", &consolidate.Model{})})
	for _, want := range []string{
		"initContainers:", "- name: libs-download", "image: busybox:1.37",
		`wget -O '/libs/a.jar' 'https://repo/a.jar' && wget -O '/libs/b.jar' 'https://repo/b.jar'`,
		"emptyDir: {}", "mountPath: /app/external/libs",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q\n%s", want, out)
		}
	}
	// download into an existing PVC instead of emptyDir
	k.Libs.Download.PVC = "dl-pvc"
	out = Render(Input{Kube: k, Defaults: &spec.Defaults{}, Instance: one(k.Deployment.Name, "x: 1\n", &consolidate.Model{})})
	if !strings.Contains(out, "claimName: dl-pvc") || strings.Contains(out, "emptyDir: {}") {
		t.Errorf("download pvc volume wrong:\n%s", out)
	}
}

// TestRenderStatusScriptMountAfterLibs pins the ordering that keeps the
// single-file status mount from being shadowed by the libs directory mount:
// the status mount must be declared after it.
func TestRenderStatusScriptMountAfterLibs(t *testing.T) {
	k := baseKube()
	k.Libs = &spec.Libs{PVC: &spec.LibsPVC{Existing: "my-pvc"}}
	out := Render(Input{Kube: k, Defaults: &spec.Defaults{}, Instance: one(k.Deployment.Name, "x: 1\n", &consolidate.Model{})})
	libsIdx := strings.Index(out, "mountPath: /app/external/libs\n")
	statusIdx := strings.Index(out, "mountPath: /app/external/.status-script")
	if libsIdx == -1 || statusIdx == -1 || statusIdx < libsIdx {
		t.Errorf("status mount must come after the libs directory mount:\n%s", out)
	}
	if !strings.Contains(out, "subPath: status") {
		t.Errorf("missing subPath: status\n%s", out)
	}
}

func TestRenderNamespaceAlwaysFirst(t *testing.T) {
	k := baseKube()
	out := Render(Input{Kube: k, Defaults: &spec.Defaults{}, Instance: one(k.Deployment.Name, "x: 1\n", &consolidate.Model{})})
	// The Namespace doc is always emitted first, before the ConfigMap, and names
	// the deployment namespace.
	nsDoc := "apiVersion: v1\nkind: Namespace\nmetadata:\n  name: ns\n"
	if !strings.HasPrefix(out, nsDoc) {
		t.Errorf("expected Namespace as the first doc, got:\n%s", out)
	}
	if strings.Index(out, "kind: Namespace") > strings.Index(out, "kind: ConfigMap") {
		t.Error("Namespace must precede the ConfigMap")
	}
}

func TestRenderExistingSecrets(t *testing.T) {
	k := baseKube()
	k.Secrets = spec.Secrets{
		Credentials: &spec.CredentialsSecret{Existing: "my-creds"},
		Stores:      &spec.StoresSecret{Existing: "my-tls"},
	}
	out := Render(Input{Kube: k, Defaults: &spec.Defaults{}, Instance: one(k.Deployment.Name, "x: 1\n", &consolidate.Model{})})
	if strings.Contains(out, "kind: Secret") {
		t.Error("existing secrets must not emit Secret docs")
	}
	if !strings.Contains(out, "secretName: my-creds") || !strings.Contains(out, "secretName: my-tls") {
		t.Errorf("existing refs missing:\n%s", out)
	}
}

// TestManagementPort pins that ManagementPort is exactly
// Defaults.EffectiveManagementPort: Kube.Service.Port is a targetPort mapping,
// not a management-port fallback (spec resolves an unset service.port to the
// management port itself, before Render/ManagementPort ever see it -- see
// internal/spec's TestPortDefaultsFollowManagementPort).
func TestManagementPort(t *testing.T) {
	if got := ManagementPort(Input{Kube: baseKube(), Defaults: &spec.Defaults{Management: spec.Management{Port: 9999}}}); got != 9999 {
		t.Errorf("mgmt = %d want 9999", got)
	}
	if got := ManagementPort(Input{Kube: baseKube(), Defaults: &spec.Defaults{}}); got != 8090 {
		t.Errorf("mgmt = %d want 8090 (default)", got)
	}
	if got := ManagementPort(Input{Kube: baseKube(), Defaults: nil}); got != 8090 {
		t.Errorf("mgmt = %d want 8090 (nil Defaults)", got)
	}
}

// TestRenderLeaderModeLabels covers the pod-template label per mode: standalone
// and active_active are statically "active", active_standby is not (only the
// actuator knows the live role there), and a nil Defaults behaves as standalone.
func TestRenderLeaderModeLabels(t *testing.T) {
	cases := []struct {
		name     string
		defs     *spec.Defaults
		mode     string
		wantRole bool
	}{
		{"nil Defaults behaves as standalone", nil, spec.LeaderStandalone, true},
		{"standalone", &spec.Defaults{}, spec.LeaderStandalone, true},
		{"active_active", &spec.Defaults{LeaderElection: spec.LeaderElection{Present: true, Mode: spec.LeaderActiveActive}}, spec.LeaderActiveActive, true},
		{"active_standby", &spec.Defaults{LeaderElection: spec.LeaderElection{Present: true, Mode: spec.LeaderActiveStby}}, spec.LeaderActiveStby, false},
	}
	for _, c := range cases {
		out := Render(Input{Kube: baseKube(), Defaults: c.defs, Instance: one("solmq", "x: 1\n", &consolidate.Model{})})
		wantMode := "        " + spec.LabelModeKey + ": " + c.mode
		if !strings.Contains(out, wantMode) {
			t.Errorf("%s: missing %q\n%s", c.name, wantMode, out)
		}
		roleLine := "        " + spec.LabelRoleKey + ": " + spec.LabelRoleActive
		if has := strings.Contains(out, roleLine); has != c.wantRole {
			t.Errorf("%s: role label present=%v want %v\n%s", c.name, has, c.wantRole, out)
		}
	}
}

// TestRenderSelectorMatchLabelsAppOnly pins that spec.selector.matchLabels
// never grows the mode/role labels even when the pod template carries them:
// a selector must stay immutable for the life of the Deployment, and both
// labels above can change with configuration or the actuator-reported leader.
func TestRenderSelectorMatchLabelsAppOnly(t *testing.T) {
	k := baseKube()
	defs := &spec.Defaults{LeaderElection: spec.LeaderElection{Present: true, Mode: spec.LeaderActiveActive}}
	out := Render(Input{Kube: k, Defaults: defs, Instance: one(k.Deployment.Name, "x: 1\n", &consolidate.Model{})})
	want := "  selector:\n    matchLabels:\n      app: solmq\n  template:"
	if !strings.Contains(out, want) {
		t.Errorf("selector must be app-only:\n%s", out)
	}
}

func TestQuoteRes(t *testing.T) {
	if quoteRes("1") != `"1"` || quoteRes("250m") != "250m" || quoteRes("512Mi") != "512Mi" {
		t.Errorf(`quoteRes: 1=>%q 250m=>%q`, quoteRes("1"), quoteRes("250m"))
	}
}

func TestRenderNoResources(t *testing.T) {
	k := baseKube()
	k.Deployment.Resources = spec.Resources{}
	if strings.Contains(Render(Input{Kube: k, Defaults: &spec.Defaults{}, Instance: one(k.Deployment.Name, "x\n", &consolidate.Model{})}), "resources:") {
		t.Error("no resources block expected when all empty")
	}
}

// TestRenderServicePort pins that Render renders whatever spec.Port it is
// given verbatim (port: Host, targetPort: Container), with no fallback logic
// of its own. An unset service.port is resolved by spec.applyKubeDefaults
// before Render ever sees it (mgmtPort:mgmtPort, including a non-default
// management.port), so that resolution is exercised here as an
// already-resolved Port, matching what production wiring (internal/gen builds
// Input from spec.Env.Kubernetes/Defaults) actually passes in.
func TestRenderServicePort(t *testing.T) {
	render := func(port spec.Port, defs *spec.Defaults) string {
		k := baseKube()
		k.Service = spec.Service{Enabled: true, Port: port}
		return Render(Input{Kube: k, Defaults: defs, Instance: one("solmq", "spring: {}\n", &consolidate.Model{})})
	}
	// The Service block is the only place both keys appear together, so matching
	// the pair pins the service port rather than the container/probe ports.
	svcBlock := func(port spec.Port) string {
		return "      port: " + strconv.Itoa(port.Host) + "\n      targetPort: " + strconv.Itoa(port.Container) + "\n"
	}
	cases := []struct {
		name string
		port spec.Port
		defs *spec.Defaults
	}{
		{"host:container distinct ports", spec.Port{Host: 8081, Container: 9000}, &spec.Defaults{}},
		{"resolved to the default management port", spec.Port{Host: 8090, Container: 8090}, &spec.Defaults{}},
		{"resolved to a non-default management port", spec.Port{Host: 9500, Container: 9500}, &spec.Defaults{Management: spec.Management{Port: 9500}}},
	}
	for _, c := range cases {
		out := render(c.port, c.defs)
		if want := svcBlock(c.port); !strings.Contains(out, want) {
			t.Errorf("%s: want\n%s\ngot\n%s", c.name, want, out)
		}
	}
}

// TestTeardownDropsTheNamespaceDocument pins half the reason Input.Teardown
// exists. A manifest carrying a Namespace, piped to `kubectl delete -f -`,
// deletes the namespace and cascades to every object inside it -- including
// workloads this tool never deployed. apply needs the document; delete must not
// have it.
func TestTeardownDropsTheNamespaceDocument(t *testing.T) {
	k := baseKube()
	inst := one(k.Deployment.Name, "x: 1\n", &consolidate.Model{})

	applyManifest := Render(Input{Kube: k, Defaults: &spec.Defaults{}, Instance: inst})
	if !strings.Contains(applyManifest, "kind: Namespace") {
		t.Errorf("apply still needs the Namespace document, got:\n%s", applyManifest)
	}

	deleteManifest := Render(Input{Kube: k, Defaults: &spec.Defaults{}, Instance: inst, Teardown: true})
	if strings.Contains(deleteManifest, "kind: Namespace") {
		t.Fatalf("a teardown manifest must carry no Namespace, got:\n%s", deleteManifest)
	}
	// Everything else is still there, and the document that was first is now the
	// ConfigMap -- so dropping doc 0 must not leave a leading separator.
	if strings.HasPrefix(deleteManifest, "---") {
		t.Errorf("omitting the first document must not leave a leading separator, got:\n%s", deleteManifest)
	}
	for _, want := range []string{"kind: ConfigMap", "kind: Deployment"} {
		if !strings.Contains(deleteManifest, want) {
			t.Errorf("teardown manifest is missing %q, got:\n%s", want, deleteManifest)
		}
	}
}

// TestNamespaceManifestMatchesWhatRenderEmits keeps the standalone document and
// the one inside Render from drifting: the namespace delete pipes the former
// through the same path the full manifest uses.
func TestNamespaceManifestMatchesWhatRenderEmits(t *testing.T) {
	k := baseKube()
	full := Render(Input{Kube: k, Defaults: &spec.Defaults{}, Instance: one(k.Deployment.Name, "x: 1\n", &consolidate.Model{})})
	solo := NamespaceManifest(k.Deployment.Namespace)
	if !strings.HasPrefix(full, solo) {
		t.Errorf("Render should start with exactly NamespaceManifest, got:\n%s\nwant prefix:\n%s", full, solo)
	}
}

// TestTeardownReversesTheDocumentOrder pins the other half of Input.Teardown,
// and the bug behind it: kubectl deletes documents in file order and waits for
// each one to actually go. In creation order the PVC is emitted before the
// Deployment, so a teardown blocks forever -- kubernetes.io/pvc-protection
// holds the claim while a pod still mounts it, and the Deployment that owns
// that pod is queued behind it. Reversed, the workload goes first.
func TestTeardownReversesTheDocumentOrder(t *testing.T) {
	k := baseKube()
	k.Libs = &spec.Libs{PVC: &spec.LibsPVC{Create: &spec.PVCCreate{
		Name: "jar-libs", Storage: "2Gi", NFS: spec.NFS{Server: "nfs1", Path: "/libs"},
	}}}
	inst := one(k.Deployment.Name, "x: 1\n", &consolidate.Model{})

	apply := Render(Input{Kube: k, Defaults: &spec.Defaults{}, Instance: inst})
	if !before(apply, "kind: PersistentVolumeClaim", "kind: Deployment") {
		t.Fatalf("apply must create the claim before the workload that mounts it:\n%s", apply)
	}

	del := Render(Input{Kube: k, Defaults: &spec.Defaults{}, Instance: inst, Teardown: true})
	if !before(del, "kind: Deployment", "kind: PersistentVolumeClaim") {
		t.Errorf("a teardown must delete the workload before the claim it holds:\n%s", del)
	}
	// The whole set is reversed, not just those two: the ConfigMap led the apply
	// and must trail the teardown.
	if !before(del, "kind: Service", "kind: ConfigMap") {
		t.Errorf("the teardown set must be fully reversed:\n%s", del)
	}
	// Reversing must not disturb the separators.
	if strings.HasPrefix(del, "---") || strings.Contains(del, "------") {
		t.Errorf("document separators are malformed:\n%s", del)
	}
	if got, want := strings.Count(del, "\n---\n"), strings.Count(apply, "\n---\n")-1; got != want {
		t.Errorf("separator count = %d, want %d (apply minus the dropped Namespace)", got, want)
	}
}

// before reports whether a appears ahead of b in s. Both must be present.
func before(s, a, b string) bool {
	i, j := strings.Index(s, a), strings.Index(s, b)
	return i >= 0 && j >= 0 && i < j
}

// TestLibsPVNameIsNamespaced is the root-cause test: a PersistentVolume is
// cluster-scoped, so two releases that pick the same libs.pvc.create.name in
// different namespaces must not derive the same PV name. They did, and the
// second apply rebound the volume out from under the first.
func TestLibsPVNameIsNamespaced(t *testing.T) {
	a := spec.LibsPVName("team-a", "jar-libs-pvc")
	b := spec.LibsPVName("team-b", "jar-libs-pvc")
	if a == b {
		t.Fatalf("two namespaces must not share a PV name, both got %q", a)
	}
	for _, n := range []string{a, b} {
		if !strings.HasSuffix(n, "-pv") {
			t.Errorf("%q should be suffixed -pv", n)
		}
	}
	if !strings.Contains(a, "team-a") {
		t.Errorf("%q should carry the namespace", a)
	}
}
