provider "aws" {
  version = "~> 2.8.0"
  alias   = "default"
}

provider "ct" {
  version = "0.3.1"
}

provider "local" {
  version = "~> 1.0"
  alias   = "default"
}

provider "null" {
  version = "~> 1.0"
  alias   = "default"
}

provider "template" {
  version = "~> 1.0"
  alias   = "default"
}

provider "tls" {
  version = "~> 1.0"
  alias   = "default"
}

provider "packet" {
  version = "~> 1.2"
  alias   = "default"
}
