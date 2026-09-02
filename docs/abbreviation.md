# solmq-conn-util abbreviations

<!-- GENERATED -- do not edit by hand.
Source of truth: cmd/solmq-conn-util/commands.go (the cliSpec model) and
cmd/solmq-conn-util/main.go (platformAliasList).
Regenerate: go test ./cmd/solmq-conn-util -run TestAbbreviationDocInSync -update
TestAbbreviationDocInSync fails the build if this file drifts from the model. -->

Every short spelling `solmq-conn-util` accepts, keyed by the abbreviation.
Each one is recognised wherever its canonical word is -- both when the command
runs and in shell completion -- but only the canonical word is ever printed by
terminal help or offered by the TAB menu, so the short forms are documented here
and in [commands.md](commands.md) rather than in the binary. Generated from the
command model in
[`cmd/solmq-conn-util/commands.go`](../cmd/solmq-conn-util/commands.go); see
[DEVELOPMENT.md](DEVELOPMENT.md#testing) to regenerate.

## Command abbreviations

| Short | Stands for | What it does |
|-------|------------|--------------|
| `dl` | `download` | Download IBM MQ or syslog encoder jars and their dependencies |
| `dp` | `deploy` | Generate for a platform, then apply it |
| `eg` | `examples` | Write a starter env.yaml + workflows |
| `gen` | `generate` | Render application.yml, or the deploy artifacts for the resolved platform |
| `lg` | `logs` | Print one instance log, where status says what but not why |
| `rm` | `remove` | Tear down what deploy created for a platform |
| `sts` | `status` | Report each instance: container (engine), application (connector), or all |
| `ver` | `version` | Print the utility name, version, Go version and OS/arch |
| `vld` | `validate` | Lint the whole env.yaml + workflows |

## Target abbreviations

The second (or third) word of a command, after the verb.

| Short | Stands for | Under | Summary |
|-------|------------|-------|---------|
| `app` | `application` | `status` | Report what the connector knows: leader-election state, health and workflows |
| `cfg` | `config` | `generate` | Emit application.yml |
| `cnt` | `container` | `status` | Report what the engine knows: state, restarts, age and image per instance |

## Platform abbreviations

Accepted as a `--platform` value by `generate`/`deploy`/`remove`/`status`/`logs`/`cli`, alongside the
canonical names.

| Short | Stands for |
|-------|------------|
| `dk` | `docker` |
| `kube` | `kubernetes` |
| `pm` | `podman` |

## Flag abbreviations

Only flags with a short form appear here; a flag spelled one way only (for
example `--platform`) is in the [commands.md flag table](commands.md#flags).

| Short | Stands for | Applies to | Meaning |
|-------|------------|------------|---------|
| `-d` | `--details` | `status` | add the enrichment lines each view can report: worker node, CPU/memory use against allocation, image digest and referenced components; app version, java version, config path and heap |
| `-e` | `--env` | all except `examples`/`download` | config file, relative or absolute path (default: `env.yaml`) |
| `-f` | `--force` | `examples`/`download` | overwrite existing files |
| `-o` | `--out` | `generate` | write output to a file (default: stdout) |
| `-w` | `--watch` | `status` | re-render the report every 5s until interrupted (Ctrl-C) |

## Notes

- An abbreviation is accepted wherever its canonical word is, but is never
  offered as a completion candidate and never shown in terminal help, so each
  menu and help page keeps exactly one spelling per command.
- `help` also answers to `-h` and `--help`. Those are flag spellings of the verb,
  not model aliases, which is why they are not in the table above.
- The platform short spellings are curated, not a prefix rule: only `kubernetes`
  has a widely recognized short form, and a prefix scheme would silently change
  meaning the day a platform is added.
