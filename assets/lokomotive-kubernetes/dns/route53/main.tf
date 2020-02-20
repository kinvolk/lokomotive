variable "entries" {
  type = list(
    object({
      name    = string
      type    = string
      ttl     = number
      records = list(string)
    })
  )
}

variable "aws_zone_id" {
  type        = string
  description = "AWS Route53 DNS Zone ID (e.g. Z3PAABBCFAKEC0)"
}

resource "aws_route53_record" "dns-records" {
  count = length(var.entries)

  # Route53 DNS Zone where record should be created
  zone_id = var.aws_zone_id

  name    = var.entries[count.index].name
  type    = var.entries[count.index].type
  ttl     = var.entries[count.index].ttl
  records = var.entries[count.index].records
}
