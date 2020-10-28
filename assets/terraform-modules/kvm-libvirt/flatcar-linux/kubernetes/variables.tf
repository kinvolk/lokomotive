# DNS records

variable "cluster_name" {
  type        = string
  description = "Unique cluster name"
}

# Nodes
variable "os_image" {
  type        = string
  description = "Path to unpacked Flatcar Container Linux image flatcar_production_qemu_image.img (probably after a qemu-img resize IMG +5G)"
}

variable "controller_count" {
  type        = number
  default     = 1
  description = "Number of controllers (i.e. masters)"
}

variable "machine_domain" {
  type        = string
  description = "Machine domain"
}

variable "node_ip_pool" {
  type        = string
  default     = "192.168.192.0/24"
  description = "Unique VM IP CIDR"
}

variable "virtual_cpus" {
  type        = number
  default     = 1
  description = "Number of virtual CPUs"
}

variable "virtual_memory" {
  type        = number
  default     = 2048
  description = "Virtual RAM in MB"
}

variable "controller_clc_snippets" {
  type        = list(string)
  description = "Controller Container Linux Config snippets"
  default     = []
}

# Configuration

variable "ssh_keys" {
  type        = list(string)
  description = "SSH public keys for user 'core'"
}

variable "asset_dir" {
  description = "Path to a directory where generated assets should be placed (contains secrets)"
  type        = string
}

variable "network_mtu" {
  description = "CNI interface MTU"
  type        = number
  default     = 1480
}

variable "network_ip_autodetection_method" {
  description = "Method to autodetect the host IPv4 address"
  type        = string
  default     = "first-found"
}

variable "pod_cidr" {
  description = "CIDR IPv4 range to assign Kubernetes pods"
  type        = string
  default     = "10.1.0.0/16"
}

variable "service_cidr" {
  description = <<EOD
CIDR IPv4 range to assign Kubernetes services.
The 1st IP will be reserved for kube_apiserver, the 10th IP will be reserved for coredns.
EOD


  type    = string
  default = "10.2.0.0/16"
}

variable "cluster_domain_suffix" {
  description = "Queries for domains with the suffix will be answered by coredns. Default is cluster.local (e.g. foo.default.svc.cluster.local) "
  type        = string
  default     = "cluster.local"
}

variable "enable_reporting" {
  type        = bool
  description = "Enable usage or analytics reporting to upstreams (Calico)"
  default     = false
}

variable "enable_aggregation" {
  description = "Enable the Kubernetes Aggregation Layer (defaults to true)"
  type        = bool
  default     = true
}

variable "kube_apiserver_extra_flags" {
  description = "Extra flags passed to self-hosted kube-apiserver."
  type        = list(string)
  default     = []
}

variable "disable_self_hosted_kubelet" {
  description = "Disable the self hosted kubelet installed by default"
  type        = bool
}

# Certificates

variable "certs_validity_period_hours" {
  description = "Validity of all the certificates in hours"
  type        = number
  default     = 8760
}

variable "encrypt_pod_traffic" {
  description = "Enable in-cluster pod traffic encryption."
  type        = bool
  default     = false
}

variable "worker_bootstrap_tokens" {
  description = "List of token-id and token-secret of each node."
  type        = list(any)
}

variable "enable_tls_bootstrap" {
  description = "Enable TLS Bootstrap for Kubelet."
  type        = bool
}
