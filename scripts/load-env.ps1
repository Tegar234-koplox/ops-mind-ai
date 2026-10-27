$ErrorActionPreference = 'Stop'
$projectRoot = Split-Path -Parent $PSScriptRoot
$envFile = Join-Path $projectRoot '.env'
if (Test-Path -LiteralPath $envFile) {
  foreach ($line in Get-Content -LiteralPath $envFile) {
    if ($line -match '^([A-Za-z_][A-Za-z0-9_]*)=(.*)$') {
      [Environment]::SetEnvironmentVariable($Matches[1], $Matches[2], 'Process')
    }
  }
}
$localGo = Join-Path $projectRoot '.tools\go\bin\go.exe'
if (Test-Path -LiteralPath $localGo) {
  $env:GOMODCACHE = Join-Path $projectRoot '.tools\gomodcache'
  $env:GOCACHE = Join-Path $projectRoot '.tools\gocache'
  $script:GoCommand = $localGo
} else {
  $script:GoCommand = 'go'
}
