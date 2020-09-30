# Self-hosted Kubernetes assets (kubeconfig, manifests)
module "bootkube" {
  source = "../../../bootkube"

  cluster_name          = var.cluster_name
  api_servers           = [format("%s.%s", var.cluster_name, var.dns_zone)]
  etcd_servers          = google_dns_record_set.etcds.*.name
  asset_dir             = var.asset_dir
  network_mtu           = 1440
  pod_cidr              = var.pod_cidr
  service_cidr          = var.service_cidr
  cluster_domain_suffix = var.cluster_domain_suffix
  enable_reporting      = var.enable_reporting
  enable_aggregation    = var.enable_aggregation
  encrypt_pod_traffic   = var.encrypt_pod_traffic

  // temporary
  external_apiserver_port = 443

  certs_validity_period_hours = var.certs_validity_period_hours
}
