#!/bin/bash

set -euo pipefail

readonly script_dir=$(cd "$(dirname "${BASH_SOURCE[0]}")" &>/dev/null && pwd -P)
readonly namespace="update-host-files"

kubectl create ns "${namespace}" --dry-run=client -o yaml | kubectl apply -f -
kubectl create -n "${namespace}" cm script --from-file "${script_dir}"/cluster.sh --dry-run=client -o yaml | kubectl apply -f -

function update_node_files() {
  nodename=$1
  mode=$2

  podname="uhf-$nodename-$RANDOM"

  echo "
apiVersion: v1
kind: Pod
metadata:
  labels:
    run: ${podname}
  name: ${podname}
  namespace: ${namespace}
spec:
  containers:
  - image: registry.fedoraproject.org/fedora:32
    name: update-host-files
    imagePullPolicy: IfNotPresent
    securityContext:
      privileged: true
    args:
    - sh
    - -c
    - bash /tmp/script/cluster.sh ${mode}
    volumeMounts:
    - name: etc-kubernetes
      mountPath: /etc/kubernetes/
    - name: script
      mountPath: /tmp/script/
    - name: flatcar-metadata
      mountPath: /run/metadata/flatcar
    - name: kubelet-service
      mountPath: /etc/systemd/system/kubelet.service
  nodeName: ${nodename}
  restartPolicy: Never
  hostPID: true
  volumes:
  - name: etc-kubernetes
    hostPath:
      path: /etc/kubernetes/
  - name: script
    configMap:
      name: script
  - name: flatcar-metadata
    hostPath:
      path: /run/metadata/flatcar
  - name: kubelet-service
    hostPath:
      path: /etc/systemd/system/kubelet.service
" | kubectl apply -f -

  echo -e "\n\nLogs: ${podname}\n\n"

  # Wait until pod exits. Show logs to the user.
  while ! kubectl -n "${namespace}" logs -f "${podname}" 2>/dev/null; do
    sleep 1
  done

  echo '-------------------------------------------------------------------------------------------'
}

function update_controller_nodes() {
  for nodename in $(kubectl get nodes -l node.kubernetes.io/master -ojsonpath='{.items[*].metadata.name}'); do
    update_node_files "${nodename}" "controller"
  done
}

function update_worker_nodes() {
  for nodename in $(kubectl get nodes -l node.kubernetes.io/master!="" -ojsonpath='{.items[*].metadata.name}'); do
    update_node_files "${nodename}" "general"
  done
}

update_controller_nodes
update_worker_nodes
