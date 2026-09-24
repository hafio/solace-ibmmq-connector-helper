package spec

import "testing"

// TestImageRef covers assembling the reference every platform pulls. The digest
// case is the one worth pinning: a sha256: tag has to join with '@', because
// name:sha256:... is not a reference any engine accepts -- and pinning by digest
// is the strongest form of the version pinning this tool asks for elsewhere.
func TestImageRef(t *testing.T) {
	tests := []struct {
		name string
		img  *Image
		want string
	}{
		{"hub image, namespace in the name", &Image{Name: "solace/connector", Tag: "2.14.1"}, "solace/connector:2.14.1"},
		{"private registry", &Image{Repo: "registry.internal:5000", Name: "team/connector", Tag: "2.14.1"}, "registry.internal:5000/team/connector:2.14.1"},
		{"digest joins with @", &Image{Name: "connector", Tag: "sha256:abc123"}, "connector@sha256:abc123"},
		{"trailing slash on repo is not doubled", &Image{Repo: "registry.internal/", Name: "connector", Tag: "1"}, "registry.internal/connector:1"},
		{"no tag (validate rejects it; Ref still renders)", &Image{Name: "connector"}, "connector"},
		{"no name at all", &Image{Tag: "1"}, ""},
		{"nil", nil, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.img.Ref(); got != tt.want {
				t.Errorf("Ref() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestImageRegistry pins the auths key credentials are stored under. The Docker
// Hub case matters because the key is the historical v1 URL, not a hostname --
// get it wrong and the kubelet looks up credentials that are not there.
//
// A Hub namespace must never reach this: solace/connector puts "solace" in the
// name, so Repo stays empty and the lookup falls back to Hub.
func TestImageRegistry(t *testing.T) {
	tests := []struct {
		name string
		img  *Image
		want string
	}{
		{"no repo means Docker Hub", &Image{Name: "solace/connector", Tag: "1"}, DefaultRegistry},
		{"nil", nil, DefaultRegistry},
		{"private registry with a port", &Image{Repo: "registry.internal:5000", Name: "c", Tag: "1"}, "registry.internal:5000"},
		{"trailing slash trimmed", &Image{Repo: "registry.internal/", Name: "c", Tag: "1"}, "registry.internal"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.img.Registry(); got != tt.want {
				t.Errorf("Registry() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestRetiredPerPlatformImageStillParses pins the mechanism behind the loud
// error: the fields stay on the structs so a stale env.yaml is caught by
// validate. Deleting them would make yaml drop the key silently and deploy an
// image the operator never chose.
func TestRetiredPerPlatformImageStillParses(t *testing.T) {
	e, err := ParseEnv([]byte("docker:\n  image: old\npodman:\n  image: old\nkubernetes:\n  deployment:\n    name: c\n    namespace: n\n    image: old\n"))
	if err != nil {
		t.Fatal(err)
	}
	if e.Docker.Image != "old" || e.Podman.Image != "old" || e.Kubernetes.Deployment.Image != "old" {
		t.Errorf("the retired keys must still parse so validate can reject them: %q / %q / %q",
			e.Docker.Image, e.Podman.Image, e.Kubernetes.Deployment.Image)
	}
}

// TestImagePullSecretCreateDefaultsFalse pins the safe default: naming a Secret
// references one the operator manages, and only an explicit create has the tool
// build it. Defaulting the other way would overwrite a Secret it did not make.
func TestImagePullSecretCreateDefaultsFalse(t *testing.T) {
	e, err := ParseEnv([]byte("kubernetes:\n  deployment:\n    name: c\n    namespace: n\n  secrets:\n    image-pull:\n      name: regcred\n"))
	if err != nil {
		t.Fatal(err)
	}
	ip := e.Kubernetes.Secrets.ImagePull
	if ip == nil {
		t.Fatal("image-pull block should parse")
	}
	if ip.Name != "regcred" {
		t.Errorf("name = %q", ip.Name)
	}
	if ip.Create {
		t.Error("create must default to false when the key is absent")
	}

	e2, err := ParseEnv([]byte("kubernetes:\n  deployment:\n    name: c\n    namespace: n\n  secrets:\n    image-pull:\n      name: regcred\n      create: true\n"))
	if err != nil {
		t.Fatal(err)
	}
	if !e2.Kubernetes.Secrets.ImagePull.Create {
		t.Error("create: true must be honoured")
	}
}
