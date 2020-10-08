# Discrete DNS records for each controller's private IPv4 for etcd usage
resource "aws_route53_record" "etcds" {
  count = var.controller_count

  # DNS Zone where record should be created
  zone_id = var.dns_zone_id

  name = format("%s-etcd%d.%s.", var.cluster_name, count.index, var.dns_zone)
  type = "A"
  ttl  = 300

  # private IPv4 address for etcd
  records = [aws_instance.controllers[count.index].private_ip]
}

# IAM Policy
resource "aws_iam_instance_profile" "csi-driver" {
  count = var.enable_csi ? 1 : 0
  role  = join("", aws_iam_role.csi-driver.*.name)
}

resource "aws_iam_role_policy" "csi-driver" {
  count = var.enable_csi ? 1 : 0
  role  = join("", aws_iam_role.csi-driver.*.id)

  policy = <<-EOF
  {
    "Version": "2012-10-17",
    "Statement": [
      {
        "Action": [
          "ec2:AttachVolume",
          "ec2:CreateSnapshot",
          "ec2:CreateTags",
          "ec2:CreateVolume",
          "ec2:DeleteSnapshot",
          "ec2:DeleteTags",
          "ec2:DeleteVolume",
          "ec2:DescribeAvailabilityZones",
          "ec2:DescribeInstances",
          "ec2:DescribeSnapshots",
          "ec2:DescribeTags",
          "ec2:DescribeVolumes",
          "ec2:DescribeVolumesModifications",
          "ec2:DetachVolume",
          "ec2:ModifyVolume"
        ],
        "Effect": "Allow",
        "Resource": "*"
      }
    ]
  }
  EOF
}

resource "aws_iam_role" "csi-driver" {
  count = var.enable_csi ? 1 : 0
  path  = "/"
  tags  = var.tags

  assume_role_policy = <<-EOF
  {
    "Version": "2012-10-17",
    "Statement": [
        {
            "Action": "sts:AssumeRole",
            "Principal": {
               "Service": "ec2.amazonaws.com"
            },
            "Effect": "Allow",
            "Sid": ""
        }
    ]
  }
  EOF
}

# Controller instances
resource "aws_instance" "controllers" {
  count = var.controller_count

  tags = merge(var.tags, {
    Name = "${var.cluster_name}-controller-${count.index}"
  })

  instance_type = var.controller_type

  ami                  = local.ami_id
  user_data            = data.ct_config.controller-ignitions[count.index].rendered
  iam_instance_profile = join("", aws_iam_instance_profile.csi-driver.*.name)

  # storage
  root_block_device {
    volume_type = var.disk_type
    volume_size = var.disk_size
    iops        = var.disk_iops
    encrypted   = true
  }

  # network
  associate_public_ip_address = true
  subnet_id                   = aws_subnet.public[count.index].id
  vpc_security_group_ids      = [aws_security_group.controller.id]

  lifecycle {
    ignore_changes = [
      ami,
      user_data,
    ]
  }
}

# Controller Ignition configs
data "ct_config" "controller-ignitions" {
  count = var.controller_count
  content = templatefile("${path.module}/cl/controller.yaml.tmpl", {
    # Cannot use cyclic dependencies on controllers or their DNS records
    etcd_name   = "etcd${count.index}"
    etcd_domain = "${var.cluster_name}-etcd${count.index}.${var.dns_zone}"
    # etcd0=https://cluster-etcd0.example.com,etcd1=https://cluster-etcd1.example.com,...
    etcd_initial_cluster   = join(",", [for i in range(var.controller_count) : format("etcd%d=https://%s-etcd%d.%s:2380", i, var.cluster_name, i, var.dns_zone)])
    ssh_keys               = jsonencode(var.ssh_keys)
    cluster_dns_service_ip = cidrhost(var.service_cidr, 10)
    cluster_domain_suffix  = var.cluster_domain_suffix
    enable_tls_bootstrap   = var.enable_tls_bootstrap
  })
  pretty_print = false
  snippets     = var.controller_clc_snippets
}
