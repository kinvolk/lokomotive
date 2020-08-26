terraform {
  required_version = ">= 0.13"

  required_providers {
    ct = {
      source = "poseidon/ct"
      version = "0.6.0"
    }
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "1.35.0"
    }
    template = {
      source  = "hashicorp/template"
      version = "2.1.2"
    }
  }
}
