---
title: Rook Ceph configuration reference for Lokomotive
linkTitle: Rook Ceph
weight: 10
---

## Introduction

[Rook](https://rook.io/) is a storage orchestrator for Kubernetes. This component installs a Ceph
cluster managed by the Rook operator.

## Prerequisites

* A Lokomotive cluster accessible via `kubectl`.

* The [Rook](rook.md) component deployed on the cluster.

## Configuration

Rook-Ceph component configuration example:

```tf
component "rook-ceph" {
  # Optional arguments
  namespace       = "rook-test"
  monitor_count   = 3
  enable_toolbox  = true
  metadata_device = "md127"
  node_affinity {
    key      = "node-role.kubernetes.io/storage"
    operator = "Exists"
  }
  node_affinity {
    key      = "storage.lokomotive.io"
    operator = "In"

    # If the `operator` is set to `"In"`, `values` should be specified.
    values = [
      "foo",
    ]
  }
  toleration {
    key      = "storage.lokomotive.io"
    operator = "Equal"
    value    = "rook-ceph"
    effect   = "NoSchedule"
  }

  storage_class {
    enable         = true
    default        = true
    reclaim_policy = "Delete"
  }
}

```

The Ceph cluster needs to be deployed in the same namespace as the Rook operator at the moment.
Additional `Roles` and `RoleBindings` need to be created if deploying across separate namespaces is
desired.

## Attribute reference

Table of all the arguments accepted by the component.

| Argument                       | Description                                                                                                                                                                                                                                                                                                                 | Default  | Type                                                                                                           | Required |
|--------------------------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|:--------:|:---------------------------------------------------------------------------------------------------------------|:--------:|
| `namespace`                    | Namespace to deploy the Ceph cluster into. Must be the same as the rook operator.                                                                                                                                                                                                                                           |  "rook"  | string                                                                                                         |  false   |
| `monitor_count`                | Number of Ceph monitors to deploy. An odd number like 3 or 5 is recommended which should also be sufficient for most cases.                                                                                                                                                                                                 |     1    | number                                                                                                         |  false   |
| `enable_toolbox`               | Deploy the [toolbox pod](https://rook.io/docs/rook/master/ceph-toolbox.html) to debug and manage the Ceph cluster.                                                                                                                                                                                                          |  false   | bool                                                                                                           |  false   |
| `node_affinity`                | Node affinity for deploying the Ceph cluster pods.                                                                                                                                                                                                                                                                          |     -    | list(object({key = string, operator = string, values = list(string)}))                                         |  false   |
| `toleration`                   | Tolerations that the Ceph cluster pods will tolerate.                                                                                                                                                                                                                                                                       |     -    | list(object({key = string, effect = string, operator = string, value = string, toleration_seconds = string })) |  false   |
| `metadata_device`              | Name of the device to store the metadata on each storage machine. **Note**: Provide just the name of the device and skip prefixing with `/dev/`.                                                                                                                                                                            |     -    | string                                                                                                         |  false   |
| `storage_class.enable`         | Install Storage Class config.                                                                                                                                                                                                                                                                                               |   false  | bool                                                                                                           |  false   |
| `storage_class.default`        | Make this Storage Class as a default one.                                                                                                                                                                                                                                                                                   |   false  | bool                                                                                                           |  false   |
| `storage_class.reclaim_policy` | Persistent volumes created with this storage class will have this reclaim policy. This field decides what happens to the volume after a user deletes a PVC. Valid values: `Retain`, `Recycle` and `Delete`. Read more in the [Kubernetes docs](https://kubernetes.io/docs/concepts/storage/persistent-volumes/#reclaiming). | `Retain` | string                                                                                                         |  false   |

## Applying

To apply the Rook-Ceph component:

```bash
lokoctl component apply rook-ceph
```

Once the Ceph cluster is ready, an object store can be deployed to start writing to Ceph.
More information is available at [configuring Ceph object store CRD](https://rook.io/docs/rook/v1.2/ceph-object-store-crd.html)

## Deleting

To destroy the component:

```bash
lokoctl component delete rook-ceph
```

After removing the component from the cluster, make sure to delete `/var/lib/rook` from the host
filesystem of all worker nodes for a clean reinstallation.
