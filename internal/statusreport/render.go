package statusreport

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/validate"
)

// View selects which halves of the report are rendered: the engine's view of
// the containers, the connector's view of itself, or both. It mirrors the
// status verb's target word, so the CLI passes the operator's choice straight
// through rather than deciding here what a word means.
type View int

const (
	// ViewContainer renders the engine-level tables only.
	ViewContainer View = iota
	// ViewApplication renders the per-instance application blocks only.
	ViewApplication
	// ViewAll renders the container tables, then the application blocks.
	ViewAll
)

// Render returns the report's lines, without a trailing blank. The caller
// prints them; nothing here writes to a stream, so the same rendering serves
// stdout, a watch redraw buffer, and a test.
func Render(r Report, view View, level Level) []string {
	var out []string
	if view != ViewApplication {
		out = append(out, renderContainers(r, level)...)
	}
	if view != ViewContainer {
		app := renderApplications(r, level)
		if len(out) > 0 && len(app) > 0 {
			out = append(out, "")
		}
		out = append(out, app...)
	}
	// Run-level notes last: they are about the collection, not about any one
	// instance, and an operator reads them after the facts they qualify.
	for _, n := range r.Notes {
		out = append(out, noteLine(n))
	}
	return out
}

// noteLine gives every note the same "status:" lead-in the in-container script
// already uses for its own, so a problem found out here and a problem found in
// there read the same way.
func noteLine(n string) string {
	if strings.HasPrefix(n, prefixNote) {
		return n
	}
	return prefixNote + " " + n
}

// ---- container view ----------------------------------------------------------

// Column headings. Kubernetes reports readiness (its own probe verdict) where
// docker and podman report a healthcheck verdict, so the two platforms differ
// by one column rather than sharing a column that means different things.
const (
	colPodOrContainer = "NAME"
	colState          = "STATE"
	colReady          = "READY"
	colHealth         = "HEALTH"
	colRestarts       = "RESTARTS"
	colAge            = "AGE"
	colNode           = "NODE"
	colImage          = "IMAGE"
)

func renderContainers(r Report, level Level) []string {
	if len(r.Instances) == 0 {
		return nil
	}
	kube := r.Platform == validate.PlatformKubernetes
	head := []string{colPodOrContainer, colState}
	if kube {
		head = append(head, colReady)
	} else {
		head = append(head, colHealth)
	}
	head = append(head, colRestarts, colAge)
	if level == LevelDetails && kube {
		head = append(head, colNode)
	}
	head = append(head, colImage)

	// The namespace has to appear somewhere: in the banner when every instance
	// shares one, or in a column of its own when they differ, which is what a
	// cluster-wide --all produces.
	sharedNS, multiNS := namespaceScope(r)
	t := NewTable(head...)
	if multiNS {
		t = NewTable(append([]string{"NAMESPACE"}, head...)...)
	}
	for _, inst := range r.Instances {
		c := inst.Container
		if c == nil {
			// The engine half was not collected for this instance (a query that
			// failed); its application block still reports, and the note above says
			// why, so the row is left out rather than filled with dashes.
			continue
		}
		row := make([]string, 0, len(head)+1)
		if multiNS {
			row = append(row, inst.Namespace)
		}
		row = append(row, inst.Name, stateCell(c))
		if kube {
			row = append(row, c.Ready)
		} else {
			row = append(row, c.Health)
		}
		row = append(row, strconv.Itoa(c.Restarts), c.Age)
		if level == LevelDetails && kube {
			row = append(row, c.Node)
		}
		row = append(row, c.Image)
		t.Row(row...)
	}
	if t.Empty() {
		return nil
	}

	out := []string{Banner(r.Platform, sharedNS, r.Group)}
	out = append(out, t.Lines()...)
	if w := renderWorkload(r.Workload); len(w) > 0 {
		out = append(out, "")
		out = append(out, w...)
	}
	if level == LevelDetails {
		for _, inst := range r.Instances {
			if d := renderContainerDetail(inst); len(d) > 0 {
				out = append(out, "")
				out = append(out, d...)
			}
		}
	}
	return out
}

// namespaceScope decides where the namespace is reported: the one every
// instance shares, or per row when they differ. The instances are the source
// rather than the run's resolved namespace, so an explicit --pod against a pod
// whose namespace was never configured still says where that pod is.
func namespaceScope(r Report) (shared string, perRow bool) {
	seen := ""
	for _, i := range r.Instances {
		if i.Namespace == "" {
			continue
		}
		if seen == "" {
			seen = i.Namespace
			continue
		}
		if seen != i.Namespace {
			return "", true
		}
	}
	if seen == "" {
		return r.Namespace, false
	}
	return seen, false
}

// stateCell is the state column: the normalised state, qualified by the
// engine's own reason when it has one, or by the exit code when that is all
// there is. The qualifier is what turns "waiting" into something actionable.
func stateCell(c *Container) string {
	switch {
	case c.Reason != "":
		return c.State + " (" + c.Reason + ")"
	case c.ExitCode != nil && c.State != StateRunning:
		return c.State + " (" + ExitCodeText(c.ExitCode) + ")"
	default:
		return c.State
	}
}

func renderWorkload(w *Workload) []string {
	if w == nil {
		return nil
	}
	var b kvBuilder
	if w.Deployment != "" {
		b.add("deployment", fmt.Sprintf("%s  %d/%d ready, %d up-to-date, %d available",
			w.Deployment, w.Ready, w.Desired, w.UpToDate, w.Available))
	}
	if w.Service != "" {
		svc := w.Service
		if len(w.ServicePorts) > 0 {
			svc += "  " + strings.Join(w.ServicePorts, ", ")
		}
		b.add("service", svc)
	}
	return KV(0, b.pairs)
}

// renderContainerDetail is the per-instance block the details level adds under
// the table: the facts that do not fit a column, each dropped when it was not
// collected. An instance with nothing extra to say prints no block at all.
func renderContainerDetail(inst Instance) []string {
	c := inst.Container
	if c == nil {
		return nil
	}
	var b kvBuilder
	b.add("digest", c.Digest)
	if c.ImageExpected != "" {
		// Only ever set when it differs from what is running, so the wording can
		// state the mismatch outright rather than leaving the reader to compare.
		b.add("image-expected", c.ImageExpected+"  (env.yaml -- this instance is not running the configured image)")
	}
	b.add("started", c.StartedAt)
	if c.RestartSource != "" {
		b.add("restarts", strconv.Itoa(c.Restarts)+" (from "+c.RestartSource+")")
	}
	b.add("cpu", resourceLine(c.CPU))
	b.add("memory", resourceLine(c.Memory))
	lines := KV(2, b.pairs)
	if len(c.Components) > 0 {
		ct := NewTable("KIND", "NAME", "STATUS", "MOUNT")
		for _, comp := range c.Components {
			ct.Row(comp.Kind, comp.Name, comp.Status, comp.Detail)
		}
		lines = append(lines, "  components:")
		for _, l := range ct.Lines() {
			lines = append(lines, "    "+l)
		}
	}
	if len(lines) == 0 {
		return nil
	}
	return append([]string{inst.Name}, lines...)
}

// ---- application view --------------------------------------------------------

func renderApplications(r Report, level Level) []string {
	var out []string
	for _, inst := range r.Instances {
		if inst.Application == nil && inst.Error == "" {
			continue
		}
		if len(out) > 0 {
			out = append(out, "")
		}
		out = append(out, InstanceBanner(r.Platform, inst.Namespace, groupOf(r, inst), inst.Name))
		out = append(out, renderApplication(inst, level)...)
	}
	return out
}

// groupOf is the name between the namespace and the instance in a banner: the
// instance's own group when it has one (docker reads the compose project back
// off the container itself), else the run's (the kubernetes deployment every
// discovered pod belongs to).
func groupOf(r Report, inst Instance) string {
	if inst.Group != "" {
		return inst.Group
	}
	return r.Group
}

func renderApplication(inst Instance, level Level) []string {
	var lines []string
	app := inst.Application
	if app != nil {
		var b kvBuilder
		b.add("leader-election mode", app.LeaderElectionMode)
		b.add("leader-election state", app.LeaderElectionState)
		b.add("health", app.Health)
		b.add("health-detail", app.HealthDetail)
		if level == LevelDetails {
			b.add("uptime", app.Uptime)
			b.add("version", app.Version)
			b.add("java", app.Java)
			b.add("config", app.Config)
			b.add("heap", resourceLine(app.Heap))
		}
		lines = append(lines, KV(2, b.pairs)...)
		if level == LevelDetails && len(app.HealthComponents) > 0 {
			lines = append(lines, "  health components:")
			pairs := make([][2]string, 0, len(app.HealthComponents))
			for _, hc := range app.HealthComponents {
				pairs = append(pairs, [2]string{hc.Name, hc.Status})
			}
			lines = append(lines, KV(4, pairs)...)
		}
		lines = append(lines, renderWorkflows(app.Workflows)...)
		for _, n := range app.Notes {
			lines = append(lines, "  "+noteLine(n))
		}
	}
	if inst.Error != "" {
		// The failure is a body line under this instance's own banner, with the
		// container facts above it, rather than a bare line on stderr with nothing
		// to explain it.
		lines = append(lines, "  "+noteLine(inst.Error))
	}
	return lines
}

// renderWorkflows prints one row per workflow, ids right-aligned so the colons
// line up and 0 stacks under 10 instead of stepping right -- the same rule the
// in-container script follows when it is run by hand.
func renderWorkflows(wfs []Workflow) []string {
	if len(wfs) == 0 {
		return nil
	}
	width := 0
	for _, w := range wfs {
		if n := len(w.ID); n > width {
			width = n
		}
	}
	out := []string{"  workflows:"}
	for _, w := range wfs {
		out = append(out, "    "+strings.Repeat(" ", width-len(w.ID))+w.ID+": "+w.State)
	}
	return out
}

// ---- json --------------------------------------------------------------------

// JSON renders the report as the --output json document: the same model the
// tables are built from, so the two can never disagree. Indented and newline
// terminated, because it is read by people as often as by programs.
//
// Notes and warnings never travel in here beyond Report.Notes -- everything
// else the run has to say goes to stderr, so stdout stays a single parseable
// document.
func JSON(r Report) (string, error) {
	r.SchemaVersion = SchemaVersion
	if r.Instances == nil {
		// An empty run is an empty list, not null: a consumer can iterate it
		// without a nil check.
		r.Instances = []Instance{}
	}
	b, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return "", fmt.Errorf("rendering the status report as json: %w", err)
	}
	return string(b) + "\n", nil
}
