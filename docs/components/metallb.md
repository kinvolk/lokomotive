# MetalLB Load Balancer

## General

[MetalLB](https://metallb.universe.tf/) is a load balancer implementation for bare metal Kubernetes
clusters, using standard routing protocols. It allows using k8s services of type `LoadBalancer` on
an infrastructure without native load balancing support.

On IaaS providers such as AWS or GCP, creating a k8s service of type `LoadBalancer` triggers an
automatic creation of a provider-specific load balancer which routes traffic to the k8s service.
However, when running k8s on bare metal environments such as [Packet](https://www.packet.com/),
which typically don't provide native load balancer support, creating a `LoadBalancer` service would
result in the service staying forever in the `Pending` state since no load balancer is created by
the infrastructure provider. MetalLB helps solve this problem by creating a "virtual" load balancer
for each `LoadBalancer` service. It does so using standard network protocols such as BGP and ARP.

## Mode of Operation

MetalLB can operate in two modes: **BGP** and **layer 2**. This component currently supports only
the BGP mode.

MetalLB operates by allocating one IPv4 address to each service of type `LoadBalancer` created on
the cluster. It then advertises this address to one or more upstream BGP routers. This enables both
high availability and load balancing: high availability is achieved since BGP naturally converges
upon node failure, and load balancing is achieved using
[ECMP](https://en.wikipedia.org/wiki/Equal-cost_multi-path_routing).

## Requirements

- A Kubernetes cluster, running Kubernetes 1.9.0 or later, that does not already have network
load-balancing functionality.
- A [compatible](https://metallb.universe.tf/installation/network-addons/) cluster networking addon.
- An IPv4 CIDR for MetalLB to allocate - one address is needed per `LoadBalancer` service.
- One or more routers capable of speaking BGP.

## Installation

Add a `component` block for MetalLB in your `*.lokocfg` file while specifying the CIDR under
`address_pools`:

```hcl
...

component "metallb" {
    address_pools = {
      default = ["147.63.8.20/32"]
    }
}
```

MetalLB will use the specified CIDR for exposing services of type `LoadBalancer`.

Install the component by running the following:

```bash
lokoctl component install metallb
```

### Address Pools

The `address_pools` parameter is a map which allows specifying one or more CIDRs which MetalLB can
use to expose services. Multiple pools may be specified, and multiple CIDRs may be specified per
pool:

```hcl
component "metallb" {
    address_pools = {
      default = ["147.63.8.20/32"]
      special_addresses = ["147.85.47.16/29", "147.85.47.24/29"]
    }
}
```

### Node Selection

Optionally, it is possible to run MetalLB selectively on a group of nodes. A common use case for
this is a cluster where a dedicated worker group serves ingress traffic.

The following can be included in the `*.lokocfg` file to force MetalLB to run on nodes with
specific labels:

```
component "metallb" {
  controller_node_selectors = {
    "kubernetes.io/hostname" = "worker3"
  }

  speaker_node_selectors = {
    "ingress_node" = "true"
    "node-role.kubernetes.io/node" = ""
  }
}
```

The above sets a k8s
[nodeSelector](https://kubernetes.io/docs/concepts/configuration/assign-pod-node/#nodeselector)
for MetalLB *controller* pods and MetalLB *speaker* pods, respectively.

>NOTE: Label *keys* containing special characters should be quoted. Label values must be quoted. It
>is safer to always quote both the keys and the values.

>NOTE: Empty label values should be specified using `""`.

More information on MetalLB configuration can be found in the official
[docs](https://metallb.universe.tf/configuration/)..

### Pod Tolerations
You can specify one or more tolerations for both speaker and controller pods as part of the component configuration in the `*.lokocfg` file.
Example:

```
component "metallb" {
  speaker_toleration {
    key = "speaker_key1"
    operator = "Equal"
    value = "value1"
  }
  speaker_toleration {
    key = "speaker_key2"
    operator = "Equal"
    value = "value2"
  }

  controller_toleration {
    key = "controller_key1"
    operator = "Equal"
    value = "value1"
  }
  controller_toleration {
    key = "controller_key2"
    operator = "Equal"
    value = "value2"
  }
}
```

### Prometheus monitoring
If you want a ServiceMonitor to be created for Prometheus to be able to scrape MetalLB metrics, add the following as part of the component's configuration in the `*.lokocfg` file.

**Note: You should already have [prometheus-operator component](/docs/components/prometheus-operator) installed before doing this.**

```
component "metallb" {
	service_monitor = true
}
```
