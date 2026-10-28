package validate

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/spec"
)

// secretDataKeyRE is what a Secret data key may look like (the API server's own
// rule for a ConfigMap or Secret key).
var secretDataKeyRE = regexp.MustCompile(`^[-._a-zA-Z0-9]+$`)

// checkStoreDefinition checks one tls.stores entry's shape: at most one
// host-side source -- create: or existing.pkcs12 -- and optionally a
// kubernetes Secret holding the store (existing.secret), at least one of the
// three; and for create: sources the PKCS12 writer can turn into one store.
// Nothing here opens a file -- a source that is missing, unreadable or not a
// certificate is reported when the validate verb or deploy builds the store
// (gen.Validate, gen.BuildStore).
func checkStoreDefinition(add func(string, string, ...any), name string, st *spec.Store) {
	if st == nil || (st.Create == nil && st.Existing == nil) {
		add(fileEnv, "tls.stores.%s needs create: (the tool builds it) or existing: (a PKCS12 you provide, or a kubernetes Secret holding one)", name)
		return
	}
	if ex := st.Existing; ex != nil {
		switch {
		case st.Create != nil && ex.PKCS12 != "":
			add(fileEnv, "tls.stores.%s takes create: or existing.pkcs12, not both: one host-side source per store (create: may sit beside existing.secret, which kubernetes reads instead)", name)
			return
		case ex.PKCS12 == "" && ex.Secret == nil && st.Create == nil:
			add(fileEnv, "tls.stores.%s.existing names neither pkcs12 nor secret: name the PKCS12 you hold on this host (pkcs12: <file>), the kubernetes Secret holding it (secret: {name: <Secret>}), or both", name)
			return
		}
		if sec := ex.Secret; sec != nil {
			checkSecretName(add, "tls.stores."+name+".existing.secret.name", sec.Name)
			if k := sec.Key; k != "" && (len(k) > 253 || !secretDataKeyRE.MatchString(k)) {
				add(fileEnv, "tls.stores.%s.existing.secret.key %q is not a valid Secret data key: letters, digits, '-', '_' and '.', at most 253 characters", name, k)
			}
		}
	}
	if st.Create == nil {
		return
	}
	c := st.Create
	if st.Secret().Empty() {
		add(fileEnv, "tls.stores.%s is created and encrypted with its own password, but sets none; add password-env", name)
	}
	certs, keys := len(c.Certs.Items) > 0, len(c.Keys) > 0
	switch {
	case !certs && !keys && c.JKS == "":
		add(fileEnv, "tls.stores.%s.create names no source: list certs, give one keys entry, or name a jks", name)
	case keys && (certs || c.JKS != ""):
		add(fileEnv, "tls.stores.%s.create mixes trusted certificates and a key: a store is a truststore or a keystore -- define a second store for the other", name)
	}
	if len(c.Keys) > 1 {
		add(fileEnv, "tls.stores.%s.create.keys has %d entries: a created keystore holds exactly one key; define one store per identity", name, len(c.Keys))
	}
	for i, k := range c.Keys {
		if k.Cert == "" || k.Key == "" {
			add(fileEnv, "tls.stores.%s.create.keys[%d] needs cert and key: the certificate and the private key it belongs to", name, i)
		}
		for j, ch := range k.Chain {
			if ch == "" {
				add(fileEnv, "tls.stores.%s.create.keys[%d].chain[%d] is empty: name a certificate file", name, i, j)
			}
		}
	}
	seen := map[string]bool{}
	for i, cs := range c.Certs.Items {
		where := fmt.Sprintf("tls.stores.%s.create.certs[%d]", name, i)
		if c.Certs.Named {
			where = fmt.Sprintf("tls.stores.%s.create.certs.%s", name, cs.Name)
			alias := strings.ToLower(cs.Name)
			switch {
			case alias == "":
				add(fileEnv, "tls.stores.%s.create.certs has an empty alias: name every certificate", name)
			case seen[alias]:
				add(fileEnv, "tls.stores.%s.create.certs: alias %q appears twice (aliases ignore case)", name, cs.Name)
			}
			seen[alias] = true
		}
		if cs.File == "" {
			add(fileEnv, "%s is empty: name a certificate file", where)
		}
	}
}

// createdRole is what a created store's sources make it, as far as env.yaml
// alone tells: certificates make a truststore, a key entry a keystore, and a
// JKS on its own is known only once deploy opens it.
func createdRole(st *spec.Store) (truststore, keystore bool) {
	if !st.Created() {
		return false, false
	}
	return len(st.Create.Certs.Items) > 0, len(st.Create.Keys) > 0
}

// checkCreatedSelection holds a connection's selections to what its created
// stores can be: a store created from a key holds no trusted certificates, one
// created from certificates holds no key, and the one key of a created keystore
// has no name for key-alias to pick (go-pkcs12 cannot write one; Java calls it
// "1").
func checkCreatedSelection(add func(string, string, ...any), file, where string, s spec.Side, stores spec.TLSConfig) {
	if _, keystore := createdRole(stores.Store(s.Truststore)); keystore {
		add(file, "%s: truststore %q is created from a key, so it holds no trusted certificates; create a second store from the CA certificates and select that", where, s.Truststore)
	}
	if truststore, _ := createdRole(stores.Store(s.Keystore)); truststore {
		add(file, "%s: keystore %q is created from certificates only, so it holds no private key; give it one keys entry", where, s.Keystore)
	}
	if s.KeyAlias != "" && stores.Store(s.Keystore).Created() {
		add(file, "%s: remove key-alias %q: tls.stores.%s is created with exactly one key, which needs no alias (Java names it \"1\")", where, s.KeyAlias, s.Keystore)
	}
}
