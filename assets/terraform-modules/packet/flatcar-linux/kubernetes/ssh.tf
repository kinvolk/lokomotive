# Secure copy etcd TLS assets to controllers.
resource "null_resource" "copy-controller-secrets" {
  count = var.controller_count

  connection {
    type    = "ssh"
    host    = packet_device.controllers[count.index].access_public_ipv4
    user    = "core"
    timeout = "60m"
  }

  provisioner "file" {
    content     = module.bootkube.etcd_ca_cert
    destination = "$HOME/etcd-client-ca.crt"
  }

  provisioner "file" {
    content     = module.bootkube.etcd_client_cert
    destination = "$HOME/etcd-client.crt"
  }

  provisioner "file" {
    content     = module.bootkube.etcd_client_key
    destination = "$HOME/etcd-client.key"
  }

  provisioner "file" {
    content     = module.bootkube.etcd_server_cert
    destination = "$HOME/etcd-server.crt"
  }

  provisioner "file" {
    content     = module.bootkube.etcd_server_key
    destination = "$HOME/etcd-server.key"
  }

  provisioner "file" {
    content     = module.bootkube.etcd_peer_cert
    destination = "$HOME/etcd-peer.crt"
  }

  provisioner "file" {
    content     = module.bootkube.etcd_peer_key
    destination = "$HOME/etcd-peer.key"
  }

  provisioner "remote-exec" {
    inline = [
      "set -e",
      # Using "etcd/." copies the etcd/ folder recursively in an idempotent
      # way. See https://unix.stackexchange.com/a/228637 for details.
      "[ -d /etc/ssl/etcd ] && sudo cp -R /etc/ssl/etcd/. /etc/ssl/etcd.old && sudo rm -rf /etc/ssl/etcd",
      "sudo mkdir -p /etc/ssl/etcd/etcd",
      "sudo mv etcd-client* /etc/ssl/etcd/",
      "sudo cp /etc/ssl/etcd/etcd-client-ca.crt /etc/ssl/etcd/etcd/server-ca.crt",
      "sudo mv etcd-server.crt /etc/ssl/etcd/etcd/server.crt",
      "sudo mv etcd-server.key /etc/ssl/etcd/etcd/server.key",
      "sudo cp /etc/ssl/etcd/etcd-client-ca.crt /etc/ssl/etcd/etcd/peer-ca.crt",
      "sudo mv etcd-peer.crt /etc/ssl/etcd/etcd/peer.crt",
      "sudo mv etcd-peer.key /etc/ssl/etcd/etcd/peer.key",
      "sudo chown -R etcd:etcd /etc/ssl/etcd",
      "sudo chmod -R 500 /etc/ssl/etcd",
      "sudo systemctl restart etcd",
    ]
  }

  triggers = {
    etcd_ca_cert     = module.bootkube.etcd_ca_cert
    etcd_server_cert = module.bootkube.etcd_server_cert
    etcd_peer_cert   = module.bootkube.etcd_peer_cert
  }
}

# Secure copy bootkube assets to ONE controller.
resource "null_resource" "copy-assets-dir" {
  depends_on = [
    module.bootkube,
    null_resource.copy-controller-secrets,
    local_file.calico_host_protection,
    local_file.calico_crds,
    local_file.packet-ccm,
  ]

  connection {
    type    = "ssh"
    host    = packet_device.controllers[0].access_public_ipv4
    user    = "core"
    timeout = "15m"
  }

  provisioner "file" {
    source      = var.asset_dir
    destination = "$HOME/assets"
  }
}

# start bootkube to perform one-time self-hosted cluster bootstrapping.
resource "null_resource" "bootkube-start" {
  depends_on = [
    module.bootkube,
    null_resource.copy-controller-secrets,
    null_resource.copy-assets-dir,
  ]

  connection {
    type    = "ssh"
    host    = packet_device.controllers[0].access_public_ipv4
    user    = "core"
    timeout = "15m"
  }

  provisioner "remote-exec" {
    inline = [
      "sudo mv $HOME/assets /opt/bootkube/",
      # This is needed, as the bootkube-start script will move all files matching
      # /opt/bootkube/asssets/manifests-*/*.yaml into /opt/bootkube/assets/manifests.
      "sudo mkdir /opt/bootkube/assets/manifests",
      # Use stdbuf to disable the buffer while printing logs to make sure everything is transmitted back to
      # Terraform before we return error. We should be able to remove it once
      # https://github.com/hashicorp/terraform/issues/27121 is resolved.
      "sudo systemctl start bootkube || (stdbuf -i0 -o0 -e0 sudo journalctl -u bootkube --no-pager; exit 1)",
    ]
  }
}
