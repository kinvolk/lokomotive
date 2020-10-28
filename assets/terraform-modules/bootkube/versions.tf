# Terraform version and plugin versions

terraform {
  required_version = ">= 0.13"
  required_providers {
    local = {
      source  = "hashicorp/local"
      version = "2.0.0"
    }
    template = {
      source  = "hashicorp/template"
      version = "2.1.2"
    }
    tls = {
      source  = "hashicorp/tls"
      version = "2.2.0"
    }
  }
}
