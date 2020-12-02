# Environment part.
locals {
  base_name = "tinkerbell-sandbox-${var.name}"
}

resource "libvirt_network" "network" {
  name = local.base_name

  # These addresses will have NAT setup on the used host.
  addresses = [var.hosts_cidr]

  # Disable DHCP as boots from the Tinkerbell stack will act as one
  # to assign IPs and schedule workflows.
  dhcp {
    enabled = false
  }

  dns {
    # Do not try to resolve all DNS names, only the static ones
    # configured below.
    local_only = true

    # Use 8.8.8.8 as default to follow defaults from Flatcar.
    forwarders {
      address = "8.8.8.8"
    }

    dynamic "hosts" {
      for_each = var.dns_hosts

      content {
        hostname = hosts.value.hostname
        ip       = hosts.value.ip
      }
    }
  }
}

resource "libvirt_pool" "pool" {
  name = local.base_name
  type = "dir"
  path = abspath(var.pool_path)
}

resource "libvirt_volume" "os_base" {
  name   = "os-base"
  source = abspath(var.flatcar_image_path)
  pool   = libvirt_pool.pool.name
  format = "qcow2"
}

# Provisioner.
locals {
  provisioner_netmask = split("/", var.hosts_cidr)[1]
  provisioner_ips = [
    # Tink requires 2 IP addresses.
    # 0 host is network address.
    # 1 host is gateway address.
    # So we start with offset of 2.
    cidrhost(var.hosts_cidr, 2),
    cidrhost(var.hosts_cidr, 3),
  ]
  provisioner_gateway = cidrhost(var.hosts_cidr, 1)

  assets_dir = "${path.module}/assets"
}

data "ct_config" "provisioner" {
  content = ""

  snippets = [
    # Set hostname to easy identify machines.
    <<EOF
storage:
  files:
  - path: /etc/hostname
    filesystem: root
    mode: 0420
    contents:
      inline: provisioner
EOF
    ,
    # Disable updates for stability.
    <<EOF
systemd:
  units:
  - name: locksmithd.service
    mask: true
  - name: update-engine.service
    mask: true
EOF
    ,
    # Enable docker to bring up containers after reboot.
    <<EOF
systemd:
  units:
  - name: docker.service
    enabled: true
EOF
    ,
    # Enable time sync to avoid issues with date and time.
    <<EOF
systemd:
  units:
  - name: systemd-timesyncd.service
    enabled: true
EOF
    ,
    # Set SSH keys for core user to be able to SSH into the machine.
    # Also configure root user as Tink is configured as root.
    <<EOF
passwd:
  users:
  - name: core
    ssh_authorized_keys:
%{for key in var.ssh_keys~}
    - ${key}
%{endfor~}
  - name: root
    ssh_authorized_keys:
%{for key in var.ssh_keys~}
    - ${key}
%{endfor~}
EOF
    ,
    # Configure DNS server as we do not use DHCP server.
    <<EOF
storage:
  files:
    - path: /etc/systemd/resolved.conf.d/dns_servers.conf
      filesystem: root
      mode: 0644
      contents:
        inline: |
          [Resolve]
          DNS=9.9.9.9
          Domains=~.
EOF
    ,
    # Configure interface with both IP addresses as Tinkerbell requires.
    <<EOF
storage:
  files:
    - path: /etc/systemd/network/30-tink.network
      filesystem: root
      mode: 0644
      contents:
        inline: |
          [Match]
          Name=eth0

          [Network]
          Address=${local.provisioner_ips[0]}/${local.provisioner_netmask}
          Address=${local.provisioner_ips[1]}/${local.provisioner_netmask}
          Gateway=${local.provisioner_gateway}
EOF
    ,
  ]
}

resource "libvirt_ignition" "provisioner" {
  name = "provisioner-ignition"
  pool = libvirt_pool.pool.name

  content = data.ct_config.provisioner.rendered
}

resource "libvirt_volume" "provisioner" {
  name           = "provisioner"
  base_volume_id = libvirt_volume.os_base.id
  pool           = libvirt_pool.pool.name
  format         = "qcow2"
  size           = 20 * 1000 * 1000 * 1000 # 20GB
}

resource "libvirt_domain" "provisioner" {
  name   = "${local.base_name}-provisioner"
  vcpu   = 2
  memory = 2048

  disk {
    volume_id = libvirt_volume.provisioner.id
  }

  fw_cfg_name     = "opt/org.flatcar-linux/config"
  coreos_ignition = libvirt_ignition.provisioner.id

  # Setup SPICE server for debugging.
  graphics {
    listen_type = "address"
  }

  network_interface {
    network_id = libvirt_network.network.id
    hostname   = "provisioner"
  }

  # Setup console, so 'virsh console' can be used to debug
  # connectivity issues.
  console {
    type        = "pty"
    target_port = "0"
  }

  connection {
    type = "ssh"
    user = "root"
    host = local.provisioner_ips[0]
  }

  provisioner "remote-exec" {
    inline = [
      "mkdir -p /root/tink/deploy",
    ]
  }

  provisioner "file" {
    source      = "${local.assets_dir}/setup.sh"
    destination = "/root/tink/setup.sh"
  }

  provisioner "file" {
    source      = "${local.assets_dir}/generate-envrc.sh"
    destination = "/root/tink/generate-envrc.sh"
  }

  provisioner "file" {
    source      = "${local.assets_dir}/deploy"
    destination = "/root/tink"
  }

  provisioner "remote-exec" {
    inline = [
      "set -e",
      "set -o pipefail",
      "mkdir -p /opt/bin",
      "test -f /opt/bin/docker-compose || wget https://github.com/docker/compose/releases/download/1.27.4/docker-compose-Linux-x86_64 -O /opt/bin/docker-compose",
      "chmod +x /opt/bin/docker-compose",
      "chmod +x /root/tink/*.sh /root/tink/deploy/tls/*.sh",
      "/root/tink/generate-envrc.sh eth0 > /root/tink/.env",
      "sed -i 's/192.168.1.1/${local.provisioner_ips[0]}/g' /root/tink/.env",
      "sed -i 's/192.168.1.2/${local.provisioner_ips[1]}/g' /root/tink/.env",
      "echo 'export PATH=\"/opt/bin:$PATH\"' >> /root/tink/.env",
      "bash -c 'until test -f /opt/bin/docker-compose; do echo \"Waiting for docker-compose binary\"; sleep 1; done'",
      "bash -c 'cd /root/tink && source .env && ./setup.sh'",
      "bash -c 'cd /root/tink/deploy && source ../.env && docker-compose up -d'",
      "cd /root/tink/deploy/flatcar-install && docker build -t ${local.provisioner_ips[0]}/flatcar-install .",
      "docker push ${local.provisioner_ips[0]}/flatcar-install",
    ]
  }
}
