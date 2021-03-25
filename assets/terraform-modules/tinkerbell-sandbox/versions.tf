terraform {
  required_version = ">= 0.13"

  required_providers {
    libvirt = {
      source  = "dmacvicar/libvirt"
      version = "0.6.2"
    }
    ct = {
      source  = "poseidon/ct"
      version = "0.8.0"
    }
  }
}
