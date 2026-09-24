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
// level: no credential *value* -- literal or -env -- ever reaches a rendered
// binder field, only a ${NAME} placeholder, and Model.Secrets records the real
// source (Literal or EnvVar) under that name so the deploy layer can resolve it.
//
// The name is the operator's own -env variable when there is one, so a
// hand-built `existing:` Secret keys on the names written in the spec; a literal
// has no name of its own and takes the derived stableName for its position. This
// pins both halves against one fixture that mixes the two forms on each side.
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

	if solBinder.ClientUser != "${SOL_USER_ENV}" {
		t.Errorf("SolaceBinder.ClientUser = %q, want the SOL_USER_ENV placeholder: an -env credential is mounted under the operator's own variable name", solBinder.ClientUser)
	}
	if solBinder.ClientPass != "${_GEN_SOL_CONN_1_CLIENT_PASSWORD}" {
		t.Errorf("SolaceBinder.ClientPass = %q, want the derived _GEN_SOL_CONN_1_CLIENT_PASSWORD placeholder (never the literal value)", solBinder.ClientPass)
	}
	if mqBinder.User != "${_GEN_MQ_CONN_1_USER}" {
		t.Errorf("MQBinder.User = %q, want the derived _GEN_MQ_CONN_1_USER placeholder (never the literal value)", mqBinder.User)
	}
	if mqBinder.Password != "${MQ_PASS_ENV}" {
		t.Errorf("MQBinder.Password = %q, want the MQ_PASS_ENV placeholder: an -env credential is mounted under the operator's own variable name", mqBinder.Password)
	}

	want := map[string]SecretRef{
		"SOL_USER_ENV":                    {Stable: "SOL_USER_ENV", EnvVar: "SOL_USER_ENV"},
		"_GEN_SOL_CONN_1_CLIENT_PASSWORD": {Stable: "_GEN_SOL_CONN_1_CLIENT_PASSWORD", Literal: "sol-literal-pass"},
		"_GEN_MQ_CONN_1_USER":             {Stable: "_GEN_MQ_CONN_1_USER", Literal: "mq-literal-user"},
		"MQ_PASS_ENV":                     {Stable: "MQ_PASS_ENV", EnvVar: "MQ_PASS_ENV"},
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

// TestEnvCredentialsShareOneMountName pins the dedup half of secretRef: two
// positions naming the same host variable are one credential, so they contribute
// one mounted file rather than two files holding the same value under two names.
// This is the normal case for a password shared across binders, and it must not
// be mistaken for the collision the next test covers.
func TestEnvCredentialsShareOneMountName(t *testing.T) {
	mqA := spec.Side{System: spec.SystemMQ, ConnName: "a(1414)", QueueManager: "QMA", Channel: "C",
		PasswordEnv: "SHARED_PASS", DestKind: spec.DestQueue, Dest: "IN"}
	mqB := spec.Side{System: spec.SystemMQ, ConnName: "b(1414)", QueueManager: "QMB", Channel: "C",
		PasswordEnv: "SHARED_PASS", DestKind: spec.DestQueue, Dest: "OUT"}
	wfs := []spec.Workflow{{File: "10.yaml", Enabled: true, SourceSet: true, TargetSet: true, Source: mqA, Target: mqB}}

	m, _ := Build(wfs, nil, Opts{MountStores: true})

	if len(m.SecretConflicts) != 0 {
		t.Errorf("SecretConflicts = %v, want none: one variable named by two positions is one credential", m.SecretConflicts)
	}
	if len(m.Secrets) != 1 || m.Secrets[0].Stable != "SHARED_PASS" {
		t.Errorf("Secrets = %+v, want exactly one SHARED_PASS entry", m.Secrets)
	}
}

// TestSecretNameConflictIsRecorded pins the collision guard against the case
// that survives the reserved prefix. validate rejects an -env variable inside
// spec.GeneratedNamePrefix, so an operator's name can no longer reach a derived
// one; what remains is two *derived* names folding together, because stableToken
// maps every run of non-alphanumeric characters to a single '_'. Two management
// users differing only in punctuation are the reachable example, and nothing
// upstream rejects them: validate checks security.users names only against the
// reserved status account.
//
// One file cannot hold two passwords, so Build records the name and both
// claiming positions for gen to refuse, rather than silently dropping one.
func TestSecretNameConflictIsRecorded(t *testing.T) {
	sol := spec.Side{System: spec.SystemSolace, Host: "tcp://b", MsgVPN: "v",
		DestKind: spec.DestQueue, Dest: "OUT"}
	mq := spec.Side{System: spec.SystemMQ, ConnName: "h(1)", QueueManager: "QM", Channel: "C",
		DestKind: spec.DestQueue, Dest: "IN"}
	// "ops.1" and "ops-1" both fold to OPS_1.
	d := &spec.Defaults{Security: spec.Security{Users: []spec.User{
		{Name: "ops.1", Password: "first-pass"},
		{Name: "ops-1", Password: "second-pass"},
	}}}
	wfs := []spec.Workflow{{File: "10.yaml", Enabled: true, SourceSet: true, TargetSet: true, Source: sol, Target: mq}}

	m, _ := Build(wfs, d, Opts{MountStores: true, StatusPassword: "status-pw"})

	want := SecretConflict{
		Name:   "_GEN_SECURITY_USER_OPS_1_PASSWORD",
		First:  "security.users[ops.1].password",
		Second: "security.users[ops-1].password",
	}
	if len(m.SecretConflicts) != 1 || m.SecretConflicts[0] != want {
		t.Fatalf("SecretConflicts = %+v, want [%+v]", m.SecretConflicts, want)
	}
	// The first credential still wins the name, so the model stays deterministic
	// for any caller that ignores the conflict list.
	for _, s := range m.Secrets {
		if s.Stable == want.Name && s.Literal != "first-pass" {
			t.Errorf("Secrets[%s] = %+v, want the first credential kept", want.Name, s)
		}
	}
}

// stableSecretName is every *derived* secret-name shape the connector emits:
// stableName(binder, suffix), securityUserPasswordName(user), and the fixed
// truststore/keystore/leader-election constants. All fold through stableToken,
// so config-controlled text (a connection name, a security-user name) can only
// ever land in the run of characters before the fixed suffix -- never replace
// it -- but this pins the actual suffix vocabulary in use today (MQBinder.User
// is "..._USER", not "..._USERNAME") rather than assuming it.
//
// It applies to literal credentials only. An -env credential is mounted under
// the operator's own variable name, which is deliberately not a derived name;
// see TestGeneratedSecretNamesStayOutOfChildEnvDanger for why that stays safe.
//
// The spec.GeneratedNamePrefix anchor is the load-bearing part: it is what keeps
// every derived name inside a namespace no operator would export, so a derived
// name and an -env name cannot collide by accident.
var stableSecretName = regexp.MustCompile(`^` + regexp.QuoteMeta(spec.GeneratedNamePrefix) + `[A-Z0-9_]+_(PASSWORD|USERNAME|USER)$`)

// TestGeneratedSecretNamesStayOutOfChildEnvDanger is the S3 guarantee behind
// runner.Docker's env param: every SecretRef.Stable this package can produce is
// injected as a name into the docker-compose child process's environment, so if
// config text (a connection name, a security-user name) could ever fold into a
// dangerous identifier like PATH or LD_PRELOAD, the config file would control
// what the child process loads. The guarantee now has two halves, because a
// credential's mount name is its -env variable when it has one and a derived
// name otherwise:
//
//   - Literal credentials take a derived name. stableToken uppercases and folds
//     adversarial input, but the real protection is the fixed, code-chosen
//     suffix appended after it (_PASSWORD/_USERNAME/_USER, or one of the four
//     exported constants) -- config text only ever supplies the run of
//     characters *before* that suffix, so it can fold to "LD_PRELOAD_PASSWORD"
//     but never to bare "LD_PRELOAD" or "PATH".
//
//   - An -env credential's name is operator-chosen and so *could* be "PATH".
//     That is safe because its name and its source are then the same variable:
//     envPairs emits `<EnvVar>=<value of EnvVar>` and runner.applyCmdEnv appends
//     it to os.Environ(), so the pair can only restate what the child already
//     inherited. Naming a dangerous variable overwrites it with its own value.
//     Stable == EnvVar is what makes that argument hold, so it is what is pinned.
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
	var literals, envs int
	for _, s := range m.Secrets {
		if s.EnvVar != "" {
			envs++
			if s.Stable != s.EnvVar {
				t.Errorf("secret %+v: an -env credential must be mounted under its own variable name, or the pair injected into the compose child stops being a restatement of what it inherited", s)
			}
			continue
		}
		literals++
		if !stableSecretName.MatchString(s.Stable) {
			t.Errorf("literal secret name %q does not match %s: an adversarial config value may be able to produce a dangerous child-process env var name", s.Stable, stableSecretName.String())
		}
	}
	// Both halves of the guarantee must actually be exercised, or a future change
	// to the fixture could silently retire one of them.
	if literals == 0 || envs == 0 {
		t.Errorf("fixture produced %d literal and %d -env secrets; it must exercise both naming paths", literals, envs)
	}
}
