resource "local_file" "calico_host_protection" {
  content = templatefile("${path.module}/calico-host-protection.yaml.tmpl", {
    host_endpoints = [
      for device in packet_device.controllers :
      {
        name           = "${device.hostname}-bond0",
        node_name      = device.hostname,
        interface_name = "bond0",
        labels = {
          "host-endpoint" = "ingress"
          "nodetype"      = "controller"
        }
      }
    ],
    management_cidrs = var.management_cidrs
    cluster_cidrs = [
      var.node_private_cidr,
      var.pod_cidr,
      var.service_cidr
    ],
  })

  filename = "${var.asset_dir}/charts/kube-system/calico-host-protection.yaml"
}

# Populate calico-host-protection chart.
# TODO: Currently, there is no way in Terraform to copy local directory, so we use `template_dir` for it.
# The downside is, that any Terraform templating syntax stored in this directory will be evaluated, which may bring unexpected results.
resource "template_dir" "calico_host_protection" {
  source_dir      = "${path.module}/calico-host-protection"
  destination_dir = "${var.asset_dir}/charts/kube-system/calico-host-protection"
}
