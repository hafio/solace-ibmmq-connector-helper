// Package tls computes the TLS / mTLS wiring of Solace and IBM MQ connections
// from the named stores in env.yaml (tls.stores.<name>). Each connection selects
// its own truststore and keystore by name, and every selected store is
// referenced as written in env.yaml for `config`, or at its mounted path under
// MountDir for `deploy` (see StorePath).
package tls

import (
	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/spec"
)

// MountDir is where the stores are mounted in the container. It aliases
// spec.DefaultStoresMountPath so the application.yml store path and the
// docker bind-mount, podman secret-mount and kubernetes volume targets share
// one source of truth.
const MountDir = spec.DefaultStoresMountPath

// KV is one ordered scalar property (Key: Val).
type KV struct{ Key, Val string }

// MountPath returns the in-container path of the store named store. The store's
// name -- not its file name -- is the file name inside the container and its key
// in a stores Secret, so two stores whose files share a name cannot collide, and
// one file defined under two names is simply mounted twice.
func MountPath(store string) string {
	return MountDir + "/" + store
}

// StorePath is where generated config references a store's file: the path
// exactly as written in env.yaml when mount is false (`config`, so the connector
// can run wherever those files already live), or the in-container MountPath of
// the store when mount is true (`deploy` mounts every selected store there).
func StorePath(store, file string, mount bool) string {
	if mount {
		return MountPath(store)
	}
	return file
}

// Selection is the stores one connection selects: the truststore it validates
// its broker against, the keystore holding the client certificate it presents,
// and the alias of one key within that keystore. Each is "" when not selected.
type Selection struct {
	Truststore string
	Keystore   string
	KeyAlias   string
}

// SolaceProps returns the ordered tool-managed Solace api-properties for a TLS
// (tcps://) binder or session: the wiring of the stores it selects and nothing
// else. Validation policy -- SSL_VALIDATE_CERTIFICATE and its _HOST/_DATE
// siblings -- is not written: the Solace API validates by default, and a
// connection that wants otherwise says so in its own api-properties, which then
// reach the API unchanged. No truststore selected writes no trust keys, so the
// API validates against the JVM default trust store; a keystore adds the client
// certificate, and SSL_PRIVATE_KEY_ALIAS only when an alias picks one key. A
// selected store without a file is skipped (validate refuses one) rather than
// emitting a broken reference, and a store without a type omits its FORMAT key,
// leaving the API's own default.
//
// secretRef records the password of the store it is given by name and returns
// the placeholder to emit, so the password itself never reaches api-properties.
func SolaceProps(d *spec.Defaults, sel Selection, mount bool, secretRef func(string, spec.Cred) string) []KV {
	var out []KV
	if ts := d.TLS.Store(sel.Truststore); ts.HasFile() {
		out = append(out,
			KV{"SSL_TRUST_STORE", StorePath(sel.Truststore, ts.File, mount)},
			KV{"SSL_TRUST_STORE_PASSWORD", secretRef(sel.Truststore, ts.Secret())},
		)
		if ts.Type != "" {
			out = append(out, KV{"SSL_TRUST_STORE_FORMAT", ts.Type})
		}
	}
	if ks := d.TLS.Store(sel.Keystore); ks.HasFile() {
		out = append(out,
			KV{"SSL_KEY_STORE", StorePath(sel.Keystore, ks.File, mount)},
			KV{"SSL_KEY_STORE_PASSWORD", secretRef(sel.Keystore, ks.Secret())},
		)
		if ks.Type != "" {
			out = append(out, KV{"SSL_KEY_STORE_FORMAT", ks.Type})
		}
		if sel.KeyAlias != "" {
			out = append(out, KV{"SSL_PRIVATE_KEY_ALIAS", sel.KeyAlias})
		}
	}
	return out
}
