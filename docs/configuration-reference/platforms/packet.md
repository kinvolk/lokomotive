# Lokomotive Packet configuration reference

## Contents

* [Introduction](#introduction)
* [Prerequisites](#prerequisites)
* [Configuration](#configuration)
* [Attribute reference](#attribute-reference)
* [Applying](#applying)
* [Destroying](#destroying)
* [ARM support and hybrid clusters](#arm-support-and-hybrid-clusters)

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
variable "workers_type" {}
variable "dns_zone" {}
variable "route53_zone_id" {}
variable "packet_project_id" {}
variable "ssh_public_keys" {}
variable "management_cidrs" {}
variable "node_private_cidr" {}
variable "state_s3_bucket" {}
variable "lock_dynamodb_table" {}

backend "s3" {
  bucket         = var.state_s3_bucket
  key            = var.state_s3_key
  region         = var.state_s3_region
  dynamodb_table = var.lock_dynamodb_table
}

# backed "local" {
#   path = "path/to/local/file"
#}

cluster "packet" {
  auth_token = var.packet_token

  asset_dir = var.asset_dir

  cluster_name = var.cluster_name

  controller_count = var.controllers_count

  controller_type = "baremetal_0"

  facility = var.facility

  os_channel = "stable"

  os_version = "current"

  os_arch = "amd64"

  ipxe_script_url = ""

  project_id = var.packet_project_id

  dns {
    zone = var.dns_zone

    provider {
      route53 {
        zone_id = var.route53_zone_id
      }
    }

    # manual {}
  }

  ssh_pubkeys = var.ssh_public_keys

  management_cidrs = var.management_cidrs

  node_private_cidr = var.node_private_cidr

  cluster_domain_suffix = "cluster.local"

  network_mtu = 1480

  tags {
    key1 = "value1"
    key2 = "value2"
  }

  enable_aggregation = true

  enable_reporting = false

  network_ip_autodetection_method = "first-found"

  pod_cidr = "10.2.0.0/16"

  service_cidr = "10.3.0.0/16"

  reservation_ids = { controller-0 = "55555f20-a1fb-55bd-1e11-11af11d11111" }

  reservation_ids_default = ""

  certs_validity_period_hours = 8760

  worker_pool "worker-pool-1" {
    count = var.workers_count

    ipxe_script_url = ""

    os_arch = "amd64"

    disable_bgp = false

    node_type = var.workers_type

    os_channel = "stable"

    os_version = "current"

    labels = "foo=bar,baz=zab"

    taints = "nodeType=storage:NoSchedule"

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

The default for node_type is `baremetal_0`. If you wish to change the default, then you
define the variable  and use it to refer in the cluster configuration.

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

| Argument                              | Description                                                                                                                                                                   | Default         | Required |
|---------------------------------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|:---------------:|:--------:|
| `auth_token`                          | Packet Auth token. Use the `PACKET_AUTH_TOKEN` environment variable instead.                                                                                                  | -               | false    |
| `asset_dir`                           | Location where Lokomotive stores cluster assets.                                                                                                                              | -               | true     |
| `cluster_name`                        | Name of the cluster.                                                                                                                                                          | -               | true     |
| `tags`                                | List of tags that will be propagated to master nodes.                                                                                                                         | -               | false    |
| `controller_count`                    | Number of controller nodes.                                                                                                                                                   | 1               | false    |
| `controller_type`                     | Packet instance type for controllers.                                                                                                                                         | "baremetal_0"   | false    |
| `dns`                                 | DNS configuration block.                                                                                                                                                      | -               | true     |
| `dns.zone`                            | DNS Zone.                                                                                                                                                                     | -               | true     |
| `dns.provider`                        | DNS Provider configuration block. Route 53 or Manual.                                                                                                                         | -               | true     |
| `dns.provider.route53`                | Route 53 DNS Configuration.                                                                                                                                                   | -               | false    |
| `dns.provider.route53.zone_id`        | Route 53 DNS Zone ID.                                                                                                                                                         | -               | true     |
| `dns.provider.route53.aws_creds_path` | AWS credentials for managing Route 53 DNS.                                                                                                                                    | -               | false    |
| `dns.provider.manual`                 | Manual DNS configuration.                                                                                                                                                     | -               | false    |
| `facility`                            | Packet facility to use for deploying the cluster.                                                                                                                             | -               | false    |
| `project_id`                          | Packet project ID.                                                                                                                                                            | -               | true     |
| `ssh_pubkeys`                         | SSH public keys for user `core`.                                                                                                                                              | -               | true     |
| `os_arch`                             | Flatcar Container Linux architecture to install (amd64, arm64).                                                                                                               | "amd64"         | false    |
| `os_channel`                          | Flatcar Container Linux channel to install from (stable, beta, alpha, edge).                                                                                                  | "stable"        | false    |
| `os_version`                          | Flatcar Container Linux version to install. Version such as "2303.3.1" or "current".                                                                                          | "current"       | false    |
| `ipxe_script_url`                     | Boot via iPXE. Required for arm64.                                                                                                                                            | -               | false    |
| `management_cidrs`                    | List of IPv4 CIDRs authorized to access or manage the cluster. Example ["0.0.0.0/0"] to allow all.                                                                            | -               | true     |
| `node_private_cidr`                   | Private IPv4 CIDR of the nodes used to allow inter-node traffic. Example "10.0.0.0/8"                                                                                         | -               | true     |
| `enable_aggregation`                  | Enable the Kubernetes Aggregation Layer.                                                                                                                                      | true            | false    |
| `network_mtu`                         | CNI interface MTU                                                                                                                                                             | 1480            | false    |
| `pod_cidr`                            | CIDR IPv4 range to assign Kubernetes pods.                                                                                                                                    | "10.2.0.0/16"   | false    |
| `service_cidr`                        | CIDR IPv4 range to assign Kubernetes services.                                                                                                                                | "10.3.0.0/16"   | false    |
| `cluster_domain_suffix`               | Cluster's DNS domain.                                                                                                                                                         | "cluster.local" | false    |
| `enable_reporting`                    | Enables usage or analytics reporting to upstream.                                                                                                                             | false           | false    |
| `reservation_ids`                     | Specify Packet hardware reservation ID for instances.                                                                                                                         | -               | false    |
| `reservation_ids_default`             | Default reservation ID for nodes not listed in the `reservation_ids`. The value`next-available` will choose any reservation that matches the pool's device type and facility. | ""              | false    |
| `certs_validity_period_hours`         | Validity of all the certificates in hours.                                                                                                                                    | 8760            | false    |
| `worker_pool`                         | Configuration block for worker pools. There can be more than one.                                                                                                             | -               | true     |
| `worker_pool.count`                   | Number of workers in the worker pool. Can be changed afterwards to add or delete workers.                                                                                     | 1               | true     |
| `worker_pool.disable_bgp`             | Disable BGP on nodes. Nodes won't be able to connect to Packet BGP peers.                                                                                                     | false           | false    |
| `worker_pool.ipxe_script_url`         | Boot via iPXE. Required for arm64.                                                                                                                                            | -               | false    |
| `worker_pool.os_arch`                 | Flatcar Container Linux architecture to install (amd64, arm64).                                                                                                               | "amd64"         | false    |
| `worker_pool.os_channel`              | Flatcar Container Linux channel to install from (stable, beta, alpha, edge).                                                                                                  | "stable"        | false    |
| `worker_pool.os_version`              | Flatcar Container Linux version to install. Version such as "2303.3.1" or "current".                                                                                          | "current"       | false    |
| `worker_pool.node_type`               | Packet instance type for worker nodes.                                                                                                                                        | "baremetal_0"   | false    |
| `worker_pool.labels`                  | Custom labels to assign to worker nodes.                                                                                                                                      | -               | false    |
| `worker_pool.taints`                  | Taints to assign to worker nodes.                                                                                                                                             | -               | false    |
| `worker_pool.setup_raid`              | Attempt to create a RAID 0 from extra disks to be used for persistent container storage. Can't be used with `setup_raid_hdd` nor `setup_raid_sdd`.                            | false           | false    |
| `worker_pool.setup_raid_hdd`          | Attempt to create a RAID 0 from extra Hard Disk drives only, to be used for persistent container storage. Can't be used with `setup_raid` nor `setup_raid_sdd`.               | false           | false    |
| `worker_pool.setup_raid_ssd`          | Attempt to create a RAID 0 from extra Solid State Drives only, to be used for persistent container storage.  Can't be used with `setup_raid` nor `setup_raid_hdd`.            | false           | false    |
| `worker_pool.setup_raid_ssd_fs`       | When set to `true` file system will be created on SSD RAID device and will be mounted on `/mnt/node-local-ssd-storage`. To use the raw device set it to `false`.              | false           | false    |

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
ipxe_script_url = "https://raw.githubusercontent.com/kinvolk/flatcar-ipxe-scripts/arm64-usr/packet.ipxe"
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
