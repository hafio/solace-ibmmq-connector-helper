package spec

import (
	"strings"
	"testing"
)

// ParseEnv's doc comment promises: "An empty/absent file yields a zero Env
// with default discovery (dir ".", "*")." Pin that contract exactly.
func TestParseEnvEmpty(t *testing.T) {
	e, err := ParseEnv([]byte(""))
	if err != nil {
		t.Fatal(err)
	}
	if e.Workflows == nil || e.Workflows.Dir != "." || e.Workflows.FilePattern != "*" {
		t.Errorf("workflows = %+v", e.Workflows)
	}
	if e.Kubernetes != nil || e.Docker != nil || e.Podman != nil {
		t.Errorf("sections should stay nil when absent: k=%+v d=%+v p=%+v", e.Kubernetes, e.Docker, e.Podman)
	}
	if e.Management.Present || e.Security.Present || e.LeaderElection.Present || e.TLS.Truststore != nil {
		t.Errorf("defaults should be zero-valued: %+v", e.Defaults)
	}
}

func TestWorkflowsFromRawDefaultWhenAbsent(t *testing.T) {
	e, err := ParseEnv([]byte("docker:\n  image: x\n"))
	if err != nil {
		t.Fatal(err)
	}
	if e.Workflows.Dir != "." || e.Workflows.FilePattern != "*" {
		t.Errorf("workflows = %+v", e.Workflows)
	}
}

func TestWorkflowsFromRawDirOverride(t *testing.T) {
	e, err := ParseEnv([]byte("workflows:\n  dir: /custom/dir\n"))
	if err != nil {
		t.Fatal(err)
	}
	if e.Workflows.Dir != "/custom/dir" {
		t.Errorf("dir = %q want /custom/dir", e.Workflows.Dir)
	}
	if e.Workflows.FilePattern != "*" {
		t.Errorf("file pattern should stay default, got %q", e.Workflows.FilePattern)
	}
}

func TestWorkflowsFromRawFilePatternOverride(t *testing.T) {
	e, err := ParseEnv([]byte("workflows:\n  file_pattern: \"*.yaml\"\n"))
	if err != nil {
		t.Fatal(err)
	}
	if e.Workflows.FilePattern != "*.yaml" {
		t.Errorf("file pattern = %q want *.yaml", e.Workflows.FilePattern)
	}
	if e.Workflows.Dir != "." {
		t.Errorf("dir should stay default, got %q", e.Workflows.Dir)
	}
}

// TestParseEnvUnknownKeyIgnored documents current behavior: ParseEnv decodes
// with plain yaml.Unmarshal (no KnownFields(true)), so an unrecognized
// top-level key is silently dropped rather than erroring. This pins that
// behavior; it does not endorse non-strict parsing as desirable.
func TestParseEnvUnknownKeyIgnored(t *testing.T) {
	e, err := ParseEnv([]byte("bogus-key: 1\ndocker:\n  image: x\n"))
	if err != nil {
		t.Fatalf("unknown top-level key should not error: %v", err)
	}
	if e.Docker == nil || e.Docker.Image != "x" {
		t.Errorf("docker section should still parse: %+v", e.Docker)
	}
}

func TestParseEnvWrongScalarTypeErrors(t *testing.T) {
	_, err := ParseEnv([]byte("management:\n  port: not-a-number\n"))
	if err == nil {
		t.Fatal("expected error for non-integer management.port")
	}
	// yaml.v3 TypeError messages embed a line number, so pin only the stable
	// "cannot unmarshal" fragment rather than the full string.
	if !strings.Contains(err.Error(), "cannot unmarshal") {
		t.Errorf("error should name the type mismatch: %v", err)
	}
}
