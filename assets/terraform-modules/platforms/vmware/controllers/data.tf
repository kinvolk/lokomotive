data "vsphere_datacenter" "main" {
  name = var.datacenter
}

data "vsphere_datastore" "main" {
  name          = var.datastore
  datacenter_id = data.vsphere_datacenter.main.id
}

data "vsphere_compute_cluster" "main" {
  name          = var.compute_cluster
  datacenter_id = data.vsphere_datacenter.main.id
}

data "vsphere_network" "main" {
  name          = var.network
  datacenter_id = data.vsphere_datacenter.main.id
}

data "vsphere_virtual_machine" "main_template" {
  name          = var.template
  datacenter_id = data.vsphere_datacenter.main.id
}
