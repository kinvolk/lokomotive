# Lokomotive Installer - General

This document includes information that is relevant for the Lokomotive installer in general,
regardless of the target platform. For platform-specific information, refer to the docs of the
relevant [platform](../../README.md#supported-platforms).

## Asset Directory

When running the `lokoctl cluster install` command, the installer generates files that are required
for the cluster bootstrap process, for example Terraform templates which are used to create the
cluster's infrastructure or TLS certificates used for securing communication between k8s components.
We refer to these files as *assets*.

Assets are generated in the **asset directory** which is specified using the `asset_dir`
configuration option in the `.lokocfg` file.

The asset directory contains the following subdirectories:

- `terraform/` - contains the Terraform [root module](https://www.terraform.io/docs/modules/index.html).
- `lokomotive-kubernetes/` - contains Terraform modules used by the root module to create a cluster.
- `cluster-assets/` - contains files generated during cluster bootstrap, e.g. k8s TLS certificates.
