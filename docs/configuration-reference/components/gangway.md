# Gangway configuration reference for Lokomotive

## Contents

* [Introduction](#introduction)
* [Prerequisites](#prerequisites)
* [Configuration](#configuration)
* [Attribute reference](#attribute-reference)
* [Applying](#applying)
* [Deleting](#deleting)

## Introduction

[Gangway](https://github.com/heptiolabs/gangway) is a web application that allows obtaining OIDC
tokens from identity providers and automatically generating kubeconfigs to be used by Kubernetes
users.

## Prerequisites

* A Lokomotive cluster accessible via `kubectl`.

* [Dex](dex.md) installed with a static client for gangway.

## Configuration

Gangway component configuration example:

```tf
# gangway.lokocfg

variable "gangway_session_key" {
  type = "string"
}

component "gangway" {
  # The name of the cluster. This is used to name the kubectl configuration context.
  cluster_name = "example"

  # Used as the `hosts` domain in the ingress resource for gangway that is
  # automatically created
  ingress_host = "gangway.example.lokomotive-k8s.org"

  session_key = var.gangway_session_key

  # Where kube-apiserver is reachable
  api_server_url = "https://example.lokomotive-k8s.org:6443"

  # Where the 'auth' endpoint is
  authorize_url = "https://dex.example.lokomotive-k8s.org/auth"

  # Where the 'token' endpoint is
  token_url = "https://dex.example.lokomotive-k8s.org/token"

  # The static client id and secret
  client_id     = var.dex_static_client_gangway_id
  client_secret = var.dex_static_client_gangway_secret

  # gangway's redirect URL, i.e. where the OIDC endpoint should callback to
  redirect_url = var.gangway_redirect_url
}
```

The secrets can be defined in another file (`lokocfg.vars`) like following:

```tf
gangway_redirect_url         = "https://gangway.example.lokomotive-k8s.org/callback"

# A random secret key (create one with `openssl rand -base64 32`)
gangway_session_key              = "5Rsz5C4qRqYFoAfYcXOedQOyQpHTXyLiWFYvtjwjtm0="
dex_static_client_gangway_secret = "2KBvQkjOZdc3iHt4KSb9GUECdenH/VDl04TwMdSyPcs="
dex_static_client_gangway_id     = "gangway"
```
## Attribute reference

Table of all the arguments accepted by the component.

Example:

| Argument                     | Description                                                                                                                                   |         Default          |  Type  | Required |
|------------------------------|-----------------------------------------------------------------------------------------------------------------------------------------------|:------------------------:|:------:|:--------:|
| `cluster_name`               | The name of the cluster.                                                                                                                      |            -             | string |   true   |
| `ingress_host`               | Used as the `hosts` domain in the ingress resource for gangway that is automatically created.                                                 |            -             | string |   true   |
| `certmanager_cluster_issuer` | `ClusterIssuer` to be used by cert-manager while issuing TLS certificates. Supported values: `letsencrypt-production`, `letsencrypt-staging`. | `letsencrypt-production` | string |  false   |
| `sesion_key`                 | Gangway session key.                                                                                                                          |            -             | string |   true   |
| `api_server_url`             | URL of Kubernetes API server.                                                                                                                 |            -             | string |   true   |
| `authorize_url`              | Auth endpoint of Dex.                                                                                                                         |            -             | string |   true   |
| `token_url`                  | Token endpoint of Dex.                                                                                                                        |            -             | string |   true   |
| `client_id`                  | Static client ID.                                                                                                                             |            -             | string |   true   |
| `client_secret`              | Static client secret.                                                                                                                         |            -             | string |   true   |
| `redirect_url`               | Gangway's redirect URL, i.e. OIDC callback endpoint.                                                                                          |            -             | string |   true   |


## Applying

To apply the Gangway component:

```bash
lokoctl component apply gangway
```
## Deleting

To destroy the component:

```bash
lokoctl component delete gangway --delete-namespace
```

