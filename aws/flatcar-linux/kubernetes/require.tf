# Terraform version and plugin versions

terraform {
  required_version = ">= 0.11.0"
}

provider "aws" {
  version = "2.25.0"
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
