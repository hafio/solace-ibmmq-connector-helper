// Package consolidate turns the parsed workflow files + defaults into a single
// fully-ordered Model: it deduplicates connections into shared binders, numbers
// workflows by (already-sorted) input order, derives destination-types, computes
// MQ durable-subscription names, and wires Solace + MQ TLS/mTLS. It assumes the
// input already passed validation (no structural or semantic errors).
package consolidate

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/spec"
	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/tls"
)

// acc accumulates the state of one deduplicated binder across all sides that map
// to it, before names and TLS wiring are finalized.
type acc struct {
	key  string
	kind string // spec.SystemSolace | spec.SystemMQ
	name string // resolved after all accs are known

	// Solace tuple. Credentials are Cred pairs (literal or host env var); the
	// rendered config carries neither, only a stable secret name.
	host, vpn  string
	user, pass spec.Cred
	// MQ tuple.
	conn, qm, ch   string
	mqUser, mqPass spec.Cred

	tls      bool   // TLS on this binder (Solace: tcps:// host; MQ: tls:true)
	keyAlias string // mTLS client key alias ("" = none)
	cipher   string // MQ JCE cipher ("" = none)

	connName string // contributing connection name (first by appearance); "" = purely inline

	pass2 []Prop // merged verbatim passthrough (api-properties / additional-properties)

	binder *Binder
}

// Opts tunes one Build.
type Opts struct {
	MountStores  bool
	ConfigImport string

	// StatusPassword is the already-resolved password for the reserved
	// spec.StatusUserName actuator account. The caller (internal/gen) resolves
	// it once per invocation and passes it through; Build never generates it
	// itself, so Build stays a pure, deterministic function of its inputs.
	StatusPassword string
}

// Build consolidates workflows (in the caller's sorted order) with defaults.
// Returns the ordered Model and any non-fatal warnings.
func Build(wfs []spec.Workflow, d *spec.Defaults, opts Opts) (*Model, []string) {
	if d == nil {
		d = &spec.Defaults{}
	}
	mountStores := opts.MountStores
	m := &Model{
		Security:     d.Security,
		Management:   Management{Port: d.EffectiveManagementPort(), HealthShowDetails: d.Management.HealthShowDetails},
		LoggingLevel: d.LoggingLevel,
		ConfigImport: opts.ConfigImport,
	}
	var warns []string
	warn := func(format string, a ...any) { warns = append(warns, fmt.Sprintf(format, a...)) }

	// secretRef records one credential under the name it is mounted as and returns
	// the placeholder the config should carry. An unset credential contributes
	// nothing and renders as "" so the caller's emptiness gate still omits the
	// line. Repeated positions (a store password referenced by both an SSL bundle
	// and the Solace api-properties) collapse onto one entry.
	//
	// A credential that names a host variable is mounted under that very name:
	// the operator writing `password-env: SOL_PASSWORD` gets the key SOL_PASSWORD
	// in the Secret, in /run/secrets and in the rendered ${...}, so a
	// hand-built `existing:` Secret needs no name derivation to reproduce. The
	// caller's derived name (stableName) is the fallback for a literal, which has
	// no name of its own.
	type claim struct {
		cred   spec.Cred
		origin string
	}
	seenSecret := map[string]claim{}
	conflicted := map[string]bool{}
	secretRef := func(stable, origin string, c spec.Cred) string {
		if c.Empty() {
			return ""
		}
		name := stable
		if c.EnvVar != "" {
			name = c.EnvVar
			origin += "-env"
		}
		// A name seen twice carrying the same credential is one credential reached
		// from two places (a store password used by both an SSL bundle and the
		// Solace api-properties), so it contributes one entry. Two *different*
		// credentials competing for one name can only be an -env variable colliding
		// with another position's derived name: one mounted file cannot hold both
		// values, so record both positions and let the caller refuse rather than
		// silently drop one.
		if prev, seen := seenSecret[name]; !seen {
			seenSecret[name] = claim{cred: c, origin: origin}
			m.Secrets = append(m.Secrets, SecretRef{Stable: name, Literal: c.Literal, EnvVar: c.EnvVar})
		} else if prev.cred != c && !conflicted[name] {
			conflicted[name] = true
			m.SecretConflicts = append(m.SecretConflicts, SecretConflict{Name: name, First: prev.origin, Second: origin})
		}
		return "${" + name + "}"
	}

	// Materialise conn-ref sides once, up front, so conn-ref and inline sides that
	// resolve to the same connection tuple consolidate into a single binder.
	type rwf struct {
		file     string
		enabled  bool
		src, tgt spec.Side
	}
	rwfs := make([]rwf, len(wfs))
	for i, wf := range wfs {
		rwfs[i] = rwf{file: wf.File, enabled: wf.Enabled, src: d.Resolve(wf.Source), tgt: d.Resolve(wf.Target)}
	}

	// ---- pass 1: register + accumulate binders --------------------------------
	byKey := map[string]*acc{}
	var accs []*acc
	register := func(s spec.Side) *acc {
		key := s.DedupKey()
		a := byKey[key]
		if a == nil {
			a = &acc{key: key, kind: s.System}
			switch s.System {
			case spec.SystemSolace:
				a.host, a.vpn = s.Host, s.MsgVPN
				a.user, a.pass = s.Username(), s.Secret()
				a.tls = isTCPS(s.Host)
			case spec.SystemMQ:
				a.conn, a.qm, a.ch = s.ConnName, s.QueueManager, s.Channel
				a.mqUser, a.mqPass = s.Username(), s.Secret()
				a.tls = s.TLS
			}
			byKey[key] = a
			accs = append(accs, a)
		}
		// The binder name derives from the first contributing connection name.
		if a.connName == "" && s.ConnRef != "" {
			a.connName = s.ConnRef
		}
		// Merge cross-side attributes (validate guarantees key-alias is consistent).
		if a.keyAlias == "" {
			a.keyAlias = s.KeyAlias
		}
		if s.System == spec.SystemMQ {
			a.tls = a.tls || s.TLS
			if s.Cipher != "" {
				if a.cipher != "" && a.cipher != s.Cipher {
					warn("binder %q: conflicting cipher values on the same binder; last (by filename) wins", displayName(a))
				}
				a.cipher = s.Cipher
			}
		}
		// Merge verbatim passthrough in workflow order.
		var node *yaml.Node
		if s.System == spec.SystemSolace {
			node = s.APIProps
		} else {
			node = s.AddlProps
		}
		for _, p := range nodeToProps(node) {
			a.pass2 = mergeProp(a.pass2, p, &warns, displayName(a))
		}
		return a
	}
	for _, w := range rwfs {
		register(w.src)
		register(w.tgt)
	}

	// ---- assign binder names ---------------------------------------------------
	assignBinderNames(accs)

	// ---- finalize binder + bundle objects -------------------------------------
	for _, a := range accs {
		switch a.kind {
		case spec.SystemSolace:
			sb := &SolaceBinder{
				Host:       a.host,
				MsgVPN:     a.vpn,
				ClientUser: secretRef(stableName(a.name, "CLIENT_USERNAME"), binderOwner(a.name)+" client-username", a.user),
				ClientPass: secretRef(stableName(a.name, "CLIENT_PASSWORD"), binderOwner(a.name)+" client-password", a.pass),
			}
			sb.Extras = nodeToProps(d.SolaceDefaults)
			var props []Prop
			if a.tls {
				for _, kv := range tls.SolaceProps(d, a.keyAlias, mountStores, storeSecret(secretRef)) {
					props = append(props, Prop{Key: kv.Key, Val: kv.Val})
				}
			}
			props = appendPassthrough(props, a.pass2, &warns, binderOwner(a.name))
			sb.APIProps = props
			a.binder = &Binder{Name: a.name, Kind: spec.SystemSolace, Solace: sb}
		case spec.SystemMQ:
			mb := &MQBinder{
				QueueManager: a.qm,
				Channel:      a.ch,
				ConnName:     a.conn,
				User:         secretRef(stableName(a.name, "USER"), binderOwner(a.name)+" user", a.mqUser),
				Password:     secretRef(stableName(a.name, "PASSWORD"), binderOwner(a.name)+" password", a.mqPass),
			}
			if a.tls {
				m.MQTLS = true
				// No truststore configured means no bundle to reference: the
				// connection still negotiates TLS, it just trusts the JVM's
				// default store. Referencing an empty bundle would break it.
				if b := buildBundle(a, d, mountStores, secretRef); b != nil {
					mb.SSLBundle = b.Name
					m.Bundles = append(m.Bundles, b)
				} else {
					warn("binder %q: tls is enabled but tls.truststore is not configured; no SSL bundle is emitted and the connection falls back to the JVM default truststore", a.name)
				}
			}
			var props []Prop
			if a.cipher != "" {
				props = append(props, Prop{Key: "WMQ_SSL_CIPHER_SUITE", Val: a.cipher})
			}
			props = appendPassthrough(props, a.pass2, &warns, binderOwner(a.name))
			mb.AddlProps = props
			a.binder = &Binder{Name: a.name, Kind: spec.SystemMQ, MQ: mb}
		}
		m.Binders = append(m.Binders, a.binder)
	}
	// Mandatory internal binder, always last.
	m.Binders = append(m.Binders, &Binder{Name: "undefined", Kind: "undefined"})

	// ---- pass 2: bindings, jms/solace options, workflow enables ----------------
	for i, w := range rwfs {
		in := fmt.Sprintf("input-%d", i)
		out := fmt.Sprintf("output-%d", i)
		srcAcc := byKey[w.src.DedupKey()]
		tgtAcc := byKey[w.tgt.DedupKey()]
		m.Bindings = append(m.Bindings,
			&Binding{Name: in, Dest: w.src.Dest, Binder: srcAcc.binder.Name},
			&Binding{Name: out, Dest: w.tgt.Dest, Binder: tgtAcc.binder.Name},
		)
		m.emitBindingOptions(in, w.src, true, w.file)
		m.emitBindingOptions(out, w.tgt, false, w.file)
		m.Workflows = append(m.Workflows, WorkflowEnable{ID: i, Enabled: w.enabled})

		if srcAcc.binder.Name == tgtAcc.binder.Name && w.src.Dest == w.tgt.Dest {
			warn("workflow %q: source and target resolve to the same binder and destination %q (possible message loop)", w.file, w.src.Dest)
		}
	}

	// A management session that resolves to a connection a workflow already uses
	// is the same credential, so it reuses that binder's stable names: one
	// secret, mounted once, rather than two files holding the same value under
	// two spellings. DedupKey excludes the destination, so a bare connection and
	// the workflow side built from it land on the same accumulator -- which is
	// what makes this work for the inline session as well as for conn-ref.
	leaderNames := func(sess spec.Side) (string, string) {
		if a := byKey[sess.DedupKey()]; a != nil && a.binder != nil && a.kind == spec.SystemSolace {
			return stableName(a.name, "CLIENT_USERNAME"), stableName(a.name, "CLIENT_PASSWORD")
		}
		return LeaderUsernameName, LeaderPasswordName
	}

	// leader-election block (active_active / active_standby); standalone/absent ⇒ nil.
	m.LeaderElection = buildLeaderElection(d, mountStores, secretRef, leaderNames, &warns)

	// The status script needs actuator access in every deployment, so the
	// exposure list and the reserved status account are forced onto the model
	// regardless of what the operator configured.
	applyStatusAccess(m, d, opts.StatusPassword, secretRef)

	return m, warns
}

// emitBindingOptions appends jms/solace binding entries for one side.
func (m *Model) emitBindingOptions(name string, s spec.Side, isInput bool, file string) {
	role := "producer"
	var passNode *yaml.Node
	if isInput {
		role = "consumer"
		passNode = s.Consumer
	} else {
		passNode = s.Producer
	}
	extra := nodeToProps(passNode)

	switch s.System {
	case spec.SystemMQ:
		jb := &JMSBinding{Name: name, Role: role, DestType: s.DestKind, Extra: extra}
		// An MQ topic source is always a durable subscription (auto-named).
		if isInput && s.DestKind == spec.DestTopic {
			jb.Durable = DurableName(s.ConnName, s.QueueManager, s.Dest, file)
		}
		m.JMSBindings = append(m.JMSBindings, jb)
	case spec.SystemSolace:
		destType := ""
		// Solace defaults: a consumer binds to a queue endpoint, a producer publishes
		// to a topic. Emit destination-type only when the side departs from that default.
		switch {
		case isInput && s.DestKind == spec.DestTopic:
			destType = spec.DestTopic // consuming a topic subscription (non-default)
		case !isInput && s.DestKind == spec.DestQueue:
			destType = spec.DestQueue // producing to a queue (non-default)
		}
		if destType == "" && len(extra) == 0 {
			return // matches the Solace default for this role: nothing to emit
		}
		m.SolaceBindings = append(m.SolaceBindings, &SolaceBinding{Name: name, Role: role, DestType: destType, Extra: extra})
	}
}

// ---- helpers -----------------------------------------------------------------

// displayName is a stable label for warnings emitted before names are assigned.
func displayName(a *acc) string {
	switch {
	case a.name != "":
		return a.name
	case a.connName != "":
		return a.connName
	case a.kind == spec.SystemSolace:
		return "solace:" + a.vpn
	default:
		return "mq:" + a.qm
	}
}

// binderOwner labels a binder in a passthrough warning. The label is the
// caller's business because the leader-election management session runs the
// same passthrough rules without being a binder -- see leaderSessionOwner.
func binderOwner(name string) string { return fmt.Sprintf("binder %q", name) }

// leaderSessionOwner is the passthrough-warning label for the leader-election
// management session, which has no binder name to quote.
const leaderSessionOwner = "leader-election session"

func isTCPS(host string) bool { return strings.HasPrefix(host, "tcps://") }

// sanitize keeps [A-Za-z0-9-] and replaces every other rune with '-'.
func sanitize(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '-':
			b.WriteRune(r)
		default:
			b.WriteRune('-')
		}
	}
	return b.String()
}

// buildBundle assembles the JKS SSL bundle for one TLS-enabled MQ binder, or nil
// when there is no truststore to point it at. A bundle exists to name a
// truststore: emitting one without would put empty location/password/type values
// into application.yml, which the connector reads as a configured-but-broken
// trust store rather than as "not configured".
func buildBundle(a *acc, d *spec.Defaults, mount bool, secretRef secretFn) *Bundle {
	ts := d.TLS.Truststore
	if ts == nil || ts.File == "" {
		return nil
	}
	b := &Bundle{Name: a.name + "-bundle"}
	b.TruststoreLoc = tls.StorePath(ts.File, mount)
	b.TruststorePwd = secretRef(TruststorePasswordName, storeOrigin(TruststorePasswordName), ts.Secret())
	b.TruststoreTyp = ts.Type
	if a.keyAlias != "" {
		if ks := d.TLS.Keystore; ks != nil {
			b.HasKeystore = true
			b.KeystoreLoc = tls.StorePath(ks.File, mount)
			b.KeystorePwd = secretRef(KeystorePasswordName, storeOrigin(KeystorePasswordName), ks.Secret())
			b.KeystoreTyp = ks.Type
			b.KeyAlias = a.keyAlias
		}
	}
	return b
}

// buildLeaderElection assembles the leader-election model for active_active /
// active_standby (nil for standalone/absent). The management session carries the
// same key set as a Solace binder -- the connector documents session.* as the
// same interface as solace.java.* -- so solace-defaults and the connection's own
// verbatim api-properties land here too, on top of the shared
// truststore/keystore wiring (config vs deploy path via mount).
func buildLeaderElection(d *spec.Defaults, mount bool, secretRef secretFn, names leaderNameFn, warns *[]string) *LeaderElectionModel {
	le := d.LeaderElection
	if !le.Present || le.Mode == "" || le.Mode == spec.LeaderStandalone {
		return nil
	}
	m := &LeaderElectionModel{Mode: le.Mode, Queue: le.Queue, FailOver: le.FailOver}
	var sess *spec.Side
	if le.ConnRef != "" {
		if c, ok := d.Connections[le.ConnRef]; ok {
			sess = &c
		}
	} else if le.Session != nil {
		sess = le.Session
	}
	if sess != nil {
		userName, passName := names(*sess)
		s := &Session{
			Host:       sess.Host,
			MsgVPN:     sess.MsgVPN,
			ClientUser: secretRef(userName, leaderSessionOwner+" client-username", sess.Username()),
			ClientPass: secretRef(passName, leaderSessionOwner+" client-password", sess.Secret()),
			Extras:     nodeToProps(d.SolaceDefaults),
		}
		var props []Prop
		if isTCPS(sess.Host) {
			for _, kv := range tls.SolaceProps(d, sess.KeyAlias, mount, storeSecret(secretRef)) {
				props = append(props, Prop{Key: kv.Key, Val: kv.Val})
			}
		}
		// One session resolves to exactly one connection, so unlike a binder --
		// which accumulates passthrough from every side that dedups onto it --
		// there is nothing to mergeProp across: the mapping goes straight through.
		s.APIProps = appendPassthrough(props, nodeToProps(sess.APIProps), warns, leaderSessionOwner)
		m.Session = s
	}
	return m
}

// managementExposure is the actuator exposure list every render carries. It is
// fixed rather than derived from the operator's configuration: the generated
// status script depends on leaderelection and workflows being exposed, and a
// fixed list is the only way the tool can promise a running instance is
// queryable.
const managementExposure = "health,info,metrics,leaderelection,workflows"

// applyStatusAccess forces the actuator surface the generated status script
// depends on: the fixed exposure list and the reserved read-only account
// (spec.StatusUserName), regardless of what the operator configured.
func applyStatusAccess(m *Model, d *spec.Defaults, statusPassword string, secretRef secretFn) {
	m.Management.Exposure = managementExposure

	// Management users authenticate against the actuator endpoint, so their
	// passwords are credentials like any other. The slice is copied first
	// because m.Security was assigned from d.Security and the two share one
	// backing array: rewriting passwords in place would replace the caller's
	// own parsed values with ${...} placeholders, leaving Defaults describing
	// something other than what the operator wrote.
	//
	// Only the password is rewritten. Roles pass through as authored -- the
	// copy still shares their backing array with the caller, which stays safe
	// precisely because nothing here mutates them.
	users := make([]spec.User, len(d.Security.Users))
	copy(users, d.Security.Users)
	for i, u := range users {
		users[i].Password = secretRef(securityUserPasswordName(u.Name), fmt.Sprintf("security.users[%s].password", u.Name), u.Secret())
		users[i].PasswordEnv = ""
	}

	// The reserved account is always appended last, after the operator's own
	// users -- management security is always on now (see spec.StatusUserName),
	// so there is no enabled/disabled branch left to gate on.
	//
	// This one password is a literal rather than a secretRef placeholder --
	// the deliberate exception to the rule the other users follow. Every
	// credential in the secrets model is resolved from the operator's
	// environment at deploy time, so routing this one through it would put
	// an export back on the operator for an account the tool owns end to
	// end and generates a fresh password for. The cost is that the value
	// sits in the generated config artifacts (on Kubernetes, a ConfigMap
	// rather than a Secret), which is why the account is read-only and the
	// password is regenerated on every render. The script's own
	// ${...} -> /run/secrets fallback still covers configs written before
	// this account existed.
	m.Security.Users = append(users, spec.User{Name: spec.StatusUserName, Password: statusPassword})
}

// nodeToProps converts a mapping node into ordered Props (scalar value -> Val,
// otherwise -> Sub). Returns nil for a nil or non-mapping node.
func nodeToProps(node *yaml.Node) []Prop {
	if node == nil || node.Kind != yaml.MappingNode {
		return nil
	}
	var out []Prop
	for i := 0; i+1 < len(node.Content); i += 2 {
		k := node.Content[i].Value
		v := node.Content[i+1]
		if v.Kind == yaml.ScalarNode {
			out = append(out, Prop{Key: k, Val: FormatScalar(v)})
		} else {
			out = append(out, Prop{Key: k, Sub: v})
		}
	}
	return out
}

// FormatScalar renders a scalar node, re-applying quoting when the source used
// it so verbatim passthrough (e.g. a quoted DN) survives faithfully. Shared with
// the render package, which calls it while walking nested (Sub) passthrough
// trees at arbitrary depth.
//
// A literal (|) or folded (>) source scalar keeps its embedded newlines here;
// the render layer re-emits those as a block scalar, since only it knows the
// indent (see render.writeScalar).
func FormatScalar(n *yaml.Node) string {
	switch n.Style {
	case yaml.DoubleQuotedStyle:
		return strconv.Quote(n.Value)
	case yaml.SingleQuotedStyle:
		return "'" + strings.ReplaceAll(n.Value, "'", "''") + "'"
	default:
		return n.Value
	}
}

// yamlIndicators are the characters that, in first position, give a plain scalar
// a meaning other than "this text": block/flow structure, anchors, tags, comments.
const yamlIndicators = "-?:,[]{}#&*!|>'\"%@`"

// plainNotString matches the plain scalars a YAML parser reads back as a bool,
// null, or number rather than a string -- so a password of "0123" or "no" needs
// quoting to survive as text.
var plainNotString = regexp.MustCompile(`^(?i:true|false|yes|no|on|off|y|n|null|~|[-+]?(\.inf|\.nan)|[-+]?[0-9][0-9_]*(\.[0-9_]*)?([eE][-+]?[0-9]+)?|0[xXbBoO][0-9a-fA-F_]+)$`)

// QuoteScalar renders a Go string as a YAML scalar, double-quoting it only when
// a plain scalar would not read back as the same string.
//
// The renderer builds application.yml by concatenating "key: " with a value, so
// an unquoted value carrying ": ", " #", a leading indicator, or an embedded
// newline silently restructures the document -- a realistic connector password
// ("p@ss #1") is enough to do it. Quoting only when required keeps every ordinary
// value (hosts, ${VAR} placeholders, queue names) byte-for-byte as before.
func QuoteScalar(s string) string {
	if !needsQuote(s) {
		return s
	}
	return strconv.Quote(s)
}

func needsQuote(s string) bool {
	if s == "" {
		return true
	}
	// Leading/trailing whitespace is stripped by the parser, and control
	// characters cannot appear in a plain scalar at all.
	if strings.TrimSpace(s) != s {
		return true
	}
	for _, r := range s {
		if r < 0x20 || r == 0x7f {
			return true
		}
	}
	if strings.IndexByte(yamlIndicators, s[0]) >= 0 || plainNotString.MatchString(s) {
		return true
	}
	// ": " opens a nested mapping, a trailing ':' a mapping key, " #" a comment.
	return strings.Contains(s, ": ") || strings.Contains(s, " #") || strings.HasSuffix(s, ":")
}

// mergeProp appends p unless its key already exists, in which case the value is
// updated (last writer wins) with a warning on a real change.
func mergeProp(list []Prop, p Prop, warns *[]string, binder string) []Prop {
	for i := range list {
		if list[i].Key == p.Key {
			if list[i].Val != p.Val || (p.Sub != nil) {
				*warns = append(*warns, fmt.Sprintf("binder %q: passthrough key %q set more than once; last (by filename) wins", binder, p.Key))
				list[i] = p
			}
			return list
		}
	}
	return append(list, p)
}

// appendPassthrough appends verbatim props after the tool-managed props, dropping
// (with a warning) any passthrough key that collides with a tool-managed key.
// owner is the already-formatted label the warning names: binderOwner(...) for a
// binder, leaderSessionOwner for the management session.
func appendPassthrough(tool []Prop, pass []Prop, warns *[]string, owner string) []Prop {
	toolKeys := map[string]bool{}
	for _, p := range tool {
		toolKeys[p.Key] = true
	}
	for _, p := range pass {
		if toolKeys[p.Key] {
			*warns = append(*warns, fmt.Sprintf("%s: passthrough overrides tool-managed key %q; tool value kept", owner, p.Key))
			continue
		}
		tool = append(tool, p)
	}
	return tool
}
