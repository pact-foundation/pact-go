$pactDir="$env:TEMP\pact"
rmdir -Recurse -Force $pactDir

Write-Verbose "Creating ${pactDir}"
New-Item -Force -ItemType Directory $pactDir

Write-Verbose "Creating pact-go binary"
go build -o "$pactDir\pact-go.exe" "github.com\pact-foundation\pact-go"

Write-Verbose "Creating pact daemon (downloading Ruby binaries)"

$downloadDir=$env:TEMP
$downloadDirMockService = "$pactDir\pact-mock-service"
$downloadDirPRoviderVerifier = "$pactDir\pact-provider-verifier"
$url = "https://github.com/pact-foundation/pact-ruby-standalone/releases/download/v1.2.5/pact-1.2.5-win32.zip"

Write-Verbose "Downloading $url"
$zip = "$downloadDir\pact.zip"
if (!(Test-Path "$zip")) {
  $downloader = new-object System.Net.WebClient
  $downloader.DownloadFile($url, $zip)
}

Write-Verbose "Extracting $zip"
Add-Type -AssemblyName System.IO.Compression.FileSystem
[System.IO.Compression.ZipFile]::ExtractToDirectory("$zip", $downloadDirMockService)
[System.IO.Compression.ZipFile]::ExtractToDirectory("$zip", $downloadDirPRoviderVerifier)

Write-Verbose "Moving binaries into position"
mv $downloadDirMockService/pact/* $pactDir/pact-mock-service/
mv $downloadDirPRoviderVerifier/pact/* $pactDir/pact-provider-verifier/
Get-ChildItem $pactDir

Write-Verbose "Starting pact daemon in background"
Start-Process -FilePath "$pactDir\pact-go.exe" -ArgumentList "daemon -v -l DEBUG"  -RedirectStandardOutput "pact.log" -RedirectStandardError "pact-error.log"

Write-Verbose "Running tests"
$env:PACT_INTEGRATED_TESTS=1
$env:PACT_BROKER_HOST="https://test.pact.dius.com.au"
$env:PACT_BROKER_USERNAME="dXfltyFMgNOFZAxr8io9wJ37iUpY42M"
$env:PACT_BROKER_PASSWORD="O5AIZWxelWbLvqMd8PkAVycBJh2Psyg1"
go test github.com/pact-foundation/pact-go

$packages = go list github.com/pact-foundation/pact-go/... |  where {$_ -inotmatch 'vendor'} | where {$_ -inotmatch 'vendor'}
$curDir=$pwd
foreach ($package in $packages) {
 Write-Verbose "Running tests for $package"
 go test -v $package
}

Write-Verbose "Testing examples"
$examples=@("github.com/pact-foundation/pact-go/examples/consumer/goconsumer", "github.com/pact-foundation/pact-go/examples/go-kit/provider", "github.com/pact-foundation/pact-go/examples/mux/provider")
foreach ($example in $examples) {
  Write-Verbose "Installing dependencies for example: $example"
  cd "$env:GOPATH\src\$example"
  go get ./...
  Write-Verbose "Running tests for $example"
  go test -v .
}

cd $curDir
Write-Verbose "Shutting down pact processes :)"

Stop-Process -Name ruby
Stop-Process -Name pact-go

Write-Verbose "Done!"