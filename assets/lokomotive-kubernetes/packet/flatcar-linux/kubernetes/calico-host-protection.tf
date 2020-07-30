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
