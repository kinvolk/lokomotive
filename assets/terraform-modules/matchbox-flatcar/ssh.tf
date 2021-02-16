resource "null_resource" "reprovision-node-when-ignition-changes" {
  # Triggered when the Ignition Config changes
  triggers = {
    ignition_config = matchbox_profile.node.raw_ignition
    kernel_args = join(" ", var.kernel_args)
    kernel_console = join(" ", var.kernel_console)
  }
  # Wait for the new Ignition config object to be ready before rebooting
  depends_on = [matchbox_group.node]
  # Trigger running Ignition on the next reboot (first_boot flag file) and reboot the instance, or, if the instance needs to be (re)provisioned, run external commands for PXE booting (also runs on the first provisioning)
  provisioner "local-exec" {
    command = templatefile("${path.module}/pxe-helper.sh.tmpl", { domain = var.node_domain, name = var.node_name, mac = var.node_mac, pxe_commands = var.pxe_commands, asset_dir = var.asset_dir, kernel_args = join(" ", var.kernel_args), kernel_console = join(" ", var.kernel_console), ignition_endpoint = format("%s/ignition", var.http_endpoint), ignore_changes = var.ignore_changes })
  }
}
