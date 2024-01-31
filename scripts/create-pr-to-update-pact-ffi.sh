#!/bin/sh

set -e

: "${1?Please supply the pact-ffi version to upgrade to}"

FFI_VERSION=$1
TYPE=${2:-fix}
DASHERISED_VERSION=$(echo "${FFI_VERSION}" | sed 's/\./\-/g')
BRANCH_NAME="chore/upgrade-to-pact-ffi-${DASHERISED_VERSION}"

git checkout master
git checkout installer/installer.go
git pull origin master

git checkout -b ${BRANCH_NAME}

cat installer/installer.go | sed "s/version:.*/version:     \"${FFI_VERSION}\",/" > tmp-install
mv tmp-install installer/installer.go

git add installer/installer.go
git commit -m "${TYPE}: update pact-ffi to ${FFI_VERSION}"
git push --set-upstream origin ${BRANCH_NAME}

gh pr create --title "${TYPE}: update pact-ffi to ${FFI_VERSION}" --fill

git checkout master
