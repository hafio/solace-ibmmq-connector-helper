// Package consolidate turns the parsed workflow files + defaults into a single
// fully-ordered Model: it deduplicates connections into shared binders, numbers
// workflows by (already-sorted) input order, derives destination-types, computes
// MQ durable-subscription names, and wires Solace + MQ TLS/mTLS. It assumes the
// input already passed validation (no structural or semantic errors).
package consolidate

import (
	"fmt"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-gen/internal/spec"
	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-gen/internal/tls"
)

// acc accumulates the state of one deduplicated binder across all sides that map
// to it, before names and TLS wiring are finalized.
type acc struct {
	key  string
	kind string // spec.SystemSolace | spec.SystemMQ
	name string // resolved after all accs are known

	// Solace tuple.
	host, vpn, user, pass string
	// MQ tuple.
	conn, qm, ch, mqUser, mqPass string

	tls      bool   // TLS on this binder (Solace: tcps:// host; MQ: tls:true)
	keyAlias string // mTLS client key alias ("" = none)
	cipher   string // MQ JCE cipher ("" = none)

	connName string // contributing connection name (first by appearance); "" = purely inline

	pass2 []Prop // merged verbatim passthrough (api-properties / additional-properties)

	binder *Binder
}

// Build consolidates workflows (in the caller's sorted order) with defaults.
// Returns the ordered Model and any non-fatal warnings.
func Build(wfs []spec.Workflow, d *spec.Defaults, mountStores bool) (*Model, []string) {
	if d == nil {
		d = &spec.Defaults{}
	}
	m := &Model{
		Security:     d.Security,
		Management:   d.Management,
		LoggingLevel: d.LoggingLevel,
	}
	var warns []string
	warn := func(format string, a ...any) { warns = append(warns, fmt.Sprintf(format, a...)) }

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
		key := dedupKey(s)
		a := byKey[key]
		if a == nil {
			a = &acc{key: key, kind: s.System}
			switch s.System {
			case spec.SystemSolace:
				a.host, a.vpn, a.user, a.pass = s.Host, s.MsgVPN, s.ClientUser, s.ClientPass
				a.tls = isTCPS(s.Host)
			case spec.SystemMQ:
				a.conn, a.qm, a.ch, a.mqUser, a.mqPass = s.ConnName, s.QueueManager, s.Channel, s.User, s.Password
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

	// ---- assign binder names: the contributing connection name when present, else a
	//      generated sol-conn-N / mq-conn-N; then disambiguate any clash between two
	//      *different* binders with -2/-3. (Dedup by tuple already happened above.) ----
	used := map[string]int{}
	assign := func(base string) string {
		if used[base] == 0 {
			used[base] = 1
			return base
		}
		used[base]++
		return fmt.Sprintf("%s-%d", base, used[base])
	}
	var solN, mqN int
	for _, a := range accs {
		var base string
		switch {
		case a.connName != "":
			base = sanitize(a.connName)
		case a.kind == spec.SystemSolace:
			solN++
			base = fmt.Sprintf("sol-conn-%d", solN)
		default:
			mqN++
			base = fmt.Sprintf("mq-conn-%d", mqN)
		}
		a.name = assign(base)
	}

	// ---- finalize binder + bundle objects -------------------------------------
	for _, a := range accs {
		switch a.kind {
		case spec.SystemSolace:
			sb := &SolaceBinder{Host: a.host, MsgVPN: a.vpn, ClientUser: a.user, ClientPass: a.pass}
			sb.Extras = nodeToProps(d.SolaceDefaults)
			var props []Prop
			if a.tls {
				for _, kv := range tls.SolaceProps(d, a.keyAlias, mountStores) {
					props = append(props, Prop{Key: kv.Key, Val: kv.Val})
				}
			}
			props = appendPassthrough(props, a.pass2, &warns, a.name)
			sb.APIProps = props
			a.binder = &Binder{Name: a.name, Kind: spec.SystemSolace, Solace: sb}
		case spec.SystemMQ:
			mb := &MQBinder{QueueManager: a.qm, Channel: a.ch, ConnName: a.conn, User: a.mqUser, Password: a.mqPass}
			if a.tls {
				mb.SSLBundle = a.name + "-bundle"
				m.MQTLS = true
				m.Bundles = append(m.Bundles, buildBundle(a, d, mountStores))
			}
			var props []Prop
			if a.cipher != "" {
				props = append(props, Prop{Key: "WMQ_SSL_CIPHER_SUITE", Val: a.cipher})
			}
			props = appendPassthrough(props, a.pass2, &warns, a.name)
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
		srcAcc := byKey[dedupKey(w.src)]
		tgtAcc := byKey[dedupKey(w.tgt)]
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

	// leader-election block (active_active / active_standby); standalone/absent ⇒ nil.
	m.LeaderElection = buildLeaderElection(d, mountStores)

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

func dedupKey(s spec.Side) string {
	if s.System == spec.SystemSolace {
		return "S|" + s.Host + "|" + s.MsgVPN + "|" + s.ClientUser
	}
	return "M|" + s.ConnName + "|" + s.QueueManager + "|" + s.Channel + "|" + s.User
}

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

func buildBundle(a *acc, d *spec.Defaults, mount bool) *Bundle {
	b := &Bundle{Name: a.name + "-bundle"}
	if ts := d.TLS.Truststore; ts != nil {
		b.TruststoreLoc = tls.StorePath(ts.File, mount)
		b.TruststorePwd = ts.Password
		b.TruststoreTyp = ts.Type
	}
	if a.keyAlias != "" {
		if ks := d.TLS.Keystore; ks != nil {
			b.HasKeystore = true
			b.KeystoreLoc = tls.StorePath(ks.File, mount)
			b.KeystorePwd = ks.Password
			b.KeystoreTyp = ks.Type
			b.KeyAlias = a.keyAlias
		}
	}
	return b
}

// buildLeaderElection assembles the leader-election model for active_active /
// active_standby (nil for standalone/absent). The management session's TLS
// api-properties reuse the shared truststore/keystore wiring (config vs deploy
// path via mount).
func buildLeaderElection(d *spec.Defaults, mount bool) *LeaderElectionModel {
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
		s := &Session{Host: sess.Host, MsgVPN: sess.MsgVPN, ClientUser: sess.ClientUser, ClientPass: sess.ClientPass}
		if isTCPS(sess.Host) {
			for _, kv := range tls.SolaceProps(d, sess.KeyAlias, mount) {
				s.APIProps = append(s.APIProps, Prop{Key: kv.Key, Val: kv.Val})
			}
		}
		m.Session = s
	}
	return m
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
			out = append(out, Prop{Key: k, Val: formatScalar(v)})
		} else {
			out = append(out, Prop{Key: k, Sub: v})
		}
	}
	return out
}

// formatScalar renders a scalar node, re-applying quoting when the source used
// it so verbatim passthrough (e.g. a quoted DN) survives faithfully.
func formatScalar(n *yaml.Node) string {
	switch n.Style {
	case yaml.DoubleQuotedStyle:
		return strconv.Quote(n.Value)
	case yaml.SingleQuotedStyle:
		return "'" + strings.ReplaceAll(n.Value, "'", "''") + "'"
	default:
		return n.Value
	}
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
func appendPassthrough(tool []Prop, pass []Prop, warns *[]string, binder string) []Prop {
	toolKeys := map[string]bool{}
	for _, p := range tool {
		toolKeys[p.Key] = true
	}
	for _, p := range pass {
		if toolKeys[p.Key] {
			*warns = append(*warns, fmt.Sprintf("binder %q: passthrough overrides tool-managed key %q; tool value kept", binder, p.Key))
			continue
		}
		tool = append(tool, p)
	}
	return tool
}
