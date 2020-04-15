# Terraform version and plugin versions

terraform {
  required_version = ">= 0.12.0"

  required_providers {
    aws      = "2.48.0"
    ct       = "0.5.0"
    local    = "~> 1.2"
    null     = "~> 2.1"
    template = "~> 2.1"
    tls      = "~> 2.0"
  }
}
