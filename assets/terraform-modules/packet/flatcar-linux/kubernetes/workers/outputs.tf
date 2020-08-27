output "worker_nodes_hostname" {
  value = packet_device.nodes.*.hostname
}

output "worker_nodes_public_ipv4" {
  value = packet_device.nodes.*.access_public_ipv4
}

# Dummy output used to create dependencies only
# Not guaranteed that won't change
output "device_ids" {
  value = packet_device.nodes.*.id
}

output "worker_bootstrap_token" {
  value = local.worker_bootstrap_token
}
