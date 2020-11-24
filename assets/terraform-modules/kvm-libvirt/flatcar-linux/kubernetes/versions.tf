# Terraform version and plugin versions

terraform {
  required_version = ">= 0.12.0"

  required_providers {
    ct       = "0.7.1"
    null     = "3.0.0"
    template = "2.2.0"
    libvirt  = "0.6.0"
    random   = "3.0.0"
  }
}
