provider "matchbox" {
  version     = "0.3.0"
  ca          = "${file(pathexpand("~/pxe-testbed/.matchbox/ca.crt"))}"
  client_cert = "${file(pathexpand("~/pxe-testbed/.matchbox/client.crt"))}"
  client_key  = "${file(pathexpand("~/pxe-testbed/.matchbox/client.key"))}"
  endpoint    = "matchbox.example.com:8081"
}

provider "ct" {
  version = "0.4.0"
}

provider "local" {
  version = "~> 1.2"
  alias   = "default"
}

provider "null" {
  version = "~> 2.1"
  alias   = "default"
}

provider "template" {
  version = "~> 2.1"
  alias   = "default"
}

provider "tls" {
  version = "~> 2.0"
  alias   = "default"
}
