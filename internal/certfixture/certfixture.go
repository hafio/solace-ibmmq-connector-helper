// Package certfixture mints the certificates, private keys and Java keystores
// tests need, at run time, so no key material is ever committed to the
// repository. It is imported only from _test.go files, so it is never linked
// into the binary.
package certfixture

import (
	"bytes"
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"testing"
	"time"

	"github.com/pavlo-v-chernykh/keystore-go/v4"
)

// Cert is a certificate and the key it was issued for.
type Cert struct {
	Cert *x509.Certificate
	Key  crypto.Signer
}

// params is what the options shape: the template, and the key type.
type params struct {
	tmpl   *x509.Certificate
	useRSA bool
}

// Option adjusts a certificate before it is signed.
type Option func(*params)

// Validity sets the certificate's validity window.
func Validity(notBefore, notAfter time.Time) Option {
	return func(p *params) { p.tmpl.NotBefore, p.tmpl.NotAfter = notBefore, notAfter }
}

// AsCA makes an issued certificate a CA -- an intermediate.
func AsCA() Option {
	return func(p *params) {
		p.tmpl.IsCA = true
		p.tmpl.KeyUsage = x509.KeyUsageCertSign | x509.KeyUsageCRLSign | x509.KeyUsageDigitalSignature
	}
}

// NotCA makes a self-signed certificate a plain leaf: CA mints it with AsCA
// first, and this undoes that.
func NotCA() Option {
	return func(p *params) {
		p.tmpl.IsCA = false
		p.tmpl.KeyUsage = x509.KeyUsageDigitalSignature
	}
}

// RSAKey gives the certificate a 2048-bit RSA key instead of ECDSA P-256.
func RSAKey() Option {
	return func(p *params) { p.useRSA = true }
}

// CA mints a self-signed CA certificate for cn.
func CA(t testing.TB, cn string, opts ...Option) Cert {
	t.Helper()
	return mint(t, nil, cn, append([]Option{AsCA()}, opts...)...)
}

// Issue mints a certificate for cn signed by parent: a leaf, or an
// intermediate with AsCA.
func Issue(t testing.TB, parent Cert, cn string, opts ...Option) Cert {
	t.Helper()
	return mint(t, &parent, cn, opts...)
}

func mint(t testing.TB, parent *Cert, cn string, opts ...Option) Cert {
	t.Helper()
	tmpl := &x509.Certificate{
		SerialNumber:          serial(t),
		Subject:               pkix.Name{CommonName: cn},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}
	p := &params{tmpl: tmpl}
	for _, o := range opts {
		o(p)
	}
	key := newKey(t, p.useRSA)
	signerCert, signerKey := tmpl, key
	if parent != nil {
		signerCert, signerKey = parent.Cert, parent.Key
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, signerCert, key.Public(), signerKey)
	must(t, err, "minting "+cn)
	c, err := x509.ParseCertificate(der)
	must(t, err, "parsing "+cn)
	return Cert{Cert: c, Key: key}
}

// must stops the test when a fixture cannot be made.
func must(t testing.TB, err error, what string) {
	t.Helper()
	if err != nil {
		t.Fatalf("certfixture: %s: %v", what, err)
	}
}

func newKey(t testing.TB, useRSA bool) crypto.Signer {
	t.Helper()
	if useRSA {
		k, err := rsa.GenerateKey(rand.Reader, 2048)
		must(t, err, "RSA key")
		return k
	}
	k, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	must(t, err, "ECDSA key")
	return k
}

func serial(t testing.TB) *big.Int {
	t.Helper()
	n, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 62))
	must(t, err, "serial")
	return n
}

// PEM encodes the certificates as CERTIFICATE blocks, in order.
func PEM(certs ...Cert) []byte {
	var buf bytes.Buffer
	for _, c := range certs {
		_ = pem.Encode(&buf, &pem.Block{Type: "CERTIFICATE", Bytes: c.Cert.Raw})
	}
	return buf.Bytes()
}

// DER concatenates the certificates' DER, in order.
func DER(certs ...Cert) []byte {
	var out []byte
	for _, c := range certs {
		out = append(out, c.Cert.Raw...)
	}
	return out
}

// KeyPKCS8 is c's key as a PEM PRIVATE KEY block.
func KeyPKCS8(t testing.TB, c Cert) []byte {
	t.Helper()
	return pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: KeyPKCS8DER(t, c)})
}

// KeyPKCS8DER is c's key as PKCS#8 DER.
func KeyPKCS8DER(t testing.TB, c Cert) []byte {
	t.Helper()
	der, err := x509.MarshalPKCS8PrivateKey(c.Key)
	must(t, err, "PKCS#8")
	return der
}

// KeyPKCS1 is c's RSA key as a PEM RSA PRIVATE KEY block.
func KeyPKCS1(t testing.TB, c Cert) []byte {
	t.Helper()
	k, ok := c.Key.(*rsa.PrivateKey)
	if !ok {
		t.Fatalf("certfixture: KeyPKCS1 needs an RSA key, got %T", c.Key)
	}
	return pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(k)})
}

// KeyEC is c's ECDSA key as a PEM EC PRIVATE KEY block (SEC1).
func KeyEC(t testing.TB, c Cert) []byte {
	t.Helper()
	k, ok := c.Key.(*ecdsa.PrivateKey)
	if !ok {
		t.Fatalf("certfixture: KeyEC needs an ECDSA key, got %T", c.Key)
	}
	der, err := x509.MarshalECPrivateKey(k)
	must(t, err, "SEC1")
	return pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: der})
}

// Trusted is one trusted-certificate entry of a JKS. Type and Content, when
// set, replace the certificate's own -- an entry as a foreign or damaged JKS
// holds it.
type Trusted struct {
	Alias   string
	Cert    Cert
	Type    string
	Content []byte
}

// Key is one private-key entry of a JKS: the key of Chain[0] -- or of KeyOf,
// when set, for an entry whose key and certificate do not belong together --
// protected with Password (the store password when empty).
type Key struct {
	Alias    string
	Chain    []Cert
	KeyOf    *Cert
	Password string
}

// JKS writes a JKS keystore holding trusted and keys, protected with password.
func JKS(t testing.TB, password string, trusted []Trusted, keys []Key) []byte {
	t.Helper()
	ks := keystore.New()
	now := time.Now()
	for _, tr := range trusted {
		cert := keystore.Certificate{Type: "X.509", Content: tr.Content}
		if tr.Type != "" {
			cert.Type = tr.Type
		}
		if tr.Content == nil {
			cert.Content = tr.Cert.Cert.Raw
		}
		must(t, ks.SetTrustedCertificateEntry(tr.Alias, keystore.TrustedCertificateEntry{CreationTime: now, Certificate: cert}), "JKS trusted "+tr.Alias)
	}
	for _, k := range keys {
		var chain []keystore.Certificate
		for _, c := range k.Chain {
			chain = append(chain, keystore.Certificate{Type: "X.509", Content: c.Cert.Raw})
		}
		pw := k.Password
		if pw == "" {
			pw = password
		}
		owner := k.KeyOf
		if owner == nil {
			owner = &k.Chain[0]
		}
		must(t, ks.SetPrivateKeyEntry(k.Alias, keystore.PrivateKeyEntry{
			CreationTime:     now,
			PrivateKey:       KeyPKCS8DER(t, *owner),
			CertificateChain: chain,
		}, []byte(pw)), "JKS key "+k.Alias)
	}
	var buf bytes.Buffer
	must(t, ks.Store(&buf, []byte(password)), "JKS store")
	return buf.Bytes()
}
