package spec

import (
	"fmt"
	"sort"

	"gopkg.in/yaml.v3"
)

// Leader-election modes.
const (
	LeaderStandalone   = "standalone"
	LeaderActiveActive = "active_active"
	LeaderActiveStby   = "active_standby"
)

// StoreFormat is the keystore type of every store: create builds a PKCS12, and
// existing names one, so the tool always writes this type -- JCSMP's own
// documentation disagrees with itself about the default, so it is never left
// to one.
const StoreFormat = "PKCS12"

// Store is one named truststore or keystore. It has at most one host-side
// source -- built by the tool from certificate files or a JKS (Create), or a
// PKCS12 file the operator provides (Existing.PKCS12) -- and optionally a
// cluster location, a kubernetes Secret the operator manages
// (Existing.Secret). Kubernetes reads a store from its Secret whenever it names
// one and never opens the file or builds the store; docker and podman ignore
// the Secret and need the host-side source. validate enforces the shape. The
// password encrypts a created store and opens an existing one, and the
// connection is wired with it either way. The tags are explicit because
// `password-env` cannot be reached by yaml.v3's default
// lowercase-the-field-name mapping.
type Store struct {
	Password    string         `yaml:"password" expand:"no"`     // credential: out of scope (rule 5)
	PasswordEnv string         `yaml:"password-env" expand:"no"` // names a host var; expanding it would defeat the -env indirection
	Create      *StoreSources  `yaml:"create"`
	Existing    *StoreExisting `yaml:"existing"`
}

// StoreSources is what a created store is built from. Trusted certificates --
// Certs, plus the trusted entries of JKS -- make a truststore; one key entry --
// Keys, or a JKS holding exactly one key -- makes a keystore. A store is never
// both: the PKCS12 writer puts either trusted certificates or one unnamed key
// in a file.
type StoreSources struct {
	Certs CertSources `yaml:"certs"`
	Keys  []KeySource `yaml:"keys"`
	JKS   string      `yaml:"jks"`
}

// CertSources is create.certs: a list of certificate files, aliased
// <store>-<n> in order, or a mapping of alias to file, aliased by name. Items
// keeps the order either was written in.
type CertSources struct {
	Items []CertSource
	Named bool
}

// CertSource is one create.certs entry. Name is the alias a mapping gives it,
// empty in a list.
type CertSource struct {
	Name string `expand:"no"` // an alias, not a path
	File string
}

// UnmarshalYAML reads a list of files or a mapping of alias: file.
// rawTLS.UnmarshalYAML has already checked the shape, naming the store.
func (c *CertSources) UnmarshalYAML(node *yaml.Node) error {
	node = deref(node)
	switch node.Kind {
	case yaml.SequenceNode:
		for _, item := range node.Content {
			var f string
			if err := item.Decode(&f); err != nil {
				return err
			}
			c.Items = append(c.Items, CertSource{File: f})
		}
	case yaml.MappingNode:
		c.Named = true
		for i := 0; i+1 < len(node.Content); i += 2 {
			var f string
			if err := node.Content[i+1].Decode(&f); err != nil {
				return err
			}
			c.Items = append(c.Items, CertSource{Name: node.Content[i].Value, File: f})
		}
	}
	return nil
}

// KeySource is one create.keys entry: the certificate the key belongs to (its
// file may also hold the chain), further chain files, and the private key.
type KeySource struct {
	Cert  string   `yaml:"cert"`
	Chain []string `yaml:"chain"`
	Key   string   `yaml:"key"`
}

// StoreExisting is a store the operator provides: a PKCS12 file on the deploy
// host, a kubernetes Secret holding it, or both -- the Secret for kubernetes,
// the file for docker and podman, so one env.yaml serves every platform.
type StoreExisting struct {
	PKCS12 string       `yaml:"pkcs12"`
	Secret *StoreSecret `yaml:"secret"`
}

// StoreSecret locates a store in a kubernetes Secret the operator manages:
// the Secret's name and the data key holding the PKCS12. Key defaults to the
// store's own name (EffectiveKey), which is the key the tool's own stores
// Secret uses. A cert-manager Certificate with keystores.pkcs12 writes its
// stores under keystore.p12 and truststore.p12, so two stores can name one
// Secret with a key each.
type StoreSecret struct {
	Name string `yaml:"name"`
	Key  string `yaml:"key"`
}

// EffectiveKey is the Secret data key the store is read from: Key, or the
// store's name when Key is unset.
func (s *StoreSecret) EffectiveKey(store string) string {
	if s == nil || s.Key == "" {
		return store
	}
	return s.Key
}

// Secret is the store's password credential.
func (s *Store) Secret() Cred {
	if s == nil {
		return Cred{}
	}
	return Cred{s.Password, s.PasswordEnv}
}

// Created reports whether the tool builds the store (create:). Nil-safe.
func (s *Store) Created() bool { return s != nil && s.Create != nil }

// Path is the PKCS12 file an existing store names, or "" -- always for a
// created store, which has no file: deploy builds it straight into the
// platform's secret.
func (s *Store) Path() string {
	if s == nil || s.Existing == nil {
		return ""
	}
	return s.Existing.PKCS12
}

// SecretRef is the kubernetes Secret an existing store names, or nil.
func (s *Store) SecretRef() *StoreSecret {
	if s == nil || s.Existing == nil {
		return nil
	}
	return s.Existing.Secret
}

// Configured reports whether the store can be wired: created, naming its
// PKCS12, or naming the kubernetes Secret that holds it. It is the one
// definition every reader shares -- the Solace api-properties, the MQ SSL
// bundle, validate and every platform's mount -- so a store validate refuses
// can never be wired by one and skipped by another. Every source is mounted at
// the same path, so the rendered config does not depend on which one it is.
// Nil-safe: a store name nothing defines is not configured.
func (s *Store) Configured() bool { return s.Created() || s.Path() != "" || s.SecretRef() != nil }

// TLSConfig is the tls: block of env.yaml: any number of named stores. A store
// has no role of its own -- it is a truststore, a keystore or both according to
// how connections select it (truststore: <name> / keystore: <name>), so Solace
// and MQ connections, and the leader-election session, each trust and present
// what they choose.
type TLSConfig struct {
	Stores map[string]*Store
}

// Store returns the store named name, or nil when name is empty or undefined.
func (c TLSConfig) Store(name string) *Store {
	if name == "" {
		return nil
	}
	return c.Stores[name]
}

// StoreNames lists the defined store names in sorted order, so every reader
// walks them the same way.
func (c TLSConfig) StoreNames() []string {
	out := make([]string, 0, len(c.Stores))
	for name := range c.Stores {
		out = append(out, name)
	}
	sort.Strings(out)
	return out
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

// GeneratedNamePrefix marks every mount name the tool derives for itself, as
// opposed to the ones operators supply through the `-env` fields. A credential
// written as `-env` is mounted under the host variable it names, so the two
// kinds of name share one namespace; this prefix keeps the derived half out of
// the way of anything an operator would plausibly export, since a collision
// would otherwise mean two credentials competing for one mounted file.
//
// The leading underscore is deliberate: it is legal in an environment-variable
// identifier, a Kubernetes Secret data key, a file name, a Spring property
// segment and a compose secret key alike, and is vanishingly rare in
// hand-written variable names.
const GeneratedNamePrefix = "_GEN_"

// ConnectorContainerName is the name of the connector container inside a
// generated kubernetes pod. It is a constant rather than a literal in the
// renderer because the status verb has to find that container again in a pod's
// containerStatuses to report its state, and a rename on one side without the
// other would silently report a pod with no container facts at all.
//
// Every kubernetes exec and log read names it outright (kubectl's -c), rather
// than letting the engine pick a default container, so a pod carrying a sidecar
// cannot be read as if the sidecar were the connector.
//
// docker and podman have no second container to disambiguate -- there the
// container *is* the instance, and its name is the operator's own
// (docker.name / podman.name in env.yaml) -- but DefaultConnectorName is this
// same value, so an instance that does not override it is called the same thing
// on all three platforms.
const ConnectorContainerName = "connector"

// Label keys/values emitted on Kubernetes pods, compose services and podman
// containers alike: solace-connector is a valid DNS-1123 subdomain prefix and
// le-mode/role are valid label names, so one spelling works on every platform.
const (
	LabelModeKey    = "solace-connector/le-mode"
	LabelRoleKey    = "solace-connector/role"
	LabelRoleActive = "active"
)

// LabelStoresDigestKey labels a docker service that carries created stores
// with the sha256 of their bytes, so compose recreates the container exactly
// when a store changes: the stores themselves travel in the environment, which
// compose's own change detection never sees.
const LabelStoresDigestKey = "solace-connector/stores-sha256"

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
	Present   bool
	Mode      string
	Queue     string
	ConnRef   string     // solace connection for the session (OR inline Session)
	Session   *Side      // inline management session (alternative to ConnRef)
	SolaceKey bool       // solace: was renamed to session:; true is a validation error
	FailOver  *yaml.Node // verbatim leader-election.fail-over mapping
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
	Syslog         *Syslog    // logging.syslog, nil when the block is absent
	Management     Management
	Security       Security
	LeaderElection LeaderElection
	SolaceDefaults *yaml.Node      // ordered mapping merged into each Solace binder's solace.java.*
	MQDefaults     *yaml.Node      // ordered mapping merged into each MQ binder's ibm.mq.*
	Connections    map[string]Side // reusable connections referenced by conn-ref

	// MisplacedTransforms lists every transform-looking key in env.yaml, as
	// dotted paths. None belongs there -- a transform is per workflow, written
	// in the workflow file -- so validate refuses each one.
	MisplacedTransforms []string
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
	MQDefaults     yaml.Node `yaml:"mq-defaults"`
}

type rawTLS struct {
	Stores map[string]*Store
}

// The keys of one tls.stores.<name> block, its create: and existing: blocks,
// and one create.keys entry.
var (
	storeKeys    = map[string]bool{"password": true, "password-env": true, "create": true, "existing": true}
	createKeys   = map[string]bool{"certs": true, "keys": true, "jks": true}
	keyEntryKeys = map[string]bool{"cert": true, "chain": true, "key": true}
	existingKeys = map[string]bool{"pkcs12": true, "secret": true}
	secretKeys   = map[string]bool{"name": true, "key": true}
)

// StoreKeys lists the keys of a tls.stores.<name> block, sorted -- like the
// three below, for the generator page's copy to be checked against.
func StoreKeys() []string { return sortedKeys(storeKeys) }

// StoreCreateKeys lists the keys of a create: block, sorted.
func StoreCreateKeys() []string { return sortedKeys(createKeys) }

// StoreKeyEntryKeys lists the keys of a create.keys entry, sorted.
func StoreKeyEntryKeys() []string { return sortedKeys(keyEntryKeys) }

// StoreExistingKeys lists the keys of an existing: block, sorted.
func StoreExistingKeys() []string { return sortedKeys(existingKeys) }

// StoreSecretKeys lists the keys of an existing.secret block, sorted.
func StoreSecretKeys() []string { return sortedKeys(secretKeys) }

func sortedKeys(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// UnmarshalYAML accepts only tls.stores, a mapping of store names to store
// blocks, and only the known keys at every level below it. ParseEnv decodes
// without KnownFields, so without this the retired tls.truststore or a store's
// retired file:, or a misspelled key (passwrd-env, cert:), would be dropped in
// silence and the connection would trust the JVM default store, or present no
// certificate, while env.yaml says otherwise.
func (r *rawTLS) UnmarshalYAML(node *yaml.Node) error {
	if node.Kind != yaml.MappingNode {
		return fmt.Errorf("tls must be a mapping holding stores:, got a %s", YAMLKind(node))
	}
	for i := 0; i+1 < len(node.Content); i += 2 {
		k, v := node.Content[i], deref(node.Content[i+1])
		switch k.Value {
		case "stores":
		case "truststore", "keystore":
			return fmt.Errorf("tls.%s is no longer read: define each store under tls.stores.<name> and select it on the connection with truststore: <name> or keystore: <name>", k.Value)
		default:
			return fmt.Errorf("tls.%s is not a key: tls takes only stores:, a mapping of store names to {password | password-env, create and/or existing}", k.Value)
		}
		if isNull(v) {
			continue
		}
		if v.Kind != yaml.MappingNode {
			return fmt.Errorf("tls.stores must be a mapping of store names to store blocks, got a %s", YAMLKind(v))
		}
		for j := 0; j+1 < len(v.Content); j += 2 {
			if block := deref(v.Content[j+1]); block.Kind == yaml.MappingNode {
				if err := checkStoreBlock(v.Content[j].Value, block); err != nil {
					return err
				}
			}
		}
		var stores map[string]*Store
		if err := v.Decode(&stores); err != nil {
			return err
		}
		r.Stores = stores
	}
	return nil
}

// checkStoreBlock holds one tls.stores.<name> block to its keys, and its
// create: and existing: blocks to theirs.
func checkStoreBlock(name string, block *yaml.Node) error {
	for n := 0; n+1 < len(block.Content); n += 2 {
		key, val := block.Content[n].Value, deref(block.Content[n+1])
		switch {
		case key == "file" || key == "type":
			return fmt.Errorf("tls.stores.%s.%s is no longer read: name a PKCS12 you provide with existing: {pkcs12: <file>}, or one a kubernetes Secret holds with existing: {secret: {name: <Secret>}}, or have the tool build one with create:", name, key)
		case !storeKeys[key]:
			return fmt.Errorf("tls.stores.%s.%s is not a store key: a store takes password or password-env, and create: and/or existing:", name, key)
		case isNull(val):
		case key == "create":
			if err := checkCreateBlock(name, val); err != nil {
				return err
			}
		case key == "existing":
			if err := checkExistingBlock(name, val); err != nil {
				return err
			}
		}
	}
	return nil
}

// checkExistingBlock holds an existing: block to pkcs12 and secret, and its
// secret: block to name and key.
func checkExistingBlock(name string, block *yaml.Node) error {
	if block.Kind != yaml.MappingNode {
		return fmt.Errorf("tls.stores.%s.existing must be a mapping holding pkcs12: <file> and/or secret: {name, key}, got a %s", name, YAMLKind(block))
	}
	for e := 0; e+1 < len(block.Content); e += 2 {
		k, v := block.Content[e].Value, deref(block.Content[e+1])
		if !existingKeys[k] {
			return fmt.Errorf("tls.stores.%s.existing.%s is not an existing key: existing takes pkcs12 (a PKCS12 file you provide) and secret (a kubernetes Secret you manage that holds it)", name, k)
		}
		if k != "secret" || isNull(v) {
			continue
		}
		if v.Kind != yaml.MappingNode {
			return fmt.Errorf("tls.stores.%s.existing.secret must be a mapping holding name: <Secret> and optionally key: <data key>, got a %s", name, YAMLKind(v))
		}
		for s := 0; s+1 < len(v.Content); s += 2 {
			if sk := v.Content[s].Value; !secretKeys[sk] {
				return fmt.Errorf("tls.stores.%s.existing.secret.%s is not a secret key: secret takes name (the kubernetes Secret) and key (its data key, by default the store name)", name, sk)
			}
		}
	}
	return nil
}

// checkCreateBlock holds a create: block to its keys and their shapes.
func checkCreateBlock(name string, block *yaml.Node) error {
	if block.Kind != yaml.MappingNode {
		return fmt.Errorf("tls.stores.%s.create must be a mapping holding certs, keys or jks, got a %s", name, YAMLKind(block))
	}
	for c := 0; c+1 < len(block.Content); c += 2 {
		key, val := block.Content[c].Value, deref(block.Content[c+1])
		if !createKeys[key] {
			return fmt.Errorf("tls.stores.%s.create.%s is not a create key: create takes certs, keys and jks", name, key)
		}
		if isNull(val) {
			continue
		}
		switch key {
		case "certs":
			if val.Kind != yaml.SequenceNode && val.Kind != yaml.MappingNode {
				return fmt.Errorf("tls.stores.%s.create.certs must be a list of files or a mapping of alias: file, got a %s", name, YAMLKind(val))
			}
		case "keys":
			if val.Kind != yaml.SequenceNode {
				return fmt.Errorf("tls.stores.%s.create.keys must be a list of {cert, chain, key} entries, got a %s", name, YAMLKind(val))
			}
			for i, entry := range val.Content {
				if err := checkKeyEntry(name, i, deref(entry)); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// checkKeyEntry holds one create.keys entry to cert, chain and key.
func checkKeyEntry(name string, i int, entry *yaml.Node) error {
	if entry.Kind != yaml.MappingNode {
		return fmt.Errorf("tls.stores.%s.create.keys[%d] must be a mapping of cert, chain and key, got a %s", name, i, YAMLKind(entry))
	}
	for e := 0; e+1 < len(entry.Content); e += 2 {
		key, val := entry.Content[e].Value, deref(entry.Content[e+1])
		if !keyEntryKeys[key] {
			return fmt.Errorf("tls.stores.%s.create.keys[%d].%s is not a key-entry key: an entry takes cert, chain and key", name, i, key)
		}
		if key == "chain" && !isNull(val) && val.Kind != yaml.SequenceNode {
			return fmt.Errorf("tls.stores.%s.create.keys[%d].chain must be a list of files, got a %s", name, i, YAMLKind(val))
		}
	}
	return nil
}

// isNull reports whether n is an explicit or implied null (key: with nothing).
func isNull(n *yaml.Node) bool {
	return n.Kind == yaml.ScalarNode && n.ShortTag() == "!!null"
}

type rawLogging struct {
	Level  yaml.Node `yaml:"level"`
	Syslog *Syslog   `yaml:"syslog"`
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
	Mode    string     `yaml:"mode"`
	Queue   string     `yaml:"queue"`
	ConnRef string     `yaml:"conn-ref"`
	Session *rawSolace `yaml:"session"` // inline management session (alternative to conn-ref)
	// solace: is the retired spelling of session:. Captured as a bare node rather
	// than a *rawSolace so ANY shape written under it parses -- a scalar, a list, a
	// half-migrated block -- and reaches validate as the rename error, instead of
	// failing here as a yaml type error that names neither key.
	Solace   yaml.Node `yaml:"solace"`
	FailOver yaml.Node `yaml:"fail-over"` // verbatim
}

// ParseDefaults decodes a standalone defaults document. An empty/absent file
// yields a zero Defaults. The unified env.yaml reuses defaultsFromRaw directly
// (see ParseEnv) so both paths apply identical mapping and defaulting.
func ParseDefaults(data []byte) (*Defaults, error) {
	var raw rawDefaults
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("env.yaml: %v", err)
	}
	d := defaultsFromRaw(raw)
	d.MisplacedTransforms = envTransforms(data)
	return d, nil
}

// envTransforms decodes env.yaml a second time, into the generic tree, to find
// the transform keys no struct has a field for. It cannot fail where the typed
// decode it follows succeeded.
func envTransforms(data []byte) []string {
	var doc yaml.Node
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return nil
	}
	return misplacedEnvTransforms(&doc)
}

// defaultsFromRaw maps a decoded rawDefaults into the public Defaults.
func defaultsFromRaw(raw rawDefaults) *Defaults {
	d := &Defaults{}
	if raw.TLS != nil {
		d.TLS.Stores = raw.TLS.Stores
	}
	if raw.Logging != nil {
		d.LoggingLevel = nodePtr(raw.Logging.Level)
		d.Syslog = raw.Logging.Syslog
		// udp is the default because it needs nothing on the classpath; tcp is
		// the opt-in that costs a jar.
		if d.Syslog != nil && d.Syslog.Protocol == "" {
			d.Syslog.Protocol = SyslogUDP
		}
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
			Present:   true,
			Mode:      raw.LeaderElection.Mode,
			Queue:     raw.LeaderElection.Queue,
			ConnRef:   raw.LeaderElection.ConnRef,
			SolaceKey: nodePtr(raw.LeaderElection.Solace) != nil,
			FailOver:  nodePtr(raw.LeaderElection.FailOver),
		}
		if raw.LeaderElection.Session != nil {
			rs := rawSide{Solace: raw.LeaderElection.Session}
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
	d.MQDefaults = nodePtr(raw.MQDefaults)
	return d
}
