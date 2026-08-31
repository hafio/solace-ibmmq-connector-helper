// Package statusreport is the CLI-side half of the status verb's reporting: the
// typed model every status view is built into, and the two renderings of it --
// the human tables/blocks and the --output json document.
//
// It is a pure package: no os/exec, no filesystem, no network, no globals. The
// caller (cmd/solmq-conn-util) runs the read-only engine queries through
// internal/runner and hands the raw output here to be parsed; nothing in this
// package can start a process. That split is what makes the whole report
// testable from captured fixtures, the same way internal/statusscript is
// testable as a pure renderer of the in-container half.
//
// One model, two renderings: every fact reaches the operator through the
// structs below, so the table view and the JSON document can never disagree
// about what was collected. Field names in the JSON document are a
// compatibility contract -- see SchemaVersion.
package statusreport

import (
	"sort"
	"strconv"
	"strings"
	"time"
)

// SchemaVersion is the --output json document's schema version. It is a
// contract with whatever parses that document: a field may be added at the
// same version (a parser ignoring unknown fields keeps working), but renaming
// or removing one, or changing a value's meaning, requires a bump.
const SchemaVersion = 1

// ImageMatch is the substring --all looks for in a container's image reference
// to decide whether it is a connector instance. A substring rather than a
// pattern: the operator's registry host, namespace, and tag are all unknown
// here, and the repository name is the only stable part of the reference.
const ImageMatch = "solace-pubsub-connector-ibmmq"

// Level selects how much of the collected model a rendering shows. The model
// itself is always whatever the caller managed to collect; the level decides
// what is printed, so one collection can be rendered at either level.
type Level int

const (
	// LevelBasic prints the identity and state lines an operator reads first.
	LevelBasic Level = iota
	// LevelDetails adds the enrichment lines: node, resources, digest,
	// components, and the application's version/java/config/heap block.
	LevelDetails
)

// Container states, normalised across the three engines so one column reads
// the same everywhere. The engine's own word for *why* travels separately in
// Container.Reason, never folded into the state, so a state stays comparable
// while the reason stays whatever the engine called it.
const (
	StateRunning    = "running"
	StateExited     = "exited"
	StateWaiting    = "waiting"
	StateRestarting = "restarting"
	StatePaused     = "paused"
	StateUnknown    = "unknown"
)

// NotApplicable is what a column shows when the engine has no such concept for
// this target -- a docker container with no HEALTHCHECK, a pod whose node is
// not yet assigned. It is deliberately distinct from an empty value, which
// means "not collected".
const NotApplicable = "n/a"

// Report is one status run: the platform it ran against and one entry per
// instance found, plus the workload summary when the platform has one above
// the instance (a kubernetes Deployment/Service).
type Report struct {
	SchemaVersion int    `json:"schemaVersion"`
	Platform      string `json:"platform"`
	// Namespace/Group locate the whole run: the kubernetes namespace and
	// deployment, or the docker compose project. Both are empty under --all,
	// where instances come from anywhere and each carries its own.
	Namespace string     `json:"namespace,omitempty"`
	Group     string     `json:"group,omitempty"`
	Workload  *Workload  `json:"workload,omitempty"`
	Instances []Instance `json:"instances"`
	// Notes are run-level problems that did not stop the report: a metrics API
	// that is not installed, a workload object that could not be read. They are
	// rendered as "status:" lines, the same idiom the in-container script uses
	// for its own notes.
	Notes []string `json:"notes,omitempty"`
}

// Workload is the object above the instances on kubernetes: the Deployment's
// replica counts and the Service that fronts it. Nil when it was not read --
// under --all, or for an instance this tool never deployed.
type Workload struct {
	Deployment string `json:"deployment,omitempty"`
	Ready      int    `json:"ready"`
	Desired    int    `json:"desired"`
	UpToDate   int    `json:"upToDate"`
	Available  int    `json:"available"`
	Service    string `json:"service,omitempty"`
	// ServicePorts renders as "8090/TCP" entries; a headless or port-less
	// service simply has none.
	ServicePorts []string `json:"servicePorts,omitempty"`
}

// Instance is one connector instance: a kubernetes pod, or a docker/podman
// container. Container carries what the engine knows about it and Application
// what the connector inside it reports; either may be nil when that half was
// not collected (a view that did not ask for it, or a query that failed).
type Instance struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace,omitempty"`
	Group     string `json:"group,omitempty"`

	// ContainerName is the container inside a kubernetes pod this report is
	// about, as connectorIndex picked it. Empty on docker/podman, where the
	// target is the container, and empty for a multi-container pod in which
	// none is the connector -- the case connectorIndex deliberately refuses to
	// guess at. A caller addressing the container by name (kubectl -c) reads it
	// here rather than deciding again.
	ContainerName string `json:"containerName,omitempty"`

	Container   *Container   `json:"container,omitempty"`
	Application *Application `json:"application,omitempty"`

	// Error is why this instance has no Application block: the script could not
	// be probed, installed, or run. It is a body line in the report rather than
	// a bare stderr message, so the container facts that explain it sit right
	// above it.
	Error string `json:"error,omitempty"`
}

// Container is the engine's view of one instance.
type Container struct {
	State  string `json:"state"`
	Reason string `json:"reason,omitempty"`
	// Ready is the kubernetes readiness verdict ("yes"/"no"), NotApplicable on
	// docker/podman, which have no such concept.
	Ready string `json:"ready,omitempty"`
	// Health is the engine's own healthcheck verdict (docker/podman
	// healthy/unhealthy/starting), NotApplicable when the container defines no
	// healthcheck. It is deliberately distinct from the connector's actuator
	// health in Application.Health: this one is the engine's opinion.
	Health string `json:"health,omitempty"`

	Restarts int `json:"restarts"`
	// RestartSource names where Restarts came from when it is not the
	// container's own counter -- "systemd" under podman quadlet, where the unit
	// is recreated rather than restarted so the container's counter reads 0.
	RestartSource string `json:"restartSource,omitempty"`
	// ExitCode is the most recent non-zero exit code known for the container --
	// its own if it has terminated, otherwise its previous termination's, which
	// is the useful one on an instance that is looping or back up. Nil when the
	// container has never exited non-zero: a clean exit adds nothing to State.
	ExitCode *int `json:"exitCode,omitempty"`

	// StartedAt is the engine's own timestamp, kept verbatim; Age is it
	// rendered against the run's clock.
	StartedAt string `json:"startedAt,omitempty"`
	Age       string `json:"age,omitempty"`

	Image  string `json:"image,omitempty"`
	Digest string `json:"digest,omitempty"`
	// ImageExpected is set only when it differs from Image: the reference
	// env.yaml asks for, so a pod still on the old tag after a failed rollout
	// says so in its own block.
	ImageExpected string `json:"imageExpected,omitempty"`

	Node string `json:"node,omitempty"`

	CPU    *Resource `json:"cpu,omitempty"`
	Memory *Resource `json:"memory,omitempty"`

	Components []Component `json:"components,omitempty"`
}

// Resource is one measured resource against whatever ceiling the engine
// reports for it. Every field is a rendered string, not a number: each engine
// reports in its own units (millicores, MiB, a percentage it computed itself)
// and converting them to a common unit here would invent precision the engine
// never gave. Percent is filled only when the engine reported it or both sides
// parsed cleanly.
type Resource struct {
	Used    string `json:"used,omitempty"`
	Limit   string `json:"limit,omitempty"`
	Percent string `json:"percent,omitempty"`
}

// Component is one object the workload references: a secret, configmap, or
// volume claim on kubernetes; a mount, network, or quadlet unit on
// docker/podman. Read from what the workload actually references rather than
// from env.yaml, so it works for an instance this tool never deployed.
type Component struct {
	Kind   string `json:"kind"`
	Name   string `json:"name"`
	Status string `json:"status,omitempty"`
	Detail string `json:"detail,omitempty"`
}

// Application is what the connector inside the container reports, parsed from
// the in-container script's output.
type Application struct {
	LeaderElectionMode  string `json:"leaderElectionMode,omitempty"`
	LeaderElectionState string `json:"leaderElectionState,omitempty"`

	Health       string `json:"health,omitempty"`
	HealthDetail string `json:"healthDetail,omitempty"`
	// HealthComponents is the health document's per-component statuses, which
	// is the app-level answer closest to "is it actually moving messages".
	HealthComponents []NameStatus `json:"healthComponents,omitempty"`

	Workflows []Workflow `json:"workflows,omitempty"`

	Uptime  string    `json:"uptime,omitempty"`
	Version string    `json:"version,omitempty"`
	Java    string    `json:"java,omitempty"`
	Config  string    `json:"config,omitempty"`
	Heap    *Resource `json:"heap,omitempty"`

	// Notes are the script's own "status:" lines plus any line this parser did
	// not recognise. Unknown lines are kept rather than dropped so an instance
	// carrying a newer or older script still reports everything it printed.
	Notes []string `json:"notes,omitempty"`
}

// Workflow is one workflow slot's reported state.
type Workflow struct {
	ID    string `json:"id"`
	State string `json:"state"`
}

// NameStatus is a named component and its status (health components).
type NameStatus struct {
	Name   string `json:"name"`
	Status string `json:"status"`
}

// ---- formatting vocabulary ---------------------------------------------------

// The report has exactly two banner levels, so a reader can tell a section
// from an instance at a glance without counting characters: a section groups
// instances, an instance heads one block. The instance form is unchanged from
// the report this verb printed before the container view existed.
const (
	sectionRule  = "=="
	instanceRule = "==="
)

// Banner renders a section header: the platform, then the names that locate
// the group on it, joined with " / ". An unset name is dropped rather than
// rendered as an empty segment, so a separator always sits between two real
// names.
func Banner(platform string, parts ...string) string {
	return banner(sectionRule, platform, parts...)
}

// InstanceBanner renders one instance's header, the identity line above its
// report block.
func InstanceBanner(platform string, parts ...string) string {
	return banner(instanceRule, platform, parts...)
}

func banner(rule, platform string, parts ...string) string {
	kept := make([]string, 0, len(parts))
	for _, s := range parts {
		if s != "" {
			kept = append(kept, s)
		}
	}
	head := platform
	if len(kept) > 0 {
		head += "  " + strings.Join(kept, " / ")
	}
	return rule + " " + head + " " + rule
}

// Table is the report's one table shape: a header row and body rows, every
// column padded to its widest cell and separated by a two-space gutter. The
// last column is never padded, so a long image reference cannot leave trailing
// whitespace on every line.
type Table struct {
	head []string
	rows [][]string
}

// NewTable starts a table with the given column headings.
func NewTable(head ...string) *Table { return &Table{head: head} }

// Row appends one row. A row shorter than the heading is padded with empty
// cells, so a caller that has nothing for a trailing column can just leave it
// off rather than passing "".
func (t *Table) Row(cells ...string) {
	row := make([]string, len(t.head))
	copy(row, cells)
	t.rows = append(t.rows, row)
}

// Empty reports whether the table has no body rows, so a caller can skip
// printing a heading with nothing under it.
func (t *Table) Empty() bool { return len(t.rows) == 0 }

// Lines renders the table, heading first. An empty cell renders as "-" so a
// column never silently collapses into the gutter and a missing value is
// visibly missing.
func (t *Table) Lines() []string {
	widths := make([]int, len(t.head))
	for i, h := range t.head {
		widths[i] = len(h)
	}
	cell := func(s string) string {
		if s == "" {
			return "-"
		}
		return s
	}
	for _, r := range t.rows {
		for i, c := range r {
			if n := len(cell(c)); n > widths[i] {
				widths[i] = n
			}
		}
	}
	out := make([]string, 0, len(t.rows)+1)
	out = append(out, joinCells(t.head, widths))
	for _, r := range t.rows {
		cells := make([]string, len(r))
		for i, c := range r {
			cells[i] = cell(c)
		}
		out = append(out, joinCells(cells, widths))
	}
	return out
}

func joinCells(cells []string, widths []int) string {
	var b strings.Builder
	for i, c := range cells {
		if i > 0 {
			b.WriteString("  ")
		}
		if i == len(cells)-1 {
			b.WriteString(c) // never pad the last column: no trailing whitespace
			continue
		}
		b.WriteString(c)
		for n := len(c); n < widths[i]; n++ {
			b.WriteByte(' ')
		}
	}
	return b.String()
}

// KV renders aligned "key: value" lines at the given indent, every value
// starting in the one column the widest key decides, so a block of facts reads
// as two columns however long its longest label is. A pair with an empty value
// is dropped by the caller (see kvBuilder), not here: this function renders
// what it is given.
func KV(indent int, pairs [][2]string) []string {
	width := 0
	for _, p := range pairs {
		if n := len(p[0]); n > width {
			width = n
		}
	}
	pad := strings.Repeat(" ", indent)
	out := make([]string, 0, len(pairs))
	for _, p := range pairs {
		key := p[0] + ":"
		for n := len(key); n < width+1; n++ {
			key += " "
		}
		out = append(out, pad+key+" "+p[1])
	}
	return out
}

// kvBuilder collects key/value pairs, skipping any whose value is empty, so a
// caller can list every possible line once and let the unavailable ones fall
// out. That is the report's noise rule in one place: a fact that was not
// collected prints nothing at all rather than a line saying so.
type kvBuilder struct{ pairs [][2]string }

func (b *kvBuilder) add(key, val string) {
	if val == "" {
		return
	}
	b.pairs = append(b.pairs, [2]string{key, val})
}

// ---- value formatting --------------------------------------------------------

// Age renders how long ago ts was, in the compact form a table column wants:
// seconds under a minute, then minutes, then hours with minutes, then days
// with hours, then days alone past a week. It is this tool's own rendering,
// not a reproduction of any engine's, since every row in these tables is
// tool-rendered.
//
// An empty or unparseable timestamp yields "", which the table shows as a
// missing value rather than as a zero age.
func Age(startedAt string, now time.Time) string {
	t, ok := parseTime(startedAt)
	if !ok {
		return ""
	}
	d := now.Sub(t)
	if d < 0 {
		// A clock skew between this host and the engine's; report the start as
		// just-now rather than a negative age.
		d = 0
	}
	switch {
	case d < time.Minute:
		return strconv.Itoa(int(d.Seconds())) + "s"
	case d < time.Hour:
		return strconv.Itoa(int(d.Minutes())) + "m"
	case d < 24*time.Hour:
		h := int(d.Hours())
		return strconv.Itoa(h) + "h" + strconv.Itoa(int(d.Minutes())-h*60) + "m"
	case d < 7*24*time.Hour:
		days := int(d.Hours()) / 24
		return strconv.Itoa(days) + "d" + strconv.Itoa(int(d.Hours())-days*24) + "h"
	default:
		return strconv.Itoa(int(d.Hours())/24) + "d"
	}
}

// parseTime accepts the timestamp shapes the engines emit: RFC3339 with or
// without fractional seconds, and docker's zero value for "never started".
func parseTime(s string) (time.Time, bool) {
	s = strings.TrimSpace(s)
	if s == "" || strings.HasPrefix(s, "0001-01-01") {
		return time.Time{}, false
	}
	for _, layout := range []string{time.RFC3339Nano, time.RFC3339} {
		if t, err := time.Parse(layout, s); err == nil {
			return t, true
		}
	}
	return time.Time{}, false
}

// quantitySuffixes maps the unit suffixes a quantity can carry to their
// multiplier, longest first so "Mi" is never read as the decimal "M" and
// "MiB" never as "Mi". Binary (Ki/Mi/Gi) and decimal (k/M/G) are both listed
// because kubernetes accepts either in a manifest and the engines report
// sizes in the "MiB"/"GB" spellings.
var quantitySuffixes = []struct {
	suffix string
	mult   float64
}{
	{"KiB", 1 << 10}, {"MiB", 1 << 20}, {"GiB", 1 << 30}, {"TiB", 1 << 40},
	{"Ki", 1 << 10}, {"Mi", 1 << 20}, {"Gi", 1 << 30}, {"Ti", 1 << 40},
	{"kB", 1e3}, {"MB", 1e6}, {"GB", 1e9}, {"TB", 1e12},
	{"k", 1e3}, {"K", 1e3}, {"M", 1e6}, {"G", 1e9}, {"T", 1e12}, {"B", 1},
}

// ParseQuantity reads a kubernetes-style quantity ("120m", "1", "512Mi") or an
// engine-style size ("512MiB", "1.5GB") into a float in base units -- cores
// for CPU, bytes for memory. The "m" milli suffix is CPU-only and handled
// first, since no size suffix ends in a lowercase m.
//
// ok is false for anything it cannot read, which is how a caller decides to
// leave a percentage out rather than print a number computed from nothing.
func ParseQuantity(s string) (float64, bool) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, false
	}
	if num, found := strings.CutSuffix(s, "m"); found {
		if v, err := strconv.ParseFloat(strings.TrimSpace(num), 64); err == nil {
			return v / 1000, true
		}
		return 0, false
	}
	num, mult := s, 1.0
	for _, q := range quantitySuffixes {
		if rest, found := strings.CutSuffix(s, q.suffix); found {
			num, mult = rest, q.mult
			break // the table is ordered longest-first, so the first hit is the right one
		}
	}
	v, err := strconv.ParseFloat(strings.TrimSpace(num), 64)
	if err != nil {
		return 0, false
	}
	return v * mult, true
}

// Percent renders used/limit as a whole-number percentage, or "" when either
// side is missing or unreadable, or the limit is zero (no ceiling set, which
// is how docker reports an unlimited container).
func Percent(used, limit string) string {
	u, uok := ParseQuantity(used)
	l, lok := ParseQuantity(limit)
	if !uok || !lok || l <= 0 {
		return ""
	}
	return strconv.Itoa(int(u/l*100+0.5)) + "%"
}

// resourceLine renders a Resource as one value: "120m of 1 (12%)", dropping
// whatever is missing -- "120m" alone when nothing bounds it, "120m of 1" when
// the percentage could not be computed.
func resourceLine(r *Resource) string {
	if r == nil || r.Used == "" {
		return ""
	}
	s := r.Used
	if r.Limit != "" {
		s += " of " + r.Limit
	}
	if r.Percent != "" {
		s += " (" + r.Percent + ")"
	}
	return s
}

// Cores renders a CPU count from a nanocpu quota, the unit docker and podman
// report a container's CPU ceiling in. Zero means no ceiling, and yields "".
func Cores(nanoCPUs int64) string {
	if nanoCPUs <= 0 {
		return ""
	}
	c := float64(nanoCPUs) / 1e9
	if c == float64(int64(c)) {
		return strconv.FormatInt(int64(c), 10)
	}
	return strconv.FormatFloat(c, 'f', 2, 64)
}

// Bytes renders a byte count in the binary unit an operator reads limits in,
// matching how the same value is written in env.yaml ("1Gi", "512Mi"). Zero
// means no limit and yields "".
func Bytes(n int64) string {
	if n <= 0 {
		return ""
	}
	switch {
	case n >= 1<<30:
		return trimFloat(float64(n)/(1<<30)) + "Gi"
	case n >= 1<<20:
		return trimFloat(float64(n)/(1<<20)) + "Mi"
	case n >= 1<<10:
		return trimFloat(float64(n)/(1<<10)) + "Ki"
	default:
		return strconv.FormatInt(n, 10)
	}
}

// trimFloat renders a value with at most one decimal place and no trailing
// ".0", so a round limit reads "1Gi" rather than "1.0Gi".
func trimFloat(f float64) string {
	s := strconv.FormatFloat(f, 'f', 1, 64)
	return strings.TrimSuffix(s, ".0")
}

// ExitCodeText renders an exit code for a report line, or "" when the
// container has not terminated. Split out so the same wording serves the table
// reason column and the detail block.
func ExitCodeText(code *int) string {
	if code == nil {
		return ""
	}
	return "exit code " + strconv.Itoa(*code)
}

// SortInstances orders instances by namespace then name, so a report over
// several namespaces (--all) groups them and a repeated run prints the same
// order however the engine listed them.
func SortInstances(in []Instance) {
	sort.SliceStable(in, func(i, j int) bool {
		if in[i].Namespace != in[j].Namespace {
			return in[i].Namespace < in[j].Namespace
		}
		return in[i].Name < in[j].Name
	})
}

// ImageMismatch reports whether a running image is not the one env.yaml asks
// for. It is the failed-rollout check: a pod still on the previous tag after a
// rollout that never completed looks healthy in every other line of the report.
//
// The comparison is deliberately tolerant of the spellings the engines use for
// the same reference. Kubernetes normalises an image to its fully qualified
// form ("docker.io/solace/x:1" for "solace/x:1") and Docker Hub's official
// namespace ("library/") appears and disappears the same way, so both sides are
// reduced to a canonical form before being compared -- otherwise every
// kubernetes instance would report a mismatch against a correct env.yaml.
//
// When env.yaml pins a digest rather than a tag, the digest actually running is
// what the answer must come from: the reference then carries no tag to compare.
func ImageMismatch(running, digest, want string) bool {
	if running == "" || want == "" {
		return false // nothing collected to compare, so nothing to claim
	}
	if i := strings.Index(want, "@sha256:"); i >= 0 {
		if digest == "" {
			return false // pinned by digest, but the engine reported none
		}
		return digest != want[i+1:]
	}
	return canonicalRef(running) != canonicalRef(want)
}

// canonicalRef strips the registry and namespace defaults an engine may add or
// omit, so two spellings of one reference compare equal.
func canonicalRef(ref string) string {
	for _, prefix := range []string{"docker.io/", "index.docker.io/", "registry-1.docker.io/"} {
		ref = strings.TrimPrefix(ref, prefix)
	}
	ref = strings.TrimPrefix(ref, "library/")
	// An untagged reference means :latest to every engine, and one side often
	// spells it while the other does not.
	if !strings.Contains(ref[strings.LastIndex(ref, "/")+1:], ":") {
		ref += ":latest"
	}
	return ref
}
