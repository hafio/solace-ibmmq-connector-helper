package validate

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/spec"
)

// The tool writes no certificate-validation policy into a Solace connection's
// api-properties: the Solace API validates the broker certificate by default,
// and a connection that wants otherwise writes SSL_VALIDATE_CERTIFICATE itself,
// which reaches the API unchanged. What the tool does instead is report the
// posture it can see, here in validate.Run so the validate verb, generate,
// deploy and the generator page all show it. Nothing here changes output or the
// exit status except P2, two sides of one binder that disagree.

// validateCertificateKey is the JCSMP property that switches broker-certificate
// validation (and with it the host-name and expiry checks) on or off.
const validateCertificateKey = "SSL_VALIDATE_CERTIFICATE"

// postureSite is one place a Solace connection's api-properties are written:
// a connection in env.yaml, an inline workflow side, or the inline
// leader-election session. A conn-ref side is never a site -- it carries its
// connection's api-properties, which warn once, on the connection.
type postureSite struct {
	file  string
	label string // what the finding opens with: "connections.edge", "source", "leader-election session"
	side  spec.Side
	conn  string // the connection's name for a connections.<name> site, else ""
}

// stated returns the SSL_VALIDATE_CERTIFICATE entry of s's api-properties in its
// exact spelling, or nil. Pairs are walked directly, the way consolidate reads
// them, rather than through mappingEntries, whose duplicate-key error would be
// new for specs that generate today.
func stated(s spec.Side) *yaml.Node {
	m := s.APIProps
	if m == nil || m.Kind != yaml.MappingNode {
		return nil
	}
	for i := 0; i+1 < len(m.Content); i += 2 {
		if m.Content[i].Value == validateCertificateKey {
			return m.Content[i+1]
		}
	}
	return nil
}

// validateLiteral classifies a SSL_VALIDATE_CERTIFICATE value: "true" or
// "false" for those words in any case, "" for anything else -- a ${VAR}, a
// yes/no/0/1, a mapping -- whose meaning the tool cannot vouch for.
func validateLiteral(v *yaml.Node) string {
	if v == nil || v.Kind != yaml.ScalarNode {
		return ""
	}
	switch strings.ToLower(strings.TrimSpace(v.Value)) {
	case "true":
		return "true"
	case "false":
		return "false"
	}
	return ""
}

// statedOff reports whether s's api-properties turn validation off.
func statedOff(s spec.Side) bool { return validateLiteral(stated(s)) == "false" }

// statedText is a value as it would be rendered, quoting included, so two
// sides whose values the connector would read differently never compare equal.
func statedText(v *yaml.Node) string {
	if v.Kind != yaml.ScalarNode {
		return "<" + spec.YAMLKind(v) + ">"
	}
	switch v.Style {
	case yaml.DoubleQuotedStyle:
		return strconv.Quote(v.Value)
	case yaml.SingleQuotedStyle:
		return "'" + strings.ReplaceAll(v.Value, "'", "''") + "'"
	}
	return v.Value
}

func isTCPSHost(s spec.Side) bool { return strings.HasPrefix(s.Host, "tcps://") }

// postureSites lists every Solace site in the order checkAllExtraKeys walks
// them: connections sorted by name, each workflow's inline sides, then the
// inline leader-election session -- the last only under active_active /
// active_standby without a conn-ref, the only case consolidate renders it.
func postureSites(wfs []spec.Workflow, d *spec.Defaults) []postureSite {
	var out []postureSite
	names := make([]string, 0, len(d.Connections))
	for name := range d.Connections {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		if c := d.Connections[name]; c.System == spec.SystemSolace {
			out = append(out, postureSite{file: fileEnv, label: "connections." + name, side: c, conn: name})
		}
	}
	for _, wf := range wfs {
		for _, s := range []struct {
			label string
			side  spec.Side
			set   bool
		}{{"source", wf.Source, wf.SourceSet}, {"target", wf.Target, wf.TargetSet}} {
			if s.set && s.side.System == spec.SystemSolace && s.side.ConnRef == "" {
				out = append(out, postureSite{file: wf.File, label: s.label, side: s.side})
			}
		}
	}
	if sess := inlineSession(d); sess != nil {
		out = append(out, postureSite{file: fileEnv, label: "leader-election session", side: *sess})
	}
	return out
}

// inlineSession is the inline leader-election session consolidate renders, or
// nil: none under standalone, and none when a conn-ref names the connection
// instead (conn-ref wins).
func inlineSession(d *spec.Defaults) *spec.Side {
	le := d.LeaderElection
	mode := le.EffectiveMode()
	if !le.Present || (mode != spec.LeaderActiveActive && mode != spec.LeaderActiveStby) {
		return nil
	}
	if le.ConnRef != "" || le.Session == nil || le.Session.System != spec.SystemSolace {
		return nil
	}
	return le.Session
}

// leaderConnRef is the connection the rendered leader-election session follows,
// or "" when it is inline or not rendered.
func leaderConnRef(d *spec.Defaults) string {
	le := d.LeaderElection
	mode := le.EffectiveMode()
	if !le.Present || (mode != spec.LeaderActiveActive && mode != spec.LeaderActiveStby) {
		return ""
	}
	return le.ConnRef
}

// checkValidatePosture reports the certificate-validation posture of every
// tcps:// Solace connection (P1, P1u, P3s), refuses sides of one binder that
// disagree on it (P2), and warns when an inline leader-election session and the
// workflow sides on the same broker connection disagree (P6). wfs are the raw
// workflows (sites are where a value is written); resolved carries every side
// with its connection materialised (binders are what consolidate builds).
func checkValidatePosture(add, warn func(string, string, ...any), wfs, resolved []spec.Workflow, d *spec.Defaults) {
	// sidesOn counts the resolved Solace workflow sides on each binder.
	sidesOn := map[string]int{}
	for _, wf := range resolved {
		for _, s := range []spec.Side{wf.Source, wf.Target} {
			if s.System == spec.SystemSolace {
				sidesOn[s.DedupKey()]++
			}
		}
	}
	leaderRef := leaderConnRef(d)

	for _, st := range postureSites(wfs, d) {
		if !isTCPSHost(st.side) {
			continue
		}
		where := st.label + " solace"
		if m := st.side.APIProps; m != nil && m.Kind == yaml.MappingNode {
			for i := 0; i+1 < len(m.Content); i += 2 {
				k := m.Content[i].Value
				if k != validateCertificateKey && spec.CanonicalKey(k) == spec.CanonicalKey(validateCertificateKey) {
					warn(st.file, "%s: api-properties %q is another spelling of %s; the Solace API ignores property names it does not recognise, so this entry likely has no effect -- write %s", where, k, validateCertificateKey, validateCertificateKey)
				}
			}
		}
		v := stated(st.side)
		if v == nil {
			continue
		}
		switch validateLiteral(v) {
		case "false":
			var shared []string
			if n := sidesOn[st.side.DedupKey()]; n > 1 {
				shared = append(shared, fmt.Sprintf("every workflow side on this connection (%d)", n))
			}
			if st.conn != "" && st.conn == leaderRef {
				shared = append(shared, "the leader-election session")
			}
			suffix := ""
			if len(shared) > 0 {
				suffix = "; it applies to " + strings.Join(shared, " and to ")
			}
			warn(st.file, "%s: api-properties %s: false turns server-certificate validation off for %s -- the broker's certificate, host name and expiry are not checked, so anything on the network path can impersonate the broker and read what this connection sends; either trust the broker's CA and remove the entry, or keep it deliberately%s", where, validateCertificateKey, st.side.Host, suffix)
		case "true":
			// The Solace API's default, stated: nothing to report.
		default:
			warn(st.file, "%s: api-properties %s is %q, not the literal true or false, so whether %s validates the broker certificate cannot be reported; write true or false", where, validateCertificateKey, statedText(v), st.side.Host)
		}
	}

	checkValidateConflicts(add, resolved)
	checkSessionPosture(warn, resolved, d)
}

// checkValidateConflicts is P2: sides that collapse onto one binder share one
// api-properties block, merged last-by-filename, so two stated values that
// differ in any way -- quoting included -- would leave one side's posture to
// file order. Absent beside stated is not a conflict: the stated value applies
// to the whole binder, which P1's count reports. Modelled on
// checkStoreSelectionConflicts, but each side is reported by where its value was
// written: a conn-ref side by its connection in env.yaml.
func checkValidateConflicts(add func(string, string, ...any), resolved []spec.Workflow) {
	type seen struct{ text, origin string }
	byKey := map[string]seen{}
	for _, wf := range resolved {
		for _, s := range []spec.Side{wf.Source, wf.Target} {
			if s.System != spec.SystemSolace || !isTCPSHost(s) {
				continue
			}
			v := stated(s)
			if v == nil {
				continue
			}
			file, origin := wf.File, wf.File
			if s.ConnRef != "" {
				file, origin = fileEnv, fileEnv+" connections."+s.ConnRef
			}
			text := statedText(v)
			key := s.DedupKey()
			if prev, ok := byKey[key]; ok {
				if prev.text != text {
					add(file, "conflicting api-properties %s %s vs %s for the same binder (also in %s): the sides share one connection, so they must agree -- set the same value on every side, or define the connection once under connections: and use conn-ref", validateCertificateKey, text, prev.text, prev.origin)
				}
				continue
			}
			byKey[key] = seen{text: text, origin: origin}
		}
	}
}

// checkSessionPosture is P6: an inline leader-election session is a separate
// connection with its own api-properties, so on the same broker tuple as
// workflow sides it can validate while they do not, or the reverse -- usually
// an oversight, and one that only shows as a failed (or unverified) handshake at
// run time. A conn-ref session follows its connection and cannot diverge.
func checkSessionPosture(warn func(string, string, ...any), resolved []spec.Workflow, d *spec.Defaults) {
	sess := inlineSession(d)
	if sess == nil || !isTCPSHost(*sess) {
		return
	}
	key := sess.DedupKey()
	files := map[string]bool{}
	groupOff := false
	for _, wf := range resolved {
		for _, s := range []spec.Side{wf.Source, wf.Target} {
			if s.System != spec.SystemSolace || s.DedupKey() != key {
				continue
			}
			files[wf.File] = true
			if statedOff(s) {
				groupOff = true
			}
		}
	}
	if len(files) == 0 || groupOff == statedOff(*sess) {
		return
	}
	list := make([]string, 0, len(files))
	for f := range files {
		list = append(list, f)
	}
	sort.Strings(list)
	sessVerb, sidesVerb := "validates", "do not validate"
	if statedOff(*sess) {
		sessVerb, sidesVerb = "does not validate", "validate"
	}
	warn(fileEnv, "the leader-election session %s the broker certificate, but the workflow sides on the same broker connection (%s) %s it; the session is configured separately -- set %s the same way in the session's api-properties, or use leader-election conn-ref so it follows the connection", sessVerb, strings.Join(list, ", "), sidesVerb, validateCertificateKey)
}
