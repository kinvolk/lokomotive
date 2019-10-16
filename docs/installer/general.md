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

## Configuring Backend

Lokomotive installer supports local or remote backend [S3 only] for storing terraform state.
Lokomotive installer also supports optional state locking feature for S3 backend.

Backend configuration is OPTIONAL. If no backend configuration is provided then terraform's default will be used.

NOTE: Installer does not support multiple backends, configure only one.

### Requirements
* S3 bucket to be used should already be created.
* DynamoDB table to be used for state locking should already be created.
* Correct IAM permissions for the S3 bucket and DynamoDB Table. At minimum the following are the permissions required by terraform
  * [S3 bucket permissions](https://www.terraform.io/docs/backends/types/s3.html#s3-bucket-permissions) 
  * [DynamoDB table permissions](https://www.terraform.io/docs/backends/types/s3.html#dynamodb-table-permissions).

In order to to start using a backend, one needs to create a configuration in the `.lokocfg` file.


Examples

* Local Backend

  ```
  backend "local" {
    # Optional
    path = "terraform.tfstate"
  }

  ```
* S3 Backend
  
  ```
  backend "s3" {
    # Required parameters
    bucket = "<bucket_name>"
    key = "<path_in_s3_bucket>"
    # Optional parameters
    region = "<aws_region>"
    aws_creds_path = "<aws_credentials_file_path>" # ~/.aws/credentials
    dynamodb_table = "<dynamodb_table_name>"
  }
  ```
  NOTE: In order for the installer to configure the credentials for S3 backend either pass them as environment variables or in the config above.

  NOTE: If no value is passed for `dynamodb_table`, installer will not use the state locking feature.
