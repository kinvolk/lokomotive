---
title: AWS
weight: 10
---

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

## TLS handshake errors in kube-apiserver logs

On the AWS platform, you may see the following logs coming from `kube-apiserver` pods:

```
I0408 05:35:02.865305       1 log.go:172] http: TLS handshake error from 127.0.0.1:45332: read tcp 127.53.210.227:7443->127.0.0.1:45332: read: connection reset by peer
I0408 05:35:12.865457       1 log.go:172] http: TLS handshake error from 127.0.0.1:45424: read tcp 127.53.210.227:7443->127.0.0.1:45424: read: connection reset by peer
I0408 05:35:22.865279       1 log.go:172] http: TLS handshake error from 127.0.0.1:45516: read tcp 127.53.210.227:7443->127.0.0.1:45516: read: connection reset by peer
```

Those logs are harmless and are caused by AWS ELBs opening TCP connections to `kube-apiserver` to probe for availability, without performing a full TLS handshake. Unfortunately, AWS ELBs do not support TLS for probe requests at the time of writing.

There is ongoing [upstream](https://github.com/kubernetes/kubernetes/pull/91277) work to resolve this issue.

## Flatcar Linux Customization

Flatcar Container Linux deployments on AWS can be customized with Container Linux Configs.
For more information, see [Flatcar Container Linux Customization](/docs/concepts/flatcar-container-linux.md#Customization).