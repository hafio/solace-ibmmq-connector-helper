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
)

// yw is an indentation-aware line writer (2 spaces per level).
type yw struct{ b strings.Builder }

func (w *yw) line(indent int, s string) {
	w.b.WriteString(strings.Repeat(" ", indent))
	w.b.WriteString(s)
	w.b.WriteByte('\n')
}

func (w *yw) String() string { return w.b.String() }

// Application renders the full application.yml (with trailing newline).
func Application(m *consolidate.Model) string {
	w := &yw{}
	w.line(0, "spring:")
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
	w.line(2, "ssl:")
	w.line(4, "bundle:")
	w.line(6, "jks:")
	for _, b := range bundles {
		w.line(8, b.Name+":")
		w.line(10, "truststore:")
		w.line(12, "location: "+b.TruststoreLoc)
		w.line(12, "password: "+b.TruststorePwd)
		w.line(12, "type: "+b.TruststoreTyp)
		if b.HasKeystore {
			w.line(10, "keystore:")
			w.line(12, "location: "+b.KeystoreLoc)
			w.line(12, "password: "+b.KeystorePwd)
			w.line(12, "type: "+b.KeystoreTyp)
			w.line(10, "key:")
			w.line(12, "alias: "+b.KeyAlias)
		}
	}
}

func renderCloudStream(w *yw, m *consolidate.Model) {
	w.line(2, "cloud:")
	w.line(4, "stream:")

	// binders
	w.line(6, "binders:")
	for _, b := range m.Binders {
		w.line(8, b.Name+":")
		switch b.Kind {
		case spec.SystemSolace:
			w.line(10, "type: solace")
			w.line(10, "environment:")
			w.line(12, "solace:")
			w.line(14, "java:")
			w.line(16, "host: "+b.Solace.Host)
			w.line(16, "msg-vpn: "+b.Solace.MsgVPN)
			if b.Solace.ClientUser != "" {
				w.line(16, "client-username: "+b.Solace.ClientUser)
			}
			if b.Solace.ClientPass != "" {
				w.line(16, "client-password: "+b.Solace.ClientPass)
			}
			for _, p := range b.Solace.Extras {
				renderProp(w, 16, p)
			}
			if len(b.Solace.APIProps) > 0 {
				w.line(16, "api-properties:")
				for _, p := range b.Solace.APIProps {
					renderProp(w, 18, p)
				}
			}
		case spec.SystemMQ:
			w.line(10, "type: jms")
			w.line(10, "environment:")
			w.line(12, "ibm:")
			w.line(14, "mq:")
			w.line(16, "queue-manager: "+b.MQ.QueueManager)
			w.line(16, "channel: "+b.MQ.Channel)
			w.line(16, "conn-name: "+b.MQ.ConnName)
			if b.MQ.User != "" {
				w.line(16, "user: "+b.MQ.User)
			}
			if b.MQ.Password != "" {
				w.line(16, "password: "+b.MQ.Password)
			}
			if b.MQ.SSLBundle != "" {
				w.line(16, "ssl-bundle: "+b.MQ.SSLBundle)
			}
			if len(b.MQ.AddlProps) > 0 {
				w.line(16, "additional-properties:")
				for _, p := range b.MQ.AddlProps {
					renderProp(w, 18, p)
				}
			}
		case "undefined":
			w.line(10, "type: undefined")
		}
	}

	// bindings
	w.line(6, "bindings:")
	for _, bd := range m.Bindings {
		w.line(8, bd.Name+":")
		w.line(10, "destination: "+bd.Dest)
		w.line(10, "binder: "+bd.Binder)
	}

	// jms bindings
	if len(m.JMSBindings) > 0 {
		w.line(6, "jms:")
		w.line(8, "bindings:")
		for _, jb := range m.JMSBindings {
			w.line(10, jb.Name+":")
			w.line(12, jb.Role+":")
			if jb.DestType != "" {
				w.line(14, "destination-type: "+jb.DestType)
			}
			if jb.Durable != "" {
				w.line(14, "durable-subscription-name: "+jb.Durable)
			}
			for _, p := range jb.Extra {
				renderProp(w, 14, p)
			}
		}
	}

	// solace bindings
	if len(m.SolaceBindings) > 0 {
		w.line(6, "solace:")
		w.line(8, "bindings:")
		for _, sb := range m.SolaceBindings {
			w.line(10, sb.Name+":")
			w.line(12, sb.Role+":")
			if sb.DestType != "" {
				w.line(14, "destination-type: "+sb.DestType)
			}
			for _, p := range sb.Extra {
				renderProp(w, 14, p)
			}
		}
	}
}

func renderConnector(w *yw, m *consolidate.Model) {
	w.line(0, "solace:")
	w.line(2, "connector:")
	w.line(4, "workflows:")
	for _, wf := range m.Workflows {
		w.line(6, strconv.Itoa(wf.ID)+":")
		w.line(8, "enabled: "+strconv.FormatBool(wf.Enabled))
	}
	if m.Security.Present {
		w.line(4, "security:")
		w.line(6, "enabled: "+strconv.FormatBool(m.Security.Enabled))
		if len(m.Security.Users) > 0 {
			w.line(6, "users:")
			for _, u := range m.Security.Users {
				w.line(8, "- name: "+u.Name)
				w.line(10, "password: "+u.Password)
			}
		}
	}
	if m.LeaderElection != nil {
		renderLeaderElection(w, m.LeaderElection)
	}
}

// renderLeaderElection emits solace.connector.management for active_active /
// active_standby, in the order used by example-application.yml.
func renderLeaderElection(w *yw, le *consolidate.LeaderElectionModel) {
	w.line(4, "management:")
	w.line(6, "leader-election:")
	w.line(8, "mode: "+le.Mode)
	if le.FailOver != nil && le.FailOver.Kind == yaml.MappingNode && len(le.FailOver.Content) > 0 {
		w.line(8, "fail-over:")
		renderContainer(w, 10, le.FailOver)
	}
	if le.Queue != "" {
		w.line(6, "queue: "+le.Queue)
	}
	if s := le.Session; s != nil {
		w.line(6, "session:")
		w.line(8, "host: "+s.Host)
		w.line(8, "msg-vpn: "+s.MsgVPN)
		if s.ClientUser != "" {
			w.line(8, "client-username: "+s.ClientUser)
		}
		if s.ClientPass != "" {
			w.line(8, "client-password: "+s.ClientPass)
		}
		if len(s.APIProps) > 0 {
			w.line(8, "api-properties:")
			for _, p := range s.APIProps {
				renderProp(w, 10, p)
			}
		}
	}
}

func renderManagement(w *yw, mg spec.Management) {
	if !mg.Present {
		return
	}
	w.line(0, "management:")
	if mg.Port != 0 {
		w.line(2, "server:")
		w.line(4, "port: "+strconv.Itoa(mg.Port))
	}
	if mg.Exposure != "" {
		w.line(2, "endpoints:")
		w.line(4, "web:")
		w.line(6, "exposure:")
		w.line(8, "include: "+mg.Exposure)
	}
	if mg.HealthShowDetails != "" {
		w.line(2, "endpoint:")
		w.line(4, "health:")
		w.line(6, "show-details: "+mg.HealthShowDetails)
	}
}

func renderLogging(w *yw, level *yaml.Node) {
	if level == nil || level.Kind != yaml.MappingNode || len(level.Content) == 0 {
		return
	}
	w.line(0, "logging:")
	w.line(2, "level:")
	for i := 0; i+1 < len(level.Content); i += 2 {
		w.line(4, level.Content[i].Value+": "+consolidate.FormatScalar(level.Content[i+1]))
	}
}

// renderProp emits one property line, recursing for non-scalar passthrough.
func renderProp(w *yw, indent int, p consolidate.Prop) {
	if p.Sub == nil {
		w.line(indent, p.Key+": "+p.Val)
		return
	}
	w.line(indent, p.Key+":")
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
				w.line(indent, k+": "+consolidate.FormatScalar(v))
			} else {
				w.line(indent, k+":")
				renderContainer(w, indent+2, v)
			}
		}
	case yaml.SequenceNode:
		for _, item := range n.Content {
			if item.Kind == yaml.ScalarNode {
				w.line(indent, "- "+consolidate.FormatScalar(item))
			} else {
				w.line(indent, "-")
				renderContainer(w, indent+2, item)
			}
		}
	}
}
