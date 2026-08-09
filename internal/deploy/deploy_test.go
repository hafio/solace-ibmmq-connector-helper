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
			Name: "solmq", Namespace: "ns", Image: "img:1", Replicas: 1,
			Resources: spec.Resources{CPU: "1", Memory: "1Gi"},
			Timezone:  "UTC",
		},
		Service: spec.Service{Enabled: true, Port: 8090},
	}
}

// one wraps a single connector instance (the common single-instance test case).
func one(name, appYAML string, m *consolidate.Model) []Instance {
	return []Instance{{Name: name, AppYAML: appYAML, Model: m}}
}

// fullFixtureInput builds the everything-on fixture shared by TestRenderFull
// and TestRenderFull_ExactDocument so the two cannot drift.
func fullFixtureInput() Input {
	k := baseKube()
	k.Secrets = spec.Secrets{
		Credentials: &spec.CredentialsSecret{Create: &spec.CredCreate{Name: "creds", Source: spec.SourceEnv, Variables: []string{"A"}}},
		Stores:      &spec.StoresSecret{Create: &spec.StoreCreate{Name: "tls"}},
	}
	return Input{
		Kube: k, Defaults: &spec.Defaults{},
		CredKVs:   []KV{{Key: "A", Val: `sec"ret`}},
		Stores:    []StoreFile{{Name: "truststore.jks", Base64: "QUJD"}},
		Instances: one(k.Deployment.Name, "spring:\n  x: 1\n\n  y: 2\n", &consolidate.Model{MQTLS: true}),
	}
}

func TestRenderFull(t *testing.T) {
	out := Render(fullFixtureInput())
	for _, w := range []string{
		"kind: ConfigMap", "name: solmq-config", "application.yml: |", "    spring:",
		"kind: Secret", "name: creds", "stringData:", `A: "sec\"ret"`,
		"name: tls", "truststore.jks: QUJD",
		"kind: Deployment", "name: JAVA_TOOL_OPTIONS", "useIBMCipherMappings=false",
		"envFrom:", "secretRef:",
		"mountPath: /app/external/classpath/truststores", "livenessProbe:", "tcpSocket:", "readinessProbe:",
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
    spec:
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
          envFrom:
            - secretRef:
                name: creds
          volumeMounts:
            - name: config
              mountPath: /app/external/spring/config/application.yml
              subPath: application.yml
              readOnly: true
            - name: stores
              mountPath: /app/external/classpath/truststores
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
	out := Render(Input{Kube: k, Defaults: &spec.Defaults{}, Instances: one(k.Deployment.Name, "x: 1\n", &consolidate.Model{MQTLS: false})})
	for _, no := range []string{
		"JAVA_TOOL_OPTIONS", "envFrom:", "kind: Secret", "kind: Service", "- name: stores",
		"logback-spring.xml", "LOGGING_SYSLOG_HOST", "- name: libs", "initContainers:",
	} {
		if strings.Contains(out, no) {
			t.Errorf("unexpected %q:\n%s", no, out)
		}
	}
}

func TestRenderSyslogUDP(t *testing.T) {
	k := baseKube()
	k.Logging = &spec.Logging{Syslog: &spec.Syslog{Host: "sys.host", Port: 514, Protocol: spec.SyslogUDP}}
	out := Render(Input{Kube: k, Defaults: &spec.Defaults{}, Instances: one(k.Deployment.Name, "x: 1\n", &consolidate.Model{})})
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
	k.Logging = &spec.Logging{Syslog: &spec.Syslog{Host: "sys.host", Port: 6514, Protocol: spec.SyslogTCP}}
	out := Render(Input{Kube: k, Defaults: &spec.Defaults{}, Instances: one(k.Deployment.Name, "x: 1\n", &consolidate.Model{})})
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
	out := Render(Input{Kube: k, Defaults: &spec.Defaults{}, Instances: one(k.Deployment.Name, "x: 1\n", &consolidate.Model{})})
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
	out := Render(Input{Kube: k, Defaults: &spec.Defaults{}, Instances: one(k.Deployment.Name, "x: 1\n", &consolidate.Model{})})
	for _, want := range []string{
		"kind: PersistentVolume", "name: jar-libs-pv", "server: nfs1", "path: /libs", "readOnly: true",
		"kind: PersistentVolumeClaim", `storageClassName: ""`, "volumeName: jar-libs-pv",
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
	out := Render(Input{Kube: k, Defaults: &spec.Defaults{}, Instances: one(k.Deployment.Name, "x: 1\n", &consolidate.Model{})})
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
	out = Render(Input{Kube: k, Defaults: &spec.Defaults{}, Instances: one(k.Deployment.Name, "x: 1\n", &consolidate.Model{})})
	if !strings.Contains(out, "claimName: dl-pvc") || strings.Contains(out, "emptyDir: {}") {
		t.Errorf("download pvc volume wrong:\n%s", out)
	}
}

func TestRenderNamespaceAlwaysFirst(t *testing.T) {
	k := baseKube()
	out := Render(Input{Kube: k, Defaults: &spec.Defaults{}, Instances: one(k.Deployment.Name, "x: 1\n", &consolidate.Model{})})
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

func TestRenderMultiInstance(t *testing.T) {
	k := baseKube()
	k.Secrets = spec.Secrets{
		Credentials: &spec.CredentialsSecret{Create: &spec.CredCreate{Name: "creds", Source: spec.SourceEnv, Variables: []string{"A"}}},
	}
	in := Input{
		Kube: k, Defaults: &spec.Defaults{},
		CredKVs: []KV{{Key: "A", Val: "v"}},
		Instances: []Instance{
			{Name: "solmq-1", AppYAML: "a: 1\n", Model: &consolidate.Model{}},
			{Name: "solmq-2", AppYAML: "b: 2\n", Model: &consolidate.Model{}},
		},
	}
	out := Render(in)
	// Shared docs exactly once.
	if got := strings.Count(out, "kind: Namespace"); got != 1 {
		t.Errorf("Namespace count = %d, want 1", got)
	}
	if got := strings.Count(out, "kind: Secret"); got != 1 {
		t.Errorf("Secret count = %d, want 1 (shared)", got)
	}
	// Per-instance docs: 2 ConfigMaps, 2 Deployments, 2 Services with -N names.
	if got := strings.Count(out, "kind: ConfigMap"); got != 2 {
		t.Errorf("ConfigMap count = %d, want 2", got)
	}
	if got := strings.Count(out, "kind: Deployment"); got != 2 {
		t.Errorf("Deployment count = %d, want 2", got)
	}
	if got := strings.Count(out, "kind: Service"); got != 2 {
		t.Errorf("Service count = %d, want 2", got)
	}
	for _, want := range []string{
		"name: solmq-1-config", "name: solmq-2-config",
		"name: solmq-1\n", "name: solmq-2\n",
		"app: solmq-1", "app: solmq-2",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q\n%s", want, out)
		}
	}
}

func TestRenderExistingSecrets(t *testing.T) {
	k := baseKube()
	k.Secrets = spec.Secrets{
		Credentials: &spec.CredentialsSecret{Existing: "my-creds"},
		Stores:      &spec.StoresSecret{Existing: "my-tls"},
	}
	out := Render(Input{Kube: k, Defaults: &spec.Defaults{}, Instances: one(k.Deployment.Name, "x: 1\n", &consolidate.Model{})})
	if strings.Contains(out, "kind: Secret") {
		t.Error("existing secrets must not emit Secret docs")
	}
	if !strings.Contains(out, "name: my-creds") || !strings.Contains(out, "secretName: my-tls") {
		t.Errorf("existing refs missing:\n%s", out)
	}
}

func TestManagementPortFallback(t *testing.T) {
	if got := managementPort(Input{Kube: baseKube(), Defaults: &spec.Defaults{Management: spec.Management{Present: true, Port: 9999}}}); got != 9999 {
		t.Errorf("mgmt = %d want 9999 (defaults)", got)
	}
	if got := managementPort(Input{Kube: baseKube(), Defaults: &spec.Defaults{}}); got != 8090 {
		t.Errorf("mgmt = %d want 8090 (service port)", got)
	}
	k := baseKube()
	k.Service.Port = 0
	if got := managementPort(Input{Kube: k, Defaults: nil}); got != 8090 {
		t.Errorf("mgmt = %d want 8090 (default)", got)
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
	if strings.Contains(Render(Input{Kube: k, Defaults: &spec.Defaults{}, Instances: one(k.Deployment.Name, "x\n", &consolidate.Model{})}), "resources:") {
		t.Error("no resources block expected when all empty")
	}
}

// TestBaseName pins baseName's current behavior, including that it does NOT
// normalize backslashes the way gen.baseName does (internal/gen/gen_extra_test.go's
// TestBaseNameB64ToIssues).
// That divergence is safe: baseName's only caller (renderDeployment) feeds it
// Libs.Download.URLs entries, and internal/validate's safeLibsURL rejects any
// libs.download url containing a backslash before GenerateKubernetes ever
// calls Render, so a Windows-style path can never reach it through the CLI.
func TestBaseName(t *testing.T) {
	if got := baseName("https://repo/a.jar"); got != "a.jar" {
		t.Errorf("baseName(url) = %q, want %q", got, "a.jar")
	}
	if got := baseName("noslash"); got != "noslash" {
		t.Errorf("baseName(no slash) = %q, want %q", got, "noslash")
	}
	if got := baseName(`a\b\c.jar`); got != `a\b\c.jar` {
		t.Errorf("baseName(backslash) = %q, want unchanged %q", got, `a\b\c.jar`)
	}
}

func TestSplitLines(t *testing.T) {
	if got := splitLines("a\nb\n"); len(got) != 2 || got[0] != "a" || got[1] != "b" {
		t.Errorf("splitLines = %v", got)
	}
	if got := splitLines(""); len(got) != 0 {
		t.Errorf("splitLines empty = %v", got)
	}
}
