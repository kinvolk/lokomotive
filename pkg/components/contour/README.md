# Contour manifests

The files in `manifest-deployment` were copied from:
	https://github.com/heptio/contour/tree/v0.11.0/deployment/deployment-grpc-v2

And files from `manifest-daemonset` were copied from:
	https://github.com/heptio/contour/tree/v0.11.0/deployment/ds-grpc-v2

The service on both directories was manually modified to:
 * [Add prometheus scrape config](#Use-statsprometheus)
 * [Use externalTrafficPolicy: Local](#Use-externalTrafficPolicy-local)

The pod spec was modified to:
 * Use `/stats/prometheus` instead of `/stats/` for `prometheus.io/path`
   [annotation](#Use-statsprometheus).
 * [Use a custom Contour docker image to fix a OOM issue on envoy](#Custom-Contour-docker-image)
 * [Use envoy 1.10](#Envoy-1.10)

You should make sure to **NOT** delete them by mistake (e.g. when updating Contour
version).

To upgrade the Contour version you need to repeat that: copy & paste from that
directory from the release you want and manually keep the modifications to the
service.  (HINT: `git checkout -p` may help you to spot the deleted manual
modifications after the copy & paste).

## Use /stats/prometheus

This change was merged upstream and will be available in Contour >= 0.12.  So
this changes can be soon be dropped. See:
https://github.com/heptio/contour/pull/1036

## Custom Contour docker image

We use a custom docker image to fix envoy consuming loads of RAM
when running under high traffic.

The image is based on Contour 0.12-dirty (i.e. before v0.12 was
released) and it just contains this simple patch:
https://github.com/johananl/contour/commit/bece50124c5398ead56e24e4bb3dcd3a2ff51035

The issue is that connections stay open until the client closes them (if
he does it at all). Under high traffic load, this ends with tons of open
connections and LOT of RAM usage. In fact, in the scenarios we've
observed, the RAM usage as well as the opened connections were constantly
inreasing until envoy run out of RAM and crashes.

This behaviour of not closing open connections is the default behaviour
of the envoy http connection manager and there is is a parameter on
envoy to tune when an idle connection will be closed (`idle_timeout`):

https://www.envoyproxy.io/docs/envoy/latest/api-v2/config/filter/network/http_connection_manager/v2/http_connection_manager.proto.html?highlight=idle_timeout

When setting this parameter on envoy, connections are closed and the RAM
usage problem is solved immediately.

However, as Contour does not expose a way to tune this envoy parameter,
the patch just mentioned hardcodes the `idle_timeout` param with a value
of 60s.

tl;dr: With this patch, the mem problem is solved immediately and
mem remains at constant usage even under heavy traffic load.

We already talked to upstream and the proper fix will be discussed on a
github issue.

## Envoy 1.10

We use envoy 1.10 because it exports more metrics in prometheus format.

Envoy, in PR https://github.com/envoyproxy/envoy/pull/5601, added
support for histograms in prometheus format and the first release to
include it is 1.10. This simplifies scraping data from prometheus and
several important metrics, like latency metrics, are exported. This is a
huge advantage from a monitoring POV.

Although Contour is released with envoy 1.9.1, it is safe to upgrade to
envoy 1.10. Also, we removed a deprecated param that Contour was
using with envoy 1.9 and is completely removed in envoy 1.10. The patch
was merged in Contour upstream:
https://github.com/heptio/contour/commit/b8e8e65e312ffeb4f41dab1af7b89a250c35026f


## Use externalTrafficPolicy: Local

This setting just makes nodes running Contour pods to receive traffic for the
Contour service. This is needed on several setups to preserve the client source
IP.
