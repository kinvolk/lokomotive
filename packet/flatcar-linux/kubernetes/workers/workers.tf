resource "packet_device" "nodes" {
  count            = "${var.count}"
  hostname         = "${var.cluster_name}-${var.pool_name}-worker-${count.index}"
  plan             = "${var.type}"
  facilities       = ["${var.facility}"]
  operating_system = "custom_ipxe"
  billing_cycle    = "hourly"
  project_id       = "${var.project_id}"
  ipxe_script_url  = "${var.ipxe_script_url}"
  always_pxe       = "false"
  user_data        = "${data.ct_config.install-ignitions.rendered}"

  # If not present in the map, it uses "" that means no reservation id
  hardware_reservation_id = "${lookup(var.reservation_ids, format("worker-%v", count.index), "")}"
}

# These configs are used for the fist boot, to run flatcar-install
data "ct_config" "install-ignitions" {
  content = "${data.template_file.install.rendered}"
}

data "template_file" "install" {
  template = "${file("${path.module}/cl/install.yaml.tmpl")}"

  vars {
    os_channel           = "${var.os_channel}"
    os_version           = "${var.os_version}"
    flatcar_linux_oem    = "packet"
    ssh_keys             = "${jsonencode("${var.ssh_keys}")}"
    postinstall_ignition = "${data.ct_config.ignitions.rendered}"
    setup_raid           = "${var.setup_raid}"
  }
}

resource "packet_bgp_session" "bgp" {
  count          = "${var.count}"
  device_id      = "${element(packet_device.nodes.*.id, count.index)}"
  address_family = "ipv4"
}

data "ct_config" "ignitions" {
  content  = "${data.template_file.configs.rendered}"
  platform = "packet"
}

data "template_file" "configs" {
  template = "${file("${path.module}/cl/worker.yaml.tmpl")}"

  vars {
    kubeconfig            = "${indent(10, "${var.kubeconfig}")}"
    ssh_keys              = "${jsonencode("${var.ssh_keys}")}"
    k8s_dns_service_ip    = "${cidrhost(var.service_cidr, 10)}"
    cluster_domain_suffix = "${var.cluster_domain_suffix}"
    worker_labels         = "${var.labels}"
  }
}
