output "kubeconfig-admin" {
  value = module.bootkube.kubeconfig-admin
}


# Outputs for worker pools

output "subnet_id" {
  value       = oci_core_subnet.subnet.id
  description = "Subnet ID for creating worker instances"
}

output "nsg_id" {
  value = oci_core_network_security_group.lokomotive.id
}

output "kubeconfig" {
  value = module.bootkube.kubeconfig-kubelet
}

output "ca_cert" {
  value = module.bootkube.ca_cert
}

output "apiserver" {
  value = local.api_server
}

# values.yaml content for all deployed charts.
output "pod-checkpointer_values" {
  value = module.bootkube.pod-checkpointer_values
}

output "kube-apiserver_values" {
  value = module.bootkube.kube-apiserver_values
}

output "kubernetes_values" {
  value = module.bootkube.kubernetes_values
}

output "kubelet_values" {
  value = module.bootkube.kubelet_values
}

output "calico_values" {
  value = module.bootkube.calico_values
}

output "lokomotive_values" {
  value = module.bootkube.lokomotive_values
}

output "bootstrap-secrets_values" {
  value = module.bootkube.bootstrap-secrets_values
}
