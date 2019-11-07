# Upgrading cluster to Kubernetes Patch Release

Lokomotive cluster runs control plane viz. `apiserver`, `scheduler`, `controller-manager`, `coredns`, `kube-proxy` and `calico` as pods, just like any other workloads. To updrage them it is mostly about changing the image tag, which can be done using `kubectl`.

**NOTE**: This document only covers upgrading from one patch version to another. So if you are on Kubernetes `1.15.1` then this guide explains how to upgrade to `1.15.5`. Since this is patch upgrade it is assumed that there were no major changes and just bug and security fixes.

## Prepare

Be ready with the latest patch versions.

* For `apiserver`, `scheduler`, `controller-manager`, `kube-proxy`

Find your current version:

```bash
kubectl version --short | grep -i server | cut -d':' -f2-
```

You can see the latest patch version on the Kubernetes release page: https://github.com/kubernetes/kubernetes/releases. After you find the latest patch version just run following command. For e.g. `export latest_kube="v1.15.5"`.

```
export latest_kube="<latest patch version of kubernetes>"
```

* For `coredns` check the latest patch version.

Find your current version:

```
kubectl -n kube-system get deploy coredns -o jsonpath='{.spec.template.spec.containers[0].image}' | cut -d':' -f2-
```

You can see the latest patch version on the CoreDNS release page: https://github.com/coredns/coredns/releases. After you find the latest patch version just run following command. For e.g. `export latest_coredns="1.6.5"`.

**Note**: There is no `v` prefixed to the version.

```
export latest_coredns="<latest patch version of coredns>"
```

* For `calico` check the latest patch version.

Find your current version:

```
kubectl -n kube-system get ds calico-node -o jsonpath='{.spec.template.spec.containers[0].image}' | cut -d':' -f2-
```

You can see the latest patch version on the Calico release page: https://github.com/projectcalico/calico/releases. After you find the latest patch version just run following command. For e.g. `export latest_calico="v3.8.4"`.

```
export latest_calico="<latest patch version of calico>"
```

## Upgrade Control Plane

While making following changes you can have a terminal where you see all the events from `kube-system` namespace:

```
kubectl -n kube-system get events -w
```

### kube-apiserver

```
kubectl set image daemonset -n kube-system kube-apiserver kube-apiserver="k8s.gcr.io/hyperkube:${latest_kube}"
```

You can verify that apiserver is back by running following command:

```
kubectl get daemonset -n kube-system kube-apiserver
```

### kube-scheduler

```
kubectl set image deployment -n kube-system kube-scheduler kube-scheduler="k8s.gcr.io/hyperkube:${latest_kube}"
```

You can verify that scheduler is back by running following command:

```
kubectl get deployment -n kube-system kube-scheduler
```

### kube-controller-manager

```
kubectl set image deployment -n kube-system kube-controller-manager kube-controller-manager="k8s.gcr.io/hyperkube:${latest_kube}"
```

You can verify that controller manager is back by running following command:

```
kubectl get deployment -n kube-system kube-controller-manager
```

### kube-proxy

```
kubectl set image daemonset -n kube-system kube-proxy kube-proxy="k8s.gcr.io/hyperkube:${latest_kube}"
```

You can verify that kube-proxy is back by running following command:

```
kubectl get daemonset -n kube-system kube-proxy
```

### CoreDNS

```
kubectl set image deployment -n kube-system coredns coredns="k8s.gcr.io/coredns:${latest_coredns}"
```

```
kubectl get deployment -n kube-system coredns
```

### Calico

```
kubectl set image daemonset -n kube-system calico-node calico-node="quay.io/calico/node:${latest_calico}" install-cni="quay.io/calico/cni:${latest_calico}"
```

You can verify that calico is back by running following command:

```
kubectl get daemonset -n kube-system calico-node
```

### Verify control plane is upgraded

```
kubectl version --short | grep -i server
```

## Update kubelets

SSH into each node(including master nodes) and run following commands. Make sure to change the version to what was exported before.

```
export latest_kube="<latest patch version of kubernetes>"
sudo sed -i "s|$(grep -i kubelet_image_tag /etc/kubernetes/kubelet.env)|KUBELET_IMAGE_TAG=${latest_kube}|g" /etc/kubernetes/kubelet.env
sudo systemctl restart kubelet
```
### Verify

```
kubectl get nodes -o yaml | grep 'kubeProxyVersion'
kubectl get nodes -o yaml | grep 'kubeletVersion'
kubectl get nodes
```

If you see the latest version for each of the above output, then the cluster is successfully updated to the latest patch version.
