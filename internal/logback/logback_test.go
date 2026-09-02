package logback

import (
	"strings"
	"testing"

	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/spec"
)

// TestXMLPicksTheAppenderTheProtocolNames pins the one decision this package
// makes. The two configs are not interchangeable: udp uses Logback's built-in
// appender and needs nothing on the classpath, while tcp needs the
// logstash-logback-encoder jar -- which is why validate warns about tcp and why
// picking the wrong one fails at runtime rather than at generate time.
func TestXMLPicksTheAppenderTheProtocolNames(t *testing.T) {
	udp := XML(spec.SyslogUDP)
	if !strings.Contains(udp, "ch.qos.logback.classic.net.SyslogAppender") {
		t.Errorf("udp should use the built-in SyslogAppender, got:\n%s", udp)
	}
	if strings.Contains(udp, "LogstashTcpSocketAppender") {
		t.Error("udp must not reference the logstash appender, which needs a jar it does not require")
	}

	tcp := XML(spec.SyslogTCP)
	if !strings.Contains(tcp, "net.logstash.logback.appender.LogstashTcpSocketAppender") {
		t.Errorf("tcp should use the logstash appender, got:\n%s", tcp)
	}
	if strings.Contains(tcp, "ch.qos.logback.classic.net.SyslogAppender") {
		t.Error("tcp must not fall back to the udp appender")
	}
}

// TestXMLDefaultsToUDP covers the unset and unknown cases together: udp is the
// safe choice because it needs nothing added to the classpath. validate rejects
// an unknown protocol long before this, so this is the guard rather than a
// path operators reach.
func TestXMLDefaultsToUDP(t *testing.T) {
	for _, protocol := range []string{"", "nonsense"} {
		if XML(protocol) != XML(spec.SyslogUDP) {
			t.Errorf("protocol %q should render the udp config", protocol)
		}
	}
}

// TestBothConfigsReadTheSameThreeProperties pins the contract between this file
// and every renderer: the XML takes host, port and appname from springProperty
// bindings, which is why all three platforms set the same LOGGING_SYSLOG_* env
// vars instead of templating values into the file.
func TestBothConfigsReadTheSameThreeProperties(t *testing.T) {
	for _, protocol := range []string{spec.SyslogUDP, spec.SyslogTCP} {
		out := XML(protocol)
		for _, want := range []string{
			`source="logging.syslog.host"`,
			`source="logging.syslog.port"`,
			`source="logging.syslog.appname"`,
		} {
			if !strings.Contains(out, want) {
				t.Errorf("%s config is missing %s, so the env var that feeds it would do nothing", protocol, want)
			}
		}
		if !strings.Contains(out, `<appender-ref ref="CONSOLE"/>`) {
			t.Errorf("%s config must keep the console appender: syslog is in addition to stdout, not instead of it", protocol)
		}
	}
}

// TestContainerPathAndFileNameAgree keeps the two constants from drifting: the
// platforms that mount a file use FileName for the disk name and ContainerPath
// for the target, and a mismatch would mount the file somewhere the connector
// does not read.
func TestContainerPathAndFileNameAgree(t *testing.T) {
	if !strings.HasSuffix(ContainerPath, "/"+FileName) {
		t.Errorf("ContainerPath %q must end in /%s", ContainerPath, FileName)
	}
}
