$ErrorActionPreference = 'Stop'

function Resolve-Executable {
  param(
    [Parameter(Mandatory = $true)]
    [string]$CommandName,
    [string[]]$CandidateDirectories = @()
  )

  $command = Get-Command $CommandName -ErrorAction SilentlyContinue
  if ($command) {
    return $command.Source
  }

  foreach ($directory in $CandidateDirectories) {
    if ([string]::IsNullOrWhiteSpace($directory)) {
      continue
    }

    $exePath = Join-Path $directory "$CommandName.exe"
    if (Test-Path $exePath) {
      $env:PATH = "$directory;$env:PATH"
      return $exePath
    }
  }

  throw "Required executable '$CommandName' was not found. Checked PATH and: $($CandidateDirectories -join ', ')"
}

function Resolve-NSISExecutable {
  return Resolve-Executable -CommandName 'makensis' -CandidateDirectories @(
    "${env:ProgramFiles(x86)}\NSIS",
    "$env:ProgramFiles\NSIS",
    "$env:ChocolateyInstall\bin"
  )
}

function Add-DirectoryToGithubPath {
  param(
    [Parameter(Mandatory = $true)]
    [string]$Directory
  )

  if ([string]::IsNullOrWhiteSpace($Directory)) {
    throw 'Directory must not be empty'
  }

  $env:PATH = "$Directory;$env:PATH"

  if (-not [string]::IsNullOrWhiteSpace($env:GITHUB_PATH)) {
    $Directory | Out-File -FilePath $env:GITHUB_PATH -Encoding utf8 -Append
  }
}
