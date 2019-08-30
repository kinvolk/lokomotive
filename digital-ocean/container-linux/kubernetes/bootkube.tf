# Self-hosted Kubernetes assets (kubeconfig, manifests)
module "bootkube" {
  source = "github.com/kinvolk/terraform-render-bootkube?ref=7e237ffa21fd85f76ddf2a215073aa7cd6ef2476"

  cluster_name = "${var.cluster_name}"
  api_servers  = ["${format("%s.%s", var.cluster_name, var.dns_zone)}"]
  etcd_servers = "${digitalocean_record.etcds.*.fqdn}"
  asset_dir    = "${var.asset_dir}"

  networking = "${var.networking}"

  # only effective with Calico networking
  network_encapsulation = "vxlan"
  network_mtu           = "1450"

  network_mtu           = 1440
  pod_cidr              = "${var.pod_cidr}"
  service_cidr          = "${var.service_cidr}"
  cluster_domain_suffix = "${var.cluster_domain_suffix}"
  enable_reporting      = "${var.enable_reporting}"
  enable_aggregation    = "${var.enable_aggregation}"
}
