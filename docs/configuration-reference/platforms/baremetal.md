---
title: Lokomotive bare metal configuration reference
weight: 10
---

## Introduction

This configuration reference provides information on configuring a Lokomotive cluster on Bare Metal with all
the configuration options available to the user.

## Prerequisites

* `lokoctl` [installed locally.](../../installer/lokoctl.md)
* `kubectl` installed locally to access the Kubernetes cluster.

### Configuration

To create a Lokomotive cluster, we need to define a configuration.

Example configuration file:

```tf
# baremetalcluster.lokocfg
variable "asset_dir" {}
variable "cluster_name" {}
variable "ssh_public_keys" {}
variable "matchbox_ca_path" {}
variable "matchbox_client_cert_path" {}
variable "matchbox_client_key_path" {}
variable "matchbox_endpoint" {}
variable "matchbox_http_endpoint" {}
variable "domain_name" {}
variable "controller_domains" {}
variable "controller_macs" {}
variable "controller_names" {}
variable "worker_domains" {}
variable "worker_macs" {}
variable "worker_names" {}
variable "management_cidrs" {}
variable "node_private_cidr" {}
variable "state_s3_bucket" {}
variable "lock_dynamodb_table" {}
variable "oidc_issuer_url" {}
variable "oidc_client_id" {}
variable "oidc_username_claim" {}
variable "oidc_groups_claim" {}
variable "kernel_args" {}

cluster "bare-metal" {
  asset_dir = var.asset_dir

  cluster_name = var.cluster_name

  ssh_pubkeys = var.ssh_public_keys

  cached_install = "true"

  matchbox_ca_path = var.matchbox_ca_path

  matchbox_client_cert_path = var.matchbox_client_cert_path

  matchbox_client_key_path = var.matchbox_client_key_path

  matchbox_endpoint = var.mathbox_endpoint

  matchbox_http_endpoint = var.matchbox_http_endpoint

  k8s_domain_name = var.domain_name

  labels = {
    "testlabel" = ""
  }

  network_mtu = 1500

  controller_domains = var.controller_domains

  controller_macs = var.controller_macs

  controller_names = var.controller_names

  worker_domains = var.worker_domains

  worker_macs = var.worker_macs

  worker_names = var.worker_names

  os_version = "current"

  os_channel = "stable"

  enable_tls_bootstrap = true

  encrypt_pod_traffic = true

  conntrack_max_per_core = 32768

  oidc {
    issuer_url     = var.oidc_issuer_url
    client_id      = var.oidc_client_id
    username_claim = var.oidc_username_claim
    groups_claim   = var.oidc_groups_claim
  }

  install_disk = "/dev/sdb"

  install_to_smallest_disk = "false"

  kernel_args = var.kernel_args

  network_ip_auto_detection = "can-reach=172.18.169.0"
}
```

**NOTE**: Should you feel differently about the default values, you can set default values using the `variable`
block in the cluster configuration.

Example:

The default for os_version is `current`. If you wish to change the default, then you
define the variable  and use it to refer in the cluster configuration.

```tf
variable "custom_default_os_version" {
  default = "2303.3.1"
}

.
.
.
os_version = var.custom_default_os_version
.
.
.

```

## Attribute reference

| Argument                          | Description                                                                                                                                                                                                                                                                                                  | Default                | Type         | Required |
|-----------------------------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|------------------------|--------------|----------|
| `asset_dir`                       | Location where Lokomotive stores cluster assets.                                                                                                                                                                                                                                                             | -                      | string       | true     |
| `cached_install`                  | Whether the operating system should PXE boot and install from matchbox /assets cache. Note that the admin must have downloaded the `os_version` into matchbox assets.                                                                                                                                        | "false"                | string       | false    |
| `cluster_name`                    | Name of the cluster.                                                                                                                                                                                                                                                                                         | -                      | string       | true     |
| `controller_domains`              | Ordered list of controller FQDNs. Example: ["node1.example.com"]                                                                                                                                                                                                                                             | -                      | list(string) | true     |
| `controller_macs`                 | Ordered list of controller identifying MAC addresses. Example: ["52:54:00:a1:9c:ae"]                                                                                                                                                                                                                         | -                      | list(string) | true     |
| `controller_names`                | Ordered list of controller names. Example: ["node1"]                                                                                                                                                                                                                                                         | -                      | list(string) | true     |
| `k8s_domain_name`                 | Controller DNS name which resolves to a controller instance. Workers and kubeconfig's will communicate with this endpoint. Example: "cluster.example.com"                                                                                                                                                    | -                      | string       | true     |
| `labels`                          | Map of extra Kubernetes Node labels for worker nodes.                                                                                                                                                                                                                                                        | -                      | map(string)  | false    |
| `matchbox_ca_path`                | Path to the CA to verify and authenticate client certificates.                                                                                                                                                                                                                                               | -                      | string       | true     |
| `matchbox_client_cert_path`       | Path to the server TLS certificate file.                                                                                                                                                                                                                                                                     | -                      | string       | true     |
| `matchbox_client_key_path`        | Path to the server TLS key file.                                                                                                                                                                                                                                                                             | -                      | string       | true     |
| `matchbox_endpoint`               | Matchbox API endpoint.                                                                                                                                                                                                                                                                                       | -                      | string       | true     |
| `matchbox_http_endpoint`          | Matchbox HTTP read-only endpoint. Example: "http://matchbox.example.com:8080"                                                                                                                                                                                                                                | -                      | string       | true     |
| `network_mtu`                     | Physical Network MTU.                                                                                                                                                                                                                                                                                        | 1500                   | number       | false    |
| `worker_names`                    | Ordered list of worker names. Example: ["node2", "node3"]                                                                                                                                                                                                                                                    | -                      | list(string) | true     |
| `worker_macs`                     | Ordered list of worker identifying MAC addresses. Example ["52:54:00:b2:2f:86", "52:54:00:c3:61:77"]                                                                                                                                                                                                         | -                      | list(string) | true     |
| `worker_domains`                  | Ordered list of worker FQDNs. Example ["node2.example.com", "node3.example.com"]                                                                                                                                                                                                                             | -                      | list(string) | true     |
| `ssh_pubkeys`                     | List of SSH public keys for user `core`. Each element must be specified in a valid OpenSSH public key format, as defined in RFC 4253 Section 6.6, e.g. "ssh-rsa AAAAB3N...".                                                                                                                                 | -                      | list(string) | true     |
| `os_version`                      | Flatcar Container Linux version to install. Version such as "2303.3.1" or "current".                                                                                                                                                                                                                         | "current"              | string       | false    |
| `os_channel`                      | Flatcar Container Linux channel to install from ("stable", "beta", "alpha", "edge").                                                                                                                                                                                                                         | "stable"               | string       | false    |
| `enable_tls_bootstrap`            | Enable TLS bootstraping for Kubelet.                                                                                                                                                                                                                                                                         | true                   | bool         | false    |
| `encrypt_pod_traffic`             | Enable in-cluster pod traffic encryption. If true `network_mtu` is reduced by 60 to make room for the encryption header.                                                                                                                                                                                     | false                  | bool         | false    |
| `ignore_x509_cn_check`            | Ignore check of common name in x509 certificates. If any application is built pre golang 1.15 then API server rejects x509 from such application, enable this to get around apiserver.                                                                                                                       | false                  | bool         | false    |
| `install_to_smallest_disk`        | Installs Flatcar Container Linux to the smallest disk.                                                                                                                                                                                                                                                       | false                  | bool         | false    |
| `install_disk`                    | Disk device where Flatcar Container Linux should be installed.                                                                                                                                                                                                                                               | "/dev/sda"             | string       | false    |
| `kernel_args`                     | Additional kernel args to provide at PXE boot.                                                                                                                                                                                                                                                               | ["kvm-intel.nested=1"] | list(string) | false    |
| `download_protocol`               | Protocol iPXE uses to download the kernel and initrd. iPXE must be compiled with crypto support for https. Unused if `cached_install` is true. Use `http` if the iPXE boot using `https` is slow or does not work due to SSL settings.                                                                       | "https"                | string       | false    |
| `network_ip_autodetection_method` | Method to detect host IPv4 address. Accepted values include 'first-found', 'can-reach=<DESTINATION>', 'interface=<INTERFACE-REGEX>', 'skip-interface=<INTERFACE-REGEX>, 'cidr=<CIDR>'.                                                                                                                       | "first-found"          | string       | false    |
| `oidc`                            | OIDC configuration block.                                                                                                                                                                                                                                                                                    | -                      | object       | false    |
| `oidc.issuer_url`                 | URL of the provider which allows the API server to discover public signing keys. Only URLs which use the https:// scheme are accepted.                                                                                                                                                                       | -                      | string       | false    |
| `oidc.client_id`                  | A client id that all tokens must be issued for.                                                                                                                                                                                                                                                              | "clusterauth"          | string       | false    |
| `oidc.username_claim`             | JWT claim to use as the user name.                                                                                                                                                                                                                                                                           | "email"                | string       | false    |
| `oidc.groups_claim`               | JWT claim to use as the user’s group.                                                                                                                                                                                                                                                                        | "groups"               | string       | false    |
| `conntrack_max_per_core`          | Maximum number of entries in conntrack table per CPU on all nodes in the cluster. If you require more fain-grained control over this value, set it to 0 and add CLC snippet setting `net.netfilter.nf_conntrack_max sysctl setting per node pool. See [Flatcar documentation about sysctl] for more details. | 32768                  | number       | false    |

## Applying

To create the cluster, execute the following command:

```console
lokoctl cluster apply
```

## Destroying

To destroy the Lokomotive cluster, execute the following command:

```console
lokoctl cluster destroy --confirm
```

[Flatcar documentation about sysctl](https://docs.flatcar-linux.org/os/other-settings/#tuning-sysctl-parameters)
