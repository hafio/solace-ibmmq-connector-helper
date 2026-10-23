package validate

import (
	"strings"
	"testing"

	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/spec"
)

// postureSolace is a clean tcps:// Solace queue side carrying y as its
// api-properties.
func postureSolace(t *testing.T, dest, y string) spec.Side {
	t.Helper()
	s := vSolace(dest, spec.DestQueue, "")
	s.APIProps = transformNode(t, y)
	return s
}

// postureConn is the tcps:// Solace connection vSolace describes, without a
// destination, carrying y as its api-properties.
func postureConn(t *testing.T, y string) spec.Side {
	t.Helper()
	s := vSolace("", "", "")
	s.APIProps = transformNode(t, y)
	return s
}

// postureIssues keeps the findings about certificate-validation posture.
func postureIssues(in []Issue) []string {
	var out []string
	for _, i := range in {
		if strings.Contains(i.Msg, "SSL_VALIDATE_CERTIFICATE") {
			out = append(out, i.String())
		}
	}
	return out
}

const p1Tail = " -- the broker's certificate, host name and expiry are not checked, so anything on the network path can impersonate the broker and read what this connection sends; either trust the broker's CA and remove the entry, or keep it deliberately"

// TestValidatePostureInlineSide pins P1, P1u and P3s on an inline workflow
// side: the tool writes no validation policy, so it reports what the side
// writes for itself, and only on a tcps:// host.
func TestValidatePostureInlineSide(t *testing.T) {
	for _, c := range []struct {
		name  string
		host  string
		props string
		want  []string
	}{
		{"false warns", "", "SSL_VALIDATE_CERTIFICATE: false\n",
			[]string{"x.yaml: source solace: api-properties SSL_VALIDATE_CERTIFICATE: false turns server-certificate validation off for tcps://b:55443" + p1Tail}},
		{"false in any case warns", "", "SSL_VALIDATE_CERTIFICATE: FALSE\n",
			[]string{"x.yaml: source solace: api-properties SSL_VALIDATE_CERTIFICATE: false turns server-certificate validation off for tcps://b:55443" + p1Tail}},
		{"true is the default, silent", "", "SSL_VALIDATE_CERTIFICATE: true\n", nil},
		{"non-literal is unknown", "", "SSL_VALIDATE_CERTIFICATE: no\n",
			[]string{`x.yaml: source solace: api-properties SSL_VALIDATE_CERTIFICATE is "no", not the literal true or false, so whether tcps://b:55443 validates the broker certificate cannot be reported; write true or false`}},
		{"another spelling", "", "ssl_validate_certificate: false\n",
			[]string{`x.yaml: source solace: api-properties "ssl_validate_certificate" is another spelling of SSL_VALIDATE_CERTIFICATE; the Solace API ignores property names it does not recognise, so this entry likely has no effect -- write SSL_VALIDATE_CERTIFICATE`}},
		{"tcp:// host is silent", "tcp://b:55555", "SSL_VALIDATE_CERTIFICATE: false\n", nil},
	} {
		t.Run(c.name, func(t *testing.T) {
			src := postureSolace(t, "Q", c.props)
			if c.host != "" {
				src.Host = c.host
			}
			_, warns := Run(Context{Workflows: []spec.Workflow{wf("x.yaml", src, vMQ("M", spec.DestQueue, false))}, Defaults: &spec.Defaults{}})
			if got := postureIssues(warns); strings.Join(got, "\n") != strings.Join(c.want, "\n") {
				t.Errorf("warns =\n%s\nwant\n%s", strings.Join(got, "\n"), strings.Join(c.want, "\n"))
			}
		})
	}
}

// TestValidatePostureConnectionWarnsOnceWithReach pins that a connection warns
// once, in env.yaml -- never per conn-ref side -- and names everything its
// value reaches: every workflow side on it and the leader-election session.
func TestValidatePostureConnectionWarnsOnceWithReach(t *testing.T) {
	d := &spec.Defaults{
		Connections:    map[string]spec.Side{"edge": postureConn(t, "SSL_VALIDATE_CERTIFICATE: false\n")},
		LeaderElection: spec.LeaderElection{Present: true, Mode: spec.LeaderActiveStby, Queue: "mgmt-q", ConnRef: "edge"},
	}
	ref := func(dest string) spec.Side {
		return spec.Side{System: spec.SystemSolace, ConnRef: "edge", DestKind: spec.DestTopic, Dest: dest}
	}
	wfs := []spec.Workflow{
		wf("a.yaml", vMQ("M1", spec.DestQueue, false), ref("t/1")),
		wf("b.yaml", vMQ("M2", spec.DestQueue, false), ref("t/2")),
	}
	_, warns := Run(Context{Workflows: wfs, Defaults: d})
	want := "env.yaml: connections.edge solace: api-properties SSL_VALIDATE_CERTIFICATE: false turns server-certificate validation off for tcps://b:55443" + p1Tail +
		"; it applies to every workflow side on this connection (2) and to the leader-election session"
	if got := postureIssues(warns); len(got) != 1 || got[0] != want {
		t.Errorf("warns = %q, want exactly [%q]", got, want)
	}
}

// TestValidatePostureBinderConflict pins P2: sides that collapse onto one
// binder share one api-properties block, so two stated values must agree --
// and each side is reported where its value was written.
func TestValidatePostureBinderConflict(t *testing.T) {
	t.Run("inline sides disagree", func(t *testing.T) {
		wfs := []spec.Workflow{
			wf("a.yaml", postureSolace(t, "Q1", "SSL_VALIDATE_CERTIFICATE: false\n"), vMQ("M1", spec.DestQueue, false)),
			wf("b.yaml", postureSolace(t, "Q2", "SSL_VALIDATE_CERTIFICATE: true\n"), vMQ("M2", spec.DestQueue, false)),
		}
		errs, _ := Run(Context{Workflows: wfs, Defaults: &spec.Defaults{}})
		want := "b.yaml: conflicting api-properties SSL_VALIDATE_CERTIFICATE true vs false for the same binder (also in a.yaml): the sides share one connection, so they must agree -- set the same value on every side, or define the connection once under connections: and use conn-ref"
		if got := postureIssues(errs); len(got) != 1 || got[0] != want {
			t.Errorf("errs = %q, want exactly [%q]", got, want)
		}
	})
	t.Run("conn-ref side is reported by its connection", func(t *testing.T) {
		d := &spec.Defaults{Connections: map[string]spec.Side{"edge": postureConn(t, "SSL_VALIDATE_CERTIFICATE: false\n")}}
		wfs := []spec.Workflow{
			wf("a.yaml", vMQ("M1", spec.DestQueue, false), spec.Side{System: spec.SystemSolace, ConnRef: "edge", DestKind: spec.DestTopic, Dest: "t/1"}),
			wf("b.yaml", postureSolace(t, "Q2", "SSL_VALIDATE_CERTIFICATE: 'false'\n"), vMQ("M2", spec.DestQueue, false)),
		}
		errs, _ := Run(Context{Workflows: wfs, Defaults: d})
		want := "b.yaml: conflicting api-properties SSL_VALIDATE_CERTIFICATE 'false' vs false for the same binder (also in env.yaml connections.edge): the sides share one connection, so they must agree -- set the same value on every side, or define the connection once under connections: and use conn-ref"
		if got := postureIssues(errs); len(got) != 1 || got[0] != want {
			t.Errorf("errs = %q, want exactly [%q]", got, want)
		}
	})
	t.Run("stated beside absent is not a conflict", func(t *testing.T) {
		wfs := []spec.Workflow{
			wf("a.yaml", postureSolace(t, "Q1", "SSL_VALIDATE_CERTIFICATE: false\n"), vMQ("M1", spec.DestQueue, false)),
			wf("b.yaml", vSolace("Q2", spec.DestQueue, ""), vMQ("M2", spec.DestQueue, false)),
		}
		errs, warns := Run(Context{Workflows: wfs, Defaults: &spec.Defaults{}})
		if got := postureIssues(errs); len(got) != 0 {
			t.Errorf("errs = %q, want none", got)
		}
		// The stated false applies to the whole binder, which P1 says.
		if got := postureIssues(warns); len(got) != 1 || !strings.HasSuffix(got[0], "; it applies to every workflow side on this connection (2)") {
			t.Errorf("warns = %q, want one P1 naming both sides", got)
		}
	})
}

// TestValidatePostureSession pins that an inline leader-election session is a
// site only when it is rendered (active_* modes), and P6: a session and the
// workflow sides on its broker tuple that disagree are flagged.
func TestValidatePostureSession(t *testing.T) {
	sess := func(y string) *spec.Side {
		s := postureConn(t, y)
		return &s
	}
	t.Run("standalone session is not a site", func(t *testing.T) {
		d := &spec.Defaults{LeaderElection: spec.LeaderElection{Present: true, Mode: spec.LeaderStandalone, Session: sess("SSL_VALIDATE_CERTIFICATE: false\n")}}
		_, warns := Run(Context{Workflows: []spec.Workflow{wf("x.yaml", vSolace("Q", spec.DestQueue, ""), vMQ("M", spec.DestQueue, false))}, Defaults: d})
		if got := postureIssues(warns); len(got) != 0 {
			t.Errorf("warns = %q, want none", got)
		}
	})
	t.Run("session off, sides validate", func(t *testing.T) {
		d := &spec.Defaults{LeaderElection: spec.LeaderElection{Present: true, Mode: spec.LeaderActiveStby, Queue: "mgmt-q", Session: sess("SSL_VALIDATE_CERTIFICATE: false\n")}}
		_, warns := Run(Context{Workflows: []spec.Workflow{wf("x.yaml", vSolace("Q", spec.DestQueue, ""), vMQ("M", spec.DestQueue, false))}, Defaults: d})
		want := []string{
			"env.yaml: leader-election session solace: api-properties SSL_VALIDATE_CERTIFICATE: false turns server-certificate validation off for tcps://b:55443" + p1Tail,
			"env.yaml: the leader-election session does not validate the broker certificate, but the workflow sides on the same broker connection (x.yaml) validate it; the session is configured separately -- set SSL_VALIDATE_CERTIFICATE the same way in the session's api-properties, or use leader-election conn-ref so it follows the connection",
		}
		if got := postureIssues(warns); strings.Join(got, "\n") != strings.Join(want, "\n") {
			t.Errorf("warns =\n%s\nwant\n%s", strings.Join(got, "\n"), strings.Join(want, "\n"))
		}
	})
	t.Run("sides off, session validates", func(t *testing.T) {
		s := postureConn(t, "REAPPLY_SUBSCRIPTIONS: true\n")
		d := &spec.Defaults{LeaderElection: spec.LeaderElection{Present: true, Mode: spec.LeaderActiveActive, Queue: "mgmt-q", Session: &s}}
		_, warns := Run(Context{Workflows: []spec.Workflow{wf("x.yaml", postureSolace(t, "Q", "SSL_VALIDATE_CERTIFICATE: false\n"), vMQ("M", spec.DestQueue, false))}, Defaults: d})
		want := "env.yaml: the leader-election session validates the broker certificate, but the workflow sides on the same broker connection (x.yaml) do not validate it; the session is configured separately -- set SSL_VALIDATE_CERTIFICATE the same way in the session's api-properties, or use leader-election conn-ref so it follows the connection"
		got := postureIssues(warns)
		if len(got) != 2 || got[1] != want {
			t.Errorf("warns = %q, want P1 for x.yaml then %q", got, want)
		}
	})
}
