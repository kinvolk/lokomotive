resource "matchbox_profile" "flatcar-install" {
  name = format(
    "flatcar-install-%s",
    var.node_name
  )

  kernel = "${var.download_protocol}://${var.os_channel}.release.flatcar-linux.net/amd64-usr/${var.os_version}/flatcar_production_pxe.vmlinuz"

  initrd = [
    "${var.download_protocol}://${var.os_channel}.release.flatcar-linux.net/amd64-usr/${var.os_version}/flatcar_production_pxe_image.cpio.gz",
  ]

  args = flatten([
    "initrd=flatcar_production_pxe_image.cpio.gz",
    "ignition.config.url=${var.http_endpoint}/ignition?uuid=$${uuid}&mac=$${mac:hexhyp}",
    "flatcar.first_boot=yes",
    var.kernel_console,
    var.kernel_args,
  ])

  container_linux_config = templatefile("${path.module}/templates/install.yaml.tmpl", {
    os_channel               = var.os_channel
    os_version               = var.os_version
    ignition_endpoint        = format("%s/ignition", var.http_endpoint)
    install_disk             = var.install_disk
    container_linux_oem      = var.container_linux_oem
    ssh_keys                 = jsonencode(var.ssh_keys)
    install_to_smallest_disk = var.install_to_smallest_disk
    kernel_console           = join(" ", var.kernel_console)
    kernel_args              = join(" ", var.kernel_args)
    # only cached-container-linux profile adds -b baseurl
    baseurl_flag = ""
  })
}

// Flatcar Container Linux Install profile (from matchbox /assets cache)
// Note: Admin must have downloaded os_version into matchbox assets/flatcar.
resource "matchbox_profile" "cached-flatcar-linux-install" {
  name = format(
    "cached-flatcar-linux-install-%s",
    var.node_name
  )

  kernel = "/assets/flatcar/${var.os_version}/flatcar_production_pxe.vmlinuz"

  initrd = [
    "/assets/flatcar/${var.os_version}/flatcar_production_pxe_image.cpio.gz",
  ]

  args = flatten([
    "initrd=flatcar_production_pxe_image.cpio.gz",
    "ignition.config.url=${var.http_endpoint}/ignition?uuid=$${uuid}&mac=$${mac:hexhyp}",
    "flatcar.first_boot=yes",
    var.kernel_console,
    var.kernel_args,
  ])

  container_linux_config = templatefile("${path.module}/templates/install.yaml.tmpl", {
    os_channel               = var.os_channel
    os_version               = var.os_version
    ignition_endpoint        = format("%s/ignition", var.http_endpoint)
    install_disk             = var.install_disk
    container_linux_oem      = var.container_linux_oem
    ssh_keys                 = jsonencode(var.ssh_keys)
    install_to_smallest_disk = var.install_to_smallest_disk
    kernel_console           = join(" ", var.kernel_console)
    kernel_args              = join(" ", var.kernel_args)
    # profile uses -b baseurl to install from matchbox cache
    baseurl_flag = "-b ${var.http_endpoint}/assets/flatcar"
  })
}

resource "matchbox_profile" "node" {
  name = format(
    "node-%s",
    var.node_name
  )
  raw_ignition = var.ignition_clc_config
}

