package spec

import (
	"strings"
	"testing"
)

func TestLeaderElectionEffectiveMode(t *testing.T) {
	tests := []struct {
		name string
		mode string
		want string
	}{
		{"empty", "", LeaderStandalone},
		{"standalone", LeaderStandalone, LeaderStandalone},
		{"active_active", LeaderActiveActive, LeaderActiveActive},
		{"active_standby", LeaderActiveStby, LeaderActiveStby},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			le := LeaderElection{Mode: tt.mode}
			if got := le.EffectiveMode(); got != tt.want {
				t.Errorf("EffectiveMode() = %q, want %q", got, tt.want)
			}
		})
	}
}

// HasFile is the one definition of a configured store: nil and a block without
// a file are both unconfigured.
func TestStoreHasFile(t *testing.T) {
	for _, c := range []struct {
		name string
		s    *Store
		want bool
	}{
		{"nil", nil, false},
		{"no file", &Store{PasswordEnv: "PW", Type: "JKS"}, false},
		{"file", &Store{File: "./t.jks"}, true},
	} {
		if got := c.s.HasFile(); got != c.want {
			t.Errorf("%s: HasFile() = %v, want %v", c.name, got, c.want)
		}
	}
}

// TestTLSConfigStoreLookup pins the store lookups every reader uses: an empty
// or undefined name is no store, and names list in sorted order.
func TestTLSConfigStoreLookup(t *testing.T) {
	c := TLSConfig{Stores: map[string]*Store{"solace-ca": {File: "s.jks"}, "mq-ca": {File: "m.jks"}}}
	if c.Store("") != nil || c.Store("nope") != nil {
		t.Error("an empty or undefined name must be no store")
	}
	if c.Store("mq-ca").File != "m.jks" {
		t.Errorf("Store(mq-ca) = %+v", c.Store("mq-ca"))
	}
	if got := strings.Join(c.StoreNames(), ","); got != "mq-ca,solace-ca" {
		t.Errorf("StoreNames = %s, want sorted", got)
	}
}

// TestParseTLSIsStrict pins rawTLS.UnmarshalYAML: tls takes only stores:, each
// store only its four keys, and the retired single truststore/keystore fail
// loud instead of being dropped -- a dropped store would leave connections
// trusting the JVM default store while env.yaml says otherwise.
func TestParseTLSIsStrict(t *testing.T) {
	for _, c := range []struct{ name, y, want string }{
		{"retired truststore", "tls:\n  truststore:\n    file: t.jks\n", "tls.truststore is no longer read"},
		{"retired keystore", "tls:\n  keystore:\n    file: k.jks\n", "tls.keystore is no longer read"},
		{"unknown key", "tls:\n  storez: {}\n", "tls.storez is not a key"},
		{"unknown store key", "tls:\n  stores:\n    mq-ca:\n      file: t.jks\n      passwrd-env: X\n", "tls.stores.mq-ca.passwrd-env is not a store key"},
		{"not a mapping", "tls: [a]\n", "tls must be a mapping"},
		{"stores not a mapping", "tls:\n  stores: [a]\n", "tls.stores must be a mapping"},
	} {
		t.Run(c.name, func(t *testing.T) {
			_, err := ParseDefaults([]byte(c.y))
			if err == nil || !strings.Contains(err.Error(), c.want) {
				t.Errorf("err = %v, want it to contain %q", err, c.want)
			}
		})
	}
	d, err := ParseDefaults([]byte("tls:\n  stores:\n    mq-ca:\n      file: t.jks\n      password-env: P\n      type: JKS\n    empty:\n"))
	if err != nil {
		t.Fatal(err)
	}
	if st := d.TLS.Store("mq-ca"); st == nil || st.File != "t.jks" || st.PasswordEnv != "P" || st.Type != "JKS" {
		t.Errorf("mq-ca = %+v", st)
	}
	if _, ok := d.TLS.Stores["empty"]; !ok || d.TLS.Store("empty").HasFile() {
		t.Error("a store written with no block is defined and has no file (validate reports it)")
	}
}
