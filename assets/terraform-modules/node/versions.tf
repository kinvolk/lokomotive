terraform {
  required_version = ">= 0.13"

  required_providers {
    ct = {
      source  = "poseidon/ct"
      version = "0.7.1"
    }
    random = {
      source  = "hashicorp/random"
      version = "3.0.0"
    }
  }
}
