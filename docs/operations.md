# solmq-conn-util operations reference

> [!WARNING]
> **Not a supported Solace product.** `solmq-conn-util` was created by Solace
> Professional Services and is supported only by Solace Professional Services --
> not by Solace Support. For help with this tool, contact your Solace
> Professional Services representative rather than opening a Solace Support
> case. This notice covers this tool only, not the Solace PubSub+ Connector for
> IBM MQ that it configures and deploys.

Every verb, what it runs, and what it prints. The `env.yaml` keys are in
[configuration.md](configuration.md); the generated flag reference with an
example per command is [commands.md](commands.md); every short spelling is in
[abbreviation.md](abbreviation.md).

## Contents

1. [Commands](#1-commands)
2. [validate](#2-validate)
3. [generate](#3-generate)
4. [deploy](#4-deploy)
5. [remove](#5-remove)
6. [start, stop, restart](#6-start-stop-restart)
7. [status](#7-status)
8. [logs](#8-logs)
9. [cli](#9-cli)
10. [download jar](#10-download-jar)
11. [What gets generated](#11-what-gets-generated)
12. [Shell completion](#12-shell-completion)

## 1. Commands

| Verb | Alias | Does |
|---|---|---|
| `generate [config]` | `gen` | render `application.yml` (`config`), or the platform's deploy artifacts |
| `deploy` | `dp` | generate for a platform, then apply it |
| `remove` | `rm` | tear down what `deploy` created |
| `start` / `stop` / `restart` | | act on the instance in place |
| `status container\|application\|all` | `sts` | report each instance |
| `logs` | `lg` | print one instance's log |
| `cli` | | open a shell in one instance, or run one command |
| `validate` | `vld` | lint `env.yaml` and the workflows |
| `examples [dir]` | `eg` | write a starter set |
| `download jar mq\|syslog [dir]` | `dl` | fetch jars from Maven Central |
| `auto-complete bash\|zsh\|fish\|powershell` | | print a completion script |
| `version` | `ver` | print the build's version |
| `help [verb]` | | usage, or one verb's page |

`-h` lists one line per verb; `help <verb>` or `<verb> -h` prints that verb's
flags and examples. Flags may go before, after or between positional arguments.

**Platform.** `generate`, `deploy`, `remove`, `start`, `stop`, `restart`,
`status`, `logs` and `cli` resolve the platform in this order:

1. `--platform kubernetes|docker|podman` (short `kube`, `dk`, `pm`). It must
   name a section present in `env.yaml`.
2. Otherwise the single platform section present in `env.yaml`.
3. Otherwise an interactive menu over the sections present.
4. No section at all is an error.

**Non-TTY rule.** The menu, the `status` install prompt, the `remove`
confirmation and an interactive `cli` all fail immediately when stdin is not a
terminal, naming the flag to pass instead. Scripts and CI pass `--platform`,
`--install` and `--no-prompt` explicitly.

**Exit codes.** 0 success; 1 a processing error (bad input, a missing variable,
a failed platform command); 2 a usage error (unknown verb, flag or combination).
`cli` passes the remote command's own status back once it is running.

**Trust model.** `env.yaml` is executable configuration: its `command:` runs
with your privileges. Review a file someone hands you before `deploy` or
`remove`. `generate` is the dry run.

## 2. validate

```sh
solmq-conn-util validate -e env.yaml
```

Runs every check across `env.yaml`, every platform section present, and the
workflows, and prints every finding in one run. Non-zero exit on any error.

- Reports every error `generate` and `deploy` would stop at, including the TLS
  store checks: it opens each existing PKCS12 with its password, builds each
  created store in memory, holds every store and the kubernetes ConfigMap to
  the platform's size limit, and checks docker bind-mount sources.
- An unset `-env` variable is a warning, so it runs without production secrets
  exported. A created store is then built with a stand-in password. A variable
  set but empty is an error.
- Logins, the image and the cluster itself are left to `deploy`.

## 3. generate

```sh
solmq-conn-util generate config -e env.yaml                 # application.yml on stdout
solmq-conn-util generate config -e env.yaml -o application.yml
solmq-conn-util generate --platform kubernetes -e env.yaml  # the manifest set
```

- `generate config` reads the workflows and the connector defaults and prints
  the `application.yml` that `deploy` ships on every platform. It never opens a
  store.
- `generate` without `config` renders the resolved platform's artifacts,
  resolving credentials and building created stores where the artifact embeds
  them (kubernetes). See [section 11](#11-what-gets-generated).
- It fails fast: the first error stops it and nothing is written. Output is
  buffered, so `-o` never leaves a half-written file. With `-o` the file is
  written mode `0600`.
- Point `-o` outside the workflow folder, or the output is scanned as a
  workflow next time.

## 4. deploy

```sh
solmq-conn-util deploy --platform kubernetes -e env.yaml [--allow-command name]
```

Generates for the resolved platform, then applies it through the section's
`command:`:

| Platform | Runs |
|---|---|
| kubernetes | reads every Secret you manage that the pod mounts and stops if one lacks a key; then `<command> apply -f -` with the manifest on stdin |
| docker | `<command> compose -f - --project-directory <env.yaml's folder> up -d`, with the compose file on stdin and credentials in the child environment |
| podman | loads secrets, writes the unit, `systemctl [--user] daemon-reload`, then `start` or `restart` |

Before anything is written or applied, a read-only preflight runs: `info` on
docker and podman, and on kubernetes `auth can-i` for the one permission the
verb needs:

| Verb | Asks |
|---|---|
| `deploy` | `create deployment` |
| `remove` | `delete deployment` |
| `start`, `stop`, `restart` | `patch deployment` |
| `status` | `get pods` |
| `logs` | `get pods --subresource log` |
| `cli` | `create pods --subresource exec` |

A failing probe stops the run, naming the missing permission and namespace, or
giving a login hint. There is no skip flag.

`command:` must start with an allowlisted bare binary (`kubectl`/`oc`,
`docker`, `podman`). `--allow-command <name>` approves another one (`sudo`, a
wrapper) for this run only. The command runs as an argv slice, never through a
shell, and every token is checked against a safe charset.

**What a redeploy restarts.** Podman restarts the unit on every deploy. Docker
recreates the container when the compose file or any selected store changed.
Kubernetes does not restart pods when only `application.yml`, the status script
or a store changed; follow such a deploy with `restart`.

## 5. remove

```sh
solmq-conn-util remove --platform kubernetes -e env.yaml [--no-prompt]
```

Tears down what `deploy` created, after asking:

```text
remove kubernetes: this tears down deployment solmq-connector in namespace prod -- continue? [y/N]
```

Anything but `y`/`yes` exits 0 having touched nothing. `--no-prompt` answers
yes to both questions and is required when stdin is not a terminal.

- kubernetes: pipes the manifest set, minus the Namespace, to `<command>
  delete -f - --ignore-not-found`, in reverse order. Secrets the tool owns are
  deleted by name; nothing is resolved. Then it lists what is left in the
  namespace. A namespace holding anything the release does not own is kept and
  the occupants listed, on every path. An empty one is offered for removal as a
  separate question. The cluster namespaces (`default`, `kube-system`, `kube-public`,
  `kube-node-lease`) are never removed.
- docker: `compose down` against the configured project. Store secrets are
  handed to compose as placeholders, so no store source is needed.
- podman: stops the unit, removes it, reloads systemd, then removes every
  secret the unit or the spec mounted.

## 6. start, stop, restart

```sh
solmq-conn-util restart --platform kubernetes -e env.yaml
```

| Platform | `stop` | `start` | `restart` |
|---|---|---|---|
| kubernetes | scale the Deployment to 0 | scale back to `replicas` | `kubectl rollout restart` |
| docker | `docker stop` | `docker start` | `docker restart` |
| podman | `systemctl [--user] stop` | `start` | `restart` |

They read only the section's names and `command:`. Nothing is rendered, no
credential resolved, no store opened. `restart` is how the connector picks up a
rotated store or a changed `application.yml` where `deploy` does not restart it.

## 7. status

```sh
solmq-conn-util status container    # what the engine knows
solmq-conn-util status application  # what the connector knows about itself
solmq-conn-util status all          # both, container first
```

| Word | Short | Answers | Read from |
|---|---|---|---|
| `container` | `cnt` | Is it up? How often has it died? Which image is running? | `kubectl get` / `docker inspect` / `podman inspect` |
| `application` | `app` | Is it active? Healthy? Are the workflows running? | the status script, run inside each instance |
| `all` | | both | |

The target word is required. Instances are every pod matching the Deployment's
`app=<name>` selector, or the configured container name, narrowed with
`--pod` / `--container` (repeatable here) or widened with `--all`.

### 7.1 Container view

```text
== kubernetes  prod / solmq-connector ==
NAME                              STATE                          READY  RESTARTS  AGE   IMAGE
solmq-connector-7d9f8c6b5-k8m1r   restarting (CrashLoopBackOff)  no     7         3d7h  solace/solace-pubsub-connector-ibmmq:2.13.0
solmq-connector-7d9f8c6b5-x2n4q   running                        yes    0         3d7h  solace/solace-pubsub-connector-ibmmq:2.13.0

deployment: solmq-connector  1/2 ready, 2 up-to-date, 1 available
service:    solmq-connector  8090/TCP
```

| Column | Meaning |
|---|---|
| `STATE` | `running`, `exited`, `waiting`, `restarting`, `paused` or `unknown`, with the engine's reason (`CrashLoopBackOff`, `OOMKilled`) or exit code |
| `READY` (kubernetes) | the readiness probe, which runs the status script's `--health` mode, so `yes` means the connector reported `UP`. `n/a` when the pod has no probe |
| `HEALTH` (docker/podman) | the engine's healthcheck verdict: `healthy`, `unhealthy`, `starting`. `n/a` when the container declares none |
| `RESTARTS` | restart count. On podman it comes from systemd, since a quadlet restart recreates the container |
| `AGE`, `IMAGE` | how long it has run, and the image reference it is running |

On kubernetes the Deployment's replica counts and the Service follow the table.
A read that fails becomes a `status:` note, not a failed run.

### 7.2 Application view

```text
=== kubernetes  prod / solmq-connector / solmq-connector-7d9f8c6b5-x2n4q ===
  leader-election mode:  active_standby
  leader-election state: active
  health:                UP
  workflows:
     0: running
     9: running
    10: stopped
```

One block per instance, under a banner naming it (namespace / deployment / pod;
compose project / container; or the container alone).

| Line | From |
|---|---|
| `leader-election mode` | `env.yaml` |
| `leader-election state` | `/actuator/leaderelection`: `active` or `standby` |
| `health` | `/actuator/health`: `UP`, or the status plus a `health-detail` line naming the failing component |
| `workflows` | `/actuator/workflows`: one `<id>: <state>` row per configured workflow, ordered numerically |

- A standby carries no workflow lines. That is the normal shape.
- Unconfigured slots (`N/A`) are filtered out. An active instance with every
  slot empty gets a note.
- An instance that cannot be reached still gets a block, with the failure as a
  `status:` line.
- On docker, a compose project that differs from `project-name` in `env.yaml`
  gets a note, since `remove` would then miss the stack.

### 7.3 Installing the status script

Every artifact the tool deploys carries the script. An instance deployed some
other way may not. The first application view that finds it missing asks once:

```text
status script missing on solmq-connector-7d9f8c6b5-x2n4q -- install it now? [y/N]
```

`--install` skips the prompt (required on a non-TTY). Declining skips those
instances and exits 1. `--install`, `--user` and `--management-port` apply to
the `application` and `all` views only; with `container` they are a usage error.

### 7.4 Flags

| Flag | Meaning |
|---|---|
| `-d`, `--details` | add the enrichment lines below. Costs one sampling call (`kubectl top`, `docker stats`); kubernetes needs a metrics API |
| `--all` | report every container whose image name contains `solace-pubsub-connector-ibmmq`: every namespace on kubernetes (a `NAMESPACE` column appears when they span several), every container on docker/podman, stopped ones included. Needs no `env.yaml` but still a `--platform`. Not with `--pod`/`--container` |
| `-w`, `--watch` | re-render every 5s until Ctrl-C (exit 0). The install prompt is asked at most once |
| `--output json` | one document on stdout, notes on stderr. Fields are a compatibility contract under `schemaVersion`; empty fields are omitted. Not with `--watch` |
| `-v`, `--verbose` | one `step:` line per engine call on stderr, with its elapsed time, instead of the spinner. Printed whether or not stderr is a terminal; nothing under `--watch` |
| `--pod`, `--container` | limit to these instances, by name or by index into the listed (alphabetical) order; repeatable |
| `--namespace` | kubernetes namespace (default: the deployment's) |
| `--management-port`, `--user` | the actuator port and account the script uses (defaults: the configured port, `solmq-status`) |
| `--command`, `--allow-command` | override the platform CLI, and approve a binary outside the allowlist |

`--details` adds, per instance:

- Container side: the worker node, image digest, start time, CPU and memory
  against allocation, and the objects the pod or container references
  (ConfigMap, Secrets, PVCs) with their presence. `image-expected` is printed
  only when the running image differs from `env.yaml`.
- Application side: uptime, version, java, the config path the script read,
  heap against `-Xmx`, and per-component health. A line whose source answers
  nothing is dropped.

```json
{
  "schemaVersion": 1,
  "platform": "kubernetes",
  "namespace": "prod",
  "workload": { "deployment": "solmq-connector", "ready": 1, "desired": 2, "upToDate": 2, "available": 1 },
  "instances": [
    {
      "name": "solmq-connector-7d9f8c6b5-x2n4q",
      "container": { "state": "running", "ready": "yes", "restarts": 0, "age": "3d7h",
                     "image": "solace/solace-pubsub-connector-ibmmq:2.13.0" },
      "application": { "leaderElectionMode": "active_standby", "leaderElectionState": "active",
                       "health": "UP", "workflows": [ { "id": "0", "state": "running" } ] }
    }
  ]
}
```

By default a spinner on stderr names the step in flight (`/ status script
<pod> 1/2  4s`) and is erased before the report. It appears only when stderr is
a terminal. Progress never changes stdout or the exit code, so
`status --output json > report.json` is the same bytes either way.

### 7.5 Exit code and cost

- **0**: every instance was reported. A standby, an unhealthy container, a
  crash loop or a missing metrics API are answers, not failures.
- **1**: an instance's script could not be probed, installed or run, or the
  run could not start (no instances, unreadable `env.yaml`, failed preflight).
- **2**: usage error.

| View | kubernetes | docker | podman |
|---|---|---|---|
| `container` | preflight + 1 `get pods` (+ `get deployment`/`get service`) | preflight + 1 `inspect` | preflight + 1 `inspect` + 1 `systemctl show` per instance |
| `application` | + 1 probe and 1 script run per instance | same | same |
| `-d` adds | 1 `top pod` + 1 `get` per referenced object | 1 `stats` + 1 `image inspect` per image | as docker |
| `--all` adds | nothing | 1 `ps` | 1 `ps` |

### 7.6 Instances this tool did not deploy

The container view needs nothing. The application view needs the script
(install it with `--install`) and a reachable actuator account:

```sh
solmq-conn-util status all --platform kubernetes --pod other-team-connector-0 \
  --namespace other-ns --management-port 9090 --user ops-readonly
```

The script finds the connector's config the way Spring does and reads the
account password from it; with no config it queries unauthenticated. A pod with
several containers and none named `connector` is reported at pod level only.

To run the script by hand:

```sh
kubectl exec <pod> -- sh /app/external/.status-script
docker exec <container> sh /app/external/.status-script
```

It always exits 0 and prints problems on stderr, so branch on its output
(`| grep -q 'state: active'`), not on `$?`. Run with `--health` it makes one
health call and exits 0 only on `UP`; that is what the probes and healthchecks
run.

## 8. logs

```sh
solmq-conn-util logs                                   # one instance, or a list to choose from
solmq-conn-util logs --tail 100
solmq-conn-util logs --pod 0                           # the first instance in listed order
solmq-conn-util logs --previous                        # what the last container printed before it died
solmq-conn-util logs --follow --pod solmq-connector-7d9f8c6b5-x2n4q
```

Prints what one connector instance wrote. Platform and instances resolve
exactly as for `status`. `--all` is not a `logs` flag.

| Flag | Meaning |
|---|---|
| `--follow` | keep the log open until Ctrl-C (exit 0). Not with `--previous` |
| `--previous` | the previous container's log; kubernetes only, refused elsewhere |
| `--tail N` | the last N lines; default `all`; `0` is a real request |
| `--since d` | a Go duration (`30s`, `10m`, `2h`), not a date |
| `--timestamps` | prefix each line with the platform's time |
| `--pod`, `--container` | one value, by name or by index. A second value is refused |
| `--namespace`, `--command`, `--allow-command` | as `status` |

When several instances match and none is named, nothing is read. The matches
are listed as commands to paste back, with the flags already typed, and the run
exits 0:

```text
$ solmq-conn-util logs --tail 50
3 pods match; logs reads one. Run one of:

  solmq-conn-util logs --platform kubernetes --tail 50 --pod solmq-connector-7d9f8c6b5-abc12
  ...
or by index, 0-2 in the order listed:

  solmq-conn-util logs --platform kubernetes --tail 50 --pod 0
```

A `logs > app.log` against several replicas therefore puts this listing in the
file. Log lines go to stdout; the platform CLI's own diagnostics go to stderr.

By hand:

```sh
kubectl logs <pod> -n <namespace> -c connector --tail 100
docker logs --tail 100 <container>
podman logs --tail 100 <container>     # quadlet units are ordinary containers
```

## 9. cli

```sh
solmq-conn-util cli -e env.yaml                              # an interactive sh
solmq-conn-util cli -e env.yaml -- ls -la /app/external/libs # one command, then exit
echo hi | solmq-conn-util cli -e env.yaml -- cat             # stdin is attached when piped
```

Opens a shell in one instance, for the questions `status` and `logs` cannot
answer: did the truststore mount, what is in `/app/external/libs`, which
`application.yml` is loaded. Same resolution, allowlist, preflight and picker
as `logs`. The picker keeps a one-shot command behind its `--`.

- The shell is `sh`: the image is Alpine with busybox, and there is no `bash`.
- Everything after `--` reaches the engine as argv, never `sh -c`. A pipe,
  redirect or glob is refused; use an interactive session for those.
- An interactive run needs a terminal on stdin. Without one, `cli` refuses and
  names the one-shot form.
- Exit status: 2 for usage, 1 for an unreachable engine or a refused token,
  then the remote command's own status. The engines overload that status, so
  a non-zero exit means either the command failed or the session could not be
  established; the engine's stderr says which.
- On kubernetes the container is always `-c connector`. A foreign pod whose
  container is named otherwise is not reachable here; use `kubectl exec`.

| Path inside the container | What is there |
|---|---|
| `/app/external/spring/config/application.yml` | the configuration the process loaded |
| `/app/external/libs` | the jars `download jar` put there |
| `/app/external/classpath/truststores` | each selected store, under its store name |
| `/app/external/var/secrets` | the credentials |

## 10. download jar

```sh
solmq-conn-util download jar mq     [dir] [-e env.yaml] [--version v] [--omit-lib-file f] [--include-provided] [--url u] [-f]
solmq-conn-util download jar syslog [dir] [-e env.yaml] [--version v] [--omit-lib-file f] [--include-provided] [--url u] [-f]
```

Fetches jars over HTTPS from Maven Central into `<dir>` (default `./libs`), for
`libs.dir` (docker, podman) or `libs.pvc` / `libs.download` (kubernetes). It
reads `env.yaml` for the `image` block only, and a missing file is not an error.

| Set | Seed | For |
|---|---|---|
| `mq` | `com.ibm.mq:com.ibm.mq.jakarta.client` (JMS 3.0) and its dependencies | the IBM MQ connection |
| `syslog` | `net.logstash.logback:logstash-logback-encoder` and its dependencies | `logging.syslog.protocol: tcp` |

**Resolution.**

1. The seed resolves to `--version`, or to its latest stable release. For
   `syslog` without `--version`, it is the newest 8.x for a connector 2.x
   image and the newest release otherwise; the `seed:` line says which.
2. Every dependency is pinned to the version the seed's own POM declares. One
   that cannot be resolved falls back to its latest stable and is reported
   under `fallback`.
3. Each dependency is compared with the image's jar list and **omitted** when
   the image ships it at an equal or newer version (Maven ordering: `Final`
   equals the release, `SP` is above it, `rc`/`alpha`/`SNAPSHOT` below). The
   seed is never omitted. Every omission is reported.
4. Each jar's `.sha1` sidecar is fetched and checked before the file lands. A
   mismatch or a missing sidecar on a Maven-resolved jar fails that artifact.

With no `--version` the exact files change over time. Pin one in scripts.

**The image jar list.** Built-in lists are captured from connector images and
each describes a range: 2.13.0 (2.10.0 and later, before 3.0.0), 3.1.0 (3.1.0
before 3.2.0), 3.2.0 (3.2.0 and later). The command picks the list whose range
holds the `image.tag` you deploy and prints it (`omit list: ...`). When none
does (a 3.0.x, `latest`, a digest, another image) it warns that omissions are
judged against the nearest list. With no `env.yaml` the newest list is used.

For a custom image, capture its list and pass it:

```sh
docker run --rm --entrypoint sh <your-image> -c 'find / -name "*.jar" -not -path "/proc/*" 2>/dev/null | xargs -n1 basename | sort' > omit-lib-file.txt
solmq-conn-util download jar mq --omit-lib-file omit-lib-file.txt
```

| Flag | Meaning |
|---|---|
| `[dir]` | destination, created if missing; default `./libs` |
| `--version v` | pin the seed release |
| `--omit-lib-file f` | a jar list that replaces the built-in one entirely; an empty file omits nothing |
| `--include-provided` | skip the omission check and download the whole closure |
| `--url u` | repeatable; download exactly these URLs, with no resolution and no omission. A `.sha1` sidecar that 404s leaves the jar written but reported `unverified` |
| `-f`, `--force` | overwrite an existing file; never reaches a jar the omission check dropped |

https is mandatory on every hop. sha1 proves integrity, not authenticity; for a
regulated environment vendor the jars yourself or use `--url` against a mirror
you verified.

Notes:

- The `mq` seed is always the Jakarta client. `com.ibm.mq.allclient`
  (`javax.jms`) cannot serve the connector, and both on one classpath clash.
- The image already ships an encoder (2.x: 8.0, 3.x: 9.0), so `download jar
  syslog` leaves two copies across the classpath. Verify tcp syslog loads in
  your deployment, or pin with `--version` or `--url`.
- The `mq`/`syslog` sets contact Maven Central even when every file exists;
  `--url` checks for an existing file first.

## 11. What gets generated

For the meaning of the connector's own keys, see the connector documentation.
This is what the tool decides for you.

**`generate config`**: one multi-binder `application.yml` with deduplicated
binders, numbered workflows, the mandatory `undefined` binder (always last),
derived destination types, auto durable subscription names, the TLS wiring,
`spring.main.sources` on every MQ binder, the configtree import, the
`solmq-status` account, and the fixed management exposure. Stores are referenced
at `/app/external/classpath/truststores/<name>` and credentials as
`${NAME}`, so the document runs only where a platform mounts them.

**Every platform** also ships the status script at
`/app/external/.status-script` and labels the instance with
`solace-connector/le-mode`. `solace-connector/role: active` is added for
`standalone` and `active_active`, where the artifact already knows; under
`active_standby` the role is left to run time. When an MQ binder uses TLS,
`JAVA_TOOL_OPTIONS=-Dcom.ibm.mq.cfg.useIBMCipherMappings=false` is set. With
`logging.syslog`, a `logback-spring.xml` and `LOGGING_SYSLOG_*` variables are
added.

| Platform | Artifacts |
|---|---|
| kubernetes | a Namespace; a ConfigMap (`application.yml` and the status script, `subPath` mounts); a Deployment with a `tcpSocket` liveness probe and a readiness probe running `--health`; an optional Service; the `<name>-credentials` and `<name>-stores` Secrets; any libs PV/PVC |
| docker | a `docker-compose.yml`: `application.yml` and the script inlined as `configs:` (`$` doubled), credentials as environment-provider `secrets:`, bind mounts for stores and libs, a healthcheck, and the stores digest label |
| podman | a `.container` quadlet unit mounting everything from podman's secret store, which `deploy` fills |

## 12. Shell completion

```sh
solmq-conn-util auto-complete bash|zsh|fish|powershell
```

Prints a completion script for verbs, targets, flags, a path after `-e`/`-o`
and a directory after `examples`. It is rendered from the command model in the
binary, so re-run it after upgrading. Install per shell:

```sh
# bash: add to ~/.bashrc
source <(solmq-conn-util auto-complete bash)

# fish: writing the file is the install
mkdir -p ~/.config/fish/completions
solmq-conn-util auto-complete fish > ~/.config/fish/completions/solmq-conn-util.fish

# zsh: the file must be named _solmq-conn-util and sit on $fpath before compinit
mkdir -p ~/.zsh/completions
solmq-conn-util auto-complete zsh > ~/.zsh/completions/_solmq-conn-util
#   in ~/.zshrc, ahead of compinit:  fpath=(~/.zsh/completions $fpath)
```

```powershell
solmq-conn-util auto-complete powershell >> $PROFILE
```

| Shell | Gotcha |
|---|---|
| bash | `/etc/bash_completion.d/` is read only where the bash-completion package is installed; the `source` line needs nothing |
| zsh | `compinit` caches: `rm -f ~/.zcompdump* && exec zsh` |
| fish | `solmq-conn-util auto-complete fish \| source` covers the current session |
| PowerShell | `\| Out-String \| Invoke-Expression` is the current session only |

The scripts complete the command named `solmq-conn-util`. Rename a downloaded
release binary (`solmq-conn-util-linux-amd64`) to that name, or completion will
not fire.
