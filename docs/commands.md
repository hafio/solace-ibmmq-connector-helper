# solmq-conn command reference

<!-- GENERATED -- do not edit by hand.
Source of truth: cmd/solmq-conn/commands.go (the cliSpec model).
Regenerate: go test ./cmd/solmq-conn -run TestCommandsDocInSync -update
TestCommandsDocInSync fails the build if this file drifts from the model. -->

The full `solmq-conn` command tree. The first argument is a **verb**; where a verb
takes a second argument it names the **target** (`generate`) or **platform**
(`deploy`/`delete`). Generated from the command model in
[`cmd/solmq-conn/commands.go`](../cmd/solmq-conn/commands.go); see
[DEVELOPMENT.md](DEVELOPMENT.md#testing) to regenerate.

## Command tree

- `generate` -> `config` | `kubernetes` | `docker` | `podman`
- `deploy` -> `kubernetes` | `docker` | `podman`
- `delete` -> `kubernetes` | `docker` | `podman`
- `validate`
- `examples` `[dir]`
- `help` (`-h`, `--help`)

## All commands

| Command | Summary |
|---------|---------|
| `solmq-conn generate config [-e env.yaml] [-o out]` | Emit application.yml |
| `solmq-conn generate kubernetes [-e env.yaml] [-o out]` | Emit ConfigMap+Deployment+Service (+Secrets) |
| `solmq-conn generate docker [-e env.yaml] [-o out]` | Emit docker-compose.yml (application.yml inlined) |
| `solmq-conn generate podman [-e env.yaml] [-o out]` | Emit a podman run script or quadlet unit |
| `solmq-conn deploy kubernetes [-e env.yaml] [--allow-command name]` | kubectl/oc apply -f - (manifest on stdin) |
| `solmq-conn deploy docker [-e env.yaml] [--allow-command name]` | docker compose up -d |
| `solmq-conn deploy podman [-e env.yaml] [--allow-command name]` | write the quadlet unit; systemctl start |
| `solmq-conn delete kubernetes [-e env.yaml] [--allow-command name]` | kubectl/oc delete -f - |
| `solmq-conn delete docker [-e env.yaml] [--allow-command name]` | docker compose down |
| `solmq-conn delete podman [-e env.yaml] [--allow-command name]` | systemctl stop; remove the unit |
| `solmq-conn validate [-e env.yaml]` | Lint the whole env.yaml + workflows |
| `solmq-conn examples [dir] [-f]` | Write a starter env.yaml + workflows |
| `solmq-conn help` | Print the usage summary (also -h, --help) |

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

#### `solmq-conn generate config [-e env.yaml] [-o out]`

Emit application.yml.

```sh
solmq-conn generate config -e env.yaml -o application.yml
```

#### `solmq-conn generate kubernetes [-e env.yaml] [-o out]`

Emit ConfigMap+Deployment+Service (+Secrets).

```sh
solmq-conn generate kubernetes -e env.yaml -o k8s.yaml
```

#### `solmq-conn generate docker [-e env.yaml] [-o out]`

Emit docker-compose.yml (application.yml inlined).

```sh
solmq-conn generate docker -e env.yaml -o docker-compose.yml
```

#### `solmq-conn generate podman [-e env.yaml] [-o out]`

Emit a podman run script or quadlet unit.

```sh
solmq-conn generate podman -e env.yaml -o run.sh
```

### deploy

Generates for the platform, then applies it by shelling out to the section's `command:` (`kubectl`/`oc`, `docker`, or `podman` + `systemctl`) through an argv slice -- never a shell. The env file must contain the matching section. `command:`'s argv[0] must be a bare, allowlisted binary name (path-free, PATH-resolved); `--allow-command` approves an extra binary for this invocation (e.g. a `sudo` prefix). Before anything is written or applied, a read-only preflight probe (login/permission check) must succeed, or the run stops with a login hint.

Flags: `-e`, `--env`; `--allow-command`.

#### `solmq-conn deploy kubernetes [-e env.yaml] [--allow-command name]`

kubectl/oc apply -f - (manifest on stdin).

```sh
solmq-conn deploy kubernetes -e env.yaml
```

#### `solmq-conn deploy docker [-e env.yaml] [--allow-command name]`

docker compose up -d.

```sh
solmq-conn deploy docker -e env.yaml
```

#### `solmq-conn deploy podman [-e env.yaml] [--allow-command name]`

write the quadlet unit; systemctl start.

```sh
solmq-conn deploy podman -e env.yaml
```

### delete

Tears down what `deploy` created for the platform, the same way (via the section's `command:`, the same binary allowlist, `--allow-command`, and the same read-only preflight probe before anything is torn down).

Flags: `-e`, `--env`; `--allow-command`.

#### `solmq-conn delete kubernetes [-e env.yaml] [--allow-command name]`

kubectl/oc delete -f -.

```sh
solmq-conn delete kubernetes -e env.yaml
```

#### `solmq-conn delete docker [-e env.yaml] [--allow-command name]`

docker compose down.

```sh
solmq-conn delete docker -e env.yaml
```

#### `solmq-conn delete podman [-e env.yaml] [--allow-command name]`

systemctl stop; remove the unit.

```sh
solmq-conn delete podman -e env.yaml
```

### validate

Runs every check across the whole `env.yaml` (including any `kubernetes:`/`docker:`/`podman:` sections) and its workflows, printing all findings. Non-zero exit if any errors. Use it as a linter.

Flags: `-e`, `--env`.

```sh
solmq-conn validate -e env.yaml
```

### examples

Writes a starter `env.yaml` plus workflow files into `dir` (default: the current directory). Use `-f` to overwrite existing files.

Flags: `-f`, `--force`.

```sh
solmq-conn examples ./myconfig
```

### help

Prints the usage summary. Same as `-h` / `--help`.

```sh
solmq-conn help
```
