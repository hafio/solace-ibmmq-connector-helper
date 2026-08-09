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
// instance with creds (env_file), a store mount, a libs mount, MQTLS, ports and
// a timezone.
func TestRenderFull_WithEverything(t *testing.T) {
	in := Input{
		Docker: &spec.Docker{
			Image:    "solace/solace-pubsub-connector-ibmmq:2.13.0",
			Name:     "solmq-connector",
			Restart:  "unless-stopped",
			Ports:    []spec.Port{{Host: 8090, Container: 8090}, {Host: 8080, Container: 8091}},
			Timezone: "Asia/Singapore",
		},
		Instances: []Instance{
			{Name: "solmq-connector", AppYAML: appYAML1, MQTLS: true},
		},
		EnvFile: "solmq-connector.env",
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
    env_file:
      - solmq-connector.env
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
`

	if got := Render(in); got != want {
		t.Errorf("Render mismatch:\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
}

// TestRenderFull_Minimal is the second full-string golden: no creds, no stores,
// no libs, MQTLS false, empty timezone, no restart and no ports -- so the
// restart, ports, environment, env_file and volumes blocks are all omitted.
func TestRenderFull_Minimal(t *testing.T) {
	in := Input{
		Docker: &spec.Docker{
			Image: "img:1",
			Name:  "solmq",
		},
		Instances: []Instance{
			{Name: "solmq", AppYAML: "key: value\n", MQTLS: false},
		},
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

// TestRenderMultiInstance checks that two shards produce two service blocks and
// two top-level configs in order, and that per-instance fields (env_file, ports,
// MQTLS) are repeated / conditional correctly.
func TestRenderMultiInstance(t *testing.T) {
	in := Input{
		Docker: &spec.Docker{
			Image:    "img:2",
			Name:     "solmq",
			Restart:  "always",
			Ports:    []spec.Port{{Host: 8090, Container: 8090}, {Host: 9090, Container: 9090}},
			Timezone: "UTC",
		},
		Instances: []Instance{
			{Name: "solmq-1", AppYAML: "a: 1\n", MQTLS: true},
			{Name: "solmq-2", AppYAML: "b: 2\n", MQTLS: false},
		},
		EnvFile: "creds.env",
	}

	out := Render(in)

	// Two service headers (indent 2, not the "-app" config keys).
	if c := strings.Count(out, "  solmq-1:\n"); c != 1 {
		t.Errorf("service header solmq-1 count = %d, want 1", c)
	}
	if c := strings.Count(out, "  solmq-2:\n"); c != 1 {
		t.Errorf("service header solmq-2 count = %d, want 1", c)
	}
	// Two top-level config keys.
	if c := strings.Count(out, "  solmq-1-app:\n"); c != 1 {
		t.Errorf("config key solmq-1-app count = %d, want 1", c)
	}
	if c := strings.Count(out, "  solmq-2-app:\n"); c != 1 {
		t.Errorf("config key solmq-2-app count = %d, want 1", c)
	}
	// Exactly two inlined configs and two service-level config sources.
	if c := strings.Count(out, "content: |"); c != 2 {
		t.Errorf("content blocks = %d, want 2", c)
	}
	if c := strings.Count(out, "- source: "); c != 2 {
		t.Errorf("service config sources = %d, want 2", c)
	}
	// Ordering: instance 1 before instance 2 in both maps.
	if strings.Index(out, "solmq-1:") >= strings.Index(out, "solmq-2:") {
		t.Error("instance order not preserved in services")
	}
	if strings.Index(out, "solmq-1-app:") >= strings.Index(out, "solmq-2-app:") {
		t.Error("instance order not preserved in configs")
	}
	// Ports repeated per service.
	if c := strings.Count(out, `- "8090:8090"`); c != 2 {
		t.Errorf(`port 8090 count = %d, want 2`, c)
	}
	if c := strings.Count(out, `- "9090:9090"`); c != 2 {
		t.Errorf(`port 9090 count = %d, want 2`, c)
	}
	// env_file and TZ repeated per service; JAVA_TOOL_OPTIONS only on shard 1.
	if c := strings.Count(out, "- creds.env"); c != 2 {
		t.Errorf("env_file entries = %d, want 2", c)
	}
	if c := strings.Count(out, "TZ: UTC"); c != 2 {
		t.Errorf("TZ entries = %d, want 2", c)
	}
	if c := strings.Count(out, "JAVA_TOOL_OPTIONS:"); c != 1 {
		t.Errorf("JAVA_TOOL_OPTIONS entries = %d, want 1", c)
	}
}

// TestEnvironmentBranches covers the two single-key environment shapes.
func TestEnvironmentBranches(t *testing.T) {
	// TZ only (no MQTLS).
	tzOnly := Render(Input{
		Docker:    &spec.Docker{Image: "img", Name: "s", Timezone: "UTC"},
		Instances: []Instance{{Name: "s", AppYAML: "k: v\n", MQTLS: false}},
	})
	if !strings.Contains(tzOnly, "    environment:\n      TZ: UTC\n") {
		t.Errorf("TZ-only environment block missing:\n%s", tzOnly)
	}
	if strings.Contains(tzOnly, "JAVA_TOOL_OPTIONS") {
		t.Error("JAVA_TOOL_OPTIONS should be absent when MQTLS is false")
	}

	// MQTLS only (no timezone).
	mqOnly := Render(Input{
		Docker:    &spec.Docker{Image: "img", Name: "s"},
		Instances: []Instance{{Name: "s", AppYAML: "k: v\n", MQTLS: true}},
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
		Docker:    &spec.Docker{Image: "img", Name: "s"},
		Instances: []Instance{{Name: "s", AppYAML: "root:\n  child: v\n\ntail: end\n"}},
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
		Docker:    &spec.Docker{Image: "img", Name: "s"},
		Instances: []Instance{{Name: "s", AppYAML: "k: v\n"}},
		Stores:    []Mount{{Source: "/h/a.jks", Target: "/c/a.jks"}},
	})
	if !strings.Contains(storesOnly, "    volumes:\n      - /h/a.jks:/c/a.jks:ro\n") {
		t.Errorf("stores-only volumes missing:\n%s", storesOnly)
	}

	libsOnly := Render(Input{
		Docker:    &spec.Docker{Image: "img", Name: "s"},
		Instances: []Instance{{Name: "s", AppYAML: "k: v\n"}},
		Libs:      &Mount{Source: "/h/libs", Target: "/c/libs"},
	})
	if !strings.Contains(libsOnly, "    volumes:\n      - /h/libs:/c/libs:ro\n") {
		t.Errorf("libs-only volumes missing:\n%s", libsOnly)
	}
}

// TestRenderNoInstances covers the guard that omits the top-level configs map
// when there is nothing to inline.
func TestRenderNoInstances(t *testing.T) {
	if got := Render(Input{Docker: &spec.Docker{}}); got != "services:\n" {
		t.Errorf("empty Render = %q, want %q", got, "services:\n")
	}
}

// TestSplitLinesNoTrailingNewline covers the branch where the app.yml does not
// end in a newline, so no trailing empty element is dropped.
func TestSplitLinesNoTrailingNewline(t *testing.T) {
	out := Render(Input{
		Docker:    &spec.Docker{Image: "img", Name: "s"},
		Instances: []Instance{{Name: "s", AppYAML: "only: line"}},
	})
	if !strings.HasSuffix(out, "    content: |\n      only: line\n") {
		t.Errorf("no-trailing-newline content not rendered as expected:\n%s", out)
	}
}
