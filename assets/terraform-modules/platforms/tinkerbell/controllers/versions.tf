terraform {
  required_version = ">= 0.13"

  required_providers {
    null = {
      source  = "hashicorp/null"
      version = "3.1.0"
    }
    tinkerbell = {
      source  = "tinkerbell/tinkerbell"
      version = "0.1.0"
    }
  }
}
