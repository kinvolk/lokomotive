# Generate a cryptographically random token id (public).
resource random_string "bootstrap_token_id" {
  length  = 6
  upper   = false
  special = false
}

# Generate a cryptographically random token secret.
resource random_string "bootstrap_token_secret" {
  length  = 16
  upper   = false
  special = false
}

locals {
  bootstrap_token = {
    token_id     = random_string.bootstrap_token_id.result
    token_secret = random_string.bootstrap_token_secret.result
  }

  bootstrap_kubeconfig = templatefile("${path.module}/templates/bootstrap-kubeconfig.yaml.tmpl", {
    token_id     = random_string.bootstrap_token_id.result
    token_secret = random_string.bootstrap_token_secret.result
    ca_cert      = var.ca_cert
    server       = "https://${var.apiserver}:6443"
  })
}
