package main

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// start, stop and restart act on the instance env.yaml names, in place: the
// platform's preflight probe, then one call -- and on kubernetes, for start and
// restart, a wait for the rollout. They need no -env variable --
// the spec's credentials are all -env here and none is exported -- and no
// store file, since nothing is rendered or resolved.
func TestLifecycleVerbsActInPlace(t *testing.T) {
	podmanQuadletHome(t) // podman resolves to the invoking user's units
	envFile := func(t *testing.T, section string) string {
		dir := t.TempDir()
		write(t, dir, "env.yaml", sharedEnv+"\ntls:\n  stores:\n    ts:\n      password-env: SOLMQ_UNSET_TS_PW\n      existing:\n        pkcs12: ./gone.p12\n"+section)
		write(t, dir, "10.yaml", tlsWF)
		return filepath.Join(dir, "env.yaml")
	}
	kube := strings.Replace(kubeEnv, "namespace: solace-connectors\n", "namespace: solace-connectors\n    replicas: 2\n", 1)
	for _, c := range []struct {
		verb, platform, section string
		want                    [][]string
	}{
		// stop has nothing to wait for: the pods are going away, not rolling.
		{"stop", "kubernetes", kube, [][]string{
			{"kubectl", "auth", "can-i", "patch", "deployments", "-n", "solace-connectors"},
			{"kubectl", "scale", "deployment/solmq-connector", "-n", "solace-connectors", "--replicas=0"}}},
		{"start", "kubernetes", kube, [][]string{
			{"kubectl", "auth", "can-i", "patch", "deployments", "-n", "solace-connectors"},
			{"kubectl", "scale", "deployment/solmq-connector", "-n", "solace-connectors", "--replicas=2"},
			{"kubectl", "rollout", "status", "deployment/solmq-connector", "-n", "solace-connectors", "--timeout=300s"}}},
		{"restart", "kubernetes", kube, [][]string{
			{"kubectl", "auth", "can-i", "patch", "deployments", "-n", "solace-connectors"},
			{"kubectl", "rollout", "restart", "deployment/solmq-connector", "-n", "solace-connectors"},
			{"kubectl", "rollout", "status", "deployment/solmq-connector", "-n", "solace-connectors", "--timeout=300s"}}},
		{"stop", "docker", dockerEnv, [][]string{{"docker", "info"}, {"docker", "stop", "solmq-conn"}}},
		{"start", "docker", dockerEnv, [][]string{{"docker", "info"}, {"docker", "start", "solmq-conn"}}},
		{"restart", "docker", dockerEnv, [][]string{{"docker", "info"}, {"docker", "restart", "solmq-conn"}}},
		{"stop", "podman", podmanEnv(), [][]string{{"podman", "info"}, {"systemctl", "--user", "stop", "solmq-conn.service"}}},
		{"start", "podman", podmanEnv(), [][]string{{"podman", "info"}, {"systemctl", "--user", "start", "solmq-conn.service"}}},
		{"restart", "podman", podmanEnv(), [][]string{{"podman", "info"}, {"systemctl", "--user", "restart", "solmq-conn.service"}}},
	} {
		t.Run(c.verb+" "+c.platform, func(t *testing.T) {
			f := useFakeRunner(t)
			var code int
			stderr := captureStderr(t, func() {
				code = dispatch([]string{c.verb, "--platform", c.platform, "-e", envFile(t, c.section)}, f)
			})
			if code != 0 {
				t.Fatalf("exit=%d\n%s", code, stderr)
			}
			if len(f.calls) != len(c.want) {
				t.Fatalf("calls = %+v, want %v", f.calls, c.want)
			}
			for i, w := range c.want {
				if !reflect.DeepEqual(f.calls[i].argv, w) {
					t.Errorf("call %d = %v, want %v", i, f.calls[i].argv, w)
				}
			}
			if !strings.Contains(stderr, "ok: "+c.verb+" "+c.platform) {
				t.Errorf("stderr = %q, want the ok line", stderr)
			}
		})
	}
}

// A failed preflight stops a lifecycle verb before it acts; a name env.yaml
// should never carry, and a positional argument, are refused before anything
// runs.
func TestLifecycleVerbsRefuseBeforeActing(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "env.yaml", sharedEnv+dockerEnv)
	write(t, dir, "10.yaml", validWF)
	env := filepath.Join(dir, "env.yaml")

	f := &fakeRunner{err: errors.New("Cannot connect to the Docker daemon")}
	if code := dispatch([]string{"stop", "--platform", "docker", "-e", env}, f); code == 0 || len(f.calls) != 1 {
		t.Errorf("a failed preflight must stop before acting: exit %d, calls %+v", code, f.calls)
	}

	write(t, dir, "env.yaml", sharedEnv+strings.Replace(dockerEnv, "name: solmq-conn", "name: --all", 1))
	f = &fakeRunner{}
	var code int
	stderr := captureStderr(t, func() { code = dispatch([]string{"restart", "--platform", "docker", "-e", env}, f) })
	if code != 2 || len(f.calls) != 0 || !strings.Contains(stderr, `docker.name "--all" is not a valid DNS-1123 label`) {
		t.Errorf("an unsafe name must be refused before anything runs: exit %d, calls %+v\n%s", code, f.calls, stderr)
	}

	f = &fakeRunner{}
	if code := dispatch([]string{"start", "docker", "-e", env}, f); code != 2 || len(f.calls) != 0 {
		t.Errorf("a positional argument is a usage error: exit %d, calls %+v", code, f.calls)
	}
	if _, err := os.Stat(filepath.Join(dir, "docker-compose.yml")); !os.IsNotExist(err) {
		t.Errorf("a lifecycle verb writes no compose file: %v", err)
	}
}

// The rollout wait after start and restart is a convenience, and the change is already
// made: a wait kubectl refuses warns and exits 0, while one that times out, or sees
// the rollout fail, is the verb's failure. stop never waits, so it never meets either.
func TestLifecycleRolloutWait(t *testing.T) {
	forbidden := `Error from server (Forbidden): deployments.apps "solmq-connector" is forbidden: User "u" cannot get resource "deployments"`
	exit1 := errors.New("exit status 1")
	for _, c := range []struct {
		name, verb string
		wait       queuedResp
		wantCode   int
		want       []string
	}{
		{"start, refused", "start", queuedResp{forbidden, exit1}, 0,
			[]string{"warning: the wait for deployment/solmq-connector to roll out was refused", "the change was applied", "ok: start kubernetes"}},
		{"restart, refused", "restart", queuedResp{forbidden, exit1}, 0,
			[]string{"warning: the wait for deployment/solmq-connector to roll out was refused", "ok: restart kubernetes"}},
		{"start, timed out", "start", queuedResp{"error: timed out waiting for the condition", exit1}, 1,
			[]string{"error: start kubernetes failed", "deployment/solmq-connector did not finish rolling out"}},
		{"restart, rollout failed", "restart", queuedResp{`error: deployment "solmq-connector" exceeded its progress deadline`, exit1}, 1,
			[]string{"error: restart kubernetes failed", "exceeded its progress deadline"}},
	} {
		t.Run(c.name, func(t *testing.T) {
			dir := t.TempDir()
			write(t, dir, "env.yaml", sharedEnv+kubeEnv)
			write(t, dir, "10.yaml", validWF)
			// patch probe, the scale or restart, then the wait.
			q := &queueRunner{resp: []queuedResp{{"yes\n", nil}, {"", nil}, c.wait}}
			var code int
			stderr := captureStderr(t, func() {
				code = dispatch([]string{c.verb, "--platform", "kubernetes", "-e", filepath.Join(dir, "env.yaml")}, q)
			})
			if code != c.wantCode {
				t.Fatalf("exit = %d, want %d\n%s", code, c.wantCode, stderr)
			}
			for _, want := range c.want {
				if !strings.Contains(stderr, want) {
					t.Errorf("stderr lacks %q:\n%s", want, stderr)
				}
			}
			if len(q.calls) != 3 {
				t.Errorf("want probe, the verb and one wait, got %+v", q.calls)
			}
		})
	}

	// stop makes no wait: two calls, and a refusal-shaped answer is never read.
	q := &queueRunner{resp: []queuedResp{{"yes\n", nil}, {"", nil}, {forbidden, exit1}}}
	dir := t.TempDir()
	write(t, dir, "env.yaml", sharedEnv+kubeEnv)
	write(t, dir, "10.yaml", validWF)
	if code := dispatch([]string{"stop", "--platform", "kubernetes", "-e", filepath.Join(dir, "env.yaml")}, q); code != 0 || len(q.calls) != 2 {
		t.Errorf("stop must not wait: exit %d, calls %+v", code, q.calls)
	}
}
