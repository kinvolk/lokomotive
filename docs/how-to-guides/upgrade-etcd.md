# Upgrading etcd

## Contents

- [Introduction](#introduction)
- [Steps](#steps)
  - [Step 1: SSH into the controller node](#step-1-ssh-into-the-controller-node)
  - [Step 2: Create necessary directories with correct permissions](#step-2-create-necessary-directories-with-correct-permissions)
  - [Step 3: Upgrade etcd](#step-3-upgrade-etcd)
  - [Step 4: Verify upgrade](#step-4-verify-upgrade)
  - [Step 5: Verify using `etcdctl`](#step-5-verify-using-etcdctl)

## Introduction

[Etcd](https://etcd.io/) is the most crucial component of a Kubernetes cluster. It stores the cluster state.

This document will provide step by step guide on upgrading etcd in Lokomotive.

## Steps

Repeat the following steps on all the controller node one node at a time.

### Step 1: SSH into the controller node

Find the IP of the controller node by visiting the cloud provider dashboard and ssh into it.

```bash
ssh core@<IP Address>
```

### Step 2: Create necessary directories with correct permissions

Latest etcd (`>= v3.4.10`) necessitates the data directory permissions to be `0700`, accordingly change the permissions. Verify the permissions are changed to `rwx------`. More information on the changes in upstream release notes can be found [here](https://github.com/etcd-io/etcd/blob/master/CHANGELOG-3.4.md#breaking-changes).

> **NOTE**: This step is needed only for the Lokomotive deployment done using `lokoctl` version `< 0.4.0`.

```bash
sudo chmod 0700 /var/lib/etcd/
sudo ls -ld /var/lib/etcd/
```

If the node reboots, we need the right settings in place so that `systemd-tmpfile` service does not alter the permissions of the data directory. To make the changes persistent run the following command:

```bash
echo "d    /var/lib/etcd 0700 etcd etcd - -" | sudo tee /etc/tmpfiles.d/etcd-wrapper.conf
```

### Step 3: Upgrade etcd

Run the following commands:

> **NOTE**: Before proceeding to other commands, set the `etcd_version` variable to the latest etcd version.

```bash
export etcd_version=<latest etcd version e.g. v3.4.10>

sudo sed -i "s,ETCD_IMAGE_TAG=.*,ETCD_IMAGE_TAG=${etcd_version}\"," \
        /etc/systemd/system/etcd-member.service.d/40-etcd-cluster.conf
sudo systemctl daemon-reload
sudo systemctl restart etcd-member
```

### Step 4: Verify upgrade

Verify that the etcd service is in `active (running)` state:

```bash
sudo systemctl status --no-pager etcd-member
```

Run the following command to see logs of the process since the last restart:

```bash
sudo journalctl _SYSTEMD_INVOCATION_ID=$(sudo systemctl \
              show -p InvocationID --value etcd-member.service)
```

> **NOTE**: Do not proceed with the upgrade of the rest of the cluster if you encounter any errors.

Once you see the following log line, you can discern that the etcd daemon has come up without errors:

```log
etcdserver: starting server... [version: 3.4.10, cluster version: to_be_decided]
```

Once you see the following log line, you can discern that the etcd has rejoined the cluster without issues:

```log
embed: serving client requests on 10.88.81.1:2379
```

### Step 5: Verify using `etcdctl`

We can use `etcdctl` client to verify the state of etcd cluster.

```bash
# Find the endpoint of this node's etcd:
export endpoint=$(grep ETCD_ADVERTISE_CLIENT_URLS /etc/kubernetes/etcd.env | cut -d"=" -f2)
export flags="--cacert=/etc/ssl/etcd/etcd-client-ca.crt \
              --cert=/etc/ssl/etcd/etcd-client.crt \
              --key=/etc/ssl/etcd/etcd-client.key"
endpoints=$(sudo ETCDCTL_API=3 etcdctl member list $flags --endpoints=${endpoint} \
            --write-out=json | jq -r '.members | map(.clientURLs) | add | join(",")')

# Verify:
sudo ETCDCTL_API=3 etcdctl member list $flags --endpoints=${endpoint}
sudo ETCDCTL_API=3 etcdctl endpoint health $flags --endpoints=${endpoints}
```

The last command should report that nodes are healthy. If it indicates otherwise then try commands from [Step 4](#step-4-verify-upgrade) to see what's wrong. If the nodes are healthy, it is safe to move forward with the next controller node.
