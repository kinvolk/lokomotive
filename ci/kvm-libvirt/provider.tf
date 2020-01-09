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

provider "libvirt" {
  version = "~> 0.6.1"
  uri     = "qemu:///system"
  alias   = "default"
}
