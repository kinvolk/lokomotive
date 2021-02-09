module "worker" {
  source = "../../../worker"

  count  = length(var.worker_names)
  count_index = count.index

  cluster_dns_service_ip = module.bootkube.cluster_dns_service_ip
  ssh_keys               = var.ssh_keys
  cluster_domain_suffix  = var.cluster_domain_suffix
  ca_cert                = module.bootkube.ca_cert
  apiserver              = format("%s.%s", var.cluster_name, var.k8s_domain_name)
  kubelet_labels         = var.labels

  clc_snippets = concat(var.clc_snippets[var.worker_names[count.index]], [
    <<EOF
storage:
  files:
  - path: /etc/hostname
    filesystem: root
    mode: 0644
    contents:
      inline: |
        ${var.cluster_name}-worker-${count.index}
EOF
    ,
  ])
}

