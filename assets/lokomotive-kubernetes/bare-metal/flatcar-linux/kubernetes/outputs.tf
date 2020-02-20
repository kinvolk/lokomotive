output "kubeconfig-admin" {
  value = module.bootkube.kubeconfig-admin
}

# values.yaml content for all deployed charts.
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

output "flannel_values" {
  value = module.bootkube.flannel_values
}

output "kube-router_values" {
  value = module.bootkube.kube-router_values
}
