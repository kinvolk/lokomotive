module "controller" {
  source                 = "../../../controller"
  count                  = length(var.controller_names)
  cluster_name           = var.cluster_name
  controllers_count      = length(var.controller_names)
  dns_zone               = var.k8s_domain_name
  count_index            = count.index
  cluster_dns_service_ip = module.bootkube.cluster_dns_service_ip
  ssh_keys               = var.ssh_keys
  apiserver              = format("%s.%s", var.cluster_name, var.k8s_domain_name)
  ca_cert                = module.bootkube.ca_cert
  kubelet_labels         = lookup(var.node_specific_labels, var.controller_names[count.index], {})
  clc_snippets           = lookup(var.clc_snippets, var.controller_names[count.index], [])
  set_standard_hostname  = true
}
