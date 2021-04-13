variable "cluster_name" {
  type        = string
  description = "Cluster name (prepended to pool name)"
}

variable "pool_name" {
  type        = string
  description = "Unique name for the worker pool"
}


# OCI
variable "tenancy_id" {
  type = string
}

variable "worker_image_id" {
  type = string
}

variable "compartment_id" {
  type = string
}

variable "worker_instance_shape" {
  type = string
}

variable "subnet_id" {
  type        = string
  description = "Must be set to `subnet_id` output by cluster"
}

variable "nsg_id" {
  type        = string
  description = "Must be set to `nsg_id` output by cluster"
}


# instances
variable "worker_count" {
  type        = number
  default     = 1
  description = "Number of instances"
}

variable "labels" {
  type        = map(string)
  description = "Map of custom labels for worker nodes."
  default     = {}
}

variable "taints" {
  type        = map(string)
  default     = {}
  description = "Map of custom taints for worker nodes."
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
  default     = 50
  description = "Size of the volume in GB"
}

variable "target_groups" {
  type        = list(string)
  description = "Additional target group ARNs to which instances should be added"
  default     = []
}

variable "clc_snippets" {
  type        = list(string)
  description = "Container Linux Config snippets"
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

variable "kubeconfig" {
  type        = string
  description = "Must be set to `kubeconfig` output by cluster"
}

variable "ca_cert" {
  description = "Kubernetes CA certificate needed in the kubeconfig file."
  type        = string
}

variable "apiserver" {
  description = "Apiserver private endpoint needed in the kubeconfig file."
  type        = string
}

variable "ssh_keys" {
  type        = list(string)
  description = "SSH public keys for user 'core'"
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

variable "enable_tls_bootstrap" {
  description = "Enable TLS Bootstrap for Kubelet."
  type        = bool
}

variable "dns_zone" {
  type        = string
  description = "AWS Route53 DNS Zone (e.g. aws.example.com)"
}

variable "worker_cpus" {
  default = 2
}

variable "worker_memory" {
  default = 6
}
