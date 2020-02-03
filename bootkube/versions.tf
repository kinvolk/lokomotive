# Terraform version and plugin versions

terraform {
  required_version = ">= 0.12"
  required_providers {
    local    = "~> 1.2"
    template = "~> 2.1"
    tls      = "~> 2.0"
  }
}
