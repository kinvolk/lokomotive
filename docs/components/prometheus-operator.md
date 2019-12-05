# Prometheus Operator

## Requirements

- A Kubernetes cluster with a
[PersistentVolume](https://kubernetes.io/docs/concepts/storage/persistent-volumes/)
plugin, e.g. [OpenEBS](/docs/components/openebs/openebs-operator.md) or one of the
[built-in](https://kubernetes.io/docs/concepts/storage/persistent-volumes/#types-of-persistent-volumes)
plugins.

>NOTE: The [OpenEBS component](/docs/components/openebs/openebs-operator.md) provides a
>k8s-native persistent storage solution which can be used on almost any
>infrastructure.

## Installation

#### `.lokocfg` file

Create file that looks like this:

```bash
$ cat prometheus.lokocfg
component "prometheus-operator" {
  namespace              = "monitoring"
  grafana_admin_password = "foobar"

  etcd_endpoints = [
    "10.88.181.1",
  ]

  prometheus_metrics_retention = "14d"
  prometheus_external_url      = "https://api.example.com/prometheus"

  alertmanager_retention    = "360h"
  alertmanager_external_url = "https://api.example.com/alertmanager"
  alertmanager_config       = "${file("alertmanager-config.yaml")}"
}
```

**Note**: Replace values in above file as necessary. You can find more information on the fields in the above file in **[Argument Reference](#argument-reference)**.

#### Prometheus Alertmanager config

Create `alertmanager-config.yaml` file if necessary. It generally looks like below. To know more about prometheus alerting read [here](https://prometheus.io/docs/alerting/configuration/#configuration-file).

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

**Note**: Please make sure it is indented to two spaces.

#### Namespace creation

Replace namespace name from above `*.lokocfg` file in following command:

```bash
kubectl create namespace <namespace>
```

#### Install Prometheus Operator component

To install run:

```bash
lokoctl component install prometheus-operator
```

## Next steps

#### ServiceMonitors

To start monitoring your applications running on Kubernetes. Just create a `ServiceMonitor` object in that namespace which looks like following:

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

Change the `labels`, `endpoints`, `namespaceSelector`, `selector` fields as you need. To know more about basics of `ServiceMonitor` [read the docs here](https://github.com/coreos/prometheus-operator/blob/master/Documentation/user-guides/getting-started.md#related-resources) and [the API Reference can be found here](https://github.com/coreos/prometheus-operator/blob/master/Documentation/api.md#servicemonitor).

#### Access Prometheus dashboard

If you have exposed prometheus dashboard via a URL that is defined by `prometheus_external_url` then visit the URL to access dashboard. But if you haven't then run following command to get access to dashboard:

```bash
kubectl -n monitoring port-forward svc/prometheus-operated 19090:9090
```

Now visit [http://localhost:19090/graph](http://localhost:19090/graph).

## Argument Reference

| Argument | Explanation | Default | Required |
|--------	|--------------|---------|----------|
| `namespace` | Namespace to deploy the prometheus operator into. | - | true |
| `grafana_admin_password` | Password for `admin` user in Grafana.  | - | true |
| `etcd_endpoints` | List of endpoints via which etcd can be reachable from Kubernetes. | [] | false |
| `prometheus_operator_node_selector` | Node selector to specify nodes where the Prometheus Operator pods should be deployed. | {} | false |
| `prometheus_metrics_retention` | Time duration Prometheus shall retain data for. Must match the regular expression `[0-9]+(ms\|s\|m\|h\|d\|w\|y)` (milliseconds seconds minutes hours days weeks years). | `10d` | false |
| `prometheus_external_url` | The external URL the Prometheus instances will be available under. This is necessary to generate correct URLs. This is necessary if Prometheus is not served from root of a DNS name. | "" | false |
| `prometheus_node_selector` | Node selector to specify nodes where the Prometheus pods should be deployed. | {} | false |
| `alertmanager_retention` | Time duration Alertmanager shall retain data for. Must match the regular expression `[0-9]+(ms\|s\|m\|h)` (milliseconds seconds minutes hours). | `120h` | false |
| `alertmanager_external_url` | The external URL the Alertmanager instances will be available under. This is necessary to generate correct URLs. This is necessary if Alertmanager is not served from root of a DNS name. | "" | false |
| `alertmanager_config` | Provide YAML file path to configure Alertmanager. See [https://prometheus.io/docs/alerting/configuration/#configuration-file](https://prometheus.io/docs/alerting/configuration/#configuration-file). | `{"global":{"resolve_timeout":"5m"},"route":{"group_by":["job"],"group_wait":"30s","group_interval":"5m","repeat_interval":"12h","receiver":"null","routes":[{"match":{"alertname":"Watchdog"},"receiver":"null"}]},"receivers":[{"name":"null"}]}` | false |
| `alertmanager_node_selector` | Node selector to specify nodes where the AlertManager pods should be deployed. | {} | false |
