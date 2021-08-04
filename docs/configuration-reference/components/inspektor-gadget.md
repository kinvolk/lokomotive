---
title: Inspektor Gadget configuration reference for Lokomotive
weight: 10
---

## Introduction

[Inspektor Gadget](https://github.com/kinvolk/inspektor-gadget) is a collection
of tools (or gadgets) for debugging and introspecting Kubernetes applications
using BPF.

This component installs the in-cluster part of Inspektor Gadget. To use the
tracing gadgets you need to install the [Inspektor Gadget kubectl
plugin](https://github.com/kinvolk/inspektor-gadget/blob/main/docs/install.md#installing-kubectl-gadget).

The [Kinvolk Web UI](web-ui.md) has integration with the traceloop gadget. When both
the Web UI and Inspektor Gadget components are installed in the cluster, a
new "Traces" menu is available on the Web UI which provides access to pod's
traces via [traceloop](https://github.com/kinvolk/traceloop).

## Prerequisites

* A Kubernetes cluster accessible via `kubectl`.

* Optionally Kinvolk Web UI to use the traceloop integration.

## Configuration

```tf
# inspektor-gadget.lokocfg

component "inspektor-gadget" {
  enable_traceloop = true
}
```

## Attribute reference

Table of all the arguments accepted by the component.

Example:

| Argument           | Description                                                                                                                   | Default       | Type   | Required |
|--------------------|-------------------------------------------------------------------------------------------------------------------------------|---------------|--------|----------|
| `namespace`        | Namespace where Inspektor Gadget will be installed.                                                                           | "kube-system" | string | false    |
| `enable_traceloop` | Whether to enable [traceloop](https://github.com/kinvolk/traceloop) or not. It has a small performance impact on the cluster. | -             | block  | false    |

## Applying

To apply the Inspektor Gadget component:

```bash
lokoctl component apply inspektor-gadget
```

## Deleting

To destroy the component:

```bash
lokoctl component delete inspektor-gadget
```
