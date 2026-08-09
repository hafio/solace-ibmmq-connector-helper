# solmq-conn -- User Guide

`solmq-conn` turns a folder of small, per-workflow YAML files plus one `env.yaml`
into one consolidated `application.yml` for the **Solace PubSub+ Connector for IBM
MQ**, generates the Kubernetes, Docker Compose, or Podman artifacts that run it, and
can apply or tear those down for you. You describe each message flow as its own small
file; the tool deduplicates connections into shared **binders**, numbers the
workflows, derives each binding's destination-type, wires **TLS/mTLS** from one
shared truststore/keystore, and passes secrets through as `${VAR}` placeholders
(never inlining real values).

All YAML shown here uses block style. Flow style (`{ }` / `[ ]`) also parses, but
`${VAR}` placeholders are invalid inside YAML flow `{ }`, so block style is
recommended everywhere.

---

## 1. Running solmq-conn

You run `solmq-conn` as a command-line tool. Use the prebuilt binary for your
platform -- on Windows it is `solmq-conn.exe`, elsewhere `solmq-conn` (release
binaries are named like `solmq-conn-linux-amd64`, `solmq-conn-darwin-arm64`,
`solmq-conn-windows-amd64.exe`). The examples below write it as `solmq-conn`; on
Windows use `.\solmq-conn.exe`, or put the binary on your `PATH` and drop the
leading `./`. To build from source, see
[docs/DEVELOPMENT.md](docs/DEVELOPMENT.md).

```sh
solmq-conn                                     # no arguments: print usage
solmq-conn --help                              # print usage
solmq-conn <verb> <target> [-e env.yaml] ...   # run a command (see section 3)
```

`solmq-conn` reads one `env.yaml` (chosen with `-e`, default `env.yaml`) plus the
workflow files it discovers alongside it. `generate`, `validate`, and `examples`
are pure build-time steps: they read and write plain files and need no network,
broker, or cluster. `deploy`/`delete` additionally shell out to the target CLI
(`kubectl`/`oc`, `docker`, or `podman` + `systemctl`) to apply or tear down what was
generated -- run those where that CLI and its context are available.

---

## 2. Quick start

```sh
solmq-conn examples examples                     # write a ready-to-edit sample set into ./examples
solmq-conn generate config -e examples/env.yaml  # print the application.yml those samples produce
```

Then edit the files under `examples/` (start with `examples/env.yaml`), drop your
`.jks` stores under `examples/certs/`, and re-run. See section 10 for the
`examples` command.

---

## 3. Commands

The first argument is a **verb**; the second names the **target** (for `generate`)
or **platform** (for `deploy`/`delete`).

```text
solmq-conn generate config     [-e env.yaml] [-o out]        Emit application.yml (per instance)
solmq-conn generate kubernetes [-e env.yaml] [-o out]        Emit ConfigMap+Deployment+Service (+Secrets)
solmq-conn generate docker     [-e env.yaml] [-o out]        Emit docker-compose.yml (application.yml inlined)
solmq-conn generate podman     [-e env.yaml] [-o out]        Emit a podman run script or quadlet unit(s)
solmq-conn deploy  kubernetes  [-e env.yaml]                 kubectl/oc apply -f - (manifest on stdin)
solmq-conn delete  kubernetes  [-e env.yaml]                 kubectl/oc delete -f -
solmq-conn deploy  docker      [-e env.yaml]                 docker compose up -d
solmq-conn delete  docker      [-e env.yaml]                 docker compose down
solmq-conn deploy  podman      [-e env.yaml]                 write quadlet unit(s); systemctl start
solmq-conn delete  podman      [-e env.yaml]                 systemctl stop; remove unit(s)
solmq-conn validate            [-e env.yaml]                 Lint the whole env.yaml + workflows
solmq-conn examples [dir] [-f]                               Write a starter env.yaml + workflows
```

The full command tree -- with an example for every command -- is the generated
reference at [docs/commands.md](docs/commands.md).

| flag | applies to | meaning |
|------|-----------|---------|
| `-e`, `--env` | all except `examples` | config file, relative or absolute path (default: `env.yaml`) |
| `-o`, `--out` | `generate` | write output to a file (default: stdout) |
| `-f`, `--force` | `examples` | overwrite existing files |

Flags may appear before, after, or between the positional arguments. Exit codes:
**0** success, **1** a processing error (bad input, unreadable file, missing env
var, a deploy command that failed), **2** a usage error (missing/unknown verb or
target, unknown flag).

- **`generate config`** reads the workflow files + the connector defaults from
  `env.yaml` and prints `application.yml`. It **fails fast**: it stops at the first
  error and writes nothing.
- **`generate kubernetes` / `docker` / `podman`** render that target's deploy
  artifacts from the matching `env.yaml` section, resolving secret values / `.jks`
  bytes as needed (see section 8). They also **fail fast**.
- **`deploy <platform>`** generates for that platform and then applies it by
  shelling out to the section's `command:` (`kubectl`/`oc`, `docker`, or
  `podman` + `systemctl`); **`delete <platform>`** tears the same thing down. The
  env file must contain the matching section (a `deploy docker` needs `docker:`).
- **`validate`** runs **every** check across the whole `env.yaml` (including any
  present `kubernetes:`/`docker:`/`podman:` sections) and its workflows, and prints
  all findings (non-zero exit if any errors). Use it as a linter.
- `generate` output is buffered and only written on full success -- you never get a
  half-written `-o` file.

---

## 4. The config file and workflow discovery

You point `solmq-conn` at one **`env.yaml`** with `-e` (default `env.yaml`). It
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
`application.yml`. With more than 20 workflows the tool automatically **splits them
across multiple connector instances** -- fill-to-20 in sorted order (instance 1 =
workflows 0-19, instance 2 = 20-39, ...). Each instance renumbers its own workflows
from `0`, and `generate` emits one set of artifacts per instance (see section 7).

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
> from `env.yaml` (§5.6). A conn-ref side then sets *only* its `queue:`/`topic:` —
> any other field is an error.

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

- A `conn-ref` side is **strict**: it may set only `conn-ref` + one `queue:`/`topic:`.
  Any connection field (host, creds, tls, cipher, key-alias, api/additional-properties)
  or per-binding tuning alongside `conn-ref` is an **error**.
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
| `management` | `exposure` | → `management.endpoints.web.exposure.include` |
| `management` | `health-show-details` | → `management.endpoint.health.show-details` |
| `security` | `enabled` | defaults to `true` when the `security` block is present |
| `security` | `users` | list of `{ name, password }` → `solace.connector.security` |
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
    port: 8090
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
| `deployment` | `name`, `namespace` | required; must be valid DNS-1123 labels; a `kind: Namespace` doc for `namespace` is always emitted first. With **>20 workflows** the name is suffixed `-1`, `-2`, ... per instance (single instance keeps the bare name); keep the base short enough that `<name>-<n>-config` stays within 63 chars |
| `deployment` | `image` | required |
| `deployment` | `replicas` | default `1`; `standalone` leader-election requires `1`, `active_*` allow more. **Replicas** are copies of one instance (leader-election picks the active one); **instances** are the separate Deployments created when workflows exceed 20 |
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
`application.yml` and the credentials env-file next to the unit/script and
bind-mount them in.

**Scope** (`quadlet.scope`) selects where units go and which systemd runs them:

- `auto` (default): root -> **system** (`/etc/containers/systemd/`, `systemctl`);
  non-root -> **user** (`~/.config/containers/systemd/`, `systemctl --user`).
- `user` / `system`: force one; `quadlet.dir` overrides the directory for the
  resolved scope.

`deploy podman` writes the units, then `systemctl [--user] daemon-reload` and
`start`; `delete podman` `stop`s, removes the units, and reloads.

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
  secrets:
    credentials:                 # rendered as --env-file / EnvironmentFile=; never logged
      create:
        name: solmq-credentials
        source: env
        variables:
          - SOL_PASSWORD
          # ...one per ${VAR} used across the workflows + defaults
      # existing: solmq-connector.env
  stores:                        # opt in to bind-mounting tls.*.file host paths
    mount-path: /app/external/classpath/truststores   # fixed in-container path; must be this value
  # libs:
  #   dir: ./libs
```

Common docker/podman options: `image` (required); `name` (a DNS-1123 label -- it flows
into filenames and a systemctl unit); `restart`; `ports` (a bare port, or
`host:container` to map a distinct host port; each 1-65535); `timezone` (container `TZ`);
`secrets.credentials`; `stores` (opt in to bind-mounting the truststore/keystore; the
in-container path is fixed); `libs.dir`.

---

## 8. Secrets model

Secret **values never appear** in `application.yml` -- the config keeps only
`${VAR}` placeholders. The `secrets.credentials` and `stores` schema is **identical
across all three targets**; only the rendering differs. Values are materialized at
generate/deploy time:

- **credentials** (`secrets.credentials.create`): resolves each value from the
  tool's environment (`source: env`) or a `values-file` (`source: file`).
  Kubernetes emits them in a Secret's `stringData` (referenced via `envFrom`);
  docker/podman write them to an **env-file** (`env_file:` / `--env-file` /
  `EnvironmentFile=`), created mode **0600** and **never logged**. An `existing:`
  reference names a Secret / env-file you manage yourself, so the tool writes
  nothing for it.
- **stores** (`secrets.stores`): the shared truststore/keystore. Kubernetes
  base64-embeds the `.jks` files named in `env.yaml` `tls.truststore.file` /
  `tls.keystore.file` into a Secret, volume-mounted at
  `/app/external/classpath/truststores/`; docker/podman **bind-mount** those files
  onto that same fixed dir instead (opt in with a `stores:` block; the path is not
  configurable, so a non-default `mount-path` is rejected).

For a **`kubernetes:`** section, `validate` additionally warns if a `${VAR}` used by
the workflows/defaults is not supplied by the credentials Secret, and if a TLS/mTLS
connection exists but no stores Secret is wired. These two warnings are
Kubernetes-specific.

---

## 9. What gets generated

**`generate config` -> one multi-binder `application.yml` per instance:** deduplicated
binders (connections sharing a broker/queue-manager tuple collapse into one binder),
numbered workflows, the mandatory `undefined` binder (always emitted, always last),
derived destination-types, auto `durable-subscription-name` for MQ topic consumers,
verbatim `api-properties` / `additional-properties`, `${VAR}` placeholders, and the
Solace + MQ TLS/mTLS blocks. With <=20 workflows this is a single document. With more,
it emits one `application.yml` per instance: to **stdout** each is preceded by a banner
(`CONNECTOR INSTANCE n OF N`); with **`-o out.yml`** they are written to `out-1.yml`,
`out-2.yml`, ... (a single instance still writes plain `out.yml`). Point `-o`
**outside** the workflow folder when sharding, so the suffixed outputs are not
re-scanned as workflows on the next run.

**Store paths differ by target.** `generate config` writes each truststore/keystore
`location` exactly as it appears in `env.yaml`, so you can run the connector wherever
those files already live. The container targets rewrite them to a mount path -- the
`application.yml` that kubernetes/docker/podman ship points at
`/app/external/classpath/truststores/<file>` because the stores are mounted there (a
Secret volume for kubernetes, a bind mount onto that same fixed dir for docker/podman).

**`generate kubernetes` -> a multi-document manifest set:** a `Namespace` doc for
`deployment.namespace` (emitted first; applying it when the namespace already
exists is a no-op), then **per instance** a `ConfigMap` mounting `application.yml` at
`/app/external/spring/config/application.yml` (via `subPath`), a `Deployment`, and an
optional `Service`; the credentials/stores `Secret`s and any libs `PersistentVolume`/
`PersistentVolumeClaim` are **shared** (emitted once). With >20 workflows the
ConfigMap/Deployment/Service names are suffixed `-1`, `-2`, ... per instance, and each
instance's leader-election coordination `queue` is suffixed to match (so the separate
connector clusters don't contend for one election queue). Probes use a **tcpSocket**
check on the management port (basic-auth protects `/actuator/*`).
When any MQ-TLS binder exists the Deployment sets
`JAVA_TOOL_OPTIONS=-Dcom.ibm.mq.cfg.useIBMCipherMappings=false`. When
`kubernetes.logging.syslog` is set, the ConfigMap also gets a `logback-spring.xml` key
(mounted at `/app/external/classpath/logback-spring.xml`) and the Deployment gets
`LOGGING_SYSLOG_*` env vars. When `kubernetes.libs` is set, the connector gets a
volume mounted at `/app/external/libs`: either a PVC (`libs.pvc`, optionally
provisioned from an NFS PV+PVC) or an initContainer that `wget`s jars into an
`emptyDir`/PVC (`libs.download`) -- see section 7.

**`generate docker` -> a `docker-compose.yml`** with `application.yml` inlined under
compose `configs:`, the credentials env-file written alongside, and bind mounts for
stores/libs. **`generate podman` -> a `podman run` script** (`mode: run`) **or
`.container` quadlet unit(s)** (`mode: quadlet`); because neither can inline file
content, the rendered `application.yml` and the credentials env-file are written next
to the script/unit and bind-mounted in. Both container targets shard past 20 workflows
the same way (one artifact set per instance).

---

## 10. `examples`

```sh
solmq-conn examples [dir] [-f]
```

Writes a ready-to-edit starter set into `<dir>` (default `examples/`):

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
`solmq-conn examples && solmq-conn generate config -e examples/env.yaml`.

---

## 11. Notes and gotchas

- **Deterministic, byte-for-byte output.** Regenerating from unchanged inputs
  produces identical bytes (an ordered emitter, not generic YAML marshaling), so
  files diff cleanly in review.
- **Workflow numbering is filename-driven**; sort order decides the ids and gaps in
  your naming are fine. More than 20 workflows are auto-split into additional
  connector instances (20 per instance), each renumbering from `0`.
- **Renaming a workflow file changes its MQ durable subscription name** (it is part
  of the UUIDv5 key). Rename deliberately, or the old durable subscription is
  orphaned.
- **`generate` fails fast; `validate` reports everything.** Use `validate` while
  authoring, `generate`/`deploy`/`delete` to produce or apply output.
- **One shared truststore + keystore** for all connections; per-connection client
  certs are selected by `key-alias` within that one keystore.
- Multi-binder syntax is always used and the `undefined` binder is always emitted —
  this is expected, not a bug.
