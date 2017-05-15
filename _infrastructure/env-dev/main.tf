provider "aws" {
  region = "${var.aws_region}"
}

resource "aws_ecr_repository" "gateway" {
  name = "gateway"
}

resource "aws_alb_target_group" "gateway" {
  name     = "${var.name}"
  port     = 80                                                     // it's overriden while registering instance with task by ECS
  protocol = "HTTP"
  vpc_id   = "${data.terraform_remote_state.infrastructure.vpc_id}"

  health_check {
    healthy_threshold = 2
    path              = "/status"
  }
}

resource "aws_alb_listener_rule" "path" {
  listener_arn = "${data.terraform_remote_state.infrastructure.api_alb_listener_id}"
  priority     = 150

  action {
    type             = "forward"
    target_group_arn = "${aws_alb_target_group.gateway.arn}"
  }

  condition {
    field  = "path-pattern"
    values = ["/v0/gateway/*"]
  }
}
