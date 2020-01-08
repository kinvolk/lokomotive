output "kubeconfig-admin" {
  value = module.bootkube.kubeconfig-admin
}

output "kubeconfig" {
  value = module.bootkube.kubeconfig-kubelet
}

output "dns_entries" {
  value = concat(
    # etcd
    [
      for device in packet_device.controllers:
      {
        name    = local.etcd_fqdn[index(packet_device.controllers, device)],
        type    = "A",
        ttl     = 300,
        records = [device.access_private_ipv4],
      }
    ],
    [
      # apiserver public
      {
        name    = local.api_external_fqdn
        type    = "A",
        ttl     = 300,
        records = packet_device.controllers.*.access_public_ipv4,
      },
      # apiserver private
      {
        name    = local.api_fqdn,
        type    = "A",
        ttl     = 300,
        records = packet_device.controllers.*.access_private_ipv4,
      },
    ]
  )
}
