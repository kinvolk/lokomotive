#!/bin/bash

set -euo pipefail

function get_latest_release() {
  version=$(curl --silent "https://api.github.com/repos/$1/releases/latest" | jq -r '.tag_name')
  if [ "${version}" != "null" ]; then
    return
  fi

  version=$(curl --silent "https://api.github.com/repos/$1/tags" | jq -r '.[0].name')
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
current_version=$(grep 'k8s.gcr.io/kube-apiserver' assets/terraform-modules/bootkube/variables.tf | cut -d":" -f2 | sed 's/"//g')

version=$(curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt)
printf "${format}" "kubernetes" "${current_version}" "${version}"

###########################
# calico
current_version=$(grep 'calico-node' assets/terraform-modules/bootkube/variables.tf | cut -d":" -f2 | sed 's/"//g')

get_latest_release projectcalico/calico
printf "${format}" "calico" "${current_version}" "${version}"

###########################
# etcd
current_version=$(grep 'IMAGE_TAG=' assets/terraform-modules/aws/flatcar-linux/kubernetes/cl/controller.yaml.tmpl | grep -v KUBELET | cut -d"=" -f2 | sed 's/"//g')

get_latest_release etcd-io/etcd
printf "${format}" "etcd" "${current_version}" "${version}"

###########################
# cert-manager

cd "${workdir}"
current_version=$(grep appVersion assets/charts/components/cert-manager/Chart.yaml | cut -d":" -f2 | sed 's/ //g')

# Download the latest chart into a temp dir and find its version.
tmpdir=$(mktemp -d)
cd "${tmpdir}"
helm repo add jetstack https://charts.jetstack.io >/dev/null 2>&1
helm repo update >/dev/null 2>&1
helm fetch --untar --untardir ./ jetstack/cert-manager >/dev/null 2>&1
version=$(grep appVersion "${tmpdir}/cert-manager/Chart.yaml" | cut -d":" -f2)

printf "${format}" "cert-manager" "${current_version}" "${version}"
rm -rf "${tmpdir}"

###########################
# contour

cd "${workdir}"
current_version=$(grep -A 1 "image: docker.io/projectcontour/contour" assets/charts/components/contour/values.yaml | grep tag | cut -d":" -f2 | sed 's/ //g')

get_latest_release projectcontour/contour
printf "${format}" "contour" "${current_version}" "${version}"

###########################
# Dex
cd "${workdir}"
current_version=$(grep appVersion assets/charts/components/dex/Chart.yaml | cut -d":" -f2 | sed 's/ //g')

get_latest_release dexidp/dex
printf "${format}" "dex" "${current_version}" "${version}"


###########################
# external dns
cd "${workdir}"
current_version=$(grep ^version assets/charts/components/external-dns/Chart.yaml | cut -d":" -f2 | sed 's/ //g' | tail -n1)

tmpdir=$(mktemp -d)
cd "${tmpdir}"
helm repo add bitnami https://charts.bitnami.com/bitnami >/dev/null 2>&1
helm repo update >/dev/null 2>&1
helm fetch --untar --untardir ./ bitnami/external-dns >/dev/null 2>&1
version=$(grep ^version "${tmpdir}/external-dns/Chart.yaml" | cut -d":" -f2)

printf "${format}" "external-dns" "${current_version}" "${version}"
rm -rf "${tmpdir}"

###########################
# flatcar-linux-update-operator

cd "${workdir}"
current_version=$(grep appVersion assets/charts/components/flatcar-linux-update-operator/Chart.yaml | cut -d":" -f2 | sed 's/ //g')

get_latest_release kinvolk/flatcar-linux-update-operator
printf "${format}" "fluo" "${current_version}" "${version}"


###########################
# gangway

cd "${workdir}"
current_version=$(grep appVersion assets/charts/components/gangway/Chart.yaml | cut -d":" -f2 | sed 's/ //g')

get_latest_release heptiolabs/gangway
printf "${format}" "gangway" "${current_version}" "${version}"

###########################
# httpbin

cd "${workdir}"
current_version=$(grep appVersion assets/charts/components/httpbin/Chart.yaml | cut -d":" -f2 | sed 's/ //g')

get_latest_release postmanlabs/httpbin
printf "${format}" "httpbin" "${current_version}" "${version}"


###########################
# metallb

cd "${workdir}"
current_version=$(grep "image: quay.io/kinvolk/metallb-controller" pkg/components/metallb/manifests.go | cut -d":" -f3)

get_latest_release metallb/metallb
printf "${format}" "metallb" "${current_version}" "${version}"

###########################
# metrics-server
cd "${workdir}"
current_version=$(grep ^version assets/charts/components/metrics-server/Chart.yaml | cut -d":" -f2 | sed 's/ //g')

tmpdir=$(mktemp -d)
cd "${tmpdir}"

helm repo add stable https://charts.helm.sh/stable >/dev/null 2>&1
helm repo update >/dev/null 2>&1
helm fetch --untar --untardir ./ stable/metrics-server >/dev/null 2>&1
version=$(grep ^version "${tmpdir}/metrics-server/Chart.yaml" | cut -d":" -f2)

printf "${format}" "metrics-server" "${current_version}" "${version}"
rm -rf "${tmpdir}"

###########################
# openebs
cd "${workdir}"
current_version=$(grep ^version assets/charts/components/openebs-operator/Chart.yaml | cut -d":" -f2 | sed 's/ //g')

tmpdir=$(mktemp -d)
cd "${tmpdir}"

helm repo add openebs https://openebs.github.io/charts >/dev/null 2>&1
helm repo update >/dev/null 2>&1
helm fetch --untar --untardir ./ openebs/openebs >/dev/null 2>&1
version=$(grep ^version "${tmpdir}/openebs/Chart.yaml" | cut -d":" -f2)

printf "${format}" "openebs" "${current_version}" "${version}"
rm -rf "${tmpdir}"

###########################
# prometheus operator
cd "${workdir}"
current_version=$(grep appVersion assets/charts/components/prometheus-operator/Chart.yaml | cut -d":" -f2 | sed 's/ //g')

get_latest_release prometheus-operator/prometheus-operator
printf "${format}" "prometheus-operator" "${current_version}" "${version}"

###########################
# rook
cd "${workdir}"
current_version=$(grep ^version assets/charts/components/rook/Chart.yaml | cut -d":" -f2 | sed 's/ //g')

get_latest_release rook/rook
printf "${format}" "rook" "${current_version}" "${version}"

###########################
# AWS EBS CSI Driver
cd "${workdir}"
current_version=$(grep appVersion assets/charts/components/aws-ebs-csi-driver/Chart.yaml | cut -d":" -f2 | sed 's/ //g' | sed 's/"//g')

tmpdir=$(mktemp -d)
cd "${tmpdir}"

helm repo add aws-ebs-csi-driver https://kubernetes-sigs.github.io/aws-ebs-csi-driver >/dev/null 2>&1
helm repo update >/dev/null 2>&1
helm fetch --untar --untardir ./ aws-ebs-csi-driver/aws-ebs-csi-driver >/dev/null 2>&1
version=$(grep ^version "${tmpdir}/aws-ebs-csi-driver/Chart.yaml" | cut -d":" -f2 | sed 's/ //g')

printf "${format}" "aws-ebs-csi-driver" "${current_version}" "${version}"
rm -rf "${tmpdir}"

###########################
# Velero
cd "${workdir}"
current_version=$(grep ^version assets/charts/components/velero/Chart.yaml | cut -d":" -f2 | sed 's/ //g')

tmpdir=$(mktemp -d)
cd "${tmpdir}"

helm repo add vmware-tanzu https://vmware-tanzu.github.io/helm-charts >/dev/null 2>&1
helm repo update >/dev/null 2>&1
helm fetch --untar --untardir ./ vmware-tanzu/velero >/dev/null 2>&1
version=$(grep ^version "${tmpdir}/velero/Chart.yaml" | cut -d":" -f2 | sed 's/ //g')

printf "${format}" "velero" "${current_version}" "${version}"
rm -rf "${tmpdir}"

###########################
# cluster-autoscaler
cd "${workdir}"
current_version=$(grep ^version assets/charts/components/cluster-autoscaler/Chart.yaml | cut -d":" -f2 | sed 's/ //g' | sed 's/"//g')

get_latest_release kubernetes/autoscaler
latest_version=$(echo $version | cut -d"-" -f4 | sed 's/ //g' | sed 's/"//g')
printf "${format}" "cluster-autoscaler" "${current_version}" "${latest_version}"

###########################
echo
# Print the column names.
printf "${format}" "TF Provider" "Current Version" "Latest Version"
printf "${format}" "-----------" "---------------" "--------------"

###########################
# Packet Provider
cd "${workdir}"
current_version=$(grep packet -A1 assets/terraform-modules/packet/flatcar-linux/kubernetes/versions.tf | tail -1 | cut -d"\"" -f2 | sed 's|~>||g' | sed 's| ||g')

get_latest_release packethost/terraform-provider-packet
printf "${format}" "Packet" "${current_version}" "${version}"

###########################
# AWS Provider
cd "${workdir}"
current_version=$(grep aws -A1 assets/terraform-modules/aws/flatcar-linux/kubernetes/versions.tf | tail -1 | cut -d"\"" -f2 | sed 's|~>||g' | sed 's| ||g')

get_latest_release hashicorp/terraform-provider-aws
printf "${format}" "AWS" "${current_version}" "${version}"

###########################
# Azure Provider
cd "${workdir}"
current_version=$(grep azurerm -A1 assets/terraform-modules/azure/flatcar-linux/kubernetes/versions.tf | tail -1 | cut -d"\"" -f2 | sed 's|~>||g' | sed 's| ||g')

get_latest_release terraform-providers/terraform-provider-azurerm
printf "${format}" "Azure" "${current_version}" "${version}"

###########################
# Local Provider
cd "${workdir}"
current_version=$(grep 'local' -A1 assets/terraform-modules/packet/flatcar-linux/kubernetes/versions.tf | tail -1 | cut -d"\"" -f2 | sed 's|~>||g' | sed 's| ||g')

get_latest_release hashicorp/terraform-provider-local
printf "${format}" "Local" "${current_version}" "${version}"

###########################
# Null Provider
cd "${workdir}"
current_version=$(grep 'null' -A1 assets/terraform-modules/packet/flatcar-linux/kubernetes/versions.tf | tail -1 | cut -d"\"" -f2 | sed 's|~>||g' | sed 's| ||g')

get_latest_release hashicorp/terraform-provider-null
printf "${format}" "Null" "${current_version}" "${version}"

###########################
# CT Provider
cd "${workdir}"
current_version=$(grep ct -A1 assets/terraform-modules/packet/flatcar-linux/kubernetes/versions.tf  | tail -1 | cut -d"\"" -f2 | sed 's|~>||g' | sed 's| ||g')

get_latest_release poseidon/terraform-provider-ct
printf "${format}" "CT" "${current_version}" "${version}"

###########################
# Matchbox Provider
cd "${workdir}"
current_version=$(grep -A1 'poseidon/matchbox' pkg/platform/baremetal/template.go | tail -1 | cut -d"\"" -f2 | sed 's|~>||g' | sed 's| ||g')

get_latest_release poseidon/terraform-provider-matchbox
printf "${format}" "Matchbox" "${current_version}" "${version}"
