resource "libvirt_volume" "worker-disk" {
  name           = "${var.cluster_name}-${var.pool_name}-worker-${count.index}.qcow2"
  count          = var.worker_count
  base_volume_id = var.libvirtbaseid
  pool           = var.libvirtpool
  format         = "qcow2"
}

resource "libvirt_ignition" "ignition" {
  name    = "${var.cluster_name}-${var.pool_name}-worker-${count.index}-ignition"
  pool    = var.libvirtpool
  count   = var.worker_count
  content = element(data.ct_config.worker-ignition.*.rendered, count.index)
}

resource "libvirt_domain" "worker-machine" {
  count  = var.worker_count
  name   = "${var.cluster_name}-${var.pool_name}-worker-${count.index}"
  vcpu   = var.virtual_cpus
  memory = var.virtual_memory

  fw_cfg_name     = "opt/org.flatcar-linux/config"
  coreos_ignition = element(libvirt_ignition.ignition.*.id, count.index)

  disk {
    volume_id = element(libvirt_volume.worker-disk.*.id, count.index)
  }

  graphics {
    listen_type = "address"
  }

  network_interface {
    network_name   = var.cluster_name
    hostname       = "${var.cluster_name}-${var.pool_name}-worker-${count.index}"
    wait_for_lease = true
  }
}

data "ct_config" "worker-ignition" {
  count    = var.worker_count
  content  = element(data.template_file.worker-config.*.rendered, count.index)
  snippets = var.clc_snippets
}

data "template_file" "worker-config" {
  count    = var.worker_count
  template = file("${path.module}/cl/worker.yaml.tmpl")

  vars = {
    domain_name            = "${var.cluster_name}-${var.pool_name}-worker-${count.index}.${var.machine_domain}"
    kubeconfig             = indent(10, var.kubeconfig)
    ssh_keys               = jsonencode(var.ssh_keys)
    cluster_dns_service_ip = cidrhost(var.service_cidr, 10)
    cluster_domain_suffix  = var.cluster_domain_suffix
  }
}

