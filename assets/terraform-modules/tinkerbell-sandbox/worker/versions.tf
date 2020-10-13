terraform {
  required_version = ">= 0.13"

  required_providers {
    tinkerbell = {
      source  = "tinkerbell/tinkerbell"
      version = "0.1.0"
    }
    libvirt = {
      source  = "dmacvicar/libvirt"
      version = "0.6.2"
    }
    random = {
      source  = "hashicorp/random"
      version = "3.0.0"
    }
  }
}
