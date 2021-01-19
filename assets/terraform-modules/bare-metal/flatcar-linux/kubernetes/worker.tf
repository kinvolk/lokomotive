module "worker" {
  source                 = "../../../worker"
  count                  = length(var.worker_names)
  count_index            = count.index
  cluster_dns_service_ip = module.bootkube.cluster_dns_service_ip
  ssh_keys               = var.ssh_keys
  cluster_domain_suffix  = var.cluster_domain_suffix
  ca_cert                = module.bootkube.ca_cert
  apiserver              = format("%s.%s", var.cluster_name, var.k8s_domain_name)
  kubelet_labels         = var.labels
  cluster_name           = var.cluster_name
  set_standard_hostname  = true
  clc_snippets           = concat(lookup(var.clc_snippets, var.worker_names[count.index], []), [
    <<EOF
filesystems:
  - name: root
    mount:
      device: /dev/disk/by-label/ROOT
      format: ext4
      wipe_filesystem: true
      label: ROOT
EOF
    ,
  ])
}

