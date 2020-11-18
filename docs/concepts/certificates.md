---
title: Certificates
weight: 10
---

# Introduction

Lokomotive takes care of generating [Kubernetes PKI](https://kubernetes.io/docs/setup/best-practices/certificates/)
on behalf of the user.

Generation of certificates is currently done in [bootkube Terraform module](https://github.com/kinvolk/lokomotive/tree/v0.5.0/assets/terraform-modules/bootkube).

Following certificates, private and public keys are generated:

- Admission Webhook Server private key
- Admission Webhook Server certificate

- Aggregation CA private key
- Aggregation CA certificate

- Aggregation Client private key
- Aggregation Client certificate

- etcd CA private key
- etcd CA certificate

- etcd client private key
- etcd client certificate

- etcd server private key
- etcd server certificate

- etcd peer private key
- etcd peer certificate

- Kubernetes CA private key
- Kubernetes CA certificate

- kube-apiserver private key
- kube-apiserver certificate

- Kubernetes admin private key
- Kubernetes admin certificate

- service account private key
- service account public key

# Manual certificate rotation

This section describes how each certificate can be rotated manually to extend it's expiry date.

## Rotating Admission Webhook Server certificate

Admission Webhook Server certificate is used in Lokomotive Admission Webhook Server.

When it expires, `admission-webhook-server` `MutatingWebhookConfiguration` will no longer be
functional, which will break API calls for whatever webhook handles.

### If certificate already exired

If the certificate already expired, `admission-webhook-server` `MutatingWebhookConfiguration` will no longer be
functional, which will break API calls for whatever webhook handles.

It is recommended to temporarily remove the webhook to make sure Helm is able to upgrade control-plane resources.

To rotate this certificate, execute the following commands:

```sh
export KUBECONFIG=$ASSET_DIR/cluster-assets/auth/kubeconfig
kubectl delete MutatingWebhookConfiguration admission-webhook-server
cd $ASSET_DIR/terraform
terrafor taint module.aws-l8e-mat.module.bootkube.tls_locally_signed_cert.admission-webhook-server
cd ../..
lokoctl cluster apply --skip-pre-update-health-check
```

### If certificate did not expire yet

If the certificate did not expire yet and should be rotated, one should taint certificate resource
using Terraform commands shown below:

```sh
cd $ASSET_DIR/terraform
terrafor taint module.aws-l8e-mat.module.bootkube.tls_locally_signed_cert.admission-webhook-server
cd ../..
lokoctl cluster apply
```

`lokoctl` should now re-create the certificate using Terraform and update it as part of control-plane
upgrade process.


## Rotating Aggregation CA private key

Aggregation CA certificate is used in kube-apiserver and in extension API servers like `metrics-server.

### If certificate already expired

If CA certificate already expired, extension API servers should no longer be able to serve the requests,
**if they are configured to validate incoming requests**.

If they are configured to do so, they will read Aggregation CA certificate from `extension-apiserver-authentication`
`ConfigMap` in `kube-system` namespace, which is managed by `kube-apiserver`.

#### Symptoms of expired Aggregation CA certificate

For example, when Aggregation CA certificate expires, you should see the following:

```console
$ kubectl top nodes
Error from server (ServiceUnavailable): the server is currently unable to handle the request (get nodes.metrics.k8s.io)
```

Then, in this case, in `metrics-server` logs, you should find logs similar to the following:

```console
E1119 09:47:18.072393       1 authentication.go:53] Unable to authenticate the request due to an error: [x509: certificate signed by unknown authority (possibly because of "crypto/rsa: verification error" while trying to verify candidate authority certificate "kubernetes-front-proxy-ca"), x509: certificate signed by unknown authority]
```

#### Identifying extension API servers

To get a list of affected services, run the following command:

```sh
kubectl get APIService | grep -v Local
```

Sample output:

```console
$ kubectl get APIService | grep -v Local
NAME                                    SERVICE                      AVAILABLE   AGE
v1beta1.metrics.k8s.io                  kube-system/metrics-server   True        3d3h
```

If the certificate already expired, servers validing client certificates will no longer be functional and
servers which do not validate it will remain functional, so there is no need to perform 2-steps rotation process
like in case when certificate did not expire yet.

#### Generating new certificates

The rotation can be started by executing the following commands:

```sh
cd $ASSET_DIR/terraform

# Re-create CA certificate with new private key.
terrafor taint module.aws-l8e-mat.module.bootkube.tls_private_key.aggregation-ca
terraform apply -target module.aws-l8e-mat.module.bootkube.tls_self_signed_cert.aggregation-ca

cd $CONFIG_DIR

lokoctl cluster apply
```

`lokoctl` should now re-create following resources:
- Aggregation CA Certificate
- Aggregation CA private key
- Aggregation client certificate

Once finished, `kube-apiserver` should get restarted and `extension-apiserver-authentication` `ConfigMap` should include new CA certificate.

#### Restaring extension API servers

Next step is to deploy it to all consuming clients.

To get a list of affected services, run the following command:

```sh
kubectl get APIService | grep -v Local
```

Sample output:

```console
$ kubectl get APIService | grep -v Local
NAME                                    SERVICE                      AVAILABLE   AGE
v1beta1.metrics.k8s.io                  kube-system/metrics-server   True        3d3h
```

Then, find all pods which are endpoints of services listed in `SERVICE` column and make sure they picked
up new CA certificate. You can check the logs if they did. If you are not sure, it is recommended to restart
all of them to make sure they have new CA certificate available. Regstarting should be done using example command:

```sh
kubectl rollout restart deployment/metrics-server
```

This command should gracefully restart all pods which serve mentioned `Service`.

Once finished, all extension API services should be functional again.

### If certificate did not expire yet

If certificate did not expire yet and should be rotated, one should start by running the commands below.

Following resources will be re-created in this procedure:
- Aggregation CA Certificate
- Aggregation CA private key
- Aggregation client certificate

```sh
cd $ASSET_DIR/terraform

# Backup old CA certificate.
cp ../cluster-assets/tls/aggregation-ca.crt ./aggregation-ca.crt

# Re-create CA certificate with new private key.
terraform taint module.aws-l8e-mat.module.bootkube.tls_private_key.aggregation-ca[0]
terraform apply -target module.aws-l8e-mat.module.bootkube.local_file.aggregation-ca-crt

# Create a CA bundle including old and new certificate.
cat ../cluster-assets/tls/aggregation-ca.crt >> ./aggregation-ca.crt

# Make sure we speak to the right cluster.
export KUBECONFIG=$ASSET_DIR/cluster-assets/auth/kubeconfig

# Upgrade kube-apiserver to include new CA certificate.
helm upgrade -n kube-system kube-apiserver --reuse-values --set apiserver.aggregationCaCert=$(cat ./aggregation-ca.crt | base64 -w0) --wait ../cluster-assets/charts/kube-system/kube-apiserver --debug
```

#### Failing "helm upgrade"

If `helm upgrade` command fails, try running it again.

If you get error like the following:

```console
Error: UPGRADE FAILED: another operation (install/upgrade/rollback) is in progress
```

Check latest revision of `kube-apiserver` release using the following command:

```sh
helm history kube-apiserver
```

Sample output:

```console
$ helm history kube-apiserver
REVISION        UPDATED                         STATUS          CHART                   APP VERSION     DESCRIPTION
1               Wed Nov 18 17:55:44 2020        deployed        kube-apiserver-0.1.2    v1.19.3         Install complete
2               Wed Nov 18 19:30:30 2020        pending-upgrade kube-apiserver-0.1.2    v1.19.3         Preparing upgrade
```

In this case, latest revision is `2`, so run the following command:

```sh
helm rollback kube-apiserver 2 --wait --debug
```

#### Identifying extension API servers

Now that `extension-apiserver-authentication` `ConfigMap` includes new CA certificate, we need to deploy it to all consuming clients.

To get a list of affected services, run the following command:

```sh
kubectl get APIService | grep -v Local
```

Sample output:

```console
$ kubectl get APIService | grep -v Local
NAME                                    SERVICE                      AVAILABLE   AGE
v1beta1.metrics.k8s.io                  kube-system/metrics-server   True        3d3h
```

#### Checking dynamic CA certificate loading

Then, find all pods which are endpoints of services listed in `SERVICE` column and make sure they picked
up new CA certificate.

You can check the logs if they did. Usually if daemon has set log level to at least `-v=2`, you should see message similar to this after `kube-apiserver` got updated:

```console
I1119 09:40:10.092021       1 tlsconfig.go:178] loaded client CA [0/"client-ca::kube-system::extension-apiserver-authentication::client-ca-file,client-ca::kube-system::extension-apiserver-authentication::requestheader-client-ca-file"]: "kubernetes-ca" [] groups=[bootkube] issuer="<self>" (2020-11-18 17:50:43 +0000 UTC to 2021-11-18 17:50:43 +0000 UTC (now=2020-11-19 09:40:10.092007151 +0000 UTC))
I1119 09:40:10.092059       1 tlsconfig.go:178] loaded client CA [1/"client-ca::kube-system::extension-apiserver-authentication::client-ca-file,client-ca::kube-system::extension-apiserver-authentication::requestheader-client-ca-file"]: "kubernetes-front-proxy-ca" [] issuer="<self>" (2020-11-18 18:29:08 +0000 UTC to 2021-11-18 18:29:08 +0000 UTC (now=2020-11-19 09:40:10.092044014 +0000 UTC))
I1119 09:40:10.092116       1 tlsconfig.go:178] loaded client CA [2/"client-ca::kube-system::extension-apiserver-authentication::client-ca-file,client-ca::kube-system::extension-apiserver-authentication::requestheader-client-ca-file"]: "kubernetes-front-proxy-ca" [] issuer="<self>" (2020-11-19 09:36:58 +0000 UTC to 2021-11-19 09:36:58 +0000 UTC (now=2020-11-19 09:40:10.092099451 +0000 UTC))
```

#### Restarting extension API servers

If you are not sure, it is recommended to restart
all of them to make sure they have new CA certificate available.

You can find appropriate `Deployment` or `DaemonSet` objects by listing the `Service` object first using command like:

```sh
kubectl get service -n kube-system metrics-server -o wide
```

Sample output:

```console
$ kubectl get svc -n kube-system metrics-server -o wide
NAME             TYPE        CLUSTER-IP   EXTERNAL-IP   PORT(S)   AGE   SELECTOR
metrics-server   ClusterIP   10.3.201.0   <none>        443/TCP   38m   app=metrics-server,release=metrics-server
```

This should include `SELECTOR` command, which can possibly be used to find associated `Deployment` or `DaemonSet` objects.

To find `Deployment`, run command similar to the following:

```sh
kubectl get deploy -l app=metrics-server,release=metrics-server
```

To find `DaemonSet`, run command similar to the following:
```sh
kubectl get daemonset -l app=metrics-server,release=metrics-server
```

Then, restarting should be done using example command:

```sh
kubectl rollout restart deployment/metrics-server
```

This command should gracefully restart all pods which serve mentioned `Service`.

You can watch restarting process using the following command:

```sh
kubectl rollout status deployment/metrics-server --watch
```

This process should be repeated for all APIService objects listed above.

#### Generating new extension API client certificate for kube-apiserver

Once you restarted all services, you can now create new client certificate for `kube-apiserver` signed by new CA.
This can be done using the following commands:

```sh
cd $ASSET_DIR/terraform
# Re-create client certificate.
terraform taint module.aws-l8e-mat.module.bootkube.tls_locally_signed_cert.aggregation-client[0]
terraform apply -target module.aws-l8e-mat.module.bootkube.local_file.aggregation-client-crt

# Make sure we speak to the right cluster.
export KUBECONFIG=$ASSET_DIR/cluster-assets/auth/kubeconfig

# Upgrade kube-apiserver to use new client certificate signed by new CA certificate.
helm upgrade -n kube-system kube-apiserver --reuse-values --set apiserver.aggregationClientCert=$(cat ../cluster-assets/tls/aggregation-client.crt | base64 -w0) --wait ../cluster-assets/charts/kube-system/kube-apiserver --debug
```

Now, you can run `lokoctl cluster apply` to make sure everything is up to date:

```sh
lokoctl cluster apply --skip-pre-update-health-check
```

#### (Optional) Removing old extension API CA certificate from trusted bundle

Once finished, you can optionally remove old CA certificate from trusted bundle and restart all extension
API servers to pick up the update. This step is optional, though highly recommended.

Start by removing authentication `ConfigMap` object, as `kube-apiserver` does not sync the CA certificates
there it seems, only adds new ones. Removing can be done by executing the following command:

```sh
kubectl delete configmap -n kube-system extension-apiserver-authentication
```

Then restart `kube-apiserver` to re-generate the `ConfigMap` using the following command:

```sh
kubectl rollout restart ds -n kube-system kube-apiserver
```

Wait for restart process to finish using the following command:

```sh
kubectl rollout status -w ds -n kube-system kube-apiserver
```

Finally, restart all extension API servers as explained above.
