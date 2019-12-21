# Wildcard DNS Record for ingress
resource "aws_route53_record" "ingress-wildcard" {
  zone_id = var.dns_zone_id
  name    = format("*.%s.%s.", var.cluster_name, var.dns_zone)
  type    = "CNAME"
  ttl     = 60
  records = [
    format("%s.%s.", var.cluster_name, var.dns_zone),
  ]
}

