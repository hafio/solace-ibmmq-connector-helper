package spec

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// Env is the fully parsed env.yaml: the shared connector defaults (embedded)
// plus the optional per-target deploy sections and the workflow-discovery
// config. It replaces the old split of defaults.yaml + kubernetes.yaml.
type Env struct {
	Defaults              // connections, tls, logging.level, management, security, leader-election, solace-defaults
	Workflows  *Workflows // discovery config (never nil after ParseEnv)
	Kubernetes *Kubernetes
	Docker     *Docker
	Podman     *Podman
}

// Workflows drives which files in a folder are treated as workflows.
type Workflows struct {
	Dir         string // relative (to env.yaml) or absolute; default "."
	FilePattern string // only leading/middle/trailing '*' wildcard; default "*"
}

// rawEnv inlines the existing rawDefaults field set (so every current top-level
// key parses unchanged) and adds the new sections.
type rawEnv struct {
	rawDefaults `yaml:",inline"`
	Workflows   *rawWorkflows `yaml:"workflows"`
	Kubernetes  *Kubernetes   `yaml:"kubernetes"`
	Docker      *Docker       `yaml:"docker"`
	Podman      *Podman       `yaml:"podman"`
}

type rawWorkflows struct {
	Dir         string `yaml:"dir"`
	FilePattern string `yaml:"file_pattern"`
}

// ParseEnv decodes env.yaml into an Env, applying every section's defaults. An
// empty/absent file yields a zero Env with default discovery (dir ".", "*").
func ParseEnv(data []byte) (*Env, error) {
	var raw rawEnv
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("env.yaml: %v", err)
	}
	e := &Env{
		Defaults:  *defaultsFromRaw(raw.rawDefaults),
		Workflows: workflowsFromRaw(raw.Workflows),
	}
	// The effective management port drives the kubernetes service default, so an
	// unset service.port targets the port the connector actually listens on
	// rather than a bare constant. docker/podman publish nothing unless the
	// operator lists ports, so neither needs it.
	if raw.Kubernetes != nil {
		applyKubeDefaults(raw.Kubernetes, e.EffectiveManagementPort())
		e.Kubernetes = raw.Kubernetes
	}
	if raw.Docker != nil {
		applyDockerDefaults(raw.Docker)
		e.Docker = raw.Docker
	}
	if raw.Podman != nil {
		applyPodmanDefaults(raw.Podman)
		e.Podman = raw.Podman
	}
	return e, nil
}

// workflowsFromRaw always returns a non-nil config so callers have a dir and
// pattern to scan with even when the section is omitted.
func workflowsFromRaw(r *rawWorkflows) *Workflows {
	w := &Workflows{Dir: ".", FilePattern: "*"}
	if r != nil {
		if r.Dir != "" {
			w.Dir = r.Dir
		}
		if r.FilePattern != "" {
			w.FilePattern = r.FilePattern
		}
	}
	return w
}
