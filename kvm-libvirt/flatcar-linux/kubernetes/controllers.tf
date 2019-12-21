resource "random_string" "volumepath" {
  length  = 6
  special = false
}

resource "libvirt_pool" "volumetmp" {
  name = "vms${random_string.volumepath.result}"
  type = "dir"
  path = "/var/tmp/vms${random_string.volumepath.result}"
}

resource "libvirt_volume" "base" {
  name   = "${var.cluster_name}-base"
  source = var.os_image_unpacked
  pool   = libvirt_pool.volumetmp.name
  format = "qcow2"
}

resource "libvirt_volume" "controller-disk" {
  name           = "${var.cluster_name}-controller-${count.index}.qcow2"
  count          = var.controller_count
  base_volume_id = libvirt_volume.base.id
  pool           = libvirt_pool.volumetmp.name
  format         = "qcow2"
}

resource "libvirt_ignition" "ignition" {
  name    = "${var.cluster_name}-controller-${count.index}-ignition"
  pool    = libvirt_pool.volumetmp.name
  count   = var.controller_count
  content = element(data.ct_config.controller-ignitions.*.rendered, count.index)
}

resource "libvirt_network" "vmnet" {
  name      = var.cluster_name
  mode      = "nat"
  domain    = var.machine_domain
  addresses = [var.node_ip_pool]

  dns {
    local_only = true
    # can specify local names here
  }
}

resource "libvirt_domain" "controller-machine" {
  count  = var.controller_count
  name   = "${var.cluster_name}-controller-${count.index}"
  vcpu   = var.virtual_cpus
  memory = var.virtual_memory

  fw_cfg_name     = "opt/org.flatcar-linux/config"
  coreos_ignition = element(libvirt_ignition.ignition.*.id, count.index)

  disk {
    volume_id = element(libvirt_volume.controller-disk.*.id, count.index)
  }

  graphics {
    listen_type = "address"
  }

  network_interface {
    network_id     = libvirt_network.vmnet.id
    hostname       = "${var.cluster_name}-controller-${count.index}"
    addresses      = [cidrhost(var.node_ip_pool, 10 + count.index)] # TODO: use as public addr in kubeconfig
    wait_for_lease = true
  }
}

data "ct_config" "controller-ignitions" {
  count = var.controller_count
  content = element(
    data.template_file.controller-configs.*.rendered,
    count.index,
  )
  snippets = var.controller_clc_snippets
}

data "template_file" "controller-configs" {
  count    = var.controller_count
  template = file("${path.module}/cl/controller.yaml.tmpl")

  vars = {
    # Cannot use cyclic dependencies on controllers or their DNS records
    etcd_name              = "${var.cluster_name}-controller-${count.index}"
    etcd_domain            = "${var.cluster_name}-controller-${count.index}.${var.machine_domain}"
    etcd_initial_cluster   = join(",", data.template_file.etcds.*.rendered)
    kubeconfig             = indent(10, module.bootkube.kubeconfig-kubelet)
    ssh_keys               = jsonencode(var.ssh_keys)
    cluster_dns_service_ip = cidrhost(var.service_cidr, 10)
    cluster_domain_suffix  = var.cluster_domain_suffix
  }
}

data "template_file" "etcds" {
  count    = var.controller_count
  template = "$${cluster_name}-controller-$${index}=https://$${cluster_name}-controller-$${index}.${var.machine_domain}:2380" # as etcd_domain above

  vars = {
    index        = count.index
    cluster_name = var.cluster_name
  }
}
