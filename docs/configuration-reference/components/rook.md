# Rook configuration reference for Lokomotive

## Contents

* [Introduction](#introduction)
* [Prerequisites](#prerequisites)
* [Configuration](#configuration)
* [Attribute reference](#attribute-reference)
* [Applying](#applying)
* [Uninstalling](#uninstalling)

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

  node_selector {
    key      = "node-role.kubernetes.io/node"
    operator = "Exists"
  }

  node_selector {
    key      = "storage.lokomotive.io"
    operator = "In"

    # If the `operator` is set to `"In"`, `values` should be specified.
    values = [
      "foo",
    ]
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
}
```

## Attribute reference

Table of all the arguments accepted by the component.

Example:

| Argument                     | Description                                                                                              | Default | Required |
|------------------------------|----------------------------------------------------------------------------------------------------------|:-------:|:--------:|
| `namespace`                  | Namespace to deploy the rook operator into.                                                              | rook    | false    |
| `node_selector`              | Node selectors for deploying the operator pod.                                                           | -       | false    |
| `toleration`                 | Tolerations that the operator's pods will tolerate.                                                      | -       | false    |
| `agent_toleration_key`       | Toleration key for the rook agent pods.                                                                  | -       | false    |
| `agent_toleration_effect`    | Toleration effect for the rook agent pods. Needs to be specified if `agent_toleration_key` is set.       | -       | false    |
| `discover_toleration_key`    | Toleration key for the rook discover pods.                                                               | -       | false    |
| `discover_toleration_effect` | Toleration effect for the rook discover pods. Needs to be specified if `discover_toleration_key` is set. | -       | false    |

## Applying

To install the Rook component:

```bash
lokoctl component apply rook
```
## Uninstalling

To uninstall the component:

```bash
lokoctl component render-manifest rook | kubectl delete -f -
```

