#!/bin/bash

set -eoux pipefail

# Move forward with the script, only if this variable is set.
echo "${CI}"

log() {
  local message="${1:-""}"
  echo -e "\\033[1;33m${message}\\033[0m"
}

# It returns the location where most instances are available for the instance
# type passed as parameter.
#
# It returns 0 if it found one or 1 if it didn't. On return it sets the
# $AVAIL_LOCATIONS variable.
best_location_for_instance_type() {
    local instance_type="${1}"

    set +x
    AVAIL_LOCATIONS=$(curl -X GET --header 'Accept: application/json' --header "X-Auth-Token: $PACKET_AUTH_TOKEN" 'https://api.packet.net/capacity' |
        jq -r '.capacity | to_entries[] | { region: .key, avail: .value."'"${instance_type}"'".level} | select(.avail == "normal") | { region: .region } | values[]')
    set -x

    if [ -n "$AVAIL_LOCATIONS" ]; then
        return 0
    fi

    echo "WARNING: no region with \"normal\" ${instance_type} capacity found. Trying \"limited\" capacity..."

    set +x
    AVAIL_LOCATIONS=$(curl -X GET --header 'Accept: application/json' --header "X-Auth-Token: $PACKET_AUTH_TOKEN" 'https://api.packet.net/capacity' |
        jq -r '.capacity | to_entries[] | { region: .key, avail: .value."'"${instance_type}"'".level} | select(.avail == "limited") | { region: .region } | values[]')
    set -x

    if [ -n "$AVAIL_LOCATIONS" ]; then
        return 0
    fi

    echo "ERROR: no region with ${instance_type} availability"
    return 1
}

generate_ssh_keys() {
  if [ "${platform}" != baremetal ] && [ "${platform}" != "kvm-libvirt" ] && [ "${platform}" != "tinkerbell" ]; then
    # Generate SSH key pair to be used by lokoctl.
    log "Generating SSH key pair for lokoctl"
    ssh-keygen -f ~/.ssh/id_rsa -N ''
  fi
}

load_ssh_keys() {
  # Add SSH key to ssh-agent so that lokoctl can SSH into nodes.
  eval "$(ssh-agent)"
  ssh-add ~/.ssh/id_rsa

  PUB_KEY=$(cat ~/.ssh/id_rsa.pub)
  export PUB_KEY
}

generate_cluster_id() {
  SUFFIX="$(printf '%b%b\n' "$(printf '\\%03o' "$((RANDOM % 26 + 97))")" "$(printf '\\%03o' "$((RANDOM % 26 + 97))")")"
  CLUSTER_ID="ci$(date +%s)-$SUFFIX"
  export CLUSTER_ID
}

finalise_packet_location() {
  case "$platform" in
  packet|packet_fluo)
    if ! best_location_for_instance_type "c2.medium.x86"; then
      exit 1
    fi
    AVAIL_MEDIUM_LOCATIONS="${AVAIL_LOCATIONS}"

    if ! best_location_for_instance_type "baremetal_0"; then
      exit 1
    fi
    AVAIL_SMALL_LOCATIONS="${AVAIL_LOCATIONS}"

    AVAIL_LOCATIONS=$(comm -12 <(echo "$AVAIL_SMALL_LOCATIONS" | tr " " "\n") <(echo "$AVAIL_MEDIUM_LOCATIONS" | tr " " "\n"))
    ;;
  packet_arm)
    if ! best_location_for_instance_type "c2.large.arm"; then
      exit 1
    fi
    ;;
  *)
    return
  esac

  # get a random available region
  AVAIL_LOCATIONS_ARRAY=($AVAIL_LOCATIONS)
  export PACKET_LOCATION=${AVAIL_LOCATIONS_ARRAY[$RANDOM % ${#AVAIL_LOCATIONS_ARRAY[@]}]}
}

install_cluster() {
  log "Running lokoctl version $(${LOKOCTL_PATH} version)"
  RET=0
  "${LOKOCTL_PATH}" cluster apply --verbose --skip-components || RET=$?
}

override_fluo() {
  if [ "${RET}" != 0 ]; then
    return
  fi

  # Tell FLUO to pause update reboots for controller nodes
  if [ "$platform" == "packet" ] || [ "$platform" == "packet_fluo" ]; then
    kubectl annotate node --all "flatcar-linux-update.v1.flatcar-linux.net/reboot-paused=true"
  fi
}

install_components() {
  if [ "${RET}" != 0 ]; then
    return
  fi

  "${LOKOCTL_PATH}" component apply || RET=$?
}

run_e2e_tests() {
  if [ "${RET}" != 0 ]; then
    return
  fi

  # move to the root of lokoctl code directory
  cd ../..
  RUN_FROM_CI='"true"' platform="${platform}" make run-e2e-tests || RET=$?
}

delete_components() {
  # Delete all components except external-dns, to give it a chance to remove managed DNS entries.
  "${LOKOCTL_PATH}" component delete $("${LOKOCTL_PATH}" component list | tail -n+2 | awk '{print $1}' | grep -Ev 'external-dns|openebs-storage-class' | tr \\n ' ') --confirm || RET=$?

  echo "Sleeping for 30 seconds. Waiting for external-dns to clear DNS records."
  sleep 30

  if [ "${RET}" = 0 ]; then
    # Delete external-dns component now.
    "${LOKOCTL_PATH}" component delete external-dns --confirm
  fi
}

delete_cluster() {
  cd "$resource_dir/lokoctl/ci/$platform"
  "${LOKOCTL_PATH}" cluster destroy --confirm --verbose
  exit $RET
}

# =======================================================================

log "Deploying test cluster on $platform"
resource_dir=$(pwd)/..

generate_ssh_keys
load_ssh_keys
generate_cluster_id
finalise_packet_location

cd "ci/$platform"
cat "$platform-cluster.lokocfg.envsubst" | envsubst '$AWS_ACCESS_KEY_ID $AWS_SECRET_ACCESS_KEY $PUB_KEY $CLUSTER_ID $AWS_DEFAULT_REGION $AWS_DNS_ZONE $AWS_DNS_ZONE_ID $PACKET_PROJECT_ID $EMAIL $GITHUB_CLIENT_ID $GITHUB_CLIENT_SECRET $DEX_STATIC_CLIENT_CLUSTERAUTH_ID $DEX_STATIC_CLIENT_CLUSTERAUTH_SECRET $GANGWAY_REDIRECT_URL $GANGWAY_SESSION_KEY $DEX_INGRESS_HOST $GANGWAY_INGRESS_HOST $ISSUER_HOST $REDIRECT_URI $API_SERVER_URL $AUTHORIZE_URL $TOKEN_URL $PACKET_LOCATION $ARM_SUBSCRIPTION_ID $ARM_TENANT_ID $ARM_CLIENT_ID $ARM_CLIENT_SECRET' >"$platform-cluster.lokocfg"

export KUBECONFIG=$HOME/lokoctl-assets/cluster-assets/auth/kubeconfig
echo "export KUBECONFIG=$KUBECONFIG" >> ~/.bashrc

LOKOCTL_PATH="${resource_dir}/lokoctl-bin/lokoctl"

install_cluster
override_fluo
install_components
run_e2e_tests
delete_components
delete_cluster
