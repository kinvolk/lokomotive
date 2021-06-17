variable "cluster_name" {
  type        = string
  description = "Unique cluster name"
}

# bare-metal

variable "matchbox_http_endpoint" {
  type        = string
  description = "Matchbox HTTP read-only endpoint (e.g. http://matchbox.example.com:8080)"
}

variable "os_channel" {
  type        = string
  default     = "stable"
  description = "Flatcar Container Linux channel to install from (stable, beta, alpha, edge)"
}

variable "os_version" {
  type        = string
  default     = "current"
  description = "Flatcar Container Linux version to install (for example '2191.5.0' - see https://www.flatcar-linux.org/releases/)"
}

# machines
# Terraform's crude "type system" does not properly support lists of maps so we do this.

variable "controller_names" {
  type        = list(string)
  description = "Ordered list of controller names (e.g. [node1])"
}

variable "controller_macs" {
  type        = list(string)
  description = "Ordered list of controller identifying MAC addresses (e.g. [52:54:00:a1:9c:ae])"
}

variable "controller_domains" {
  type        = list(string)
  description = "Ordered list of controller FQDNs (e.g. [node1.example.com])"
}

variable "worker_names" {
  type        = list(string)
  description = "Ordered list of worker names (e.g. [node2, node3])"
}

variable "worker_macs" {
  type        = list(string)
  description = "Ordered list of worker identifying MAC addresses (e.g. [52:54:00:b2:2f:86, 52:54:00:c3:61:77])"
}

variable "worker_domains" {
  type        = list(string)
  description = "Ordered list of worker FQDNs (e.g. [node2.example.com, node3.example.com])"
}

variable "clc_snippets" {
  type        = map(list(string))
  description = "Map from machine names to lists of Container Linux Config snippets"
  default     = {}
}

variable "installer_clc_snippets" {
  type        = map(list(string))
  description = "Map from machine names to lists of Container Linux Config snippets, applied for the PXE-booted installer OS"
  default     = {}
}

variable "labels" {
  type        = map(string)
  description = "Map of labels for worker nodes."
  default     = {}
}

# configuration

variable "k8s_domain_name" {
  description = "Controller DNS name which resolves to a controller instance. Workers and kubeconfig's will communicate with this endpoint (e.g. cluster.example.com)"
  type        = string
}

variable "ssh_keys" {
  type        = list(string)
  description = "SSH public keys for user 'core'"
}

variable "asset_dir" {
  description = "Path to a directory where generated assets should be placed (contains secrets)"
  type        = string
}

variable "network_mtu" {
  description = "Physical Network MTU."
  type        = number
}

variable "network_ip_autodetection_method" {
  description = "Method to autodetect the host IPv4 address"
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

# optional

variable "cluster_domain_suffix" {
  description = "Queries for domains with the suffix will be answered by coredns. Default is cluster.local (e.g. foo.default.svc.cluster.local) "
  type        = string
  default     = "cluster.local"
}

variable "download_protocol" {
  type        = string
  default     = "https"
  description = "Protocol iPXE should use to download the kernel and initrd. Defaults to https, which requires iPXE compiled with crypto support. Unused if cached_install is true."
}

variable "cached_install" {
  type        = bool
  default     = false
  description = "Whether the operating system should PXE boot and install from matchbox /assets cache. Note that the admin must have downloaded the os_version into matchbox assets."
}

variable "install_disk" {
  type        = string
  default     = "/dev/sda"
  description = "Disk device to which the install profiles should install the operating system (e.g. /dev/sda)"
}

variable "container_linux_oem" {
  type        = string
  default     = ""
  description = "DEPRECATED: Specify an OEM image id to use as base for the installation (e.g. ami, vmware_raw, xen) or leave blank for the default image"
}

variable "kernel_args" {
  description = "Additional kernel arguments to provide at PXE boot."
  type        = list(string)
  default     = []
}

variable "kernel_console" {
  description = "The kernel arguments to configure the console at PXE boot and in /usr/share/oem/grub.cfg."
  type        = list(string)
  default     = ["console=tty0", "console=ttyS0"]
}

variable "enable_reporting" {
  type        = bool
  description = "Enable usage or analytics reporting to upstreams (Calico)"
  default     = false
}

variable "enable_aggregation" {
  description = "Enable the Kubernetes Aggregation Layer (defaults to true)"
  type        = bool
  default     = true
}

variable "disable_self_hosted_kubelet" {
  description = "Disable the self hosted kubelet installed by default"
  type        = bool
}

# Certificates

variable "certs_validity_period_hours" {
  description = "Validity of all the certificates in hours"
  type        = number
  default     = 8760
}

variable "kube_apiserver_extra_flags" {
  description = "Extra flags passed to self-hosted kube-apiserver."
  type        = list(string)
  default     = []
}

variable "encrypt_pod_traffic" {
  description = "Enable in-cluster pod traffic encryption."
  type        = bool
  default     = false
}

variable "ignore_x509_cn_check" {
  description = "Ignore CN checks in x509 certificates."
  type        = bool
  default     = false
}

variable "conntrack_max_per_core" {
  description = "--conntrack-max-per-core value for kube-proxy. Maximum number of NAT connections to track per CPU core (0 to leave the limit as-is and ignore the conntrack-min kube-proxy flag)."
  type        = number
}

variable "install_to_smallest_disk" {
  description = "Install Flatcar Container Linux to the smallest disk."
  type        = bool
  default     = false
}

variable "node_specific_labels" {
  type        = map(map(string))
  description = "Map of node specific labels map."
  default     = {}
}

variable "wipe_additional_disks" {
  type        = bool
  description = "Wipes any additional disks attached, if set to true"
  default     = false
}
