package statusreport

import (
	"strings"
	"testing"
	"time"
)

// now is the fixed clock every age assertion measures against, so a test
// result can never depend on when the suite runs.
var now = time.Date(2026, 8, 21, 12, 0, 0, 0, time.UTC)

func ts(t time.Time) string { return t.Format(time.RFC3339) }

func TestAge(t *testing.T) {
	cases := []struct {
		name  string
		start string
		want  string
	}{
		{"seconds", ts(now.Add(-45 * time.Second)), "45s"},
		{"minutes", ts(now.Add(-20 * time.Minute)), "20m"},
		{"hours and minutes", ts(now.Add(-2*time.Hour - 13*time.Minute)), "2h13m"},
		{"days and hours", ts(now.Add(-76 * time.Hour)), "3d4h"},
		{"past a week, days alone", ts(now.Add(-30 * 24 * time.Hour)), "30d"},
		{"fractional seconds parse", now.Add(-time.Hour).Format(time.RFC3339Nano), "1h0m"},
		// A container that never started reports docker's zero time; an empty or
		// unparseable stamp is not an age of zero, it is no age at all.
		{"docker zero time", "0001-01-01T00:00:00Z", ""},
		{"empty", "", ""},
		{"unparseable", "not-a-time", ""},
		// A host clock behind the engine's must not render a negative age.
		{"clock skew ahead", ts(now.Add(30 * time.Second)), "0s"},
	}
	for _, c := range cases {
		if got := Age(c.start, now); got != c.want {
			t.Errorf("%s: Age(%q) = %q, want %q", c.name, c.start, got, c.want)
		}
	}
}

func TestParseQuantity(t *testing.T) {
	cases := []struct {
		in   string
		want float64
		ok   bool
	}{
		{"1", 1, true},
		{"120m", 0.12, true},       // kubernetes millicores
		{"512Mi", 536870912, true}, // binary suffix
		{"1Gi", 1073741824, true},
		{"512MiB", 536870912, true}, // the engines' spelling of the same unit
		{"1.5GiB", 1610612736, true},
		{"1kB", 1000, true}, // decimal, and never read as the binary Ki
		{"2M", 2e6, true},
		{"432013312", 432013312, true},
		// Micrometer hands a large heap to Jackson, which may serialise it in
		// scientific notation; the script passes the value through verbatim, so
		// this is the form that actually arrives.
		{"4.32013312E8", 432013312, true},
		{"", 0, false},
		{"n/a", 0, false},
		{"12x", 0, false},
	}
	for _, c := range cases {
		got, ok := ParseQuantity(c.in)
		if ok != c.ok {
			t.Errorf("ParseQuantity(%q) ok = %v, want %v", c.in, ok, c.ok)
			continue
		}
		if ok && got != c.want {
			t.Errorf("ParseQuantity(%q) = %v, want %v", c.in, got, c.want)
		}
	}
}

func TestPercent(t *testing.T) {
	cases := []struct {
		used, limit, want string
	}{
		{"512Mi", "1Gi", "50%"},
		{"120m", "1", "12%"},
		{"432013312", "1073741824", "40%"},
		// No ceiling, an unreadable side, or a zero limit (docker's way of saying
		// unlimited) all mean there is no percentage to state.
		{"512Mi", "", ""},
		{"", "1Gi", ""},
		{"512Mi", "0", ""},
		{"n/a", "1Gi", ""},
	}
	for _, c := range cases {
		if got := Percent(c.used, c.limit); got != c.want {
			t.Errorf("Percent(%q, %q) = %q, want %q", c.used, c.limit, got, c.want)
		}
	}
}

func TestBytesAndCores(t *testing.T) {
	byteCases := []struct {
		in   int64
		want string
	}{
		{0, ""},  // docker reports no memory limit as 0
		{-1, ""}, // a JVM with no maximum heap
		{1073741824, "1Gi"},
		{536870912, "512Mi"},
		{1610612736, "1.5Gi"},
		{2048, "2Ki"},
		{512, "512"},
	}
	for _, c := range byteCases {
		if got := Bytes(c.in); got != c.want {
			t.Errorf("Bytes(%d) = %q, want %q", c.in, got, c.want)
		}
	}
	coreCases := []struct {
		in   int64
		want string
	}{
		{0, ""},
		{1000000000, "1"},
		{2500000000, "2.50"},
	}
	for _, c := range coreCases {
		if got := Cores(c.in); got != c.want {
			t.Errorf("Cores(%d) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestBanners(t *testing.T) {
	// A name that is not set is dropped, so a separator always sits between two
	// real names -- and the instance form is unchanged from what this verb
	// printed before the container view existed.
	cases := []struct {
		name string
		got  string
		want string
	}{
		{"section, all names", Banner("kubernetes", "prod", "solmq-connector"), "== kubernetes  prod / solmq-connector =="},
		{"section, no names", Banner("docker"), "== docker =="},
		{"section, gap in the middle", Banner("docker", "", "solmq-connector"), "== docker  solmq-connector =="},
		{"instance", InstanceBanner("kubernetes", "prod", "solmq-connector", "pod-a"), "=== kubernetes  prod / solmq-connector / pod-a ==="},
		{"instance, podman has only the container", InstanceBanner("podman", "", "", "solmq"), "=== podman  solmq ==="},
	}
	for _, c := range cases {
		if c.got != c.want {
			t.Errorf("%s: got %q, want %q", c.name, c.got, c.want)
		}
	}
}

func TestTableAlignsColumnsAndNeverPadsTheLast(t *testing.T) {
	tb := NewTable("NAME", "STATE", "IMAGE")
	tb.Row("pod-a", "running", "solace/x:1")
	tb.Row("much-longer-pod-b", "exited", "solace/x:2")
	tb.Row("pod-c") // a row that stops early is padded, not a panic
	got := tb.Lines()
	want := []string{
		"NAME               STATE    IMAGE",
		"pod-a              running  solace/x:1",
		"much-longer-pod-b  exited   solace/x:2",
		"pod-c              -        -",
	}
	if len(got) != len(want) {
		t.Fatalf("got %d lines, want %d:\n%s", len(got), len(want), strings.Join(got, "\n"))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("line %d = %q, want %q", i, got[i], want[i])
		}
		if strings.HasSuffix(got[i], " ") {
			t.Errorf("line %d has trailing whitespace: %q", i, got[i])
		}
	}
}

func TestTableEmptyReportsNoRows(t *testing.T) {
	tb := NewTable("NAME")
	if !tb.Empty() {
		t.Error("a table with no rows should report Empty")
	}
	tb.Row("pod-a")
	if tb.Empty() {
		t.Error("a table with a row should not report Empty")
	}
}

func TestKVAlignsValuesOnTheWidestKey(t *testing.T) {
	got := KV(2, [][2]string{
		{"health", "UP"},
		{"leader-election state", "active"},
	})
	want := []string{
		"  health:                UP",
		"  leader-election state: active",
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("line %d = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestKVBuilderDropsEmptyValues(t *testing.T) {
	// The report's noise rule in one place: a fact that was not collected prints
	// nothing rather than a line saying it is unknown.
	var b kvBuilder
	b.add("version", "2.14.1")
	b.add("java", "")
	b.add("config", "/app/external/spring/config/application.yml")
	if len(b.pairs) != 2 {
		t.Fatalf("got %d pairs, want 2: %v", len(b.pairs), b.pairs)
	}
	if b.pairs[1][0] != "config" {
		t.Errorf("the empty pair should be dropped, leaving config second: %v", b.pairs)
	}
}

func TestResourceLineDropsWhatIsMissing(t *testing.T) {
	cases := []struct {
		name string
		in   *Resource
		want string
	}{
		{"nil", nil, ""},
		{"no usage collected", &Resource{Limit: "1Gi"}, ""},
		{"usage alone", &Resource{Used: "512Mi"}, "512Mi"},
		{"usage and limit", &Resource{Used: "512Mi", Limit: "1Gi"}, "512Mi of 1Gi"},
		{"all three", &Resource{Used: "512Mi", Limit: "1Gi", Percent: "50%"}, "512Mi of 1Gi (50%)"},
		// docker's stats already carry both sides in one string.
		{"engine-formatted usage", &Resource{Used: "512MiB / 15.6GiB", Percent: "3.20%"}, "512MiB / 15.6GiB (3.20%)"},
	}
	for _, c := range cases {
		if got := resourceLine(c.in); got != c.want {
			t.Errorf("%s: got %q, want %q", c.name, got, c.want)
		}
	}
}

func TestImageMismatch(t *testing.T) {
	const want = "solace/solace-pubsub-connector-ibmmq:2.14.1"
	cases := []struct {
		name            string
		running, digest string
		expect          string
		mismatch        bool
	}{
		{"identical", want, "", want, false},
		// kubernetes normalises a reference to its fully qualified form, so the
		// same image must not read as a mismatch on kubernetes alone.
		{"kubernetes-normalised", "docker.io/" + want, "", want, false},
		{"library namespace", "docker.io/library/busybox:1.37", "", "busybox:1.37", false},
		{"implied latest on one side", "solace/x", "", "solace/x:latest", false},
		{"different tag -- the failed rollout", "solace/solace-pubsub-connector-ibmmq:2.13.0", "", want, true},
		{"different repository", "other/connector:2.14.1", "", want, true},
		// Nothing collected on either side is not a claim of a mismatch.
		{"no running image", "", "", want, false},
		{"no expectation", want, "", "", false},
		// Pinned by digest: the tag cannot answer, so the digest must.
		{"digest pin matches", "solace/x@sha256:abc", "sha256:abc", "solace/x@sha256:abc", false},
		{"digest pin differs", "solace/x@sha256:def", "sha256:def", "solace/x@sha256:abc", true},
		{"digest pin, engine reported none", "solace/x", "", "solace/x@sha256:abc", false},
	}
	for _, c := range cases {
		if got := ImageMismatch(c.running, c.digest, c.expect); got != c.mismatch {
			t.Errorf("%s: ImageMismatch(%q, %q, %q) = %v, want %v", c.name, c.running, c.digest, c.expect, got, c.mismatch)
		}
	}
}

func TestSortInstancesGroupsByNamespaceThenName(t *testing.T) {
	in := []Instance{
		{Name: "pod-b", Namespace: "prod"},
		{Name: "pod-a", Namespace: "prod"},
		{Name: "pod-z", Namespace: "dev"},
	}
	SortInstances(in)
	got := []string{in[0].Namespace + "/" + in[0].Name, in[1].Namespace + "/" + in[1].Name, in[2].Namespace + "/" + in[2].Name}
	want := []string{"dev/pod-z", "prod/pod-a", "prod/pod-b"}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("position %d = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestExitCodeText(t *testing.T) {
	if got := ExitCodeText(nil); got != "" {
		t.Errorf("a container that has not terminated has no exit code, got %q", got)
	}
	code := 137
	if got := ExitCodeText(&code); got != "exit code 137" {
		t.Errorf("got %q", got)
	}
}
