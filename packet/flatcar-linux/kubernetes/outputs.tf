output "kubeconfig-admin" {
  value = module.bootkube.kubeconfig-admin
}

output "kubeconfig" {
  value = module.bootkube.kubeconfig-kubelet
}

output "dns_entries" {
  value = local.dns_entries
}
