variable "http_endpoint" {
  type        = string
  description = "Matchbox HTTP read-only endpoint (e.g. http://matchbox.example.com:8080)."
}

variable "os_channel" {
  type        = string
  description = "Flatcar Container Linux channel to install from (stable, beta, alpha, edge)."
  default     = "stable"
}

variable "os_version" {
  type        = string
  description = "Flatcar Container Linux version to install (for example '2191.5.0' - see https://www.flatcar-linux.org/releases/)."
  default     = "current"
}

variable "download_protocol" {
  type        = string
  description = "Protocol iPXE should use to download the kernel and initrd. Defaults to https, which requires iPXE compiled with crypto support. Unused if cached_install is true."
  default     = "https"
}

variable "cached_install" {
  type        = bool
  description = "Whether the operating system should PXE boot and install from matchbox /assets cache. Note that the admin must have downloaded the os_version into matchbox assets."
  default     = false
}

variable "install_disk" {
  type        = string
  description = "Disk device to which the install profiles should install the operating system (e.g. /dev/sda)."
  default     = "/dev/sda"
}

variable "container_linux_oem" {
  type        = string
  description = "DEPRECATED: Specify an OEM image id to use as base for the installation (e.g. ami, vmware_raw, xen) or leave blank for the default image."
  default     = ""
}

variable "kernel_args" {
  type        = list(string)
  description = "Additional kernel arguments to provide at PXE boot."
  default     = []
}

variable "install_to_smallest_disk" {
  type        = bool
  description = "Install Flatcar Container Linux to the smallest disk."
  default     = false
}

variable "ssh_keys" {
  type        = list(string)
  description = "SSH public keys for user 'core'."
}

variable "ignition_clc_config" {
  type        = string
  description = "Ignition CLC snippets to include in the configuration."
}

variable "node_name" {
  type        = string
  description = "Name of the node/machine."
}

variable "node_mac" {
  type        = string
  description = "MAC address identifying the node/machine (e.g. 52:54:00:a1:9c:ae)."
}

variable "node_domain" {
  type        = string
  description = "Node FQDN (e.g node1.example.com)."
}

variable "pxe_commands" {
  type        = string
  description = "shell commands to execute for PXE (re)provisioning, with access to the variables $mac (the MAC address), $name (the node name), and $domain (the domain name), e.g., 'bmc=bmc-$domain; ipmitool -H $bmc power off; ipmitool -H $bmc chassis bootdev pxe; ipmitool -H $bmc power on'."
  default     = "echo 'you must (re)provision the node by booting via iPXE from http://MATCHBOX/boot.ipxe'; exit 1"
}

variable "install_pre_reboot_cmds" {
  type        = string
  description = "shell commands to execute on the provisioned host after installation finished and before reboot, e.g., docker run --privileged --net host --rm debian sh -c 'apt update && apt install -y ipmitool && ipmitool chassis bootdev disk options=persistent'."
  default     = "true"
}

variable "kernel_console" {
  type        = list(string)
  description = "The kernel arguments to configure the console at PXE boot and in /usr/share/oem/grub.cfg."
  default     = ["console=tty0", "console=ttyS0"]
}

variable "ignore_changes" {
  description = "When set to true, ignores the reprovisioning of the node."
  type        = bool
  default     = false
}

variable "asset_dir" {
  description = "Path to a directory where generated assets should be placed (contains secrets)"
  type        = string
}
