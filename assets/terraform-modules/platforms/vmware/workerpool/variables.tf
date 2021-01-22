# Required variables.
variable "datacenter" {
  type        = string
  description = "The name of the VMware datacenter. This can be a name or path."
}

variable "datastore" {
  type        = string
  description = "The name of the VMware datastore. This can be a name or path."
}

variable "compute_cluster" {
  type        = string
  description = "The name of the VMware computer cluster used for the placement of the virtual machine."
}

variable "network" {
  type        = string
  description = "The name of the VMware network to connect the main interface to. This can be a name or path."
}

variable "template" {
  type        = string
  description = "The name of the VMware template used for the creation of the instance."
}

variable "nodes_ips" {
  type        = list(string)
  description = "IP addresses of the virtual machines."
}

variable "hosts_cidr" {
  type        = string
  description = "CIDR for all hosts."
}

variable "node_count" {
  type        = number
  description = "Number of nodes to create."
}

variable "name" {
  type        = string
  description = "Workerpool name. Must be unique across cluster."
}

variable "cluster_name" {
  type        = string
  description = "Cluster identifier which will be used in controller node names."
}

variable "kubeconfig" {
  type        = string
  description = "Content of kubelet's kubeconfig file."
}

# Optional variables.
variable "folder" {
  type        = string
  description = "The path to the folder to put this virtual machine in, relative to the datacenter that the resource pool is in."
  default     = ""
}

variable "cpus_count" {
  type        = number
  description = "The total number of virtual processor cores to assign to this virtual machine."
  default     = 4
}

variable "memory" {
  type        = number
  description = "The size of the virtual machine's memory, in MB."
  default     = 4096
}

variable "disk_size" {
  type        = number
  description = "The size of the virtual machine's disk, in GB."
  default     = 30
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

variable "dns_servers" {
  type        = list(string)
  description = "DNS servers for the network interface."
  default     = []
}
