package spec

// Syslog is the top-level logging.syslog block of env.yaml: where the connector
// ships its log lines, in addition to the console.
//
// It lives beside logging.level rather than under a deploy section because it
// describes the connector, not a platform -- one declaration serves kubernetes,
// docker and podman alike. It used to be kubernetes.logging.syslog, which meant
// docker and podman users had no way to ask for it at all.
//
// There is no enabled field: writing the block is what turns syslog on, so
// every consumer nil-guards instead of reading a flag that could disagree with
// the block being present.
type Syslog struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Protocol string `yaml:"protocol"` // SyslogUDP (default) | SyslogTCP
}

// Syslog protocols. udp is Logback's built-in RFC 3164 appender; tcp needs the
// logstash-logback-encoder jar on the classpath, which is why validate warns
// about it rather than assuming it is there.
const (
	SyslogUDP = "udp"
	SyslogTCP = "tcp"
)

// Logging is the retired kubernetes.logging block.
//
// The type stays because Kubernetes.Logging still parses it: ParseEnv decodes
// non-strict, so deleting the field would make a stale config silently lose
// syslog instead of failing. validate rejects a non-nil value and names the new
// top-level key.
type Logging struct {
	Syslog *Syslog `yaml:"syslog"`
}
