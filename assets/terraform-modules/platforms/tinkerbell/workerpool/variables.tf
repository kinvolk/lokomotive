# Required variables.
variable "name" {
  type        = string
  description = "Workerpool name. Must be unique across cluster."
}

variable "cluster_name" {
  type        = string
  description = "Cluster identifier which will be used in controller node names."
}

variable "ip_addresses" {
  type        = list(string)
  description = "List of IP addresses of Tinkerbell workers where controller nodes should be provisioned."
}

variable "kubeconfig" {
  type        = string
  description = "Content of kubelet's kubeconfig file."
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

variable "cluster_domain_suffix" {
  type        = string
  description = "Cluster domain suffix. Passed to kubelet as --cluster_domain flag."
  default     = "cluster.local"
}

variable "host_dns_ip" {
  type        = string
  description = "IP address of DNS server to configure on the nodes."
  default     = "8.8.8.8"
}

variable "ca_cert" {
  description = "Kubernetes CA certificate needed in the kubeconfig file."
  type        = string
}

variable "apiserver" {
  description = "Apiserver private endpoint needed in the kubeconfig file."
  type        = string
}
