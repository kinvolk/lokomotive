---
title: Amazon EBS CSI Driver configuration reference for Lokomotive
weight: 10
---

## Introduction

The [CSI Driver for Amazon EBS](https://github.com/kubernetes-sigs/aws-ebs-csi-driver)
provides a CSI interface used by container orchestrators to manage the lifecycle
of Amazon EBS volumes. It provides a storage class for AWS, backed by Amazon EBS
volumes.

## Prerequisites

* A Lokomotive cluster accessible via `kubectl` deployed on Packet.

## Configuration

To run a cluster with the CSI Driver component, `enable_csi` needs
to be set to `true` in the `cluster` block of your lokocfg. The flag and the component
should only be used for clusters deployed on AWS.

Sample config:

```hcl
# aws-ebs-csi-driver.lokocfg
component "aws-ebs-csi-driver" {
    enable_default_storage_class = true
}
```

## Attribute reference

Table of all the arguments accepted by the component.

| Argument                       | Description                                                                   | Default | Type | Required |
|--------------------------------|-------------------------------------------------------------------------------|---------|------|----------|
| `enable_default_storage_class` | Use the storage class provided by the component as the default storage class. | true    | bool | false    |

## Applying

To apply the CSI Driver component, run the following command:

```bash
lokoctl component apply aws-ebs-csi-driver
```
By default, the CSI Driver pods run in the `kube-system` namespace.

## Deleting

To delete the component, run the following command:

```bash
lokoctl component delete aws-ebs-csi-driver
```

**WARNING: Before destroying a cluster or deleting the component, EBS volumes
must be cleaned up manually.** Failing to do so would result in EBS volumes
being left behind.
