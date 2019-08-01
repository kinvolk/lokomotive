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

* An ingress controller such as Contour for HTTP ingress
* A certificate manager such as cert-manager for valid certificates

Both Contour and cert-manager are available as lokoctl components.

### Configuration

The dex lokoctl component currently supports the following options:

```tf
# dex.lokocfg

variable "google_client_id" {
  type = "string"
}

variable "google_client_secret" {
  type = "string"
}

variable "dex_static_client_gangway_id" {
  type = "string"
}

variable "dex_static_client_gangway_secret" {
  type = "string"
}

variable "gangway_redirect_url" {
  type = "string"
}

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
      # follow: https://developers.google.com/adwords/api/docs/guides/authentication#webapp

      # The OAuth app client id
      client_id = "${var.google_client_id}"
      # The OAuth app client secret
      client_secret = "${var.google_client_secret}"

      # The authorization callback URL
      # Authorize this redirect URL while creating above credentials in the
      # Restrictions -> Authorized redirect URIs
      redirect_uri = "https://dex.example.lokomotive-k8s.org/callback"

      # The OIDC issuer endpoint
      issuer = "https://accounts.google.com"
    }
  }

  # A Google native connector
  connector "google" {
    id   = "google"
    name = "Google"

    config {
      # With Google, the OAuth app credentials can be created in the
      # cloud console via
      # APIs & Services -> Credentials ->
      # Create credentials -> OAuth client id -> Web application
      # follow: https://developers.google.com/adwords/api/docs/guides/authentication#webapp

      # The OAuth app client id
      client_id = "${var.google_client_id}"

      # The OAuth app client secret
      client_secret = "${var.google_client_secret}"

      # The authorization callback URL
      # Authorize this redirect URL while creating above credentials in the
      # Restrictions -> Authorized redirect URIs
      redirect_uri = "https://dex.example.lokomotive-k8s.org/callback"

      # This should be the email of a GSuite super user. The service account
      # you created earlier will impersonate this user when making calls to
      # the admin API.
      admin_email = "foobar@example.io"
    }
  }
  # only to be defined with Google connector
  # Path to the Gsuite Service Account JSON file, more information at the end of
  # this file in [G Suite specific instructions](#g-suite-specific-instructions)
  gsuite_json_config_path = "project-testing-123456-er12t34y56ui.json"


  # You can configure one or more static clients, i.e. apps that use
  # dex (https://github.com/dexidp/dex/blob/master/Documentation/using-dex.md#configuring-your-app).
  # If you use for example gangway to drive authentication flows,
  # the config would look like the following snippet:
  static_client {
    name   = "gangway"
    id     = "${var.dex_static_client_gangway_id}"
    secret = "${var.dex_static_client_gangway_secret}"

    redirect_uris = ["${var.gangway_redirect_url}"]
  }
}
```

The secrets can be defined in another file (`lokocfg.vars`) like following:

```tf
google_client_id     = "1234567890123-SqDIX1KFvKPYmuV9Sa8eL92cvxtS3TuP.apps.googleusercontent.com"
google_client_secret = "63zYPITtigLxLaYBEjNP9Taw"

# A random secret key (create one with `openssl rand -base64 32`)
dex_static_client_gangway_secret = "2KBvQkjOZdc3iHt4KSb9GUECdenH/VDl04TwMdSyPcs="
dex_static_client_gangway_id     = "gangway"

gangway_redirect_url = "https://gangway.example.lokomotive-k8s.org/callback"
```

**Note**: More information on the variables used in above dex config can be
found in the [gangway doc](gangway.md#configuration).

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

## Detailed Information

### G Suite specific instructions

You need to create a service account on your google suite account and authorize
it to view groups on your domain.

#### Perform G Suite Domain-Wide Delegation of Authority

- Follow instructions [here](https://developers.google.com/admin-sdk/directory/v1/guides/delegation)
to create service account.

- During Service Account creation a JSON file will be downloaded, give the path
of this file in dex's config for field `gsuite_json_config_path`.

- While **delegating domain-wide authority to your service account** you will be
asked to assign scope in that field select scope
**`https://www.googleapis.com/auth/admin.directory.group.readonly`** only.

#### Enable admin SDK

Admin SDK lets administrators of enterprise domains to view and manage resources
like user, groups etc. To enable it [click here](https://console.developers.google.com/apis/library/admin.googleapis.com/).
