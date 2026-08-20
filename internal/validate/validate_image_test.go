package validate

import (
	"strings"
	"testing"

	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/spec"
)

// kubeWithImage is a fully valid kubernetes section, so a test asserting on one
// error is not also fighting the command/service ones -- checkKube requires a
// command and an in-range service port regardless of what is under test.
func kubeWithImage() *spec.Kubernetes {
	return &spec.Kubernetes{
		Deployment: baseKubeDeploy(),
		Command:    spec.DefaultKubeCommand,
		Service:    baseKubeService(),
	}
}

// TestRetiredPerPlatformImageRejected pins the loud failure for each of the
// three keys the image hoist retired. yaml drops an unknown key silently, so
// without these a stale env.yaml would deploy whatever the top-level block said
// while the operator believed the per-platform value was in force -- the exact
// drift that motivated the hoist (the shipped example carried two versions).
//
// Each message must name the replacement, not merely reject the key.
func TestRetiredPerPlatformImageRejected(t *testing.T) {
	t.Run("kubernetes.deployment.image", func(t *testing.T) {
		k := kubeWithImage()
		k.Deployment.Image = "old:1"
		errs, _ := Run(Context{Workflows: wfOK(), Defaults: &spec.Defaults{}, Image: imageOK(), Kube: k, CheckKubernetes: true})
		if !hasErr(errs, "kubernetes.deployment.image is no longer configured here") {
			t.Fatalf("want a retired-key error, got %v", errs)
		}
		if !hasErr(errs, "top-level image:") {
			t.Errorf("the error must name the replacement, got %v", errs)
		}
	})
	t.Run("docker.image", func(t *testing.T) {
		d := dockerOK()
		d.Image = "old:1"
		errs, _ := Run(Context{Workflows: wfOK(), Defaults: &spec.Defaults{}, Image: imageOK(), Docker: d, CheckDocker: true})
		if !hasErr(errs, "docker.image is no longer configured here") {
			t.Fatalf("want a retired-key error, got %v", errs)
		}
	})
	t.Run("podman.image", func(t *testing.T) {
		p := podmanOK()
		p.Image = "old:1"
		errs, _ := Run(Context{Workflows: wfOK(), Defaults: &spec.Defaults{}, Image: imageOK(), Podman: p, CheckPodman: true})
		if !hasErr(errs, "podman.image is no longer configured here") {
			t.Fatalf("want a retired-key error, got %v", errs)
		}
	})
}

// TestImageBlockRequired covers the top-level block itself. tag is required
// rather than defaulted to latest: an unpinned image is the one dependency this
// tool would otherwise leave floating.
func TestImageBlockRequired(t *testing.T) {
	tests := []struct {
		name string
		img  *spec.Image
		want string
	}{
		{"absent entirely", nil, "image: is required"},
		{"no name", &spec.Image{Tag: "1"}, "image.name is required"},
		{"no tag", &spec.Image{Name: "c"}, "image.tag is required"},
		{"unsafe name", &spec.Image{Name: "c;rm -rf /", Tag: "1"}, "image.name"},
		{"unsafe repo", &spec.Image{Repo: "reg $(evil)", Name: "c", Tag: "1"}, "image.repo"},
		{"unsafe tag", &spec.Image{Name: "c", Tag: "1 2"}, "image.tag"},
		{"unsafe repo-username", &spec.Image{Name: "c", Tag: "1", RepoUser: "a b"}, "image.repo-username"},
		{"unsafe repo-password-env", &spec.Image{Name: "c", Tag: "1", RepoPassEnv: "A B"}, "image.repo-password-env"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs, _ := Run(Context{Workflows: wfOK(), Defaults: &spec.Defaults{}, Image: tt.img, Kube: kubeWithImage(), CheckKubernetes: true})
			if !hasErr(errs, tt.want) {
				t.Fatalf("want an error containing %q, got %v", tt.want, errs)
			}
		})
	}
}

// TestImageNotRequiredWithoutAPlatform pins the gate: `generate config` renders
// application.yml alone and never pulls anything, so demanding an image there
// would reject a perfectly good config-only run.
func TestImageNotRequiredWithoutAPlatform(t *testing.T) {
	errs, _ := Run(Context{Workflows: wfOK(), Defaults: &spec.Defaults{}})
	if hasErr(errs, "image") {
		t.Errorf("no platform checked, so no image is needed: %v", errs)
	}
}

// TestImagePullSecretChecks covers the three states of the block. The middle
// one is the reason create exists: naming a Secret must not oblige the operator
// to hand the tool registry credentials it does not need.
func TestImagePullSecretChecks(t *testing.T) {
	run := func(ip *spec.ImagePullSecret, img *spec.Image, env func(string) (string, bool)) ([]Issue, []Issue) {
		k := kubeWithImage()
		k.Secrets = spec.Secrets{ImagePull: ip}
		return Run(Context{Workflows: wfOK(), Defaults: &spec.Defaults{}, Image: img, Kube: k, CheckKubernetes: true, Env: env})
	}

	t.Run("name alone needs no registry credentials", func(t *testing.T) {
		errs, _ := run(&spec.ImagePullSecret{Name: "regcred"}, imageOK(), nil)
		if len(errs) != 0 {
			t.Fatalf("referencing a Secret should not require credentials: %v", errs)
		}
	})
	t.Run("name is required", func(t *testing.T) {
		errs, _ := run(&spec.ImagePullSecret{}, imageOK(), nil)
		if !hasErr(errs, "kubernetes.secrets.image-pull.name is required") {
			t.Fatalf("want a required-name error, got %v", errs)
		}
	})
	t.Run("name must be DNS-1123", func(t *testing.T) {
		errs, _ := run(&spec.ImagePullSecret{Name: "Bad_Name"}, imageOK(), nil)
		if !hasErr(errs, "not a valid DNS-1123 label") {
			t.Fatalf("want a DNS-1123 error, got %v", errs)
		}
	})
	t.Run("create without registry credentials", func(t *testing.T) {
		errs, _ := run(&spec.ImagePullSecret{Name: "regcred", Create: true}, imageOK(), nil)
		if !hasErr(errs, "requires image.repo-username and image.repo-password-env") {
			t.Fatalf("want a missing-credentials error, got %v", errs)
		}
	})
	// wfOK() uses TLS, so an unrelated "secrets.stores is omitted" warning is
	// always present -- these look for their own warning rather than counting.
	hasWarn := func(warns []Issue, sub string) bool {
		for _, w := range warns {
			if strings.Contains(w.Msg, sub) {
				return true
			}
		}
		return false
	}
	credsImage := func() *spec.Image {
		img := imageOK()
		img.RepoUser = "svc"
		img.RepoPassEnv = "REGISTRY_PASSWORD"
		return img
	}

	t.Run("create with them, variable unset, warns", func(t *testing.T) {
		errs, warns := run(&spec.ImagePullSecret{Name: "regcred", Create: true}, credsImage(),
			func(string) (string, bool) { return "", false })
		if len(errs) != 0 {
			t.Fatalf("an unset variable is a warning, not an error: %v", errs)
		}
		if !hasWarn(warns, "REGISTRY_PASSWORD") {
			t.Fatalf("want a warning naming the variable, got %v", warns)
		}
	})
	t.Run("create with them, variable set, no warning of its own", func(t *testing.T) {
		errs, warns := run(&spec.ImagePullSecret{Name: "regcred", Create: true}, credsImage(),
			func(string) (string, bool) { return "pw", true })
		if len(errs) != 0 {
			t.Fatalf("errs = %v", errs)
		}
		if hasWarn(warns, "REGISTRY_PASSWORD") {
			t.Fatalf("a set variable must not warn, got %v", warns)
		}
	})
}

// TestRetiredPerPlatformTimezoneRejected mirrors the image retirement for the
// other value that moved to the top level. Same reasoning: yaml would drop the
// key silently, so the container would run in UTC while the operator believed
// their per-platform setting was in force.
func TestRetiredPerPlatformTimezoneRejected(t *testing.T) {
	t.Run("kubernetes.deployment.timezone", func(t *testing.T) {
		k := kubeWithImage()
		k.Deployment.Timezone = "Asia/Singapore"
		errs, _ := Run(Context{Workflows: wfOK(), Defaults: &spec.Defaults{}, Image: imageOK(), Kube: k, CheckKubernetes: true})
		if !hasErr(errs, "kubernetes.deployment.timezone is no longer configured here") {
			t.Fatalf("want a retired-key error, got %v", errs)
		}
		if !hasErr(errs, "top-level timezone:") {
			t.Errorf("the error must name the replacement, got %v", errs)
		}
	})
	t.Run("docker.timezone", func(t *testing.T) {
		d := dockerOK()
		d.Timezone = "Asia/Singapore"
		errs, _ := Run(Context{Workflows: wfOK(), Defaults: &spec.Defaults{}, Image: imageOK(), Docker: d, CheckDocker: true})
		if !hasErr(errs, "docker.timezone is no longer configured here") {
			t.Fatalf("want a retired-key error, got %v", errs)
		}
	})
	t.Run("podman.timezone", func(t *testing.T) {
		p := podmanOK()
		p.Timezone = "Asia/Singapore"
		errs, _ := Run(Context{Workflows: wfOK(), Defaults: &spec.Defaults{}, Image: imageOK(), Podman: p, CheckPodman: true})
		if !hasErr(errs, "podman.timezone is no longer configured here") {
			t.Fatalf("want a retired-key error, got %v", errs)
		}
	})
}

// TestTopLevelTimezoneUnsafe keeps the charset gate the per-platform key used to
// carry: the value is concatenated unquoted into a compose file, a podman argv
// and a manifest, so a metacharacter must be refused. It is optional, so an
// empty value is not an error -- only an unsafe one.
func TestTopLevelTimezoneUnsafe(t *testing.T) {
	for _, tz := range []string{"Asia/Singapore`id`", "UTC\nprivileged: true", "A B"} {
		errs, _ := Run(Context{Workflows: wfOK(), Defaults: &spec.Defaults{}, Image: imageOK(), Timezone: tz,
			Kube: kubeWithImage(), CheckKubernetes: true})
		if !hasErr(errs, "timezone") {
			t.Errorf("timezone %q should be rejected, got %v", tz, errs)
		}
	}
	for _, tz := range []string{"", "Asia/Singapore", "UTC"} {
		errs, _ := Run(Context{Workflows: wfOK(), Defaults: &spec.Defaults{}, Image: imageOK(), Timezone: tz,
			Kube: kubeWithImage(), CheckKubernetes: true})
		if len(errs) != 0 {
			t.Errorf("timezone %q should pass, got %v", tz, errs)
		}
	}
}
