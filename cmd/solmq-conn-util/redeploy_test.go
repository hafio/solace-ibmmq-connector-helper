package main

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	gdurunner "github.com/hafio/go-deploy-utility/runner"
	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/dockergen"
	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/spec"
)

// The redeploy prompt: a deploy that would restart or recreate a running docker or
// podman service whose artifact or secrets changed asks first, and a "no" leaves
// the host exactly as it was. These tests pin the order -- compare, ask, write --
// and the answers: --no-prompt approves, no terminal fails naming the flag, and an
// unchanged deploy asks nothing.
//
// The reserved status account's password is inlined in the artifact. It is derived
// from the deployment's credentials and name (gen.resolveStatusPassword), so these
// specs render the same artifact on every deploy with no pin: an unchanged redeploy
// really is unchanged.

// ---- docker -----------------------------------------------------------------------

const (
	dockerRunningRow = "solmq-conn\timg:1\tUp 2 hours\n"
	dockerOtherRow   = "another-container\timg:1\tUp 2 hours\n"
)

// dockerSpec writes a docker env.yaml and returns its path.
func dockerSpec(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	write(t, dir, "env.yaml", sharedEnv+dockerEnv)
	write(t, dir, "10.yaml", validWF)
	return filepath.Join(dir, "env.yaml")
}

// dockerDigestOf deploys the spec once, with nothing running, and returns the
// deploy digest it handed compose: what an unchanged redeploy would compute again.
func dockerDigestOf(t *testing.T, envPath string) string {
	t.Helper()
	f := useFakeRunner(t)
	if code := dispatch([]string{"deploy", "--platform", "docker", "-e", envPath}, testApp(f)); code != 0 {
		t.Fatalf("first deploy exit=%d", code)
	}
	digest, ok := envValue(f.calls[2].env, dockergen.DeployDigestVar)
	if !ok || len(digest) != 64 {
		t.Fatalf("deploy digest = %q, want 64 hex characters", digest)
	}
	return digest
}

// dockerInspectWithLabel is `docker container inspect` for the running container,
// carrying label as its deploy digest ("" for a container with no such label).
func dockerInspectWithLabel(label string) string {
	labels := `{}`
	if label != "" {
		labels = `{"` + spec.LabelDeployDigestKey + `":"` + label + `"}`
	}
	return `[{"Name":"/solmq-conn","State":{"Status":"running"},"Config":{"Image":"img:1","Labels":` + labels + `}}]`
}

// deployDocker runs a docker deploy against the queued answers: preflight, the
// listing, then (for a running container) the inspect. It returns the exit code,
// stderr and what the prompter was asked.
func deployDocker(t *testing.T, envPath string, q *queueRunner, terminal bool, input string, extra ...string) (code int, stderr, asked string) {
	t.Helper()
	a := testApp(nil)
	out := withPrompter(&a, terminal, input)
	args := append([]string{"deploy", "--platform", "docker", "-e", envPath}, extra...)
	stderr = captureStderr(t, func() { code = dispatch(args, a.withRunner(q)) })
	return code, stderr, out.String()
}

// verbs is the docker subcommand of each recorded call, for asserting the order.
func verbs(calls []fakeCall) []string {
	var out []string
	for _, c := range calls {
		switch {
		case len(c.argv) > 2 && c.argv[1] == "container":
			out = append(out, c.argv[2])
		case len(c.argv) > 1:
			out = append(out, c.argv[1])
		}
	}
	return out
}

// A running container whose label already holds this deploy's digest asks nothing:
// compose is told to bring it up, finds its configuration unchanged and leaves it
// running. Even with no terminal at all, since nothing is asked.
func TestDeployDockerUnchangedRunningAsksNothingAndRunsUp(t *testing.T) {
	envPath := dockerSpec(t)
	digest := dockerDigestOf(t, envPath)
	q := &queueRunner{resp: []queuedResp{{"", nil}, {dockerRunningRow, nil}, {dockerInspectWithLabel(digest), nil}}}
	code, stderr, asked := deployDocker(t, envPath, q, false, "")
	if code != 0 || asked != "" {
		t.Fatalf("exit=%d, asked %q; want 0 and no question:\n%s", code, asked, stderr)
	}
	if got, want := verbs(q.calls), []string{"info", "ps", "inspect", "compose"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("calls = %v, want %v", got, want)
	}
	up := q.calls[3].argv
	if up[len(up)-2] != "up" || up[len(up)-1] != "-d" || containsToken(up, "--force-recreate") {
		t.Errorf("up argv = %v, want a plain `up -d`: compose decides", up)
	}
	if got, _ := envValue(q.calls[3].env, dockergen.DeployDigestVar); got != digest {
		t.Errorf("the child env carries digest %q, want the one it compared, %q", got, digest)
	}
}

// A container that is not running is never inspected and never asked about, so a
// deploy over a stopped or absent one is as quiet as a first deploy.
func TestDeployDockerNotRunningAsksNothing(t *testing.T) {
	envPath := dockerSpec(t)
	for name, list := range map[string]string{"absent": "", "another container": dockerOtherRow, "a sibling": "solmq-conn-2\timg:1\tUp 1 hour\n"} {
		t.Run(name, func(t *testing.T) {
			q := &queueRunner{resp: []queuedResp{{"", nil}, {list, nil}}}
			code, stderr, asked := deployDocker(t, envPath, q, false, "")
			if code != 0 || asked != "" {
				t.Fatalf("exit=%d, asked %q:\n%s", code, asked, stderr)
			}
			if got, want := verbs(q.calls), []string{"info", "ps", "compose"}; !reflect.DeepEqual(got, want) {
				t.Errorf("calls = %v, want %v", got, want)
			}
		})
	}
}

// A different label, a missing one (a container an earlier build created) and an
// inspect that fails each count as changed: compose would recreate the container,
// so the deploy asks first, naming it, and "y" lets compose run.
func TestDeployDockerChangedAsksBeforeCompose(t *testing.T) {
	envPath := dockerSpec(t)
	for _, c := range []struct {
		name    string
		inspect queuedResp
	}{
		{"a different digest", queuedResp{dockerInspectWithLabel("0000"), nil}},
		{"no label", queuedResp{dockerInspectWithLabel(""), nil}},
		{"the inspect fails", queuedResp{"no such container", errors.New("exit status 1")}},
		{"the inspect is not json", queuedResp{"not json", nil}},
	} {
		t.Run(c.name, func(t *testing.T) {
			q := &queueRunner{resp: []queuedResp{{"", nil}, {dockerRunningRow, nil}, c.inspect}}
			code, stderr, asked := deployDocker(t, envPath, q, true, "y\n")
			if code != 0 {
				t.Fatalf("exit=%d:\n%s", code, stderr)
			}
			if !strings.Contains(asked, "container solmq-conn is running") || !strings.Contains(asked, "compose will recreate it") ||
				!strings.HasSuffix(asked, "[y/N] ") {
				t.Errorf("the question should name the container and what compose will do, got %q", asked)
			}
			if got, want := verbs(q.calls), []string{"info", "ps", "inspect", "compose"}; !reflect.DeepEqual(got, want) {
				t.Errorf("calls = %v, want %v", got, want)
			}
		})
	}
}

// --no-prompt approves without asking, so a script or CI job needs no terminal.
func TestDeployDockerNoPromptApproves(t *testing.T) {
	envPath := dockerSpec(t)
	q := &queueRunner{resp: []queuedResp{{"", nil}, {dockerRunningRow, nil}, {dockerInspectWithLabel("0000"), nil}}}
	code, stderr, asked := deployDocker(t, envPath, q, false, "", noPromptFlagName)
	if code != 0 || asked != "" {
		t.Fatalf("exit=%d, asked %q:\n%s", code, asked, stderr)
	}
	if got, want := verbs(q.calls), []string{"info", "ps", "inspect", "compose"}; !reflect.DeepEqual(got, want) {
		t.Errorf("calls = %v, want %v", got, want)
	}
}

// With no terminal and no --no-prompt the question cannot be asked: exit 1 naming
// the flag, before compose runs.
func TestDeployDockerNonTTYFailsNamingTheFlag(t *testing.T) {
	envPath := dockerSpec(t)
	q := &queueRunner{resp: []queuedResp{{"", nil}, {dockerRunningRow, nil}, {dockerInspectWithLabel("0000"), nil}}}
	code, stderr, asked := deployDocker(t, envPath, q, false, "")
	if code != 1 || asked != "" {
		t.Fatalf("exit=%d, asked %q, want 1 and nothing written to the prompt:\n%s", code, asked, stderr)
	}
	if !strings.Contains(stderr, noPromptFlagName) || !strings.Contains(stderr, "the redeploy") {
		t.Errorf("the refusal must name %s and what it is for, got:\n%s", noPromptFlagName, stderr)
	}
	if got, want := verbs(q.calls), []string{"info", "ps", "inspect"}; !reflect.DeepEqual(got, want) {
		t.Errorf("calls = %v, want %v: compose must not run", got, want)
	}
}

// Anything but y/yes declines: nothing runs after the probes, and exit 0.
func TestDeployDockerDeclinedRunsNothing(t *testing.T) {
	envPath := dockerSpec(t)
	for _, answer := range []string{"n\n", "\n", "maybe\n"} {
		q := &queueRunner{resp: []queuedResp{{"", nil}, {dockerRunningRow, nil}, {dockerInspectWithLabel("0000"), nil}}}
		code, stderr, _ := deployDocker(t, envPath, q, true, answer)
		if code != 0 || !strings.Contains(stderr, "deploy: cancelled; nothing changed") {
			t.Errorf("answer %q: exit=%d, want 0 and the cancelled message:\n%s", answer, code, stderr)
		}
		if got, want := verbs(q.calls), []string{"info", "ps", "inspect"}; !reflect.DeepEqual(got, want) {
			t.Errorf("answer %q: calls = %v, want %v: compose must not run", answer, got, want)
		}
	}
}

// A listing that cannot be read is an error, exit 1, naming the container: with no
// answer the deploy cannot say whether it would restart a running service, and it
// must not guess.
func TestDeployDockerFailedRunningProbeIsAnError(t *testing.T) {
	envPath := dockerSpec(t)
	q := &queueRunner{resp: []queuedResp{{"", nil}, {"Cannot connect to the Docker daemon", errors.New("exit status 1")}}}
	code, stderr, asked := deployDocker(t, envPath, q, true, "y\n")
	if code != 1 || asked != "" {
		t.Fatalf("exit=%d, asked %q, want 1 and no question:\n%s", code, asked, stderr)
	}
	for _, w := range []string{"solmq-conn", "Cannot connect to the Docker daemon"} {
		if !strings.Contains(stderr, w) {
			t.Errorf("stderr should carry %q, got:\n%s", w, stderr)
		}
	}
	if got, want := verbs(q.calls), []string{"info", "ps"}; !reflect.DeepEqual(got, want) {
		t.Errorf("calls = %v, want %v", got, want)
	}
}

// ---- kubernetes -------------------------------------------------------------------

// kubernetes asks nothing -- kubectl apply and the rollout decide -- so a deploy
// never writes a prompt, and --no-prompt there is a flag that would do nothing: a
// usage error (exit 2) saying so, with nothing run.
func TestDeployKubernetesNeverPromptsAndRefusesNoPrompt(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "env.yaml", sharedEnv+kubeEnv)
	write(t, dir, "10.yaml", validWF)
	envPath := filepath.Join(dir, "env.yaml")

	f := useFakeRunner(t)
	a := testApp(nil)
	out := withPrompter(&a, true, "n\n")
	if code := dispatch([]string{"deploy", "--platform", "kubernetes", "-e", envPath}, a.withRunner(f)); code != 0 {
		t.Fatalf("deploy exit=%d, want 0", code)
	}
	if out.Len() != 0 {
		t.Errorf("a kubernetes deploy must never prompt, got %q", out.String())
	}

	f2 := useFakeRunner(t)
	var code int
	stderr := captureStderr(t, func() {
		code = dispatch([]string{"deploy", "--platform", "kubernetes", noPromptFlagName, "-e", envPath}, a.withRunner(f2))
	})
	if code != 2 || len(f2.calls) != 0 {
		t.Fatalf("deploy %s: exit=%d with %d calls, want 2 and nothing run", noPromptFlagName, code, len(f2.calls))
	}
	if !strings.Contains(stderr, noPromptFlagName) || !strings.Contains(stderr, "kubernetes deploy, which asks nothing") {
		t.Errorf("the refusal should name the flag and say kubernetes asks nothing, got:\n%s", stderr)
	}
}

// ---- podman -----------------------------------------------------------------------

// podmanDeployRunner answers what a podman redeploy asks, by what each command is:
// the client version, the value stored under each secret (stored, by secret name; a
// name not in the map is a missing secret, which fails) and the unit's state. Every
// other command succeeds with no output.
type podmanDeployRunner struct {
	noAttach
	calls   []fakeCall
	stored  map[string]string
	state   string
	version string // "" answers 5.8.2
}

func (p *podmanDeployRunner) Run(_ context.Context, c gdurunner.Cmd, stdout, _ io.Writer) error {
	p.calls = append(p.calls, fakeCall{argv: c.Argv, stdin: string(c.Stdin), env: c.Env})
	switch {
	case containsToken(c.Argv, "version"):
		v := p.version
		if v == "" {
			v = "5.8.2"
		}
		scriptOut(stdout, v+"\n")
	case containsToken(c.Argv, "secret") && containsToken(c.Argv, "inspect"):
		v, ok := p.stored[c.Argv[len(c.Argv)-1]]
		if !ok {
			return errors.New("no such secret")
		}
		scriptOut(stdout, v+"\n")
	case containsToken(c.Argv, "is-active"):
		scriptOut(stdout, p.state+"\n")
		if p.state != "active" && p.state != "reloading" {
			return errors.New("exit status 3")
		}
	}
	return nil
}

// did reports whether any call began with prefix.
func (p *podmanDeployRunner) did(prefix ...string) bool {
	for _, c := range p.calls {
		if len(c.argv) >= len(prefix) && reflect.DeepEqual(c.argv[:len(prefix)], prefix) {
			return true
		}
	}
	return false
}

// creates are the secrets a deploy stored: name -> the value on the create's stdin.
func creates(calls []fakeCall) map[string]string {
	out := map[string]string{}
	for _, c := range calls {
		if len(c.argv) > 3 && c.argv[1] == "secret" && c.argv[2] == "create" {
			out[c.argv[len(c.argv)-2]] = c.stdin
		}
	}
	return out
}

// podmanRedeploy is a podman deploy that has already happened once: the unit is on
// disk, and podman holds every secret's value as the first deploy stored it.
type podmanRedeploy struct {
	envPath, unitPath string
	stored            map[string]string
	// host is the ordinary user whose quadlet directory holds the unit, which
	// every later run has to dispatch with to find it.
	host hostFacts
}

func deployedPodman(t *testing.T) podmanRedeploy {
	t.Helper()
	a := testApp(nil)
	quadletDir := podmanQuadletHome(t, &a)
	dir := t.TempDir()
	write(t, dir, "env.yaml", sharedEnv+podmanEnv())
	write(t, dir, "10.yaml", validWF)
	r := podmanRedeploy{envPath: filepath.Join(dir, "env.yaml"), unitPath: filepath.Join(quadletDir, "solmq-conn.container"), host: a.host}
	f := useFakeRunner(t)
	if code := dispatch([]string{"deploy", "--platform", "podman", "-e", r.envPath}, a.withRunner(f)); code != 0 {
		t.Fatalf("first deploy exit=%d", code)
	}
	r.stored = creates(f.calls)
	if len(r.stored) != 6 {
		t.Fatalf("the first deploy stored %d secrets, want 6 (four credentials, application.yml, the status script): %d names", len(r.stored), len(r.stored))
	}
	return r
}

func (r podmanRedeploy) run(t *testing.T, p *podmanDeployRunner, terminal bool, input string, extra ...string) (code int, stderr, asked string) {
	t.Helper()
	a := testApp(nil)
	useHost(&a, r.host)
	out := withPrompter(&a, terminal, input)
	args := append([]string{"deploy", "--platform", "podman", "-e", r.envPath}, extra...)
	stderr = captureStderr(t, func() { code = dispatch(args, a.withRunner(p)) })
	return code, stderr, out.String()
}

func (r podmanRedeploy) unit(t *testing.T) string {
	t.Helper()
	b, err := os.ReadFile(r.unitPath)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// An unchanged deploy of a running unit writes nothing and restarts nothing: the
// unit matches the file on disk, every stored secret reads back as this value, so
// there is nothing to ask and nothing to do -- and it says how to restart.
func TestDeployPodmanUnchangedRunningIsLeftAlone(t *testing.T) {
	r := deployedPodman(t)
	before := r.unit(t)
	p := &podmanDeployRunner{stored: r.stored, state: "active"}
	code, stderr, asked := r.run(t, p, false, "")
	if code != 0 || asked != "" {
		t.Fatalf("exit=%d, asked %q; want 0 and no question (no terminal is needed):\n%s", code, asked, stderr)
	}
	for _, prefix := range [][]string{{"podman", "secret", "create"}, {"podman", "secret", "rm"},
		{"systemctl", "--user", "daemon-reload"}, {"systemctl", "--user", "start"}, {"systemctl", "--user", "restart"}} {
		if p.did(prefix...) {
			t.Errorf("an unchanged deploy ran %v: %+v", prefix, p.calls)
		}
	}
	if !strings.Contains(stderr, "solmq-conn.service is unchanged and left running") || !strings.Contains(stderr, "solmq-conn-util restart") {
		t.Errorf("stderr should say it is unchanged and how to restart it, got:\n%s", stderr)
	}
	if r.unit(t) != before {
		t.Error("the unit must not be rewritten")
	}
	if !p.did("systemctl", "--user", "is-active", "solmq-conn.service") {
		t.Errorf("the unit's state is always read, got %+v", p.calls)
	}
}

// An unchanged deploy of a unit that is not running starts it: nothing is written
// and nothing is reloaded, but a stopped service is brought up.
func TestDeployPodmanUnchangedStoppedIsStarted(t *testing.T) {
	r := deployedPodman(t)
	p := &podmanDeployRunner{stored: r.stored, state: "inactive"}
	code, stderr, asked := r.run(t, p, false, "")
	if code != 0 || asked != "" {
		t.Fatalf("exit=%d, asked %q:\n%s", code, asked, stderr)
	}
	if !p.did("systemctl", "--user", "start", "solmq-conn.service") {
		t.Errorf("a stopped unit is started, got %+v", p.calls)
	}
	for _, prefix := range [][]string{{"podman", "secret", "create"}, {"systemctl", "--user", "daemon-reload"}, {"systemctl", "--user", "restart"}} {
		if p.did(prefix...) {
			t.Errorf("nothing changed, yet the deploy ran %v", prefix)
		}
	}
}

// One changed secret asks, names the secret (never a value), writes only
// that secret -- the unchanged ones are skipped -- leaves the unchanged unit as it is,
// reloads and restarts. A secret with no label, or one that does not exist, is a
// change too.
func TestDeployPodmanChangedSecretAsksThenWritesOnlyIt(t *testing.T) {
	for _, c := range []struct {
		name  string
		stale func(stored map[string]string) map[string]string
	}{
		{"a different value", func(l map[string]string) map[string]string { l["solmq-conn-application.yml"] = "old"; return l }},
		{"no such secret", func(l map[string]string) map[string]string { delete(l, "solmq-conn-application.yml"); return l }},
	} {
		t.Run(c.name, func(t *testing.T) {
			r := deployedPodman(t)
			before := r.unit(t)
			stored := map[string]string{}
			for k, v := range r.stored {
				stored[k] = v
			}
			p := &podmanDeployRunner{stored: c.stale(stored), state: "active"}
			code, stderr, asked := r.run(t, p, true, "y\n")
			if code != 0 {
				t.Fatalf("exit=%d:\n%s", code, stderr)
			}
			if !strings.Contains(asked, "solmq-conn.service is running") || !strings.Contains(asked, "the secrets solmq-conn-application.yml") ||
				strings.Contains(asked, "its unit") {
				t.Errorf("the question should name the service and only the changed secret, got %q", asked)
			}
			if strings.Contains(asked+stderr, r.stored["solmq-conn-application.yml"]) {
				t.Errorf("a secret's value must never be printed, got %q / %q", asked, stderr)
			}
			got := creates(p.calls)
			if len(got) != 1 || got["solmq-conn-application.yml"] == "" {
				t.Errorf("only application.yml is stored, got %v", got)
			}
			if got["solmq-conn-application.yml"] != r.stored["solmq-conn-application.yml"] {
				t.Error("the stored value is the one the first deploy rendered")
			}
			for _, c := range p.calls {
				if len(c.argv) > 3 && c.argv[2] == "create" && containsToken(c.argv, got["solmq-conn-application.yml"]) {
					t.Error("a value must never reach argv")
				}
			}
			if !p.did("systemctl", "--user", "daemon-reload") || !p.did("systemctl", "--user", "restart", "solmq-conn.service") || p.did("systemctl", "--user", "start") {
				t.Errorf("a changed, running unit is reloaded and restarted, got %+v", p.calls)
			}
			if r.unit(t) != before {
				t.Error("the unit did not change and must not be rewritten")
			}
		})
	}
}

// Reading a value back needs podman 4.7: below it, or with a version that cannot be
// read, every secret counts as changed and none is probed. A matching value on 4.7
// and newer is skipped, and a probe that fails (above, "no such secret") is a change.
func TestDeployPodmanOlderPodmanCountsEverySecretChangedAndProbesNone(t *testing.T) {
	for _, v := range []string{"4.6.2", "not a version"} {
		r := deployedPodman(t)
		p := &podmanDeployRunner{stored: r.stored, state: "inactive", version: v}
		code, stderr, asked := r.run(t, p, false, "")
		if code != 0 || asked != "" {
			t.Fatalf("version %q: exit=%d, asked %q:\n%s", v, code, asked, stderr)
		}
		for _, c := range p.calls {
			if containsToken(c.argv, "--showsecret") {
				t.Errorf("version %q: no secret may be read back, got %v", v, c.argv)
			}
		}
		if len(creates(p.calls)) != 6 {
			t.Errorf("version %q: every secret is stored again, got %d", v, len(creates(p.calls)))
		}
		for _, c := range p.calls {
			if len(c.argv) > 3 && c.argv[2] == "create" && containsToken(c.argv, "--replace") {
				t.Errorf("version %q: no --replace below 4.7: %v", v, c.argv)
			}
		}
	}
}

// On 4.7 and newer the read-back is per secret, by the exact value, and the version
// is read before any comparison.
func TestDeployPodmanReadsTheVersionThenEachSecretBack(t *testing.T) {
	r := deployedPodman(t)
	p := &podmanDeployRunner{stored: r.stored, state: "active"}
	if code, stderr, _ := r.run(t, p, false, ""); code != 0 {
		t.Fatalf("exit=%d:\n%s", code, stderr)
	}
	var seq []string
	for _, c := range p.calls {
		switch {
		case containsToken(c.argv, "version"):
			seq = append(seq, "version")
		case containsToken(c.argv, "--showsecret"):
			seq = append(seq, "read")
			if want := "{{.SecretData}}"; !containsToken(c.argv, want) {
				t.Errorf("read-back argv %v lacks %s", c.argv, want)
			}
		}
	}
	if want := []string{"version", "read", "read", "read", "read", "read", "read"}; !reflect.DeepEqual(seq, want) {
		t.Errorf("sequence = %v, want %v", seq, want)
	}
}

// A changed unit alone asks too, and is the one thing written: no secret is touched.
func TestDeployPodmanChangedUnitAsksAndWritesOnlyIt(t *testing.T) {
	r := deployedPodman(t)
	if err := os.WriteFile(r.unitPath, []byte("[Container]\nImage=old\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	p := &podmanDeployRunner{stored: r.stored, state: "reloading"}
	code, stderr, asked := r.run(t, p, true, "yes\n")
	if code != 0 || !strings.Contains(asked, "its unit changed") {
		t.Fatalf("exit=%d, asked %q:\n%s", code, asked, stderr)
	}
	if p.did("podman", "secret", "create") {
		t.Errorf("no secret changed, yet one was stored: %+v", p.calls)
	}
	if got := r.unit(t); strings.Contains(got, "Image=old") {
		t.Errorf("the changed unit is replaced, got:\n%s", got)
	}
	if !p.did("systemctl", "--user", "restart", "solmq-conn.service") {
		t.Errorf("a reloading unit counts as running and is restarted, got %+v", p.calls)
	}
}

// A "no" leaves the host exactly as it was: no secret stored or removed, the unit
// as it was, nothing reloaded or restarted -- and exit 0, so the next deploy asks
// again.
func TestDeployPodmanDeclinedWritesNothing(t *testing.T) {
	r := deployedPodman(t)
	const old = "[Container]\nImage=old\n"
	if err := os.WriteFile(r.unitPath, []byte(old), 0o644); err != nil {
		t.Fatal(err)
	}
	p := &podmanDeployRunner{stored: map[string]string{}, state: "active"} // every secret reads as missing: all changed
	code, stderr, asked := r.run(t, p, true, "n\n")
	if code != 0 || asked == "" || !strings.Contains(stderr, "deploy: cancelled; nothing changed") {
		t.Fatalf("exit=%d, asked %q, want 0 after a question and the cancelled message:\n%s", code, asked, stderr)
	}
	if r.unit(t) != old {
		t.Error("a declined deploy must leave the unit as it was")
	}
	for _, prefix := range [][]string{{"podman", "secret", "create"}, {"podman", "secret", "rm"},
		{"systemctl", "--user", "daemon-reload"}, {"systemctl", "--user", "start"}, {"systemctl", "--user", "restart"}} {
		if p.did(prefix...) {
			t.Errorf("a declined deploy ran %v: %+v", prefix, p.calls)
		}
	}
}

// With no terminal and no --no-prompt the question cannot be asked: exit 1 naming
// the flag, and nothing written.
func TestDeployPodmanNonTTYFailsNamingTheFlagAndWritesNothing(t *testing.T) {
	r := deployedPodman(t)
	const old = "[Container]\nImage=old\n"
	if err := os.WriteFile(r.unitPath, []byte(old), 0o644); err != nil {
		t.Fatal(err)
	}
	p := &podmanDeployRunner{stored: r.stored, state: "active"}
	code, stderr, _ := r.run(t, p, false, "")
	if code != 1 || !strings.Contains(stderr, noPromptFlagName) || !strings.Contains(stderr, "the redeploy") {
		t.Fatalf("exit=%d, want 1 naming %s:\n%s", code, noPromptFlagName, stderr)
	}
	if r.unit(t) != old || p.did("podman", "secret", "create") || p.did("systemctl", "--user", "restart") {
		t.Errorf("nothing may be written or restarted: %+v", p.calls)
	}
}

// --no-prompt approves without asking, so the changed unit is written and the
// running service restarted with no terminal.
func TestDeployPodmanNoPromptApproves(t *testing.T) {
	r := deployedPodman(t)
	if err := os.WriteFile(r.unitPath, []byte("[Container]\nImage=old\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	p := &podmanDeployRunner{stored: r.stored, state: "active"}
	code, stderr, asked := r.run(t, p, false, "", noPromptFlagName)
	if code != 0 || asked != "" {
		t.Fatalf("exit=%d, asked %q:\n%s", code, asked, stderr)
	}
	if strings.Contains(r.unit(t), "Image=old") || !p.did("systemctl", "--user", "restart", "solmq-conn.service") {
		t.Errorf("the changed unit is written and the service restarted, got %+v", p.calls)
	}
}

// A changed deploy of a service that is not running asks nothing -- nothing running
// is disturbed -- and starts it.
func TestDeployPodmanChangedButStoppedAsksNothing(t *testing.T) {
	r := deployedPodman(t)
	p := &podmanDeployRunner{stored: map[string]string{}, state: "failed"}
	code, stderr, asked := r.run(t, p, false, "")
	if code != 0 || asked != "" {
		t.Fatalf("exit=%d, asked %q:\n%s", code, asked, stderr)
	}
	if len(creates(p.calls)) != 6 || !p.did("systemctl", "--user", "start", "solmq-conn.service") || p.did("systemctl", "--user", "restart") {
		t.Errorf("every secret is stored and the unit started, got %+v", p.calls)
	}
}

// podmanChangeSummary names what differs and nothing else.
func TestPodmanChangeSummary(t *testing.T) {
	two := []podmanSecret{{name: "a", value: "v1"}, {name: "b", value: "v2"}}
	for _, c := range []struct {
		unit    bool
		secrets []podmanSecret
		want    string
	}{
		{true, nil, "its unit"},
		{false, two, "the secrets a, b"},
		{true, two[:1], "its unit and the secrets a"},
	} {
		if got := podmanChangeSummary(c.unit, c.secrets); got != c.want {
			t.Errorf("podmanChangeSummary(%v, %v) = %q, want %q", c.unit, c.secrets, got, c.want)
		}
	}
}

// A session's rootless answer only picks the hint a failed preflight carries: the
// rule deploy applies (rootless unless podman.rootful asks and root may), false off
// podman, and false rather than an error where the rule refuses -- the verbs that
// resolve the full scope report that refusal.
func TestSessionRootless(t *testing.T) {
	h := testHost(t.TempDir()) // euid 1000, not root
	podman := func(rootful bool) *spec.Env { return &spec.Env{Podman: &spec.Podman{Rootful: rootful}} }
	for _, c := range []struct {
		name string
		s    instanceSession
		want bool
	}{
		{"podman, rootless by default", instanceSession{platform: "podman", env: podman(false), host: h}, true},
		{"podman with no env.yaml at all", instanceSession{platform: "podman", host: h}, true},
		{"podman with no podman section", instanceSession{platform: "podman", env: &spec.Env{}, host: h}, true},
		{"rootful asked by a non-root user is refused, so not rootless", instanceSession{platform: "podman", env: podman(true), host: h}, false},
		{"docker", instanceSession{platform: "docker", env: podman(false), host: h}, false},
		{"kubernetes", instanceSession{platform: "kubernetes", host: h}, false},
	} {
		if got := c.s.rootless(); got != c.want {
			t.Errorf("%s: rootless() = %v, want %v", c.name, got, c.want)
		}
	}
}
