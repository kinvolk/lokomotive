# Lokomotive release process

This document describes Lokomotive releasing process.

## Changelog


### Generating

This paragraph should describe how to generate/prepare changelog to include it in the release.


### Breaking changes

This paragraph should describe how to inform users about breaking changes.


## Versioning


### Versioning scheme

This paragraph should describe, that we use SemVer-like versioning etc. Perhaps it should point to the versioning document.


### Identifying latest release

The source of truth for available releases is [Git tags](https://git-scm.com/book/en/v2/Git-Basics-Tagging) in the repository.

To list available releases, run `git tags` command. It will list all available releases. Example output:

```console
$ git tag
v0.1.0
v0.1.1
```

In the following case, release `v0.1.1` is the latest release.


### Increasing version number

This paragraph should describe, how to decide, which version number should be increased.


#### Patch release

Patch release version should be increased, if the planned release:
* contain only bug fixes and improvements

That means, if the current latest version is `1.2.3`, it should be increased to `1.2.4`.


#### Minor release

Minor release version should be increased if the planned release:
* contains new features and does not include breaking changes
* include breaking changes, but it’s done before `1.0.0` release

That means, if the current latest version is `1.2.3`, it should be increased to `1.3.0`.


#### Major release

Major release version should be increased if the planned release:
* contains breaking changes
* contains major improvements

That means, if the current latest version is `1.2.3`, it should be increased to `2.0.0`.


## Release signing

Each Lokomotive release must be signed using trusted GPG key.


## Generating new GPG key

Please follow [https://help.github.com/en/github/authenticating-to-github/generating-a-new-gpg-key](https://help.github.com/en/github/authenticating-to-github/generating-a-new-gpg-key) for generating new keys for signing.


## Adding new GPG key to list of trusted keys

Before signing a release with new GPG key, it should be signed by other trusted keys and added to the list of trusted keys in the repository (insert link to some file in GitHub repository).


# Creating a release


## Checking latest release

First step in the release process is to identify what will be the next release version. See [Increasing version number](#increasing-version-number) for more details.


## Prepare GPG key

Before creating a release, you need to make sure that your GPG key is available for signing Git tags. Check [Release signing](#release-signing) for more details.


## Tagging a release

To tag a new release, first checkout the desired branch in the repository using the `git checkout` command.

Hint: before running the command, replace `<new release version>` with the desired release version, e.g. `v0.2.0`.

Once on the target branch, run the following command:

```sh
export RELEASE=v0.2.0; git tag --sign -m "Release $RELEASE" $RELEASE
```


## Pushing tag

Newly created tags should be pushed to remote repository. This can be done using the following command:

```sh
git push upstream <new release version>
```

Hint: before running the command, replace `<new release version>` with the desired release version, e.g. `v0.2.0`.


## Creating GitHub release and release artifacts

Once the release tag is published (pushed), we also need to create GitHub release and associated release artifacts. To automate this process, we use [goreleaser](https://goreleaser.com/) tool.


### Installing goreleaser

Please follow [this](https://goreleaser.com/install/) instruction to get it installed on your local machine.


### Getting GitHub API token

Gorelaser uses GitHub API to create release and upload release artifacts, so you need to have valid GitHub API token exported as an environment variable before running goreleser.

To create a new API token, visit [https://github.com/settings/tokens](https://github.com/settings/tokens).

To make it available for goreleaser, run the following command:

```sh
export GITHUB_TOKEN=<insert your GitHub token here>
```


### Running goreleaser

To automatically create release artifacts, create GitHub release and upload release artifacts, run the following command:

```sh
goreleaser
```


### Verifying release creation

Visit [https://github.com/kinvolk/lokomotive/releases](https://github.com/kinvolk/lokomotive/releases) to verify that a new release has been created successfully.


### Sample release process output

Here you can see how the sample release creation process should look like:

```console
$ git tag -a v0.1.1 -s -m "Release v0.1.1" # To tag a release.
$ git push origin v0.1.1 # To publish a tag.
Enumerating objects: 11, done.
Counting objects: 100% (11/11), done.
Delta compression using up to 12 threads
Compressing objects: 100% (7/7), done.
Writing objects: 100% (9/9), 2.15 KiB | 1.08 MiB/s, done.
Total 9 (delta 2), reused 0 (delta 0)
remote: Resolving deltas: 100% (2/2), completed with 1 local object.
To github.com:kinvolk/lokomotive-release-testing.git
 * [new tag]         v0.1.1 -> v0.1.1
$ export GITHUB_TOKEN=<GitHub token> # So goreleaser binary has access to it.
$ goreleaser # To build the binary and upload it to GitHub, create GitHub release etc.

   • releasing using goreleaser 0.128.0...
   • loading config file       file=.goreleaser.yml
   • RUNNING BEFORE HOOKS
   • LOADING ENVIRONMENT VARIABLES
   • GETTING AND VALIDATING GIT STATE
      • releasing v0.1.1, commit eb19beea4483a67cbacfb7a53a6331c77861ce29
   • PARSING TAG
   • SETTING DEFAULTS
      • LOADING ENVIRONMENT VARIABLES
      • SNAPSHOTING
      • GITHUB/GITLAB/GITEA RELEASES
      • PROJECT NAME
      • BUILDING BINARIES
      • ARCHIVES
      • LINUX PACKAGES WITH NFPM
      • SNAPCRAFT PACKAGES
      • CALCULATING CHECKSUMS
      • SIGNING ARTIFACTS
      • DOCKER IMAGES
      • ARTIFACTORY
      • BLOB
      • HOMEBREW TAP FORMULA
      • SCOOP MANIFEST
   • SNAPSHOTING
      • pipe skipped              error=not a snapshot
   • CHECKING ./DIST
   • WRITING EFFECTIVE CONFIG FILE
      • writing                   config=dist/config.yaml
   • GENERATING CHANGELOG
      • writing                   changelog=dist/CHANGELOG.md
   • BUILDING BINARIES
      • building                  binary=/home/invidian/repos/lokomotive-release-testing/dist/lokoctl_linux_386/lokoctl
      • building                  binary=/home/invidian/repos/lokomotive-release-testing/dist/lokoctl_linux_amd64/lokoctl
   • ARCHIVES
      • creating                  archive=dist/lokoctl_0.1.1_linux_386.tar.gz
      • creating                  archive=dist/lokoctl_0.1.1_linux_amd64.tar.gz
   • LINUX PACKAGES WITH NFPM
   • SNAPCRAFT PACKAGES
   • CALCULATING CHECKSUMS
      • checksumming              file=lokoctl_0.1.1_linux_amd64.tar.gz
      • checksumming              file=lokoctl_0.1.1_linux_386.tar.gz
   • SIGNING ARTIFACTS
      • signing                   cmd=[gpg --output dist/lokoctl_0.1.1_checksums.txt.sig --detach-sig dist/lokoctl_0.1.1_checksums.txt]
   • DOCKER IMAGES
      • pipe skipped              error=docker section is not configured
   • PUBLISHING
      • BLOB
         • pipe skipped              error=Blob section is not configured
      • HTTP UPLOAD
         • pipe skipped              error=uploads section is not configured
      • ARTIFACTORY
         • pipe skipped              error=artifactory section is not configured
      • DOCKER IMAGES
      • SNAPCRAFT PACKAGES
      • GITHUB/GITLAB/GITEA RELEASES
         • creating or updating release repo=kinvolk/lokomotive-release-testing tag=v0.1.1
         • release updated           url=https://github.com/kinvolk/lokomotive-release-testing/releases/tag/v0.1.1
         • uploading to release      file=dist/lokoctl_0.1.1_checksums.txt.sig name=lokoctl_0.1.1_checksums.txt.sig
         • uploading to release      file=dist/lokoctl_0.1.1_linux_amd64.tar.gz name=lokoctl_0.1.1_linux_amd64.tar.gz
         • uploading to release      file=dist/lokoctl_0.1.1_linux_386.tar.gz name=lokoctl_0.1.1_linux_386.tar.gz
         • uploading to release      file=dist/lokoctl_0.1.1_checksums.txt name=lokoctl_0.1.1_checksums.txt
      • HOMEBREW TAP FORMULA
      • SCOOP MANIFEST
         • pipe skipped              error=scoop section is not configured
   • release succeeded after 6.39s
```
