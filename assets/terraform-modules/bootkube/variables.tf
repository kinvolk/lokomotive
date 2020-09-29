variable "cluster_name" {
  description = "Cluster name"
  type        = string
}

variable "api_servers" {
  description = "List of domain names used to reach kube-apiserver from within the cluster"
  type        = list(string)
}

# When not set, the value of var.api_servers will be used.
variable "api_servers_external" {
  description = "List of domain names used to reach kube-apiserver from an external network"
  type        = list(string)
  default     = []
}

variable "api_servers_ips" {
  description = "List of additional IPv4 addresses to be included in the kube-apiserver TLS certificate"
  type        = list(string)
  default     = []
}

variable "etcd_servers" {
  description = "List of domain names used to reach etcd servers."
  type        = list(string)
}

variable "etcd_endpoints" {
  description = "List of Private IPv4 addresses of the controller nodes running etcd."
  type        = list(string)
  default     = []
}

variable "asset_dir" {
  description = "Path to a directory where generated assets should be placed (contains secrets)"
  type        = string
}

variable "cloud_provider" {
  description = "The provider for cloud services (empty string for no provider)"
  type        = string
  default     = ""
}

variable "network_mtu" {
  description = "CNI interface MTU"
  type        = number
  default     = 1500
}

variable "network_encapsulation" {
  description = "Network encapsulation mode either ipip or vxlan (only applies to calico)"
  type        = string
  default     = "ipip"
}

variable "network_ip_autodetection_method" {
  description = "Method to autodetect the host IPv4 address (only applies to calico)"
  type        = string
  default     = "first-found"
}

variable "pod_cidr" {
  description = "CIDR IP range to assign Kubernetes pods"
  type        = string
  default     = "10.2.0.0/16"
}

variable "service_cidr" {
  description = <<EOD
CIDR IP range to assign Kubernetes services.
The 1st IP will be reserved for kube_apiserver, the 10th IP will be reserved for kube-dns.
EOD


  type    = string
  default = "10.3.0.0/24"
}

variable "cluster_domain_suffix" {
  description = "Queries for domains with the suffix will be answered by kube-dns"
  type        = string
  default     = "cluster.local"
}

variable "container_arch" {
  description = "Architecture suffix for the container image coredns/coredns:coredns- (e.g., arm64)"
  type        = string
  default     = "amd64"
}

variable "container_images" {
  description = "Container images to use (the coredns entry will get -$${var.container_arch} appended)"
  type        = map(string)

  default = {
    calico                  = "calico/node:v3.15.2"
    calico_cni              = "calico/cni:v3.15.2"
    calico_controllers      = "calico/kube-controllers:v3.15.2"
    flexvol_driver_image    = "calico/pod2daemon-flexvol:v3.15.2"
    kubelet_image           = "quay.io/poseidon/kubelet:v1.19.2"
    coredns                 = "coredns/coredns:coredns-"
    pod_checkpointer        = "kinvolk/pod-checkpointer:d1c58443fe7d7d33aa5bf7d80d65d299be6e5847"
    kube_apiserver          = "k8s.gcr.io/kube-apiserver:v1.19.2"
    kube_controller_manager = "k8s.gcr.io/kube-controller-manager:v1.19.2"
    kube_scheduler          = "k8s.gcr.io/kube-scheduler:v1.19.2"
    kube_proxy              = "k8s.gcr.io/kube-proxy:v1.19.2"
  }
}

variable "enable_reporting" {
  type        = bool
  description = "Enable usage or analytics reporting to upstream component owners (Tigera: Calico)"
  default     = false
}

variable "trusted_certs_dir" {
  description = "Path to the directory on cluster nodes where trust TLS certs are kept"
  type        = string
  default     = "/usr/share/ca-certificates"
}

variable "certs_validity_period_hours" {
  description = "Validity of all the certificates in hours"
  type        = number
  default     = 8760
}

variable "enable_aggregation" {
  description = "Enable the Kubernetes Aggregation Layer (defaults to false, recommended)"
  type        = bool
  default     = false
}

# unofficial, temporary, may be removed without notice

variable "external_apiserver_port" {
  description = "External kube-apiserver port (e.g. 6443 to match internal kube-apiserver port)"
  type        = number
  default     = 6443
}

variable "disable_self_hosted_kubelet" {
  description = "Disable the self hosted kubelet installed by default"
  type        = bool
}

variable "kube_apiserver_extra_flags" {
  description = "Extra flags passed to self-hosted kube-apiserver."
  type        = list(string)
  default     = []
}

variable "blocked_metadata_cidrs" {
  description = "List of platform metadata CIDRs to block access to for all pods"
  type        = list(string)
  default     = []
}

variable "bootstrap_tokens" {
  description = "List of bootstrap tokens for all the nodes in the cluster in the form token-id and token-secret."
  type        = list(any)
}

variable "enable_tls_bootstrap" {
  description = "Enable TLS Bootstrap for Kubelet."
  type        = bool
}

variable "failsafe_inbound_host_ports" {
  description = "UDP/TCP/SCTP protocol/port pairs to allow incoming traffic on regardless of the security policy."
  type        = list(any)
  default     = null
}

variable "encrypt_pod_traffic" {
  description = "Enable in-cluster pod traffic encryption."
  type        = bool
  default     = false
}
