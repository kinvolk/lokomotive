---
title: Developer docs
weight: 10
---

This is a developer docs directory. Not meant for end users.

## Spell checking

As part of the CI process, we run [codespell](https://github.com/codespell-project/codespell) to find potential
typos in our code base. This section describes how to configure and use it.

### Running locally

To run `codespell` locally, install it either using the [official method](https://github.com/codespell-project/codespell#installation)
or using a package manager specific to your Linux distribution.

With `codespell` installed, run the following command to run spell checking locally:
```sh
make codespell
```

### Ignoring specific words

Sometimes, `codespell` detects some words as typos, like `AKS`, `ACI` or `IAM`, which cause the CI to
report errors. To resolve that, we define our list of ignored words in the `.codespell.ignorewords` file,
which is then passed to `codespell` via the `--ignore-words` flag.

If you find that some word you used should be ignored, add it to this file.

### Ignoring vendored files

Lokomotive ships some 3rd party code like vendored Go dependencies and Helm charts, which are maintained
by other people and may also contain some typos. It is recommended to configure `codespell` to **SKIP** those
files from checking, instead of **FIXING** them in our repository, to avoid deriving from upstream code, which
may make the update process of this code more difficult in the future (merge conflicts, extensive diffs etc.).

To skip some file or directory (recursively) from spell checking, add the path to `.codespell.skip` file.
The content of this file is then passed to the `--skip` flag of `codespell`.

Additionally, if time allows, consider fixing the found typos in the respective upstream projects.

### Spell checking git commit messages

Spell checking of git commit messages is not done as part of the CI process, but it is recommended to do it locally
before submitting patches.

To check for typos in commit messages of your feature branch, you can run the following command:
```sh
git log master..HEAD | codespell -
```

It will spell check both commit messages and new code, so it is recommended to first fix typos in the code before
running it, to only find typos in commit messages.
