# Terraform version and plugin versions

terraform {
  required_version = ">= 0.12"
  required_providers {
    local    = "1.4.0"
    template = "2.1.2"
    tls      = "2.2.0"
    random   = "2.3.0"
  }
}
