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
type User struct {
	Name        string `yaml:"name"`
	Password    string `yaml:"password" expand:"no"`     // credential: out of scope (rule 5)
	PasswordEnv string `yaml:"password-env" expand:"no"` // names a host var; expanding it would defeat the -env indirection
}

// Secret is the user's password credential.
func (u User) Secret() Cred { return Cred{u.Password, u.PasswordEnv} }

// Security mirrors solace.connector.security.
type Security struct {
	Present bool
	Enabled bool
	Users   []User
}

// Management mirrors the Spring actuator settings.
type Management struct {
	Present           bool
	Port              int
	Exposure          string
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
	Port              int    `yaml:"port"`
	Exposure          string `yaml:"exposure"`
	HealthShowDetails string `yaml:"health-show-details"`
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
			Present:           true,
			Port:              raw.Management.Port,
			Exposure:          raw.Management.Exposure,
			HealthShowDetails: raw.Management.HealthShowDetails,
		}
	}
	if raw.Security != nil {
		d.Security = Security{Present: true, Enabled: true, Users: raw.Security.Users}
		if raw.Security.Enabled != nil {
			d.Security.Enabled = *raw.Security.Enabled
		}
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
