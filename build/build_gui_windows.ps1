$ErrorActionPreference = 'Stop'

$repoRoot = Split-Path -Parent $PSScriptRoot
$distDir = Join-Path $repoRoot 'dist'
$guiDir = Join-Path $repoRoot 'src/gui'

New-Item -ItemType Directory -Force -Path $distDir | Out-Null

Set-Location $guiDir
$env:PATH = "$env:USERPROFILE\go\bin;$env:PATH"
$env:CGO_ENABLED = '1'

wails build -clean -nsis -platform windows/amd64

$portable = Get-ChildItem build/bin -File -Filter '*.exe' |
  Where-Object { $_.Name -notmatch 'installer|setup' } |
  Select-Object -First 1
if (-not $portable) {
  throw 'Portable GUI executable not found'
}

$installer = Get-ChildItem build/bin -File -Filter '*.exe' |
  Where-Object { $_.Name -match 'installer|setup' } |
  Select-Object -First 1
if (-not $installer) {
  throw 'Windows installer not found'
}

Copy-Item $installer.FullName (Join-Path $distDir 'env-sync-gui-windows-amd64-installer.exe')
Compress-Archive -Path $portable.FullName -DestinationPath (Join-Path $distDir 'env-sync-gui-windows-amd64-portable.zip') -Force
