---
title: Add worker pool in different facility on Equinix Metal
weight: 10
---

## Introduction

Equinix Metal (EM) supports inter-facility network connectivity. Building on that Lokomotive
supports adding worker pool to a Lokomotive cluster in a different facility. The reasons to add a
worker pool in a separate pool could be numerous viz. facility-wide HA, node type availability,
proximity to the application users, etc.

This document provides a step by step guide on adding a worker-pool to existing Lokomotive cluster
but in a different facility than the control plane.

## Prerequisites

* A Lokomotive cluster accessible via `kubectl` deployed on a supported provider.

* Access to Equinix Metal console and permissions to edit project-level settings.

## Steps

### Step 1: Enable "Backend Transfer"

Go to the Equinix Metal console of your project and enable "Backend Transfer" on it, follow [this
document](https://metal.equinix.com/developers/docs/networking/features/#backend-transfer) for
detailed information.

### Step 2: Add private CIDR of the new facility

Go to **Equinix Metal console** > **IPs & Networks** > **IPs**.

Now spot the _"Management"_ IP block (CIDR) for the facility of your choice and make a note of the
`10.xx.xx.xx/25` range.

Open your Lokomotive cluster's `lokocfg` file and add it to the existing `node_private_cidrs` list.

```tf
node_private_cidrs = ["10.10.10.128/25", "10.xx.xx.xx/25"]
```
### Step 3: Add worker pool with a different facility

Add the following snippet to your existing `lokocfg` file, under the `cluster "equinixmetal"` section:

```tf
  worker_pool "worker-new-facility" {
    count    = 1
    facility = "<new facility>"
  }
```

### Step 4: Apply changes

Execute the following command to apply the above changes:

```
lokoctl cluster apply -v --skip-components
```

Once the above command is successfully executed, you will have a worker pool in a separate facility
connected to your existing cluster.
