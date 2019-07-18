# Contribution Guidelines

## Setup developer environment

```bash
git clone git@github.com:kinvolk/lokoctl.git
cd lokoctl
```

## Build the code

```bash
make
```

To build a "dev" version of lokoctl, use

```
make update-lk-submodule
make build-slim
```

The resulting binary won't include the Lokomotive Kubernetes assets and
requires the lokomotive-kubernetes code in the submodule directory.


## Build with docker

Alternatively, you can use Docker environment to build the binary.

```bash
docker build .
```

## Update the lokomotive-kubernetes to current master

To update the local git submodule `lokomotive-kubernetes` dir to the latest master run following commands:

```bash
cd lokomotive-kubernetes
git pull --ff-only origin master
cd ..
```

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
components/cert-manager: Update manifest to 0.2


Upstream charts for cert-manager has been released to 0.2. This commit
updates the component to use the latest upstream charts.
```

Bad:
```
Update manifest of cert-manager to 0.2
```


Acceptable:
```
cert-manager: Update manifest to 0.2
```

This format is acceptable as sometimes nesting parts of the codebase
in the title can take up a lot of characters. Also at the same time,
using `pkg/components/` is redundant in the title, unless there is
another directory of the same name but with a different parent
directory, for eg: `cli/components`.
