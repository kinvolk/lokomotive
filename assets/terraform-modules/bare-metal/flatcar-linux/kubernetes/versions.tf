# Terraform version and plugin versions

terraform {
  required_version = ">= 0.12.0"

  required_providers {
    ct       = "0.6.1"
    local    = "1.4.0"
    null     = "2.1.2"
    template = "2.1.2"
    tls      = "2.2.0"
    matchbox = "0.4.1"
  }
}
