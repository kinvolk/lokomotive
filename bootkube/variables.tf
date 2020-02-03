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

variable "networking" {
  description = "Choice of networking provider (flannel or calico or kube-router)"
  type        = string
  default     = "flannel"
}

variable "network_mtu" {
  description = "CNI interface MTU (only applies to calico and kube-router)"
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
  description = "Architecture suffix for the container images coredns/coredns:coredns- and quay.io/coreos/flannel:v0.11.0- (e.g., arm64)"
  type        = string
  default     = "amd64"
}

variable "container_images" {
  description = "Container images to use (the coredns and flannel entry will get -$${var.container_arch} appended)"
  type        = map(string)

  default = {
    calico     = "calico/node:v3.11.2"
    calico_cni = "calico/cni:v3.11.2"
    flannel    = "quay.io/coreos/flannel:v0.11.0-"
    # only amd64 images available for flannel_cni
    flannel_cni = "quay.io/coreos/flannel-cni:v0.3.0"
    # only amd64 images available for cloudnativelabs/kube-router
    kube_router      = "cloudnativelabs/kube-router:v0.3.2"
    hyperkube        = "k8s.gcr.io/hyperkube:v1.17.2"
    coredns          = "coredns/coredns:coredns-"
    pod_checkpointer = "kinvolk/pod-checkpointer:83e25e5968391b9eb342042c435d1b3eeddb2be1"
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
