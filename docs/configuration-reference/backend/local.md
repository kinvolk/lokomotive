---
title: Local backend configuration reference
weight: 10
---

## Introduction

Lokomotive supports local backend for storing Terraform state.

Backend configuration is **OPTIONAL**. If no backend configuration is provided then local backend is
used.

>NOTE: lokoctl does not support multiple backends, configure only one.

## Prerequisites

There are no prerequisites for using local backend.

## Configuration

To use a backend, we need to define a configuration in the `.lokocfg` file.

Example configuration file `local_backend.lokocfg`:

```hcl
backend "local" {
  path = "terraform.tfstate"
}
```

## Attribute reference

Default backend is local.

| Argument             | Description                                         | Default |  Type  | Required |
|----------------------|-----------------------------------------------------|:-------:|:------:|:--------:|
| `backend.local`      | Local backend configuration block.                  |    -    | object |  false   |
| `backend.local.path` | Location where Lokomotive stores the cluster state. |    -    | string |  false   |


