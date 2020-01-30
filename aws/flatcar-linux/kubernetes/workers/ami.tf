locals {
  # Pick a CoreOS Container Linux derivative
  ami_id = local.flavor == "flatcar" ? data.aws_ami.flatcar.image_id : data.aws_ami.coreos.image_id

  flavor  = var.os_name
  channel = var.os_channel
  ver = var.os_version == "current" ? "" : var.os_version
}

data "aws_ami" "coreos" {
  most_recent = true
  owners      = ["595879546273"]

  filter {
    name   = "architecture"
    values = ["x86_64"]
  }

  filter {
    name   = "virtualization-type"
    values = ["hvm"]
  }

  filter {
    name   = "name"
    values = ["CoreOS-${local.flavor == "coreos" ? local.channel : "stable"}-${local.flavor == "coreos" ? local.ver : ""}*"]
  }
}

data "aws_ami" "flatcar" {
  most_recent = true
  owners      = ["075585003325"]

  filter {
    name   = "architecture"
    values = ["x86_64"]
  }

  filter {
    name   = "virtualization-type"
    values = ["hvm"]
  }

  filter {
    name   = "name"
    values = ["Flatcar-${local.flavor == "flatcar" ? local.channel : "stable"}-${local.flavor == "flatcar" ? local.ver : ""}*"]
  }
}
