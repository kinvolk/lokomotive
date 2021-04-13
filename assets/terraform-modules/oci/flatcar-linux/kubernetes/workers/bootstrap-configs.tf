locals {
  worker_bootstrap_token = var.enable_tls_bootstrap ? {
    token_id     = random_string.bootstrap_token_id[0].result
    token_secret = random_string.bootstrap_token_secret[0].result
  } : {}
}

# Generate a cryptographically random token id (public).
resource "random_string" "bootstrap_token_id" {
  count = var.enable_tls_bootstrap == true ? 1 : 0

  length  = 6
  upper   = false
  special = false
}

# Generate a cryptographically random token secret.
resource "random_string" "bootstrap_token_secret" {
  count = var.enable_tls_bootstrap == true ? 1 : 0

  length  = 16
  upper   = false
  special = false
}
