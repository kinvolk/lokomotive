---
title: Istio operator configuration reference for Lokomotive
weight: 10
---

## Introduction

Istio is an open-source cloud native service mesh. It uses sidecar pattern to inject Envoy proxies in the pods, which acts as data plane of the Istio service mesh. Control plane components take in user input and implement policies using Envoy proxy.

This component installs the Istio operator.

> **NOTE**: This is an **unsupported and experimental component**. It is not recommended to use this in production. The UX of this component can change anytime as it is tested overtime. If there are fundamental problems with the component then it could be removed without prior notice.

## Prerequisites

* A Lokomotive cluster accessible via `kubectl`.

## Configuration

Istio operator component configuration example:

```tf
component "experimental-istio-operator" {
  enable_monitoring = true
  profile           = "minimal"
}
```

## Attribute reference

Table of all the arguments accepted by the component.

| Argument            | Description                                                                                                                                                                                                                                                                                    | Default     | Type     | Required |
|---------------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|:-----------:|:---------|:--------:|
| `enable_monitoring` | Enable Monitoring for the Istio operator, `istiod`, istio-proxy sub-systems. Make sure that the Prometheus Operator is installed.                                                                                                                                                              | `false`     | `bool`   | false    |
| `profile`           | Istio [configuration profile](https://istio.io/latest/docs/setup/additional-setup/config-profiles/). The profiles provide customization of the Istio control plane and of the sidecars for the Istio data plane. Supported values: `default`, `demo`, `minimal`, `remote`, `empty`, `preview`. | `"minimal"` | `string` | false    |

## Applying

To apply the component:

```bash
lokoctl component apply experimental-istio-operator
```
## Deleting

To destroy the component:

```bash
lokoctl component delete experimental-istio-operator --delete-namespace
```
