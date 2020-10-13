# Required variables.
variable "name" {
  type        = string
  description = "Unique identifier of created sandbox used to identify created libvirt objects."
}

variable "ssh_keys" {
  type        = list(string)
  description = "List of SSH keys for provisioner machine for core and root users."
}

variable "hosts_cidr" {
  type        = string
  description = "CIDR for all hosts."
}

variable "flatcar_image_path" {
  type        = string
  description = "Path to Flatcar image."
}

variable "pool_path" {
  type        = string
  description = "Path to store virtual machines disk images."
}

variable "dns_hosts" {
  type = list(object({
    hostname = string
    ip       = string
  }))
  description = "List of DNS entries to add to libvirt DNS server."
}
