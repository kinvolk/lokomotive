output "kubeconfig-admin" {
  value = module.bootkube.kubeconfig-admin
}

# values.yaml content for all deployed charts.
output "kubernetes_values" {
  value = module.bootkube.kubernetes_values
}

output "kubelet_values" {
  value = module.bootkube.kubelet_values
}

output "calico_values" {
  value = module.bootkube.calico_values
}
