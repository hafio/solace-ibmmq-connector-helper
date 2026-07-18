// Package tls computes the automatic TLS / mTLS wiring shared by Solace and
// IBM MQ connections. One truststore + one keystore (from defaults.yaml) are
// referenced by every TLS connection — as written in defaults.yaml for `config`,
// or at the mounted MountDir for `deploy` (see StorePath).
package tls

import (
	"strings"

	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-gen/internal/spec"
)

// MountDir is where the shared stores are mounted in the container.
const MountDir = "/app/external/classpath/truststores"

// KV is one ordered scalar property (Key: Val).
type KV struct{ Key, Val string }

// MountPath returns the in-container path for a store file, using its base name.
// It strips both '/' and '\' regardless of host OS (not the OS-dependent
// filepath.Base), so a Windows-authored defaults.yaml path resolves to the same
// mounted base name when the CLI runs on Linux/Mac, matching the stores-Secret
// data key so the bundle location always points at a file the volume mounts.
func MountPath(storeFile string) string {
	return MountDir + "/" + base(storeFile)
}

// base returns the final path element, splitting on both '/' and '\'.
func base(p string) string {
	p = strings.ReplaceAll(p, "\\", "/")
	if i := strings.LastIndexByte(p, '/'); i >= 0 {
		return p[i+1:]
	}
	return p
}

// StorePath is where generated config references a store file: the path exactly as
// written in defaults.yaml when mount is false (`config`, so the connector can run
// wherever those files already live), or the in-container MountPath when mount is
// true (`deploy` mounts the stores Secret at MountDir).
func StorePath(storeFile string, mount bool) string {
	if mount {
		return MountPath(storeFile)
	}
	return storeFile
}

// SolaceProps returns the ordered tool-managed Solace api-properties for a TLS
// (tcps://) binder. keyAlias != "" adds the mTLS keystore selection. Missing
// stores are skipped (validate warns) so the connection falls back to the JVM
// default trust store rather than emitting broken references.
func SolaceProps(d *spec.Defaults, keyAlias string, mount bool) []KV {
	var out []KV
	if ts := d.TLS.Truststore; ts != nil {
		out = append(out,
			KV{"SSL_VALIDATE_CERTIFICATE", "true"},
			KV{"SSL_TRUST_STORE", StorePath(ts.File, mount)},
			KV{"SSL_TRUST_STORE_PASSWORD", ts.Password},
			KV{"SSL_TRUST_STORE_FORMAT", ts.Type},
		)
	}
	if keyAlias != "" {
		if ks := d.TLS.Keystore; ks != nil {
			out = append(out,
				KV{"SSL_KEY_STORE", StorePath(ks.File, mount)},
				KV{"SSL_KEY_STORE_PASSWORD", ks.Password},
				KV{"SSL_KEY_STORE_FORMAT", ks.Type},
				KV{"SSL_PRIVATE_KEY_ALIAS", keyAlias},
			)
		}
	}
	return out
}
