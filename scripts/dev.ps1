# solmq-gen task runner (mirror of dev.sh). Run from anywhere.
#   .\scripts\dev.ps1 <task> [task...]
# Correctness gates (build/vet/test/cov) are fatal; vuln is report-only.
param([Parameter(ValueFromRemainingArguments = $true)][string[]]$Tasks)

$ErrorActionPreference = 'Continue'
$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$Root = Split-Path -Parent $ScriptDir          # module root = parent of scripts/
$Logs = Join-Path $ScriptDir 'logs'
New-Item -ItemType Directory -Force -Path $Logs | Out-Null
Set-Location $Root

# Python for the graphify update (override with $env:PYTHON).
$Python = if ($env:PYTHON) { $env:PYTHON }
  elseif (Get-Command python -ErrorAction SilentlyContinue) { 'python' }
  elseif (Get-Command python3 -ErrorAction SilentlyContinue) { 'python3' }
  else { 'python' }

# Platforms cross-compiled by `dist` (static, CGO-free — pure Go + yaml.v3).
$DistTargets = @('linux/amd64', 'linux/arm64', 'darwin/amd64', 'darwin/arm64', 'windows/amd64', 'windows/arm64')

function Step($m) { Write-Host "==> $m" -ForegroundColor Yellow }
function Ok($m)   { Write-Host "ok: $m" -ForegroundColor Green }
function Warn($m) { Write-Host "warn: $m" -ForegroundColor Yellow }
function Die($m)  { Write-Host "FAIL: $m" -ForegroundColor Red; exit 1 }

# Run <logname> <scriptblock>: run cmd, tee combined output to a UTF-8 log, and
# return ONLY the exit code (captured output must not leak into the pipeline).
function Run($name, [scriptblock]$cmd) {
  $log = Join-Path $Logs "$name.log"
  $output = & $cmd 2>&1 | Out-String
  $code = $LASTEXITCODE
  "# $(Get-Date)`r`n$output" | Out-File -Encoding utf8 $log
  if ($output.Trim()) { Write-Host $output.TrimEnd() }
  return $code
}

function Task-Build { Step build; if ((Run 'build' { go build -o solmq-gen.exe ./cmd/solmq-gen }) -ne 0) { Die build }; Ok 'build (-> .\solmq-gen.exe)' }
function Task-Vet   { Step vet;   if ((Run 'vet'   { go vet ./... })  -ne 0) { Die vet };   Ok vet }
function Task-Test  { Step test;  if ((Run 'test'  { go test -count=1 ./... }) -ne 0) { Die test }; Ok test }
function Task-Cov {
  Step cov
  if ((Run 'cov' { go test "-coverprofile=coverage.out" -count=1 ./... }) -ne 0) { Die 'cov (tests)' }
  go tool cover "-html=coverage.out" -o coverage.html | Out-Null
  go tool cover "-func=coverage.out" | Select-Object -Last 1
  Ok 'cov (coverage.html + total above)'
}
function Task-Dist {  # cross-compiled static binaries for every platform -> dist\
  Step dist
  Remove-Item -Recurse -Force dist -ErrorAction SilentlyContinue
  New-Item -ItemType Directory -Force -Path dist | Out-Null
  $log = @("# $(Get-Date)"); $failed = 0
  foreach ($t in $DistTargets) {
    $parts = $t.Split('/'); $os = $parts[0]; $arch = $parts[1]
    $ext = if ($os -eq 'windows') { '.exe' } else { '' }
    $out = "dist/solmq-gen-$os-$arch$ext"
    $env:CGO_ENABLED = '0'; $env:GOOS = $os; $env:GOARCH = $arch
    $o = go build -trimpath -o $out ./cmd/solmq-gen 2>&1 | Out-String
    if ($LASTEXITCODE -eq 0) { Ok "  $out"; $log += "ok: $out" }
    else { Warn "  build FAILED: $t"; if ($o.Trim()) { Write-Host $o.TrimEnd() }; $log += "FAIL: $t`n$o"; $failed = 1 }
  }
  Remove-Item Env:CGO_ENABLED, Env:GOOS, Env:GOARCH -ErrorAction SilentlyContinue
  $log -join "`r`n" | Out-File -Encoding utf8 (Join-Path $Logs 'dist.log')
  if ($failed -ne 0) { Die 'dist (a target failed; see logs/dist.log)' }
  Ok 'dist (-> .\dist\)'
}
function Task-Vuln { # report-only
  Step vuln
  if (-not (Get-Command govulncheck -ErrorAction SilentlyContinue)) {
    Warn 'govulncheck not installed (go install golang.org/x/vuln/cmd/govulncheck@latest); skipping'; return
  }
  if ((Run 'vuln' { govulncheck ./... }) -ne 0) { Warn 'govulncheck reported findings (report-only)' } else { Ok vuln }
}
# Incrementally update the graphify knowledge graph (AST-only, no API cost).
# Report-only: a missing python/graphify or absent graph warns, never aborts.
function Task-Graphify {
  Step 'graphify'
  $env:NO_COLOR = '1'
  if ((Run 'graphify' { & $Python -m graphify update . }) -eq 0) { Ok 'graphify' } else { Warn 'graphify (report-only)' }
}
function Task-All  { Task-Build; Task-Vet; Task-Test; Task-Cov; Task-Graphify }
function Task-Full { Task-All; Task-Vuln; Task-Dist }

function Usage {
  @'
solmq-gen dev tasks:
  build   go build -o solmq-gen.exe ./cmd/solmq-gen (fatal; writes .\solmq-gen.exe)
  vet     go vet ./...                          (fatal)
  test    go test ./...  - golden + unit        (fatal)
  cov     coverage profile -> coverage.html + printed total (fatal)
  vuln    govulncheck ./...                     (report-only)
  gpfy    python -m graphify update .
  dist    cross-compile static binaries for all platforms -> dist\
  all     build + vet + test + cov
  full    all + vuln + dist

Not applicable to this generator (no Dockerfile / local stack): image, trivy, up, down.
'@ | Write-Host
}

if (-not $Tasks -or $Tasks.Count -eq 0) { Usage; exit 0 }
foreach ($t in $Tasks) {
  switch ($t) {
    'build' { Task-Build } 'vet' { Task-Vet } 'test' { Task-Test }
    'cov' { Task-Cov } 'dist' { Task-Dist } 'vuln' { Task-Vuln }
    'gpfy' { Task-Graphify } 
    'all' { Task-All } 'full' { Task-Full }
    { $_ -in 'help', '-h', '--help' } { Usage }
    default { Die "unknown task: $t" }
  }
}
