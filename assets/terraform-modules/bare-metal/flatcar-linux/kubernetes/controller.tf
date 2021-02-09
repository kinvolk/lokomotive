module "controller" {
  source = "../../../controller"
  count  = length(var.controller_names)
  cluster_name = var.cluster_name
  controllers_count = length(var.controller_names)
  dns_zone = var.k8s_domain_name
  count_index = count.index
  cluster_dns_service_ip = module.bootkube.cluster_dns_service_ip
  ssh_keys = var.ssh_keys
  apiserver              = format("%s.%s", var.cluster_name, var.k8s_domain_name)

  ca_cert = module.bootkube.ca_cert
  clc_snippets = concat(var.clc_snippets[var.controller_names[count.index]], [
    <<EOF
storage:
  files:
  - path: /etc/hostname
    filesystem: root
    mode: 0644
    contents:
      inline: |
        ${var.cluster_name}-controller${count.index}
EOF
    ,
  ])
}
