#!/bin/bash

set -euo pipefail

readonly kubelet_env="/etc/kubernetes/kubelet.env"
kubelet_needs_restart=false

function run_on_host() {
  nsenter -a -t 1 /bin/sh -c "${1}"
}

function update_kubelet_version() {
  readonly kubelet_version="v1.20.4"

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

update_kubelet_version
restart_host_kubelet
