locals {
  controller_bootstrap_token = var.enable_tls_bootstrap ? {
    token_id     = random_string.bootstrap_token_id_controller[0].result
    token_secret = random_string.bootstrap_token_secret_controller[0].result
  } : {}

  worker_bootstrap_token = var.enable_tls_bootstrap ? {
    token_id     = random_string.bootstrap_token_id_worker[0].result
    token_secret = random_string.bootstrap_token_secret_worker[0].result
  } : {}
}

# Generate a cryptographically random token id (public).
resource random_string "bootstrap_token_id_controller" {
  count = var.enable_tls_bootstrap == true ? 1 : 0

  length  = 6
  upper   = false
  special = false
}

# Generate a cryptographically random token secret.
resource random_string "bootstrap_token_secret_controller" {
  count = var.enable_tls_bootstrap == true ? 1 : 0

  length  = 16
  upper   = false
  special = false
}

# Generate a cryptographically random token id (public).
resource random_string "bootstrap_token_id_worker" {
  count = var.enable_tls_bootstrap == true ? 1 : 0

  length  = 6
  upper   = false
  special = false
}

# Generate a cryptographically random token secret.
resource random_string "bootstrap_token_secret_worker" {
  count = var.enable_tls_bootstrap == true ? 1 : 0

  length  = 16
  upper   = false
  special = false
}
