#!/bin/bash

set -e

mkdir -p build
cd build

# Provider Verifier
if [ ! -d "pact-provider-verifier-${PACT_PROVIDER_VERIFIER_VERSION}" ]; then
  wget https://github.com/pact-foundation/pact-provider-verifier/archive/v${PACT_PROVIDER_VERIFIER_VERSION}.zip -O temp.zip
  unzip temp.zip
  rm temp.zip
  cd pact-provider-verifier-${PACT_PROVIDER_VERIFIER_VERSION}
  bundle
  bundle exec rake package
else
  echo "pact provider verifier already generated, run './scripts/clean.sh' to generate a new package"
fi

# Mock Service
if [ ! -d "pact-mock_service-${PACT_MOCK_SERVICE_VERSION}" ]; then
  wget https://github.com/bethesque/pact-mock_service/archive/v${PACT_MOCK_SERVICE_VERSION}.zip -O temp.zip
  unzip temp.zip
  rm temp.zip
  cd pact-mock_service-${PACT_MOCK_SERVICE_VERSION}
  bundle
  bundle exec rake package
else
  echo "pact mock service already generated, run './scripts/clean.sh' to generate a new package"
fi
