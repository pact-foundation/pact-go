#!/bin/bash +e

libDir=$(dirname "$0")
. "${libDir}/lib"

echo "Installing CLI tools into ${libDir}"
if [ ! -d "build/pact" ]; then
    step "Installing CLI tools locally"
    mkdir -p build/pact
    cd build
    # curl -fsSL https://raw.githubusercontent.com/pact-foundation/pact-ruby-standalone/master/install.sh | bash
    response=$(curl -s -v https://github.com/pact-foundation/pact-ruby-standalone/releases/latest 2>&1)
    tag=$(echo "$response" | grep -o "Location: .*" | sed -e 's/[[:space:]]*$//' | grep -o "Location: .*" | grep -o '[^/]*$')
    version=${tag#v}
    os="linux-x86_64"
    curl -LO https://github.com/pact-foundation/pact-ruby-standalone/releases/download/${tag}/pact-${version}-${os}.tar.gz
    tar xzf pact-${version}-${os}.tar.gz
    rm pact-${version}-${os}.tar.gz

    log "Done!"
else
    log "Skipping installation of CLI tools, as they are present"
fi