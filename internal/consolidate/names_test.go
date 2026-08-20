package consolidate

import (
	"regexp"
	"testing"

	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/spec"
)

// TestStableTokenFolding pins stableToken's folding rules: upper-snake letters
// and digits pass through, any other rune becomes '_' with runs collapsed and
// leading/trailing '_' trimmed, a leading digit gets an 'X' prefix, and an empty
// result (nothing foldable) becomes "X" rather than a bare "_" or "".
func TestStableTokenFolding(t *testing.T) {
	cases := []struct{ in, want string }{
		{"mq-conn-1", "MQ_CONN_1"},
		{"svc.a", "SVC_A"},
		{"svc/a", "SVC_A"},
		{"  weird--name!!", "WEIRD_NAME"},
		{"123abc", "X123ABC"},
		{"", "X"},
		{"___", "X"},
		{"a_b__c", "A_B_C"},
	}
	for _, c := range cases {
		if got := stableToken(c.in); got != c.want {
			t.Errorf("stableToken(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

// TestBinderFieldsCarryStablePlaceholders is the S3 guarantee at the binder
// level: no credential value or host env-var name -- literal or -env -- ever
// reaches a rendered binder field, only the derived ${STABLE_NAME} placeholder,
// and Model.Secrets records the real source (Literal or EnvVar) under that name
// so the deploy layer can resolve it.
func TestBinderFieldsCarryStablePlaceholders(t *testing.T) {
	sol := spec.Side{System: spec.SystemSolace, Host: "tcps://b", MsgVPN: "v",
		ClientUserEnv: "SOL_USER_ENV", ClientPass: "sol-literal-pass",
		DestKind: spec.DestQueue, Dest: "OUT"}
	mq := spec.Side{System: spec.SystemMQ, ConnName: "h(1)", QueueManager: "QM", Channel: "C",
		User: "mq-literal-user", PasswordEnv: "MQ_PASS_ENV",
		DestKind: spec.DestQueue, Dest: "IN"}
	wfs := []spec.Workflow{{File: "10.yaml", Enabled: true, SourceSet: true, TargetSet: true, Source: sol, Target: mq}}

	m, _ := Build(wfs, nil, Opts{MountStores: true})

	var solBinder *SolaceBinder
	var mqBinder *MQBinder
	for _, b := range m.Binders {
		switch b.Name {
		case "sol-conn-1":
			solBinder = b.Solace
		case "mq-conn-1":
			mqBinder = b.MQ
		}
	}
	if solBinder == nil || mqBinder == nil {
		t.Fatalf("binders = %v, want sol-conn-1 and mq-conn-1", binderNames(m))
	}

	if solBinder.ClientUser != "${SOL_CONN_1_CLIENT_USERNAME}" {
		t.Errorf("SolaceBinder.ClientUser = %q, want the SOL_CONN_1_CLIENT_USERNAME placeholder (never SOL_USER_ENV)", solBinder.ClientUser)
	}
	if solBinder.ClientPass != "${SOL_CONN_1_CLIENT_PASSWORD}" {
		t.Errorf("SolaceBinder.ClientPass = %q, want the SOL_CONN_1_CLIENT_PASSWORD placeholder (never the literal)", solBinder.ClientPass)
	}
	if mqBinder.User != "${MQ_CONN_1_USER}" {
		t.Errorf("MQBinder.User = %q, want the MQ_CONN_1_USER placeholder (never the literal)", mqBinder.User)
	}
	if mqBinder.Password != "${MQ_CONN_1_PASSWORD}" {
		t.Errorf("MQBinder.Password = %q, want the MQ_CONN_1_PASSWORD placeholder (never MQ_PASS_ENV)", mqBinder.Password)
	}

	want := map[string]SecretRef{
		"SOL_CONN_1_CLIENT_USERNAME": {Stable: "SOL_CONN_1_CLIENT_USERNAME", EnvVar: "SOL_USER_ENV"},
		"SOL_CONN_1_CLIENT_PASSWORD": {Stable: "SOL_CONN_1_CLIENT_PASSWORD", Literal: "sol-literal-pass"},
		"MQ_CONN_1_USER":             {Stable: "MQ_CONN_1_USER", Literal: "mq-literal-user"},
		"MQ_CONN_1_PASSWORD":         {Stable: "MQ_CONN_1_PASSWORD", EnvVar: "MQ_PASS_ENV"},
	}
	if len(m.Secrets) != len(want) {
		t.Fatalf("Secrets = %+v, want %d entries", m.Secrets, len(want))
	}
	for _, got := range m.Secrets {
		w, ok := want[got.Stable]
		if !ok {
			t.Errorf("unexpected Secrets entry %+v", got)
			continue
		}
		if got != w {
			t.Errorf("Secrets[%s] = %+v, want %+v", got.Stable, got, w)
		}
	}
}

// stableSecretName is every generated-secret-name shape the connector emits:
// stableName(binder, suffix), securityUserPasswordName(user), and the fixed
// truststore/keystore/leader-election constants. All fold through stableToken,
// so config-controlled text (a connection name, a security-user name) can only
// ever land in the run of characters before the fixed suffix -- never replace
// it -- but this pins the actual suffix vocabulary in use today (MQBinder.User
// is "..._USER", not "..._USERNAME") rather than assuming it.
var stableSecretName = regexp.MustCompile(`^[A-Z0-9_]+_(PASSWORD|USERNAME|USER)$`)

// TestGeneratedSecretNamesStayOutOfChildEnvDanger is the S3 guarantee behind
// runner.Docker's env param: every SecretRef.Stable this package can produce is
// injected as a name into the docker-compose child process's environment, so if
// config text (a connection name, a security-user name) could ever fold into a
// dangerous identifier like PATH or LD_PRELOAD, the config file would control
// what the child process loads. stableToken uppercases and folds adversarial
// input, but the real guarantee is the fixed, code-chosen suffix appended after
// it (_PASSWORD/_USERNAME/_USER, or one of the four exported constants) --
// config text only ever supplies the run of characters *before* that suffix, so
// it can fold to "LD_PRELOAD_PASSWORD" but never to bare "LD_PRELOAD" or "PATH".
func TestGeneratedSecretNamesStayOutOfChildEnvDanger(t *testing.T) {
	d := &spec.Defaults{
		Connections: map[string]spec.Side{
			// Adversarial conn-ref names chosen to fold toward dangerous env-var
			// identifiers if the fixed suffix were ever dropped.
			"path": {System: spec.SystemSolace, Host: "tcps://h", MsgVPN: "v",
				ClientUserEnv: "SRC_USER_ENV", ClientPass: "src-literal-pass"},
			"ld-preload": {System: spec.SystemMQ, ConnName: "h(1414)", QueueManager: "QM", Channel: "C",
				User: "mq-literal-user", PasswordEnv: "MQ_PASS_ENV", TLS: true, KeyAlias: "alias1"},
		},
		Security: spec.Security{Users: []spec.User{
			{Name: "LD", Password: "sekret"},
			{Name: "../../etc/passwd", PasswordEnv: "WEIRD_ENV"},
		}},
		TLS: spec.TLSConfig{
			Truststore: &spec.Store{File: "ts.jks", Password: "ts-literal-pass", Type: "JKS"},
			Keystore:   &spec.Store{File: "ks.jks", PasswordEnv: "KS_PASS_ENV", Type: "JKS"},
		},
		LeaderElection: spec.LeaderElection{Present: true, Mode: spec.LeaderActiveActive, Queue: "leader-q",
			Session: &spec.Side{System: spec.SystemSolace, Host: "tcps://lh", MsgVPN: "lv",
				ClientUserEnv: "LEADER_USER_ENV", ClientPass: "leader-literal-pass"}},
	}
	wfs := []spec.Workflow{{File: "10.yaml", Enabled: true, SourceSet: true, TargetSet: true,
		Source: spec.Side{System: spec.SystemSolace, ConnRef: "path", DestKind: spec.DestQueue, Dest: "IN"},
		Target: spec.Side{System: spec.SystemMQ, ConnRef: "ld-preload", DestKind: spec.DestQueue, Dest: "OUT"},
	}}

	m, _ := Build(wfs, d, Opts{MountStores: true})

	if len(m.Secrets) == 0 {
		t.Fatal("Secrets is empty; the fixture did not exercise any secret-name producer")
	}
	for _, s := range m.Secrets {
		if !stableSecretName.MatchString(s.Stable) {
			t.Errorf("secret name %q does not match %s: an adversarial config value may be able to produce a dangerous child-process env var name", s.Stable, stableSecretName.String())
		}
	}
}
