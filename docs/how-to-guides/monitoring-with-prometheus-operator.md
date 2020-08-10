---
title: Set up monitoring on Lokomotive
weight: 10
---

## Introduction

This guide provides the steps for deploying a monitoring stack using the `prometheus-operator` Lokomotive component and explains how to access Prometheus, Alertmanager and Grafana.

## Prerequisites

* A Lokomotive cluster deployed on a supported provider and accessible via `kubectl`.

<!---
TODO: Once we have tutorials on how to deploy and configure OpenEBS, point the following to those tutorials.
-->

* A storage provider component ([`rook` and `rook-ceph`](./rook-ceph-storage.md), or `openebs-operator` and `openebs-storage-class`) deployed with a default storage class that can provision volumes for the [PVCs](https://kubernetes.io/docs/concepts/storage/persistent-volumes/#persistentvolumeclaims) created by Alertmanager and Prometheus.

<!---
TODO: Once we have a tutorial on how to deploy and configure Contour and cert-manager, point the following to that tutorial.
-->
> **NOTE**: If you wish to [expose Grafana to the public internet](#using-ingress), the following Lokomotive components should be installed:
> * [`metallb`](./ingress-with-contour-metallb.md) (only on Packet and bare-metal)
> * [`contour`](../configuration-reference/components/contour.md)
> * [`cert-manager`](../configuration-reference/components/cert-manager.md)


## Steps: Deploy Prometheus Operator

### Step 1: Configure Prometheus Operator

Create a file named `monitoring.lokocfg` with the following contents:

```tf
component "prometheus-operator" {}
```

For information about all the available configuration options for the `prometheus-operator` component, visit the component's [configuration reference](../configuration-reference/components/prometheus-operator.md). If you would like to add custom Alerts and Grafana dashboards then look at the section ["Add custom Grafana dashboards"](#add-custom-grafana-dashboards) and subsequent sections.

### Step 2: Install Prometheus Operator

Execute the following command to deploy the `prometheus-operator` component:

```bash
lokoctl component apply prometheus-operator
```

Verify the pods in the `monitoring` namespace are in the `Running` state (this may take a few minutes):

```bash
kubectl -n monitoring get pods
```

## Access Prometheus, Alertmanager and Grafana

### Access Prometheus

#### Using port forward

Execute the following command to forward port `9090` locally to the Prometheus pod:

```bash
kubectl -n monitoring port-forward svc/prometheus-operator-prometheus 9090
```

Open the following URL: [http://localhost:9090](http://localhost:9090).

#### Using Ingress

**NOTE**: NOT RECOMMENDED IN PRODUCTION. Prometheus does not support any authentication out of the box, it has to be enabled at the Ingress layer which is not supported in Lokomotive Ingress at the moment. Therefore, adding following config exposes Prometheus to the outside world and anyone can access it.

To expose Prometheus to the internet using Ingress, provide the `host` field. The configuration for Prometheus in the `prometheus-operator` component should look like the following:

```tf
component "prometheus-operator" {
  prometheus {
    ingress {
      host = "prometheus.<cluster name>.<DNS zone>"
    }
  }
}
```

> **NOTE**: On Packet, you either need to create a DNS entry for `prometheus.<cluster name>.<DNS zone>` and point it to the Packet external IP for the contour service (see the [Packet ingress guide for more details](./ingress-with-contour-metallb.md)) or use the [External DNS component](../configuration-reference/components/external-dns.md).

Open the following URL: `https://prometheus.<cluster name>.<DNS zone>`.

### Access Alertmanager

Execute the following command to forward port `9093` locally to the Alertmanager pod:

```bash
kubectl -n monitoring port-forward svc/prometheus-operator-alertmanager 9093
```

Open the following URL: [http://localhost:9093](http://localhost:9093).

### Access Grafana

#### Using port forward

Execute the following command to forward port `8080` locally to the Grafana dashboard pod on port `80`:

```bash
kubectl -n monitoring port-forward svc/prometheus-operator-grafana 8080:80
```

Obtain the password for the `admin` Grafana user by running the following command:

```bash
kubectl -n monitoring get secret prometheus-operator-grafana -o jsonpath='{.data.admin-password}' | base64 -d && echo
```

Open the following URL: [http://localhost:8080](http://localhost:8080). Enter the username `admin` and password obtained from the previous step.

#### Using Ingress

To expose Grafana to the internet using Ingress, provide the host field. The configuration for Grafana in the `prometheus-operator` component should look like the following:

```tf
component "prometheus-operator" {
  grafana {
    ingress {
      host = "grafana.<cluster name>.<DNS zone>"
    }
  }
}
```

> **NOTE**: On Packet, you either need to create a DNS entry for `grafana.<cluster name>.<DNS zone>` and point it to the Packet external IP for the contour service (see the [Packet ingress guide for more details](./ingress-with-contour-metallb.md)) or use the [External DNS component](../configuration-reference/components/external-dns.md).

Obtain the password for the `admin` Grafana user by running the following command:

```bash
kubectl -n monitoring get secret prometheus-operator-grafana -o jsonpath='{.data.admin-password}' | base64 -d && echo
```

Open the following URL: `https://grafana.<cluster name>.<DNS zone>`. Enter the username `admin` and the password obtained from the previous step.

## Add custom Grafana dashboards

Create a ConfigMap with keys as the dashboard file names and values as JSON dashboard. See the following example:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: grafana-dashboards
  namespace: myapp
  labels:
    grafana_dashboard: "true"
data:
  grafana-dashboard1.json: |
    {
      "annotations": {
[REDACTED]
```

Add the label `grafana_dashboard: "true"` so that grafana automatically picks up the dashboards in the ConfigMaps across the cluster.

This can also be done by using the following two imperative commands:

```bash
kubectl -n myapp create cm grafana-dashboards \
  --from-file=grafana-dashboard1.json \
  --from-file=grafana-dashboard2.json \
  --dry-run -o yaml | kubectl apply -f -

kubectl -n myapp label cm grafana-dashboards grafana_dashboard=true
```

## Add new ServiceMonitors

### Default Prometheus operator setting

Create a ServiceMonitor with the required configuration and make sure to add the following label, so that the prometheus-operator will track it:

```yaml
metadata:
  labels:
    release: prometheus-operator
```

### Custom Prometheus operator setting

Deploy the prometheus-operator with the following setting, and it watches all ServiceMonitors across the cluster:

```tf
watch_labeled_service_monitors = "false"
```

Then there is no need to add any label to ServiceMonitor, at all. Create a ServiceMonitor, and prometheus-operator tracks it.

## Add custom alerts for Alertmanager

### Default Prometheus operator setting

Create a PrometheuRule object with the required configuration and make sure to add the following labels, so that prometheus-operator will track it:

```yaml
metadata:
  labels:
    release: prometheus-operator
    app: prometheus-operator
```

### Custom Prometheus operator setting

Deploy the prometheus-operator with the following setting, and it watches all PrometheusRules across the cluster:

```tf
watch_labeled_prometheus_rules = "false"
```

Then there is no need to add any label to PrometheusRule, at all. Create a PrometheusRule, and prometheus-operator tracks it.

## Additional resources

- `prometheus-operator` component [configuration reference](../configuration-reference/components/prometheus-operator.md).
- ServiceMonitor API docs https://github.com/coreos/prometheus-operator/blob/master/Documentation/api.md#servicemonitor
- PrometheusRule API docs https://github.com/coreos/prometheus-operator/blob/master/Documentation/api.md#prometheusrule
