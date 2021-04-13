# Discrete DNS records for each controller's private IPv4 for etcd usage
resource "aws_route53_record" "controllers" {
  count = var.controller_count

  # DNS Zone where record should be created
  zone_id = var.dns_zone_id

  name = format("%s-%d.%s.", var.cluster_name, count.index, var.dns_zone)
  type = "A"
  ttl  = 300

  # private IPv4 address for etcd
  records = [oci_core_instance.controllers[count.index].private_ip]
}

resource "aws_route53_record" "etcds" {
  count = var.controller_count

  # DNS Zone where record should be created
  zone_id = var.dns_zone_id

  name = format("%s-etcd%d.%s.", var.cluster_name, count.index, var.dns_zone)
  type = "A"
  ttl  = 300

  # private IPv4 address for etcd
  records = [oci_core_instance.controllers[count.index].private_ip]
}

data "oci_core_image" "flatcar" {
  image_id = var.controller_image_id
}

# Controller instances

data "oci_identity_availability_domain" "ad" {
  compartment_id = var.tenancy_id
  ad_number      = 1
}

resource "oci_core_instance" "controllers" {
  count = var.controller_count

  availability_domain = data.oci_identity_availability_domain.ad.name
  compartment_id = var.compartment_id
  shape = var.controller_instance_shape

  create_vnic_details {
    assign_public_ip = true
    display_name     = "flatcar-vnic"
    freeform_tags = merge(var.tags, {
      "Name" = var.cluster_name
    })
    hostname_label = "${var.cluster_name}-${count.index}"
    subnet_id = oci_core_subnet.subnet.id
    nsg_ids = [oci_core_network_security_group.lokomotive.id]
  }

  freeform_tags =  merge(var.tags, {
    Name = "${var.cluster_name}-controller-${count.index}"
  })

  display_name = "${var.cluster_name}-${count.index}"

  source_details {
    source_id = data.oci_core_image.flatcar.id
    source_type = "image"

    boot_volume_size_in_gbs = var.disk_size
  }

  shape_config {
    ocpus = var.controller_cpus
    memory_in_gbs = var.controller_memory
  }

  metadata = {
    user_data = base64encode(data.ct_config.controller-ignitions[count.index].rendered)
  }
}

# Controller Ignition configs
data "ct_config" "controller-ignitions" {
  count = var.controller_count

  content = templatefile("${path.module}/cl/controller.yaml.tmpl", {
    # Cannot use cyclic dependencies on controllers or their DNS records
    etcd_name   = "etcd${count.index}"
    etcd_arch_tag_suffix = var.os_arch == "arm64" ? "-arm64" : ""
    etcd_arch_options    = var.os_arch == "arm64" ? "ETCD_UNSUPPORTED_ARCH=arm64" : ""
    etcd_domain = "${var.cluster_name}-etcd${count.index}.${var.dns_zone}"
    # etcd0=https://cluster-etcd0.example.com,etcd1=https://cluster-etcd1.example.com,...
    etcd_initial_cluster   = join(",", [for i in range(var.controller_count) : format("etcd%d=https://%s-etcd%d.%s:2380", i, var.cluster_name, i, var.dns_zone)])
    ssh_keys               = jsonencode(var.ssh_keys)
    cluster_dns_service_ip = cidrhost(var.service_cidr, 10)
    cluster_domain_suffix  = var.cluster_domain_suffix
    enable_tls_bootstrap   = var.enable_tls_bootstrap
    domain_name            = "${var.cluster_name}-${count.index}.${var.dns_zone}"
  })
  pretty_print = false
  snippets     = var.controller_clc_snippets
}
