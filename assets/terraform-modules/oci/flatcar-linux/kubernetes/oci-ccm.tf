resource "local_file" "oci_ccm" {
  filename = "${var.asset_dir}/charts/kube-system/oci-ccm.yaml"
  content = templatefile("${path.module}/oci-ccm.yaml.tmpl", {
    region = var.region
    tenancy = var.tenancy_id
    user = var.user
    fingerprint = var.fingerprint
    compartment = var.compartment_id
    vcn = oci_core_vcn.network.id
    subnet = oci_core_subnet.subnet.id
    key = var.key
  })
}
