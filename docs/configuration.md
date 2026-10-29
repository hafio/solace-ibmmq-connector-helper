# solmq-conn-util configuration reference

> [!WARNING]
> **Not a supported Solace product.** `solmq-conn-util` was created by Solace
> Professional Services and is supported only by Solace Professional Services --
> not by Solace Support. For help with this tool, contact your Solace
> Professional Services representative rather than opening a Solace Support
> case. This notice covers this tool only, not the Solace PubSub+ Connector for
> IBM MQ that it configures and deploys.

Every key of `env.yaml` and of a workflow file. New here? Start with
[getting-started.md](getting-started.md). For the commands, see
[operations.md](operations.md); for known pitfalls,
[troubleshooting.md](troubleshooting.md); for why things are the way they are,
[DESIGN.md](DESIGN.md).

Write YAML in block style. Flow style (`{ }` / `[ ]`) parses, but `${VAR}` is
invalid inside flow braces.

## Contents

1. [env.yaml layout](#1-envyaml-layout)
2. [Workflow discovery](#2-workflow-discovery)
3. [Variable expansion](#3-variable-expansion)
4. [Workflow files](#4-workflow-files)
5. [Connections](#5-connections)
6. [TLS stores](#6-tls-stores)
7. [Connector defaults](#7-connector-defaults)
8. [Image, timezone and JVM options](#8-image-timezone-and-jvm-options)
9. [Platform sections](#9-platform-sections)
10. [Secrets](#10-secrets)

## 1. env.yaml layout

One `env.yaml` (chosen with `-e`, default `env.yaml`) holds everything but the
workflows:

| Top-level key | Purpose | Section |
|---|---|---|
| `workflows` | which files are workflow files | [2](#2-workflow-discovery) |
| `connections` | reusable connections, referenced by `conn-ref` | [5](#5-connections) |
| `tls.stores` | named truststores and keystores | [6](#6-tls-stores) |
| `logging`, `management`, `security`, `leader-election`, `solace-defaults`, `mq-defaults` | connector defaults | [7](#7-connector-defaults) |
| `image`, `timezone`, `java-options` | shared by every platform | [8](#8-image-timezone-and-jvm-options) |
| `kubernetes`, `docker`, `podman` | one section per platform you deploy to | [9](#9-platform-sections) |

Every section is optional. `env.yaml` holds no secret values, only literals you
choose to put there and `-env` variable names ([section 10](#10-secrets)).

Relative paths in `env.yaml` (`workflows.dir`, store files, `libs.dir`) resolve
against the env file's directory, so a config folder is portable.

## 2. Workflow discovery

```yaml
workflows:
  dir: .                 # folder to scan; default "." (beside env.yaml)
  file_pattern: "*"      # matches workflow base names; default "*"
```

- Every `*.yaml` / `*.yml` file in `dir` whose base name matches `file_pattern`
  is a workflow. The env file itself is always excluded.
- `file_pattern` supports only the `*` wildcard. `?`, `[` and `]` are errors.
- Workflows are numbered `0..N` by sorted file name, after filtering. The number
  is the workflow id in the output (`input-<N>`, `output-<N>`,
  `solace.connector.workflows.<N>`).
- The sort is byte order, as `LC_ALL=C ls` lists the folder: `workflow-10.yaml`
  sorts before `workflow-2.yaml`. Zero-pad past 9 (`workflow-00.yaml` ...
  `workflow-19.yaml`) to keep a file's number equal to its id.
- A folder holds at most 20 workflows (ids 0 to 19). More is an error. Split
  the workflows across folders, each with its own `env.yaml` and deployment
  name, and deploy each as its own connector.

> [!WARNING]
> Renaming or adding files renumbers the ones after them, and their bindings
> and `solace.connector.workflows.<N>` settings move with them. Durable
> subscription names come from the file's base name and do not change.

## 3. Variable expansion

Any non-credential string field may reference the environment `solmq-conn-util`
runs in:

```yaml
host: tcps://${BROKER_HOST}:55443
msg-vpn: ${VPN:prod}          # default used when VPN is unset
```

- Only the braced form expands. A bare `$VAR` is left as text.
- Unset with no default: the text is left verbatim and a warning names the
  variable.
- Credential fields never expand. Use the `-env` form instead
  ([section 10.1](#101-declaring-a-credential)). A `${...}` inside a literal
  credential warns.
- Verbatim passthrough never expands: `api-properties`,
  `additional-properties`, any other `solace:`/`mq:` key, `consumer`,
  `producer`, `transform`, `transform-headers`, `solace-defaults`,
  `mq-defaults`, `logging.level` and `leader-election.fail-over` reach the
  connector as typed, for Spring to resolve at run time.
- The generator page cannot expand: it previews `${VAR}` verbatim and says so.
- Expansion is the one exception to byte-for-byte output. Pin the variables
  identically everywhere the file is rendered.

## 4. Workflow files

One file is one flow: consume from `source`, produce to `target`.

```yaml
enabled: true          # optional; default true
source:                # exactly one of solace:/mq:, with exactly one queue:/topic:
  mq:
    conn-name: mqhost.internal(1414)
    queue-manager: QM1
    channel: DSU.SVRCONN
    user: appuser
    password-env: MQ_PASSWORD
    tls: true
    cipher: TLS_RSA_WITH_AES_256_CBC_SHA256
    truststore: mq-ca          # a tls.stores name
    keystore: mq-client        # presents a client certificate (mTLS)
    key-alias: mq-client
    queue: ORDERS.OUT
target:
  solace:
    host: tcps://broker.internal:55443
    msg-vpn: prod
    client-username: connector
    client-password-env: SOL_PASSWORD
    truststore: solace-ca
    keystore: solace-client
    key-alias: solace-client
    queue: Q.FROM-MQ.ORDERS
```

Each side names exactly one system and exactly one destination. Any direction
works. The one rejected shape is a Solace `topic:` as a source
([section 4.7](#47-destination-rules)). You never write `destination-type` or a
durable name; both are derived.

### 4.1 Top-level keys

| Key | Required | Default | Meaning |
|---|---|---|---|
| `enabled` | no | `true` | `false` emits the workflow disabled |
| `source` | yes | | the consuming side |
| `target` | yes | | the producing side |
| `transform` | no | | header and payload transforms ([section 4.8](#48-transforms)) |
| `transform-headers` | no | | deprecated header-only transforms; not with `transform` |

### 4.2 `solace:` side

| Key | Required | Meaning |
|---|---|---|
| `host` | yes | `tcp://` (plaintext) or `tcps://` (TLS) |
| `msg-vpn` | yes | message VPN |
| `client-username` | no | omit only for certificate or OAuth auth |
| `client-password` / `-env` | no | prefer `-env` ([section 10.1](#101-declaring-a-credential)) |
| `truststore` | no | `tls.stores` name the broker certificate is validated against; needs `tcps://`. Unset: the JVM's default CA bundle |
| `keystore` | no | `tls.stores` name holding the client certificate (mTLS); needs `tcps://` |
| `key-alias` | no | picks one key in that keystore; needs `keystore` |
| `queue` / `topic` | one of | `topic` is valid on a target only |
| `api-properties` | no | verbatim map to `solace.java.api-properties` |
| `consumer` / `producer` | no | verbatim per-binding tuning |
| any other key | no | passed through as `solace.java.<key>` ([section 4.6](#46-passthrough-keys)) |

Notes:

- For certificate authentication by the broker, also set
  `AUTHENTICATION_SCHEME: AUTHENTICATION_SCHEME_CLIENT_CERTIFICATE` in
  `api-properties`. The tool does not write it.
- The tool writes only store wiring into `api-properties`
  (`SSL_TRUST_STORE*`, `SSL_KEY_STORE*`, `SSL_PRIVATE_KEY_ALIAS`). Validation
  policy is yours: `SSL_VALIDATE_CERTIFICATE`,
  `SSL_VALIDATE_CERTIFICATE_HOST`, `SSL_VALIDATE_CERTIFICATE_DATE`
  ([section 6.3](#63-certificate-validation)).

### 4.3 `mq:` side

| Key | Required | Meaning |
|---|---|---|
| `conn-name` | yes | `host(port)`; comma-separate for multi-instance queue managers |
| `queue-manager` | yes | |
| `channel` | yes | server-connection channel |
| `user` / `-env` | no | omit for certificate or channel (MCA) auth |
| `password` / `-env` | no | prefer `-env` |
| `tls` | no | `true` enables TLS |
| `cipher` | no | JCE cipher name, to `WMQ_SSL_CIPHER_SUITE`; needs `tls: true` |
| `truststore` | no | `tls.stores` name; needs `tls: true`. Unset: the JVM's default CA bundle |
| `keystore` | no | `tls.stores` name holding the client certificate (mTLS); needs `tls: true` |
| `key-alias` | no | picks one key in that keystore; needs `keystore` |
| `queue` / `topic` | one of | a `topic:` source is a durable subscription, auto-named |
| `additional-properties` | no | verbatim map to `ibm.mq.additional-properties` |
| `consumer` / `producer` | no | verbatim per-binding tuning |
| any other key | no | passed through as `ibm.mq.<key>` ([section 4.6](#46-passthrough-keys)) |

Notes:

- `tls: true` alone does not negotiate TLS. The IBM MQ client needs a cipher
  suite, so without `cipher:` (or an `ssl-cipher-suite` / `ssl-cipher-spec`
  key) the connection is plaintext. `validate` warns. Name the JCE equivalent
  of the channel's `SSLCIPH`, for example `TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384`
  for `ECDHE_RSA_AES_256_GCM_SHA384`.
- `additional-properties` keys are `WMQ_*` constant names or raw `XMSC_*`
  names. `USER_AUTHENTICATION_MQCSP` must be written
  `XMSC_USER_AUTHENTICATION_MQCSP`. A value that reads as a number or boolean
  is typed as such.
- Common passthrough keys: `user-authentication-mqcsp: false` (pre-9.2.1
  password mode), `application-name`, `client-id`, `ccdt-url`, `reconnect`, a
  `pool:` block.

### 4.4 Reusing a connection (`conn-ref`)

A side may set `conn-ref: <name>` instead of connection fields, naming an entry
of `connections:` ([section 5](#5-connections)):

```yaml
source:
  mq:
    conn-ref: mq-archive
    queue: ARCHIVE.REPLAY.OUT
target:
  solace:
    conn-ref: prod-solace
    topic: archive/from-mq
```

- A `conn-ref` side carries only `queue:`/`topic:` and `consumer:`/`producer:`.
  Any connection field beside `conn-ref` is an error.
- The referenced connection must exist and match the side's system.

### 4.5 Binders and durable names

- Two sides with the same connection details, inline or by `conn-ref`,
  collapse into one binder. Only connections a workflow uses become binders.
- Binder names come from the connection name. Inline connections get
  `sol-conn-N` / `mq-conn-N`. A clash is disambiguated with `-2`, `-3`.
- Two sides on one binder must agree on password, `truststore`, `keystore` and
  `key-alias`. Disagreement is an error. A differing `cipher` warns and takes
  the last value.
- An MQ `topic:` source gets a stable, auto-generated
  `durable-subscription-name` derived from the connection, the topic and the
  file name. Renaming the file changes the name and orphans the old
  subscription.

### 4.6 Passthrough keys

Any key under `solace:` or `mq:` that the tool does not read itself passes
through verbatim, as `solace.java.<key>`, `ibm.mq.<key>`, or
`solace.connector.management.session.<key>` for the management session. There
is no allowlist: the connector owns that vocabulary and ignores what it does not
read. Rules, checked on every command:

| Found | Result |
|---|---|
| another spelling of a key the tool reads (`queueManager`, `clientPassword`) | error: Spring would bind it twice, or a credential would land as a literal |
| `ssl-bundle` | error: derived from `tls:` and the selected stores |
| a key that is not a plain property name (letters, digits, `. _ - [ ]`) | error |
| `ssl-cipher-suite` beside `cipher:` on one side | error: the cipher suite set twice |
| a key within a typo of one the tool reads | warning ("did you mean") |
| an `-env` suffix on a non-credential key | warning: it is a plain property |
| a secret-looking key with a literal value | warning: reference a mounted credential as `${NAME}` |
| `use-ibm-cipher-mappings` or `jks.*` on a `tls: true` side | warning: fights the tool's TLS wiring |
| a fixed `client-name` under `active_*` | warning: every replica shares it; use `${HOSTNAME}` |

Two sides on one binder merge their passthrough keys; the last by file name
wins a disagreement, with a warning.

### 4.7 Destination rules

| Shape | Result | Fix |
|---|---|---|
| Solace `topic:` as a source | error | consume from a queue subscribed to the topic |
| Solace `queue:` as a target | warning | publish to a topic and let the broker route |
| MQ `topic:` as a source | allowed; durable subscription auto-named | |
| any MQ queue or topic | allowed | |

Warnings never block generation.

### 4.8 Transforms

The connector's `transform:` block goes at the top level of the workflow file
and is written to `solace.connector.workflows.<N>.transform` verbatim:

```yaml
transform:
  source-payload:
    content-type: application/json
  target-payload:
    content-type: application/json
  expressions:
    - transform: "target['payload'] = source['payload']"
    - transform: "var['split'] = #splitString(source['payload']['passengers'], ',', 3)"
    - transform: "target['payload']['passengers'] = {:}"
    - transform: "target['payload']['passengers']['capacity'] = #convertStringToNumber(var['split'][0])"
```

| Key | Meaning |
|---|---|
| `source-payload.content-type` | `application/json`, or the default `application/vnd.solace.micro-integration.unspecified` (payload unreadable) |
| `target-payload.content-type` | the same for the target |
| `expressions` | ordered list of `- transform: "<SpEL>"` |

Behaviour worth knowing:

- No source header reaches the target unless an expression sets it.
- Setting `scst_targetDestination` routes the message to that destination.
- Until an expression writes `target['payload']`, the payload is copied through.
- Quote each expression. Unquoted, YAML cuts it at the first ` #`, which every
  `#function()` call has. That cut is reported as an error.
- The tool checks the block's shape, not the SpEL. Reference: Solace's
  [Message transforms](https://docs.solace.com/Micro-Integrations/Self-Managed/Message-transforms.htm)
  and [Transformation functions](https://docs.solace.com/Micro-Integrations/Self-Managed/Transformation-functions.htm).

Shape checks, on every command:

| Found | Result |
|---|---|
| `transform:` not a mapping, `expressions:` not a list, a malformed item, a key given twice | error, with the correct shape |
| an unknown key at any level | warning: passed through |
| a `content-type` other than the two above | warning: passed through |
| empty `expressions:`, or neither expressions nor a content type | warning: the block does nothing |
| `transform` or `transform-headers` under a side or its blocks | error: move it to the top level |
| `transform-payload(s)` anywhere | error: write payload expressions in `transform:` |
| any other `transform...` key, in a workflow file or in `env.yaml` | error: a typo, or the wrong file |

**`transform-headers` is deprecated** (by Solace since connector 2.9.0, still
read by 3.1.0). The tool passes it through and `validate` warns per file. A file
carrying both sections is an error. To migrate:

```yaml
# before
transform-headers:
  expressions:
    route: "T(String).format('%s/%s', headers.region, headers.status)"
    count: "headers.count.toString()"
```

```yaml
# after
transform:
  expressions:
    - transform: "target['headers']['route'] = #joinString('/', source['headers']['region'], source['headers']['status'])"
    - transform: "target['headers']['count'] = #convertNumberToString(source['headers']['count'])"
```

1. Each `<header>: <expr>` becomes `- transform: "target['headers']['<header>'] = <expr>"`.
2. `headers.<name>` becomes `source['headers']['<name>']`; Java methods become
   Solace functions (`T(String).join()` to `#joinString()`, `.split()` to
   `#splitString()`, `.toString()` to `#convertNumberToString()`).
3. Add an expression for every header the target still needs.
   `transform-headers` carried headers over by default; `transform:` carries
   none.

## 5. Connections

Define a connection once and reference it with `conn-ref`:

```yaml
connections:
  prod-solace:
    solace:
      host: tcps://broker.internal:55443
      msg-vpn: prod
      client-username: connector
      client-password-env: SOL_PASSWORD
      truststore: solace-ca
      keystore: solace-client
      key-alias: solace-client
      api-properties:
        REAPPLY_SUBSCRIPTIONS: true
  mq-archive:
    mq:
      conn-name: mqhost-archive.internal(1414)
      queue-manager: QM_ARCHIVE
      channel: DSU.SVRCONN
      user: appuser
      password-env: MQ_ARCHIVE_PASSWORD
      tls: true
      cipher: TLS_RSA_WITH_AES_256_CBC_SHA256
      truststore: mq-ca
```

A connection takes every key of a side except `queue:`/`topic:`. Identical
connections, referenced or inline, share one binder.

## 6. TLS stores

```yaml
tls:
  stores:
    solace-ca:                         # a PKCS12 you provide
      password-env: SOLACE_TRUSTSTORE_PASSWORD
      existing:
        pkcs12: ./certs/solace-ca.p12
    mq-ca:                             # built at deploy from certificate files
      password-env: MQ_TRUSTSTORE_PASSWORD
      create:
        certs:
          root: ./pem/mq-root.pem
          issuing: ./pem/mq-issuing.pem
```

| Key | Meaning |
|---|---|
| `<name>` | lowercase letters, digits and single hyphens, at most 63 characters; it names the mounted file |
| `password` / `password-env` | opens an existing store, encrypts a created one; required |
| `existing.pkcs12` | a PKCS12 file on the deploy host |
| `existing.secret` | kubernetes only: `{name, key}` of a Secret you manage that holds the store ([section 9.2](#92-kubernetes)) |
| `create` | built by `deploy` from certificate files or a JKS ([section 6.2](#62-created-stores)) |

Rules:

- A store has at most one host-side source, `existing.pkcs12` or `create`, and
  may name a Secret beside it. Kubernetes then reads the Secret; docker and
  podman use the host source. A store with only a Secret is kubernetes-only.
- A store with neither `create` nor `existing` is an error.
- Every store is PKCS12. The retired `tls.truststore`, `tls.keystore`, and a
  store's `file` / `type` are errors naming the replacement.
- A store no connection selects is neither mounted nor built, and warns.
- Inside the container every selected store is
  `/app/external/classpath/truststores/<name>`, on every platform.

### 6.1 Existing stores

Every verb that reads a PKCS12 (`validate`, kubernetes `generate`/`deploy`,
docker and podman `deploy`) opens it with its password first. It refuses a file
that is not PKCS12 (a renamed JKS, a PEM) or that the password does not open.
It warns about a truststore without Java's trusted-certificate attribute, which
Java loads as empty; an `openssl pkcs12 -export -nokeys` store made without
`-jdktrust` is the usual case. Let the tool build it instead.

### 6.2 Created stores

```yaml
tls:
  stores:
    broker-ca:                     # truststore from a list (aliases broker-ca-1, -2, ...)
      password-env: BROKER_CA_PASSWORD
      create:
        certs:
          - ./pem/broker-root.pem
          - ./pem/broker-issuing.pem
    partner-ca:                    # truststore from a map (each name is the alias)
      password-env: PARTNER_CA_PASSWORD
      create:
        certs:
          root: ./pem/partner-root.pem
    client-id:                     # keystore
      password-env: CLIENT_ID_PASSWORD
      create:
        keys:
          - cert: ./pem/client.pem   # may hold its chain
            chain:
              - ./pem/broker-issuing.pem
            key: ./pem/client.key    # unencrypted PKCS#1, PKCS#8 or SEC1
    legacy-ca:
      password-env: LEGACY_CA_PASSWORD
      create:
        jks: ./legacy/truststore.jks
```

| `create:` key | Takes | Makes |
|---|---|---|
| `certs` | a list or alias map of certificate files (PEM or DER, several per file) | a truststore, each certificate a trust anchor |
| `keys` | exactly one `{cert, chain?, key}` | a keystore holding that key with its chain, ordered by issuer |
| `jks` | a JKS opened with the store's password | a truststore of its trusted entries, or a keystore if it holds one private key and nothing else |

Rules:

- A created store is a truststore or a keystore, never both. Define a second
  store for the other role.
- Selecting a key-based store as `truststore:`, or a certificate-based one as
  `keystore:`, is an error.
- A created keystore holds one key, which Java names `1`. A `key-alias` on a
  connection selecting it is an error.
- Aliases ignore case; two that differ only in case are an error.

What is checked, and when:

- `deploy`, and `generate --platform kubernetes`, read the sources and the
  password, so both must be present. `validate` builds the store too, for every
  platform section, with a stand-in password when the variable is unset. A
  variable set but empty is an error. `generate config` never opens stores.
- Errors: a file with no certificate, an encrypted private key (decrypt with
  `openssl pkey -in key.pem -out plain.pem`), more than one key in a key file,
  no certificate matching the key, an expired client certificate, a JKS the
  password does not open, a JCEKS (convert with `keytool -importkeystore
  -srcstoretype JCEKS -deststoretype PKCS12`), a file that is already PKCS12
  (use `existing:`).
- Warnings: a certificate expired or expiring within 30 days, a certificate
  that is not a CA, a certificate outside the key's chain (left out), a
  repeated certificate (kept once).

Where a created store goes:

| Platform | Delivery |
|---|---|
| kubernetes | the `<name>-stores` Secret, or not built when the store also names `existing.secret` |
| docker | a base64 compose secret from deploy's environment, decoded into memory at every start |
| podman | the podman secret `<name>-tls-<store>` |
| `generate config` | referenced at its mount path, never built |

### 6.3 Certificate validation

A connection that selects no truststore validates against the JVM's default CA
bundle (an MQ binder that selects nothing gets no SSL bundle). One that selects
no keystore presents no client certificate.

The tool writes no validation policy for Solace. The API validates the broker
certificate by default. A `tcps://` connection whose `api-properties` set
`SSL_VALIDATE_CERTIFICATE: false` gets exactly that, truststore or not. Every
command reports it:

| Found | Result |
|---|---|
| `SSL_VALIDATE_CERTIFICATE: false` | warning naming the host and every side it reaches |
| a value other than `true`/`false` (`no`, `0`, `${VAR}`) | warning: posture cannot be reported |
| another spelling (`ssl_validate_certificate`) | warning: likely ignored by the API |
| two sides on one binder with different values | error |
| an inline management session differing from the workflow sides | warning: use `conn-ref` |

Prefer trusting the broker's CA. Reach for `SSL_VALIDATE_CERTIFICATE_HOST:
false` or `SSL_VALIDATE_CERTIFICATE_DATE: false` when only the host name or the
expiry is the problem.

### 6.4 Rotating a store

Replace the file (a created store's source or an existing PKCS12) and run
`deploy`. The connector reads stores only at start:

| Platform | After `deploy` |
|---|---|
| podman | the unit is restarted by `deploy` |
| docker | the container is recreated when any selected store changed |
| kubernetes | pods are not restarted: run `solmq-conn-util restart`, also after updating a Secret you manage |

## 7. Connector defaults

```yaml
logging:
  level:
    root: INFO
    com.solace.connector: INFO
  syslog:                        # optional; also ships log lines to syslog
    host: syslog.internal
    port: 514
    protocol: udp
management:
  port: 8090
  health-show-details: always
security:                        # optional; the tool's own read-only account is added regardless
  users:
    - name: ops
      password-env: OPS_PASSWORD
      roles:
        - admin                  # needed to POST to /actuator/workflows
leader-election:                 # standalone | active_active | active_standby
  mode: active_standby
  queue: solmq-connector-mgmt    # exclusive management queue (active_* only)
  conn-ref: prod-solace          # management session (or an inline session: block)
  fail-over:                     # optional; emitted verbatim
    max-attempts: 5
solace-defaults:
  connect-retries: -1
  reconnect-retries: -1
mq-defaults:
  user-authentication-mqcsp: false
```

| Key | Meaning |
|---|---|
| `logging.level` | verbatim, order preserved |
| `logging.syslog` | `host`, `port`, `protocol` (`udp` default, or `tcp`). Emits `logback-spring.xml` and `LOGGING_SYSLOG_*` env vars on every platform; appname is the instance name. `tcp` needs the `logstash-logback-encoder` jar on the classpath: fetch it with `download jar syslog` and ship it via `libs` |
| `management.port` | always emitted, default `8090`; the published port and the kubernetes `service.port` default to it. Keep it above 1023 |
| `management.health-show-details` | default `always`, so `/actuator/health` names each component; `when-authorized` or `never` restricts it |
| `security.users` | extra accounts: `{name, password or password-env, roles}`. Omit `roles` for a read-only account; add `admin` for read/write. Roles are passed through verbatim, not checked against a list |
| `leader-election.mode` | `standalone` (default; requires `replicas: 1`), `active_active`, `active_standby` |
| `leader-election.queue` | management queue; required for `active_*` |
| `leader-election.conn-ref` / `session` | the Solace management session; required for `active_*`. Exactly one. `session:` is a full Solace connection block; `queue:`, `topic:`, `consumer:` and `producer:` inside it are errors |
| `leader-election.fail-over` | optional map, emitted verbatim |
| `solace-defaults` | merged into every Solace binder's `solace.java.*` and into the management session. A connection's own key of the same name wins |
| `mq-defaults` | merged into every MQ binder's `ibm.mq.*`. A connection's own key wins. `channel` and `user` are rejected here |

Not configurable: management security is always on, and the actuator exposure
list is always `health,info,metrics,leaderelection,workflows`. An `env.yaml`
carrying `security.enabled` or `management.exposure` is rejected.

The management session reuses a binder's credential secrets when it points at a
broker the workflows already bind. Changing `solace-defaults` changes the
rendered session too, so the next `deploy` rolls the pods.

### 7.1 The reserved status account

Every generated `application.yml` carries a read-only actuator account named
`solmq-status`. The `status` script and the health probes authenticate as it.

- The name is reserved: a `security.users` entry called `solmq-status` is an
  error.
- The password comes from `SECURITY_USER_SOLMQ_STATUS_PASSWORD` when that
  variable is set and non-empty. Otherwise the tool generates one (32 lowercase hex characters from a
  CSPRNG), and a generated password changes on every regenerate. Pin the variable if the value
  must survive a redeploy.
- It is rendered as a literal inside `application.yml`, so anyone who can read
  the ConfigMap, the compose file or the podman secret can read it. Never reuse
  a password you use elsewhere.
- The account is GET-only on `health`, `leaderelection` and `workflows`.

## 8. Image, timezone and JVM options

These are declared once, at the top level. A per-platform `image:` or
`timezone:` is an error.

```yaml
image:
  repo: registry.internal:5000   # registry host only; omit for Docker Hub
  name: solace/solace-pubsub-connector-ibmmq
  tag: 2.13.0
  user: svc-puller               # only for kubernetes.secrets.image-pull.create
  pass-env: REGISTRY_PASSWORD
timezone: Asia/Singapore         # optional; the container's TZ
java-options:
  tool: >-                       # JAVA_TOOL_OPTIONS
    -Xms512m
    -Xmx${CONNECTOR_HEAP:1g}
    -XX:+HeapDumpOnOutOfMemoryError
    -XX:HeapDumpPath=/tmp
  jdk: --add-opens=java.base/java.lang=ALL-UNNAMED   # JDK_JAVA_OPTIONS
```

| Key | Meaning |
|---|---|
| `image.repo` | registry host and port only. A Docker Hub namespace such as `solace/` belongs in `name` |
| `image.name` | required; the repository path |
| `image.tag` | required. Pin it: `latest` moves between connector lines, and a `sha256:` digest pins exactly |
| `image.user`, `image.pass` / `-env` | registry account, used only to build a pull Secret ([section 10.4](#104-registry-credentials)) |
| `timezone` | unset: no `TZ` is set |
| `java-options.tool` | JVM options only (`-X`, `-XX`, `-D`, `-javaagent`); read by every JVM in the container |
| `java-options.jdk` | anything the `java` launcher accepts, `--add-opens` and `@argfiles` included; not `-jar` |

`java-options` rules:

- Each value is one string of space-separated options; a `>-` block is the same
  thing. A list is a parse error.
- Quotes, backslash, backtick, a `$` left after expansion and control characters
  are errors. A ` #` inside a `>-` block is an error (it would be part of the
  value).
- When an MQ binder uses TLS the tool sets
  `-Dcom.ibm.mq.cfg.useIBMCipherMappings=false` first and appends `tool` after
  it, so your value of the same property wins.
- Both variables make the JVM print a `Picked up ...` line at start.
- The run-as identity cannot write to `/`, so point a heap dump at `/tmp`.

### Connector 2.x and 3.x

| | 2.x | 3.x |
|---|---|---|
| Spring Boot | 3 | 4 |
| Java | 17 | 25 |
| Jackson | 2 | 3 |
| Solace Spring Cloud | 5 | 6 |

Driven by `tag`: `download jar` picks the jar list and the syslog encoder line
for the release you deploy, and the status script reads health under either
Spring Boot's key order. The generated `application.yml` is the same for both.
Solace Spring Cloud 6 changes some message headers (`contentType` no longer
mapped to `httpContentType`, `solace_scst_messageVersion` and `SOURCE_DATA` no
longer set); check your `api-properties` and downstream consumers.

## 9. Platform sections

Each section is optional. `generate` renders the resolved platform's artifacts,
`deploy` applies them, `remove` tears them down
([operations.md](operations.md#1-commands)). A run whose resolved section is
absent is an error.

### 9.1 Keys common to every platform

| Key | Meaning |
|---|---|
| `command` | the CLI to shell out to (default `kubectl` / `docker` / `podman`). Extra global arguments go inside it: `kubectl --context prod`, `oc` |
| `deployment.name` / `name` | the instance name, a DNS-1123 label (docker and podman default to `connector`). Every object the tool creates is named after it, and `status`, `logs` and `cli` find instances by it |
| `run-as` | `"<uid>:<gid>"`, `"<uid>"` (gid 0), or `image`; default `1000:0` ([section 9.5](#95-run-as-and-hardening)) |
| `libs` | where the IBM MQ jars come from; fill with `download jar mq` ([operations.md](operations.md#10-download-jar)) |

`command` rules: the first token must be a bare binary on the platform's
allowlist (`kubectl`/`oc`, `docker`, `podman`), later tokens must be
flag-shaped, and every token is checked against a safe charset. Another binary
(`sudo`, a wrapper) needs `--allow-command` on the command line.

No section takes a `secrets:` block for application credentials (kubernetes
takes one for how they are delivered), and none configures in-container paths.

### 9.2 kubernetes

```yaml
kubernetes:
  command: kubectl
  deployment:
    name: solmq-connector
    namespace: solace-connectors
    replicas: 2                  # standalone needs 1
    resources:                   # requests and limits are the same value
      cpu: "1"
      memory: 1Gi
    node-selector:               # optional
      kubernetes.io/os: linux
    tolerations:                 # optional; Kubernetes' own shape
      - key: dedicated
        operator: Equal
        value: solace
        effect: NoSchedule
    spread: preferred            # optional; none (default) | preferred | required
    # affinity: {}               # optional; the pod's affinity block, written as it stands
  service:
    enabled: true
    port: 8090                   # optional; bare or "host:container"; defaults to the management port
  run-as: "1000:0"
  libs:                          # optional; exactly one of pvc / download
    pvc:
      existing: jar-libs-pvc
      # create:                  # or an NFS-backed PV + PVC named <name>-libs
      #   storage: 1Gi
      #   nfs: {server: nfs1.corp, path: /solace-libs}
    # download:                  # or an initContainer that wget's the jars at start
    #   urls: [https://repo1/ibmmq-1.jar]
    #   image: busybox:1.37
    #   pvc: jar-libs-pvc        # optional; otherwise an emptyDir
  secrets:
    credentials:
      create: true               # create XOR existing
      # existing: my-credentials
    # image-pull:
    #   name: regcred            # or create: true
```

| Key | Meaning |
|---|---|
| `deployment.name`, `namespace` | required DNS-1123 labels. A `kind: Namespace` document is always emitted first. `<name>-credentials` and `<namespace>-<name>-libs-pv` must stay within 63 characters |
| `deployment.replicas` | default `1`; `standalone` requires `1` |
| `deployment.resources` | `cpu` and `memory`, Kubernetes quantities, written as identical requests and limits |
| `deployment.node-selector` | node labels the node must carry; keys written sorted |
| `deployment.tolerations` | Kubernetes' own fields; `validate` applies the API server's rules |
| `deployment.spread` | pod anti-affinity on `kubernetes.io/hostname`. `required` leaves replicas beyond the node count `Pending` |
| `deployment.affinity` | written as it stands; `validate` checks only the outline and refuses `podAntiAffinity` together with a `spread` other than `none` |
| `service.enabled`, `port` | a Service on this port; defaults to the management port |
| `libs.pvc.existing` | a PVC that already holds the jars; never deleted by `remove` |
| `libs.pvc.create` | `storage` (default `1Gi`) and `nfs.server` / `nfs.path`: an NFS PersistentVolume `<namespace>-<name>-libs-pv` and claim `<name>-libs`. Capacity and NFS settings are immutable once created; `deploy` warns when it re-applies an existing PV |
| `libs.download` | `urls`, `image` (default `busybox:1.37`), optional `pvc`; an initContainer wgets each URL into `/libs` |
| `secrets.credentials` | `create: true` builds `<name>-credentials` from the resolved values; `existing: <Secret>` mounts one you manage, which must carry the keys of [section 10.2](#102-mount-names). Required once any credential is set |
| `secrets.image-pull` | `name` references a pull Secret you made; `create: true` builds `<name>-image-pull` from the `image` block ([section 10.4](#104-registry-credentials)) |
| `secrets.stores` | retired; see [troubleshooting.md](troubleshooting.md#upgrading-from-an-earlier-release) |

TLS stores need no kubernetes setting. Every store a connection selects reaches
the pod in one projected volume at `/app/external/classpath/truststores`, mode
`0400`, readable through `fsGroup`:

- A store the tool holds (`existing.pkcs12` or `create`) travels in
  `<name>-stores`. Over 256 KiB, the limit of client-side `kubectl apply`, the
  run stops and names the store.
- A store in a Secret you manage names it:

  ```yaml
  tls:
    stores:
      broker-id:
        password-env: BROKER_ID_PASSWORD
        existing:
          secret:
            name: broker-cert          # e.g. a cert-manager Certificate's secretName
            key: keystore.p12          # defaults to the store's name
  ```

  Only a PKCS12 key works; cert-manager's `keystores.jks` output cannot be
  read. The store's password stays a credential: export it under the `-env`
  name or carry it in your `credentials.existing` Secret. Keep `pkcs12:` or
  `create:` beside `secret:` when a `docker:` or `podman:` section reads the
  same file. `deploy` checks the Secret holds every key before applying.

> [!IMPORTANT]
> **Redeploying does not restart the pods.** `application.yml` and the status
> script are `subPath` mounts, which kubernetes never refreshes. After a
> `deploy` that changes either, or after rotating a store, run:
>
> ```sh
> solmq-conn-util restart --platform kubernetes -e env.yaml
> ```

Everything the tool creates is named after `deployment.name`: `<name>-config`,
`<name>-credentials`, `<name>-stores`, `<name>-image-pull`, `<name>-libs`. Two
instances in one namespace can never share an object. To share one libs claim
on purpose, create it yourself and reference it with `libs.pvc.existing`.

### 9.3 docker

```yaml
docker:
  command: docker                # or "docker --context foo"
  name: solmq-connector
  project-name: solace-ibmmq-connectors   # the compose project
  restart: unless-stopped
  run-as: "1000:0"
  ports:
    - 8090                       # publish 8090:8090
    - "8081:8090"                # or host:container
  libs:
    dir: ./libs                  # bind-mounted at /app/external/libs
```

| Key | Meaning |
|---|---|
| `project-name` | the compose project, emitted as the file's `name:`; default `solace-ibmmq-connectors`; DNS-1123. `status` reads it back off the container and flags a mismatch |
| `restart` | compose's words: `no`, `always`, `on-failure`, `on-failure:<n>`, `unless-stopped` (default) |
| `ports` | bare port or `host:container`, each 1 to 65535. Omit it and nothing is published; the tool never needs one |
| `libs.dir` | the only `libs` key: a host directory bind-mounted read-only |

What `deploy` produces:

- `application.yml` and the status script are inlined under compose
  `configs:`, with every `$` doubled. Keep the doubling if you edit a generated
  file by hand.
- Credentials are compose `secrets:` read from the environment `deploy` hands
  compose. A spec written entirely in the `-env` form, with no created store,
  yields a file you can also bring up yourself with those variables exported.
  Requires Compose v2.23.1 or newer.
- Each selected existing store is bind-mounted at
  `/app/external/classpath/truststores/<store name>`. `deploy` and `validate`
  refuse a missing file or a directory there.
- A created store travels as a base64 compose secret and is decoded into memory
  at every start. The file then also carries `command:` and an init script. A
  hand-run `docker compose up` on such a file fails naming the unset
  `_GEN_TLS_<STORE>_STORE` variable.
- The service is labelled with a digest of every selected store, so `compose
  up` recreates the container exactly when a store changed.
- A healthcheck runs the status script's `--health` mode every 30s, which fills
  the `HEALTH` column of `status container`. Nothing restarts on failure.

Notes:

- `COMPOSE_PROJECT_NAME` in your shell overrides `project-name` on both
  `deploy` and `remove`. Unset it to make `env.yaml` the authority.
- Container names are host-unique regardless of project. Two stacks need
  different `name` values, not just different projects.
- A project-name mismatch reported by `status` means `remove` would leave that
  stack running. Set `project-name` to the reported value, or run `docker
  compose -p <name> down`.

### 9.4 podman

```yaml
podman:
  command: podman
  name: solmq-connector
  restart: always                # systemd's words; this is the default
  run-as: "1000:0"
  ports:
    - 8090
    - "8081:8090"
  libs:
    dir: ./libs
```

| Key | Meaning |
|---|---|
| `restart` | systemd's `Restart=` values: `no`, `always` (default), `on-success`, `on-failure`, `on-abnormal`, `on-abort`, `on-watchdog`. `unless-stopped` is an error |
| `ports`, `libs.dir` | as docker |

What `deploy` produces: one `.container` quadlet unit, and nothing else on
disk. Credentials, `application.yml`, the status script, the logback config and
every selected store are loaded into podman's secret store and mounted back by
`Secret=` lines:

| Secret | Mounted at |
|---|---|
| `<name>-<KEY>`, one per credential | `/app/external/var/secrets/<KEY>` |
| `<name>-application.yml` | `/app/external/spring/config/application.yml` |
| `<name>-tls-<store>`, one per selected store | `/app/external/classpath/truststores/<store>` |
| `<name>-status-script` | `/app/external/.status-script` |
| `<name>-logback-spring.xml` (syslog only) | `/app/external/classpath/logback-spring.xml` |

- The unit goes to `/etc/containers/systemd/` when run as root, otherwise to
  `~/.config/containers/systemd/`, and `systemctl` or `systemctl --user`
  manages it. This is not configurable; `podman.quadlet`, `podman.mode` and
  `podman.base-dir` are rejected or ignored.
- `deploy` resolves everything first, loads the secrets, deletes the ones the
  spec no longer mounts, writes the unit, reloads systemd, then starts a stopped
  unit or restarts a running one. `remove` stops the unit, removes it, then
  removes the secrets.
- Secrets hold 1 to 511999 bytes; an empty or larger store is refused.
- The unit carries a healthcheck (`HealthCmd=`, 30s) that fills the `HEALTH`
  column. Nothing restarts on failure.
- `generate` writes nothing, so its unit is a preview naming secrets only
  `deploy` creates.
- With the default `file` driver the secrets live under
  `~/.local/share/containers/storage/secrets/` (rootless) or
  `/var/lib/containers/storage/secrets/` (root). Requires podman 4.5 or newer.

### 9.5 Run-as and hardening

The connector image declares no `USER`. Every artifact states the identity and a
fixed set of restrictions:

| `run-as` | Means |
|---|---|
| unset | `1000:0`: an unprivileged uid and the group the image's files belong to |
| `"<uid>:<gid>"` | decimal ids, up to 2147483647 |
| `"<uid>"` | that uid with gid 0 |
| `image` | the image's own user (root); no ids written. Use on OpenShift, where the SCC assigns them |

Quote the value. `"0"` warns. The restrictions are always written: kubernetes
`runAsNonRoot`, `fsGroup`, `seccompProfile: RuntimeDefault`,
`allowPrivilegeEscalation: false`, all capabilities dropped; docker `privileged:
false`, `cap_drop: [ALL]`, `no-new-privileges`; podman `NoNewPrivileges=true`,
`DropCapability=all`. Under `image` the connector is root without root's
capabilities.

What the identity has to read:

- **kubernetes**: Secrets mount `0400` and `fsGroup` makes them readable. An
  NFS volume behind `libs` ignores `fsGroup`, so the export must be readable
  (and writable for `libs.download`) by the run-as uid or gid.
- **docker**: store files and `libs.dir` are bind-mounted with their host
  modes. `deploy` warns on Linux about each path the identity cannot read, with
  a `chmod`/`chgrp` fix. A keystore should be `chgrp` to the run-as gid rather
  than opened to everyone.
- **podman**: only `libs.dir` is a bind mount. Rootless, a non-zero uid maps to
  a subordinate id that owns no host file, so jars must be readable by others
  or by your group.

`download jar` writes jars `0644`, so a directory it filled needs nothing.

**Rootless podman** needs the run-as ids inside your `/etc/subuid` and
`/etc/subgid` ranges (65536 by default). `deploy` warns when they are not.
Widen the range (`usermod --add-subuids`, then `podman system migrate`), pick a
smaller id, deploy as root, or set `run-as: image`.

**OpenShift**: set `run-as: image`. The restrictions match the `restricted-v2`
SCC.

## 10. Secrets

Each credential becomes a file under `/app/external/var/secrets/`, and the
connector reads it through `spring.config.import: optional:configtree:`. No
credential value appears in `application.yml`, a compose file, a quadlet unit or
any file the tool leaves on disk. Two exceptions: the `solmq-status` password
([section 7.1](#71-the-reserved-status-account)) and the kubernetes manifest
`generate --platform kubernetes` prints, which embeds every Secret.

### 10.1 Declaring a credential

Every credential is a pair: a literal, or the name of a host environment
variable. Never both.

| Literal | Environment reference |
|---|---|
| `client-username` / `client-password` | `client-username-env` / `client-password-env` |
| `user` / `password` (mq) | `user-env` / `password-env` |
| `tls.stores.<name>.password` | `password-env` |
| `security.users[].password` | `password-env` |
| `image.pass` | `image.pass-env` |

An `-env` value is a bare name (`SOL_PASSWORD`), not `${SOL_PASSWORD}`. A name
starting with `_GEN_` is rejected. Prefer `-env` everywhere: the file then
stays safe to commit.

### 10.2 Mount names

| Declared as | Mounted under |
|---|---|
| `-env` | the variable name you wrote |
| a literal | a derived name prefixed `_GEN_`: `_GEN_<BINDER>_CLIENT_PASSWORD`, `_GEN_TLS_<STORE>_PASSWORD`, `_GEN_SECURITY_USER_<NAME>_PASSWORD`, ... |

`<BINDER>`, `<STORE>` and `<NAME>` are upper-cased with every run of
punctuation folded to `_`. Two positions naming the same variable are one
credential, mounted once. Two different credentials folding to one name (users
`ops.1` and `ops-1`) are an error naming both.

A spec written entirely with `-env` mounts its credentials under exactly the
names it wrote, so `application.yml` reads `${SOL_PASSWORD}` and a
`credentials.existing` Secret can be built by hand from those names.

### 10.3 How each platform delivers them

| Platform | Delivery | Requires |
|---|---|---|
| kubernetes | the `<name>-credentials` Secret (or your `existing` one) mounted as a volume, `0400`, no `envFrom`; `automountServiceAccountToken: false` | `secrets.credentials` once any credential is set |
| docker | compose `secrets:` from the environment provider, injected into the compose child process only | Compose v2.23.1+ |
| podman | podman secrets `<name>-<KEY>`, mounted by `Secret=` lines; removed by `remove` | podman 4.5+ |

For a kubernetes `existing` Secret, create it from your `-env` names:

```sh
kubectl create secret generic my-credentials \
  --from-literal=SOL_PASSWORD="$SOL_PASSWORD" \
  --from-literal=MQ_CORE_PASSWORD="$MQ_CORE_PASSWORD"
```

A literal credential still needs its derived key. Run `generate --platform
kubernetes` once with `create: true` and read the `stringData:` keys off the
manifest for the authoritative list. A missing key fails quietly: the pod
starts and the connector fails later to authenticate.

### 10.4 Registry credentials

Pulling the image is the engine's job, not the connector's.

| Platform | Mechanism |
|---|---|
| kubernetes | `secrets.image-pull.name` references a Secret you made (`kubectl create secret docker-registry`); `create: true` builds `<name>-image-pull` from the `image` block instead. Omitted, nothing is created |
| docker, podman | none generated. Run `docker login` / `podman login` on the host first |
