resource "aws_iam_user" "gateway" {
  name = "${var.name}-${var.environment}-${var.aws_region}"
}

resource "aws_iam_access_key" "gateway" {
  user = "${aws_iam_user.gateway.name}"
}

resource "aws_iam_user_policy" "lambda" {
  name = "lambda-invoke"
  user = "${aws_iam_user.gateway.name}"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "lambda:InvokeFunction"
      ],
      "Resource": "*"
    }
  ]
}
EOF
}
