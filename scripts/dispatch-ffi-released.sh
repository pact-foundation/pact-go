#!/bin/sh

# Script to trigger an update of the pact ffi from pact-foundation/pact-reference to listening repos
# Requires a Github API token with repo scope stored in the
# environment variable GITHUB_ACCESS_TOKEN_FOR_PF_RELEASES

: "${GITHUB_ACCESS_TOKEN_FOR_PF_RELEASES:?Please set environment variable GITHUB_ACCESS_TOKEN_FOR_PF_RELEASES}"

if [ -n "$1" ]; then
  name="\"${1}\""
else
  echo "name not provided as first param"
  exit 1
fi

if [ -n "$2" ]; then
  version="\"${2}\""
else
  echo "name not provided as second param"
  exit 1
fi

repository_slug=$(git remote get-url origin | cut -d':' -f2 | sed 's/\.git//')

output=$(curl -v https://api.github.com/repos/${repository_slug}/dispatches \
      -H 'Accept: application/vnd.github.everest-preview+json' \
      -H "Authorization: Bearer $GITHUB_ACCESS_TOKEN_FOR_PF_RELEASES" \
      -d "{\"event_type\": \"pact-ffi-released\", \"client_payload\": {\"name\": ${name}, \"version\" : ${version}}}" 2>&1)

if  ! echo "${output}" | grep "HTTP\/.* 204" > /dev/null; then
  echo "$output" | sed  "s/${GITHUB_ACCESS_TOKEN_FOR_PF_RELEASES}/********/g"
  echo "Failed to trigger update"
  exit 1
else
  echo "Update workflow triggered"
fi

echo "See https://github.com/${repository_slug}/actions?query=workflow%3A%22Update+Pact+FFI+Library%22"