---
title: Cert-Manager configuration reference for Lokomotive
weight: 10
---

## Introduction

[cert-manager](https://cert-manager.io/docs/) is a Kubernetes service that provisions TLS
certificates from Let’s Encrypt and other certificate authorities and manages their lifecycles.

## Prerequisites

* A Lokomotive cluster accessible via `kubectl`.

## Configuration

If you run a cluster `enable_aggregation` set to `false`, make sure you disable the webhooks
feature, which will not work without aggregation enabled.

cert-manager component configuration example:

```tf
component "cert-manager" {
  email = "example@example.com"
  namespace = "cert-manager"
}
```

## Attribute reference

Table of all the arguments accepted by the component.

| Argument          | Description                                                    |   Default    |  Type  | Required |
|-------------------|----------------------------------------------------------------|:------------:|:------:|:--------:|
| `email`           | Email used for certificates to receive expiry notifications.   |      -       | string |   true   |
| `namespace`       | Namespace to deploy the cert-manager into.                     | cert-manager | string |  false   |
| `service_monitor` | Specifies how metrics can be retrieved from a set of services. |    false     |  bool  |  false   |


## Applying

To apply the cert-manager component:

```bash
lokoctl component apply cert-manager
```
## Deleting

To destroy the component:

```bash
lokoctl component delete cert-manager --delete-namespace
```
