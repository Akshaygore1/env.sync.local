$ErrorActionPreference = 'Stop'

. (Join-Path $PSScriptRoot 'windows_helpers.ps1')

choco install nsis --no-progress -y

$makensisPath = Resolve-NSISExecutable
$nsisDir = Split-Path -Parent $makensisPath

Add-DirectoryToGithubPath -Directory $nsisDir

Write-Host "Using makensis from $makensisPath"
& $makensisPath /VERSION
