locals {
  dns_name = "www.example.com"
}

resource "aws_route53_record" "www" {
  zone_id = aws_route53_zone.primary.zone_id
  name    = local.dns_name
  type    = "A"
  ttl     = 300
  records = [aws_eip.lb.public_ip]
}

module "route53_records" {
  count = 10

  source = "./module"
}
