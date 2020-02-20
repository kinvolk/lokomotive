# Assets generated only when certain options are chosen

# Populate flannel chart values file named flannel.yaml.
resource "local_file" "flannel" {
  count    = var.networking == "flannel" ? 1 : 0
  content  = templatefile("${path.module}/resources/charts/flannel.yaml",{
    flannel_image     = "${var.container_images["flannel"]}${var.container_arch}"
    flannel_cni_image = var.container_images["flannel_cni"]
    pod_cidr          = var.pod_cidr
  })
  filename = "${var.asset_dir}/charts/kube-system/flannel.yaml"
}

# Populate flannel chart.
# TODO: Currently, there is no way in Terraform to copy local directory, so we use `template_dir` for it.
# The downside is, that any Terraform templating syntax stored in this directory will be evaluated, which may bring unexpected results.
resource "template_dir" "flannel" {
  count           = var.networking == "flannel" ? 1 : 0
  source_dir      = "${replace(path.module, path.cwd, ".")}/resources/charts/flannel"
  destination_dir = "${var.asset_dir}/charts/kube-system/flannel"
}

# Render flannel.yaml for flannel chart.
data "template_file" "flannel" {
  count    = var.networking == "flannel" ? 1 : 0
  template = "${file("${path.module}/resources/charts/flannel.yaml")}"

  vars = {
    flannel_image     = "${var.container_images["flannel"]}${var.container_arch}"
    flannel_cni_image = var.container_images["flannel_cni"]
    pod_cidr          = var.pod_cidr
  }
}

# Populate calico chart values file named calico.yaml.
resource "local_file" "calico" {
  count    = var.networking == "calico" ? 1 : 0
  content  = templatefile("${path.module}/resources/charts/calico.yaml",{
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
  count           = var.networking == "calico" ? 1 : 0
  source_dir      = "${replace(path.module, path.cwd, ".")}/resources/charts/calico"
  destination_dir = "${var.asset_dir}/charts/kube-system/calico"
}

# Populate kube-router chart values file named kube-router.yaml.
resource "local_file" "kube-router" {
  count    = var.networking == "kube-router" ? 1 : 0
  content  = templatefile("${path.module}/resources/charts/kube-router.yaml",{
    kube_router_image = var.container_images["kube_router"]
    flannel_cni_image = var.container_images["flannel_cni"]
    network_mtu       = var.network_mtu
  })
  filename = "${var.asset_dir}/charts/kube-system/kube-router.yaml"
}

# Populate kube-router chart.
# TODO: Currently, there is no way in Terraform to copy local directory, so we use `template_dir` for it.
# The downside is, that any Terraform templating syntax stored in this directory will be evaluated, which may bring unexpected results.
resource "template_dir" "kube-router" {
  count           = var.networking == "kube-router" ? 1 : 0
  source_dir      = "${replace(path.module, path.cwd, ".")}/resources/charts/kube-router"
  destination_dir = "${var.asset_dir}/charts/kube-system/kube-router"
}
