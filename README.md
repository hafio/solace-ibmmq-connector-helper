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
- **`status` reports each instance from either side**: `status container` reads
  the engine from outside (state, restarts, age, the image actually running),
  `status application` execs into each instance and reads its own actuator (which
  one is active, health, workflows), and `status all` reports both. `-d` adds
  node, CPU/memory, digest and referenced objects; `--all` finds every connector
  instance by image name; `--output json` emits the same facts as one document.
  `version` prints the build's own version plus its Go/OS/arch, for bug reports.
- **`download jar` fetches the IBM MQ client jars (or the syslog encoder jar)**
  from Maven Central over HTTPS into a local directory, so the `libs.dir` bind
  mount and the `libs.pvc`/`libs.download` deploy sections have something to
  point at. Each jar's sha1 is checked against Maven Central's own published
  digest before it lands on disk. It is **image-aware**: a jar the connector
  image already ships at an equal-or-newer version is skipped and reported
  rather than re-fetched -- but the built-in image jar list is a snapshot of
  one specific image, so deploying to a different one needs its own list (see
  [userguide.md](userguide.md) section 12).

Every term above -- binders, durable subscriptions, leader-election, the secrets
model -- is explained in depth in [userguide.md](userguide.md).

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
      client-password-env: SOL_PASSWORD
```

`specs/workflow-0.yaml` bridges an MQ queue to a Solace topic, referencing it:

```yaml
source:
  mq:
    conn-name: mqhost.internal(1414)
    queue-manager: QM1
    channel: DEV.APP.SVRCONN
    user: appuser
    password-env: MQ_PASSWORD
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
>
> **Breaking change:** the teardown verb `delete` (alias `del`) is now `remove`
> (alias `rm`). The old spellings are not accepted -- update any script that ran
> `solmq-conn-util delete ...`. Only the verb changed: a kubernetes teardown
> still issues `kubectl delete -f -`.

```text
solmq-conn-util generate [config] [--platform kubernetes|docker|podman] [-e env.yaml] [-o out]
                                                                   Emit application.yml, or the resolved platform's artifacts
solmq-conn-util deploy  [--platform kubernetes|docker|podman] [-e env.yaml]  Generate for the resolved platform, then apply it
solmq-conn-util remove  [--platform kubernetes|docker|podman] [-e env.yaml]  Tear the same platform down
solmq-conn-util status  <container|application|all> [-d] [-w] [--all] [--output table|json]
                        [--install] [--platform kubernetes|docker|podman] [-e env.yaml]
                                                                   Report each instance: the engine's view, the connector's own, or both
solmq-conn-util logs    [--follow] [--previous] [--tail N] [--since d] [--timestamps] [--all]
                        [--platform kubernetes|docker|podman] [-e env.yaml]
                                                                   Print each instance log -- what status says happened, and why
solmq-conn-util version                                           Print the utility name, version, Go version and OS/arch
solmq-conn-util validate            [-e env.yaml]                 Lint the whole env.yaml + workflows
solmq-conn-util examples [dir] [-f]                               Write a starter env.yaml + workflows
solmq-conn-util auto-complete bash|zsh|fish|powershell            Print a shell completion script
solmq-conn-util download jar mq|syslog [dir] [-e env.yaml] [--version v] [--omit-lib-file file]
                                       [--include-provided] [--url u] [-f]
                                                                   Fetch IBM MQ or syslog jars from Maven Central into a local directory

# The in-binary help is shorter than this table: `solmq-conn-util -h` lists the
# commands, and `solmq-conn-util help <command>` (or `<command> -h`) prints that
# command's arguments, flags, and examples.

-e, --env     Config file, relative or absolute path (default: env.yaml)
-o, --out     generate output file (default: stdout)
-f, --force   examples/download jar: overwrite existing files
--version     download jar: pin the seed release instead of resolving latest stable
--omit-lib-file  download jar: jar list that REPLACES the embedded default (default: embedded list)
--include-provided  download jar: download the whole closure even where the image already provides it
--url         download jar: repeatable; exact URLs to fetch, skipping Maven resolution and image-aware omission
--follow      logs: keep the log open and print new lines until interrupted (one instance)
--previous    logs: read the previous container's log -- why a restarting pod died (kubernetes only)
--tail        logs: read only the last N lines, or all (default: all)
--since       logs: read only lines newer than this duration, e.g. 10m
--timestamps  logs: prefix every line with the time the platform recorded
```

Every verb above except `auto-complete` and `help` also has a short alias
(`gen`, `dp`, `rm`, `sts`, `lg`, `ver`, `vld`, `eg`, `dl`), and `--platform` accepts
`kube`, `dk` and `pm` -- see [userguide.md](userguide.md) section 3 for both
tables.

`generate` fails fast (stops at the first error, writes nothing); `validate`
reports every finding; generate output is buffered and only written on full
success -- never a half-written `-o`. `deploy`/`remove`/`status`/`logs` run the CLI
named by each section's `command:` through an argv slice -- never a shell --
with every token checked against a safe charset, argv[0] checked against a
per-platform binary allowlist (escape hatch: `--allow-command`), and a
read-only login/daemon preflight before anything is written, applied, or
queried. Details and exit codes: [userguide.md](userguide.md) section 3; the
full generated command reference: [docs/commands.md](docs/commands.md).

Tab completion for all of the above: `solmq-conn-util auto-complete bash|zsh|fish|powershell`
prints a script for your shell, rendered from the binary's own command model so it
never drifts from the commands that binary accepts
([userguide.md](userguide.md) section 1.1).

## Documentation

- [userguide.md](userguide.md) -- the complete user reference: commands (section
  3), the config file and workflow discovery (section 4), workflow files (section
  5), the `env.yaml` connector defaults (section 6), the deploy targets --
  kubernetes/docker/podman (section 7), the secrets model (section 8), what gets
  generated (section 9), determining which instance is active (section 10), the
  sample set (section 11), fetching the IBM MQ and syslog jars (section 12),
  reading instance logs (section 13), and gotchas (section 14).
- [docs/commands.md](docs/commands.md) -- the full command tree / reference,
  generated from the command model and gated against drift.
- [docs/abbreviation.md](docs/abbreviation.md) -- every short spelling the CLI
  accepts (commands, status targets, platforms, flags), keyed by the
  abbreviation; generated from the same model and gated the same way.
- [docs/DEVELOPMENT.md](docs/DEVELOPMENT.md) -- building, dev-script tasks, tests
  and golden fixtures, CI release, design notes.
- [solmq-conn-util-generator.html](solmq-conn-util-generator.html) -- a standalone browser
  page (no install, no server) that generates the `env.yaml` + workflow files,
  reports the same findings as `validate`, and previews the `application.yml`.
- [doc.md](doc.md) -- a configuration reference for the underlying connector's own
  `application.yml` (the raw `solace.connector.*` and Spring keys), for hand-tuning
  or debugging generated output beyond what `env.yaml` exposes. Parts are still
  marked unverified -- check a key against the connector's official docs before
  relying on it.
