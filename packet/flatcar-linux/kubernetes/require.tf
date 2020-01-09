# Terraform version and plugin versions

terraform {
  required_version = ">= 0.12.0"
}

provider "ct" {
  version = "0.4.0"
}

provider "local" {
  version = "~> 1.2"
}

provider "null" {
  version = "~> 2.1"
}

provider "template" {
  version = "~> 2.1"
}

provider "tls" {
  version = "~> 2.0"
}

provider "packet" {
  version = "~> 2.7.3"
}

provider "aws" {
  version = "2.31.0"
}
