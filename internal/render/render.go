// Package render emits the consolidated application.yml.
//
// The connector's application.yml has an intentional, non-alphabetical key
// order (e.g. spring.ssl before spring.cloud; a fixed api-properties order) that
// generic YAML marshaling cannot reproduce. Rendering therefore walks the
// ordered Model with a small indentation writer so output is byte-for-byte
// stable and diffs cleanly on regeneration.
package render

import (
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/consolidate"
	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/spec"
	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/yamlwriter"
)

// yw is the indentation-aware line writer shared by every renderer here.
type yw = yamlwriter.Writer

// q renders a user-supplied value as a YAML scalar, quoting only when a plain
// one would not read back unchanged. Every value that originates in the spec
// (hosts, credentials, destinations, store paths) goes through it; generated
// identifiers (binder/binding names, bundle names, enum-valued modes) do not,
// because their charset already makes them safe.
func q(s string) string { return consolidate.QuoteScalar(s) }

// writeScalar emits "key: value" -- or a block scalar when the value spans more
// than one line, since a plain "key: line1\nline2" would put line2 at the
// document's top level. The indicator preserves the trailing newline the value
// actually has (| keeps one, |- keeps none), so a round-tripped literal block
// reads back byte-for-byte.
func writeScalar(w *yw, indent int, key, val string) {
	if !strings.Contains(val, "\n") {
		w.Line(indent, key+": "+val)
		return
	}
	indicator, body := blockIndicator(val)
	w.Line(indent, key+": "+indicator)
	writeBlockBody(w, indent+2, body)
}

// blockIndicator picks the chomping indicator that preserves the value's own
// trailing newline (| keeps one, |- keeps none) and returns the body to emit.
func blockIndicator(val string) (indicator, body string) {
	if strings.HasSuffix(val, "\n") {
		return "|", strings.TrimSuffix(val, "\n")
	}
	return "|-", val
}

// writeBlockBody emits the lines of a block scalar. An empty line is written
// bare so the block carries no trailing whitespace.
func writeBlockBody(w *yw, indent int, body string) {
	for _, ln := range strings.Split(body, "\n") {
		if ln == "" {
			w.Raw("\n")
			continue
		}
		w.Line(indent, ln)
	}
}

// Application renders the full application.yml (with trailing newline).
func Application(m *consolidate.Model) string {
	w := &yw{}
	w.Line(0, "spring:")
	// Credentials are mounted as files, one per stable secret name, and read back
	// as properties from there -- so every ${...} the config below references
	// resolves without a single credential appearing in this document. `optional:`
	// keeps it inert wherever nothing is mounted.
	if m.ConfigImport != "" {
		w.Line(2, "config:")
		w.Line(4, "import: "+q(m.ConfigImport))
	}
	if len(m.Bundles) > 0 {
		renderBundles(w, m.Bundles)
	}
	renderCloudStream(w, m)
	renderConnector(w, m)
	renderManagement(w, m.Management)
	renderLogging(w, m.LoggingLevel)
	return w.String()
}

func renderBundles(w *yw, bundles []*consolidate.Bundle) {
	w.Line(2, "ssl:")
	w.Line(4, "bundle:")
	w.Line(6, "jks:")
	for _, b := range bundles {
		w.Line(8, b.Name+":")
		w.Line(10, "truststore:")
		w.Line(12, "location: "+q(b.TruststoreLoc))
		w.Line(12, "password: "+q(b.TruststorePwd))
		w.Line(12, "type: "+q(b.TruststoreTyp))
		if b.HasKeystore {
			w.Line(10, "keystore:")
			w.Line(12, "location: "+q(b.KeystoreLoc))
			w.Line(12, "password: "+q(b.KeystorePwd))
			w.Line(12, "type: "+q(b.KeystoreTyp))
			w.Line(10, "key:")
			w.Line(12, "alias: "+q(b.KeyAlias))
		}
	}
}

func renderCloudStream(w *yw, m *consolidate.Model) {
	w.Line(2, "cloud:")
	w.Line(4, "stream:")

	// binders
	w.Line(6, "binders:")
	for _, b := range m.Binders {
		w.Line(8, b.Name+":")
		switch b.Kind {
		case spec.SystemSolace:
			w.Line(10, "type: solace")
			w.Line(10, "environment:")
			w.Line(12, "solace:")
			w.Line(14, "java:")
			w.Line(16, "host: "+q(b.Solace.Host))
			w.Line(16, "msg-vpn: "+q(b.Solace.MsgVPN))
			if b.Solace.ClientUser != "" {
				w.Line(16, "client-username: "+q(b.Solace.ClientUser))
			}
			if b.Solace.ClientPass != "" {
				w.Line(16, "client-password: "+q(b.Solace.ClientPass))
			}
			for _, p := range b.Solace.Extras {
				renderProp(w, 16, p)
			}
			if len(b.Solace.APIProps) > 0 {
				w.Line(16, "api-properties:")
				for _, p := range b.Solace.APIProps {
					renderProp(w, 18, p)
				}
			}
		case spec.SystemMQ:
			w.Line(10, "type: jms")
			w.Line(10, "environment:")
			w.Line(12, "ibm:")
			w.Line(14, "mq:")
			w.Line(16, "queue-manager: "+q(b.MQ.QueueManager))
			w.Line(16, "channel: "+q(b.MQ.Channel))
			w.Line(16, "conn-name: "+q(b.MQ.ConnName))
			if b.MQ.User != "" {
				w.Line(16, "user: "+q(b.MQ.User))
			}
			if b.MQ.Password != "" {
				w.Line(16, "password: "+q(b.MQ.Password))
			}
			if b.MQ.SSLBundle != "" {
				w.Line(16, "ssl-bundle: "+b.MQ.SSLBundle)
			}
			if len(b.MQ.AddlProps) > 0 {
				w.Line(16, "additional-properties:")
				for _, p := range b.MQ.AddlProps {
					renderProp(w, 18, p)
				}
			}
		case "undefined":
			w.Line(10, "type: undefined")
		}
	}

	// bindings
	w.Line(6, "bindings:")
	for _, bd := range m.Bindings {
		w.Line(8, bd.Name+":")
		w.Line(10, "destination: "+q(bd.Dest))
		w.Line(10, "binder: "+bd.Binder)
	}

	// jms bindings
	if len(m.JMSBindings) > 0 {
		w.Line(6, "jms:")
		w.Line(8, "bindings:")
		for _, jb := range m.JMSBindings {
			w.Line(10, jb.Name+":")
			w.Line(12, jb.Role+":")
			if jb.DestType != "" {
				w.Line(14, "destination-type: "+jb.DestType)
			}
			if jb.Durable != "" {
				w.Line(14, "durable-subscription-name: "+jb.Durable)
			}
			for _, p := range jb.Extra {
				renderProp(w, 14, p)
			}
		}
	}

	// solace bindings
	if len(m.SolaceBindings) > 0 {
		w.Line(6, "solace:")
		w.Line(8, "bindings:")
		for _, sb := range m.SolaceBindings {
			w.Line(10, sb.Name+":")
			w.Line(12, sb.Role+":")
			if sb.DestType != "" {
				w.Line(14, "destination-type: "+sb.DestType)
			}
			for _, p := range sb.Extra {
				renderProp(w, 14, p)
			}
		}
	}
}

func renderConnector(w *yw, m *consolidate.Model) {
	w.Line(0, "solace:")
	w.Line(2, "connector:")
	w.Line(4, "workflows:")
	for _, wf := range m.Workflows {
		w.Line(6, strconv.Itoa(wf.ID)+":")
		w.Line(8, "enabled: "+strconv.FormatBool(wf.Enabled))
	}
	// Management security is always on now (see spec.StatusUserName): the
	// block is unconditional and enabled is always true, so there is no
	// operator toggle left to gate on.
	w.Line(4, "security:")
	w.Line(6, "enabled: true")
	if len(m.Security.Users) > 0 {
		w.Line(6, "users:")
		for _, u := range m.Security.Users {
			w.Line(8, "- name: "+q(u.Name))
			w.Line(10, "password: "+q(u.Password))
			// Omitted entirely when empty: an empty list is the connector's
			// read-only default, so a role-less user (every user before roles
			// were supported, and the reserved status account) renders exactly
			// as it did before.
			if len(u.Roles) > 0 {
				w.Line(10, "roles:")
				for _, r := range u.Roles {
					w.Line(12, "- "+q(r))
				}
			}
		}
	}
	if m.LeaderElection != nil {
		renderLeaderElection(w, m.LeaderElection)
	}
}

// renderLeaderElection emits solace.connector.management for active_active /
// active_standby, in the order used by testdata/golden/application.yml.
func renderLeaderElection(w *yw, le *consolidate.LeaderElectionModel) {
	w.Line(4, "management:")
	w.Line(6, "leader-election:")
	w.Line(8, "mode: "+le.Mode)
	if le.FailOver != nil && le.FailOver.Kind == yaml.MappingNode && len(le.FailOver.Content) > 0 {
		w.Line(8, "fail-over:")
		renderContainer(w, 10, le.FailOver)
	}
	if le.Queue != "" {
		w.Line(6, "queue: "+q(le.Queue))
	}
	if s := le.Session; s != nil {
		w.Line(6, "session:")
		w.Line(8, "host: "+q(s.Host))
		w.Line(8, "msg-vpn: "+q(s.MsgVPN))
		if s.ClientUser != "" {
			w.Line(8, "client-username: "+q(s.ClientUser))
		}
		if s.ClientPass != "" {
			w.Line(8, "client-password: "+q(s.ClientPass))
		}
		for _, p := range s.Extras {
			renderProp(w, 8, p)
		}
		if len(s.APIProps) > 0 {
			w.Line(8, "api-properties:")
			for _, p := range s.APIProps {
				renderProp(w, 10, p)
			}
		}
	}
}

// renderManagement always emits the management block: server.port and the
// fixed actuator exposure list are unconditional (consolidate.applyStatusAccess
// guarantees both are always set), endpoint.health.show-details only when the
// operator configured one.
func renderManagement(w *yw, mg consolidate.Management) {
	w.Line(0, "management:")
	w.Line(2, "server:")
	w.Line(4, "port: "+strconv.Itoa(mg.Port))
	w.Line(2, "endpoints:")
	w.Line(4, "web:")
	w.Line(6, "exposure:")
	w.Line(8, "include: "+mg.Exposure)
	if mg.HealthShowDetails != "" {
		w.Line(2, "endpoint:")
		w.Line(4, "health:")
		w.Line(6, "show-details: "+q(mg.HealthShowDetails))
	}
}

func renderLogging(w *yw, level *yaml.Node) {
	if level == nil || level.Kind != yaml.MappingNode || len(level.Content) == 0 {
		return
	}
	w.Line(0, "logging:")
	w.Line(2, "level:")
	for i := 0; i+1 < len(level.Content); i += 2 {
		writeScalar(w, 4, level.Content[i].Value, consolidate.FormatScalar(level.Content[i+1]))
	}
}

// renderProp emits one property line, recursing for non-scalar passthrough.
func renderProp(w *yw, indent int, p consolidate.Prop) {
	if p.Sub == nil {
		writeScalar(w, indent, p.Key, p.Val)
		return
	}
	w.Line(indent, p.Key+":")
	renderContainer(w, indent+2, p.Sub)
}

// renderContainer emits the children of a mapping or sequence node.
func renderContainer(w *yw, indent int, n *yaml.Node) {
	switch n.Kind {
	case yaml.MappingNode:
		for i := 0; i+1 < len(n.Content); i += 2 {
			k := n.Content[i].Value
			v := n.Content[i+1]
			if v.Kind == yaml.ScalarNode {
				writeScalar(w, indent, k, consolidate.FormatScalar(v))
			} else {
				w.Line(indent, k+":")
				renderContainer(w, indent+2, v)
			}
		}
	case yaml.SequenceNode:
		for _, item := range n.Content {
			if item.Kind == yaml.ScalarNode {
				writeSeqItem(w, indent, consolidate.FormatScalar(item))
			} else {
				w.Line(indent, "-")
				renderContainer(w, indent+2, item)
			}
		}
	}
}

// writeSeqItem emits "- value", putting the block indicator on the dash line
// when the value spans several lines.
func writeSeqItem(w *yw, indent int, val string) {
	if !strings.Contains(val, "\n") {
		w.Line(indent, "- "+val)
		return
	}
	indicator, body := blockIndicator(val)
	w.Line(indent, "- "+indicator)
	writeBlockBody(w, indent+2, body)
}
