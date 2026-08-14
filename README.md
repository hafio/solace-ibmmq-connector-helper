# solmq-conn -- Solace IBM MQ Connector config generator and deployer

`solmq-conn` turns a folder of small, per-workflow YAML files plus one `env.yaml`
into a consolidated `application.yml` for the **Solace PubSub+ Connector for IBM
MQ** (`solace/solace-pubsub-connector-ibmmq:2.13.0`), generates the Kubernetes,
Docker Compose, or Podman artifacts that run it, and can apply or tear those down
by shelling out to `kubectl`/`oc`, `docker`, or `podman`/`systemctl`. One folder is
one connector instance, holding up to 20 workflows; a folder with more is rejected,
so splitting them across connectors stays your decision.

- **One config file.** `env.yaml` holds the connector defaults, workflow
  discovery, and a per-target (`kubernetes:` / `docker:` / `podman:`) deploy section.
- **Reusable connections**: define `connections.<name>` once in `env.yaml`,
  reference it with `conn-ref`; identical connections dedup into shared **binders**.
- Auto-numbers workflows by sorted filename, derives destination-types from
  `queue:`/`topic:`, and auto-names a **durable subscription** for every MQ topic
  source.
- Implements **leader-election** (`standalone` / `active_active` / `active_standby`).
- Wires **TLS + mTLS** for both Solace and MQ from one shared truststore/keystore.
- **One secrets model everywhere**: each credential is declared as a literal or an
  `-env` variable name, rendered into config only as a derived stable name, and
  mounted as a file under `/run/secrets/` -- a Kubernetes Secret volume, a compose
  environment-provider secret, or a podman secret. No credential value or host
  variable name reaches any generated file, and nothing secret is written to disk.

## Quick start

Grab the release binary for your platform (`solmq-conn-linux-amd64`,
`solmq-conn-darwin-arm64`, `solmq-conn-windows-amd64.exe`, ...) or build from
source (see [docs/DEVELOPMENT.md](docs/DEVELOPMENT.md)), then:

```sh
solmq-conn examples specs                     # write a ready-to-edit sample set into ./specs
solmq-conn generate config -e specs/env.yaml  # print the application.yml those samples produce
```

The sample set is four cross-platform (MQ/Solace) workflows that together cover
every connection style -- referenced vs. inline, reused vs. single-use -- across
mTLS, TLS-only, and plaintext transports, plus an MQ topic source exercising the
auto-named durable subscription. Full breakdown: [userguide.md](userguide.md) section 10.

Prefer a form to a text editor? Open
[solmq-conn-generator.html](solmq-conn-generator.html) in a browser (no server, no
install): it builds the whole spec folder, lints it with the same rules `validate`
enforces, previews the `application.yml` it consolidates into, and downloads the
set as a zip.

## Minimal working example

One `env.yaml` plus one workflow file in a folder (`specs/` here) are a complete
spec. `specs/env.yaml` sets the workflow discovery and a reusable connection:

```yaml
workflows:
  dir: .
  file_pattern: "*"
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
solmq-conn generate config -e specs/env.yaml                     # application.yml on stdout
solmq-conn generate config -e specs/env.yaml -o application.yml  # ...or written to a file
```

Add a `kubernetes:` section ([userguide.md](userguide.md) section 7) and
`solmq-conn generate kubernetes -e specs/env.yaml` emits the full manifest set
(Namespace, ConfigMap, Deployment, Service, Secrets).
`solmq-conn deploy kubernetes -e specs/env.yaml` then applies it by piping the
manifest to `kubectl`/`oc`.

## Commands

```text
solmq-conn generate config     [-e env.yaml] [-o out]        Emit application.yml
solmq-conn generate kubernetes [-e env.yaml] [-o out]        Emit ConfigMap+Deployment+Service (+Secrets)
solmq-conn generate docker     [-e env.yaml] [-o out]        Emit docker-compose.yml (application.yml inlined)
solmq-conn generate podman     [-e env.yaml] [-o out]        Emit a podman run script or quadlet unit
solmq-conn deploy  kubernetes|docker|podman  [-e env.yaml]   Apply: kubectl/oc, docker compose up, or systemctl start
solmq-conn delete  kubernetes|docker|podman  [-e env.yaml]   Tear the same target down
solmq-conn validate            [-e env.yaml]                 Lint the whole env.yaml + workflows
solmq-conn examples [dir] [-f]                               Write a starter env.yaml + workflows

-e, --env     Config file, relative or absolute path (default: env.yaml)
-o, --out     generate output file (default: stdout)
-f, --force   examples: overwrite existing files
```

`generate` fails fast (stops at the first error, writes nothing); `validate`
reports every finding; generate output is buffered and only written on full
success -- never a half-written `-o`. `deploy`/`delete` run the CLI named by each
section's `command:` through an argv slice -- never a shell -- with every token
checked against a safe charset, argv[0] checked against a per-platform binary
allowlist (escape hatch: `--allow-command`), and a read-only login/daemon
preflight before anything is written or applied. Details and exit codes:
[userguide.md](userguide.md) section 3; the full generated command reference:
[docs/commands.md](docs/commands.md).

## Documentation

- [userguide.md](userguide.md) -- the complete user reference: commands (section
  3), the config file and workflow discovery (section 4), workflow files (section
  5), the `env.yaml` connector defaults (section 6), the deploy targets --
  kubernetes/docker/podman (section 7), the secrets model (section 8), what gets
  generated (section 9), the sample set (section 10), and gotchas (section 11).
- [docs/commands.md](docs/commands.md) -- the full command tree / reference,
  generated from the command model and gated against drift.
- [docs/DEVELOPMENT.md](docs/DEVELOPMENT.md) -- building, dev-script tasks, tests
  and golden fixtures, CI release, design notes.
- [solmq-conn-generator.html](solmq-conn-generator.html) -- a standalone browser
  page (no install, no server) that generates the `env.yaml` + workflow files,
  reports the same findings as `validate`, and previews the `application.yml`.
