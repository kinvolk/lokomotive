# Terraform version and plugin versions

terraform {
  required_version = ">= 0.12.0"

  required_providers {
    ct       = "= 0.5.0"
    null     = "~> 2.1"
    template = "~> 2.1"
    libvirt  = "~> 0.6.0"
    random   = "~> 2.2"
  }
}
