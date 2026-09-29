package spec

import (
	"strings"

	"gopkg.in/yaml.v3"
)

// TransformHeadersKey is the one place a header transform is read: the top
// level of a workflow file. The block is passed through verbatim to that
// workflow's solace.connector.workflows.<N>.transform-headers, where the
// connector evaluates each transform-headers.expressions.<header> as a SpEL
// expression. It has to live in the workflow file rather than env.yaml because
// the tool numbers the workflows (by sorted file name), so only the file knows
// which <N> it becomes.
const TransformHeadersKey = "transform-headers"

// transformPrefix is what every spelling of a transform key starts with. A key
// beginning with it anywhere other than the one right place is almost certainly
// a transform the author meant to apply -- transform:, transform-header:, or
// transform-headers: nested under a side -- and both files decode without
// KnownFields, so without this scan it would be dropped in silence and the
// connector would start without it.
const transformPrefix = "transform"

// misplacedWorkflowTransforms lists every transform-looking key in a workflow
// file except a top-level transform-headers:, as dotted paths, at each level a
// misplaced one plausibly lands: the top level, each side, the side's
// solace:/mq: block, and that block's consumer:/producer: tuning.
func misplacedWorkflowTransforms(doc *yaml.Node) []string {
	top := documentMapping(doc)
	if top == nil {
		return nil
	}
	var out []string
	out = appendTransformKeys(out, top, "", TransformHeadersKey)
	for _, side := range []string{"source", "target"} {
		s := mappingChild(top, side)
		out = appendTransformKeys(out, s, side+".", "")
		for _, system := range []string{SystemSolace, SystemMQ} {
			b := mappingChild(s, system)
			out = appendTransformKeys(out, b, side+"."+system+".", "")
			for _, tuning := range []string{"consumer", "producer"} {
				out = appendTransformKeys(out, mappingChild(b, tuning), side+"."+system+"."+tuning+".", "")
			}
		}
	}
	return out
}

// misplacedEnvTransforms is misplacedWorkflowTransforms for env.yaml, where no
// transform key belongs at all: the top level, and each connection with its
// solace:/mq: block.
func misplacedEnvTransforms(doc *yaml.Node) []string {
	top := documentMapping(doc)
	if top == nil {
		return nil
	}
	out := appendTransformKeys(nil, top, "", "")
	conns := mappingChild(top, "connections")
	if conns == nil {
		return out
	}
	for i := 0; i+1 < len(conns.Content); i += 2 {
		name, c := conns.Content[i].Value, conns.Content[i+1]
		if c.Kind != yaml.MappingNode {
			continue
		}
		prefix := "connections." + name + "."
		out = appendTransformKeys(out, c, prefix, "")
		for _, system := range []string{SystemSolace, SystemMQ} {
			out = appendTransformKeys(out, mappingChild(c, system), prefix+system+".", "")
		}
	}
	return out
}

// appendTransformKeys appends prefix+key for every key of mapping m that
// starts with transformPrefix, other than allowed. A nil or non-mapping m has
// none.
func appendTransformKeys(out []string, m *yaml.Node, prefix, allowed string) []string {
	if m == nil || m.Kind != yaml.MappingNode {
		return out
	}
	for i := 0; i+1 < len(m.Content); i += 2 {
		if k := m.Content[i].Value; strings.HasPrefix(k, transformPrefix) && k != allowed {
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
// mapping, the key is absent, or its value is not a mapping.
func mappingChild(m *yaml.Node, key string) *yaml.Node {
	if m == nil || m.Kind != yaml.MappingNode {
		return nil
	}
	for i := 0; i+1 < len(m.Content); i += 2 {
		if m.Content[i].Value == key {
			if v := m.Content[i+1]; v.Kind == yaml.MappingNode {
				return v
			}
			return nil
		}
	}
	return nil
}
