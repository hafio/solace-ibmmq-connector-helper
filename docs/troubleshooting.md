# solmq-conn-util troubleshooting

> [!WARNING]
> **Not a supported Solace product.** `solmq-conn-util` was created by Solace
> Professional Services and is supported only by Solace Professional Services --
> not by Solace Support. For help with this tool, contact your Solace
> Professional Services representative rather than opening a Solace Support
> case. This notice covers this tool only, not the Solace PubSub+ Connector for
> IBM MQ that it configures and deploys.

Things that surprise people, grouped by where they bite. The keys are explained
in [configuration.md](configuration.md), the verbs in
[operations.md](operations.md).

## Contents

1. [Workflows and numbering](#workflows-and-numbering)
2. [Passthrough keys](#passthrough-keys)
3. [TLS](#tls)
4. [Secrets and credentials](#secrets-and-credentials)
5. [Platforms](#platforms)
6. [Status, logs and cli](#status-logs-and-cli)
7. [Upgrading from an earlier release](#upgrading-from-an-earlier-release)

## Workflows and numbering

- **`workflow-10.yaml` is workflow 2, not 10.** Files sort in byte order, so
  `10` comes before `2`. Zero-pad past 9. Renaming or adding a file renumbers
  everything after it, and the `input-<N>`/`output-<N>` bindings move too.
- **More than 20 workflows is an error.** The connector binds ids 0 to 19 and
  silently ignores the rest. Split the folder; each folder is one connector.
- **Renaming a workflow file orphans its MQ durable subscription.** The file
  name is part of the derived name.
- **A Solace `topic:` source is rejected.** The connector would start cleanly
  and never consume. Consume from a queue subscribed to the topic.
- **The `undefined` binder is always emitted.** Multi-binder syntax is always
  used; this is expected.

## Passthrough keys

- **Any other key under `solace:`/`mq:` reaches the connector, typos
  included.** The connector ignores what it does not read. `validate` warns on a
  key within a typo of one the tool reads; a wrong name with no near neighbour
  passes through unremarked.
- **An `-env` suffix means something only on credential pairs.**
  `client-id-env` under `mq:` is a plain property, passed as written.
- **A cipher clash across two sides of one binder is not detected.** `cipher:`
  on one side and `ssl-cipher-suite:` on another that share a binder both reach
  the connector. On one side it is an error.
- **Two sides on one connection must share a password.** Disagreeing passwords
  are an error, not a last-wins merge. A differing `cipher` warns and takes the
  last value.

## TLS

- **`tls: true` alone does not make an MQ connection use TLS.** The IBM MQ
  client negotiates TLS only with a cipher suite set. Without `cipher:` it
  connects in plaintext and a TLS-only channel refuses it. `validate` warns.
- **A hand-written `application.yml` and the tool's differ on MQ TLS.** Each
  MQ binder has its own `environment:` and so its own Spring context, which is
  why every MQ binder carries `spring.main.sources`. Moving `ibm.mq` into a
  binder `environment:` by hand without it silently drops `ssl-bundle` and the
  connection fails with `MQRC_JSSE_ERROR` (2397).
- **An `openssl pkcs12 -nokeys` truststore loads in Java as empty.** Java
  keeps only certificates marked as trust anchors, which OpenSSL writes only
  from 3.2 with `-jdktrust anyExtendedKeyUsage`. Let the tool build the store
  (`create: {certs: [...]}`). The error is "the trustAnchors parameter must be
  non-empty".
- **A created keystore takes no `key-alias`.** It holds one key, which Java
  names `1`. Use one created store per client identity.
- **TLS without a store selection trusts the JVM's CAs.** A `tcps://` host or
  `tls: true` with no `truststore:` validates against the default CA bundle. An
  MQ binder that selects nothing gets no SSL bundle at all.
- **The tool writes no certificate-validation policy.** `SSL_VALIDATE_CERTIFICATE`
  is whatever the connection's `api-properties` say, truststore or not, and
  every command warns when it is `false`.
- **A store is `existing` or `create`, always PKCS12.** A store with neither is
  an error. `file:` and `type:` are retired; convert a JKS with `create: {jks:
  <file>}` or `keytool -importkeystore`.
- **Stores mount under their names.** Every selected store is
  `/app/external/classpath/truststores/<name>` on every platform, whatever its
  file is called. Two stores may share a file, or be one file.
- **The connector reads stores at start only.** After rotating one, podman
  `deploy` restarts the unit and docker `deploy` recreates the container, but
  kubernetes needs `solmq-conn-util restart`.

## Secrets and credentials

- **The `solmq-status` password is a literal in `application.yml`.** Anyone
  who can read the ConfigMap, the compose file or the podman secret can read
  it. Never reuse a password you use elsewhere, and pin
  `SECURITY_USER_SOLMQ_STATUS_PASSWORD` if it must survive a redeploy.
- **`generate --platform kubernetes` embeds every Secret.** The manifest is
  written `0600` with `-o`; treat it as sensitive. Docker and podman write no
  credential to disk.
- **A kubernetes `credentials.existing` Secret needs the derived keys too.**
  `-env` credentials use your variable names; a literal needs its `_GEN_...`
  key. Run `generate --platform kubernetes` with `create: true` once and read
  the `stringData:` keys. A missing key fails quietly: the pod starts and the
  connector fails later to authenticate.
- **Two names that differ only in punctuation fold together.** Users `ops.1`
  and `ops-1` both mount as `_GEN_SECURITY_USER_OPS_1_PASSWORD`; the run
  refuses, naming both.
- **`${VAR}` inside a credential is an error.** Use the `-env` form.
- **Every `$` in a generated compose file is doubled.** Compose would otherwise
  substitute `${...}` placeholders from the environment and write plaintext
  credentials into the file. Keep the doubling when editing by hand.

## Platforms

- **The connector runs as `1000:0`, not root.** A host file that identity
  cannot read (a `0600` truststore on docker, a `0600` jar) fails the connector
  at run time, not at deploy. `run-as: image` does not help: root runs with
  every capability dropped. `deploy` warns on Linux about each unreadable
  mounted path.
- **Rootless podman needs the run-as ids inside your subuid/subgid ranges.**
  The default 1000 fits a standard 65536 range; a bigger id does not. Widen
  the range, pick a smaller id, deploy as root, or set `run-as: image`.
- **OpenShift assigns its own ids.** Set `kubernetes.run-as: image`.
- **The safe-charset gate covers more than `command:`.** `image`, `restart`,
  `timezone`, bind-mounted store files, `libs.dir`, referenced Secret names and
  `libs.pvc.create.nfs.*` are rejected when they carry whitespace, quotes,
  control characters or shell metacharacters.
- **`java-options` has its own gate.** Spaces separate options and shell
  metacharacters are fine, since no value reaches a shell. Quotes, backslash,
  backtick, a leftover `$`, control characters and a `#` comment inside a `>-`
  block are rejected.
- **`name` and `docker.project-name` are DNS-1123 labels**, stricter than
  compose's own rule: lowercase alphanumerics and hyphens, starting and ending
  alphanumeric.
- **Docker: `COMPOSE_PROJECT_NAME` in your shell wins** over `project-name`.
  Unset it.
- **Docker: a project-name mismatch means `remove` misses the stack.** `status`
  flags it. Set `project-name` to the reported value, or run `docker compose -p
  <name> down`.
- **Docker: two stacks need two `name` values.** Container names are
  host-unique regardless of project.
- **Docker: a hand-run `docker compose up` with a created store fails.** The
  store secret is filled only by `deploy`; the error names the unset
  `_GEN_TLS_<STORE>_STORE` variable.
- **Docker: a created store must fit one environment variable.** 131072 bytes
  on Linux, 32767 characters on Windows, base64. Trim the certificates or
  deploy that store with podman or kubernetes.
- **Podman: a store of 512000 bytes or more is refused.** Podman secrets hold
  up to 511999 bytes.
- **Podman: `restart: unless-stopped` is an error.** Use systemd's words;
  `always` means the same thing.
- **Kubernetes: `deploy` does not restart pods** when only `application.yml`,
  the status script or a store changed. Run `restart`.
- **Kubernetes: a store in `<name>-stores` is limited to 256 KiB.**
  Client-side `kubectl apply` cannot carry more. Keep a store in a Secret you
  manage (`existing.secret`) instead.
- **Kubernetes: PV capacity and NFS settings are immutable.** Raising `storage`
  and redeploying looks like it worked but leaves the claim asking for more
  than the volume has. `deploy` warns; delete the volume and deploy again.
- **Kubernetes: `remove` reverses the manifest order.** Deleting in creation
  order deadlocks on `pvc-protection` when a `libs.pvc.create` claim precedes
  the Deployment that mounts it. The libs PV and PVC go with the release; a
  `libs.pvc.existing` claim is never deleted.
- **Kubernetes: a namespace with anything else in it is never removed.** The
  occupants are listed and the namespace kept, with or without `--no-prompt`.
- **A piped run hangs? It does not; it fails.** Every prompt refuses a
  non-TTY and names the flag to pass: `--platform`, `--install`, `--no-prompt`,
  or the one-shot `cli` form.

## Status, logs and cli

- **`status` exit 0 does not mean healthy.** It means every instance was
  reported. Read the report, or `--output json`, for the state.
- **A standby shows no workflow lines.** That is the normal shape.
- **Podman restart counts come from systemd.** A quadlet restart recreates the
  container, so podman's own counter reads 0.
- **`HEALTH` reads `n/a`.** The container declares no healthcheck: created
  before this tool emitted one, deployed without the script, or run by hand.
- **A line you expect under `-d` is missing.** The installed script may be out
  of date. Reinstall it with `--install`.
- **The script's `heap:` line is raw bytes.** The CLI converts to `Mi`/`Gi`;
  busybox arithmetic cannot handle scientific notation.
- **`logs > app.log` wrote a list of commands.** Several instances matched.
  Name one with `--pod`/`--container`, by name or by index.
- **`cli` exited non-zero: command or connection?** The engines use the same
  code for both. The engine's stderr says which.
- **`cli` cannot reach a foreign pod** whose connector container is not named
  `connector`. Use `kubectl exec` directly.
- **`cli -- cat file | grep x` is refused.** The one-shot form is argv, not a
  shell line. Pipe outside, or use an interactive session.
- **`java:` under `-d` is dropped.** The image has no `java` on `PATH`.

## Upgrading from an earlier release

Each of these is reported by `validate`. `generate` and `deploy` keep working
where noted.

- **Unpadded workflow names past 9 renumber.** Earlier releases sorted
  numerically. `workflow-10.yaml` becomes workflow 2, and its bindings move.
  Durable names do not change.
- **Passthrough keys that were dropped now reach the connector.** Regenerating
  can add keys to `application.yml` that were silently ignored before, and
  `deploy` then rolls the pods. A camelCase spelling of a tool key that used to
  be ignored is now an error.
- **`tls.truststore` / `tls.keystore` are retired.** Define each store under
  `tls.stores.<name>` and select it on the connection with `truststore:` /
  `keystore:`. `key-alias` now needs a `keystore:`. Both old keys are errors
  naming the replacement.
- **A store's `file:` / `type:` are retired.** Use `existing: {pkcs12: <file>}`
  or `create: {jks: <file>}`.
- **`SSL_VALIDATE_CERTIFICATE: true` is no longer written.** The API's default
  applies unless the connection sets it.
- **Kubernetes: `secrets.stores` is retired.** With `create: true`, nothing
  changes but a `validate` error asking for the block to go. With `existing:
  <Secret>`, `generate` and `deploy` stop until each store names the Secret
  itself (`existing.secret: {name: <Secret>}`; the key defaults to the store's
  name). `remove` only warns. The first deploy rolls the pods once.
- **Kubernetes: object names are now derived from `deployment.name`.** The old
  `secrets.credentials.create.name`, `secrets.image-pull.name` with `create:
  true`, and `libs.pvc.create.name` are ignored by `generate`/`deploy` and
  reported by `validate`. The first deploy creates the new objects and moves
  the pods; the old ones stay behind and `remove` does not know them:

  ```sh
  kubectl -n <namespace> delete secret <old-credentials-name> <old-stores-name> <old-image-pull-name>
  kubectl -n <namespace> delete pvc <old-claim-name>
  kubectl delete pv <namespace>-<old-claim-name>-pv
  ```

- **Kubernetes: omitting `credentials` once a credential is set is now an
  error.** Earlier releases warned. Add `credentials: {create: true}` or
  `credentials: {existing: <Secret>}`.
- **Podman: files move into the secret store.** The first `deploy` loads
  `application.yml`, the status script, the logback config and the stores into
  podman's secret store and restarts the unit. Files earlier releases wrote
  under `podman.base-dir` (`<name>-application.yml`, `<name>-status`,
  `<name>-logback-spring.xml`) stay until you delete them; the first holds the
  status password in clear. Delete the files and the key.
- **The connector now runs as `1000:0`.** The first `deploy` rolls the pods,
  recreates the compose service and restarts the podman unit. Check that
  mounted store files and jars are readable by that identity.
