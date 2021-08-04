---
title: Conformance tests
weight: 50
---

This document enumerates the steps required to run conformance tests for various platforms supported
by Lokomotive.

## Step 1: Platform specific preparations

### Step 1.1: AWS

For AWS you need to make sure that node ports are allowed in the security group. To do so, make sure
you set the `expose_nodeports` cluster property to `true` in the AWS config. Read more about this
flag in the [AWS reference docs](configuration-reference/platforms/aws.md).

To install the AWS cluster, follow the [AWS quick start guide](quickstarts/aws.md).

### Step 1.2: Equinix Metal

#### Step 1.2.1: Cluster Requirements

Create a cluster with at least two worker nodes. Follow the [Equinix Metal quick start
guide](quickstarts/equinix-metal.md) to install a cluster.

#### Step 1.2.2: Expose kube-proxy

Edit the kube-proxy Daemonset config to expose the metrics port on all interfaces. Run the following
command to edit the configuration:

```bash
kubectl -n kube-system edit ds kube-proxy
```

Change the following flag from `- --metrics-bind-address=$(HOST_IP)` to
`- --metrics-bind-address=0.0.0.0`.

#### Step 1.2.3: Expose the node ports on the worker nodes

Apply the following Calico configuration to expose the ports in the default node port range `30000`
to `32767`:

```yaml
echo "
apiVersion: crd.projectcalico.org/v1
kind: GlobalNetworkPolicy
metadata:
  name: allow-nodeport
spec:
  applyOnForward: true
  ingress:
  - action: Allow
    destination:
      ports:
      - 30000:32767
    protocol: TCP
  order: 20
  preDNAT: true
  selector: nodetype == 'worker'
" | kubectl apply -f -
```

## Step 2: Disable the mutating webhook server

Run the following commands to disable the mutating webhook server that disallows the usage `default`
service account tokens:

```bash
kubectl delete MutatingWebhookConfiguration admission-webhook-server
```

## Step 3: Running conformance tests

Follow the canonical document
[here](https://github.com/cncf/k8s-conformance/blob/master/instructions.md) which instructs on
installing sonobuoy and running tests.
