# Required variables.
variable "dns_zone" {
  type        = string
  description = "Domain name for the the cluster. E.g. 'example.com'"
}

variable "cluster_name" {
  type        = string
  description = "Cluster name."
}

variable "count_index" {
  type        = number
  description = "Index passed as count.index from count on module."
}

variable "controllers_count" {
  type        = number
  description = "Number of controller nodes in the cluster."
}

variable "apiserver" {
  type        = string
  description = "FQDN or IP address for kubelet to use for talking to Kubernetes API server."
}

variable "ca_cert" {
  type        = string
  description = "Kubernetes CA certificate in PEM format for bootstrap kubeconfig for kubelet."
}

# Optional variables.
variable "bootkube_image_name" {
  type        = string
  description = "Docker image name to use for container running bootkube."
  default     = "quay.io/kinvolk/bootkube"
}

variable "bootkube_image_tag" {
  type        = string
  description = "Docker image tag to use for container running bootkube."
  default     = "v0.14.0-helm4"
}

variable "cluster_domain_suffix" {
  type        = string
  description = "Cluster domain suffix. Passed to kubelet as --cluster_domain flag."
  default     = "cluster.local"
}

variable "ssh_keys" {
  type        = list(string)
  description = "List of SSH public keys for user `core`. Each element must be specified in a valid OpenSSH public key format, as defined in RFC 4253 Section 6.6, e.g. 'ssh-rsa AAAAB3N...'."
  default     = []
}

variable "cluster_dns_service_ip" {
  type        = string
  description = "IP address of cluster DNS Service. Passed to kubelet as --cluster_dns parameter."
  default     = "10.3.0.10"
}

variable "clc_snippets" {
  type        = list(string)
  description = "Extra CLC snippets to include in the configuration."
  default     = []
}

variable "kubelet_image_name" {
  type        = string
  description = "Source of kubelet Docker image."
  default     = "quay.io/kinvolk/kubelet"
}

variable "kubelet_image_tag" {
  type        = string
  description = "Tag for kubelet Docker image."
  default     = "v1.20.4"
}

variable "host_dns_ip" {
  type        = string
  description = "IP address of DNS server to configure on the nodes."
  default     = "8.8.8.8"
}
