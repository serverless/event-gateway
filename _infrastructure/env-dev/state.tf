data "terraform_remote_state" "infrastructure" {
  backend = "s3"

  config {
    bucket = "serverlessdev-tfstate"
    key    = "infrastructure/env-dev.tfstate"
    region = "${var.aws_region}"
  }
}
