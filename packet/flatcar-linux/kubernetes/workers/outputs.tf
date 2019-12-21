output "worker_nodes_hostname" {
  value = packet_device.nodes.*.hostname
}

output "worker_nodes_public_ipv4" {
  value = packet_device.nodes.*.access_public_ipv4
}

