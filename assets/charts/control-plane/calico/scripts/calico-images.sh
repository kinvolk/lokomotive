#!/bin/bash
# This script downloads the calico images from docker hub and uploads them to quay.io.
# Run this script once on AMD and then on ARM machine.

version=$1
if [ -z "${version}" ]; then
  echo "Please provide the version of Calico. Check https://github.com/projectcalico/calico/releases/."
  echo "e.g."
  echo "./calico-images.sh v3.19.1"
  exit 1
fi

set -euo pipefail

declare -a StringArray=(
  "node"
  "cni"
  "kube-controllers"
  "pod2daemon-flexvol"
)

function find_arch() {
  if [ "$(uname -m)" = "x86_64" ]; then
    arch="amd64"
  elif [ "$(uname -m)" = "aarch64" ]; then
    arch="arm64"
  else
    echo "Unknown architecture: $(uname -m). Only x86_64 and aarch64 are supported."
    exit 1
  fi

  echo "Identified architecture: ${arch}."
}

function push_to_quay() {
  for image in "${StringArray[@]}"; do
    quay_url="quay.io/kinvolk/calico-${image}:${version}-${arch}"

    docker pull "docker.io/calico/${image}:${version}"
    docker tag "docker.io/calico/${image}:${version}" "${quay_url}"
    docker push "${quay_url}"
  done
}

function create_arch_agnostic_tag() {
  for image in "${StringArray[@]}"; do
    quay_url="quay.io/kinvolk/calico-${image}:${version}"

    # At this point the script should have pushed both arch images. Check if that is the case.
    if ! docker pull "${quay_url}-amd64"; then
      echo "Run the script on AMD machine as well!"
      exit 1
    fi

    if ! docker pull "${quay_url}-arm64"; then
      echo "Run the script on ARM machine as well!"
      exit 1
    fi

    docker manifest create "${quay_url}" \
      --amend "${quay_url}-amd64" \
      --amend "${quay_url}-arm64"

    docker manifest annotate "${quay_url}" \
      "${quay_url}-amd64" --arch=amd64 --os=linux

    docker manifest annotate "${quay_url}" \
      "${quay_url}-arm64" --arch=arm64 --os=linux

    docker manifest push "${quay_url}"
  done
}

find_arch
push_to_quay
create_arch_agnostic_tag
