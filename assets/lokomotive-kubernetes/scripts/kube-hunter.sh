#!/bin/bash

set -euo pipefail

readonly dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
readonly maxtries=200

function check_tries() {
  if [ "${count}" -gt "${maxtries}" ]; then
    echo "Reached maximum number of retries."
    echo "--------------------------------------"
    echo "info dump:"
    set -x
    kubectl get pods,job
    kubectl get events
    kubectl logs "${pod_name}"
    kubectl get "${pod_name}" -o yaml
    set +x
    echo "--------------------------------------"
    exit 1
  fi
  count=$((count + 1))
}

# Wait until the cluster is responsive
count=0
until kubectl get nodes; do
  check_tries

  echo "Waiting for the cluster to be responsive..."
  sleep 2
done

# Install the kube-hunter job
count=0
until kubectl apply -f "${dir}/kube-hunter-manifests/config.yaml"; do
  check_tries

  echo "Applying kube-hunter manifests..."
  sleep 2
done

# the pod might take sometime to show up
count=0
while [ "$(kubectl get pod -o name -l job-name=kube-hunter)" == "" ]; do
  check_tries

  echo "The pod has not started yet..."
  sleep 2
done
pod_name=$(kubectl get pod -o name -l job-name=kube-hunter)

count=0
while [ "$(kubectl get "${pod_name}" -o jsonpath='{.status.phase}')" != "Succeeded" ]; do
  check_tries

  echo "Waiting until the 'kube-hunter' job is complete..."
  sleep 2
done

echo "--------------------------------------------------"
echo "kube-hunter output:"
echo "--------------------------------------------------"
kubectl logs "${pod_name}"
