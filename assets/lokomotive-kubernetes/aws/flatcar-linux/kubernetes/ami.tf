locals {
  # Pick a Container Linux derivative
  ami_id = "${element(concat(data.aws_ami.flatcar.*.image_id, data.aws_ami.coreos.*.image_id, list("")), 0)}"

  flavor  = var.os_name
  channel = var.os_channel
  ver     = var.os_version == "current" ? "" : var.os_version
}

data "aws_ami" "coreos" {
  count = "${local.flavor == "coreos" ? 1 : 0}"

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
    values = ["CoreOS-${local.channel}-${local.ver}*"]
  }
}

data "aws_ami" "flatcar" {
  count = "${local.flavor == "flatcar" ? 1 : 0}"

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
    values = ["Flatcar-${local.channel}-${local.ver}*"]
  }
}

# test
