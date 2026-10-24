package storebuild

import (
	"bytes"
	"crypto"
	"crypto/x509"
	"encoding/binary"
	"fmt"
	"strings"

	"github.com/pavlo-v-chernykh/keystore-go/v4"
)

// format is what a keystore file's first bytes say it is.
type format int

const (
	formatUnknown format = iota
	formatJKS
	formatJCEKS
	formatPKCS12
)

// sniff tells a JKS (magic FEEDFEED) from a JCEKS (CECECECE) and a PKCS#12
// (a DER SEQUENCE). Many a .jks file written by keytool since JDK 9 is really
// PKCS#12, which is why the extension is never trusted.
func sniff(data []byte) format {
	if len(data) >= 4 {
		switch binary.BigEndian.Uint32(data) {
		case 0xFEEDFEED:
			return formatJKS
		case 0xCECECECE:
			return formatJCEKS
		}
	}
	if len(data) > 0 && data[0] == 0x30 {
		return formatPKCS12
	}
	return formatUnknown
}

// jksKey is one private-key entry of a JKS: its alias, key and chain (leaf
// first).
type jksKey struct {
	alias string
	key   crypto.Signer
	chain []*x509.Certificate
}

// fromJKS builds a store from create.jks alone: a JKS holding only trusted
// certificates becomes a truststore keeping their aliases; one holding exactly
// one private key and nothing else becomes a keystore. Anything else is refused,
// since a created store is one or the other.
func (b *builder) fromJKS() ([]byte, error) {
	src := *b.in.JKS
	where := src.where(b.in.Store)
	ts, keys, err := readJKS(src, b.in.Store, b.in.Password)
	if err != nil {
		return nil, err
	}
	switch {
	case len(keys) == 0 && len(ts) > 0:
		return b.encodeTrust(ts)
	case len(keys) == 1 && len(ts) == 0:
		k := keys[0]
		from := fmt.Sprintf("%s entry %q", where, k.alias)
		if !matches(k.key, k.chain[0]) {
			return nil, fmt.Errorf("%s: the key does not match the first certificate of its chain", from)
		}
		return b.encodeKey(k.key, k.chain[0], k.chain[1:], from)
	case len(keys) == 0:
		return nil, fmt.Errorf("%s holds no entries", where)
	default:
		var aliases []string
		for _, t := range ts {
			aliases = append(aliases, fmt.Sprintf("%q (trusted)", t.alias))
		}
		for _, k := range keys {
			aliases = append(aliases, fmt.Sprintf("%q (key)", k.alias))
		}
		return nil, fmt.Errorf("%s holds %d trusted certificates and %d private keys -- %s; a created store is a truststore or holds exactly one key, so split the JKS with keytool, or list certs or keys instead", where, len(ts), len(keys), strings.Join(aliases, ", "))
	}
}

// readJKS opens src with the store's password and returns its trusted
// certificates and private keys, in alias order. The password opens both the
// keystore and its keys: keytool gives a key the store's password unless told
// otherwise, and a created store has one password.
func readJKS(src Source, store, password string) (ts []trusted, keys []jksKey, err error) {
	where := src.where(store)
	switch sniff(src.Data) {
	case formatJKS:
	case formatJCEKS:
		return nil, nil, fmt.Errorf("%s is a JCEKS keystore, which cannot be read here: convert it with keytool -importkeystore -srcstoretype JCEKS -deststoretype PKCS12 and name the result with existing: {pkcs12: <file>}", where)
	case formatPKCS12:
		return nil, nil, fmt.Errorf("%s is already a PKCS12 file, not a JKS: name it with existing: {pkcs12: <file>} instead of create.jks", where)
	default:
		return nil, nil, fmt.Errorf("%s is not a JKS keystore", where)
	}
	ks := keystore.New(keystore.WithCaseExactAliases(), keystore.WithOrderedAliases())
	if err := ks.Load(bytes.NewReader(src.Data), []byte(password)); err != nil {
		// keystore-go has no sentinel for a failed integrity check; its
		// message is the only signal that the password is the problem.
		if strings.Contains(err.Error(), "invalid digest") {
			return nil, nil, fmt.Errorf("%s does not open with the password of tls.stores.%s (the JKS integrity check failed); a created store opens its jks with its own password", where, store)
		}
		return nil, nil, fmt.Errorf("%s is not a readable JKS: %v", where, err)
	}
	for _, alias := range ks.Aliases() {
		switch {
		case ks.IsTrustedCertificateEntry(alias):
			e, err := ks.GetTrustedCertificateEntry(alias)
			if err != nil {
				return nil, nil, fmt.Errorf("%s: entry %q: %v", where, alias, err)
			}
			c, err := jksCert(e.Certificate, where, alias)
			if err != nil {
				return nil, nil, err
			}
			ts = append(ts, trusted{alias: alias, cert: c, from: fmt.Sprintf("%s entry %q", where, alias)})
		case ks.IsPrivateKeyEntry(alias):
			e, err := ks.GetPrivateKeyEntry(alias, []byte(password))
			if err != nil {
				return nil, nil, fmt.Errorf("%s: the key %q does not open with the store password; give it the store password first (keytool -keypasswd -alias %s)", where, alias, alias)
			}
			k, err := x509.ParsePKCS8PrivateKey(e.PrivateKey)
			if err != nil {
				return nil, nil, fmt.Errorf("%s: the key %q does not parse: %v", where, alias, err)
			}
			s, err := signer(k, fmt.Sprintf("%s entry %q", where, alias))
			if err != nil {
				return nil, nil, err
			}
			var chain []*x509.Certificate
			for _, jc := range e.CertificateChain {
				c, err := jksCert(jc, where, alias)
				if err != nil {
					return nil, nil, err
				}
				chain = append(chain, c)
			}
			if len(chain) == 0 {
				return nil, nil, fmt.Errorf("%s: the key %q has no certificate", where, alias)
			}
			keys = append(keys, jksKey{alias: alias, key: s, chain: chain})
		}
	}
	return ts, keys, nil
}

// jksCert parses one certificate of a JKS entry. keytool records the type as
// "X.509"; keystore-go writes "X509".
func jksCert(c keystore.Certificate, where, alias string) (*x509.Certificate, error) {
	if t := strings.ReplaceAll(strings.ToUpper(c.Type), ".", ""); t != "X509" {
		return nil, fmt.Errorf("%s: entry %q holds a %s certificate; only X.509 is supported", where, alias, c.Type)
	}
	cert, err := x509.ParseCertificate(c.Content)
	if err != nil {
		return nil, fmt.Errorf("%s: entry %q: the certificate does not parse: %v", where, alias, err)
	}
	return cert, nil
}

// joinAliases lists the aliases of keys for a message.
func joinAliases(keys []jksKey) string {
	var out []string
	for _, k := range keys {
		out = append(out, fmt.Sprintf("%q", k.alias))
	}
	return strings.Join(out, ", ")
}
