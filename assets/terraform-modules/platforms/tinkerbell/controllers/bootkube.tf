# Self-hosted Kubernetes assets (kubeconfig, manifests).
module "bootkube" {
  source = "../../../bootkube"

  cluster_name          = var.cluster_name
  api_servers           = [format("%s.%s", var.cluster_name, var.dns_zone)]
  api_servers_external  = var.ip_addresses
  api_servers_ips       = var.ip_addresses
  etcd_servers          = module.controller[0].etcd_servers
  asset_dir             = var.asset_dir
  network_mtu           = var.network_mtu
  pod_cidr              = var.pod_cidr
  service_cidr          = var.service_cidr
  cluster_domain_suffix = var.cluster_domain_suffix
  enable_reporting      = var.enable_reporting
  enable_aggregation    = var.enable_aggregation
  controller_count      = var.node_count

  certs_validity_period_hours = var.certs_validity_period_hours

  bootstrap_tokens            = concat(module.controller.*.bootstrap_token, var.worker_bootstrap_tokens)
  enable_tls_bootstrap        = true
  disable_self_hosted_kubelet = false
  conntrack_max_per_core      = var.conntrack_max_per_core
}
