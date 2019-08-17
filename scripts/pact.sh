#!/bin/bash +e

LIBDIR=$(dirname "$0")
. "${LIBDIR}/lib"

curDir=${pwd}
echo $curDir
trap shutdown INT
exitCode=0

function shutdown() {
    if [ "${exitCode}" != "0" ]; then
      log "Reviewing log output: "
      cat logs/*
    fi
}

if [ ! -d "build/pact" ]; then
    step "Installing CLI tools locally"
    mkdir -p build/pact
    cd build
    curl -fsSL https://raw.githubusercontent.com/pact-foundation/pact-ruby-standalone/master/install.sh | bash
    log "Done!"
fi

cd ${curDir}
export PACT_INTEGRATED_TESTS=1
export PACT_BROKER_HOST="https://test.pact.dius.com.au"
export PACT_BROKER_USERNAME="dXfltyFMgNOFZAxr8io9wJ37iUpY42M"
export PACT_BROKER_PASSWORD="O5AIZWxelWbLvqMd8PkAVycBJh2Psyg1"
export PATH="../build/pact/bin:${PATH}"

step "Running E2E regression and example projects"
examples=("github.com/pact-foundation/pact-go/examples/consumer/goconsumer" "github.com/pact-foundation/pact-go/examples/go-kit/provider" "github.com/pact-foundation/pact-go/examples/mux/provider" "github.com/pact-foundation/pact-go/examples/gin/provider" "github.com/pact-foundation/pact-go/examples/messages/consumer" "github.com/pact-foundation/pact-go/examples/messages/provider" "github.com/pact-foundation/pact-go/examples/httpbin")

for example in "${examples[@]}"
do
  log "Installing dependencies for example: $example"
  cd "${GOPATH}/src/${example}"
  go get ./...

  log "Running tests for $example"
  go test -v .
  if [ $? -ne 0 ]; then
    log "ERROR: Test failed, logging failure"
    exitCode=1
  fi
done
cd ..

shutdown

if [ "${exitCode}" = "0" ]; then
  step "Integration testing succeeded!"
else
  step "Integration testing failed, see stack trace above"
fi

exit $exitCode