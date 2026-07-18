#!/usr/bin/env bash
# solmq-gen task runner (mirror of dev.ps1). Run from anywhere.
#   ./scripts/dev.sh <task> [task...]
# Correctness gates (build/vet/test/cov) are fatal; vuln is report-only.
set -uo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"   # module root = parent of scripts/
LOGS="$SCRIPT_DIR/logs"
mkdir -p "$LOGS"
cd "$ROOT"

# Python for the graphify update (override with PYTHON env var).
PYTHON="${PYTHON:-python}"; command -v "$PYTHON" >/dev/null 2>&1 || PYTHON=python3

# Platforms cross-compiled by `dist` (static, CGO-free — pure Go + yaml.v3).
DIST_TARGETS="linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64 windows/arm64"

c_g=$'\033[32m'; c_y=$'\033[33m'; c_r=$'\033[31m'; c_0=$'\033[0m'
step() { echo "${c_y}==> $*${c_0}"; }
ok()   { echo "${c_g}ok: $*${c_0}"; }
warn() { echo "${c_y}warn: $*${c_0}"; }
die()  { echo "${c_r}FAIL: $*${c_0}"; exit 1; }

# run <logname> <cmd...>  — tee combined output to logs/<logname>.log
run() {
  local name="$1"; shift
  { echo "# $(date) :: $*"; "$@"; } 2>&1 | tee "$LOGS/$name.log"
  return "${PIPESTATUS[0]}"
}

task_build() { step build; run build go build -o solmq-gen ./cmd/solmq-gen || die build; ok "build (-> ./solmq-gen)"; }
task_vet()   { step vet;   run vet   go vet ./...   || die vet;   ok vet; }
task_test()  { step test;  run test  go test ./...  || die test;  ok test; }
task_cov()   {
  step cov
  run cov go test -coverprofile=coverage.out ./... || die "cov (tests)"
  go tool cover -html=coverage.out -o coverage.html
  go tool cover -func=coverage.out | tail -n1
  ok "cov (coverage.html + total above)"
}
task_dist()  {  # cross-compiled static binaries for every platform -> dist/
  step dist
  rm -rf dist; mkdir -p dist
  local log="$LOGS/dist.log"; echo "# $(date)" > "$log"
  local failed=0 os arch ext out
  for t in $DIST_TARGETS; do
    os="${t%/*}"; arch="${t#*/}"; ext=""
    [ "$os" = windows ] && ext=".exe"
    out="dist/solmq-gen-${os}-${arch}${ext}"
    if CGO_ENABLED=0 GOOS="$os" GOARCH="$arch" go build -trimpath -o "$out" ./cmd/solmq-gen 2>>"$log"; then
      ok "  $out"; echo "ok: $out" >> "$log"
    else
      warn "  build FAILED: $t"; echo "FAIL: $t" >> "$log"; failed=1
    fi
  done
  [ "$failed" = 0 ] || die "dist (a target failed; see $log)"
  ok "dist (-> ./dist/)"
}
task_vuln()  { # report-only: never aborts
  step vuln
  if ! command -v govulncheck >/dev/null 2>&1; then
    warn "govulncheck not installed (go install golang.org/x/vuln/cmd/govulncheck@latest); skipping"
    return 0
  fi
  run vuln govulncheck ./... || warn "govulncheck reported findings (report-only)"
  ok vuln
}
# Incrementally update the graphify knowledge graph (AST-only, no API cost).
# Report-only: a missing python/graphify or absent graph warns, never aborts.
task_graphify() { step "graphify"; NO_COLOR=1 run graphify "$PYTHON" -m graphify update . && ok graphify || warn "graphify (report-only)"; }
task_all()   { task_build; task_vet; task_test; task_cov; }   # fast post-change loop
task_full()  { task_all; task_vuln; task_dist; }              # + vuln + cross-platform dist

usage() { cat <<USAGE
solmq-gen dev tasks:
  build   go build -o solmq-gen ./cmd/solmq-gen  (fatal; writes ./solmq-gen)
  vet     go vet ./...                          (fatal)
  test    go test ./...  — golden + unit        (fatal)
  cov     coverage profile -> coverage.html + printed total (fatal)
  vuln    govulncheck ./...                     (report-only)
  gpfy    python -m graphify update .
  dist    cross-compile static binaries for all platforms -> dist/
  all     build + vet + test + cov
  full    all + vuln + dist

Not applicable to this generator (no Dockerfile / local stack): image, trivy, up, down.
USAGE
}

[ $# -eq 0 ] && { usage; exit 0; }
for t in "$@"; do
  case "$t" in
    build) task_build ;; vet) task_vet ;; test) task_test ;;
    cov) task_cov ;; dist) task_dist ;; vuln) task_vuln ;;
    gpfy) task_graphify ;;
    all) task_all ;; full) task_full ;;
    help|-h|--help) usage ;;
    *) die "unknown task: $t" ;;
  esac
done
