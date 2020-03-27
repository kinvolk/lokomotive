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
    calico                  = "calico/node:v3.13.1"
    calico_cni              = "calico/cni:v3.13.1"
    hyperkube               = "k8s.gcr.io/hyperkube:v1.18.0"
    coredns                 = "coredns/coredns:coredns-"
    pod_checkpointer        = "kinvolk/pod-checkpointer:83e25e5968391b9eb342042c435d1b3eeddb2be1"
    kube_apiserver          = "k8s.gcr.io/kube-apiserver:v1.18.0"
    kube_controller_manager = "k8s.gcr.io/kube-controller-manager:v1.18.0"
    kube_scheduler          = "k8s.gcr.io/kube-scheduler:v1.18.0"
    kube_proxy              = "k8s.gcr.io/kube-proxy:v1.17.4"
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

variable "expose_on_all_interfaces" {
  description = "If true, kube-apiserver will be exposed on all controller node interfaces on port 6443. If false, it will be exposed only one kubelet's node IP."
  type        = bool
  default     = false
}
