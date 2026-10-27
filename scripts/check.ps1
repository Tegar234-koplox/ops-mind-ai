. "$PSScriptRoot\load-env.ps1"
Set-Location -LiteralPath $projectRoot
& $GoCommand vet ./...
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
& $GoCommand test ./...
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
Push-Location web
try { npm.cmd run build; if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE } } finally { Pop-Location }
