# OpenEBS

OpenEBS has many components, which can be grouped into the following categories.

- **Control plane components** - Provisioner, API Server, volume exports, and volume sidecars

- **Data plane components** - Jiva and cStor

- **Node disk manager** - Discover, monitor, and manage the media attached to the Kubernetes node

- **Integrations with cloud-native tools** - Integrations are done with Prometheus, Grafana, Fluentd, and Jaeger

According to the [docs](https://docs.openebs.io/docs/next/cstor.html), cStor is the recommended storage engine in OpenEBS.

## Requirements
- At least 3 workers with available disks.
- iSCSI client configured and iscsid service running on the worker nodes. In our current setup, we have iscsid service automatically enabled and running on all worker nodes. 

**NOTE:** OpenEBS requires available disks, i.e. disks that aren't mounted by anything. This means that by default, OpenEBS will not work on machines with just a single physical disk, e.g. Packet's t1.small.x86 (because the disk will be used for the operating system).

### Setup nodeSelectors for Node Disk Manager (NDM)
**This is an optional step**

If you want to consider only some nodes in Kubernetes cluster to be used for OpenEBS storage (for hosting cStor Storage Pool instances), then do the following to use nodeSelector field of NDM PodSpec and dedicate those nodes to NDM.

First, label the required nodes with an appropriate label. In the following example command, the required nodes for storage nodes are labelled as `node=openebs`.
```
kubectl label nodes <node-name> node=openebs
```

Then, add the following to `.lokocfg` file, modifying the values as appropriate.
```
component "openebs-operator" {
	ndm_selector_label = "node"
	ndm_selector_value = "openebs"
}
```

### Installation
```bash
 ✗ ./lokoctl component install openebs-operator                           
Waiting for api-server...
Creating assets...
Created  Namespace openebs
Created  ServiceAccount openebs/openebs-maya-operator
Created  ClusterRole openebs-maya-operator
Created  ClusterRoleBinding openebs-maya-operator
Created  Deployment openebs/maya-apiserver
Created  Service openebs/maya-apiserver-service
Created  Deployment openebs/openebs-provisioner
Created  Deployment openebs/openebs-snapshot-operator
Created  ConfigMap openebs/openebs-ndm-config
Created  DaemonSet openebs/openebs-ndm
```

**Verify**

After a while, all the OpenEBS components should be done creating and in running state.
Please ensure to verify that everything works properly before next steps.
```
✗ kubectl get pods -n openebs
NAME                                        READY   STATUS    RESTARTS   AGE
cstor-sparse-pool-4cif-7cbc757dc8-r9nsq     2/2     Running   0          25s
cstor-sparse-pool-58nd-84c9f66786-tflg9     2/2     Running   0          25s
cstor-sparse-pool-w6kv-7889bf9755-mnfd9     2/2     Running   0          25s
maya-apiserver-54f87dcb7b-7gxsd             1/1     Running   0          60s
openebs-ndm-4t7ws                           1/1     Running   0          57s
openebs-ndm-jwxwf                           1/1     Running   0          57s
openebs-ndm-l2v9s                           1/1     Running   0          57s
openebs-provisioner-57554c764-dvffn         1/1     Running   0          58s
openebs-snapshot-operator-96464fd9d-pscdk   2/2     Running   0          58s
```
If you initially setup nodeSelectors for Node Disk Manager(NDM), you should see that pods are only scheduled on labelled nodes with ` kubectl get pods -n openebs -o wide`.

```
✗ kubectl get storageclass
NAME                        PROVISIONER                                                AGE
openebs-cstor-sparse        openebs.io/provisioner-iscsi                               2m58s
openebs-jiva-default        openebs.io/provisioner-iscsi                               2m58s
openebs-snapshot-promoter   volumesnapshot.external-storage.k8s.io/snapshot-promoter   2m58s

✗ kubectl get disk
NAME                                      SIZE           STATUS   AGE
disk-344d7dfe24baa1fdd554f8c2071653e6     240057409536   Active   3m
disk-3999e2a03f5704b2b61ff22986480078     480103981056   Active   3m
disk-4ec248b90328f89f907e0cfeb7557808     480103981056   Active   3m
disk-74a39bb7df08013b8eb1c2cd2254b29c     480103981056   Active   3m
disk-8d5346546ab17be83c53ca8107436cfe     240057409536   Active   3m
disk-9061c7de65d0675484d74d983521494d     480103981056   Active   3m
disk-ba89435750dfad96ade5412d12834705     480103981056   Active   3m
disk-f7da3ed775a6c70b9ae5c55e6b3f379b     480103981056   Active   3m
disk-fc0852f715cd521535595e4431b93d8c     240057409536   Active   3m
sparse-272d30ebb237ec76969455adf0633d16   10737418240    Active   3m
sparse-5b7d76d2b64f2a787351b8dc5eeb262a   10737418240    Active   3m
sparse-e699b48ea97a8a36faa17c29a3411b04   10737418240    Active   3m

✗ kubectl get cstorpools
NAME                     ALLOCATED   FREE    CAPACITY   STATUS    TYPE      AGE
cstor-sparse-pool-4cif   270K        9.94G   9.94G      Healthy   striped   2m
cstor-sparse-pool-58nd   80K         9.94G   9.94G      Healthy   striped   2m
cstor-sparse-pool-w6kv   270K        9.94G   9.94G      Healthy   striped   2m

✗ kubectl get storagepoolclaims
NAME                AGE
cstor-sparse-pool   3m

✗ kubectl get storagepools
NAME                     AGE
cstor-sparse-pool-4cif   3m
cstor-sparse-pool-58nd   3m
cstor-sparse-pool-w6kv   3m
default                  3m
```

### Create storage pool and storage class using physical disks
OpenEBS supports creation of a cStorPool even when diskList is not specified in the YAML specification. In this case, one pool instance on each node is created with just one striped disk. Here is the link to the [docs](https://docs.openebs.io/docs/next/configurepools.html#auto-mode) for more information on how to customise the pool creation and storage class according to your needs.

```
✗ ./lokoctl component install openebs-default-storage-class
Waiting for api-server...
Creating assets...
Created  StoragePoolClaim cstor-pool
Created  StorageClass openebs-cstor-disk
```
This creates a storage pool and a storage class, and makes the installed storage class the default.

**Verify**
```
✗ kubectl get cstorpools                                            
NAME                                ALLOCATED   FREE    CAPACITY   STATUS    TYPE      AGE
cstor-pool-dog7                     83K         444G    444G       Healthy   striped   112s
cstor-pool-mdd5                     83K         444G    444G       Healthy   striped   112s
cstor-pool-rv2s                     83K         444G    444G       Healthy   striped   112s
cstor-sparse-pool-4fvc              626K        9.94G   9.94G      Healthy   striped   69m
cstor-sparse-pool-h8mn              623K        9.94G   9.94G      Healthy   striped   69m
cstor-sparse-pool-kzf2              554K        9.94G   9.94G      Healthy   striped   69m

✗ kubectl get pods -n openebs
NAME                                                 READY   STATUS    RESTARTS   AGE
cstor-pool-dog7-7d6fbdbb98-drp78                     2/2     Running   1          2m23s
cstor-pool-mdd5-5b9d866d5f-4clxt                     2/2     Running   1          2m23s
cstor-pool-rv2s-f5db654d9-tg2m9                      2/2     Running   1          2m23s
cstor-sparse-pool-4fvc-5959f74896-n9j7z              2/2     Running   0          69m
cstor-sparse-pool-h8mn-7f66b74bf-gt4pz               2/2     Running   0          69m
cstor-sparse-pool-kzf2-6cb995bd7c-vth9d              2/2     Running   0          69m
maya-apiserver-b69f79645-l7sgq                       1/1     Running   0          70m
openebs-ndm-bg657                                    1/1     Running   0          70m
openebs-ndm-bsh7p                                    1/1     Running   0          70m
openebs-ndm-f4vcs                                    1/1     Running   0          70m
openebs-provisioner-cfb9564d8-pmwgp                  1/1     Running   0          70m
openebs-snapshot-operator-67678bb6b5-zcvqm           2/2     Running   0          70m

✗ kubectl get storageclass
NAME                           PROVISIONER                                                AGE
openebs-cstor-disk (default)   openebs.io/provisioner-iscsi                               2m52s
openebs-cstor-sparse           openebs.io/provisioner-iscsi                               70m
openebs-jiva-default           openebs.io/provisioner-iscsi                               70m
openebs-snapshot-promoter      volumesnapshot.external-storage.k8s.io/snapshot-promoter   70m
```

### Using OpenEBS storage in a workload
You can optionally specify the storage class created if it is not the default in the PVC spec. [Here](manifests/demo-pvc.yaml) is an example pvc spec file.
You can see that `openebs-cstor-disk` is the storage class created by OpenEBS in the steps above.
More information is on the OpenEBS [docs](https://docs.openebs.io/docs/next/provisionvols.html).

### Running demo workload to test
```
✗ kubectl apply -f manifests/demo-pvc.yaml
persistentvolumeclaim/fio-cstor-claim created

✗ kubectl get persistentvolumeclaims
NAME              STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS                AGE
fio-cstor-claim   Bound    pvc-4b2e5f35-610f-11e9-818a-ec0d9a9eadd6   4G         RWO            default-custom-openebs-sc   48s

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
kubecNAME               READY   STATUS    RESTARTS   AGE
fio-cstor-sparse   1/1     Running   0          7m3s

✗ kubectl delete pods fio-cstor 
pod "fio-cstor" deleted

✗ kubectl apply -f cstor/demo-fio-cstor-workload.yaml
pod/fio-cstor created
persistentvolumeclaim/fio-cstor-claim unchanged

✗ kubectl get pods               
NAME               READY   STATUS    RESTARTS   AGE
fio-cstor   1/1     Running   0          18s

✗ kubectl exec -it fio-cstor bash

root@fio-cstor:/# ls /datadir/
basic.test.file  kosyfile.txt  lost+found

root@fio-cstor:/# cat /datadir/kosyfile.txt
RANDOM TEST FOR STORAGE
```