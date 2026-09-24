package validate

import (
	"strings"
	"testing"

	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/spec"
)

// javaCtx is a docker run that is otherwise clean, carrying j as its
// java-options: block -- so any error it returns is the block's.
func javaCtx(j *spec.JavaOptions) Context {
	return Context{Workflows: wfOK(), Defaults: &spec.Defaults{}, Image: imageOK(), Docker: dockerOK(), CheckDocker: true, JavaOptions: j}
}

// TestJavaOptionsAcceptsJVMSyntax pins what the gate must let through. JVM
// options are full of characters SafeToken refuses -- spaces between options,
// the '*' and ':' of -Xlog, the ',' and '=' of an agent string, the '%p' of an
// error-file path, an @argfile -- and none of them ever reaches a shell. The
// newlines of a literal block are collapsed before the check, so they are not
// control characters either.
func TestJavaOptionsAcceptsJVMSyntax(t *testing.T) {
	for _, v := range []string{
		"-Xms512m -Xmx1g -XX:+UseG1GC",
		"-Xlog:gc*:file=/tmp/gc.log:time,uptime",
		"-XX:ErrorFile=/tmp/hs_err_%p.log -XX:+HeapDumpOnOutOfMemoryError",
		"-agentlib:jdwp=transport=dt_socket,server=y,suspend=n,address=*:5005",
		"--add-opens=java.base/java.lang=ALL-UNNAMED @/app/external/jvm.args",
		"-Xms512m\n-Xmx1g\n\n  -Dx=y ",
		// A '#' inside one option is not a comment and stays allowed.
		"-Dmarker=a#b",
	} {
		for _, j := range []*spec.JavaOptions{{Tool: v}, {JDK: v}} {
			if errs, _ := Run(javaCtx(j)); len(errs) != 0 {
				t.Errorf("java-options %+v should pass, got %v", *j, errs)
			}
		}
	}
}

// TestJavaOptionsRejectsWhatThePlatformsReadDifferently covers every character
// the gate refuses. Each would mean one thing in a compose file, another in a
// kubernetes manifest and a third in a systemd unit, so the error names the
// key, the variable it feeds and the character, and says what to do instead.
func TestJavaOptionsRejectsWhatThePlatformsReadDifferently(t *testing.T) {
	cases := []struct {
		name, value string
		want        []string
	}{
		// What an unset ${VAR} with no default leaves behind after expansion.
		{"a leftover variable reference", "-Xmx${HEAP}", []string{"contains '$'", "unset variable", ":default"}},
		{"a double quote", `-Dgreeting="hi there"`, []string{`contains '"'`, "no quoting is needed"}},
		{"a single quote", "-Dgreeting='hi'", []string{`contains '\''`, "no quoting is needed"}},
		{"a backslash", `-Dpath=C:\temp`, []string{`contains '\\'`, "no quoting is needed"}},
		{"a backtick", "-Dx=`id`", []string{"contains '`'", "no quoting is needed"}},
		{"a control character", "-Dx=a\x07b", []string{"control character", "remove it"}},
		// What a trailing comment inside a >- block leaves in the value.
		{"a comment inside a block", "-Xms512m # initial heap\n-Xmx1g", []string{"reads as a comment", "above the java-options: key"}},
		{"a block that opens with a comment", "# heap\n-Xmx1g", []string{"reads as a comment"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			for _, sub := range []struct {
				key, envVar string
				j           *spec.JavaOptions
			}{
				{"java-options.tool", "JAVA_TOOL_OPTIONS", &spec.JavaOptions{Tool: c.value}},
				{"java-options.jdk", "JDK_JAVA_OPTIONS", &spec.JavaOptions{JDK: c.value}},
			} {
				errs, _ := Run(javaCtx(sub.j))
				if len(errs) != 1 {
					t.Fatalf("%s = %q: want exactly one error, got %v", sub.key, c.value, errs)
				}
				msg := errs[0].String()
				for _, w := range append([]string{sub.key, sub.envVar}, c.want...) {
					if !strings.Contains(msg, w) {
						t.Errorf("%s = %q: error %q should contain %q", sub.key, c.value, msg, w)
					}
				}
			}
		})
	}
}

// TestJavaOptionsCheckedOnlyWhenAPlatformIsInPlay pins the gate to the runs
// that render a container: `generate config` writes application.yml alone and
// never sets a JVM options variable, so it has nothing to refuse.
func TestJavaOptionsCheckedOnlyWhenAPlatformIsInPlay(t *testing.T) {
	bad := &spec.JavaOptions{Tool: `-Dx="y"`}
	errs, _ := Run(Context{Workflows: wfOK(), Defaults: &spec.Defaults{}, JavaOptions: bad})
	if hasErr(errs, "java-options") {
		t.Errorf("a config-only run must not check java-options, got %v", errs)
	}
	for name, ctx := range map[string]Context{
		"kubernetes": {Workflows: wfOK(), Defaults: &spec.Defaults{}, Image: imageOK(), CheckKubernetes: true, JavaOptions: bad},
		"docker":     {Workflows: wfOK(), Defaults: &spec.Defaults{}, Image: imageOK(), CheckDocker: true, JavaOptions: bad},
		"podman":     {Workflows: wfOK(), Defaults: &spec.Defaults{}, Image: imageOK(), CheckPodman: true, JavaOptions: bad},
	} {
		if errs, _ := Run(ctx); !hasErr(errs, "java-options.tool (JAVA_TOOL_OPTIONS)") {
			t.Errorf("%s: java-options must be checked, got %v", name, errs)
		}
	}
}
