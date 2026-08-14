package validate

import (
	"fmt"
	"strings"
	"testing"

	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/spec"
)

// TestCheckDeployCommandAcceptReject pins the accept/reject matrix from the
// deploy-command lockdown: a bare, allowlisted argv[0] with flag-shaped
// arguments passes; a path, an unlisted binary, a bare positional argument, an
// unapproved chained binary, a bare end-of-flags marker, and an empty command
// are all rejected. extraAllowed (--allow-command) is what turns "sudo podman"
// from a reject into an accept.
func TestCheckDeployCommandAcceptReject(t *testing.T) {
	cases := []struct {
		name     string
		platform string
		cmd      string
		extra    []string
		wantErr  bool
	}{
		{"kubectl", PlatformKubernetes, "kubectl", nil, false},
		{"oc", PlatformKubernetes, "oc", nil, false},
		{"kubectl with context and namespace flags", PlatformKubernetes, "kubectl --context prod -n ns", nil, false},
		{"docker with context flag", PlatformDocker, "docker --context foo", nil, false},
		{"podman", PlatformPodman, "podman", nil, false},
		{"kubectl.exe (Windows-authored config)", PlatformKubernetes, "kubectl.exe", nil, false},
		{"sudo podman, sudo approved via extraAllowed", PlatformPodman, "sudo podman", []string{"sudo"}, false},

		{"curl is not on any platform allowlist", PlatformKubernetes, "curl", nil, true},
		{"absolute path argv0 rejected", PlatformKubernetes, "/tmp/evil", nil, true},
		{"relative path argv0 rejected", PlatformKubernetes, "./kubectl", nil, true},
		{"bare positional argument after argv0", PlatformKubernetes, "kubectl delete ns prod", nil, true},
		{"sudo podman rejected without extraAllowed", PlatformPodman, "sudo podman", nil, true},
		{"bare end-of-flags marker as argv0", PlatformKubernetes, "--", nil, true},
		{"empty command", PlatformKubernetes, "", nil, true},
	}
	for _, c := range cases {
		err := CheckDeployCommand(c.platform, c.cmd, c.extra)
		if (err != nil) != c.wantErr {
			t.Errorf("%s: CheckDeployCommand(%q, extra=%v) = %v, want error=%v", c.name, c.cmd, c.extra, err, c.wantErr)
		}
	}
}

// TestCheckDeployCommandEndOfFlagsMarkerMidCommand exercises the "--" rule
// specifically at a non-argv0 position (kubectl is a valid, allowlisted
// binary; "--" as its first argument is still rejected as an end-of-flags
// marker, distinct from the argv0 allowlist rejection covered above).
func TestCheckDeployCommandEndOfFlagsMarkerMidCommand(t *testing.T) {
	err := CheckDeployCommand(PlatformKubernetes, "kubectl --", nil)
	if err == nil || !strings.Contains(err.Error(), `token "--": end-of-flags marker is not accepted`) {
		t.Fatalf(`want end-of-flags rejection, got %v`, err)
	}
}

// TestCheckDeployCommandErrorTexts pins the canonical error wording verbatim
// (the CLI validator and the generator page's JS validator must emit the same
// words for the same mistake).
func TestCheckDeployCommandErrorTexts(t *testing.T) {
	if err := CheckDeployCommand(PlatformKubernetes, "/tmp/evil", nil); err == nil ||
		err.Error() != `"/tmp/evil": a path is not accepted here; use a bare binary name, resolved from PATH` {
		t.Errorf("path error text = %v", err)
	}
	wantAllowlist := fmt.Sprintf(`"curl": binary must be one of %s; deploy/delete can approve another with --allow-command <name>`,
		strings.Join(AllowedCommands[PlatformKubernetes], ", "))
	if err := CheckDeployCommand(PlatformKubernetes, "curl", nil); err == nil || err.Error() != wantAllowlist {
		t.Errorf("allowlist error text = %v, want %v", err, wantAllowlist)
	}
	if err := CheckDeployCommand(PlatformKubernetes, "kubectl --", nil); err == nil ||
		err.Error() != `token "--": end-of-flags marker is not accepted` {
		t.Errorf("end-of-flags error text = %v", err)
	}
	if err := CheckDeployCommand(PlatformKubernetes, "kubectl delete", nil); err == nil ||
		err.Error() != `token "delete": arguments must be flag-shaped (-x, --flag, --flag=value, or a flag's value); solmq-conn appends its own subcommand` {
		t.Errorf("flag-shape error text = %v", err)
	}
	if err := CheckDeployCommand(PlatformKubernetes, "", nil); err == nil || err.Error() != "must not be empty" {
		t.Errorf("empty error text = %v", err)
	}
}

// TestCheckKubeCommandNowValidated is the regression pin for the confirmed
// gap: kubernetes.command used to reach neither checkCommand nor any other
// check, so an unsafe command validated clean for k8s and only failed at
// deploy time. It must now fail validate too, same as docker/podman always
// have.
func TestCheckKubeCommandNowValidated(t *testing.T) {
	k := &spec.Kubernetes{Deployment: baseKubeDeploy(), Command: "kubectl; rm -rf /"}
	errs, _ := Run(Context{Workflows: wfOK(), Defaults: &spec.Defaults{}, Kube: k, CheckKubernetes: true})
	if !hasErr(errs, "kubernetes.command") {
		t.Errorf("want kubernetes.command validated (was previously skipped entirely), got %v", errs)
	}

	// A safe, allowlisted kubernetes.command produces no such error.
	k2 := &spec.Kubernetes{Deployment: baseKubeDeploy(), Command: "kubectl --context prod -n ns"}
	if errs, _ := Run(Context{Workflows: wfOK(), Defaults: &spec.Defaults{}, Kube: k2, CheckKubernetes: true}); hasErr(errs, "kubernetes.command") {
		t.Errorf("valid kubernetes.command should not be rejected, got %v", errs)
	}
}

// TestCheckKubeCommandDefaultKubectlUnvalidated pins that the zero-value
// default (spec.Kubernetes.Default fills Command with DefaultKubeCommand
// before validate ever runs) validates clean.
func TestCheckKubeCommandDefaultKubectlUnvalidated(t *testing.T) {
	k := &spec.Kubernetes{Deployment: baseKubeDeploy(), Command: spec.DefaultKubeCommand}
	if errs, _ := Run(Context{Workflows: wfOK(), Defaults: &spec.Defaults{}, Kube: k, CheckKubernetes: true}); hasErr(errs, "kubernetes.command") {
		t.Errorf("default kubernetes.command should validate clean, got %v", errs)
	}
}

// TestContextAllowCommandsHonored covers Context.AllowCommands threading into
// checkContainerTarget (docker/podman) and checkKube (kubernetes): a chained
// binary is rejected when AllowCommands is nil (plain `validate`) and accepted
// once the operator's --allow-command value is present.
func TestContextAllowCommandsHonored(t *testing.T) {
	d := dockerOK()
	d.Command = "sudo docker"
	if e, _ := Run(Context{Workflows: wfOK(), Defaults: &spec.Defaults{}, Docker: d, CheckDocker: true}); !hasErr(e, "docker.command") {
		t.Errorf("sudo docker without AllowCommands should be rejected, got %v", e)
	}
	if e, _ := Run(Context{Workflows: wfOK(), Defaults: &spec.Defaults{}, Docker: d, CheckDocker: true, AllowCommands: []string{"sudo"}}); hasErr(e, "docker.command") {
		t.Errorf("sudo docker with AllowCommands=[sudo] should be accepted, got %v", e)
	}

	p := podmanOK()
	p.Command = "sudo podman"
	if e, _ := Run(Context{Workflows: wfOK(), Defaults: &spec.Defaults{}, Podman: p, CheckPodman: true}); !hasErr(e, "podman.command") {
		t.Errorf("sudo podman without AllowCommands should be rejected, got %v", e)
	}
	if e, _ := Run(Context{Workflows: wfOK(), Defaults: &spec.Defaults{}, Podman: p, CheckPodman: true, AllowCommands: []string{"sudo"}}); hasErr(e, "podman.command") {
		t.Errorf("sudo podman with AllowCommands=[sudo] should be accepted, got %v", e)
	}

	k := &spec.Kubernetes{Deployment: baseKubeDeploy(), Command: "sudo kubectl"}
	if e, _ := Run(Context{Workflows: wfOK(), Defaults: &spec.Defaults{}, Kube: k, CheckKubernetes: true}); !hasErr(e, "kubernetes.command") {
		t.Errorf("sudo kubectl without AllowCommands should be rejected, got %v", e)
	}
	if e, _ := Run(Context{Workflows: wfOK(), Defaults: &spec.Defaults{}, Kube: k, CheckKubernetes: true, AllowCommands: []string{"sudo"}}); hasErr(e, "kubernetes.command") {
		t.Errorf("sudo kubectl with AllowCommands=[sudo] should be accepted, got %v", e)
	}
}

// TestCheckContainerCommandUnlistedBinaryRejected pins that docker/podman's
// command field is now held to the platform allowlist (curl, an absolute
// path), not merely the safe-charset check checkCommand used to apply.
func TestCheckContainerCommandUnlistedBinaryRejected(t *testing.T) {
	d := dockerOK()
	d.Command = "curl"
	if e, _ := Run(Context{Workflows: wfOK(), Defaults: &spec.Defaults{}, Docker: d, CheckDocker: true}); !hasErr(e, "binary must be one of") {
		t.Errorf("docker.command curl should be rejected by the allowlist, got %v", e)
	}
	p := podmanOK()
	p.Command = "/tmp/evil"
	if e, _ := Run(Context{Workflows: wfOK(), Defaults: &spec.Defaults{}, Podman: p, CheckPodman: true}); !hasErr(e, "a path is not accepted here") {
		t.Errorf("podman.command as a path should be rejected, got %v", e)
	}
}
