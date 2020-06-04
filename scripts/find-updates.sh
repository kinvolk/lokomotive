#!/bin/bash

set -euo pipefail

function get_latest_release() {
  version=$(curl --silent "https://api.github.com/repos/$1/releases/latest" | jq -r '.tag_name')
}

# Make sure we keep track of the Lokomotive code repository.
workdir=$(pwd)

# Use temporary directory for helm config and data storage.
tmphelm=$(mktemp -d)
export XDG_CACHE_HOME="${tmphelm}"
export XDG_CONFIG_HOME="${tmphelm}"
export XDG_DATA_HOME="${tmphelm}"

# Make sure that we have a format for printing values.
format="%-20s %18s %18s\n"

# Print the column names.
printf "${format}" "Component" "Current Version" "Latest Version"
printf "${format}" "---------" "---------------" "--------------"

###########################
# k8s
current_version=$(grep 'k8s.gcr.io/kube-apiserver' assets/lokomotive-kubernetes/bootkube/variables.tf | cut -d":" -f2 | sed 's/"//g')

version=$(curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt)
printf "${format}" "kubernetes" "${current_version}" "${version}"

###########################
# calico
current_version=$(grep 'calico/node' assets/lokomotive-kubernetes/bootkube/variables.tf | cut -d":" -f2 | sed 's/"//g')

get_latest_release projectcalico/calico
printf "${format}" "calico" "${current_version}" "${version}"

###########################
# etcd
current_version=$(grep 'ETCD_IMAGE_TAG=' assets/lokomotive-kubernetes/aws/flatcar-linux/kubernetes/cl/controller.yaml.tmpl | cut -d"=" -f3 | sed 's/"//g')

get_latest_release etcd-io/etcd
printf "${format}" "etcd" "${current_version}" "${version}"

###########################
# cert-manager

cd "${workdir}"
current_version=$(grep appVersion assets/components/cert-manager/manifests/Chart.yaml | cut -d":" -f2 | sed 's/ //g')

# Download the latest chart into a temp dir and find its version.
tmpdir=$(mktemp -d)
cd "${tmpdir}"
helm repo add jetstack https://charts.jetstack.io >/dev/null 2>&1
helm repo update >/dev/null 2>&1
helm fetch --untar --untardir ./ jetstack/cert-manager
version=$(grep appVersion "${tmpdir}/cert-manager/Chart.yaml" | cut -d":" -f2)

printf "${format}" "cert-manager" "${current_version}" "${version}"
rm -rf "${tmpdir}"

###########################
# contour

cd "${workdir}"
current_version=$(grep -A 1 "image: docker.io/projectcontour/contour" assets/components/contour/values.yaml | grep tag | cut -d":" -f2 | sed 's/ //g')

get_latest_release projectcontour/contour
printf "${format}" "contour" "${current_version}" "${version}"

###########################
# Dex
cd "${workdir}"
current_version=$(grep "image: quay.io/dexidp/dex" pkg/components/dex/component.go | cut -d":" -f3)

get_latest_release dexidp/dex
printf "${format}" "dex" "${current_version}" "${version}"


###########################
# external dns
cd "${workdir}"
current_version=$(grep version assets/components/external-dns/manifests/Chart.yaml | cut -d":" -f2 | sed 's/ //g')

tmpdir=$(mktemp -d)
cd "${tmpdir}"
helm repo add bitnami https://charts.bitnami.com/bitnami >/dev/null 2>&1
helm repo update >/dev/null 2>&1
helm fetch --untar --untardir ./ bitnami/external-dns
version=$(grep version "${tmpdir}/external-dns/Chart.yaml" | cut -d":" -f2)

printf "${format}" "external-dns" "${current_version}" "${version}"
rm -rf "${tmpdir}"

###########################
# gangway

cd "${workdir}"
current_version=$(grep "image: gcr.io/heptio-images/gangway" pkg/components/gangway/component.go | cut -d":" -f3)

get_latest_release heptiolabs/gangway
printf "${format}" "gangway" "${current_version}" "${version}"

###########################
# metallb

cd "${workdir}"
current_version=$(grep "image: quay.io/kinvolk/metallb-controller" pkg/components/metallb/manifests.go | cut -d":" -f3)

get_latest_release metallb/metallb
printf "${format}" "metallb" "${current_version}" "${version}"

###########################
# metrics-server
cd "${workdir}"
current_version=$(grep version assets/components/metrics-server/Chart.yaml | cut -d":" -f2 | sed 's/ //g')

tmpdir=$(mktemp -d)
cd "${tmpdir}"

helm repo add stable https://kubernetes-charts.storage.googleapis.com >/dev/null 2>&1
helm repo update >/dev/null 2>&1
helm fetch --untar --untardir ./ stable/metrics-server
version=$(grep version "${tmpdir}/metrics-server/Chart.yaml" | cut -d":" -f2)

printf "${format}" "metrics-server" "${current_version}" "${version}"
rm -rf "${tmpdir}"

###########################
# openebs
cd "${workdir}"
current_version=$(grep version assets/components/openebs/Chart.yaml | cut -d":" -f2 | sed 's/ //g')

tmpdir=$(mktemp -d)
cd "${tmpdir}"

helm fetch --untar --untardir ./ stable/openebs
version=$(grep version "${tmpdir}/openebs/Chart.yaml" | cut -d":" -f2)

printf "${format}" "openebs" "${current_version}" "${version}"
rm -rf "${tmpdir}"
