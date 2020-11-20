---
title: Using Cloudflare as a DNS provider for Lokomotive
weight: 10
---

## Introduction

This guide explains how to use [Cloudflare](https://www.cloudflare.com/) as a
DNS provider for a Lokomotive cluster.

## Prerequisites

The following are required:

- A Cloudflare account with a hosted DNS zone.
- A [Cloudflare API token](https://developers.cloudflare.com/api/tokens/create)
is required. This token will be used by the Lokomotive tooling to create DNS
records on Cloudflare on your behalf.

The token needs to have the following permissions:

- `Zone.Zone` - read
- `Zone.DNS` - edit

## Steps

### Step 1: Configure Cloudflare as the DNS provider

In your cluster configuration file, configure Cloudflare as the DNS provider:

```hcl
cluster "packet" {
  cluster_name = "my-cluster"
  ...
  dns {
    zone     = "example.com"
    provider = "cloudflare"
  }
  ...
}
```

### Step 2: Deploy the cluster

Set your Cloudflare API token in the `CLOUDFLARE_API_TOKEN` environment
variable:

```
export CLOUDFLARE_API_TOKEN=6IgMwujC2q92AvSgOJ60aCQD5uo9sA
```

Deploy the cluster:

```
lokoctl cluster apply
```

## Additional resources

For full information about how to deploy a Lokomotive cluster, refer to the
[configuration reference](../configuration-reference) or to one of the
[quickstart guides](../quickstarts).
