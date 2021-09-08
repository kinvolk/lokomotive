resource "local_file" "bootstrap-apiserver" {
  filename = "${var.asset_dir}/bootstrap-manifests/bootstrap-apiserver.yaml"
  content = templatefile("${path.module}/resources/bootstrap-manifests/bootstrap-apiserver.yaml", {
    kube_apiserver_image = var.container_images["kube_apiserver"]
    cloud_provider       = var.cloud_provider
    etcd_servers         = join(",", formatlist("https://%s:2379", var.etcd_servers))
    service_cidr         = var.service_cidr
    trusted_certs_dir    = var.trusted_certs_dir
    enable_tls_bootstrap = var.enable_tls_bootstrap
  })
}

resource "local_file" "bootstrap-controller-manager" {
  filename = "${var.asset_dir}/bootstrap-manifests/bootstrap-controller-manager.yaml"
  content = templatefile("${path.module}/resources/bootstrap-manifests/bootstrap-controller-manager.yaml", {
    kube_controller_manager_image = var.container_images["kube_controller_manager"]
    pod_cidr                      = var.pod_cidr
    service_cidr                  = var.service_cidr
    cloud_provider                = var.cloud_provider
    trusted_certs_dir             = var.trusted_certs_dir
  })
}

resource "local_file" "bootstrap-scheduler" {
  filename = "${var.asset_dir}/bootstrap-manifests/bootstrap-scheduler.yaml"
  content = templatefile("${path.module}/resources/bootstrap-manifests/bootstrap-scheduler.yaml", {
    kube_scheduler_image = var.container_images["kube_scheduler"]
  })
}

resource "local_file" "kube-apiserver" {
  filename = "${var.asset_dir}/charts/kube-system/kube-apiserver.yaml"
  content = templatefile("${path.module}/resources/charts/kube-apiserver.yaml", {
    kube_apiserver_image    = var.container_images["kube_apiserver"]
    etcd_servers            = join(",", formatlist("https://%s:2379", var.etcd_servers))
    cloud_provider          = var.cloud_provider
    service_cidr            = var.service_cidr
    trusted_certs_dir       = var.trusted_certs_dir
    ca_cert                 = base64encode(tls_self_signed_cert.kube-ca.cert_pem)
    apiserver_key           = base64encode(tls_private_key.apiserver.private_key_pem)
    apiserver_cert          = base64encode(tls_locally_signed_cert.apiserver.cert_pem)
    serviceaccount_private  = base64encode(tls_private_key.service-account.private_key_pem)
    etcd_ca_cert            = base64encode(tls_self_signed_cert.etcd-ca.cert_pem)
    etcd_client_cert        = base64encode(tls_locally_signed_cert.client.cert_pem)
    etcd_client_key         = base64encode(tls_private_key.client.private_key_pem)
    enable_aggregation      = var.enable_aggregation
    aggregation_ca_cert     = var.enable_aggregation == true ? base64encode(join(" ", tls_self_signed_cert.aggregation-ca.*.cert_pem)) : ""
    aggregation_client_cert = var.enable_aggregation == true ? base64encode(join(" ", tls_locally_signed_cert.aggregation-client.*.cert_pem)) : ""
    aggregation_client_key  = var.enable_aggregation == true ? base64encode(join(" ", tls_private_key.aggregation-client.*.private_key_pem)) : ""
    replicas                = var.controller_count
    extra_flags             = var.kube_apiserver_extra_flags
    enable_tls_bootstrap    = var.enable_tls_bootstrap
    ignore_x509_cn_check    = var.ignore_x509_cn_check
  })
}

resource "local_file" "pod-checkpointer" {
  filename = "${var.asset_dir}/charts/kube-system/pod-checkpointer.yaml"
  content = templatefile("${path.module}/resources/charts/pod-checkpointer.yaml", {
    pod_checkpointer_image = var.container_images["pod_checkpointer"]
  })
}

# Populate kubernetes chart values file named kubernetes.yaml.
resource "local_file" "kubernetes" {
  filename = "${var.asset_dir}/charts/kube-system/kubernetes.yaml"
  content = templatefile("${path.module}/resources/charts/kubernetes.yaml", {
    kube_controller_manager_image = var.container_images["kube_controller_manager"]
    kube_scheduler_image          = var.container_images["kube_scheduler"]
    kube_proxy_image              = var.container_images["kube_proxy"]
    coredns_image                 = var.container_images["coredns"]
    control_plane_replicas        = var.controller_count
    cloud_provider                = var.cloud_provider
    pod_cidr                      = var.pod_cidr
    service_cidr                  = var.service_cidr
    cluster_domain_suffix         = var.cluster_domain_suffix
    cluster_dns_service_ip        = cidrhost(var.service_cidr, 10)
    trusted_certs_dir             = var.trusted_certs_dir
    ca_cert                       = base64encode(tls_self_signed_cert.kube-ca.cert_pem)
    ca_key                        = base64encode(tls_private_key.kube-ca.private_key_pem)
    server                        = format("https://%s:%s", var.api_servers[0], var.external_apiserver_port)
    serviceaccount_key            = base64encode(tls_private_key.service-account.private_key_pem)
    etcd_endpoints                = var.etcd_endpoints
    enable_tls_bootstrap          = var.enable_tls_bootstrap
    conntrack_max_per_core        = var.conntrack_max_per_core
  })
}

# Populate node-local-dns chart values file named node-local-dns.yaml.
resource "local_file" "node-local-dns" {
  count = var.enable_node_local_dns ? 1 : 0

  filename = "${var.asset_dir}/charts/kube-system/node-local-dns.yaml"
  content = templatefile("${path.module}/resources/charts/node-local-dns.yaml", {
    cluster_dns_service_ip = cidrhost(var.service_cidr, 10)
    cluster_domain_suffix  = var.cluster_domain_suffix
    node_local_dns_ip      = var.node_local_dns_ip
  })
}

locals {
  bootstrap_secrets = templatefile("${path.module}/resources/charts/bootstrap-secrets.yaml", {
    bootstrap_tokens = var.bootstrap_tokens
  })
}

# Populate bootstrap-secrets chart values file named bootstrap-secrets.yaml.
resource "local_file" "bootstrap-secrets" {
  count = var.enable_tls_bootstrap == true ? 1 : 0

  filename = "${var.asset_dir}/charts/kube-system/bootstrap-secrets.yaml"
  content  = local.bootstrap_secrets
}

locals {
  kubelet = var.disable_self_hosted_kubelet == false ? 1 : 0
  # Render kubelet.yaml for kubelet chart
  kubelet_content = templatefile("${path.module}/resources/charts/kubelet.yaml", {
    kubelet_image          = var.container_images["kubelet_image"]
    cluster_dns_service_ip = cidrhost(var.service_cidr, 10)
    cluster_domain_suffix  = var.cluster_domain_suffix
    enable_tls_bootstrap   = var.enable_tls_bootstrap
    cloud_provider         = var.cloud_provider
    kubernetes_ca_cert     = base64encode(tls_self_signed_cert.kube-ca.cert_pem)
  })

  kubeconfig_kubelet_content = templatefile("${path.module}/resources/kubeconfig-kubelet", {
    ca_cert      = base64encode(tls_self_signed_cert.kube-ca.cert_pem)
    kubelet_cert = base64encode(tls_locally_signed_cert.kubelet.cert_pem)
    kubelet_key  = base64encode(tls_private_key.kubelet.private_key_pem)
    server       = format("https://%s:%s", var.api_servers[0], var.external_apiserver_port)
  })

  kubeconfig_admin_content = templatefile("${path.module}/resources/kubeconfig-admin", {
    name         = var.cluster_name
    ca_cert      = base64encode(tls_self_signed_cert.kube-ca.cert_pem)
    kubelet_cert = base64encode(tls_locally_signed_cert.admin.cert_pem)
    kubelet_key  = base64encode(tls_private_key.admin.private_key_pem)
    server       = format("https://%s:%s", local.api_servers_external[0], var.external_apiserver_port)
  })
}

# Populate kubelet chart values file named kubelet.yaml.
resource "local_file" "kubelet" {
  count = local.kubelet

  content  = join("", [for i in range(0, 1) : local.kubelet_content])
  filename = "${var.asset_dir}/charts/kube-system/kubelet.yaml"
}

# Generated kubeconfig for Kubelets
resource "local_file" "kubeconfig-kubelet" {
  content  = local.kubeconfig_kubelet_content
  filename = "${var.asset_dir}/auth/kubeconfig-kubelet"
}

# Generated admin kubeconfig (bootkube requires it be at auth/kubeconfig)
# https://github.com/kubernetes-incubator/bootkube/blob/master/pkg/bootkube/bootkube.go#L42
resource "local_file" "kubeconfig-admin" {
  content  = local.kubeconfig_admin_content
  filename = "${var.asset_dir}/auth/kubeconfig"
}

# Generated admin kubeconfig in a file named after the cluster
resource "local_file" "kubeconfig-admin-named" {
  content  = local.kubeconfig_admin_content
  filename = "${var.asset_dir}/auth/${var.cluster_name}-config"
}

# If var.api_servers_external isn't set, use var.api_servers.
# This is for supporting separate API server URLs for external clients in a backward-compatible way.
# The use of split() and join() here is because Terraform's conditional operator ('?') cannot be
# used with lists.
locals {
  api_servers_external = split(",", join(",", var.api_servers_external) == "" ? join(",", var.api_servers) : join(",", var.api_servers_external))
}
