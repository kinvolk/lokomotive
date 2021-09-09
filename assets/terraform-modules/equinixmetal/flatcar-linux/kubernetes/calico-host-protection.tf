// TODO: Currently, bootkube installs all charts in parallel, which causes
// calico-host-protection chart installation to fail occasionally, as it
// gets installed before the Calico chart itself.
//
// To workaround this issue, we make an extra copy of the Calico CRDs into
// the bootkube manifests directory, as those will be applied
// before all Helm charts. This ensures that the CRDs are created before
// the calico-host-protection chart is installed.
//
// See https://github.com/kinvolk/lokomotive/issues/1175 for more details.
resource "local_file" "calico_crds" {
  for_each = fileset("${var.asset_dir}/charts/kube-system/calico/crds", "*.yaml")

  content  = file("${var.asset_dir}/charts/kube-system/calico/crds/${each.value}")
  filename = "${var.asset_dir}/manifests-calico-cdrs/${each.value}"
}

resource "local_file" "calico_host_protection" {
  content = templatefile("${path.module}/calico-host-protection.yaml.tmpl", {
    host_endpoints = [
      for device in metal_device.controllers :
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
    cluster_cidrs = concat([
      var.pod_cidr,
      var.service_cidr
    ], var.node_private_cidrs),
  })

  filename = "${var.asset_dir}/charts/kube-system/calico-host-protection.yaml"
}
