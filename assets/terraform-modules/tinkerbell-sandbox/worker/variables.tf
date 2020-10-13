variable "name" {
  type        = string
  description = "Worker hostname."
}

variable "ip" {
  type        = string
  description = "IP address to assign to the node."
}

variable "sandbox" {
  type = object({
    sandbox_name      = string
    volumes_pool_name = string
    network_id        = string
    netmask           = string
    gateway           = string
  })
  description = "Output from main sandbox module."
}
