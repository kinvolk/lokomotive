---
title: Setting up an HTTP ingress controller on Lokomotive with MetalLB and Contour on Packet
weight: 10
---

## Introduction

Kubernetes passes on the responsibility of creating a load balancer for services of type `LoadBalancer`
to the underlying cloud provider. Bare metal providers such as Packet, however, typically don't have an implementation
of network load-balancers. Therefore, these services always remain in the `Pending` state forever.

MetalLB aims to address this problem by offering a load balancer implementation for bare metal Kubernetes
clusters using standard routing protocols.

Contour, on the other hand, addresses the need for ingress traffic management.

Contour is an Ingress controller for Kubernetes that works by deploying the Envoy proxy as a reverse proxy and load balancer.

This guide provides installation steps to configure MetalLB and Contour to help you set up HTTP load balancing
on a Lokomotive cluster with Packet provider.

This how-to guide is expected to take about 15 minutes.

## Learning Objectives

This guide assumes familiarity with Kubernetes and has a basic understanding of Ingress and load balancers.

Upon completion of this guide, you will be able to use Service type `LoadBalancer` in your Lokomotive cluster on Packet.

## Prerequisites

To set up HTTP load balancing, we need the following:

* A Lokomotive cluster accessible via `kubectl` [deployed on Packet](../quickstarts/packet.md).

* IPv4 address pools for MetalLB to allocate — one address per LoadBalancer Service. On Packet, you need to create [Public Elastic IPs](https://support.packet.com/kb/articles/elastic-ips).

## Steps

### Step 1: Configure MetalLB and Contour

MetalLB and Contour are available as a Lokomotive components. A configuration file is needed to install them.

MetalLB operates in two modes: BGP and Layer 2. Lokomotive supports MetalLB in BGP mode.

Create a file named `ingress.lokocfg` with the below contents.

```hcl
# MetalLB component configuration.
component "metallb" {
  address_pools = {
    default = ["a.b.c.d/X"]
  }
}

# Contour component configuration.
component "contour" {}
```

Change "a.b.c.d/X" to the IP address pool CIDR you've created before.

### Step 2: Install MetalLB and Contour

To install, execute:

```bash
lokoctl component apply
```

MetalLB installs in `metallb-system` namespace, whereas Contour installs in `projectcontour` namespace.

In few minutes pods from MetalLB and Contour are in `Running` state.

To verify that the BGP sessions are established, check the logs of the MetalLB speaker pods:

```bash
$ kubectl -n metallb-system logs -l app=metallb,component=speaker
...
{"caller":"bgp.go:63","event":"sessionUp","localASN":65000,"msg":"BGP session established","peer":"10.88.72.128:179","peerASN":65530,"ts":"2019-09-17T13:10:43.194650355Z"}
```

Contour service has an external IP address if it is properly set up with MetalLB.

```bash
kubectl get svc contour -n projectcontour
NAME      TYPE           CLUSTER-IP    EXTERNAL-IP      PORT(S)                      AGE
contour   LoadBalancer   10.3.101.86   1XX.7X.XX9.XXX   80:30511/TCP,443:32317/TCP   5m
```

## Summary

This guide provided step-by-step instructions for setting up MetalLB and Contour on a Lokomotive cluster running on Packet.

In short, MetalLB allows you to create Kubernetes services of type `LoadBalancer` on bare metal cloud providers
that don't provide load balancing capabilities that Kubernetes can make use of.
Contour provides a high-performance Ingress controller for Kubernetes as an alternative to the Nginx Ingress controller.

You can now go ahead and create Ingress resources for your applications using Contour.

## Troubleshooting

**MetalLB**

* Ensure compatibility with cloud providers. You can check compatibility on the MetalLB website under
the [cloud providers section](https://metallb.universe.tf/installation/clouds/).
* Ensure you have assigned an IPv4 address block for MetalLB to use and there are unused IPv4 addresses available to use.

**Contour**

* Envoy container not listening on port 8080 or 8443.

  Contour does not configure Envoy to listen on a port unless there is traffic to be served. For example,
  if you have not configured any TLS ingress objects then Contour does not command Envoy to open port.

## Additional resources

For more extensive and complex configuration for MetalLB, you can visit the MetalLB website for
[configuration options](https://metallb.universe.tf/configuration/).

For more in-depth documentation on Contour, please can visit the [Contour Documentation](https://projectcontour.io/docs/v1.1.0/).
