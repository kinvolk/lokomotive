## Release process

This section shows how to perform a release of Lokomotive. Only parts of the
procedure are automated; this is somewhat intentional (manual steps for sanity
checking) but it can probably be further scripted.

We use goreleaser to automate some parts of the release, follow [the
instructions](https://goreleaser.com/install/) to get it installed on your
local machine.

The following example assumes we’re going from version 0.1.0 (`v0.1.0`) to
0.2.0 (`v0.2.0`). Check [Increasing version number](#increasing-version-number)
for details on how to identify what the next version should be.

- Start at the relevant milestone on GitHub (e.g.
  https://github.com/kinvolk/lokomotive/milestones/v0.2.0): ensure all
  referenced issues are closed or moved elsewhere. Close the milestone.

- Create a release branch from latest master
  - e.g. `git fetch origin && git checkout -b user/release-v0.2.0 origin/master`

- Make sure your git status is clean.

- Ensure the build is clean
  - `git clean -ffdx && make all` should work.
  - CI should be green.

- Update the [release notes](../CHANGELOG.md). Try to capture most of the salient
  changes since the last release, but don't go into unnecessary detail (better
  to link/reference the documentation wherever possible).
  `scripts/changelog.sh` will help generating an initial list of changes.
  Correct/fix entries if necessary, and group them by category.

Even though it is set at build time, the Lokomotive version is also hardcoded
in the repository, so the first thing to do is bump it:

- Run `scripts/bump-release.sh v0.2.0`. This should generate two commits: a bump
  to the actual release (e.g. v0.2.0, including CHANGELOG updates), and then a
  bump to the release+git (e.g. v0.2.0+git). The actual release version should
  only exist in a single commit! Sanity check what the script did with `git
  diff HEAD^^` or similar.

- If the script didn't work, yell at the author and/or fix it. It can almost certainly be improved.

- File a PR and get a review from another maintainer. This is useful to a)
  sanity check the diff, and b) be very explicit/public that a release is
  happening.

- Ensure the CI on the release PR is green!

- Merge the PR.

Now we'll tag the release.

- Check out the release commit: `git checkout HEAD^` should work. You want to
  be at the commit where the version is without "+git". Sanity check
  `pkg/version/version.go`.

- Create a signed tag. Check [Release signing](#release-signing) for details.
  - `git tag -a v0.2.0 -s -m "Release v0.2.0"`

- Push the tag to git.
  - `git push --tags origin v0.2.0`

- Export your GitHub token (check [Getting a GitHub API token](#getting-a-github-api-token) for details).
  - `export GITHUB_TOKEN=<GitHub token>`

- Build the binary, sign it, upload it to GitHub, create draft GitHub release.
  - `make build-and-publish-release`

- Go to the [releases page](https://github.com/kinvolk/lokomotive/releases) and
  check everything looks good.

- Use the GitHub UI to publish the release.

- Clean your git tree.
  - `git clean -ffdx`

## Increasing version number

This attempts to describe how to decide what kind of release to do.

### Patch release

Patch release version should be increased, if the planned release:

- contains only bug fixes and improvements

That means, if the current latest version is `1.2.3`, it should be increased to `1.2.4`.

### Minor release

Minor release version should be increased if the planned release either:

- contains new features and does not include breaking changes

- includes breaking changes, but it’s done before `1.0.0` release

That means, if the current latest version is `1.2.3`, it should be increased to `1.3.0`.

### Major release

Major release version should be increased if the planned release either:

- contains breaking changes

- contains major improvements

That means, if the current latest version is `1.2.3`, it should be increased to `2.0.0`.

## Getting a GitHub API token

goreleaser uses the GitHub API to create a release and upload release
artifacts, so you need to have valid GitHub API token exported as an
environment variable before running it.

To create a new API token, visit
[https://github.com/settings/tokens](https://github.com/settings/tokens).

## Release signing

Each Lokomotive release must be signed by a trusted GPG key.

### Generating new GPG key

Please follow [Generating a new GPG
key](https://help.github.com/en/github/authenticating-to-github/generating-a-new-gpg-key)
for generating new keys for signing.

## Adding new GPG key to list of trusted keys

Before signing a release with a new GPG key, it should be signed by other trusted
keys and added to the [list of trusted keys in the repository](KEYS.md).
