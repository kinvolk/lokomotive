---
title: Troubleshooting OpenEBS
weight: 10
---

## Introduction

This guide enlists various issues with OpenEBS and steps to mitigate those problems.

## Prerequisites

* A Kubernetes cluster accessible via `kubectl`.

* [OpenEBS component](../../configuration-reference/components/openebs-operator) installed.

## iSCSI volume mount issue

If you are running Flatcar version `2605.9.0` and deploy OpenEBS with version `0.5.0` of Lokoctl,
you might run into volume mount issue. The pod trying to mount the volume continues to stay in
`ContainerCreating` state. To fix this issue follow the following steps. This volume mount issue is
fixed in [this PR](https://github.com/kinvolk/lokomotive/pull/1266) in Lokomotive and is release as
a part of `0.6.0`. The nodes deployed using lokoctl version 0.5.0 or before need to do following
fix.

### Step 1: Fix the self-hosted Kubelet

> **NOTE**: If any of the following commands fail then you might not be running a self-hosted
> Kubelet in your cluster.

#### Step 1.1: Update the image

```bash
kubectl -n kube-system set image ds kubelet kubelet=quay.io/kinvolk/kubelet:v1.19.4
```

#### Step 1.2: Remove the iSCSI mount

```bash
kubectl -n kube-system edit ds kubelet
```

Remove the following snippet from the kubelet Daemonset configuration:

```yaml
- mountPath: /usr/sbin/iscsiadm
  name: iscsiadm
```

### Step 2: Fix the `iscsid` service on hosts

SSH into all the worker nodes that will try to mount the volume created by OpenEBS. Run the
following set of commands to fix the `iscsid` service.

Acquire root credentials:

```bash
sudo -i
```

Create the `iscsid` service dropin and restart the service with new config, copy paste the following commands:

```bash
mkdir /etc/systemd/system/iscsid.service.d
cat <<EOF > /etc/systemd/system/iscsid.service.d/00-iscsid.conf
[Service]
ExecStartPre=/bin/bash -c 'echo "InitiatorName=$(/sbin/iscsi-iname -p iqn.2020-01.io.kinvolk:01)" > /etc/iscsi/initiatorname.iscsi'
EOF
systemctl daemon-reload
systemctl restart iscsid
systemctl status --no-page iscsid
journalctl --no-pager -u iscsid
```

### Step 3: Fix the bootstrap kubelet

Remove all the occurrences of `iscsiadm` from kubelet service file:

```bash
sed -i '/iscsiadm/d' /etc/systemd/system/kubelet.service
```

Restart kubelet process:

```bash
systemctl daemon-reload
systemctl restart kubelet
systemctl status --no-page kubelet
journalctl --no-pager -u kubelet
```

### Step 4: Verify the stuck pod

Check if the pod stuck in `ContainerCreating` state is in `Running` state now.
