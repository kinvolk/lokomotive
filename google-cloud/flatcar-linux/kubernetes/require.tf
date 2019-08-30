# Terraform version and plugin versions

terraform {
  required_version = ">= 0.11.0"
}

provider "google" {
  version = ">= 1.19, < 3.0"
}

provider "local" {
  version = "~> 1.0"
}

provider "null" {
  version = "~> 2.1"
}

provider "template" {
  version = "~> 2.1"
}

provider "tls" {
  version = "~> 1.0"
}
