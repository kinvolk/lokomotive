output "kubeconfig-admin" {
  value = module.bootkube.kubeconfig-admin
}

output "kubeconfig" {
  value = module.bootkube.kubeconfig-kubelet
}

output "machine_domain" {
  value = var.machine_domain
}

output "cluster_name" {
  value = var.cluster_name
}

output "ssh_keys" {
  value = var.ssh_keys
}

output "libvirtpool" {
  value = libvirt_pool.volumetmp.name
}

output "libvirtbaseid" {
  value = libvirt_volume.base.id
}
