locals {
  api_server = format("%s.%s", var.cluster_name, var.dns_zone)
}

# Self-hosted Kubernetes assets (kubeconfig, manifests)
module "bootkube" {
  source = "../../../bootkube"

  cluster_name                = var.cluster_name
  api_servers                 = [local.api_server]
  etcd_servers                = aws_route53_record.etcds.*.fqdn
  etcd_endpoints              = aws_instance.controllers.*.private_ip
  asset_dir                   = var.asset_dir
  network_mtu                 = var.network_mtu
  pod_cidr                    = var.pod_cidr
  service_cidr                = var.service_cidr
  cluster_domain_suffix       = var.cluster_domain_suffix
  enable_reporting            = var.enable_reporting
  enable_aggregation          = var.enable_aggregation
  kube_apiserver_extra_flags  = var.kube_apiserver_extra_flags
  certs_validity_period_hours = var.certs_validity_period_hours

  # Disable the self hosted kubelet.
  disable_self_hosted_kubelet = var.disable_self_hosted_kubelet

  # Block access to AWS metadata for all pods.
  #
  # https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ec2-instance-metadata.html
  blocked_metadata_cidrs = ["169.254.169.254/32"]

  bootstrap_tokens     = var.enable_tls_bootstrap ? concat([local.controller_bootstrap_token], var.worker_bootstrap_tokens) : []
  enable_tls_bootstrap = var.enable_tls_bootstrap
  encrypt_pod_traffic  = var.encrypt_pod_traffic
}
