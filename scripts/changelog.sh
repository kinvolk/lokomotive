#!/usr/bin/env bash
#
# Copyright 2017 The rkt Authors
# Copyright 2020 Kinvolk
#
# Create a changelog.
#
# The env variable RANGE specifies the range of commits to be searched for the changelog.
# If unset the latest tag until origin/master will be set.
#
# The env variable GITHUB_TOKEN can specify a GitHub personal access token.
# Otherwise one could run into GitHub rate limits. Go to
# https://github.com/settings/tokens to generate a token.

set -e

jq --version >/dev/null 2>&1 || {
    echo "could not find jq (JSON command line processor), is it installed?"
    exit 255
}

if [ -z "${RANGE}" ]; then
    LATEST_TAG=$(git describe --tags --abbrev=0)
    RANGE="${LATEST_TAG}..origin/master"
fi

if [ ! -z "${GITHUB_TOKEN}" ]; then
    GITHUB_AUTH="--header \"authorization: Bearer ${GITHUB_TOKEN}\""
fi

for pr in $(git log --pretty=%s --first-parent "${RANGE}" | egrep -o '#\w+' | tr -d '#'); do
    body=$(curl -s ${GITHUB_AUTH} https://api.github.com/repos/kinvolk/lokomotive/pulls/"${pr}" | \
                  jq -r '{title: .title, body: .body}')

    echo "-" \
         "$(echo "${body}" | jq -r .title | sed 's/\.$//g')" \
         "([#${pr}](https://github.com/kinvolk/lokomotive/pull/$pr))." \
         "$(echo "${body}" | jq -r .body | awk -v RS='\r\n\r\n' NR==1 | tr -d '\r')"
done
