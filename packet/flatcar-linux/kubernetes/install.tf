provider "packet" {}

provider "aws" {
  region = "${var.aws_region}"
}

# Discrete DNS records for each controller's private IPv4 for etcd usage
resource "aws_route53_record" "etcds" {
  count = "${var.controller_count}"

  # DNS Zone where record should be created
  zone_id = "${var.dns_zone_id}"

  name = "${format("%s-etcd%d.%s.", var.cluster_name, count.index, var.dns_zone)}"
  type = "A"
  ttl  = 300

  # private IPv4 address for etcd
  records = ["${element(packet_device.controllers.*.access_private_ipv4, count.index)}"]
}

# DNS record for the API servers
resource "aws_route53_record" "apiservers" {
  zone_id = "${var.dns_zone_id}"

  name = "${format("%s.%s.", var.cluster_name, var.dns_zone)}"
  type = "A"
  ttl  = "300"

  # TODO - figure out a way to access the API servers when they don't have a public IPv4
  # TODO - verify that a multi-controller setup actually works
  records = ["${packet_device.controllers.*.access_public_ipv4}"]
}

resource "packet_device" "controllers" {
  count            = "${var.controller_count}"
  hostname         = "controller-${count.index}"
  plan             = "${var.controller_type}"
  facility         = "${var.cluster_region}"
  operating_system = "custom_ipxe"
  billing_cycle    = "hourly"
  project_id       = "${var.project_id}"
  ipxe_script_url  = "${var.ipxe_script_url}"
  always_pxe       = "false"
  user_data        = "${element(data.ct_config.controller-ignitions.*.rendered, count.index)}"
}

data "ct_config" "controller-ignitions" {
  count   = "${var.controller_count}"
  content = "${element(data.template_file.controller-configs.*.rendered, count.index)}"
}

data "template_file" "controller-configs" {
  count    = "${var.controller_count}"
  template = "${file("${path.module}/cl/controller.yaml.tmpl")}"

  vars = {
    ssh_keys = "${jsonencode("${var.ssh_keys}")}"
  }
}

resource "packet_device" "worker_nodes" {
  count            = "${var.worker_count}"
  hostname         = "worker-${count.index}"
  plan             = "${var.worker_type}"
  facility         = "${var.cluster_region}"
  operating_system = "custom_ipxe"
  billing_cycle    = "hourly"
  project_id       = "${var.project_id}"
  ipxe_script_url  = "${var.ipxe_script_url}"
  always_pxe       = "false"
  user_data        = "${element(data.ct_config.controller-ignitions.*.rendered, count.index)}"
}

data "ct_config" "worker-ignitions" {
  count   = "${var.worker_count}"
  content = "${element(data.template_file.worker-configs.*.rendered, count.index)}"
}

data "template_file" "worker-configs" {
  count    = "${var.worker_count}"
  template = "${file("${path.module}/cl/worker.yaml.tmpl")}"

  vars = {
    ssh_keys = "${jsonencode("${var.ssh_keys}")}"
  }
}
