#!/usr/bin/env bash
# Dev tasks for solmq-conn-util. The only place that knows how to build/vet/test/scan
# this repo. CI calls task names only. Keep dev.ps1 behaviourally identical.
#
# solmq-conn-util is a pure-Go CLI with no Dockerfile and no compose stack, so the
# Docker tasks (image, up, down) do not apply and are omitted.
set -uo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
LOG_DIR="$SCRIPT_DIR/logs"
DIST="$REPO_ROOT/dist"
mkdir -p "$LOG_DIR"
cd "$REPO_ROOT"

export NO_COLOR=1

# --- output helpers ---------------------------------------------------------
c() { printf '\033[%sm%s\033[0m\n' "$1" "$2"; }
step() { c '1;36' "==> $*"; }
ok()   { c '1;32' "ok: $*"; }
warn() { c '1;33' "warn: $*"; }
die()  { c '1;31' "error: $*"; exit 1; }

now() { date +%Y-%m-%dT%H:%M:%S%z; }

# Truncate this task's log with a header, then everything tees onto it.
log_begin() {
  printf '=== %s | %s ===\n' "$(now)" "$1" > "$LOG_DIR/$1.log"
}

# finish <task> <exit-code> <elapsed-seconds>
finish() {
  local task=$1 code=$2 secs=$3 status
  if [ "$code" -eq 0 ]; then status=OK; else status="FAILED (exit $code)"; fi
  printf '%s | %s | %ss | %s\n' "$(now)" "$task" "$secs" "$status" \
    | tee -a "$LOG_DIR/$task.log"
}

# Strip ANSI/CSI so logs stay readable plain text.
strip_csi() { sed -E $'s/\x1b\\[[0-9;?]*[a-zA-Z]//g'; }

# run <task> <cmd...> -- tees combined output, returns the command's code.
run() {
  local task=$1; shift
  "$@" 2>&1 | strip_csi | tee -a "$LOG_DIR/$task.log"
  return "${PIPESTATUS[0]}"
}

# --- target resolution ------------------------------------------------------
# CI sets TARGET_OS/TARGET_ARCH; unset means host. Translate to Go's GOOS/GOARCH
# here and nowhere else. Binary name carries the target so the release job can
# merge every cross-compile leg into one directory without collisions.
host_os()   { uname -s | tr '[:upper:]' '[:lower:]'; }
host_arch() { case "$(uname -m)" in x86_64|amd64) echo amd64;; aarch64|arm64) echo arm64;; *) uname -m;; esac; }
T_OS="${TARGET_OS:-$(host_os)}"
T_ARCH="${TARGET_ARCH:-$(host_arch)}"
BIN_NAME="solmq-conn-util-$T_OS-$T_ARCH"
[ "$T_OS" = "windows" ] && BIN_NAME="$BIN_NAME.exe"

# --- tasks ------------------------------------------------------------------
task_build() {
  mkdir -p "$DIST"
  run build env CGO_ENABLED=0 GOOS="$T_OS" GOARCH="$T_ARCH" \
    go build -trimpath -ldflags "-s -w" -o "$DIST/$BIN_NAME" ./cmd/solmq-conn-util
}

task_vet() { run vet go vet ./...; }

task_test() { run test go test -count=1 ./...; }

task_cov() {
  # Coverage profile -> HTML, and PRINT the total so the footer captures it.
  # The previous total in logs/cov.log is the floor (local only -- CI is a
  # fresh checkout with no prior log).
  # -coverpkg credits cross-package execution (e.g. spec parsing driven by gen
  # tests); without it those lines report 0% and the floor lies. Adding it is
  # a one-time step change in the total -- the floor resets at that run.
  run cov go test -coverpkg=./... -coverprofile=coverage.out -count=1 ./... || return $?
  go tool cover -html=coverage.out -o coverage.html
  go tool cover -func=coverage.out | tail -n1 | tee -a "$LOG_DIR/cov.log"
}

# One task, every applicable check. FATAL on fixable CVEs. solmq-conn-util ships no
# image, so this is the Go dependency scan alone. govulncheck reports only vulns
# reachable from called code -- all actionable -- so a non-zero exit stops the
# run. NOT `go run ...@version`: that form ignores this go.mod, so the scanner
# is built by whatever toolchain its own module asks for, and a govulncheck
# built with go1.25 refuses to load packages from a module declaring go 1.26.
# `go tool` builds it in this module, on this toolchain, at the version go.mod
# pins -- which is also what stops an unpinned scanner arriving mid-release.
task_scan() {
  run scan go tool govulncheck ./...
}

# Local only: the graph is a developer artifact, not a CI output.
task_graphify() {
  [ -n "${CI:-}" ] && { warn "graphify is local-only; skipping in CI"; return 0; }
  command -v graphify >/dev/null || { warn "graphify not on PATH; skipping"; return 0; }
  run graphify graphify update .
}

# --- dispatch ---------------------------------------------------------------
ALL="build vet test"
FULL="build vet test cov scan graphify"

usage() {
  cat <<EOF
usage: $(basename "$0") <task>...

  build vet test cov scan graphify
  all   = $ALL            (what CI runs, as: all scan)
  full  = $FULL   (pre-tag sweep)
EOF
}

expand() {
  case "$1" in
    all)  echo "$ALL" ;;
    full) echo "$FULL" ;;
    *)    echo "$1" ;;
  esac
}

[ $# -eq 0 ] && { usage; exit 0; }
case "${1:-}" in -h|--help|help) usage; exit 0 ;; esac

TASKS=""
for a in "$@"; do TASKS="$TASKS $(expand "$a")"; done

FAILED=0
for task in $TASKS; do
  type "task_$task" >/dev/null 2>&1 || die "unknown task: $task"
  step "$task"
  log_begin "$task"
  start=$SECONDS
  code=0
  "task_$task" || code=$?
  finish "$task" "$code" "$((SECONDS - start))"
  if [ "$code" -ne 0 ]; then
    FAILED=1
    warn "$task failed; stopping"
    break   # build/vet/test/scan are all fatal
  fi
  ok "$task"
done
exit "$FAILED"
