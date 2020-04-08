resource "local_file" "cloud_controller_manager" {
  filename = "${var.asset_dir}/charts/kube-system/cloud-controller-manager.yaml"
}

resource "template_dir" "cloud_controller_manager" {
  source_dir      = "${path.module}/cloud-controller-manager"
  destination_dir = "${var.asset_dir}/charts/kube-system/cloud-controller-manager"
}