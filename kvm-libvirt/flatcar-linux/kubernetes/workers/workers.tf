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
  content = data.ct_config.worker-ignition[count.index].rendered
}

resource "libvirt_domain" "worker-machine" {
  count  = var.worker_count
  name   = "${var.cluster_name}-${var.pool_name}-worker-${count.index}"
  vcpu   = var.virtual_cpus
  memory = var.virtual_memory

  fw_cfg_name     = "opt/org.flatcar-linux/config"
  coreos_ignition = libvirt_ignition.ignition[count.index].id

  disk {
    volume_id = libvirt_volume.worker-disk[count.index].id
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
  content  = data.template_file.worker-config[count.index].rendered
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
