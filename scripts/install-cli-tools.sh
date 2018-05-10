#!/bin/bash +e

libDir=$(dirname "$0")
. "${libDir}/lib"

pactDir="build/pact"
version=$(grep "var cliToolsVersion" command/version.go | grep -E -o "([0-9\.]+)")
echo "Installing CLI tools into ${libDir}"

if [ -d "${pactDir}" ]; then
  rm -rf ${pactDir}
fi

step "Installing CLI tools locally"
mkdir -p ${pactDir}
cd build
response=$(curl -s -v https://github.com/pact-foundation/pact-ruby-standalone/releases/latest 2>&1)
tag=$(echo "$response" | grep -o "Location: .*" | sed -e 's/[[:space:]]*$//' | grep -o "Location: .*" | grep -o '[^/]*$')
version=${tag#v}
os="linux-x86_64"
curl -LO https://github.com/pact-foundation/pact-ruby-standalone/releases/download/${tag}/pact-${version}-${os}.tar.gz
tar xzf pact-${version}-${os}.tar.gz
rm pact-${version}-${os}.tar.gz

log "Done!"