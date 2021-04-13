# Security Groups (instance firewalls)

# Controller security group

# Currently set to insecure for first PoC

resource "oci_core_network_security_group" "lokomotive" {
  #Required
  compartment_id = var.compartment_id
  vcn_id = oci_core_vcn.network.id

  #Optional
  display_name = "Lokomotive Insecure"
}

resource "oci_core_network_security_group_security_rule" "allow_all" {
  network_security_group_id = oci_core_network_security_group.lokomotive.id
  direction = "INGRESS"
  protocol = "all"

  source = "0.0.0.0/0"
  source_type = "CIDR_BLOCK"
}