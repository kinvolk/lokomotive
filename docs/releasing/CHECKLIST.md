## Check list

This document shows what actions you should perform before creating a new release. This is a manual process as CI does not test everything.

The following example assumes we’re going from version 0.1.0 (`v0.1.0`) to
0.2.0 (`v0.2.0`).

- Get the latest `lokoctl` binary e.g. from [GitHub](https://github.com/kinvolk/lokomotive/releases)
  and put it into directory where you are going to
  - Copy `lokoctl` binary to your assets directory.

- Deploy `lokomotive` with old release
  - e.g. `./lokoctl cluster apply`
  - Deploy a single cluster (controller node) first.

- Configure env variable `$KUBECONFIG` as per your assets directory location.

- Make sure all nodes healthy and all Pod are running.
  - e.g. `kubectl get po -A && kubectl get no`

- Delete the old `lokoctl` binary.

- Checkout to latest `master` branch. Make sure you have the latest code fetched and rebased.
  - e.g. `git checkout master && git fetch origin && git rebase origin/master`
  - Build lokoctl binary from latest master `make build`

- Copy `lokoctl` binary to your assets directory and re-apply the cluster
  - e.g. `./lokoctl cluster apply`

- Make sure all nodes healthy and all Pod are running.
  - e.g. `kubectl get po -A && kubectl get no`

### Components test

This sections checks if components work as desired.

- Check all certificates are valid
  - e.g. `kubectl get certificates -A`
  - Certificates for all your components are valid.

- Check external IP is assigned to contour service. This will verify that MetalLB is assigning IP to service of type `LoadBalancer`.
  - `kubectl get svc -n projectcontour`

- Check routes are added to AWS for your components. If you have used route53 DNS provider, you can check them [here](https://console.aws.amazon.com/route53/v2/home#Dashboard). Make sure to check the correct hosted zone.

- Check Gangway Ingress Host URL that you have configured works fine.

- Check httpbin Ingress Host URL that you have configured works fine.

- Do some **blackbox testing** by sending HTTP requests through MetalLB + Contour + cert-manager.

- Check metrics for your cluster by going to Prometheus Ingress Host URL.

- Check velero component works fine, by testing it for a namespace.
  - Run the following commands:
  ```sh
  # Create test namespace.
  kubectl create ns test

  # Create a serviceaccount.
  kubectl create sa test

  # Create velero backup.
  velero backup create serviceaccount-backup --include-namespaces test

  # Delete namespace test.
  kubectl delete ns test

  # Restore namespace using velero.
  velero restore create --from-backup serviceaccount-backup

  # Check serviceaccount test exist.
  kubectl get sa test

  ```

- Check web-ui Ingress Host URL that you have configured works fine.

**IMPORTANT**: Follow the whole process again with multi-cluster (controller node).

If everything works fine, continue with the release process.
