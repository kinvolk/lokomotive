---
title: Rotate cluster certificates
weight: 10
---

## Introduction

Kubernetes uses PKI certificates for authentication over TLS.
Lokomotive generates the required certificates automatically and the Certificate Authority (CA) has an expiration date of 1 year by default.
To continue operating the cluster, certificates need to be rotated before their expiration date.

This document provides a step by step guide on rotating the cluster certificates.

## Prerequisites

* A Lokomotive cluster accessible via `kubectl`
* The openssl CLI tool

## Steps

### Step 1: Check current CA expiration date

Find out the address of the cluster:

```
kubectl cluster-info
```

Check expiration date of the cluster CA certificate:

```
openssl s_client -connect cluster.example.com:6443 -servername cluster.example.com 2> /dev/null | openssl x509 -noout  -dates
```

The output will be similar to the following:

```
notBefore=May 16 15:13:58 2020 GMT
notAfter=May 16 15:13:58 2021 GMT
```

The date in the `notAfter` line is the expiration date of the cluster CA certificate.

## Step 2: Rotate certificates

Run the lokoctl certificate rotation command:

```
lokoctl cluster certificate rotate
```

Lokomotive will make sure your cluster is up to date and will start the certificate rotation process.
This process takes about 20 minutes and will restart the cluster control plane
components several times so you might lose access to the cluster in a non-HA setup.

## Step 3: Check new CA expiration date

Run the same command as in Step 1 and check the CA certificate has a new expiration date 1 year from now:

```
openssl s_client -connect cluster.example.com:6443 -servername cluster.example.com 2> /dev/null | openssl x509 -noout  -dates
```

Assuming we rotated certificates on May 12 2021, the output should be similar
to the following:

```
notBefore=May 12 09:13:58 2021 GMT
notAfter=May 12 09:13:58 2022 GMT
```

Note the expiration date is one year after the time we did the rotation.
