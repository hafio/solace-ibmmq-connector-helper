# solmq-conn-util -- development guide

> [!WARNING]
> **Not a supported Solace product.** `solmq-conn-util` was created by Solace
> Professional Services and is supported only by Solace Professional Services --
> not by Solace Support. For help with this tool, contact your Solace
> Professional Services representative rather than opening a Solace Support
> case. This notice covers this tool only, not the Solace PubSub+ Connector for
> IBM MQ that it configures and deploys.

Building, testing and releasing `solmq-conn-util`. For why the code is shaped
the way it is, see [DESIGN.md](DESIGN.md). For using the tool, start at
[getting-started.md](getting-started.md).

## Build

Requires Go 1.27+ (the `go` directive in `go.mod`). Dependencies:
`gopkg.in/yaml.v3`, `software.sslmate.com/src/go-pkcs12` (writes created
stores) and `github.com/pavlo-v-chernykh/keystore-go/v4` (reads a `create.jks`).

```sh
go mod tidy
go build -o solmq-conn-util ./cmd/solmq-conn-util   # CGO_ENABLED=0 for a static binary
```

Or use the task runner (`dev.sh` / `dev.ps1`, identical on both platforms):

```sh
./scripts/dev.sh all          # build + vet + test            (the CI gates; CI runs `all scan`)
./scripts/dev.sh full         # all + cov + scan + graphify   (pre-tag sweep)
./scripts/dev.sh build        # one binary into dist/, for the host or TARGET_OS/TARGET_ARCH
.\scripts\dev.ps1 full
```

| Task | Does |
|---|---|
| `build` | writes `dist/solmq-conn-util-<os>-<arch>[.exe]` |
| `vet` | `go vet ./...` plus a `gofmt -l` check; lists, never writes |
| `test` | `go test ./...`, including every doc and golden gate below |
| `cov` | coverage |
| `scan` | `go tool govulncheck ./...`, fatal on any called vulnerability. One uncalled advisory is expected: `golang.org/x/crypto`'s frozen `openpgp`, which nothing here imports |
| `regen` | runs every `go:generate`; rewrites the generated docs and completion goldens. Local only, in no aggregate, because `test` is what fails on drift |
| `graphify` | refreshes `graphify-out/`; local only |

The Go toolchain is pinned exactly in `go.mod`, so the laptop and CI run the
same Go. govulncheck is a `go.mod` tool dependency, not `go run ...@latest`,
so it builds against the pinned toolchain.

## Testing

A worked golden example lives in [`testdata/golden/`](../testdata/golden) and
is asserted byte for byte by `internal/gen/golden_test.go`.

**Generated docs.** [`commands.md`](commands.md) and
[`abbreviation.md`](abbreviation.md) are rendered from the command model in
[`cmd/solmq-conn-util/commands.go`](../cmd/solmq-conn-util/commands.go). Do not
edit them by hand; `TestCommandsDocInSync` and `TestAbbreviationDocInSync` fail
the build on drift. After changing a command:

```sh
./scripts/dev.sh regen              # = go generate ./...
```

**Hand-written docs.** `TestSupportNoticeInHandWrittenDocs` requires the
support notice from `cmd/solmq-conn-util/support.go` verbatim under the H1 of
every hand-written doc, and `TestSupportNoticeInGeneratorPage` holds the
generator page to the same text.

**Test catalogue.** [`test.md`](test.md) lists every test and case by package.
Keep it in sync when a test or case changes. `TestTestCatalogSnapshotInSync`
checks its `_Snapshot: N test functions, M case rows across P packages._` line
against the repo and names whichever count drifted. There is no `-update` flag;
edit the line by hand after fixing the rows.

**Shell completion.** `auto-complete` renders from the same model as the help
and `commands.md`, in [`completion.go`](../cmd/solmq-conn-util/completion.go).
`TestCompletionGoldenInSync` compares each script with a snapshot under
[`testdata/completions/`](../cmd/solmq-conn-util/testdata/completions), which
`regen` rewrites; read that diff. `TestCompletionCoversModel` and its siblings
check that every modeled name reaches every shell with the right value kind.
`TestCompletionModelMetadataComplete` fails when a new verb, target or flag
lacks its `Blurb`/`Summary`, `PosArg` or `Arg` metadata.

## The spec generator (`solmq-conn-util-generator.html`)

[`solmq-conn-util-generator.html`](../solmq-conn-util-generator.html) is a
standalone page with no build step. It carries a JavaScript port of
`validate.Run`, `consolidate.Build`, `render.Application`, `tls.SolaceProps`
and the durable-name derivation, so it can lint a spec and preview the
`application.yml` in the browser. The port can drift, so:

- The page embeds [`testdata/golden/application.yml`](../testdata/golden/application.yml)
  and the findings `gen.Validate` reports for the golden spec folder. Its
  **Self-test** button diffs the sample set against both.
  `TestGeneratorPageGoldenInSync` and `TestGeneratorPageFindingsGoldenInSync`
  (`internal/gen/htmlgolden_test.go`) keep the embedded copies current;
  regenerate with `go test ./internal/gen -run <TestName> -update-html-golden`.
- Changing any ported function means re-running Self-test (**Load sample**,
  **Self-test**) and porting the change.
- Changing the `env.yaml` schema means porting it to the page: the form field,
  `readModel`, `emitEnv`, the `validate` mirror, `sampleModel` and the load
  path. `TestGeneratorPageKnownKeysInSync` pins the page's key sets against
  `spec`, so a key added to `rawSolace`/`rawMQ` fails the build until the page
  lists it.
- The port covers the static rules only. Store reading and building, and the
  kubernetes `tolerations`/`affinity` checks, are left to the CLI.

The script can be syntax-checked outside a browser:

```sh
python -c "import re;print(re.findall(r'<script(?![^>]*text/plain)[^>]*>(.*?)</script>',open('solmq-conn-util-generator.html',encoding='utf-8').read(),re.S)[0])" > /tmp/page.js
node --check /tmp/page.js
```

The page never ships secrets: every password field expects a `${VAR}`
placeholder, and a literal is reported as a finding.

## Release (CI)

Two workflows, both calling dev-script task names only:

- [`.github/workflows/ci.yml`](../.github/workflows/ci.yml): the gates. Runs
  `all scan` on Ubuntu and Windows. It fires when `tag.yml` calls it, on manual
  dispatch, and on PRs that touch CI config or the dev scripts.
- [`.github/workflows/tag.yml`](../.github/workflows/tag.yml): the only
  automatic pipeline. Pushing a `v*` tag runs plan, gates, binaries, release:
  the six platform binaries (per the `BUILD_TARGETS` variable) plus
  `SHA256SUMS.txt` go to a GitHub Release. A failure anywhere publishes
  nothing.

```sh
git tag v1.0.0 && git push origin v1.0.0
```

There is no Dockerfile; the image job self-skips. Actions are SHA-pinned and
tracked by [`.github/dependabot.yml`](../.github/dependabot.yml).

### Release checklist

What the gates cannot check, before tagging:

- **The connector image still starts the way the docker created-store path
  assumes.** For each supported connector tag, compare `docker image inspect
  --format '{{json .Config.Entrypoint}} {{json .Config.Cmd}}' <image>` with
  `storescript.ImageEntrypoint` / `storescript.ImageCmd` in
  [`internal/storescript/storescript.go`](../internal/storescript/storescript.go).
  `deploy` refuses an image that starts otherwise.
- **A changed store path was run on a real engine.** Store delivery is
  unit-tested only. After a change, run one deploy with a created and an
  existing store on that platform and open each store inside the container
  (`keytool -list -storetype PKCS12`, which the image ships).
