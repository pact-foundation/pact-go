#!/bin/bash

set -e

mkdir -p build
cd build

PACT_STANDALONE_VERSION=1.30.0
urls=(https://github.com/pact-foundation/pact-ruby-standalone/releases/download/v$PACT_STANDALONE_VERSION/pact-$PACT_STANDALONE_VERSION-linux-x86.tar.gz https://github.com/pact-foundation/pact-ruby-standalone/releases/download/v$PACT_STANDALONE_VERSION/pact-$PACT_STANDALONE_VERSION-linux-x86_64.tar.gz https://github.com/pact-foundation/pact-ruby-standalone/releases/download/v$PACT_STANDALONE_VERSION/pact-$PACT_STANDALONE_VERSION-osx.tar.gz https://github.com/pact-foundation/pact-ruby-standalone/releases/download/v$PACT_STANDALONE_VERSION/pact-$PACT_STANDALONE_VERSION-win32.zip)

# Provider Verifier
echo "--> Downloading Ruby Engine"
if [ ! -f pact-$PACT_STANDALONE_VERSION-linux-x86.tar.gz ]; then
  for url in "${urls[@]}"
  do
    wget $url
  done
else
  echo "Pact CLI tools already downloaded, run 'make clean' to refresh"
fi