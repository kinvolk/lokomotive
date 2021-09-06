#!/bin/bash

set -euo pipefail

readonly kubelet_env="/etc/kubernetes/kubelet.env"
kubelet_needs_restart=false
mode="${1}"

function run_on_host() {
  nsenter -a -t 1 /bin/sh -c "${1}"
}

function update_kubelet_version() {
  readonly kubelet_version="v1.21.4"

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

function restart_host_kubelet() {
  if ! "${kubelet_needs_restart}"; then
    return
  fi

  echo -e "\nRestarting Kubelet...\n"
  run_on_host "systemctl daemon-reload && systemctl restart kubelet && systemctl status --no-pager kubelet"
}

function update_etcd() {
  readonly etcd_version="v3.4.16"

  if [ "${mode}" != "controller" ]; then
    echo "Nothing to do. Not a controller node."
    return
  fi

  docker_etcd_cfg="/etc/kubernetes/etcd.env"

  if grep "^IMAGE_TAG=${etcd_version}" "${docker_etcd_cfg}" >/dev/null; then
    echo "etcd env var file ${docker_etcd_cfg} is already updated."
    return
  fi

  sed_cmd="sed 's|^IMAGE_TAG.*|IMAGE_TAG=${etcd_version}|g' ${docker_etcd_cfg} > /tmp/etcd.env"
  restart_etcd_command="systemctl is-active etcd && systemctl restart etcd && systemctl status --no-pager etcd"

  echo -e "\nUpdating etcd file...\nOld etcd file:\n"
  cat "${docker_etcd_cfg}"

  eval "${sed_cmd}"

  cat /tmp/etcd.env >"${docker_etcd_cfg}"

  echo -e "\nNew etcd file...\n"
  cat "${docker_etcd_cfg}"

  echo -e "\nRestarting etcd...\n"
  run_on_host "${restart_etcd_command}"
}

update_etcd
update_kubelet_version
restart_host_kubelet
