# Contribution Guidelines

## Setup developer environment

```bash
git clone git@github.com:kinvolk/lokomotive.git
cd lokomotive
```

## Requirements

- Minimum Go version **1.4**
- [golangci-lint](https://github.com/golangci/golangci-lint) installed locally

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

## Authoring PRs

These are general guidelines for making PRs/commits easier to review:

 * Commits should be atomic and self-contained. Divide logically separate changes
   to separate commits. This principle is best explained in the the Linux Kernel
   [submitting patches][linux-sep-changes] guide.

 * Commit messages should explain the intention, _why_ something is done. This,
   too, is best explained in [this section][linux-desc-changes] from the Linux
   Kernel patch submission guide.

 * Commit titles (the first line in a commit) should be meaningful and describe
   _what_ the commit does.

 * Don't add code you will change in a later commit (it makes it pointless to
   review that commit), nor create a commit to add code an earlier commit should
   have added. Consider squashing the relevant commits instead.

 * It's not important to retain your development history when contributing a
   change. Use `git rebase` to squash and order commits in a way that makes them easy to
   review. Keep the final, well-structured commits and not your development history
   that led to the final state.

 * Consider reviewing the changes yourself before opening a PR. It is likely
   you will catch errors when looking at your code critically and thus save the
   reviewers (and yourself) time.

 * Use the PR's description as a "cover letter" and give the context you think
   reviewers might need. Use the PR's description to explain why you are
   proposing the change, give an overview, raise questions about yet-unresolved
   issues in your PR, list TODO items etc.

PRs which follow these rules are easier to review and merge.

[linux-sep-changes]: https://www.kernel.org/doc/html/v4.17/process/submitting-patches.html#separate-your-changes
[linux-desc-changes]: https://www.kernel.org/doc/html/v4.17/process/submitting-patches.html#describe-your-changes

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
directory, e.g.: `cli/components`.

## Updating dependencies

In order to update dependencies managed with Go modules, run `make update`,
which will ensure that all steps needed for an update are taken (tidy and vendoring).
