---
title: Web UI configuration reference for Lokomotive
weight: 10
---

## Introduction

Web UI is a web interface to your Lokomotive cluster. It is based on the
[Headlamp](https://github.com/kinvolk/headlamp) project, an easy-to-use and
versatile dashboard for Kubernetes.

It has a clean and modern UI and supports the most common operations for
Kubernetes clusters.

## Prerequisites

* A Kubernetes cluster accessible via `kubectl`.

* An ingress controller such as [Contour](../contour) for HTTP ingress.

* [cert-manager](../cert-manager) to generate TLS certificates.

* Optionally [dex](../dex) to use OIDC for authentication.

## Configuration

```tf
# web-ui.lokocfg

component "web-ui" {
  ingress {
    host                       = "web-ui.example.lokomotive-k8s.org"
    class                      = "contour"
    certmanager_cluster_issuer = "letsencrypt-production"
  }
  oidc {
    client_id     = var.dex_static_client_clusterauth_id
    client_secret = var.dex_static_client_clusterauth_secret
    issuer_url    = "https://dex.example.lokomotive-k8s.org"
  }
}
```

Secrets can be defined in another file (`lokocfg.vars`) like following:

```tf
# A random secret key (create one with `openssl rand -base64 32`)
dex_static_client_clusterauth_secret = "2KBvQkjOZdc3iHt4KSb9GUECdenH/VDl04TwMdSyPcs="
dex_static_client_clusterauth_id     = "clusterauth"
```

### OIDC

To use OIDC for authentication make sure you first have [authentication with
Dex and Gangway](../../../how-to-guides/authentication-with-dex-gangway)
configured. Additionally, you need to add the Web UI redirect URL to the
`static_client.redirect_uris` argument in the dex configuration.

The Web UI redirect URL is `https://web-ui.<CLUSTER_NAME>.<DOMAIN_NAME>/oidc-callback`.

Example:

```tf
  static_client {
    ...

    redirect_uris = [..., "https://web-ui.example.lokomotive-k8s.org/oidc-callback"]
  }
```

Finally, configure the oidc arguments in the Web UI component following the
description in the [Attribute reference](#attribute-reference).

## Attribute reference

Table of all the arguments accepted by the component.

| Argument                             | Description                                                                                                                                   | Default                  | Type   | Required |
|--------------------------------------|-----------------------------------------------------------------------------------------------------------------------------------------------|--------------------------|--------|----------|
| `namespace`                          | Namespace where the Web UI will be installed.                                                                                                 | "lokomotive-system"      | string | false    |
| `ingress`                            | Configuration block for exposing the Web UI through an Ingress resource.                                                                      | -                        | block  | false    |
| `ingress.host`                       | Used as the `hosts` domain in the Ingress resource for web-ui that is automatically created.                                                  | -                        | string | true     |
| `ingress.class`                      | Ingress class to use for the Web UI Ingress.                                                                                                  | `contour`                | string | false    |
| `ingress.certmanager_cluster_issuer` | `ClusterIssuer` to be used by cert-manager while issuing TLS certificates. Supported values: `letsencrypt-production`, `letsencrypt-staging`. | `letsencrypt-production` | string | false    |
| `oidc`                               | Configuration block for setting up OIDC authentication against dex.                                                                           | -                        | block  | false    |
| `oidc.client_id`                     | Static client id. It must match the dex `static_client` name.                                                                                 | -                        | string | true     |
| `oidc.client_secret`                 | Static client secret. It must match the dex `static_client` secret.                                                                           | -                        | string | true     |
| `oidc.issuer_url`                    | Dex's issuer URL. It must match the dex `issuer_host`.                                                                                        | -                        | string | true     |

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
