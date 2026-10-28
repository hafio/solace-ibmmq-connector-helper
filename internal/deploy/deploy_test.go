package deploy

import (
	"strconv"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/consolidate"
	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/spec"
	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/statusscript"
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

// one builds a connector instance (the common test-fixture case), running as
// the default non-root identity.
func one(name, appYAML string, m *consolidate.Model) Instance {
	return Instance{Name: name, Image: "img:1", Timezone: "UTC", AppYAML: appYAML, StatusScript: "#!/bin/sh\necho status\n", Model: m, RunAs: spec.RunAsID{UID: 1000, GID: 0}}
}

// fullFixtureInput builds the everything-on fixture shared by TestRenderFull
// and TestRenderFull_ExactDocument so the two cannot drift.
func fullFixtureInput() Input {
	k := baseKube()
	k.Secrets = spec.Secrets{
		Credentials: &spec.CredentialsSecret{Create: &spec.CredCreate{}},
	}
	return Input{
		Kube: k, Defaults: &spec.Defaults{},
		CredKVs:  []KV{{Key: "A", Val: `sec"ret`}},
		Stores:   []StoreFile{{Name: "mq-ca", Base64: "QUJD"}},
		Instance: one(k.Deployment.Name, "spring:\n  x: 1\n\n  y: 2\n", &consolidate.Model{MQTLS: true, Stores: []consolidate.StoreUse{{Name: "mq-ca", File: "./mq-ca.p12", Trusted: true}}}),
	}
}

func TestRenderFull(t *testing.T) {
	out := Render(fullFixtureInput())
	for _, w := range []string{
		"kind: ConfigMap", "name: solmq-config", "application.yml: |", "    spring:",
		"status: |", "    #!/bin/sh", "    echo status",
		"kind: Secret", "name: solmq-credentials", "stringData:", `A: "sec\"ret"`,
		"name: solmq-stores", "mq-ca: QUJD",
		"kind: Deployment", "automountServiceAccountToken: false", "name: JAVA_TOOL_OPTIONS", "useIBMCipherMappings=false",
		"solace-connector/le-mode: standalone", "solace-connector/role: active",
		"- name: secrets", "mountPath: " + SecretsMountPath, "defaultMode: 0400",
		"mountPath: /app/external/classpath/truststores", "livenessProbe:", "tcpSocket:", "readinessProbe:", "exec:",
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
  name: solmq-credentials
  namespace: ns
type: Opaque
stringData:
  A: "sec\"ret"
---
apiVersion: v1
kind: Secret
metadata:
  name: solmq-stores
  namespace: ns
type: Opaque
data:
  mq-ca: QUJD
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
      securityContext:
        runAsUser: 1000
        runAsGroup: 0
        fsGroup: 0
        runAsNonRoot: true
        seccompProfile:
          type: RuntimeDefault
      containers:
        - name: connector
          image: img:1
          securityContext:
            allowPrivilegeEscalation: false
            privileged: false
            capabilities:
              drop:
                - ALL
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
            exec:
              command:
                - sh
                - /app/external/.status-script
                - --health
            initialDelaySeconds: 15
            periodSeconds: 10
            timeoutSeconds: 5
            failureThreshold: 3
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
            secretName: solmq-credentials
            defaultMode: 0400
        - name: stores
          projected:
            defaultMode: 0400
            sources:
              - secret:
                  name: solmq-stores
                  items:
                    - key: "mq-ca"
                      path: mq-ca
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

// The pod carries the node selector the operator set: keys sorted, so an
// unchanged selector renders the same manifest every run, and a key or value
// YAML would read as anything but a string quoted. None set writes none.
func TestRenderDeploymentNodeSelector(t *testing.T) {
	k := baseKube()
	k.Deployment.NodeSelector = map[string]string{"zone": "a", "kubernetes.io/os": "linux", "gpu": "true", "spare": "", "2": "x"}
	out := Render(Input{Kube: k, Defaults: &spec.Defaults{}, Instance: one(k.Deployment.Name, "x: 1\n", &consolidate.Model{})})
	want := "      automountServiceAccountToken: false\n" +
		"      nodeSelector:\n" +
		"        \"2\": x\n" +
		"        gpu: \"true\"\n" +
		"        kubernetes.io/os: linux\n" +
		"        spare: \"\"\n" +
		"        zone: a\n" +
		"      securityContext:\n"
	if !strings.Contains(out, want) {
		t.Errorf("want the sorted, quoted nodeSelector in the pod spec\n%s\n---\n%s", want, out)
	}
	k.Deployment.NodeSelector = nil
	if out := Render(Input{Kube: k, Defaults: &spec.Defaults{}, Instance: one(k.Deployment.Name, "x: 1\n", &consolidate.Model{})}); strings.Contains(out, "nodeSelector") {
		t.Errorf("no node selector set, so none may be written\n%s", out)
	}
}

// Tolerations are written in Kubernetes' shape, each field only when set,
// after the node selector and before the pod's securityContext.
func TestRenderDeploymentTolerations(t *testing.T) {
	k := baseKube()
	secs := int64(300)
	k.Deployment.Tolerations = []spec.Toleration{
		{Key: "dedicated", Operator: "Equal", Value: "true", Effect: "NoSchedule"},
		{Operator: "Exists", Effect: "NoExecute", TolerationSeconds: &secs},
	}
	out := Render(Input{Kube: k, Defaults: &spec.Defaults{}, Instance: one(k.Deployment.Name, "x: 1\n", &consolidate.Model{})})
	want := "      tolerations:\n" +
		"        - key: dedicated\n" +
		"          operator: Equal\n" +
		"          value: \"true\"\n" +
		"          effect: NoSchedule\n" +
		"        - operator: Exists\n" +
		"          effect: NoExecute\n" +
		"          tolerationSeconds: 300\n" +
		"      securityContext:\n"
	if !strings.Contains(out, want) {
		t.Errorf("want the tolerations in the pod spec\n%s\n---\n%s", want, out)
	}
}

// The affinity block is written as it stands, then spread's anti-affinity on
// kubernetes.io/hostname against the Deployment's own app label: preferred
// with weight 100, or required. spread: none and no affinity write nothing.
func TestRenderDeploymentAffinityAndSpread(t *testing.T) {
	var doc yaml.Node
	if err := yaml.Unmarshal([]byte("nodeAffinity:\n  requiredDuringSchedulingIgnoredDuringExecution:\n    nodeSelectorTerms:\n      - matchExpressions:\n          - key: kubernetes.io/arch\n            operator: In\n            values:\n              - amd64\n"), &doc); err != nil {
		t.Fatal(err)
	}
	render := func(spread string, aff *yaml.Node) string {
		k := baseKube()
		k.Deployment.Spread = spread
		if aff != nil {
			k.Deployment.Affinity = *aff
		}
		return Render(Input{Kube: k, Defaults: &spec.Defaults{}, Instance: one(k.Deployment.Name, "x: 1\n", &consolidate.Model{})})
	}
	verbatim := "      affinity:\n" +
		"        nodeAffinity:\n" +
		"          requiredDuringSchedulingIgnoredDuringExecution:\n" +
		"            nodeSelectorTerms:\n" +
		"              -\n" +
		"                matchExpressions:\n" +
		"                  -\n" +
		"                    key: kubernetes.io/arch\n" +
		"                    operator: In\n" +
		"                    values:\n" +
		"                      - amd64\n"
	preferred := "        podAntiAffinity:\n" +
		"          preferredDuringSchedulingIgnoredDuringExecution:\n" +
		"            - weight: 100\n" +
		"              podAffinityTerm:\n" +
		"                topologyKey: kubernetes.io/hostname\n" +
		"                labelSelector:\n" +
		"                  matchLabels:\n" +
		"                    app: solmq\n"
	if out := render(spec.SpreadPreferred, doc.Content[0]); !strings.Contains(out, verbatim+preferred+"      securityContext:\n") {
		t.Errorf("want the verbatim affinity, then the preferred spread\n---\n%s", out)
	}
	required := "      affinity:\n" +
		"        podAntiAffinity:\n" +
		"          requiredDuringSchedulingIgnoredDuringExecution:\n" +
		"            - topologyKey: kubernetes.io/hostname\n" +
		"              labelSelector:\n" +
		"                matchLabels:\n" +
		"                  app: solmq\n"
	if out := render(spec.SpreadRequired, nil); !strings.Contains(out, required) {
		t.Errorf("want the required spread alone\n---\n%s", out)
	}
	for _, spread := range []string{"", spec.SpreadNone} {
		if out := render(spread, nil); strings.Contains(out, "affinity:") {
			t.Errorf("spread %q and no affinity must write none\n%s", spread, out)
		}
	}
}

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

// TestRenderJavaOptionsEnv covers the java-options: block on kubernetes: both
// variables become container env entries, the MQ TLS flag stays first in
// JAVA_TOOL_OPTIONS, and the operator's options alone are enough to open the
// env: list when there is no timezone, TLS or syslog to put in it.
func TestRenderJavaOptionsEnv(t *testing.T) {
	k := baseKube()
	inst := one(k.Deployment.Name, "x: 1\n", &consolidate.Model{MQTLS: true})
	inst.JavaOptions = &spec.JavaOptions{Tool: "-Xms512m\n-Xmx1g", JDK: "-Duser.language=en"}
	out := Render(Input{Kube: k, Defaults: &spec.Defaults{}, Instance: inst})
	want := "            - name: JAVA_TOOL_OPTIONS\n" +
		`              value: "-Dcom.ibm.mq.cfg.useIBMCipherMappings=false -Xms512m -Xmx1g"` + "\n" +
		"            - name: JDK_JAVA_OPTIONS\n" +
		`              value: "-Duser.language=en"` + "\n"
	if !strings.Contains(out, want) {
		t.Errorf("env entries missing or wrong, want:\n%s\ngot:\n%s", want, out)
	}

	bare := one(k.Deployment.Name, "x: 1\n", &consolidate.Model{})
	bare.Timezone = ""
	bare.JavaOptions = &spec.JavaOptions{JDK: "-Duser.language=en"}
	out = Render(Input{Kube: k, Defaults: &spec.Defaults{}, Instance: bare})
	if !strings.Contains(out, "          env:\n            - name: JDK_JAVA_OPTIONS\n") {
		t.Errorf("java-options alone must open the env: list:\n%s", out)
	}
	if strings.Contains(out, "JAVA_TOOL_OPTIONS") {
		t.Errorf("JAVA_TOOL_OPTIONS must be absent with no TLS and no tool options:\n%s", out)
	}
}

// TestEnvValueQuoteEscapes pins the render-side half of the java-options gate:
// validate refuses these characters, but a value that reached the renderer
// anyway must still reach the container as written -- '$' doubled, since
// kubernetes expands $(VAR) inside env values and reads $$ as one '$', and
// quotes and backslashes escaped for YAML.
func TestEnvValueQuoteEscapes(t *testing.T) {
	for in, want := range map[string]string{
		"-Xmx1g":        `"-Xmx1g"`,
		"-Dx=$(HOME)":   `"-Dx=$$(HOME)"`,
		`-Dx="a"`:       `"-Dx=\"a\""`,
		`-Dpath=C:\tmp`: `"-Dpath=C:\\tmp"`,
	} {
		if got := envValueQuote(in); got != want {
			t.Errorf("envValueQuote(%q) = %s, want %s", in, got, want)
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

// TestCreatedSecretsTakeDerivedNames pins the fix for two instances sharing
// one namespace: every Secret the tool builds is named after deployment.name,
// so neither instance's deploy can overwrite the other's credentials and
// neither one's remove can delete them. The retired create.name keys an old
// env.yaml still carries are ignored rather than rejected, so that file keeps
// deploying; only validate reports them.
func TestCreatedSecretsTakeDerivedNames(t *testing.T) {
	k := baseKube()
	k.Secrets = spec.Secrets{
		Credentials: &spec.CredentialsSecret{Create: &spec.CredCreate{Name: "shared-creds"}},
		Stores:      &spec.StoresSecret{Create: &spec.StoreCreate{Name: "shared-tls"}},
	}
	out := Render(Input{
		Kube: k, Defaults: &spec.Defaults{},
		CredKVs:  []KV{{Key: "A", Val: "x"}},
		Stores:   []StoreFile{{Name: "mq-ca", Base64: "QUJD"}},
		Instance: one(k.Deployment.Name, "x: 1\n", &consolidate.Model{Stores: []consolidate.StoreUse{{Name: "mq-ca", File: "./t.p12"}}}),
	})
	for _, want := range []string{
		"  name: solmq" + spec.CredentialsSecretSuffix + "\n",
		"  name: solmq" + spec.StoresSecretSuffix + "\n",
		"secretName: solmq" + spec.CredentialsSecretSuffix + "\n",
		"- secret:\n                  name: solmq" + spec.StoresSecretSuffix + "\n",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q:\n%s", want, out)
		}
	}
	for _, retired := range []string{"shared-creds", "shared-tls"} {
		if strings.Contains(out, retired) {
			t.Errorf("the retired name %q must be ignored, but it reached the manifest:\n%s", retired, out)
		}
	}
}

func TestRenderLibsPVCCreate(t *testing.T) {
	k := baseKube()
	// Name is the retired key: it must not reach the manifest at all.
	k.Libs = &spec.Libs{PVC: &spec.LibsPVC{Create: &spec.PVCCreate{
		Name: "jar-libs", Storage: "2Gi", NFS: spec.NFS{Server: "nfs1", Path: "/libs"},
	}}}
	out := Render(Input{Kube: k, Defaults: &spec.Defaults{}, Instance: one(k.Deployment.Name, "x: 1\n", &consolidate.Model{})})
	claim := k.Deployment.Name + spec.LibsPVCSuffix
	for _, want := range []string{
		// The claim is named after the deployment, so no other instance in the
		// namespace can share it; the PV name carries the namespace on top,
		// because the PV is cluster-scoped. Both must agree on volumeName.
		"kind: PersistentVolume", "name: " + spec.LibsPVName(k.Deployment.Namespace, claim),
		"persistentVolumeReclaimPolicy: Retain",
		"server: nfs1", "path: /libs", "readOnly: true",
		"kind: PersistentVolumeClaim", "name: " + claim + "\n", `storageClassName: ""`,
		"volumeName: " + spec.LibsPVName(k.Deployment.Namespace, claim),
		"storage: 2Gi", "claimName: " + claim,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q\n%s", want, out)
		}
	}
	if strings.Contains(out, "jar-libs") {
		t.Errorf("the retired libs.pvc.create.name must be ignored, but it reached the manifest:\n%s", out)
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
	}
	m := &consolidate.Model{Stores: []consolidate.StoreUse{{Name: "mq-ca", Secret: &spec.StoreSecret{Name: "my-tls"}}}}
	out := Render(Input{Kube: k, Defaults: &spec.Defaults{}, Instance: one(k.Deployment.Name, "x: 1\n", m)})
	if strings.Contains(out, "kind: Secret") {
		t.Error("existing secrets must not emit Secret docs")
	}
	if !strings.Contains(out, "secretName: my-creds") || !strings.Contains(out, "- secret:\n                  name: my-tls\n") {
		t.Errorf("existing refs missing:\n%s", out)
	}
}

// The stores volume is one projected volume mounting each store a connection
// selects (Model.Stores) under its tls.stores name, whatever the file is
// called: the tool's stores Secret, keyed by store name, for every store read
// from a file or built; each operator Secret a store names, once, with the key
// the store names (its own name by default). The tool's Secret comes first;
// with no store selected there is no volume at all.
func TestStoresVolumeProjectsEverySource(t *testing.T) {
	item := func(key, path string) string {
		return "                    - key: \"" + key + "\"\n                      path: " + path + "\n"
	}
	src := func(secret string, items ...string) string {
		return "              - secret:\n                  name: " + secret + "\n                  items:\n" + strings.Join(items, "")
	}
	cert := &spec.StoreSecret{Name: "broker-cert", Key: "keystore.p12"}
	for _, c := range []struct {
		name    string
		uses    []consolidate.StoreUse
		sources string // the sources: block of the stores volume; "" for no volume
	}{
		{"two host-held stores", []consolidate.StoreUse{{Name: "mq-ca", File: "./certs/store.p12"}, {Name: "solace-client", Created: true, Private: true}},
			src("solmq-stores", item("mq-ca", "mq-ca"), item("solace-client", "solace-client"))},
		{"one file under two names", []consolidate.StoreUse{{Name: "a", File: "./c.p12"}, {Name: "b", File: "./c.p12"}},
			src("solmq-stores", item("a", "a"), item("b", "b"))},
		{"two stores from one operator Secret", []consolidate.StoreUse{{Name: "cm-id", Secret: cert}, {Name: "cm-ca", Secret: &spec.StoreSecret{Name: "broker-cert", Key: "truststore.p12"}}},
			src("broker-cert", item("keystore.p12", "cm-id"), item("truststore.p12", "cm-ca"))},
		{"a Secret first selected after an operator one", []consolidate.StoreUse{{Name: "cm-id", Secret: cert}, {Name: "mq-ca", File: "./m.p12"}, {Name: "x", Secret: &spec.StoreSecret{Name: "other"}}},
			src("solmq-stores", item("mq-ca", "mq-ca")) + src("broker-cert", item("keystore.p12", "cm-id")) + src("other", item("x", "x"))},
		{"a file beside a Secret is read from the Secret", []consolidate.StoreUse{{Name: "shared-ca", File: "./s.p12", Secret: &spec.StoreSecret{Name: "shared"}}},
			src("shared", item("shared-ca", "shared-ca"))},
		{"no store selected", nil, ""},
	} {
		t.Run(c.name, func(t *testing.T) {
			k := baseKube()
			out := Render(Input{Kube: k, Defaults: &spec.Defaults{}, Instance: one(k.Deployment.Name, "x: 1\n", &consolidate.Model{Stores: c.uses})})
			if c.sources == "" {
				if strings.Contains(out, "- name: stores") {
					t.Errorf("no store selected, no stores volume:\n%s", out)
				}
				return
			}
			head := "        - name: stores\n          projected:\n            defaultMode: 0400\n            sources:\n"
			if want := head + c.sources; !strings.Contains(out, want) {
				t.Errorf("stores volume missing\n%s\nin:\n%s", want, out)
			}
			if !strings.Contains(out, "              mountPath: /app/external/classpath/truststores\n") {
				t.Errorf("stores not mounted at the store dir:\n%s", out)
			}
		})
	}
	if got := storeSources(Input{Kube: baseKube()}); got != nil {
		t.Errorf("storeSources with no model = %v, want nil", got)
	}
}

// TestProbesSplitLivenessFromReadiness pins the deliberate asymmetry between
// the two probes, which is the whole reason they are not the same check.
//
// Readiness execs the same --health check the compose and quadlet artifacts
// run, so READY means the connector reports itself UP rather than merely that
// its port is open, and the verdict means the same thing on all three
// platforms.
//
// Liveness stays a tcpSocket probe, because it decides whether kubelet
// restarts the container. Running the health check there would turn a slow or
// flapping downstream into a restart loop.
func TestProbesSplitLivenessFromReadiness(t *testing.T) {
	out := Render(fullFixtureInput())

	live := strings.Index(out, "livenessProbe:")
	ready := strings.Index(out, "readinessProbe:")
	if live == -1 || ready == -1 || ready < live {
		t.Fatalf("expected livenessProbe then readinessProbe:\n%s", out)
	}
	if !strings.Contains(out[live:ready], "tcpSocket:") {
		t.Errorf("livenessProbe must stay a tcpSocket probe, or a slow downstream becomes a restart loop:\n%s", out[live:ready])
	}
	if strings.Contains(out[live:ready], "exec:") {
		t.Errorf("livenessProbe must not exec the health check:\n%s", out[live:ready])
	}

	readyBlock := out[ready:]
	if end := strings.Index(readyBlock, "resources:"); end != -1 {
		readyBlock = readyBlock[:end]
	}
	// Built from the shared constants, and passed to a shell: the ConfigMap
	// mounts the script through a subPath, which never carries the execute
	// bit, so exec'ing the path directly would fail with a permission error
	// that reads as an unhealthy connector.
	for _, want := range []string{
		"exec:",
		"command:",
		"- " + statusscript.HealthShell,
		"- " + statusscript.ContainerPath,
		"- " + statusscript.HealthArg,
	} {
		if !strings.Contains(readyBlock, want) {
			t.Errorf("readinessProbe missing %q:\n%s", want, readyBlock)
		}
	}
	// kubernetes defaults timeoutSeconds to 1s, which an HTTP round trip plus
	// a shell will lose -- an unset timeout reports a healthy connector as not
	// ready. The cadence stays the probe's own rather than the slower engine
	// interval, so a rollout is not dragged out.
	for _, want := range []string{
		"initialDelaySeconds: 15",
		"periodSeconds: 10",
		"timeoutSeconds: 5",
		"failureThreshold: 3",
	} {
		if !strings.Contains(readyBlock, want) {
			t.Errorf("readinessProbe missing %q:\n%s", want, readyBlock)
		}
	}
	// The probe execs a path the pod's own volumeMounts are what provide.
	if !strings.Contains(out, "mountPath: "+statusscript.ContainerPath) {
		t.Errorf("readinessProbe execs %s but the pod does not mount it:\n%s", statusscript.ContainerPath, out)
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
		Storage: "2Gi", NFS: spec.NFS{Server: "nfs1", Path: "/libs"},
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
	// Minus the dropped Namespace, plus the stores Secret a teardown always
	// names (this apply selects no store, so it renders none).
	if got, want := strings.Count(del, "\n---\n"), strings.Count(apply, "\n---\n"); got != want {
		t.Errorf("separator count = %d, want %d (apply minus the Namespace, plus the stores Secret)", got, want)
	}
}

// TestTeardownNamesSecretsOnly pins what remove hands kubectl delete for the
// Secrets: each one the tool owns by kind and name alone -- delete needs
// nothing else, so remove resolves no credential and opens no store -- the
// stores Secret always, since its name is the instance's and an earlier deploy
// may have left one, and never a Secret the operator manages.
func TestTeardownNamesSecretsOnly(t *testing.T) {
	nameOnly := func(name string) string {
		return "apiVersion: v1\nkind: Secret\nmetadata:\n  name: " + name + "\n  namespace: ns\n---"
	}
	m := &consolidate.Model{Stores: []consolidate.StoreUse{{Name: "mq-ca", File: "./m.p12"}}}

	k := baseKube()
	k.Secrets = spec.Secrets{
		Credentials: &spec.CredentialsSecret{Create: &spec.CredCreate{}},
		ImagePull:   &spec.ImagePullSecret{Create: true},
	}
	del := Render(Input{Kube: k, Defaults: &spec.Defaults{}, Instance: one(k.Deployment.Name, "x: 1\n", m), Teardown: true,
		ImagePull: &PullSecret{Name: "solmq-image-pull"}})
	for _, want := range []string{nameOnly("solmq-credentials"), nameOnly("solmq-stores"), nameOnly("solmq-image-pull")} {
		if !strings.Contains(del+"---", want) {
			t.Errorf("teardown missing the name-only Secret\n%s\nin:\n%s", want, del)
		}
	}
	for _, payload := range []string{"stringData:", "data:\n  mq-ca", ".dockerconfigjson", "type: Opaque"} {
		if strings.Contains(del, payload) {
			t.Errorf("a teardown carries no Secret payload, found %q:\n%s", payload, del)
		}
	}

	// Secrets the operator manages are never named for delete.
	k = baseKube()
	k.Secrets = spec.Secrets{
		Credentials: &spec.CredentialsSecret{Existing: "their-creds"},
		ImagePull:   &spec.ImagePullSecret{Name: "regcred"},
	}
	opm := &consolidate.Model{Stores: []consolidate.StoreUse{{Name: "cm-id", Secret: &spec.StoreSecret{Name: "broker-cert"}}}}
	del = Render(Input{Kube: k, Defaults: &spec.Defaults{}, Instance: one(k.Deployment.Name, "x: 1\n", opm), Teardown: true,
		ImagePull: &PullSecret{Name: "regcred"}})
	for _, theirs := range []string{"  name: their-creds\n", "  name: regcred\n", "  name: broker-cert\n  namespace"} {
		if strings.Contains(del, theirs) {
			t.Errorf("a teardown must not name an operator Secret (%q):\n%s", theirs, del)
		}
	}
	if !strings.Contains(del+"---", nameOnly("solmq-stores")) {
		t.Errorf("the stores Secret is named even when this spec renders none:\n%s", del)
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
