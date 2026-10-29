# solmq-conn-util -- Solace IBM MQ Connector config generator and deployer

> [!WARNING]
> **Not a supported Solace product.** `solmq-conn-util` was created by Solace
> Professional Services and is supported only by Solace Professional Services --
> not by Solace Support. For help with this tool, contact your Solace
> Professional Services representative rather than opening a Solace Support
> case. This notice covers this tool only, not the Solace PubSub+ Connector for
> IBM MQ that it configures and deploys.

`solmq-conn-util` turns a folder of small per-workflow YAML files plus one
`env.yaml` into the `application.yml` for the **Solace PubSub+ Connector for
IBM MQ** (2.x and 3.x), generates the Kubernetes, Docker Compose or Podman
artifacts that run it, and deploys, inspects and tears those down by driving
`kubectl`/`oc`, `docker` or `podman`.

- One `env.yaml` holds the connector defaults, reusable connections, named TLS
  stores and a section per platform.
- Credentials are declared as environment variable names and reach the
  connector as mounted files, never as environment variables or literals.
- TLS and mTLS are wired per connection from PKCS12 stores you provide, stores
  `deploy` builds from your certificate files, or Secrets you already manage.
- Every artifact runs the connector non-root with every capability dropped.
- `status`, `logs` and `cli` report, read and enter the running instances on
  any of the three platforms.

## Quick start

Grab the release binary for your platform, rename it `solmq-conn-util`, then:

```sh
solmq-conn-util examples examples                     # write a sample spec into ./examples
solmq-conn-util validate -e examples/env.yaml         # lint it
solmq-conn-util generate config -e examples/env.yaml  # print the application.yml it produces
solmq-conn-util deploy --platform docker -e examples/env.yaml
```

Prefer a form to an editor? Open
[solmq-conn-util-generator.html](solmq-conn-util-generator.html) in a browser.
It builds the spec folder, lints it and downloads it as a zip.

The full walkthrough is [docs/getting-started.md](docs/getting-started.md).

## Commands

| Verb | Alias | What it does |
| --- | --- | --- |
| `generate [config]` | `gen` | Render `application.yml`, or the deploy artifacts for a platform |
| `deploy` | `dp` | Generate for a platform, then apply it |
| `remove` | `rm` | Tear down what `deploy` created |
| `start` / `stop` / `restart` | | Act on the instance in place |
| `status container\|application\|all` | `sts` | Report each instance: engine view, connector view, or both |
| `logs` | `lg` | Print one instance's log |
| `cli` | | Open a shell inside one instance, or run one command in it |
| `validate` | `vld` | Lint `env.yaml` and the workflows |
| `examples` | `eg` | Write a starter spec |
| `download jar mq\|syslog` | `dl` | Fetch IBM MQ or syslog encoder jars |
| `auto-complete` | | Print a shell completion script |
| `version`, `help` | `ver` | |

The platform comes from `--platform`, or from `env.yaml` when it has exactly one
platform section. Scripts and CI must pass `--platform`: every prompt refuses to
block when stdin is not a terminal.

## Documentation

- [docs/getting-started.md](docs/getting-started.md): install, write a spec,
  validate, deploy, check, tear down.
- [docs/configuration.md](docs/configuration.md): every `env.yaml` and
  workflow key.
- [docs/operations.md](docs/operations.md): every verb, flag, exit code and
  what gets generated.
- [docs/troubleshooting.md](docs/troubleshooting.md): known pitfalls and
  upgrade notes.
- [docs/commands.md](docs/commands.md) and
  [docs/abbreviation.md](docs/abbreviation.md): the generated command reference
  and every short spelling.
- [docs/DESIGN.md](docs/DESIGN.md): why the tool behaves the way it does.
- [docs/DEVELOPMENT.md](docs/DEVELOPMENT.md): building, testing and releasing.
