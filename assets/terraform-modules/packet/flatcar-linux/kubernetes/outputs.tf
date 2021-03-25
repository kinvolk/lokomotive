output "kubeconfig-admin" {
  value = module.bootkube.kubeconfig-admin
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

output "calico-host-protection_values" {
  value = join("", local_file.calico_host_protection.*.content)
}

output "lokomotive_values" {
  value = module.bootkube.lokomotive_values
}

output "bootstrap-secrets_values" {
  value = module.bootkube.bootstrap-secrets_values
}

# Dummy output used to create dependencies only
# Not guaranteed that won't change
output "device_ids" {
  value = packet_device.controllers.*.id
}

output "controllers_public_ipv4" {
  value = packet_device.controllers.*.access_public_ipv4
}

output "controllers_private_ipv4" {
  value = packet_device.controllers.*.access_private_ipv4
}
