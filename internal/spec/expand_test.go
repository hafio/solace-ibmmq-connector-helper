package spec

import (
	"fmt"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// lookupOf builds an Expander.Lookup from a plain map, mirroring os.LookupEnv.
func lookupOf(env map[string]string) func(string) (string, bool) {
	return func(name string) (string, bool) {
		v, ok := env[name]
		return v, ok
	}
}

func TestExpandBracedVar(t *testing.T) {
	wfs := []Workflow{{Source: Side{Host: "tcps://${HOST}:55443"}}}
	Expand(Expander{Lookup: lookupOf(map[string]string{"HOST": "broker.internal"})}, &Env{}, wfs)
	if got, want := wfs[0].Source.Host, "tcps://broker.internal:55443"; got != want {
		t.Errorf("Host = %q, want %q", got, want)
	}
}

func TestExpandDefaultVarSetUsesValue(t *testing.T) {
	wfs := []Workflow{{Source: Side{MsgVPN: "${VPN:fallback}"}}}
	Expand(Expander{Lookup: lookupOf(map[string]string{"VPN": "prod"})}, &Env{}, wfs)
	if got, want := wfs[0].Source.MsgVPN, "prod"; got != want {
		t.Errorf("MsgVPN = %q, want %q", got, want)
	}
}

func TestExpandDefaultVarUnsetUsesDefault(t *testing.T) {
	wfs := []Workflow{{Source: Side{MsgVPN: "${VPN:fallback}"}}}
	Expand(Expander{Lookup: lookupOf(nil)}, &Env{}, wfs)
	if got, want := wfs[0].Source.MsgVPN, "fallback"; got != want {
		t.Errorf("MsgVPN = %q, want %q", got, want)
	}
}

func TestExpandUnsetNoDefaultPassesThroughWithWarning(t *testing.T) {
	wfs := []Workflow{{Source: Side{MsgVPN: "${TYPO}"}}}
	var warned []string
	Expand(Expander{
		Lookup: lookupOf(nil),
		Warn:   func(format string, a ...any) { warned = append(warned, fmt.Sprintf(format, a...)) },
	}, &Env{}, wfs)
	if got, want := wfs[0].Source.MsgVPN, "${TYPO}"; got != want {
		t.Errorf("MsgVPN = %q, want verbatim %q", got, want)
	}
	if len(warned) != 1 {
		t.Fatalf("warnings = %v, want exactly 1", warned)
	}
	if !strings.Contains(warned[0], "TYPO") {
		t.Errorf("warning %q should name the variable TYPO", warned[0])
	}
}

func TestExpandBareDollarVarUntouched(t *testing.T) {
	wfs := []Workflow{{Source: Side{MsgVPN: "$VPN"}}}
	Expand(Expander{Lookup: lookupOf(map[string]string{"VPN": "prod"})}, &Env{}, wfs)
	if got, want := wfs[0].Source.MsgVPN, "$VPN"; got != want {
		t.Errorf("bare $VAR must stay literal: MsgVPN = %q, want %q", got, want)
	}
}

func TestExpandCredentialFieldLeftAlone(t *testing.T) {
	wfs := []Workflow{{Source: Side{
		System:      SystemMQ,
		Password:    "${MQ_PASSWORD}",
		PasswordEnv: "${SOME_VAR}",
	}}}
	Expand(Expander{Lookup: lookupOf(map[string]string{"MQ_PASSWORD": "secret", "SOME_VAR": "x"})}, &Env{}, wfs)
	if got, want := wfs[0].Source.Password, "${MQ_PASSWORD}"; got != want {
		t.Errorf("credential Password must never expand: got %q, want %q", got, want)
	}
	if got, want := wfs[0].Source.PasswordEnv, "${SOME_VAR}"; got != want {
		t.Errorf("credential PasswordEnv must never expand: got %q, want %q", got, want)
	}
}

func TestExpandYAMLNodePassthroughLeftAlone(t *testing.T) {
	var doc yaml.Node
	if err := yaml.Unmarshal([]byte("KEY: ${SHOULD_NOT_EXPAND}\n"), &doc); err != nil {
		t.Fatal(err)
	}
	// api-properties is captured as a mapping node (see nodePtr); a document
	// node's real content is one level down under Content[0].
	mapping := &doc
	if doc.Kind == yaml.DocumentNode {
		mapping = doc.Content[0]
	}
	wfs := []Workflow{{Source: Side{APIProps: mapping}}}
	before := wfs[0].Source.APIProps.Content[1].Value
	Expand(Expander{Lookup: lookupOf(map[string]string{"SHOULD_NOT_EXPAND": "changed"})}, &Env{}, wfs)
	after := wfs[0].Source.APIProps.Content[1].Value
	if before != after || after != "${SHOULD_NOT_EXPAND}" {
		t.Errorf("yaml.Node passthrough must never expand: before=%q after=%q", before, after)
	}
}

func TestExpandDefaultsConnectionsMapEntry(t *testing.T) {
	e := &Env{Defaults: Defaults{Connections: map[string]Side{
		"edge": {Host: "tcps://${HOST}:55443"},
	}}}
	Expand(Expander{Lookup: lookupOf(map[string]string{"HOST": "broker.internal"})}, e, nil)
	if got, want := e.Connections["edge"].Host, "tcps://broker.internal:55443"; got != want {
		t.Errorf("Connections[edge].Host = %q, want %q", got, want)
	}
}

// TestExpandSecurityUserRole covers the one expansion consequence of User.Roles
// not being tagged expand:"no": a role is an identity like Name, not a
// credential, so a ${VAR} in it resolves. Proves Expand reaches into a []string
// field inside a slice of structs under Defaults.
func TestExpandSecurityUserRole(t *testing.T) {
	e := &Env{Defaults: Defaults{Security: Security{Users: []User{
		{Name: "ops", Roles: []string{"${ROLE}"}},
	}}}}
	Expand(Expander{Lookup: lookupOf(map[string]string{"ROLE": "admin"})}, e, nil)
	if got, want := e.Security.Users[0].Roles[0], "admin"; got != want {
		t.Errorf("role = %q, want the expanded %q", got, want)
	}
}

func TestExpandNilLookupDisablesEverything(t *testing.T) {
	wfs := []Workflow{{Source: Side{Host: "tcps://${HOST}:55443"}}}
	Expand(Expander{}, &Env{}, wfs)
	if got, want := wfs[0].Source.Host, "tcps://${HOST}:55443"; got != want {
		t.Errorf("nil Lookup must disable expansion entirely: got %q, want %q", got, want)
	}
}

// A store's source paths expand like any other path -- the existing PKCS12,
// every certs file, the key entry's files and the jks -- while its password
// fields and a certs alias stay as written: a password names a credential, and
// an alias is a name, not a path.
func TestExpandStorePathsNotPasswordsOrAliases(t *testing.T) {
	e := &Env{Defaults: Defaults{TLS: TLSConfig{Stores: map[string]*Store{
		"have": {PasswordEnv: "${PW}", Existing: &StoreExisting{PKCS12: "${D}/t.p12"}},
		"make": {Password: "${PW}", Create: &StoreSources{
			Certs: CertSources{Named: true, Items: []CertSource{{Name: "${D}", File: "${D}/ca.pem"}}},
			Keys:  []KeySource{{Cert: "${D}/c.pem", Chain: []string{"${D}/i.pem"}, Key: "${D}/c.key"}},
			JKS:   "${D}/old.jks",
		}},
	}}}}
	Expand(Expander{Lookup: lookupOf(map[string]string{"D": "/certs", "PW": "leaked"})}, e, nil)
	have, made := e.Defaults.TLS.Stores["have"], e.Defaults.TLS.Stores["make"]
	c := made.Create
	for _, got := range []string{have.Path(), c.Certs.Items[0].File, c.Keys[0].Cert, c.Keys[0].Chain[0], c.Keys[0].Key, c.JKS} {
		if !strings.HasPrefix(got, "/certs/") {
			t.Errorf("path %q was not expanded", got)
		}
	}
	if have.PasswordEnv != "${PW}" || made.Password != "${PW}" || c.Certs.Items[0].Name != "${D}" {
		t.Errorf("password-env %q, password %q, alias %q must stay as written", have.PasswordEnv, made.Password, c.Certs.Items[0].Name)
	}
}
