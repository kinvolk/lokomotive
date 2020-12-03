---
title: Upgrading bootstrap kubelet
weight: 10
---

## Introduction

[Kubelet](https://kubernetes.io/docs/reference/command-line-tools-reference/kubelet/) is a daemon
that runs on every node and is responsible for managing Pods on the node.

Lokomotive cluster runs two different sets of kubelet processes. Initially, **bootstrap** kubelet
configured on the node as a `rkt` pod joins the cluster, and then kubelet pod managed using
DaemonSet (self-hosted kubelet) takes over the bootstrap kubelet. Self-hosted kubelet allows
seamless updates between Kubernetes patch versions and node configuration using tools like
`kubectl`.

Currently `lokoctl` cannot update bootstrap kubelet, so this document explains how to perform this
operation manually.

## Steps

Perform the following steps on each node, one node at a time.

### Step 1: Drain the node

> **Caution:** If you are using a local directory as a storage for a workload, it will be disturbed
> by this operation. To avoid this move the workload to another node and let the application
> replicate the data. If the application does not support data replication across instances, then
> expect downtime.

```bash
kubectl drain --ignore-daemonsets <node name>
```

### Step 2: Find out the IP and SSH

Find the IP of the node by visiting the cloud provider dashboard. Then, connect to selected machine
using SSH.

```bash
ssh core@<IP Address>
```

### Step 3: Upgrade kubelet on node

Run the following commands:

> **NOTE**: Before proceeding to other commands, set the `latest_kube` variable to the latest
> Kubernetes version. Latest Kubernetes version can be found by running this command after a cluster
> upgrade: `kubectl version -ojson | jq -r '.serverVersion.gitVersion'`.

```bash
export latest_kube=<latest kubernetes version e.g. v1.18.0>
sudo sed -i "s|.*KUBELET_IMAGE_TAG.*|KUBELET_IMAGE_TAG=${latest_kube}|g" /etc/kubernetes/kubelet.env
sudo systemctl restart kubelet
sudo journalctl -fu kubelet
```

Check the logs carefully. If kubelet fails to restart and instructs to do something (e.g. deleting
some file), follow the instructions and reboot the node:

```bash
sudo reboot
```

### Step 4: Verify

**When `disable_self_hosted_kubelet` is `true`**:

Once the node reboots and kubelet rejoins the cluster, output of following command will show new
version across the node name:

```bash
kubectl get nodes
```

**When `disable_self_hosted_kubelet` is `false`**:

Verify that the kubelet service is in active (running) state:

```bash
sudo systemctl status --no-pager kubelet
```

Run the following command to see logs of the process since the last restart:

```bash
sudo journalctl _SYSTEMD_INVOCATION_ID=$(sudo systemctl \
                show -p InvocationID --value kubelet)
```

Once you see the following log lines, you can discern that the kubelet daemon has come up without
errors. Kubelet daemon tries to rejoin the cluster it is taken over by the self hosted kubelet pod
and you see the following logs:

```
Version: <latest_kube>
acquiring file lock on "/var/run/lock/kubelet.lock"
```

### Step 5: Uncordon the node

To finish the upgrade process, uncordon upgraded node to mark it as schedulable for pods, using
the following command:
```bash
kubectl uncordon <node name>
```

## Caveats

- When upgrading kubelet on nodes which are running Rook Ceph, verify that the Ceph cluster is in
  the **`HEALTH_OK`** state. If it is in any other state, **do not proceed with the upgrade** as
  doing so could lead to data loss. When the cluster is in the `HEALTH_OK` state it can tolerate the
  downtime caused by rebooting nodes.
