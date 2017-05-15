terraform {
  required_version = ">= 0.9.5"

  backend "s3" {
    bucket = "serverlessdev-tfstate"
    region = "us-east-1"
    key    = "gateway/env-dev.tfstate"
  }
}
