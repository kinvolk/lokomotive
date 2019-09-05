# Lokokomotive AWS installation guide

This guide walks through the installation of Lokomotive on AWS.

## Requirements

* AWS Account and IAM credentials
* AWS Route53 DNS Zone (registered Domain Name or delegated subdomain)
* [Terraform v0.11.x](https://www.terraform.io/downloads.html)
* [Terraform-provider-ct](https://github.com/coreos/terraform-provider-ct) installed locally
    ```bash
    wget https://github.com/poseidon/terraform-provider-ct/releases/download/v0.3.1/terraform-provider-ct-v0.3.1-linux-amd64.tar.gz
    tar xzf terraform-provider-ct-v0.3.1-linux-amd64.tar.gz
    mkdir -p ~/.terraform.d/plugins
    mv terraform-provider-ct-v0.3.1-linux-amd64/terraform-provider-ct ~/.terraform.d/plugins/terraform-provider-ct_v0.3.1
    ```

## Installing the cluster

The [aws credentials file](https://docs.aws.amazon.com/cli/latest/userguide/cli-chap-configure.html) can be found at `~/.aws/credentials` if you have set up and configured AWS CLI before.
If you want to use that account, you don't need to specify any credentials for lokoctl.

You can also take any other credentials mechanism used by the AWS CLI but [environment variables](https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-envvars.html)
may be the safest option. Either prepend them when starting `lokoctl` or export each of them once in the current terminal session:

```
$ AWS_ACCESS_KEY_ID=abc AWS_SECRET_ACCESS_KEY=xyz lokoctl ...
```

If you want to use a credentials file other than the default, add a valid AWS access key ID and secret access key for your IAM user, e.g:

```
[default]
aws_access_key_id=AKIAIOSFODNN7EXAMPLE
aws_secret_access_key=wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY
```

You can specify to use this file by setting the `AWS_SHARED_CREDENTIALS_FILE` environment variable or the `aws_creds`/`creds_path` variable in the following cluster configuration.

Create a `my-cluster.lokocfg` file to define your cluster and, optionally,
components that should be installed. Example:

```
variable "asset_dir" {
	type = "string"
}

#variable "aws_creds" {
#	type = "string"
#}

variable "ssh_pubkey" {
	type = "string"
}

cluster "aws" {
	asset_dir = "${pathexpand(var.asset_dir)}"
	# creds_path = "${pathexpand(var.aws_creds)}"
	cluster_name = "test"
	os_image = "flatcar-stable"
	dns_zone = "example.com"
	dns_zone_id = "XXX"
	ssh_pubkey = "${pathexpand(var.ssh_pubkey)}"
}

component "contour" {
  install_mode = "daemonset"
}
```

The maximal length for a cluster name is 18 characters.

Create a `lokocfg.vars` file and define all needed variables. Example:

```
asset_dir = "~/lokoctl-assets/mycluster"
#aws_creds = "~/.aws/credentials"
ssh_pubkey = "~/.ssh/id_rsa.pub"
```

>NOTE: The asset directory should be kept for the lifetime of the cluster. For more information
>regarding the asset directory, see [here](general.md#asset-directory).

To apply the configuration, run

```
lokoctl cluster install
```

## Destroying the cluster

```bash
cd <asset_dir>/terraform/
terraform destroy
```
