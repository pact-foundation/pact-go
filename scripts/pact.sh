#!/bin/bash +e

############################################################
##        Start and stop the Pact Mock Server daemon      ## 
############################################################

LIBDIR=$(dirname "$0")
. "${LIBDIR}/lib"

trap shutdown INT
CUR_DIR=$(pwd)
exitCode=0

function shutdown() {
    step "Shutting down stub server"
    log "Finding Pact daemon PID"
    PID=$(ps -ef | grep "pact-go daemon"| grep -v grep | awk -F" " '{print $2}' | head -n 1)
    if [ "${PID}" != "" ]; then
      log "Killing ${PID}"
      kill $PID
    fi
    cd $CUR_DIR
    
    if [ "${exitCode}" != "0" ]; then
      log "Reviewing log output: "
      cat logs/*
    fi
}

if [ ! -f "dist/pact-go" ]; then
    cd dist
    platform=$(detect_os)
    archive="pact-go_${platform}_amd64.tar.gz"
    step "Installing Pact Go for ${platform}"

    if [ ! -f "${archive}" ]; then
      log "Cannot find distribution package ${archive}, please run 'make package' first"
      exit 1
    fi

    log "Expanding archive"
    if [[ $platform == 'linux' ]]; then
      tar -xf $archive
    elif [[ $platform == 'darwin' ]]; then
      tar -xf $archive
    else
      log "Unsupported platform ${platform}"
      exit 1
    fi

    log "Done!"
fi

step "Starting Daemon"
mkdir -p ./logs
./dist/pact-go daemon -v -l DEBUG > logs/daemon.log 2>&1 &

export PACT_INTEGRATED_TESTS=1
export PACT_BROKER_HOST="https://test.pact.dius.com.au"
export PACT_BROKER_USERNAME="dXfltyFMgNOFZAxr8io9wJ37iUpY42M"
export PACT_BROKER_PASSWORD="O5AIZWxelWbLvqMd8PkAVycBJh2Psyg1"

step "Running E2E regression and example projects"
examples=("github.com/pact-foundation/pact-go/examples/consumer/goconsumer" "github.com/pact-foundation/pact-go/examples/go-kit/provider" "github.com/pact-foundation/pact-go/examples/mux/provider" "github.com/pact-foundation/pact-go/examples/gin/provider")

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