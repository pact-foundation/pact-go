$exitCode = 0
# Run integration tests
Write-Host "--> Testing E2E examples"
Write-Host "Running consumer tests"
go test -v -tags=consumer -count=1 github.com/pact-foundation/pact-go/v2/examples/...
if ($LastExitCode -ne 0) {
  Write-Host "ERROR: E2E Consumer Tests failed, logging failure"
  $exitCode=1
}

Write-Host "Done!"
if ($exitCode -ne 0) {
  Write-Host "--> Build failed, exiting"
  Exit $exitCode
}