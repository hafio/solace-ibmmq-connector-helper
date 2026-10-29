# Getting started with solmq-conn-util

> [!WARNING]
> **Not a supported Solace product.** `solmq-conn-util` was created by Solace
> Professional Services and is supported only by Solace Professional Services --
> not by Solace Support. For help with this tool, contact your Solace
> Professional Services representative rather than opening a Solace Support
> case. This notice covers this tool only, not the Solace PubSub+ Connector for
> IBM MQ that it configures and deploys.

From nothing to a running connector in six steps. Each step links to the
reference when you need more.

## 1. Install

Download the release binary for your platform (`solmq-conn-util-linux-amd64`,
`solmq-conn-util-darwin-arm64`, `solmq-conn-util-windows-amd64.exe`), rename it
to `solmq-conn-util` and put it on your `PATH`. Nothing else to install. To
build from source, see [DEVELOPMENT.md](DEVELOPMENT.md).

```sh
solmq-conn-util --help          # usage
solmq-conn-util help deploy     # one verb's flags and examples
solmq-conn-util version
```

Optional: tab completion for your shell
([operations.md section 12](operations.md#12-shell-completion)).

What each verb needs:

| Verbs | Needs |
|---|---|
| `validate`, `generate`, `examples` | files only; no network, broker or cluster |
| `deploy`, `remove`, `start`, `stop`, `restart`, `status`, `logs`, `cli` | the platform CLI (`kubectl`/`oc`, `docker`, or `podman` + `systemctl`) logged in |
| `download jar` | HTTPS to Maven Central |

## 2. Write a spec

A spec is a folder: one `env.yaml` plus one small YAML file per workflow.
Start from the sample set:

```sh
solmq-conn-util examples examples        # write a ready-to-edit set into ./examples
```

| File | Contents |
|---|---|
| `env.yaml` | connections, TLS stores, logging, leader election, and `kubernetes:` / `docker:` / `podman:` sections |
| `workflow-0.yaml` | IBM MQ queue to Solace, inline MQ connection, `conn-ref` to Solace |
| `workflow-1.yaml` | Solace to IBM MQ, the reverse |
| `workflow-2.yaml` | IBM MQ to Solace, both sides by `conn-ref` |
| `workflow-3.yaml` | IBM MQ topic to Solace over plaintext, with an auto-named durable subscription |

Together they cover referenced and inline connections, reused and single-use,
over mTLS, TLS-only and plaintext. Existing files are left alone unless you
pass `-f`.

Or write the minimum by hand. `specs/env.yaml`:

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

`specs/workflow-0.yaml`, bridging an MQ queue to a Solace topic:

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

Credentials are written as `-env` names and read from your environment at
deploy, so the folder is safe to commit. Every key is explained in
[configuration.md](configuration.md).

**Prefer a form?** Open
[solmq-conn-util-generator.html](../solmq-conn-util-generator.html) in a
browser. No server, no install. It builds the whole folder, lints it with the
same rules as `validate`, previews the `application.yml`, and downloads the set
as a zip. The CLI stays authoritative where the two differ.

## 3. Validate and preview

```sh
solmq-conn-util validate -e specs/env.yaml           # every finding, in one run
solmq-conn-util generate config -e specs/env.yaml    # the application.yml it produces
```

`validate` reports every error and warning at once, including the TLS store
checks, and runs without production secrets exported. `generate` stops at the
first error and writes nothing.

## 4. Add a platform section

Add one of `kubernetes:`, `docker:` or `podman:` to `env.yaml`. The smallest
kubernetes section:

```yaml
kubernetes:
  command: kubectl                 # or "oc", or "kubectl --context prod"
  deployment:
    name: solmq-connector
    namespace: solace-connectors
    resources:
      cpu: "1"
      memory: 1Gi
  secrets:
    credentials:
      create: true
```

The smallest docker or podman section:

```yaml
docker:
  name: solmq-connector
```

Everything else has a default. The connector runs as uid 1000, non-root, with
every capability dropped. See
[configuration.md section 9](configuration.md#9-platform-sections).

If a workflow uses IBM MQ, fetch the MQ client jars first and point the
section's `libs` at them:

```sh
solmq-conn-util download jar mq ./libs -e specs/env.yaml
```

## 5. Deploy

```sh
export SOL_PASSWORD=... MQ_PASSWORD=...
solmq-conn-util generate --platform kubernetes -e specs/env.yaml   # inspect first (dry run)
solmq-conn-util deploy   --platform kubernetes -e specs/env.yaml
```

`deploy` renders the artifacts, runs a read-only preflight against the platform,
then applies them by piping to `kubectl`, `docker compose` or `systemctl`.
Credentials travel as mounted files, never as environment variables. With one
platform section in `env.yaml`, `--platform` can be omitted.

On kubernetes, a later `deploy` that changes only `application.yml` or a TLS
store does not restart the pods. Follow it with:

```sh
solmq-conn-util restart --platform kubernetes -e specs/env.yaml
```

## 6. Check, read logs, tear down

```sh
solmq-conn-util status all -e specs/env.yaml     # engine view + the connector's own view
solmq-conn-util logs -e specs/env.yaml           # one instance's log
solmq-conn-util cli -e specs/env.yaml            # a shell inside it
solmq-conn-util remove -e specs/env.yaml         # asks first; --no-prompt in scripts
```

`status container` says whether it is up and what image it runs. `status
application` says which instance is active, whether it is healthy, and which
workflows run. Both are in
[operations.md section 7](operations.md#7-status).

## Where next

- [configuration.md](configuration.md): every `env.yaml` and workflow key.
- [operations.md](operations.md): every verb, flag and exit code.
- [troubleshooting.md](troubleshooting.md): known pitfalls and upgrade notes.
- [commands.md](commands.md): the generated command reference, with an example
  per command.
