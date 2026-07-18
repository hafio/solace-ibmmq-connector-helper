package deploy

import (
	"strings"
	"testing"

	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-gen/internal/consolidate"
	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-gen/internal/spec"
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

func TestRenderFull(t *testing.T) {
	k := baseKube()
	k.Secrets = spec.Secrets{
		Credentials: &spec.CredentialsSecret{Create: &spec.CredCreate{Name: "creds", Source: spec.SourceEnv, Variables: []string{"A"}}},
		Stores:      &spec.StoresSecret{Create: &spec.StoreCreate{Name: "tls"}},
	}
	in := Input{
		Kube: k, Defaults: &spec.Defaults{}, Model: &consolidate.Model{MQTLS: true},
		AppYAML: "spring:\n  x: 1\n\n  y: 2\n",
		CredKVs: []KV{{Key: "A", Val: `sec"ret`}},
		Stores:  []StoreFile{{Name: "truststore.jks", Base64: "QUJD"}},
	}
	out := Render(in)
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

func TestRenderNoSecretsNoServiceNoTLS(t *testing.T) {
	k := baseKube()
	k.Service.Enabled = false
	out := Render(Input{Kube: k, Defaults: &spec.Defaults{}, Model: &consolidate.Model{MQTLS: false}, AppYAML: "x: 1\n"})
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
	out := Render(Input{Kube: k, Defaults: &spec.Defaults{}, Model: &consolidate.Model{}, AppYAML: "x: 1\n"})
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
	out := Render(Input{Kube: k, Defaults: &spec.Defaults{}, Model: &consolidate.Model{}, AppYAML: "x: 1\n"})
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
	out := Render(Input{Kube: k, Defaults: &spec.Defaults{}, Model: &consolidate.Model{}, AppYAML: "x: 1\n"})
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
	out := Render(Input{Kube: k, Defaults: &spec.Defaults{}, Model: &consolidate.Model{}, AppYAML: "x: 1\n"})
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
	out := Render(Input{Kube: k, Defaults: &spec.Defaults{}, Model: &consolidate.Model{}, AppYAML: "x: 1\n"})
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
	out = Render(Input{Kube: k, Defaults: &spec.Defaults{}, Model: &consolidate.Model{}, AppYAML: "x: 1\n"})
	if !strings.Contains(out, "claimName: dl-pvc") || strings.Contains(out, "emptyDir: {}") {
		t.Errorf("download pvc volume wrong:\n%s", out)
	}
}

func TestRenderNamespaceAlwaysFirst(t *testing.T) {
	out := Render(Input{Kube: baseKube(), Defaults: &spec.Defaults{}, Model: &consolidate.Model{}, AppYAML: "x: 1\n"})
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
	out := Render(Input{Kube: k, Defaults: &spec.Defaults{}, Model: &consolidate.Model{}, AppYAML: "x: 1\n"})
	if strings.Contains(out, "kind: Secret") {
		t.Error("existing secrets must not emit Secret docs")
	}
	if !strings.Contains(out, "name: my-creds") || !strings.Contains(out, "secretName: my-tls") {
		t.Errorf("existing refs missing:\n%s", out)
	}
}

func TestManagementPortFallback(t *testing.T) {
	if got := managementPort(Input{Kube: baseKube(), Defaults: &spec.Defaults{Management: spec.Management{Present: true, Port: 9999}}, Model: &consolidate.Model{}}); got != 9999 {
		t.Errorf("mgmt = %d want 9999 (defaults)", got)
	}
	if got := managementPort(Input{Kube: baseKube(), Defaults: &spec.Defaults{}, Model: &consolidate.Model{}}); got != 8090 {
		t.Errorf("mgmt = %d want 8090 (service port)", got)
	}
	k := baseKube()
	k.Service.Port = 0
	if got := managementPort(Input{Kube: k, Defaults: nil, Model: &consolidate.Model{}}); got != 8090 {
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
	if strings.Contains(Render(Input{Kube: k, Defaults: &spec.Defaults{}, Model: &consolidate.Model{}, AppYAML: "x\n"}), "resources:") {
		t.Error("no resources block expected when all empty")
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
