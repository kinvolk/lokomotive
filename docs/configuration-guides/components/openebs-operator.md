# OpenEBS operator configuration reference for Lokomotive

## Contents

* [Introduction](#introduction)
* [Prerequisites](#prerequisites)
* [Configuration](#configuration)
* [Argument reference](#argument-reference)
* [Installation](#installation)
* [Uninstallation](#uninstallation)

## Introduction

[OpenEBS](https://docs.openebs.io/) is a container native storage provider, which supports dynamic
storage provisioning, which allows creating persistent volume claims to be automatically bound by
created persistent volumes. OpenEBS adopts Container Attached Storage (CAS) approach, where each
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
node]https://kubernetes.io/docs/tasks/configure-pod-container/assign-pods-nodes/#add-a-label-to-a-node).

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

## Argument reference

Table of all the arguments accepted by the component.

Example:

| Argument             | Description              | Default | Required |
|----------------------|------------------------- |:-------:|:--------:|
| `ndm_selector_label` | Name of the node label.  | -       | false    |
| `ndm_selector_value` | Value of the node label  | -       | false    |

## Installation

To install the OpenEBS operator component:

```bash
lokoctl component install openebs-operator
```

This component only concerns with the installation of openebs-operator. To configure the
storageclass and storage pool claim, check out the [openebs-storage-class](openebs-storage-class.md)
component.

## Uninstallation

To uninstall the component:

```bash
lokoctl component render-manifest openebs-operator | kubectl delete -f -
```
