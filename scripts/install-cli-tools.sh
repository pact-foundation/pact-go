#!/bin/bash +e

libDir=$(dirname "$0")
. "${libDir}/lib"

buildDir="build"
pactDir="${buildDir}/pact"
version=$(grep "var cliToolsVersion" command/version.go | grep -E -o "([0-9\.]+)")
step "Installing CLI tools locally into ${pactDir}"
log "Expecting version to be at least ${version}"
log "Installing CLI tools into ${libDir}"

if [ -d "${pactDir}" ]; then
  log "Removing existing directory"
  rm -rf ${pactDir}
fi
mkdir -p ${pactDir}
cd ${buildDir}

# Detect OS, default to linux 64
uname_output=$(uname)
log "Detecting OS. Output of 'uname': ${uname_output}"
case $uname_output in
  'Linux')
    linux_uname_output=$(uname -i)
    case $linux_uname_output in
      'x86_64')
        os='linux-x86_64'
        ;;
      'i686')
        os='linux-x86'
        ;;
      *)
        log "Can't determine OS, defaulting to Linux 64bit"
        os='linux-x86_64'
        ;;
    esac
    ;;
  'Darwin')
    os='osx'
    ;;
  *)
  log "Can't determine OS, defaulting to Linux 64bit"
  os='linux-x86_64'
    ;;
esac

log "OS Detected: ${os}"
log "Finding latest version from GitHub"
response=$(curl -s -v https://github.com/pact-foundation/pact-ruby-standalone/releases/latest 2>&1)
tag=$(echo "$response" | grep -o "Location: .*" | sed -e 's/[[:space:]]*$//' | grep -o "Location: .*" | grep -o '[^/]*$')
version=${tag#v}

log "Downloading version ${version}"
curl -LO https://github.com/pact-foundation/pact-ruby-standalone/releases/download/${tag}/pact-${version}-${os}.tar.gz
tar xzf pact-${version}-${os}.tar.gz
rm pact-${version}-${os}.tar.gz

log "Done!"