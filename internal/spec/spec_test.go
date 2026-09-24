package spec

import (
	"slices"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestParseWorkflowSolaceAndMQ(t *testing.T) {
	data := []byte(`
enabled: false
source:
  solace:
    host: tcps://b:55443
    msg-vpn: prod
    client-username-env: EDGE_SOLACE_USER
    client-password-env: EDGE_SOLACE_PASS
    key-alias: sc
    queue: Q.IN
    api-properties:
      REAPPLY_SUBSCRIPTIONS: true
target:
  mq:
    conn-name: h(1414)
    queue-manager: QM1
    channel: CH
    user-env: EDGE_MQ_USER
    password-env: EDGE_MQ_PASS
    tls: true
    cipher: TLS_X
    key-alias: mc
    topic: T/1
    additional-properties:
      WMQ_SSL_PEER_NAME: "CN=x"
`)
	wf, err := ParseWorkflow(data, "dir/10-x.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if wf.File != "10-x.yaml" || wf.Enabled {
		t.Fatalf("file/enabled: %+v", wf)
	}
	if !wf.SourceSet || !wf.TargetSet {
		t.Fatal("source/target flags")
	}
	s := wf.Source
	if s.System != SystemSolace || s.DestKind != DestQueue || s.Dest != "Q.IN" || s.KeyAlias != "sc" || s.APIProps == nil {
		t.Fatalf("source: %+v (apiprops nil=%v)", s, s.APIProps == nil)
	}
	if s.Username().EnvVar != "EDGE_SOLACE_USER" || s.Secret().EnvVar != "EDGE_SOLACE_PASS" {
		t.Fatalf("source solace -env fields not parsed: %+v", s)
	}
	m := wf.Target
	if m.System != SystemMQ || m.DestKind != DestTopic || m.Dest != "T/1" || !m.TLS || m.Cipher != "TLS_X" || m.AddlProps == nil {
		t.Fatalf("target: %+v", m)
	}
	if m.Username().EnvVar != "EDGE_MQ_USER" || m.Secret().EnvVar != "EDGE_MQ_PASS" {
		t.Fatalf("target mq -env fields not parsed: %+v", m)
	}
}

func TestParseWorkflowConnRef(t *testing.T) {
	wf, err := ParseWorkflow([]byte("source:\n  mq:\n    conn-ref: my-mq\n    queue: A.IN\n"), "10.yaml")
	if err != nil {
		t.Fatal(err)
	}
	s := wf.Source
	if s.System != SystemMQ || s.ConnRef != "my-mq" || s.DestKind != DestQueue || s.Dest != "A.IN" {
		t.Fatalf("conn-ref side: %+v", s)
	}
	if s.SetsConnFields() {
		t.Error("a bare conn-ref side must not report connection fields")
	}
}

func TestConnRefSideMayTuneBinding(t *testing.T) {
	// consumer:/producer: tune one binding, not the connection, so a conn-ref
	// side may set them: SetsConnFields ignores both, and Resolve carries them
	// onto the resolved side alongside the referenced tuple.
	wf, err := ParseWorkflow([]byte("source:\n  mq:\n    conn-ref: my-mq\n    queue: A.IN\n    consumer:\n      concurrency: 4\n"), "10.yaml")
	if err != nil {
		t.Fatal(err)
	}
	s := wf.Source
	if s.Consumer == nil {
		t.Fatal("consumer block should parse on a conn-ref side")
	}
	if s.SetsConnFields() {
		t.Error("consumer/producer are per-binding, so they must not count as connection fields")
	}

	d := &Defaults{Connections: map[string]Side{"my-mq": {
		System: SystemMQ, ConnName: "h(1414)", QueueManager: "QM1", Channel: "CH",
		UserEnv: "MY_MQ_USER", PasswordEnv: "MY_MQ_PASS",
	}}}
	r := d.Resolve(s)
	if r.ConnName != "h(1414)" || r.QueueManager != "QM1" {
		t.Errorf("resolved side lost the referenced tuple: %+v", r)
	}
	if r.Consumer == nil {
		t.Error("resolved side must keep the referring side's consumer block")
	}
	if r.Dest != "A.IN" || r.DestKind != DestQueue {
		t.Errorf("resolved side lost its destination: %+v", r)
	}
}

func TestParseDefaultsConnectionsAndLeaderElection(t *testing.T) {
	d, err := ParseDefaults([]byte(`
connections:
  edge:
    solace:
      host: tcps://b:55443
      msg-vpn: prod
      client-username: u
      client-password-env: EDGE_SOLACE_PASS
      key-alias: sc
  qm:
    mq:
      conn-name: h(1414)
      queue-manager: QM1
      channel: C
      user: u
      password-env: QM_PASS
leader-election:
  mode: active_standby
  queue: mgmt-q
  conn-ref: edge
  fail-over:
    max-attempts: 5
    back-off-multiplier: 1.5
`))
	if err != nil {
		t.Fatal(err)
	}
	if len(d.Connections) != 2 {
		t.Fatalf("connections = %v", d.Connections)
	}
	if e := d.Connections["edge"]; e.System != SystemSolace || e.Host != "tcps://b:55443" || e.KeyAlias != "sc" {
		t.Fatalf("edge connection: %+v", e)
	}
	if e := d.Connections["edge"]; e.Username().Literal != "u" || e.Secret().EnvVar != "EDGE_SOLACE_PASS" {
		t.Fatalf("edge connection credentials: username=%+v secret=%+v", e.Username(), e.Secret())
	}
	if q := d.Connections["qm"]; q.System != SystemMQ || q.QueueManager != "QM1" {
		t.Fatalf("qm connection: %+v", q)
	}
	if q := d.Connections["qm"]; q.Username().Literal != "u" || q.Secret().EnvVar != "QM_PASS" {
		t.Fatalf("qm connection credentials: username=%+v secret=%+v", q.Username(), q.Secret())
	}
	le := d.LeaderElection
	if !le.Present || le.Mode != LeaderActiveStby || le.Queue != "mgmt-q" || le.ConnRef != "edge" || le.FailOver == nil {
		t.Fatalf("leader-election: %+v", le)
	}
}

// TestParseDefaultsLeaderSession covers the inline management session under its
// current key. session: shares the solace block shape with a workflow side --
// that is what makes session.* the same interface as solace.java.* -- so
// everything a connection can carry parses here, api-properties included.
func TestParseDefaultsLeaderSession(t *testing.T) {
	d, err := ParseDefaults([]byte(`
leader-election:
  mode: active_standby
  queue: mgmt-q
  session:
    host: tcps://b:55443
    msg-vpn: prod
    client-username: u
    client-password-env: MGMT_PASS
    key-alias: sc
    api-properties:
      REAPPLY_SUBSCRIPTIONS: true
`))
	if err != nil {
		t.Fatal(err)
	}
	le := d.LeaderElection
	if le.SolaceKey {
		t.Error("session: must not set the retired-key marker")
	}
	if le.Session == nil {
		t.Fatalf("leader-election: %+v", le)
	}
	if le.Session.System != SystemSolace || le.Session.Host != "tcps://b:55443" || le.Session.KeyAlias != "sc" {
		t.Fatalf("session: %+v", *le.Session)
	}
	if le.Session.Secret().EnvVar != "MGMT_PASS" {
		t.Errorf("session credentials: %+v", le.Session.Secret())
	}
	if le.Session.APIProps == nil {
		t.Error("session api-properties must parse; consolidate renders them")
	}
}

// TestParseDefaultsLeaderSolaceKeyRetired pins the retired spelling. It must
// PARSE, whatever shape it holds, so validate can answer with the rename rather
// than the parser answering with a yaml type error that names neither key.
func TestParseDefaultsLeaderSolaceKeyRetired(t *testing.T) {
	for _, c := range []struct {
		name string
		body string
	}{
		{"mapping", "  solace:\n    host: tcps://b:55443\n    msg-vpn: prod\n"},
		{"scalar", "  solace: 5\n"},
		{"list", "  solace:\n    - a\n"},
	} {
		t.Run(c.name, func(t *testing.T) {
			d, err := ParseDefaults([]byte("leader-election:\n  mode: active_standby\n  queue: mgmt-q\n" + c.body))
			if err != nil {
				t.Fatalf("the retired key must parse so validate can name it: %v", err)
			}
			if !d.LeaderElection.SolaceKey {
				t.Error("SolaceKey = false, want the retired-key marker set")
			}
			if d.LeaderElection.Session != nil {
				t.Errorf("the retired key must not populate Session, got %+v", *d.LeaderElection.Session)
			}
		})
	}
}

// TestSideBindingFields pins the predicate validate uses to reject binding keys
// inside a management session. It is the complement of SetsConnFields, which
// deliberately ignores every key here.
func TestSideBindingFields(t *testing.T) {
	node := &yaml.Node{Kind: yaml.MappingNode}
	for _, c := range []struct {
		name string
		side Side
		want []string
	}{
		{"bare tuple", Side{System: SystemSolace, Host: "tcps://b", MsgVPN: "v"}, nil},
		{"queue", Side{DestKind: DestQueue, Dest: "Q"}, []string{DestQueue}},
		{"topic", Side{DestKind: DestTopic, Dest: "T"}, []string{DestTopic}},
		{"consumer", Side{Consumer: node}, []string{"consumer"}},
		{"producer", Side{Producer: node}, []string{"producer"}},
		{"consumer and producer", Side{Consumer: node, Producer: node}, []string{"consumer", "producer"}},
	} {
		t.Run(c.name, func(t *testing.T) {
			if got := c.side.BindingFields(); !slices.Equal(got, c.want) {
				t.Errorf("BindingFields() = %q, want %q", got, c.want)
			}
		})
	}
}

func TestResolveConnRef(t *testing.T) {
	d := &Defaults{Connections: map[string]Side{
		"edge": {System: SystemSolace, Host: "tcps://b", MsgVPN: "v", ClientUser: "u", ClientPassEnv: "EDGE_PASS", KeyAlias: "sc"},
	}}
	r := d.Resolve(Side{System: SystemSolace, ConnRef: "edge", DestKind: DestQueue, Dest: "Q"})
	if r.Host != "tcps://b" || r.MsgVPN != "v" || r.KeyAlias != "sc" || r.Dest != "Q" || r.DestKind != DestQueue {
		t.Fatalf("resolved: %+v", r)
	}
	// Unknown ref is returned unchanged (validate flags it).
	u := d.Resolve(Side{System: SystemMQ, ConnRef: "nope", DestKind: DestQueue, Dest: "Q"})
	if u.ConnRef != "nope" || u.Host != "" {
		t.Fatalf("unknown ref: %+v", u)
	}
}

func TestParseWorkflowEnabledDefaultsTrue(t *testing.T) {
	wf, err := ParseWorkflow([]byte("source:\n  mq: {conn-name: h(1), queue-manager: QM, channel: C, user: u, password: p, queue: Q}\n"), "a.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if !wf.Enabled {
		t.Error("enabled should default true")
	}
	if wf.TargetSet {
		t.Error("target should be unset")
	}
}

func TestParseWorkflowAmbiguousSystemAndDest(t *testing.T) {
	wf, err := ParseWorkflow([]byte("source:\n  solace: {host: x}\n  mq: {conn-name: y}\n"), "a.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if wf.Source.HasSystem() {
		t.Error("both systems -> no system")
	}
	wf2, err := ParseWorkflow([]byte("target:\n  solace:\n    host: x\n    queue: Q\n    topic: T\n"), "a.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if wf2.Target.DestKind != "" {
		t.Errorf("queue+topic should be ambiguous, got %q", wf2.Target.DestKind)
	}
}

func TestParseWorkflowSyntaxError(t *testing.T) {
	if _, err := ParseWorkflow([]byte("source: {bad: [unclosed\n"), "bad.yaml"); err == nil {
		t.Fatal("expected yaml error")
	}
}

// TestParseSecurityUserRoles pins roles parsing on security.users[]: absent
// (the connector's read-only default), one, and several, plus that a role is
// carried verbatim. Roles are not a credential, so unlike password/password-env
// they are expanded like any other identity field -- covered by the ${VAR} case.
func TestParseSecurityUserRoles(t *testing.T) {
	d, err := ParseDefaults([]byte(`
security:
  users:
    - name: readonly
      password-env: RO_PASS
    - name: ops
      password-env: OPS_PASS
      roles:
        - admin
    - name: multi
      password-env: MULTI_PASS
      roles:
        - admin
        - auditor
`))
	if err != nil {
		t.Fatal(err)
	}
	if len(d.Security.Users) != 3 {
		t.Fatalf("users = %+v, want 3", d.Security.Users)
	}
	if got := d.Security.Users[0].Roles; len(got) != 0 {
		t.Errorf("readonly roles = %v, want none (an absent list is the read-only default)", got)
	}
	if got := d.Security.Users[1].Roles; len(got) != 1 || got[0] != "admin" {
		t.Errorf("ops roles = %v, want [admin]", got)
	}
	if got := d.Security.Users[2].Roles; len(got) != 2 || got[0] != "admin" || got[1] != "auditor" {
		t.Errorf("multi roles = %v, want [admin auditor] in order", got)
	}
}

func TestParseDefaultsFull(t *testing.T) {
	d, err := ParseDefaults([]byte(`
tls:
  truststore:
    file: ./t.jks
    password-env: TRUSTSTORE_PASS
    type: JKS
  keystore:
    file: ./k.jks
    password: keystore-literal-pw
    type: JKS
logging:
  level: {root: INFO}
management:
  port: 8090
  exposure: health
  health-show-details: always
security:
  enabled: false
  users:
    - name: hc
      password-env: HEALTHCHECK_PASS
leader-election:
  mode: standalone
solace-defaults:
  connect-retries: -1
`))
	if err != nil {
		t.Fatal(err)
	}
	if d.TLS.Truststore == nil || d.TLS.Keystore == nil {
		t.Fatal("stores not parsed")
	}
	if d.TLS.Truststore.Secret().EnvVar != "TRUSTSTORE_PASS" {
		t.Errorf("truststore secret = %+v", d.TLS.Truststore.Secret())
	}
	if d.TLS.Keystore.Secret().Literal != "keystore-literal-pw" {
		t.Errorf("keystore secret = %+v", d.TLS.Keystore.Secret())
	}
	if d.Management.Port != 8090 || d.Management.Exposure == nil || *d.Management.Exposure != "health" {
		t.Errorf("management: %+v", d.Management)
	}
	if d.Security.Enabled == nil || *d.Security.Enabled || len(d.Security.Users) != 1 {
		t.Errorf("security: %+v", d.Security)
	}
	if d.Security.Users[0].Secret().EnvVar != "HEALTHCHECK_PASS" {
		t.Errorf("user secret = %+v", d.Security.Users[0].Secret())
	}
	if !d.LeaderElection.Present || d.LeaderElection.Mode != LeaderStandalone {
		t.Errorf("leader: %+v", d.LeaderElection)
	}
	if d.LoggingLevel == nil || d.SolaceDefaults == nil {
		t.Error("logging/solace-defaults nodes not captured")
	}
}

func TestParseDefaultsSecurityEnabledKeyOmittedStaysNil(t *testing.T) {
	// security.enabled is a removed key: defaultsFromRaw must copy it straight
	// through rather than default it, so an absent key is reachable as nil and
	// validate can tell "omitted" apart from "set to false".
	d, err := ParseDefaults([]byte("security:\n  users: []\n"))
	if err != nil {
		t.Fatal(err)
	}
	if d.Security.Enabled != nil {
		t.Errorf("security.enabled should stay nil when omitted: %+v", d.Security.Enabled)
	}
}

func TestParseDefaultsEmpty(t *testing.T) {
	d, err := ParseDefaults([]byte(""))
	if err != nil {
		t.Fatal(err)
	}
	if d.Management != (Management{}) || d.Security.Enabled != nil || len(d.Security.Users) != 0 || d.TLS.Truststore != nil {
		t.Errorf("empty defaults should be zero-valued: %+v", d)
	}
}

func TestParseDefaultsError(t *testing.T) {
	if _, err := ParseDefaults([]byte("tls: [bad\n")); err == nil {
		t.Fatal("expected error")
	}
}

func TestParseKubernetesReplicasDefault(t *testing.T) {
	k, err := ParseKubernetes([]byte("deployment:\n  name: c\n  namespace: ns\n  image: img\n"))
	if err != nil {
		t.Fatal(err)
	}
	if k.Deployment.Replicas != 1 {
		t.Errorf("replicas default = %d want 1", k.Deployment.Replicas)
	}
}

func TestParseKubernetesFull(t *testing.T) {
	k, err := ParseKubernetes([]byte(`
deployment: {name: c, namespace: ns, image: img, replicas: 2}
service: {enabled: true, port: 8090}
secrets:
  credentials:
    create: {name: s, source: env, variables: [A, B]}
  stores:
    create: {name: t}
`))
	if err != nil {
		t.Fatal(err)
	}
	if k.Deployment.Replicas != 2 || !k.Service.Enabled || k.Service.Port != (Port{Host: 8090, Container: 8090}) {
		t.Errorf("kube: %+v", k)
	}
	if k.Secrets.Credentials == nil || k.Secrets.Credentials.Create == nil || k.Secrets.Credentials.Create.Name != "s" {
		t.Errorf("cred: %+v", k.Secrets.Credentials)
	}
	// source/variables are removed keys: they still parse (so RemovedKeys can see
	// them) but no longer carry meaning -- RemovedKeys is how a caller rejects them.
	if rk := k.Secrets.Credentials.Create.RemovedKeys(); len(rk) != 2 || rk[0] != "source" || rk[1] != "variables" {
		t.Errorf("removed keys = %v", rk)
	}
	if k.Secrets.Stores == nil || k.Secrets.Stores.Create == nil {
		t.Errorf("stores: %+v", k.Secrets.Stores)
	}
}

func TestParseKubernetesError(t *testing.T) {
	if _, err := ParseKubernetes([]byte("deployment: [not a map\n")); err == nil {
		t.Fatal("expected error")
	}
}

func TestParseKubernetesResources(t *testing.T) {
	k, err := ParseKubernetes([]byte("deployment:\n  name: c\n  namespace: ns\n  image: img\n  resources:\n    cpu: \"1\"\n    memory: 1Gi\n"))
	if err != nil {
		t.Fatal(err)
	}
	if k.Deployment.Resources.CPU != "1" || k.Deployment.Resources.Memory != "1Gi" {
		t.Fatalf("resources = %+v", k.Deployment.Resources)
	}
}

func TestParseKubernetesLoggingLibsDefaults(t *testing.T) {
	k, err := ParseKubernetes([]byte(`
deployment: {name: c, namespace: ns, image: img}
logging:
  syslog: {host: h, port: 514}
libs:
  download: {urls: [https://x/a.jar]}
`))
	if err != nil {
		t.Fatal(err)
	}
	// kubernetes.logging still PARSES -- that is the precondition for validate
	// rejecting it by name instead of it being silently dropped -- but nothing
	// defaults it here any more: the live block is top-level.
	if k.Logging == nil || k.Logging.Syslog == nil {
		t.Fatal("the retired kubernetes.logging block must still parse, so validate can reject it")
	}
	if k.Logging.Syslog.Protocol != "" {
		t.Errorf("the retired block must not be defaulted, got protocol %q", k.Logging.Syslog.Protocol)
	}
	if k.Libs.Download.Image != "busybox:1.37" {
		t.Errorf("download image default = %q", k.Libs.Download.Image)
	}
	k2, err := ParseKubernetes([]byte(`
deployment: {name: c, namespace: ns, image: img}
libs:
  pvc: {create: {name: p, nfs: {server: s, path: /x}}}
`))
	if err != nil {
		t.Fatal(err)
	}
	if k2.Libs.PVC.Create.Storage != "1Gi" {
		t.Errorf("storage default = %q", k2.Libs.PVC.Create.Storage)
	}
	k3, err := ParseKubernetes([]byte("deployment: {name: c, namespace: ns, image: img}\n"))
	if err != nil {
		t.Fatal(err)
	}
	if k3.Logging != nil || k3.Libs != nil {
		t.Error("absent logging/libs blocks must stay nil")
	}
}

func TestCredEmptyBothKeyDescribe(t *testing.T) {
	tests := []struct {
		name      string
		c         Cred
		wantEmpty bool
		wantBoth  bool
		wantKey   string
		wantDesc  string
	}{
		{"unset", Cred{}, true, false, "", "nothing"},
		{"literal only", Cred{Literal: "s3cret"}, false, false, "L:s3cret", "a literal value"},
		{"env only", Cred{EnvVar: "EDGE_PASS"}, false, false, "E:EDGE_PASS", "the environment variable EDGE_PASS"},
		// Both set is an over-specification validate rejects, but Key/Describe must
		// still resolve deterministically (env wins) rather than panic.
		{"both set", Cred{Literal: "s3cret", EnvVar: "EDGE_PASS"}, false, true, "E:EDGE_PASS", "the environment variable EDGE_PASS"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.c.Empty(); got != tt.wantEmpty {
				t.Errorf("Empty() = %v want %v", got, tt.wantEmpty)
			}
			if got := tt.c.Both(); got != tt.wantBoth {
				t.Errorf("Both() = %v want %v", got, tt.wantBoth)
			}
			if got := tt.c.Key(); got != tt.wantKey {
				t.Errorf("Key() = %q want %q", got, tt.wantKey)
			}
			if got := tt.c.Describe(); got != tt.wantDesc {
				t.Errorf("Describe() = %q want %q", got, tt.wantDesc)
			}
		})
	}
}

func TestSideUsernameSecretBothSystems(t *testing.T) {
	solace := Side{
		System:        SystemSolace,
		ClientUser:    "sol-user",
		ClientPassEnv: "SOLACE_PASS",
		// MQ fields are set too, to prove dispatch is by System, not by whichever
		// pair happens to be non-empty.
		User:        "mq-user",
		PasswordEnv: "MQ_PASS",
	}
	if got := solace.Username(); got.Literal != "sol-user" {
		t.Errorf("solace Username() = %+v want literal sol-user", got)
	}
	if got := solace.Secret(); got.EnvVar != "SOLACE_PASS" {
		t.Errorf("solace Secret() = %+v want env SOLACE_PASS", got)
	}

	mq := Side{
		System:        SystemMQ,
		User:          "mq-user",
		PasswordEnv:   "MQ_PASS",
		ClientUser:    "sol-user",
		ClientPassEnv: "SOLACE_PASS",
	}
	if got := mq.Username(); got.Literal != "mq-user" {
		t.Errorf("mq Username() = %+v want literal mq-user", got)
	}
	if got := mq.Secret(); got.EnvVar != "MQ_PASS" {
		t.Errorf("mq Secret() = %+v want env MQ_PASS", got)
	}
}

func TestStoreSecretNilSafe(t *testing.T) {
	var nilStore *Store
	if got := nilStore.Secret(); !got.Empty() {
		t.Errorf("nil *Store Secret() = %+v want empty", got)
	}

	literal := &Store{Password: "lit-pw"}
	if got := literal.Secret(); got.Literal != "lit-pw" || got.EnvVar != "" {
		t.Errorf("literal store Secret() = %+v", got)
	}

	env := &Store{PasswordEnv: "STORE_PASS"}
	if got := env.Secret(); got.EnvVar != "STORE_PASS" {
		t.Errorf("env store Secret() = %+v want env STORE_PASS", got)
	}
}

func TestUserSecretLiteralAndEnv(t *testing.T) {
	literal := User{Name: "hc", Password: "lit-pw"}
	if got := literal.Secret(); got.Literal != "lit-pw" {
		t.Errorf("literal user Secret() = %+v", got)
	}
	env := User{Name: "hc", PasswordEnv: "HC_PASS"}
	if got := env.Secret(); got.EnvVar != "HC_PASS" {
		t.Errorf("env user Secret() = %+v want env HC_PASS", got)
	}
}

func TestCredCreateRemovedKeys(t *testing.T) {
	var nilCreate *CredCreate
	if got := nilCreate.RemovedKeys(); got != nil {
		t.Errorf("nil *CredCreate RemovedKeys() = %v want nil", got)
	}
	if got := (&CredCreate{Name: "s"}).RemovedKeys(); got != nil {
		t.Errorf("no removed keys set, RemovedKeys() = %v want nil", got)
	}
	if got := (&CredCreate{Source: "env"}).RemovedKeys(); len(got) != 1 || got[0] != "source" {
		t.Errorf("source only RemovedKeys() = %v want [source]", got)
	}
	if got := (&CredCreate{Variables: []string{"A"}}).RemovedKeys(); len(got) != 1 || got[0] != "variables" {
		t.Errorf("variables only RemovedKeys() = %v want [variables]", got)
	}
	if got := (&CredCreate{ValuesFile: "vals.env"}).RemovedKeys(); len(got) != 1 || got[0] != "values-file" {
		t.Errorf("values-file only RemovedKeys() = %v want [values-file]", got)
	}
	all := &CredCreate{Source: "env", Variables: []string{"A", "B"}, ValuesFile: "vals.env"}
	if got := all.RemovedKeys(); len(got) != 3 || got[0] != "source" || got[1] != "variables" || got[2] != "values-file" {
		t.Errorf("all removed keys = %v want [source variables values-file]", got)
	}
}

func TestBaseName(t *testing.T) {
	// One definition serves the store mount path, the Kubernetes stores-Secret
	// data key, and the libs download filename, so it must resolve a
	// Windows-authored path the same way wherever the CLI runs. The deploy copy
	// used to leave backslashes alone, which only stayed safe because validation
	// happened to reject them upstream.
	cases := []struct{ in, want string }{
		{"https://repo/a.jar", "a.jar"},
		{"a/b/c.jks", "c.jks"},
		{`a\b\c.jks`, "c.jks"},
		{`C:\certs\truststore.jks`, "truststore.jks"},
		{"noslash", "noslash"},
		{"", ""},
	}
	for _, c := range cases {
		if got := BaseName(c.in); got != c.want {
			t.Errorf("BaseName(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

// TestWorkflowFileLess covers the ordering workflow ids are assigned from:
// digit runs compare as numbers, everything else byte by byte. The sort/reverse
// pairs also pin that the comparator is a strict order (never true both ways),
// which sort.Slice relies on.
func TestWorkflowFileLess(t *testing.T) {
	cases := []struct {
		a, b string
		want bool // want WorkflowFileLess(a, b); the reverse must be false
	}{
		{"2.yaml", "10.yaml", true},                   // the case plain lexical order gets wrong
		{"9.yaml", "10.yaml", true},                   // ...and its 9/10 boundary
		{"workflow-2.yaml", "workflow-10.yaml", true}, // digits after a shared prefix
		{"1.yaml", "2.yaml", true},                    // same width, still numeric
		{"10.yaml", "10.yml", true},                   // equal numbers fall through to the suffix
		{"a.yaml", "b.yaml", true},                    // no digits at all: byte order
		{"7.yaml", "007.yaml", false},                 // same value, padding breaks the tie
		{"2.yaml", "2b.yaml", true},                   // equal digits, then '.' before 'b'
	}
	for _, c := range cases {
		if got := WorkflowFileLess(c.a, c.b); got != c.want {
			t.Errorf("WorkflowFileLess(%q, %q) = %v, want %v", c.a, c.b, got, c.want)
		}
		if got := WorkflowFileLess(c.b, c.a); got == c.want {
			t.Errorf("WorkflowFileLess(%q, %q) = %v: not a strict order", c.b, c.a, got)
		}
	}
	// A name compared with itself is never less than itself -- the irreflexivity
	// sort.Slice needs to avoid an inconsistent comparator.
	if WorkflowFileLess("10.yaml", "10.yaml") {
		t.Error("WorkflowFileLess(x, x) must be false")
	}
}

// TestParseEnvTopLevelSyslog pins the live location: logging.syslog sits beside
// logging.level at the top level, and udp is filled in when protocol is unset.
func TestParseEnvTopLevelSyslog(t *testing.T) {
	e, err := ParseEnv([]byte(`
logging:
  level:
    root: INFO
  syslog:
    host: syslog.corp
    port: 514
`))
	if err != nil {
		t.Fatal(err)
	}
	if e.Syslog == nil {
		t.Fatal("top-level logging.syslog did not parse")
	}
	if e.Syslog.Host != "syslog.corp" || e.Syslog.Port != 514 {
		t.Errorf("syslog = %+v", e.Syslog)
	}
	if e.Syslog.Protocol != SyslogUDP {
		t.Errorf("protocol default = %q, want udp", e.Syslog.Protocol)
	}
	if e.LoggingLevel == nil {
		t.Error("logging.level must still parse alongside syslog")
	}

	// Absent block stays nil: presence is what turns syslog on.
	e2, err := ParseEnv([]byte("logging:\n  level:\n    root: INFO\n"))
	if err != nil {
		t.Fatal(err)
	}
	if e2.Syslog != nil {
		t.Error("an absent syslog block must stay nil")
	}
}
