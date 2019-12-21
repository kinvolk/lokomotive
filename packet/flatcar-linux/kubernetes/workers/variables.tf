# Cluster

variable "cluster_name" {
  type        = string
  description = "Unique cluster name (prepended to dns_zone)"
}

variable "project_id" {
  description = "Packet project ID (e.g. 405efe9c-cce9-4c71-87c1-949c290b27dc)"
}

# Nodes

variable "pool_name" {
  type        = string
  description = "Unique worker pool name (prepended to hostname)"
}

variable "worker_count" {
  type        = number
  default     = 1
  description = "Number of workers, can be changed afterwards to delete or add nodes"
}

variable "type" {
  type        = string
  default     = "baremetal_0"
  description = "Packet instance type for workers, can be changed afterwards to recreate the nodes"
}

variable "clc_snippets" {
  type        = list(string)
  description = "Container Linux Config snippets"
  default     = []
}

# TODO: migrate to `templatefile` when Terraform `0.12` is out and use `{% for ~}`
# to avoid specifying `--node-labels` again when the var is empty.
variable "labels" {
  type        = string
  default     = ""
  description = "Custom labels to assign to worker nodes. Provide comma separated key=value pairs as labels. e.g. 'foo=oof,bar=,baz=zab'"
}

variable "taints" {
  type        = string
  default     = ""
  description = "Comma separated list of taints. eg. 'clusterType=staging:NoSchedule,nodeType=storage:NoSchedule'"
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

variable "cluster_domain_suffix" {
  description = "Queries for domains with the suffix will be answered by coredns. Default is cluster.local (e.g. foo.default.svc.cluster.local) "
  type        = string
  default     = "cluster.local"
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
  default = "10.3.0.0/16"
}

variable "setup_raid" {
  description = "Attempt to create a RAID 0 from extra disks to be used for persistent container storage. Can't be used with setup_raid_hdd nor setup_raid_sdd. Valid values: \"true\", \"false\""
  type        = bool
  default     = false
}

variable "setup_raid_hdd" {
  description = "Attempt to create a RAID 0 from extra Hard Disk drives only, to be used for persistent container storage. Can't be used with setup_raid nor setup_raid_sdd. Valid values: \"true\", \"false\""
  type        = bool
  default     = false
}

variable "setup_raid_ssd" {
  description = "Attempt to create a RAID 0 from extra Solid State Drives only, to be used for persistent container storage.  Can't be used with setup_raid nor setup_raid_hdd. Valid values: \"true\", \"false\""
  type        = bool
  default     = false
}

variable "setup_raid_ssd_fs" {
  description = "When set to \"true\" file system will be created on SSD RAID device and will be mounted on /mnt/node-local-ssd-storage. To use the raw device set it to \"false\". Valid values: \"true\", \"false\""
  type        = bool
  default     = true
}

variable "reservation_ids" {
  description = "Specify Packet hardware_reservation_id for instances. A map where the key format is 'worker-$${index}' and the value is the reservation ID. Nodes not present in the map will use the value of `reservation_ids_default` variable. Example: reservation_ids = { worker-0 = \"<reservation_id>\" }"
  type        = map(string)
  default     = {}
}

variable "reservation_ids_default" {
  description = <<EOD
Possible values: "" and "next-available".

Specify a default reservation ID for nodes not listed in the `reservation_ids`
map. An empty string means "use no hardware reservation". `next-available` will
choose any reservation that matches the worker pool's device type and facility.
EOD


  type    = string
  default = ""
}

variable "disable_bgp" {
  description = "Disable BGP on nodes. Nodes won't be able to connect to Packet BGP peers. Note that BGP is used to receive internet traffic directed to Packet elastic IPs"
  type        = bool
  default     = false
}
