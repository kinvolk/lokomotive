---
title: Azure Arc onboarding configuration reference for Lokomotive
weight: 10
---

## Introduction

Azure Arc offers simplified management, faster app development, and consistent Azure services.

With Azure Arc, you can:

- Centrally manage a wide range of resources, including
[Windows](https://azure.microsoft.com/en-in/campaigns/windows-server/) and
[Linux](https://azure.microsoft.com/en-in/overview/linux-on-azure/) servers, SQL server,
[Kubernetes](https://azure.microsoft.com/en-in/services/kubernetes-service/) clusters and [Azure
services](https://azure.microsoft.com/en-in/services/azure-arc/hybrid-data-services/).

- Establish central visibility in the [Azure portal](https://azure.microsoft.com/en-in/features/azure-portal/)
and enable multi-environment search with Azure Resource Graph.

- Meet [governance](https://azure.microsoft.com/en-in/solutions/governance/) and compliance standards for
apps, infrastructure and data with [Azure Policy](https://azure.microsoft.com/en-in/services/azure-policy/).

- Delegate access and manage security policies for resources using role-based access control (RBAC) and [Azure
Lighthouse](https://azure.microsoft.com/en-in/services/azure-lighthouse/).

- Organise and inventory assets through a variety of Azure scopes, such as management groups, subscriptions,
resource groups and tags.

This component onboards or removes a Lokomotive cluster with Azure Arc.

## Prerequisites

* Microsoft Azure account with permissions to create ResourceGroup and register an application with the
the Microsoft Identity Platform.

  Detailed instructions and execution steps are mentioned in the
  [How-to-guide](../../how-to-guides/gitops-using-azure-arc-onboarding-component.md#prerequisites).

## Configuration

Azure Arc onboarding component configuration example:

```tf
# azure-arc-onboarding.lokocfg

component "azure-arc-onboarding" {
  application_client_id = "29348jdw-9g23-9kot-21sa-opw129831c2k"
  application_password  = "foobar"
  tenant_id             = "s38kjs4k-x123-89h2-7f21-89uffo109921"
  resource_group        = "azure-arc-lokomotive-resource"
  cluster_name          = "mercury"
}
```

## Attribute reference

Table of all the arguments accepted by the component.

| Argument                | Description                                                                                  | Default  | Type   | Required |
|-------------------------|----------------------------------------------------------------------------------------------|:--------:|:------:|:--------:|
| `application_client_id` | Application ID that uniquely identifies your application within the Azure identity platform. | -        | string | true     |
| `application_password`  | A string value generated that your application can use to identity itself.                   | -        | string | true     |
| `tenant_id`             | Unique ID of the Azure Active Directory tenant.                                              | -        | string | true     |
| `resource_group`        | Name or Id of the Azure resource group.                                                      | -        | string | true     |
| `cluster_name`          | Name of the Lokomotive cluster as provided in the cluster configuration.                     | -        | string | true     |

## Applying

To apply the Azure Arc onboarding component:

```bash
lokoctl component apply azure-arc-onboarding
```
## Deleting

To destroy the component:

```bash
lokoctl component delete azure-arc-onboarding --delete-namespace
```
