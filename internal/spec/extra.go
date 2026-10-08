package spec

import (
	"reflect"
	"sort"
	"strings"
	"unicode"

	"gopkg.in/yaml.v3"
)

// A solace:/mq: block is read into a fixed struct, and the connector reads far
// more connection properties than that struct names -- IBM's MQ starter alone
// has dozens (user-authentication-mqcsp, client-id, ccdt-url, pool.*, ...).
// Rather than enumerate them, every key the struct does not name is carried
// through verbatim as Side.Extra and written beside the keys the tool emits,
// under the binder's ibm.mq.* or solace.java.* (and the leader-election
// session). This file is the capture half; validate/extra.go is the check half.

// knownKeys reads the yaml keys a raw side struct decodes, the way yaml.v3
// names them: the tag's first element, a "-" tag decodes nothing, and an
// untagged field is its name in lower case.
func knownKeys(t reflect.Type) map[string]bool {
	keys := map[string]bool{}
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if f.PkgPath != "" {
			continue // unexported: yaml.v3 skips it too
		}
		name, _, _ := strings.Cut(f.Tag.Get("yaml"), ",")
		switch name {
		case "-":
			continue
		case "":
			name = strings.ToLower(f.Name)
		}
		keys[name] = true
	}
	return keys
}

var (
	solaceKeys = knownKeys(reflect.TypeOf(rawSolace{}))
	mqKeys     = knownKeys(reflect.TypeOf(rawMQ{}))
)

// KnownKeys lists the keys the tool itself reads under a solace: or mq: block,
// sorted -- the schema, derived from the raw struct so it cannot drift from
// what ParseWorkflow decodes. Every other key is passed through (Side.Extra).
func KnownKeys(system string) []string {
	var set map[string]bool
	switch system {
	case SystemSolace:
		set = solaceKeys
	case SystemMQ:
		set = mqKeys
	default:
		return nil
	}
	out := make([]string, 0, len(set))
	for k := range set {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// ToolManagedKeys lists the keys the tool writes itself under a binder's
// solace.java.* or ibm.mq.* (render.go), so a passed-through key may not set
// them: the connector would receive the property twice. ssl-bundle is the one
// that is not also a schema key -- it is derived from tls: and tls.truststore.
func ToolManagedKeys(system string) []string {
	switch system {
	case SystemSolace:
		return []string{"host", "msg-vpn", "client-username", "client-password", "api-properties"}
	case SystemMQ:
		return []string{"queue-manager", "channel", "conn-name", "user", "password", "ssl-bundle", "additional-properties"}
	}
	return nil
}

// CredentialKeys lists the keys whose values are credentials under each
// system: each goes through the secrets model, never into application.yml as
// written, so another spelling of one must not pass through.
func CredentialKeys(system string) []string {
	switch system {
	case SystemSolace:
		return []string{"client-username", "client-password"}
	case SystemMQ:
		return []string{"user", "password"}
	}
	return nil
}

// CanonicalKey folds a property name the way Spring's relaxed binding does
// when it matches a key to a property: case and the - and _ separators are
// ignored, so user-authentication-mqcsp, userAuthenticationMQCSP and
// USER_AUTHENTICATION_MQCSP are one property.
func CanonicalKey(k string) string {
	return strings.Map(func(r rune) rune {
		if r == '-' || r == '_' {
			return -1
		}
		return unicode.ToLower(r)
	}, k)
}

// solaceBlock and mqBlock are the raw structs without their methods, so the
// typed decode inside UnmarshalYAML cannot recurse into itself.
type (
	solaceBlock rawSolace
	mqBlock     rawMQ
)

// UnmarshalYAML decodes the known keys as before and keeps every other key of
// the block in Extra. yaml.v3 calls it with the block already resolved, so a
// `solace: *shared` alias is read the same as the block it points at.
func (r *rawSolace) UnmarshalYAML(n *yaml.Node) error {
	if err := n.Decode((*solaceBlock)(r)); err != nil {
		return err
	}
	if x := extraKeys(n, solaceKeys); x != nil {
		r.Extra = *x
	}
	return nil
}

// UnmarshalYAML is rawSolace's, for an mq: block.
func (r *rawMQ) UnmarshalYAML(n *yaml.Node) error {
	if err := n.Decode((*mqBlock)(r)); err != nil {
		return err
	}
	if x := extraKeys(n, mqKeys); x != nil {
		r.Extra = *x
	}
	return nil
}

// extraKeys returns a mapping of every entry of block whose key is not in
// known -- in file order, each key and value as written, so a value's quoting
// and a nested mapping survive -- or nil when there is none. A key starting
// with transformPrefix is left out: the transform scan reports it as misplaced.
func extraKeys(block *yaml.Node, known map[string]bool) *yaml.Node {
	block = deref(block)
	if block == nil || block.Kind != yaml.MappingNode {
		return nil
	}
	out := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
	for _, kv := range flattenBlock(block) {
		if k := kv.key.Value; known[k] || strings.HasPrefix(k, transformPrefix) {
			continue
		}
		out.Content = append(out.Content, kv.key, kv.value)
	}
	if len(out.Content) == 0 {
		return nil
	}
	return out
}

// keyValue is one entry of a mapping, both nodes with any alias followed.
type keyValue struct {
	key, value *yaml.Node
}

// flattenBlock lists block's entries the way the typed decode sees them. A <<
// merge key is expanded in place: a key a mapping writes itself beats one it
// merges in (at every level, so a merge source's own key beats what that
// source merges in turn), an earlier merge source beats a later one, and the
// merge key itself is never an entry. A key written twice in one mapping never
// gets here: the typed decode refuses it first.
func flattenBlock(block *yaml.Node) []keyValue {
	merged := map[string]bool{}
	var out []keyValue
	// walk lists m's entries; blocked holds the keys the mappings above m write
	// themselves, which shadow anything m merges in.
	var walk func(m *yaml.Node, viaMerge bool, blocked map[string]bool)
	walk = func(m *yaml.Node, viaMerge bool, blocked map[string]bool) {
		m = deref(m)
		if m == nil || m.Kind != yaml.MappingNode {
			return
		}
		own := map[string]bool{}
		for k := range blocked {
			own[k] = true
		}
		for i := 0; i+1 < len(m.Content); i += 2 {
			if k := deref(m.Content[i]); !isMergeKey(k) {
				own[k.Value] = true
			}
		}
		for i := 0; i+1 < len(m.Content); i += 2 {
			k, v := deref(m.Content[i]), deref(m.Content[i+1])
			if isMergeKey(k) {
				if v != nil && v.Kind == yaml.SequenceNode {
					for _, src := range v.Content {
						walk(src, true, own)
					}
				} else {
					walk(v, true, own)
				}
				continue
			}
			if viaMerge {
				if blocked[k.Value] || merged[k.Value] {
					continue
				}
				merged[k.Value] = true
			}
			out = append(out, keyValue{key: k, value: v})
		}
	}
	walk(block, false, nil)
	return out
}

// isMergeKey reports whether k is the YAML merge key (<<).
func isMergeKey(k *yaml.Node) bool {
	return k != nil && (k.Tag == "!!merge" || (k.Kind == yaml.ScalarNode && k.Value == "<<" && k.Tag != "!!str"))
}

// deref follows an alias to the node it names.
func deref(n *yaml.Node) *yaml.Node {
	for n != nil && n.Kind == yaml.AliasNode {
		n = n.Alias
	}
	return n
}
