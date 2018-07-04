data "aws_caller_identity" "current" {}

locals {
  container_linux_version = "${data.external.version.result["version"]}"
}

data "external" "version" {
  program = ["sh", "-c", "curl https://${var.container_linux_channel}.release.core-os.net/amd64-usr/current/version.txt | sed -n 's/COREOS_VERSION=\\(.*\\)$/{\"version\": \"\\1\"}/p'"]
}

module "etcd" {
  source = "github.com/coreos/tectonic-installer//modules/aws/etcd?ref=0a22c73d39f67ba4bb99106a9e72322a47179736"

  base_domain             = "${var.base_domain}"
  cluster_id              = "${var.cluster_name}"
  cluster_name            = "${var.cluster_name}"
  container_image         = "${var.container_image}"
  container_linux_channel = "${var.container_linux_channel}"
  container_linux_version = "${local.container_linux_version}"
  ec2_type                = "${var.ec2_type}"
  external_endpoints      = []
  extra_tags              = "${var.tags}"
  ign_etcd_crt_id_list    = "${local.etcd_crt_id_list}"
  ign_etcd_dropin_id_list = "${data.ignition_systemd_unit.etcd.*.id}"
  instance_count          = "${var.instance_count}"
  root_volume_iops        = "${var.root_volume_iops}"
  root_volume_size        = "${var.root_volume_size}"
  root_volume_type        = "${var.root_volume_type}"
  s3_bucket               = "${aws_s3_bucket.eg-etcd-ignition.id}"
  sg_ids                  = "${aws_security_group.etcd.*.id}"
  ssh_key                 = "${var.ssh_key}"
  subnets                 = "${var.subnets}"
  tls_enabled             = "${var.tls_enabled}"
}

module "etcd_certs" {
  source = "github.com/coreos/tectonic-installer//modules/tls/etcd/signed?ref=0a22c73d39f67ba4bb99106a9e72322a47179736"

  etcd_ca_cert_path     = "/dev/null"
  etcd_cert_dns_names   = "${data.template_file.etcd_hostname_list.*.rendered}"
  etcd_client_cert_path = "/dev/null"
  etcd_client_key_path  = "/dev/null"
  self_signed           = "true"
  service_cidr          = "10.3.0.0/16"
}

resource "aws_s3_bucket" "eg-etcd-ignition" {
  bucket = "${data.aws_caller_identity.current.account_id}-eg-etc-ignition"
  acl    = "private"

  tags = "${var.tags}"
}

resource "aws_security_group" "etcd" {
  name   = "eg-etcd"
  vpc_id = "${var.vpc_id}"

  ingress {
    protocol        = "tcp"
    from_port       = "2379"
    to_port         = "2380"
    self            = true
    security_groups = ["${var.security_groups}"]
  }

  egress {
    protocol    = "-1"
    from_port   = 0
    to_port     = 0
    cidr_blocks = ["0.0.0.0/0"]
  }
}

# If bastion instance is present in a public subnet
resource "aws_security_group_rule" "allow_ssh" {
  count = "${var.bastion_enabled ? 1 : 0}"

  type      = "ingress"
  from_port = 22
  to_port   = 22
  protocol  = "tcp"

  cidr_blocks = ["0.0.0.0/0"]

  security_group_id = "${aws_security_group.etcd.id}"
}
