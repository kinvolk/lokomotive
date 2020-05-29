provider "aws" {
  # The Route 53 service doesn't need a specific region to operate, however
  # the AWS Terraform provider needs it and the documentation suggests to use
  # "us-east-1": https://docs.aws.amazon.com/general/latest/gr/r53.html.
  region = "us-east-1"
}

data "aws_route53_zone" "selected" {
  name = "${var.dns_zone}."
}

resource "aws_route53_record" "apiserver_public" {
  zone_id = data.aws_route53_zone.selected.zone_id
  name    = format("%s.%s.", var.cluster_name, var.dns_zone)
  type    = "A"
  ttl     = 300
  records = var.controllers_public_ipv4
}

resource "aws_route53_record" "apiserver_private" {
  zone_id = data.aws_route53_zone.selected.zone_id
  name    = format("%s-private.%s.", var.cluster_name, var.dns_zone)
  type    = "A"
  ttl     = 300
  records = var.controllers_private_ipv4
}

resource "aws_route53_record" "etcd" {
  count = length(var.controllers_private_ipv4)

  zone_id = data.aws_route53_zone.selected.zone_id
  name    = format("%s-etcd%d.%s.", var.cluster_name, count.index, var.dns_zone)
  type    = "A"
  ttl     = 300
  records = [var.controllers_private_ipv4[count.index]]
}
