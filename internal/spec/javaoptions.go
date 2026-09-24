package spec

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// MQTLSJavaToolOption is the JAVA_TOOL_OPTIONS flag the connector needs whenever
// an MQ binder uses TLS: it selects the IBM cipher-suite mappings the IBM MQ
// client expects instead of the Oracle names. It is named once, here, because
// every platform's renderer merges it with the operator's own java-options.tool.
const MQTLSJavaToolOption = "-Dcom.ibm.mq.cfg.useIBMCipherMappings=false"

// JavaOptions is the top-level java-options: block of env.yaml -- extra JVM
// options every platform hands the connector through the environment, declared
// once like the image and the timezone.
//
// The two variables are not interchangeable. JAVA_TOOL_OPTIONS (Tool) is read
// by the JVM itself, so it reaches every JVM started in the container, and it
// takes JVM options only. JDK_JAVA_OPTIONS (JDK) is read by the java launcher
// alone (JDK 9+), so it also takes launcher options such as --add-opens, and
// it is applied after JAVA_TOOL_OPTIONS.
//
// Each is one string of whitespace-separated options, written plain or as a
// folded (>-) block when the list runs long; NormalizeJavaOptions makes the
// two spellings identical.
type JavaOptions struct {
	Tool string // -> JAVA_TOOL_OPTIONS
	JDK  string // -> JDK_JAVA_OPTIONS
}

// UnmarshalYAML accepts only a mapping of tool:/jdk: to one scalar each.
// ParseEnv decodes without KnownFields, so without this a misspelled sub-key,
// or a list of options where a string belongs, would be dropped in silence and
// the connector would start without the options the operator asked for. The
// charset is validate's job; parse stays about shape.
func (j *JavaOptions) UnmarshalYAML(node *yaml.Node) error {
	if node.Kind != yaml.MappingNode {
		return fmt.Errorf("java-options must be a mapping with tool: (JAVA_TOOL_OPTIONS) and/or jdk: (JDK_JAVA_OPTIONS), got a %s", yamlKind(node))
	}
	seen := map[string]bool{}
	for i := 0; i+1 < len(node.Content); i += 2 {
		k, v := node.Content[i], node.Content[i+1]
		var dst *string
		switch k.Value {
		case "tool":
			dst = &j.Tool
		case "jdk":
			dst = &j.JDK
		default:
			return fmt.Errorf("java-options.%s is not a key: use tool: for JAVA_TOOL_OPTIONS or jdk: for JDK_JAVA_OPTIONS", k.Value)
		}
		if seen[k.Value] {
			return fmt.Errorf("java-options.%s is set twice; give all its options in one string", k.Value)
		}
		seen[k.Value] = true
		if v.Kind == yaml.AliasNode && v.Alias != nil {
			v = v.Alias
		}
		if v.Kind != yaml.ScalarNode {
			return fmt.Errorf("java-options.%s must be one string of space-separated options (plain, or a >- folded block to spread it over several lines), got a %s", k.Value, yamlKind(v))
		}
		*dst = v.Value
	}
	return nil
}

// yamlKind names a node's shape the way an env.yaml author would.
func yamlKind(n *yaml.Node) string {
	switch n.Kind {
	case yaml.SequenceNode:
		return "list"
	case yaml.MappingNode:
		return "mapping"
	}
	return "scalar"
}

// NormalizeJavaOptions collapses every run of whitespace to one space and trims
// the ends. A folded (>-) block already joins its lines with spaces, but a
// literal (|) block, a more-indented folded line, or a ${VAR} that expanded to
// several lines would not; the JVM splits both variables on whitespace anyway,
// so this changes no option -- it only guarantees the single-line value that a
// quadlet Environment= line and a manifest or compose scalar all need.
func NormalizeJavaOptions(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

// JavaEnv returns the JAVA_TOOL_OPTIONS and JDK_JAVA_OPTIONS values the
// connector container gets, each empty when that variable is not to be set at
// all. It is the one place the tool's MQ TLS flag and the operator's options
// meet, shared by every renderer so the three platforms cannot disagree.
//
// The TLS flag goes first and the operator's options after it: the JVM applies
// the options left to right, so an operator who sets the same property
// deliberately wins.
func JavaEnv(mqTLS bool, j *JavaOptions) (tool, jdk string) {
	var parts []string
	if mqTLS {
		parts = append(parts, MQTLSJavaToolOption)
	}
	if j != nil {
		if t := NormalizeJavaOptions(j.Tool); t != "" {
			parts = append(parts, t)
		}
		jdk = NormalizeJavaOptions(j.JDK)
	}
	return strings.Join(parts, " "), jdk
}
