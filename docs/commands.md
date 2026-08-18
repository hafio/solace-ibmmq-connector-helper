# solmq-conn-util command reference

<!-- GENERATED -- do not edit by hand.
Source of truth: cmd/solmq-conn-util/commands.go (the cliSpec model).
Regenerate: go test ./cmd/solmq-conn-util -run TestCommandsDocInSync -update
TestCommandsDocInSync fails the build if this file drifts from the model. -->

The full `solmq-conn-util` command tree. The first argument is a **verb**.
`generate` takes an optional second argument, `config`, to render
`application.yml` instead of a platform's artifacts. `generate`, `deploy`,
`delete`, and `status` all accept `--platform` to pick a **platform**
(`kubernetes`, `docker`, or `podman`) instead of resolving it from
`env.yaml`. Generated from the command model in
[`cmd/solmq-conn-util/commands.go`](../cmd/solmq-conn-util/commands.go); see
[DEVELOPMENT.md](DEVELOPMENT.md#testing) to regenerate.

## Command tree

- `generate` `[config]` `[--platform kubernetes|docker|podman]`
- `deploy` `[--platform kubernetes|docker|podman]`
- `delete` `[--platform kubernetes|docker|podman]`
- `status` `[--install]` `[--platform kubernetes|docker|podman]`
- `version`
- `validate`
- `examples` `[dir]`
- `completion` -> `bash` | `zsh` | `fish` | `powershell`
- `help` (`-h`, `--help`)

## All commands

| Command | Summary |
|---------|---------|
| `solmq-conn-util generate config [--platform kubernetes\|docker\|podman] [-e env.yaml] [-o out]` | Emit application.yml |
| `solmq-conn-util generate [--platform kubernetes\|docker\|podman] [-e env.yaml] [-o out]` | Render the artifacts for the resolved platform to stdout or a file |
| `solmq-conn-util deploy [--platform kubernetes\|docker\|podman] [-e env.yaml] [--allow-command name]` | Generate for a platform, then apply it |
| `solmq-conn-util delete [--platform kubernetes\|docker\|podman] [-e env.yaml] [--allow-command name]` | Tear down what deploy created for a platform |
| `solmq-conn-util status [--install] [--platform kubernetes\|docker\|podman] [-e env.yaml] [--pod name] [--container name] [--namespace ns] [--management-port port] [--user name] [--command name] [--allow-command name]` | Ensure and run the status script, printing per-target leader-election and workflow state |
| `solmq-conn-util version` | Print the utility name, version, Go version and OS/arch |
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
| `--allow-command` | `deploy`/`delete`/`status` | approve an extra command binary beyond the `command:` allowlist; repeatable |
| `--platform` | `generate`/`deploy`/`delete`/`status` | the platform: `kubernetes`, `docker`, or `podman` (default: resolved from env.yaml, or an interactive menu -- see Command details) |
| `--install` | `status` | install the status script on every target without prompting |
| `--pod` | `status` | limit checks to this kubernetes pod name; repeatable (default: every running pod); no effect on docker/podman |
| `--container` | `status` | limit checks to this docker/podman container name; repeatable (default: every running container); no effect on kubernetes |
| `--namespace` | `status` | kubernetes namespace to query (default: the namespace of the deployment in env.yaml); no effect on docker/podman |
| `--management-port` | `status` | actuator management port to reach inside each target (default: the management port configured for the target) |
| `--user` | `status` | actuator account the status script authenticates as (default `solmq-status`) |
| `--command` | `status` | override the platform CLI binary (`kubectl`/`oc`, `docker`, or `podman`) used to reach each target, instead of the `command:` in that section |

Flags may appear before, after, or between the positional arguments.

## Exit codes

| Code | Meaning |
|------|---------|
| `0` | success |
| `1` | processing error (bad input, unreadable file, missing env var, a deploy command that failed) |
| `2` | usage error (missing/unknown verb or target, unknown flag) |

## Command details

### generate

Renders artifacts and prints them to stdout (or `-o`). Fails fast: stops at the first error and writes nothing; output is buffered, so a failed run never leaves a half-written `-o` file. The `config` positional renders `application.yml` from env.yaml and never involves a platform (`--platform` is ignored); leaving it off renders the resolved platform's artifacts instead. The platform is resolved in order: `--platform`, if given; otherwise the single `kubernetes:`/`docker:`/`podman:` section in env.yaml, when exactly one is present; otherwise an interactive menu, when more than one is present. A `--platform` value with no matching section in env.yaml is a loud error, and so are zero sections. The menu -- and, under `status`, the install confirmation prompt -- never block when stdin is not a TTY; both fail with the same guidance instead of hanging.

Flags: `--platform`; `-e`, `--env`; `-o`, `--out`.

```sh
solmq-conn-util generate --platform kubernetes -e env.yaml -o k8s.yaml
```

#### `solmq-conn-util generate config [--platform kubernetes|docker|podman] [-e env.yaml] [-o out]`

Emit application.yml.

```sh
solmq-conn-util generate config -e env.yaml -o application.yml
```

### deploy

Generates for the platform, then applies it by shelling out to the section's `command:` (`kubectl`/`oc`, `docker`, or `podman` + `systemctl`) through an argv slice -- never a shell. The env file must contain the matching section. `command:`'s argv[0] must be a bare, allowlisted binary name (path-free, PATH-resolved); `--allow-command` approves an extra binary for this invocation (e.g. a `sudo` prefix). Before anything is written or applied, a read-only preflight probe (login/permission check) must succeed, or the run stops with a login hint. The platform is resolved in order: `--platform`, if given; otherwise the single `kubernetes:`/`docker:`/`podman:` section in env.yaml, when exactly one is present; otherwise an interactive menu, when more than one is present. A `--platform` value with no matching section in env.yaml is a loud error, and so are zero sections. The menu -- and, under `status`, the install confirmation prompt -- never block when stdin is not a TTY; both fail with the same guidance instead of hanging.

Flags: `--platform`; `-e`, `--env`; `--allow-command`.

```sh
solmq-conn-util deploy --platform kubernetes -e env.yaml
```

### delete

Tears down what `deploy` created for the platform, the same way (via the section's `command:`, the same binary allowlist, `--allow-command`, and the same read-only preflight probe before anything is torn down). The platform is resolved in order: `--platform`, if given; otherwise the single `kubernetes:`/`docker:`/`podman:` section in env.yaml, when exactly one is present; otherwise an interactive menu, when more than one is present. A `--platform` value with no matching section in env.yaml is a loud error, and so are zero sections. The menu -- and, under `status`, the install confirmation prompt -- never block when stdin is not a TTY; both fail with the same guidance instead of hanging.

Flags: `--platform`; `-e`, `--env`; `--allow-command`.

```sh
solmq-conn-util delete --platform kubernetes -e env.yaml
```

### status

For the resolved platform, execs into each running instance (a `kubernetes` pod, or a `docker`/`podman` container) and ensures the status script is present: `--install` installs it without asking; without it, a declined per-target install prompt just skips that target. Then runs the script and prints, per target, the leader-election result and workflow state from the instance's own actuator endpoint. `--pod` and `--container` (both repeatable) narrow which running targets are checked; `--namespace` overrides the kubernetes namespace and `--management-port` the actuator port; `--user` names the read-only actuator account the script authenticates as, for an instance whose config does not carry the reserved `solmq-status` account. `--command` overrides the platform CLI binary used to reach each target, and `--allow-command` approves an extra one, the same as deploy/delete. The platform is resolved in order: `--platform`, if given; otherwise the single `kubernetes:`/`docker:`/`podman:` section in env.yaml, when exactly one is present; otherwise an interactive menu, when more than one is present. A `--platform` value with no matching section in env.yaml is a loud error, and so are zero sections. The menu -- and, under `status`, the install confirmation prompt -- never block when stdin is not a TTY; both fail with the same guidance instead of hanging.

Flags: `--install`; `--platform`; `-e`, `--env`; `--pod`; `--container`; `--namespace`; `--management-port`; `--user`; `--command`; `--allow-command`.

```sh
solmq-conn-util status --platform kubernetes -e env.yaml
```

### version

Prints solmq-conn-util's own version (stamped in at build time), the Go version it was built with, and its OS/arch (`GOOS`/`GOARCH`) -- for bug reports and to confirm which build is installed. Takes no flags.

```sh
solmq-conn-util version
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
