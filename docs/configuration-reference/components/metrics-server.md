# Metrics Server configuration reference for Lokomotive

## Contents

* [Introduction](#introduction)
* [Prerequisites](#prerequisites)
* [Configuration](#configuration)
* [Attribute reference](#attribute-reference)
* [Applying](#applying)
* [Destroying](#destroying)

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

To install the Metrics server component:

```bash
lokoctl component apply metrics-server
```

## Destroying

To uninstall the component:

```bash
lokoctl component render-manifest metrics-server | kubectl delete -f -
```

