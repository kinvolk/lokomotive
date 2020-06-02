#!/usr/bin/env bash
#
# Copyright 2020 Kinvolk
#
# Get the release notes text of the current checked out tag.

set -e

# make sure we are running in a toplevel directory
if ! [[ "$0" =~ "scripts/print-version-changelog" ]]; then
    echo "This script must be run in a toplevel Lokomotive directory"
    exit 255
fi

CURRENT_TAG=${CURRENT_TAG:-$(git describe --tags)}

if ! [[ "${CURRENT_TAG}" =~ ^v[[:digit:]]+.[[:digit:]]+.[[:digit:]]$ ]]; then
    echo "Not running on a tag"
    exit 255
fi

# filter text between "## ${CURRENT_TAG}" and the next tag in CHANGELOG.md
awk "/## ${CURRENT_TAG}/{flag=1;next}/## v[[:digit:]]+.[[:digit:]]+.[[:digit:]]+/{flag=0}flag" CHANGELOG.md |
    sed '1{/^[[:space:]]*$/d}' | # trim first line if empty
    sed '${/^$/d;}' # trim last line if empty
