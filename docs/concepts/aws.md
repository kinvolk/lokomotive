# AWS

AWS is one of the cloud platforms supported by Lokomotive. This document explains various architecture decisions and details specific to this platform.

## Access to EC2 metadata endpoint

By default, access to [EC2 instance metadata](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ec2-instance-metadata.html) is blocked for all pods. This is to prevent possible exploitation of IAM roles which might be attached to the cluster nodes.

To allow an application to access the metadata service, you can create a NetworkPolicy selecting the application.

Here's a simple NetworkPolicy that allows pods with the label `foo: foo` to send packets to any IP address including the metadata service.

For simplicity, this is a very open NetworkPolicy. You should consider creating a more restrictive one for your production clusters.

```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: allow-metadata-access
spec:
  podSelector:
    matchLabels:
      foo: foo
  policyTypes:
  - Egress
  egress:
  - to:
    - ipBlock:
        cidr: 0.0.0.0/0
```
