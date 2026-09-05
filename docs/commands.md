# solmq-conn-util command reference

<!-- GENERATED -- do not edit by hand.
Source of truth: cmd/solmq-conn-util/commands.go (the cliSpec model).
Regenerate: go test ./cmd/solmq-conn-util -run TestCommandsDocInSync -update
TestCommandsDocInSync fails the build if this file drifts from the model. -->

The full `solmq-conn-util` command tree. The first argument is a **verb**.
`generate` takes an optional second argument, `config`, to render
`application.yml` instead of a platform's artifacts.
`generate`/`deploy`/`remove`/`status`/`logs`/`cli` all accept
`--platform` to pick a **platform**
(`kubernetes`, `docker`, or `podman`) instead of resolving it from
`env.yaml`. Every short spelling in here is also listed on its own, keyed
by the abbreviation, in [abbreviation.md](abbreviation.md). Generated from the
command model in
[`cmd/solmq-conn-util/commands.go`](../cmd/solmq-conn-util/commands.go); see
[DEVELOPMENT.md](DEVELOPMENT.md#testing) to regenerate.

## Command tree

- `generate` (`gen`) `[config]` `[--platform kubernetes|docker|podman]`
- `deploy` (`dp`) `[--platform kubernetes|docker|podman]`
- `remove` (`rm`) `[--platform kubernetes|docker|podman]` `[--no-prompt]`
- `status` (`sts`) `<container|application|all>` `[--details]` `[--platform kubernetes|docker|podman]`
- `logs` (`lg`) `[--follow]` `[--previous]` `[--platform kubernetes|docker|podman]`
- `cli` `[-- command ...]` `[--platform kubernetes|docker|podman]`
- `version` (`ver`)
- `validate` (`vld`)
- `examples` (`eg`) `[dir]`
- `download` (`dl`) -> `jar` -> `mq` | `syslog`
- `auto-complete` -> `bash` | `zsh` | `fish` | `powershell`
- `help` (`-h`, `--help`)

## All commands

| Command | Summary |
|---------|---------|
| `solmq-conn-util generate config [--platform kubernetes\|docker\|podman] [-e env.yaml] [-o out]` | Emit application.yml |
| `solmq-conn-util generate [--platform kubernetes\|docker\|podman] [-e env.yaml] [-o out]` | Render the artifacts for the resolved platform to stdout or a file |
| `solmq-conn-util deploy [--platform kubernetes\|docker\|podman] [-e env.yaml] [--allow-command name]` | Generate for a platform, then apply it |
| `solmq-conn-util remove [--platform kubernetes\|docker\|podman] [--no-prompt] [-e env.yaml] [--allow-command name]` | Tear down what deploy created for a platform |
| `solmq-conn-util status container [--details] [--watch] [--all] [--output table\|json] [--install] [--platform kubernetes\|docker\|podman] [-e env.yaml] [--pod name] [--container name] [--namespace ns] [--management-port port] [--user name] [--command name] [--allow-command name]` | Report what the engine knows: state, restarts, age and image per instance |
| `solmq-conn-util status application [--details] [--watch] [--all] [--output table\|json] [--install] [--platform kubernetes\|docker\|podman] [-e env.yaml] [--pod name] [--container name] [--namespace ns] [--management-port port] [--user name] [--command name] [--allow-command name]` | Report what the connector knows: leader-election state, health and workflows |
| `solmq-conn-util status all [--details] [--watch] [--all] [--output table\|json] [--install] [--platform kubernetes\|docker\|podman] [-e env.yaml] [--pod name] [--container name] [--namespace ns] [--management-port port] [--user name] [--command name] [--allow-command name]` | Report both halves: the container table, then the application block per instance |
| `solmq-conn-util logs [--follow] [--previous] [--tail N] [--since d] [--timestamps] [--platform kubernetes\|docker\|podman] [-e env.yaml] [--pod name\|index] [--container name\|index] [--namespace ns] [--command name] [--allow-command name]` | Print one instance log, where status says what but not why |
| `solmq-conn-util cli [--platform kubernetes\|docker\|podman] [-e env.yaml] [--pod name\|index] [--container name\|index] [--namespace ns] [--command name] [--allow-command name] [-- command ...]` | Open a shell inside one instance, or run one command in it |
| `solmq-conn-util version` | Print the utility name, version, Go version and OS/arch |
| `solmq-conn-util validate [-e env.yaml]` | Lint the whole env.yaml + workflows |
| `solmq-conn-util examples [dir] [-f]` | Write a starter env.yaml + workflows |
| `solmq-conn-util download jar mq [dir] [-e env.yaml] [--url u] [--version v] [--omit-lib-file file] [--include-provided] [-f]` | Download the IBM MQ client jar and its dependencies |
| `solmq-conn-util download jar syslog [dir] [-e env.yaml] [--url u] [--version v] [--omit-lib-file file] [--include-provided] [-f]` | Download the logstash syslog encoder jar and its dependencies |
| `solmq-conn-util auto-complete bash` | Print the bash completion script |
| `solmq-conn-util auto-complete zsh` | Print the zsh completion script |
| `solmq-conn-util auto-complete fish` | Print the fish completion script |
| `solmq-conn-util auto-complete powershell` | Print the PowerShell completion script |
| `solmq-conn-util help [verb]` | Print this summary, or the help page of one verb |

## Flags

| Flag | Applies to | Meaning |
|------|-----------|---------|
| `-e`, `--env` | all except `examples`/`version`/`auto-complete`/`help` | config file, relative or absolute path (default: `env.yaml`) |
| `-o`, `--out` | `generate` | write output to a file (default: stdout) |
| `-f`, `--force` | `examples`/`download` | overwrite existing files |
| `--allow-command` | `deploy`/`remove`/`status`/`logs`/`cli` | approve an extra command binary beyond the `command:` allowlist; repeatable |
| `--platform` | `generate`/`deploy`/`remove`/`status`/`logs`/`cli` | the platform: `kubernetes`, `docker`, or `podman` (short: `kube`, `dk`, `pm`; default: resolved from env.yaml, or an interactive menu -- see [Platform resolution](commands.md#platform-resolution)) |
| `--url` | `download` | exact URL to download instead of Maven resolution; repeatable; when given, no resolution happens at all |
| `--version` | `download` | pin the seed release (the IBM MQ client jar, or the syslog encoder jar) instead of resolving latest stable; empty means latest stable |
| `--omit-lib-file` | `download` | a jar list that replaces (never merges with) the embedded default the omission rule compares against; an empty file omits nothing |
| `--include-provided` | `download` | download the whole closure even where the connector image already provides a jar, instead of omitting it |
| `--install` | `status` | install the status script on every instance without prompting |
| `--no-prompt` | `remove` | tear down without asking anything -- what a script or CI job passes, since the prompts refuse to read a non-TTY rather than hang. It covers both questions: the teardown confirmation, and whether to remove a namespace that turned out to be empty. It cannot authorise more than that: a namespace holding anything this release does not own is never removed, with or without it |
| `-d`, `--details` | `status` | add the enrichment lines each view can report: worker node, CPU/memory use against allocation, image digest and referenced components; app version, java version, config path and heap |
| `-w`, `--watch` | `status` | re-render the report every 5s until interrupted (Ctrl-C) |
| `--all` | `status` | reach every connector instance found by image name (`solace-pubsub-connector-ibmmq`) instead of the ones `env.yaml` describes -- every namespace on kubernetes, every container on docker/podman; cannot be combined with `--pod`/`--container` |
| `--output` | `status` | output format: `table` (default) or `json`, one machine-readable document per run; `json` cannot be combined with `--watch` |
| `--pod` | `status`/`logs`/`cli` | the kubernetes pod to reach, by name or by index into the listed order (alphabetical, the order `status` prints); a name always wins over the index reading. Repeatable on `status`; on `logs` and `cli`, which reach one instance, it may be given once. Default: every running pod on `status`, and on `logs`/`cli` the matching instances are listed instead. No effect on docker/podman |
| `--container` | `status`/`logs`/`cli` | the docker/podman container to reach, by name or by index into the listed order (alphabetical, the order `status` prints); a name always wins over the index reading. Repeatable on `status`; on `logs` and `cli`, which reach one instance, it may be given once. Default: every running container on `status`, and on `logs`/`cli` the one the section in env.yaml names. No effect on kubernetes |
| `--namespace` | `status`/`logs`/`cli` | kubernetes namespace to query (default: the namespace of the deployment in env.yaml); no effect on docker/podman |
| `--management-port` | `status` | actuator management port to reach inside each instance (default: the configured management port) |
| `--user` | `status` | actuator account the status script authenticates as (default `solmq-status`) |
| `--command` | `status`/`logs`/`cli` | override the platform CLI binary (`kubectl`/`oc`, `docker`, or `podman`) used to reach each instance, instead of the `command:` in that section |
| `--follow` | `logs` | keep the log open and print new lines as they arrive, until interrupted (Ctrl-C); reads one instance, so it cannot be combined with `--all` or `--previous` |
| `--previous` | `logs` | read the log of the previous container instead of the running one -- what a pod that is restarting printed before it died; kubernetes only, since neither docker nor podman keeps a prior run under the same name |
| `--tail` | `logs` | read only the last N lines, or `all` for the whole log (default: `all`) |
| `--since` | `logs` | read only lines newer than this duration, spelled as a Go duration (`30s`, `10m`, `2h`) |
| `--timestamps` | `logs` | prefix every line with the time the platform recorded for it |

Flags may appear before, after, or between the positional arguments.

## Platform resolution

The platform is resolved in order: `--platform` (which accepts the short spellings `kube`, `dk` and `pm`), if given; otherwise the single `kubernetes:`/`docker:`/`podman:` section in env.yaml, when exactly one is present; otherwise an interactive menu, when more than one is present. A `--platform` value with no matching section in env.yaml is a loud error, and so are zero sections.

`status`, `logs` and `cli` are the exception, and only when the operator has already named the instances themselves (`--pod`/`--container`, or `--all` under `status`) alongside an explicit `--platform`: there is then nothing left to read from env.yaml, so a missing file and a section-less platform are both fine -- which is how an instance this tool never deployed is reached.

The menu -- and `status`'s install confirmation, and `remove`'s teardown confirmation -- never block when stdin is not a TTY; all three fail with the same guidance (naming the flag that skips them) instead of hanging.

## Exit codes

| Code | Meaning |
|------|---------|
| `0` | success |
| `1` | processing error (bad input, unreadable file, missing env var, a deploy command that failed) |
| `2` | usage error (missing/unknown verb or target, unknown flag) |
| _other_ | `cli` only: once a session or command is running inside the instance, its own exit status is passed straight back. Anything that fails before that still uses the three codes above |

## Command details

### generate

Alias: `gen`.

Renders artifacts and prints them to stdout (or `-o`). Fails fast: stops at the first error and writes nothing; output is buffered, so a failed run never leaves a half-written `-o` file.

The `config` positional renders `application.yml` from env.yaml and never involves a platform (`--platform` is ignored); leaving it off renders the resolved platform's artifacts instead. For how the platform is picked, see [Platform resolution](#platform-resolution).

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

Alias: `dp`.

Generates for the platform, then applies it by shelling out to the section's `command:` (`kubectl`/`oc`, `docker`, or `podman` + `systemctl`) through an argv slice -- never a shell. The env file must contain the matching section.

`command:`'s argv[0] must be a bare, allowlisted binary name (path-free, PATH-resolved); `--allow-command` approves an extra binary for this invocation (e.g. a `sudo` prefix). Before anything is written or applied, a read-only preflight probe (login/permission check) must succeed, or the run stops with a login hint. For how the platform is picked, see [Platform resolution](#platform-resolution).

Flags: `--platform`; `-e`, `--env`; `--allow-command`.

```sh
solmq-conn-util deploy --platform kubernetes -e env.yaml
```

### remove

Alias: `rm`.

Tears down what `deploy` created for the platform, the same way (via the section's `command:`, the same binary allowlist, `--allow-command`, and the same read-only preflight probe before anything is torn down). For how the platform is picked, see [Platform resolution](#platform-resolution).

Because it is the one verb that destroys rather than creates, it asks for confirmation first, naming what it is about to tear down -- the deployment and its namespace on kubernetes, the container on docker/podman -- so a run pointed at the wrong `env.yaml` is caught before anything is deleted. Answering anything but `y`/`yes` cancels and exits 0, having touched nothing at all: the question comes before even the preflight probe. `--no-prompt` skips the prompt, which is what a script or CI job passes -- without it a non-TTY run fails fast with that hint rather than hanging on a read that will never return.

On kubernetes the namespace is handled separately, and never as part of the manifest delete: deleting a Namespace cascades to every object inside it, including workloads this tool never deployed. So once the teardown succeeds, `remove` looks at what is left in the namespace. Anything this release does not own -- another deployment, a stateful set, a volume claim, a secret -- is listed and the namespace is kept. Only a namespace with nothing else in it is offered for removal, as its own separate question.

That is an invariant rather than a default: no flag removes a namespace that still holds someone else's work. The cluster namespaces (`default`, `kube-system`, `kube-public`, `kube-node-lease`) are never removed either, and a check that cannot run leaves the namespace alone rather than assuming it is empty.

Flags: `--no-prompt`; `--platform`; `-e`, `--env`; `--allow-command`.

```sh
solmq-conn-util remove --platform kubernetes -e env.yaml
```

### status

Alias: `sts`.

Reports the state of every connector instance of the resolved platform. The target word picks which half is reported, because they answer different questions and come from different places: `container` is what the container engine knows, read from outside through read-only `kubectl`/`docker`/`podman` queries -- state, restarts, age and the image actually running, in one table per platform; `application` is what the connector knows about itself, read from inside by running the generated status script in each instance -- leader-election mode and state, health, and per-workflow state; `all` reports both, container first, since a container that is not running is the reason the application half is missing. Each word has a short spelling (`cnt`, `app`); the word itself is required, and `status` on its own prints this list.

`-d`/`--details` adds the enrichment lines to whichever view is being printed: worker node, CPU and memory use against allocation, the image digest, and the objects the workload references (secrets, config maps, volume claims, mounts) on the container side; app version, java version, the configuration file the report was read from, and JVM heap use on the application side. The one sampling query it needs (`kubectl top`, `docker stats`) is why those lines are opt-in: on kubernetes it also needs a metrics API in the cluster, and reports a note instead of the lines when there is none.

`-w`/`--watch` re-renders the report every 5s until interrupted. `--output` `json` emits one machine-readable document per run instead of the tables, carrying every fact either view collected. `--all` ignores the instance names in `env.yaml` and reports every connector instance it can find by image name (`solace-pubsub-connector-ibmmq`) -- across every namespace on kubernetes, and every container, running or not, on docker/podman.

For the application views, `--install` installs the status script without asking where it is missing; without it, a declined install prompt just skips the instances that lack it. `--pod` and `--container` (both repeatable) narrow which instances are reported; `--namespace` overrides the kubernetes namespace and `--management-port` the actuator port; `--user` names the read-only actuator account the script authenticates as, for an instance whose config does not carry the reserved `solmq-status` account. `--command` overrides the platform CLI binary used to reach each instance, and `--allow-command` approves an extra one, the same as deploy/remove. For how the platform is picked, see [Platform resolution](#platform-resolution).

Flags: `-d`, `--details`; `-w`, `--watch`; `--all`; `--output`; `--install`; `--platform`; `-e`, `--env`; `--pod`; `--container`; `--namespace`; `--management-port`; `--user`; `--command`; `--allow-command`.

#### `solmq-conn-util status container [--details] [--watch] [--all] [--output table|json] [--install] [--platform kubernetes|docker|podman] [-e env.yaml] [--pod name] [--container name] [--namespace ns] [--management-port port] [--user name] [--command name] [--allow-command name]`

Report what the engine knows: state, restarts, age and image per instance.

```sh
solmq-conn-util status container --platform kubernetes -e env.yaml
```

#### `solmq-conn-util status application [--details] [--watch] [--all] [--output table|json] [--install] [--platform kubernetes|docker|podman] [-e env.yaml] [--pod name] [--container name] [--namespace ns] [--management-port port] [--user name] [--command name] [--allow-command name]`

Report what the connector knows: leader-election state, health and workflows.

```sh
solmq-conn-util status application -e env.yaml
```

#### `solmq-conn-util status all [--details] [--watch] [--all] [--output table|json] [--install] [--platform kubernetes|docker|podman] [-e env.yaml] [--pod name] [--container name] [--namespace ns] [--management-port port] [--user name] [--command name] [--allow-command name]`

Report both halves: the container table, then the application block per instance.

```sh
solmq-conn-util status all -d
```

### logs

Alias: `lg`.

Prints what one connector instance of the resolved platform has written, read through the same read-only `kubectl`/`docker`/`podman` path `status` uses -- the same `command:`, the same binary allowlist, `--allow-command`, and the same preflight probe -- and discovering the same instances from `env.yaml`, so the two verbs can never disagree about which ones they mean. It is the answer to the question `status` leaves open: that view reports a restart count and an exit code, and this one reports the lines that preceded them.

One instance is read per run. When discovery finds several and none was named, nothing is read: the matching instances are listed on stdout as commands that can be pasted back verbatim, carrying the flags already typed, and the run exits 0. `--pod` and `--container` name the instance, and each may be given once -- a second is refused rather than silently losing the first. Either accepts an **index** as well as a name: `--pod 0` is the first instance in the listed order, which is alphabetical by name and the same order `status` prints. A name always wins, so an instance genuinely called `0` is still reachable by name.

`--previous` is the pairing that matters most -- it reads what the previous container printed before it died, which is the only place a crash loop explains itself; kubernetes keeps that log, docker and podman do not, so it is refused there rather than quietly ignored. `--follow` keeps the log open and prints new lines until interrupted (Ctrl-C, which is a clean exit, not a failure). `--tail` limits how far back the read starts (`--tail all` is the default), `--since` limits it by age, and `--timestamps` prefixes each line with the time the platform recorded.

`--namespace` overrides the kubernetes namespace, `--command` overrides the platform CLI binary, and `--allow-command` approves an extra one, the same as status. The log itself goes to stdout and everything else to stderr, so `logs > app.log` captures the log and nothing but. For how the platform is picked, see [Platform resolution](#platform-resolution).

Flags: `--follow`; `--previous`; `--tail`; `--since`; `--timestamps`; `--platform`; `-e`, `--env`; `--pod`; `--container`; `--namespace`; `--command`; `--allow-command`.

```sh
solmq-conn-util logs --tail 100 -e env.yaml
```

### cli

Opens an interactive shell inside one connector instance of the resolved platform, reached through the same read-only `kubectl`/`docker`/`podman` path `status` and `logs` use -- the same `command:`, the same binary allowlist, `--allow-command`, the same preflight probe, and the same instance discovery from `env.yaml`, so the three verbs can never disagree about which instance they mean. It is where the questions `status` and `logs` cannot answer get settled: whether the truststore really mounted, what is actually in `/app/external/libs`, which `application.yml` the process is running. For how the platform is picked, see [Platform resolution](#platform-resolution).

Everything after `--` is run in the instance instead of a shell, non-interactively, which is the form for a script -- an interactive run with no terminal on stdin is refused with that as the next step rather than opening a session nobody can type into. The shell is `sh`: the connector image is Alpine, so its userland is busybox and there is no `bash` to ask for.

One instance is reached per run. When discovery finds several and none was named, nothing is opened: the matching instances are listed on stdout as commands that can be pasted back verbatim, carrying the flags already typed, and the run exits 0. `--pod` and `--container` name the instance, each may be given once, and either accepts an **index** as well as a name -- `--pod 0` is the first instance in the listed order, alphabetical by name and the same order `status` prints. A name always wins. On kubernetes the container is always named explicitly (`-c connector`), so a pod carrying a sidecar cannot be entered by mistake; a pod with no such container is refused by the platform rather than guessed at.

Every token of a `--` command is held to the same safe charset as every other value that reaches an argv, because `cli` runs an argv and never a shell line: a pipe, a redirect or a glob has to be written inside the session instead. **Exit status** is the one place `cli` leaves the usual contract: once the session or command is running, whatever it exited with is passed straight back. The engines give no way to be more precise -- `kubectl exec` reports an unreachable pod and a command that exited non-zero with the same status -- so a non-zero `cli` exit means one of the two, and the message the engine printed on stderr is what says which.

Flags: `--platform`; `-e`, `--env`; `--pod`; `--container`; `--namespace`; `--command`; `--allow-command`.

```sh
solmq-conn-util cli -e env.yaml
```

### version

Alias: `ver`.

Prints solmq-conn-util's own version (stamped in at build time), the Go version it was built with, and its OS/arch (`GOOS`/`GOARCH`) -- for bug reports and to confirm which build is installed. Takes no flags.

```sh
solmq-conn-util version
```

### validate

Alias: `vld`.

Runs every check across the whole `env.yaml` (including any `kubernetes:`/`docker:`/`podman:` sections) and its workflows, printing all findings. Non-zero exit if any errors. Use it as a linter.

Flags: `-e`, `--env`.

```sh
solmq-conn-util validate -e env.yaml
```

### examples

Alias: `eg`.

Writes a starter `env.yaml` plus workflow files into `dir` (default: the current directory). Use `-f` to overwrite existing files.

Flags: `-f`, `--force`.

```sh
solmq-conn-util examples ./myconfig
```

### download

Alias: `dl`.

Downloads a fixed set of jars and their dependencies into a directory. All three words -- `download`, then `jar`, then `mq` or `syslog` -- are required; a missing or unknown word is a loud error listing the valid words. `mq` seeds from the IBM MQ client jar; `syslog` seeds from the logstash syslog encoder jar. The `mq` seed is the Jakarta build of the client, and there is no flag to change it: the connector image is a Jakarta stack, so the javax build could only ever produce a classpath that fails at run time.

`-e` is read for one thing only: the `image` block, so the command can say when the jar list it omits against does not describe the image being deployed. It reads no credentials, no platform and no workflows, and a missing env.yaml is not an error -- download is the verb you run before you have a deployment.

The seed artifact resolves to its latest stable release, or to the exact release named by `--version` when given; an empty value means latest stable, the same as leaving the flag off. Every dependency version instead comes from the Maven POM chain of the resolved seed release. `--url` (repeatable) overrides all of that: when given, exactly those URLs are downloaded and no Maven resolution happens.

By default, an artifact resolved through Maven is omitted when the connector image already ships that jar, matched by artifact base name, at a version equal to or newer than the one resolved here; every omission is reported, never silent. The seed artifact -- the jar the command was run to fetch in the first place -- is never a candidate for omission no matter what the omit file says about it, so the command stays useful against an older image that lacks it entirely, and a stale or hostile omit file can never cause the one jar that matters to be skipped. `--omit-lib-file` replaces the embedded jar list (captured from the shipped connector image) with one captured from a different image, so the comparison runs against that image instead; it replaces the embedded list completely rather than merging with it, so an omit file containing nothing omits nothing. `--include-provided` disables omission entirely and downloads the whole closure regardless of what the image already has. Omission never applies to an explicit `--url`: the operator named that URL directly, so it is always downloaded verbatim and never second-guessed.

Matching is by jar artifact base name plus version, since a jar filename carries no groupId; this is why Jackson 3 (`tools.jackson.core`) still downloads for the `syslog` set even though the image already ships Jackson 2: its 3.x versions compare higher than the 2.x copies the image carries, so the version comparison still gets the right answer.

The destination is the trailing `[dir]` positional (default `./libs`). Every jar is checked against the sha1 digest the repository publishes beside it before it is written, catching a truncated or corrupted transfer; that is integrity, not authenticity -- it is not proof against a compromised repository. https is still required on the initial URL and on every redirect hop. An existing file is skipped unless `-f` is given, exactly like `examples`.

Flags: `-e`, `--env`; `--url`; `--version`; `--omit-lib-file`; `--include-provided`; `-f`, `--force`.

#### `solmq-conn-util download jar mq [dir] [-e env.yaml] [--url u] [--version v] [--omit-lib-file file] [--include-provided] [-f]`

Download the IBM MQ client jar and its dependencies.

```sh
solmq-conn-util download jar mq ./libs
```

#### `solmq-conn-util download jar syslog [dir] [-e env.yaml] [--url u] [--version v] [--omit-lib-file file] [--include-provided] [-f]`

Download the logstash syslog encoder jar and its dependencies.

```sh
solmq-conn-util download jar syslog ./libs
```

### auto-complete

Prints a completion script for the named shell on stdout, for you to source or drop into the shell's completion directory (see the per-shell examples below). The script is rendered from the same command model as this help, so the completion a binary prints always matches the verbs that binary accepts.

#### `solmq-conn-util auto-complete bash`

Print the bash completion script.

```sh
solmq-conn-util auto-complete bash > /etc/bash_completion.d/solmq-conn-util
```

#### `solmq-conn-util auto-complete zsh`

Print the zsh completion script.

```sh
solmq-conn-util auto-complete zsh > ~/.zsh/completions/_solmq-conn-util
```

#### `solmq-conn-util auto-complete fish`

Print the fish completion script.

```sh
solmq-conn-util auto-complete fish > ~/.config/fish/completions/solmq-conn-util.fish
```

#### `solmq-conn-util auto-complete powershell`

Print the PowerShell completion script.

```sh
solmq-conn-util auto-complete powershell > solmq-conn-util-completion.ps1
```

### help

Prints the command summary (same as `-h` / `--help`). With a verb name, prints that verb's own page -- its arguments, flags, and examples -- exactly like `<verb> -h`. Requested help goes to stdout and exits 0; the same pages printed after a usage mistake go to stderr with exit 2.

```sh
solmq-conn-util help status
```
