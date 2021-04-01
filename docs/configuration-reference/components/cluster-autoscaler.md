---
title: Cluster Autoscaler configuration reference for Lokomotive
linkTitle: Cluster Autoscaler
weight: 10
---

## Introduction

The [Cluster Autoscaler](https://github.com/kubernetes/autoscaler/tree/master/cluster-autoscaler) is
a tool that automatically adjusts the size of a Kubernetes cluster when there are pods that failed
to run in the cluster due to insufficient resources or when there are nodes that have been
underutilized for an extended period of time and pods can be placed on other existing nodes.

## Prerequisites

* A Lokomotive cluster accessible via `kubectl` deployed on Packet.

* For existing worker nodes in the cluster, you need to tag them manually for the Cluster Autoscaler
  to consider.

  ```bash
  k8s-cluster-${cluster_name}
  k8s-nodepool-${worker_pool}
  ```

## Configuration

Currently Lokomotive supports the Cluster Autoscaler component only on the Packet platform. Support
for other platforms will be added in the future.

Cluster Autoscaler component configuration example:

```tf
# cluster-autoscaler.lokocfg
component "cluster-autoscaler" {

  # Required arguments
  cluster_name = "cluster"
  worker_pool = "pool"

  # Optional arguments
  provider = "packet"
  namespace = "kube-system"
  min_workers = 1
  max_workers = 4
  scale_down_unneeded_time = "10m"
  scale_down_delay_after_add = "10m"
  scale_down_unready_time = "20m"

  packet {
    # Required arguments for Packet platform
    project_id = "0b5dc3d2-949a-447d-82cd-43fbdc1ae8c0"
    facility = "sjc1"

    # Optional arguments
    worker_type = "c3.small.x86"
    worker_channel = "stable"
  }
}
```

## Attribute reference

Table of all the arguments accepted by the component.

| Argument                     | Description                                                                              | Default        |  Type  | Required |
|------------------------------|------------------------------------------------------------------------------------------|:---------------|:------:|:--------:|
| `cluster_name`               | Name of the  cluster.                                                                    | -              | string |   true   |
| `worker_pool`                | Name of the worker pool.                                                                 | -              | string |   true   |
| `namespace`                  | Namespace where the Cluster Autoscaler will be installed.                                | "kube-system"  | string |  false   |
| `min_workers`                | Minimum number of workers in the worker pool.                                            | 1              | number |  false   |
| `max_workers`                | Maximum number of workers in the worker pool.                                            | 4              | number |  false   |
| `scale_down_unneeded_time`   | How long a node should be unneeded before it is eligible for scale down.                 | "10m"          | string |  false   |
| `scale_down_delay_after_add` | How long scale down should wait after a scale up.                                        | "10m"          | string |  false   |
| `scale_down_unready_time`    | How long an unready node should be unneeded before it is eligible for scale down.        | "20m"          | string |  false   |
| `provider`                   | Supported provider, currently Packet.                                                    | "packet"       | string |  false   |
| `service_monitor`            | Specifies how metrics can be retrieved from a set of services.                           | false          |  bool  |  false   |
| `packet.project_id`          | Packet Project ID where the cluster is running.                                          | -              | string |   true   |
| `packet.facility`            | Packet Facility where the cluster is running.                                            | -              | string |   true   |
| `packet.worker_type`         | Machine type for workers spawned by the Cluster Autoscaler.                              | "c3.small.x86" | string |  false   |
| `packet_worker_channel`      | Flatcar Container Linux channel to be used in workers spawned by the Cluster Autoscaler. | "stable"       | string |  false   |


## Applying

To apply the cluster autoscaler component:

```bash
lokoctl component apply cluster-autoscaler
```
By default, the cluster Autoscaler pods run in the `kube-system` namespace

## Deleting

To destroy the component:

```bash
lokoctl component delete cluster-autoscaler
```
