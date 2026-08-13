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

// TestRenderFull_WithEverything is the primary full-string golden: a single
// instance with creds (rendered as a per-service secrets list plus a
// top-level environment-provider secrets block -- never an env_file), a store
// mount, a libs mount, MQTLS, ports and a timezone.
func TestRenderFull_WithEverything(t *testing.T) {
	in := Input{
		Docker: &spec.Docker{
			Image:    "solace/solace-pubsub-connector-ibmmq:2.13.0",
			Name:     "solmq-connector",
			Restart:  "unless-stopped",
			Ports:    []spec.Port{{Host: 8090, Container: 8090}, {Host: 8080, Container: 8091}},
			Timezone: "Asia/Singapore",
		},
		Instance: Instance{Name: "solmq-connector", AppYAML: appYAML1, MQTLS: true},
		Secrets:  []string{"SOLACE_CLIENT_USERNAME", "SOLACE_CLIENT_PASSWORD", "TRUSTSTORE_PASSWORD"},
		Stores: []Mount{
			{Source: "/abs/certs/truststore.jks", Target: "/app/external/classpath/truststores/truststore.jks"},
		},
		Libs: &Mount{Source: "/abs/libs", Target: "/app/external/libs"},
	}

	want := `services:
  solmq-connector:
    image: solace/solace-pubsub-connector-ibmmq:2.13.0
    container_name: solmq-connector
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
			Image: "img:1",
			Name:  "solmq",
		},
		Instance: Instance{Name: "solmq", AppYAML: "key: value\n", MQTLS: false},
	}

	want := `services:
  solmq:
    image: img:1
    container_name: solmq
    configs:
      - source: solmq-app
        target: /app/external/spring/config/application.yml
configs:
  solmq-app:
    content: |
      key: value
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
		Docker:   &spec.Docker{Image: "img", Name: "s"},
		Instance: Instance{Name: "s", AppYAML: "k: v\n"},
		Secrets:  []string{"MQ_USER"},
	})
	if !strings.Contains(withSecret, "    secrets:\n      - MQ_USER\n") {
		t.Errorf("per-service secrets list missing:\n%s", withSecret)
	}
	if !strings.Contains(withSecret, "secrets:\n  MQ_USER:\n    environment: MQ_USER\n") {
		t.Errorf("top-level environment-provider secrets block missing:\n%s", withSecret)
	}

	noSecrets := Render(Input{
		Docker:   &spec.Docker{Image: "img", Name: "s"},
		Instance: Instance{Name: "s", AppYAML: "k: v\n"},
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
		Docker:   &spec.Docker{Image: "img", Name: "s", Timezone: "UTC"},
		Instance: Instance{Name: "s", AppYAML: "k: v\n", MQTLS: false},
	})
	if !strings.Contains(tzOnly, "    environment:\n      TZ: UTC\n") {
		t.Errorf("TZ-only environment block missing:\n%s", tzOnly)
	}
	if strings.Contains(tzOnly, "JAVA_TOOL_OPTIONS") {
		t.Error("JAVA_TOOL_OPTIONS should be absent when MQTLS is false")
	}

	// MQTLS only (no timezone).
	mqOnly := Render(Input{
		Docker:   &spec.Docker{Image: "img", Name: "s"},
		Instance: Instance{Name: "s", AppYAML: "k: v\n", MQTLS: true},
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
		Docker:   &spec.Docker{Image: "img", Name: "s"},
		Instance: Instance{Name: "s", AppYAML: "root:\n  child: v\n\ntail: end\n"},
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
		Docker:   &spec.Docker{Image: "img", Name: "s"},
		Instance: Instance{Name: "s", AppYAML: "k: v\n"},
		Stores:   []Mount{{Source: "/h/a.jks", Target: "/c/a.jks"}},
	})
	if !strings.Contains(storesOnly, "    volumes:\n      - /h/a.jks:/c/a.jks:ro\n") {
		t.Errorf("stores-only volumes missing:\n%s", storesOnly)
	}

	libsOnly := Render(Input{
		Docker:   &spec.Docker{Image: "img", Name: "s"},
		Instance: Instance{Name: "s", AppYAML: "k: v\n"},
		Libs:     &Mount{Source: "/h/libs", Target: "/c/libs"},
	})
	if !strings.Contains(libsOnly, "    volumes:\n      - /h/libs:/c/libs:ro\n") {
		t.Errorf("libs-only volumes missing:\n%s", libsOnly)
	}
}

// TestSplitLinesNoTrailingNewline covers the branch where the app.yml does not
// end in a newline, so no trailing empty element is dropped.
func TestSplitLinesNoTrailingNewline(t *testing.T) {
	out := Render(Input{
		Docker:   &spec.Docker{Image: "img", Name: "s"},
		Instance: Instance{Name: "s", AppYAML: "only: line"},
	})
	if !strings.HasSuffix(out, "    content: |\n      only: line\n") {
		t.Errorf("no-trailing-newline content not rendered as expected:\n%s", out)
	}
}
