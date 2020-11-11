---
title: Setting up third party OAuth for Grafana
weight: 10
---

## Introduction

Grafana is a sub-component deployed as a part of Lokomotive's `prometheus-operator` component. By
default you can provide an admin user password for Grafana, but what if you want to allow your team
members to view the dashboards? Sharing a single password in such circumstances is cumbersome and
insecure. OAuth comes to our rescue and Grafana supports multiple OAuth providers out of the box.
This document explains how to enable any supported auth provider on Grafana.

## Prerequisites

- A Lokomotive cluster deployed on AWS or Packet.

- [MetalLB](https://metallb.universe.tf/) deployed on the cluster.

  **NOTE**: Required only for the Packet provider.

  Installation instructions for [MetalLB](./ingress-with-contour-metallb.md) component.

- [Contour](https://projectcontour.io/) deployed on the cluster.

  Installation instructions for [Contour](../configuration-reference/components/contour.md)
  component.

- [cert-manager](https://cert-manager.io/docs/) deployed on the cluster.

  Installation instructions for
  [cert-manager](../configuration-reference/components/cert-manager.md) Lokomotive component.

- [ExternalDNS](https://github.com/kubernetes-sigs/external-dns) deployed on the cluster.

  Installation instructions for [ExternalDNS](../configuration-reference/components/external-dns.md)
  component.

## Steps

> **NOTE**: This guide assumes that the OAuth provider is GitHub. For other OAuth providers, the
> steps are the same, but the secret environment variables will change, as mentioned in [Step
> 2](#step-2-add-prometheus-operator-component-configuration). Grafana docs explain how to convert
> the `ini` config to environment variables
> [here](https://grafana.com/docs/grafana/latest/administration/configuration/#configure-with-environment-variables).

### Step 1: Create Github application

- Create a GitHub OAuth application as documented in the [Grafana
  docs](https://grafana.com/docs/grafana/latest/auth/github/).

- Set **Homepage URL** to `https://grafana.<cluster name>.<DNS zone>`. This should be same as the
  `prometheus-operator.grafana.ingress.host` as shown in [Step
  2](#step-2-add-prometheus-operator-component-configuration).

- Set **Authorization callback URL** to `https://grafana.<cluster name>.<DNS zone>/login/github`.

- Make a note of `Client ID` and `Client Secret`, they will be needed in [Step
  3](#step-3-add-secret-information).

### Step 2: Add `prometheus-operator` component configuration

Create a file named `prometheus-operator.lokocfg` with the following contents or if you already
have `prometheus-operator` installed then add the following contents to the existing configuration:

```tf
variable "gf_auth_github_client_id" {}
variable "gf_auth_github_client_secret" {}
variable "gf_auth_github_allowed_orgs" {}

component "prometheus-operator" {
  grafana {
    secret_env = {
      "GF_AUTH_GITHUB_ENABLED"               = "'true'"
      "GF_AUTH_GITHUB_ALLOW_SIGN_UP"         = "'true'"
      "GF_AUTH_GITHUB_SCOPES"                = "user:email,read:org"
      "GF_AUTH_GITHUB_AUTH_URL"              = "https://github.com/login/oauth/authorize"
      "GF_AUTH_GITHUB_TOKEN_URL"             = "https://github.com/login/oauth/access_token"
      "GF_AUTH_GITHUB_API_URL"               = "https://api.github.com/user"
      "GF_AUTH_GITHUB_CLIENT_ID"             = var.gf_auth_github_client_id
      "GF_AUTH_GITHUB_CLIENT_SECRET"         = var.gf_auth_github_client_secret
      "GF_AUTH_GITHUB_ALLOWED_ORGANIZATIONS" = var.gf_auth_github_allowed_orgs
    }

    ingress {
      host = "grafana.<cluster name>.<DNS zone>"
    }
  }
}
```

> **NOTE**: On Packet, you either need to create a DNS entry for `grafana.<cluster name>.<DNS zone>`
> and point it to the Packet external IP for the contour service (see the [Packet ingress guide for
> more details](./ingress-with-contour-metallb.md)) or use the [External DNS
> component](../configuration-reference/components/external-dns.md).

> **NOTE**: In the above configuration, boolean values are set to `"'true'"` instead of bare
> `"true"` because Kubernetes expects the key-value pair to be of type `map[string]string` and not
> `map[string]bool`.

### Step 3: Add secret information

Create a `lokofg.vars` file or add the following to an existing file, setting the values of this
secret as needed:

```tf
gf_auth_github_client_id     = "YOUR_GITHUB_APP_CLIENT_ID"
gf_auth_github_client_secret = "YOUR_GITHUB_APP_CLIENT_SECRET"
gf_auth_github_allowed_orgs  = "YOUR_GITHUB_ALLOWED_ORGANIZATIONS"
```

Replace `YOUR_GITHUB_APP_CLIENT_ID` with `Client ID` and `YOUR_GITHUB_APP_CLIENT_SECRET` with
`Client Secret` collected in [Step 1](#step-1-create-github-application). And replace
`YOUR_GITHUB_ALLOWED_ORGANIZATIONS` with the Github organization that your users belong to.

### Step 4: Deploy and access the dashboard

Deploy the `prometheus-operator` component using the following command:

```bash
lokoctl component apply prometheus-operator
```

Go to `https://grafana.<cluster name>.<DNS zone>` and use the **Sign in with GitHub** button, to
sign in with Github.

## Additional resources

- Other auth providers for Grafana:
  https://grafana.com/docs/grafana/latest/auth/overview/#user-authentication-overview

- Component `prometheus-operator`'s configuration reference can be found
  [here](../configuration-reference/components/prometheus-operator.md).

- Find details on how to setup monitoring with the `prometheus-operator` component
  [here](./monitoring-with-prometheus-operator.md).
