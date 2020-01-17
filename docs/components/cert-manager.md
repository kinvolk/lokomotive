# Cert-manager

[cert-manager](https://docs.cert-manager.io/en/latest/) is a native Kubernetes certificate management controller.

## Requirements

**If you run a cluster `enable_aggregation` set to `false`, make sure you disable the webhooks
feature, which will not work without aggregation enabled.**

You can do that by setting `webhooks = false`

## Installation

#### `.lokocfg` file

Declare the component's config:

```
component "cert-manager" {
  email = "example@example.com"
}
```

## Argument Reference

| Argument    | Explanation                                                  | Default      | Required |
|-------------|--------------------------------------------------------------|--------------|----------|
| `email`     | Email used for certificates to receive expiry notifications. | -            | true     |
| `namespace` | Namespace to deploy the cert-manager into.                   | cert-manager | false    |
| `webhooks`  | Controls if webhooks should be deployed.                     | true         | false    |
