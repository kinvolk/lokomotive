# Web UI configuration reference for Lokomotive

## Contents

* [Introduction](#introduction)
* [Prerequisites](#prerequisites)
* [Configuration](#configuration)
* [Attribute reference](#attribute-reference)
* [Applying](#applying)
* [Deleting](#deleting)

## Introduction

Web UI is a web interface to your Lokomotive cluster. It is based on the
[Headlamp](https://github.com/kinvolk/headlamp) project, an easy-to-use and
versatile dashboard for Kubernetes.

It has a clean and modern UI and supports the most common operations for
Kubernetes clusters.

## Prerequisites

* A Kubernetes cluster accessible via `kubectl`.

* An ingress controller such as [Contour](contour.md) for HTTP ingress.

* [cert-manager](cert-manager.md) to generate TLS certificates.

## Configuration

```tf
# web-ui.lokocfg

component "web-ui" {
  ingress {
    host                       = "web-ui.example.lokomotive-k8s.org"
    class                      = "contour"
    certmanager_cluster_issuer = "letsencrypt-production"
  }
}
```

## Attribute reference

Table of all the arguments accepted by the component.

| Argument                             | Description                                                                                                                                   | Default                  | Type   | Required |
|--------------------------------------|-----------------------------------------------------------------------------------------------------------------------------------------------|--------------------------|--------|----------|
| `namespace`                          | Namespace where the Web UI will be installed.                                                                                                 | "lokomotive-system"      | string | false    |
| `ingress`                            | Configuration block for exposing the Web UI through an Ingress resource.                                                                      | -                        | block  | false    |
| `ingress.host`                       | Used as the `hosts` domain in the Ingress resource for web-ui that is automatically created.                                                  | -                        | string | true     |
| `ingress.class`                      | Ingress class to use for the Web UI Ingress.                                                                                                  | `contour`                | string | false    |
| `ingress.certmanager_cluster_issuer` | `ClusterIssuer` to be used by cert-manager while issuing TLS certificates. Supported values: `letsencrypt-production`, `letsencrypt-staging`. | `letsencrypt-production` | string | false    |

## Applying

To apply the Web UI component:

```bash
lokoctl component apply web-ui
```

## Deleting

To destroy the component:

```bash
lokoctl component delete web-ui
```
