package deploy

import (
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"

	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/consolidate"
	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/spec"
)

// renderWithPull renders the manifest set for a given pull-secret wiring.
func renderWithPull(ps *PullSecret) string {
	k := baseKube()
	return Render(Input{
		Kube: k, Defaults: &spec.Defaults{}, ImagePull: ps,
		Instance: one(k.Deployment.Name, "x: 1\n", &consolidate.Model{}),
	})
}

// TestImagePullSecretStates covers all three shapes of the block, and the
// middle one is the one that matters most: naming a Secret must render the
// reference and nothing else. Creating one anyway would overwrite the Secret
// the operator built with `kubectl create secret docker-registry`, and the
// deploy would still look like it succeeded.
func TestImagePullSecretStates(t *testing.T) {
	t.Run("no block: no reference and no Secret", func(t *testing.T) {
		out := renderWithPull(nil)
		if strings.Contains(out, "imagePullSecrets") {
			t.Errorf("imagePullSecrets should be absent:\n%s", out)
		}
		if strings.Contains(out, "dockerconfigjson") {
			t.Errorf("no dockerconfigjson Secret should be rendered:\n%s", out)
		}
	})

	t.Run("name alone: reference only", func(t *testing.T) {
		out := renderWithPull(&PullSecret{Name: "regcred"})
		if !strings.Contains(out, "      imagePullSecrets:\n        - name: regcred\n") {
			t.Errorf("imagePullSecrets entry missing:\n%s", out)
		}
		if strings.Contains(out, "dockerconfigjson") {
			t.Errorf("a reference-only block must render no Secret, or it would overwrite the operator's:\n%s", out)
		}
	})

	t.Run("created: reference and Secret", func(t *testing.T) {
		out := renderWithPull(&PullSecret{Name: "regcred", DockerConfigJSON: "PAYLOAD"})
		if !strings.Contains(out, "      imagePullSecrets:\n        - name: regcred\n") {
			t.Errorf("imagePullSecrets entry missing:\n%s", out)
		}
		if !strings.Contains(out, "kind: Secret\nmetadata:\n  name: regcred\n  namespace: ns\ntype: kubernetes.io/dockerconfigjson\ndata:\n  .dockerconfigjson: PAYLOAD\n") {
			t.Errorf("dockerconfigjson Secret missing or malformed:\n%s", out)
		}
	})
}

// TestImagePullSecretPayloadIsOpaqueToDeploy pins the package boundary: deploy
// is pure and never builds or resolves the credential, it only places the
// already-encoded payload. The value arrives from gen, which read it from the
// environment -- so nothing here can leak it anywhere but into the Secret.
func TestImagePullSecretPayloadIsOpaqueToDeploy(t *testing.T) {
	// A realistic payload, so the test also documents what gen produces.
	doc, err := json.Marshal(map[string]any{"auths": map[string]any{
		"registry.internal": map[string]string{
			"username": "svc",
			"password": "hunter2",
			"auth":     base64.StdEncoding.EncodeToString([]byte("svc:hunter2")),
		},
	}})
	if err != nil {
		t.Fatal(err)
	}
	payload := base64.StdEncoding.EncodeToString(doc)
	out := renderWithPull(&PullSecret{Name: "regcred", DockerConfigJSON: payload})

	if !strings.Contains(out, ".dockerconfigjson: "+payload+"\n") {
		t.Errorf("payload not placed verbatim:\n%s", out)
	}
	// The password may appear only inside the base64 payload, never in the clear
	// -- not in the Deployment, not as an env var, not anywhere else.
	if strings.Contains(strings.ReplaceAll(out, payload, ""), "hunter2") {
		t.Errorf("the registry password leaked outside the Secret payload:\n%s", out)
	}
}

// TestEnvBlockOmittedWhenEmpty covers the guard the timezone hoist required.
// TZ used to be unconditional because a per-platform timezone was always there
// to fill it; now that it is one optional top-level key, every entry under env:
// is optional -- and an "env:" with nothing beneath it is a null, not an empty
// list, which is not what the pod spec means.
func TestEnvBlockOmittedWhenEmpty(t *testing.T) {
	k := baseKube()
	inst := one(k.Deployment.Name, "x: 1\n", &consolidate.Model{MQTLS: false})
	inst.Timezone = ""
	out := Render(Input{Kube: k, Defaults: &spec.Defaults{}, Instance: inst})
	if strings.Contains(out, "env:") {
		t.Errorf("env: should be omitted when nothing goes under it:\n%s", out)
	}

	// Any one of the three entries brings the block back.
	inst.Timezone = "UTC"
	withTZ := Render(Input{Kube: k, Defaults: &spec.Defaults{}, Instance: inst})
	if !strings.Contains(withTZ, "          env:\n            - name: TZ\n              value: UTC\n") {
		t.Errorf("TZ entry missing:\n%s", withTZ)
	}

	inst.Timezone = ""
	inst.Model = &consolidate.Model{MQTLS: true}
	withMQTLS := Render(Input{Kube: k, Defaults: &spec.Defaults{}, Instance: inst})
	if !strings.Contains(withMQTLS, "env:") || !strings.Contains(withMQTLS, "JAVA_TOOL_OPTIONS") {
		t.Errorf("MQTLS alone should still open the env block:\n%s", withMQTLS)
	}
	if strings.Contains(withMQTLS, "- name: TZ") {
		t.Errorf("no timezone set, so no TZ entry:\n%s", withMQTLS)
	}
}
