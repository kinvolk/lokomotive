---
title: Backup Rook Ceph volume on S3 using Velero
weight: 10
---

## Introduction

[Rook](https://rook.io/) is a component of Lokomotive which provides storage on Equinix Metal. Taking
regular backup of the data to a remote server is an essential strategy for disaster recovery.

[Velero](https://velero.io/) is another component of Lokomotive which helps you to backup entire
namespaces, including volume data in them.

## Learning objectives

This document will walk you through the process of backing up a namespace including the volume in
it.

## Prerequisites

- A Lokomotive cluster deployed on a Equinix Metal and accessible via `kubectl`.

- Rook Ceph installed by following [this guide](./rook-ceph-storage.md).

- `aws` CLI tool [installed](https://docs.aws.amazon.com/cli/latest/userguide/install-cliv2.html).

- S3 bucket created by following [these
  instructions](https://github.com/vmware-tanzu/velero-plugin-for-aws/blob/8d31a11/README.md#create-s3-bucket).

- Velero user in AWS created by following [these
  instructions](https://github.com/vmware-tanzu/velero-plugin-for-aws/blob/8d31a11/README.md#option-1-set-permissions-with-an-iam-user).

- Velero CLI tool [downloaded](https://github.com/vmware-tanzu/velero/releases/tag/v1.4.2) and
  installed in the `PATH`.

## Steps

### Step 1: Deploy Velero

#### Config

Create a file `velero.lokocfg` with the following contents:

```tf
component "velero" {
  provider = "restic"

  restic {
    credentials = file("./credentials-velero")
    backup_storage_location {
      provider = "aws"
      bucket   = "rook-ceph-backup"
      region   = "us-west-1"
    }
  }
}
```

In the above config `region` should match the region of bucket created previously using `aws` CLI.

Replace the `./credentials-velero` with path to AWS credentials file for the `velero` user.

#### Deploy

Execute the following command to deploy the `velero` component:

```bash
lokoctl component apply velero
```

Verify the pods in the `velero` namespace are in the `Running` state (this may take a few minutes):

```console
$ kubectl -n velero get pods
NAME                     READY   STATUS    RESTARTS   AGE
restic-c27rq             1/1     Running   0          2m
velero-66d5d67b5-g54x7   1/1     Running   0          2m
```

### Step 2: Deploy sample workload

If you already have an application you want to backup, then skip this step.

Let us deloy a stateful application and save some demo data in it. Save the following YAML config in
a file named `stateful.yaml`:

```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: demo-ns
---
apiVersion: apps/v1
kind: StatefulSet
metadata:
  labels:
    app: demo-app
  name: demo-app
  namespace: demo-ns
spec:
  replicas: 1
  serviceName: "demo-app"
  selector:
    matchLabels:
      app: demo-app
  template:
    metadata:
      labels:
        app: demo-app
    spec:
      securityContext:
        runAsNonRoot: true
        runAsUser: 65534
        runAsGroup: 65534
      containers:
      - image: busybox:1
        name: app
        command: ["/bin/sh"]
        args:
        - -c
        - "echo 'sleeping' && sleep infinity"
        volumeMounts:
        - mountPath: "/data"
          name: data
  volumeClaimTemplates:
  - metadata:
      name: data
    spec:
      accessModes:
      - ReadWriteOnce
      resources:
        requests:
          storage: 100Mi
```

Execute the following command to deploy the application:

```bash
kubectl apply -f stateful.yaml
```

Verify the application is running fine:

```console
$ kubectl -n demo-ns get pods
NAME         READY   STATUS    RESTARTS   AGE
demo-app-0   1/1     Running   0          16s
```

Execute the following command to generate some dummy data:

```console
kubectl -n demo-ns exec -it demo-app-0 -- /bin/sh -c \
    'dd if=/dev/zero of=/data/file.txt count=40 bs=1048576'
```

Verify that the data is generated:

```console
$ kubectl -n demo-ns exec -it demo-app-0 -- /bin/sh -c 'du -s /data'
40960   /data
```

### Step 3: Annotate pods

Annotate the pods with volumes attached to them with their volume names so that Velero takes backup
of volume data. Replace the pod name and the volume name as needed in the following command:

```bash
kubectl -n demo-ns annotate pod demo-app-0 backup.velero.io/backup-volumes=data
```

> **NOTE**: Modify pod template in Deployment `spec` or StatefulSet `spec` to always backup
> persistent volumes attached to them. This permanent setting will render this step unnecessary.

### Step 4: Backup entire namespace

Execute the following command to start the backup of the concerned namespace. In our demo
application's case it is `demo-ns`:

```bash
velero backup create backup-demo-app-ns --include-namespaces demo-ns --wait
```

Above operation may take some time, depending on the size of the data.

### Step 5: Restore Volumes

#### Same Cluster

If you plan to restore in the same cluster, then delete the namespace. In case of our demo
application run the following command:

```bash
kubectl delete ns demo-ns
```

> **NOTE**: If you are restoring a stateful component of Lokomotive like `prometheus-operator`, then
> delete the component namespace by running `kubectl delete ns monitoring`.

#### Different Cluster

In another cluster deploy components `rook`, `rook-ceph` and `velero` with the same configuration
for a successful restore.

#### Restore

Execute the following command to start the restore:

```bash
velero restore create --from-backup backup-demo-app-ns
```

Verify if Velero restored the application successfully:

```console
$ kubectl -n demo-ns get pods
NAME         READY   STATUS    RESTARTS   AGE
demo-app-0   1/1     Running   0          51s
```

> **NOTE**: If you are restoring a stateful component of Lokomotive like `prometheus-operator`, then
> once pods in `monitoring` namespace are in `Running` state, then run `lokoctl component apply
> prometheus-operator` to ensure the latest configs are applied.

Verify that the data is restored correctly:

```console
$ kubectl -n demo-ns exec -it demo-app-0 -- /bin/sh -c 'du -s /data'
40960   /data
```

## Additional resources

- Velero [Restic Docs](https://velero.io/docs/v1.4/restic/).
- Lokomotive `velero` component [configuration reference
  document](../configuration-reference/components/velero.md).
- Lokomotive `rook` component [configuration reference
  document](../configuration-reference/components/rook.md).
- Lokomotive `rook-ceph` component [configuration reference
  document](../configuration-reference/components/rook-ceph.md).
