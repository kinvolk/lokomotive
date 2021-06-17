module "worker_profile" {
  source                   = "../../../matchbox-flatcar"
  count                    = length(var.worker_names)
  node_name                = var.worker_names[count.index]
  node_mac                 = var.worker_macs[count.index]
  download_protocol        = var.download_protocol
  os_channel               = var.os_channel
  os_version               = var.os_version
  http_endpoint            = var.matchbox_http_endpoint
  kernel_args              = var.kernel_args
  kernel_console           = var.kernel_console
  install_disk             = var.install_disk
  install_to_smallest_disk = var.install_to_smallest_disk
  container_linux_oem      = var.container_linux_oem
  ssh_keys                 = var.ssh_keys
  ignition_clc_config      = module.worker[count.index].clc_config
  cached_install           = var.cached_install
}
