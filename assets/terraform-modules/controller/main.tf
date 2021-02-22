locals {
  kubelet_require_kubeconfig = <<EOF
systemd:
  units:
  - name: kubelet.service
    dropins:
    - name: 10-controller.conf
      contents: |
        [Service]
        ConditionPathExists=/etc/kubernetes/kubeconfig
        ExecStartPre=/bin/mkdir -p /etc/kubernetes/checkpoint-secrets
        ExecStartPre=/bin/mkdir -p /etc/kubernetes/inactive-manifests
EOF

  bootkube = templatefile("${path.module}/templates/bootkube.yaml.tmpl", {
    bootkube_image_name = var.bootkube_image_name
    bootkube_image_tag  = var.bootkube_image_tag
    kubelet_image_name  = var.kubelet_image_name
    kubelet_image_tag   = var.kubelet_image_tag
  })

  etcd_servers = [for i in range(var.controllers_count) : format("%s-etcd%d.%s", var.cluster_name, i, var.dns_zone)]

  etcd = templatefile("${path.module}/templates/etcd.yaml.tmpl", {
    etcd_name   = "etcd${var.count_index}"
    etcd_domain = "${var.cluster_name}-etcd${var.count_index}.${var.dns_zone}"

    # etcd0=https://cluster-etcd0.example.com,etcd1=https://cluster-etcd1.example.com,...
    etcd_initial_cluster = join(",", [for i, server in local.etcd_servers : format("etcd%d=https://%s:2380", i, server)])
  })

  snippets = [
    local.kubelet_require_kubeconfig,
    local.bootkube,
    local.etcd,
  ]
}

data "ct_config" "config" {
  pretty_print = false

  content = templatefile("${path.module}/templates/node.yaml.tmpl", {
    ssh_keys                  = jsonencode(var.ssh_keys)
    cluster_dns_service_ip    = var.cluster_dns_service_ip
    cluster_domain_suffix     = var.cluster_domain_suffix
    kubelet_image_name        = var.kubelet_image_name
    kubelet_image_tag         = var.kubelet_image_tag
    kubelet_docker_extra_args = []
    hostname                  = var.set_standard_hostname == true ? "${var.cluster_name}-controller-${var.count_index}" : ""
    kubelet_labels = {
      "node.kubernetes.io/master"     = "",
      "node.kubernetes.io/controller" = "true",
    }
    kubelet_taints = {
      "node-role.kubernetes.io/master" = ":NoSchedule"
    }
  })

  snippets = concat(local.snippets, var.clc_snippets)
}
