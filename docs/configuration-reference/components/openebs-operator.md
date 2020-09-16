---
title: OpenEBS operator configuration reference for Lokomotive
weight: 10
---

## Introduction

[OpenEBS](https://docs.openebs.io/) is a container native storage provider supporting dynamic
storage provisioning, which allows creating persistent volume claims to be automatically bound by
created persistent volumes. OpenEBS a adopts Container Attached Storage (CAS) approach, where each
workload is provided with a dedicated storage controller.

This component installs the OpenEBS operator.

## Prerequisites

* A Lokomotive cluster accessible via `kubectl`.

* At least 3 workers with available disks.

* iSCSI client configured and `iscsid` service running on the worker nodes. In our current setup, we
  have `iscsid` service automatically enabled and running on all worker nodes.

**NOTE:** OpenEBS requires available disks, i.e. disks that aren't mounted by anything. This means
that by default, OpenEBS does not work on machines with just a single physical disk, e.g. Packet's
t1.small.x86 (because the disk is used for the operating system).

## Configuration

To only use a subset of worker nodes for OpenEBS storage, you must manually label the nodes before
configuring the OpenEBS operator component.

Refer to the Kubernetes documentation on [adding labels to a
node](https://kubernetes.io/docs/tasks/configure-pod-container/assign-pods-nodes/#add-a-label-to-a-node).

OpenEBS operator component configuration example:

```tf
component "openebs-operator" {
  # Optional arguments.
  # Example node labels to consider for OpenEBS storage.
  ndm_selector_label = "node"
  ndm_selector_value = "openebs"
}
```

**NOTE**: If `ndm_selector_label` and `ndm_selector_value` are not provided, all worker nodes are
considered by OpenEBS for storage.

## Attribute reference

Table of all the arguments accepted by the component.

| Argument             | Description             | Default |  Type  | Required |
|----------------------|-------------------------|:-------:|:------:|:--------:|
| `ndm_selector_label` | Name of the node label. |    -    | string |  false   |
| `ndm_selector_value` | Value of the node label |    -    | string |  false   |


## Applying

To apply the OpenEBS operator component:

```bash
lokoctl component apply openebs-operator
```

This component only concerns with the installation of openebs-operator. To configure the storage
class and storage pool claim, check out the [openebs-storage-class](../openebs-storage-class)
component.

## Deleting

To destroy the component:

```bash
lokoctl component delete openebs-operator --delete-namespace
```
