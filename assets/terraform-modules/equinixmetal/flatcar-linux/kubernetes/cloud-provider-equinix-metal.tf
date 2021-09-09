resource "local_file" "cloud-provider-equinix-metal" {
  filename = "${var.asset_dir}/charts/kube-system/cloud-provider-equinix-metal.yaml"
  content = templatefile("${path.module}/cloud-provider-equinix-metal.yaml.tmpl", {
    api_key    = var.auth_token
    project_id = var.project_id
  })
}
