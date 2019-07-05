# DNS records

variable "cluster_name" {
  type        = "string"
  description = "Unique cluster name (prepended to dns_zone)"
}

variable "dns_zone" {
  type        = "string"
  description = "AWS Route53 DNS Zone (e.g. aws.example.com)"
}

variable "dns_zone_id" {
  type        = "string"
  description = "AWS Route53 DNS Zone ID (e.g. Z3PAABBCFAKEC0)"
}

variable "project_id" {
  description = "Packet project ID (e.g. 405efe9c-cce9-4c71-87c1-949c290b27dc)"
}

# Nodes

variable "os_channel" {
  type        = "string"
  default     = "stable"
  description = "Flatcar Linux channel to install from (stable, beta, alpha, edge)"
}

variable "controller_count" {
  type        = "string"
  default     = "1"
  description = "Number of controllers (i.e. masters)"
}

variable "worker_count" {
  type        = "string"
  description = "Number of workers"
}

variable "worker_nodes_hostnames" {
  type        = "list"
  description = "List of hostname of packet_device resources"
}

variable "controller_type" {
  type        = "string"
  default     = "baremetal_0"
  description = "Packet instance type for controllers"
}

variable "facility" {
  type        = "string"
  description = "Packet facility to deploy the cluster in"
}

# Configuration

variable "ssh_keys" {
  type        = "list"
  description = "SSH public keys for user 'core'"
}

variable "asset_dir" {
  description = "Path to a directory where generated assets should be placed (contains secrets)"
  type        = "string"
}

variable "networking" {
  description = "Choice of networking provider (flannel or calico)"
  type        = "string"
  default     = "calico"
}

variable "network_mtu" {
  description = "CNI interface MTU (applies to calico only)"
  type        = "string"
  default     = "1480"
}

variable "network_ip_autodetection_method" {
  description = "Method to autodetect the host IPv4 address (applies to calico only)"
  type        = "string"
  default     = "first-found"
}

variable "pod_cidr" {
  description = "CIDR IPv4 range to assign Kubernetes pods"
  type        = "string"
  default     = "10.2.0.0/16"
}

variable "service_cidr" {
  description = <<EOD
CIDR IPv4 range to assign Kubernetes services.
The 1st IP will be reserved for kube_apiserver, the 10th IP will be reserved for coredns.
EOD

  type    = "string"
  default = "10.3.0.0/16"
}

variable "cluster_domain_suffix" {
  description = "Queries for domains with the suffix will be answered by coredns. Default is cluster.local (e.g. foo.default.svc.cluster.local) "
  type        = "string"
  default     = "cluster.local"
}

variable "enable_reporting" {
  type        = "string"
  description = "Enable usage or analytics reporting to upstreams (Calico)"
  default     = "false"
}

variable "management_cidrs" {
  description = "List of IPv4 CIDRs authorized to access or manage the cluster"
  type        = "list"
}

variable "node_private_cidr" {
  description = "Private IPv4 CIDR of the nodes used to allow inter-node traffic"
  type        = "string"
}

variable "enable_aggregation" {
  description = "Enable the Kubernetes Aggregation Layer (defaults to false)"
  type        = "string"
  default     = "false"
}
