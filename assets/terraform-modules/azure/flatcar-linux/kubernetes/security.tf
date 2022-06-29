# Controller security group

resource "azurerm_network_security_group" "controller" {
  resource_group_name = azurerm_resource_group.cluster.name

  name     = "${var.cluster_name}-controller"
  location = azurerm_resource_group.cluster.location
}

resource "azurerm_network_security_rule" "controller-icmp" {
  resource_group_name = azurerm_resource_group.cluster.name

  name                        = "allow-icmp"
  network_security_group_name = azurerm_network_security_group.controller.name
  priority                    = "1995"
  access                      = "Allow"
  direction                   = "Inbound"
  protocol                    = "Icmp"
  source_port_range           = "*"
  destination_port_range      = "*"
  source_address_prefixes     = [azurerm_subnet.controller.address_prefix, azurerm_subnet.worker.address_prefix]
  destination_address_prefix  = azurerm_subnet.controller.address_prefix
}

resource "azurerm_network_security_rule" "controller-ssh" {
  resource_group_name = azurerm_resource_group.cluster.name

  name                        = "allow-ssh"
  network_security_group_name = azurerm_network_security_group.controller.name
  priority                    = "2000"
  access                      = "Allow"
  direction                   = "Inbound"
  protocol                    = "Tcp"
  source_port_range           = "*"
  destination_port_range      = "22"
  source_address_prefix       = "*"
  destination_address_prefix  = azurerm_subnet.controller.address_prefix
}

resource "azurerm_network_security_rule" "controller-etcd" {
  resource_group_name = azurerm_resource_group.cluster.name

  name                        = "allow-etcd"
  network_security_group_name = azurerm_network_security_group.controller.name
  priority                    = "2005"
  access                      = "Allow"
  direction                   = "Inbound"
  protocol                    = "Tcp"
  source_port_range           = "*"
  destination_port_range      = "2379-2380"
  source_address_prefix       = azurerm_subnet.controller.address_prefix
  destination_address_prefix  = azurerm_subnet.controller.address_prefix
}

# Allow Prometheus to scrape etcd metrics
resource "azurerm_network_security_rule" "controller-etcd-metrics" {
  resource_group_name = azurerm_resource_group.cluster.name

  name                        = "allow-etcd-metrics"
  network_security_group_name = azurerm_network_security_group.controller.name
  priority                    = "2010"
  access                      = "Allow"
  direction                   = "Inbound"
  protocol                    = "Tcp"
  source_port_range           = "*"
  destination_port_range      = "2381"
  source_address_prefix       = azurerm_subnet.worker.address_prefix
  destination_address_prefix  = azurerm_subnet.controller.address_prefix
}

# Allow Prometheus to scrape kube-proxy metrics
resource "azurerm_network_security_rule" "controller-kube-proxy" {
  resource_group_name = azurerm_resource_group.cluster.name

  name                        = "allow-kube-proxy-metrics"
  network_security_group_name = azurerm_network_security_group.controller.name
  priority                    = "2011"
  access                      = "Allow"
  direction                   = "Inbound"
  protocol                    = "Tcp"
  source_port_range           = "*"
  destination_port_range      = "10249"
  source_address_prefix       = azurerm_subnet.worker.address_prefix
  destination_address_prefix  = azurerm_subnet.controller.address_prefix
}

# Allow Prometheus to scrape kube-scheduler and kube-controller-manager metrics
resource "azurerm_network_security_rule" "controller-kube-metrics" {
  resource_group_name = azurerm_resource_group.cluster.name

  name                        = "allow-kube-metrics"
  network_security_group_name = azurerm_network_security_group.controller.name
  priority                    = "2012"
  access                      = "Allow"
  direction                   = "Inbound"
  protocol                    = "Tcp"
  source_port_range           = "*"
  destination_port_range      = "10257-10259"
  source_address_prefix       = azurerm_subnet.worker.address_prefix
  destination_address_prefix  = azurerm_subnet.controller.address_prefix
}

resource "azurerm_network_security_rule" "controller-apiserver" {
  resource_group_name = azurerm_resource_group.cluster.name

  name                        = "allow-apiserver"
  network_security_group_name = azurerm_network_security_group.controller.name
  priority                    = "2015"
  access                      = "Allow"
  direction                   = "Inbound"
  protocol                    = "Tcp"
  source_port_range           = "*"
  destination_port_range      = "6443"
  source_address_prefix       = "*"
  destination_address_prefix  = azurerm_subnet.controller.address_prefix
}

resource "azurerm_network_security_rule" "controller-vxlan" {
  resource_group_name = azurerm_resource_group.cluster.name

  name                        = "allow-vxlan"
  network_security_group_name = azurerm_network_security_group.controller.name
  priority                    = "2020"
  access                      = "Allow"
  direction                   = "Inbound"
  protocol                    = "Udp"
  source_port_range           = "*"
  destination_port_range      = "4789"
  source_address_prefixes     = [azurerm_subnet.controller.address_prefix, azurerm_subnet.worker.address_prefix]
  destination_address_prefix  = azurerm_subnet.controller.address_prefix
}

resource "azurerm_network_security_rule" "controller-linux-vxlan" {
  resource_group_name = azurerm_resource_group.cluster.name

  name                        = "allow-linux-vxlan"
  network_security_group_name = azurerm_network_security_group.controller.name
  priority                    = "2021"
  access                      = "Allow"
  direction                   = "Inbound"
  protocol                    = "Udp"
  source_port_range           = "*"
  destination_port_range      = "8472"
  source_address_prefixes     = [azurerm_subnet.controller.address_prefix, azurerm_subnet.worker.address_prefix]
  destination_address_prefix  = azurerm_subnet.controller.address_prefix
}

# Allow Prometheus to scrape node-exporter daemonset
resource "azurerm_network_security_rule" "controller-node-exporter" {
  resource_group_name = azurerm_resource_group.cluster.name

  name                        = "allow-node-exporter"
  network_security_group_name = azurerm_network_security_group.controller.name
  priority                    = "2025"
  access                      = "Allow"
  direction                   = "Inbound"
  protocol                    = "Tcp"
  source_port_range           = "*"
  destination_port_range      = "9100"
  source_address_prefix       = azurerm_subnet.worker.address_prefix
  destination_address_prefix  = azurerm_subnet.controller.address_prefix
}

# Allow apiserver to access kubelet's for exec, log, port-forward
resource "azurerm_network_security_rule" "controller-kubelet" {
  resource_group_name = azurerm_resource_group.cluster.name

  name                        = "allow-kubelet"
  network_security_group_name = azurerm_network_security_group.controller.name
  priority                    = "2030"
  access                      = "Allow"
  direction                   = "Inbound"
  protocol                    = "Tcp"
  source_port_range           = "*"
  destination_port_range      = "10250"

  # allow Prometheus to scrape kubelet metrics too
  source_address_prefixes    = [azurerm_subnet.controller.address_prefix, azurerm_subnet.worker.address_prefix]
  destination_address_prefix = azurerm_subnet.controller.address_prefix
}

# Override Azure AllowVNetInBound and AllowAzureLoadBalancerInBound
# https://docs.microsoft.com/en-us/azure/virtual-network/security-overview#default-security-rules

resource "azurerm_network_security_rule" "controller-allow-loadblancer" {
  resource_group_name = azurerm_resource_group.cluster.name

  name                        = "allow-loadbalancer"
  network_security_group_name = azurerm_network_security_group.controller.name
  priority                    = "3000"
  access                      = "Allow"
  direction                   = "Inbound"
  protocol                    = "*"
  source_port_range           = "*"
  destination_port_range      = "*"
  source_address_prefix       = "AzureLoadBalancer"
  destination_address_prefix  = "*"
}

resource "azurerm_network_security_rule" "controller-deny-all" {
  resource_group_name = azurerm_resource_group.cluster.name

  name                        = "deny-all"
  network_security_group_name = azurerm_network_security_group.controller.name
  priority                    = "3005"
  access                      = "Deny"
  direction                   = "Inbound"
  protocol                    = "*"
  source_port_range           = "*"
  destination_port_range      = "*"
  source_address_prefix       = "*"
  destination_address_prefix  = "*"
}

# Worker security group

resource "azurerm_network_security_group" "worker" {
  resource_group_name = azurerm_resource_group.cluster.name

  name     = "${var.cluster_name}-worker"
  location = azurerm_resource_group.cluster.location
}

resource "azurerm_network_security_rule" "worker-icmp" {
  resource_group_name = azurerm_resource_group.cluster.name

  name                        = "allow-icmp"
  network_security_group_name = azurerm_network_security_group.worker.name
  priority                    = "1995"
  access                      = "Allow"
  direction                   = "Inbound"
  protocol                    = "Icmp"
  source_port_range           = "*"
  destination_port_range      = "*"
  source_address_prefixes     = [azurerm_subnet.controller.address_prefix, azurerm_subnet.worker.address_prefix]
  destination_address_prefix  = azurerm_subnet.worker.address_prefix
}

resource "azurerm_network_security_rule" "worker-ssh" {
  resource_group_name = azurerm_resource_group.cluster.name

  name                        = "allow-ssh"
  network_security_group_name = azurerm_network_security_group.worker.name
  priority                    = "2000"
  access                      = "Allow"
  direction                   = "Inbound"
  protocol                    = "Tcp"
  source_port_range           = "*"
  destination_port_range      = "22"
  source_address_prefix       = azurerm_subnet.controller.address_prefix
  destination_address_prefix  = azurerm_subnet.worker.address_prefix
}

resource "azurerm_network_security_rule" "worker-http" {
  resource_group_name = azurerm_resource_group.cluster.name

  name                        = "allow-http"
  network_security_group_name = azurerm_network_security_group.worker.name
  priority                    = "2005"
  access                      = "Allow"
  direction                   = "Inbound"
  protocol                    = "Tcp"
  source_port_range           = "*"
  destination_port_range      = "80"
  source_address_prefix       = "*"
  destination_address_prefix  = azurerm_subnet.worker.address_prefix
}

resource "azurerm_network_security_rule" "worker-https" {
  resource_group_name = azurerm_resource_group.cluster.name

  name                        = "allow-https"
  network_security_group_name = azurerm_network_security_group.worker.name
  priority                    = "2010"
  access                      = "Allow"
  direction                   = "Inbound"
  protocol                    = "Tcp"
  source_port_range           = "*"
  destination_port_range      = "443"
  source_address_prefix       = "*"
  destination_address_prefix  = azurerm_subnet.worker.address_prefix
}

resource "azurerm_network_security_rule" "worker-vxlan" {
  resource_group_name = azurerm_resource_group.cluster.name

  name                        = "allow-vxlan"
  network_security_group_name = azurerm_network_security_group.worker.name
  priority                    = "2015"
  access                      = "Allow"
  direction                   = "Inbound"
  protocol                    = "Udp"
  source_port_range           = "*"
  destination_port_range      = "4789"
  source_address_prefixes     = [azurerm_subnet.controller.address_prefix, azurerm_subnet.worker.address_prefix]
  destination_address_prefix  = azurerm_subnet.worker.address_prefix
}

resource "azurerm_network_security_rule" "worker-linux-vxlan" {
  resource_group_name = azurerm_resource_group.cluster.name

  name                        = "allow-linux-vxlan"
  network_security_group_name = azurerm_network_security_group.worker.name
  priority                    = "2016"
  access                      = "Allow"
  direction                   = "Inbound"
  protocol                    = "Udp"
  source_port_range           = "*"
  destination_port_range      = "8472"
  source_address_prefixes     = [azurerm_subnet.controller.address_prefix, azurerm_subnet.worker.address_prefix]
  destination_address_prefix  = azurerm_subnet.worker.address_prefix
}

# Allow Prometheus to scrape node-exporter daemonset
resource "azurerm_network_security_rule" "worker-node-exporter" {
  resource_group_name = azurerm_resource_group.cluster.name

  name                        = "allow-node-exporter"
  network_security_group_name = azurerm_network_security_group.worker.name
  priority                    = "2020"
  access                      = "Allow"
  direction                   = "Inbound"
  protocol                    = "Tcp"
  source_port_range           = "*"
  destination_port_range      = "9100"
  source_address_prefix       = azurerm_subnet.worker.address_prefix
  destination_address_prefix  = azurerm_subnet.worker.address_prefix
}

# Allow Prometheus to scrape kube-proxy
resource "azurerm_network_security_rule" "worker-kube-proxy" {
  resource_group_name = azurerm_resource_group.cluster.name

  name                        = "allow-kube-proxy"
  network_security_group_name = azurerm_network_security_group.worker.name
  priority                    = "2024"
  access                      = "Allow"
  direction                   = "Inbound"
  protocol                    = "Tcp"
  source_port_range           = "*"
  destination_port_range      = "10249"
  source_address_prefix       = azurerm_subnet.worker.address_prefix
  destination_address_prefix  = azurerm_subnet.worker.address_prefix
}

# Allow apiserver to access kubelet's for exec, log, port-forward
resource "azurerm_network_security_rule" "worker-kubelet" {
  resource_group_name = azurerm_resource_group.cluster.name

  name                        = "allow-kubelet"
  network_security_group_name = azurerm_network_security_group.worker.name
  priority                    = "2025"
  access                      = "Allow"
  direction                   = "Inbound"
  protocol                    = "Tcp"
  source_port_range           = "*"
  destination_port_range      = "10250"

  # allow Prometheus to scrape kubelet metrics too
  source_address_prefixes    = [azurerm_subnet.controller.address_prefix, azurerm_subnet.worker.address_prefix]
  destination_address_prefix = azurerm_subnet.worker.address_prefix
}

# Override Azure AllowVNetInBound and AllowAzureLoadBalancerInBound
# https://docs.microsoft.com/en-us/azure/virtual-network/security-overview#default-security-rules

resource "azurerm_network_security_rule" "worker-allow-loadblancer" {
  resource_group_name = azurerm_resource_group.cluster.name

  name                        = "allow-loadbalancer"
  network_security_group_name = azurerm_network_security_group.worker.name
  priority                    = "3000"
  access                      = "Allow"
  direction                   = "Inbound"
  protocol                    = "*"
  source_port_range           = "*"
  destination_port_range      = "*"
  source_address_prefix       = "AzureLoadBalancer"
  destination_address_prefix  = "*"
}

resource "azurerm_network_security_rule" "worker-deny-all" {
  resource_group_name = azurerm_resource_group.cluster.name

  name                        = "deny-all"
  network_security_group_name = azurerm_network_security_group.worker.name
  priority                    = "3005"
  access                      = "Deny"
  direction                   = "Inbound"
  protocol                    = "*"
  source_port_range           = "*"
  destination_port_range      = "*"
  source_address_prefix       = "*"
  destination_address_prefix  = "*"
}
