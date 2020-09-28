# Terraform version and plugin versions

terraform {
  required_version = ">= 0.12.0"

  required_providers {
    ct       = "0.6.0"
    azurerm  = "1.35.0"
    null     = "2.1.2"
    template = "2.1.2"
  }
}
