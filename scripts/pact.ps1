$pactDir = "$env:APPVEYOR_BUILD_FOLDER\pact"
$exitCode = 0

# Set environment
if (!($env:GOPATH)) {
  $env:GOPATH = "c:\go"
}
$env:PACT_BROKER_HOST = "https://test.pact.dius.com.au"
$env:PACT_BROKER_USERNAME = "dXfltyFMgNOFZAxr8io9wJ37iUpY42M"
$env:PACT_BROKER_PASSWORD = "O5AIZWxelWbLvqMd8PkAVycBJh2Psyg1"

if (Test-Path "$pactDir") {
  Write-Host "-> Deleting old pact directory"
  rmdir -Recurse -Force $pactDir
}

Write-Host "--> Creating ${pactDir}"
New-Item -Force -ItemType Directory $pactDir

Write-Host "--> Downloading Latest Ruby binaries)"
$downloadDir = $env:TEMP
$latestRelease = Invoke-WebRequest https://github.com/pact-foundation/pact-ruby-standalone/releases/latest -Headers @{"Accept"="application/json"}
$json = $latestRelease.Content | ConvertFrom-Json
$latestVersion = $json.tag_name
$url = "https://github.com/pact-foundation/pact-ruby-standalone/releases/download/v$latestVersion/pact-$latestVersion-win32.zip"

Write-Host "    Downloading $url"
$zip = "$downloadDir\pact.zip"
if (Test-Path "$zip") {
  Remove-Item $zip
}

$downloader = new-object System.Net.WebClient
$downloader.DownloadFile($url, $zip)

Write-Host "    Extracting $zip"
Add-Type -AssemblyName System.IO.Compression.FileSystem
[System.IO.Compression.ZipFile]::ExtractToDirectory("$zip", $pactDir)

Write-Host "    Moving binaries into position"
Get-ChildItem $pactDir\pact

Write-Host "--> Adding pact binaries to path"
$pactBinariesPath = "$pactDir\pact\bin"
$env:PATH += ";$pactBinariesPath"
Write-Host $env:PATH
Get-ChildItem $pactBinariesPath
pact-broker version

Write-Host "--> Running tests"
$packages = go list github.com/pact-foundation/pact-go/... |  where {$_ -inotmatch 'vendor'} | where {$_ -inotmatch 'examples'}
$curDir=$pwd

foreach ($package in $packages) {
  Write-Host "    Running tests for $package"
  go test -v $package
  if ($LastExitCode -ne 0) {
    Write-Host "    ERROR: Test failed, logging failure"
    $exitCode=1
  }
}

Write-Host "--> Testing E2E examples"
$env:PACT_INTEGRATED_TESTS=1

$examples=@("github.com/pact-foundation/pact-go/examples/consumer/goconsumer", "github.com/pact-foundation/pact-go/examples/go-kit/provider", "github.com/pact-foundation/pact-go/examples/mux/provider", "github.com/pact-foundation/pact-go/examples/gin/provider")
foreach ($example in $examples) {
  Write-Host "    Installing dependencies for example: $example"
  cd "$env:GOPATH\src\$example"
  go get ./...
  Write-Host "    Running tests for $example"
  go test -v .
  if ($LastExitCode -ne 0) {
    Write-Host "    ERROR: Test failed, logging failure"
    $exitCode=1
  }
}
cd $curDir

Write-Host "    Shutting down any remaining pact processes :)"
Stop-Process -Name ruby

Write-Host "    Done!"
if ($exitCode -ne 0) {
  Write-Host "--> Build failed, exiting"
  Exit $exitCode
}