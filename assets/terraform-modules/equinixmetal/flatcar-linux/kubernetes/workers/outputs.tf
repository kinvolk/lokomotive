output "worker_nodes_hostname" {
  value = metal_device.nodes.*.hostname
}

output "worker_nodes_public_ipv4" {
  value = metal_device.nodes.*.access_public_ipv4
}

# Dummy output used to create dependencies only
# Not guaranteed that won't change
output "device_ids" {
  value = metal_device.nodes.*.id
}

output "worker_bootstrap_token" {
  value = local.worker_bootstrap_token
}
