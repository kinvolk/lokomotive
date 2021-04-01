---
title: Contour Ingress Controller configuration reference for Lokomotive
linkTitle: Contour Ingress Controller
weight: 10
---

## Introduction

[Contour](https://github.com/projectcontour/contour) is an Ingress controller for Kubernetes that
deploys the Envoy proxy as a reverse proxy and load balancer.

The Contour Ingress component has different requirements on different platforms. The reason for this
is that an Ingress Controller needs traffic to be routed to their ingress pods, and the network
configurations needed to achieve that differ on each platform.

## Prerequisites

* A Lokomotive cluster accessible via `kubectl`.

## Configuration

Contour component configuration example:

```tf
component "contour" {
  # Optional arguments
  enable_monitoring = false
  service_type      = "NodePort"

  node_affinity {
    key      = "node-role.kubernetes.io/node"
    operator = "Exists"
  }

  node_affinity {
    key      = "network.lokomotive.io"
    operator = "In"

    # If the `operator` is set to `"In"`, `values` should be specified.
    values = [
      "foo",
    ]
  }

  toleration {
    key      = "network.lokomotive.io"
    operator = "Equal"
    value    = "contour"
    effect   = "NoSchedule"
  }

  envoy {
    metrics_scrape_interval = "30s"
  }
}
```

## Attribute reference

Table of all the arguments accepted by the component.

| Argument                        | Description                                                                                             | Default        | Type                                                                                                           | Required |
|---------------------------------|---------------------------------------------------------------------------------------------------------|----------------|----------------------------------------------------------------------------------------------------------------|----------|
| `enable_monitoring`             | Create Prometheus Operator configs to scrape Contour and Envoy metrics. Also deploys Grafana Dashboard. | false          | bool                                                                                                           | false    |
| `node_affinity`                 | Node affinity for deploying the operator pod and envoy daemonset.                                       | -              | list(object({key = string, operator = string, values = list(string)}))                                         | false    |
| `service_type`                  | The type of Kubernetes service used to expose Envoy. Set as "NodePort" on the **AWS** platform.         | "LoadBalancer" | string                                                                                                         | false    |
| `toleration`                    | Tolerations that the operator and envoy pods will tolerate.                                             | -              | list(object({key = string, effect = string, operator = string, value = string, toleration_seconds = string })) | false    |
| `envoy.metrics_scrape_interval` | Interval at which Prometheus will scrape Envoy. Valid only when `enable_monitoring` is set to `true`.   | -              | string                                                                                                         | false    |


## Applying

To apply the Contour component:

```bash
lokoctl component apply contour
```

This component is installed in the `projectcontour` namespace.

## Deleting

To destroy the component:

```bash
lokoctl component delete contour --delete-namespace
```
