---
title: Lokomotive admission webhooks
weight: 10
---

As part of the cluster control plane Lokomotive creates additional admission webhooks, which add extra features to the cluster. This document describes the webhooks we install and what they do.

## Mutating webhook - `default` service account

When you create a pod, if you do not specify a service account, the pod is automatically assigned the `default` service account in the same namespace. If you get the raw JSON or YAML for a pod you have created (for example using `kubectl get pods/<podname> -o yaml`), you can see the `spec.serviceAccountName` field has been automatically set.

By default, created pods can authenticate to the Kubernetes API, which is a potential security threat. Not every pod needs the ability to utilize the API. If your application doesn't integrate with Kubernetes and doesn't utilize its API, the application shouldn't have access to API credentials.

To avoid having to manually disable automounting of the `default` service account, we have a webhook server that patches `default` service accounts whenever you either apply any [lokomotive component](../components) or create a new namespace. To see how it works, follow the steps below.

```bash
# Create a namespace.
kubectl create ns foo

# Get default service account for namespace foo.
kubectl get sa default -o yaml -n foo
```

By following the steps above you can see that the `automountServiceAccountToken` field is set to `false`.

For more information see the [kubernetes docs](https://kubernetes.io/docs/tasks/configure-pod-container/configure-service-account/).

### Current limitations

When creating a cluster, the `default` service account for the `default` namespace is not patched. Note that it is generally not recommended to use the `default` namespace for running workloads. However, if you want to patch the `default` namespace, too, you can delete the `default` service account after your cluster is created.

```bash
kubectl delete sa default -n default
```

### Enabling mounting of the `default` service account for pods

It is recommended to create a dedicated service account for your application, giving it the permissions it needs. However, if you still want to enable mounting of the `default` service account, you can opt into automounting API credentials for a particular pod by setting the `spec.automountServiceAccountToken` to `true`. The pod spec takes precedence over the service account if both specify a value for the `spec.automountServiceAccountToken` field.
