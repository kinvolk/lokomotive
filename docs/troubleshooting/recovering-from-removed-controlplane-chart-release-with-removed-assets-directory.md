---
title: Recovering from removed controlplane chart release with removed assets directory
weight: 10
---

## Introduction

If you see the following error while running `lokoctl cluster apply` command:
```console
Controlplane component 'lokomotive' is missing, reinstalling...FATA[0086] Applying cluster failed: ensuring controlplane component "lokomotive": loading chart from assets: loading chart from asset directory "lokomotive-assets/cluster-assets/charts/lokomotive-system/lokomotive": stat lokomotive-assets/cluster-assets/charts/lokomotive-system/lokomotive: no such file or directory  args="[]" command="lokoctl cluster apply"
```

This means that the Helm release of some of the controlplane components has been uninstalled from your
cluster and `lokoctl` is not able to find a local copy of this controlplane component Helm chart in the
cluster assets directory.

As part of the `cluster apply` process, `lokoctl` ensures that the cluster is aligned with the last applied
configuration before the cluster is updated. This is done to ensure that the update process is consistent.

If a Helm release of some of the controlplane component is uninstalled from the cluster, `lokoctl cluster apply`
will re-install it before proceeding with the upgrade process to ensure that the cluster update is performed only
on a stable cluster. The re-installation process will be performed before the cluster assets directory is updated.

If you end up in a situation where the two conditions mentioned are met, `lokoctl cluster apply` has no way of knowing
which `lokoctl` version was used for the last cluster configuration, so it cannot guarantee that assets embedded in the
running binary will not update the cluster. To ensure safety, an error is returned in such cases.

Follow the steps below to resolve it.

### Checking which lokoctl version was used last time to manage the cluster

If you are sure that your local `lokoctl` binary is the same one which was used the last time you ran `lokoctl cluster apply`
successfully, you can run `lokoctl cluster apply` with the `--skip-pre-update-health-check` flag. This will skip the initial health
check and perform an upgrade from assets embedded in the binary, which should resolve the issue.

If you are not sure which version of `lokoctl` was used the last time, you can find it in the Terraform state.

To do that, first go to the cluster Terraform directory:

```sh
cd $ASSET_DIR/terraform
```

Where `$ASSET_DIR` matches the one defined in `asset_dir` in the `cluster` block.

Then, run the following command:

```sh
terraform state pull | grep lokoctl-version | uniq
```

It should print something similar to the following:

```console
$ terraform state pull | grep lokoctl-version | uniq
              "lokoctl-version": "v0.5.0"
```

Please note that `lokoctl-version` may not be available in Terraform state in all supported providers.

Then, you can compare the value with the output of `lokoctl version`.

### if version does not match

If the version of your local `lokoctl` binary does not match, get the right binary from the
[Releases](https://github.com/kinvolk/lokomotive/releases) page, then proceed with the steps below.

### If version matches

If the version matches, run the following command to fix your cluster:

```sh
lokoctl cluster apply --skip-pre-update-health-check
```

This will skip pre-update health checks, unpack embedded charts from the binary and run the upgrade process,
which should only re-install the missing releases.
