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

// MountPath must strip both '/' and '\' independent of host OS, so a Windows-
// authored env.yaml path (which may use backslashes) resolves to the same
// mounted base name even when the CLI runs on Linux/Mac — the base name the
// stores Secret mounts, not left unsplit.
func TestMountPathSeparatorAgnostic(t *testing.T) {
	cases := map[string]string{
		"truststore.jks":              MountDir + "/truststore.jks",
		"./certs/truststore.jks":      MountDir + "/truststore.jks",
		`.\certs\truststore.jks`:      MountDir + "/truststore.jks",
		`C:\app\certs\keystore.jks`:   MountDir + "/keystore.jks",
		"/abs/unix/path/keystore.jks": MountDir + "/keystore.jks",
		`mixed/dir\store.jks`:         MountDir + "/store.jks",
	}
	for in, want := range cases {
		if got := MountPath(in); got != want {
			t.Errorf("MountPath(%q) = %q, want %q", in, got, want)
		}
	}
}

// The store locations wired into Solace api-properties must be the mounted base
// name, not the raw (possibly backslash) path from env.yaml — otherwise the
// bundle/location references a file the volume never mounts.
func TestSolacePropsUseMountedBaseName(t *testing.T) {
	d := &spec.Defaults{TLS: spec.TLSConfig{
		Truststore: &spec.Store{File: `.\certs\truststore.jks`, PasswordEnv: "TRUSTSTORE_PASSWORD_ENV", Type: "JKS"},
		Keystore:   &spec.Store{File: `.\certs\keystore.jks`, PasswordEnv: "KEYSTORE_PASSWORD_ENV", Type: "JKS"},
	}}
	got := map[string]string{}
	for _, kv := range SolaceProps(d, "solace-client", true, placeholderSecretRef) {
		got[kv.Key] = kv.Val
	}
	want := map[string]string{
		"SSL_TRUST_STORE": MountDir + "/truststore.jks",
		"SSL_KEY_STORE":   MountDir + "/keystore.jks",
	}
	for k, w := range want {
		if got[k] != w {
			t.Errorf("SolaceProps[%s] = %q, want %q", k, got[k], w)
		}
	}
}

func TestStorePathConfigVsDeploy(t *testing.T) {
	if StorePath("./certs/t.jks", false) != "./certs/t.jks" {
		t.Error("mount=false must return the raw defaults path (config)")
	}
	if got := StorePath(`.\certs\t.jks`, true); got != MountDir+"/t.jks" {
		t.Errorf("mount=true must return the mount path base name, got %q", got)
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
