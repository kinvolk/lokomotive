# OpenEBS storage class configuration reference for Lokomotive

## Contents

* [Introduction](#introduction)
* [Prerequisites](#prerequisites)
* [Configuration](#configuration)
* [Attribute reference](#attribute-reference)
* [Installing](#installing)
* [Uninstalling](#uninstalling)

## Introduction

OpenEBS has many components, which can be grouped into the following categories.

- **Control plane components** - Provisioner, API Server, volume exports, and volume sidecars.

- **Data plane components** - Jiva and cStor.

- **Node disk manager** - Discover, monitor, and manage the media attached to the Kubernetes node.

- **Integrations with cloud-native tools** - Integrations are done with Prometheus, Grafana, Fluentd, and Jaeger.

According to OpenEBS, [cStor](https://docs.openebs.io/docs/next/cstor.html) is the recommended storage engine in OpenEBS.

This component configures the storage class and storage pool claim for OpenEBS the with cStor storage engine.

## Prerequisites

* Openebs operator installed and in running state.

## Configuration

For a default component configuration, one need not specify a configuration file.

This component supports configuring multiple storage classes and providing disks to use for storage.

OpenEBS storage class component configuration example:

```tf
# openebs-storage-class.lokocfg
component "openebs-storage-class" {
  # Optional arguments
  storage-class "openebs-replica1" {
    replica_count = 1
  }
  storage-class "openebs-replica3" {
    replica_count = 3
    default = true
    disks = [
      "blockdevice-0565dd2d566cab012b7bc35e54874d9f",
      "blockdevice-17901367ccd9e1ead797a7e233de8cc8",
      "blockdevice-1f4315cb4acbb4b0dbf5202adcdb70d8"
    ]
  }
}
```

## Attribute reference

Table of all the arguments accepted by the component.

Example:

| Argument        | Description                                                                                                                   | Default | Required |
|-----------------|-------------------------------------------------------------------------------------------------------------------------------|:-------:|:--------:|
| `replica_count` | Defines the number of cStor volume replicas.                                                                                  | 3       | false    |
| `default`       | Indicates whether the storage class is default or not.                                                                        | false   | false    |
| `disks`         | List of selected unclaimed BlockDevice CRs which are unmounted and do not contain a filesystem in each participating node.    | -       | false    |

## Installing

To install the OpenEBS storage class component:

```bash
lokoctl component apply openebs-storage-class
```
## Uninstalling

To uninstall the component:

```bash
lokoctl component render-manifest openebs-storage-class | kubectl delete -f -
```
