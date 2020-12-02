# Required variables.
variable "dns_zone" {
  type        = string
  description = "Domain name for the the cluster. E.g. 'example.com'"
}

variable "cluster_name" {
  type        = string
  description = "Cluster identifier which will be used in controller node names."
}

variable "ip_addresses" {
  type        = list(string)
  description = "List of IP addresses of Tinkerbell workers where controller nodes should be provisioned."
}

variable "asset_dir" {
  type        = string
  description = "Path to a directory where generated assets should be placed (contains secrets)."
}

variable "worker_bootstrap_tokens" {
  type = list(object({
    token_id     = string
    token_secret = string
  }))
  description = "List of token-id and token-secret of each node."
}

variable "conntrack_max_per_core" {
  description = "--conntrack-max-per-core value for kube-proxy. Maximum number of NAT connections to track per CPU core (0 to leave the limit as-is and ignore the conntrack-min kube-proxy flag)."
  type        = number
}

# Optional variables.
variable "flatcar_install_base_url" {
  type        = string
  description = "URL passed to the `flatcar-install` script to fetch Flatcar images from."
  default     = ""
}

variable "os_version" {
  type        = string
  description = "Flatcar version to install."
  default     = ""
}

variable "os_channel" {
  type        = string
  description = "Flatcar channel to use for installation."
  default     = ""
}

variable "ssh_keys" {
  type        = list(string)
  description = "List of SSH public keys for user `core`. Each element must be specified in a valid OpenSSH public key format, as defined in RFC 4253 Section 6.6, e.g. 'ssh-rsa AAAAB3N...'."
  default     = []
}

variable "node_count" {
  type        = number
  description = "Number of nodes to create."
  default     = 1
}

variable "clc_snippets" {
  type        = list(string)
  description = "Extra CLC snippets to include in the configuration."
  default     = []
}

variable "cluster_domain_suffix" {
  type        = string
  description = "Cluster domain suffix. Passed to kubelet as --cluster_domain flag."
  default     = "cluster.local"
}

variable "network_mtu" {
  type        = number
  description = "CNI interface MTU."
  default     = 1500
}

variable "pod_cidr" {
  type        = string
  description = "CIDR IP range to assign Kubernetes pods."
  default     = "10.2.0.0/16"
}

variable "service_cidr" {
  type        = string
  description = <<EOF
CIDR IP range to assign Kubernetes services.
The 1st IP will be reserved for kube_apiserver, the 10th IP will be reserved for kube-dns.
EOF
  default     = "10.3.0.0/24"
}

variable "enable_reporting" {
  type        = bool
  description = "Enable usage or analytics reporting to upstream component owners (Tigera: Calico)."
  default     = false
}

variable "certs_validity_period_hours" {
  type        = number
  description = "Validity of all the certificates in hours."
  default     = 8760
}

variable "enable_aggregation" {
  type        = bool
  description = "Enable the Kubernetes Aggregation Layer (defaults to false, recommended)."
  default     = true
}

variable "host_dns_ip" {
  type        = string
  description = "IP address of DNS server to configure on the nodes."
  default     = "8.8.8.8"
}
