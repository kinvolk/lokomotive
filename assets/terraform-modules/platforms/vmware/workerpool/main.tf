locals {
  nodes_mask = split("/", var.hosts_cidr)[1]
  nodes_gw   = cidrhost(var.hosts_cidr, 1)
}

module "worker" {
  source = "../../../worker"

  count       = var.node_count
  count_index = count.index

  cluster_dns_service_ip = var.cluster_dns_service_ip
  ssh_keys               = var.ssh_keys
  cluster_domain_suffix  = var.cluster_domain_suffix
  host_dns_ip            = var.host_dns_ip
  ca_cert                = var.ca_cert
  apiserver              = var.apiserver

  clc_snippets = concat(var.clc_snippets, [
    <<EOF
storage:
  files:
  - path: /etc/hostname
    filesystem: root
    mode: 0644
    contents:
      inline: |
        ${var.cluster_name}-worker-${var.name}-${count.index}
EOF
    ,
    <<EOF
networkd:
  units:
  - name: 00-ens192.network
    contents: |
      [Match]
      Name=ens192

      [Network]
      Address=${var.nodes_ips[count.index]}/${local.nodes_mask}
      Gateway=${local.nodes_gw}
EOF
    ,
    ],
  )
}

resource "vsphere_virtual_machine" "main" {
  count = var.node_count

  name             = "${var.cluster_name}-worker-${var.name}-${count.index}"
  resource_pool_id = data.vsphere_compute_cluster.main.resource_pool_id
  datastore_id     = data.vsphere_datastore.main.id
  folder           = var.folder


  num_cpus = var.cpus_count
  memory   = var.memory
  guest_id = data.vsphere_virtual_machine.main_template.guest_id

  network_interface {
    network_id = data.vsphere_network.main.id
  }

  disk {
    label            = "disk0"
    size             = var.disk_size
    eagerly_scrub    = data.vsphere_virtual_machine.main_template.disks[0].eagerly_scrub
    thin_provisioned = data.vsphere_virtual_machine.main_template.disks[0].thin_provisioned
  }

  clone {
    template_uuid = data.vsphere_virtual_machine.main_template.id
  }

  extra_config = {
    "guestinfo.ignition.config.data.encoding" = "base64"
    "guestinfo.ignition.config.data"          = base64encode(module.worker[count.index].clc_config)
  }
}
