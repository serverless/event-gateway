data "aws_caller_identity" "current" {}

resource "aws_ecs_cluster" "event-gateway" {
  name = "eg-ecs-cluster"
}

resource "aws_cloudwatch_log_group" "eg" {
  name = "eg-ecs-group"
  tags = "${var.tags}"
}

resource "aws_ecs_task_definition" "eg" {
  family                   = "eg"
  network_mode             = "awsvpc"
  requires_compatibilities = ["FARGATE"]
  cpu                      = "${var.fargate_cpu}"
  memory                   = "${var.fargate_memory}"
  execution_role_arn       = "arn:aws:iam::${data.aws_caller_identity.current.account_id}:role/ecsTaskExecutionRole"

  container_definitions = <<EOF
[
  {
    "name": "event-gateway",
    "image": "${var.eg_image}",
    "command": ["${join("\",\"", var.command_list)}"],
    "networkMode": "awsvpc",  
    "portMappings": [
      {
        "containerPort": ${var.config_port}
      },
      {
        "containerPort": ${var.events_port}
      }
    ],
    "logConfiguration": {
        "logDriver": "awslogs",
        "options": {
            "awslogs-group": "${aws_cloudwatch_log_group.eg.name}",
            "awslogs-region": "${var.aws_region}",
            "awslogs-stream-prefix": "ecs"
        }
    }
  }
]
EOF
}

resource "aws_ecs_service" "config" {
  name            = "eg-config"
  cluster         = "${aws_ecs_cluster.event-gateway.id}"
  task_definition = "${aws_ecs_task_definition.eg.arn}"
  desired_count   = "${var.task_count}"
  launch_type     = "FARGATE"

  network_configuration {
    security_groups = ["${aws_security_group.sg-container-ingress.id}"]
    subnets         = ["${module.vpc.private_subnets}"]
  }

  load_balancer {
    target_group_arn = "${module.alb-config.target_group_arns[0]}"
    container_name   = "event-gateway"
    container_port   = "${var.config_port}"
  }

  depends_on = [
    "module.alb-config",
  ]
}

resource "aws_ecs_service" "events" {
  name            = "eg-events"
  cluster         = "${aws_ecs_cluster.event-gateway.id}"
  task_definition = "${aws_ecs_task_definition.eg.arn}"
  desired_count   = "${var.task_count}"
  launch_type     = "FARGATE"

  network_configuration {
    security_groups = ["${aws_security_group.sg-container-ingress.id}"]
    subnets         = ["${module.vpc.private_subnets}"]
  }

  load_balancer {
    target_group_arn = "${module.alb-events.target_group_arns[0]}"
    container_name   = "event-gateway"
    container_port   = "${var.events_port}"
  }

  depends_on = [
    "module.alb-events",
  ]
}

resource "aws_security_group" "sg-container-ingress" {
  name        = "eg-event-service"
  description = "Allow only into events API"
  vpc_id      = "${module.vpc.vpc_id}"

  ingress {
    protocol        = "tcp"
    from_port       = "${var.events_port}"
    to_port         = "${var.config_port}"
    security_groups = ["${aws_security_group.alb.id}"]
  }

  egress {
    protocol    = "-1"
    from_port   = 0
    to_port     = 0
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = "${var.tags}"
}
