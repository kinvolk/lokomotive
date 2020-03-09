# Contribution Guidelines

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
and `lokomotive-kubernetes`. Examples:

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

## Commit guidelines

The title of the commit message should describe the _what_ about the
changes. Additionally, it is helpful to add the _why_ in the body of
the commit changes. Make sure to include any assumptions that you
might have made in this commit. Changes to unrelated parts of the
codebase should be kept as separate commits.

### Commit Format

```
<area>: <description of changes>

Detailed information about the commit message goes here
```

The title should not exceed 80 chars, although keeping it under 72
chars is appreciated.

Please wrap the body of commit message at a
maximum of 80 chars.

Here are a few example commit messages:

Good:
```
components/cert-manager: update manifest to 0.2


Upstream charts for cert-manager has been released to 0.2. This commit
updates the component to use the latest upstream charts.
```

Bad:
```
Update manifest of cert-manager to 0.2
```


Acceptable:
```
cert-manager: update manifest to 0.2
```

This format is acceptable as sometimes nesting parts of the codebase
in the title can take up a lot of characters. Also at the same time,
using `pkg/components/` is redundant in the title, unless there is
another directory of the same name but with a different parent
directory, for eg: `cli/components`.

## Updating dependencies

In order to update dependencies managed with Go modules, run `make update`,
which will ensure that all steps needed for an update are taken (tidy and vendoring).
