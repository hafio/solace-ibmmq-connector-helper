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
// extraAllowed (deploy/remove's repeatable --allow-command flag), which is
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
	Image     *spec.Image      // the one image every platform deploys; nil when the block is absent
	Timezone  string           // the container TZ every platform sets; empty when unset
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
	// threaded from deploy/remove's repeatable --allow-command flag. Plain
	// `validate` leaves this nil: an exotic command (a chained binary like `sudo
	// podman`) validates clean only at deploy/remove time with the flag, and the
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

	// Retired keys: still parsed (spec.Security.Enabled, spec.Management.Exposure,
	// spec.LeaderElection.SolaceKey), always rejected, so an old env.yaml fails
	// loudly instead of silently keeping a toggle that no longer has any effect, or
	// a block under a name nothing reads (mirrors docker/podman .secrets).
	checkRemovedDefaultsKeys(add, d)

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
	checkLeaderSessionPassword(add, d, resolved)

	// Reserved status-account name, and its password override charset.
	checkStatusUser(add, ctx.Env, d)
	checkSecurityUserRoles(add, d)

	// The image every platform deploys, and then the per-target deploy-grade checks.
	checkImage(add, warn, ctx)
	// Unconditional: syslog is a top-level key now, so a docker-only run or a
	// bare `generate config` has to validate it too. Under kubernetes only, it
	// would go unchecked everywhere else it now applies.
	checkSyslog(add, warn, d.Syslog)
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

// checkRemovedDefaultsKeys rejects the connector-defaults keys the tool retired:
// security.enabled (management security can no longer be turned off) and
// management.exposure (the actuator exposure list is no longer configurable),
// now that management is unconditionally locked down, plus the old
// leader-election.solace spelling of the management session. Every one of them
// still parses a value through from env.yaml purely so a stale key is caught
// here instead of silently doing nothing. This runs whatever the leader-election
// mode is, so the rename is reported even under standalone.
func checkRemovedDefaultsKeys(add func(string, string, ...any), d *spec.Defaults) {
	if d.Security.Enabled != nil {
		add(fileEnv, "security.enabled is no longer configurable: the tool always injects a read-only actuator account (%s) so a running instance can always be queried for its leader-election state, and disabling auth would also leave the write-capable /actuator/workflows endpoint open. Remove the key", spec.StatusUserName)
	}
	if d.Management.Exposure != nil {
		add(fileEnv, "management.exposure is no longer configurable: the tool always exposes exactly health,info,metrics,leaderelection,workflows. Remove the key")
	}
	if d.LeaderElection.SolaceKey {
		add(fileEnv, "leader-election.solace has been renamed to leader-election.session, which is what it renders to (solace.connector.management.session). Rename the key")
	}
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
		// Warn: credentials are referenced but the credentials Secret is not
		// wired. Without the block there is no volume and no mount, so every
		// ${...} the config carries stays unresolved -- and because the
		// configtree import is optional, the connector starts anyway and fails
		// later against whatever it could not authenticate to. Nothing else
		// says so, which is the whole reason this warning exists.
		if usesCredentials(resolved, ctx.Defaults) && !credentialsWired(ctx.Kube) {
			warn(fileEnv, "credentials are referenced but kubernetes.secrets.credentials is omitted; they will not be mounted at "+secretsMountPath+" and every ${...} in the generated config will stay unresolved at runtime")
		}
		// Warn: TLS/mTLS in use but the stores Secret is not wired.
		if usesTLS(resolved) && !storesWired(ctx.Kube) {
			warn(fileEnv, "a TLS/mTLS connection exists but secrets.stores is omitted; the store files will be missing at runtime")
		}
	}
	// docker and podman carry no stores warning: the tls.*.file paths are always
	// bind-mounted now, so "TLS configured but nothing mounted" cannot arise. The
	// kubernetes one above stays -- it embeds store content in a Secret, which is
	// still something the operator has to wire up.
	if ctx.CheckDocker && ctx.Docker != nil {
		checkDocker(add, ctx)
	}
	if ctx.CheckPodman && ctx.Podman != nil {
		checkPodman(add, ctx)
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
		case strings.HasPrefix(c.EnvVar, spec.GeneratedNamePrefix):
			// An -env credential is mounted under the variable name it gives, and
			// so shares one namespace with the names the tool derives for literals.
			// Reserving this prefix is what keeps the two apart: without it a
			// variable spelled like a derived name would put two credentials on one
			// mounted file. envNameRE would accept it (a leading underscore is a
			// legal identifier), so it needs its own case.
			add(file, "%s-env %q starts with %s, which is reserved for the mount names this tool derives for itself; rename the variable so it falls outside that prefix", label, c.EnvVar, spec.GeneratedNamePrefix)
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
// a Solace management session (conn-ref to a solace connection, or inline session).
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
	// conn-ref wins in consolidate, so an inline block alongside one is dead
	// config -- and dead config that is still credential-checked, which produces
	// errors about a session nothing renders.
	if le.ConnRef != "" && le.Session != nil {
		add(fileEnv, "leader-election sets both conn-ref %q and an inline session: block; keep one", le.ConnRef)
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
		add(fileEnv, "leader-election mode %q requires a solace session (conn-ref or inline session:)", le.Mode)
	}
	// A management session coordinates leadership; it has no binding, so a
	// destination or consumer/producer tuning written here renders nothing. The
	// management queue is leader-election.queue, one level up -- which is the
	// mistake this catches.
	if le.Session != nil {
		if got := le.Session.BindingFields(); len(got) > 0 {
			add(fileEnv, "leader-election.session may not set queue/topic/consumer/producer: a management session is a connection, not a binding (it sets %s here), and the management queue is leader-election.queue, one level up", strings.Join(got, ", "))
		}
	}
}

// checkLeaderSessionPassword flags a management session whose broker tuple is
// one the workflows already bind but whose password disagrees with it. The dedup
// tuple omits the password, so the two are one binder and one mounted credential
// -- consolidate shares the stable secret name on purpose, so the credential is
// mounted once -- and the binder's password, recorded first, is the one the
// session would silently authenticate with. Fatal for the same reason
// checkPasswordConflicts is.
func checkLeaderSessionPassword(add func(string, string, ...any), d *spec.Defaults, wfs []spec.Workflow) {
	le := d.LeaderElection
	if !le.Present || le.Mode == "" || le.Mode == spec.LeaderStandalone {
		return
	}
	sess := le.Session
	if le.ConnRef != "" {
		c, ok := d.Connections[le.ConnRef]
		if !ok {
			return // the dangling ref is already reported by checkLeaderElection
		}
		sess = &c
	}
	if sess == nil || sess.System != spec.SystemSolace {
		return
	}
	cred := sess.Secret()
	if cred.Empty() {
		return
	}
	key := sess.DedupKey()
	for _, wf := range wfs {
		for _, s := range []spec.Side{wf.Source, wf.Target} {
			if s.System != spec.SystemSolace || s.DedupKey() != key {
				continue
			}
			if b := s.Secret(); !b.Empty() && b.Key() != cred.Key() {
				add(fileEnv, "the leader-election session uses %s, but %s reaches the same broker tuple with a different password: they collapse onto one binder and one mounted credential, so give them a single password or make the tuples distinct", cred.Describe(), wf.File)
				return
			}
		}
	}
}

// checkStatusUser guards the reserved read-only actuator account the tool
// injects into every rendered application.yml (named spec.StatusUserName).
// Management security is always on, so this always runs. No author-configured
// user may claim that name -- it would collide with the account the tool adds
// itself. If the account's generated password is overridden via
// spec.StatusUserPasswordEnvVar, the override must survive being rendered as
// a literal into application.yml and parsed back out of that file by the
// generated status script.
func checkStatusUser(add func(string, string, ...any), env func(string) (string, bool), d *spec.Defaults) {
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

// checkSecurityUserRoles gates the roles an operator grants their own actuator
// accounts. Roles are passed through to the connector verbatim, so this is not
// an allowlist of known names -- the connector owns that vocabulary and a
// hardcoded list here would reject a role it adds later. What is checked is that
// each entry is usable at all: an empty role would render as a bare "- " item
// the connector cannot map to an authority, and an unsafe-charset role is far
// more likely a typo than a real authority name.
//
// The reserved status account is not covered here: the tool appends it with no
// roles at all (consolidate.applyStatusAccess), which is what keeps it
// read-only, and checkStatusUser already rejects an operator claiming its name.
func checkSecurityUserRoles(add func(string, string, ...any), d *spec.Defaults) {
	for i, u := range d.Security.Users {
		for j, r := range u.Roles {
			if strings.TrimSpace(r) == "" {
				add(fileEnv, "security.users[%d].roles[%d] is empty: give the role a name (for example %q, which grants the read/write access needed to POST to /actuator/workflows) or drop the entry", i, j, "admin")
				continue
			}
			if !SafeToken(r) {
				add(fileEnv, "security.users[%d].roles[%d] %q contains an unsafe character (%s); a role is an authority name passed to the connector verbatim", i, j, r, UnsafeTokenReason)
			}
		}
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
	if dep.Image != "" {
		add(fileEnv, "kubernetes.deployment.image is no longer configured here: the image moved to the top-level image: block (repo/name/tag) so one declaration serves every platform. Remove kubernetes.deployment.image")
	}
	if dep.Timezone != "" {
		add(fileEnv, "kubernetes.deployment.timezone is no longer configured here: the container timezone moved to the top-level timezone: key so one declaration serves every platform. Remove kubernetes.deployment.timezone")
	}

	// service.port is the same scalar / "host:container" shape as docker/podman
	// ports (Port.UnmarshalYAML enforces the shape itself at parse); range-check
	// both sides regardless of service.enabled, since an unset port is already
	// defaulted to management.port by the time validate sees it.
	checkPortRange(add, "kubernetes.service.port", ctx.Kube.Service.Port)

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
		// create and existing are mutually exclusive: Render takes the create
		// branch when both are set, which would emit a Secret doc over the very
		// object 'existing' names. Omitting the whole credentials block stays
		// legal -- it is a present-but-undecided block that is rejected here.
		if (c.Create != nil) == (c.Existing != "") {
			add(fileEnv, "kubernetes.secrets.credentials must set exactly one of 'create' or 'existing'")
		}
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
		if (s.Create != nil) == (s.Existing != "") {
			add(fileEnv, "kubernetes.secrets.stores must set exactly one of 'create' or 'existing'")
		}
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
	checkImagePull(add, warn, ctx)

	if ctx.Kube.Logging != nil {
		add(fileEnv, "kubernetes.logging is no longer configured here: syslog moved to the top-level logging: block (beside logging.level) so one declaration serves every platform. Remove kubernetes.logging")
	}
	checkLibs(add, ctx.Kube)
}

// checkSyslog validates the optional top-level logging.syslog block.
//
// It takes the block rather than a section because it is no longer a kubernetes
// key: the same settings drive the docker and podman renderers, so the checks
// have to run whichever platform is being generated.
func checkSyslog(add, warn func(string, string, ...any), s *spec.Syslog) {
	if s == nil {
		return
	}
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
			// The PersistentVolume is cluster-scoped, so its name carries the
			// namespace to stop two releases fighting over one object -- and the
			// result is still a single DNS-1123 label. Caught here rather than by
			// the API server mid-apply, the same way the ConfigMap name is
			// checked against deployment.name above.
			if pv := spec.LibsPVName(k.Deployment.Namespace, c.Name); len(pv) > 63 {
				add(fileEnv, "libs.pvc.create.name %q derives the PersistentVolume name %q, which exceeds the 63-char DNS-1123 limit: shorten it or deployment.namespace", c.Name, pv)
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

// credentialsWired reports whether the kubernetes section actually asks for a
// credentials Secret. An omitted block is legal -- a config with no credentials
// at all needs none -- so this is only half the question; usesCredentials
// answers the other half.
func credentialsWired(k *spec.Kubernetes) bool {
	if k == nil || k.Secrets.Credentials == nil {
		return false
	}
	return k.Secrets.Credentials.Create != nil || k.Secrets.Credentials.Existing != ""
}

// usesCredentials reports whether anything in the config resolves to a mounted
// credential, which is what makes an omitted credentials block a problem rather
// than a choice.
//
// It walks the same positions consolidate collects from -- both sides of every
// workflow, the management session, the management accounts, and the two store
// passwords -- because a name that reaches the config as ${...} has to be
// readable from the mount, wherever in the spec it was written.
func usesCredentials(wfs []spec.Workflow, d *spec.Defaults) bool {
	for _, wf := range wfs {
		for _, s := range []spec.Side{wf.Source, wf.Target} {
			if !s.Username().Empty() || !s.Secret().Empty() {
				return true
			}
		}
	}
	if d == nil {
		return false
	}
	if s := d.LeaderElection.Session; s != nil {
		if !s.Username().Empty() || !s.Secret().Empty() {
			return true
		}
	}
	for _, u := range d.Security.Users {
		if !u.Secret().Empty() {
			return true
		}
	}
	return !d.TLS.Truststore.Secret().Empty() || !d.TLS.Keystore.Secret().Empty()
}

func isDNS1123(s string) bool { return len(s) <= 63 && dns1123RE.MatchString(s) }

// checkSecretName gates a Secret name that is emitted verbatim into the manifest
// (metadata.name, secretRef.name, volumes[].secret.secretName). An empty name is
// its own message: a create block without one renders a nameless object.
// checkImage validates the top-level image: block -- the one declaration every
// platform deploys from. It runs only when a platform is in play: `generate
// config` renders application.yml alone and never needs an image.
func checkImage(add, warn func(string, string, ...any), ctx Context) {
	if !ctx.CheckKubernetes && !ctx.CheckDocker && !ctx.CheckPodman {
		return
	}
	// The timezone rides along here: it is the other value every platform now
	// shares, it lands in a manifest, a compose file and a podman argv exactly as
	// the image fields do, and it is optional -- so only the charset is checked.
	if ctx.Timezone != "" && !SafeToken(ctx.Timezone) {
		add(fileEnv, "timezone %q contains an unsafe character (%s)", ctx.Timezone, UnsafeTokenReason)
	}
	i := ctx.Image
	if i == nil {
		add(fileEnv, "image: is required: set image.name and image.tag, plus image.repo for a registry other than Docker Hub")
		return
	}
	if i.Name == "" {
		add(fileEnv, "image.name is required")
	}
	// An untagged reference resolves to :latest, which pins nothing -- the same
	// reason every other dependency here carries an explicit version.
	if i.Tag == "" {
		add(fileEnv, "image.tag is required: an untagged image resolves to :latest, which pins nothing. Use a version tag, or a sha256: digest to pin exactly")
	}
	// Every field lands in a manifest, a compose file or a podman argv, so all of
	// them are held to the safe charset rather than only the ones that reach a
	// command line.
	for _, f := range []struct{ key, val string }{
		{"repo", i.Repo},
		{"name", i.Name},
		{"tag", i.Tag},
	} {
		if f.val != "" && !SafeToken(f.val) {
			add(fileEnv, "image.%s %q contains an unsafe character (%s)", f.key, f.val, UnsafeTokenReason)
		}
	}
	// The registry account is a credential pair like every other one, so it gets
	// the shared check: literal xor -env, a valid variable name, and a warning
	// when the variable is not exported.
	checkCred(add, warn, ctx.Env, fileEnv, "image.user", i.UserCred())
	checkCred(add, warn, ctx.Env, fileEnv, "image.pass", i.PassCred())
}

// checkImagePull validates kubernetes.secrets.image-pull. The name alone
// references a Secret the operator manages; create additionally builds one from
// the registry account in the image: block, which is the only case that needs
// those credentials -- so they are required there and optional everywhere else.
func checkImagePull(add, warn func(string, string, ...any), ctx Context) {
	ip := ctx.Kube.Secrets.ImagePull
	if ip == nil {
		return
	}
	checkSecretName(add, "kubernetes.secrets.image-pull.name", ip.Name)
	if !ip.Create {
		return
	}
	if ctx.Image.UserCred().Empty() || ctx.Image.PassCred().Empty() {
		add(fileEnv, "kubernetes.secrets.image-pull.create requires image.user and image.pass (or their -env forms): the Secret is built from them. Omit create to reference a Secret you manage yourself instead")
	}
}

func checkSecretName(add func(string, string, ...any), field, name string) {
	if name == "" {
		add(fileEnv, "%s is required", field)
		return
	}
	if !isDNS1123(name) {
		add(fileEnv, "%s %q is not a valid DNS-1123 label", field, name)
	}
}

// checkPortRange flags a spec.Port whose host or container side falls outside
// the valid 1-65535 range (the "host:container" shape itself is enforced at
// parse in spec.Port). label is the field path prefixed onto "host port"/
// "container port" in the message, e.g. "docker.ports" or
// "kubernetes.service.port".
func checkPortRange(add func(string, string, ...any), label string, p spec.Port) {
	if p.Host < 1 || p.Host > 65535 {
		add(fileEnv, "%s host port %d must be 1-65535", label, p.Host)
	}
	if p.Container < 1 || p.Container > 65535 {
		add(fileEnv, "%s container port %d must be 1-65535", label, p.Container)
	}
}

// checkPorts runs checkPortRange over a docker/podman ports list.
func checkPorts(add func(string, string, ...any), section string, ports []spec.Port) {
	for _, p := range ports {
		checkPortRange(add, section+".ports", p)
	}
}

// CheckDeployCommand is the single shared gate for a deploy/remove `command:`
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
		return fmt.Errorf("%q: binary must be one of %s; deploy/remove can approve another with --allow-command <name>", bin, strings.Join(allowed, ", "))
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
	Secrets  *spec.Secrets     // removed from the schema; non-nil is rejected
	Stores   *spec.StoresMount // removed from the schema; non-nil is rejected
	Libs     *spec.LibsMount
}

// secretsMountPath is where credentials are mounted on every platform, named
// here for error messages; spec.SecretsMountPath is the definition.
const secretsMountPath = spec.SecretsMountPath

// checkContainerTarget runs the checks common to the docker and podman sections:
// a DNS-1123 name (it flows into filesystem paths and a systemctl unit token), a
// deploy command gated by CheckDeployCommand (bare, allowlisted binary plus
// flag-shaped arguments), a required image, valid ports, credentials wiring, and
// host-provided stores/libs paths.
//
// image/restart/timezone and the host paths are gated here because both renderers
// concatenate them unquoted into artifacts that are later executed or parsed --
// a quadlet directive, a compose YAML line -- so an embedded newline or
// metacharacter would add content the spec never declared.
func checkContainerTarget(add func(string, string, ...any), ctx Context, t containerTarget) {
	section := t.Section
	// The section is gone: credentials are derived from the config's own
	// credential fields and delivered as platform secrets. yaml ignores unknown
	// keys, so without this an old env.yaml would parse and silently drop them.
	if t.Secrets != nil {
		add(fileEnv, "%s.secrets is no longer configured: credentials come from the connection fields themselves (client-password / client-password-env, password / password-env, ...) and are mounted at %s. Remove the %s.secrets section", section, secretsMountPath, section)
	}
	// Gone for the same reason: its one field could only ever hold the fixed
	// in-container path, and the host side always came from tls.*.file, so the
	// block decided nothing. The bind mount is now derived from those paths.
	if t.Stores != nil {
		add(fileEnv, "%s.stores is no longer configured: the tls.truststore.file / tls.keystore.file store files are bind-mounted at %s whenever they are set. Remove the %s.stores section", section, spec.DefaultStoresMountPath, section)
	}
	if !isDNS1123(t.Name) {
		add(fileEnv, "%s.name %q is not a valid DNS-1123 label", section, t.Name)
	}
	if err := CheckDeployCommand(t.Platform, t.Command, ctx.AllowCommands); err != nil {
		add(fileEnv, "%s.command %s", section, err)
	}
	if t.Image != "" {
		add(fileEnv, "%s.image is no longer configured here: the image moved to the top-level image: block (repo/name/tag) so one declaration serves every platform. Remove %s.image", section, section)
	}
	if t.Restart != "" && !SafeToken(t.Restart) {
		add(fileEnv, "%s.restart %q contains an unsafe character (no spaces, quotes, backslash, control chars, or shell metacharacters)", section, t.Restart)
	}
	if t.Timezone != "" {
		add(fileEnv, "%s.timezone is no longer configured here: the container timezone moved to the top-level timezone: key so one declaration serves every platform. Remove %s.timezone", section, section)
	}
	checkPorts(add, section, t.Ports)
	// The tls.*.file paths are always bind-mount sources in a docker or podman
	// artifact, so they are always gated -- an unsafe character would add content
	// to a compose YAML line or a quadlet Volume= directive that the spec never
	// declared. This runs unconditionally because the mount is no longer opt-in.
	// Kubernetes is still exempt: it embeds the store content in a Secret rather
	// than naming a host path, and it does not reach this function.
	for _, st := range []struct {
		field string
		store *spec.Store
	}{{"tls.truststore.file", ctx.Defaults.TLS.Truststore}, {"tls.keystore.file", ctx.Defaults.TLS.Keystore}} {
		if st.store == nil || st.store.File == "" {
			continue
		}
		if !safeHostPath(st.store.File) {
			add(fileEnv, "%s bind-mounts %s %q, which contains an unsafe character (no whitespace, quotes, control chars, or shell metacharacters)", section, st.field, st.store.File)
		}
	}
	if t.Libs != nil {
		// dir is the only key libs takes. The container side is fixed at
		// DefaultLibsMountPath by the image, which launches with that directory
		// literally on its classpath, so a custom mount-path put the jars where the
		// JVM never looks -- and nothing caught it, unlike stores.mount-path, which
		// at least failed loud.
		if t.Libs.MountPath != "" {
			add(fileEnv, "%s.libs.mount-path is no longer configured: the in-container libs path is fixed at %s, which is on the connector image's classpath. Remove %s.libs.mount-path", section, spec.DefaultLibsMountPath, section)
		}
		if t.Libs.Dir == "" {
			add(fileEnv, "%s.libs.dir is required when libs is set", section)
		} else if !safeHostPath(t.Libs.Dir) {
			add(fileEnv, "%s.libs.dir %q contains an unsafe character (no whitespace, quotes, control chars, or shell metacharacters)", section, t.Libs.Dir)
		}
	}
}

// checkDocker validates the docker (compose) section, including the compose
// project name.
//
// project-name is checked here rather than in checkContainerTarget because
// containerTarget is shared with podman, which has no project concept to hold
// one -- the mirror of how checkPodman keeps mode/scope out of the shared
// struct. It is held to the same DNS-1123 rule as name: docker compose's own
// grammar additionally allows an underscore and a trailing hyphen, so this is
// stricter than compose requires but can never accept a project name compose
// would then reject (an explicit one it dislikes is a hard error, not a
// sanitised value). Checked unconditionally, as name is: ParseEnv always
// defaults it, so an empty value means the section was built without defaults.
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
	if !isDNS1123(d.ProjectName) {
		add(fileEnv, "docker.project-name %q is not a valid DNS-1123 label", d.ProjectName)
	}
}

// checkPodman validates the podman section, including the quadlet scope.
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
	// The mode used to choose between a `podman run` script and a quadlet unit,
	// but only for generate -- deploy and remove always installed the unit. The
	// script was an artifact the tool could emit and then never manage, so it is
	// gone and the unit is the only output. Rejected rather than ignored: a
	// section asking for a script should be told it will get a unit.
	if p.Mode != "" {
		add(fileEnv, "podman.mode is no longer configured: generate emits the .container quadlet unit that deploy and remove install. Remove podman.mode")
	}
	// base-dir is required, not defaulted: it is where the mounted application.yml
	// and status script are written, and the path is baked into the unit's Volume=
	// lines. There is no safe default -- the quadlet directory would put generated
	// data among systemd's own units, and a guess would be silently wrong on a host
	// where it is unwritable. Better to be told once than to find files somewhere
	// unexpected.
	if p.BaseDir == "" {
		add(fileEnv, "podman.base-dir is required: it names the host directory the rendered application.yml and status script are written to and bind-mounted from. Relative paths resolve against env.yaml")
	} else if !safeHostPath(p.BaseDir) {
		add(fileEnv, "podman.base-dir %q contains an unsafe character (no whitespace, quotes, control chars, or shell metacharacters)", p.BaseDir)
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
