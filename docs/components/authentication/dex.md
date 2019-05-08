[Dex](https://github.com/dexidp/dex) is an OpenID Connect (OIDC) and OAuth 2.0
provider with connectors to many other identity providers such as GitHub,
Google or LDAP.

"Dex acts as a portal to other identity providers through "connectors." This
lets dex defer authentication to LDAP servers, SAML providers, or established
identity providers like GitHub, Google, and Active Directory. Clients write
their authentication logic once to talk to dex, then dex handles the protocols
for a given backend."

In Kubernetes context, dex enables

* usage of authentication providers that don't support OIDC itself and
* grouping of multiple authentication providers.

## Lokomotive component

Dex is available as a component in lokoctl.

### Requirements

* An ingress controller such as `ingress-nginx` for HTTP ingress
* A certificate manager such as `cert-manager` for valid certificates

Both `ingress-nginx` and `cert-manager` are available as lokoctl components.

### Configuration

The dex lokoctl component currently supports the following options:

```
# dex.lokocfg

component "dex" {
  # Used as the `hosts` domain in the ingress resource for dex that is
  # automatically created
  ingress_host = "dex.example.lokomotive-k8s.org"

  # Used as the `issuer` URL
  issuer_host = "https://dex.example.lokomotive-k8s.org"

  # You can configure one or more connectors. Currently only GitHub and
  # OIDC (for example with Google) are supported from lokoctl.

  # A GitHub connector
  # Requires GitHub OAuth app credentials from https://github.com/settings/developers
  connector "github" {
    id = "github"
    name = "GitHub"

    config {
      # The OAuth app client id
      client_id = "${var.github_client_id}"
      # The OAuth app client secret
      client_secret = "${var.github_client_secret}"
      # The authorization callback URL as configured with GitHub
      # (i.e. where your dex instance is reachable)
      redirect_uri = "https://dex.example.lokomotive-k8s.org/callback"

      # Can be 'name', 'slug' or 'both', see https://github.com/dexidp/dex/blob/master/Documentation/connectors/github.md
      team_name_field = "slug"

      # Define one or more organisations and teams
      org {
        name = "kinvolk"
        teams = [
          "lokomotive-developers",
        ]
      }
    }
  }

  # A OIDC connector
  # Here configured for use with Google
  connector "oidc" {
    id = "google"
    name = "Google"

    config {
      # With Google, the OAuth app credentials can be created in the
      # cloud console via
      # APIs & Services -> Credentials ->
      # Create credentials -> OAuth client id -> Web application

      # The OAuth app client id
      client_id = "${var.google_client_id}"
      # The OAuth app client secret
      client_secret = "${var.google_client_secret}"

      # The authorization callback URL
      redirect_uri = "https://dex.example.lokomotive-k8s.org/callback"

      # The OIDC issuer endpoint
      issuer = "https://accounts.google.com"
    }
  }

  # You can configure one or more static clients, i.e. apps that use
  # dex (https://github.com/dexidp/dex/blob/master/Documentation/using-dex.md#configuring-your-app).
  # If you use for example gangway to drive authentication flows,
  # the config would look like the following snippet:
  static_client {
    id = "gangway"
    name = "gangway"

    redirect_uris = [
      "https://gangway.example.lokomotive-k8s.org/callback",
    ]
    secret = "${var.dex_static_client_gangway_secret}"
  }
}
```

### Installation

After preparing your configuration in a lokocfg file (e.g. `dex.lokocfg`), you
can install the component with

```
lokoctl component install dex
```

dex pods run in the dex namespace, see `kubectl get pods -n dex`.

You can verify that dex is up and running by sending a HTTP request to

```
curl https://dex.example.lokomotive-k8s.org/.well-known/openid-configuration
{
  "issuer": "https://dex.example.lokomotive-k8s.org",
...
}
```

To actually make the kube-apiserver use dex as an OIDC authenticator,
you need to [reconfigure it](https://kubernetes.io/docs/reference/access-authn-authz/authentication/#configuring-the-api-server).
This can be done by editing the `kube-apiserver` daemonset as follows:

```
kubectl -n kube-system edit ds kube-apiserver
```

Add the following parameters:

```
--oidc-issuer-url=https://dex.example.lokomotive-k8s.org
--oidc-client-id=gangway
--oidc-username-claim=email
--oidc-groups-claim=groups
```

It's important that `--oidc-client-id` matches the client id that your
tokens are issued for.

It takes a few minutes for the kube-apiserver daemonset to be restarted
and updated. You can watch

```
kubectl -n kube-system get pods
```

## Next steps

Once you have successfully installed and configured dex, you can use it
to authenticate users.

Note that users need a dex client in order to retrieve a JWT token to
authenticate. We recommend using [gangway](gangway.md), as it enables
users to easily self-configure their kubectl configuration with the right
parameters.

By default, users or user groups don't have any
permissions on the cluster and they need to be authorized through
[RBAC](https://kubernetes.io/docs/reference/access-authn-authz/rbac/).

Examples:

```
# Make members of a group 'cluster-admin'
kubectl create clusterrolebinding cluster-admin-lokomotive-developers --clusterrole cluster-admin --group='kinvolk:lokomotive-developers'

# Give a user 'edit' access
kubectl create clusterrolebinding jane-edit --clusterrole edit --user='jane@example.com'
```
