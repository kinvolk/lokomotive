variable "cluster_name" {
  type        = string
  description = "Unique cluster name (prepended to dns_zone)"
}

variable "pool_name" {
  type        = string
  description = "Unique name for the worker pool"
}

# Azure

variable "region" {
  type        = string
  description = "Must be set to the Azure Region of cluster"
}

variable "resource_group_name" {
  type        = string
  description = "Must be set to the resource group name of cluster"
}

variable "subnet_id" {
  type        = string
  description = "Must be set to the `worker_subnet_id` output by cluster"
}

variable "security_group_id" {
  type        = string
  description = "Must be set to the `worker_security_group_id` output by cluster"
}

variable "backend_address_pool_id" {
  type        = string
  description = "Must be set to the `worker_backend_address_pool_id` output by cluster"
}

variable "dns_zone" {
  type        = string
  description = "DNS Zone (e.g. example.com)"
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


# variable "custom_image_resource_group_name" {
#   type        = string
#   description = "The name of the Resource Group in which the Custom Image exists."
# }

# variable "custom_image_name" {
#   type        = string
#   description = "The name of the Custom Image to provision this Virtual Machine from."
# }

# instances

variable "worker_count" {
  type        = number
  default     = 1
  description = "Number of instances"
}

variable "vm_type" {
  type        = string
  default     = "Standard_DS1_v2"
  description = "Machine type for instances (see `az vm list-skus --location centralus`)"
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

variable "priority" {
  type        = string
  default     = "Regular"
  description = "Set priority to Low to use reduced cost surplus capacity, with the tradeoff that instances can be evicted at any time."
}

variable "clc_snippets" {
  type        = list(string)
  description = "Container Linux Config snippets"
  default     = []
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

variable "cpu_manager_policy" {
  description = "CPU Manager policy to use for the worker pool. Possible values: `none`, `static`."
  default     = "none"
  type        = string
}

variable "kube_reserved_cpu" {
  description = "CPU cores reserved for the Worker Kubernetes components like kubelet, etc."
  default     = "300m"
  type        = string
}

variable "system_reserved_cpu" {
  description = "CPU cores reserved for the host services like Docker, sshd, kernel, etc."
  default     = "500m"
  type        = string
}
