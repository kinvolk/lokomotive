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
    ]
  }

  triggers = {
    controller_id = packet_device.controllers[count.index].id
  }
}

# Secure copy bootkube assets to ONE controller.
resource "null_resource" "copy-assets-dir" {
  depends_on = [
    module.bootkube,
    aws_route53_record.apiservers,
    null_resource.copy-controller-secrets,
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

// copy the calico hostendpoint controller manifests only when networking is calico
// TODO: convert these templates to helm charts for calico host endpoint controller
// Currently the manfifests are copied over to the calico helm chart.
resource "null_resource" "calico-host-endpoint-manifests" {
  count = var.networking == "calico" ? 1 : 0

  depends_on = [
    module.bootkube,
    aws_route53_record.apiservers,
    null_resource.copy-controller-secrets,
    null_resource.copy-assets-dir,
  ]

  connection {
    type    = "ssh"
    host    = packet_device.controllers[0].access_public_ipv4
    user    = "core"
    timeout = "15m"
  }

  provisioner "file" {
    content     = data.template_file.host_protection_policy.rendered
    destination = "$HOME/assets/charts/kube-system/calico/templates/calico-policy.yaml"
  }

  provisioner "file" {
    source      = "${path.module}/calico/host-endpoint-controller.yaml"
    destination = "$HOME/assets/charts/kube-system/calico/templates/host-endpoint-controller.yaml"
  }
}

# start bootkube to perform one-time self-hosted cluster bootstrapping.
resource "null_resource" "bootkube-start" {
  depends_on = [
    module.bootkube,
    aws_route53_record.apiservers,
    null_resource.copy-controller-secrets,
    null_resource.copy-assets-dir,
    null_resource.calico-host-endpoint-manifests,
  ]

  connection {
    type    = "ssh"
    host    = packet_device.controllers[0].access_public_ipv4
    user    = "core"
    timeout = "15m"
  }

  provisioner "remote-exec" {
    inline = [
      "sudo mv $HOME/assets /opt/bootkube",
      "sudo systemctl start bootkube",
    ]
  }
}

data "template_file" "controller_host_endpoints" {
  count    = var.controller_count
  template = file("${path.module}/calico/controller-host-endpoint.yaml.tmpl")

  vars = {
    node_name = packet_device.controllers[count.index].hostname
  }
}

data "template_file" "host_protection_policy" {
  template = file("${path.module}/calico/host-protection.yaml.tmpl")

  vars = {
    controller_host_endpoints = join(
      "\n",
      data.template_file.controller_host_endpoints.*.rendered,
    )
    management_cidrs       = jsonencode(var.management_cidrs)
    cluster_internal_cidrs = jsonencode([var.node_private_cidr, var.pod_cidr, var.service_cidr])
    etcd_server_cidrs      = jsonencode(packet_device.controllers.*.access_private_ipv4)
  }
}
