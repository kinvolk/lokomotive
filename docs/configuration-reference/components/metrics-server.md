---
title: Metrics Server configuration reference for Lokomotive
weight: 10
---

## Introduction

Metrics server is a cluster addon that is required for supporting the [Horizontal Pod Autoscaler
(HPA)](https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale/).

## Prerequisites

* A Lokomotive cluster with `enable_aggregation` set to true.

## Configuration

This component does not require any specific configuration.

Empty configuration block is also accepted as valid configuration.

Metrics server component configuration example:

```tf
component "metrics-server" {}
```

## Attribute reference

This component does not accept any arguments in its configuration.

## Applying

To apply the Metrics server component:

```bash
lokoctl component apply metrics-server
```

## Deleting

To destroy the component:

```bash
lokoctl component delete metrics-server
```

