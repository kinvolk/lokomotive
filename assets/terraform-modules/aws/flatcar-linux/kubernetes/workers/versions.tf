# Terraform version and plugin versions

terraform {
  required_version = ">= 0.12.0"

  required_providers {
    aws      = "3.3.0"
    ct       = "0.6.1"
    template = "2.1.2"
  }
}
