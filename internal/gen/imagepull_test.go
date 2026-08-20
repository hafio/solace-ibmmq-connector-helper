package gen

import (
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"

	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/spec"
)

// decodeAuths unwraps the double-encoded payload: base64 of the JSON document
// kubernetes.io/dockerconfigjson Secrets carry.
func decodeAuths(t *testing.T, payload string) map[string]map[string]string {
	t.Helper()
	raw, err := base64.StdEncoding.DecodeString(payload)
	if err != nil {
		t.Fatalf("payload is not base64: %v", err)
	}
	var doc struct {
		Auths map[string]map[string]string `json:"auths"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatalf("payload is not the expected JSON: %v (%s)", err, raw)
	}
	return doc.Auths
}

// TestDockerConfigJSON pins the payload shape the kubelet reads: an auths map
// keyed by registry, carrying the account plus the base64 "user:password" the
// engines actually send. Both the Docker Hub fallback key and a private
// registry are covered, since getting the key wrong means the credential is
// simply never found.
func TestDockerConfigJSON(t *testing.T) {
	for _, tt := range []struct {
		name, registry string
	}{
		{"private registry", "registry.internal:5000"},
		{"docker hub fallback", spec.DefaultRegistry},
	} {
		t.Run(tt.name, func(t *testing.T) {
			payload, err := dockerConfigJSON(tt.registry, "svc", "hunter2")
			if err != nil {
				t.Fatal(err)
			}
			auths := decodeAuths(t, payload)
			entry, ok := auths[tt.registry]
			if !ok {
				t.Fatalf("auths has no entry for %q: %v", tt.registry, auths)
			}
			if entry["username"] != "svc" || entry["password"] != "hunter2" {
				t.Errorf("account = %v", entry)
			}
			if want := base64.StdEncoding.EncodeToString([]byte("svc:hunter2")); entry["auth"] != want {
				t.Errorf("auth = %q, want %q", entry["auth"], want)
			}
		})
	}
}

// TestDockerConfigJSONEscapesAwkwardValues is why the document is marshalled
// rather than concatenated: a password carrying a quote or a backslash must end
// up as data, not reshape the JSON around it.
func TestDockerConfigJSONEscapesAwkwardValues(t *testing.T) {
	pass := `he said "hi"\ and {"auths":{}}`
	payload, err := dockerConfigJSON("registry.internal", "svc", pass)
	if err != nil {
		t.Fatal(err)
	}
	auths := decodeAuths(t, payload)
	if got := auths["registry.internal"]["password"]; got != pass {
		t.Errorf("password round-trip = %q, want %q", got, pass)
	}
	if len(auths) != 1 {
		t.Errorf("the password reshaped the document: %v", auths)
	}
}

// TestResolvePullSecret covers the three ways the block resolves, including the
// reference-only case that must not touch the environment at all.
func TestResolvePullSecret(t *testing.T) {
	img := &spec.Image{Repo: "registry.internal", Name: "c", Tag: "1", User: "svc", PassEnv: "REGISTRY_PASSWORD"}

	t.Run("reference only: no payload, no environment read", func(t *testing.T) {
		read := false
		ps, err := resolvePullSecret(&spec.ImagePullSecret{Name: "regcred"}, img,
			Resolver{Env: func(string) (string, bool) { read = true; return "", true }})
		if err != nil {
			t.Fatal(err)
		}
		if ps.Name != "regcred" || ps.DockerConfigJSON != "" {
			t.Errorf("ps = %+v, want the name alone", ps)
		}
		if read {
			t.Error("referencing a Secret must not read the registry password")
		}
	})

	// Both halves are literal/-env pairs, so all four combinations have to reach
	// the same payload. The -env forms are the ones a committed env.yaml should
	// use, and the literal forms are what every other credential position here
	// also allows -- neither is a special case.
	t.Run("both credentials in their -env form", func(t *testing.T) {
		envImg := &spec.Image{Repo: "registry.internal", Name: "c", Tag: "1",
			UserEnv: "REGISTRY_USER", PassEnv: "REGISTRY_PASSWORD"}
		ps, err := resolvePullSecret(&spec.ImagePullSecret{Name: "regcred", Create: true}, envImg,
			Resolver{Env: func(k string) (string, bool) {
				switch k {
				case "REGISTRY_USER":
					return "svc", true
				case "REGISTRY_PASSWORD":
					return "hunter2", true
				}
				return "", false
			}})
		if err != nil {
			t.Fatal(err)
		}
		got := decodeAuths(t, ps.DockerConfigJSON)["registry.internal"]
		if got["username"] != "svc" || got["password"] != "hunter2" {
			t.Errorf("auths entry = %v, want the values read from the environment", got)
		}
	})

	t.Run("both credentials in their literal form", func(t *testing.T) {
		litImg := &spec.Image{Repo: "registry.internal", Name: "c", Tag: "1", User: "svc", Pass: "hunter2"}
		// No environment access at all: a literal pair must not need one.
		ps, err := resolvePullSecret(&spec.ImagePullSecret{Name: "regcred", Create: true}, litImg, Resolver{})
		if err != nil {
			t.Fatal(err)
		}
		got := decodeAuths(t, ps.DockerConfigJSON)["registry.internal"]
		if got["username"] != "svc" || got["password"] != "hunter2" {
			t.Errorf("auths entry = %v", got)
		}
	})

	t.Run("an unset user-env fails, naming that variable", func(t *testing.T) {
		envImg := &spec.Image{Repo: "registry.internal", Name: "c", Tag: "1",
			UserEnv: "REGISTRY_USER", PassEnv: "REGISTRY_PASSWORD"}
		_, err := resolvePullSecret(&spec.ImagePullSecret{Name: "regcred", Create: true}, envImg,
			Resolver{Env: func(k string) (string, bool) { return "pw", k == "REGISTRY_PASSWORD" }})
		if err == nil {
			t.Fatal("want an error when the user variable is unset")
		}
		if !strings.Contains(err.Error(), "REGISTRY_USER") {
			t.Errorf("error %q should name the user variable, not just the password one", err)
		}
	})

	t.Run("create builds the payload", func(t *testing.T) {
		ps, err := resolvePullSecret(&spec.ImagePullSecret{Name: "regcred", Create: true}, img,
			Resolver{Env: func(k string) (string, bool) { return "hunter2", k == "REGISTRY_PASSWORD" }})
		if err != nil {
			t.Fatal(err)
		}
		if auths := decodeAuths(t, ps.DockerConfigJSON); auths["registry.internal"]["password"] != "hunter2" {
			t.Errorf("auths = %v", auths)
		}
	})

	t.Run("create with the variable unset fails, naming it", func(t *testing.T) {
		_, err := resolvePullSecret(&spec.ImagePullSecret{Name: "regcred", Create: true}, img,
			Resolver{Env: func(string) (string, bool) { return "", false }})
		if err == nil {
			t.Fatal("want an error when the variable is unset")
		}
		if !strings.Contains(err.Error(), "REGISTRY_PASSWORD") {
			t.Errorf("error %q should name the variable", err)
		}
	})

	t.Run("create with no environment access fails loudly", func(t *testing.T) {
		if _, err := resolvePullSecret(&spec.ImagePullSecret{Name: "regcred", Create: true}, img, Resolver{}); err == nil {
			t.Fatal("want an error when the resolver has no environment access")
		}
	})

	// validate rejects create without both credential fields, so reaching here
	// means a caller skipped it. The guard exists so that is an error rather
	// than a Secret built from an empty account.
	t.Run("create with no image block at all", func(t *testing.T) {
		if _, err := resolvePullSecret(&spec.ImagePullSecret{Name: "regcred", Create: true}, nil,
			Resolver{Env: func(string) (string, bool) { return "pw", true }}); err == nil {
			t.Fatal("want an error when the image block is missing entirely")
		}
	})
	t.Run("create with a partial image block", func(t *testing.T) {
		for _, partial := range []*spec.Image{
			{Name: "c", Tag: "1", PassEnv: "REGISTRY_PASSWORD"}, // no user
			{Name: "c", Tag: "1", User: "svc"},                  // no pass in either form
		} {
			if _, err := resolvePullSecret(&spec.ImagePullSecret{Name: "regcred", Create: true}, partial,
				Resolver{Env: func(string) (string, bool) { return "pw", true }}); err == nil {
				t.Errorf("want an error for %+v", partial)
			}
		}
	})
}

// kubeEnvWithPull builds an env.yaml carrying the image block plus a kubernetes
// section whose secrets.image-pull is in the requested mode. creds controls
// whether the registry account is present, so the create-without-credentials
// path can be driven too.
func kubeEnvWithPull(mode string, creds bool) []byte {
	env := "image:\n  repo: registry.internal\n  name: c\n  tag: v1\n"
	if creds {
		env += "  user: svc\n  pass-env: REGISTRY_PASSWORD\n"
	}
	env += "kubernetes:\n  command: kubectl\n  deployment:\n    name: solmq\n    namespace: ns\n"
	if mode != "" {
		env += "  secrets:\n    image-pull:\n      name: regcred\n"
		if mode == "create" {
			env += "      create: true\n"
		}
	}
	return []byte(env)
}

// TestGenerateKubernetesImagePull covers the wiring between the image-pull
// config and the rendered manifest -- the integration point TestResolvePullSecret
// skips by calling the helper directly. All three states are asserted on the
// manifest itself, because that is the only place the difference between
// "reference" and "create" is observable.
func TestGenerateKubernetesImagePull(t *testing.T) {
	run := func(mode string, creds bool, env func(string) (string, bool)) (string, []Issue) {
		req := Request{Env: &File{Name: "env.yaml", Data: kubeEnvWithPull(mode, creds)}, Workflows: synthWorkflowFiles(1)}
		out, errs, _ := GenerateKubernetes(req, Resolver{Env: env, Rand: fixedStatusRand})
		return out, errs
	}
	withPass := func(k string) (string, bool) { return "hunter2", k == "REGISTRY_PASSWORD" }

	t.Run("no block: neither reference nor Secret", func(t *testing.T) {
		out, errs := run("", false, withPass)
		if len(errs) > 0 {
			t.Fatalf("unexpected errors: %v", errs)
		}
		if strings.Contains(out, "imagePullSecrets") || strings.Contains(out, "dockerconfigjson") {
			t.Errorf("nothing should be emitted:\n%s", out)
		}
	})

	t.Run("reference: the entry, and no Secret", func(t *testing.T) {
		out, errs := run("reference", false, withPass)
		if len(errs) > 0 {
			t.Fatalf("unexpected errors: %v", errs)
		}
		if !strings.Contains(out, "imagePullSecrets:\n        - name: regcred\n") {
			t.Errorf("imagePullSecrets entry missing:\n%s", out)
		}
		if strings.Contains(out, "dockerconfigjson") {
			t.Errorf("a reference must not render a Secret -- it would overwrite the operator's:\n%s", out)
		}
	})

	t.Run("create: the entry and the Secret, carrying the resolved password", func(t *testing.T) {
		out, errs := run("create", true, withPass)
		if len(errs) > 0 {
			t.Fatalf("unexpected errors: %v", errs)
		}
		if !strings.Contains(out, "imagePullSecrets:\n        - name: regcred\n") {
			t.Errorf("imagePullSecrets entry missing:\n%s", out)
		}
		if !strings.Contains(out, "type: kubernetes.io/dockerconfigjson\n") {
			t.Fatalf("dockerconfigjson Secret missing:\n%s", out)
		}
		// Pull the payload back out and confirm it round-trips to the real
		// account -- and that the password is nowhere else in the manifest.
		const key = ".dockerconfigjson: "
		i := strings.Index(out, key)
		payload := strings.SplitN(out[i+len(key):], "\n", 2)[0]
		auths := decodeAuths(t, payload)
		if got := auths["registry.internal"]; got["username"] != "svc" || got["password"] != "hunter2" {
			t.Errorf("auths entry = %v", got)
		}
		if strings.Contains(strings.ReplaceAll(out, payload, ""), "hunter2") {
			t.Errorf("the registry password leaked outside the Secret payload:\n%s", out)
		}
	})

	t.Run("create with the variable unset fails the generate", func(t *testing.T) {
		_, errs := run("create", true, func(string) (string, bool) { return "", false })
		if len(errs) == 0 {
			t.Fatal("want an issue when the registry password is not exported")
		}
		if !strings.Contains(errs[0].Msg, "REGISTRY_PASSWORD") {
			t.Errorf("issue %q should name the variable", errs[0].Msg)
		}
	})
}
