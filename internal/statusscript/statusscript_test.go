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

	for _, want := range []string{ContainerPath, "/run/secrets", `from_configs "/^[[:space:]]*-[[:space:]]*name:`} {
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

	// Every exit in the script is an exit 0, including the early error returns.
	for i, line := range strings.Split(out, "\n") {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, "exit ") && !strings.Contains(trimmed, ") exit ") {
			continue
		}
		if strings.Contains(trimmed, "exit 0") {
			continue
		}
		t.Errorf("line %d exits non-zero: %q", i+1, trimmed)
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

// TestRenderSendsStatusToStdoutAndProblemsToStderr pins the stream split: the
// report is stdout, every diagnostic is redirected to stderr.
func TestRenderSendsStatusToStdoutAndProblemsToStderr(t *testing.T) {
	out := Render(8090, "solmq-status")

	// Each report line must appear exactly as written -- no >&2 on the end, or
	// the report would land on the error stream.
	for _, want := range []string{
		`echo "leader-election mode: ${MODE:-unknown}"` + "\n",
		`echo "leader-election state: ${STATE:-unknown}"` + "\n",
		`echo "workflow $id: ${st:-unknown}"` + "\n",
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
	if Filename != "status" {
		t.Errorf("Filename = %q, want %q", Filename, "status")
	}
	if ContainerDir != "/app/external/libs" {
		t.Errorf("ContainerDir = %q, want %q", ContainerDir, "/app/external/libs")
	}
	if want := ContainerDir + "/" + Filename; ContainerPath != want {
		t.Errorf("ContainerPath = %q, want %q", ContainerPath, want)
	}
	if ConfigPath != "/app/external/spring/config/application.yml" {
		t.Errorf("ConfigPath = %q, want %q", ConfigPath, "/app/external/spring/config/application.yml")
	}
}
