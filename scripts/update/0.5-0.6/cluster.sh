#!/bin/bash

# This script:
# 1) Appends the label `lokomotive.alpha.kinvolk.io/bgp-enabled=true` in the env file for the nodes
#    running MetalLB.
# 2) Update the image tag of Kubelet.
# 3) Update etcd if it is a controller node.

set -euo pipefail

mode="${1}"
update_kubelet_etcd="${2}"

readonly kubelet_env="/etc/kubernetes/kubelet.env"
kubelet_needs_restart=false
packet_environment=false

function run_on_host() {
  nsenter -a -t 1 /bin/sh -c "${1}"
}

function is_packet_environment() {
  if grep -i packet /run/metadata/flatcar > /dev/null; then
    packet_environment=true
  fi
}

is_packet_environment

function update_kubelet_version() {
  readonly kubelet_version="v1.19.4"

  if grep "${kubelet_version}" "${kubelet_env}" >/dev/null; then
    echo "Kubelet env var file ${kubelet_env} already updated, version ${kubelet_version} exists."
    return
  fi

  echo -e "\nUpdating Kubelet env file...\nOld Kubelet env file:\n"
  cat "${kubelet_env}"

  # Update the kubelet image version.
  sed "s|^KUBELET_IMAGE_TAG.*|KUBELET_IMAGE_TAG=${kubelet_version}|g" "${kubelet_env}" >/tmp/kubelet.env

  # This copy is needed because `sed -i` tries to create a new file, this changes the file inode and
  # docker does not allow it. We save changes using `sed` to a temporary file and then overwrite
  # contents of actual file from temporary file.
  cat /tmp/kubelet.env >"${kubelet_env}"

  echo -e "\nNew Kubelet env file:\n"
  cat "${kubelet_env}"

  kubelet_needs_restart=true
}

function update_kubelet_labels() {
  # Update the label only on MetalLB nodes.
  if [ "${mode}" != "metallb" ]; then
    echo "Nothing to do. Not a MetalLB node."
    return
  fi

  readonly metallb_label="lokomotive.alpha.kinvolk.io/bgp-enabled=true"

  if grep "${metallb_label}" "${kubelet_env}" >/dev/null; then
    echo "Kubelet env var file ${kubelet_env} already updated, label ${metallb_label} exists."
    return
  fi

  label=$(grep ^NODE_LABELS "${kubelet_env}")
  label_prefix="${label::-1}"
  augmented_label="${label_prefix},${metallb_label}\""

  echo -e "\nUpdating Kubelet env file...\nOld Kubelet env file:\n"
  cat "${kubelet_env}"

  # Update the kubelet image version.
  sed "s|^NODE_LABELS.*|${augmented_label}|g" "${kubelet_env}" >/tmp/kubelet.env

  cat /tmp/kubelet.env >"${kubelet_env}"

  echo -e "\nNew Kubelet env file:\n"
  cat "${kubelet_env}"

  kubelet_needs_restart=true
}

function update_kubelet_service_file() {
  if ! "${packet_environment}"; then
    echo "Nothing to do. Not a Packet node."
    return
  fi

  readonly kubeletsvcfile="/etc/systemd/system/kubelet.service"
  readonly newline='--cloud-provider=external'

  if grep "cloud-provider=external" "${kubeletsvcfile}" >/dev/null; then
    echo "Kubelet service file ${kubeletsvcfile} is already updated."
    return
  fi

  echo -e "\nUpdating Kubelet service file...\nOld Kubelet service file:\n"
  cat "${kubeletsvcfile}"

  sed '/client-ca-file.*/a \ \ --cloud-provider=external \\' "${kubeletsvcfile}" > /tmp/kubeletsvcfile
  cat /tmp/kubeletsvcfile > "${kubeletsvcfile}"

  echo -e "\nNew Kubelet service file:\n"
  cat "${kubeletsvcfile}"

  kubelet_needs_restart=true
}

function restart_host_kubelet() {
  if ! "${kubelet_needs_restart}"; then
    return
  fi

  echo -e "\nRestarting Kubelet...\n"
  run_on_host "systemctl daemon-reload && systemctl restart kubelet && systemctl status --no-pager kubelet"
}

function update_etcd() {
  if [ "${mode}" != "controller" ]; then
    echo "Nothing to do. Not a controller node."
    return
  fi

  rkt_etcd_cfg="/etc/systemd/system/etcd-member.service.d/40-etcd-cluster.conf"
  docker_etcd_cfg="/etc/kubernetes/etcd.env"
  readonly etcd_version="v3.4.14"

  if [ -f "${rkt_etcd_cfg}" ]; then
    cfg_file="${rkt_etcd_cfg}"
    sed_cmd="sed 's|^Environment=\"ETCD_IMAGE_TAG.*|Environment=\"ETCD_IMAGE_TAG=${etcd_version}\"|g' ${cfg_file} > /tmp/etcd.env"
    restart_etcd_command="systemctl is-active etcd-member && systemctl restart etcd-member && systemctl status --no-pager etcd-member"

  elif [ -f "${docker_etcd_cfg}" ]; then
    cfg_file="${docker_etcd_cfg}"
    sed_cmd="sed 's|^ETCD_IMAGE_TAG.*|ETCD_IMAGE_TAG=${etcd_version}|g' ${cfg_file} > /tmp/etcd.env"
    restart_etcd_command="systemctl is-active etcd && systemctl restart etcd && systemctl status --no-pager etcd"
  fi

  if grep "${etcd_version}" "${cfg_file}" >/dev/null; then
    echo "etcd env var file ${cfg_file} is already updated."
    return
  fi

  echo -e "\nUpdating etcd file...\nOld etcd file:\n"
  cat "${cfg_file}"

  eval "${sed_cmd}"

  cat /tmp/etcd.env >"${cfg_file}"

  echo -e "\nNew etcd file...\n"
  cat "${cfg_file}"

  echo -e "\nRestarting etcd...\n"
  run_on_host "${restart_etcd_command}"
}

if "${update_kubelet_etcd}"; then
  update_etcd
  update_kubelet_version
fi

update_kubelet_labels
update_kubelet_service_file
restart_host_kubelet
