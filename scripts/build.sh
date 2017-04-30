#!/bin/bash

set -ex

# Create go binary and package verifier + mock service into distribution
VERSION=$(go version)
echo "==> Go version ${VERSION}"

echo "==> Getting dependencies..."
go get github.com/mitchellh/gox

echo "==> Creating binaries..."
gox -os="darwin" -arch="amd64" -output="build/pact-go_{{.OS}}_{{.Arch}}"
gox -os="windows" -arch="386" -output="build/pact-go_{{.OS}}_{{.Arch}}"
gox -os="linux" -arch="386" -output="build/pact-go_{{.OS}}_{{.Arch}}"
gox -os="linux" -arch="amd64" -output="build/pact-go_{{.OS}}_{{.Arch}}"

echo
echo "==> Results:"
ls -hl build/
