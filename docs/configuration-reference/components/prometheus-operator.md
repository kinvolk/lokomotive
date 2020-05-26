# Prometheus Operator configuration reference for Lokomotive

## Contents

* [Introduction](#introduction)
* [Prerequisites](#prerequisites)
* [Configuration](#configuration)
* [Attribute reference](#attribute-reference)
* [Applying](#applying)
* [Deleting](#deleting)

## Introduction

The [Prometheus Operator](https://github.com/coreos/prometheus-operator) for Kubernetes provides
easy monitoring definitions for Kubernetes services and deployment and management of Prometheus
instances.

## Prerequisites

* A Lokomotive cluster with a
  [PersistentVolume](https://kubernetes.io/docs/concepts/storage/persistent-volumes/) plugin, e.g.
  [OpenEBS](openebs.md) or one of the
  [built-in](https://kubernetes.io/docs/concepts/storage/persistent-volumes/#types-of-persistent-volumes)
  plugins.

## Configuration

Prometheus Operator component configuration example:

```tf
component "prometheus-operator" {
  namespace              = "monitoring"
  grafana_admin_password = "foobar"

  etcd_endpoints = [
    "10.88.181.1",
  ]

  prometheus_metrics_retention = "14d"
  prometheus_external_url      = "https://api.example.com/prometheus"
  prometheus_node_selector = {
    "kubernetes.io/hostname" = "worker3"
  }


  alertmanager_retention    = "360h"
  alertmanager_external_url = "https://api.example.com/alertmanager"
  alertmanager_config       = file("alertmanager-config.yaml")
  alertmanager_node_selector = {
    "kubernetes.io/hostname" = "worker3"
  }
```

Create `alertmanager-config.yaml` file if necessary. Visit the [alertmanager
configuration](https://prometheus.io/docs/alerting/configuration/#configuration-file) for more
information.

**Note**: Make sure the whole file is indented two spaces. That is, there are two spaces before the top level block.

```yaml
  config:
    global:
      resolve_timeout: 5m
    route:
      group_by:
      - job
      group_wait: 30s
      group_interval: 5m
      repeat_interval: 12h
      receiver: 'null'
      routes:
      - match:
          alertname: Watchdog
        receiver: 'null'
    receivers:
    - name: 'null'
```

## Attribute reference

Table of all the arguments accepted by the component.

Example:

| Argument | Description | Default | Required |
|--------	|--------------|:-------:|:--------:|
| `namespace` | Namespace to deploy the Prometheus Operator. | - | true |
| `grafana_admin_password` | Password for `admin` user in Grafana.  | - | true |
| `etcd_endpoints` | List of endpoints where etcd can be reachable from Kubernetes. | [] | false |
| `prometheus_operator_node_selector` | Node selector to specify nodes where the Prometheus Operator pods should be deployed. | {} | false |
| `prometheus_metrics_retention` | Time duration Prometheus shall retain data for. Must match the regular expression `[0-9]+(ms\|s\|m\|h\|d\|w\|y)` (milliseconds, seconds, minutes, hours, days, weeks and years). | `10d` | false |
| `prometheus_external_url` | The external URL Prometheus instances will be available under. This is necessary to generate correct URLs. This is necessary if Prometheus is not served from root of a DNS name. | "" | false |
| `prometheus_node_selector` | Node selector to specify nodes where the Prometheus pods should be deployed. | {} | false |
| `prometheus_storage_size` | Storage capacity for the Prometheus in bytes. You can express storage as a fixed-point integer using one of these suffixes: E, P, T, G, M, K. You can also use the power-of-two equivalents: Ei, Pi, Ti, Gi, Mi, Ki. | "50Gi" | false |
| `watch_labeled_service_monitors` | By default prometheus operator watches only the ServiceMonitor objects in the cluster that are labeled `release: prometheus-operator`. If set to `false` then all the ServiceMonitors will be watched. | `true` | false |
| `watch_labeled_prometheus_rules` | By default prometheus operator watches only the PrometheusRule objects in the cluster that are labeled `release: prometheus-operator` and `app: prometheus-operator`. If set to `false` then all the PrometheusRule will be watched. | `true` | false |
| `alertmanager_retention` | Time duration Alertmanager shall retain data for. Must match the regular expression `[0-9]+(ms\|s\|m\|h)` (milliseconds, seconds, minutes and hours). | `120h` | false |
| `alertmanager_external_url` | The external URL the Alertmanager instances will be available under. This is necessary to generate correct URLs. This is necessary if Alertmanager is not served from root of a DNS name. | "" | false |
| `alertmanager_config` | Provide YAML file path to configure Alertmanager. See [https://prometheus.io/docs/alerting/configuration/#configuration-file](https://prometheus.io/docs/alerting/configuration/#configuration-file). | `{"global":{"resolve_timeout":"5m"},"route":{"group_by":["job"],"group_wait":"30s","group_interval":"5m","repeat_interval":"12h","receiver":"null","routes":[{"match":{"alertname":"Watchdog"},"receiver":"null"}]},"receivers":[{"name":"null"}]}` | false |
| `alertmanager_node_selector` | Node selector to specify nodes where the AlertManager pods should be deployed. | {} | false |
| `alertmanager_storage_size` | Storage capacity for the Alertmanager in bytes. You can express storage as a fixed-point integer using one of these suffixes: E, P, T, G, M, K. You can also use the power-of-two equivalents: Ei, Pi, Ti, Gi, Mi, Ki. | "50Gi" | false |
| `disable_webhooks` | Disables validation and mutation webhooks. This might be required on older versions of Kubernetes to install successfully. | false | false |
| `monitor` | Block, which allows to disable scraping of individual Kubernetes components. | - | false |
| `monitor.etcd` | Controls if the default Prometheus instance should scrape etcd metrics. | true | false |
| `monitor.kube_controller_manager` | Controls if the default Prometheus instance should scrape kube-controller-manager metrics. | true | false |
| `monitor.kube_scheduler` | Controls if the default Prometheus instance should scrape kube-scheduler metrics. | true | false |
| `monitor.kube_proxy` | Controls if the default Prometheus instance should scrape kube-proxy metrics. | true | false |
| `monitor.kubelet` | Controls if the default Prometheus instance should scrape kubelet metrics. | true | false |
| `coredns` | Block, which allows to customize, how CoreDNS is scraped. | - | false |
| `coredns.selector` | Defines, how CoreDNS pods should be selected for scraping. | {"k8s-app":"coredns","tier":"control-plane"} | false |
| `storage_class` | Storage Class to use for the storage allowed for Prometheus and Alertmanager. | - | false |

## Applying

To apply the Prometheus Operator component:

```bash
lokoctl component apply prometheus-operator
```

### Post-installation

To start monitoring your applications running on Kubernetes. Just create a `ServiceMonitor` object
in that namespace which looks like following:

```yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  labels:
    app: openebs
  name: openebs
  namespace: openebs
spec:
  endpoints:
  - path: /metrics
    port: exporter
  namespaceSelector:
    matchNames:
    - openebs
  selector:
    matchLabels:
      openebs.io/cas-type: cstor
```

Change the `labels`, `endpoints`, `namespaceSelector`, `selector` fields as you need. To learn more
about basics of `ServiceMonitor` [read the docs
here](https://github.com/coreos/prometheus-operator/blob/master/Documentation/user-guides/getting-started.md#related-resources)
and [the API Reference can be found
here](https://github.com/coreos/prometheus-operator/blob/master/Documentation/api.md#servicemonitor).

## Deleting

To destroy the component:

```bash
lokoctl component delete prometheus-operator --delete-namespace
```
