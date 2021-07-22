---
title: Lokomotive Tinkerbell configuration reference
weight: 10
---

## Introduction

This configuration reference provides information on configuring a Lokomotive cluster on bare metal
using Tinkerbell with all the configuration options available to the user.

## Prerequisites

* `lokoctl` [installed locally](../../installer/lokoctl.md)
* `kubectl` installed locally to access the Kubernetes cluster.

### Configuration

To create a Lokomotive cluster, we need to define a configuration.

Example configuration file:

```tf
# mycluster.lokocfg
variable "state_s3_bucket" {}
variable "state_s3_key" {}
variable "state_s3_region" {}
variable "lock_dynamodb_table" {}

backend "s3" {
  bucket         = var.state_s3_bucket
  key            = var.state_s3_key
  region         = var.state_s3_region
  dynamodb_table = var.lock_dynamodb_table
}

# backend "local" {
#   path = "path/to/local/file"
#}

variable "asset_dir" {}
variable "cluster_name" {}
variable "dns_zone" {}
variable "ssh_public_keys" {}
variable "controller_ip_addresses" {}
variable "controller_clc_snippets" {}
variable "controller_flatcar_install_base_url" {}
variable "os_channel" {}
variable "os_version" {}
variable "hosts_cidr" {}
variable "flatcar_image_path" {}
variable "pool_path" {}
variable "enable_aggregation" {}
variable "enable_reporting" {}
variable "pod_cidr" {}
variable "service_cidr" {}
variable "cluster_domain_suffix" {}
variable "certs_validity_period_hours" {}
variable "network_mtu" {}
variable "disable_self_hosted_kubelet" {}
variable "ip_addresses" {}
variable "flatcar_install_base_url" {}
variable "clc_snippets" {}
variable "labels" {}
variable "taints" {}

cluster "tinkerbell" {
  asset_dir = var.asset_dir

  name = var.cluster_name

  dns_zone = var.dns_zone

  ssh_public_keys = var.ssh_public_keys

  controller_ip_addresses = var.controller_ip_addresses

  controller_clc_snippets = var.controller_clc_snippets

  controller_flatcar_install_base_url = var.controller_flatcar_install_base_url

  os_channel = var.os_channel

  os_version = var.os_version

  experimental_sandbox {
    hosts_cidr         = var.hosts_cidr
    flatcar_image_path = var.flatcar_image_path
    pool_path          = var.pool_path
  }

  enable_aggregation = var.enable_aggregation

  disable_self_hosted_kubelet = var.disable_self_hosted_kubelet

  enable_reporting = var.enable_reporting

  pod_cidr = var.pod_cidr

  service_cidr = var.service_cidr

  cluster_domain_suffix = var.cluster_domain_suffix

  certs_validity_period_hours = var.certs_validity_period_hours

  network_mtu = var.network_mtu

  disable_self_hosted_kubelet = var.disable_self_hosted_kubelet

  worker_pool "pool1" {
    ip_addresses = var.ip_addresses

    ssh_public_keys = var.ssh_public_keys

    os_channel = var.os_channel

    os_version = var.os_version

    flatcar_install_base_url = var.flatcar_install_base_url

    clc_snippets = var.clc_snippets

    labels = var.labels

    taints = var.taints
  }
}
```

## Attribute reference

| Argument                                  | Description                                                                                                                                                                                                                                | Default         | Type         | Required |
|-------------------------------------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|-----------------|--------------|----------|
| `asset_dir`                               | Location where Lokomotive stores cluster assets.                                                                                                                                                                                           | -               | string       | true     |
| `name`                                    | Name of the cluster.                                                                                                                                                                                                                       | -               | string       | true     |
| `dns_zone`                                | DNS Zone name which will be used for cluster DNS entries. E.g. If you set it to "example.com", then `<name>.example.com` must be set to point to `controller_ip_addresses`. With `experimental_sandbox` DNS entries are set automatically. | -               | string       | true     |
| `ssh_public_keys`                         | List of SSH public keys for user `core` on controller nodes. Each element must be specified in a valid OpenSSH public key format, as defined in RFC 4253 Section 6.6, e.g. "ssh-rsa AAAAB3N...".                                           | -               | list(string) | true     |
| `controller_ip_addresses`                 | List of IP addresses of Tinkerbell hardware to be used for controller nodes. With `experimental_sandbox`, machines will be created with these IP addresses.                                                                                | -               | list(string) | true     |
| `controller_clc_snippets`                 | Controller Flatcar Container Linux Config snippets.                                                                                                                                                                                        | []              | list(string) | false    |
| `controller_flatcar_install_base_url`     | URL passed to the `flatcar-install` script to fetch Flatcar images from.                                                                                                                                                                   | -               | string       | false    |
| `os_channel`                              | Flatcar Container Linux channel to install from (stable, beta, alpha, edge).                                                                                                                                                               | "stable"        | string       | false    |
| `os_version`                              | Flatcar Container Linux version to install. Version such as "2303.3.1" or "current".                                                                                                                                                       | "current"       | string       | false    |
| `experimental_sandbox`                    | Configuration block for experimental local Tinkerbell sandbox setup using libvirt.                                                                                                                                                         | -               | object       | false    |
| `experimental_sandbox.hosts_cidr`         | CIDR for all hosts in the cluster, which will be NATed to the outside world for internet access.                                                                                                                                           | -               | string       | true     |
| `experimental_sandbox.flatcar_image_path` | Absolute path on the local filesystem to an unpacked Flatcar QEMU image, which will be used as a base OS image for Tinkerbell provisioner server.                                                                                          | -               | string       | true     |
| `experimental_sandbox.pool_path           | Absolute path on the local filesystem where all VM disk images will be stored. At least 25GB of free disk space is required.                                                                                                               | -               | string       | true     |
| `enable_aggregation`                      | Enable the Kubernetes Aggregation Layer.                                                                                                                                                                                                   | true            | bool         | false    |
| `enable_reporting`                        | Enables usage or analytics reporting to upstream.                                                                                                                                                                                          | false           | bool         | false    |
| `pod_cidr`                                | CIDR IPv4 range to assign Kubernetes pods.                                                                                                                                                                                                 | "10.2.0.0/16"   | string       | false    |
| `service_cidr`                            | CIDR IPv4 range to assign Kubernetes services.                                                                                                                                                                                             | "10.3.0.0/16"   | string       | false    |
| `cluster_domain_suffix`                   | Cluster's DNS domain.                                                                                                                                                                                                                      | "cluster.local" | string       | false    |
| `certs_validity_period_hours`             | Validity of all the certificates in hours.                                                                                                                                                                                                 | 8760            | number       | false    |
| `network_mtu`                             | Physical Network MTU.                                                                                                                                                                                                                      | 1500            | number       | false    |
| `disable_self_hosted_kubelet`             | If true, self-hosted kubelet won't be installed on the cluster.                                                                                                                                                                            | false           | bool         | false    |
| `worker_pool`                             | Configuration block for worker pools. There can be more than one.                                                                                                                                                                          | -               | list(object) | true     |
| `worker_pool.ip_addresses`                | List of IP addresses of Tinkerbell hardware to be used for worker pool nodes. With `experimental_sandbox`, machines will be created with these IP addresses.                                                                               | -               | list(string) | true     |
| `worker_pool.ssh_public_keys`             | List of SSH public keys for user `core` on worker pool nodes. Each element must be specified in a valid OpenSSH public key format, as defined in RFC 4253 Section 6.6, e.g. "ssh-rsa AAAAB3N...".                                          | []              | list(string) | false    |
| `worker_pool.os_channel`                  | Flatcar Container Linux channel to install from (stable, beta, alpha, edge).                                                                                                                                                               | "stable"        | string       | false    |
| `worker_pool.os_version`                  | Flatcar Container Linux version to install. Version such as "2303.3.1" or "current".                                                                                                                                                       | "current"       | string       | false    |
| `worker_pool.flatcar_install_base_url`    | URL passed to `flatcar-install` script to fetch Flatcar images from.                                                                                                                                                                       | -               | string       | false    |
| `worker_pool.clc_snippets`                | Flatcar Container Linux Config snippets for nodes in the worker pool.                                                                                                                                                                      | []              | list(string) | false    |
| `worker_pool.labels`                      | Map of extra Kubernetes Node labels for worker nodes.                                                                                                                                                                                      | -               | map(string)  | false    |
| `worker_pool.taints`                      | Map of Taints to assign to worker nodes.                                                                                                                                                                                                   | -               | map(string)  | false    |
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
