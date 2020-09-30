variable "cluster_name" {
  type        = string
  description = "Unique cluster name (prepended to dns_zone)"
}

# AWS

variable "dns_zone" {
  type        = string
  description = "AWS Route53 DNS Zone (e.g. aws.example.com)"
}

variable "dns_zone_id" {
  type        = string
  description = "AWS Route53 DNS Zone ID (e.g. Z3PAABBCFAKEC0)"
}

variable "expose_nodeports" {
  type        = bool
  default     = false
  description = "Expose node ports 30000-32767 in the security group"
}

# instances

variable "controller_count" {
  type        = number
  default     = 1
  description = "Number of controllers (i.e. masters)"
}

variable "controller_type" {
  type = string
  # When doing the upgrades of controlplane on t3.small instance type when
  # having one single controlplane node, t3.small has not enough memory (2GB)
  # to run more than one instance of kube-apiserver in parallel, so we need to use
  # a bigger instance. With HA controlplane, t3.small should be fine, though for
  # production setups, it's recommended to use instances with more RAM, to
  # give plenty of usable memory for etcd and kube-apiserver.
  default     = "t3.medium"
  description = "EC2 instance type for controllers"
}

variable "os_name" {
  type        = string
  default     = "flatcar"
  description = "Name of Operating System (coreos or flatcar)"
}

variable "os_channel" {
  type        = string
  default     = "stable"
  description = "AMI channel for the OS (stable, beta, alpha, edge)"
}

variable "os_version" {
  type        = string
  default     = "current"
  description = "Version of the OS (current or numeric version such as 2261.99.0)"
}

variable "disk_size" {
  type        = number
  default     = 40
  description = "Size of the EBS volume in GB"
}

variable "disk_type" {
  type        = string
  default     = "gp2"
  description = "Type of the EBS volume (e.g. standard, gp2, io1)"
}

variable "disk_iops" {
  type        = number
  default     = 0
  description = "IOPS of the EBS volume (e.g. 100)"
}

variable "controller_clc_snippets" {
  type        = list(string)
  description = "Controller Container Linux Config snippets"
  default     = []
}

variable "tags" {
  type = map
  default = {
    "ManagedBy" = "Lokomotive"
    "CreatedBy" = "Unspecified"
  }
  description = "Optional details to tag on AWS resources"
}

variable "enable_csi" {
  type        = bool
  default     = false
  description = "Set up IAM role needed for dynamic volumes provisioning to work on AWS"
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
