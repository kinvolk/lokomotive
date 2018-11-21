# cert-manager

This will help you install the cert-manager component.


## Reference Answers

- `namespace`: Provide the namespace in which you want to deploy the cert-manager, if this is not provided by default value is `cert-manager`.

- `email`: Provide the email address which should be used for ACME registration.

## Sample answers file

```yaml
namespace: cert-manager
email: foo@bar.com
```
