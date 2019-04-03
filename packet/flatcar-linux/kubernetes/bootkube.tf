module "bootkube" {
  source = "github.com/kinvolk/terraform-render-bootkube?ref=6ef9683afb6d15d0a9545cbb9d478ea5b9528f41"

  cluster_name = "${var.cluster_name}"

  # Cannot use cyclic dependencies on controllers or their DNS records
  api_servers                     = ["${format("%s-private.%s", var.cluster_name, var.dns_zone)}"]
  api_servers_external            = ["${format("%s.%s", var.cluster_name, var.dns_zone)}"]
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
