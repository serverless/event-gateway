resource "aws_route53_zone" "etcd_priv" {
  name    = "${var.base_domain}"
  vpc_id  = "${var.vpc_id}"
  comment = "Managed by Terraform"

  tags = "${var.tags}"
}

resource "aws_route53_record" "etcd_a_nodes" {
  count   = "${var.instance_count}"
  type    = "A"
  ttl     = "60"
  zone_id = "${aws_route53_zone.etcd_priv.zone_id}"
  name    = "${var.cluster_name}-etcd-${count.index}"
  records = ["${module.etcd.ip_addresses[count.index]}"]
}
