locals {
  # Channel for a CoreOS Container Linux derivative
  # coreos-stable -> CoreOS Container Linux Stable
  channel = split("-", var.os_image)[1]
}

# data "azurerm_image" "custom_workers" {
#   name                = var.custom_image_name
#   resource_group_name = var.custom_image_resource_group_name
# }

# Workers scale set
resource "azurerm_linux_virtual_machine_scale_set" "workers" {
  resource_group_name = var.resource_group_name

  name                   = "${var.pool_name}-worker"
  location               = var.region
  sku                    = var.vm_type
  instances              = var.worker_count
  single_placement_group = false


  # storage
  os_disk {
    storage_account_type = "Standard_LRS"
    caching              = "ReadWrite"
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

  computer_name_prefix = "${var.pool_name}-worker"
  admin_username       = "core"
  custom_data          = base64encode(data.ct_config.worker-ignition.rendered)

  # Azure mandates setting an ssh_key, provide just a single key as the
  # others are handled with Ignition custom_data.
  disable_password_authentication = true
  admin_ssh_key {
    username   = "core"
    public_key = var.ssh_keys[0]
  }

  # network
  network_interface {
    name                      = "nic0"
    primary                   = true
    network_security_group_id = var.security_group_id

    ip_configuration {
      name      = "ip0"
      primary   = true
      subnet_id = var.subnet_id

      # backend address pool to which the NIC should be added
      load_balancer_backend_address_pool_ids = [var.backend_address_pool_id]
    }
  }

  # lifecycle
  upgrade_mode = "Manual"
  # eviction policy may only be set when priority is Spot
  priority        = var.priority
  eviction_policy = var.priority == "Spot" ? "Delete" : null
}

# Scale up or down to maintain desired number, tolerating deallocations.
resource "azurerm_monitor_autoscale_setting" "workers" {
  resource_group_name = var.resource_group_name

  name     = "${var.pool_name}-maintain-desired"
  location = var.region

  # autoscale
  enabled            = true
  target_resource_id = azurerm_linux_virtual_machine_scale_set.workers.id

  profile {
    name = "default"

    capacity {
      minimum = var.worker_count
      default = var.worker_count
      maximum = var.worker_count
    }
  }
}

# Worker Ignition configs
data "ct_config" "worker-ignition" {
  content = templatefile("${path.module}/cl/worker.yaml.tmpl", {
    kubeconfig = var.enable_tls_bootstrap ? indent(10, templatefile("${path.module}/cl/bootstrap-kubeconfig.yaml.tmpl", {
      token_id     = random_string.bootstrap_token_id[0].result
      token_secret = random_string.bootstrap_token_secret[0].result
      ca_cert      = var.ca_cert
      server       = "https://${var.apiserver}:6443"
    })) : indent(10, var.kubeconfig)
    ssh_keys               = jsonencode(var.ssh_keys)
    cluster_dns_service_ip = cidrhost(var.service_cidr, 10)
    cluster_domain_suffix  = var.cluster_domain_suffix
    enable_tls_bootstrap   = var.enable_tls_bootstrap
    cpu_manager_policy     = var.cpu_manager_policy
    system_reserved_cpu    = var.system_reserved_cpu
    kube_reserved_cpu      = var.kube_reserved_cpu
    node_labels            = merge({ "node.kubernetes.io/node" = "" }, var.labels)
    taints                 = var.taints
    dns_zone               = var.dns_zone
    cluster_name           = var.cluster_name
  })
  pretty_print = false
  snippets     = var.clc_snippets
}
