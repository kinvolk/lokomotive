[gangway](https://github.com/heptiolabs/gangway) is a web application
"to easily enable authentication flows via OIDC for a kubernetes cluster."

## Lokomotive component

gangway is available as a component in lokoctl

### Requirements

* An ingress controller such as `ingress-nginx` for HTTP ingress
* A certificate manager such as `cert-manager` for valid certificates
* [Dex](dex.md) installed with a static client for gangway

### Configuration

The gangway lokoctl component currently supports the following options:

```tf
# gangway.lokocfg

variable "gangway_session_key" {
  type = "string"
}

component "gangway" {
  # The name of the cluster. This is used to name the kubectl configuration context.
  cluster_name = "example"

  # Used as the `hosts` domain in the ingress resource for gangway that is
  # automatically created
  ingress_host = "gangway.example.lokomotive-k8s.org"

  session_key = "${var.gangway_session_key}"

  # Where kube-apiserver is reachable
  api_server_url = "https://example.lokomotive-k8s.org:6443"

  # Where the 'auth' endpoint is
  authorize_url = "https://dex.example.lokomotive-k8s.org/auth"

  # Where the 'token' endpoint is
  token_url = "https://dex.example.lokomotive-k8s.org/token"

  # The static client id and secret
  client_id     = "${var.dex_static_client_gangway_id}"
  client_secret = "${var.dex_static_client_gangway_secret}"

  # gangway's redirect URL, i.e. where the OIDC endpoint should callback to
  redirect_url = "${var.gangway_redirect_url}"
}
```

The secrets can be defined in another file (`lokocfg.vars`) like following:

```tf
gangway_redirect_url         = "https://gangway.example.lokomotive-k8s.org/callback"

# A random secret key (create one with `openssl rand -base64 32`)
gangway_session_key              = "5Rsz5C4qRqYFoAfYcXOedQOyQpHTXyLiWFYvtjwjtm0="
dex_static_client_gangway_secret = "2KBvQkjOZdc3iHt4KSb9GUECdenH/VDl04TwMdSyPcs="
dex_static_client_gangway_id     = "gangway"
```

### Installation

After preparing your configuration in a lokocfg file, you can install
gangway with

```
lokoctl component install gangway
```

gangway should be available under the configured `ingress_host` domain
shortly after.
