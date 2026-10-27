param([switch]$Mock)
. "$PSScriptRoot\load-env.ps1"
Set-Location -LiteralPath $projectRoot
if ($Mock) { $env:MODEL_MODE='mock' }
& $GoCommand run ./cmd/opsmind serve
exit $LASTEXITCODE
