---
title: Linkerd configuration reference for Lokomotive
weight: 10
---

## Introduction

Linkerd is an open-source, cloud-native and lightweight service mesh. Linkerd uses sidecar proxy (linkerd-proxy) to create a mesh among all pods. Control plane components are installed in linkerd namespace.

> **NOTE**: This is an **unsupported and experimental component**. It is not recommended to use this in production. The UX of this component can change anytime as it is tested overtime. If there are fundamental problems with the component then it could be removed without prior notice.

## Prerequisites

* A Lokomotive cluster accessible via `kubectl`.

## Configuration

Linkerd configuration example:

```tf
component "experimental-linkerd" {
  controller_replicas = 2
  enable_monitoring   = true
  prometheus_url      = "http://prometheus-operator-prometheus.monitoring:9090"
}
```

## Attribute reference

Table of all the arguments accepted by the component.

| Argument              | Description                                                                                                                                                                    | Default | Type      | Required |
|-----------------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|---------|-----------|----------|
| `enable_monitoring`   | Enable Monitoring for the Linkerd control plane components. Make sure that the [Prometheus Operator](../../how-to-guides/monitoring-with-prometheus-operator.md) is installed. | `false` | `bool`    | false    |
| `controller_replicas` | Number of replicas of control plane components like: `controller`, `destination`, `identity`, `proxy-injector`, `sp-validator`, `tab`.                                         | `1`     | `integer` | false    |
| `prometheus_url`      | URL of the external prometheus, Linkerd will scrape its control plane metrics from this Prometheus instance.                                                                   | `""`    | `string`  | false    |

## Applying

To apply the component:

```bash
lokoctl component apply experimental-linkerd
```

## Deleting

To destroy the component:

```bash
lokoctl component delete experimental-linkerd --delete-namespace
```
