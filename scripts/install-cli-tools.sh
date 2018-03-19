#!/bin/bash +e

libDir=$(dirname "$0")
. "${libDir}/lib"

if [ ! -d "build/pact" ]; then
    step "Installing CLI tools locally"
    mkdir -p build/pact
    cd build
    curl -fsSL https://raw.githubusercontent.com/pact-foundation/pact-ruby-standalone/master/install.sh | bash
    log "Done!"
else
    log "Skipping installation of CLI tools, as they are present"
fi