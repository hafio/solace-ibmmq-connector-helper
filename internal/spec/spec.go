// Package spec defines the on-disk input schema for solmq-conn -- the per-workflow
// files and the unified env.yaml (connector defaults plus the workflows,
// kubernetes, docker, and podman sections) -- and parses them into typed values.
//
// Verbatim passthrough maps (Solace api-properties, MQ additional-properties, and
// per-binding consumer/producer tuning) are captured as *yaml.Node so their key
// order and scalar formatting survive into the rendered output unchanged.
package spec

import (
	"fmt"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// System identifiers for one side of a workflow.
const (
	SystemSolace = "solace"
	SystemMQ     = "mq"
)

// Destination kinds.
const (
	DestQueue = "queue"
	DestTopic = "topic"
)

// Side is one end (source or target) of a workflow: exactly one system
// (Solace or IBM MQ) pointing at exactly one destination (queue or topic).
type Side struct {
	System   string // SystemSolace | SystemMQ
	DestKind string // DestQueue | DestTopic
	Dest     string // destination name

	// Solace connection tuple. Each credential is a pair: the literal value, or
	// the name of a host environment variable holding it (never both -- validate
	// enforces the exclusivity). Neither ever reaches the generated config: both
	// resolve to a stable secret name, mounted as a file under /run/secrets.
	Host          string
	MsgVPN        string
	ClientUser    string `expand:"no"` // credential: out of scope (rule 5)
	ClientUserEnv string `expand:"no"` // names a host var; expanding it would defeat the -env indirection
	ClientPass    string `expand:"no"` // credential: out of scope (rule 5)
	ClientPassEnv string `expand:"no"` // names a host var; expanding it would defeat the -env indirection

	// MQ connection tuple.
	ConnName     string
	QueueManager string
	Channel      string
	User         string `expand:"no"` // credential: out of scope (rule 5)
	UserEnv      string `expand:"no"` // names a host var; expanding it would defeat the -env indirection
	Password     string `expand:"no"` // credential: out of scope (rule 5)
	PasswordEnv  string `expand:"no"` // names a host var; expanding it would defeat the -env indirection
	TLS          bool
	Cipher       string

	// mTLS: presence selects a client key from the shared keystore.
	KeyAlias string

	// ConnRef names a reusable connection defined in env.yaml (connections.<name>).
	// When set, this side carries only ConnRef + the destination; the connection supplies
	// the tuple / TLS / passthrough (materialised by Defaults.Resolve).
	ConnRef string

	// Verbatim passthrough (order-preserving mapping nodes; nil when absent).
	APIProps  *yaml.Node // Solace  -> solace.java.api-properties
	AddlProps *yaml.Node // MQ      -> ibm.mq.additional-properties
	Consumer  *yaml.Node // per-binding consumer tuning
	Producer  *yaml.Node // per-binding producer tuning
}

// Workflow is one parsed workflow file (one input-<N>/output-<N> pair).
type Workflow struct {
	File      string // basename of the source file (used in durable-sub key + ordering)
	Enabled   bool
	Source    Side
	Target    Side
	SourceSet bool // a source: block was present in the file
	TargetSet bool // a target: block was present in the file
}

// ---- raw YAML shapes ---------------------------------------------------------

type rawWorkflow struct {
	Enabled   *bool      `yaml:"enabled"`
	Source    *rawSide   `yaml:"source"`
	Target    *rawSide   `yaml:"target"`
	Transform *yaml.Node `yaml:"transform"` // absorbed and ignored (non-goal)
}

type rawSide struct {
	Solace    *rawSolace `yaml:"solace"`
	MQ        *rawMQ     `yaml:"mq"`
	Transform *yaml.Node `yaml:"transform"` // absorbed and ignored
}

// NOTE: node-capturing fields MUST be `yaml.Node` (value), not `*yaml.Node`.
// yaml.v3 v3.0.1 only captures a raw subtree when the destination type is
// exactly yaml.Node; a *yaml.Node field is decoded as a struct and stays empty.
type rawSolace struct {
	ConnRef       string    `yaml:"conn-ref"`
	Host          string    `yaml:"host"`
	MsgVPN        string    `yaml:"msg-vpn"`
	ClientUser    string    `yaml:"client-username"`
	ClientUserEnv string    `yaml:"client-username-env"`
	ClientPass    string    `yaml:"client-password"`
	ClientPassEnv string    `yaml:"client-password-env"`
	KeyAlias      string    `yaml:"key-alias"`
	Queue         *string   `yaml:"queue"`
	Topic         *string   `yaml:"topic"`
	APIProps      yaml.Node `yaml:"api-properties"`
	Consumer      yaml.Node `yaml:"consumer"`
	Producer      yaml.Node `yaml:"producer"`
}

type rawMQ struct {
	ConnRef      string    `yaml:"conn-ref"`
	ConnName     string    `yaml:"conn-name"`
	QueueManager string    `yaml:"queue-manager"`
	Channel      string    `yaml:"channel"`
	User         string    `yaml:"user"`
	UserEnv      string    `yaml:"user-env"`
	Password     string    `yaml:"password"`
	PasswordEnv  string    `yaml:"password-env"`
	TLS          bool      `yaml:"tls"`
	Cipher       string    `yaml:"cipher"`
	KeyAlias     string    `yaml:"key-alias"`
	Queue        *string   `yaml:"queue"`
	Topic        *string   `yaml:"topic"`
	AddlProps    yaml.Node `yaml:"additional-properties"`
	Consumer     yaml.Node `yaml:"consumer"`
	Producer     yaml.Node `yaml:"producer"`
}

// nodePtr returns a pointer to n, or nil when n captured nothing (absent key).
func nodePtr(n yaml.Node) *yaml.Node {
	if n.Kind == 0 {
		return nil
	}
	return &n
}

// ParseWorkflow decodes one workflow file. A YAML syntax error is returned as
// the error result. Structural conformance (exactly one system/destination per
// side, both sides present, required tuple fields) is left to the validate
// package so that `validate` can report every problem at once; ParseWorkflow
// fills in whatever it can and never panics on partial input.
func ParseWorkflow(data []byte, path string) (*Workflow, error) {
	var raw rawWorkflow
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("%s: %v", filepath.Base(path), err)
	}
	wf := &Workflow{File: filepath.Base(path), Enabled: true}
	if raw.Enabled != nil {
		wf.Enabled = *raw.Enabled
	}
	if raw.Source != nil {
		wf.Source = raw.Source.toSide()
		wf.SourceSet = true
	}
	if raw.Target != nil {
		wf.Target = raw.Target.toSide()
		wf.TargetSet = true
	}
	return wf, nil
}

func (r *rawSide) toSide() Side {
	if r.Solace != nil && r.MQ != nil {
		return Side{} // ambiguous: both systems set -> System "" flags a validation error
	}
	switch {
	case r.Solace != nil:
		s := Side{
			System:        SystemSolace,
			ConnRef:       r.Solace.ConnRef,
			Host:          r.Solace.Host,
			MsgVPN:        r.Solace.MsgVPN,
			ClientUser:    r.Solace.ClientUser,
			ClientUserEnv: r.Solace.ClientUserEnv,
			ClientPass:    r.Solace.ClientPass,
			ClientPassEnv: r.Solace.ClientPassEnv,
			KeyAlias:      r.Solace.KeyAlias,
			APIProps:      nodePtr(r.Solace.APIProps),
			Consumer:      nodePtr(r.Solace.Consumer),
			Producer:      nodePtr(r.Solace.Producer),
		}
		applyDest(&s, r.Solace.Queue, r.Solace.Topic)
		return s
	case r.MQ != nil:
		s := Side{
			System:       SystemMQ,
			ConnRef:      r.MQ.ConnRef,
			ConnName:     r.MQ.ConnName,
			QueueManager: r.MQ.QueueManager,
			Channel:      r.MQ.Channel,
			User:         r.MQ.User,
			UserEnv:      r.MQ.UserEnv,
			Password:     r.MQ.Password,
			PasswordEnv:  r.MQ.PasswordEnv,
			TLS:          r.MQ.TLS,
			Cipher:       r.MQ.Cipher,
			KeyAlias:     r.MQ.KeyAlias,
			AddlProps:    nodePtr(r.MQ.AddlProps),
			Consumer:     nodePtr(r.MQ.Consumer),
			Producer:     nodePtr(r.MQ.Producer),
		}
		applyDest(&s, r.MQ.Queue, r.MQ.Topic)
		return s
	default:
		return Side{}
	}
}

// applyDest records the destination kind/name. If both or neither are set the
// System stays as-is and DestKind is left blank/ambiguous for validate to flag.
func applyDest(s *Side, queue, topic *string) {
	switch {
	case queue != nil && topic == nil:
		s.DestKind, s.Dest = DestQueue, *queue
	case topic != nil && queue == nil:
		s.DestKind, s.Dest = DestTopic, *topic
	case queue != nil && topic != nil:
		s.DestKind = "" // ambiguous: both set
	default:
		s.DestKind = "" // missing: neither set
	}
}

// HasSystem reports whether the side named a system block at all.
func (s Side) HasSystem() bool { return s.System == SystemSolace || s.System == SystemMQ }

// SetsConnFields reports whether the side declared any *connection* field beyond
// its system, conn-ref, and destination — used to enforce that a conn-ref side
// sets only what it is allowed to.
//
// consumer:/producer: are deliberately excluded: they tune one binding, not the
// connection, so a side is free to reference a shared connection and still set
// them. Defaults.Resolve carries both onto the resolved side for exactly that
// reason.
func (s Side) SetsConnFields() bool {
	return s.Host != "" || s.MsgVPN != "" || s.ConnName != "" || s.QueueManager != "" ||
		s.Channel != "" || s.TLS || s.Cipher != "" || s.KeyAlias != "" ||
		s.APIProps != nil || s.AddlProps != nil ||
		!s.Username().Empty() || !s.Secret().Empty()
}

// Cred is one credential position: either a literal value or the name of a host
// environment variable holding it, never both. Both forms resolve to the same
// stable secret name, so the generated config never carries either one.
type Cred struct {
	Literal string
	EnvVar  string
}

// Empty reports whether the position was left unset entirely.
func (c Cred) Empty() bool { return c.Literal == "" && c.EnvVar == "" }

// Both reports whether the pair was over-specified (validate rejects it).
func (c Cred) Both() bool { return c.Literal != "" && c.EnvVar != "" }

// Key is a comparison token for a credential that never exposes the literal
// value: two positions with the same Key request the same credential. Used for
// binder identity and conflict detection.
func (c Cred) Key() string {
	if c.EnvVar != "" {
		return "E:" + c.EnvVar
	}
	if c.Literal == "" {
		return ""
	}
	return "L:" + c.Literal
}

// Describe names the credential's source for an error message without ever
// printing the literal value itself (S3: secrets never reach a log).
func (c Cred) Describe() string {
	switch {
	case c.EnvVar != "":
		return "the environment variable " + c.EnvVar
	case c.Literal != "":
		return "a literal value"
	default:
		return "nothing"
	}
}

// DedupKey is binder identity: two sides sharing it are the same connection and
// collapse into one rendered binder.
//
// It lives here, on Side, because both the renderer and the validator must agree
// on it exactly. They previously kept private copies, and the copies drifted the
// moment credentials became literal/-env pairs -- the validator went on keying off
// the literal username, so every side using client-username-env keyed as if it had
// no username at all, and the validator's conflict checks reasoned about a
// different grouping than the renderer produced.
//
// The password is deliberately excluded: two sides may legitimately repeat a
// connection without repeating its password, and validate.checkPasswordConflicts
// catches the case where they genuinely disagree.
func (s Side) DedupKey() string {
	if s.System == SystemSolace {
		return "S|" + s.Host + "|" + s.MsgVPN + "|" + s.Username().Key()
	}
	return "M|" + s.ConnName + "|" + s.QueueManager + "|" + s.Channel + "|" + s.Username().Key()
}

// Username is the side's username credential, whichever system it uses.
func (s Side) Username() Cred {
	if s.System == SystemMQ {
		return Cred{s.User, s.UserEnv}
	}
	return Cred{s.ClientUser, s.ClientUserEnv}
}

// Secret is the side's password credential, whichever system it uses.
func (s Side) Secret() Cred {
	if s.System == SystemMQ {
		return Cred{s.Password, s.PasswordEnv}
	}
	return Cred{s.ClientPass, s.ClientPassEnv}
}
