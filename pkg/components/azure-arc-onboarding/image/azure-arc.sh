#!/bin/bash

set -euo pipefail

# Type of action: register or remove the Lokomotive cluster from Azure Arc.
action=$1

function generate_kubeconfig() {
  ca="$(cat /var/run/secrets/kubernetes.io/serviceaccount/ca.crt | base64 | tr -d "\n")"
  namespace="$(cat /var/run/secrets/kubernetes.io/serviceaccount/namespace)"
  token="$(cat /var/run/secrets/kubernetes.io/serviceaccount/token | tr -d "\n")"
  server="https://kubernetes.default.svc"

  echo "
  apiVersion: v1
  kind: Config
  clusters:
  - name: default-cluster
    cluster:
      certificate-authority-data: ${ca}
      server: ${server}
  contexts:
  - name: default-context
    context:
      cluster: default-cluster
      namespace: ${namespace}
      user: default-user
  current-context: default-context
  users:
  - name: default-user
    user:
      token: ${token}
  " > sa.kubeconfig
}

function azure_login() {
  az login --service-principal -u ${AZURE_APPLICATION_CLIENT_ID} -p ${AZURE_APPLICATION_PASSWORD} --tenant ${AZURE_TENANT_ID}
}

function perform_action() {
  case ${action} in
    register)
      # Connect the cluster to Azure Arc.
      az connectedk8s connect --kube-config sa.kubeconfig --name ${CONNECTED_CLUSTER_NAME} --resource-group ${AZURE_RESOURCE_GROUP}
      ;;
    remove)
      # Remove the cluster from Azure Arc.
      az connectedk8s delete --yes --kube-config sa.kubeconfig --name ${CONNECTED_CLUSTER_NAME} --resource-group ${AZURE_RESOURCE_GROUP}
      ;;
    *)
      echo "Unknown action"
      exit 1
      ;;
  esac
}

function checkEnvVars() {
  if [ -z "${AZURE_APPLICATION_CLIENT_ID}" ]; then
    echo "AZURE_APPLICATION_CLIENT_ID is empty"
    exit 1
  fi

  if [ -z "${AZURE_APPLICATION_PASSWORD}" ]; then
    echo "AZURE_APPLICATION_PASSWORD is empty"
    exit 1
  fi

  if [ -z "${CONNECTED_CLUSTER_NAME}" ]; then
    echo "CONNECTED_CLUSTER_NAME is empty"
    exit 1
  fi

  if [ -z "${AZURE_RESOURCE_GROUP}" ]; then
    echo "AZURE_RESOURCE_GROUP is empty"
    exit 1
  fi
}

# Validate environment variables.
checkEnvVars
# Generate kubeconfig from service account.
generate_kubeconfig
# Login to the AZURE CLI.
azure_login
# Register or remove the Lokomotive cluster from Azure arc.
perform_action
