module "bootkube" {
  source = "git::https://github.com/kinvolk/terraform-render-bootkube?ref=e369831758e871a4c133c4a6ae9d4b2cd8865053"

  cluster_name = "${var.cluster_name}"

  # Cannot use cyclic dependencies on controllers or their DNS records
  api_servers                     = ["${format("%s.%s", var.cluster_name, var.dns_zone)}"]
  etcd_servers                    = "${aws_route53_record.etcds.*.name}"
  asset_dir                       = "${var.asset_dir}"
  networking                      = "${var.networking}"
  network_mtu                     = "${var.network_mtu}"
  network_ip_autodetection_method = "${var.network_ip_autodetection_method}"
  pod_cidr                        = "${var.pod_cidr}"
  service_cidr                    = "${var.service_cidr}"
  cluster_domain_suffix           = "${var.cluster_domain_suffix}"
  enable_reporting                = "${var.enable_reporting}"
}
