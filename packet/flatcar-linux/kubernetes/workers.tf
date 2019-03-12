resource "packet_device" "worker_nodes" {
  count            = "${var.worker_count}"
  hostname         = "${var.cluster_name}-worker-${count.index}"
  plan             = "${var.worker_type}"
  facility         = "${var.facility}"
  operating_system = "custom_ipxe"
  billing_cycle    = "hourly"
  project_id       = "${var.project_id}"
  ipxe_script_url  = "${var.ipxe_script_url}"
  always_pxe       = "false"
  user_data        = "${element(data.ct_config.worker-ignitions.rendered, count.index)}"
}

data "ct_config" "worker-ignitions" {
  content = "${data.template_file.worker-configs.rendered}"
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
