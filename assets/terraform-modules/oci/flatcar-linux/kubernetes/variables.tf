variable "cluster_name" {
  type        = string
  description = "Unique cluster name (prepended to dns_zone)"
}

variable "dns_zone" {
  type        = string
  description = "AWS Route53 DNS Zone (e.g. aws.example.com)"
}

variable "dns_zone_id" {
  type        = string
  description = "AWS Route53 DNS Zone ID (e.g. Z3PAABBCFAKEC0)"
}

# OCI
variable "compartment_id" {
  type = string
}

variable "tenancy_id" {
  type = string
}

variable "controller_instance_shape" {
  type = string
}

variable "region" {
  type = string
}

variable "user" {
  type = string
}

variable "fingerprint" {
  type = string
}

variable "key" {
  type = string
}

# instances

variable "controller_count" {
  type        = number
  default     = 1
  description = "Number of controllers (i.e. masters)"
}

variable "controller_image_id" {
  type = string
}

variable "os_channel" {
  type        = string
  default     = "stable"
  description = "Channel for the OS (stable, beta, alpha, edge)"
}

variable "os_arch" {
  type        = string
  default     = "amd64"
  description = "Arch for the OS (amd64, arm64)"
}

variable "os_version" {
  type        = string
  default     = "current"
  description = "Version of the OS (current or numeric version such as 2261.99.0)"
}

variable "disk_size" {
  type        = number
  default     = 50
  description = "Size of the boot volume in GB"
}

variable "controller_cpus" {
  default = 2
}

variable "controller_memory" {
  default = 6
}

variable "controller_ad_number" {
  default = 1
}

variable "controller_clc_snippets" {
  type        = list(string)
  description = "Controller Container Linux Config snippets"
  default     = []
}

variable "tags" {
  type = map(any)
  default = {
    "ManagedBy" = "Lokomotive"
    "CreatedBy" = "Unspecified"
  }
  description = "Optional details to tag on AWS resources"
}

# configuration

variable "ssh_keys" {
  type        = list(string)
  description = "SSH public keys for user 'core'"
}

variable "asset_dir" {
  description = "Path to a directory where generated assets should be placed (contains secrets)"
  type        = string
}

variable "network_mtu" {
  description = "Physical Network MTU. Use 9001 if using instances types with Jumbo frames."
  type        = number
}

variable "host_cidr" {
  description = "CIDR IPv4 range to assign to EC2 nodes"
  type        = string
  default     = "10.0.0.0/16"
}

variable "pod_cidr" {
  description = "CIDR IPv4 range to assign Kubernetes pods"
  type        = string
  default     = "10.2.0.0/16"
}

variable "service_cidr" {
  description = <<EOD
CIDR IPv4 range to assign Kubernetes services.
The 1st IP will be reserved for kube_apiserver, the 10th IP will be reserved for coredns.
EOD


  type    = string
  default = "10.3.0.0/16"
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

variable "disable_self_hosted_kubelet" {
  description = "Disable the self hosted kubelet installed by default"
  type        = bool
}

variable "enable_tls_bootstrap" {
  description = "Enable TLS Bootstrap for Kubelet."
  type        = bool
}

variable "worker_bootstrap_tokens" {
  description = "List of token-id and token-secret of each node."
  type        = list(any)
}

# Certificates

variable "certs_validity_period_hours" {
  description = "Validity of all the certificates in hours"
  type        = number
  default     = 8760
}

variable "kube_apiserver_extra_flags" {
  description = "Extra flags passed to self-hosted kube-apiserver."
  type        = list(string)
  default     = []
}

variable "encrypt_pod_traffic" {
  description = "Enable in-cluster pod traffic encryption."
  type        = bool
  default     = false
}

variable "ignore_x509_cn_check" {
  description = "Ignore CN checks in x509 certificates."
  type        = bool
  default     = false
}

variable "conntrack_max_per_core" {
  description = "--conntrack-max-per-core value for kube-proxy. Maximum number of NAT connections to track per CPU core (0 to leave the limit as-is and ignore the conntrack-min kube-proxy flag)."
  type        = number
}
