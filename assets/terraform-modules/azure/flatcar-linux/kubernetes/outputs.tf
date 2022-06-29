output "kubeconfig-admin" {
  value = module.bootkube.kubeconfig-admin
}

# Outputs for Kubernetes Ingress

output "ingress_static_ipv4" {
  value       = azurerm_public_ip.ingress-ipv4.ip_address
  description = "IPv4 address of the load balancer for distributing traffic to Ingress controllers"
}

# Outputs for worker pools

output "region" {
  value = azurerm_resource_group.cluster.location
}

output "resource_group_name" {
  value = azurerm_resource_group.cluster.name
}

output "resource_group_id" {
  value = azurerm_resource_group.cluster.id
}

output "subnet_id" {
  value = azurerm_subnet.worker.id
}

output "security_group_id" {
  value = azurerm_network_security_group.worker.id
}

output "kubeconfig" {
  value = module.bootkube.kubeconfig-kubelet
}

output "ca_cert" {
  value = module.bootkube.ca_cert
}

output "apiserver" {
  value = local.api_server
}

# Outputs for custom firewalling

output "worker_security_group_name" {
  value = azurerm_network_security_group.worker.name
}

output "worker_address_prefix" {
  description = "Worker network subnet CIDR address (for source/destination)"
  value       = azurerm_subnet.worker.address_prefix
}

# Outputs for custom load balancing

output "loadbalancer_id" {
  description = "ID of the cluster load balancer"
  value       = azurerm_lb.cluster.id
}

output "backend_address_pool_id" {
  description = "ID of the worker backend address pool"
  value       = azurerm_lb_backend_address_pool.worker.id
}

# values.yaml content for all deployed charts.
output "pod-checkpointer_values" {
  value = module.bootkube.pod-checkpointer_values
}

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

output "lokomotive_values" {
  value = module.bootkube.lokomotive_values
}

output "bootstrap-secrets_values" {
  value = module.bootkube.bootstrap-secrets_values
}

output "node-local-dns_values" {
  value = module.bootkube.node-local-dns_values
}

output "controllers_public_ipv4" {
  value = azurerm_linux_virtual_machine.controllers.*.public_ip_address
}

output "controllers_private_ipv4" {
  value = azurerm_linux_virtual_machine.controllers.*.private_ip_address
}
