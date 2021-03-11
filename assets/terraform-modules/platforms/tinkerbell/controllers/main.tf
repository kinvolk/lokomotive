module "controller" {
  source = "../../../controller"

  count = var.node_count

  cluster_name           = var.cluster_name
  controllers_count      = var.node_count
  dns_zone               = var.dns_zone
  count_index            = count.index
  cluster_dns_service_ip = module.bootkube.cluster_dns_service_ip
  ssh_keys               = var.ssh_keys
  cluster_domain_suffix  = var.cluster_domain_suffix
  host_dns_ip            = var.host_dns_ip
  apiserver              = format("%s.%s", var.cluster_name, var.dns_zone)
  ca_cert                = module.bootkube.ca_cert

  clc_snippets = concat(var.clc_snippets, [
    <<EOF
storage:
  files:
  - path: /etc/hostname
    filesystem: root
    mode: 0644
    contents:
      inline: |
        ${var.cluster_name}-controller-${count.index}
EOF
    ,
  ])
}

resource "tinkerbell_template" "main" {
  count = var.node_count

  name = "${var.cluster_name}-controller-${count.index}"

  content = templatefile("${path.module}/templates/flatcar-install.tmpl", {
    ignition_config          = module.controller[count.index].clc_config
    flatcar_install_base_url = var.flatcar_install_base_url
    os_version               = var.os_version
    os_channel               = var.os_channel
  })
}

resource "tinkerbell_workflow" "main" {
  count = var.node_count

  hardwares = <<EOF
{"device_1": "${var.ip_addresses[count.index]}"}
EOF
  template  = tinkerbell_template.main[count.index].id
}
