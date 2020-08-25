# Terraform version and plugin versions

terraform {
  required_version = ">= 0.12.0"

  required_providers {
    ct       = "= 0.5.0"
    local    = "~> 1.2"
    null     = "~> 2.1"
    template = "~> 2.1"
    tls      = "~> 2.0"
    libvirt  = "~> 0.6.0"
    packet   = "~> 2.7.3"
    random   = "~> 2.2"
  }
}
