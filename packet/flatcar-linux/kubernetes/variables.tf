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

# instances
variable "controller_count" {
  type        = "string"
  default     = "1"
  description = "Number of controllers (i.e. masters)"
}

variable "worker_count" {
  type        = "string"
  default     = "1"
  description = "Number of workers"
}

variable "controller_type" {
  type        = "string"
  default     = "baremetal_0"
  description = "Packet instance type for controllers"
}

variable "worker_type" {
  type        = "string"
  default     = "baremetal_0"
  description = "Packet instance type for workers"
}

variable "ipxe_script_url" {
  type        = "string"
  default     = "https://raw.githubusercontent.com/kinvolk/flatcar-ipxe-scripts/4fe69534f69013b9681d8da7e61853407e4c1c59/packet.ipxe"
  description = "Location to load the pxe boot script from"
}

variable "cluster_region" {
  type        = "string"
  default     = "ams1"
  description = "Location of the packet datacenter"
}

# configuration
variable "ssh_keys" {
  type        = "list"
  description = "SSH public keys for user 'core'"
}
