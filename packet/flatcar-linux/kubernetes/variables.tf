# DNS records

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

variable "project_id" {
  description = "Packet project ID (e.g. 405efe9c-cce9-4c71-87c1-949c290b27dc)"
}

# Nodes

variable "os_arch" {
  type        = string
  default     = "amd64"
  description = "Flatcar Container Linux architecture to install (amd64, arm64)"
}

variable "os_channel" {
  type        = string
  default     = "stable"
  description = "Flatcar Container Linux channel to install from (stable, beta, alpha, edge)"
}

variable "os_version" {
  type        = string
  default     = "current"
  description = "Flatcar Container Linux version to install (for example '2191.5.0' - see https://www.flatcar-linux.org/releases/), only for iPXE"
}

variable "controller_count" {
  type        = number
  default     = 1
  description = "Number of controllers (i.e. masters)"
}

variable "controller_type" {
  type        = string
  default     = "baremetal_0"
  description = "Packet instance type for controllers"
}

variable "ipxe_script_url" {
  type = string

  # Note: iPXE-booting Flatcar on Packet over HTTPS is failing due to a bug in iPXE.
  # This patch is supposed to fix this: http://git.ipxe.org/ipxe.git/commitdiff/b6ffe28a2
  # However, the upstream fix can work only when the HTTPS server does not rely on elliptic
  # curves. So we should use HTTPS only for servers without elliptic curves, and otherwise
  # use HTTP. Fortunately, since stable.release.flatcar-linux.net does not rely on elliptic
  # curves. it should not be a problem in that case.
  # It has been possible to natively install Flatcar images as official OS option on Packet,
  # but only for amd64. There is no arm64 Flatcar image available on Packet.
  default = ""

  description = "Location to load the pxe boot script from"
}

variable "facility" {
  type        = string
  description = "Packet facility to deploy the cluster in"
}

variable "controller_clc_snippets" {
  type        = list(string)
  description = "Controller Container Linux Config snippets"
  default     = []
}

# Configuration

variable "ssh_keys" {
  type        = list(string)
  description = "SSH public keys for user 'core'"
}

variable "asset_dir" {
  description = "Path to a directory where generated assets should be placed (contains secrets)"
  type        = string
}

variable "networking" {
  description = "Choice of networking provider (flannel or calico)"
  type        = string
  default     = "calico"
}

variable "network_mtu" {
  description = "CNI interface MTU (applies to calico only)"
  type        = number
  default     = 1480
}

variable "network_ip_autodetection_method" {
  description = "Method to autodetect the host IPv4 address (applies to calico only)"
  type        = string
  default     = "first-found"
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

variable "management_cidrs" {
  description = "List of IPv4 CIDRs authorized to access or manage the cluster"
  type        = list(string)
}

variable "node_private_cidr" {
  description = "Private IPv4 CIDR of the nodes used to allow inter-node traffic"
  type        = string
}

variable "enable_aggregation" {
  description = "Enable the Kubernetes Aggregation Layer (defaults to true)"
  type        = bool
  default     = true
}

variable "reservation_ids" {
  description = "Specify Packet hardware_reservation_id for instances. A map where the key format is 'controller-$${index}' and the value is the reservation ID. Nodes not present in the map will use the value of `reservation_ids_default` variable. Example: reservation_ids = { controller-0 = \"<reservation_id>\" }"
  type        = map(string)
  default     = {}
}

variable "reservation_ids_default" {
  description = <<EOD
Possible values: "" and "next-available".

Specify a default reservation ID for nodes not listed in the `reservation_ids`
map. An empty string means "use no hardware reservation". `next-available` will
choose any reservation that matches the pool's device type and facility.
EOD


  type    = string
  default = ""
}

# Certificates

variable "certs_validity_period_hours" {
  description = "Validity of all the certificates in hours"
  type        = number
  default     = 8760
}
