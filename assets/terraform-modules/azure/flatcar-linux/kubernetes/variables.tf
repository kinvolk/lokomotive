variable "cluster_name" {
  type        = string
  description = "Unique cluster name (prepended to dns_zone)"
}

# Azure

variable "region" {
  type        = string
  description = "Azure Region (e.g. centralus , see `az account list-locations --output table`)"
}

variable "dns_zone" {
  type        = string
  description = "DNS Zone (e.g. example.com)"
}

# variable "dns_zone_group" {
#   type        = string
#   description = "Resource group where the Azure DNS Zone resides (e.g. global)"
# }

# variable "custom_image_resource_group_name" {
#   type        = string
#   description = "The name of the Resource Group in which the Custom Image exists."
# }

# variable "custom_image_name" {
#   type        = string
#   description = "The name of the Custom Image to provision this Virtual Machine from."
# }

variable "tags" {
  type = map(any)
  default = {
    "ManagedBy" = "Lokomotive"
    "CreatedBy" = "Unspecified"
  }
  description = "Optional details to tag on AWS resources"
}

# instances

variable "controller_count" {
  type        = number
  default     = 1
  description = "Number of controllers (i.e. masters)"
}

variable "worker_count" {
  type        = number
  default     = 1
  description = "Number of workers"
}

variable "controller_type" {
  type        = string
  default     = "Standard_B2s"
  description = "Machine type for controllers (see `az vm list-skus --location centralus`)"
}

variable "worker_type" {
  type        = string
  default     = "Standard_DS1_v2"
  description = "Machine type for workers (see `az vm list-skus --location centralus`)"
}

variable "os_image" {
  type        = string
  description = "Channel for a Container Linux derivative (flatcar-stable, flatcar-beta, flatcar-alpha)"
  default     = "flatcar-stable"

  validation {
    condition     = contains(["flatcar-stable", "flatcar-beta", "flatcar-alpha"], var.os_image)
    error_message = "The os_image must be flatcar-stable, flatcar-beta, or flatcar-alpha."
  }
}

variable "disk_size" {
  type        = number
  default     = 30
  description = "Size of the disk in GB"
}

variable "controller_clc_snippets" {
  type        = list(string)
  description = "Controller Container Linux Config snippets"
  default     = []
}

variable "clc_snippets" {
  type        = list(string)
  description = "Worker Container Linux Config snippets"
  default     = []
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

variable "host_cidr" {
  description = "CIDR IPv4 range to assign to instances"
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

variable "encrypt_pod_traffic" {
  description = "Enable in-cluster pod traffic encryption."
  type        = bool
  default     = false
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

variable "conntrack_max_per_core" {
  description = "--conntrack-max-per-core value for kube-proxy. Maximum number of NAT connections to track per CPU core (0 to leave the limit as-is and ignore the conntrack-min kube-proxy flag)."
  type        = number
}
