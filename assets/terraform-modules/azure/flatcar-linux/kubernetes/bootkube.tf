# Self-hosted Kubernetes assets (kubeconfig, manifests)
locals {
  api_server = format("%s.%s", var.cluster_name, var.dns_zone)
}

module "bootkube" {
  source = "../../../bootkube"

  cluster_name     = var.cluster_name
  api_servers      = [local.api_server]
  etcd_servers     = [for i, d in azurerm_linux_virtual_machine.controllers : format("%s-etcd%d.%s", var.cluster_name, i, var.dns_zone)]
  etcd_endpoints   = azurerm_linux_virtual_machine.controllers.*.private_ip_address
  asset_dir        = var.asset_dir
  controller_count = var.controller_count

  network_encapsulation = "vxlan"

  # we should be able to use 1450 MTU, but in practice, 1410 was needed
  network_mtu = "1410"

  conntrack_max_per_core = var.conntrack_max_per_core
  pod_cidr               = var.pod_cidr
  service_cidr           = var.service_cidr
  cluster_domain_suffix  = var.cluster_domain_suffix
  bootstrap_tokens       = var.enable_tls_bootstrap ? concat([local.controller_bootstrap_token], var.worker_bootstrap_tokens) : []
  enable_tls_bootstrap   = var.enable_tls_bootstrap
  enable_reporting       = var.enable_reporting
  enable_aggregation     = var.enable_aggregation
  encrypt_pod_traffic    = var.encrypt_pod_traffic
  # Disable the self hosted kubelet.
  disable_self_hosted_kubelet = var.disable_self_hosted_kubelet
  certs_validity_period_hours = var.certs_validity_period_hours
}
