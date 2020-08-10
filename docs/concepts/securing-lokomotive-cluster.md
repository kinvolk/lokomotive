---
title: Securing Lokomotive clusters
weight: 10
---

Lokomotive clusters are deployed with security as a top concern.

## Cluster wide Pod Security Policy

A [Pod Security Policy](https://kubernetes.io/docs/concepts/policy/pod-security-policy/) is a
cluster-level resource that controls security sensitive aspects of the pod specification. The
PodSecurityPolicy objects define a set of conditions that a pod must run with in order to be
accepted into the system, as well as defaults for the related fields.

Lokomotive clusters have PodSecurityPolicy (PSP) enabled by default. The cluster comes with two
default PSPs for general purpose application usage:

  * [restricted](../../assets/charts/control-plane/kubernetes/templates/psp-restricted.yaml)

    Allowed to all the workloads in all namespaces except `kube-system` namespace. This PSP has the
    following restrictions:

    * Does not allow pods to be run as root.
    * Allows only whitelisted volumes.
    * Allows only whitelisted capabilities.
    * The Linux kernel host namespace sharing is not allowed for any pod using this PSP. The default
      docker seccomp profiles are used.

    To allow special permissions to workloads, the recommended way is to create a bespoke PSP and
    allow the workloads’ ServiceAccount to use that PSP. Refer to the
    [documentation](https://kubernetes.io/docs/concepts/policy/pod-security-policy/) on creating a
    Pod Security Policy.

  * [privileged](../../assets/charts/control-plane/kubernetes/templates/psp-privileged.yaml)

    Allowed to workloads in `kube-system` namespace only. This PSP does not restrict any workloads
    from any required permissions.

Many projects provide their own Pod Security Policies tailored to their needs which can be used
when deploying if the policies provided by Lokomotive are too strict.

## Global network policy (for Packet platform only)

Lokomotive installs Calico’s
[GlobalNetworkPolicy](https://docs.projectcalico.org/security/calico-network-policy) by default.
This helps with restricting access to the nodes from outside the cluster.

The policy named `ssh` is worth noting, since there we can define who can ssh into the nodes. To edit
the IP address list, run following command:

```console
kubectl edit globalnetworkpolicies ssh
```

Add the IP addresses to the whitelist you want to allow ssh access to the host from.

The list is present at `json path {.spec.ingress[0].source.nets}`.

Also remove the IP block `0.0.0.0/0` from the whitelist, if there is any.

>NOTE: Executing the above step will be overwritten next time a user runs `lokoctl cluster apply`.

To make the changes permanent, the canonical way of doing such an operation is
to edit the cluster configuration file, updating the `management_cidrs` field
and running `lokoctl cluster apply` again.
