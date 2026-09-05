# solmq-conn-util -- Solace IBM MQ Connector config generator and deployer

`solmq-conn-util` turns a folder of small, per-workflow YAML files plus one `env.yaml`
into a consolidated `application.yml` for the **Solace PubSub+ Connector for IBM
MQ** (`solace/solace-pubsub-connector-ibmmq:2.13.0`), generates the Kubernetes,
Docker Compose, or Podman artifacts that run it, and can apply or tear those down
by shelling out to `kubectl`/`oc`, `docker`, or `podman`/`systemctl`. One folder is
one connector instance, holding up to 20 workflows; a folder with more is rejected,
so splitting them across connectors stays your decision.

- **One config file.** `env.yaml` holds the connector defaults, workflow
  discovery, and a per-platform (`kubernetes:` / `docker:` / `podman:`) deploy
  section ([section 8](docs/userguide.md#8-platform-sections-kubernetes-docker-podman)).
- **Reusable connections**: define `connections.<name>` once in `env.yaml`,
  reference it with `conn-ref`; identical connections dedup into shared
  **[binders](docs/userguide.md#66-reusable-connections-conn-ref)**.
- Auto-numbers workflows by sorted filename, derives destination-types from
  `queue:`/`topic:`, and auto-names a
  **[durable subscription](docs/userguide.md#64-destinations-durable-names-passthrough)**
  for every MQ topic source.
- Implements **[leader-election](docs/userguide.md#7-connector-defaults-envyaml-top-level)**
  (`standalone` / `active_active` / `active_standby`).
- Wires **[TLS + mTLS](docs/userguide.md#7-connector-defaults-envyaml-top-level)**
  for both Solace and MQ from one shared truststore/keystore.
- **[One secrets model](docs/userguide.md#9-secrets-model) everywhere**: each
  credential is declared as a literal or an `-env` variable name and mounted
  as a file under `/app/external/var/secrets/` -- a Kubernetes Secret volume,
  a compose environment-provider secret, or a podman secret, never an
  environment variable and never written to disk as a value. The one
  exception is the tool's own reserved `solmq-status` account, whose password
  is rendered as a literal by design
  ([section 7.1](docs/userguide.md#71-the-reserved-status-account-solmq-status)).
- **[`status`](docs/userguide.md#12-status-the-container-the-connector-or-both)
  reports each instance from either side**: `status container` reads the
  engine from outside (state, restarts, age, the image actually running),
  `status application` execs into each instance and reads its own actuator
  (which one is active, health, workflows), and `status all` reports both.
  `-d` adds node, CPU/memory, digest and referenced objects; `--all` finds
  every connector instance by image name; `--output json` emits the same
  facts as one document.
- **[`download jar`](docs/userguide.md#10-download-jar) fetches the IBM MQ client
  jars (or the syslog encoder jar)** from Maven Central over HTTPS,
  sha1-verified, into a local directory for the `libs.dir`/`libs.pvc`/
  `libs.download` deploy options. It is image-aware: a jar the connector
  image already ships at an equal-or-newer version is skipped and reported
  rather than re-fetched, against a snapshot of one specific image --
  deploying to a different one needs its own list.

Every term above links straight through to where
[userguide.md](docs/userguide.md) explains it in full.

## Quick start

Grab the release binary for your platform (`solmq-conn-util-linux-amd64`,
`solmq-conn-util-darwin-arm64`, `solmq-conn-util-windows-amd64.exe`, ...) -- no
install, nothing else to run -- or build from source (Go 1.27+; see
[docs/DEVELOPMENT.md](docs/DEVELOPMENT.md)), then:

```sh
solmq-conn-util examples examples                     # write a ready-to-edit sample set into ./examples
solmq-conn-util generate config -e examples/env.yaml  # print the application.yml those samples produce
```

The sample set is four cross-platform (MQ/Solace) workflows that together cover
every connection style -- referenced vs. inline, reused vs. single-use -- across
mTLS, TLS-only, and plaintext transports, plus an MQ topic source exercising the
auto-named durable subscription. Full breakdown:
[section 4](docs/userguide.md#4-examples).

Prefer a form to a text editor? Open
[solmq-conn-util-generator.html](solmq-conn-util-generator.html) in a browser (no server, no
install): it builds the whole spec folder, lints it with the same rules `validate`
enforces, previews the `application.yml` it consolidates into, and downloads the
set as a zip.

## Minimal working example

One `env.yaml` plus one workflow file in a folder (`specs/` here -- any name
works; this is a hand-written pair, not the four-workflow set `examples`
writes into `./examples`, [section 4](docs/userguide.md#4-examples))
are a complete spec. `specs/env.yaml` sets the workflow discovery and a
reusable connection:

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

Add a `kubernetes:` section
([section 8](docs/userguide.md#8-platform-sections-kubernetes-docker-podman)) and
`solmq-conn-util generate --platform kubernetes -e
specs/env.yaml` emits the full manifest set (Namespace, ConfigMap, Deployment,
Service, Secrets). `solmq-conn-util deploy --platform kubernetes -e
specs/env.yaml` then applies it by piping the manifest to `kubectl`/`oc`.

## Commands

> [!IMPORTANT]
> The platform comes from the `--platform` flag, from `env.yaml` when it has
> exactly one platform section, or from an interactive menu otherwise -- see
> [userguide.md section 3](docs/userguide.md#3-commands). CI and scripts must pass
> `--platform` explicitly: the menu refuses to block when stdin is not a TTY.

| Verb | Aliases | What it does |
| --- | --- | --- |
| `generate` | `gen` | Render application.yml, or the deploy artifacts for the resolved platform |
| `deploy` | `dp` | Generate for a platform, then apply it |
| `remove` | `rm` | Tear down what deploy created for a platform |
| `status` | `sts` | Report each instance: container (engine), application (connector), or all |
| `logs` | `lg` | Print one instance log, where status says what but not why |
| `cli` | _(none)_ | Open a shell inside one instance, or run one command in it |
| `version` | `ver` | Print the utility name, version, Go version and OS/arch |
| `validate` | `vld` | Lint the whole env.yaml + workflows |
| `examples` | `eg` | Write a starter env.yaml + workflows |
| `download` | `dl` | Download IBM MQ or syslog encoder jars and their dependencies |
| `auto-complete` | _(none)_ | Print a shell completion script |
| `help` | _(none)_ | Print the CLI usage summary, or the help page of one verb |

Full flag reference and synopses: [docs/commands.md](docs/commands.md); every
alias, target word, platform short form, and flag abbreviation:
[docs/abbreviation.md](docs/abbreviation.md).

`generate` fails fast: it stops at the first error and writes nothing, and its
output is buffered so a failed run never leaves a half-written `-o` file;
`validate` instead reports every finding it can.

`deploy`, `remove`, `status`, `logs`, and `cli` all run the platform CLI named
by each section's `command:` through an argv slice -- never a shell. Every
token is checked against a safe charset, `argv[0]` is checked against a
per-platform binary allowlist (escape hatch: `--allow-command`), and a
read-only login/daemon preflight must succeed before anything is written,
applied, or queried.

`remove` additionally confirms first, naming what it will tear down;
`--no-prompt` skips that prompt and is required for a non-TTY run. On
kubernetes the namespace is a separate, checked step: a namespace holding
anything the release does not own is listed and kept, and only an empty one is
offered for removal.

`cli` is the one verb whose exit code can leave the 0/1/2 contract: once a
session or command is running inside the instance, its own status is passed
straight back. The engines give no way to be more precise -- `kubectl exec`
reports an unreachable pod and a command that exited non-zero with the same
status -- so a non-zero `cli` exit means one of the two, and the message the
engine printed on stderr says which.

Every kubernetes `exec` and `logs` this tool runs names its container
explicitly (`-c connector`), so a pod carrying a sidecar cannot be entered or
read by mistake; docker and podman default to the same name. Details and exit
codes: [section 3](docs/userguide.md#3-commands); the full generated command
reference: [docs/commands.md](docs/commands.md).

Tab completion for all of the above: `solmq-conn-util auto-complete
bash|zsh|fish|powershell` prints a script for your shell, rendered from the
binary's own command model so it never drifts from the commands that binary
accepts ([section 1.1](docs/userguide.md#11-shell-completion)).

## Documentation

- [docs/userguide.md](docs/userguide.md) -- the complete user reference: commands
  ([section 3](docs/userguide.md#3-commands)), the sample set
  ([section 4](docs/userguide.md#4-examples)), the config file and workflow
  discovery ([section 5](docs/userguide.md#5-the-config-file-and-workflow-discovery)),
  workflow files ([section 6](docs/userguide.md#6-workflow-file)), the `env.yaml`
  connector defaults ([section 7](docs/userguide.md#7-connector-defaults-envyaml-top-level)),
  the platform sections -- kubernetes/docker/podman
  ([section 8](docs/userguide.md#8-platform-sections-kubernetes-docker-podman)),
  the secrets model ([section 9](docs/userguide.md#9-secrets-model)),
  `download jar` ([section 10](docs/userguide.md#10-download-jar)), what gets
  generated ([section 11](docs/userguide.md#11-what-gets-generated)), status
  ([section 12](docs/userguide.md#12-status-the-container-the-connector-or-both)),
  reading instance logs
  ([section 13](docs/userguide.md#13-logs-the-lines-behind-the-state)), opening a
  shell inside an instance
  ([section 14](docs/userguide.md#14-cli-a-shell-inside-the-instance)), and
  gotchas ([section 15](docs/userguide.md#15-notes-and-gotchas)).
- [docs/commands.md](docs/commands.md) -- the full command tree / reference,
  generated from the command model and gated against drift.
- [docs/abbreviation.md](docs/abbreviation.md) -- every short spelling the CLI
  accepts (commands, status targets, platforms, flags), keyed by the
  abbreviation; generated from the same model and gated the same way.
- [docs/DEVELOPMENT.md](docs/DEVELOPMENT.md) -- building, dev-script tasks, tests
  and golden fixtures, CI release, design notes.
- [docs/test.md](docs/test.md) -- the test catalogue: every test grouped by
  package, so you can see what behavior is covered and jump to the test that
  covers it.
- [solmq-conn-util-generator.html](solmq-conn-util-generator.html) -- the
  browser-based spec generator described in [Quick start](#quick-start) above.
