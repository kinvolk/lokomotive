module "workers" {
  source = "./workers"
  name   = "${var.cluster_name}"

  # Azure
  resource_group_name     = "${azurerm_resource_group.cluster.name}"
  region                  = "${azurerm_resource_group.cluster.location}"
  subnet_id               = "${azurerm_subnet.worker.id}"
  security_group_id       = "${azurerm_network_security_group.worker.id}"
  backend_address_pool_id = "${azurerm_lb_backend_address_pool.worker.id}"

  worker_count = "${var.worker_count}"
  vm_type      = "${var.worker_type}"
  os_image     = "${var.os_image}"
  priority     = "${var.worker_priority}"

  custom_image_resource_group_name = "${var.custom_image_resource_group_name}"
  custom_image_name = "${var.custom_image_name}"

  # configuration
  kubeconfig            = "${module.bootkube.kubeconfig-kubelet}"
  ssh_keys              = "${var.ssh_keys}"
  service_cidr          = "${var.service_cidr}"
  cluster_domain_suffix = "${var.cluster_domain_suffix}"
  clc_snippets          = "${var.worker_clc_snippets}"
}
