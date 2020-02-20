output "cluster_dns_service_ip" {
  value = cidrhost(var.service_cidr, 10)
}

// Generated kubeconfig for Kubelets (i.e. lower privilege than admin)
output "kubeconfig-kubelet" {
  value = data.template_file.kubeconfig-kubelet.rendered
}

// Generated kubeconfig for admins (i.e. human super-user)
output "kubeconfig-admin" {
  value = data.template_file.kubeconfig-admin.rendered
}

# etcd TLS assets

output "etcd_ca_cert" {
  value = tls_self_signed_cert.etcd-ca.cert_pem
}

output "etcd_client_cert" {
  value = tls_locally_signed_cert.client.cert_pem
}

output "etcd_client_key" {
  value = tls_private_key.client.private_key_pem
}

output "etcd_server_cert" {
  value = tls_locally_signed_cert.server.cert_pem
}

output "etcd_server_key" {
  value = tls_private_key.server.private_key_pem
}

output "etcd_peer_cert" {
  value = tls_locally_signed_cert.peer.cert_pem
}

output "etcd_peer_key" {
  value = tls_private_key.peer.private_key_pem
}

# Some platforms may need to reconstruct the kubeconfig directly in user-data.
# That can't be done with the way template_file interpolates multi-line
# contents so the raw components of the kubeconfig may be needed.

output "ca_cert" {
  value = base64encode(tls_self_signed_cert.kube-ca.cert_pem)
}

output "kubelet_cert" {
  value = base64encode(tls_locally_signed_cert.kubelet.cert_pem)
}

output "kubelet_key" {
  value = base64encode(tls_private_key.kubelet.private_key_pem)
}

output "server" {
  value = format("https://%s:%s", var.api_servers[0], var.external_apiserver_port)
}

output "server_admin" {
  value = format("https://%s:%s", element(local.api_servers_external, 0), var.external_apiserver_port)
}

# values.yaml content for all deployed charts.
output "kube-apiserver_values" {
  value = local_file.kube-apiserver.content
}

output "kubernetes_values" {
  value = data.template_file.kubernetes.rendered
}

output "kubelet_values" {
  value = data.template_file.kubelet.rendered
}

output "calico_values" {
  value = join("", local_file.calico.*.content)
}

output "flannel_values" {
  value = join("", local_file.flannel.*.content)
}

output "kube-router_values" {
  value = join("", local_file.kube-router.*.content)
}
