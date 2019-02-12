resource "packet_device" "worker_nodes" {
  count            = "${var.worker_count}"
  hostname         = "${var.cluster_name}-worker-${count.index}"
  plan             = "${var.worker_type}"
  facilities       = ["${var.facility}"]
  operating_system = "custom_ipxe"
  billing_cycle    = "hourly"
  project_id       = "${var.project_id}"
  ipxe_script_url  = "${var.ipxe_script_url}"
  always_pxe       = "false"
  user_data        = "${data.ct_config.worker-install-ignitions.rendered}"
}

# These configs are used for the fist boot, to run flatcar-install
data "ct_config" "worker-install-ignitions" {
  content = "${data.template_file.worker-install.rendered}"
}

data "template_file" "worker-install" {
  template = "${file("${path.module}/cl/install.yaml.tmpl")}"

  vars {
    os_channel           = "${var.os_channel}"
    install_disk         = "${var.install_disk}"
    flatcar_linux_oem    = "packet"
    ssh_keys             = "${jsonencode("${var.ssh_keys}")}"
    postinstall_ignition = "${data.ct_config.worker-ignitions.rendered}"
  }
}

resource "packet_bgp_session" "bgp" {
  count          = "${var.worker_count}"
  device_id      = "${element(packet_device.worker_nodes.*.id, count.index)}"
  address_family = "ipv4"
}

data "ct_config" "worker-ignitions" {
  content = "${data.template_file.worker-configs.rendered}"
  platform = "packet"
}

data "template_file" "worker-configs" {
  template = "${file("${path.module}/cl/worker.yaml.tmpl")}"

  vars {
    kubeconfig            = "${indent(10, module.bootkube.kubeconfig-kubelet)}"
    ssh_keys              = "${jsonencode("${var.ssh_keys}")}"
    k8s_dns_service_ip    = "${cidrhost(var.service_cidr, 10)}"
    cluster_domain_suffix = "${var.cluster_domain_suffix}"
  }
}
