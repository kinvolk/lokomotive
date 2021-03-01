resource "packet_device" "nodes" {
  count            = var.worker_count
  hostname         = "${var.cluster_name}-${var.pool_name}-worker-${count.index}"
  plan             = var.type
  facilities       = [var.facility]
  operating_system = var.ipxe_script_url != "" ? "custom_ipxe" : format("flatcar_%s", var.os_channel)
  billing_cycle    = "hourly"
  project_id       = var.project_id
  ipxe_script_url  = var.ipxe_script_url
  always_pxe       = false
  user_data        = var.ipxe_script_url != "" ? data.ct_config.install-ignitions.rendered : data.ct_config.ignitions.rendered

  # If not present in the map, it uses ${var.reservation_ids_default}.
  hardware_reservation_id = lookup(
    var.reservation_ids,
    format("worker-%v", count.index),
    var.reservation_ids_default,
  )

  lifecycle {
    ignore_changes = [
      // With newer Packet provider, changing userdata causes re-creation of the device,
      // which we want to silent to avoid destroying nodes, as they may contain local data.
      user_data,
    ]
  }

  tags = var.tags

  # This way to handle dependencies was inspired in this:
  # https://discuss.hashicorp.com/t/tips-howto-implement-module-depends-on-emulation/2305/2
  depends_on = [var.nodes_depend_on]
}

# These configs are used for the fist boot, to run flatcar-install
data "ct_config" "install-ignitions" {
  content = templatefile("${path.module}/cl/install.yaml.tmpl", {
    os_channel           = var.os_channel
    os_version           = var.os_version
    flatcar_linux_oem    = "packet"
    ssh_keys             = jsonencode(var.ssh_keys)
    postinstall_ignition = data.ct_config.ignitions.rendered
  })
}

data "ct_config" "ignitions" {
  content = templatefile(
    "${path.module}/cl/worker.yaml.tmpl",
    {
      os_arch = var.os_arch
      kubeconfig = var.enable_tls_bootstrap ? indent(10, templatefile("${path.module}/cl/bootstrap-kubeconfig.yaml.tmpl", {
        token_id     = random_string.bootstrap_token_id[0].result
        token_secret = random_string.bootstrap_token_secret[0].result
        ca_cert      = var.ca_cert
        server       = "https://${var.apiserver}:6443"
      })) : indent(10, var.kubeconfig)
      ssh_keys              = jsonencode(var.ssh_keys)
      k8s_dns_service_ip    = cidrhost(var.service_cidr, 10)
      cluster_domain_suffix = var.cluster_domain_suffix
      node_labels = merge({
        "node.kubernetes.io/node"                 = "",
        "lokomotive.alpha.kinvolk.io/bgp-enabled" = format("%t", ! var.disable_bgp),
      }, var.labels)
      taints               = var.taints
      setup_raid           = var.setup_raid
      setup_raid_hdd       = var.setup_raid_hdd
      setup_raid_ssd       = var.setup_raid_ssd
      setup_raid_ssd_fs    = var.setup_raid_ssd_fs
      cluster_name         = var.cluster_name
      dns_zone             = var.dns_zone
      enable_tls_bootstrap = var.enable_tls_bootstrap
      cpu_manager_policy   = var.cpu_manager_policy
    }
  )
  platform = "packet"
  snippets = var.clc_snippets
}

data "packet_project" "project" {
  project_id = var.project_id
}
