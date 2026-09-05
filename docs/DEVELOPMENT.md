# solmq-conn-util -- development guide

Building, testing, releasing, and design notes for `solmq-conn-util`. For using the
tool, see [userguide.md](userguide.md); for a quick start, [README.md](../README.md).

## Build

Requires Go 1.27+ (the `go` directive in `go.mod`). Dependency: `gopkg.in/yaml.v3` only.

```sh
# from the repo root
go mod tidy          # offline if the module cache already has yaml.v3
go build -o solmq-conn-util ./cmd/solmq-conn-util   # CGO_ENABLED=0 for a static binary
```

Or use the mirrored task runner (`dev.sh` / `dev.ps1`, behaviorally identical):

```sh
./scripts/dev.sh all          # build + vet + test              (fatal gates; CI runs `all scan`)
./scripts/dev.sh full         # all + cov + scan + graphify     (pre-tag sweep)
./scripts/dev.sh build        # one binary -> dist/ (host, or the TARGET_OS/TARGET_ARCH pair)
.\scripts\dev.ps1 full        # same on Windows PowerShell
```

Tasks: `build vet test cov scan regen graphify`, plus aggregates `all` (= build vet test, run by CI
as `all scan`) and `full` (= all + cov + scan + graphify, the pre-tag sweep). Gates
build/vet/test/scan are fatal.

`vet` is `go vet ./...` **plus a formatting check**: it fails when `gofmt -l .` lists anything,
naming the files and the `gofmt -w` line that fixes them. The check only lists -- it never
writes -- which is what lets it live inside an aggregate at all; a `gofmt -w` task would have to
stay out of every aggregate for the same reason `regen` does (below), since a gate that repairs
what it measures proves nothing. `go vet` and the format check both always run, so one pass
reports both rather than hiding vet errors behind a formatting failure.

`scan` is `go tool govulncheck ./...`, fatal on any finding -- every Go vuln-DB finding is
reachable-and-fixable, so there is nothing to warn-and-pass on.

The scanner is pinned in `go.mod` as a tool dependency rather than invoked as `go run ...@latest`:
that form ignores `go.mod` and builds the scanner on whatever toolchain its own module requires,
which then cannot load packages from a module on a newer `go` directive.

The Go toolchain itself is pinned exactly (`toolchain` directive in `go.mod`), so the laptop and
both CI runners download and run the identical Go rather than whatever the runner image
preinstalls -- bump the pin deliberately when upgrading locally.

`build` honors `TARGET_OS`/`TARGET_ARCH` and writes `dist/solmq-conn-util-<os>-<arch>[.exe]` (host
os/arch when unset), so one task serves the laptop and the CI matrix.

`regen` runs every `go:generate` directive, rewriting the generated docs and completion goldens;
like `graphify` it is local-only (warn-skips under CI) and belongs to no aggregate, because `test`
is what fails on drift and a regen inside a sweep would rewrite the evidence instead of reporting
it.

`image`/`up`/`down` are omitted -- the tool ships no Dockerfile or local stack (it generates
artifacts for other engines; it is not itself containerized).

## Testing

A worked golden example lives in [`testdata/golden/`](../testdata/golden) and is
asserted byte-for-byte by the tests (`internal/gen/golden_test.go`).

[`test.md`](test.md) is the full test catalogue -- every test, grouped by package and
expanded to individual cases. Keep it in sync: a test or case added, removed, or renamed
updates the matching row in the same change. `TestTestCatalogSnapshotInSync`
(`cmd/solmq-conn-util/testcatalog_test.go`) gates the doc's own "_Snapshot: N test
functions, M case rows across P packages._" line against the repo: it recounts `func Test`
declarations across every `*_test.go` file, the doc's own case rows, and its package
sections, and fails naming whichever count drifted. There is no `-update` flag for it --
the rows above the snapshot line are hand-written prose, not a rendered model, so a human
edits the line by hand after fixing the rows.

[`commands.md`](commands.md) is the **generated** command reference and
[`abbreviation.md`](abbreviation.md) the **generated** lookup of every short spelling, both
rendered from the command model in
[`cmd/solmq-conn-util/commands.go`](../cmd/solmq-conn-util/commands.go) (the abbreviation
renderer lives in [`abbreviation.go`](../cmd/solmq-conn-util/abbreviation.go)). Do not edit
either by hand; `TestCommandsDocInSync` and `TestAbbreviationDocInSync` fail the build if one
drifts from the model. Regenerate after changing a command:

```sh
./scripts/dev.sh regen              # = go generate ./...
go generate ./cmd/solmq-conn-util   # = go test . -run "TestCommandsDocInSync|TestAbbreviationDocInSync|TestCompletionGoldenInSync" -update
```

### Shell completion

`solmq-conn-util auto-complete bash|zsh|fish|powershell` prints a completion script, rendered by
[`cmd/solmq-conn-util/completion.go`](../cmd/solmq-conn-util/completion.go) from the same
`cliVerbs`/`cliFlags` model as the help and `commands.md` -- so a verb, target or flag added
to the model reaches all four shells with no second edit, and the script a binary prints
always matches the commands that binary accepts. Nothing is shipped as a file; the binary
is the distribution.

Two things gate that, both under `dev.sh test` (so CI, so a tag):

- **`TestCompletionGoldenInSync`** compares each rendered script against a snapshot in
  [`cmd/solmq-conn-util/testdata/completions/`](../cmd/solmq-conn-util/testdata/completions). Those
  are test fixtures, not artifacts: they exist so a model change shows up as a reviewable
  diff in the generated shell code. The `go generate` line above rewrites them along with
  `commands.md`, so **changing a command means regenerating and reading that diff**.
- **`TestCompletionCoversModel`** and its siblings assert the semantics a snapshot cannot:
  every modeled name reaches every shell, the value kinds (`-e`/`-o` complete a path,
  `examples` completes a directory) survive into each script, each script keeps the one
  registration line that makes it load, and the output stays plain ASCII with LF endings.

Adding a verb, target or flag therefore also means giving it the metadata the renderers
need -- a `Blurb` (verbs with targets) or `Summary` (leaf verbs), a `PosArg` kind, and an
`Arg` kind on a flag. `TestCompletionModelMetadataComplete` fails the build when one is
missing or unrecognized rather than letting it render as an empty tooltip.

## The spec generator (`solmq-conn-util-generator.html`)

[`solmq-conn-util-generator.html`](../solmq-conn-util-generator.html) is a standalone, dependency-free
page that builds a spec folder (`env.yaml` + workflow files) from a form. It has **no build
step** and is not part of any dev-script task -- open it from disk and edit it in place.

It carries a JavaScript port of `validate.Run`, `consolidate.Build`, `render.Application`,
`tls.SolaceProps` and the UUIDv5 durable-name derivation so it can lint the spec and preview
the consolidated `application.yml` in the browser. That port is a second implementation and
can drift, so the page embeds a copy of [`testdata/golden/application.yml`](../testdata/golden/application.yml)
in a `<script type="text/plain" id="golden">` block and its **Self-test** button diffs the
shipped sample set against it. A second `<script type="text/plain" id="golden-findings">`
block holds the findings `gen.Validate` reports for that same golden spec folder
(`testdata/golden/specs`), and Self-test also diffs the JS `validateModel` port's findings
against it -- so Self-test checks rule parity as well as the rendered bytes. Both blocks
are kept in sync by `TestGeneratorPageGoldenInSync` and `TestGeneratorPageFindingsGoldenInSync`
(`internal/gen/htmlgolden_test.go`), regenerated the same way: `go test ./internal/gen -run
<TestName> -update-html-golden`. Three rules follow:

- Changing the golden file or the golden spec folder means refreshing the matching embedded
  copy (`-update-html-golden`) in the same change.
- Changing `validate.Run`, `consolidate.Build`, `render.Application`, `tls.SolaceProps` or the
  durable-name derivation means re-running Self-test (open the page, **Load sample**,
  **Self-test**) and porting the change.
- Changing the **env.yaml schema** -- adding, moving or retiring a key -- means porting it
  to the page too: the form field, `readModel`, `emitEnv` and the `validate` mirror, plus
  `sampleModel` and the load path. Self-test's findings diff catches a drift in the ported
  rule itself, but not a schema change that never triggers one -- a page left behind can
  still emit configs the CLI rejects. The image and timezone hoist is the worked example.

The port has no automated syntax gate, but it is plain JavaScript in one `<script>` block,
so it can be checked outside a browser:

```sh
python -c "import re;print(re.findall(r'<script(?![^>]*text/plain)[^>]*>(.*?)</script>',open('solmq-conn-util-generator.html',encoding='utf-8').read(),re.S)[0])" > /tmp/page.js
node --check /tmp/page.js
```

On Windows PowerShell, redirect to a Windows temp path instead:

```powershell
python -c "import re;print(re.findall(r'<script(?![^>]*text/plain)[^>]*>(.*?)</script>',open('solmq-conn-util-generator.html',encoding='utf-8').read(),re.S)[0])" > $env:TEMP\page.js
node --check $env:TEMP\page.js
```

The page never ships secrets: every password field expects a `${VAR}` placeholder and a
literal value is reported as a finding.

## Release (CI)

Two workflows, both calling dev-script task names only (never build commands):

- [`.github/workflows/ci.yml`](../.github/workflows/ci.yml) -- the gates. Runs `all scan` on
  Ubuntu + Windows. Nothing runs on a normal push or PR; it fires when `tag.yml` calls it, on
  manual dispatch, and on PRs that touch CI config or the dev scripts (which gates Dependabot's
  action bumps).
- [`.github/workflows/tag.yml`](../.github/workflows/tag.yml) -- the only automatic pipeline.
  Pushing a `v*` tag runs `plan -> gates -> binaries -> release`, cross-compiling the six
  platform binaries (per the `BUILD_TARGETS` repo variable) and publishing them with a
  `SHA256SUMS.txt` to a **GitHub Release**. A failure anywhere publishes nothing.

```sh
git tag v1.0.0 && git push origin v1.0.0
```

Binaries-only: with no Dockerfile the image job self-skips. Cross-compilation reuses
`dev.sh build` with `TARGET_OS`/`TARGET_ARCH` supplied by the matrix, so local and CI build
the same way. Actions are SHA-pinned and tracked by [`.github/dependabot.yml`](../.github/dependabot.yml);
Go module pins (including govulncheck and the toolchain) move deliberately, gates re-run.

## Design notes

- **Byte-for-byte output.** `application.yml` and the manifests use a deterministic ordered
  emitter (not generic YAML marshaling, which would re-sort keys), so regenerated files diff
  cleanly and match the golden fixture exactly.
- **Layered core.** The CLI (`cmd/solmq-conn-util`) is a thin shell over `internal/gen`, which
  ties parse -> validate -> consolidate -> render together. Packages: `scan`, `spec`,
  `consolidate`, `tls`, `render`, `deploy`, `dockergen`, `podmangen`, `runner`, `validate`,
  `examples`, `gen`, `libs`, `statusscript`, `statusreport`, `yamlwriter`, `logback`.
- **Deploy exec layer.** `internal/runner` shells out to the CLI named by each section's
  `command:` through an `os/exec` argv slice -- never `sh -c`. Every config-derived token is
  validated against a safe charset before it reaches argv (shell metacharacters and control
  chars are rejected with an actionable error), and on top of that
  `validate.CheckDeployCommand` pins argv[0] to a bare, per-platform allowlisted binary
  (kubectl/oc, docker, podman; `--allow-command` approves an extra one per invocation) and
  requires later tokens to be flag-shaped. Before anything mutating runs, `runner.Preflight`
  probes login/daemon reachability read-only; the real Runner resolves argv[0] via
  `exec.LookPath` (rejecting an `exec.ErrDot` same-directory match) and prints
  nothing of its own -- the child's combined output is returned to the caller,
  which reports it. Credential
  env-files are written `0600` and never logged. Kubernetes manifests are piped on
  **stdin** (`apply -f -`), not argv.
- **Teardown reverses the manifest order.** `remove` pipes the document set `deploy`
  renders -- minus the Namespace document, which is handled as its own confirmed
  step -- reversed, to `<command> delete -f -`. kubectl deletes documents
  in file order and waits for each one to go, so creation order deadlocks whenever a
  `libs.pvc.create` PVC precedes the Deployment that mounts it: `kubernetes.io/pvc-protection`
  holds the claim while a pod still mounts it, and the Deployment that owns that pod
  is queued behind it in the file. Reversed, the workload goes first and everything
  it holds is free by the time its turn comes.
- **base-dir vs quadlet-dir split (podman).** The spec names `podman.base-dir`, and that
  is where the rendered `application.yml`, status script, and logback config land. The
  unit directory is deliberately not configurable: systemd only scans
  `/etc/containers/systemd` (root) or `~/.config/containers/systemd` (everyone else), so
  the deploy derives it from the invoking uid and writes only the `.container` unit
  there -- a spec key could only ever name a directory systemd ignores or the account
  cannot write.
- **Streaming seam.** `runner.Runner` is one method (`Run`) and fully buffered: it returns
  only once the process has exited, with stdout and stderr merged into one string for error
  context. `runner.Streamer` is the optional second capability for the case that cannot
  express -- `logs --follow`, where output has to arrive while the process is still running.
  It is a **separate interface, asked for by type assertion**, not a second method on
  `Runner`: most Runners are fakes that only record argv and have no business growing a
  streaming implementation, and `var _ Streamer = OS{}` makes the production assertion a
  compile-time fact. Its two writers are separate on purpose, so a redirected `logs > app.log`
  captures the log and leaves the platform CLI's diagnostics on the terminal. A cancelled
  context is reported as success -- a follow the operator ended with Ctrl-C did not fail --
  and `cmd.Cancel` interrupts, falling back to `Kill` where a signal cannot be delivered
  (Windows), with `WaitDelay` as the backstop for a child that ignores it.
- **Splitting seam.** `runner.Splitter` is the optional capability of returning a process's
  stdout and stderr apart, both fully buffered, instead of `Run`'s single merged string --
  asked for by the same type assertion as `Streamer` and `Attacher`. `runParsed` is how every
  helper that parses output runs it: through `Splitter` when the Runner has it, falling back
  to `Run` otherwise, so every fake Runner that only records argv keeps working unchanged. The
  nine helpers that parse rather than scan their output go through it this way -- the five
  JSON readers (`KubernetesPodsJSON`, `KubernetesGetJSON`, `KubernetesListJSON`,
  `EngineInspectJSON`, `EngineImageInspectJSON`), plus `KubernetesTop`, `EngineList`,
  `EngineStats` and `SystemctlNRestarts` -- because a platform CLI writes warnings to stderr
  whenever it feels like it (OpenShift greets `oc get` with a DeploymentConfig deprecation
  notice), and merged into stdout that line lands in front of the JSON, where it fails to
  parse on its first character. `ScriptInstalled` and `RunStatusScript` deliberately still
  call `Run` directly: they tolerate an extra line, and `RunStatusScript` wants the script's
  own stderr advisories folded into what it returns. `OS` implements `Splitter` too, pinned
  alongside `Streamer` and `Attacher`.
- **Terminal attach seam.** `runner.Attacher` is the third capability beside `Runner` and
  `Streamer`, asked for by the same type assertion, and it exists for the one case neither can
  express: `cli`, where the child must be given the operator's real terminal. Its parameters are
  `*os.File` rather than `io.Reader`/`io.Writer`, and that is the whole point -- `os/exec`
  interposes an OS pipe for any writer that is not an `*os.File`, a pipe is not a terminal, and
  the engine would then refuse the pty it was asked for, leaving a shell with no prompt, no line
  editing and no job control. It returns `(int, error)`: the int is the child's own exit status
  (an `*exec.ExitError` unwrapped here so `os/exec` stays out of the CLI layer), and the error is
  reserved for a child that never started. No `context.Context`, unlike `Stream`: a followed log
  is ended from outside, an attached session from inside by the operator typing `exit`.
  Ctrl-C is handled by *not* handling it -- `signal.Notify` onto a channel nobody reads keeps the
  parent alive while the terminal delivers the interrupt to the child directly. Deliberately not
  `signal.Ignore`, whose SIG_IGN would be inherited across exec and leave the container's shell
  uninterruptible.
- **One container name, always named.** `spec.ConnectorContainerName` (`connector`) is the
  container this tool renders into a kubernetes pod, and `spec.DefaultConnectorName` is the same
  word, so docker and podman default to it too. Every kubernetes `exec` and `logs` argv names it
  outright (`-c`), built in `runner.ExecArgv`/`runner.LogsArgv` from the constant rather than from
  anything discovery returned. A pod carrying a sidecar therefore cannot have the wrong half of it
  read, and a pod with no such container fails loudly instead of being guessed at -- the same
  judgement `statusreport.connectorIndex` records for the reporting side. The accepted cost is
  that a pod this tool did not deploy, whose container is called something else, is not
  reachable this way. `statusreport.Instance.ContainerName` still exists, but it is reported (it
  is a field of the `--output json` document) rather than used to address anything.
- **Shared instance resolution.** `cmd/solmq-conn-util/instances.go` holds what every verb
  that reaches into running instances needs: platform resolution, namespace resolution, the
  `SafeToken` checks on every operator-supplied name, `ParseCommand`, the pod/container
  discovery branches, and -- for the single-instance verbs -- `resolveOneInstance` and the
  paste-back picker. `status`, `logs` and `cli` all call it rather than each carrying a copy --
  a debugging set of verbs that disagreed about which instances it meant would be worse than any
  one of them alone. Because `-c` is a constant, naming an instance outright costs no query at
  all: no `get pods <name>` call is needed first just to learn the container name.
- **Durable names** use UUIDv5 (namespace `6ba7f4e2-9c1d-5a3b-8e47-2f9a0c7d13e5`, key =
  `conn-name || queue-manager || topic || file-basename` joined by `0x1F`). Renaming a workflow
  file changes its durable name and orphans the old subscription -- rename deliberately.
