# Terraform version and plugin versions

terraform {
  required_version = ">= 0.13"

  required_providers {
    ct = {
      source  = "poseidon/ct"
      version = "0.6.1"
    }
    local = {
      source  = "hashicorp/local"
      version = "2.0.0"
    }
    null = {
      source  = "hashicorp/null"
      version = "3.0.0"
    }
    packet = {
      source  = "packethost/packet"
      version = "3.0.1"
    }
    random = {
      source  = "hashicorp/random"
      version = "3.0.0"
    }
  }
}
