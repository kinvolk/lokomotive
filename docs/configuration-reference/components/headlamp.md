# Headlamp configuration reference for Lokomotive

## Contents

* [Introduction](#introduction)
* [Prerequisites](#prerequisites)
* [Configuration](#configuration)
* [Attribute reference](#attribute-reference)
* [Applying](#applying)
* [Deleting](#deleting)

## Introduction

[Headlamp](https://github.com/kinvolk/headlamp) is an easy-to-use and versatile
dashboard for Kubernetes.

It has a clean and modern UI, it is vendor independent, generic, and supports
the most common operations for Kubernetes clusters.

## Prerequisites

* A Kubernetes cluster accessible via `kubectl`.

* An ingress controller such as [Contour](contour.md) for HTTP ingress.

* [cert-manager](cert-manager.md) to generate TLS certificates.

## Configuration

```tf
# headlamp.lokocfg

component "headlamp" {
  ingress {
    host                       = "headlamp.example.lokomotive-k8s.org"
    class                      = "contour"
    certmanager_cluster_issuer = "letsencrypt-production"
  }
}
```

## Attribute reference

Table of all the arguments accepted by the component.

Example:

| Argument                             | Description                                                                                                                                   | Default                  | Type   | Required |
|--------------------------------------|-----------------------------------------------------------------------------------------------------------------------------------------------|--:-:---------------------|--:-:---|--:-:-----|
| `namespace`                          | Namespace where Headlamp will be installed.                                                                                                   | "lokomotive-system"      | string | false    |
| `ingress`                            | Configuration block for exposing Headlamp through an Ingress resource.                                                                        | -                        | block  | false    |
| `ingress.host`                       | Used as the `hosts` domain in the Ingress resource for headlamp that is automatically created.                                                | -                        | string | true     |
| `ingress.class`                      | Ingress class to use for the Headlamp Ingress.                                                                                                | `contour`                | string | false    |
| `ingress.certmanager_cluster_issuer` | `ClusterIssuer` to be used by cert-manager while issuing TLS certificates. Supported values: `letsencrypt-production`, `letsencrypt-staging`. | `letsencrypt-production` | string | false    |

## Applying

To apply the Headlamp component:

```bash
lokoctl component apply headlamp
```

## Deleting

To destroy the component:

```bash
lokoctl component delete headlamp
```
