#!/bin/bash

set -e

mkdir -p build
cd build

PACT_STANDALONE_VERSION=1.4.2
urls=(https://github.com/pact-foundation/pact-ruby-standalone/releases/download/v$PACT_STANDALONE_VERSION/pact-$PACT_STANDALONE_VERSION-linux-x86.tar.gz https://github.com/pact-foundation/pact-ruby-standalone/releases/download/v$PACT_STANDALONE_VERSION/pact-$PACT_STANDALONE_VERSION-linux-x86_64.tar.gz https://github.com/pact-foundation/pact-ruby-standalone/releases/download/v$PACT_STANDALONE_VERSION/pact-$PACT_STANDALONE_VERSION-osx.tar.gz https://github.com/pact-foundation/pact-ruby-standalone/releases/download/v$PACT_STANDALONE_VERSION/pact-$PACT_STANDALONE_VERSION-win32.zip)

# Provider Verifier
echo "--> Downloading Ruby Engine"
if [ ! -f pact-$PACT_STANDALONE_VERSION-linux-x86.tar.gz ]; then
  for url in "${urls[@]}"
  do
    wget $url
  done
else
  echo "pact provider verifier already generated, run 'make clean' to generate a new package"
fi

osarchs=(osx win32 linux-x86 linux-x86_64)

echo "--> Packaging distributions"
mkdir -p ../dist
for os in "${osarchs[@]}" 
do
  echo "Building ${os}"
  osarch=""
  if [ "${os}" = "win32" ]; then
    osarch="windows_386"
    cp pact-go_$osarch.exe pact-go
    zip -u pact-$PACT_STANDALONE_VERSION-$os.zip pact-go
    cp pact-$PACT_STANDALONE_VERSION-$os.zip ../dist/pact-go_$osarch.zip
  else
    if [ "${os}" = "osx" ]; then
      osarch="darwin_amd64"
      cp pact-go_$osarch pact-go
    elif [ "${os}" = "linux-x86" ]; then
      osarch="linux_386"
      cp pact-go_$osarch pact-go
    elif [ "${os}" = "linux-x86_64" ]; then
      osarch="linux_amd64"
      cp pact-go_$osarch pact-go
    fi

    gunzip pact-$PACT_STANDALONE_VERSION-$os.tar.gz
    tar -rf pact-$PACT_STANDALONE_VERSION-$os.tar pact-go
    gzip pact-$PACT_STANDALONE_VERSION-$os.tar
    cp pact-$PACT_STANDALONE_VERSION-$os.tar.gz ../dist/pact-go_$osarch.tar.gz
  fi
done

 