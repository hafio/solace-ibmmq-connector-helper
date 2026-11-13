# Decisions on the 2026-11-11 critical review

Owner decisions on every finding in
[2026-11-11-critical-review.md](2026-11-11-critical-review.md), taken 2026-11-13. Nothing here
is built yet. Evidence and suggested diffs stay in the report.

Home says where the work lands:

- `here`: this repository.
- `gdu N`: item N of `docs/todo.md` in go-deploy-utility, for the part both tools share.
- `init-devops`: the skill's templates or skeletons first, then a re-deploy here.

## Decisions

| ID | Severity | Decision | Home |
|----|----------|----------|------|
| CR-0001 | minor | `remove` on docker and podman takes the names-only path that `start`, `stop` and `restart` use. | here |
| CR-0002 | minor | The three pipelines follow one order: validate and resolve, preflight, act. | here, gdu 3 |
| CR-0003 | nit | Apply the report's patch. | here |
| CR-0004 | nit | `status` takes its clock from the app value. | here |
| CR-0005 | blocker | validate holds every passthrough key to a grammar at every depth, the way solace-util holds values to the sink's grammar. render also quotes a key that is not plain. | here |
| CR-0006 | major | Retire `${VAR}` expansion. Expand only a leading `~`. Pass a `${VAR}` on as written. | here, gdu 2 |
| CR-0007 | major | Accepted here: namespace removal stays as it is. Ownership labels on every artifact are planned in the shared module. | gdu 1 |
| CR-0008 | major | Docker `remove` resolves no credential. The names-only path of CR-0001 covers it. | here |
| CR-0009 | minor | validate holds `cipher` to a cipher-suite grammar and refuses control characters in `key-alias`. Both are quoted when written. | here |
| CR-0010 | minor | An aliased passthrough block is dereferenced. A non-mapping value is a validate error. | here |
| CR-0011 | minor | `solace-defaults` gets the checks and the filter `mq-defaults` has. | here |
| CR-0012 | minor | Accepted. | - |
| CR-0013 | minor | The tool calls the shared name grammars at its boundary for `status`, `logs`, `cli` and the lifecycle verbs. The private regex goes (report patch). | here, gdu 4 |
| CR-0014 | minor | The manifest quotes names, namespaces and label values. | here |
| CR-0015 | minor | `replicas` must be 1 or more, and an explicit 0 is an error. `management.port` must be 1-65535. | here |
| CR-0016 | minor | Apply the report's patch. | here |
| CR-0017 | minor | Apply the report's one-line patch, with a test. | here |
| CR-0018 | minor | A second YAML document in env.yaml or a workflow file is an error that names the file. | here |
| CR-0019 | minor | Accepted. | - |
| CR-0020 | nit | After `--` every argument is positional. | here |
| CR-0021 | nit | The sub-id check runs first. Linger is enabled only after the question is answered yes. | here |
| CR-0022 | major | Zero workflows is a validate error naming `workflows.dir`, `file_pattern` and the folder. | here |
| CR-0023 | major | A POM fetch is retried a small, capped number of times. An unresolved branch fails the run and is listed in the report. | here |
| CR-0024 | minor | The unit is written through a temporary file and a rename. Stale secrets are removed after the reload, and a failure there is a warning. | here, gdu 7 |
| CR-0025 | minor | Preflight and read probes get a bounded context. Acting calls get a signal-derived context and no fixed deadline. | here |
| CR-0026 | minor | The app value carries stdout and stderr writers, threaded through every print site. | here |
| CR-0027 | nit | One `isNull`, one `seconds` and one `lineDiff` in this repository. | here, gdu 5 |
| CR-0028 | nit | Apply the report's patch. | here |
| CR-0029 | nit | Both values move into the libs `Input`. | here |
| CR-0030 | major | Remove the generator page. A new page is planned below. | here |
| CR-0031 | minor | The `docs/test.md` gate compares test names. | here, gdu 8 |
| CR-0032 | minor | Split `main_test.go` by source file, with shared fakes in one helper file. | here |
| CR-0033 | nit | The three tool-dependent tests fail instead of skipping when `CI` is set. | here |
| CR-0034 | minor | Re-deploy `tag.yml` from the current template. | init-devops |
| CR-0035 | minor | CI runs `all cov scan`. DEVELOPMENT.md follows. | init-devops |
| CR-0036 | minor | Both dev scripts come up to the contract: race detector, cov-queued skip, toolchain gofmt. | init-devops |
| CR-0037 | minor | DEVELOPMENT.md gains the `GOWORK=off` release step. Each task log says whether a workspace was in effect. `full` sets `GOWORK=off`. | here |
| CR-0038 | minor | Dependabot gains a weekly `gomod` block. `go.mod` and `go.sum` join the pull-request paths in `ci.yml`. | init-devops |
| CR-0039 | minor | The tag pipeline gains a step that compares the binary's `version` with the tag. The stamp still comes from `git describe`. | init-devops |
| CR-0040 | nit | `host_os` maps `mingw*`, `msys*` and `cygwin*` to `windows`. | init-devops |
| CR-0041 | major | `libs.download` keeps wget and http. Each entry may carry an optional `sha256`, checked in the init container when given. | here |
| CR-0042 | minor | New config keys set the network mode and the root filesystem. They default to bridge and read-only, and every artifact states both. | here, gdu 6 |
| CR-0043 | minor | No change here until the shared quadlet builder escapes `%`. | gdu 4 |
| CR-0044 | minor | DESIGN.md records that the status account's credential is on wget's argv, and what that account can reach. | here |
| CR-0045 | major | The picker goes to stderr with exit 2 when scripted: a one-shot `cli --`, and `logs` when stdout is not a terminal. | here |
| CR-0046 | major | An unknown key in an env.yaml section or at a workflow's top level is a validate error with a near-match hint. | here |
| CR-0047 | minor | A question that cannot be asked keeps exit 1 and operations.md names it as the exception. A missing store file and a bad argument exit 2 on every platform. | here |
| CR-0048 | minor | One "acting on" line before the first engine call names the instance, namespace, env.yaml path and image. | here |
| CR-0049 | minor | The quick start and `examples` both say the stores and their password variables come first. | here |
| CR-0050 | minor | Each "user guide" pointer becomes the configuration.md or operations.md section it means. | here |
| CR-0051 | minor | Apply both report patches and correct the two comments in gen.go. | here |
| CR-0052 | minor | Old finding ids, stage labels and rule numbers become test names or plain words. | here |
| CR-0053 | minor | The two validate messages and the four starter workflows with their golden copies become plain ASCII. | here |
| CR-0054 | minor | The notice check covers README.md and every `docs/*.md` through the shared `AssertDocs`. `docs/review/` is exempt and `docs/test.md` gains the notice. | here |
| CR-0055 | minor | Apply the report's patch. | here |
| CR-0056 | minor | CLAUDE.md asks the user to run `./scripts/dev.sh graphify`. The same wording goes into the graphify wiring. | here, init-devops |
| CR-0057 | minor | The file-read sites wrap with `%w`. A parse error flattened on purpose says so in a comment. | here |
| CR-0058 | minor | Untrack `graphify-out/cache` and the settings backup. The owner runs the command. | here |
| CR-0059 | nit | DEVELOPMENT.md names go-deploy-utility among the dependencies. | here |
| CR-0060 | nit | Accepted. | - |
| CR-0061 | nit | The module path in go.mod changes to match the repository. | here |
| CR-0062 | suspected | `restart` scales to the configured replicas when the deployment has none. | here |
| CR-0063 | suspected | The usage parser reads five columns under `--all-namespaces` and keys by namespace and pod. | here |
| CR-0064 | suspected | Every string the kubernetes status path prints passes through `inspect.Printable`. | here |
| CR-0065 | confirmed | The download file name drops the query and fragment. validate reports an empty name, a name without `.jar` and a duplicate name. | here |
| CR-0066 | suspected | Podman `remove` skips the stop when the unit is absent and not loaded, removes what is left and exits 0. | here, gdu 9 |

CR-0065 was confirmed on 2026-11-13 by reading `spec.BaseName`, which keeps everything after
the last `/`. The ledger still lists it as suspected until the next review run.

## Planned work the report does not hold

### A new generator page

The current page goes with CR-0030, together with its golden tests, its support-notice check
and its section in DEVELOPMENT.md. A replacement is planned and not designed. Its one fixed
requirement comes from the finding: a gate must run whatever logic the page carries, so the
page cannot drift from the CLI unnoticed.

Undecided:

- Whether the page ports the Go logic again or runs the Go code itself, for example compiled
  to WebAssembly.
- What the page does: edit a spec, preview the rendered files, or both.
- Where it ships: in the release, in the repository or both.

### Shared module

`docs/todo.md` in go-deploy-utility holds nine items raised by these decisions:

1. Ownership labels on every artifact, and a wider namespace listing.
2. One guideline for expansion in config values.
3. One step order for deploy and remove on every platform.
4. Builders refuse a name outside its grammar, and quadlet lines escape `%`.
5. One diff helper for golden tests.
6. Every artifact states its network mode.
7. One atomic write for files a tool installs.
8. One gate for the test catalogue.
9. Remove is idempotent on every platform.

### Stale objects after a config change

Raised 2026-11-13 from solace-util. Its owner decided that stale secrets and objects are
removed by ownership labels on podman and on Kubernetes, with both tools aligned. The labels
and the removal are one standard behaviour in go-deploy-utility that both tools import, and
the design is worked through there. Items 1, 16 and 18 of its `docs/todo.md` hold it. For
this tool:

- Podman: to do, so this tool matches solace-util. The stale-secret removal moves from the
  old-unit read to the ownership labels of gdu 1 (gdu 16, decided 2026-11-13):
  - Every secret is created with the labels, through `podman.CreateSecret`. They name the
    tool and the instance, and nothing about a value.
  - A deploy and a remove list the secret names, read `Spec.Labels` with `secret inspect`,
    and delete the instance's labelled secrets that env.yaml no longer names. An unlabelled
    secret is never deleted.
  - A secret env.yaml still names and that carries no label is created again once with the
    same value, so it gains the labels.
  - `previousStoreSecrets` and `gen.PodmanUnitStoreSecrets` are dropped in the same step.
    The labels are the one source (decided 2026-11-13).
  - CR-0024 stands: the removal runs after the unit is written and reloaded.
  - Questions come first and actions after (decided 2026-11-13). The write that adds the
    labels and every removal follow the redeploy question, as CR-0021 already does for
    linger. The same holds on Kubernetes. A step that needs no question just runs: a deploy
    whose only change is the write that adds the labels asks nothing, and removing a stale
    secret or object asks nothing either. The tool names what it removed.
- Kubernetes: a deploy and a remove also delete the objects an earlier deploy created and
  env.yaml no longer names. Today the image-pull Secret, the libs PV and PVC and the Service
  are rendered only while env.yaml asks for them, for a deploy and for a teardown alike, so a
  dropped one stays and a Service that was switched off keeps routing
  (`internal/deploy/deploy.go`). The removal goes by the ownership labels of gdu 1: every
  object the tool creates carries them, the instance's labelled objects are listed, and the
  ones env.yaml no longer names are deleted. An unlabelled object is never deleted, which
  protects a Secret the operator made (gdu 18, decided 2026-11-13).
- Docker: nothing to do. A compose secret is copied into the container and nothing is stored
  outside it (decided 2026-11-13).

Undecided:

- Which kinds the removal covers. A PVC holds data and may be left out.
- What a deploy does when the listing is refused.

## Consequences to check before the work starts

- CR-0006 breaks every env.yaml that relies on `${VAR}` today. A `${VAR}` passed on as
  written reaches application.yml, where the connector's own property resolution reads it
  from the container's environment. Compose output doubles every `$`.
- CR-0030 also removes the two page references counted under CR-0050.
- CR-0041 and CR-0065 change a `libs.download` entry from a plain URL to a URL with an
  optional `sha256`. The plain form stays valid.
- CR-0042 needs a real deploy on each platform. A read-only root must be shown not to stop
  the connector, and the paths the JVM writes need a tmpfs.
- CR-0046 and CR-0018 turn files that load today into validate errors.
- CR-0061 rewrites every import path in the repository. It is its own change.

## Steps the owner runs

- CR-0058: `git rm -r --cached graphify-out/cache .claude/settings.json.graphify-bak`, then
  commit.
- CR-0034 to CR-0036 and CR-0038 to CR-0040: the init-devops skill, once its templates carry
  the changes.
- In go-deploy-utility, the docs test after `docs/todo.md` was added: `./scripts/dev.sh test`.
