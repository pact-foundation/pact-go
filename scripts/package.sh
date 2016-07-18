#!/bin/bash -e

set -e

export PACT_MOCK_SERVICE_VERSION=0.8.2-golang # Forked mock service with pinned deps.
export PACT_PROVIDER_VERIFIER_VERSION=0.0.12

# Create the OS specific versions of the mock service and verifier
echo "==> Building Ruby Binaries..."
scripts/build_standalone_packages.sh

# Build each go package for specific OS, bundling the mock service and verifier
echo "==> Creating OS distributions..."
scripts/compile.sh

echo
echo "==> Results:"
ls -hl dist/
