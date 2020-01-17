# Metrics Server

Metrics server is a cluster addon that is required for supporting [Horizontal Pod Autoscaler (HPA)](https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale/).

## Requirements

**A cluster with `enable_aggregation` set to `true`(default).**

This is important as currently changing the option once the cluster has already been setup with [Lokomotive](https://github.com/kinvolk/lokomotive-kubernetes) does not have any effect.

## Installation

```bash
lokoctl component install metrics-server
```
