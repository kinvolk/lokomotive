---
title: Configuration
weight: 10
---

## Overview

Lokomotive uses a [HCL2](https://github.com/hashicorp/hcl) based configuration language to allow
users to configure clusters and components. This configuration is read from `.lokocfg` files by
`lokoctl`.

Typically, configuration for a Lokomotive cluster consists of one or more `.lokocfg` files as well
as a `lokocfg.vars` file for variables and/or secret values. The details are explained below.

A configuration directory for a Lokomotive cluster for example could look like this:

```console
my-cluster/
├── cert-manager.lokocfg
├── cluster.lokocfg
├── dex.lokocfg
├── gangway.lokocfg
├── httpbin.lokocfg
├── ingress.lokocfg
└── lokocfg.vars
```

## Configuration through `.lokocfg` files

By default, lokoctl loads and merges configuration from all `.lokocfg` files in the current
directory. Depending on the issued command, different configuration blocks then get parsed and used.

For example, when running `lokoctl cluster apply`, lokoctl will look for a `cluster "<provider>" {
... }` block in all loaded lokocfg files, install or apply new configuration to the cluster and
afterwards proceed with installing or applying new configuration to all configured components, too.

Another example, when running `lokoctl component apply`, lokoctl will attempt to install or apply
new configuration to all configured components from all loaded lokocfg files. On the other hand,
`lokoctl component apply cert-manager` would only evaluate the `component "cert-manager" { ... }`
block and only install or apply new configuration to the cert-manager component.

With the `--lokocfg` command-line parameter, it is possible to load `.lokocfg` files from a
different directory or to load only a single file:

```console
lokoctl cluster apply --lokocfg path/to/my-cluster.lokocfg
```

## Variables and the `lokocfg.vars` file

It is possible to define variables for values that should be configurable or secrets in a
`lokocfg.vars` file.

The `lokocfg.vars` file is **not** meant to be stored in a source code repository.

For example, if you define a variables in a `.lokocfg` file as below:

```hcl
variable "github_client_id" {
  type = "string"
}

component "foo" {
  github_client_id = var.github_client_id
}
```

The corresponding value definition in `lokocfg.vars` would be:

```hcl
github_client_id = "aaabbbccc"
```

With the `--lokocfg-vars` command-line flag, you can specify the path to the `lokocfg.vars` file to
load.

## Interpolation functions

lokoctl supports the following interpolation functions when loading `.lokocfg` files and the
`lokocfg.vars` file.

`pathexpand`: expands a path with `~` in it.

Example:

```hcl
foo_path = pathexpand("~/foo")
```

`file`: reads the content of the passed file and returns it as string. Example:

```hcl
snippet = file("my-snippets/snippet.txt")
```
