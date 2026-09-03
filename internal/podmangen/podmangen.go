// Package podmangen renders the podman deployment artifact for the Solace PubSub+
// Connector for IBM MQ: a systemd .container quadlet unit. It is pure: no os/exec,
// no filesystem, no globals, no network -- it only builds strings. The caller
// resolves credentials to secret-store references and stores/libs to host bind
// mounts before calling; exec safety is the runner's job, so this package emits
// the values it is given verbatim.
package podmangen

import (
	"strconv"

	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/logback"
	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/spec"
	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/statusscript"
	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/yamlwriter"
)

// appYAMLTarget is the in-container path the application.yml is bind-mounted to
// (read-only); podman cannot inline file content.
const appYAMLTarget = "/app/external/spring/config/application.yml"

// statusTarget is the in-container path the rendered status script is
// bind-mounted to (read-only); like application.yml, podman cannot inline
// file content, so the script has to be a bind mount too. It comes from
// statusscript rather than being repeated here, so moving the path is one edit
// instead of four.
const statusTarget = statusscript.ContainerPath

// javaToolOptions is the JAVA_TOOL_OPTIONS value set when the connector uses MQ TLS,
// selecting the IBM cipher mappings the connector expects.
const javaToolOptions = "-Dcom.ibm.mq.cfg.useIBMCipherMappings=false"

// Instance is the connector: its container name and on-disk config path.
type Instance struct {
	Name             string // container name
	Image            string // the reference to pull, from the top-level image: block
	Timezone         string // container TZ, from the top-level timezone: key
	AppYAMLPath      string // host path to the application.yml on disk (bind-mounted)
	MQTLS            bool   // when true, add JAVA_TOOL_OPTIONS env for IBM cipher mappings
	StatusScriptPath string // host path to the rendered status script on disk (bind-mounted); empty omits the mount
	// LogbackPath is the host path to the rendered logback-spring.xml
	// (bind-mounted); empty omits the mount. podman cannot inline file content,
	// so unlike compose this has to exist on disk before the unit starts.
	LogbackPath string
	LeaderMode  string // leader-election mode; empty means standalone (see leaderLabels)
}

// Mount is one read-only bind mount (host path -> container path).
type Mount struct {
	Source string // host path
	Target string // absolute container path
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

// Input is everything the podman renderers need. The caller resolves stores/libs
// to host bind mounts and credentials to secret-store references before calling.
type Input struct {
	Podman   *spec.Podman
	Instance Instance
	// Syslog is the top-level logging.syslog block, nil when absent. It supplies
	// the three env vars the mounted logback config reads at runtime.
	Syslog  *spec.Syslog
	Secrets []SecretRef // podman secret store entries to mount; nil/empty when none
	Stores  []Mount     // .jks bind mounts (read-only); nil/empty when none
	Libs    *Mount      // libs dir bind mount (read-only); nil when none
}

// Unit is one rendered quadlet file.
type Unit struct {
	Filename string // e.g. "solmq-connector.container"
	Content  string
}

// sw is the shared line writer; podman's lines are never nested, so every
// call below passes indent 0 to Line.
type sw = yamlwriter.Writer

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
	if inst.MQTLS {
		w.Line(0, "Environment=JAVA_TOOL_OPTIONS="+javaToolOptions)
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
		w.Line(0, "Secret="+s.StoreName+",type=mount,target="+spec.SecretsMountPath+"/"+s.Target)
	}
	w.Line(0, "Volume="+inst.AppYAMLPath+":"+appYAMLTarget+":ro")
	for _, m := range in.Stores {
		w.Line(0, "Volume="+m.Source+":"+m.Target+":ro")
	}
	if in.Libs != nil {
		w.Line(0, "Volume="+in.Libs.Source+":"+in.Libs.Target+":ro")
	}
	if inst.StatusScriptPath != "" {
		// Emitted after the libs volume so this single-file mount nests inside it
		// instead of being shadowed by a libs directory mount at the same path.
		w.Line(0, "Volume="+inst.StatusScriptPath+":"+statusTarget+":ro")
	}
	if inst.LogbackPath != "" {
		w.Line(0, "Volume="+inst.LogbackPath+":"+logback.ContainerPath+":ro")
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
