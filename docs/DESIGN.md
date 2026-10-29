# solmq-conn-util design notes

> [!WARNING]
> **Not a supported Solace product.** `solmq-conn-util` was created by Solace
> Professional Services and is supported only by Solace Professional Services --
> not by Solace Support. For help with this tool, contact your Solace
> Professional Services representative rather than opening a Solace Support
> case. This notice covers this tool only, not the Solace PubSub+ Connector for
> IBM MQ that it configures and deploys.

Why the tool behaves the way it does. This is for someone changing it or judging
a change. For using it, read [getting-started.md](getting-started.md),
[configuration.md](configuration.md) and [operations.md](operations.md). For
building and releasing it, read [DEVELOPMENT.md](DEVELOPMENT.md).

## Contents

1. [Architecture](#1-architecture)
2. [CLI](#2-cli)
3. [Deploy execution](#3-deploy-execution)
4. [Kubernetes](#4-kubernetes)
5. [Docker](#5-docker)
6. [Podman](#6-podman)
7. [Container posture](#7-container-posture)
8. [TLS](#8-tls)
9. [Secrets](#9-secrets)
10. [Status, logs and cli](#10-status-logs-and-cli)
11. [download jar](#11-download-jar)
12. [Runner seams](#12-runner-seams)

## 1. Architecture

**Layered core.** The CLI (`cmd/solmq-conn-util`) is a thin shell over
`internal/gen`, which ties parse, validate, consolidate and render together.
Packages: `scan`, `spec`, `consolidate`, `tls`, `render`, `deploy`, `dockergen`,
`podmangen`, `runner`, `validate`, `examples`, `gen`, `libs`, `statusscript`,
`statusreport`, `yamlwriter`, `logback`, `storebuild`, `storescript`.
`certfixture` mints certificates, keys and JKS files for tests only.

**Byte-for-byte output.** `application.yml` and every artifact come from a
deterministic ordered emitter, not generic YAML marshaling, which would re-sort
keys. Regenerated files diff cleanly and match the golden fixture exactly.
Scalars are quoted only when YAML would misread them. `${VAR}` expansion is the
one exception: output then depends on the environment the tool runs in.

**One folder is one connector instance.** The connector binds workflow ids 0 to
19 and silently ignores anything higher. A 21st workflow would generate, deploy
and never run, with nothing saying why. The tool therefore refuses a folder with
more than 20 workflows instead of warning. Which flows belong together is a
deployment decision: an instance shares one leader election, one credential set
and one resource budget, so the tool never splits a folder for you.

**Workflow ids follow byte order.** Files are numbered in the order `LC_ALL=C
ls` lists them, so the ids are identical on every machine. A UTF-8 locale can
order punctuation differently. Earlier releases sorted numerically, which is why
an unpadded folder renumbers on upgrade.

**Solace topic sources are an error, not a warning.** The connector's Solace
consumer binds to an endpoint. A topic source generates, deploys and starts
cleanly, then never consumes, with nothing in the output saying why. Producing
to a Solace queue only warns: it works, it just couples the flow to one
endpoint against publish-subscribe practice.

**Transforms live in the workflow file.** Workflows are numbered by sorted file
name, so only the file knows which `<N>` it becomes. Both files are parsed
leniently, so a transform in any other place would be dropped silently. Every
key beginning `transform` outside a workflow file's top level is therefore an
error.

**Durable subscription names** are UUIDv5 (namespace
`6ba7f4e2-9c1d-5a3b-8e47-2f9a0c7d13e5`, key = `conn-name || queue-manager ||
topic || file-basename` joined by `0x1F`). The file name is part of the key, so
renaming a workflow file orphans its old subscription.

## 2. CLI

**Terminal help is shorter than the docs.** `-h` lists one line per verb; each
verb's page carries its flags and examples. Aliases work everywhere but appear
only in the generated reference, never in terminal help or TAB completion. The
completion menu lists the canonical verbs only, so an alias is a convenience for
people who already know it.

**Platform short spellings** (`kube`, `dk`, `pm`) are resolved before anything
reads the value, so errors name the `env.yaml` section, not the spelling. `kube`
is in common use; `dk` and `pm` are this tool's own, which is why the docs list
them.

**Prompts never block a non-TTY.** The platform menu, the `status` install
confirmation, the `remove` confirmation and an interactive `cli` session all fail
immediately when stdin is not a terminal, naming the flag that replaces them.
A CI job that hangs on a hidden prompt is worse than one that fails with a
message.

**`validate` reports everything; `generate` fails fast.** `validate` is the
linter and must surface every error `generate` or `deploy` would hit, TLS store
files included. `generate` stops at the first error and writes nothing. Its
output is buffered, so a failed run never leaves a half-written `-o` file.

**Exit codes are 0, 1 or 2**, except for `cli`. Once a session or command runs
inside the instance, its own status is passed back. The engines overload that
status: `kubectl exec` reports an unreachable pod and a failing command with the
same code, and docker adds 125, 126 and 127. `cli` does not guess between them;
the engine's stderr says which it was.

**The connector container is always named `connector`.** Every kubernetes `exec`
and `logs` argv names it with `-c`, built from the constant, never from
discovery. A pod with a sidecar cannot have the wrong container read, and a pod
without a `connector` container fails loudly. The cost is that a foreign pod
whose container is called something else is not reachable this way. Docker and
podman default `name` to the same word.

**One instance resolver.** `status`, `logs` and `cli` share
`cmd/solmq-conn-util/instances.go` for platform and namespace resolution, the
safe-token checks, command parsing and discovery. A debugging set of verbs that
disagreed about which instance they meant would be worse than any one alone.
Because the container name is a constant, naming an instance costs no extra
query.

**`logs` and `cli` read one instance.** When several match and none is named,
they print the matching commands, ready to paste, and exit 0. A guess or an
interleave would be worse than a list.

## 3. Deploy execution

**Never a shell.** `internal/runner` runs the CLI named by each section's
`command:` through an argv slice. Every config-derived token is checked against a
safe charset first. `validate.CheckDeployCommand` pins argv[0] to a bare,
per-platform allowlisted binary and requires later tokens to be flag-shaped,
since the tool appends its own subcommand and a stray positional would land in
the wrong place. argv[0] is resolved with `exec.LookPath`, rejecting an
`exec.ErrDot` same-directory match.

**Authority lives with whoever runs the command.** `--allow-command` approves an
extra binary for one invocation. It is a flag, not an `env.yaml` key, so a config
someone hands you cannot approve `sudo` for itself. `env.yaml` is still
executable configuration: its `command:` runs with your privileges, so review it
like a Makefile. `generate` is the dry run.

**Preflight is read-only and asks for the least it needs.** Docker and podman run
`info`. Kubernetes asks `auth can-i` for the one permission the verb needs
(`create deployment` for deploy, `get pods --subresource log` for logs, and so
on), so a read-only account can still run `status` and `logs`. A failing probe
stops the run before anything is written. There is no skip flag: the real
command would fail anyway.

**Secret material crosses on stdin, never argv**, and is never logged. Podman
secrets, kubernetes manifests and the compose file all travel that way.

**`remove` asks before it probes.** It names what it will tear down, with the
namespace or container name, because a run against the wrong `env.yaml` looks
identical until you read that line. Declining exits 0 having read nothing.
`deploy` has no prompt and rejects `--no-prompt`: it is additive and
re-runnable.

**The namespace is never in the manifest `remove` deletes.** Deleting a
Namespace cascades. A namespace holding anything the release does not own is
never removed, on any path, with or without `--no-prompt`. An empty one is
offered as a separate question, because consenting to a deployment's removal is
not consenting to the namespace's. Cluster namespaces are never removed, and a
check that cannot run leaves the namespace alone.

**Teardown reverses the manifest order.** kubectl deletes documents in file
order and waits on each. In creation order a `libs.pvc.create` claim precedes the
Deployment that mounts it, and `kubernetes.io/pvc-protection` holds the claim
while a pod still mounts it, so the delete deadlocks. Reversed, the workload goes
first. Every Secret the tool owns is reduced to kind and name, so `remove`
resolves no credential and opens no store.

## 4. Kubernetes

**`deploy` does not restart pods.** `application.yml` and the status script are
mounted from the ConfigMap with `subPath`, and kubernetes never refreshes a
`subPath` mount when its ConfigMap changes. The `restart` verb exists for this:
it runs `kubectl rollout restart`.

**Everything is named after `deployment.name`.** `<name>-config`,
`<name>-credentials`, `<name>-stores`, `<name>-image-pull`, `<name>-libs`. A
Deployment name is unique in its namespace, so these are too. Earlier releases
let each object take its own name, and two instances given the same one shared
it: each deploy overwrote the other's credentials, either `remove` deleted them
for both, and a shared libs claim made `remove` hang on `pvc-protection`. A
PersistentVolume is cluster-scoped, so its name also carries the namespace.

**Liveness and readiness are different checks.** Liveness is a `tcpSocket` probe
on the management port, because it decides whether kubelet restarts the
container, and a slow downstream must not become a restart loop. Readiness execs
the status script's `--health` mode, so `READY` means the connector reported
`UP`. The docker and podman healthchecks run the same mode, so the verdict means
the same thing everywhere.

**Credentials are a volume, never `envFrom`.** The Secret mounts at
`/app/external/var/secrets`, `0400`, and Spring reads it as a configtree. The
pod sets `automountServiceAccountToken: false`, since the connector never calls
the API and the token would land in the same tree. The files belong to root, so
`fsGroup` (the run-as gid) is what makes them readable to a non-root uid.

**`generate --platform kubernetes` embeds secrets.** The manifest has to be
complete for `kubectl apply`, so the credentials Secret and every store the tool
holds are rendered into it. With `-o` the file is written `0600`. The other
platforms never write a credential to disk.

**`image-pull.create` defaults to false.** Building a Secret is a mutation, and
naming one you manage must not overwrite it. It also keeps the registry account
optional in `env.yaml`.

**A store kept in a Secret you manage** (`existing.secret`) is read by the pod
directly. `deploy` checks that the Secret holds every key the pod reads before
applying, because `validate` cannot see the cluster. The tool never writes or
deletes such a Secret.

**Omitting `credentials` once a credential is set is an error.** The connector
would still start, since the configtree import is optional, and fail later
against whatever it could not authenticate to. `remove` only warns: a teardown
needs no credential.

## 5. Docker

**The compose file is handed to compose on stdin.** It inlines
`application.yml`, which carries the status account's password, so `deploy`
never writes it. `generate -o` writes it `0600` for a hand run.

**Secrets use compose's environment provider** and the long `secrets:` syntax.
The short form always lands at `/run/secrets/<name>`, and the connector reads
only `/app/external/var/secrets/`. The CLI injects each value into the compose
child process only. Requires Compose v2.23.1 or newer.

**A created store rides in an environment variable, base64.** Docker has no
secret store for a binary file, and an environment variable cannot hold NUL
bytes. The service starts through `/app/external/.stores-init`, which decodes
every such store into a `tmpfs` at `/app/external/classpath/truststores` on
every start and then runs the image's own command. Setting `entrypoint:` drops
the image's command, so the compose file carries it as `command:`; `deploy`
compares that with the image the engine reports and refuses an image that starts
any other way. A store too large for one environment variable (131072 bytes on
Linux, 32767 characters on Windows) fails the deploy by name.

**The `stores-sha256` label makes compose recreate on store change.** The stores
are not in the compose file, so compose's own change detection cannot see them.
The label carries a digest of every selected store's bytes, so `compose up`
recreates the container exactly when one changed. A bind mount pins an inode, so
a store replaced in place needs that recreate, not a restart.

**Every `$` in inlined `configs:` is doubled.** Compose interpolates across the
whole document and turns `$$` back into `$`. Without doubling, the status
script's `$PORT` would be blanked and `application.yml`'s `${...}` placeholders
would be substituted from the CLI's environment, writing plaintext credentials
into the file.

**The healthcheck is exec form and never restarts.** Exec form gives compose no
shell string to reinterpret. What to do about an unhealthy instance stays yours.

**`project-name` is declared in the file.** Without `name:` compose names the
project after the directory, which would make `status`'s grouping an accident of
where `env.yaml` lives. Compose resolves `-p`, then `COMPOSE_PROJECT_NAME`, then
the file, so an exported variable silently wins. `container_name` is host-unique
regardless of project, so two stacks need different `name` values too.

**`name` and `project-name` are DNS-1123 labels**, stricter than compose's own
rule, which also allows an underscore and a trailing hyphen. One grammar for
every name in the spec is worth more than those two spellings.

**Bind-mount sources are checked before `up`.** Compose creates a missing bind
source as an empty directory, so the connector would start without its store and
the directory would then sit where the PKCS12 belongs. A directory found there is
refused for the same reason.

## 6. Podman

**Everything the connector reads from a file comes from podman's secret store.**
A quadlet unit cannot inline content, and writing files beside `env.yaml` would
leave the status password on disk. `deploy` loads credentials, `application.yml`,
the status script, the logback config and every selected store as secrets, on
stdin, and writes only the `.container` unit. `libs.dir` is the one `Volume=`.
Secrets are named `<name>-<role>` so `deploy` can delete the ones a spec stops
mounting and `remove` can delete them all.

**`target=` is an absolute path.** Given a bare name podman places a mount secret
under `/run/secrets`, and the connector reads credentials only from
`/app/external/var/secrets/`. Requires podman 4.5 or newer.

**The unit directory is not configurable.** systemd scans only
`/etc/containers/systemd` (root) and `~/.config/containers/systemd` (everyone
else), so the deploy derives the directory from the invoking uid. A key could
only name a directory systemd ignores or the account cannot write.

**`restart` takes systemd's words.** The value lands in the unit as `Restart=`.
systemd never restarts a unit you stopped, so its `always` is what compose's
`unless-stopped` means; `unless-stopped` itself is an error on podman, since
systemd would ignore it and never restart the connector.

**Restart counts come from systemd.** Under quadlet a restart recreates the
container, so podman's own `RestartCount` reads 0 forever. `status` asks
`systemctl show -p NRestarts` and falls back to the container's counter where
systemd cannot answer.

**`deploy` restarts a running unit.** Podman copies a mount secret into the
container when it creates it, so only a new container sees the configuration
this deploy loaded. A stopped or failed unit is started. On podman 4.7 and newer
secrets are loaded with `--replace`, so a failed load leaves the previous secret
intact.

**`HealthOnFailure=` is left unset.** The check reports a verdict and never
restarts, kills or stops anything.

## 7. Container posture

**Posture is stated, never inherited.** The connector image declares no `USER`,
so left alone it runs as root, and on docker or rootful podman that is root on
the host. Every artifact writes the run-as identity (default `1000:0`) and a
fixed set of restrictions: kubernetes `allowPrivilegeEscalation: false`,
`privileged: false`, `capabilities.drop: [ALL]` and `seccompProfile:
RuntimeDefault`; compose `privileged: false`, `cap_drop: [ALL]`,
`security_opt: [no-new-privileges=true]`; quadlet `NoNewPrivileges=true` and
`DropCapability=all`. `TestArtifactsStateTheirPosture` pins what each platform
writes and `TestArtifactsCarryNoWideningTokens` forbids the spellings that would
widen it.

**`run-as: image` keeps the restrictions.** It drops only the identity. Root
without `CAP_DAC_OVERRIDE` reads only what the file modes allow, so deploy's
mount check (`cmd/solmq-conn-util/runas.go`) judges root by the mode bits like
any uid. Group 0 is the default gid because the image's own files belong to it.

**Rootless podman** maps uid 0 to the invoking user and higher ids into the
subordinate ranges, which `useradd` sizes at 65536. `deploy` reads `/etc/subuid`
and `/etc/subgid` and warns when the run-as ids need more than is granted.

## 8. TLS

**Stores are named and selected per connection.** Each connection picks its own
`truststore:` and `keystore:`, so Solace and MQ can trust different CAs and
present different certificates, and several queue managers can each have their
own. Earlier releases had one `tls.truststore` and `tls.keystore` for the whole
instance; those keys are now an error naming the replacement. A store no
connection selects is neither mounted nor built.

**The tool writes no certificate-validation policy.** Earlier releases wrote
`SSL_VALIDATE_CERTIFICATE: true` on every `tcps://` connection with a truststore
and dropped a connection's own `false`. Now the Solace API's default (validate)
applies unless the connection's `api-properties` say otherwise, and every
command warns when one turns it off.

**Every MQ binder carries `spring.main.sources:
org.springframework.boot.autoconfigure.ssl.SslAutoConfiguration`.** Spring Cloud
Stream runs a binder with its own `environment:` in a child context with no
parent: it inherits properties, `spring.ssl.bundle.*` included, but no beans and
no auto-configuration. Without that line the binder context has no SSL bundles,
the IBM MQ connection factory silently skips `ssl-bundle`, and the connection
uses the JVM's default SSL context, which a queue manager with a private CA or
mutual TLS refuses with `MQRC_JSSE_ERROR` (2397). A hand-written
`application.yml` with `ibm.mq` at the top level needs none of this, at the cost
of one queue manager per connector.

**`useIBMCipherMappings=false`** is set in `JAVA_TOOL_OPTIONS` whenever an MQ
binder uses TLS, because `cipher:` takes a JCE name. User `java-options.tool`
values are appended after it, so a user option that sets the same property wins.
IBM MQ 9.3.3 CD and 9.4.0 removed the property, so with the client `download
jar mq` fetches by default the flag is inert.

**Created stores are built, not made with `openssl pkcs12 -export`.** A
truststore made with `-nokeys` loads in Java as empty: Java keeps only
certificates carrying Oracle's trusted-certificate attribute, which OpenSSL
writes only from 3.2 with `-jdktrust`. The connection then fails with "the
trustAnchors parameter must be non-empty". A created store always carries the
attribute, and the tool warns about an existing truststore that lacks it.

**Created stores are deterministic.** PKCS12 with AES-256 (PBES2,
PBKDF2-HMAC-SHA-256, 10000 iterations) and an HMAC-SHA-256 MAC, readable by Java
12 and later. The bytes are a function of the password and the sources, so an
unchanged store builds to the same bytes and the kubernetes Secret stays
unchanged across deploys.

**A created keystore holds exactly one key.** The PKCS12 writer cannot name it,
so Java lists it as alias `1` and a `key-alias` on such a store is an error. One
store per client identity replaces a multi-key keystore.

**Store paths are the same everywhere.** `generate config` writes the
`application.yml` that `deploy` ships, so every store is read from
`/app/external/classpath/truststores/<store name>` on every platform, whatever
its host file is called. The mount follows the configuration and needs no
switch.

## 9. Secrets

**One mechanism everywhere.** Each credential becomes a file under
`/app/external/var/secrets/` and Spring imports the directory as an optional
configtree. No credential value reaches `application.yml`, a compose file, a
quadlet unit or any file the tool leaves on disk, apart from the kubernetes
manifest `generate` prints.

**`_GEN_` is a reserved namespace.** An `-env` credential mounts under the
variable name you wrote; a literal mounts under a derived name. Both land in one
directory, so the prefix is what stops a derived name from colliding with a
variable you exported. Two derived names can still fold together, since
punctuation collapses to `_`; that is rejected by name.

**The leader-election session shares the binder's credentials** when its broker
tuple matches one the workflows bind. The `_GEN_LEADER_ELECTION_*` pair is a
fallback for a session on a broker no workflow uses.

**The `solmq-status` password is a literal.** The status script reads it back
out of the mounted `application.yml` at run time, so there is nothing for the
secrets model to mount separately. Anyone who can read the artifact can read it:
the ConfigMap, the compose file, or the podman secret entry. A generated password
rotates on every regenerate unless `SECURITY_USER_SOLMQ_STATUS_PASSWORD` pins it.
The account is GET-only, and the tool never adds a role to it.

**Management security is always on and the exposure list is fixed**
(`health,info,metrics,leaderelection,workflows`). The status script and the
readiness probe depend on both, so neither is configurable.

## 10. Status, logs and cli

**Two halves from two places.** The container view reads the engine from
outside with read-only `get`/`inspect`. The application view runs the status
script inside each instance against its own actuator. The script cannot know
what surrounds it, so the banner (namespace, deployment, compose project) is
read from the engine, and the compose project from the container's own label
rather than guessed from a directory.

**The script always exits 0.** The report goes to stdout and every problem to
stderr, so a non-zero exec exit means exactly one thing: the exec did not reach
the instance. `--health` is the one exception and has to be asked for by name:
it makes one health call and exits 0 only on `UP`. It does none of the rest of
the report, because the full report makes seven actuator calls and starts a
JVM, far too much for a timer in a CPU-limited container.

**Health is read under either key order.** Spring Boot 3 (connector 2.x) writes
the instance's own status first; Spring Boot 4 (connector 3.x) sorts every key,
so the status comes last. The script finds the instance's own status either way,
never a component's.

**Degraded queries are notes, not exit codes.** `status` exits 1 only when an
instance could not be reached or run, or the run could not start. A standby, a
crash loop or a missing metrics API are answers, printed in the report. A
component that cannot be checked becomes a `status:` line in the same idiom the
in-container script uses.

**An instance that cannot answer still gets a block**, with the failure under
the container facts that explain it. The crash-looping instance is the one you
most need to see.

**One call for many instances.** Every engine query lists or inspects once for
the run, so cost barely grows with replica count. `-d` is opt-in because it adds
a sampling call (`top`, `stats`) and on kubernetes needs a metrics API.

**Progress goes to stderr and only stderr.** The report owns stdout, so
`--output json > report.json` produces the same bytes with a spinner, with `-v`
or with neither. The spinner draws with carriage return and spaces, no ANSI,
capped at 79 columns so a wrapped line never leaves residue above the cursor.
`--watch` prints no progress: the redraw owns the screen.

**JSON is the same model the tables render from**, so they cannot disagree.
Fields are a compatibility contract: additions keep `schemaVersion`, renames
bump it, and empty fields are omitted so presence is a meaningful test.

**Workflows are filtered and numerically ordered.** The connector reports every
slot and marks the unconfigured ones `N/A`; one real workflow would otherwise
print twenty lines. Real states are never filtered against a fixed list.

**The script prints heap in raw bytes.** Micrometer may report a large heap in
scientific notation, which busybox integer arithmetic would read as `4`. The
CLI converts.

**The script finds the config the way Spring does**: `SPRING_CONFIG_LOCATION`,
then `SPRING_CONFIG_ADDITIONAL_LOCATION`, then `/app/external/spring/config/`,
then `./` and `./config/`, with `SPRING_CONFIG_NAME` and both extensions. A
`${PLACEHOLDER}` password is followed into `/app/external/var/secrets/`. That
is what lets it report a foreign instance. With no config at all it queries
unauthenticated, which is right for a foreign instance with security disabled.

**`cli` opens `sh`.** The image is Alpine with busybox; there is no `bash` to
ask for, so there is no `--shell` flag. A one-shot command reaches the engine as
argv, never `sh -c`, so shell metacharacters are refused rather than misread.

**`cli` attaches the real terminal.** See `runner.Attacher` in
[section 12](#12-runner-seams).

## 11. download jar

**The `mq` seed is always `com.ibm.mq.jakarta.client`.** The connector image is
a Jakarta stack on every line, and `com.ibm.mq.allclient` (`javax.jms`) cannot
satisfy a `jakarta.jms` binder. Both builds carry the same packages, so with
both on one classpath load order decides, silently.

**Dependencies are pinned to the seed's POM**, not to each dependency's latest.
This keeps a multi-jar family such as BouncyCastle's three jars mutually
consistent. A version that cannot be resolved through the parent chain falls
back to that artifact's latest stable and is reported as a fallback.

**The seed is never omitted.** Omission applies only to dependencies. The seed
is the jar you ran the command for, the same command has to work against an
older image that never shipped it, and a crafted jar list cannot omit the one
jar that matters most.

**A jar list is a declaration, not something the command verifies.** Every
non-seed entry is trusted as written, so a list that does not describe the
image you deploy produces wrong omissions. That is why the command reads the
`image` block, picks the built-in list whose range holds your release, and
warns when none does. Deploying a covered release is silent, since a warning on
every correct run is noise.

**Matching is by jar filename, not groupId.** A filename carries no groupId, so
`com.fasterxml.jackson.core:jackson-databind` and
`tools.jackson.core:jackson-databind` share a base name. The version comparison
still tells the generations apart, because Jackson 3's version is higher.

**The syslog encoder follows the connector line.** `logstash-logback-encoder`
8.x is built on Jackson 2 (connector 2.x); 9.0 moved to Jackson 3 (connector
3.x). With no `--version` the command resolves the newest 8.x for a 2.x
connector and says why on its `seed:` line. The image already ships an encoder,
so two copies end up across the classpath; nothing in the tool confirms the
result loads with `protocol: tcp`.

**sha1 proves integrity, not authenticity.** The digest comes from the same host
as the jar, so it catches truncation and corruption, not a compromised
repository. A Maven-resolved artifact must verify; a `--url` artifact whose
sidecar 404s is written and reported unverified. https is mandatory on every hop,
with no opt-out.

## 12. Runner seams

`runner.Runner` is one method, `Run`, fully buffered: it returns once the process
has exited with stdout and stderr merged. Most Runners in tests are fakes that
record argv, so the other capabilities are separate interfaces asked for by type
assertion, and `var _ Streamer = OS{}` makes the production assertion a
compile-time fact.

- **`Streamer`** is for `logs --follow`, where output must arrive while the
  process runs. Its two writers are separate, so `logs > app.log` captures the
  log and leaves the platform CLI's diagnostics on the terminal. A cancelled
  context is reported as success: a follow the operator ended with Ctrl-C did
  not fail. `cmd.Cancel` interrupts, falling back to `Kill` on Windows, with
  `WaitDelay` as the backstop.
- **`Splitter`** returns stdout and stderr apart, both buffered. `runParsed`
  uses it for every helper that parses output (the five JSON readers,
  `KubernetesTop`, `EngineList`, `EngineStats`, `SystemctlNRestarts`), because a
  platform CLI writes warnings to stderr whenever it likes (OpenShift greets
  `oc get` with a deprecation notice), and merged into stdout that line lands in
  front of the JSON. `ScriptInstalled` and `RunStatusScript` still call `Run`:
  they tolerate an extra line and want the script's stderr advisories.
- **`Attacher`** is for `cli`, where the child must have the operator's real
  terminal. Its parameters are `*os.File`, because `os/exec` interposes a pipe
  for any other writer, a pipe is not a terminal, and the engine would refuse
  the pty. It returns the child's own exit status and reserves the error for a
  child that never started. No context: a session is ended from inside by
  typing `exit`. Ctrl-C is handled by not handling it: `signal.Notify` onto a
  channel nobody reads keeps the parent alive while the terminal delivers the
  interrupt to the child. Not `signal.Ignore`, whose `SIG_IGN` would be
  inherited across exec and leave the container's shell uninterruptible.

Every runner prints nothing of its own. The child's output is returned to the
caller, which reports it followed by `ok: <action> <platform>` or `error: ...`.
