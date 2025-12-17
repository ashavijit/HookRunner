$ErrorActionPreference = "Stop"

$Repo = "ashavijit/HookRunner"
$InstallDir = if ($env:INSTALL_DIR) { $env:INSTALL_DIR } else { "$env:LOCALAPPDATA\hookrunner" }
$Version = if ($env:VERSION) { $env:VERSION } else { "latest" }
$Arch = if ([Environment]::Is64BitOperatingSystem) { "amd64" } else { "386" }

if ($Version -eq "latest") {
    try {
        $Release = Invoke-RestMethod -Uri "https://api.github.com/repos/$Repo/releases/latest" -Headers @{"User-Agent"="HookRunner-Installer"}
        $Version = $Release.tag_name
    } catch {
        $Version = "v0.19.0"
    }
}

$Binary = "hookrunner-windows-$Arch.exe"
$Url = "https://github.com/$Repo/releases/download/$Version/$Binary"

Write-Host ""
Write-Host "Installing HookRunner $Version..." -ForegroundColor Cyan
Write-Host "  Architecture: $Arch"
Write-Host "  Install Dir: $InstallDir"
Write-Host ""

if (-not (Test-Path $InstallDir)) {
    New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
}

$DestPath = Join-Path $InstallDir "hookrunner.exe"

Write-Host "Downloading..." -ForegroundColor Yellow
try {
    Invoke-WebRequest -Uri $Url -OutFile $DestPath -UseBasicParsing
} catch {
    Write-Error "Download failed: $Url"
    Write-Host "Try: go install github.com/ashavijit/hookrunner/cmd/hookrunner@latest" -ForegroundColor Yellow
    exit 1
}

Write-Host "Downloaded!" -ForegroundColor Green

$UserPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($UserPath -notlike "*$InstallDir*") {
    [Environment]::SetEnvironmentVariable("Path", "$UserPath;$InstallDir", "User")
    $env:Path = "$env:Path;$InstallDir"
    Write-Host "Added to PATH" -ForegroundColor Green
}

Write-Host ""
Write-Host "HookRunner installed!" -ForegroundColor Green
Write-Host ""
Write-Host "  hookrunner init --lang go"
Write-Host "  hookrunner install"
Write-Host "  hookrunner --help"
Write-Host ""

try { & $DestPath version } catch { }
