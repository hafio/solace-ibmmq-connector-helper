package gen_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/deploy"
	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/gen"
	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/scan"
	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/spec"
	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/statusscript"
)

// goldenStatusPassword is the obviously-synthetic literal the golden fixture's
// solmq-status account carries. Setting spec.StatusUserPasswordEnvVar makes
// gen.resolveStatusPassword use it verbatim instead of generating a random
// one, so testdata/golden/application.yml can pin an exact byte-for-byte value.
const goldenStatusPassword = "0123456789abcdef0123456789abcdef"

const specsDir = "../../testdata/golden/specs"

// loadSpecs reads the golden spec folder into a gen.Request (mirrors the CLI):
// env.yaml is the config, every other *.yaml/*.yml is a workflow.
func loadSpecs(t *testing.T) gen.Request {
	t.Helper()
	envPath := filepath.Join(specsDir, "env.yaml")
	absEnv, err := filepath.Abs(envPath)
	if err != nil {
		t.Fatalf("abs: %v", err)
	}
	res, err := scan.Scan(specsDir, "*", absEnv)
	if err != nil {
		t.Fatalf("scan: %v", err)
	}
	req := gen.Request{Env: &gen.File{Name: "env.yaml", Data: mustRead(t, envPath)}}
	for _, p := range res.WorkflowFiles {
		req.Workflows = append(req.Workflows, gen.File{Name: filepath.Base(p), Data: mustRead(t, p)})
	}
	return req
}

// envWithKube returns the golden env.yaml with its kubernetes: section replaced
// by kubeBlock, so a variant test keeps the exact defaults (and therefore the
// exact application.yml) while swapping only the deploy settings.
func envWithKube(t *testing.T, kubeBlock string) []byte {
	t.Helper()
	full := mustRead(t, filepath.Join(specsDir, "env.yaml"))
	i := bytes.Index(full, []byte("\nkubernetes:\n"))
	if i < 0 {
		t.Fatal("golden env.yaml has no kubernetes: section")
	}
	base := append([]byte(nil), full[:i+1]...) // defaults + workflows, trailing newline
	return append(base, kubeBlock...)
}

// setGoldenCredEnv sets the six host environment variables the golden spec's
// -env credential positions name, plus spec.StatusUserPasswordEnvVar so the
// reserved status account's password is the fixed goldenStatusPassword rather
// than a freshly generated one. validate checks every -env reference resolves
// to a set variable regardless of target (S4a: catch a missing/misspelled
// name while linting rather than only at deploy time), so even
// TestGoldenConfig -- which never reads the values -- needs them set.
func setGoldenCredEnv(t *testing.T) {
	t.Helper()
	for k, v := range map[string]string{
		"SOL_PASSWORD": "sol-pw", "MQ_CORE_PASSWORD": "mqcore-pw",
		"MQ_ARCHIVE_PASSWORD": "mqarchive-pw", "EDGE_SOL_PASSWORD": "edge-pw",
		"TRUSTSTORE_PASSWORD": "ts-pw", "KEYSTORE_PASSWORD": "ks-pw",
		spec.StatusUserPasswordEnvVar: goldenStatusPassword,
	} {
		t.Setenv(k, v)
	}
}

func dirReader() func(string) ([]byte, error) {
	return func(p string) ([]byte, error) {
		if !filepath.IsAbs(p) {
			p = filepath.Join(specsDir, p)
		}
		return os.ReadFile(p)
	}
}

func mustRead(t *testing.T, p string) []byte {
	t.Helper()
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatalf("read %s: %v", p, err)
	}
	return b
}

// norm collapses CRLF to LF so byte-for-byte comparison is independent of the
// checkout's line-ending setting (this is a Windows-first repo; the tool's
// output contract is defined in LF).
func norm(s string) string { return strings.ReplaceAll(s, "\r\n", "\n") }

func TestGoldenConfig(t *testing.T) {
	setGoldenCredEnv(t)
	req := loadSpecs(t)
	out, errs, _ := gen.Config(req, gen.Resolver{Env: os.LookupEnv, ReadFile: dirReader()})
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	want := norm(string(mustRead(t, "../../testdata/golden/application.yml")))
	if norm(out) != want {
		t.Errorf("config output mismatch\n%s", lineDiff(want, norm(out)))
	}
}

func TestGoldenKubernetesCreate(t *testing.T) {
	// source: env -- provide the six -env credential values (the literal
	// positions -- usernames -- need no environment at all).
	setGoldenCredEnv(t)
	req := loadSpecs(t)
	out, errs, _ := gen.GenerateKubernetes(req, gen.Resolver{Env: os.LookupEnv, ReadFile: dirReader()})
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}

	appYAML := norm(string(mustRead(t, "../../testdata/golden/application.yml")))
	// kubernetes mounts the stores at MountDir, so the embedded application.yml uses
	// the mount path where `config` keeps the raw env.yaml path.
	appMounted := strings.NewReplacer(
		"./certs/truststore.jks", "/app/external/classpath/truststores/truststore.jks",
		"./certs/keystore.jks", "/app/external/classpath/truststores/keystore.jks",
	).Replace(appYAML)
	want := norm(strings.Join([]string{
		namespaceDoc, // deploy always emits the Namespace first
		configMapDoc(appMounted, true),
		credDoc,
		storesDoc,
		pvDoc,
		pvcDoc,
		deploymentDoc(true, true, true, true), // envFrom + stores + syslog + libs
		serviceDoc,
	}, "---\n"))
	if norm(out) != want {
		t.Errorf("kubernetes(create) mismatch\n%s", lineDiff(want, norm(out)))
	}
}

func TestGoldenKubernetesNoSecrets(t *testing.T) {
	setGoldenCredEnv(t)
	req := loadSpecs(t)
	// Swap in a secrets/syslog/libs-free kubernetes: section (defaults unchanged).
	req.Env = &gen.File{Name: "env.yaml", Data: envWithKube(t, kubeNoSecrets)}
	out, errs, _ := gen.GenerateKubernetes(req, gen.Resolver{Env: os.LookupEnv, ReadFile: dirReader()})
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	appYAML := norm(string(mustRead(t, "../../testdata/golden/application.yml")))
	// kubernetes mounts the stores at MountDir, so the embedded application.yml uses
	// the mount path where `config` keeps the raw env.yaml path.
	appMounted := strings.NewReplacer(
		"./certs/truststore.jks", "/app/external/classpath/truststores/truststore.jks",
		"./certs/keystore.jks", "/app/external/classpath/truststores/keystore.jks",
	).Replace(appYAML)
	want := norm(strings.Join([]string{
		namespaceDoc,
		configMapDoc(appMounted, false),
		deploymentDoc(false, false, false, false), // no envFrom, stores, syslog, or libs
		serviceDoc,
	}, "---\n"))
	if norm(out) != want {
		t.Errorf("kubernetes(no-secrets) mismatch\n%s", lineDiff(want, norm(out)))
	}
}

// ---- expected document builders ---------------------------------------------

const namespaceDoc = `apiVersion: v1
kind: Namespace
metadata:
  name: solace-connectors
`

func configMapDoc(appYAML string, logback bool) string {
	var b strings.Builder
	b.WriteString("apiVersion: v1\nkind: ConfigMap\nmetadata:\n")
	b.WriteString("  name: solmq-connector-config\n  namespace: solace-connectors\n")
	b.WriteString("data:\n  application.yml: |\n")
	for _, ln := range strings.Split(strings.TrimSuffix(appYAML, "\n"), "\n") {
		if ln == "" {
			b.WriteString("\n")
		} else {
			b.WriteString("    " + ln + "\n")
		}
	}
	if logback {
		b.WriteString("  logback-spring.xml: |\n")
		for _, ln := range strings.Split(strings.TrimSuffix(deploy.LogbackXML("udp"), "\n"), "\n") {
			if ln == "" {
				b.WriteString("\n")
			} else {
				b.WriteString("    " + ln + "\n")
			}
		}
	}
	// The status script is always present, regardless of secrets/syslog/libs:
	// the operator execs it inside the container in every configuration. 8090
	// mirrors the golden fixture's management.port, unaffected by which
	// kubernetes: block (create vs. no-secrets) the caller swapped in.
	b.WriteString("  status: |\n")
	for _, ln := range strings.Split(strings.TrimSuffix(statusscript.Render(8090, spec.StatusUserName), "\n"), "\n") {
		if ln == "" {
			b.WriteString("\n")
		} else {
			b.WriteString("    " + ln + "\n")
		}
	}
	return b.String()
}

// credDoc lists every credential the golden spec's config references, in
// first-use order: the two mq-conn-1 positions, the two shared store
// passwords (referenced first by the mq-conn-1 bundle), prod-solace,
// mq-archive, sol-conn-1, and the leader-election session (a distinct secret
// from prod-solace even though it carries the same underlying value). The
// golden spec configures no operator security.users, so the reserved
// solmq-status account (goldenStatusPassword, a literal rather than a
// secretRef) is the only management user and never appears here. Usernames
// are literal in the spec, so their values are the literal text; passwords
// are -env, so their values are the six t.Setenv values set by
// setGoldenCredEnv.
const credDoc = `apiVersion: v1
kind: Secret
metadata:
  name: solmq-credentials
  namespace: solace-connectors
type: Opaque
stringData:
  MQ_CONN_1_USER: "appuser"
  MQ_CONN_1_PASSWORD: "mqcore-pw"
  TRUSTSTORE_PASSWORD: "ts-pw"
  KEYSTORE_PASSWORD: "ks-pw"
  PROD_SOLACE_CLIENT_USERNAME: "connector"
  PROD_SOLACE_CLIENT_PASSWORD: "sol-pw"
  MQ_ARCHIVE_USER: "appuser"
  MQ_ARCHIVE_PASSWORD: "mqarchive-pw"
  SOL_CONN_1_CLIENT_USERNAME: "bridge"
  SOL_CONN_1_CLIENT_PASSWORD: "edge-pw"
  LEADER_ELECTION_CLIENT_USERNAME: "connector"
  LEADER_ELECTION_CLIENT_PASSWORD: "sol-pw"
`

const storesDoc = `apiVersion: v1
kind: Secret
metadata:
  name: solmq-tls
  namespace: solace-connectors
type: Opaque
data:
  truststore.jks: VFJVU1RTVE9SRQo=
  keystore.jks: S0VZU1RPUkUK
`

const pvDoc = `apiVersion: v1
kind: PersistentVolume
metadata:
  name: jar-libs-pvc-pv
spec:
  capacity:
    storage: 1Gi
  accessModes:
    - ReadWriteMany
  nfs:
    server: nfs1.corp.example
    path: /solace-libs
    readOnly: true
`

const pvcDoc = `apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: jar-libs-pvc
  namespace: solace-connectors
spec:
  storageClassName: ""
  volumeName: jar-libs-pvc-pv
  accessModes:
    - ReadWriteMany
  resources:
    requests:
      storage: 1Gi
`

const serviceDoc = `apiVersion: v1
kind: Service
metadata:
  name: solmq-connector
  namespace: solace-connectors
spec:
  selector:
    app: solmq-connector
  ports:
    - name: management
      port: 8090
      targetPort: 8090
`

// deploymentDoc builds the expected Deployment manifest. hasCreds controls
// the credentials Secret's volume + mount: no envFrom exists any more --
// credentials are mounted as files under /run/secrets (matching the
// connector's spring.config.import configtree, S3: no secret in an env var a
// child process or crash dump could see) -- and automountServiceAccountToken
// is always false regardless of hasCreds (the connector never calls the
// Kubernetes API).
func deploymentDoc(hasCreds, stores, syslog, libs bool) string {
	var b strings.Builder
	b.WriteString(`apiVersion: apps/v1
kind: Deployment
metadata:
  name: solmq-connector
  namespace: solace-connectors
spec:
  replicas: 2
  selector:
    matchLabels:
      app: solmq-connector
  template:
    metadata:
      labels:
        app: solmq-connector
        solace-connector/le-mode: active_standby
    spec:
      automountServiceAccountToken: false
      containers:
        - name: connector
          image: solace/solace-pubsub-connector-ibmmq:2.13.0
          ports:
            - name: management
              containerPort: 8090
          env:
            - name: TZ
              value: Asia/Singapore
            - name: JAVA_TOOL_OPTIONS
              value: "-Dcom.ibm.mq.cfg.useIBMCipherMappings=false"
`)
	if syslog {
		b.WriteString(`            - name: LOGGING_SYSLOG_APPNAME
              value: solmq-connector
            - name: LOGGING_SYSLOG_HOST
              value: syslog.corp.example
            - name: LOGGING_SYSLOG_PORT
              value: "514"
`)
	}
	b.WriteString("          volumeMounts:\n")
	if hasCreds {
		b.WriteString("            - name: secrets\n              mountPath: /run/secrets\n              readOnly: true\n")
	}
	b.WriteString(`            - name: config
              mountPath: /app/external/spring/config/application.yml
              subPath: application.yml
              readOnly: true
`)
	if syslog {
		b.WriteString(`            - name: config
              mountPath: /app/external/classpath/logback-spring.xml
              subPath: logback-spring.xml
              readOnly: true
`)
	}
	if stores {
		b.WriteString("            - name: stores\n              mountPath: /app/external/classpath/truststores\n              readOnly: true\n")
	}
	if libs {
		b.WriteString("            - name: libs\n              mountPath: /app/external/libs\n              readOnly: true\n")
	}
	// The status-script mount is unconditional and always last (deploy.go:
	// renderDeployment), so a libs directory mount above never shadows it.
	b.WriteString("            - name: config\n              mountPath: /app/external/.status-script\n              subPath: status\n              readOnly: true\n")
	b.WriteString(`          livenessProbe:
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
            name: solmq-connector-config
`)
	if hasCreds {
		b.WriteString("        - name: secrets\n          secret:\n            secretName: solmq-credentials\n            defaultMode: 0400\n")
	}
	if stores {
		b.WriteString("        - name: stores\n          secret:\n            secretName: solmq-tls\n")
	}
	if libs {
		b.WriteString(`        - name: libs
          persistentVolumeClaim:
            claimName: jar-libs-pvc
`)
	}
	return b.String()
}

// kubeNoSecrets is a full kubernetes: section with no secrets, syslog, or libs.
const kubeNoSecrets = `kubernetes:
  command: kubectl
  deployment:
    name: solmq-connector
    namespace: solace-connectors
    replicas: 2
    resources:
      cpu: "1"
      memory: 1Gi
  service:
    enabled: true
    port: 8090
`

// lineDiff returns a compact first-divergence report for two multi-line strings.
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
			return "first diff at line " + itoa(i+1) + ":\n  want: " + quote(wv) + "\n  got:  " + quote(gv)
		}
	}
	return "(strings differ only in length/trailing content)"
}

func quote(s string) string { return "\"" + s + "\"" }

func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	var d []byte
	for i > 0 {
		d = append([]byte{byte('0' + i%10)}, d...)
		i /= 10
	}
	return string(d)
}
