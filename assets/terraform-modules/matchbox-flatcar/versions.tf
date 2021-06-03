terraform {
  required_version = ">= 0.13"

  required_providers {
    template = {
      source  = "hashicorp/template"
      version = "2.2.0"
    }
    matchbox = {
      source  = "poseidon/matchbox"
      version = "0.4.1"
    }
    random = {
      source  = "hashicorp/random"
      version = "3.0.0"
    }
  }
}
