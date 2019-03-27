# README

This module installs a Lokomotive cluster to Packet.


### Persistent Storage

The current approach of supporting persistent storage is a hack and a temporary workaround. Ideally we'd like to use a storage provider instead.

The only way to run apps that require persistent storage is to create a local storage class like:

```
kind: StorageClass
apiVersion: storage.k8s.io/v1
metadata:
  name: local-storage
provisioner: kubernetes.io/no-provisioner
volumeBindingMode: WaitForFirstConsumer
```

To support local storage class, we mount the host path `/mnt` of the worker nodes into the kubelet being run as a rkt pod. To use persistent storage, 

1. Create a directory under `/mnt` on the worker node(s). Eg: `/mnt/busybox`
2. Create `PersistentVolume`. For eg:

```
apiVersion: v1
kind: PersistentVolume
metadata:
  name: busybox
spec:
  capacity:
    storage: 1Gi
  claimRef:
    apiVersion: v1
    kind: PersistentVolumeClaim
    name: busybox-claim
    namespace: default
  accessModes:
  - ReadWriteOnce
  persistentVolumeReclaimPolicy: Retain
  storageClassName: local-storage
  local:
    path: /mnt/busybox
  nodeAffinity:
    required:
      nodeSelectorTerms:
      - matchExpressions:
        - key: kubernetes.io/hostname
          operator: In
          values:
          - your-worker-0
```

This will create a `PersistentVolume` named `busybox` that refers a `PersistentVolumeClaim` named `busybox-claim` with a storage capacity of `1Gi` (ensure that your node has this free space available). The volume will be mounted onto `/mnt/busybox` on the hostpath of the worker node named `your-worker-0`. Make sure this matches a worker node in your cluster. If you don't want to create this `PersistentVolume` in a specific worker node, skip the entire `nodeAffinity` section.


3. Create a `PersistentVolumeClaim`. For eg:

```
kind: PersistentVolumeClaim
apiVersion: v1
metadata:
  name: busybox
  namespace: default
spec:
  accessModes:
  - ReadWriteOnce
  storageClassName: local-storage
  resources:
    requests:
      storage: 1Gi
```

4. Deploy a workload that uses this `PersistentVolumeClaim`. For eg:

```
apiVersion: v1
kind: Pod
metadata:
  name: busybox
  namespace: default
spec:
  containers:
  - image: busybox
    imagePullPolicy: IfNotPresent
    name: busybox
    command: ["/bin/sh", "-c", "sleep 99999"]
    volumeMounts:
    - mountPath: /busybox
      name: test-volume
  volumes:
  - name: test-volume
    persistentVolumeClaim:
      claimName: busybox-0
  nodeSelector:
    kubernetes.io/hostname: your-worker-0

```

Once again, if you did not specify a `nodeAffinity` in your `PersistentVolume`, then skip the `nodeSelector` section from your `PodSpec`.
