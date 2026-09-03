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
//
// DefaultConnectorName is deliberately the same word as ConnectorContainerName,
// the name this tool gives the container inside a generated kubernetes pod, so
// that the connector is called one thing on every platform. It was
// "solmq-connector" before, which meant an operator reading a `docker exec`
// line and a `kubectl exec -c` line saw two different names for the same
// process.
const (
	DefaultConnectorName = ConnectorContainerName
	DefaultRestart       = "unless-stopped"
	DefaultMgmtPort      = 8090
)

// DefaultComposeProject is the compose project every generated compose file
// declares, so a stack's grouping comes from the spec rather than from the
// basename of whichever directory env.yaml happens to live in.
//
// Docker only, and deliberately not in the shared block above: podman has no
// project or pod concept for it to mean anything against.
const DefaultComposeProject = "solace-ibmmq-connectors"

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

	// SecretsMountPath is where the credentials Secret is mounted, one file per
	// name, on every platform. The connector's configtree import reads exactly
	// this directory, so one value keeps the generated application.yml
	// platform-independent -- which it has to be, since `generate config` takes
	// no platform at all.
	//
	// It is deliberately NOT /run/secrets, the conventional place and where this
	// tool used to put it. On RHEL and OpenShift, CRI-O bind-mounts host
	// subscription data over /run/secrets in every container by default, which
	// silently shadows a kubelet volume mounted at the same path: `oc describe
	// pod` reports the mount, /proc/mounts inside the container shows CRI-O's
	// tmpfs instead, the directory holds only rhsm, and the connector starts
	// with every credential placeholder unresolved because the import is
	// optional. Nothing owns /app/external, so nothing can take it away, and
	// keeping docker and podman on the same path costs one flag each and leaves
	// one path to know rather than two.
	SecretsMountPath = "/app/external/var/secrets"
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

// StoresMount was the docker/podman stores: block. The block is removed: the
// tls.*.file store files are bind-mounted whenever they are set, onto the fixed
// in-container path DefaultStoresMountPath, so there was nothing left for it to
// decide -- its one field could only ever hold that same fixed path.
//
// The type survives with no fields so an old env.yaml still decodes into a non-nil
// pointer and validate can reject it by name. ParseEnv decodes without
// KnownFields, so deleting the field outright would drop the section in silence --
// and the behaviour change runs both ways, since omitting stores: used to mean "do
// not bind-mount" and now means nothing at all.
type StoresMount struct{}

// LibsMount bind-mounts a host directory of IBM MQ jars into the container.
//
// Only the host side is configurable. The container side is fixed at
// DefaultLibsMountPath because the connector image launches with that directory
// literally on its classpath (-cp /app/external/libs), so mounting anywhere else
// puts the jars where the JVM never looks -- and kubernetes, which has no
// mount-path key at all, always used the fixed path anyway.
type LibsMount struct {
	Dir       string `yaml:"dir"`        // host dir (relative to env.yaml or absolute)
	MountPath string `yaml:"mount-path"` // removed; non-empty is a validation error
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
// Image, Timezone and Stores are kept on the same terms: the first two moved to
// their own top-level keys, and the store bind-mount is now derived from the
// tls.*.file paths themselves.
//
// ProjectName has no Podman counterpart: it names a compose project, and podman
// has no equivalent grouping to attach one to.
type Docker struct {
	Command string `yaml:"command"` // default docker; e.g. "podman" or "docker --context foo"
	Image   string `yaml:"image"`   // removed; non-empty is a validation error
	Name    string `yaml:"name"`
	// ProjectName becomes the compose file's top-level name:, which is what
	// labels every container in the stack as belonging to one project.
	ProjectName string       `yaml:"project-name"`
	Restart     string       `yaml:"restart"`
	Ports       []Port       `yaml:"ports"`
	Timezone    string       `yaml:"timezone"` // removed; non-empty is a validation error
	Secrets     *Secrets     `yaml:"secrets"`  // removed; non-nil is a validation error
	Stores      *StoresMount `yaml:"stores"`   // removed; non-nil is a validation error
	Libs        *LibsMount   `yaml:"libs"`
}

// Quadlet locates the systemd generator directory and its scope.
type Quadlet struct {
	Scope string `yaml:"scope"` // auto (default) | user | system
	Dir   string `yaml:"dir"`   // overrides the default dir for the resolved scope
}

// Podman is the parsed podman section of env.yaml. generate renders a .container
// quadlet unit; deploy/remove install and tear that same unit down via systemctl.
type Podman struct {
	Command  string       `yaml:"command"` // default podman
	Mode     string       `yaml:"mode"`    // removed; non-empty is a validation error
	Quadlet  *Quadlet     `yaml:"quadlet"`
	Image    string       `yaml:"image"` // removed; non-empty is a validation error (see Docker)
	Name     string       `yaml:"name"`
	Ports    []Port       `yaml:"ports"`
	Restart  string       `yaml:"restart"`
	Timezone string       `yaml:"timezone"` // removed; non-empty is a validation error (see Docker)
	Secrets  *Secrets     `yaml:"secrets"`  // removed; non-nil is a validation error (see Docker)
	Stores   *StoresMount `yaml:"stores"`   // removed; non-nil is a validation error (see Docker)
	Libs     *LibsMount   `yaml:"libs"`
}

// applyDockerDefaults fills command/name/project-name/restart.
//
// ports is deliberately not defaulted: publishing a container port to the host
// is an exposure decision, and nothing the tool does needs one -- status execs
// into the container rather than reaching it over a published port. An omitted
// ports: therefore publishes nothing, and the operator opts in by listing the
// ports they want.
func applyDockerDefaults(d *Docker) {
	if d.Command == "" {
		d.Command = DefaultDockerCommand
	}
	if d.Name == "" {
		d.Name = DefaultConnectorName
	}
	if d.ProjectName == "" {
		d.ProjectName = DefaultComposeProject
	}
	if d.Restart == "" {
		d.Restart = DefaultRestart
	}
}

// applyPodmanDefaults mirrors applyDockerDefaults (ports left unpublished
// included) and additionally defaults the quadlet scope (auto).
//
// Mode is deliberately not defaulted: the key is removed, and validate rejects a
// non-empty one. Defaulting it here would make every section fail that check for a
// value the operator never wrote. The same goes for libs.mount-path, which is why
// neither of these functions fills in a mount path any more: both container-side
// paths are fixed by the image, so there is nothing left to default.
func applyPodmanDefaults(p *Podman) {
	if p.Command == "" {
		p.Command = DefaultPodmanCommand
	}
	if p.Name == "" {
		p.Name = DefaultConnectorName
	}
	if p.Restart == "" {
		p.Restart = DefaultRestart
	}
	if p.Quadlet == nil {
		p.Quadlet = &Quadlet{}
	}
	if p.Quadlet.Scope == "" {
		p.Quadlet.Scope = QuadletScopeAuto
	}
}
