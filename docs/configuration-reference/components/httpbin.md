# httpbin configuration reference for Lokomotive

## Contents

* [Introduction](#introduction)
* [Prerequisites](#prerequisites)
* [Configuration](#configuration)
* [Attribute reference](#attribute-reference)
* [Applying](#applying)
* [Deleting](#deleting)

## Introduction

[httpbin](https://httpbin.org/) is a simple HTTP request & response service.
It's used mostly for testing purposes.

## Prerequisites

* A Lokomotive cluster accessible via `kubectl`.

## Configuration

httpbin component configuration example:

```tf
component "httpbin" {
  ingress_host = "httpbin.example.lokomotive-k8s.org"
}
```

## Attribute reference

Table of all the arguments accepted by the component.

Example:

| Argument           | Description                                                                                     Default      | Required |
|--------------------|-----------------------------------------------------------------------------------------------|:------------:|:--------:|
| `ingress_host`     | Used as the `hosts` domain in the ingress resource for httpbin that is automatically created. | -            | true     |
| `certmanager_cluster_issuer` | `ClusterIssuer` to be used by cert-manager while issuing TLS certificates. Supported values: `letsencrypt-production`, `letsencrypt-staging`. | `letsencrypt-production` | false |

## Applying

To apply the httpbin component:

```bash
lokoctl component apply httpbin
```

## Deleting

To destroy the component:

```bash
lokoctl component delete httpbin --delete-namespace
```
