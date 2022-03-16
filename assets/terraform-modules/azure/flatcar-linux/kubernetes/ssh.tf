# Secure copy etcd TLS assets to controllers.
resource "null_resource" "copy-controller-secrets" {
  count = var.controller_count

  depends_on = [azurerm_linux_virtual_machine.controllers]

  connection {
    type    = "ssh"
    host    = azurerm_linux_virtual_machine.controllers[count.index].public_ip_address
    user    = "core"
    timeout = "15m"
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

  provisioner "file" {
    content = var.enable_tls_bootstrap ? templatefile("${path.module}/workers/cl/bootstrap-kubeconfig.yaml.tmpl", {
      token_id     = random_string.bootstrap_token_id[0].result
      token_secret = random_string.bootstrap_token_secret[0].result
      ca_cert      = module.bootkube.ca_cert
      server       = "https://${local.api_server}:6443"
    }) : module.bootkube.kubeconfig-kubelet

    destination = "$HOME/kubeconfig"
  }

  provisioner "remote-exec" {
    inline = [
      "set -e",
      "sudo mv $HOME/kubeconfig /etc/kubernetes/kubeconfig",
      "sudo chown root:root /etc/kubernetes/kubeconfig",
      "sudo chmod 600 /etc/kubernetes/kubeconfig",
      "sudo systemctl stop etcd",
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
      # Use stdbuf to disable the buffer while printing logs to make sure everything is transmitted back to
      # Terraform before we return error. We should be able to remove it once
      # https://github.com/hashicorp/terraform/issues/27121 is resolved.
      "sudo systemctl start etcd || (stdbuf -i0 -o0 -e0 sudo journalctl -u etcd --no-pager; exit 1)",
    ]
  }

  triggers = {
    etcd_ca_cert     = module.bootkube.etcd_ca_cert
    etcd_server_cert = module.bootkube.etcd_server_cert
    etcd_peer_cert   = module.bootkube.etcd_peer_cert
  }
}

# Secure copy bootkube assets to ONE controller and start bootkube to perform
# one-time self-hosted cluster bootstrapping.
resource "null_resource" "bootkube-start" {
  depends_on = [
    module.bootkube,
    null_resource.copy-controller-secrets,
  ]

  connection {
    type    = "ssh"
    host    = azurerm_linux_virtual_machine.controllers[0].public_ip_address
    user    = "core"
    timeout = "15m"
  }

  provisioner "file" {
    source      = var.asset_dir
    destination = "$HOME/assets"
  }

  provisioner "remote-exec" {
    inline = [
      "sudo mv $HOME/assets /opt/bootkube",
      # Use stdbuf to disable the buffer while printing logs to make sure everything is transmitted back to
      # Terraform before we return error. We should be able to remove it once
      # https://github.com/hashicorp/terraform/issues/27121 is resolved.
      "sudo systemctl start bootkube || (stdbuf -i0 -o0 -e0 sudo journalctl -u bootkube --no-pager; exit 1)",
    ]
  }
}
