output "provisioner_ip" {
  description = "IP address where Tink server is exposed."
  value       = local.provisioner_ips[0]
}

output "sandbox_name" {
  description = "Sandbox name, which should be passed to worker module instances."
  value       = local.base_name
}

output "volumes_pool_name" {
  description = "Name of the storage pool created for sandbox. Should be passed to worker module instances."
  value       = libvirt_pool.pool.name
}

output "network_id" {
  description = "ID of sandbox network. Should be passed to worker module instances."
  value       = libvirt_network.network.id
}

output "netmask" {
  description = "Calculated hosts CIDR network mask for Tinkerbell hardware entries for workers."
  value       = cidrnetmask(var.hosts_cidr)
}

output "gateway" {
  description = "Gateway IP address for Tink workers to have DNS resolution and internet access."
  value       = local.provisioner_gateway
}
