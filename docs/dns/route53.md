# Route53 DNS provider

This module uses the AWS Route53 service to setup the DNS entries required for the cluster.

Include the following snippet in your configuration to use this module:

```tf
module "dns" {
  source = "git::https://github.com/kinvolk/lokomotive-kubernetes//dns/route53?ref=<hash>"

  entries = module.controller.dns_entries

  aws_zone_id = "<zone_id>" # e.g. Z3PAABBCFAKEC0
}
```

## AWS Credentials

Login to your AWS IAM dashboard and find your IAM user. Select "Security Credentials" and create an access key. Save the id and secret to a file that can be referenced in configs.

```
[default]
aws_access_key_id = xxx
aws_secret_access_key = yyy
```

!!! tip
    Other standard AWS authentication methods can be used instead of specifying `shared_credentials_file` under the provider's config. For more information see the [docs](https://www.terraform.io/docs/providers/aws/#authentication).

Configure the AWS provider to use your access key credentials in a `providers.tf` file.

```
provider "aws" {
  version = "2.31.0"
  alias   = "default"

  region                  = "eu-central-1"
  shared_credentials_file = "/home/user/.config/aws/credentials"
}
```

Additional configuration options are described in the `aws` provider [docs](https://www.terraform.io/docs/providers/aws/).

