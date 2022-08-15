#!/bin/bash -e
#
# Usage:
#   $ curl -fsSL https://raw.githubusercontent.com/pact-foundation/pact-plugins/master/install-cli.sh | bash
# or
#   $ wget -q https://raw.githubusercontent.com/pact-foundation/pact-plugins/master/install-cli.sh -O- | bash
#

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
        *)
        echo "Sorry, you'll need to install the plugin CLI manually."
        exit 1
            ;;
    esac
}

package=$(curl https://api.github.com/repos/pact-foundation/pact-plugins/releases | grep pact-plugin-cli | grep tag_name | head -n1 | egrep -o "pact-plugin-cli-v.[0-9\.]+")
detect_osarch

if [ ! -f ~/.pact/bin/pact-plugin-cli ]; then
    echo "--- 🐿  Installing plugins CLI tool"
    mkdir -p ~/.pact/bin
    wget https://github.com/pact-foundation/pact-plugins/releases/download/${package}/pact-plugin-cli-${os}-${arch}.gz -O ~/.pact/bin/pact-plugin-cli-${os}-${arch}.gz
    gunzip -N ~/.pact/bin/pact-plugin-cli-${os}-${arch}.gz
    chmod +x ~/.pact/bin/pact-plugin-cli
fi