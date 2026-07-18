// Package validate runs every schema and semantic check shared by all commands.
// It only collects problems; the caller decides reporting (validate prints all,
// config/deploy stop at the first). Errors are fatal; warnings are advisory.
package validate

import (
	"fmt"
	"net/url"
	"regexp"
	"sort"
	"strings"

	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-gen/internal/spec"
)

// MaxWorkflows is the hard cap on workflows per folder.
const MaxWorkflows = 20

// Issue is one validation finding (File "" means a global/cross-file problem).
type Issue struct {
	File string
	Msg  string
}

func (i Issue) String() string {
	if i.File == "" {
		return i.Msg
	}
	return i.File + ": " + i.Msg
}

// Context is the input to a validation run.
type Context struct {
	Workflows []spec.Workflow
	Defaults  *spec.Defaults
	Kube      *spec.Kubernetes // nil for `config`
	Deploy    bool             // deploy command: enables secret/store + kube checks

	// Env, when non-nil, is used to check that credentials.create source:env
	// variables are present (the CLI supplies os.LookupEnv; nil skips the check).
	Env func(string) (string, bool)
	// ValuesFileKeys is the set of keys available from a source:file values-file
	// (nil when not applicable).
	ValuesFileKeys map[string]bool
}

var (
	connNameRE = regexp.MustCompile(`^[^()\s,]+\(\d+\)(,[^()\s,]+\(\d+\))*$`)
	dns1123RE  = regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`)
	varRE      = regexp.MustCompile(`\$\{([A-Za-z_][A-Za-z0-9_]*)(:[^}]*)?\}`)
	hostRE     = regexp.MustCompile(`^[A-Za-z0-9._:-]+$`)
)

// Run executes all checks, returning fatal errors and non-fatal warnings.
func Run(ctx Context) (errs, warns []Issue) {
	add := func(file, format string, a ...any) { errs = append(errs, Issue{file, fmt.Sprintf(format, a...)}) }
	warn := func(file, format string, a ...any) { warns = append(warns, Issue{file, fmt.Sprintf(format, a...)}) }

	if ctx.Defaults == nil {
		ctx.Defaults = &spec.Defaults{}
	}
	d := ctx.Defaults

	if len(ctx.Workflows) > MaxWorkflows {
		add("", "too many workflows: %d (max %d)", len(ctx.Workflows), MaxWorkflows)
	}

	haveKeystore := d.TLS.Keystore != nil && d.TLS.Keystore.File != ""

	// Resolve conn-ref sides once for cross-workflow checks that key on the tuple.
	resolved := make([]spec.Workflow, len(ctx.Workflows))
	for i, wf := range ctx.Workflows {
		resolved[i] = wf
		resolved[i].Source = d.Resolve(wf.Source)
		resolved[i].Target = d.Resolve(wf.Target)
	}

	// Reusable connection definitions.
	checkConnections(add, d, haveKeystore)

	// Per-workflow structural + semantic checks (on the raw sides, so conn-ref rules apply).
	for _, wf := range ctx.Workflows {
		if !wf.SourceSet {
			add(wf.File, "missing 'source'")
		}
		if !wf.TargetSet {
			add(wf.File, "missing 'target'")
		}
		if wf.SourceSet {
			checkSide(add, warn, wf.File, "source", wf.Source, true, haveKeystore, d.Connections)
		}
		if wf.TargetSet {
			checkSide(add, warn, wf.File, "target", wf.Target, false, haveKeystore, d.Connections)
		}
	}

	// Cross-workflow (on resolved tuples): key-alias conflicts and duplicate sources.
	checkKeyAliasConflicts(add, resolved)
	checkDuplicateSources(warn, resolved)

	// Leader-election (standalone | active_active | active_standby).
	checkLeaderElection(add, d, haveKeystore)

	// Deploy-only checks.
	if ctx.Deploy && ctx.Kube != nil {
		checkKube(add, warn, ctx)
	}

	// Warn: TLS/mTLS in use but the stores Secret is not wired (deploy only).
	if ctx.Deploy {
		if usesTLS(resolved) && !storesWired(ctx.Kube) {
			warn("kubernetes.yaml", "a TLS/mTLS connection exists but secrets.stores is omitted; the store files will be missing at runtime")
		}
	}

	return errs, warns
}

// checkSide validates one workflow side. A conn-ref side is strict (only the
// destination); an inline side is checked as a full connection tuple. Both emit
// the EDA advisories for Solace topic-source / queue-destination.
func checkSide(add, warn func(string, string, ...any), file, which string, s spec.Side, isSource, haveKeystore bool, conns map[string]spec.Side) {
	if !s.HasSystem() {
		add(file, "%s must specify exactly one of 'solace:' or 'mq:'", which)
		return
	}
	if s.DestKind == "" {
		add(file, "%s (%s) must specify exactly one of 'queue:' or 'topic:'", which, s.System)
		return
	}
	if s.ConnRef != "" {
		if s.SetsConnFields() {
			add(file, "%s (%s) uses conn-ref %q, so it may set only queue/topic — move the connection details into connections.%s", which, s.System, s.ConnRef, s.ConnRef)
		}
		if conn, ok := conns[s.ConnRef]; !ok {
			add(file, "%s (%s) conn-ref %q is not defined under connections in defaults.yaml", which, s.System, s.ConnRef)
		} else if conn.System != s.System {
			add(file, "%s conn-ref %q is a %q connection but is referenced under %q", which, s.ConnRef, conn.System, s.System)
		}
		edaAdvisory(warn, file, s, isSource)
		return
	}
	checkTuple(add, file, which, s, haveKeystore)
	edaAdvisory(warn, file, s, isSource)
}

// checkTuple validates the connection fields of a side (no destination), shared
// by inline workflow sides, connection definitions, and the leader-election session.
func checkTuple(add func(string, string, ...any), file, label string, s spec.Side, haveKeystore bool) {
	switch s.System {
	case spec.SystemSolace:
		if s.Host == "" {
			add(file, "%s solace: missing 'host'", label)
		} else if !strings.HasPrefix(s.Host, "tcp://") && !strings.HasPrefix(s.Host, "tcps://") {
			add(file, "%s solace: host %q must start with tcp:// or tcps://", label, s.Host)
		}
		if s.MsgVPN == "" {
			add(file, "%s solace: missing 'msg-vpn'", label)
		}
		// user and pass can be optional
		// if s.ClientUser == "" {
		// 	add(file, "%s solace: missing 'client-username'", label)
		// }
		// if s.ClientPass == "" {
		// 	add(file, "%s solace: missing 'client-password'", label)
		// }
		if s.KeyAlias != "" {
			if !strings.HasPrefix(s.Host, "tcps://") {
				add(file, "%s solace: key-alias (mTLS) requires a tcps:// host", label)
			}
			if !haveKeystore {
				add(file, "%s solace: key-alias set but no keystore defined in defaults.yaml", label)
			}
		}
	case spec.SystemMQ:
		if s.ConnName == "" {
			add(file, "%s mq: missing 'conn-name'", label)
		} else if !connNameRE.MatchString(s.ConnName) {
			add(file, "%s mq: conn-name %q must be host(port)[,host(port)]", label, s.ConnName)
		}
		if s.QueueManager == "" {
			add(file, "%s mq: missing 'queue-manager'", label)
		}
		if s.Channel == "" {
			add(file, "%s mq: missing 'channel'", label)
		}
		// user and password can be optional for connection
		// if s.User == "" {
		// 	add(file, "%s mq: missing 'user'", label)
		// }
		// if s.Password == "" {
		// 	add(file, "%s mq: missing 'password'", label)
		// }
		if (s.Cipher != "" || s.KeyAlias != "") && !s.TLS {
			add(file, "%s mq: cipher/key-alias require 'tls: true'", label)
		}
		if s.KeyAlias != "" && !haveKeystore {
			add(file, "%s mq: key-alias set but no keystore defined in defaults.yaml", label)
		}
	}
}

// edaAdvisory emits the event-driven-architecture warnings (allowed, not errors):
// a Solace topic source is a non-durable subscription; a Solace queue destination
// is point-to-point.
func edaAdvisory(warn func(string, string, ...any), file string, s spec.Side, isSource bool) {
	if s.System != spec.SystemSolace {
		return
	}
	if isSource && s.DestKind == spec.DestTopic {
		warn(file, "source solace: topic %q as a source is a direct, non-durable subscription (at-most-once) — events published while this connector is down are lost. EDA guaranteed-delivery favors consuming from a queue subscribed to the topic, decoupling producer and consumer availability", s.Dest)
	}
	if !isSource && s.DestKind == spec.DestQueue {
		warn(file, "target solace: producing to queue %q is point-to-point and couples this flow to one endpoint. EDA favors publishing to a topic and letting the broker fan out to subscribed queues, so producers stay unaware of consumers (loose coupling)", s.Dest)
	}
}

// checkConnections validates each reusable connection definition (no destination,
// no nested conn-ref, complete tuple), in sorted order for stable messages.
func checkConnections(add func(string, string, ...any), d *spec.Defaults, haveKeystore bool) {
	names := make([]string, 0, len(d.Connections))
	for name := range d.Connections {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		c := d.Connections[name]
		label := "connections." + name
		if !c.HasSystem() {
			add("defaults.yaml", "%s must specify exactly one of 'solace:' or 'mq:'", label)
			continue
		}
		if c.ConnRef != "" {
			add("defaults.yaml", "%s must not use conn-ref (a connection is the referent, not a referrer)", label)
		}
		if c.DestKind != "" {
			add("defaults.yaml", "%s must not define queue/topic (a connection carries only connection details)", label)
		}
		checkTuple(add, "defaults.yaml", label, c, haveKeystore)
	}
}

// checkLeaderElection validates leader-election mode + (for active_*) queue and
// a Solace management session (conn-ref to a solace connection, or inline solace).
func checkLeaderElection(add func(string, string, ...any), d *spec.Defaults, haveKeystore bool) {
	le := d.LeaderElection
	if !le.Present || le.Mode == "" || le.Mode == spec.LeaderStandalone {
		return
	}
	if le.Mode != spec.LeaderActiveActive && le.Mode != spec.LeaderActiveStby {
		add("defaults.yaml", "leader-election mode %q is invalid (want standalone, active_active, or active_standby)", le.Mode)
		return
	}
	if le.Queue == "" {
		add("defaults.yaml", "leader-election mode %q requires a 'queue'", le.Mode)
	}
	switch {
	case le.ConnRef != "":
		if conn, ok := d.Connections[le.ConnRef]; !ok {
			add("defaults.yaml", "leader-election conn-ref %q is not defined under connections", le.ConnRef)
		} else if conn.System != spec.SystemSolace {
			add("defaults.yaml", "leader-election conn-ref %q must be a solace connection (got %q)", le.ConnRef, conn.System)
		}
	case le.Session != nil:
		if le.Session.System != spec.SystemSolace {
			add("defaults.yaml", "leader-election session must be a solace connection")
		} else {
			checkTuple(add, "defaults.yaml", "leader-election session", *le.Session, haveKeystore)
		}
	default:
		add("defaults.yaml", "leader-election mode %q requires a solace session (conn-ref or inline solace:)", le.Mode)
	}
}

// checkKeyAliasConflicts flags binders (same dedup tuple) whose sides declare
// different key-alias values — one binder cannot present two client certificates.
func checkKeyAliasConflicts(add func(string, string, ...any), wfs []spec.Workflow) {
	type seen struct {
		alias string
		file  string
	}
	byKey := map[string]seen{}
	visit := func(file string, s spec.Side) {
		if !s.HasSystem() || s.KeyAlias == "" {
			return
		}
		key := dedupKey(s)
		if prev, ok := byKey[key]; ok {
			if prev.alias != s.KeyAlias {
				add(file, "conflicting key-alias %q vs %q for the same binder (also in %s)", s.KeyAlias, prev.alias, prev.file)
			}
			return
		}
		byKey[key] = seen{s.KeyAlias, file}
	}
	for _, wf := range wfs {
		visit(wf.File, wf.Source)
		visit(wf.File, wf.Target)
	}
}

func checkDuplicateSources(warn func(string, string, ...any), wfs []spec.Workflow) {
	type seen struct{ file string }
	byKey := map[string]seen{}
	for _, wf := range wfs {
		s := wf.Source
		if !s.HasSystem() {
			continue
		}
		key := dedupKey(s) + "|" + s.Dest
		if prev, ok := byKey[key]; ok {
			warn(wf.File, "source %q is also consumed in %s (duplicate source on the same binder)", s.Dest, prev.file)
			continue
		}
		byKey[key] = seen{wf.File}
	}
}

func checkKube(add, warn func(string, string, ...any), ctx Context) {
	dep := ctx.Kube.Deployment
	if dep.Name == "" {
		add("kubernetes.yaml", "deployment.name is required")
	} else if !isDNS1123(dep.Name) {
		add("kubernetes.yaml", "deployment.name %q is not a valid DNS-1123 label", dep.Name)
	}
	if dep.Namespace == "" {
		add("kubernetes.yaml", "deployment.namespace is required")
	} else if !isDNS1123(dep.Namespace) {
		add("kubernetes.yaml", "deployment.namespace %q is not a valid DNS-1123 label", dep.Namespace)
	}
	if dep.Image == "" {
		add("kubernetes.yaml", "deployment.image is required")
	}

	// standalone requires a single replica.
	mode := ctx.Defaults.LeaderElection.Mode
	if (mode == "" || mode == spec.LeaderStandalone) && dep.Replicas > 1 {
		add("kubernetes.yaml", "leader-election standalone requires replicas: 1 (got %d)", dep.Replicas)
	}

	// Secret wiring checks.
	if c := ctx.Kube.Secrets.Credentials; c != nil && c.Create != nil {
		switch c.Create.Source {
		case spec.SourceEnv:
			if len(c.Create.Variables) == 0 {
				add("kubernetes.yaml", "secrets.credentials.create source: env requires a non-empty 'variables' list")
			}
			if ctx.Env != nil {
				for _, v := range c.Create.Variables {
					if _, ok := ctx.Env(v); !ok {
						add("kubernetes.yaml", "credentials variable %q is not set in the environment", v)
					}
				}
			}
		case spec.SourceFile:
			if c.Create.ValuesFile == "" {
				add("kubernetes.yaml", "secrets.credentials.create source: file requires 'values-file'")
			}
		default:
			add("kubernetes.yaml", "secrets.credentials.create source must be %q or %q", spec.SourceEnv, spec.SourceFile)
		}
	}
	if s := ctx.Kube.Secrets.Stores; s != nil && s.Create != nil {
		if ctx.Defaults.TLS.Truststore == nil {
			add("kubernetes.yaml", "secrets.stores.create requires defaults.yaml tls.truststore")
		}
	}

	checkSyslog(add, warn, ctx.Kube)
	checkLibs(add, ctx.Kube)

	// Warn: ${VAR} used in config but not supplied by the credentials Secret.
	checkUnsuppliedVars(warn, ctx)
}

// checkSyslog validates the optional logging.syslog block.
func checkSyslog(add, warn func(string, string, ...any), k *spec.Kubernetes) {
	if k.Logging == nil || k.Logging.Syslog == nil {
		return
	}
	s := k.Logging.Syslog
	if s.Host == "" {
		add("kubernetes.yaml", "logging.syslog.host is required")
	} else if !hostRE.MatchString(s.Host) {
		add("kubernetes.yaml", "logging.syslog.host %q may only contain letters, digits, '.', ':', '_' and '-'", s.Host)
	}
	if s.Port < 1 || s.Port > 65535 {
		add("kubernetes.yaml", "logging.syslog.port must be 1-65535 (got %d)", s.Port)
	}
	switch s.Protocol {
	case spec.SyslogUDP:
	case spec.SyslogTCP:
		warn("kubernetes.yaml", "logging.syslog protocol tcp requires the logstash-logback-encoder jar on the connector classpath (provide it via libs)")
	default:
		add("kubernetes.yaml", "logging.syslog.protocol must be %q or %q (got %q)", spec.SyslogUDP, spec.SyslogTCP, s.Protocol)
	}
}

// checkLibs validates the optional libs block (exactly one provisioning mode).
func checkLibs(add func(string, string, ...any), k *spec.Kubernetes) {
	lb := k.Libs
	if lb == nil {
		return
	}
	modes := 0
	for _, set := range []bool{lb.PVC != nil, lb.Download != nil} {
		if set {
			modes++
		}
	}
	if modes != 1 {
		add("kubernetes.yaml", "libs must set exactly one of 'pvc' or 'download' (got %d)", modes)
		return
	}
	switch {
	case lb.PVC != nil:
		p := lb.PVC
		if (p.Create != nil) == (p.Existing != "") {
			add("kubernetes.yaml", "libs.pvc must set exactly one of 'create' or 'existing'")
			return
		}
		if p.Existing != "" && !isDNS1123(p.Existing) {
			add("kubernetes.yaml", "libs.pvc.existing %q is not a valid DNS-1123 label", p.Existing)
		}
		if c := p.Create; c != nil {
			if !isDNS1123(c.Name) {
				add("kubernetes.yaml", "libs.pvc.create.name %q is not a valid DNS-1123 label", c.Name)
			}
			if c.NFS.Server == "" || c.NFS.Path == "" {
				add("kubernetes.yaml", "libs.pvc.create requires nfs.server and nfs.path")
			}
		}
	case lb.Download != nil:
		d := lb.Download
		if len(d.URLs) == 0 {
			add("kubernetes.yaml", "libs.download requires a non-empty 'urls' list")
		}
		for _, u := range d.URLs {
			if !safeLibsURL(u) {
				add("kubernetes.yaml", "libs.download url %q must be http(s) and contain no spaces, quotes, or control characters", u)
			}
		}
		if d.PVC != "" && !isDNS1123(d.PVC) {
			add("kubernetes.yaml", "libs.download.pvc %q is not a valid DNS-1123 label", d.PVC)
		}
	}
}

// safeLibsURL accepts only http(s) URLs that cannot escape the single-quoted
// wget argument in the generated initContainer command (defense-in-depth, §3).
func safeLibsURL(raw string) bool {
	u, err := url.Parse(raw)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return false
	}
	return safeShellChars(raw)
}

func safeShellChars(s string) bool {
	for _, r := range s {
		if r <= 0x20 || r == 0x7f || r == '\'' || r == '"' || r == '\\' || r == '`' || r == '$' {
			return false
		}
	}
	return true
}

// checkUnsuppliedVars warns when a ${VAR} placeholder used by the workflows or
// defaults has no matching key in the credentials Secret (create only).
func checkUnsuppliedVars(warn func(string, string, ...any), ctx Context) {
	c := ctx.Kube.Secrets.Credentials
	if c == nil || c.Create == nil {
		return
	}
	supplied := map[string]bool{}
	switch c.Create.Source {
	case spec.SourceEnv:
		for _, v := range c.Create.Variables {
			supplied[v] = true
		}
	case spec.SourceFile:
		for k := range ctx.ValuesFileKeys {
			supplied[k] = true
		}
	}
	seen := map[string]bool{}
	for v := range collectVars(ctx) {
		if !supplied[v] && !seen[v] {
			seen[v] = true
			warn("kubernetes.yaml", "${%s} is used but not supplied by the credentials Secret", v)
		}
	}
}

// collectVars gathers ${VAR} names referenced by workflow secret fields and the
// defaults store passwords (placeholders that must be resolved at runtime).
func collectVars(ctx Context) map[string]bool {
	out := map[string]bool{}
	scan := func(s string) {
		for _, m := range varRE.FindAllStringSubmatch(s, -1) {
			out[m[1]] = true
		}
	}
	for _, wf := range ctx.Workflows {
		for _, s := range []spec.Side{ctx.Defaults.Resolve(wf.Source), ctx.Defaults.Resolve(wf.Target)} {
			scan(s.ClientPass)
			scan(s.Password)
		}
	}
	if d := ctx.Defaults; d != nil {
		if d.TLS.Truststore != nil {
			scan(d.TLS.Truststore.Password)
		}
		if d.TLS.Keystore != nil {
			scan(d.TLS.Keystore.Password)
		}
		for _, u := range d.Security.Users {
			scan(u.Password)
		}
		if le := d.LeaderElection; le.ConnRef != "" {
			if c, ok := d.Connections[le.ConnRef]; ok {
				scan(c.ClientPass)
			}
		} else if le.Session != nil {
			scan(le.Session.ClientPass)
		}
	}
	return out
}

func usesTLS(wfs []spec.Workflow) bool {
	for _, wf := range wfs {
		for _, s := range []spec.Side{wf.Source, wf.Target} {
			if s.System == spec.SystemSolace && strings.HasPrefix(s.Host, "tcps://") {
				return true
			}
			if s.System == spec.SystemMQ && s.TLS {
				return true
			}
		}
	}
	return false
}

func storesWired(k *spec.Kubernetes) bool {
	if k == nil || k.Secrets.Stores == nil {
		return false
	}
	return k.Secrets.Stores.Create != nil || k.Secrets.Stores.Existing != ""
}

func isDNS1123(s string) bool { return len(s) <= 63 && dns1123RE.MatchString(s) }

func dedupKey(s spec.Side) string {
	if s.System == spec.SystemSolace {
		return "S|" + s.Host + "|" + s.MsgVPN + "|" + s.ClientUser
	}
	return "M|" + s.ConnName + "|" + s.QueueManager + "|" + s.Channel + "|" + s.User
}
