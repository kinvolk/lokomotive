# Assets generated only when certain options are chosen

# Populate calico chart values file named calico.yaml.
resource "local_file" "calico" {
  content = templatefile("${path.module}/resources/charts/calico.yaml", {
    calico_image                    = var.container_images["calico"]
    calico_cni_image                = var.container_images["calico_cni"]
    calico_controllers_image        = var.container_images["calico_controllers"]
    flexvol_driver_image            = var.container_images["flexvol_driver_image"]
    network_mtu                     = var.network_mtu
    network_encapsulation           = indent(2, var.network_encapsulation == "vxlan" ? "vxlanMode: Always" : "ipipMode: Always")
    ipip_enabled                    = var.network_encapsulation == "ipip" ? true : false
    vxlan_enabled                   = var.network_encapsulation == "vxlan" ? true : false
    network_ip_autodetection_method = var.network_ip_autodetection_method
    pod_cidr                        = var.pod_cidr
    enable_reporting                = var.enable_reporting
    blocked_metadata_cidrs          = var.blocked_metadata_cidrs
  })
  filename = "${var.asset_dir}/charts/kube-system/calico.yaml"
}
