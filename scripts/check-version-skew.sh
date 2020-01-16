#!/bin/bash
#
# This test makes sure that apiserver and kubelet are always on the same version. Such version skew
# might show up if terraform render bootkube repo is not correctly updated in lokomotive-kubernetes.
#
# Look at the PR here: https://github.com/kinvolk/lokomotive-kubernetes/pull/104
# This PR was created because when updating lokomotive-kubernetes to 1.16.3 from 1.16.2. All the
# providers were updated except for Packet. And this PR is created to fix the problem.
# But this test tries to avoid such version skews.

set -euo pipefail

function match_with_apiserver_version() {
  kubelet_version=$(kubectl get "${node_name}" -o jsonpath='{.status.nodeInfo.kubeletVersion}')
  if [ "${kubelet_version}" != "${apiserver_version}" ]; then
    echo "${node_name} kubelet version ${kubelet_version}, does not match apiserver."
    echo "--------------------------------------"
    exit 1
  fi
}

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

echo "--------------------------------------"
echo "Testing version skew"
# find the apiserver version
apiserver_version=$(kubectl version -o json | python -c 'import sys, json; print(json.load(sys.stdin)["serverVersion"]["gitVersion"])')
echo "apiserver version: ${apiserver_version}"

# iterate over all the nodes
for node_name in $(kubectl get nodes -o name); do
  match_with_apiserver_version
done
echo "apiserver and kubelet at same version"
echo "--------------------------------------"
