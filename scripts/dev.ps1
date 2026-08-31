#requires -Version 5.1
# Dev tasks for solmq-conn-util. Behaviourally identical to dev.sh -- same task names,
# same gating, same footer format.
#
# solmq-conn-util is a pure-Go CLI with no Dockerfile and no compose stack, so the
# Docker tasks (image, up, down) do not apply and are omitted.
[CmdletBinding()]
param([Parameter(ValueFromRemainingArguments = $true)][string[]]$Tasks)

$ErrorActionPreference = 'Continue'
$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$RepoRoot  = Split-Path -Parent $ScriptDir
$LogDir    = Join-Path $ScriptDir 'logs'
$Dist      = Join-Path $RepoRoot 'dist'
New-Item -ItemType Directory -Force -Path $LogDir | Out-Null
Set-Location $RepoRoot
$env:NO_COLOR = '1'

# --- output helpers ---------------------------------------------------------
function Step { param($m) Write-Host "==> $m" -ForegroundColor Cyan }
function Ok   { param($m) Write-Host "ok: $m"    -ForegroundColor Green }
function Warn { param($m) Write-Host "warn: $m"  -ForegroundColor Yellow }
function Die  { param($m) Write-Host "error: $m" -ForegroundColor Red; exit 1 }

# Offset with no colon, matching dev.sh's `date +%z` (+0800, not +08:00) so both
# scripts write byte-identical footer shapes.
function Get-Now {
  $d = Get-Date
  '{0}{1}' -f $d.ToString('yyyy-MM-ddTHH:mm:ss'), ($d.ToString('zzz') -replace ':', '')
}
function Get-Log { param($Task) Join-Path $LogDir "$Task.log" }

function Start-TaskLog {
  param($Task)
  Set-Content -Path (Get-Log $Task) -Encoding utf8 `
    -Value ("=== {0} | {1} ===" -f (Get-Now), $Task)
}

function Write-Finish {
  param([string]$Task, [int]$Code, [int]$Seconds)
  $status = if ($Code -eq 0) { 'OK' } else { "FAILED (exit $Code)" }
  $line = '{0} | {1} | {2}s | {3}' -f (Get-Now), $Task, $Seconds, $status
  # Add-Content, never Tee-Object: Tee doubles lines and writes UTF-16.
  Add-Content -Path (Get-Log $Task) -Value $line -Encoding utf8
  Write-Host $line
}

# Capture once, write once. "$_" flattens stderr ErrorRecords; -Width stops
# column wrap; the CSI strip keeps the file readable plain text.
function Invoke-Logged {
  # NB: parameter is $CmdArgs, not $Args -- $Args is an automatic variable, so a
  # param named $Args binds empty and `& $Exe @Args` would run $Exe with no args.
  param([string]$Task, [string]$Exe, [string[]]$CmdArgs)
  $out = (& $Exe @CmdArgs 2>&1 | ForEach-Object { "$_" } | Out-String -Width 4096)
  $code = $LASTEXITCODE
  $out = $out -replace "\x1b\[[0-9;?]*[a-zA-Z]", ""
  Add-Content -Path (Get-Log $Task) -Value $out -Encoding utf8
  Write-Host $out
  return $code
}

# --- target resolution ------------------------------------------------------
function Get-HostArch {
  switch ($env:PROCESSOR_ARCHITECTURE) {
    'AMD64' { 'amd64' } 'ARM64' { 'arm64' } default { 'amd64' }
  }
}
$TOs   = if ($env:TARGET_OS)   { $env:TARGET_OS }   else { 'windows' }
$TArch = if ($env:TARGET_ARCH) { $env:TARGET_ARCH } else { Get-HostArch }
$BinName = "solmq-conn-util-{0}-{1}" -f $TOs, $TArch
if ($TOs -eq 'windows') { $BinName = "$BinName.exe" }

# --- tasks ------------------------------------------------------------------
# Version stamp: git describe, falling back to "dev" when git fails or the
# checkout has no tags (a fresh clone, a shallow CI checkout with no tags
# fetched). Tag-triggered release builds check out the exact tag, so
# --dirty can never fire there -- only a local build off an unclean tree
# ever reports a "-dirty" suffix.
$Version = (git describe --tags --dirty 2>$null)
if ($LASTEXITCODE -ne 0 -or -not $Version) { $Version = 'dev' }

function Task-build {
  New-Item -ItemType Directory -Force -Path $Dist | Out-Null
  $env:CGO_ENABLED = '0'; $env:GOOS = $TOs; $env:GOARCH = $TArch
  $code = Invoke-Logged 'build' 'go' @(
    'build','-trimpath','-ldflags',"-s -w -X main.version=$Version",'-o',(Join-Path $Dist $BinName),'./cmd/solmq-conn-util')
  Remove-Item Env:CGO_ENABLED, Env:GOOS, Env:GOARCH -ErrorAction SilentlyContinue
  return $code
}

function Task-vet  { Invoke-Logged 'vet'  'go' @('vet','./...') }
function Task-test { Invoke-Logged 'test' 'go' @('test','-count=1','./...') }

function Task-cov {
  # Coverage profile -> HTML, and PRINT the total so the footer captures it.
  # -coverpkg credits cross-package execution (e.g. spec parsing driven by gen
  # tests); without it those lines report 0% and the floor lies. Adding it is
  # a one-time step change in the total -- the floor resets at that run.
  $code = Invoke-Logged 'cov' 'go' @('test','-coverpkg=./...','-coverprofile=coverage.out','-count=1','./...')
  if ($code -ne 0) { return $code }
  go tool cover "-html=coverage.out" -o coverage.html | Out-Null
  $total = (go tool cover "-func=coverage.out" | Select-Object -Last 1)
  Add-Content -Path (Get-Log 'cov') -Value $total -Encoding utf8
  Write-Host $total
  return 0
}

# One task, every applicable check. FATAL on fixable CVEs. solmq-conn-util ships no
# image, so this is the Go dependency scan alone. govulncheck reports only vulns
# reachable from called code -- all actionable -- so a non-zero exit stops the
# run. NOT `go run ...@version`: that form ignores this go.mod, so the scanner
# is built by whatever toolchain its own module asks for, and a govulncheck
# built with go1.25 refuses to load packages from a module declaring go 1.26.
# `go tool` builds it in this module, on this toolchain, at the version go.mod
# pins -- which is also what stops an unpinned scanner arriving mid-release.
function Task-scan {
  return (Invoke-Logged 'scan' 'go' @('tool','govulncheck','./...'))
}

# Rewrite the committed artifacts the command model owns -- docs/commands.md,
# docs/abbreviation.md and the four completion goldens -- by running every
# go:generate directive. './...' rather than the one package that has a directive
# today, so a second one is picked up without editing this script.
#
# Local only, and deliberately in NO aggregate: `test` is what fails on drift, so
# a regen inside all/full would quietly rewrite the evidence instead of
# reporting it. Run it after changing a command, then run the gates.
function Task-regen {
  if ($env:CI) { Warn 'regen rewrites committed files; skipping in CI'; return 0 }
  return (Invoke-Logged 'regen' 'go' @('generate','./...'))
}

function Task-graphify {
  if ($env:CI) { Warn 'graphify is local-only; skipping in CI'; return 0 }
  if (-not (Get-Command graphify -ErrorAction SilentlyContinue)) {
    Warn 'graphify not on PATH; skipping'; return 0
  }
  return (Invoke-Logged 'graphify' 'graphify' @('update','.'))
}

# --- dispatch ---------------------------------------------------------------
$All  = @('build','vet','test')
$Full = @('build','vet','test','cov','scan','graphify')

function Show-Usage {
  @"
usage: dev.ps1 <task>...

  build vet test cov scan regen graphify
  all   = $($All -join ' ')            (what CI runs, as: all scan)
  full  = $($Full -join ' ')   (pre-tag sweep)

  regen is local-only and in no aggregate: it rewrites generated files that
  test gates, so run it deliberately, then run the gates.
"@ | Write-Host
}

if (-not $Tasks -or $Tasks[0] -in @('-h','--help','help')) { Show-Usage; exit 0 }

$queue = @()
foreach ($t in $Tasks) {
  switch ($t) { 'all' { $queue += $All } 'full' { $queue += $Full } default { $queue += $t } }
}

$failed = 0
foreach ($task in $queue) {
  if (-not (Get-Command "Task-$task" -ErrorAction SilentlyContinue)) { Die "unknown task: $task" }
  Step $task
  Start-TaskLog $task
  $sw = [Diagnostics.Stopwatch]::StartNew()
  $code = 0
  try { $code = & "Task-$task" } catch { $code = 1 }
  if ($null -eq $code) { $code = 0 }
  $sw.Stop()
  Write-Finish -Task $task -Code $code -Seconds ([int]$sw.Elapsed.TotalSeconds)
  if ($code -ne 0) { $failed = 1; Warn "$task failed; stopping"; break }
  Ok $task
}
exit $failed
