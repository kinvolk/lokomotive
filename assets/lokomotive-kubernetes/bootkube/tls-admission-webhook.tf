resource "tls_private_key" "admission-webhook-server" {
  algorithm = "RSA"
  rsa_bits  = "2048"
}

resource "tls_cert_request" "admission-webhook-server" {
  key_algorithm   = tls_private_key.admission-webhook-server.algorithm
  private_key_pem = tls_private_key.admission-webhook-server.private_key_pem

  subject {
    common_name  = "admission-webhook-server"
    organization = "kinvolk"
  }

  dns_names = [
    "admission-webhook-server",
    "admission-webhook-server.lokomotive-system",
    "admission-webhook-server.lokomotive-system.svc",
    "admission-webhook-server.lokomotive-system.svc.cluster",
    "admission-webhook-server.lokomotive-system.svc.cluster.local",
  ]
}

resource "tls_locally_signed_cert" "admission-webhook-server" {
  cert_request_pem = tls_cert_request.admission-webhook-server.cert_request_pem

  ca_key_algorithm   = tls_self_signed_cert.kube-ca.key_algorithm
  ca_private_key_pem = tls_private_key.kube-ca.private_key_pem
  ca_cert_pem        = tls_self_signed_cert.kube-ca.cert_pem

  validity_period_hours = var.certs_validity_period_hours

  allowed_uses = [
    "key_encipherment",
    "digital_signature",
    "server_auth",
  ]
}

resource "local_file" "lokomotive" {
  filename = "${var.asset_dir}/charts/lokomotive-system/lokomotive.yaml"
  content = templatefile("${path.module}/resources/charts/lokomotive.yaml", {
    serving_key  = base64encode(tls_private_key.admission-webhook-server.private_key_pem)
    serving_cert = base64encode(tls_locally_signed_cert.admission-webhook-server.cert_pem)
  })
}
