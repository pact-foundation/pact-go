#!/bin/bash -e
#
# Usage:
#   $ curl -fsSL https://raw.githubusercontent.com/pact-foundation/pact-plugins/master/install-cli.sh | bash
# or
#   $ wget -q https://raw.githubusercontent.com/pact-foundation/pact-plugins/master/install-cli.sh -O- | bash
#

set -x

function detect_osarch() {
    case $(uname -sm) in
        'Linux x86_64')
            os='linux'
            arch='x86_64'
            ;;
        'Darwin x86' | 'Darwin x86_64')
            os='osx'
            arch='x86_64'
            ;;
        'Darwin arm64')
            os='osx'
            arch='aarch64'
            ;;
        CYGWIN*|MINGW32*|MSYS*|MINGW*)
            os="windows"
            arch='x86_64'
            ext='.exe'
            ;;
        *)
        echo "Sorry, you'll need to install the plugin CLI manually."
        exit 1
            ;;
    esac
}

TAG=$(curl -s https://api.github.com/repos/pact-foundation/pact-plugins/releases | jq '.[0].tag_name')
VERSION=$(echo "$TAG" | tr -d '[:alpha:]-"')
detect_osarch

if [ ! -f ~/.pact/bin/pact-plugin-cli ]; then
    echo "--- 🐿  Installing plugins CLI version '${VERSION}' (from tag ${TAG})"
    mkdir -p ~/.pact/bin
    DOWNLOAD_LOCATION=https://github.com/pact-foundation/pact-plugins/releases/download/pact-plugin-cli-v${VERSION}/pact-plugin-cli-${os}-${arch}${ext}.gz
    echo "        Downloading from: ${DOWNLOAD_LOCATION}"
    curl -L -o ~/.pact/bin/pact-plugin-cli-${os}-${arch}.gz "${DOWNLOAD_LOCATION}"
    echo "        Downloaded $(file ~/.pact/bin/pact-plugin-cli-${os}-${arch}.gz)"
    gunzip -N -f ~/.pact/bin/pact-plugin-cli-${os}-${arch}.gz
    chmod +x ~/.pact/bin/pact-plugin-cli
fi