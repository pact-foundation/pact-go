#!/bin/bash

libDir=$(dirname "$0")
. "${libDir}/lib"

trap cleanup TERM

function cleanup() {
  log "ERROR generating release, please check your git logs and working tree to ensure things are in order."
}

step "Releasing Pact Go"

# Get current versions
log "Finding current version"
version=$(cat command/version.go | egrep -o "v([0-9\.]+)-?([a-zA-Z\-\+\.0-9]+)?")
lastVersion=$(git log  --grep='chore(release)' | grep chore | head -n1 | egrep -o "v([0-9\.]+)-?([a-zA-Z\-]+)?")
date=$(date "+%d %B %Y")

# Check tags
log "Checking if ${version} exists"
tagExists=$(git rev-parse --verify -q ${version})
if [ $? = 0 ]; then
  log "ERROR: tag already exists, this could break API compatibility, exiting."
  exit 1
fi

# Generate changelog
step "Generating changelog"
if [ $? = 0 ]; then
  log=$(git log --pretty=format:'  * [%h](https://github.com/pact-foundation/pact-go/commit/%h) - %s (%an, %ad)' ${lastVersion}..HEAD | egrep -v "wip(:|\()" | grep -v "docs(" | grep -v "chore(" | grep -v Merge | grep -v "test(")

  log "Updating CHANGELOG"
  ed CHANGELOG.md << END
7i

### $version ($date)
$log
.
w
q
END

  log "Changelog updated"
  step "Committing changes"
  git reset HEAD
  git add CHANGELOG.md
  git add command/version.go
  git commit -m "chore(release): release ${version}"

  step "Creating tag ${version}"
  git tag ${version} -m "chore(release): release ${version}"

  log "Done - check your git logs, and then run 'git push --follow-tags'."
else
  log "ERROR: Version ${version} does not exist, exiting."
  log "To fix this, ensure RELEASE_VERSION in the Wercker build is
        set to the correct tag (https://app.wercker.com/Pact-Foundation/pact-go/environment)"
  exit 1
fi

