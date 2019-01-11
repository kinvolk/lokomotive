# Lokoctl AWS installation
This guide walks through the installation of Lokomotive on AWS.

## Requirements
* AWS Account and IAM credentials
* AWS Route53 DNS Zone (registered Domain Name or delegated subdomain)
* [Terraform v0.11.x](https://www.terraform.io/downloads.html)
* [Terraform-provider-ct](https://github.com/coreos/terraform-provider-ct) installed locally
    ```bash
    wget https://github.com/coreos/terraform-provider-ct/releases/download/v0.3.0/terraform-provider-ct-v0.3.0-linux-amd64.tar.gz
    tar xzf terraform-provider-ct-v0.3.0-linux-amd64.tar.gz
    mv terraform-provider-ct-v0.3.0-linux-amd64/terraform-provider-ct ~/.terraform.d/plugins/terraform-provider-ct_v0.3.0
    ```

## Lokomotive installer
Get [Lokoctl](https://github.com/kinvolk/lokoctl) and build with `make` in the project root.

Run `./lokoctl install aws` with the required flags added appropriately e.g
``` bash
./lokoctl install aws \
    --cluster-name <insert-your-cluster-name> \
    --dns-zone <insert-your-dns-zone> \
    --dns-zone-id <insert-your-zone-id> \
    --creds <path-to-aws-credentials-file>
```

The [aws credentials file](https://docs.aws.amazon.com/cli/latest/userguide/cli-chap-configure.html) can be found at `~/.aws/credentials` if you have set up and configured AWS CLI, and want to use that.

An alternative will be to create credentials file and add a valid AWS access key ID and secret access key for your IAM user e.g
```
[default]
aws_access_key_id=AKIAIOSFODNN7EXAMPLE
aws_secret_access_key=wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY
```

## Clean up
```bash
cd ~/.lokoctl/cluster-name.your-dns-zone/terraform/ # expected default if --assets flag was not set in the lokoctl install command
# else
# cd ~/.lokoctl/value-of-assets/terraform/
terraform destroy
```
