package tls

import (
	"strings"
	"testing"

	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/spec"
)

// placeholderSecretRef is the readable stand-in for consolidate's real
// secretRef: it returns the stable placeholder for any credential, without
// caring whether the caller supplied a literal or an -env reference.
func placeholderSecretRef(n string, c spec.Cred) string { return "${" + n + "}" }

// Inside the container a store is named by its role alone, so the two stores
// never share a path -- not when their files share a name, nor when one file
// serves as both.
func TestMountPathIsFixedPerRole(t *testing.T) {
	for _, c := range []struct {
		role Role
		want string
	}{
		{Truststore, MountDir + "/truststore"},
		{Keystore, MountDir + "/keystore"},
	} {
		if got := MountPath(c.role); got != c.want {
			t.Errorf("MountPath(%q) = %q, want %q", c.role, got, c.want)
		}
	}
}

// The store locations wired into Solace api-properties must be the fixed
// mounted names, whatever the files in env.yaml are called and however their
// paths are written -- here both stores share one file name, written with
// backslashes.
func TestSolacePropsUseFixedMountPaths(t *testing.T) {
	d := &spec.Defaults{TLS: spec.TLSConfig{
		Truststore: &spec.Store{File: `.\trust\store.p12`, PasswordEnv: "TRUSTSTORE_PASSWORD_ENV", Type: "PKCS12"},
		Keystore:   &spec.Store{File: `.\key\store.p12`, PasswordEnv: "KEYSTORE_PASSWORD_ENV", Type: "PKCS12"},
	}}
	got := map[string]string{}
	for _, kv := range SolaceProps(d, "solace-client", true, placeholderSecretRef) {
		got[kv.Key] = kv.Val
	}
	want := map[string]string{
		"SSL_TRUST_STORE": MountDir + "/truststore",
		"SSL_KEY_STORE":   MountDir + "/keystore",
	}
	for k, w := range want {
		if got[k] != w {
			t.Errorf("SolaceProps[%s] = %q, want %q", k, got[k], w)
		}
	}
}

func TestStorePathConfigVsDeploy(t *testing.T) {
	for _, c := range []struct {
		role  Role
		file  string
		mount bool
		want  string
	}{
		{Truststore, "./certs/t.jks", false, "./certs/t.jks"},
		{Keystore, `.\certs\k.p12`, false, `.\certs\k.p12`},
		{Truststore, `.\certs\t.jks`, true, MountDir + "/truststore"},
		{Keystore, "./certs/t.jks", true, MountDir + "/keystore"},
	} {
		if got := StorePath(c.role, c.file, c.mount); got != c.want {
			t.Errorf("StorePath(%q, %q, %v) = %q, want %q", c.role, c.file, c.mount, got, c.want)
		}
	}
}

func TestSolacePropsRawPathWhenNotMounted(t *testing.T) {
	d := &spec.Defaults{TLS: spec.TLSConfig{
		Truststore: &spec.Store{File: "./certs/truststore.jks", PasswordEnv: "TRUSTSTORE_PASSWORD_ENV", Type: "JKS"},
	}}
	got := map[string]string{}
	for _, kv := range SolaceProps(d, "", false, placeholderSecretRef) {
		got[kv.Key] = kv.Val
	}
	if got["SSL_TRUST_STORE"] != "./certs/truststore.jks" {
		t.Errorf("mount=false should keep the raw defaults path, got %q", got["SSL_TRUST_STORE"])
	}
}

// The store passwords must reach api-properties as their mount-name placeholder
// (the two package constants, which carry spec.GeneratedNamePrefix) -- never as
// the literal
// value or the name of the host env var backing it (S3: no credential value
// or host variable name may reach the rendered config).
func TestSolacePropsStorePasswordIsStablePlaceholderNeverLiteral(t *testing.T) {
	d := &spec.Defaults{TLS: spec.TLSConfig{
		Truststore: &spec.Store{File: "./certs/truststore.jks", PasswordEnv: "TRUSTSTORE_PASSWORD_ENV", Type: "JKS"},
		Keystore:   &spec.Store{File: "./certs/keystore.jks", Password: "s3cr3t-literal", Type: "JKS"},
	}}

	var stables []string
	var creds []spec.Cred
	rec := func(stable string, c spec.Cred) string {
		stables = append(stables, stable)
		creds = append(creds, c)
		return "${" + stable + "}"
	}

	got := map[string]string{}
	for _, kv := range SolaceProps(d, "solace-client", false, rec) {
		got[kv.Key] = kv.Val
	}

	if want := "${" + TruststorePasswordName + "}"; got["SSL_TRUST_STORE_PASSWORD"] != want {
		t.Errorf("SSL_TRUST_STORE_PASSWORD = %q, want %q", got["SSL_TRUST_STORE_PASSWORD"], want)
	}
	if want := "${" + KeystorePasswordName + "}"; got["SSL_KEY_STORE_PASSWORD"] != want {
		t.Errorf("SSL_KEY_STORE_PASSWORD = %q, want %q", got["SSL_KEY_STORE_PASSWORD"], want)
	}
	for _, v := range got {
		if strings.Contains(v, "TRUSTSTORE_PASSWORD_ENV") || strings.Contains(v, "s3cr3t-literal") {
			t.Errorf("api-properties leaked a credential value/env name: %q", v)
		}
	}

	// The stable names passed to secretRef must be the package's own constants,
	// paired with the store's actual Secret() (never the other store's, never
	// swapped).
	if len(stables) != 2 || stables[0] != TruststorePasswordName || stables[1] != KeystorePasswordName {
		t.Fatalf("secretRef stable names = %v, want [%s %s]", stables, TruststorePasswordName, KeystorePasswordName)
	}
	if creds[0] != (spec.Cred{EnvVar: "TRUSTSTORE_PASSWORD_ENV"}) {
		t.Errorf("truststore secretRef got Cred %+v, want the truststore's -env credential", creds[0])
	}
	if creds[1] != (spec.Cred{Literal: "s3cr3t-literal"}) {
		t.Errorf("keystore secretRef got Cred %+v, want the keystore's literal credential", creds[1])
	}
}

// Missing stores must not invoke secretRef at all: no store, no password
// position, so nothing should be recorded for it.
func TestSolacePropsSkipsSecretRefWhenStoreMissing(t *testing.T) {
	d := &spec.Defaults{} // no truststore, no keystore
	calls := 0
	rec := func(stable string, c spec.Cred) string {
		calls++
		return "${" + stable + "}"
	}
	if got := SolaceProps(d, "solace-client", false, rec); len(got) != 0 {
		t.Errorf("SolaceProps with no stores = %v, want empty", got)
	}
	if calls != 0 {
		t.Errorf("secretRef called %d times, want 0", calls)
	}
}
