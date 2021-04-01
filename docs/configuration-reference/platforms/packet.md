---
title: Lokomotive Packet configuration reference
linkTitle: Packet
weight: 10
---

## Introduction

This configuration reference provides information on configuring a Lokomotive cluster on Packet with all
the configuration options available to the user.

## Prerequisites

* `lokoctl` [installed locally.](../../installer/lokoctl.md)
* `kubectl` installed locally to access the Kubernetes cluster.

### Configuration

To create a Lokomotive cluster, we need to define a configuration.

Example configuration file:

```tf
# mycluster.lokocfg
variable "packet_token" {}
variable "asset_dir" {}
variable "facility" {}
variable "cluster_name" {}
variable "controllers_count" {}
variable "workers_count" {}
variable "controller_type" {}
variable "controller_clc_snippets" {}
variable "workers_type" {}
variable "dns_zone" {}
variable "route53_zone_id" {}
variable "packet_project_id" {}
variable "ssh_public_keys" {}
variable "management_cidrs" {}
variable "node_private_cidrs" {}
variable "state_s3_bucket" {}
variable "lock_dynamodb_table" {}
variable "oidc_issuer_url" {}
variable "oidc_client_id" {}
variable "oidc_username_claim" {}
variable "oidc_groups_claim" {}
variable "worker_clc_snippets" {}
variable "worker_pool_facility" {}

backend "s3" {
  bucket         = var.state_s3_bucket
  key            = var.state_s3_key
  region         = var.state_s3_region
  dynamodb_table = var.lock_dynamodb_table
}

# backend "local" {
#   path = "path/to/local/file"
#}

cluster "packet" {
  auth_token = var.packet_token

  asset_dir = var.asset_dir

  cluster_name = var.cluster_name

  controller_count = var.controllers_count

  controller_type = "c3.small.x86"

  controller_clc_snippets = var.controller_clc.snippets

  facility = var.facility

  os_channel = "stable"

  os_version = "current"

  os_arch = "amd64"

  ipxe_script_url = ""

  project_id = var.packet_project_id

  dns {
    zone     = var.dns_zone
    provider = "route53"
  }

  ssh_pubkeys = var.ssh_public_keys

  management_cidrs = var.management_cidrs

  node_private_cidrs = var.node_private_cidrs

  cluster_domain_suffix = "cluster.local"

  network_mtu = 1500

  tags {
    key1 = "value1"
    key2 = "value2"
  }

  enable_aggregation = true

  enable_tls_bootstrap = true

  encrypt_pod_traffic = true

  enable_reporting = false

  network_ip_autodetection_method = "first-found"

  pod_cidr = "10.2.0.0/16"

  service_cidr = "10.3.0.0/16"

  reservation_ids = { controller-0 = "55555f20-a1fb-55bd-1e11-11af11d11111" }

  reservation_ids_default = ""

  certs_validity_period_hours = 8760

  conntrack_max_per_core = 32768

  oidc {
    issuer_url     = var.oidc_issuer_url
    client_id      = var.oidc_client_id
    username_claim = var.oidc_username_claim
    groups_claim   = var.oidc_groups_claim
  }

  worker_pool "worker-pool-1" {
    count = var.workers_count

    clc_snippets = var.worker_clc_snippets

    tags = {
      pool = "storage"
    }

    ipxe_script_url = ""

    os_arch = "amd64"

    disable_bgp = false

    facility = var.worker_pool_facility

    node_type = var.workers_type

    os_channel = "stable"

    os_version = "current"

    labels = {
      "testlabel" = ""
    }

    taints = {
      "nodeType" = "storage:NoSchedule"
    }

    setup_raid = false

    setup_raid_hdd = false

    setup_raid_ssd = false

    setup_raid_ssd_fs = false
  }
}
```

**NOTE**: Should you feel differently about the default values, you can set default values using the `variable`
block in the cluster configuration.

Example:

The default for node_type is `c3.small.x86`. If you wish to change the default, then you
define the variable and use it to refer in the cluster configuration.

```tf
variable "custom_default_worker_type" {
  default = "c2.medium.x86"
}

.
.
.
node_type = var.custom_default_worker_type
.
.
.

```

## Attribute reference

| Argument                              | Description                                                                                                                                                                                                                                                                                                                        | Default         | Type         | Required |
|---------------------------------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|-----------------|--------------|----------|
| `auth_token`                          | Packet Project API token. Can be supplied using `PACKET_AUTH_TOKEN` environment variable.                                                                                                                                                                                                                                          | -               | string       | false    |
| `asset_dir`                           | Location where Lokomotive stores cluster assets.                                                                                                                                                                                                                                                                                   | -               | string       | true     |
| `cluster_name`                        | Name of the cluster.                                                                                                                                                                                                                                                                                                               | -               | string       | true     |
| `tags`                                | List of tags that will be propagated to master nodes.                                                                                                                                                                                                                                                                              | -               | map(string)  | false    |
| `controller_count`                    | Number of controller nodes.                                                                                                                                                                                                                                                                                                        | 1               | number       | false    |
| `controller_type`                     | Packet instance type for controllers. Changing this field on existing cluster will be ignored. To actually apply it, cluster needs to be destroyed and re-created.                                                                                                                                                                 | "c3.small.x86"  | string       | false    |
| `controller_clc_snippets`             | Controller Flatcar Container Linux Config snippets.                                                                                                                                                                                                                                                                                | []              | list(string) | false    |
| `dns`                                 | DNS configuration block.                                                                                                                                                                                                                                                                                                           | -               | object       | true     |
| `dns.zone`                            | A DNS zone to use for the cluster. The following format is used for cluster-related DNS records: `<record>.<cluster_name>.<dns_zone>`                                                                                                                                                                                              | -               | string       | true     |
| `dns.provider`                        | DNS provider to use for the cluster. Valid values: `cloudflare`, `route53`, `manual`.                                                                                                                                                                                                                                              | -               | string       | true     |
| `oidc`                                | OIDC configuration block.                                                                                                                                                                                                                                                                                                          | -               | object       | false    |
| `oidc.issuer_url`                     | URL of the provider which allows the API server to discover public signing keys. Only URLs which use the https:// scheme are accepted.                                                                                                                                                                                             | -               | string       | false    |
| `oidc.client_id`                      | A client id that all tokens must be issued for.                                                                                                                                                                                                                                                                                    | "clusterauth"   | string       | false    |
| `oidc.username_claim`                 | JWT claim to use as the user name.                                                                                                                                                                                                                                                                                                 | "email"         | string       | false    |
| `oidc.groups_claim`                   | JWT claim to use as the user’s group.                                                                                                                                                                                                                                                                                              | "groups"        | string       | false    |
| `facility`                            | Packet facility to use for deploying the cluster.                                                                                                                                                                                                                                                                                  | -               | string       | false    |
| `project_id`                          | Packet project ID.                                                                                                                                                                                                                                                                                                                 | -               | string       | true     |
| `ssh_pubkeys`                         | List of SSH public keys for user `core`. Each element must be specified in a valid OpenSSH public key format, as defined in RFC 4253 Section 6.6, e.g. "ssh-rsa AAAAB3N...".                                                                                                                                                       | -               | list(string) | true     |
| `os_arch`                             | Flatcar Container Linux architecture to install (amd64, arm64).                                                                                                                                                                                                                                                                    | "amd64"         | string       | false    |
| `os_channel`                          | Flatcar Container Linux channel to install from (stable, beta, alpha, edge).                                                                                                                                                                                                                                                       | "stable"        | string       | false    |
| `os_version`                          | Flatcar Container Linux version to install. Version such as "2303.3.1" or "current".                                                                                                                                                                                                                                               | "current"       | string       | false    |
| `ipxe_script_url`                     | Boot via iPXE. Required for arm64.                                                                                                                                                                                                                                                                                                 | -               | string       | false    |
| `management_cidrs`                    | List of IPv4 CIDRs authorized to access or manage the cluster. Example ["0.0.0.0/0"] to allow all.                                                                                                                                                                                                                                 | -               | list(string) | true     |
| `node_private_cidr`                   | (Deprecated, use `node_private_cidrs` instead) Private IPv4 CIDR of the nodes used to allow inter-node traffic. Example "10.0.0.0/8".                                                                                                                                                                                              | -               | string       | true     |
| `node_private_cidrs`                  | List of Private IPv4 CIDRs of the nodes used to allow inter-node traffic. Example ["10.0.0.0/8"].                                                                                                                                                                                                                                  | -               | list(string) | true     |
| `enable_aggregation`                  | Enable the Kubernetes Aggregation Layer.                                                                                                                                                                                                                                                                                           | true            | bool         | false    |
| `enable_tls_bootstrap`                | Enable TLS bootstraping for Kubelet.                                                                                                                                                                                                                                                                                               | true            | bool         | false    |
| `encrypt_pod_traffic`                 | Enable in-cluster pod traffic encryption. If true `network_mtu` is reduced by 60 to make room for the encryption header.                                                                                                                                                                                                           | false           | bool         | false    |
| `ignore_x509_cn_check`                | Ignore check of common name in x509 certificates. If any application is built pre golang 1.15 then API server rejects x509 from such application, enable this to get around apiserver.                                                                                                                                             | false           | bool         | false    |
| `network_mtu`                         | Physical Network MTU.                                                                                                                                                                                                                                                                                                              | 1500            | number       | false    |
| `pod_cidr`                            | CIDR IPv4 range to assign Kubernetes pods.                                                                                                                                                                                                                                                                                         | "10.2.0.0/16"   | string       | false    |
| `service_cidr`                        | CIDR IPv4 range to assign Kubernetes services.                                                                                                                                                                                                                                                                                     | "10.3.0.0/16"   | string       | false    |
| `cluster_domain_suffix`               | Cluster's DNS domain.                                                                                                                                                                                                                                                                                                              | "cluster.local" | string       | false    |
| `enable_reporting`                    | Enables usage or analytics reporting to upstream.                                                                                                                                                                                                                                                                                  | false           | bool         | false    |
| `reservation_ids`                     | Block with Packet hardware reservation IDs for controller nodes. Each key must have the format `controller-${index}` and the value is the reservation UUID. Can't be combined with `reservation_ids_default`. Key indexes must be sequential and start from 0. Example: `reservation_ids = { controller-0 = "<reservation_id>" }`. | -               | map(string)  | false    |
| `reservation_ids_default`             | Default reservation ID for controllers. The value`next-available` will choose any reservation that matches the pool's device type and facility. Can't be combined with `reservation_ids`                                                                                                                                           | -               | string       | false    |
| `certs_validity_period_hours`         | Validity of all the certificates in hours.                                                                                                                                                                                                                                                                                         | 8760            | number       | false    |
| `conntrack_max_per_core`      				| Maximum number of entries in conntrack table per CPU on all nodes in the cluster. If you require more fain-grained control over this value, set it to 0 and add CLC snippet setting `net.netfilter.nf_conntrack_max sysctl setting per node pool. See [Flatcar documentation about sysctl] for more details.                       | 32768           | number       | false    |
| `worker_pool`                         | Configuration block for worker pools. There can be more than one.                                                                                                                                                                                                                                                                  | -               | list(object) | true     |
| `worker_pool.count`                   | Number of workers in the worker pool. Can be changed afterwards to add or delete workers.                                                                                                                                                                                                                                          | 1               | number       | true     |
| `worker_pool.clc_snippets`            | Flatcar Container Linux Config snippets for nodes in the worker pool.                                                                                                                                                                                                                                                              | []              | list(string) | false    |
| `worker_pool.tags`                    | List of tags that will be propagated to nodes in the worker pool.                                                                                                                                                                                                                                                                  | -               | map(string)  | false    |
| `worker_pool.disable_bgp`             | Disable BGP on nodes. Nodes won't be able to connect to Packet BGP peers.                                                                                                                                                                                                                                                          | false           | bool         | false    |
| `worker_pool.ipxe_script_url`         | Boot via iPXE. Required for arm64.                                                                                                                                                                                                                                                                                                 | -               | string       | false    |
| `worker_pool.os_arch`                 | Flatcar Container Linux architecture to install (amd64, arm64).                                                                                                                                                                                                                                                                    | "amd64"         | string       | false    |
| `worker_pool.os_channel`              | Flatcar Container Linux channel to install from (stable, beta, alpha, edge).                                                                                                                                                                                                                                                       | "stable"        | string       | false    |
| `worker_pool.os_version`              | Flatcar Container Linux version to install. Version such as "2303.3.1" or "current".                                                                                                                                                                                                                                               | "current"       | string       | false    |
| `worker_pool.node_type`               | Packet instance type for worker nodes.                                                                                                                                                                                                                                                                                             | "c3.small.x86"  | string       | false    |
| `worker_pool.facility`                | Packet facility to use for deploying the worker pool. Enable ["Backend Transfer"](https://metal.equinix.com/developers/docs/networking/features/#backend-transfer) on the Equinix Metal project for this to work.                                                                                                        | Same as controller nodes. | string       | false    |
| `worker_pool.labels`                  | Map of extra Kubernetes Node labels for worker nodes.                                                                                                                                                                                                                                                                              | -               | map(string)  | false    |
| `worker_pool.taints`                  | Map of Taints to assign to worker nodes.                                                                                                                                                                                                                                                                                           | -               | map(string)  | false    |
| `worker_pool.reservation_ids`         | Block with Packet hardware reservation IDs for worker nodes. Each key must have the format `worker-${index}` and the value is the reservation UUID. Can't be combined with `reservation_ids_default`. Key indexes must be sequential and start from 0. Example: `reservation_ids = { worker-0 = "<reservation_id>" }`.             | -               | map(string)  | false    |
| `worker_pool.reservation_ids_default` | Default reservation ID for workers. The value`next-available` will choose any reservation that matches the pool's device type and facility. Can't be combined with `reservation_ids`.                                                                                                                                              | -               | string       | false    |
| `worker_pool.setup_raid`              | Attempt to create a RAID 0 from extra disks to be used for persistent container storage. Can't be used with `setup_raid_hdd` nor `setup_raid_sdd`.                                                                                                                                                                                 | false           | bool         | false    |
| `worker_pool.setup_raid_hdd`          | Attempt to create a RAID 0 from extra Hard Disk drives only, to be used for persistent container storage. Can't be used with `setup_raid` nor `setup_raid_sdd`.                                                                                                                                                                    | false           | bool         | false    |
| `worker_pool.setup_raid_ssd`          | Attempt to create a RAID 0 from extra Solid State Drives only, to be used for persistent container storage.  Can't be used with `setup_raid` nor `setup_raid_hdd`.                                                                                                                                                                 | false           | bool         | false    |
| `worker_pool.setup_raid_ssd_fs`       | When set to `true` file system will be created on SSD RAID device and will be mounted on `/mnt/node-local-ssd-storage`. To use the raw device set it to `false`.                                                                                                                                                                   | false           | bool         | false    |

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

## ARM support and hybrid clusters

Lokomotive and Flatcar currently include an alpha-quality preview for the
Packet arm64 server type `c2.large.arm`. They can be used for both worker and
controller nodes. Besides specifying them in `controller_type`/`node_type` you
need to configure some additional variables in the respective controller/worker
module:

```
os_arch = "arm64"
os_channel = "alpha"
ipxe_script_url = "https://alpha.release.flatcar-linux.net/arm64-usr/current/flatcar_production_packet.ipxe"
```

The iPXE boot variable can be removed once Flatcar is available for
installation in the Packet OS menu for the ARM servers.

If you have a hybrid cluster with both x86 and ARM nodes, you need to either
use Docker multiarch images such as the standard `debian:latest` or
`python:latest` images, or restrict Pods to nodes of the correct architecture
with an entry like this for ARM (or with `amd64` for x86) in your YAML
deployment:

```
nodeSelector:
  kubernetes.io/arch: arm64
```

An example on how to build multiarch images yourself is
[here](https://github.com/kinvolk/calico-hostendpoint-controller/#building).

**Note**: [Lokomotive Components](../../concepts/components.md) are not
supported in this preview, so if you need them you'll have to install them
manually using Kubernetes directly.
