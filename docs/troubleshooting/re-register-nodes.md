---
title: Reregistering a worker node
weight: 10
---


## Re-register nodes

### Scenario

On Equinix Metal, formerly Packet, if a cluster is updated from `v0.5.0` to `v0.6.0`, it also installs Equinix Metal's Cloud Controller Manager (CCM). The nodes added before `v0.6.0` were configured with BGP using Terraform. From `v0.6.0` onwards they are configured to use CCM for BGP setup. During the update, the older nodes won't see BGP resource removal, but it is possible that someone can disable the BGP from the Equinix Metal console. To make the nodes resilient against such human errors, follow the next steps.

### Steps

> **NOTE:** SSH into a given node in a separate console upfront, before starting to follow steps.

This step ensures that you don't see any abrupt changes. Any workloads running on this node are evicted and scheduled to other nodes. The node is marked as unschedulable after running this command.

```bash
export nodename=""
kubectl drain --ignore-daemonsets $nodename
```

Delete the node object:

```bash
kubectl delete node $nodename
```

SSH into the node:

```bash
sudo systemctl restart kubelet
```
