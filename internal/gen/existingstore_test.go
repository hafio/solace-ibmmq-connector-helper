package gen

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/certfixture"
	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/consolidate"
	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/spec"
)

// checkExistingStore opens an existing store before the connector does: a
// file that is not a PKCS12 -- a JKS, a JCEKS, a PEM, a certificate -- is an
// error with or without the password; with it, a store it does not open is an
// error, and one it opens but cannot read, or a truststore without Java's
// trust attribute, only warns. With no password to use only the shape is
// checked.
func TestCheckExistingStore(t *testing.T) {
	ca := certfixture.CA(t, "Root")
	leaf := certfixture.Issue(t, ca, "client")
	trust := certfixture.TrustStorePKCS12(t, storePW, ca)
	key := certfixture.KeyStorePKCS12(t, storePW, leaf, ca)
	jks := certfixture.JKS(t, storePW, []certfixture.Trusted{{Alias: "root", Cert: ca}}, nil)
	ts := consolidate.StoreUse{Name: "ts", File: "./t.p12", Trusted: true}
	ks := consolidate.StoreUse{Name: "ks", File: "./k.p12", Private: true}
	env := spec.Cred{EnvVar: "CA_PW"}
	res := Resolver{Env: storeEnv}
	for _, c := range []struct {
		name string
		use  consolidate.StoreUse
		data []byte
		pw   spec.Cred
		err  string
		warn string
	}{
		{"a truststore", ts, trust, env, "", ""},
		{"a keystore", ks, key, env, "", ""},
		{"a store used both ways", consolidate.StoreUse{Name: "ks", File: "./k.p12", Trusted: true, Private: true}, key, env, "", ""},
		{"the wrong password", ts, trust, spec.Cred{EnvVar: "WRONG"}, `tls.stores.ts.existing.pkcs12 "./t.p12" does not open with the value of WRONG (tls.stores.ts.password-env)`, ""},
		{"the wrong literal", ks, key, spec.Cred{Literal: "nope"}, `tls.stores.ks.existing.pkcs12 "./k.p12" does not open with tls.stores.ks.password`, ""},
		{"a JKS", ts, jks, env, "is a JKS keystore, not a PKCS12", ""},
		{"a JKS, no password", ts, jks, spec.Cred{EnvVar: "NOPE"}, "is a JKS keystore, not a PKCS12", ""},
		{"a JCEKS", ts, []byte{0xCE, 0xCE, 0xCE, 0xCE, 0, 0}, env, "is a JCEKS keystore, not a PKCS12", ""},
		{"a PEM", ts, certfixture.PEM(ca), env, `"./t.p12" is not a PKCS12 file`, ""},
		{"a certificate renamed .p12", ts, certfixture.DER(ca), spec.Cred{EnvVar: "NOPE"}, "is not a PKCS12 file", ""},
		{"no password to open it", ts, trust, spec.Cred{EnvVar: "NOPE"}, "", ""},
		{"an empty password variable", ts, trust, spec.Cred{EnvVar: "EMPTY"}, "", ""},
		{"a keystore read as a truststore", ts, key, env, "", "opens with its password, but this tool cannot read its contents"},
	} {
		t.Run(c.name, func(t *testing.T) {
			r := res
			if c.pw.EnvVar == "WRONG" {
				r.Env = func(string) (string, bool) { return "not-it", true }
			}
			warns, err := checkExistingStore(c.use, c.data, c.pw, r)
			switch {
			case c.err == "" && err != nil:
				t.Fatalf("unexpected error: %v", err)
			case c.err != "" && (err == nil || !strings.Contains(err.Error(), c.err)):
				t.Fatalf("err = %v, want %q", err, c.err)
			}
			if c.warn == "" && len(warns) > 0 || c.warn != "" && !hasIssue(warns, c.warn) {
				t.Errorf("warns = %v, want %q", warns, c.warn)
			}
		})
	}
}

// Stores openssl makes: a keystore opens as made; a truststore made with
// -nokeys carries no Java trust attribute before OpenSSL 3.2 or without
// -jdktrust, which Java loads as empty -- a warning naming the fix. Skipped
// where openssl is not on PATH, like the keytool check in storebuild.
func TestCheckExistingStoreFromOpenSSL(t *testing.T) {
	openssl, err := exec.LookPath("openssl")
	if err != nil {
		t.Skip("openssl is not on PATH")
	}
	dir := t.TempDir()
	ca := certfixture.CA(t, "Root")
	leaf := certfixture.Issue(t, ca, "client")
	write := func(name string, b []byte) string {
		p := filepath.Join(dir, name)
		if err := os.WriteFile(p, b, 0o600); err != nil {
			t.Fatal(err)
		}
		return p
	}
	caPEM, leafPEM, keyPEM := write("ca.pem", certfixture.PEM(ca)), write("leaf.pem", certfixture.PEM(leaf)), write("leaf.key", certfixture.KeyPKCS8(t, leaf))
	run := func(out string, args ...string) []byte {
		cmd := exec.Command(openssl, append([]string{"pkcs12", "-export", "-out", filepath.Join(dir, out), "-passout", "pass:" + storePW}, args...)...)
		if b, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("openssl %v: %v\n%s", args, err, b)
		}
		b, err := os.ReadFile(filepath.Join(dir, out))
		if err != nil {
			t.Fatal(err)
		}
		return b
	}
	env := spec.Cred{EnvVar: "CA_PW"}
	res := Resolver{Env: storeEnv}

	keystore := run("k.p12", "-in", leafPEM, "-inkey", keyPEM, "-certfile", caPEM)
	if w, err := checkExistingStore(consolidate.StoreUse{Name: "ks", File: "./k.p12", Private: true}, keystore, env, res); err != nil || len(w) != 0 {
		t.Errorf("openssl keystore: err %v, warns %v", err, w)
	}
	nokeys := run("t.p12", "-nokeys", "-in", caPEM)
	w, err := checkExistingStore(consolidate.StoreUse{Name: "ts", File: "./t.p12", Trusted: true}, nokeys, env, res)
	if err != nil {
		t.Fatalf("openssl -nokeys truststore: %v", err)
	}
	if !hasIssue(w, "without Java's trusted-certificate attribute") {
		t.Errorf("an openssl -nokeys truststore must warn, got %v", w)
	}
}

// ResolvePodmanFiles opens an existing store before loading it, as
// kubernetes' resolveStores does.
func TestResolvePodmanFilesOpensExistingStores(t *testing.T) {
	ca := certfixture.CA(t, "Root")
	trust := certfixture.TrustStorePKCS12(t, storePW, ca)
	f := PodmanFile{StoreName: "c-tls-ts", Source: "./t.p12", Field: "tls.stores.ts.existing.pkcs12",
		Use: consolidate.StoreUse{Name: "ts", File: "./t.p12", Trusted: true}, Password: spec.Cred{EnvVar: "CA_PW"}}
	res := Resolver{Env: storeEnv, ReadFile: storeFiles(map[string][]byte{"./t.p12": trust})}
	if kvs, _, err := ResolvePodmanFiles([]PodmanFile{f}, res); err != nil || len(kvs) != 1 {
		t.Fatalf("a good store loads: %v %v", kvs, err)
	}
	res.Env = func(string) (string, bool) { return "not-it", true }
	if _, _, err := ResolvePodmanFiles([]PodmanFile{f}, res); err == nil || !strings.Contains(err.Error(), "does not open with the value of CA_PW") {
		t.Errorf("a store the password does not open stops the deploy, got %v", err)
	}
}
