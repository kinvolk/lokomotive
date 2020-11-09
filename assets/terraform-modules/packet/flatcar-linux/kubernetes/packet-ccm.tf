resource "local_file" "packet-ccm" {
  filename = "${var.asset_dir}/charts/kube-system/packet-ccm.yaml"
  content = templatefile("${path.module}/packet-ccm.yaml.tmpl", {
    api_key    = var.auth_token
    project_id = var.project_id
  })
}
