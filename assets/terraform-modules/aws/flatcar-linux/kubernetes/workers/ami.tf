locals {
  ami_id = data.aws_ami.flatcar.*.image_id

  channel = var.os_channel
  ver     = var.os_version == "current" ? "" : var.os_version
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
    values = ["Flatcar-${local.channel}-${local.ver}*"]
  }
}
