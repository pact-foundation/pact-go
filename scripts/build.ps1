$pactDir = $env:USERPROFILE + "\.pact"
$pactBinDir = $pactDir + "\bin"
$exitCode = 0

# go build -buildvcs=false -o build/pact-go.exe
go build -o build/pact-go.exe
if ($LastExitCode -ne 0) {
  Write-Host "ERROR: Failed to build pact-go"
  $exitCode=1
}

./build/pact-go.exe -l DEBUG install --libDir $env:TMP
if ($LastExitCode -ne 0) {
  Write-Host "ERROR: Failed to install pact-go library to $env:TMP"
  $exitCode=1
}

# Install CLI Tools
if (!(Test-Path -Path $pactBinDir\pact-plugin-cli.exe)) {
Write-Host "--> Creating ${pactDir}"
New-Item -Force -ItemType Directory $pactDir
New-Item -Force -ItemType Directory $pactBinDir

Write-Host "--> Downloading Pact plugin CLI tools"
$downloadDir = $env:TEMP
$pactPluginCliVersion = "0.1.2"
$url = "https://github.com/pact-foundation/pact-plugins/releases/download/pact-plugin-cli-v${pactPluginCliVersion}/pact-plugin-cli-windows-x86_64.exe.gz"

Write-Host "Downloading $url"
$gz = "$downloadDir\pact-plugin-cli-windows-x86_64.exe.gz"
if (Test-Path "$gz") {
  Remove-Item $gz
}

$downloader = new-object System.Net.WebClient
$downloader.DownloadFile($url, $gz)

Write-Host "Extracting $gz"
Add-Type -AssemblyName System.IO.Compression.FileSystem
gzip -d $gz
Move-Item "$downloadDir\pact-plugin-cli-windows-x86_64.exe" "$pactBinDir\pact-plugin-cli.exe"
}

Write-Host "--> Adding pact binaries to path"
$env:PATH += ";$pactBinDir"
Write-Host $env:PATH
Get-ChildItem $pactBinDir
Write-Host "Added pact binaries to path"

pact-plugin-cli.exe --version
if ($LastExitCode -ne 0) {
  Write-Host "ERROR: Failed to install pact-plugin-cli"
  $exitCode=1
}

pact-plugin-cli.exe -y install https://github.com/mefellows/pact-matt-plugin/releases/tag/v0.1.1 --skip-if-installed
if ($LastExitCode -ne 0) {
  Write-Host "ERROR: Failed to install pact-matt-plugin"
  $exitCode=1
}

Write-Host "Done!"
if ($exitCode -ne 0) {
  Write-Host "--> Build failed, exiting"
  Exit $exitCode
}