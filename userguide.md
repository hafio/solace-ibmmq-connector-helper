# solmq-conn-util -- User Guide

`solmq-conn-util` turns a folder of small, per-workflow YAML files plus one `env.yaml`
into one consolidated `application.yml` for the **Solace PubSub+ Connector for IBM
MQ**, generates the Kubernetes, Docker Compose, or Podman artifacts that run it, and
can apply or tear those down for you. You describe each message flow as its own small
file; the tool deduplicates connections into shared **binders**, numbers the
workflows, derives each binding's destination-type, wires **TLS/mTLS** from one
shared truststore/keystore, and turns every credential into a file mounted at
`/run/secrets/` -- so no credential value ever reaches a generated file (section 8).

All YAML shown here uses block style. Flow style (`{ }` / `[ ]`) also parses, but
`${VAR}` placeholders are invalid inside YAML flow `{ }`, so block style is
recommended everywhere.

> **Breaking change: platforms are no longer positional.** `generate`, `deploy`,
> `delete`, and `status` used to take the platform as a second positional argument
> (`deploy kubernetes`). That is now the `--platform` flag: `deploy --platform
> kubernetes`, or just `deploy` when `env.yaml` has exactly one of
> `kubernetes:`/`docker:`/`podman:`. The platform resolves in this order:
>
> 1. `--platform`, if given -- it must name a section actually present in
>    `env.yaml`, or the run fails loudly.
> 2. Otherwise, the single `kubernetes:`/`docker:`/`podman:` section present in
>    `env.yaml`, when exactly one is present.
> 3. Otherwise, an interactive numbered menu over the sections present, when more
>    than one is present.
> 4. Zero sections present is always a loud error, with or without `--platform`.
>
> The menu -- and, under `status`, its install-confirmation prompt (section 10) --
> refuse to block when stdin is not a TTY: both fail immediately with a hint to
> pass the flag explicitly instead of hanging. **CI and scripts must therefore
> pass `--platform` (and, for `status`, `--install`) rather than rely on either
> prompt.**

---

## 1. Running solmq-conn-util

You run `solmq-conn-util` as a command-line tool. Use the prebuilt binary for your
platform -- on Windows it is `solmq-conn-util.exe`, elsewhere `solmq-conn-util` (release
binaries are named like `solmq-conn-util-linux-amd64`, `solmq-conn-util-darwin-arm64`,
`solmq-conn-util-windows-amd64.exe`). The examples below write it as `solmq-conn-util`; on
Windows use `.\solmq-conn-util.exe`, or put the binary on your `PATH` and drop the
leading `./`. To build from source, see
[docs/DEVELOPMENT.md](docs/DEVELOPMENT.md).

```sh
solmq-conn-util                                     # no arguments: print usage
solmq-conn-util --help                              # print usage
solmq-conn-util <verb> <target> [-e env.yaml] ...   # run a command (see section 3)
```

`solmq-conn-util` reads one `env.yaml` (chosen with `-e`, default `env.yaml`) plus the
workflow files it discovers alongside it. `generate`, `validate`, and `examples`
are pure build-time steps: they read and write plain files and need no network,
broker, or cluster. `deploy`/`delete` additionally shell out to the target CLI
(`kubectl`/`oc`, `docker`, or `podman` + `systemctl`) to apply or tear down what was
generated -- run those where that CLI and its context are available.

### 1.1 Shell completion

`solmq-conn-util completion <shell>` prints a completion script on stdout for `bash`,
`zsh`, `fish`, or `powershell`. It completes the verbs, each verb's targets, the
flags valid under that verb, a file path after `-e`/`-o`, and a directory after
`examples` -- including when a flag comes before the target, as in
`solmq-conn-util generate -e env.yaml <TAB>`.

The script is rendered from the command model compiled into the binary, so it
always matches the commands that binary accepts. **Re-run the command after
upgrading `solmq-conn-util`** to pick up new commands.

Printing the script installs nothing -- it goes to stdout. What makes a shell
*use* it differs per shell, so follow the recipe for yours:

```sh
# bash -- add the source line to ~/.bashrc; depends on nothing but bash
source <(solmq-conn-util completion bash)

# fish -- writing the file IS the install; fish autoloads that path
mkdir -p ~/.config/fish/completions
solmq-conn-util completion fish > ~/.config/fish/completions/solmq-conn-util.fish

# zsh -- the file must be named _solmq-conn-util and sit on $fpath BEFORE compinit
mkdir -p ~/.zsh/completions
solmq-conn-util completion zsh > ~/.zsh/completions/_solmq-conn-util
#   ...then in ~/.zshrc, ahead of any existing compinit call:
#   fpath=(~/.zsh/completions $fpath)
#   autoload -Uz compinit && compinit
```

```powershell
# PowerShell -- Register-ArgumentCompleter is per-session, so persist it
solmq-conn-util completion powershell >> $PROFILE
```

Each shell has one thing that catches people out:

| Shell | Gotcha |
|-------|--------|
| bash | `/etc/bash_completion.d/` works only where the **bash-completion package** is installed and sourced from the profile -- it is what reads that directory. Without it the file is never loaded; the `~/.bashrc` `source` line has no such dependency. |
| zsh | `compinit` caches. If nothing completes, `rm -f ~/.zcompdump* && exec zsh`. |
| fish | None -- it autoloads. `solmq-conn-util completion fish \| source` covers the current session if you would rather not install. |
| PowerShell | `\| Out-String \| Invoke-Expression` is the current session only; it does not survive a restart. |

The scripts complete the command named `solmq-conn-util` (and `solmq-conn-util.exe` on
Windows). Release binaries carry their platform in the filename, so rename the one
you downloaded -- `solmq-conn-util-linux-amd64` to `solmq-conn-util` -- or completion will not
fire for it.

---

## 2. Quick start

```sh
solmq-conn-util examples examples                     # write a ready-to-edit sample set into ./examples
solmq-conn-util generate config -e examples/env.yaml  # print the application.yml those samples produce
```

Then edit the files under `examples/` (start with `examples/env.yaml`), drop your
`.jks` stores under `examples/certs/`, and re-run. See section 10 for the
`examples` command.

### The spec generator (no editor required)

If you would rather fill in a form than hand-write YAML, open
[solmq-conn-util-generator.html](solmq-conn-util-generator.html) in any browser -- it is a
single self-contained page, so there is nothing to install and no server to run.
It builds the whole spec folder for you:

- a form for every `env.yaml` section (section 6) and for each deploy target
  (section 7), plus repeatable cards for `connections` and for the workflow files
  (section 5), so a `conn-ref` is picked from a list rather than typed;
- live findings using the same rules and wording as `solmq-conn-util validate`
  (section 3), including the EDA advisories;
- a preview of the `application.yml` the spec consolidates into (section 9);
- **Download all (.zip)** writes `specs/env.yaml` plus one file per workflow,
  ready to unzip and hand to `-e specs/env.yaml`; **Load sample** fills in the
  same starter set `examples` writes (section 10).

Credentials are entered as the name of an environment variable (the `-env` form,
section 8) rather than as values, so the generated `env.yaml` stays safe to commit.
The preview is a convenience -- `solmq-conn-util generate config` remains authoritative.
The one case where the two are known to differ is `${VAR}`: the page cannot read
the environment the CLI will run in, so it previews the reference verbatim and
says so (section 4.1).

---

## 3. Commands

The first argument is a **verb**; the second names the **target** (for `generate`)
or **platform** (for `deploy`/`delete`).

```text
solmq-conn-util generate [config] [--platform kubernetes|docker|podman] [-e env.yaml] [-o out]
                                                                   Emit application.yml, or the resolved platform's artifacts
solmq-conn-util deploy  [--platform kubernetes|docker|podman] [-e env.yaml] [--allow-command name]
                                                                   Generate for the resolved platform, then apply it
solmq-conn-util delete  [--platform kubernetes|docker|podman] [-e env.yaml] [--allow-command name]
                                                                   Tear down what deploy created for the resolved platform
solmq-conn-util status  [--install] [--platform kubernetes|docker|podman] [-e env.yaml] [flags...]
                                                                   Ensure and run the status script; print per-target leader-election + workflow state
solmq-conn-util version                                           Print the utility name, version, Go version and OS/arch
solmq-conn-util validate            [-e env.yaml]                 Lint the whole env.yaml + workflows
solmq-conn-util examples [dir] [-f]                               Write a starter env.yaml + workflows
solmq-conn-util completion bash|zsh|fish|powershell               Print a shell completion script (section 1.1)
```

The full command tree -- with an example for every command -- is the generated
reference at [docs/commands.md](docs/commands.md).

| flag | applies to | meaning |
|------|-----------|---------|
| `-e`, `--env` | all except `examples` | config file, relative or absolute path (default: `env.yaml`) |
| `-o`, `--out` | `generate` | write output to a file (default: stdout) |
| `-f`, `--force` | `examples` | overwrite existing files |
| `--platform` | `generate`/`deploy`/`delete`/`status` | the platform: `kubernetes`, `docker`, or `podman` (default: resolved from `env.yaml`, or an interactive menu -- see the breaking-change note above) |
| `--allow-command` | `deploy`/`delete`/`status` | approve an extra command binary beyond the `command:` allowlist; repeatable |
| `--install` | `status` | install the status script on every target without prompting |
| `--pod` | `status` | limit checks to this kubernetes pod name; repeatable (default: every running pod) |
| `--container` | `status` | limit checks to this docker/podman container name; repeatable (default: every running container) |
| `--namespace` | `status` | kubernetes namespace to query (default: the deployment's namespace in `env.yaml`) |
| `--management-port` | `status` | actuator management port to reach inside each target (default: the management port configured for the target) |
| `--user` | `status` | actuator account the status script authenticates as (default `solmq-status`) |
| `--command` | `status` | override the platform CLI binary used to reach each target, instead of the `command:` in that section |

Flags may appear before, after, or between the positional arguments. Exit codes:
**0** success, **1** a processing error (bad input, unreadable file, missing env
var, a deploy command that failed), **2** a usage error (missing/unknown verb or
target, unknown flag). `status`'s own exit code is about whether every target
could be reached and run, not whether each instance is active -- see section 10
for how to read the per-target report.

- **`generate config`** reads the workflow files + the connector defaults from
  `env.yaml` and prints `application.yml`. It **fails fast**: it stops at the first
  error and writes nothing.
- **`generate` with no `config`** (or `generate --platform <platform>`) renders the
  resolved platform's deploy artifacts from the matching `env.yaml` section,
  resolving secret values / `.jks` bytes as needed (see section 8). It also
  **fails fast**.
- **`deploy`** generates for the resolved platform and then applies it by
  shelling out to the section's `command:` (`kubectl`/`oc`, `docker`, or
  `podman` + `systemctl`); **`delete`** tears the same thing down. The env file
  must contain the resolved section (a docker deploy needs `docker:`).
  `command:`'s binary must be on the platform allowlist (or approved with
  `--allow-command`), and both verbs run a read-only login/daemon preflight before
  writing or applying anything -- see section 7.
- **`status`** ensures the status script is present on each running instance of
  the resolved platform, runs it, and prints per-target leader-election and
  workflow state -- see section 10.
- **`version`** prints the build's own version (stamped in at build time), the Go
  version it was built with, and its `GOOS`/`GOARCH` -- for bug reports and to
  confirm which build is installed. Takes no flags.
- **`validate`** runs **every** check across the whole `env.yaml` (including any
  present `kubernetes:`/`docker:`/`podman:` sections) and its workflows, and prints
  all findings (non-zero exit if any errors). Use it as a linter.
- **`completion <shell>`** prints a completion script for `bash`, `zsh`, `fish` or
  `powershell` on stdout -- see section 1.1. It reads no `env.yaml` and touches
  nothing on disk.
- `generate` output is buffered and only written on full success -- you never get a
  half-written `-o` file.

---

## 4. The config file and workflow discovery

You point `solmq-conn-util` at one **`env.yaml`** with `-e` (default `env.yaml`). It
carries the connector defaults (section 6), a `workflows:` discovery block, and one
optional section per deploy target -- `kubernetes:`, `docker:`, `podman:` (section
7).

The `workflows:` block chooses which files around `env.yaml` are **workflow** files:

```yaml
workflows:
  dir: .                 # optional; folder to scan, relative to env.yaml or absolute; default "."
  file_pattern: "*"      # optional; matches workflow base names; only the '*' wildcard; default "*"
```

- The scan considers every `*.yaml` / `*.yml` file in `dir`, keeps those whose
  **base name** matches `file_pattern`, and **always excludes the env file itself**
  (regardless of `dir`/`file_pattern`).
- `file_pattern` supports **only the `*` wildcard** (leading, middle, and/or
  trailing), e.g. `workflow-*.yaml` or `*edge*.yml`. `?`, `[`, `]` are rejected with
  an actionable error.
- Relative paths in `env.yaml` (`workflows.dir`, `tls.*.file`, `libs.dir`,
  `values-file`, ...) resolve **relative to the env file's directory**, so a config
  folder is portable regardless of your current directory. `workflows.dir: "."`
  therefore means "alongside `env.yaml`".

Workflows are **numbered `0..N` by sorted filename**. The number becomes the
workflow's id in the output (`input-<N>` / `output-<N>` and
`solace.connector.workflows.<N>`). Naming files `workflow-0.yaml`,
`workflow-1.yaml`, ... keeps the mapping obvious, but any names work -- only sort
order matters. Filtering with `file_pattern` happens **before** numbering, so the
surviving files are numbered `0..N` among themselves.

The connector runtime holds **up to 20 workflows** (ids `0..19`) per
`application.yml`, so **one folder is one connector instance**. A folder holding more
than 20 is rejected with an error naming the count and the cap: split the workflows
across separate folders, each with its own `env.yaml` and its own
`deployment.name` / `docker.name` / `podman.name`, and deploy each as its own
connector. The tool does not split them for you -- which flows belong together is a
deployment decision, since an instance shares one leader election, one set of
credentials, and one resource budget.

This is an error rather than a warning because the connector does **not** complain
about the extra workflows: it binds ids `0..19` and silently ignores anything
numbered higher. A 21st workflow would generate cleanly, deploy cleanly, and simply
never run, with nothing at runtime saying why -- so the tool refuses up front
instead.

### 4.1 Variable expansion (`${VAR}`)

Any non-credential string field -- hosts, `msg-vpn`, `conn-name`, `queue-manager`,
`channel`, `cipher`, `key-alias`, destinations, names, namespaces, images, dirs,
mount paths, the syslog host, URLs, ... -- may reference the tool's own environment
at generate time:

```yaml
host: tcps://${BROKER_HOST}:55443
msg-vpn: ${VPN:prod}          # ${VAR:default} -- default used when VAR is unset
```

- `${VAR}` and `${VAR:default}` expand from the environment `solmq-conn-util` runs in.
  Only the braced form expands -- a bare `$VAR` is left as literal text.
- If `VAR` is unset and a default is given, the default is used.
- If `VAR` is unset and there is **no default**, the text is left **verbatim** and
  a **warning** names the variable. A typo must not vanish silently.
- **Credential fields never expand** (`client-username`/`-env`,
  `client-password`/`-env`, `user`/`-env`, `password`/`-env`, and the TLS store
  passwords): a credential is either a literal value or the name of a host
  variable in the matching `-env` field (see section 8.1) -- use that form instead of
  `${...}` inside a credential. A `${...}` inside a literal credential already
  triggers its own warning telling you to switch to `-env`.
- **Verbatim passthrough never expands** either -- `api-properties`,
  `additional-properties`, `consumer`, `producer`, `solace-defaults`,
  `logging.level` and the leader-election `fail-over` block are copied through
  untouched (section 5.4), so a `${...}` inside one reaches the connector as typed
  and is resolved by Spring at runtime, not by `solmq-conn-util` at generate time.
- **The generator page cannot expand.** A browser has no access to the
  environment `solmq-conn-util` will run in, so
  [solmq-conn-util-generator.html](solmq-conn-util-generator.html) previews a `${VAR}`
  verbatim and raises an advisory saying the generated file may differ. The
  `env.yaml` it writes is still correct -- expansion happens when you generate,
  not when you author.
- **Determinism caveat**: the tool's core promise is byte-for-byte reproducible
  output. Variable expansion is the one exception -- the rendered output now
  depends on the environment `solmq-conn-util` runs in, so the same `env.yaml` can
  render differently across machines or CI runs unless the referenced variables
  are pinned identically everywhere it runs.

---

## 5. Workflow file

One file describes one flow: consume from a **source**, produce to a **target**.

```yaml
enabled: true          # optional; default true. Set false to emit the workflow disabled.
source:                # exactly one of solace:/mq:, pointing at exactly one queue:/topic:
  mq:
    conn-name: mqhost.internal(1414)
    queue-manager: QM1
    channel: DSU.SVRCONN
    user: appuser
    password: ${MQ_PASSWORD}
    tls: true
    cipher: TLS_RSA_WITH_AES_256_CBC_SHA256
    key-alias: mq-client
    queue: ORDERS.OUT
target:
  solace:
    host: tcps://broker.internal:55443
    msg-vpn: prod
    client-username: connector
    client-password: ${SOL_PASSWORD}
    key-alias: solace-client
    queue: Q.FROM-MQ.ORDERS
```

Rules per side: **exactly one system** (`solace:` or `mq:`) and **exactly one
destination** (`queue:` or `topic:`). Any direction is allowed — `mq→solace`,
`solace→mq`, `mq→mq`, `solace→solace` — and every queue/topic combination is
permitted (two Solace patterns emit an advisory **warning**, see §5.5). You never write
`destination-type` or a durable name — both are derived.

### 5.1 Top-level

| field | required | default | notes |
|-------|----------|---------|-------|
| `enabled` | no | `true` | `false` emits the workflow but marks it disabled |
| `source` | yes | — | the consuming side |
| `target` | yes | — | the producing side |

### 5.2 `solace:` options

| field | required | notes |
|-------|----------|-------|
| `host` | yes | must start with `tcp://` (plaintext) or `tcps://` (TLS) |
| `msg-vpn` | yes | Solace message VPN |
| `client-username` | no | needed by most brokers; omit only for cert-only/OAuth auth |
| `client-password` | no | use a `${VAR}` placeholder; omit for cert-only/OAuth auth |
| `key-alias` | no | selects a client key from the shared keystore ⇒ **mTLS**; requires a `tcps://` host and a keystore in `env.yaml` |
| `queue` | one of | consume from / produce to a Solace queue |
| `topic` | one of | Solace topic; also allowed as a `source`, but a topic source warns (§5.5) |
| `api-properties` | no | verbatim map → `solace.java.api-properties` |
| `consumer` / `producer` | no | verbatim per-binding tuning |

### 5.3 `mq:` options

| field | required | notes |
|-------|----------|-------|
| `conn-name` | yes | `host(port)` — comma-separate for multi-instance QMs, e.g. `h1(1414),h2(1414)` |
| `queue-manager` | yes | |
| `channel` | yes | server-connection channel |
| `user` | no | omit for cert-based or channel (MCA) auth |
| `password` | no | use a `${VAR}` placeholder; omit when not using password auth |
| `tls` | no | `true` opts the connection into TLS (MQ has no URL scheme) |
| `cipher` | no | JCE cipher → `WMQ_SSL_CIPHER_SUITE`; requires `tls: true` |
| `key-alias` | no | client key from the shared keystore ⇒ **mTLS**; requires `tls: true` and a keystore |
| `queue` | one of | consume from / produce to an MQ queue |
| `topic` | one of | MQ topic; a `topic:` **source** is always a durable subscription (auto-named) |
| `additional-properties` | no | verbatim map → `ibm.mq.additional-properties` |
| `consumer` / `producer` | no | verbatim per-binding tuning |

> A `solace:`/`mq:` side may instead set **`conn-ref: <name>`** to reuse a connection
> from `env.yaml` (§5.6). A conn-ref side then sets *only* its `queue:`/`topic:` plus
> the per-binding `consumer:`/`producer:` tuning — any *connection* field is an error.

### 5.4 Destinations, durable names, passthrough

- The tool derives `destination-type` from whether you wrote `queue:` or `topic:`.
- An **MQ `topic:` source** always gets an auto `durable-subscription-name` (guaranteed
  delivery). The name is a stable UUIDv5 (namespace
  `6ba7f4e2-9c1d-5a3b-8e47-2f9a0c7d13e5`, key = `conn-name ‖ queue-manager ‖ topic ‖
  file-basename` joined by `0x1F`) — so **renaming a workflow file changes its durable
  name** and orphans the old subscription. Rename deliberately.
- `api-properties`, `additional-properties`, `consumer`, and `producer` are copied
  through **verbatim**, preserving key order and scalar quoting.

### 5.5 Event-driven guidance (warnings)

All four Solace↔MQ destination combinations are allowed. Two Solace patterns are
still generated, but flagged with a **warning** (never an error), because they run
against event-driven architecture (EDA) principles:

- **Solace `topic:` as a `source`.** Consuming directly from a topic is a direct,
  non-durable subscription (at-most-once): events published while this connector is
  down are lost. EDA's guaranteed-delivery / durable-state principle favors binding
  a Solace **queue** subscribed to the topic — the broker persists events, so the
  producer's uptime is decoupled from the consumer's and a restart never drops data.
- **Solace `queue:` as a `destination`.** Producing to a queue is point-to-point and
  couples the flow to one endpoint. EDA's publish-subscribe / loose-coupling
  principle favors publishing to a **topic** and letting the broker route to any
  subscribed queues, so producers stay unaware of consumers and new consumers can be
  added without touching the producer.

Warnings never block generation — use these patterns deliberately (for example,
best-effort telemetry from a topic source, or a controlled point-to-point handoff to
a queue). MQ topic/queue sources and destinations are never warned.

### 5.6 Reusable connections (`conn-ref`)

Define a connection once under `connections.<name>` in `env.yaml` (§6), then
reference it from a workflow side with `conn-ref`. This avoids repeating host,
credentials, and TLS on every workflow:

```yaml
# workflow file
source:
  mq:
    conn-ref: mq-archive
    queue: ARCHIVE.REPLAY.OUT
target:
  solace:
    conn-ref: prod-solace
    topic: archive/from-mq
```

- A `conn-ref` side is **strict** about *connection* fields: host, creds, tls, cipher,
  key-alias and api/additional-properties alongside `conn-ref` are an **error** — they
  belong on the connection itself. `queue:`/`topic:` and the per-binding
  `consumer:`/`producer:` blocks are the side's own and stay allowed.
- The referenced connection must exist and its system must match the side's
  `solace:`/`mq:` block.
- **Consolidation is by connection *details*.** Two sides that resolve to the same
  connection tuple — whether via `conn-ref` or written inline — collapse into a single
  binder. Only connections a workflow references become binders.
- **Binder names** come from the connection name (sanitized); purely-inline connections
  get generated `sol-conn-N` / `mq-conn-N` names. A clash between two different binders
  is disambiguated with `-2`/`-3`.

---

## 6. Connector defaults (`env.yaml` top level)

The **top level** of `env.yaml` holds the connector defaults -- the keys that shape
`application.yml`. Every section here is optional and holds **no secret values**
(only `${VAR}` placeholders). The `workflows:` block (section 4) and the deploy
sections (section 7) sit in the same file, alongside these keys.

```yaml
connections:                     # reusable connections, referenced by conn-ref (§5.6)
  prod-solace:
    solace:
      host: tcps://broker.internal:55443
      msg-vpn: prod
      client-username: connector
      client-password: ${SOL_PASSWORD}
      key-alias: solace-client
      api-properties:              # optional; verbatim -> solace.java.api-properties
        REAPPLY_SUBSCRIPTIONS: true
  mq-archive:
    mq:
      conn-name: mqhost-archive.internal(1414)
      queue-manager: QM_ARCHIVE
      channel: DSU.SVRCONN
      user: appuser
      password: ${MQ_ARCHIVE_PASSWORD}
      tls: true                    # TLS without key-alias => server-auth only (no mTLS)
      cipher: TLS_RSA_WITH_AES_256_CBC_SHA256
tls:
  truststore:                    # one shared truststore for ALL Solace + MQ TLS connections
    file: ./certs/truststore.jks
    password: ${TRUSTSTORE_PASSWORD}
    type: JKS
  keystore:                      # one shared keystore; only needed when a side sets key-alias (mTLS)
    file: ./certs/keystore.jks
    password: ${KEYSTORE_PASSWORD}
    type: JKS
logging:
  level:
    root: INFO
    com.solace.connector: INFO
management:
  port: 8090
  exposure: health,info,workflows
  health-show-details: always
security:
  enabled: true
  users:
    - name: healthcheck
      password: ${HEALTHCHECK_PASSWORD}
leader-election:                 # standalone | active_active | active_standby
  mode: active_standby
  queue: solmq-connector-mgmt    # exclusive management queue (active_* only)
  conn-ref: prod-solace          # solace management session (or inline `solace: {...}`)
  fail-over:                     # optional; emitted verbatim
    max-attempts: 5
    back-off-initial-interval: 1000
    back-off-max-interval: 10000
    back-off-multiplier: 1.5
solace-defaults:
  connect-retries: -1
  reconnect-retries: -1
```

| section | option | notes |
|---------|--------|-------|
| `connections.<name>` | `solace:`/`mq:` block | a reusable connection tuple (no destination); referenced by `conn-ref` (§5.6) |
| `tls.truststore` | `file`, `password`, `type` | the single shared truststore; `type` is `JKS` or `PKCS12` |
| `tls.keystore` | `file`, `password`, `type` | the single shared keystore; required only for mTLS (`key-alias`) |
| `logging.level` | `<logger>: <level>` | verbatim, order preserved → `logging.level` |
| `management` | `port` | → `management.server.port` |
| `management` | `exposure` | → `management.endpoints.web.exposure.include`; `leaderelection` and `workflows` are always appended on top of whatever you list (or Spring's own `health` default when you list nothing) -- see 6.1 |
| `management` | `health-show-details` | → `management.endpoint.health.show-details` |
| `security` | `enabled` | defaults to `true` when the `security` block is present |
| `security` | `users` | list of `{ name, password }` → `solace.connector.security`; a reserved account is always added on top -- see 6.1 |
| `leader-election` | `mode` | `standalone` (default; omitted from output), `active_active`, or `active_standby` |
| `leader-election` | `queue` | management queue; **required** for `active_*` |
| `leader-election` | `conn-ref` / `solace` | the Solace management **session** (`conn-ref` to a solace connection, or inline `solace:`); required for `active_*` |
| `leader-election` | `fail-over` | optional map, emitted verbatim under `leader-election.fail-over` |
| `solace-defaults` | `<key>: <value>` | merged verbatim into every Solace binder's `solace.java.*` (e.g. connect/reconnect retries) |

`active_active` and `active_standby` render a `solace.connector.management.leader-election`
block with the `queue` and a Solace `session` (TLS wired from the shared stores);
`standalone` (or an absent block) emits nothing. `standalone` requires `replicas: 1`;
`active_*` allow more.

**TLS is shared:** there is exactly one truststore and one keystore, referenced by
every TLS connection (Solace via api-properties, MQ via an ssl-bundle). Different
connections can present different client certificates by using a different
`key-alias` **within** that one shared keystore.

### 6.1 The reserved status account (`solmq-status`)

Every generated `application.yml` carries a read-only actuator account named
`solmq-status` whenever security is **effectively enabled** -- that is, the
`security` block is absent (its default) or present with `enabled: true`; only an
explicit `security.enabled: false` turns it off along with the rest of the
management endpoints' access control. It is what the generated `status` script
(section 10) authenticates as, so it exists whether or not you configured a
`security:` block yourself, and in every leader-election mode.

- **The name is reserved.** Naming a `security.users[]` entry `solmq-status`
  yourself is a `validate`/`generate` error -- choose a different name for your
  own accounts.
- **The password** comes from the optional host env var
  `SECURITY_USER_SOLMQ_STATUS_PASSWORD` at `generate`/`deploy` time, when it is
  set and non-empty; otherwise the tool generates one (32 lowercase hex
  characters from a CSPRNG). **A generated password rotates on every
  regenerate** -- it is not persisted or reused across runs -- so pin
  `SECURITY_USER_SOLMQ_STATUS_PASSWORD` if you need the value to survive a
  redeploy unchanged.
- **It is a literal, not a secrets-model credential.** Unlike every other
  password this tool handles (section 8), `solmq-status`'s password is rendered
  as a plain value inside the generated `application.yml` -- the status script
  reads it back out of that same mounted file at run time, so there is nothing
  for the secrets model to mount separately. That means the password is
  **readable by anyone who can read the artifact it lands in**: the Kubernetes
  ConfigMap, the Docker Compose file (which inlines `application.yml`), or the
  Podman on-disk `application.yml` -- which is why that last one is written with
  file mode `0600` (section 7.3) rather than the platform's normal default.
- **The account is GET-only** against the actuator endpoints it needs
  (`health`, `leaderelection`, `workflows`); it cannot mutate anything on the
  connector.
- **Never reuse a password you use elsewhere** for
  `SECURITY_USER_SOLMQ_STATUS_PASSWORD` -- the value lands as a literal in build
  artifacts that may be checked into a ConfigMap, shared as a compose file, or
  simply sit on a host's disk, so treat it as no more sensitive than that.

---

## 7. Deploy targets (`kubernetes:`, `docker:`, `podman:`)

Each deploy target is an optional top-level section of `env.yaml`. `generate
<target>` renders that target's artifacts, `deploy <platform>` renders and applies
them, and `delete <platform>` tears them down. A command whose section is absent
errors (a `deploy docker` needs `docker:`).

**The `command:` field** names the CLI each target shells out to (default `kubectl`
/ `docker` / `podman`). Put any extra global arguments inside it -- e.g. `command:
kubectl --context prod -n solace-connectors` or `command: oc`. The string is
tokenized to an **argv slice and executed directly (never via a shell)**, and
**every token is validated against a safe charset** -- shell metacharacters and
control characters are rejected with an error naming the offending token. Kubernetes
manifests are passed on **stdin**, never as arguments.

On top of the charset gate, `command:`'s first token (argv[0]) is checked against
a **per-platform binary allowlist** -- `kubectl`/`oc` for `kubernetes:`, `docker`
for `docker:`, `podman` for `podman:` (a trailing `.exe`/`.EXE` is stripped before
comparing, so Windows configs stay portable). It must be a **bare name resolved
from PATH** -- a token containing `/` (or, since backslash is already
charset-rejected, any path) is an error. Every later token must be **flag-shaped**
(`-x`, `--flag`, `--flag=value`, or the value belonging to the preceding flag) --
a literal `--` or a bare positional argument is rejected, since `solmq-conn-util`
appends its own subcommand (`apply -f -`, `compose up -d`, etc.) and a stray
positional would land in the wrong place. `validate` runs this same check on
`kubernetes.command` too, not just `docker`/`podman`.

Need a binary outside the allowlist (a `sudo` prefix, an internal wrapper)?
`deploy`/`delete` take a repeatable **`--allow-command <name>`** flag that approves
it for that invocation only -- authority lives with whoever runs the command, never
with whoever wrote `env.yaml`. `command: sudo podman` fails both `validate` and
`deploy podman` until you add `deploy podman --allow-command sudo`.

Before `deploy`/`delete` write or apply anything, they run a **read-only preflight
probe** against the target: kubernetes checks `auth can-i create|delete deployment`
(the verb `deploy`/`delete` will actually need); docker/podman check `info`
(daemon/socket reachability). A failing probe stops the run before any file is
written or any mutating command runs, with the underlying CLI error plus a login
hint (e.g. "log in or select a context first ... then re-run" for kubernetes,
a daemon-unreachable hint for docker/podman). There is no way to skip it -- a
failing preflight means the real command would fail anyway.

The real runner also resolves argv[0] with `exec.LookPath` before exec'ing (a
same-directory match via `exec.ErrDot` is rejected, matching Go 1.19+'s hardening)
and echoes the resolved path to stderr first, e.g. `exec: /usr/local/bin/kubectl
apply -f -`, so you can see exactly what ran.

**Trust model:** `env.yaml` is executable configuration, not passive data -- its
`command:` (plus manifests, compose files, and quadlet units it drives) runs with
your privileges. Review a config someone else handed you the way you would a
Makefile before running `deploy`/`delete` against it. `generate` is the dry-run: it
renders the same artifacts without shelling out to anything, so you can inspect
them first.

Credentials and stores use the **same schema across all three targets** -- see
section 8.

### 7.1 kubernetes

`generate kubernetes` emits the manifest set; `deploy kubernetes` pipes it to
`<command> apply -f -`, `delete kubernetes` to `<command> delete -f -`.

```yaml
kubernetes:
  command: kubectl               # or "oc", or "kubectl --context prod -n solace-connectors"
  deployment:
    name: solmq-connector        # DNS-1123 label
    namespace: solace-connectors # DNS-1123 label; also emitted as a kind: Namespace doc
    image: solace/solace-pubsub-connector-ibmmq:2.13.0
    replicas: 2                  # active_standby: 1 leader + standbys (standalone needs 1)
    resources:                   # requests and limits are set to the same value
      cpu: "1"
      memory: 1Gi
    timezone: Asia/Singapore     # -> container TZ env var
  service:
    enabled: true
    port: 8090                   # optional; defaults to the management port (§below)
  logging:                       # entirely optional
    syslog:
      host: syslog.corp
      port: 514
      protocol: udp              # udp (default) | tcp (needs the logstash-logback-encoder jar on the classpath)
  libs:                          # entirely optional; exactly one of pvc/download
    pvc:
      existing: jar-libs-pvc     # ...a PVC that already holds the IBM MQ jars
      # create:                  # ...or provision an NFS-backed PV + PVC
      #   name: jar-libs-pvc
      #   storage: 1Gi
      #   nfs:
      #     server: nfs1.corp
      #     path: /solace-libs
    # download:                  # ...or an initContainer wget's the jars at pod start
    #   urls:
    #   - https://repo1/ibmmq-1.jar
    #   - https://repo1/ibmmq-2.jar
    #   image: busybox:1.37
    #   pvc: jar-libs-pvc        # optional: download into this existing PVC instead of an emptyDir
  secrets:                       # entirely optional
    credentials:                 # env-var Secret (mounted via envFrom)
      create:
        name: solmq-credentials
        source: env              # read each variable from the tool's environment at deploy time
        variables:               # one per ${VAR} across the workflows + defaults
          - SOL_PASSWORD
          - MQ_CORE_PASSWORD
          - MQ_ARCHIVE_PASSWORD
          - EDGE_SOL_PASSWORD
          - TRUSTSTORE_PASSWORD
          - KEYSTORE_PASSWORD
          - HEALTHCHECK_PASSWORD
      # or read values from a file instead of the environment:
      # create:
      #   name: solmq-credentials
      #   source: file
      #   values-file: ./secrets.env
      # existing: my-credentials # ...or reference a pre-existing Secret (create XOR existing)
    stores:                      # truststore/keystore Secret (volume-mounted)
      create:
        name: solmq-tls          # base64-embeds the .jks files from env.yaml tls.*.file
      # existing: my-tls
```

| section | option | notes |
|---------|--------|-------|
| `deployment` | `name`, `namespace` | required; must be valid DNS-1123 labels; a `kind: Namespace` doc for `namespace` is always emitted first. Keep `name` short enough that the derived `<name>-config` stays within 63 chars |
| `deployment` | `image` | required |
| `deployment` | `replicas` | default `1`; `standalone` leader-election requires `1`, `active_*` allow more. Replicas are copies of the one connector, and leader-election picks the active one |
| `deployment` | `resources.cpu`, `resources.memory` | one value each; emitted as identical requests **and** limits (guaranteed QoS); a bare integer like `cpu: 1` is auto-quoted |
| `deployment` | `timezone` | sets the container `TZ` env var |
| `service` | `enabled`, `port` | emit a Service on this port |
| `logging.syslog` | `host`, `port`, `protocol` | optional; emits `logback-spring.xml` into the ConfigMap (mounted at `/app/external/classpath/logback-spring.xml`) plus `LOGGING_SYSLOG_*` env vars (appname = `deployment.name`); `protocol` is `udp` (default) or `tcp` -- **tcp requires the `logstash-logback-encoder` jar on the connector classpath** (provide it via `libs`) |
| `libs` | `pvc` \| `download` | optional; exactly one mode; makes the IBM MQ java libraries available at `/app/external/libs` (read-only) |
| `libs.pvc` | `create` / `existing` | `create` emits an NFS-backed PersistentVolume (`<name>-pv`) + PersistentVolumeClaim; `existing` references a pre-provisioned PVC (`create` XOR `existing`) |
| `libs.download` | `urls`, `image`, `pvc` | an initContainer `wget`s each URL into `/libs` at pod start; the shared volume is an `emptyDir` unless `pvc` names an existing PVC |
| `secrets.credentials.create` | `name`, `source` | `source: env` reads listed vars from the environment; `source: file` reads `KEY=VALUE` (or YAML) from `values-file` |
| `secrets.credentials.create` | `variables` | required for `source: env` |
| `secrets.credentials.create` | `values-file` | required for `source: file` |
| `secrets.credentials.existing` | `<name>` | reference a pre-existing Secret instead of creating one (`create` XOR `existing`) |
| `secrets.stores.create` | `name` | base64-embeds the `env.yaml` `tls.*.file` stores; requires a `tls.truststore` |
| `secrets.stores.existing` | `<name>` | reference a pre-existing stores Secret |

Omit a `secrets` block and that credentials Secret / env-file is not produced; omit
`stores` and the stores are not mounted.

### 7.2 docker

`generate docker` emits a `docker-compose.yml` with `application.yml` inlined under
compose `configs:`; `deploy docker` runs `<command> compose -f <file> up -d`;
`delete docker` runs `<command> compose -f <file> down`. Credentials render to an
env-file (`env_file:`, mode `0600`, never logged) written alongside the compose file;
a `stores:` block bind-mounts the `tls.*.file` host paths onto the fixed in-container
store dir `/app/external/classpath/truststores`, and `libs.dir` bind-mounts a host jar
directory to `/app/external/libs`.

```yaml
docker:
  command: docker                # or "docker --context foo"
  image: solace/solace-pubsub-connector-ibmmq:2.13.0
  name: solmq-connector
  restart: unless-stopped
  ports:
    - 8090                       # bare: publish to the same host port (8090:8090)
    - "8081:8090"                # or "host:container" to map a distinct host port
  timezone: Asia/Singapore
  secrets:
    credentials:                 # rendered as an env_file; never logged
      create:
        name: solmq-credentials
        source: env
        variables:
          - SOL_PASSWORD
          # ...one per ${VAR} used across the workflows + defaults
      # existing: solmq-connector.env   # ...or reference an existing env-file by name
  stores:                        # opt in to bind-mounting tls.*.file host paths
    mount-path: /app/external/classpath/truststores   # fixed in-container path; must be this value
  # libs:
  #   dir: ./libs                # host dir bind-mounted to /app/external/libs
```

### 7.3 podman

`generate podman` honors `mode:` -- `run` (default) emits a `podman run` script,
`quadlet` emits `.container` unit file(s). `deploy podman` / `delete podman` are
**always quadlet + systemctl**, regardless of `mode`. Because a quadlet unit or run
script cannot inline file content, `generate`/`deploy` also write the rendered
`application.yml` next to the unit/script and bind-mount it in. Credentials do not
go on disk: `deploy` loads each into **podman's secret store** and the unit mounts
it (`Secret=<name>,type=mount,target=<KEY>`). Requires **podman 4.5+**.

**Scope** (`quadlet.scope`) selects where units go and which systemd runs them:

- `auto` (default): root -> **system** (`/etc/containers/systemd/`, `systemctl`);
  non-root -> **user** (`~/.config/containers/systemd/`, `systemctl --user`).
- `user` / `system`: force one; `quadlet.dir` overrides the directory for the
  resolved scope.

`deploy podman` loads the credentials into podman's secret store, writes the units,
then `systemctl [--user] daemon-reload` and `start`; `delete podman` `stop`s,
removes the units, reloads, and removes the secrets. A `generate podman` run script
in `mode: run` carries a preamble that creates the same secrets from **your**
environment (`${NAME:?...}`) -- so the script itself holds no credential values.

```yaml
podman:
  command: podman
  mode: run                      # run (default) | quadlet -- controls `generate` only
  quadlet:
    scope: auto                  # auto | user | system
    dir: ""                      # overrides the default dir for the resolved scope
  image: solace/solace-pubsub-connector-ibmmq:2.13.0
  name: solmq-connector
  restart: unless-stopped
  ports:
    - 8090                       # bare: publish to the same host port (8090:8090)
    - "8081:8090"                # or "host:container" to map a distinct host port
  timezone: Asia/Singapore
  # No secrets: section -- credentials come from the connection fields themselves.
  stores:                        # opt in to bind-mounting tls.*.file host paths
    mount-path: /app/external/classpath/truststores   # fixed in-container path; must be this value
  # libs:
  #   dir: ./libs
```

Common docker/podman options: `image` (required); `name` (a DNS-1123 label -- it flows
into filenames, a systemctl unit, and the podman secret-store namespace); `restart`;
`ports` (a bare port, or `host:container` to map a distinct host port; each 1-65535);
`timezone` (container `TZ`); `stores` (opt in to bind-mounting the truststore/keystore;
the in-container path is fixed); `libs.dir`. Neither section takes a `secrets:` key --
setting one is an error naming the credential fields that replaced it.

---

## 8. Secrets model

**One mechanism on every platform: each credential becomes a file under
`/run/secrets/`, and the connector reads it as a property.** No credential value
and no host variable name ever appears in `application.yml`, in a compose file, in
a quadlet unit, or in any file this tool leaves on disk.

### 8.1 Declaring a credential

Every credential field is a pair -- give it a literal value, or name the host
environment variable that holds it. Never both:

| Literal | Environment reference |
|---------|-----------------------|
| `client-username` / `client-password` | `client-username-env` / `client-password-env` |
| `user` / `password` (mq) | `user-env` / `password-env` |
| `tls.truststore.password`, `tls.keystore.password` | `...password-env` |
| `security.users[].password` | `...password-env` |

```yaml
connections:
  prod-solace:
    solace:
      host: tcps://broker.internal:55443
      msg-vpn: prod
      client-username: connector          # not sensitive -> literal is fine
      client-password-env: SOL_PASSWORD   # read from the environment at deploy
```

An `-env` value is a bare variable name (`SOL_PASSWORD`), not `${SOL_PASSWORD}`.
Writing `${...}` there is an error, and a `${` inside the *literal* key produces a
warning pointing you at the `-env` form.

### 8.2 Stable names

The container-side name is derived from the config, never from your variable name,
so the same `application.yml` runs anywhere: `<BINDER>_CLIENT_USERNAME`,
`<BINDER>_CLIENT_PASSWORD`, `<BINDER>_USER`, `<BINDER>_PASSWORD`,
`TRUSTSTORE_PASSWORD`, `KEYSTORE_PASSWORD`, `SECURITY_USER_<NAME>_PASSWORD`,
`LEADER_ELECTION_CLIENT_USERNAME` / `_CLIENT_PASSWORD`. The rendered config
references `${PROD_SOLACE_CLIENT_PASSWORD}`; `SOL_PASSWORD` is only ever a
deploy-time input on your host.

Every generated `application.yml` therefore begins with:

```yaml
spring:
  config:
    import: optional:configtree:/run/secrets/
```

`optional:` keeps it inert where nothing is mounted. Note that OS environment
variables still outrank configtree files in Spring's precedence, and that a bare
`generate config` document is not runnable on its own -- nothing mounts the stable
names until a deploy target does.

### 8.3 How each platform delivers them

- **kubernetes**: `secrets.credentials.create.name` (or `existing:`) is mounted as
  a volume at `/run/secrets`, read-only, `defaultMode: 0400`. There is no `envFrom`
  -- credentials are never environment variables. The pod also sets
  `automountServiceAccountToken: false`, since the connector never calls the API and
  an automounted token would land in the same tree configtree reads. An `existing:`
  Secret must already use the stable names as its keys.
- **docker**: compose `secrets:` using the **environment provider**, so values are
  read from the environment `docker compose` itself runs with -- the CLI injects
  them into that child process only. Nothing is written to disk. Requires
  **Docker Compose v2.23.1+**.
- **podman**: `deploy` loads each credential into **podman's secret store**
  (values on stdin, never in argv) and units mount them with
  `Secret=<name>,type=mount,target=<KEY>`. Store entries are namespaced by the
  container name, since that store is shared across every project on the host.
  `delete` removes them. Requires **podman 4.5+**.

The `stores` wiring is unchanged and separate: the shared truststore/keystore is
base64-embedded into a Kubernetes Secret mounted at
`/app/external/classpath/truststores/`, or bind-mounted there from the host for
docker/podman (opt in with a `stores:` block; the in-container path is fixed).

There is **no `secrets:` section under `docker:` or `podman:`** -- credentials come
from the connection fields, so there is nothing to configure. Setting one is an
error, as is a `kubernetes.secrets.credentials.create` still carrying the removed
`source` / `variables` / `values-file` keys.

---

## 9. What gets generated

**`generate config` -> one multi-binder `application.yml`:** deduplicated
binders (connections sharing a broker/queue-manager tuple collapse into one binder),
numbered workflows, the mandatory `undefined` binder (always emitted, always last),
derived destination-types, auto `durable-subscription-name` for MQ topic consumers,
verbatim `api-properties` / `additional-properties` (the one place a `${VAR}`
survives into the output, for Spring to resolve at runtime -- everywhere else it is
expanded at generate time, section 4.1), and the Solace + MQ TLS/mTLS blocks. It is always a single document (a folder holding more
than 20 workflows is rejected -- see section 4). Point `-o` **outside** the workflow
folder so the output is not re-scanned as a workflow on the next run.

**Store paths differ by target.** `generate config` writes each truststore/keystore
`location` exactly as it appears in `env.yaml`, so you can run the connector wherever
those files already live. The container targets rewrite them to a mount path -- the
`application.yml` that kubernetes/docker/podman ship points at
`/app/external/classpath/truststores/<file>` because the stores are mounted there (a
Secret volume for kubernetes, a bind mount onto that same fixed dir for docker/podman).

**The status script and its labels ship on every target, every platform.** All
three deploy targets always carry the rendered status script (section 10) at
`/app/external/libs/status` inside the container -- there is no option to omit it
-- and always set the static label `solace-connector/le-mode` to the instance's
leader-election mode (`standalone` / `active_active` / `active_standby`). A second
label, `solace-connector/role: active`, is added only for `standalone` and
`active_active`, where the deploy artifact itself already determines which
instance is active; `active_standby` leaves the role for runtime (the connector's
own leader election) to decide, so no static value would stay accurate.

**`generate kubernetes` -> a multi-document manifest set:** a `Namespace` doc for
`deployment.namespace` (emitted first; applying it when the namespace already
exists is a no-op), then a `ConfigMap` mounting `application.yml` at
`/app/external/spring/config/application.yml` (via `subPath`) and the status script
under its own `status` key, mounted at `/app/external/libs/status`, a `Deployment`
(pod template carrying the `solace-connector/le-mode`/`role` labels above), an
optional `Service`, the credentials/stores `Secret`s, and any libs
`PersistentVolume`/`PersistentVolumeClaim`. Probes use a **tcpSocket**
check on the management port (basic-auth protects `/actuator/*`).
When any MQ-TLS binder exists the Deployment sets
`JAVA_TOOL_OPTIONS=-Dcom.ibm.mq.cfg.useIBMCipherMappings=false`. When
`kubernetes.logging.syslog` is set, the ConfigMap also gets a `logback-spring.xml` key
(mounted at `/app/external/classpath/logback-spring.xml`) and the Deployment gets
`LOGGING_SYSLOG_*` env vars. When `kubernetes.libs` is set, the connector gets a
volume mounted at `/app/external/libs`: either a PVC (`libs.pvc`, optionally
provisioned from an NFS PV+PVC) or an initContainer that `wget`s jars into an
`emptyDir`/PVC (`libs.download`) -- see section 7.

**`generate docker` -> a `docker-compose.yml`** with `application.yml` inlined
under compose `configs:`, the status script inlined as a second `configs:` entry
mounted at `/app/external/libs/status`, the credentials env-file written
alongside, bind mounts for stores/libs, and the `solace-connector/le-mode`/`role`
labels above on the service. **`generate podman` -> a `podman run` script**
(`mode: run`) **or a `.container` quadlet unit** (`mode: quadlet`); because
neither can inline file content, the rendered `application.yml` **and** the
status script are each written to disk next to the script/unit and bind-mounted
in read-only (credentials go to the platform's secret store, not to disk -- see
section 8), with the same labels applied via `--label`/`Label=`.

---

## 10. Status: which instance is active

`solmq-conn-util status` answers "which of my instances is the active one right
now?" by exec'ing into each running instance and asking its own Spring actuator
endpoints -- no separate monitoring stack, and no assumption made from outside
the container.

```sh
solmq-conn-util status --platform kubernetes -e env.yaml
```

The platform resolves the same way as `generate`/`deploy`/`delete` (section 3):
`--platform` if given, else the single section present in `env.yaml`, else an
interactive menu. `status` then discovers every running instance of that
platform -- every pod matching the deployment's `app=<name>` selector on
kubernetes, or the configured container name on docker/podman -- unless you
narrow it with repeatable `--pod` / `--container`.

### First run: installing the script

Every artifact `generate`/`deploy` produces already carries the status script
(section 9), but an instance that predates this feature, or one this tool never
deployed, may not have it installed. The first time `status` finds it missing on
a target, it asks once:

```
status script missing on solmq-connector-7d9f8c6b5-x2n4q -- install it now? [y/N]
```

Answer `y` to install it on every missing target for this run, or pass
`--install` up front to skip the prompt entirely. Declining (or a blank answer)
skips that target for this run without installing anything, and `status` exits
with code 1. **The prompt refuses to block when stdin is not a TTY** -- a CI job
or script must pass `--install` (or ensure the script is already installed), or
the run fails immediately with that guidance instead of hanging.

### Sample output

```
solmq-connector-7d9f8c6b5-x2n4q: leader-election mode: active_standby
  leader-election state: active
  workflow 0: STARTED
  workflow 1: STARTED
solmq-connector-7d9f8c6b5-k8m1r: leader-election mode: active_standby
  leader-election state: standby
  workflow 0: STOPPED
  workflow 1: STOPPED
```

One block per target: the leader-election **mode** (`standalone` /
`active_active` / `active_standby`, from `leader-election.mode`, section 6), the
leader-election **state** (`active` or `standby`) read from the instance's own
`/actuator/leaderelection`, then one `workflow <id>: <state>` line per workflow
from `/actuator/workflows`. A target that does not expose `workflows` in
`management.endpoints.web.exposure.include` still reports leader-election, with a
warning on stderr that per-workflow state is missing; one that does not expose
`leaderelection` reports nothing and explains why on stderr, since that is the
endpoint the whole report is built from. Either way the script still exits 0 --
see below.

### The manual alternative

`status` is a thin wrapper around a script the connector container can run on its
own. To check an instance yourself, without the CLI:

```sh
kubectl exec <pod> -- sh /app/external/libs/status
docker exec <container> sh /app/external/libs/status
podman exec <container> sh /app/external/libs/status
```

To sweep every pod behind a Deployment:

```sh
for p in $(kubectl get pods -l app=solmq-connector -o name); do
  echo "$p:"; kubectl exec "$p" -- sh /app/external/libs/status
done
```

The script **always exits 0**. The report goes to stdout, and anything that went
wrong -- unreachable actuator, an endpoint that is not exposed, no config to read
the account from, an unrecognized state -- goes to stderr. Nothing is encoded in
the exit status, so a non-zero exit from `kubectl exec` / `docker exec` /
`podman exec` means exactly one thing: the exec did not reach or run the
instance. A standby instance and a misconfigured endpoint never look like that.

To branch on the result, read the output rather than `$?`:

```sh
kubectl exec <pod> -- sh /app/external/libs/status | grep -q 'state: active'
```

The `status` command's own exit code follows the same idea: 0 when every target
could be reached and run, 1 when one could not -- never a statement about which
instance is active. Read the per-target report lines for that.

### Instances this tool did not deploy

`status` also works against a **foreign** instance -- one deployed by hand, or by
an older version of this tool -- as long as the status script and the reserved
`solmq-status` account (section 6.1) are present on it. Combine `--platform` with
explicit `--pod` / `--container` targets and `status` needs nothing else from
`env.yaml` at all; `--management-port` and `--user` cover a non-default setup (a
management port other than the connector default, or an actuator account under a
different name):

```sh
solmq-conn-util status --platform kubernetes --pod other-team-connector-0 \
  --namespace other-ns --management-port 9090 --user ops-readonly
```

`--command` overrides the platform CLI binary used to reach a target (instead of
the matching section's `command:`, which may not even exist for a foreign
instance), and `--allow-command` approves an extra one beyond the platform
allowlist, the same as `deploy`/`delete`.

The script does not assume this tool mounted the config. Inside the container it
looks for the connector's configuration the way Spring itself resolves it, taking
the first file that answers: `SPRING_CONFIG_LOCATION`, then
`SPRING_CONFIG_ADDITIONAL_LOCATION`, then `/app/external/spring/config/` (what
the connector image's own entrypoint passes), then `./` and `./config/`. The file
name follows `SPRING_CONFIG_NAME` (default `application`) and both `.yml` and
`.yaml` are tried; `classpath:` locations are skipped, since a shell cannot read
inside the jar. That is where it reads the account password from, and a
`${PLACEHOLDER}` value there is followed to `/run/secrets/<PLACEHOLDER>` -- so an
instance whose password still flows through the secrets model works too. If no
config can be found at all, the script says so and still queries the endpoint
unauthenticated, which is the right answer for an instance with
`security.enabled: false`.

---

## 11. `examples`

```sh
solmq-conn-util examples [dir] [-f]
```

Writes a ready-to-edit starter set into `<dir>` (default: the current directory):

| file | contents |
|------|----------|
| `env.yaml` | the single config file: `workflows` discovery, `connections` (prod-solace, mq-archive), shared TLS stores, logging/management/security, `active_standby` leader-election, and the `kubernetes:` / `docker:` / `podman:` deploy sections |
| `workflow-0.yaml` | IBM MQ -> Solace -- inline `mq-core` / `conn-ref: prod-solace` |
| `workflow-1.yaml` | Solace -> IBM MQ -- `conn-ref: prod-solace` / inline `mq-core` (same block as wf-0) |
| `workflow-2.yaml` | IBM MQ -> Solace -- `conn-ref: mq-archive` / `conn-ref: prod-solace` |
| `workflow-3.yaml` | IBM MQ (topic source) -> Solace -- inline `mq-core` (`topic:` => durable subscription) / inline `edge-solace` (plaintext) |

The set demonstrates all four connection styles at once -- a reference vs. inline
connection, each in a reused and a single-use flavour:

| connection | style | reuse | transport |
|------------|-------|-------|-----------|
| `prod-solace` | referenced (`conn-ref`) | reused (wf 0-2) | Solace mTLS |
| `mq-core` | manual / inline | reused (wf 0, 1, 3) -> one `mq-conn-1` binder | MQ mTLS |
| `mq-archive` | referenced (`conn-ref`) | single-use (wf 2) | MQ TLS, no mTLS |
| `edge-solace` | manual / inline | single-use (wf 3) -> `sol-conn-1` binder | Solace plaintext |

Every workflow is cross-platform (MQ<->Solace). `mq-core`'s identical inline block in
workflow-0, workflow-1 and workflow-3 collapses into one binder -- the manual
equivalent of a shared `conn-ref` -- while `mq-archive` (a different queue-manager)
stays its own binder. `workflow-3` consumes from an MQ **topic**, so it also exercises
the auto-named durable subscription (section 5.4). Only the *referenced* connections
live in `env.yaml`; the inline ones live on the sides that use them. Existing files are
left untouched (and reported) unless you pass `-f` / `--force`. The freshly written set
always generates cleanly:
`solmq-conn-util examples && solmq-conn-util generate config -e examples/env.yaml`.

---

## 11. Notes and gotchas

- **Deterministic, byte-for-byte output.** Regenerating from unchanged inputs
  produces identical bytes (an ordered emitter, not generic YAML marshaling), so
  files diff cleanly in review.
- **Workflow numbering is filename-driven**; sort order decides the ids and gaps in
  your naming are fine. A folder is capped at 20 workflows (ids `0..19`) -- past that
  the run fails and you split the folder yourself.
- **Renaming a workflow file changes its MQ durable subscription name** (it is part
  of the UUIDv5 key). Rename deliberately, or the old durable subscription is
  orphaned.
- **`generate` fails fast; `validate` reports everything.** Use `validate` while
  authoring, `generate`/`deploy`/`delete` to produce or apply output.
- **One shared truststore + keystore** for all connections; per-connection client
  certs are selected by `key-alias` within that one keystore.
- **TLS without a truststore emits no SSL bundle.** `tls: true` (or a `tcps://`
  host) with no `tls.truststore` still negotiates TLS, but falls back to the JVM's
  default trust store and warns — rather than writing a bundle with empty
  `location`/`password`/`type`, which the connector reads as configured-but-broken.
- **Two sides that share a connection tuple must share its password.** They collapse
  into one binder, so disagreeing passwords are an **error**, not a last-wins merge
  (a differing `cipher` still warns and takes the last value).
- **Values are quoted only when they need it.** A generated scalar carrying `": "`,
  `" #"`, a leading YAML indicator, or something that reads back as a bool/number
  (`no`, `0123`) is double-quoted; everything else stays plain, so output is stable.
  Multi-line passthrough values keep their block form.
- **The safe-charset gate covers more than `command:`.** `image`, `restart` and
  `timezone` in the docker/podman sections, the `tls.*.file` paths those sections
  bind-mount, `libs.dir`, the kubernetes Secret names, and `libs.pvc.create.nfs.*`
  are all rejected when they carry whitespace, quotes, control characters, or shell
  metacharacters — each one lands unquoted in a generated script, unit, or manifest.
- Multi-binder syntax is always used and the `undefined` binder is always emitted —
  this is expected, not a bug.
