package tls

import (
	"strings"
	"testing"

	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/spec"
)

// placeholderSecretRef is the readable stand-in for consolidate's real
// storeSecret: it returns a placeholder named after the store, without caring
// whether the store's password is a literal or an -env reference.
func placeholderSecretRef(store string, c spec.Cred) string { return "${" + store + "}" }

// defaults builds Defaults holding the given named stores.
func defaults(stores map[string]*spec.Store) *spec.Defaults {
	return &spec.Defaults{TLS: spec.TLSConfig{Stores: stores}}
}

// existing is a store naming the PKCS12 at path.
func existing(path, passwordEnv string) *spec.Store {
	return &spec.Store{PasswordEnv: passwordEnv, Existing: &spec.StoreExisting{PKCS12: path}}
}

// keys lists the api-property keys SolaceProps returned, in order.
func keys(kvs []KV) string {
	var out []string
	for _, kv := range kvs {
		out = append(out, kv.Key)
	}
	return strings.Join(out, ",")
}

// Inside the container a store is named by its tls.stores name, so two stores
// never share a path -- not when their files share a name, nor when one file is
// defined under two names.
func TestMountPathIsPerStore(t *testing.T) {
	for _, c := range []struct{ store, want string }{
		{"mq-ca", MountDir + "/mq-ca"},
		{"solace-client", MountDir + "/solace-client"},
	} {
		if got := MountPath(c.store); got != c.want {
			t.Errorf("MountPath(%q) = %q, want %q", c.store, got, c.want)
		}
	}
}

// The store locations wired into Solace api-properties are the mounted names of
// the selected stores, whatever the files in env.yaml are called and however
// their paths are written -- here both stores share one file name, written with
// backslashes.
func TestSolacePropsUseStoreMountPaths(t *testing.T) {
	d := defaults(map[string]*spec.Store{
		"solace-ca":     existing(`.\trust\store.p12`, "TS_PW"),
		"solace-client": existing(`.\key\store.p12`, "KS_PW"),
	})
	got := map[string]string{}
	for _, kv := range SolaceProps(d, Selection{Truststore: "solace-ca", Keystore: "solace-client", KeyAlias: "c"}, placeholderSecretRef) {
		got[kv.Key] = kv.Val
	}
	want := map[string]string{
		"SSL_TRUST_STORE": MountDir + "/solace-ca",
		"SSL_KEY_STORE":   MountDir + "/solace-client",
	}
	for k, w := range want {
		if got[k] != w {
			t.Errorf("SolaceProps[%s] = %q, want %q", k, got[k], w)
		}
	}
}

// A created store has no file of its own: deploy builds it into the platform's
// secret, mounted under its name like any other, so that is the one path it
// is wired at.
func TestSolacePropsWireACreatedStoreAtItsMountPath(t *testing.T) {
	d := defaults(map[string]*spec.Store{
		"made": {PasswordEnv: "PW", Create: &spec.StoreSources{Certs: spec.CertSources{Items: []spec.CertSource{{File: "./ca.pem"}}}}},
	})
	got := map[string]string{}
	for _, kv := range SolaceProps(d, Selection{Truststore: "made"}, placeholderSecretRef) {
		got[kv.Key] = kv.Val
	}
	if got["SSL_TRUST_STORE"] != MountDir+"/made" || got["SSL_TRUST_STORE_FORMAT"] != "PKCS12" || got["SSL_TRUST_STORE_PASSWORD"] != "${made}" {
		t.Errorf("SolaceProps = %v", got)
	}
}

// The store passwords must reach api-properties as the placeholder secretRef
// returns -- never as the literal value or the name of the host env var backing
// it (S3) -- and secretRef must get each selected store's own name and
// credential, never the other's.
func TestSolacePropsStorePasswordIsPlaceholderNeverLiteral(t *testing.T) {
	d := defaults(map[string]*spec.Store{
		"solace-ca":     existing("./certs/truststore.p12", "TS_PW_ENV"),
		"solace-client": {Password: "s3cr3t-literal", Existing: &spec.StoreExisting{PKCS12: "./certs/keystore.p12"}},
	})
	var stores []string
	var creds []spec.Cred
	rec := func(store string, c spec.Cred) string {
		stores = append(stores, store)
		creds = append(creds, c)
		return "${" + store + "}"
	}
	got := map[string]string{}
	for _, kv := range SolaceProps(d, Selection{Truststore: "solace-ca", Keystore: "solace-client", KeyAlias: "c"}, rec) {
		got[kv.Key] = kv.Val
	}
	if got["SSL_TRUST_STORE_PASSWORD"] != "${solace-ca}" || got["SSL_KEY_STORE_PASSWORD"] != "${solace-client}" {
		t.Errorf("password placeholders = %q / %q", got["SSL_TRUST_STORE_PASSWORD"], got["SSL_KEY_STORE_PASSWORD"])
	}
	for _, v := range got {
		if strings.Contains(v, "TS_PW_ENV") || strings.Contains(v, "s3cr3t-literal") {
			t.Errorf("api-properties leaked a credential value/env name: %q", v)
		}
	}
	if strings.Join(stores, ",") != "solace-ca,solace-client" {
		t.Fatalf("secretRef stores = %v, want [solace-ca solace-client]", stores)
	}
	if creds[0] != (spec.Cred{EnvVar: "TS_PW_ENV"}) || creds[1] != (spec.Cred{Literal: "s3cr3t-literal"}) {
		t.Errorf("secretRef creds = %+v, want each store's own credential", creds)
	}
}

// Nothing selected -- or a name nothing defines -- writes nothing and records
// no password: validate reports the undefined name.
func TestSolacePropsSkipsSecretRefWhenNothingSelected(t *testing.T) {
	d := defaults(map[string]*spec.Store{"solace-ca": existing("./t.p12", "TS_PW")})
	calls := 0
	rec := func(store string, c spec.Cred) string { calls++; return "${" + store + "}" }
	for _, sel := range []Selection{{}, {Truststore: "nope", Keystore: "nope"}} {
		if got := SolaceProps(d, sel, rec); len(got) != 0 {
			t.Errorf("SolaceProps(%+v) = %v, want empty", sel, got)
		}
	}
	if calls != 0 {
		t.Errorf("secretRef called %d times, want 0", calls)
	}
}

// The tool wires stores and writes no validation policy: SSL_VALIDATE_CERTIFICATE
// is the Solace API's default (true) or the connection's own api-properties, so
// no selection may emit it -- an operator's false would otherwise be dropped as
// a collision with the tool's true. A keystore alone presents its certificate;
// the alias is written only when one picks a key.
func TestSolacePropsWritesNoValidationPolicy(t *testing.T) {
	d := defaults(map[string]*spec.Store{
		"ca":     existing("./t.p12", "TS_PW"),
		"client": existing("./k.p12", "KS_PW"),
	})
	for _, c := range []struct {
		name string
		sel  Selection
		want string
	}{
		{"none", Selection{}, ""},
		{"truststore", Selection{Truststore: "ca"}, "SSL_TRUST_STORE,SSL_TRUST_STORE_PASSWORD,SSL_TRUST_STORE_FORMAT"},
		{"keystore without alias", Selection{Keystore: "client"}, "SSL_KEY_STORE,SSL_KEY_STORE_PASSWORD,SSL_KEY_STORE_FORMAT"},
		{"keystore with alias", Selection{Keystore: "client", KeyAlias: "a"}, "SSL_KEY_STORE,SSL_KEY_STORE_PASSWORD,SSL_KEY_STORE_FORMAT,SSL_PRIVATE_KEY_ALIAS"},
		{"both with alias", Selection{Truststore: "ca", Keystore: "client", KeyAlias: "a"}, "SSL_TRUST_STORE,SSL_TRUST_STORE_PASSWORD,SSL_TRUST_STORE_FORMAT,SSL_KEY_STORE,SSL_KEY_STORE_PASSWORD,SSL_KEY_STORE_FORMAT,SSL_PRIVATE_KEY_ALIAS"},
	} {
		t.Run(c.name, func(t *testing.T) {
			got := SolaceProps(d, c.sel, placeholderSecretRef)
			for _, kv := range got {
				if kv.Key == "SSL_VALIDATE_CERTIFICATE" {
					t.Errorf("SolaceProps wrote validation policy %v", kv)
				}
			}
			if keys(got) != c.want {
				t.Errorf("keys = %s, want %s", keys(got), c.want)
			}
		})
	}
}

// Every store is PKCS12 -- created ones are built as one, existing ones must
// be one -- so both FORMAT keys always say so, never left to JCSMP's default.
func TestSolacePropsAlwaysWritePKCS12(t *testing.T) {
	d := defaults(map[string]*spec.Store{
		"ca":     existing("./t.p12", "TS_PW"),
		"client": existing("./k.p12", "KS_PW"),
	})
	got := map[string]string{}
	for _, kv := range SolaceProps(d, Selection{Truststore: "ca", Keystore: "client"}, placeholderSecretRef) {
		got[kv.Key] = kv.Val
	}
	if got["SSL_TRUST_STORE_FORMAT"] != "PKCS12" || got["SSL_KEY_STORE_FORMAT"] != "PKCS12" {
		t.Errorf("formats = %q / %q, want PKCS12", got["SSL_TRUST_STORE_FORMAT"], got["SSL_KEY_STORE_FORMAT"])
	}
}

// A selected store that is not configured -- no source, or an existing block
// without a path -- is not wired (validate refuses it), and its password
// position is never recorded.
func TestSolacePropsSkipsUnconfiguredStore(t *testing.T) {
	d := defaults(map[string]*spec.Store{
		"ca":     {PasswordEnv: "TS_PW"},
		"client": {PasswordEnv: "KS_PW", Existing: &spec.StoreExisting{}},
	})
	calls := 0
	rec := func(store string, c spec.Cred) string { calls++; return "${" + store + "}" }
	if got := SolaceProps(d, Selection{Truststore: "ca", Keystore: "client", KeyAlias: "a"}, rec); len(got) != 0 {
		t.Errorf("SolaceProps with unconfigured stores = %v, want empty", got)
	}
	if calls != 0 {
		t.Errorf("secretRef called %d times, want 0", calls)
	}
}
