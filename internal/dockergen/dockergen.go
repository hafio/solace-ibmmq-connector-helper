// Package dockergen renders a docker-compose.yml for the Solace PubSub+
// Connector for IBM MQ. It is pure: no os/exec, no filesystem, no globals, no
// network -- the caller resolves credentials to an env-file and stores/libs to
// host bind mounts, then this package just builds the compose string.
package dockergen

import (
	"strings"

	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/spec"
)

// Instance is one connector shard (a folder with more than the per-instance
// workflow cap is sharded into several shards).
type Instance struct {
	Name    string // service/container name for this shard (already suffixed by caller if sharded)
	AppYAML string // this shard's rendered application.yml (has a trailing newline)
	MQTLS   bool   // when true, add JAVA_TOOL_OPTIONS for IBM cipher mappings
}

// Mount is one read-only bind mount (host path -> container path).
type Mount struct {
	Source string // host path, used verbatim (caller has resolved it)
	Target string // absolute container path
}

// Input is everything the compose renderer needs. The caller resolves creds to
// an env-file and stores/libs to host bind mounts before calling.
type Input struct {
	Docker    *spec.Docker
	Instances []Instance // one or more shards; >=1
	EnvFile   string     // env_file entry (a path); "" when there are no credentials
	Stores    []Mount    // .jks bind mounts (read-only); nil/empty when no stores
	Libs      *Mount     // libs dir bind mount (read-only); nil when no libs
}

type yw struct{ b strings.Builder }

func (w *yw) line(indent int, s string) {
	w.b.WriteString(strings.Repeat(" ", indent))
	w.b.WriteString(s)
	w.b.WriteByte('\n')
}
func (w *yw) raw(s string)   { w.b.WriteString(s) }
func (w *yw) String() string { return w.b.String() }

// Render returns the full docker-compose.yml. Each instance's application.yml is
// inlined via a compose top-level configs entry with content:, mounted into the
// service at /app/external/spring/config/application.yml. Instance order is
// preserved across both the services and the configs maps.
func Render(in Input) string {
	w := &yw{}

	w.line(0, "services:")
	for _, inst := range in.Instances {
		renderService(w, in, inst)
	}

	// configs: only when there is at least one instance to inline (always the
	// case for a valid Input, since every instance carries its own config).
	if len(in.Instances) > 0 {
		w.line(0, "configs:")
		for _, inst := range in.Instances {
			renderConfig(w, inst)
		}
	}

	return w.String()
}

// renderService emits one service block for a shard: image, container name, and
// the conditional restart / ports / environment / env_file / configs / volumes
// sub-blocks. A sub-block whose contents would be empty is omitted entirely so
// no dangling key with a null value is produced.
func renderService(w *yw, in Input, inst Instance) {
	d := in.Docker
	w.line(2, inst.Name+":")
	w.line(4, "image: "+d.Image)
	w.line(4, "container_name: "+inst.Name)
	if d.Restart != "" {
		w.line(4, "restart: "+d.Restart)
	}
	if len(d.Ports) > 0 {
		w.line(4, "ports:")
		for _, p := range d.Ports {
			w.line(6, `- "`+p.String()+`"`)
		}
	}
	// environment: TZ only when set, JAVA_TOOL_OPTIONS only when this shard uses
	// MQ TLS. Omit the whole key when neither applies.
	if d.Timezone != "" || inst.MQTLS {
		w.line(4, "environment:")
		if d.Timezone != "" {
			w.line(6, "TZ: "+d.Timezone)
		}
		if inst.MQTLS {
			w.line(6, `JAVA_TOOL_OPTIONS: "-Dcom.ibm.mq.cfg.useIBMCipherMappings=false"`)
		}
	}
	if in.EnvFile != "" {
		w.line(4, "env_file:")
		w.line(6, "- "+in.EnvFile)
	}
	// configs: always one entry -- this shard's inlined application.yml.
	w.line(4, "configs:")
	w.line(6, "- source: "+inst.Name+"-app")
	w.line(8, "target: /app/external/spring/config/application.yml")
	// volumes: stores first, then libs; omit the key when there are neither.
	if len(in.Stores) > 0 || in.Libs != nil {
		w.line(4, "volumes:")
		for _, s := range in.Stores {
			w.line(6, "- "+s.Source+":"+s.Target+":ro")
		}
		if in.Libs != nil {
			w.line(6, "- "+in.Libs.Source+":"+in.Libs.Target+":ro")
		}
	}
}

// renderConfig emits one top-level configs entry inlining the shard's
// application.yml as a block scalar under content:. Blank lines are preserved as
// truly empty lines (no indent, no trailing spaces).
func renderConfig(w *yw, inst Instance) {
	w.line(2, inst.Name+"-app:")
	w.line(4, "content: |")
	for _, ln := range splitLines(inst.AppYAML) {
		if ln == "" {
			w.raw("\n")
		} else {
			w.line(6, ln)
		}
	}
}

// splitLines splits s on '\n' and drops a single trailing empty element that a
// terminating newline produces, so the block scalar keeps exactly one final
// newline (mirrors deploy.splitLines).
func splitLines(s string) []string {
	lines := strings.Split(s, "\n")
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return lines
}
