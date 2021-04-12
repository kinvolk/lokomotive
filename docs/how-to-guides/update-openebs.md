---
title: Updating OpenEBS control plane and data plane components
weight: 10
---

## Introduction

This guide provides step-by-step instructions on how to update
[OpenEBS](https://openebs.io/) control plane and data plane components from one
version to another.

To get the current version installed, execute:
```bash
kubectl get pods -n openebs -o jsonpath='{.items[*].metadata.labels.openebs\.io/version}'
```

Refer to the [Lokomotive release notes](../../../CHANGELOG) to get the version
to update.

Make note of the current version installed and the version to update and set
the values accordingly in the document.

As an example, in this document we are going to update from `v2.2.0` to
`v2.6.0`.

## Prerequisites

To update the OpenEBS data plane components, we need the following:

* A Kubernetes cluster accessible via `kubectl`.

* [OpenEBS component](../../configuration-reference/components/openebs-operator)
  installed. You can check if OpenEBS is indeed in expected version:
    ```bash
    kubectl get pods -n openebs -l openebs.io/version=2.2.0
    ```

* Update process should not be disruptive, but it is recommended to schedule a
  downtime for the applications consuming the OpenEBS PV and make sure to take a
  backup of the data before starting the below update procedure in case some
  problem arises. Lokomotive provides
  [Velero](../../configuration-reference/components/velero) component for backup
  and restore.

* Ensure the cluster and OpenEBS volumes are is in healthy state before proceeding.
    ```bash
    $ lokoctl health
    Node                    Ready    Reason          Message

    alpha-controller-0      True     KubeletReady    kubelet is posting ready status
    alpha-large-worker-0    True     KubeletReady    kubelet is posting ready status
    Name      Status    Message              Error

    etcd-0    True      {"health":"true"}


   $ kubectl get cstorpools -n openebs # Status should be Healthy.

   NAME                                ALLOCATED   FREE   CAPACITY   STATUS    READONLY   TYPE      AGE
   cstor-pool-openebs-replica-1-w3r7   6.98G       437G   444G       Healthy   false      striped   4d22h


   $ kubectl get cstorvolume -n openebs # Status should be Healthy.

   NAME                                       STATUS    AGE     CAPACITY
   pvc-183dc3df-a7d9-4273-a8e2-7f66d2e19f4e   Healthy   3d20h   50Gi
   pvc-5d4b4c2b-2ed5-4e75-aeb9-fbd2918ebb78   Healthy   3d20h   50Gi
   ```
## Steps

Set the following environment variables in the terminal to assist in the update process:
```bash
export OPENEBS_OLD_VERSION=2.6.0
export OPENEBS_NEW_VERSION=2.7.0
```
### Step 1: Update OpenEBS control plane components

Lokomotive provides an easy way of updating the OpenEBS control plane.

Execute the following command to update OpenEBS control plane:

```bash
lokoctl component apply openebs-operator
```

Doing so, terminates the OpenEBS resources associated with the old version and new
resources are created for the new version.

Verify all the pods are in `Running` state before proceeding:

```bash
kubectl get pods -n openebs -l openebs.io/version=${OPENEBS_NEW_VERSION}
```

### Step 2: Update OpenEBS data plane components

OpenEBS control plane and data plane components work independently. Even if the
control plane components are updated to the new version, the data plane components
continue to work with the older version.

#### Update cStor pools

Create a Job to update the existing cStor pools:

```bash
cat > update-cstor-pools-from-${OPENEBS_OLD_VERSION}-to-${OPENEBS_NEW_VERSION}.yaml <<EOF
apiVersion: batch/v1
kind: Job
metadata:
  name: cstor-spc-${OPENEBS_OLD_VERSION}-to-${OPENEBS_NEW_VERSION}
  namespace: openebs
spec:
  backoffLimit: 4
  template:
    spec:
      serviceAccountName: openebs-operator
      containers:
      - name:  update
        args:
        - "cstor-spc"
        - "--v=4"
        - "--from-version=${OPENEBS_OLD_VERSION}"
        - "--to-version=${OPENEBS_NEW_VERSION}"
$(kubectl get spc -n openebs --no-headers -o custom-columns=":metadata.name" | sed 's/.*/        - "&"/')
        #DO NOT CHANGE BELOW PARAMETERS
        env:
        - name: OPENEBS_NAMESPACE
          valueFrom:
            fieldRef:
              fieldPath: metadata.namespace
        tty: true
        image: openebs/m-upgrade:${OPENEBS_NEW_VERSION}
        imagePullPolicy: Always
      restartPolicy: OnFailure
EOF
```

Apply the job to start the update for cStor pools:

```bash
kubectl apply -f update-cstor-pools-from-${OPENEBS_OLD_VERSION}-to-${OPENEBS_NEW_VERSION}.yaml
```

Ensure the job runs to completion:
```bash
kubectl describe job -n openebs cstor-spc-${OPENEBS_OLD_VERSION}-to-${OPENEBS_NEW_VERSION}

# Should see a similar outpu of 'Completed'
  Type    Reason            Age    From            Message
  ----    ------            ----   ----            -------
  Normal  SuccessfulCreate  2m35s  job-controller  Created pod: cstor-spc-2.6.0-to-2.7.0-hpcxl
  Normal  Completed         52s    job-controller  Job completed
```


#### Update cStor volumes

Create a Kubernetes job to update the existing cStor volumes:

```bash
cat > update-cstor-vols-from-${OPENEBS_OLD_VERSION}-to-${OPENEBS_NEW_VERSION}.yaml <<EOF
apiVersion: batch/v1
kind: Job
metadata:
  name: cstor-vol-${OPENEBS_OLD_VERSION}-to-${OPENEBS_NEW_VERSION}
  namespace: openebs
spec:
  backoffLimit: 4
  template:
    spec:
      serviceAccountName: openebs-operator
      containers:
      - name:  update
        args:
        - "cstor-volume"
        - "--v=4"
        - "--from-version=${OPENEBS_OLD_VERSION}"
        - "--to-version=${OPENEBS_NEW_VERSION}"
$(kubectl get cstorvolumes -n openebs --no-headers -o custom-columns=":metadata.name" | sed 's/.*/        - "&"/')
        #DO NOT CHANGE BELOW PARAMETERS
        env:
        - name: OPENEBS_NAMESPACE
          valueFrom:
            fieldRef:
              fieldPath: metadata.namespace
        tty: true
        image: openebs/m-upgrade:${OPENEBS_NEW_VERSION}
        imagePullPolicy: Always
      restartPolicy: OnFailure
EOF
```

```bash
kubectl apply -f update-cstor-vols-from-${OPENEBS_OLD_VERSION}-to-${OPENEBS_NEW_VERSION}.yaml
```

Ensure the job runs to completion:
```bash
kubectl describe job -n openebs cstor-vol-${OPENEBS_OLD_VERSION}-to-${OPENEBS_NEW_VERSION}

# Should see a similar outpu of 'Completed'
  Type    Reason            Age    From            Message
  ----    ------            ----   ----            -------
  Normal  SuccessfulCreate  2m35s  job-controller  Created pod: cstor-vol-2.6.0-to-2.7.0-gtwsd
  Normal  Completed         1m36s  job-controller  Job completed
```

### Step 3: Verify

To check if the update process was successful, execute:

```bash
$ kubectl get pods -n openebs -l openebs.io/version=${OPENEBS_NEW_VERSION}
NAME                                                              READY   STATUS    RESTARTS   AGE
cstor-pool-openebs-replica-3-6qxo-bd59f5554-d97l4                 3/3     Running   0          27m
cstor-pool-openebs-replica-3-ag9v-78cfbb64b4-vzcs9                3/3     Running   0          26m
cstor-pool-openebs-replica-3-gt7g-8ddd9b457-jqwz4                 3/3     Running   0          25m
openebs-operator-admission-server-6b5fb6dff5-6dc8d                1/1     Running   2          34m
openebs-operator-apiserver-5c467bc588-fsl58                       1/1     Running   0          34m
openebs-operator-localpv-provisioner-74d76d55b-s8zwv              1/1     Running   0          33m
openebs-operator-ndm-cfnxp                                        1/1     Running   0          34m
openebs-operator-ndm-nsqkg                                        1/1     Running   0          33m
openebs-operator-ndm-operator-758fdbc5f4-5qqxx                    1/1     Running   0          34m
openebs-operator-ndm-x44d8                                        1/1     Running   0          34m
openebs-operator-provisioner-59c6dc5dfc-q9k65                     1/1     Running   0          34m
openebs-operator-snapshot-operator-bf49c5dc6-xqqlq                2/2     Running   0          34m
pvc-5b7d8ee3-59b8-4a02-bd6d-a4e66fbecf9f-target-68984f69b7l42sl   3/3     Running   0          22m
pvc-e8f8b268-b81c-4201-841a-283f557c44a7-target-7484b968c4pzld6   3/3     Running   0          20m
```

To check if all the `StoragePoolClaims` and `CStorVolumes` have been updated,
execute:

```bash
kubectl get pods -n openebs -l openebs.io/version=${OPENEBS_OLD_VERSION}
```
No output should be displayed.

## Summary

This guide helps to update the OpenEBS control plane and data plane components.

## Additional resources

[OpenEBS troubleshooting guide](https://docs.openebs.io/docs/next/troubleshooting.html).

For additional information regarding the update steps, see [OpenEBS update
documentation](https://github.com/openebs/openebs/blob/master/k8s/upgrades/README.md).
