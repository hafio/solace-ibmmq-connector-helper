package gen

import (
	"strings"
	"testing"
)

// renderAll renders the kubernetes manifests, the compose file and the quadlet
// unit for one connector with the given run-as line (empty: none written)
// under every platform section, and fails the test on any error.
func renderAll(t *testing.T, runAs string) map[string]string {
	t.Helper()
	line := ""
	if runAs != "" {
		line = "  run-as: \"" + runAs + "\"\n"
	}
	const base = "image:\n  name: img\n  tag: v1\n"
	kube := "kubernetes:\n  command: kubectl\n  deployment:\n    name: solmq\n    namespace: ns\n" + line +
		"  libs:\n    download:\n      urls:\n        - https://repo1.maven.org/maven2/a/b/1/b-1.jar\n      pvc: libs\n"
	docker := "docker:\n  command: docker\n  name: solmq\n" + line
	podman := "podman:\n  command: podman\n  name: solmq\n" + line
	res := Resolver{Rand: fixedStatusRand, Env: func(string) (string, bool) { return "v", true }}
	req := func(env string) Request {
		return Request{Env: &File{Name: "env.yaml", Data: []byte(env)}, Workflows: synthWorkflowFiles(1)}
	}
	k8s, errs, _ := GenerateKubernetes(req(base+kube), res, KubeOpts{})
	if len(errs) > 0 {
		t.Fatalf("kubernetes run-as %q: %v", runAs, errs)
	}
	compose, errs, _ := GenerateDocker(req(base+docker), res)
	if len(errs) > 0 {
		t.Fatalf("docker run-as %q: %v", runAs, errs)
	}
	quadlet, errs, _ := GeneratePodman(req(base+podman), res)
	if len(errs) > 0 {
		t.Fatalf("podman run-as %q: %v", runAs, errs)
	}
	return map[string]string{"kubernetes": k8s, "docker": compose.Compose, "podman": quadlet.Unit.Content}
}

// TestArtifactsStateTheirPosture pins what every artifact says about the
// container's identity and privileges, for the default (no run-as written), an
// explicit id, uid 0 and image. The posture -- no privilege escalation or
// new privileges, every capability dropped, not privileged, the runtime's
// seccomp profile -- holds whatever the run-as; the ids follow it, image
// writes none, and only a non-root uid asks the kubelet for runAsNonRoot. The
// kubernetes libs-download init container carries the same container posture.
func TestArtifactsStateTheirPosture(t *testing.T) {
	posture := map[string][]string{
		"kubernetes": {
			"seccompProfile:\n          type: RuntimeDefault\n",
			"          securityContext:\n            allowPrivilegeEscalation: false\n            privileged: false\n            capabilities:\n              drop:\n                - ALL\n",
		},
		"docker": {"    privileged: false\n    cap_drop:\n      - ALL\n    security_opt:\n      - no-new-privileges=true\n"},
		"podman": {"NoNewPrivileges=true\nDropCapability=all\n"},
	}
	for _, c := range []struct {
		runAs                 string
		kube, docker, podman  []string
		noKube, noDock, noPod []string
	}{
		{"", []string{"runAsUser: 1000\n", "runAsGroup: 0\n", "fsGroup: 0\n", "runAsNonRoot: true\n"}, []string{`user: "1000:0"`}, []string{"User=1000\nGroup=0\n"}, nil, nil, nil},
		{"1234:5678", []string{"runAsUser: 1234\n", "runAsGroup: 5678\n", "fsGroup: 5678\n", "runAsNonRoot: true\n"}, []string{`user: "1234:5678"`}, []string{"User=1234\nGroup=5678\n"}, nil, nil, nil},
		{"0", []string{"runAsUser: 0\n", "runAsGroup: 0\n", "fsGroup: 0\n"}, []string{`user: "0:0"`}, []string{"User=0\nGroup=0\n"}, []string{"runAsNonRoot"}, nil, nil},
		// The compose service's own user: key sits at four spaces; the inlined
		// application.yml carries user: keys of its own, deeper in.
		{"image", nil, nil, nil, []string{"runAsUser", "runAsGroup", "fsGroup", "runAsNonRoot"}, []string{"\n    user: "}, []string{"User=", "Group="}},
	} {
		t.Run("run-as "+c.runAs, func(t *testing.T) {
			out := renderAll(t, c.runAs)
			for platform, want := range map[string][]string{"kubernetes": c.kube, "docker": c.docker, "podman": c.podman} {
				for _, w := range append(append([]string(nil), posture[platform]...), want...) {
					if !strings.Contains(out[platform], w) {
						t.Errorf("%s artifact missing %q:\n%s", platform, w, out[platform])
					}
				}
			}
			for platform, no := range map[string][]string{"kubernetes": c.noKube, "docker": c.noDock, "podman": c.noPod} {
				for _, n := range no {
					if strings.Contains(out[platform], n) {
						t.Errorf("%s artifact must not carry %q for run-as %q:\n%s", platform, n, c.runAs, out[platform])
					}
				}
			}
			// The init container gets the container posture too: two
			// container securityContexts, one per container.
			if n := strings.Count(out["kubernetes"], "allowPrivilegeEscalation: false"); n != 2 {
				t.Errorf("want the connector and the libs-download init container hardened, got %d container securityContexts", n)
			}
		})
	}
}

// TestArtifactsCarryNoWideningTokens forbids, in every artifact and for every
// run-as, the spellings that would widen what the container may do: running
// privileged, escalating privileges, adding capabilities, dropping the
// seccomp or SELinux confinement, or sharing the host's namespaces. None is
// emitted today; this keeps it that way.
func TestArtifactsCarryNoWideningTokens(t *testing.T) {
	widening := []string{
		"privileged: true", "--privileged", "Privileged=true",
		"allowPrivilegeEscalation: true", "no-new-privileges:false", "no-new-privileges=false", "NoNewPrivileges=false",
		"cap_add", "AddCapability", "SYS_ADMIN",
		"Unconfined", "unconfined", "SecurityLabelDisable", "label=disable", "label:disable",
		"hostNetwork", "hostPID", "hostIPC", "network_mode: host", "Network=host", "pid: host", "ipc: host",
	}
	for _, runAs := range []string{"", "1234:5678", "0", "image"} {
		for platform, out := range renderAll(t, runAs) {
			for _, w := range widening {
				if strings.Contains(out, w) {
					t.Errorf("%s artifact for run-as %q carries the widening token %q:\n%s", platform, runAs, w, out)
				}
			}
		}
	}
}

// TestRunAsOfRefusesWhatValidateRefuses covers the renderers' own parse of a
// run-as. validate refuses a bad value first, so this only fires for a caller
// that rendered without validating -- and then it must be an error, never the
// zero identity, which is root.
func TestRunAsOfRefusesWhatValidateRefuses(t *testing.T) {
	if id, errs := runAsOf("docker", "1000:0"); errs != nil || id.UID != 1000 || id.GID != 0 {
		t.Errorf("runAsOf(1000:0) = %+v, %v", id, errs)
	}
	id, errs := runAsOf("podman", "root")
	if len(errs) != 1 || !strings.Contains(errs[0].Msg, `podman.run-as "root": the uid "root" is not a decimal number`) {
		t.Errorf("runAsOf(root) errs = %v", errs)
	}
	if id.UID != 0 || id.Image {
		t.Errorf("a refused run-as returns the zero id alongside its error, got %+v", id)
	}
}
