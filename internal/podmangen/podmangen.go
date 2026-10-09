// Package podmangen renders the podman deployment artifact for the Solace PubSub+
// Connector for IBM MQ: a systemd .container quadlet unit. It is pure: no os/exec,
// no filesystem, no globals, no network -- it only builds strings. The caller
// resolves credentials, the rendered documents and the TLS stores to
// secret-store references and libs to a host bind mount before calling; exec
// safety is the runner's job, so this package emits the values it is given
// verbatim.
package podmangen

import (
	"strconv"
	"strings"

	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/logback"
	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/spec"
	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/statusscript"
	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/yamlwriter"
)

// appYAMLTarget is the in-container path the application.yml secret is mounted
// at. A quadlet unit cannot inline file content, so the document reaches the
// container from podman's secret store, like the credentials.
const appYAMLTarget = "/app/external/spring/config/application.yml"

// statusTarget is the in-container path the rendered status script's secret is
// mounted at. It comes from statusscript rather than being repeated here, so
// moving the path is one edit instead of four.
const statusTarget = statusscript.ContainerPath

// Instance is the connector: its container name and the secret-store names of
// the documents rendered for it.
type Instance struct {
	Name     string // container name
	Image    string // the reference to pull, from the top-level image: block
	Timezone string // container TZ, from the top-level timezone: key
	// AppYAMLSecret is the secret-store name holding the rendered
	// application.yml, mounted at appYAMLTarget; empty omits the mount.
	AppYAMLSecret string
	MQTLS         bool // when true, add JAVA_TOOL_OPTIONS env for IBM cipher mappings
	// StatusScriptSecret is the secret-store name holding the rendered status
	// script; empty omits the mount and the healthcheck that runs it.
	StatusScriptSecret string
	// LogbackSecret is the secret-store name holding the rendered
	// logback-spring.xml; empty (no syslog) omits the mount.
	LogbackSecret string
	LeaderMode    string // leader-election mode; empty means standalone (see leaderLabels)
	// JavaOptions is the top-level java-options: block, nil when absent. It is
	// merged with the MQTLS flag by spec.JavaEnv, as on every platform.
	JavaOptions *spec.JavaOptions
}

// Mount is one read-only bind mount (host path -> container path). Only the
// libs directory is one: everything else comes from the secret store.
type Mount struct {
	Source string // host path
	Target string // absolute container path
}

// FileSecret is one file mounted out of podman's secret store at an absolute
// in-container path: a TLS store, at the path application.yml names it by.
type FileSecret struct {
	StoreName string
	Target    string // absolute container path
}

// SecretRef is one credential from podman's secret store, mounted into the
// container. StoreName is how podman knows it (namespaced by container name,
// since the store is shared across every project on the host); Target is the
// stable name the connector's config references, which is also the file name it
// appears under in spec.SecretsMountPath.
type SecretRef struct {
	StoreName string
	Target    string
}

// Input is everything the podman renderers need. The caller resolves
// credentials and stores to secret-store references and libs to a host bind
// mount before calling.
type Input struct {
	Podman   *spec.Podman
	Instance Instance
	// Syslog is the top-level logging.syslog block, nil when absent. It supplies
	// the three env vars the mounted logback config reads at runtime.
	Syslog  *spec.Syslog
	Secrets []SecretRef  // credentials from podman's secret store; nil/empty when none
	Stores  []FileSecret // TLS stores from podman's secret store; nil/empty when none
	Libs    *Mount       // libs dir bind mount (read-only); nil when none
}

// Unit is one rendered quadlet file.
type Unit struct {
	Filename string // e.g. "solmq-connector.container"
	Content  string
}

// sw is the shared line writer; podman's lines are never nested, so every
// call below passes indent 0 to Line.
type sw = yamlwriter.Writer

// systemdEnv renders one Environment= assignment for an operator-supplied value.
// systemd expands % specifiers in unit files -- including in what quadlet
// passes on to the generated service -- so a literal '%' (the %p in
// -XX:ErrorFile=hs_err_%p.log) is doubled, or the container would get the
// unit's name in its place. An assignment carrying a space has to be quoted
// whole, or systemd splits it into several assignments. A value with neither
// renders exactly as the unquoted form always has. Quotes and backslashes are
// escaped too, though validate rejects both before anything renders.
func systemdEnv(name, value string) string {
	v := strings.ReplaceAll(value, "%", "%%")
	if !strings.ContainsAny(v, " \t\"'\\") {
		return "Environment=" + name + "=" + v
	}
	return `Environment="` + name + "=" + strings.NewReplacer(`\`, `\\`, `"`, `\"`).Replace(v) + `"`
}

// secretMount writes one Secret= line mounting the named secret as a file at
// the absolute target. No mode= is set: podman's default, 0444 owned by root,
// matches what compose gives a config and what the credentials already get. An
// empty name writes nothing.
func secretMount(w *sw, name, target string) {
	if name == "" {
		return
	}
	w.Line(0, "Secret="+name+",type=mount,target="+target)
}

// seconds spells a duration the way the quadlet Health* keys want it.
func seconds(n int) string { return strconv.Itoa(n) + "s" }

// leaderLabels returns the ordered (key, value) label pairs the unit carries.
// The mode label is always present. The role label marks this
// instance active only when that is knowable at render time: standalone
// always is, and every active_active member is; active_standby's active side
// flips at runtime, so it never gets a static role label.
func leaderLabels(mode string) [][2]string {
	if mode == "" {
		mode = spec.LeaderStandalone
	}
	labels := [][2]string{{spec.LabelModeKey, mode}}
	if mode == spec.LeaderStandalone || mode == spec.LeaderActiveActive {
		labels = append(labels, [2]string{spec.LabelRoleKey, spec.LabelRoleActive})
	}
	return labels
}

// RenderQuadlet returns the .container quadlet unit; the filename is
// "<Name>.container". Sections are emitted in order [Unit], [Container],
// [Service], [Install], separated by blank lines. The [Service] section is
// omitted entirely when Restart is empty.
func RenderQuadlet(in Input) Unit {
	p := in.Podman
	inst := in.Instance
	w := &sw{}
	w.Line(0, "# Generated by solmq-conn-util -- podman quadlet unit for the Solace PubSub+ Connector for IBM MQ.")

	w.Line(0, "[Unit]")
	w.Line(0, "Description=Solace PubSub+ Connector for IBM MQ ("+inst.Name+")")
	w.Line(0, "After=network-online.target")
	w.Line(0, "Wants=network-online.target")
	w.Line(0, "")

	w.Line(0, "[Container]")
	w.Line(0, "Image="+inst.Image)
	w.Line(0, "ContainerName="+inst.Name)
	for _, l := range leaderLabels(inst.LeaderMode) {
		w.Line(0, "Label="+l[0]+"="+l[1])
	}
	for _, port := range p.Ports {
		w.Line(0, "PublishPort="+port.String())
	}
	if inst.Timezone != "" {
		w.Line(0, "Environment=TZ="+inst.Timezone)
	}
	tool, jdk := spec.JavaEnv(inst.MQTLS, inst.JavaOptions)
	if tool != "" {
		w.Line(0, systemdEnv("JAVA_TOOL_OPTIONS", tool))
	}
	if jdk != "" {
		w.Line(0, systemdEnv("JDK_JAVA_OPTIONS", jdk))
	}
	if sl := in.Syslog; sl != nil {
		w.Line(0, "Environment=LOGGING_SYSLOG_APPNAME="+inst.Name)
		w.Line(0, "Environment=LOGGING_SYSLOG_HOST="+sl.Host)
		w.Line(0, "Environment=LOGGING_SYSLOG_PORT="+strconv.Itoa(sl.Port))
	}
	for _, s := range in.Secrets {
		// An absolute target rather than a bare file name: bare would leave the
		// secret in podman's default /run/secrets, the directory
		// spec.SecretsMountPath exists to avoid. Needs podman 4.x or newer, where
		// target= accepts a path.
		secretMount(w, s.StoreName, spec.SecretsMountPath+"/"+s.Target)
	}
	// The rendered documents and the TLS stores come from the secret store too,
	// each at the fixed path the image or application.yml reads it from, so the
	// unit names no host file: the libs directory below is the one bind mount.
	secretMount(w, inst.AppYAMLSecret, appYAMLTarget)
	for _, st := range in.Stores {
		secretMount(w, st.StoreName, st.Target)
	}
	secretMount(w, inst.StatusScriptSecret, statusTarget)
	secretMount(w, inst.LogbackSecret, logback.ContainerPath)
	if in.Libs != nil {
		w.Line(0, "Volume="+in.Libs.Source+":"+in.Libs.Target+":ro")
	}
	// The healthcheck runs the status script in its --health mode, so it is
	// emitted only when that script is actually mounted -- unlike compose,
	// where the mount is unconditional. Without these keys podman populates no
	// .State.Health and `status container` reports n/a in the HEALTH column.
	//
	// HealthOnFailure is deliberately left unset (podman defaults it to none):
	// the check reports a verdict and never acts on it, matching the compose
	// side and leaving what to do about an unhealthy instance to the operator.
	// Podman drives the check from a transient systemd timer in whichever
	// scope the unit lives in, so this works the same rootful and rootless.
	//
	// These keys need podman 4.5+, which is already this tool's floor.
	if inst.StatusScriptSecret != "" {
		w.Line(0, "HealthCmd="+statusscript.HealthShell+" "+statusTarget+" "+statusscript.HealthArg)
		w.Line(0, "HealthInterval="+seconds(statusscript.HealthIntervalSeconds))
		w.Line(0, "HealthTimeout="+seconds(statusscript.HealthTimeoutSeconds))
		w.Line(0, "HealthRetries="+strconv.Itoa(statusscript.HealthRetries))
		w.Line(0, "HealthStartPeriod="+seconds(statusscript.HealthStartPeriodSeconds))
	}
	w.Line(0, "")

	if p.Restart != "" {
		w.Line(0, "[Service]")
		w.Line(0, "Restart="+p.Restart)
		w.Line(0, "")
	}

	w.Line(0, "[Install]")
	w.Line(0, "WantedBy=default.target")

	return Unit{Filename: inst.Name + ".container", Content: w.String()}
}
