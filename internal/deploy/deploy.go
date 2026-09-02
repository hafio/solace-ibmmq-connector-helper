// Package deploy renders the Kubernetes manifests for a consolidated connector:
// ConfigMap (embedding application.yml) + optional Secrets + Deployment +
// optional Service. It is pure: secret values and .jks bytes are resolved by the
// caller (CLI from env/file) and passed in, so this package stays pure.
package deploy

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/consolidate"
	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/logback"
	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/spec"
	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/statusscript"
	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/yamlwriter"
)

// SecretsMountPath is where the credentials Secret is mounted, one file per
// key. Named here because the manifest reads better for it; spec.SecretsMountPath
// is the definition, and it is shared with docker, podman and the configtree
// import so one application.yml resolves everywhere.
const SecretsMountPath = spec.SecretsMountPath

// KV is one resolved credential entry (plaintext; K8s base64-encodes stringData).
type KV struct{ Key, Val string }

// StoreFile is one resolved store entry for the stores Secret.
type StoreFile struct {
	Name   string // base name, e.g. truststore.jks
	Base64 string // base64 of the file bytes
}

// PullSecret is the resolved registry-credential wiring for the pod template.
//
// Name is set whenever an image-pull block is present and always reaches the
// pod as an imagePullSecrets entry. DockerConfigJSON is the base64
// .dockerconfigjson payload and is non-empty only when the tool was asked to
// build the Secret itself -- referencing one the operator manages leaves it
// empty, so no Secret object is rendered and theirs is not overwritten.
type PullSecret struct {
	Name             string
	DockerConfigJSON string
}

// Instance is the connector: its ConfigMap + Deployment + Service.
type Instance struct {
	Name         string             // deployment name
	Image        string             // the reference to pull, from the top-level image: block
	Timezone     string             // container TZ, from the top-level timezone: key
	AppYAML      string             // the rendered application.yml (trailing newline)
	StatusScript string             // the rendered status script the operator execs inside the container
	Model        *consolidate.Model // the consolidated model (drives MQTLS etc.)
}

// Input is everything needed to render the manifests.
type Input struct {
	Kube     *spec.Kubernetes
	Defaults *spec.Defaults
	// Syslog is the top-level logging.syslog block, nil when absent. It arrives
	// as its own field rather than being read off Kube because it is no longer a
	// kubernetes key: the same value drives the docker and podman renderers.
	Syslog *spec.Syslog
	// Teardown renders the set for `kubectl delete -f -` rather than `apply -f -`,
	// which changes two things.
	//
	// The Namespace document is dropped. Piping one to delete removes the
	// namespace and cascades to everything inside it, including workloads this
	// tool never deployed -- so remove renders without it and handles the
	// namespace as its own, separately confirmed step. apply always wants it:
	// creating a namespace that already exists is a no-op, and the objects below
	// need it to exist in the same apply.
	//
	// And the documents come out in reverse. kubectl deletes them in file order
	// and waits for each one to go, so creation order deadlocks: the PVC is
	// emitted before the Deployment, kubectl blocks on it because
	// kubernetes.io/pvc-protection holds it while a pod still mounts it, and the
	// Deployment that owns that pod is never reached. Reversed, the workload goes
	// first and everything it holds is free by the time its turn comes.
	Teardown  bool
	CredKVs   []KV        // resolved credential values (only when credentials.create)
	Stores    []StoreFile // resolved .jks files (only when stores.create)
	ImagePull *PullSecret // nil when no image-pull block is configured
	Instance  Instance
}

// yw is the indentation-aware line writer used by every renderer here.
type yw = yamlwriter.Writer

var intOnly = regexp.MustCompile(`^[0-9]+$`)

// quoteRes quotes a resource quantity only when it is a bare integer (so cpu "1"
// stays a string while 250m / 512Mi / 1Gi remain unquoted).
func quoteRes(v string) string {
	if intOnly.MatchString(v) {
		return `"` + v + `"`
	}
	return v
}

// leaderMode returns the pod's leader-election mode label value, defaulting
// to standalone when Defaults is absent (mirrors the ManagementPort nil-guard
// style: a *Defaults that was never parsed still renders).
func leaderMode(d *spec.Defaults) string {
	if d == nil {
		return spec.LeaderStandalone
	}
	return d.LeaderElection.EffectiveMode()
}

// Render produces the full multi-doc manifest set, in the order Namespace,
// ConfigMap, Secrets, libs PV/PVC, Deployment, Service -- reversed when
// in.Teardown, for the reason recorded there.
func Render(in Input) string {
	dep := in.Kube.Deployment
	ns := dep.Namespace

	// Secret references / emission flags.
	credRef, emitCred := "", false
	if c := in.Kube.Secrets.Credentials; c != nil {
		if c.Create != nil {
			credRef, emitCred = c.Create.Name, true
		} else if c.Existing != "" {
			credRef = c.Existing
		}
	}
	storeRef, emitStores := "", false
	if s := in.Kube.Secrets.Stores; s != nil {
		if s.Create != nil {
			storeRef, emitStores = s.Create.Name, true
		} else if s.Existing != "" {
			storeRef = s.Existing
		}
	}
	hasStores := storeRef != ""
	mgmtPort := ManagementPort(in)

	// Each document is rendered into its own buffer rather than appended to one
	// stream, so the set can be emitted in either order at the end. The writer
	// below is rebound per document; nothing else in this function changes.
	var docs []string
	var w *yw
	sep := func() {
		if w != nil {
			docs = append(docs, w.String())
		}
		w = &yw{}
	}
	flush := func() {
		if w != nil {
			docs = append(docs, w.String())
		}
	}

	// 0. Namespace: emitted first so the objects below land in a namespace that
	// exists in the same apply (applying it when it already exists is a no-op).
	// Omitted for a teardown -- see Input.OmitNamespace for why that matters.
	if !in.Teardown {
		sep()
		w.Raw(NamespaceManifest(ns))
	}

	// 1. ConfigMap.
	sep()
	renderConfigMap(w, in.Instance.Name+"-config", ns, in.Instance.AppYAML, in.Instance.StatusScript, in.Syslog)

	// 2. credentials Secret (stringData).
	if emitCred {
		sep()
		w.Line(0, "apiVersion: v1")
		w.Line(0, "kind: Secret")
		w.Line(0, "metadata:")
		w.Line(2, "name: "+credRef)
		w.Line(2, "namespace: "+ns)
		w.Line(0, "type: Opaque")
		w.Line(0, "stringData:")
		for _, kv := range in.CredKVs {
			w.Line(2, kv.Key+": "+strconv.Quote(kv.Val))
		}
	}

	// 3. stores Secret (base64 data).
	if emitStores {
		sep()
		w.Line(0, "apiVersion: v1")
		w.Line(0, "kind: Secret")
		w.Line(0, "metadata:")
		w.Line(2, "name: "+storeRef)
		w.Line(2, "namespace: "+ns)
		w.Line(0, "type: Opaque")
		w.Line(0, "data:")
		for _, sf := range in.Stores {
			w.Line(2, sf.Name+": "+sf.Base64)
		}
	}

	// 3c. image-pull Secret, only when the tool was asked to build one. A
	// reference-only block renders no Secret here, so a Secret the operator
	// created and manages is never overwritten by an apply.
	if ip := in.ImagePull; ip != nil && ip.DockerConfigJSON != "" {
		sep()
		w.Line(0, "apiVersion: v1")
		w.Line(0, "kind: Secret")
		w.Line(0, "metadata:")
		w.Line(2, "name: "+ip.Name)
		w.Line(2, "namespace: "+ns)
		w.Line(0, "type: kubernetes.io/dockerconfigjson")
		w.Line(0, "data:")
		w.Line(2, ".dockerconfigjson: "+ip.DockerConfigJSON)
	}

	// 3b. libs PV + PVC (only for libs.pvc.create; PV is cluster-scoped).
	if lb := in.Kube.Libs; lb != nil && lb.PVC != nil && lb.PVC.Create != nil {
		c := lb.PVC.Create
		sep()
		w.Line(0, "apiVersion: v1")
		w.Line(0, "kind: PersistentVolume")
		w.Line(0, "metadata:")
		w.Line(2, "name: "+spec.LibsPVName(ns, c.Name))
		w.Line(0, "spec:")
		// Stated rather than left to the default, which is the same value: a
		// Retain PV outlives its claim and comes back as Released, which never
		// rebinds, so the operator reading this manifest should not have to know
		// the default to see that.
		w.Line(2, "persistentVolumeReclaimPolicy: Retain")
		w.Line(2, "capacity:")
		w.Line(4, "storage: "+c.Storage)
		w.Line(2, "accessModes:")
		w.Line(4, "- ReadWriteMany")
		w.Line(2, "nfs:")
		w.Line(4, "server: "+c.NFS.Server)
		w.Line(4, "path: "+c.NFS.Path)
		w.Line(4, "readOnly: true")
		sep()
		w.Line(0, "apiVersion: v1")
		w.Line(0, "kind: PersistentVolumeClaim")
		w.Line(0, "metadata:")
		w.Line(2, "name: "+c.Name)
		w.Line(2, "namespace: "+ns)
		w.Line(0, "spec:")
		w.Line(2, `storageClassName: ""`)
		w.Line(2, "volumeName: "+spec.LibsPVName(ns, c.Name))
		w.Line(2, "accessModes:")
		w.Line(4, "- ReadWriteMany")
		w.Line(2, "resources:")
		w.Line(4, "requests:")
		w.Line(6, "storage: "+c.Storage)
	}

	// 4. Deployment.
	sep()
	renderDeployment(w, in, in.Instance, ns, credRef, storeRef, hasStores, mgmtPort)

	// 5. Service (when enabled). spec already defaults an unset service.port to
	// mgmtPort:mgmtPort (applyKubeDefaults), so there is no fallback to redo here.
	if in.Kube.Service.Enabled {
		sep()
		renderService(w, in.Instance.Name, ns, in.Kube.Service.Port)
	}
	flush()

	if in.Teardown {
		for i, j := 0, len(docs)-1; i < j; i, j = i+1, j-1 {
			docs[i], docs[j] = docs[j], docs[i]
		}
	}
	return strings.Join(docs, "---\n")
}

// renderConfigMap emits the ConfigMap embedding the application.yml, the
// status script (always present, so the operator can exec it inside the
// container regardless of leader-election mode) and, when syslog is
// configured, the shared logback-spring.xml.
func renderConfigMap(w *yw, cmName, ns, appYAML, statusScript string, sys *spec.Syslog) {
	w.Line(0, "apiVersion: v1")
	w.Line(0, "kind: ConfigMap")
	w.Line(0, "metadata:")
	w.Line(2, "name: "+cmName)
	w.Line(2, "namespace: "+ns)
	w.Line(0, "data:")
	w.Line(2, "application.yml: |")
	for _, ln := range yamlwriter.SplitLines(appYAML) {
		if ln == "" {
			w.Raw("\n")
		} else {
			w.Line(4, ln)
		}
	}
	if sys != nil {
		w.Line(2, logback.FileName+": |")
		for _, ln := range yamlwriter.SplitLines(logback.XML(sys.Protocol)) {
			if ln == "" {
				w.Raw("\n")
			} else {
				w.Line(4, ln)
			}
		}
	}
	w.Line(2, "status: |")
	for _, ln := range yamlwriter.SplitLines(statusScript) {
		if ln == "" {
			w.Raw("\n")
		} else {
			w.Line(4, ln)
		}
	}
}

func renderDeployment(w *yw, in Input, inst Instance, ns, credRef, storeRef string, hasStores bool, mgmtPort int) {
	dep := in.Kube.Deployment
	name := inst.Name
	cmName := inst.Name + "-config"
	w.Line(0, "apiVersion: apps/v1")
	w.Line(0, "kind: Deployment")
	w.Line(0, "metadata:")
	w.Line(2, "name: "+name)
	w.Line(2, "namespace: "+ns)
	w.Line(0, "spec:")
	w.Line(2, "replicas: "+strconv.Itoa(dep.Replicas))
	w.Line(2, "selector:")
	w.Line(4, "matchLabels:")
	// app only: a selector must be immutable for the life of the Deployment, and
	// the role label below changes with the actuator-reported leader, not with
	// anything the selector could pin.
	w.Line(6, "app: "+name)
	w.Line(2, "template:")
	w.Line(4, "metadata:")
	w.Line(6, "labels:")
	w.Line(8, "app: "+name)
	mode := leaderMode(in.Defaults)
	w.Line(8, spec.LabelModeKey+": "+mode)
	// active_standby gets no role label: which pod is currently active is only
	// knowable at runtime from the actuator, which is exactly what the status
	// script (rendered into the ConfigMap above) execs to answer. standalone
	// and active_active are both statically "active" -- there is no standby to
	// distinguish from.
	if mode == spec.LeaderStandalone || mode == spec.LeaderActiveActive {
		w.Line(8, spec.LabelRoleKey+": "+spec.LabelRoleActive)
	}
	w.Line(4, "spec:")
	// The connector never calls the Kubernetes API, and an automounted service
	// account token would land under a /run/secrets tree the configtree import
	// also reads -- turning the token into stray connector properties.
	w.Line(6, "automountServiceAccountToken: false")
	// The kubelet, not the connector, consumes this -- it is how the image is
	// pulled at all from a registry that needs credentials. Emitted for a
	// created and an existing Secret alike; only rendering the Secret differs.
	if ip := in.ImagePull; ip != nil && ip.Name != "" {
		w.Line(6, "imagePullSecrets:")
		w.Line(8, "- name: "+ip.Name)
	}
	lb := in.Kube.Libs
	if lb != nil && lb.Download != nil {
		d := lb.Download
		cmds := make([]string, 0, len(d.URLs))
		for _, u := range d.URLs {
			cmds = append(cmds, "wget -O '/libs/"+spec.BaseName(u)+"' '"+u+"'")
		}
		w.Line(6, "initContainers:")
		w.Line(8, "- name: libs-download")
		w.Line(10, "image: "+d.Image)
		w.Line(10, `command: ["sh", "-c", `+strconv.Quote(strings.Join(cmds, " && "))+`]`)
		w.Line(10, "volumeMounts:")
		w.Line(12, "- name: libs")
		w.Line(14, "mountPath: /libs")
	}
	w.Line(6, "containers:")
	w.Line(8, "- name: "+spec.ConnectorContainerName)
	w.Line(10, "image: "+inst.Image)
	w.Line(10, "ports:")
	w.Line(12, "- name: management")
	w.Line(14, "containerPort: "+strconv.Itoa(mgmtPort))
	// env: guarded as a whole, because every entry under it is optional now that
	// the timezone is a top-level key rather than a required per-platform one --
	// an unguarded "env:" with nothing beneath it is a null, not an empty list.
	sys := in.Syslog
	if inst.Timezone != "" || inst.Model.MQTLS || sys != nil {
		w.Line(10, "env:")
	}
	if inst.Timezone != "" {
		w.Line(12, "- name: TZ")
		w.Line(14, "value: "+inst.Timezone)
	}
	if inst.Model.MQTLS {
		w.Line(12, "- name: JAVA_TOOL_OPTIONS")
		w.Line(14, `value: "-Dcom.ibm.mq.cfg.useIBMCipherMappings=false"`)
	}
	if sys != nil {
		w.Line(12, "- name: LOGGING_SYSLOG_APPNAME")
		w.Line(14, "value: "+name)
		w.Line(12, "- name: LOGGING_SYSLOG_HOST")
		w.Line(14, "value: "+sys.Host)
		w.Line(12, "- name: LOGGING_SYSLOG_PORT")
		w.Line(14, `value: "`+strconv.Itoa(sys.Port)+`"`)
	}
	// volumeMounts
	w.Line(10, "volumeMounts:")
	if credRef != "" {
		// Credentials are mounted as one file per key, which the connector reads
		// through its configtree import -- never injected as environment
		// variables, where any child process or crash dump would see them.
		w.Line(12, "- name: secrets")
		w.Line(14, "mountPath: "+SecretsMountPath)
		w.Line(14, "readOnly: true")
	}
	w.Line(12, "- name: config")
	w.Line(14, "mountPath: /app/external/spring/config/application.yml")
	w.Line(14, "subPath: application.yml")
	w.Line(14, "readOnly: true")
	if in.Syslog != nil {
		w.Line(12, "- name: config")
		w.Line(14, "mountPath: "+logback.ContainerPath)
		w.Line(14, "subPath: "+logback.FileName)
		w.Line(14, "readOnly: true")
	}
	if hasStores {
		w.Line(12, "- name: stores")
		w.Line(14, "mountPath: /app/external/classpath/truststores")
		w.Line(14, "readOnly: true")
	}
	if lb != nil {
		w.Line(12, "- name: libs")
		w.Line(14, "mountPath: /app/external/libs")
		w.Line(14, "readOnly: true")
	}
	// The status script is a sibling of the libs/spring/classpath mounts rather
	// than nested inside one, so no ordering between them matters any more. The
	// subPath is the ConfigMap key, which is unrelated to the container path.
	w.Line(12, "- name: config")
	w.Line(14, "mountPath: "+statusscript.ContainerPath)
	w.Line(14, "subPath: status")
	w.Line(14, "readOnly: true")
	// probes (tcpSocket — see prompt.md "Probes"; keep isolated for a later switch)
	w.Line(10, "livenessProbe:")
	w.Line(12, "tcpSocket:")
	w.Line(14, "port: "+strconv.Itoa(mgmtPort))
	w.Line(12, "initialDelaySeconds: 30")
	w.Line(12, "periodSeconds: 15")
	w.Line(10, "readinessProbe:")
	w.Line(12, "tcpSocket:")
	w.Line(14, "port: "+strconv.Itoa(mgmtPort))
	w.Line(12, "initialDelaySeconds: 15")
	w.Line(12, "periodSeconds: 10")
	// resources
	renderResources(w, dep.Resources)
	// volumes
	w.Line(6, "volumes:")
	w.Line(8, "- name: config")
	w.Line(10, "configMap:")
	w.Line(12, "name: "+cmName)
	if credRef != "" {
		w.Line(8, "- name: secrets")
		w.Line(10, "secret:")
		w.Line(12, "secretName: "+credRef)
		w.Line(12, "defaultMode: 0400")
	}
	if hasStores {
		w.Line(8, "- name: stores")
		w.Line(10, "secret:")
		w.Line(12, "secretName: "+storeRef)
	}
	if lb != nil {
		w.Line(8, "- name: libs")
		switch {
		case lb.PVC != nil && lb.PVC.Existing != "":
			w.Line(10, "persistentVolumeClaim:")
			w.Line(12, "claimName: "+lb.PVC.Existing)
		case lb.PVC != nil && lb.PVC.Create != nil:
			w.Line(10, "persistentVolumeClaim:")
			w.Line(12, "claimName: "+lb.PVC.Create.Name)
		case lb.Download != nil && lb.Download.PVC != "":
			w.Line(10, "persistentVolumeClaim:")
			w.Line(12, "claimName: "+lb.Download.PVC)
		default:
			w.Line(10, "emptyDir: {}")
		}
	}
}

func renderResources(w *yw, r spec.Resources) {
	if r.CPU == "" && r.Memory == "" {
		return
	}
	w.Line(10, "resources:")
	// requests and limits are set identically (guaranteed QoS).
	for _, kind := range []string{"requests", "limits"} {
		w.Line(12, kind+":")
		if r.CPU != "" {
			w.Line(14, "cpu: "+quoteRes(r.CPU))
		}
		if r.Memory != "" {
			w.Line(14, "memory: "+quoteRes(r.Memory))
		}
	}
}

func renderService(w *yw, name, ns string, port spec.Port) {
	w.Line(0, "apiVersion: v1")
	w.Line(0, "kind: Service")
	w.Line(0, "metadata:")
	w.Line(2, "name: "+name)
	w.Line(2, "namespace: "+ns)
	w.Line(0, "spec:")
	w.Line(2, "selector:")
	w.Line(4, "app: "+name)
	w.Line(2, "ports:")
	w.Line(4, "- name: management")
	w.Line(6, "port: "+strconv.Itoa(port.Host))
	w.Line(6, "targetPort: "+strconv.Itoa(port.Container))
}

// ManagementPort returns the effective management port from Defaults.
// Kube.Service.Port plays no part: it is a spec.Port whose Container side is a
// targetPort, not necessarily the management port, and an unset service.port
// is already resolved to mgmtPort:mgmtPort by spec's applyKubeDefaults before
// Input is ever built.
func ManagementPort(in Input) int {
	return in.Defaults.EffectiveManagementPort()
}

// NamespaceManifest is the Namespace document on its own, so the one caller
// that deletes a namespace can pipe it through the same runner.Kubernetes path
// (and therefore the same binary allowlist and verb mapping) as every other
// manifest, instead of building a bespoke argv.
//
// It is the same four lines Render emits, spelled once so the two cannot drift.
func NamespaceManifest(ns string) string {
	w := &yw{}
	w.Line(0, "apiVersion: v1")
	w.Line(0, "kind: Namespace")
	w.Line(0, "metadata:")
	w.Line(2, "name: "+ns)
	return w.String()
}
