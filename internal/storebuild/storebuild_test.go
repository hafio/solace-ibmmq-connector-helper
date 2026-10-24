package storebuild

import (
	"bytes"
	"crypto/ecdh"
	"crypto/rand"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	pkcs12 "software.sslmate.com/src/go-pkcs12"

	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/certfixture"
)

const pw = "store-pass-123"

func src(field, path string, data []byte) Source {
	return Source{Field: field, Path: path, Data: data}
}

// aliases lists the aliases a build reports, in order.
func aliases(res Result) string {
	var out []string
	for _, e := range res.Entries {
		out = append(out, e.Alias)
	}
	return strings.Join(out, ",")
}

func hasWarning(res Result, sub string) bool {
	for _, w := range res.Warnings {
		if strings.Contains(w, sub) {
			return true
		}
	}
	return false
}

// A certs list numbers every certificate of every file in order, PEM and DER
// alike, and the result is a Java truststore: DecodeTrustStore accepts only
// certificates carrying Java's trusted-cert attribute.
func TestTruststoreFromCertsList(t *testing.T) {
	a, b, c := certfixture.CA(t, "Root A"), certfixture.CA(t, "Root B"), certfixture.CA(t, "Root C")
	res, err := Build(Input{Store: "ca", Password: pw, Certs: []CertSource{
		{Source: src("create.certs[0]", "./ab.pem", certfixture.PEM(a, b))},
		{Source: src("create.certs[1]", "./c.der", certfixture.DER(c))},
	}})
	if err != nil {
		t.Fatal(err)
	}
	if res.Keystore || aliases(res) != "ca-1,ca-2,ca-3" || len(res.Warnings) != 0 {
		t.Fatalf("keystore=%v aliases=%s warnings=%v", res.Keystore, aliases(res), res.Warnings)
	}
	got, err := pkcs12.DecodeTrustStore(res.Data, pw)
	if err != nil {
		t.Fatal(err)
	}
	if !sameCerts(got, []*x509.Certificate{a.Cert, b.Cert, c.Cert}) {
		t.Errorf("decoded truststore does not hold the three certificates in order")
	}
}

// A named map aliases each certificate by its name; a file holding several
// certificates under one name numbers the rest, and Java's lower-casing shows.
func TestTruststoreNamedAliases(t *testing.T) {
	a, b, c := certfixture.CA(t, "Root A"), certfixture.CA(t, "Root B"), certfixture.CA(t, "Issuing")
	res, err := Build(Input{Store: "ca", Password: pw, Certs: []CertSource{
		{Alias: "roots", Source: src("create.certs.roots", "./roots.pem", certfixture.PEM(a, b))},
		{Alias: "Issuing", Source: src("create.certs.Issuing", "./issuing.pem", certfixture.PEM(c))},
	}})
	if err != nil {
		t.Fatal(err)
	}
	if aliases(res) != "roots,roots-2,issuing" {
		t.Errorf("aliases = %s, want roots,roots-2,issuing", aliases(res))
	}
}

// Aliases ignore case in Java, so two names differing only in case name one
// entry -- a build error naming both sources rather than a silently lost
// certificate.
func TestTruststoreAliasCollision(t *testing.T) {
	a, b := certfixture.CA(t, "Root A"), certfixture.CA(t, "Root B")
	_, err := Build(Input{Store: "ca", Password: pw, Certs: []CertSource{
		{Alias: "Root", Source: src("create.certs.Root", "./a.pem", certfixture.PEM(a))},
		{Alias: "root", Source: src("create.certs.root", "./b.pem", certfixture.PEM(b))},
	}})
	if err == nil || !strings.Contains(err.Error(), `alias "root" names two certificates`) || !strings.Contains(err.Error(), "./a.pem") || !strings.Contains(err.Error(), "./b.pem") {
		t.Fatalf("err = %v", err)
	}
}

func TestTruststoreWarnings(t *testing.T) {
	now := time.Now()
	ca := certfixture.CA(t, "Root")
	for _, c := range []struct {
		name  string
		data  []byte
		want  string
		count int
	}{
		{"a self-signed leaf is pinned", certfixture.PEM(certfixture.CA(t, "broker", certfixture.NotCA())), "is not a CA certificate; it is trusted as-is", 1},
		{"expired", certfixture.PEM(certfixture.CA(t, "Old", certfixture.Validity(now.Add(-48*time.Hour), now.Add(-24*time.Hour)))), "expired on", 1},
		{"expiring", certfixture.PEM(certfixture.CA(t, "Soon", certfixture.Validity(now.Add(-time.Hour), now.Add(10*24*time.Hour)))), "expires on", 1},
		{"repeated", certfixture.PEM(ca, ca), "repeats a certificate", 1},
		{"a key beside the certificates", append(certfixture.PEM(ca), certfixture.KeyPKCS8(t, ca)...), "also holds a private key, which a truststore leaves out", 1},
	} {
		t.Run(c.name, func(t *testing.T) {
			res, err := Build(Input{Store: "ca", Password: pw, Now: now, Certs: []CertSource{{Source: src("create.certs[0]", "./ca.pem", c.data)}}})
			if err != nil {
				t.Fatal(err)
			}
			if !hasWarning(res, c.want) || len(res.Entries) != c.count {
				t.Errorf("warnings = %v entries = %d, want %q and %d entries", res.Warnings, len(res.Entries), c.want, c.count)
			}
		})
	}
}

func TestTruststoreRejects(t *testing.T) {
	ca := certfixture.CA(t, "Root")
	for _, c := range []struct {
		name  string
		certs []CertSource
		want  string
	}{
		{"not a certificate", []CertSource{{Source: src("create.certs[0]", "./junk.txt", []byte("hello"))}}, "tls.stores.ca.create.certs[0] (./junk.txt) holds no certificate"},
		{"only a key", []CertSource{{Source: src("create.certs[0]", "./key.pem", certfixture.KeyPKCS8(t, ca))}}, "holds no certificate"},
		{"a broken certificate", []CertSource{{Source: src("create.certs[0]", "./bad.pem", pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: []byte("junk")}))}}, "certificate 1 does not parse"},
		{"no source", nil, "tls.stores.ca holds no certificate to trust"},
	} {
		t.Run(c.name, func(t *testing.T) {
			_, err := Build(Input{Store: "ca", Password: pw, Certs: c.certs})
			if err == nil || !strings.Contains(err.Error(), c.want) {
				t.Errorf("err = %v, want %q", err, c.want)
			}
		})
	}
}

// chain mints root -> intermediate -> leaf.
func chain(t *testing.T, opts ...certfixture.Option) (root, inter, leaf certfixture.Cert) {
	root = certfixture.CA(t, "Root")
	inter = certfixture.Issue(t, root, "Issuing", certfixture.AsCA())
	leaf = certfixture.Issue(t, inter, "connector", opts...)
	return root, inter, leaf
}

// The leaf is found by its key wherever it sits, the chain is ordered by
// issuer whatever order the files list it in, and Java will list the one key
// as alias 1.
func TestKeystoreFromPEMChain(t *testing.T) {
	root, inter, leaf := chain(t)
	res, err := Build(Input{Store: "client", Password: pw, Keys: []KeySource{{
		Cert:  src("create.keys[0].cert", "./client.pem", certfixture.PEM(inter, leaf)),
		Chain: []Source{src("create.keys[0].chain[0]", "./root.pem", certfixture.PEM(root))},
		Key:   src("create.keys[0].key", "./client.key", certfixture.KeyPKCS8(t, leaf)),
	}}})
	if err != nil {
		t.Fatal(err)
	}
	if !res.Keystore || aliases(res) != "1,," || !res.Entries[0].Key || len(res.Warnings) != 0 {
		t.Fatalf("keystore=%v aliases=%q entries=%+v warnings=%v", res.Keystore, aliases(res), res.Entries, res.Warnings)
	}
	key, gotLeaf, gotChain, err := pkcs12.DecodeChain(res.Data, pw)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(gotLeaf.Raw, leaf.Cert.Raw) || !sameCerts(gotChain, []*x509.Certificate{inter.Cert, root.Cert}) || !matchesKey(key, leaf.Key) {
		t.Errorf("decoded keystore does not hold the leaf, [issuing, root] and the key")
	}
}

// Every unencrypted key encoding a CA or openssl hands out builds.
func TestKeystoreKeyEncodings(t *testing.T) {
	ec := certfixture.CA(t, "ec", certfixture.NotCA())
	rsaCert := certfixture.CA(t, "rsa", certfixture.NotCA(), certfixture.RSAKey())
	for _, c := range []struct {
		name string
		cert certfixture.Cert
		key  []byte
	}{
		{"PKCS#8 PEM", ec, certfixture.KeyPKCS8(t, ec)},
		{"SEC1 PEM", ec, certfixture.KeyEC(t, ec)},
		{"PKCS#1 PEM", rsaCert, certfixture.KeyPKCS1(t, rsaCert)},
		{"PKCS#8 DER", ec, certfixture.KeyPKCS8DER(t, ec)},
		{"PKCS#1 DER", rsaCert, derOf(certfixture.KeyPKCS1(t, rsaCert))},
		{"SEC1 DER", ec, derOf(certfixture.KeyEC(t, ec))},
		{"one file with the certificate and the key", ec, append(certfixture.PEM(ec), certfixture.KeyPKCS8(t, ec)...)},
	} {
		t.Run(c.name, func(t *testing.T) {
			_, err := Build(Input{Store: "client", Password: pw, Keys: []KeySource{{
				Cert: src("create.keys[0].cert", "./c.pem", certfixture.PEM(c.cert)),
				Key:  src("create.keys[0].key", "./c.key", c.key),
			}}})
			if err != nil {
				t.Error(err)
			}
		})
	}
}

func TestKeystoreRejects(t *testing.T) {
	now := time.Now()
	_, _, leaf := chain(t)
	other := certfixture.CA(t, "other", certfixture.NotCA())
	expired := certfixture.CA(t, "old", certfixture.NotCA(), certfixture.Validity(now.Add(-48*time.Hour), now.Add(-time.Hour)))
	encrypted := pem.EncodeToMemory(&pem.Block{Type: "ENCRYPTED PRIVATE KEY", Bytes: []byte("x")})
	legacy := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Headers: map[string]string{"Proc-Type": "4,ENCRYPTED", "DEK-Info": "AES-256-CBC,00"}, Bytes: []byte("x")})
	for _, c := range []struct {
		name string
		cert certfixture.Cert
		key  []byte
		want string
	}{
		{"encrypted PKCS#8", leaf, encrypted, "holds an encrypted private key, which cannot be read here"},
		{"legacy encrypted PEM", leaf, legacy, "holds an encrypted private key"},
		{"two keys", leaf, append(certfixture.KeyPKCS8(t, leaf), certfixture.KeyPKCS8(t, other)...), "holds more than one private key"},
		{"no key", leaf, certfixture.PEM(leaf), "holds no private key"},
		{"a broken key", leaf, pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: []byte("junk")}), "the private key does not parse"},
		{"a key of another certificate", leaf, certfixture.KeyPKCS8(t, other), "no certificate in tls.stores.client.create.keys[0].cert (./c.pem) matches the private key in tls.stores.client.create.keys[0].key (./c.key)"},
		{"an expired certificate", expired, certfixture.KeyPKCS8(t, expired), "expired on"},
		{"an X25519 key", leaf, x25519Key(t), "a private key of type *ecdh.PrivateKey is not supported"},
	} {
		t.Run(c.name, func(t *testing.T) {
			_, err := Build(Input{Store: "client", Password: pw, Now: now, Keys: []KeySource{{
				Cert: src("create.keys[0].cert", "./c.pem", certfixture.PEM(c.cert)),
				Key:  src("create.keys[0].key", "./c.key", c.key),
			}}})
			if err == nil || !strings.Contains(err.Error(), c.want) {
				t.Errorf("err = %v, want %q", err, c.want)
			}
		})
	}
}

// A certificate outside the leaf's chain is left out with a warning, and a
// leaf that is not valid yet or expires soon is reported but still built.
func TestKeystoreWarnings(t *testing.T) {
	now := time.Now()
	root, inter, leaf := chain(t, certfixture.Validity(now.Add(time.Hour), now.Add(10*24*time.Hour)))
	stranger := certfixture.CA(t, "Stranger")
	res, err := Build(Input{Store: "client", Password: pw, Now: now, Keys: []KeySource{{
		Cert: src("create.keys[0].cert", "./c.pem", certfixture.PEM(leaf, inter, root, stranger)),
		Key:  src("create.keys[0].key", "./c.key", certfixture.KeyPKCS8(t, leaf)),
	}}})
	if err != nil {
		t.Fatal(err)
	}
	for _, w := range []string{"CN=Stranger is not part of the chain of CN=connector and is left out", "is not valid until"} {
		if !hasWarning(res, w) {
			t.Errorf("warnings = %v, want %q", res.Warnings, w)
		}
	}
	if _, _, gotChain, err := pkcs12.DecodeChain(res.Data, pw); err != nil || len(gotChain) != 2 {
		t.Errorf("chain = %d certificates, err %v; want issuing and root only", len(gotChain), err)
	}
}

// A JKS of trusted certificates becomes a truststore keeping its aliases.
func TestJKSTrustedEntriesBecomeATruststore(t *testing.T) {
	a, b := certfixture.CA(t, "Root A"), certfixture.CA(t, "Root B")
	jks := certfixture.JKS(t, pw, []certfixture.Trusted{{Alias: "beta", Cert: b}, {Alias: "alpha", Cert: a}}, nil)
	res, err := Build(Input{Store: "mq-ca", Password: pw, JKS: &Source{Field: "create.jks", Path: "./old.jks", Data: jks}})
	if err != nil {
		t.Fatal(err)
	}
	if res.Keystore || aliases(res) != "alpha,beta" {
		t.Errorf("keystore=%v aliases=%s, want a truststore aliased alpha,beta", res.Keystore, aliases(res))
	}
}

// A JKS holding exactly one key and nothing else becomes a keystore.
func TestJKSSingleKeyBecomesAKeystore(t *testing.T) {
	root, inter, leaf := chain(t)
	jks := certfixture.JKS(t, pw, nil, []certfixture.Key{{Alias: "client", Chain: []certfixture.Cert{leaf, inter, root}}})
	res, err := Build(Input{Store: "mq-client", Password: pw, JKS: &Source{Field: "create.jks", Path: "./old.jks", Data: jks}})
	if err != nil {
		t.Fatal(err)
	}
	_, gotLeaf, gotChain, err := pkcs12.DecodeChain(res.Data, pw)
	if !res.Keystore || err != nil || !bytes.Equal(gotLeaf.Raw, leaf.Cert.Raw) || len(gotChain) != 2 {
		t.Errorf("keystore=%v err=%v; want the leaf and a two-certificate chain", res.Keystore, err)
	}
}

func TestJKSRejects(t *testing.T) {
	_, _, leaf := chain(t)
	ca := certfixture.CA(t, "Root")
	key := certfixture.Key{Alias: "client", Chain: []certfixture.Cert{leaf}}
	built, err := Build(Input{Store: "x", Password: pw, Certs: []CertSource{{Source: src("create.certs[0]", "./ca.pem", certfixture.PEM(ca))}}})
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range []struct {
		name string
		data []byte
		want string
	}{
		{"trusted and a key", certfixture.JKS(t, pw, []certfixture.Trusted{{Alias: "ca", Cert: ca}}, []certfixture.Key{key}), `holds 1 trusted certificates and 1 private keys -- "ca" (trusted), "client" (key)`},
		{"two keys", certfixture.JKS(t, pw, nil, []certfixture.Key{key, {Alias: "second", Chain: []certfixture.Cert{leaf}}}), "holds 0 trusted certificates and 2 private keys"},
		{"empty", certfixture.JKS(t, pw, nil, nil), "holds no entries"},
		{"another password", certfixture.JKS(t, "other-pass", []certfixture.Trusted{{Alias: "ca", Cert: ca}}, nil), "does not open with the password of tls.stores.mq (the JKS integrity check failed)"},
		{"a key with its own password", certfixture.JKS(t, pw, nil, []certfixture.Key{{Alias: "client", Chain: []certfixture.Cert{leaf}, Password: "key-pass"}}), `the key "client" does not open with the store password`},
		{"JCEKS", []byte{0xCE, 0xCE, 0xCE, 0xCE, 0, 0, 0, 2}, "is a JCEKS keystore, which cannot be read here"},
		{"already PKCS12", built.Data, "is already a PKCS12 file, not a JKS: name it with existing: {pkcs12: <file>}"},
		{"not a keystore", []byte("hello"), "is not a JKS keystore"},
		{"truncated", certfixture.JKS(t, pw, []certfixture.Trusted{{Alias: "ca", Cert: ca}}, nil)[:20], "is not a readable JKS"},
		{"a key that is not its certificate's", certfixture.JKS(t, pw, nil, []certfixture.Key{{Alias: "client", Chain: []certfixture.Cert{leaf}, KeyOf: &ca}}), `entry "client": the key does not match the first certificate of its chain`},
		{"a key without a certificate", certfixture.JKS(t, pw, nil, []certfixture.Key{{Alias: "client", KeyOf: &leaf}}), `the key "client" has no certificate`},
		{"a non-X.509 entry", certfixture.JKS(t, pw, []certfixture.Trusted{{Alias: "pgp", Type: "PGP", Content: []byte("x")}}, nil), `entry "pgp" holds a PGP certificate; only X.509 is supported`},
		{"a damaged certificate", certfixture.JKS(t, pw, []certfixture.Trusted{{Alias: "bad", Content: []byte("junk")}}, nil), `entry "bad": the certificate does not parse`},
	} {
		t.Run(c.name, func(t *testing.T) {
			_, err := Build(Input{Store: "mq", Password: pw, JKS: &Source{Field: "create.jks", Path: "./old.jks", Data: c.data}})
			if err == nil || !strings.Contains(err.Error(), c.want) {
				t.Errorf("err = %v, want %q", err, c.want)
			}
		})
	}
}

// certs and a jks together make one truststore; a jks holding a key cannot
// join a truststore.
func TestCertsWithJKS(t *testing.T) {
	a, b := certfixture.CA(t, "Root A"), certfixture.CA(t, "Root B")
	certs := []CertSource{{Source: src("create.certs[0]", "./a.pem", certfixture.PEM(a))}}
	res, err := Build(Input{Store: "ca", Password: pw, Certs: certs, JKS: &Source{Field: "create.jks", Path: "./old.jks", Data: certfixture.JKS(t, pw, []certfixture.Trusted{{Alias: "legacy", Cert: b}}, nil)}})
	if err != nil || aliases(res) != "ca-1,legacy" {
		t.Fatalf("aliases = %s, err %v; want ca-1,legacy", aliases(res), err)
	}
	_, _, leaf := chain(t)
	withKey := certfixture.JKS(t, pw, nil, []certfixture.Key{{Alias: "client", Chain: []certfixture.Cert{leaf}}})
	_, err = Build(Input{Store: "ca", Password: pw, Certs: certs, JKS: &Source{Field: "create.jks", Path: "./old.jks", Data: withKey}})
	if err == nil || !strings.Contains(err.Error(), `holds private keys ("client"), but tls.stores.ca is a truststore`) {
		t.Errorf("err = %v", err)
	}
	wrong := certfixture.JKS(t, "other-pass", []certfixture.Trusted{{Alias: "legacy", Cert: b}}, nil)
	_, err = Build(Input{Store: "ca", Password: pw, Certs: certs, JKS: &Source{Field: "create.jks", Path: "./old.jks", Data: wrong}})
	if err == nil || !strings.Contains(err.Error(), "does not open with the password of tls.stores.ca") {
		t.Errorf("a JKS beside certs that the store password does not open: err = %v", err)
	}
}

// derOf is the DER inside a single PEM block.
func derOf(p []byte) []byte {
	blk, _ := pem.Decode(p)
	return blk.Bytes
}

// x25519Key is a PKCS#8 X25519 key: a valid private key, but a key-agreement
// one that cannot sign, so no TLS certificate can be presented with it.
func x25519Key(t *testing.T) []byte {
	t.Helper()
	k, err := ecdh.X25519().GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	der, err := x509.MarshalPKCS8PrivateKey(k)
	if err != nil {
		t.Fatal(err)
	}
	return pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der})
}

// A certificate or chain file that holds no certificate fails naming its own
// position, not the key's.
func TestKeystoreRejectsBrokenFiles(t *testing.T) {
	_, _, leaf := chain(t)
	for _, c := range []struct {
		name  string
		cert  []byte
		chain []Source
		want  string
	}{
		{"the cert file", []byte("hello"), nil, "tls.stores.client.create.keys[0].cert (./c.pem) holds no certificate"},
		{"a chain file", certfixture.PEM(leaf), []Source{src("create.keys[0].chain[0]", "./chain.pem", []byte("hello"))}, "tls.stores.client.create.keys[0].chain[0] (./chain.pem) holds no certificate"},
	} {
		t.Run(c.name, func(t *testing.T) {
			_, err := Build(Input{Store: "client", Password: pw, Keys: []KeySource{{
				Cert:  src("create.keys[0].cert", "./c.pem", c.cert),
				Chain: c.chain,
				Key:   src("create.keys[0].key", "./c.key", certfixture.KeyPKCS8(t, leaf)),
			}}})
			if err == nil || !strings.Contains(err.Error(), c.want) {
				t.Errorf("err = %v, want %q", err, c.want)
			}
		})
	}
}

// A leaf given without its issuer still builds -- the peer may already hold
// the intermediate -- with an empty chain.
func TestKeystoreLeafWithoutItsIssuer(t *testing.T) {
	_, _, leaf := chain(t)
	res, err := Build(Input{Store: "client", Password: pw, Keys: []KeySource{{
		Cert: src("create.keys[0].cert", "./c.pem", certfixture.PEM(leaf)),
		Key:  src("create.keys[0].key", "./c.key", certfixture.KeyPKCS8(t, leaf)),
	}}})
	if err != nil {
		t.Fatal(err)
	}
	if _, _, gotChain, err := pkcs12.DecodeChain(res.Data, pw); err != nil || len(gotChain) != 0 {
		t.Errorf("chain = %d certificates, err %v; want none", len(gotChain), err)
	}
}

// PKCS#12 encodes the password as UCS-2, so a character outside the Basic
// Multilingual Plane cannot be written: both kinds of store say so instead of
// producing a file nothing could open.
func TestPasswordOutsideBMP(t *testing.T) {
	const emoji = "pass-\U0001F600"
	ca := certfixture.CA(t, "Root")
	_, _, leaf := chain(t)
	for _, c := range []struct {
		name string
		in   Input
		want string
	}{
		{"truststore", Input{Store: "ca", Password: emoji, Certs: []CertSource{{Source: src("create.certs[0]", "./ca.pem", certfixture.PEM(ca))}}}, "tls.stores.ca: encoding the truststore failed"},
		{"keystore", Input{Store: "client", Password: emoji, Keys: []KeySource{{
			Cert: src("create.keys[0].cert", "./c.pem", certfixture.PEM(leaf)),
			Key:  src("create.keys[0].key", "./c.key", certfixture.KeyPKCS8(t, leaf)),
		}}}, "tls.stores.client: encoding the keystore failed"},
	} {
		t.Run(c.name, func(t *testing.T) {
			_, err := Build(c.in)
			if err == nil || !strings.Contains(err.Error(), c.want) || !strings.Contains(err.Error(), "UCS-2") {
				t.Errorf("err = %v, want %q and UCS-2", err, c.want)
			}
		})
	}
}

// A store that does not read back as built is reported as a bug, with the
// decoder's error when there is one.
func TestReadBackError(t *testing.T) {
	for _, c := range []struct {
		err  error
		want string
	}{
		{nil, "tls.stores.ca: the encoded store did not read back as built (its content differs from what was encoded); this is a bug in solmq-conn-util -- please report it"},
		{errors.New("pkcs12: bad MAC"), "tls.stores.ca: the encoded store did not read back as built (pkcs12: bad MAC); this is a bug in solmq-conn-util -- please report it"},
	} {
		if got := readBackError("ca", c.err).Error(); got != c.want {
			t.Errorf("readBackError(%v) = %q, want %q", c.err, got, c.want)
		}
	}
}

// The same inputs and password encode to the same bytes -- what keeps a
// redeploy from rewriting the platform secret -- while another password or
// other content does not, and moving a file changes nothing.
func TestBuildIsDeterministic(t *testing.T) {
	a, b := certfixture.CA(t, "Root A"), certfixture.CA(t, "Root B")
	in := func(password, path string, cert certfixture.Cert) Input {
		return Input{Store: "ca", Password: password, Certs: []CertSource{{Source: src("create.certs[0]", path, certfixture.PEM(cert))}}}
	}
	build := func(i Input) []byte {
		t.Helper()
		res, err := Build(i)
		if err != nil {
			t.Fatal(err)
		}
		return res.Data
	}
	first := build(in(pw, "./a.pem", a))
	if !bytes.Equal(first, build(in(pw, "./a.pem", a))) {
		t.Error("the same inputs encoded to different bytes")
	}
	if !bytes.Equal(first, build(in(pw, "./moved/a.pem", a))) {
		t.Error("moving a source file changed the bytes")
	}
	if bytes.Equal(first, build(in("another-pass", "./a.pem", a))) {
		t.Error("another password encoded to the same bytes")
	}
	if bytes.Equal(first, build(in(pw, "./a.pem", b))) {
		t.Error("another certificate encoded to the same bytes")
	}
}

func TestBuildNeedsPasswordAndBoundedSources(t *testing.T) {
	ca := certfixture.CA(t, "Root")
	if _, err := Build(Input{Store: "ca", Certs: []CertSource{{Source: src("create.certs[0]", "./a.pem", certfixture.PEM(ca))}}}); err == nil || !strings.Contains(err.Error(), "tls.stores.ca is created and encrypted with its own password, but the password is empty") {
		t.Errorf("empty password: err = %v", err)
	}
	big := make([]byte, MaxSourceBytes+1)
	if _, err := Build(Input{Store: "ca", Password: pw, Certs: []CertSource{{Source: src("create.certs[0]", "./big.pem", big)}}}); err == nil || !strings.Contains(err.Error(), "tls.stores.ca.create.certs[0] (./big.pem) is 1048577 bytes, over the 1048576") {
		t.Errorf("oversized source: err = %v", err)
	}
}

// A panic inside a library comes back as an error naming the store and the
// panic's type only -- never its value, which may quote the input.
func TestGuardTurnsPanicIntoError(t *testing.T) {
	_, err := guard("ca", func() (Result, error) { panic("S3cr3t inside") })
	if err == nil || err.Error() != "tls.stores.ca: building the store failed inside a library (string); check its source files" {
		t.Errorf("err = %v", err)
	}
}

// No error repeats the password or the key's bytes, whatever fails.
func TestErrorsCarryNoSecrets(t *testing.T) {
	const secret = "S3cr3t-store-pass"
	_, _, leaf := chain(t)
	other := certfixture.CA(t, "other", certfixture.NotCA())
	keyPEM := certfixture.KeyPKCS8(t, other)
	keyBody := base64.StdEncoding.EncodeToString(certfixture.KeyPKCS8DER(t, other))[:24]
	for _, in := range []Input{
		{Store: "c", Password: secret, Keys: []KeySource{{Cert: src("create.keys[0].cert", "./c.pem", certfixture.PEM(leaf)), Key: src("create.keys[0].key", "./c.key", keyPEM)}}},
		{Store: "c", Password: secret, JKS: &Source{Field: "create.jks", Path: "./x.jks", Data: certfixture.JKS(t, "wrong-pass", []certfixture.Trusted{{Alias: "a", Cert: other}}, nil)}},
		{Store: "c", Password: secret, Certs: []CertSource{{Source: src("create.certs[0]", "./k.pem", keyPEM)}}},
	} {
		_, err := Build(in)
		if err == nil {
			t.Fatalf("%+v built", in.Store)
		}
		if strings.Contains(err.Error(), secret) || strings.Contains(err.Error(), keyBody) {
			t.Errorf("error leaks a secret: %v", err)
		}
	}
}

// A certificate is untrusted input: control characters in its subject never
// reach a terminal.
func TestSubjectIsPrintable(t *testing.T) {
	c := certfixture.CA(t, "evil\x1b[31mred")
	if got := subject(c.Cert); strings.ContainsRune(got, 0x1b) || !strings.Contains(got, "evil?[31mred") {
		t.Errorf("subject = %q", got)
	}
}

func TestSniff(t *testing.T) {
	for _, c := range []struct {
		data []byte
		want format
	}{
		{[]byte{0xFE, 0xED, 0xFE, 0xED, 0, 0, 0, 2}, formatJKS},
		{[]byte{0xCE, 0xCE, 0xCE, 0xCE}, formatJCEKS},
		{[]byte{0x30, 0x82, 0x01}, formatPKCS12},
		{[]byte("-----BEGIN"), formatUnknown},
		{nil, formatUnknown},
	} {
		if got := sniff(c.data); got != c.want {
			t.Errorf("sniff(% x) = %d, want %d", c.data, got, c.want)
		}
	}
}

// Java reads both kinds of store: the truststore's entries are trusted
// certificates under their aliases, and the keystore's one key is alias 1 --
// the fact the key-alias rule for created keystores rests on. Runs only where
// keytool is on PATH; every assertion above it checks the encoding from Go.
func TestJavaReadsBuiltStores(t *testing.T) {
	keytool, err := exec.LookPath("keytool")
	if err != nil {
		t.Skip("keytool is not on PATH; this Java interop check runs where a JDK is installed")
	}
	root, _, leaf := chain(t)
	trust, err := Build(Input{Store: "ca", Password: pw, Certs: []CertSource{{Alias: "root", Source: src("create.certs.root", "./root.pem", certfixture.PEM(root))}}})
	if err != nil {
		t.Fatal(err)
	}
	key, err := Build(Input{Store: "client", Password: pw, Keys: []KeySource{{
		Cert: src("create.keys[0].cert", "./c.pem", certfixture.PEM(leaf)),
		Key:  src("create.keys[0].key", "./c.key", certfixture.KeyPKCS8(t, leaf)),
	}}})
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range []struct {
		name string
		data []byte
		want []string
	}{
		{"truststore", trust.Data, []string{"Alias name: root", "Entry type: trustedCertEntry"}},
		{"keystore", key.Data, []string{"Alias name: 1", "Entry type: PrivateKeyEntry"}},
	} {
		path := filepath.Join(t.TempDir(), c.name+".p12")
		if err := os.WriteFile(path, c.data, 0o600); err != nil {
			t.Fatal(err)
		}
		cmd := exec.Command(keytool, "-J-Duser.language=en", "-list", "-v", "-storetype", "PKCS12", "-keystore", path, "-storepass:env", "SOLMQ_TEST_STOREPASS")
		cmd.Env = append(os.Environ(), "SOLMQ_TEST_STOREPASS="+pw)
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("%s: keytool: %v\n%s", c.name, err, out)
		}
		for _, w := range c.want {
			if !strings.Contains(string(out), w) {
				t.Errorf("%s: keytool output lacks %q:\n%s", c.name, w, out)
			}
		}
	}
}
