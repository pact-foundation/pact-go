$exitCode = 0
Write-Host "--> Testing E2E examples"
Write-Host "Running provider tests"
$env:SKIP_PUBLISH='true'
go test -v -timeout=30s -tags=provider -count=1 github.com/pact-foundation/pact-go/v2/examples/...
if ($LastExitCode -ne 0) {
  Write-Host "ERROR: E2E Provider Tests failed, logging failure"
  $exitCode=1
}

Write-Host "Done!"
if ($exitCode -ne 0) {
  Write-Host "--> Build failed, exiting"
  Exit $exitCode
}