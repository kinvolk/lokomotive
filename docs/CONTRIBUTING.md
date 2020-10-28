# Contribution Guidelines

## Contents

- [Code of Conduct](#code-of-conduct)
- [Setup developer environment](#setup-developer-environment)
- [Build the code](#build-the-code)
- [Build with docker](#build-with-docker)
- [Update assets](#update-assets)
- [Authoring PRs](#authoring-prs)
  - [Commit Format](#commit-format)
- [Updating dependencies](#updating-dependencies)
- [Testing and linting requirements](#testing-and-linting-requirements)

## Code of Conduct

Please refer to the Kinvolk [Code of Conduct](https://github.com/kinvolk/contribution/blob/master/CODE_OF_CONDUCT.md).

## Setup developer environment

```bash
git clone git@github.com:kinvolk/lokomotive.git
cd lokomotive
```

## Build the code

```bash
make
```

To use the assets from disk instead of the ones embedded in the binary,
use the `LOKOCTL_USE_FS_ASSETS` environment variable.

Empty value means that lokoctl will search for assets in `assets`
directory where the binary is.
Non empty value should point to the `assets` directory.
The `assets` directory should contain subdirectories like `components`
and `terraform-modules`. Examples:

```bash
LOKOCTL_USE_FS_ASSETS='' ./lokoctl help
LOKOCTL_USE_FS_ASSETS='./assets' ./lokoctl help
```

## Build with docker

Alternatively, you can use a Docker environment to build the binary.

```bash
make build-in-docker
```

## Update assets

When changing code under `assets/` you need to regenerate assets before
contributing:

```bash
make update-assets
```

Commit and submit a PR.

## Authoring PRs

For the general guidelines on making PRs/commits easier to review, please check out
Kinvolk's
[contribution guidelines on git](https://github.com/kinvolk/contribution/tree/master/topics/git.md).

## Updating dependencies

In order to update dependencies managed with Go modules, run `make update`,
which will ensure that all steps needed for an update are taken (tidy and vendoring).

## Testing and linting requirements

- Minimum Go version **1.14**
- [golangci-lint](https://github.com/golangci/golangci-lint) installed locally
