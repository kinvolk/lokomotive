locals {
  # api_server = module.bootkube.api_servers
  api_server = data.template_file.controllernames[0].rendered
}

module "bootkube" {
  source       = "../../../bootkube"
  cluster_name = var.cluster_name

  # Cannot use cyclic dependencies on controllers or their DNS records
  api_servers          = [data.template_file.controllernames[0].rendered]
  api_servers_external = libvirt_domain.controller-machine.*.network_interface.0.addresses.0
  api_servers_ips      = libvirt_domain.controller-machine.*.network_interface.0.addresses.0
  etcd_servers         = data.template_file.controllernames.*.rendered
  asset_dir            = var.asset_dir
  network_mtu          = var.network_mtu

  network_ip_autodetection_method = var.network_ip_autodetection_method

  pod_cidr              = var.pod_cidr
  service_cidr          = var.service_cidr
  cluster_domain_suffix = var.cluster_domain_suffix
  enable_reporting      = var.enable_reporting
  enable_aggregation    = var.enable_aggregation
  encrypt_pod_traffic   = var.encrypt_pod_traffic

  # Disable the self hosted kubelet.
  disable_self_hosted_kubelet = var.disable_self_hosted_kubelet
  # Extra flags to API server.
  kube_apiserver_extra_flags = var.kube_apiserver_extra_flags

  bootstrap_tokens     = var.enable_tls_bootstrap ? concat([local.controller_bootstrap_token], var.worker_bootstrap_tokens) : []
  enable_tls_bootstrap = var.enable_tls_bootstrap

  certs_validity_period_hours = var.certs_validity_period_hours
}

data "template_file" "controllernames" {
  count    = var.controller_count
  template = "$${cluster_name}-controller-$${index}.$${machine_domain}"

  vars = {
    index          = count.index
    cluster_name   = var.cluster_name
    machine_domain = var.machine_domain
  }
}
