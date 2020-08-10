---
title: Flatcar Container Linux
weight: 10
---

Lokomotive uses [Flatcar Container Linux](https://www.flatcar-linux.org/) as the underlying operating system.

Flatcar Container Linux is an open source immutable Linux distribution for containers. It is a
friendly fork of CoreOS Container Linux and as such, compatible with it.

## Why Flatcar Container Linux ?

* Minimal distribution required for containers.
  * Reduced dependencies.
  * Reduced attack surface area.
* Immutable file system.
  * Operational simplicity.
  * Removes entire category of security threats. (e.g. runc vulnerability CVE-2019-5736)
    https://kinvolk.io/blog/2019/02/runc-breakout-vulnerability-mitigated-on-flatcar-linux/
* Automated, streamlined and atomic updates.
  * Easily apply all latest security patches.
  * Rollback partition.
* Declarative and immutable configuration.
* Optimization for containerized applications.

### Container runtime properties

Flatcar Container Linux uses Docker as its container runtime. This is the default configuration:


| Property               | Value      |
|------------------------|------------|
| cgroup driver          | cgroupfs   |
| logging driver         | json-file  |
| storage driver         | overlay2   |

### Directory locations

Lokomotive conventional directories:

| Kubelet setting   | Host location                  |
|-------------------|--------------------------------|
| cni-conf-dir      | /etc/cni/net.d                 |
| pod-manifest-path | /etc/kubernetes/manifests      |
| volume-plugin-dir | /var/lib/kubelet/volumeplugins |

### Kubelet mounts

Kubelet mount points on Flatcar Container Linux:

| Mount location             | Host location              | Options   |
|----------------------------|----------------------------|-----------|
| /etc/kubernetes            | /etc/kubernetes            | ro        |
| /etc/ssl/certs             | /etc/ssl/certs             | ro        |
| /usr/share/ca-certificates | /usr/share/ca-certificates | ro        |
| /var/lib/kubelet           | /var/lib/kubelet           | recursive |
| /var/lib/docker            | /var/lib/docker            |           |
| /var/lib/cni               | /var/lib/cni               |           |
| /var/lib/calico            | /var/lib/calico            |           |
| /var/log                   | /var/log                   |           |
| /etc/os-release            | /usr/lib/os-release        | ro        |
| /run                       | /run                       |           |
| /lib/modules               | /lib/modules               | ro        |
| /etc/resolv.conf           | /etc/resolv.conf           |           |
| /opt/cni/bin               | /opt/cni/bin               |           |

### Customization

Flatcar Container Linux can be customized via [Container Linux Configs](https://docs.flatcar-linux.org/container-linux-config-transpiler/doc/examples/) (CLC)
that are interpreted by [Ignition](https://docs.flatcar-linux.org/ignition/what-is-ignition/) (see some [examples](https://github.com/coreos/container-linux-config-transpiler/blob/master/doc/examples.md)).

Lokomotive supports defining CLC snippets for clusters running on Packet and AWS.
CLC snippets can be defined both for controllers and for workers with `controller_clc_snippets` in the
controller definition, and `clc_snippets` for the worker pool definition.

An example CLC snippet:

```hcl
  controller_clc_snippets = [
    file("./snippet/controller-snippet.yaml"),
  ]
```

`clc_snippets` and `controller_clc_snippets` also accept inline text:

```hcl
  clc_snippets = [
  <<EOF
systemd:
  units:
  - name: helloworld.service
    dropins:
      - name: 10-helloworld.conf
        contents: |
          [Install]
          WantedBy=multi-user.target
EOF
        ,
  ]
```