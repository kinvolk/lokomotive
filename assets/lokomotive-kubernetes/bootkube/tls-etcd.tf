# etcd-ca.crt
resource "local_file" "etcd_ca_crt" {
  content  = tls_self_signed_cert.etcd-ca.cert_pem
  filename = "${var.asset_dir}/tls/etcd-ca.crt"
}

# etcd-ca.key
resource "local_file" "etcd_ca_key" {
  content  = tls_private_key.etcd-ca.private_key_pem
  filename = "${var.asset_dir}/tls/etcd-ca.key"
}

# etcd-client-ca.crt
resource "local_file" "etcd_client_ca_crt" {
  content  = tls_self_signed_cert.etcd-ca.cert_pem
  filename = "${var.asset_dir}/tls/etcd-client-ca.crt"
}

# etcd-client.crt
resource "local_file" "etcd_client_crt" {
  content  = tls_locally_signed_cert.client.cert_pem
  filename = "${var.asset_dir}/tls/etcd-client.crt"
}

# etcd-client.key
resource "local_file" "etcd_client_key" {
  content  = tls_private_key.client.private_key_pem
  filename = "${var.asset_dir}/tls/etcd-client.key"
}

# server-ca.crt
resource "local_file" "etcd_server_ca_crt" {
  content  = tls_self_signed_cert.etcd-ca.cert_pem
  filename = "${var.asset_dir}/tls/etcd/server-ca.crt"
}

# server.crt
resource "local_file" "etcd_server_crt" {
  content  = tls_locally_signed_cert.server.cert_pem
  filename = "${var.asset_dir}/tls/etcd/server.crt"
}

# server.key
resource "local_file" "etcd_server_key" {
  content  = tls_private_key.server.private_key_pem
  filename = "${var.asset_dir}/tls/etcd/server.key"
}

# peer-ca.crt
resource "local_file" "etcd_peer_ca_crt" {
  content  = tls_self_signed_cert.etcd-ca.cert_pem
  filename = "${var.asset_dir}/tls/etcd/peer-ca.crt"
}

# peer.crt
resource "local_file" "etcd_peer_crt" {
  content  = tls_locally_signed_cert.peer.cert_pem
  filename = "${var.asset_dir}/tls/etcd/peer.crt"
}

# peer.key
resource "local_file" "etcd_peer_key" {
  content  = tls_private_key.peer.private_key_pem
  filename = "${var.asset_dir}/tls/etcd/peer.key"
}

# certificates and keys

resource "tls_private_key" "etcd-ca" {
  algorithm = "RSA"
  rsa_bits  = "2048"
}

resource "tls_self_signed_cert" "etcd-ca" {
  key_algorithm   = tls_private_key.etcd-ca.algorithm
  private_key_pem = tls_private_key.etcd-ca.private_key_pem

  subject {
    common_name  = "etcd-ca"
    organization = "etcd"
  }

  is_ca_certificate     = true
  validity_period_hours = var.certs_validity_period_hours

  allowed_uses = [
    "key_encipherment",
    "digital_signature",
    "cert_signing",
  ]
}

# client certs are used for client (apiserver, locksmith, etcd-operator)
# to etcd communication
resource "tls_private_key" "client" {
  algorithm = "RSA"
  rsa_bits  = "2048"
}

resource "tls_cert_request" "client" {
  key_algorithm   = tls_private_key.client.algorithm
  private_key_pem = tls_private_key.client.private_key_pem

  subject {
    common_name  = "etcd-client"
    organization = "etcd"
  }

  ip_addresses = [
    "127.0.0.1",
  ]

  dns_names = concat(var.etcd_servers, ["localhost"])
}

resource "tls_locally_signed_cert" "client" {
  cert_request_pem = tls_cert_request.client.cert_request_pem

  ca_key_algorithm   = join(" ", tls_self_signed_cert.etcd-ca.*.key_algorithm)
  ca_private_key_pem = join(" ", tls_private_key.etcd-ca.*.private_key_pem)
  ca_cert_pem        = join(" ", tls_self_signed_cert.etcd-ca.*.cert_pem)

  validity_period_hours = var.certs_validity_period_hours

  allowed_uses = [
    "key_encipherment",
    "digital_signature",
    "server_auth",
    "client_auth",
  ]
}

resource "tls_private_key" "server" {
  algorithm = "RSA"
  rsa_bits  = "2048"
}

resource "tls_cert_request" "server" {
  key_algorithm   = tls_private_key.server.algorithm
  private_key_pem = tls_private_key.server.private_key_pem

  subject {
    common_name  = "etcd-server"
    organization = "etcd"
  }

  ip_addresses = [
    "127.0.0.1",
  ]

  dns_names = concat(var.etcd_servers, ["localhost"])
}

resource "tls_locally_signed_cert" "server" {
  cert_request_pem = tls_cert_request.server.cert_request_pem

  ca_key_algorithm   = join(" ", tls_self_signed_cert.etcd-ca.*.key_algorithm)
  ca_private_key_pem = join(" ", tls_private_key.etcd-ca.*.private_key_pem)
  ca_cert_pem        = join(" ", tls_self_signed_cert.etcd-ca.*.cert_pem)

  validity_period_hours = var.certs_validity_period_hours

  allowed_uses = [
    "key_encipherment",
    "digital_signature",
    "server_auth",
    "client_auth",
  ]
}

resource "tls_private_key" "peer" {
  algorithm = "RSA"
  rsa_bits  = "2048"
}

resource "tls_cert_request" "peer" {
  key_algorithm   = tls_private_key.peer.algorithm
  private_key_pem = tls_private_key.peer.private_key_pem

  subject {
    common_name  = "etcd-peer"
    organization = "etcd"
  }

  dns_names = var.etcd_servers
}

resource "tls_locally_signed_cert" "peer" {
  cert_request_pem = tls_cert_request.peer.cert_request_pem

  ca_key_algorithm   = join(" ", tls_self_signed_cert.etcd-ca.*.key_algorithm)
  ca_private_key_pem = join(" ", tls_private_key.etcd-ca.*.private_key_pem)
  ca_cert_pem        = join(" ", tls_self_signed_cert.etcd-ca.*.cert_pem)

  validity_period_hours = var.certs_validity_period_hours

  allowed_uses = [
    "key_encipherment",
    "digital_signature",
    "server_auth",
    "client_auth",
  ]
}
