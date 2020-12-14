locals {
  manage_iscsid_service = <<EOF
systemd:
  units:
    - name: iscsid.service
      enable: true
      enabled: true
      dropins:
      - name: 00-iscsid.conf
        contents: |
          [Service]
          ExecStartPre=/bin/bash -c 'echo "InitiatorName=$(/sbin/iscsi-iname -p iqn.2020-01.io.kinvolk:01)" > /etc/iscsi/initiatorname.iscsi'
EOF

  kubeconfig = <<EOF
storage:
  files:
    - path: /etc/kubernetes/kubeconfig
      filesystem: root
      mode: 0600
      contents:
        inline: |
          ${indent(10, local.bootstrap_kubeconfig)}
EOF

  snippets = [
    local.manage_iscsid_service,
    local.kubeconfig,
  ]
}

data "ct_config" "config" {
  pretty_print = false

  content = templatefile("${path.module}/templates/node.yaml.tmpl", {
    ssh_keys               = jsonencode(var.ssh_keys)
    cluster_dns_service_ip = var.cluster_dns_service_ip
    cluster_domain_suffix  = var.cluster_domain_suffix
    host_dns_ip            = var.host_dns_ip
    kubelet_docker_extra_args = []
    # Here we set default labels for worker nodes.
    kubelet_labels = length(var.kubelet_labels) > 0 ? var.kubelet_labels : {
      "node.kubernetes.io/node" = ""
    }
    kubelet_taints = var.kubelet_taints

    kubelet_image_name = var.kubelet_image_name != "" ? var.kubelet_image_name : null
    kubelet_image_tag  = var.kubelet_image_tag != "" ? var.kubelet_image_tag : null
  })

  snippets = concat(local.snippets, var.clc_snippets)
}
