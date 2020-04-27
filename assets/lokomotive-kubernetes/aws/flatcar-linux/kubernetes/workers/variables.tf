variable "cluster_name" {
  type        = string
  description = "Cluster name (prepended to pool name)"
}

variable "pool_name" {
  type        = string
  description = "Unique name for the worker pool"
}

# AWS

variable "vpc_id" {
  type        = string
  description = "Must be set to `vpc_id` output by cluster"
}

variable "subnet_ids" {
  type        = list(string)
  description = "Must be set to `subnet_ids` output by cluster"
}

variable "security_groups" {
  type        = list(string)
  description = "Must be set to `worker_security_groups` output by cluster"
}

# instances

variable "worker_count" {
  type        = number
  default     = 1
  description = "Number of instances"
}

variable "instance_type" {
  type        = string
  default     = "t3.small"
  description = "EC2 instance type"
}

variable "os_name" {
  type        = string
  default     = "flatcar"
  description = "Name of Operating System (coreos or flatcar)"
}

variable "os_channel" {
  type        = string
  default     = "stable"
  description = "AMI channel for the OS (stable, beta, alpha, edge)"
}

variable "os_version" {
  type        = string
  default     = "current"
  description = "Version of the OS (current or numeric version such as 2261.99.0)"
}

variable "disk_size" {
  type        = number
  default     = 40
  description = "Size of the EBS volume in GB"
}

variable "disk_type" {
  type        = string
  default     = "gp2"
  description = "Type of the EBS volume (e.g. standard, gp2, io1)"
}

variable "disk_iops" {
  type        = number
  default     = 0
  description = "IOPS of the EBS volume (required for io1)"
}

variable "spot_price" {
  type        = string
  default     = ""
  description = "Spot price in USD for autoscaling group spot instances. Leave as default empty string for autoscaling group to use on-demand instances. Note, switching in-place from spot to on-demand is not possible: https://github.com/terraform-providers/terraform-provider-aws/issues/4320"
}

variable "target_groups" {
  type        = list(string)
  description = "Additional target group ARNs to which instances should be added"
  default     = []
}

variable "clc_snippets" {
  type        = list(string)
  description = "Container Linux Config snippets"
  default     = []
}

variable "tags" {
  type = map
  default = {
    "ManagedBy" = "Lokomotive"
    "CreatedBy" = "Unspecified"
  }
  description = "Optional details to tag on AWS resources"
}

# configuration

variable "kubeconfig" {
  type        = string
  description = "Must be set to `kubeconfig` output by cluster"
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

variable "lb_arn" {
  description = "ARN of the load balancer on which to create listeners for this worker pool"
}

variable "lb_http_port" {
  description = "Port the load balancer should listen on for HTTP connections"
  type        = number
  default     = 80
}

variable "lb_https_port" {
  description = "Port the load balancer should listen on for HTTPS connections"
  type        = number
  default     = 443
}
