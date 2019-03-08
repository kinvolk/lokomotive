The files in `manifest-deployment` were copied from:
	https://github.com/heptio/contour/tree/v0.11.0/deployment/deployment-grpc-v2

And files from `manifest-daemonset` were copied from:
	https://github.com/heptio/contour/tree/v0.11.0/deployment/ds-grpc-v2


The service on both directories was manually modified to:
 * Add prometheus scrape config [1]
 * Use externalTrafficPolicy: Local

And the pod spec to use `/stats/prometheus` instead of `/stats/` for
`prometheus.io/path` annotation. [1]

You should make sure to **NOT** delete them by mistake (e.g. when updating contour
version).

To upgrade the contour version you need to repeat that: copy & paste from that
directory from the release you want and manually keep the modifications to the
service.  (HINT: `git checkout -p` may help you to spot the deleted manual
modifications after the copy & paste).


[1]: This changes were merged upstream and will be available in contour >= 0.12.
So this changes can be soon be dropped. See:
https://github.com/heptio/contour/pull/1036
