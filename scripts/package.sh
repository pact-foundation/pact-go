#!/bin/bash -e

set -e

echo "==> Building App"
make bin

# Create the OS specific versions of the mock service and verifier
echo "==> Building Ruby Binaries..."
scripts/build_standalone_packages.sh

echo
echo "==> Results:"
ls -hl dist/
