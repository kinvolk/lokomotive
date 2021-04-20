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

## Cert Manager

Run the following commands in the root of this repository:

```bash
cd assets/components/cert-manager
helm repo add jetstack https://charts.jetstack.io
helm repo update

rm -rf manifests
helm fetch --untar --untardir ./ jetstack/cert-manager
mv cert-manager manifests

git checkout ./manifests/templates/letsencrypt-clusterissuer-prod.yaml
git checkout ./manifests/templates/letsencrypt-clusterissuer-staging.yaml
git checkout ./manifests/templates/namespace.yaml
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
cd assets/components
rm -rf openebs
helm repo add openebs https://openebs.github.io/charts
helm repo update
helm fetch --untar --untardir ./ openebs/openebs
git checkout openebs/crds/storagepoolclaims.yaml
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

