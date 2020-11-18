---
title: Certificates
weight: 10
---

Lokomotive takes care of generating [Kubernetes PKI](https://kubernetes.io/docs/setup/best-practices/certificates/)
on behalf of the user.

Generation of certificates is currently done in [bootkube Terraform module](https://github.com/kinvolk/lokomotive/tree/v0.5.0/assets/terraform-modules/bootkube).

Following certificates, private and public keys are generated:

- Admission Webhook Server private key
- Admission Webhook Server certificate

- Aggregation CA private key
- Aggregation CA certificate

- Aggregation Client private key
- Aggregation Client certificate

- etcd CA private key
- etcd CA certificate

- etcd client private key
- etcd client certificate

- etcd server private key
- etcd server certificate

- etcd peer private key
- etcd peer certificate

- Kubernetes CA private key
- Kubernetes CA certificate

- kube-apiserver private key
- kube-apiserver certificate

- Kubernetes admin private key
- Kubernetes admin certificate

- service account private key
- service account public key
