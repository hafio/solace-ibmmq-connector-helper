// Package dockergen renders a docker-compose.yml for the Solace PubSub+
// Connector for IBM MQ. It is pure: no os/exec, no filesystem, no globals, no
// network -- the caller resolves credentials to their stable names and
// stores/libs to host bind mounts, then this package just builds the compose
// string.
package dockergen

import (
	"strings"

	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/spec"
	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/yamlwriter"
)

// Instance is the connector: its compose service name and rendered config.
type Instance struct {
	Name         string // service/container name
	Image        string // the reference to pull, from the top-level image: block
	Timezone     string // container TZ, from the top-level timezone: key
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
	renderContentConfig(w, in.Instance.Name+"-app", in.Instance.AppYAML)
	renderContentConfig(w, in.Instance.Name+"-status", in.Instance.StatusScript)

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
	w.Line(4, "image: "+inst.Image)
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
	if inst.Timezone != "" || inst.MQTLS {
		w.Line(4, "environment:")
		if inst.Timezone != "" {
			w.Line(6, "TZ: "+inst.Timezone)
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

// renderContentConfig emits one top-level configs entry inlining payload as a
// block scalar under content: -- the application.yml and the status script are
// both rendered through it. Blank lines are preserved as truly empty lines (no
// indent, no trailing spaces), and every line is composeEscape'd.
//
// The two entries share one renderer so neither can be escaped without the
// other: an unescaped content block is not a cosmetic difference but a broken
// deploy (the status script) or a leaked credential (application.yml).
func renderContentConfig(w *yw, name, payload string) {
	w.Line(2, name+":")
	w.Line(4, "content: |")
	for _, ln := range yamlwriter.SplitLines(payload) {
		if ln == "" {
			w.Raw("\n")
		} else {
			w.Line(6, composeEscape(ln))
		}
	}
}

// composeEscape doubles every '$'. Docker Compose interpolates the whole
// document -- configs content included -- and renders "$$" back to a single
// "$", so this is what makes the file the container receives identical to what
// was generated. It applies to the inlined content only; the rest of the
// document is built from charset-validated values that cannot carry a '$'.
//
// Without it, both content blocks are corrupted:
//
//   - The status script is shell. Its variables ($PORT, $CONFIGS, ...) are
//     replaced with blanks, and compose refuses the document outright on the
//     first $( -- command substitution is not valid interpolation syntax.
//   - application.yml's ${...} placeholders are Spring's, resolved from the
//     configtree import of /run/secrets. Compose resolves them first, from the
//     environment the CLI hands the compose child (runner.Cmd.Env), inlining
//     the credential values as plaintext into the document and bypassing the
//     secrets model entirely.
//
// Only the compose renderer needs this: the kubernetes ConfigMap and the podman
// quadlet (which writes both files to disk beside the unit) have no
// interpolation layer between this generator and the container.
func composeEscape(s string) string { return strings.ReplaceAll(s, "$", "$$") }
