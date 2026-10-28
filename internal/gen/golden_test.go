package gen_test

import (
	"bytes"
	"encoding/base64"
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/certfixture"
	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/deploy"
	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/gen"
	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/logback"
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

// updateFixtures rewrites the golden spec's PKCS12 fixtures --
// certs/truststore.p12 (opened with TRUSTSTORE_PASSWORD's ts-pw) and
// certs/keystore.p12 (KEYSTORE_PASSWORD's ks-pw) -- from freshly minted
// certificates. They are real stores because every verb that reads an existing
// store opens it with its password; the manifests embed whatever bytes are
// committed, so a rewrite needs no other change.
var updateFixtures = flag.Bool("update-fixtures", false, "rewrite testdata/golden/specs/certs/*.p12")

// TestGoldenStoreFixtures checks the committed fixtures open with the
// passwords the golden spec gives them, and rewrites them with -update-fixtures.
func TestGoldenStoreFixtures(t *testing.T) {
	trust, key := filepath.Join(specsDir, "certs", "truststore.p12"), filepath.Join(specsDir, "certs", "keystore.p12")
	if *updateFixtures {
		ca := certfixture.CA(t, "Golden Root")
		for path, data := range map[string][]byte{
			trust: certfixture.TrustStorePKCS12(t, "ts-pw", ca),
			key:   certfixture.KeyStorePKCS12(t, "ks-pw", certfixture.Issue(t, ca, "golden-client"), ca),
		} {
			if err := os.WriteFile(path, data, 0o644); err != nil {
				t.Fatal(err)
			}
		}
	}
	setGoldenCredEnv(t)
	errs, warns := gen.Validate(loadSpecs(t), gen.Resolver{Env: os.LookupEnv, ReadFile: dirReader()})
	for _, i := range append(errs, warns...) {
		if strings.Contains(i.Msg, "certs/") {
			t.Errorf("a golden store fixture is not a store its password opens: %s (rewrite them with -update-fixtures)", i.Msg)
		}
	}
}

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

// envWithKubeNoSyslog is envWithKube with the top-level logging.syslog block
// dropped as well.
//
// syslog is a top-level key now, so a variant that wants none has to remove it
// there: swapping the kubernetes: section no longer takes it out, and without
// this the "no secrets" golden would silently start asserting a rendering that
// includes syslog.
func envWithKubeNoSyslog(t *testing.T, kubeBlock string) []byte {
	t.Helper()
	var out []string
	skip := false
	for _, ln := range strings.Split(string(envWithKube(t, kubeBlock)), "\n") {
		if strings.HasPrefix(ln, "  syslog:") {
			skip = true
			continue
		}
		if skip {
			if strings.HasPrefix(ln, "    ") {
				continue
			}
			skip = false
		}
		out = append(out, ln)
	}
	return []byte(strings.Join(out, "\n"))
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

// generate config writes the application.yml deploy ships: the kubernetes
// goldens embed the very file TestGoldenConfig pins, and here the podman unit
// mounts it byte for byte from podman's secret store and the compose file
// inlines it -- indented, with compose's every '$' doubled.
func TestConfigIsWhatDeployShips(t *testing.T) {
	setGoldenCredEnv(t)
	req := loadSpecs(t)
	res := gen.Resolver{Env: os.LookupEnv, ReadFile: dirReader()}
	cfg, errs, _ := gen.Config(req, res)
	if len(errs) > 0 {
		t.Fatalf("config: %v", errs)
	}
	withSection := func(section string) gen.Request {
		r := req
		r.Env = &gen.File{Name: "env.yaml", Data: append(append([]byte(nil), req.Env.Data...), section...)}
		return r
	}
	plan, errs, _ := gen.GeneratePodman(withSection("podman:\n  command: podman\n  name: c\n"), res)
	if len(errs) > 0 {
		t.Fatalf("podman: %v", errs)
	}
	var shipped *gen.PodmanFile
	for i := range plan.Files {
		if plan.Files[i].StoreName == gen.PodmanSecretStoreName("c", "application.yml") {
			shipped = &plan.Files[i]
		}
	}
	if shipped == nil || shipped.Data != cfg {
		t.Errorf("podman ships a different application.yml than generate config writes")
	}
	dp, errs, _ := gen.GenerateDocker(withSection("docker:\n  command: docker\n  name: c\n"), res)
	if len(errs) > 0 {
		t.Fatalf("docker: %v", errs)
	}
	var inlined strings.Builder
	for _, ln := range strings.SplitAfter(cfg, "\n") {
		if ln == "\n" || ln == "" {
			inlined.WriteString(ln)
			continue
		}
		inlined.WriteString("      " + strings.ReplaceAll(ln, "$", "$$"))
	}
	if !strings.Contains(dp.Compose, "    content: |\n"+inlined.String()) {
		t.Errorf("docker ships a different application.yml than generate config writes:\n%s", dp.Compose)
	}
}

func TestGoldenKubernetesCreate(t *testing.T) {
	// source: env -- provide the six -env credential values (the literal
	// positions -- usernames -- need no environment at all).
	setGoldenCredEnv(t)
	req := loadSpecs(t)
	out, errs, _ := gen.GenerateKubernetes(req, gen.Resolver{Env: os.LookupEnv, ReadFile: dirReader()}, gen.KubeOpts{})
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}

	appYAML := norm(string(mustRead(t, "../../testdata/golden/application.yml")))
	want := norm(strings.Join([]string{
		namespaceDoc, // deploy always emits the Namespace first
		configMapDoc(appYAML, true),
		credDoc,
		storesDoc(t),
		pvDoc,
		pvcDoc,
		deploymentDoc("solmq-connector-credentials", true, true, true), // credentials + stores + syslog + libs
		serviceDoc,
	}, "---\n"))
	if norm(out) != want {
		t.Errorf("kubernetes(create) mismatch\n%s", lineDiff(want, norm(out)))
	}
}

func TestGoldenKubernetesNoSecrets(t *testing.T) {
	setGoldenCredEnv(t)
	req := loadSpecs(t)
	// Swap in a syslog/libs-free kubernetes: section that creates no Secret but
	// the stores (defaults unchanged).
	req.Env = &gen.File{Name: "env.yaml", Data: envWithKubeNoSyslog(t, kubeNoSecrets)}
	out, errs, _ := gen.GenerateKubernetes(req, gen.Resolver{Env: os.LookupEnv, ReadFile: dirReader()}, gen.KubeOpts{})
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	appYAML := norm(string(mustRead(t, "../../testdata/golden/application.yml")))
	// No Secret but the stores is created: the credentials are the operator's,
	// and a selected store the tool holds is delivered with nothing to
	// configure, as on docker and podman.
	want := norm(strings.Join([]string{
		namespaceDoc,
		configMapDoc(appYAML, false),
		storesDoc(t),
		deploymentDoc("their-creds", true, false, false), // their credentials + stores; no syslog or libs
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

func configMapDoc(appYAML string, withLogback bool) string {
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
	if withLogback {
		b.WriteString("  logback-spring.xml: |\n")
		for _, ln := range strings.Split(strings.TrimSuffix(logback.XML("udp"), "\n"), "\n") {
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
// mq-archive and sol-conn-1. The leader-election session adds nothing: it
// conn-refs prod-solace, so it reuses that binder's two names rather than
// mounting the same value a second time. The
// golden spec configures no operator security.users, so the reserved
// solmq-status account (goldenStatusPassword, a literal rather than a
// secretRef) is the only management user and never appears here.
//
// The key names show both halves of the naming rule: usernames are literal in
// the spec, so they take the derived name for their position
// (MQ_CONN_1_USER, PROD_SOLACE_CLIENT_USERNAME, ...), while every password is
// -env and is keyed by the host variable the spec names -- so SOL_PASSWORD, not
// PROD_SOLACE_CLIENT_PASSWORD. That is what makes this list reproducible by hand
// for a `credentials.existing` Secret: the keys are read straight off the spec.
// The values are the literal text and the six t.Setenv values set by
// setGoldenCredEnv respectively.
const credDoc = `apiVersion: v1
kind: Secret
metadata:
  name: solmq-connector-credentials
  namespace: solace-connectors
type: Opaque
stringData:
  _GEN_MQ_CONN_1_USER: "appuser"
  MQ_CORE_PASSWORD: "mqcore-pw"
  TRUSTSTORE_PASSWORD: "ts-pw"
  KEYSTORE_PASSWORD: "ks-pw"
  _GEN_PROD_SOLACE_CLIENT_USERNAME: "connector"
  SOL_PASSWORD: "sol-pw"
  _GEN_MQ_ARCHIVE_USER: "appuser"
  MQ_ARCHIVE_PASSWORD: "mqarchive-pw"
  _GEN_SOL_CONN_1_CLIENT_USERNAME: "bridge"
  EDGE_SOL_PASSWORD: "edge-pw"
`

// storesDoc is the stores Secret the golden spec renders: each selected store
// keyed by its name, holding the committed fixture it names, base64-encoded.
func storesDoc(t *testing.T) string {
	t.Helper()
	enc := func(file string) string {
		return base64.StdEncoding.EncodeToString(mustRead(t, filepath.Join(specsDir, "certs", file)))
	}
	return `apiVersion: v1
kind: Secret
metadata:
  name: solmq-connector-stores
  namespace: solace-connectors
type: Opaque
data:
  mq-ca: ` + enc("truststore.p12") + `
  mq-client: ` + enc("keystore.p12") + `
  solace-ca: ` + enc("truststore.p12") + `
  solace-client: ` + enc("keystore.p12") + `
`
}

const pvDoc = `apiVersion: v1
kind: PersistentVolume
metadata:
  name: solace-connectors-solmq-connector-libs-pv
spec:
  persistentVolumeReclaimPolicy: Retain
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
  name: solmq-connector-libs
  namespace: solace-connectors
spec:
  storageClassName: ""
  volumeName: solace-connectors-solmq-connector-libs-pv
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

// deploymentDoc builds the expected Deployment manifest. credsSecret names the
// Secret behind the credentials volume + mount ("" for none): no envFrom exists any more --
// credentials are mounted as files under deploy.SecretsMountPath (one of the
// two directories the connector's spring.config.import configtree names, S3: no secret in an env var a
// child process or crash dump could see) -- and automountServiceAccountToken
// is always false regardless of credsSecret (the connector never calls the
// Kubernetes API). The pod runs as the default non-root run-as with the
// posture every pod carries.
func deploymentDoc(credsSecret string, stores, syslog, libs bool) string {
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
      securityContext:
        runAsUser: 1000
        runAsGroup: 0
        fsGroup: 0
        runAsNonRoot: true
        seccompProfile:
          type: RuntimeDefault
      containers:
        - name: connector
          image: solace/solace-pubsub-connector-ibmmq:2.13.0
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
	if credsSecret != "" {
		b.WriteString("            - name: secrets\n              mountPath: " + deploy.SecretsMountPath + "\n              readOnly: true\n")
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
            name: solmq-connector-config
`)
	if credsSecret != "" {
		b.WriteString("        - name: secrets\n          secret:\n            secretName: " + credsSecret + "\n            defaultMode: 0400\n")
	}
	if stores {
		b.WriteString("        - name: stores\n          projected:\n            defaultMode: 0400\n            sources:\n" +
			"              - secret:\n                  name: solmq-connector-stores\n                  items:\n" +
			"                    - key: \"mq-ca\"\n                      path: mq-ca\n" +
			"                    - key: \"mq-client\"\n                      path: mq-client\n" +
			"                    - key: \"solace-ca\"\n                      path: solace-ca\n" +
			"                    - key: \"solace-client\"\n                      path: solace-client\n")
	}
	if libs {
		b.WriteString(`        - name: libs
          persistentVolumeClaim:
            claimName: solmq-connector-libs
`)
	}
	return b.String()
}

// kubeNoSecrets is a full kubernetes: section with no syslog or libs, and no
// Secret the tool creates: the credentials Secret, which the config's ${...}
// references require, is the operator's own.
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
  secrets:
    credentials:
      existing: their-creds
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
