package statusreport

import (
	"encoding/json"
	"strings"
	"testing"
)

// sample builds a two-instance kubernetes report: one healthy, one crash
// looping. Every rendering test works from this, so a change in one view's
// shape is visible against the same facts in the other.
func sample() Report {
	code := 137
	return Report{
		Platform:  "kubernetes",
		Namespace: "prod",
		Group:     "solmq-connector",
		Workload: &Workload{
			Deployment: "solmq-connector", Ready: 1, Desired: 2, UpToDate: 1, Available: 2,
			Service: "solmq-connector", ServicePorts: []string{"8090/TCP"},
		},
		Instances: []Instance{
			{
				Name: "pod-a", Namespace: "prod",
				Container: &Container{
					State: StateRunning, Ready: "yes", Health: NotApplicable,
					Age: "3d7h", Image: "solace/x:2.14.1", Digest: "sha256:9f2a1c",
					Node: "node-a", StartedAt: "2026-08-18T04:12:07Z",
					CPU:    &Resource{Used: "120m", Limit: "1", Percent: "12%"},
					Memory: &Resource{Used: "512Mi", Limit: "1Gi", Percent: "50%"},
					Components: []Component{
						{Kind: "configmap", Name: "solmq-connector-config", Status: "present", Detail: "/app/external/spring/config/application.yml"},
						{Kind: "secret", Name: "solmq-connector-stores", Status: "MISSING"},
					},
				},
				Application: &Application{
					LeaderElectionMode: "active_standby", LeaderElectionState: "active",
					Health: "UP", Uptime: "3d 4h 31m", Version: "2.14.1",
					Java: "openjdk 17.0.9", Config: "/app/external/spring/config/application.yml",
					Heap:             &Resource{Used: "412Mi", Limit: "1Gi", Percent: "40%"},
					HealthComponents: []NameStatus{{"solace", "UP"}, {"ibmmq", "UP"}},
					Workflows:        []Workflow{{"0", "running"}, {"10", "stopped"}},
				},
			},
			{
				Name: "pod-b", Namespace: "prod",
				Container: &Container{
					State: StateRestarting, Reason: "CrashLoopBackOff", Ready: "no", Health: NotApplicable,
					Restarts: 7, ExitCode: &code, Age: "3d7h",
					Image: "solace/x:2.13.0", ImageExpected: "solace/x:2.14.1",
				},
				Error: "could not run the status script: command terminated with exit code 126",
			},
		},
	}
}

func render(t *testing.T, r Report, v View, l Level) string {
	t.Helper()
	return strings.Join(Render(r, v, l), "\n") + "\n"
}

func TestRenderContainerViewBasic(t *testing.T) {
	got := render(t, sample(), ViewContainer, LevelBasic)
	// The table pads to its widest cell, so compare column by column rather
	// than guessing the padding: assert the parts that carry meaning.
	for _, must := range []string{
		"== kubernetes  prod / solmq-connector ==",
		"NAME ", "STATE", "READY", "RESTARTS", "AGE", "IMAGE",
		"pod-a", "running", "pod-b", "restarting (CrashLoopBackOff)",
		"deployment: solmq-connector  1/2 ready, 1 up-to-date, 2 available",
		"service:    solmq-connector  8090/TCP",
	} {
		if !strings.Contains(got, must) {
			t.Errorf("container view is missing %q, got:\n%s", must, got)
		}
	}
	// The basic level has no NODE column and no per-instance detail block.
	if strings.Contains(got, "NODE") || strings.Contains(got, "digest") {
		t.Errorf("basic level must not carry the details columns, got:\n%s", got)
	}
	// A kubernetes table reports readiness, never a healthcheck verdict.
	if strings.Contains(got, "HEALTH") {
		t.Errorf("kubernetes reports READY, not HEALTH, got:\n%s", got)
	}
}

func TestRenderContainerViewDetails(t *testing.T) {
	got := render(t, sample(), ViewContainer, LevelDetails)
	for _, must := range []string{
		"NODE",
		"node-a",
		"digest:", "sha256:9f2a1c",
		"cpu:", "120m of 1 (12%)",
		"memory:", "512Mi of 1Gi (50%)",
		"components:",
		"configmap", "solmq-connector-config", "present",
		"secret", "solmq-connector-stores", "MISSING",
		// Only ever rendered on a mismatch, so its presence is the finding.
		"image-expected:", "solace/x:2.14.1",
		"not running the configured image",
	} {
		if !strings.Contains(got, must) {
			t.Errorf("details view is missing %q, got:\n%s", must, got)
		}
	}
	// pod-b collected no resources, so it gets no resource lines at all rather
	// than lines saying unknown.
	detail := got[strings.Index(got, "pod-b\n"):]
	if strings.Contains(detail, "cpu:") {
		t.Errorf("an instance with no sample must not carry a cpu line:\n%s", detail)
	}
}

func TestRenderContainerViewDockerUsesHealthColumn(t *testing.T) {
	r := Report{
		Platform: "docker",
		Instances: []Instance{{
			Name: "solmq-connector", Group: "eg",
			Container: &Container{State: StateRunning, Health: "healthy", Ready: NotApplicable, Age: "3d7h", Image: "solace/x:2.14.1"},
		}},
	}
	got := render(t, r, ViewContainer, LevelBasic)
	if !strings.Contains(got, "HEALTH") {
		t.Errorf("docker reports the engine healthcheck verdict, got:\n%s", got)
	}
	if strings.Contains(got, "READY") {
		t.Errorf("docker has no readiness concept, got:\n%s", got)
	}
	if strings.Contains(got, "NODE") {
		t.Errorf("only kubernetes has nodes, got:\n%s", got)
	}
}

func TestRenderContainerViewAllNamespacesLeadsWithNamespace(t *testing.T) {
	// Under --all instances come from anywhere, so the run has no one namespace
	// to put in the banner and each row carries its own.
	r := Report{
		Platform: "kubernetes",
		Instances: []Instance{
			{Name: "pod-a", Namespace: "prod", Container: &Container{State: StateRunning, Age: "1h0m", Image: "solace/x:1"}},
			{Name: "pod-b", Namespace: "dev", Container: &Container{State: StateRunning, Age: "2h0m", Image: "solace/x:1"}},
		},
	}
	got := render(t, r, ViewContainer, LevelBasic)
	if !strings.Contains(got, "NAMESPACE") {
		t.Errorf("a cluster-wide report needs the namespace column, got:\n%s", got)
	}
	if !strings.Contains(got, "prod") || !strings.Contains(got, "dev") {
		t.Errorf("both namespaces should appear, got:\n%s", got)
	}
}

func TestRenderApplicationViewBasicAndDetails(t *testing.T) {
	basic := render(t, sample(), ViewApplication, LevelBasic)
	// The banner is unchanged from what this verb printed before the container
	// view existed.
	if !strings.Contains(basic, "=== kubernetes  prod / solmq-connector / pod-a ===") {
		t.Errorf("instance banner missing, got:\n%s", basic)
	}
	for _, must := range []string{
		"leader-election mode:  active_standby",
		"leader-election state: active",
		"health:                UP",
		"  workflows:",
		"     0: running",
		"    10: stopped",
	} {
		if !strings.Contains(basic, must) {
			t.Errorf("basic application block is missing %q, got:\n%s", must, basic)
		}
	}
	// Enrichment is opt-in.
	for _, mustNot := range []string{"uptime:", "version:", "java:", "config:", "heap:", "health components:"} {
		if strings.Contains(basic, mustNot) {
			t.Errorf("basic level must not carry %q, got:\n%s", mustNot, basic)
		}
	}
	// The container view is not in this rendering at all.
	if strings.Contains(basic, "NAME ") || strings.Contains(basic, "== kubernetes  prod / solmq-connector ==") {
		t.Errorf("the application view must not print the container table, got:\n%s", basic)
	}

	details := render(t, sample(), ViewApplication, LevelDetails)
	for _, must := range []string{
		"uptime:                3d 4h 31m",
		"version:               2.14.1",
		"java:                  openjdk 17.0.9",
		"config:                /app/external/spring/config/application.yml",
		"heap:                  412Mi of 1Gi (40%)",
		"  health components:",
		"    solace: UP",
		"    ibmmq:  UP",
	} {
		if !strings.Contains(details, must) {
			t.Errorf("details application block is missing %q, got:\n%s", must, details)
		}
	}
}

func TestRenderFailedInstanceKeepsItsBlock(t *testing.T) {
	// The whole point of the rework: an instance whose script could not run
	// still gets a block, with the failure as a body line under the container
	// facts that explain it.
	got := render(t, sample(), ViewAll, LevelBasic)
	if !strings.Contains(got, "=== kubernetes  prod / solmq-connector / pod-b ===") {
		t.Errorf("a failed instance must still get a banner, got:\n%s", got)
	}
	if !strings.Contains(got, "  status: could not run the status script: command terminated with exit code 126") {
		t.Errorf("the failure should read as a status note, got:\n%s", got)
	}
	// And the container table above it says why.
	if strings.Index(got, "restarting (CrashLoopBackOff)") > strings.Index(got, "could not run the status script") {
		t.Error("the container table should come before the application blocks")
	}
}

func TestRenderNotesAndScriptNotesShareOneIdiom(t *testing.T) {
	r := sample()
	r.Notes = []string{"no resource usage: metrics API not available"}
	r.Instances[0].Application.Notes = []string{"status: workflows is not exposed"}
	got := render(t, r, ViewAll, LevelBasic)
	// A note the CLI made and a note the in-container script made read the same.
	if !strings.Contains(got, "  status: workflows is not exposed") {
		t.Errorf("the script's own note should keep its prefix, got:\n%s", got)
	}
	if !strings.Contains(got, "status: no resource usage: metrics API not available") {
		t.Errorf("a run-level note should be given the same prefix, got:\n%s", got)
	}
	// Run-level notes are about the collection, so they come last.
	if strings.Index(got, "no resource usage") < strings.Index(got, "leader-election mode") {
		t.Error("run-level notes belong after the facts they qualify")
	}
}

func TestRenderEmptyReport(t *testing.T) {
	if got := Render(Report{Platform: "docker"}, ViewAll, LevelBasic); len(got) != 0 {
		t.Errorf("a report with no instances renders nothing, got %q", got)
	}
}

func TestRenderInstanceWithNoContainerFactsStillReportsItsApplication(t *testing.T) {
	// An engine query that failed leaves the container half nil; the
	// application half is still worth printing, and the row is left out rather
	// than filled with dashes.
	r := Report{
		Platform:  "kubernetes",
		Namespace: "prod",
		Instances: []Instance{{Name: "pod-a", Namespace: "prod", Application: &Application{LeaderElectionState: "active"}}},
		Notes:     []string{"could not read the pod list"},
	}
	got := render(t, r, ViewAll, LevelBasic)
	if strings.Contains(got, "NAME ") {
		t.Errorf("no table should be printed when nothing was collected, got:\n%s", got)
	}
	if !strings.Contains(got, "leader-election state: active") {
		t.Errorf("the application half should still report, got:\n%s", got)
	}
}

func TestJSONIsTheSameModelTheTablesRender(t *testing.T) {
	doc, err := JSON(sample())
	if err != nil {
		t.Fatalf("JSON: %v", err)
	}
	var round Report
	if err := json.Unmarshal([]byte(doc), &round); err != nil {
		t.Fatalf("the document must parse back: %v", err)
	}
	if round.SchemaVersion != SchemaVersion {
		t.Errorf("schemaVersion = %d, want %d", round.SchemaVersion, SchemaVersion)
	}
	if len(round.Instances) != 2 {
		t.Fatalf("instances = %d", len(round.Instances))
	}
	if round.Instances[0].Application.Workflows[1].ID != "10" {
		t.Errorf("workflows did not survive the round trip: %+v", round.Instances[0].Application)
	}
	if round.Instances[1].Error == "" {
		t.Error("a failed instance's reason belongs in the document too")
	}
	if !strings.HasSuffix(doc, "\n") {
		t.Error("the document should end in a newline")
	}
	// Field names are a compatibility contract: assert the spellings a consumer
	// would key off, so a rename has to bump SchemaVersion deliberately.
	for _, key := range []string{
		`"schemaVersion": 1`,
		`"platform": "kubernetes"`,
		`"namespace": "prod"`,
		`"workload"`,
		`"upToDate"`,
		`"servicePorts"`,
		`"instances"`,
		`"container"`,
		`"application"`,
		`"leaderElectionState"`,
		`"healthComponents"`,
		`"workflows"`,
		`"exitCode"`,
		`"imageExpected"`,
		`"components"`,
		`"error"`,
	} {
		if !strings.Contains(doc, key) {
			t.Errorf("json document is missing %s, got:\n%s", key, doc)
		}
	}
	// A field with nothing in it is left out entirely rather than emitted as
	// null or an empty string, so a consumer can test for presence.
	if strings.Contains(doc, `"restartSource"`) {
		t.Errorf("an unset field should be omitted, got:\n%s", doc)
	}
}

func TestJSONEmptyRunIsAnEmptyList(t *testing.T) {
	doc, err := JSON(Report{Platform: "docker"})
	if err != nil {
		t.Fatal(err)
	}
	// A consumer must be able to iterate instances without a nil check.
	if !strings.Contains(doc, `"instances": []`) {
		t.Errorf("want an empty list, got:\n%s", doc)
	}
}
