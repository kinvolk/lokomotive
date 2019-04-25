After installing Lokomotive, you can find a 'cluster-admin' kubeconfig
file in `<asset_dir>/auth/kubeconfig`. It is not advised to share
that file with other users.

Instead, we recommend to use an authentication service and service accounts.
While you are free to choose how you authenticate your users (an overview
can be found
[here](https://kubernetes.io/docs/reference/access-authn-authz/authentication/)),
Lokomotive provides [dex](dex.md) and [gangway](gangway.md) components for
authentication of "normal users" via OpenID Connect (OIDC).
