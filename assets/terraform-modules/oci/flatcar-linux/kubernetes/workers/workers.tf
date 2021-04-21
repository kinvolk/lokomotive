data "oci_core_image" "flatcar" {
  image_id =  var.worker_image_id
}

data "oci_identity_availability_domain" "ad" {
  compartment_id = var.tenancy_id
  ad_number      = var.worker_ad_number
}

resource "oci_core_instance" "controllers" {
  count = var.worker_count

  availability_domain = data.oci_identity_availability_domain.ad.name
  compartment_id = var.compartment_id
  shape = var.worker_instance_shape

  create_vnic_details {
    assign_public_ip = true
    display_name     = "flatcar-vnic"
    freeform_tags = merge(var.tags, {
      "Name" = var.cluster_name
    })
    hostname_label = "${var.pool_name}-${count.index}"
    subnet_id = var.subnet_id
    nsg_ids = [var.nsg_id]
  }

  freeform_tags =  merge(var.tags, {
    Name = "${var.pool_name}-${count.index}"
  })
  display_name = "${var.pool_name}-${count.index}"

  source_details {
    source_id = data.oci_core_image.flatcar.id
    source_type = "image"

    boot_volume_size_in_gbs = var.disk_size
  }

  shape_config {
    ocpus = var.worker_cpus
    memory_in_gbs = var.worker_memory
  }

  metadata = {
    user_data = base64encode(data.ct_config.worker-ignition[count.index].rendered)
  }
}

# Worker Ignition config
data "ct_config" "worker-ignition" {
  count = var.worker_count

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
    node_labels            = merge({ "node.kubernetes.io/node" = "" }, var.labels)
    taints                 = var.taints
    enable_tls_bootstrap   = var.enable_tls_bootstrap
    domain_name            = "${var.pool_name}-${count.index}.${var.dns_zone}"
  })
  pretty_print = false
  snippets     = var.clc_snippets
}
