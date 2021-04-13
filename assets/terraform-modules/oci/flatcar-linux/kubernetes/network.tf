# Network VPC, gateway, and routes

resource "oci_core_vcn" "network" {
  compartment_id = var.compartment_id

  dns_label = var.cluster_name // instance.cluster_name.vcn1.oraclevcn.com

  cidr_block = var.host_cidr
  freeform_tags = merge(var.tags, {
    "Name" = var.cluster_name
  })
}

resource "oci_core_internet_gateway" "gateway" {
  compartment_id = var.compartment_id
  vcn_id = oci_core_vcn.network.id

  freeform_tags = merge(var.tags, {
    "Name" = var.cluster_name
  })
}

resource "oci_core_subnet" "subnet" {
  cidr_block = var.host_cidr
  compartment_id = var.compartment_id
  vcn_id = oci_core_vcn.network.id

  dns_label = var.cluster_name // instance.cluster_name.vcn1.oraclevcn.com

  freeform_tags = merge(var.tags, {
    "Name" = var.cluster_name
  })
}

resource "oci_core_route_table" "default" {
  compartment_id = var.compartment_id
  vcn_id = oci_core_vcn.network.id

  freeform_tags = merge(var.tags, {
    "Name" = var.cluster_name
  })
  route_rules {
    network_entity_id = oci_core_internet_gateway.gateway.id
    destination = "0.0.0.0/0"
  }
}

resource "oci_core_route_table_attachment" "test_route_table_attachment" {
  subnet_id = oci_core_subnet.subnet.id
  route_table_id =oci_core_route_table.default.id
}
