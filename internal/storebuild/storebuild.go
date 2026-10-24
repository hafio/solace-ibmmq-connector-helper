// Package storebuild builds the PKCS#12 truststore or keystore a tls.stores
// entry declares under create:, entirely in memory. It is a pure builder: it is
// handed the source files' bytes and the store's password and returns the
// store's bytes -- no filesystem, no environment, no globals. Callers deliver
// the result straight into the platform's secret (the kubernetes stores Secret,
// a podman secret, a docker compose secret), so a created store never lands on
// the operator's disk.
//
// Every store is PKCS#12, encoded with go-pkcs12's Modern2023 profile (PBES2
// with PBKDF2-HMAC-SHA-256 and AES-256-CBC, an HMAC-SHA-256 MAC), which Java 12+
// reads -- the connector images ship Java 17 and 25. A store is a truststore or
// a keystore, never both: go-pkcs12 writes trusted certificates (marked with
// the attribute Java needs to treat them as trust anchors) or exactly one
// private key with its chain, and cannot name that key, so Java lists it as
// alias "1".
package storebuild

import (
	"bytes"
	"crypto"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/x509"
	"encoding/binary"
	"encoding/pem"
	"fmt"
	"io"
	"math/rand/v2"
	"strings"
	"time"

	pkcs12 "software.sslmate.com/src/go-pkcs12"
)

// Iterations is the PBKDF2 iteration count for the store's encryption and MAC
// keys: keytool's default since JDK 12, and far below the 5,000,000 the JDK
// refuses on load.
const Iterations = 10000

// MaxSourceBytes caps one source file. A CA bundle of every public root is a
// few hundred kilobytes; anything near this is not a certificate file.
const MaxSourceBytes = 1 << 20

// expiringWithin is how far ahead an expiry is worth a warning.
const expiringWithin = 30 * 24 * time.Hour

// seedLabel versions the deterministic randomness (see seededRand): changing
// how salts and IVs are derived must change this label, never silently reuse it.
const seedLabel = "solmq-conn-util/storebuild/v1"

// Source is one file a created store is built from.
type Source struct {
	// Field is where env.yaml names the file, below tls.stores.<store>:
	// "create.certs[0]", "create.certs.root", "create.keys[0].key", "create.jks".
	Field string
	// Path is the file as env.yaml spells it -- what messages name.
	Path string
	Data []byte
}

// where labels s in messages: the env.yaml position and the file.
func (s Source) where(store string) string {
	return "tls.stores." + store + "." + s.Field + " (" + s.Path + ")"
}

// CertSource is one create.certs entry. Alias is the entry's name when certs is
// a named map, and empty when certs is a list.
type CertSource struct {
	Alias string
	Source
}

// KeySource is one create.keys entry: the certificate the key belongs to (its
// file may also hold the chain), optional further chain files, and the key.
type KeySource struct {
	Cert  Source
	Chain []Source
	Key   Source
}

// Input is everything one created store is built from.
type Input struct {
	// Store is the tls.stores name: it prefixes generated aliases and every
	// message.
	Store string
	// Password encrypts the store, and opens a JKS source.
	Password string
	Certs    []CertSource
	// Keys holds at most one entry; validate refuses more.
	Keys []KeySource
	JKS  *Source
	// Now is the time expiry is judged against; zero means time.Now().
	Now time.Time
}

// Entry describes one certificate in a built store, for the summary a command
// prints. It carries no key material.
type Entry struct {
	// Alias is the entry's name as Java lists it (lower-cased). The key
	// entry's leaf is "1" -- Java's name for a key go-pkcs12 cannot name -- and
	// the certificates of its chain have none.
	Alias    string
	Subject  string
	NotAfter time.Time
	// SHA256 is the certificate's fingerprint, colon-separated upper-case hex
	// as keytool prints it.
	SHA256 string
	// Key marks the key entry's own certificate.
	Key bool
}

// Result is a built store.
type Result struct {
	Data []byte
	// Keystore is true for a store holding a private key, false for a
	// truststore.
	Keystore bool
	Entries  []Entry
	// Warnings are findings that do not stop the build: an expiring
	// certificate, a certificate left out of a chain, a pinned leaf.
	Warnings []string
}

// Build builds the store in declares. The bytes are a deterministic function of
// the password and the sources (see seededRand), so an unchanged store encodes
// to the same bytes on every run: a redeploy leaves the platform secret
// unchanged and generated manifests stable.
//
// A panic inside a library -- x/crypto's PBKDF2 panics on error, and the JKS
// reader parses an untrusted file -- comes back as an error naming the store
// only, never the input.
func Build(in Input) (Result, error) {
	return guard(in.Store, func() (Result, error) { return build(in) })
}

// guard runs f and turns a panic into an error naming the store only: the
// panic value may quote the input, so it never reaches the message.
func guard(store string, f func() (Result, error)) (res Result, err error) {
	defer func() {
		if r := recover(); r != nil {
			res, err = Result{}, fmt.Errorf("tls.stores.%s: building the store failed inside a library (%T); check its source files", store, r)
		}
	}()
	return f()
}

func build(in Input) (Result, error) {
	if in.Password == "" {
		return Result{}, fmt.Errorf("tls.stores.%s is created and encrypted with its own password, but the password is empty", in.Store)
	}
	if in.Now.IsZero() {
		in.Now = time.Now()
	}
	for _, s := range in.sources() {
		if len(s.Data) > MaxSourceBytes {
			return Result{}, fmt.Errorf("%s is %d bytes, over the %d a source file may be; name the certificate files themselves", s.where(in.Store), len(s.Data), MaxSourceBytes)
		}
	}
	b := &builder{in: in}
	var data []byte
	var err error
	switch {
	case len(in.Keys) > 0:
		data, err = b.keystore(in.Keys[0])
	case len(in.Certs) == 0 && in.JKS != nil:
		data, err = b.fromJKS()
	default:
		data, err = b.truststore()
	}
	if err != nil {
		return Result{}, err
	}
	return Result{Data: data, Keystore: b.keystoreBuilt, Entries: b.entries, Warnings: b.warnings}, nil
}

// sources lists every source file of in, in a fixed order.
func (in Input) sources() []Source {
	var out []Source
	for _, c := range in.Certs {
		out = append(out, c.Source)
	}
	for _, k := range in.Keys {
		out = append(out, k.Cert)
		out = append(out, k.Chain...)
		out = append(out, k.Key)
	}
	if in.JKS != nil {
		out = append(out, *in.JKS)
	}
	return out
}

// builder accumulates one build's entries and warnings.
type builder struct {
	in            Input
	entries       []Entry
	warnings      []string
	keystoreBuilt bool
}

func (b *builder) warn(format string, a ...any) {
	b.warnings = append(b.warnings, fmt.Sprintf(format, a...))
}

// encoder is Modern2023 at Iterations, drawing its salts and IVs from the
// store's deterministic stream.
func (b *builder) encoder() *pkcs12.Encoder {
	return pkcs12.Modern2023.WithIterations(Iterations).WithRand(seededRand(b.in))
}

// seededRand returns the stream go-pkcs12 draws salts and IVs from: ChaCha8 (a
// cryptographically strong generator) seeded with an HMAC-SHA-256, keyed by the
// store's password, over a length-prefixed encoding of the store's name, every
// alias and every source's bytes. The same inputs and password therefore
// encode to the same bytes, which keeps `kubectl apply` a no-op and a store
// digest meaningful; the salts stay unpredictable without the password, and two
// stores with different content or passwords never share them. Source paths
// are deliberately left out: moving a file changes nothing.
func seededRand(in Input) io.Reader {
	mac := hmac.New(sha256.New, []byte(in.Password))
	field := func(p []byte) {
		var n [8]byte
		binary.BigEndian.PutUint64(n[:], uint64(len(p)))
		mac.Write(n[:])
		mac.Write(p)
	}
	field([]byte(seedLabel))
	field([]byte(in.Store))
	for _, c := range in.Certs {
		field([]byte(c.Alias))
		field(c.Data)
	}
	for _, k := range in.Keys {
		field(k.Cert.Data)
		for _, ch := range k.Chain {
			field(ch.Data)
		}
		field(k.Key.Data)
	}
	if in.JKS != nil {
		field(in.JKS.Data)
	}
	var seed [32]byte
	copy(seed[:], mac.Sum(nil))
	return rand.NewChaCha8(seed)
}

// ---- truststore ----------------------------------------------------------------

// trusted is one certificate on its way into a truststore.
type trusted struct {
	alias string
	cert  *x509.Certificate
	from  string // the source it came from, for messages
}

// truststore builds a store of trusted certificates from create.certs, plus the
// trusted entries of create.jks when both are given.
func (b *builder) truststore() ([]byte, error) {
	var all []trusted
	for _, cs := range b.in.Certs {
		certs, sawKey, err := parseCerts(cs.Source, b.in.Store)
		if err != nil {
			return nil, err
		}
		if sawKey {
			b.warn("%s also holds a private key, which a truststore leaves out", cs.where(b.in.Store))
		}
		for j, c := range certs {
			all = append(all, trusted{alias: certAlias(b.in.Store, cs.Alias, j, len(all)), cert: c, from: cs.where(b.in.Store)})
		}
	}
	if b.in.JKS != nil {
		ts, keys, err := readJKS(*b.in.JKS, b.in.Store, b.in.Password)
		if err != nil {
			return nil, err
		}
		if len(keys) > 0 {
			return nil, fmt.Errorf("%s holds private keys (%s), but tls.stores.%s is a truststore (it lists certs); a truststore takes only trusted certificates", b.in.JKS.where(b.in.Store), joinAliases(keys), b.in.Store)
		}
		all = append(all, ts...)
	}
	return b.encodeTrust(all)
}

// certAlias names the j-th certificate of a certs entry: a named entry's own
// name (name, name-2, ... when its file holds several), or <store>-<n> numbering
// every certificate of a list in order (seen certificates came before it).
func certAlias(store, name string, j, seen int) string {
	if name == "" {
		return fmt.Sprintf("%s-%d", store, seen+1)
	}
	if j == 0 {
		return name
	}
	return fmt.Sprintf("%s-%d", name, j+1)
}

// encodeTrust dedupes, checks and encodes trusted certificates.
func (b *builder) encodeTrust(all []trusted) ([]byte, error) {
	if len(all) == 0 {
		return nil, fmt.Errorf("tls.stores.%s holds no certificate to trust", b.in.Store)
	}
	seenCert := map[[32]byte]string{}
	seenAlias := map[string]string{}
	var entries []pkcs12.TrustStoreEntry
	var keep []*x509.Certificate
	for _, t := range all {
		fp := sha256.Sum256(t.cert.Raw)
		if prev, ok := seenCert[fp]; ok {
			b.warn("tls.stores.%s: %s repeats a certificate from %s (%s); it is kept once", b.in.Store, t.from, prev, subject(t.cert))
			continue
		}
		seenCert[fp] = t.from
		key := strings.ToLower(t.alias)
		if prev, ok := seenAlias[key]; ok {
			return nil, fmt.Errorf("tls.stores.%s: alias %q names two certificates (from %s and %s); aliases ignore case, so rename one", b.in.Store, t.alias, prev, t.from)
		}
		seenAlias[key] = t.from
		b.checkTrusted(t)
		entries = append(entries, pkcs12.TrustStoreEntry{Cert: t.cert, FriendlyName: t.alias})
		keep = append(keep, t.cert)
		b.entries = append(b.entries, entry(key, t.cert, false))
	}
	data, err := b.encoder().EncodeTrustStoreEntries(entries, b.in.Password)
	if err != nil {
		return nil, fmt.Errorf("tls.stores.%s: encoding the truststore failed: %v", b.in.Store, err)
	}
	got, err := pkcs12.DecodeTrustStore(data, b.in.Password)
	if err != nil || !sameCerts(got, keep) {
		return nil, readBackError(b.in.Store, err)
	}
	return data, nil
}

// checkTrusted warns about a trusted certificate that is not a CA (so only that
// exact certificate is accepted), has expired, or expires soon.
func (b *builder) checkTrusted(t trusted) {
	if !t.cert.IsCA {
		b.warn("tls.stores.%s: %s (%s) is not a CA certificate; it is trusted as-is, so only that exact certificate is accepted", b.in.Store, t.from, subject(t.cert))
	}
	b.checkExpiry(t.from, t.cert)
}

// checkExpiry warns about an expired, not yet valid, or soon expiring
// certificate.
func (b *builder) checkExpiry(from string, c *x509.Certificate) {
	now := b.in.Now
	switch {
	case now.After(c.NotAfter):
		b.warn("tls.stores.%s: %s (%s) expired on %s", b.in.Store, from, subject(c), day(c.NotAfter))
	case now.Before(c.NotBefore):
		b.warn("tls.stores.%s: %s (%s) is not valid until %s", b.in.Store, from, subject(c), day(c.NotBefore))
	case c.NotAfter.Sub(now) < expiringWithin:
		b.warn("tls.stores.%s: %s (%s) expires on %s, in %d days", b.in.Store, from, subject(c), day(c.NotAfter), int(c.NotAfter.Sub(now).Hours()/24))
	}
}

// ---- keystore ------------------------------------------------------------------

// keystore builds a store holding the one key of create.keys with its chain.
func (b *builder) keystore(k KeySource) ([]byte, error) {
	store := b.in.Store
	key, err := parseKey(k.Key, store)
	if err != nil {
		return nil, err
	}
	certs, _, err := parseCerts(k.Cert, store)
	if err != nil {
		return nil, err
	}
	pool := append([]*x509.Certificate(nil), certs...)
	for _, ch := range k.Chain {
		more, _, err := parseCerts(ch, store)
		if err != nil {
			return nil, err
		}
		pool = append(pool, more...)
	}
	pool = dedupe(pool)
	leafAt := -1
	for i, c := range pool {
		if matches(key, c) {
			leafAt = i
			break
		}
	}
	if leafAt < 0 {
		return nil, fmt.Errorf("tls.stores.%s: no certificate in %s matches the private key in %s; check that the certificate and the key belong together", store, k.Cert.where(store), k.Key.where(store))
	}
	leaf := pool[leafAt]
	rest := append(append([]*x509.Certificate(nil), pool[:leafAt]...), pool[leafAt+1:]...)
	return b.encodeKey(key, leaf, rest, k.Cert.where(store))
}

// encodeKey orders the chain, checks the leaf and encodes the key entry. from
// labels the leaf in messages.
func (b *builder) encodeKey(key crypto.Signer, leaf *x509.Certificate, pool []*x509.Certificate, from string) ([]byte, error) {
	store := b.in.Store
	if b.in.Now.After(leaf.NotAfter) {
		return nil, fmt.Errorf("tls.stores.%s: the certificate in %s (%s) expired on %s; renew it -- the peer refuses an expired certificate", store, from, subject(leaf), day(leaf.NotAfter))
	}
	b.checkExpiry(from, leaf)
	chain, extra := orderChain(leaf, pool)
	for _, c := range extra {
		b.warn("tls.stores.%s: %s is not part of the chain of %s and is left out", store, subject(c), subject(leaf))
	}
	for _, c := range chain {
		b.checkExpiry("the chain of "+subject(leaf), c)
	}
	b.keystoreBuilt = true
	b.entries = append(b.entries, entry("1", leaf, true))
	for _, c := range chain {
		b.entries = append(b.entries, entry("", c, false))
	}
	data, err := b.encoder().Encode(key, leaf, chain, b.in.Password)
	if err != nil {
		return nil, fmt.Errorf("tls.stores.%s: encoding the keystore failed: %v", store, err)
	}
	gotKey, gotLeaf, gotChain, err := pkcs12.DecodeChain(data, b.in.Password)
	if err != nil || !bytes.Equal(gotLeaf.Raw, leaf.Raw) || !sameCerts(gotChain, chain) || !matchesKey(gotKey, key) {
		return nil, readBackError(store, err)
	}
	return data, nil
}

// orderChain follows issuer links up from leaf through pool, verifying each
// signature, until a self-signed certificate or no issuer is found. Whatever is
// not part of that path comes back as extra.
func orderChain(leaf *x509.Certificate, pool []*x509.Certificate) (chain, extra []*x509.Certificate) {
	used := make([]bool, len(pool))
	cur := leaf
	for !selfSigned(cur) {
		next := -1
		for i, c := range pool {
			if !used[i] && bytes.Equal(cur.RawIssuer, c.RawSubject) && cur.CheckSignatureFrom(c) == nil {
				next = i
				break
			}
		}
		if next < 0 {
			break
		}
		used[next] = true
		chain = append(chain, pool[next])
		cur = pool[next]
	}
	for i, c := range pool {
		if !used[i] {
			extra = append(extra, c)
		}
	}
	return chain, extra
}

// selfSigned reports whether c is its own issuer and signed by its own key.
// CheckSignature rather than CheckSignatureFrom, which would also demand that a
// self-signed leaf be a CA.
func selfSigned(c *x509.Certificate) bool {
	return bytes.Equal(c.RawIssuer, c.RawSubject) && c.CheckSignature(c.SignatureAlgorithm, c.RawTBSCertificate, c.Signature) == nil
}

// ---- parsing -------------------------------------------------------------------

// parseCerts reads every certificate in src: its PEM CERTIFICATE blocks, or --
// when it holds no PEM at all -- DER, one certificate or several concatenated.
// sawKey reports a private-key block alongside the certificates.
func parseCerts(src Source, store string) (certs []*x509.Certificate, sawKey bool, err error) {
	rest := src.Data
	blocks := 0
	for {
		var blk *pem.Block
		blk, rest = pem.Decode(rest)
		if blk == nil {
			break
		}
		blocks++
		switch blk.Type {
		case "CERTIFICATE":
			c, perr := x509.ParseCertificate(blk.Bytes)
			if perr != nil {
				return nil, false, fmt.Errorf("%s: certificate %d does not parse: %v", src.where(store), len(certs)+1, perr)
			}
			certs = append(certs, c)
		case "PRIVATE KEY", "RSA PRIVATE KEY", "EC PRIVATE KEY", "ENCRYPTED PRIVATE KEY":
			sawKey = true
		}
	}
	if blocks == 0 {
		if der, derr := x509.ParseCertificates(src.Data); derr == nil && len(der) > 0 {
			return der, false, nil
		}
	}
	if len(certs) == 0 {
		return nil, false, fmt.Errorf("%s holds no certificate: give PEM CERTIFICATE blocks, or DER", src.where(store))
	}
	return certs, sawKey, nil
}

// parseKey reads the one unencrypted private key in src: a PEM PRIVATE KEY
// (PKCS#8), RSA PRIVATE KEY (PKCS#1) or EC PRIVATE KEY (SEC1) block, or -- when
// src holds no PEM at all -- DER in any of those forms. Other blocks, such as
// the certificates of a combined file, are skipped.
func parseKey(src Source, store string) (crypto.Signer, error) {
	where := src.where(store)
	var key crypto.Signer
	rest := src.Data
	blocks := 0
	for {
		var blk *pem.Block
		blk, rest = pem.Decode(rest)
		if blk == nil {
			break
		}
		blocks++
		var k any
		var err error
		switch blk.Type {
		case "ENCRYPTED PRIVATE KEY":
			return nil, encryptedKey(where)
		case "PRIVATE KEY":
			k, err = x509.ParsePKCS8PrivateKey(blk.Bytes)
		case "RSA PRIVATE KEY", "EC PRIVATE KEY":
			if strings.Contains(blk.Headers["Proc-Type"], "ENCRYPTED") {
				return nil, encryptedKey(where)
			}
			if blk.Type == "RSA PRIVATE KEY" {
				k, err = x509.ParsePKCS1PrivateKey(blk.Bytes)
			} else {
				k, err = x509.ParseECPrivateKey(blk.Bytes)
			}
		default:
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("%s: the private key does not parse: %v", where, err)
		}
		if key != nil {
			return nil, fmt.Errorf("%s holds more than one private key; a key entry takes exactly one", where)
		}
		if key, err = signer(k, where); err != nil {
			return nil, err
		}
	}
	if blocks == 0 {
		for _, parse := range []func([]byte) (any, error){
			x509.ParsePKCS8PrivateKey,
			func(d []byte) (any, error) { return x509.ParsePKCS1PrivateKey(d) },
			func(d []byte) (any, error) { return x509.ParseECPrivateKey(d) },
		} {
			if k, err := parse(src.Data); err == nil {
				return signer(k, where)
			}
		}
	}
	if key == nil {
		return nil, fmt.Errorf("%s holds no private key: give a PEM PRIVATE KEY, RSA PRIVATE KEY or EC PRIVATE KEY block, or DER PKCS#8", where)
	}
	return key, nil
}

// signer narrows a parsed key to a crypto.Signer, the shape every supported
// key type (RSA, ECDSA, Ed25519) has.
func signer(k any, where string) (crypto.Signer, error) {
	s, ok := k.(crypto.Signer)
	if !ok {
		return nil, fmt.Errorf("%s: a private key of type %T is not supported", where, k)
	}
	return s, nil
}

func encryptedKey(where string) error {
	return fmt.Errorf("%s holds an encrypted private key, which cannot be read here: decrypt it to a file only you can read (openssl pkey -in <file> -out <plain>) and name that file instead", where)
}

// ---- helpers -------------------------------------------------------------------

// publicKeyEqual is the Equal method every standard public key type has.
type publicKeyEqual interface {
	Equal(crypto.PublicKey) bool
}

// matches reports whether c is the certificate of key.
func matches(key crypto.Signer, c *x509.Certificate) bool {
	pub, ok := key.Public().(publicKeyEqual)
	return ok && pub.Equal(c.PublicKey)
}

// matchesKey reports whether a decoded key is the same key as want.
func matchesKey(got any, want crypto.Signer) bool {
	s, ok := got.(crypto.Signer)
	if !ok {
		return false
	}
	pub, ok := s.Public().(publicKeyEqual)
	return ok && pub.Equal(want.Public())
}

// dedupe drops repeated certificates, keeping the first of each.
func dedupe(certs []*x509.Certificate) []*x509.Certificate {
	seen := map[[32]byte]bool{}
	var out []*x509.Certificate
	for _, c := range certs {
		fp := sha256.Sum256(c.Raw)
		if !seen[fp] {
			seen[fp] = true
			out = append(out, c)
		}
	}
	return out
}

// sameCerts reports whether got holds exactly want, in order.
func sameCerts(got, want []*x509.Certificate) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if !bytes.Equal(got[i].Raw, want[i].Raw) {
			return false
		}
	}
	return true
}

func readBackError(store string, err error) error {
	if err == nil {
		err = fmt.Errorf("its content differs from what was encoded")
	}
	return fmt.Errorf("tls.stores.%s: the encoded store did not read back as built (%v); this is a bug in solmq-conn-util -- please report it", store, err)
}

// entry describes c for a Result.
func entry(alias string, c *x509.Certificate, key bool) Entry {
	return Entry{Alias: alias, Subject: subject(c), NotAfter: c.NotAfter, SHA256: fingerprint(c.Raw), Key: key}
}

// fingerprint is the SHA-256 of der as keytool prints it: colon-separated
// upper-case hex pairs.
func fingerprint(der []byte) string {
	sum := sha256.Sum256(der)
	var sb strings.Builder
	for i, b := range sum {
		if i > 0 {
			sb.WriteByte(':')
		}
		fmt.Fprintf(&sb, "%02X", b)
	}
	return sb.String()
}

// subject is c's subject made safe to print: a certificate is untrusted input,
// so control characters are replaced rather than handed to a terminal.
func subject(c *x509.Certificate) string {
	return strings.Map(func(r rune) rune {
		if r < 0x20 || r == 0x7f {
			return '?'
		}
		return r
	}, c.Subject.String())
}

// day formats t as the date it falls on, in UTC.
func day(t time.Time) string { return t.UTC().Format("2006-01-02") }
