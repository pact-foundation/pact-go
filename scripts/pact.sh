#!/bin/bash +e

############################################################
##        Start and stop the Pact Mock Server daemon      ## 
############################################################

LIBDIR=$(dirname "$0")
. "${LIBDIR}/lib"

trap shutdown INT
CUR_DIR=$(pwd)
function shutdown() {
    step "Shutting down stub server"
    log "Finding Pact daemon PID"
    PID=$(ps -ef | grep "pact-go daemon"| grep -v grep | awk -F" " '{print $2}' | head -n 1)
    if [ "${PID}" != "" ]; then
      log "Killing ${PID}"
      kill $PID
    fi
    cd $CUR_DIR
}

if [ ! -f "dist/pact-go" ]; then
    cd dist
    platform=$(detect_os)
    archive="${platform}-amd64.tar.gz"
    step "Installing Pact Go for ${platform}"

    if [ ! -f "${archive}" ]; then
      log "Cannot find distribution package ${archive}, please run 'make package' first"
      exit 1
    fi

    log "Expanding archive"
    if [[ $platform == 'linux' ]]; then
      tar -xf $archive
      mv pact-go_linux_amd64 pact-go
    elif [[ $platform == 'darwin' ]]; then
      tar -xf $archive
      mv pact-go_darwin_amd64 pact-go
    else
      log "Unsupported platform ${platform}"
      exit 1
    fi

    cd ..
    log "Done"
fi

step "Starting Daemon"
mkdir -p ./log
./dist/pact-go daemon -v -l DEBUG > log/daemon.log 2>&1 &

step "Running integration tests"
export PACT_INTEGRATED_TESTS=1
export PACT_BROKER_HOST="https://test.pact.dius.com.au"
export PACT_BROKER_USERNAME="dXfltyFMgNOFZAxr8io9wJ37iUpY42M"
export PACT_BROKER_PASSWORD="O5AIZWxelWbLvqMd8PkAVycBJh2Psyg1"
cd dsl
go test -v -run TestPact_Integration
SCRIPT_STATUS=$?
cd ..

shutdown

if [ "${SCRIPT_STATUS}" = "0" ]; then
  step "Integration testing succeeded!"
else
  step "Integration testing failed, see stack trace above"
fi

exit $SCRIPT_STATUS