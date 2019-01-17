provider "packet" {}

resource "packet_device" "controller_nodes" {
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
