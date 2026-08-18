package spec

import (
	"fmt"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// Deploy CLI defaults per target section.
const (
	DefaultDockerCommand = "docker"
	DefaultPodmanCommand = "podman"
)

// Shared docker/podman container defaults.
const (
	DefaultConnectorName = "solmq-connector"
	DefaultRestart       = "unless-stopped"
	DefaultMgmtPort      = 8090
)

// EffectiveManagementPort returns d.Management.Port, falling back to
// DefaultMgmtPort when unset. Nil-receiver safe so callers can invoke it on a
// *Defaults that was never parsed (no defaults.yaml/env.yaml section) without
// a guard at every call site.
func (d *Defaults) EffectiveManagementPort() int {
	if d == nil || d.Management.Port == 0 {
		return DefaultMgmtPort
	}
	return d.Management.Port
}

// Podman generate modes.
const (
	PodmanModeRun     = "run"     // emit a `podman run` script (default)
	PodmanModeQuadlet = "quadlet" // emit a .container quadlet unit
)

// Quadlet scopes. auto resolves at deploy time from the effective uid
// (root -> system, otherwise user).
const (
	QuadletScopeAuto   = "auto"
	QuadletScopeUser   = "user"
	QuadletScopeSystem = "system"
)

// Container mount points inside the connector image (mirror the k8s layout so
// application.yml keystore/truststore/libs paths are identical everywhere).
const (
	DefaultStoresMountPath = "/app/external/classpath/truststores"
	DefaultLibsMountPath   = "/app/external/libs"
)

// BaseName returns the final element of a path, splitting on both '/' and '\'
// so a config authored on Windows resolves to the same name when the CLI runs on
// Linux. It is the single definition: the store-mount path, the Kubernetes
// stores-Secret data key, and the libs download filename must all agree, and
// separate copies previously disagreed about backslashes.
func BaseName(p string) string {
	p = strings.ReplaceAll(p, "\\", "/")
	if i := strings.LastIndexByte(p, '/'); i >= 0 {
		return p[i+1:]
	}
	return p
}

// StoresMount bind-mounts the host tls.*.file directory into the container.
type StoresMount struct {
	MountPath string `yaml:"mount-path"` // default /app/external/classpath/truststores
}

// LibsMount bind-mounts a host directory of IBM MQ jars into the container.
type LibsMount struct {
	Dir       string `yaml:"dir"`        // host dir (relative to env.yaml or absolute)
	MountPath string `yaml:"mount-path"` // default /app/external/libs
}

// Port is one container port publication. A bare YAML scalar (e.g. 8090)
// publishes to the same host port (Host == Container); a "host:container" string
// (e.g. "8080:8090") maps distinct ports. Both renderers emit it as Host:Container.
type Port struct {
	Host      int
	Container int
}

// UnmarshalYAML accepts either a scalar int (host == container) or a
// "host:container" string. A non-scalar node, a non-integer field, or more than
// one colon errors at parse; the 1-65535 range check lives in validate so parse
// stays about shape.
func (p *Port) UnmarshalYAML(node *yaml.Node) error {
	if node.Kind != yaml.ScalarNode {
		return fmt.Errorf(`ports entry must be an integer or "host:container", got a %s`, node.Tag)
	}
	s := strings.TrimSpace(node.Value)
	host, container, hasColon := strings.Cut(s, ":")
	if !hasColon {
		n, err := strconv.Atoi(s)
		if err != nil {
			return fmt.Errorf(`ports entry %q must be an integer or "host:container"`, s)
		}
		p.Host, p.Container = n, n
		return nil
	}
	if strings.Contains(container, ":") {
		return fmt.Errorf(`ports entry %q must be "host:container" (exactly one colon)`, s)
	}
	h, herr := strconv.Atoi(strings.TrimSpace(host))
	c, cerr := strconv.Atoi(strings.TrimSpace(container))
	if herr != nil || cerr != nil {
		return fmt.Errorf(`ports entry %q must be "host:container" with integer ports`, s)
	}
	p.Host, p.Container = h, c
	return nil
}

// String renders the mapping as host:container, the form docker compose and
// podman both expect (a bare port yields "N:N").
func (p Port) String() string {
	return strconv.Itoa(p.Host) + ":" + strconv.Itoa(p.Container)
}

// Docker is the parsed docker section of env.yaml. Deploy uses `command compose`.
//
// There is no secrets: section: credentials are derived from the config's own
// credential fields and delivered as compose secrets, so there is nothing left
// to configure here. Secrets is kept unexported-in-spirit (parsed but rejected)
// so an old env.yaml fails loudly instead of silently losing its credentials.
type Docker struct {
	Command  string       `yaml:"command"` // default docker; e.g. "podman" or "docker --context foo"
	Image    string       `yaml:"image"`
	Name     string       `yaml:"name"`
	Restart  string       `yaml:"restart"`
	Ports    []Port       `yaml:"ports"`
	Timezone string       `yaml:"timezone"`
	Secrets  *Secrets     `yaml:"secrets"` // removed; non-nil is a validation error
	Stores   *StoresMount `yaml:"stores"`
	Libs     *LibsMount   `yaml:"libs"`
}

// Quadlet locates the systemd generator directory and its scope.
type Quadlet struct {
	Scope string `yaml:"scope"` // auto (default) | user | system
	Dir   string `yaml:"dir"`   // overrides the default dir for the resolved scope
}

// Podman is the parsed podman section of env.yaml. generate honours Mode
// (run|quadlet); deploy/delete are always quadlet + systemctl.
type Podman struct {
	Command  string       `yaml:"command"` // default podman
	Mode     string       `yaml:"mode"`    // run (default) | quadlet -- controls generate only
	Quadlet  *Quadlet     `yaml:"quadlet"`
	Image    string       `yaml:"image"`
	Name     string       `yaml:"name"`
	Ports    []Port       `yaml:"ports"`
	Restart  string       `yaml:"restart"`
	Timezone string       `yaml:"timezone"`
	Secrets  *Secrets     `yaml:"secrets"` // removed; non-nil is a validation error (see Docker)
	Stores   *StoresMount `yaml:"stores"`
	Libs     *LibsMount   `yaml:"libs"`
}

// applyDockerDefaults fills command/name/restart/ports and the mount paths.
func applyDockerDefaults(d *Docker) {
	if d.Command == "" {
		d.Command = DefaultDockerCommand
	}
	if d.Name == "" {
		d.Name = DefaultConnectorName
	}
	if d.Restart == "" {
		d.Restart = DefaultRestart
	}
	if len(d.Ports) == 0 {
		d.Ports = []Port{{Host: DefaultMgmtPort, Container: DefaultMgmtPort}}
	}
	applyMountDefaults(d.Stores, d.Libs)
}

// applyPodmanDefaults mirrors applyDockerDefaults and additionally defaults the
// generate mode and the quadlet scope (auto).
func applyPodmanDefaults(p *Podman) {
	if p.Command == "" {
		p.Command = DefaultPodmanCommand
	}
	if p.Mode == "" {
		p.Mode = PodmanModeRun
	}
	if p.Name == "" {
		p.Name = DefaultConnectorName
	}
	if p.Restart == "" {
		p.Restart = DefaultRestart
	}
	if len(p.Ports) == 0 {
		p.Ports = []Port{{Host: DefaultMgmtPort, Container: DefaultMgmtPort}}
	}
	if p.Quadlet == nil {
		p.Quadlet = &Quadlet{}
	}
	if p.Quadlet.Scope == "" {
		p.Quadlet.Scope = QuadletScopeAuto
	}
	applyMountDefaults(p.Stores, p.Libs)
}

// applyMountDefaults fills the container-side mount paths shared by docker/podman.
func applyMountDefaults(stores *StoresMount, libs *LibsMount) {
	if stores != nil && stores.MountPath == "" {
		stores.MountPath = DefaultStoresMountPath
	}
	if libs != nil && libs.MountPath == "" {
		libs.MountPath = DefaultLibsMountPath
	}
}
