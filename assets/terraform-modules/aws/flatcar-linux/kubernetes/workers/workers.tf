# Workers AutoScaling Group
resource "aws_autoscaling_group" "workers" {
  name = "${var.pool_name}-worker"

  # count
  desired_capacity          = var.worker_count
  min_size                  = var.worker_count
  max_size                  = var.worker_count + 2
  default_cooldown          = 30
  health_check_grace_period = 30

  # network
  vpc_zone_identifier = var.subnet_ids

  # template
  launch_configuration = aws_launch_configuration.worker.name

  # target groups to which instances should be added
  target_group_arns = flatten([
    aws_lb_target_group.workers_http.id,
    aws_lb_target_group.workers_https.id,
    var.target_groups,
  ])

  lifecycle {
    # override the default destroy and replace update behavior
    create_before_destroy = true
  }

  # Waiting for instance creation delays adding the ASG to state. If instances
  # can't be created (e.g. spot price too low), the ASG will be orphaned.
  # Orphaned ASGs escape cleanup, can't be updated, and keep bidding if spot is
  # used. Disable wait to avoid issues and align with other clouds.
  wait_for_capacity_timeout = "0"

  tags = flatten([
    [
      {
        key                 = "Name"
        value               = "${var.cluster_name}-${var.pool_name}-worker"
        propagate_at_launch = true
      },
    ],
    [
      for tag in keys(var.tags) :
      {
        key                 = tag == "Name" ? "X-Name" : tag
        value               = var.tags[tag]
        propagate_at_launch = true
      }
    ],
  ])
}

# IAM Policy.
resource "aws_iam_instance_profile" "csi_driver" {
  count = var.enable_csi ? 1 : 0
  role  = join("", aws_iam_role.csi_driver.*.name)
}

resource "aws_iam_role_policy" "csi_driver" {
  count = var.enable_csi ? 1 : 0
  role  = join("", aws_iam_role.csi_driver.*.id)

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

resource "aws_iam_role" "csi_driver" {
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

# Worker template
resource "aws_launch_configuration" "worker" {
  name_prefix       = "${var.cluster_name}-${var.pool_name}-"
  image_id          = data.aws_ami.flatcar.image_id
  instance_type     = var.instance_type
  spot_price        = var.spot_price
  enable_monitoring = false

  user_data = data.ct_config.worker-ignition.rendered

  # storage
  root_block_device {
    volume_type = var.disk_type
    volume_size = var.disk_size
    iops        = var.disk_iops
    encrypted   = true
  }

  # network
  security_groups = var.security_groups

  lifecycle {
    // Override the default destroy and replace update behavior
    create_before_destroy = true
    ignore_changes        = [image_id]
  }

  iam_instance_profile = join("", aws_iam_instance_profile.csi_driver.*.name)
}

# Worker Ignition config
data "ct_config" "worker-ignition" {
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
  })
  pretty_print = false
  snippets     = var.clc_snippets
}
