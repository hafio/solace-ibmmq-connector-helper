# solmq-conn-util -- Solace IBM MQ Connector config generator and deployer

`solmq-conn-util` turns a folder of small, per-workflow YAML files plus one `env.yaml`
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
- **`status` tells you which instance is active**, on any platform, by exec'ing
  into each running instance and reading its own actuator; `version` prints the
  build's own version plus its Go/OS/arch, for bug reports.

## Quick start

Grab the release binary for your platform (`solmq-conn-util-linux-amd64`,
`solmq-conn-util-darwin-arm64`, `solmq-conn-util-windows-amd64.exe`, ...) or build from
source (see [docs/DEVELOPMENT.md](docs/DEVELOPMENT.md)), then:

```sh
solmq-conn-util examples specs                     # write a ready-to-edit sample set into ./specs
solmq-conn-util generate config -e specs/env.yaml  # print the application.yml those samples produce
```

The sample set is four cross-platform (MQ/Solace) workflows that together cover
every connection style -- referenced vs. inline, reused vs. single-use -- across
mTLS, TLS-only, and plaintext transports, plus an MQ topic source exercising the
auto-named durable subscription. Full breakdown: [userguide.md](userguide.md) section 11.

Prefer a form to a text editor? Open
[solmq-conn-util-generator.html](solmq-conn-util-generator.html) in a browser (no server, no
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
solmq-conn-util generate config -e specs/env.yaml                     # application.yml on stdout
solmq-conn-util generate config -e specs/env.yaml -o application.yml  # ...or written to a file
```

Add a `kubernetes:` section ([userguide.md](userguide.md) section 7) and
`solmq-conn-util generate --platform kubernetes -e specs/env.yaml` emits the full
manifest set (Namespace, ConfigMap, Deployment, Service, Secrets).
`solmq-conn-util deploy --platform kubernetes -e specs/env.yaml` then applies it by
piping the manifest to `kubectl`/`oc`.

## Commands

> **Breaking change:** the platform used to be a second positional argument
> (`deploy kubernetes`); it is now `--platform kubernetes`, or just `deploy` when
> `env.yaml` has exactly one platform section. See [userguide.md](userguide.md)
> for the full resolution order (flag, single section, interactive menu, or a
> loud error) and why CI must pass `--platform` explicitly.

```text
solmq-conn-util generate [config] [--platform kubernetes|docker|podman] [-e env.yaml] [-o out]
                                                                   Emit application.yml, or the resolved platform's artifacts
solmq-conn-util deploy  [--platform kubernetes|docker|podman] [-e env.yaml]  Generate for the resolved platform, then apply it
solmq-conn-util delete  [--platform kubernetes|docker|podman] [-e env.yaml]  Tear the same platform down
solmq-conn-util status  [--install] [--platform kubernetes|docker|podman] [-e env.yaml]
                                                                   Ensure and run the status script; report per-target leader-election + workflow state
solmq-conn-util version                                           Print the utility name, version, Go version and OS/arch
solmq-conn-util validate            [-e env.yaml]                 Lint the whole env.yaml + workflows
solmq-conn-util examples [dir] [-f]                               Write a starter env.yaml + workflows
solmq-conn-util completion bash|zsh|fish|powershell               Print a shell completion script

-e, --env     Config file, relative or absolute path (default: env.yaml)
-o, --out     generate output file (default: stdout)
-f, --force   examples: overwrite existing files
```

`generate` fails fast (stops at the first error, writes nothing); `validate`
reports every finding; generate output is buffered and only written on full
success -- never a half-written `-o`. `deploy`/`delete`/`status` run the CLI
named by each section's `command:` through an argv slice -- never a shell --
with every token checked against a safe charset, argv[0] checked against a
per-platform binary allowlist (escape hatch: `--allow-command`), and a
read-only login/daemon preflight before anything is written, applied, or
queried. Details and exit codes: [userguide.md](userguide.md) section 3; the
full generated command reference: [docs/commands.md](docs/commands.md).

Tab completion for all of the above: `solmq-conn-util completion bash|zsh|fish|powershell`
prints a script for your shell, rendered from the binary's own command model so it
never drifts from the commands that binary accepts
([userguide.md](userguide.md) section 1.1).

## Documentation

- [userguide.md](userguide.md) -- the complete user reference: commands (section
  3), the config file and workflow discovery (section 4), workflow files (section
  5), the `env.yaml` connector defaults (section 6), the deploy targets --
  kubernetes/docker/podman (section 7), the secrets model (section 8), what gets
  generated (section 9), determining which instance is active (section 10), the
  sample set (section 11), and gotchas (section 12).
- [docs/commands.md](docs/commands.md) -- the full command tree / reference,
  generated from the command model and gated against drift.
- [docs/DEVELOPMENT.md](docs/DEVELOPMENT.md) -- building, dev-script tasks, tests
  and golden fixtures, CI release, design notes.
- [solmq-conn-util-generator.html](solmq-conn-util-generator.html) -- a standalone browser
  page (no install, no server) that generates the `env.yaml` + workflow files,
  reports the same findings as `validate`, and previews the `application.yml`.
