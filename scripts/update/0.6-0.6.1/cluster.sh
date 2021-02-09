#!/bin/bash

set -euo pipefail

mode="${1}"

function run_on_host() {
  nsenter -a -t 1 /bin/sh -c "${1}"
}

function update_etcd() {
  if [ "${mode}" != "controller" ]; then
    echo "Nothing to do. Not a controller node."
    return
  fi

  rkt_etcd_cfg="/etc/systemd/system/etcd-member.service.d/40-etcd-cluster.conf"
  docker_etcd_cfg="/etc/kubernetes/etcd.env"
  docker_etcd_svc="/etc/systemd/system/etcd.service"

  if [ -f "${rkt_etcd_cfg}" ]; then
    echo "Nothing to do. Rkt based etcd node."
    return
  fi

  if grep "^IMAGE_TAG" "${docker_etcd_cfg}" >/dev/null; then
    echo "etcd env var file ${docker_etcd_cfg} is already updated."
    return
  fi

  echo -e "\nUpdating etcd file...\nOld etcd config file:\n"
  cat "${docker_etcd_cfg}"
  sed 's|ETCD_IMAGE_TAG|IMAGE_TAG|g; s|ETCD_IMAGE_URL|IMAGE_URL|g; s|ETCD_SSL_DIR|SSL_DIR|g; s|ETCD_USER|USER|g' ${docker_etcd_cfg} >/tmp/etcd.env
  cat /tmp/etcd.env >"${docker_etcd_cfg}"
  echo -e "\nNew etcd config file:\n"
  cat "${docker_etcd_cfg}"

  echo -e "\nOld etcd service file:\n"
  cat "${docker_etcd_svc}"
  sed 's|ETCD_IMAGE_TAG|IMAGE_TAG|g; s|ETCD_IMAGE_URL|IMAGE_URL|g; s|ETCD_SSL_DIR|SSL_DIR|g; s|ETCD_USER|USER|g' ${docker_etcd_svc} >/tmp/etcd.service
  cat /tmp/etcd.service >"${docker_etcd_svc}"
  echo -e "\nNew etcd service file:\n"
  cat "${docker_etcd_svc}"

  echo -e "\nRestarting etcd...\n"
  run_on_host "systemctl daemon-reload && systemctl is-active etcd && systemctl restart etcd && systemctl status --no-pager etcd"
}

update_etcd
