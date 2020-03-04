output "kubeconfig-admin" {
  value = module.bootkube.kubeconfig-admin
}

# Outputs for Kubernetes Ingress

output "ingress_static_ipv4" {
  description = "Global IPv4 address for proxy load balancing to the nearest Ingress controller"
  value       = google_compute_global_address.ingress-ipv4.address
}

output "ingress_static_ipv6" {
  description = "Global IPv6 address for proxy load balancing to the nearest Ingress controller"
  value       = google_compute_global_address.ingress-ipv6.address
}

# Outputs for worker pools

output "network_name" {
  value = google_compute_network.network.name
}

output "kubeconfig" {
  value = module.bootkube.kubeconfig-kubelet
}

# Outputs for custom firewalling

output "network_self_link" {
  value = google_compute_network.network.self_link
}

# Outputs for custom load balancing

output "worker_instance_group" {
  description = "Worker managed instance group full URL"
  value       = module.workers.instance_group
}

output "worker_target_pool" {
  description = "Worker target pool self link"
  value       = module.workers.target_pool
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
