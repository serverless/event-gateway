module "alb-events" {
  source                   = "terraform-aws-modules/alb/aws"
  load_balancer_name       = "${var.events_alb_name}"
  security_groups          = ["${aws_security_group.alb.id}"]
  logging_enabled          = "false"
  subnets                  = ["${module.vpc.public_subnets}"]
  tags                     = "${merge(var.tags, map("Name", var.events_alb_name))}"
  vpc_id                   = "${module.vpc.vpc_id}"
  http_tcp_listeners       = "${list(map("port", "80", "protocol", "HTTP"))}"
  http_tcp_listeners_count = "1"
  target_groups            = "${list(map("health_check_path", "/v1/status", "health_check_port", "4001", "name", "tg-events", "backend_protocol", "HTTP", "backend_port", "80", "target_type", "ip"))}"
  target_groups_count      = "1"
}

module "alb-config" {
  source                   = "terraform-aws-modules/alb/aws"
  load_balancer_name       = "${var.config_alb_name}"
  security_groups          = ["${aws_security_group.alb.id}"]
  logging_enabled          = "false"
  subnets                  = ["${module.vpc.public_subnets}"]
  tags                     = "${merge(var.tags, map("Name", var.config_alb_name))}"
  vpc_id                   = "${module.vpc.vpc_id}"
  http_tcp_listeners       = "${list(map("port", "80", "protocol", "HTTP"))}"
  http_tcp_listeners_count = "1"
  target_groups            = "${list(map("health_check_path", "/v1/status", "name", "tg-config", "backend_protocol", "HTTP", "backend_port", "80", "target_type", "ip"))}"
  target_groups_count      = "1"
}

resource "aws_security_group" "alb" {
  name   = "eg-alb"
  vpc_id = "${module.vpc.vpc_id}"

  ingress {
    protocol    = "tcp"
    from_port   = 80
    to_port     = 80
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = "${var.tags}"
}
