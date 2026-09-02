// Package tls computes the automatic TLS / mTLS wiring shared by Solace and
// IBM MQ connections. One truststore + one keystore (from the tls section of
// env.yaml) are referenced by every TLS connection -- as written in env.yaml for
// `config`, or at the mounted MountDir for `deploy` (see StorePath).
package tls

import (
	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/spec"
)

// MountDir is where the shared stores are mounted in the container. It aliases
// spec.DefaultStoresMountPath so the application.yml store path and the
// docker/podman bind-mount target share one source of truth.
const MountDir = spec.DefaultStoresMountPath

// KV is one ordered scalar property (Key: Val).
type KV struct{ Key, Val string }

// MountPath returns the in-container path for a store file, using its base name.
// It strips both '/' and '\' regardless of host OS (not the OS-dependent
// filepath.Base), so a Windows-authored env.yaml path resolves to the same
// mounted base name when the CLI runs on Linux/Mac, matching the stores-Secret
// data key so the bundle location always points at a file the volume mounts.
func MountPath(storeFile string) string {
	return MountDir + "/" + spec.BaseName(storeFile)
}

// StorePath is where generated config references a store file: the path exactly as
// written in env.yaml when mount is false (`config`, so the connector can run
// wherever those files already live), or the in-container MountPath when mount is
// true (`deploy` mounts the stores Secret at MountDir).
func StorePath(storeFile string, mount bool) string {
	if mount {
		return MountPath(storeFile)
	}
	return storeFile
}

// Derived mount names for the two shared store passwords. They are declared here
// and re-exported by internal/consolidate so this package and the SSL-bundle
// builder cannot drift into naming the same password two different things. Both
// carry spec.GeneratedNamePrefix, since they are names the tool chose rather
// than ones an operator wrote in a `-env` field. They apply only when the store
// password is a literal: an `-env` store password is mounted under its own
// variable name like any other.
const (
	TruststorePasswordName = spec.GeneratedNamePrefix + "TRUSTSTORE_PASSWORD"
	KeystorePasswordName   = spec.GeneratedNamePrefix + "KEYSTORE_PASSWORD"
)

// SolaceProps returns the ordered tool-managed Solace api-properties for a TLS
// (tcps://) binder. keyAlias != "" adds the mTLS keystore selection. Missing
// stores are skipped (validate warns) so the connection falls back to the JVM
// default trust store rather than emitting broken references.
//
// secretRef records a store password under its stable name and returns the
// placeholder to emit, so the password itself never reaches api-properties.
func SolaceProps(d *spec.Defaults, keyAlias string, mount bool, secretRef func(string, spec.Cred) string) []KV {
	var out []KV
	if ts := d.TLS.Truststore; ts != nil {
		out = append(out,
			KV{"SSL_VALIDATE_CERTIFICATE", "true"},
			KV{"SSL_TRUST_STORE", StorePath(ts.File, mount)},
			KV{"SSL_TRUST_STORE_PASSWORD", secretRef(TruststorePasswordName, ts.Secret())},
			KV{"SSL_TRUST_STORE_FORMAT", ts.Type},
		)
	}
	if keyAlias != "" {
		if ks := d.TLS.Keystore; ks != nil {
			out = append(out,
				KV{"SSL_KEY_STORE", StorePath(ks.File, mount)},
				KV{"SSL_KEY_STORE_PASSWORD", secretRef(KeystorePasswordName, ks.Secret())},
				KV{"SSL_KEY_STORE_FORMAT", ks.Type},
				KV{"SSL_PRIVATE_KEY_ALIAS", keyAlias},
			)
		}
	}
	return out
}
