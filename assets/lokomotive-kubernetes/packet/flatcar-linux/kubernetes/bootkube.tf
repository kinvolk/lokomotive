module "bootkube" {
  source       = "../../../bootkube"
  cluster_name = var.cluster_name

  # Cannot use cyclic dependencies on controllers or their DNS records
  api_servers          = [local.api_fqdn]
  api_servers_external = [local.api_external_fqdn]
  etcd_servers         = local.etcd_fqdn
  asset_dir            = var.asset_dir
  network_mtu          = var.network_mtu

  # Select private Packet NIC by using the can-reach Calico autodetection option with the first
  # host in our private CIDR.
  network_ip_autodetection_method = "can-reach=${cidrhost(var.node_private_cidr, 1)}"

  pod_cidr              = var.pod_cidr
  service_cidr          = var.service_cidr
  cluster_domain_suffix = var.cluster_domain_suffix
  enable_reporting      = var.enable_reporting
  enable_aggregation    = var.enable_aggregation

  certs_validity_period_hours = var.certs_validity_period_hours

  container_arch = var.os_arch

  expose_on_all_interfaces = true

  # Disable the self hosted kubelet.
  disable_self_hosted_kubelet = var.disable_self_hosted_kubelet
}
