locals {
  api_server = format("%s-private.%s", var.cluster_name, var.dns_zone)
}

module "bootkube" {
  source       = "../../../bootkube"
  cluster_name = var.cluster_name

  # Cannot use cyclic dependencies on controllers or their DNS records
  api_servers          = [local.api_server]
  api_servers_external = [format("%s.%s", var.cluster_name, var.dns_zone)]
  etcd_servers         = [for i, d in packet_device.controllers : format("%s-etcd%d.%s", var.cluster_name, i, var.dns_zone)]
  asset_dir            = var.asset_dir
  network_mtu          = var.network_mtu
  etcd_endpoints       = packet_device.controllers.*.access_private_ipv4

  # Select private Packet NIC by using the can-reach Calico autodetection option with the first
  # host in our private CIDR.
  network_ip_autodetection_method = "can-reach=${cidrhost(var.node_private_cidr, 1)}"

  pod_cidr              = var.pod_cidr
  service_cidr          = var.service_cidr
  cluster_domain_suffix = var.cluster_domain_suffix
  enable_reporting      = var.enable_reporting
  enable_aggregation    = var.enable_aggregation

  certs_validity_period_hours = var.certs_validity_period_hours

  # Disable the self hosted kubelet.
  disable_self_hosted_kubelet = var.disable_self_hosted_kubelet
  # Extra flags to API server.
  kube_apiserver_extra_flags = var.kube_apiserver_extra_flags

  # Block access to Packet metadata service.
  #
  # https://www.packet.com/developers/docs/servers/key-features/metadata/
  #
  # metadata.packet.net should always resolve to 192.80.8.124.
  blocked_metadata_cidrs = ["192.80.8.124/32"]

  bootstrap_tokens     = var.enable_tls_bootstrap ? concat([local.controller_bootstrap_token], var.worker_bootstrap_tokens) : []
  enable_tls_bootstrap = var.enable_tls_bootstrap

  # We install calico-host-protection chart on Packet which ships GNPs, so we can disable failsafe ports in Calico.
  failsafe_inbound_host_ports = []
  encrypt_pod_traffic         = var.encrypt_pod_traffic

  ignore_x509_cn_check = var.ignore_x509_cn_check

  conntrack_max_per_core = var.conntrack_max_per_core

  cloud_provider = "external"

  install_prometheus_operator_crds = var.install_prometheus_operator_crds
}
