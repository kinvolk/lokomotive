# Terraform version and plugin versions

terraform {
  required_version = ">= 0.11.0"
}

provider "ct" {
  version = "~> 0.3"
}

provider "local" {
  version = "~> 1.0"
}

provider "template" {
  version = "~> 1.0"
}

provider "tls" {
  version = "~> 1.0"
}

provider "libvirt" {
  version = "~> 0.5.2"
  uri = "qemu:///system"
}

