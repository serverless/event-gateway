resource "aws_cloudwatch_log_group" "gateway" {
  name              = "${var.name}"
  retention_in_days = 30
}
