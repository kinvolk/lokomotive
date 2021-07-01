terraform {
  required_version = ">= 0.13"

  required_providers {
    ct = {
      source  = "poseidon/ct"
      version = "0.8.0"
    }
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "2.65.0"
    }
    template = {
      source  = "hashicorp/template"
      version = "2.2.0"
    }
  }
}
