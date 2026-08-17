# solmq-conn-util -- Development Guide

Building, testing, releasing, and design notes for `solmq-conn-util`. For using the
tool, see the [user guide](../userguide.md); for a quick start, the
[README](../README.md).

## Build

Requires Go 1.24+ (developed against 1.26). Dependency: `gopkg.in/yaml.v3` only.

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

Tasks: `build vet test cov scan graphify`, plus aggregates `all` (= build vet test, run by CI
as `all scan`) and `full` (= all + cov + scan + graphify, the pre-tag sweep). Gates
build/vet/test/scan are fatal. `scan` is `go tool govulncheck ./...`, fatal on any finding — every
Go vuln-DB finding is reachable-and-fixable, so there is nothing to warn-and-pass on. The scanner is
pinned in `go.mod` as a tool dependency rather than invoked as `go run ...@latest`: that form ignores
`go.mod` and builds the scanner on whatever toolchain its own module requires, which then cannot load
packages from a module on a newer `go` directive. The Go toolchain itself is pinned exactly
(`toolchain` directive in `go.mod`), so the laptop and both CI runners download and run the identical
Go rather than whatever the runner image preinstalls -- bump the pin deliberately when upgrading
locally. `build` honors `TARGET_OS`/`TARGET_ARCH` and writes
`dist/solmq-conn-util-<os>-<arch>[.exe]` (host os/arch when unset), so one task serves the laptop and
the CI matrix. `graphify` is local-only (warn-skips under CI). `image`/`up`/`down` are omitted --
the tool ships no Dockerfile or local stack (it generates artifacts for other engines; it is not
itself containerized).

## Testing

A worked golden example lives in [`testdata/golden/`](../testdata/golden) and is
asserted byte-for-byte by the tests (`internal/gen/golden_test.go`).

[`test.md`](test.md) is the full test catalogue -- every test, grouped by package and
expanded to individual cases. Keep it in sync: a test or case added, removed, or renamed
updates the matching row in the same change.

[`commands.md`](commands.md) is the **generated** command reference, rendered from the
command model in [`cmd/solmq-conn-util/commands.go`](../cmd/solmq-conn-util/commands.go). Do not edit
it by hand; `TestCommandsDocInSync` fails the build if it drifts from the model. Regenerate
after changing a command:

```sh
go generate ./cmd/solmq-conn-util   # = go test . -run "TestCommandsDocInSync|TestCompletionGoldenInSync" -update
```

### Shell completion

`solmq-conn-util completion bash|zsh|fish|powershell` prints a completion script, rendered by
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
shipped sample set against it. Two rules follow:

- Changing the golden file means refreshing the embedded copy in the same change.
- Changing `consolidate`, `render`, `tls` or the durable-name derivation means re-running
  Self-test (open the page, **Load sample**, **Self-test**) and porting the change.

The page never ships secrets: every password field expects a `${VAR}` placeholder and a
literal value is reported as a finding.

## Release (CI)

Two workflows, both calling dev-script task names only (never build commands):

- [`.github/workflows/ci.yml`](../.github/workflows/ci.yml) — the gates. Runs `all scan` on
  Ubuntu + Windows. Nothing runs on a normal push or PR; it fires when `tag.yml` calls it, on
  manual dispatch, and on PRs that touch CI config or the dev scripts (which gates Dependabot's
  action bumps).
- [`.github/workflows/tag.yml`](../.github/workflows/tag.yml) — the only automatic pipeline.
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
  `examples`, `gen`.
- **Deploy exec layer.** `internal/runner` shells out to the CLI named by each target's
  `command:` through an `os/exec` argv slice -- never `sh -c`. Every config-derived token is
  validated against a safe charset before it reaches argv (shell metacharacters and control
  chars are rejected with an actionable error), and on top of that
  `validate.CheckDeployCommand` pins argv[0] to a bare, per-platform allowlisted binary
  (kubectl/oc, docker, podman; `--allow-command` approves an extra one per invocation) and
  requires later tokens to be flag-shaped. Before anything mutating runs, `runner.Preflight`
  probes login/daemon reachability read-only; the real Runner resolves argv[0] via
  `exec.LookPath` and echoes the resolved path to stderr before exec'ing. Credential
  env-files are written `0600` and never logged. Kubernetes manifests are piped on
  **stdin** (`apply -f -`), not argv.
- **Durable names** use UUIDv5 (namespace `6ba7f4e2-9c1d-5a3b-8e47-2f9a0c7d13e5`, key =
  `conn-name ‖ queue-manager ‖ topic ‖ file-basename` joined by `0x1F`). Renaming a workflow
  file changes its durable name and orphans the old subscription — rename deliberately.
