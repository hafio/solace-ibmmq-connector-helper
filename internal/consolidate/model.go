package consolidate

import (
	"gopkg.in/yaml.v3"

	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/spec"
)

// Prop is one ordered property line. Scalar values render as "key: val"; when
// Sub is non-nil the value is a nested YAML node rendered from that subtree
// (used for the rare non-scalar verbatim passthrough).
type Prop struct {
	Key string
	Val string
	Sub *yaml.Node
}

// SolaceBinder holds the Solace connection block for a binder.
type SolaceBinder struct {
	Host       string
	MsgVPN     string
	ClientUser string
	ClientPass string
	Extras     []Prop // from solace-defaults (connect-retries, reconnect-retries, ...)
	APIProps   []Prop // ordered api-properties (tool TLS keys first, then verbatim passthrough)
}

// MQBinder holds the IBM MQ connection block for a binder.
type MQBinder struct {
	QueueManager string
	Channel      string
	ConnName     string
	User         string
	Password     string
	SSLBundle    string // "" when the binder has no TLS bundle
	AddlProps    []Prop // ordered additional-properties (tool cipher first, then passthrough)
}

// Binder is one entry under spring.cloud.stream.binders.
type Binder struct {
	Name   string // resolved binder name (solace-<vpn> / mq-<qm> / undefined)
	Kind   string // spec.SystemSolace | spec.SystemMQ | "undefined"
	Solace *SolaceBinder
	MQ     *MQBinder
}

// Bundle is one entry under spring.ssl.bundle.jks (MQ TLS only).
type Bundle struct {
	Name          string // "<binder>-bundle"
	TruststoreLoc string
	TruststorePwd string
	TruststoreTyp string
	HasKeystore   bool
	KeystoreLoc   string
	KeystorePwd   string
	KeystoreTyp   string
	KeyAlias      string // "" when not mTLS
}

// Binding is one entry under spring.cloud.stream.bindings.
type Binding struct {
	Name   string // "input-<N>" | "output-<N>"
	Dest   string
	Binder string
}

// JMSBinding is one entry under spring.cloud.stream.jms.bindings.
type JMSBinding struct {
	Name     string // "input-<N>" | "output-<N>"
	Role     string // "consumer" | "producer"
	DestType string // "queue" | "topic"
	Durable  string // durable-subscription-name (consumer topic); "" if none
	Extra    []Prop // verbatim consumer/producer passthrough
}

// SolaceBinding is one entry under spring.cloud.stream.solace.bindings.
type SolaceBinding struct {
	Name     string // "input-<N>" | "output-<N>"
	Role     string // "consumer" | "producer"
	DestType string // "queue" (topic is the default and emits nothing); "" if only Extra
	Extra    []Prop // verbatim consumer/producer passthrough
}

// WorkflowEnable is one entry under solace.connector.workflows.<N>.
type WorkflowEnable struct {
	ID      int
	Enabled bool
}

// Session is the Solace management session for leader-election (active_standby /
// active_active).
type Session struct {
	Host       string
	MsgVPN     string
	ClientUser string
	ClientPass string
	APIProps   []Prop // tool-managed TLS api-properties (empty when not tcps://)
}

// LeaderElectionModel is the rendered solace.connector.management.leader-election
// block (nil for standalone/absent).
type LeaderElectionModel struct {
	Mode     string
	Queue    string
	FailOver *yaml.Node // verbatim fail-over mapping (nil when absent)
	Session  *Session   // nil when no session configured
}

// SecretRef is one credential the connector needs at runtime, named by the
// stable in-container name it is mounted under. Exactly one of Literal/EnvVar
// carries the source: a literal value from the spec, or the name of a host
// environment variable read at deploy time. Neither ever reaches a generated
// artifact -- the config references only Stable, and the value is resolved
// straight into the platform's secret store.
type SecretRef struct {
	Stable  string // e.g. PROD_SOLACE_CLIENT_PASSWORD
	Literal string
	EnvVar  string
}

// Model is the fully-ordered result of consolidation, consumed by render/deploy.
type Model struct {
	Bundles        []*Bundle
	Binders        []*Binder // appearance order; the undefined binder is last
	Bindings       []*Binding
	JMSBindings    []*JMSBinding
	SolaceBindings []*SolaceBinding
	Workflows      []WorkflowEnable

	Security       spec.Security
	Management     spec.Management
	LoggingLevel   *yaml.Node
	LeaderElection *LeaderElectionModel // nil for standalone/absent

	// Secrets is every credential this instance references, in first-use order.
	// The rendered config carries `${Stable}` placeholders; the deploy layer turns
	// this list into mounted files under /run/secrets.
	Secrets []SecretRef

	// ConfigImport, when set, emits spring.config.import so the connector reads
	// the mounted secret files as properties.
	ConfigImport string

	// MQTLS is true when any MQ binder uses TLS (drives JAVA_TOOL_OPTIONS in deploy).
	MQTLS bool
}
