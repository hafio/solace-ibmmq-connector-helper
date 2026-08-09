package spec

import "testing"

func TestParseWorkflowSolaceAndMQ(t *testing.T) {
	data := []byte(`
enabled: false
source:
  solace:
    host: tcps://b:55443
    msg-vpn: prod
    client-username: u
    client-password: ${P}
    key-alias: sc
    queue: Q.IN
    api-properties:
      REAPPLY_SUBSCRIPTIONS: true
target:
  mq:
    conn-name: h(1414)
    queue-manager: QM1
    channel: CH
    user: app
    password: ${MQ}
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
	m := wf.Target
	if m.System != SystemMQ || m.DestKind != DestTopic || m.Dest != "T/1" || !m.TLS || m.Cipher != "TLS_X" || m.AddlProps == nil {
		t.Fatalf("target: %+v", m)
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

func TestParseDefaultsConnectionsAndLeaderElection(t *testing.T) {
	d, err := ParseDefaults([]byte(`
connections:
  edge:
    solace:
      host: tcps://b:55443
      msg-vpn: prod
      client-username: u
      client-password: ${P}
      key-alias: sc
  qm:
    mq:
      conn-name: h(1414)
      queue-manager: QM1
      channel: C
      user: u
      password: ${MQ}
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
	if q := d.Connections["qm"]; q.System != SystemMQ || q.QueueManager != "QM1" {
		t.Fatalf("qm connection: %+v", q)
	}
	le := d.LeaderElection
	if !le.Present || le.Mode != LeaderActiveStby || le.Queue != "mgmt-q" || le.ConnRef != "edge" || le.FailOver == nil {
		t.Fatalf("leader-election: %+v", le)
	}
}

func TestResolveConnRef(t *testing.T) {
	d := &Defaults{Connections: map[string]Side{
		"edge": {System: SystemSolace, Host: "tcps://b", MsgVPN: "v", ClientUser: "u", ClientPass: "${P}", KeyAlias: "sc"},
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

func TestParseDefaultsFull(t *testing.T) {
	d, err := ParseDefaults([]byte(`
tls:
  truststore:
    file: ./t.jks
    password: ${T}
    type: JKS
  keystore:
    file: ./k.jks
    password: ${K}
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
      password: ${H}
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
	if !d.Management.Present || d.Management.Port != 8090 {
		t.Errorf("management: %+v", d.Management)
	}
	if !d.Security.Present || d.Security.Enabled || len(d.Security.Users) != 1 {
		t.Errorf("security: %+v", d.Security)
	}
	if !d.LeaderElection.Present || d.LeaderElection.Mode != LeaderStandalone {
		t.Errorf("leader: %+v", d.LeaderElection)
	}
	if d.LoggingLevel == nil || d.SolaceDefaults == nil {
		t.Error("logging/solace-defaults nodes not captured")
	}
}

func TestParseDefaultsSecurityDefaultsEnabled(t *testing.T) {
	d, err := ParseDefaults([]byte("security:\n  users: []\n"))
	if err != nil {
		t.Fatal(err)
	}
	if !d.Security.Present || !d.Security.Enabled {
		t.Errorf("security should default enabled when present: %+v", d.Security)
	}
}

func TestParseDefaultsEmpty(t *testing.T) {
	d, err := ParseDefaults([]byte(""))
	if err != nil {
		t.Fatal(err)
	}
	if d.Management.Present || d.Security.Present || d.TLS.Truststore != nil {
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
	if k.Deployment.Replicas != 2 || !k.Service.Enabled || k.Service.Port != 8090 {
		t.Errorf("kube: %+v", k)
	}
	if k.Secrets.Credentials == nil || k.Secrets.Credentials.Create == nil || k.Secrets.Credentials.Create.Source != SourceEnv {
		t.Errorf("cred: %+v", k.Secrets.Credentials)
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
	if k.Logging.Syslog.Protocol != SyslogUDP {
		t.Errorf("protocol default = %q want udp", k.Logging.Syslog.Protocol)
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
