# solmq-conn-util command reference

<!-- GENERATED -- do not edit by hand.
Source of truth: cmd/solmq-conn-util/commands.go (the cliSpec model).
Regenerate: go test ./cmd/solmq-conn-util -run TestCommandsDocInSync -update
TestCommandsDocInSync fails the build if this file drifts from the model. -->

The full `solmq-conn-util` command tree. The first argument is a **verb**; where a verb
takes a second argument it names the **target** (`generate`) or **platform**
(`deploy`/`delete`). Generated from the command model in
[`cmd/solmq-conn-util/commands.go`](../cmd/solmq-conn-util/commands.go); see
[DEVELOPMENT.md](DEVELOPMENT.md#testing) to regenerate.

## Command tree

- `generate` -> `config` | `kubernetes` | `docker` | `podman`
- `deploy` -> `kubernetes` | `docker` | `podman`
- `delete` -> `kubernetes` | `docker` | `podman`
- `validate`
- `examples` `[dir]`
- `completion` -> `bash` | `zsh` | `fish` | `powershell`
- `help` (`-h`, `--help`)

## All commands

| Command | Summary |
|---------|---------|
| `solmq-conn-util generate config [-e env.yaml] [-o out]` | Emit application.yml |
| `solmq-conn-util generate kubernetes [-e env.yaml] [-o out]` | Emit ConfigMap+Deployment+Service (+Secrets) |
| `solmq-conn-util generate docker [-e env.yaml] [-o out]` | Emit docker-compose.yml (application.yml inlined) |
| `solmq-conn-util generate podman [-e env.yaml] [-o out]` | Emit a podman run script or quadlet unit |
| `solmq-conn-util deploy kubernetes [-e env.yaml] [--allow-command name]` | kubectl/oc apply -f - (manifest on stdin) |
| `solmq-conn-util deploy docker [-e env.yaml] [--allow-command name]` | docker compose up -d |
| `solmq-conn-util deploy podman [-e env.yaml] [--allow-command name]` | write the quadlet unit; systemctl start |
| `solmq-conn-util delete kubernetes [-e env.yaml] [--allow-command name]` | kubectl/oc delete -f - |
| `solmq-conn-util delete docker [-e env.yaml] [--allow-command name]` | docker compose down |
| `solmq-conn-util delete podman [-e env.yaml] [--allow-command name]` | systemctl stop; remove the unit |
| `solmq-conn-util validate [-e env.yaml]` | Lint the whole env.yaml + workflows |
| `solmq-conn-util examples [dir] [-f]` | Write a starter env.yaml + workflows |
| `solmq-conn-util completion bash` | Print the bash completion script |
| `solmq-conn-util completion zsh` | Print the zsh completion script |
| `solmq-conn-util completion fish` | Print the fish completion script |
| `solmq-conn-util completion powershell` | Print the PowerShell completion script |
| `solmq-conn-util help` | Print the usage summary (also -h, --help) |

## Flags

| Flag | Applies to | Meaning |
|------|-----------|---------|
| `-e`, `--env` | all except `examples` | config file, relative or absolute path (default: `env.yaml`) |
| `-o`, `--out` | `generate` | write output to a file (default: stdout) |
| `-f`, `--force` | `examples` | overwrite existing files |
| `--allow-command` | `deploy`/`delete` | approve an extra command binary beyond the `command:` allowlist; repeatable |

Flags may appear before, after, or between the positional arguments.

## Exit codes

| Code | Meaning |
|------|---------|
| `0` | success |
| `1` | processing error (bad input, unreadable file, missing env var, a deploy command that failed) |
| `2` | usage error (missing/unknown verb or target, unknown flag) |

## Command details

### generate

Renders the target's artifacts from `env.yaml` and prints them to stdout (or `-o`). Fails fast: stops at the first error and writes nothing; output is buffered, so a failed run never leaves a half-written `-o` file.

Flags: `-e`, `--env`; `-o`, `--out`.

#### `solmq-conn-util generate config [-e env.yaml] [-o out]`

Emit application.yml.

```sh
solmq-conn-util generate config -e env.yaml -o application.yml
```

#### `solmq-conn-util generate kubernetes [-e env.yaml] [-o out]`

Emit ConfigMap+Deployment+Service (+Secrets).

```sh
solmq-conn-util generate kubernetes -e env.yaml -o k8s.yaml
```

#### `solmq-conn-util generate docker [-e env.yaml] [-o out]`

Emit docker-compose.yml (application.yml inlined).

```sh
solmq-conn-util generate docker -e env.yaml -o docker-compose.yml
```

#### `solmq-conn-util generate podman [-e env.yaml] [-o out]`

Emit a podman run script or quadlet unit.

```sh
solmq-conn-util generate podman -e env.yaml -o run.sh
```

### deploy

Generates for the platform, then applies it by shelling out to the section's `command:` (`kubectl`/`oc`, `docker`, or `podman` + `systemctl`) through an argv slice -- never a shell. The env file must contain the matching section. `command:`'s argv[0] must be a bare, allowlisted binary name (path-free, PATH-resolved); `--allow-command` approves an extra binary for this invocation (e.g. a `sudo` prefix). Before anything is written or applied, a read-only preflight probe (login/permission check) must succeed, or the run stops with a login hint.

Flags: `-e`, `--env`; `--allow-command`.

#### `solmq-conn-util deploy kubernetes [-e env.yaml] [--allow-command name]`

kubectl/oc apply -f - (manifest on stdin).

```sh
solmq-conn-util deploy kubernetes -e env.yaml
```

#### `solmq-conn-util deploy docker [-e env.yaml] [--allow-command name]`

docker compose up -d.

```sh
solmq-conn-util deploy docker -e env.yaml
```

#### `solmq-conn-util deploy podman [-e env.yaml] [--allow-command name]`

write the quadlet unit; systemctl start.

```sh
solmq-conn-util deploy podman -e env.yaml
```

### delete

Tears down what `deploy` created for the platform, the same way (via the section's `command:`, the same binary allowlist, `--allow-command`, and the same read-only preflight probe before anything is torn down).

Flags: `-e`, `--env`; `--allow-command`.

#### `solmq-conn-util delete kubernetes [-e env.yaml] [--allow-command name]`

kubectl/oc delete -f -.

```sh
solmq-conn-util delete kubernetes -e env.yaml
```

#### `solmq-conn-util delete docker [-e env.yaml] [--allow-command name]`

docker compose down.

```sh
solmq-conn-util delete docker -e env.yaml
```

#### `solmq-conn-util delete podman [-e env.yaml] [--allow-command name]`

systemctl stop; remove the unit.

```sh
solmq-conn-util delete podman -e env.yaml
```

### validate

Runs every check across the whole `env.yaml` (including any `kubernetes:`/`docker:`/`podman:` sections) and its workflows, printing all findings. Non-zero exit if any errors. Use it as a linter.

Flags: `-e`, `--env`.

```sh
solmq-conn-util validate -e env.yaml
```

### examples

Writes a starter `env.yaml` plus workflow files into `dir` (default: the current directory). Use `-f` to overwrite existing files.

Flags: `-f`, `--force`.

```sh
solmq-conn-util examples ./myconfig
```

### completion

Prints a completion script for the named shell on stdout, for you to source or drop into the shell's completion directory (see the per-shell examples below). The script is rendered from the same command model as this help, so the completion a binary prints always matches the commands that binary accepts.

#### `solmq-conn-util completion bash`

Print the bash completion script.

```sh
solmq-conn-util completion bash > /etc/bash_completion.d/solmq-conn-util
```

#### `solmq-conn-util completion zsh`

Print the zsh completion script.

```sh
solmq-conn-util completion zsh > ~/.zsh/completions/_solmq-conn-util
```

#### `solmq-conn-util completion fish`

Print the fish completion script.

```sh
solmq-conn-util completion fish > ~/.config/fish/completions/solmq-conn-util.fish
```

#### `solmq-conn-util completion powershell`

Print the PowerShell completion script.

```sh
solmq-conn-util completion powershell > solmq-conn-util-completion.ps1
```

### help

Prints the usage summary. Same as `-h` / `--help`.

```sh
solmq-conn-util help
```
