package validate

import (
	"regexp"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/spec"
)

// Every key of a solace:/mq: block the tool does not read itself is passed
// through verbatim to the binder (spec.Side.Extra, captured in spec/extra.go).
// There is no allowlist -- the connector owns that vocabulary -- so what is
// checked here is the handful of ways a passthrough goes wrong: another
// spelling of a key the tool handles (the connector would get the property
// twice, or a credential would land in application.yml as a literal), a key
// that cannot be written as a YAML key, and the warnings worth a look -- a
// likely typo, a secret written as a literal, an -env name the tool does not
// resolve, and a few keys that fight the tool's own TLS wiring.

// extraKeyRE is what a passed-through key may look like: it is written into
// application.yml unquoted, so it is held to a plain property name.
var extraKeyRE = regexp.MustCompile(`^[A-Za-z0-9_][A-Za-z0-9_.\-\[\]]*$`)

// secretKeyRE marks a key whose value is probably sensitive.
var secretKeyRE = regexp.MustCompile(`(?i)password|passwd|secret|token|passphrase`)

// extraSite is one place a block's other keys are read from.
type extraSite struct {
	file   string
	label  string // what the finding opens with: "source", "connections.edge", "leader-election session", "mq-defaults"
	prop   string // the property prefix the keys land under: ibm.mq, solace.java, solace.connector.management.session
	system string
	side   spec.Side // the block's known keys, for the checks that depend on them (cipher, tls)
	extra  *yaml.Node
	// defaults marks mq-defaults, a block with no connection of its own: a key
	// the tool reads per connection has no place there at all.
	defaults bool
}

// checkAllExtraKeys runs checkExtraKeys over every block that can carry other
// keys: each connection, each workflow's raw sides (so a key is reported where
// it was written, once), the inline leader-election session, and mq-defaults.
func checkAllExtraKeys(add, warn func(string, string, ...any), wfs []spec.Workflow, d *spec.Defaults) {
	mode := d.LeaderElection.EffectiveMode()
	names := make([]string, 0, len(d.Connections))
	for name := range d.Connections {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		c := d.Connections[name]
		checkExtraKeys(add, warn, extraSite{file: fileEnv, label: "connections." + name, prop: propPrefix(c.System), system: c.System, side: c, extra: c.Extra}, d, mode)
	}
	for _, wf := range wfs {
		for _, s := range []struct {
			label string
			side  spec.Side
			set   bool
		}{{"source", wf.Source, wf.SourceSet}, {"target", wf.Target, wf.TargetSet}} {
			if s.set && s.side.HasSystem() {
				checkExtraKeys(add, warn, extraSite{file: wf.File, label: s.label, prop: propPrefix(s.side.System), system: s.side.System, side: s.side, extra: s.side.Extra}, d, mode)
			}
		}
	}
	if sess := d.LeaderElection.Session; sess != nil {
		checkExtraKeys(add, warn, extraSite{file: fileEnv, label: "leader-election session", prop: "solace.connector.management.session", system: spec.SystemSolace, side: *sess, extra: sess.Extra}, d, mode)
	}
	if d.MQDefaults != nil {
		checkExtraKeys(add, warn, extraSite{file: fileEnv, label: "mq-defaults", prop: "ibm.mq", system: spec.SystemMQ, extra: d.MQDefaults, defaults: true}, d, mode)
	}
}

// propPrefix is where a system's connection keys land in application.yml.
func propPrefix(system string) string {
	if system == spec.SystemSolace {
		return "solace.java"
	}
	return "ibm.mq"
}

// checkExtraKeys checks one block's other keys. Errors stop at the first one
// per key; the warnings are advisory and can stack.
func checkExtraKeys(add, warn func(string, string, ...any), at extraSite, d *spec.Defaults, leaderMode string) {
	if at.extra == nil || at.extra.Kind != yaml.MappingNode {
		return
	}
	where := at.label + " " + at.system
	creds := spec.CredentialKeys(at.system)
	known := spec.KnownKeys(at.system)
	for _, kv := range mappingEntries(add, at.file, where, at.extra) {
		k := kv.key
		if !extraKeyRE.MatchString(k) {
			add(at.file, "%s: key %q is not a plain property name (letters, digits and . _ - [ ] only), so it cannot be written into application.yml as a key; rename it", where, k)
			continue
		}
		ck := spec.CanonicalKey(k)
		if at.system == spec.SystemMQ && ck == "sslbundle" {
			add(at.file, "%s: %q is written by the tool itself (from tls: true and tls.truststore), so it cannot be set here; remove it and enable TLS with tls: true and a tls.truststore in env.yaml instead", where, k)
			continue
		}
		if cred := canonicalMatch(ck, creds); cred != "" {
			if at.defaults {
				add(at.file, "%s: %q is a credential, which belongs on each connection as %s (or %s-env), not in %s", where, k, cred, cred, at.label)
			} else {
				add(at.file, "%s: %q is another spelling of %s, a credential: written this way its value would land in application.yml as a literal instead of being mounted as a secret; write %s (or %s-env) instead", where, k, cred, cred, cred)
			}
			continue
		}
		if schema := canonicalMatch(ck, known); schema != "" {
			if at.defaults {
				add(at.file, "%s: %q is a key the tool reads on each connection, not a default; set %s on the connection instead", where, k, schema)
			} else {
				add(at.file, "%s: %q is another spelling of %s, which the tool reads only in that form; write %s", where, k, schema, schema)
			}
			continue
		}
		if at.system == spec.SystemMQ && ck == "sslciphersuite" && at.side.Cipher != "" {
			add(at.file, "%s: %q and cipher: both set the cipher suite (the tool writes cipher: as additional-properties.WMQ_SSL_CIPHER_SUITE); keep one of the two", where, k)
			continue
		}
		if at.system == spec.SystemMQ && ck == "useibmciphermappings" && at.side.TLS && !isFalse(kv.value) {
			warn(at.file, "%s: %q overrides the -Dcom.ibm.mq.cfg.useIBMCipherMappings=false the tool sets for a TLS connection, so the JCE cipher names in cipher: may no longer apply", where, k)
		}
		if at.system == spec.SystemMQ && ck == "jks" && at.side.TLS && d.TLS.Truststore != nil && d.TLS.Truststore.File != "" {
			warn(at.file, "%s: jks.* is ignored by the connector while ssl-bundle is set; the stores in effect are env.yaml tls.truststore/keystore", where)
		}
		if at.system == spec.SystemSolace && ck == "clientname" && (leaderMode == spec.LeaderActiveActive || leaderMode == spec.LeaderActiveStby) && isLiteral(kv.value) {
			warn(at.file, "%s: %q is fixed, but every replica shares this file and a client name must be unique per connection; use a per-instance value such as ${HOSTNAME}, or leave it for the connector to generate", where, k)
		}
		if lk := strings.ToLower(k); strings.HasSuffix(lk, "-env") || strings.HasSuffix(lk, "_env") {
			warn(at.file, "%s: %q is not a credential the tool resolves, so it is passed through as %s.%s unchanged and the connector reads it as a plain property; the credential keys are %s and %s, each with an -env twin", where, k, at.prop, k, creds[0], creds[1])
		} else {
			warnSecretLiterals(warn, at, k, kv.value)
		}
		if near := nearestKnownKey(ck, known); near != "" {
			warn(at.file, "%s: %q is not a key this tool knows; did you mean %q? It is passed through as %s.%s, which the connector ignores unless it reads that name", where, k, near, at.prop, k)
		}
	}
}

// warnSecretLiterals warns on a value that looks sensitive and is written as a
// literal: the key (or, inside a nested mapping, the leaf key) reads like a
// secret, and the value is a non-empty scalar carrying no ${...} placeholder.
func warnSecretLiterals(warn func(string, string, ...any), at extraSite, path string, v *yaml.Node) {
	switch v.Kind {
	case yaml.MappingNode:
		for i := 0; i+1 < len(v.Content); i += 2 {
			warnSecretLiterals(warn, at, path+"."+v.Content[i].Value, v.Content[i+1])
		}
	case yaml.ScalarNode:
		leaf := path[strings.LastIndex(path, ".")+1:]
		if secretKeyRE.MatchString(leaf) && isLiteral(v) {
			warn(at.file, "%s %s: %q looks like a secret, but its value is written into application.yml as a literal; if it is sensitive, reference a credential the tool already mounts as ${NAME} instead (user guide section 9.2)", at.label, at.system, path)
		}
	}
}

// isLiteral reports whether v is a non-empty scalar with no ${...} placeholder
// in it -- a value Spring will use exactly as written.
func isLiteral(v *yaml.Node) bool {
	return v != nil && v.Kind == yaml.ScalarNode && strings.TrimSpace(v.Value) != "" && !strings.Contains(v.Value, "${")
}

// isFalse reports whether v is the boolean false in any of YAML's spellings.
func isFalse(v *yaml.Node) bool {
	if v == nil || v.Kind != yaml.ScalarNode {
		return false
	}
	switch strings.ToLower(strings.TrimSpace(v.Value)) {
	case "false", "no", "off", "0":
		return true
	}
	return false
}

// canonicalMatch returns the key in keys that ck names in another spelling, or
// "" -- the exact spelling is a known key, handled before any of this runs.
func canonicalMatch(ck string, keys []string) string {
	for _, k := range keys {
		if spec.CanonicalKey(k) == ck {
			return k
		}
	}
	return ""
}

// nearestKnownKey returns the schema key ck is most likely a typo of: within
// one edit of a short key (under six characters once folded) or two of a longer
// one, never equal, and only for a key long enough to be a typo rather than a
// different word (four characters). Ties go to the first key in sorted order.
func nearestKnownKey(ck string, known []string) string {
	if len(ck) < 4 {
		return ""
	}
	best, bestDist := "", -1
	for _, k := range known {
		kc := spec.CanonicalKey(k)
		limit := 2
		if len(kc) < 6 {
			limit = 1
		}
		dist := editDistance(ck, kc)
		if dist == 0 || dist > limit {
			continue
		}
		if bestDist < 0 || dist < bestDist {
			best, bestDist = k, dist
		}
	}
	return best
}

// editDistance is the Levenshtein distance between a and b, over runes.
func editDistance(a, b string) int {
	ra, rb := []rune(a), []rune(b)
	prev := make([]int, len(rb)+1)
	cur := make([]int, len(rb)+1)
	for j := range prev {
		prev[j] = j
	}
	for i := 1; i <= len(ra); i++ {
		cur[0] = i
		for j := 1; j <= len(rb); j++ {
			cost := 1
			if ra[i-1] == rb[j-1] {
				cost = 0
			}
			cur[j] = min(prev[j]+1, cur[j-1]+1, prev[j-1]+cost)
		}
		prev, cur = cur, prev
	}
	return prev[len(rb)]
}
