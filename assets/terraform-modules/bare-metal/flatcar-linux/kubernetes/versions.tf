# Terraform version and plugin versions

terraform {
  required_version = ">= 0.13"

  required_providers {
    ct = {
      source  = "poseidon/ct"
      version = "0.6.1"
    }
    null = {
      source  = "hashicorp/null"
      version = "3.0.0"
    }
    template = {
      source  = "hashicorp/template"
      version = "2.1.2"
    }
    matchbox = {
      source  = "poseidon/matchbox"
      version = "0.4.1"
    }
    random = {
      source  = "hashicorp/random"
      version = "2.3.0"
    }
  }
}
