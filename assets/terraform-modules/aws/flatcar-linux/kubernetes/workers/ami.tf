locals {
  channel = var.os_channel
  ver     = var.os_version == "current" ? "" : var.os_version
  arch    = var.os_arch
}

data "aws_ami" "flatcar" {
  most_recent = true
  owners      = ["075585003325"]

  filter {
    name   = "architecture"
    values = [local.arch]
  }

  filter {
    name   = "virtualization-type"
    values = ["hvm"]
  }

  filter {
    name   = "name"
    values = ["Flatcar-${local.channel}-${local.ver}*"]
  }
}
