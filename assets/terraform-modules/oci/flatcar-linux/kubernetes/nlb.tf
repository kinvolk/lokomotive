# Network Load Balancer DNS Record
resource "aws_route53_record" "apiserver" {
  zone_id = var.dns_zone_id

  name = format("%s.%s.", var.cluster_name, var.dns_zone)
  type = "A"
  ttl = 300

  records = [oci_core_instance.controllers[0].public_ip]
}

// No NLB for now, just to get it into a minimal working state without fighting the cloud too much

/*
resource "oci_core_public_ip" "nlb_public_ip" {
  compartment_id = var.compartment_id
  lifetime = "ephemeral"
}

# Network Load Balancer for apiservers and ingress
resource "oci_network_load_balancer_network_load_balancer" "nlb" {
  compartment_id = var.compartment_id
  display_name  = "${var.cluster_name}-nlb"
  subnet_id = oci_core_subnet.subnet.id

  is_preserve_source_destination = true
  is_private = false
  reserved_ips {
    id = oci_core_public_ip.nlb_public_ip.id
  }
}

# Forward TCP apiserver traffic to controllers
resource "oci_network_load_balancer_backend_set" "apiserver-https" {
  health_checker {
    protocol = "TCP"
    port = "6443"
  }
  name = "${var.cluster_name}-apiserver-https"
  network_load_balancer_id = oci_network_load_balancer_network_load_balancer.nlb.id

  is_preserve_source = true

  policy = "FIVE_TUPLE"
}

resource "oci_network_load_balancer_backend" "apiserver-https" {
  count = var.controller_count

  backend_set_name = oci_network_load_balancer_backend_set.apiserver-https.name
  network_load_balancer_id = oci_network_load_balancer_network_load_balancer.nlb.id
  port = "6443"
  ip_address = oci_core_instance.controllers[count.index].private_ip
  name = oci_core_instance.controllers[count.index].display_name
}

resource "oci_network_load_balancer_listener" "apiserver-https" {
  default_backend_set_name = oci_network_load_balancer_backend_set.apiserver-https.name
  name = "${var.cluster_name}-apiserver-https-nlb"
  network_load_balancer_id = oci_network_load_balancer_network_load_balancer.nlb.id
  port = "6443"
  protocol = "TCP"
}
*/