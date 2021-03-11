locals {
  controller_bootstrap_token = {
    token_id     = random_string.bootstrap_token_id_controller.result
    token_secret = random_string.bootstrap_token_secret_controller.result
  }

  worker_bootstrap_token = {
    token_id     = random_string.bootstrap_token_id_worker.result
    token_secret = random_string.bootstrap_token_secret_worker.result
  }
}

# Generate a cryptographically random token id (public).
resource "random_string" "bootstrap_token_id_controller" {
  length  = 6
  upper   = false
  special = false
}

# Generate a cryptographically random token secret.
resource "random_string" "bootstrap_token_secret_controller" {
  length  = 16
  upper   = false
  special = false
}

# Generate a cryptographically random token id (public).
resource "random_string" "bootstrap_token_id_worker" {
  length  = 6
  upper   = false
  special = false
}

# Generate a cryptographically random token secret.
resource "random_string" "bootstrap_token_secret_worker" {
  length  = 16
  upper   = false
  special = false
}
