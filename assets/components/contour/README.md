# Contour manifests

The files in `manifest-deployment` were copied from:
	https://github.com/heptio/contour/tree/v0.14.0/examples/deployment-grpc-v2

And files from `manifest-daemonset` were copied from:
	https://github.com/heptio/contour/tree/v0.14.0/examples/ds-grpc-v2

The service on both directories was manually modified to:
 * [Use externalTrafficPolicy: Local](#Use-externalTrafficPolicy-local)

You should make sure to **NOT** delete them by mistake (e.g. when updating Contour
version).

To upgrade the Contour version you need to repeat that: copy & paste from that
directory from the release you want and manually keep the modifications to the
service.  (HINT: `git checkout -p` may help you to spot the deleted manual
modifications after the copy & paste).

## Use externalTrafficPolicy: Local

This setting just makes nodes running Contour pods to receive traffic for the
Contour service. This is needed on several setups to preserve the client source
IP.
