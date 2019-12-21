resource "matchbox_group" "install" {
  count = length(var.controller_names) + length(var.worker_names)

  name = format(
    "install-%s",
    concat(var.controller_names, var.worker_names)[count.index]
  )

  profile = local.flavor == "flatcar" ? var.cached_install == "true" ? matchbox_profile.cached-flatcar-linux-install[count.index].name : matchbox_profile.flatcar-install[count.index].name : var.cached_install == "true" ? matchbox_profile.cached-container-linux-install[count.index].name : matchbox_profile.container-linux-install[count.index].name

  selector = {
    mac = concat(var.controller_macs, var.worker_macs)[count.index]
  }
}

resource "matchbox_group" "controller" {
  count = length(var.controller_names)
  name = format(
    "%s-%s",
    var.cluster_name,
    var.controller_names[count.index]
  )
  profile = matchbox_profile.controllers[count.index].name

  selector = {
    mac = var.controller_macs[count.index]
    os  = "installed"
  }
}

resource "matchbox_group" "worker" {
  count = length(var.worker_names)
  name = format(
    "%s-%s",
    var.cluster_name,
    var.worker_names[count.index]
  )
  profile = matchbox_profile.workers[count.index].name

  selector = {
    mac = var.worker_macs[count.index]
    os  = "installed"
  }
}
