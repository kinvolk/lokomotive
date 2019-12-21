# Cluster

variable "cluster_name" {
  type        = string
  description = "Unique cluster name (prepended to dns_zone)"
}

variable "machine_domain" {
  type        = string
  description = "Machine domain"
}

variable "pool_name" {
  type        = string
  description = "Unique worker pool name (prepended to hostname)"
}

variable "worker_count" {
  type        = string
  default     = "1"
  description = "Number of workers"
}

variable "virtual_cpus" {
  type        = string
  default     = "1"
  description = "Number of virtual CPUs"
}

variable "virtual_memory" {
  type        = string
  default     = "2048"
  description = "Virtual RAM in MB"
}

# TODO: migrate to `templatefile` when Terraform `0.12` is out and use `{% for ~}`
# to avoid specifying `--node-labels` again when the var is empty.
variable "labels" {
  type        = string
  default     = ""
  description = "Custom labels to assign to worker nodes. Provide comma separated key=value pairs as labels. e.g. 'foo=oof,bar=,baz=zab'"
}

variable "libvirtpool" {
  type        = string
  description = "libvirt volume pool with base image"
}

variable "libvirtbaseid" {
  type        = string
  description = "base image id for libvirt"
}

variable "cluster_domain_suffix" {
  description = "Queries for domains with the suffix will be answered by coredns. Default is cluster.local (e.g. foo.default.svc.cluster.local) "
  type        = string
  default     = "cluster.local"
}

variable "clc_snippets" {
  type        = list(string)
  description = "Container Linux Config snippets"
  default     = []
}

variable "kubeconfig" {
  description = "Kubeconfig file"
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
  default = "10.2.0.0/16"
}

