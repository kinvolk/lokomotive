# OpenEBS operator

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
Created  Deployment openebs/openebs-localpv-provisioner
Created  Deployment openebs/openebs-admission-server
Created  Deployment openebs/openebs-ndm-operator
Created  Deployment openebs/openebs-snapshot-operator
Created  ConfigMap openebs/openebs-ndm-config
Created  DaemonSet openebs/openebs-ndm
```

**Verify**

After a while, all the OpenEBS components should be done creating and in running state.
Please ensure to verify that everything works properly before next steps.
```
✗ kubectl get pods -n openebs
NAME                                            READY   STATUS    RESTARTS   AGE
maya-apiserver-54f87dcb7b-7gxsd                 1/1     Running   0          60s
openebs-localpv-provisioner-8657cdcd6f-c7p2f    1/1     Running   0          60s
openebs-admission-server-7bc49854bb-wmqww       1/1     Running   0          60s
openebs-ndm-operator-74b4fb6fcd-49xlk           1/1     Running   0          57s
openebs-ndm-4t7ws                               1/1     Running   0          57s
openebs-ndm-jwxwf                               1/1     Running   0          57s
openebs-ndm-l2v9s                               1/1     Running   0          57s
openebs-provisioner-57554c764-dvffn             1/1     Running   0          58s
openebs-snapshot-operator-96464fd9d-pscdk       2/2     Running   0          58s
```
If you initially setup nodeSelectors for Node Disk Manager(NDM), you should see that pods are only scheduled on labelled nodes with ` kubectl get pods -n openebs -o wide`.

This component only concerns with the setup of openebs-operator. To configure the storageclass and storage pool claim, check out the [openebs-storage-class](openebs-storage-class.md) component.