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

The [aws credentials file](https://docs.aws.amazon.com/cli/latest/userguide/cli-chap-configure.html) can be found at `~/.aws/credentials` if you have set up and configured AWS CLI before, and want to use that account.

An alternative will be to create credentials file and add a valid AWS access key ID and secret access key for your IAM user e.g

```
[default]
aws_access_key_id=AKIAIOSFODNN7EXAMPLE
aws_secret_access_key=wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY
```

Create a `my-cluster.lokocfg` file to define your cluster and, optionally,
components that should be installed. Example:

```
variable "asset_dir" {
	type = "string"
}

variable "aws_creds" {
	type = "string"
}

variable "ssh_pubkey" {
	type = "string"
}

cluster "aws" {
	asset_dir = "${pathexpand(var.asset_dir)}"
	creds_path = "${pathexpand(var.aws_creds)}"
	cluster_name = "test"
	os_image = "flatcar-stable"
	dns_zone = "example.com"
	dns_zone_id = "XXX"
	ssh_pubkey = "${pathexpand(var.ssh_pubkey)}"
}

component "ingress-nginx" {
}
```

The maximal length for a cluster name is 18 characters.

Create a `lokocfg.vars` file and define all needed variables. Example:

```
asset_dir = "~/lokoctl-assets"
aws_creds = "~/.aws/credentials"
ssh_pubkey = "~/.ssh/id_rsa.pub"
```

Note that the asset directory should be kept for the lifetime of the cluster.
The path cannot be relative at the moment.

To apply the configuration, run

```
lokoctl cluster install
```

## Destroying the cluster

```bash
cd <asset_dir>/terraform/
terraform destroy
```
