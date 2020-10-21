# Terraform version and plugin versions

terraform {
  required_version = ">= 0.13"

  required_providers {

    ct = {
      source  = "poseidon/ct"
      version = "0.6.1"
    }

    libvirt = {
      source = "dmacvicar/libvirt"
      uri     = "qemu:///system"
      version = "0.6.2"
    }

    null     = "2.1.2"
    template = "2.1.2"
    random   = "2.3.0"
  }
}
