# Metrics Server configuration reference for Lokomotive

## Contents

* [Introduction](#introduction)
* [Prerequisites](#prerequisites)
* [Configuration](#configuration)
* [Argument reference](#argument-reference)
* [Installation](#installation)
* [Uninstalling](#uninstalling)

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

## Argument reference

This component does not accept any arguments in its configuration.

## Installation

To install the Metrics server component:

```bash
lokoctl component install metrics-server
```

## Uninstalling

To uninstall the component:

```bash
lokoctl component render-manifest metrics-server | kubectl delete -f -
```

