package spec

import (
	"slices"
	"strings"

	"gopkg.in/yaml.v3"
)

// TransformKey is where a workflow's transforms are read: the top level of a
// workflow file. The block -- source-payload:/target-payload: content types and
// an expressions: list of SpEL expressions -- is passed through verbatim to that
// workflow's solace.connector.workflows.<N>.transform, the connector's current
// transform section. It has to live in the workflow file rather than env.yaml
// because the tool numbers the workflows (by sorted file name), so only the
// file knows which <N> it becomes.
const TransformKey = "transform"

// TransformHeadersKey is the connector's earlier, header-only transform
// section, deprecated by Solace since connector 2.9.0 but still read. It is
// read in the same place as TransformKey and passed through the same way, to
// solace.connector.workflows.<N>.transform-headers, where the connector
// evaluates each transform-headers.expressions.<header> as a SpEL expression.
// The connector cannot use the two sections together.
const TransformHeadersKey = "transform-headers"

// The connector's legacy payload transform section, spelled both ways in
// Solace's documentation. The tool has never carried it -- payload transforms
// are written in a transform: block -- so either spelling is always misplaced.
const (
	TransformPayloadKey  = "transform-payload"
	TransformPayloadsKey = "transform-payloads"
)

// transformPrefix is what every spelling of a transform key starts with. A key
// beginning with it anywhere other than the two right places is almost
// certainly a transform the author meant to apply -- transform-header:,
// transform-payload:, or a transform block nested under a side -- and both
// files decode without KnownFields, so without this scan it would be dropped in
// silence and the connector would start without it.
const transformPrefix = "transform"

// misplacedWorkflowTransforms lists every transform-looking key in a workflow
// file except a top-level transform: or transform-headers:, as dotted paths, at
// each level a misplaced one plausibly lands: the top level, each side, the
// side's solace:/mq: block, and that block's consumer:/producer: tuning.
func misplacedWorkflowTransforms(doc *yaml.Node) []string {
	top := documentMapping(doc)
	if top == nil {
		return nil
	}
	var out []string
	out = appendTransformKeys(out, top, "", TransformKey, TransformHeadersKey)
	for _, side := range []string{"source", "target"} {
		s := mappingChild(top, side)
		out = appendTransformKeys(out, s, side+".")
		for _, system := range []string{SystemSolace, SystemMQ} {
			b := mappingChild(s, system)
			out = appendTransformKeys(out, b, side+"."+system+".")
			for _, tuning := range []string{"consumer", "producer"} {
				out = appendTransformKeys(out, mappingChild(b, tuning), side+"."+system+"."+tuning+".")
			}
		}
	}
	return out
}

// misplacedEnvTransforms is misplacedWorkflowTransforms for env.yaml, where no
// transform key belongs at all: the top level, each connection with its
// solace:/mq: block, and the inline leader-election session.
func misplacedEnvTransforms(doc *yaml.Node) []string {
	top := documentMapping(doc)
	if top == nil {
		return nil
	}
	out := appendTransformKeys(nil, top, "")
	if conns := mappingChild(top, "connections"); conns != nil {
		for i := 0; i+1 < len(conns.Content); i += 2 {
			name, c := conns.Content[i].Value, deref(conns.Content[i+1])
			if c == nil || c.Kind != yaml.MappingNode {
				continue
			}
			prefix := "connections." + name + "."
			out = appendTransformKeys(out, c, prefix)
			for _, system := range []string{SystemSolace, SystemMQ} {
				out = appendTransformKeys(out, mappingChild(c, system), prefix+system+".")
			}
		}
	}
	// The inline management session is a connection block too, read through the
	// same struct as a side, so a transform under it would vanish the same way.
	return appendTransformKeys(out, mappingChild(mappingChild(top, "leader-election"), "session"), "leader-election.session.")
}

// appendTransformKeys appends prefix+key for every key of mapping m that
// starts with transformPrefix, other than the allowed ones -- a key merged in
// with << or reached through an alias included, since the typed decode reads
// those the same way. A nil or non-mapping m has none.
func appendTransformKeys(out []string, m *yaml.Node, prefix string, allowed ...string) []string {
	m = deref(m)
	if m == nil || m.Kind != yaml.MappingNode {
		return out
	}
	for _, kv := range flattenBlock(m) {
		if k := kv.key.Value; strings.HasPrefix(k, transformPrefix) && !slices.Contains(allowed, k) {
			out = append(out, prefix+k)
		}
	}
	return out
}

// documentMapping returns the top-level mapping of a decoded document, or nil
// when the document is empty or is not a mapping.
func documentMapping(doc *yaml.Node) *yaml.Node {
	if doc == nil || doc.Kind != yaml.DocumentNode || len(doc.Content) == 0 {
		return nil
	}
	if top := doc.Content[0]; top.Kind == yaml.MappingNode {
		return top
	}
	return nil
}

// mappingChild returns the mapping value of key in m, or nil when m is not a
// mapping, the key is absent, or its value is not a mapping. An alias is
// followed and a << merge expanded, as the typed decode does.
func mappingChild(m *yaml.Node, key string) *yaml.Node {
	m = deref(m)
	if m == nil || m.Kind != yaml.MappingNode {
		return nil
	}
	for _, kv := range flattenBlock(m) {
		if kv.key.Value == key {
			if kv.value != nil && kv.value.Kind == yaml.MappingNode {
				return kv.value
			}
			return nil
		}
	}
	return nil
}
