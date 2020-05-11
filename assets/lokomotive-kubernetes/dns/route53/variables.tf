variable "cluster_name" {
  type        = string
  description = "Unique cluster name (prepended to dns_zone)"
}

variable "controllers_public_ipv4" {
  type        = list(string)
  description = "Public IPv4 addresses of all the controllers in the cluster"
}

variable "controllers_private_ipv4" {
  type        = list(string)
  description = "Private IPv4 addresses of all the controllers in the cluster"
}

variable "dns_zone" {
  type        = string
  description = "Zone name under which records should be created (e.g. example.com)"
}
