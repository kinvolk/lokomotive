resource "matchbox_group" "install" {
  name = format(
    "install-%s",
    var.node_name
  )

  profile = var.cached_install == true ? matchbox_profile.cached-flatcar-linux-install.name : matchbox_profile.flatcar-install.name
  selector = {
    mac = var.node_mac
  }
}

resource "matchbox_group" "node" {
  name = format(
    "%s",
    var.node_name
  )
  profile = matchbox_profile.node.name

  selector = {
    mac = var.node_mac
    os  = "installed"
  }
}

