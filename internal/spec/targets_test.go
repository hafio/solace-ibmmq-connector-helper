package spec

import "testing"

func TestParseEnvPortsValid(t *testing.T) {
	tests := []struct {
		name          string
		entry         string // YAML sequence item for docker.ports
		wantHost      int
		wantContainer int
		wantString    string
	}{
		{"bare int", "8090", 8090, 8090, "8090:8090"},
		{"host:container", "8080:8090", 8080, 8090, "8080:8090"},
		{"padded host:container (TrimSpace)", `" 8080 : 8090 "`, 8080, 8090, "8080:8090"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := []byte("docker:\n  ports:\n    - " + tt.entry + "\n")
			e, err := ParseEnv(data)
			if err != nil {
				t.Fatalf("ParseEnv: %v", err)
			}
			if len(e.Docker.Ports) != 1 {
				t.Fatalf("ports = %+v", e.Docker.Ports)
			}
			p := e.Docker.Ports[0]
			if p.Host != tt.wantHost || p.Container != tt.wantContainer {
				t.Fatalf("port = %+v, want host=%d container=%d", p, tt.wantHost, tt.wantContainer)
			}
			if got := p.String(); got != tt.wantString {
				t.Errorf("String() = %q, want %q", got, tt.wantString)
			}
		})
	}
}

func TestParseEnvPortsInvalid(t *testing.T) {
	tests := []struct {
		name    string
		entry   string // YAML sequence item for docker.ports
		wantErr string
	}{
		{"non-integer", "abc", `env.yaml: ports entry "abc" must be an integer or "host:container"`},
		{"more than one colon", "1:2:3", `env.yaml: ports entry "1:2:3" must be "host:container" (exactly one colon)`},
		{"non-integer host and container", "a:b", `env.yaml: ports entry "a:b" must be "host:container" with integer ports`},
		{"mapping node", "{a: 1}", `env.yaml: ports entry must be an integer or "host:container", got a !!map`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := []byte("docker:\n  ports:\n    - " + tt.entry + "\n")
			_, err := ParseEnv(data)
			if err == nil {
				t.Fatal("expected error")
			}
			if err.Error() != tt.wantErr {
				t.Errorf("err = %q, want %q", err.Error(), tt.wantErr)
			}
		})
	}
}

func TestApplyDockerDefaultsFillsMissing(t *testing.T) {
	data := []byte(`
docker:
  image: myimg
`)
	e, err := ParseEnv(data)
	if err != nil {
		t.Fatal(err)
	}
	d := e.Docker
	if d == nil {
		t.Fatal("docker section not parsed")
	}
	if d.Command != DefaultDockerCommand {
		t.Errorf("command = %q want %q", d.Command, DefaultDockerCommand)
	}
	if d.Name != DefaultConnectorName {
		t.Errorf("name = %q want %q", d.Name, DefaultConnectorName)
	}
	if d.Restart != DefaultRestart {
		t.Errorf("restart = %q want %q", d.Restart, DefaultRestart)
	}
	want := Port{Host: DefaultMgmtPort, Container: DefaultMgmtPort}
	if len(d.Ports) != 1 || d.Ports[0] != want {
		t.Errorf("ports = %+v want [%+v]", d.Ports, want)
	}
	if d.Stores != nil || d.Libs != nil {
		t.Errorf("stores/libs should stay nil when absent: %+v / %+v", d.Stores, d.Libs)
	}
}

func TestApplyDockerDefaultsOverrideWins(t *testing.T) {
	data := []byte(`
docker:
  command: docker --context foo
  image: myimg
  name: custom-name
  restart: always
  ports:
    - 9000
    - "8000:8001"
`)
	e, err := ParseEnv(data)
	if err != nil {
		t.Fatal(err)
	}
	d := e.Docker
	if d.Command != "docker --context foo" {
		t.Errorf("command = %q", d.Command)
	}
	if d.Name != "custom-name" {
		t.Errorf("name = %q", d.Name)
	}
	if d.Restart != "always" {
		t.Errorf("restart = %q", d.Restart)
	}
	want := []Port{{Host: 9000, Container: 9000}, {Host: 8000, Container: 8001}}
	if len(d.Ports) != len(want) || d.Ports[0] != want[0] || d.Ports[1] != want[1] {
		t.Errorf("ports = %+v want %+v", d.Ports, want)
	}
}

// applyPodmanDefaults must allocate a Quadlet even when the podman: section
// omits quadlet: entirely -- cmd/solmq-conn-util/main.go dereferences it
// unconditionally on the podman deploy path, so a nil Quadlet here would
// later panic.
func TestApplyPodmanDefaultsFillsMissing(t *testing.T) {
	data := []byte(`
podman:
  image: myimg
`)
	e, err := ParseEnv(data)
	if err != nil {
		t.Fatal(err)
	}
	p := e.Podman
	if p == nil {
		t.Fatal("podman section not parsed")
	}
	if p.Command != DefaultPodmanCommand {
		t.Errorf("command = %q want %q", p.Command, DefaultPodmanCommand)
	}
	if p.Mode != PodmanModeRun {
		t.Errorf("mode = %q want %q", p.Mode, PodmanModeRun)
	}
	if p.Name != DefaultConnectorName {
		t.Errorf("name = %q want %q", p.Name, DefaultConnectorName)
	}
	if p.Restart != DefaultRestart {
		t.Errorf("restart = %q want %q", p.Restart, DefaultRestart)
	}
	want := Port{Host: DefaultMgmtPort, Container: DefaultMgmtPort}
	if len(p.Ports) != 1 || p.Ports[0] != want {
		t.Errorf("ports = %+v want [%+v]", p.Ports, want)
	}
	if p.Quadlet == nil {
		t.Fatal("quadlet should default to non-nil")
	}
	if p.Quadlet.Scope != QuadletScopeAuto {
		t.Errorf("quadlet scope = %q want %q", p.Quadlet.Scope, QuadletScopeAuto)
	}
	if p.Quadlet.Dir != "" {
		t.Errorf("quadlet dir = %q want empty", p.Quadlet.Dir)
	}
}

func TestApplyPodmanDefaultsOverrideWins(t *testing.T) {
	data := []byte(`
podman:
  command: podman --context foo
  image: myimg
  mode: quadlet
  name: custom-name
  restart: always
  ports:
    - 9000
  quadlet:
    scope: system
    dir: /custom/dir
`)
	e, err := ParseEnv(data)
	if err != nil {
		t.Fatal(err)
	}
	p := e.Podman
	if p.Command != "podman --context foo" {
		t.Errorf("command = %q", p.Command)
	}
	if p.Mode != PodmanModeQuadlet {
		t.Errorf("mode = %q want %q", p.Mode, PodmanModeQuadlet)
	}
	if p.Name != "custom-name" || p.Restart != "always" {
		t.Errorf("name/restart = %q/%q", p.Name, p.Restart)
	}
	want := Port{Host: 9000, Container: 9000}
	if len(p.Ports) != 1 || p.Ports[0] != want {
		t.Errorf("ports = %+v want [%+v]", p.Ports, want)
	}
	if p.Quadlet == nil || p.Quadlet.Scope != QuadletScopeSystem || p.Quadlet.Dir != "/custom/dir" {
		t.Errorf("quadlet = %+v", p.Quadlet)
	}
}

func TestApplyMountDefaultsFillsMissing(t *testing.T) {
	data := []byte(`
docker:
  image: myimg
  stores: {}
  libs:
    dir: /host/libs
`)
	e, err := ParseEnv(data)
	if err != nil {
		t.Fatal(err)
	}
	d := e.Docker
	if d.Stores == nil || d.Stores.MountPath != DefaultStoresMountPath {
		t.Errorf("stores = %+v want mount-path %q", d.Stores, DefaultStoresMountPath)
	}
	if d.Libs == nil || d.Libs.Dir != "/host/libs" || d.Libs.MountPath != DefaultLibsMountPath {
		t.Errorf("libs = %+v want dir /host/libs mount-path %q", d.Libs, DefaultLibsMountPath)
	}
}

func TestEffectiveManagementPort(t *testing.T) {
	tests := []struct {
		name string
		d    *Defaults
		want int
	}{
		{"nil receiver", nil, DefaultMgmtPort},
		{"unset port", &Defaults{}, DefaultMgmtPort},
		{"set port", &Defaults{Management: Management{Port: 9091}}, 9091},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.d.EffectiveManagementPort(); got != tt.want {
				t.Errorf("EffectiveManagementPort() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestApplyMountDefaultsOverrideWins(t *testing.T) {
	data := []byte(`
docker:
  image: myimg
  stores:
    mount-path: /custom/store
  libs:
    dir: /host/libs
    mount-path: /custom/libs
`)
	e, err := ParseEnv(data)
	if err != nil {
		t.Fatal(err)
	}
	d := e.Docker
	if d.Stores == nil || d.Stores.MountPath != "/custom/store" {
		t.Errorf("stores = %+v", d.Stores)
	}
	if d.Libs == nil || d.Libs.MountPath != "/custom/libs" {
		t.Errorf("libs = %+v", d.Libs)
	}
}
