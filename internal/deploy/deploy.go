// Package deploy renders the Kubernetes manifests for a consolidated connector:
// ConfigMap (embedding application.yml) + optional Secrets + Deployment +
// optional Service. It is pure: secret values and .jks bytes are resolved by the
// caller (CLI from env/file) and passed in, so this package stays pure.
package deploy

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-gen/internal/consolidate"
	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-gen/internal/spec"
)

// KV is one resolved credential entry (plaintext; K8s base64-encodes stringData).
type KV struct{ Key, Val string }

// StoreFile is one resolved store entry for the stores Secret.
type StoreFile struct {
	Name   string // base name, e.g. truststore.jks
	Base64 string // base64 of the file bytes
}

// Instance is one connector instance: its own ConfigMap + Deployment + Service.
// A folder with more than MaxWorkflowsPerInstance workflows is sharded into
// several instances that share one namespace, secrets, and libs volume.
type Instance struct {
	Name    string             // deployment name: the base name, or base-N when sharded
	AppYAML string             // this instance's rendered application.yml (trailing newline)
	Model   *consolidate.Model // this instance's model (drives per-instance MQTLS etc.)
}

// Input is everything needed to render the manifests. Shared objects (namespace,
// secrets, libs PV/PVC) are emitted once; ConfigMap/Deployment/Service are
// emitted per Instance.
type Input struct {
	Kube      *spec.Kubernetes
	Defaults  *spec.Defaults
	CredKVs   []KV        // resolved credential values (only when credentials.create)
	Stores    []StoreFile // resolved .jks files (only when stores.create)
	Instances []Instance  // one or more connector instances
}

type yw struct{ b strings.Builder }

func (w *yw) line(indent int, s string) {
	w.b.WriteString(strings.Repeat(" ", indent))
	w.b.WriteString(s)
	w.b.WriteByte('\n')
}
func (w *yw) raw(s string)   { w.b.WriteString(s) }
func (w *yw) String() string { return w.b.String() }

var intOnly = regexp.MustCompile(`^[0-9]+$`)

// quoteRes quotes a resource quantity only when it is a bare integer (so cpu "1"
// stays a string while 250m / 512Mi / 1Gi remain unquoted).
func quoteRes(v string) string {
	if intOnly.MatchString(v) {
		return `"` + v + `"`
	}
	return v
}

// LogbackXML returns the logback-spring.xml content for the given syslog
// protocol. Host/port/appname flow in at runtime via springProperty from the
// LOGGING_SYSLOG_* env vars set on the Deployment, so the XML itself is static.
func LogbackXML(protocol string) string {
	if protocol == spec.SyslogTCP {
		return logbackTCP
	}
	return logbackUDP
}

// logbackUDP mirrors the reference deployment: Logback's built-in UDP
// (RFC 3164) SyslogAppender behind an async wrapper.
const logbackUDP = `<?xml version="1.0" encoding="UTF-8"?>
<configuration>
  <!-- Spring Boot defaults: conversion rules, CONSOLE appender -->
  <include resource="org/springframework/boot/logging/logback/defaults.xml"/>
  <include resource="org/springframework/boot/logging/logback/console-appender.xml"/>

  <!-- Spring property substitution from application.yml or env vars -->
  <springProperty scope="context" name="SYSLOG_HOST" source="logging.syslog.host" />
  <springProperty scope="context" name="SYSLOG_PORT" source="logging.syslog.port" />
  <springProperty scope="context" name="SYSLOG_APPNAME" source="logging.syslog.appname" defaultValue="solace-ibmmq-connector-default"/>

  <!-- UDP syslog, RFC 3164. This is Logback's built-in, no extra deps. -->
  <appender name="SYSLOG" class="ch.qos.logback.classic.net.SyslogAppender">
    <syslogHost>${SYSLOG_HOST}</syslogHost>
    <port>${SYSLOG_PORT}</port>
    <facility>LOCAL0</facility>
    <suffixPattern>${SYSLOG_APPNAME}[%thread] %-5level %logger{36} - %msg</suffixPattern>
    <stackTracePattern>${SYSLOG_APPNAME}[%thread] \t</stackTracePattern>
    <throwableExcluded>false</throwableExcluded>
  </appender>

  <!-- Async wrapper. With UDP this is less critical (sends never block long)
   but still good practice - keeps any DNS resolution or socket setup off
   the app threads. -->
  <appender name="ASYNC_SYSLOG" class="ch.qos.logback.classic.AsyncAppender">
    <appender-ref ref="SYSLOG"/>
    <queueSize>2048</queueSize>
    <discardingThreshold>0</discardingThreshold>
    <neverBlock>true</neverBlock>
    <includeCallerData>false</includeCallerData>
  </appender>

  <root level="INFO">
    <appender-ref ref="CONSOLE"/>
    <appender-ref ref="ASYNC_SYSLOG"/>
  </root>
</configuration>
`

// logbackTCP sends newline-framed lines over TCP via logstash-logback-encoder's
// LogstashTcpSocketAppender (internally async, so no AsyncAppender wrapper).
// The jar must be provided on the connector classpath, e.g. via libs.
const logbackTCP = `<?xml version="1.0" encoding="UTF-8"?>
<configuration>
  <include resource="org/springframework/boot/logging/logback/defaults.xml"/>
  <include resource="org/springframework/boot/logging/logback/console-appender.xml"/>

  <springProperty scope="context" name="SYSLOG_HOST" source="logging.syslog.host" />
  <springProperty scope="context" name="SYSLOG_PORT" source="logging.syslog.port" />
  <springProperty scope="context" name="SYSLOG_APPNAME" source="logging.syslog.appname" defaultValue="solace-ibmmq-connector-default"/>

  <!-- TCP syslog via logstash-logback-encoder (jar required on the classpath). -->
  <appender name="SYSLOG" class="net.logstash.logback.appender.LogstashTcpSocketAppender">
    <destination>${SYSLOG_HOST}:${SYSLOG_PORT}</destination>
    <encoder class="ch.qos.logback.classic.encoder.PatternLayoutEncoder">
      <pattern>${SYSLOG_APPNAME}[%thread] %-5level %logger{36} - %msg%n</pattern>
    </encoder>
  </appender>

  <root level="INFO">
    <appender-ref ref="CONSOLE"/>
    <appender-ref ref="SYSLOG"/>
  </root>
</configuration>
`

// syslogOf returns the syslog settings, or nil when the block is absent.
func syslogOf(k *spec.Kubernetes) *spec.Syslog {
	if k.Logging == nil {
		return nil
	}
	return k.Logging.Syslog
}

// baseName returns the final path element of a URL (mirrors gen.baseName).
func baseName(p string) string {
	if i := strings.LastIndexByte(p, '/'); i >= 0 {
		return p[i+1:]
	}
	return p
}

// Render produces the full multi-doc manifest set. Shared objects (Namespace,
// Secrets, libs PV/PVC) are emitted once; the ConfigMap, Deployment, and Service
// are emitted per instance. The emission order (Namespace, all ConfigMaps, shared
// Secrets, PV/PVC, all Deployments, all Services) is chosen so that a single
// instance reproduces the historical byte-for-byte layout.
func Render(in Input) string {
	dep := in.Kube.Deployment
	ns := dep.Namespace

	// Secret references / emission flags (shared across instances).
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
	mgmtPort := managementPort(in)

	w := &yw{}
	docs := 0
	sep := func() {
		if docs > 0 {
			w.raw("---\n")
		}
		docs++
	}

	// 0. Namespace: emitted first so the objects below land in a namespace that
	// exists in the same apply (applying it when it already exists is a no-op).
	sep()
	w.line(0, "apiVersion: v1")
	w.line(0, "kind: Namespace")
	w.line(0, "metadata:")
	w.line(2, "name: "+ns)

	// 1. ConfigMap — one per instance.
	for _, inst := range in.Instances {
		sep()
		renderConfigMap(w, inst.Name+"-config", ns, inst.AppYAML, syslogOf(in.Kube))
	}

	// 2. credentials Secret (stringData) — shared.
	if emitCred {
		sep()
		w.line(0, "apiVersion: v1")
		w.line(0, "kind: Secret")
		w.line(0, "metadata:")
		w.line(2, "name: "+credRef)
		w.line(2, "namespace: "+ns)
		w.line(0, "type: Opaque")
		w.line(0, "stringData:")
		for _, kv := range in.CredKVs {
			w.line(2, kv.Key+": "+strconv.Quote(kv.Val))
		}
	}

	// 3. stores Secret (base64 data) — shared.
	if emitStores {
		sep()
		w.line(0, "apiVersion: v1")
		w.line(0, "kind: Secret")
		w.line(0, "metadata:")
		w.line(2, "name: "+storeRef)
		w.line(2, "namespace: "+ns)
		w.line(0, "type: Opaque")
		w.line(0, "data:")
		for _, sf := range in.Stores {
			w.line(2, sf.Name+": "+sf.Base64)
		}
	}

	// 3b. libs PV + PVC — shared (only for libs.pvc.create; PV is cluster-scoped).
	if lb := in.Kube.Libs; lb != nil && lb.PVC != nil && lb.PVC.Create != nil {
		c := lb.PVC.Create
		sep()
		w.line(0, "apiVersion: v1")
		w.line(0, "kind: PersistentVolume")
		w.line(0, "metadata:")
		w.line(2, "name: "+c.Name+"-pv")
		w.line(0, "spec:")
		w.line(2, "capacity:")
		w.line(4, "storage: "+c.Storage)
		w.line(2, "accessModes:")
		w.line(4, "- ReadWriteMany")
		w.line(2, "nfs:")
		w.line(4, "server: "+c.NFS.Server)
		w.line(4, "path: "+c.NFS.Path)
		w.line(4, "readOnly: true")
		sep()
		w.line(0, "apiVersion: v1")
		w.line(0, "kind: PersistentVolumeClaim")
		w.line(0, "metadata:")
		w.line(2, "name: "+c.Name)
		w.line(2, "namespace: "+ns)
		w.line(0, "spec:")
		w.line(2, `storageClassName: ""`)
		w.line(2, "volumeName: "+c.Name+"-pv")
		w.line(2, "accessModes:")
		w.line(4, "- ReadWriteMany")
		w.line(2, "resources:")
		w.line(4, "requests:")
		w.line(6, "storage: "+c.Storage)
	}

	// 4. Deployment — one per instance.
	for _, inst := range in.Instances {
		sep()
		renderDeployment(w, in, inst, ns, credRef, storeRef, hasStores, mgmtPort)
	}

	// 5. Service — one per instance (when enabled).
	if in.Kube.Service.Enabled {
		for _, inst := range in.Instances {
			sep()
			renderService(w, inst.Name, ns, in.Kube.Service.Port, mgmtPort)
		}
	}

	return w.String()
}

// renderConfigMap emits one ConfigMap embedding this instance's application.yml
// and, when syslog is configured, the shared logback-spring.xml.
func renderConfigMap(w *yw, cmName, ns, appYAML string, sys *spec.Syslog) {
	w.line(0, "apiVersion: v1")
	w.line(0, "kind: ConfigMap")
	w.line(0, "metadata:")
	w.line(2, "name: "+cmName)
	w.line(2, "namespace: "+ns)
	w.line(0, "data:")
	w.line(2, "application.yml: |")
	for _, ln := range splitLines(appYAML) {
		if ln == "" {
			w.raw("\n")
		} else {
			w.line(4, ln)
		}
	}
	if sys != nil {
		w.line(2, "logback-spring.xml: |")
		for _, ln := range splitLines(LogbackXML(sys.Protocol)) {
			if ln == "" {
				w.raw("\n")
			} else {
				w.line(4, ln)
			}
		}
	}
}

func renderDeployment(w *yw, in Input, inst Instance, ns, credRef, storeRef string, hasStores bool, mgmtPort int) {
	dep := in.Kube.Deployment
	name := inst.Name
	cmName := inst.Name + "-config"
	w.line(0, "apiVersion: apps/v1")
	w.line(0, "kind: Deployment")
	w.line(0, "metadata:")
	w.line(2, "name: "+name)
	w.line(2, "namespace: "+ns)
	w.line(0, "spec:")
	w.line(2, "replicas: "+strconv.Itoa(dep.Replicas))
	w.line(2, "selector:")
	w.line(4, "matchLabels:")
	w.line(6, "app: "+name)
	w.line(2, "template:")
	w.line(4, "metadata:")
	w.line(6, "labels:")
	w.line(8, "app: "+name)
	w.line(4, "spec:")
	lb := in.Kube.Libs
	if lb != nil && lb.Download != nil {
		d := lb.Download
		cmds := make([]string, 0, len(d.URLs))
		for _, u := range d.URLs {
			cmds = append(cmds, "wget -O '/libs/"+baseName(u)+"' '"+u+"'")
		}
		w.line(6, "initContainers:")
		w.line(8, "- name: libs-download")
		w.line(10, "image: "+d.Image)
		w.line(10, `command: ["sh", "-c", `+strconv.Quote(strings.Join(cmds, " && "))+`]`)
		w.line(10, "volumeMounts:")
		w.line(12, "- name: libs")
		w.line(14, "mountPath: /libs")
	}
	w.line(6, "containers:")
	w.line(8, "- name: connector")
	w.line(10, "image: "+dep.Image)
	w.line(10, "ports:")
	w.line(12, "- name: management")
	w.line(14, "containerPort: "+strconv.Itoa(mgmtPort))
	// env
	w.line(10, "env:")
	w.line(12, "- name: TZ")
	w.line(14, "value: "+dep.Timezone)
	if inst.Model.MQTLS {
		w.line(12, "- name: JAVA_TOOL_OPTIONS")
		w.line(14, `value: "-Dcom.ibm.mq.cfg.useIBMCipherMappings=false"`)
	}
	if sys := syslogOf(in.Kube); sys != nil {
		w.line(12, "- name: LOGGING_SYSLOG_APPNAME")
		w.line(14, "value: "+name)
		w.line(12, "- name: LOGGING_SYSLOG_HOST")
		w.line(14, "value: "+sys.Host)
		w.line(12, "- name: LOGGING_SYSLOG_PORT")
		w.line(14, `value: "`+strconv.Itoa(sys.Port)+`"`)
	}
	if credRef != "" {
		w.line(10, "envFrom:")
		w.line(12, "- secretRef:")
		w.line(16, "name: "+credRef) // child of secretRef (nested map in list item)
	}
	// volumeMounts
	w.line(10, "volumeMounts:")
	w.line(12, "- name: config")
	w.line(14, "mountPath: /app/external/spring/config/application.yml")
	w.line(14, "subPath: application.yml")
	w.line(14, "readOnly: true")
	if syslogOf(in.Kube) != nil {
		w.line(12, "- name: config")
		w.line(14, "mountPath: /app/external/classpath/logback-spring.xml")
		w.line(14, "subPath: logback-spring.xml")
		w.line(14, "readOnly: true")
	}
	if hasStores {
		w.line(12, "- name: stores")
		w.line(14, "mountPath: /app/external/classpath/truststores")
		w.line(14, "readOnly: true")
	}
	if lb != nil {
		w.line(12, "- name: libs")
		w.line(14, "mountPath: /app/external/libs")
		w.line(14, "readOnly: true")
	}
	// probes (tcpSocket — see prompt.md "Probes"; keep isolated for a later switch)
	w.line(10, "livenessProbe:")
	w.line(12, "tcpSocket:")
	w.line(14, "port: "+strconv.Itoa(mgmtPort))
	w.line(12, "initialDelaySeconds: 30")
	w.line(12, "periodSeconds: 15")
	w.line(10, "readinessProbe:")
	w.line(12, "tcpSocket:")
	w.line(14, "port: "+strconv.Itoa(mgmtPort))
	w.line(12, "initialDelaySeconds: 15")
	w.line(12, "periodSeconds: 10")
	// resources
	renderResources(w, dep.Resources)
	// volumes
	w.line(6, "volumes:")
	w.line(8, "- name: config")
	w.line(10, "configMap:")
	w.line(12, "name: "+cmName)
	if hasStores {
		w.line(8, "- name: stores")
		w.line(10, "secret:")
		w.line(12, "secretName: "+storeRef)
	}
	if lb != nil {
		w.line(8, "- name: libs")
		switch {
		case lb.PVC != nil && lb.PVC.Existing != "":
			w.line(10, "persistentVolumeClaim:")
			w.line(12, "claimName: "+lb.PVC.Existing)
		case lb.PVC != nil && lb.PVC.Create != nil:
			w.line(10, "persistentVolumeClaim:")
			w.line(12, "claimName: "+lb.PVC.Create.Name)
		case lb.Download != nil && lb.Download.PVC != "":
			w.line(10, "persistentVolumeClaim:")
			w.line(12, "claimName: "+lb.Download.PVC)
		default:
			w.line(10, "emptyDir: {}")
		}
	}
}

func renderResources(w *yw, r spec.Resources) {
	if r.CPU == "" && r.Memory == "" {
		return
	}
	w.line(10, "resources:")
	// requests and limits are set identically (guaranteed QoS).
	for _, kind := range []string{"requests", "limits"} {
		w.line(12, kind+":")
		if r.CPU != "" {
			w.line(14, "cpu: "+quoteRes(r.CPU))
		}
		if r.Memory != "" {
			w.line(14, "memory: "+quoteRes(r.Memory))
		}
	}
}

func renderService(w *yw, name, ns string, port, targetPort int) {
	w.line(0, "apiVersion: v1")
	w.line(0, "kind: Service")
	w.line(0, "metadata:")
	w.line(2, "name: "+name)
	w.line(2, "namespace: "+ns)
	w.line(0, "spec:")
	w.line(2, "selector:")
	w.line(4, "app: "+name)
	w.line(2, "ports:")
	w.line(4, "- name: management")
	w.line(6, "port: "+strconv.Itoa(port))
	w.line(6, "targetPort: "+strconv.Itoa(targetPort))
}

// managementPort prefers the configured management port, then the service port,
// then the connector default 8090.
func managementPort(in Input) int {
	if in.Defaults != nil && in.Defaults.Management.Port != 0 {
		return in.Defaults.Management.Port
	}
	if in.Kube.Service.Port != 0 {
		return in.Kube.Service.Port
	}
	return 8090
}

// splitLines splits s on '\n' and drops a single trailing empty element that a
// terminating newline produces, so the block scalar keeps exactly one final newline.
func splitLines(s string) []string {
	lines := strings.Split(s, "\n")
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return lines
}
