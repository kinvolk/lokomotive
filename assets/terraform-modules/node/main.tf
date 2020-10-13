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
    host_dns_ip               = var.host_dns_ip
  })

  snippets = concat(var.clc_snippets, [
    # Allow to pass unique snippets per controller node. For example, to set the hostname.
    var.clc_snippet_index != "" ? format(var.clc_snippet_index, count.index) : "",
  ])
}
