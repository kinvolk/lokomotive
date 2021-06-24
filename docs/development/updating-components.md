---
title: How to update components?
weight: 10
---

## Introduction

This document explains how to upgrade a particular component in lokoctl. This includes steps to pull config from upstream.

## Find updates

To figure out which component is out dated run the following command in the root of this repository:

```bash
./scripts/find-updates.sh
```

## Etcd

Find the old version and newer version from the `./scripts/find-updates.sh` script and export it accordingly.

```bash
export OLD_VERSION="<old version>"
export NEW_VERSION="<new version>"
```

Now run the following commands in the root of this repository:

```bash
sed -i "s|$OLD_VERSION|$NEW_VERSION|g" assets/terraform-modules/*/flatcar-linux/kubernetes/cl/controller.yaml.tmpl
make update-assets
```

- Releases: https://github.com/etcd-io/etcd/releases.

## Calico

To update Calico update the image tags in following files:

- `assets/charts/control-plane/calico/values.yaml`
- `assets/terraform-modules/bootkube/variables.tf`

Update helm chart in `assets/charts/control-plane/calico` after comparing it with config [here](https://docs.projectcalico.org/manifests/calico.yaml).

- Releases: https://github.com/projectcalico/calico/releases.

## cert-manager

Run the following commands in the root of this repository:

```bash
cd assets/charts/components/
helm repo add jetstack https://charts.jetstack.io
helm repo update

rm -rf cert-manager
helm fetch --untar --untardir ./ jetstack/cert-manager

git checkout ./cert-manager/templates/letsencrypt-clusterissuer-prod.yaml
git checkout ./cert-manager/templates/letsencrypt-clusterissuer-staging.yaml
git checkout ./cert-manager/templates/namespace.yaml
```

Run the following command one at a time and see if the Lokomotive specific changes are not
overridden:

```
git diff ./cert-manager/templates/psp.yaml
git checkout ./cert-manager/templates/psp.yaml

git diff ./cert-manager/templates/cainjector-psp.yaml
git checkout ./cert-manager/templates/cainjector-psp.yaml

git diff ./cert-manager/templates/webhook-psp.yaml
git checkout ./cert-manager/templates/webhook-psp.yaml
```

- Releases: https://github.com/jetstack/cert-manager/releases.

## Metrics Server

Run the following commands in the root of this repository:

```bash
cd assets/components
rm -rf metrics-server
helm fetch --untar --untardir ./ stable/metrics-server
```

- More information about the chart can be found here: https://github.com/helm/charts/tree/master/stable/metrics-server.
- Code repository: https://github.com/kubernetes-sigs/metrics-server.

## OpenEBS

Run the following commands in the root of this repository:

```bash
cd assets/charts/components
rm -rf openebs-operator
helm repo add openebs https://openebs.github.io/charts
helm repo update
helm fetch --untar --untardir ./ openebs/openebs
mv openebs openebs-operator
git checkout openebs-operator/crds/storagepoolclaims.yaml
```

- Installation instructions: https://openebs.github.io/charts/.
- More information about the chart: https://github.com/openebs/charts.
- Code repository: https://github.com/openebs/openebs.

## Prometheus operator

Run the following commands in the root of this repository:

```bash
cd assets/charts/components
rm -rf prometheus-operator
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo update
helm fetch --untar --untardir ./ prometheus-community/kube-prometheus-stack
mv kube-prometheus-stack prometheus-operator
```

- Installation instructions: https://github.com/prometheus-community/helm-charts/tree/main/charts/kube-prometheus-stack.
- Code repository: https://github.com/prometheus-operator/kube-prometheus.

## external-dns

Run the following commands in the root of this repository:

```bash
cd assets/charts/components
rm -rf external-dns
helm repo add bitnami https://charts.bitnami.com/bitnami
helm repo update
helm fetch --untar --untardir ./ bitnami/external-dns
```

- Chart location: https://github.com/bitnami/charts/tree/master/bitnami/external-dns.

## aws-ebs-csi-driver

Run the following commands in the root of this repository:

```bash
cd assets/charts/components
rm -rf aws-ebs-csi-driver
helm repo add aws-ebs-csi-driver https://kubernetes-sigs.github.io/aws-ebs-csi-driver
helm repo update
helm fetch --untar --untardir ./ aws-ebs-csi-driver/aws-ebs-csi-driver
git checkout aws-ebs-csi-driver/templates/networkpolicy.yaml
git checkout aws-ebs-csi-driver/templates/volumesnapshotclass.yaml
git checkout aws-ebs-csi-driver/crds
```

Update the CRDs by following the changes in this [upstream CRD config](https://github.com/kubernetes-sigs/aws-ebs-csi-driver/blob/master/deploy/kubernetes/cluster/crd_snapshotter.yaml).

- Code repository: https://github.com/kubernetes-sigs/aws-ebs-csi-driver.

## velero

Run the following commands in the root of this repository:

```bash
cd assets/charts/components
rm -rf velero
helm repo add vmware-tanzu https://vmware-tanzu.github.io/helm-charts
helm repo update
helm fetch --untar --untardir ./ vmware-tanzu/velero
```

- Chart location: https://github.com/vmware-tanzu/helm-charts/tree/main/charts/velero

## node-problem-detector

Run the following commands in the root of this repository:

```bash
cd assets/charts/components
rm -rf node-problem-detector
helm repo add deliveryhero https://charts.deliveryhero.io/
helm repo update
helm fetch --untar --untardir ./ deliveryhero/node-problem-detector
```

- Chart location: https://github.com/deliveryhero/helm-charts/blob/master/stable/node-problem-detector/Chart.yaml.

## Rook

Run the following commands in the root of this repository:

```bash
cd assets/charts/components
rm -rf rook
helm repo add rook-release https://charts.rook.io/release
helm repo update
helm fetch --untar --untardir ./ rook-release/rook-ceph
mv rook-ceph rook
git checkout rook/templates/service-monitor.yaml
git checkout rook/templates/prometheus-ceph-v14-rules-for-prometheus-operator-0.43.2.yaml
git checkout rook/templates/prometheus-ceph-v14-rules.yaml
git checkout rook/templates/ceph-cluster.yaml
git checkout rook/templates/ceph-osd.yaml
git checkout rook/templates/ceph-pools.yaml
git checkout rook/templates/csi-metrics-service-monitor.yaml
git checkout rook/dashboards
```

- More information about the chart: https://rook.io/docs/rook/v1.5/helm-operator.html.
- Code repository: https://github.com/rook/rook.

## cluster-autoscaler

Run the following commands in the root of this repository:

```bash
cd assets/charts/components
rm -rf cluster-autoscaler
helm repo add autoscaler https://kubernetes.github.io/autoscaler
helm repo update
helm fetch --untar --untardir ./ autoscaler/cluster-autoscaler
```

- Chart location: https://github.com/kubernetes/autoscaler/tree/master/charts/cluster-autoscaler
