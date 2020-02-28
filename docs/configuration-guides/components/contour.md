# Contour Ingress controller configuration reference for Lokomotive

## Contents

* [Introduction](#introduction)
* [Prerequisites](#prerequisites)
* [Configuration](#configuration)
* [Argument reference](#argument-reference)
* [Installation](#installation)
* [Uninstalling](#uninstalling)

## Introduction

[Contour](https://github.com/projectcontour/contour) is an Ingress controller for Kubernetes that
deploys the Envoy proxy as a reverse proxy and load balancer.

The Contour Ingress component has different requirements on different platforms. The reason for this
is that an Ingress Controller needs traffic to be routed to their ingress pods, and the network
configurations needed to achieve that differ on each platform.

Currently the supported platform is Packet.

## Prerequisites

* A Lokomotive cluster accessible via `kubectl` deployed on Packet.

* [MetalLB component](metallb.md) installed and configured.

## Configuration

Contour component configuration example:

```tf
component "contour" {
  # Optional arguments
  service_monitor = false
  ingress_hosts = ["*.example.lokomotive.org"]
}
```

## Argument reference

Table of all the arguments accepted by the component.

Example:

| Argument         | Description                                                                                 | Default | Required |
|------------------|---------------------------------------------------------------------------------------------|:-------:|:--------:|
| `service_monitor`| Create ServiceMonitor for Prometheus to scrape Contour and Envoy metrics.                   | false   | false    |
| `ingress_hosts`  | [ExternalDNS component](external-dns.md) creates DNS entries from the values provided.      | ""      | false    |

## Installation

To install the Contour component:

```bash
lokoctl component install contour
```

This component is installed in the `projectcontour` namespace.

## Uninstalling

To uninstall the component:

```bash
lokoctl component render-manifest contour | kubectl delete -f -
```
