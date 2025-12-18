# HookRunner Build Script for Windows
# Optimized build with stripped symbols for smaller binaries

param(
    [Parameter(Position=0)]
    [ValidateSet("build", "build-dev", "install", "test", "clean", "release", "size", "help")]
    [string]$Target = "build"
)

# Get version info
$VERSION = if (Test-Path "VERSION") { (Get-Content VERSION).Trim() } else { "dev" }
$GIT_COMMIT = git rev-parse --short HEAD 2>$null
if (-not $GIT_COMMIT) { $GIT_COMMIT = "unknown" }
$BUILD_DATE = (Get-Date).ToUniversalTime().ToString("yyyy-MM-ddTHH:mm:ssZ")

# Optimized ldflags: -s strips symbol table, -w strips DWARF debug info
$LDFLAGS = "-s -w -X github.com/ashavijit/hookrunner/internal/version.Version=$VERSION -X github.com/ashavijit/hookrunner/internal/version.GitCommit=$GIT_COMMIT -X github.com/ashavijit/hookrunner/internal/version.BuildDate=$BUILD_DATE"

$BINARY = "hookrunner.exe"

function Build-Optimized {
    Write-Host "Building optimized binary..." -ForegroundColor Cyan
    go build -ldflags $LDFLAGS -o $BINARY ./cmd/hookrunner
    $size = (Get-Item $BINARY).Length / 1MB
    Write-Host "Built $BINARY ($([math]::Round($size, 2)) MB)" -ForegroundColor Green
}

function Build-Dev {
    Write-Host "Building development binary..." -ForegroundColor Cyan
    go build -o $BINARY ./cmd/hookrunner
    $size = (Get-Item $BINARY).Length / 1MB
    Write-Host "Built $BINARY (dev mode, $([math]::Round($size, 2)) MB)" -ForegroundColor Yellow
}

function Install-Binary {
    Write-Host "Installing hookrunner..." -ForegroundColor Cyan
    go install -ldflags $LDFLAGS ./cmd/hookrunner
    $gopath = go env GOPATH
    Write-Host "Installed to $gopath\bin\hookrunner.exe" -ForegroundColor Green
}

function Run-Tests {
    Write-Host "Running tests..." -ForegroundColor Cyan
    go test ./... -v -cover
}

function Clean-Artifacts {
    Write-Host "Cleaning build artifacts..." -ForegroundColor Cyan
    Remove-Item -Force -ErrorAction SilentlyContinue hookrunner.exe, hookrunner, hookrunner_optimized.exe
    Remove-Item -Force -ErrorAction SilentlyContinue coverage.out, coverage.html
    Remove-Item -Recurse -Force -ErrorAction SilentlyContinue dist
    Write-Host "Cleaned!" -ForegroundColor Green
}

function Build-Release {
    Write-Host "Building release binaries..." -ForegroundColor Cyan
    New-Item -ItemType Directory -Force -Path dist | Out-Null

    $env:GOOS = "windows"; $env:GOARCH = "amd64"
    go build -ldflags $LDFLAGS -o dist/hookrunner-windows-amd64.exe ./cmd/hookrunner

    $env:GOOS = "linux"; $env:GOARCH = "amd64"
    go build -ldflags $LDFLAGS -o dist/hookrunner-linux-amd64 ./cmd/hookrunner

    $env:GOOS = "darwin"; $env:GOARCH = "amd64"
    go build -ldflags $LDFLAGS -o dist/hookrunner-darwin-amd64 ./cmd/hookrunner

    $env:GOOS = "darwin"; $env:GOARCH = "arm64"
    go build -ldflags $LDFLAGS -o dist/hookrunner-darwin-arm64 ./cmd/hookrunner

    # Reset
    $env:GOOS = ""; $env:GOARCH = ""

    Write-Host "Release binaries:" -ForegroundColor Green
    Get-ChildItem dist | Format-Table Name, @{N='Size';E={"{0:N2} MB" -f ($_.Length/1MB)}}
}

function Show-Size {
    Write-Host "=== Binary Size Comparison ===" -ForegroundColor Cyan

    Write-Host "`nOptimized build:" -ForegroundColor Green
    go build -ldflags "-s -w" -o hookrunner_opt.exe ./cmd/hookrunner
    $optSize = (Get-Item hookrunner_opt.exe).Length / 1MB
    Write-Host "  $([math]::Round($optSize, 2)) MB"

    Write-Host "`nDebug build:" -ForegroundColor Yellow
    go build -o hookrunner_dbg.exe ./cmd/hookrunner
    $dbgSize = (Get-Item hookrunner_dbg.exe).Length / 1MB
    Write-Host "  $([math]::Round($dbgSize, 2)) MB"

    $savings = (1 - $optSize/$dbgSize) * 100
    Write-Host "`nSavings: $([math]::Round($savings, 1))%" -ForegroundColor Cyan

    Remove-Item -Force hookrunner_opt.exe, hookrunner_dbg.exe
}

function Show-Help {
    Write-Host @"
HookRunner Build Script

Usage: .\build.ps1 [target]

Targets:
  build      Build optimized binary (default)
  build-dev  Build with debug info (faster compile)
  install    Install to GOPATH/bin
  test       Run tests
  clean      Remove build artifacts
  release    Cross-compile for all platforms
  size       Compare optimized vs debug binary sizes
  help       Show this help
"@
}

# Execute target
switch ($Target) {
    "build"     { Build-Optimized }
    "build-dev" { Build-Dev }
    "install"   { Install-Binary }
    "test"      { Run-Tests }
    "clean"     { Clean-Artifacts }
    "release"   { Build-Release }
    "size"      { Show-Size }
    "help"      { Show-Help }
}
