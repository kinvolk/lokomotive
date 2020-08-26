# Terraform version and plugin versions

terraform {
  required_version = ">= 0.13"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "3.3.0"
    }
    ct = {
      source  = "poseidon/ct"
      version = "0.6.1"
    }
    null = {
      source  = "hashicorp/null"
      version = "2.1.2"
    }
    template = {
      source  = "hashicorp/template"
      version = "2.1.2"
    }
    random = {
      source  = "hashicorp/random"
      version = "2.3.0"
    }
  }
}
