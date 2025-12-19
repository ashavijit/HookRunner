$packageName = 'hookrunner'
$version      = '0.45.0'
$url          = 'https://github.com/ashavijit/HookRunner/releases/download/v0.45.0/hookrunner-windows-amd64.exe'
$checksum     = '716F37EE95D5978D9F2BE2A6546738D0688FA58CB09F6BAE5804C7334392634E'

$toolsDir     = "$(Split-Path -parent $MyInvocation.MyCommand.Definition)"

Install-ChocolateyPackage `
  -PackageName "$packageName" `
  -Url "$url" `
  -Checksum "$checksum" `
  -ChecksumType "sha256" `
  -FileFullPath "$toolsDir\hookrunner.exe"

# Warn if not 64-bit (optional, since URL above is amd64 specific)
if (![Environment]::Is64BitOperatingSystem) {
  Write-Warning "This package currently supports Windows 64-bit only via Chocolatey."
}
