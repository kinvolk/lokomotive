output "kubeconfig-admin" {
  value = module.bootkube.kubeconfig-admin
}

output "kubeconfig" {
  value = module.bootkube.kubeconfig-kubelet
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
