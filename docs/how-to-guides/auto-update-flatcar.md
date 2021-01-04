---
title: Auto-updating Flatcar Container Linux
weight: 10
---

## Introduction

At the moment, Lokomotive supports only [Flatcar Container Linux](https://www.flatcar-linux.org/) as
the underlying operating system for nodes. While Flatcar can keep itself up to date, when running
Kubernetes on top of it, we recommend using the [Flatcar Linux Update Operator
(FLUO)](https://github.com/kinvolk/flatcar-linux-update-operator) to perform updates to avoid
rebooting too many nodes at a time which could cause a service outage.

FLUO is a supported Lokomotive component.

This guide will show how to install it and disable its
auto-update feature for specific nodes, which might be useful when you run services, which require
special care before shutting down (e.g. storage clusters).

> **NOTE:** If you want to use the FLUO component on a non-Lokomotive cluster, make sure your
> Flatcar nodes have `locksmithd.service` systemd unit disabled to avoid nodes rebooting on their
> own. On Lokomotive, it is disabled by default.

## Prerequisites

- A Lokomotive cluster accessible via `kubectl`.

## Steps

### Step 1: Disable auto-update for sensitive nodes

If you want to update the particular nodes manually in a controlled fashion, there is a way to
disable automatic updates. Disabling updates can come in handy when the workloads run by these
machines are storage or ingress network related, where applications cannot tolerate the abrupt
reboot of node.

Please annotate the nodes as follows:

```bash
kubectl annotate node <node name> "flatcar-linux-update.v1.flatcar-linux.net/reboot-paused=true"
```

### Step 2: Configure FLUO

Add the following content to your cluster configuration (e.g. in `fluo.lokocfg` file):

```tf
component "flatcar-linux-update-operator" {}
```

### Step 3: Install FLUO

Execute the following command to deploy the `flatcar-linux-update-operator` component:

```bash
lokoctl component apply flatcar-linux-update-operator
```

Verify that pods in the `reboot-coordinator` namespace are running (this may take a
few minutes):

```bash
kubectl -n reboot-coordinator get pods
```

Now that you have installed FLUO, nodes without annotation
`flatcar-linux-update.v1.flatcar-linux.net/reboot-paused=true` will be updated automatically when
a new version of Flatcar is available. One-by-one, the selected node is first drained before the
reboot for an update.

### Step 4: Test installation (optional)

You can annotate a node to trigger an automatic reboot:

```bash
export NODE="<node name>"
kubectl annotate node $NODE --overwrite \
    flatcar-linux-update.v1.flatcar-linux.net/reboot-needed="true"
```

You can also SSH into a node and trigger an update check by running
`update_engine_client -check_for_update` or simulate a reboot is needed by running
`locksmithctl send-need-reboot`.
