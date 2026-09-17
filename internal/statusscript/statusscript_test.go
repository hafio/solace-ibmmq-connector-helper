package statusscript

import (
	"strconv"
	"strings"
	"testing"
)

// TestRenderSubstitution covers port and user substitution, including
// non-default values, and pins the two assignment lines and the endpoint
// URLs (which are built from PORT at run time, not by Render) into shape.
func TestRenderSubstitution(t *testing.T) {
	tests := []struct {
		name     string
		mgmtPort int
		user     string
	}{
		{name: "defaults", mgmtPort: 8090, user: "solmq-status"},
		{name: "non-default port and user", mgmtPort: 19090, user: "custom-mgmt-user"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			out := Render(tc.mgmtPort, tc.user)

			wantPort := "PORT=" + strconv.Itoa(tc.mgmtPort) + "\n"
			if !strings.Contains(out, wantPort) {
				t.Errorf("missing %q in:\n%s", wantPort, out)
			}
			wantUser := "USER_NAME=" + tc.user + "\n"
			if !strings.Contains(out, wantUser) {
				t.Errorf("missing %q in:\n%s", wantUser, out)
			}

			// The leaderelection and workflows endpoints are built from
			// $PORT/$BASE at run time inside the script, not by Render, so
			// pin that indirection rather than a literal port in the URL.
			if !strings.Contains(out, `BASE="http://127.0.0.1:$PORT/actuator"`) {
				t.Errorf("BASE not built from $PORT:\n%s", out)
			}
			if !strings.Contains(out, `$BASE/leaderelection`) {
				t.Errorf("leaderelection endpoint missing:\n%s", out)
			}
			if !strings.Contains(out, `$BASE/workflows`) {
				t.Errorf("workflows endpoint missing:\n%s", out)
			}
		})
	}
}

// TestRenderIsPureASCIINoCRLF guards the plain-ASCII, LF-only contract: the
// script runs through busybox sh, which chokes on a CRLF shebang line and
// this tool must never emit non-ASCII into a generated artifact.
func TestRenderIsPureASCIINoCRLF(t *testing.T) {
	out := Render(8090, "solmq-status")

	if strings.Contains(out, "\r") {
		t.Error("output contains a carriage return -- must be LF-only")
	}
	for i := 0; i < len(out); i++ {
		if out[i] > 127 {
			t.Fatalf("output contains a non-ASCII byte at offset %d: %q", i, out[i])
		}
	}
	if !strings.HasSuffix(out, "\n") {
		t.Error("output must end with a trailing newline")
	}
}

// TestRenderHeaderHasExecOneLiners pins the three documented ways to invoke
// the script, each built from ContainerPath so they can't drift from the
// path the caller actually mounts it at.
func TestRenderHeaderHasExecOneLiners(t *testing.T) {
	out := Render(8090, "solmq-status")

	for _, want := range []string{
		"kubectl exec <pod> -- sh " + ContainerPath,
		"docker exec <container> sh " + ContainerPath,
		"podman exec <container> sh " + ContainerPath,
		// The healthcheck is the one caller that passes an argument, so the
		// header has to name it or the only non-zero exit looks unexplained.
		"sh " + ContainerPath + " " + HealthArg,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing exec one-liner %q in:\n%s", want, out)
		}
	}
}

// TestRenderPasswordResolution covers the password-lookup chain: the config
// file path, the secrets-model fallback directory, and that no literal
// credential is ever embedded in the script.
func TestRenderPasswordResolution(t *testing.T) {
	out := Render(8090, "solmq-status")

	for _, want := range []string{ContainerPath, SecretsDir, `from_configs "/^[[:space:]]*-[[:space:]]*name:`} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in password-resolution logic:\n%s", want, out)
		}
	}
	// The script resolves the password from the config it finds at run time --
	// it must never carry one of its own, so no assignment may hold a value.
	if strings.Contains(out, "PASS=solmq") || strings.Contains(out, `PASS="solmq`) {
		t.Error("script must not embed a credential")
	}
}

// TestRenderSearchesSpringConfigLocations pins the config search order. The
// script cannot assume this tool mounted the config: a foreign instance may
// point Spring somewhere else entirely, and reading the wrong file would mean
// authenticating with a stale password or misjudging the exposure list.
func TestRenderSearchesSpringConfigLocations(t *testing.T) {
	out := Render(8090, "solmq-status")

	for _, want := range []string{
		// Spring's own override variables, in precedence order.
		"${SPRING_CONFIG_LOCATION:-}",
		"${SPRING_CONFIG_ADDITIONAL_LOCATION:-}",
		"${SPRING_CONFIG_NAME:-application}",
		// The directory the image's entrypoint passes as an additional
		// location, including its wildcard subdirectory form.
		ConfigDir + "/",
		ConfigDir + "/*/",
		// Spring's working-directory defaults.
		"./ ./config/ ./config/*/",
		// Both YAML extensions Spring accepts for the config name.
		`"$loc$CONFIG_NAME.yml" "$loc$CONFIG_NAME.yaml"`,
		// A location list is comma-separated, and each entry may carry
		// optional: and file: prefixes.
		"IFS=','",
		"loc=${loc#optional:}",
		"loc=${loc#file:}",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in config search:\n%s", want, out)
		}
	}
	// classpath: locations live inside the jar, so the script must skip them
	// rather than test them as filesystem paths.
	if !strings.Contains(out, "classpath:*) continue ;;") {
		t.Errorf("classpath locations must be skipped:\n%s", out)
	}
	// Searching must happen before both consumers of the config.
	searchIdx := strings.Index(out, "CONFIGS=$(config_files)")
	exposureIdx := strings.Index(out, `has_entry "$EXPOSURE" leaderelection`)
	passIdx := strings.Index(out, "PASS=$(from_configs")
	if searchIdx < 0 || exposureIdx < searchIdx || passIdx < searchIdx {
		t.Errorf("config search must precede the exposure check and password lookup:\n%s", out)
	}
}

// TestRenderAlwaysExitsZero pins the exit contract: the script reports through
// its output, never through the exit status, so an exec wrapper can read a
// non-zero exit as "could not reach or run this instance" without a standby
// instance or a misconfigured endpoint looking the same.
func TestRenderAlwaysExitsZero(t *testing.T) {
	out := Render(8090, "solmq-status")
	first, block, rest := splitHealthBlock(t, out)

	// Every exit in the report path is an exit 0, including the early error
	// returns. The healthcheck block is excluded because its whole job is to
	// answer with an exit status; splitHealthBlock proves it is the only part
	// of the script carved out here, and TestRenderHealthModeExitsOnVerdict
	// pins what it exits with.
	for i, line := range strings.Split(first+rest, "\n") {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, "exit ") && !strings.Contains(trimmed, ") exit ") {
			continue
		}
		if strings.Contains(trimmed, "exit 0") {
			continue
		}
		t.Errorf("report-path line %d exits non-zero: %q", i+1, trimmed)
	}
	// And the carve-out is exactly one block: a second non-zero exit anywhere
	// would mean some other path had quietly started reporting failure.
	if n := strings.Count(out, "exit 1"); n != 1 {
		t.Errorf("found %d `exit 1` in the script, want exactly 1 (the healthcheck verdict)", n)
	}
	if !strings.Contains(block, "exit 1") {
		t.Errorf("the only non-zero exit is not the healthcheck verdict; block was:\n%s", block)
	}
	// set -e would abort mid-script with a non-zero status, so it must be off,
	// while set -u stays on to catch a mistyped variable name.
	if strings.Contains(out, "set -eu") || strings.Contains(out, "set -e\n") {
		t.Errorf("set -e must not be used, it would break the always-0 contract:\n%s", out)
	}
	if !strings.Contains(out, "set -u") {
		t.Errorf("set -u should stay on:\n%s", out)
	}
	// The trap holds the contract even if set -u aborts on an unguarded name.
	if !strings.Contains(out, "trap 'exit 0' EXIT") {
		t.Errorf("missing the EXIT trap that guarantees a 0 exit:\n%s", out)
	}
	// active and standby are both normal, already-printed answers, so neither
	// branch may report anything on stderr.
	if !strings.Contains(out, "active | standby) ;;") {
		t.Errorf("active and standby must both be quiet, normal outcomes:\n%s", out)
	}
}

// splitHealthBlock cuts the rendered script into the part before the --health
// branch, the branch itself, and the part after it. The branch is the script's
// one deliberate departure from the always-exit-0 contract, so several tests
// need to reason about it separately from the report path.
func splitHealthBlock(t *testing.T, out string) (before, block, after string) {
	t.Helper()

	guard := `if [ "${1:-}" = "` + HealthArg + `" ]; then` + "\n"
	i := strings.Index(out, guard)
	if i < 0 {
		t.Fatalf("no %s branch in:\n%s", HealthArg, out)
	}
	// The branch nests only if/fi, so the first line that is exactly "fi" ends
	// it -- the inner `if [ "$H" = "UP" ]` closes with an indented "fi".
	rel := strings.Index(out[i:], "\nfi\n")
	if rel < 0 {
		t.Fatalf("%s branch is never closed in:\n%s", HealthArg, out)
	}
	end := i + rel + len("\nfi\n")
	return out[:i], out[i:end], out[end:]
}

// TestRenderHealthModeShortCircuits pins the reason the healthcheck can run on
// a timer at all: it must answer with one call to /actuator/health and stop.
// The full report makes seven actuator calls and spawns a JVM for
// `java -version`, which is far too much to repeat every interval inside a
// container that is already memory- and CPU-limited.
//
// It also pins the branch's position. The exposure check below it exits 0 when
// leaderelection is unexposed, which a healthcheck would read as healthy, so
// the branch has to sit above it -- and below get(), whose credentials it
// needs, since the actuator is always behind the injected read-only account.
func TestRenderHealthModeShortCircuits(t *testing.T) {
	out := Render(8090, "solmq-status")
	before, block, _ := splitHealthBlock(t, out)

	// Below get() and the password resolution it depends on.
	for _, want := range []string{"get() {", "PASS=$(from_configs"} {
		if !strings.Contains(before, want) {
			t.Errorf("%q must be resolved before the %s branch; it is not in:\n%s", want, HealthArg, before)
		}
	}
	// Above the exposure check and the report's first query.
	for _, notYet := range []string{`if [ -n "$CONFIGS" ]; then`, `$BASE/leaderelection`} {
		if strings.Contains(before, notYet) {
			t.Errorf("%q appears before the %s branch, so the branch cannot short-circuit it", notYet, HealthArg)
		}
	}
	// One call, and only the health one.
	if !strings.Contains(block, `get "$BASE/health"`) {
		t.Errorf("the %s branch does not query /actuator/health:\n%s", HealthArg, block)
	}
	if n := strings.Count(block, "get \"$BASE/"); n != 1 {
		t.Errorf("the %s branch makes %d actuator calls, want exactly 1:\n%s", HealthArg, n, block)
	}
	// None of the report's enrichment, above all not the JVM spawn.
	for _, tooMuch := range []string{"metric ", "java -version", "$BASE/info", "$BASE/workflows"} {
		if strings.Contains(block, tooMuch) {
			t.Errorf("the %s branch does %q, which belongs to the report path only:\n%s", HealthArg, tooMuch, block)
		}
	}
}

// TestRenderHealthModeExitsOnVerdict pins what the engine actually reads. The
// EXIT trap that guarantees the report's exit 0 would swallow a non-zero
// verdict, so the branch has to clear it before answering; without that, every
// container reports healthy forever.
func TestRenderHealthModeExitsOnVerdict(t *testing.T) {
	out := Render(8090, "solmq-status")
	_, block, _ := splitHealthBlock(t, out)

	trap := strings.Index(block, "trap - EXIT")
	if trap < 0 {
		t.Fatalf("the %s branch never clears the EXIT trap, so its exit status cannot escape:\n%s", HealthArg, block)
	}
	if e := strings.Index(block, "exit "); e >= 0 && e < trap {
		t.Errorf("the %s branch exits before clearing the EXIT trap:\n%s", HealthArg, block)
	}
	// UP and nothing else is healthy: a DOWN, an OUT_OF_SERVICE or an
	// unreachable actuator all have to come out non-zero.
	if !strings.Contains(block, `if [ "$H" = "UP" ]; then`) {
		t.Errorf("the %s branch does not gate its success on UP:\n%s", HealthArg, block)
	}
	if !strings.Contains(block, "exit 0") || !strings.Contains(block, "exit 1") {
		t.Errorf("the %s branch must exit 0 on UP and 1 otherwise:\n%s", HealthArg, block)
	}
	// The engine keeps stdout in the container's health log, so the verdict is
	// what a later inspect shows. It is a report line, not a diagnostic, so it
	// must not carry the "status:" prefix the stderr convention reserves.
	if !strings.Contains(block, `echo "${H:-unreachable}"`) {
		t.Errorf("the %s branch does not report the verdict it saw:\n%s", HealthArg, block)
	}
	if strings.Contains(block, `echo "status:`) {
		t.Errorf("the %s branch emits a status: diagnostic, which belongs on stderr:\n%s", HealthArg, block)
	}
}

// TestRenderSendsStatusToStdoutAndProblemsToStderr pins the stream split: the
// report is stdout, every diagnostic is redirected to stderr.
func TestRenderSendsStatusToStdoutAndProblemsToStderr(t *testing.T) {
	out := Render(8090, "solmq-status")

	// Each report line must appear exactly as written -- no >&2 on the end, or
	// the report would land on the error stream.
	for _, want := range []string{
		`echo "leader-election mode: ${MODE:-unknown}"` + "\n",
		`echo "leader-election state: ${STATE:-unknown}"` + "\n",
		`echo "workflows:"` + "\n",
		`echo "  $wf_id: $wf_state"` + "\n",
		`echo "health: UP"` + "\n",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing unredirected stdout report line %q in:\n%s", want, out)
		}
	}
	// Every "status:" diagnostic is a problem report and must go to stderr.
	for i, line := range strings.Split(out, "\n") {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, `echo "status:`) {
			continue
		}
		if !strings.HasSuffix(trimmed, ">&2") {
			t.Errorf("line %d is a diagnostic but not redirected to stderr: %q", i+1, trimmed)
		}
	}
}

// TestRenderVerifiesExposure pins the up-front exposure check: an unexposed
// leaderelection endpoint answers 404, which is indistinguishable from a wrong
// port or a bad credential once it reaches wget, so the script must read the
// mounted config and fail loudly before querying.
func TestRenderVerifiesExposure(t *testing.T) {
	out := Render(8090, "solmq-status")

	// The include list is located under the exposure key rather than by a bare
	// include match, so an unrelated include elsewhere in the file cannot be
	// mistaken for it.
	if !strings.Contains(out, "exposure:") || !strings.Contains(out, "include:") {
		t.Errorf("exposure include lookup missing:\n%s", out)
	}
	// Comma-delimited membership, so leaderelection2 cannot satisfy the check
	// and Spring's "*" wildcard does.
	for _, want := range []string{
		"has_entry() {",
		`*",$2,"*) return 0 ;;`,
		`*",*,"*) return 0 ;;`,
		`if ! has_entry "$EXPOSURE" leaderelection; then`,
		`if ! has_entry "$EXPOSURE" workflows; then`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in exposure check:\n%s", want, out)
		}
	}
	// An unexposed leaderelection stops the run (there is nothing left to
	// report), while missing workflows only degrades the report and an
	// unlocatable config leaves both unverified rather than failing an
	// otherwise healthy instance. Stopping still exits 0 -- see
	// TestRenderAlwaysExitsZero.
	leIdx := strings.Index(out, `has_entry "$EXPOSURE" leaderelection`)
	wfIdx := strings.Index(out, `has_entry "$EXPOSURE" workflows`)
	if leIdx < 0 || wfIdx <= leIdx {
		t.Fatalf("expected the leaderelection check before the workflows check:\n%s", out)
	}
	gate := out[leIdx:wfIdx]
	if !strings.Contains(gate, "exit 0") {
		t.Errorf("unexposed leaderelection must stop the run:\n%s", out)
	}
	if !strings.Contains(gate, ">&2") {
		t.Errorf("the unexposed-leaderelection message must go to stderr:\n%s", out)
	}
	if !strings.Contains(out, "found no application config") {
		t.Errorf("a config that cannot be located must warn rather than fail:\n%s", out)
	}
	// The check must run before the first request, or it saves nothing.
	if reqIdx := strings.Index(out, `LE=$(get`); reqIdx < 0 || leIdx > reqIdx {
		t.Error("exposure check must precede the first actuator request")
	}
}

// TestRenderAlignsWorkflowColumn pins the layout of the workflows block: a
// bare header, then one indented row per workflow with the ids right-aligned so
// every colon sits in the same column. Left-aligning them steps 10 one place
// right of 0, which is exactly what makes a twenty-workflow report hard to scan.
//
// The width is measured from the ids actually present rather than assumed from
// the connector's twenty-slot cap, so a report of single-digit ids is not padded
// for a two-digit id that is not there.
func TestRenderAlignsWorkflowColumn(t *testing.T) {
	out := Render(8090, "solmq-status")
	for _, want := range []string{
		`echo "workflows:"`,
		`WF_WIDTH=1`,
		`if [ ${#wf_id} -gt "$WF_WIDTH" ]; then`,
		`while [ ${#wf_id} -lt "$WF_WIDTH" ]; do`,
		`echo "  $wf_id: $wf_state"`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in:\n%s", want, out)
		}
	}
}

// TestRenderReportsHealthUptimeAndVersion covers the three enrichment lines.
// All three endpoints are in the fixed exposure list this tool writes, so they
// need no extra configuration -- but each is read defensively, because a
// hand-written config or a future image may not answer.
//
// Each line is dropped when its endpoint says nothing, rather than reported as
// missing: none of them tells you whether the instance is doing its job, so a
// blank or a warning would be noise next to the leader-election and workflow
// lines that do.
func TestRenderReportsHealthUptimeAndVersion(t *testing.T) {
	out := Render(8090, "solmq-status")
	for _, want := range []string{
		`HEALTH=$(get "$BASE/health")`,
		`echo "health: UP"`,
		`echo "health-detail: $HEALTH"`,
		`get "$BASE/metrics/$1"`,
		`UPTIME=$(metric process.uptime)`,
		`INFO=$(get "$BASE/info")`,
		`echo "version: $VERSION"`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in:\n%s", want, out)
		}
	}
	// Each enrichment line sits behind a non-empty guard, so an endpoint that
	// answers nothing drops its line instead of printing an empty value.
	for _, guard := range []string{
		`if [ -n "$HEALTH" ]; then`,
		`if [ -n "$UPTIME" ]; then`,
		`if [ -n "$VERSION" ]; then`,
	} {
		if !strings.Contains(out, guard) {
			t.Errorf("enrichment line is not guarded (%q missing):\n%s", guard, out)
		}
	}
	// The instance's own health is the first status in the document; matching the
	// last would report a component's status as the whole instance's.
	if !strings.Contains(out, `s/^[^{]*{`) {
		t.Errorf("health status match is not anchored at the opening brace:\n%s", out)
	}
}

// TestRenderEscapesUserForSedAddress covers the sanitize-at-use half of the
// account-name gate: the name is spliced into a sed address, so it is emitted
// regex-escaped there while the Authorization header still gets the raw name.
// Without this a name carrying '/' would end the address early and leave the
// script silently unauthenticated rather than failing.
func TestRenderEscapesUserForSedAddress(t *testing.T) {
	tests := []struct {
		user      string
		wantMatch string
	}{
		{user: "solmq-status", wantMatch: `USER_MATCH='solmq-status'`},
		{user: "svc.status", wantMatch: `USER_MATCH='svc\.status'`},
		// Defense in depth: these never pass validate.SafeActuatorUser, but the
		// escaping must hold for anything that reaches Render by another route.
		{user: "a/b", wantMatch: `USER_MATCH='a\/b'`},
		{user: "a[b]", wantMatch: `USER_MATCH='a\[b\]'`},
		{user: "a*b", wantMatch: `USER_MATCH='a\*b'`},
		{user: `a\b`, wantMatch: `USER_MATCH='a\\b'`},
		{user: "^ab$", wantMatch: `USER_MATCH='\^ab\$'`},
	}
	for _, tc := range tests {
		out := Render(8090, tc.user)
		if !strings.Contains(out, tc.wantMatch) {
			t.Errorf("Render(user=%q): missing %q", tc.user, tc.wantMatch)
		}
		// The raw name is what the actuator must see in the credential.
		if !strings.Contains(out, "USER_NAME="+tc.user+"\n") {
			t.Errorf("Render(user=%q): auth name must stay unescaped", tc.user)
		}
		// The sed address must consume the escaped copy, never the raw one.
		if !strings.Contains(out, `name:[[:space:]]*$USER_MATCH[[:space:]]*$/`) {
			t.Errorf("Render(user=%q): sed address must use USER_MATCH:\n%s", tc.user, out)
		}
	}
}

// TestFilenameAndPathConstants pins the exported path constants against each
// other so ContainerPath can never silently diverge from Filename/ContainerDir.
func TestFilenameAndPathConstants(t *testing.T) {
	if Filename != ".status-script" {
		t.Errorf("Filename = %q, want %q", Filename, ".status-script")
	}
	// Not the libs mount, and not nested in any other mount: the script used to
	// live under /app/external/libs, where a libs bind mount could shadow it
	// unless every renderer declared the mounts in the right order.
	if ContainerDir != "/app/external" {
		t.Errorf("ContainerDir = %q, want %q", ContainerDir, "/app/external")
	}
	for _, mount := range []string{"/app/external/libs", "/app/external/spring/config", "/app/external/classpath"} {
		if strings.HasPrefix(ContainerPath, mount+"/") {
			t.Errorf("ContainerPath %q is nested inside the %q mount, which is the conflict this path avoids", ContainerPath, mount)
		}
	}
	if want := ContainerDir + "/" + Filename; ContainerPath != want {
		t.Errorf("ContainerPath = %q, want %q", ContainerPath, want)
	}
	if ConfigPath != "/app/external/spring/config/application.yml" {
		t.Errorf("ConfigPath = %q, want %q", ConfigPath, "/app/external/spring/config/application.yml")
	}
	// The healthcheck contract the three deployment renderers build their
	// commands from. HealthShell matters because the mounted script never
	// carries the execute bit on any platform.
	if HealthArg != "--health" {
		t.Errorf("HealthArg = %q, want %q", HealthArg, "--health")
	}
	if HealthShell != "sh" {
		t.Errorf("HealthShell = %q, want %q", HealthShell, "sh")
	}
	// A timeout at or past the interval would let one slow check overlap the
	// next, and a start period shorter than the timeout budget the retries
	// need would report a still-booting JVM as unhealthy.
	if HealthTimeoutSeconds >= HealthIntervalSeconds {
		t.Errorf("HealthTimeoutSeconds = %d must be below HealthIntervalSeconds = %d", HealthTimeoutSeconds, HealthIntervalSeconds)
	}
	if HealthRetries < 1 {
		t.Errorf("HealthRetries = %d, want at least 1", HealthRetries)
	}
	if HealthStartPeriodSeconds <= HealthIntervalSeconds {
		t.Errorf("HealthStartPeriodSeconds = %d should exceed HealthIntervalSeconds = %d", HealthStartPeriodSeconds, HealthIntervalSeconds)
	}
}

// TestRenderReportsEveryWorkflowInNumericOrder pins the two things that decide
// whether the workflow half of the report is trustworthy at all.
//
// Order: the actuator hands the workflows back in its own map order, which
// lists 10 before 2, so the loop tags each line with the id as a leading
// tab-separated sort key, sorts numerically, and cuts the key back off.
//
// Completeness: the response is fed in with a terminating newline. read
// returns non-zero on a final line that carries no newline and the loop body
// is skipped for it, so a bare printf of the response dropped the last
// workflow from every report -- the failure mode that silently under-reports
// rather than looking wrong.
func TestRenderReportsEveryWorkflowInNumericOrder(t *testing.T) {
	out := Render(8090, "solmq-status")
	for _, want := range []string{
		`printf '%s\n' "$WF" | tr '{'`,
		`done | sort -n)`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in:\n%s", want, out)
		}
	}
	// A bare printf of the response would reintroduce the dropped-last-workflow
	// bug, so the unterminated form must not come back.
	if strings.Contains(out, `printf %s "$WF"`) {
		t.Errorf("workflows response fed in without a terminating newline:\n%s", out)
	}
}

// TestRenderReportsOnlyConfiguredWorkflows pins the noise filter. Two things
// get dropped:
//
//   - a chunk with no state at all. Splitting the response on '{' also hands
//     the loop whatever nested objects an element contains, and such a fragment
//     can carry an id of its own.
//   - a state of N/A, case-insensitively. The connector reports its full set of
//     workflow slots and marks every unconfigured one N/A, which is how one real
//     workflow produced twenty lines of report.
//
// What is deliberately NOT filtered is an allowlist of the real states:
// running/stopped/paused/error are not a closed set, so only the placeholder
// goes, and a state this script has never seen still reaches the operator.
//
// Filtering every entry away is the systemic case, not the per-entry one, so it
// is reported on stderr rather than leaving the workflow half of the report
// silently blank (S4a fail-loud vs skip).
func TestRenderReportsOnlyConfiguredWorkflows(t *testing.T) {
	out := Render(8090, "solmq-status")
	for _, want := range []string{
		`[ -n "$st" ] || continue`,
		`case "$st" in [Nn]/[Aa]) continue ;; esac`,
		`if [ -n "$WF_ROWS" ]; then`,
		`echo "status: $BASE/workflows reported no workflow on this active instance`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in:\n%s", want, out)
		}
	}
	// The unknown-state placeholder is what the id/state guard replaced; leaving
	// it in would put back exactly the lines this is meant to drop.
	if strings.Contains(out, `${st:-unknown}`) {
		t.Errorf("stateless workflows are still padded with an unknown line:\n%s", out)
	}
}

// TestRenderWarnsOnEmptyWorkflowsOnlyWhenActive pins the standby case. An
// active_standby deployment with replicas > 1 always has a standby, and a
// standby runs no workflow -- so an empty workflow list there is the normal
// shape, not a fault. Warning on it unconditionally made every standby pod read
// as broken, which at replicas: 2 is half of every report.
//
// The gate is the leader-election state the script has already resolved, so
// standalone and active_active (both of which report active) keep the warning,
// and an unset or unrecognized state falls through silently -- the script flags
// that separately at the end.
func TestRenderWarnsOnEmptyWorkflowsOnlyWhenActive(t *testing.T) {
	out := Render(8090, "solmq-status")
	want := `elif [ "$STATE" = "active" ]; then`
	if !strings.Contains(out, want) {
		t.Errorf("empty-workflow warning is not gated on the active state (%q missing):\n%s", want, out)
	}
	// The ungated form is the regression: a bare else here warns on every
	// standby. $STATE is resolved well above this block, so there is no ordering
	// reason to fall back to one.
	if strings.Contains(out, "  else\n"+`    echo "status: $BASE/workflows`) {
		t.Errorf("empty-workflow warning is still ungated:\n%s", out)
	}
}

// TestRenderReportsHealthComponents covers the per-component breakdown, which
// is the app-level fact closest to "is it actually moving messages". Spring
// serialises each component as "<name>":{"status":"X"}, so the parse puts a
// newline before every {"status" and carries the name forward from the line
// above -- the assertions below pin that mechanism, since a regression would
// silently report no components at all rather than fail.
func TestRenderReportsHealthComponents(t *testing.T) {
	out := Render(8090, "solmq-status")
	for _, want := range []string{
		`echo "health components:"`,
		`sed -e 's/{[[:space:]]*"status"/`,
		`if [ -n "$st" ] && [ -n "${pending:-}" ]; then`,
		`pending=$(printf %s "$chunk" | sed -n 's/.*"`,
		`if [ -n "$HC_ROWS" ]; then`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in:\n%s", want, out)
		}
	}
	// The block only prints when something was parsed, and the whole thing sits
	// inside the health guard -- no health document, no components.
	if strings.Index(out, "HC_ROWS=") < strings.Index(out, `HEALTH=$(get "$BASE/health")`) {
		t.Error("the component parse must come after the health document is fetched")
	}
	// $pending is read with set -u in force, so it needs the default-value form.
	if strings.Contains(out, `[ -n "$pending" ]`) {
		t.Errorf("an unguarded $pending would abort under set -u:\n%s", out)
	}
}

// TestRenderReportsJavaConfigAndHeap covers the three details-level lines that
// come from outside the actuator's report endpoints: the JVM the instance runs
// on, the configuration the report was read from, and heap use.
func TestRenderReportsJavaConfigAndHeap(t *testing.T) {
	out := Render(8090, "solmq-status")
	for _, want := range []string{
		// java writes -version to stderr, so the redirect is load-bearing.
		`if command -v java >/dev/null 2>&1; then`,
		`JAVA_RAW=$(java -version 2>&1 | head -n 1)`,
		`echo "java: $JAVA_LINE"`,
		`echo "java: $JAVA_RAW"`,
		`echo "config: $CONFIG_LIST"`,
		`HEAP_USED=$(metric 'jvm.memory.used?tag=area:heap')`,
		`HEAP_MAX=$(metric 'jvm.memory.max?tag=area:heap')`,
		`echo "heap: $HEAP_USED of $HEAP_MAX"`,
		`echo "heap: $HEAP_USED"`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in:\n%s", want, out)
		}
	}
	// Every one of the three is guarded, so an image without java, an instance
	// with no readable config, and a JVM that reports no heap metric each drop
	// their line rather than printing an empty value.
	for _, guard := range []string{
		`if [ -n "${JAVA_LINE:-}" ]; then`,
		`if [ -n "$CONFIG_LIST" ]; then`,
		`if [ -n "$HEAP_USED" ]; then`,
	} {
		if !strings.Contains(out, guard) {
			t.Errorf("line is not guarded (%q missing):\n%s", guard, out)
		}
	}
	// An unbounded heap reports a negative maximum, which must not be printed as
	// a ceiling.
	if !strings.Contains(out, `"" | -*) echo "heap: $HEAP_USED" ;;`) {
		t.Errorf("a negative or absent heap maximum must be left out:\n%s", out)
	}
	// The area:heap tag is what makes the number comparable with -Xmx rather
	// than including metaspace and the code cache.
	if !strings.Contains(out, "tag=area:heap") {
		t.Errorf("heap metrics must be tagged to the heap area:\n%s", out)
	}
	// The bytes are handed over raw on purpose: busybox arithmetic would read
	// Jackson's scientific notation (4.32013312E8) as 4.
	if strings.Contains(out, "HEAP_USED / 1048576") {
		t.Errorf("the script must not do the byte arithmetic itself:\n%s", out)
	}
}

// TestRenderHeaderNamesEveryReportedFact keeps the script's own header honest:
// it is the first thing an operator reads when running the script by hand, so
// it has to name what the script now reports.
func TestRenderHeaderNamesEveryReportedFact(t *testing.T) {
	out := Render(8090, "solmq-status")
	header := out[:strings.Index(out, "PORT=")]
	for _, want := range []string{"leader-election", "health", "java", "heap", "version"} {
		if !strings.Contains(header, want) {
			t.Errorf("the header does not mention %q:\n%s", want, header)
		}
	}
}
