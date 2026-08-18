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

// MaxWorkflows is the connector runtime's cap on workflow IDs (0..19) per
// application.yml, and therefore the most one connector instance can run. The
// Spring Boot app binds ids 0..19 only: a workflow numbered 20 or higher is
// silently ignored, not rejected.
//
// That silence is why exceeding this is a hard error here rather than a warning.
// A 21st workflow would otherwise generate cleanly, deploy cleanly, and simply
// never run -- with nothing in the config or the connector's own output saying
// so. Splitting the folder is the user's call, since only they know which flows
// belong on which connector.
const MaxWorkflows = 20

// Platform keys shared by validate, the runner, and the CLI: the section name
// in env.yaml, the argv[0] allowlist key, and the --allow-command threading
// all use these same three strings.
const (
	PlatformKubernetes = "kubernetes"
	PlatformDocker     = "docker"
	PlatformPodman     = "podman"
)

// AllowedCommands is the argv[0] allowlist for each deploy platform's
// `command:` field. CheckDeployCommand also honors any names passed as
// extraAllowed (deploy/delete's repeatable --allow-command flag), which is
// how an operator approves a chained binary (e.g. `sudo podman`) that the
// yaml author cannot grant on their own.
var AllowedCommands = map[string][]string{
	PlatformKubernetes: {"kubectl", "oc"},
	PlatformDocker:     {"docker"},
	PlatformPodman:     {"podman"},
}

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
	CheckKubernetes bool // kubernetes checks
	CheckDocker     bool // docker checks
	CheckPodman     bool // podman checks

	// Env, when non-nil, looks up a host environment variable so a `-env`
	// credential reference can be checked for presence at validate time (the CLI
	// supplies os.LookupEnv; nil skips the check).
	Env func(string) (string, bool)

	// AllowCommands extends the platform binary allowlist for CheckDeployCommand,
	// threaded from deploy/delete's repeatable --allow-command flag. Plain
	// `validate` leaves this nil: an exotic command (a chained binary like `sudo
	// podman`) validates clean only at deploy/delete time with the flag, and the
	// error text says so.
	AllowCommands []string
}

var (
	connNameRE = regexp.MustCompile(`^[^()\s,]+\(\d+\)(,[^()\s,]+\(\d+\))*$`)
	dns1123RE  = regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`)
	hostRE     = regexp.MustCompile(`^[A-Za-z0-9._:-]+$`)
	envNameRE  = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)
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

	checkWorkflowCount(add, len(ctx.Workflows))

	// Reusable connection definitions + per-workflow structural checks.
	checkConnections(add, warn, ctx.Env, d, haveKeystore)
	checkWorkflowSides(add, warn, ctx.Env, ctx.Workflows, haveKeystore, d.Connections)
	checkDefaultsCredentials(add, warn, ctx, d)

	// Cross-workflow (on resolved tuples): binder-level conflicts and duplicate sources.
	checkKeyAliasConflicts(add, resolved)
	checkPasswordConflicts(add, resolved)
	checkDuplicateSources(warn, resolved)

	// Leader-election (standalone | active_active | active_standby).
	checkLeaderElection(add, d, haveKeystore)

	// Reserved status-account name, and its password override charset.
	checkStatusUser(add, ctx.Env, d)

	// Per-target deploy-grade checks.
	checkTargets(add, warn, ctx, resolved)

	return errs, warns
}

// checkWorkflowCount rejects a folder holding more workflows than one connector
// instance can run. The tool does not split them: which flows belong on which
// connector is a deployment decision (they share a leader election, a set of
// credentials, and a resource budget), so it names the remedy instead of guessing.
func checkWorkflowCount(add func(string, string, ...any), n int) {
	if n <= MaxWorkflows {
		return
	}
	add(fileEnv, "%d workflows found, but one connector instance runs at most %d (workflow ids 0..%d). Split them across separate folders, each with its own env.yaml and its own deployment.name/docker.name/podman.name, and deploy each as its own connector",
		n, MaxWorkflows, MaxWorkflows-1)
}

// checkWorkflowSides runs the per-workflow structural + semantic checks on the
// raw sides (so conn-ref rules apply before resolution).
func checkWorkflowSides(add, warn func(string, string, ...any), env func(string) (string, bool), wfs []spec.Workflow, haveKeystore bool, conns map[string]spec.Side) {
	for _, wf := range wfs {
		if !wf.SourceSet {
			add(wf.File, "missing 'source'")
		}
		if !wf.TargetSet {
			add(wf.File, "missing 'target'")
		}
		if wf.SourceSet {
			checkSide(add, warn, env, wf.File, "source", wf.Source, true, haveKeystore, conns)
		}
		if wf.TargetSet {
			checkSide(add, warn, env, wf.File, "target", wf.Target, false, haveKeystore, conns)
		}
	}
}

// checkTargets runs the checks for whichever deploy targets are enabled. Each
// gate is independent so `config` skips them all, a per-target command lints
// only its section, and `validate` enables every section present.
func checkTargets(add, warn func(string, string, ...any), ctx Context, resolved []spec.Workflow) {
	if ctx.CheckKubernetes && ctx.Kube != nil {
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
func checkSide(add, warn func(string, string, ...any), env func(string) (string, bool), file, which string, s spec.Side, isSource, haveKeystore bool, conns map[string]spec.Side) {
	if !s.HasSystem() {
		add(file, "%s must specify exactly one of 'solace:' or 'mq:'", which)
		return
	}
	checkSideCredentials(add, warn, env, file, which, s)
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

// checkCred validates one credential position: the literal and `-env` forms are
// mutually exclusive, an `-env` value must be an environment-variable name (not
// a `${...}` reference or a value), and a `${` in the literal form is almost
// always someone reaching for the `-env` key.
//
// env, when non-nil, also confirms the named variable is actually set, so a
// missing credential is caught while linting rather than mid-deploy.
func checkCred(add, warn func(string, string, ...any), env func(string) (string, bool), file, label string, c spec.Cred) {
	if c.Both() {
		add(file, "%s sets both a literal value and %s-env; use one or the other", label, label)
		return
	}
	if c.EnvVar != "" {
		switch {
		case strings.Contains(c.EnvVar, "${"):
			add(file, "%s-env %q must be a bare variable name, not a ${...} reference", label, c.EnvVar)
		case !envNameRE.MatchString(c.EnvVar):
			add(file, "%s-env %q is not a valid environment variable name (letters, digits and underscore; not starting with a digit)", label, c.EnvVar)
		case env != nil:
			// Advisory only. Authoring a spec, generating a config, or linting on
			// a laptop must not require every production credential to be
			// exported; the deploy path resolves values and fails hard there.
			if _, ok := env(c.EnvVar); !ok {
				warn(file, "%s-env names %s, which is not set in this environment; deploying will fail until it is exported", label, c.EnvVar)
			}
		}
		return
	}
	if strings.Contains(c.Literal, "${") {
		warn(file, "%s looks like a variable reference; it is used as a literal value. Use %s-env: VAR to read it from the environment instead", label, label)
	}
}

// checkSideCredentials validates both credential positions of one side.
func checkSideCredentials(add, warn func(string, string, ...any), env func(string) (string, bool), file, label string, s spec.Side) {
	userKey, passKey := "client-username", "client-password"
	if s.System == spec.SystemMQ {
		userKey, passKey = "user", "password"
	}
	checkCred(add, warn, env, file, label+" "+userKey, s.Username())
	checkCred(add, warn, env, file, label+" "+passKey, s.Secret())
}

// checkDefaultsCredentials covers the credential positions that live in env.yaml
// itself rather than on a workflow side: the two store passwords, the management
// users, and the leader-election session.
func checkDefaultsCredentials(add, warn func(string, string, ...any), ctx Context, d *spec.Defaults) {
	checkCred(add, warn, ctx.Env, fileEnv, "tls.truststore.password", d.TLS.Truststore.Secret())
	checkCred(add, warn, ctx.Env, fileEnv, "tls.keystore.password", d.TLS.Keystore.Secret())
	for i, u := range d.Security.Users {
		label := fmt.Sprintf("security.users[%d].password", i)
		checkCred(add, warn, ctx.Env, fileEnv, label, u.Secret())
	}
	if s := d.LeaderElection.Session; s != nil {
		checkSideCredentials(add, warn, ctx.Env, fileEnv, "leader-election session", *s)
	}
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
func checkConnections(add, warn func(string, string, ...any), env func(string) (string, bool), d *spec.Defaults, haveKeystore bool) {
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
		checkSideCredentials(add, warn, env, fileEnv, label, c)
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

// checkStatusUser guards the reserved read-only actuator account the tool
// injects into every rendered application.yml (named spec.StatusUserName)
// once management security is effectively enabled. No author-configured user
// may claim that name -- it would collide with the account the tool adds
// itself. If the account's generated password is overridden via
// spec.StatusUserPasswordEnvVar, the override must survive being rendered as
// a literal into application.yml and parsed back out of that file by the
// generated status script.
func checkStatusUser(add func(string, string, ...any), env func(string) (string, bool), d *spec.Defaults) {
	if !d.Security.EffectivelyEnabled() {
		return
	}
	for i, u := range d.Security.Users {
		if u.Name == spec.StatusUserName {
			add(fileEnv, "security.users[%d].name %q is reserved for the tool's own read-only status account used by the generated status script; choose a different name", i, u.Name)
		}
	}
	if env == nil {
		return
	}
	v, ok := env(spec.StatusUserPasswordEnvVar)
	if !ok || v == "" {
		return
	}
	if !safeStatusPassword(v) {
		add(fileEnv, "%s is set but its value contains a character the generated status script cannot round-trip: the password is rendered as a literal into application.yml and parsed back out by a shell script, so whitespace, quotes, backslash, and ${...} would corrupt the YAML or break the parse. Use only printable ASCII, excluding those characters", spec.StatusUserPasswordEnvVar)
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
		key := s.DedupKey()
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

// checkPasswordConflicts flags binders (same dedup tuple) whose sides declare
// different passwords. The dedup tuple deliberately omits the password, so two
// sides that disagree still collapse into one binder and consolidate keeps only
// the first (by filename) -- silently authenticating every workflow on that
// binder with a credential the later file never asked for. An auth mismatch on
// one connection is a real misconfiguration, so unlike the cipher conflict
// (last-wins warning) this is fatal.
func checkPasswordConflicts(add func(string, string, ...any), wfs []spec.Workflow) {
	type seen struct{ pass, file string }
	byKey := map[string]seen{}
	visit := func(file string, s spec.Side) {
		if !s.HasSystem() {
			return
		}
		c := s.Secret()
		if c.Empty() {
			return
		}
		key := s.DedupKey()
		if prev, ok := byKey[key]; ok {
			if prev.pass != c.Key() {
				add(file, "conflicting password for the same binder: this side uses %s, %s uses a different one — one connection cannot authenticate two ways, so give it a single password or make the tuples distinct", c.Describe(), prev.file)
			}
			return
		}
		byKey[key] = seen{c.Key(), file}
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
		key := s.DedupKey() + "|" + s.Dest
		if prev, ok := byKey[key]; ok {
			warn(wf.File, "source %q is also consumed in %s (duplicate source on the same binder)", s.Dest, prev.file)
			continue
		}
		byKey[key] = seen{wf.File}
	}
}

func checkKube(add, warn func(string, string, ...any), ctx Context) {
	if err := CheckDeployCommand(PlatformKubernetes, ctx.Kube.Command, ctx.AllowCommands); err != nil {
		add(fileEnv, "kubernetes.command %s", err)
	}

	dep := ctx.Kube.Deployment
	if dep.Name == "" {
		add(fileEnv, "deployment.name is required")
	} else if !isDNS1123(dep.Name) {
		add(fileEnv, "deployment.name %q is not a valid DNS-1123 label", dep.Name)
	} else if longest := dep.Name + "-config"; len(longest) > 63 {
		// The ConfigMap name is the longest object derived from deployment.name
		// and must still be a valid DNS-1123 label.
		add(fileEnv, "deployment.name %q is too long: it derives the ConfigMap name %q, which exceeds the 63-char DNS-1123 limit", dep.Name, longest)
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

	// Secret wiring checks. Every name here is emitted verbatim as a manifest
	// metadata.name / secretRef, so it is held to the same DNS-1123 rule the
	// cluster would apply anyway -- caught at the gate with a readable message
	// instead of by kubectl mid-apply (libs.pvc.create.name is checked the same
	// way in checkLibs).
	if c := ctx.Kube.Secrets.Credentials; c != nil {
		if c.Create != nil {
			checkSecretName(add, "kubernetes.secrets.credentials.create.name", c.Create.Name)
			if removed := c.Create.RemovedKeys(); len(removed) > 0 {
				add(fileEnv, "kubernetes.secrets.credentials.create no longer takes %s: the Secret's keys are every credential the config references, and their values come from the literals and -env variables those fields name. Remove %s", strings.Join(removed, "/"), strings.Join(removed, ", "))
			}
		}
		if c.Existing != "" {
			checkSecretName(add, "kubernetes.secrets.credentials.existing", c.Existing)
		}
	}
	if s := ctx.Kube.Secrets.Stores; s != nil {
		if s.Create != nil {
			if ctx.Defaults.TLS.Truststore == nil {
				add(fileEnv, "kubernetes.secrets.stores.create requires tls.truststore")
			}
			checkSecretName(add, "kubernetes.secrets.stores.create.name", s.Create.Name)
		}
		if s.Existing != "" {
			checkSecretName(add, "kubernetes.secrets.stores.existing", s.Existing)
		}
	}

	checkSyslog(add, warn, ctx.Kube)
	checkLibs(add, ctx.Kube)
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
			} else {
				// Both land unquoted in the PersistentVolume manifest piped to
				// kubectl, so a newline would inject a sibling key.
				if !hostRE.MatchString(c.NFS.Server) {
					add(fileEnv, "libs.pvc.create nfs.server %q may only contain letters, digits, '.', ':', '_' and '-'", c.NFS.Server)
				}
				if !safeHostPath(c.NFS.Path) {
					add(fileEnv, "libs.pvc.create nfs.path %q contains an unsafe character (no whitespace, quotes, control chars, or shell metacharacters)", c.NFS.Path)
				}
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

// safeHostPath gates a config-declared host path that a docker/podman renderer
// concatenates unquoted into a mount argument (`-v src:dst:ro` in the run script,
// `Volume=` in a quadlet unit, a compose `volumes:` entry). It rejects whitespace,
// control characters, quotes, backtick and '$' plus the shell metacharacters: a
// newline would open a new script line, unit directive, or YAML key, and a space
// would split the mount argument in two.
//
// Unlike safeShellChars it permits '\' and ':' so a Windows-authored path
// (C:\certs\truststore.jks) still validates -- neither can escape any of those
// three sinks. Only the value as written in env.yaml is checked, never the
// absolute path it resolves to, which is the developer's own working directory.
func safeHostPath(s string) bool {
	for _, r := range s {
		if r <= 0x20 || r == 0x7f || r == '\'' || r == '"' || r == '`' || r == '$' {
			return false
		}
	}
	return !strings.ContainsAny(s, shellMeta)
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

// checkSecretName gates a Secret name that is emitted verbatim into the manifest
// (metadata.name, secretRef.name, volumes[].secret.secretName). An empty name is
// its own message: a create block without one renders a nameless object.
func checkSecretName(add func(string, string, ...any), field, name string) {
	if name == "" {
		add(fileEnv, "%s is required", field)
		return
	}
	if !isDNS1123(name) {
		add(fileEnv, "%s %q is not a valid DNS-1123 label", field, name)
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

// CheckDeployCommand is the single shared gate for a deploy/delete `command:`
// field, used by both the validator (nil extraAllowed) and the runner (the
// operator's --allow-command values). It layers three checks on top of
// Tokenize's charset gate, in order, and returns nil or one error naming the
// first offending token:
//
//  1. Empty command -> "must not be empty".
//  2. argv[0] (a ".exe"/".EXE" suffix is stripped before comparing, so a
//     Windows-authored config stays portable): a path separator is rejected
//     outright (a bare PATH-resolved name only); otherwise the token must be
//     in AllowedCommands[platform] or extraAllowed.
//  3. Every later token must be flag-shaped: it starts with "-"; or it is
//     itself allowlisted/extraAllowed after the same .exe-strip (this is what
//     lets a chained binary like `sudo podman` validate once "sudo" is
//     extraAllowed); or the previous token starts with "-" and contains no
//     "=" (so this token is that flag's value). A literal "--" is rejected
//     outright: as an end-of-flags marker it would let a positional argument
//     smuggle past this rule right after it.
//
// Honest limitation: a flag's value is not further inspectable (arity is
// unknowable from the token stream alone), so the hard guarantee here covers
// argv[0] and every bare token -- not what a flag's value itself contains.
func CheckDeployCommand(platform, cmd string, extraAllowed []string) error {
	tokens, bad := Tokenize(cmd)
	if bad != nil {
		if len(bad) == 0 {
			return fmt.Errorf("must not be empty")
		}
		return fmt.Errorf("token %q contains an unsafe character (%s)", bad[0], UnsafeTokenReason)
	}

	stripExe := func(tok string) string {
		if s, ok := strings.CutSuffix(tok, ".exe"); ok {
			return s
		}
		if s, ok := strings.CutSuffix(tok, ".EXE"); ok {
			return s
		}
		return tok
	}
	allowed := AllowedCommands[platform]
	isAllowed := func(name string) bool {
		for _, a := range allowed {
			if a == name {
				return true
			}
		}
		for _, a := range extraAllowed {
			if a == name {
				return true
			}
		}
		return false
	}

	bin := tokens[0]
	if binName := stripExe(bin); strings.Contains(binName, "/") {
		return fmt.Errorf("%q: a path is not accepted here; use a bare binary name, resolved from PATH", bin)
	} else if !isAllowed(binName) {
		return fmt.Errorf("%q: binary must be one of %s; deploy/delete can approve another with --allow-command <name>", bin, strings.Join(allowed, ", "))
	}

	for i := 1; i < len(tokens); i++ {
		tok := tokens[i]
		switch {
		case tok == "--":
			return fmt.Errorf(`token "--": end-of-flags marker is not accepted`)
		case strings.HasPrefix(tok, "-"):
		case isAllowed(stripExe(tok)):
		case strings.HasPrefix(tokens[i-1], "-") && !strings.Contains(tokens[i-1], "="):
			// This token is the previous flag's value.
		default:
			return fmt.Errorf("token %q: arguments must be flag-shaped (-x, --flag, --flag=value, or a flag's value); solmq-conn-util appends its own subcommand", tok)
		}
	}
	return nil
}

// UnsafeTokenReason describes the safe-token charset, shared so the validator and
// the runner reject a token with the same words.
const UnsafeTokenReason = "no spaces, quotes, backslash, control chars, or shell metacharacters"

// Tokenize splits a command string into argv on whitespace and reports which
// tokens fail the safe charset. It is the single definition of "how a command:
// becomes argv", used by the validator (which reports every bad token) and by
// the runner (which refuses to start a process on the first one) -- so what
// validation accepts and what actually executes can never drift apart.
//
// bad is nil when every token is safe; a non-nil, empty bad means the command
// was empty.
func Tokenize(cmd string) (tokens, bad []string) {
	tokens = strings.Fields(cmd)
	if len(tokens) == 0 {
		return nil, []string{}
	}
	for _, t := range tokens {
		if !SafeToken(t) {
			bad = append(bad, t)
		}
	}
	return tokens, bad
}

// containerTarget is one docker/podman section flattened to the fields the shared
// checks need. Grouping them keeps checkContainerTarget's call sites named rather
// than positional, so a new shared attribute is one struct field instead of a
// wider signature at three places.
type containerTarget struct {
	Section  string
	Platform string
	Name     string
	Command  string
	Image    string
	Restart  string
	Timezone string
	Ports    []spec.Port
	Secrets  *spec.Secrets // removed from the schema; non-nil is rejected
	Stores   *spec.StoresMount
	Libs     *spec.LibsMount
}

// secretsMountPath is where credentials are mounted on every platform. It is
// named here for error messages; internal/deploy owns the manifest-side constant.
const secretsMountPath = "/run/secrets"

// checkContainerTarget runs the checks common to the docker and podman sections:
// a DNS-1123 name (it flows into filesystem paths and a systemctl unit token), a
// deploy command gated by CheckDeployCommand (bare, allowlisted binary plus
// flag-shaped arguments), a required image, valid ports, credentials wiring, and
// host-provided stores/libs paths.
//
// image/restart/timezone and the host paths are gated here because both renderers
// concatenate them unquoted into artifacts that are later executed or parsed --
// a `podman run` script line, a quadlet directive, a compose YAML line -- so an
// embedded newline or metacharacter would add content the spec never declared.
func checkContainerTarget(add func(string, string, ...any), ctx Context, t containerTarget) {
	section := t.Section
	// The section is gone: credentials are derived from the config's own
	// credential fields and delivered as platform secrets. yaml ignores unknown
	// keys, so without this an old env.yaml would parse and silently drop them.
	if t.Secrets != nil {
		add(fileEnv, "%s.secrets is no longer configured: credentials come from the connection fields themselves (client-password / client-password-env, password / password-env, ...) and are mounted at %s. Remove the %s.secrets section", section, secretsMountPath, section)
	}
	if !isDNS1123(t.Name) {
		add(fileEnv, "%s.name %q is not a valid DNS-1123 label", section, t.Name)
	}
	if err := CheckDeployCommand(t.Platform, t.Command, ctx.AllowCommands); err != nil {
		add(fileEnv, "%s.command %s", section, err)
	}
	if t.Image == "" {
		add(fileEnv, "%s.image is required", section)
	} else if !SafeToken(t.Image) {
		add(fileEnv, "%s.image %q contains an unsafe character (no spaces, quotes, backslash, control chars, or shell metacharacters)", section, t.Image)
	}
	if t.Restart != "" && !SafeToken(t.Restart) {
		add(fileEnv, "%s.restart %q contains an unsafe character (no spaces, quotes, backslash, control chars, or shell metacharacters)", section, t.Restart)
	}
	if t.Timezone != "" && !SafeToken(t.Timezone) {
		add(fileEnv, "%s.timezone %q contains an unsafe character (no spaces, quotes, backslash, control chars, or shell metacharacters)", section, t.Timezone)
	}
	checkPorts(add, section, t.Ports)
	if t.Stores != nil {
		if ctx.Defaults.TLS.Truststore == nil {
			add(fileEnv, "%s.stores requires tls.truststore", section)
		}
		// The in-container store path is fixed (application.yml always points at
		// it); stores only chooses the host source to bind-mount onto it. A custom
		// mount-path would silently break TLS, so reject it (S4a fail-loud).
		if t.Stores.MountPath != spec.DefaultStoresMountPath {
			add(fileEnv, "%s.stores.mount-path %q is not supported; the in-container store path is fixed at %q", section, t.Stores.MountPath, spec.DefaultStoresMountPath)
		}
		// Only when stores opts in do the tls.*.file paths become bind-mount
		// sources in a generated artifact; the kubernetes path embeds their
		// content instead, so it is not gated here.
		for _, st := range []struct {
			field string
			store *spec.Store
		}{{"tls.truststore.file", ctx.Defaults.TLS.Truststore}, {"tls.keystore.file", ctx.Defaults.TLS.Keystore}} {
			if st.store == nil || st.store.File == "" {
				continue
			}
			if !safeHostPath(st.store.File) {
				add(fileEnv, "%s.stores bind-mounts %s %q, which contains an unsafe character (no whitespace, quotes, control chars, or shell metacharacters)", section, st.field, st.store.File)
			}
		}
	}
	if t.Libs != nil {
		if t.Libs.Dir == "" {
			add(fileEnv, "%s.libs.dir is required when libs is set", section)
		} else if !safeHostPath(t.Libs.Dir) {
			add(fileEnv, "%s.libs.dir %q contains an unsafe character (no whitespace, quotes, control chars, or shell metacharacters)", section, t.Libs.Dir)
		}
	}
}

// checkDocker validates the docker (compose) section.
func checkDocker(add func(string, string, ...any), ctx Context) {
	d := ctx.Docker
	checkContainerTarget(add, ctx, containerTarget{
		Section:  "docker",
		Platform: PlatformDocker,
		Name:     d.Name,
		Command:  d.Command,
		Image:    d.Image,
		Restart:  d.Restart,
		Timezone: d.Timezone,
		Ports:    d.Ports,
		Secrets:  d.Secrets,
		Stores:   d.Stores,
		Libs:     d.Libs,
	})
}

// checkPodman validates the podman section, including the generate mode and the
// quadlet scope (deploy/delete are always quadlet + systemctl).
func checkPodman(add func(string, string, ...any), ctx Context) {
	p := ctx.Podman
	checkContainerTarget(add, ctx, containerTarget{
		Section:  "podman",
		Platform: PlatformPodman,
		Name:     p.Name,
		Command:  p.Command,
		Image:    p.Image,
		Restart:  p.Restart,
		Timezone: p.Timezone,
		Ports:    p.Ports,
		Secrets:  p.Secrets,
		Stores:   p.Stores,
		Libs:     p.Libs,
	})
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

// SafeActuatorUser reports whether s is safe to use as the management account
// name the generated status script authenticates as. On top of SafeToken's
// argv rules the charset is an allowlist of letters, digits, '.', '-' and '_',
// because the name is also spliced into a sed address inside the script: '/'
// would end the address early and leave the script silently unauthenticated,
// and a regex metacharacter could match an unintended account. The script
// escapes the name as well (statusscript.breEscape) -- this is the boundary
// half of that pair.
func SafeActuatorUser(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9':
		case r == '.' || r == '-' || r == '_':
		default:
			return false
		}
	}
	return true
}

// SafeActuatorUserReason explains the SafeActuatorUser charset in errors.
const SafeActuatorUserReason = "only letters, digits, '.', '-' and '_' are allowed"

// safeStatusPassword gates a status-account password override
// (spec.StatusUserPasswordEnvVar): the value is rendered as a literal into
// application.yml and later parsed back out of that file by the generated
// status script's sed invocation, so it must be non-empty and consist only
// of printable ASCII, excluding whitespace, single/double quotes, backslash,
// and the '$', '{', '}' characters that make up a ${...} placeholder -- any
// of those would corrupt the YAML literal or break the sed match. Unlike
// SafeToken this value is never passed to a shell as a token, so the shell
// metacharacters SafeToken also rejects (!, #, ;, ...) are fine here.
func safeStatusPassword(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < 0x21 || r > 0x7e || r == '\'' || r == '"' || r == '\\' || r == '$' || r == '{' || r == '}' {
			return false
		}
	}
	return true
}
