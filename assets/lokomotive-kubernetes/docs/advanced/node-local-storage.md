# Using Node Local Storage on Kubernetes

This guide shows different ways to use node local storage on Lokomotive, to achieve near
bare-metal storage performance inside Kubernetes pods.

Before you continue, make sure to read generalities about these concepts:
 * [Volumes](https://kubernetes.io/docs/concepts/storage/volumes/)
 * [Persistent Volumes and PVC](https://kubernetes.io/docs/concepts/storage/persistent-volumes/)
 * [StorageClasses](https://kubernetes.io/docs/concepts/storage/storage-classes/)

More details about specific volume types used and how to use them on Lokomotive
will follow, so there is no need to look into volume specifics on the first
read.

**NOTE**: This has been tested on Packet provider only, some Lokomotive changes may need to be
ported to other providers for this to work.

## Local Persistent Volumes in Filesystem Mode

This way of using node local storage works by exposing a filesystem mounted on
the host as a Persistent Volume to pods.

A single Persistent Volume can be bound to only one PVC in ReadWriteOnce access
mode, so only one pod will be able to use one Persistent Volume in the host.

See the section below to see how to create more than one Persistent
Volume per node.

The Kubernetes documentation on Local Persistent Volumes using filesystem mode is
[here](https://kubernetes.io/docs/concepts/storage/volumes/#local) (mixed with
block mode).

Note that this section is about the filesystem mode, not the block mode.
Information about the block mode will be covered on a different section of this document

### Example

In this example, two nodes with Local Persistent Volumes will be created and used
by two Stateful Set pods.

Create a Packet cluster and make sure to create a worker pool that has
`setup_raid = "true"`. For example (note that this example is not
auto-contained, as local variables are not defined):

```yaml

module "worker-pool" {
  source = "git::ssh://git@github.com/kinvolk/lokomotive-kubernetes.git//packet/flatcar-linux/kubernetes/workers?ref=546e48004d76cb7d99c8211f04a170e20e918fe9"

  providers = {
    local    = "local.default"
    template = "template.default"
    tls      = "tls.default"
    packet   = "packet.default"
  }

  ssh_keys     = "${local.ssh_keys}"
  cluster_name = "${local.cluster_name}"
  project_id   = "${local.project_id}"
  facility     = "${local.facility}"

  pool_name = "storage"

  count = 2

  type = "c1.small.x86"

  # Make sure this is set on your workers
  setup_raid = "true"

  ipxe_script_url = "https://raw.githubusercontent.com/kinvolk/flatcar-ipxe-scripts/no-https/packet.ipxe"
  kubeconfig      = "${module.controller.kubeconfig}"
}
```

For the rest of the example it is assumed the nodes names are:
`poc-storage-worker-0` and `poc-storage-worker-1`.

Create a Storage Class and a Local Persistent Volume exposing the mounted
filesystem on each node:

```yaml
kind: StorageClass
apiVersion: storage.k8s.io/v1
metadata:
  name: local-storage
provisioner: kubernetes.io/no-provisioner
volumeBindingMode: WaitForFirstConsumer
---
apiVersion: v1
kind: PersistentVolume
metadata:
  name: worker-0-pv
spec:
  capacity:
    storage: 10Gi
  volumeMode: Filesystem
  accessModes:
  - ReadWriteOnce
  persistentVolumeReclaimPolicy: Retain
  storageClassName: local-storage
  local:
    path: /mnt
  nodeAffinity:
    required:
      nodeSelectorTerms:
      - matchExpressions:
        - key: kubernetes.io/hostname
          operator: In
          values:
          - poc-storage-worker-0
---
apiVersion: v1
kind: PersistentVolume
metadata:
  name: worker-1-pv
spec:
  capacity:
    storage: 10Gi
  volumeMode: Filesystem
  accessModes:
  - ReadWriteOnce
  persistentVolumeReclaimPolicy: Retain
  storageClassName: local-storage
  local:
    path: /mnt
  nodeAffinity:
    required:
      nodeSelectorTerms:
      - matchExpressions:
        - key: kubernetes.io/hostname
          operator: In
          values:
          - poc-storage-worker-1
```

The PV should be available:

```sh
$ kubectl get pv
NAME          CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS      CLAIM   STORAGECLASS    REASON   AGE
worker-0-pv   10Gi       RWO            Retain           Available           local-storage            95m
worker-1-pv   10Gi       RWO            Retain           Available           local-storage            95m
```

As no PVC was created yet, you should also see no PVC:

```sh
$ kubectl get pvc
No resources found.
```

Create a simple Stateful Set that will use the volumes (via a PVC):

```yaml
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: local-test
spec:
  serviceName: "local-service"
  replicas: 2
  selector:
    matchLabels:
      app: local-test
  template:
    metadata:
      labels:
        app: local-test
    spec:
      securityContext:
        runAsNonRoot: true
        runAsUser: 65534
      containers:
      - name: test-container
        image: k8s.gcr.io/busybox
        command:
        - "/bin/sh"
        args:
        - "-c"
        - "sleep 100000"
        volumeMounts:
        - name: local-vol
          mountPath: /usr/test-pod
  volumeClaimTemplates:
  - metadata:
      name: local-vol
    spec:
      accessModes: [ "ReadWriteOnce" ]
      storageClassName: "local-storage"
      resources:
        requests:
          storage: 1Gi
```

The PVC should be listed and the PV bound:

```sh
$ kubectl get pods
NAME           READY   STATUS    RESTARTS   AGE
local-test-0   1/1     Running   0          54s
local-test-1   1/1     Running   0          48s

$ kubectl get pvc
NAME                     STATUS   VOLUME        CAPACITY   ACCESS MODES   STORAGECLASS    AGE
local-vol-local-test-0   Bound    worker-1-pv   10Gi       RWO            local-storage   7s
local-vol-local-test-1   Bound    worker-0-pv   10Gi       RWO            local-storage   1s

$ kubectl get pv
NAME          CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                            STORAGECLASS    REASON   AGE
worker-0-pv   10Gi       RWO            Retain           Bound    default/local-vol-local-test-1   local-storage            97m
worker-1-pv   10Gi       RWO            Retain           Bound    default/local-vol-local-test-0   local-storage            97m
```

You can also connect to the pod and write to the volume which is mounted on
`/usr/test-pod`, you will see in the host on `/mnt`.

If you delete a pod, as it is part of this Stateful Set, it will be recreated and
continue to use the volume.

**What if you want to delete this application and use the PVs for a different
application?**

Basically, as the retain policy is set to `Retain` (just to be safe), you need
to do some manual operations documented [here][1]. Note that when using the `Delete`
policy this is different, so you may want to choose a policy that better fits
your use case.

[1]: https://kubernetes.io/docs/concepts/storage/persistent-volumes/#retain

### Automatic Local Persistent Volume Creation and Handling

The example manually created the `Kind: PersistentVolume` for each node. Also,
the PV will continue to exist even if the PVC is deleted and, therefore, the
lifecycle of the PV needs to be handled manually.

Those limitation don't scale well with the number of nodes and applications,
for that reason an external provisioner that will automatically create those
objects when nodes are available and manage the lifecycle of the volume was
created by sig-storage:
	https://github.com/kubernetes-sigs/sig-storage-local-static-provisioner/

### Limitations

There are some limitations you might want to have in mind:

 * Lokomotive Kubernetes currently only exposes `/mnt` to the kubelet, so only that
directory can be used for node local storage.

 * Using `setup_raid = "true"` will create a single RAID 0 array with all
the spare disks and mount it in `/mnt`. Therefore all the
node storage will be available to only one PV, thus it will be available to only
one PVC and one pod.  This is the only automated setup supported in Lokomotive
at the time of writing.

 * When using `setup_raid = "true"` the filesystem is `ext4` and there is no
way to use different one right now.

### Operation and FAQs

* [Recommendations and tips on how to operate][1]. Bear in mind, however, that the provisioner is not being used in Lokomotive
right now.

 * You may also want to check out the [FAQs][2].

[1]: https://github.com/kubernetes-sigs/sig-storage-local-static-provisioner/blob/master/docs/operations.md
[2]: https://github.com/kubernetes-sigs/sig-storage-local-static-provisioner/blob/master/docs/faqs.md

### Useful Links

There are several other useful links:

 * [Best practices][1]. Lokomotive is following them at the time of this writing, except for using UUID in fstab.

 * More documentation about [access control][2] that might be useful for several apps sharing a volume.

[1]: https://github.com/kubernetes-sigs/sig-storage-local-static-provisioner/blob/master/docs/best-practices.md.
[2]: https://kubernetes.io/docs/tasks/configure-pod-container/configure-persistent-volume-storage/#access-control

Furthermore, in [this blog post][3] there are some highlights not well
documented anywhere else and some useful tips from Uber's experience when
operating Local Persistent Volumes.

[3]: https://kubernetes.io/blog/2019/04/04/kubernetes-1.14-local-persistent-volumes-ga/

Here is a short copy paste of some highlights, in case the blog post is lost in
the future.

> When first testing local volumes, we wanted to have a thorough understanding of
> the effect disruptions (voluntary and involuntary) would have on pods using
> local storage, and so we began testing some failure scenarios. We found that
> when a local volume becomes unavailable while the node remains available (such
> as when performing maintenance on the disk), a pod using the local volume will
> be stuck in a ContainerCreating state until it can mount the volume. If a node
> becomes unavailable, for example if it is removed from the cluster or is
> drained, then pods using local volumes on that node are stuck in an Unknown or
> Pending state depending on whether or not the node was removed gracefully.

> Recovering pods from these interim states means having to delete the PVC binding
> the pod to its local volume and then delete the pod in order for it to be
> rescheduled (or wait until the node and disk are available again). We took this
> into account when building our operator for M3DB, which makes changes to the
> cluster topology when a pod is rescheduled such that the new one gracefully
> streams data from the remaining two peers. Eventually we plan to automate the
> deletion and rescheduling process entirely.
>
> Alerts on pod states can help call attention to stuck local volumes, and
> workload-specific controllers or operators can remediate them automatically.
> Because of these constraints, it’s best to exclude nodes with local volumes from
> automatic upgrades or repairs, and in fact some cloud providers explicitly
> mention this as a best practice.

And some information of the current pain points that might be handled better in
the future:

> One of the most frequent asks has been for a controller that can help with
> recovery from failed nodes or disks, which is currently a manual process (or
> something that has to be built into an operator). SIG Storage is investigating
> creating a common controller that can be used by workloads with simple and
> similar recovery processes.
>
> Another popular ask has been to support dynamic provisioning using lvm. This can
> simplify disk management, and improve disk utilization. SIG Storage is
> evaluating the performance tradeoffs for the viability of this feature.


## Local Persistent Volumes in Raw Block Device Mode

This mode of using node local storage has not been explored so far, but it's
listed so it can be explored in the future.

Some links that might be relevant when doing so:

- https://kubernetes.io/blog/2019/03/07/raw-block-volume-support-to-beta/
- https://kubernetes.io/docs/concepts/storage/volumes/#local
- https://docs.okd.io/latest/install_config/persistent_storage/persistent_storage_local.html
- https://docs.okd.io/latest/install_config/configuring_local.html#install-config-configuring-local
- https://docs.okd.io/latest/install_config/configuring_local.html#local-volume-raw-block-devices


## Using OpenEBS Local PV

OpenEBS has Local PV engine, which can be used for consuming node local storage.

To set it up, follow OpenEBS documentation:

- https://docs.openebs.io/docs/next/localpv.html
- https://blog.openebs.io/preview-dynamic-provisioning-of-kubernetes-local-pvs-using-openebs-a530c25cf13d
- https://docs.openebs.io/docs/next/uglocalpv.html


## Further investigation

Other container attached storage (CAS) solutions, like [Rook](https://rook.io/),
might be considered for consuming node local storage as well.
