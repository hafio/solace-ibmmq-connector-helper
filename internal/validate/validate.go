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

	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/spec"
)

// MaxWorkflowsPerInstance is the connector runtime's cap on workflow IDs
// (0..19) per application.yml. Folders with more workflows are sharded across
// that many connector instances by the gen layer, so this is a per-instance
// size — not a hard folder cap.
const MaxWorkflowsPerInstance = 20

// fileEnv is the single config file label on every issue (everything now lives
// in env.yaml; the message body carries the section path).
const fileEnv = "env.yaml"

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
	Kube      *spec.Kubernetes // nil when the kubernetes section is absent/unused
	Docker    *spec.Docker     // nil when the docker section is absent/unused
	Podman    *spec.Podman     // nil when the podman section is absent/unused

	// Target gates: enable the deploy-grade checks for the selected target
	// (secret/store wiring, image required, command safe). `validate` sets all
	// three so it lints every section present in env.yaml.
	Deploy      bool // kubernetes checks
	CheckDocker bool // docker checks
	CheckPodman bool // podman checks

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
	haveKeystore := d.TLS.Keystore != nil && d.TLS.Keystore.File != ""

	// Resolve conn-ref sides once for cross-workflow checks that key on the tuple.
	resolved := make([]spec.Workflow, len(ctx.Workflows))
	for i, wf := range ctx.Workflows {
		resolved[i] = wf
		resolved[i].Source = d.Resolve(wf.Source)
		resolved[i].Target = d.Resolve(wf.Target)
	}

	// Reusable connection definitions + per-workflow structural checks.
	checkConnections(add, d, haveKeystore)
	checkWorkflowSides(add, warn, ctx.Workflows, haveKeystore, d.Connections)

	// Cross-workflow (on resolved tuples): key-alias conflicts and duplicate sources.
	checkKeyAliasConflicts(add, resolved)
	checkDuplicateSources(warn, resolved)

	// Leader-election (standalone | active_active | active_standby).
	checkLeaderElection(add, d, haveKeystore)

	// Per-target deploy-grade checks.
	checkTargets(add, warn, ctx, resolved)

	return errs, warns
}

// checkWorkflowSides runs the per-workflow structural + semantic checks on the
// raw sides (so conn-ref rules apply before resolution).
func checkWorkflowSides(add, warn func(string, string, ...any), wfs []spec.Workflow, haveKeystore bool, conns map[string]spec.Side) {
	for _, wf := range wfs {
		if !wf.SourceSet {
			add(wf.File, "missing 'source'")
		}
		if !wf.TargetSet {
			add(wf.File, "missing 'target'")
		}
		if wf.SourceSet {
			checkSide(add, warn, wf.File, "source", wf.Source, true, haveKeystore, conns)
		}
		if wf.TargetSet {
			checkSide(add, warn, wf.File, "target", wf.Target, false, haveKeystore, conns)
		}
	}
}

// checkTargets runs the checks for whichever deploy targets are enabled. Each
// gate is independent so `config` skips them all, a per-target command lints
// only its section, and `validate` enables every section present.
func checkTargets(add, warn func(string, string, ...any), ctx Context, resolved []spec.Workflow) {
	if ctx.Deploy && ctx.Kube != nil {
		checkKube(add, warn, ctx)
		// Warn: TLS/mTLS in use but the stores Secret is not wired.
		if usesTLS(resolved) && !storesWired(ctx.Kube) {
			warn(fileEnv, "a TLS/mTLS connection exists but secrets.stores is omitted; the store files will be missing at runtime")
		}
	}
	if ctx.CheckDocker && ctx.Docker != nil {
		checkDocker(add, ctx)
		// Warn: TLS/mTLS in use but no host stores are bind-mounted (the container
		// image ships none, so application.yml would point at absent files).
		if usesTLS(resolved) && ctx.Docker.Stores == nil {
			warn(fileEnv, "a TLS/mTLS connection exists but docker.stores is omitted; the store files will be missing at runtime")
		}
	}
	if ctx.CheckPodman && ctx.Podman != nil {
		checkPodman(add, ctx)
		if usesTLS(resolved) && ctx.Podman.Stores == nil {
			warn(fileEnv, "a TLS/mTLS connection exists but podman.stores is omitted; the store files will be missing at runtime")
		}
	}
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
			add(file, "%s (%s) conn-ref %q is not defined under connections", which, s.System, s.ConnRef)
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
				add(file, "%s solace: key-alias set but no keystore defined", label)
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
			add(file, "%s mq: key-alias set but no keystore defined", label)
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
			add(fileEnv, "%s must specify exactly one of 'solace:' or 'mq:'", label)
			continue
		}
		if c.ConnRef != "" {
			add(fileEnv, "%s must not use conn-ref (a connection is the referent, not a referrer)", label)
		}
		if c.DestKind != "" {
			add(fileEnv, "%s must not define queue/topic (a connection carries only connection details)", label)
		}
		checkTuple(add, fileEnv, label, c, haveKeystore)
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
		add(fileEnv, "leader-election mode %q is invalid (want standalone, active_active, or active_standby)", le.Mode)
		return
	}
	if le.Queue == "" {
		add(fileEnv, "leader-election mode %q requires a 'queue'", le.Mode)
	}
	switch {
	case le.ConnRef != "":
		if conn, ok := d.Connections[le.ConnRef]; !ok {
			add(fileEnv, "leader-election conn-ref %q is not defined under connections", le.ConnRef)
		} else if conn.System != spec.SystemSolace {
			add(fileEnv, "leader-election conn-ref %q must be a solace connection (got %q)", le.ConnRef, conn.System)
		}
	case le.Session != nil:
		if le.Session.System != spec.SystemSolace {
			add(fileEnv, "leader-election session must be a solace connection")
		} else {
			checkTuple(add, fileEnv, "leader-election session", *le.Session, haveKeystore)
		}
	default:
		add(fileEnv, "leader-election mode %q requires a solace session (conn-ref or inline solace:)", le.Mode)
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
		add(fileEnv, "deployment.name is required")
	} else if !isDNS1123(dep.Name) {
		add(fileEnv, "deployment.name %q is not a valid DNS-1123 label", dep.Name)
	} else if instances := shardCount(len(ctx.Workflows)); instances > 1 {
		// With >1 instance the name is suffixed -<n>; the ConfigMap (<name>-<n>-config)
		// is the longest derived name and must stay within the 63-char DNS-1123 limit.
		if longest := fmt.Sprintf("%s-%d-config", dep.Name, instances); len(longest) > 63 {
			add(fileEnv, "deployment.name %q is too long: %d instances generate names up to %q which exceeds the 63-char DNS-1123 limit", dep.Name, instances, longest)
		}
	}
	if dep.Namespace == "" {
		add(fileEnv, "deployment.namespace is required")
	} else if !isDNS1123(dep.Namespace) {
		add(fileEnv, "deployment.namespace %q is not a valid DNS-1123 label", dep.Namespace)
	}
	if dep.Image == "" {
		add(fileEnv, "deployment.image is required")
	}

	// standalone requires a single replica.
	mode := ctx.Defaults.LeaderElection.Mode
	if (mode == "" || mode == spec.LeaderStandalone) && dep.Replicas > 1 {
		add(fileEnv, "leader-election standalone requires replicas: 1 (got %d)", dep.Replicas)
	}

	// Secret wiring checks.
	if c := ctx.Kube.Secrets.Credentials; c != nil && c.Create != nil {
		checkCredentialsCreate(add, ctx.Env, "kubernetes", c.Create)
	}
	if s := ctx.Kube.Secrets.Stores; s != nil && s.Create != nil {
		if ctx.Defaults.TLS.Truststore == nil {
			add(fileEnv, "kubernetes.secrets.stores.create requires tls.truststore")
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
		add(fileEnv, "logging.syslog.host is required")
	} else if !hostRE.MatchString(s.Host) {
		add(fileEnv, "logging.syslog.host %q may only contain letters, digits, '.', ':', '_' and '-'", s.Host)
	}
	if s.Port < 1 || s.Port > 65535 {
		add(fileEnv, "logging.syslog.port must be 1-65535 (got %d)", s.Port)
	}
	switch s.Protocol {
	case spec.SyslogUDP:
	case spec.SyslogTCP:
		warn(fileEnv, "logging.syslog protocol tcp requires the logstash-logback-encoder jar on the connector classpath (provide it via libs)")
	default:
		add(fileEnv, "logging.syslog.protocol must be %q or %q (got %q)", spec.SyslogUDP, spec.SyslogTCP, s.Protocol)
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
		add(fileEnv, "libs must set exactly one of 'pvc' or 'download' (got %d)", modes)
		return
	}
	switch {
	case lb.PVC != nil:
		p := lb.PVC
		if (p.Create != nil) == (p.Existing != "") {
			add(fileEnv, "libs.pvc must set exactly one of 'create' or 'existing'")
			return
		}
		if p.Existing != "" && !isDNS1123(p.Existing) {
			add(fileEnv, "libs.pvc.existing %q is not a valid DNS-1123 label", p.Existing)
		}
		if c := p.Create; c != nil {
			if !isDNS1123(c.Name) {
				add(fileEnv, "libs.pvc.create.name %q is not a valid DNS-1123 label", c.Name)
			}
			if c.NFS.Server == "" || c.NFS.Path == "" {
				add(fileEnv, "libs.pvc.create requires nfs.server and nfs.path")
			}
		}
	case lb.Download != nil:
		d := lb.Download
		if len(d.URLs) == 0 {
			add(fileEnv, "libs.download requires a non-empty 'urls' list")
		}
		for _, u := range d.URLs {
			if !safeLibsURL(u) {
				add(fileEnv, "libs.download url %q must be http(s) and contain no spaces, quotes, or control characters", u)
			}
		}
		if d.PVC != "" && !isDNS1123(d.PVC) {
			add(fileEnv, "libs.download.pvc %q is not a valid DNS-1123 label", d.PVC)
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

// safeShellChars rejects whitespace, control chars, 0x7f, quotes, backslash,
// backtick, and '$' -- the characters that could break out of the single-quoted
// wget argument safeLibsURL guards. It deliberately does NOT reject other shell
// metacharacters (';', '&', '?', '#', ...): those are harmless inside single
// quotes and legitimately appear in download URLs (query strings, fragments), so
// widening this set would reject valid URLs. Argv tokens use the stricter
// SafeToken, which layers shellMeta on top.
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
			warn(fileEnv, "${%s} is used but not supplied by the credentials Secret", v)
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

// shardCount is the number of connector instances n workflows split into
// (ceil(n / MaxWorkflowsPerInstance)); 0 for an empty folder.
func shardCount(n int) int {
	return (n + MaxWorkflowsPerInstance - 1) / MaxWorkflowsPerInstance
}

func dedupKey(s spec.Side) string {
	if s.System == spec.SystemSolace {
		return "S|" + s.Host + "|" + s.MsgVPN + "|" + s.ClientUser
	}
	return "M|" + s.ConnName + "|" + s.QueueManager + "|" + s.Channel + "|" + s.User
}

// checkCredentialsCreate validates a credentials Secret's create block for any
// target section (kubernetes/docker/podman). env, when non-nil, confirms each
// source:env variable is actually present in the process environment.
func checkCredentialsCreate(add func(string, string, ...any), env func(string) (string, bool), section string, c *spec.CredCreate) {
	switch c.Source {
	case spec.SourceEnv:
		if len(c.Variables) == 0 {
			add(fileEnv, "%s.secrets.credentials.create source: env requires a non-empty 'variables' list", section)
		}
		if env != nil {
			for _, v := range c.Variables {
				if _, ok := env(v); !ok {
					add(fileEnv, "%s.secrets.credentials variable %q is not set in the environment", section, v)
				}
			}
		}
	case spec.SourceFile:
		if c.ValuesFile == "" {
			add(fileEnv, "%s.secrets.credentials.create source: file requires 'values-file'", section)
		}
	default:
		add(fileEnv, "%s.secrets.credentials.create source must be %q or %q", section, spec.SourceEnv, spec.SourceFile)
	}
}

// checkPorts flags any host or container port outside the valid 1-65535 range
// (the "host:container" shape itself is enforced at parse in spec.Port).
func checkPorts(add func(string, string, ...any), section string, ports []spec.Port) {
	for _, p := range ports {
		if p.Host < 1 || p.Host > 65535 {
			add(fileEnv, "%s.ports host port %d must be 1-65535", section, p.Host)
		}
		if p.Container < 1 || p.Container > 65535 {
			add(fileEnv, "%s.ports container port %d must be 1-65535", section, p.Container)
		}
	}
}

// checkCommand splits a deploy command into whitespace-separated tokens and
// gates each against SafeToken. It is a validation-grade check; the runner does
// the authoritative quote-aware parse. A command with spaces in its path is
// intentionally unsupported (the safe charset rejects space/quote/backslash).
func checkCommand(add func(string, string, ...any), field, cmd string) {
	tokens := strings.Fields(cmd)
	if len(tokens) == 0 {
		add(fileEnv, "%s must not be empty", field)
		return
	}
	for _, tok := range tokens {
		if !SafeToken(tok) {
			add(fileEnv, "%s token %q contains an unsafe character (no spaces, quotes, backslash, control chars, or shell metacharacters)", field, tok)
		}
	}
}

// checkContainerTarget runs the checks common to the docker and podman sections:
// a DNS-1123 name (it flows into filesystem paths and a systemctl unit token), a
// safe non-empty command, a required image, valid ports, credentials wiring, and
// host-provided stores/libs paths.
func checkContainerTarget(add func(string, string, ...any), ctx Context, section, name, command, image string, ports []spec.Port, creds *spec.CredentialsSecret, stores *spec.StoresMount, libs *spec.LibsMount) {
	if !isDNS1123(name) {
		add(fileEnv, "%s.name %q is not a valid DNS-1123 label", section, name)
	}
	checkCommand(add, section+".command", command)
	if image == "" {
		add(fileEnv, "%s.image is required", section)
	}
	checkPorts(add, section, ports)
	if creds != nil && creds.Create != nil {
		checkCredentialsCreate(add, ctx.Env, section, creds.Create)
	}
	if stores != nil {
		if ctx.Defaults.TLS.Truststore == nil {
			add(fileEnv, "%s.stores requires tls.truststore", section)
		}
		// The in-container store path is fixed (application.yml always points at
		// it); stores only chooses the host source to bind-mount onto it. A custom
		// mount-path would silently break TLS, so reject it (S4a fail-loud).
		if stores.MountPath != spec.DefaultStoresMountPath {
			add(fileEnv, "%s.stores.mount-path %q is not supported; the in-container store path is fixed at %q", section, stores.MountPath, spec.DefaultStoresMountPath)
		}
	}
	if libs != nil && libs.Dir == "" {
		add(fileEnv, "%s.libs.dir is required when libs is set", section)
	}
}

// checkDocker validates the docker (compose) section.
func checkDocker(add func(string, string, ...any), ctx Context) {
	d := ctx.Docker
	checkContainerTarget(add, ctx, "docker", d.Name, d.Command, d.Image, d.Ports, d.Secrets.Credentials, d.Stores, d.Libs)
}

// checkPodman validates the podman section, including the generate mode and the
// quadlet scope (deploy/delete are always quadlet + systemctl).
func checkPodman(add func(string, string, ...any), ctx Context) {
	p := ctx.Podman
	checkContainerTarget(add, ctx, "podman", p.Name, p.Command, p.Image, p.Ports, p.Secrets.Credentials, p.Stores, p.Libs)
	switch p.Mode {
	case spec.PodmanModeRun, spec.PodmanModeQuadlet:
	default:
		add(fileEnv, "podman.mode must be %q or %q (got %q)", spec.PodmanModeRun, spec.PodmanModeQuadlet, p.Mode)
	}
	if q := p.Quadlet; q != nil {
		switch q.Scope {
		case spec.QuadletScopeAuto, spec.QuadletScopeUser, spec.QuadletScopeSystem:
		default:
			add(fileEnv, "podman.quadlet.scope must be auto, user, or system (got %q)", q.Scope)
		}
	}
}

// shellMeta lists the shell metacharacters SafeToken rejects on top of the
// safeShellChars set (which already covers spaces, quotes, backslash, backtick,
// and '$'). No shell is ever invoked -- the runner runs an argv slice via os/exec,
// never sh -c -- so this is defense-in-depth against a token being reinterpreted by
// a shell downstream. Legitimate CLI tokens (paths, flags, contexts) use only
// letters, digits, and - . / : _ = @ , +, none of which appear here.
const shellMeta = ";|&<>()*?[]{}~#!"

// SafeToken reports whether s is safe to pass as a single argv token: on top of
// safeShellChars (spaces, quotes, backslash, backtick, '$', control chars) it also
// rejects the shell metacharacters in shellMeta. The runner reuses this gate before
// executing any config-derived command token.
func SafeToken(s string) bool {
	return safeShellChars(s) && !strings.ContainsAny(s, shellMeta)
}
