# solmq-gen — Solace IBM MQ Connector config generator

`solmq-gen` turns a folder of small, per-workflow YAML files into one consolidated
`application.yml` for the **Solace PubSub+ Connector for IBM MQ**
(`solace/solace-pubsub-connector-ibmmq:2.13.0`), and can emit the Kubernetes
manifests that run it.

- **Reusable connections**: define `connections.<name>` once in `defaults.yaml`,
  reference it with `conn-ref`; identical connections dedup into shared **binders**.
- Auto-numbers workflows by sorted filename, derives destination-types from
  `queue:`/`topic:`, and auto-names a **durable subscription** for every MQ topic source.
- Implements **leader-election** (`standalone` / `active_active` / `active_standby`).
- Wires **TLS + mTLS** for both Solace and MQ from one shared truststore/keystore.
- Secrets stay `${VAR}` placeholders — never inlined into config.

## Quick start

Grab the release binary for your platform (or build from source — see
[docs/DEVELOPMENT.md](docs/DEVELOPMENT.md)), then:

```sh
solmq-gen examples examples    # write a ready-to-edit sample set into ./examples
solmq-gen config examples      # print the application.yml those samples produce
```

The sample set is four cross-platform (MQ↔Solace) workflows that together cover
every connection style — referenced vs. inline, reused vs. single-use — across
mTLS, TLS-only, and plaintext transports, plus an MQ topic source exercising the
auto-named durable subscription. Full breakdown: [userguide.md](userguide.md) §10.

## Minimal working example

Two files in a folder (`specs/` here) are a complete spec.
`specs/defaults.yaml` defines a reusable connection:

```yaml
connections:
  prod-solace:
    solace:
      host: tcp://broker.internal:55555
      msg-vpn: prod
      client-username: connector
      client-password: ${SOL_PASSWORD}
```

`specs/workflow-0.yaml` bridges an MQ queue to a Solace topic, referencing it:

```yaml
source:
  mq:
    conn-name: mqhost.internal(1414)
    queue-manager: QM1
    channel: DEV.APP.SVRCONN
    user: appuser
    password: ${MQ_PASSWORD}
    queue: ORDERS.OUT
target:
  solace:
    conn-ref: prod-solace
    topic: orders/from-mq
```

```sh
solmq-gen config specs                     # application.yml on stdout
solmq-gen config specs -o application.yml  # ...or written to a file
```

Add a `kubernetes.yaml` ([userguide.md](userguide.md) §7) and `solmq-gen deploy specs`
emits the full manifest set (Namespace, ConfigMap, Deployment, Service, Secrets).

## Commands

```text
solmq-gen config   <dir> [-o out]            Emit application.yml (fails fast)
solmq-gen deploy   <dir> [-k kube] [-o out]  Emit ConfigMap+Deployment+Service (+Secrets)
solmq-gen validate <dir> [-k kube]           Lint only; report every error
solmq-gen examples [dir] [-f]                Write sample spec files (default dir: examples)

-o, --out    Output file (default: stdout)
-k, --kube   Kubernetes settings file (default: kubernetes.yaml)
-f, --force  Overwrite existing files (examples)
```

`config`/`deploy` fail fast (stop at the first error, write nothing); `validate`
reports every finding; output is buffered and only written on full success —
never a half-written `-o`. Details and exit codes: [userguide.md](userguide.md) §3.

## Documentation

- [userguide.md](userguide.md) — the complete user reference: commands (§3), the
  spec folder (§4), workflow files (§5), `defaults.yaml` (§6), `kubernetes.yaml`
  (§7), the secrets model (§8), what gets generated (§9), the sample set (§10),
  and gotchas (§11).
- [docs/DEVELOPMENT.md](docs/DEVELOPMENT.md) — building, dev-script tasks, tests
  and golden fixtures, CI release, design notes.
