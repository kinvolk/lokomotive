# Rook Ceph configuration reference for Lokomotive

## Contents

* [Introduction](#introduction)
* [Prerequisites](#prerequisites)
* [Configuration](#configuration)
* [Attribute reference](#attribute-reference)
* [Applying](#applying)
* [Destroying](#destroying)

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
  namespace = "rook-test"
  monitor_count = 3
  metadata_device = "md127"
  node_selector {
    key      = "node-role.kubernetes.io/storage"
    operator = "Exists"
  }
  node_selector {
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
}
```

The Ceph cluster needs to be deployed in the same namespace as the Rook operator at the moment.
Additional `Roles` and `RoleBindings` need to be created if deploying across separate namespaces is
desired.

## Attribute reference

Table of all the arguments accepted by the component.

Example:

| Argument            | Description                                                                                                                                        | Default | Required |
|---------------------|----------------------------------------------------------------------------------------------------------------------------------------------------|:-------:|:--------:|
| `namespace`         | Namespace to deploy the Ceph cluster into. Must be the same as the rook operator.                                                                  | rook    | false    |
| `monitor_count`     | Number of Ceph monitors to deploy. An odd number like 3 or 5 is recommended which should also be sufficient for most cases.                        | 1       | false    |
| `node_selector`     | Node selectors for deploying the Ceph cluster pods.                                                                                                | -       | false    |
| `toleration`        | Tolerations that the Ceph cluster pods will tolerate.                                                                                              | -       | false    |
| `metadata_device`   | Name of the device to store the metadata on each storage machine. **Note**: Provide just the name of the device and skip prefixing with `/dev/`.   | -       | false    |

## Applying

To install the Rook-Ceph component:

```bash
lokoctl component apply rook-ceph
```

Once the Ceph cluster is ready, an object store can be deployed to start writing to Ceph.
More information is available at [configuring Ceph object store CRD](https://rook.io/docs/rook/v1.2/ceph-object-store-crd.html)

## Destroying

To destroy the component:

```bash
lokoctl component render-manifest rook-ceph | kubectl delete -f -
```

After removing the component from the cluster, make sure to delete `/var/lib/rook` from the host
filesystem of all worker nodes for a clean reinstallation.
