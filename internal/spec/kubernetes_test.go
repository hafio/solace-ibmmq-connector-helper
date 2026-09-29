package spec

import (
	"reflect"
	"strings"
	"testing"
)

// parseKube decodes a kubernetes: section through ParseEnv, the path every
// command takes, with a fixed deployment so derived names are predictable.
func parseKube(t *testing.T, secretsAndLibs string) *Kubernetes {
	t.Helper()
	e, err := ParseEnv([]byte("kubernetes:\n  deployment:\n    name: solmq\n    namespace: ns\n" + secretsAndLibs))
	if err != nil {
		t.Fatalf("ParseEnv: %v", err)
	}
	return e.Kubernetes
}

// TestSecretsCreateSpellings covers both spellings create: now takes. true is
// the new one; the mapping form an env.yaml written before names were derived
// still carries must keep meaning "create", so that file keeps deploying --
// with its name decoded only so validate can report it. false, empty and
// absent all mean the tool builds nothing.
func TestSecretsCreateSpellings(t *testing.T) {
	cases := []struct {
		name, block string
		create      bool
		legacyName  string
	}{
		{"create: true", "create: true", true, ""},
		{"the retired mapping form", "create:\n        name: old-name", true, "old-name"},
		{"an empty mapping", "create: {}", true, ""},
		{"create: false", "create: false", false, ""},
		{"create left empty", "create:", false, ""},
		{"only existing", "existing: theirs", false, ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			k := parseKube(t, "  secrets:\n    credentials:\n      "+c.block+"\n    stores:\n      "+c.block+"\n")
			cred, stores := k.Secrets.Credentials, k.Secrets.Stores
			if cred == nil || stores == nil {
				t.Fatal("both blocks should parse")
			}
			if (cred.Create != nil) != c.create || (stores.Create != nil) != c.create {
				t.Fatalf("create = %v/%v, want %v", cred.Create != nil, stores.Create != nil, c.create)
			}
			if c.create && (cred.Create.Name != c.legacyName || stores.Create.Name != c.legacyName) {
				t.Errorf("retired name = %q/%q, want %q", cred.Create.Name, stores.Create.Name, c.legacyName)
			}
		})
	}
	// existing: survives the custom decoding untouched.
	k := parseKube(t, "  secrets:\n    credentials:\n      existing: their-creds\n")
	if got := k.Secrets.Credentials.Existing; got != "their-creds" {
		t.Errorf("existing = %q, want their-creds", got)
	}
}

// TestSecretsCreateRejectsOtherShapes pins that a create: value that is
// neither a boolean nor the retired mapping fails at parse, naming the key --
// silently reading it as false would start the pod without its Secret.
func TestSecretsCreateRejectsOtherShapes(t *testing.T) {
	for _, c := range []struct{ name, secrets, want string }{
		{"a word", "credentials:\n      create: please", `kubernetes.secrets.credentials.create must be true or false, got "please"`},
		{"a list", "credentials:\n      create: [a, b]", "kubernetes.secrets.credentials.create must be true or false, got a list"},
		{"stores, a word", "stores:\n      create: please", `kubernetes.secrets.stores.create must be true or false, got "please"`},
		// The retired mapping is decoded field by field, so a wrong type in it
		// is still a parse error rather than a silently empty name.
		{"a retired mapping with a list for its name", "credentials:\n      create:\n        name: [a, b]", "cannot unmarshal"},
		{"stores, a retired mapping with a list for its name", "stores:\n      create:\n        name: [a, b]", "cannot unmarshal"},
		// The block itself has to be a mapping before create: can be read.
		{"credentials as a bare value", "credentials: true", "cannot unmarshal"},
		{"stores as a bare value", "stores: true", "cannot unmarshal"},
	} {
		t.Run(c.name, func(t *testing.T) {
			_, err := ParseEnv([]byte("kubernetes:\n  secrets:\n    " + c.secrets + "\n"))
			if err == nil || !strings.Contains(err.Error(), c.want) {
				t.Errorf("error = %v, want it to contain %q", err, c.want)
			}
		})
	}
}

// TestSecretsCreateFollowsAnAlias pins that a YAML alias is resolved before
// create: is read, so a shared anchor -- create: *on -- is not reported as the
// wrong shape.
func TestSecretsCreateFollowsAnAlias(t *testing.T) {
	e, err := ParseEnv([]byte("x-on: &on true\nkubernetes:\n  deployment:\n    name: solmq\n  secrets:\n    credentials:\n      create: *on\n"))
	if err != nil {
		t.Fatalf("ParseEnv: %v", err)
	}
	if n, created := e.Kubernetes.CredentialsSecretName(); !created || n != "solmq-credentials" {
		t.Errorf("an aliased create: true = %q/%v, want solmq-credentials created", n, created)
	}
}

// TestDerivedObjectNames pins the fix for two instances in one namespace: every
// object the tool creates is named after deployment.name, so no two instances
// can render -- and later tear down -- the same Secret or claim. A retired name
// key is ignored, and a referenced object keeps the operator's own name.
func TestDerivedObjectNames(t *testing.T) {
	created := parseKube(t, `  secrets:
    credentials:
      create:
        name: shared-creds
    stores:
      create: true
    image-pull:
      name: regcred
      create: true
  libs:
    pvc:
      create:
        name: shared-libs
        nfs: {server: nfs1, path: /libs}
`)
	for _, c := range []struct {
		what      string
		got       string
		gotCreate bool
		want      string
	}{
		{"credentials", first(created.CredentialsSecretName()), second(created.CredentialsSecretName()), "solmq-credentials"},
		{"stores", first(created.StoresSecretName()), second(created.StoresSecretName()), "solmq-stores"},
		{"image-pull", first(created.ImagePullSecretName()), second(created.ImagePullSecretName()), "solmq-image-pull"},
	} {
		if c.got != c.want || !c.gotCreate {
			t.Errorf("%s = %q (created %v), want %q created", c.what, c.got, c.gotCreate, c.want)
		}
	}
	if got := created.LibsPVCName(); got != "solmq-libs" {
		t.Errorf("libs claim = %q, want solmq-libs", got)
	}

	referenced := parseKube(t, `  secrets:
    credentials:
      existing: their-creds
    stores:
      existing: their-tls
    image-pull:
      name: regcred
  libs:
    pvc:
      existing: their-libs
`)
	for _, c := range []struct {
		what, got string
		created   bool
		want      string
	}{
		{"credentials", first(referenced.CredentialsSecretName()), second(referenced.CredentialsSecretName()), "their-creds"},
		{"stores", first(referenced.StoresSecretName()), second(referenced.StoresSecretName()), "their-tls"},
		{"image-pull", first(referenced.ImagePullSecretName()), second(referenced.ImagePullSecretName()), "regcred"},
	} {
		if c.got != c.want || c.created {
			t.Errorf("%s = %q (created %v), want the operator's %q, not created", c.what, c.got, c.created, c.want)
		}
	}
	if got := referenced.LibsPVCName(); got != "" {
		t.Errorf("an existing claim is not created, got %q", got)
	}

	none := parseKube(t, "")
	if n, c := none.CredentialsSecretName(); n != "" || c {
		t.Errorf("no credentials block = %q/%v, want nothing", n, c)
	}
	if n, c := none.ImagePullSecretName(); n != "" || c {
		t.Errorf("no image-pull block = %q/%v, want nothing", n, c)
	}
}

// TestCreatedNames pins the list remove uses to recognise its own objects and
// validate uses to hold each name to a DNS-1123 label: the Deployment and its
// ConfigMap always, then exactly the Secrets and claim the config asks for.
func TestCreatedNames(t *testing.T) {
	all := parseKube(t, `  secrets:
    credentials:
      create: true
    stores:
      create: true
    image-pull:
      create: true
  libs:
    pvc:
      create:
        nfs: {server: nfs1, path: /libs}
`)
	want := []string{"solmq", "solmq-config", "solmq-credentials", "solmq-stores", "solmq-image-pull", "solmq-libs"}
	if got := all.CreatedNames(); !reflect.DeepEqual(got, want) {
		t.Errorf("CreatedNames = %v, want %v", got, want)
	}

	referenced := parseKube(t, "  secrets:\n    credentials:\n      existing: their-creds\n    image-pull:\n      name: regcred\n")
	if got, want := referenced.CreatedNames(), []string{"solmq", "solmq-config"}; !reflect.DeepEqual(got, want) {
		t.Errorf("CreatedNames = %v, want %v -- a referenced object is not the tool's", got, want)
	}
}

func first(s string, _ bool) string { return s }
func second(_ string, b bool) bool  { return b }
