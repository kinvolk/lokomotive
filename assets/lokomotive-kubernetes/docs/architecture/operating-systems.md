# Operating Systems

Lokomotive supports [Flatcar Container Linux](https://www.flatcar-linux.org/). Flatcar Container Linux is a friendly fork of CoreOS Container Linux and was chosen because:

* Minimalism and focus on clustered operation
* Automated and atomic operating system upgrades
* Declarative and immutable configuration
* Optimization for containerized applications


## Host Properties

| Property          | Flatcar Container Linux |
|-------------------|-----------------|
| host spec (bare-metal) | Container Linux Config |
| host spec (cloud)      | Container Linux Config |
| container runtime | docker    |
| cgroup driver     | cgroupfs  |
| logging driver    | json-file |
| storage driver    | overlay2  |

## Kubernetes Properties

| Property          | Flatcar Container Linux |
|-------------------|-----------------|
| single-master     | all platforms |
| multi-master      | planned |
| control plane     | self-hosted   |
| kubelet image     | upstream hyperkube |
| control plane images | upstream hyperkube |
| on-host etcd      | rkt-fly   |
| on-host kubelet   | rkt-fly   |
| CNI plugins       | calico or flannel |
| coordinated drain & OS update | [CLUO](https://github.com/coreos/container-linux-update-operator) addon |

## Directory Locations

Lokomotive conventional directories.

| Kubelet setting   | Host location                  |
|-------------------|--------------------------------|
| cni-conf-dir      | /etc/kubernetes/cni/net.d      |
| pod-manifest-path | /etc/kubernetes/manifests      |
| volume-plugin-dir | /var/lib/kubelet/volumeplugins |

## Kubelet Mounts

### Flatcar Container Linux

| Mount location    | Host location     | Options |
|-------------------|-------------------|---------|
| /etc/kubernetes   | /etc/kubernetes   | ro |
| /etc/ssl/certs    | /etc/ssl/certs    | ro |
| /usr/share/ca-certificates | /usr/share/ca-certificates | ro |
| /var/lib/kubelet  | /var/lib/kubelet  | recursive |
| /var/lib/docker   | /var/lib/docker   | |
| /var/lib/cni      | /var/lib/cni      | |
| /var/lib/calico   | /var/lib/calico   | |
| /var/log          | /var/log          | |
| /etc/os-release   | /usr/lib/os-release | ro |
| /run              | /run |            |
| /lib/modules      | /lib/modules | ro |
| /etc/resolv.conf  | /etc/resolv.conf  | |
| /opt/cni/bin      | /opt/cni/bin      | |
