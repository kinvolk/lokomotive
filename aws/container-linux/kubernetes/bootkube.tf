# Self-hosted Kubernetes assets (kubeconfig, manifests)
module "bootkube" {
  source = "git::ssh://git@github.com/kinvolk/terraform-render-bootkube.git?ref=8deaadbd7e1258b0c51e0e4f98e1f116ec31a966"

  cluster_name          = "${var.cluster_name}"
  api_servers           = ["${format("%s.%s", var.cluster_name, var.dns_zone)}"]
  etcd_servers          = ["${aws_route53_record.etcds.*.fqdn}"]
  asset_dir             = "${var.asset_dir}"
  networking            = "${var.networking}"
  network_mtu           = "${var.network_mtu}"
  pod_cidr              = "${var.pod_cidr}"
  service_cidr          = "${var.service_cidr}"
  cluster_domain_suffix = "${var.cluster_domain_suffix}"
  enable_reporting      = "${var.enable_reporting}"
}
