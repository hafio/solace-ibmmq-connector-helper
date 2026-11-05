package validate

import (
	"strings"
	"testing"

	"github.com/hafio/go-deploy-utility/execguard"
	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/spec"
)

// approved builds the --allow-command approvals a test hands to Context, failing
// the test on a value Approve refuses.
func approved(t *testing.T, names ...string) execguard.Approvals {
	t.Helper()
	a, err := execguard.Approve("solmq-conn-util", names)
	if err != nil {
		t.Fatalf("Approve(%v) = %v", names, err)
	}
	return a
}

// TestCommandRulesPerPlatform pins the one Rules value per command field: the
// field the messages name, the platform as scope, and AllowedCommands as the
// source of Allowed (copied, so a caller cannot widen the package's list). An
// unknown platform has no allowlist, which the guard reports rather than
// accepting anything.
func TestCommandRulesPerPlatform(t *testing.T) {
	for platform, allowed := range AllowedCommands {
		r := CommandRules(platform)
		if r.Field != platform+".command" || r.Scope != platform || r.Subword != "" || strings.Join(r.Allowed, ",") != strings.Join(allowed, ",") {
			t.Errorf("CommandRules(%q) = %+v, want field %s.command, scope %s, allowed %v and no subword", platform, r, platform, platform, allowed)
		}
	}
	r := CommandRules(PlatformKubernetes)
	r.Allowed[0] = "curl"
	if AllowedCommands[PlatformKubernetes][0] != "kubectl" {
		t.Error("CommandRules must copy AllowedCommands, not alias it")
	}
	if err := execguard.Check(CommandRules("nomad"), []string{"nomad"}, execguard.Approvals{}); err == nil ||
		!strings.Contains(err.Error(), "Rules.Allowed is empty for nomad.command") {
		t.Errorf("an unknown platform must fail as a wiring error, got %v", err)
	}
}

// TestCommandGuardMessages pins what the shared guard says for each mistake an
// env file actually makes, named by the field it was found in. The words come
// from execguard; this pins that the platform rules reach them intact.
func TestCommandGuardMessages(t *testing.T) {
	cases := []struct {
		name     string
		platform string
		cmd      string
		want     string
	}{
		{"path", PlatformKubernetes, "/tmp/evil",
			`kubernetes.command[0] = "/tmp/evil" must be a bare binary name, not a path`},
		{"unlisted binary", PlatformKubernetes, "curl",
			`kubernetes.command[0] = "curl" is not a binary this tool runs: allowed on kubernetes are kubectl, oc -- correct the env file, or approve this one for a single run with --allow-command curl`},
		{"end of flags", PlatformKubernetes, "kubectl --",
			`kubernetes.command[1] = "--" is not allowed: end-of-flags`},
		{"bare word", PlatformKubernetes, "kubectl delete",
			`kubernetes.command[1] = "delete" is not allowed in subcommand position`},
		{"empty", PlatformKubernetes, "",
			`kubernetes.command is empty: it must name the binary to run (one of: kubectl, oc)`},
		{"unsafe character", PlatformDocker, "docker --host $(evil)",
			`docker.command[2] = "$(evil)" contains "$", which is not allowed in a command token`},
		{"sudo is an escalator", PlatformPodman, "sudo podman",
			`podman.command[0] = "sudo" is a privilege-escalation wrapper (sudo), which is never permitted in a command, not even with --allow-command`},
		{"sudo after argv0", PlatformPodman, "podman sudo",
			`podman.command[1] = "sudo" is a privilege-escalation wrapper (sudo)`},
	}
	for _, c := range cases {
		err := execguard.Check(CommandRules(c.platform), strings.Fields(c.cmd), execguard.Approvals{})
		if err == nil || !strings.Contains(err.Error(), c.want) {
			t.Errorf("%s: Check(%q) = %v, want it to contain %q", c.name, c.cmd, err, c.want)
		}
	}
}

// TestCommandGuardAccepts pins what the guard now lets through that the old gate
// refused: an embedded tilde (a Windows 8.3 short name in a flag value) and a
// ".Exe" suffix in any case, next to the plain forms.
func TestCommandGuardAccepts(t *testing.T) {
	for _, c := range []struct{ platform, cmd string }{
		{PlatformKubernetes, "kubectl"},
		{PlatformKubernetes, "oc"},
		{PlatformKubernetes, "kubectl --context prod -n ns"},
		{PlatformDocker, "docker --context foo"},
		{PlatformPodman, "podman"},
		{PlatformKubernetes, "kubectl.exe"},
		{PlatformKubernetes, "kubectl.Exe"},
		{PlatformKubernetes, "kubectl --kubeconfig=C:/RUNNER~1/config"},
		{PlatformKubernetes, "kubectl --kubeconfig C:/PROGRA~1/kube/config"},
	} {
		if err := execguard.Check(CommandRules(c.platform), strings.Fields(c.cmd), execguard.Approvals{}); err != nil {
			t.Errorf("%s %q must be accepted, got %v", c.platform, c.cmd, err)
		}
	}
	if err := execguard.Check(CommandRules(PlatformKubernetes), []string{"kubectl", "--kubeconfig", "~/kc"}, execguard.Approvals{}); err == nil ||
		!strings.Contains(err.Error(), "starts with '~'") {
		t.Errorf("a leading tilde must still be refused, got %v", err)
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
// binary is rejected when AllowCommands is the zero value (plain `validate`) and
// accepted once the operator's --allow-command approval is present. A
// privilege-escalation wrapper is the one chained binary no approval can reach:
// it cannot even be approved, and the command that names one is refused whatever
// the approvals say.
func TestContextAllowCommandsHonored(t *testing.T) {
	lima := approved(t, "lima")

	d := dockerOK()
	d.Command = "lima docker"
	if e, _ := Run(Context{Workflows: wfOK(), Defaults: &spec.Defaults{}, Docker: d, CheckDocker: true}); !hasErr(e, "docker.command[0] = \"lima\"") {
		t.Errorf("lima docker without AllowCommands should be rejected, got %v", e)
	}
	if e, _ := Run(Context{Workflows: wfOK(), Defaults: &spec.Defaults{}, Docker: d, CheckDocker: true, AllowCommands: lima}); hasErr(e, "docker.command") {
		t.Errorf("lima docker with lima approved should be accepted, got %v", e)
	}

	p := podmanOK()
	p.Command = "lima podman"
	if e, _ := Run(Context{Workflows: wfOK(), Defaults: &spec.Defaults{}, Podman: p, CheckPodman: true}); !hasErr(e, "podman.command[0] = \"lima\"") {
		t.Errorf("lima podman without AllowCommands should be rejected, got %v", e)
	}
	if e, _ := Run(Context{Workflows: wfOK(), Defaults: &spec.Defaults{}, Podman: p, CheckPodman: true, AllowCommands: lima}); hasErr(e, "podman.command") {
		t.Errorf("lima podman with lima approved should be accepted, got %v", e)
	}

	k := &spec.Kubernetes{Deployment: baseKubeDeploy(), Command: "lima kubectl"}
	if e, _ := Run(Context{Workflows: wfOK(), Defaults: &spec.Defaults{}, Kube: k, CheckKubernetes: true}); !hasErr(e, "kubernetes.command[0] = \"lima\"") {
		t.Errorf("lima kubectl without AllowCommands should be rejected, got %v", e)
	}
	if e, _ := Run(Context{Workflows: wfOK(), Defaults: &spec.Defaults{}, Kube: k, CheckKubernetes: true, AllowCommands: lima}); hasErr(e, "kubernetes.command") {
		t.Errorf("lima kubectl with lima approved should be accepted, got %v", e)
	}

	// sudo: refused with no approvals, and Approve will not produce one.
	p.Command = "sudo podman"
	if e, _ := Run(Context{Workflows: wfOK(), Defaults: &spec.Defaults{}, Podman: p, CheckPodman: true, AllowCommands: lima}); !hasErr(e, "podman.command[0] = \"sudo\" is a privilege-escalation wrapper") {
		t.Errorf("sudo podman must be refused even with other approvals, got %v", e)
	}
	if _, err := execguard.Approve("solmq-conn-util", []string{"sudo"}); err == nil || !strings.Contains(err.Error(), "sudo solmq-conn-util") {
		t.Errorf("approving sudo must be refused with the run-it-under-sudo advice, got %v", err)
	}
}

// TestCheckContainerCommandUnlistedBinaryRejected pins that docker/podman's
// command field is now held to the platform allowlist (curl, an absolute
// path), not merely the safe-charset check checkCommand used to apply.
func TestCheckContainerCommandUnlistedBinaryRejected(t *testing.T) {
	d := dockerOK()
	d.Command = "curl"
	if e, _ := Run(Context{Workflows: wfOK(), Defaults: &spec.Defaults{}, Docker: d, CheckDocker: true}); !hasErr(e, `docker.command[0] = "curl" is not a binary this tool runs: allowed on docker are docker`) {
		t.Errorf("docker.command curl should be rejected by the allowlist, got %v", e)
	}
	p := podmanOK()
	p.Command = "/tmp/evil"
	if e, _ := Run(Context{Workflows: wfOK(), Defaults: &spec.Defaults{}, Podman: p, CheckPodman: true}); !hasErr(e, "podman.command[0] = \"/tmp/evil\" must be a bare binary name, not a path") {
		t.Errorf("podman.command as a path should be rejected, got %v", e)
	}
}
