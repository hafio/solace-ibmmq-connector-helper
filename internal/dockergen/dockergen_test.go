package dockergen

import (
	"strings"
	"testing"

	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/spec"
)

// appYAML1 exercises nested keys, a blank line, and a trailing newline so the
// block-scalar indentation and blank-line preservation are both covered.
const appYAML1 = `spring:
  application:
    name: solmq

logging:
  level: INFO
`

// statusScript1 exercises a blank line too, so the status config's content
// block is checked against the same blank-line-preservation behavior as
// application.yml's.
const statusScript1 = `#!/bin/sh
STATUS_URL="http://localhost:8080/actuator/health"

curl -fsS "$STATUS_URL"
`

// TestRenderFull_WithEverything is the primary full-string golden: a single
// instance with creds (rendered as a per-service secrets list plus a
// top-level environment-provider secrets block -- never an env_file), a store
// mount, a libs mount, MQTLS, ports and a timezone.
func TestRenderFull_WithEverything(t *testing.T) {
	in := Input{
		Docker: &spec.Docker{
			Name:    "solmq-connector",
			Restart: "unless-stopped",
			Ports:   []spec.Port{{Host: 8090, Container: 8090}, {Host: 8080, Container: 8091}},
		},
		Instance: Instance{
			Name:         "solmq-connector",
			Image:        "solace/solace-pubsub-connector-ibmmq:2.13.0",
			Timezone:     "Asia/Singapore",
			AppYAML:      appYAML1,
			MQTLS:        true,
			StatusScript: statusScript1,
			LeaderMode:   spec.LeaderActiveActive,
		},
		Secrets: []string{"SOLACE_CLIENT_USERNAME", "SOLACE_CLIENT_PASSWORD", "TRUSTSTORE_PASSWORD"},
		Stores: []Mount{
			{Source: "/abs/certs/truststore.jks", Target: "/app/external/classpath/truststores/truststore.jks"},
		},
		Libs: &Mount{Source: "/abs/libs", Target: "/app/external/libs"},
	}

	want := `services:
  solmq-connector:
    image: solace/solace-pubsub-connector-ibmmq:2.13.0
    container_name: solmq-connector
    labels:
      solace-connector/le-mode: active_active
      solace-connector/role: active
    restart: unless-stopped
    ports:
      - "8090:8090"
      - "8080:8091"
    environment:
      TZ: Asia/Singapore
      JAVA_TOOL_OPTIONS: "-Dcom.ibm.mq.cfg.useIBMCipherMappings=false"
    secrets:
      - SOLACE_CLIENT_USERNAME
      - SOLACE_CLIENT_PASSWORD
      - TRUSTSTORE_PASSWORD
    configs:
      - source: solmq-connector-app
        target: /app/external/spring/config/application.yml
      - source: solmq-connector-status
        target: /app/external/.status-script
    volumes:
      - /abs/certs/truststore.jks:/app/external/classpath/truststores/truststore.jks:ro
      - /abs/libs:/app/external/libs:ro
configs:
  solmq-connector-app:
    content: |
      spring:
        application:
          name: solmq

      logging:
        level: INFO
  solmq-connector-status:
    content: |
      #!/bin/sh
      STATUS_URL="http://localhost:8080/actuator/health"

      curl -fsS "$$STATUS_URL"
secrets:
  SOLACE_CLIENT_USERNAME:
    environment: SOLACE_CLIENT_USERNAME
  SOLACE_CLIENT_PASSWORD:
    environment: SOLACE_CLIENT_PASSWORD
  TRUSTSTORE_PASSWORD:
    environment: TRUSTSTORE_PASSWORD
`

	if got := Render(in); got != want {
		t.Errorf("Render mismatch:\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
	if strings.Contains(want, "env_file") {
		t.Fatal("test fixture itself contains env_file -- fixture is stale")
	}
	// One instance means one service, so exactly one ports: block even though it
	// lists two ports -- pins the single-instance shape now that sharding (which
	// used to emit one ports: block per shard, each publishing the same host
	// port) is gone.
	if c := strings.Count(want, "ports:"); c != 1 {
		t.Errorf("ports: block count = %d, want 1", c)
	}
}

// TestRenderFull_Minimal is the second full-string golden: no creds, no stores,
// no libs, MQTLS false, empty timezone, no restart and no ports -- so the
// restart, ports, environment, secrets and volumes blocks are all omitted.
func TestRenderFull_Minimal(t *testing.T) {
	in := Input{
		Docker: &spec.Docker{
			Name: "solmq",
		},
		Instance: Instance{Name: "solmq", Image: "img:1", AppYAML: "key: value\n", MQTLS: false, StatusScript: "echo ok\n"},
	}

	want := `services:
  solmq:
    image: img:1
    container_name: solmq
    labels:
      solace-connector/le-mode: standalone
      solace-connector/role: active
    configs:
      - source: solmq-app
        target: /app/external/spring/config/application.yml
      - source: solmq-status
        target: /app/external/.status-script
configs:
  solmq-app:
    content: |
      key: value
  solmq-status:
    content: |
      echo ok
`

	if got := Render(in); got != want {
		t.Errorf("Render mismatch:\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
}

// TestSecretsBranches covers the two secrets shapes directly: at least one
// secret (both the per-service secrets list and the top-level
// environment-provider block are emitted) and no secrets (neither block, nor
// any env_file line, is ever emitted).
func TestSecretsBranches(t *testing.T) {
	withSecret := Render(Input{
		Docker:   &spec.Docker{Name: "s"},
		Instance: Instance{Name: "s", Image: "img", AppYAML: "k: v\n", StatusScript: "echo ok\n", LeaderMode: spec.LeaderStandalone},
		Secrets:  []string{"MQ_USER"},
	})
	if !strings.Contains(withSecret, "    secrets:\n      - MQ_USER\n") {
		t.Errorf("per-service secrets list missing:\n%s", withSecret)
	}
	if !strings.Contains(withSecret, "secrets:\n  MQ_USER:\n    environment: MQ_USER\n") {
		t.Errorf("top-level environment-provider secrets block missing:\n%s", withSecret)
	}

	noSecrets := Render(Input{
		Docker:   &spec.Docker{Name: "s"},
		Instance: Instance{Name: "s", Image: "img", AppYAML: "k: v\n", StatusScript: "echo ok\n", LeaderMode: spec.LeaderStandalone},
	})
	if strings.Contains(noSecrets, "secrets:") {
		t.Errorf("secrets block should be absent when Secrets is empty:\n%s", noSecrets)
	}
	if strings.Contains(noSecrets, "env_file") {
		t.Error("env_file should never be emitted")
	}
}

// TestEnvironmentBranches covers the two single-key environment shapes.
func TestEnvironmentBranches(t *testing.T) {
	// TZ only (no MQTLS).
	tzOnly := Render(Input{
		Docker:   &spec.Docker{Name: "s"},
		Instance: Instance{Name: "s", Image: "img", Timezone: "UTC", AppYAML: "k: v\n", MQTLS: false, StatusScript: "echo ok\n", LeaderMode: spec.LeaderStandalone},
	})
	if !strings.Contains(tzOnly, "    environment:\n      TZ: UTC\n") {
		t.Errorf("TZ-only environment block missing:\n%s", tzOnly)
	}
	if strings.Contains(tzOnly, "JAVA_TOOL_OPTIONS") {
		t.Error("JAVA_TOOL_OPTIONS should be absent when MQTLS is false")
	}

	// MQTLS only (no timezone).
	mqOnly := Render(Input{
		Docker:   &spec.Docker{Name: "s"},
		Instance: Instance{Name: "s", Image: "img", AppYAML: "k: v\n", MQTLS: true, StatusScript: "echo ok\n", LeaderMode: spec.LeaderStandalone},
	})
	if !strings.Contains(mqOnly, `    environment:
      JAVA_TOOL_OPTIONS: "-Dcom.ibm.mq.cfg.useIBMCipherMappings=false"
`) {
		t.Errorf("MQTLS-only environment block missing:\n%s", mqOnly)
	}
	if strings.Contains(mqOnly, "TZ:") {
		t.Error("TZ should be absent when timezone is empty")
	}
}

// TestContentIndentationAndBlankLines checks that inlined application.yml lines
// are indented 6 spaces under content: and that blank lines stay truly empty.
func TestContentIndentationAndBlankLines(t *testing.T) {
	out := Render(Input{
		Docker:   &spec.Docker{Name: "s"},
		Instance: Instance{Name: "s", Image: "img", AppYAML: "root:\n  child: v\n\ntail: end\n", StatusScript: "echo ok\n", LeaderMode: spec.LeaderStandalone},
	})
	// A nested line keeps its own indentation on top of the 6-space block indent.
	if !strings.Contains(out, "      root:\n        child: v\n") {
		t.Errorf("nested content not indented as expected:\n%s", out)
	}
	// The blank line is preserved with no spaces.
	if !strings.Contains(out, "        child: v\n\n      tail: end\n") {
		t.Errorf("blank line not preserved as empty:\n%s", out)
	}
	// No content line carries trailing spaces.
	for _, ln := range strings.Split(out, "\n") {
		if ln != strings.TrimRight(ln, " ") {
			t.Errorf("line has trailing spaces: %q", ln)
		}
	}
}

// TestStoresOnlyAndLibsOnly covers the two partial-volumes shapes.
func TestStoresOnlyAndLibsOnly(t *testing.T) {
	storesOnly := Render(Input{
		Docker:   &spec.Docker{Name: "s"},
		Instance: Instance{Name: "s", Image: "img", AppYAML: "k: v\n", StatusScript: "echo ok\n", LeaderMode: spec.LeaderStandalone},
		Stores:   []Mount{{Source: "/h/a.jks", Target: "/c/a.jks"}},
	})
	if !strings.Contains(storesOnly, "    volumes:\n      - /h/a.jks:/c/a.jks:ro\n") {
		t.Errorf("stores-only volumes missing:\n%s", storesOnly)
	}

	libsOnly := Render(Input{
		Docker:   &spec.Docker{Name: "s"},
		Instance: Instance{Name: "s", Image: "img", AppYAML: "k: v\n", StatusScript: "echo ok\n", LeaderMode: spec.LeaderStandalone},
		Libs:     &Mount{Source: "/h/libs", Target: "/c/libs"},
	})
	if !strings.Contains(libsOnly, "    volumes:\n      - /h/libs:/c/libs:ro\n") {
		t.Errorf("libs-only volumes missing:\n%s", libsOnly)
	}
}

// TestSplitLinesNoTrailingNewline covers the branch where the app.yml does not
// end in a newline, so no trailing empty element is dropped. The status
// config now renders after the app config, so this checks the app content is
// present rather than that it is the final block.
func TestSplitLinesNoTrailingNewline(t *testing.T) {
	out := Render(Input{
		Docker:   &spec.Docker{Name: "s"},
		Instance: Instance{Name: "s", Image: "img", AppYAML: "only: line", StatusScript: "echo ok\n", LeaderMode: spec.LeaderStandalone},
	})
	if !strings.Contains(out, "    content: |\n      only: line\n") {
		t.Errorf("no-trailing-newline content not rendered as expected:\n%s", out)
	}
}

// TestLabelsPerMode covers the labels block for every leader-election mode:
// le-mode is always emitted, and role: active is added for standalone and
// active_active but withheld for active_standby, whose role is only knowable
// live from the actuator. An empty mode defaults to standalone.
func TestLabelsPerMode(t *testing.T) {
	tests := []struct {
		name     string
		mode     string
		wantMode string
		wantRole bool
	}{
		{"empty defaults to standalone", "", spec.LeaderStandalone, true},
		{"standalone", spec.LeaderStandalone, spec.LeaderStandalone, true},
		{"active_active", spec.LeaderActiveActive, spec.LeaderActiveActive, true},
		{"active_standby", spec.LeaderActiveStby, spec.LeaderActiveStby, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := Render(Input{
				Docker:   &spec.Docker{Name: "s"},
				Instance: Instance{Name: "s", Image: "img", AppYAML: "k: v\n", StatusScript: "echo ok\n", LeaderMode: tt.mode},
			})
			wantModeLine := "    labels:\n      " + spec.LabelModeKey + ": " + tt.wantMode + "\n"
			if !strings.Contains(out, wantModeLine) {
				t.Errorf("le-mode label missing or wrong:\n%s", out)
			}
			roleLine := "      " + spec.LabelRoleKey + ": " + spec.LabelRoleActive + "\n"
			if got := strings.Contains(out, roleLine); got != tt.wantRole {
				t.Errorf("role label present = %v, want %v:\n%s", got, tt.wantRole, out)
			}
		})
	}
}

// TestStatusScriptConfigSourceAndTarget covers the second configs entry: the
// service references <name>-status and mounts it at
// /app/external/.status-script.
func TestStatusScriptConfigSourceAndTarget(t *testing.T) {
	out := Render(Input{
		Docker:   &spec.Docker{Name: "s"},
		Instance: Instance{Name: "solmq", Image: "img", AppYAML: "k: v\n", StatusScript: "echo ok\n", LeaderMode: spec.LeaderStandalone},
	})
	if !strings.Contains(out, "      - source: solmq-status\n        target: /app/external/.status-script\n") {
		t.Errorf("status config source/target pair missing:\n%s", out)
	}
	if !strings.Contains(out, "  solmq-status:\n    content: |\n") {
		t.Errorf("top-level solmq-status config entry missing:\n%s", out)
	}
}

// TestStatusScriptContentIsEscaped checks the status script body is inlined
// under the status config's content: block, indented 6 spaces, with its blank
// line preserved as truly empty and its shell '$' doubled -- the same rules
// renderContentConfig applies to application.yml.
func TestStatusScriptContentIsEscaped(t *testing.T) {
	out := Render(Input{
		Docker:   &spec.Docker{Name: "s"},
		Instance: Instance{Name: "s", Image: "img", AppYAML: "k: v\n", StatusScript: statusScript1, LeaderMode: spec.LeaderStandalone},
	})
	want := "  s-status:\n    content: |\n      #!/bin/sh\n      STATUS_URL=\"http://localhost:8080/actuator/health\"\n\n      curl -fsS \"$$STATUS_URL\"\n"
	if !strings.Contains(out, want) {
		t.Errorf("status script content not inlined as expected:\n%s", out)
	}
}

// TestContentEscapesDollarsForCompose covers every '$' shape the two payloads
// actually carry. Docker Compose interpolates the whole document including
// configs content, so each must reach the file doubled: a bare $VAR and a
// ${VAR} would otherwise be substituted away, a ${VAR:-default} likewise, and a
// $( makes compose reject the document ("invalid interpolation format") because
// command substitution is not interpolation syntax at all. A '$' that is
// already doubled in the payload is data like any other and becomes four.
func TestContentEscapesDollarsForCompose(t *testing.T) {
	script := "P=$PORT\nQ=${VAR}\nR=${SPRING_CONFIG_LOCATION:-}\nS=$(printf %s x)\nT=$$\n"
	out := Render(Input{
		Docker:   &spec.Docker{Name: "s"},
		Instance: Instance{Name: "s", Image: "img", AppYAML: "k: $LITERAL\n", StatusScript: script, LeaderMode: spec.LeaderStandalone},
	})
	for _, want := range []string{
		"      k: $$LITERAL\n",
		"      P=$$PORT\n",
		"      Q=$${VAR}\n",
		"      R=$${SPRING_CONFIG_LOCATION:-}\n",
		"      S=$$(printf %s x)\n",
		"      T=$$$$\n",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing escaped line %q in:\n%s", want, out)
		}
	}
	// No lone '$' may survive anywhere in the document: compose reads one as the
	// start of an interpolation, so a single survivor is the whole bug back
	// again. Dropping every "$$" pair leaves exactly the unescaped ones behind.
	if bare := strings.ReplaceAll(out, "$$", ""); strings.Contains(bare, "$") {
		t.Errorf("unescaped '$' survives in:\n%s", out)
	}
}

// TestAppYAMLSecretPlaceholdersAreNotInterpolated pins the security half of the
// escape. The ${...} placeholders in application.yml are Spring's, resolved
// from the configtree import of /run/secrets. The CLI hands the compose child
// those same names as environment entries so the secrets: environment provider
// can mount them, which means an unescaped placeholder is not merely blanked --
// compose substitutes the real credential and the plaintext lands in the
// document. The placeholder must survive to the container intact.
func TestAppYAMLSecretPlaceholdersAreNotInterpolated(t *testing.T) {
	out := Render(Input{
		Docker: &spec.Docker{Name: "s"},
		Instance: Instance{
			Name:         "s",
			AppYAML:      "client-username: ${LOCAL_SOLACE_CLIENT_USERNAME}\nclient-password: ${LOCAL_SOLACE_CLIENT_PASSWORD}\n",
			StatusScript: "echo ok\n",
			LeaderMode:   spec.LeaderStandalone,
		},
		Secrets: []string{"LOCAL_SOLACE_CLIENT_USERNAME", "LOCAL_SOLACE_CLIENT_PASSWORD"},
	})
	if !strings.Contains(out, "      client-username: $${LOCAL_SOLACE_CLIENT_USERNAME}\n") {
		t.Errorf("username placeholder not escaped:\n%s", out)
	}
	if !strings.Contains(out, "      client-password: $${LOCAL_SOLACE_CLIENT_PASSWORD}\n") {
		t.Errorf("password placeholder not escaped:\n%s", out)
	}
	// The secrets blocks reference the same names and must stay unescaped:
	// those are compose's own, not Spring's.
	if !strings.Contains(out, "  LOCAL_SOLACE_CLIENT_USERNAME:\n    environment: LOCAL_SOLACE_CLIENT_USERNAME\n") {
		t.Errorf("secrets provider entry should not be escaped:\n%s", out)
	}
}
