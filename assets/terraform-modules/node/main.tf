data "ct_config" "config" {
  count = var.node_count

  pretty_print = false

  content = templatefile("${path.module}/templates/node.yaml.tmpl", {
    ssh_keys                  = jsonencode(var.ssh_keys)
    cluster_dns_service_ip    = var.cluster_dns_service_ip
    cluster_domain_suffix     = var.cluster_domain_suffix
    kubelet_image_name        = var.kubelet_image_name
    kubelet_image_tag         = var.kubelet_image_tag
    kubelet_labels            = var.kubelet_labels
    kubelet_taints            = var.kubelet_taints
    kubelet_docker_extra_args = var.kubelet_docker_extra_args
    hostname                  = ""
  })

  snippets = var.clc_snippets
}
