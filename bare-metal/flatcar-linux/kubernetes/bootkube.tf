# Self-hosted Kubernetes assets (kubeconfig, manifests)
module "bootkube" {
  source = "github.com/kinvolk/terraform-render-bootkube?ref=5809204d5abba8aa4f43d194eeef65cfd8660aa5"

  cluster_name                    = var.cluster_name
  api_servers                     = [var.k8s_domain_name]
  etcd_servers                    = var.controller_domains
  asset_dir                       = var.asset_dir
  networking                      = var.networking
  network_mtu                     = var.network_mtu
  network_ip_autodetection_method = var.network_ip_autodetection_method
  pod_cidr                        = var.pod_cidr
  service_cidr                    = var.service_cidr
  cluster_domain_suffix           = var.cluster_domain_suffix
  enable_reporting                = var.enable_reporting
  enable_aggregation              = var.enable_aggregation

  certs_validity_period_hours = var.certs_validity_period_hours
}
