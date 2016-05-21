#!/bin/bash

set -ex

# Create go binary and package verifier + mock service into distribution
rm -rf output/*
rm -rf dist/*
go get github.com/mitchellh/gox
go get -d ./...

# if [ -n "$(type -t rvm)" ]; then
#   rvm use 2.1.5
# fi

###
### Darwin/OSX
###

rm -rf output/*
# gox -os="darwin" -arch="amd64" -output="output/{{.Dir}}_{{.OS}}_{{.Arch}}"

cd output
cp ../build/pact-provider-verifier-*/pkg/pact-provider-verifier-*osx* .
cp ../build/pact-mock_service-*/pkg/pact-mock-service-*osx*.tar.gz .

tar -xzf pact-provider-verifier-*osx*.tar.gz && rm pact-provider*.tar.gz && mv pact-provider-verifier* pact-provider-verifier
tar -xzf pact-mock-service-*osx*.tar.gz && rm pact-mock-service-*.tar.gz && mv pact-mock-service* pact-mock-service

rm -rf pact-provider*.tar.gz
rm -rf pact-mock*.tar.gz
tar -czf darwin-amd64.tar.gz * && mv darwin-amd64.tar.gz ../dist
cd ..


####
#### Windows 32bit
####

rm -rf output/*
gox -os="windows" -arch="386" -output="output/{{.Dir}}_{{.OS}}_{{.Arch}}"

cd output
cp ../build/pact-provider-verifier-*/pkg/pact-provider-verifier-*win32*.zip .
cp ../build/pact-mock_service-*/pkg/pact-mock-service-*win32*.zip .

unzip pact-provider-verifier-*win32*.zip && rm pact-provider*.zip && mv pact-provider-verifier* pact-provider-verifier
unzip pact-mock-service-*win32*.zip && rm pact-mock-service-*.zip && mv pact-mock-service* pact-mock-service

rm -rf pact-provider*.zip
rm -rf pact-mock*.zip
tar -czf windows-386.tar.gz *  && mv windows-386.tar.gz ../dist
cd ..

####
#### Linux 32bit
####

rm -rf output/*
gox -os="linux" -arch="386" -output="output/{{.Dir}}_{{.OS}}_{{.Arch}}"

cd output
cp ../build/pact-provider-verifier-*/pkg/pact-provider-verifier-*linux-x86.tar.gz .
cp ../build/pact-mock_service-*/pkg/pact-mock-service-*linux-x86.tar.gz .

tar -xzf pact-provider-verifier-*linux-x86.tar.gz && rm pact-provider*.tar.gz && mv pact-provider-verifier* pact-provider-verifier
tar -xzf pact-mock-service-*linux-x86.tar.gz && rm pact-mock-service-*.tar.gz && mv pact-mock-service* pact-mock-service

rm -rf pact-provider*.tar.gz
rm -rf pact-mock*.tar.gz
tar -czf linux-386.tar.gz * && mv linux-386.tar.gz ../dist
cd ..

####
#### Linux 64bit
####

rm -rf output/*
gox -os="linux" -arch="amd64" -output="output/{{.Dir}}_{{.OS}}_{{.Arch}}"

cd output
cp ../build/pact-provider-verifier-*/pkg/pact-provider-verifier-*linux-x86_64* .
cp ../build/pact-mock_service-*/pkg/pact-mock-service-*linux-x86_64*.tar.gz .

tar -xzf pact-provider-verifier-*linux-x86_64*.tar.gz && rm pact-provider*.tar.gz && mv pact-provider-verifier* pact-provider-verifier
tar -xzf pact-mock-service-*linux-x86_64*.tar.gz && rm pact-mock-service-*.tar.gz && mv pact-mock-service* pact-mock-service

rm -rf pact-provider*.tar.gz
rm -rf pact-mock*.tar.gz
tar -czf linux-amd64.tar.gz * && mv linux-amd64.tar.gz ../dist
cd ..
