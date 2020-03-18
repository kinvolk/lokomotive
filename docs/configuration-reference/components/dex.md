# Dex configuration reference for Lokomotive

## Contents

* [Introduction](#introduction)
* [Prerequisites](#prerequisites)
* [Configuration](#configuration)
* [Attribute reference](#attribute-reference)
* [Applying](#applying)
* [Destroying](#destroying)

## Introduction

[Dex](https://github.com/dexidp/dex) is an OpenID Connect (OIDC) and OAuth 2.0 provider with
connectors to many other identity providers such as GitHub, Google or LDAP.

> Dex acts as a portal to other identity providers through "connectors." This lets dex defer
authentication to LDAP servers, SAML providers, or established identity providers like GitHub,
Google, and Active Directory. Clients write their authentication logic once to talk to dex, then dex
handles the protocols for a given backend.

In a Kubernetes context, dex enables:

* usage of authentication providers that don't support OIDC itself and
* grouping of multiple authentication providers.

## Prerequisites

* A Lokomotive cluster accessible via `kubectl`.

* An ingress controller such as [Contour](contour.md) for HTTP ingress.

* A certificate manager such as [cert-manager](cert-manager.md) for valid certificates.

## Configuration

```tf
# dex.lokocfg

variable "google_client_id" {
  type = "string"
}

variable "google_client_secret" {
  type = "string"
}

variable "github_client_id" {
  type = "string"
}

variable "github_client_secret" {
  type = "string"
}

variable "dex_static_client_gangway_id" {
  type = "string"
}

variable "dex_static_client_gangway_secret" {
  type = "string"
}

variable "gangway_redirect_url" {
  type = "string"
}

component "dex" {
  ingress_host = "dex.example.lokomotive-k8s.org"
  issuer_host = "https://dex.example.lokomotive-k8s.org"

  # You can configure one or more connectors. Currently only GitHub and
  # OIDC (for example with Google) are supported from lokoctl.

  # A GitHub connector
  # Requires GitHub OAuth app credentials from https://github.com/settings/developers
  connector "github" {
    id = "github"
    name = "GitHub"

    config {
      client_id = var.github_client_id
      client_secret = var.github_client_secret
      redirect_uri = "https://dex.example.lokomotive-k8s.org/callback"
      team_name_field = "slug"

      org {
        name = "kinvolk"
        teams = [
          "lokomotive-developers",
        ]
      }
    }
  }

  # A OIDC connector
  # Here configured for use with Google
  connector "oidc" {
    id = "google"
    name = "Google"

    config {
      client_id = var.google_client_id
      client_secret = var.google_client_secret
      redirect_uri = "https://dex.example.lokomotive-k8s.org/callback"
      issuer = "https://accounts.google.com"
    }
  }

  # A Google native connector
  connector "google" {
    id   = "google"
    name = "Google"

    config {
      client_id = var.google_client_id
      client_secret = var.google_client_secret
      redirect_uri = "https://dex.example.lokomotive-k8s.org/callback"
      admin_email = "foobar@example.io"
    }
  }
  # only to be defined with Google connector
  gsuite_json_config_path = "project-testing-123456-er12t34y56ui.json"

  static_client {
    name   = "gangway"
    id     = var.dex_static_client_gangway_id
    secret = var.dex_static_client_gangway_secret
    redirect_uris = [var.gangway_redirect_url]
  }
}
```

The secrets can be defined in another file (`lokocfg.vars`) like following:

```tf
google_client_id     = "1234567890123-SqDIX1KFvKPYmuV9Sa8eL92cvxtS3TuP.apps.googleusercontent.com"
google_client_secret = "63zYPITtigLxLaYBEjNP9Taw"

# A random secret key (create one with `openssl rand -base64 32`)
dex_static_client_gangway_secret = "2KBvQkjOZdc3iHt4KSb9GUECdenH/VDl04TwMdSyPcs="
dex_static_client_gangway_id     = "gangway"

gangway_redirect_url = "https://gangway.example.lokomotive-k8s.org/callback"
```

**Note**: More information on the variables used in above dex config can be found in the [gangway
doc](gangway.md#configuration).

## Attribute reference

Table of all the arguments accepted by the component.

Example:

| Argument                                | Description                                                                                                                                                                 | Default | Required |
|-----------------------------------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------|:-------:|:--------:|
| `ingress_host`                          | Used as the `hosts` domain in the ingress resource for dex that is automatically created.                                                                                   | -       | true     |
| `issuer_host`                           | Dex's issuer URL.                                                                                                                                                           | -       | true     |
| `connector`                             | Dex implements connectors that target OpenID Connect and specific platforms such as GitHub, Google etc. Currently only GitHub and OIDC (Google) are supported from lokoctl. | -       | true     |
| `connector.id`                          | ID of the connector.                                                                                                                                                        | -       | true     |
| `connector.name`                        | Name of the connector.                                                                                                                                                      | -       | true     |
| `connector.config`                      | Configuration for the chosen connector.                                                                                                                                     | -       | true     |
| `connector.config.client_id`            | OAuth app client id.                                                                                                                                                        | -       | true     |
| `connector.config.client_secret`        | OAuth app client secret.                                                                                                                                                    | -       | true     |
| `connector.config.issuer`               | The OIDC issuer endpoint. For `oidc` connector only.                                                                                                                        | -       | true     |
| `connector.config.redirect_uri`         | The authorization callback URL.                                                                                                                                             | -       | true     |
| `connector.config.team_name`            | Can be 'name', 'slug' or 'both', see https://github.com/dexidp/dex/blob/master/Documentation/connectors/github.md. For `github` connector only.                             | -       | true     |
| `connector.config.admin_email`          | The email of a GSuite super user. For `google` connector only.                                                                                                              | -       | false    |
| `connector.config.hosted_domains`       | If this field is nonempty, only users from a listed domain will be allowed to log in. For `oidc` and `google` connectors only.                                              | -       | false    |
| `connector.config.org`                  | Define one or more organizations and teams. For `github` connector only.                                                                                                    | -       | true     |
| `connector.config.org.name`             | Name of the GitHub organization.                                                                                                                                            | -       | true     |
| `connector.config.org.teams`            | Name of the team in the provided GitHub organization.                                                                                                                       | -       | true     |
| `gsuite_json_config_path`               | Path to the Gsuite Service Account JSON file. For `google` connector only.                                                                                                  | -       | true     |
| `connector.static_client`               | Configure one or more static clients, i.e. apps that use dex. Example: gangway                                                                                              | -       | true     |
| `connector.static_client.id`            | Client ID used to identify the static client.                                                                                                                               | -       | true     |
| `connector.static_client.secret`        | Client secret used to identify the static client.                                                                                                                          | -       | true     |
| `connector.static_client.name`          | Name used when displaying this client to the end user.                                                                                                                     | -       | true     |
| `connector.static_client.redirect_uris` | A registered set of redirect URIs. When redirecting from dex to the client, the URI requested to redirect to MUST match one of these values.                                | -       | true     |

## Applying

To install the Dex component:

```bash
lokoctl component apply dex
```

### G Suite specific instructions

You need to create a service account on your google suite account and authorize it to view groups on
your domain.

#### Perform G Suite domain-wide delegation of authority

- Follow instructions [here](https://developers.google.com/admin-sdk/directory/v1/guides/delegation)
  to create a service account.

- During Service Account creation a JSON file will be downloaded, give the path to this file in
  dex's config for field `gsuite_json_config_path`.

- While **delegating domain-wide authority to your service account** you will be asked to assign
  scope. In that field select scope
  **`https://www.googleapis.com/auth/admin.directory.group.readonly`** only.

#### Enable admin SDK

Admin SDK lets administrators of enterprise domains to view and manage resources like users, groups
etc. To enable it [click
here](https://console.developers.google.com/apis/library/admin.googleapis.com/).

## Destroying

To destroy the component:

```bash
lokoctl component render-manifest dex | kubectl delete -f -
```
