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

// Configured is the one definition of a usable store: created, or naming its
// PKCS12. Nil, an empty block and an existing block without a path are not;
// Path is the existing PKCS12, empty for a created store.
func TestStoreConfigured(t *testing.T) {
	for _, c := range []struct {
		name    string
		s       *Store
		want    bool
		created bool
		path    string
	}{
		{"nil", nil, false, false, ""},
		{"no source", &Store{PasswordEnv: "PW"}, false, false, ""},
		{"existing without a path", &Store{Existing: &StoreExisting{}}, false, false, ""},
		{"existing", &Store{Existing: &StoreExisting{PKCS12: "./t.p12"}}, true, false, "./t.p12"},
		{"created", &Store{Create: &StoreSources{JKS: "./old.jks"}}, true, true, ""},
	} {
		if got := c.s.Configured(); got != c.want {
			t.Errorf("%s: Configured() = %v, want %v", c.name, got, c.want)
		}
		if c.s.Created() != c.created || c.s.Path() != c.path {
			t.Errorf("%s: Created() = %v, Path() = %q; want %v, %q", c.name, c.s.Created(), c.s.Path(), c.created, c.path)
		}
	}
}

// TestTLSConfigStoreLookup pins the store lookups every reader uses: an empty
// or undefined name is no store, and names list in sorted order.
func TestTLSConfigStoreLookup(t *testing.T) {
	c := TLSConfig{Stores: map[string]*Store{"solace-ca": {Existing: &StoreExisting{PKCS12: "s.p12"}}, "mq-ca": {Existing: &StoreExisting{PKCS12: "m.p12"}}}}
	if c.Store("") != nil || c.Store("nope") != nil {
		t.Error("an empty or undefined name must be no store")
	}
	if c.Store("mq-ca").Path() != "m.p12" {
		t.Errorf("Store(mq-ca) = %+v", c.Store("mq-ca"))
	}
	if got := strings.Join(c.StoreNames(), ","); got != "mq-ca,solace-ca" {
		t.Errorf("StoreNames = %s, want sorted", got)
	}
}

// TestParseTLSIsStrict pins rawTLS.UnmarshalYAML: tls takes only stores:, each
// store, create: block, key entry and existing: block only their own keys in
// their own shapes, and the retired single truststore/keystore and a store's
// retired file/type fail loud instead of being dropped -- a dropped store would
// leave connections trusting the JVM default store while env.yaml says
// otherwise.
func TestParseTLSIsStrict(t *testing.T) {
	for _, c := range []struct{ name, y, want string }{
		{"retired truststore", "tls:\n  truststore:\n    file: t.jks\n", "tls.truststore is no longer read"},
		{"retired keystore", "tls:\n  keystore:\n    file: k.jks\n", "tls.keystore is no longer read"},
		{"unknown key", "tls:\n  storez: {}\n", "tls.storez is not a key"},
		{"retired file", "tls:\n  stores:\n    mq-ca:\n      file: t.jks\n", "tls.stores.mq-ca.file is no longer read: name a PKCS12 you provide with existing: {pkcs12: <file>}, or have the tool build one with create:"},
		{"retired type", "tls:\n  stores:\n    mq-ca:\n      type: JKS\n", "tls.stores.mq-ca.type is no longer read"},
		{"unknown store key", "tls:\n  stores:\n    mq-ca:\n      passwrd-env: X\n", "tls.stores.mq-ca.passwrd-env is not a store key: a store takes password or password-env, and create: or existing:"},
		{"existing not a mapping", "tls:\n  stores:\n    mq-ca:\n      existing: ./t.p12\n", "tls.stores.mq-ca.existing must be a mapping holding pkcs12: <file>, got a scalar"},
		{"unknown existing key", "tls:\n  stores:\n    mq-ca:\n      existing:\n        file: ./t.p12\n", "tls.stores.mq-ca.existing.file is not an existing key"},
		{"create not a mapping", "tls:\n  stores:\n    mq-ca:\n      create: [a]\n", "tls.stores.mq-ca.create must be a mapping holding certs, keys or jks, got a list"},
		{"unknown create key", "tls:\n  stores:\n    mq-ca:\n      create:\n        cert: a.pem\n", "tls.stores.mq-ca.create.cert is not a create key: create takes certs, keys and jks"},
		{"certs a scalar", "tls:\n  stores:\n    mq-ca:\n      create:\n        certs: a.pem\n", "tls.stores.mq-ca.create.certs must be a list of files or a mapping of alias: file, got a scalar"},
		{"keys a mapping", "tls:\n  stores:\n    mq-ca:\n      create:\n        keys:\n          cert: a.pem\n", "tls.stores.mq-ca.create.keys must be a list of {cert, chain, key} entries, got a mapping"},
		{"key entry a scalar", "tls:\n  stores:\n    mq-ca:\n      create:\n        keys:\n          - a.pem\n", "tls.stores.mq-ca.create.keys[0] must be a mapping of cert, chain and key, got a scalar"},
		{"unknown key-entry key", "tls:\n  stores:\n    mq-ca:\n      create:\n        keys:\n          - certs: a.pem\n", "tls.stores.mq-ca.create.keys[0].certs is not a key-entry key"},
		{"chain a scalar", "tls:\n  stores:\n    mq-ca:\n      create:\n        keys:\n          - cert: a.pem\n            chain: b.pem\n", "tls.stores.mq-ca.create.keys[0].chain must be a list of files, got a scalar"},
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
	d, err := ParseDefaults([]byte("tls:\n  stores:\n    mq-ca:\n      password-env: P\n      existing:\n        pkcs12: t.p12\n    empty:\n"))
	if err != nil {
		t.Fatal(err)
	}
	if st := d.TLS.Store("mq-ca"); st == nil || st.Path() != "t.p12" || st.PasswordEnv != "P" {
		t.Errorf("mq-ca = %+v", st)
	}
	if _, ok := d.TLS.Stores["empty"]; !ok || d.TLS.Store("empty").Configured() {
		t.Error("a store written with no block is defined and not configured (validate reports it)")
	}
}

// TestParseCreateBlock pins how a create: block decodes: certs as a list (in
// order, unnamed) or as a map (in written order, named), one key entry with
// its chain, a jks -- and a create: written with nothing in it is no create
// block at all, which validate reports as a store with no source.
func TestParseCreateBlock(t *testing.T) {
	d, err := ParseDefaults([]byte(`tls:
  stores:
    list:
      create:
        certs:
          - ./b.pem
          - ./a.pem
    named:
      create:
        certs:
          zeta: ./z.pem
          alpha: ./a.pem
        jks: ./old.jks
    key:
      create:
        keys:
          - cert: ./c.pem
            chain:
              - ./i.pem
            key: ./c.key
    bare:
      create:
`))
	if err != nil {
		t.Fatal(err)
	}
	list := d.TLS.Store("list").Create.Certs
	if list.Named || len(list.Items) != 2 || list.Items[0] != (CertSource{File: "./b.pem"}) || list.Items[1] != (CertSource{File: "./a.pem"}) {
		t.Errorf("list certs = %+v", list)
	}
	named := d.TLS.Store("named").Create
	if !named.Certs.Named || len(named.Certs.Items) != 2 || named.Certs.Items[0] != (CertSource{Name: "zeta", File: "./z.pem"}) || named.JKS != "./old.jks" {
		t.Errorf("named = %+v", named)
	}
	k := d.TLS.Store("key").Create.Keys
	if len(k) != 1 || k[0].Cert != "./c.pem" || strings.Join(k[0].Chain, ",") != "./i.pem" || k[0].Key != "./c.key" {
		t.Errorf("keys = %+v", k)
	}
	if d.TLS.Store("bare").Created() {
		t.Error("create: with nothing in it must decode as no create block")
	}
}

// TestStoreKeyListsAreSorted pins the key lists the generator page is checked
// against.
func TestStoreKeyListsAreSorted(t *testing.T) {
	for _, c := range []struct {
		got  []string
		want string
	}{
		{StoreKeys(), "create,existing,password,password-env"},
		{StoreCreateKeys(), "certs,jks,keys"},
		{StoreKeyEntryKeys(), "cert,chain,key"},
		{StoreExistingKeys(), "pkcs12"},
	} {
		if got := strings.Join(c.got, ","); got != c.want {
			t.Errorf("keys = %s, want %s", got, c.want)
		}
	}
}
