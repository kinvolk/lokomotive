// Flatcar Container Linux install profile (from release.flatcar-linux.net)
resource "matchbox_profile" "flatcar-install" {
  count = length(var.controller_names) + length(var.worker_names)
  name = format(
    "%s-flatcar-install-%s",
    var.cluster_name,
    concat(var.controller_names, var.worker_names)[count.index]
  )

  kernel = "${var.download_protocol}://${var.os_channel}.release.flatcar-linux.net/amd64-usr/${var.os_version}/flatcar_production_pxe.vmlinuz"

  initrd = [
    "${var.download_protocol}://${var.os_channel}.release.flatcar-linux.net/amd64-usr/${var.os_version}/flatcar_production_pxe_image.cpio.gz",
  ]

  args = flatten([
    "initrd=flatcar_production_pxe_image.cpio.gz",
    "ignition.config.url=${var.matchbox_http_endpoint}/ignition?uuid=$${uuid}&mac=$${mac:hexhyp}",
    "flatcar.first_boot=yes",
    "console=tty0",
    "console=ttyS0",
    var.kernel_args,
  ])

  container_linux_config = templatefile("${path.module}/cl/install.yaml.tmpl", {
    os_channel          = var.os_channel
    os_version          = var.os_version
    ignition_endpoint   = format("%s/ignition", var.matchbox_http_endpoint)
    install_disk        = var.install_disk
    container_linux_oem = var.container_linux_oem
    ssh_keys            = jsonencode(var.ssh_keys)
    # only cached-container-linux profile adds -b baseurl
    baseurl_flag = ""
  })
}

// Flatcar Container Linux Install profile (from matchbox /assets cache)
// Note: Admin must have downloaded os_version into matchbox assets/flatcar.
resource "matchbox_profile" "cached-flatcar-linux-install" {
  count = length(var.controller_names) + length(var.worker_names)
  name = format(
    "%s-cached-flatcar-linux-install-%s",
    var.cluster_name,
    concat(var.controller_names, var.worker_names)[count.index]
  )

  kernel = "/assets/flatcar/${var.os_version}/flatcar_production_pxe.vmlinuz"

  initrd = [
    "/assets/flatcar/${var.os_version}/flatcar_production_pxe_image.cpio.gz",
  ]

  args = flatten([
    "initrd=flatcar_production_pxe_image.cpio.gz",
    "ignition.config.url=${var.matchbox_http_endpoint}/ignition?uuid=$${uuid}&mac=$${mac:hexhyp}",
    "flatcar.first_boot=yes",
    "console=tty0",
    "console=ttyS0",
    var.kernel_args,
  ])

  container_linux_config = templatefile("${path.module}/cl/install.yaml.tmpl", {
    os_channel          = var.os_channel
    os_version          = var.os_version
    ignition_endpoint   = format("%s/ignition", var.matchbox_http_endpoint)
    install_disk        = var.install_disk
    container_linux_oem = var.container_linux_oem
    ssh_keys            = jsonencode(var.ssh_keys)
    # profile uses -b baseurl to install from matchbox cache
    baseurl_flag = "-b ${var.matchbox_http_endpoint}/assets/flatcar"
  })
}

// Kubernetes Controller profiles
resource "matchbox_profile" "controllers" {
  count = length(var.controller_names)
  name = format(
    "%s-controller-%s",
    var.cluster_name,
    var.controller_names[count.index]
  )
  raw_ignition = data.ct_config.controller-ignitions[count.index].rendered
}

data "ct_config" "controller-ignitions" {
  count = length(var.controller_names)
  content = templatefile("${path.module}/cl/controller.yaml.tmpl", {
    domain_name = var.controller_domains[count.index]
    etcd_name   = var.controller_names[count.index]
    etcd_initial_cluster = join(
      ",",
      formatlist(
        "%s=https://%s:2380",
        var.controller_names,
        var.controller_domains,
      ),
    )
    cluster_dns_service_ip = module.bootkube.cluster_dns_service_ip
    cluster_domain_suffix  = var.cluster_domain_suffix
    ssh_keys               = jsonencode(var.ssh_keys)
    enable_tls_bootstrap   = var.enable_tls_bootstrap
  })
  pretty_print = false

  # Must use direct lookup. Cannot use lookup(map, key) since it only works for flat maps
  snippets = local.clc_map[var.controller_names[count.index]]
}

// Kubernetes Worker profiles
resource "matchbox_profile" "workers" {
  count = length(var.worker_names)
  name = format(
    "%s-worker-%s",
    var.cluster_name,
    var.worker_names[count.index]
  )
  raw_ignition = data.ct_config.worker-ignitions[count.index].rendered
}

data "ct_config" "worker-ignitions" {
  count = length(var.worker_names)
  content = templatefile("${path.module}/cl/worker.yaml.tmpl", {
    domain_name            = var.worker_domains[count.index]
    cluster_dns_service_ip = module.bootkube.cluster_dns_service_ip
    cluster_domain_suffix  = var.cluster_domain_suffix
    ssh_keys               = jsonencode(var.ssh_keys)
    kubelet_labels         = merge({ "node.kubernetes.io/node" = "" }, var.labels),
    enable_tls_bootstrap   = var.enable_tls_bootstrap
  })
  pretty_print = false

  # Must use direct lookup. Cannot use lookup(map, key) since it only works for flat maps
  snippets = local.clc_map[var.worker_names[count.index]]
}

locals {
  # TODO: Probably it is not needed anymore with terraform 0.12
  # Hack to workaround https://github.com/hashicorp/terraform/issues/17251
  # Default Flatcar Container Linux config snippets map every node names to list("\n") so
  # all lookups succeed
  total_length = length(var.controller_names) + length(var.worker_names)
  clc_defaults = zipmap(
    concat(var.controller_names, var.worker_names),
    chunklist([for i in range(local.total_length) : "\n"], 1),
  )

  # Union of the default and user specific snippets, later overrides prior.
  clc_map = merge(local.clc_defaults, var.clc_snippets)
}
