# Self-hosted Kubernetes assets (kubeconfig, manifests)
module "bootkube" {
  source = "../../../bootkube"

  cluster_name = var.cluster_name
  api_servers  = [format("%s.%s", var.cluster_name, var.k8s_domain_name)]
  # Each instance of controller module generates the same set of etcd_servers.
  etcd_servers                    = module.controller[0].etcd_servers
  etcd_endpoints                  = []
  asset_dir                       = var.asset_dir
  network_mtu                     = var.network_mtu
  network_ip_autodetection_method = var.network_ip_autodetection_method
  pod_cidr                        = var.pod_cidr
  service_cidr                    = var.service_cidr
  cluster_domain_suffix           = var.cluster_domain_suffix
  enable_reporting                = var.enable_reporting
  enable_aggregation              = var.enable_aggregation
  kube_apiserver_extra_flags      = var.kube_apiserver_extra_flags
  controller_count                = length(var.controller_domains)

  certs_validity_period_hours = var.certs_validity_period_hours

  # Disable the self hosted kubelet.
  disable_self_hosted_kubelet = var.disable_self_hosted_kubelet

  bootstrap_tokens     = concat(module.controller.*.bootstrap_token, module.worker.*.bootstrap_token)
  enable_tls_bootstrap = true
  encrypt_pod_traffic  = var.encrypt_pod_traffic

  ignore_x509_cn_check = var.ignore_x509_cn_check

  conntrack_max_per_core = var.conntrack_max_per_core

  # Node Local DNS configuration.
  enable_node_local_dns = var.enable_node_local_dns
  node_local_dns_ip     = var.node_local_dns_ip
}
