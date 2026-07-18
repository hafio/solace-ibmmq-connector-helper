# solmq-gen — Development Guide

Building, testing, releasing, and design notes for `solmq-gen`. For using the
tool, see the [user guide](../userguide.md); for a quick start, the
[README](../README.md).

## Build

Requires Go 1.24+ (developed against 1.26). Dependency: `gopkg.in/yaml.v3` only.

```sh
# from connectors/ibmmq/solmq-gen
go mod tidy          # offline if the module cache already has yaml.v3
go build -o solmq-gen ./cmd/solmq-gen     # CGO_ENABLED=0 for a static binary
```

Or use the mirrored task runner (`dev.sh` / `dev.ps1`, behaviorally identical):

```sh
./scripts/dev.sh all          # build + vet + test + cov      (fatal gates)
./scripts/dev.sh full         # all + vuln + cross-platform dist
./scripts/dev.sh dist         # just the cross-compiled binaries -> dist/
.\scripts\dev.ps1 full        # same on Windows PowerShell
```

Tasks: `build vet test cov vuln dist all full`. Correctness gates (build/vet/test/cov)
are fatal; `vuln` (govulncheck) is report-only. `dist` cross-compiles static, CGO-free
binaries for linux/darwin/windows (amd64 + arm64) into `dist/`. `image`/`trivy`/`up`/`down`
are not applicable — the generator ships no Dockerfile or local stack.

## Testing

A worked golden example lives in [`testdata/golden/`](../testdata/golden) and is
asserted byte-for-byte by the tests (`internal/gen/golden_test.go`).

## Release (CI)

[`.github/workflows/release.yml`](../.github/workflows/release.yml) cross-compiles the
six platform binaries and publishes them to a **GitHub Release** when a tag matching
`solmq-gen-v*` is pushed:

```sh
git tag solmq-gen-v1.0.0 && git push origin solmq-gen-v1.0.0
```

> **Monorepo note:** GitHub only runs workflows found at the *repository root*
> `.github/workflows/`. This file is staged under `solmq-gen/.github/workflows/`; move it
> to the repo root (its header repeats this). It scopes to a `solmq-gen-v*` tag prefix and
> builds with `working-directory: connectors/ibmmq/solmq-gen`.

## Design notes

- **Byte-for-byte output.** `application.yml` and the manifests use a deterministic ordered
  emitter (not generic YAML marshaling, which would re-sort keys), so regenerated files diff
  cleanly and match the golden fixture exactly.
- **Layered core.** The CLI (`cmd/solmq-gen`) is a thin shell over `internal/gen`, which
  ties parse → validate → consolidate → render/deploy together. Packages: `scan`, `spec`,
  `consolidate`, `tls`, `render`, `deploy`, `validate`, `gen`.
- **Durable names** use UUIDv5 (namespace `6ba7f4e2-9c1d-5a3b-8e47-2f9a0c7d13e5`, key =
  `conn-name ‖ queue-manager ‖ topic ‖ file-basename` joined by `0x1F`). Renaming a workflow
  file changes its durable name and orphans the old subscription — rename deliberately.
