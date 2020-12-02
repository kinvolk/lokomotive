---
title: Upgrading Rook Ceph
weight: 10
---

## Introduction

[Rook Ceph](../configuration-reference/components/rook.md) is one of the storage providers of
Lokomotive. With a distributed system as complex as Ceph, the upgrade process is not trivial. This
document enlists steps on how to perform the upgrade and how to monitor this process.

## Prerequisites

- A Lokomotive cluster accessible via `kubectl`.

- Upgrade process should not be disruptive, but it is recommended to schedule a downtime for the
  applications consuming the Rook Ceph PVs and make sure to take a backup of the data before
  starting the below upgrade procedure in case some problem arises. Read more about the [backup
  process](backup-rook-ceph-volumes.md).

## Steps

The following steps are inspired by the [`rook`](https://rook.io/docs/rook/master/ceph-upgrade.html)
docs.

### Step 1: Ensure `AUTOSCALE` is set to `on`

Start a shell in the toolbox pod as specified in [this
doc](rook-ceph-storage.md#enable-and-access-toolbox) and run the following command:

```console
# ceph osd pool autoscale-status | grep replicapool
POOL           SIZE  TARGET SIZE  RATE  RAW CAPACITY   RATIO  TARGET RATIO  EFFECTIVE RATIO  BIAS  PG_NUM  NEW PG_NUM  AUTOSCALE
replicapool      0                 3.0         3241G  0.0000                                  1.0      32              on
```

Ensure that the `AUTOSCALE` column outputs `on` and not `warn`. This should always be `on` but it's
especially important for upgrades. If the output of the `AUTOSCALE` column says `warn`, then run the
command below to make sure that pool autoscaling is enabled. It is required to ensure that the
placement groups scale up as the data in the cluster increases.

```bash
ceph osd pool set replicapool pg_autoscale_mode on
```

### Step 2: Watch

Watch events, updates and pods.

#### Step 2.1: Ceph status

Leave the following running in the toolbox pod:

```bash
watch ceph status
```

Ensure that the output says that `health:` is `HEALTH_OK`. Match the output such that everything
looks as explained in the [rook upgrade
docs](https://rook.io/docs/rook/master/ceph-upgrade.html#status-output).

> **IMPORTANT**: Don't proceed further if the output is anything other than `HEALTH_OK`.

During the ongoing upgrade and after completion, the output should stay in `HEALTH_OK` state,
although if the cluster is more than 60% full, it can sometimes turn into `HEALTH_WARN` temporarily.

#### Step 2.2: Pods in rook namespace

Open another terminal window and keep an eye on the `STATUS` field of the following output. Make
sure that the pods are restarted correctly and don't go into `CrashLoopBackOff` state. Leave the
following command running:

```bash
watch kubectl -n rook get pods -o wide
```

#### Step 2.3: Rook version update

Run the following command in a new terminal window to keep an eye on the rook version as it is
updated for all the sub-components:

```bash
watch --exec kubectl -n rook get deployments -l rook_cluster=rook -o \
  jsonpath='{range .items[*]}{.metadata.name}{"  \treq/upd/avl: "}{.spec.replicas}{"/"}{.status.updatedReplicas}{"/"}{.status.readyReplicas}{"  \trook-version="}{.metadata.labels.rook-version}{"\n"}{end}'
```

```bash
watch --exec kubectl -n rook get jobs -o \
  jsonpath='{range .items[*]}{.metadata.name}{"  \tsucceeded: "}{.status.succeeded}{"      \trook-version="}{.metadata.labels.rook-version}{"\n"}{end}'
```

You should see that `rook-version` slowly changes to `v1.4.6`.

#### Step 2.4: Ceph version update

Run the following command to keep an eye on the Ceph version update as the new pods come up in a new
terminal window:

```bash
watch --exec kubectl -n rook get deployments -l rook_cluster=rook -o \
  jsonpath='{range .items[*]}{.metadata.name}{"  \treq/upd/avl: "}{.spec.replicas}{"/"}{.status.updatedReplicas}{"/"}{.status.readyReplicas}{"  \tceph-version="}{.metadata.labels.ceph-version}{"\n"}{end}'
```

You should see that `ceph-version` slowly changes to `15.2.5`.

#### Step 2.5: Events in rook namespace

In a new terminal leave the following command running, to keep track of the events happening in the
`rook` namespace. Keep an eye on the column `TYPE` of the following output and especially events
that are not of type `Normal`.

```bash
kubectl -n rook get events -w
```

### Step 3: Dashboards

Monitor various dashboards.

#### Step 3.1: Ceph

Open the Ceph dashboard in a browser window. Read the docs
[here](rook-ceph-storage.md#access-the-ceph-dashboard) to access the dashboard.

> **NOTE**: Accessing the dashboard can be a hassle because while the components are upgrading you
> may lose access to it multiple times.

#### Step 3.2: Grafana

Gain access to the Grafana dashboard as instructed
[here](monitoring-with-prometheus-operator.md#access-grafana). And keep an eye on the dashboard
named `Ceph - Cluster`.

> **NOTE**: The data in the Grafana dashboard will always be outdated compared to the `watch ceph
> status` running inside the toolbox pod.

### Step 4: Make a note of existing image versions

Make a note of the images of the pods in the rook namespace:

```bash
kubectl -n rook get pod -o \
  jsonpath='{range .items[*]}{.metadata.name}{"\n\t"}{.status.phase}{"\t\t"}{.spec.containers[0].image}{"\t"}{.spec.initContainers[0].image}{"\n\n"}{end}'
```

After the upgrade is complete, we can verify the output of the above command to see if the workloads
now run updated images.

### Step 5: Perform updates

With everything monitored, you can start the update process now by executing the following commands:

```bash
kubectl apply -f https://raw.githubusercontent.com/kinvolk/lokomotive/v0.5.0/assets/charts/components/rook/templates/resources.yaml
lokoctl component apply rook rook-ceph
```

### Step 6: Verify that the CSI images are updated

Verify if the images were updated, comparing it with the output of the [Step
4](#step-4-make-a-note-of-existing-image-versions).

```bash
kubectl -n rook get pod -o \
  jsonpath='{range .items[*]}{.metadata.name}{"\n\t"}{.status.phase}{"\t\t"}{.spec.containers[0].image}{"\t"}{.spec.initContainers[0].image}{"\n\n"}{end}'
```

### Step 7: Final checks

Once everything is up to date, then run the following commands in the toolbox pod, to verify if all
the OSDs are in `up` state:

```bash
ceph osd status
```

## Additional resources

- [Rook Upgrade docs](https://rook.io/docs/rook/v1.4/ceph-upgrade.html).
- [General Troubleshooting](https://rook.io/docs/rook/v1.5/common-issues.html).
- [Ceph Troubleshooting](https://rook.io/docs/rook/v1.4/ceph-common-issues.html).
