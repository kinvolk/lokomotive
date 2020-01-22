# Terraform version and plugin versions

terraform {
  required_version = ">= 0.12.0"

  required_providers {
    ct       = "~> 0.4"
    local    = "~> 1.2"
    null     = "~> 2.1"
    template = "~> 2.1"
    tls      = "~> 2.0"
    matchbox = "~> 0.3"
  }
}
