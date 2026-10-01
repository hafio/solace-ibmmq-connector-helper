package validate

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/spec"
)

// The content types Solace documents for the IBM MQ connector's transform
// payload blocks. application/xml is on Solace's generic message-transform
// page, but the IBM MQ connector's release notes support JSON payloads only.
const (
	contentTypeJSON        = "application/json"
	contentTypeUnspecified = "application/vnd.solace.micro-integration.unspecified"
	contentTypeXML         = "application/xml"
)

// checkTransforms validates every workflow's transforms: the transform: block,
// the deprecated transform-headers: block, and every transform-looking key
// found anywhere a transform is not read (spec.Workflow.MisplacedTransforms,
// spec.Defaults.MisplacedTransforms). Both files decode without KnownFields,
// so a transform-header: typo or a transform: nested under a side would
// otherwise vanish, and the connector would start and run without it.
//
// Both blocks are passed through verbatim, so only their shape is checked --
// the SpEL is the connector's to evaluate. A key this tool does not know is a
// warning rather than an error: it is passed through as written, and nothing
// here can know every key a later connector release reads.
//
// lint adds the one finding generate and deploy stay quiet about: that Solace
// has deprecated transform-headers. The connector still reads it, so a file
// that uses it keeps deploying; validate is where it is asked to move.
func checkTransforms(add, warn func(string, string, ...any), wfs []spec.Workflow, d *spec.Defaults, lint bool) {
	for _, p := range d.MisplacedTransforms {
		add(fileEnv, "%s is not read here: a transform belongs to one workflow, so write transform: (or the deprecated transform-headers:) at the top level of each workflow file it applies to", p)
	}
	for _, wf := range wfs {
		for _, p := range wf.MisplacedTransforms {
			checkMisplacedTransform(add, wf.File, p)
		}
		if wf.Transform != nil && wf.TransformHeaders != nil {
			add(wf.File, `transform and transform-headers are both set, and the connector cannot use the two together ("The previous and current configuration sections cannot be used together."): keep transform: and move each transform-headers.expressions.<header> into transform.expressions as - transform: "target['headers']['<header>'] = <SpEL expression>" (user guide section 6.7), or drop transform: to stay on the deprecated transform-headers: for now`)
		}
		if wf.Transform != nil {
			checkTransform(add, warn, wf.File, wf.Transform)
		}
		if wf.TransformHeaders != nil {
			checkTransformHeaders(add, warn, wf.File, wf.TransformHeaders)
			if lint {
				warn(wf.File, "transform-headers is deprecated by Solace (since connector 2.9.0; connector 3.1.0 still reads it, a future release removes it): move to transform:, which carries no source header over to the target unless an expression sets it (user guide section 6.7)")
			}
		}
	}
}

// checkTransformHeaders checks the shape of one transform-headers: block -- the
// deprecated section: a mapping whose expressions: maps each header name to
// one SpEL expression.
func checkTransformHeaders(add, warn func(string, string, ...any), file string, n *yaml.Node) {
	const key = spec.TransformHeadersKey
	if n.Kind != yaml.MappingNode {
		add(file, "%s must be a mapping with expressions: <header>: <SpEL expression>, got a %s", key, spec.YAMLKind(n))
		return
	}
	var exprs *yaml.Node
	for i := 0; i+1 < len(n.Content); i += 2 {
		k, v := n.Content[i].Value, n.Content[i+1]
		if k == "expressions" {
			exprs = v
			continue
		}
		warn(file, "%s.%s is not a key this tool knows: it is passed through as written, but the connector reads header transforms from %s.expressions -- check the spelling", key, k, key)
	}
	switch {
	case exprs == nil:
		add(file, "%s has no expressions: -- the connector reads header transforms only from %s.expressions (<header>: <SpEL expression>)", key, key)
		return
	case exprs.Kind != yaml.MappingNode:
		add(file, "%s.expressions must be a mapping of <header>: <SpEL expression>, got a %s", key, spec.YAMLKind(exprs))
		return
	case len(exprs.Content) == 0:
		warn(file, "%s.expressions is empty, so this workflow transforms no headers", key)
		return
	}
	seen := map[string]bool{}
	for i := 0; i+1 < len(exprs.Content); i += 2 {
		h, v := exprs.Content[i].Value, exprs.Content[i+1]
		if seen[h] {
			add(file, "%s.expressions sets %q twice; give each header one expression", key, h)
		}
		seen[h] = true
		if v.Kind != yaml.ScalarNode {
			add(file, "%s.expressions.%s must be one SpEL expression, got a %s", key, h, spec.YAMLKind(v))
		}
	}
}

// checkMisplacedTransform reports one transform-looking key found where no
// transform is read, at dotted path p, with the advice that fits it: a real
// block in the wrong place is told where it goes, a legacy payload section how
// payload transforms are written now, and anything else that it is not a key.
func checkMisplacedTransform(add func(string, string, ...any), file, p string) {
	switch p[strings.LastIndex(p, ".")+1:] {
	case spec.TransformKey:
		add(file, "%s is in the wrong place: a transform applies to the whole workflow, so transform: goes at the top level of this file, beside source: and target:", p)
	case spec.TransformHeadersKey:
		add(file, "%s is in the wrong place: header transforms apply to the whole workflow, so transform-headers: goes at the top level of this file, beside source: and target:", p)
	case spec.TransformPayloadKey, spec.TransformPayloadsKey:
		add(file, "%s is not read by this tool: payload transforms are written as target['payload'] expressions in a transform: block at the top level of this file, not as a transform-payload section", p)
	default:
		add(file, "%s is not a key: a workflow file's transforms are written transform: (current) or transform-headers: (deprecated) at the top level of this file", p)
	}
}

// checkTransform checks the shape of one transform: block, the connector's
// current transform section: optional source-payload: and target-payload:,
// each naming the content-type that payload is read as, and expressions: as an
// ordered list of - transform: <SpEL expression> items.
//
// enabled is the one key Solace documents inconsistently -- its IBM MQ
// migration guide adds enabled: true, its property table does not list it --
// so until that is settled it is passed through as written and not checked.
func checkTransform(add, warn func(string, string, ...any), file string, n *yaml.Node) {
	const key = spec.TransformKey
	if n.Kind != yaml.MappingNode {
		add(file, "%s must be a mapping (source-payload:, target-payload:, expressions:), got a %s", key, spec.YAMLKind(n))
		return
	}
	var exprs *yaml.Node
	contentType := false
	for _, kv := range mappingEntries(add, file, key, n) {
		switch kv.key {
		case "expressions":
			exprs = kv.value
		case "source-payload", "target-payload":
			if checkPayloadBlock(add, warn, file, key+"."+kv.key, kv.value) {
				contentType = true
			}
		case "enabled":
			// Passed through unchecked; see the doc comment above.
		default:
			warn(file, "%s.%s is not a key this tool knows: it is passed through as written, but the connector reads a transform from source-payload, target-payload and expressions -- check the spelling", key, kv.key)
		}
	}
	if isNull(exprs) {
		if !contentType {
			warn(file, "%s sets no expressions and no payload content-type, so it does nothing", key)
		}
		return
	}
	checkTransformExpressions(add, warn, file, exprs)
}

// checkPayloadBlock checks transform.source-payload or transform.target-payload
// (path): a mapping whose content-type names how the connector reads that
// payload. It reports whether a content-type was set.
func checkPayloadBlock(add, warn func(string, string, ...any), file, path string, n *yaml.Node) bool {
	if isNull(n) {
		return false
	}
	if n.Kind != yaml.MappingNode {
		add(file, "%s must be a mapping holding content-type:, got a %s", path, spec.YAMLKind(n))
		return false
	}
	set := false
	for _, kv := range mappingEntries(add, file, path, n) {
		switch {
		case kv.key != "content-type":
			warn(file, "%s.%s is not a key this tool knows: it is passed through as written, but the connector reads only content-type here -- check the spelling", path, kv.key)
		case isNull(kv.value):
			// No value: the connector's default content type applies.
		case kv.value.Kind != yaml.ScalarNode:
			add(file, "%s.content-type must be one value such as application/json, got a %s", path, spec.YAMLKind(kv.value))
		default:
			set = true
			checkContentType(warn, file, path, kv.value.Value)
		}
	}
	return set
}

// checkContentType warns about a content-type the IBM MQ connector is not
// documented to read. A warning, not an error: the value is passed through as
// written, and a later connector release may read more. A ${...} placeholder
// is Spring's to resolve, so it is not second-guessed.
func checkContentType(warn func(string, string, ...any), file, path, ct string) {
	if ct == contentTypeJSON || ct == contentTypeUnspecified || strings.Contains(ct, "${") {
		return
	}
	note := ""
	if ct == contentTypeXML {
		note = " -- the IBM MQ connector's release notes support JSON payloads only"
	}
	warn(file, "%s.content-type %q is not a value documented for the IBM MQ connector (%s, %s); it is passed through as written%s", path, ct, contentTypeJSON, contentTypeUnspecified, note)
}

// checkTransformExpressions checks transform.expressions: an ordered list
// whose every item is a mapping holding one SpEL expression under transform.
// A mapping here is the transform-headers shape carried over unchanged, so it
// gets the migration hint rather than a bare type error.
func checkTransformExpressions(add, warn func(string, string, ...any), file string, n *yaml.Node) {
	const path = spec.TransformKey + ".expressions"
	switch n.Kind {
	case yaml.SequenceNode:
	case yaml.MappingNode:
		add(file, `%s must be a list of - transform: <SpEL expression> items, got a mapping -- that is the transform-headers.expressions shape; write each <header>: <expression> as - transform: "target['headers']['<header>'] = <expression>" (user guide section 6.7)`, path)
		return
	default:
		add(file, "%s must be a list of - transform: <SpEL expression> items, got a %s", path, spec.YAMLKind(n))
		return
	}
	if len(n.Content) == 0 {
		warn(file, "%s is empty, so this workflow applies no transform expressions", path)
		return
	}
	for i, item := range n.Content {
		checkTransformItem(add, warn, file, fmt.Sprintf("%s[%d]", path, i), item)
	}
}

// checkTransformItem checks one transform.expressions item (path): a mapping
// holding one SpEL expression under transform.
func checkTransformItem(add, warn func(string, string, ...any), file, path string, item *yaml.Node) {
	if item.Kind != yaml.MappingNode {
		add(file, "%s must be a mapping holding transform: <SpEL expression>, got a %s", path, spec.YAMLKind(item))
		return
	}
	var expr *yaml.Node
	for _, kv := range mappingEntries(add, file, path, item) {
		if kv.key == spec.TransformKey {
			expr = kv.value
			continue
		}
		warn(file, "%s.%s is not a key this tool knows: it is passed through as written, but the connector reads one SpEL expression per item from transform -- check the spelling", path, kv.key)
	}
	switch {
	case expr == nil:
		add(file, "%s has no transform: -- each item is - transform: <SpEL expression>", path)
	case expr.Kind != yaml.ScalarNode:
		add(file, "%s.transform must be one SpEL expression, got a %s", path, spec.YAMLKind(expr))
	case isNull(expr) || strings.TrimSpace(expr.Value) == "":
		add(file, "%s.transform must be one SpEL expression, got an empty value", path)
	case expr.Style&(yaml.DoubleQuotedStyle|yaml.SingleQuotedStyle|yaml.LiteralStyle|yaml.FoldedStyle) == 0 && strings.HasSuffix(strings.TrimSpace(expr.Value), "="):
		// Unquoted, YAML ends the value at its first " #" -- which every
		// #function(...) call has -- and no SpEL expression ends in "=".
		add(file, `%s.transform ends at "=": unquoted, the expression stops at its first " #", which YAML reads as the start of a comment -- quote the whole expression`, path)
	}
}

// keyedNode is one key and its value from a YAML mapping.
type keyedNode struct {
	key   string
	value *yaml.Node
}

// mappingEntries returns m's entries in order, reporting any key set twice at
// path: a block captured as a node keeps both, and only one of them would
// reach the connector.
func mappingEntries(add func(string, string, ...any), file, path string, m *yaml.Node) []keyedNode {
	var out []keyedNode
	seen := map[string]bool{}
	for i := 0; i+1 < len(m.Content); i += 2 {
		k := m.Content[i].Value
		if seen[k] {
			add(file, "%s sets %q twice; keep one", path, k)
		}
		seen[k] = true
		out = append(out, keyedNode{key: k, value: m.Content[i+1]})
	}
	return out
}

// isNull reports whether n is absent or an explicit YAML null -- a key written
// with no value, which the connector reads the same as no key at all.
func isNull(n *yaml.Node) bool {
	return n == nil || (n.Kind == yaml.ScalarNode && n.ShortTag() == "!!null")
}
