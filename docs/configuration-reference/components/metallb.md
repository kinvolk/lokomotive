# MetalLB configuration reference for Lokomotive

## Contents

* [Introduction](#introduction)
* [Prerequisites](#prerequisites)
* [Configuration](#configuration)
* [Attribute reference](#attribute-reference)
* [Applying](#applying)
* [Deleting](#deleting)

## Introduction

[MetalLB](https://metallb.universe.tf/) is a load balancer implementation for bare metal Kubernetes
clusters, using standard routing protocols. It allows using Kubernetes Service of type `LoadBalancer`
on an infrastructure without native load balancing support.

On IaaS providers such as AWS or GCP, creating a Kubernetes Service of type `LoadBalancer` triggers an
automatic creation of a provider-specific load balancer which routes traffic to the Kubernetes service.
However, when running Kubernetes on bare metal environments such as [Packet](https://www.packet.com/),
which typically don't provide native load balancer support, creating a `LoadBalancer` Service would
result in the service staying forever in the `Pending` state since no load balancer is created by
the infrastructure provider. MetalLB helps solve this problem by creating a "virtual" load balancer
for each `LoadBalancer` service. It does so using standard network protocols such as BGP and ARP.

## Prerequisites

* A Lokomotive cluster accessible via `kubectl` deployed on Packet.

* A [compatible](https://metallb.universe.tf/installation/network-addons/) cluster networking addon.

* At least one IPv4 address pool for MetalLB to allocate - one address is needed per `LoadBalancer` service.

* One or more routers capable of speaking BGP.

## Configuration

MetalLB can operate in two modes: **BGP** and **layer 2**. This component currently supports only
the BGP mode.

MetalLB operates by allocating one IPv4 address to each service of type `LoadBalancer` created on
the cluster. It then advertises this address to one or more upstream BGP routers. This enables both
high availability and load balancing: high availability is achieved since BGP naturally converges
upon node failure, and load balancing is achieved using
[ECMP](https://en.wikipedia.org/wiki/Equal-cost_multi-path_routing).


MetalLB component configuration example:

```tf
component "metallb" {
  address_pools = {
    pool1 = ["147.63.8.20/32"]
    pool2 = ["147.85.47.16/29", "147.85.47.24/29"]
  }
  controller_node_selectors = {
    "kubernetes.io/hostname" = "worker3"
  }
  speaker_node_selectors = {
    "ingress_node" = "true"
    "node-role.kubernetes.io/node" = ""
  }
  speaker_toleration {
    key = "speaker_key1"
    operator = "Equal"
    value = "value1"
  }
  controller_toleration {
    key = "controller_key1"
    operator = "Equal"
    value = "value1"
  }
  # If true, then Prometheus Operator component should be installed.
  service_monitor = true
}
```

MetalLB will use the specified CIDR for exposing services of type `LoadBalancer`.

### Advanced IP allocation

By default, MetalLB uses all specified address pools to allocate IP addresses to services. To
request an address from a specific pool, set the `metallb.universe.tf/address-pool` annotation for
the relevant service:

```yaml
apiVersion: v1
kind: Service
metadata:
  name: nginx
  annotations:
    metallb.universe.tf/address-pool: pool2
spec:
  ports:
  - port: 80
    targetPort: 80
  selector:
    app: nginx
  type: LoadBalancer
```

## Attribute reference

Table of all the arguments accepted by the component.

Example:

| Argument                    | Description                                                                                | Default | Type                                                                                                           | Required |
|-----------------------------|--------------------------------------------------------------------------------------------|:-------:|:---------------------------------------------------------------------------------------------------------------|:--------:|
| `address_pools`             | A map which allows specifying one or more CIDRs which MetalLB can use to expose services.  |    -    | map(list(string))                                                                                              |   true   |
| `controller_node_selectors` | A map with specific labels to run MetalLB controller pods selectively on a group of nodes. |    -    | map(string)                                                                                                    |  false   |
| `speaker_node_selectors`    | A map with specific labels to run MetalLB speaker pods selectively on a group of nodes.    |    -    | map(string)                                                                                                    |  false   |
| `controller_toleration`     | Specify one or more tolerations for controller pods.                                       |    -    | list(object({key = string, effect = string, operator = string, value = string, toleration_seconds = string })) |  false   |
| `speaker_toleration`        | Specify one or more tolerations for speaker pods.                                          |    -    | list(object({key = string, effect = string, operator = string, value = string, toleration_seconds = string })) |  false   |
| `service_monitor`           | Create ServiceMonitor for Prometheus to scrape MetalLB metrics.                            |  false  | bool                                                                                                           |  false   |


## Applying

To apply the MetalLB component:

```bash
lokoctl component apply metallb
```

## Deleting

To destroy the component:

```bash
lokoctl component delete metallb --delete-namespace
```
