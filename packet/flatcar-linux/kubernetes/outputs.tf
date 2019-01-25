output "kubeconfig-admin" {
  value = "${module.bootkube.user-kubeconfig}"
}

output "kubeconfig" {
  value = "${module.bootkube.kubeconfig}"
}
