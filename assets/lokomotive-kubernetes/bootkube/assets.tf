# Self-hosted Kubernetes bootstrap-manifests
resource "template_dir" "bootstrap-manifests" {
  source_dir      = "${replace(path.module, path.cwd, ".")}/resources/bootstrap-manifests"
  destination_dir = "${var.asset_dir}/bootstrap-manifests"

  vars = {
    hyperkube_image   = var.container_images["hyperkube"]
    etcd_servers      = join(",", formatlist("https://%s:2379", var.etcd_servers))
    cloud_provider    = var.cloud_provider
    pod_cidr          = var.pod_cidr
    service_cidr      = var.service_cidr
    trusted_certs_dir = var.trusted_certs_dir
  }
}

resource "local_file" "kube-apiserver" {
  filename = "${var.asset_dir}/charts/kube-system/kube-apiserver.yaml"
  content  = templatefile("${path.module}/resources/charts/kube-apiserver.yaml", {
    hyperkube_image         = var.container_images["hyperkube"]
    pod_checkpointer_image  = var.container_images["pod_checkpointer"]
    etcd_servers            = join(",", formatlist("https://%s:2379", var.etcd_servers))
    cloud_provider          = var.cloud_provider
    service_cidr            = var.service_cidr
    trusted_certs_dir       = var.trusted_certs_dir
    ca_cert                 = base64encode(tls_self_signed_cert.kube-ca.cert_pem)
    apiserver_key           = base64encode(tls_private_key.apiserver.private_key_pem)
    apiserver_cert          = base64encode(tls_locally_signed_cert.apiserver.cert_pem)
    serviceaccount_pub      = base64encode(tls_private_key.service-account.public_key_pem)
    etcd_ca_cert            = base64encode(tls_self_signed_cert.etcd-ca.cert_pem)
    etcd_client_cert        = base64encode(tls_locally_signed_cert.client.cert_pem)
    etcd_client_key         = base64encode(tls_private_key.client.private_key_pem)
    enable_aggregation      = var.enable_aggregation
    aggregation_ca_cert     = var.enable_aggregation == true ? base64encode(join(" ", tls_self_signed_cert.aggregation-ca.*.cert_pem)) : ""
    aggregation_client_cert = var.enable_aggregation == true ? base64encode(join(" ", tls_locally_signed_cert.aggregation-client.*.cert_pem)) : ""
    aggregation_client_key  = var.enable_aggregation == true ? base64encode(join(" ", tls_private_key.aggregation-client.*.private_key_pem)) : ""
  })
}

resource "template_dir" "kube-apiserver" {
  source_dir      = "${replace(path.module, path.cwd, ".")}/resources/charts/kube-apiserver"
  destination_dir = "${var.asset_dir}/charts/kube-system/kube-apiserver"
}

resource "local_file" "podcheckpointer" {
  filename = "${var.asset_dir}/charts/kube-system/podcheckpointer.yaml"
  content  = templatefile("${path.module}/resources/charts/podcheckpointer.yaml", {
    pod_checkpointer_image  = var.container_images["pod_checkpointer"]
  })
}

resource "template_dir" "podcheckpointer" {
  source_dir      = "${replace(path.module, path.cwd, ".")}/resources/charts/podcheckpointer"
  destination_dir = "${var.asset_dir}/charts/kube-system/podcheckpointer"
}

# Populate kubernetes chart values file named kubernetes.yaml.
resource "local_file" "kubernetes" {
  content  = data.template_file.kubernetes.rendered
  filename = "${var.asset_dir}/charts/kube-system/kubernetes.yaml"
}

# Populate kubernetes control plane chart.
# TODO: Currently, there is no way in Terraform to copy local directory, so we use `template_dir` for it.
# The downside is, that any Terraform templating syntax stored in this directory will be evaluated, which may bring unexpected results.
resource "template_dir" "kubernetes" {
  source_dir      = "${replace(path.module, path.cwd, ".")}/resources/charts/kubernetes"
  destination_dir = "${var.asset_dir}/charts/kube-system/kubernetes"
}

# Render kubernetes.yaml for kubernetes chart.
data "template_file" "kubernetes" {
  template = "${file("${path.module}/resources/charts/kubernetes.yaml")}"

  vars = {
    hyperkube_image         = var.container_images["hyperkube"]
    coredns_image           = "${var.container_images["coredns"]}${var.container_arch}"
    control_plane_replicas  = max(2, length(var.etcd_servers))
    cloud_provider          = var.cloud_provider
    pod_cidr                = var.pod_cidr
    service_cidr            = var.service_cidr
    cluster_domain_suffix   = var.cluster_domain_suffix
    cluster_dns_service_ip  = cidrhost(var.service_cidr, 10)
    trusted_certs_dir       = var.trusted_certs_dir
    ca_cert                 = base64encode(tls_self_signed_cert.kube-ca.cert_pem)
    ca_key                  = base64encode(tls_private_key.kube-ca.private_key_pem)
    server                  = format("https://%s:%s", var.api_servers[0], var.external_apiserver_port)
    serviceaccount_key      = base64encode(tls_private_key.service-account.private_key_pem)
  }
}

# Render kubelet.yaml for kubelet chart
data "template_file" "kubelet" {
  template = "${file("${path.module}/resources/charts/kubelet.yaml")}"

  vars = {
    hyperkube_image         = var.container_images["hyperkube"]
    cluster_dns_service_ip  = cidrhost(var.service_cidr, 10)
    cluster_domain_suffix   = var.cluster_domain_suffix
  }
}

# Populate kubelet chart values file named kubelet.yaml.
resource "local_file" "kubelet" {
  content = data.template_file.kubelet.rendered
  filename = "${var.asset_dir}/charts/kube-system/kubelet.yaml"
}

# Populate kubelet chart.
# TODO: Currently, there is no way in Terraform to copy local directory, so we use `template_dir` for it.
# The downside is, that any Terraform templating syntax stored in this directory will be evaluated, which may bring unexpected results.
resource "template_dir" "kubelet" {
  source_dir = "${replace(path.module, path.cwd, ".")}/resources/charts/kubelet"
  destination_dir = "${var.asset_dir}/charts/kube-system/kubelet"
}

# Generated kubeconfig for Kubelets
resource "local_file" "kubeconfig-kubelet" {
  content  = data.template_file.kubeconfig-kubelet.rendered
  filename = "${var.asset_dir}/auth/kubeconfig-kubelet"
}

# Generated admin kubeconfig (bootkube requires it be at auth/kubeconfig)
# https://github.com/kubernetes-incubator/bootkube/blob/master/pkg/bootkube/bootkube.go#L42
resource "local_file" "kubeconfig-admin" {
  content  = data.template_file.kubeconfig-admin.rendered
  filename = "${var.asset_dir}/auth/kubeconfig"
}

# Generated admin kubeconfig in a file named after the cluster
resource "local_file" "kubeconfig-admin-named" {
  content  = data.template_file.kubeconfig-admin.rendered
  filename = "${var.asset_dir}/auth/${var.cluster_name}-config"
}

data "template_file" "kubeconfig-kubelet" {
  template = file("${path.module}/resources/kubeconfig-kubelet")

  vars = {
    ca_cert      = base64encode(tls_self_signed_cert.kube-ca.cert_pem)
    kubelet_cert = base64encode(tls_locally_signed_cert.kubelet.cert_pem)
    kubelet_key  = base64encode(tls_private_key.kubelet.private_key_pem)
    server       = format("https://%s:%s", var.api_servers[0], var.external_apiserver_port)
  }
}

# If var.api_servers_external isn't set, use var.api_servers.
# This is for supporting separate API server URLs for external clients in a backward-compatible way.
# The use of split() and join() here is because Terraform's conditional operator ('?') cannot be
# used with lists.
locals {
  api_servers_external = split(",", join(",", var.api_servers_external) == "" ? join(",", var.api_servers) : join(",", var.api_servers_external))
}

data "template_file" "kubeconfig-admin" {
  template = file("${path.module}/resources/kubeconfig-admin")

  vars = {
    name         = var.cluster_name
    ca_cert      = base64encode(tls_self_signed_cert.kube-ca.cert_pem)
    kubelet_cert = base64encode(tls_locally_signed_cert.admin.cert_pem)
    kubelet_key  = base64encode(tls_private_key.admin.private_key_pem)
    server       = format("https://%s:%s", local.api_servers_external[0], var.external_apiserver_port)
  }
}
