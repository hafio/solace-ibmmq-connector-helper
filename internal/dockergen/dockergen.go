// Package dockergen renders a docker-compose.yml for the Solace PubSub+
// Connector for IBM MQ. It is pure: no os/exec, no filesystem, no globals, no
// network -- the caller resolves credentials to their stable names and
// stores/libs to host bind mounts, then this package just builds the compose
// string.
package dockergen

import (
	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/spec"
	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/yamlwriter"
)

// Instance is the connector: its compose service name and rendered config.
type Instance struct {
	Name         string // service/container name
	AppYAML      string // the rendered application.yml (has a trailing newline)
	MQTLS        bool   // when true, add JAVA_TOOL_OPTIONS for IBM cipher mappings
	StatusScript string // the rendered status script, inlined as a second compose config
	LeaderMode   string // effective leader-election mode; empty means standalone (normalized in the renderer)
}

// Mount is one read-only bind mount (host path -> container path).
type Mount struct {
	Source string // host path, used verbatim (caller has resolved it)
	Target string // absolute container path
}

// Input is everything the compose renderer needs. The caller resolves stores/libs
// to host bind mounts and reduces credentials to their stable names before
// calling.
type Input struct {
	Docker   *spec.Docker
	Instance Instance
	Secrets  []string // stable secret names; nil/empty when there are no credentials
	Stores   []Mount  // .jks bind mounts (read-only); nil/empty when no stores
	Libs     *Mount   // libs dir bind mount (read-only); nil when no libs
}

// yw is the indentation-aware line writer used by every renderer here.
type yw = yamlwriter.Writer

// Render returns the full docker-compose.yml. The application.yml and the
// status script are each inlined via their own compose top-level configs entry
// with content:, mounted into the service at /app/external/spring/config/application.yml
// and /app/external/libs/status respectively.
func Render(in Input) string {
	w := &yw{}

	w.Line(0, "services:")
	renderService(w, in, in.Instance)

	w.Line(0, "configs:")
	renderConfig(w, in.Instance)
	renderStatusScriptConfig(w, in.Instance)

	renderSecrets(w, in.Secrets)

	return w.String()
}

// renderSecrets emits the top-level secrets map. Each entry uses compose's
// environment provider, so the value is read from the environment `docker
// compose` itself runs with and mounted at /run/secrets/<name> -- it is never
// written to a file by this tool, and never appears in the compose document.
// Requires Docker Compose v2.23.1 or newer.
func renderSecrets(w *yw, names []string) {
	if len(names) == 0 {
		return
	}
	w.Line(0, "secrets:")
	for _, n := range names {
		w.Line(2, n+":")
		w.Line(4, "environment: "+n)
	}
}

// renderService emits the service block: image, container name, labels, and
// the conditional restart / ports / environment / secrets / configs / volumes
// sub-blocks. A sub-block whose contents would be empty is omitted entirely so
// no dangling key with a null value is produced.
func renderService(w *yw, in Input, inst Instance) {
	d := in.Docker
	w.Line(2, inst.Name+":")
	w.Line(4, "image: "+d.Image)
	w.Line(4, "container_name: "+inst.Name)
	// labels: le-mode is always set; role: active only for standalone and
	// active_active -- an active_standby role is only knowable live from the
	// actuator, so it is never asserted statically here.
	mode := inst.LeaderMode
	if mode == "" {
		mode = spec.LeaderStandalone
	}
	w.Line(4, "labels:")
	w.Line(6, spec.LabelModeKey+": "+mode)
	if mode == spec.LeaderStandalone || mode == spec.LeaderActiveActive {
		w.Line(6, spec.LabelRoleKey+": "+spec.LabelRoleActive)
	}
	if d.Restart != "" {
		w.Line(4, "restart: "+d.Restart)
	}
	if len(d.Ports) > 0 {
		w.Line(4, "ports:")
		for _, p := range d.Ports {
			w.Line(6, `- "`+p.String()+`"`)
		}
	}
	// environment: TZ only when set, JAVA_TOOL_OPTIONS only when the connector
	// uses MQ TLS. Omit the whole key when neither applies.
	if d.Timezone != "" || inst.MQTLS {
		w.Line(4, "environment:")
		if d.Timezone != "" {
			w.Line(6, "TZ: "+d.Timezone)
		}
		if inst.MQTLS {
			w.Line(6, `JAVA_TOOL_OPTIONS: "-Dcom.ibm.mq.cfg.useIBMCipherMappings=false"`)
		}
	}
	// secrets: each is mounted at /run/secrets/<name>, where the connector's
	// configtree import reads it as a property.
	if len(in.Secrets) > 0 {
		w.Line(4, "secrets:")
		for _, n := range in.Secrets {
			w.Line(6, "- "+n)
		}
	}
	// configs: always two entries -- the inlined application.yml and status script.
	//
	// The status script's target sits inside the libs directory a libs volume
	// mounts below, and unlike the kubernetes and podman renderers -- where
	// emitting the file mount last keeps it nested -- compose has no ordering to
	// control here: configs and volumes are separate fields, both handed to the
	// engine, which mounts by destination depth so a parent lands before its
	// child. That makes the nesting the engine's job rather than this file's, so
	// verify it against a real engine when both are configured (see the docker
	// smoke test in userguide.md) rather than trusting key order.
	w.Line(4, "configs:")
	w.Line(6, "- source: "+inst.Name+"-app")
	w.Line(8, "target: /app/external/spring/config/application.yml")
	w.Line(6, "- source: "+inst.Name+"-status")
	w.Line(8, "target: /app/external/libs/status")
	// volumes: stores first, then libs; omit the key when there are neither.
	if len(in.Stores) > 0 || in.Libs != nil {
		w.Line(4, "volumes:")
		for _, s := range in.Stores {
			w.Line(6, "- "+s.Source+":"+s.Target+":ro")
		}
		if in.Libs != nil {
			w.Line(6, "- "+in.Libs.Source+":"+in.Libs.Target+":ro")
		}
	}
}

// renderConfig emits the top-level configs entry inlining the application.yml as
// a block scalar under content:. Blank lines are preserved as
// truly empty lines (no indent, no trailing spaces).
func renderConfig(w *yw, inst Instance) {
	w.Line(2, inst.Name+"-app:")
	w.Line(4, "content: |")
	for _, ln := range yamlwriter.SplitLines(inst.AppYAML) {
		if ln == "" {
			w.Raw("\n")
		} else {
			w.Line(6, ln)
		}
	}
}

// renderStatusScriptConfig emits the top-level configs entry inlining the
// status script as a block scalar under content:, mirroring renderConfig.
// Blank lines are preserved as truly empty lines (no indent, no trailing
// spaces).
func renderStatusScriptConfig(w *yw, inst Instance) {
	w.Line(2, inst.Name+"-status:")
	w.Line(4, "content: |")
	for _, ln := range yamlwriter.SplitLines(inst.StatusScript) {
		if ln == "" {
			w.Raw("\n")
		} else {
			w.Line(6, ln)
		}
	}
}
