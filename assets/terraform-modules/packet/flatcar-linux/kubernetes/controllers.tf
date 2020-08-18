resource "packet_device" "controllers" {
  count            = var.controller_count
  hostname         = "${var.cluster_name}-controller-${count.index}"
  plan             = var.controller_type
  facilities       = [var.facility]
  operating_system = var.ipxe_script_url != "" ? "custom_ipxe" : format("flatcar_%s", var.os_channel)
  billing_cycle    = "hourly"
  project_id       = var.project_id
  user_data        = var.ipxe_script_url != "" ? data.ct_config.controller-install-ignitions[count.index].rendered : data.ct_config.controller-ignitions[count.index].rendered

  # If not present in the map, it uses ${var.reservation_ids_default}.
  hardware_reservation_id = lookup(
    var.reservation_ids,
    format("controller-%v", count.index),
    var.reservation_ids_default,
  )

  ipxe_script_url = var.ipxe_script_url
  always_pxe      = false
  tags            = var.tags

  # This way to handle dependencies was inspired in this:
  # https://discuss.hashicorp.com/t/tips-howto-implement-module-depends-on-emulation/2305/2
  depends_on = [var.nodes_depend_on]
}

data "ct_config" "controller-install-ignitions" {
  count = var.controller_count
  content = templatefile("${path.module}/cl/controller-install.yaml.tmpl", {
    os_channel           = var.os_channel
    os_version           = var.os_version
    flatcar_linux_oem    = "packet"
    ssh_keys             = jsonencode(var.ssh_keys)
    postinstall_ignition = data.ct_config.controller-ignitions[count.index].rendered
  })
}

data "ct_config" "controller-ignitions" {
  count    = var.controller_count
  platform = "packet"
  content = templatefile("${path.module}/cl/controller.yaml.tmpl", {
    os_arch = var.os_arch
    # Cannot use cyclic dependencies on controllers or their DNS records
    etcd_name            = "etcd${count.index}"
    etcd_domain          = "${var.cluster_name}-etcd${count.index}.${var.dns_zone}"
    etcd_arch_tag_suffix = var.os_arch == "arm64" ? "-arm64" : ""
    etcd_arch_options    = var.os_arch == "arm64" ? "ETCD_UNSUPPORTED_ARCH=arm64" : ""
    # etcd0=https://cluster-etcd0.example.com,etcd1=https://cluster-etcd1.example.com,...
    etcd_initial_cluster = join(",", data.template_file.etcds.*.rendered)
    kubeconfig = var.enable_tls_bootstrap ? indent(10, templatefile("${path.module}/workers/cl/bootstrap-kubeconfig.yaml.tmpl", {
      token_id     = random_string.bootstrap_token_id[0].result
      token_secret = random_string.bootstrap_token_secret[0].result
      ca_cert      = module.bootkube.ca_cert
      server       = "https://${local.api_server}:6443"
    })) : indent(10, module.bootkube.kubeconfig-kubelet)
    ssh_keys              = jsonencode(var.ssh_keys)
    k8s_dns_service_ip    = cidrhost(var.service_cidr, 10)
    cluster_domain_suffix = var.cluster_domain_suffix
    controller_count      = var.controller_count
    dns_zone              = var.dns_zone
    cluster_name          = var.cluster_name
    enable_tls_bootstrap  = var.enable_tls_bootstrap
  })
  snippets = var.controller_clc_snippets
}

data "template_file" "etcds" {
  count    = var.controller_count
  template = "etcd$${index}=https://$${cluster_name}-etcd$${index}.$${dns_zone}:2380"

  vars = {
    index        = count.index
    cluster_name = var.cluster_name
    dns_zone     = var.dns_zone
  }
}
