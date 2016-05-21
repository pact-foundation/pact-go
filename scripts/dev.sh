#!/bin/bash -e

set -e

./scripts/package.sh

# Setup dev
echo "==> Creating OS distributions..."
tar -xzf dist/darwin-amd64.tar.gz -C ..
