module "worker" {
  source                 = "../../../worker"
  count                  = length(var.worker_names)
  count_index            = count.index
  cluster_dns_service_ip = module.bootkube.cluster_dns_service_ip
  ssh_keys               = var.ssh_keys
  cluster_domain_suffix  = var.cluster_domain_suffix
  ca_cert                = module.bootkube.ca_cert
  apiserver              = format("%s.%s", var.cluster_name, var.k8s_domain_name)
  kubelet_labels         = merge(lookup(var.node_specific_labels, var.worker_names[count.index], {}), var.labels)
  cluster_name           = var.cluster_name
  set_standard_hostname  = false
  clc_snippets = concat(lookup(var.clc_snippets, var.worker_names[count.index], []), [
    <<EOF
filesystems:
  - name: root
    mount:
      device: /dev/disk/by-label/ROOT
      format: ext4
      wipe_filesystem: true
      label: ROOT
storage:
  files:
    - path: /ignition_ran
      filesystem: root
      mode: 0644
      contents:
        inline: |
          Flag file indicating that Ignition ran.
          Should be deleted by the SSH step that checks it.
    - path: /etc/hostname
      filesystem: root
      mode: 0644
      contents:
        inline: ${var.worker_names[count.index]}
EOF
    ,
  ])
}

