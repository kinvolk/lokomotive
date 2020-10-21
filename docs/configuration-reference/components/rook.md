# Rook configuration reference for Lokomotive

## Contents

* [Introduction](#introduction)
* [Prerequisites](#prerequisites)
* [Configuration](#configuration)
* [Attribute reference](#attribute-reference)
* [Applying](#applying)
* [Deleting](#deleting)

## Introduction

[Rook](https://rook.io/docs/rook/v1.2/) is an open-source cloud native storage orchestrator for
Kubernetes, providing the platform, framework, and support for a diverse set of storage solutions to
natively integrate with cloud-native environments.

This component installs the Rook operator.

## Prerequisites

* A Lokomotive cluster accessible via `kubectl`.

## Configuration

Rook component configuration example:

```tf
component "rook" {
  # Optional arguments
  namespace = "rook-test"

  node_selector = {
    "storage.lokomotive.io" = "ceph"
  }

  toleration {
    key      = "storage.lokomotive.io"
    operator = "Equal"
    value    = "rook-ceph"
    effect   = "NoSchedule"
  }

  agent_toleration_key    = "storage.lokomotive.io"
  agent_toleration_effect = "NoSchedule"

  discover_toleration_key    = "storage.lokomotive.io"
  discover_toleration_effect = "NoSchedule"

  enable_monitoring = true
}
```

## Attribute reference

Table of all the arguments accepted by the component.

| Argument                     | Description                                                                                              | Default | Type                                                                                                           | Required |
|------------------------------|----------------------------------------------------------------------------------------------------------|:-------:|:---------------------------------------------------------------------------------------------------------------|:--------:|
| `namespace`                  | Namespace to deploy the rook operator into.                                                              | "rook"  | string                                                                                                         |  false   |
| `node_selector`              | A map with specific labels to run Rook pods selectively on a group of nodes.                             |    -    | map(string)                                                                                                    |  false   |
| `toleration`                 | Tolerations that the operator's pods will tolerate.                                                      |    -    | list(object({key = string, effect = string, operator = string, value = string, toleration_seconds = string })) |  false   |
| `agent_toleration_key`       | Toleration key for the rook agent pods.                                                                  |    -    | string                                                                                                         |  false   |
| `agent_toleration_effect`    | Toleration effect for the rook agent pods. Needs to be specified if `agent_toleration_key` is set.       |    -    | string                                                                                                         |  false   |
| `discover_toleration_key`    | Toleration key for the rook discover pods.                                                               |    -    | string                                                                                                         |  false   |
| `discover_toleration_effect` | Toleration effect for the rook discover pods. Needs to be specified if `discover_toleration_key` is set. |    -    | string                                                                                                         |  false   |
| `enable_monitoring`          | Enable Monitoring for the Rook sub-systems. Make sure that the Prometheus Operator is installed.         |  false  | bool                                                                                                           |  false   |
| `csi_plugin_node_selector`   | A map with specific labels to install Rook CSI plugins on a group of nodes.                              |    -    | map(string)                                                                                                    |  false   |
| `csi_plugin_toleration`      | Tolerations that the Rook CSI plugin installation pods will tolerate.                                    |    -    | list(object({key = string, effect = string, operator = string, value = string, toleration_seconds = string })) |  false   |


## Applying

To apply the Rook component:

```bash
lokoctl component apply rook
```
## Deleting

To destroy the component:

```bash
lokoctl component delete rook --delete-namespace
```
