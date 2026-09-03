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
	// Unlike ports, project-name IS defaulted: the compose project is not an
	// exposure decision, and leaving it empty would hand the grouping back to
	// compose's basename-of-the-directory guess.
	if d.ProjectName != DefaultComposeProject {
		t.Errorf("project-name = %q want %q", d.ProjectName, DefaultComposeProject)
	}
	if d.Restart != DefaultRestart {
		t.Errorf("restart = %q want %q", d.Restart, DefaultRestart)
	}
	// Publishing a port is an exposure decision, so an omitted ports: stays
	// omitted rather than defaulting to the management port.
	if len(d.Ports) != 0 {
		t.Errorf("ports = %+v, want none when ports: is omitted", d.Ports)
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
  project-name: custom-project
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
	// project-name is its own key, not derived from name: a custom name must not
	// drag the project along with it.
	if d.ProjectName != "custom-project" {
		t.Errorf("project-name = %q", d.ProjectName)
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
	// mode: is a removed key and must not be defaulted -- validate rejects any
	// non-empty value, so a default here would fail every section.
	if p.Mode != "" {
		t.Errorf("mode = %q, want it left empty", p.Mode)
	}
	if p.Name != DefaultConnectorName {
		t.Errorf("name = %q want %q", p.Name, DefaultConnectorName)
	}
	if p.Restart != DefaultRestart {
		t.Errorf("restart = %q want %q", p.Restart, DefaultRestart)
	}
	if len(p.Ports) != 0 {
		t.Errorf("ports = %+v, want none when ports: is omitted", p.Ports)
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

// TestRemovedMountKeysDecodeButAreNotDefaulted covers what is left of the mount
// wiring now that both container-side paths are fixed by the image. Nothing is
// filled in: the removed keys only have to survive decoding, so validate can name
// them in an error rather than yaml dropping them in silence.
//
// The mount-path assertion is the load-bearing one. Were it still defaulted,
// every libs: block would carry a non-empty value and trip validate's rejection
// for something the operator never wrote -- the same trap as podman.mode.
func TestRemovedMountKeysDecodeButAreNotDefaulted(t *testing.T) {
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
	if d.Libs == nil || d.Libs.Dir != "/host/libs" {
		t.Errorf("libs = %+v want dir /host/libs", d.Libs)
	}
	if d.Libs != nil && d.Libs.MountPath != "" {
		t.Errorf("libs mount-path = %q, want it left empty", d.Libs.MountPath)
	}
	if d.Stores == nil {
		t.Error("a present stores: must still decode non-nil so validate can reject it")
	}
}

// The kubernetes service port follows the effective management port (not the
// bare DefaultMgmtPort constant), since that is the only port the connector
// actually listens on once management.port is overridden. docker and podman
// publish nothing they were not asked to: an omitted ports: stays omitted on
// both, whatever management.port says.
func TestPortDefaultsFollowManagementPort(t *testing.T) {
	data := []byte(`
management:
  port: 9091
docker:
  image: myimg
podman:
  image: myimg
kubernetes:
  deployment: {name: c, namespace: ns, image: img}
`)
	e, err := ParseEnv(data)
	if err != nil {
		t.Fatal(err)
	}
	want := Port{Host: 9091, Container: 9091}
	if len(e.Docker.Ports) != 0 {
		t.Errorf("docker ports = %+v, want none when ports: is omitted", e.Docker.Ports)
	}
	if len(e.Podman.Ports) != 0 {
		t.Errorf("podman ports = %+v, want none when ports: is omitted", e.Podman.Ports)
	}
	if e.Kubernetes.Service.Port != want {
		t.Errorf("kubernetes service.port = %+v want %+v", e.Kubernetes.Service.Port, want)
	}
}

// With no management.port set, the kubernetes service port falls back to
// DefaultMgmtPort (8090); docker and podman still publish nothing.
func TestPortDefaultsFallBackWhenManagementPortUnset(t *testing.T) {
	data := []byte(`
docker:
  image: myimg
podman:
  image: myimg
kubernetes:
  deployment: {name: c, namespace: ns, image: img}
`)
	e, err := ParseEnv(data)
	if err != nil {
		t.Fatal(err)
	}
	want := Port{Host: DefaultMgmtPort, Container: DefaultMgmtPort}
	if len(e.Docker.Ports) != 0 {
		t.Errorf("docker ports = %+v, want none when ports: is omitted", e.Docker.Ports)
	}
	if len(e.Podman.Ports) != 0 {
		t.Errorf("podman ports = %+v, want none when ports: is omitted", e.Podman.Ports)
	}
	if e.Kubernetes.Service.Port != want {
		t.Errorf("kubernetes service.port = %+v want %+v", e.Kubernetes.Service.Port, want)
	}
}

func TestKubernetesServicePortAcceptsBareAndHostContainerForms(t *testing.T) {
	tests := []struct {
		name  string
		entry string
		want  Port
	}{
		{"bare int", "8090", Port{Host: 8090, Container: 8090}},
		{"host:container", "8080:8090", Port{Host: 8080, Container: 8090}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := []byte("kubernetes:\n  deployment: {name: c, namespace: ns, image: img}\n  service: {port: " + tt.entry + "}\n")
			e, err := ParseEnv(data)
			if err != nil {
				t.Fatalf("ParseEnv: %v", err)
			}
			if e.Kubernetes.Service.Port != tt.want {
				t.Errorf("service.port = %+v want %+v", e.Kubernetes.Service.Port, tt.want)
			}
		})
	}
}

func TestKubernetesServicePortRejectsInvalidForms(t *testing.T) {
	tests := []struct {
		name    string
		entry   string
		wantErr string
	}{
		{"multi-colon", "1:2:3", `env.yaml: ports entry "1:2:3" must be "host:container" (exactly one colon)`},
		{"mapping node", "{a: 1}", `env.yaml: ports entry must be an integer or "host:container", got a !!map`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := []byte("kubernetes:\n  deployment: {name: c, namespace: ns, image: img}\n  service: {port: " + tt.entry + "}\n")
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

// TestWrittenLibsMountPathSurvivesDecoding is the other half of the removed-key
// contract: a mount-path an operator actually wrote must reach validate intact so
// it can be rejected by name. Decoding is non-strict, so dropping the field from
// the struct would silently discard the value instead -- and this is the case
// where that matters most, since a custom path used to be honoured and produced a
// container whose jars the JVM never saw.
func TestWrittenLibsMountPathSurvivesDecoding(t *testing.T) {
	data := []byte(`
docker:
  image: myimg
  libs:
    dir: /host/libs
    mount-path: /custom/libs
`)
	e, err := ParseEnv(data)
	if err != nil {
		t.Fatal(err)
	}
	d := e.Docker
	if d.Libs == nil || d.Libs.MountPath != "/custom/libs" {
		t.Errorf("libs = %+v, want the written mount-path preserved for validate", d.Libs)
	}
}
