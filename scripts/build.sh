#!/bin/bash

set -ex

# Create go binary and package verifier + mock service into distribution
VERSION=$(go version)
echo "==> Go version ${VERSION}"

echo "==> Getting dependencies..."
go get github.com/mitchellh/gox
# Fetch dependencies using `dep`
go get github.com/golang/dep/cmd/dep
dep ensure -v -vendor-only


echo "==> Creating binaries..."
gox -os="darwin" -arch="amd64" -output="build/pact-go_{{.OS}}_{{.Arch}}"
gox -os="windows" -arch="386" -output="build/pact-go_{{.OS}}_{{.Arch}}"
gox -os="linux" -arch="386" -output="build/pact-go_{{.OS}}_{{.Arch}}"
gox -os="linux" -arch="amd64" -output="build/pact-go_{{.OS}}_{{.Arch}}"

echo
echo "==> Results:"
ls -hl build/
