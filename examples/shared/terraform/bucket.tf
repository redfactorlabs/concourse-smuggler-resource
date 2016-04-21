variable "suffix" {}

provider "aws" {
}

resource "aws_s3_bucket" "s3-test" {
  bucket = "smuggler-s3-test-${var.suffix}"
  acl = "private"
  force_destroy = "true"
  versioning {
    enabled = true
  }
}

