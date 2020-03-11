# Assets generated only when certain options are chosen

# Populate calico chart values file named calico.yaml.
resource "local_file" "calico" {
  content = templatefile("${path.module}/resources/charts/calico.yaml", {
    calico_image                    = var.container_images["calico"]
    calico_cni_image                = var.container_images["calico_cni"]
    network_mtu                     = var.network_mtu
    network_encapsulation           = indent(2, var.network_encapsulation == "vxlan" ? "vxlanMode: Always" : "ipipMode: Always")
    ipip_enabled                    = var.network_encapsulation == "ipip" ? true : false
    ipip_readiness                  = var.network_encapsulation == "ipip" ? indent(16, "- --bird-ready") : ""
    vxlan_enabled                   = var.network_encapsulation == "vxlan" ? true : false
    network_ip_autodetection_method = var.network_ip_autodetection_method
    pod_cidr                        = var.pod_cidr
    enable_reporting                = var.enable_reporting
  })
  filename = "${var.asset_dir}/charts/kube-system/calico.yaml"
}

# Populate calico chart.
# TODO: Currently, there is no way in Terraform to copy local directory, so we use `template_dir` for it.
# The downside is, that any Terraform templating syntax stored in this directory will be evaluated, which may bring unexpected results.
resource "template_dir" "calico" {
  source_dir      = "${replace(path.module, path.cwd, ".")}/resources/charts/calico"
  destination_dir = "${var.asset_dir}/charts/kube-system/calico"
}

# Render calico.yaml for calico chart.
data "template_file" "calico" {
  template = "${file("${path.module}/resources/charts/calico.yaml")}"

  vars = {
    calico_image                    = var.container_images["calico"]
    calico_cni_image                = var.container_images["calico_cni"]
    network_mtu                     = var.network_mtu
    network_encapsulation           = indent(2, var.network_encapsulation == "vxlan" ? "vxlanMode: Always" : "ipipMode: Always")
    ipip_enabled                    = var.network_encapsulation == "ipip" ? true : false
    ipip_readiness                  = var.network_encapsulation == "ipip" ? indent(16, "- --bird-ready") : ""
    vxlan_enabled                   = var.network_encapsulation == "vxlan" ? true : false
    network_ip_autodetection_method = var.network_ip_autodetection_method
    pod_cidr                        = var.pod_cidr
    enable_reporting                = var.enable_reporting
  }
}
