# Cert-manager

[cert-manager](https://docs.cert-manager.io/en/latest/) is a native Kubernetes certificate management controller.

## Requirements

**A cluster with `enable_aggregation` set to `true`.**

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
