resource "aws_s3_bucket" "data" {
  bucket_prefix = "osmptv-${var.env}-data-"
  force_destroy = true
}
