package tls

import (
	"testing"

	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/spec"
)

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
		Truststore: &spec.Store{File: `.\certs\truststore.jks`, Password: "${T}", Type: "JKS"},
		Keystore:   &spec.Store{File: `.\certs\keystore.jks`, Password: "${K}", Type: "JKS"},
	}}
	got := map[string]string{}
	for _, kv := range SolaceProps(d, "solace-client", true) {
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
		Truststore: &spec.Store{File: "./certs/truststore.jks", Password: "${T}", Type: "JKS"},
	}}
	got := map[string]string{}
	for _, kv := range SolaceProps(d, "", false) {
		got[kv.Key] = kv.Val
	}
	if got["SSL_TRUST_STORE"] != "./certs/truststore.jks" {
		t.Errorf("mount=false should keep the raw defaults path, got %q", got["SSL_TRUST_STORE"])
	}
}
