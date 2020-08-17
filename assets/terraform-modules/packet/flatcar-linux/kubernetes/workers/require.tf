# Terraform version and plugin versions

terraform {
  required_version = ">= 0.12.0"

  required_providers {
    ct       = "= 0.5.0"
    packet   = "~> 2.7.3"
  }
}
