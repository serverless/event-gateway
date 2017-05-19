resource "aws_iam_role" "ecs_service" {
  name = "${var.name}-${var.environment}-${var.aws_region}-ecs-service"

  assume_role_policy = <<ROLE
{
  "Version": "2008-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "ecs.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
ROLE
}

resource "aws_iam_role_policy" "service" {
  name = "service"
  role = "${aws_iam_role.ecs_service.name}"

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "ec2:AuthorizeSecurityGroupIngress",
        "ec2:Describe*",
        "elasticloadbalancing:RegisterTargets",
        "elasticloadbalancing:DeregisterTargets",
        "elasticloadbalancing:Describe*",
        "lambda:InvokeFunction"
      ],
      "Resource": "*"
    }
  ]
}
POLICY
}
