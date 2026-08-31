package statusreport

import (
	"strings"
	"testing"
)

// The documents below are the shapes the engines actually answer with, trimmed
// to the fields this package reads. They are the fixtures every parse case
// works from; a literal is only built inline where a case is deliberately
// malformed.

// podListJSON is two pods of one deployment: one healthy and one crash-looping,
// which is the pairing the whole container view exists for.
const podListJSON = `{
  "apiVersion": "v1",
  "kind": "List",
  "items": [
    {
      "metadata": {"name": "solmq-connector-7d9f8c-x2n4q", "namespace": "prod", "creationTimestamp": "2026-08-18T04:11:00Z"},
      "spec": {
        "nodeName": "node-a",
        "containers": [{
          "name": "connector",
          "image": "solace/solace-pubsub-connector-ibmmq:2.14.1",
          "readinessProbe": {"tcpSocket": {"port": 8090}},
          "resources": {"limits": {"cpu": "1", "memory": "1Gi"}},
          "volumeMounts": [
            {"name": "config", "mountPath": "/app/external/spring/config/application.yml"},
            {"name": "secrets", "mountPath": "/run/secrets"},
            {"name": "libs", "mountPath": "/app/external/libs"}
          ],
          "envFrom": [{"secretRef": {"name": "solmq-connector-extra-env"}}]
        }],
        "volumes": [
          {"name": "config", "configMap": {"name": "solmq-connector-config"}},
          {"name": "secrets", "secret": {"secretName": "solmq-connector-credentials"}},
          {"name": "libs", "persistentVolumeClaim": {"claimName": "solmq-libs"}},
          {"name": "scratch", "emptyDir": {}}
        ],
        "imagePullSecrets": [{"name": "solmq-pull"}]
      },
      "status": {
        "phase": "Running",
        "startTime": "2026-08-18T04:11:05Z",
        "containerStatuses": [{
          "name": "connector",
          "ready": true,
          "restartCount": 0,
          "image": "docker.io/solace/solace-pubsub-connector-ibmmq:2.14.1",
          "imageID": "docker-pullable://solace/solace-pubsub-connector-ibmmq@sha256:9f2a1c4d8e",
          "state": {"running": {"startedAt": "2026-08-18T04:12:07Z"}}
        }]
      }
    },
    {
      "metadata": {"name": "solmq-connector-7d9f8c-k8m1r", "namespace": "prod"},
      "spec": {
        "nodeName": "node-b",
        "containers": [{
          "name": "connector",
          "image": "solace/solace-pubsub-connector-ibmmq:2.14.1",
          "readinessProbe": {"tcpSocket": {"port": 8090}}
        }]
      },
      "status": {
        "phase": "Running",
        "startTime": "2026-08-18T04:11:05Z",
        "containerStatuses": [{
          "name": "connector",
          "ready": false,
          "restartCount": 7,
          "image": "docker.io/solace/solace-pubsub-connector-ibmmq:2.14.1",
          "imageID": "docker-pullable://solace/solace-pubsub-connector-ibmmq@sha256:9f2a1c4d8e",
          "state": {"waiting": {"reason": "CrashLoopBackOff", "message": "back-off 5m0s"}},
          "lastState": {"terminated": {"exitCode": 137, "reason": "OOMKilled"}}
        }]
      }
    }
  ]
}`

func TestParsePodsReadsBothHalvesOfARealPair(t *testing.T) {
	insts, err := ParsePods(podListJSON, now, "")
	if err != nil {
		t.Fatalf("ParsePods: %v", err)
	}
	if len(insts) != 2 {
		t.Fatalf("got %d instances, want 2", len(insts))
	}

	up := insts[0]
	if up.Name != "solmq-connector-7d9f8c-x2n4q" || up.Namespace != "prod" {
		t.Errorf("identity = %q/%q", up.Namespace, up.Name)
	}
	c := up.Container
	if c.State != StateRunning {
		t.Errorf("state = %q, want %q", c.State, StateRunning)
	}
	if c.Ready != "yes" {
		t.Errorf("ready = %q, want yes (the pod declares a readiness probe)", c.Ready)
	}
	if c.Restarts != 0 || c.ExitCode != nil {
		t.Errorf("a pod that has never restarted carries no exit code: restarts=%d exit=%v", c.Restarts, c.ExitCode)
	}
	// The container's own running-since stamp wins over the pod's startTime.
	if c.StartedAt != "2026-08-18T04:12:07Z" {
		t.Errorf("startedAt = %q, want the container's own", c.StartedAt)
	}
	if c.Digest != "sha256:9f2a1c4d8e" {
		t.Errorf("digest = %q, want the digest alone out of the imageID", c.Digest)
	}
	if c.Node != "node-a" {
		t.Errorf("node = %q", c.Node)
	}
	if c.CPU == nil || c.CPU.Limit != "1" || c.Memory == nil || c.Memory.Limit != "1Gi" {
		t.Errorf("limits should be read even before any usage sample: cpu=%v memory=%v", c.CPU, c.Memory)
	}
	if c.CPU.Used != "" {
		t.Errorf("no usage is collected without a top sample, got %q", c.CPU.Used)
	}

	down := insts[1].Container
	// A back-off is the engine waiting between restarts, which is what an
	// operator means by "restarting"; the engine's own word is kept beside it.
	if down.State != StateRestarting {
		t.Errorf("state = %q, want %q", down.State, StateRestarting)
	}
	if down.Reason != "CrashLoopBackOff" {
		t.Errorf("reason = %q", down.Reason)
	}
	if down.Restarts != 7 {
		t.Errorf("restarts = %d, want 7", down.Restarts)
	}
	// The previous termination is the useful one on a looping pod: it carries
	// the code that killed it.
	if down.ExitCode == nil || *down.ExitCode != 137 {
		t.Errorf("exit code = %v, want the last termination's 137", down.ExitCode)
	}
	if down.Ready != "no" {
		t.Errorf("ready = %q, want no", down.Ready)
	}
}

func TestParsePodsComponentsComeFromWhatThePodReferences(t *testing.T) {
	insts, err := ParsePods(podListJSON, now, "")
	if err != nil {
		t.Fatal(err)
	}
	got := map[string]Component{}
	for _, comp := range insts[0].Container.Components {
		got[comp.Kind+"/"+comp.Name] = comp
	}
	for _, want := range []string{
		"configmap/solmq-connector-config",
		"secret/solmq-connector-credentials",
		"persistentvolumeclaim/solmq-libs",
		"secret/solmq-connector-extra-env", // envFrom, not a volume
		"secret/solmq-pull",                // imagePullSecrets
	} {
		if _, ok := got[want]; !ok {
			t.Errorf("missing component %q; got %v", want, keys(got))
		}
	}
	// An emptyDir has no object to check, so it is not a component.
	for k := range got {
		if strings.Contains(k, "scratch") {
			t.Errorf("emptyDir should not be listed as a component: %q", k)
		}
	}
	// A mounted volume carries where it is mounted; an envFrom reference says so
	// instead, since there is no path.
	if d := got["configmap/solmq-connector-config"].Detail; d != "/app/external/spring/config/application.yml" {
		t.Errorf("configmap detail = %q, want its mount path", d)
	}
	if d := got["secret/solmq-connector-extra-env"].Detail; d != "envFrom" {
		t.Errorf("envFrom detail = %q", d)
	}
	// Nothing is claimed about existence here: that is a separate probe.
	if s := got["configmap/solmq-connector-config"].Status; s != "" {
		t.Errorf("parsing must not claim a status, got %q", s)
	}
}

func keys(m map[string]Component) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}

func TestParsePodsSingleObjectDocument(t *testing.T) {
	// `kubectl get pod <name> -o json` answers the object itself, with no items
	// array; both shapes have to parse, since status uses both forms.
	doc := `{"kind":"Pod","metadata":{"name":"pod-a","namespace":"prod"},
	  "spec":{"containers":[{"name":"connector","image":"solace/x:1"}]},
	  "status":{"phase":"Running","containerStatuses":[{"name":"connector","ready":true,"restartCount":0,"state":{"running":{"startedAt":"2026-08-21T11:00:00Z"}}}]}}`
	insts, err := ParsePods(doc, now, "")
	if err != nil {
		t.Fatalf("ParsePods: %v", err)
	}
	if len(insts) != 1 || insts[0].Name != "pod-a" {
		t.Fatalf("got %+v", insts)
	}
	if insts[0].Container.Age != "1h0m" {
		t.Errorf("age = %q, want 1h0m", insts[0].Container.Age)
	}
}

func TestParsePodsPendingAndTerminatedStates(t *testing.T) {
	cases := []struct {
		name       string
		doc        string
		wantState  string
		wantReason string
	}{
		{
			// No container status yet: the phase is all there is, and it is still
			// worth a row -- a pod that cannot be scheduled is exactly what someone
			// is looking for.
			name:      "pending, no container status",
			doc:       `{"kind":"Pod","metadata":{"name":"p"},"spec":{"containers":[{"name":"connector"}]},"status":{"phase":"Pending"}}`,
			wantState: StateWaiting,
		},
		{
			name:       "waiting on an image pull",
			doc:        `{"kind":"Pod","metadata":{"name":"p"},"spec":{"containers":[{"name":"connector"}]},"status":{"phase":"Pending","containerStatuses":[{"name":"connector","state":{"waiting":{"reason":"ImagePullBackOff"}}}]}}`,
			wantState:  StateWaiting,
			wantReason: "ImagePullBackOff",
		},
		{
			name:       "terminated",
			doc:        `{"kind":"Pod","metadata":{"name":"p"},"spec":{"containers":[{"name":"connector"}]},"status":{"phase":"Failed","containerStatuses":[{"name":"connector","restartCount":2,"state":{"terminated":{"exitCode":1,"reason":"Error"}}}]}}`,
			wantState:  StateExited,
			wantReason: "Error",
		},
	}
	for _, c := range cases {
		insts, err := ParsePods(c.doc, now, "")
		if err != nil {
			t.Fatalf("%s: %v", c.name, err)
		}
		got := insts[0].Container
		if got.State != c.wantState {
			t.Errorf("%s: state = %q, want %q", c.name, got.State, c.wantState)
		}
		if got.Reason != c.wantReason {
			t.Errorf("%s: reason = %q, want %q", c.name, got.Reason, c.wantReason)
		}
	}
}

func TestParsePodsReadinessIsOnlyAVerdictWhenAProbeExists(t *testing.T) {
	// Without a readiness probe kubernetes reports ready as soon as the
	// container runs, which says nothing the state column has not already said.
	doc := `{"kind":"Pod","metadata":{"name":"p"},"spec":{"containers":[{"name":"connector"}]},
	  "status":{"phase":"Running","containerStatuses":[{"name":"connector","ready":true,"state":{"running":{"startedAt":"2026-08-21T11:00:00Z"}}}]}}`
	insts, err := ParsePods(doc, now, "")
	if err != nil {
		t.Fatal(err)
	}
	if got := insts[0].Container.Ready; got != NotApplicable {
		t.Errorf("ready = %q, want %q when the pod declares no readiness probe", got, NotApplicable)
	}
}

func TestParsePodsPicksTheConnectorContainer(t *testing.T) {
	// A sidecar must not be reported as the connector; the container this tool
	// names in its manifests is the one the report is about.
	doc := `{"kind":"Pod","metadata":{"name":"p"},"spec":{"containers":[
	    {"name":"sidecar","image":"other:1"},
	    {"name":"connector","image":"solace/x:1","resources":{"limits":{"memory":"2Gi"}}}]},
	  "status":{"phase":"Running","containerStatuses":[
	    {"name":"sidecar","ready":true,"restartCount":99,"state":{"running":{"startedAt":"2026-08-21T11:00:00Z"}}},
	    {"name":"connector","ready":true,"restartCount":1,"image":"solace/x:1","state":{"running":{"startedAt":"2026-08-21T11:00:00Z"}}}]}}`
	insts, err := ParsePods(doc, now, "")
	if err != nil {
		t.Fatal(err)
	}
	c := insts[0].Container
	if c.Restarts != 1 {
		t.Errorf("restarts = %d, want the connector's 1 rather than the sidecar's 99", c.Restarts)
	}
	if c.Memory == nil || c.Memory.Limit != "2Gi" {
		t.Errorf("limits should come from the connector container, got %v", c.Memory)
	}
}

func TestParsePodsSeveralContainersNoneNamedConnector(t *testing.T) {
	// Guessing which of them is the connector would be worse than reporting the
	// pod alone, so no container-level fact is claimed.
	doc := `{"kind":"Pod","metadata":{"name":"p"},"spec":{"containers":[{"name":"a"},{"name":"b"}]},
	  "status":{"phase":"Running","containerStatuses":[
	    {"name":"a","ready":true,"restartCount":3},{"name":"b","ready":false,"restartCount":4}]}}`
	insts, err := ParsePods(doc, now, "")
	if err != nil {
		t.Fatal(err)
	}
	c := insts[0].Container
	if c.State != StateRunning {
		t.Errorf("the pod phase still gives a state, got %q", c.State)
	}
	if c.Restarts != 0 || c.Ready != "" {
		t.Errorf("no container-level fact should be claimed: restarts=%d ready=%q", c.Restarts, c.Ready)
	}
}

func TestParsePodsImageFilterIsWhatAllSearchesBy(t *testing.T) {
	doc := `{"items":[
	  {"metadata":{"name":"ours"},"spec":{"containers":[{"name":"connector","image":"reg.internal/solace/solace-pubsub-connector-ibmmq:2.14.1"}]},"status":{"phase":"Running"}},
	  {"metadata":{"name":"theirs"},"spec":{"containers":[{"name":"app","image":"nginx:1.27"}]},"status":{"phase":"Running"}}]}`
	insts, err := ParsePods(doc, now, ImageMatch)
	if err != nil {
		t.Fatal(err)
	}
	if len(insts) != 1 || insts[0].Name != "ours" {
		t.Fatalf("the filter should keep only the connector pod, got %+v", insts)
	}
	// Without a filter, both are reported: the filter is --all's, not the
	// parser's own opinion.
	all, err := ParsePods(doc, now, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 2 {
		t.Errorf("unfiltered parse should keep both, got %d", len(all))
	}
}

func TestParsePodsErrorsAndSkips(t *testing.T) {
	if _, err := ParsePods("", now, ""); err == nil {
		t.Error("an empty response is an error, not an empty report")
	}
	if _, err := ParsePods("not json", now, ""); err == nil {
		t.Error("an undecodable response is an error")
	}
	// One unreadable item in a list is skipped so the rest of the report
	// survives -- skip the item, not the run.
	doc := `{"items":[ "not-an-object", {"metadata":{"name":"pod-a"},"status":{"phase":"Running"}} ]}`
	insts, err := ParsePods(doc, now, "")
	if err != nil {
		t.Fatalf("one bad item must not fail the parse: %v", err)
	}
	if len(insts) != 1 || insts[0].Name != "pod-a" {
		t.Errorf("got %+v", insts)
	}
}

func TestParseDeploymentAndService(t *testing.T) {
	dep := `{"kind":"Deployment","metadata":{"name":"solmq-connector"},"spec":{"replicas":2},
	  "status":{"readyReplicas":1,"updatedReplicas":1,"availableReplicas":2}}`
	w, err := ParseDeployment(dep)
	if err != nil {
		t.Fatalf("ParseDeployment: %v", err)
	}
	if w.Desired != 2 || w.Ready != 1 || w.UpToDate != 1 || w.Available != 2 {
		t.Errorf("counts = %+v", w)
	}

	// replicas omitted means one, the API's own default.
	w2, err := ParseDeployment(`{"kind":"Deployment","metadata":{"name":"x"},"spec":{},"status":{}}`)
	if err != nil {
		t.Fatal(err)
	}
	if w2.Desired != 1 {
		t.Errorf("desired = %d, want the API default of 1", w2.Desired)
	}

	svc := `{"kind":"Service","metadata":{"name":"solmq-connector"},"spec":{"ports":[{"port":8090},{"port":9090,"protocol":"UDP"}]}}`
	merged, err := MergeService(w, svc)
	if err != nil {
		t.Fatalf("MergeService: %v", err)
	}
	if merged.Deployment != "solmq-connector" {
		t.Error("merging a service must not lose the deployment counts")
	}
	if got := strings.Join(merged.ServicePorts, ","); got != "8090/TCP,9090/UDP" {
		t.Errorf("ports = %q (an omitted protocol is TCP, the API default)", got)
	}

	// A service read without a readable deployment still reports.
	only, err := MergeService(nil, svc)
	if err != nil || only == nil || only.Service != "solmq-connector" {
		t.Errorf("service-only workload = %+v, err %v", only, err)
	}
	if _, err := MergeService(w, "nope"); err == nil {
		t.Error("an undecodable service document is an error")
	}
}

func TestObjectExists(t *testing.T) {
	cases := []struct {
		name       string
		doc        string
		wantOK     bool
		wantStatus string
	}{
		{"secret", `{"kind":"Secret","metadata":{"name":"s"}}`, true, "present"},
		// A claim's phase says more than "present", and is the only status here
		// that can be bad while the object exists.
		{"bound claim", `{"kind":"PersistentVolumeClaim","status":{"phase":"Bound"}}`, true, "Bound"},
		{"pending claim", `{"kind":"PersistentVolumeClaim","status":{"phase":"Pending"}}`, true, "Pending"},
		{"not a document", `oops`, false, ""},
		{"empty", ``, false, ""},
	}
	for _, c := range cases {
		ok, status := ObjectExists(c.doc)
		if ok != c.wantOK || status != c.wantStatus {
			t.Errorf("%s: got (%v, %q), want (%v, %q)", c.name, ok, status, c.wantOK, c.wantStatus)
		}
	}
}

func TestApplyTop(t *testing.T) {
	insts := []Instance{
		{Name: "pod-a", Container: &Container{CPU: &Resource{Limit: "1"}, Memory: &Resource{Limit: "1Gi"}}},
		{Name: "pod-b", Container: &Container{}},
		{Name: "pod-none", Container: &Container{}},
	}
	// --containers rows: pod, container, cpu, memory. A sidecar row for the same
	// pod must not win over the connector's.
	out := "pod-a   sidecar     900m   900Mi\n" +
		"pod-a   connector   120m   512Mi\n" +
		"pod-b   connector   50m    64Mi\n" +
		"\n"
	ApplyTop(insts, out)

	a := insts[0].Container
	if a.CPU.Used != "120m" || a.CPU.Percent != "12%" {
		t.Errorf("pod-a cpu = %+v, want the connector row plus a percentage of its limit", a.CPU)
	}
	if a.Memory.Used != "512Mi" || a.Memory.Percent != "50%" {
		t.Errorf("pod-a memory = %+v", a.Memory)
	}
	// A pod with no limits still reports usage, just without a percentage.
	b := insts[1].Container
	if b.CPU == nil || b.CPU.Used != "50m" || b.CPU.Percent != "" {
		t.Errorf("pod-b cpu = %+v", b.CPU)
	}
	// A pod the metrics API said nothing about keeps no usage at all.
	if insts[2].Container.CPU != nil {
		t.Errorf("pod-none should have no usage, got %+v", insts[2].Container.CPU)
	}
}

// ---- docker / podman ---------------------------------------------------------

const dockerInspectJSON = `[
  {
    "Name": "/solmq-connector",
    "Created": "2026-08-18T04:10:00Z",
    "State": {
      "Status": "running", "Running": true, "ExitCode": 0,
      "StartedAt": "2026-08-18T04:12:07Z",
      "Health": {"Status": "healthy"}
    },
    "RestartCount": 0,
    "Image": "sha256:localimageid",
    "Config": {
      "Image": "solace/solace-pubsub-connector-ibmmq:2.14.1",
      "Labels": {"com.docker.compose.project": "eg"}
    },
    "HostConfig": {"NanoCpus": 1000000000, "Memory": 1073741824},
    "Mounts": [
      {"Type": "bind", "Source": "/srv/solmq/application.yml", "Destination": "/app/external/spring/config/application.yml"},
      {"Type": "volume", "Name": "solmq-libs", "Destination": "/app/external/libs"}
    ],
    "NetworkSettings": {"Networks": {"eg_default": {}}}
  }
]`

func TestParseInspectDocker(t *testing.T) {
	insts, err := ParseInspect(dockerInspectJSON, now, "")
	if err != nil {
		t.Fatalf("ParseInspect: %v", err)
	}
	if len(insts) != 1 {
		t.Fatalf("got %d instances", len(insts))
	}
	inst := insts[0]
	// docker prefixes a container name with a slash; podman does not.
	if inst.Name != "solmq-connector" {
		t.Errorf("name = %q, want the slash stripped", inst.Name)
	}
	// The compose project comes off the container's own label, in the same call
	// that answered everything else.
	if inst.Group != "eg" {
		t.Errorf("group = %q, want the compose project", inst.Group)
	}
	c := inst.Container
	if c.State != StateRunning || c.Health != "healthy" {
		t.Errorf("state = %q, health = %q", c.State, c.Health)
	}
	// Readiness is a kubernetes concept; docker has none.
	if c.Ready != NotApplicable {
		t.Errorf("ready = %q, want %q", c.Ready, NotApplicable)
	}
	if c.Image != "solace/solace-pubsub-connector-ibmmq:2.14.1" {
		t.Errorf("image = %q, want the configured reference rather than the local id", c.Image)
	}
	if c.CPU == nil || c.CPU.Limit != "1 cpu" {
		t.Errorf("cpu ceiling = %+v, want the nanocpu quota named with its unit", c.CPU)
	}
	if c.Memory == nil || c.Memory.Limit != "1Gi" {
		t.Errorf("memory ceiling = %+v", c.Memory)
	}
	if c.Age != "3d7h" {
		t.Errorf("age = %q", c.Age)
	}
	kinds := map[string]string{}
	for _, comp := range c.Components {
		kinds[comp.Kind+"/"+comp.Name] = comp.Status
	}
	for _, want := range []string{"bind//srv/solmq/application.yml", "volume/solmq-libs", "network/eg_default"} {
		if kinds[want] != "attached" {
			t.Errorf("component %q = %q, want attached (it exists by definition)", want, kinds[want])
		}
	}
}

func TestParseInspectStatesAndHealthSpellings(t *testing.T) {
	cases := []struct {
		name        string
		doc         string
		wantState   string
		wantHealth  string
		wantReason  string
		wantExitSet bool
	}{
		{
			name:        "exited with a code",
			doc:         `[{"Name":"c","State":{"Status":"exited","ExitCode":1},"Config":{"Image":"i"}}]`,
			wantState:   StateExited,
			wantHealth:  NotApplicable,
			wantExitSet: true,
		},
		{
			name:        "oom killed",
			doc:         `[{"Name":"c","State":{"Status":"exited","ExitCode":137,"OOMKilled":true},"Config":{"Image":"i"}}]`,
			wantState:   StateExited,
			wantHealth:  NotApplicable,
			wantReason:  "OOMKilled",
			wantExitSet: true,
		},
		{"restarting", `[{"Name":"c","State":{"Status":"restarting"},"Config":{"Image":"i"}}]`, StateRestarting, NotApplicable, "", false},
		{"paused", `[{"Name":"c","State":{"Status":"paused"},"Config":{"Image":"i"}}]`, StatePaused, NotApplicable, "", false},
		{"created has not started", `[{"Name":"c","State":{"Status":"created"},"Config":{"Image":"i"}}]`, StateWaiting, NotApplicable, "", false},
		// A clean stop reports no code: "exited" already says it stopped, and
		// "exited (exit code 0)" would be noise in the column meant for a
		// diagnosis.
		{"podman stopped cleanly", `[{"Name":"c","State":{"Status":"stopped","ExitCode":0},"Config":{"Image":"i"}}]`, StateExited, NotApplicable, "", false},
		{"a status this tool has not seen", `[{"Name":"c","State":{"Status":"wat"},"Config":{"Image":"i"}}]`, StateUnknown, NotApplicable, "", false},
		// podman has spelled the healthcheck block both ways across versions.
		{"podman healthcheck key", `[{"Name":"c","State":{"Status":"running","Healthcheck":{"Status":"starting"}},"Config":{"Image":"i"}}]`, StateRunning, "starting", "", false},
		// The compose and quadlet artifacts this tool generates declare no
		// healthcheck, so this is the common case.
		{"no healthcheck at all", `[{"Name":"c","State":{"Status":"running"},"Config":{"Image":"i"}}]`, StateRunning, NotApplicable, "", false},
	}
	for _, c := range cases {
		insts, err := ParseInspect(c.doc, now, "")
		if err != nil {
			t.Fatalf("%s: %v", c.name, err)
		}
		got := insts[0].Container
		if got.State != c.wantState {
			t.Errorf("%s: state = %q, want %q", c.name, got.State, c.wantState)
		}
		if got.Health != c.wantHealth {
			t.Errorf("%s: health = %q, want %q", c.name, got.Health, c.wantHealth)
		}
		if got.Reason != c.wantReason {
			t.Errorf("%s: reason = %q, want %q", c.name, got.Reason, c.wantReason)
		}
		if (got.ExitCode != nil) != c.wantExitSet {
			t.Errorf("%s: exit code set = %v, want %v", c.name, got.ExitCode != nil, c.wantExitSet)
		}
	}
}

func TestParseInspectFilterAndErrors(t *testing.T) {
	doc := `[{"Name":"/ours","State":{"Status":"running"},"Config":{"Image":"solace/solace-pubsub-connector-ibmmq:2.14.1"}},
	         {"Name":"/theirs","State":{"Status":"running"},"Config":{"Image":"redis:7"}}]`
	insts, err := ParseInspect(doc, now, ImageMatch)
	if err != nil {
		t.Fatal(err)
	}
	if len(insts) != 1 || insts[0].Name != "ours" {
		t.Errorf("got %+v", insts)
	}
	if _, err := ParseInspect("", now, ""); err == nil {
		t.Error("an empty response is an error")
	}
	if _, err := ParseInspect("{not an array}", now, ""); err == nil {
		t.Error("an undecodable response is an error")
	}
}

func TestParseImageDigest(t *testing.T) {
	doc := `[{"Id":"sha256:local","RepoDigests":["solace/x@sha256:9f2a1c","solace/x@sha256:other"]}]`
	if got := ParseImageDigest(doc); got != "sha256:9f2a1c" {
		t.Errorf("got %q, want the first repo digest", got)
	}
	// An image built locally and never pushed carries none, which is not an
	// error: the digest line is simply left out.
	if got := ParseImageDigest(`[{"Id":"sha256:local","RepoDigests":[]}]`); got != "" {
		t.Errorf("got %q, want empty", got)
	}
	if got := ParseImageDigest(`nope`); got != "" {
		t.Errorf("got %q, want empty", got)
	}
}

func TestApplyStats(t *testing.T) {
	insts := []Instance{
		{Name: "solmq-connector", Container: &Container{CPU: &Resource{Limit: "1 cpu"}, Memory: &Resource{Limit: "1Gi"}}},
		{Name: "unsampled", Container: &Container{}},
	}
	// The tab-separated template both engines render the same way.
	out := "solmq-connector\t0.15%\t512MiB / 1GiB\t50.00%\n\t\t\t\n"
	ApplyStats(insts, out)

	c := insts[0].Container
	if c.CPU.Used != "0.15%" || c.CPU.Limit != "1 cpu" {
		t.Errorf("cpu = %+v, want the engine's percentage against the recorded quota", c.CPU)
	}
	// The engine's own memory string already carries both sides, so it is kept
	// whole rather than re-spelled -- and its percentage is the engine's.
	if c.Memory.Used != "512MiB / 1GiB" || c.Memory.Limit != "" || c.Memory.Percent != "50.00%" {
		t.Errorf("memory = %+v", c.Memory)
	}
	if insts[1].Container.CPU != nil {
		t.Errorf("a container with no sample keeps no usage, got %+v", insts[1].Container.CPU)
	}
}

func TestEngineNamesByImage(t *testing.T) {
	out := "solmq-connector\tsolace/solace-pubsub-connector-ibmmq:2.14.1\n" +
		"old-connector\treg.internal/solace/solace-pubsub-connector-ibmmq:2.13.0\n" +
		"redis\tredis:7\n" +
		"broken-row\n" +
		"\n"
	got := EngineNamesByImage(out, ImageMatch)
	want := []string{"solmq-connector", "old-connector"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Errorf("got %v, want %v", got, want)
	}
}

// ---- the in-container script's report ----------------------------------------

// scriptOutput is what the generated script prints on a healthy active
// instance at the details level, including the stderr note that arrives on the
// same combined stream.
const scriptOutput = `leader-election mode: active_standby
leader-election state: active
health: UP
health components:
  solace: UP
  ibmmq: UP
uptime: 3d 4h 31m
version: 2.14.1
java: openjdk 17.0.9
config: /app/external/spring/config/application.yml
heap: 432013312 of 1073741824
workflows:
   0: running
  10: stopped
status: workflows is not in the exposure list
`

func TestParseApplication(t *testing.T) {
	app := ParseApplication(scriptOutput)
	if app.LeaderElectionMode != "active_standby" || app.LeaderElectionState != "active" {
		t.Errorf("leader election = %q/%q", app.LeaderElectionMode, app.LeaderElectionState)
	}
	if app.Health != "UP" || app.HealthDetail != "" {
		t.Errorf("health = %q, detail = %q", app.Health, app.HealthDetail)
	}
	if len(app.HealthComponents) != 2 || app.HealthComponents[0] != (NameStatus{"solace", "UP"}) {
		t.Errorf("health components = %+v", app.HealthComponents)
	}
	if app.Uptime != "3d 4h 31m" || app.Version != "2.14.1" || app.Java != "openjdk 17.0.9" {
		t.Errorf("enrichment = %q / %q / %q", app.Uptime, app.Version, app.Java)
	}
	if app.Config != "/app/external/spring/config/application.yml" {
		t.Errorf("config = %q", app.Config)
	}
	// The script hands over raw bytes on purpose (busybox cannot do the
	// arithmetic safely); the rendering happens out here.
	if app.Heap == nil || app.Heap.Used != "412Mi" || app.Heap.Limit != "1Gi" || app.Heap.Percent != "40%" {
		t.Errorf("heap = %+v", app.Heap)
	}
	if len(app.Workflows) != 2 || app.Workflows[0] != (Workflow{"0", "running"}) || app.Workflows[1] != (Workflow{"10", "stopped"}) {
		t.Errorf("workflows = %+v", app.Workflows)
	}
	if len(app.Notes) != 1 || !strings.Contains(app.Notes[0], "exposure list") {
		t.Errorf("notes = %+v", app.Notes)
	}
}

func TestParseApplicationKeepsWhatItDoesNotRecognise(t *testing.T) {
	// An instance carrying a newer script must still report everything it
	// printed: an unknown line is kept as a note, never dropped.
	app := ParseApplication("leader-election state: standby\nsomething-new: 42\n")
	if app.LeaderElectionState != "standby" {
		t.Errorf("state = %q", app.LeaderElectionState)
	}
	if len(app.Notes) != 1 || app.Notes[0] != "something-new: 42" {
		t.Errorf("notes = %+v, want the unrecognised line kept verbatim", app.Notes)
	}
}

func TestParseApplicationHealthDetailAndBareHeap(t *testing.T) {
	app := ParseApplication("health: DOWN\nhealth-detail: {\"status\":\"DOWN\"}\nheap: 432013312\n")
	if app.Health != "DOWN" {
		t.Errorf("health = %q", app.Health)
	}
	// health-detail must not be swallowed by the health prefix that starts the
	// same way.
	if app.HealthDetail != `{"status":"DOWN"}` {
		t.Errorf("detail = %q", app.HealthDetail)
	}
	// An unbounded heap reports no maximum, so there is no percentage to state.
	if app.Heap == nil || app.Heap.Used != "412Mi" || app.Heap.Limit != "" || app.Heap.Percent != "" {
		t.Errorf("heap = %+v", app.Heap)
	}
}

func TestParseApplicationCRLFAndBlankLines(t *testing.T) {
	app := ParseApplication("leader-election mode: standalone\r\n\r\nworkflows:\r\n   0: running\r\n")
	if app.LeaderElectionMode != "standalone" {
		t.Errorf("mode = %q", app.LeaderElectionMode)
	}
	if len(app.Workflows) != 1 || app.Workflows[0].State != "running" {
		t.Errorf("workflows = %+v", app.Workflows)
	}
}

func TestParseApplicationIndentedLineWithNoBlockIsANote(t *testing.T) {
	app := ParseApplication("  stray: value\n")
	if len(app.Notes) != 1 || app.Notes[0] != "stray: value" {
		t.Errorf("notes = %+v", app.Notes)
	}
	if len(app.Workflows) != 0 {
		t.Errorf("nothing should be read as a workflow: %+v", app.Workflows)
	}
}
