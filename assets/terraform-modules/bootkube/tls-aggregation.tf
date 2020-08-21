# NOTE: Across this module, the following workaround is used:
# `"${var.some_var == "condition" ? join(" ", tls_private_key.aggregation-ca.*.private_key_pem) : ""}"`
# Due to https://github.com/hashicorp/hil/issues/50, both sides of conditions
# are evaluated, until one of them is discarded. When a `count` is used resources
# can be referenced as lists with the `.*` notation, and arrays are allowed to be
# empty. The `join()` interpolation function is then used to cast them back to
# a string. Since `count` can only be 0 or 1, the returned value is either empty
# (and discarded anyways) or the desired value.

# Kubernetes Aggregation CA (i.e. front-proxy-ca)
# Files: tls/{aggregation-ca.crt,aggregation-ca.key}

resource "tls_private_key" "aggregation-ca" {
  count = var.enable_aggregation == true ? 1 : 0

  algorithm = "RSA"
  rsa_bits  = "2048"
}

resource "tls_self_signed_cert" "aggregation-ca" {
  count = var.enable_aggregation == true ? 1 : 0

  key_algorithm   = tls_private_key.aggregation-ca[0].algorithm
  private_key_pem = tls_private_key.aggregation-ca[0].private_key_pem

  subject {
    common_name = "kubernetes-front-proxy-ca"
  }

  is_ca_certificate     = true
  validity_period_hours = var.certs_validity_period_hours

  allowed_uses = [
    "key_encipherment",
    "digital_signature",
    "cert_signing",
  ]
}

resource "local_file" "aggregation-ca-key" {
  count = var.enable_aggregation == true ? 1 : 0

  content  = tls_private_key.aggregation-ca[0].private_key_pem
  filename = "${var.asset_dir}/tls/aggregation-ca.key"
}

resource "local_file" "aggregation-ca-crt" {
  count = var.enable_aggregation == true ? 1 : 0

  content  = tls_self_signed_cert.aggregation-ca[0].cert_pem
  filename = "${var.asset_dir}/tls/aggregation-ca.crt"
}

# Kubernetes apiserver (i.e. front-proxy-client)
# Files: tls/{aggregation-client.crt,aggregation-client.key}

resource "tls_private_key" "aggregation-client" {
  count = var.enable_aggregation == true ? 1 : 0

  algorithm = "RSA"
  rsa_bits  = "2048"
}

resource "tls_cert_request" "aggregation-client" {
  count = var.enable_aggregation == true ? 1 : 0

  key_algorithm   = tls_private_key.aggregation-client[0].algorithm
  private_key_pem = tls_private_key.aggregation-client[0].private_key_pem

  subject {
    common_name = "kube-apiserver"
  }
}

resource "tls_locally_signed_cert" "aggregation-client" {
  count = var.enable_aggregation == true ? 1 : 0

  cert_request_pem = tls_cert_request.aggregation-client[0].cert_request_pem

  ca_key_algorithm   = tls_self_signed_cert.aggregation-ca[0].key_algorithm
  ca_private_key_pem = tls_private_key.aggregation-ca[0].private_key_pem
  ca_cert_pem        = tls_self_signed_cert.aggregation-ca[0].cert_pem

  validity_period_hours = var.certs_validity_period_hours

  allowed_uses = [
    "key_encipherment",
    "digital_signature",
    "client_auth",
  ]
}

resource "local_file" "aggregation-client-key" {
  count = var.enable_aggregation == true ? 1 : 0

  content  = tls_private_key.aggregation-client[0].private_key_pem
  filename = "${var.asset_dir}/tls/aggregation-client.key"
}

resource "local_file" "aggregation-client-crt" {
  count = var.enable_aggregation == true ? 1 : 0

  content  = tls_locally_signed_cert.aggregation-client[0].cert_pem
  filename = "${var.asset_dir}/tls/aggregation-client.crt"
}
