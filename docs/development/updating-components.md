# How to update components?

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
sed -i "s|$OLD_VERSION|$NEW_VERSION|g" assets/lokomotive-kubernetes/*/flatcar-linux/kubernetes/cl/controller.yaml.tmpl
make update-assets
```

- Releases: https://github.com/etcd-io/etcd/releases.

## Calico

To update Calico update the image tags in following files:

- `assets/lokomotive-kubernetes/bootkube/resources/charts/calico/values.yaml`
- `assets/lokomotive-kubernetes/bootkube/variables.tf`

If there are changes necessary in the helm chart, make them in `assets/lokomotive-kubernetes/bootkube/resources/charts/calico`.

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
helm fetch --untar --untardir ./ stable/openebs
git checkout openebs/crds/storagepoolclaims.yaml
```

- Installation instructions: https://docs.openebs.io/docs/next/installation.html.
- More information about the chart: https://github.com/helm/charts/tree/master/stable/openebs.
- Code repository: https://github.com/openebs/openebs.
