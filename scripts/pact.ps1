$pactDir = "$env:APPVEYOR_BUILD_FOLDER\pact"
$exitCode = 0

# Set environment
if (!($env:GOPATH)) {
  $env:GOPATH = "c:\go"
}
$env:PACT_BROKER_PROTO = "http"
$env:PACT_BROKER_URL = "localhost"
$env:PACT_BROKER_USERNAME = "pact_workshop"
$env:PACT_BROKER_PASSWORD = "pact_workshop"

if (Test-Path "$pactDir") {
  Write-Host "-> Deleting old pact directory"
  rmdir -Recurse -Force $pactDir
}

# Install CLI Tools
Write-Host "--> Creating ${pactDir}"
New-Item -Force -ItemType Directory $pactDir

Write-Host "--> Downloading Latest Ruby binaries)"
$downloadDir = $env:TEMP
$latestRelease = Invoke-WebRequest https://github.com/pact-foundation/pact-ruby-standalone/releases/latest -Headers @{"Accept"="application/json"}
$json = $latestRelease.Content | ConvertFrom-Json
$tag = $json.tag_name
$latestVersion = $tag.Substring(1)
$url = "https://github.com/pact-foundation/pact-ruby-standalone/releases/download/$tag/pact-$latestVersion-win32.zip"

Write-Host "Downloading $url"
$zip = "$downloadDir\pact.zip"
if (Test-Path "$zip") {
  Remove-Item $zip
}

$downloader = new-object System.Net.WebClient
$downloader.DownloadFile($url, $zip)

Write-Host "Extracting $zip"
Add-Type -AssemblyName System.IO.Compression.FileSystem
[System.IO.Compression.ZipFile]::ExtractToDirectory("$zip", $pactDir)

Write-Host "Moving binaries into position"
Get-ChildItem $pactDir\pact

Write-Host "--> Adding pact binaries to path"
$pactBinariesPath = "$pactDir\pact\bin"
$env:PATH += ";$pactBinariesPath"
Write-Host $env:PATH
Get-ChildItem $pactBinariesPath
pact-broker version


# Run tests
Write-Host "--> Running tests"
$packages = go list github.com/pact-foundation/pact-go/... |  where {$_ -inotmatch 'vendor'} | where {$_ -inotmatch 'examples'}
$curDir=$pwd

foreach ($package in $packages) {
  Write-Host "Running tests for $package"
  go test -v $package
  if ($LastExitCode -ne 0) {
    Write-Host "ERROR: Test failed, logging failure"
    $exitCode=1
  }
}


# Run integration tests
Write-Host "--> Testing E2E examples"
Write-Host "Running consumer tests"
docker compose up -d
go test -tags=consumer -count=1 github.com/pact-foundation/pact-go/examples/... -run TestExample
if ($LastExitCode -ne 0) {
  Write-Host "ERROR: Test failed, logging failure"
  $exitCode=1
}

Write-Host "Running provider tests"
go test -tags=provider -count=1 github.com/pact-foundation/pact-go/examples/... -run TestExample
if ($LastExitCode -ne 0) {
  Write-Host "ERROR: Test failed, logging failure"
  $exitCode=1
}

# Shutdown
Write-Host "Shutting down any remaining pact processes :)"
Stop-Process -Name ruby

Write-Host "Done!"
if ($exitCode -ne 0) {
  Write-Host "--> Build failed, exiting"
  Exit $exitCode
}