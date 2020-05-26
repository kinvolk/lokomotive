# Setting up monitoring on Lokomotive

## Contents

* [Introduction](#introduction)
* [Prerequisites](#prerequisites)
* [Deploy Prometheus Operator](#deploy-prometheus-operator)
  * [Configure Prometheus Operator](#configure-prometheus-operator)
  * [Install Prometheus Operator](#install-prometheus-operator)
* [Accessing Prometheus, Alertmanager and Grafana](#accessing-prometheus-operator-sub-components)
  * [Accessing Prometheus](#accessing-prometheus)
  * [Accessing Alertmanager](#accessing-alertmanager)
  * [Accessing Grafana](#accessing-grafana)
    * [Using port forward](#using-port-forward)
    * [Using Ingress](#using-ingress)
* [Additional resources](#additional-resources)

## Introduction

This guide provides the steps for deploying a monitoring stack using the `prometheus-operator` Lokomotive component and explains how to access Prometheus, Alertmanager and Grafana.

## Prerequisites

* A Lokomotive cluster deployed on a supported provider and accessible via `kubectl`.

<!---
TODO: Once we have tutorials on how to deploy and configure Rook or OpenEBS, point the following to those tutorials.
-->

* A storage provider component (`rook` and `rook-ceph`, or `openebs-operator` and `openebs-storage-class`) deployed with a default storage class that can provision volumes for the [PVCs](https://kubernetes.io/docs/concepts/storage/persistent-volumes/#persistentvolumeclaims) created by Alertmanager and Prometheus.

<!---
TODO: Once we have a tutorial on how to deploy and configure Contour and cert-manager, point the following to that tutorial.
-->
> **NOTE**: If you wish to [expose Grafana to the public internet](#using-ingress), the following Lokomotive components should be installed:
> * [`metallb`](./ingress-with-contour-metallb.md) (only on Packet and bare-metal)
> * [`contour`](../configuration-reference/components/contour.md)
> * [`cert-manager`](../configuration-reference/components/cert-manager.md)


## Deploy Prometheus Operator

### Configure Prometheus Operator

Create a file named `monitoring.lokocfg` with the following contents:

```tf
component "prometheus-operator" {}
```

For information about all the available configuration options for the `prometheus-operator` component, visit the component's [configuration reference](../configuration-reference/components/prometheus-operator.md).

### Install Prometheus Operator

Execute the following command to deploy the `prometheus-operator` component:

```bash
lokoctl component apply prometheus-operator
```

Verify the pods in the `monitoring` namespace are in the `Running` state (this may take a few minutes):

```bash
kubectl -n monitoring get pods
```

## Accessing Prometheus, Alertmanager and Grafana

### Accessing Prometheus

Execute the following command to forward port `9090` locally to the Prometheus pod:

```bash
kubectl -n monitoring port-forward svc/prometheus-operator-prometheus 9090
```

Open the following URL: [http://localhost:9090](http://localhost:9090).

### Accessing Alertmanager

Execute the following command to forward port `9093` locally to the Alertmanager pod:

```bash
kubectl -n monitoring port-forward svc/prometheus-operator-alertmanager 9093
```

Open the following URL: [http://localhost:9093](http://localhost:9093).

### Accessing Grafana

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

## Additional resources

- `prometheus-operator` component [configuration reference](../configuration-reference/components/prometheus-operator.md).
