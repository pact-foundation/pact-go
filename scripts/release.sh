#!/bin/bash

# Get current versions
echo "==> Finding current version"
version=$(cat command/version.go | egrep -o "v[0-9\.]+")
date=$(date "+%d %B %Y")

# Check tags
echo "==> Checking if ${version} exists"
git rev-parse ${version}

if [ $? = 0 ]; then
  log=$(git log --pretty=format:'  * [%h](https://github.com/pact-foundation/pact-go/commit/%h) - %s (%an, %ad)' ${version}..HEAD | egrep -v "wip(:|\()" | grep -v "docs(" | grep -v "chore(" | grep -v Merge | grep -v "test(")

  echo "==> Updating CHANGELOG"
  ed CHANGELOG.md << END
7i

### $version ($date)
$log
.
w
q
END

  echo "==> Changelog updated"
  echo "==> Committing changes"
  git reset HEAD
  git add CHANGELOG.md
  git commit -m "chore(release): release ${version}"

  echo "==> Done - check your git logs, and then `git push`!"
else
  echo "ERROR: Version ${version} does not exist, exiting."
  echo "To fix this, ensure RELEASE_VERSION in the Wercker build is 
        set to the correct tag (https://app.wercker.com/Pact-Foundation/pact-go/environment)"
  exit 1
fi

