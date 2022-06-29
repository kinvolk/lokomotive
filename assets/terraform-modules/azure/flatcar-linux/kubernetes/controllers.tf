
locals {
  # Container Linux derivative
  # flatcar-stable -> Flatcar Linux Stable
  channel = split("-", var.os_image)[1]
}

# Controller availability set to spread controllers
resource "azurerm_availability_set" "controllers" {
  resource_group_name = azurerm_resource_group.cluster.name

  name                         = "${var.cluster_name}-controllers"
  location                     = var.region
  platform_fault_domain_count  = 2
  platform_update_domain_count = 4
  managed                      = true
}

# data "azurerm_image" "custom" {
#   name                = var.custom_image_name
#   resource_group_name = var.custom_image_resource_group_name
# }

# Controller instances
resource "azurerm_linux_virtual_machine" "controllers" {
  count               = var.controller_count
  resource_group_name = azurerm_resource_group.cluster.name

  name                = "${var.cluster_name}-controller-${count.index}"
  location            = var.region
  availability_set_id = azurerm_availability_set.controllers.id

  size = var.controller_type

  # storage
  os_disk {
    name                 = "${var.cluster_name}-controller-${count.index}"
    caching              = "None"
    disk_size_gb         = var.disk_size
    storage_account_type = "Premium_LRS"
  }

  # Flatcar Container Linux
  source_image_reference {
    publisher = "Kinvolk"
    offer     = "flatcar-container-linux-free"
    sku       = local.channel
    version   = "latest"
  }

  plan {
    name      = local.channel
    publisher = "kinvolk"
    product   = "flatcar-container-linux-free"
  }

  # network
  network_interface_ids = [
    azurerm_network_interface.controllers.*.id[count.index]
  ]

  # Azure requires setting admin_ssh_key, though Ignition custom_data handles it too
  computer_name  = "${var.cluster_name}-controller-${count.index}"
  custom_data    = base64encode(data.ct_config.controller-ignitions.*.rendered[count.index])
  admin_username = "core"
  admin_ssh_key {
    username   = "core"
    public_key = var.ssh_keys[0]
  }
  # Azure mandates setting an ssh_key, provide just a single key as the
  # others are handled with Ignition custom_data.

  lifecycle {
    ignore_changes = [
      custom_data,
      os_disk,
    ]
  }
}

# Controller NICs with public and private IPv4
resource "azurerm_network_interface" "controllers" {
  count               = var.controller_count
  resource_group_name = azurerm_resource_group.cluster.name

  name     = "${var.cluster_name}-controller-${count.index}"
  location = azurerm_resource_group.cluster.location

  ip_configuration {
    name                          = "ip0"
    subnet_id                     = azurerm_subnet.controller.id
    private_ip_address_allocation = "Dynamic"
    # instance public IPv4
    public_ip_address_id = azurerm_public_ip.controllers.*.id[count.index]
  }
}

# Associate controller network interface with controller security group
resource "azurerm_network_interface_security_group_association" "controllers" {
  count = var.controller_count

  network_interface_id      = azurerm_network_interface.controllers[count.index].id
  network_security_group_id = azurerm_network_security_group.controller.id
}


# Associate controller network interface with controller backend address pool
resource "azurerm_network_interface_backend_address_pool_association" "controllers" {
  count = var.controller_count

  network_interface_id    = azurerm_network_interface.controllers[count.index].id
  ip_configuration_name   = "ip0"
  backend_address_pool_id = azurerm_lb_backend_address_pool.controller.id
}

# Controller public IPv4 addresses
resource "azurerm_public_ip" "controllers" {
  count               = var.controller_count
  resource_group_name = azurerm_resource_group.cluster.name

  name              = "${var.cluster_name}-controller-${count.index}"
  location          = azurerm_resource_group.cluster.location
  sku               = "Standard"
  allocation_method = "Static"
}

# Controller Ignition configs
data "ct_config" "controller-ignitions" {
  count        = var.controller_count
  content      = data.template_file.controller-configs[count.index].rendered
  pretty_print = false
  snippets     = var.controller_clc_snippets
}

# Controller Container Linux configs
data "template_file" "controller-configs" {
  count = var.controller_count

  template = file("${path.module}/cl/controller.yaml.tmpl")

  vars = {
    cluster_name = var.cluster_name
    # Cannot use cyclic dependencies on controllers or their DNS records
    etcd_name   = "etcd${count.index}"
    etcd_domain = "${var.cluster_name}-etcd${count.index}.${var.dns_zone}"
    # etcd0=https://cluster-etcd0.example.com,etcd1=https://cluster-etcd1.example.com,...
    etcd_initial_cluster   = join(",", [for i in range(var.controller_count) : format("etcd%d=https://%s-etcd%d.%s:2380", i, var.cluster_name, i, var.dns_zone)])
    kubeconfig             = indent(10, module.bootkube.kubeconfig-kubelet)
    ssh_keys               = jsonencode(var.ssh_keys)
    cluster_dns_service_ip = cidrhost(var.service_cidr, 10)
    cluster_domain_suffix  = var.cluster_domain_suffix
    enable_tls_bootstrap   = var.enable_tls_bootstrap
    dns_zone               = var.dns_zone
  }
}

data "template_file" "etcds" {
  count    = var.controller_count
  template = "etcd$${index}=https://$${cluster_name}-etcd$${index}.$${dns_zone}:2380"

  vars = {
    index        = count.index
    cluster_name = var.cluster_name
    dns_zone     = var.dns_zone
  }
}
