$exitCode = 0

# Run unit tests
Write-Host "--> Running unit tests"
$packages = go list -buildvcs=false ./... |  Where-Object {$_ -inotmatch 'vendor'} | Where-Object {$_ -inotmatch 'examples'}

foreach ($package in $packages) {
  Write-Host "Running tests for $package"
  go test -race -v $package
  if ($LastExitCode -ne 0) {
    Write-Host "ERROR: Test failed, logging failure"
    $exitCode=1
  }
}

Write-Host "Done!"
if ($exitCode -ne 0) {
  Write-Host "--> Build failed, exiting"
  Exit $exitCode
}