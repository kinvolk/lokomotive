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
- An IPv4 address pool for MetalLB to allocate - one address is needed per `LoadBalancer` service.
- One or more routers capable of speaking BGP.

## Installation

Install the component by running the following:

```bash
lokoctl component install metallb
```

MetalLB requires a ConfigMap specifying the BGP peering configuration as well as the address pool(s)
to allocate IPs from. The ConfigMap isn't automatically created by `lokoctl`.

Create a `ConfigMap` similar to the following, changing the values as indicated below:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  namespace: metallb-system
  name: config
data:
  config: |
    peers:
    # One peer should be defined for each worker node which should form a BGP session with the
    # infrastructure provider. Typically these would be "ingress nodes", or nodes through which
    # traffic enters the cluster. Using more nodes improves both high availability and network
    # traffic distribution, however some infrastructure providers have a limit on the number of
    # ECMP next hops, which may require to carefully consider which nodes to run BGP on.
    - node-selectors:
      - match-labels:
          # K8s node name of the worker.
          kubernetes.io/hostname: worker-0
      # Address of the *external* BGP router with which the worker node should form a BGP session.
      # On Packet this is the gateway address for the private IPv4 network.
      # This can be found under the "GATEWAY" field for the IPv4 private IP on the
      # "Overview" section in: https://app.packet.net/devices/<device_id>.
      peer-address: 10.64.43.10
      # The BGP autonomous system number on the external BGP router.
      peer-asn: 65530
      # The BGP autonomous system number MetalLB should use on the node. This will likely be
      # dictated by the infrastructure provider.
      my-asn: 65000
      # Affects BGP convergence time in case of failure. In stable network environments (i.e.
      # environments in which packet loss between MetalLB and the BGP routers is unlikely), this
      # option should probably be set as low as possible. More info:
      # https://www.juniper.net/documentation/en_US/junos/topics/reference/configuration-statement/hold-time-edit-protocols-bgp.html
      hold-time: 3s
    - node-selectors:
      - match-labels:
          kubernetes.io/hostname: worker-1
      peer-address: 10.64.54.2
      peer-asn: 65530
      my-asn: 65000
      hold-time: 3s

    address-pools:
    - name: default
      protocol: bgp
      # An IPv4 IP address pool to be allocated to k8s services by MetalLB. These should be
      # internet-routable ("public") IPv4 addresses if the k8s services need to be reachable from
      # the internet.
      # On Packet, an Elastic IP block of type "public" (not "global"!) should be used. While
      # "global" EIPs can be technically used, they are unnecessary for most use cases and are much
      # more expensive than "public" EIPs.
      # One address is required for each service which needs to be exposed by MetalLB.
      addresses:
      - 147.75.40.46/32
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
