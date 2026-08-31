package spec

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// Leader-election modes.
const (
	LeaderStandalone   = "standalone"
	LeaderActiveActive = "active_active"
	LeaderActiveStby   = "active_standby"
)

// Store is one JKS/PKCS12 truststore or keystore definition. The tags are
// explicit because `password-env` cannot be reached by yaml.v3's default
// lowercase-the-field-name mapping; the other three keep the names they always
// had.
type Store struct {
	File        string `yaml:"file"`
	Password    string `yaml:"password" expand:"no"`     // credential: out of scope (rule 5)
	PasswordEnv string `yaml:"password-env" expand:"no"` // names a host var; expanding it would defeat the -env indirection
	Type        string `yaml:"type"`
}

// Secret is the store's password credential.
func (s *Store) Secret() Cred {
	if s == nil {
		return Cred{}
	}
	return Cred{s.Password, s.PasswordEnv}
}

// TLSConfig is the single shared truststore (+ optional keystore) used by both
// Solace and MQ TLS connections.
type TLSConfig struct {
	Truststore *Store
	Keystore   *Store
}

// User is a management-endpoint basic-auth user. Name is an identity, not a
// credential, so it stays a plain literal; only the password is a Cred pair.
//
// Roles are the connector's own authority names, passed through verbatim. An
// empty list is the connector's read-only default (GET only); "admin" is what
// grants GET+POST, the only way to reach /actuator/workflows. Deliberately not
// expand:"no" -- a role is an identity like Name, so ${VAR} resolves in it.
// The tool never adds a role itself: the reserved status account it injects
// carries none, which is exactly what keeps that account read-only.
type User struct {
	Name        string   `yaml:"name"`
	Password    string   `yaml:"password" expand:"no"`     // credential: out of scope (rule 5)
	PasswordEnv string   `yaml:"password-env" expand:"no"` // names a host var; expanding it would defeat the -env indirection
	Roles       []string `yaml:"roles"`
}

// Secret is the user's password credential.
func (u User) Secret() Cred { return Cred{u.Password, u.PasswordEnv} }

// StatusUserName is the tool-reserved, read-only actuator account used by the
// generated status script. StatusUserPasswordEnvVar is the optional host env
// var that overrides its generated password at render time.
const (
	StatusUserName           = "solmq-status"
	StatusUserPasswordEnvVar = "SECURITY_USER_SOLMQ_STATUS_PASSWORD"
)

// ConnectorContainerName is the name of the connector container inside a
// generated kubernetes pod. It is a constant rather than a literal in the
// renderer because the status verb has to find that container again in a pod's
// containerStatuses to report its state, and a rename on one side without the
// other would silently report a pod with no container facts at all.
//
// docker and podman need no equivalent: there the container *is* the instance
// and its name is the operator's own (docker.name / podman.name in env.yaml).
const ConnectorContainerName = "connector"

// Label keys/values emitted on Kubernetes pods, compose services and podman
// containers alike: solace-connector is a valid DNS-1123 subdomain prefix and
// le-mode/role are valid label names, so one spelling works on every platform.
const (
	LabelModeKey    = "solace-connector/le-mode"
	LabelRoleKey    = "solace-connector/role"
	LabelRoleActive = "active"
)

// Security mirrors solace.connector.security. Management security is always
// on now (see StatusUserName), so there is no enabled/disabled toggle to
// model here; Enabled only carries a parsed security.enabled value through
// to validate so a stale key is rejected instead of silently ignored.
type Security struct {
	Users   []User
	Enabled *bool // removed; non-nil is a validation error
}

// Management mirrors the Spring actuator settings. The block is always
// emitted now, so there is nothing conditional left in this struct; Exposure
// only carries a parsed management.exposure value through to validate so a
// stale key is rejected instead of silently ignored.
type Management struct {
	Port              int
	Exposure          *string // removed; non-nil is a validation error
	HealthShowDetails string
}

// LeaderElection captures solace.connector.management.leader-election plus the
// management queue and Solace session used by active_active / active_standby.
type LeaderElection struct {
	Present  bool
	Mode     string
	Queue    string
	ConnRef  string     // solace connection for the session (OR inline Session)
	Session  *Side      // inline solace session (alternative to ConnRef)
	FailOver *yaml.Node // verbatim leader-election.fail-over mapping
}

// EffectiveMode returns le.Mode, defaulting to LeaderStandalone: an absent or
// empty leader-election mode means standalone.
func (le LeaderElection) EffectiveMode() string {
	if le.Mode == "" {
		return LeaderStandalone
	}
	return le.Mode
}

// Defaults is the parsed connector-defaults section of env.yaml (all subsections optional).
type Defaults struct {
	TLS            TLSConfig
	LoggingLevel   *yaml.Node // ordered mapping under logging.level
	Management     Management
	Security       Security
	LeaderElection LeaderElection
	SolaceDefaults *yaml.Node      // ordered mapping merged into each Solace binder's solace.java.*
	Connections    map[string]Side // reusable connections referenced by conn-ref
}

// Resolve materialises a side that uses conn-ref: it returns the referenced
// connection's tuple carrying the side's own destination (and per-binding
// passthrough). Sides without conn-ref, or with an unknown ref (validate flags
// it), are returned unchanged.
func (d *Defaults) Resolve(s Side) Side {
	if s.ConnRef == "" {
		return s
	}
	conn, ok := d.Connections[s.ConnRef]
	if !ok {
		return s
	}
	r := conn // connection tuple (System, host/creds/tls/api-props, key-alias)
	r.ConnRef = s.ConnRef
	r.DestKind = s.DestKind
	r.Dest = s.Dest
	r.Consumer = s.Consumer
	r.Producer = s.Producer
	return r
}

type rawDefaults struct {
	TLS            *rawTLS            `yaml:"tls"`
	Logging        *rawLogging        `yaml:"logging"`
	Management     *rawManagement     `yaml:"management"`
	Security       *rawSecurity       `yaml:"security"`
	LeaderElection *rawLeader         `yaml:"leader-election"`
	Connections    map[string]rawSide `yaml:"connections"`
	// yaml.Node value (not *yaml.Node) so the subtree is actually captured.
	SolaceDefaults yaml.Node `yaml:"solace-defaults"`
}

type rawTLS struct {
	Truststore *Store `yaml:"truststore"`
	Keystore   *Store `yaml:"keystore"`
}

type rawLogging struct {
	Level yaml.Node `yaml:"level"`
}

type rawManagement struct {
	Port              int     `yaml:"port"`
	Exposure          *string `yaml:"exposure"`
	HealthShowDetails string  `yaml:"health-show-details"`
}

type rawSecurity struct {
	Enabled *bool  `yaml:"enabled"`
	Users   []User `yaml:"users"`
}

type rawLeader struct {
	Mode     string     `yaml:"mode"`
	Queue    string     `yaml:"queue"`
	ConnRef  string     `yaml:"conn-ref"`
	Solace   *rawSolace `yaml:"solace"`    // inline session (alternative to conn-ref)
	FailOver yaml.Node  `yaml:"fail-over"` // verbatim
}

// ParseDefaults decodes a standalone defaults document. An empty/absent file
// yields a zero Defaults. The unified env.yaml reuses defaultsFromRaw directly
// (see ParseEnv) so both paths apply identical mapping and defaulting.
func ParseDefaults(data []byte) (*Defaults, error) {
	var raw rawDefaults
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("env.yaml: %v", err)
	}
	return defaultsFromRaw(raw), nil
}

// defaultsFromRaw maps a decoded rawDefaults into the public Defaults.
func defaultsFromRaw(raw rawDefaults) *Defaults {
	d := &Defaults{}
	if raw.TLS != nil {
		d.TLS.Truststore = raw.TLS.Truststore
		d.TLS.Keystore = raw.TLS.Keystore
	}
	if raw.Logging != nil {
		d.LoggingLevel = nodePtr(raw.Logging.Level)
	}
	if raw.Management != nil {
		d.Management = Management{
			Port:              raw.Management.Port,
			Exposure:          raw.Management.Exposure,
			HealthShowDetails: raw.Management.HealthShowDetails,
		}
	}
	if raw.Security != nil {
		d.Security = Security{Enabled: raw.Security.Enabled, Users: raw.Security.Users}
	}
	if raw.LeaderElection != nil {
		le := LeaderElection{
			Present:  true,
			Mode:     raw.LeaderElection.Mode,
			Queue:    raw.LeaderElection.Queue,
			ConnRef:  raw.LeaderElection.ConnRef,
			FailOver: nodePtr(raw.LeaderElection.FailOver),
		}
		if raw.LeaderElection.Solace != nil {
			rs := rawSide{Solace: raw.LeaderElection.Solace}
			sess := rs.toSide()
			le.Session = &sess
		}
		d.LeaderElection = le
	}
	if len(raw.Connections) > 0 {
		d.Connections = make(map[string]Side, len(raw.Connections))
		for name, rc := range raw.Connections {
			rc := rc
			d.Connections[name] = rc.toSide()
		}
	}
	d.SolaceDefaults = nodePtr(raw.SolaceDefaults)
	return d
}
