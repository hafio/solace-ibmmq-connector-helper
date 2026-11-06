// Package dockergen renders a docker-compose.yml for the Solace PubSub+
// Connector for IBM MQ. It is pure: no os/exec, no filesystem, no globals, no
// network -- the caller resolves credentials to their stable names and
// stores/libs to host bind mounts, then this package just builds the compose
// string.
package dockergen

import (
	"strconv"

	"github.com/hafio/go-deploy-utility/containerart/compose"
	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/logback"
	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/spec"
	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/statusscript"
	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/storescript"
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

	// JavaOptions is the top-level java-options: block, nil when absent. It is
	// merged with the MQTLS flag by spec.JavaEnv, as on every platform.
	JavaOptions *spec.JavaOptions
	// RunAs is the identity the container runs as, from docker.run-as. Its
	// zero value is root, so the caller always sets it.
	RunAs spec.RunAsID
}

// Mount is one read-only bind mount (host path -> container path).
type Mount struct {
	Source string // host path, used verbatim (caller has resolved it)
	Target string // absolute container path
}

// CreatedStore is a created store the service carries: its tls.stores name,
// and the compose secret -- read from the environment variable of the same
// name -- its base64-encoded bytes travel in.
type CreatedStore struct {
	Name   string
	Secret string
}

// StoresDigestVar is the environment variable compose fills the
// spec.LabelStoresDigestKey label from: deploy sets it to the sha256 of the
// created stores, so a changed store changes the service and compose recreates
// the container.
const StoresDigestVar = "_GEN_TLS_STORES_SHA256"

// Input is everything the compose renderer needs. The caller resolves stores/libs
// to host bind mounts and reduces credentials to their stable names before
// calling.
type Input struct {
	Docker   *spec.Docker
	Instance Instance
	// Syslog is the top-level logging.syslog block, nil when absent. Compose can
	// inline file content, so the logback config it implies is a third configs
	// entry rather than a file on disk.
	Syslog  *spec.Syslog
	Secrets []string // stable secret names; nil/empty when there are no credentials
	Stores  []Mount  // existing stores' bind mounts (read-only); nil/empty when none
	Libs    *Mount   // libs dir bind mount (read-only); nil when no libs
	// Created are the created stores the service carries, decoded at every
	// start by the storescript entrypoint; nil when none, and then nothing
	// about the service changes.
	Created []CreatedStore
	// Digest labels the service with the stores digest: some store, created
	// or existing, is selected.
	Digest bool
}

// yw is the indentation-aware line writer used by every renderer here.
type yw = yamlwriter.Writer

// Render returns the full docker-compose.yml. The application.yml and the
// status script are each inlined via their own compose top-level configs entry
// with content:, mounted into the service at /app/external/spring/config/application.yml
// and statusscript.ContainerPath respectively.
//
// Every name written here -- the project, the service, the container and each
// config and secret -- goes through compose.Quote, so a name YAML would read as
// a boolean or a number stays the string the operator wrote.
func Render(in Input) string {
	w := &yw{}

	// Omitted when empty, like every other optional line here, so a caller that
	// builds a spec.Docker without going through ParseEnv renders as before.
	if in.Docker.ProjectName != "" {
		w.Line(0, "name: "+compose.Quote(in.Docker.ProjectName))
	}
	w.Line(0, "services:")
	renderService(w, in, in.Instance)

	w.Line(0, "configs:")
	renderContentConfig(w, in.Instance.Name+"-app", in.Instance.AppYAML)
	renderContentConfig(w, in.Instance.Name+"-status", in.Instance.StatusScript)
	if in.Syslog != nil {
		renderContentConfig(w, in.Instance.Name+"-logback", logback.XML(in.Syslog.Protocol))
	}
	if len(in.Created) > 0 {
		renderContentConfig(w, in.Instance.Name+"-stores-init", storescript.Render())
	}

	// The top-level secrets map. Each entry uses compose's environment
	// provider, so the value is read from the environment `docker compose`
	// itself runs with -- it is never written to a file by this tool, and never
	// appears in the compose document. Requires Docker Compose v2.23.1 or
	// newer. Where each one lands in the container is the service's business;
	// see renderService.
	w.Raw(compose.TopSecrets(secretList(in)))

	return w.String()
}

// secretList is the service's secrets in the order both the service block and
// the top-level map list them: the credentials, each mounted at
// spec.SecretsMountPath/<name> where the connector's configtree import reads it
// as a property, then the created stores, each encoded under its store name in
// spec.EncodedStoresMountPath for the storescript to decode. Every secret is
// read from the environment variable of its own name.
func secretList(in Input) []compose.Secret {
	var out []compose.Secret
	for _, n := range in.Secrets {
		out = append(out, compose.Secret{Name: n, Target: spec.SecretsMountPath + "/" + n, EnvVar: n})
	}
	for _, c := range in.Created {
		out = append(out, compose.Secret{Name: c.Secret, Target: spec.EncodedStoresMountPath + "/" + c.Name, EnvVar: c.Secret})
	}
	return out
}

// renderService emits the service block: image, container name, labels, and
// the conditional restart / ports / environment / secrets / configs / volumes
// sub-blocks. A sub-block whose contents would be empty is omitted entirely so
// no dangling key with a null value is produced.
func renderService(w *yw, in Input, inst Instance) {
	d := in.Docker
	w.Line(2, compose.Quote(inst.Name)+":")
	w.Line(4, "image: "+inst.Image)
	w.Line(4, "container_name: "+compose.Quote(inst.Name))
	// A service carrying created stores starts through the storescript, which
	// decodes them and then execs the image's own command, handed over here --
	// setting entrypoint: drops the image's CMD too.
	if len(in.Created) > 0 {
		w.Line(4, "entrypoint: "+compose.FlowList([]string{storescript.Shell, storescript.ContainerPath}))
		w.Line(4, "command: "+compose.FlowList(storescript.Command()))
	}
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
	// Left for compose to interpolate from deploy's environment: the stores
	// are not in this file, so their digest cannot be either. The :- default
	// keeps a hand-run compose up silent, with an empty label.
	if in.Digest {
		w.Line(6, spec.LabelStoresDigestKey+": ${"+StoresDigestVar+":-}")
	}
	if d.Restart != "" {
		w.Line(4, "restart: "+d.Restart)
	}
	// Identity and privileges are stated, never inherited. The image declares
	// no USER, so without user: the connector would run as root -- host root,
	// since docker runs it without a user namespace. run-as: image leaves out
	// the user: line alone; the rest of the posture holds either way, and the
	// connector needs no capability (it binds an unprivileged port).
	if !inst.RunAs.Image {
		w.Line(4, "user: "+compose.Quote(inst.RunAs.String()))
	}
	w.Raw(compose.Posture)
	if len(d.Ports) > 0 {
		w.Line(4, "ports:")
		for _, p := range d.Ports {
			w.Line(6, "- "+compose.Quote(p.String()))
		}
	}
	// environment: TZ only when set, each JVM options variable only when it has
	// something in it -- the MQ TLS flag, the operator's java-options, or both.
	// Omit the whole key when nothing applies.
	tool, jdk := spec.JavaEnv(inst.MQTLS, inst.JavaOptions)
	if inst.Timezone != "" || tool != "" || jdk != "" || in.Syslog != nil {
		w.Line(4, "environment:")
		if inst.Timezone != "" {
			w.Line(6, "TZ: "+compose.Quote(inst.Timezone))
		}
		if tool != "" {
			w.Line(6, "JAVA_TOOL_OPTIONS: "+compose.Quote(tool))
		}
		if jdk != "" {
			w.Line(6, "JDK_JAVA_OPTIONS: "+compose.Quote(jdk))
		}
		// The logback config reads host/port/appname from these at runtime, the
		// same three the kubernetes Deployment sets.
		if s := in.Syslog; s != nil {
			w.Line(6, "LOGGING_SYSLOG_APPNAME: "+compose.Quote(inst.Name))
			w.Line(6, "LOGGING_SYSLOG_HOST: "+compose.Quote(s.Host))
			w.Line(6, "LOGGING_SYSLOG_PORT: "+compose.Quote(strconv.Itoa(s.Port)))
		}
	}
	// secrets: long syntax, not the one-line `- name` form, because the short
	// form mounts at compose's own /run/secrets/<name> with no way to say
	// otherwise, and /run/secrets is exactly the directory this tool had to move
	// off (see spec.SecretsMountPath). An absolute target is what compose takes
	// instead. See secretList for where each one lands.
	w.Raw(compose.ServiceSecrets(secretList(in)))
	// configs: the inlined application.yml and status script, plus the logback
	// config when syslog is configured.
	// Neither target is nested inside a volume mount below, so the engine's
	// mount order between configs and volumes cannot shadow either of them. That
	// was not true while the status script lived under /app/external/libs, where
	// a libs volume mounted over it (see statusscript.ContainerDir).
	w.Line(4, "configs:")
	w.Line(6, "- source: "+compose.Quote(inst.Name+"-app"))
	w.Line(8, "target: /app/external/spring/config/application.yml")
	w.Line(6, "- source: "+compose.Quote(inst.Name+"-status"))
	w.Line(8, "target: "+statusscript.ContainerPath)
	if in.Syslog != nil {
		w.Line(6, "- source: "+compose.Quote(inst.Name+"-logback"))
		w.Line(8, "target: "+logback.ContainerPath)
	}
	if len(in.Created) > 0 {
		w.Line(6, "- source: "+compose.Quote(inst.Name+"-stores-init"))
		w.Line(8, "target: "+storescript.ContainerPath)
		// The decoded stores live in memory, owned by the identity the
		// connector runs as and closed to everyone else; the existing stores
		// bind-mounted below land inside it, the engine mounting the
		// directory first.
		opts := "mode=0700"
		if !inst.RunAs.Image {
			opts += ",uid=" + strconv.Itoa(inst.RunAs.UID) + ",gid=" + strconv.Itoa(inst.RunAs.GID)
		}
		w.Line(4, "tmpfs:")
		w.Line(6, "- "+spec.DefaultStoresMountPath+":"+opts)
	}
	renderHealthcheck(w)
	// volumes: stores first, then libs; omit the key when there are neither.
	if len(in.Stores) > 0 || in.Libs != nil {
		w.Line(4, "volumes:")
		for _, s := range in.Stores {
			w.Line(6, "- "+compose.Quote(s.Source+":"+s.Target+":ro"))
		}
		if in.Libs != nil {
			w.Line(6, "- "+compose.Quote(in.Libs.Source+":"+in.Libs.Target+":ro"))
		}
	}
}

// renderHealthcheck emits the service's healthcheck block: the status script
// in its --health mode, which answers with an exit status instead of a report.
// Without it the engine has no healthcheck to run and `status container`
// reports n/a in the HEALTH column, since docker then populates no
// .State.Health at all.
//
// It is emitted unconditionally because the status script's configs entry is
// too (see renderService): compose always mounts the script, so the check can
// always reach it. The test is the exec form rather than CMD-SHELL -- there is
// no shell syntax here to need one, and the exec form cannot be reinterpreted
// by compose's own interpolation.
//
// Deliberately no autoheal or restart behaviour hangs off this: the check
// reports, and what to do about an unhealthy instance stays the operator's
// call.
func renderHealthcheck(w *yw) {
	w.Line(4, "healthcheck:")
	w.Line(6, "test: "+compose.FlowList([]string{"CMD", statusscript.HealthShell, statusscript.ContainerPath, statusscript.HealthArg}))
	w.Line(6, "interval: "+seconds(statusscript.HealthIntervalSeconds))
	w.Line(6, "timeout: "+seconds(statusscript.HealthTimeoutSeconds))
	w.Line(6, "retries: "+strconv.Itoa(statusscript.HealthRetries))
	w.Line(6, "start_period: "+seconds(statusscript.HealthStartPeriodSeconds))
}

// seconds spells a duration the way compose wants it.
func seconds(n int) string { return strconv.Itoa(n) + "s" }

// renderContentConfig emits one top-level configs entry inlining payload as a
// block scalar under content: -- the application.yml and the status script are
// both rendered through it. Blank lines are preserved as truly empty lines (no
// indent, no trailing spaces), and every line goes through compose.Escape, which
// says why: unescaped, the status script breaks the deploy and application.yml's
// placeholders leak the credentials. The entries share one renderer so neither
// can be left unescaped.
func renderContentConfig(w *yw, name, payload string) {
	w.Line(2, compose.Quote(name)+":")
	w.Line(4, "content: |")
	for _, ln := range yamlwriter.SplitLines(payload) {
		if ln == "" {
			w.Raw("\n")
		} else {
			w.Line(6, compose.Escape(ln))
		}
	}
}
