## v0.3.0 - 2020-07-31

We're happy to announce the release of Lokomotive v0.3.0 (Coast Starlight).

This release packs new features and bugfixes. Some of the highlights are:

* Kubernetes 1.18.6
* For Lokomotive clusters running on top of AKS, Kubernetes 1.16.10 is installed.
* [Component updates](#component-updates)

### Changes in v0.3.0

#### Kubernetes updates

- Update Kubernetes to
[v1.18.6](https://github.com/kubernetes/kubernetes/blob/master/CHANGELOG/CHANGELOG-1.18.md#v1186)
([#726](https://github.com/kinvolk/lokomotive/pull/726)).

#### Platform updates

##### Packet
- Update default machine type from `t1.small.x86` to `c3.small.x86`, since
    `t1.small.x86` are EOL and no longer available in new Packet projects
    ([#612](https://github.com/kinvolk/lokomotive/pull/612)).

  **WARNING**: If you haven't explicitly defined the controller_type and/or
  worker_pool.node_type configuration options, upgrading to this release will
  replace your controller and/or worker nodes with c3.small.x86 machines thereby
  losing all your cluster data. To avoid this, set these configuration options
  to the desired values.

  Make sure that the below attributes are explicitly defined in your cluster
  configuration. This only applies to machine type `t1.small.x86`.

  ```hcl
  cluster "packet" {
    .
    .
    controller_type = "t1.small.x86"
    .
    .
    worker_pool "pool-name" {
      .
      node_type = "t1.small.x86"
      .
    }
  }
  ```
##### AKS

- Update Kubernetes version to 1.16.10
    ([#712](https://github.com/kinvolk/lokomotive/pull/712)).

#### Component updates

- openebs: update to 1.11.0 ([#673](https://github.com/kinvolk/lokomotive/pull/673)).

- calico: update to
  [v3.15.0](https://docs.projectcalico.org/release-notes/#v3150)
  ([#652](https://github.com/kinvolk/lokomotive/pull/652)).

#### UX

- prometheus-operator: Organize Prometheus related attributes under a
  `prometheus` block in the configuration
  ([#710](https://github.com/kinvolk/lokomotive/pull/710)).

- Use `prometheus.ingress.host` to expose Prometheus instead of
    `prometheus_external_url`
    ([#710](https://github.com/kinvolk/lokomotive/pull/710)).

- contour: Remove `ingress_hosts` from contour configuration
    ([#635](https://github.com/kinvolk/lokomotive/pull/635)).

#### Features

- Add `enable_toolbox` attribute to [rook-ceph
component](https://github.com/kinvolk/lokomotive/blob/v0.3.0/docs/configuration-reference/components/rook-ceph)
([#649](https://github.com/kinvolk/lokomotive/pull/649)). This allows managing
and configuring Ceph using toolbox pod.

- Add Prometheus feature `external_labels` for federated clusters to Prometheus
    operator component. This helps to identify metrics queried from
    different clusters.
    ([#710](https://github.com/kinvolk/lokomotive/pull/710)).

#### Docs

- Add `Type` column to Attribute reference table in [configuration
 references](https://github.com/kinvolk/lokomotive/blob/v0.3.0/docs/configuration-reference)
 ([#651](https://github.com/kinvolk/lokomotive/pull/651)).

- Update contour configuration reference for usage with AWS
   ([#674](https://github.com/kinvolk/lokomotive/pull/674)).

- Add documentation related to the usage of `clc_snippets` for Packet and AWS
    ([#657](https://github.com/kinvolk/lokomotive/pull/657)).

- Improve documentation on using remote
    [backends](https://github.com/kinvolk/lokomotive/blob/v0.3.0/docs/configuration-reference/backend/s3.md)
    ([#670](https://github.com/kinvolk/lokomotive/pull/670)).

- How to guide for setting up monitoring on Lokomotive
    ([#480](https://github.com/kinvolk/lokomotive/pull/480)).

- Add `codespell` section in [development
    documentation](https://github.com/kinvolk/lokomotive/blob/v0.3.0/docs/development/README.md)
    ([#700](https://github.com/kinvolk/lokomotive/pull/700)).

- Include a demo GIF in the readme
    ([#636](https://github.com/kinvolk/lokomotive/pull/636)).

#### Bugfixes

- Remove contour ingress workaround (due to an [upstream issue](https://github.com/projectcontour/contour/issues/403))
  for ExternalDNS ([#635](https://github.com/kinvolk/lokomotive/pull/635)).

#### Development

- Do not show Helm release values in terraform output
    ([#627](https://github.com/kinvolk/lokomotive/pull/627)).

- Remove Terraform provider aliases from platforms code
    ([#617](https://github.com/kinvolk/lokomotive/pull/617)).


#### Miscellaneous

- Following flatcar-linux/Flatcar#123, Flatcar 2513.1.0 for ARM contains the dig
    binary so the workaround is no longer needed
    ([#703](https://github.com/kinvolk/lokomotive/pull/703)).

- Improve error message for `wait-for-dns` output
    ([#735](https://github.com/kinvolk/lokomotive/pull/735)).

- Add `codespell` to enable spell check on all PRs ([#661](https://github.com/kinvolk/lokomotive/pull/661)).

### Upgrading from v0.2.1

#### Configuration syntax changes

There have been some minor changes to the configurations of following components:
* contour
* prometheus-operator.

Please make sure new the configuration structure is in place before the upgrade.

##### Contour component

Optional `ingress_hosts` attribute is now removed.

old:

```hcl
component "contour" {
  .
  .
  ingress_hosts = ["*.example.lokomotive-k8s.net"]
}
```

new:

```hcl
component "contour" {
  .
  .
}

```

##### Prometheus-operator component

* Prometheus specific attributes are now under a `prometheus` block.
* A new optional `prometheus.ingress` sub-block is introduced to expose
Prometheus over ingress.
* Attribute `external_url` is now removed and now configured under
    `prometheus.ingress.host`. Remove URL scheme (e.g. `https://`) and URI (e.g.
    `/prometheus`) when configuring. URI is no longer supported and protocol is
    always HTTPS.

old:

```hcl
component "prometheus-operator" {
  .
  .
  prometheus_metrics_retention = "14d"
  prometheus_external_url      = "https://prometheus.example.lokomotive-k8s.net"
  prometheus_storage_size      = "50GiB"
  prometheus_node_selector = {
    "kubernetes.io/hostname" = "worker3"
  }
  .
  .
}
```

new:

```hcl
component "prometheus-operator" {
  .
  .
  prometheus {
    metrics_retention = "14d"
    storage_size      = "50GiB"
    node_selector = {
      "kubernetes.io/hostname" = "worker3"
    }

    ingress {
      host                       = "prometheus.example.lokomotive-k8s.net"
    }
    .
    .
  }
  .
  .
}
```

Check out the new syntax in the [Prometheus Operator configuration
reference](https://github.com/kinvolk/lokomotive/blob/v0.3.0/docs/configuration-reference/components/prometheus-operator.md)
for details.

##### Upgrade steps

Go to your cluster's directory and run the following command.

```
lokoctl cluster apply
```

The update process typically takes about 10 minutes.
After the update, running `lokoctl health` should result in an output similar to the following.

```
Node                     Ready    Reason          Message

lokomotive-controller-0  True     KubeletReady    kubelet is posting ready status
lokomotive-1-worker-0    True     KubeletReady    kubelet is posting ready status
lokomotive-1-worker-1    True     KubeletReady    kubelet is posting ready status
lokomotive-1-worker-2    True     KubeletReady    kubelet is posting ready status
Name      Status    Message              Error

etcd-0    True      {"health":"true"}
```

##### Post upgrade steps

##### Openebs

OpenEBS control plane components and data plane components work independently.
Even after the OpenEBS Control Plane components have been upgraded to 1.11.0,
the Storage Pools and Volumes (both jiva and cStor) will continue to work with
older versions.

>Upgrade functionality is still under active development. It is highly
>recommended to schedule a downtime for the application using the OpenEBS PV
>while performing this upgrade. Also, make sure you have taken a backup of the
>data before starting the below upgrade procedure. - [Openebs
documentation](https://github.com/openebs/openebs/blob/master/k8s/upgrades/README.md#step-3-upgrade-the-openebs-pools-and-volumes)


###### Upgrade cStor Pools

* Extract the SPC name using `kubectl get spc`:

```bash
NAME                          AGE
cstor-pool-openebs-replica1   24h
```

The Job spec for upgrade cstor pools is:

```yaml
#This is an example YAML for upgrading cstor SPC.
#Some of the values below needs to be changed to
#match your openebs installation. The fields are
#indicated with VERIFY
---
apiVersion: batch/v1
kind: Job
metadata:
  #VERIFY that you have provided a unique name for this upgrade job.
  #The name can be any valid K8s string for name. This example uses
  #the following convention: cstor-spc-<flattened-from-to-versions>
  name: cstor-spc-1001120

  #VERIFY the value of namespace is same as the namespace where openebs components
  # are installed. You can verify using the command:
  # `kubectl get pods -n <openebs-namespace> -l openebs.io/component-name=maya-apiserver`
  # The above command should return status of the openebs-apiserver.
  namespace: openebs
spec:
  backoffLimit: 4
  template:
    spec:
      #VERIFY the value of serviceAccountName is pointing to service account
      # created within openebs namespace. Use the non-default account.
      # by running `kubectl get sa -n <openebs-namespace>`
      serviceAccountName: openebs-operator
      containers:
      - name:  upgrade
        args:
        - "cstor-spc"

        # --from-version is the current version of the pool
        - "--from-version=1.10.0"

        # --to-version is the version desired upgrade version
        - "--to-version=1.11.0"

        # Bulk upgrade is supported from 1.9
        # To make use of it, please provide the list of SPCs
        # as mentioned below
        - "cstor-pool-name"
        # For upgrades older than 1.9.0, use
        # '--spc-name=<spc_name> format as
        # below commented line
        # - "--spc-name=cstor-sparse-pool"

        #Following are optional parameters
        #Log Level
        - "--v=4"
        #DO NOT CHANGE BELOW PARAMETERS
        env:
        - name: OPENEBS_NAMESPACE
          valueFrom:
            fieldRef:
              fieldPath: metadata.namespace
        tty: true

        # the image version should be same as the --to-version mentioned above
        # in the args of the job
        image: quay.io/openebs/m-upgrade:1.11.0
        imagePullPolicy: Always
      restartPolicy: OnFailure
```

Apply the Job manifest using `kubectl`. Check the logs of the pod started by the Job:

```bash
$ kubectl get logs -n openebs cstor-spc-1001120-dc7kx
..
..
..
I0728 15:15:41.321450       1 spc_upgrade.go:102] Upgrade Successful for spc cstor-pool-openebs-replica1
I0728 15:15:41.321473       1 cstor_spc.go:120] Successfully upgraded storagePoolClaim{cstor-pool-openebs-replica1} from 1.10.0 to 1.11.0
```

###### Upgrade cStor volumes

Extract the PV name using kubectl get pv:

```bash
$ kubectl get pv
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                                                             STORAGECLASS       REASON   AGE
pvc-b69260c4-5cc1-4461-b762-851fa53629d9   50Gi       RWO            Delete           Bound    monitoring/data-alertmanager-prometheus-operator-alertmanager-0   openebs-replica1            24h
pvc-da29e4fe-1841-4da9-a8f6-4e3c92943cbb   50Gi       RWO            Delete           Bound    monitoring/data-prometheus-prometheus-operator-prometheus-0       openebs-replica1            24h
```

Create a Kubernetes Job spec for upgrading the cstor volume. An example spec is
as follows:

```yaml
#This is an example YAML for upgrading cstor volume.
#Some of the values below needs to be changed to
#match your openebs installation. The fields are
#indicated with VERIFY
---
apiVersion: batch/v1
kind: Job
metadata:
  #VERIFY that you have provided a unique name for this upgrade job.
  #The name can be any valid K8s string for name. This example uses
  #the following convention: cstor-vol-<flattened-from-to-versions>
  name: cstor-vol-1001120

  #VERIFY the value of namespace is same as the namespace where openebs components
  # are installed. You can verify using the command:
  # `kubectl get pods -n <openebs-namespace> -l openebs.io/component-name=maya-apiserver`
  # The above command should return status of the openebs-apiserver.
  namespace: openebs

spec:
  backoffLimit: 4
  template:
    spec:
      #VERIFY the value of serviceAccountName is pointing to service account
      # created within openebs namespace. Use the non-default account.
      # by running `kubectl get sa -n <openebs-namespace>`
      serviceAccountName: openebs-operator
      containers:
      - name:  upgrade
        args:
        - "cstor-volume"

        # --from-version is the current version of the volume
        - "--from-version=1.10.0"

        # --to-version is the version desired upgrade version
        - "--to-version=1.11.0"

        # Bulk upgrade is supported from 1.9
        # To make use of it, please provide the list of PVs
        # as mentioned below
        - "pvc-b69260c4-5cc1-4461-b762-851fa53629d9"
        - "pvc-da29e4fe-1841-4da9-a8f6-4e3c92943cbb"
        # For upgrades older than 1.9.0, use
        # '--pv-name=<pv_name> format as
        # below commented line
        # - "--pv-name=pvc-c630f6d5-afd2-11e9-8e79-42010a800065"

        #Following are optional parameters
        #Log Level
        - "--v=4"
        #DO NOT CHANGE BELOW PARAMETERS
        env:
        - name: OPENEBS_NAMESPACE
          valueFrom:
            fieldRef:
              fieldPath: metadata.namespace
        tty: true

        # the image version should be same as the --to-version mentioned above
        # in the args of the job
        image: quay.io/openebs/m-upgrade:1.11.0
        imagePullPolicy: Always
      restartPolicy: OnFailure
---

```

Apply the Job manifest using `kubectl`. Check the logs of the pod started by the Job:

```bash
$ kubectl get logs -n openebs cstor-vol-1001120-8b2h9
..
..
..
I0728 15:19:48.496031       1 cstor_volume_upgrade.go:609] Upgrade Successful for cstor volume pvc-da29e4fe-1841-4da9-a8f6-4e3c92943cbb
I0728 15:19:48.502876       1 cstor_volume.go:119] Successfully upgraded cstorVolume{pvc-da29e4fe-1841-4da9-a8f6-4e3c92943cbb} from 1.10.0 to 1.11.0
```

## v0.2.1 - 2020-06-24

This is a patch release to fix AKS platform deployments.

### Changes in v0.2.1

#### Kubernetes updates

* Updated Kubernetes version on AKS platform to 1.16.9 ([#626](https://github.com/kinvolk/lokomotive/pull/626)). This fixes deploying AKS clusters, as the previously used version is not available anymore.

#### Security

* Updated `golang.org/x/text` dependency to v0.3.3 ([#648](https://github.com/kinvolk/lokomotive/pull/648)) to address [CVE-2020-14040](https://cve.mitre.org/cgi-bin/cvename.cgi?name=CVE-2020-14040).

#### Bugfixes

* Fixes [example](https://github.com/kinvolk/lokomotive/tree/master/examples) configuration for AKS platform ([#626](https://github.com/kinvolk/lokomotive/pull/626)). Contour component configuration syntax changed and those files had not been updated.

#### Misc

* Bootkube Docker images are now pulled using Docker protocol, as quay.io plans to deprecate pulling images using ACI ([#656](https://github.com/kinvolk/lokomotive/pull/656).

#### Development

* AKS platform is now being tested for every pull request and `master` branch changes in the CI.
* Added script for finding available component updates in upstream repositories ([#375](https://github.com/kinvolk/lokomotive/pull/375)).

## v0.2.0 - 2020-06-19

We're happy to announce Lokomotive v0.2.0 (Bernina Express).

This release includes a ton of new features, changes and bugfixes.
Here are some highlights:

* Kubernetes v1.18.3.
* Many [component updates](#component-updates).
* AKS platform support.
* Cloudflare DNS support.
* Monitoring dashboards fixes.
* Dynamic provisioning of Persistent Volumes on AWS.
* [Security improvements](#security).

Check the [full list of changes](#changes-in-v020) for more details.

### Upgrading from v0.1.0

#### Prerequisites

##### All platforms

* The Calico component has a new CRD that needs to be applied manually.

    ```
    kubectl apply -f https://raw.githubusercontent.com/kinvolk/lokomotive/v0.2.0/assets/lokomotive-kubernetes/bootkube/resources/charts/calico/crds/kubecontrollersconfigurations.yaml
    ```

* Some component objects changed `apiVersion` so they need to be labeled and annotated manually to be able to upgrade them.

  * Dex

    ```
    kubectl -n dex label ingress dex app.kubernetes.io/managed-by=Helm
    kubectl -n dex annotate ingress dex meta.helm.sh/release-name=dex
    kubectl -n dex annotate ingress dex meta.helm.sh/release-namespace=dex
    ```

  * Gangway

    ```
    kubectl -n gangway label ingress gangway app.kubernetes.io/managed-by=Helm
    kubectl -n gangway annotate ingress gangway meta.helm.sh/release-name=gangway
    kubectl -n gangway annotate ingress gangway meta.helm.sh/release-namespace=gangway
    ```

  * Metrics Server

    ```
    kubectl -n kube-system label rolebinding metrics-server-auth-reader app.kubernetes.io/managed-by=Helm
    kubectl -n kube-system annotate rolebinding metrics-server-auth-reader meta.helm.sh/release-namespace=kube-system
    kubectl -n kube-system annotate rolebinding metrics-server-auth-reader meta.helm.sh/release-name=metrics-server
    ```

  * httpbin

    ```
    kubectl -n httpbin label ingress httpbin app.kubernetes.io/managed-by=Helm
    kubectl -n httpbin annotate ingress httpbin meta.helm.sh/release-namespace=httpbin
    kubectl -n httpbin annotate ingress httpbin meta.helm.sh/release-name=httpbin
    ```

##### AWS

You need to remove an asset we've updated from your assets directory:

```
rm $ASSETS_DIRECTORY/lokomotive-kubernetes/aws/flatcar-linux/kubernetes/workers.tf
```

#### Upgrading

##### lokocfg syntax changes

Before upgrading, make sure your lokocfg configuration follows the new v0.2.0 syntax.
Here we describe the changes.

###### DNS for the Packet platform

The DNS configuration syntax for the Packet platform has been simplified.

Here's an example for the Route 53 provider.

Old:

```hcl
dns {
    zone = "<DNS_ZONE>"
    provider {
        route53 {
            zone_id = "<ZONE_ID>"
        }
    }
}
```

New:

```hcl
dns {
    zone     = "<DNS_ZONE>"
    provider = "route53"
}
```

Check out the new syntax in the [Packet configuration reference](https://github.com/kinvolk/lokomotive/blob/v0.2.0/docs/configuration-reference/platforms/packet.md#attribute-reference) for details.

###### External DNS component

The [`owner_id` field](https://github.com/kinvolk/lokomotive/blob/v0.2.0/docs/configuration-reference/components/external-dns.md#attribute-reference) is now required.

###### Prometheus Operator component

There is a specific block for Grafana now.

Here's an example of the changed syntax.

Old:

```hcl
component "prometheus-operator" {
    namespace              = "<NAMESPACE>"
    grafana_admin_password = "<GRAFANA_PASSWORD>"
    etcd_endpoints         = ["<ETCD_IP>"]
}
```

New:

```hcl
component "prometheus-operator" {
    namespace = "<NAMESPACE>"
    grafana {
        admin_password = "<GRAFANA_PASSWORD>"
    }
    # etcd_endpoints is not needed anymore
}
```

Check out the new syntax in the [Prometheus Operator configuration reference](https://github.com/kinvolk/lokomotive/blob/v0.2.0/docs/configuration-reference/components/prometheus-operator.md) for details.

##### Upgrade

Go to your cluster's directory and run the following command.

```
lokoctl cluster apply
```

The update process typically takes about 10 minutes.
After the update, running `lokoctl health` should result in an output similar to the following.

```
Node                     Ready    Reason          Message

lokomotive-controller-0  True     KubeletReady    kubelet is posting ready status
lokomotive-1-worker-0    True     KubeletReady    kubelet is posting ready status
lokomotive-1-worker-1    True     KubeletReady    kubelet is posting ready status
lokomotive-1-worker-2    True     KubeletReady    kubelet is posting ready status
Name      Status    Message              Error

etcd-0    True      {"health":"true"}
```

If you have the cert-manager component installed, you will get an error on the first update and need to do a second one.
Run the following to upgrade your components again.

```
lokoctl component apply
```

### Changes in v0.2.0

#### Kubernetes updates

- Update Kubernetes to v1.18.3 ([#459](https://github.com/kinvolk/lokomotive/pull/459)).

#### Component updates

- openebs: update to 1.10.0 ([#528](https://github.com/kinvolk/lokomotive/pull/528)).
- dex: update to v2.24.0 ([#525](https://github.com/kinvolk/lokomotive/pull/525)).
- contour: update to v1.5.0 ([#524](https://github.com/kinvolk/lokomotive/pull/524)).
- cert-manager: update to v0.15.1 ([#522](https://github.com/kinvolk/lokomotive/pull/522)).
- calico: update to v3.14.1 ([#415](https://github.com/kinvolk/lokomotive/pull/415)).
- metrics-server: update to 0.3.6 ([#343](https://github.com/kinvolk/lokomotive/pull/343)).
- external-dns: update to 2.21.2 ([#340](https://github.com/kinvolk/lokomotive/pull/340)).
- rook: update to v1.3.1 ([#300](https://github.com/kinvolk/lokomotive/pull/300)).
- etcd: Update to v3.4.9 ([#521](https://github.com/kinvolk/lokomotive/pull/521)).

#### New platforms

- Add AKS platform support ([#219](https://github.com/kinvolk/lokomotive/pull/219)).

#### Bugfixes

- Handle OS interrupts in lokoctl to fix leaking terraform process ([#483](https://github.com/kinvolk/lokomotive/pull/483)).
- Fix self-hosted Kubelet on bare metal platform ([#436](https://github.com/kinvolk/lokomotive/pull/436)). It wasn't working.
- grafana: remove cluster label in kubelet dashboard ([#474](https://github.com/kinvolk/lokomotive/pull/474)). This fixes missing information in the Kubelet Grafana dashboard.
- Rook Ceph: Fix dashboard templating ([#476](https://github.com/kinvolk/lokomotive/pull/476)). Some graphs were not showing information.
- pod-checkpointer: update to pod-checkpointer image ([#498](https://github.com/kinvolk/lokomotive/pull/498)). Fixes communication between the pod checkpointer and the kubelet.
- Fix AWS worker pool handling ([#367](https://github.com/kinvolk/lokomotive/pull/367)). Remove invisible worker pool of size 0 and fix NLB listener wiring to fix ingress.
- Fix rendering of `ingress_hosts` in Contour component ([#417](https://github.com/kinvolk/lokomotive/pull/417)). Fixes having a wildcard subdomain as ingress for Contour.
- kube-apiserver: fix TLS handshake errors on Packet ([#297](https://github.com/kinvolk/lokomotive/pull/297)). Removes harmless error message.
- calico-host-protection: fix node name of HostEndpoint objects ([#201](https://github.com/kinvolk/lokomotive/pull/201)). Fixes GlobalNetworkPolcies for nodes.

#### Features

- aws: add the AWS EBS CSI driver ([#423](https://github.com/kinvolk/lokomotive/pull/423)). This allows dynamic provisioning of Persistent Volunmes on AWS.
- grafana: provide root_url in grafana.ini conf ([#547](https://github.com/kinvolk/lokomotive/pull/547)). So Grafana exposes its URL and not localhost.
- packet: add Cloudflare DNS support ([#422](https://github.com/kinvolk/lokomotive/pull/422)).
- Monitor etcd by default ([#493](https://github.com/kinvolk/lokomotive/pull/493)). It wasn't being monitored before.
- Add variable `grafana_ingress_host` to expose Grafana ([#468](https://github.com/kinvolk/lokomotive/pull/468)). Allows exposing Grafana through Ingress.
- Add ability to provide oidc configuration ([#182](https://github.com/kinvolk/lokomotive/pull/182)). Allows to configure the API Server to use OIDC for authentication. Previously this was a manual operation.
- Parameterise ClusterIssuer for Dex, Gangway, HTTPBin ([#482](https://github.com/kinvolk/lokomotive/pull/482)). Allows using a different cluster issuer.
- grafana: enable piechart plugin for the Prometheus Operator chart ([#469](https://github.com/kinvolk/lokomotive/pull/469)). Pie chart graphs weren't showing.
- Add a knob to disable self hosted kubelet ([#425](https://github.com/kinvolk/lokomotive/pull/425)).
- rook-ceph: add StorageClass config ([#402](https://github.com/kinvolk/lokomotive/pull/402)). This allows setting up rook-ceph as the default storage class.
- Add monitoring config and variable to rook component ([#405](https://github.com/kinvolk/lokomotive/pull/405)). This allows monitoring rook.
- packet: add support for hardware reservations ([#299](https://github.com/kinvolk/lokomotive/pull/299)).
- Add support for `lokoctl component delete` ([#268](https://github.com/kinvolk/lokomotive/pull/268)).
- bootkube: add calico-kube-controllers ([#283](https://github.com/kinvolk/lokomotive/pull/283)).
- metallb: add AlertManager rules ([#140](https://github.com/kinvolk/lokomotive/pull/140)).
- Label service-monitors so that they are discovered by Prometheus ([#200](https://github.com/kinvolk/lokomotive/pull/200)). This makes sure all components are discovered by Prometheus.
- external-dns: expose owner_id ([#207](https://github.com/kinvolk/lokomotive/pull/207)). Otherwise several clusters in the same DNS Zone will interact badly with each other.
- contour: add Alertmanager rules ([#193](https://github.com/kinvolk/lokomotive/pull/193)).
- contour: add nodeAffinity and tolerations ([#386](https://github.com/kinvolk/lokomotive/pull/386)). This allows using ingress in a subset of cluster nodes.
- prometheus-operator: add storage class & size options ([#387](https://github.com/kinvolk/lokomotive/pull/387)).
- grafana: add secret_env variable ([#541](https://github.com/kinvolk/lokomotive/pull/541)). This allows users to provide arbitrary key values pairs that will be exposed as environment variables inside the Grafana pod.
- rook-ceph: allow volume resizing ([#640](https://github.com/kinvolk/lokomotive/pull/640)). This enables the PVs created by the storage class to be resized on the fly.

#### Security

- Block access to metadata servers for all components by default ([#464](https://github.com/kinvolk/lokomotive/pull/464)). Most components don't need it and it is a security risk.
- packet: disable syncing allowed SSH keys on nodes ([#471](https://github.com/kinvolk/lokomotive/pull/471)). So nodes aren't accessible to all authorized SSH keys in the Packet user and project keys.
- packet: tighten up node bootstrap iptables rules ([#202](https://github.com/kinvolk/lokomotive/pull/202)). So nodes are better protected during bootstrap.
- PSP: Rename `restricted` to `zz-minimal` ([#293](https://github.com/kinvolk/lokomotive/pull/293)). So PSPs apply in the right order.
- kubelet: don't automount service account token ([#306](https://github.com/kinvolk/lokomotive/pull/306)). The Kubelet doesn't need it.
Apiserver, mounted using HostPath.
- prometheus-operator: add seccomp annotations to kube-state-metrics ([#288](https://github.com/kinvolk/lokomotive/pull/288)). This reduces the attack surface by blocking unneeded syscalls.
- prometheus operator: add seccomp annotations to PSP ([#294](https://github.com/kinvolk/lokomotive/pull/294)). So Prometheus Operator pods have seccomp enabled.
- Binding improvements ([#194](https://github.com/kinvolk/lokomotive/pull/194)). This makes the `kubelet`, `kube-proxy` and `calico-node` processes listen on the Host internal IP.

#### UX

- Add `--confirm` flag to delete component without asking for confirmation ([#568](https://github.com/kinvolk/lokomotive/pull/568)).
- Add error message for missing ipxe_script_url ([#540](https://github.com/kinvolk/lokomotive/pull/540)).
- Show logs when terraform fails in `lokoctl cluster apply/destroy` ([#323](https://github.com/kinvolk/lokomotive/pull/323)).
- cli/cmd: rename --kubeconfig flag to --kubeconfig-file ([#602](https://github.com/kinvolk/lokomotive/pull/602)). This is because cobra/viper consider the KUBECONFIG environment variable and the --kubeconfig flag the same and this can cause surprising behavior.

#### Docs

- docs: make Packet quickstart quick ([#332](https://github.com/kinvolk/lokomotive/pull/332)).
- docs: document Route 53 and S3+DynamoDB permissions ([#561](https://github.com/kinvolk/lokomotive/pull/561)).
- docs/quickstart/aws: Fix flatcar-linux-update-operator link ([#552](https://github.com/kinvolk/lokomotive/pull/552)).
- docs: Add detailed contributing guidelines ([#404](https://github.com/kinvolk/lokomotive/pull/404)).
- docs: Add instructions to run conformance tests ([#236](https://github.com/kinvolk/lokomotive/pull/236)).
- docs/quickstarts: add reference to usage docs and PSP note ([#233](https://github.com/kinvolk/lokomotive/pull/233)).
- docs: clarify values for ssh_pubkeys ([#230](https://github.com/kinvolk/lokomotive/pull/230)).
- docs/quickstarts: fix kubeconfig path ([#229](https://github.com/kinvolk/lokomotive/pull/229)).
- docs/prometheus-operator: clarify alertmanager config indentation ([#199](https://github.com/kinvolk/lokomotive/pull/199)).
- quickstart-docs: Add ssh-agent instructions ([#325](https://github.com/kinvolk/lokomotive/pull/325)).
- docs: provide alternate way of declaring alertmanager config ([#570](https://github.com/kinvolk/lokomotive/pull/570)).
- examples: make Flatcar channels explicit ([#565](https://github.com/kinvolk/lokomotive/pull/565)).
- docs/aws: document TLS handshake errors in kube-apiserver ([#599](https://github.com/kinvolk/lokomotive/pull/599)).

#### Misc

- Update terraform-provider-ct to v0.5.0 and mention it in the docs ([#281](https://github.com/kinvolk/lokomotive/pull/281)).
- Update broken links ([#569](https://github.com/kinvolk/lokomotive/pull/569)).
- Fix example configs and typos ([#535](https://github.com/kinvolk/lokomotive/pull/535)).
- docs/httpbin: Fix table ([#510](https://github.com/kinvolk/lokomotive/pull/510)).
- Add missing bracket in Prometheus Operator docs ([#490](https://github.com/kinvolk/lokomotive/pull/490)).
- docs: Update the component deleting steps ([#481](https://github.com/kinvolk/lokomotive/pull/481)).
- Fix broken bare metal config link ([#473](https://github.com/kinvolk/lokomotive/pull/473)).
- Remove period (.) from flag descriptions ([#574](https://github.com/kinvolk/lokomotive/pull/574)).
- Several fixes to make updates from v0.1.0 smooth ([#638](https://github.com/kinvolk/lokomotive/pull/638), [#639](https://github.com/kinvolk/lokomotive/pull/639), [#642](https://github.com/kinvolk/lokomotive/pull/642))
- baremetal quickstart: Add double quotes ([#633](https://github.com/kinvolk/lokomotive/pull/633)).
- pkg/components/util: improvements ([#605](https://github.com/kinvolk/lokomotive/pull/605)).
- New internal package for helper functions ([#588](https://github.com/kinvolk/lokomotive/pull/588)).
- Remove vars from assets that were unused by tmpl file ([#620](https://github.com/kinvolk/lokomotive/pull/620)).
- keys: iago's key is kinvolk.io, not gmail 🤓 ([#616](https://github.com/kinvolk/lokomotive/pull/616)).

## v0.1.0 - 2020-03-18

Initial release.

* Kubernetes v1.18.0
* Running on [Flatcar Container Linux](https://www.flatcar-linux.org/)
* Fully self-hosted, including the kubelet
* Single or multi-master
* Calico networking
* On-cluster etcd with TLS, RBAC-enabled, PSP-enabled, network policies
* In-place upgrades, including experimental kubelet upgrades
* Supported on:
    * [Packet](https://github.com/kinvolk/lokomotive/blob/v0.1.0/docs/configuration-reference/platforms/packet.md)
    * [AWS](https://github.com/kinvolk/lokomotive/blob/v0.1.0/docs/configuration-reference/platforms/aws.md)
    * [Bare metal](https://github.com/kinvolk/lokomotive/blob/v0.1.0/docs/configuration-reference/platforms/baremetal.md)
* Initial Lokomotive Components:
    * [calico-hostendpoint-controller](https://github.com/kinvolk/lokomotive/blob/v0.1.0/docs/configuration-reference/components/calico-hostendpoint-controller.md)
    * [cert-manager](https://github.com/kinvolk/lokomotive/blob/v0.1.0/docs/configuration-reference/components/cert-manager.md)
    * [cluster-autoscaler](https://github.com/kinvolk/lokomotive/blob/v0.1.0/docs/configuration-reference/components/cluster-autoscaler.md)
    * [contour](https://github.com/kinvolk/lokomotive/blob/v0.1.0/docs/configuration-reference/components/contour.md)
    * [dex](https://github.com/kinvolk/lokomotive/blob/v0.1.0/docs/configuration-reference/components/dex.md)
    * [external-dns](https://github.com/kinvolk/lokomotive/blob/v0.1.0/docs/configuration-reference/components/external-dns.md)
    * [flatcar-linux-update-operator](https://github.com/kinvolk/lokomotive/blob/v0.1.0/docs/configuration-reference/components/flatcar-linux-update-operator.md)
    * [gangway](https://github.com/kinvolk/lokomotive/blob/v0.1.0/docs/configuration-reference/components/gangway.md)
    * [httpbin](https://github.com/kinvolk/lokomotive/blob/v0.1.0/docs/configuration-reference/components/httpbin.md)
    * [metallb](https://github.com/kinvolk/lokomotive/blob/v0.1.0/docs/configuration-reference/components/metallb.md)
    * [metrics-server](https://github.com/kinvolk/lokomotive/blob/v0.1.0/docs/configuration-reference/components/metrics-server.md)
    * [openebs-operator](https://github.com/kinvolk/lokomotive/blob/v0.1.0/docs/configuration-reference/components/openebs-operator.md)
    * [openebs-storage-class](https://github.com/kinvolk/lokomotive/blob/v0.1.0/docs/configuration-reference/components/openebs-storage-class.md)
    * [prometheus-operator](https://github.com/kinvolk/lokomotive/blob/v0.1.0/docs/configuration-reference/components/prometheus-operator.md)
    * [rook](https://github.com/kinvolk/lokomotive/blob/v0.1.0/docs/configuration-reference/components/rook.md)
    * [rook-ceph](https://github.com/kinvolk/lokomotive/blob/v0.1.0/docs/configuration-reference/components/rook-ceph.md)
    * [velero](https://github.com/kinvolk/lokomotive/blob/v0.1.0/docs/configuration-reference/components/velero.md)

## v0.0.0
