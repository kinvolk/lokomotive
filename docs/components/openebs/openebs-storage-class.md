# OpenEBS Storage Class

This component configures the storage class and storage pool claim for OpenEBS.

## Requirements
- Openebs operator installed and in running state.

## Argument Reference

The openebs-storage-class component supports creation of multiple storage classes. The component supports the following parameters listed below

| Argument        | Explanation                                                                                                                   | Default | Required |
|-----------------|-------------------------------------------------------------------------------------------------------------------------------|---------|----------|
| `replica_count` | Defines the number of cStor volume replicas                                                                                   | 3       | false    |
| `default`       | Indicates whether the storage class is default or not                                                                         | false   | false    |
| `disks`         | List of selected unclaimed blockDevice CRs which are unmounted and does not contain a filesystem in each participating nodes. | -       | false    |

## Example Configurations

For a default configuration, no configuration file is needed.

The below configuration is treated as default
```hcl
component "openebs-storage-class" {
  storage-class "openebs-cstor-disk-replica-3" {
    replica_count = 3
    default = true
  }
}

or

component "openebs-storage-class" {}
```
For additional configuration or multiple storage classes, follow the example config below.

OpenEBS supports creation of a cStorPool even when diskList is not specified.

```hcl
# cluster-openebs-storage-class.lokocfg
component "openebs-storage-class" {
  storage-class "openebs-replica1" {
    replica_count = 1
  }
  storage-class "openebs-replica3" {
    replica_count = 3
    default = true
  }
}
```

One can also provide the disks which should be used by a storage class.

To get the name of the disks

```bash
✗ kubectl get blockdevices -n openebs
NAME                                           NODENAME               SIZE            CLAIMSTATE   STATUS   AGE
blockdevice-0565dd2d566cab012b7bc35e54874d9f   node-pool-1-worker-0   480103981056    Unclaimed    Active   16h
blockdevice-15e9fe2adab21465e6ccf5a649eb27ff   node-pool-1-worker-0   2000398934016   Unclaimed    Active   16h
blockdevice-17901367ccd9e1ead797a7e233de8cc8   node-pool-1-worker-1   2000398934016   Unclaimed    Active   16h
blockdevice-1cf72b0fed59c0c1a6767baca41422a0   node-pool-1-worker-0   2000398934016   Unclaimed    Active   16h
blockdevice-1f4315cb4acbb4b0dbf5202adcdb70d8   node-pool-1-worker-2   2000398934016   Unclaimed    Active   16h
blockdevice-219ed0a0878c8a4517ef52c54b9f5b54   node-pool-1-worker-2   2000398934016   Unclaimed    Active   16h
```

Example configuration with disks provided

```hcl
# cluster-openebs-storage-class.lokocfg
component "openebs-storage-class" {
  storage-class "openebs-replica-3" {
    replica_count = 3
    default = true
    disks = [
      "blockdevice-0565dd2d566cab012b7bc35e54874d9f",
      "blockdevice-17901367ccd9e1ead797a7e233de8cc8",
      "blockdevice-1f4315cb4acbb4b0dbf5202adcdb70d8"
    ]
  }
}

```
## Installation

Installation of the component is straightforward.

```bash
✗ ./lokoctl component install openebs-storage-class
Waiting for api-server...
Creating assets...
Created  StoragePoolClaim cstor-pool-openebs-cstor-disk-replica-3
Created  StorageClass openebs-cstor-disk-replica-3
```
This creates a storage pool and a storage class, based on the above disk based ocnfiguration.

**Verify**

```
✗ kubectl get storageclass
NAME                                     PROVISIONER                                                AGE
openebs-cstor-disk-replica-3 (default)   openebs.io/provisioner-iscsi                               2m23s
openebs-device                           openebs.io/local                                           4m58s
openebs-hostpath                         openebs.io/local                                           4m58s
openebs-jiva-default                     openebs.io/provisioner-iscsi                               4m58s
openebs-snapshot-promoter                volumesnapshot.external-storage.k8s.io/snapshot-promoter   4m58s

✗ kubectl get blockdevices -n openebs
NAME                                           NODENAME               SIZE            CLAIMSTATE   STATUS   AGE
blockdevice-0565dd2d566cab012b7bc35e54874d9f   node-pool-1-worker-0   480103981056    Claimed      Active   16h
blockdevice-15e9fe2adab21465e6ccf5a649eb27ff   node-pool-1-worker-0   2000398934016   Unclaimed    Active   16h
blockdevice-17901367ccd9e1ead797a7e233de8cc8   node-pool-1-worker-1   2000398934016   Claimed      Active   16h
blockdevice-1cf72b0fed59c0c1a6767baca41422a0   node-pool-1-worker-0   2000398934016   Unclaimed    Active   16h
blockdevice-1f4315cb4acbb4b0dbf5202adcdb70d8   node-pool-1-worker-2   2000398934016   Claimed      Active   16h
blockdevice-219ed0a0878c8a4517ef52c54b9f5b54   node-pool-1-worker-2   2000398934016   Unclaimed    Active   16h

✗ kubectl get cstorpools
NAME                                           ALLOCATED   FREE    CAPACITY   STATUS    TYPE      AGE
cstor-pool-openebs-cstor-disk-replica-3-lfr6   968K        444G    444G       Healthy   striped   2m
cstor-pool-openebs-cstor-disk-replica-3-mip9   1018K       1.81T   1.81T      Healthy   striped   2m
cstor-pool-openebs-cstor-disk-replica-3-yd8y   3.65M       1.81T   1.81T      Healthy   striped   2m

✗ kubectl get storagepoolclaims
NAME                                      AGE
cstor-pool-openebs-cstor-disk-replica-3   3m

✗ kubectl get storagepools
NAME                     AGE
default                  3m
```

### Running demo workload to test
```
✗ kubectl apply -f manifests/demo-pvc.yaml
persistentvolumeclaim/fio-cstor-claim created

✗ kubectl get persistentvolumeclaims
NAME              STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS                   AGE
fio-cstor-claim   Bound    pvc-4b2e5f35-610f-11e9-818a-ec0d9a9eadd6   4G         RWO            openebs-cstor-disk-replica-3   48s

✗ kubectl apply -f manifests/demo-fio-cstor-workload.yaml
pod/fio-cstor created

✗ kubectl get pods
NAME        READY   STATUS    RESTARTS   AGE
fio-cstor   1/1     Running   0          71s
```

**Verify**
```
✗ kubectl exec -it fio-cstor bash

root@fio-cstor:/# cd /datadir/

root@fio-cstor:/datadir# ls
basic.test.file  lost+found

root@fio-cstor:/datadir# echo "RANDOM TEST FOR STORAGE" > kosyfile.txt

root@fio-cstor:/datadir# cat kosyfile.txt
RANDOM TEST FOR STORAGE

root@fio-cstor:/datadir# exit
exit

✗ kubectl get pods
NAME        READY   STATUS    RESTARTS   AGE
fio-cstor   1/1     Running   0          7m3s

✗ kubectl delete pods fio-cstor
pod "fio-cstor" deleted

✗ kubectl apply -f cstor/demo-fio-cstor-workload.yaml
pod/fio-cstor created
persistentvolumeclaim/fio-cstor-claim unchanged

✗ kubectl get pods
NAME        READY   STATUS    RESTARTS   AGE
fio-cstor   1/1     Running   0          18s

✗ kubectl exec -it fio-cstor bash

root@fio-cstor:/# ls /datadir/
basic.test.file  kosyfile.txt  lost+found

root@fio-cstor:/# cat /datadir/kosyfile.txt
RANDOM TEST FOR STORAGE
```
