resource "libvirt_volume" "worker" {
  name = var.name
  pool = var.sandbox.volumes_pool_name
  size = 20 * 1000 * 1000 * 1000 # 20GB should be sufficient for Kubernetes node.
}

# Generate random numbers in hex format to generate worker MAC address.
resource "random_id" "mac_address" {
  count       = 5
  byte_length = 1
}

locals {
  # 52 is a Locally Administered Address Range, so it's safe to use.
  mac = "52:${join(":", random_id.mac_address.*.hex)}"
}

resource "libvirt_domain" "worker" {
  name = "${var.sandbox.sandbox_name}-${var.name}"
  # 4 vCPUs and 4GB of RAM should be sufficient for both Kubernetes controller and worker
  # node, so to keep things simple we hardcode these values.
  vcpu   = 4
  memory = 4096

  disk {
    volume_id = libvirt_volume.worker.id
  }

  # Setup SPICE server for debugging.
  graphics {
    listen_type = "address"
  }

  # Primarily boot from disk, but also allow network boot to run Tinkerbell workflow,
  # which should install host OS.
  boot_device {
    dev = ["hd", "network"]
  }

  # Setup console, so 'virsh console' can be used to debug
  # connectivity issues.
  console {
    type        = "pty"
    target_port = "0"
  }

  network_interface {
    network_id = var.sandbox.network_id
    hostname   = var.name
    mac        = local.mac
  }
}

resource "random_uuid" "worker" {}

resource "tinkerbell_hardware" "worker" {
  data = templatefile("${path.module}/assets/hardware_data.tpl", {
    id            = random_uuid.worker.result
    facility_code = "libvirt"
    plan_slug     = "libvirt"
    address       = var.ip
    mac           = local.mac
    gateway       = var.sandbox.gateway
    netmask       = var.sandbox.netmask
  })
}
