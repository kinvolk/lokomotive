module "worker" {
  source = "../../../worker"

  count       = var.node_count
  count_index = count.index

  cluster_dns_service_ip = var.cluster_dns_service_ip
  ssh_keys               = var.ssh_keys
  cluster_domain_suffix  = var.cluster_domain_suffix
  ca_cert                = var.ca_cert
  apiserver              = var.apiserver
  set_standard_hostname  = true
  clc_snippets = concat(var.clc_snippets, [
    <<EOF
storage:
  files:
  - path: /etc/systemd/resolved.conf.d/dns_servers.conf
    filesystem: root
    mode: 0644
    contents:
      inline: |
        [Resolve]
        DNS=${var.host_dns_ip}
        Domains=~.
EOF
    ,
  ])
}

resource "tinkerbell_template" "main" {
  count = var.node_count

  name = "${var.cluster_name}-worker-${var.name}-${count.index}"

  content = templatefile("${path.module}/templates/flatcar-install.tmpl", {
    ignition_config          = module.worker[count.index].clc_config
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
